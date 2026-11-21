package store

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	DriverNakama     = "nakama"
	DriverPostgres   = "postgres"
	DriverClickHouse = "clickhouse"
	DriverMongo      = "mongodb"

	MetricCount = "count"
	MetricSum   = "sum"
	MetricAvg   = "avg"
	MetricMin   = "min"
	MetricMax   = "max"
	MetricUsers = "users"

	BucketNone = "none"
	BucketHour = "hour"
	BucketDay  = "day"
	BucketWeek = "week"

	ModeClassic = "classic"
	ModeRolling = "rolling"
)

var (
	ErrBadQuery = errors.New("insights store: bad query")

	groupKey   = regexp.MustCompile(`^(tag:[a-z][a-z0-9_]{0,31}|name|platform|appVersion|user|source|game)$`)
	tagGroupRe = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
)

type Event struct {
	ID         string            `json:"id"`
	Game       string            `json:"game"`
	Name       string            `json:"name"`
	Value      *float64          `json:"value,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
	Fields     map[string]any    `json:"fields,omitempty"`
	UserID     string            `json:"userId"`
	SessionID  string            `json:"sessionId,omitempty"`
	At         time.Time         `json:"at"`
	Day        string            `json:"day"`
	ServerAt   time.Time         `json:"serverAt"`
	Platform   string            `json:"platform,omitempty"`
	AppVersion string            `json:"appVersion,omitempty"`
	Source     string            `json:"source"`
	Flagged    bool              `json:"flagged,omitempty"`
	FlagReason string            `json:"flagReason,omitempty"`
}

type Filter struct {
	From           string
	To             string
	TZ             string
	Game           *string
	Names          []string
	Platform       string
	AppVersion     string
	UserID         string
	SessionID      string
	Source         string
	Tags           map[string][]string
	IncludeFlagged bool
}

func (f Filter) TagGroups() []string {
	out := make([]string, 0, len(f.Tags))
	for g := range f.Tags {
		out = append(out, g)
	}
	sort.Strings(out)
	return out
}

func (f Filter) Validate() error {
	if f.From != "" {
		if _, err := time.Parse(time.DateOnly, f.From); err != nil {
			return ErrBadQuery
		}
	}
	if f.To != "" {
		if _, err := time.Parse(time.DateOnly, f.To); err != nil {
			return ErrBadQuery
		}
	}
	for g, ids := range f.Tags {
		if !tagGroupRe.MatchString(g) || len(ids) == 0 || len(ids) > 50 {
			return ErrBadQuery
		}
	}
	if len(f.Names) > 50 {
		return ErrBadQuery
	}
	return nil
}

func (f Filter) Location() *time.Location {
	if loc, err := time.LoadLocation(f.TZ); err == nil && f.TZ != "" {
		return loc
	}
	return time.UTC
}

type AggQuery struct {
	Filter
	Metric  string
	GroupBy []string
	Bucket  string
	Limit   int
}

func (q *AggQuery) Normalize() error {
	if err := q.Filter.Validate(); err != nil {
		return err
	}
	switch q.Metric {
	case "":
		q.Metric = MetricCount
	case MetricCount, MetricSum, MetricAvg, MetricMin, MetricMax, MetricUsers:
	default:
		return ErrBadQuery
	}
	switch q.Bucket {
	case "":
		q.Bucket = BucketNone
	case BucketNone, BucketHour, BucketDay, BucketWeek:
	default:
		return ErrBadQuery
	}
	if len(q.GroupBy) > 4 {
		return ErrBadQuery
	}
	seen := map[string]bool{}
	for _, g := range q.GroupBy {
		if !groupKey.MatchString(g) || seen[g] {
			return ErrBadQuery
		}
		seen[g] = true
	}
	if q.Limit <= 0 {
		q.Limit = 100
	}
	if q.Limit > 1000 {
		q.Limit = 1000
	}
	return nil
}

type AggRow struct {
	Bucket string   `json:"bucket,omitempty"`
	Keys   []string `json:"keys"`
	Value  float64  `json:"value"`
	Count  int64    `json:"count"`
	Users  int64    `json:"users"`
}

type ActivityDay struct {
	Day       string `json:"day"`
	DAU       int64  `json:"dau"`
	NewUsers  int64  `json:"newUsers"`
	Returning int64  `json:"returning"`
	Events    int64  `json:"events"`
}

type Activity struct {
	Days     []ActivityDay `json:"days"`
	DAU      int64         `json:"dau"`
	WAU      int64         `json:"wau"`
	MAU      int64         `json:"mau"`
	NewUsers int64         `json:"newUsers"`
	Users    int64         `json:"users"`
	Events   int64         `json:"events"`
}

type RetentionQuery struct {
	Filter
	Days  []int
	Mode  string
	Rows  int
	Today string
}

type RetentionSummary struct {
	Day      int      `json:"day"`
	Cohort   int64    `json:"cohort"`
	Returned int64    `json:"returned"`
	Rate     *float64 `json:"rate"`
}

type CohortCell struct {
	Day      int      `json:"day"`
	Returned *int64   `json:"returned"`
	Rate     *float64 `json:"rate"`
}

type Cohort struct {
	Day   string       `json:"day"`
	Size  int64        `json:"size"`
	Cells []CohortCell `json:"cells"`
}

type Retention struct {
	Summary []RetentionSummary `json:"summary"`
	Cohorts []Cohort           `json:"cohorts"`
}

type Step struct {
	Name string              `json:"name"`
	Tags map[string][]string `json:"tags,omitempty"`
}

type FunnelQuery struct {
	Filter
	Steps  []Step
	Window time.Duration
}

type FunnelStep struct {
	Index         int      `json:"index"`
	Name          string   `json:"name"`
	Users         int64    `json:"users"`
	MedianSeconds *float64 `json:"medianSeconds"`
}

type EventQuery struct {
	Filter
	Limit    int
	Asc      bool
	BeforeAt *time.Time
	BeforeID string
}

type ValuesQuery struct {
	Filter
	Kind  string
	Limit int
}

type ValueCount struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

type Store interface {
	Kind() string
	Ensure(ctx context.Context) error
	Write(ctx context.Context, events []Event) error
	Aggregate(ctx context.Context, q AggQuery) ([]AggRow, error)
	Activity(ctx context.Context, f Filter) (Activity, error)
	Retention(ctx context.Context, q RetentionQuery) (Retention, error)
	Funnel(ctx context.Context, q FunnelQuery) ([]FunnelStep, error)
	Events(ctx context.Context, q EventQuery) ([]Event, error)
	Values(ctx context.Context, q ValuesQuery) ([]ValueCount, error)
	Sweep(ctx context.Context, beforeDay string) (int64, error)
	Ping(ctx context.Context) error
	Close() error
}

func Rate(a, b int64) *float64 {
	if b <= 0 {
		return nil
	}
	v := float64(int64(float64(a)/float64(b)*10000+0.5)) / 10000
	return &v
}

func ValidateFunnel(q *FunnelQuery) error {
	if err := q.Filter.Validate(); err != nil {
		return err
	}
	if len(q.Steps) < 2 || len(q.Steps) > 12 {
		return ErrBadQuery
	}
	for _, s := range q.Steps {
		if strings.TrimSpace(s.Name) == "" {
			return ErrBadQuery
		}
		for g, ids := range s.Tags {
			if !tagGroupRe.MatchString(g) || len(ids) == 0 {
				return ErrBadQuery
			}
		}
	}
	if q.Window < time.Minute || q.Window > 31*24*time.Hour {
		return ErrBadQuery
	}
	return nil
}

func ValidateRetention(q *RetentionQuery) error {
	if err := q.Filter.Validate(); err != nil {
		return err
	}
	if len(q.Days) == 0 || len(q.Days) > 12 {
		return ErrBadQuery
	}
	for _, d := range q.Days {
		if d < 1 || d > 365 {
			return ErrBadQuery
		}
	}
	if q.Mode == "" {
		q.Mode = ModeClassic
	}
	if q.Mode != ModeClassic && q.Mode != ModeRolling {
		return ErrBadQuery
	}
	if q.Rows <= 0 || q.Rows > 366 {
		q.Rows = 30
	}
	if _, err := time.Parse(time.DateOnly, q.Today); err != nil {
		return ErrBadQuery
	}
	return nil
}

func AddDays(day string, n int) string {
	t, err := time.Parse(time.DateOnly, day)
	if err != nil {
		return day
	}
	return t.AddDate(0, 0, n).Format(time.DateOnly)
}

func Median(v []float64) *float64 {
	if len(v) == 0 {
		return nil
	}
	s := append([]float64{}, v...)
	sort.Float64s(s)
	var m float64
	if n := len(s); n%2 == 1 {
		m = s[n/2]
	} else {
		m = (s[n/2-1] + s[n/2]) / 2
	}
	return &m
}
