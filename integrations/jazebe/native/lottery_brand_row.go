package main

import (
	"context"
	"database/sql"
	"errors"
	"math/big"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Source brand cards and detail use the same full record projection.
func lotteryBrandRow(ctx context.Context, db *sql.DB, planetID, lotteryID string) (map[string]any, error) {
	var title, state string
	var short, long, banner, reason sql.NullString
	var start, draw, created, updated time.Time
	var drawn, cancelled sql.NullTime
	var price, tickets int
	var userCap, groupCap sql.NullInt64
	err := db.QueryRowContext(ctx, `SELECT title,status,short_description,long_description,banner_asset,starts_at,draw_at,ticket_price_points,total_ticket_count,per_user_ticket_cap,per_group_ticket_cap,drawn_at,cancelled_at,cancel_reason,created_at,updated_at FROM jazebe_lottery WHERE id=$1 AND planet_id=$2`, lotteryID, planetID).Scan(&title, &state, &short, &long, &banner, &start, &draw, &price, &tickets, &userCap, &groupCap, &drawn, &cancelled, &reason, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "lottery not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery lookup failed")
	}
	var groupCount int
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_lottery_group_stat WHERE lottery_id=$1 AND ticket_count>0`, lotteryID).Scan(&groupCount); err != nil {
		return nil, status.Error(codes.Internal, "lottery group count failed")
	}
	rows, err := db.QueryContext(ctx, `SELECT id,title,description,type,value_rial::text,priority,image_asset FROM jazebe_lottery_prize WHERE lottery_id=$1 AND deleted_at IS NULL ORDER BY priority`, lotteryID)
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery prizes failed")
	}
	prizes := []map[string]any{}
	valueSum := new(big.Int)
	for rows.Next() {
		var id, name, kind, value string
		var description, image sql.NullString
		var priority int
		if err = rows.Scan(&id, &name, &description, &kind, &value, &priority, &image); err != nil {
			rows.Close()
			return nil, status.Error(codes.Internal, "lottery prizes failed")
		}
		n, ok := new(big.Int).SetString(value, 10)
		if !ok {
			rows.Close()
			return nil, status.Error(codes.Internal, "lottery prize value invalid")
		}
		valueSum.Add(valueSum, n)
		prizes = append(prizes, map[string]any{"id": id, "title": name, "description": nullString(description), "type": kind, "valueRial": value, "priority": priority, "imageAsset": nullString(image)})
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery prizes failed")
	}
	var drawnAt, cancelledAt any
	if drawn.Valid {
		drawnAt = drawn.Time.Format(time.RFC3339Nano)
	}
	if cancelled.Valid {
		cancelledAt = cancelled.Time.Format(time.RFC3339Nano)
	}
	return map[string]any{"id": lotteryID, "planetId": planetID, "title": title, "shortDescription": nullString(short), "longDescription": nullString(long), "bannerAsset": nullString(banner), "startsAt": start.Format(time.RFC3339Nano), "drawAt": draw.Format(time.RFC3339Nano), "ticketPricePoints": price, "perUserTicketCap": nullInt(userCap), "perGroupTicketCap": nullInt(groupCap), "status": state, "drawnAt": drawnAt, "cancelledAt": cancelledAt, "cancelReason": nullString(reason), "totalTicketCount": tickets, "groupCount": groupCount, "prizeCount": len(prizes), "prizeValueRial": valueSum.String(), "prizes": prizes, "createdAt": created.Format(time.RFC3339Nano), "updatedAt": updated.Format(time.RFC3339Nano)}, nil
}
