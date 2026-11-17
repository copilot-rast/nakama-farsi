package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/settings"
)

type upFile struct {
	name string
	data []byte
}

func lap(n int, header bool, player string) string {
	var b strings.Builder
	if header {
		b.WriteString("t,x,z,player,kmh\n")
	}
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "%.2f,%d,%d,%s,%d\n", float64(i)*0.1, i, i%7, player, 80+i%50)
	}
	return b.String()
}

func gzBytes(s string) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return b.Bytes()
}

func zipBytes(t *testing.T, files map[string]string) []byte {
	var b bytes.Buffer
	zw := zip.NewWriter(&b)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func (f *fixture) upload(key string, fields map[string]string, files ...upFile) (int, map[string]any) {
	f.t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	for _, file := range files {
		w, _ := mw.CreateFormFile("file", file.name)
		_, _ = w.Write(file.data)
	}
	_ = mw.Close()
	r := httptest.NewRequest(http.MethodPost, "/import", &body)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	r.SetPathValue("key", key)
	w := httptest.NewRecorder()
	f.s.importCSV(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func files(t *testing.T, out map[string]any) []map[string]any {
	t.Helper()
	list, _ := out["files"].([]any)
	res := []map[string]any{}
	for _, it := range list {
		res = append(res, it.(map[string]any))
	}
	return res
}

func TestHeatmapImportFormats(t *testing.T) {
	f := newFixture(t)
	key := f.prefix + "_road"
	code, out := f.upload(key, map[string]string{"mode": "add"}, upFile{"a.csv", []byte(lap(10, true, "me"))})
	if code != http.StatusNotFound {
		t.Fatalf("unknown map without create = %d %v", code, out)
	}
	bad := "t,x,z,player,kmh\n1,2,3,me,50\nbroken\n1,abc,3,me,50\n2,4,5,me,99999\n3,6,7,me,40\n"
	semi := "time;posX;posY;posZ;speed;name\n0;10;1;20;55;ali\n0.5;11;1;21;60;ali\n"
	code, out = f.upload(key, map[string]string{"create": "true", "name": "Road", "cellSize": "2"},
		upFile{"plain.csv", []byte(lap(50, true, "me"))},
		upFile{"nohead.csv", []byte(lap(30, false, "bot"))},
		upFile{"packed.csv.gz", gzBytes(lap(20, true, "gz"))},
		upFile{"bad.csv", []byte(bad)},
		upFile{"semi.csv", []byte(semi)},
		upFile{"race.zip", zipBytes(t, map[string]string{"r1.csv": lap(5, true, "z1"), "dir/r2.csv.gz": string(gzBytes(lap(6, false, "z2"))), "__MACOSX/._r1.csv": "x", "notes.md": "hi"})},
	)
	if code != http.StatusOK {
		t.Fatalf("import = %d %v", code, out)
	}
	list := files(t, out)
	want := map[string][3]float64{"plain.csv": {50, 0, 50}, "nohead.csv": {30, 0, 30}, "packed.csv.gz": {20, 0, 20}, "bad.csv": {5, 3, 2}, "semi.csv": {2, 0, 2}, "race.zip/r1.csv": {5, 0, 5}, "race.zip/dir/r2.csv.gz": {6, 0, 6}}
	if len(list) != len(want) {
		t.Fatalf("files = %v", list)
	}
	for _, it := range list {
		w, ok := want[it["name"].(string)]
		if !ok || it["status"] != "stored" || it["rows"] != w[0] || it["dropped"] != w[1] || it["points"] != w[2] {
			t.Fatalf("file %v, want %v", it, w)
		}
	}
	for _, it := range list {
		if it["name"] == "bad.csv" {
			r := it["reasons"].(map[string]any)
			if r["columns"] != 1.0 || r["number"] != 1.0 || r["range"] != 1.0 {
				t.Fatalf("reasons = %v", r)
			}
		}
		if it["name"] == "plain.csv" && it["label"] != "plain" {
			t.Fatalf("label = %v", it["label"])
		}
	}
	m := out["map"].(map[string]any)
	if m["name"] != "Road" || m["cellSize"] != 2.0 || m["points"] != 115.0 || m["uploads"] != 7.0 {
		t.Fatalf("map = %v", m)
	}
	cells := get(t, f.s.mapCells, "/cells?from=2026-06-01&to=2026-06-01", map[string]string{"key": key})
	if tt := cells["totals"].(map[string]any); tt["points"] != 115.0 || tt["players"] != 1.0 || cells["size"] != 2.0 {
		t.Fatalf("cells = %v", cells["totals"])
	}
	ups := get(t, f.s.mapUploads, "/uploads?limit=100", map[string]string{"key": key})
	var semiID string
	for _, it := range ups["items"].([]any) {
		u := it.(map[string]any)
		if u["source"] != "console" || u["hasTrace"] != true {
			t.Fatalf("upload = %v", u)
		}
		if u["fileName"] == "semi.csv" {
			semiID = u["id"].(string)
		}
	}
	tr := get(t, f.s.uploadTrace, "/trace", map[string]string{"id": semiID})
	pl := tr["players"].([]any)[0].(map[string]any)
	if pl["player"] != "ali" || tr["points"] != 2.0 {
		t.Fatalf("trace = %v", tr)
	}
	pts := pl["points"].([]any)
	if p := pts[1].([]any); p[1] != 11.0 || p[2] != 21.0 || p[3] != 60.0 {
		t.Fatalf("semi point = %v", p)
	}
}

func TestHeatmapImportDuplicateReplaceAndLimits(t *testing.T) {
	f := newFixture(t)
	key := f.prefix + "_dup"
	one := lap(40, true, "me")
	code, out := f.upload(key, map[string]string{"create": "true", "race": "label", "label": "Cup final", "player": "fixed", "playerName": "Sara"}, upFile{"one.csv", []byte(one)})
	if code != 200 || files(t, out)[0]["status"] != "stored" || files(t, out)[0]["label"] != "Cup final" {
		t.Fatalf("first = %d %v", code, out)
	}
	_, out = f.upload(key, nil, upFile{"again.csv.gz", gzBytes(one)}, upFile{"two.csv", []byte(lap(10, true, "x"))})
	list := files(t, out)
	if list[0]["status"] != "duplicate" || list[0]["duplicate"] != true || list[0]["points"] != 0.0 || list[1]["status"] != "stored" {
		t.Fatalf("second = %v", list)
	}
	if m := out["map"].(map[string]any); m["points"] != 50.0 || m["uploads"] != 2.0 {
		t.Fatalf("add map = %v", m)
	}
	ups := get(t, f.s.mapUploads, "/uploads", map[string]string{"key": key})
	var id string
	for _, it := range ups["items"].([]any) {
		if u := it.(map[string]any); u["fileName"] == "one.csv" {
			id = u["id"].(string)
			if u["label"] != "Cup final" {
				t.Fatalf("label = %v", u)
			}
		}
	}
	tr := get(t, f.s.uploadTrace, "/trace", map[string]string{"id": id})
	if tr["players"].([]any)[0].(map[string]any)["player"] != "Sara" {
		t.Fatalf("fixed player = %v", tr["players"])
	}
	_, out = f.upload(key, map[string]string{"mode": "replace"}, upFile{"one.csv", []byte(one)})
	if files(t, out)[0]["status"] != "stored" || out["replaced"] != true {
		t.Fatalf("replace = %v", out)
	}
	if m := out["map"].(map[string]any); m["points"] != 40.0 || m["uploads"] != 1.0 {
		t.Fatalf("replace map = %v", m)
	}
	if code, out := f.upload(key, map[string]string{"race": "label"}, upFile{"x.csv", []byte(one)}); code != 400 {
		t.Fatalf("label required = %d %v", code, out)
	}
	if code, _ := f.upload(key, map[string]string{"mode": "wipe"}, upFile{"x.csv", []byte(one)}); code != 400 {
		t.Fatalf("bad mode = %d", code)
	}
	if code, _ := f.upload(key, map[string]string{"mode": "add"}); code != 400 {
		t.Fatalf("no file = %d", code)
	}
	f.settings(func(c *settings.All) {
		c.Heatmap.ImportMaxRows = 100
		c.Heatmap.ImportMaxFiles = 2
		c.Heatmap.ImportMaxMB = 1
	})
	big := lap(150, true, "big")
	huge := strings.Repeat("1,2,3,"+strings.Repeat("a", 60000)+",50\n", 20)
	_, out = f.upload(key, nil, upFile{"big.csv", []byte(big)}, upFile{"huge.csv", []byte(huge)}, upFile{"third.csv", []byte(lap(3, true, "t"))})
	list = files(t, out)
	if list[0]["status"] != "failed" || list[0]["error"] != "HEATMAP_TOO_LARGE" || list[0]["meta"].(map[string]any)["maxRows"] != 100.0 {
		t.Fatalf("rows limit = %v", list[0])
	}
	if list[1]["status"] != "failed" || list[1]["error"] != "HEATMAP_TOO_LARGE" {
		t.Fatalf("bytes limit = %v", list[1])
	}
	if list[2]["status"] != "skipped" || list[2]["error"] != "TOO_MANY_FILES" {
		t.Fatalf("files limit = %v", list[2])
	}
	_, out = f.upload(key, nil, upFile{"empty.csv", []byte("t,x,z\nfoo,bar,baz\n")})
	if l := files(t, out); l[0]["status"] != "failed" || l[0]["error"] != "HEATMAP_INVALID" || l[0]["dropped"] != 1.0 {
		t.Fatalf("no rows = %v", l)
	}
}

func TestScanCSVLayouts(t *testing.T) {
	cases := []struct {
		in      string
		rows    int
		dropped int
		first   point
	}{
		{"t,x,z,player,kmh\n0.5,1,2,me,30\n", 1, 0, point{T: 0.5, X: 1, Z: 2, Player: "me", Kmh: 30}},
		{"0.5,1,2,me,30\n1,2,3,me,40\n", 2, 0, point{T: 0.5, X: 1, Z: 2, Player: "me", Kmh: 30}},
		{"\ufeffPlayer,Speed (km/h),PosX,PosZ,Time\nali,70,5,6,1.5\n", 1, 0, point{T: 1.5, X: 5, Z: 6, Player: "ali", Kmh: 70}},
		{"0\t1\t2\tme\t9\n", 1, 0, point{X: 1, Z: 2, Player: "me", Kmh: 9}},
		{"1,2,3\n\n1,2\n", 2, 1, point{T: 1, X: 2, Z: 3}},
	}
	for i, c := range cases {
		var got []point
		st, err := scanCSV(strings.NewReader(c.in), func(p point) error { got = append(got, p); return nil })
		if err != nil || st.Rows != c.rows || st.Dropped != c.dropped || len(got) == 0 || got[0] != c.first {
			t.Fatalf("case %d: %+v %v %v", i, st, got, err)
		}
	}
}
