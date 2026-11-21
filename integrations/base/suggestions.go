package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"math/rand/v2"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

const (
	coplayLease    = "base.coplay.record"
	coplayLeaseTTL = 2 * time.Minute
	coplayTick     = 20 * time.Second
	coplayBatch    = 200
	coplayMaxSeats = 16
)

type suggestSettings struct {
	Enabled             bool   `json:"enabled"`
	Visible             int    `json:"visible"`
	Reserve             int    `json:"reserve"`
	LookbackDays        int    `json:"lookbackDays"`
	CoplayQuota         int    `json:"coplayQuota"`
	MutualQuota         int    `json:"mutualQuota"`
	ActivityQuota       int    `json:"activityQuota"`
	LevelQuota          int    `json:"levelQuota"`
	PresenceDays        int    `json:"presenceDays"`
	DecayAfter          int    `json:"decayAfter"`
	CandidateMaxFriends int    `json:"candidateMaxFriends"`
	LevelTable          string `json:"levelTable"`
	LevelDelta          int    `json:"levelDelta"`
	ScanMax             int    `json:"scanMax"`
	LivePatch           bool   `json:"livePatch"`
}

func defaultSuggestSettings() suggestSettings {
	return suggestSettings{Enabled: true, Visible: 50, Reserve: 20, LookbackDays: 30, CoplayQuota: 20, MutualQuota: 15, ActivityQuota: 10, LevelQuota: 5,
		PresenceDays: 7, DecayAfter: 4, CandidateMaxFriends: 50, LevelDelta: 2, ScanMax: 300, LivePatch: true}
}

var suggestStore = settingsrow.New("base_friend_suggest_settings", defaultSuggestSettings)

func (v suggestSettings) clean() (suggestSettings, error) {
	check := func(field string, n, lo, hi int) error {
		if n < lo || n > hi {
			return badRequest("INVALID_RULE", field, "between "+itoa(lo)+" and "+itoa(hi))
		}
		return nil
	}
	for _, c := range []struct {
		f         string
		n, lo, hi int
	}{
		{"visible", v.Visible, 1, 100}, {"reserve", v.Reserve, 0, 100}, {"lookbackDays", v.LookbackDays, 1, 365},
		{"coplayQuota", v.CoplayQuota, 0, 100}, {"mutualQuota", v.MutualQuota, 0, 100}, {"activityQuota", v.ActivityQuota, 0, 100}, {"levelQuota", v.LevelQuota, 0, 100},
		{"presenceDays", v.PresenceDays, 1, 60}, {"decayAfter", v.DecayAfter, 0, 365}, {"candidateMaxFriends", v.CandidateMaxFriends, 0, 10000},
		{"levelDelta", v.LevelDelta, 0, 1000}, {"scanMax", v.ScanMax, 10, 2000},
	} {
		if err := check(c.f, c.n, c.lo, c.hi); err != nil {
			return v, err
		}
	}
	return v, nil
}

type suggestRow struct {
	Candidate  string
	Source     string
	Score      float64
	Mutual     int
	LastPlayed sql.NullTime
	Shows      int
}

func (s *service) suggestCycle(ctx context.Context) (string, time.Time) {
	zone := projectZone(ctx, s.db)
	cycle, next := dayBounds(s.now(), zone)
	return cycle, next.AddDate(0, 0, -1)
}

func (s *service) suggestExcluded(ctx context.Context, userID, cycle string) (map[string]bool, error) {
	out := map[string]bool{userID: true}
	rows, err := s.db.QueryContext(ctx, `SELECT destination_id::text FROM user_edge WHERE source_id=$1::uuid
		UNION SELECT source_id::text FROM user_edge WHERE destination_id=$1::uuid
		UNION SELECT candidate_id::text FROM base_friend_suggest_skip WHERE user_id=$1::uuid AND until_cycle >= $2`, userID, cycle)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func activityBucket(days, window int) int {
	if days <= 0 || window <= 0 {
		return 0
	}
	f := float64(days) / float64(window)
	switch {
	case f >= 6.0/7.0:
		return 3
	case f >= 3.0/7.0:
		return 2
	default:
		return 1
	}
}

func suggestDecay(shows, after int) float64 {
	if shows <= after {
		return 1
	}
	return math.Pow(0.5, float64(shows-after))
}

func scanPool(rows *sql.Rows, fn func(*sql.Rows) (suggestRow, error)) ([]suggestRow, error) {
	defer rows.Close()
	out := []suggestRow{}
	for rows.Next() {
		r, err := fn(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *service) suggestPools(ctx context.Context, v suggestSettings, userID string, now time.Time) (map[string][]suggestRow, error) {
	pools := map[string][]suggestRow{}
	rows, err := s.db.QueryContext(ctx, `SELECT other_id::text, last_at FROM base_coplay WHERE user_id=$1 AND last_at > $2 ORDER BY last_at DESC, other_id LIMIT $3`,
		userID, now.AddDate(0, 0, -v.LookbackDays), v.ScanMax)
	if err != nil {
		return nil, err
	}
	if pools["coplay"], err = scanPool(rows, func(r *sql.Rows) (suggestRow, error) {
		var x suggestRow
		var at time.Time
		err := r.Scan(&x.Candidate, &at)
		x.Source, x.Score, x.LastPlayed = "coplay", float64(at.UnixMilli()), sql.NullTime{Time: at, Valid: true}
		return x, err
	}); err != nil {
		return nil, err
	}
	rows, err = s.db.QueryContext(ctx, `SELECT e2.destination_id::text, count(*) FROM user_edge e1 JOIN user_edge e2 ON e2.source_id=e1.destination_id AND e2.state=0
		WHERE e1.source_id=$1::uuid AND e1.state=0 AND e2.destination_id<>$1::uuid GROUP BY 1 ORDER BY 2 DESC, 1 LIMIT $2`, userID, v.ScanMax)
	if err != nil {
		return nil, err
	}
	if pools["mutual"], err = scanPool(rows, func(r *sql.Rows) (suggestRow, error) {
		var x suggestRow
		err := r.Scan(&x.Candidate, &x.Mutual)
		x.Source, x.Score = "mutual", float64(x.Mutual)
		return x, err
	}); err != nil {
		return nil, err
	}
	zone := projectZone(ctx, s.db)
	today, _ := dayBounds(now, zone)
	start, _ := time.ParseInLocation("2006-01-02", today, zone)
	since := start.AddDate(0, 0, -(v.PresenceDays - 1)).Format("2006-01-02")
	var mine int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM base_presence_day WHERE user_id=$1 AND day >= $2::date`, userID, since).Scan(&mine); err != nil {
		return nil, err
	}
	if vb := activityBucket(mine, v.PresenceDays); vb > 0 {
		rows, err = s.db.QueryContext(ctx, `SELECT user_id::text, count(*), max(day) FROM base_presence_day WHERE day >= $1::date AND user_id<>$2 GROUP BY user_id ORDER BY max(day) DESC, count(*) DESC, user_id LIMIT $3`,
			since, userID, v.ScanMax)
		if err != nil {
			return nil, err
		}
		list, err := scanPool(rows, func(r *sql.Rows) (suggestRow, error) {
			var x suggestRow
			var n int
			var last time.Time
			err := r.Scan(&x.Candidate, &n, &last)
			x.Source = "activity"
			d := activityBucket(n, v.PresenceDays) - vb
			if d < 0 {
				d = -d
			}
			x.Score = float64(2 - d)
			if d > 1 {
				x.Score = -1
			}
			return x, err
		})
		if err != nil {
			return nil, err
		}
		keep := list[:0]
		for _, x := range list {
			if x.Score >= 0 {
				keep = append(keep, x)
			}
		}
		pools["activity"] = keep
	}
	if v.LevelTable != "" {
		var level sql.NullInt64
		var tableID sql.NullString
		err := s.db.QueryRowContext(ctx, `SELECT t.id::text, (SELECT ul.level FROM economy_user_level ul WHERE ul.table_id=t.id AND ul.user_id=$2) FROM economy_level_table t WHERE t.key=$1`,
			v.LevelTable, userID).Scan(&tableID, &level)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if tableID.Valid {
			mineLevel := int64(0)
			if level.Valid {
				mineLevel = level.Int64
			}
			rows, err = s.db.QueryContext(ctx, `SELECT user_id::text, level FROM economy_user_level WHERE table_id=$1 AND user_id<>$2 AND abs(level - $3) <= $4
				ORDER BY abs(level - $3), updated_at DESC, user_id LIMIT $5`, tableID.String, userID, mineLevel, v.LevelDelta, v.ScanMax)
			if err != nil {
				return nil, err
			}
			if pools["level"], err = scanPool(rows, func(r *sql.Rows) (suggestRow, error) {
				var x suggestRow
				var lv int64
				err := r.Scan(&x.Candidate, &lv)
				d := lv - mineLevel
				if d < 0 {
					d = -d
				}
				x.Source, x.Score = "level", float64(int64(v.LevelDelta)+1-d)
				return x, err
			}); err != nil {
				return nil, err
			}
		}
	}
	return pools, nil
}

func (s *service) eligibleCandidates(ctx context.Context, v suggestSettings, ids []string) (map[string]bool, error) {
	out := map[string]bool{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT u.id::text, (SELECT count(*) FROM user_edge e WHERE e.source_id=u.id AND e.state=0) FROM users u
		WHERE u.id = ANY($1::uuid[]) AND u.disable_time <= '1970-01-01 00:00:00+00'`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var friends int
		if err := rows.Scan(&id, &friends); err != nil {
			return nil, err
		}
		if v.CandidateMaxFriends == 0 || friends < v.CandidateMaxFriends {
			out[id] = true
		}
	}
	return out, rows.Err()
}

func rankSuggestions(v suggestSettings, pools map[string][]suggestRow, ok func(string) bool, shows map[string]int, fresh map[string]bool, shuffle func([]suggestRow)) []suggestRow {
	order := []struct {
		name  string
		quota int
	}{{"coplay", v.CoplayQuota}, {"mutual", v.MutualQuota}, {"activity", v.ActivityQuota}, {"level", v.LevelQuota}}
	ranked := map[string][]suggestRow{}
	for _, o := range order {
		list := []suggestRow{}
		for _, r := range pools[o.name] {
			if !ok(r.Candidate) {
				continue
			}
			if !fresh[r.Candidate] {
				r.Shows = shows[r.Candidate]
			}
			r.Score *= suggestDecay(r.Shows, v.DecayAfter)
			list = append(list, r)
		}
		sort.SliceStable(list, func(i, j int) bool { return list[i].Score > list[j].Score })
		ranked[o.name] = list
	}
	total := v.Visible + v.Reserve
	taken := map[string]bool{}
	out := []suggestRow{}
	used := map[string]int{}
	for _, o := range order {
		for _, r := range ranked[o.name] {
			if len(out) >= total || used[o.name] >= o.quota {
				break
			}
			if taken[r.Candidate] {
				continue
			}
			taken[r.Candidate] = true
			used[o.name]++
			out = append(out, r)
		}
	}
	for _, o := range order {
		for _, r := range ranked[o.name] {
			if len(out) >= total {
				break
			}
			if !taken[r.Candidate] {
				taken[r.Candidate] = true
				out = append(out, r)
			}
		}
	}
	keep := min(v.CoplayQuota, len(out))
	vis := min(v.Visible, len(out))
	if vis > keep && shuffle != nil {
		shuffle(out[keep:vis])
	}
	return out
}

func (s *service) generateSuggestions(ctx context.Context, userID, by string) error {
	v, err := suggestStore.Load(ctx, s.db)
	if err != nil {
		return err
	}
	cycle, cycleStart := s.suggestCycle(ctx)
	now := s.now().UTC()
	excluded, err := s.suggestExcluded(ctx, userID, cycle)
	if err != nil {
		return err
	}
	pools, err := s.suggestPools(ctx, v, userID, now)
	if err != nil {
		return err
	}
	ids := []string{}
	seen := map[string]bool{}
	for _, list := range pools {
		for _, r := range list {
			if !excluded[r.Candidate] && !seen[r.Candidate] {
				seen[r.Candidate] = true
				ids = append(ids, r.Candidate)
			}
		}
	}
	eligible, err := s.eligibleCandidates(ctx, v, ids)
	if err != nil {
		return err
	}
	shows := map[string]int{}
	prev, err := s.db.QueryContext(ctx, `SELECT candidate_id::text, shows FROM base_friend_suggest WHERE user_id=$1`, userID)
	if err != nil {
		return err
	}
	for prev.Next() {
		var id string
		var n int
		if err := prev.Scan(&id, &n); err != nil {
			prev.Close()
			return err
		}
		shows[id] = n
	}
	prev.Close()
	fresh := map[string]bool{}
	for _, r := range pools["coplay"] {
		if r.LastPlayed.Valid && !r.LastPlayed.Time.Before(cycleStart) {
			fresh[r.Candidate] = true
		}
	}
	list := rankSuggestions(v, pools, func(id string) bool { return !excluded[id] && eligible[id] }, shows, fresh, func(rs []suggestRow) {
		rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })
	})
	return s.saveSuggestions(ctx, userID, cycle, by, list, true)
}

func (s *service) saveSuggestions(ctx context.Context, userID, cycle, by string, list []suggestRow, stamp bool) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('base.suggest:' || $1::text))`, userID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM base_friend_suggest WHERE user_id=$1`, userID); err != nil {
			return err
		}
		for i, r := range list {
			var last any
			if r.LastPlayed.Valid {
				last = r.LastPlayed.Time
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO base_friend_suggest (user_id, candidate_id, slot, source, score, mutual, last_played_at, shows) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
				ON CONFLICT (user_id, candidate_id) DO NOTHING`, userID, r.Candidate, i+1, r.Source, r.Score, r.Mutual, last, r.Shows); err != nil {
				return err
			}
		}
		if !stamp {
			return nil
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO base_friend_suggest_list (user_id, cycle, generated_at, generated_by) VALUES ($1,$2,now(),$3)
			ON CONFLICT (user_id) DO UPDATE SET cycle=EXCLUDED.cycle, generated_at=now(), generated_by=EXCLUDED.generated_by`, userID, cycle, by)
		return err
	})
}

func (s *service) loadSuggestRows(ctx context.Context, userID string) ([]suggestRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT candidate_id::text, source, score, mutual, last_played_at, shows FROM base_friend_suggest WHERE user_id=$1 ORDER BY slot`, userID)
	if err != nil {
		return nil, err
	}
	return scanPool(rows, func(r *sql.Rows) (suggestRow, error) {
		var x suggestRow
		err := r.Scan(&x.Candidate, &x.Source, &x.Score, &x.Mutual, &x.LastPlayed, &x.Shows)
		return x, err
	})
}

func (s *service) listCycle(ctx context.Context, userID string) (string, string, error) {
	var cycle, imp string
	err := s.db.QueryRowContext(ctx, `SELECT cycle, impression_cycle FROM base_friend_suggest_list WHERE user_id=$1`, userID).Scan(&cycle, &imp)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	return cycle, imp, err
}

type suggestUser struct {
	UserID       string  `json:"userId"`
	Username     string  `json:"username"`
	DisplayName  string  `json:"displayName"`
	AvatarURL    string  `json:"avatarUrl"`
	Code         string  `json:"code"`
	Online       bool    `json:"online"`
	Source       string  `json:"source"`
	Mutual       int     `json:"mutual"`
	LastPlayedAt *string `json:"lastPlayedAt,omitempty"`
	Score        float64 `json:"score,omitempty"`
	Shows        int     `json:"shows,omitempty"`
}

func (s *service) suggestUsers(ctx context.Context, list []suggestRow, admin bool) ([]suggestUser, error) {
	ids := make([]string, 0, len(list))
	for _, r := range list {
		ids = append(ids, r.Candidate)
	}
	info := map[string]suggestUser{}
	if len(ids) > 0 {
		rows, err := s.db.QueryContext(ctx, `SELECT u.id::text, u.username, COALESCE(u.display_name,''), COALESCE(u.avatar_url,''), COALESCE(c.code,'') FROM users u
			LEFT JOIN base_player_code c ON c.user_id=u.id WHERE u.id = ANY($1::uuid[])`, ids)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var x suggestUser
			if err := rows.Scan(&x.UserID, &x.Username, &x.DisplayName, &x.AvatarURL, &x.Code); err != nil {
				rows.Close()
				return nil, err
			}
			info[x.UserID] = x
		}
		rows.Close()
	}
	out := []suggestUser{}
	for _, r := range list {
		x, ok := info[r.Candidate]
		if !ok {
			continue
		}
		x.Source, x.Mutual = r.Source, r.Mutual
		if r.LastPlayed.Valid {
			at := r.LastPlayed.Time.UTC().Format(time.RFC3339)
			x.LastPlayedAt = &at
		}
		x.Online = s.isOnline(x.UserID)
		if admin {
			x.Score, x.Shows = r.Score, r.Shows
		}
		out = append(out, x)
	}
	return out, nil
}

func (s *service) rpcFriendSuggestions(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	v, err := suggestStore.Load(ctx, s.db)
	if err != nil {
		return nil, err
	}
	cycle, _ := s.suggestCycle(ctx)
	if !v.Enabled {
		return map[string]any{"enabled": false, "cycle": cycle, "suggestions": []suggestUser{}}, nil
	}
	have, imp, err := s.listCycle(ctx, userID)
	if err != nil {
		return nil, err
	}
	if have != cycle {
		if err := s.generateSuggestions(ctx, userID, ""); err != nil {
			return nil, err
		}
		imp = ""
	}
	if imp != cycle {
		res, err := s.db.ExecContext(ctx, `UPDATE base_friend_suggest_list SET impression_cycle=$2 WHERE user_id=$1 AND impression_cycle<>$2`, userID, cycle)
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			if _, err := s.db.ExecContext(ctx, `UPDATE base_friend_suggest SET shows=shows+1 WHERE user_id=$1 AND slot <= $2`, userID, v.Visible); err != nil {
				return nil, err
			}
		}
	}
	all, err := s.loadSuggestRows(ctx, userID)
	if err != nil {
		return nil, err
	}
	excluded, err := s.suggestExcluded(ctx, userID, cycle)
	if err != nil {
		return nil, err
	}
	rows := []suggestRow{}
	for _, r := range all {
		if !excluded[r.Candidate] {
			rows = append(rows, r)
		}
	}
	if len(rows) > v.Visible {
		rows = rows[:v.Visible]
	}
	users, err := s.suggestUsers(ctx, rows, false)
	if err != nil {
		return nil, err
	}
	_, next := dayBounds(s.now(), projectZone(ctx, s.db))
	return map[string]any{"enabled": true, "cycle": cycle, "refreshesAt": next.UTC().Format(time.RFC3339), "suggestions": users}, nil
}

func (s *service) dropSuggestionPairs(ctx context.Context, userID string, others []string) {
	if len(others) == 0 {
		return
	}
	owners := append([]string{userID}, others...)
	if _, err := s.db.ExecContext(ctx, `DELETE FROM base_friend_suggest WHERE (user_id=$1 AND candidate_id = ANY($2::uuid[])) OR (candidate_id=$1 AND user_id = ANY($2::uuid[]))`, userID, others); err != nil {
		s.logger.Warn("Base friend suggestions not cleared: %v", err)
		return
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE base_friend_suggest f SET slot=r.n FROM (SELECT user_id, candidate_id, row_number() OVER (PARTITION BY user_id ORDER BY slot) AS n FROM base_friend_suggest WHERE user_id = ANY($1::uuid[])) r
		WHERE f.user_id=r.user_id AND f.candidate_id=r.candidate_id AND f.slot<>r.n`, owners); err != nil {
		s.logger.Warn("Base friend suggestions not compacted: %v", err)
	}
}

type coplayMatch struct {
	ID      string
	GameKey string
	EndedAt time.Time
	Players []string
}

func (s *service) coplayLoop(ctx context.Context) {
	ticker := time.NewTicker(coplayTick)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			_ = clusterkit.ReleaseLease(context.Background(), s.db, coplayLease, holder)
			return
		case <-ticker.C:
		}
		ok, err := clusterkit.AcquireLease(ctx, s.db, coplayLease, holder, coplayLeaseTTL)
		if err != nil || !ok {
			continue
		}
		if _, err := s.recordCoplay(ctx, coplayBatch); err != nil && ctx.Err() == nil {
			s.logger.Warn("Base co-play recording failed: %v", err)
		}
	}
}

func (s *service) recordCoplay(ctx context.Context, limit int) (int, error) {
	var at time.Time
	var after string
	if err := s.db.QueryRowContext(ctx, `SELECT ended_at, match_id FROM base_coplay_cursor WHERE id=1`).Scan(&at, &after); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
		at = s.now().UTC().AddDate(0, 0, -30)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT m.id, m.game_key, m.ended_at, COALESCE(string_agg(e.user_id::text, ',' ORDER BY e.created_at) FILTER (WHERE e.status IN ('pending','paid','free','left') AND u.id IS NOT NULL), '')
		FROM economy_game_match m LEFT JOIN economy_game_entry e ON e.match_id=m.id LEFT JOIN users u ON u.id=e.user_id
		WHERE m.status='finished' AND m.ended_at IS NOT NULL AND (m.ended_at > $1 OR (m.ended_at = $1 AND m.id > $2))
		GROUP BY m.id, m.game_key, m.ended_at ORDER BY m.ended_at, m.id LIMIT $3`, at, after, limit)
	if err != nil {
		return 0, err
	}
	matches := []coplayMatch{}
	for rows.Next() {
		var m coplayMatch
		var joined string
		if err := rows.Scan(&m.ID, &m.GameKey, &m.EndedAt, &joined); err != nil {
			rows.Close()
			return 0, err
		}
		players := []string{}
		if joined != "" {
			players = strings.Split(joined, ",")
		}
		if len(players) > coplayMaxSeats {
			players = players[:coplayMaxSeats]
		}
		m.Players = players
		matches = append(matches, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	for _, m := range matches {
		if err := s.saveCoplay(ctx, m); err != nil {
			return 0, err
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE base_coplay_cursor SET ended_at=$1, match_id=$2 WHERE id=1`, m.EndedAt, m.ID); err != nil {
			return 0, err
		}
	}
	return len(matches), nil
}

func (s *service) saveCoplay(ctx context.Context, m coplayMatch) error {
	if len(m.Players) < 2 {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO base_coplay (user_id, other_id, last_match_id, last_game_key, last_at, matches)
		SELECT a, b, $2, $3, $4, 1 FROM unnest($1::uuid[]) a CROSS JOIN unnest($1::uuid[]) b WHERE a <> b
		ON CONFLICT (user_id, other_id) DO UPDATE SET last_match_id=EXCLUDED.last_match_id, last_game_key=EXCLUDED.last_game_key, last_at=GREATEST(base_coplay.last_at, EXCLUDED.last_at),
			matches=base_coplay.matches + CASE WHEN base_coplay.last_match_id = EXCLUDED.last_match_id THEN 0 ELSE 1 END
		WHERE base_coplay.last_at <= EXCLUDED.last_at`, m.Players, m.ID, m.GameKey, m.EndedAt); err != nil {
		return err
	}
	v, err := suggestStore.Load(ctx, s.db)
	if err != nil || !v.Enabled || !v.LivePatch {
		return err
	}
	cycle, _ := s.suggestCycle(ctx)
	for _, viewer := range m.Players {
		have, _, err := s.listCycle(ctx, viewer)
		if err != nil {
			return err
		}
		if have != cycle {
			continue
		}
		if err := s.patchCoplay(ctx, v, viewer, cycle, m); err != nil {
			s.logger.Warn("Base friend suggestions for %s not patched: %v", viewer, err)
		}
	}
	return nil
}

func (s *service) patchCoplay(ctx context.Context, v suggestSettings, viewer, cycle string, m coplayMatch) error {
	excluded, err := s.suggestExcluded(ctx, viewer, cycle)
	if err != nil {
		return err
	}
	others := []string{}
	for _, p := range m.Players {
		if !excluded[p] {
			others = append(others, p)
		}
	}
	eligible, err := s.eligibleCandidates(ctx, v, others)
	if err != nil {
		return err
	}
	current, err := s.loadSuggestRows(ctx, viewer)
	if err != nil {
		return err
	}
	front := []suggestRow{}
	fresh := map[string]bool{}
	for _, p := range others {
		if eligible[p] {
			fresh[p] = true
			front = append(front, suggestRow{Candidate: p, Source: "coplay", Score: float64(m.EndedAt.UnixMilli()), LastPlayed: sql.NullTime{Time: m.EndedAt, Valid: true}})
		}
	}
	if len(front) == 0 {
		return nil
	}
	list := append([]suggestRow{}, front...)
	for _, r := range current {
		if !fresh[r.Candidate] {
			list = append(list, r)
		}
	}
	if limit := v.Visible + v.Reserve; len(list) > limit {
		list = list[:limit]
	}
	return s.saveSuggestions(ctx, viewer, cycle, "", list, false)
}

func (s *service) onSessionStart(ctx context.Context, _ runtime.Logger, _ *api.Event) {
	userID := ctxUser(ctx)
	if userID == "" {
		return
	}
	day, _ := dayBounds(s.now(), projectZone(ctx, s.db))
	if _, err := s.db.ExecContext(ctx, `INSERT INTO base_presence_day (user_id, day) VALUES ($1, $2::date) ON CONFLICT DO NOTHING`, userID, day); err != nil {
		s.logger.Warn("Base presence day not saved: %v", err)
	}
}

func (s *service) registerSuggestionHooks(initializer runtime.Initializer) error {
	return initializer.RegisterEventSessionStart(s.onSessionStart)
}

func (s *service) suggestSettingsGet(w http.ResponseWriter, r *http.Request) {
	v, meta, err := suggestStore.Read(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	raw, _ := json.Marshal(v)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	out["updatedBy"], out["updatedAt"] = meta.UpdatedBy, meta.UpdatedAt
	writeJSON(w, http.StatusOK, out)
}

func (s *service) suggestSettingsPut(w http.ResponseWriter, r *http.Request) {
	in := defaultSuggestSettings()
	var raw map[string]json.RawMessage
	if err := readBody(r, &raw); err != nil {
		s.writeError(w, err)
		return
	}
	delete(raw, "updatedBy")
	delete(raw, "updatedAt")
	b, _ := json.Marshal(raw)
	if err := json.Unmarshal(b, &in); err != nil {
		s.writeError(w, badRequest("INVALID_JSON", "", "settings are not valid"))
		return
	}
	v, err := in.clean()
	if err == nil && v.LevelTable != "" {
		var ok bool
		if err = s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM economy_level_table WHERE key=$1)`, v.LevelTable).Scan(&ok); err == nil && !ok {
			err = badRequest("LEVEL_TABLE_NOT_FOUND", "levelTable", "no level table with this key")
		}
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := suggestStore.Save(r.Context(), s.db, v, consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	s.suggestSettingsGet(w, r)
}

func (s *service) playerSuggestions(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	ctx := r.Context()
	if err := s.userExists(ctx, s.db, id); err != nil {
		s.writeError(w, err)
		return
	}
	v, err := suggestStore.Load(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	out := map[string]any{"cycle": "", "impressionCycle": "", "generatedAt": nil, "generatedBy": "", "visible": []suggestUser{}, "reserve": []suggestUser{}, "coplay": []any{}}
	var cycle, imp, by string
	var at time.Time
	err = s.db.QueryRowContext(ctx, `SELECT cycle, impression_cycle, generated_at, generated_by FROM base_friend_suggest_list WHERE user_id=$1`, id).Scan(&cycle, &imp, &at, &by)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, err)
		return
	}
	if err == nil {
		out["cycle"], out["impressionCycle"], out["generatedAt"], out["generatedBy"] = cycle, imp, at.UTC().Format(time.RFC3339), by
	}
	rows, err := s.loadSuggestRows(ctx, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	users, err := s.suggestUsers(ctx, rows, true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	cut := min(v.Visible, len(users))
	out["visible"], out["reserve"] = users[:cut], users[cut:]
	cp, err := s.db.QueryContext(ctx, `SELECT jsonb_build_object('userId',c.other_id,'username',COALESCE(u.username,''),'matches',c.matches,'lastAt',c.last_at,'gameKey',c.last_game_key,'matchId',c.last_match_id)::text
		FROM base_coplay c LEFT JOIN users u ON u.id=c.other_id WHERE c.user_id=$1 ORDER BY c.last_at DESC LIMIT 20`, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	list, err := rawRows(cp)
	if err != nil {
		s.writeError(w, err)
		return
	}
	out["coplay"] = list
	current, _ := s.suggestCycle(ctx)
	out["currentCycle"] = current
	writeJSON(w, http.StatusOK, out)
}

func (s *service) playerSuggestionsRegenerate(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if err := s.userExists(r.Context(), s.db, id); err != nil {
		s.writeError(w, err)
		return
	}
	who := consoleUser(r)
	if who == "" {
		who = "console"
	}
	if err := s.generateSuggestions(r.Context(), id, who); err != nil {
		s.writeError(w, err)
		return
	}
	s.playerSuggestions(w, r)
}
