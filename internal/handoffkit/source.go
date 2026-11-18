package handoffkit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
)

const (
	ScopeGeneral = "general"
	ScopeProject = "project"
)

type Entry struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	TitleEn string `json:"titleEn"`
	Section string `json:"section,omitempty"`
	Scope   string `json:"scope"`
	Anchor  string `json:"anchor,omitempty"`
}

type Meta struct {
	Name      string  `json:"name"`
	Title     string  `json:"title"`
	TitleEn   string  `json:"titleEn"`
	Summary   string  `json:"summary"`
	SummaryEn string  `json:"summaryEn"`
	Changelog []Entry `json:"changelog"`
}

type Section struct {
	Scope   string `json:"scope"`
	Path    string `json:"path"`
	ID      string `json:"id"`
	Title   string `json:"title"`
	TitleEn string `json:"titleEn,omitempty"`
	Summary string `json:"summary,omitempty"`
	kind    string
	body    string
}

type Part struct {
	Scope    string
	Meta     Meta
	Sections []*Section
}

type Bundle struct {
	General Part
	Project *Part
}

var (
	dateRe    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	idCleanRe = regexp.MustCompile(`[^a-z0-9]+`)
)

func Load(fsys fs.FS) (*Bundle, error) {
	g, err := loadPart(fsys, ScopeGeneral)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, errors.New("handoff: general/handoff.json is missing")
	}
	b := &Bundle{General: *g}
	p, err := loadPart(fsys, ScopeProject)
	if err != nil {
		return nil, err
	}
	if p != nil && len(p.Sections) > 0 {
		b.Project = p
	}
	return b, nil
}

func loadPart(fsys fs.FS, scope string) (*Part, error) {
	raw, err := fs.ReadFile(fsys, scope+"/handoff.json")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p := &Part{Scope: scope}
	if err := json.Unmarshal(raw, &p.Meta); err != nil {
		return nil, fmt.Errorf("handoff: %s/handoff.json: %w", scope, err)
	}
	err = fs.WalkDir(fsys, scope, func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := path.Ext(name)
		if ext != ".md" && ext != ".html" {
			return nil
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(name, scope+"/")
		s := &Section{Scope: scope, Path: rel, ID: sectionID(scope, rel), kind: ext}
		front, body := frontMatter(string(data))
		s.Title, s.TitleEn, s.Summary = front["title"], front["titleEn"], front["summary"]
		s.body = body
		p.Sections = append(p.Sections, s)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(p.Sections, func(i, j int) bool { return p.Sections[i].Path < p.Sections[j].Path })
	for i := range p.Meta.Changelog {
		e := &p.Meta.Changelog[i]
		if !dateRe.MatchString(e.Date) {
			return nil, fmt.Errorf("handoff: %s changelog entry %d: date %q is not YYYY-MM-DD", scope, i, e.Date)
		}
		if strings.TrimSpace(e.Title) == "" {
			return nil, fmt.Errorf("handoff: %s changelog entry %d has no title", scope, i)
		}
		e.Scope = scope
		if e.Section != "" {
			found := false
			for _, s := range p.Sections {
				if s.Path == e.Section {
					e.Anchor, found = s.ID, true
				}
			}
			if !found {
				return nil, fmt.Errorf("handoff: %s changelog entry %d points to missing section %q", scope, i, e.Section)
			}
		}
	}
	return p, nil
}

func sectionID(scope, rel string) string {
	p := "g"
	if scope == ScopeProject {
		p = "p"
	}
	return p + "-" + strings.Trim(idCleanRe.ReplaceAllString(strings.ToLower(strings.TrimSuffix(rel, path.Ext(rel))), "-"), "-")
}

func frontMatter(src string) (map[string]string, string) {
	out := map[string]string{}
	src = strings.TrimPrefix(src, "\uFEFF")
	if !strings.HasPrefix(src, "---\n") {
		return out, src
	}
	end := strings.Index(src[4:], "\n---")
	if end < 0 {
		return out, src
	}
	for _, l := range strings.Split(src[4:4+end], "\n") {
		k, v, ok := strings.Cut(l, ":")
		if ok {
			out[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	rest := src[4+end+4:]
	return out, strings.TrimPrefix(rest, "\n")
}

func (b *Bundle) Changelog() []Entry {
	var all []Entry
	all = append(all, b.General.Meta.Changelog...)
	if b.Project != nil {
		all = append(all, b.Project.Meta.Changelog...)
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].Date > all[j].Date })
	return all
}

func (b *Bundle) Sections() []*Section {
	out := append([]*Section{}, b.General.Sections...)
	if b.Project != nil {
		out = append(out, b.Project.Sections...)
	}
	return out
}
