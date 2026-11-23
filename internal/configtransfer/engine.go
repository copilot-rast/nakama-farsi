package configtransfer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

const StatusError = "error"

type ExportOptions struct {
	Areas  []string
	Source Source
	Now    time.Time
}

type Options struct {
	Areas     []string
	Rows      map[string][]string
	SwitchOff map[string]bool
	Commit    bool
	By        string
	Audit     func(ctx context.Context, q Querier, res *Result) error
}

type Counts struct {
	New          int `json:"new"`
	Changed      int `json:"changed"`
	Same         int `json:"same"`
	MissingRefs  int `json:"missingRefs"`
	SecretNeeded int `json:"secretNeeded"`
	Skipped      int `json:"skipped"`
	SwitchedOff  int `json:"switchedOff"`
	Errors       int `json:"errors"`
}

func (c *Counts) add(status string) {
	switch status {
	case StatusNew:
		c.New++
	case StatusChanged:
		c.Changed++
	case StatusSame:
		c.Same++
	case StatusMissingRefs:
		c.MissingRefs++
	case StatusSecretNeeded:
		c.SecretNeeded++
	case StatusSkipped:
		c.Skipped++
	case StatusSwitchedOff:
		c.SwitchedOff++
	case StatusError:
		c.Errors++
	}
}

func (c *Counts) plus(o Counts) {
	c.New += o.New
	c.Changed += o.Changed
	c.Same += o.Same
	c.MissingRefs += o.MissingRefs
	c.SecretNeeded += o.SecretNeeded
	c.Skipped += o.Skipped
	c.SwitchedOff += o.SwitchedOff
	c.Errors += o.Errors
}

type Change struct {
	Field string `json:"field"`
	From  any    `json:"from"`
	To    any    `json:"to"`
}

type Missing struct {
	Field string `json:"field"`
	Table string `json:"table"`
	Key   string `json:"key"`
}

type RowResult struct {
	Table   string    `json:"table"`
	Key     string    `json:"key"`
	Status  string    `json:"status"`
	Changes []Change  `json:"changes,omitempty"`
	Missing []Missing `json:"missing,omitempty"`
	Reason  string    `json:"reason,omitempty"`
}

type AreaResult struct {
	Key       string            `json:"key"`
	Owner     string            `json:"owner"`
	Title     map[string]string `json:"title"`
	Selected  bool              `json:"selected"`
	SwitchOff bool              `json:"switchOff"`
	CanSwitch bool              `json:"canSwitchOff"`
	InFile    int               `json:"inFile"`
	Counts    Counts            `json:"counts"`
	Rows      []RowResult       `json:"rows"`
	Warnings  []string          `json:"warnings,omitempty"`
}

type Result struct {
	Applied  bool         `json:"applied"`
	Source   Source       `json:"source"`
	Version  int          `json:"version"`
	Totals   Counts       `json:"totals"`
	Areas    []AreaResult `json:"areas"`
	Unknown  []string     `json:"unknownAreas"`
	Warnings []string     `json:"warnings,omitempty"`
}

func Export(ctx context.Context, db *sql.DB, o ExportOptions) (*File, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SET LOCAL TimeZone = 'UTC'`); err != nil {
		return nil, err
	}
	now := o.Now
	if now.IsZero() {
		now = time.Now()
	}
	src := o.Source
	src.ExportedAt = now.UTC().Format(time.RFC3339)
	f := &File{Format: Format, Version: FormatVersion, Source: src, Counts: Counted{}, Areas: []FileArea{}}
	want := toSet(o.Areas)
	s := newSession(tx)
	for _, a := range Areas() {
		if len(want) > 0 && !want[a.Key] {
			continue
		}
		fa := FileArea{Key: a.Key, Owner: a.Owner, Tables: []FileTable{}}
		if a.Custom != nil {
			tables, err := a.Custom.Export(ctx, tx)
			if err != nil {
				return nil, err
			}
			for _, t := range tables {
				t.Count = len(t.Rows)
				fa.Count += t.Count
				fa.Tables = append(fa.Tables, t)
			}
		} else {
			for _, t := range a.Tables {
				ti, err := s.info(ctx, t)
				if err != nil {
					return nil, err
				}
				if ti == nil {
					continue
				}
				rows, err := s.readTable(ctx, ti, true)
				if err != nil {
					return nil, err
				}
				ft := FileTable{Table: t.Name, Rows: make([]Row, 0, len(rows))}
				for _, r := range rows {
					ft.Rows = append(ft.Rows, Row{Key: r.key, Values: r.values})
				}
				ft.Count = len(ft.Rows)
				fa.Count += ft.Count
				fa.Tables = append(fa.Tables, ft)
			}
		}
		f.Counts[a.Key] = fa.Count
		f.Areas = append(f.Areas, fa)
	}
	return f, nil
}

func toSet(list []string) map[string]bool {
	out := map[string]bool{}
	for _, v := range list {
		if v = strings.TrimSpace(v); v != "" {
			out[v] = true
		}
	}
	return out
}

type job struct {
	area   Area
	ai     int
	table  Table
	ti     int
	ft     *FileTable
	custom []FileTable
}

func Apply(ctx context.Context, db *sql.DB, f *File, o Options) (*Result, error) {
	if f == nil || f.Format != Format {
		return nil, ErrFormat
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if _, err := tx.ExecContext(ctx, `SET LOCAL TimeZone = 'UTC'`); err != nil {
		return nil, err
	}
	res := &Result{Source: f.Source, Version: f.Version, Areas: []AreaResult{}, Unknown: []string{}}
	want := toSet(o.Areas)
	results := map[string]*AreaResult{}
	var jobs []job
	order := Areas()
	for ai, a := range order {
		fa := f.Area(a.Key)
		if fa == nil {
			continue
		}
		ar := &AreaResult{Key: a.Key, Owner: a.Owner, Title: a.Title, Selected: len(want) == 0 || want[a.Key], InFile: fa.Count, Rows: []RowResult{}}
		for _, t := range a.Tables {
			if t.Active != "" {
				ar.CanSwitch = true
			}
		}
		ar.SwitchOff = ar.Selected && ar.CanSwitch && o.SwitchOff[a.Key]
		results[a.Key] = ar
		if a.Custom != nil {
			jobs = append(jobs, job{area: a, ai: ai, custom: fa.Tables})
			continue
		}
		known := map[string]bool{}
		for ti, t := range a.Tables {
			known[t.Name] = true
			ft := fa.Table(t.Name)
			if ft == nil {
				continue
			}
			jobs = append(jobs, job{area: a, ai: ai, table: t, ti: ti, ft: ft})
		}
		for _, ft := range fa.Tables {
			if !known[ft.Table] {
				ar.Warnings = append(ar.Warnings, "table_unknown:"+ft.Table)
			}
		}
	}
	for _, fa := range f.Areas {
		if _, ok := results[fa.Key]; !ok {
			res.Unknown = append(res.Unknown, fa.Key)
		}
	}
	jobs = sortJobs(jobs)
	s := newSession(tx)
	for _, j := range jobs {
		ar := results[j.area.Key]
		picks := pickFunc(o.Rows, j.area.Key, ar.Selected)
		if j.area.Custom != nil {
			if !ar.Selected {
				for _, ft := range j.custom {
					for _, r := range ft.Rows {
						ar.Rows = append(ar.Rows, RowResult{Table: ft.Table, Key: r.Key, Status: StatusSkipped})
					}
				}
				continue
			}
			rows, err := j.area.Custom.Apply(ctx, tx, j.custom, picks, o.By)
			if err != nil {
				return nil, err
			}
			ar.Rows = append(ar.Rows, rows...)
			continue
		}
		ti, err := s.info(ctx, j.table)
		if err != nil {
			return nil, err
		}
		if ti == nil {
			ar.Warnings = append(ar.Warnings, "table_missing:"+j.table.Name)
			for _, r := range j.ft.Rows {
				ar.Rows = append(ar.Rows, RowResult{Table: j.table.Name, Key: r.Key, Status: StatusSkipped, Reason: "table_missing"})
			}
			continue
		}
		rows, warns, err := s.applyTable(ctx, tx, ti, j.ft, picks, ar.SwitchOff, o.By)
		if err != nil {
			return nil, err
		}
		ar.Rows = append(ar.Rows, rows...)
		ar.Warnings = append(ar.Warnings, warns...)
	}
	for _, a := range order {
		ar, ok := results[a.Key]
		if !ok {
			continue
		}
		for _, r := range ar.Rows {
			ar.Counts.add(r.Status)
		}
		res.Totals.plus(ar.Counts)
		res.Areas = append(res.Areas, *ar)
	}
	if !o.Commit || res.Totals.Errors > 0 {
		return res, nil
	}
	if o.Audit != nil {
		if err := o.Audit(ctx, tx, res); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	res.Applied = true
	for i := range res.Areas {
		a, ok := Lookup(res.Areas[i].Key)
		if !ok || a.After == nil || !res.Areas[i].Selected {
			continue
		}
		res.Areas[i].Warnings = append(res.Areas[i].Warnings, a.After(ctx, res.Areas[i])...)
	}
	return res, nil
}

func pickFunc(rows map[string][]string, area string, selected bool) func(table, key string) bool {
	if !selected {
		return func(string, string) bool { return false }
	}
	list, ok := rows[area]
	if !ok || list == nil {
		return func(string, string) bool { return true }
	}
	set := map[string]bool{}
	for _, r := range list {
		set[r] = true
	}
	return func(table, key string) bool { return set[table+":"+key] }
}

func sortJobs(jobs []job) []job {
	n := len(jobs)
	idx := map[string][]int{}
	for i, j := range jobs {
		if j.area.Custom == nil {
			idx[j.table.Name] = append(idx[j.table.Name], i)
		}
	}
	deps := make([]map[int]bool, n)
	for i, j := range jobs {
		deps[i] = map[int]bool{}
		for _, r := range j.table.Refs {
			if r.Table == j.table.Name {
				continue
			}
			for _, d := range idx[r.Table] {
				deps[i][d] = true
			}
		}
	}
	done := make([]bool, n)
	out := make([]job, 0, n)
	for len(out) < n {
		cands := make([]int, 0)
		for i := range jobs {
			if done[i] {
				continue
			}
			ready := true
			for d := range deps[i] {
				if !done[d] {
					ready = false
					break
				}
			}
			if ready {
				cands = append(cands, i)
			}
		}
		if len(cands) == 0 {
			for i := range jobs {
				if !done[i] {
					cands = append(cands, i)
					break
				}
			}
		}
		sort.SliceStable(cands, func(a, b int) bool {
			ja, jb := jobs[cands[a]], jobs[cands[b]]
			if ja.ai != jb.ai {
				return ja.ai < jb.ai
			}
			return ja.ti < jb.ti
		})
		i := cands[0]
		done[i] = true
		out = append(out, jobs[i])
	}
	return out
}

type fileRow struct {
	row   Row
	key   string
	ident string
	pick  bool
	done  bool
}

func (s *session) applyTable(ctx context.Context, tx Querier, ti *tableInfo, ft *FileTable, pick func(table, key string) bool, switchOff bool, by string) ([]RowResult, []string, error) {
	cur, err := s.readTable(ctx, ti, false)
	if err != nil {
		return nil, nil, err
	}
	curBy := make(map[string]*tableRow, len(cur))
	for i := range cur {
		curBy[cur[i].ident] = &cur[i]
	}
	var warns []string
	unknown := map[string]bool{}
	rows := make([]*fileRow, 0, len(ft.Rows))
	inFile := map[string]bool{}
	pendingSelf := map[string]bool{}
	for _, r := range ft.Rows {
		for c := range r.Values {
			if !ti.exportS[c] && !unknown[c] {
				unknown[c] = true
				warns = append(warns, "column_unknown:"+ti.t.Name+"."+c)
			}
		}
		key, ident := ti.identity(r.Values)
		fr := &fileRow{row: r, key: key, ident: ident, pick: pick(ti.t.Name, key)}
		inFile[ident] = true
		rows = append(rows, fr)
		if fr.pick {
			pendingSelf[key] = true
		}
	}
	out := make([]RowResult, len(rows))
	for progress := true; progress; {
		progress = false
		for i, fr := range rows {
			if fr.done {
				continue
			}
			if !fr.pick {
				fr.done = true
				out[i] = RowResult{Table: ti.t.Name, Key: fr.key, Status: StatusSkipped}
				continue
			}
			resolved, missing, deferred, err := s.resolve(ctx, ti, fr.row.Values, pendingSelf, fr.key)
			if err != nil {
				return nil, nil, err
			}
			if deferred {
				continue
			}
			fr.done = true
			progress = true
			delete(pendingSelf, fr.key)
			if len(missing) > 0 {
				out[i] = RowResult{Table: ti.t.Name, Key: fr.key, Status: StatusMissingRefs, Missing: missing}
				continue
			}
			rr, err := s.writeRow(ctx, tx, ti, fr, resolved, curBy[fr.ident], by)
			if err != nil {
				return nil, nil, err
			}
			out[i] = rr
		}
	}
	for i, fr := range rows {
		if fr.done {
			continue
		}
		_, missing, _, err := s.resolve(ctx, ti, fr.row.Values, map[string]bool{}, fr.key)
		if err != nil {
			return nil, nil, err
		}
		out[i] = RowResult{Table: ti.t.Name, Key: fr.key, Status: StatusMissingRefs, Missing: missing, Reason: "cycle"}
	}
	if switchOff && ti.t.Active != "" {
		for i := range cur {
			c := &cur[i]
			if inFile[c.ident] {
				continue
			}
			if on, _ := c.raw[ti.t.Active].(bool); !on {
				continue
			}
			rr, err := s.switchOff(ctx, tx, ti, c, by)
			if err != nil {
				return nil, nil, err
			}
			out = append(out, rr)
		}
	}
	return out, warns, nil
}

func (s *session) resolve(ctx context.Context, ti *tableInfo, values map[string]any, pendingSelf map[string]bool, self string) (map[string]any, []Missing, bool, error) {
	out := make(map[string]any, len(values))
	for c, v := range values {
		if ti.exportS[c] {
			out[c] = deepCopy(v)
		}
	}
	var missing []Missing
	deferred := false
	for i, r := range ti.t.Refs {
		v, ok := out[r.Column]
		if !ok {
			continue
		}
		ix, err := s.index(ctx, r)
		if err != nil {
			return nil, nil, false, err
		}
		out[r.Column] = mapPath(v, ti.paths[i], func(v any, inArray bool) (any, bool) {
			key, _ := v.(string)
			if !strings.HasPrefix(key, missingPrefix) {
				if id, ok := ix.byKey[key]; ok {
					return id, true
				}
				if r.Table == ti.t.Name && key != self && pendingSelf[key] {
					deferred = true
					return v, true
				}
			}
			missing = append(missing, Missing{Field: r.Column, Table: r.Table, Key: strings.TrimPrefix(key, missingPrefix)})
			return v, true
		})
	}
	return out, missing, deferred, nil
}

func (s *session) writeRow(ctx context.Context, tx Querier, ti *tableInfo, fr *fileRow, resolved map[string]any, cur *tableRow, by string) (RowResult, error) {
	rr := RowResult{Table: ti.t.Name, Key: fr.key}
	rec := map[string]any{}
	var cols []string
	if cur == nil {
		rr.Status = StatusNew
		for _, c := range ti.export {
			if v, ok := resolved[c]; ok {
				rec[c] = v
				cols = append(cols, c)
			}
		}
		if ti.uuidID {
			rec["id"] = uuid.Must(uuid.NewV4()).String()
			cols = append(cols, "id")
		}
		if _, ok := ti.types["updated_by"]; ok {
			rec["updated_by"] = by
			cols = append(cols, "updated_by")
		}
	} else {
		for _, c := range ti.export {
			fv, ok := fr.row.Values[c]
			if !ok {
				continue
			}
			if canon(fv) != canon(cur.values[c]) {
				rr.Changes = append(rr.Changes, Change{Field: c, From: cur.values[c], To: fv})
				v := resolved[c]
				if ti.isJSON(c) {
					v = RestoreSecrets(deepCopy(v), cur.raw[c])
				}
				rec[c] = v
				cols = append(cols, c)
			}
		}
		if len(cols) == 0 {
			rr.Status = StatusSame
			return rr, nil
		}
		rr.Status = StatusChanged
		for _, k := range ti.t.Key {
			rec[k] = resolved[k]
		}
	}
	raw, err := json.Marshal(rec)
	if err != nil {
		return rr, err
	}
	if _, err := tx.ExecContext(ctx, `SAVEPOINT ct_row`); err != nil {
		return rr, err
	}
	if cur == nil {
		quoted := make([]string, len(cols))
		for i, c := range cols {
			quoted[i] = quote(c)
		}
		list := strings.Join(quoted, ",")
		_, err = tx.ExecContext(ctx, `INSERT INTO `+quote(ti.t.Name)+` (`+list+`) SELECT `+list+` FROM jsonb_populate_record(NULL::`+quote(ti.t.Name)+`, $1::jsonb)`, string(raw))
	} else {
		sets := make([]string, 0, len(cols)+2)
		for _, c := range cols {
			sets = append(sets, quote(c)+" = n."+quote(c))
		}
		if _, ok := ti.types["updated_at"]; ok {
			sets = append(sets, `"updated_at" = now()`)
		}
		if _, ok := ti.types["updated_by"]; ok {
			sets = append(sets, `"updated_by" = $2`)
		}
		where := make([]string, 0, len(ti.t.Key))
		for _, k := range ti.t.Key {
			where = append(where, "d."+quote(k)+" IS NOT DISTINCT FROM n."+quote(k))
		}
		q := `UPDATE ` + quote(ti.t.Name) + ` AS d SET ` + strings.Join(sets, ", ") + ` FROM jsonb_populate_record(NULL::` + quote(ti.t.Name) + `, $1::jsonb) AS n WHERE ` + strings.Join(where, " AND ")
		if _, ok := ti.types["updated_by"]; ok {
			_, err = tx.ExecContext(ctx, q, string(raw), by)
		} else {
			_, err = tx.ExecContext(ctx, q, string(raw))
		}
	}
	if err != nil {
		if _, rerr := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT ct_row`); rerr != nil {
			return rr, errors.Join(err, rerr)
		}
		rr.Status = StatusError
		rr.Reason = pgMessage(err)
		return rr, nil
	}
	if _, err := tx.ExecContext(ctx, `RELEASE SAVEPOINT ct_row`); err != nil {
		return rr, err
	}
	if cur == nil {
		s.added(ti.t.Name, rec)
	}
	return rr, nil
}

func (s *session) added(table string, rec map[string]any) {
	for name, ix := range s.indexes {
		parts := strings.Split(name, "|")
		if len(parts) != 3 || parts[0] != table {
			continue
		}
		id, ok := rec[parts[1]]
		if !ok {
			delete(s.indexes, name)
			continue
		}
		var key []any
		for _, k := range strings.Split(parts[2], ",") {
			key = append(key, rec[k])
		}
		label := keyLabel(key)
		ids := partLabel(id)
		ix.byID[ids] = label
		ix.byKey[label] = ids
	}
}

func (s *session) switchOff(ctx context.Context, tx Querier, ti *tableInfo, c *tableRow, by string) (RowResult, error) {
	rr := RowResult{Table: ti.t.Name, Key: c.key, Status: StatusSwitchedOff, Changes: []Change{{Field: ti.t.Active, From: true, To: false}}}
	rec := map[string]any{}
	for _, k := range ti.t.Key {
		rec[k] = c.raw[k]
	}
	raw, err := json.Marshal(rec)
	if err != nil {
		return rr, err
	}
	sets := []string{quote(ti.t.Active) + " = false"}
	if _, ok := ti.types["updated_at"]; ok {
		sets = append(sets, `"updated_at" = now()`)
	}
	where := make([]string, 0, len(ti.t.Key))
	for _, k := range ti.t.Key {
		where = append(where, "d."+quote(k)+" IS NOT DISTINCT FROM n."+quote(k))
	}
	args := []any{string(raw)}
	if _, ok := ti.types["updated_by"]; ok {
		sets = append(sets, `"updated_by" = $2`)
		args = append(args, by)
	}
	_, err = tx.ExecContext(ctx, `UPDATE `+quote(ti.t.Name)+` AS d SET `+strings.Join(sets, ", ")+` FROM jsonb_populate_record(NULL::`+quote(ti.t.Name)+`, $1::jsonb) AS n WHERE `+strings.Join(where, " AND "), args...)
	return rr, err
}

func pgMessage(err error) string {
	var pg interface{ SQLState() string }
	msg := err.Error()
	if errors.As(err, &pg) {
		return pg.SQLState() + ": " + msg
	}
	return msg
}
