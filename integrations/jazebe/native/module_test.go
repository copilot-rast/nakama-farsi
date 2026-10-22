package main

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNakamaIdentityIsRequired(t *testing.T) {
	if _, err := userID(context.Background()); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("missing session must be denied: %v", err)
	}
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, "not-a-uuid")
	if _, err := userID(ctx); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("invalid identity must be denied: %v", err)
	}
	id := uuid.Must(uuid.NewV4()).String()
	ctx = context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, id)
	if got, err := userID(ctx); err != nil || got != id {
		t.Fatalf("valid Nakama identity rejected: %s, %v", got, err)
	}
}

func TestRequestValidation(t *testing.T) {
	var input struct {
		Name string `json:"name"`
	}
	for _, payload := range []string{`{"name":"a","unexpected":1}`, `{"name":`, `{} {}`, strings.Repeat("a", 8193)} {
		if err := decode(payload, &input); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("payload %q accepted: %v", payload, err)
		}
	}
	if _, err := requiredName(" "); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("empty name accepted: %v", err)
	}
	if _, err := requiredName(strings.Repeat("آ", 121)); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("oversized Persian name accepted: %v", err)
	}
	if name, err := requiredName("  سیاره آبی  "); err != nil || name != "سیاره آبی" {
		t.Fatalf("valid Persian name rejected: %q, %v", name, err)
	}
	if _, err := requiredUUID("other-user"); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid UUID accepted: %v", err)
	}
}

func TestPlanetDraftValidation(t *testing.T) {
	for _, value := range []string{"", "a", strings.Repeat("س", 129)} {
		if _, err := planetName(value); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("invalid planet name %q accepted: %v", value, err)
		}
	}
	if got, err := planetName("  سیاره آبی  "); err != nil || got != "سیاره آبی" {
		t.Fatalf("valid planet name rejected: %q, %v", got, err)
	}
	for _, value := range []string{"", "-bad", "bad-", "white space", strings.Repeat("a", 65)} {
		if _, err := validatedSlug(value); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("invalid slug %q accepted: %v", value, err)
		}
	}
	if got, err := validatedSlug("  Blue-Planet  "); err != nil || got != "blue-planet" {
		t.Fatalf("valid slug rejected: %q, %v", got, err)
	}
}

func TestBrandCreationRequiresConfiguredOperator(t *testing.T) {
	id := uuid.Must(uuid.NewV4()).String()
	if isOperator(id, "") || isOperator(id, "not-a-user") {
		t.Fatal("operator role must be disabled without an exact allowlist match")
	}
	if !isOperator(id, "not-a-user, "+strings.ToUpper(id)+" ") {
		t.Fatal("configured operator was denied")
	}
}

func TestPlanetIdentityUpdateValidation(t *testing.T) {
	for _, tc := range []struct {
		field, value string
	}{
		{"brandColor", `"red"`},
		{"deploymentMode", `"remote"`},
		{"pointName", `""`},
		{"name", `null`},
		{"slug", `"-invalid"`},
		{"categoryId", `"not-a-uuid"`},
	} {
		if _, err := planetUpdateValue(tc.field, []byte(tc.value)); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("invalid %s accepted: %v", tc.field, err)
		}
	}
	if got, err := planetUpdateValue("brandColor", []byte(`"#ABCDEF"`)); err != nil || got != "#ABCDEF" {
		t.Fatalf("valid color rejected: %v, %v", got, err)
	}
}

func TestPlanetPublishRequirements(t *testing.T) {
	missing := publishMissing(publishFields{}, false)
	if len(missing) != 7 || missing[6] != "active_mission" {
		t.Fatalf("unexpected missing fields: %v", missing)
	}
	ready := publishFields{
		name: "Planet", shortDescription: "Description", pointName: "Points",
		categoryID: sql.NullString{String: "category", Valid: true},
		brandColor: sql.NullString{String: "#AABBCC", Valid: true},
		glyph:      sql.NullString{String: "P", Valid: true},
	}
	if missing := publishMissing(ready, true); len(missing) != 0 {
		t.Fatalf("ready planet rejected: %v", missing)
	}
}

func TestPlanetOperatorActionDeniedByDefault(t *testing.T) {
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", "")
	id := uuid.Must(uuid.NewV4()).String()
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, id)
	if _, err := planetReject(ctx, nil, nil, nil, `{"planetId":"`+id+`"}`); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unconfigured operator changed planet: %v", err)
	}
	if _, err := planetApprove(ctx, nil, nil, nil, `{"planetId":"`+id+`"}`); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unconfigured operator approved planet: %v", err)
	}
}

func TestMediaMissionDraftValidation(t *testing.T) {
	var draft mediaMissionDraft
	draft.Title = "Watch the clip"
	draft.DetectionMethod = "platform_auto"
	draft.StartsAt = "2026-10-22T00:00:00Z"
	draft.EndsAt = "2026-10-23T00:00:00Z"
	draft.Media.AssetURL = "https://example.test/clip.mp4"
	draft.Media.TotalSec = 60
	draft.Media.MinWatchSec = 30
	if _, _, err := validateMediaMissionDraft(draft); err != nil {
		t.Fatalf("valid media mission rejected: %v", err)
	}
	draft.Media.MinWatchSec = 61
	if _, _, err := validateMediaMissionDraft(draft); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("watch time longer than clip accepted: %v", err)
	}
	draft.Media.MinWatchSec = 30
	draft.TargetType = "group"
	draft.GroupTargetValue = 2
	draft.RewardTickets = 1
	draft.TargetLotteryID = uuid.Must(uuid.NewV4()).String()
	if _, _, err := validateMediaMissionDraft(draft); err != nil {
		t.Fatalf("valid group media mission rejected: %v", err)
	}
	draft.RewardPoints = 1
	if _, _, err := validateMediaMissionDraft(draft); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("group points accepted: %v", err)
	}
	draft.RewardPoints = 0
	draft.TargetLotteryID = ""
	if _, _, err := validateMediaMissionDraft(draft); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("ticket reward without lottery accepted: %v", err)
	}
	draft.TargetLotteryID = uuid.Must(uuid.NewV4()).String()
	draft.EndsAt = draft.StartsAt
	if _, _, err := validateMediaMissionDraft(draft); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid mission period accepted: %v", err)
	}
}

func TestBrandRoleActions(t *testing.T) {
	for _, role := range []string{"technical", "marketing"} {
		for _, action := range brandActions(role) {
			if action == "team" {
				t.Fatalf("%s administrator may manage the team", role)
			}
		}
	}
	if len(brandActions("main")) != 8 || len(brandActions("unknown")) != 0 {
		t.Fatal("brand role actions do not match the source permission map")
	}
}

func TestMediaMissionActionGuards(t *testing.T) {
	refID := uuid.Must(uuid.NewV4()).String()
	if _, err := readMissionRef(`{"planetId":"` + refID + `","missionId":"invalid"}`); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid mission reference accepted: %v", err)
	}
	m := mediaMission{Status: "active", PlanetStatus: "active", StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(time.Hour)}
	if err := mediaOpen(m, time.Now()); err != nil {
		t.Fatalf("open media mission rejected: %v", err)
	}
	m.PlanetStatus = "setup"
	if err := mediaOpen(m, time.Now()); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("setup planet mission accepted: %v", err)
	}
	for _, config := range []string{`{}`, `{"assetUrl":"x","totalSec":5,"minWatchSec":6}`} {
		if _, err := mediaWatchSeconds([]byte(config)); err == nil {
			t.Fatalf("invalid config accepted: %s", config)
		}
	}
	if need, err := mediaWatchSeconds([]byte(`{"assetUrl":"x","totalSec":5,"minWatchSec":3}`)); err != nil || need != 3 {
		t.Fatalf("valid media config rejected: %d %v", need, err)
	}
}

func TestMissionReviewValidationAndLevels(t *testing.T) {
	id := uuid.Must(uuid.NewV4()).String()
	_, err := validateMissionReview(missionReviewInput{PlanetID: id, CompletionID: id, Decision: "reject"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("reasonless rejection accepted: %v", err)
	}
	_, err = validateMissionReview(missionReviewInput{PlanetID: id, CompletionID: id, Decision: "approve"})
	if err != nil {
		t.Fatalf("approval rejected: %v", err)
	}
	ladder := []byte(`[{"level":3,"key":"veteran","minXp":100},{"level":1,"key":"newcomer","minXp":0},{"level":2,"key":"active","minXp":50}]`)
	if level, key := levelAt(75, ladder, 1); level != 2 || key != "active" {
		t.Fatalf("level threshold mismatch: %d %s", level, key)
	}
	if level, _ := levelAt(10, ladder, 3); level != 3 {
		t.Fatalf("level demotion: %d", level)
	}
}

func TestCodeMissionRules(t *testing.T) {
	var draft codeMissionDraft
	draft.Title = "Enter a code"
	draft.DetectionMethod = "platform_auto"
	draft.Code.CodeMode = "shared"
	draft.Code.MaxUses = 2
	draft.StartsAt = "2026-10-22T00:00:00Z"
	draft.EndsAt = "2026-10-23T00:00:00Z"
	if _, _, err := validateCodeDraft(draft); err != nil {
		t.Fatalf("valid code mission rejected: %v", err)
	}
	draft.Code.CodeMode = "unique"
	if _, _, err := validateCodeDraft(draft); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("unique code with multiple uses accepted: %v", err)
	}
	code, err := renderMissionCode("JZ-####")
	if err != nil || len(code) != 7 || !strings.HasPrefix(code, "JZ-") {
		t.Fatalf("generated code invalid: %q %v", code, err)
	}
	for _, r := range code[3:] {
		if !strings.ContainsRune(missionCodeAlphabet, r) {
			t.Fatalf("code contains ambiguous character: %q", code)
		}
	}
}

func TestWalletAdjustmentGuards(t *testing.T) {
	id := uuid.Must(uuid.NewV4()).String()
	input := walletAdjustmentInput{PlanetID: id, UserID: id, Direction: "debit", Amount: 100, Reason: "Support correction", IdempotencyKey: "case-1234"}
	if _, err := validateWalletAdjustment(input); err != nil {
		t.Fatalf("valid adjustment rejected: %v", err)
	}
	input.Reason = "x"
	if _, err := validateWalletAdjustment(input); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("short reason accepted: %v", err)
	}
	input.Reason = "Support correction"
	input.IdempotencyKey = "short"
	if _, err := validateWalletAdjustment(input); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("short key accepted: %v", err)
	}
	a := adjustmentSourceID("adjust:" + id + ":case-1234")
	if a != adjustmentSourceID("adjust:"+id+":case-1234") || a == adjustmentSourceID("adjust:"+id+":case-5678") {
		t.Fatal("adjustment source ID is not stable or scoped")
	}
	if _, err := uuid.FromString(a); err != nil {
		t.Fatalf("adjustment source ID is not UUID: %s", a)
	}
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", "")
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, id)
	if _, err := walletAdjust(ctx, nil, nil, nil, `{}`); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unconfigured wallet adjustment allowed: %v", err)
	}
}

func TestContentEvidenceRules(t *testing.T) {
	var draft contentMissionDraft
	draft.Title = "Write a review"
	draft.StartsAt = "2026-10-22T00:00:00Z"
	draft.EndsAt = "2026-10-23T00:00:00Z"
	draft.Content.EvidenceTypes = []string{"text"}
	if _, _, err := validateContentDraft(draft); err != nil {
		t.Fatalf("text evidence draft rejected: %v", err)
	}
	draft.Content.EvidenceTypes = []string{"img"}
	if _, _, err := validateContentDraft(draft); err != nil {
		t.Fatalf("image evidence draft rejected: %v", err)
	}
	draft.Content.EvidenceTypes = []string{"video"}
	if _, _, err := validateContentDraft(draft); err != nil {
		t.Fatalf("video evidence draft rejected: %v", err)
	}
	draft.Content.EvidenceTypes = []string{"binary"}
	if _, _, err := validateContentDraft(draft); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("unsupported evidence type accepted: %v", err)
	}
	if _, err := contentEvidenceTypes([]byte(`{"evidenceTypes":["text","code"]}`)); err != nil {
		t.Fatalf("text/code config rejected: %v", err)
	}
}

func TestMissionOperatorModerationGuards(t *testing.T) {
	id := uuid.Must(uuid.NewV4()).String()
	if _, err := validateAdminMissionInput(adminMissionInput{MissionID: id, Reason: "policy violation"}, true); err != nil {
		t.Fatalf("valid suspension rejected: %v", err)
	}
	if _, err := validateAdminMissionInput(adminMissionInput{MissionID: id}, true); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("suspension without reason accepted: %v", err)
	}
	if _, err := validateAdminMissionInput(adminMissionInput{MissionID: id}, false); err != nil {
		t.Fatalf("unsuspension without reason rejected: %v", err)
	}
	if _, err := validateAdminMissionInput(adminMissionInput{MissionID: id, Reason: strings.Repeat("a", 513)}, false); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("oversized reason accepted: %v", err)
	}
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", "")
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, id)
	if _, err := missionOperator(ctx); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unconfigured operator allowed: %v", err)
	}
}

func TestMissionDisplayConditions(t *testing.T) {
	now := time.Date(2026, 10, 22, 20, 0, 0, 0, time.UTC) // 23:30 in Tehran.
	weekday, minute, err := tehranClock(now)
	if err != nil || minute != 23*60+30 {
		t.Fatalf("Tehran clock mismatch: weekday=%d minute=%d err=%v", weekday, minute, err)
	}
	viewer := missionViewer{Level: 2, Completed: map[string]bool{"prereq": true}, Now: now}
	condition := `{"minLevel":2,"prerequisiteMissionIds":["prereq"],"time":{"from":"22:00","to":"02:00"}}`
	if ok, err := missionDisplaySatisfied([]byte(condition), viewer); err != nil || !ok {
		t.Fatalf("eligible mission hidden: %v", err)
	}
	viewer.Level = 1
	if ok, err := missionDisplaySatisfied([]byte(condition), viewer); err != nil || ok {
		t.Fatalf("level-gated mission exposed: %v", err)
	}
	viewer.Level = 2
	viewer.Completed = map[string]bool{}
	if ok, err := missionDisplaySatisfied([]byte(condition), viewer); err != nil || ok {
		t.Fatalf("prerequisite-gated mission exposed: %v", err)
	}
	viewer.Completed["prereq"] = true
	viewer.Now = time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	if ok, err := missionDisplaySatisfied([]byte(condition), viewer); err != nil || ok {
		t.Fatalf("time-gated mission exposed: %v", err)
	}
	if ok, err := missionDisplaySatisfied([]byte(`{"weekdays":[7]}`), viewer); err != nil || ok {
		t.Fatalf("weekday-gated mission exposed: %v", err)
	}
	if _, err := missionDisplaySatisfied([]byte(`{"time":{"from":"8am"}}`), viewer); err == nil {
		t.Fatal("invalid time condition accepted")
	}
	cond := &missionDisplayConditions{MinLevel: 3, Weekdays: []int{0, 6}}
	cond.Time.From, cond.Time.To = "22:00", "02:00"
	if encoded, err := validateMissionDisplayConditions(cond); err != nil || encoded == nil {
		t.Fatalf("valid conditions rejected: %v", err)
	}
	cond.Weekdays = []int{0, 0}
	if _, err := validateMissionDisplayConditions(cond); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("duplicate weekdays accepted: %v", err)
	}
	cond.Weekdays = nil
	cond.PrerequisiteMissionIDs = []string{"invalid"}
	if _, err := validateMissionDisplayConditions(cond); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid prerequisite accepted: %v", err)
	}
}

func TestCheapestSoloPointFallback(t *testing.T) {
	mission := func(kind string, points, tickets int) map[string]any {
		return map[string]any{"type": kind, "reward": map[string]any{"points": points, "tickets": tickets}}
	}
	var got map[string]any
	got = cheaperFallbackMission(got, mission("group", 1, 0))
	got = cheaperFallbackMission(got, mission("solo", 0, 0))
	got = cheaperFallbackMission(got, mission("solo", 2, 1))
	if got != nil {
		t.Fatal("ineligible fallback selected")
	}
	got = cheaperFallbackMission(got, mission("solo", 10, 0))
	got = cheaperFallbackMission(got, mission("solo", 3, 0))
	got = cheaperFallbackMission(got, mission("solo", 8, 0))
	if got["reward"].(map[string]any)["points"] != 3 {
		t.Fatalf("wrong fallback: %#v", got)
	}
}

func TestMissionCatalogRejectsInvalidFilters(t *testing.T) {
	id := uuid.Must(uuid.NewV4()).String()
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, id)
	for _, payload := range []string{
		`{"planetId":"` + id + `","repeat":"hourly"}`,
		`{"planetId":"` + id + `","type":"operator"}`,
		`{"planetId":"` + id + `","sort":"random"}`,
	} {
		if _, err := missionList(ctx, nil, nil, nil, payload); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("invalid catalog filter accepted: %s, %v", payload, err)
		}
	}
}

func TestTehranMissionPeriodKeys(t *testing.T) {
	before := time.Date(2026, 10, 22, 20, 29, 0, 0, time.UTC)
	after := before.Add(time.Minute)
	oldKey, err := missionPeriodKey("periodic", "daily", before)
	if err != nil || oldKey != "2026-10-22" {
		t.Fatalf("wrong pre-midnight key: %s %v", oldKey, err)
	}
	newKey, err := missionPeriodKey("periodic", "daily", after)
	if err != nil || newKey != "2026-10-23" {
		t.Fatalf("wrong post-midnight key: %s %v", newKey, err)
	}
	if key, err := missionPeriodKey("periodic", "monthly", after); err != nil || key != "2026-10" {
		t.Fatalf("wrong monthly key: %s %v", key, err)
	}
	if key, err := missionPeriodKey("once", "", after); err != nil || key != "" {
		t.Fatalf("wrong once key: %s %v", key, err)
	}
	if _, err := missionPeriodKey("periodic", "invalid", after); err == nil {
		t.Fatal("invalid repeat period accepted")
	}
	var draft mediaMissionDraft
	draft.Title = "Watch"
	draft.DetectionMethod = "platform_auto"
	draft.Media.AssetURL = "https://example.test/video.mp4"
	draft.Media.TotalSec, draft.Media.MinWatchSec = 20, 10
	draft.StartsAt, draft.EndsAt = "2026-10-22T00:00:00Z", "2026-10-24T00:00:00Z"
	draft.RepeatPattern, draft.RepeatPeriod = "periodic", "daily"
	if _, _, err := validateMediaMissionDraft(draft); err != nil {
		t.Fatalf("periodic media draft rejected: %v", err)
	}
	draft.RepeatPattern = "unlimited"
	if _, _, err := validateMediaMissionDraft(draft); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("unsupported unlimited watch accepted: %v", err)
	}
	if pattern, err := supportedMissionRepeat("periodic", "weekly"); err != nil || pattern != "periodic" {
		t.Fatalf("weekly repeat rejected: %q %v", pattern, err)
	}
	if _, err := supportedMissionRepeat("once", "daily"); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("period on once mission accepted: %v", err)
	}
	if pattern, err := supportedReviewedRepeat("unlimited", ""); err != nil || pattern != "unlimited" {
		t.Fatalf("verified unlimited repeat rejected: %q %v", pattern, err)
	}
	if _, err := supportedReviewedRepeat("unlimited", "daily"); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("unlimited with period accepted: %v", err)
	}
	first, err := missionPeriodKey("unlimited", "", after)
	if err != nil || len(first) > 24 || !strings.HasPrefix(first, "U-") {
		t.Fatalf("invalid unlimited key: %q %v", first, err)
	}
	second, err := missionPeriodKey("unlimited", "", after)
	if err != nil || first == second {
		t.Fatalf("unlimited keys repeated: %q %q %v", first, second, err)
	}
}
