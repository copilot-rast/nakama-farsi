package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type profileReportInput struct {
	PlanetID string  `json:"planetId"`
	UserID   string  `json:"userId"`
	Reason   string  `json:"reason"`
	Note     *string `json:"note"`
}

func validateProfileReport(in profileReportInput, reporter string) (profileReportInput, error) {
	var err error
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return in, err
	}
	in.UserID, err = requiredUUID(in.UserID)
	if err != nil {
		return in, err
	}
	if in.UserID == reporter {
		return in, status.Error(codes.InvalidArgument, "cannot report own profile")
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if !utf8.ValidString(in.Reason) || utf8.RuneCountInString(in.Reason) < 1 || utf8.RuneCountInString(in.Reason) > 64 {
		return in, status.Error(codes.InvalidArgument, "invalid report reason")
	}
	if in.Note != nil {
		v := strings.TrimSpace(*in.Note)
		if !utf8.ValidString(v) || utf8.RuneCountInString(v) > 512 {
			return in, status.Error(codes.InvalidArgument, "invalid report note")
		}
		in.Note = &v
	}
	return in, nil
}

func profileReport(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	reporter, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in profileReportInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in, err = validateProfileReport(in, reporter)
	if err != nil {
		return "", err
	}
	var visible bool
	err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet p JOIN jazebe_brand b ON b.id=p.brand_id
		LEFT JOIN jazebe_planet_membership vm ON vm.planet_id=p.id AND vm.user_id=$2
		JOIN users vu ON vu.id=$2 AND vu.disable_time='1970-01-01 00:00:00 UTC'
		LEFT JOIN jazebe_user_account va ON va.user_id=vu.id
		JOIN users target ON target.id=$3 AND target.disable_time='1970-01-01 00:00:00 UTC'
		LEFT JOIN jazebe_user_account a ON a.user_id=target.id
		WHERE p.id=$1 AND p.status='active' AND b.is_active
		AND (vm.user_id IS NULL OR vm.is_active) AND COALESCE(va.state,'active')='active'
		AND COALESCE(a.state,'active')='active')`, in.PlanetID, reporter, in.UserID).Scan(&visible)
	if err != nil {
		return "", err
	}
	if !visible {
		return "", status.Error(codes.NotFound, "public profile not found")
	}
	id := uuid.Must(uuid.NewV4()).String()
	err = db.QueryRowContext(ctx, `INSERT INTO jazebe_user_report(id,planet_id,reporter_user_id,subject_type,subject_id,reason,note)
		VALUES($1,$2,$3,'profile',$4,$5,$6)
		ON CONFLICT(planet_id,reporter_user_id,subject_type,subject_id)
			WHERE status='open' AND subject_type='profile' DO NOTHING
		RETURNING id`, id, in.PlanetID, reporter, in.UserID, in.Reason, nullableGameText(in.Note)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		err = db.QueryRowContext(ctx, `SELECT id FROM jazebe_user_report WHERE planet_id=$1 AND reporter_user_id=$2
			AND subject_type='profile' AND subject_id=$3 AND status='open'`, in.PlanetID, reporter, in.UserID).Scan(&id)
	}
	if err != nil {
		return "", err
	}
	return encode(map[string]any{"id": id})
}
