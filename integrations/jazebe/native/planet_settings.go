package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Brand planet settings (source planet-settings.service.ts, PRD US-16): every
// write is validated against the platform policy as a whole merged row, a
// violation saves nothing, and new XP thresholds re-derive member levels
// upward only (PRD 10.2, US-77 rule 5).

type planetSettings struct {
	MissionRewardPointMin          int            `json:"missionRewardPointMin"`
	MissionRewardPointMax          int            `json:"missionRewardPointMax"`
	MissionRewardTicketMax         int            `json:"missionRewardTicketMax"`
	TicketPriceMin                 int            `json:"ticketPriceMin"`
	TicketPriceMax                 int            `json:"ticketPriceMax"`
	LotteryFeePercent              float64        `json:"lotteryFeePercent"`
	GroupMemberCapMin              int            `json:"groupMemberCapMin"`
	GroupMemberCapMax              int            `json:"groupMemberCapMax"`
	GroupMaxCount                  *int           `json:"groupMaxCount"`
	GroupAdminCap                  int            `json:"groupAdminCap"`
	GroupCreationEnabled           bool           `json:"groupCreationEnabled"`
	AllowedGroupModes              []string       `json:"allowedGroupModes"`
	XpLevelThresholds              []profileLevel `json:"xpLevelThresholds"`
	FreeGameDailyCap               *int           `json:"freeGameDailyCap"`
	XpPerPurchase                  int            `json:"xpPerPurchase"`
	XpPerGroupJoin                 int            `json:"xpPerGroupJoin"`
	DisplayNameChangeLimitPerMonth int            `json:"displayNameChangeLimitPerMonth"`
	PointToTomanRate               int            `json:"pointToTomanRate"`
}

const planetSettingsColumns = `mission_reward_point_min,mission_reward_point_max,mission_reward_ticket_max,ticket_price_min,ticket_price_max,
	lottery_fee_percent,group_member_cap_min,group_member_cap_max,group_max_count,group_admin_cap,group_creation_enabled,
	allowed_group_modes,COALESCE(xp_level_thresholds,'[]'::jsonb),free_game_daily_cap,xp_per_purchase,xp_per_group_join,
	display_name_change_limit_per_month,point_to_toman_rate`

func scanPlanetSettings(row interface{ Scan(...any) error }) (planetSettings, error) {
	var s planetSettings
	var maxCount, freeCap sql.NullInt64
	var modes, levels []byte
	err := row.Scan(&s.MissionRewardPointMin, &s.MissionRewardPointMax, &s.MissionRewardTicketMax, &s.TicketPriceMin, &s.TicketPriceMax,
		&s.LotteryFeePercent, &s.GroupMemberCapMin, &s.GroupMemberCapMax, &maxCount, &s.GroupAdminCap, &s.GroupCreationEnabled,
		&modes, &levels, &freeCap, &s.XpPerPurchase, &s.XpPerGroupJoin, &s.DisplayNameChangeLimitPerMonth, &s.PointToTomanRate)
	if err != nil {
		return s, err
	}
	if maxCount.Valid {
		v := int(maxCount.Int64)
		s.GroupMaxCount = &v
	}
	if freeCap.Valid {
		v := int(freeCap.Int64)
		s.FreeGameDailyCap = &v
	}
	_ = json.Unmarshal(modes, &s.AllowedGroupModes)
	_ = json.Unmarshal(levels, &s.XpLevelThresholds)
	if s.XpLevelThresholds == nil {
		s.XpLevelThresholds = []profileLevel{}
	}
	return s, nil
}

type platformPolicyBounds struct {
	MissionRewardPointMin, MissionRewardPointMax, MissionRewardTicketMax int
	TicketPriceMin, TicketPriceMax                                       int
	FeeMin, FeeMax                                                       float64
	GroupMemberCapMin, GroupMemberCapMax, GroupAdminCapMax               int
	GroupMaxCountMax                                                     sql.NullInt64
	DisplayNameChangeLimitPerMonth                                       int
}

type policyViolation struct {
	Field   string `json:"field"`
	Value   any    `json:"value"`
	Kind    string `json:"kind"`
	Bound   any    `json:"bound,omitempty"`
	Message string `json:"message"`
}

// validatePlanetSettings is the source validate(): bounds and cross-field rules.
func validatePlanetSettings(s planetSettings, p platformPolicyBounds) []policyViolation {
	v := []policyViolation{}
	lo := func(field string, value, bound int) {
		if value < bound {
			v = append(v, policyViolation{field, value, "min", bound, field + " < " + strconv.Itoa(bound)})
		}
	}
	hi := func(field string, value, bound int) {
		if value > bound {
			v = append(v, policyViolation{field, value, "max", bound, field + " > " + strconv.Itoa(bound)})
		}
	}
	rule := func(field string, ok bool, value any, message string) {
		if !ok {
			v = append(v, policyViolation{Field: field, Value: value, Kind: "rule", Message: message})
		}
	}
	lo("missionRewardPointMin", s.MissionRewardPointMin, p.MissionRewardPointMin)
	hi("missionRewardPointMin", s.MissionRewardPointMin, p.MissionRewardPointMax)
	lo("missionRewardPointMax", s.MissionRewardPointMax, p.MissionRewardPointMin)
	hi("missionRewardPointMax", s.MissionRewardPointMax, p.MissionRewardPointMax)
	rule("missionRewardPointMax", s.MissionRewardPointMax >= s.MissionRewardPointMin, s.MissionRewardPointMax, "missionRewardPointMax must be ≥ missionRewardPointMin")
	lo("missionRewardTicketMax", s.MissionRewardTicketMax, 0)
	hi("missionRewardTicketMax", s.MissionRewardTicketMax, p.MissionRewardTicketMax)
	lo("ticketPriceMin", s.TicketPriceMin, p.TicketPriceMin)
	hi("ticketPriceMin", s.TicketPriceMin, p.TicketPriceMax)
	lo("ticketPriceMax", s.TicketPriceMax, p.TicketPriceMin)
	hi("ticketPriceMax", s.TicketPriceMax, p.TicketPriceMax)
	rule("ticketPriceMax", s.TicketPriceMax >= s.TicketPriceMin, s.TicketPriceMax, "ticketPriceMax must be ≥ ticketPriceMin")
	rule("lotteryFeePercent", s.LotteryFeePercent >= p.FeeMin && s.LotteryFeePercent <= p.FeeMax, s.LotteryFeePercent,
		"lotteryFeePercent must be within "+strconv.FormatFloat(p.FeeMin, 'f', -1, 64)+"…"+strconv.FormatFloat(p.FeeMax, 'f', -1, 64))
	lo("groupMemberCapMin", s.GroupMemberCapMin, p.GroupMemberCapMin)
	hi("groupMemberCapMin", s.GroupMemberCapMin, p.GroupMemberCapMax)
	lo("groupMemberCapMax", s.GroupMemberCapMax, p.GroupMemberCapMin)
	hi("groupMemberCapMax", s.GroupMemberCapMax, p.GroupMemberCapMax)
	rule("groupMemberCapMax", s.GroupMemberCapMax >= s.GroupMemberCapMin, s.GroupMemberCapMax, "groupMemberCapMax must be ≥ groupMemberCapMin")
	if s.GroupMaxCount != nil {
		lo("groupMaxCount", *s.GroupMaxCount, 0)
		if p.GroupMaxCountMax.Valid {
			hi("groupMaxCount", *s.GroupMaxCount, int(p.GroupMaxCountMax.Int64))
		}
	} else {
		rule("groupMaxCount", !p.GroupMaxCountMax.Valid, nil, "groupMaxCount is required (platform cap "+strconv.FormatInt(p.GroupMaxCountMax.Int64, 10)+")")
	}
	lo("groupAdminCap", s.GroupAdminCap, 1)
	hi("groupAdminCap", s.GroupAdminCap, p.GroupAdminCapMax)
	modesOK := len(s.AllowedGroupModes) > 0
	for _, m := range s.AllowedGroupModes {
		modesOK = modesOK && (m == "open" || m == "approval")
	}
	rule("allowedGroupModes", modesOK, s.AllowedGroupModes, "allowedGroupModes must be a non-empty subset of open|approval")
	levels := s.XpLevelThresholds
	levelsOK := len(levels) == 0
	if len(levels) == 3 {
		sorted := append([]profileLevel(nil), levels...)
		for i := 1; i < len(sorted); i++ {
			for j := i; j > 0 && sorted[j].Level < sorted[j-1].Level; j-- {
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			}
		}
		levelsOK = sorted[0].MinXp == 0 && sorted[1].MinXp > sorted[0].MinXp && sorted[2].MinXp > sorted[1].MinXp
		for i, l := range sorted {
			levelsOK = levelsOK && l.Level == i+1 && l.MinXp >= 0 && strings.TrimSpace(l.TitleFa) != ""
		}
	}
	rule("xpLevelThresholds", levelsOK, levels, "xpLevelThresholds must be exactly 3 levels (1…3) with strictly ascending minXp starting at 0")
	if s.FreeGameDailyCap != nil {
		lo("freeGameDailyCap", *s.FreeGameDailyCap, 0)
	}
	lo("displayNameChangeLimitPerMonth", s.DisplayNameChangeLimitPerMonth, 0)
	hi("displayNameChangeLimitPerMonth", s.DisplayNameChangeLimitPerMonth, p.DisplayNameChangeLimitPerMonth)
	lo("pointToTomanRate", s.PointToTomanRate, 0)
	return v
}

// brandPlanetRole returns the caller's role on the planet's active brand.
func brandPlanetRole(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, planetID, user string) (string, error) {
	var role string
	err := q.QueryRowContext(ctx, `SELECT a.role FROM jazebe_planet p JOIN jazebe_active_brand_admin a ON a.brand_id=p.brand_id
		WHERE p.id=$1 AND a.user_id=$2 AND a.is_active`, planetID, user).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "brand authorization unavailable")
	}
	return role, nil
}

func planetIDFrom(payload string) (string, error) {
	var in struct {
		PlanetID string `json:"planetId"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	return requiredUUID(in.PlanetID)
}

func brandPlanetSettingsGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	s, err := scanPlanetSettings(db.QueryRowContext(ctx, `SELECT `+planetSettingsColumns+` FROM jazebe_planet_settings WHERE planet_id=$1`, planetID))
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet settings not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "planet settings unavailable")
	}
	return encode(s)
}

func brandPlanetSettingsUpdate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var patch map[string]json.RawMessage
	if err := decode(payload, &patch); err != nil {
		return "", err
	}
	var planetID string
	if json.Unmarshal(patch["planetId"], &planetID) != nil {
		return "", status.Error(codes.InvalidArgument, "planetId required")
	}
	if planetID, err = requiredUUID(planetID); err != nil {
		return "", err
	}
	delete(patch, "planetId")
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "planet settings unavailable")
	}
	defer tx.Rollback()
	role, err := brandPlanetRole(ctx, tx, planetID, user)
	if err != nil {
		return "", err
	}
	if role != "main" && role != "marketing" {
		return "", status.Error(codes.PermissionDenied, "brand economy role required")
	}
	current, err := scanPlanetSettings(tx.QueryRowContext(ctx, `SELECT `+planetSettingsColumns+` FROM jazebe_planet_settings WHERE planet_id=$1 FOR UPDATE`, planetID))
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet settings not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "planet settings unavailable")
	}
	// Merge the patch over the stored row through the JSON contract so every
	// field keeps its documented name and type.
	merged := map[string]json.RawMessage{}
	base, _ := json.Marshal(current)
	_ = json.Unmarshal(base, &merged)
	changed := []string{}
	for key, raw := range patch {
		if _, ok := merged[key]; !ok {
			return "", status.Error(codes.InvalidArgument, "unsupported settings field: "+key)
		}
		merged[key] = raw
		changed = append(changed, key)
	}
	mergedJSON, _ := json.Marshal(merged)
	var next planetSettings
	if err := json.Unmarshal(mergedJSON, &next); err != nil {
		return "", status.Error(codes.InvalidArgument, "invalid settings value")
	}
	if next.LotteryFeePercent < 0 || next.LotteryFeePercent > 100 || math.Round(next.LotteryFeePercent*100) != next.LotteryFeePercent*100 ||
		next.XpPerPurchase < 0 || next.XpPerPurchase > 1000 || next.XpPerGroupJoin < 0 || next.XpPerGroupJoin > 1000 {
		return "", status.Error(codes.InvalidArgument, "invalid settings value")
	}
	for _, l := range next.XpLevelThresholds {
		if !utf8.ValidString(l.TitleFa) || utf8.RuneCountInString(l.TitleFa) > 64 || len(l.Key) > 32 {
			return "", status.Error(codes.InvalidArgument, "invalid level")
		}
	}
	var p platformPolicyBounds
	if err := tx.QueryRowContext(ctx, `SELECT mission_reward_point_min,mission_reward_point_max,mission_reward_ticket_max,ticket_price_min,ticket_price_max,
		lottery_fee_percent_min,lottery_fee_percent_max,group_member_cap_min,group_member_cap_max,group_admin_cap_max,group_max_count_max,
		display_name_change_limit_per_month FROM jazebe_platform_policy WHERE key='default'`).Scan(&p.MissionRewardPointMin, &p.MissionRewardPointMax,
		&p.MissionRewardTicketMax, &p.TicketPriceMin, &p.TicketPriceMax, &p.FeeMin, &p.FeeMax, &p.GroupMemberCapMin, &p.GroupMemberCapMax,
		&p.GroupAdminCapMax, &p.GroupMaxCountMax, &p.DisplayNameChangeLimitPerMonth); err != nil {
		return "", status.Error(codes.Internal, "platform policy unavailable")
	}
	if violations := validatePlanetSettings(next, p); len(violations) > 0 {
		raw, _ := json.Marshal(violations)
		return "", status.Error(codes.InvalidArgument, "PLANET_SETTINGS_OUT_OF_POLICY "+string(raw))
	}
	modes, _ := json.Marshal(next.AllowedGroupModes)
	levels, _ := json.Marshal(next.XpLevelThresholds)
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet_settings SET mission_reward_point_min=$2,mission_reward_point_max=$3,
		mission_reward_ticket_max=$4,ticket_price_min=$5,ticket_price_max=$6,lottery_fee_percent=$7,group_member_cap_min=$8,
		group_member_cap_max=$9,group_max_count=$10,group_admin_cap=$11,group_creation_enabled=$12,allowed_group_modes=$13::jsonb,
		xp_level_thresholds=$14::jsonb,free_game_daily_cap=$15,xp_per_purchase=$16,xp_per_group_join=$17,
		display_name_change_limit_per_month=$18,point_to_toman_rate=$19 WHERE planet_id=$1`, planetID,
		next.MissionRewardPointMin, next.MissionRewardPointMax, next.MissionRewardTicketMax, next.TicketPriceMin, next.TicketPriceMax,
		next.LotteryFeePercent, next.GroupMemberCapMin, next.GroupMemberCapMax, next.GroupMaxCount, next.GroupAdminCap,
		next.GroupCreationEnabled, string(modes), string(levels), next.FreeGameDailyCap, next.XpPerPurchase, next.XpPerGroupJoin,
		next.DisplayNameChangeLimitPerMonth, next.PointToTomanRate); err != nil {
		logger.Error("Jazebe planet settings update failed: %v", err)
		return "", status.Error(codes.Internal, "planet settings update failed")
	}
	before, _ := json.Marshal(current.XpLevelThresholds)
	rederived := int64(0)
	if string(before) != string(levels) {
		if rederived, err = rederivePlanetLevels(ctx, tx, planetID, next.XpLevelThresholds); err != nil {
			return "", err
		}
	}
	raw, _ := json.Marshal(map[string]any{"fields": changed, "levelsRederived": rederived})
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('planet.settings.update','brand_admin',$1,'planet',$2,$2,$3::jsonb)`, user, planetID, string(raw)); err != nil {
		return "", status.Error(codes.Internal, "planet settings audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "planet settings update failed")
	}
	return encode(next)
}

// rederivePlanetLevels raises each member's stored level to what their XP
// now reaches; a member never loses a level (source rederiveLevels).
func rederivePlanetLevels(ctx context.Context, tx *sql.Tx, planetID string, levels []profileLevel) (int64, error) {
	if len(levels) != 3 {
		return 0, nil
	}
	sorted := append([]profileLevel(nil), levels...)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j].Level < sorted[j-1].Level; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	// Validation guarantees level 1 starts at 0 and the ladder ascends.
	res, err := tx.ExecContext(ctx, `WITH target AS (
			SELECT planet_id,user_id,CASE WHEN total_xp>=$3 THEN 3 WHEN total_xp>=$2 THEN 2 ELSE 1 END AS reached
			FROM jazebe_user_planet_stats WHERE planet_id=$1)
		UPDATE jazebe_user_planet_stats s SET level=t.reached,
			level_key=CASE t.reached WHEN 3 THEN $6 WHEN 2 THEN $5 ELSE $4 END
		FROM target t WHERE s.planet_id=t.planet_id AND s.user_id=t.user_id AND t.reached>s.level`,
		planetID, sorted[1].MinXp, sorted[2].MinXp, sorted[0].Key, sorted[1].Key, sorted[2].Key)
	if err != nil {
		return 0, status.Error(codes.Internal, "level rederivation failed")
	}
	return res.RowsAffected()
}

// brandPlanetSuspend lets a brand administrator with the planets action take
// an active planet offline (source planet-brand.service.ts suspend).
func brandPlanetSuspend(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		PlanetID string  `json:"planetId"`
		Reason   *string `json:"reason"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.PlanetID, err = requiredUUID(in.PlanetID); err != nil {
		return "", err
	}
	var reason any
	if in.Reason != nil {
		r := strings.TrimSpace(*in.Reason)
		if !utf8.ValidString(r) || utf8.RuneCountInString(r) > 512 {
			return "", status.Error(codes.InvalidArgument, "reason is too long")
		}
		if r != "" {
			reason = r
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "planet suspension unavailable")
	}
	defer tx.Rollback()
	role, err := brandPlanetRole(ctx, tx, in.PlanetID, user)
	if err != nil {
		return "", err
	}
	if role != "main" && role != "technical" {
		return "", status.Error(codes.PermissionDenied, "brand planets role required")
	}
	var from string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM jazebe_planet WHERE id=$1 FOR UPDATE`, in.PlanetID).Scan(&from); err != nil {
		return "", status.Error(codes.Internal, "planet unavailable")
	}
	if from != "active" {
		return "", status.Error(codes.FailedPrecondition, "invalid planet transition")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet SET status='suspended',suspended_at=now() WHERE id=$1`, in.PlanetID); err != nil {
		return "", status.Error(codes.Internal, "planet suspension failed")
	}
	raw, _ := json.Marshal(map[string]any{"from": from, "to": "suspended", "reason": reason, "by": "brand"})
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES('planet.suspend','brand_admin',$1,'planet',$2,$2,$3::jsonb)`, user, in.PlanetID, string(raw)); err != nil {
		return "", status.Error(codes.Internal, "planet audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "planet suspension failed")
	}
	return encode(map[string]any{"planetId": in.PlanetID, "status": "suspended", "from": from})
}
