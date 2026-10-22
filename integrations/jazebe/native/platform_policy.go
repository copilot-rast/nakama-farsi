package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var policyTermsPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,32}$`)

type policyField struct {
	column   string
	min, max float64
	decimal  bool
	nullable bool
}

var policyFields = map[string]policyField{
	"missionRewardPointMin":          {"mission_reward_point_min", 0, math.MaxInt32, false, false},
	"missionRewardPointMax":          {"mission_reward_point_max", 0, math.MaxInt32, false, false},
	"missionRewardTicketMax":         {"mission_reward_ticket_max", 0, math.MaxInt32, false, false},
	"missionRewardXpMax":             {"mission_reward_xp_max", 0, math.MaxInt32, false, false},
	"ticketPriceMin":                 {"ticket_price_min", 0, math.MaxInt32, false, false},
	"ticketPriceMax":                 {"ticket_price_max", 0, math.MaxInt32, false, false},
	"lotteryFeePercentMin":           {"lottery_fee_percent_min", 0, 100, true, false},
	"lotteryFeePercentMax":           {"lottery_fee_percent_max", 0, 100, true, false},
	"groupMemberCapMin":              {"group_member_cap_min", 1, math.MaxInt32, false, false},
	"groupMemberCapMax":              {"group_member_cap_max", 1, math.MaxInt32, false, false},
	"groupAdminCapMax":               {"group_admin_cap_max", 1, math.MaxInt32, false, false},
	"groupMaxCountMax":               {"group_max_count_max", 1, math.MaxInt32, false, true},
	"otpLength":                      {"otp_length", 4, 8, false, false},
	"otpTtlSeconds":                  {"otp_ttl_seconds", 30, math.MaxInt32, false, false},
	"otpMaxAttempts":                 {"otp_max_attempts", 1, math.MaxInt32, false, false},
	"otpResendCooldownSeconds":       {"otp_resend_cooldown_seconds", 0, math.MaxInt32, false, false},
	"rateLimitPerMinute":             {"rate_limit_per_minute", 1, math.MaxInt32, false, false},
	"displayNameChangeLimitPerMonth": {"display_name_change_limit_per_month", 0, math.MaxInt32, false, false},
}

func validatePolicyPatch(fields map[string]json.RawMessage) (map[string]any, error) {
	parsed := map[string]any{}
	for key, raw := range fields {
		if key == "termsVersion" {
			var value string
			if json.Unmarshal(raw, &value) != nil || !policyTermsPattern.MatchString(value) {
				return nil, status.Error(codes.InvalidArgument, "invalid terms version")
			}
			parsed[key] = value
			continue
		}
		rule, ok := policyFields[key]
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "unknown platform policy field")
		}
		if string(raw) == "null" && rule.nullable {
			parsed[key] = nil
			continue
		}
		if string(raw) == "null" {
			return nil, status.Error(codes.InvalidArgument, "platform policy field cannot be null")
		}
		var input any
		if json.Unmarshal(raw, &input) != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid platform policy value")
		}
		n, ok := medalNumber(input)
		if !ok || math.IsNaN(n) || math.IsInf(n, 0) || n < rule.min || n > rule.max {
			return nil, status.Error(codes.InvalidArgument, "invalid platform policy value")
		}
		if rule.decimal {
			if math.Abs(math.Round(n*100)-n*100) > 1e-8 {
				return nil, status.Error(codes.InvalidArgument, "policy percent has more than two decimals")
			}
		}
		if !rule.decimal && math.Trunc(n) != n {
			return nil, status.Error(codes.InvalidArgument, "policy field requires integer")
		}
		parsed[key] = n
	}
	return parsed, nil
}

type platformPolicy struct {
	ID                             string          `json:"id"`
	Key                            string          `json:"key"`
	MissionRewardPointMin          int             `json:"missionRewardPointMin"`
	MissionRewardPointMax          int             `json:"missionRewardPointMax"`
	MissionRewardTicketMax         int             `json:"missionRewardTicketMax"`
	MissionRewardXpMax             int             `json:"missionRewardXpMax"`
	TicketPriceMin                 int             `json:"ticketPriceMin"`
	TicketPriceMax                 int             `json:"ticketPriceMax"`
	LotteryFeePercentMin           float64         `json:"lotteryFeePercentMin"`
	LotteryFeePercentMax           float64         `json:"lotteryFeePercentMax"`
	GroupMemberCapMin              int             `json:"groupMemberCapMin"`
	GroupMemberCapMax              int             `json:"groupMemberCapMax"`
	GroupAdminCapMax               int             `json:"groupAdminCapMax"`
	GroupMaxCountMax               *int            `json:"groupMaxCountMax"`
	OtpLength                      int             `json:"otpLength"`
	OtpTtlSeconds                  int             `json:"otpTtlSeconds"`
	OtpMaxAttempts                 int             `json:"otpMaxAttempts"`
	OtpResendCooldownSeconds       int             `json:"otpResendCooldownSeconds"`
	RateLimitPerMinute             int             `json:"rateLimitPerMinute"`
	DisplayNameChangeLimitPerMonth int             `json:"displayNameChangeLimitPerMonth"`
	TermsVersion                   string          `json:"termsVersion"`
	Extra                          json.RawMessage `json:"extra"`
	UpdatedBy                      *string         `json:"updatedBy"`
	CreatedAt                      time.Time       `json:"createdAt"`
	UpdatedAt                      time.Time       `json:"updatedAt"`
}

const platformPolicySelect = `SELECT id,key,mission_reward_point_min,mission_reward_point_max,mission_reward_ticket_max,mission_reward_xp_max,
	 ticket_price_min,ticket_price_max,lottery_fee_percent_min::float8,lottery_fee_percent_max::float8,
	 group_member_cap_min,group_member_cap_max,group_admin_cap_max,group_max_count_max,
	 otp_length,otp_ttl_seconds,otp_max_attempts,otp_resend_cooldown_seconds,rate_limit_per_minute,
	 display_name_change_limit_per_month,terms_version,extra,updated_by,created_at,updated_at
	 FROM jazebe_platform_policy WHERE key='default'`

func scanPlatformPolicy(scanner interface{ Scan(...any) error }) (platformPolicy, error) {
	var row platformPolicy
	var group sql.NullInt64
	var updated sql.NullString
	var extra []byte
	err := scanner.Scan(&row.ID, &row.Key, &row.MissionRewardPointMin, &row.MissionRewardPointMax, &row.MissionRewardTicketMax, &row.MissionRewardXpMax,
		&row.TicketPriceMin, &row.TicketPriceMax, &row.LotteryFeePercentMin, &row.LotteryFeePercentMax,
		&row.GroupMemberCapMin, &row.GroupMemberCapMax, &row.GroupAdminCapMax, &group,
		&row.OtpLength, &row.OtpTtlSeconds, &row.OtpMaxAttempts, &row.OtpResendCooldownSeconds, &row.RateLimitPerMinute,
		&row.DisplayNameChangeLimitPerMonth, &row.TermsVersion, &extra, &updated, &row.CreatedAt, &row.UpdatedAt)
	if group.Valid {
		v := int(group.Int64)
		row.GroupMaxCountMax = &v
	}
	if updated.Valid {
		row.UpdatedBy = &updated.String
	}
	row.Extra = extra
	return row, err
}

func loadPlatformPolicy(ctx context.Context, db *sql.DB) (platformPolicy, error) {
	row, err := scanPlatformPolicy(db.QueryRowContext(ctx, platformPolicySelect))
	if errors.Is(err, sql.ErrNoRows) {
		return row, status.Error(codes.NotFound, "platform policy not found")
	}
	if err != nil {
		return row, status.Error(codes.Internal, "platform policy unavailable")
	}
	return row, nil
}

func platformPolicyGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, _ string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	row, err := loadPlatformPolicy(ctx, db)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func brandPlatformPolicyGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		BrandID string `json:"brandId"`
	}
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.BrandID, err = requiredUUID(in.BrandID)
	if err != nil {
		return "", err
	}
	var allowed bool
	if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_brand_admin WHERE brand_id=$1 AND user_id=$2 AND is_active)`, in.BrandID, actor).Scan(&allowed); err != nil {
		return "", status.Error(codes.Internal, "brand authorization unavailable")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	row, err := loadPlatformPolicy(ctx, db)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func platformPolicyUpdate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := missionOperator(ctx)
	if err != nil {
		return "", err
	}
	if !isOperator(actor, os.Getenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS")) && !isOperator(actor, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "managerial operator role required")
	}
	var input map[string]json.RawMessage
	if err = decode(payload, &input); err != nil {
		return "", err
	}
	parsed, err := validatePolicyPatch(input)
	if err != nil {
		return "", err
	}
	if len(parsed) == 0 {
		row, err := loadPlatformPolicy(ctx, db)
		if err != nil {
			return "", err
		}
		return encode(row)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "platform policy unavailable")
	}
	defer tx.Rollback()
	before, err := scanPlatformPolicy(tx.QueryRowContext(ctx, platformPolicySelect+` FOR UPDATE`))
	if err != nil {
		return "", status.Error(codes.Internal, "platform policy unavailable")
	}
	keys := make([]string, 0, len(parsed))
	for k := range parsed {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sets := make([]string, 0, len(keys)+2)
	args := []any{actor}
	changed := map[string]any{}
	for _, key := range keys {
		column := "terms_version"
		if rule, ok := policyFields[key]; ok {
			column = rule.column
		}
		sets = append(sets, fmt.Sprintf("%s=$%d", column, len(args)+1))
		args = append(args, parsed[key])
		changed[key] = parsed[key]
	}
	sets = append(sets, "updated_by=$1", "updated_at=now()")
	query := `UPDATE jazebe_platform_policy SET ` + strings.Join(sets, ",") + ` WHERE key='default'`
	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == pgerrcode.CheckViolation {
			return "", status.Error(codes.InvalidArgument, "platform policy bounds violated")
		}
		return "", status.Error(codes.Internal, "platform policy update failed")
	}
	params, _ := json.Marshal(changed)
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,params)
		VALUES('platform_policy.update','operator',$1,'platform_policy',$2,$3::jsonb)`, actor, before.ID, string(params)); err != nil {
		return "", status.Error(codes.Internal, "platform policy audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "platform policy update failed")
	}
	row, err := loadPlatformPolicy(ctx, db)
	if err != nil {
		return "", err
	}
	return encode(row)
}
