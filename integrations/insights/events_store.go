package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/store"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
)

var (
	storageBox   = secretbox.Box{Salt: "nakama-insights", AAD: "insights-storage"}
	connectorBox = secretbox.Box{Salt: "nakama-insights", AAD: "insights-connector"}
)

const storeCheckEvery = 10 * time.Second

type storageConfig struct {
	Postgres   pgConfig    `json:"postgres"`
	ClickHouse chConfig    `json:"clickhouse"`
	Mongo      mongoConfig `json:"mongodb"`
}

type pgConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	User     string `json:"user"`
	SSLMode  string `json:"sslMode"`
	Schema   string `json:"schema"`
}

type chConfig struct {
	URL      string `json:"url"`
	Database string `json:"database"`
	User     string `json:"user"`
	Table    string `json:"table"`
}

type mongoConfig struct {
	Hosts      string `json:"hosts"`
	Database   string `json:"database"`
	User       string `json:"user"`
	AuthSource string `json:"authSource"`
	TLS        bool   `json:"tls"`
	Collection string `json:"collection"`
}

type opener func(ctx context.Context, cfg storageConfig, secret string) (store.Store, error)

var openers = map[string]opener{}

type storeManager struct {
	mu      sync.Mutex
	current store.Store
	driver  string
	version time.Time
	checked time.Time
	err     error
}

func (s *service) eventStore(ctx context.Context) (store.Store, error) {
	m := &s.stores
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current != nil && time.Since(m.checked) < storeCheckEvery {
		return m.current, nil
	}
	driver, cfg, sealed, updated, err := s.readStorage(ctx)
	if err != nil {
		if m.current != nil {
			return m.current, nil
		}
		return nil, err
	}
	m.checked = time.Now()
	if m.current != nil && m.driver == driver && m.version.Equal(updated) {
		return m.current, nil
	}
	next, err := s.openStore(ctx, driver, cfg, sealed)
	if err != nil {
		m.err = err
		if m.current != nil && m.driver == driver {
			return m.current, nil
		}
		return nil, err
	}
	old := m.current
	m.current, m.driver, m.version, m.err = next, driver, updated, nil
	if old != nil {
		go func() { time.Sleep(time.Minute); _ = old.Close() }()
	}
	return next, nil
}

func (s *service) readStorage(ctx context.Context) (string, storageConfig, string, time.Time, error) {
	var driver, sealed string
	var raw []byte
	var updated time.Time
	var cfg storageConfig
	err := s.db.QueryRowContext(ctx, `SELECT driver, config, secret, updated_at FROM insights_storage WHERE id=1`).Scan(&driver, &raw, &sealed, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return store.DriverNakama, cfg, "", time.Time{}, nil
	}
	if err != nil {
		return "", cfg, "", time.Time{}, err
	}
	_ = json.Unmarshal(raw, &cfg)
	return driver, cfg, sealed, updated, nil
}

func (s *service) openSecret(sealed string) (string, error) {
	if sealed == "" {
		return "", nil
	}
	if s.secretKey == nil {
		return "", errors.New("ECONOMY_SECRET_KEY is not set; sealed secrets cannot be opened")
	}
	b, err := storageBox.Open(s.secretKey, sealed)
	return string(b), err
}

func (s *service) openStore(ctx context.Context, driver string, cfg storageConfig, sealed string) (store.Store, error) {
	if driver == store.DriverNakama || driver == "" {
		return store.NewNakamaPG(s.db), nil
	}
	fn, ok := openers[driver]
	if !ok {
		return nil, errors.New("unknown insights store driver " + driver)
	}
	secret, err := s.openSecret(sealed)
	if err != nil {
		return nil, err
	}
	st, err := fn(ctx, cfg, secret)
	if err != nil {
		return nil, err
	}
	if err := st.Ensure(ctx); err != nil {
		_ = st.Close()
		return nil, err
	}
	return st, nil
}

func (s *service) kick() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func claimToken() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type outboxRow struct {
	id       int64
	attempts int
	events   []store.Event
}

func (s *service) drainOnce(ctx context.Context) (int, error) {
	cfg := s.settings(ctx)
	token := claimToken()
	rows, err := s.db.QueryContext(ctx, `UPDATE insights_outbox SET claimed_until = now() + INTERVAL '60 seconds', claim_token = $1
		WHERE id IN (SELECT id FROM insights_outbox WHERE NOT dead AND next_at <= now() AND (claimed_until IS NULL OR claimed_until < now()) ORDER BY id LIMIT $2 FOR UPDATE SKIP LOCKED)
		RETURNING id, attempts, events`, token, cfg.Storage.BatchSize)
	if err != nil {
		return 0, err
	}
	list := []outboxRow{}
	for rows.Next() {
		var r outboxRow
		var raw []byte
		if err := rows.Scan(&r.id, &r.attempts, &raw); err != nil {
			rows.Close()
			return 0, err
		}
		if err := json.Unmarshal(raw, &r.events); err != nil {
			r.events = nil
		}
		list = append(list, r)
	}
	rows.Close()
	if len(list) == 0 {
		return 0, nil
	}
	st, err := s.eventStore(ctx)
	ids := make([]int64, 0, len(list))
	all := []store.Event{}
	for _, r := range list {
		ids = append(ids, r.id)
		all = append(all, r.events...)
	}
	if err == nil {
		wctx, cancel := context.WithTimeout(ctx, 45*time.Second)
		err = st.Write(wctx, all)
		cancel()
	}
	if err != nil {
		msg := err.Error()
		if len(msg) > 500 {
			msg = msg[:500]
		}
		_, uerr := s.db.ExecContext(ctx, `UPDATE insights_outbox SET attempts = attempts + 1, claimed_until = NULL, claim_token = '', last_error = $3,
			next_at = now() + LEAST(300, 2 * power(2, LEAST(attempts, 8)))::int * INTERVAL '1 second', dead = attempts + 1 >= $4
			WHERE id = ANY($1::bigint[]) AND claim_token = $2`, ids, token, msg, cfg.Storage.DeadAfter)
		if uerr != nil {
			return 0, uerr
		}
		return 0, err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM insights_outbox WHERE id = ANY($1::bigint[]) AND claim_token = $2`, ids, token); err != nil {
		return 0, err
	}
	return len(all), nil
}

func (s *service) eventLoop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	n := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-s.wake:
		}
		for i := 0; i < 20; i++ {
			got, err := s.drainOnce(ctx)
			if err != nil {
				if s.logger != nil {
					s.logger.Warn("Insights outbox delivery failed: %v", err)
				}
				break
			}
			if got == 0 {
				break
			}
		}
		if n%3 == 0 {
			var pending int64
			if err := s.db.QueryRowContext(ctx, `SELECT coalesce(sum(count),0)::int8 FROM insights_outbox WHERE NOT dead`).Scan(&pending); err == nil {
				s.pending.Store(pending)
			}
		}
		if n%900 == 1 {
			if ok, err := clusterkit.AcquireLease(ctx, s.db, "insights-event-sweep", holder, 20*time.Minute); err == nil && ok {
				s.sweepEvents(ctx)
			}
		}
		n++
	}
}

func (s *service) sweepEvents(ctx context.Context) {
	cfg := s.settings(ctx)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM insights_batch WHERE ctid IN (SELECT ctid FROM insights_batch WHERE created_at < now() - make_interval(days => $1) LIMIT 50000)`, cfg.Events.ReceiptDays)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM insights_import_run WHERE started_at < now() - INTERVAL '90 days'`)
	if cfg.Storage.KeepDays <= 0 {
		return
	}
	st, err := s.eventStore(ctx)
	if err != nil {
		return
	}
	before := s.now().In(cfg.Location()).AddDate(0, 0, -cfg.Storage.KeepDays).Format(time.DateOnly)
	if n, err := st.Sweep(ctx, before); err != nil && s.logger != nil {
		s.logger.Warn("Insights event sweep failed: %v", err)
	} else if n > 0 && s.logger != nil {
		s.logger.Info("Insights event sweep removed %d events before %s", n, before)
	}
}

func (s *service) outboxGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var pending, rows, dead, deadEvents int64
	var oldest sql.NullTime
	if err := s.db.QueryRowContext(ctx, `SELECT coalesce(sum(count) FILTER (WHERE NOT dead),0)::int8, count(*) FILTER (WHERE NOT dead)::int8, count(*) FILTER (WHERE dead)::int8,
		coalesce(sum(count) FILTER (WHERE dead),0)::int8, min(created_at) FILTER (WHERE NOT dead) FROM insights_outbox`).Scan(&pending, &rows, &dead, &deadEvents, &oldest); err != nil {
		s.fail(w, err)
		return
	}
	failing := []map[string]any{}
	rs, err := s.db.QueryContext(ctx, `SELECT id, source, count, attempts, dead, last_error, created_at, next_at FROM insights_outbox WHERE attempts > 0 ORDER BY dead DESC, id LIMIT 50`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rs.Close()
	for rs.Next() {
		var id int64
		var source, last string
		var count, attempts int
		var isDead bool
		var created, next time.Time
		if err := rs.Scan(&id, &source, &count, &attempts, &isDead, &last, &created, &next); err != nil {
			s.fail(w, err)
			return
		}
		failing = append(failing, map[string]any{"id": id, "source": source, "count": count, "attempts": attempts, "dead": isDead, "lastError": last, "createdAt": created, "nextAt": next})
	}
	out := map[string]any{"pending": pending, "rows": rows, "dead": dead, "deadEvents": deadEvents, "oldest": nil, "failing": failing, "maxPendingEvents": s.settings(ctx).Storage.MaxPendingEvents}
	if oldest.Valid {
		out["oldest"] = oldest.Time
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) outboxRetry(w http.ResponseWriter, r *http.Request) {
	res, err := s.db.ExecContext(r.Context(), `UPDATE insights_outbox SET dead = false, attempts = 0, next_at = now(), claimed_until = NULL, claim_token = '' WHERE dead OR attempts > 0`)
	if err != nil {
		s.fail(w, err)
		return
	}
	n, _ := res.RowsAffected()
	s.kick()
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"rows": n})
}

func (s *service) outboxDiscard(w http.ResponseWriter, r *http.Request) {
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM insights_outbox WHERE dead`)
	if err != nil {
		s.fail(w, err)
		return
	}
	n, _ := res.RowsAffected()
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"rows": n})
}

type storageIn struct {
	Driver      string        `json:"driver"`
	Config      storageConfig `json:"config"`
	Secret      *string       `json:"secret"`
	ClearSecret bool          `json:"clearSecret"`
}

func validDriver(d string) bool {
	if d == store.DriverNakama {
		return true
	}
	_, ok := openers[d]
	return ok
}

func cleanStorage(in *storageIn) error {
	in.Driver = strings.TrimSpace(in.Driver)
	if !validDriver(in.Driver) {
		return pluginkit.Bad("INVALID_DRIVER", "driver", "")
	}
	c := &in.Config
	c.Postgres.Host, c.Postgres.Database, c.Postgres.User = strings.TrimSpace(c.Postgres.Host), strings.TrimSpace(c.Postgres.Database), strings.TrimSpace(c.Postgres.User)
	c.Postgres.Schema = strings.TrimSpace(c.Postgres.Schema)
	switch in.Driver {
	case store.DriverPostgres:
		if c.Postgres.Host == "" {
			return pluginkit.Bad("REQUIRED", "config.postgres.host", "")
		}
		if c.Postgres.Database == "" {
			return pluginkit.Bad("REQUIRED", "config.postgres.database", "")
		}
		if c.Postgres.Port == 0 {
			c.Postgres.Port = 5432
		}
		if c.Postgres.Port < 1 || c.Postgres.Port > 65535 {
			return pluginkit.Bad("VALUE_OUT_OF_RANGE", "config.postgres.port", "")
		}
		switch c.Postgres.SSLMode {
		case "":
			c.Postgres.SSLMode = "require"
		case "disable", "require", "verify-ca", "verify-full", "prefer":
		default:
			return pluginkit.Bad("INVALID_POLICY", "config.postgres.sslMode", "")
		}
		if c.Postgres.Schema == "" {
			c.Postgres.Schema = "public"
		}
	case store.DriverClickHouse:
		c.ClickHouse.URL = strings.TrimRight(strings.TrimSpace(c.ClickHouse.URL), "/")
		if !strings.HasPrefix(c.ClickHouse.URL, "http://") && !strings.HasPrefix(c.ClickHouse.URL, "https://") {
			return pluginkit.Bad("INVALID_URL", "config.clickhouse.url", "")
		}
		if c.ClickHouse.Database == "" {
			c.ClickHouse.Database = "default"
		}
	case store.DriverMongo:
		c.Mongo.Hosts = strings.TrimSpace(c.Mongo.Hosts)
		if c.Mongo.Hosts == "" {
			return pluginkit.Bad("REQUIRED", "config.mongodb.hosts", "")
		}
		if strings.TrimSpace(c.Mongo.Database) == "" {
			return pluginkit.Bad("REQUIRED", "config.mongodb.database", "")
		}
	}
	return nil
}

func (s *service) storageView(ctx context.Context) (map[string]any, error) {
	driver, cfg, sealed, updated, err := s.readStorage(ctx)
	if err != nil {
		return nil, err
	}
	var by string
	_ = s.db.QueryRowContext(ctx, `SELECT updated_by FROM insights_storage WHERE id=1`).Scan(&by)
	drivers := []string{store.DriverNakama}
	for _, d := range []string{store.DriverPostgres, store.DriverClickHouse, store.DriverMongo} {
		if _, ok := openers[d]; ok {
			drivers = append(drivers, d)
		}
	}
	out := map[string]any{"driver": driver, "config": cfg, "hasSecret": sealed != "", "drivers": drivers, "secretKey": s.secretKey != nil, "updatedBy": by, "updatedAt": nil, "status": "ok", "error": ""}
	if !updated.IsZero() {
		out["updatedAt"] = updated
	}
	s.stores.mu.Lock()
	if s.stores.err != nil {
		out["status"], out["error"] = "error", s.stores.err.Error()
	}
	s.stores.mu.Unlock()
	return out, nil
}

func (s *service) storageGet(w http.ResponseWriter, r *http.Request) {
	out, err := s.storageView(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) resolveSecret(ctx context.Context, in storageIn) (string, string, error) {
	_, _, sealed, _, err := s.readStorage(ctx)
	if err != nil {
		return "", "", err
	}
	if in.ClearSecret {
		return "", "", nil
	}
	if in.Secret == nil {
		plain, err := s.openSecret(sealed)
		return plain, sealed, err
	}
	if *in.Secret == "" {
		return "", "", nil
	}
	if s.secretKey == nil {
		return "", "", pluginkit.Bad("SECRET_KEY_MISSING", "secret", "ECONOMY_SECRET_KEY is not set")
	}
	next, err := storageBox.Seal(s.secretKey, []byte(*in.Secret))
	return *in.Secret, next, err
}

func (s *service) storageTest(w http.ResponseWriter, r *http.Request) {
	var in storageIn
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.fail(w, err)
		return
	}
	if err := cleanStorage(&in); err != nil {
		s.fail(w, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	plain, _, err := s.resolveSecret(ctx, in)
	if err != nil {
		s.fail(w, err)
		return
	}
	started := time.Now()
	result := map[string]any{"ok": true, "error": "", "driver": in.Driver}
	var st store.Store
	if in.Driver == store.DriverNakama {
		st = store.NewNakamaPG(s.db)
	} else if st, err = openers[in.Driver](ctx, in.Config, plain); err == nil {
		defer st.Close()
		err = st.Ensure(ctx)
	}
	if err == nil {
		err = st.Ping(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
	}
	if err != nil {
		result["ok"], result["error"] = false, err.Error()
	}
	result["millis"] = time.Since(started).Milliseconds()
	pluginkit.WriteJSON(w, http.StatusOK, result)
}

func (s *service) storagePut(w http.ResponseWriter, r *http.Request) {
	var in storageIn
	if err := pluginkit.ReadBody(r, &in, 64*1024); err != nil {
		s.fail(w, err)
		return
	}
	if err := cleanStorage(&in); err != nil {
		s.fail(w, err)
		return
	}
	ctx := r.Context()
	_, sealed, err := s.resolveSecret(ctx, in)
	if err != nil {
		s.fail(w, err)
		return
	}
	raw, _ := json.Marshal(in.Config)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO insights_storage (id, driver, config, secret, updated_by, updated_at) VALUES (1, $1, $2, $3, $4, now())
		ON CONFLICT (id) DO UPDATE SET driver=EXCLUDED.driver, config=EXCLUDED.config, secret=EXCLUDED.secret, updated_by=EXCLUDED.updated_by, updated_at=now()`,
		in.Driver, raw, sealed, pluginkit.ConsoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	s.stores.mu.Lock()
	s.stores.checked = time.Time{}
	s.stores.mu.Unlock()
	if _, err := s.eventStore(ctx); err != nil && s.logger != nil {
		s.logger.Warn("Insights store %s does not open yet: %v", in.Driver, err)
	}
	s.kick()
	s.storageGet(w, r)
}
