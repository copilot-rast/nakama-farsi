package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type groupUpdateInput struct {
	PlanetID string          `json:"planetId"`
	GroupID  string          `json:"groupId"`
	Name     json.RawMessage `json:"name"`
	Desc     json.RawMessage `json:"desc"`
	Icon     json.RawMessage `json:"icon"`
	Color    json.RawMessage `json:"color"`
	Mode     json.RawMessage `json:"mode"`
	Cap      json.RawMessage `json:"cap"`
}

func groupOptionalString(raw json.RawMessage, limit int, nullable bool) (string, bool, error) {
	if raw == nil {
		return "", false, nil
	}
	if string(raw) == "null" && nullable {
		return "", true, nil
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return "", false, status.Error(codes.InvalidArgument, "invalid group field")
	}
	value = strings.TrimSpace(value)
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) > limit || (!nullable && value == "") {
		return "", false, status.Error(codes.InvalidArgument, "invalid group field")
	}
	return value, true, nil
}

func groupUpdate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input groupUpdateInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	input.PlanetID, err = requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	input.GroupID, err = requiredUUID(input.GroupID)
	if err != nil {
		return "", err
	}
	name, hasName, err := groupOptionalString(input.Name, 64, false)
	if err != nil {
		return "", err
	}
	desc, hasDesc, err := groupOptionalString(input.Desc, 280, true)
	if err != nil {
		return "", err
	}
	icon, hasIcon, err := groupOptionalString(input.Icon, 32, true)
	if err != nil {
		return "", err
	}
	if hasIcon && icon != "" {
		switch icon {
		case "star", "bolt", "trophy", "ticket", "crown", "user", "magnifier", "drink":
		default:
			return "", status.Error(codes.InvalidArgument, "invalid group icon")
		}
	}
	color, hasColor, err := groupOptionalString(input.Color, 7, true)
	if err != nil {
		return "", err
	}
	if hasColor && color != "" && !groupHexColor.MatchString(color) {
		return "", status.Error(codes.InvalidArgument, "invalid group color")
	}
	mode, hasMode, err := groupOptionalString(input.Mode, 8, false)
	if err != nil {
		return "", err
	}
	if hasMode && mode != "open" && mode != "approval" {
		return "", status.Error(codes.InvalidArgument, "invalid group mode")
	}
	var cap int
	hasCap := input.Cap != nil
	if hasCap && (json.Unmarshal(input.Cap, &cap) != nil || cap < 1 || cap > 10000) {
		return "", status.Error(codes.InvalidArgument, "invalid group cap")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group update unavailable")
	}
	defer tx.Rollback()
	var oldName, oldMode, groupStatus, creator string
	var oldDesc, oldIcon, oldColor sql.NullString
	var oldCap, members int
	err = tx.QueryRowContext(ctx, `SELECT name,description,icon,color,mode,member_cap,member_count,status,creator_user_id
		FROM jazebe_group WHERE id=$1 AND planet_id=$2 FOR UPDATE`, input.GroupID, input.PlanetID).
		Scan(&oldName, &oldDesc, &oldIcon, &oldColor, &oldMode, &oldCap, &members, &groupStatus, &creator)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group unavailable")
	}
	if groupStatus != "active" {
		return "", status.Error(codes.FailedPrecondition, "group inactive")
	}
	if creator != actor {
		return "", status.Error(codes.PermissionDenied, "group creator required")
	}
	if err := requireUsablePlanetMembership(ctx, tx, input.PlanetID, actor); err != nil {
		return "", err
	}
	var enabled bool
	var modes []byte
	var capMin, capMax int
	err = tx.QueryRowContext(ctx, `SELECT s.group_creation_enabled,s.allowed_group_modes,s.group_member_cap_min,s.group_member_cap_max
		FROM jazebe_planet_settings s JOIN jazebe_planet p ON p.id=s.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		WHERE p.id=$1 AND p.status='active' AND b.is_active`, input.PlanetID).Scan(&enabled, &modes, &capMin, &capMax)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.FailedPrecondition, "planet unavailable")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group settings unavailable")
	}
	if hasMode {
		var allowed []string
		if json.Unmarshal(modes, &allowed) != nil {
			return "", status.Error(codes.Internal, "group settings invalid")
		}
		found := false
		for _, allowedMode := range allowed {
			if mode == allowedMode {
				found = true
				break
			}
		}
		if !found {
			return "", status.Error(codes.FailedPrecondition, "group mode not allowed")
		}
	}
	if hasCap && (cap < capMin || cap > capMax || cap < members) {
		return "", status.Error(codes.InvalidArgument, "group cap outside allowed range")
	}
	_ = enabled // Creation policy does not prevent an existing creator from editing the group.
	if !hasName {
		name = oldName
	}
	if !hasDesc {
		desc = oldDesc.String
	}
	if !hasIcon {
		icon = oldIcon.String
	}
	if !hasColor {
		color = oldColor.String
	}
	if !hasMode {
		mode = oldMode
	}
	if !hasCap {
		cap = oldCap
	}
	changed := make([]string, 0, 6)
	if name != oldName {
		changed = append(changed, "name")
	}
	if desc != oldDesc.String {
		changed = append(changed, "description")
	}
	if icon != oldIcon.String {
		changed = append(changed, "icon")
	}
	if color != oldColor.String {
		changed = append(changed, "color")
	}
	if mode != oldMode {
		changed = append(changed, "mode")
	}
	if cap != oldCap {
		changed = append(changed, "memberCap")
	}
	if len(changed) > 0 {
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_group SET name=$2,description=NULLIF($3,''),icon=NULLIF($4,''),
			color=NULLIF($5,''),mode=$6,member_cap=$7,updated_at=now() WHERE id=$1`, input.GroupID, name, desc, icon, color, mode, cap)
		if err != nil {
			var pg *pgconn.PgError
			if errors.As(err, &pg) && pg.Code == pgerrcode.UniqueViolation {
				return "", status.Error(codes.AlreadyExists, "group name already reserved")
			}
			return "", status.Error(codes.Internal, "group update failed")
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
			VALUES('group.settings.update','user',$1,'group',$2,$3,jsonb_build_object('fields',$4::jsonb))`, actor, input.GroupID, input.PlanetID, jsonString(changed)); err != nil {
			return "", status.Error(codes.Internal, "group audit failed")
		}
		if mode != oldMode {
			rows, err := tx.QueryContext(ctx, `SELECT user_id FROM jazebe_group_membership WHERE group_id=$1 AND status='active' AND user_id<>$2`, input.GroupID, actor)
			if err != nil {
				return "", status.Error(codes.Internal, "group members unavailable")
			}
			ids := make([]string, 0)
			for rows.Next() {
				var id string
				if rows.Scan(&id) != nil {
					rows.Close()
					return "", status.Error(codes.Internal, "group members unavailable")
				}
				ids = append(ids, id)
			}
			if err = rows.Err(); err != nil {
				rows.Close()
				return "", status.Error(codes.Internal, "group members unavailable")
			}
			rows.Close()
			eventID := uuid.Must(uuid.NewV4()).String()
			for _, id := range ids {
				if err := persistJazebeNotification(ctx, tx, id, "group-mode:"+eventID+":"+id,
					"حالت گروه تغییر کرد", map[string]any{"category": "group", "actionLink": "/group", "groupId": input.GroupID, "mode": mode}, time.Now().UTC()); err != nil {
					return "", status.Error(codes.Internal, "group notice failed")
				}
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "group update failed")
	}
	return encode(map[string]any{"id": input.GroupID, "name": name, "desc": desc, "icon": icon, "color": color, "mode": mode, "cap": cap, "members": members, "changed": changed})
}

func jsonString(value any) string { data, _ := json.Marshal(value); return string(data) }

func groupSetRole(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		GroupID  string `json:"groupId"`
		MemberID string `json:"memberId"`
		Role     string `json:"role"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	input.PlanetID, err = requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	input.GroupID, err = requiredUUID(input.GroupID)
	if err != nil {
		return "", err
	}
	input.MemberID, err = requiredUUID(input.MemberID)
	if err != nil {
		return "", err
	}
	if input.Role != "admin" && input.Role != "member" && input.Role != "creator" {
		return "", status.Error(codes.InvalidArgument, "unsupported role transition")
	}
	if actor == input.MemberID {
		return "", status.Error(codes.InvalidArgument, "cannot change own role")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group role unavailable")
	}
	defer tx.Rollback()
	var creator, groupStatus, name string
	err = tx.QueryRowContext(ctx, `SELECT g.creator_user_id,g.status,g.name FROM jazebe_group g
		JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		WHERE g.id=$1 AND g.planet_id=$2 AND p.status='active' AND b.is_active FOR UPDATE OF g`, input.GroupID, input.PlanetID).Scan(&creator, &groupStatus, &name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group unavailable")
	}
	if groupStatus != "active" {
		return "", status.Error(codes.FailedPrecondition, "group inactive")
	}
	if creator != actor {
		return "", status.Error(codes.PermissionDenied, "group creator required")
	}
	if err := requireUsablePlanetMembership(ctx, tx, input.PlanetID, actor); err != nil {
		return "", err
	}
	var targetRole string
	err = tx.QueryRowContext(ctx, `SELECT role FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active' FOR UPDATE`, input.GroupID, input.MemberID).Scan(&targetRole)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "member not in group")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group member unavailable")
	}
	if input.Role == "creator" {
		if targetRole != "admin" {
			return "", status.Error(codes.PermissionDenied, "creator offer requires an administrator")
		}
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_group_leadership_offer SET status='cancelled',decided_at=now()
			WHERE group_id=$1 AND status='pending'`, input.GroupID); err != nil {
			return "", status.Error(codes.Internal, "offer replacement failed")
		}
		offerID := uuid.Must(uuid.NewV4()).String()
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_leadership_offer(id,planet_id,group_id,from_user_id,to_user_id,expires_at)
			VALUES($1,$2,$3,$4,$5,now()+interval '72 hours')`, offerID, input.PlanetID, input.GroupID, actor, input.MemberID)
		if err != nil {
			return "", status.Error(codes.Internal, "leadership offer failed")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
			VALUES('group.transfer.offer','user',$1,'group',$2,$3,jsonb_build_object('to',$4::text,'offerId',$5::text))`, actor, input.GroupID, input.PlanetID, input.MemberID, offerID); err != nil {
			return "", status.Error(codes.Internal, "offer audit failed")
		}
		if err := persistJazebeNotification(ctx, tx, input.MemberID, "group-offer:"+offerID,
			"پیشنهاد رهبری گروه", map[string]any{"category": "group", "actionLink": "/group", "groupId": input.GroupID, "offerId": offerID}, time.Now().UTC()); err != nil {
			return "", status.Error(codes.Internal, "offer notice failed")
		}
		if err := tx.Commit(); err != nil {
			return "", status.Error(codes.Internal, "leadership offer failed")
		}
		return encode(map[string]any{"groupId": input.GroupID, "userId": input.MemberID, "role": "admin", "offerId": offerID, "offered": true})
	}
	if (input.Role == "admin" && targetRole != "member") || (input.Role == "member" && targetRole != "admin") {
		return "", status.Error(codes.FailedPrecondition, "invalid role transition")
	}
	if input.Role == "admin" {
		var adminCap, adminCount int
		err = tx.QueryRowContext(ctx, `SELECT s.group_admin_cap,(SELECT count(*) FROM jazebe_group_membership m WHERE m.group_id=$2 AND m.status='active' AND m.role='admin')
			FROM jazebe_planet_settings s WHERE s.planet_id=$1`, input.PlanetID, input.GroupID).Scan(&adminCap, &adminCount)
		if err != nil {
			return "", status.Error(codes.Internal, "group policy unavailable")
		}
		if adminCount >= adminCap {
			return "", status.Error(codes.ResourceExhausted, "group admin cap reached")
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_membership SET role=$3 WHERE group_id=$1 AND user_id=$2 AND status='active'`, input.GroupID, input.MemberID, input.Role); err != nil {
		return "", status.Error(codes.Internal, "role change failed")
	}
	if input.Role == "member" {
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_group_leadership_offer SET status='cancelled',decided_at=now() WHERE group_id=$1 AND to_user_id=$2 AND status='pending'`, input.GroupID, input.MemberID); err != nil {
			return "", status.Error(codes.Internal, "offer cancellation failed")
		}
	}
	messageKey := "member.promote"
	action := "group.member.promote"
	notice := "به مدیر گروه ارتقا یافتی"
	if input.Role == "member" {
		messageKey = "member.demote"
		action = "group.member.demote"
		notice = "نقش تو در گروه به عضو عادی تغییر کرد"
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES($1,'user',$2,'group',$3,$4,jsonb_build_object('userId',$5::text))`, action, actor, input.GroupID, input.PlanetID, input.MemberID); err != nil {
		return "", status.Error(codes.Internal, "group audit failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload)
		VALUES($1,$2,$3,'role',$4,$5,jsonb_build_object('userId',$6::text,'role',$7::text))`, uuid.Must(uuid.NewV4()).String(), input.PlanetID, input.GroupID, actor, messageKey, input.MemberID, input.Role); err != nil {
		return "", status.Error(codes.Internal, "group activity failed")
	}
	if err := persistJazebeNotification(ctx, tx, input.MemberID, "group-role:"+uuid.Must(uuid.NewV4()).String(),
		notice, map[string]any{"category": "group", "actionLink": "/group", "groupId": input.GroupID, "groupName": name, "role": input.Role}, time.Now().UTC()); err != nil {
		return "", status.Error(codes.Internal, "group notice failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "role change failed")
	}
	return encode(map[string]any{"groupId": input.GroupID, "userId": input.MemberID, "role": input.Role})
}
