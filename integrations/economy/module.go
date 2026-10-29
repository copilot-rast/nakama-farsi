package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

type settings struct {
	secret      []byte
	sealKey     []byte
	ticketKey   []byte
	publicBase  string
	deepLink    string
	webRedirect string
	pendingTTL  time.Duration
	uploadDir   string
}

type service struct {
	db     *sql.DB
	nk     runtime.NakamaModule
	logger runtime.Logger
	cfg    settings
	client *http.Client
	now    func() time.Time
}

const defaultDeepLink = "nakama://payment?success={success}&paymentId={paymentId}"

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	env, _ := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	cfg, err := loadSettings(env, os.Args)
	if err != nil {
		logger.Error("Economy module refused to start: %v", err)
		return err
	}
	svc := newService(db, nk, logger, cfg)
	if err := os.MkdirAll(cfg.uploadDir, 0o750); err != nil {
		return err
	}
	for name, fn := range svc.clientRPCs() {
		if err := initializer.RegisterRpc(name, svc.rpc(fn)); err != nil {
			return err
		}
	}
	if err := svc.registerPublicHTTP(initializer); err != nil {
		return err
	}
	if err := svc.registerConsoleHTTP(initializer); err != nil {
		return err
	}
	go svc.reconcileLoop(ctx)
	logger.Info("Economy module registered")
	return nil
}

func newService(db *sql.DB, nk runtime.NakamaModule, logger runtime.Logger, cfg settings) *service {
	return &service{db: db, nk: nk, logger: logger, cfg: cfg, client: &http.Client{Timeout: 20 * time.Second}, now: time.Now}
}

func loadSettings(env map[string]string, args []string) (settings, error) {
	var cfg settings
	secret := strings.TrimSpace(env["ECONOMY_SECRET_KEY"])
	if len(secret) < 32 {
		return cfg, errors.New("ECONOMY_SECRET_KEY must be at least 32 characters")
	}
	cfg.secret = []byte(secret)
	var err error
	if cfg.sealKey, err = deriveKey(cfg.secret, "economy-config-seal"); err != nil {
		return cfg, err
	}
	if cfg.ticketKey, err = deriveKey(cfg.secret, "economy-callback-ticket"); err != nil {
		return cfg, err
	}
	cfg.publicBase = strings.TrimRight(strings.TrimSpace(env["ECONOMY_PUBLIC_BASE_URL"]), "/")
	if !strings.HasPrefix(cfg.publicBase, "http://") && !strings.HasPrefix(cfg.publicBase, "https://") {
		return cfg, errors.New("ECONOMY_PUBLIC_BASE_URL must be an http(s) URL")
	}
	cfg.deepLink = strings.TrimSpace(env["ECONOMY_DEEP_LINK"])
	if cfg.deepLink == "" {
		cfg.deepLink = defaultDeepLink
	}
	cfg.webRedirect = strings.TrimSpace(env["ECONOMY_WEB_REDIRECT_URL"])
	cfg.pendingTTL = 45 * time.Minute
	if raw := strings.TrimSpace(env["ECONOMY_PENDING_TTL_MINUTES"]); raw != "" {
		minutes, err := strconv.Atoi(raw)
		if err != nil || minutes < 5 || minutes > 10080 {
			return cfg, errors.New("ECONOMY_PENDING_TTL_MINUTES must be between 5 and 10080")
		}
		cfg.pendingTTL = time.Duration(minutes) * time.Minute
	}
	cfg.uploadDir = strings.TrimSpace(env["ECONOMY_UPLOAD_DIR"])
	if cfg.uploadDir == "" {
		cfg.uploadDir = filepath.Join(dataDirFromArgs(args), "economy-uploads")
	}
	return cfg, nil
}

func dataDirFromArgs(args []string) string {
	for i, arg := range args {
		for _, flag := range []string{"--data_dir", "-data_dir"} {
			if arg == flag && i+1 < len(args) {
				return args[i+1]
			}
			if strings.HasPrefix(arg, flag+"=") {
				return strings.TrimPrefix(arg, flag+"=")
			}
		}
	}
	return "./data"
}

func (s *service) redirectTarget(success bool, paymentID string) string {
	target := s.cfg.deepLink
	if s.cfg.webRedirect != "" {
		target = s.cfg.webRedirect
	}
	flag := "false"
	if success {
		flag = "true"
	}
	return strings.NewReplacer("{success}", flag, "{paymentId}", paymentID).Replace(target)
}
