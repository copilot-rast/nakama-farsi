package eventcontext

import (
	"context"
	"testing"
	"time"
)

func policy(conflict string) Policy {
	return Policy{Enabled: true, Conflict: conflict, FlagMismatch: true, KeepClientOutside: true,
		Fills: []Fill{{Group: "mission", Source: "campaign.stage", Enabled: true}, {Group: "car", Source: "campaign.item", Enabled: true}, {Group: "chapter", Source: "campaign.chapter"}}}
}

func TestTimelinePicksTheOpenSpan(t *testing.T) {
	base := time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)
	Register(Provider{Name: "test", Spans: func(_ context.Context, userID string, from, to time.Time) ([]Span, error) {
		return []Span{
			{ID: "a", From: base, To: base.Add(10 * time.Minute), Values: map[string]string{"campaign.stage": "s1", "campaign.item": "tiggo8"}},
			{ID: "b", From: base.Add(20 * time.Minute), Values: map[string]string{"campaign.stage": "s2"}},
		}, nil
	}, Sources: func() []Source { return []Source{{Key: "campaign.stage"}} }})
	defer Unregister("test")
	if !Has() || len(Sources()) != 1 || Sources()[0].Owner != "test" {
		t.Fatalf("sources %+v", Sources())
	}
	tl, err := Load(context.Background(), "u", base, base.Add(time.Hour), 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if v := tl.At(base.Add(time.Minute)); v["campaign.stage"] != "s1" || v["campaign.item"] != "tiggo8" {
		t.Fatalf("inside a: %v", v)
	}
	if v := tl.At(base.Add(10*time.Minute + 4*time.Second)); v["campaign.stage"] != "s1" {
		t.Fatalf("grace: %v", v)
	}
	if v := tl.At(base.Add(15 * time.Minute)); v != nil {
		t.Fatalf("between: %v", v)
	}
	if v := tl.At(base.Add(5 * time.Hour)); v["campaign.stage"] != "s2" {
		t.Fatalf("still open b: %v", v)
	}
}

func TestApplyModes(t *testing.T) {
	in := map[string]string{"campaign.stage": "chapter-1-level-01", "campaign.item": "tiggo8", "campaign.chapter": "chapter-1"}
	tags := map[string]string{"weather": "rain"}
	o := Apply(policy(ConflictOverride), in, tags, nil)
	if tags["mission"] != "chapter-1-level-01" || tags["car"] != "tiggo8" || tags["chapter"] != "" || len(o.Filled) != 2 || o.Mismatched() {
		t.Fatalf("fill: %v %+v", tags, o)
	}
	tags = map[string]string{"mission": "3434343", "car": "Tiggo8"}
	o = Apply(policy(ConflictOverride), in, tags, nil)
	if tags["mission"] != "chapter-1-level-01" || tags["car"] != "tiggo8" || o.Mismatch["mission"] != "3434343" || len(o.Mismatch) != 1 {
		t.Fatalf("override: %v %+v", tags, o)
	}
	tags = map[string]string{"mission": "3434343"}
	o = Apply(policy(ConflictKeepClient), in, tags, nil)
	if tags["mission"] != "3434343" || tags["car"] != "tiggo8" || !o.Mismatched() {
		t.Fatalf("keep client: %v %+v", tags, o)
	}
	tags = map[string]string{"mission": "3434343"}
	if o = Apply(policy(ConflictReject), in, tags, nil); o.Reject != "mission" {
		t.Fatalf("reject: %+v", o)
	}
	tags = map[string]string{"car": "audiq2"}
	if o = Apply(policy(ConflictReject), nil, tags, nil); tags["car"] != "audiq2" || o.Inside {
		t.Fatalf("outside keep: %v %+v", tags, o)
	}
	p := policy(ConflictOverride)
	p.KeepClientOutside = false
	if o = Apply(p, nil, tags, nil); tags["car"] != "" || len(o.Dropped) != 1 {
		t.Fatalf("outside drop: %v %+v", tags, o)
	}
	tags = map[string]string{"car": "audiq2"}
	if o = Apply(policy(ConflictOverride), map[string]string{"campaign.stage": "s"}, tags, func(v string) bool { return v != "s" }); tags["car"] != "audiq2" || tags["mission"] != "" {
		t.Fatalf("no server item, invalid stage: %v %+v", tags, o)
	}
	p.Enabled = false
	tags = map[string]string{}
	if Apply(p, in, tags, nil); len(tags) != 0 {
		t.Fatalf("disabled: %v", tags)
	}
}

func TestValidateAndClamp(t *testing.T) {
	p := policy(ConflictOverride)
	if f, c := p.Validate("context"); c != "" {
		t.Fatal(f, c)
	}
	p.Conflict = "x"
	if f, _ := p.Validate("context"); f != "context.conflict" {
		t.Fatal(f)
	}
	p = policy(ConflictOverride)
	p.Fills = append(p.Fills, Fill{Group: "car", Source: "campaign.item"})
	if _, c := p.Validate("context"); c != "DUPLICATE_TAG_GROUP" {
		t.Fatal(c)
	}
	now := time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)
	if got := Clamp(now.Add(-100*time.Hour), now, 72*time.Hour, 5*time.Minute); !got.Equal(now.Add(-72 * time.Hour)) {
		t.Fatal(got)
	}
	if got := Clamp(now.Add(time.Hour), now, 72*time.Hour, 5*time.Minute); !got.Equal(now.Add(5 * time.Minute)) {
		t.Fatal(got)
	}
}
