package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lotteryAdminListInput struct {
	Status, PlanetID, Search string
	Page, Limit              int
}
type lotteryAdminTicketsInput struct {
	LotteryID, Search string
	Page, Limit       int
}

func lotteryAdminPage(page, limit int) (int, int, error) {
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 20
	}
	if page < 1 || limit < 1 || limit > 100 {
		return 0, 0, status.Error(codes.InvalidArgument, "invalid page")
	}
	return page, limit, nil
}
func lotteryAdminRef(payload string) (lotteryDrawInput, error) {
	var in lotteryDrawInput
	err := decode(payload, &in)
	if err != nil {
		return in, err
	}
	in.LotteryID, err = requiredUUID(in.LotteryID)
	return in, err
}

func lotteryAdminRow(ctx context.Context, db *sql.DB, id string) (map[string]any, error) {
	var title, state, planetID, planetName, brandID, brandName string
	var draw time.Time
	var drawn, cancelled sql.NullTime
	var tickets, prizes int
	err := db.QueryRowContext(ctx, `SELECT l.title,l.status,p.id,p.name,b.id,b.name,l.draw_at,l.drawn_at,l.cancelled_at,l.total_ticket_count,(SELECT count(*) FROM jazebe_lottery_prize pr WHERE pr.lottery_id=l.id AND pr.deleted_at IS NULL) FROM jazebe_lottery l JOIN jazebe_planet p ON p.id=l.planet_id JOIN jazebe_brand b ON b.id=p.brand_id WHERE l.id=$1`, id).Scan(&title, &state, &planetID, &planetName, &brandID, &brandName, &draw, &drawn, &cancelled, &tickets, &prizes)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "lottery not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery audit lookup failed")
	}
	var drawnAt, cancelledAt any
	if drawn.Valid {
		drawnAt = drawn.Time.Format(time.RFC3339Nano)
	}
	if cancelled.Valid {
		cancelledAt = cancelled.Time.Format(time.RFC3339Nano)
	}
	return map[string]any{"id": id, "title": title, "status": state, "planet": map[string]any{"id": planetID, "name": planetName}, "brand": map[string]any{"productId": brandID, "name": brandName}, "drawAt": draw.Format(time.RFC3339Nano), "drawnAt": drawnAt, "cancelledAt": cancelledAt, "totalTicketCount": tickets, "prizeCount": prizes}, nil
}

func lotteryAdminList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var in lotteryAdminListInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.PlanetID != "" {
		var err error
		in.PlanetID, err = requiredUUID(in.PlanetID)
		if err != nil {
			return "", err
		}
	}
	if in.Status != "" {
		switch in.Status {
		case "draft", "active", "drawing", "finished", "cancelled":
		default:
			return "", status.Error(codes.InvalidArgument, "invalid lottery status")
		}
	}
	in.Search = strings.TrimSpace(in.Search)
	if len(in.Search) > 128 {
		return "", status.Error(codes.InvalidArgument, "search too long")
	}
	var err error
	in.Page, in.Limit, err = lotteryAdminPage(in.Page, in.Limit)
	if err != nil {
		return "", err
	}
	var total int
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_lottery WHERE ($1='' OR status=$1) AND ($2::uuid IS NULL OR planet_id=$2::uuid) AND ($3='' OR title ILIKE '%'||$3||'%')`, in.Status, nullableText(in.PlanetID), in.Search).Scan(&total)
	if err != nil {
		return "", status.Error(codes.Internal, "operator lottery list failed")
	}
	rows, err := db.QueryContext(ctx, `SELECT id FROM jazebe_lottery WHERE ($1='' OR status=$1) AND ($2::uuid IS NULL OR planet_id=$2::uuid) AND ($3='' OR title ILIKE '%'||$3||'%') ORDER BY COALESCE(drawn_at,draw_at) DESC,id DESC LIMIT $4 OFFSET $5`, in.Status, nullableText(in.PlanetID), in.Search, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "operator lottery list failed")
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "operator lottery list failed")
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "operator lottery list failed")
	}
	items := []map[string]any{}
	for _, id := range ids {
		row, err := lotteryAdminRow(ctx, db, id)
		if err != nil {
			return "", err
		}
		items = append(items, row)
	}
	return encode(map[string]any{"items": items, "total": total, "page": in.Page, "limit": in.Limit})
}

func lotteryAdminAudit(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	in, err := lotteryAdminRef(payload)
	if err != nil {
		return "", err
	}
	audit, err := lotteryAuditData(ctx, db, in.LotteryID)
	if err != nil {
		return "", err
	}
	return encode(audit)
}

func lotteryAuditData(ctx context.Context, db *sql.DB, id string) (map[string]any, error) {
	row, err := lotteryAdminRow(ctx, db, id)
	if err != nil {
		return nil, err
	}
	var seed, reason sql.NullString
	var raw []byte
	var planetID string
	if err = db.QueryRowContext(ctx, `SELECT planet_id,draw_seed,draw_meta,cancel_reason FROM jazebe_lottery WHERE id=$1`, id).Scan(&planetID, &seed, &raw, &reason); err != nil {
		return nil, status.Error(codes.Internal, "lottery audit unavailable")
	}
	meta := map[string]any{}
	if len(raw) > 0 && json.Unmarshal(raw, &meta) != nil {
		return nil, status.Error(codes.Internal, "draw metadata invalid")
	}
	var publicSeed any
	if row["status"] == "finished" || row["status"] == "drawing" {
		publicSeed = nullString(seed)
	}
	algorithm := meta["algorithm"]
	if algorithm == nil {
		algorithm = lotteryDrawAlgorithm
	}
	uns := meta["unawardedPrizeIds"]
	if uns == nil {
		uns = []any{}
	}
	row["seed"] = publicSeed
	row["algorithm"] = algorithm
	row["snapshotAt"] = meta["snapshotAt"]
	row["ticketDigest"] = meta["ticketDigest"]
	row["cancelReason"] = nullString(reason)
	row["unawardedPrizeIds"] = uns
	prizeRows, err := db.QueryContext(ctx, `SELECT id,priority,title,type,value_rial::text FROM jazebe_lottery_prize WHERE lottery_id=$1 AND deleted_at IS NULL ORDER BY priority`, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery audit prizes unavailable")
	}
	prizes := []map[string]any{}
	for prizeRows.Next() {
		var id, title, kind, value string
		var rank int
		if err = prizeRows.Scan(&id, &rank, &title, &kind, &value); err != nil {
			prizeRows.Close()
			return nil, status.Error(codes.Internal, "lottery audit prizes unavailable")
		}
		prizes = append(prizes, map[string]any{"id": id, "priority": rank, "title": title, "type": kind, "valueRial": value})
	}
	err = prizeRows.Err()
	prizeRows.Close()
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery audit prizes unavailable")
	}
	winners, err := lotteryWinnerRows(ctx, db, planetID, id)
	if err != nil {
		return nil, err
	}
	trailRows, err := db.QueryContext(ctx, `SELECT id,action,actor_type,actor_id,result,params,created_at FROM jazebe_audit_log WHERE target_type='lottery' AND target_id=$1 ORDER BY created_at DESC LIMIT 200`, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery audit trail unavailable")
	}
	trail := []map[string]any{}
	for trailRows.Next() {
		var seq int64
		var action, actorType, result string
		var actor sql.NullString
		var params []byte
		var at time.Time
		if err = trailRows.Scan(&seq, &action, &actorType, &actor, &result, &params, &at); err != nil {
			trailRows.Close()
			return nil, status.Error(codes.Internal, "lottery audit trail unavailable")
		}
		var p any
		if len(params) > 0 && json.Unmarshal(params, &p) != nil {
			trailRows.Close()
			return nil, status.Error(codes.Internal, "lottery audit trail invalid")
		}
		trail = append(trail, map[string]any{"id": strconv.FormatInt(seq, 10), "action": action, "actorType": actorType, "actorId": nullString(actor), "result": result, "params": p, "createdAt": at.Format(time.RFC3339Nano)})
	}
	err = trailRows.Err()
	trailRows.Close()
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery audit trail unavailable")
	}
	return map[string]any{"lottery": row, "prizes": prizes, "winners": winners, "trail": trail}, nil
}
