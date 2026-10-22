package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type operatorReportInput struct {
	ReportID    string `json:"reportId"`
	PlanetID    string `json:"planetId"`
	Status      string `json:"status"`
	SubjectType string `json:"subjectType"`
	Outcome     string `json:"outcome"`
	Reason      string `json:"reason"`
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
}

func operatorReportActor(ctx context.Context) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	if !isOperator(actor, os.Getenv("JAZEBE_OPERATOR_USER_IDS")) &&
		!isOperator(actor, os.Getenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS")) &&
		!isOperator(actor, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "operator required")
	}
	return actor, nil
}

func validateOperatorReportList(in operatorReportInput) (operatorReportInput, error) {
	if in.PlanetID != "" {
		var err error
		in.PlanetID, err = requiredUUID(in.PlanetID)
		if err != nil {
			return in, err
		}
	}
	if in.Status == "" {
		in.Status = "open"
	}
	switch in.Status {
	case "open", "actioned", "dismissed", "all":
	default:
		return in, status.Error(codes.InvalidArgument, "invalid report status")
	}
	switch in.SubjectType {
	case "", "profile", "group", "message":
	default:
		return in, status.Error(codes.InvalidArgument, "invalid report subject")
	}
	if in.Page < 0 || in.Page > 100000 || in.Limit < 0 || in.Limit > 100 {
		return in, status.Error(codes.InvalidArgument, "invalid report page")
	}
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	return in, nil
}

const operatorReportSelect = `SELECT r.id,r.subject_type,r.subject_id,
	CASE r.subject_type WHEN 'profile' THEN up.display_name WHEN 'group' THEN g.name ELSE m.body END,
	r.reason,r.note,r.status,r.reporter_user_id,COALESCE(rp.display_name,ru.display_name),
	r.planet_id,p.name,b.name,r.created_at,r.handled_at
	FROM jazebe_user_report r
	JOIN jazebe_planet p ON p.id=r.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
	LEFT JOIN jazebe_user_profile up ON r.subject_type='profile' AND up.user_id=r.subject_id
	LEFT JOIN jazebe_group g ON r.subject_type='group' AND g.id=r.subject_id
	LEFT JOIN jazebe_group_message m ON r.subject_type='message' AND m.id=r.subject_id
	LEFT JOIN users ru ON ru.id=r.reporter_user_id
	LEFT JOIN jazebe_user_profile rp ON rp.user_id=r.reporter_user_id`

func scanOperatorReport(rows interface{ Scan(...any) error }) (map[string]any, error) {
	var id, subjectType, subjectID, reason, state, reporterID, planetID, planetName, brandName string
	var label, note, reporterName sql.NullString
	var handled sql.NullTime
	var created any
	err := rows.Scan(&id, &subjectType, &subjectID, &label, &reason, &note, &state,
		&reporterID, &reporterName, &planetID, &planetName, &brandName, &created, &handled)
	if err != nil {
		return nil, err
	}
	var subjectLabel, noteValue, reporterLabel, handledAt any
	if label.Valid {
		subjectLabel = label.String
	}
	if note.Valid {
		noteValue = note.String
	}
	if reporterName.Valid {
		reporterLabel = reporterName.String
	}
	if handled.Valid {
		handledAt = handled.Time
	}
	return map[string]any{"id": id, "subjectType": subjectType, "subjectId": subjectID,
		"subjectLabel": subjectLabel, "reason": reason, "note": noteValue, "status": state,
		"reporter": map[string]any{"id": reporterID, "name": reporterLabel},
		"planet":   map[string]any{"id": planetID, "name": planetName}, "brand": brandName,
		"createdAt": created, "handledAt": handledAt}, nil
}

func operatorReportList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := operatorReportActor(ctx); err != nil {
		return "", err
	}
	var in operatorReportInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	in, err := validateOperatorReportList(in)
	if err != nil {
		return "", err
	}
	filter := ` WHERE ($1::uuid IS NULL OR r.planet_id=$1)
		AND ($2='' OR r.subject_type=$2)
		AND ($3='all' OR ($3='open' AND r.status IN ('open','reviewing')) OR r.status=$3)`
	var total int
	if err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_user_report r`+filter,
		nullableID(in.PlanetID), in.SubjectType, in.Status).Scan(&total); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, operatorReportSelect+filter+` ORDER BY r.created_at DESC,r.id DESC LIMIT $4 OFFSET $5`,
		nullableID(in.PlanetID), in.SubjectType, in.Status, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		item, e := scanOperatorReport(rows)
		if e != nil {
			return "", e
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	return encode(map[string]any{"items": items, "total": total, "page": in.Page, "limit": in.Limit})
}

func operatorReportOne(ctx context.Context, db *sql.DB, id string) (string, error) {
	item, err := scanOperatorReport(db.QueryRowContext(ctx, operatorReportSelect+` WHERE r.id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "report not found")
	}
	if err != nil {
		return "", err
	}
	return encode(item)
}

func operatorReportReview(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return operatorReportTransition(ctx, db, payload, true)
}

func operatorReportResolve(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return operatorReportTransition(ctx, db, payload, false)
}

func operatorReportTransition(ctx context.Context, db *sql.DB, payload string, review bool) (string, error) {
	actor, err := operatorReportActor(ctx)
	if err != nil {
		return "", err
	}
	var in operatorReportInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.ReportID, err = requiredUUID(in.ReportID)
	if err != nil {
		return "", err
	}
	if !review {
		in.Reason = strings.TrimSpace(in.Reason)
		if (in.Outcome != "actioned" && in.Outcome != "dismissed") ||
			!utf8.ValidString(in.Reason) || utf8.RuneCountInString(in.Reason) < 1 || utf8.RuneCountInString(in.Reason) > 512 {
			return "", status.Error(codes.InvalidArgument, "valid resolution outcome and reason required")
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var state, subjectType, subjectID, planetID, brandID string
	err = tx.QueryRowContext(ctx, `SELECT r.status,r.subject_type,r.subject_id,r.planet_id,p.brand_id
		FROM jazebe_user_report r JOIN jazebe_planet p ON p.id=r.planet_id
		WHERE r.id=$1 FOR UPDATE OF r`, in.ReportID).Scan(&state, &subjectType, &subjectID, &planetID, &brandID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "report not found")
	}
	if err != nil {
		return "", err
	}
	if state == "actioned" || state == "dismissed" {
		return "", status.Error(codes.AlreadyExists, "report already resolved")
	}
	action := "user_report.review"
	if review {
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_user_report SET status='reviewing',handled_by_admin_id=$2 WHERE id=$1`, in.ReportID, actor)
	} else {
		if in.Outcome == "actioned" {
			action = "user_report.action"
		} else {
			action = "user_report.dismiss"
		}
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_user_report SET status=$2,handled_by_admin_id=$3,handled_at=now() WHERE id=$1`, in.ReportID, in.Outcome, actor)
	}
	if err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,brand_id,planet_id,params)
		VALUES($1,'operator',$2,'user_report',$3,$4,$5,jsonb_build_object('reason',$6,'subjectType',$7,'subjectId',$8))`,
		action, actor, in.ReportID, brandID, planetID, in.Reason, subjectType, subjectID)
	if err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return operatorReportOne(ctx, db, in.ReportID)
}
