package main

import (
	"context"
	"database/sql"
	"unicode"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Caller-scoped medal wall from Jazebe's profile progression projection.
// Earned medals remain visible when paused; unearned paused medals do not.
func medalWall(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		PlanetID string `json:"planetId"`
	}
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	var exists bool
	if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet WHERE id=$1)`, planetID).Scan(&exists); err != nil {
		return "", status.Error(codes.Internal, "planet unavailable")
	}
	if !exists {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	rows, err := db.QueryContext(ctx, `SELECT l.name,l.icon,pm.status,a.acquired_at
		FROM jazebe_planet_medal pm JOIN jazebe_medal_library l ON l.id=pm.medal_library_id
		LEFT JOIN jazebe_medal_acquisition a ON a.planet_medal_id=pm.id AND a.user_id=$2
		WHERE pm.planet_id=$1 AND (pm.status='active' OR a.id IS NOT NULL)
		ORDER BY pm.created_at,pm.id`, planetID, user)
	if err != nil {
		return "", status.Error(codes.Internal, "medal wall unavailable")
	}
	defer rows.Close()
	medals := []map[string]any{}
	for rows.Next() {
		var name, state string
		var icon sql.NullString
		var acquired sql.NullTime
		if err = rows.Scan(&name, &icon, &state, &acquired); err != nil {
			return "", status.Error(codes.Internal, "medal wall unavailable")
		}
		glyph := "star"
		if icon.Valid {
			glyph = medalDisplayIcon(icon.String)
		}
		item := map[string]any{"fa": name, "icon": glyph, "earned": acquired.Valid}
		if acquired.Valid {
			item["date"] = lotteryDateFa(acquired.Time)
		}
		medals = append(medals, item)
	}
	if err = rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "medal wall unavailable")
	}
	return encode(medals)
}

func medalDisplayIcon(icon string) string {
	switch icon {
	case "target", "heart", "star", "compass", "bolt":
		return icon
	}
	for _, r := range icon {
		if unicode.Is(unicode.So, r) {
			return icon
		}
	}
	return "star"
}
