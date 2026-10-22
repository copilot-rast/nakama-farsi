package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type referralInvite struct {
	ID, PlanetID, Inviter       string
	MissionID, Invitee, GroupID sql.NullString
	Status                      string
	Config                      []byte
}

type referralSignupContextKey struct{}

func referralMissionAcceptAllowed(ctx context.Context) bool {
	return ctx.Value(referralSignupContextKey{}) == true
}

var referralTokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{6,64}$`)

type referralMissionConfig struct {
	Scope            string `json:"scope"`
	SuccessCondition string `json:"successCondition"`
	TargetCount      int    `json:"targetCount"`
	WindowDays       int    `json:"windowDays"`
}

func parseReferralMissionConfig(raw []byte) (referralMissionConfig, error) {
	var c referralMissionConfig
	if len(raw) > 0 && json.Unmarshal(raw, &c) != nil {
		return c, status.Error(codes.FailedPrecondition, "invalid invite mission configuration")
	}
	if c.Scope == "" {
		c.Scope = "planet"
	}
	if c.SuccessCondition == "" {
		c.SuccessCondition = "signup"
	}
	if c.Scope != "planet" && c.Scope != "group" || c.SuccessCondition != "signup" && c.SuccessCondition != "first_activity" || c.TargetCount < 0 || c.WindowDays < 0 {
		return c, status.Error(codes.FailedPrecondition, "invalid invite mission configuration")
	}
	return c, nil
}

func referralToken(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func referralIdentity(ctx context.Context, tx *sql.Tx, user string) (string, error) {
	var code string
	err := tx.QueryRowContext(ctx, `SELECT invite_code FROM jazebe_referral_identity WHERE user_id=$1`, user).Scan(&code)
	if err == nil {
		return code, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	for attempt := 0; attempt < 3; attempt++ {
		code, err = referralToken(9)
		if err != nil {
			return "", err
		}
		var saved string
		err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_referral_identity(user_id,invite_code) VALUES($1,$2)
			ON CONFLICT DO NOTHING RETURNING invite_code`, user, code).Scan(&saved)
		if err == nil {
			return saved, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
		err = tx.QueryRowContext(ctx, `SELECT invite_code FROM jazebe_referral_identity WHERE user_id=$1`, user).Scan(&saved)
		if err == nil {
			return saved, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	return "", status.Error(codes.ResourceExhausted, "referral code unavailable")
}

func referralMyCode(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	code, err := referralIdentity(ctx, tx, user)
	if err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return encode(map[string]any{"inviteCode": code})
}

func referralStats(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	var issued, registered, qualified int64
	err = db.QueryRowContext(ctx, `SELECT count(*) FILTER(WHERE invitee_user_id IS NULL),
		count(*) FILTER(WHERE status='registered'),count(*) FILTER(WHERE status='qualified')
		FROM jazebe_mission_invite WHERE planet_id=$1 AND inviter_user_id=$2`, planetID, user).Scan(&issued, &registered, &qualified)
	if err != nil {
		return "", err
	}
	return encode(map[string]any{"issuedLinks": issued, "registered": registered, "qualified": qualified})
}

func referralMissionLink(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readMissionRef(payload)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var cfg []byte
	var statusName, planetStatus string
	var start, end time.Time
	err = tx.QueryRowContext(ctx, `SELECT m.type_config,m.status,p.status,m.starts_at,m.ends_at FROM jazebe_mission m
		JOIN jazebe_planet p ON p.id=m.planet_id WHERE m.id=$1 AND m.planet_id=$2 AND m.functional_type='invite'`, ref.MissionID, ref.PlanetID).
		Scan(&cfg, &statusName, &planetStatus, &start, &end)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "invite mission not found")
	}
	if err != nil {
		return "", err
	}
	if statusName != "active" || planetStatus != "active" || now.Before(start) || !now.Before(end) {
		return "", status.Error(codes.FailedPrecondition, "invite mission closed")
	}
	c, err := parseReferralMissionConfig(cfg)
	if err != nil {
		return "", err
	}
	var member bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet_membership WHERE planet_id=$1 AND user_id=$2)`, ref.PlanetID, user).Scan(&member); err != nil {
		return "", err
	}
	if !member {
		return "", status.Error(codes.PermissionDenied, "planet membership required")
	}
	var group sql.NullString
	if c.Scope == "group" {
		err = tx.QueryRowContext(ctx, `SELECT group_id FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active'`, ref.PlanetID, user).Scan(&group)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.FailedPrecondition, "active group required")
		}
		if err != nil {
			return "", err
		}
	}
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "jazebe_referral_link:"+ref.MissionID+":"+user); err != nil {
		return "", err
	}
	var token string
	var inviteID string
	err = tx.QueryRowContext(ctx, `SELECT id,token FROM jazebe_mission_invite WHERE mission_id=$1 AND inviter_user_id=$2 AND invitee_user_id IS NULL FOR UPDATE`, ref.MissionID, user).Scan(&inviteID, &token)
	if errors.Is(err, sql.ErrNoRows) {
		token, err = referralToken(32)
		if err != nil {
			return "", err
		}
		inviteID = uuid.Must(uuid.NewV4()).String()
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_mission_invite(id,planet_id,mission_id,inviter_user_id,token,group_id)
			VALUES($1,$2,$3,$4,$5,$6)`, inviteID, ref.PlanetID, ref.MissionID, user, token, group)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_mission_invite SET group_id=$2,updated_at=$3 WHERE id=$1`, inviteID, group, now); err != nil {
		return "", err
	}
	var qualified int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_mission_invite WHERE mission_id=$1 AND inviter_user_id=$2
		AND (status='qualified' OR ($3::boolean AND status='registered')) AND completion_id IS NULL
		AND ($4::timestamptz IS NULL OR COALESCE(qualified_at,registered_at)>=$4)`,
		ref.MissionID, user, c.Scope == "planet" && c.SuccessCondition == "signup", referralWindow(c, now)).Scan(&qualified); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	base := strings.TrimRight(os.Getenv("JAZEBE_MAIN_APP_URL"), "/")
	url := base + "/i/" + token
	return encode(map[string]any{"current": qualified, "target": referralTarget(c), "link": url, "url": url})
}

func referralWantedStatuses(c referralMissionConfig) []string {
	if c.Scope == "planet" && c.SuccessCondition == "signup" {
		return []string{"registered", "qualified"}
	}
	return []string{"qualified"}
}
func referralWindow(c referralMissionConfig, now time.Time) any {
	if c.WindowDays <= 0 {
		return nil
	}
	return now.Add(-time.Duration(c.WindowDays) * 24 * time.Hour)
}
func referralTarget(c referralMissionConfig) int {
	if c.TargetCount < 1 {
		return 1
	}
	return c.TargetCount
}

func referralLookup(ctx context.Context, tx *sql.Tx, token string) (referralInvite, error) {
	var r referralInvite
	err := tx.QueryRowContext(ctx, `SELECT i.id,i.planet_id,i.inviter_user_id,i.mission_id,i.invitee_user_id,i.group_id,i.status,
		COALESCE(m.type_config,'{}'::jsonb) FROM jazebe_mission_invite i LEFT JOIN jazebe_mission m ON m.id=i.mission_id
		WHERE i.token=$1 FOR UPDATE OF i`, token).Scan(&r.ID, &r.PlanetID, &r.Inviter, &r.MissionID, &r.Invitee, &r.GroupID, &r.Status, &r.Config)
	return r, err
}

func referralAccept(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		Token string `json:"token"`
	}
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	token := strings.TrimSpace(in.Token)
	if !referralTokenPattern.MatchString(token) {
		return "", status.Error(codes.InvalidArgument, "invalid referral token")
	}
	now := time.Now().UTC()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "jazebe_referral_user:"+user); err != nil {
		return "", err
	}
	invite, err := referralLookup(ctx, tx, token)
	if errors.Is(err, sql.ErrNoRows) {
		var inviter string
		err = tx.QueryRowContext(ctx, `SELECT user_id FROM jazebe_referral_identity WHERE invite_code=$1`, token).Scan(&inviter)
		if errors.Is(err, sql.ErrNoRows) {
			return encode(map[string]any{"kind": "none"})
		}
		if err != nil {
			return "", err
		}
		if inviter == user {
			return encode(map[string]any{"kind": "none"})
		}
		if _, err = referralIdentity(ctx, tx, user); err != nil {
			return "", err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_referral_identity SET referred_by_user_id=$2,updated_at=$3
			WHERE user_id=$1 AND referred_by_user_id IS NULL AND user_id<>$2`, user, inviter, now); err != nil {
			return "", err
		}
		if err = tx.Commit(); err != nil {
			return "", err
		}
		return encode(map[string]any{"kind": "invite_code", "inviterUserId": inviter})
	}
	if err != nil {
		return "", err
	}
	if invite.Inviter == user {
		return encode(map[string]any{"kind": "none"})
	}
	if invite.MissionID.Valid && !referralMissionAcceptAllowed(ctx) {
		return "", status.Error(codes.FailedPrecondition, "accept mission invitation through group join-by-invite")
	}
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "jazebe_referral_attr:"+invite.Inviter+":"+user+":"+invite.MissionID.String); err != nil {
		return "", err
	}
	var already string
	err = tx.QueryRowContext(ctx, `SELECT id FROM jazebe_mission_invite WHERE inviter_user_id=$1 AND invitee_user_id=$2
		AND mission_id IS NOT DISTINCT FROM $3::uuid`, invite.Inviter, user, nullableInviteID(invite.MissionID)).Scan(&already)
	if err == nil {
		return encode(map[string]any{"kind": "none"})
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	inviteID := invite.ID
	if invite.Invitee.Valid {
		newToken, e := referralToken(32)
		if e != nil {
			return "", e
		}
		inviteID = uuid.Must(uuid.NewV4()).String()
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_mission_invite(id,planet_id,mission_id,inviter_user_id,invitee_user_id,token,group_id,status,registered_at)
			VALUES($1,$2,$3,$4,$5,$6,$7,'registered',$8)`, inviteID, invite.PlanetID, nullableInviteID(invite.MissionID), invite.Inviter, user, newToken, nullableInviteID(invite.GroupID), now)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_mission_invite SET invitee_user_id=$2,status='registered',registered_at=$3,updated_at=$3
			WHERE id=$1 AND status='pending'`, inviteID, user, now)
	}
	if err != nil {
		return "", err
	}
	if _, err = referralIdentity(ctx, tx, user); err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_referral_identity SET referred_by_user_id=$2,updated_at=$3
		WHERE user_id=$1 AND referred_by_user_id IS NULL AND user_id<>$2`, user, invite.Inviter, now); err != nil {
		return "", err
	}
	c, err := parseReferralMissionConfig(invite.Config)
	if err != nil {
		return "", err
	}
	settle := !invite.MissionID.Valid || c.Scope == "planet" && c.SuccessCondition == "signup"
	if c.Scope == "group" && invite.GroupID.Valid {
		var belongs bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active')`, invite.GroupID.String, user).Scan(&belongs); err != nil {
			return "", err
		}
		settle = belongs
	}
	if settle {
		if err = referralQualify(ctx, tx, inviteID, invite.PlanetID, invite.Inviter, now); err != nil {
			return "", err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return encode(map[string]any{"kind": "mission_invite", "inviteId": inviteID, "inviterUserId": invite.Inviter, "settled": settle})
}

func referralLanding(ctx context.Context, db *sql.DB, token string) (map[string]any, error) {
	none := map[string]any{"kind": "none", "planet": nil, "group": nil}
	if !referralTokenPattern.MatchString(token) {
		return none, nil
	}
	var planetID, planetName, planetState string
	var slug, groupID, groupName, groupState sql.NullString
	var cfg []byte
	err := db.QueryRowContext(ctx, `SELECT p.id,p.name,p.status,p.slug,i.group_id,g.name,g.status,COALESCE(m.type_config,'{}'::jsonb)
		FROM jazebe_mission_invite i JOIN jazebe_planet p ON p.id=i.planet_id
		LEFT JOIN jazebe_group g ON g.id=i.group_id LEFT JOIN jazebe_mission m ON m.id=i.mission_id
		WHERE i.token=$1`, token).Scan(&planetID, &planetName, &planetState, &slug, &groupID, &groupName, &groupState, &cfg)
	if err == nil {
		c, e := parseReferralMissionConfig(cfg)
		if e != nil {
			return nil, e
		}
		result := map[string]any{"kind": "mission_invite", "planet": nil, "group": nil}
		if planetState == "active" {
			result["planet"] = map[string]any{"id": planetID, "name": planetName, "slug": slug.String}
		}
		if c.Scope == "group" && groupID.Valid && groupState.String == "active" {
			result["group"] = map[string]any{"id": groupID.String, "name": groupName.String}
		}
		return result, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	var exists bool
	if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_referral_identity WHERE invite_code=$1)`, token).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		none["kind"] = "invite_code"
	}
	return none, nil
}

func registerReferralHTTP(initializer runtime.Initializer, db *sql.DB) error {
	return initializer.RegisterHttp("/v2/jazebe/referrals/{token}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		result, err := referralLanding(ctx, db, r.PathValue("token"))
		if err != nil {
			http.Error(w, "referral landing unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(result)
	}), http.MethodGet)
}

func nullableInviteID(v sql.NullString) any {
	if v.Valid {
		return v.String
	}
	return nil
}

func referralQualify(ctx context.Context, tx *sql.Tx, inviteID, planetID, inviter string, now time.Time) error {
	var missionID sql.NullString
	err := tx.QueryRowContext(ctx, `UPDATE jazebe_mission_invite SET status='qualified',qualified_at=$2,updated_at=$2
		WHERE id=$1 AND status='registered' RETURNING mission_id`, inviteID, now).Scan(&missionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if missionID.Valid {
		if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_referral_completion_job(mission_id,inviter_user_id,planet_id)
			VALUES($1,$2,$3) ON CONFLICT(mission_id,inviter_user_id) DO UPDATE SET pending=true,next_attempt_at=now(),last_error=NULL,updated_at=now()`,
			missionID.String, inviter, planetID); err != nil {
			return err
		}
	}
	_, err = evaluateMedalsInTx(ctx, tx, planetID, inviter, "referral", "", "", now)
	return err
}

// Called while the durable activity event is locked. A qualifying referral
// and any medal it triggers commit before that event is marked processed.
func settleReferralActivityTx(ctx context.Context, tx *sql.Tx, planetID, invitee string, now time.Time) error {
	rows, err := tx.QueryContext(ctx, `SELECT i.id,i.inviter_user_id,i.group_id,COALESCE(m.type_config,'{}'::jsonb)
		FROM jazebe_mission_invite i LEFT JOIN jazebe_mission m ON m.id=i.mission_id
		WHERE i.planet_id=$1 AND i.invitee_user_id=$2 AND i.status='registered' ORDER BY i.id`, planetID, invitee)
	if err != nil {
		return err
	}
	type pending struct {
		id, inviter string
		group       sql.NullString
		config      []byte
	}
	all := []pending{}
	for rows.Next() {
		var p pending
		if err = rows.Scan(&p.id, &p.inviter, &p.group, &p.config); err != nil {
			rows.Close()
			return err
		}
		all = append(all, p)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, p := range all {
		c, e := parseReferralMissionConfig(p.config)
		if e != nil {
			return e
		}
		qualified := c.Scope == "planet" && c.SuccessCondition == "first_activity"
		if c.Scope == "group" && p.group.Valid {
			if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_group_membership WHERE group_id=$1 AND user_id=$2 AND status='active')`, p.group.String, invitee).Scan(&qualified); err != nil {
				return err
			}
		}
		if qualified {
			if err = referralQualify(ctx, tx, p.id, planetID, p.inviter, now); err != nil {
				return err
			}
		}
	}
	return nil
}
