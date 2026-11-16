package plugintest

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const DBEnv = "PLUGIN_DISPOSABLE_TEST_DB_URL"

func OpenDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv(DBEnv)
	if dsn == "" {
		t.Skip(DBEnv + " not set; PostgreSQL NOT_RUN")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(cfg.Database, "_disposable") {
		t.Fatalf("database %q is not disposable", cfg.Database)
	}
	if ip := net.ParseIP(cfg.Host); cfg.Host != "localhost" && !strings.HasPrefix(cfg.Host, "/") && (ip == nil || !ip.IsLoopback()) {
		t.Fatalf("database host %q is not local", cfg.Host)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

type Logger struct{}

func (Logger) Debug(string, ...any)                       {}
func (Logger) Info(string, ...any)                        {}
func (Logger) Warn(string, ...any)                        {}
func (Logger) Error(string, ...any)                       {}
func (l Logger) WithField(string, any) runtime.Logger     { return l }
func (l Logger) WithFields(map[string]any) runtime.Logger { return l }
func (Logger) Fields() map[string]any                     { return nil }

type Wallet struct {
	runtime.NakamaModule
	DB    *sql.DB
	mu    sync.Mutex
	Calls int
	Fail  bool
}

func (n *Wallet) MultiUpdate(ctx context.Context, _ []*runtime.AccountUpdate, _ []*runtime.StorageWrite, _ []*runtime.StorageDelete, wallets []*runtime.WalletUpdate, ledger bool) ([]*api.StorageObjectAck, []*runtime.WalletUpdateResult, error) {
	n.mu.Lock()
	n.Calls++
	failing := n.Fail
	n.mu.Unlock()
	if failing {
		return nil, nil, errors.New("wallet service is down")
	}
	tx, err := n.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()
	for _, w := range wallets {
		var raw string
		if err := tx.QueryRowContext(ctx, `SELECT wallet::text FROM users WHERE id=$1 FOR UPDATE`, w.UserID).Scan(&raw); err != nil {
			return nil, nil, err
		}
		wallet := map[string]int64{}
		_ = json.Unmarshal([]byte(raw), &wallet)
		for k, v := range w.Changeset {
			if wallet[k]+v < 0 {
				return nil, nil, &runtime.WalletNegativeError{UserID: w.UserID, Path: k, Current: wallet[k], Amount: v}
			}
			wallet[k] += v
		}
		out, _ := json.Marshal(wallet)
		if _, err := tx.ExecContext(ctx, `UPDATE users SET wallet=$2 WHERE id=$1`, w.UserID, string(out)); err != nil {
			return nil, nil, err
		}
		if ledger {
			cs, _ := json.Marshal(w.Changeset)
			md, _ := json.Marshal(w.Metadata)
			if _, err := tx.ExecContext(ctx, `INSERT INTO wallet_ledger (id, user_id, changeset, metadata) VALUES ($1,$2,$3,$4)`, uuid.Must(uuid.NewV4()).String(), w.UserID, string(cs), string(md)); err != nil {
				return nil, nil, err
			}
		}
	}
	return nil, nil, tx.Commit()
}

func (n *Wallet) Give(t *testing.T, userID string, amounts map[string]int64, meta map[string]any) {
	t.Helper()
	if _, _, err := n.MultiUpdate(context.Background(), nil, nil, nil, []*runtime.WalletUpdate{{UserID: userID, Changeset: amounts, Metadata: meta}}, true); err != nil {
		t.Fatal(err)
	}
}

func Balance(t *testing.T, db *sql.DB, userID string) map[string]int64 {
	t.Helper()
	var raw string
	if err := db.QueryRow(`SELECT wallet::text FROM users WHERE id=$1`, userID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	out := map[string]int64{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func NewUser(t *testing.T, db *sql.DB) string {
	t.Helper()
	id := uuid.Must(uuid.NewV4()).String()
	if _, err := db.Exec(`INSERT INTO users (id, username) VALUES ($1,$2)`, id, "pt"+strings.ReplaceAll(id[:13], "-", "")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, q := range []string{
			`DELETE FROM economy_user_item WHERE user_id=$1`,
			`DELETE FROM ledger_item_entry WHERE user_id=$1`,
			`DELETE FROM inventoryx_user_unlock WHERE user_id=$1`,
			`DELETE FROM inventoryx_user_track WHERE user_id=$1`,
			`DELETE FROM inventoryx_request WHERE user_id=$1`,
			`DELETE FROM publication_phone_account WHERE user_id=$1`,
			`DELETE FROM users WHERE id=$1`,
		} {
			_, _ = db.Exec(q, id)
		}
	})
	return id
}

func Exec(t *testing.T, db *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := db.Exec(q, args...); err != nil {
		t.Fatalf("%v\n%s", err, q)
	}
}

func Key(prefix string) string {
	return prefix + "_" + strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
}
