package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func lotteryAdminTickets(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var in lotteryAdminTicketsInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	var err error
	in.LotteryID, err = requiredUUID(in.LotteryID)
	if err != nil {
		return "", err
	}
	in.Page, in.Limit, err = lotteryAdminPage(in.Page, in.Limit)
	if err != nil {
		return "", err
	}
	in.Search = strings.TrimSpace(in.Search)
	if len(in.Search) > 128 {
		return "", status.Error(codes.InvalidArgument, "search too long")
	}
	var exists bool
	if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_lottery WHERE id=$1)`, in.LotteryID).Scan(&exists); err != nil {
		return "", status.Error(codes.Internal, "lottery lookup failed")
	}
	if !exists {
		return "", status.Error(codes.NotFound, "lottery not found")
	}
	var total int
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_lottery_ticket t JOIN jazebe_group g ON g.id=t.group_id WHERE t.lottery_id=$1 AND ($2='' OR t.serial ILIKE '%'||$2||'%' OR g.name ILIKE '%'||$2||'%')`, in.LotteryID, in.Search).Scan(&total)
	if err != nil {
		return "", status.Error(codes.Internal, "ticket audit unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT t.id,t.serial,t.group_id,g.name,t.creator_type,t.creator_user_id,COALESCE(NULLIF(trim(u.display_name),''),'کاربر'),t.source,t.status,t.created_at,p.priority FROM jazebe_lottery_ticket t JOIN jazebe_group g ON g.id=t.group_id LEFT JOIN users u ON u.id=t.creator_user_id LEFT JOIN jazebe_lottery_winner w ON w.ticket_id=t.id LEFT JOIN jazebe_lottery_prize p ON p.id=w.prize_id WHERE t.lottery_id=$1 AND ($2='' OR t.serial ILIKE '%'||$2||'%' OR g.name ILIKE '%'||$2||'%') ORDER BY t.serial LIMIT $3 OFFSET $4`, in.LotteryID, in.Search, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "ticket audit unavailable")
	}
	items := []map[string]any{}
	for rows.Next() {
		var id, serial, groupID, groupName, creatorType, creatorName, source, state string
		var creator sql.NullString
		var priority sql.NullInt64
		var at time.Time
		if err = rows.Scan(&id, &serial, &groupID, &groupName, &creatorType, &creator, &creatorName, &source, &state, &at, &priority); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "ticket audit unavailable")
		}
		var creatorView, won any
		if creator.Valid {
			creatorView = map[string]any{"userId": creator.String, "name": creatorName}
		}
		if priority.Valid {
			won = priority.Int64
		}
		items = append(items, map[string]any{"id": id, "serial": serial, "group": map[string]any{"id": groupID, "name": groupName}, "creatorType": creatorType, "creator": creatorView, "source": source, "status": state, "createdAt": at.Format(time.RFC3339Nano), "wonPriority": won})
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "ticket audit unavailable")
	}
	return encode(map[string]any{"items": items, "total": total, "page": in.Page, "limit": in.Limit})
}

func lotteryAdminReport(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
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
	rows, err := db.QueryContext(ctx, `SELECT s.group_id,g.name,s.ticket_count,s.individual_ticket_count,s.collective_ticket_count FROM jazebe_lottery_group_stat s JOIN jazebe_group g ON g.id=s.group_id WHERE s.lottery_id=$1 AND s.ticket_count>0 ORDER BY s.ticket_count DESC`, in.LotteryID)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery report pool unavailable")
	}
	pool := []map[string]any{}
	for rows.Next() {
		var id, name string
		var tickets, individual, collective int
		if err = rows.Scan(&id, &name, &tickets, &individual, &collective); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "lottery report pool unavailable")
		}
		pool = append(pool, map[string]any{"groupId": id, "name": name, "tickets": tickets, "individual": individual, "collective": collective})
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "lottery report pool unavailable")
	}
	var replay any
	lottery := audit["lottery"].(map[string]any)
	if lottery["status"] == "finished" {
		raw, err := lotteryAdminReplay(ctx, logger, db, nk, payload)
		if err != nil {
			return "", err
		}
		if json.Unmarshal([]byte(raw), &replay) != nil {
			return "", status.Error(codes.Internal, "replay response invalid")
		}
	}
	audit["generatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
	audit["pool"] = pool
	audit["replay"] = replay
	return encode(audit)
}
