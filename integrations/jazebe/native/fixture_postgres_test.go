package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
)

// recordingNakama implements the NakamaModule calls the Jazebe module makes
// and records them; every other call panics through the nil embedded value.
type recordingNakama struct {
	runtime.NakamaModule
	mu         sync.Mutex
	present    []string // user IDs StreamUserList reports as connected
	sends      []string
	recipients map[string]int
}

type stubPresence struct {
	runtime.Presence
	user string
}

func (p stubPresence) GetUserId() string { return p.user }

func (n *recordingNakama) StreamSend(_ uint8, subject, subcontext, label, data string, presences []runtime.Presence, _ bool) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.sends = append(n.sends, data)
	if n.recipients == nil {
		n.recipients = map[string]int{}
	}
	for _, p := range presences {
		n.recipients[p.GetUserId()]++
	}
	return nil
}
func (n *recordingNakama) StreamUserList(uint8, string, string, string, bool, bool) ([]runtime.Presence, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]runtime.Presence, 0, len(n.present))
	for _, id := range n.present {
		out = append(out, stubPresence{user: id})
	}
	return out, nil
}
func (n *recordingNakama) StreamUserJoin(uint8, string, string, string, string, string, bool, bool, string) (bool, error) {
	return true, nil
}
func (n *recordingNakama) StreamUserLeave(uint8, string, string, string, string, string) error {
	return nil
}
func (n *recordingNakama) StreamUserKick(uint8, string, string, string, runtime.Presence) error {
	return nil
}
func (n *recordingNakama) StreamClose(uint8, string, string, string) error { return nil }
func (n *recordingNakama) SessionDisconnect(context.Context, string, ...runtime.PresenceReason) error {
	return nil
}
func (n *recordingNakama) SessionLogout(string, string, string) error { return nil }

// jazebeFixture is an isolated brand, planet and user set in a disposable
// database. Every RPC goes through the same wrappers InitModule registers.
type jazebeFixture struct {
	t        *testing.T
	db       *sql.DB
	ctx      context.Context
	nk       *recordingNakama
	logger   *sweepLogger
	rpcs     map[string]jazebeRPC
	users    map[string]string
	operator string
	owner    string
	brandID  string
	planetID string
}

type rpcErr struct {
	Code    int
	Message string
}

func newJazebeFixture(t *testing.T, members ...string) *jazebeFixture {
	t.Helper()
	db := openDisposableTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	f := &jazebeFixture{t: t, db: db, ctx: ctx, nk: &recordingNakama{}, logger: &sweepLogger{},
		rpcs: map[string]jazebeRPC{}, users: map[string]string{}}
	for _, entry := range jazebeRPCEntries() {
		f.rpcs[entry.name] = nakamaRPCError(guardJazebeRPC(entry.fn))
	}
	t.Cleanup(func() {
		f.cleanup()
		cancel()
		db.Close()
	})
	f.operator = f.user("operator")
	f.owner = f.user("owner")
	t.Setenv("JAZEBE_OPERATOR_USER_IDS", f.operator)
	t.Setenv("JAZEBE_SUPER_OPERATOR_USER_IDS", "")
	t.Setenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS", "")
	brand := f.ok(f.operator, "jazebe_brand_create", map[string]any{"name": "برند " + f.owner[:8], "ownerUserId": f.owner})
	f.brandID = brand["id"].(string)
	planet := f.ok(f.owner, "jazebe_planet_create", map[string]any{"brandId": f.brandID, "name": "سیاره " + f.owner[:8],
		"slug": "fx-" + f.owner[:8], "latinName": "Fixture Planet", "pointName": "امتیاز"})
	f.planetID = planet["id"].(string)
	// Publishing needs an active mission and a planet category, so the
	// fixture activates directly; the lifecycle RPCs are tested separately.
	f.exec(`UPDATE jazebe_planet SET status='active',activated_at=now() WHERE id=$1`, f.planetID)
	for _, name := range members {
		id := f.user(name)
		f.ok(id, "jazebe_planet_join", map[string]any{"planetId": f.planetID})
	}
	return f
}

func (f *jazebeFixture) user(name string) string {
	f.t.Helper()
	id := uuid.Must(uuid.NewV4()).String()
	f.exec(`INSERT INTO users(id,username,display_name) VALUES($1,$2,$3)`, id, "fx-"+name+"-"+id[:8], "کاربر "+name)
	f.users[name] = id
	return id
}

func (f *jazebeFixture) exec(query string, args ...any) {
	f.t.Helper()
	if _, err := f.db.ExecContext(f.ctx, query, args...); err != nil {
		f.t.Fatalf("fixture SQL failed: %v", err)
	}
}

// call invokes an RPC as the given user ("" for anonymous) and returns the
// decoded payload or the client-visible runtime error.
func (f *jazebeFixture) call(user, rpc string, payload any) (map[string]any, *rpcErr) {
	f.t.Helper()
	fn, ok := f.rpcs[rpc]
	if !ok {
		f.t.Fatalf("RPC %s is not registered", rpc)
	}
	b, err := json.Marshal(payload)
	if err != nil {
		f.t.Fatal(err)
	}
	ctx := f.ctx
	if user != "" {
		ctx = context.WithValue(ctx, runtime.RUNTIME_CTX_USER_ID, user)
	}
	out, err := fn(ctx, f.logger, f.db, f.nk, string(b))
	if err != nil {
		var re *runtime.Error
		if !errors.As(err, &re) {
			f.t.Fatalf("%s returned a non-runtime error: %v", rpc, err)
		}
		if re.Code == 13 {
			f.t.Logf("%s internal error; module log: %v", rpc, f.logger.take())
		}
		return nil, &rpcErr{Code: re.Code, Message: re.Message}
	}
	var decoded map[string]any
	if out != "" && json.Unmarshal([]byte(out), &decoded) != nil {
		decoded = map[string]any{"_raw": out}
	}
	return decoded, nil
}

func (f *jazebeFixture) ok(user, rpc string, payload any) map[string]any {
	f.t.Helper()
	out, err := f.call(user, rpc, payload)
	if err != nil {
		f.t.Fatalf("%s: code %d %q", rpc, err.Code, err.Message)
	}
	return out
}

func (f *jazebeFixture) fails(user, rpc string, payload any, code int) {
	f.t.Helper()
	if _, err := f.call(user, rpc, payload); err == nil || err.Code != code {
		f.t.Fatalf("%s: got %+v, want code %d", rpc, err, code)
	}
}

func (f *jazebeFixture) count(query string, args ...any) int {
	f.t.Helper()
	var n int
	if err := f.db.QueryRowContext(f.ctx, query, args...).Scan(&n); err != nil {
		f.t.Fatalf("count failed: %v", err)
	}
	return n
}

// audit returns the params of the only audit row for an action and target.
func (f *jazebeFixture) audit(action, targetID string) (actorType, actor string, params map[string]any) {
	f.t.Helper()
	var raw []byte
	var actorID sql.NullString
	rows := f.count(`SELECT count(*) FROM jazebe_audit_log WHERE action=$1 AND target_id=$2`, action, targetID)
	if rows != 1 {
		f.t.Fatalf("audit %s for %s: %d rows, want 1", action, targetID, rows)
	}
	if err := f.db.QueryRowContext(f.ctx, `SELECT actor_type,actor_id,params FROM jazebe_audit_log WHERE action=$1 AND target_id=$2`,
		action, targetID).Scan(&actorType, &actorID, &raw); err != nil {
		f.t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		f.t.Fatal(err)
	}
	return actorType, actorID.String, params
}

// activity returns the payload of the newest group activity with a key.
func (f *jazebeFixture) activity(groupID, key string) (actor string, payload map[string]any) {
	f.t.Helper()
	var raw []byte
	var actorID sql.NullString
	if err := f.db.QueryRowContext(f.ctx, `SELECT actor_user_id,payload FROM jazebe_group_activity
		WHERE group_id=$1 AND message_key=$2 ORDER BY created_at DESC LIMIT 1`, groupID, key).Scan(&actorID, &raw); err != nil {
		f.t.Fatalf("activity %s for group %s: %v", key, groupID, err)
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		f.t.Fatal(err)
	}
	return actorID.String, payload
}

func expectJSON(t *testing.T, label string, got map[string]any, want map[string]any) {
	t.Helper()
	for k, v := range want {
		if fmt.Sprint(got[k]) != fmt.Sprint(v) {
			t.Errorf("%s: %s=%v (%T), want %v", label, k, got[k], got[k], v)
		}
	}
}

// cleanup deletes only rows tied to this fixture's users, brand and planet,
// retrying in passes until foreign keys are satisfied, then verifies that
// nothing tied to the fixture remains.
func (f *jazebeFixture) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	users := make([]string, 0, len(f.users))
	for _, id := range f.users {
		users = append(users, id)
	}
	if len(users) == 0 {
		return
	}
	scope := map[string]string{}
	rows, err := f.db.QueryContext(ctx, `SELECT c.conrelid::regclass::text, a.attname, c.confrelid::regclass::text
		FROM pg_constraint c JOIN pg_attribute a ON a.attrelid=c.conrelid AND a.attnum=ANY(c.conkey)
		WHERE c.contype='f' AND c.conrelid::regclass::text LIKE 'jazebe\_%'
		AND c.confrelid::regclass::text IN ('users','jazebe_brand','jazebe_planet')`)
	if err != nil {
		f.t.Errorf("cleanup discovery: %v", err)
		return
	}
	for rows.Next() {
		var table, column, parent string
		if err := rows.Scan(&table, &column, &parent); err != nil {
			f.t.Errorf("cleanup discovery: %v", err)
			rows.Close()
			return
		}
		arg := "$1"
		switch parent {
		case "jazebe_brand":
			arg = "$2"
		case "jazebe_planet":
			arg = "$3"
		}
		cond := fmt.Sprintf("%q::text = ANY(%s::text[])", column, arg)
		if scope[table] == "" {
			scope[table] = cond
		} else if !strings.Contains(scope[table], cond) {
			scope[table] += " OR " + cond
		}
	}
	rows.Close()
	scope["jazebe_audit_log"] = `actor_id::text = ANY($1::text[]) OR target_id::text = ANY($1::text[]) OR brand_id::text = ANY($2::text[]) OR planet_id::text = ANY($3::text[])`
	scope["jazebe_lottery_group_stat"] = `lottery_id IN (SELECT id FROM jazebe_lottery WHERE planet_id::text = ANY($3::text[]))`
	scope["jazebe_lottery_prize"] = `planet_id::text = ANY($3::text[])`
	scope["jazebe_media_chunk"] = `upload_id IN (SELECT id FROM jazebe_media_upload WHERE owner_user_id::text = ANY($1::text[]))`
	scope["notification"] = `user_id::text = ANY($1::text[])`
	scope["wallet_ledger"] = `user_id::text = ANY($1::text[])`
	scope["storage"] = `user_id::text = ANY($1::text[])`
	// Every statement references all three parameters so none is untyped.
	const typed = ` AND cardinality($1::text[]) + cardinality($2::text[]) + cardinality($3::text[]) >= 0`
	for table, cond := range scope {
		scope[table] = "(" + cond + ")" + typed
	}
	planets := []string{f.planetID}
	if f.brandID != "" {
		if rows, err := f.db.QueryContext(ctx, `SELECT id::text FROM jazebe_planet WHERE brand_id=$1`, f.brandID); err == nil {
			for rows.Next() {
				var id string
				if rows.Scan(&id) == nil && id != f.planetID {
					planets = append(planets, id)
				}
			}
			rows.Close()
		}
	}
	args := []any{users, []string{f.brandID}, planets}
	var lastErr error
	for pass := 0; pass < 25; pass++ {
		pending := 0
		for table, cond := range scope {
			if _, err := f.db.ExecContext(ctx, fmt.Sprintf("DELETE FROM %q WHERE %s", table, cond), args...); err != nil {
				pending++
				lastErr = err
			}
		}
		if _, err := f.db.ExecContext(ctx, `DELETE FROM users WHERE id::text = ANY($1::text[])`, users); err != nil {
			pending++
			lastErr = err
		}
		if pending == 0 {
			break
		}
	}
	remaining := 0
	var left []string
	for table, cond := range scope {
		var n int
		if err := f.db.QueryRowContext(ctx, fmt.Sprintf("SELECT count(*) FROM %q WHERE %s", table, cond), args...).Scan(&n); err != nil {
			f.t.Errorf("cleanup verification of %s: %v", table, err)
			continue
		}
		remaining += n
		if n > 0 {
			left = append(left, fmt.Sprintf("%s=%d", table, n))
		}
	}
	var leftUsers int
	_ = f.db.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE id::text = ANY($1::text[])`, users).Scan(&leftUsers)
	if remaining+leftUsers != 0 {
		f.t.Errorf("fixture cleanup left %d rows %v and %d users: %v", remaining, left, leftUsers, lastErr)
	}
}
