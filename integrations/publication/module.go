package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"github.com/heroiclabs/nakama/v3/internal/inapp"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
)

type settings struct {
	sealKey      []byte
	otpKey       []byte
	workerEvery  time.Duration
	workerBatch  int
	maxAttempts  int
	retryBackoff time.Duration
	claimLease   time.Duration
	sendTimeout  time.Duration
	autoEvery    time.Duration
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
	timeScan *timeScanGate
}

type timeScanGate struct {
	mu   sync.Mutex
	last time.Time
}

func (g *timeScanGate) due(now time.Time, every time.Duration) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.last.IsZero() && now.Sub(g.last) < every {
		return false
	}
	g.last = now
	return true
}

var defaultBaseURLs = map[string]string{
	"kavenegar": "https://api.kavenegar.com",
	"smsir":     "https://api.sms.ir",
	"twilio":    "https://api.twilio.com",
	"resend":    "https://api.resend.com",
	"fcm":       "https://fcm.googleapis.com",
	"fcmauth":   "https://oauth2.googleapis.com/token",
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
	inapp.SetSender(svc.inappSender())
	go svc.workerLoop(ctx)
	go svc.otpSweepLoop(ctx)
	go svc.bannerSweepLoop(ctx)
	go svc.autoLoop(ctx)
	logger.Info("Publication module registered")
	return nil
}

func newService(db *sql.DB, nk runtime.NakamaModule, logger runtime.Logger, cfg settings) *service {
	urls := map[string]string{}
	for k, v := range defaultBaseURLs {
		urls[k] = v
	}
	return &service{db: db, nk: nk, logger: logger, cfg: cfg, client: &http.Client{Timeout: 20 * time.Second}, baseURLs: urls, now: time.Now, timeScan: &timeScanGate{}}
}

func loadSettings(env map[string]string) (settings, error) {
	cfg := settings{workerEvery: 5 * time.Second, workerBatch: 50, maxAttempts: 3, retryBackoff: 30 * time.Second, claimLease: 2 * time.Minute, sendTimeout: 45 * time.Second, autoEvery: 10 * time.Second}
	if v := strings.TrimSpace(env["PUBLICATION_AUTO_EVERY_SECONDS"]); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 2 || n > 3600 {
			return cfg, errors.New("PUBLICATION_AUTO_EVERY_SECONDS must be 2 to 3600")
		}
		cfg.autoEvery = time.Duration(n) * time.Second
	}
	secret, err := secretbox.Secret(env)
	if err != nil {
		return cfg, err
	}
	if cfg.sealKey, err = deriveKey(secret, secretbox.PublicationKeyInfo); err != nil {
		return cfg, err
	}
	if cfg.otpKey, err = deriveKey(secret, "publication-otp-code"); err != nil {
		return cfg, err
	}
	return cfg, nil
}
