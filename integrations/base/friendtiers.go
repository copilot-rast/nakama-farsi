package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
)

type friendLimits struct {
	DailyRequests       int
	MaxFriends          int
	RejectCooldownHours int
	BaseDaily           int
	BaseMax             int
	BonusDaily          int
	BonusMax            int
	TierID              string
	TierName            string
}

type friendTier struct {
	id       string
	name     string
	daily    sql.NullInt64
	max      sql.NullInt64
	cooldown sql.NullInt64
}

func generousLimit(cur int, v int64) int {
	if cur == 0 || v == 0 {
		return 0
	}
	return max(cur, int(v))
}

func (s *service) metaBonus(ctx context.Context, userID, key string) (int, error) {
	if key == "" {
		return 0, nil
	}
	var v sql.NullFloat64
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(um.value, m.default_value)::float8 FROM economy_meta m LEFT JOIN economy_user_meta um ON um.meta_id=m.id AND um.user_id=$1::uuid WHERE m.key=$2`, userID, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) || !v.Valid {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if v.Float64 <= 0 || math.IsNaN(v.Float64) {
		return 0, nil
	}
	return int(math.Min(math.Floor(v.Float64), 100000)), nil
}

func (s *service) effectiveFriendLimits(ctx context.Context, rules friendRules, userID string) (friendLimits, error) {
	out := friendLimits{DailyRequests: rules.DailyRequests, MaxFriends: rules.MaxFriends, RejectCooldownHours: rules.RejectCooldownHours}
	rows, err := s.db.QueryContext(ctx, `SELECT r.id::text, r.name, r.daily_requests, r.max_friends, r.reject_cooldown_hours FROM base_friend_rule r
		WHERE r.is_active AND EXISTS (SELECT 1 FROM base_user_tag ut WHERE ut.user_id=$1::uuid AND ut.tag_id::text = ANY(r.tag_ids) AND (ut.expires_at IS NULL OR ut.expires_at > now()))
		ORDER BY r.priority DESC, r.created_at, r.id`, userID)
	if err != nil {
		return out, err
	}
	tiers := []friendTier{}
	for rows.Next() {
		var t friendTier
		if err := rows.Scan(&t.id, &t.name, &t.daily, &t.max, &t.cooldown); err != nil {
			rows.Close()
			return out, err
		}
		tiers = append(tiers, t)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return out, err
	}
	if len(tiers) > 0 {
		out.TierID, out.TierName = tiers[0].id, tiers[0].name
		if rules.MatchPolicy == "generous" {
			setDaily, setMax, setCool := false, false, false
			for _, t := range tiers {
				if t.daily.Valid {
					if !setDaily {
						out.DailyRequests, setDaily = int(t.daily.Int64), true
					} else {
						out.DailyRequests = generousLimit(out.DailyRequests, t.daily.Int64)
					}
				}
				if t.max.Valid {
					if !setMax {
						out.MaxFriends, setMax = int(t.max.Int64), true
					} else {
						out.MaxFriends = generousLimit(out.MaxFriends, t.max.Int64)
					}
				}
				if t.cooldown.Valid {
					if !setCool {
						out.RejectCooldownHours, setCool = int(t.cooldown.Int64), true
					} else {
						out.RejectCooldownHours = min(out.RejectCooldownHours, int(t.cooldown.Int64))
					}
				}
			}
		} else {
			t := tiers[0]
			if t.daily.Valid {
				out.DailyRequests = int(t.daily.Int64)
			}
			if t.max.Valid {
				out.MaxFriends = int(t.max.Int64)
			}
			if t.cooldown.Valid {
				out.RejectCooldownHours = int(t.cooldown.Int64)
			}
		}
	}
	out.BaseDaily, out.BaseMax = out.DailyRequests, out.MaxFriends
	if out.DailyRequests > 0 {
		if out.BonusDaily, err = s.metaBonus(ctx, userID, rules.RequestsBonusMeta); err != nil {
			return out, err
		}
		out.DailyRequests += out.BonusDaily
	}
	if out.MaxFriends > 0 {
		if out.BonusMax, err = s.metaBonus(ctx, userID, rules.FriendsBonusMeta); err != nil {
			return out, err
		}
		out.MaxFriends += out.BonusMax
	}
	return out, nil
}

const friendTierJSON = `jsonb_build_object('id',r.id,'name',r.name,'tagIds',to_jsonb(r.tag_ids),'priority',r.priority,'dailyRequests',r.daily_requests,'maxFriends',r.max_friends,
	'rejectCooldownHours',r.reject_cooldown_hours,'isActive',r.is_active,'createdBy',r.created_by,'createdAt',r.created_at,'updatedAt',r.updated_at,
	'tags',(SELECT COALESCE(jsonb_agg(jsonb_build_object('id',g.id,'key',g.key,'name',g.name) ORDER BY g.key),'[]'::jsonb) FROM base_tag g WHERE g.id::text = ANY(r.tag_ids)),
	'players',(SELECT count(DISTINCT ut.user_id) FROM base_user_tag ut WHERE ut.tag_id::text = ANY(r.tag_ids) AND (ut.expires_at IS NULL OR ut.expires_at > now())))`

func (s *service) friendTierList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		where.add("r.name ILIKE ?", likePattern(p.Q))
	}
	if v := r.URL.Query().Get("isActive"); v == "true" || v == "false" {
		where.add("r.is_active = ?", v == "true")
	}
	out, err := s.listRaw(r, ` FROM base_friend_rule r`, friendTierJSON, "r.priority DESC, r.created_at, r.id", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func fetchFriendTier(ctx context.Context, q queryer, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+friendTierJSON+`::text FROM base_friend_rule r WHERE r.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func (s *service) friendTierGet(w http.ResponseWriter, r *http.Request) {
	raw, err := fetchFriendTier(r.Context(), s.db, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

func optLimit(field string, v *int, hi int) (any, error) {
	if v == nil {
		return nil, nil
	}
	if *v < 0 || *v > hi {
		return nil, badRequest("INVALID_RULE", field, "out of range")
	}
	return *v, nil
}

func (s *service) friendTierWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Name                string   `json:"name"`
			TagIDs              []string `json:"tagIds"`
			Priority            int      `json:"priority"`
			DailyRequests       *int     `json:"dailyRequests"`
			MaxFriends          *int     `json:"maxFriends"`
			RejectCooldownHours *int     `json:"rejectCooldownHours"`
			IsActive            *bool    `json:"isActive"`
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
		daily, err := optLimit("dailyRequests", in.DailyRequests, 1000)
		if err != nil {
			s.writeError(w, err)
			return
		}
		maxF, err := optLimit("maxFriends", in.MaxFriends, 10000)
		if err != nil {
			s.writeError(w, err)
			return
		}
		cool, err := optLimit("rejectCooldownHours", in.RejectCooldownHours, 8760)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if daily == nil && maxF == nil && cool == nil {
			s.writeError(w, badRequest("LIMIT_REQUIRED", "dailyRequests", "set at least one limit"))
			return
		}
		active := in.IsActive == nil || *in.IsActive
		var out json.RawMessage
		err = s.inTx(r.Context(), func(tx *sql.Tx) error {
			tags, err := s.userTagIDs(r.Context(), tx, in.TagIDs)
			if err != nil {
				return err
			}
			if create {
				_, err = tx.ExecContext(r.Context(), `INSERT INTO base_friend_rule (id, name, tag_ids, priority, daily_requests, max_friends, reject_cooldown_hours, is_active, created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
					id, name, tags, in.Priority, daily, maxF, cool, active, consoleUser(r))
			} else {
				var res sql.Result
				res, err = tx.ExecContext(r.Context(), `UPDATE base_friend_rule SET name=$2, tag_ids=$3, priority=$4, daily_requests=$5, max_friends=$6, reject_cooldown_hours=$7, is_active=$8, updated_at=now() WHERE id=$1`,
					id, name, tags, in.Priority, daily, maxF, cool, active)
				if err == nil {
					if n, _ := res.RowsAffected(); n == 0 {
						return notFound("NOT_FOUND")
					}
				}
			}
			if err != nil {
				return err
			}
			out, err = fetchFriendTier(r.Context(), tx, id)
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

func (s *service) friendTierDelete(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM base_friend_rule WHERE id=$1`, id)
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

func (s *service) userTagIDs(ctx context.Context, q interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}, ids []string) ([]string, error) {
	clean := []string{}
	seen := map[string]bool{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if _, err := uuid.FromString(id); err != nil {
			return nil, badRequest("TAG_NOT_FOUND", "tagIds", "not a tag id")
		}
		if !seen[id] {
			seen[id] = true
			clean = append(clean, id)
		}
	}
	if len(clean) == 0 {
		return nil, badRequest("TAGS_REQUIRED", "tagIds", "pick at least one player tag")
	}
	if len(clean) > 20 {
		return nil, badRequest("TOO_MANY_TAGS", "tagIds", "at most 20 tags")
	}
	rows, err := q.QueryContext(ctx, `SELECT id::text, key_type FROM base_tag WHERE id = ANY($1::uuid[])`, clean)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	found := 0
	for rows.Next() {
		var id, kt string
		if err := rows.Scan(&id, &kt); err != nil {
			return nil, err
		}
		if kt != "user" {
			return nil, badRequest("TAG_KIND_MISMATCH", "tagIds", "only player tags")
		}
		found++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if found != len(clean) {
		return nil, badRequest("TAG_NOT_FOUND", "tagIds", "a tag does not exist")
	}
	return clean, nil
}
