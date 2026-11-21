package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var chIdent = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,62}$`)

type ClickHouse struct {
	base     string
	database string
	table    string
	user     string
	password string
	client   *http.Client
	final    bool
}

func NewClickHouse(base, database, table, user, password string, client *http.Client) (*ClickHouse, error) {
	if database == "" {
		database = "default"
	}
	if table == "" {
		table = "insights_event"
	}
	if !chIdent.MatchString(database) || !chIdent.MatchString(table) {
		return nil, ErrBadQuery
	}
	u, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, errors.New("clickhouse url must be http(s)://host:port")
	}
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	return &ClickHouse{base: strings.TrimRight(base, "/"), database: database, table: table, user: user, password: password, client: client, final: true}, nil
}

func (c *ClickHouse) Kind() string { return DriverClickHouse }

func (c *ClickHouse) Close() error { return nil }

func (c *ClickHouse) t() string {
	from := "`" + c.database + "`.`" + c.table + "`"
	if c.final {
		from += " FINAL"
	}
	return from
}

type chq struct {
	where  []string
	params url.Values
	n      int
}

func newCHQ() *chq { return &chq{params: url.Values{}} }

func chEscape(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `'`, `\'`, "\t", `\t`, "\n", `\n`)
	return r.Replace(s)
}

func (b *chq) arg(v any, typ string) string {
	b.n++
	name := "p" + strconv.Itoa(b.n)
	switch x := v.(type) {
	case []string:
		parts := make([]string, len(x))
		for i, s := range x {
			parts[i] = "'" + chEscape(s) + "'"
		}
		b.params.Set("param_"+name, "["+strings.Join(parts, ",")+"]")
	case time.Time:
		b.params.Set("param_"+name, x.UTC().Format("2006-01-02 15:04:05.000"))
	default:
		b.params.Set("param_"+name, fmt.Sprint(v))
	}
	return "{" + name + ":" + typ + "}"
}

func (b *chq) add(cond string) { b.where = append(b.where, cond) }

func (b *chq) sql() string {
	if len(b.where) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(b.where, " AND ")
}

func (c *ClickHouse) do(ctx context.Context, query string, params url.Values, body io.Reader) ([]byte, error) {
	q := url.Values{}
	for k, v := range params {
		q[k] = v
	}
	q.Set("database", c.database)
	q.Set("output_format_json_quote_64bit_integers", "0")
	method := http.MethodPost
	var req *http.Request
	var err error
	if body == nil {
		req, err = http.NewRequestWithContext(ctx, method, c.base+"/?"+q.Encode(), strings.NewReader(query))
	} else {
		q.Set("query", query)
		req, err = http.NewRequestWithContext(ctx, method, c.base+"/?"+q.Encode(), body)
	}
	if err != nil {
		return nil, err
	}
	if c.user != "" {
		req.Header.Set("X-ClickHouse-User", c.user)
	}
	if c.password != "" {
		req.Header.Set("X-ClickHouse-Key", c.password)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 256<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(raw))
		if len(msg) > 400 {
			msg = msg[:400]
		}
		return nil, fmt.Errorf("clickhouse HTTP %d: %s", resp.StatusCode, msg)
	}
	return raw, nil
}

func (c *ClickHouse) rows(ctx context.Context, query string, params url.Values) ([][]any, error) {
	raw, err := c.do(ctx, query+" FORMAT JSONCompact", params, nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Data [][]any `json:"data"`
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return nil, errors.New("clickhouse answer is not JSON")
	}
	return out.Data, nil
}

func chStr(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case nil:
		return ""
	}
	return fmt.Sprint(v)
}

func chF(v any) float64 {
	switch x := v.(type) {
	case json.Number:
		f, _ := x.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	case float64:
		return x
	}
	return 0
}

func chI(v any) int64 { return int64(chF(v)) }

func (c *ClickHouse) Ensure(ctx context.Context) error {
	ddl := "CREATE TABLE IF NOT EXISTS `" + c.database + "`.`" + c.table + "` (" +
		"id UUID, game LowCardinality(String), name LowCardinality(String), value Nullable(Float64), tags Map(String, String), fields String, " +
		"user_id String, session_id String, at DateTime64(3, 'UTC'), day Date, server_at DateTime64(3, 'UTC'), platform LowCardinality(String), " +
		"app_version LowCardinality(String), source LowCardinality(String), flagged UInt8, flag_reason String) " +
		"ENGINE = ReplacingMergeTree PARTITION BY toYYYYMM(day) ORDER BY (game, name, day, id)"
	_, err := c.do(ctx, ddl, nil, nil)
	return err
}

func (c *ClickHouse) Ping(ctx context.Context) error {
	_, err := c.do(ctx, "SELECT 1", nil, nil)
	return err
}

func (c *ClickHouse) Write(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	var buf bytes.Buffer
	ids := make([]string, 0, len(events))
	for _, e := range events {
		fields := "{}"
		if len(e.Fields) > 0 {
			raw, _ := json.Marshal(e.Fields)
			fields = string(raw)
		}
		tags := e.Tags
		if tags == nil {
			tags = map[string]string{}
		}
		flag := 0
		if e.Flagged {
			flag = 1
		}
		row := map[string]any{"id": e.ID, "game": e.Game, "name": e.Name, "value": e.Value, "tags": tags, "fields": fields, "user_id": e.UserID, "session_id": e.SessionID,
			"at": e.At.UTC().Format("2006-01-02 15:04:05.000"), "day": e.Day, "server_at": e.ServerAt.UTC().Format("2006-01-02 15:04:05.000"), "platform": e.Platform,
			"app_version": e.AppVersion, "source": e.Source, "flagged": flag, "flag_reason": e.FlagReason}
		raw, err := json.Marshal(row)
		if err != nil {
			return err
		}
		buf.Write(raw)
		buf.WriteByte('\n')
		ids = append(ids, e.ID)
	}
	sort.Strings(ids)
	sum := sha256.Sum256([]byte(strings.Join(ids, ",")))
	params := url.Values{"insert_deduplication_token": {hex.EncodeToString(sum[:16])}, "input_format_skip_unknown_fields": {"1"}}
	_, err := c.do(ctx, "INSERT INTO `"+c.database+"`.`"+c.table+"` FORMAT JSONEachRow", params, &buf)
	return err
}

func (c *ClickHouse) filter(b *chq, f Filter) {
	if f.From != "" {
		b.add("day >= " + b.arg(f.From, "Date"))
	}
	if f.To != "" {
		b.add("day <= " + b.arg(f.To, "Date"))
	}
	if f.Game != nil {
		b.add("game = " + b.arg(*f.Game, "String"))
	}
	if len(f.Names) > 0 {
		b.add("name IN " + b.arg(f.Names, "Array(String)"))
	}
	for _, kv := range [][2]string{{"platform", f.Platform}, {"app_version", f.AppVersion}, {"user_id", f.UserID}, {"session_id", f.SessionID}, {"source", f.Source}} {
		if kv[1] != "" {
			b.add(kv[0] + " = " + b.arg(kv[1], "String"))
		}
	}
	chTags(b, f.Tags)
	if !f.IncludeFlagged {
		b.add("flagged = 0")
	}
}

func chTags(b *chq, tags map[string][]string) {
	for _, g := range sortedKeys(tags) {
		b.add("tags[" + b.arg(g, "String") + "] IN " + b.arg(tags[g], "Array(String)"))
	}
}

func chGroupExpr(b *chq, g string) string {
	switch {
	case strings.HasPrefix(g, "tag:"):
		return "tags[" + b.arg(strings.TrimPrefix(g, "tag:"), "String") + "]"
	case g == "appVersion":
		return "app_version"
	case g == "user":
		return "user_id"
	}
	return g
}

func (c *ClickHouse) Aggregate(ctx context.Context, q AggQuery) ([]AggRow, error) {
	if err := q.Normalize(); err != nil {
		return nil, err
	}
	b := newCHQ()
	c.filter(b, q.Filter)
	for _, g := range q.GroupBy {
		if strings.HasPrefix(g, "tag:") {
			b.add("mapContains(tags, " + b.arg(strings.TrimPrefix(g, "tag:"), "String") + ")")
		}
	}
	bucket := "''"
	switch q.Bucket {
	case BucketDay:
		bucket = "toString(day)"
	case BucketWeek:
		bucket = "toString(toMonday(day))"
	case BucketHour:
		bucket = "formatDateTime(toStartOfHour(toTimeZone(at, " + b.arg(q.Filter.Location().String(), "String") + ")), '%Y-%m-%dT%H:00')"
	}
	metric := map[string]string{MetricCount: "toFloat64(count())", MetricSum: "toFloat64(sum(ifNull(value, 0)))", MetricAvg: "toFloat64(ifNull(avg(value), 0))",
		MetricMin: "toFloat64(ifNull(min(value), 0))", MetricMax: "toFloat64(ifNull(max(value), 0))", MetricUsers: "toFloat64(uniqExact(user_id))"}[q.Metric]
	cols := []string{bucket + " AS b"}
	groups := []string{"b"}
	for i, g := range q.GroupBy {
		k := "k" + strconv.Itoa(i)
		cols = append(cols, chGroupExpr(b, g)+" AS "+k)
		groups = append(groups, k)
	}
	order := "b, v DESC"
	if q.Bucket == BucketNone {
		order = "v DESC"
	}
	sqlText := "SELECT " + strings.Join(cols, ", ") + ", " + metric + " AS v, count(), uniqExact(user_id) FROM " + c.t() + b.sql() + " GROUP BY " + strings.Join(groups, ", ") +
		" ORDER BY " + order + " LIMIT " + strconv.Itoa(q.Limit)
	data, err := c.rows(ctx, sqlText, b.params)
	if err != nil {
		return nil, err
	}
	out := []AggRow{}
	for _, r := range data {
		if len(r) < 4+len(q.GroupBy) {
			continue
		}
		row := AggRow{Bucket: chStr(r[0]), Keys: make([]string, len(q.GroupBy))}
		for i := range q.GroupBy {
			row.Keys[i] = chStr(r[1+i])
		}
		n := 1 + len(q.GroupBy)
		row.Value, row.Count, row.Users = chF(r[n]), chI(r[n+1]), chI(r[n+2])
		out = append(out, row)
	}
	return out, nil
}

func (c *ClickHouse) userBase(b *chq, f Filter) {
	if f.Game != nil {
		b.add("game = " + b.arg(*f.Game, "String"))
	}
	if f.Platform != "" {
		b.add("platform = " + b.arg(f.Platform, "String"))
	}
	if f.AppVersion != "" {
		b.add("app_version = " + b.arg(f.AppVersion, "String"))
	}
	b.add("user_id != ''")
	b.add("flagged = 0")
}

func (c *ClickHouse) Activity(ctx context.Context, f Filter) (Activity, error) {
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
	b := newCHQ()
	c.userBase(b, f)
	b.add("day >= " + b.arg(f.From, "Date"))
	b.add("day <= " + b.arg(f.To, "Date"))
	data, err := c.rows(ctx, "SELECT toString(day), uniqExact(user_id), count() FROM "+c.t()+b.sql()+" GROUP BY day", b.params)
	if err != nil {
		return out, err
	}
	for _, r := range data {
		if i, ok := idx[chStr(r[0])]; ok {
			out.Days[i].DAU, out.Days[i].Events = chI(r[1]), chI(r[2])
			out.Events += chI(r[2])
		}
	}
	nb := newCHQ()
	c.userBase(nb, f)
	fa, ta := nb.arg(f.From, "Date"), nb.arg(f.To, "Date")
	data, err = c.rows(ctx, "SELECT toString(first), count() FROM (SELECT user_id, min(day) AS first FROM "+c.t()+nb.sql()+" GROUP BY user_id) WHERE first >= "+fa+" AND first <= "+ta+" GROUP BY first", nb.params)
	if err != nil {
		return out, err
	}
	for _, r := range data {
		if i, ok := idx[chStr(r[0])]; ok {
			out.Days[i].NewUsers = chI(r[1])
			out.NewUsers += chI(r[1])
		}
	}
	for i := range out.Days {
		if d := out.Days[i].DAU - out.Days[i].NewUsers; d > 0 {
			out.Days[i].Returning = d
		}
	}
	wb := newCHQ()
	c.userBase(wb, f)
	toA := wb.arg(f.To, "Date")
	w := func(back int) string {
		return "uniqExactIf(user_id, day >= " + wb.arg(AddDays(f.To, -back), "Date") + " AND day <= " + toA + ")"
	}
	data, err = c.rows(ctx, "SELECT "+w(0)+", "+w(6)+", "+w(29)+", uniqExactIf(user_id, day >= "+wb.arg(f.From, "Date")+" AND day <= "+toA+") FROM "+c.t()+wb.sql(), wb.params)
	if err != nil {
		return out, err
	}
	if len(data) == 1 && len(data[0]) == 4 {
		out.DAU, out.WAU, out.MAU, out.Users = chI(data[0][0]), chI(data[0][1]), chI(data[0][2]), chI(data[0][3])
	}
	return out, nil
}

func (c *ClickHouse) Retention(ctx context.Context, q RetentionQuery) (Retention, error) {
	out := Retention{Summary: []RetentionSummary{}, Cohorts: []Cohort{}}
	if err := ValidateRetention(&q); err != nil || q.From == "" || q.To == "" {
		return out, ErrBadQuery
	}
	b := newCHQ()
	c.userBase(b, q.Filter)
	base := b.sql()
	fa, ta := b.arg(q.From, "Date"), b.arg(q.To, "Date")
	cmp := "="
	if q.Mode == ModeRolling {
		cmp = ">="
	}
	cols := []string{}
	for _, n := range q.Days {
		cols = append(cols, "uniqExactIf(u.user_id, a.day "+cmp+" addDays(u.first, "+strconv.Itoa(n)+"))")
	}
	sqlText := "WITH u AS (SELECT user_id, min(day) AS first FROM " + c.t() + base + " GROUP BY user_id HAVING first >= " + fa + " AND first <= " + ta + "), " +
		"a AS (SELECT DISTINCT user_id, day FROM " + c.t() + base + " AND user_id IN (SELECT user_id FROM u)) " +
		"SELECT toString(u.first), uniqExact(u.user_id), " + strings.Join(cols, ", ") + " FROM u LEFT JOIN a ON a.user_id = u.user_id GROUP BY u.first ORDER BY u.first DESC"
	data, err := c.rows(ctx, sqlText, b.params)
	if err != nil {
		return out, err
	}
	return retentionFrom(q, func(yield func(day string, size int64, vals []int64)) {
		for _, r := range data {
			vals := make([]int64, len(q.Days))
			for i := range q.Days {
				if 2+i < len(r) {
					vals[i] = chI(r[2+i])
				}
			}
			yield(chStr(r[0]), chI(r[1]), vals)
		}
	}), nil
}

func retentionFrom(q RetentionQuery, each func(func(day string, size int64, vals []int64))) Retention {
	out := Retention{Summary: make([]RetentionSummary, len(q.Days)), Cohorts: []Cohort{}}
	for i, n := range q.Days {
		out.Summary[i].Day = n
	}
	each(func(day string, size int64, vals []int64) {
		c := Cohort{Day: day, Size: size}
		for i, n := range q.Days {
			cell := CohortCell{Day: n}
			if AddDays(day, n) <= q.Today {
				v := vals[i]
				cell.Returned, cell.Rate = &v, Rate(v, size)
				out.Summary[i].Cohort += size
				out.Summary[i].Returned += v
			}
			c.Cells = append(c.Cells, cell)
		}
		if len(out.Cohorts) < q.Rows {
			out.Cohorts = append(out.Cohorts, c)
		}
	})
	for i := range out.Summary {
		out.Summary[i].Rate = Rate(out.Summary[i].Returned, out.Summary[i].Cohort)
	}
	return out
}

func (c *ClickHouse) stepCond(b *chq, s Step) string {
	parts := []string{"name = " + b.arg(s.Name, "String")}
	for _, g := range sortedKeys(s.Tags) {
		parts = append(parts, "tags["+b.arg(g, "String")+"] IN "+b.arg(s.Tags[g], "Array(String)"))
	}
	return "(" + strings.Join(parts, " AND ") + ")"
}

func (c *ClickHouse) Funnel(ctx context.Context, q FunnelQuery) ([]FunnelStep, error) {
	if err := ValidateFunnel(&q); err != nil {
		return nil, err
	}
	b := newCHQ()
	f := q.Filter
	f.Names, f.UserID, f.SessionID = nil, "", ""
	to := f.To
	f.To = ""
	c.filter(b, f)
	if to != "" {
		days := int(q.Window.Hours()/24) + 1
		b.add("day <= " + b.arg(AddDays(to, days), "Date"))
	}
	b.add("user_id != ''")
	conds := []string{}
	for i, s := range q.Steps {
		cond := c.stepCond(b, s)
		if i == 0 && to != "" {
			cond = "(" + cond + " AND day <= " + b.arg(to, "Date") + ")"
		}
		conds = append(conds, cond)
	}
	b.add("(" + strings.Join(conds, " OR ") + ")")
	sqlText := "SELECT level, count() FROM (SELECT user_id, windowFunnel(" + strconv.Itoa(int(q.Window.Seconds())) + ")(toDateTime(at), " + strings.Join(conds, ", ") + ") AS level FROM " +
		c.t() + b.sql() + " GROUP BY user_id) GROUP BY level"
	data, err := c.rows(ctx, sqlText, b.params)
	if err != nil {
		return nil, err
	}
	levels := map[int64]int64{}
	for _, r := range data {
		levels[chI(r[0])] += chI(r[1])
	}
	out := make([]FunnelStep, len(q.Steps))
	for i, s := range q.Steps {
		out[i] = FunnelStep{Index: i, Name: s.Name}
		for lvl, n := range levels {
			if lvl >= int64(i+1) {
				out[i].Users += n
			}
		}
	}
	return out, nil
}

func (c *ClickHouse) Events(ctx context.Context, q EventQuery) ([]Event, error) {
	if err := q.Filter.Validate(); err != nil {
		return nil, err
	}
	if q.Limit <= 0 || q.Limit > 1000 {
		q.Limit = 100
	}
	b := newCHQ()
	c.filter(b, q.Filter)
	dir, op := "DESC", "<"
	if q.Asc {
		dir, op = "ASC", ">"
	}
	if q.BeforeAt != nil {
		b.add("(at, toString(id)) " + op + " (" + b.arg(*q.BeforeAt, "DateTime64(3, 'UTC')") + ", " + b.arg(q.BeforeID, "String") + ")")
	}
	data, err := c.rows(ctx, "SELECT toString(id), game, name, value, tags, fields, user_id, session_id, toString(at), toString(day), toString(server_at), platform, app_version, source, flagged, flag_reason FROM "+
		c.t()+b.sql()+" ORDER BY at "+dir+", toString(id) "+dir+" LIMIT "+strconv.Itoa(q.Limit), b.params)
	if err != nil {
		return nil, err
	}
	out := []Event{}
	for _, r := range data {
		if len(r) < 16 {
			continue
		}
		e := Event{ID: chStr(r[0]), Game: chStr(r[1]), Name: chStr(r[2]), UserID: chStr(r[6]), SessionID: chStr(r[7]), Day: chStr(r[9]), Platform: chStr(r[11]),
			AppVersion: chStr(r[12]), Source: chStr(r[13]), Flagged: chI(r[14]) == 1, FlagReason: chStr(r[15])}
		if r[3] != nil {
			v := chF(r[3])
			e.Value = &v
		}
		if m, ok := r[4].(map[string]any); ok {
			e.Tags = map[string]string{}
			for k, v := range m {
				e.Tags[k] = chStr(v)
			}
		}
		if s := chStr(r[5]); s != "" && s != "{}" {
			_ = json.Unmarshal([]byte(s), &e.Fields)
		}
		e.At, _ = time.Parse("2006-01-02 15:04:05.000", chStr(r[8]))
		e.ServerAt, _ = time.Parse("2006-01-02 15:04:05.000", chStr(r[10]))
		out = append(out, e)
	}
	return out, nil
}

func (c *ClickHouse) Values(ctx context.Context, q ValuesQuery) ([]ValueCount, error) {
	if err := q.Filter.Validate(); err != nil {
		return nil, err
	}
	if q.Limit <= 0 || q.Limit > 500 {
		q.Limit = 100
	}
	b := newCHQ()
	c.filter(b, q.Filter)
	expr := ""
	switch {
	case q.Kind == "name", q.Kind == "platform", q.Kind == "source", q.Kind == "game":
		expr = q.Kind
	case q.Kind == "appVersion":
		expr = "app_version"
	case q.Kind == "tags":
		expr = "arrayJoin(mapKeys(tags))"
	case strings.HasPrefix(q.Kind, "tag:") && tagGroupRe.MatchString(strings.TrimPrefix(q.Kind, "tag:")):
		g := b.arg(strings.TrimPrefix(q.Kind, "tag:"), "String")
		expr = "tags[" + g + "]"
		b.add("mapContains(tags, " + g + ")")
	default:
		return nil, ErrBadQuery
	}
	data, err := c.rows(ctx, "SELECT "+expr+" AS v, count() AS n FROM "+c.t()+b.sql()+" GROUP BY v ORDER BY n DESC, v LIMIT "+strconv.Itoa(q.Limit), b.params)
	if err != nil {
		return nil, err
	}
	out := []ValueCount{}
	for _, r := range data {
		if len(r) >= 2 {
			out = append(out, ValueCount{Value: chStr(r[0]), Count: chI(r[1])})
		}
	}
	return out, nil
}

func (c *ClickHouse) Sweep(ctx context.Context, beforeDay string) (int64, error) {
	if _, err := time.Parse(time.DateOnly, beforeDay); err != nil {
		return 0, ErrBadQuery
	}
	b := newCHQ()
	_, err := c.do(ctx, "ALTER TABLE `"+c.database+"`.`"+c.table+"` DELETE WHERE day < "+b.arg(beforeDay, "Date"), b.params, nil)
	return 0, err
}
