package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/buildinfo"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
	"github.com/heroiclabs/nakama/v3/internal/upstream"
)

const (
	updatesLease       = "base.updates.check"
	updatesLeaseTTL    = 90 * time.Second
	updatesTick        = 5 * time.Second
	updatesErrorRetry  = 30 * time.Minute
	updatesCheckLimit  = 5
	updatesCheckWindow = 10 * time.Minute
	updatesAPIEnv      = "BASE_UPDATES_API_URL"
	updatesDefaultRepo = "heroiclabs/nakama"
	updatesDefaultRef  = "master"
	updatesForkRepo    = "copilot-rast/nakama-farsi"
	secretMask         = "********"
)

var tokenPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{20,255}$`)

type updatesConfig struct {
	key  []byte
	api  string
	http *http.Client
	info func() buildinfo.Info
}

type updateSettings struct {
	Enabled       bool      `json:"enabled"`
	Provider      string    `json:"provider"`
	Repo          string    `json:"repo"`
	Branch        string    `json:"branch"`
	IntervalHours int       `json:"intervalHours"`
	Token         string    `json:"token"`
	HasToken      bool      `json:"hasToken"`
	Saved         bool      `json:"saved"`
	UpdatedBy     string    `json:"updatedBy"`
	UpdatedAt     time.Time `json:"updatedAt"`
	sealed        string
}

type updateState struct {
	Repo        string     `json:"repo"`
	Branch      string     `json:"branch"`
	Base        string     `json:"base"`
	Status      string     `json:"status"`
	Error       string     `json:"error"`
	Checks      int64      `json:"checks"`
	CheckedAt   *time.Time `json:"checkedAt"`
	LastOKAt    *time.Time `json:"lastOkAt"`
	NextAt      time.Time  `json:"nextAt"`
	RequestedAt *time.Time `json:"requestedAt"`
	RequestedBy string     `json:"requestedBy"`
	CheckedBy   string     `json:"checkedBy"`
	Queued      bool       `json:"queued"`
	result      *upstream.Result
}

func (s *service) configureUpdates(env map[string]string) {
	s.updates = updatesConfig{api: strings.TrimSpace(env[updatesAPIEnv]), info: buildinfo.Get}
	if s.updates.api != "" {
		u, err := url.Parse(s.updates.api)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			s.logger.Warn("Base updates: %s is not an http(s) URL, using GitHub", updatesAPIEnv)
			s.updates.api = ""
		}
	}
	if secret, err := secretbox.Secret(env); err == nil {
		if key, err := secretbox.Base.DeriveKey(secret, secretbox.BaseUpdatesInfo); err == nil {
			s.updates.key = key
		}
	}
}

func (s *service) buildInfo() buildinfo.Info {
	if s.updates.info != nil {
		return s.updates.info()
	}
	return buildinfo.Get()
}

func (s *service) defaultUpdateSettings() updateSettings {
	out := updateSettings{Enabled: true, Provider: "github", Repo: updatesDefaultRepo, Branch: updatesDefaultRef, IntervalHours: 6}
	if p := s.buildInfo().Parent; p != nil {
		if repo, ok := upstream.NormalizeRepo(p.Repo); ok && upstream.ValidBranch(p.Branch) {
			out.Repo, out.Branch = repo, p.Branch
		}
	}
	return out
}

func (s *service) loadUpdateSettings(ctx context.Context, q queryer) (updateSettings, error) {
	out := s.defaultUpdateSettings()
	err := q.QueryRowContext(ctx, `SELECT enabled, provider, repo, branch, interval_hours, token_sealed, updated_by, updated_at FROM base_updates_settings WHERE id=1`).
		Scan(&out.Enabled, &out.Provider, &out.Repo, &out.Branch, &out.IntervalHours, &out.sealed, &out.UpdatedBy, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	out.Saved = true
	out.HasToken = out.sealed != ""
	return out, nil
}

func (u updateSettings) view() updateSettings {
	u.Token = ""
	if u.HasToken {
		u.Token = secretMask
	}
	return u
}

func (s *service) loadUpdateState(ctx context.Context, q queryer, withResult bool) (updateState, error) {
	var st updateState
	var raw []byte
	cols := `repo, branch, base, status, error, checks, checked_at, last_ok_at, next_at, requested_at, requested_by, checked_by, NULL::jsonb`
	if withResult {
		cols = strings.Replace(cols, "NULL::jsonb", "result", 1)
	}
	err := q.QueryRowContext(ctx, `SELECT `+cols+` FROM base_updates_state WHERE id=1`).
		Scan(&st.Repo, &st.Branch, &st.Base, &st.Status, &st.Error, &st.Checks, &st.CheckedAt, &st.LastOKAt, &st.NextAt, &st.RequestedAt, &st.RequestedBy, &st.CheckedBy, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return updateState{Status: "never", NextAt: s.now()}, nil
	}
	if err != nil {
		return st, err
	}
	st.Queued = st.RequestedAt != nil && (st.CheckedAt == nil || st.RequestedAt.After(*st.CheckedAt))
	if len(raw) > 0 {
		var r upstream.Result
		if json.Unmarshal(raw, &r) == nil {
			st.result = &r
		}
	}
	return st, nil
}

func baseOf(info buildinfo.Info) string {
	if info.Parent == nil {
		return ""
	}
	return info.Parent.Commit
}

func (st updateState) due(cfg updateSettings, base string, now time.Time) bool {
	if !cfg.Enabled {
		return false
	}
	if st.Status == "rate_limited" && st.NextAt.After(now) {
		return false
	}
	if st.Queued || st.Repo != cfg.Repo || st.Branch != cfg.Branch || st.Base != base {
		return true
	}
	return !st.NextAt.After(now)
}

func publicResult(r *upstream.Result) *upstream.Result {
	if r == nil {
		return nil
	}
	c := *r
	c.Sources = nil
	c.AllReleases = nil
	c.AllTags = nil
	c.Compare = nil
	return &c
}

func (s *service) updatesLoop(ctx context.Context) {
	ticker := time.NewTicker(updatesTick)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			_ = clusterkit.ReleaseLease(context.Background(), s.db, updatesLease, holder)
			return
		case <-ticker.C:
		}
		if _, err := s.updatesTick(ctx, holder); err != nil && ctx.Err() == nil {
			s.logger.Warn("Base updates check failed: %v", err)
		}
	}
}

func (s *service) updatesTick(ctx context.Context, holder string) (bool, error) {
	cfg, err := s.loadUpdateSettings(ctx, s.db)
	if err != nil {
		return false, err
	}
	st, err := s.loadUpdateState(ctx, s.db, false)
	if err != nil {
		return false, err
	}
	base := baseOf(s.buildInfo())
	if !st.due(cfg, base, s.now()) {
		return false, nil
	}
	ok, err := clusterkit.AcquireLease(ctx, s.db, updatesLease, holder, updatesLeaseTTL)
	if err != nil || !ok {
		return false, err
	}
	if st, err = s.loadUpdateState(ctx, s.db, true); err != nil {
		return false, err
	}
	if !st.due(cfg, base, s.now()) {
		return false, nil
	}
	return true, s.runUpdateCheck(ctx, holder, cfg, st)
}

func (s *service) runUpdateCheck(ctx context.Context, holder string, cfg updateSettings, st updateState) error {
	now := s.now().UTC()
	interval := time.Duration(cfg.IntervalHours) * time.Hour
	info := s.buildInfo()
	next := now.Add(interval)
	status, msg := "ok", ""
	var res *upstream.Result
	base := ""
	parent := info.Parent
	switch {
	case parent == nil:
		status, msg = "no_base", "this build has no base commit; build with build/buildinfo.sh"
	case !strings.EqualFold(parent.Repo, cfg.Repo):
		status, msg = "base_mismatch", "the base commit was recorded for "+parent.Repo
		base = parent.Commit
	default:
		base = parent.Commit
		var token string
		if cfg.sealed != "" && s.updates.key != nil {
			if plain, err := secretbox.Base.Open(s.updates.key, cfg.sealed); err == nil {
				token = string(plain)
			}
		}
		var baseDate time.Time
		if t, err := time.Parse(time.RFC3339, parent.Date); err == nil {
			baseDate = t
		}
		prev := st.result
		if st.Repo != cfg.Repo || st.Branch != cfg.Branch || st.Base != base {
			prev = nil
		}
		client := &upstream.Client{BaseURL: s.updates.api, Token: token, HTTP: s.updates.http, Now: s.now}
		r, err := client.Check(ctx, upstream.Request{Repo: cfg.Repo, Branch: cfg.Branch, Base: base, BaseDate: baseDate}, prev)
		var rl *upstream.RateLimitError
		switch {
		case err == nil:
			res = r
		case errors.As(err, &rl):
			status, msg = "rate_limited", err.Error()
			next = rl.Until
			if min := now.Add(time.Minute); next.Before(min) {
				next = min
			}
		case upstream.IsNotFound(err):
			status, msg = "base_missing", err.Error()
		default:
			status, msg = "error", err.Error()
			if updatesErrorRetry < interval {
				next = now.Add(updatesErrorRetry)
			}
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	ok, err := clusterkit.AcquireLease(ctx, s.db, updatesLease, holder, updatesLeaseTTL)
	if err != nil || !ok {
		return err
	}
	var raw any
	keep := status != "ok" && st.Repo == cfg.Repo && st.Branch == cfg.Branch && st.Base == base
	if res != nil {
		b, err := json.Marshal(res)
		if err != nil {
			return err
		}
		raw = string(b)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO base_updates_state (id, repo, branch, base, status, error, result, checks, checked_at, last_ok_at, next_at, checked_by, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, $6::jsonb, 1, $7, CASE WHEN $4 = 'ok' THEN $7::timestamptz END, $8, $9, now())
		ON CONFLICT (id) DO UPDATE SET repo=EXCLUDED.repo, branch=EXCLUDED.branch, base=EXCLUDED.base, status=EXCLUDED.status, error=EXCLUDED.error,
			result=CASE WHEN EXCLUDED.result IS NOT NULL THEN EXCLUDED.result WHEN $10 THEN base_updates_state.result END,
			checks=base_updates_state.checks+1, checked_at=EXCLUDED.checked_at,
			last_ok_at=COALESCE(EXCLUDED.last_ok_at, CASE WHEN $10 THEN base_updates_state.last_ok_at END),
			next_at=EXCLUDED.next_at, checked_by=EXCLUDED.checked_by, updated_at=now()`,
		cfg.Repo, cfg.Branch, base, status, truncate(msg, 500), raw, now, next, holder, keep)
	return err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.ToValidUTF8(s[:n], "")
}

func (s *service) updatesView(ctx context.Context) (map[string]any, error) {
	cfg, err := s.loadUpdateSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	st, err := s.loadUpdateState(ctx, s.db, true)
	if err != nil {
		return nil, err
	}
	info := s.buildInfo()
	current := st.Repo == cfg.Repo && st.Branch == cfg.Branch && st.Base == baseOf(info)
	return map[string]any{
		"build":    info,
		"settings": cfg.view(),
		"state":    st,
		"current":  current,
		"result":   publicResult(st.result),
		"forkRepo": updatesForkRepo,
	}, nil
}

func (s *service) updatesGet(w http.ResponseWriter, r *http.Request) {
	out, err := s.updatesView(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) updatesSummary(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.loadUpdateSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	st, err := s.loadUpdateState(r.Context(), s.db, true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	out := map[string]any{"enabled": cfg.Enabled, "status": st.Status, "checkedAt": st.CheckedAt, "repo": cfg.Repo, "branch": cfg.Branch, "newCommits": 0, "queued": st.Queued}
	if st.result != nil && st.Repo == cfg.Repo && st.Branch == cfg.Branch && st.Base == baseOf(s.buildInfo()) && st.Status != "base_mismatch" && cfg.Enabled {
		out["newCommits"] = st.result.AheadBy
		out["releases"] = len(st.result.Releases)
		out["warnings"] = len(st.result.Warnings)
		if st.result.LatestRelease != nil {
			out["latestRelease"] = st.result.LatestRelease.Tag
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) updatesCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg, err := s.loadUpdateSettings(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if !cfg.Enabled {
		s.writeError(w, conflict("UPDATES_OFF", "update checks are switched off"))
		return
	}
	d, err := clusterkit.Allow(ctx, s.db, "base:updates:check", updatesCheckLimit, updatesCheckWindow)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if !d.Allowed {
		secs := int(d.RetryAfter.Seconds()) + 1
		w.Header().Set("Retry-After", itoa(secs))
		writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "TOO_MANY_CHECKS", "message": "too many checks, try again later", "retryAfter": secs})
		return
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO base_updates_state (id, requested_at, requested_by) VALUES (1, now(), $1)
		ON CONFLICT (id) DO UPDATE SET requested_at=now(), requested_by=EXCLUDED.requested_by, updated_at=now()`, consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	out, err := s.updatesView(ctx)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, out)
}

func (s *service) updatesSettingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.loadUpdateSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.settingsBody(cfg))
}

func (s *service) settingsBody(cfg updateSettings) map[string]any {
	def := s.defaultUpdateSettings()
	return map[string]any{
		"enabled": cfg.Enabled, "provider": cfg.Provider, "repo": cfg.Repo, "branch": cfg.Branch, "intervalHours": cfg.IntervalHours,
		"token": cfg.view().Token, "hasToken": cfg.HasToken, "saved": cfg.Saved, "updatedBy": cfg.UpdatedBy, "updatedAt": cfg.UpdatedAt,
		"canSeal": s.updates.key != nil, "forkRepo": updatesForkRepo, "defaultRepo": def.Repo, "defaultBranch": def.Branch,
		"branches": map[string][]string{updatesForkRepo: {"main", "stage", "dev"}, updatesDefaultRepo: {"master"}},
	}
}

type updateSettingsInput struct {
	Enabled       *bool   `json:"enabled"`
	Provider      string  `json:"provider"`
	Repo          string  `json:"repo"`
	Branch        string  `json:"branch"`
	IntervalHours *int    `json:"intervalHours"`
	Token         *string `json:"token"`
}

func (s *service) updatesSettingsPut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in updateSettingsInput
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	cur, err := s.loadUpdateSettings(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	next := cur
	if in.Enabled != nil {
		next.Enabled = *in.Enabled
	}
	provider := strings.TrimSpace(in.Provider)
	if provider == "" {
		provider = "github"
	}
	if provider != "github" {
		s.writeError(w, badRequest("UNSUPPORTED_PROVIDER", "provider", "only GitHub parents are supported"))
		return
	}
	next.Provider = provider
	repo, ok := upstream.NormalizeRepo(in.Repo)
	if !ok {
		s.writeError(w, badRequest("INVALID_REPO", "repo", "use owner/name or a github.com URL"))
		return
	}
	next.Repo = repo
	branch := strings.TrimSpace(in.Branch)
	if !upstream.ValidBranch(branch) {
		s.writeError(w, badRequest("INVALID_BRANCH", "branch", "branch name is not valid"))
		return
	}
	next.Branch = branch
	if in.IntervalHours != nil {
		next.IntervalHours = *in.IntervalHours
	}
	if next.IntervalHours < 1 || next.IntervalHours > 168 {
		s.writeError(w, badRequest("INVALID_INTERVAL", "intervalHours", "check every 1 to 168 hours"))
		return
	}
	sealed := cur.sealed
	if in.Token != nil {
		tok := strings.TrimSpace(*in.Token)
		switch {
		case tok == secretMask:
		case tok == "":
			sealed = ""
		case !tokenPattern.MatchString(tok):
			s.writeError(w, badRequest("INVALID_GITHUB_TOKEN", "token", "the token has unexpected characters or length"))
			return
		case s.updates.key == nil:
			s.writeError(w, badRequest("SECRET_KEY_MISSING", "token", secretbox.SecretEnv+" is not set, so a token cannot be stored"))
			return
		default:
			if sealed, err = secretbox.Base.Seal(s.updates.key, []byte(tok)); err != nil {
				s.writeError(w, err)
				return
			}
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO base_updates_settings (id, enabled, provider, repo, branch, interval_hours, token_sealed, updated_by, updated_at)
		VALUES (1,$1,$2,$3,$4,$5,$6,$7,now())
		ON CONFLICT (id) DO UPDATE SET enabled=EXCLUDED.enabled, provider=EXCLUDED.provider, repo=EXCLUDED.repo, branch=EXCLUDED.branch,
			interval_hours=EXCLUDED.interval_hours, token_sealed=EXCLUDED.token_sealed, updated_by=EXCLUDED.updated_by, updated_at=now()`,
		next.Enabled, next.Provider, next.Repo, next.Branch, next.IntervalHours, sealed, consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	changed := next.Repo != cur.Repo || next.Branch != cur.Branch || sealed != cur.sealed || (next.Enabled && !cur.Enabled)
	if changed {
		if _, err = tx.ExecContext(ctx, `INSERT INTO base_updates_state (id, requested_at, requested_by) VALUES (1, now(), $1)
			ON CONFLICT (id) DO UPDATE SET requested_at=now(), requested_by=EXCLUDED.requested_by, updated_at=now()`, consoleUser(r)); err != nil {
			s.writeError(w, err)
			return
		}
	} else if next.IntervalHours != cur.IntervalHours {
		if _, err = tx.ExecContext(ctx, `UPDATE base_updates_state SET next_at = COALESCE(checked_at, now()) + $1::INT * INTERVAL '1 hour', updated_at=now() WHERE id=1 AND status='ok'`, next.IntervalHours); err != nil {
			s.writeError(w, err)
			return
		}
	}
	if err = tx.Commit(); err != nil {
		s.writeError(w, err)
		return
	}
	saved, err := s.loadUpdateSettings(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.settingsBody(saved))
}
