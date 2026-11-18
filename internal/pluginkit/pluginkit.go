package pluginkit

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/heroiclabs/nakama-common/runtime"
)

const ConsoleUserHeader = "X-Nakama-Console-Username"

type ConsoleError struct {
	Status  int
	Code    string
	Field   string
	Message string
	Meta    map[string]any
}

func (e *ConsoleError) Error() string { return e.Code }

func Bad(code, field, msg string) *ConsoleError {
	return &ConsoleError{Status: http.StatusBadRequest, Code: code, Field: field, Message: msg}
}

func Missing(code string) *ConsoleError {
	return &ConsoleError{Status: http.StatusNotFound, Code: code}
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(body)
}

func WriteError(w http.ResponseWriter, err error) bool {
	var ce *ConsoleError
	if !errors.As(err, &ce) {
		return false
	}
	msg := ce.Message
	if msg == "" {
		msg = ce.Code
	}
	body := map[string]any{"error": ce.Code, "message": msg, "field": ce.Field}
	if len(ce.Meta) > 0 {
		body["meta"] = ce.Meta
	}
	WriteJSON(w, ce.Status, body)
	return true
}

func ReadBody(r *http.Request, out any, max int64) error {
	if max <= 0 {
		max = 1 << 20
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, max+1))
	if err != nil {
		return Bad("INVALID_JSON", "", "cannot read body")
	}
	if int64(len(raw)) > max {
		return &ConsoleError{Status: http.StatusRequestEntityTooLarge, Code: "REQUEST_TOO_LARGE", Message: "body is too large"}
	}
	if strings.TrimSpace(string(raw)) == "" {
		raw = []byte("{}")
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return Bad("INVALID_JSON", "", "request body is not valid JSON: "+err.Error())
	}
	return nil
}

func PathVar(r *http.Request, name string) string {
	if v := r.PathValue(name); v != "" {
		return v
	}
	return mux.Vars(r)[name]
}

func ConsoleUser(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get(ConsoleUserHeader))
}

func Page(r *http.Request, def, max int) (int, int) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > max {
		limit = def
	}
	return page, limit
}

type Args struct{ List []any }

func (a *Args) Add(v any) string {
	a.List = append(a.List, v)
	return "$" + strconv.Itoa(len(a.List))
}

func LikePattern(q string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + r.Replace(strings.TrimSpace(q)) + "%"
}

func InTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

type Queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func WalletApplied(ctx context.Context, q Queryer, userID, grantID string) (bool, error) {
	var found bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM wallet_ledger WHERE user_id=$1 AND metadata->>'grantId'=$2)`, userID, grantID).Scan(&found)
	return found, err
}

type ErrInsufficient struct{ Path string }

func (e *ErrInsufficient) Error() string { return "not enough " + e.Path }

func Debit(ctx context.Context, nk runtime.NakamaModule, q Queryer, userID, grantID string, amounts map[string]int64, meta map[string]any) error {
	changes := map[string]int64{}
	for k, v := range amounts {
		if v != 0 {
			changes[k] = -v
		}
	}
	if len(changes) == 0 {
		return nil
	}
	if tx, ok := q.(*sql.Tx); ok {
		return WalletTx(ctx, tx, userID, grantID, changes, meta)
	}
	applied, err := WalletApplied(ctx, q, userID, grantID)
	if err != nil || applied {
		return err
	}
	md := map[string]any{"grantId": grantID}
	for k, v := range meta {
		md[k] = v
	}
	if _, _, err := nk.MultiUpdate(ctx, nil, nil, nil, []*runtime.WalletUpdate{{UserID: userID, Changeset: changes, Metadata: md}}, true); err != nil {
		var neg *runtime.WalletNegativeError
		if errors.As(err, &neg) {
			return &ErrInsufficient{Path: neg.Path}
		}
		return err
	}
	return nil
}

type TxQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func WalletTx(ctx context.Context, tx TxQueryer, userID, grantID string, changes map[string]int64, meta map[string]any) error {
	clean := map[string]int64{}
	for k, v := range changes {
		if v != 0 {
			clean[k] = v
		}
	}
	if len(clean) == 0 {
		return nil
	}
	var raw []byte
	if err := tx.QueryRowContext(ctx, `SELECT wallet FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&raw); err != nil {
		return err
	}
	applied, err := WalletApplied(ctx, tx, userID, grantID)
	if err != nil || applied {
		return err
	}
	wallet := map[string]int64{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &wallet); err != nil {
			return err
		}
	}
	for k, v := range clean {
		if wallet[k]+v < 0 {
			return &ErrInsufficient{Path: k}
		}
		wallet[k] += v
	}
	md := map[string]any{"grantId": grantID}
	for k, v := range meta {
		md[k] = v
	}
	walletJSON, _ := json.Marshal(wallet)
	changeJSON, _ := json.Marshal(clean)
	metaJSON, _ := json.Marshal(md)
	if _, err := tx.ExecContext(ctx, `UPDATE users SET update_time=now(), wallet=$2 WHERE id=$1`, userID, walletJSON); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO wallet_ledger (id, user_id, changeset, metadata) VALUES (gen_random_uuid(), $1, $2, $3)`, userID, changeJSON, metaJSON)
	return err
}
