package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/configtransfer"
	"github.com/heroiclabs/nakama/v3/internal/handoffkit"
	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

const (
	ctPath              = "/config-transfer"
	ctDefaultMaxFileMB  = 20
	ctMaxFileMBLimit    = 200
	ctDefaultKeepRuns   = 50
	CodeCTFileRequired  = "CONFIG_FILE_REQUIRED"
	CodeCTFileInvalid   = "CONFIG_FILE_INVALID"
	CodeCTFileVersion   = "CONFIG_FILE_VERSION"
	CodeCTFileTooLarge  = "CONFIG_FILE_TOO_LARGE"
	CodeCTExportOff     = "CONFIG_EXPORT_OFF"
	CodeCTImportOff     = "CONFIG_IMPORT_OFF"
	CodeCTAreaUnknown   = "CONFIG_AREA_UNKNOWN"
	CodeCTNothingPicked = "CONFIG_NOTHING_SELECTED"
	CodeCTApplyFailed   = "CONFIG_APPLY_FAILED"
)

var ctFileName = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)

type ctSettings struct {
	ServerName    string `json:"serverName"`
	MaxFileMB     int    `json:"maxFileMB"`
	ExportEnabled bool   `json:"exportEnabled"`
	ImportEnabled bool   `json:"importEnabled"`
	KeepRuns      int    `json:"keepRuns"`
}

func defaultCTSettings() ctSettings {
	return ctSettings{MaxFileMB: ctDefaultMaxFileMB, ExportEnabled: true, ImportEnabled: true, KeepRuns: ctDefaultKeepRuns}
}

var ctSettingsStore = settingsrow.New("base_config_transfer_settings", defaultCTSettings)

func (c ctSettings) clean() (ctSettings, error) {
	c.ServerName = strings.TrimSpace(c.ServerName)
	if len([]rune(c.ServerName)) > 64 {
		return c, badRequest("INVALID_VALUE", "serverName", "at most 64 characters")
	}
	if c.MaxFileMB < 1 || c.MaxFileMB > ctMaxFileMBLimit {
		return c, badRequest("INVALID_VALUE", "maxFileMB", "between 1 and 200")
	}
	if c.KeepRuns < 1 || c.KeepRuns > 500 {
		return c, badRequest("INVALID_VALUE", "keepRuns", "between 1 and 500")
	}
	return c, nil
}

func (s *service) registerConfigTransfer(initializer runtime.Initializer) error {
	if err := configtransfer.Register(configtransfer.BaseAreas()...); err != nil {
		return err
	}
	routes := []struct {
		path, method string
		fn           http.HandlerFunc
	}{
		{consolePrefix + ctPath, http.MethodGet, s.ctOverview},
		{consolePrefix + ctPath + "/settings", http.MethodPut, s.ctSettingsPut},
		{consolePrefix + ctPath + "/export", http.MethodPost, s.ctExport},
		{consolePrefix + ctPath + "/preview", http.MethodPost, s.ctImport(false)},
		{consolePrefix + ctPath + "/apply", http.MethodPost, s.ctImport(true)},
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.fn, r.method); err != nil {
			return err
		}
	}
	return nil
}

func isConsoleAdmin(r *http.Request) bool {
	return r.Header.Get(handoffkit.AdminHeader) == "true"
}

func (s *service) ctServerName(st ctSettings) string {
	if st.ServerName != "" {
		return st.ServerName
	}
	if s.config != nil {
		return s.config.GetName()
	}
	return ""
}

type ctRun struct {
	ID     string                `json:"id"`
	Kind   string                `json:"kind"`
	By     string                `json:"by"`
	Source configtransfer.Source `json:"source"`
	Areas  []string              `json:"areas"`
	Totals map[string]any        `json:"totals"`
	At     string                `json:"at"`
}

func ctRuns(ctx context.Context, db *sql.DB, limit int) ([]ctRun, error) {
	rows, err := db.QueryContext(ctx, `SELECT id::text, kind, run_by, source::text, to_jsonb(areas)::text, totals::text, created_at FROM base_config_transfer_run ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ctRun{}
	for rows.Next() {
		var r ctRun
		var src, areas, totals string
		var at time.Time
		if err := rows.Scan(&r.ID, &r.Kind, &r.By, &src, &areas, &totals, &at); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(src), &r.Source)
		_ = json.Unmarshal([]byte(areas), &r.Areas)
		_ = json.Unmarshal([]byte(totals), &r.Totals)
		r.At = at.UTC().Format(time.RFC3339)
		out = append(out, r)
	}
	return out, rows.Err()
}

func ctRecord(ctx context.Context, q configtransfer.Querier, keep int, kind, by string, src configtransfer.Source, areas []string, totals any) (string, error) {
	id := uuid.Must(uuid.NewV4()).String()
	srcRaw, _ := json.Marshal(src)
	totRaw, _ := json.Marshal(totals)
	if areas == nil {
		areas = []string{}
	}
	areasRaw, _ := json.Marshal(areas)
	if _, err := q.ExecContext(ctx, `INSERT INTO base_config_transfer_run (id, kind, run_by, source, areas, totals) VALUES ($1,$2,$3,$4::jsonb,ARRAY(SELECT jsonb_array_elements_text($5::jsonb)),$6::jsonb)`, id, kind, by, string(srcRaw), string(areasRaw), string(totRaw)); err != nil {
		return "", err
	}
	if keep > 0 {
		if _, err := q.ExecContext(ctx, `DELETE FROM base_config_transfer_run WHERE id IN (SELECT id FROM base_config_transfer_run ORDER BY created_at DESC OFFSET $1)`, keep); err != nil {
			return "", err
		}
	}
	return id, nil
}

func ctAudit(ctx context.Context, q configtransfer.Querier, username, action, message string, meta any) error {
	raw, _ := json.Marshal(meta)
	var userID string
	err := q.QueryRowContext(ctx, `SELECT id::text FROM console_user WHERE username=$1`, username).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) || userID == "" {
		userID = uuid.Nil.String()
	} else if err != nil {
		return err
	}
	var email string
	_ = q.QueryRowContext(ctx, `SELECT email FROM console_user WHERE username=$1`, username).Scan(&email)
	_, err = q.ExecContext(ctx, `INSERT INTO console_audit_log (id, console_user_id, console_username, email, resource, action, metadata, message, create_time) VALUES ($1,$2,$3,$4,'SETTINGS',$5,$6::jsonb,$7,now())`,
		uuid.Must(uuid.NewV4()).String(), userID, username, email, action, string(raw), message)
	return err
}

func (s *service) ctOverview(w http.ResponseWriter, r *http.Request) {
	st, meta, err := ctSettingsStore.Read(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	runs, err := ctRuns(r.Context(), s.db, 20)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"areas": configtransfer.List(), "settings": st, "defaults": defaultCTSettings(), "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt,
		"serverName": s.ctServerName(st), "format": configtransfer.Format, "version": configtransfer.FormatVersion,
		"runs": runs, "admin": isConsoleAdmin(r),
	})
}

func (s *service) ctSettingsPut(w http.ResponseWriter, r *http.Request) {
	if !isConsoleAdmin(r) {
		s.writeError(w, fail(http.StatusForbidden, CodeAdminOnly, "", "only a console admin can change config transfer settings"))
		return
	}
	in := defaultCTSettings()
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	v, err := in.clean()
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := ctSettingsStore.Save(r.Context(), s.db, v, consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	s.ctOverview(w, r)
}

func ctUnknownAreas(list []string) []string {
	var out []string
	for _, k := range list {
		if _, ok := configtransfer.Lookup(k); !ok {
			out = append(out, k)
		}
	}
	return out
}

func (s *service) ctExport(w http.ResponseWriter, r *http.Request) {
	if !isConsoleAdmin(r) {
		s.writeError(w, fail(http.StatusForbidden, CodeAdminOnly, "", "only a console admin can export the config"))
		return
	}
	st, err := ctSettingsStore.Load(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if !st.ExportEnabled {
		s.writeError(w, fail(http.StatusForbidden, CodeCTExportOff, "", "export is switched off in the config transfer settings"))
		return
	}
	var in struct {
		Areas []string `json:"areas"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if bad := ctUnknownAreas(in.Areas); len(bad) > 0 {
		s.writeError(w, badRequest(CodeCTAreaUnknown, "areas", "unknown area "+bad[0]))
		return
	}
	by := consoleUser(r)
	src := configtransfer.Source{Name: s.ctServerName(st), By: by}
	if s.config != nil {
		src.Node = s.config.GetName()
	}
	f, err := configtransfer.Export(r.Context(), s.db, configtransfer.ExportOptions{Areas: in.Areas, Source: src, Now: s.now()})
	if err != nil {
		s.writeError(w, err)
		return
	}
	keys := make([]string, 0, len(f.Areas))
	for _, a := range f.Areas {
		keys = append(keys, a.Key)
	}
	if _, err := ctRecord(r.Context(), s.db, st.KeepRuns, "export", by, f.Source, keys, f.Counts); err != nil {
		s.writeError(w, err)
		return
	}
	if err := ctAudit(r.Context(), s.db, by, "EXPORT", "base config transfer export", map[string]any{"areas": keys, "counts": f.Counts}); err != nil && s.logger != nil {
		s.logger.Warn("Config transfer export audit entry not written: %v", err)
	}
	name := strings.Trim(ctFileName.ReplaceAllString(f.Source.Name, "-"), "-")
	if name == "" {
		name = "server"
	}
	w.Header().Set("Content-Disposition", `attachment; filename="config-`+name+`-`+s.now().UTC().Format("20060102-150405")+`.json"`)
	writeJSON(w, http.StatusOK, f)
}

type ctImportIn struct {
	File      json.RawMessage     `json:"file"`
	Areas     []string            `json:"areas"`
	Rows      map[string][]string `json:"rows"`
	SwitchOff map[string]bool     `json:"switchOff"`
}

func (s *service) ctReadImport(w http.ResponseWriter, r *http.Request, max int64) (*ctImportIn, []byte, error) {
	if r.ContentLength > max+(1<<20) {
		return nil, nil, fail(http.StatusRequestEntityTooLarge, CodeCTFileTooLarge, "file", "the file is larger than the limit in the settings")
	}
	r.Body = http.MaxBytesReader(w, r.Body, max+(1<<20))
	in := &ctImportIn{}
	var raw []byte
	if strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "multipart/") {
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			var mbe *http.MaxBytesError
			if errors.As(err, &mbe) {
				return nil, nil, fail(http.StatusRequestEntityTooLarge, CodeCTFileTooLarge, "file", "the file is larger than the limit in the settings")
			}
			return nil, nil, badRequest(CodeCTFileInvalid, "file", "cannot read the upload")
		}
		defer func() {
			if r.MultipartForm != nil {
				_ = r.MultipartForm.RemoveAll()
			}
		}()
		if opts := r.FormValue("options"); strings.TrimSpace(opts) != "" {
			if err := json.Unmarshal([]byte(opts), in); err != nil {
				return nil, nil, badRequest("INVALID_JSON", "options", "options are not valid JSON")
			}
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			return nil, nil, badRequest(CodeCTFileRequired, "file", "attach the exported file")
		}
		defer file.Close()
		raw, err = io.ReadAll(io.LimitReader(file, max+1))
		if err != nil {
			return nil, nil, badRequest(CodeCTFileInvalid, "file", "cannot read the upload")
		}
	} else {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			var mbe *http.MaxBytesError
			if errors.As(err, &mbe) {
				return nil, nil, fail(http.StatusRequestEntityTooLarge, CodeCTFileTooLarge, "file", "the file is larger than the limit in the settings")
			}
			return nil, nil, badRequest(CodeCTFileInvalid, "file", "cannot read the body")
		}
		if err := json.Unmarshal(body, in); err != nil {
			return nil, nil, badRequest("INVALID_JSON", "", "body is not valid JSON")
		}
		raw = in.File
	}
	if len(strings.TrimSpace(string(raw))) == 0 || string(raw) == "null" {
		return nil, nil, badRequest(CodeCTFileRequired, "file", "attach the exported file")
	}
	if int64(len(raw)) > max {
		return nil, nil, fail(http.StatusRequestEntityTooLarge, CodeCTFileTooLarge, "file", "the file is larger than the limit in the settings")
	}
	return in, raw, nil
}

func (s *service) ctImport(commit bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isConsoleAdmin(r) {
			s.writeError(w, fail(http.StatusForbidden, CodeAdminOnly, "", "only a console admin can import a config file"))
			return
		}
		st, err := ctSettingsStore.Load(r.Context(), s.db)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if !st.ImportEnabled {
			s.writeError(w, fail(http.StatusForbidden, CodeCTImportOff, "", "import is switched off in the config transfer settings"))
			return
		}
		in, raw, err := s.ctReadImport(w, r, int64(st.MaxFileMB)<<20)
		if err != nil {
			s.writeError(w, err)
			return
		}
		f, err := configtransfer.Decode(raw)
		if errors.Is(err, configtransfer.ErrVersion) {
			s.writeError(w, badRequest(CodeCTFileVersion, "file", "this file was made by a newer server"))
			return
		}
		if err != nil {
			s.writeError(w, badRequest(CodeCTFileInvalid, "file", "this is not a config transfer file"))
			return
		}
		if bad := ctUnknownAreas(in.Areas); len(bad) > 0 {
			s.writeError(w, badRequest(CodeCTAreaUnknown, "areas", "unknown area "+bad[0]))
			return
		}
		if commit && in.Areas != nil && len(in.Areas) == 0 {
			s.writeError(w, badRequest(CodeCTNothingPicked, "areas", "pick at least one area"))
			return
		}
		by := consoleUser(r)
		var runID string
		opts := configtransfer.Options{Areas: in.Areas, Rows: in.Rows, SwitchOff: in.SwitchOff, Commit: commit, By: "config-transfer:" + by}
		if commit {
			opts.Audit = func(ctx context.Context, q configtransfer.Querier, res *configtransfer.Result) error {
				var picked []string
				areas := map[string]configtransfer.Counts{}
				for _, a := range res.Areas {
					if a.Selected {
						picked = append(picked, a.Key)
						areas[a.Key] = a.Counts
					}
				}
				id, err := ctRecord(ctx, q, st.KeepRuns, "import", by, res.Source, picked, res.Totals)
				if err != nil {
					return err
				}
				runID = id
				return ctAudit(ctx, q, by, "IMPORT", "base config transfer apply", map[string]any{"runId": id, "source": res.Source, "totals": res.Totals, "areas": areas})
			}
		}
		res, err := configtransfer.Apply(r.Context(), s.db, f, opts)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if commit && !res.Applied {
			writeJSON(w, http.StatusConflict, map[string]any{"error": CodeCTApplyFailed, "message": "some rows could not be written; nothing was changed", "result": res})
			return
		}
		if commit && s.logger != nil {
			s.logger.Info("Config transfer by %s from %q: %d new, %d changed, %d switched off", by, res.Source.Name, res.Totals.New, res.Totals.Changed, res.Totals.SwitchedOff)
		}
		writeJSON(w, http.StatusOK, map[string]any{"result": res, "runId": runID})
	}
}
