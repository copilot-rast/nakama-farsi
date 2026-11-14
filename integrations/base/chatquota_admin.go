package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

func (s *service) chatQuotaSettingsBody(ctx context.Context) (map[string]any, error) {
	st, err := loadChatQuotaSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(st)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	products, err := s.chatPackProducts(ctx, st, nil, "en", false)
	if err != nil {
		return nil, err
	}
	out["detectedProducts"] = products
	return out, nil
}

func (s *service) chatQuotaSettingsGet(w http.ResponseWriter, r *http.Request) {
	out, err := s.chatQuotaSettingsBody(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) checkKeysExist(ctx context.Context, table, field string, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM `+table+` WHERE key = ANY($1::text[])`, keys).Scan(&n); err != nil {
		return err
	}
	if n != len(keys) {
		code := "ITEM_NOT_FOUND"
		if table == "economy_product" {
			code = "PRODUCT_NOT_FOUND"
		}
		return badRequest(code, field, "a key does not exist")
	}
	return nil
}

func (s *service) chatQuotaSettingsPut(w http.ResponseWriter, r *http.Request) {
	in := defaultChatQuotaSettings()
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	st, err := in.clean()
	ctx := r.Context()
	if err == nil {
		err = s.checkMetaKey(ctx, "packMeta", st.PackMeta)
	}
	if err == nil {
		err = s.checkMetaKey(ctx, "bonusMeta", st.BonusMeta)
	}
	if err == nil {
		err = s.checkMetaKey(ctx, "unlimitedMeta", st.UnlimitedMeta)
	}
	if err == nil && st.PackMeta != "" && (st.PackMeta == st.BonusMeta || st.PackMeta == st.UnlimitedMeta) {
		err = badRequest("SAME_META", "packMeta", "the pack meta is spent as packs; use another meta for the bonus or unlimited")
	}
	if err == nil {
		err = s.checkKeysExist(ctx, "economy_item", "unlimitedItems", st.UnlimitedItems)
	}
	if err == nil {
		err = s.checkKeysExist(ctx, "economy_product", "packProducts", st.PackProducts)
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	st.UpdatedBy, st.UpdatedAt = "", time.Time{}
	raw, _ := json.Marshal(st)
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	delete(m, "updatedAt")
	delete(m, "updatedBy")
	raw, _ = json.Marshal(m)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO base_chat_quota_settings (id, settings, updated_by, updated_at) VALUES (1, $1::jsonb, $2, now())
		ON CONFLICT (id) DO UPDATE SET settings=EXCLUDED.settings, updated_by=EXCLUDED.updated_by, updated_at=now()`, string(raw), consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	s.chatQuotaSettingsGet(w, r)
}

const chatRuleJSON = `jsonb_build_object('id',r.id,'name',r.name,'tagIds',to_jsonb(r.tag_ids),'priority',r.priority,'unlimited',r.unlimited,
	'dailyMessages',r.daily_messages,'dailyConversations',r.daily_conversations,'friendDailyMessages',r.friend_daily_messages,'friendDailyConversations',r.friend_daily_conversations,
	'isActive',r.is_active,'createdBy',r.created_by,'createdAt',r.created_at,'updatedAt',r.updated_at,
	'tags',(SELECT COALESCE(jsonb_agg(jsonb_build_object('id',g.id,'key',g.key,'name',g.name) ORDER BY g.key),'[]'::jsonb) FROM base_tag g WHERE g.id::text = ANY(r.tag_ids)),
	'players',(SELECT count(DISTINCT ut.user_id) FROM base_user_tag ut WHERE ut.tag_id::text = ANY(r.tag_ids)))`

func (s *service) chatRuleList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		where.add("r.name ILIKE ?", likePattern(p.Q))
	}
	if v := r.URL.Query().Get("isActive"); v == "true" || v == "false" {
		where.add("r.is_active = ?", v == "true")
	}
	out, err := s.listRaw(r, ` FROM base_chat_quota_rule r`, chatRuleJSON, "r.priority DESC, r.created_at, r.id", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func fetchChatRule(ctx context.Context, q queryer, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+chatRuleJSON+`::text FROM base_chat_quota_rule r WHERE r.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func (s *service) chatRuleGet(w http.ResponseWriter, r *http.Request) {
	raw, err := fetchChatRule(r.Context(), s.db, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

func (s *service) chatRuleWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Name                     string   `json:"name"`
			TagIDs                   []string `json:"tagIds"`
			Priority                 int      `json:"priority"`
			Unlimited                bool     `json:"unlimited"`
			DailyMessages            *int     `json:"dailyMessages"`
			DailyConversations       *int     `json:"dailyConversations"`
			FriendDailyMessages      *int     `json:"friendDailyMessages"`
			FriendDailyConversations *int     `json:"friendDailyConversations"`
			IsActive                 *bool    `json:"isActive"`
		}
		if err := readBody(r, &in); err != nil {
			s.writeError(w, err)
			return
		}
		id := pathVar(r, "id")
		if create {
			id = uuid.Must(uuid.NewV4()).String()
		} else if _, err := uuid.FromString(id); err != nil {
			s.writeError(w, notFound("NOT_FOUND"))
			return
		}
		name := strings.TrimSpace(in.Name)
		if n := len([]rune(name)); n == 0 || n > 100 {
			s.writeError(w, badRequest("INVALID_NAME", "name", "1 to 100 letters"))
			return
		}
		if in.Priority < -1000 || in.Priority > 1000 {
			s.writeError(w, badRequest("INVALID_RULE", "priority", "between -1000 and 1000"))
			return
		}
		vals := make([]any, 4)
		for i, f := range []struct {
			name string
			v    *int
		}{{"dailyMessages", in.DailyMessages}, {"dailyConversations", in.DailyConversations}, {"friendDailyMessages", in.FriendDailyMessages}, {"friendDailyConversations", in.FriendDailyConversations}} {
			v, err := optLimit(f.name, f.v, chatQuotaMaxLimit)
			if err != nil {
				s.writeError(w, err)
				return
			}
			vals[i] = v
		}
		if !in.Unlimited && vals[0] == nil && vals[1] == nil && vals[2] == nil && vals[3] == nil {
			s.writeError(w, badRequest("LIMIT_REQUIRED", "dailyMessages", "set a limit or make the rule unlimited"))
			return
		}
		active := in.IsActive == nil || *in.IsActive
		var out json.RawMessage
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			tags, err := s.userTagIDs(r.Context(), tx, in.TagIDs)
			if err != nil {
				return err
			}
			if create {
				_, err = tx.ExecContext(r.Context(), `INSERT INTO base_chat_quota_rule (id, name, tag_ids, priority, unlimited, daily_messages, daily_conversations, friend_daily_messages, friend_daily_conversations, is_active, created_by)
					VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, id, name, tags, in.Priority, in.Unlimited, vals[0], vals[1], vals[2], vals[3], active, consoleUser(r))
			} else {
				var res sql.Result
				res, err = tx.ExecContext(r.Context(), `UPDATE base_chat_quota_rule SET name=$2, tag_ids=$3, priority=$4, unlimited=$5, daily_messages=$6, daily_conversations=$7, friend_daily_messages=$8, friend_daily_conversations=$9, is_active=$10, updated_at=now() WHERE id=$1`,
					id, name, tags, in.Priority, in.Unlimited, vals[0], vals[1], vals[2], vals[3], active)
				if err == nil {
					if n, _ := res.RowsAffected(); n == 0 {
						return notFound("NOT_FOUND")
					}
				}
			}
			if err != nil {
				return err
			}
			out, err = fetchChatRule(r.Context(), tx, id)
			return err
		})
		if err != nil {
			s.writeError(w, err)
			return
		}
		status := http.StatusOK
		if create {
			status = http.StatusCreated
		}
		writeRaw(w, status, out)
	}
}

func (s *service) chatRuleDelete(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM base_chat_quota_rule WHERE id=$1`, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) playerExists(ctx context.Context, id string) error {
	if _, err := uuid.FromString(id); err != nil {
		return notFound("PLAYER_NOT_FOUND")
	}
	var ok bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1::uuid)`, id).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return notFound("PLAYER_NOT_FOUND")
	}
	return nil
}

func (s *service) playerChatQuota(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if err := s.playerExists(r.Context(), id); err != nil {
		s.writeError(w, err)
		return
	}
	out, err := s.chatQuotaStatus(r.Context(), id, r.URL.Query().Get("lang"), 20)
	if err != nil {
		s.writeError(w, err)
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT day::text, total, messages + friend_messages, conversations + friend_conversations, pack_used, free_sent, unlimited_sent, blocked
		FROM base_chat_quota_day WHERE user_id=$1::uuid ORDER BY day DESC LIMIT 14`, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	days := []map[string]any{}
	for rows.Next() {
		var day string
		var total, msgs, convs, packs, free, unl, blocked int
		if err := rows.Scan(&day, &total, &msgs, &convs, &packs, &free, &unl, &blocked); err != nil {
			s.writeError(w, err)
			return
		}
		days = append(days, map[string]any{"day": day, "sent": total, "messages": msgs, "conversations": convs, "fromPacks": packs, "free": free, "unlimited": unl, "blocked": blocked})
	}
	if err := rows.Err(); err != nil {
		s.writeError(w, err)
		return
	}
	out["days"] = days
	writeJSON(w, http.StatusOK, out)
}

func (s *service) playerChatQuotaGrant(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var in struct {
		Amount        int    `json:"amount"`
		ExpiresInDays *int   `json:"expiresInDays"`
		Note          string `json:"note"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if err := s.playerExists(r.Context(), id); err != nil {
		s.writeError(w, err)
		return
	}
	if in.Amount < 1 || in.Amount > chatQuotaMaxLimit {
		s.writeError(w, badRequest("INVALID_AMOUNT", "amount", "between 1 and 100000"))
		return
	}
	note := strings.TrimSpace(in.Note)
	if len([]rune(note)) > 200 {
		s.writeError(w, badRequest("TEXT_TOO_LONG", "note", "at most 200 letters"))
		return
	}
	days := 0
	if in.ExpiresInDays == nil {
		st, err := loadChatQuotaSettings(r.Context(), s.db)
		if err != nil {
			s.writeError(w, err)
			return
		}
		days = st.PackExpiryDays
	} else {
		days = *in.ExpiresInDays
	}
	if days < 0 || days > 3650 {
		s.writeError(w, badRequest("INVALID_SETTING", "expiresInDays", "between 0 and 3650"))
		return
	}
	var exp any
	if days > 0 {
		exp = s.now().Add(time.Duration(days) * 24 * time.Hour)
	}
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO base_chat_quota_lot (id, user_id, amount, remaining, source, note, expires_at, created_by) VALUES ($1,$2,$3,$3,'admin',$4,$5,$6)`,
		uuid.Must(uuid.NewV4()).String(), id, in.Amount, note, exp, consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	s.playerChatQuota(w, r)
}

func (s *service) playerChatQuotaReset(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if err := s.playerExists(r.Context(), id); err != nil {
		s.writeError(w, err)
		return
	}
	st, err := loadChatQuotaSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	day, _, _ := s.chatPeriod(st)
	if _, err := s.db.ExecContext(r.Context(), `UPDATE base_chat_quota_day SET messages=0, conversations=0, friend_messages=0, friend_conversations=0, notified=false, updated_at=now() WHERE user_id=$1::uuid AND day=$2::date`, id, day); err != nil {
		s.writeError(w, err)
		return
	}
	s.playerChatQuota(w, r)
}

func (s *service) playerChatQuotaRevoke(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	lot := pathVar(r, "lot")
	if _, err := uuid.FromString(lot); err != nil {
		s.writeError(w, notFound("LOT_NOT_FOUND"))
		return
	}
	if err := s.playerExists(r.Context(), id); err != nil {
		s.writeError(w, err)
		return
	}
	res, err := s.db.ExecContext(r.Context(), `UPDATE base_chat_quota_lot SET remaining=0, updated_at=now() WHERE id=$1::uuid AND user_id=$2::uuid`, lot, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("LOT_NOT_FOUND"))
		return
	}
	s.playerChatQuota(w, r)
}

func (s *service) chatQuotaReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days < 1 || days > 365 {
		days = 30
	}
	st, err := loadChatQuotaSettings(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	_, start, _ := s.chatPeriod(st)
	sinceT := start.AddDate(0, 0, -(days - 1))
	since := sinceT.Format("2006-01-02")
	var t struct{ sent, blocked, packs, free, unl, daily, senders, blockedPlayers int }
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(sum(total),0)::int, COALESCE(sum(blocked),0)::int, COALESCE(sum(pack_used),0)::int, COALESCE(sum(free_sent),0)::int, COALESCE(sum(unlimited_sent),0)::int,
			COALESCE(sum(messages + friend_messages),0)::int, count(DISTINCT user_id) FILTER (WHERE total > 0), count(DISTINCT user_id) FILTER (WHERE blocked > 0)
		FROM base_chat_quota_day WHERE day >= $1::date`, since).Scan(&t.sent, &t.blocked, &t.packs, &t.free, &t.unl, &t.daily, &t.senders, &t.blockedPlayers); err != nil {
		s.writeError(w, err)
		return
	}
	daily := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT day::text, sum(total)::int, sum(blocked)::int, sum(pack_used)::int FROM base_chat_quota_day WHERE day >= $1::date GROUP BY day ORDER BY day`, since)
	if err != nil {
		s.writeError(w, err)
		return
	}
	for rows.Next() {
		var day string
		var sent, blocked, packs int
		if err := rows.Scan(&day, &sent, &blocked, &packs); err != nil {
			rows.Close()
			s.writeError(w, err)
			return
		}
		daily = append(daily, map[string]any{"day": day, "sent": sent, "blocked": blocked, "fromPacks": packs})
	}
	rows.Close()
	top := []map[string]any{}
	rows, err = s.db.QueryContext(ctx, `SELECT d.user_id::text, u.username, COALESCE(u.display_name,''), sum(d.blocked)::int FROM base_chat_quota_day d JOIN users u ON u.id=d.user_id
		WHERE d.day >= $1::date AND d.blocked > 0 GROUP BY d.user_id, u.username, u.display_name ORDER BY sum(d.blocked) DESC, u.username LIMIT 10`, since)
	if err != nil {
		s.writeError(w, err)
		return
	}
	for rows.Next() {
		var id, username, display string
		var n int
		if err := rows.Scan(&id, &username, &display, &n); err != nil {
			rows.Close()
			s.writeError(w, err)
			return
		}
		top = append(top, map[string]any{"userId": id, "username": username, "displayName": display, "count": n})
	}
	rows.Close()
	products, err := s.chatPackProducts(ctx, st, nil, "en", false)
	if err != nil {
		s.writeError(w, err)
		return
	}
	ids := make([]string, 0, len(products))
	for _, p := range products {
		ids = append(ids, p.ID)
	}
	sales := []map[string]any{}
	var salesCount, salesToman int64
	coins := map[string]int64{}
	if len(ids) > 0 {
		rows, err = s.db.QueryContext(ctx, `SELECT p.id::text, p.key, p.name::text, count(*)::bigint, COALESCE(sum(pay.final_toman) FILTER (WHERE pay.coin_currency IS NULL),0)::bigint, COALESCE(pay.coin_currency,''), COALESCE(sum(pay.coin_amount),0)::bigint
			FROM economy_payment pay JOIN economy_product p ON p.id=pay.product_id
			WHERE pay.status='success' AND pay.product_id = ANY($1::uuid[]) AND pay.created_at >= $2
			GROUP BY p.id, p.key, p.name, pay.coin_currency ORDER BY count(*) DESC`, ids, sinceT)
		if err != nil {
			s.writeError(w, err)
			return
		}
		byID := map[string]map[string]any{}
		for rows.Next() {
			var id, key, name, cur string
			var n, toman, coin int64
			if err := rows.Scan(&id, &key, &name, &n, &toman, &cur, &coin); err != nil {
				rows.Close()
				s.writeError(w, err)
				return
			}
			row, ok := byID[id]
			if !ok {
				row = map[string]any{"productId": id, "key": key, "name": json.RawMessage(name), "count": int64(0), "toman": int64(0), "coins": map[string]int64{}}
				byID[id] = row
				sales = append(sales, row)
			}
			row["count"] = row["count"].(int64) + n
			row["toman"] = row["toman"].(int64) + toman
			if cur != "" {
				row["coins"].(map[string]int64)[cur] += coin
				coins[cur] += coin
			}
			salesCount += n
			salesToman += toman
		}
		rows.Close()
	}
	granted := map[string]int64{"meta": 0, "admin": 0}
	rows, err = s.db.QueryContext(ctx, `SELECT source, COALESCE(sum(amount),0)::bigint FROM base_chat_quota_lot WHERE created_at >= $1 GROUP BY source`, sinceT)
	if err != nil {
		s.writeError(w, err)
		return
	}
	for rows.Next() {
		var src string
		var n int64
		if err := rows.Scan(&src, &n); err != nil {
			rows.Close()
			s.writeError(w, err)
			return
		}
		granted[src] = n
	}
	rows.Close()
	writeJSON(w, http.StatusOK, map[string]any{
		"days":   days,
		"since":  since,
		"totals": map[string]any{"sent": t.sent, "blocked": t.blocked, "fromPacks": t.packs, "free": t.free, "unlimited": t.unl, "daily": t.daily, "senders": t.senders, "blockedPlayers": t.blockedPlayers},
		"daily":  daily, "topBlocked": top,
		"packs": map[string]any{"sold": salesCount, "toman": salesToman, "coins": coins, "products": sales, "granted": granted, "packProducts": products},
	})
}
