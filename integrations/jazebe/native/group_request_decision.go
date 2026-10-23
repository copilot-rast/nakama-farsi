package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type groupDecisionInput struct {
	PlanetID  string `json:"planetId"`
	GroupID   string `json:"groupId"`
	RequestID string `json:"requestId"`
	Decision  string `json:"decision"`
}

func validateGroupDecision(input groupDecisionInput) (groupDecisionInput, error) {
	var err error
	input.PlanetID, err = requiredUUID(input.PlanetID)
	if err != nil {
		return input, err
	}
	input.GroupID, err = requiredUUID(input.GroupID)
	if err != nil {
		return input, err
	}
	input.RequestID, err = requiredUUID(input.RequestID)
	if err != nil {
		return input, err
	}
	if input.Decision != "approve" && input.Decision != "reject" {
		return input, status.Error(codes.InvalidArgument, "invalid join request decision")
	}
	return input, nil
}

// A full group or a member who joined elsewhere still consumes the pending
// request. The refusal is committed before its conflict response, as in Jazebe.
func groupDecideRequest(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input groupDecisionInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	input, err = validateGroupDecision(input)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group decision unavailable")
	}
	defer tx.Rollback()
	// Join and manager approval take locks in the same order: user, then group.
	// This preserves the source's committed refusal when a user joins elsewhere.
	var applicant string
	err = tx.QueryRowContext(ctx, `SELECT user_id FROM jazebe_group_join_request
		WHERE id=$1 AND planet_id=$2 AND group_id=$3`, input.RequestID, input.PlanetID, input.GroupID).Scan(&applicant)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "join request not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group request unavailable")
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(774103,hashtext($1))`, input.PlanetID+applicant); err != nil {
		return "", status.Error(codes.Internal, "group decision unavailable")
	}
	var groupStatus, name string
	var count, cap int
	err = tx.QueryRowContext(ctx, `SELECT status,name,member_count,member_cap FROM jazebe_group
		WHERE id=$1 AND planet_id=$2 FOR UPDATE`, input.GroupID, input.PlanetID).Scan(&groupStatus, &name, &count, &cap)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group decision unavailable")
	}
	if groupStatus != "active" {
		return "", status.Error(codes.FailedPrecondition, "group unavailable")
	}
	var role string
	err = tx.QueryRowContext(ctx, `SELECT role FROM jazebe_group_membership
		WHERE group_id=$1 AND user_id=$2 AND status='active'`, input.GroupID, actor).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.PermissionDenied, "group manager required")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group manager lookup failed")
	}
	if role != "creator" && role != "admin" {
		return "", status.Error(codes.PermissionDenied, "group manager required")
	}
	var requester, requestStatus string
	err = tx.QueryRowContext(ctx, `SELECT user_id,status FROM jazebe_group_join_request
		WHERE id=$1 AND planet_id=$2 AND group_id=$3 FOR UPDATE`, input.RequestID, input.PlanetID, input.GroupID).
		Scan(&requester, &requestStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "join request not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group request unavailable")
	}
	if requestStatus != "pending" {
		return "", status.Error(codes.AlreadyExists, "join request already decided")
	}
	refusal := ""
	decision := input.Decision
	if decision == "approve" && count >= cap {
		decision = "reject"
		refusal = "group full"
	}
	var membershipID string
	if decision == "approve" {
		var elsewhere bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_group_membership
			WHERE planet_id=$1 AND user_id=$2 AND status='active')`, input.PlanetID, requester).Scan(&elsewhere); err != nil {
			return "", status.Error(codes.Internal, "group membership unavailable")
		}
		if elsewhere {
			decision = "auto_cancelled"
			refusal = "already in group"
		}
	}
	// The stored status is the source GroupJoinRequestStatus, not the decision verb.
	decidedStatus := map[string]string{"approve": "approved", "reject": "rejected", "auto_cancelled": "auto_cancelled"}[decision]
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group_join_request
		SET status=$2,decided_at=now(),decided_by_user_id=$3 WHERE id=$1`, input.RequestID, decidedStatus, actor); err != nil {
		return "", status.Error(codes.Internal, "group decision failed")
	}
	if decision == "approve" {
		var displayName sql.NullString
		if err := tx.QueryRowContext(ctx, `SELECT nullif(trim(display_name),'') FROM users WHERE id=$1`, requester).Scan(&displayName); err != nil {
			return "", status.Error(codes.Internal, "profile unavailable")
		}
		nameForFeed := displayName.String
		if nameForFeed == "" {
			nameForFeed = "کاربر ناشناس"
		}
		membershipID = uuid.Must(uuid.NewV4()).String()
		if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_group_membership(id,group_id,planet_id,user_id,role,source)
			VALUES($1,$2,$3,$4,'member','approved_request')`, membershipID, input.GroupID, input.PlanetID, requester); err != nil {
			return "", status.Error(codes.Internal, "group membership approval failed")
		}
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group SET member_count=member_count+1,updated_at=now() WHERE id=$1`, input.GroupID); err != nil {
			return "", status.Error(codes.Internal, "group count update failed")
		}
		var first bool
		if err := tx.QueryRowContext(ctx, `INSERT INTO jazebe_planet_membership(planet_id,user_id,join_path,last_activity_at)
			VALUES($1,$2,'first_activity',now()) ON CONFLICT(planet_id,user_id) DO NOTHING RETURNING true`,
			input.PlanetID, requester).Scan(&first); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.Internal, "planet membership creation failed")
		}
		if first {
			if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet SET fans_count=fans_count+1 WHERE id=$1`, input.PlanetID); err != nil {
				return "", status.Error(codes.Internal, "fan count update failed")
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET last_activity_at=now() WHERE planet_id=$1 AND user_id=$2`, input.PlanetID, requester); err != nil {
			return "", status.Error(codes.Internal, "planet activity update failed")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload)
			VALUES($1,$2,$3,'join',$4,'member.joined',jsonb_build_object('userId',$4::uuid,'name',$5::text,'source','approved_request','text',$6::text))`,
			uuid.Must(uuid.NewV4()).String(), input.PlanetID, input.GroupID, requester, nameForFeed, nameForFeed+" به گروه پیوست"); err != nil {
			return "", status.Error(codes.Internal, "group activity failed")
		}
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group_join_request SET status='auto_cancelled',decided_at=now()
			WHERE planet_id=$1 AND user_id=$2 AND status='pending'`, input.PlanetID, requester); err != nil {
			return "", status.Error(codes.Internal, "group request cleanup failed")
		}
	}
	if decision != "auto_cancelled" {
		if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
			VALUES('group.request.decision','user',$1,'group_request',$2,$3,jsonb_build_object('decision',$4::text,'groupId',$5::text))`,
			actor, input.RequestID, input.PlanetID, decision, input.GroupID); err != nil {
			return "", status.Error(codes.Internal, "group decision audit failed")
		}
		message := "درخواست عضویت در گروه تأیید شد"
		if decision == "reject" {
			message = "درخواست عضویت در گروه پذیرفته نشد"
		}
		if err := persistJazebeNotification(ctx, tx, requester, "group-request:"+input.RequestID,
			message, map[string]any{"category": "group", "actionLink": "/groups", "groupId": input.GroupID, "requestId": input.RequestID, "groupName": name, "decision": decision}, time.Now().UTC()); err != nil {
			return "", status.Error(codes.Internal, "group decision notification failed")
		}
	}
	if membershipID != "" {
		if err := enqueueActivityXP(ctx, tx, input.PlanetID, requester, membershipID, "group_join"); err != nil {
			return "", status.Error(codes.Internal, "group progression event failed")
		}
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "group decision failed")
	}
	if membershipID != "" {
		groupJoinProgression(ctx, logger, db, input.PlanetID, requester, membershipID)
	}
	if refusal == "group full" {
		return "", status.Error(codes.ResourceExhausted, "group full")
	}
	if refusal == "already in group" {
		return "", status.Error(codes.AlreadyExists, "requester already joined a group")
	}
	return encode(map[string]any{"requestId": input.RequestID, "status": decision})
}

// Jazebe's bulk form reports each request independently. A failed request
// never rolls back decisions committed for earlier IDs in the batch.
func groupDecideMany(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	if _, err := userID(ctx); err != nil {
		return "", err
	}
	var input struct {
		PlanetID   string   `json:"planetId"`
		GroupID    string   `json:"groupId"`
		RequestIDs []string `json:"requestIds"`
		Decision   string   `json:"decision"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	if len(input.RequestIDs) < 1 || len(input.RequestIDs) > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid request count")
	}
	base, err := validateGroupDecision(groupDecisionInput{PlanetID: input.PlanetID, GroupID: input.GroupID,
		RequestID: input.RequestIDs[0], Decision: input.Decision})
	if err != nil {
		return "", err
	}
	seen := make(map[string]bool)
	ids := make([]string, 0, len(input.RequestIDs))
	for _, raw := range input.RequestIDs {
		id, err := requiredUUID(raw)
		if err != nil {
			return "", err
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	approved := make([]string, 0)
	rejected := make([]string, 0)
	failed := make([]map[string]string, 0)
	for _, id := range ids {
		base.RequestID = id
		one, err := encode(base)
		if err != nil {
			return "", status.Error(codes.Internal, "group decision encoding failed")
		}
		_, err = groupDecideRequest(ctx, logger, db, nk, one)
		if err != nil {
			failed = append(failed, map[string]string{"requestId": id, "code": status.Code(err).String()})
			continue
		}
		if input.Decision == "approve" {
			approved = append(approved, id)
		} else {
			rejected = append(rejected, id)
		}
	}
	return encode(map[string]any{"approved": approved, "rejected": rejected, "failed": failed})
}
