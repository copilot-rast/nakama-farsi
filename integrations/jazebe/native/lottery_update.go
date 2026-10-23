package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lotteryUpdateInput struct {
	PlanetID          string          `json:"planetId"`
	LotteryID         string          `json:"lotteryId"`
	Title             json.RawMessage `json:"title"`
	ShortDescription  json.RawMessage `json:"shortDescription"`
	LongDescription   json.RawMessage `json:"longDescription"`
	BannerAsset       json.RawMessage `json:"bannerAsset"`
	StartsAt          json.RawMessage `json:"startsAt"`
	DrawAt            json.RawMessage `json:"drawAt"`
	TicketPricePoints json.RawMessage `json:"ticketPricePoints"`
	PerUserTicketCap  json.RawMessage `json:"perUserTicketCap"`
	PerGroupTicketCap json.RawMessage `json:"perGroupTicketCap"`
	Prizes            json.RawMessage `json:"prizes"`
}

func decodeLotteryString(raw json.RawMessage, target *string) error {
	if raw == nil {
		return nil
	}
	if string(raw) == "null" {
		*target = ""
		return nil
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return status.Error(codes.InvalidArgument, "invalid lottery text")
	}
	return nil
}
func decodeLotteryInt(raw json.RawMessage, target *int) error {
	if raw == nil {
		return nil
	}
	if string(raw) == "null" {
		return status.Error(codes.InvalidArgument, "required lottery number cannot be null")
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return status.Error(codes.InvalidArgument, "invalid lottery number")
	}
	return nil
}
func decodeLotteryCap(raw json.RawMessage, target **int) error {
	if raw == nil {
		return nil
	}
	if string(raw) == "null" {
		*target = nil
		return nil
	}
	var value int
	if err := json.Unmarshal(raw, &value); err != nil {
		return status.Error(codes.InvalidArgument, "invalid lottery cap")
	}
	*target = &value
	return nil
}

// A draft accepts structural edits. An active lottery accepts only texts and
// banner, preserving the schedule and odds that members entered under.
func lotteryBrandUpdate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var req lotteryUpdateInput
	if err = decodeLotteryPayload(payload, &req); err != nil {
		return "", err
	}
	req.PlanetID, err = requiredUUID(req.PlanetID)
	if err != nil {
		return "", err
	}
	req.LotteryID, err = requiredUUID(req.LotteryID)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery update unavailable")
	}
	defer tx.Rollback()
	var draft lotteryDraftInput
	var state string
	var short, long, banner sql.NullString
	var start, draw time.Time
	var userCap, groupCap sql.NullInt64
	var minPrice, maxPrice int
	var allowed bool
	err = tx.QueryRowContext(ctx, `SELECT l.title,l.short_description,l.long_description,l.banner_asset,l.starts_at,l.draw_at,l.ticket_price_points,l.per_user_ticket_cap,l.per_group_ticket_cap,l.status,s.ticket_price_min,s.ticket_price_max,EXISTS(SELECT 1 FROM jazebe_brand_admin a WHERE a.brand_id=p.brand_id AND a.user_id=$3 AND a.is_active AND a.role IN ('main','marketing')) FROM jazebe_lottery l JOIN jazebe_planet p ON p.id=l.planet_id JOIN jazebe_planet_settings s ON s.planet_id=p.id WHERE l.id=$1 AND l.planet_id=$2 FOR UPDATE OF l`, req.LotteryID, req.PlanetID, actor).Scan(&draft.Title, &short, &long, &banner, &start, &draw, &draft.TicketPricePoints, &userCap, &groupCap, &state, &minPrice, &maxPrice, &allowed)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "lottery not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "lottery lookup failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	structural := req.StartsAt != nil || req.DrawAt != nil || req.TicketPricePoints != nil || req.PerUserTicketCap != nil || req.PerGroupTicketCap != nil || req.Prizes != nil
	if state != "draft" && (state != "active" || structural) {
		return "", status.Error(codes.FailedPrecondition, "lottery is not editable")
	}
	draft.PlanetID = req.PlanetID
	draft.ShortDescription = short.String
	draft.LongDescription = long.String
	draft.BannerAsset = banner.String
	draft.StartsAt = start.Format(time.RFC3339Nano)
	draft.DrawAt = draw.Format(time.RFC3339Nano)
	if userCap.Valid {
		n := int(userCap.Int64)
		draft.PerUserTicketCap = &n
	}
	if groupCap.Valid {
		n := int(groupCap.Int64)
		draft.PerGroupTicketCap = &n
	}
	if err = decodeLotteryString(req.Title, &draft.Title); err != nil {
		return "", err
	}
	if err = decodeLotteryString(req.ShortDescription, &draft.ShortDescription); err != nil {
		return "", err
	}
	if err = decodeLotteryString(req.LongDescription, &draft.LongDescription); err != nil {
		return "", err
	}
	if err = decodeLotteryString(req.BannerAsset, &draft.BannerAsset); err != nil {
		return "", err
	}
	if err = decodeLotteryString(req.StartsAt, &draft.StartsAt); err != nil {
		return "", err
	}
	if err = decodeLotteryString(req.DrawAt, &draft.DrawAt); err != nil {
		return "", err
	}
	if err = decodeLotteryInt(req.TicketPricePoints, &draft.TicketPricePoints); err != nil {
		return "", err
	}
	if err = decodeLotteryCap(req.PerUserTicketCap, &draft.PerUserTicketCap); err != nil {
		return "", err
	}
	if err = decodeLotteryCap(req.PerGroupTicketCap, &draft.PerGroupTicketCap); err != nil {
		return "", err
	}
	rows, err := tx.QueryContext(ctx, `SELECT title,COALESCE(description,''),type,value_rial,priority,COALESCE(image_asset,'') FROM jazebe_lottery_prize WHERE lottery_id=$1 AND deleted_at IS NULL ORDER BY priority`, req.LotteryID)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery prizes unavailable")
	}
	prizes := make([]lotteryPrizeInput, 0)
	for rows.Next() {
		var p lotteryPrizeInput
		if err = rows.Scan(&p.Title, &p.Description, &p.Type, &p.ValueRial, &p.Priority, &p.ImageAsset); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "lottery prizes unavailable")
		}
		prizes = append(prizes, p)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "lottery prizes unavailable")
	}
	draft.Prizes = prizes
	if req.Prizes != nil {
		if string(req.Prizes) == "null" || json.Unmarshal(req.Prizes, &draft.Prizes) != nil {
			return "", status.Error(codes.InvalidArgument, "invalid prizes")
		}
	}
	validatedStart, validatedDraw := start, draw
	if state == "draft" && structural {
		validatedStart, validatedDraw, err = validateLotteryDraft(draft, time.Now(), minPrice, maxPrice)
		if err != nil {
			return "", err
		}
	} else if !utf8.ValidString(draft.Title) || utf8.RuneCountInString(strings.TrimSpace(draft.Title)) < 1 || utf8.RuneCountInString(draft.Title) > 128 || !utf8.ValidString(draft.ShortDescription) || utf8.RuneCountInString(draft.ShortDescription) > 512 || !utf8.ValidString(draft.LongDescription) || utf8.RuneCountInString(draft.LongDescription) > 8000 || !utf8.ValidString(draft.BannerAsset) || utf8.RuneCountInString(draft.BannerAsset) > 512 {
		return "", status.Error(codes.InvalidArgument, "invalid lottery text")
	}
	_, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery SET title=$2,short_description=$3,long_description=$4,banner_asset=$5,starts_at=$6,draw_at=$7,ticket_price_points=$8,per_user_ticket_cap=$9,per_group_ticket_cap=$10,updated_at=now() WHERE id=$1`, req.LotteryID, strings.TrimSpace(draft.Title), nullableText(draft.ShortDescription), nullableText(draft.LongDescription), nullableText(draft.BannerAsset), validatedStart, validatedDraw, draft.TicketPricePoints, draft.PerUserTicketCap, draft.PerGroupTicketCap)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery update failed")
	}
	if req.Prizes != nil {
		keep := make([]int, 0, len(draft.Prizes))
		for _, p := range draft.Prizes {
			keep = append(keep, p.Priority)
			_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_lottery_prize(id,planet_id,lottery_id,title,description,type,value_rial,priority,image_asset) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(lottery_id,priority) DO UPDATE SET title=EXCLUDED.title,description=EXCLUDED.description,type=EXCLUDED.type,value_rial=EXCLUDED.value_rial,image_asset=EXCLUDED.image_asset,deleted_at=NULL`, uuid.Must(uuid.NewV4()).String(), req.PlanetID, req.LotteryID, strings.TrimSpace(p.Title), nullableText(p.Description), p.Type, p.ValueRial, p.Priority, nullableText(p.ImageAsset))
			if err != nil {
				return "", status.Error(codes.Internal, "lottery prize update failed")
			}
		}
		// At most 50 slots, so a bounded array keeps the soft deletion in one SQL statement.
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery_prize SET deleted_at=now() WHERE lottery_id=$1 AND NOT(priority=ANY($2::smallint[])) AND deleted_at IS NULL`, req.LotteryID, smallintArray(keep)); err != nil {
			return "", status.Error(codes.Internal, "lottery prize update failed")
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params) VALUES('lottery.update','brand_admin',$1,'lottery',$2,$3,jsonb_build_object('structural',$4::boolean))`, actor, req.LotteryID, req.PlanetID, structural); err != nil {
		return "", status.Error(codes.Internal, "lottery audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "lottery update failed")
	}
	return encode(map[string]any{"id": req.LotteryID, "planetId": req.PlanetID, "status": state})
}

type smallintArray []int

func (a smallintArray) Value() (driver.Value, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i, n := range a {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(n))
	}
	b.WriteByte('}')
	return b.String(), nil
}
