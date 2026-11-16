package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

type importDoc struct {
	Version   int                 `json:"version"`
	PackKey   string              `json:"packKey"`
	Lang      string              `json:"lang"`
	DryRun    bool                `json:"dryRun"`
	Missions  []importMission     `json:"missions"`
	Packs     []importPack        `json:"packs"`
	Questions []importQuestion    `json:"questions"`
	Rows      []map[string]string `json:"rows"`
}

type importQuestion struct {
	questionBody
	HintList []hint `json:"hints"`
	IsActive *bool  `json:"isActive"`
}

type importMission struct {
	mission
	IsActive *bool `json:"isActive"`
}

type importPack struct {
	packBody
	IsActive *bool `json:"isActive"`
}

func activeOr(v *bool) bool { return v == nil || *v }

type importError struct {
	Section string `json:"section"`
	Index   int    `json:"index"`
	Field   string `json:"field"`
	Code    string `json:"code"`
}

func errCode(err error) (string, string) {
	var ce *pluginkit.ConsoleError
	if errors.As(err, &ce) {
		return ce.Code, ce.Field
	}
	return "INVALID", ""
}

func rowQuestion(row map[string]string, lang string) (importQuestion, string, string) {
	get := func(k string) string {
		for key, v := range row {
			if strings.EqualFold(strings.TrimSpace(key), k) {
				return strings.TrimSpace(v)
			}
		}
		return ""
	}
	var q importQuestion
	q.Active = true
	q.Key = get("Key")
	q.PackKey = get("Pack Key")
	q.Type = strings.ToLower(get("Type"))
	if q.Type == "" {
		q.Type = "options"
	}
	l := get("Answer Language")
	if l == "" {
		l = lang
	}
	if v := get("Question"); v != "" {
		q.Question = map[string]string{lang: v}
	}
	if v := get("Explanation"); v != "" {
		q.Explanation = map[string]string{lang: v}
	}
	q.Category = get("Category")
	if v := get("Difficulty"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return q, "Difficulty", "VALUE_OUT_OF_RANGE"
		}
		q.Difficulty = n
	}
	if v := get("Time"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return q, "Time", "VALUE_OUT_OF_RANGE"
		}
		q.TimeSeconds = &n
	}
	if q.Type == "options" {
		right := map[int]bool{}
		for _, part := range strings.FieldsFunc(get("Correct Options"), func(r rune) bool { return r == ',' || r == '|' || r == '/' || r == ' ' }) {
			n, err := strconv.Atoi(part)
			if err != nil || n < 1 {
				return q, "Correct Options", "QUIZ_OPTION_INDEX_NOT_FOUND"
			}
			right[n] = true
		}
		for i := 1; i <= 6; i++ {
			if v := get("Option " + strconv.Itoa(i)); v != "" {
				q.Options = append(q.Options, Option{ID: len(q.Options), Text: map[string]string{lang: v}, Correct: right[i]})
			}
		}
		for n := range right {
			if n > len(q.Options) {
				return q, "Correct Options", "QUIZ_OPTION_INDEX_NOT_FOUND"
			}
		}
	} else {
		for _, a := range strings.FieldsFunc(get("Accepted Answers"), func(r rune) bool { return r == '|' || r == '\n' || r == ';' }) {
			if a = strings.TrimSpace(a); a != "" {
				q.Answers = map[string][]string{l: append(q.Answers[l], a)}
			}
		}
	}
	if q.PackKey == "" && get("Pack Name") != "" {
		q.Metadata = map[string]any{"packName": get("Pack Name")}
	}
	return q, "", ""
}

func (s *service) importPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var doc importDoc
	if err := pluginkit.ReadBody(r, &doc, 8<<20); err != nil {
		s.fail(w, err)
		return
	}
	out, err := s.importDocument(ctx, &doc, pluginkit.ConsoleUser(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) importJSON(ctx context.Context, raw []byte, by string) (map[string]any, error) {
	var doc importDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, bad("INVALID_JSON", "")
	}
	return s.importDocument(ctx, &doc, by)
}

func (s *service) importDocument(ctx context.Context, doc *importDoc, by string) (map[string]any, error) {
	if len(doc.Missions)+len(doc.Packs)+len(doc.Questions)+len(doc.Rows) == 0 {
		return nil, bad("QUIZ_IMPORT_FILE_IS_REQUIRED", "questions")
	}
	if len(doc.Questions)+len(doc.Rows) > 5000 || len(doc.Packs) > 1000 || len(doc.Missions) > 500 {
		return nil, bad("QUIZ_IMPORT_TOO_LARGE", "questions")
	}
	if doc.Lang == "" {
		doc.Lang = "fa"
	}
	errs := []importError{}
	add := func(section string, i int, err error) {
		code, field := errCode(err)
		errs = append(errs, importError{Section: section, Index: i, Field: field, Code: code})
	}
	missionKeys, packKeys := map[string]string{}, map[string]bool{}
	for i := range doc.Missions {
		m := &doc.Missions[i].mission
		m.Active = activeOr(doc.Missions[i].IsActive)
		m.Key = strings.TrimSpace(m.Key)
		if m.Kind == "" {
			m.Kind = "mission"
		}
		if m.OrderMode == "" {
			m.OrderMode = "all_at_once"
		}
		var err error
		switch {
		case !settings.KeyPattern.MatchString(m.Key):
			err = bad("INVALID_KEY", "key")
		case m.Kind != "mission" && m.Kind != "game":
			err = bad("INVALID_KIND", "kind")
		case m.OrderMode != "all_at_once" && m.OrderMode != "sequential":
			err = bad("QUIZ_MISSION_ORDER_INVALID", "orderMode")
		case m.TimeSeconds < 0:
			err = bad("VALUE_OUT_OF_RANGE", "timeSeconds")
		}
		if err == nil {
			if m.Name, err = s.cleanText(ctx, "name", m.Name, true, 120); err == nil {
				if m.Description, err = s.cleanText(ctx, "description", m.Description, false, 2000); err == nil {
					err = cleanReward("reward", &m.Reward)
				}
			}
		}
		if err != nil {
			add("missions", i, err)
			continue
		}
		missionKeys[m.Key] = m.Kind
	}
	for i := range doc.Missions {
		m := &doc.Missions[i].mission
		if m.ParentKey != "" && missionKeys[m.ParentKey] != "game" && s.idByKey(ctx, "quiz_mission", m.ParentKey, ` AND kind='game'`) == "" {
			add("missions", i, bad("QUIZ_GAME_NOT_FOUND", "parentKey"))
		}
	}
	for i := range doc.Packs {
		k := &doc.Packs[i].packBody
		k.Active = activeOr(doc.Packs[i].IsActive)
		k.Key = strings.TrimSpace(k.Key)
		if !settings.KeyPattern.MatchString(k.Key) {
			add("packs", i, bad("INVALID_KEY", "key"))
			continue
		}
		mk := strings.TrimSpace(k.MissionKey)
		k.MissionKey, k.MissionID = "", ""
		if err := s.cleanPack(ctx, k); err != nil {
			add("packs", i, err)
			continue
		}
		k.MissionKey = mk
		if mk != "" && missionKeys[mk] != "mission" && s.idByKey(ctx, "quiz_mission", mk, ` AND kind='mission'`) == "" {
			add("packs", i, bad("QUIZ_MISSION_NOT_FOUND", "missionKey"))
		}
		packKeys[k.Key] = true
	}
	questions := append([]importQuestion{}, doc.Questions...)
	sections := make([]string, len(questions))
	index := make([]int, len(questions))
	for i := range questions {
		sections[i], index[i] = "questions", i
	}
	for i, row := range doc.Rows {
		q, field, code := rowQuestion(row, doc.Lang)
		if code != "" {
			errs = append(errs, importError{Section: "rows", Index: i, Field: field, Code: code})
			continue
		}
		questions = append(questions, q)
		sections, index = append(sections, "rows"), append(index, i)
	}
	if doc.PackKey != "" && !packKeys[doc.PackKey] && s.idByKey(ctx, "quiz_pack", doc.PackKey, "") == "" {
		return nil, bad("QUIZ_PACK_NOT_FOUND", "packKey")
	}
	seenKeys := map[string]bool{}
	for i := range questions {
		q := &questions[i]
		if q.IsActive != nil || sections[i] == "questions" {
			q.Active = activeOr(q.IsActive)
		}
		pk := strings.TrimSpace(q.PackKey)
		if doc.PackKey != "" {
			pk = doc.PackKey
		}
		q.PackKey, q.PackID = "", ""
		if err := s.cleanQuestion(ctx, &q.questionBody); err != nil {
			add(sections[i], index[i], err)
			continue
		}
		if q.Key != "" {
			if seenKeys[q.Key] {
				add(sections[i], index[i], bad("DUPLICATE_KEY", "key"))
			}
			seenKeys[q.Key] = true
		}
		q.PackKey = pk
		if pk != "" && !packKeys[pk] && s.idByKey(ctx, "quiz_pack", pk, "") == "" {
			add(sections[i], index[i], bad("QUIZ_PACK_NOT_FOUND", "packKey"))
		}
		for j := range q.HintList {
			h := &q.HintList[j]
			var err error
			if h.Name, err = s.cleanText(ctx, "hints.name", h.Name, true, 120); err == nil {
				h.Text, err = s.cleanText(ctx, "hints.text", h.Text, false, 2000)
			}
			if err == nil && h.Price.Amount > 0 && !settings.KeyPattern.MatchString(h.Price.CurrencyKey) {
				err = bad("QUIZ_HINT_PRICE_REQUIRES_CURRENCY_ID", "hints.price")
			}
			if err != nil {
				add(sections[i], index[i], err)
			}
		}
	}
	if len(errs) > 0 {
		return nil, &pluginkit.ConsoleError{Status: http.StatusBadRequest, Code: "QUIZ_IMPORT_HAS_ERRORS", Message: strconv.Itoa(len(errs)) + " error(s); nothing was imported", Meta: map[string]any{"errors": errs}}
	}
	counts := map[string]map[string]int{"missions": {"created": 0, "updated": 0}, "packs": {"created": 0, "updated": 0}, "questions": {"created": 0, "updated": 0}}
	if doc.DryRun {
		counts["missions"]["created"], counts["packs"]["created"], counts["questions"]["created"] = len(doc.Missions), len(doc.Packs), len(questions)
		return map[string]any{"dryRun": true, "counts": counts}, nil
	}
	err := pluginkit.InTx(ctx, s.db, func(tx *sql.Tx) error {
		lookup := func(table, key string) string {
			var id string
			_ = tx.QueryRowContext(ctx, `SELECT id::text FROM `+table+` WHERE key=$1`, key).Scan(&id)
			return id
		}
		for _, pass := range []string{"game", "mission"} {
			for i := range doc.Missions {
				m := &doc.Missions[i].mission
				if m.Kind != pass {
					continue
				}
				parent := ""
				if m.ParentKey != "" && m.Kind == "mission" {
					parent = lookup("quiz_mission", m.ParentKey)
				}
				if m.Metadata == nil {
					m.Metadata = map[string]any{}
				}
				id := lookup("quiz_mission", m.Key)
				if id == "" {
					counts["missions"]["created"]++
					id = uuid.Must(uuid.NewV4()).String()
				} else {
					counts["missions"]["updated"]++
				}
				if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_mission (id, key, kind, parent_id, name, description, icon_url, sort, order_mode, time_seconds, send_answers, reward, is_active, metadata, updated_by)
					VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
					ON CONFLICT (key) DO UPDATE SET kind=EXCLUDED.kind, parent_id=EXCLUDED.parent_id, name=EXCLUDED.name, description=EXCLUDED.description, icon_url=EXCLUDED.icon_url,
					sort=EXCLUDED.sort, order_mode=EXCLUDED.order_mode, time_seconds=EXCLUDED.time_seconds, send_answers=EXCLUDED.send_answers, reward=EXCLUDED.reward,
					is_active=EXCLUDED.is_active, metadata=EXCLUDED.metadata, updated_by=EXCLUDED.updated_by, updated_at=now()`,
					id, m.Key, m.Kind, nullable(parent), jsonOf(m.Name), jsonOf(m.Description), m.IconURL, m.Sort, m.OrderMode, m.TimeSeconds, m.SendAnswers, jsonOf(m.Reward), m.Active, jsonOf(m.Metadata), by); err != nil {
					return err
				}
			}
		}
		for i := range doc.Packs {
			k := &doc.Packs[i].packBody
			if k.MissionKey != "" {
				k.MissionID = lookup("quiz_mission", k.MissionKey)
			}
			id := lookup("quiz_pack", k.Key)
			create := id == ""
			if create {
				id = uuid.Must(uuid.NewV4()).String()
				counts["packs"]["created"]++
			} else {
				counts["packs"]["updated"]++
			}
			if _, err := savePack(ctx, tx, id, create, k, by); err != nil {
				return err
			}
		}
		for i := range questions {
			q := &questions[i]
			if q.PackKey != "" {
				q.PackID = lookup("quiz_pack", q.PackKey)
			}
			id := ""
			if q.Key != "" {
				_ = tx.QueryRowContext(ctx, `SELECT id::text FROM quiz_question WHERE key=$1`, q.Key).Scan(&id)
			}
			create := id == ""
			if create {
				id = uuid.Must(uuid.NewV4()).String()
				counts["questions"]["created"]++
			} else {
				counts["questions"]["updated"]++
			}
			q.question.TimeSeconds = q.TimeSeconds
			if _, err := saveQuestion(ctx, tx, id, create, &q.questionBody, by); err != nil {
				return err
			}
			if create {
				for j, h := range q.HintList {
					if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_hint (id, question_id, name, text, media_url, price, sort, is_active, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,true,$8)`,
						uuid.Must(uuid.NewV4()).String(), id, jsonOf(h.Name), jsonOf(h.Text), h.MediaURL, jsonOf(h.Price), j*10, by); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		if isUnique(err) {
			return nil, conflict("ALREADY_EXISTS", "key")
		}
		return nil, err
	}
	return map[string]any{"dryRun": false, "counts": counts}, nil
}

func (s *service) exportGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	packWhere, args := `TRUE`, []any{}
	if v := q.Get("packId"); v != "" {
		packWhere, args = `k.id::text=$1`, []any{v}
	} else if v := q.Get("missionId"); v != "" {
		packWhere, args = `k.mission_id::text=$1`, []any{v}
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+packCols+packFrom+` WHERE `+packWhere+` ORDER BY k.sort, k.key`, args...)
	if err != nil {
		s.fail(w, err)
		return
	}
	packs := []*pack{}
	missionIDs := map[string]bool{}
	ids := []string{}
	for rows.Next() {
		k, err := scanPack(rows)
		if err != nil {
			rows.Close()
			s.fail(w, err)
			return
		}
		packs = append(packs, k)
		ids = append(ids, k.ID)
		if k.MissionID != "" {
			missionIDs[k.MissionID] = true
		}
	}
	rows.Close()
	where, qargs := `q.pack_id::text = ANY($1::text[])`, []any{ids}
	if len(args) == 0 {
		where, qargs = `TRUE`, nil
	}
	questions, err := listQuestions(ctx, s.db, where, qargs...)
	if err != nil {
		s.fail(w, err)
		return
	}
	qids := []string{}
	for _, x := range questions {
		qids = append(qids, x.ID)
	}
	hints, err := listHints(ctx, s.db, qids, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	missions := []*mission{}
	mrows, err := s.db.QueryContext(ctx, `SELECT `+missionCols+missionFrom+` ORDER BY m.kind DESC, m.sort, m.key`)
	if err == nil {
		for mrows.Next() {
			m, err := scanMission(mrows)
			if err == nil && (len(args) == 0 || missionIDs[m.ID]) {
				missions = append(missions, m)
			}
		}
		mrows.Close()
	}
	outQ := []map[string]any{}
	for _, x := range questions {
		hl := []map[string]any{}
		for _, h := range hints[x.ID] {
			hl = append(hl, map[string]any{"name": h.Name, "text": h.Text, "mediaUrl": h.MediaURL, "price": h.Price})
		}
		outQ = append(outQ, map[string]any{"key": x.Key, "packKey": x.PackKey, "type": x.Type, "question": x.Question, "mediaUrl": x.MediaURL, "options": x.Options,
			"answers": x.Answers, "explanation": x.Explanation, "difficulty": x.Difficulty, "category": x.Category, "tags": x.Tags, "timeSeconds": x.TimeSeconds,
			"reward": x.Reward, "sort": x.Sort, "isActive": x.Active, "metadata": x.Metadata, "hints": hl})
	}
	w.Header().Set("Content-Disposition", `attachment; filename="quiz-export.json"`)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"version": 1, "missions": missions, "packs": packs, "questions": outQ})
}
