package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"net/http"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
)

type settings struct {
	sealKey      []byte
	workerEvery  time.Duration
	workerBatch  int
	maxAttempts  int
	retryBackoff time.Duration
	claimLease   time.Duration
}

type service struct {
	db       *sql.DB
	nk       runtime.NakamaModule
	logger   runtime.Logger
	cfg      settings
	client   *http.Client
	tls      *tls.Config
	baseURLs map[string]string
	now      func() time.Time
}

var defaultBaseURLs = map[string]string{
	"kavenegar": "https://api.kavenegar.com",
	"smsir":     "https://api.sms.ir",
	"twilio":    "https://api.twilio.com",
	"resend":    "https://api.resend.com",
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	env, _ := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	cfg, err := loadSettings(env)
	if err != nil {
		logger.Error("Publication module refused to start: %v", err)
		return err
	}
	svc := newService(db, nk, logger, cfg)
	for name, fn := range svc.clientRPCs() {
		if err := initializer.RegisterRpc(name, svc.rpc(fn)); err != nil {
			return err
		}
	}
	if err := svc.registerConsoleHTTP(initializer); err != nil {
		return err
	}
	go svc.workerLoop(ctx)
	logger.Info("Publication module registered")
	return nil
}

func newService(db *sql.DB, nk runtime.NakamaModule, logger runtime.Logger, cfg settings) *service {
	urls := map[string]string{}
	for k, v := range defaultBaseURLs {
		urls[k] = v
	}
	return &service{db: db, nk: nk, logger: logger, cfg: cfg, client: &http.Client{Timeout: 20 * time.Second}, baseURLs: urls, now: time.Now}
}

func loadSettings(env map[string]string) (settings, error) {
	cfg := settings{workerEvery: 5 * time.Second, workerBatch: 50, maxAttempts: 3, retryBackoff: 30 * time.Second, claimLease: 2 * time.Minute}
	secret, err := secretbox.Secret(env)
	if err != nil {
		return cfg, err
	}
	if cfg.sealKey, err = deriveKey(secret, secretbox.PublicationKeyInfo); err != nil {
		return cfg, err
	}
	return cfg, nil
}
