package upstream

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	maxCommits     = 100
	maxReleases    = 10
	maxReleaseBody = 8000
	maxTags        = 20
)

type Request struct {
	Repo     string
	Branch   string
	Base     string
	BaseDate time.Time
}

type Commit struct {
	SHA    string    `json:"sha"`
	Title  string    `json:"title"`
	Author string    `json:"author"`
	Date   time.Time `json:"date"`
	URL    string    `json:"url"`
	Merge  bool      `json:"merge,omitempty"`
	Breaks bool      `json:"breaks,omitempty"`
}

type File struct {
	Path      string `json:"path"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Previous  string `json:"previous,omitempty"`
	patch     string
}

type Area struct {
	Key       string `json:"key"`
	Count     int    `json:"count"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Files     []File `json:"files"`
}

type Release struct {
	Tag         string    `json:"tag"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"publishedAt"`
	URL         string    `json:"url"`
	Body        string    `json:"body"`
	Prerelease  bool      `json:"prerelease,omitempty"`
	Breaks      bool      `json:"breaks,omitempty"`
}

type Tag struct {
	Name string `json:"name"`
	SHA  string `json:"sha"`
	URL  string `json:"url"`
}

type Warning struct {
	Kind  string   `json:"kind"`
	Items []string `json:"items"`
}

type Source struct {
	Path string `json:"path"`
	ETag string `json:"etag"`
}

type Result struct {
	Repo          string            `json:"repo"`
	Branch        string            `json:"branch"`
	Base          string            `json:"base"`
	CheckedAt     time.Time         `json:"checkedAt"`
	Status        string            `json:"status"`
	AheadBy       int               `json:"aheadBy"`
	BehindBy      int               `json:"behindBy"`
	HeadSHA       string            `json:"headSha"`
	HeadDate      *time.Time        `json:"headDate,omitempty"`
	CompareURL    string            `json:"compareUrl"`
	Truncated     bool              `json:"truncated"`
	Commits       []Commit          `json:"commits"`
	FileCount     int               `json:"fileCount"`
	Areas         []Area            `json:"areas"`
	Releases      []Release         `json:"releases"`
	Tags          []Tag             `json:"tags"`
	LatestRelease *Release          `json:"latestRelease,omitempty"`
	LatestTag     *Tag              `json:"latestTag,omitempty"`
	Warnings      []Warning         `json:"warnings"`
	Rate          Rate              `json:"rate"`
	NotModified   []string          `json:"notModified,omitempty"`
	Sources       map[string]Source `json:"sources,omitempty"`
	AllReleases   []Release         `json:"allReleases,omitempty"`
	AllTags       []Tag             `json:"allTags,omitempty"`
	Compare       *CompareData      `json:"compare,omitempty"`
}

type CompareData struct {
	Status   string   `json:"status"`
	AheadBy  int      `json:"aheadBy"`
	BehindBy int      `json:"behindBy"`
	Total    int      `json:"total"`
	Commits  []Commit `json:"commits"`
	Files    []File   `json:"files"`
	Patches  []string `json:"patches,omitempty"`
}

type ghCompare struct {
	Status       string `json:"status"`
	AheadBy      int    `json:"ahead_by"`
	BehindBy     int    `json:"behind_by"`
	TotalCommits int    `json:"total_commits"`
	Commits      []struct {
		SHA    string `json:"sha"`
		URL    string `json:"html_url"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name string    `json:"name"`
				Date time.Time `json:"date"`
			} `json:"author"`
			Committer struct {
				Date time.Time `json:"date"`
			} `json:"committer"`
		} `json:"commit"`
		Author *struct {
			Login string `json:"login"`
		} `json:"author"`
		Parents []struct {
			SHA string `json:"sha"`
		} `json:"parents"`
	} `json:"commits"`
	Files []struct {
		Filename  string `json:"filename"`
		Status    string `json:"status"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
		Previous  string `json:"previous_filename"`
		Patch     string `json:"patch"`
	} `json:"files"`
}

type ghRelease struct {
	TagName     string     `json:"tag_name"`
	Name        string     `json:"name"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	URL         string     `json:"html_url"`
	Body        string     `json:"body"`
	Draft       bool       `json:"draft"`
	Prerelease  bool       `json:"prerelease"`
}

type ghTag struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
}

func escapePath(repo string) string {
	parts := strings.SplitN(repo, "/", 2)
	return url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1])
}

func (c *Client) Check(ctx context.Context, req Request, prev *Result) (*Result, error) {
	repo := escapePath(req.Repo)
	res := &Result{Repo: req.Repo, Branch: req.Branch, Base: req.Base, CheckedAt: c.now().UTC(), Sources: map[string]Source{}}
	old := func(key, path string) string {
		if prev == nil || prev.Sources == nil {
			return ""
		}
		if s, ok := prev.Sources[key]; ok && s.Path == path {
			return s.ETag
		}
		return ""
	}
	comparePath := "/repos/" + repo + "/compare/" + url.PathEscape(req.Base) + "..." + url.PathEscape(req.Branch) + "?per_page=250"
	r, err := c.get(ctx, comparePath, old("compare", comparePath))
	if err != nil {
		return nil, err
	}
	res.Rate = r.rate
	if r.notModified && prev != nil && prev.Compare != nil {
		res.Compare = prev.Compare
		res.NotModified = append(res.NotModified, "compare")
	} else {
		var gc ghCompare
		if err := json.Unmarshal(r.body, &gc); err != nil {
			return nil, err
		}
		res.Compare = compareData(gc)
	}
	res.Sources["compare"] = Source{Path: comparePath, ETag: r.etag}
	if r.notModified && r.etag == "" {
		res.Sources["compare"] = Source{Path: comparePath, ETag: old("compare", comparePath)}
	}

	releasesPath := "/repos/" + repo + "/releases?per_page=30"
	r, err = c.get(ctx, releasesPath, old("releases", releasesPath))
	if err != nil {
		return nil, err
	}
	res.Rate = r.rate
	if r.notModified && prev != nil {
		res.AllReleases = prev.AllReleases
		res.NotModified = append(res.NotModified, "releases")
		if r.etag == "" {
			r.etag = old("releases", releasesPath)
		}
	} else {
		var list []ghRelease
		if err := json.Unmarshal(r.body, &list); err != nil {
			return nil, err
		}
		res.AllReleases = releases(list)
	}
	res.Sources["releases"] = Source{Path: releasesPath, ETag: r.etag}

	tagsPath := "/repos/" + repo + "/tags?per_page=50"
	r, err = c.get(ctx, tagsPath, old("tags", tagsPath))
	if err != nil {
		return nil, err
	}
	res.Rate = r.rate
	if r.notModified && prev != nil {
		res.AllTags = prev.AllTags
		res.NotModified = append(res.NotModified, "tags")
		if r.etag == "" {
			r.etag = old("tags", tagsPath)
		}
	} else {
		var list []ghTag
		if err := json.Unmarshal(r.body, &list); err != nil {
			return nil, err
		}
		for _, t := range list {
			res.AllTags = append(res.AllTags, Tag{Name: t.Name, SHA: t.Commit.SHA, URL: "https://github.com/" + req.Repo + "/releases/tag/" + url.PathEscape(t.Name)})
		}
	}
	res.Sources["tags"] = Source{Path: tagsPath, ETag: r.etag}
	derive(res, req)
	return res, nil
}

func compareData(gc ghCompare) *CompareData {
	cd := &CompareData{Status: gc.Status, AheadBy: gc.AheadBy, BehindBy: gc.BehindBy, Total: gc.TotalCommits}
	for i := len(gc.Commits) - 1; i >= 0; i-- {
		gcm := gc.Commits[i]
		author := gcm.Commit.Author.Name
		if gcm.Author != nil && gcm.Author.Login != "" {
			author = gcm.Author.Login
		}
		date := gcm.Commit.Committer.Date
		if date.IsZero() {
			date = gcm.Commit.Author.Date
		}
		msg := gcm.Commit.Message
		cd.Commits = append(cd.Commits, Commit{SHA: gcm.SHA, Title: firstLine(msg), Author: author, Date: date.UTC(), URL: gcm.URL, Merge: len(gcm.Parents) > 1, Breaks: Breaking(msg)})
	}
	for _, f := range gc.Files {
		cd.Files = append(cd.Files, File{Path: f.Filename, Status: f.Status, Additions: f.Additions, Deletions: f.Deletions, Previous: f.Previous, patch: f.Patch})
	}
	cd.Patches = patchHints(cd.Files)
	return cd
}

func releases(list []ghRelease) []Release {
	out := []Release{}
	for _, r := range list {
		if r.Draft {
			continue
		}
		at := r.CreatedAt
		if r.PublishedAt != nil {
			at = *r.PublishedAt
		}
		body := r.Body
		if len(body) > maxReleaseBody {
			body = strings.ToValidUTF8(body[:maxReleaseBody], "") + "\n…"
		}
		out = append(out, Release{Tag: r.TagName, Name: r.Name, PublishedAt: at.UTC(), URL: r.URL, Body: body, Prerelease: r.Prerelease, Breaks: Breaking(r.Body)})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PublishedAt.After(out[j].PublishedAt) })
	return out
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if len(s) > 200 {
		s = strings.ToValidUTF8(s[:200], "") + "…"
	}
	return s
}

func derive(res *Result, req Request) {
	cd := res.Compare
	if cd == nil {
		cd = &CompareData{}
	}
	res.Status = cd.Status
	res.AheadBy = cd.AheadBy
	res.BehindBy = cd.BehindBy
	res.CompareURL = "https://github.com/" + req.Repo + "/compare/" + req.Base + "..." + req.Branch
	res.Truncated = cd.Total > len(cd.Commits) || len(cd.Files) >= 300
	res.Commits = cd.Commits
	if len(res.Commits) > maxCommits {
		res.Commits = res.Commits[:maxCommits]
	}
	if len(cd.Commits) > 0 {
		res.HeadSHA = cd.Commits[0].SHA
		d := cd.Commits[0].Date
		res.HeadDate = &d
	}
	res.FileCount = len(cd.Files)
	res.Areas = Areas(cd.Files)
	newSHAs := map[string]bool{}
	for _, c := range cd.Commits {
		newSHAs[c.SHA] = true
	}
	res.Tags = []Tag{}
	for _, t := range res.AllTags {
		if newSHAs[t.SHA] && len(res.Tags) < maxTags {
			res.Tags = append(res.Tags, t)
		}
	}
	if len(res.AllTags) > 0 {
		t := res.AllTags[0]
		res.LatestTag = &t
	}
	res.Releases = []Release{}
	newTags := map[string]bool{}
	for _, t := range res.Tags {
		newTags[t.Name] = true
	}
	for _, r := range res.AllReleases {
		if (r.PublishedAt.After(req.BaseDate) || newTags[r.Tag]) && len(res.Releases) < maxReleases {
			res.Releases = append(res.Releases, r)
		}
	}
	if len(res.AllReleases) > 0 {
		r := res.AllReleases[0]
		res.LatestRelease = &r
	}
	res.Warnings = Warnings(cd, res.Releases)
}
