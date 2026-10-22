package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type awardableMedal struct {
	id, name, kind  string
	schema, params  []byte
	points, tickets int
}

// evaluateMedalsInTx reads the actual activity ledgers. The caller owns the
// transaction so acquisition and every monetary side effect commit together.
func evaluateMedalsInTx(ctx context.Context, tx *sql.Tx, planetID, userID, activity, eventKey, only string, now time.Time) (int, error) {
	rows, err := tx.QueryContext(ctx, `SELECT pm.id, l.name, l.condition_type, l.condition_param_schema, pm.condition_params,
		pm.reward_points, pm.reward_tickets FROM jazebe_planet_medal pm
		JOIN jazebe_medal_library l ON l.id=pm.medal_library_id
		WHERE pm.planet_id=$1 AND pm.status='active' AND ($2='' OR pm.id::text=$2)
		AND NOT EXISTS (SELECT 1 FROM jazebe_medal_acquisition a WHERE a.planet_medal_id=pm.id AND a.user_id=$3)
		ORDER BY pm.id`, planetID, only, userID)
	if err != nil {
		return 0, err
	}
	medals := []awardableMedal{}
	for rows.Next() {
		var m awardableMedal
		if err = rows.Scan(&m.id, &m.name, &m.kind, &m.schema, &m.params, &m.points, &m.tickets); err != nil {
			rows.Close()
			return 0, err
		}
		medals = append(medals, m)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, err
	}
	granted := 0
	for _, m := range medals {
		condition, err := resolveMedalCondition(m.kind, m.schema, m.params)
		if err != nil {
			return granted, err
		}
		if activity != "" && !medalListensTo(condition, activity) {
			continue
		}
		met, err := medalEligible(ctx, tx, planetID, userID, eventKey, condition, now)
		if err != nil {
			return granted, err
		}
		if !met {
			continue
		}
		acquired, err := grantMedalInTx(ctx, tx, planetID, userID, m, now)
		if err != nil {
			return granted, err
		}
		if acquired {
			granted++
		}
	}
	return granted, nil
}

func medalEligible(ctx context.Context, tx *sql.Tx, planetID, userID, eventKey string, c resolvedMedalCondition, now time.Time) (bool, error) {
	if c.Type == "special_event" {
		return c.EventKey != "" && c.EventKey == eventKey, nil
	}
	if c.Threshold < 1 {
		return false, errors.New("invalid medal threshold")
	}
	if c.Type == "duration" {
		var joined time.Time
		err := tx.QueryRowContext(ctx, `SELECT joined_at FROM jazebe_planet_membership WHERE planet_id=$1 AND user_id=$2`, planetID, userID).Scan(&joined)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return int(now.Sub(joined).Hours()/24) >= c.Threshold, nil
	}
	if c.Type != "activity_count" {
		return false, errors.New("unsupported medal condition")
	}
	queries := map[string]string{
		"purchase":      `SELECT count(*) FROM jazebe_order WHERE planet_id=$1 AND user_id=$2 AND status='delivered'`,
		"mission":       `SELECT count(*) FROM jazebe_mission_completion WHERE planet_id=$1 AND user_id=$2 AND status='approved' AND deleted_at IS NULL`,
		"game":          `SELECT count(*) FROM jazebe_game_session WHERE planet_id=$1 AND user_id=$2 AND status='finished' AND score IS NOT NULL`,
		"referral":      `SELECT count(*) FROM jazebe_mission_invite WHERE planet_id=$1 AND inviter_user_id=$2 AND status='qualified'`,
		"group_members": `SELECT COALESCE(MAX(member_count),0) FROM jazebe_group WHERE planet_id=$1 AND creator_user_id=$2 AND status='active'`,
	}
	query, ok := queries[c.Activity]
	if !ok {
		return false, fmt.Errorf("medal activity %q requires an unintegrated source ledger", c.Activity)
	}
	var count int
	if err := tx.QueryRowContext(ctx, query, planetID, userID).Scan(&count); err != nil {
		return false, err
	}
	return count >= c.Threshold, nil
}

func grantMedalInTx(ctx context.Context, tx *sql.Tx, planetID, userID string, m awardableMedal, now time.Time) (bool, error) {
	id := uuid.Must(uuid.NewV4()).String()
	var inserted string
	err := tx.QueryRowContext(ctx, `INSERT INTO jazebe_medal_acquisition(id,planet_id,planet_medal_id,user_id,acquired_at)
		VALUES($1,$2,$3,$4,$5) ON CONFLICT(planet_medal_id,user_id) DO NOTHING RETURNING id`, id, planetID, m.id, userID, now).Scan(&inserted)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var pointID any
	if m.points > 0 {
		movement, err := applyPointMovement(ctx, tx, planetID, userID, "credit", int64(m.points), "medal", id, "Medal: "+m.name, now)
		if err != nil {
			return false, err
		}
		pointID = movement.TransactionID
	}
	ticketIDs, err := medalRewardTickets(ctx, tx, planetID, userID, id, m.tickets, now)
	if err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_medal_acquisition SET reward_point_transaction_id=$2,
		reward_ticket_ids=CASE WHEN $3::boolean THEN ARRAY(SELECT t.id FROM jazebe_lottery_ticket t WHERE t.source='medal' AND t.source_id=$1) ELSE NULL END
		WHERE id=$1`, id, pointID, len(ticketIDs) > 0); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_planet_stats(planet_id,user_id) VALUES($1,$2) ON CONFLICT(planet_id,user_id) DO NOTHING`, planetID, userID); err != nil {
		return false, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_user_planet_stats SET medals_count=medals_count+1,tickets_created=tickets_created+$3,last_activity_at=$4 WHERE planet_id=$1 AND user_id=$2`, planetID, userID, len(ticketIDs), now); err != nil {
		return false, err
	}
	content := map[string]any{"category": "mission_game", "actionLink": "/profile", "medalId": m.id, "rewardPoints": m.points, "rewardTickets": len(ticketIDs)}
	if err = persistJazebeNotification(ctx, tx, userID, "medal:"+id, "مدال جدید دریافت کردی", content, now); err != nil {
		return false, err
	}
	return true, nil
}

// Ticket caps and closed draws may prevent the optional ticket part of a
// medal. A savepoint keeps PostgreSQL usable after the issuer rejects it.
func medalRewardTickets(ctx context.Context, tx *sql.Tx, planetID, userID, acquisitionID string, count int, now time.Time) ([]string, error) {
	if count == 0 {
		return nil, nil
	}
	var groupID, lotteryID string
	err := tx.QueryRowContext(ctx, `SELECT group_id FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active' LIMIT 1`, planetID, userID).Scan(&groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	err = tx.QueryRowContext(ctx, `SELECT id FROM jazebe_lottery WHERE planet_id=$1 AND status='active' AND draw_at>$2 ORDER BY draw_at,id LIMIT 1`, planetID, now).Scan(&lotteryID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `SAVEPOINT medal_tickets`); err != nil {
		return nil, err
	}
	_, err = issueLotteryTickets(ctx, tx, ticketIssue{LotteryID: lotteryID, PlanetID: planetID, GroupID: groupID, UserID: userID, Source: "medal", SourceID: acquisitionID, CreatorType: "member", Count: count}, now)
	if err != nil {
		if _, rollbackErr := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT medal_tickets`); rollbackErr != nil {
			return nil, rollbackErr
		}
		if _, releaseErr := tx.ExecContext(ctx, `RELEASE SAVEPOINT medal_tickets`); releaseErr != nil {
			return nil, releaseErr
		}
		if code := status.Code(err); code == codes.ResourceExhausted || code == codes.FailedPrecondition || code == codes.NotFound || code == codes.AlreadyExists {
			return nil, nil
		}
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `RELEASE SAVEPOINT medal_tickets`); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND source='medal' AND source_id=$2 ORDER BY id`, lotteryID, acquisitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
