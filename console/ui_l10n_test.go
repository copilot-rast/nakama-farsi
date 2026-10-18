// Copyright 2026 The Nakama Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package console

import (
	"bytes"
	"compress/gzip"
	"io"
	"regexp"
	"strings"
	"testing"
)

func readUIFile(t *testing.T, name string) []byte {
	t.Helper()
	f, err := UIFS.Open(name)
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return b
}

// TestLocalizedIndex verifies the embedded console loads the Persian
// localization layer (console/ui/l10n) and that every asset it references is
// embedded, including the .gz copies the console server serves for /static
// scripts and stylesheets.
func TestLocalizedIndex(t *testing.T) {
	index := string(readUIFile(t, "index.html"))

	if !strings.Contains(index, `<html lang="fa-IR" dir="rtl">`) {
		t.Error("index.html should default to lang=fa-IR dir=rtl")
	}
	start := strings.Index(index, "<!-- nakama-l10n:start -->")
	end := strings.Index(index, "<!-- nakama-l10n:end -->")
	bundle := strings.Index(index, `<script type="module"`)
	if start == -1 || end == -1 || bundle == -1 || !(start < end && end < bundle) {
		t.Fatal("localization block must be present and load before the console bundle")
	}
	if !strings.Contains(index, "{{nt}}") {
		t.Error("index.html must keep the {{nt}} placeholder replaced by the server")
	}

	assets := regexp.MustCompile(`(?:href|src)="/static/([^"]+)"`).FindAllStringSubmatch(index, -1)
	if len(assets) < 4 {
		t.Fatalf("expected the console and localization assets to be referenced, got %d", len(assets))
	}
	for _, m := range assets {
		name := "static/" + m[1]
		content := readUIFile(t, name)
		if !strings.HasSuffix(name, ".js") && !strings.HasSuffix(name, ".css") {
			continue
		}
		gz, err := gzip.NewReader(bytes.NewReader(readUIFile(t, name+".gz")))
		if err != nil {
			t.Fatalf("%s.gz: %v", name, err)
		}
		unzipped, err := io.ReadAll(gz)
		if err != nil {
			t.Fatalf("%s.gz: %v", name, err)
		}
		if !bytes.Equal(unzipped, content) {
			t.Errorf("%s.gz does not match %s", name, name)
		}
	}

	// The original stylesheet stays embedded next to its logical-property copy.
	source := regexp.MustCompile(`data-l10n-source="/static/([^"]+)"`).FindStringSubmatch(index)
	if source == nil {
		t.Fatal("logical stylesheet should record its source stylesheet")
	}
	readUIFile(t, "static/"+source[1])

	runtime := regexp.MustCompile(`<script src="/static/(l10n-[a-f0-9]+\.js)">`).FindStringSubmatch(index)
	if runtime == nil {
		t.Fatal("localization runtime script not referenced")
	}
	js := string(readUIFile(t, "static/"+runtime[1]))
	for _, want := range []string{"__NAKAMA_L10N_BUILD__", `"defaultLocale":"fa-IR"`, "بازیکنان"} {
		if !strings.Contains(js, want) {
			t.Errorf("runtime should contain %q", want)
		}
	}
	readUIFile(t, "static/vazirmatn-OFL.txt")
}
