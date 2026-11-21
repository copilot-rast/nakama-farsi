package events

import (
	"bytes"
	"encoding/json"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	StatusStored   = "stored"
	StatusFlagged  = "flagged"
	StatusRejected = "rejected"
	StatusDropped  = "dropped"

	PolicyAccept = "accept"
	PolicyFlag   = "flag"
	PolicyReject = "reject"

	MaxPathParts = 6
)

var (
	BatchIDPattern  = regexp.MustCompile(`^[A-Za-z0-9_.:-]{8,64}$`)
	GroupPattern    = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)
	GamePattern     = regexp.MustCompile(`^[a-z0-9_-]{0,32}$`)
	namePattern     = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]*$`)
	SessionPattern  = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,64}$`)
	typedPartRegexp = regexp.MustCompile(`^[A-Za-z0-9_.\- ]{1,64}$`)
)

type Tag struct {
	Group string `json:"tagGroup"`
	ID    string `json:"id"`
}

type Raw struct {
	Name     string          `json:"name,omitempty"`
	Path     string          `json:"path,omitempty"`
	Tags     json.RawMessage `json:"tags,omitempty"`
	Value    any             `json:"value,omitempty"`
	Time     any             `json:"time,omitempty"`
	Fields   map[string]any  `json:"fields,omitempty"`
	Type     string          `json:"type,omitempty"`
	ID       string          `json:"id,omitempty"`
	Status   string          `json:"status,omitempty"`
	P1       string          `json:"p1,omitempty"`
	P2       string          `json:"p2,omitempty"`
	P3       string          `json:"p3,omitempty"`
	Score    any             `json:"score,omitempty"`
	Flow     string          `json:"flow,omitempty"`
	Currency string          `json:"currency,omitempty"`
	Amount   any             `json:"amount,omitempty"`
	ItemType string          `json:"itemType,omitempty"`
	Item     string          `json:"item,omitempty"`
	CartType string          `json:"cartType,omitempty"`
	Store    string          `json:"store,omitempty"`
	Severity string          `json:"severity,omitempty"`
	Message  string          `json:"message,omitempty"`
}

type Group struct {
	Key     string          `json:"key"`
	Closed  bool            `json:"closed"`
	Enabled bool            `json:"enabled"`
	Values  map[string]bool `json:"-"`
}

type ValueRule struct {
	Required bool     `json:"required"`
	Min      *float64 `json:"min"`
	Max      *float64 `json:"max"`
	Unit     string   `json:"unit"`
}

type Def struct {
	Name         string    `json:"name"`
	Enabled      bool      `json:"enabled"`
	PathGroups   []string  `json:"pathGroups"`
	RequiredTags []string  `json:"requiredTags"`
	AllowedTags  []string  `json:"allowedTags"`
	Value        ValueRule `json:"value"`
}

type Catalog struct {
	Groups map[string]Group
	Events map[string]Def
}

func (c *Catalog) group(key string) (Group, bool) {
	if c == nil || c.Groups == nil {
		return Group{}, false
	}
	g, ok := c.Groups[key]
	return g, ok
}

func (c *Catalog) def(name string) (Def, bool) {
	if c == nil || c.Events == nil {
		return Def{}, false
	}
	d, ok := c.Events[name]
	return d, ok
}

type Policy struct {
	MaxTags          int
	MaxNameLength    int
	MaxFields        int
	MaxStringLength  int
	ValueMin         float64
	ValueMax         float64
	PastHours        int
	FutureMinutes    int
	UnknownEvents    string
	UnknownTagGroups string
	UnknownTagValues string
	ValueOutOfRange  string
}

func DefaultPolicy() Policy {
	return Policy{MaxTags: 8, MaxNameLength: 64, MaxFields: 16, MaxStringLength: 256, ValueMin: -1e12, ValueMax: 1e12, PastHours: 72, FutureMinutes: 5,
		UnknownEvents: PolicyAccept, UnknownTagGroups: PolicyAccept, UnknownTagValues: PolicyFlag, ValueOutOfRange: PolicyReject}
}

type Prepared struct {
	Name       string
	Tags       map[string]string
	Value      *float64
	Fields     map[string]any
	ClientTime *time.Time
	Flagged    bool
	Reasons    []string
}

func (p *Prepared) flag(reason string) {
	p.Flagged = true
	for _, r := range p.Reasons {
		if r == reason {
			return
		}
	}
	p.Reasons = append(p.Reasons, reason)
}

func (p Prepared) Reason() string { return strings.Join(p.Reasons, ",") }

type Result struct {
	Index    int      `json:"index"`
	Name     string   `json:"name,omitempty"`
	Status   string   `json:"status"`
	Code     string   `json:"code,omitempty"`
	Problems []string `json:"problems,omitempty"`
}

func ValidName(name string, max int) bool {
	return len(name) <= max && namePattern.MatchString(name)
}

func ValidTagID(id string) bool {
	if id == "" || utf8.RuneCountInString(id) > 64 || strings.Contains(id, ":") {
		return false
	}
	for _, r := range id {
		if unicode.IsControl(r) {
			return false
		}
	}
	return strings.TrimSpace(id) == id
}

func ParseTags(raw json.RawMessage) ([]Tag, string) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return nil, ""
	}
	if raw[0] == '{' {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, "ANALYTICS_INVALID_TAG"
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make([]Tag, 0, len(m))
		for _, k := range keys {
			id, ok := tagString(m[k])
			if !ok {
				return nil, "ANALYTICS_INVALID_TAG"
			}
			out = append(out, Tag{Group: k, ID: id})
		}
		return out, ""
	}
	var list []map[string]any
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, "ANALYTICS_INVALID_TAG"
	}
	out := make([]Tag, 0, len(list))
	for _, item := range list {
		g, _ := item["tagGroup"].(string)
		if g == "" {
			g, _ = item["group"].(string)
		}
		id, ok := tagString(item["id"])
		if !ok {
			return nil, "ANALYTICS_INVALID_TAG"
		}
		out = append(out, Tag{Group: g, ID: id})
	}
	return out, ""
}

func tagString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case float64:
		if x == math.Trunc(x) && math.Abs(x) < 1e15 {
			return strconv.FormatInt(int64(x), 10), true
		}
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case json.Number:
		return x.String(), true
	case bool:
		return strconv.FormatBool(x), true
	}
	return "", false
}

func Number(v any) (*float64, bool) {
	switch x := v.(type) {
	case nil:
		return nil, true
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return nil, false
		}
		return &x, true
	case int:
		f := float64(x)
		return &f, true
	case int64:
		f := float64(x)
		return &f, true
	case json.Number:
		f, err := x.Float64()
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, false
		}
		return &f, true
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return nil, true
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, false
		}
		return &f, true
	}
	return nil, false
}

func ParseTime(v any) (time.Time, bool) {
	switch x := v.(type) {
	case float64:
		return unixAny(x)
	case int64:
		return unixAny(float64(x))
	case json.Number:
		f, err := x.Float64()
		if err != nil {
			return time.Time{}, false
		}
		return unixAny(f)
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return time.Time{}, false
		}
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return t, true
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return unixAny(f)
		}
	}
	return time.Time{}, false
}

func unixAny(f float64) (time.Time, bool) {
	if math.IsNaN(f) || math.IsInf(f, 0) || f <= 0 {
		return time.Time{}, false
	}
	switch {
	case f > 1e17:
		return time.Unix(0, int64(f)*1000).UTC(), true
	case f > 1e14:
		return time.UnixMicro(int64(f)).UTC(), true
	case f > 1e11:
		return time.UnixMilli(int64(f)).UTC(), true
	}
	sec, frac := math.Modf(f)
	return time.Unix(int64(sec), int64(frac*1e9)).UTC(), true
}

func cleanFields(in map[string]any, maxFields, maxString int) (map[string]any, bool) {
	if len(in) == 0 {
		return nil, true
	}
	if len(in) > maxFields {
		return nil, false
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		if !GroupPattern.MatchString(strings.ToLower(k)) && !namePattern.MatchString(k) || len(k) > 64 {
			return nil, false
		}
		switch x := v.(type) {
		case nil:
			continue
		case string:
			if utf8.RuneCountInString(x) > maxString {
				return nil, false
			}
			out[k] = x
		case float64:
			if math.IsNaN(x) || math.IsInf(x, 0) {
				return nil, false
			}
			out[k] = x
		case bool:
			out[k] = x
		case json.Number:
			f, err := x.Float64()
			if err != nil {
				return nil, false
			}
			out[k] = f
		default:
			return nil, false
		}
	}
	if len(out) == 0 {
		return nil, true
	}
	return out, true
}

func typedPart(v string) (string, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", true
	}
	return v, typedPartRegexp.MatchString(v)
}

func typed(ev Raw) (name, path string, tags []Tag, value *float64, fields map[string]any, code string) {
	typ := strings.ToLower(strings.TrimSpace(ev.Type))
	add := func(group, v string) bool {
		p, ok := typedPart(v)
		if !ok {
			return false
		}
		if p != "" {
			tags = append(tags, Tag{Group: group, ID: p})
		}
		return true
	}
	bad := "ANALYTICS_INVALID_TYPED_EVENT"
	switch typ {
	case "design":
		if strings.TrimSpace(ev.ID) == "" {
			return "", "", nil, nil, nil, bad
		}
		v, ok := Number(ev.Value)
		if !ok {
			return "", "", nil, nil, nil, "ANALYTICS_INVALID_VALUE"
		}
		return "", strings.TrimSpace(ev.ID), nil, v, nil, ""
	case "progression":
		st := strings.ToLower(strings.TrimSpace(ev.Status))
		if st != "start" && st != "complete" && st != "fail" {
			return "", "", nil, nil, nil, bad
		}
		if strings.TrimSpace(ev.P1) == "" || !add("status", st) || !add("progression1", ev.P1) || !add("progression2", ev.P2) || !add("progression3", ev.P3) {
			return "", "", nil, nil, nil, bad
		}
		v, ok := Number(ev.Score)
		if !ok {
			return "", "", nil, nil, nil, "ANALYTICS_INVALID_VALUE"
		}
		return "progression", "", tags, v, nil, ""
	case "resource":
		flow := strings.ToLower(strings.TrimSpace(ev.Flow))
		if flow != "source" && flow != "sink" {
			return "", "", nil, nil, nil, bad
		}
		v, ok := Number(ev.Amount)
		if !ok || v == nil || strings.TrimSpace(ev.Currency) == "" {
			return "", "", nil, nil, nil, bad
		}
		a := math.Abs(*v)
		if flow == "sink" {
			a = -a
		}
		if !add("flow", flow) || !add("currency", ev.Currency) || !add("item_type", ev.ItemType) || !add("item", ev.Item) {
			return "", "", nil, nil, nil, bad
		}
		return "resource", "", tags, &a, nil, ""
	case "business":
		v, ok := Number(ev.Amount)
		if !ok || v == nil || strings.TrimSpace(ev.Currency) == "" {
			return "", "", nil, nil, nil, bad
		}
		if !add("currency", ev.Currency) || !add("item_type", ev.ItemType) || !add("item", ev.Item) || !add("cart_type", ev.CartType) || !add("store", ev.Store) {
			return "", "", nil, nil, nil, bad
		}
		return "business", "", tags, v, nil, ""
	case "error":
		sev := strings.ToLower(strings.TrimSpace(ev.Severity))
		switch sev {
		case "debug", "info", "warning", "error", "critical":
		default:
			return "", "", nil, nil, nil, bad
		}
		add("severity", sev)
		msg := ev.Message
		if r := []rune(msg); len(r) > 1024 {
			msg = string(r[:1024])
		}
		if msg != "" {
			fields = map[string]any{"message": msg}
		}
		return "error", "", tags, nil, fields, ""
	}
	return "", "", nil, nil, nil, bad
}

func SplitPath(path string, cat *Catalog) (string, []Tag, bool, bool) {
	parts := strings.Split(path, ":")
	if len(parts) > MaxPathParts {
		return "", nil, false, false
	}
	name := strings.TrimSpace(parts[0])
	def, known := cat.def(name)
	tags := []Tag{}
	extra := false
	slot := 0
	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		if p == "" {
			slot++
			continue
		}
		group := ""
		if known && slot < len(def.PathGroups) {
			group = def.PathGroups[slot]
		} else {
			group = "p" + strconv.Itoa(slot+1)
			extra = true
		}
		tags = append(tags, Tag{Group: group, ID: p})
		slot++
	}
	return name, tags, extra, true
}

func Prepare(i int, ev Raw, cat *Catalog, pol Policy, now time.Time) (Prepared, Result) {
	res := Result{Index: i, Status: StatusStored}
	reject := func(code string, problems ...string) (Prepared, Result) {
		res.Status, res.Code = StatusRejected, code
		if len(problems) > 0 {
			res.Problems = problems
		}
		return Prepared{}, res
	}
	var p Prepared
	name := strings.TrimSpace(ev.Name)
	path := strings.TrimSpace(ev.Path)
	var tags []Tag
	var value *float64
	var fields map[string]any
	if strings.TrimSpace(ev.Type) != "" {
		n, pth, tg, v, f, code := typed(ev)
		if code != "" {
			return reject(code)
		}
		name, path, tags, value, fields = n, pth, tg, v, f
	} else {
		v, ok := Number(ev.Value)
		if !ok {
			return reject("ANALYTICS_INVALID_VALUE")
		}
		value = v
		list, code := ParseTags(ev.Tags)
		if code != "" {
			return reject(code)
		}
		tags = list
	}
	if path != "" {
		if name != "" && name != strings.Split(path, ":")[0] {
			return reject("ANALYTICS_INVALID_EVENT_NAME")
		}
		n, pt, extra, ok := SplitPath(path, cat)
		if !ok {
			return reject("ANALYTICS_TOO_MANY_TAGS")
		}
		name = n
		if extra {
			p.flag("path_groups")
		}
		tags = append(pt, tags...)
	}
	res.Name = name
	if !ValidName(name, pol.MaxNameLength) {
		return reject("ANALYTICS_INVALID_EVENT_NAME")
	}
	if len(tags) > pol.MaxTags {
		return reject("ANALYTICS_TOO_MANY_TAGS")
	}
	p.Name = name
	p.Tags = make(map[string]string, len(tags))
	for _, t := range tags {
		g := strings.TrimSpace(t.Group)
		if !GroupPattern.MatchString(g) || !ValidTagID(t.ID) {
			return reject("ANALYTICS_INVALID_TAG", g)
		}
		if _, dup := p.Tags[g]; dup {
			return reject("ANALYTICS_DUPLICATE_TAG_GROUP", g)
		}
		p.Tags[g] = t.ID
	}
	extraFields, ok := cleanFields(ev.Fields, pol.MaxFields, pol.MaxStringLength)
	if !ok {
		return reject("ANALYTICS_INVALID_FIELDS")
	}
	for k, v := range extraFields {
		if fields == nil {
			fields = map[string]any{}
		}
		fields[k] = v
	}
	p.Fields = fields
	p.Value = value
	if t, ok := ParseTime(ev.Time); ok {
		if !t.Before(now.Add(-time.Duration(pol.PastHours)*time.Hour)) && !t.After(now.Add(time.Duration(pol.FutureMinutes)*time.Minute)) {
			ct := t.UTC()
			p.ClientTime = &ct
		} else {
			p.flag("client_time")
		}
	}
	def, known := cat.def(name)
	switch {
	case !known:
		switch pol.UnknownEvents {
		case PolicyReject:
			return reject("ANALYTICS_UNKNOWN_EVENT")
		case PolicyFlag:
			p.flag("unknown_event")
		}
	case !def.Enabled:
		res.Status, res.Code = StatusDropped, "ANALYTICS_EVENT_DISABLED"
		return Prepared{}, res
	}
	if known {
		for _, g := range def.RequiredTags {
			if _, ok := p.Tags[g]; !ok {
				return reject("ANALYTICS_MISSING_TAG", g)
			}
		}
		if len(def.AllowedTags) > 0 {
			allowed := map[string]bool{}
			for _, g := range def.AllowedTags {
				allowed[g] = true
			}
			for _, g := range def.RequiredTags {
				allowed[g] = true
			}
			for _, g := range def.PathGroups {
				allowed[g] = true
			}
			for g := range p.Tags {
				if !allowed[g] {
					return reject("ANALYTICS_UNKNOWN_TAG_GROUP", g)
				}
			}
		}
		if def.Value.Required && p.Value == nil {
			return reject("ANALYTICS_VALUE_REQUIRED")
		}
	}
	lo, hi := pol.ValueMin, pol.ValueMax
	if known && def.Value.Min != nil {
		lo = *def.Value.Min
	}
	if known && def.Value.Max != nil {
		hi = *def.Value.Max
	}
	if p.Value != nil && (*p.Value < lo || *p.Value > hi) {
		if pol.ValueOutOfRange == PolicyFlag {
			p.flag("value_range")
		} else {
			return reject("ANALYTICS_VALUE_OUT_OF_RANGE")
		}
	}
	groups := make([]string, 0, len(p.Tags))
	for g := range p.Tags {
		groups = append(groups, g)
	}
	sort.Strings(groups)
	for _, g := range groups {
		grp, ok := cat.group(g)
		switch {
		case !ok || !grp.Enabled:
			switch pol.UnknownTagGroups {
			case PolicyReject:
				return reject("ANALYTICS_UNKNOWN_TAG_GROUP", g)
			case PolicyFlag:
				p.flag("unknown_tag_group")
			}
		case grp.Closed && !grp.Values[p.Tags[g]]:
			if pol.UnknownTagValues == PolicyReject {
				return reject("ANALYTICS_UNKNOWN_TAG_VALUE", g)
			}
			p.flag("unknown_tag_value")
		}
	}
	if p.Flagged {
		res.Status = StatusFlagged
		switch p.Reasons[0] {
		case "client_time":
			res.Code = "ANALYTICS_CLIENT_TIME_OUT_OF_RANGE"
		case "unknown_event":
			res.Code = "ANALYTICS_UNKNOWN_EVENT"
		case "unknown_tag_group":
			res.Code = "ANALYTICS_UNKNOWN_TAG_GROUP"
		case "unknown_tag_value":
			res.Code = "ANALYTICS_UNKNOWN_TAG_VALUE"
		case "value_range":
			res.Code = "ANALYTICS_VALUE_OUT_OF_RANGE"
		default:
			res.Code = "ANALYTICS_PATH_GROUPS"
		}
		res.Problems = append([]string{}, p.Reasons...)
	}
	return p, res
}
