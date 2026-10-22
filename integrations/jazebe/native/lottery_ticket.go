package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ticketIssue struct {
	LotteryID, PlanetID, GroupID, UserID, Source, SourceID string
	CreatorType                                            string
	Count                                                  int
	Transfer                                               bool
}

func validTicketSource(s string) bool {
	switch s {
	case "purchase", "solo_mission", "group_mission", "game", "previous_lottery_prize", "medal":
		return true
	}
	return false
}

// issueLotteryTickets runs only inside the caller's transaction. The lottery
// row serializes cap checks, counter changes, and issuance across instances.
func issueLotteryTickets(ctx context.Context, tx *sql.Tx, in ticketIssue, now time.Time) ([]string, error) {
	if in.Count < 1 || in.Count > 100000 || !validTicketSource(in.Source) || (in.CreatorType != "member" && in.CreatorType != "group") {
		return nil, status.Error(codes.InvalidArgument, "invalid ticket request")
	}
	if in.CreatorType == "member" && in.UserID == "" || in.CreatorType == "group" && in.UserID != "" || in.GroupID == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket creator or group missing")
	}
	var state string
	var starts, draw time.Time
	var userCap, groupCap sql.NullInt64
	err := tx.QueryRowContext(ctx, `SELECT status,starts_at,draw_at,per_user_ticket_cap,per_group_ticket_cap FROM jazebe_lottery WHERE id=$1 AND planet_id=$2 FOR UPDATE`, in.LotteryID, in.PlanetID).Scan(&state, &starts, &draw, &userCap, &groupCap)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "lottery not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery lookup failed")
	}
	if state != "active" || !now.Before(draw) || !in.Transfer && now.Before(starts) {
		return nil, status.Error(codes.FailedPrecondition, "lottery not open")
	}
	var groupState string
	err = tx.QueryRowContext(ctx, `SELECT status FROM jazebe_group WHERE id=$1 AND planet_id=$2`, in.GroupID, in.PlanetID).Scan(&groupState)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "group lookup failed")
	}
	if !in.Transfer && groupState != "active" {
		return nil, status.Error(codes.FailedPrecondition, "group not active")
	}
	if in.CreatorType == "member" && !in.Transfer {
		var belongs bool
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_group_membership WHERE group_id=$1 AND planet_id=$2 AND user_id=$3 AND status='active')`, in.GroupID, in.PlanetID, in.UserID).Scan(&belongs)
		if err != nil {
			return nil, status.Error(codes.Internal, "group membership lookup failed")
		}
		if !belongs {
			return nil, status.Error(codes.PermissionDenied, "active group membership required")
		}
	}
	if !in.Transfer && userCap.Valid && in.CreatorType == "member" {
		var held int
		err = tx.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND creator_user_id=$2 AND status='active'`, in.LotteryID, in.UserID).Scan(&held)
		if err != nil {
			return nil, status.Error(codes.Internal, "ticket cap lookup failed")
		}
		if int64(held)+int64(in.Count) > userCap.Int64 {
			return nil, status.Error(codes.ResourceExhausted, "user ticket cap reached")
		}
	}
	if !in.Transfer && groupCap.Valid {
		var held int
		err = tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT ticket_count FROM jazebe_lottery_group_stat WHERE lottery_id=$1 AND group_id=$2),0)`, in.LotteryID, in.GroupID).Scan(&held)
		if err != nil {
			return nil, status.Error(codes.Internal, "group ticket cap lookup failed")
		}
		if int64(held)+int64(in.Count) > groupCap.Int64 {
			return nil, status.Error(codes.ResourceExhausted, "group ticket cap reached")
		}
	}
	if in.SourceID != "" && !in.Transfer {
		var issued bool
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND source=$2 AND source_id=$3)`, in.LotteryID, in.Source, in.SourceID).Scan(&issued)
		if err != nil {
			return nil, status.Error(codes.Internal, "ticket idempotency lookup failed")
		}
		if issued {
			return nil, status.Error(codes.AlreadyExists, "tickets already issued for source")
		}
	}
	serials := make([]string, 0, in.Count)
	for i := 0; i < in.Count; i++ {
		serial, err := newTicketSerial(now)
		if err != nil {
			return nil, status.Error(codes.Internal, "ticket serial unavailable")
		}
		var creator any
		if in.CreatorType == "member" {
			creator = in.UserID
		}
		var sourceID any
		if in.SourceID != "" {
			sourceID = in.SourceID
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_lottery_ticket(id,lottery_id,planet_id,group_id,creator_type,creator_user_id,source,source_id,serial) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, uuid.Must(uuid.NewV4()).String(), in.LotteryID, in.PlanetID, in.GroupID, in.CreatorType, creator, in.Source, sourceID, serial)
		if err != nil {
			return nil, status.Error(codes.Internal, "ticket issuance failed")
		}
		serials = append(serials, serial)
	}
	individual, collective := 0, 0
	if in.CreatorType == "member" {
		individual = in.Count
	} else {
		collective = in.Count
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_lottery_group_stat(lottery_id,group_id,ticket_count,individual_ticket_count,collective_ticket_count) VALUES($1,$2,$3,$4,$5) ON CONFLICT(lottery_id,group_id) DO UPDATE SET ticket_count=jazebe_lottery_group_stat.ticket_count+$3,individual_ticket_count=jazebe_lottery_group_stat.individual_ticket_count+$4,collective_ticket_count=jazebe_lottery_group_stat.collective_ticket_count+$5,updated_at=now()`, in.LotteryID, in.GroupID, in.Count, individual, collective)
	if err != nil {
		return nil, status.Error(codes.Internal, "ticket statistics failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery SET total_ticket_count=total_ticket_count+$2,updated_at=now() WHERE id=$1`, in.LotteryID, in.Count); err != nil {
		return nil, status.Error(codes.Internal, "lottery counter failed")
	}
	return serials, nil
}

func newTicketSerial(now time.Time) (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	serial := "T-" + strings.ToUpper(strconv.FormatInt(now.UnixMilli(), 36)) + "-" + strings.ToUpper(hex.EncodeToString(b))
	if len(serial) > 24 {
		serial = serial[:24]
	}
	return serial, nil
}
