package main

import (
	"context"
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// A suspended planet membership bars group actions even when the group seat
// still exists. An absent planet membership is valid: joining can create it.
func requireUsablePlanetMembership(ctx context.Context, tx *sql.Tx, planetID, userID string) error {
	var suspended bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet_membership
		WHERE planet_id=$1 AND user_id=$2 AND NOT is_active)`, planetID, userID).Scan(&suspended); err != nil {
		return status.Error(codes.Internal, "planet membership unavailable")
	}
	if suspended {
		return status.Error(codes.PermissionDenied, "planet membership suspended")
	}
	return nil
}
