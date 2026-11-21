package store

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	mw "github.com/heroiclabs/nakama/v3/integrations/insights/internal/mongowire"
)

const mongoMaxRows = 500000

type Mongo struct {
	cl   *mw.Client
	db   string
	coll string
}

func NewMongo(cl *mw.Client, db, coll string) (*Mongo, error) {
	if coll == "" {
		coll = "insights_event"
	}
	if !chIdent.MatchString(db) || !chIdent.MatchString(coll) {
		return nil, ErrBadQuery
	}
	return &Mongo{cl: cl, db: db, coll: coll}, nil
}

func (m *Mongo) Kind() string { return DriverMongo }

func (m *Mongo) Close() error { return m.cl.Close() }

func (m *Mongo) Ping(ctx context.Context) error {
	_, err := m.cl.Command(ctx, m.db, mw.D{{Key: "ping", Val: int32(1)}})
	return err
}

func idx(name string, keys mw.D) mw.D {
	return mw.D{{Key: "key", Val: keys}, {Key: "name", Val: name}}
}

func (m *Mongo) Ensure(ctx context.Context) error {
	_, err := m.cl.Command(ctx, m.db, mw.D{{Key: "createIndexes", Val: m.coll}, {Key: "indexes", Val: mw.A{
		idx("game_name_day", mw.D{{Key: "game", Val: int32(1)}, {Key: "name", Val: int32(1)}, {Key: "day", Val: int32(1)}}),
		idx("user_at", mw.D{{Key: "user_id", Val: int32(1)}, {Key: "at", Val: int32(1)}}),
		idx("session_at", mw.D{{Key: "session_id", Val: int32(1)}, {Key: "at", Val: int32(1)}}),
		idx("day", mw.D{{Key: "day", Val: int32(1)}}),
	}}})
	return err
}

func (m *Mongo) Write(ctx context.Context, events []Event) error {
	docs := make([]mw.D, 0, len(events))
	for _, e := range events {
		tags := map[string]string{}
		for k, v := range e.Tags {
			tags[k] = v
		}
		var fields map[string]any
		if len(e.Fields) > 0 {
			fields = e.Fields
		} else {
			fields = map[string]any{}
		}
		docs = append(docs, mw.D{{Key: "_id", Val: e.ID}, {Key: "game", Val: e.Game}, {Key: "name", Val: e.Name}, {Key: "value", Val: e.Value}, {Key: "tags", Val: tags},
			{Key: "fields", Val: fields}, {Key: "user_id", Val: e.UserID}, {Key: "session_id", Val: e.SessionID}, {Key: "at", Val: e.At.UTC()}, {Key: "day", Val: e.Day},
			{Key: "server_at", Val: e.ServerAt.UTC()}, {Key: "platform", Val: e.Platform}, {Key: "app_version", Val: e.AppVersion}, {Key: "source", Val: e.Source},
			{Key: "flagged", Val: e.Flagged}, {Key: "flag_reason", Val: e.FlagReason}})
	}
	for len(docs) > 0 {
		n := len(docs)
		if n > 1000 {
			n = 1000
		}
		if _, err := m.cl.InsertMany(ctx, m.db, m.coll, docs[:n]); err != nil {
			return err
		}
		docs = docs[n:]
	}
	return nil
}

func strList(v []string) mw.A {
	a := make(mw.A, len(v))
	for i, s := range v {
		a[i] = s
	}
	return a
}

func mongoMatch(f Filter) mw.D {
	m := mw.D{}
	day := mw.D{}
	if f.From != "" {
		day = append(day, mw.E{Key: "$gte", Val: f.From})
	}
	if f.To != "" {
		day = append(day, mw.E{Key: "$lte", Val: f.To})
	}
	if len(day) > 0 {
		m = append(m, mw.E{Key: "day", Val: day})
	}
	if f.Game != nil {
		m = append(m, mw.E{Key: "game", Val: *f.Game})
	}
	if len(f.Names) > 0 {
		m = append(m, mw.E{Key: "name", Val: mw.D{{Key: "$in", Val: strList(f.Names)}}})
	}
	for _, kv := range [][2]string{{"platform", f.Platform}, {"app_version", f.AppVersion}, {"user_id", f.UserID}, {"session_id", f.SessionID}, {"source", f.Source}} {
		if kv[1] != "" {
			m = append(m, mw.E{Key: kv[0], Val: kv[1]})
		}
	}
	for _, g := range f.TagGroups() {
		m = append(m, mw.E{Key: "tags." + g, Val: mw.D{{Key: "$in", Val: strList(f.Tags[g])}}})
	}
	if !f.IncludeFlagged {
		m = append(m, mw.E{Key: "flagged", Val: false})
	}
	return m
}

func mongoGroupField(g string) string {
	switch {
	case strings.HasPrefix(g, "tag:"):
		return "$tags." + strings.TrimPrefix(g, "tag:")
	case g == "appVersion":
		return "$app_version"
	case g == "user":
		return "$user_id"
	}
	return "$" + g
}

func mstr(v any) string {
	s, _ := v.(string)
	return s
}

func mnum(v any) float64 {
	f, _ := mw.Num(v)
	return f
}

func (m *Mongo) agg(ctx context.Context, pipeline []mw.D) ([]mw.D, error) {
	return m.cl.Aggregate(ctx, m.db, m.coll, pipeline, mongoMaxRows)
}

func (m *Mongo) Aggregate(ctx context.Context, q AggQuery) ([]AggRow, error) {
	if err := q.Normalize(); err != nil {
		return nil, err
	}
	match := mongoMatch(q.Filter)
	for _, g := range q.GroupBy {
		if strings.HasPrefix(g, "tag:") {
			match = append(match, mw.E{Key: "tags." + strings.TrimPrefix(g, "tag:"), Val: mw.D{{Key: "$exists", Val: true}}})
		}
	}
	tz := q.Filter.Location().String()
	var bucket any = ""
	switch q.Bucket {
	case BucketDay:
		bucket = "$day"
	case BucketHour:
		bucket = mw.D{{Key: "$dateToString", Val: mw.D{{Key: "format", Val: "%Y-%m-%dT%H:00"}, {Key: "date", Val: "$at"}, {Key: "timezone", Val: tz}}}}
	case BucketWeek:
		bucket = mw.D{{Key: "$dateToString", Val: mw.D{{Key: "format", Val: "%Y-%m-%d"}, {Key: "date", Val: mw.D{{Key: "$dateTrunc", Val: mw.D{{Key: "date", Val: "$at"}, {Key: "unit", Val: "week"}, {Key: "startOfWeek", Val: "monday"}, {Key: "timezone", Val: tz}}}}}}}}
	}
	id := mw.D{{Key: "b", Val: bucket}}
	for i, g := range q.GroupBy {
		id = append(id, mw.E{Key: "k" + itoa(i), Val: mw.D{{Key: "$ifNull", Val: mw.A{mongoGroupField(g), ""}}}})
	}
	metric := map[string]mw.D{MetricCount: {{Key: "$sum", Val: int32(1)}}, MetricSum: {{Key: "$sum", Val: "$value"}}, MetricAvg: {{Key: "$avg", Val: "$value"}},
		MetricMin: {{Key: "$min", Val: "$value"}}, MetricMax: {{Key: "$max", Val: "$value"}}, MetricUsers: {{Key: "$sum", Val: int32(0)}}}[q.Metric]
	sortDoc := mw.D{{Key: "v", Val: int32(-1)}}
	if q.Bucket != BucketNone {
		sortDoc = mw.D{{Key: "_id.b", Val: int32(1)}, {Key: "v", Val: int32(-1)}}
	}
	group := mw.D{{Key: "_id", Val: id}, {Key: "v", Val: metric}, {Key: "c", Val: mw.D{{Key: "$sum", Val: int32(1)}}}, {Key: "u", Val: mw.D{{Key: "$addToSet", Val: "$user_id"}}}}
	project := mw.D{{Key: "v", Val: "$v"}, {Key: "c", Val: "$c"}, {Key: "u", Val: mw.D{{Key: "$size", Val: "$u"}}}}
	if q.Metric == MetricUsers {
		project[0] = mw.E{Key: "v", Val: mw.D{{Key: "$size", Val: "$u"}}}
	}
	rows, err := m.agg(ctx, []mw.D{{{Key: "$match", Val: match}}, {{Key: "$group", Val: group}}, {{Key: "$project", Val: project}}, {{Key: "$sort", Val: sortDoc}}, {{Key: "$limit", Val: int32(q.Limit)}}})
	if err != nil {
		return nil, err
	}
	out := []AggRow{}
	for _, r := range rows {
		id, _ := r.Get("_id").(mw.D)
		row := AggRow{Bucket: mstr(id.Get("b")), Keys: make([]string, len(q.GroupBy)), Value: mnum(r.Get("v")), Count: int64(mnum(r.Get("c"))), Users: int64(mnum(r.Get("u")))}
		for i := range q.GroupBy {
			row.Keys[i] = mstr(id.Get("k" + itoa(i)))
		}
		out = append(out, row)
	}
	return out, nil
}

func itoa(i int) string { return strconv.Itoa(i) }

func (m *Mongo) userMatch(f Filter) mw.D {
	u := Filter{Game: f.Game, Platform: f.Platform, AppVersion: f.AppVersion}
	match := mongoMatch(u)
	return append(match, mw.E{Key: "user_id", Val: mw.D{{Key: "$ne", Val: ""}}})
}

func (m *Mongo) firsts(ctx context.Context, f Filter) (map[string]string, error) {
	rows, err := m.agg(ctx, []mw.D{{{Key: "$match", Val: m.userMatch(f)}}, {{Key: "$group", Val: mw.D{{Key: "_id", Val: "$user_id"}, {Key: "f", Val: mw.D{{Key: "$min", Val: "$day"}}}}}},
		{{Key: "$match", Val: mw.D{{Key: "f", Val: mw.D{{Key: "$gte", Val: f.From}, {Key: "$lte", Val: f.To}}}}}}})
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, r := range rows {
		out[mstr(r.Get("_id"))] = mstr(r.Get("f"))
	}
	return out, nil
}

func (m *Mongo) userDays(ctx context.Context, f Filter, from, to string, users map[string]string) ([][2]string, map[string]int64, error) {
	match := m.userMatch(f)
	match = append(match, mw.E{Key: "day", Val: mw.D{{Key: "$gte", Val: from}, {Key: "$lte", Val: to}}})
	if users != nil {
		list := make([]string, 0, len(users))
		for u := range users {
			list = append(list, u)
		}
		match = append(match, mw.E{Key: "user_id", Val: mw.D{{Key: "$in", Val: strList(list)}}})
	}
	rows, err := m.agg(ctx, []mw.D{{{Key: "$match", Val: match}}, {{Key: "$group", Val: mw.D{{Key: "_id", Val: mw.D{{Key: "d", Val: "$day"}, {Key: "u", Val: "$user_id"}}}, {Key: "n", Val: mw.D{{Key: "$sum", Val: int32(1)}}}}}}})
	if err != nil {
		return nil, nil, err
	}
	pairs := make([][2]string, 0, len(rows))
	events := map[string]int64{}
	for _, r := range rows {
		id, _ := r.Get("_id").(mw.D)
		d, u := mstr(id.Get("d")), mstr(id.Get("u"))
		pairs = append(pairs, [2]string{d, u})
		events[d] += int64(mnum(r.Get("n")))
	}
	return pairs, events, nil
}

func ActivityFrom(f Filter, pairs [][2]string, events map[string]int64, firsts map[string]string) Activity {
	out := Activity{Days: []ActivityDay{}}
	from, _ := time.Parse(time.DateOnly, f.From)
	to, _ := time.Parse(time.DateOnly, f.To)
	idx := map[string]int{}
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		idx[d.Format(time.DateOnly)] = len(out.Days)
		out.Days = append(out.Days, ActivityDay{Day: d.Format(time.DateOnly)})
	}
	wau, mau, dau, all := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	w0, m0 := AddDays(f.To, -6), AddDays(f.To, -29)
	for _, p := range pairs {
		d, u := p[0], p[1]
		if i, ok := idx[d]; ok {
			out.Days[i].DAU++
			all[u] = true
		}
		if d == f.To {
			dau[u] = true
		}
		if d >= w0 && d <= f.To {
			wau[u] = true
		}
		if d >= m0 && d <= f.To {
			mau[u] = true
		}
	}
	for d, n := range events {
		if i, ok := idx[d]; ok {
			out.Days[i].Events = n
			out.Events += n
		}
	}
	for _, first := range firsts {
		if i, ok := idx[first]; ok {
			out.Days[i].NewUsers++
			out.NewUsers++
		}
	}
	for i := range out.Days {
		if r := out.Days[i].DAU - out.Days[i].NewUsers; r > 0 {
			out.Days[i].Returning = r
		}
	}
	out.DAU, out.WAU, out.MAU, out.Users = int64(len(dau)), int64(len(wau)), int64(len(mau)), int64(len(all))
	return out
}

func (m *Mongo) Activity(ctx context.Context, f Filter) (Activity, error) {
	if err := f.Validate(); err != nil || f.From == "" || f.To == "" {
		return Activity{}, ErrBadQuery
	}
	start := f.From
	if s := AddDays(f.To, -29); s < start {
		start = s
	}
	pairs, events, err := m.userDays(ctx, f, start, f.To, nil)
	if err != nil {
		return Activity{}, err
	}
	firsts, err := m.firsts(ctx, f)
	if err != nil {
		return Activity{}, err
	}
	return ActivityFrom(f, pairs, events, firsts), nil
}

func RetentionFromDays(q RetentionQuery, firsts map[string]string, active map[string]map[string]bool) Retention {
	type cohort struct {
		users []string
	}
	byDay := map[string]*cohort{}
	for u, d := range firsts {
		c := byDay[d]
		if c == nil {
			c = &cohort{}
			byDay[d] = c
		}
		c.users = append(c.users, u)
	}
	days := make([]string, 0, len(byDay))
	for d := range byDay {
		days = append(days, d)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(days)))
	return retentionFrom(q, func(yield func(string, int64, []int64)) {
		for _, d := range days {
			vals := make([]int64, len(q.Days))
			for i, n := range q.Days {
				target := AddDays(d, n)
				for _, u := range byDay[d].users {
					for ad := range active[u] {
						if ad == target || (q.Mode == ModeRolling && ad > target) {
							vals[i]++
							break
						}
					}
				}
			}
			yield(d, int64(len(byDay[d].users)), vals)
		}
	})
}

func (m *Mongo) Retention(ctx context.Context, q RetentionQuery) (Retention, error) {
	if err := ValidateRetention(&q); err != nil || q.From == "" || q.To == "" {
		return Retention{}, ErrBadQuery
	}
	firsts, err := m.firsts(ctx, q.Filter)
	if err != nil {
		return Retention{}, err
	}
	maxN := 0
	for _, n := range q.Days {
		if n > maxN {
			maxN = n
		}
	}
	end := AddDays(q.To, maxN)
	if q.Mode == ModeRolling || end > q.Today {
		end = q.Today
	}
	active := map[string]map[string]bool{}
	if len(firsts) > 0 {
		pairs, _, err := m.userDays(ctx, q.Filter, AddDays(q.From, 1), end, firsts)
		if err != nil {
			return Retention{}, err
		}
		for _, p := range pairs {
			if active[p[1]] == nil {
				active[p[1]] = map[string]bool{}
			}
			active[p[1]][p[0]] = true
		}
	}
	return RetentionFromDays(q, firsts, active), nil
}

type FunnelHit struct {
	User string
	At   time.Time
	Name string
	Tags map[string]string
}

func stepMatches(s Step, h FunnelHit) bool {
	if s.Name != h.Name {
		return false
	}
	for g, ids := range s.Tags {
		ok := false
		for _, id := range ids {
			if h.Tags[g] == id {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

func FunnelFromHits(q FunnelQuery, hits []FunnelHit) []FunnelStep {
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].User != hits[j].User {
			return hits[i].User < hits[j].User
		}
		return hits[i].At.Before(hits[j].At)
	})
	out := make([]FunnelStep, len(q.Steps))
	gaps := make([][]float64, len(q.Steps))
	for i, s := range q.Steps {
		out[i] = FunnelStep{Index: i, Name: s.Name}
	}
	for i := 0; i < len(hits); {
		j := i
		for j < len(hits) && hits[j].User == hits[i].User {
			j++
		}
		user := hits[i:j]
		var t0, prev time.Time
		step := 0
		for _, h := range user {
			if step == 0 {
				if stepMatches(q.Steps[0], h) {
					t0, prev, step = h.At, h.At, 1
					out[0].Users++
				}
				continue
			}
			if step >= len(q.Steps) {
				break
			}
			if h.At.Sub(t0) > q.Window {
				break
			}
			if h.At.After(prev) && stepMatches(q.Steps[step], h) {
				gaps[step] = append(gaps[step], h.At.Sub(prev).Seconds())
				out[step].Users++
				prev = h.At
				step++
			}
		}
		i = j
	}
	for i := 1; i < len(out); i++ {
		if m := Median(gaps[i]); m != nil {
			v := float64(int64(*m*10+0.5)) / 10
			out[i].MedianSeconds = &v
		}
	}
	return out
}

func (m *Mongo) Funnel(ctx context.Context, q FunnelQuery) ([]FunnelStep, error) {
	if err := ValidateFunnel(&q); err != nil {
		return nil, err
	}
	f := q.Filter
	f.UserID, f.SessionID = "", ""
	names := []string{}
	seen := map[string]bool{}
	for _, s := range q.Steps {
		if !seen[s.Name] {
			seen[s.Name] = true
			names = append(names, s.Name)
		}
	}
	f.Names = names
	if f.To != "" {
		f.To = AddDays(f.To, int(q.Window.Hours()/24)+1)
	}
	match := mongoMatch(f)
	match = append(match, mw.E{Key: "user_id", Val: mw.D{{Key: "$ne", Val: ""}}})
	rows, err := m.agg(ctx, []mw.D{{{Key: "$match", Val: match}}, {{Key: "$sort", Val: mw.D{{Key: "user_id", Val: int32(1)}, {Key: "at", Val: int32(1)}}}},
		{{Key: "$project", Val: mw.D{{Key: "user_id", Val: int32(1)}, {Key: "at", Val: int32(1)}, {Key: "name", Val: int32(1)}, {Key: "tags", Val: int32(1)}, {Key: "day", Val: int32(1)}}}}})
	if err != nil {
		return nil, err
	}
	hits := make([]FunnelHit, 0, len(rows))
	firstLimit := q.Filter.To
	for _, r := range rows {
		h := FunnelHit{User: mstr(r.Get("user_id")), Name: mstr(r.Get("name")), Tags: map[string]string{}}
		h.At, _ = r.Get("at").(time.Time)
		if t, ok := r.Get("tags").(mw.D); ok {
			for _, e := range t {
				h.Tags[e.Key] = mstr(e.Val)
			}
		}
		if firstLimit != "" && mstr(r.Get("day")) > firstLimit && stepMatches(q.Steps[0], h) {
			h.Name = "\x00late"
		}
		hits = append(hits, h)
	}
	return FunnelFromHits(q, hits), nil
}

func (m *Mongo) Events(ctx context.Context, q EventQuery) ([]Event, error) {
	if err := q.Filter.Validate(); err != nil {
		return nil, err
	}
	if q.Limit <= 0 || q.Limit > 1000 {
		q.Limit = 100
	}
	match := mongoMatch(q.Filter)
	dir := int32(-1)
	op := "$lt"
	if q.Asc {
		dir, op = 1, "$gt"
	}
	if q.BeforeAt != nil {
		match = append(match, mw.E{Key: "$or", Val: mw.A{
			mw.D{{Key: "at", Val: mw.D{{Key: op, Val: q.BeforeAt.UTC()}}}},
			mw.D{{Key: "at", Val: q.BeforeAt.UTC()}, {Key: "_id", Val: mw.D{{Key: op, Val: q.BeforeID}}}},
		}})
	}
	rows, err := m.agg(ctx, []mw.D{{{Key: "$match", Val: match}}, {{Key: "$sort", Val: mw.D{{Key: "at", Val: dir}, {Key: "_id", Val: dir}}}}, {{Key: "$limit", Val: int32(q.Limit)}}})
	if err != nil {
		return nil, err
	}
	out := []Event{}
	for _, r := range rows {
		e := Event{ID: mstr(r.Get("_id")), Game: mstr(r.Get("game")), Name: mstr(r.Get("name")), UserID: mstr(r.Get("user_id")), SessionID: mstr(r.Get("session_id")),
			Day: mstr(r.Get("day")), Platform: mstr(r.Get("platform")), AppVersion: mstr(r.Get("app_version")), Source: mstr(r.Get("source")), FlagReason: mstr(r.Get("flag_reason"))}
		e.Flagged, _ = r.Get("flagged").(bool)
		if v, ok := mw.Num(r.Get("value")); ok {
			e.Value = &v
		}
		e.At, _ = r.Get("at").(time.Time)
		e.ServerAt, _ = r.Get("server_at").(time.Time)
		if t, ok := r.Get("tags").(mw.D); ok && len(t) > 0 {
			e.Tags = map[string]string{}
			for _, kv := range t {
				e.Tags[kv.Key] = mstr(kv.Val)
			}
		}
		if fd, ok := r.Get("fields").(mw.D); ok && len(fd) > 0 {
			raw, _ := json.Marshal(fd.Map())
			_ = json.Unmarshal(raw, &e.Fields)
		}
		out = append(out, e)
	}
	return out, nil
}

func (m *Mongo) Values(ctx context.Context, q ValuesQuery) ([]ValueCount, error) {
	if err := q.Filter.Validate(); err != nil {
		return nil, err
	}
	if q.Limit <= 0 || q.Limit > 500 {
		q.Limit = 100
	}
	match := mongoMatch(q.Filter)
	pipeline := []mw.D{}
	var key any
	switch {
	case q.Kind == "name", q.Kind == "platform", q.Kind == "source", q.Kind == "game":
		key = "$" + q.Kind
	case q.Kind == "appVersion":
		key = "$app_version"
	case q.Kind == "tags":
		pipeline = append(pipeline, mw.D{{Key: "$match", Val: match}}, mw.D{{Key: "$project", Val: mw.D{{Key: "k", Val: mw.D{{Key: "$objectToArray", Val: "$tags"}}}}}}, mw.D{{Key: "$unwind", Val: "$k"}})
		key = "$k.k"
	case strings.HasPrefix(q.Kind, "tag:") && tagGroupRe.MatchString(strings.TrimPrefix(q.Kind, "tag:")):
		g := strings.TrimPrefix(q.Kind, "tag:")
		match = append(match, mw.E{Key: "tags." + g, Val: mw.D{{Key: "$exists", Val: true}}})
		key = "$tags." + g
	default:
		return nil, ErrBadQuery
	}
	if len(pipeline) == 0 {
		pipeline = append(pipeline, mw.D{{Key: "$match", Val: match}})
	}
	pipeline = append(pipeline, mw.D{{Key: "$group", Val: mw.D{{Key: "_id", Val: key}, {Key: "n", Val: mw.D{{Key: "$sum", Val: int32(1)}}}}}},
		mw.D{{Key: "$sort", Val: mw.D{{Key: "n", Val: int32(-1)}, {Key: "_id", Val: int32(1)}}}}, mw.D{{Key: "$limit", Val: int32(q.Limit)}})
	rows, err := m.agg(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	out := []ValueCount{}
	for _, r := range rows {
		out = append(out, ValueCount{Value: mstr(r.Get("_id")), Count: int64(mnum(r.Get("n")))})
	}
	return out, nil
}

func (m *Mongo) Sweep(ctx context.Context, beforeDay string) (int64, error) {
	if _, err := time.Parse(time.DateOnly, beforeDay); err != nil {
		return 0, ErrBadQuery
	}
	out, err := m.cl.Command(ctx, m.db, mw.D{{Key: "delete", Val: m.coll}, {Key: "deletes", Val: mw.A{mw.D{{Key: "q", Val: mw.D{{Key: "day", Val: mw.D{{Key: "$lt", Val: beforeDay}}}}}, {Key: "limit", Val: int32(0)}}}}})
	if err != nil {
		return 0, err
	}
	return int64(mnum(out.Get("n"))), nil
}
