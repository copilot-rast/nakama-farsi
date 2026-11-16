package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/leaderboardx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func iso(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

type texts struct {
	langs *localized.Set
	lang  string
}

func (s *service) texts(ctx context.Context, userID, lang string) texts {
	langs, l, _ := localized.ForUser(ctx, s.db, userID, lang)
	return texts{langs: langs, lang: l}
}

func (t texts) of(m map[string]string, fallback string) string {
	if t.langs != nil {
		if v := t.langs.Resolve(localized.Text(m), t.lang); v != "" {
			return v
		}
	}
	for _, v := range m {
		if v != "" {
			return v
		}
	}
	return fallback
}

func visible(b *board, cfg settings.All, now time.Time) bool {
	switch b.status(now) {
	case "inactive":
		return false
	case "upcoming":
		return cfg.Display.ShowUpcoming
	case "ended":
		end := b.EndsAt
		if b.FinishedAt != nil && (end == nil || b.FinishedAt.Before(*end)) {
			end = b.FinishedAt
		}
		return end != nil && cfg.Display.ShowEndedDays > 0 && now.Before(end.Add(time.Duration(cfg.Display.ShowEndedDays)*24*time.Hour))
	}
	return true
}

func (s *service) boardInfo(b *board, tx texts, now time.Time) map[string]any {
	start, end := periodBounds(b.ResetSchedule, now)
	if b.StartsAt != nil && (start == nil || start.Before(*b.StartsAt)) {
		start = b.StartsAt
	}
	if b.EndsAt != nil && (end == nil || b.EndsAt.Before(*end)) {
		end = b.EndsAt
	}
	tiers := make([]map[string]any, 0, len(b.Tiers))
	for _, t := range b.Tiers {
		tiers = append(tiers, map[string]any{"key": t.Key, "name": tx.of(t.Name, t.Key), "iconUrl": t.IconURL, "threshold": t.Threshold})
	}
	return map[string]any{"key": b.Key, "name": tx.of(b.Name, b.Key), "description": tx.of(b.Description, ""), "iconUrl": b.IconURL, "sort": b.Sort,
		"status": b.status(now), "sortOrder": b.SortOrder, "operator": b.Operator, "period": b.Period, "resetSchedule": b.ResetSchedule,
		"periodStart": iso(start), "periodEnd": iso(end), "startsAt": iso(b.StartsAt), "endsAt": iso(b.EndsAt),
		"clientSubmit": b.Limits.ClientSubmit, "tierBasis": b.TierBasis, "tiers": tiers, "rewards": rewardPreview(b, nil), "rewardMinScore": b.RewardMinScore,
		"payMode": b.payMode(s.settings(context.Background())), "metadata": b.Metadata}
}

func (s *service) rpcList(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	now := s.now()
	tx := s.texts(ctx, userID, in.Lang)
	list, err := s.boards(ctx, "active")
	if err != nil {
		return nil, err
	}
	pending := map[string]int{}
	rows, err := s.db.QueryContext(ctx, `SELECT p.board_key, count(*) FROM leaderboardx_result r JOIN leaderboardx_period p ON p.id=r.period_id WHERE r.user_id=$1 AND r.status='claimable' GROUP BY 1`, userID)
	if err == nil {
		for rows.Next() {
			var k string
			var n int
			if rows.Scan(&k, &n) == nil {
				pending[k] = n
			}
		}
		rows.Close()
	}
	out := []map[string]any{}
	for _, b := range list {
		if !visible(b, cfg, now) {
			continue
		}
		info := s.boardInfo(b, tx, now)
		info["me"] = nil
		if b.status(now) == "active" {
			if _, own, _, err := s.core.List(ctx, b.Key, []string{userID}, 1, ""); err == nil && len(own) > 0 {
				info["me"] = s.meRow(ctx, b, own[0])
			}
		}
		info["claimable"] = pending[b.Key]
		out = append(out, info)
	}
	return map[string]any{"serverTime": now.UTC().Format(time.RFC3339), "boards": out}, nil
}

func (s *service) meRow(ctx context.Context, b *board, r *api.LeaderboardRecord) map[string]any {
	total := 0
	if b.TierBasis == "percent" {
		total = countPlayers(ctx, s.db, b.Key, expiry(b.ResetSchedule, s.now()))
	}
	idx := tierIndex(b, int(r.Rank), r.Score, total)
	me := map[string]any{"rank": r.Rank, "score": r.Score, "subscore": r.Subscore, "tier": nil, "nextTier": nextTier(b, idx, int(r.Rank), r.Score, total)}
	if idx >= 0 {
		me["tier"] = b.Tiers[idx].Key
	}
	return me
}

func (s *service) rows(ctx context.Context, b *board, recs []*api.LeaderboardRecord, userID string, total int, cfg settings.All) []map[string]any {
	ids := make([]string, 0, len(recs))
	for _, r := range recs {
		ids = append(ids, r.OwnerId)
	}
	var people map[string]person
	if cfg.Display.ShowDisplayName {
		people = s.people(ctx, ids)
	}
	out := make([]map[string]any, 0, len(recs))
	for _, r := range recs {
		row := map[string]any{"rank": r.Rank, "userId": r.OwnerId, "username": "", "score": r.Score, "subscore": r.Subscore, "me": r.OwnerId == userID, "tier": nil}
		if r.Username != nil {
			row["username"] = r.Username.Value
		}
		if p, ok := people[r.OwnerId]; ok {
			row["displayName"], row["avatarUrl"] = p.DisplayName, p.AvatarURL
			if p.Username != "" {
				row["username"] = p.Username
			}
		}
		if idx := tierIndex(b, int(r.Rank), r.Score, total); idx >= 0 {
			row["tier"] = b.Tiers[idx].Key
		}
		out = append(out, row)
	}
	return out
}

func (s *service) rpcBoard(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key    string `json:"key"`
		Lang   string `json:"lang"`
		Limit  int    `json:"limit"`
		Around *int   `json:"around"`
		Cursor string `json:"cursor"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	now := s.now()
	b, err := s.visibleBoard(ctx, in.Key, cfg, now)
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = cfg.Display.TopLimit
	}
	if limit > cfg.Display.MaxLimit {
		limit = cfg.Display.MaxLimit
	}
	around := cfg.Display.AroundLimit
	if in.Around != nil && *in.Around >= 0 && *in.Around <= cfg.Display.MaxLimit {
		around = *in.Around
	}
	info := s.boardInfo(b, s.texts(ctx, userID, in.Lang), now)
	exp := expiry(b.ResetSchedule, now)
	total := countPlayers(ctx, s.db, b.Key, exp)
	info["players"] = total
	info["top"], info["around"], info["me"], info["nextCursor"] = []any{}, []any{}, nil, ""
	if b.status(now) == "upcoming" {
		return info, nil
	}
	if b.status(now) == "ended" {

		return s.endedView(ctx, b, info, userID, limit, cfg)
	}
	recs, own, next, err := s.core.List(ctx, b.Key, []string{userID}, limit, strings.TrimSpace(in.Cursor))
	if err != nil {
		return nil, apierr.BadRequest("LEADERBOARD_INVALID_CURSOR").WithField("cursor")
	}
	info["top"], info["nextCursor"] = s.rows(ctx, b, recs, userID, total, cfg), next
	if len(own) > 0 {
		info["me"] = s.meRow(ctx, b, own[0])
		if around > 0 {
			if hay, err := s.core.Haystack(ctx, b.Key, userID, around); err == nil {
				info["around"] = s.rows(ctx, b, hay, userID, total, cfg)
			}
		}
	}
	return info, nil
}

func (s *service) visibleBoard(ctx context.Context, key string, cfg settings.All, now time.Time) (*board, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, apierr.BadRequest("LEADERBOARD_KEY_REQUIRED").WithField("key")
	}
	b, err := s.boardByKey(ctx, s.db, key)
	if errors.Is(err, errNoBoard) || (err == nil && !visible(b, cfg, now)) {
		return nil, apierr.NotFound("LEADERBOARD_NOT_FOUND").WithField("key")
	}
	return b, err
}

func (s *service) endedView(ctx context.Context, b *board, info map[string]any, userID string, limit int, cfg settings.All) (any, error) {
	var periodID string
	var players int
	err := s.db.QueryRowContext(ctx, `SELECT id::text, players FROM leaderboardx_period WHERE board_key=$1 ORDER BY period_end DESC LIMIT 1`, b.Key).Scan(&periodID, &players)
	if errors.Is(err, sql.ErrNoRows) {
		return info, nil
	}
	if err != nil {
		return nil, err
	}
	info["players"], info["periodId"] = players, periodID
	top, err := s.periodRows(ctx, periodID, `r.rank <= $2`, []any{limit}, userID)
	if err != nil {
		return nil, err
	}
	info["top"] = top
	mine, err := s.periodRows(ctx, periodID, `r.user_id::text = $2`, []any{userID}, userID)
	if err == nil && len(mine) > 0 {
		info["me"] = map[string]any{"rank": mine[0]["rank"], "score": mine[0]["score"], "subscore": mine[0]["subscore"], "tier": mine[0]["tier"], "nextTier": nil}
	}
	return info, nil
}

func (s *service) periodRows(ctx context.Context, periodID, where string, args []any, userID string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT r.user_id::text, r.rank, r.tier, r.score, r.subscore, COALESCE(NULLIF(u.username,''), r.username), COALESCE(u.display_name,''), COALESCE(u.avatar_url,''), r.status, r.plan
		FROM leaderboardx_result r LEFT JOIN users u ON u.id=r.user_id WHERE r.period_id=$1 AND `+where+` ORDER BY r.rank LIMIT 500`, append([]any{periodID}, args...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var uid, tier, name, display, avatar, status string
		var rank int
		var score, sub int64
		var rawPlan []byte
		if err := rows.Scan(&uid, &rank, &tier, &score, &sub, &name, &display, &avatar, &status, &rawPlan); err != nil {
			return nil, err
		}
		var pl plan
		unmarshal(rawPlan, &pl)
		prizes := []map[string]any{}
		for _, p := range pl.Prizes {
			prizes = append(prizes, map[string]any{"prizeKey": p.PrizeKey, "prizeGroupKey": p.PrizeGroupKey})
		}
		var t any
		if tier != "" {
			t = tier
		}
		out = append(out, map[string]any{"rank": rank, "userId": uid, "username": name, "displayName": display, "avatarUrl": avatar, "score": score, "subscore": sub, "tier": t,
			"me": uid == userID, "rewards": map[string]any{"currencies": currencyList(pl.Currencies), "prizes": prizes}})
	}
	return out, rows.Err()
}

func (s *service) rpcHistory(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key    string `json:"key"`
		Lang   string `json:"lang"`
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
		Status string `json:"status"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	limit := in.Limit
	if limit <= 0 || limit > cfg.History.PlayerLimit {
		limit = cfg.History.PlayerLimit
	}
	if in.Offset < 0 {
		in.Offset = 0
	}
	tx := s.texts(ctx, userID, in.Lang)
	where, args := `r.user_id=$1`, []any{userID}
	if k := strings.TrimSpace(in.Key); k != "" {
		args = append(args, k)
		where += ` AND p.board_key=$2`
	}
	switch in.Status {
	case "":
	case "claimable", "pending", "paid", "expired", "failed", "none":
		args = append(args, in.Status)
		where += ` AND r.status=$` + itoa(len(args))
	default:
		return nil, apierr.BadRequest("LEADERBOARD_INVALID_STATUS").WithField("status")
	}
	args = append(args, limit+1, in.Offset)
	rows, err := s.db.QueryContext(ctx, `SELECT p.id::text, p.board_key, p.board_name::text, p.period_start, p.period_end, p.players, r.rank, r.tier_rank, r.tier, r.score, r.subscore, r.username, r.status, r.plan, r.paid, r.claim_expires_at, r.paid_at
		FROM leaderboardx_result r JOIN leaderboardx_period p ON p.id=r.period_id WHERE `+where+` ORDER BY p.period_end DESC, p.board_key LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var pid, key, rawName, tier, uname, status string
		var start *time.Time
		var end time.Time
		var players, rank, tierRank int
		var score, sub int64
		var rawPlan, rawPaid []byte
		var claimUntil, paidAt *time.Time
		if err := rows.Scan(&pid, &key, &rawName, &start, &end, &players, &rank, &tierRank, &tier, &score, &sub, &uname, &status, &rawPlan, &rawPaid, &claimUntil, &paidAt); err != nil {
			return nil, err
		}
		row := resultRow(pid, userID, rank, tierRank, tier, score, sub, uname, status, rawPlan, rawPaid, claimUntil, paidAt, "", 0)
		delete(row, "lastError")
		delete(row, "payTries")
		name := map[string]string{}
		unmarshal([]byte(rawName), &name)
		row["boardKey"], row["boardName"], row["periodStart"], row["periodEnd"], row["players"] = key, tx.of(name, key), iso(start), iso(&end), players
		items = append(items, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	var claimable int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM leaderboardx_result WHERE user_id=$1 AND status='claimable'`, userID).Scan(&claimable)
	return map[string]any{"items": items, "more": more, "claimable": claimable}, nil
}

func (s *service) rpcClaim(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		PeriodID string `json:"periodId"`
		All      bool   `json:"all"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	pid := strings.TrimSpace(in.PeriodID)
	if pid == "" && !in.All {
		return nil, apierr.BadRequest("LEADERBOARD_PERIOD_REQUIRED").WithField("periodId")
	}
	ids := []string{}
	if pid != "" {
		if !uuidPattern.MatchString(pid) {
			return nil, apierr.NotFound("LEADERBOARD_REWARD_NOT_FOUND").WithField("periodId")
		}
		ids = append(ids, pid)
	} else {
		rows, err := s.db.QueryContext(ctx, `SELECT period_id::text FROM leaderboardx_result WHERE user_id=$1 AND status='claimable' ORDER BY period_id LIMIT 50`, userID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			if rows.Scan(&id) == nil {
				ids = append(ids, id)
			}
		}
		rows.Close()
	}
	claimed := []map[string]any{}
	for _, id := range ids {
		out, err := s.payResult(ctx, id, userID, true)
		if err != nil {
			if pid != "" {
				return nil, err
			}
			continue
		}
		claimed = append(claimed, out)
	}
	return map[string]any{"claimed": claimed}, nil
}

func (s *service) rpcWinners(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key      string `json:"key"`
		PeriodID string `json:"periodId"`
		Lang     string `json:"lang"`
		Limit    int    `json:"limit"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	b, err := s.visibleBoard(ctx, in.Key, cfg, s.now())
	if err != nil {
		return nil, err
	}
	limit := in.Limit
	if limit <= 0 || limit > cfg.History.WinnersLimit {
		limit = cfg.History.WinnersLimit
	}
	q, args := `SELECT id::text, period_start, period_end, players, reason FROM leaderboardx_period WHERE board_key=$1`, []any{b.Key}
	if p := strings.TrimSpace(in.PeriodID); p != "" {
		if !uuidPattern.MatchString(p) {
			return nil, apierr.NotFound("LEADERBOARD_PERIOD_NOT_FOUND").WithField("periodId")
		}
		q, args = q+` AND id::text=$2`, append(args, p)
	}
	var pid, reason string
	var start *time.Time
	var end time.Time
	var players int
	err = s.db.QueryRowContext(ctx, q+` ORDER BY period_end DESC LIMIT 1`, args...).Scan(&pid, &start, &end, &players, &reason)
	if errors.Is(err, sql.ErrNoRows) {
		if in.PeriodID != "" {
			return nil, apierr.NotFound("LEADERBOARD_PERIOD_NOT_FOUND").WithField("periodId")
		}
		return map[string]any{"key": b.Key, "period": nil, "winners": []any{}, "periods": []any{}}, nil
	}
	if err != nil {
		return nil, err
	}
	winners, err := s.periodRows(ctx, pid, `r.rank <= $2`, []any{limit}, userID)
	if err != nil {
		return nil, err
	}
	periods := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, period_start, period_end, players FROM leaderboardx_period WHERE board_key=$1 ORDER BY period_end DESC LIMIT 20`, b.Key)
	if err == nil {
		for rows.Next() {
			var id string
			var st *time.Time
			var en time.Time
			var n int
			if rows.Scan(&id, &st, &en, &n) == nil {
				periods = append(periods, map[string]any{"periodId": id, "periodStart": iso(st), "periodEnd": iso(&en), "players": n})
			}
		}
		rows.Close()
	}
	return map[string]any{"key": b.Key, "period": map[string]any{"periodId": pid, "periodStart": iso(start), "periodEnd": iso(&end), "players": players, "reason": reason},
		"winners": winners, "periods": periods}, nil
}
