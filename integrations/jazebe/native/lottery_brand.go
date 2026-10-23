package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lotteryPrizeInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	ValueRial   string `json:"valueRial"`
	Priority    int    `json:"priority"`
	ImageAsset  string `json:"imageAsset"`
}
type lotteryDraftInput struct {
	PlanetID          string              `json:"planetId"`
	Title             string              `json:"title"`
	ShortDescription  string              `json:"shortDescription"`
	LongDescription   string              `json:"longDescription"`
	BannerAsset       string              `json:"bannerAsset"`
	StartsAt          string              `json:"startsAt"`
	DrawAt            string              `json:"drawAt"`
	TicketPricePoints int                 `json:"ticketPricePoints"`
	PerUserTicketCap  *int                `json:"perUserTicketCap"`
	PerGroupTicketCap *int                `json:"perGroupTicketCap"`
	Prizes            []lotteryPrizeInput `json:"prizes"`
}
type lotteryRef struct {
	PlanetID  string `json:"planetId"`
	LotteryID string `json:"lotteryId"`
}

func nullableText(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}

// Source draft forms may contain 50 prizes with descriptions. Keep their
// payload limit separate from the smaller general RPC limit.
func decodeLotteryPayload(payload string, out any) error {
	if len(payload) > 256<<10 {
		return status.Error(codes.InvalidArgument, "lottery request too large")
	}
	if payload == "" {
		payload = "{}"
	}
	d := json.NewDecoder(strings.NewReader(payload))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return status.Error(codes.InvalidArgument, "invalid lottery JSON")
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		return status.Error(codes.InvalidArgument, "multiple JSON values")
	}
	return nil
}

func validateLotteryDraft(in lotteryDraftInput, now time.Time, minPrice, maxPrice int) (time.Time, time.Time, error) {
	bad := func(s string) (time.Time, time.Time, error) {
		return time.Time{}, time.Time{}, status.Error(codes.InvalidArgument, s)
	}
	if !utf8.ValidString(in.Title) || utf8.RuneCountInString(strings.TrimSpace(in.Title)) < 1 || utf8.RuneCountInString(in.Title) > 128 ||
		!utf8.ValidString(in.ShortDescription) || utf8.RuneCountInString(in.ShortDescription) > 512 ||
		!utf8.ValidString(in.LongDescription) || utf8.RuneCountInString(in.LongDescription) > 8000 ||
		!utf8.ValidString(in.BannerAsset) || utf8.RuneCountInString(in.BannerAsset) > 512 {
		return bad("invalid lottery text")
	}
	start, e1 := time.Parse(time.RFC3339, in.StartsAt)
	draw, e2 := time.Parse(time.RFC3339, in.DrawAt)
	if e1 != nil || e2 != nil || !draw.After(start) || !draw.After(now) {
		return bad("draw must be future and after start")
	}
	if in.TicketPricePoints < minPrice || in.TicketPricePoints > maxPrice {
		return bad("ticket price outside planet policy")
	}
	if in.PerUserTicketCap != nil && *in.PerUserTicketCap < 1 || in.PerGroupTicketCap != nil && *in.PerGroupTicketCap < 1 {
		return bad("invalid ticket cap")
	}
	if in.PerUserTicketCap != nil && in.PerGroupTicketCap != nil && *in.PerUserTicketCap > *in.PerGroupTicketCap {
		return bad("user cap exceeds group cap")
	}
	if len(in.Prizes) < 1 || len(in.Prizes) > 50 {
		return bad("lottery requires 1 to 50 prizes")
	}
	seen := map[int]bool{}
	for _, p := range in.Prizes {
		if !utf8.ValidString(p.Title) || utf8.RuneCountInString(strings.TrimSpace(p.Title)) < 1 || utf8.RuneCountInString(p.Title) > 128 ||
			!utf8.ValidString(p.Description) || utf8.RuneCountInString(p.Description) > 2000 ||
			!utf8.ValidString(p.ImageAsset) || utf8.RuneCountInString(p.ImageAsset) > 512 || p.Priority < 1 || p.Priority > 50 || seen[p.Priority] {
			return bad("invalid prize")
		}
		seen[p.Priority] = true
		switch p.Type {
		case "cash", "brand_credit", "planet_points", "discount_coupon":
		default:
			return bad("invalid prize type")
		}
		if len(p.ValueRial) < 1 || len(p.ValueRial) > 15 {
			return bad("invalid prize value")
		}
		for _, c := range p.ValueRial {
			if c < '0' || c > '9' {
				return bad("invalid prize value")
			}
		}
		if _, err := strconv.ParseInt(p.ValueRial, 10, 64); err != nil {
			return bad("invalid prize value")
		}
	}
	return start, draw, nil
}

func lotteryDraftCreate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in lotteryDraftInput
	if err = decodeLotteryPayload(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery creation unavailable")
	}
	defer tx.Rollback()
	var minPrice, maxPrice int
	var allowed bool
	err = tx.QueryRowContext(ctx, `SELECT s.ticket_price_min,s.ticket_price_max,EXISTS(SELECT 1 FROM jazebe_active_brand_admin a WHERE a.brand_id=p.brand_id AND a.user_id=$2 AND a.is_active AND a.role IN ('main','marketing')) FROM jazebe_planet p JOIN jazebe_planet_settings s ON s.planet_id=p.id WHERE p.id=$1`, in.PlanetID, actor).Scan(&minPrice, &maxPrice, &allowed)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "lottery policy unavailable")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	start, draw, err := validateLotteryDraft(in, time.Now(), minPrice, maxPrice)
	if err != nil {
		return "", err
	}
	id := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_lottery(id,planet_id,title,short_description,long_description,banner_asset,starts_at,draw_at,ticket_price_points,per_user_ticket_cap,per_group_ticket_cap,status) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'draft')`, id, in.PlanetID, strings.TrimSpace(in.Title), nullableText(in.ShortDescription), nullableText(in.LongDescription), nullableText(in.BannerAsset), start, draw, in.TicketPricePoints, in.PerUserTicketCap, in.PerGroupTicketCap)
	if err != nil {
		logger.Error("Lottery insert: %v", err)
		return "", status.Error(codes.Internal, "lottery creation failed")
	}
	for _, p := range in.Prizes {
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_lottery_prize(id,planet_id,lottery_id,title,description,type,value_rial,priority,image_asset) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, uuid.Must(uuid.NewV4()).String(), in.PlanetID, id, strings.TrimSpace(p.Title), nullableText(p.Description), p.Type, p.ValueRial, p.Priority, nullableText(p.ImageAsset))
		if err != nil {
			return "", status.Error(codes.Internal, "lottery prize insert failed")
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params) VALUES('lottery.create','brand_admin',$1,'lottery',$2,$3,'{"status":"draft"}'::jsonb)`, actor, id, in.PlanetID); err != nil {
		return "", status.Error(codes.Internal, "lottery audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "lottery creation failed")
	}
	return encode(map[string]any{"id": id, "planetId": in.PlanetID, "status": "draft"})
}

func lotteryActivate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in lotteryRef
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	in.LotteryID, err = requiredUUID(in.LotteryID)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery activation unavailable")
	}
	defer tx.Rollback()
	var state string
	var start, draw time.Time
	var price, minPrice, maxPrice, prizes int
	var allowed bool
	err = tx.QueryRowContext(ctx, `SELECT l.status,l.starts_at,l.draw_at,l.ticket_price_points,s.ticket_price_min,s.ticket_price_max,EXISTS(SELECT 1 FROM jazebe_active_brand_admin a WHERE a.brand_id=p.brand_id AND a.user_id=$3 AND a.is_active AND a.role IN ('main','marketing')),(SELECT count(*) FROM jazebe_lottery_prize pr WHERE pr.lottery_id=l.id AND pr.deleted_at IS NULL) FROM jazebe_lottery l JOIN jazebe_planet p ON p.id=l.planet_id JOIN jazebe_planet_settings s ON s.planet_id=p.id WHERE l.id=$1 AND l.planet_id=$2 FOR UPDATE OF l`, in.LotteryID, in.PlanetID, actor).Scan(&state, &start, &draw, &price, &minPrice, &maxPrice, &allowed, &prizes)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "lottery not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "lottery lookup failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if state != "draft" {
		return "", status.Error(codes.FailedPrecondition, "lottery is not draft")
	}
	if !draw.After(time.Now()) || !draw.After(start) || price < minPrice || price > maxPrice || prizes < 1 {
		return "", status.Error(codes.FailedPrecondition, "lottery configuration is invalid")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery SET status='active',updated_at=now() WHERE id=$1`, in.LotteryID); err != nil {
		return "", status.Error(codes.Internal, "lottery activation failed")
	}
	adopted, err := adoptStrandedLotteryTickets(ctx, tx, in.PlanetID, in.LotteryID, time.Now().UTC())
	if err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params) VALUES('lottery.activate','brand_admin',$1,'lottery',$2,$3,jsonb_build_object('drawAt',$4::timestamptz,'adoptedTickets',$5::bigint))`, actor, in.LotteryID, in.PlanetID, draw, adopted); err != nil {
		return "", status.Error(codes.Internal, "lottery audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "lottery activation failed")
	}
	return encode(map[string]any{"id": in.LotteryID, "planetId": in.PlanetID, "status": "active"})
}
