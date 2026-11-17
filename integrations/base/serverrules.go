package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/apidocs"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/rpcgate"
	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
	"google.golang.org/grpc/codes"
)

const (
	CodeMaintenance          = "MAINTENANCE"
	CodeStorageWriteBlocked  = "STORAGE_WRITE_BLOCKED"
	CodeStorageDeleteBlocked = "STORAGE_DELETE_BLOCKED"
	CodeAccountDeleteOff     = "ACCOUNT_DELETE_DISABLED"
	CodeConfirmRequired      = "CONFIRM_REQUIRED"
	CodeNoPendingDeletion    = "NO_PENDING_DELETION"
	accountDeleteLease       = "base-account-delete"
)

type maintenanceRules struct {
	Enabled   bool           `json:"enabled"`
	Message   localized.Text `json:"message"`
	Until     *time.Time     `json:"until"`
	AllowRpcs []string       `json:"allowRpcs"`
	AllowTags []string       `json:"allowTagIds"`
}

type storageRules struct {
	BlockWrites      bool     `json:"blockWrites"`
	BlockDeletes     bool     `json:"blockDeletes"`
	AllowCollections []string `json:"allowCollections"`
}

type deleteRules struct {
	Enabled   bool   `json:"enabled"`
	Mode      string `json:"mode"`
	GraceDays int    `json:"graceDays"`
}

type serverRules struct {
	Maintenance   maintenanceRules `json:"maintenance"`
	Storage       storageRules     `json:"storage"`
	AccountDelete deleteRules      `json:"accountDelete"`
	Docs          apidocs.Settings `json:"docs"`
}

func defaultServerRules() serverRules {
	return serverRules{
		Maintenance:   maintenanceRules{Message: localized.Text{}, AllowRpcs: []string{}, AllowTags: []string{}},
		Storage:       storageRules{AllowCollections: []string{}},
		AccountDelete: deleteRules{Enabled: true, Mode: "grace", GraceDays: 14},
		Docs:          apidocs.DefaultSettings(),
	}
}

var serverRulesStore = settingsrow.New("base_server_rules", defaultServerRules)

var maintenanceAlwaysOpen = map[string]bool{"base_app_version": true, "base_server_status": true, "base_languages": true, "base_set_language": true, "base_brand": true}

var maintenanceMessages = localized.Text{"fa": "بازی در حال به‌روزرسانی است. کمی بعد دوباره امتحان کن.", "en": "The game is under maintenance. Please try again soon.", "ar": "اللعبة قيد الصيانة. حاول مرة أخرى بعد قليل."}

func (r serverRules) maintenanceOn(now time.Time) bool {
	return r.Maintenance.Enabled && (r.Maintenance.Until == nil || now.Before(*r.Maintenance.Until))
}

func (r serverRules) clean() (serverRules, error) {
	r.Maintenance.AllowRpcs = cleanList(r.Maintenance.AllowRpcs, true)
	r.Maintenance.AllowTags = cleanList(r.Maintenance.AllowTags, false)
	r.Storage.AllowCollections = cleanList(r.Storage.AllowCollections, false)
	if r.Maintenance.Message == nil {
		r.Maintenance.Message = localized.Text{}
	}
	for k, v := range r.Maintenance.Message {
		if len([]rune(v)) > 500 {
			return r, badRequest("TEXT_TOO_LONG", "maintenance.message."+k, "at most 500 letters")
		}
	}
	if len(r.Maintenance.AllowRpcs) > 100 || len(r.Maintenance.AllowTags) > 50 || len(r.Storage.AllowCollections) > 100 {
		return r, badRequest("TOO_MANY", "maintenance.allowRpcs", "too many entries")
	}
	switch r.AccountDelete.Mode {
	case "grace", "immediate":
	default:
		return r, badRequest("INVALID_RULE", "accountDelete.mode", "grace or immediate")
	}
	if r.AccountDelete.GraceDays < 1 || r.AccountDelete.GraceDays > 365 {
		return r, badRequest("INVALID_RULE", "accountDelete.graceDays", "between 1 and 365")
	}
	return r, nil
}

func cleanList(in []string, lower bool) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, v := range in {
		v = strings.TrimSpace(v)
		if lower {
			v = strings.ToLower(v)
		}
		if v == "" || len(v) > 128 || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func (s *service) maintenanceView(ctx context.Context, userID, lang string) map[string]any {
	r, _ := serverRulesStore.Load(ctx, s.db)
	on := r.maintenanceOn(time.Now())
	out := map[string]any{"on": on}
	if !on {
		return out
	}
	langs, l, err := localized.ForUser(ctx, s.db, userID, lang)
	text := ""
	if err == nil && langs != nil {
		text = langs.Resolve(r.Maintenance.Message, l)
		if text == "" {
			text = langs.Resolve(maintenanceMessages, l)
		}
	}
	if text == "" {
		text = maintenanceMessages["en"]
	}
	out["message"] = text
	if r.Maintenance.Until != nil {
		out["until"] = r.Maintenance.Until.UTC()
	}
	if userID != "" {
		out["bypass"] = s.maintenanceBypass(ctx, r, userID)
	}
	return out
}

func (s *service) maintenanceBypass(ctx context.Context, r serverRules, userID string) bool {
	if len(r.Maintenance.AllowTags) == 0 || userID == "" {
		return false
	}
	var ok bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_user_tag WHERE user_id=$1::uuid AND tag_id::text = ANY($2::text[]) AND (expires_at IS NULL OR expires_at > now()))`, userID, r.Maintenance.AllowTags).Scan(&ok)
	return ok
}

func (s *service) maintenanceGate(ctx context.Context, rpcID, userID string) (codes.Code, error) {
	r, _ := serverRulesStore.Load(ctx, s.db)
	if !r.maintenanceOn(time.Now()) || maintenanceAlwaysOpen[rpcID] {
		return codes.OK, nil
	}
	for _, id := range r.Maintenance.AllowRpcs {
		if id == rpcID {
			return codes.OK, nil
		}
	}
	if s.maintenanceBypass(ctx, r, userID) {
		return codes.OK, nil
	}
	view := s.maintenanceView(ctx, userID, "")
	body := map[string]any{"code": CodeMaintenance, "status": http.StatusServiceUnavailable, "message": view["message"]}
	if v, ok := view["until"]; ok {
		body["meta"] = map[string]any{"until": v}
	}
	raw, _ := json.Marshal(body)
	return codes.Unavailable, runtime.NewError(string(raw), int(codes.Unavailable))
}

func (s *service) rpcServerStatus(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	r, _ := serverRulesStore.Load(ctx, s.db)
	out := map[string]any{"maintenance": s.maintenanceView(ctx, userID, in.Lang), "accountDelete": map[string]any{"enabled": r.AccountDelete.Enabled, "mode": r.AccountDelete.Mode, "graceDays": r.AccountDelete.GraceDays}}
	var due sql.NullTime
	_ = s.db.QueryRowContext(ctx, `SELECT due_at FROM base_account_deletion WHERE user_id=$1::uuid`, userID).Scan(&due)
	if due.Valid {
		out["deletionDueAt"] = due.Time.UTC()
	}
	return out, nil
}

func (s *service) appVersionWithMaintenance(ctx context.Context, userID, payload string) (any, error) {
	out, err := s.rpcAppVersion(ctx, userID, payload)
	if err != nil {
		return nil, err
	}
	if m, ok := out.(map[string]any); ok {
		lang, _ := m["lang"].(string)
		m["maintenance"] = s.maintenanceView(ctx, userID, lang)
	}
	return out, nil
}

func storageRefusal(code string) error {
	raw, _ := json.Marshal(map[string]any{"code": code, "status": http.StatusForbidden})
	return runtime.NewError(string(raw), int(codes.PermissionDenied))
}

func (r serverRules) collectionAllowed(c string) bool {
	for _, a := range r.Storage.AllowCollections {
		if a == c {
			return true
		}
	}
	return false
}

func (s *service) beforeStorageWrite(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.WriteStorageObjectsRequest) (*api.WriteStorageObjectsRequest, error) {
	r, _ := serverRulesStore.Load(ctx, s.db)
	if !r.Storage.BlockWrites || ctxUser(ctx) == "" {
		return in, nil
	}
	for _, o := range in.GetObjects() {
		if !r.collectionAllowed(o.GetCollection()) {
			return nil, storageRefusal(CodeStorageWriteBlocked)
		}
	}
	return in, nil
}

func (s *service) beforeStorageDelete(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.DeleteStorageObjectsRequest) (*api.DeleteStorageObjectsRequest, error) {
	r, _ := serverRulesStore.Load(ctx, s.db)
	if !r.Storage.BlockDeletes || ctxUser(ctx) == "" {
		return in, nil
	}
	for _, o := range in.GetObjectIds() {
		if !r.collectionAllowed(o.GetCollection()) {
			return nil, storageRefusal(CodeStorageDeleteBlocked)
		}
	}
	return in, nil
}

func (s *service) rpcAccountDelete(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Confirm bool `json:"confirm"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if !in.Confirm {
		return nil, badRequest(CodeConfirmRequired, "confirm", "send confirm: true")
	}
	r, _ := serverRulesStore.Load(ctx, s.db)
	if !r.AccountDelete.Enabled {
		return nil, fail(http.StatusForbidden, CodeAccountDeleteOff, "", "account deletion is turned off")
	}
	if r.AccountDelete.Mode == "immediate" {
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindSignIn, Action: "account_deleted", Actor: "player"})
		if err := s.nk.AccountDeleteId(ctx, userID, true); err != nil {
			return nil, err
		}
		return map[string]any{"deleted": true}, nil
	}
	due := time.Now().Add(time.Duration(r.AccountDelete.GraceDays) * 24 * time.Hour)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO base_account_deletion (user_id, due_at) VALUES ($1,$2) ON CONFLICT (user_id) DO NOTHING`, userID, due); err != nil {
		return nil, err
	}
	var at time.Time
	if err := s.db.QueryRowContext(ctx, `SELECT due_at FROM base_account_deletion WHERE user_id=$1`, userID).Scan(&at); err != nil {
		return nil, err
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindSignIn, Action: "account_delete_requested", Actor: "player", Data: map[string]any{"dueAt": at.UTC()}})
	return map[string]any{"deleted": false, "dueAt": at.UTC(), "graceDays": r.AccountDelete.GraceDays}, nil
}

func (s *service) rpcAccountDeleteCancel(ctx context.Context, userID, payload string) (any, error) {
	if err := decodeStrict(payload, &struct{}{}); err != nil {
		return nil, err
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM base_account_deletion WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, notFound(CodeNoPendingDeletion)
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindSignIn, Action: "account_delete_cancelled", Actor: "player"})
	return map[string]any{"cancelled": true}, nil
}

func (s *service) accountDeleteLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if ok, err := clusterkit.AcquireLease(ctx, s.db, accountDeleteLease, holder, 50*time.Second); err != nil || !ok {
			continue
		}
		s.deleteDueAccounts(ctx)
	}
}

func (s *service) deleteDueAccounts(ctx context.Context) int {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text FROM base_account_deletion WHERE due_at <= now() ORDER BY due_at LIMIT 100`)
	if err != nil {
		return 0
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	n := 0
	for _, id := range ids {
		if err := s.nk.AccountDeleteId(ctx, id, true); err != nil {
			s.logger.Warn("Base account %s not deleted: %v", id, err)
			continue
		}
		_, _ = s.db.ExecContext(ctx, `DELETE FROM base_account_deletion WHERE user_id=$1`, id)
		n++
	}
	return n
}

func (s *service) serverRulesGet(w http.ResponseWriter, r *http.Request) {
	v, meta, err := serverRulesStore.Read(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	raw, _ := json.Marshal(v)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	out["updatedBy"], out["updatedAt"] = meta.UpdatedBy, meta.UpdatedAt
	var pending int
	_ = s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM base_account_deletion`).Scan(&pending)
	out["pendingDeletions"] = pending
	writeJSON(w, http.StatusOK, out)
}

func (s *service) serverRulesPut(w http.ResponseWriter, r *http.Request) {
	in := defaultServerRules()
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	v, err := in.clean()
	if err == nil {
		langs, lerr := localized.Load(r.Context(), s.db)
		if lerr != nil {
			err = lerr
		} else {
			v.Maintenance.Message, err = cleanText(langs, "maintenance.message", v.Maintenance.Message, false, 500)
		}
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := serverRulesStore.Save(r.Context(), s.db, v, consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	apidocs.Invalidate()
	s.serverRulesGet(w, r)
}

func registerServerRules(s *service, initializer runtime.Initializer) error {
	rpcgate.Set(s.maintenanceGate)
	if err := initializer.RegisterBeforeWriteStorageObjects(s.beforeStorageWrite); err != nil {
		return err
	}
	return initializer.RegisterBeforeDeleteStorageObjects(s.beforeStorageDelete)
}
