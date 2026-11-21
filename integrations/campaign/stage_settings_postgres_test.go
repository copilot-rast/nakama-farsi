package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

func f64(v float64) *float64 { return &v }

func (f *fixture) chapters(user string) map[string]any {
	f.t.Helper()
	out, err := f.s.rpcChapters(context.Background(), user, `{"lang":"fa"}`)
	if err != nil {
		f.t.Fatal(err)
	}
	return roundTrip(out)
}

func findChapter(out map[string]any, key string) map[string]any {
	for _, c := range out["chapters"].([]any) {
		if m := c.(map[string]any); m["key"] == key {
			return m
		}
	}
	return nil
}

func findStage(ch map[string]any, key string) map[string]any {
	for _, s := range ch["stages"].([]any) {
		if m := s.(map[string]any); m["key"] == key {
			return m
		}
	}
	return nil
}

func sixSteps() []map[string]any {
	out := []map[string]any{}
	for i, p := range []float64{16.6667, 33.3333, 50, 66.6667, 83.3333, 100} {
		out = append(out, map[string]any{"id": []string{"1", "2", "3", "4", "5", "6"}[i], "percent": p, "currencies": map[string]int64{"pt_gem": int64(10 * (i + 1))}})
	}
	return out
}

func TestPostgresMilestonesPayAtExactSharesAndKeepTheirIDs(t *testing.T) {
	f := newFixture(t)
	f.settings(func(a *settings.All) { a.Attempts.DefaultMinSeconds = 0; a.Stars.MissingField = settings.MissingPass })
	chKey := plugintest.Key("pt_ch")
	ch := f.chapter(map[string]any{"key": chKey, "milestones": sixSteps()})
	var stages []string
	for i := 0; i < 10; i++ {
		stages = append(stages, f.stage(ch, map[string]any{"sort": i + 1, "timeLimitSeconds": 100}))
	}
	user := f.user()
	f.play(user, stages[0], 50, 60)
	out := f.play(user, stages[1], 50, 60)
	if out["stars"] != 2.0 || len(out["milestoneIds"].([]any)) != 0 {
		t.Fatalf("4 stars paid a step: %v", out)
	}
	out = f.play(user, stages[2], 95, 60)
	if ids := out["milestoneIds"].([]any); out["stars"] != 1.0 || len(ids) != 1 || ids[0] != "1" {
		t.Fatalf("5 of 30 stars = 16.67%% must pay step 1: %v", out)
	}
	code_, saved := f.console(http.MethodGet, f.s.chapterGet, ch, nil)
	if code_ != 200 {
		t.Fatal(saved)
	}
	ms := saved["milestones"].([]any)
	ms[0].(map[string]any)["percent"] = 10.0
	saved["milestones"] = ms
	if c, o := f.console(http.MethodPut, f.s.chapterSave, ch, saved); c != 200 {
		t.Fatalf("edit percent = %d %v", c, o)
	}
	out = f.play(user, stages[3], 50, 60)
	if len(out["milestoneIds"].([]any)) != 0 {
		t.Fatalf("a changed percent paid step 1 again: %v", out)
	}
	for i := 4; i < 9; i++ {
		f.play(user, stages[i], 50, 60)
	}
	if out = f.play(user, stages[9], 10, 100); out["stars"] != 3.0 {
		t.Fatalf("last = %v", out)
	}
	cs := findChapter(f.chapters(user), chKey)
	if cs["stars"] != 20.0 {
		t.Fatalf("stars = %v", cs["stars"])
	}
	for _, m := range cs["milestones"].([]any) {
		mm := m.(map[string]any)
		if mm["id"] == "4" && (mm["claimed"] != true || mm["reached"] != true) {
			t.Fatalf("20 of 30 = 66.67%% must pay step 4: %v", mm)
		}
	}
	var paid string
	_ = f.db.QueryRow(`SELECT milestones::text FROM campaign_user_chapter WHERE user_id=$1`, user).Scan(&paid)
	if paid != `["1", "2", "3", "4"]` {
		t.Fatalf("paid ids = %s", paid)
	}

	code_, created := f.console(http.MethodPost, f.s.chapterSave, "", map[string]any{"key": plugintest.Key("pt_ch"), "name": map[string]string{"fa": "x"},
		"milestones": []map[string]any{{"percent": 50, "currencies": map[string]int64{"pt_gem": 1}}, {"id": "7", "percent": 100, "currencies": map[string]int64{"pt_gem": 1}}}})
	if code_ != 200 {
		t.Fatal(created)
	}
	got := created["milestones"].([]any)
	if got[0].(map[string]any)["id"] != "8" || got[1].(map[string]any)["id"] != "7" {
		t.Fatalf("ids = %v", got)
	}
	if c, o := f.console(http.MethodPost, f.s.chapterSave, "", map[string]any{"key": plugintest.Key("pt_ch"), "name": map[string]string{"fa": "x"},
		"milestones": []map[string]any{{"id": "a", "percent": 50, "currencies": map[string]int64{"pt_gem": 1}}, {"id": "a", "percent": 100, "currencies": map[string]int64{"pt_gem": 1}}}}); c != 400 || o["error"] != "DUPLICATE_MILESTONE" {
		t.Fatalf("duplicate id = %d %v", c, o)
	}
}

func TestPostgresMigrationGivesLegacyMilestonesIDs(t *testing.T) {
	f := newFixture(t)
	raw, err := os.ReadFile("../../migrate/sql/20270126000000-campaign-stage-settings.sql")
	if err != nil {
		t.Fatal(err)
	}
	up := strings.SplitN(strings.SplitN(string(raw), "-- +migrate Down", 2)[0], "-- +migrate Up", 2)[1]
	ch := f.chapter(map[string]any{"milestones": []map[string]any{{"percent": 16.67, "currencies": map[string]int64{"pt_gem": 20}}, {"percent": 100, "currencies": map[string]int64{"pt_gem": 25}}}})
	plugintest.Exec(t, f.db, `UPDATE campaign_chapter SET milestones='[{"percent":16.67,"currencies":{"pt_gem":20}},{"percent":100,"currencies":{"pt_gem":25}}]' WHERE id=$1`, ch)
	user := f.user()
	plugintest.Exec(t, f.db, `INSERT INTO campaign_user_chapter (user_id, chapter_id, milestones) VALUES ($1,$2,'[16.67]')`, user, ch)
	for _, stmt := range strings.Split(up, ";\n") {
		if strings.TrimSpace(stmt) != "" {
			plugintest.Exec(t, f.db, stmt)
		}
	}
	var ms, paid string
	_ = f.db.QueryRow(`SELECT milestones::text FROM campaign_chapter WHERE id=$1`, ch).Scan(&ms)
	_ = f.db.QueryRow(`SELECT milestones::text FROM campaign_user_chapter WHERE user_id=$1`, user).Scan(&paid)
	if !strings.Contains(ms, `"id": "1"`) || !strings.Contains(ms, `"id": "2"`) || paid != `["1"]` {
		t.Fatalf("migrated chapter=%s paid=%s", ms, paid)
	}
}

func TestPostgresTemplatesCriteriaReplayLockAndStageFields(t *testing.T) {
	f := newFixture(t)
	chKey := plugintest.Key("pt_ch")
	f.settings(func(a *settings.All) {
		a.Attempts.DefaultMinSeconds = 0
		a.Fields.Stage = []settings.FieldDef{
			{Key: "unityLevelId", Type: settings.FieldText},
			{Key: "brief", Type: settings.FieldI18n},
			{Key: "reachRadius", Type: settings.FieldNumber, Default: 8.0, Min: f64(0), Max: f64(500)},
			{Key: "missions", Type: settings.FieldList, Fields: []settings.FieldDef{
				{Key: "mode", Type: settings.FieldSelect, Options: []string{"target", "capture"}, Default: "target"},
				{Key: "enemies", Type: settings.FieldList, Fields: []settings.FieldDef{{Key: "speedMax", Type: settings.FieldNumber}}},
			}},
			{Key: "quizOnly", Type: settings.FieldBool, Chapters: []string{"pt_quiz_only"}},
			{Key: "trigger", Type: settings.FieldBool, Kinds: []string{"poi"}, Default: true},
		}
		a.Fields.Chapter = []settings.FieldDef{{Key: "offerEvery", Type: settings.FieldInteger, Default: 120.0}}
	})
	ch := f.chapter(map[string]any{"key": chKey})
	race := f.stage(ch, map[string]any{"mode": "repeatable", "timeLimitSeconds": 100, "starTemplate": "race", "minSeconds": 20, "noReplayAfterWin": true,
		"settings": map[string]any{"unityLevelId": "M_Race", "brief": map[string]any{"fa": "خلاصه", "en": "Brief"}, "missions": []any{map[string]any{"mode": "capture", "enemies": []any{map[string]any{"speedMax": 170.0}}}}}})
	for _, bad := range []struct {
		body map[string]any
		code string
	}{
		{map[string]any{"starTemplate": "nope"}, "INVALID_TEMPLATE"},
		{map[string]any{"settings": map[string]any{"typo": 1}}, "UNKNOWN_FIELD"},
		{map[string]any{"settings": map[string]any{"reachRadius": 900}}, "VALUE_OUT_OF_RANGE"},
		{map[string]any{"settings": map[string]any{"missions": []any{map[string]any{"mode": "fly"}}}}, "INVALID_OPTION"},
		{map[string]any{"settings": map[string]any{"quizOnly": true}}, "UNKNOWN_FIELD"},
	} {
		body := bad.body
		body["key"], body["chapterId"], body["name"] = plugintest.Key("pt_st"), ch, map[string]string{"fa": "x"}
		if c, o := f.console(http.MethodPost, f.s.stageSave, "", body); c != 400 || o["error"] != bad.code {
			t.Fatalf("%v = %d %v", bad.body, c, o)
		}
	}
	user := f.user()
	st := findStage(findChapter(f.chapters(user), chKey), race)
	set := st["settings"].(map[string]any)
	if set["unityLevelId"] != "M_Race" || set["brief"] != "خلاصه" || set["reachRadius"] != 8.0 {
		t.Fatalf("settings = %v", set)
	}
	if _, ok := set["quizOnly"]; ok {
		t.Fatalf("chapter-only field leaked: %v", set)
	}
	if _, ok := set["trigger"]; ok {
		t.Fatalf("poi-only field leaked: %v", set)
	}
	poiKey := plugintest.Key("pt_ch")
	poi := f.chapter(map[string]any{"key": poiKey, "metadata": map[string]any{"kind": "poi"}})
	village := f.stage(poi, map[string]any{"settings": map[string]any{"trigger": false}})
	if got := findStage(findChapter(f.chapters(user), poiKey), village)["settings"].(map[string]any)["trigger"]; got != false {
		t.Fatalf("poi kind field = %v", got)
	}
	m := set["missions"].([]any)[0].(map[string]any)
	if m["mode"] != "capture" || m["enemies"].([]any)[0].(map[string]any)["speedMax"] != 170.0 {
		t.Fatalf("missions = %v", m)
	}
	if st["starTemplate"] != "race" || st["minSeconds"] != 20.0 || st["replayable"] != true || len(st["starRules"].([]any)) != 2 {
		t.Fatalf("stage = %v", st)
	}
	if findChapter(f.chapters(user), chKey)["settings"].(map[string]any)["offerEvery"] != 120.0 {
		t.Fatal("chapter default missing")
	}
	a := f.start(user, race)
	f.advance(41 * time.Second)
	raw := `{"attemptId":"` + a + `","success":true,"elapsedSeconds":40,"healthPercent":5,"fields":{"timesSeen":1}}`
	res, err := f.s.rpcComplete(context.Background(), user, raw)
	if err != nil {
		t.Fatal(err)
	}
	out := roundTrip(res)
	if out["stars"] != 2.0 || out["minSeconds"] != 20.0 {
		t.Fatalf("race ignores health; timesSeen 1 caps at 2: %v", out)
	}
	tiers := map[string]float64{}
	for _, c := range out["criteria"].([]any) {
		cc := c.(map[string]any)
		tiers[cc["field"].(string)] = cc["stars"].(float64)
	}
	if tiers["timeSavedPercent"] != 3 || tiers["timesSeen"] != 2 || len(tiers) != 2 {
		t.Fatalf("criteria = %v", out["criteria"])
	}
	if out["replayable"] != false {
		t.Fatalf("replayable after a win = %v", out["replayable"])
	}
	if _, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+race+`"}`); code(err) != "STAGE_ALREADY_WON" {
		t.Fatalf("replay after win = %v", err)
	}
	if _, err := f.s.claimFreeride(context.Background(), user, "anything", "r1"); code(err) != "FREERIDE_NOT_FOUND" {
		t.Fatalf("freeride table off by default = %v", err)
	}
}

func TestSettingsFieldsValidate(t *testing.T) {
	cfg := settings.Defaults()
	if f, c := cfg.Validate(); c != "" {
		t.Fatalf("defaults: %s %s", f, c)
	}
	if cfg.Stars.Template("base") == nil || cfg.Stars.Template("stealth") == nil || cfg.Freeride.TableEnabled {
		t.Fatal("defaults")
	}
	cfg.Fields.Stage = []settings.FieldDef{{Key: "a", Type: settings.FieldSelect}}
	if f, c := cfg.Validate(); c != "OPTIONS_REQUIRED" {
		t.Fatalf("%s %s", f, c)
	}
	deep := settings.FieldDef{Key: "a", Type: settings.FieldList, Fields: []settings.FieldDef{{Key: "b", Type: settings.FieldList, Fields: []settings.FieldDef{{Key: "c", Type: settings.FieldList, Fields: []settings.FieldDef{{Key: "d", Type: settings.FieldText}}}}}}}
	cfg.Fields.Stage = []settings.FieldDef{deep}
	if _, c := cfg.Validate(); c != "FIELDS_TOO_DEEP" {
		t.Fatalf("deep = %s", c)
	}
	cfg.Fields.Stage = []settings.FieldDef{{Key: "n", Type: settings.FieldNumber, Default: "x"}}
	if _, c := cfg.Validate(); c != "INVALID_FIELD_VALUE" {
		t.Fatalf("bad default = %s", c)
	}
}
