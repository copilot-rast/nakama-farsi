package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var schemaPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

type PG struct {
	db     *sql.DB
	kind   string
	schema string
	owned  bool
}

func NewNakamaPG(db *sql.DB) *PG { return &PG{db: db, kind: DriverNakama} }

func NewExternalPG(db *sql.DB, schema string) (*PG, error) {
	if schema == "" {
		schema = "public"
	}
	if !schemaPattern.MatchString(schema) {
		return nil, ErrBadQuery
	}
	return &PG{db: db, kind: DriverPostgres, schema: schema, owned: true}, nil
}

func (p *PG) Kind() string { return p.kind }

func (p *PG) t(name string) string {
	if p.schema == "" {
		return name
	}
	return `"` + p.schema + `".` + name
}

func PGSchema(schema string) []string {
	q := func(n string) string {
		if schema == "" {
			return n
		}
		return `"` + schema + `".` + n
	}
	ix := func(n string) string { return n }
	return []string{
		`CREATE TABLE IF NOT EXISTS ` + q("insights_event") + ` (
    id UUID PRIMARY KEY,
    game TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    value DOUBLE PRECISION,
    tags JSONB NOT NULL DEFAULT '{}',
    fields JSONB NOT NULL DEFAULT '{}',
    user_id TEXT NOT NULL DEFAULT '',
    session_id TEXT NOT NULL DEFAULT '',
    at TIMESTAMPTZ NOT NULL,
    day DATE NOT NULL,
    server_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    platform TEXT NOT NULL DEFAULT '',
    app_version TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'client',
    flagged BOOLEAN NOT NULL DEFAULT false,
    flag_reason TEXT NOT NULL DEFAULT ''
)`,
		`CREATE INDEX IF NOT EXISTS ` + ix("insights_event_name_day_idx") + ` ON ` + q("insights_event") + ` (name, day)`,
		`CREATE INDEX IF NOT EXISTS ` + ix("insights_event_user_at_idx") + ` ON ` + q("insights_event") + ` (user_id, at)`,
		`CREATE INDEX IF NOT EXISTS ` + ix("insights_event_day_idx") + ` ON ` + q("insights_event") + ` (day)`,
		`CREATE INDEX IF NOT EXISTS ` + ix("insights_event_session_idx") + ` ON ` + q("insights_event") + ` (session_id, at) WHERE session_id <> ''`,
		`CREATE INDEX IF NOT EXISTS ` + ix("insights_event_tags_idx") + ` ON ` + q("insights_event") + ` USING GIN (tags jsonb_path_ops)`,
		`CREATE TABLE IF NOT EXISTS ` + q("insights_event_user") + ` (
    game TEXT NOT NULL DEFAULT '',
    user_id TEXT NOT NULL,
    first_day DATE NOT NULL,
    first_at TIMESTAMPTZ NOT NULL,
    last_day DATE NOT NULL,
    platform TEXT NOT NULL DEFAULT '',
    app_version TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (game, user_id)
)`,
		`CREATE INDEX IF NOT EXISTS ` + ix("insights_event_user_first_idx") + ` ON ` + q("insights_event_user") + ` (first_day)`,
		`CREATE TABLE IF NOT EXISTS ` + q("insights_event_user_day") + ` (
    game TEXT NOT NULL DEFAULT '',
    day DATE NOT NULL,
    user_id TEXT NOT NULL,
    events BIGINT NOT NULL DEFAULT 0,
    platform TEXT NOT NULL DEFAULT '',
    app_version TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (game, day, user_id)
)`,
		`CREATE INDEX IF NOT EXISTS ` + ix("insights_event_user_day_user_idx") + ` ON ` + q("insights_event_user_day") + ` (user_id, day)`,
		`CREATE INDEX IF NOT EXISTS ` + ix("insights_event_user_day_day_idx") + ` ON ` + q("insights_event_user_day") + ` (day)`,
	}
}

func (p *PG) Ensure(ctx context.Context) error {
	if p.kind == DriverNakama {
		return nil
	}
	if _, err := p.db.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS "`+p.schema+`"`); err != nil {
		return err
	}
	for _, stmt := range PGSchema(p.schema) {
		if _, err := p.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (p *PG) Ping(ctx context.Context) error {
	var one int
	return p.db.QueryRowContext(ctx, `SELECT 1 FROM `+p.t("insights_event")+` LIMIT 1`).Scan(&one)
}

func (p *PG) Close() error {
	if p.owned {
		return p.db.Close()
	}
	return nil
}

type pgRow struct {
	ID         string            `json:"id"`
	Game       string            `json:"game"`
	Name       string            `json:"name"`
	Value      *float64          `json:"value"`
	Tags       map[string]string `json:"tags"`
	Fields     map[string]any    `json:"fields"`
	UserID     string            `json:"user_id"`
	SessionID  string            `json:"session_id"`
	At         time.Time         `json:"at"`
	Day        string            `json:"day"`
	ServerAt   time.Time         `json:"server_at"`
	Platform   string            `json:"platform"`
	AppVersion string            `json:"app_version"`
	Source     string            `json:"source"`
	Flagged    bool              `json:"flagged"`
	FlagReason string            `json:"flag_reason"`
}

func (p *PG) Write(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	rows := make([]pgRow, len(events))
	for i, e := range events {
		tags, fields := e.Tags, e.Fields
		if tags == nil {
			tags = map[string]string{}
		}
		if fields == nil {
			fields = map[string]any{}
		}
		rows[i] = pgRow{e.ID, e.Game, e.Name, e.Value, tags, fields, e.UserID, e.SessionID, e.At.UTC(), e.Day, e.ServerAt.UTC(), e.Platform, e.AppVersion, e.Source, e.Flagged, e.FlagReason}
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	ev, u, ud := p.t("insights_event"), p.t("insights_event_user"), p.t("insights_event_user_day")
	q := `WITH ins AS (
	INSERT INTO ` + ev + ` (id, game, name, value, tags, fields, user_id, session_id, at, day, server_at, platform, app_version, source, flagged, flag_reason)
	SELECT id, game, name, value, coalesce(tags,'{}'), coalesce(fields,'{}'), user_id, session_id, at, day, server_at, platform, app_version, source, flagged, flag_reason
	FROM jsonb_to_recordset($1::jsonb) AS x(id uuid, game text, name text, value float8, tags jsonb, fields jsonb, user_id text, session_id text, at timestamptz, day date, server_at timestamptz, platform text, app_version text, source text, flagged bool, flag_reason text)
	ON CONFLICT (id) DO NOTHING
	RETURNING game, user_id, day, at, platform, app_version, flagged
), act AS (SELECT * FROM ins WHERE user_id <> '' AND NOT flagged),
ud AS (
	INSERT INTO ` + ud + ` (game, day, user_id, events, platform, app_version)
	SELECT game, day, user_id, count(*), max(platform), max(app_version) FROM act GROUP BY game, day, user_id ORDER BY game, day, user_id
	ON CONFLICT (game, day, user_id) DO UPDATE SET events = ` + ud + `.events + EXCLUDED.events,
		platform = CASE WHEN EXCLUDED.platform <> '' THEN EXCLUDED.platform ELSE ` + ud + `.platform END,
		app_version = CASE WHEN EXCLUDED.app_version <> '' THEN EXCLUDED.app_version ELSE ` + ud + `.app_version END
	RETURNING 1
)
INSERT INTO ` + u + ` (game, user_id, first_day, first_at, last_day, platform, app_version)
SELECT game, user_id, min(day), min(at), max(day), max(platform), max(app_version) FROM act GROUP BY game, user_id ORDER BY game, user_id
ON CONFLICT (game, user_id) DO UPDATE SET first_day = LEAST(` + u + `.first_day, EXCLUDED.first_day), first_at = LEAST(` + u + `.first_at, EXCLUDED.first_at),
	last_day = GREATEST(` + u + `.last_day, EXCLUDED.last_day),
	platform = CASE WHEN EXCLUDED.platform <> '' THEN EXCLUDED.platform ELSE ` + u + `.platform END,
	app_version = CASE WHEN EXCLUDED.app_version <> '' THEN EXCLUDED.app_version ELSE ` + u + `.app_version END`
	_, err = p.db.ExecContext(ctx, q, raw)
	return err
}

type pgq struct {
	where []string
	args  []any
}

func (b *pgq) arg(v any) string {
	b.args = append(b.args, v)
	return "$" + strconv.Itoa(len(b.args))
}

func (b *pgq) add(cond string) { b.where = append(b.where, cond) }

func (b *pgq) sql() string {
	if len(b.where) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(b.where, " AND ")
}

func tagCond(b *pgq, col string, tags map[string][]string) {
	for _, g := range sortedKeys(tags) {
		alts := []string{}
		for _, id := range tags[g] {
			raw, _ := json.Marshal(map[string]string{g: id})
			alts = append(alts, col+" @> "+b.arg(string(raw))+"::jsonb")
		}
		b.add("(" + strings.Join(alts, " OR ") + ")")
	}
}

func sortedKeys(m map[string][]string) []string {
	f := Filter{Tags: m}
	return f.TagGroups()
}

func (p *PG) eventFilter(b *pgq, f Filter, prefix string) {
	c := func(n string) string { return prefix + n }
	if f.From != "" {
		b.add(c("day") + " >= " + b.arg(f.From) + "::date")
	}
	if f.To != "" {
		b.add(c("day") + " <= " + b.arg(f.To) + "::date")
	}
	if f.Game != nil {
		b.add(c("game") + " = " + b.arg(*f.Game))
	}
	if len(f.Names) == 1 {
		b.add(c("name") + " = " + b.arg(f.Names[0]))
	} else if len(f.Names) > 1 {
		b.add(c("name") + " = ANY(" + b.arg(f.Names) + "::text[])")
	}
	for col, v := range map[string]string{"platform": f.Platform, "app_version": f.AppVersion, "user_id": f.UserID, "session_id": f.SessionID, "source": f.Source} {
		if v != "" {
			b.add(c(col) + " = " + b.arg(v))
		}
	}
	tagCond(b, c("tags"), f.Tags)
	if !f.IncludeFlagged {
		b.add("NOT " + c("flagged"))
	}
}

func pgGroupExpr(g string) string {
	switch {
	case strings.HasPrefix(g, "tag:"):
		return "coalesce(tags->>'" + strings.TrimPrefix(g, "tag:") + "', '')"
	case g == "appVersion":
		return "app_version"
	case g == "user":
		return "user_id"
	}
	return g
}

func (p *PG) Aggregate(ctx context.Context, q AggQuery) ([]AggRow, error) {
	if err := q.Normalize(); err != nil {
		return nil, err
	}
	b := &pgq{}
	p.eventFilter(b, q.Filter, "")
	for _, g := range q.GroupBy {
		if strings.HasPrefix(g, "tag:") {
			b.add("tags ? " + b.arg(strings.TrimPrefix(g, "tag:")))
		}
	}
	bucket := "''"
	switch q.Bucket {
	case BucketDay:
		bucket = "day::text"
	case BucketWeek:
		bucket = "date_trunc('week', day)::date::text"
	case BucketHour:
		bucket = "to_char(date_trunc('hour', at AT TIME ZONE " + b.arg(q.Filter.Location().String()) + "), 'YYYY-MM-DD\"T\"HH24:00')"
	}
	metric := map[string]string{MetricCount: "count(*)::float8", MetricSum: "coalesce(sum(value),0)::float8", MetricAvg: "coalesce(avg(value),0)::float8",
		MetricMin: "coalesce(min(value),0)::float8", MetricMax: "coalesce(max(value),0)::float8", MetricUsers: "count(DISTINCT user_id)::float8"}[q.Metric]
	cols := []string{bucket + " AS b"}
	groups := []string{"1"}
	for i, g := range q.GroupBy {
		cols = append(cols, pgGroupExpr(g)+" AS k"+strconv.Itoa(i))
		groups = append(groups, strconv.Itoa(i+2))
	}
	order := "1, v DESC"
	if q.Bucket == BucketNone {
		order = "v DESC"
	}
	sqlText := `SELECT ` + strings.Join(cols, ", ") + `, ` + metric + ` AS v, count(*)::int8, count(DISTINCT user_id)::int8 FROM ` + p.t("insights_event") + b.sql() +
		` GROUP BY ` + strings.Join(groups, ", ") + ` ORDER BY ` + order + ` LIMIT ` + b.arg(q.Limit)
	rows, err := p.db.QueryContext(ctx, sqlText, b.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AggRow{}
	for rows.Next() {
		var r AggRow
		keys := make([]string, len(q.GroupBy))
		dest := []any{&r.Bucket}
		for i := range keys {
			dest = append(dest, &keys[i])
		}
		dest = append(dest, &r.Value, &r.Count, &r.Users)
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		r.Keys = keys
		out = append(out, r)
	}
	return out, rows.Err()
}

func (p *PG) userFilter(b *pgq, f Filter, prefix string) {
	if f.Game != nil {
		b.add(prefix + "game = " + b.arg(*f.Game))
	}
	if f.Platform != "" {
		b.add(prefix + "platform = " + b.arg(f.Platform))
	}
	if f.AppVersion != "" {
		b.add(prefix + "app_version = " + b.arg(f.AppVersion))
	}
}

func (p *PG) Activity(ctx context.Context, f Filter) (Activity, error) {
	out := Activity{Days: []ActivityDay{}}
	if err := f.Validate(); err != nil || f.From == "" || f.To == "" {
		return out, ErrBadQuery
	}
	from, _ := time.Parse(time.DateOnly, f.From)
	to, _ := time.Parse(time.DateOnly, f.To)
	idx := map[string]int{}
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		idx[d.Format(time.DateOnly)] = len(out.Days)
		out.Days = append(out.Days, ActivityDay{Day: d.Format(time.DateOnly)})
	}
	b := &pgq{}
	p.userFilter(b, f, "")
	b.add("day >= " + b.arg(f.From) + "::date")
	b.add("day <= " + b.arg(f.To) + "::date")
	rows, err := p.db.QueryContext(ctx, `SELECT day::text, count(DISTINCT user_id)::int8, coalesce(sum(events),0)::int8 FROM `+p.t("insights_event_user_day")+b.sql()+` GROUP BY 1`, b.args...)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var d string
		var n, e int64
		if err := rows.Scan(&d, &n, &e); err != nil {
			rows.Close()
			return out, err
		}
		if i, ok := idx[d]; ok {
			out.Days[i].DAU, out.Days[i].Events = n, e
			out.Events += e
		}
	}
	rows.Close()
	nb := &pgq{}
	p.userFilter(nb, f, "")
	firsts := `SELECT user_id, min(first_day) AS first_day FROM ` + p.t("insights_event_user") + nb.sql() + ` GROUP BY user_id`
	fromA, toA := nb.arg(f.From), nb.arg(f.To)
	rows, err = p.db.QueryContext(ctx, `SELECT first_day::text, count(*)::int8 FROM (`+firsts+`) u WHERE first_day >= `+fromA+`::date AND first_day <= `+toA+`::date GROUP BY 1`, nb.args...)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var d string
		var n int64
		if err := rows.Scan(&d, &n); err != nil {
			rows.Close()
			return out, err
		}
		if i, ok := idx[d]; ok {
			out.Days[i].NewUsers = n
			out.NewUsers += n
		}
	}
	rows.Close()
	for i := range out.Days {
		out.Days[i].Returning = out.Days[i].DAU - out.Days[i].NewUsers
		if out.Days[i].Returning < 0 {
			out.Days[i].Returning = 0
		}
	}
	count := func(fromDay string) (int64, error) {
		wb := &pgq{}
		p.userFilter(wb, f, "")
		wb.add("day >= " + wb.arg(fromDay) + "::date")
		wb.add("day <= " + wb.arg(f.To) + "::date")
		var n int64
		err := p.db.QueryRowContext(ctx, `SELECT count(DISTINCT user_id)::int8 FROM `+p.t("insights_event_user_day")+wb.sql(), wb.args...).Scan(&n)
		return n, err
	}
	var e error
	if out.DAU, e = count(f.To); e != nil {
		return out, e
	}
	if out.WAU, e = count(AddDays(f.To, -6)); e != nil {
		return out, e
	}
	if out.MAU, e = count(AddDays(f.To, -29)); e != nil {
		return out, e
	}
	out.Users, e = count(f.From)
	return out, e
}

func (p *PG) Retention(ctx context.Context, q RetentionQuery) (Retention, error) {
	out := Retention{Summary: []RetentionSummary{}, Cohorts: []Cohort{}}
	if err := ValidateRetention(&q); err != nil {
		return out, err
	}
	if q.From == "" || q.To == "" {
		return out, ErrBadQuery
	}
	cmp := "="
	if q.Mode == ModeRolling {
		cmp = ">="
	}
	b := &pgq{}
	p.userFilter(b, q.Filter, "")
	firsts := `SELECT user_id, min(first_day) AS first_day FROM ` + p.t("insights_event_user") + b.sql() + ` GROUP BY user_id`
	dayGame := ""
	if q.Game != nil {
		dayGame = " AND d.game = " + b.arg(*q.Game)
	}
	today := b.arg(q.Today) + "::date"
	cols := []string{}
	for _, n := range q.Days {
		a := b.arg(n) + "::int"
		cols = append(cols, `CASE WHEN u.first_day + `+a+` <= `+today+` THEN count(*) FILTER (WHERE EXISTS (SELECT 1 FROM `+p.t("insights_event_user_day")+` d WHERE d.user_id = u.user_id AND d.day `+cmp+` u.first_day + `+a+dayGame+`)) END`)
	}
	where := ` WHERE u.first_day >= ` + b.arg(q.From) + `::date AND u.first_day <= ` + b.arg(q.To) + `::date`
	rows, err := p.db.QueryContext(ctx, `SELECT u.first_day::text, count(*)::int8, `+strings.Join(cols, ", ")+` FROM (`+firsts+`) u`+where+` GROUP BY u.first_day ORDER BY u.first_day DESC`, b.args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	sums := make([]RetentionSummary, len(q.Days))
	for i, n := range q.Days {
		sums[i].Day = n
	}
	for rows.Next() {
		var c Cohort
		vals := make([]sql.NullInt64, len(q.Days))
		dest := []any{&c.Day, &c.Size}
		for i := range vals {
			dest = append(dest, &vals[i])
		}
		if err := rows.Scan(dest...); err != nil {
			return out, err
		}
		for i, n := range q.Days {
			cell := CohortCell{Day: n}
			if vals[i].Valid {
				v := vals[i].Int64
				cell.Returned, cell.Rate = &v, Rate(v, c.Size)
				sums[i].Cohort += c.Size
				sums[i].Returned += v
			}
			c.Cells = append(c.Cells, cell)
		}
		if len(out.Cohorts) < q.Rows {
			out.Cohorts = append(out.Cohorts, c)
		}
	}
	for i := range sums {
		sums[i].Rate = Rate(sums[i].Returned, sums[i].Cohort)
	}
	out.Summary = sums
	return out, rows.Err()
}

func (p *PG) stepCond(b *pgq, s Step, prefix string) string {
	sb := &pgq{args: b.args}
	sb.add(prefix + "name = " + sb.arg(s.Name))
	tagCond(sb, prefix+"tags", s.Tags)
	b.args = sb.args
	return strings.Join(sb.where, " AND ")
}

func (p *PG) Funnel(ctx context.Context, q FunnelQuery) ([]FunnelStep, error) {
	if err := ValidateFunnel(&q); err != nil {
		return nil, err
	}
	b := &pgq{}
	base := q.Filter
	base.Names, base.UserID, base.SessionID = nil, "", ""
	first := &pgq{}
	p.eventFilter(first, base, "")
	first.add("user_id <> ''")
	b.args = first.args
	firstWhere := strings.Join(append(first.where, p.stepCond(b, q.Steps[0], "")), " AND ")
	later := base
	later.From, later.To = "", ""
	lb := &pgq{args: b.args}
	p.eventFilter(lb, later, "e.")
	if base.From != "" {
		lb.add("e.day >= " + lb.arg(base.From) + "::date")
	}
	b.args = lb.args
	win := b.arg(q.Window.Seconds()) + "::float8"
	ctes := []string{`s0 AS (SELECT user_id, min(at) AS t0, min(at) AS t, 0::float8 AS dt FROM ` + p.t("insights_event") + ` WHERE ` + firstWhere + ` GROUP BY user_id)`}
	for i := 1; i < len(q.Steps); i++ {
		cond := p.stepCond(b, q.Steps[i], "e.")
		conds := append(append([]string{}, lb.where...), cond)
		prev := "s" + strconv.Itoa(i-1)
		ctes = append(ctes, `s`+strconv.Itoa(i)+` AS (SELECT s.user_id, s.t0, min(e.at) AS t, EXTRACT(EPOCH FROM (min(e.at) - s.t))::float8 AS dt FROM `+prev+` s JOIN `+p.t("insights_event")+` e ON e.user_id = s.user_id AND e.at > s.t AND e.at <= s.t0 + `+win+` * INTERVAL '1 second' WHERE `+strings.Join(conds, " AND ")+` GROUP BY s.user_id, s.t0, s.t)`)
	}
	sel := []string{}
	for i := range q.Steps {
		med := "NULL::float8"
		if i > 0 {
			med = "percentile_cont(0.5) WITHIN GROUP (ORDER BY dt)::float8"
		}
		sel = append(sel, `SELECT `+strconv.Itoa(i)+`, count(*)::int8, `+med+` FROM s`+strconv.Itoa(i))
	}
	rows, err := p.db.QueryContext(ctx, `WITH `+strings.Join(ctes, ", ")+` `+strings.Join(sel, " UNION ALL ")+` ORDER BY 1`, b.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]FunnelStep, len(q.Steps))
	for i, s := range q.Steps {
		out[i] = FunnelStep{Index: i, Name: s.Name}
	}
	for rows.Next() {
		var i int
		var n int64
		var m sql.NullFloat64
		if err := rows.Scan(&i, &n, &m); err != nil {
			return nil, err
		}
		out[i].Users = n
		if m.Valid {
			v := float64(int64(m.Float64*10+0.5)) / 10
			out[i].MedianSeconds = &v
		}
	}
	return out, rows.Err()
}

func (p *PG) Events(ctx context.Context, q EventQuery) ([]Event, error) {
	if err := q.Filter.Validate(); err != nil {
		return nil, err
	}
	if q.Limit <= 0 || q.Limit > 1000 {
		q.Limit = 100
	}
	b := &pgq{}
	p.eventFilter(b, q.Filter, "")
	dir := "DESC"
	if q.Asc {
		dir = "ASC"
	}
	if q.BeforeAt != nil {
		op := "<"
		if q.Asc {
			op = ">"
		}
		b.add("(at, id) " + op + " (" + b.arg(*q.BeforeAt) + "::timestamptz, " + b.arg(q.BeforeID) + "::uuid)")
	}
	rows, err := p.db.QueryContext(ctx, `SELECT id::text, game, name, value, tags, fields, user_id, session_id, at, day::text, server_at, platform, app_version, source, flagged, flag_reason FROM `+p.t("insights_event")+b.sql()+
		` ORDER BY at `+dir+`, id `+dir+` LIMIT `+b.arg(q.Limit), b.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Event{}
	for rows.Next() {
		var e Event
		var v sql.NullFloat64
		var tags, fields []byte
		if err := rows.Scan(&e.ID, &e.Game, &e.Name, &v, &tags, &fields, &e.UserID, &e.SessionID, &e.At, &e.Day, &e.ServerAt, &e.Platform, &e.AppVersion, &e.Source, &e.Flagged, &e.FlagReason); err != nil {
			return nil, err
		}
		if v.Valid {
			x := v.Float64
			e.Value = &x
		}
		_ = json.Unmarshal(tags, &e.Tags)
		_ = json.Unmarshal(fields, &e.Fields)
		if len(e.Fields) == 0 {
			e.Fields = nil
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (p *PG) Values(ctx context.Context, q ValuesQuery) ([]ValueCount, error) {
	if err := q.Filter.Validate(); err != nil {
		return nil, err
	}
	if q.Limit <= 0 || q.Limit > 500 {
		q.Limit = 100
	}
	b := &pgq{}
	p.eventFilter(b, q.Filter, "")
	expr := ""
	from := p.t("insights_event")
	switch {
	case q.Kind == "name":
		expr = "name"
	case q.Kind == "platform":
		expr = "platform"
	case q.Kind == "appVersion":
		expr = "app_version"
	case q.Kind == "source":
		expr = "source"
	case q.Kind == "game":
		expr = "game"
	case q.Kind == "tags":
		expr = "k"
		from = p.t("insights_event") + ", jsonb_object_keys(tags) AS k"
	case strings.HasPrefix(q.Kind, "tag:") && tagGroupRe.MatchString(strings.TrimPrefix(q.Kind, "tag:")):
		g := strings.TrimPrefix(q.Kind, "tag:")
		expr = "tags->>'" + g + "'"
		b.add("tags ? " + b.arg(g))
	default:
		return nil, ErrBadQuery
	}
	rows, err := p.db.QueryContext(ctx, `SELECT `+expr+`, count(*)::int8 FROM `+from+b.sql()+` GROUP BY 1 ORDER BY 2 DESC, 1 LIMIT `+b.arg(q.Limit), b.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ValueCount{}
	for rows.Next() {
		var v ValueCount
		if err := rows.Scan(&v.Value, &v.Count); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (p *PG) Sweep(ctx context.Context, beforeDay string) (int64, error) {
	if _, err := time.Parse(time.DateOnly, beforeDay); err != nil {
		return 0, ErrBadQuery
	}
	var total int64
	for i := 0; i < 200; i++ {
		res, err := p.db.ExecContext(ctx, `DELETE FROM `+p.t("insights_event")+` WHERE id IN (SELECT id FROM `+p.t("insights_event")+` WHERE day < $1::date LIMIT 5000)`, beforeDay)
		if err != nil {
			return total, err
		}
		n, _ := res.RowsAffected()
		total += n
		if n < 5000 {
			break
		}
	}
	if _, err := p.db.ExecContext(ctx, `DELETE FROM `+p.t("insights_event_user_day")+` WHERE day < $1::date`, beforeDay); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return total, err
	}
	return total, nil
}
