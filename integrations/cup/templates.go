package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/cup/internal/settings"
)

type templateRow struct {
	ID              string            `json:"id"`
	Key             string            `json:"key"`
	Name            map[string]string `json:"name"`
	Description     map[string]string `json:"description"`
	IconURL         string            `json:"iconUrl"`
	Active          bool              `json:"active"`
	Schedule        string            `json:"schedule"`
	SignupMinutes   int               `json:"signupMinutes"`
	AnnounceMinutes int               `json:"announceMinutes"`
	Rules           settings.Rules    `json:"rules"`
	Sort            int               `json:"sort"`
	LastStartsAt    *time.Time        `json:"lastStartsAt"`
	UpdatedBy       string            `json:"updatedBy"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

const templateCols = `id::text, key, name::text, description::text, icon_url, active, schedule, signup_minutes, announce_minutes, rules::text, sort, last_starts_at, updated_by, updated_at`

func scanTemplate(row interface{ Scan(...any) error }) (*templateRow, error) {
	t := &templateRow{}
	var name, desc, rules string
	if err := row.Scan(&t.ID, &t.Key, &name, &desc, &t.IconURL, &t.Active, &t.Schedule, &t.SignupMinutes, &t.AnnounceMinutes, &rules, &t.Sort, &t.LastStartsAt, &t.UpdatedBy, &t.UpdatedAt); err != nil {
		return nil, err
	}
	t.Name, t.Description = textMap(name), textMap(desc)
	t.Rules = settings.DefaultRules()
	_ = json.Unmarshal([]byte(rules), &t.Rules)
	t.Rules.Normalize()
	return t, nil
}

func (s *service) templates(ctx context.Context, activeOnly bool) ([]*templateRow, error) {
	q := `SELECT ` + templateCols + ` FROM cup_template`
	if activeOnly {
		q += ` WHERE active`
	}
	rows, err := s.db.QueryContext(ctx, q+` ORDER BY sort, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*templateRow{}
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
