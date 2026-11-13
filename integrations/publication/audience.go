package main

import (
	"context"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

type audienceRule struct {
	Audience     []string
	Exclude      []string
	LevelTable   string
	MinLevel     *int
	MaxLevel     *int
	Platforms    []string
	MinVersion   string
	MaxVersion   string
	Countries    []string
	Languages    []string
	PlayerAge    string
	NewDays      int
	InactiveDays int
}

type audienceInput struct {
	AudienceTagIDs []string `json:"audienceTagIds"`
	ExcludeTagIDs  []string `json:"excludeTagIds"`
	LevelTableID   string   `json:"levelTableId"`
	MinLevel       *int     `json:"minLevel"`
	MaxLevel       *int     `json:"maxLevel"`
	Platforms      []string `json:"platforms"`
	MinVersion     string   `json:"minVersion"`
	MaxVersion     string   `json:"maxVersion"`
	Countries      []string `json:"countries"`
	Languages      []string `json:"languages"`
	PlayerAge      string   `json:"playerAge"`
	NewDays        *int     `json:"newDays"`
	InactiveDays   *int     `json:"inactiveDays,omitempty"`
}

func (a audienceRule) empty() bool {
	return len(a.Audience) == 0 && len(a.Exclude) == 0 && a.MinLevel == nil && a.MaxLevel == nil && len(a.Platforms) == 0 && a.MinVersion == "" && a.MaxVersion == "" &&
		len(a.Countries) == 0 && len(a.Languages) == 0 && (a.PlayerAge == "" || a.PlayerAge == "all") && a.InactiveDays == 0
}

func (a audienceRule) input() audienceInput {
	nd := a.NewDays
	out := audienceInput{AudienceTagIDs: a.Audience, ExcludeTagIDs: a.Exclude, LevelTableID: a.LevelTable, MinLevel: a.MinLevel, MaxLevel: a.MaxLevel, Platforms: a.Platforms,
		MinVersion: a.MinVersion, MaxVersion: a.MaxVersion, Countries: a.Countries, Languages: a.Languages, PlayerAge: a.PlayerAge, NewDays: &nd}
	if a.InactiveDays > 0 {
		d := a.InactiveDays
		out.InactiveDays = &d
	}
	return out
}

type txQueryer interface {
	queryer
	rowsQueryer
}

func cleanAudience(ctx context.Context, tx txQueryer, in audienceInput) (audienceRule, error) {
	var out audienceRule
	langs, err := localized.Load(ctx, tx)
	if err != nil {
		return out, err
	}
	if out.Audience, err = cleanTagIDs(ctx, tx, "audienceTagIds", in.AudienceTagIDs, "user"); err != nil {
		return out, err
	}
	if out.Exclude, err = cleanTagIDs(ctx, tx, "excludeTagIds", in.ExcludeTagIDs, "user"); err != nil {
		return out, err
	}
	for _, x := range out.Exclude {
		if containsStr(out.Audience, x) {
			return out, badRequest("TAG_BOTH_WAYS", "excludeTagIds", "a tag cannot be both shown to and hidden from")
		}
	}
	levelTable, err := cleanUUID("levelTableId", in.LevelTableID, false)
	if err != nil {
		return out, err
	}
	if levelTable != nil {
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_level_table WHERE id=$1)`, *levelTable).Scan(&ok); err != nil {
			return out, err
		}
		if !ok {
			return out, badRequest("LEVEL_TABLE_NOT_FOUND", "levelTableId", "level table does not exist")
		}
	}
	for _, p := range []struct {
		f string
		v *int
	}{{"minLevel", in.MinLevel}, {"maxLevel", in.MaxLevel}} {
		if p.v != nil && (*p.v < 0 || *p.v > 100000) {
			return out, badRequest("INVALID_LEVEL", p.f, "between 0 and 100000")
		}
	}
	if in.MinLevel != nil && in.MaxLevel != nil && *in.MinLevel > *in.MaxLevel {
		return out, badRequest("INVALID_RANGE", "maxLevel", "max level is below min level")
	}
	out.MinLevel, out.MaxLevel = in.MinLevel, in.MaxLevel
	if levelTable != nil && (in.MinLevel != nil || in.MaxLevel != nil) {
		out.LevelTable = *levelTable
	}
	if out.Platforms, err = cleanList("platforms", in.Platforms, 10, strings.ToLower, func(v string) bool { return containsStr(bannerPlatforms, v) }, "INVALID_PLATFORM"); err != nil {
		return out, err
	}
	out.MinVersion = strings.TrimPrefix(strings.TrimSpace(latinDigits(in.MinVersion)), "v")
	out.MaxVersion = strings.TrimPrefix(strings.TrimSpace(latinDigits(in.MaxVersion)), "v")
	for _, p := range []struct{ f, v string }{{"minVersion", out.MinVersion}, {"maxVersion", out.MaxVersion}} {
		if p.v != "" {
			if _, ok := bannerVersionParts(p.v); !ok || len(p.v) > 40 {
				return out, badRequest("INVALID_VERSION", p.f, "use a version like 1.4.0")
			}
		}
	}
	if out.MinVersion != "" && out.MaxVersion != "" {
		if c, _ := compareBannerVersions(out.MinVersion, out.MaxVersion); c > 0 {
			return out, badRequest("INVALID_RANGE", "maxVersion", "max version is below min version")
		}
	}
	if out.Countries, err = cleanList("countries", in.Countries, 250, strings.ToUpper, func(v string) bool {
		return len(v) == 2 && v[0] >= 'A' && v[0] <= 'Z' && v[1] >= 'A' && v[1] <= 'Z'
	}, "INVALID_COUNTRY_CODE"); err != nil {
		return out, err
	}
	if out.Languages, err = cleanList("languages", in.Languages, 50, localized.Normalize, langs.Known, "INVALID_LANGUAGE"); err != nil {
		return out, err
	}
	out.PlayerAge = in.PlayerAge
	if out.PlayerAge == "" {
		out.PlayerAge = "all"
	}
	if !containsStr(bannerAges, out.PlayerAge) {
		return out, badRequest("INVALID_PLAYER_AGE", "playerAge", "playerAge must be all, new or returning")
	}
	out.NewDays = 7
	if in.NewDays != nil {
		out.NewDays = *in.NewDays
	}
	if out.NewDays < 1 || out.NewDays > 365 {
		return out, badRequest("INVALID_NEW_DAYS", "newDays", "between 1 and 365 days")
	}
	if in.InactiveDays != nil && *in.InactiveDays != 0 {
		if *in.InactiveDays < 1 || *in.InactiveDays > 365 {
			return out, badRequest("INVALID_DAYS", "inactiveDays", "between 1 and 365 days")
		}
		out.InactiveDays = *in.InactiveDays
	}
	return out, nil
}

func (a audienceRule) match(ctx context.Context, v *viewer) (string, error) {
	if len(a.Audience) > 0 {
		hit := false
		for _, id := range a.Audience {
			if v.Tags[id] {
				hit = true
			}
		}
		if !hit {
			return "tag", nil
		}
	}
	for _, id := range a.Exclude {
		if v.Tags[id] {
			return "excluded", nil
		}
	}
	if len(a.Platforms) > 0 && !containsStr(a.Platforms, v.Platform) {
		return "platform", nil
	}
	if a.MinVersion != "" || a.MaxVersion != "" {
		if v.Version == "" {
			return "version", nil
		}
		if a.MinVersion != "" {
			if c, ok := compareBannerVersions(v.Version, a.MinVersion); !ok || c < 0 {
				return "version", nil
			}
		}
		if a.MaxVersion != "" {
			if c, ok := compareBannerVersions(v.Version, a.MaxVersion); !ok || c > 0 {
				return "version", nil
			}
		}
	}
	if len(a.Countries) > 0 && !containsStr(a.Countries, v.Country) {
		return "country", nil
	}
	if len(a.Languages) > 0 && !containsStr(a.Languages, v.Lang) {
		return "language", nil
	}
	switch a.PlayerAge {
	case "new":
		if v.Now.Sub(v.Created) >= time.Duration(a.NewDays)*24*time.Hour {
			return "new", nil
		}
	case "returning":
		if v.Now.Sub(v.Created) < time.Duration(a.NewDays)*24*time.Hour {
			return "returning", nil
		}
	}
	if a.MinLevel != nil || a.MaxLevel != nil {
		lv, err := v.level(ctx, a.LevelTable)
		if err != nil {
			return "", err
		}
		if a.MinLevel != nil && lv < *a.MinLevel {
			return "level", nil
		}
		if a.MaxLevel != nil && lv > *a.MaxLevel {
			return "level", nil
		}
	}
	return "", nil
}
