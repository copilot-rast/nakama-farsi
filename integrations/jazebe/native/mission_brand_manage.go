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

// Brand mission builder reads and edits (source mission-brand.service.ts:
// meta, get, update, decideMany). Every route needs the brand content role
// (main/technical/marketing) on the planet's active brand.

const missionReviewBulkMax = 100

func requireContentRole(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, planetID, user string) error {
	role, err := brandPlanetRole(ctx, q, planetID, user)
	if err != nil {
		return err
	}
	for _, action := range brandActions(role) {
		if action == "content" {
			return nil
		}
	}
	return status.Error(codes.PermissionDenied, "brand content role required")
}

func brandMissionMeta(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	planetID, err := planetIDFrom(payload)
	if err != nil {
		return "", err
	}
	if err := requireContentRole(ctx, db, planetID, user); err != nil {
		return "", err
	}
	var pointName string
	var pointMin, pointMax, ticketMax sql.NullInt64
	if err := db.QueryRowContext(ctx, `SELECT p.point_name,s.mission_reward_point_min,s.mission_reward_point_max,s.mission_reward_ticket_max
		FROM jazebe_planet p LEFT JOIN jazebe_planet_settings s ON s.planet_id=p.id WHERE p.id=$1`, planetID).Scan(&pointName, &pointMin, &pointMax, &ticketMax); err != nil {
		return "", status.Error(codes.Internal, "mission meta unavailable")
	}
	or := func(v sql.NullInt64, d int64) int64 {
		if v.Valid {
			return v.Int64
		}
		return d
	}
	rows, err := db.QueryContext(ctx, `SELECT id,title,status,draw_at FROM jazebe_lottery WHERE planet_id=$1 AND status IN ('draft','active') ORDER BY draw_at ASC`, planetID)
	if err != nil {
		return "", status.Error(codes.Internal, "mission meta unavailable")
	}
	defer rows.Close()
	lotteries := []map[string]any{}
	for rows.Next() {
		var id, title, state string
		var draw time.Time
		if err := rows.Scan(&id, &title, &state, &draw); err != nil {
			return "", status.Error(codes.Internal, "mission meta unavailable")
		}
		lotteries = append(lotteries, map[string]any{"id": id, "title": title, "status": state, "drawAt": draw})
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "mission meta unavailable")
	}
	return encode(map[string]any{"pointName": pointName,
		"bounds":    map[string]int64{"pointMin": or(pointMin, 0), "pointMax": or(pointMax, 100), "ticketMax": or(ticketMax, 5)},
		"lotteries": lotteries})
}

const brandMissionColumns = `m.id,m.planet_id,m.title,m.short_description,m.long_description,m.icon,m.functional_type,m.target_type,
	m.repeat_pattern,m.repeat_period,m.detection_method,m.reward_points,m.reward_tickets,m.reward_xp,m.target_lottery_id::text,
	m.group_target_value,m.display_conditions,m.type_config,m.starts_at,m.ends_at,m.status,m.sort_order,m.created_at,
	(SELECT count(*) FROM jazebe_mission_completion c WHERE c.mission_id=m.id AND c.status='approved'),
	(SELECT count(*) FROM jazebe_mission_completion c WHERE c.mission_id=m.id AND c.status='pending_review'),
	(SELECT count(*) FROM jazebe_mission_code k WHERE k.mission_id=m.id),
	(SELECT COALESCE(sum(k.used_count),0) FROM jazebe_mission_code k WHERE k.mission_id=m.id)`

func brandMissionRow(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, planetID, missionID string) (map[string]any, error) {
	var id, planet, title, functional, target, repeat, detection, state string
	var short, long, icon, period, lottery sql.NullString
	var points, tickets, xp, sortOrder, completions, pending, codesTotal, codesUsed int
	var groupTarget sql.NullInt64
	var display, config []byte
	var starts, ends, created time.Time
	err := q.QueryRowContext(ctx, `SELECT `+brandMissionColumns+` FROM jazebe_mission m WHERE m.id=$1 AND m.planet_id=$2`, missionID, planetID).Scan(
		&id, &planet, &title, &short, &long, &icon, &functional, &target, &repeat, &period, &detection, &points, &tickets, &xp, &lottery,
		&groupTarget, &display, &config, &starts, &ends, &state, &sortOrder, &created, &completions, &pending, &codesTotal, &codesUsed)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "mission not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "mission unavailable")
	}
	var displayValue, configValue any
	_ = json.Unmarshal(display, &displayValue)
	_ = json.Unmarshal(config, &configValue)
	if configValue == nil {
		configValue = map[string]any{}
	}
	var groupTargetValue any
	if groupTarget.Valid {
		groupTargetValue = groupTarget.Int64
	}
	return map[string]any{"id": id, "planetId": planet, "title": title, "shortDescription": nullable(short), "longDescription": nullable(long),
		"icon": nullable(icon), "functionalType": functional, "targetType": target, "repeatPattern": repeat, "repeatPeriod": nullable(period),
		"detectionMethod": detection, "rewardPoints": points, "rewardTickets": tickets, "rewardXp": xp, "targetLotteryId": nullable(lottery),
		"groupTargetValue": groupTargetValue, "displayConditions": displayValue, "typeConfig": configValue, "startsAt": starts, "endsAt": ends,
		"status": state, "sortOrder": sortOrder, "createdAt": created,
		"stats": map[string]int{"completions": completions, "pendingReview": pending, "codesTotal": codesTotal, "codesUsed": codesUsed}}, nil
}

func brandMissionGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readMissionRef(payload)
	if err != nil {
		return "", err
	}
	if err := requireContentRole(ctx, db, ref.PlanetID, user); err != nil {
		return "", err
	}
	row, err := brandMissionRow(ctx, db, ref.PlanetID, ref.MissionID)
	if err != nil {
		return "", err
	}
	return encode(row)
}

// brandMissionUpdate edits the fields every mission type shares. Shape and
// type-specific fields (targetType, repeat, detection, groupTargetValue,
// typeConfig) are refused: their per-type merge is not ported yet.
func brandMissionUpdate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in map[string]json.RawMessage
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	var ref missionRef
	_ = json.Unmarshal(in["planetId"], &ref.PlanetID)
	_ = json.Unmarshal(in["missionId"], &ref.MissionID)
	if ref.PlanetID, err = requiredUUID(ref.PlanetID); err != nil {
		return "", err
	}
	if ref.MissionID, err = requiredUUID(ref.MissionID); err != nil {
		return "", err
	}
	delete(in, "planetId")
	delete(in, "missionId")
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "mission update unavailable")
	}
	defer tx.Rollback()
	if err := requireContentRole(ctx, tx, ref.PlanetID, user); err != nil {
		return "", err
	}
	var state, target string
	var points, tickets int
	var lottery sql.NullString
	var starts, ends time.Time
	err = tx.QueryRowContext(ctx, `SELECT status,target_type,reward_points,reward_tickets,target_lottery_id::text,starts_at,ends_at
		FROM jazebe_mission WHERE id=$1 AND planet_id=$2 FOR UPDATE`, ref.MissionID, ref.PlanetID).Scan(&state, &target, &points, &tickets, &lottery, &starts, &ends)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "mission not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "mission update unavailable")
	}
	if state == "suspended" {
		return "", status.Error(codes.FailedPrecondition, "suspended mission is controlled by operator")
	}
	sets := []string{}
	args := []any{ref.MissionID}
	changed := []string{}
	add := func(column string, value any) {
		args = append(args, value)
		sets = append(sets, column+"=$"+stringInt(len(args)))
	}
	text := func(key string, min, max int) (any, error) {
		var v *string
		if json.Unmarshal(in[key], &v) != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid "+key)
		}
		if v == nil {
			if min > 0 {
				return nil, status.Error(codes.InvalidArgument, key+" cannot be null")
			}
			return nil, nil
		}
		s := strings.TrimSpace(*v)
		if !utf8.ValidString(s) || utf8.RuneCountInString(s) < min || utf8.RuneCountInString(s) > max {
			return nil, status.Error(codes.InvalidArgument, "invalid "+key)
		}
		return s, nil
	}
	integer := func(key string, min, max int) (int, error) {
		var v int
		if json.Unmarshal(in[key], &v) != nil || v < min || v > max {
			return 0, status.Error(codes.InvalidArgument, "invalid "+key)
		}
		return v, nil
	}
	for key := range in {
		changed = append(changed, key)
		switch key {
		case "title":
			v, err := text(key, 1, 128)
			if err != nil {
				return "", err
			}
			add("title", v)
		case "shortDescription":
			v, err := text(key, 0, 512)
			if err != nil {
				return "", err
			}
			add("short_description", v)
		case "longDescription":
			v, err := text(key, 0, 8000)
			if err != nil {
				return "", err
			}
			add("long_description", v)
		case "icon":
			v, err := text(key, 0, 512)
			if err != nil {
				return "", err
			}
			add("icon", v)
		case "sortOrder":
			v, err := integer(key, 0, 1_000_000)
			if err != nil {
				return "", err
			}
			add("sort_order", v)
		case "rewardXp":
			v, err := integer(key, 0, 1_000_000)
			if err != nil {
				return "", err
			}
			add("reward_xp", v)
		case "rewardPoints":
			if points, err = integer(key, 0, 1_000_000); err != nil {
				return "", err
			}
		case "rewardTickets":
			if tickets, err = integer(key, 0, 1_000_000); err != nil {
				return "", err
			}
		case "targetLotteryId":
			var v *string
			if json.Unmarshal(in[key], &v) != nil {
				return "", status.Error(codes.InvalidArgument, "invalid targetLotteryId")
			}
			lottery = sql.NullString{}
			if v != nil && *v != "" {
				id, err := requiredUUID(*v)
				if err != nil {
					return "", err
				}
				lottery = sql.NullString{String: id, Valid: true}
			}
		case "startsAt", "endsAt":
			var v string
			at, err := time.Parse(time.RFC3339, func() string { _ = json.Unmarshal(in[key], &v); return v }())
			if err != nil {
				return "", status.Error(codes.InvalidArgument, "invalid "+key)
			}
			if key == "startsAt" {
				starts = at
			} else {
				ends = at
			}
		case "displayConditions":
			var cond *missionDisplayConditions
			if json.Unmarshal(in[key], &cond) != nil {
				return "", status.Error(codes.InvalidArgument, "invalid displayConditions")
			}
			value, err := validateMissionDisplayConditions(cond)
			if err != nil {
				return "", err
			}
			if cond != nil {
				for _, id := range cond.PrerequisiteMissionIDs {
					if id == ref.MissionID {
						return "", status.Error(codes.InvalidArgument, "a mission cannot require itself")
					}
				}
				var found int
				if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_mission WHERE planet_id=$1 AND id::text = ANY($2::text[])`, ref.PlanetID, cond.PrerequisiteMissionIDs).Scan(&found); err != nil {
					return "", status.Error(codes.Internal, "mission update unavailable")
				}
				if found != len(cond.PrerequisiteMissionIDs) {
					return "", status.Error(codes.InvalidArgument, "prerequisite missions must belong to this planet")
				}
			}
			raw, _ := json.Marshal(value)
			if value == nil {
				add("display_conditions", nil)
			} else {
				add("display_conditions", string(raw))
				sets[len(sets)-1] += "::jsonb"
			}
		default:
			return "", status.Error(codes.InvalidArgument, "field not editable: "+key)
		}
	}
	if !ends.After(starts) {
		return "", status.Error(codes.InvalidArgument, "invalid end")
	}
	if err := validateMissionTargetRewards(target, 1, points, tickets, lottery.String); err != nil {
		return "", err
	}
	var pointMin, pointMax, ticketMax sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT mission_reward_point_min,mission_reward_point_max,mission_reward_ticket_max FROM jazebe_planet_settings WHERE planet_id=$1`, ref.PlanetID).Scan(&pointMin, &pointMax, &ticketMax); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "planet settings unavailable")
	}
	lo, hi, tmax := int64(0), int64(100), int64(5)
	if pointMin.Valid {
		lo, hi, tmax = pointMin.Int64, pointMax.Int64, ticketMax.Int64
	}
	if points > 0 && (int64(points) < lo || int64(points) > hi) || int64(tickets) > tmax {
		return "", status.Error(codes.InvalidArgument, "mission reward outside planet bounds")
	}
	if lottery.Valid {
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_lottery WHERE id=$1 AND planet_id=$2 AND status IN ('draft','active'))`, lottery.String, ref.PlanetID).Scan(&ok); err != nil {
			return "", status.Error(codes.Internal, "lottery unavailable")
		}
		if !ok {
			return "", status.Error(codes.InvalidArgument, "invalid target lottery")
		}
	}
	// The planet keeps an active solo point mission (source assertNotLastSoloPointMission).
	if state == "active" && !(target == "solo" && points > 0 && tickets == 0) {
		var other bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_mission WHERE planet_id=$1 AND id<>$2
			AND status='active' AND target_type='solo' AND reward_points>0 AND reward_tickets=0)`, ref.PlanetID, ref.MissionID).Scan(&other); err != nil {
			return "", status.Error(codes.Internal, "mission update unavailable")
		}
		if !other {
			return "", status.Error(codes.FailedPrecondition, "last solo point mission must remain active")
		}
	}
	add("reward_points", points)
	add("reward_tickets", tickets)
	add("target_lottery_id", nullUUID(lottery.String))
	sets[len(sets)-1] += "::uuid"
	add("starts_at", starts)
	add("ends_at", ends)
	// Column names come from the fixed cases above, never from the payload.
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_mission SET `+strings.Join(sets, ",")+` WHERE id=$1`, args...); err != nil {
		logger.Error("Jazebe mission update failed: %v", err)
		return "", status.Error(codes.Internal, "mission update failed")
	}
	raw, _ := json.Marshal(map[string]any{"fields": changed, "rewardPoints": points, "rewardTickets": tickets, "status": state})
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('mission.update','brand_admin',$1,'mission',$2,$3,$4::jsonb)`, user, ref.MissionID, ref.PlanetID, string(raw)); err != nil {
		return "", status.Error(codes.Internal, "mission audit failed")
	}
	row, err := brandMissionRow(ctx, tx, ref.PlanetID, ref.MissionID)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "mission update failed")
	}
	return encode(row)
}

// brandMissionReviewBulk applies one decision to up to 100 completions; each
// is decided on its own and reported, as in the source decideMany.
func brandMissionReviewBulk(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		PlanetID      string   `json:"planetId"`
		CompletionIDs []string `json:"completionIds"`
		Decision      string   `json:"decision"`
		Reason        string   `json:"reason"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.PlanetID, err = requiredUUID(in.PlanetID); err != nil {
		return "", err
	}
	// Authorize once up front so a non-administrator gets a denial, not a
	// list of per-item failures.
	if err := requireContentRole(ctx, db, in.PlanetID, user); err != nil {
		return "", err
	}
	if len(in.CompletionIDs) == 0 || len(in.CompletionIDs) > missionReviewBulkMax {
		return "", status.Error(codes.InvalidArgument, "completionIds must hold 1 to 100 items")
	}
	for _, id := range in.CompletionIDs {
		if _, err := requiredUUID(id); err != nil {
			return "", err
		}
	}
	out := []map[string]any{}
	for _, id := range in.CompletionIDs {
		_, err := mediaMissionReview(ctx, logger, db, nk, jsonString(map[string]string{"planetId": in.PlanetID, "completionId": id, "decision": in.Decision, "reason": in.Reason}))
		if err == nil {
			out = append(out, map[string]any{"id": id, "ok": true})
			continue
		}
		if st, ok := status.FromError(err); ok && st.Code() == codes.InvalidArgument {
			return "", err
		}
		code := "CONFLICT"
		if st, ok := status.FromError(err); ok {
			code = upperSnake(st.Code().String())
		}
		out = append(out, map[string]any{"id": id, "ok": false, "code": code})
	}
	return encode(map[string]any{"results": out})
}

// upperSnake turns a gRPC code name (NotFound) into NOT_FOUND.
func upperSnake(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}
