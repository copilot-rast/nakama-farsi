package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/activity"
)

const referralJSON = `jsonb_build_object('id',r.id,'referrerId',r.referrer_id,'referrerName',ru.username,'refereeId',r.referee_id,'refereeName',eu.username,'code',r.code,'status',r.status,
	'flags',to_jsonb(r.flags),'goalEventTypeId',r.goal_event_type_id,'goalEventKey',(SELECT et.key FROM economy_event_type et WHERE et.id=r.goal_event_type_id),'goalMeasure',r.goal_measure,'goalValue',r.goal_value,
	'goalProgress',(SELECT CASE r.goal_measure WHEN 'sum' THEN t.sum WHEN 'max' THEN COALESCE(t.max,0) ELSE t.count END FROM economy_player_event_total t WHERE t.user_id=r.referee_id AND t.type_id=r.goal_event_type_id),
	'ip',r.ip,'note',r.note,'decidedBy',r.decided_by,'decidedAt',r.decided_at,'createdAt',r.created_at,'completedAt',r.completed_at,'refereeCreatedAt',eu.create_time,
	'claims',(SELECT COALESCE(jsonb_agg(jsonb_build_object('id',c.id,'scope',c.scope,'userId',c.user_id,'status',c.status,'prizeId',c.prize_id,'prizeKey',(SELECT p.key FROM economy_prize p WHERE p.id=c.prize_id)) ORDER BY c.scope),'[]'::jsonb) FROM economy_reward_claim c WHERE c.kind='referral' AND c.ref_id=r.id))`

const referralFrom = ` FROM economy_referral r JOIN users ru ON ru.id=r.referrer_id JOIN users eu ON eu.id=r.referee_id`

func consoleName(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Nakama-Console-Username")); v != "" {
		return v
	}
	return "admin"
}

func (s *service) referralConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/referrals", http.MethodGet, s.consoleReferralList},
		{consolePrefix + "/referrals/summary", http.MethodGet, s.consoleReferralSummary},
		{consolePrefix + "/referrals/tree/{userId}", http.MethodGet, s.consoleReferralTree},
		{consolePrefix + "/referrals/{id}", http.MethodGet, s.consoleReferralGet},
		{consolePrefix + "/referrals/{id}/approve", http.MethodPost, s.consoleReferralDecide(true)},
		{consolePrefix + "/referrals/{id}/reject", http.MethodPost, s.consoleReferralDecide(false)},
		{consolePrefix + "/referral-settings", http.MethodGet, s.consoleReferralSettingsGet},
		{consolePrefix + "/referral-settings", http.MethodPut, s.consoleReferralSettingsPut},
	}
}

func (s *service) consoleReferralList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	q := r.URL.Query()
	where := &sqlWhere{}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(ru.username ILIKE "+n+" OR eu.username ILIKE "+n+" OR r.code ILIKE "+n+" OR r.referrer_id::text ILIKE "+n+" OR r.referee_id::text ILIKE "+n+")")
	}
	if v := q.Get("status"); v == "pending" || v == "held" || v == "completed" || v == "rejected" {
		where.add("r.status = ?", v)
	}
	if q.Get("flagged") == "true" {
		where.add("cardinality(r.flags) > 0")
	}
	if v := q.Get("flag"); v != "" {
		where.add("? = ANY(r.flags)", v)
	}
	for _, f := range []string{"referrerId", "refereeId"} {
		if v := q.Get(f); v != "" {
			if _, err := uuid.FromString(v); err == nil {
				col := "r.referrer_id"
				if f == "refereeId" {
					col = "r.referee_id"
				}
				where.add(col+" = ?::uuid", v)
			}
		}
	}
	s.listRaw(w, r, `SELECT count(*)`+referralFrom, `SELECT `+referralJSON+`::text`+referralFrom, where, ` ORDER BY r.created_at DESC`, p)
}

func (s *service) consoleReferralGet(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("REFERRAL_NOT_FOUND"))
		return
	}
	var raw string
	err := s.db.QueryRowContext(r.Context(), `SELECT `+referralJSON+`::text`+referralFrom+` WHERE r.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, notFound("REFERRAL_NOT_FOUND"))
		return
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, []byte(raw))
}

func (s *service) consoleReferralSummary(w http.ResponseWriter, r *http.Request) {
	var total, completed, pending, held, rejected, flagged, today int
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*), count(*) FILTER (WHERE status='completed'), count(*) FILTER (WHERE status='pending'), count(*) FILTER (WHERE status='held'),
		count(*) FILTER (WHERE status='rejected'), count(*) FILTER (WHERE cardinality(flags) > 0), count(*) FILTER (WHERE created_at > now() - INTERVAL '1 day') FROM economy_referral`).
		Scan(&total, &completed, &pending, &held, &rejected, &flagged, &today); err != nil {
		s.writeError(w, err)
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT r.referrer_id::text, u.username, count(*), count(*) FILTER (WHERE r.status='completed'), count(*) FILTER (WHERE cardinality(r.flags) > 0)
		FROM economy_referral r JOIN users u ON u.id=r.referrer_id WHERE r.status <> 'rejected' GROUP BY 1, 2 ORDER BY 3 DESC, 2 LIMIT 10`)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	top := []map[string]any{}
	for rows.Next() {
		var id, name string
		var n, done, fl int
		if rows.Scan(&id, &name, &n, &done, &fl) == nil {
			top = append(top, map[string]any{"userId": id, "username": name, "total": n, "completed": done, "flagged": fl})
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"total": total, "completed": completed, "pending": pending, "held": held, "rejected": rejected, "flagged": flagged, "today": today, "topReferrers": top})
}

func (s *service) consoleReferralTree(w http.ResponseWriter, r *http.Request) {
	userID := pathVar(r, "userId")
	if _, err := uuid.FromString(userID); err != nil {
		s.writeError(w, notFound("USER_NOT_FOUND"))
		return
	}
	var name string
	if err := s.db.QueryRowContext(r.Context(), `SELECT username FROM users WHERE id=$1`, userID).Scan(&name); err != nil {
		s.writeError(w, notFound("USER_NOT_FOUND"))
		return
	}
	var code sql.NullString
	_ = s.db.QueryRowContext(r.Context(), `SELECT code FROM economy_referral_code WHERE user_id=$1`, userID).Scan(&code)
	up := []map[string]any{}
	cur := userID
	seen := map[string]bool{userID: true}
	for i := 0; i < 5; i++ {
		var rid, pid, pname, status string
		err := s.db.QueryRowContext(r.Context(), `SELECT r.id::text, r.referrer_id::text, u.username, r.status FROM economy_referral r JOIN users u ON u.id=r.referrer_id WHERE r.referee_id=$1`, cur).Scan(&rid, &pid, &pname, &status)
		if err != nil || seen[pid] {
			break
		}
		seen[pid] = true
		up = append(up, map[string]any{"referralId": rid, "userId": pid, "username": pname, "status": status})
		cur = pid
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT r.id::text, r.referee_id::text, u.username, r.status, to_jsonb(r.flags)::text, r.created_at,
		(SELECT count(*) FROM economy_referral c WHERE c.referrer_id=r.referee_id AND c.status <> 'rejected')
		FROM economy_referral r JOIN users u ON u.id=r.referee_id WHERE r.referrer_id=$1 ORDER BY r.created_at DESC LIMIT 200`, userID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	children := []map[string]any{}
	for rows.Next() {
		var rid, cid, cname, status, flags string
		var at time.Time
		var n int
		if err := rows.Scan(&rid, &cid, &cname, &status, &flags, &at, &n); err != nil {
			s.writeError(w, err)
			return
		}
		children = append(children, map[string]any{"referralId": rid, "userId": cid, "username": cname, "status": status, "flags": json.RawMessage(flags), "createdAt": at.UTC().Format(time.RFC3339), "referrals": n})
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": map[string]any{"id": userID, "username": name, "code": code.String}, "referredBy": up, "children": children})
}

func (s *service) consoleReferralDecide(approve bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Note string `json:"note"`
		}
		if err := readBody(r, &in); err != nil {
			s.writeError(w, err)
			return
		}
		id := pathVar(r, "id")
		if _, err := uuid.FromString(id); err != nil {
			s.writeError(w, notFound("REFERRAL_NOT_FOUND"))
			return
		}
		note := strings.TrimSpace(in.Note)
		if len([]rune(note)) > 500 {
			s.writeError(w, badRequest("TEXT_TOO_LONG", "note", "at most 500 letters"))
			return
		}
		me := consoleName(r)
		var claims []claimRef
		var completed bool
		var referrer, referee string
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			var status string
			var goal sql.NullString
			err := tx.QueryRowContext(r.Context(), `SELECT status, goal_event_type_id::text, referrer_id::text, referee_id::text FROM economy_referral WHERE id=$1 FOR UPDATE`, id).Scan(&status, &goal, &referrer, &referee)
			if errors.Is(err, sql.ErrNoRows) {
				return notFound("REFERRAL_NOT_FOUND")
			}
			if err != nil {
				return err
			}
			if status == "completed" || status == "rejected" {
				return conflict("REFERRAL_DECIDED", "this referral is already "+status)
			}
			if !approve {
				_, err := tx.ExecContext(r.Context(), `UPDATE economy_referral SET status='rejected', note=$2, decided_by=$3, decided_at=now() WHERE id=$1`, id, note, me)
				return err
			}
			if status == "pending" {
				return conflict("NOT_HELD", "only a held referral needs approval")
			}
			if _, err := tx.ExecContext(r.Context(), `UPDATE economy_referral SET note=$2, decided_by=$3, decided_at=now() WHERE id=$1`, id, note, me); err != nil {
				return err
			}
			if goal.Valid {
				reached := false
				_ = tx.QueryRowContext(r.Context(), `SELECT (CASE r.goal_measure WHEN 'sum' THEN t.sum WHEN 'max' THEN COALESCE(t.max,0) ELSE t.count END) >= r.goal_value
					FROM economy_referral r JOIN economy_player_event_total t ON t.user_id=r.referee_id AND t.type_id=r.goal_event_type_id WHERE r.id=$1`, id).Scan(&reached)
				if !reached {
					_, err := tx.ExecContext(r.Context(), `UPDATE economy_referral SET status='pending' WHERE id=$1`, id)
					return err
				}
			}
			claims, err = s.completeReferral(r.Context(), tx, id)
			completed = true
			return err
		})
		if err != nil {
			s.writeError(w, err)
			return
		}
		s.settleReferralClaims(r.Context(), claims)
		action := "rejected"
		if approve {
			action = "approved"
		}
		activity.Safe(r.Context(), s.db, s.logger, activity.Entry{UserID: referee, Kind: activity.KindReferral, Action: action, RefID: id, Actor: "console:" + me, Data: map[string]any{"note": note}})
		if completed {
			s.noteReferralDone(r.Context(), []string{id})
		}
		s.consoleReferralGet(w, r)
	}
}

func (s *service) consoleReferralSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := loadReferralSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *service) consoleReferralSettingsPut(w http.ResponseWriter, r *http.Request) {
	var in referralSettings
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if in.NewAccountDays < 0 || in.NewAccountDays > 3650 {
		s.writeError(w, badRequest("INVALID_VALUE", "newAccountDays", "days are 0 to 3650"))
		return
	}
	if in.MaxPerReferrer < 1 || in.MaxPerReferrer > 100000 {
		s.writeError(w, badRequest("INVALID_VALUE", "maxPerReferrer", "1 to 100000"))
		return
	}
	if in.DailyPerReferer < 1 || in.DailyPerReferer > 100000 {
		s.writeError(w, badRequest("INVALID_VALUE", "dailyPerReferrer", "1 to 100000"))
		return
	}
	if in.GoalMeasure == "" {
		in.GoalMeasure = "count"
	}
	if in.GoalMeasure != "count" && in.GoalMeasure != "sum" && in.GoalMeasure != "max" {
		s.writeError(w, badRequest("INVALID_MEASURE", "goalMeasure", "count, sum or max"))
		return
	}
	if in.GoalValue <= 0 || math.IsNaN(in.GoalValue) || in.GoalValue >= 1e15 {
		if in.GoalEventTypeID != "" {
			s.writeError(w, badRequest("INVALID_VALUE", "goalValue", "goal must be above zero"))
			return
		}
		in.GoalValue = 1
	}
	refs := []struct {
		field, table, code string
		val                *string
	}{{"referrerPrizeId", "economy_prize", "PRIZE_NOT_FOUND", &in.ReferrerPrizeID}, {"refereePrizeId", "economy_prize", "PRIZE_NOT_FOUND", &in.RefereePrizeID}, {"goalEventTypeId", "economy_event_type", "EVENT_TYPE_NOT_FOUND", &in.GoalEventTypeID}}
	args := []any{}
	for _, ref := range refs {
		v := strings.TrimSpace(*ref.val)
		*ref.val = v
		if v == "" {
			args = append(args, nil)
			continue
		}
		var ok bool
		if _, err := uuid.FromString(v); err == nil {
			_ = s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM `+ref.table+` WHERE id=$1)`, v).Scan(&ok)
		}
		if !ok {
			s.writeError(w, badRequest(ref.code, ref.field, "does not exist"))
			return
		}
		args = append(args, v)
	}
	_, err := s.db.ExecContext(r.Context(), `INSERT INTO economy_referral_settings (id, is_enabled, new_account_days, referrer_prize_id, referee_prize_id, goal_event_type_id, goal_measure, goal_value, max_per_referrer, daily_per_referrer, one_per_device, one_per_phone, hold_flagged, updated_at)
		VALUES (1,$1,$2,$3,$4,$5,$6,$7::numeric,$8,$9,$10,$11,$12,now())
		ON CONFLICT (id) DO UPDATE SET is_enabled=EXCLUDED.is_enabled, new_account_days=EXCLUDED.new_account_days, referrer_prize_id=EXCLUDED.referrer_prize_id, referee_prize_id=EXCLUDED.referee_prize_id,
		goal_event_type_id=EXCLUDED.goal_event_type_id, goal_measure=EXCLUDED.goal_measure, goal_value=EXCLUDED.goal_value, max_per_referrer=EXCLUDED.max_per_referrer, daily_per_referrer=EXCLUDED.daily_per_referrer,
		one_per_device=EXCLUDED.one_per_device, one_per_phone=EXCLUDED.one_per_phone, hold_flagged=EXCLUDED.hold_flagged, updated_at=now()`,
		in.IsEnabled, in.NewAccountDays, args[0], args[1], args[2], in.GoalMeasure, formatAmount(in.GoalValue), in.MaxPerReferrer, in.DailyPerReferer, in.OnePerDevice, in.OnePerPhone, in.HoldFlagged)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleReferralSettingsGet(w, r)
}
