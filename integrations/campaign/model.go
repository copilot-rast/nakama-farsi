package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"

	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
)

type Rule struct {
	Type  string  `json:"type"`
	Key   string  `json:"key,omitempty"`
	Value float64 `json:"value,omitempty"`
	Table string  `json:"table,omitempty"`
}

var ruleTypes = map[string]bool{"level": true, "meta": true, "item": true, "stage": true, "chapter": true}

type Reward struct {
	Currencies    map[string]int64 `json:"currencies,omitempty"`
	PrizeKey      string           `json:"prizeKey,omitempty"`
	PrizeGroupKey string           `json:"prizeGroupKey,omitempty"`
}

func (r Reward) empty() bool {
	for _, v := range r.Currencies {
		if v > 0 {
			return false
		}
	}
	return r.PrizeKey == "" && r.PrizeGroupKey == ""
}

type Milestone struct {
	ID      string  `json:"id"`
	Percent float64 `json:"percent"`
	Reward
}

type TierPrize struct {
	Stars         int    `json:"stars"`
	PrizeKey      string `json:"prizeKey,omitempty"`
	PrizeGroupKey string `json:"prizeGroupKey,omitempty"`
}

type chapter struct {
	ID          string            `json:"id"`
	Key         string            `json:"key"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	IconURL     string            `json:"iconUrl"`
	Address     string            `json:"address"`
	Sort        int               `json:"sort"`
	Active      bool              `json:"isActive"`
	Unlock      []Rule            `json:"unlock"`
	Milestones  []Milestone       `json:"milestones"`
	Metadata    map[string]any    `json:"metadata"`
	Settings    map[string]any    `json:"settings"`
	UpdatedBy   string            `json:"updatedBy"`
	UpdatedAt   string            `json:"updatedAt"`
	StageCount  int               `json:"stageCount"`
}

type stage struct {
	ID            string              `json:"id"`
	Key           string              `json:"key"`
	ChapterID     string              `json:"chapterId"`
	ChapterKey    string              `json:"chapterKey"`
	Name          map[string]string   `json:"name"`
	Description   map[string]string   `json:"description"`
	IconURL       string              `json:"iconUrl"`
	Address       string              `json:"address"`
	Sort          int                 `json:"sort"`
	Active        bool                `json:"isActive"`
	Mode          string              `json:"mode"`
	Unlock        []Rule              `json:"unlock"`
	OwnUnlock     bool                `json:"ownUnlock"`
	TimeLimit     *float64            `json:"timeLimitSeconds"`
	MinSeconds    *float64            `json:"minSeconds"`
	StarRules     []settings.StarRule `json:"starRules"`
	OwnRules      bool                `json:"ownStarRules"`
	StarTemplate  string              `json:"starTemplate"`
	NoReplay      bool                `json:"noReplayAfterWin"`
	Multipliers   []float64           `json:"starMultipliers"`
	OwnMultiplier bool                `json:"ownMultipliers"`
	Rewards       Reward              `json:"rewards"`
	XP            int                 `json:"xp"`
	TierPrizes    []TierPrize         `json:"tierPrizes"`
	RewardGroup   string              `json:"rewardGroup"`
	Cooldown      int                 `json:"cooldownSeconds"`
	DailyLimit    int                 `json:"dailyLimit"`
	XPWhenLimited bool                `json:"xpWhenLimited"`
	Metadata      map[string]any      `json:"metadata"`
	Settings      map[string]any      `json:"settings"`
	UpdatedBy     string              `json:"updatedBy"`
	UpdatedAt     string              `json:"updatedAt"`
}

func (s *stage) group() string {
	if s.RewardGroup != "" {
		return s.RewardGroup
	}
	return "stage:" + s.Key
}

type freeride struct {
	ID          string            `json:"id"`
	Key         string            `json:"key"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	IconURL     string            `json:"iconUrl"`
	Sort        int               `json:"sort"`
	Active      bool              `json:"isActive"`
	Rewards     Reward            `json:"rewards"`
	XP          int               `json:"xp"`
	RewardGroup string            `json:"rewardGroup"`
	Cooldown    int               `json:"cooldownSeconds"`
	DailyLimit  int               `json:"dailyLimit"`
	MinLevel    int               `json:"minLevel"`
	UpdatedBy   string            `json:"updatedBy"`
	UpdatedAt   string            `json:"updatedAt"`
}

func (f *freeride) group() string {
	if f.RewardGroup != "" {
		return f.RewardGroup
	}
	return "freeride:" + f.Key
}

type catalog struct {
	chapters []*chapter
	stages   []*stage
	byKey    map[string]*stage
	chapter  map[string]*chapter
	chKey    map[string]*chapter
}

func (c *catalog) stagesOf(chapterID string) []*stage {
	var out []*stage
	for _, s := range c.stages {
		if s.ChapterID == chapterID {
			out = append(out, s)
		}
	}
	return out
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func unmarshal(raw []byte, out any) {
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, out)
	}
}

const chapterCols = `c.id::text, c.key, c.name, c.description, c.icon_url, c.address, c.sort, c.is_active, c.unlock, c.milestones, c.metadata, c.settings, c.updated_by, c.updated_at,
	(SELECT count(*) FROM campaign_stage s WHERE s.chapter_id=c.id)`

func scanChapter(sc interface{ Scan(...any) error }) (*chapter, error) {
	c := &chapter{}
	var name, desc, unlock, ms, md, set []byte
	var at sql.NullTime
	if err := sc.Scan(&c.ID, &c.Key, &name, &desc, &c.IconURL, &c.Address, &c.Sort, &c.Active, &unlock, &ms, &md, &set, &c.UpdatedBy, &at, &c.StageCount); err != nil {
		return nil, err
	}
	c.Name, c.Description, c.Unlock, c.Milestones, c.Metadata, c.Settings = map[string]string{}, map[string]string{}, []Rule{}, []Milestone{}, map[string]any{}, map[string]any{}
	unmarshal(set, &c.Settings)
	if c.Settings == nil {
		c.Settings = map[string]any{}
	}
	unmarshal(name, &c.Name)
	unmarshal(desc, &c.Description)
	unmarshal(unlock, &c.Unlock)
	unmarshal(ms, &c.Milestones)
	if c.Milestones == nil {
		c.Milestones = []Milestone{}
	}
	if c.Unlock == nil {
		c.Unlock = []Rule{}
	}
	unmarshal(md, &c.Metadata)
	if at.Valid {
		c.UpdatedAt = at.Time.UTC().Format("2006-01-02T15:04:05Z")
	}
	return c, nil
}

const stageCols = `s.id::text, s.key, s.chapter_id::text, c.key, s.name, s.description, s.icon_url, s.address, s.sort, s.is_active, s.mode, s.unlock, s.time_limit_seconds::float8, s.min_seconds::float8,
	s.star_rules, s.star_multipliers, s.rewards, s.xp, s.tier_prizes, s.reward_group, s.cooldown_seconds, s.daily_limit, s.xp_when_limited, s.metadata, s.settings, s.star_template, s.no_replay_after_win, s.updated_by, s.updated_at`

func scanStage(sc interface{ Scan(...any) error }) (*stage, error) {
	s := &stage{}
	var name, desc, unlock, rules, mult, rewards, tiers, md, set []byte
	var limit, minSec sql.NullFloat64
	var at sql.NullTime
	if err := sc.Scan(&s.ID, &s.Key, &s.ChapterID, &s.ChapterKey, &name, &desc, &s.IconURL, &s.Address, &s.Sort, &s.Active, &s.Mode, &unlock, &limit, &minSec,
		&rules, &mult, &rewards, &s.XP, &tiers, &s.RewardGroup, &s.Cooldown, &s.DailyLimit, &s.XPWhenLimited, &md, &set, &s.StarTemplate, &s.NoReplay, &s.UpdatedBy, &at); err != nil {
		return nil, err
	}
	s.Name, s.Description, s.Metadata, s.TierPrizes, s.Settings = map[string]string{}, map[string]string{}, map[string]any{}, []TierPrize{}, map[string]any{}
	unmarshal(set, &s.Settings)
	if s.Settings == nil {
		s.Settings = map[string]any{}
	}
	unmarshal(name, &s.Name)
	unmarshal(desc, &s.Description)
	unmarshal(md, &s.Metadata)
	unmarshal(tiers, &s.TierPrizes)
	if s.TierPrizes == nil {
		s.TierPrizes = []TierPrize{}
	}
	unmarshal(rewards, &s.Rewards)
	if unlock != nil {
		s.OwnUnlock = true
		s.Unlock = []Rule{}
		unmarshal(unlock, &s.Unlock)
	}
	if rules != nil {
		s.OwnRules = true
		unmarshal(rules, &s.StarRules)
	}
	if mult != nil {
		s.OwnMultiplier = true
		unmarshal(mult, &s.Multipliers)
	}
	if limit.Valid {
		v := limit.Float64
		s.TimeLimit = &v
	}
	if minSec.Valid {
		v := minSec.Float64
		s.MinSeconds = &v
	}
	if at.Valid {
		s.UpdatedAt = at.Time.UTC().Format("2006-01-02T15:04:05Z")
	}
	return s, nil
}

func loadCatalog(ctx context.Context, q queryer, activeOnly bool) (*catalog, error) {
	where := ""
	if activeOnly {
		where = ` WHERE c.is_active`
	}
	rows, err := q.QueryContext(ctx, `SELECT `+chapterCols+` FROM campaign_chapter c`+where+` ORDER BY c.sort, c.key`)
	if err != nil {
		return nil, err
	}
	cat := &catalog{byKey: map[string]*stage{}, chapter: map[string]*chapter{}, chKey: map[string]*chapter{}}
	for rows.Next() {
		c, err := scanChapter(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		cat.chapters = append(cat.chapters, c)
		cat.chapter[c.ID], cat.chKey[c.Key] = c, c
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	swhere := ""
	if activeOnly {
		swhere = ` WHERE s.is_active AND c.is_active`
	}
	rows, err = q.QueryContext(ctx, `SELECT `+stageCols+` FROM campaign_stage s JOIN campaign_chapter c ON c.id=s.chapter_id`+swhere+` ORDER BY c.sort, c.key, s.sort, s.key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		s, err := scanStage(rows)
		if err != nil {
			return nil, err
		}
		cat.stages = append(cat.stages, s)
		cat.byKey[s.Key] = s
	}
	sort.SliceStable(cat.stages, func(i, j int) bool {
		ci, cj := cat.chapter[cat.stages[i].ChapterID], cat.chapter[cat.stages[j].ChapterID]
		if ci != nil && cj != nil && ci != cj {
			if ci.Sort != cj.Sort {
				return ci.Sort < cj.Sort
			}
			return ci.Key < cj.Key
		}
		return false
	})
	return cat, rows.Err()
}

const freerideCols = `id::text, key, name, description, icon_url, sort, is_active, rewards, prize_key, prize_group_key, xp, reward_group, cooldown_seconds, daily_limit, min_level, updated_by, updated_at`

func scanFreeride(sc interface{ Scan(...any) error }) (*freeride, error) {
	f := &freeride{}
	var name, desc, rewards []byte
	var at sql.NullTime
	if err := sc.Scan(&f.ID, &f.Key, &name, &desc, &f.IconURL, &f.Sort, &f.Active, &rewards, &f.Rewards.PrizeKey, &f.Rewards.PrizeGroupKey, &f.XP, &f.RewardGroup, &f.Cooldown, &f.DailyLimit, &f.MinLevel, &f.UpdatedBy, &at); err != nil {
		return nil, err
	}
	f.Name, f.Description = map[string]string{}, map[string]string{}
	unmarshal(name, &f.Name)
	unmarshal(desc, &f.Description)
	var r Reward
	unmarshal(rewards, &r)
	f.Rewards.Currencies = r.Currencies
	if at.Valid {
		f.UpdatedAt = at.Time.UTC().Format("2006-01-02T15:04:05Z")
	}
	return f, nil
}
