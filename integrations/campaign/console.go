package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
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
		{p + "/chapters", http.MethodGet, s.chapterList},
		{p + "/chapters", http.MethodPost, s.chapterSave},
		{p + "/chapters/order", http.MethodPut, s.chapterOrder},
		{p + "/chapters/{id}", http.MethodGet, s.chapterGet},
		{p + "/chapters/{id}", http.MethodPut, s.chapterSave},
		{p + "/chapters/{id}", http.MethodDelete, s.chapterDelete},
		{p + "/stages", http.MethodGet, s.stageList},
		{p + "/stages", http.MethodPost, s.stageSave},
		{p + "/stages/order", http.MethodPut, s.stageOrder},
		{p + "/stages/{id}", http.MethodGet, s.stageGet},
		{p + "/stages/{id}", http.MethodPut, s.stageSave},
		{p + "/stages/{id}", http.MethodDelete, s.stageDelete},
		{p + "/freeride", http.MethodGet, s.freerideList},
		{p + "/freeride", http.MethodPost, s.freerideSave},
		{p + "/freeride/{id}", http.MethodPut, s.freerideSave},
		{p + "/freeride/{id}", http.MethodDelete, s.freerideDelete},
		{p + "/players/{id}", http.MethodGet, s.playerGet},
		{p + "/players/{id}/stages/{stageId}", http.MethodDelete, s.playerReset},
		{p + "/report", http.MethodGet, s.report},
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

func bad(code, field string) error { return pluginkit.Bad(code, field, "") }

func (s *service) settingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, meta, err := settings.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	_, lvErr := s.level(r.Context(), uuid.Nil.String(), "")
	metaOK := cfg.XP.MetaKey == ""
	if !metaOK {
		_ = s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM economy_meta WHERE key=$1)`, cfg.XP.MetaKey).Scan(&metaOK)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"settings": cfg, "defaults": settings.Defaults(), "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt,
		"economy": lvErr == nil || !errors.Is(lvErr, economyhooks.ErrNoLevels), "prizes": economyhooks.PrizesAvailable(), "xpMetaFound": metaOK})
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
	if err := settings.Save(r.Context(), s.db, in, pluginkit.ConsoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	s.settingsGet(w, r)
}

func (s *service) refs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	langs, lang, _ := localized.ForUser(ctx, s.db, "", r.URL.Query().Get("lang"))
	out := map[string]any{}
	for name, q := range map[string]string{
		"currencies":  `SELECT key, name::text FROM economy_currency ORDER BY key`,
		"metas":       `SELECT key, name::text FROM economy_meta ORDER BY key`,
		"items":       `SELECT key, name::text FROM economy_item ORDER BY key LIMIT 2000`,
		"prizes":      `SELECT key, name::text FROM economy_prize ORDER BY key LIMIT 2000`,
		"prizeGroups": `SELECT key, name::text FROM economy_prize_group ORDER BY key`,
		"levelTables": `SELECT key, name::text FROM economy_level_table ORDER BY sort, key`,
		"stages":      `SELECT key, name::text FROM campaign_stage ORDER BY key`,
		"chapters":    `SELECT key, name::text FROM campaign_chapter ORDER BY sort, key`,
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
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func resolveName(langs *localized.Set, lang, raw, fallback string) string {
	if langs != nil {
		if v := langs.Resolve(localized.Parse(raw), lang); v != "" {
			return v
		}
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
	return nil
}

func cleanRules(field string, rules []Rule) ([]Rule, error) {
	if len(rules) > 10 {
		return nil, bad("TOO_MANY_RULES", field)
	}
	out := make([]Rule, 0, len(rules))
	for i, r := range rules {
		f := field + "[" + strconv.Itoa(i) + "]"
		r.Type, r.Key, r.Table = strings.TrimSpace(r.Type), strings.TrimSpace(r.Key), strings.TrimSpace(r.Table)
		if !ruleTypes[r.Type] {
			return nil, bad("INVALID_RULE", f+".type")
		}
		if r.Type != "level" && !settings.KeyPattern.MatchString(r.Key) {
			return nil, bad("INVALID_KEY", f+".key")
		}
		if r.Type == "level" {
			r.Key = ""
			if r.Table != "" && !settings.KeyPattern.MatchString(r.Table) {
				return nil, bad("INVALID_KEY", f+".table")
			}
		} else {
			r.Table = ""
		}
		if r.Value < 0 || r.Value > 1e9 || math.IsNaN(r.Value) || (r.Type == "chapter" && r.Value > 100) {
			return nil, bad("VALUE_OUT_OF_RANGE", f+".value")
		}
		if r.Type == "item" {
			r.Value = 0
		}
		out = append(out, r)
	}
	return out, nil
}

func cleanAddress(field, v string) (string, error) {
	v = strings.TrimSpace(v)
	if utf8.RuneCountInString(v) > 500 {
		return "", bad("TEXT_TOO_LONG", field)
	}
	return v, nil
}

func jsonOf(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func isUnique(err error) bool {
	var pg interface{ SQLState() string }
	return errors.As(err, &pg) && pg.SQLState() == "23505"
}

func isFK(err error) bool {
	var pg interface{ SQLState() string }
	return errors.As(err, &pg) && pg.SQLState() == "23503"
}

func (s *service) chapterList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+chapterCols+` FROM campaign_chapter c ORDER BY c.sort, c.key`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []*chapter{}
	for rows.Next() {
		c, err := scanChapter(rows)
		if err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, c)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}

func (s *service) chapterGet(w http.ResponseWriter, r *http.Request) {
	c, err := scanChapter(s.db.QueryRowContext(r.Context(), `SELECT `+chapterCols+` FROM campaign_chapter c WHERE c.id::text=$1`, pluginkit.PathVar(r, "id")))
	if errors.Is(err, sql.ErrNoRows) {
		s.fail(w, pluginkit.Missing("CHAPTER_NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, c)
}

func (s *service) chapterSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in chapter
	if err := pluginkit.ReadBody(r, &in, 256*1024); err != nil {
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
	if in.Description, err = s.cleanText(ctx, "description", in.Description, false, 2000); err != nil {
		s.fail(w, err)
		return
	}
	if in.IconURL, err = cleanAddress("iconUrl", in.IconURL); err != nil {
		s.fail(w, err)
		return
	}
	if in.Address, err = cleanAddress("address", in.Address); err != nil {
		s.fail(w, err)
		return
	}
	if in.Unlock, err = cleanRules("unlock", in.Unlock); err != nil {
		s.fail(w, err)
		return
	}
	if len(in.Milestones) > 20 {
		s.fail(w, bad("TOO_MANY_MILESTONES", "milestones"))
		return
	}
	seen := map[float64]bool{}
	for i := range in.Milestones {
		f := "milestones[" + strconv.Itoa(i) + "]"
		m := &in.Milestones[i]
		if m.Percent <= 0 || m.Percent > 100 {
			s.fail(w, bad("VALUE_OUT_OF_RANGE", f+".percent"))
			return
		}
		if seen[m.Percent] {
			s.fail(w, bad("DUPLICATE_MILESTONE", f+".percent"))
			return
		}
		seen[m.Percent] = true
		if err := cleanReward(f, &m.Reward); err != nil {
			s.fail(w, err)
			return
		}
		if m.Reward.empty() {
			s.fail(w, bad("REWARD_REQUIRED", f))
			return
		}
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	by := pluginkit.ConsoleUser(r)
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
		_, err = s.db.ExecContext(ctx, `INSERT INTO campaign_chapter (id, key, name, description, icon_url, address, sort, is_active, unlock, milestones, metadata, updated_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, id, in.Key, jsonOf(in.Name), jsonOf(in.Description), in.IconURL, in.Address, in.Sort, in.Active,
			jsonOf(in.Unlock), jsonOf(in.Milestones), jsonOf(in.Metadata), by)
	} else {
		var res sql.Result
		res, err = s.db.ExecContext(ctx, `UPDATE campaign_chapter SET name=$2, description=$3, icon_url=$4, address=$5, sort=$6, is_active=$7, unlock=$8, milestones=$9, metadata=$10, updated_by=$11, updated_at=now()
			WHERE id::text=$1`, id, jsonOf(in.Name), jsonOf(in.Description), in.IconURL, in.Address, in.Sort, in.Active, jsonOf(in.Unlock), jsonOf(in.Milestones), jsonOf(in.Metadata), by)
		if err == nil {
			if n, _ := res.RowsAffected(); n == 0 {
				s.fail(w, pluginkit.Missing("CHAPTER_NOT_FOUND"))
				return
			}
		}
	}
	if isUnique(err) {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "ALREADY_EXISTS", Field: "key"})
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	r.SetPathValue("id", id)
	s.chapterGet(w, r)
}

func (s *service) chapterOrder(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs []string `json:"ids"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	err := pluginkit.InTx(r.Context(), s.db, func(tx *sql.Tx) error {
		for i, id := range in.IDs {
			if _, err := tx.ExecContext(r.Context(), `UPDATE campaign_chapter SET sort=$2, updated_at=now() WHERE id::text=$1`, id, (i+1)*10); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	s.chapterList(w, r)
}

func (s *service) chapterDelete(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM campaign_chapter WHERE id::text=$1`, id)
	if isFK(err) {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "IN_USE", Message: "the chapter still has stages"})
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("CHAPTER_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

type stageBody struct {
	stage
	Unlock      *[]Rule              `json:"unlock"`
	StarRules   *[]settings.StarRule `json:"starRules"`
	Multipliers *[]float64           `json:"starMultipliers"`
}

func (s *service) stageList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	page, limit := pluginkit.Page(r, 50, 200)
	args := &pluginkit.Args{}
	where := []string{"TRUE"}
	if v := strings.TrimSpace(q.Get("chapterId")); v != "" {
		where = append(where, `s.chapter_id::text=`+args.Add(v))
	}
	if v := strings.TrimSpace(q.Get("mode")); v != "" {
		where = append(where, `s.mode=`+args.Add(v))
	}
	if v := strings.TrimSpace(q.Get("q")); v != "" {
		like := args.Add(pluginkit.LikePattern(v))
		where = append(where, `(s.key ILIKE `+like+` OR s.name::text ILIKE `+like+`)`)
	}
	cond := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM campaign_stage s WHERE `+cond, args.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	lim, off := args.Add(limit), args.Add((page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT `+stageCols+`, (SELECT count(*) FROM campaign_user_stage u WHERE u.stage_id=s.id) FROM campaign_stage s JOIN campaign_chapter c ON c.id=s.chapter_id
		WHERE `+cond+` ORDER BY c.sort, c.key, s.sort, s.key LIMIT `+lim+` OFFSET `+off, args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var players int
		var st *stage
		st, err = scanStage(scanPlus{rows, &players})
		if err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, map[string]any{"stage": st, "players": players})
	}
	out := make([]any, 0, len(items))
	for _, it := range items {
		raw, _ := json.Marshal(it["stage"])
		m := map[string]any{}
		_ = json.Unmarshal(raw, &m)
		m["players"] = it["players"]
		out = append(out, m)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": out, "total": total, "page": page, "limit": limit})
}

type scanPlus struct {
	rows  *sql.Rows
	extra *int
}

func (p scanPlus) Scan(dest ...any) error { return p.rows.Scan(append(dest, p.extra)...) }

func (s *service) stageGet(w http.ResponseWriter, r *http.Request) {
	st, err := scanStage(s.db.QueryRowContext(r.Context(), `SELECT `+stageCols+` FROM campaign_stage s JOIN campaign_chapter c ON c.id=s.chapter_id WHERE s.id::text=$1`, pluginkit.PathVar(r, "id")))
	if errors.Is(err, sql.ErrNoRows) {
		s.fail(w, pluginkit.Missing("STAGE_NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, st)
}

func (s *service) stageSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in stageBody
	if err := pluginkit.ReadBody(r, &in, 256*1024); err != nil {
		s.fail(w, err)
		return
	}
	cfg := s.settings(ctx)
	in.Key = strings.TrimSpace(in.Key)
	if id == "" && !settings.KeyPattern.MatchString(in.Key) {
		s.fail(w, bad("INVALID_KEY", "key"))
		return
	}
	if in.ChapterID == "" && in.ChapterKey != "" {
		_ = s.db.QueryRowContext(ctx, `SELECT id::text FROM campaign_chapter WHERE key=$1`, in.ChapterKey).Scan(&in.ChapterID)
	}
	var exists bool
	if _, err := uuid.FromString(in.ChapterID); err == nil {
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM campaign_chapter WHERE id=$1)`, in.ChapterID).Scan(&exists)
	}
	if !exists {
		s.fail(w, bad("CHAPTER_NOT_FOUND", "chapterId"))
		return
	}
	if in.Mode == "" {
		in.Mode = settings.ModeOrdered
	}
	if in.Mode != settings.ModeOrdered && in.Mode != settings.ModeRepeatable && in.Mode != settings.ModeOnce {
		s.fail(w, bad("INVALID_MODE", "mode"))
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
	if in.IconURL, err = cleanAddress("iconUrl", in.IconURL); err != nil {
		s.fail(w, err)
		return
	}
	if in.Address, err = cleanAddress("address", in.Address); err != nil {
		s.fail(w, err)
		return
	}
	var unlock, rules, mult []byte
	if in.Unlock != nil {
		list, err := cleanRules("unlock", *in.Unlock)
		if err != nil {
			s.fail(w, err)
			return
		}
		unlock = jsonOf(list)
	}
	if in.StarRules != nil {
		if f, code := settings.ValidateRules(*in.StarRules, cfg.Stars.MaxStars); code != "" {
			s.fail(w, bad(code, "starRules"+f))
			return
		}
		rules = jsonOf(*in.StarRules)
	}
	if in.Multipliers != nil {
		if !settings.ValidateMultipliers(*in.Multipliers, cfg.Stars.MaxStars) {
			s.fail(w, bad("INVALID_MULTIPLIERS", "starMultipliers"))
			return
		}
		mult = jsonOf(*in.Multipliers)
	}
	if in.TimeLimit != nil && (*in.TimeLimit <= 0 || *in.TimeLimit > 86400) {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "timeLimitSeconds"))
		return
	}
	if in.MinSeconds != nil && (*in.MinSeconds < 0 || *in.MinSeconds > 86400 || in.TimeLimit != nil && *in.MinSeconds >= *in.TimeLimit) {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "minSeconds"))
		return
	}
	in.Rewards.PrizeKey, in.Rewards.PrizeGroupKey = "", ""
	if in.TierPrizes == nil {
		in.TierPrizes = []TierPrize{}
	}
	if err := cleanReward("rewards", &in.Rewards); err != nil {
		s.fail(w, err)
		return
	}
	if in.XP < 0 || in.XP > 1e9 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "xp"))
		return
	}
	if len(in.TierPrizes) > 10 {
		s.fail(w, bad("TOO_MANY_PRIZES", "tierPrizes"))
		return
	}
	for i := range in.TierPrizes {
		t := &in.TierPrizes[i]
		f := "tierPrizes[" + strconv.Itoa(i) + "]"
		if t.Stars < 1 || t.Stars > cfg.Stars.MaxStars {
			s.fail(w, bad("VALUE_OUT_OF_RANGE", f+".stars"))
			return
		}
		rw := Reward{PrizeKey: t.PrizeKey, PrizeGroupKey: t.PrizeGroupKey}
		if err := cleanReward(f, &rw); err != nil {
			s.fail(w, err)
			return
		}
		if rw.PrizeKey == "" && rw.PrizeGroupKey == "" {
			s.fail(w, bad("REWARD_REQUIRED", f))
			return
		}
		t.PrizeKey, t.PrizeGroupKey = rw.PrizeKey, rw.PrizeGroupKey
	}
	in.RewardGroup = strings.TrimSpace(in.RewardGroup)
	if in.RewardGroup != "" && !settings.KeyPattern.MatchString(in.RewardGroup) {
		s.fail(w, bad("INVALID_KEY", "rewardGroup"))
		return
	}
	if in.Cooldown < 0 || in.Cooldown > 30*86400 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "cooldownSeconds"))
		return
	}
	if in.DailyLimit < 0 || in.DailyLimit > 10000 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "dailyLimit"))
		return
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	by := pluginkit.ConsoleUser(r)
	args := []any{jsonOf(in.Name), jsonOf(in.Description), in.IconURL, in.Address, in.Sort, in.Active, in.Mode, unlock, in.TimeLimit, in.MinSeconds, rules, mult,
		jsonOf(in.Rewards), in.XP, jsonOf(in.TierPrizes), in.RewardGroup, in.Cooldown, in.DailyLimit, in.XPWhenLimited, jsonOf(in.Metadata), by, in.ChapterID}
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
		_, err = s.db.ExecContext(ctx, `INSERT INTO campaign_stage (name, description, icon_url, address, sort, is_active, mode, unlock, time_limit_seconds, min_seconds, star_rules, star_multipliers,
			rewards, xp, tier_prizes, reward_group, cooldown_seconds, daily_limit, xp_when_limited, metadata, updated_by, chapter_id, id, key)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)`, append(args, id, in.Key)...)
	} else {
		var res sql.Result
		res, err = s.db.ExecContext(ctx, `UPDATE campaign_stage SET name=$1, description=$2, icon_url=$3, address=$4, sort=$5, is_active=$6, mode=$7, unlock=$8, time_limit_seconds=$9,
			min_seconds=$10, star_rules=$11, star_multipliers=$12, rewards=$13, xp=$14, tier_prizes=$15, reward_group=$16, cooldown_seconds=$17, daily_limit=$18, xp_when_limited=$19,
			metadata=$20, updated_by=$21, chapter_id=$22, updated_at=now() WHERE id::text=$23`, append(args, id)...)
		if err == nil {
			if n, _ := res.RowsAffected(); n == 0 {
				s.fail(w, pluginkit.Missing("STAGE_NOT_FOUND"))
				return
			}
		}
	}
	if isUnique(err) {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "ALREADY_EXISTS", Field: "key"})
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	r.SetPathValue("id", id)
	s.stageGet(w, r)
}

func (s *service) stageOrder(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs []string `json:"ids"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	err := pluginkit.InTx(r.Context(), s.db, func(tx *sql.Tx) error {
		for i, id := range in.IDs {
			if _, err := tx.ExecContext(r.Context(), `UPDATE campaign_stage SET sort=$2, updated_at=now() WHERE id::text=$1`, id, (i+1)*10); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *service) stageDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var used bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM campaign_user_stage WHERE stage_id::text=$1)`, id).Scan(&used)
	if used {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "IN_USE", Message: "players have progress on this stage; turn it off instead"})
		return
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM campaign_stage WHERE id::text=$1`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("STAGE_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) freerideList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+freerideCols+` FROM campaign_freeride ORDER BY sort, key`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []*freeride{}
	for rows.Next() {
		f, err := scanFreeride(rows)
		if err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, f)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}

func (s *service) freerideSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in freeride
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
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
	if in.Description, err = s.cleanText(ctx, "description", in.Description, false, 2000); err != nil {
		s.fail(w, err)
		return
	}
	if in.IconURL, err = cleanAddress("iconUrl", in.IconURL); err != nil {
		s.fail(w, err)
		return
	}
	if err := cleanReward("rewards", &in.Rewards); err != nil {
		s.fail(w, err)
		return
	}
	if in.Rewards.empty() && in.XP == 0 {
		s.fail(w, bad("REWARD_REQUIRED", "rewards"))
		return
	}
	in.RewardGroup = strings.TrimSpace(in.RewardGroup)
	switch {
	case in.RewardGroup != "" && !settings.KeyPattern.MatchString(in.RewardGroup):
		err = bad("INVALID_KEY", "rewardGroup")
	case in.XP < 0 || in.XP > 1e9:
		err = bad("VALUE_OUT_OF_RANGE", "xp")
	case in.Cooldown < 0 || in.Cooldown > 30*86400:
		err = bad("VALUE_OUT_OF_RANGE", "cooldownSeconds")
	case in.DailyLimit < 0 || in.DailyLimit > 10000:
		err = bad("VALUE_OUT_OF_RANGE", "dailyLimit")
	case in.MinLevel < 0 || in.MinLevel > 10000:
		err = bad("VALUE_OUT_OF_RANGE", "minLevel")
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	cur := Reward{Currencies: in.Rewards.Currencies}
	args := []any{jsonOf(in.Name), jsonOf(in.Description), in.IconURL, in.Sort, in.Active, jsonOf(cur), in.Rewards.PrizeKey, in.Rewards.PrizeGroupKey, in.XP, in.RewardGroup,
		in.Cooldown, in.DailyLimit, in.MinLevel, pluginkit.ConsoleUser(r)}
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
		_, err = s.db.ExecContext(ctx, `INSERT INTO campaign_freeride (name, description, icon_url, sort, is_active, rewards, prize_key, prize_group_key, xp, reward_group, cooldown_seconds, daily_limit, min_level, updated_by, id, key)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`, append(args, id, in.Key)...)
	} else {
		var res sql.Result
		res, err = s.db.ExecContext(ctx, `UPDATE campaign_freeride SET name=$1, description=$2, icon_url=$3, sort=$4, is_active=$5, rewards=$6, prize_key=$7, prize_group_key=$8, xp=$9, reward_group=$10,
			cooldown_seconds=$11, daily_limit=$12, min_level=$13, updated_by=$14, updated_at=now() WHERE id::text=$15`, append(args, id)...)
		if err == nil {
			if n, _ := res.RowsAffected(); n == 0 {
				s.fail(w, pluginkit.Missing("FREERIDE_NOT_FOUND"))
				return
			}
		}
	}
	if isUnique(err) {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "ALREADY_EXISTS", Field: "key"})
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	f, err := scanFreeride(s.db.QueryRowContext(ctx, `SELECT `+freerideCols+` FROM campaign_freeride WHERE id=$1`, id))
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, f)
}

func (s *service) freerideDelete(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM campaign_freeride WHERE id::text=$1`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("FREERIDE_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) playerGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(userID); err != nil {
		s.fail(w, pluginkit.Missing("PLAYER_NOT_FOUND"))
		return
	}
	cfg := s.settings(ctx)
	cat, err := loadCatalog(ctx, s.db, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	p, err := s.loadPlayer(ctx, s.db, userID, cfg, cat)
	if err != nil {
		s.fail(w, err)
		return
	}
	langs, lang, _ := localized.ForUser(ctx, s.db, "", r.URL.Query().Get("lang"))
	chapters := []map[string]any{}
	for _, ch := range cat.chapters {
		earned, max, pct := p.chapterStars(ch)
		stages := []map[string]any{}
		for _, st := range cat.stagesOf(ch.ID) {
			row, ok := p.rows[st.ID]
			if !ok {
				row = p.row(st.ID)
			}
			stages = append(stages, map[string]any{"id": st.ID, "key": st.Key, "name": text(langs, lang, st.Name), "mode": st.Mode, "isActive": st.Active, "played": ok,
				"unlocked": len(p.stageLocks(ctx, st)) == 0, "bestStars": row.BestStars, "completions": row.Completions, "wins": row.Wins, "bestSeconds": row.BestSeconds,
				"claimed": row.Claimed, "claimedXp": row.ClaimedXP, "lastCompletedAt": iso(row.LastAt)})
		}
		chapters = append(chapters, map[string]any{"id": ch.ID, "key": ch.Key, "name": text(langs, lang, ch.Name), "stars": earned, "maxStars": max, "starPercent": math.Round(pct*100) / 100, "stages": stages})
	}
	attempts := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT a.id::text, s.key, a.status, a.started_at, a.completed_at, a.success, a.stars, a.rejected, a.last_reject FROM campaign_attempt a JOIN campaign_stage s ON s.id=a.stage_id
		WHERE a.user_id=$1 ORDER BY a.started_at DESC LIMIT 10`, userID)
	if err == nil {
		for rows.Next() {
			var id, key, status, rej string
			var started time.Time
			var completed sql.NullTime
			var success sql.NullBool
			var stars, rejected int
			if rows.Scan(&id, &key, &status, &started, &completed, &success, &stars, &rejected, &rej) == nil {
				var c *time.Time
				if completed.Valid {
					c = &completed.Time
				}
				attempts = append(attempts, map[string]any{"id": id, "stageKey": key, "status": status, "startedAt": iso(&started), "completedAt": iso(c), "success": success.Bool, "stars": stars, "rejected": rejected, "lastReject": rej})
			}
		}
		rows.Close()
	}
	xp, level := s.xpLevel(ctx, p)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"xp": xp, "level": level, "chapters": chapters, "attempts": attempts})
}

func (s *service) playerReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, stageID := pluginkit.PathVar(r, "id"), pluginkit.PathVar(r, "stageId")
	var res sql.Result
	var err error
	if r.URL.Query().Get("claims") == "reset" {
		res, err = s.db.ExecContext(ctx, `DELETE FROM campaign_user_stage WHERE user_id::text=$1 AND stage_id::text=$2`, userID, stageID)
	} else {
		res, err = s.db.ExecContext(ctx, `UPDATE campaign_user_stage SET best_stars=0, completions=0, wins=0, best_seconds=NULL, first_win_at=NULL, last_completed_at=NULL, updated_at=now()
			WHERE user_id::text=$1 AND stage_id::text=$2`, userID, stageID)
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("STAGE_NOT_FOUND"))
		return
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE campaign_attempt SET status='closed', closed_at=now() WHERE user_id::text=$1 AND stage_id::text=$2 AND status='open'`, userID, stageID)
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindEvent, Action: "stage_reset", RefID: stageID, Actor: "admin", Data: map[string]any{"stageId": stageID, "admin": pluginkit.ConsoleUser(r), "claims": r.URL.Query().Get("claims")}})
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"reset": true})
}

func (s *service) report(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	days, _ := strconv.Atoi(q.Get("days"))
	if days < 1 || days > 365 {
		days = 30
	}
	args := []any{days}
	cond := ""
	if v := strings.TrimSpace(q.Get("chapterId")); v != "" {
		args = append(args, v)
		cond = ` AND s.chapter_id::text=$2`
	}
	rows, err := s.db.QueryContext(ctx, `SELECT s.id::text, s.key, s.name::text, c.key, s.mode,
		COUNT(DISTINCT a.user_id), COUNT(a.id), COUNT(a.id) FILTER (WHERE a.success), COUNT(DISTINCT a.user_id) FILTER (WHERE a.success), COALESCE(SUM(a.rejected),0),
		COALESCE((SELECT AVG(u.best_stars)::float8 FROM campaign_user_stage u WHERE u.stage_id=s.id AND u.wins > 0),0),
		COALESCE((SELECT AVG(a2.stars)::float8 FROM campaign_attempt a2 WHERE a2.stage_id=s.id AND a2.success AND a2.started_at > now() - make_interval(days => $1)),0)
		FROM campaign_stage s JOIN campaign_chapter c ON c.id=s.chapter_id
		LEFT JOIN campaign_attempt a ON a.stage_id=s.id AND a.started_at > now() - make_interval(days => $1)
		WHERE TRUE`+cond+` GROUP BY s.id, c.sort, c.key ORDER BY c.sort, c.key, s.sort, s.key`, args...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	langs, lang, _ := localized.ForUser(ctx, s.db, "", q.Get("lang"))
	items := []map[string]any{}
	prevWinners := map[string]int{}
	totals := map[string]float64{}
	for rows.Next() {
		var id, key, name, chKey, mode string
		var players, attempts, wins, winners, rejected int
		var avgBest, avgStars float64
		if err := rows.Scan(&id, &key, &name, &chKey, &mode, &players, &attempts, &wins, &winners, &rejected, &avgBest, &avgStars); err != nil {
			s.fail(w, err)
			return
		}
		rate := 0.0
		if players > 0 {
			rate = float64(winners) * 100 / float64(players)
		}
		var dropOff any
		if mode == settings.ModeOrdered {
			if prev, ok := prevWinners[chKey]; ok && prev > 0 {
				dropOff = math.Round(math.Max(0, float64(prev-winners))*10000/float64(prev)) / 100
			}
			prevWinners[chKey] = winners
		}
		totals["players"] += float64(players)
		totals["attempts"] += float64(attempts)
		totals["wins"] += float64(wins)
		totals["rejected"] += float64(rejected)
		items = append(items, map[string]any{"id": id, "key": key, "name": resolveName(langs, lang, name, key), "chapterKey": chKey, "mode": mode, "players": players, "attempts": attempts,
			"wins": wins, "winners": winners, "rejected": rejected, "completionRate": math.Round(rate*100) / 100, "avgBestStars": math.Round(avgBest*100) / 100,
			"avgStars": math.Round(avgStars*100) / 100, "dropOff": dropOff})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"days": days, "items": items, "totals": totals})
}
