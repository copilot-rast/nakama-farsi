package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/cup/internal/settings"
)

type queryer interface {
	QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
}

type cupRow struct {
	ID           string
	TemplateID   string
	TemplateKey  string
	Name         map[string]string
	Description  map[string]string
	IconURL      string
	Status       string
	AnnounceAt   *time.Time
	SignupAt     time.Time
	StartsAt     time.Time
	Rules        settings.Rules
	Size         int
	Rounds       int
	CurrentRound int
	Players      int
	Pool         poolValue
	Champion     string
	CancelReason string
	CreatedBy    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	StartedAt    *time.Time
	FinishedAt   *time.Time
}

type poolValue struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
	Fees        int64  `json:"fees"`
}

const cupCols = `id::text, COALESCE(template_id::text,''), template_key, name::text, description::text, icon_url, status, announce_at, signup_at, starts_at, rules::text,
	size, rounds, current_round, players, pool::text, COALESCE(champion::text,''), cancel_reason, created_by, created_at, updated_at, started_at, finished_at`

func scanCup(row interface{ Scan(...any) error }) (*cupRow, error) {
	c := &cupRow{}
	var name, desc, rules, pool string
	if err := row.Scan(&c.ID, &c.TemplateID, &c.TemplateKey, &name, &desc, &c.IconURL, &c.Status, &c.AnnounceAt, &c.SignupAt, &c.StartsAt, &rules,
		&c.Size, &c.Rounds, &c.CurrentRound, &c.Players, &pool, &c.Champion, &c.CancelReason, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &c.StartedAt, &c.FinishedAt); err != nil {
		return nil, err
	}
	c.Name, c.Description = textMap(name), textMap(desc)
	c.Rules = settings.DefaultRules()
	_ = json.Unmarshal([]byte(rules), &c.Rules)
	c.Rules.Normalize()
	_ = json.Unmarshal([]byte(pool), &c.Pool)
	return c, nil
}

func textMap(raw string) map[string]string {
	out := map[string]string{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func loadCup(ctx context.Context, q queryer, id string, lock string) (*cupRow, error) {
	if !uuidPattern.MatchString(id) {
		return nil, sql.ErrNoRows
	}
	return scanCup(q.QueryRowContext(ctx, `SELECT `+cupCols+` FROM cup WHERE id=$1::uuid `+lock, id))
}

type playerRow struct {
	UserID   string
	Status   string
	Seed     int
	Trophies int
	Place    int
	Reached  int
	Fee      feeValue
	FeeState string
	AddedBy  string
	JoinedAt time.Time
}

type feeValue struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
	Grant       string `json:"grant"`
	Refunded    int64  `json:"refunded"`
}

const playerCols = `user_id::text, status, seed, trophies, place, reached, fee::text, fee_state, added_by, joined_at`

func scanPlayer(row interface{ Scan(...any) error }) (*playerRow, error) {
	p := &playerRow{}
	var fee string
	if err := row.Scan(&p.UserID, &p.Status, &p.Seed, &p.Trophies, &p.Place, &p.Reached, &fee, &p.FeeState, &p.AddedBy, &p.JoinedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(fee), &p.Fee)
	return p, nil
}

func loadPlayers(ctx context.Context, q queryer, cupID string) ([]*playerRow, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+playerCols+` FROM cup_player WHERE cup_id=$1::uuid ORDER BY seed, joined_at, user_id`, cupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*playerRow{}
	for rows.Next() {
		p, err := scanPlayer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

type matchRow struct {
	ID          string
	Round       int
	Slot        int
	Kind        string
	Status      string
	BestOf      int
	A, B        string
	SeedA       int
	SeedB       int
	WinsA       int
	WinsB       int
	Draws       int
	Games       int
	ReadyA      *time.Time
	ReadyB      *time.Time
	ShowedA     bool
	ShowedB     bool
	OpensAt     *time.Time
	StartsAt    *time.Time
	ClosesAt    *time.Time
	Winner      string
	Loser       string
	Reason      string
	GameMatchID string
	Actor       string
	FinishedAt  *time.Time
	dirty       bool
}

const matchCols = `id::text, round, slot, kind, status, best_of, COALESCE(user_a::text,''), COALESCE(user_b::text,''), seed_a, seed_b, wins_a, wins_b, draws, games,
	ready_a, ready_b, showed_a, showed_b, opens_at, starts_at, closes_at, COALESCE(winner::text,''), COALESCE(loser::text,''), reason, game_match_id, actor, finished_at`

func scanMatch(row interface{ Scan(...any) error }) (*matchRow, error) {
	m := &matchRow{}
	if err := row.Scan(&m.ID, &m.Round, &m.Slot, &m.Kind, &m.Status, &m.BestOf, &m.A, &m.B, &m.SeedA, &m.SeedB, &m.WinsA, &m.WinsB, &m.Draws, &m.Games,
		&m.ReadyA, &m.ReadyB, &m.ShowedA, &m.ShowedB, &m.OpensAt, &m.StartsAt, &m.ClosesAt, &m.Winner, &m.Loser, &m.Reason, &m.GameMatchID, &m.Actor, &m.FinishedAt); err != nil {
		return nil, err
	}
	return m, nil
}

func loadMatches(ctx context.Context, q queryer, cupID string) ([]*matchRow, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+matchCols+` FROM cup_match WHERE cup_id=$1::uuid ORDER BY kind, round, slot`, cupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*matchRow{}
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func nullUUID(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func saveMatch(ctx context.Context, q queryer, m *matchRow) error {
	_, err := q.ExecContext(ctx, `UPDATE cup_match SET status=$2, user_a=$3::uuid, user_b=$4::uuid, seed_a=$5, seed_b=$6, wins_a=$7, wins_b=$8, draws=$9, games=$10,
		ready_a=$11, ready_b=$12, showed_a=$13, showed_b=$14, opens_at=$15, starts_at=$16, closes_at=$17, winner=$18::uuid, loser=$19::uuid, reason=$20, game_match_id=$21,
		actor=$22, finished_at=$23, best_of=$24, updated_at=now() WHERE id=$1::uuid`,
		m.ID, m.Status, nullUUID(m.A), nullUUID(m.B), m.SeedA, m.SeedB, m.WinsA, m.WinsB, m.Draws, m.Games, m.ReadyA, m.ReadyB, m.ShowedA, m.ShowedB,
		m.OpensAt, m.StartsAt, m.ClosesAt, nullUUID(m.Winner), nullUUID(m.Loser), m.Reason, m.GameMatchID, m.Actor, m.FinishedAt, m.BestOf)
	return err
}

type gameRow struct {
	ID          string
	MatchID     string
	CupID       string
	GameNo      int
	GameMatchID string
	Status      string
	Winner      string
	Draw        bool
	Score       map[string]int
	Present     []string
	Reason      string
	CreatedAt   time.Time
	FinishedAt  *time.Time
}

const gameCols = `id::text, match_id::text, cup_id::text, game_no, game_match_id, status, COALESCE(winner::text,''), draw, score::text, present::text, reason, created_at, finished_at`

func scanGame(row interface{ Scan(...any) error }) (*gameRow, error) {
	g := &gameRow{}
	var score, present string
	if err := row.Scan(&g.ID, &g.MatchID, &g.CupID, &g.GameNo, &g.GameMatchID, &g.Status, &g.Winner, &g.Draw, &score, &present, &g.Reason, &g.CreatedAt, &g.FinishedAt); err != nil {
		return nil, err
	}
	g.Score, g.Present = map[string]int{}, []string{}
	_ = json.Unmarshal([]byte(score), &g.Score)
	_ = json.Unmarshal([]byte(present), &g.Present)
	return g, nil
}

func loadGames(ctx context.Context, q queryer, cupID string) (map[string][]*gameRow, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+gameCols+` FROM cup_game WHERE cup_id=$1::uuid ORDER BY match_id, game_no`, cupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]*gameRow{}
	for rows.Next() {
		g, err := scanGame(rows)
		if err != nil {
			return nil, err
		}
		out[g.MatchID] = append(out[g.MatchID], g)
	}
	return out, rows.Err()
}

func jsonOf(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}

func iso(t *time.Time) any {
	if t == nil || t.IsZero() {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

func isoT(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func ptr(t time.Time) *time.Time { return &t }

func feeFrom(raw string) feeValue {
	var f feeValue
	_ = json.Unmarshal([]byte(raw), &f)
	return f
}

type payoutItem struct {
	Kind     string           `json:"kind"`
	PrizeKey string           `json:"prizeKey"`
	Amounts  map[string]int64 `json:"amounts"`
	Status   string           `json:"status"`
}

func payoutList(raw string) []payoutItem {
	out := []payoutItem{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func jsonUnmarshal(raw string, out any) error { return json.Unmarshal([]byte(raw), out) }
