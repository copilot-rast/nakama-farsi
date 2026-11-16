package main

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

func (s *service) registerConsoleHTTP(initializer runtime.Initializer) error {
	type route struct {
		path, method string
		handler      http.HandlerFunc
	}
	p := consolePrefix
	routes := []route{
		{p + "/settings", http.MethodGet, s.settingsGet},
		{p + "/settings", http.MethodPut, s.settingsPut},
		{p + "/refs", http.MethodGet, s.refs},
		{p + "/missions", http.MethodGet, s.missionList},
		{p + "/missions", http.MethodPost, s.missionSave},
		{p + "/missions/{id}", http.MethodGet, s.missionGet},
		{p + "/missions/{id}", http.MethodPut, s.missionSave},
		{p + "/missions/{id}", http.MethodDelete, s.missionDelete},
		{p + "/packs", http.MethodGet, s.packList},
		{p + "/packs", http.MethodPost, s.packSave},
		{p + "/packs/{id}", http.MethodGet, s.packGet},
		{p + "/packs/{id}", http.MethodPut, s.packSave},
		{p + "/packs/{id}", http.MethodDelete, s.packDelete},
		{p + "/questions", http.MethodGet, s.questionList},
		{p + "/questions", http.MethodPost, s.questionSave},
		{p + "/questions/{id}", http.MethodGet, s.questionGet},
		{p + "/questions/{id}", http.MethodPut, s.questionSave},
		{p + "/questions/{id}", http.MethodDelete, s.questionDelete},
		{p + "/questions/{id}/hints", http.MethodGet, s.hintList},
		{p + "/questions/{id}/hints", http.MethodPost, s.hintSave},
		{p + "/hints/{hid}", http.MethodPut, s.hintSave},
		{p + "/hints/{hid}", http.MethodDelete, s.hintDelete},
		{p + "/import", http.MethodPost, s.importPost},
		{p + "/export", http.MethodGet, s.exportGet},
		{p + "/reports", http.MethodGet, s.reportList},
		{p + "/reports/{id}", http.MethodPut, s.reportReview},
		{p + "/reasons", http.MethodGet, s.reasonList},
		{p + "/reasons", http.MethodPost, s.reasonSave},
		{p + "/reasons/{id}", http.MethodPut, s.reasonSave},
		{p + "/reasons/{id}", http.MethodDelete, s.reasonDelete},
		{p + "/templates", http.MethodGet, s.templateList},
		{p + "/templates", http.MethodPost, s.templateSave},
		{p + "/templates/{id}", http.MethodPut, s.templateSave},
		{p + "/templates/{id}", http.MethodDelete, s.templateDelete},
		{p + "/templates/{id}/deliveries", http.MethodGet, s.deliveryList},
		{p + "/stats", http.MethodGet, s.stats},
		{p + "/players/{id}", http.MethodGet, s.playerGet},
		{p + "/players/{id}/packs/{packId}", http.MethodDelete, s.playerReset},
	}
	for _, d := range s.duelRoutes() {
		routes = append(routes, route{d.path, d.method, d.handler})
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

func bad(code, field string) error { return pluginkit.Bad(code, field, "") }

func conflict(code, field string) error {
	return &pluginkit.ConsoleError{Status: http.StatusConflict, Code: code, Field: field}
}

func (s *service) settingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, meta, err := settings.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"settings": cfg, "defaults": settings.Defaults(), "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt,
		"prizes": economyhooks.PrizesAvailable(), "campaign": s.campaignOn(), "secrets": s.secretKey != nil})
}

func (s *service) settingsPut(w http.ResponseWriter, r *http.Request) {
	in := settings.Defaults()
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if field, code := in.Validate(); code != "" {
		s.fail(w, bad(code, field))
		return
	}
	for field, key := range map[string]string{"duel.winPrize": in.Duel.WinPrize, "duel.losePrize": in.Duel.LosePrize, "duel.drawPrize": in.Duel.DrawPrize} {
		if key == "" {
			continue
		}
		var ok bool
		if err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM economy_prize WHERE key=$1)`, key).Scan(&ok); err == nil && !ok {
			s.fail(w, bad("PRIZE_NOT_FOUND", field))
			return
		}
	}
	if err := settings.Save(r.Context(), s.db, in, pluginkit.ConsoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	settings.Invalidate()
	s.settingsGet(w, r)
}

func (s *service) refs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	langs, lang, _ := localized.ForUser(ctx, s.db, "", r.URL.Query().Get("lang"))
	out := map[string]any{}
	for name, q := range map[string]string{
		"currencies":     `SELECT key, name::text FROM economy_currency ORDER BY key`,
		"prizes":         `SELECT key, name::text FROM economy_prize ORDER BY key LIMIT 2000`,
		"prizeGroups":    `SELECT key, name::text FROM economy_prize_group ORDER BY key`,
		"missions":       `SELECT key, name::text FROM quiz_mission WHERE kind='mission' ORDER BY sort, key`,
		"games":          `SELECT key, name::text FROM quiz_mission WHERE kind='game' ORDER BY sort, key`,
		"packs":          `SELECT key, name::text FROM quiz_pack ORDER BY sort, key`,
		"stages":         `SELECT key, name::text FROM campaign_stage ORDER BY (mode='once') DESC, key`,
		"categories":     `SELECT DISTINCT c, c FROM (SELECT category AS c FROM quiz_question UNION SELECT category FROM quiz_pack UNION SELECT key FROM quiz_category) x WHERE c <> '' ORDER BY 1`,
		"duelCategories": `SELECT key, name::text FROM quiz_category ORDER BY sort, key`,
		"tags":           `SELECT DISTINCT t, t FROM (SELECT unnest(tags) AS t FROM quiz_question UNION SELECT unnest(tags) FROM quiz_pack) x ORDER BY 1 LIMIT 500`,
	} {
		list := []map[string]string{}
		rows, err := s.db.QueryContext(ctx, q)
		if err == nil {
			for rows.Next() {
				var key, name string
				if rows.Scan(&key, &name) == nil {
					list = append(list, map[string]string{"key": key, "name": resolveName(langs, lang, name, key)})
				}
			}
			rows.Close()
		}
		out[name] = list
	}
	out["placeholders"] = placeholders
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func resolveName(langs *localized.Set, lang, raw, fallback string) string {
	if langs != nil && strings.HasPrefix(raw, "{") {
		if v := langs.Resolve(localized.Parse(raw), lang); v != "" {
			return v
		}
	}
	if raw != "" && !strings.HasPrefix(raw, "{") {
		return raw
	}
	return fallback
}

func (s *service) cleanText(ctx context.Context, field string, in map[string]string, required bool, max int) (map[string]string, error) {
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		return nil, err
	}
	out, err := langs.Clean(field, localized.Text(in), required, max)
	if err != nil {
		var le *localized.Error
		if errors.As(err, &le) {
			return nil, bad(le.Code, field)
		}
		return nil, err
	}
	if out == nil {
		out = localized.Text{}
	}
	return out, nil
}

func cleanReward(field string, r *Reward) error {
	clean := map[string]int64{}
	for k, v := range r.Currencies {
		k = strings.TrimSpace(k)
		if !settings.KeyPattern.MatchString(k) {
			return bad("INVALID_KEY", field+".currencies")
		}
		if v < 0 || v > 1e12 {
			return bad("VALUE_OUT_OF_RANGE", field+".currencies."+k)
		}
		if v > 0 {
			clean[k] = v
		}
	}
	if len(clean) > 20 {
		return bad("TOO_MANY_CURRENCIES", field+".currencies")
	}
	r.Currencies = clean
	r.PrizeKey, r.PrizeGroupKey = strings.TrimSpace(r.PrizeKey), strings.TrimSpace(r.PrizeGroupKey)
	if r.PrizeKey != "" && !settings.KeyPattern.MatchString(r.PrizeKey) {
		return bad("INVALID_PRIZE_KEY", field+".prizeKey")
	}
	if r.PrizeGroupKey != "" && !settings.KeyPattern.MatchString(r.PrizeGroupKey) {
		return bad("INVALID_PRIZE_KEY", field+".prizeGroupKey")
	}
	if r.PrizeKey != "" && r.PrizeGroupKey != "" {
		return bad("ONE_PRIZE_ONLY", field+".prizeGroupKey")
	}
	return nil
}

func cleanURL(field, v string) (string, error) {
	v = strings.TrimSpace(v)
	if utf8.RuneCountInString(v) > 1000 {
		return "", bad("TEXT_TOO_LONG", field)
	}
	return v, nil
}

func cleanTags(field string, in []string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	for _, t := range in {
		t = strings.TrimSpace(t)
		if t == "" || seen[t] {
			continue
		}
		if utf8.RuneCountInString(t) > 40 {
			return nil, bad("TEXT_TOO_LONG", field)
		}
		seen[t] = true
		out = append(out, t)
	}
	if len(out) > 20 {
		return nil, bad("TOO_MANY_TAGS", field)
	}
	return out, nil
}

func (s *service) idByKey(ctx context.Context, table, key string, extra string) string {
	if key == "" {
		return ""
	}
	var id string
	_ = s.db.QueryRowContext(ctx, `SELECT id::text FROM `+table+` WHERE key=$1`+extra, key).Scan(&id)
	return id
}

func nullable(id string) any {
	if id == "" {
		return nil
	}
	return id
}

func (s *service) missionList(w http.ResponseWriter, r *http.Request) {
	args := &pluginkit.Args{}
	where := "TRUE"
	if v := r.URL.Query().Get("kind"); v != "" {
		where = `m.kind=` + args.Add(v)
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+missionCols+missionFrom+` WHERE `+where+` ORDER BY m.kind DESC, m.sort, m.key`, args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []*mission{}
	for rows.Next() {
		m, err := scanMission(rows)
		if err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, m)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}

func (s *service) missionGet(w http.ResponseWriter, r *http.Request) {
	m, err := loadMission(r.Context(), s.db, `m.id::text=$1`, pluginkit.PathVar(r, "id"))
	if errors.Is(err, sql.ErrNoRows) {
		s.fail(w, pluginkit.Missing("QUIZ_MISSION_NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, m)
}

func (s *service) missionSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in mission
	in.Active = true
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.fail(w, err)
		return
	}
	in.Key = strings.TrimSpace(in.Key)
	if id == "" && !settings.KeyPattern.MatchString(in.Key) {
		s.fail(w, bad("INVALID_KEY", "key"))
		return
	}
	if in.Kind == "" {
		in.Kind = "mission"
	}
	if in.Kind != "mission" && in.Kind != "game" {
		s.fail(w, bad("INVALID_KIND", "kind"))
		return
	}
	if in.OrderMode == "" {
		in.OrderMode = "all_at_once"
	}
	if in.OrderMode != "all_at_once" && in.OrderMode != "sequential" {
		s.fail(w, bad("QUIZ_MISSION_ORDER_INVALID", "orderMode"))
		return
	}
	if in.TimeSeconds < 0 || in.TimeSeconds > 365*86400 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "timeSeconds"))
		return
	}
	var err error
	if in.Name, err = s.cleanText(ctx, "name", in.Name, true, 120); err != nil {
		s.fail(w, err)
		return
	}
	if in.Description, err = s.cleanText(ctx, "description", in.Description, false, 2000); err != nil {
		s.fail(w, err)
		return
	}
	if in.IconURL, err = cleanURL("iconUrl", in.IconURL); err != nil {
		s.fail(w, err)
		return
	}
	if err := cleanReward("reward", &in.Reward); err != nil {
		s.fail(w, err)
		return
	}
	parent := ""
	if in.Kind == "mission" {
		parent = in.ParentID
		if parent == "" && in.ParentKey != "" {
			parent = s.idByKey(ctx, "quiz_mission", in.ParentKey, ` AND kind='game'`)
			if parent == "" {
				s.fail(w, bad("QUIZ_GAME_NOT_FOUND", "parentKey"))
				return
			}
		}
		if parent != "" {
			var ok bool
			if _, err := uuid.FromString(parent); err == nil {
				_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM quiz_mission WHERE id=$1 AND kind='game')`, parent).Scan(&ok)
			}
			if !ok {
				s.fail(w, bad("QUIZ_GAME_NOT_FOUND", "parentId"))
				return
			}
		}
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	by := pluginkit.ConsoleUser(r)
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
		_, err = s.db.ExecContext(ctx, `INSERT INTO quiz_mission (id, key, kind, parent_id, name, description, icon_url, sort, order_mode, time_seconds, send_answers, reward, is_active, metadata, updated_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`, id, in.Key, in.Kind, nullable(parent), jsonOf(in.Name), jsonOf(in.Description), in.IconURL, in.Sort, in.OrderMode,
			in.TimeSeconds, in.SendAnswers, jsonOf(in.Reward), in.Active, jsonOf(in.Metadata), by)
	} else {
		var res sql.Result
		res, err = s.db.ExecContext(ctx, `UPDATE quiz_mission SET parent_id=CASE WHEN kind='mission' THEN $2::uuid ELSE NULL END, name=$3, description=$4, icon_url=$5, sort=$6, order_mode=$7, time_seconds=$8,
			send_answers=$9, reward=$10, is_active=$11, metadata=$12, updated_by=$13, updated_at=now() WHERE id::text=$1`, id, nullable(parent), jsonOf(in.Name), jsonOf(in.Description), in.IconURL, in.Sort,
			in.OrderMode, in.TimeSeconds, in.SendAnswers, jsonOf(in.Reward), in.Active, jsonOf(in.Metadata), by)
		if err == nil {
			if n, _ := res.RowsAffected(); n == 0 {
				s.fail(w, pluginkit.Missing("QUIZ_MISSION_NOT_FOUND"))
				return
			}
		}
	}
	if isUnique(err) {
		s.fail(w, conflict("ALREADY_EXISTS", "key"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	r.SetPathValue("id", id)
	s.missionGet(w, r)
}

func (s *service) missionDelete(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	var items int
	_ = s.db.QueryRowContext(r.Context(), `SELECT (SELECT count(*) FROM quiz_pack WHERE mission_id::text=$1) + (SELECT count(*) FROM quiz_mission WHERE parent_id::text=$1)`, id).Scan(&items)
	if items > 0 {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "IN_USE", Message: "the mission still has packs or missions"})
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM quiz_mission WHERE id::text=$1`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("QUIZ_MISSION_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) packList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	page, limit := pluginkit.Page(r, 50, 200)
	args := &pluginkit.Args{}
	where := []string{"TRUE"}
	if v := strings.TrimSpace(q.Get("missionId")); v != "" {
		if v == "none" {
			where = append(where, `k.mission_id IS NULL`)
		} else {
			where = append(where, `k.mission_id::text=`+args.Add(v))
		}
	}
	if v := strings.TrimSpace(q.Get("category")); v != "" {
		where = append(where, `k.category=`+args.Add(v))
	}
	if v := strings.TrimSpace(q.Get("q")); v != "" {
		like := args.Add(pluginkit.LikePattern(v))
		where = append(where, `(k.key ILIKE `+like+` OR k.name::text ILIKE `+like+`)`)
	}
	cond := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM quiz_pack k WHERE `+cond, args.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	lim, off := args.Add(limit), args.Add((page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT `+packCols+packFrom+` WHERE `+cond+` ORDER BY m.sort NULLS FIRST, k.sort, k.key LIMIT `+lim+` OFFSET `+off, args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []*pack{}
	for rows.Next() {
		k, err := scanPack(rows)
		if err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, k)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) packGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	k, err := loadPack(ctx, s.db, `k.id::text=$1`, pluginkit.PathVar(r, "id"))
	if errors.Is(err, sql.ErrNoRows) {
		s.fail(w, pluginkit.Missing("QUIZ_PACK_NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	var players, plays, wins int
	var avg float64
	_ = s.db.QueryRowContext(ctx, `SELECT count(*), COALESCE(sum(plays),0), COALESCE(sum(wins),0), COALESCE(avg(best_score) FILTER (WHERE plays>0),0)::float8 FROM quiz_user_pack WHERE pack_id=$1`, k.ID).Scan(&players, &plays, &wins, &avg)
	rate := 0.0
	if plays > 0 {
		rate = math.Round(float64(wins)*10000/float64(plays)) / 100
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"pack": k, "stats": map[string]any{"players": players, "plays": plays, "wins": wins, "winRate": rate, "avgBestScore": math.Round(avg*100) / 100}})
}

type packBody struct {
	pack
	PassPercent      *float64 `json:"passPercent"`
	PassCorrect      *int     `json:"passCorrect"`
	ShuffleQuestions *bool    `json:"shuffleQuestions"`
	ShuffleOptions   *bool    `json:"shuffleOptions"`
}

func (s *service) cleanPack(ctx context.Context, in *packBody) error {
	var err error
	if in.Name, err = s.cleanText(ctx, "name", in.Name, true, 120); err != nil {
		return err
	}
	if in.Description, err = s.cleanText(ctx, "description", in.Description, false, 2000); err != nil {
		return err
	}
	if in.IconURL, err = cleanURL("iconUrl", in.IconURL); err != nil {
		return err
	}
	if in.Tags, err = cleanTags("tags", in.Tags); err != nil {
		return err
	}
	in.Category, in.PoolCategory, in.Language, in.CampaignStageKey = strings.TrimSpace(in.Category), strings.TrimSpace(in.PoolCategory), strings.TrimSpace(in.Language), strings.TrimSpace(in.CampaignStageKey)
	switch {
	case utf8.RuneCountInString(in.Category) > 80 || utf8.RuneCountInString(in.PoolCategory) > 80:
		return bad("TEXT_TOO_LONG", "category")
	case in.Language != "" && !localized.ValidCode(in.Language):
		return bad("INVALID_LANGUAGE", "language")
	case in.TimeSeconds < 0 || in.TimeSeconds > 86400:
		return bad("VALUE_OUT_OF_RANGE", "timeSeconds")
	case in.QuestionSeconds < 0 || in.QuestionSeconds > 3600:
		return bad("VALUE_OUT_OF_RANGE", "questionSeconds")
	case in.QuestionCount < 0 || in.QuestionCount > 500:
		return bad("VALUE_OUT_OF_RANGE", "questionCount")
	case in.PassPercent != nil && (*in.PassPercent < 0 || *in.PassPercent > 100):
		return bad("QUIZ_PACK_WIN_THRESHOLD_INVALID", "passPercent")
	case in.PassCorrect != nil && (*in.PassCorrect < 0 || *in.PassCorrect > 500):
		return bad("QUIZ_PACK_WIN_THRESHOLD_INVALID", "passCorrect")
	case in.PassCorrect != nil && in.QuestionCount > 0 && *in.PassCorrect > in.QuestionCount:
		return bad("QUIZ_PACK_WIN_THRESHOLD_INVALID", "passCorrect")
	case in.CampaignStageKey != "" && !settings.KeyPattern.MatchString(in.CampaignStageKey):
		return bad("INVALID_KEY", "campaignStageKey")
	}
	if in.CampaignStageKey != "" {
		var mode string
		if err := s.db.QueryRowContext(ctx, `SELECT mode FROM campaign_stage WHERE key=$1`, in.CampaignStageKey).Scan(&mode); err != nil {
			return bad("STAGE_NOT_FOUND", "campaignStageKey")
		}
	}
	if err := cleanReward("reward", &in.Reward); err != nil {
		return err
	}
	if in.MissionID == "" && in.MissionKey != "" {
		if in.MissionID = s.idByKey(ctx, "quiz_mission", in.MissionKey, ` AND kind='mission'`); in.MissionID == "" {
			return bad("QUIZ_MISSION_NOT_FOUND", "missionKey")
		}
	}
	if in.MissionID != "" {
		var ok bool
		if _, err := uuid.FromString(in.MissionID); err == nil {
			_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM quiz_mission WHERE id=$1 AND kind='mission')`, in.MissionID).Scan(&ok)
		}
		if !ok {
			return bad("QUIZ_MISSION_NOT_FOUND", "missionId")
		}
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	return nil
}

func savePack(ctx context.Context, q interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}, id string, create bool, in *packBody, by string) (bool, error) {
	args := []any{id, nullable(in.MissionID), jsonOf(in.Name), jsonOf(in.Description), in.IconURL, in.Sort, in.Category, in.Tags, in.Language, in.Difficulty,
		in.TimeSeconds, in.QuestionSeconds, in.PassPercent, in.PassCorrect, in.QuestionCount, in.PoolCategory, in.ShuffleQuestions, in.ShuffleOptions, in.SendAnswers,
		jsonOf(in.Reward), in.CampaignStageKey, in.Active, jsonOf(in.Metadata), by}
	if create {
		_, err := q.ExecContext(ctx, `INSERT INTO quiz_pack (id, mission_id, name, description, icon_url, sort, category, tags, language, difficulty, time_seconds, question_seconds,
			pass_percent, pass_correct, question_count, pool_category, shuffle_questions, shuffle_options, send_answers, reward, campaign_stage_key, is_active, metadata, updated_by, key)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8::text[],$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)`, append(args, in.Key)...)
		return true, err
	}
	res, err := q.ExecContext(ctx, `UPDATE quiz_pack SET mission_id=$2, name=$3, description=$4, icon_url=$5, sort=$6, category=$7, tags=$8::text[], language=$9, difficulty=$10,
		time_seconds=$11, question_seconds=$12, pass_percent=$13, pass_correct=$14, question_count=$15, pool_category=$16, shuffle_questions=$17, shuffle_options=$18, send_answers=$19,
		reward=$20, campaign_stage_key=$21, is_active=$22, metadata=$23, updated_by=$24, updated_at=now() WHERE id::text=$1`, args...)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (s *service) packSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in packBody
	in.Active = true
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.fail(w, err)
		return
	}
	in.Key = strings.TrimSpace(in.Key)
	if id == "" && !settings.KeyPattern.MatchString(in.Key) {
		s.fail(w, bad("INVALID_KEY", "key"))
		return
	}
	if err := s.cleanPack(ctx, &in); err != nil {
		s.fail(w, err)
		return
	}
	create := id == ""
	if create {
		id = uuid.Must(uuid.NewV4()).String()
	}
	found, err := savePack(ctx, s.db, id, create, &in, pluginkit.ConsoleUser(r))
	if isUnique(err) {
		field := "key"
		if strings.Contains(err.Error(), "stage") {
			field = "campaignStageKey"
		}
		s.fail(w, conflict("ALREADY_EXISTS", field))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	if !found {
		s.fail(w, pluginkit.Missing("QUIZ_PACK_NOT_FOUND"))
		return
	}
	r.SetPathValue("id", id)
	s.packGet(w, r)
}

func (s *service) packDelete(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	var played bool
	_ = s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM quiz_session WHERE pack_id::text=$1)`, id).Scan(&played)
	if played && r.URL.Query().Get("force") != "true" {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "IN_USE", Message: "the pack was played; switch it off or delete with force=true"})
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM quiz_pack WHERE id::text=$1`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("QUIZ_PACK_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) questionList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	page, limit := pluginkit.Page(r, 50, 500)
	args := &pluginkit.Args{}
	where := []string{"TRUE"}
	if v := strings.TrimSpace(q.Get("packId")); v != "" {
		if v == "none" {
			where = append(where, `q.pack_id IS NULL`)
		} else {
			where = append(where, `q.pack_id::text=`+args.Add(v))
		}
	}
	for _, f := range []string{"category", "type"} {
		if v := strings.TrimSpace(q.Get(f)); v != "" {
			where = append(where, `q.`+f+`=`+args.Add(v))
		}
	}
	if v := strings.TrimSpace(q.Get("tag")); v != "" {
		where = append(where, args.Add(v)+` = ANY(q.tags)`)
	}
	switch q.Get("active") {
	case "true":
		where = append(where, `q.is_active`)
	case "false":
		where = append(where, `NOT q.is_active`)
	}
	if q.Get("reported") == "true" {
		where = append(where, `EXISTS (SELECT 1 FROM quiz_question_report r WHERE r.question_id=q.id AND r.status='open')`)
	}
	if v := strings.TrimSpace(q.Get("q")); v != "" {
		like := args.Add(pluginkit.LikePattern(v))
		where = append(where, `(q.key ILIKE `+like+` OR q.question::text ILIKE `+like+`)`)
	}
	cond := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM quiz_question q WHERE `+cond, args.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	lim, off := args.Add(limit), args.Add((page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT `+questionCols+questionFrom+` WHERE `+cond+` ORDER BY k.key NULLS FIRST, q.sort, q.created_at LIMIT `+lim+` OFFSET `+off, args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []*question{}
	for rows.Next() {
		x, err := scanQuestion(rows)
		if err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, x)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) questionGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := listQuestions(ctx, s.db, `q.id::text=$1`, pluginkit.PathVar(r, "id"))
	if err != nil {
		s.fail(w, err)
		return
	}
	if len(list) == 0 {
		s.fail(w, pluginkit.Missing("QUIZ_NOT_FOUND"))
		return
	}
	x := list[0]
	var asked, right int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE outcome='correct') FROM quiz_answer WHERE question_id=$1`, x.ID).Scan(&asked, &right)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"question": x, "stats": map[string]any{"answers": asked, "correct": right}})
}

type questionBody struct {
	question
	TimeSeconds *int `json:"timeSeconds"`
}

func (s *service) cleanQuestion(ctx context.Context, in *questionBody) error {
	var err error
	in.Key = strings.TrimSpace(in.Key)
	if in.Key != "" && !settings.KeyPattern.MatchString(in.Key) {
		return bad("INVALID_KEY", "key")
	}
	if in.Type == "" {
		in.Type = "options"
	}
	if in.Type != "options" && in.Type != "text" {
		return bad("QUIZ_TYPE_INVALID", "type")
	}
	if in.Question, err = s.cleanText(ctx, "question", in.Question, false, 2000); err != nil {
		return err
	}
	if in.MediaURL, err = cleanURL("mediaUrl", in.MediaURL); err != nil {
		return err
	}
	if len(in.Question) == 0 && in.MediaURL == "" {
		return bad("QUIZ_QUESTION_TEXT_OR_CONTENT_REQUIRED", "question")
	}
	if in.Explanation, err = s.cleanText(ctx, "explanation", in.Explanation, false, 2000); err != nil {
		return err
	}
	if in.Tags, err = cleanTags("tags", in.Tags); err != nil {
		return err
	}
	in.Category = strings.TrimSpace(in.Category)
	if utf8.RuneCountInString(in.Category) > 80 {
		return bad("TEXT_TOO_LONG", "category")
	}
	if in.TimeSeconds != nil && (*in.TimeSeconds < 0 || *in.TimeSeconds > 3600) {
		return bad("VALUE_OUT_OF_RANGE", "timeSeconds")
	}
	if in.Difficulty < 0 || in.Difficulty > 100 {
		return bad("VALUE_OUT_OF_RANGE", "difficulty")
	}
	if in.Type == "options" {
		for i := range in.Options {
			o := &in.Options[i]
			if o.Text, err = s.cleanText(ctx, "options["+strconv.Itoa(i)+"].text", o.Text, false, 500); err != nil {
				return err
			}
			if o.MediaURL, err = cleanURL("options["+strconv.Itoa(i)+"].mediaUrl", o.MediaURL); err != nil {
				return err
			}
		}
		if code := validOptions(in.Options); code != "" {
			return bad(code, "options")
		}
		in.Answers = map[string][]string{}
	} else {
		clean := map[string][]string{}
		for lang, list := range in.Answers {
			if !localized.ValidCode(lang) {
				return bad("INVALID_LANGUAGE", "answers")
			}
			for _, a := range list {
				a = strings.TrimSpace(a)
				if a == "" {
					continue
				}
				if utf8.RuneCountInString(a) > 200 {
					return bad("TEXT_TOO_LONG", "answers")
				}
				clean[lang] = append(clean[lang], a)
			}
		}
		if len(clean) == 0 {
			return bad("QUIZ_TEXT_ANSWERS_REQUIRED", "answers")
		}
		in.Answers, in.Options = clean, []Option{}
	}
	if err := cleanReward("reward", &in.Reward); err != nil {
		return err
	}
	if in.PackID == "" && in.PackKey != "" {
		if in.PackID = s.idByKey(ctx, "quiz_pack", in.PackKey, ""); in.PackID == "" {
			return bad("QUIZ_PACK_NOT_FOUND", "packKey")
		}
	}
	if in.PackID != "" {
		var ok bool
		if _, err := uuid.FromString(in.PackID); err == nil {
			_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM quiz_pack WHERE id=$1)`, in.PackID).Scan(&ok)
		}
		if !ok {
			return bad("QUIZ_PACK_NOT_FOUND", "packId")
		}
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	return nil
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func saveQuestion(ctx context.Context, q execer, id string, create bool, in *questionBody, by string) (bool, error) {
	args := []any{id, nullable(strings.TrimSpace(in.Key)), nullable(in.PackID), in.Type, jsonOf(in.Question), in.MediaURL, jsonOf(in.Options), jsonOf(in.Answers), jsonOf(in.Explanation),
		in.Difficulty, in.Category, in.Tags, in.TimeSeconds, jsonOf(in.Reward), in.Sort, in.Active, jsonOf(in.Metadata), by}
	if create {
		_, err := q.ExecContext(ctx, `INSERT INTO quiz_question (id, key, pack_id, type, question, media_url, options, answers, explanation, difficulty, category, tags, time_seconds, reward, sort, is_active, metadata, updated_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::text[],$13,$14,$15,$16,$17,$18)`, args...)
		return true, err
	}
	res, err := q.ExecContext(ctx, `UPDATE quiz_question SET key=$2, pack_id=$3, type=$4, question=$5, media_url=$6, options=$7, answers=$8, explanation=$9, difficulty=$10, category=$11,
		tags=$12::text[], time_seconds=$13, reward=$14, sort=$15, is_active=$16, metadata=$17, updated_by=$18, updated_at=now() WHERE id::text=$1`, args...)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (s *service) questionSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in questionBody
	in.Active = true
	if err := pluginkit.ReadBody(r, &in, 128*1024); err != nil {
		s.fail(w, err)
		return
	}
	if err := s.cleanQuestion(ctx, &in); err != nil {
		s.fail(w, err)
		return
	}
	in.question.TimeSeconds = in.TimeSeconds
	create := id == ""
	if create {
		id = uuid.Must(uuid.NewV4()).String()
	}
	found, err := saveQuestion(ctx, s.db, id, create, &in, pluginkit.ConsoleUser(r))
	if isUnique(err) {
		s.fail(w, conflict("ALREADY_EXISTS", "key"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	if !found {
		s.fail(w, pluginkit.Missing("QUIZ_NOT_FOUND"))
		return
	}
	r.SetPathValue("id", id)
	s.questionGet(w, r)
}

func (s *service) questionDelete(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	var answered bool
	_ = s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM quiz_answer WHERE question_id::text=$1)`, id).Scan(&answered)
	if answered && r.URL.Query().Get("force") != "true" {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "IN_USE", Message: "players answered this question; switch it off or delete with force=true"})
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM quiz_question WHERE id::text=$1`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("QUIZ_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) hintList(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	hints, err := listHints(r.Context(), s.db, []string{id}, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	items := hints[id]
	if items == nil {
		items = []*hint{}
	}
	out := []map[string]any{}
	for _, h := range items {
		var bought int
		_ = s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM quiz_user_hint WHERE hint_id=$1 AND status='done'`, h.ID).Scan(&bought)
		out = append(out, map[string]any{"id": h.ID, "questionId": h.QuestionID, "name": h.Name, "text": h.Text, "mediaUrl": h.MediaURL, "price": h.Price, "sort": h.Sort,
			"isActive": h.Active, "updatedAt": h.UpdatedAt, "bought": bought})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": out, "total": len(out)})
}

func (s *service) hintSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hid := pluginkit.PathVar(r, "hid")
	var in hint
	in.Active = true
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.fail(w, err)
		return
	}
	var err error
	if in.Name, err = s.cleanText(ctx, "name", in.Name, true, 120); err != nil {
		s.fail(w, err)
		return
	}
	if in.Text, err = s.cleanText(ctx, "text", in.Text, false, 2000); err != nil {
		s.fail(w, err)
		return
	}
	if in.MediaURL, err = cleanURL("mediaUrl", in.MediaURL); err != nil {
		s.fail(w, err)
		return
	}
	if len(in.Text) == 0 && in.MediaURL == "" {
		s.fail(w, bad("QUIZ_HINT_VALUE_REQUIRED", "text"))
		return
	}
	in.Price.CurrencyKey = strings.TrimSpace(in.Price.CurrencyKey)
	if in.Price.Amount < 0 || in.Price.Amount > 1e12 {
		s.fail(w, bad("QUIZ_HINT_PRICE_INVALID", "price.amount"))
		return
	}
	if in.Price.Amount > 0 && !settings.KeyPattern.MatchString(in.Price.CurrencyKey) {
		s.fail(w, bad("QUIZ_HINT_PRICE_REQUIRES_CURRENCY_ID", "price.currencyKey"))
		return
	}
	by := pluginkit.ConsoleUser(r)
	if hid == "" {
		qid := pluginkit.PathVar(r, "id")
		var ok bool
		if _, err := uuid.FromString(qid); err == nil {
			_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM quiz_question WHERE id=$1)`, qid).Scan(&ok)
		}
		if !ok {
			s.fail(w, pluginkit.Missing("QUIZ_NOT_FOUND"))
			return
		}
		hid = uuid.Must(uuid.NewV4()).String()
		_, err = s.db.ExecContext(ctx, `INSERT INTO quiz_hint (id, question_id, name, text, media_url, price, sort, is_active, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			hid, qid, jsonOf(in.Name), jsonOf(in.Text), in.MediaURL, jsonOf(in.Price), in.Sort, in.Active, by)
	} else {
		var res sql.Result
		res, err = s.db.ExecContext(ctx, `UPDATE quiz_hint SET name=$2, text=$3, media_url=$4, price=$5, sort=$6, is_active=$7, updated_by=$8, updated_at=now() WHERE id::text=$1`,
			hid, jsonOf(in.Name), jsonOf(in.Text), in.MediaURL, jsonOf(in.Price), in.Sort, in.Active, by)
		if err == nil {
			if n, _ := res.RowsAffected(); n == 0 {
				s.fail(w, pluginkit.Missing("QUIZ_HINT_NOT_FOUND"))
				return
			}
		}
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	h, err := scanHint(s.db.QueryRowContext(ctx, `SELECT `+hintCols+` FROM quiz_hint h WHERE h.id=$1`, hid))
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, h)
}

func (s *service) hintDelete(w http.ResponseWriter, r *http.Request) {
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM quiz_hint WHERE id::text=$1`, pluginkit.PathVar(r, "hid"))
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("QUIZ_HINT_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) reportList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	page, limit := pluginkit.Page(r, 50, 200)
	args := &pluginkit.Args{}
	where := []string{"TRUE"}
	if v := q.Get("status"); v != "" {
		where = append(where, `r.status=`+args.Add(v))
	}
	if v := q.Get("questionId"); v != "" {
		where = append(where, `r.question_id::text=`+args.Add(v))
	}
	if v := q.Get("reasonKey"); v != "" {
		where = append(where, `r.reason_key=`+args.Add(v))
	}
	cond := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM quiz_question_report r WHERE `+cond, args.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	lim, off := args.Add(limit), args.Add((page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT r.id::text, r.user_id::text, COALESCE(u.username,''), r.question_id::text, q.question::text, q.is_active, COALESCE(k.key,''), r.reason_key, COALESCE(rr.name::text,'{}'),
		r.note, r.status, r.resolution, r.reviewed_by, r.reviewed_at, r.created_at, (SELECT count(*) FROM quiz_question_report o WHERE o.question_id=r.question_id AND o.status='open')
		FROM quiz_question_report r JOIN quiz_question q ON q.id=r.question_id LEFT JOIN quiz_pack k ON k.id=q.pack_id LEFT JOIN users u ON u.id=r.user_id
		LEFT JOIN quiz_report_reason rr ON rr.key=r.reason_key WHERE `+cond+` ORDER BY r.created_at DESC LIMIT `+lim+` OFFSET `+off, args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, user, username, qid, qtext, packKey, reason, reasonName, note, status, resolution, by string
		var active bool
		var reviewed sql.NullTime
		var created time.Time
		var open int
		if err := rows.Scan(&id, &user, &username, &qid, &qtext, &active, &packKey, &reason, &reasonName, &note, &status, &resolution, &by, &reviewed, &created, &open); err != nil {
			s.fail(w, err)
			return
		}
		qm, rn := map[string]string{}, map[string]string{}
		unmarshal([]byte(qtext), &qm)
		unmarshal([]byte(reasonName), &rn)
		items = append(items, map[string]any{"id": id, "userId": user, "username": username, "questionId": qid, "question": qm, "questionActive": active, "packKey": packKey,
			"reasonKey": reason, "reasonName": rn, "note": note, "status": status, "resolution": resolution, "reviewedBy": by, "reviewedAt": stamp(reviewed),
			"createdAt": created.UTC().Format(time.RFC3339), "openReports": open})
	}
	var openTotal int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM quiz_question_report WHERE status='open'`).Scan(&openTotal)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit, "open": openTotal})
}

func (s *service) reportReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in struct {
		Status          string `json:"status"`
		Resolution      string `json:"resolution"`
		SameQuestion    bool   `json:"sameQuestion"`
		DisableQuestion bool   `json:"disableQuestion"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if in.Status != "resolved" && in.Status != "dismissed" && in.Status != "open" {
		s.fail(w, bad("INVALID_STATUS", "status"))
		return
	}
	if utf8.RuneCountInString(in.Resolution) > 1000 {
		s.fail(w, bad("TEXT_TOO_LONG", "resolution"))
		return
	}
	by := pluginkit.ConsoleUser(r)
	var qid, user string
	if err := s.db.QueryRowContext(ctx, `SELECT question_id::text, user_id::text FROM quiz_question_report WHERE id::text=$1`, id).Scan(&qid, &user); err != nil {
		s.fail(w, pluginkit.Missing("QUIZ_REPORT_NOT_FOUND"))
		return
	}
	err := pluginkit.InTx(ctx, s.db, func(tx *sql.Tx) error {
		where, arg := `id::text=$1`, id
		if in.SameQuestion {
			where, arg = `question_id::text=$1 AND status='open'`, qid
		}
		if in.Status == "open" {
			if _, err := tx.ExecContext(ctx, `UPDATE quiz_question_report SET status='open', resolution=$2, reviewed_by=$3, reviewed_at=now() WHERE `+where, arg, in.Resolution, by); err != nil {
				return err
			}
		} else if _, err := tx.ExecContext(ctx, `UPDATE quiz_question_report SET status=$2, resolution=$3, reviewed_by=$4, reviewed_at=now() WHERE `+where, arg, in.Status, in.Resolution, by); err != nil {
			return err
		}
		if in.DisableQuestion {
			_, err := tx.ExecContext(ctx, `UPDATE quiz_question SET is_active=false, updated_by=$2, updated_at=now() WHERE id=$1`, qid, by)
			return err
		}
		return nil
	})
	if isUnique(err) {
		s.fail(w, conflict("ALREADY_EXISTS", "status"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: user, Kind: activity.KindReport, Action: "quiz_report_" + in.Status, RefID: id, Actor: by, Data: map[string]any{"questionId": qid}})
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"id": id, "status": in.Status})
}

func (s *service) reasonList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT id::text, key, name, sort, is_active, (SELECT count(*) FROM quiz_question_report x WHERE x.reason_key=rr.key) FROM quiz_report_reason rr ORDER BY sort, key`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, key string
		var raw []byte
		var sort, used int
		var active bool
		if err := rows.Scan(&id, &key, &raw, &sort, &active, &used); err != nil {
			s.fail(w, err)
			return
		}
		name := map[string]string{}
		unmarshal(raw, &name)
		items = append(items, map[string]any{"id": id, "key": key, "name": name, "sort": sort, "isActive": active, "used": used})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}

func (s *service) reasonSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in struct {
		Key    string            `json:"key"`
		Name   map[string]string `json:"name"`
		Sort   int               `json:"sort"`
		Active *bool             `json:"isActive"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	in.Key = strings.TrimSpace(in.Key)
	if id == "" && !settings.KeyPattern.MatchString(in.Key) {
		s.fail(w, bad("INVALID_KEY", "key"))
		return
	}
	var err error
	if in.Name, err = s.cleanText(ctx, "name", in.Name, true, 120); err != nil {
		s.fail(w, err)
		return
	}
	active := in.Active == nil || *in.Active
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
		_, err = s.db.ExecContext(ctx, `INSERT INTO quiz_report_reason (id, key, name, sort, is_active) VALUES ($1,$2,$3,$4,$5)`, id, in.Key, jsonOf(in.Name), in.Sort, active)
	} else {
		_, err = s.db.ExecContext(ctx, `UPDATE quiz_report_reason SET name=$2, sort=$3, is_active=$4, updated_at=now() WHERE id::text=$1`, id, jsonOf(in.Name), in.Sort, active)
	}
	if isUnique(err) {
		s.fail(w, conflict("ALREADY_EXISTS", "key"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	s.reasonList(w, r)
}

func (s *service) reasonDelete(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	var used bool
	_ = s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM quiz_question_report x JOIN quiz_report_reason rr ON rr.key=x.reason_key WHERE rr.id::text=$1)`, id).Scan(&used)
	if used {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "IN_USE", Message: "reports use this reason; switch it off instead"})
		return
	}
	if _, err := s.db.ExecContext(r.Context(), `DELETE FROM quiz_report_reason WHERE id::text=$1`, id); err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) templateList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+templateCols+templateFrom+` ORDER BY t.key`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []*resultTemplate{}
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, t)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items), "placeholders": placeholders, "secrets": s.secretKey != nil})
}

func (s *service) templateSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in struct {
		resultTemplate
		WebhookSecret *string `json:"webhookSecret"`
	}
	in.Active = true
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.fail(w, err)
		return
	}
	in.Key, in.Name, in.Body = strings.TrimSpace(in.Key), strings.TrimSpace(in.Name), strings.TrimSpace(in.Body)
	in.WonText, in.LostText, in.WebhookURL = strings.TrimSpace(in.WonText), strings.TrimSpace(in.LostText), strings.TrimSpace(in.WebhookURL)
	switch {
	case id == "" && !settings.KeyPattern.MatchString(in.Key):
		s.fail(w, bad("INVALID_KEY", "key"))
		return
	case in.Name == "" || utf8.RuneCountInString(in.Name) > 120:
		s.fail(w, bad("QUIZ_REPORT_TEMPLATE_NAME_IS_REQUIRED", "name"))
		return
	case in.Body == "" || utf8.RuneCountInString(in.Body) > 20000:
		s.fail(w, bad("QUIZ_REPORT_TEMPLATE_BODY_IS_REQUIRED", "body"))
		return
	case in.WebhookURL != "" && !webhookAllowed(in.WebhookURL):
		s.fail(w, bad("QUIZ_REPORT_WEBHOOK_URL_IS_INVALID", "webhookUrl"))
		return
	}
	if in.WonText == "" {
		in.WonText = "won"
	}
	if in.LostText == "" {
		in.LostText = "lost"
	}
	if in.MissionID == "" && in.MissionKey != "" {
		if in.MissionID = s.idByKey(ctx, "quiz_mission", in.MissionKey, ""); in.MissionID == "" {
			s.fail(w, bad("QUIZ_MISSION_NOT_FOUND", "missionKey"))
			return
		}
	}
	secretSet, sealed := false, ""
	if in.WebhookSecret != nil {
		secretSet = true
		if v := strings.TrimSpace(*in.WebhookSecret); v != "" {
			if s.secretKey == nil {
				s.fail(w, bad("QUIZ_SECRET_UNAVAILABLE", "webhookSecret"))
				return
			}
			var err error
			if sealed, err = webhookBox.Seal(s.secretKey, []byte(v)); err != nil {
				s.fail(w, err)
				return
			}
		}
	}
	by := pluginkit.ConsoleUser(r)
	var err error
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
		_, err = s.db.ExecContext(ctx, `INSERT INTO quiz_result_template (id, key, name, mission_id, body, won_text, lost_text, webhook_url, webhook_secret, is_active, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			id, in.Key, in.Name, nullable(in.MissionID), in.Body, in.WonText, in.LostText, in.WebhookURL, sealed, in.Active, by)
	} else {
		_, err = s.db.ExecContext(ctx, `UPDATE quiz_result_template SET name=$2, mission_id=$3, body=$4, won_text=$5, lost_text=$6, webhook_url=$7,
			webhook_secret=CASE WHEN $8 THEN $9 ELSE webhook_secret END, is_active=$10, updated_by=$11, updated_at=now() WHERE id::text=$1`,
			id, in.Name, nullable(in.MissionID), in.Body, in.WonText, in.LostText, in.WebhookURL, secretSet, sealed, in.Active, by)
	}
	if isUnique(err) {
		s.fail(w, conflict("ALREADY_EXISTS", "key"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	t, err := scanTemplate(s.db.QueryRowContext(ctx, `SELECT `+templateCols+templateFrom+` WHERE t.id::text=$1`, id))
	if err != nil {
		s.fail(w, pluginkit.Missing("QUIZ_REPORT_TEMPLATE_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, t)
}

func (s *service) templateDelete(w http.ResponseWriter, r *http.Request) {
	if _, err := s.db.ExecContext(r.Context(), `DELETE FROM quiz_result_template WHERE id::text=$1`, pluginkit.PathVar(r, "id")); err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) deliveryList(w http.ResponseWriter, r *http.Request) {
	page, limit := pluginkit.Page(r, 50, 200)
	rows, err := s.db.QueryContext(r.Context(), `SELECT id::text, user_id::text, url, status, attempts, last_error, next_attempt_at, delivered_at, created_at FROM quiz_result_delivery
		WHERE template_id::text=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, pluginkit.PathVar(r, "id"), limit, (page-1)*limit)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, user, u, status, last string
		var attempts int
		var next, created time.Time
		var delivered sql.NullTime
		if err := rows.Scan(&id, &user, &u, &status, &attempts, &last, &next, &delivered, &created); err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, map[string]any{"id": id, "userId": user, "url": u, "status": status, "attempts": attempts, "lastError": last,
			"nextAttemptAt": next.UTC().Format(time.RFC3339), "deliveredAt": stamp(delivered), "createdAt": created.UTC().Format(time.RFC3339)})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "page": page, "limit": limit})
}

func (s *service) stats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	days, _ := strconv.Atoi(q.Get("days"))
	if days < 1 || days > 365 {
		days = 30
	}
	langs, lang, _ := localized.ForUser(ctx, s.db, "", q.Get("lang"))
	rows, err := s.db.QueryContext(ctx, `SELECT k.id::text, k.key, k.name::text, COALESCE(m.key,''), COUNT(DISTINCT x.user_id), COUNT(x.id), COUNT(x.id) FILTER (WHERE x.status IN ('paying','done')),
		COUNT(x.id) FILTER (WHERE x.passed), COALESCE(AVG(x.score) FILTER (WHERE x.status IN ('paying','done')),0)::float8,
		COALESCE(AVG(EXTRACT(EPOCH FROM (x.completed_at - x.started_at))) FILTER (WHERE x.status IN ('paying','done')),0)::float8
		FROM quiz_pack k LEFT JOIN quiz_mission m ON m.id=k.mission_id LEFT JOIN quiz_session x ON x.pack_id=k.id AND x.started_at > now() - make_interval(days => $1)
		GROUP BY k.id, m.key, m.sort ORDER BY m.sort NULLS FIRST, k.sort, k.key`, days)
	if err != nil {
		s.fail(w, err)
		return
	}
	packs := []map[string]any{}
	totals := map[string]float64{}
	for rows.Next() {
		var id, key, name, mkey string
		var players, plays, finished, passes int
		var avg, secs float64
		if err := rows.Scan(&id, &key, &name, &mkey, &players, &plays, &finished, &passes, &avg, &secs); err != nil {
			rows.Close()
			s.fail(w, err)
			return
		}
		rate := 0.0
		if finished > 0 {
			rate = math.Round(float64(passes)*10000/float64(finished)) / 100
		}
		totals["plays"] += float64(plays)
		totals["finished"] += float64(finished)
		totals["passes"] += float64(passes)
		packs = append(packs, map[string]any{"id": id, "key": key, "name": resolveName(langs, lang, name, key), "missionKey": mkey, "players": players, "plays": plays,
			"finished": finished, "passes": passes, "passRate": rate, "avgScore": math.Round(avg*100) / 100, "avgSeconds": math.Round(secs*10) / 10})
	}
	rows.Close()
	var players int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT user_id) FROM quiz_session WHERE started_at > now() - make_interval(days => $1)`, days).Scan(&players)
	totals["players"] = float64(players)
	hard := []map[string]any{}
	rows, err = s.db.QueryContext(ctx, `SELECT a.question_id::text, q.question::text, COALESCE(k.key,''), COUNT(*), COUNT(*) FILTER (WHERE a.outcome='correct'), COUNT(*) FILTER (WHERE a.outcome IN ('skip','late'))
		FROM quiz_answer a JOIN quiz_question q ON q.id=a.question_id LEFT JOIN quiz_pack k ON k.id=q.pack_id WHERE a.answered_at > now() - make_interval(days => $1)
		GROUP BY a.question_id, q.question::text, k.key HAVING COUNT(*) >= 3 ORDER BY COUNT(*) FILTER (WHERE a.outcome='correct')::float / COUNT(*) ASC, COUNT(*) DESC LIMIT 20`, days)
	if err == nil {
		for rows.Next() {
			var id, raw, pk string
			var asked, right, skipped int
			if rows.Scan(&id, &raw, &pk, &asked, &right, &skipped) == nil {
				hard = append(hard, map[string]any{"id": id, "question": resolveName(langs, lang, raw, id), "packKey": pk, "answers": asked, "correct": right, "skipped": skipped,
					"correctRate": math.Round(float64(right)*10000/float64(asked)) / 100})
			}
		}
		rows.Close()
	}
	var hints, reports int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM quiz_user_hint WHERE status='done' AND created_at > now() - make_interval(days => $1)`, days).Scan(&hints)
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM quiz_question_report WHERE status='open'`).Scan(&reports)
	totals["hintsBought"], totals["openReports"] = float64(hints), float64(reports)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"days": days, "packs": packs, "hardest": hard, "totals": totals})
}

func (s *service) playerGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(userID); err != nil {
		s.fail(w, pluginkit.Missing("PLAYER_NOT_FOUND"))
		return
	}
	langs, lang, _ := localized.ForUser(ctx, s.db, "", r.URL.Query().Get("lang"))
	packs := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT k.id::text, k.key, k.name::text, u.plays, u.wins, u.best_score::float8, u.last_score::float8, u.rewards, u.first_pass_at, u.last_played_at
		FROM quiz_user_pack u JOIN quiz_pack k ON k.id=u.pack_id WHERE u.user_id=$1 ORDER BY u.last_played_at DESC NULLS LAST LIMIT 200`, userID)
	if err != nil {
		s.fail(w, err)
		return
	}
	for rows.Next() {
		var id, key, name string
		var plays, wins, rewards int
		var best, last float64
		var fp, lp sql.NullTime
		if rows.Scan(&id, &key, &name, &plays, &wins, &best, &last, &rewards, &fp, &lp) == nil {
			packs = append(packs, map[string]any{"id": id, "key": key, "name": resolveName(langs, lang, name, key), "plays": plays, "wins": wins, "bestScore": best, "lastScore": last,
				"rewards": rewards, "firstPassAt": stamp(fp), "lastPlayedAt": stamp(lp)})
		}
	}
	rows.Close()
	missions := []map[string]any{}
	rows, err = s.db.QueryContext(ctx, `SELECT m.key, m.kind, m.name::text, u.started_at, u.completed_at, u.rewarded FROM quiz_user_mission u JOIN quiz_mission m ON m.id=u.mission_id WHERE u.user_id=$1 ORDER BY u.started_at DESC`, userID)
	if err == nil {
		for rows.Next() {
			var key, kind, name string
			var started time.Time
			var completed sql.NullTime
			var rewarded bool
			if rows.Scan(&key, &kind, &name, &started, &completed, &rewarded) == nil {
				missions = append(missions, map[string]any{"key": key, "kind": kind, "name": resolveName(langs, lang, name, key), "startedAt": started.UTC().Format(time.RFC3339),
					"completedAt": stamp(completed), "rewarded": rewarded})
			}
		}
		rows.Close()
	}
	sessions := []map[string]any{}
	rows, err = s.db.QueryContext(ctx, `SELECT x.id::text, k.key, x.status, x.started_at, x.completed_at, x.answered, x.correct, jsonb_array_length(x.questions), x.score::float8, x.passed
		FROM quiz_session x JOIN quiz_pack k ON k.id=x.pack_id WHERE x.user_id=$1 ORDER BY x.started_at DESC LIMIT 10`, userID)
	if err == nil {
		for rows.Next() {
			var id, key, status string
			var started time.Time
			var completed sql.NullTime
			var answered, correct, total int
			var score sql.NullFloat64
			var passed sql.NullBool
			if rows.Scan(&id, &key, &status, &started, &completed, &answered, &correct, &total, &score, &passed) == nil {
				sessions = append(sessions, map[string]any{"id": id, "packKey": key, "status": status, "startedAt": started.UTC().Format(time.RFC3339), "completedAt": stamp(completed),
					"answered": answered, "correct": correct, "total": total, "score": score.Float64, "passed": passed.Bool})
			}
		}
		rows.Close()
	}
	var hints, reports int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM quiz_user_hint WHERE user_id=$1 AND status='done'`, userID).Scan(&hints)
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM quiz_question_report WHERE user_id=$1`, userID).Scan(&reports)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"packs": packs, "missions": missions, "sessions": sessions, "hintsBought": hints, "reports": reports})
}

func (s *service) playerReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, packID := pluginkit.PathVar(r, "id"), pluginkit.PathVar(r, "packId")
	if _, err := uuid.FromString(userID); err != nil {
		s.fail(w, pluginkit.Missing("PLAYER_NOT_FOUND"))
		return
	}
	full := r.URL.Query().Get("rewards") == "reset"
	err := pluginkit.InTx(ctx, s.db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `UPDATE quiz_session SET status='closed', completed_at=now() WHERE user_id=$1 AND pack_id::text=$2 AND status='open'`, userID, packID); err != nil {
			return err
		}
		if full {
			_, err := tx.ExecContext(ctx, `DELETE FROM quiz_user_pack WHERE user_id=$1 AND pack_id::text=$2`, userID, packID)
			return err
		}
		_, err := tx.ExecContext(ctx, `UPDATE quiz_user_pack SET plays=0, wins=0, best_score=0, last_score=0, last_correct=0, last_total=0, first_pass_at=NULL, last_played_at=NULL WHERE user_id=$1 AND pack_id::text=$2`, userID, packID)
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindConfig, Action: "quiz_reset", RefID: packID, Actor: pluginkit.ConsoleUser(r), Data: map[string]any{"rewards": full}})
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"reset": true, "rewards": full})
}
