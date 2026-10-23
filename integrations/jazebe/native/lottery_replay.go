package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lotteryReplayWinner struct{ PrizeID, TicketID, Serial, GroupID, WinnerID string }

// Operator replay recomputes the public draw from its recorded seed and
// snapshot, including the immutable prize-share rows. It never changes a
// winner; it writes an audit row recording match or mismatch.
func lotteryAdminReplay(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := missionOperator(ctx)
	if err != nil {
		return "", err
	}
	var in lotteryDrawInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.LotteryID, err = requiredUUID(in.LotteryID)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "replay unavailable")
	}
	defer tx.Rollback()
	var state, seed, planetID string
	var drawAt time.Time
	var raw []byte
	err = tx.QueryRowContext(ctx, `SELECT status,COALESCE(draw_seed,''),planet_id,draw_at,draw_meta FROM jazebe_lottery WHERE id=$1`, in.LotteryID).Scan(&state, &seed, &planetID, &drawAt, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "lottery not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "replay lookup failed")
	}
	if state != "finished" || seed == "" {
		return "", status.Error(codes.FailedPrecondition, "lottery not drawn")
	}
	var meta struct{ Algorithm, SnapshotAt, TicketDigest string }
	if json.Unmarshal(raw, &meta) != nil {
		return "", status.Error(codes.Internal, "draw metadata invalid")
	}
	at, err := time.Parse(time.RFC3339Nano, meta.SnapshotAt)
	if err != nil {
		at = drawAt
	}
	prizes, err := loadLotteryDrawPrizes(ctx, tx, in.LotteryID)
	if err != nil {
		return "", err
	}
	tickets, err := loadLotteryDrawTickets(ctx, tx, in.LotteryID, at)
	if err != nil {
		return "", err
	}
	pure := make([]drawPrize, len(prizes))
	for i, p := range prizes {
		pure[i] = drawPrize{p.ID, p.Priority}
	}
	out, err := drawLottery(seed, tickets, pure)
	if err != nil {
		return "", status.Error(codes.Internal, "replay computation failed")
	}
	winnerRows, err := tx.QueryContext(ctx, `SELECT w.id,w.prize_id,w.ticket_id,t.serial,w.group_id FROM jazebe_lottery_winner w JOIN jazebe_lottery_ticket t ON t.id=w.ticket_id WHERE w.lottery_id=$1`, in.LotteryID)
	if err != nil {
		return "", status.Error(codes.Internal, "recorded winners unavailable")
	}
	byPrize := map[string]lotteryReplayWinner{}
	for winnerRows.Next() {
		var w lotteryReplayWinner
		if err = winnerRows.Scan(&w.WinnerID, &w.PrizeID, &w.TicketID, &w.Serial, &w.GroupID); err != nil {
			winnerRows.Close()
			return "", status.Error(codes.Internal, "recorded winners unavailable")
		}
		byPrize[w.PrizeID] = w
	}
	err = winnerRows.Err()
	winnerRows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "recorded winners unavailable")
	}
	byPick := map[string]drawPick{}
	for _, p := range out.Picks {
		byPick[p.PrizeID] = p
	}
	rows := make([]map[string]any, 0, len(prizes))
	shareMismatches := 0
	prizeMatch := true
	for _, p := range prizes {
		w, wok := byPrize[p.ID]
		pick, pok := byPick[p.ID]
		match := wok == pok && (!wok || w.TicketID == pick.TicketID && w.GroupID == pick.GroupID)
		if !match {
			prizeMatch = false
		}
		var recorded, replayed any
		if wok {
			recorded = map[string]any{"ticketId": w.TicketID, "serial": w.Serial, "groupId": w.GroupID}
		}
		if pok {
			replayed = map[string]any{"ticketId": pick.TicketID, "serial": pick.Serial, "groupId": pick.GroupID}
		}
		rows = append(rows, map[string]any{"priority": p.Priority, "prizeId": p.ID, "recorded": recorded, "replayed": replayed, "match": match})
		if wok {
			value, ok := new(big.Int).SetString(p.Value, 10)
			if !ok {
				return "", status.Error(codes.Internal, "invalid prize value")
			}
			members, err := lotteryShareMembers(ctx, tx, in.LotteryID, w.GroupID, at)
			if err != nil {
				return "", err
			}
			expected, err := splitLotteryPrize(value, members)
			if err != nil {
				return "", status.Error(codes.Internal, "share replay failed")
			}
			actualRows, err := tx.QueryContext(ctx, `SELECT user_id,individual_ticket_count,share_percent::text,share_value_rial::text FROM jazebe_lottery_winner_share WHERE winner_id=$1`, w.WinnerID)
			if err != nil {
				return "", status.Error(codes.Internal, "recorded shares unavailable")
			}
			type saved struct {
				Tickets        int
				Percent, Value string
			}
			actual := map[string]saved{}
			for actualRows.Next() {
				var id string
				var s saved
				if err = actualRows.Scan(&id, &s.Tickets, &s.Percent, &s.Value); err != nil {
					actualRows.Close()
					return "", status.Error(codes.Internal, "recorded shares unavailable")
				}
				actual[id] = s
			}
			err = actualRows.Err()
			actualRows.Close()
			if err != nil {
				return "", status.Error(codes.Internal, "recorded shares unavailable")
			}
			if len(actual) != len(expected) {
				shareMismatches++
			}
			for _, s := range expected {
				a, ok := actual[s.UserID]
				if !ok || a.Tickets != s.IndividualTicketCount || a.Percent != s.SharePercent || a.Value != s.ShareValueRial {
					shareMismatches++
				}
			}
		}
	}
	digestMatch := meta.TicketDigest == "" || meta.TicketDigest == out.TicketDigest
	verdict := "match"
	if !digestMatch || !prizeMatch || shareMismatches > 0 {
		verdict = "mismatch"
	}
	result := "success"
	if verdict == "mismatch" {
		result = "failure"
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params,result) VALUES('lottery.replay','operator',$1,'lottery',$2,$3,jsonb_build_object('verdict',$4::text,'ticketDigest',$5::text,'shareMismatches',$6::bigint),$7)`, actor, in.LotteryID, planetID, verdict, out.TicketDigest, shareMismatches, result)
	if err != nil {
		return "", status.Error(codes.Internal, "replay audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "replay commit failed")
	}
	algorithm := meta.Algorithm
	if algorithm == "" {
		algorithm = lotteryDrawAlgorithm
	}
	var recordedDigest any
	if meta.TicketDigest != "" {
		recordedDigest = meta.TicketDigest
	}
	return encode(map[string]any{"verdict": verdict, "seed": seed, "algorithm": algorithm, "ticketCount": out.TicketCount, "ticketDigest": map[string]any{"recorded": recordedDigest, "replayed": out.TicketDigest, "match": digestMatch}, "prizes": rows, "shares": map[string]any{"match": shareMismatches == 0, "mismatches": shareMismatches}})
}
