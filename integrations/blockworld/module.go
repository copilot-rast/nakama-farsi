package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

//go:embed blocks.json
var blocksJSON []byte

type config struct {
	issuer      string
	keyFile     string
	createKey   bool
	currency    string
	tokenTTL    time.Duration
	winPrize    int64
	startCoins  int64
	redelivery  time.Duration
	syncEvery   time.Duration
	seedEnabled bool
}

type service struct {
	db     *sql.DB
	nk     runtime.NakamaModule
	logger runtime.Logger
	cfg    config
	signer *signer
	now    func() time.Time
}

type rpcFn func(ctx context.Context, userID, payload string) (any, error)

func envInt(env map[string]string, key string, def int64) int64 {
	v := strings.TrimSpace(env[key])
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func dataDirFromArgs(args []string) string {
	for i, a := range args {
		if a == "--data_dir" && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(a, "--data_dir=") {
			return strings.TrimPrefix(a, "--data_dir=")
		}
	}
	return "./data"
}

func loadConfig(env map[string]string, dataDir string) (config, error) {
	c := config{
		issuer:      strings.TrimRight(strings.TrimSpace(env["BLOCKWORLD_ISSUER"]), "/"),
		keyFile:     strings.TrimSpace(env["BLOCKWORLD_JWT_KEY_FILE"]),
		currency:    strings.TrimSpace(env["BLOCKWORLD_CURRENCY"]),
		tokenTTL:    time.Duration(envInt(env, "BLOCKWORLD_TOKEN_TTL_SECONDS", 3600)) * time.Second,
		winPrize:    envInt(env, "BLOCKWORLD_WIN_PRIZE", 100),
		startCoins:  envInt(env, "BLOCKWORLD_START_COINS", 10000),
		redelivery:  time.Duration(envInt(env, "BLOCKWORLD_REDELIVERY_SECONDS", 10)) * time.Second,
		syncEvery:   time.Duration(envInt(env, "BLOCKWORLD_SYNC_MS", 1500)) * time.Millisecond,
		seedEnabled: strings.TrimSpace(env["BLOCKWORLD_SEED"]) != "0",
	}
	if c.issuer == "" {
		base := strings.TrimRight(strings.TrimSpace(env["ECONOMY_PUBLIC_BASE_URL"]), "/")
		if base == "" {
			return c, errors.New("set BLOCKWORLD_ISSUER (or ECONOMY_PUBLIC_BASE_URL) to the public URL of the API port plus /blockworld")
		}
		c.issuer = base + "/blockworld"
	}
	if !strings.HasPrefix(c.issuer, "http://") && !strings.HasPrefix(c.issuer, "https://") {
		return c, errors.New("BLOCKWORLD_ISSUER must start with http:// or https://")
	}
	if len(c.issuer) > 128 {
		return c, errors.New("BLOCKWORLD_ISSUER must be at most 128 characters")
	}
	if c.keyFile == "" {
		c.keyFile = filepath.Join(dataDir, "blockworld", "jwt-rs256.pem")
		c.createKey = true
	}
	if c.currency == "" {
		c.currency = "bw_coin"
	}
	if c.tokenTTL < time.Minute || c.tokenTTL > 24*time.Hour {
		return c, errors.New("BLOCKWORLD_TOKEN_TTL_SECONDS must be between 60 and 86400")
	}
	if c.redelivery < time.Second {
		c.redelivery = time.Second
	}
	if c.syncEvery < 200*time.Millisecond {
		c.syncEvery = 200 * time.Millisecond
	}
	return c, nil
}

func issuerPath(issuer string) string {
	rest := issuer
	if i := strings.Index(rest, "://"); i >= 0 {
		rest = rest[i+3:]
	}
	if i := strings.Index(rest, "/"); i >= 0 {
		return strings.TrimRight(rest[i:], "/")
	}
	return ""
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	env, _ := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	cfg, err := loadConfig(env, dataDirFromArgs(os.Args))
	if err != nil {
		return fmt.Errorf("blockworld: %w", err)
	}
	key, err := loadOrCreateKey(cfg.keyFile, cfg.createKey)
	if err != nil {
		return fmt.Errorf("blockworld signing key: %w", err)
	}
	svc := &service{db: db, nk: nk, logger: logger, cfg: cfg, signer: newSigner(key, cfg.issuer), now: time.Now}
	if cfg.seedEnabled {
		if err := svc.seed(ctx); err != nil {
			return fmt.Errorf("blockworld seed: %w", err)
		}
	}
	for name, fn := range svc.clientRPCs() {
		if err := initializer.RegisterRpc(name, svc.wrap(fn, false)); err != nil {
			return err
		}
	}
	for name, fn := range svc.bridgeRPCs() {
		if err := initializer.RegisterRpc(name, svc.wrap(fn, true)); err != nil {
			return err
		}
	}
	base := issuerPath(cfg.issuer)
	if base == "" {
		return errors.New("blockworld: BLOCKWORLD_ISSUER needs a path such as /blockworld")
	}
	if err := initializer.RegisterHttp(base+"/.well-known/openid-configuration", svc.discoveryHTTP, http.MethodGet, http.MethodHead); err != nil {
		return err
	}
	if err := initializer.RegisterHttp(base+"/jwks.json", svc.jwksHTTP, http.MethodGet, http.MethodHead); err != nil {
		return err
	}
	go svc.syncLoop(ctx)
	logger.Info("BlockWorld bridge ready: issuer %s, key id %s, currency %s", cfg.issuer, svc.signer.kid, cfg.currency)
	return nil
}

func (s *service) clientRPCs() map[string]rpcFn {
	return map[string]rpcFn{
		"blockworld_token":     s.rpcToken,
		"blockworld_wallet":    s.rpcWallet,
		"blockworld_buy_block": s.rpcBuyBlock,
	}
}

func (s *service) bridgeRPCs() map[string]rpcFn {
	return map[string]rpcFn{
		"blockworld_bridge_pull":   s.rpcBridgePull,
		"blockworld_bridge_ack":    s.rpcBridgeAck,
		"blockworld_bridge_report": s.rpcBridgeReport,
		"blockworld_bridge_flags":  s.rpcBridgeFlags,
	}
}

func (s *service) wrap(fn rpcFn, bridge bool) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		if bridge && userID != "" {
			return "", runtime.NewError("SERVER_ONLY", 7)
		}
		if !bridge && userID == "" {
			return "", runtime.NewError("UNAUTHENTICATED", 16)
		}
		if len(payload) > 65536 {
			return "", runtime.NewError("PAYLOAD_TOO_LARGE", 3)
		}
		out, err := fn(ctx, userID, payload)
		if err != nil {
			var rt *runtime.Error
			if errors.As(err, &rt) {
				return "", rt
			}
			if ctx.Err() == nil {
				s.logger.Error("BlockWorld RPC failed: %v", err)
			}
			return "", runtime.NewError("INTERNAL_ERROR", 13)
		}
		raw, err := json.Marshal(out)
		if err != nil {
			return "", runtime.NewError("INTERNAL_ERROR", 13)
		}
		return string(raw), nil
	}
}

func decode(payload string, out any) error {
	if strings.TrimSpace(payload) == "" {
		payload = "{}"
	}
	dec := json.NewDecoder(strings.NewReader(payload))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return runtime.NewError("INVALID_JSON", 3)
	}
	return nil
}

func (s *service) discoveryHTTP(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.signer.discovery(), "public, max-age=300")
}

func (s *service) jwksHTTP(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.signer.jwks(), "public, max-age=300")
}

func writeJSON(w http.ResponseWriter, body any, cache string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", cache)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(body)
}

func (s *service) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
