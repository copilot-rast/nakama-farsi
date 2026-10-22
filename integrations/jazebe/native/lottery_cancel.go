package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lotteryCancelInput struct {
	PlanetID  string `json:"planetId"`
	LotteryID string `json:"lotteryId"`
	Reason    string `json:"reason"`
}
type cancelledTicket struct{ ID, GroupID, CreatorType, CreatorUserID, Source, SourceID string }
type ticketBundle struct {
	Ticket cancelledTicket
	Count  int
}

func validateLotteryCancel(in lotteryCancelInput) (lotteryCancelInput, error) {
	var err error
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return in, err
	}
	in.LotteryID, err = requiredUUID(in.LotteryID)
	if err != nil {
		return in, err
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if !utf8.ValidString(in.Reason) || utf8.RuneCountInString(in.Reason) < 3 || utf8.RuneCountInString(in.Reason) > 512 {
		return in, status.Error(codes.InvalidArgument, "cancellation reason must contain 3 to 512 characters")
	}
	return in, nil
}
func bundleCancelledTickets(tickets []cancelledTicket) []ticketBundle {
	bundles := make([]ticketBundle, 0)
	byKey := map[string]int{}
	for _, t := range tickets {
		key := strings.Join([]string{t.GroupID, t.CreatorType, t.CreatorUserID, t.Source, t.SourceID}, "|")
		if i, ok := byKey[key]; ok {
			bundles[i].Count++
		} else {
			byKey[key] = len(bundles)
			bundles = append(bundles, ticketBundle{Ticket: t, Count: 1})
		}
	}
	return bundles
}

func lotteryCancel(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in lotteryCancelInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in, err = validateLotteryCancel(in)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery cancellation unavailable")
	}
	defer tx.Rollback()
	var state, title string
	var price int
	var allowed bool
	err = tx.QueryRowContext(ctx, `SELECT l.status,l.title,l.ticket_price_points,EXISTS(SELECT 1 FROM jazebe_planet p JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id WHERE p.id=l.planet_id AND a.user_id=$3 AND a.is_active AND a.role IN ('main','marketing')) FROM jazebe_lottery l WHERE l.id=$1 AND l.planet_id=$2 FOR UPDATE OF l`, in.LotteryID, in.PlanetID, actor).Scan(&state, &title, &price, &allowed)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "lottery not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "lottery lookup failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if state == "cancelled" {
		return lotteryCancelReplay(ctx, tx, in)
	}
	if state != "draft" && state != "active" {
		return "", status.Error(codes.FailedPrecondition, "lottery cannot be cancelled")
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,group_id,creator_type,COALESCE(creator_user_id::text,''),source,COALESCE(source_id::text,'') FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND status='active' ORDER BY serial`, in.LotteryID)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery tickets unavailable")
	}
	all := make([]cancelledTicket, 0)
	purchased := make([]cancelledTicket, 0)
	earned := make([]cancelledTicket, 0)
	for rows.Next() {
		var t cancelledTicket
		if err = rows.Scan(&t.ID, &t.GroupID, &t.CreatorType, &t.CreatorUserID, &t.Source, &t.SourceID); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "lottery tickets unavailable")
		}
		all = append(all, t)
		if t.Source == "purchase" {
			purchased = append(purchased, t)
		} else {
			earned = append(earned, t)
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "lottery tickets unavailable")
	}
	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery SET status='cancelled',cancel_reason=$2,cancelled_at=$3,updated_at=$3 WHERE id=$1`, in.LotteryID, in.Reason, now); err != nil {
		return "", status.Error(codes.Internal, "lottery cancellation failed")
	}
	refunded := int64(0)
	for _, t := range purchased {
		if t.CreatorUserID == "" {
			continue
		}
		paid := price
		if t.SourceID != "" {
			var amount int
			err = tx.QueryRowContext(ctx, `SELECT paid_points FROM jazebe_order WHERE id=$1`, t.SourceID).Scan(&amount)
			if err == nil {
				paid = amount
			} else if !errors.Is(err, sql.ErrNoRows) {
				return "", status.Error(codes.Internal, "ticket order lookup failed")
			}
		}
		if paid <= 0 {
			continue
		}
		movement, err := applyPointMovement(ctx, tx, in.PlanetID, t.CreatorUserID, "credit", int64(paid), "lottery_refund", t.ID, "Lottery ticket refund: "+title, now)
		if err != nil {
			return "", status.Error(codes.Internal, "lottery refund failed")
		}
		if movement.Applied {
			refunded += int64(paid)
		}
	}
	var targetID, targetTitle string
	if len(earned) > 0 {
		err = tx.QueryRowContext(ctx, `SELECT id,title FROM jazebe_lottery WHERE planet_id=$1 AND id<>$2 AND status='active' AND draw_at>$3 ORDER BY draw_at ASC,id ASC LIMIT 1`, in.PlanetID, in.LotteryID, now).Scan(&targetID, &targetTitle)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.Internal, "next lottery lookup failed")
		}
	}
	if targetID != "" {
		for _, b := range bundleCancelledTickets(earned) {
			_, err = issueLotteryTickets(ctx, tx, ticketIssue{LotteryID: targetID, PlanetID: in.PlanetID, GroupID: b.Ticket.GroupID, CreatorType: b.Ticket.CreatorType, UserID: b.Ticket.CreatorUserID, Source: b.Ticket.Source, SourceID: b.Ticket.SourceID, Count: b.Count, Transfer: true}, now)
			if err != nil {
				return "", err
			}
		}
	}
	voided := purchased
	if targetID != "" {
		voided = append(voided, earned...)
	}
	if err = voidLotteryTickets(ctx, tx, in.LotteryID, voided); err != nil {
		return "", err
	}
	notified, err := notifyLotteryCancellation(ctx, tx, in.PlanetID, in.LotteryID, title, in.Reason, targetTitle, all, now)
	if err != nil {
		return "", err
	}
	moved := 0
	if targetID != "" {
		moved = len(earned)
	}
	var movedTo any
	if targetID != "" {
		movedTo = targetID
	}
	params, _ := json.Marshal(map[string]any{"reason": in.Reason, "refundedTickets": len(purchased), "refundedPoints": refunded, "movedTickets": moved, "movedTo": movedTo, "unmovedTickets": len(earned) - moved})
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params) VALUES('lottery.cancel','brand_admin',$1,'lottery',$2,$3,$4::jsonb)`, actor, in.LotteryID, in.PlanetID, string(params)); err != nil {
		return "", status.Error(codes.Internal, "lottery cancellation audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "lottery cancellation failed")
	}
	var target any
	if targetID != "" {
		target = map[string]any{"id": targetID, "title": targetTitle}
	}
	return encode(map[string]any{"lotteryId": in.LotteryID, "status": "cancelled", "refundedTickets": len(purchased), "refundedPoints": refunded, "movedTickets": moved, "movedTo": target, "unmovedTickets": len(earned) - moved, "notifiedUsers": notified})
}

func voidLotteryTickets(ctx context.Context, tx *sql.Tx, lotteryID string, tickets []cancelledTicket) error {
	if len(tickets) == 0 {
		return nil
	}
	grouped := map[string][3]int{}
	for _, t := range tickets {
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_lottery_ticket SET status='void' WHERE id=$1 AND status='active'`, t.ID); err != nil {
			return status.Error(codes.Internal, "ticket void failed")
		}
		counts := grouped[t.GroupID]
		counts[0]++
		if t.CreatorType == "member" {
			counts[1]++
		} else {
			counts[2]++
		}
		grouped[t.GroupID] = counts
	}
	for groupID, c := range grouped {
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_lottery_group_stat SET ticket_count=ticket_count-$3,individual_ticket_count=individual_ticket_count-$4,collective_ticket_count=collective_ticket_count-$5,updated_at=now() WHERE lottery_id=$1 AND group_id=$2`, lotteryID, groupID, c[0], c[1], c[2]); err != nil {
			return status.Error(codes.Internal, "lottery pool decrement failed")
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_lottery SET total_ticket_count=total_ticket_count-$2,updated_at=now() WHERE id=$1`, lotteryID, len(tickets)); err != nil {
		return status.Error(codes.Internal, "lottery total decrement failed")
	}
	return nil
}

func notifyLotteryCancellation(ctx context.Context, tx *sql.Tx, planetID, lotteryID, title, reason, targetTitle string, tickets []cancelledTicket, now time.Time) (int, error) {
	if len(tickets) == 0 {
		return 0, nil
	}
	recipients := map[string]bool{}
	for _, t := range tickets {
		if t.CreatorUserID != "" {
			recipients[t.CreatorUserID] = true
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT gm.user_id FROM jazebe_group_membership gm WHERE gm.status='active' AND gm.group_id IN (SELECT DISTINCT group_id FROM jazebe_lottery_ticket WHERE lottery_id=$1)`, lotteryID)
	if err != nil {
		return 0, status.Error(codes.Internal, "lottery notice recipients unavailable")
	}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return 0, status.Error(codes.Internal, "lottery notice recipients unavailable")
		}
		recipients[id] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, status.Error(codes.Internal, "lottery notice recipients unavailable")
	}
	fate := "بلیت‌های کسب‌شده تا فعال شدن قرعه‌کشی بعدی نگه‌داری می‌شوند."
	if targetTitle != "" {
		fate = "بلیت‌های کسب‌شده به قرعه‌کشی «" + targetTitle + "» منتقل شدند."
	}
	for id := range recipients {
		if err = persistJazebeNotification(ctx, tx, id, "lottery-cancel:"+lotteryID, "قرعه‌کشی لغو شد", map[string]any{"category": "critical", "priority": "urgent", "actionLink": "/raffles", "lotteryId": lotteryID, "planetId": planetID, "reason": reason, "body": fmt.Sprintf("قرعه‌کشی «%s» لغو شد. دلیل: %s. بلیت‌های خریداری‌شده به‌صورت امتیاز بازگردانده شدند. %s", title, reason, fate)}, now); err != nil {
			return 0, status.Error(codes.Internal, "lottery notice failed")
		}
	}
	return len(recipients), nil
}

func lotteryCancelReplay(ctx context.Context, tx *sql.Tx, in lotteryCancelInput) (string, error) {
	var params []byte
	err := tx.QueryRowContext(ctx, `SELECT params FROM jazebe_audit_log WHERE action='lottery.cancel' AND target_id=$1 ORDER BY created_at DESC LIMIT 1`, in.LotteryID).Scan(&params)
	if err != nil {
		return "", status.Error(codes.Internal, "cancellation history unavailable")
	}
	var p struct {
		RefundedTickets int     `json:"refundedTickets"`
		RefundedPoints  int64   `json:"refundedPoints"`
		MovedTickets    int     `json:"movedTickets"`
		MovedTo         *string `json:"movedTo"`
		UnmovedTickets  int     `json:"unmovedTickets"`
	}
	if json.Unmarshal(params, &p) != nil {
		return "", status.Error(codes.Internal, "cancellation history invalid")
	}
	var target any
	if p.MovedTo != nil {
		var title string
		if tx.QueryRowContext(ctx, `SELECT title FROM jazebe_lottery WHERE id=$1`, *p.MovedTo).Scan(&title) == nil {
			target = map[string]any{"id": *p.MovedTo, "title": title}
		}
	}
	return encode(map[string]any{"lotteryId": in.LotteryID, "status": "cancelled", "refundedTickets": p.RefundedTickets, "refundedPoints": p.RefundedPoints, "movedTickets": p.MovedTickets, "movedTo": target, "unmovedTickets": p.UnmovedTickets, "notifiedUsers": 0})
}

// Earned tickets stranded in cancelled pools follow the next lottery once it
// becomes active. Row locks prevent two activations from adopting the same row.
func adoptStrandedLotteryTickets(ctx context.Context, tx *sql.Tx, planetID, targetID string, now time.Time) (int, error) {
	rows, err := tx.QueryContext(ctx, `SELECT t.id,t.group_id,t.creator_type,COALESCE(t.creator_user_id::text,''),t.source,COALESCE(t.source_id::text,''),t.lottery_id FROM jazebe_lottery_ticket t JOIN jazebe_lottery l ON l.id=t.lottery_id WHERE t.planet_id=$1 AND t.status='active' AND t.source<>'purchase' AND l.status='cancelled' AND l.id<>$2 ORDER BY t.serial FOR UPDATE OF t`, planetID, targetID)
	if err != nil {
		return 0, status.Error(codes.Internal, "stranded ticket lookup failed")
	}
	type stranded struct {
		Ticket    cancelledTicket
		LotteryID string
	}
	found := make([]stranded, 0)
	for rows.Next() {
		var r stranded
		if err = rows.Scan(&r.Ticket.ID, &r.Ticket.GroupID, &r.Ticket.CreatorType, &r.Ticket.CreatorUserID, &r.Ticket.Source, &r.Ticket.SourceID, &r.LotteryID); err != nil {
			rows.Close()
			return 0, status.Error(codes.Internal, "stranded ticket lookup failed")
		}
		found = append(found, r)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, status.Error(codes.Internal, "stranded ticket lookup failed")
	}
	if len(found) == 0 {
		return 0, nil
	}
	bySource := map[string][]cancelledTicket{}
	bundles := make([]cancelledTicket, 0, len(found))
	for _, r := range found {
		bySource[r.LotteryID] = append(bySource[r.LotteryID], r.Ticket)
		bundles = append(bundles, r.Ticket)
	}
	for _, b := range bundleCancelledTickets(bundles) {
		if _, err = issueLotteryTickets(ctx, tx, ticketIssue{LotteryID: targetID, PlanetID: planetID, GroupID: b.Ticket.GroupID, CreatorType: b.Ticket.CreatorType, UserID: b.Ticket.CreatorUserID, Source: b.Ticket.Source, SourceID: b.Ticket.SourceID, Count: b.Count, Transfer: true}, now); err != nil {
			return 0, err
		}
	}
	for sourceID, tickets := range bySource {
		if err = voidLotteryTickets(ctx, tx, sourceID, tickets); err != nil {
			return 0, err
		}
	}
	return len(found), nil
}
