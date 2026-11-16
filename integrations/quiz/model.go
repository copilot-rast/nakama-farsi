package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
)

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

type Option struct {
	ID       int               `json:"id"`
	Text     map[string]string `json:"text"`
	MediaURL string            `json:"mediaUrl,omitempty"`
	Correct  bool              `json:"correct"`
}

type mission struct {
	ID          string            `json:"id"`
	Key         string            `json:"key"`
	Kind        string            `json:"kind"`
	ParentID    string            `json:"parentId"`
	ParentKey   string            `json:"parentKey"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	IconURL     string            `json:"iconUrl"`
	Sort        int               `json:"sort"`
	OrderMode   string            `json:"orderMode"`
	TimeSeconds int               `json:"timeSeconds"`
	SendAnswers bool              `json:"sendAnswers"`
	Reward      Reward            `json:"reward"`
	Active      bool              `json:"isActive"`
	Metadata    map[string]any    `json:"metadata"`
	UpdatedBy   string            `json:"updatedBy"`
	UpdatedAt   string            `json:"updatedAt"`
	Items       int               `json:"items"`
}

type pack struct {
	ID               string            `json:"id"`
	Key              string            `json:"key"`
	MissionID        string            `json:"missionId"`
	MissionKey       string            `json:"missionKey"`
	Name             map[string]string `json:"name"`
	Description      map[string]string `json:"description"`
	IconURL          string            `json:"iconUrl"`
	Sort             int               `json:"sort"`
	Category         string            `json:"category"`
	Tags             []string          `json:"tags"`
	Language         string            `json:"language"`
	Difficulty       int               `json:"difficulty"`
	TimeSeconds      int               `json:"timeSeconds"`
	QuestionSeconds  int               `json:"questionSeconds"`
	PassPercent      *float64          `json:"passPercent"`
	PassCorrect      *int              `json:"passCorrect"`
	QuestionCount    int               `json:"questionCount"`
	PoolCategory     string            `json:"poolCategory"`
	ShuffleQuestions *bool             `json:"shuffleQuestions"`
	ShuffleOptions   *bool             `json:"shuffleOptions"`
	SendAnswers      bool              `json:"sendAnswers"`
	Reward           Reward            `json:"reward"`
	CampaignStageKey string            `json:"campaignStageKey"`
	Active           bool              `json:"isActive"`
	Metadata         map[string]any    `json:"metadata"`
	UpdatedBy        string            `json:"updatedBy"`
	UpdatedAt        string            `json:"updatedAt"`
	Questions        int               `json:"questions"`
}

type question struct {
	ID          string              `json:"id"`
	Key         string              `json:"key"`
	PackID      string              `json:"packId"`
	PackKey     string              `json:"packKey"`
	Type        string              `json:"type"`
	Question    map[string]string   `json:"question"`
	MediaURL    string              `json:"mediaUrl"`
	Options     []Option            `json:"options"`
	Answers     map[string][]string `json:"answers"`
	Explanation map[string]string   `json:"explanation"`
	Difficulty  int                 `json:"difficulty"`
	Category    string              `json:"category"`
	Tags        []string            `json:"tags"`
	TimeSeconds *int                `json:"timeSeconds"`
	Reward      Reward              `json:"reward"`
	Sort        int                 `json:"sort"`
	Active      bool                `json:"isActive"`
	Metadata    map[string]any      `json:"metadata"`
	UpdatedBy   string              `json:"updatedBy"`
	UpdatedAt   string              `json:"updatedAt"`
	Hints       int                 `json:"hints"`
	Reports     int                 `json:"openReports"`
}

type hint struct {
	ID         string            `json:"id"`
	QuestionID string            `json:"questionId"`
	Name       map[string]string `json:"name"`
	Text       map[string]string `json:"text"`
	MediaURL   string            `json:"mediaUrl"`
	Price      settings.Price    `json:"price"`
	Sort       int               `json:"sort"`
	Active     bool              `json:"isActive"`
	UpdatedAt  string            `json:"updatedAt"`
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type scanner interface{ Scan(...any) error }

func unmarshal(raw []byte, out any) {
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, out)
	}
}

func jsonOf(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func stamp(at sql.NullTime) string {
	if at.Valid {
		return at.Time.UTC().Format(time.RFC3339)
	}
	return ""
}

const missionCols = `m.id::text, m.key, m.kind, COALESCE(m.parent_id::text,''), COALESCE(p.key,''), m.name, m.description, m.icon_url, m.sort, m.order_mode, m.time_seconds,
	m.send_answers, m.reward, m.is_active, m.metadata, m.updated_by, m.updated_at,
	CASE WHEN m.kind='game' THEN (SELECT count(*) FROM quiz_mission c WHERE c.parent_id=m.id) ELSE (SELECT count(*) FROM quiz_pack k WHERE k.mission_id=m.id) END`

const missionFrom = ` FROM quiz_mission m LEFT JOIN quiz_mission p ON p.id=m.parent_id`

func scanMission(sc scanner) (*mission, error) {
	m := &mission{}
	var name, desc, reward, md []byte
	var at sql.NullTime
	if err := sc.Scan(&m.ID, &m.Key, &m.Kind, &m.ParentID, &m.ParentKey, &name, &desc, &m.IconURL, &m.Sort, &m.OrderMode, &m.TimeSeconds, &m.SendAnswers, &reward, &m.Active, &md, &m.UpdatedBy, &at, &m.Items); err != nil {
		return nil, err
	}
	m.Name, m.Description, m.Metadata = map[string]string{}, map[string]string{}, map[string]any{}
	unmarshal(name, &m.Name)
	unmarshal(desc, &m.Description)
	unmarshal(reward, &m.Reward)
	unmarshal(md, &m.Metadata)
	m.UpdatedAt = stamp(at)
	return m, nil
}

const packCols = `k.id::text, k.key, COALESCE(k.mission_id::text,''), COALESCE(m.key,''), k.name, k.description, k.icon_url, k.sort, k.category, array_to_json(k.tags)::text, k.language, k.difficulty,
	k.time_seconds, k.question_seconds, k.pass_percent::float8, k.pass_correct, k.question_count, k.pool_category, k.shuffle_questions, k.shuffle_options, k.send_answers,
	k.reward, k.campaign_stage_key, k.is_active, k.metadata, k.updated_by, k.updated_at,
	CASE WHEN k.pool_category <> '' THEN (SELECT count(*) FROM quiz_question q WHERE q.category=k.pool_category AND q.is_active) ELSE (SELECT count(*) FROM quiz_question q WHERE q.pack_id=k.id AND q.is_active) END`

const packFrom = ` FROM quiz_pack k LEFT JOIN quiz_mission m ON m.id=k.mission_id`

func scanPack(sc scanner) (*pack, error) {
	k := &pack{}
	var name, desc, reward, md []byte
	var tags []byte
	var pct sql.NullFloat64
	var pc sql.NullInt64
	var sq, so sql.NullBool
	var at sql.NullTime
	if err := sc.Scan(&k.ID, &k.Key, &k.MissionID, &k.MissionKey, &name, &desc, &k.IconURL, &k.Sort, &k.Category, &tags, &k.Language, &k.Difficulty,
		&k.TimeSeconds, &k.QuestionSeconds, &pct, &pc, &k.QuestionCount, &k.PoolCategory, &sq, &so, &k.SendAnswers,
		&reward, &k.CampaignStageKey, &k.Active, &md, &k.UpdatedBy, &at, &k.Questions); err != nil {
		return nil, err
	}
	k.Name, k.Description, k.Metadata, k.Tags = map[string]string{}, map[string]string{}, map[string]any{}, []string{}
	unmarshal(tags, &k.Tags)
	if k.Tags == nil {
		k.Tags = []string{}
	}
	unmarshal(name, &k.Name)
	unmarshal(desc, &k.Description)
	unmarshal(reward, &k.Reward)
	unmarshal(md, &k.Metadata)
	if pct.Valid {
		v := pct.Float64
		k.PassPercent = &v
	}
	if pc.Valid {
		v := int(pc.Int64)
		k.PassCorrect = &v
	}
	if sq.Valid {
		v := sq.Bool
		k.ShuffleQuestions = &v
	}
	if so.Valid {
		v := so.Bool
		k.ShuffleOptions = &v
	}
	k.UpdatedAt = stamp(at)
	return k, nil
}

const questionCols = `q.id::text, COALESCE(q.key,''), COALESCE(q.pack_id::text,''), COALESCE(k.key,''), q.type, q.question, q.media_url, q.options, q.answers, q.explanation, q.difficulty,
	q.category, array_to_json(q.tags)::text, q.time_seconds, q.reward, q.sort, q.is_active, q.metadata, q.updated_by, q.updated_at,
	(SELECT count(*) FROM quiz_hint h WHERE h.question_id=q.id), (SELECT count(*) FROM quiz_question_report r WHERE r.question_id=q.id AND r.status='open')`

const questionFrom = ` FROM quiz_question q LEFT JOIN quiz_pack k ON k.id=q.pack_id`

func scanQuestion(sc scanner) (*question, error) {
	q := &question{}
	var text, opts, answers, expl, reward, md []byte
	var tags []byte
	var ts sql.NullInt64
	var at sql.NullTime
	if err := sc.Scan(&q.ID, &q.Key, &q.PackID, &q.PackKey, &q.Type, &text, &q.MediaURL, &opts, &answers, &expl, &q.Difficulty,
		&q.Category, &tags, &ts, &reward, &q.Sort, &q.Active, &md, &q.UpdatedBy, &at, &q.Hints, &q.Reports); err != nil {
		return nil, err
	}
	q.Question, q.Explanation, q.Answers, q.Metadata, q.Options, q.Tags = map[string]string{}, map[string]string{}, map[string][]string{}, map[string]any{}, []Option{}, []string{}
	unmarshal(tags, &q.Tags)
	if q.Tags == nil {
		q.Tags = []string{}
	}
	unmarshal(text, &q.Question)
	unmarshal(opts, &q.Options)
	unmarshal(answers, &q.Answers)
	unmarshal(expl, &q.Explanation)
	unmarshal(reward, &q.Reward)
	unmarshal(md, &q.Metadata)
	if q.Options == nil {
		q.Options = []Option{}
	}
	if ts.Valid {
		v := int(ts.Int64)
		q.TimeSeconds = &v
	}
	q.UpdatedAt = stamp(at)
	return q, nil
}

const hintCols = `h.id::text, h.question_id::text, h.name, h.text, h.media_url, h.price, h.sort, h.is_active, h.updated_at`

func scanHint(sc scanner) (*hint, error) {
	h := &hint{}
	var name, text, price []byte
	var at sql.NullTime
	if err := sc.Scan(&h.ID, &h.QuestionID, &name, &text, &h.MediaURL, &price, &h.Sort, &h.Active, &at); err != nil {
		return nil, err
	}
	h.Name, h.Text = map[string]string{}, map[string]string{}
	unmarshal(name, &h.Name)
	unmarshal(text, &h.Text)
	unmarshal(price, &h.Price)
	h.UpdatedAt = stamp(at)
	return h, nil
}

func loadPack(ctx context.Context, q queryer, where string, args ...any) (*pack, error) {
	return scanPack(q.QueryRowContext(ctx, `SELECT `+packCols+packFrom+` WHERE `+where, args...))
}

func loadMission(ctx context.Context, q queryer, where string, args ...any) (*mission, error) {
	return scanMission(q.QueryRowContext(ctx, `SELECT `+missionCols+missionFrom+` WHERE `+where, args...))
}

func listQuestions(ctx context.Context, q queryer, where string, args ...any) ([]*question, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+questionCols+questionFrom+` WHERE `+where+` ORDER BY q.sort, q.created_at, q.id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*question{}
	for rows.Next() {
		x, err := scanQuestion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func poolOf(ctx context.Context, q queryer, k *pack) ([]*question, error) {
	if k.PoolCategory != "" {
		return listQuestions(ctx, q, `q.category=$1 AND q.is_active`, k.PoolCategory)
	}
	return listQuestions(ctx, q, `q.pack_id=$1 AND q.is_active`, k.ID)
}

func listHints(ctx context.Context, q queryer, questionIDs []string, activeOnly bool) (map[string][]*hint, error) {
	out := map[string][]*hint{}
	if len(questionIDs) == 0 {
		return out, nil
	}
	where := ``
	if activeOnly {
		where = ` AND h.is_active`
	}
	rows, err := q.QueryContext(ctx, `SELECT `+hintCols+` FROM quiz_hint h WHERE h.question_id::text = ANY($1::text[])`+where+` ORDER BY h.sort, h.created_at`, questionIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		h, err := scanHint(rows)
		if err != nil {
			return nil, err
		}
		out[h.QuestionID] = append(out[h.QuestionID], h)
	}
	return out, rows.Err()
}

type picked struct {
	ID    string `json:"id"`
	Order []int  `json:"order,omitempty"`
}
