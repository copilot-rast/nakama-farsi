package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var purchaseKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{8,128}$`)

type ticketPurchaseInput struct {
	PlanetID       string `json:"planetId"`
	ProductID      string `json:"productId"`
	IdempotencyKey string `json:"idempotencyKey"`
}

// Matches Jazebe's SHA-256 based uuidFromString for an idempotent order ID.
func purchaseOrderID(user, product, key string) string {
	digest := sha256.Sum256([]byte("purchase:" + user + ":" + product + ":" + key))
	h := hex.EncodeToString(digest[:])
	variant := "89ab"[(int(digest[8])>>4)&3]
	return h[:8] + "-" + h[8:12] + "-5" + h[13:16] + "-" + string(variant) + h[17:20] + "-" + h[20:32]
}
func validateTicketPurchase(in ticketPurchaseInput) (ticketPurchaseInput, error) {
	var err error
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return in, err
	}
	in.ProductID, err = requiredUUID(in.ProductID)
	if err != nil {
		return in, err
	}
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)
	if !purchaseKeyPattern.MatchString(in.IdempotencyKey) {
		return in, status.Error(codes.InvalidArgument, "invalid idempotency key")
	}
	return in, nil
}

func ticketPurchaseReplay(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, in ticketPurchaseInput, user, orderID string) (string, bool, error) {
	var productID, name, kind string
	var paid int
	var delivery []byte
	var purchased time.Time
	var balance int64
	err := q.QueryRowContext(ctx, `SELECT o.shop_product_id,o.product_name_snapshot,o.product_type_snapshot,o.paid_points,o.delivery,o.purchased_at,COALESCE(w.balance,0) FROM jazebe_order o LEFT JOIN jazebe_point_wallet w ON w.planet_id=o.planet_id AND w.user_id=o.user_id WHERE o.id=$1 AND o.planet_id=$2 AND o.user_id=$3`, orderID, in.PlanetID, user).Scan(&productID, &name, &kind, &paid, &delivery, &purchased, &balance)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, status.Error(codes.Internal, "purchase replay failed")
	}
	var detail map[string]any
	if json.Unmarshal(delivery, &detail) != nil {
		return "", false, status.Error(codes.Internal, "order delivery invalid")
	}
	lotteryID, ok := detail["lotteryId"].(string)
	if !ok || kind != "ticket" {
		return "", false, status.Error(codes.Internal, "order delivery invalid")
	}
	var lotteryTitle string
	if err := q.QueryRowContext(ctx, `SELECT COALESCE((SELECT title FROM jazebe_lottery WHERE id=$1 AND planet_id=$2),'')`, lotteryID, in.PlanetID).Scan(&lotteryTitle); err != nil {
		return "", false, status.Error(codes.Internal, "order lottery unavailable")
	}
	record := shopTicketRecord(orderID, productID, name, paid, purchased, lotteryID, lotteryTitle, time.Now().UTC())
	out, err := encode(map[string]any{"record": record, "balance": balance})
	return out, true, err
}

func shopTicketPurchase(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	if !shopRateAlreadyChecked(ctx) {
		if err = shopPurchaseRate(ctx, db, user); err != nil {
			return "", err
		}
	}
	var in ticketPurchaseInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in, err = validateTicketPurchase(in)
	if err != nil {
		return "", err
	}
	if err = lotteryClientGuard(ctx, db, in.PlanetID, user); err != nil {
		return "", err
	}
	orderID := purchaseOrderID(user, in.ProductID, in.IdempotencyKey)
	if out, found, err := ticketPurchaseReplay(ctx, db, in, user, orderID); err != nil || found {
		return out, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "purchase unavailable")
	}
	defer tx.Rollback()
	var name, kind, state string
	var price int
	var config []byte
	err = tx.QueryRowContext(ctx, `SELECT name,type,status,price_points,type_config FROM jazebe_shop_product WHERE id=$1 AND planet_id=$2`, in.ProductID, in.PlanetID).Scan(&name, &kind, &state, &price, &config)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "product not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "product lookup failed")
	}
	if kind != "ticket" {
		return "", status.Error(codes.FailedPrecondition, "product delivery type unavailable")
	}
	if state != "active" {
		return "", status.Error(codes.FailedPrecondition, "product not active")
	}
	var ticketConfig struct {
		LotteryID string `json:"lotteryId"`
	}
	if json.Unmarshal(config, &ticketConfig) != nil || ticketConfig.LotteryID == "" {
		return "", status.Error(codes.FailedPrecondition, "ticket product configuration invalid")
	}
	var lotteryTitle string
	if err = tx.QueryRowContext(ctx, `SELECT title FROM jazebe_lottery WHERE id=$1 AND planet_id=$2`, ticketConfig.LotteryID, in.PlanetID).Scan(&lotteryTitle); err != nil {
		return "", status.Error(codes.FailedPrecondition, "ticket lottery unavailable")
	}
	group, err := activeGroupForMission(ctx, tx, in.PlanetID, user, true)
	if err != nil {
		return "", err
	}
	movement, err := applyPointMovement(ctx, tx, in.PlanetID, user, "debit", int64(price), "purchase", orderID, name, time.Now().UTC())
	if errors.Is(err, errInsufficientPoints) {
		return "", status.Error(codes.FailedPrecondition, "insufficient points")
	}
	if err != nil {
		logger.Error("Purchase point debit: %v", err)
		return "", status.Error(codes.Internal, "purchase failed")
	}
	if !movement.Applied {
		return ticketPurchaseReplayAfterRollback(ctx, tx, db, in, user, orderID)
	}
	result, err := tx.ExecContext(ctx, `UPDATE jazebe_shop_product SET stock=CASE WHEN stock IS NULL THEN NULL ELSE stock-1 END,sold_count=sold_count+1,updated_at=now() WHERE id=$1 AND status='active' AND (stock IS NULL OR stock>0)`, in.ProductID)
	if err != nil {
		return "", status.Error(codes.Internal, "product stock update failed")
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return "", status.Error(codes.ResourceExhausted, "product out of stock")
	}
	serials, err := issueLotteryTickets(ctx, tx, ticketIssue{LotteryID: ticketConfig.LotteryID, PlanetID: in.PlanetID, GroupID: group.String, UserID: user, CreatorType: "member", Source: "purchase", SourceID: orderID, Count: 1}, time.Now().UTC())
	if err != nil {
		return "", err
	}
	var ticketID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND source='purchase' AND source_id=$2`, ticketConfig.LotteryID, orderID).Scan(&ticketID)
	if err != nil {
		return "", status.Error(codes.Internal, "ticket delivery failed")
	}
	delivery, _ := json.Marshal(map[string]any{"ticketId": ticketID, "ticketSerial": serials[0], "lotteryId": ticketConfig.LotteryID})
	now := time.Now().UTC()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_order(id,planet_id,user_id,shop_product_id,product_name_snapshot,product_type_snapshot,paid_points,status,delivery,point_transaction_id,purchased_at) VALUES($1,$2,$3,$4,$5,'ticket',$6,'delivered',$7::jsonb,$8,$9)`, orderID, in.PlanetID, user, in.ProductID, name, price, string(delivery), movement.TransactionID, now)
	if err != nil {
		return "", status.Error(codes.Internal, "order record failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_planet_stats(planet_id,user_id) VALUES($1,$2) ON CONFLICT(planet_id,user_id) DO NOTHING`, in.PlanetID, user); err != nil {
		return "", status.Error(codes.Internal, "purchase stats failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_user_planet_stats SET tickets_created=tickets_created+1,last_activity_at=$3 WHERE planet_id=$1 AND user_id=$2`, in.PlanetID, user, now); err != nil {
		return "", status.Error(codes.Internal, "purchase stats failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload) VALUES($1,$2,$3,'ticket',$4,'ticket.bought',jsonb_build_object('lotteryId',$5::uuid,'tickets',1,'serials',jsonb_build_array($6::text)))`, uuid.Must(uuid.NewV4()).String(), in.PlanetID, group.String, user, ticketConfig.LotteryID, serials[0]); err != nil {
		return "", status.Error(codes.Internal, "purchase group activity failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET last_activity_at=$3 WHERE planet_id=$1 AND user_id=$2`, in.PlanetID, user, now); err != nil {
		return "", status.Error(codes.Internal, "purchase membership activity failed")
	}
	if err = enqueueActivityXP(ctx, tx, in.PlanetID, user, orderID, "purchase"); err != nil {
		return "", status.Error(codes.Internal, "purchase progression event failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "purchase failed")
	}
	shopPurchaseProgression(ctx, logger, db, in.PlanetID, user, orderID)
	record := shopTicketRecord(orderID, in.ProductID, name, price, now, ticketConfig.LotteryID, lotteryTitle, now)
	return encode(map[string]any{"record": record, "balance": movement.Balance})
}

func ticketPurchaseReplayAfterRollback(ctx context.Context, tx *sql.Tx, db *sql.DB, in ticketPurchaseInput, user, orderID string) (string, error) {
	_ = tx.Rollback()
	out, found, err := ticketPurchaseReplay(ctx, db, in, user, orderID)
	if err != nil {
		return "", err
	}
	if !found {
		return "", status.Error(codes.Internal, "purchase replay missing order")
	}
	return out, nil
}
