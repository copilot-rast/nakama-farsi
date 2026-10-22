package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"
	_ "time/tzdata"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type missionDisplayConditions struct {
	MinLevel               int      `json:"minLevel"`
	Weekdays               []int    `json:"weekdays"`
	PrerequisiteMissionIDs []string `json:"prerequisiteMissionIds"`
	Time                   struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"time"`
}

func validateMissionDisplayConditions(cond *missionDisplayConditions) (any, error) {
	if cond == nil {
		return nil, nil
	}
	if cond.MinLevel < 0 || cond.MinLevel > 3 || len(cond.Weekdays) > 7 || len(cond.PrerequisiteMissionIDs) > 20 {
		return nil, status.Error(codes.InvalidArgument, "invalid mission display conditions")
	}
	seenDay := make(map[int]bool)
	for _, day := range cond.Weekdays {
		if day < 0 || day > 6 || seenDay[day] {
			return nil, status.Error(codes.InvalidArgument, "invalid mission weekdays")
		}
		seenDay[day] = true
	}
	seenMission := make(map[string]bool)
	for _, id := range cond.PrerequisiteMissionIDs {
		if _, err := requiredUUID(id); err != nil || seenMission[id] {
			return nil, status.Error(codes.InvalidArgument, "invalid prerequisite mission")
		}
		seenMission[id] = true
	}
	if _, err := parseMissionMinute(cond.Time.From, 0); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mission display time")
	}
	if _, err := parseMissionMinute(cond.Time.To, 1440); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mission display time")
	}
	encoded, err := json.Marshal(cond)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid mission display conditions")
	}
	return string(encoded), nil
}

func verifyMissionPrerequisites(ctx context.Context, tx *sql.Tx, planetID string, cond *missionDisplayConditions) error {
	if cond == nil {
		return nil
	}
	for _, id := range cond.PrerequisiteMissionIDs {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_mission WHERE planet_id=$1 AND id=$2)`, planetID, id).Scan(&exists); err != nil {
			return status.Error(codes.Internal, "mission prerequisite check failed")
		}
		if !exists {
			return status.Error(codes.FailedPrecondition, "prerequisite mission must belong to the same planet")
		}
	}
	return nil
}

type missionViewer struct {
	Level     int
	Completed map[string]bool
	Now       time.Time
}

type missionQueryRow interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func missionViewerFor(ctx context.Context, db missionQueryRow, planetID, userID string, now time.Time) (missionViewer, error) {
	viewer := missionViewer{Now: now, Completed: make(map[string]bool)}
	if err := db.QueryRowContext(ctx, `SELECT COALESCE((SELECT level FROM jazebe_user_planet_stats
		WHERE planet_id=$1 AND user_id=$2),1)`, planetID, userID).Scan(&viewer.Level); err != nil {
		return viewer, err
	}
	rows, err := db.QueryContext(ctx, `SELECT DISTINCT mission_id FROM jazebe_mission_completion
		WHERE planet_id=$1 AND user_id=$2 AND status='approved' AND deleted_at IS NULL`, planetID, userID)
	if err != nil {
		return viewer, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return viewer, err
		}
		viewer.Completed[id] = true
	}
	return viewer, rows.Err()
}

func tehranClock(now time.Time) (weekday int, minute int, err error) {
	zone, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		return 0, 0, err
	}
	local := now.In(zone)
	return (int(local.Weekday()) + 1) % 7, local.Hour()*60 + local.Minute(), nil
}

func parseMissionMinute(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	if len(value) != 5 || value[2] != ':' {
		return 0, errors.New("invalid mission time window")
	}
	hour, hourErr := strconv.Atoi(value[:2])
	minute, minuteErr := strconv.Atoi(value[3:])
	if hourErr != nil || minuteErr != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, errors.New("invalid mission time window")
	}
	return hour*60 + minute, nil
}

// Matches the source's all-of visibility rule. Unknown integration-specific
// data keys do not hide the mission; the brand API owns those decisions.
func missionDisplaySatisfied(raw []byte, viewer missionViewer) (bool, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return true, nil
	}
	var cond missionDisplayConditions
	if err := json.Unmarshal(raw, &cond); err != nil {
		return false, err
	}
	if cond.MinLevel > 0 && viewer.Level < cond.MinLevel {
		return false, nil
	}
	weekday, minute, err := tehranClock(viewer.Now)
	if err != nil {
		return false, err
	}
	if len(cond.Weekdays) > 0 {
		found := false
		for _, day := range cond.Weekdays {
			if day == weekday {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	for _, id := range cond.PrerequisiteMissionIDs {
		if !viewer.Completed[id] {
			return false, nil
		}
	}
	if cond.Time.From != "" || cond.Time.To != "" {
		from, err := parseMissionMinute(cond.Time.From, 0)
		if err != nil {
			return false, err
		}
		to, err := parseMissionMinute(cond.Time.To, 24*60)
		if err != nil {
			return false, err
		}
		if from <= to && (minute < from || minute >= to) {
			return false, nil
		}
		if from > to && (minute < from && minute >= to) {
			return false, nil
		}
	}
	return true, nil
}

func requireMissionVisible(ctx context.Context, tx *sql.Tx, ref missionRef, userID string, raw []byte, now time.Time) error {
	viewer, err := missionViewerFor(ctx, tx, ref.PlanetID, userID, now)
	if err != nil {
		return status.Error(codes.Internal, "mission visibility unavailable")
	}
	visible, err := missionDisplaySatisfied(raw, viewer)
	if err != nil {
		return status.Error(codes.Internal, "mission visibility unavailable")
	}
	if !visible {
		return status.Error(codes.NotFound, "mission not found")
	}
	return nil
}
