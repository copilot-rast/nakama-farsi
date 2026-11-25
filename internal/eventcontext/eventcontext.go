package eventcontext

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	ConflictOverride   = "override"
	ConflictKeepClient = "keep_client"
	ConflictReject     = "reject"

	MaxFills = 16
)

var (
	GroupPattern  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,31}$`)
	SourcePattern = regexp.MustCompile(`^[a-z][A-Za-z0-9_.]{0,63}$`)
)

type Source struct {
	Key     string `json:"key"`
	Owner   string `json:"owner"`
	LabelFa string `json:"labelFa"`
	LabelEn string `json:"labelEn"`
}

type Span struct {
	ID     string
	From   time.Time
	To     time.Time
	Values map[string]string
}

func (s Span) Contains(at time.Time, grace time.Duration) bool {
	if at.Before(s.From) {
		return false
	}
	return s.To.IsZero() || !at.After(s.To.Add(grace))
}

type Provider struct {
	Name    string
	Sources func() []Source
	Spans   func(ctx context.Context, userID string, from, to time.Time) ([]Span, error)
}

var (
	mu        sync.RWMutex
	providers = map[string]Provider{}
)

func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	if p.Spans == nil {
		delete(providers, p.Name)
		return
	}
	providers[p.Name] = p
}

func Unregister(name string) {
	mu.Lock()
	defer mu.Unlock()
	delete(providers, name)
}

func Has() bool {
	mu.RLock()
	defer mu.RUnlock()
	return len(providers) > 0
}

func list() []Provider {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(providers))
	for n := range providers {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]Provider, 0, len(names))
	for _, n := range names {
		out = append(out, providers[n])
	}
	return out
}

func Sources() []Source {
	out := []Source{}
	for _, p := range list() {
		if p.Sources == nil {
			continue
		}
		for _, s := range p.Sources() {
			if s.Owner == "" {
				s.Owner = p.Name
			}
			out = append(out, s)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

type Timeline struct {
	Spans []Span
	Grace time.Duration
}

func Load(ctx context.Context, userID string, from, to time.Time, grace time.Duration) (Timeline, error) {
	t := Timeline{Grace: grace}
	if strings.TrimSpace(userID) == "" {
		return t, nil
	}
	for _, p := range list() {
		spans, err := p.Spans(ctx, userID, from.Add(-grace), to)
		if err != nil {
			return t, err
		}
		t.Spans = append(t.Spans, spans...)
	}
	sort.SliceStable(t.Spans, func(i, j int) bool { return t.Spans[i].From.Before(t.Spans[j].From) })
	return t, nil
}

func (t Timeline) At(at time.Time) map[string]string {
	var out map[string]string
	for _, s := range t.Spans {
		if !s.Contains(at, t.Grace) {
			continue
		}
		if out == nil {
			out = map[string]string{}
		}
		for k, v := range s.Values {
			if v != "" {
				out[k] = v
			}
		}
	}
	return out
}

type Fill struct {
	Group   string `json:"group"`
	Source  string `json:"source"`
	Enabled bool   `json:"enabled"`
}

type Policy struct {
	Enabled           bool   `json:"enabled"`
	Fills             []Fill `json:"fills"`
	Conflict          string `json:"conflict"`
	FlagMismatch      bool   `json:"flagMismatch"`
	KeepClientOutside bool   `json:"keepClientOutside"`
	GraceSeconds      int    `json:"graceSeconds"`
}

func (p Policy) Grace() time.Duration { return time.Duration(p.GraceSeconds) * time.Second }

func (p Policy) Active() bool {
	if !p.Enabled {
		return false
	}
	for _, f := range p.Fills {
		if f.Enabled {
			return true
		}
	}
	return false
}

func (p Policy) Validate(prefix string) (string, string) {
	switch p.Conflict {
	case ConflictOverride, ConflictKeepClient, ConflictReject:
	default:
		return prefix + ".conflict", "INVALID_POLICY"
	}
	if p.GraceSeconds < 0 || p.GraceSeconds > 3600 {
		return prefix + ".graceSeconds", "VALUE_OUT_OF_RANGE"
	}
	if len(p.Fills) > MaxFills {
		return prefix + ".fills", "VALUE_OUT_OF_RANGE"
	}
	seen := map[string]bool{}
	for _, f := range p.Fills {
		switch {
		case !GroupPattern.MatchString(f.Group):
			return prefix + ".fills.group", "INVALID_TAG_GROUP"
		case !SourcePattern.MatchString(f.Source):
			return prefix + ".fills.source", "INVALID_SOURCE"
		case seen[f.Group]:
			return prefix + ".fills.group", "DUPLICATE_TAG_GROUP"
		}
		seen[f.Group] = true
	}
	return "", ""
}

type Outcome struct {
	Filled   []string
	Mismatch map[string]string
	Server   map[string]string
	Dropped  []string
	Reject   string
	Inside   bool
}

func (o Outcome) Mismatched() bool { return len(o.Mismatch) > 0 }

func Apply(p Policy, values map[string]string, tags map[string]string, valid func(string) bool) Outcome {
	out := Outcome{Inside: values != nil}
	if !p.Enabled || tags == nil {
		return out
	}
	for _, f := range p.Fills {
		if !f.Enabled {
			continue
		}
		cv, has := tags[f.Group]
		if !out.Inside {
			if has && !p.KeepClientOutside {
				delete(tags, f.Group)
				out.Dropped = append(out.Dropped, f.Group)
			}
			continue
		}
		sv := strings.TrimSpace(values[f.Source])
		if sv == "" || valid != nil && !valid(sv) {
			continue
		}
		switch {
		case !has || strings.TrimSpace(cv) == "":
			tags[f.Group] = sv
			out.Filled = append(out.Filled, f.Group)
		case strings.EqualFold(strings.TrimSpace(cv), sv):
			tags[f.Group] = sv
		default:
			if out.Mismatch == nil {
				out.Mismatch, out.Server = map[string]string{}, map[string]string{}
			}
			out.Mismatch[f.Group], out.Server[f.Group] = cv, sv
			switch p.Conflict {
			case ConflictReject:
				out.Reject = f.Group
				return out
			case ConflictKeepClient:
			default:
				tags[f.Group] = sv
			}
		}
	}
	return out
}

func Clamp(at, now time.Time, past, future time.Duration) time.Time {
	if lo := now.Add(-past); past > 0 && at.Before(lo) {
		return lo
	}
	if hi := now.Add(future); at.After(hi) {
		return hi
	}
	return at
}
