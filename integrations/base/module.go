package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/objectstore"

	"github.com/heroiclabs/nakama-common/runtime"
)

type service struct {
	db         *sql.DB
	nk         runtime.NakamaModule
	logger     runtime.Logger
	now        func() time.Time
	files      objectstore.Store
	fileKey    []byte
	publicBase string
	updates    updatesConfig
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	svc := &service{db: db, nk: nk, logger: logger, now: time.Now}
	env, _ := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	if err := svc.configureFiles(env, dataDirFromArgs(os.Args)); err != nil {
		logger.Warn("Base support attachments are off: %v", err)
	}
	svc.configureUpdates(env)
	for name, fn := range svc.clientRPCs() {
		limit := 16384
		if name == "base_support_upload_part" {
			limit = 180000
		}
		if err := initializer.RegisterRpc(name, svc.rpcSized(fn, limit)); err != nil {
			return err
		}
	}
	if err := initializer.RegisterHttp("/v2/support/files/{key}", svc.fileHTTP, http.MethodGet, http.MethodHead); err != nil {
		return err
	}
	if err := svc.registerConsoleHTTP(initializer); err != nil {
		return err
	}
	if err := svc.registerRuleHooks(initializer); err != nil {
		return err
	}
	if err := svc.registerModerationHooks(initializer); err != nil {
		return err
	}
	if err := svc.registerActivityHooks(initializer); err != nil {
		return err
	}
	go svc.restrictionLoop(ctx)
	go svc.supportLoop(ctx)
	go svc.updatesLoop(ctx)
	logger.Info("Base module registered")
	return nil
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
