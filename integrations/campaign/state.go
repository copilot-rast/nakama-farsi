package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
)

type userStage struct {
	BestStars   int              `json:"bestStars"`
	Claimed     map[string]int64 `json:"claimed"`
	ClaimedXP   int              `json:"claimedXp"`
	TiersPaid   int              `json:"tiersPaid"`
	Completions int              `json:"completions"`
	Wins        int              `json:"wins"`
	BestSeconds *float64         `json:"bestSeconds"`
	LastAt      *time.Time       `json:"lastCompletedAt"`
}

type player struct {
	s      *service
	q      queryer
	userID string
	cfg    settings.All
	cat    *catalog
	rows   map[string]*userStage
	levels map[string]int
	metas  map[string]float64
	items  map[string]bool
	groups map[string]groupState
}

func (s *service) loadPlayer(ctx context.Context, q queryer, userID string, cfg settings.All, cat *catalog) (*player, error) {
	p := &player{s: s, q: q, userID: userID, cfg: cfg, cat: cat, rows: map[string]*userStage{}, levels: map[string]int{}, metas: map[string]float64{}, items: map[string]bool{}, groups: map[string]groupState{}}
	rows, err := q.QueryContext(ctx, `SELECT stage_id::text, best_stars, claimed, claimed_xp, tiers_paid, completions, wins, best_seconds::float8, last_completed_at FROM campaign_user_stage WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var raw []byte
		var best sql.NullFloat64
		var at sql.NullTime
		u := &userStage{Claimed: map[string]int64{}}
		if err := rows.Scan(&id, &u.BestStars, &raw, &u.ClaimedXP, &u.TiersPaid, &u.Completions, &u.Wins, &best, &at); err != nil {
			return nil, err
		}
		unmarshal(raw, &u.Claimed)
		if best.Valid {
			v := best.Float64
			u.BestSeconds = &v
		}
		if at.Valid {
			v := at.Time
			u.LastAt = &v
		}
		p.rows[id] = u
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	grows, err := q.QueryContext(ctx, `SELECT group_key, last_at, day, count FROM campaign_reward_group WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer grows.Close()
	for grows.Next() {
		var key string
		var at sql.NullTime
		var st groupState
		if err := grows.Scan(&key, &at, &st.Day, &st.Count); err != nil {
			return nil, err
		}
		if at.Valid {
			v := at.Time
			st.LastAt = &v
		}
		p.groups[key] = st
	}
	return p, grows.Err()
}

func (p *player) row(stageID string) *userStage {
	if r := p.rows[stageID]; r != nil {
		return r
	}
	return &userStage{Claimed: map[string]int64{}}
}

func (p *player) stars(stageKey string) int {
	st := p.cat.byKey[stageKey]
	if st == nil {
		return 0
	}
	return p.row(st.ID).BestStars
}

func (p *player) level(ctx context.Context, table string) int {
	if v, ok := p.levels[table]; ok {
		return v
	}
	id := ""
	if table != "" {
		if err := p.q.QueryRowContext(ctx, `SELECT id::text FROM economy_level_table WHERE key=$1`, table).Scan(&id); err != nil {
			p.levels[table] = 0
			return 0
		}
	}
	lv, err := p.s.level(ctx, p.userID, id)
	if err != nil && !errors.Is(err, economyhooks.ErrNoLevels) && p.s.logger != nil {
		p.s.logger.Warn("Campaign could not read level %q: %v", table, err)
	}
	p.levels[table] = lv
	return lv
}

func (p *player) meta(ctx context.Context, key string) float64 {
	if v, ok := p.metas[key]; ok {
		return v
	}
	var v float64
	_ = p.q.QueryRowContext(ctx, `SELECT COALESCE((SELECT um.value FROM economy_user_meta um WHERE um.user_id=$1 AND um.meta_id=m.id), m.default_value)::float8 FROM economy_meta m WHERE m.key=$2`, p.userID, key).Scan(&v)
	p.metas[key] = v
	return v
}

func (p *player) owns(ctx context.Context, key string) bool {
	if v, ok := p.items[key]; ok {
		return v
	}
	var v bool
	_ = p.q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_item i WHERE i.key=$2 AND (i.is_starter OR EXISTS (SELECT 1 FROM economy_user_item u WHERE u.user_id=$1 AND u.item_id=i.id)))`, p.userID, key).Scan(&v)
	p.items[key] = v
	return v
}

func (p *player) countedStages(ch *chapter) []*stage {
	var out []*stage
	for _, st := range p.cat.stagesOf(ch.ID) {
		if counted(st.Mode, p.cfg.Rewards.MilestoneModes) {
			out = append(out, st)
		}
	}
	return out
}

func (p *player) chapterStars(ch *chapter) (int, int, float64) {
	list := p.countedStages(ch)
	sum := 0
	for _, st := range list {
		sum += p.row(st.ID).BestStars
	}
	return sum, len(list) * p.cfg.Stars.MaxStars, chapterPercent(sum, len(list), p.cfg.Stars.MaxStars)
}

func (p *player) chapterPassed(ch *chapter) bool {
	list := p.countedStages(ch)
	for _, st := range list {
		if p.row(st.ID).BestStars < p.cfg.Unlock.MinStars {
			return false
		}
	}
	return len(list) > 0
}

func (p *player) stageRules(st *stage) []Rule {
	if st.OwnUnlock {
		return st.Unlock
	}
	if st.Mode != settings.ModeOrdered || p.cfg.Unlock.StageDefault == settings.UnlockNone {
		return nil
	}
	var earlier []*stage
	for _, o := range p.cat.stagesOf(st.ChapterID) {
		if o.ID == st.ID {
			break
		}
		if o.Mode == settings.ModeOrdered {
			earlier = append(earlier, o)
		}
	}
	if len(earlier) == 0 {
		return nil
	}
	if p.cfg.Unlock.StageDefault == settings.UnlockPrevious {
		earlier = earlier[len(earlier)-1:]
	}
	out := make([]Rule, 0, len(earlier))
	for _, o := range earlier {
		out = append(out, Rule{Type: "stage", Key: o.Key})
	}
	return out
}

func (p *player) chapterRules(ch *chapter) []Rule {
	if len(ch.Unlock) > 0 || p.cfg.Unlock.ChapterDefault != settings.UnlockPrevious {
		return ch.Unlock
	}
	var prev *chapter
	for _, c := range p.cat.chapters {
		if c.ID == ch.ID {
			break
		}
		prev = c
	}
	if prev == nil {
		return nil
	}
	return []Rule{{Type: "chapter", Key: prev.Key}}
}

func (p *player) holds(ctx context.Context, r Rule) bool {
	switch r.Type {
	case "level":
		return float64(p.level(ctx, r.Table)) >= r.Value
	case "meta":
		return p.meta(ctx, r.Key) >= r.Value
	case "item":
		return p.owns(ctx, r.Key)
	case "stage":
		need := int(r.Value)
		if need <= 0 {
			need = p.cfg.Unlock.MinStars
		}
		return p.stars(r.Key) >= need
	case "chapter":
		ch := p.cat.chKey[r.Key]
		if ch == nil {
			return false
		}
		if r.Value <= 0 {
			return p.chapterPassed(ch)
		}
		_, _, pct := p.chapterStars(ch)
		return pct >= r.Value
	}
	return false
}

func (p *player) unmet(ctx context.Context, rules []Rule) []Rule {
	out := []Rule{}
	for _, r := range rules {
		if !p.holds(ctx, r) {
			out = append(out, r)
		}
	}
	return out
}

func (p *player) stageLocks(ctx context.Context, st *stage) []Rule {
	out := []Rule{}
	if ch := p.cat.chapter[st.ChapterID]; ch != nil {
		out = append(out, p.unmet(ctx, p.chapterRules(ch))...)
	}
	return append(out, p.unmet(ctx, p.stageRules(st))...)
}

func (p *player) group(key string) groupState { return p.groups[key] }
