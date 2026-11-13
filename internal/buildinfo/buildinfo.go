package buildinfo

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

var (
	Commit   = ""
	Branch   = ""
	Date     = ""
	Dirty    = ""
	OwnAhead = ""
	Upstream = ""
)

type Parent struct {
	Provider   string `json:"provider"`
	Repo       string `json:"repo"`
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	Date       string `json:"date"`
	Subject    string `json:"subject,omitempty"`
	RecordedAt string `json:"recordedAt,omitempty"`
}

type Info struct {
	Commit   string  `json:"commit"`
	Branch   string  `json:"branch"`
	Date     string  `json:"date"`
	Dirty    bool    `json:"dirty"`
	OwnAhead int     `json:"ownAhead"`
	Known    bool    `json:"known"`
	Parent   *Parent `json:"parent"`
}

func Get() Info {
	return From(Commit, Branch, Date, Dirty, OwnAhead, Upstream)
}

func From(commit, branch, date, dirty, ahead, upstream string) Info {
	info := Info{
		Commit: strings.TrimSpace(commit),
		Branch: strings.TrimSpace(branch),
		Dirty:  dirty == "true" || dirty == "1",
	}
	if t, err := time.Parse(time.RFC3339, strings.TrimSpace(date)); err == nil {
		info.Date = t.UTC().Format(time.RFC3339)
	}
	if n, err := strconv.Atoi(strings.TrimSpace(ahead)); err == nil && n >= 0 {
		info.OwnAhead = n
	}
	info.Known = info.Commit != ""
	info.Parent = ParseParent(upstream)
	return info
}

func ParseParent(text string) *Parent {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	raw := []byte(text)
	if !strings.HasPrefix(text, "{") {
		decoded, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return nil
		}
		raw = decoded
	}
	var p Parent
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil
	}
	p.Repo = strings.TrimSpace(p.Repo)
	p.Commit = strings.ToLower(strings.TrimSpace(p.Commit))
	if p.Repo == "" || p.Commit == "" {
		return nil
	}
	if p.Provider == "" {
		p.Provider = "github"
	}
	return &p
}
