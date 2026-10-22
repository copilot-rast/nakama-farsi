package main

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func walletSummary(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	var name string
	var rate int
	var balance int64
	err = db.QueryRowContext(ctx, `SELECT p.point_name,GREATEST(1,s.point_to_toman_rate),COALESCE(w.balance,0)
		FROM jazebe_planet p JOIN jazebe_planet_settings s ON s.planet_id=p.id
		LEFT JOIN jazebe_point_wallet w ON w.planet_id=p.id AND w.user_id=$2 WHERE p.id=$1 AND p.status='active'`, planetID, user).Scan(&name, &rate, &balance)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "active planet not found")
	}
	if err != nil {
		logger.Error("Wallet summary: %v", err)
		return "", status.Error(codes.Internal, "wallet unavailable")
	}
	return encode(map[string]any{"pointName": name, "balance": balance, "tomanPerPoint": rate})
}

func walletTransactions(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		Range    string `json:"range"`
		Type     string `json:"type"`
		Cursor   string `json:"cursor"`
		Limit    int    `json:"limit"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	if input.Range != "" && input.Range != "all" && input.Range != "month" && input.Range != "quarter" {
		return "", status.Error(codes.InvalidArgument, "invalid range")
	}
	if input.Type != "" && input.Type != "all" && input.Type != "mission" && input.Type != "raffle" && input.Type != "bonus" && input.Type != "purchase" && input.Type != "refund" && input.Type != "adjust" {
		return "", status.Error(codes.InvalidArgument, "invalid transaction type")
	}
	if input.Limit == 0 {
		input.Limit = 50
	}
	if input.Limit < 1 || input.Limit > 200 {
		return "", status.Error(codes.InvalidArgument, "invalid limit")
	}
	var cursor int64
	if input.Cursor != "" {
		cursor, err = strconv.ParseInt(input.Cursor, 10, 64)
		if err != nil || cursor < 1 {
			return "", status.Error(codes.InvalidArgument, "invalid cursor")
		}
	}
	var since any
	switch input.Range {
	case "month":
		since = time.Now().UTC().Add(-30 * 24 * time.Hour)
	case "quarter":
		since = time.Now().UTC().Add(-120 * 24 * time.Hour)
	}
	category := `CASE source
		WHEN 'mission' THEN 'mission' WHEN 'game' THEN 'mission' WHEN 'medal' THEN 'mission'
		WHEN 'lottery_prize' THEN 'raffle' WHEN 'special_bonus' THEN 'bonus'
		WHEN 'purchase' THEN 'purchase' WHEN 'lottery_refund' THEN 'refund'
		WHEN 'game_refund' THEN 'refund' WHEN 'operator_adjustment' THEN 'adjust' ELSE 'bonus' END`
	rows, err := db.QueryContext(ctx, `SELECT id,seq,direction,amount,balance_after,source,description,created_at,`+category+` AS category
		FROM jazebe_point_transaction WHERE planet_id=$1 AND user_id=$2
		AND ($3::bigint=0 OR seq<$3) AND ($4::timestamptz IS NULL OR created_at>=$4)
		AND ($5::text IN ('','all') OR `+category+`=$5)
		ORDER BY seq DESC LIMIT $6`, planetID, user, cursor, since, input.Type, input.Limit+1)
	if err != nil {
		logger.Error("Wallet transactions: %v", err)
		return "", status.Error(codes.Internal, "wallet history unavailable")
	}
	defer rows.Close()
	items := make([]map[string]any, 0, input.Limit+1)
	var last int64
	hasMore := false
	for rows.Next() {
		var id, direction, source, category string
		var seq, amount, balance int64
		var desc sql.NullString
		var created time.Time
		if err := rows.Scan(&id, &seq, &direction, &amount, &balance, &source, &desc, &created, &category); err != nil {
			return "", status.Error(codes.Internal, "wallet history unavailable")
		}
		if len(items) == input.Limit {
			hasMore = true
			break
		}
		last = seq
		signed := amount
		if direction == "debit" && category != "purchase" {
			signed = -amount
		}
		item := map[string]any{"id": id, "seq": strconv.FormatInt(seq, 10), "type": category, "amt": signed, "balanceAfter": balance, "source": source, "createdAt": created}
		if desc.Valid {
			item["desc"] = desc.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "wallet history unavailable")
	}
	var next any
	if hasMore {
		next = strconv.FormatInt(last, 10)
	}
	return encode(map[string]any{"items": items, "nextCursor": next})
}
