package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type reportRules struct {
	PerHour         int       `json:"perHour"`
	DuplicateHours  int       `json:"duplicateHours"`
	MinAccountHours int       `json:"minAccountHours"`
	NoteRequired    bool      `json:"noteRequired"`
	TrustedTagIDs   []string  `json:"trustedTagIds"`
	TrustedPerHour  int       `json:"trustedPerHour"`
	UpdatedBy       string    `json:"updatedBy,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func defaultReportRules() reportRules {
	return reportRules{PerHour: playerReportLimit, DuplicateHours: int(playerReportDupWindow.Hours()), TrustedTagIDs: []string{}, TrustedPerHour: 30}
}

func loadReportRules(ctx context.Context, q queryer) (reportRules, error) {
	r := defaultReportRules()
	var tags string
	err := q.QueryRowContext(ctx, `SELECT per_hour, duplicate_hours, min_account_hours, note_required, to_jsonb(trusted_tag_ids)::text, trusted_per_hour, updated_by, updated_at FROM base_report_rules WHERE id=1`).
		Scan(&r.PerHour, &r.DuplicateHours, &r.MinAccountHours, &r.NoteRequired, &tags, &r.TrustedPerHour, &r.UpdatedBy, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultReportRules(), nil
	}
	if err != nil {
		return r, err
	}
	_ = json.Unmarshal([]byte(tags), &r.TrustedTagIDs)
	if r.TrustedTagIDs == nil {
		r.TrustedTagIDs = []string{}
	}
	return r, nil
}

func (s *service) reportRulesGet(w http.ResponseWriter, r *http.Request) {
	rules, err := loadReportRules(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (s *service) reportRulesPut(w http.ResponseWriter, r *http.Request) {
	in := defaultReportRules()
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	switch {
	case in.PerHour < 0 || in.PerHour > 1000:
		s.writeError(w, badRequest("INVALID_RULE", "perHour", "between 0 and 1000"))
		return
	case in.DuplicateHours < 0 || in.DuplicateHours > 720:
		s.writeError(w, badRequest("INVALID_RULE", "duplicateHours", "between 0 and 720"))
		return
	case in.MinAccountHours < 0 || in.MinAccountHours > 720:
		s.writeError(w, badRequest("INVALID_RULE", "minAccountHours", "between 0 and 720"))
		return
	case in.TrustedPerHour < 0 || in.TrustedPerHour > 1000:
		s.writeError(w, badRequest("INVALID_RULE", "trustedPerHour", "between 0 and 1000"))
		return
	}
	tags := []string{}
	if len(in.TrustedTagIDs) > 0 {
		var err error
		if tags, err = s.userTagIDs(r.Context(), s.db, in.TrustedTagIDs); err != nil {
			if app, ok := asApp(err); ok {
				app.Field = "trustedTagIds"
			}
			s.writeError(w, err)
			return
		}
	}
	_, err := s.db.ExecContext(r.Context(), `INSERT INTO base_report_rules (id, per_hour, duplicate_hours, min_account_hours, note_required, trusted_tag_ids, trusted_per_hour, updated_by, updated_at)
		VALUES (1,$1,$2,$3,$4,$5,$6,$7,now())
		ON CONFLICT (id) DO UPDATE SET per_hour=EXCLUDED.per_hour, duplicate_hours=EXCLUDED.duplicate_hours, min_account_hours=EXCLUDED.min_account_hours,
		note_required=EXCLUDED.note_required, trusted_tag_ids=EXCLUDED.trusted_tag_ids, trusted_per_hour=EXCLUDED.trusted_per_hour, updated_by=EXCLUDED.updated_by, updated_at=now()`,
		in.PerHour, in.DuplicateHours, in.MinAccountHours, in.NoteRequired, tags, in.TrustedPerHour, consoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.reportRulesGet(w, r)
}

func (s *service) hasAnyTag(ctx context.Context, userID string, tagIDs []string) (bool, error) {
	if len(tagIDs) == 0 {
		return false, nil
	}
	var ok bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_user_tag WHERE user_id=$1::uuid AND tag_id::text = ANY($2::text[]))`, userID, tagIDs).Scan(&ok)
	return ok, err
}

func (s *service) reportLimitFor(ctx context.Context, rules reportRules, userID string) (int, bool, error) {
	trusted, err := s.hasAnyTag(ctx, userID, rules.TrustedTagIDs)
	if err != nil {
		return 0, false, err
	}
	if trusted {
		return rules.TrustedPerHour, true, nil
	}
	return rules.PerHour, false, nil
}

func (s *service) reportsLeft(ctx context.Context, userID string, limit int) (any, error) {
	if limit == 0 {
		return nil, nil
	}
	var used int
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE((SELECT count FROM cluster_rate_limit WHERE key=$1 AND expires_at > now()), 0)`, "player-report:"+userID).Scan(&used)
	return max(limit-used, 0), err
}
