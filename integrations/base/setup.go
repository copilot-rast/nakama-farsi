package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

var setupSteps = []string{"brand", "languages", "currency", "payment", "sms", "email"}

var setupStatuses = map[string]bool{"pending": true, "in_progress": true, "completed": true, "skipped": true}

var stepStatuses = map[string]bool{"done": true, "skipped": true, "pending": true}

type setupState struct {
	Status        string            `json:"status"`
	Steps         map[string]string `json:"steps"`
	FirstOpenedBy string            `json:"firstOpenedBy"`
	FirstOpenedAt *time.Time        `json:"firstOpenedAt"`
	CompletedBy   string            `json:"completedBy"`
	CompletedAt   *time.Time        `json:"completedAt"`
	UpdatedBy     string            `json:"updatedBy"`
	UpdatedAt     time.Time         `json:"updatedAt"`
	NextStep      string            `json:"nextStep"`
	StepOrder     []string          `json:"stepOrder"`
}

func nextSetupStep(steps map[string]string) string {
	for _, k := range setupSteps {
		if v := steps[k]; v != "done" && v != "skipped" {
			return k
		}
	}
	return "review"
}

func scanSetup(row interface{ Scan(...any) error }) (setupState, error) {
	var st setupState
	var steps string
	var opened, completed sql.NullTime
	if err := row.Scan(&st.Status, &steps, &st.FirstOpenedBy, &opened, &st.CompletedBy, &completed, &st.UpdatedBy, &st.UpdatedAt); err != nil {
		return st, err
	}
	st.Steps = map[string]string{}
	var raw map[string]any
	if err := json.Unmarshal([]byte(steps), &raw); err == nil {
		for _, k := range setupSteps {
			if v, ok := raw[k].(string); ok && (v == "done" || v == "skipped") {
				st.Steps[k] = v
			}
		}
	}
	if opened.Valid {
		t := opened.Time
		st.FirstOpenedAt = &t
	}
	if completed.Valid {
		t := completed.Time
		st.CompletedAt = &t
	}
	st.NextStep = nextSetupStep(st.Steps)
	st.StepOrder = setupSteps
	return st, nil
}

const setupColumns = `status, steps::text, first_opened_by, first_opened_at, completed_by, completed_at, updated_by, updated_at`

func (s *service) ensureSetupRow(ctx context.Context, q interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}) error {
	_, err := q.ExecContext(ctx, `INSERT INTO base_setup (id) VALUES (1) ON CONFLICT DO NOTHING`)
	return err
}

func (s *service) loadSetup(ctx context.Context, q queryer) (setupState, error) {
	return scanSetup(q.QueryRowContext(ctx, `SELECT `+setupColumns+` FROM base_setup WHERE id=1`))
}

func (s *service) setupGet(w http.ResponseWriter, r *http.Request) {
	if err := s.ensureSetupRow(r.Context(), s.db); err != nil {
		s.writeError(w, err)
		return
	}
	st, err := s.loadSetup(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *service) claimSetup(ctx context.Context, by string) (bool, setupState, error) {
	if err := s.ensureSetupRow(ctx, s.db); err != nil {
		return false, setupState{}, err
	}
	st, err := scanSetup(s.db.QueryRowContext(ctx, `UPDATE base_setup SET status='in_progress', first_opened_by=$1, first_opened_at=now(), updated_by=$1, updated_at=now()
		WHERE id=1 AND status='pending' RETURNING `+setupColumns, by))
	if err == nil {
		return true, st, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, st, err
	}
	st, err = s.loadSetup(ctx, s.db)
	return false, st, err
}

func (s *service) setupClaim(w http.ResponseWriter, r *http.Request) {
	claimed, st, err := s.claimSetup(r.Context(), consoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"claimed": claimed, "setup": st})
}

type setupInput struct {
	Status *string           `json:"status"`
	Steps  map[string]string `json:"steps"`
}

func applySetup(cur setupState, in setupInput) (setupState, error) {
	next := cur
	next.Steps = map[string]string{}
	for k, v := range cur.Steps {
		next.Steps[k] = v
	}
	for k, v := range in.Steps {
		valid := false
		for _, s := range setupSteps {
			if s == k {
				valid = true
			}
		}
		if !valid {
			return cur, badRequest("INVALID_STEP", "steps", "unknown setup step "+k)
		}
		if !stepStatuses[v] {
			return cur, badRequest("INVALID_STEP_STATUS", "steps", "step status must be done, skipped or pending")
		}
		if v == "pending" {
			delete(next.Steps, k)
		} else {
			next.Steps[k] = v
		}
	}
	if in.Status != nil {
		status := *in.Status
		if !setupStatuses[status] {
			return cur, badRequest("INVALID_STATUS", "status", "status must be pending, in_progress, completed or skipped")
		}
		switch {
		case status == "pending":
			next.Status = "pending"
			next.FirstOpenedBy, next.FirstOpenedAt = "", nil
			next.CompletedBy, next.CompletedAt = "", nil
			next.Steps = map[string]string{}
		case status == "skipped" && cur.Status == "completed":
		case status == "in_progress" && cur.Status == "completed":
		default:
			next.Status = status
		}
	}
	next.NextStep = nextSetupStep(next.Steps)
	return next, nil
}

func (s *service) setupPut(w http.ResponseWriter, r *http.Request) {
	var in setupInput
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	by := consoleUser(r)
	var out setupState
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		if err := s.ensureSetupRow(r.Context(), tx); err != nil {
			return err
		}
		cur, err := scanSetup(tx.QueryRowContext(r.Context(), `SELECT `+setupColumns+` FROM base_setup WHERE id=1 FOR UPDATE`))
		if err != nil {
			return err
		}
		next, err := applySetup(cur, in)
		if err != nil {
			return err
		}
		steps, _ := json.Marshal(next.Steps)
		out, err = scanSetup(tx.QueryRowContext(r.Context(), `UPDATE base_setup SET status=$1, steps=$2,
			first_opened_by = CASE WHEN $1='pending' THEN '' WHEN first_opened_at IS NULL AND $1<>'pending' THEN $3 ELSE first_opened_by END,
			first_opened_at = CASE WHEN $1='pending' THEN NULL WHEN first_opened_at IS NULL AND $1<>'pending' THEN now() ELSE first_opened_at END,
			completed_by = CASE WHEN $1='pending' THEN '' WHEN $1='completed' AND status<>'completed' THEN $3 ELSE completed_by END,
			completed_at = CASE WHEN $1='pending' THEN NULL WHEN $1='completed' AND status<>'completed' THEN now() ELSE completed_at END,
			updated_by=$3, updated_at=now()
			WHERE id=1 RETURNING `+setupColumns, next.Status, string(steps), by))
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
