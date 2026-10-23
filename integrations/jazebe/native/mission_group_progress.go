package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func activeGroupForMission(ctx context.Context, tx *sql.Tx, planetID, user string, required bool) (sql.NullString, error) {
	if !required {
		return sql.NullString{}, nil
	}
	var id string
	err := tx.QueryRowContext(ctx, `SELECT gm.group_id FROM jazebe_group_membership gm JOIN jazebe_group g ON g.id=gm.group_id WHERE gm.planet_id=$1 AND gm.user_id=$2 AND gm.status='active' AND g.status='active'`, planetID, user).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return sql.NullString{}, status.Error(codes.PermissionDenied, "active group membership required")
	}
	if err != nil {
		return sql.NullString{}, status.Error(codes.Internal, "group membership lookup failed")
	}
	return sql.NullString{String: id, Valid: true}, nil
}

func validateMissionTargetRewards(target string, groupTarget, points, tickets int, lotteryID string) error {
	if target != "" && target != "solo" && target != "group" {
		return status.Error(codes.InvalidArgument, "invalid mission target")
	}
	if tickets < 0 || tickets > 1_000_000 {
		return status.Error(codes.InvalidArgument, "invalid ticket reward")
	}
	if target == "group" && (groupTarget < 1 || groupTarget > 1_000_000 || points > 0) {
		return status.Error(codes.InvalidArgument, "invalid group mission configuration")
	}
	if tickets > 0 && lotteryID == "" {
		return status.Error(codes.InvalidArgument, "ticket reward requires lottery")
	}
	return nil
}

// advanceGroupMission is called after one member's completion becomes approved,
// in that completion's transaction. The progress lock is the payout barrier:
// only the transaction crossing the target writes a group completion/reward.
func advanceGroupMission(ctx context.Context, tx *sql.Tx, ref missionRef, user, groupID, memberCompletionID string, now time.Time) error {
	var target, tickets, xp int
	var lotteryID sql.NullString
	var title, kind string
	err := tx.QueryRowContext(ctx, `SELECT target_type,group_target_value,reward_tickets,reward_xp,target_lottery_id,title FROM jazebe_mission WHERE id=$1 AND planet_id=$2`, ref.MissionID, ref.PlanetID).Scan(&kind, &target, &tickets, &xp, &lotteryID, &title)
	if errors.Is(err, sql.ErrNoRows) {
		return status.Error(codes.NotFound, "mission not found")
	}
	if err != nil {
		return status.Error(codes.Internal, "group mission lookup failed")
	}
	if kind != "group" || target < 1 {
		return status.Error(codes.FailedPrecondition, "mission is not a group mission")
	}
	// The member's group was captured when the action was submitted. An
	// administrator may review it after the member has left that group.
	progressID := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_mission_progress(id,planet_id,mission_id,group_id,target_value) VALUES($1,$2,$3,$4,$5) ON CONFLICT(mission_id,group_id) DO NOTHING`, progressID, ref.PlanetID, ref.MissionID, groupID, target)
	if err != nil {
		return status.Error(codes.Internal, "group progress unavailable")
	}
	var current, storedTarget int
	var state string
	err = tx.QueryRowContext(ctx, `SELECT id,current_value,target_value,status FROM jazebe_group_mission_progress WHERE mission_id=$1 AND group_id=$2 FOR UPDATE`, ref.MissionID, groupID).Scan(&progressID, &current, &storedTarget, &state)
	if err != nil {
		return status.Error(codes.Internal, "group progress unavailable")
	}
	current++
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_mission_contribution(progress_id,planet_id,user_id,value) VALUES($1,$2,$3,1) ON CONFLICT(progress_id,user_id) DO UPDATE SET value=jazebe_group_mission_contribution.value+1,updated_at=now()`, progressID, ref.PlanetID, user); err != nil {
		return status.Error(codes.Internal, "group contribution failed")
	}
	reached := state == "in_progress" && current >= storedTarget
	if reached {
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_group_mission_progress SET current_value=$2,status='completed',completed_at=$3 WHERE id=$1`, progressID, current, now)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_group_mission_progress SET current_value=$2 WHERE id=$1`, progressID, current)
	}
	if err != nil {
		return status.Error(codes.Internal, "group progress failed")
	}
	// The contributor's approved completion counts for personal mission medals,
	// even when the group's shared target has not yet been reached. Process
	// after commit so a medal failure cannot reverse a successful mission.
	if err = enqueueActivityXP(ctx, tx, ref.PlanetID, user, memberCompletionID, "mission"); err != nil {
		return err
	}
	if !reached {
		return nil
	}
	completionID := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_mission_completion(id,planet_id,mission_id,user_id,group_id,status,submitted_at,awarded_tickets) VALUES($1,$2,$3,NULL,$4,'approved',$5,$6)`, completionID, ref.PlanetID, ref.MissionID, groupID, now, tickets)
	if err != nil {
		return status.Error(codes.Internal, "group completion failed")
	}
	if tickets > 0 {
		if !lotteryID.Valid {
			return status.Error(codes.FailedPrecondition, "mission lottery required")
		}
		if _, err = issueLotteryTickets(ctx, tx, ticketIssue{LotteryID: lotteryID.String, PlanetID: ref.PlanetID, GroupID: groupID, CreatorType: "group", Source: "group_mission", SourceID: completionID, Count: tickets}, now); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload) VALUES($1,$2,$3,'mission',NULL,'mission.completed',jsonb_build_object('missionId',$4::uuid,'tickets',$5::bigint))`, uuid.Must(uuid.NewV4()).String(), ref.PlanetID, groupID, ref.MissionID, tickets); err != nil {
		return status.Error(codes.Internal, "group activity failed")
	}
	rows, err := tx.QueryContext(ctx, `SELECT user_id FROM jazebe_group_mission_contribution WHERE progress_id=$1 AND value>0`, progressID)
	if err != nil {
		return status.Error(codes.Internal, "group contributors unavailable")
	}
	contributors := make([]string, 0)
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return status.Error(codes.Internal, "group contributors unavailable")
		}
		contributors = append(contributors, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return status.Error(codes.Internal, "group contributors unavailable")
	}
	for _, id := range contributors {
		if err = awardGroupMissionXp(ctx, tx, ref.PlanetID, id, completionID, xp, now); err != nil {
			return err
		}
	}
	members, err := tx.QueryContext(ctx, `SELECT user_id FROM jazebe_group_membership WHERE group_id=$1 AND status='active'`, groupID)
	if err != nil {
		return status.Error(codes.Internal, "group notification failed")
	}
	recipients := make([]string, 0)
	for members.Next() {
		var id string
		if err = members.Scan(&id); err != nil {
			members.Close()
			return status.Error(codes.Internal, "group notification failed")
		}
		recipients = append(recipients, id)
	}
	err = members.Err()
	members.Close()
	if err != nil {
		return status.Error(codes.Internal, "group notification failed")
	}
	for _, id := range recipients {
		if err = persistJazebeNotification(ctx, tx, id, "group-mission:"+completionID, "ماموریت گروهی کامل شد", map[string]any{"category": "mission_game", "actionLink": "/missions", "missionId": ref.MissionID, "groupId": groupID, "tickets": tickets, "title": title}, now); err != nil {
			return status.Error(codes.Internal, "group notification failed")
		}
	}
	return nil
}

func awardGroupMissionXp(ctx context.Context, tx *sql.Tx, planetID, user, completionID string, amount int, now time.Time) error {
	if amount <= 0 {
		return nil
	}
	var applied bool
	err := tx.QueryRowContext(ctx, `INSERT INTO jazebe_xp_record(id,planet_id,user_id,amount,source,source_id) VALUES($1,$2,$3,$4,'mission',$5) ON CONFLICT(planet_id,user_id,source,source_id) DO NOTHING RETURNING true`, uuid.Must(uuid.NewV4()).String(), planetID, user, amount, completionID).Scan(&applied)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return status.Error(codes.Internal, "group XP award failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_planet_stats(planet_id,user_id) VALUES($1,$2) ON CONFLICT(planet_id,user_id) DO NOTHING`, planetID, user); err != nil {
		return status.Error(codes.Internal, "group XP award failed")
	}
	var total int64
	var current int
	if err = tx.QueryRowContext(ctx, `UPDATE jazebe_user_planet_stats SET total_xp=total_xp+$3,last_activity_at=$4 WHERE planet_id=$1 AND user_id=$2 RETURNING total_xp,level`, planetID, user, amount, now).Scan(&total, &current); err != nil {
		return status.Error(codes.Internal, "group XP award failed")
	}
	var thresholds []byte
	if err = tx.QueryRowContext(ctx, `SELECT xp_level_thresholds FROM jazebe_planet_settings WHERE planet_id=$1`, planetID).Scan(&thresholds); err != nil {
		return status.Error(codes.Internal, "group XP policy unavailable")
	}
	level, key := levelAt(total, thresholds, current)
	if level > current {
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_user_planet_stats SET level=$3,level_key=$4 WHERE planet_id=$1 AND user_id=$2`, planetID, user, level, key); err != nil {
			return status.Error(codes.Internal, "group level update failed")
		}
	}
	return nil
}
