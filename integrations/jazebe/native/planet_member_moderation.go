package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Suspending a member from one of the brand's planets (source
// brand-moderation.service.ts, phase 14). A suspended member keeps their
// history but every planet guard refuses them until the brand lifts it; the
// reason is required, every change is audited and the member is told
// (critical notice).

func brandMemberSetSuspended(suspend bool) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		actor, err := userID(ctx)
		if err != nil {
			return "", err
		}
		var in struct {
			PlanetID string `json:"planetId"`
			UserID   string `json:"userId"`
			Reason   string `json:"reason"`
		}
		if err := decode(payload, &in); err != nil {
			return "", err
		}
		if in.PlanetID, err = requiredUUID(in.PlanetID); err != nil {
			return "", err
		}
		if in.UserID, err = requiredUUID(in.UserID); err != nil {
			return "", err
		}
		in.Reason = strings.TrimSpace(in.Reason)
		if n := utf8.RuneCountInString(in.Reason); !utf8.ValidString(in.Reason) || n < 3 || n > 512 {
			return "", status.Error(codes.InvalidArgument, "reason must be 3 to 512 characters")
		}
		var brandID, planetName string
		err = db.QueryRowContext(ctx, `SELECT brand_id,name FROM jazebe_planet WHERE id=$1`, in.PlanetID).Scan(&brandID, &planetName)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.NotFound, "planet not found")
		}
		if err != nil {
			return "", status.Error(codes.Internal, "planet unavailable")
		}
		actorType, err := chatModerationRole(ctx, db, brandID, actor)
		if err != nil {
			return "", err
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return "", status.Error(codes.Internal, "member moderation unavailable")
		}
		defer tx.Rollback()
		var res sql.Result
		if suspend {
			res, err = tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET is_active=false,suspended_at=now(),suspended_by_user_id=$3,suspended_reason=$4
				WHERE planet_id=$1 AND user_id=$2 AND is_active`, in.PlanetID, in.UserID, actor, in.Reason)
		} else {
			res, err = tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET is_active=true,suspended_at=NULL,suspended_by_user_id=NULL,suspended_reason=NULL
				WHERE planet_id=$1 AND user_id=$2 AND NOT is_active`, in.PlanetID, in.UserID)
		}
		if err != nil {
			return "", status.Error(codes.Internal, "member moderation failed")
		}
		if n, _ := res.RowsAffected(); n == 0 {
			if suspend {
				return "", status.Error(codes.FailedPrecondition, "PLANET_MEMBER_SUSPENDED: not an active member of this planet")
			}
			return "", status.Error(codes.FailedPrecondition, "PLANET_MEMBER_NOT_SUSPENDED: member is not suspended")
		}
		// fans_count counts active memberships, as the migration backfill and reconcile do.
		delta := 1
		if suspend {
			delta = -1
		}
		if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet SET fans_count=GREATEST(fans_count+$2,0) WHERE id=$1`, in.PlanetID, delta); err != nil {
			return "", status.Error(codes.Internal, "member moderation failed")
		}
		title, body, action := "دسترسی تو به سیاره محدود شد", "مدیریت سیاره «"+planetName+"» حضور تو را معلق کرد. دلیل: "+in.Reason, "planet.member.suspend"
		if !suspend {
			title, body, action = "دسترسی تو به سیاره برگشت", "دسترسی تو به سیاره «"+planetName+"» دوباره برقرار شد.", "planet.member.unsuspend"
		}
		at := time.Now().UTC()
		if err := persistJazebeNotification(ctx, tx, in.UserID, action+":"+in.PlanetID+":"+in.UserID+":"+at.Format(time.RFC3339Nano), title,
			map[string]any{"category": "critical", "priority": "urgent", "body": body, "planetId": in.PlanetID, "subjectType": "planet_membership"}, at); err != nil {
			return "", status.Error(codes.Internal, "member notice failed")
		}
		params, _ := json.Marshal(map[string]any{"reason": in.Reason})
		if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,brand_id,planet_id,params)
			VALUES($1,$2,$3,'user',$4,$5,$6,$7::jsonb)`, action, actorType, actor, in.UserID, brandID, in.PlanetID, string(params)); err != nil {
			return "", status.Error(codes.Internal, "member moderation audit failed")
		}
		// A suspended member stops receiving the planet's group chats at once.
		var groups []string
		if suspend {
			rows, err := tx.QueryContext(ctx, `SELECT group_id FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active'`, in.PlanetID, in.UserID)
			if err != nil {
				return "", status.Error(codes.Internal, "member groups unavailable")
			}
			for rows.Next() {
				var id string
				if err := rows.Scan(&id); err != nil {
					rows.Close()
					return "", status.Error(codes.Internal, "member groups unavailable")
				}
				groups = append(groups, id)
			}
			rows.Close()
		}
		if err := tx.Commit(); err != nil {
			return "", status.Error(codes.Internal, "member moderation failed")
		}
		for _, group := range groups {
			if err := groupChatKickUser(nk, groupRef{PlanetID: in.PlanetID, GroupID: group}, in.UserID); err != nil {
				logger.Warn("Jazebe group stream eviction failed: %v", err)
			}
		}
		return encode(map[string]any{"userId": in.UserID, "suspended": suspend})
	}
}

var (
	brandMemberSuspend   = brandMemberSetSuspended(true)
	brandMemberUnsuspend = brandMemberSetSuspended(false)
)

// brandSuspendedMembers lists the members suspended from the brand's planets,
// newest first, with display names only.
func brandSuspendedMembers(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		BrandID  string `json:"brandId"`
		PlanetID string `json:"planetId"`
		Page     int    `json:"page"`
		Limit    int    `json:"limit"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.BrandID, err = requiredUUID(in.BrandID); err != nil {
		return "", err
	}
	if in.PlanetID != "" {
		if in.PlanetID, err = requiredUUID(in.PlanetID); err != nil {
			return "", err
		}
	}
	if in.Page < 0 || in.Page > 100000 || in.Limit < 0 || in.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid page")
	}
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	if _, err := chatModerationRole(ctx, db, in.BrandID, actor); err != nil {
		return "", err
	}
	where := ` FROM jazebe_planet_membership m JOIN jazebe_planet p ON p.id=m.planet_id JOIN users u ON u.id=m.user_id
		WHERE p.brand_id=$1 AND NOT m.is_active AND ($2::uuid IS NULL OR m.planet_id=$2::uuid)`
	var total int
	if err := db.QueryRowContext(ctx, `SELECT count(*)`+where, in.BrandID, nullUUID(in.PlanetID)).Scan(&total); err != nil {
		return "", status.Error(codes.Internal, "suspended members unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT m.user_id,COALESCE(NULLIF(trim(u.display_name),''),'کاربر ناشناس'),p.id,p.name,m.suspended_at
		`+where+`
		ORDER BY m.suspended_at DESC NULLS LAST,m.user_id LIMIT $3 OFFSET $4`, in.BrandID, nullUUID(in.PlanetID), in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "suspended members unavailable")
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var user, name, planetID, planetName string
		var at sql.NullTime
		if err := rows.Scan(&user, &name, &planetID, &planetName, &at); err != nil {
			return "", status.Error(codes.Internal, "suspended members unavailable")
		}
		items = append(items, map[string]any{"userId": user, "displayName": name,
			"planet": map[string]string{"id": planetID, "name": planetName}, "suspendedAt": nullTime(at)})
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "suspended members unavailable")
	}
	return encode(map[string]any{"items": items, "total": total, "page": in.Page, "limit": in.Limit})
}
