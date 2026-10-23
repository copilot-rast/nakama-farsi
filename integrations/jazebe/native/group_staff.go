package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Staff views of groups and operator group moderation (source
// group-admin.controller.ts, group-brand.controller.ts,
// GroupModerationService and GroupService suspend/unsuspend/disband, PRD
// US-37/38). An operator reaches every group; a brand administrator with the
// moderation role (main or marketing) reaches the groups of their brand,
// addressed through the planet. Staff reads include suspended and disbanded
// groups.

type staffGroupInput struct {
	PlanetID string `json:"planetId"`
	GroupID  string `json:"groupId"`
}

// staffGroupAccess authorizes the caller on one group and returns its
// reference and the audit actor type.
func staffGroupAccess(ctx context.Context, db *sql.DB, user string, in staffGroupInput) (groupRef, string, error) {
	groupID, err := requiredUUID(in.GroupID)
	if err != nil {
		return groupRef{}, "", err
	}
	var planetID, brandID string
	err = db.QueryRowContext(ctx, `SELECT g.planet_id,p.brand_id FROM jazebe_group g JOIN jazebe_planet p ON p.id=g.planet_id WHERE g.id=$1`, groupID).Scan(&planetID, &brandID)
	if errors.Is(err, sql.ErrNoRows) {
		return groupRef{}, "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return groupRef{}, "", status.Error(codes.Internal, "group unavailable")
	}
	ref := groupRef{PlanetID: planetID, GroupID: groupID}
	if isOperator(user, os.Getenv("JAZEBE_OPERATOR_USER_IDS")) {
		return ref, "operator", nil
	}
	// The brand panel addresses a group through its planet; another planet's
	// group reads as missing.
	if in.PlanetID == "" {
		return groupRef{}, "", status.Error(codes.InvalidArgument, "planetId required")
	}
	if p, err := requiredUUID(in.PlanetID); err != nil {
		return groupRef{}, "", err
	} else if p != planetID {
		return groupRef{}, "", status.Error(codes.NotFound, "group not found")
	}
	actorType, err := chatModerationRole(ctx, db, brandID, user)
	return ref, actorType, err
}

func staffGroupCall(read func(ctx context.Context, db *sql.DB, ref groupRef) (string, error)) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
		user, err := userID(ctx)
		if err != nil {
			return "", err
		}
		var in staffGroupInput
		if err := decode(payload, &in); err != nil {
			return "", err
		}
		ref, _, err := staffGroupAccess(ctx, db, user, in)
		if err != nil {
			return "", err
		}
		return read(ctx, db, ref)
	}
}

var (
	staffGroupMembers  = staffGroupCall(groupMemberRows)
	staffGroupRequests = staffGroupCall(groupRequestRows)
	staffGroupActivity = staffGroupCall(groupActivityRows)
	staffGroupHistory  = staffGroupCall(groupHistoryRows)
	staffGroupGet      = staffGroupCall(func(ctx context.Context, db *sql.DB, ref groupRef) (string, error) {
		rows, _, err := staffGroupRows(ctx, db, `g.id=$1`, []any{ref.GroupID}, "", 1, 0)
		if err != nil {
			return "", err
		}
		if len(rows) == 0 {
			return "", status.Error(codes.NotFound, "group not found")
		}
		return encode(rows[0])
	})
)

const staffGroupSelect = `SELECT g.id,g.planet_id,p.brand_id,g.name,g.description,g.icon,g.color,g.mode,g.status,g.member_count,g.member_cap,
	g.creator_user_id,COALESCE(NULLIF(trim(u.display_name),''),'کاربر ناشناس'),
	(SELECT count(*) FROM jazebe_group_join_request r WHERE r.group_id=g.id AND r.status='pending'),
	g.created_at,g.suspended_at,g.suspended_reason,g.disbanded_at,p.name,p.slug,b.name
	FROM jazebe_group g JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
	LEFT JOIN users u ON u.id=g.creator_user_id WHERE `

// staffGroupRows returns the staff rows matching where, and the total count.
func staffGroupRows(ctx context.Context, db *sql.DB, where string, args []any, order string, limit, offset int) ([]map[string]any, int, error) {
	var total int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_group g JOIN jazebe_planet p ON p.id=g.planet_id WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, status.Error(codes.Internal, "groups unavailable")
	}
	if order == "" {
		order = `g.created_at DESC,g.id DESC`
	}
	n := len(args)
	rows, err := db.QueryContext(ctx, staffGroupSelect+where+` ORDER BY `+order+` LIMIT $`+stringInt(n+1)+` OFFSET $`+stringInt(n+2), append(args, limit, offset)...)
	if err != nil {
		return nil, 0, status.Error(codes.Internal, "groups unavailable")
	}
	defer rows.Close()
	out := []map[string]any{}
	ranked := map[string]map[string]int{}
	for rows.Next() {
		var id, planetID, brandID, name, mode, state, creator, creatorName, planetName, planetSlug, brandName string
		var description, icon, color, reason sql.NullString
		var members, cap, pending int
		var created time.Time
		var suspended, disbanded sql.NullTime
		if err := rows.Scan(&id, &planetID, &brandID, &name, &description, &icon, &color, &mode, &state, &members, &cap,
			&creator, &creatorName, &pending, &created, &suspended, &reason, &disbanded, &planetName, &planetSlug, &brandName); err != nil {
			return nil, 0, status.Error(codes.Internal, "groups unavailable")
		}
		var rank any
		if state == "active" {
			if ranked[planetID] == nil {
				lottery, err := groupRunningLottery(ctx, db, planetID, time.Now().UTC())
				if err != nil {
					return nil, 0, err
				}
				if ranked[planetID], err = groupRanks(ctx, db, planetID, lottery); err != nil {
					return nil, 0, err
				}
			}
			if r, ok := ranked[planetID][id]; ok {
				rank = r
			}
		}
		out = append(out, map[string]any{"id": id, "planetId": planetID, "brandId": brandID, "name": name,
			"description": nullable(description), "icon": nullable(icon), "color": nullable(color), "mode": mode, "status": state,
			"memberCount": members, "memberCap": cap, "rank": rank, "creator": map[string]string{"id": creator, "name": creatorName},
			"pendingRequests": pending, "createdAt": created, "suspendedAt": nullTime(suspended), "suspendedReason": nullable(reason),
			"disbandedAt": nullTime(disbanded), "planet": map[string]string{"id": planetID, "name": planetName, "slug": planetSlug},
			"brand": map[string]string{"id": brandID, "name": brandName}})
	}
	if rows.Err() != nil {
		return nil, 0, status.Error(codes.Internal, "groups unavailable")
	}
	return out, total, nil
}

func nullTime(t sql.NullTime) any {
	if t.Valid {
		return t.Time
	}
	return nil
}

// staffGroupList lists groups: an operator across the platform (optionally by
// brand or planet), a brand administrator across their brand or one planet.
func staffGroupList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		BrandID  string `json:"brandId"`
		PlanetID string `json:"planetId"`
		Status   string `json:"status"`
		Search   string `json:"search"`
		Sort     string `json:"sort"`
		Page     int    `json:"page"`
		Limit    int    `json:"limit"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	for _, id := range []*string{&in.BrandID, &in.PlanetID} {
		if *id != "" {
			if *id, err = requiredUUID(*id); err != nil {
				return "", err
			}
		}
	}
	switch in.Status {
	case "", "active", "suspended", "disbanded":
	default:
		return "", status.Error(codes.InvalidArgument, "invalid status")
	}
	order := ""
	switch in.Sort {
	case "", "recent":
	case "members":
		order = `g.member_count DESC,g.created_at DESC,g.id DESC`
	default:
		return "", status.Error(codes.InvalidArgument, "invalid sort")
	}
	in.Search = strings.TrimSpace(in.Search)
	if utf8.RuneCountInString(in.Search) > 64 || in.Page < 0 || in.Page > 100000 || in.Limit < 0 || in.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid filter")
	}
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	if !isOperator(user, os.Getenv("JAZEBE_OPERATOR_USER_IDS")) {
		if in.PlanetID != "" {
			var brandID string
			err := db.QueryRowContext(ctx, `SELECT brand_id FROM jazebe_planet WHERE id=$1`, in.PlanetID).Scan(&brandID)
			if errors.Is(err, sql.ErrNoRows) || (err == nil && in.BrandID != "" && brandID != in.BrandID) {
				return "", status.Error(codes.NotFound, "planet not found")
			}
			if err != nil {
				return "", status.Error(codes.Internal, "planet unavailable")
			}
			in.BrandID = brandID
		}
		if in.BrandID == "" {
			return "", status.Error(codes.InvalidArgument, "brandId or planetId required")
		}
		if _, err := chatModerationRole(ctx, db, in.BrandID, user); err != nil {
			return "", err
		}
	}
	search := ""
	if in.Search != "" {
		search = "%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(in.Search) + "%"
	}
	where := `($1::uuid IS NULL OR p.brand_id=$1::uuid) AND ($2::uuid IS NULL OR g.planet_id=$2::uuid)
		AND ($3::text='' OR g.status=$3::text) AND ($4::text='' OR g.name ILIKE $4::text)`
	items, total, err := staffGroupRows(ctx, db, where, []any{nullUUID(in.BrandID), nullUUID(in.PlanetID), in.Status, search}, order, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", err
	}
	return encode(map[string]any{"items": items, "total": total, "page": in.Page, "limit": in.Limit})
}

// staffGroupMessages pages a group's chat for staff; every look is audited
// as group.chat.view (PRD US-37 rule 3).
func staffGroupMessages(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		staffGroupInput
		Before int64 `json:"before"`
		Limit  int   `json:"limit"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.Before < 0 || in.Limit < 0 || in.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid message page")
	}
	if in.Limit == 0 {
		in.Limit = 50
	}
	ref, actorType, err := staffGroupAccess(ctx, db, user, in.staffGroupInput)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "chat unavailable")
	}
	defer tx.Rollback()
	items, err := chatRows(ctx, tx, ref.GroupID, user, in.Before, in.Limit+1)
	if err != nil {
		return "", err
	}
	hasMore := len(items) > in.Limit
	if hasMore {
		items = items[1:]
	}
	params := "null"
	if in.Before > 0 {
		params = `{"before":` + stringInt(int(in.Before)) + `}`
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('group.chat.view',$1,$2,'group',$3,$4,$5::jsonb)`, actorType, user, ref.GroupID, ref.PlanetID, params); err != nil {
		return "", status.Error(codes.Internal, "chat view audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "chat unavailable")
	}
	return encode(map[string]any{"items": items, "hasMore": hasMore})
}

type staffActionInput struct {
	PlanetID string `json:"planetId"`
	GroupID  string `json:"groupId"`
	UserID   string `json:"userId"`
	Reason   string `json:"reason"`
}

func (in staffActionInput) scope() staffGroupInput {
	return staffGroupInput{PlanetID: in.PlanetID, GroupID: in.GroupID}
}

func readStaffActionInput(payload string, reasonRequired bool) (staffActionInput, error) {
	var in staffActionInput
	if err := decode(payload, &in); err != nil {
		return in, err
	}
	var err error
	if in.GroupID, err = requiredUUID(in.GroupID); err != nil {
		return in, err
	}
	in.Reason = strings.TrimSpace(in.Reason)
	n := utf8.RuneCountInString(in.Reason)
	if !utf8.ValidString(in.Reason) || n > 512 || (reasonRequired && n < 3) {
		return in, status.Error(codes.InvalidArgument, "reason must be 3 to 512 characters")
	}
	return in, nil
}

// lockStaffGroup locks a group for an operator action.
func lockStaffGroup(ctx context.Context, tx *sql.Tx, groupID string) (groupRef, string, string, error) {
	var ref groupRef
	var name, state string
	err := tx.QueryRowContext(ctx, `SELECT id,planet_id,name,status FROM jazebe_group WHERE id=$1 FOR UPDATE`, groupID).Scan(&ref.GroupID, &ref.PlanetID, &name, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return ref, "", "", status.Error(codes.NotFound, "group not found")
	}
	if err != nil {
		return ref, "", "", status.Error(codes.Internal, "group unavailable")
	}
	return ref, name, state, nil
}

// groupStaffNotice writes the chat system line and tells every active member.
func groupStaffNotice(ctx context.Context, tx *sql.Tx, ref groupRef, event, line, title, body string) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_group_message(id,planet_id,group_id,kind,body) VALUES($1,$2,$3,'system',$4)`,
		uuid.Must(uuid.NewV4()).String(), ref.PlanetID, ref.GroupID, line); err != nil {
		return status.Error(codes.Internal, "group system message failed")
	}
	rows, err := tx.QueryContext(ctx, `SELECT user_id FROM jazebe_group_membership WHERE group_id=$1 AND status='active'`, ref.GroupID)
	if err != nil {
		return status.Error(codes.Internal, "group members unavailable")
	}
	var members []string
	for rows.Next() {
		var user string
		if err := rows.Scan(&user); err != nil {
			rows.Close()
			return status.Error(codes.Internal, "group members unavailable")
		}
		members = append(members, user)
	}
	rows.Close()
	if rows.Err() != nil {
		return status.Error(codes.Internal, "group members unavailable")
	}
	eventID := uuid.Must(uuid.NewV4()).String()
	for _, user := range members {
		if err := persistJazebeNotification(ctx, tx, user, event+":"+eventID+":"+user, title,
			map[string]any{"category": "group", "body": body, "actionLink": "/group", "groupId": ref.GroupID}, time.Now().UTC()); err != nil {
			return status.Error(codes.Internal, "group notice failed")
		}
	}
	return nil
}

func staffGroupAudit(ctx context.Context, tx *sql.Tx, action, actorType, actor string, ref groupRef, reason string) error {
	params, _ := json.Marshal(map[string]any{"reason": nullableReason(reason)})
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES($1,$2,$3,'group',$4,$5,$6::jsonb)`, action, actorType, actor, ref.GroupID, ref.PlanetID, string(params)); err != nil {
		return status.Error(codes.Internal, "group audit failed")
	}
	return nil
}

func nullableReason(reason string) any {
	if reason == "" {
		return nil
	}
	return reason
}

// staffGroupSetSuspended suspends an active group (members keep their data
// but cannot chat or act) or lifts a suspension; operators and brand
// moderators (through the planet) may.
func staffGroupSetSuspended(suspend bool) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		user, err := userID(ctx)
		if err != nil {
			return "", err
		}
		in, err := readStaffActionInput(payload, suspend)
		if err != nil {
			return "", err
		}
		_, actorType, err := staffGroupAccess(ctx, db, user, in.scope())
		if err != nil {
			return "", err
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return "", status.Error(codes.Internal, "group moderation unavailable")
		}
		defer tx.Rollback()
		ref, name, state, err := lockStaffGroup(ctx, tx, in.GroupID)
		if err != nil {
			return "", err
		}
		from, action := "active", "group.suspend"
		if !suspend {
			from, action = "suspended", "group.unsuspend"
		}
		if state != from {
			return "", status.Error(codes.FailedPrecondition, "GROUP_INVALID_TRANSITION: group is "+state)
		}
		if suspend {
			_, err = tx.ExecContext(ctx, `UPDATE jazebe_group SET status='suspended',suspended_at=now(),suspended_by_user_id=$2,suspended_reason=$3,updated_at=now() WHERE id=$1`, ref.GroupID, user, in.Reason)
		} else {
			_, err = tx.ExecContext(ctx, `UPDATE jazebe_group SET status='active',suspended_at=NULL,suspended_by_user_id=NULL,suspended_reason=NULL,updated_at=now() WHERE id=$1`, ref.GroupID)
		}
		if err != nil {
			return "", status.Error(codes.Internal, "group moderation failed")
		}
		if err := staffGroupAudit(ctx, tx, action, actorType, user, ref, in.Reason); err != nil {
			return "", err
		}
		if suspend {
			err = groupStaffNotice(ctx, tx, ref, "group-suspend", "گروه به‌طور موقت تعلیق شد؛ ارسال پیام ممکن نیست.", "تعلیق گروه", "گروه «"+name+"» به‌طور موقت تعلیق شد.")
		} else {
			err = groupStaffNotice(ctx, tx, ref, "group-unsuspend", "تعلیق گروه برداشته شد.", "رفع تعلیق گروه", "تعلیق گروه «"+name+"» برداشته شد؛ گفت‌وگو و فعالیت گروه دوباره باز است.")
		}
		if err != nil {
			return "", err
		}
		if err := tx.Commit(); err != nil {
			return "", status.Error(codes.Internal, "group moderation failed")
		}
		if suspend {
			if err := nk.StreamClose(groupChatStreamMode, ref.GroupID, ref.PlanetID, groupChatStreamLabel); err != nil {
				logger.Warn("Jazebe group stream closure failed: %v", err)
			}
		}
		return staffGroupGet(ctx, logger, db, nk, jsonString(in.scope()))
	}
}

var (
	staffGroupSuspend   = staffGroupSetSuspended(true)
	staffGroupUnsuspend = staffGroupSetSuspended(false)
)

// operatorGroupDisband disbands any group that is not already disbanded; it
// needs the managerial operator level.
func operatorGroupDisband(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	user, err := managerialOperator(ctx)
	if err != nil {
		return "", err
	}
	in, err := readStaffActionInput(payload, true)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group disband unavailable")
	}
	defer tx.Rollback()
	ref, name, state, err := lockStaffGroup(ctx, tx, in.GroupID)
	if err != nil {
		return "", err
	}
	if state == "disbanded" {
		return "", status.Error(codes.FailedPrecondition, "GROUP_INVALID_TRANSITION: group is disbanded")
	}
	if _, err := disbandGroupTx(ctx, tx, ref, name, user, "operator", in.Reason); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "group disband failed")
	}
	if err := nk.StreamClose(groupChatStreamMode, ref.GroupID, ref.PlanetID, groupChatStreamLabel); err != nil {
		logger.Warn("Jazebe group stream closure failed: %v", err)
	}
	return staffGroupGet(ctx, logger, db, nk, jsonString(staffGroupInput{GroupID: ref.GroupID}))
}

// staffGroupRemoveMember removes any member but the creator from a group
// that is not disbanded (source GroupService.removeMemberByStaff).
func staffGroupRemoveMember(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	in, err := readStaffActionInput(payload, true)
	if err != nil {
		return "", err
	}
	if in.UserID, err = requiredUUID(in.UserID); err != nil {
		return "", err
	}
	_, actorType, err := staffGroupAccess(ctx, db, user, in.scope())
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group removal unavailable")
	}
	defer tx.Rollback()
	ref, name, state, err := lockStaffGroup(ctx, tx, in.GroupID)
	if err != nil {
		return "", err
	}
	if state == "disbanded" {
		return "", status.Error(codes.FailedPrecondition, "GROUP_INVALID_TRANSITION: group is disbanded")
	}
	var role string
	err = tx.QueryRowContext(ctx, `SELECT role FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active' FOR UPDATE`, ref.GroupID, in.UserID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "member not in group")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group member unavailable")
	}
	if role == "creator" {
		return "", status.Error(codes.PermissionDenied, "GROUP_ROLE_FORBIDDEN: the creator cannot be removed")
	}
	if err := removeGroupMemberTx(ctx, tx, ref, name, user, actorType, in.UserID, in.Reason); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "member removal failed")
	}
	if err := groupChatKickUser(nk, ref, in.UserID); err != nil {
		logger.Warn("Jazebe group stream eviction failed: %v", err)
	}
	return encode(map[string]any{"groupId": ref.GroupID, "userId": in.UserID, "removed": true})
}
