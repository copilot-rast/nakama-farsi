package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func chatModerationRole(ctx context.Context, db *sql.DB, brandID, actor string) (string, error) {
	if isOperator(actor, os.Getenv("JAZEBE_OPERATOR_USER_IDS")) {
		return "operator", nil
	}
	var allowed bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_brand_admin a JOIN jazebe_brand b ON b.id=a.brand_id
		WHERE a.brand_id=$1 AND a.user_id=$2 AND a.is_active AND b.is_active AND a.role IN ('main','marketing'))`, brandID, actor).Scan(&allowed)
	if err != nil {
		return "", status.Error(codes.Internal, "moderation authorization unavailable")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "brand moderation role required")
	}
	return "brand_admin", nil
}

func groupChatReports(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		BrandID  string `json:"brandId"`
		PlanetID string `json:"planetId"`
		Status   string `json:"status"`
		Page     int    `json:"page"`
		Limit    int    `json:"limit"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	if input.BrandID != "" {
		input.BrandID, err = requiredUUID(input.BrandID)
		if err != nil {
			return "", err
		}
	}
	if input.PlanetID != "" {
		input.PlanetID, err = requiredUUID(input.PlanetID)
		if err != nil {
			return "", err
		}
	}
	if input.Status == "" {
		input.Status = "open"
	}
	switch input.Status {
	case "all", "open", "reviewing", "actioned", "dismissed":
	default:
		return "", status.Error(codes.InvalidArgument, "invalid report status")
	}
	if input.Page < 0 || input.Limit < 0 || input.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid report page")
	}
	if input.Page == 0 {
		input.Page = 1
	}
	if input.Limit == 0 {
		input.Limit = 20
	}
	if input.Page > 100000 {
		return "", status.Error(codes.InvalidArgument, "report page too large")
	}
	operator := isOperator(actor, os.Getenv("JAZEBE_OPERATOR_USER_IDS"))
	if !operator {
		if input.BrandID == "" {
			return "", status.Error(codes.PermissionDenied, "brand scope required")
		}
		if _, err := chatModerationRole(ctx, db, input.BrandID, actor); err != nil {
			return "", err
		}
	}
	var total int
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_user_report r JOIN jazebe_planet p ON p.id=r.planet_id
		WHERE r.subject_type='message' AND ($1::uuid IS NULL OR p.brand_id=$1) AND ($2::uuid IS NULL OR r.planet_id=$2)
		AND ($3='all' OR r.status=$3)`, nullableID(input.BrandID), nullableID(input.PlanetID), input.Status).Scan(&total)
	if err != nil {
		return "", status.Error(codes.Internal, "message reports unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT r.id,r.status,r.reason,r.note,r.created_at,r.handled_at,
		r.reporter_user_id,COALESCE(NULLIF(trim(ru.display_name),''),'کاربر ناشناس'),
		p.id,p.name,p.brand_id,b.name,g.id,g.name,g.status,m.id,m.seq,m.body,m.deleted_at,m.deleted_as_admin,m.created_at,
		m.sender_user_id,COALESCE(NULLIF(trim(su.display_name),''),'کاربر ناشناس')
		FROM jazebe_user_report r JOIN jazebe_planet p ON p.id=r.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		LEFT JOIN jazebe_group_message m ON m.id=r.subject_id LEFT JOIN jazebe_group g ON g.id=m.group_id
		LEFT JOIN users ru ON ru.id=r.reporter_user_id LEFT JOIN users su ON su.id=m.sender_user_id
		WHERE r.subject_type='message' AND ($1::uuid IS NULL OR p.brand_id=$1) AND ($2::uuid IS NULL OR r.planet_id=$2)
		AND ($3='all' OR r.status=$3) ORDER BY r.created_at DESC,r.id DESC LIMIT $4 OFFSET $5`,
		nullableID(input.BrandID), nullableID(input.PlanetID), input.Status, input.Limit, (input.Page-1)*input.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "message reports unavailable")
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, state, reason, reporterID, reporterName, planetID, planetName, brandID, brandName string
		var note sql.NullString
		var created time.Time
		var handled sql.NullTime
		var groupID, groupName, groupStatus, messageID, body, senderID sql.NullString
		var seq sql.NullInt64
		var deleted, messageCreated sql.NullTime
		var deletedAsAdmin sql.NullBool
		var senderName string
		err = rows.Scan(&id, &state, &reason, &note, &created, &handled, &reporterID, &reporterName, &planetID, &planetName, &brandID, &brandName,
			&groupID, &groupName, &groupStatus, &messageID, &seq, &body, &deleted, &deletedAsAdmin, &messageCreated, &senderID, &senderName)
		if err != nil {
			return "", status.Error(codes.Internal, "message reports unavailable")
		}
		var noteValue, handledValue, groupValue, messageValue any
		if note.Valid {
			noteValue = note.String
		}
		if handled.Valid {
			handledValue = handled.Time
		}
		if groupID.Valid {
			groupValue = map[string]any{"id": groupID.String, "name": groupName.String, "status": groupStatus.String}
		}
		if messageID.Valid {
			var text any
			if !deleted.Valid && body.Valid {
				text = body.String
			}
			var sender any
			if senderID.Valid {
				sender = map[string]any{"id": senderID.String, "name": senderName}
			}
			messageValue = map[string]any{"id": messageID.String, "seq": seq.Int64, "text": text, "deleted": deleted.Valid, "deletedAsAdmin": deletedAsAdmin.Bool, "createdAt": messageCreated.Time, "sender": sender}
		}
		items = append(items, map[string]any{"id": id, "status": state, "reason": reason, "note": noteValue, "createdAt": created, "handledAt": handledValue,
			"reporter": map[string]any{"id": reporterID, "name": reporterName}, "planet": map[string]any{"id": planetID, "name": planetName},
			"brand": map[string]any{"productId": brandID, "tradeName": brandName}, "group": groupValue, "message": messageValue})
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "message reports unavailable")
	}
	return encode(map[string]any{"items": items, "total": total, "page": input.Page, "limit": input.Limit})
}

func nullableID(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func groupChatDismissReport(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		ReportID string `json:"reportId"`
		Reason   string `json:"reason"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	input.ReportID, err = requiredUUID(input.ReportID)
	if err != nil {
		return "", err
	}
	input.Reason = strings.TrimSpace(input.Reason)
	if !utf8.ValidString(input.Reason) || utf8.RuneCountInString(input.Reason) < 1 || utf8.RuneCountInString(input.Reason) > 512 {
		return "", status.Error(codes.InvalidArgument, "dismissal reason required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "report dismissal unavailable")
	}
	defer tx.Rollback()
	var planetID, brandID, state string
	err = tx.QueryRowContext(ctx, `SELECT r.planet_id,p.brand_id,r.status FROM jazebe_user_report r
		JOIN jazebe_planet p ON p.id=r.planet_id WHERE r.id=$1 AND r.subject_type='message' FOR UPDATE OF r`, input.ReportID).Scan(&planetID, &brandID, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "report not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "report unavailable")
	}
	actorType := "operator"
	if !isOperator(actor, os.Getenv("JAZEBE_OPERATOR_USER_IDS")) {
		var allowed bool
		err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_brand_admin a JOIN jazebe_brand b ON b.id=a.brand_id
			WHERE a.brand_id=$1 AND a.user_id=$2 AND a.is_active AND b.is_active AND a.role IN ('main','marketing'))`, brandID, actor).Scan(&allowed)
		if err != nil {
			return "", status.Error(codes.Internal, "moderation authorization unavailable")
		}
		if !allowed {
			return "", status.Error(codes.PermissionDenied, "brand moderation role required")
		}
		actorType = "brand_admin"
	}
	if state == "actioned" || state == "dismissed" {
		return "", status.Error(codes.AlreadyExists, "report already resolved")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_user_report SET status='dismissed',handled_by_admin_id=$2,handled_at=now() WHERE id=$1`, input.ReportID, actor); err != nil {
		return "", status.Error(codes.Internal, "report dismissal failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('user_report.dismiss',$1,$2,'user_report',$3,$4,jsonb_build_object('reason',$5::text))`, actorType, actor, input.ReportID, planetID, input.Reason); err != nil {
		return "", status.Error(codes.Internal, "report audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "report dismissal failed")
	}
	return encode(map[string]any{"id": input.ReportID, "status": "dismissed"})
}
