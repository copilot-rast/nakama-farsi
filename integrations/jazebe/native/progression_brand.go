package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Brand progression panel (source progression-brand.service.ts xpOverview,
// setLevels, setActivityXp, PRD US-76/77). Reads are open to every active
// brand administrator; writes need the economy role.

var defaultLevelNames = []profileLevel{{1, "newcomer", "تازه‌وارد", 0}, {2, "active", "عضو فعال", 0}, {3, "veteran", "کهنه‌کار", 0}}

func xpOverview(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, planetID string) (map[string]any, error) {
	var raw []byte
	var xpMax int
	if err := q.QueryRowContext(ctx, `SELECT COALESCE(s.xp_level_thresholds,'[]'::jsonb),(SELECT mission_reward_xp_max FROM jazebe_platform_policy WHERE key='default')
		FROM jazebe_planet_settings s WHERE s.planet_id=$1`, planetID).Scan(&raw, &xpMax); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "planet settings not found")
		}
		return nil, status.Error(codes.Internal, "progression unavailable")
	}
	var configured []profileLevel
	_ = json.Unmarshal(raw, &configured)
	ladder := defaultLevelNames
	if len(configured) == 3 {
		ladder = configured
	}
	counts := map[int]int{}
	rows, err := q.QueryContext(ctx, `SELECT level,count(*) FROM jazebe_user_planet_stats WHERE planet_id=$1 GROUP BY level`, planetID)
	if err != nil {
		return nil, status.Error(codes.Internal, "progression unavailable")
	}
	for rows.Next() {
		var level, n int
		if err := rows.Scan(&level, &n); err != nil {
			rows.Close()
			return nil, status.Error(codes.Internal, "progression unavailable")
		}
		counts[level] = n
	}
	rows.Close()
	levels := []map[string]any{}
	for _, l := range ladder {
		levels = append(levels, map[string]any{"level": l.Level, "key": l.Key, "titleFa": l.TitleFa, "minXp": l.MinXp, "members": counts[l.Level]})
	}
	activities := []map[string]any{}
	rows, err = q.QueryContext(ctx, `SELECT 'mission',id,title,status,reward_xp,reward_points,created_at FROM jazebe_mission WHERE planet_id=$1
		UNION ALL SELECT 'game',id,name,status,reward_xp,NULL,created_at FROM jazebe_game WHERE planet_id=$1
		ORDER BY 1 DESC, 7 DESC`, planetID)
	if err != nil {
		return nil, status.Error(codes.Internal, "progression unavailable")
	}
	defer rows.Close()
	for rows.Next() {
		var kind, id, title, state string
		var xp int
		var points sql.NullInt64
		var created any
		if err := rows.Scan(&kind, &id, &title, &state, &xp, &points, &created); err != nil {
			return nil, status.Error(codes.Internal, "progression unavailable")
		}
		var pointsValue any
		if points.Valid {
			pointsValue = points.Int64
		}
		activities = append(activities, map[string]any{"kind": kind, "id": id, "title": title, "status": state, "rewardXp": xp, "rewardPoints": pointsValue})
	}
	if rows.Err() != nil {
		return nil, status.Error(codes.Internal, "progression unavailable")
	}
	return map[string]any{"configured": len(configured) == 3, "levels": levels, "xpMax": xpMax, "activities": activities}, nil
}

func brandProgressionXp(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	planetID, err := planetIDFrom(payload)
	if err != nil {
		return "", err
	}
	if _, err := brandPlanetRole(ctx, db, planetID, user); err != nil {
		return "", err
	}
	out, err := xpOverview(ctx, db, planetID)
	if err != nil {
		return "", err
	}
	return encode(out)
}

func requireEconomyRole(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, planetID, user string) error {
	role, err := brandPlanetRole(ctx, q, planetID, user)
	if err != nil {
		return err
	}
	if role != "main" && role != "marketing" {
		return status.Error(codes.PermissionDenied, "brand economy role required")
	}
	return nil
}

// brandProgressionSetLevels sets where levels 2 and 3 start (the names are
// fixed, PRD 10.2) and re-derives member levels upward only.
func brandProgressionSetLevels(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		PlanetID    string `json:"planetId"`
		Level2MinXp int64  `json:"level2MinXp"`
		Level3MinXp int64  `json:"level3MinXp"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.PlanetID, err = requiredUUID(in.PlanetID); err != nil {
		return "", err
	}
	if !(in.Level2MinXp > 0 && in.Level3MinXp > in.Level2MinXp) {
		return "", status.Error(codes.InvalidArgument, "LEVEL_CONFIG_INVALID: level thresholds must ascend above 0")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "progression unavailable")
	}
	defer tx.Rollback()
	if err := requireEconomyRole(ctx, tx, in.PlanetID, user); err != nil {
		return "", err
	}
	ladder := append([]profileLevel(nil), defaultLevelNames...)
	ladder[1].MinXp, ladder[2].MinXp = in.Level2MinXp, in.Level3MinXp
	raw, _ := json.Marshal(ladder)
	res, err := tx.ExecContext(ctx, `UPDATE jazebe_planet_settings SET xp_level_thresholds=$2::jsonb WHERE planet_id=$1`, in.PlanetID, string(raw))
	if err != nil {
		return "", status.Error(codes.Internal, "level update failed")
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return "", status.Error(codes.NotFound, "planet settings not found")
	}
	rederived, err := rederivePlanetLevels(ctx, tx, in.PlanetID, ladder)
	if err != nil {
		return "", err
	}
	params, _ := json.Marshal(map[string]any{"thresholds": []int64{0, in.Level2MinXp, in.Level3MinXp}, "rederived": rederived})
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('planet.levels','brand_admin',$1,'planet',$2,$2,$3::jsonb)`, user, in.PlanetID, string(params)); err != nil {
		return "", status.Error(codes.Internal, "level audit failed")
	}
	out, err := xpOverview(ctx, tx, in.PlanetID)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "level update failed")
	}
	return encode(out)
}

// brandProgressionSetXp sets the XP of up to 500 missions and games of the
// planet, each within the platform's mission XP cap.
func brandProgressionSetXp(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		PlanetID string `json:"planetId"`
		Items    []struct {
			Kind     string `json:"kind"`
			ID       string `json:"id"`
			RewardXp int    `json:"rewardXp"`
		} `json:"items"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.PlanetID, err = requiredUUID(in.PlanetID); err != nil {
		return "", err
	}
	if len(in.Items) > 500 {
		return "", status.Error(codes.InvalidArgument, "at most 500 items")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "progression unavailable")
	}
	defer tx.Rollback()
	if err := requireEconomyRole(ctx, tx, in.PlanetID, user); err != nil {
		return "", err
	}
	var xpMax int
	if err := tx.QueryRowContext(ctx, `SELECT mission_reward_xp_max FROM jazebe_platform_policy WHERE key='default'`).Scan(&xpMax); err != nil {
		return "", status.Error(codes.Internal, "platform policy unavailable")
	}
	for _, item := range in.Items {
		if item.Kind != "mission" && item.Kind != "game" {
			return "", status.Error(codes.InvalidArgument, "kind must be mission or game")
		}
		if _, err := requiredUUID(item.ID); err != nil {
			return "", err
		}
		if item.RewardXp < 0 || item.RewardXp > 1_000_000 {
			return "", status.Error(codes.InvalidArgument, "invalid rewardXp")
		}
		if item.RewardXp > xpMax {
			return "", status.Error(codes.InvalidArgument, "PROGRESSION_OUT_OF_POLICY: rewardXp above "+stringInt(xpMax))
		}
		table := "jazebe_mission"
		if item.Kind == "game" {
			table = "jazebe_game"
		}
		// The table name comes from the fixed choice above.
		res, err := tx.ExecContext(ctx, `UPDATE `+table+` SET reward_xp=$3 WHERE id=$1 AND planet_id=$2`, item.ID, in.PlanetID, item.RewardXp)
		if err != nil {
			return "", status.Error(codes.Internal, "xp update failed")
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return "", status.Error(codes.NotFound, "activity not found")
		}
	}
	out, err := xpOverview(ctx, tx, in.PlanetID)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "xp update failed")
	}
	return encode(out)
}
