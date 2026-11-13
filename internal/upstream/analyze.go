package upstream

import (
	"path"
	"regexp"
	"sort"
	"strings"
)

var AreaOrder = []string{"migrations", "core", "plugins", "console", "deps", "build", "e2e", "docs", "other"}

func AreaOf(p string) string {
	base := path.Base(p)
	switch {
	case strings.HasPrefix(p, "vendor/") || p == "go.mod" || p == "go.sum":
		return "deps"
	case strings.HasPrefix(p, "migrate/sql/"):
		return "migrations"
	case strings.HasPrefix(p, "e2e/") || strings.HasSuffix(p, "_test.go") || strings.Contains(p, "/testdata/"):
		return "e2e"
	case strings.HasPrefix(p, "console/"):
		return "console"
	case strings.HasPrefix(p, "integrations/") || strings.HasPrefix(p, "samples/") || strings.HasPrefix(p, "sample_go_module/") || strings.HasPrefix(p, "data/modules/"):
		return "plugins"
	case strings.HasPrefix(p, "build/") || strings.HasPrefix(p, ".github/") || strings.HasPrefix(base, "docker-compose") || strings.HasPrefix(base, "Dockerfile") || base == "Makefile" || strings.HasPrefix(base, "buf.") || base == ".dockerignore" || base == ".goreleaser.yml":
		return "build"
	case strings.HasPrefix(p, "docs/") || strings.HasSuffix(strings.ToLower(p), ".md") || base == "LICENSE":
		return "docs"
	case strings.HasPrefix(p, "server/") || strings.HasPrefix(p, "internal/") || strings.HasPrefix(p, "apigrpc/") || strings.HasPrefix(p, "social/") || strings.HasPrefix(p, "iap/") || strings.HasPrefix(p, "se/") || strings.HasPrefix(p, "flags/") || strings.HasPrefix(p, "migrate/") || p == "main.go":
		return "core"
	}
	return "other"
}

func Areas(files []File) []Area {
	by := map[string]*Area{}
	for _, f := range files {
		k := AreaOf(f.Path)
		a := by[k]
		if a == nil {
			a = &Area{Key: k, Files: []File{}}
			by[k] = a
		}
		a.Count++
		a.Additions += f.Additions
		a.Deletions += f.Deletions
		if k != "deps" || !strings.HasPrefix(f.Path, "vendor/") || len(a.Files) < 20 {
			a.Files = append(a.Files, File{Path: f.Path, Status: f.Status, Additions: f.Additions, Deletions: f.Deletions, Previous: f.Previous})
		}
	}
	out := []Area{}
	for _, k := range AreaOrder {
		if a := by[k]; a != nil {
			sort.SliceStable(a.Files, func(i, j int) bool { return a.Files[i].Path < a.Files[j].Path })
			out = append(out, *a)
		}
	}
	return out
}

var breakingPattern = regexp.MustCompile(`(?i)\bbreaking\b|\bbackwards?[- ]incompatible\b|\bincompatible change|\bdeprecat(ed|es|ion)\b|^[a-z]+(\([^)]*\))?!:|^#+\s*removed\b`)

func Breaking(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		if breakingPattern.MatchString(strings.TrimSpace(line)) {
			return true
		}
	}
	return false
}

var (
	envGo     = regexp.MustCompile(`(?:Getenv|LookupEnv)\("([A-Z][A-Z0-9_]{2,})"\)`)
	envDollar = regexp.MustCompile(`\$\{?([A-Z][A-Z0-9]*_[A-Z0-9_]+)`)
	envAssign = regexp.MustCompile(`(?:^|[\s"'\x60(])([A-Z][A-Z0-9]*_[A-Z0-9_]+)=`)
	envYAML   = regexp.MustCompile(`^\s*-?\s*([A-Z][A-Z0-9]*_[A-Z0-9_]+):\s`)
	goLine    = regexp.MustCompile(`^(go|toolchain)\s+(go)?([0-9][0-9.a-z]*)\s*$`)
	goImage   = regexp.MustCompile(`(?i)^FROM\s+golang:([0-9][0-9.]*)`)
	modLine   = regexp.MustCompile(`^\s*(require\s+)?([a-z0-9][^\s]*\.[^\s]+)\s+v[0-9][^\s]*`)
)

func envNames(p, line string) []string {
	var out []string
	add := func(re *regexp.Regexp) {
		for _, m := range re.FindAllStringSubmatch(line, -1) {
			out = append(out, m[1])
		}
	}
	switch ext := path.Ext(p); {
	case ext == ".go":
		add(envGo)
	case ext == ".md":
		add(envAssign)
	case ext == ".yml" || ext == ".yaml":
		add(envYAML)
		add(envDollar)
		add(envAssign)
	default:
		add(envDollar)
		add(envAssign)
	}
	return out
}

func envFile(p string) bool {
	if strings.HasPrefix(p, "vendor/") || strings.HasSuffix(p, "_test.go") || strings.HasPrefix(p, "e2e/") {
		return false
	}
	base := path.Base(p)
	switch {
	case strings.HasSuffix(p, ".go"), strings.HasSuffix(p, ".sh"), strings.HasPrefix(base, "docker-compose"), strings.HasPrefix(base, "Dockerfile"), strings.Contains(base, ".env"), strings.HasSuffix(p, ".yml"), strings.HasSuffix(p, ".yaml"):
		return true
	case strings.HasSuffix(p, ".md") && (strings.HasPrefix(p, "docs/") || strings.HasPrefix(p, "build/") || base == "README.md"):
		return true
	}
	return false
}

func patchHints(files []File) []string {
	added := map[string]bool{}
	removed := map[string]bool{}
	hints := []string{}
	seen := map[string]bool{}
	hint := func(h string) {
		if !seen[h] {
			seen[h] = true
			hints = append(hints, h)
		}
	}
	for _, f := range files {
		if f.patch == "" {
			continue
		}
		for _, raw := range strings.Split(f.patch, "\n") {
			if raw == "" || strings.HasPrefix(raw, "+++") || strings.HasPrefix(raw, "---") {
				continue
			}
			sign, line := raw[0], raw[1:]
			if sign != '+' && sign != '-' {
				continue
			}
			if envFile(f.Path) {
				for _, n := range envNames(f.Path, line) {
					if sign == '+' {
						added[n] = true
					} else {
						removed[n] = true
					}
				}
			}
			if f.Path == "go.mod" {
				if m := goLine.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
					hint("go:" + string(sign) + m[1] + " " + m[3])
				} else if m := modLine.FindStringSubmatch(line); m != nil && sign == '+' {
					hint("dep:" + m[2])
				}
			}
			if strings.HasPrefix(path.Base(f.Path), "Dockerfile") && sign == '+' {
				if m := goImage.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
					hint("go:+image " + m[1])
				}
			}
		}
	}
	names := []string{}
	for n := range added {
		if !removed[n] {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	for _, n := range names {
		hint("env:" + n)
	}
	return hints
}

func Warnings(cd *CompareData, rel []Release) []Warning {
	out := []Warning{}
	if cd == nil {
		return out
	}
	var newMig, changedMig []string
	goMod := false
	vendor := 0
	for _, f := range cd.Files {
		switch {
		case strings.HasPrefix(f.Path, "migrate/sql/") && (f.Status == "added" || f.Status == "copied"):
			newMig = append(newMig, path.Base(f.Path))
		case strings.HasPrefix(f.Path, "migrate/sql/"):
			changedMig = append(changedMig, path.Base(f.Path))
		case f.Path == "go.mod":
			goMod = true
		case strings.HasPrefix(f.Path, "vendor/"):
			vendor++
		}
	}
	sort.Strings(newMig)
	sort.Strings(changedMig)
	if len(newMig) > 0 {
		out = append(out, Warning{Kind: "migrations", Items: newMig})
	}
	if len(changedMig) > 0 {
		out = append(out, Warning{Kind: "migrationsChanged", Items: changedMig})
	}
	var env, goVer, deps []string
	for _, h := range cd.Patches {
		switch {
		case strings.HasPrefix(h, "env:"):
			env = append(env, h[4:])
		case strings.HasPrefix(h, "go:"):
			goVer = append(goVer, h[3:])
		case strings.HasPrefix(h, "dep:"):
			deps = append(deps, h[4:])
		}
	}
	if len(env) > 0 {
		out = append(out, Warning{Kind: "env", Items: env})
	}
	if len(goVer) > 0 {
		out = append(out, Warning{Kind: "goVersion", Items: goChanges(goVer)})
	}
	if goMod || vendor > 0 {
		items := deps
		if len(items) > 20 {
			items = items[:20]
		}
		if items == nil {
			items = []string{}
		}
		out = append(out, Warning{Kind: "deps", Items: items})
	}
	var brk []string
	for _, c := range cd.Commits {
		if c.Breaks {
			brk = append(brk, shortSHA(c.SHA)+" "+c.Title)
		}
	}
	for _, r := range rel {
		if r.Breaks {
			brk = append(brk, r.Tag)
		}
	}
	if len(brk) > 0 {
		if len(brk) > 20 {
			brk = brk[:20]
		}
		out = append(out, Warning{Kind: "breaking", Items: brk})
	}
	return out
}

func shortSHA(s string) string {
	if len(s) > 7 {
		return s[:7]
	}
	return s
}

func goChanges(hints []string) []string {
	old := map[string]string{}
	neu := map[string]string{}
	var order []string
	for _, h := range hints {
		if len(h) < 2 {
			continue
		}
		sign, rest := h[0], h[1:]
		name, ver, ok := strings.Cut(rest, " ")
		if !ok {
			continue
		}
		if _, seen := old[name]; !seen {
			if _, seen2 := neu[name]; !seen2 {
				order = append(order, name)
			}
		}
		if sign == '-' {
			old[name] = ver
		} else {
			neu[name] = ver
		}
	}
	out := []string{}
	for _, name := range order {
		label := name
		if name == "image" {
			label = "golang image"
		}
		switch {
		case old[name] != "" && neu[name] != "":
			out = append(out, label+" "+old[name]+" → "+neu[name])
		case neu[name] != "":
			out = append(out, label+" → "+neu[name])
		default:
			out = append(out, label+" "+old[name]+" removed")
		}
	}
	return out
}
