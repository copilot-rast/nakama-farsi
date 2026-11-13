package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
)

const (
	roundCollection = "stargame_rounds"
	currencyKey     = "coin"
	minRound        = 25 * time.Second
	maxRound        = 5 * time.Minute
	maxScore        = 400
	maxCoins        = 30
)

type round struct {
	StartedAt  int64 `json:"startedAt"`
	Finished   bool  `json:"finished"`
	FinishedAt int64 `json:"finishedAt,omitempty"`
	Score      int64 `json:"score,omitempty"`
	Seconds    int   `json:"seconds,omitempty"`
	MaxCoins   int64 `json:"maxCoins,omitempty"`
	Coins      int64 `json:"coins,omitempty"`
}

var (
	errUnauthenticated = runtime.NewError("UNAUTHENTICATED", 16)
	errInvalidJSON     = runtime.NewError("INVALID_JSON", 3)
	errRoundNotFound   = runtime.NewError("ROUND_NOT_FOUND", 5)
	errRoundFinished   = runtime.NewError("ROUND_ALREADY_FINISHED", 9)
	errRoundTooShort   = runtime.NewError("ROUND_TOO_SHORT", 9)
	errRoundExpired    = runtime.NewError("ROUND_EXPIRED", 9)
	errScoreTooHigh    = runtime.NewError("SCORE_TOO_HIGH", 3)
	errInternal        = runtime.NewError("INTERNAL_ERROR", 13)
	errGameDisabled    = runtime.NewError("GAME_DISABLED", 9)
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	if err := initializer.RegisterRpc("stargame_start", rpcStart); err != nil {
		return err
	}
	if err := initializer.RegisterRpc("stargame_finish", rpcFinish); err != nil {
		return err
	}
	if err := registerGames(initializer); err != nil {
		return err
	}
	if err := nk.LeaderboardCreate(ctx, tttLeaderboard, true, "desc", "incr", "", map[string]interface{}{"game": tttGameKey}, true); err != nil {
		logger.Warn("Tic-tac-toe leaderboard not created: %v", err)
	}
	env, _ := ctx.Value(runtime.RUNTIME_CTX_ENV).(map[string]string)
	site, err := newWebSite(strings.TrimSpace(env["STARGAME_SERVER_KEY"]))
	if err != nil {
		return err
	}
	if err := site.register(initializer); err != nil {
		return err
	}
	if seedEnabled(env) {
		if report, err := seed(ctx, db, env); err != nil {
			logger.Error("Star Catcher seed failed: %v", err)
		} else {
			logger.Info("Star Catcher seed: %s", report)
		}
	} else {
		logger.Info("Star Catcher seed disabled by STARGAME_SEED=false")
	}
	logger.Info("Star Catcher sample module registered")
	return nil
}

func userID(ctx context.Context) (string, error) {
	id, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if _, err := uuid.FromString(id); err != nil {
		return "", errUnauthenticated
	}
	return id, nil
}

func rpcStart(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	uid, err := userID(ctx)
	if err != nil {
		return "", err
	}
	seconds, coins, err := roundSettings(ctx, logger)
	if err != nil {
		return "", err
	}
	id := uuid.Must(uuid.NewV4()).String()
	value, _ := json.Marshal(round{StartedAt: time.Now().UnixMilli(), Seconds: seconds, MaxCoins: coins})
	if _, err := nk.StorageWrite(ctx, []*runtime.StorageWrite{{
		Collection: roundCollection, Key: id, UserID: uid, Value: string(value), Version: "*", PermissionRead: 0, PermissionWrite: 0,
	}}); err != nil {
		logger.Error("stargame_start write failed: %v", err)
		return "", errInternal
	}
	out, _ := json.Marshal(map[string]any{"roundId": id, "seconds": seconds, "maxCoins": coins})
	return string(out), nil
}

func rpcFinish(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	uid, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		RoundID string `json:"roundId"`
		Score   int64  `json:"score"`
	}
	if err := json.Unmarshal([]byte(payload), &in); err != nil {
		return "", errInvalidJSON
	}
	if _, err := uuid.FromString(in.RoundID); err != nil {
		return "", errRoundNotFound
	}
	if in.Score > maxScore {
		return "", errScoreTooHigh
	}
	objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{{Collection: roundCollection, Key: in.RoundID, UserID: uid}})
	if err != nil {
		logger.Error("stargame_finish read failed: %v", err)
		return "", errInternal
	}
	if len(objects) == 0 {
		return "", errRoundNotFound
	}
	var r round
	if err := json.Unmarshal([]byte(objects[0].Value), &r); err != nil {
		return "", errInternal
	}
	if r.Finished {
		return "", errRoundFinished
	}
	now := time.Now()
	elapsed := now.Sub(time.UnixMilli(r.StartedAt))
	least, limit := minRound, int64(maxCoins)
	if r.Seconds > 0 {
		least = time.Duration(r.Seconds-5) * time.Second
		limit = r.MaxCoins
	}
	if elapsed < least {
		return "", errRoundTooShort
	}
	if elapsed > maxRound {
		return "", errRoundExpired
	}
	coins := in.Score / 5
	if coins > limit {
		coins = limit
	}
	if coins < 0 {
		coins = 0
	}
	r.Finished, r.FinishedAt, r.Score, r.Coins = true, now.UnixMilli(), in.Score, coins
	value, _ := json.Marshal(r)
	writes := []*runtime.StorageWrite{{
		Collection: roundCollection, Key: in.RoundID, UserID: uid, Value: string(value), Version: objects[0].Version, PermissionRead: 0, PermissionWrite: 0,
	}}
	var wallets []*runtime.WalletUpdate
	if coins > 0 {
		wallets = append(wallets, &runtime.WalletUpdate{UserID: uid, Changeset: map[string]int64{currencyKey: coins}, Metadata: map[string]any{
			"module": "stargame", "reason": "round", "roundId": in.RoundID, "score": in.Score,
		}})
	}
	_, results, err := nk.MultiUpdate(ctx, nil, writes, nil, wallets, true)
	if err != nil {
		if errors.Is(err, runtime.ErrStorageRejectedVersion) {
			return "", errRoundFinished
		}
		logger.Error("stargame_finish update failed: %v", err)
		return "", errInternal
	}
	var balance int64
	if len(results) > 0 {
		balance = results[0].Updated[currencyKey]
	} else {
		account, err := nk.AccountGetId(ctx, uid)
		if err != nil {
			return "", errInternal
		}
		wallet := map[string]int64{}
		_ = json.Unmarshal([]byte(account.Wallet), &wallet)
		balance = wallet[currencyKey]
	}
	result := map[string]any{"coins": coins, "balance": balance, "score": in.Score}
	if events, err := playerevents.Record(ctx, playerevents.Event{UserID: uid, Key: "round_finished", Value: float64(in.Score), Metadata: map[string]any{"roundId": in.RoundID, "coins": coins}}); err == nil {
		result["events"] = events
	} else if !errors.Is(err, playerevents.ErrUnavailable) {
		logger.Warn("stargame_finish event not recorded: %v", err)
	}
	out, _ := json.Marshal(result)
	return string(out), nil
}
