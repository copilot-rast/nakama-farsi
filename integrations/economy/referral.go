package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	referralAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	referralCodeLen  = 8
	referralMarkSalt = "nakama-referral-mark:"
)

type referralSettings struct {
	IsEnabled       bool    `json:"isEnabled"`
	NewAccountDays  int     `json:"newAccountDays"`
	ReferrerPrizeID string  `json:"referrerPrizeId"`
	RefereePrizeID  string  `json:"refereePrizeId"`
	GoalEventTypeID string  `json:"goalEventTypeId"`
	GoalMeasure     string  `json:"goalMeasure"`
	GoalValue       float64 `json:"goalValue"`
	MaxPerReferrer  int     `json:"maxPerReferrer"`
	DailyPerReferer int     `json:"dailyPerReferrer"`
	OnePerDevice    bool    `json:"onePerDevice"`
	OnePerPhone     bool    `json:"onePerPhone"`
	HoldFlagged     bool    `json:"holdFlagged"`
	UpdatedAt       string  `json:"updatedAt,omitempty"`
}

func loadReferralSettings(ctx context.Context, q queryer) (referralSettings, error) {
	st := referralSettings{IsEnabled: true, NewAccountDays: 7, GoalMeasure: "count", GoalValue: 1, MaxPerReferrer: 100, DailyPerReferer: 10, OnePerDevice: true, OnePerPhone: true, HoldFlagged: true}
	var ref, ree, goal sql.NullString
	var updated time.Time
	err := q.QueryRowContext(ctx, `SELECT is_enabled, new_account_days, referrer_prize_id::text, referee_prize_id::text, goal_event_type_id::text, goal_measure, goal_value::float8, max_per_referrer, daily_per_referrer, one_per_device, one_per_phone, hold_flagged, updated_at FROM economy_referral_settings WHERE id=1`).
		Scan(&st.IsEnabled, &st.NewAccountDays, &ref, &ree, &goal, &st.GoalMeasure, &st.GoalValue, &st.MaxPerReferrer, &st.DailyPerReferer, &st.OnePerDevice, &st.OnePerPhone, &st.HoldFlagged, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	st.ReferrerPrizeID, st.RefereePrizeID, st.GoalEventTypeID = ref.String, ree.String, goal.String
	st.UpdatedAt = updated.UTC().Format(time.RFC3339)
	return st, nil
}

func normalizeReferralCode(code string) string {
	code = strings.Map(func(r rune) rune {
		switch {
		case r >= '۰' && r <= '۹':
			return '0' + (r - '۰')
		case r >= '٠' && r <= '٩':
			return '0' + (r - '٠')
		case r == ' ' || r == '-' || r == '‌':
			return -1
		}
		return r
	}, strings.TrimSpace(code))
	return strings.ToUpper(code)
}

func newReferralCode() string {
	b := make([]byte, referralCodeLen)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(referralAlphabet))))
		b[i] = referralAlphabet[n.Int64()]
	}
	return string(b)
}

func (s *service) referralCode(ctx context.Context, userID string) (string, error) {
	var code string
	err := s.db.QueryRowContext(ctx, `SELECT code FROM economy_referral_code WHERE user_id=$1`, userID).Scan(&code)
	if err == nil {
		return code, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	for i := 0; i < 8; i++ {
		code = newReferralCode()
		_, err = s.db.ExecContext(ctx, `INSERT INTO economy_referral_code (user_id, code) VALUES ($1,$2) ON CONFLICT (user_id) DO NOTHING`, userID, code)
		if c, unique := isUniqueViolation(err); unique && strings.Contains(c, "code") {
			continue
		}
		if isForeignKeyViolation(err) {
			return "", notFound("USER_NOT_FOUND")
		}
		if err != nil {
			return "", err
		}
		break
	}
	err = s.db.QueryRowContext(ctx, `SELECT code FROM economy_referral_code WHERE user_id=$1`, userID).Scan(&code)
	return code, err
}

func markHash(kind, value string) string {
	sum := sha256.Sum256([]byte(referralMarkSalt + kind + ":" + strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func stringsOf(ctx context.Context, tx *sql.Tx, q string, args ...any) ([]string, error) {
	rows, err := tx.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var v sql.NullString
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		if strings.TrimSpace(v.String) != "" {
			out = append(out, v.String)
		}
	}
	return out, rows.Err()
}

func userPhones(ctx context.Context, tx *sql.Tx, userID string) ([]string, error) {
	return stringsOf(ctx, tx, `SELECT metadata->>'phone' FROM users WHERE id=$1
		UNION SELECT phone FROM publication_phone_account WHERE user_id=$1`, userID)
}

type referralOutcome struct {
	ID         string
	ReferrerID string
	Status     string
	Flags      []string
	Claims     []claimRef
}

func (s *service) completeReferral(ctx context.Context, tx *sql.Tx, id string) ([]claimRef, error) {
	var referrer, referee string
	var refPrize, reePrize sql.NullString
	err := tx.QueryRowContext(ctx, `UPDATE economy_referral SET status='completed', completed_at=now() WHERE id=$1 AND status IN ('pending','held')
		RETURNING referrer_id::text, referee_id::text, referrer_prize_id::text, referee_prize_id::text`, id).Scan(&referrer, &referee, &refPrize, &reePrize)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	claims := []claimRef{}
	for _, side := range []struct {
		user, scope string
		prize       sql.NullString
	}{{referrer, "referrer", refPrize}, {referee, "referee", reePrize}} {
		if !side.prize.Valid {
			continue
		}
		c, err := insertClaim(ctx, tx, side.user, "referral", id, id, side.scope, side.prize.String, "auto", 0)
		if err != nil {
			return nil, err
		}
		if c != nil {
			claims = append(claims, *c)
		}
	}
	return claims, nil
}

func (s *service) enterReferral(ctx context.Context, refereeID, rawCode, ip string) (*referralOutcome, error) {
	code := normalizeReferralCode(rawCode)
	if len(code) < 4 || len(code) > 16 {
		return nil, badRequest("CODE_NOT_FOUND", "code", "this code does not exist")
	}
	st, err := loadReferralSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if !st.IsEnabled {
		return nil, fail(http.StatusForbidden, "REFERRAL_DISABLED", "", "referral codes are turned off")
	}
	out := &referralOutcome{ID: uuid.Must(uuid.NewV4()).String()}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('referral:' || $1::text))`, refereeID); err != nil {
			return err
		}
		var referrer string
		err := tx.QueryRowContext(ctx, `SELECT user_id::text FROM economy_referral_code WHERE code=$1 FOR UPDATE`, code).Scan(&referrer)
		if errors.Is(err, sql.ErrNoRows) {
			return badRequest("CODE_NOT_FOUND", "code", "this code does not exist")
		}
		if err != nil {
			return err
		}
		if referrer == refereeID {
			return badRequest("SELF_REFERRAL", "code", "you cannot use your own code")
		}
		var already bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_referral WHERE referee_id=$1)`, refereeID).Scan(&already); err != nil {
			return err
		}
		if already {
			return conflict("ALREADY_REFERRED", "you have already entered a code")
		}
		var created time.Time
		var refDisabled time.Time
		if err := tx.QueryRowContext(ctx, `SELECT create_time FROM users WHERE id=$1`, refereeID).Scan(&created); err != nil {
			return err
		}
		if err := tx.QueryRowContext(ctx, `SELECT disable_time FROM users WHERE id=$1`, referrer).Scan(&refDisabled); err != nil {
			return badRequest("CODE_NOT_FOUND", "code", "this code does not exist")
		}
		if refDisabled.Unix() > 0 {
			return badRequest("CODE_NOT_FOUND", "code", "this code does not exist")
		}
		if st.NewAccountDays > 0 && s.now().Sub(created) > time.Duration(st.NewAccountDays)*24*time.Hour {
			return fail(http.StatusForbidden, "ACCOUNT_TOO_OLD", "", "only new accounts can enter a code")
		}
		var cycle bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_referral WHERE referrer_id=$1 AND referee_id=$2)`, refereeID, referrer).Scan(&cycle); err != nil {
			return err
		}
		if cycle {
			return badRequest("REFERRAL_CYCLE", "code", "you invited this player")
		}
		var total, today int
		if err := tx.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE created_at > now() - INTERVAL '1 day') FROM economy_referral WHERE referrer_id=$1 AND status <> 'rejected'`, referrer).Scan(&total, &today); err != nil {
			return err
		}
		if total >= st.MaxPerReferrer {
			return conflict("REFERRER_LIMIT", "this code cannot be used any more")
		}
		flags := []string{}
		if today >= st.DailyPerReferer {
			flags = append(flags, "burst")
		}
		phones, err := userPhones(ctx, tx, refereeID)
		if err != nil {
			return err
		}
		refPhones, err := userPhones(ctx, tx, referrer)
		if err != nil {
			return err
		}
		for _, p := range phones {
			for _, rp := range refPhones {
				if p == rp {
					return badRequest("SELF_REFERRAL", "code", "you cannot use your own code")
				}
			}
		}
		if ip != "" {
			var sameIP bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_activity WHERE user_id=$1 AND kind='signin' AND ip=$2 AND created_at > now() - INTERVAL '30 days')`, referrer, ip).Scan(&sameIP); err == nil && sameIP {
				flags = append(flags, "same_ip")
			}
			var fromIP int
			if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM economy_referral WHERE referrer_id=$1 AND ip=$2 AND created_at > now() - INTERVAL '1 day'`, referrer, ip).Scan(&fromIP); err == nil && fromIP >= 2 {
				flags = append(flags, "ip_burst")
			}
		}
		devices, err := stringsOf(ctx, tx, `SELECT id FROM user_device WHERE user_id=$1`, refereeID)
		if err != nil {
			return err
		}
		marks := []struct{ kind, value string }{}
		for _, d := range devices {
			marks = append(marks, struct{ kind, value string }{"device", d})
		}
		for _, p := range phones {
			marks = append(marks, struct{ kind, value string }{"phone", p})
		}
		for _, m := range marks {
			var got string
			err := tx.QueryRowContext(ctx, `INSERT INTO economy_referral_mark (kind, hash, referral_id) VALUES ($1,$2,$3) ON CONFLICT (kind, hash) DO NOTHING RETURNING kind`, m.kind, markHash(m.kind, m.value), out.ID).Scan(&got)
			if errors.Is(err, sql.ErrNoRows) {
				if (m.kind == "device" && st.OnePerDevice) || (m.kind == "phone" && st.OnePerPhone) {
					return conflict(strings.ToUpper(m.kind)+"_ALREADY_USED", "this "+m.kind+" was already used for a referral")
				}
				flags = append(flags, "shared_"+m.kind)
				continue
			}
			if err != nil {
				return err
			}
		}
		status := "completed"
		if len(flags) > 0 && st.HoldFlagged {
			status = "held"
		} else if st.GoalEventTypeID != "" {
			status = "pending"
		}
		insertStatus := status
		if status == "completed" {
			insertStatus = "pending"
		}
		var goal *string
		if st.GoalEventTypeID != "" {
			goal = &st.GoalEventTypeID
		}
		nullable := func(v string) *string {
			if v == "" {
				return nil
			}
			return &v
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_referral (id, referrer_id, referee_id, code, status, flags, referrer_prize_id, referee_prize_id, goal_event_type_id, goal_measure, goal_value, ip)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::numeric,$12)`,
			out.ID, referrer, refereeID, code, insertStatus, flags, nullable(st.ReferrerPrizeID), nullable(st.RefereePrizeID), goal, st.GoalMeasure, formatAmount(st.GoalValue), ip); err != nil {
			if _, unique := isUniqueViolation(err); unique {
				return conflict("ALREADY_REFERRED", "you have already entered a code")
			}
			return err
		}
		if status == "completed" {
			if out.Claims, err = s.completeReferral(ctx, tx, out.ID); err != nil {
				return err
			}
		}
		out.ReferrerID, out.Status, out.Flags = referrer, status, flags
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.settleReferralClaims(ctx, out.Claims)
	return out, nil
}

func (s *service) settleReferralClaims(ctx context.Context, claims []claimRef) {
	for i := range claims {
		if _, err := s.grantClaim(ctx, claims[i].ClaimID, 1); err != nil {
			claims[i].Status = "failed"
			s.logger.Warn("Economy referral reward %s not granted yet: %v", claims[i].ClaimID, err)
			continue
		}
		claims[i].Status = "granted"
	}
}

func (s *service) applyReferralGoal(ctx context.Context, tx *sql.Tx, userID, typeID string, out *eventOutcome) error {
	var id string
	err := tx.QueryRowContext(ctx, `SELECT r.id::text FROM economy_referral r JOIN economy_player_event_total t ON t.user_id=r.referee_id AND t.type_id=r.goal_event_type_id
		WHERE r.referee_id=$1 AND r.status='pending' AND r.goal_event_type_id=$2
		AND (CASE r.goal_measure WHEN 'sum' THEN t.sum WHEN 'max' THEN COALESCE(t.max, 0) ELSE t.count END) >= r.goal_value FOR UPDATE OF r`, userID, typeID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	claims, err := s.completeReferral(ctx, tx, id)
	if err != nil {
		return err
	}
	out.Claims = append(out.Claims, claims...)
	out.referrals = append(out.referrals, id)
	return nil
}

func (s *service) noteReferralDone(ctx context.Context, ids []string) {
	for _, id := range ids {
		var referrer, referee string
		if err := s.db.QueryRowContext(ctx, `SELECT referrer_id::text, referee_id::text FROM economy_referral WHERE id=$1`, id).Scan(&referrer, &referee); err != nil {
			continue
		}
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: referrer, Kind: activity.KindReferral, Action: "completed", RefID: id, Actor: "system", Data: map[string]any{"side": "referrer", "other": referee}})
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: referee, Kind: activity.KindReferral, Action: "completed", RefID: id, Actor: "system", Data: map[string]any{"side": "referee", "other": referrer}})
	}
}

func (s *service) referralRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_referral_me":    s.rpcReferralMe,
		"economy_referral_enter": s.rpcReferralEnter,
	}
}

func (s *service) referralView(ctx context.Context, userID, lang string) (map[string]any, error) {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, lang)
	if err != nil {
		return nil, err
	}
	st, err := loadReferralSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	code, err := s.referralCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	var created time.Time
	if err := s.db.QueryRowContext(ctx, `SELECT create_time FROM users WHERE id=$1`, userID).Scan(&created); err != nil {
		return nil, err
	}
	out := map[string]any{"lang": lang, "code": code, "enabled": st.IsEnabled, "referredBy": nil, "newAccountDays": st.NewAccountDays}
	var rid, rname, rstatus string
	var rat time.Time
	err = s.db.QueryRowContext(ctx, `SELECT r.id::text, u.username, r.status, r.created_at FROM economy_referral r JOIN users u ON u.id=r.referrer_id WHERE r.referee_id=$1`, userID).Scan(&rid, &rname, &rstatus, &rat)
	if err == nil {
		out["referredBy"] = map[string]any{"id": rid, "username": rname, "status": rstatus, "createdAt": rat.UTC().Format(time.RFC3339)}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	canEnter := st.IsEnabled && out["referredBy"] == nil
	if st.NewAccountDays > 0 {
		until := created.Add(time.Duration(st.NewAccountDays) * 24 * time.Hour)
		out["enterUntil"] = until.UTC().Format(time.RFC3339)
		if s.now().After(until) {
			canEnter = false
		}
	}
	out["canEnter"] = canEnter
	var total, completed, pending, held int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE status='completed'), count(*) FILTER (WHERE status='pending'), count(*) FILTER (WHERE status='held') FROM economy_referral WHERE referrer_id=$1 AND status <> 'rejected'`, userID).Scan(&total, &completed, &pending, &held); err != nil {
		return nil, err
	}
	out["counts"] = map[string]int{"total": total, "completed": completed, "pending": pending + held}
	rows, err := s.db.QueryContext(ctx, `SELECT u.username, r.status, r.created_at FROM economy_referral r JOIN users u ON u.id=r.referee_id WHERE r.referrer_id=$1 AND r.status <> 'rejected' ORDER BY r.created_at DESC LIMIT 10`, userID)
	if err != nil {
		return nil, err
	}
	recent := []map[string]any{}
	for rows.Next() {
		var name, status string
		var at time.Time
		if rows.Scan(&name, &status, &at) == nil {
			if status == "held" {
				status = "pending"
			}
			recent = append(recent, map[string]any{"username": name, "status": status, "createdAt": at.UTC().Format(time.RFC3339)})
		}
	}
	rows.Close()
	out["recent"] = recent
	ids := []string{}
	for _, id := range []string{st.ReferrerPrizeID, st.RefereePrizeID} {
		if id != "" {
			ids = append(ids, id)
		}
	}
	names, err := s.prizeNames(ctx, langs, lang, ids)
	if err != nil {
		return nil, err
	}
	rewards := map[string]any{"referrer": nil, "referee": nil}
	if p, ok := names[st.ReferrerPrizeID]; ok {
		rewards["referrer"] = prizeJSON(p, st.ReferrerPrizeID)
	}
	if p, ok := names[st.RefereePrizeID]; ok {
		rewards["referee"] = prizeJSON(p, st.RefereePrizeID)
	}
	out["rewards"] = rewards
	out["goal"] = nil
	if st.GoalEventTypeID != "" {
		var key, name string
		if err := s.db.QueryRowContext(ctx, `SELECT key, name::text FROM economy_event_type WHERE id=$1`, st.GoalEventTypeID).Scan(&key, &name); err == nil {
			out["goal"] = map[string]any{"eventKey": key, "eventName": langs.Resolve(localized.Parse(name), lang), "measure": st.GoalMeasure, "value": st.GoalValue}
		}
	}
	return out, nil
}

func (s *service) rpcReferralMe(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	return s.referralView(ctx, userID, in.Lang)
}

func (s *service) rpcReferralEnter(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Code string `json:"code"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	d, err := allowReferral(ctx, s, userID)
	if err != nil {
		return nil, err
	}
	if !d {
		return nil, fail(http.StatusTooManyRequests, "TOO_MANY_TRIES", "", "too many tries, wait a little")
	}
	ip := activity.ClientIP(ctx)
	res, err := s.enterReferral(ctx, userID, in.Code, ip)
	if err != nil {
		return nil, err
	}
	var refName string
	_ = s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id=$1`, res.ReferrerID).Scan(&refName)
	activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindReferral, Action: "entered", RefID: res.ID, Actor: "player", Data: map[string]any{"referrer": refName, "status": res.Status}}))
	var meName string
	_ = s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id=$1`, userID).Scan(&meName)
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: res.ReferrerID, Kind: activity.KindReferral, Action: "joined", RefID: res.ID, Actor: "player", Data: map[string]any{"referee": meName, "status": res.Status}})
	if res.Status == "completed" {
		s.noteReferralDone(ctx, []string{res.ID})
	}
	view, err := s.referralView(ctx, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	view["entered"] = map[string]any{"id": res.ID, "status": res.Status, "claims": res.Claims}
	return view, nil
}

func allowReferral(ctx context.Context, s *service, userID string) (bool, error) {
	d, err := clusterkit.Allow(ctx, s.db, "ref:"+userID, 10, time.Hour)
	return d.Allowed, err
}
