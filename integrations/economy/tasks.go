package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math/rand/v2"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/inapp"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

type dbRunner interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type taskDef struct {
	ID             string
	Key            string
	Name           string
	Description    string
	IconURL        string
	CategoryID     string
	TagIDs         []string
	Condition      string
	EventTypeID    string
	Metric         string
	Target         float64
	ValueMin       *float64
	ValueMax       *float64
	Repeat         string
	RepeatMinutes  int
	MaxCompletions int
	ParentID       string
	DependsOn      []string
	GroupID        string
	PrizeID        string
	GrantMode      string
	Notify         bool
	Audience       []string
	LevelTableID   string
	MinLevel       *int
	MaxLevel       *int
	StartsAt       *time.Time
	EndsAt         *time.Time
	Active         bool
	Sort           int
	CreatedAt      time.Time
}

type groupDef struct {
	ID           string
	Key          string
	Name         string
	Description  string
	IconURL      string
	CategoryID   string
	ActiveCount  int
	OrderMode    string
	Handout      string
	HandoutMins  int
	Repeatable   bool
	Audience     []string
	LevelTableID string
	MinLevel     *int
	MaxLevel     *int
	StartsAt     *time.Time
	EndsAt       *time.Time
	Active       bool
	Sort         int
}

const taskColumns = `t.id::text, t.key, t.name::text, t.description::text, t.icon_url, COALESCE(t.category_id::text,''), t.tag_ids::text, t.condition, COALESCE(t.event_type_id::text,''),
	t.metric, t.target::float8, t.value_min::float8, t.value_max::float8, t.repeat, COALESCE(t.repeat_minutes,0), COALESCE(t.max_completions,0), COALESCE(t.parent_id::text,''),
	t.depends_on::text, COALESCE(t.group_id::text,''), COALESCE(t.prize_id::text,''), t.grant_mode, t.notify, t.audience_tag_ids::text, COALESCE(t.level_table_id::text,''),
	t.min_level, t.max_level, t.starts_at, t.ends_at, t.is_active, t.sort, t.created_at`

func scanTask(rows interface{ Scan(dest ...any) error }) (*taskDef, error) {
	var t taskDef
	var tags, deps, audience string
	var vmin, vmax sql.NullFloat64
	var minL, maxL sql.NullInt64
	var starts, ends sql.NullTime
	if err := rows.Scan(&t.ID, &t.Key, &t.Name, &t.Description, &t.IconURL, &t.CategoryID, &tags, &t.Condition, &t.EventTypeID, &t.Metric, &t.Target, &vmin, &vmax,
		&t.Repeat, &t.RepeatMinutes, &t.MaxCompletions, &t.ParentID, &deps, &t.GroupID, &t.PrizeID, &t.GrantMode, &t.Notify, &audience, &t.LevelTableID,
		&minL, &maxL, &starts, &ends, &t.Active, &t.Sort, &t.CreatedAt); err != nil {
		return nil, err
	}
	t.TagIDs = parseTextArray(tags)
	t.DependsOn = parseTextArray(deps)
	t.Audience = parseTextArray(audience)
	if vmin.Valid {
		t.ValueMin = &vmin.Float64
	}
	if vmax.Valid {
		t.ValueMax = &vmax.Float64
	}
	if minL.Valid {
		v := int(minL.Int64)
		t.MinLevel = &v
	}
	if maxL.Valid {
		v := int(maxL.Int64)
		t.MaxLevel = &v
	}
	if starts.Valid {
		t.StartsAt = &starts.Time
	}
	if ends.Valid {
		t.EndsAt = &ends.Time
	}
	return &t, nil
}

func loadTasks(ctx context.Context, q dbRunner, where string, args ...any) ([]*taskDef, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+taskColumns+` FROM economy_task t `+where+` ORDER BY t.sort, t.created_at, t.id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*taskDef
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

const groupColumns = `g.id::text, g.key, g.name::text, g.description::text, g.icon_url, COALESCE(g.category_id::text,''), g.active_count, g.order_mode, g.handout, g.handout_minutes,
	g.is_repeatable, g.audience_tag_ids::text, COALESCE(g.level_table_id::text,''), g.min_level, g.max_level, g.starts_at, g.ends_at, g.is_active, g.sort`

func loadGroups(ctx context.Context, q dbRunner, where string, args ...any) ([]*groupDef, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+groupColumns+` FROM economy_task_group g `+where+` ORDER BY g.sort, g.created_at, g.id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*groupDef
	for rows.Next() {
		var g groupDef
		var audience string
		var minL, maxL sql.NullInt64
		var starts, ends sql.NullTime
		if err := rows.Scan(&g.ID, &g.Key, &g.Name, &g.Description, &g.IconURL, &g.CategoryID, &g.ActiveCount, &g.OrderMode, &g.Handout, &g.HandoutMins,
			&g.Repeatable, &audience, &g.LevelTableID, &minL, &maxL, &starts, &ends, &g.Active, &g.Sort); err != nil {
			return nil, err
		}
		g.Audience = parseTextArray(audience)
		if minL.Valid {
			v := int(minL.Int64)
			g.MinLevel = &v
		}
		if maxL.Valid {
			v := int(maxL.Int64)
			g.MaxLevel = &v
		}
		if starts.Valid {
			g.StartsAt = &starts.Time
		}
		if ends.Valid {
			g.EndsAt = &ends.Time
		}
		out = append(out, &g)
	}
	return out, rows.Err()
}

type userTaskRow struct {
	TaskID      string
	Scope       string
	WindowStart sql.NullTime
	Status      string
	Count       int64
	Sum         float64
	Max         float64
	CompletedAt sql.NullTime
	ClaimID     string
	ClaimedAt   sql.NullTime
	Error       string
	CreatedAt   time.Time
}

type groupState struct {
	Cycle  int
	Handed []string
	NextAt *time.Time
	Exists bool
}

type taskEnv struct {
	s       *service
	q       dbRunner
	userID  string
	now     time.Time
	clk     eventClock
	tags    func() (map[string]bool, error)
	tasks   map[string]*taskDef
	groups  map[string]*groupDef
	gstate  map[string]*groupState
	levels  map[string]int
	rows    map[string][]userTaskRow
	loaded  map[string]bool
	ignoreG bool
}

func (s *service) newTaskEnv(ctx context.Context, q dbRunner, userID string, now time.Time, clk eventClock, tags func() (map[string]bool, error)) *taskEnv {
	if tags == nil {
		var cached map[string]bool
		tags = func() (map[string]bool, error) {
			if cached != nil {
				return cached, nil
			}
			t, err := s.userTagIDs(ctx, userID)
			cached = t
			return t, err
		}
	}
	return &taskEnv{s: s, q: q, userID: userID, now: now, clk: clk, tags: tags, tasks: map[string]*taskDef{}, groups: map[string]*groupDef{}, gstate: map[string]*groupState{},
		levels: map[string]int{}, rows: map[string][]userTaskRow{}, loaded: map[string]bool{}}
}

func (e *taskEnv) addTasks(list []*taskDef) {
	for _, t := range list {
		e.tasks[t.ID] = t
	}
}

func (e *taskEnv) task(ctx context.Context, id string) (*taskDef, error) {
	if t, ok := e.tasks[id]; ok {
		return t, nil
	}
	if _, err := uuid.FromString(id); err != nil {
		return nil, nil
	}
	list, err := loadTasks(ctx, e.q, `WHERE t.id=$1`, id)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		e.tasks[id] = nil
		return nil, nil
	}
	e.tasks[id] = list[0]
	return list[0], nil
}

func (e *taskEnv) group(ctx context.Context, id string) (*groupDef, error) {
	if g, ok := e.groups[id]; ok {
		return g, nil
	}
	list, err := loadGroups(ctx, e.q, `WHERE g.id=$1`, id)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		e.groups[id] = nil
		return nil, nil
	}
	e.groups[id] = list[0]
	return list[0], nil
}

func (e *taskEnv) groupState(ctx context.Context, id string) (*groupState, error) {
	if st, ok := e.gstate[id]; ok {
		return st, nil
	}
	st := &groupState{Cycle: 1}
	var handed string
	var next sql.NullTime
	err := e.q.QueryRowContext(ctx, `SELECT cycle, handed::text, next_at FROM economy_user_task_group WHERE group_id=$1 AND user_id=$2`, id, e.userID).Scan(&st.Cycle, &handed, &next)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err == nil {
		st.Exists = true
		st.Handed = parseTextArray(handed)
		if next.Valid {
			st.NextAt = &next.Time
		}
	}
	e.gstate[id] = st
	return st, nil
}

func (e *taskEnv) ensureRows(ctx context.Context, ids []string) error {
	need := []string{}
	for _, id := range ids {
		if !e.loaded[id] {
			need = append(need, id)
		}
	}
	if len(need) == 0 {
		return nil
	}
	rows, err := e.q.QueryContext(ctx, `SELECT task_id::text, scope, window_start, status, count, sum::float8, COALESCE(max,0)::float8, completed_at, COALESCE(claim_id::text,''), claimed_at, COALESCE(error,''), created_at
		FROM economy_user_task WHERE user_id=$1 AND task_id::text = ANY($2) ORDER BY created_at`, e.userID, need)
	if err != nil {
		return err
	}
	defer rows.Close()
	for _, id := range need {
		e.loaded[id] = true
		e.rows[id] = nil
	}
	for rows.Next() {
		var r userTaskRow
		if err := rows.Scan(&r.TaskID, &r.Scope, &r.WindowStart, &r.Status, &r.Count, &r.Sum, &r.Max, &r.CompletedAt, &r.ClaimID, &r.ClaimedAt, &r.Error, &r.CreatedAt); err != nil {
			return err
		}
		e.rows[r.TaskID] = append(e.rows[r.TaskID], r)
	}
	return rows.Err()
}

func (e *taskEnv) forget(id string) {
	delete(e.loaded, id)
	delete(e.rows, id)
}

func (e *taskEnv) row(ctx context.Context, taskID, scope string) (*userTaskRow, error) {
	if err := e.ensureRows(ctx, []string{taskID}); err != nil {
		return nil, err
	}
	for i := range e.rows[taskID] {
		if e.rows[taskID][i].Scope == scope {
			return &e.rows[taskID][i], nil
		}
	}
	return nil, nil
}

func (e *taskEnv) level(ctx context.Context, tableID string) (int, error) {
	if v, ok := e.levels[tableID]; ok {
		return v, nil
	}
	var lv int
	err := e.q.QueryRowContext(ctx, `SELECT GREATEST(COALESCE((SELECT level FROM economy_user_level WHERE table_id=$1 AND user_id=$2),0),
			COALESCE((SELECT max(s.level) FROM economy_level_step s JOIN economy_level_table lt ON lt.id=s.table_id JOIN economy_meta m ON m.id=lt.meta_id
				LEFT JOIN economy_user_meta um ON um.meta_id=lt.meta_id AND um.user_id=$2 WHERE s.table_id=$1 AND s.threshold <= COALESCE(um.value, m.default_value)),0))`, tableID, e.userID).Scan(&lv)
	if err != nil {
		return 0, err
	}
	e.levels[tableID] = lv
	return lv, nil
}

func (e *taskEnv) targeted(ctx context.Context, audience []string, tableID string, minL, maxL *int) (bool, error) {
	ok, err := audienceAllows(audience, e.tags)
	if err != nil || !ok {
		return false, err
	}
	if tableID != "" && (minL != nil || maxL != nil) {
		lv, err := e.level(ctx, tableID)
		if err != nil {
			return false, err
		}
		if minL != nil && lv < *minL {
			return false, nil
		}
		if maxL != nil && lv > *maxL {
			return false, nil
		}
	}
	return true, nil
}

func inWindow(starts, ends *time.Time, now time.Time) bool {
	if starts != nil && now.Before(*starts) {
		return false
	}
	if ends != nil && !now.Before(*ends) {
		return false
	}
	return true
}

func (e *taskEnv) depsDone(ctx context.Context, t *taskDef) (bool, error) {
	if len(t.DependsOn) == 0 {
		return true, nil
	}
	if err := e.ensureRows(ctx, t.DependsOn); err != nil {
		return false, err
	}
	for _, id := range t.DependsOn {
		done := false
		for _, r := range e.rows[id] {
			if r.Status != "active" {
				done = true
				break
			}
		}
		if !done {
			return false, nil
		}
	}
	return true, nil
}

type taskSlot struct {
	State       string
	Reason      string
	Scope       string
	WindowStart *time.Time
	PeriodEnd   *time.Time
	ExpiresAt   *time.Time
	NextAt      *time.Time
	Completions int
}

func (sl taskSlot) open() bool { return sl.State == "open" }

func minTime(a, b *time.Time) *time.Time {
	if a == nil {
		return b
	}
	if b == nil || a.Before(*b) {
		return a
	}
	return b
}

func (e *taskEnv) groupOpen(ctx context.Context, g *groupDef) (bool, string, error) {
	if g == nil || !g.Active {
		return false, "groupInactive", nil
	}
	if !inWindow(g.StartsAt, g.EndsAt, e.now) {
		return false, "groupClosed", nil
	}
	ok, err := e.targeted(ctx, g.Audience, g.LevelTableID, g.MinLevel, g.MaxLevel)
	if err != nil {
		return false, "", err
	}
	if !ok {
		return false, "notTargeted", nil
	}
	return true, "", nil
}

func (e *taskEnv) slot(ctx context.Context, t *taskDef) (taskSlot, error) {
	if t == nil || !t.Active {
		return taskSlot{State: "hidden", Reason: "inactive"}, nil
	}
	if t.ParentID != "" {
		parent, err := e.task(ctx, t.ParentID)
		if err != nil {
			return taskSlot{}, err
		}
		ps, err := e.slot(ctx, parent)
		if err != nil {
			return taskSlot{}, err
		}
		if !ps.open() {
			if ps.State == "hidden" {
				ps.Reason = "parent"
			}
			return ps, nil
		}
		ok, err := e.depsDone(ctx, t)
		if err != nil {
			return taskSlot{}, err
		}
		if !ok {
			ps.State, ps.Reason = "locked", "dependsOn"
		}
		return ps, nil
	}
	if !inWindow(t.StartsAt, t.EndsAt, e.now) {
		reason := "ended"
		if t.StartsAt != nil && e.now.Before(*t.StartsAt) {
			reason = "notStarted"
		}
		return taskSlot{State: "hidden", Reason: reason}, nil
	}
	ok, err := e.targeted(ctx, t.Audience, t.LevelTableID, t.MinLevel, t.MaxLevel)
	if err != nil {
		return taskSlot{}, err
	}
	if !ok {
		return taskSlot{State: "hidden", Reason: "notTargeted"}, nil
	}
	sl := taskSlot{ExpiresAt: t.EndsAt}
	if t.GroupID != "" && !e.ignoreG {
		g, err := e.group(ctx, t.GroupID)
		if err != nil {
			return taskSlot{}, err
		}
		open, reason, err := e.groupOpen(ctx, g)
		if err != nil {
			return taskSlot{}, err
		}
		if !open {
			return taskSlot{State: "hidden", Reason: reason}, nil
		}
		st, err := e.groupState(ctx, t.GroupID)
		if err != nil {
			return taskSlot{}, err
		}
		if !containsString(st.Handed, t.ID) {
			return taskSlot{State: "hidden", Reason: "notHanded"}, nil
		}
		sl.Scope = "g" + strconv.Itoa(st.Cycle)
		sl.ExpiresAt = minTime(t.EndsAt, g.EndsAt)
		ok, err := e.depsDone(ctx, t)
		if err != nil {
			return taskSlot{}, err
		}
		sl.State = "open"
		if !ok {
			sl.State, sl.Reason = "locked", "dependsOn"
		}
		return sl, nil
	}
	ok, err = e.depsDone(ctx, t)
	if err != nil {
		return taskSlot{}, err
	}
	if !ok {
		return taskSlot{State: "locked", Reason: "dependsOn", ExpiresAt: t.EndsAt}, nil
	}
	if err := e.ensureRows(ctx, []string{t.ID}); err != nil {
		return taskSlot{}, err
	}
	rows := e.rows[t.ID]
	completions := 0
	var lastDone *time.Time
	var latest *userTaskRow
	for i := range rows {
		r := &rows[i]
		if r.Status != "active" {
			completions++
			if r.CompletedAt.Valid && (lastDone == nil || r.CompletedAt.Time.After(*lastDone)) {
				d := r.CompletedAt.Time
				lastDone = &d
			}
		}
		if latest == nil || r.CreatedAt.After(latest.CreatedAt) {
			latest = r
		}
	}
	sl.Completions = completions
	sl.State = "open"
	switch t.Repeat {
	case "daily", "weekly":
		period := "day"
		prefix := "d"
		if t.Repeat == "weekly" {
			period, prefix = "week", "w"
		}
		start := e.clk.periodStart(period, e.now)
		sl.Scope = prefix + dateKey(start)
		sl.WindowStart = &start
		sl.PeriodEnd = e.clk.periodEnd(period, e.now)
		sl.ExpiresAt = minTime(sl.PeriodEnd, t.EndsAt)
		if t.MaxCompletions > 0 {
			current := false
			for _, r := range rows {
				if r.Scope == sl.Scope && r.Status != "active" {
					current = true
				}
			}
			if !current && completions >= t.MaxCompletions {
				sl.State, sl.Reason = "done", "maxCompletions"
			}
		}
	case "interval":
		if latest != nil && latest.Status == "active" {
			sl.Scope = latest.Scope
			return sl, nil
		}
		if t.MaxCompletions > 0 && completions >= t.MaxCompletions {
			sl.State, sl.Reason = "done", "maxCompletions"
			if latest != nil {
				sl.Scope = latest.Scope
			}
			return sl, nil
		}
		if lastDone != nil {
			next := lastDone.Add(time.Duration(t.RepeatMinutes) * time.Minute)
			if e.now.Before(next) {
				sl.State, sl.Reason, sl.NextAt = "cooldown", "interval", &next
				if latest != nil {
					sl.Scope = latest.Scope
				}
				return sl, nil
			}
		}
		sl.Scope = "i" + strconv.Itoa(completions+1)
	default:
		sl.Scope = "once"
	}
	return sl, nil
}

func containsString(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func taskMeasure(metric string, count int64, sum, max float64) float64 {
	switch metric {
	case "sum_value":
		return sum
	case "high_value":
		return max
	}
	return float64(count)
}

type taskChange struct {
	TaskID    string  `json:"taskId"`
	Key       string  `json:"key"`
	Scope     string  `json:"scope"`
	Value     float64 `json:"value"`
	Target    float64 `json:"target"`
	Completed bool    `json:"completed"`
	Status    string  `json:"status"`
	ClaimID   string  `json:"claimId,omitempty"`
	GrantMode string  `json:"grantMode,omitempty"`
	notify    bool
	groupID   string
	name      string
}

func (e *taskEnv) upsertProgress(ctx context.Context, t *taskDef, sl taskSlot, value float64) (bool, int64, float64, float64, error) {
	var ws any
	if sl.WindowStart != nil {
		ws = dateKey(*sl.WindowStart)
	}
	var count int64
	var sum, max float64
	err := e.q.QueryRowContext(ctx, `INSERT INTO economy_user_task AS p (task_id, user_id, scope, window_start, status, count, sum, max, created_at, updated_at) VALUES ($1,$2,$3,$4::date,'active',1,$5::numeric,$5::numeric,$6,$6)
		ON CONFLICT (task_id, user_id, scope) DO UPDATE SET count=p.count+1, sum=p.sum+EXCLUDED.sum, max=GREATEST(p.max, EXCLUDED.max), updated_at=EXCLUDED.updated_at WHERE p.status='active'
		RETURNING count, sum::float8, COALESCE(max,0)::float8`, t.ID, e.userID, sl.Scope, ws, formatAmount(value), e.now).Scan(&count, &sum, &max)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, 0, 0, nil
	}
	if err != nil {
		return false, 0, 0, 0, err
	}
	return true, count, sum, max, nil
}

func (e *taskEnv) lockRow(ctx context.Context, t *taskDef, sl taskSlot) (string, error) {
	var ws any
	if sl.WindowStart != nil {
		ws = dateKey(*sl.WindowStart)
	}
	var status string
	err := e.q.QueryRowContext(ctx, `INSERT INTO economy_user_task AS p (task_id, user_id, scope, window_start, status, created_at, updated_at) VALUES ($1,$2,$3,$4::date,'active',$5,$5)
		ON CONFLICT (task_id, user_id, scope) DO UPDATE SET updated_at=p.updated_at RETURNING status`, t.ID, e.userID, sl.Scope, ws, e.now).Scan(&status)
	return status, err
}

func (e *taskEnv) complete(ctx context.Context, t *taskDef, scope, by string, value float64, out *eventOutcome) error {
	claimID := ""
	status := "claimed"
	if t.PrizeID != "" {
		claimID = uuid.Must(uuid.NewV4()).String()
		status = "completed"
	}
	var claimArg any
	if claimID != "" {
		claimArg = claimID
	}
	res, err := e.q.ExecContext(ctx, `UPDATE economy_user_task SET status=$4, completed_at=$5, completed_by=$6, claim_id=$7, claimed_at=CASE WHEN $4='claimed' THEN $5::timestamptz END, updated_at=$5
		WHERE task_id=$1 AND user_id=$2 AND scope=$3 AND status='active'`, t.ID, e.userID, scope, status, e.now, by, claimArg)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil
	}
	e.forget(t.ID)
	ch := taskChange{TaskID: t.ID, Key: t.Key, Scope: scope, Value: value, Target: t.Target, Completed: true, Status: status, notify: t.Notify, groupID: t.GroupID, name: t.Name}
	if claimID != "" {
		ch.GrantMode = t.GrantMode
		if t.GrantMode == "auto" {
			ref, err := insertClaimWithID(ctx, e.q, claimID, e.userID, "task", t.ID, t.ID, scope, t.PrizeID, "auto", value)
			if err != nil {
				return err
			}
			if ref != nil {
				out.Claims = append(out.Claims, *ref)
				ch.ClaimID = ref.ClaimID
			}
		} else {
			ch.ClaimID = claimID
		}
	}
	if t.GroupID == "" && t.ParentID != "" {
		if p, _ := e.task(ctx, t.ParentID); p != nil {
			ch.groupID = p.GroupID
		}
	}
	out.Tasks = append(out.Tasks, ch)
	if t.ParentID != "" {
		parent, err := e.task(ctx, t.ParentID)
		if err != nil {
			return err
		}
		if parent != nil && parent.Active && parent.Condition == "children" {
			return e.checkParent(ctx, parent, scope, out)
		}
	}
	return nil
}

func (e *taskEnv) checkParent(ctx context.Context, parent *taskDef, scope string, out *eventOutcome) error {
	sl, err := e.slot(ctx, parent)
	if err != nil {
		return err
	}
	if !sl.open() || sl.Scope != scope {
		return nil
	}
	status, err := e.lockRow(ctx, parent, sl)
	if err != nil {
		return err
	}
	if status != "active" {
		return nil
	}
	var total, done int64
	if err := e.q.QueryRowContext(ctx, `SELECT count(*), count(ut.task_id) FROM economy_task c
		LEFT JOIN economy_user_task ut ON ut.task_id=c.id AND ut.user_id=$2 AND ut.scope=$3 AND ut.status<>'active'
		WHERE c.parent_id=$1 AND c.is_active`, parent.ID, e.userID, scope).Scan(&total, &done); err != nil {
		return err
	}
	if _, err := e.q.ExecContext(ctx, `UPDATE economy_user_task SET count=$4, updated_at=$5 WHERE task_id=$1 AND user_id=$2 AND scope=$3 AND status='active'`, parent.ID, e.userID, scope, done, e.now); err != nil {
		return err
	}
	e.forget(parent.ID)
	if total > 0 && done >= total {
		return e.complete(ctx, parent, scope, "player", float64(done), out)
	}
	return nil
}

func valueAllowed(t *taskDef, value float64) bool {
	if t.ValueMin != nil && value < *t.ValueMin {
		return false
	}
	if t.ValueMax != nil && value > *t.ValueMax {
		return false
	}
	return true
}

func (s *service) applyTasks(ctx context.Context, tx *sql.Tx, userID, typeID string, value float64, now time.Time, clk eventClock, tags func() (map[string]bool, error), out *eventOutcome) error {
	list, err := loadTasks(ctx, tx, `WHERE t.event_type_id=$1 AND t.condition='event' AND t.is_active`, typeID)
	if err != nil || len(list) == 0 {
		return err
	}
	e := s.newTaskEnv(ctx, tx, userID, now, clk, tags)
	e.addTasks(list)
	type due struct {
		t  *taskDef
		sl taskSlot
	}
	var todo []due
	for _, t := range list {
		if !valueAllowed(t, value) {
			continue
		}
		sl, err := e.slot(ctx, t)
		if err != nil {
			return err
		}
		if sl.open() {
			todo = append(todo, due{t, sl})
		}
	}
	for _, d := range todo {
		t, sl := d.t, d.sl
		ok, count, sum, max, err := e.upsertProgress(ctx, t, sl, value)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		e.forget(t.ID)
		measured := taskMeasure(t.Metric, count, sum, max)
		if measured+1e-9 >= t.Target {
			if err := e.complete(ctx, t, sl.Scope, "player", measured, out); err != nil {
				return err
			}
		} else {
			out.Tasks = append(out.Tasks, taskChange{TaskID: t.ID, Key: t.Key, Scope: sl.Scope, Value: measured, Target: t.Target, Status: "active"})
		}
	}
	return nil
}

func insertClaimWithID(ctx context.Context, q dbRunner, id, userID, kind, ownerID, refID, scope, prizeID, mode string, value float64) (*claimRef, error) {
	var got string
	err := q.QueryRowContext(ctx, `INSERT INTO economy_reward_claim (id, user_id, kind, owner_id, ref_id, scope, prize_id, grant_mode, value) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::numeric)
		ON CONFLICT (kind, ref_id, user_id, scope) DO NOTHING RETURNING id::text`, id, userID, kind, ownerID, refID, scope, prizeID, mode, formatAmount(value)).Scan(&got)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &claimRef{ClaimID: got, Kind: kind, PrizeID: prizeID, Mode: mode, Status: "ready"}, nil
}

func (s *service) afterTaskGranted(ctx context.Context, claimID string) {
	var userID, groupID string
	err := s.db.QueryRowContext(ctx, `UPDATE economy_user_task ut SET status='claimed', claimed_at=now(), error=NULL, updated_at=now() FROM economy_task t
		WHERE ut.claim_id=$1 AND ut.status='completed' AND t.id=ut.task_id RETURNING ut.user_id::text, COALESCE(t.group_id::text, (SELECT p.group_id::text FROM economy_task p WHERE p.id=t.parent_id), '')`, claimID).Scan(&userID, &groupID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			s.logger.Warn("Economy task state after reward %s not saved: %v", claimID, err)
		}
		return
	}
	if groupID != "" {
		if _, err := s.refreshGroups(ctx, userID, []string{groupID}, ""); err != nil {
			s.logger.Warn("Economy task group refresh for %s failed: %v", userID, err)
		}
	}
}

func (s *service) settleTasks(ctx context.Context, userID string, out *eventOutcome) {
	groups := []string{}
	var notify []taskChange
	for _, ch := range out.Tasks {
		if !ch.Completed {
			continue
		}
		if ch.groupID != "" && !containsString(groups, ch.groupID) {
			groups = append(groups, ch.groupID)
		}
		if ch.notify {
			notify = append(notify, ch)
		}
	}
	if len(groups) > 0 {
		if _, err := s.refreshGroups(ctx, userID, groups, ""); err != nil {
			s.logger.Warn("Economy task group refresh for %s failed: %v", userID, err)
		}
	}
	for _, ch := range notify {
		s.notifyTask(ctx, userID, ch)
	}
}

var taskNoticeTexts = map[string]map[string]string{
	"title":        {"fa": "مأموریت انجام شد", "en": "Mission complete", "ar": "اكتملت المهمة"},
	"body.claim":   {"fa": "«{name}» تمام شد. جایزه‌ات را بگیر.", "en": "“{name}” is done. Collect your prize.", "ar": "اكتملت «{name}». اجمع جائزتك."},
	"body.done":    {"fa": "«{name}» تمام شد.", "en": "“{name}” is done.", "ar": "اكتملت «{name}»."},
	"action.label": {"fa": "دیدن مأموریت‌ها", "en": "Open missions", "ar": "عرض المهام"},
}

func (s *service) notifyTask(ctx context.Context, userID string, ch taskChange) {
	if !inapp.Available() {
		return
	}
	names := localized.Parse(ch.name)
	key := "body.done"
	if ch.ClaimID != "" && ch.GrantMode == "claim" {
		key = "body.claim"
	}
	body := map[string]string{}
	for lang, tpl := range taskNoticeTexts[key] {
		name := names[lang]
		if name == "" {
			name = names["fa"]
		}
		if name == "" {
			name = ch.Key
		}
		body[lang] = strings.ReplaceAll(tpl, "{name}", name)
	}
	if _, err := inapp.Send(ctx, inapp.Message{UserIDs: []string{userID}, Title: taskNoticeTexts["title"], Body: body,
		Action: &inapp.Action{Type: "goto", Value: "tasks", Label: taskNoticeTexts["action.label"]}, CreatedBy: "economy:task"}); err != nil {
		s.logger.Warn("Economy task notice for %s not sent: %v", userID, err)
	}
}

type taskClaimResult struct {
	ClaimID string
	Rewards []prizeItem
	PrizeID string
	Extra   *eventOutcome
	Task    *taskDef
}

func (s *service) claimTask(ctx context.Context, userID, taskID, claimID, actor string) (*taskClaimResult, error) {
	var res taskClaimResult
	var scope, status, cid string
	var err error
	if claimID != "" {
		err = s.db.QueryRowContext(ctx, `SELECT task_id::text, scope, status, COALESCE(claim_id::text,'') FROM economy_user_task WHERE claim_id=$2 AND user_id=$1`, userID, claimID).Scan(&taskID, &scope, &status, &cid)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("CLAIM_NOT_FOUND")
		}
	} else {
		err = s.db.QueryRowContext(ctx, `SELECT task_id::text, scope, status, COALESCE(claim_id::text,'') FROM economy_user_task WHERE task_id=$2 AND user_id=$1 AND status='completed' ORDER BY completed_at DESC LIMIT 1`, userID, taskID).Scan(&taskID, &scope, &status, &cid)
		if errors.Is(err, sql.ErrNoRows) {
			var claimed, active bool
			if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_user_task WHERE task_id=$1 AND user_id=$2 AND status='claimed'),
				EXISTS (SELECT 1 FROM economy_user_task WHERE task_id=$1 AND user_id=$2 AND status='active')`, taskID, userID).Scan(&claimed, &active); err != nil {
				return nil, err
			}
			if claimed && !active {
				return nil, conflict("ALREADY_CLAIMED", "this reward was already received")
			}
			return nil, conflict("TASK_NOT_COMPLETED", "this mission is not finished yet")
		}
	}
	if err != nil {
		return nil, err
	}
	if status == "claimed" {
		return nil, conflict("ALREADY_CLAIMED", "this reward was already received")
	}
	if status != "completed" || cid == "" {
		return nil, conflict("TASK_NOT_COMPLETED", "this mission is not finished yet")
	}
	list, err := loadTasks(ctx, s.db, `WHERE t.id=$1`, taskID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, notFound("TASK_NOT_FOUND")
	}
	t := list[0]
	res.Task = t
	groupID := t.GroupID
	if groupID == "" && t.ParentID != "" {
		_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(group_id::text,'') FROM economy_task WHERE id=$1`, t.ParentID).Scan(&groupID)
	}
	createdHere := false
	var existingPrize string
	err = s.db.QueryRowContext(ctx, `SELECT prize_id::text FROM economy_reward_claim WHERE id=$1`, cid).Scan(&existingPrize)
	if errors.Is(err, sql.ErrNoRows) {
		if t.PrizeID == "" {
			return nil, conflict("TASK_HAS_NO_PRIZE", "this mission has no prize")
		}
		ref, err := insertClaimWithID(ctx, s.db, cid, userID, "task", t.ID, t.ID, scope, t.PrizeID, "claim", 0)
		if err != nil {
			return nil, err
		}
		createdHere = ref != nil
	} else if err != nil {
		return nil, err
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var prizeID, claimStatus string
		err := tx.QueryRowContext(ctx, `SELECT prize_id::text, status FROM economy_reward_claim WHERE id=$1 AND user_id=$2 FOR UPDATE`, cid, userID).Scan(&prizeID, &claimStatus)
		if errors.Is(err, sql.ErrNoRows) {
			return conflict("TASK_NOT_COMPLETED", "this mission is not finished yet")
		}
		if err != nil {
			return err
		}
		if claimStatus == "granted" {
			if _, err := tx.ExecContext(ctx, `UPDATE economy_user_task SET status='claimed', claimed_at=COALESCE(claimed_at, now()), error=NULL, updated_at=now() WHERE claim_id=$1 AND status<>'claimed'`, cid); err != nil {
				return err
			}
			return conflict("ALREADY_CLAIMED", "this reward was already received")
		}
		rewards, err := s.grant(ctx, grantRequest{UserID: userID, PrizeID: prizeID, SourceType: "task", SourceID: cid, Reason: "task_reward"})
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_reward_claim SET status='granted', error=NULL, granted_at=now(), attempted_at=now(), attempts=attempts+1 WHERE id=$1`, cid); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_user_task SET status='claimed', claimed_at=now(), error=NULL, updated_at=now(), completed_by=COALESCE(completed_by,$2) WHERE claim_id=$1`, cid, actor); err != nil {
			return err
		}
		res.ClaimID, res.Rewards, res.PrizeID = cid, rewards, prizeID
		return nil
	})
	if err != nil {
		if app, ok := asApp(err); ok && (app.Code == "ALREADY_CLAIMED" || app.Code == "TASK_NOT_COMPLETED") {
			return nil, err
		}
		msg := err.Error()
		if app, ok := asApp(err); ok {
			msg = app.Code
		}
		if createdHere {
			_, _ = s.db.ExecContext(ctx, `DELETE FROM economy_reward_claim WHERE id=$1 AND status<>'granted' AND grant_mode='claim'`, cid)
		} else {
			_, _ = s.db.ExecContext(ctx, `UPDATE economy_reward_claim SET status='failed', error=$2, attempted_at=now() WHERE id=$1 AND status<>'granted'`, cid, truncate(msg, 500))
		}
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_user_task SET error=$2, updated_at=now() WHERE claim_id=$1 AND status='completed'`, cid, truncate(msg, 500))
		return nil, err
	}
	if res.Rewards == nil {
		res.Rewards = []prizeItem{}
	}
	var metaKeys []string
	for _, item := range res.Rewards {
		if item.Type == "meta" {
			metaKeys = append(metaKeys, item.Key)
		}
	}
	res.Extra = newOutcome()
	if len(metaKeys) > 0 {
		if ids, err := s.metaIDsByKey(ctx, metaKeys); err == nil {
			extra, err := s.evaluateLevels(ctx, userID, ids, 2)
			if err != nil {
				s.logger.Warn("Economy level check after task reward %s failed: %v", res.ClaimID, err)
			}
			res.Extra.merge(extra)
		}
	}
	if groupID != "" {
		if _, err := s.refreshGroups(ctx, userID, []string{groupID}, ""); err != nil {
			s.logger.Warn("Economy task group refresh for %s failed: %v", userID, err)
		}
	}
	return &res, nil
}

type groupRefresh struct {
	GroupID string
	Handed  []string
	Error   *appError
}

func (s *service) refreshGroups(ctx context.Context, userID string, groupIDs []string, request string) ([]groupRefresh, error) {
	settings, err := loadEventSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if groupIDs == nil {
		groups, err := loadGroups(ctx, s.db, `WHERE g.is_active`)
		if err != nil {
			return nil, err
		}
		for _, g := range groups {
			groupIDs = append(groupIDs, g.ID)
		}
	}
	out := []groupRefresh{}
	for _, gid := range groupIDs {
		var r groupRefresh
		r.GroupID = gid
		err := s.inTx(ctx, func(tx *sql.Tx) error {
			handed, err := s.refreshGroup(ctx, tx, userID, gid, request == gid, settings.clock())
			r.Handed = handed
			return err
		})
		if err != nil {
			if app, ok := asApp(err); ok && request == gid {
				r.Error = app
			} else if !ok {
				return out, err
			}
		}
		out = append(out, r)
	}
	return out, nil
}

func (s *service) refreshGroup(ctx context.Context, tx *sql.Tx, userID, groupID string, request bool, clk eventClock) ([]string, error) {
	now := s.now()
	e := s.newTaskEnv(ctx, tx, userID, now, clk, nil)
	g, err := e.group(ctx, groupID)
	if err != nil {
		return nil, err
	}
	open, _, err := e.groupOpen(ctx, g)
	if err != nil {
		return nil, err
	}
	if !open {
		if request {
			return nil, notFound("TASK_GROUP_NOT_FOUND")
		}
		return nil, nil
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_task_group (group_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, groupID, userID); err != nil {
		if isForeignKeyViolation(err) {
			return nil, notFound("USER_NOT_FOUND")
		}
		return nil, err
	}
	st := &groupState{Exists: true}
	var handedRaw string
	var next sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT cycle, handed::text, next_at FROM economy_user_task_group WHERE group_id=$1 AND user_id=$2 FOR UPDATE`, groupID, userID).Scan(&st.Cycle, &handedRaw, &next); err != nil {
		return nil, err
	}
	st.Handed = parseTextArray(handedRaw)
	if next.Valid {
		st.NextAt = &next.Time
	}
	e.gstate[groupID] = st
	members, err := loadTasks(ctx, tx, `WHERE t.group_id=$1 AND t.parent_id IS NULL`, groupID)
	if err != nil {
		return nil, err
	}
	e.addTasks(members)
	ids := make([]string, 0, len(members))
	for _, t := range members {
		ids = append(ids, t.ID)
	}
	if err := e.ensureRows(ctx, ids); err != nil {
		return nil, err
	}
	scope := "g" + strconv.Itoa(st.Cycle)
	finishedInCycle := func(id string) bool {
		t := e.tasks[id]
		if t == nil || !t.Active {
			return true
		}
		for _, r := range e.rows[id] {
			if r.Scope == scope && r.Status == "claimed" {
				return true
			}
		}
		return false
	}
	e.ignoreG = true
	eligible := func(t *taskDef) (bool, error) {
		sl, err := e.slot(ctx, t)
		if err != nil {
			return false, err
		}
		return sl.open(), nil
	}
	pool := func() ([]*taskDef, error) {
		var out []*taskDef
		for _, t := range members {
			if containsString(st.Handed, t.ID) {
				continue
			}
			ok, err := eligible(t)
			if err != nil {
				return nil, err
			}
			if ok {
				out = append(out, t)
			}
		}
		return out, nil
	}
	unfinished := 0
	for _, id := range st.Handed {
		if !finishedInCycle(id) {
			unfinished++
		}
	}
	remaining, err := pool()
	if err != nil {
		return nil, err
	}
	changed := false
	if unfinished == 0 && len(remaining) == 0 && len(st.Handed) > 0 {
		if !g.Repeatable {
			if request {
				return nil, conflict("TASK_GROUP_FINISHED", "every mission of this group is done")
			}
			return nil, nil
		}
		st.Cycle++
		st.Handed = nil
		st.NextAt = nil
		scope = "g" + strconv.Itoa(st.Cycle)
		changed = true
		if g.Handout == "timer" && g.HandoutMins > 0 {
			n := now.Add(time.Duration(g.HandoutMins) * time.Minute)
			st.NextAt = &n
		}
		remaining, err = pool()
		if err != nil {
			return nil, err
		}
	}
	free := g.ActiveCount - unfinished
	if free < 0 {
		free = 0
	}
	give := 0
	switch g.Handout {
	case "auto":
		give = free
	case "request":
		if request {
			if free == 0 {
				return nil, conflict("CURRENT_TASK_MUST_BE_DONE", "finish the current missions first")
			}
			give = 1
		}
	case "timer":
		first := st.Cycle == 1 && len(st.Handed) == 0 && st.NextAt == nil
		if free == 0 {
			if st.NextAt != nil {
				st.NextAt = nil
				changed = true
			}
			if request {
				return nil, conflict("CURRENT_TASK_MUST_BE_DONE", "finish the current missions first")
			}
		} else if first || g.HandoutMins == 0 {
			give = free
		} else if st.NextAt == nil {
			n := now.Add(time.Duration(g.HandoutMins) * time.Minute)
			st.NextAt = &n
			changed = true
		} else if !now.Before(*st.NextAt) {
			give = free
			st.NextAt = nil
			changed = true
		} else if request {
			return nil, fail(http.StatusConflict, "DURATION_NOT_ELAPSED", "nextAt", "the next mission comes at "+st.NextAt.UTC().Format(time.RFC3339))
		}
	}
	if give > len(remaining) {
		give = len(remaining)
	}
	if request && give == 0 && len(remaining) == 0 {
		return nil, conflict("NO_TASK_AVAILABLE", "no mission is available right now")
	}
	if g.OrderMode == "random" {
		rand.Shuffle(len(remaining), func(i, j int) { remaining[i], remaining[j] = remaining[j], remaining[i] })
	}
	handed := []string{}
	for i := 0; i < give; i++ {
		t := remaining[i]
		st.Handed = append(st.Handed, t.ID)
		handed = append(handed, t.ID)
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_task (task_id, user_id, scope, status, created_at, updated_at) VALUES ($1,$2,$3,'active',$4,$4) ON CONFLICT DO NOTHING`, t.ID, userID, scope, now); err != nil {
			return nil, err
		}
		changed = true
	}
	if changed {
		var nextArg any
		if st.NextAt != nil {
			nextArg = *st.NextAt
		}
		if st.Handed == nil {
			st.Handed = []string{}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_user_task_group SET cycle=$3, handed=$4, next_at=$5, updated_at=now() WHERE group_id=$1 AND user_id=$2`, groupID, userID, st.Cycle, st.Handed, nextArg); err != nil {
			return nil, err
		}
	}
	return handed, nil
}

func sortTasks(list []*taskDef) {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Sort != list[j].Sort {
			return list[i].Sort < list[j].Sort
		}
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})
}

func rawJSON(v string) json.RawMessage {
	if v == "" {
		return json.RawMessage("{}")
	}
	return json.RawMessage(v)
}
