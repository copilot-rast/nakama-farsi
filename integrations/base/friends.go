package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/hookchain"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"google.golang.org/grpc/codes"
)

const (
	friendsLease      = "base.friends.expire"
	friendsLeaseTTL   = 2 * time.Minute
	friendsTick       = time.Minute
	friendsSweepBatch = 500
	friendLookupMax   = 100

	edgeFriend   = 0
	edgeSent     = 1
	edgeReceived = 2
	edgeBlocked  = 3

	CodeFriendDailyLimit   = "FRIEND_DAILY_LIMIT"
	CodeFriendLimit        = "FRIEND_LIMIT_REACHED"
	CodeFriendTargetLimit  = "FRIEND_TARGET_LIMIT_REACHED"
	CodeFriendCooldown     = "FRIEND_REQUEST_COOLDOWN"
	CodeFriendTooManyNames = "FRIEND_TOO_MANY_USERS"
)

type friendRules struct {
	DailyRequests       int       `json:"dailyRequests"`
	MaxFriends          int       `json:"maxFriends"`
	ExpiryDays          int       `json:"expiryDays"`
	RejectCooldownHours int       `json:"rejectCooldownHours"`
	RequestsBonusMeta   string    `json:"requestsBonusMeta"`
	FriendsBonusMeta    string    `json:"friendsBonusMeta"`
	MatchPolicy         string    `json:"matchPolicy"`
	OverLimit           string    `json:"overLimit"`
	NotifyDeclined      bool      `json:"notifyDeclined"`
	NotifyRemoved       bool      `json:"notifyRemoved"`
	PresenceFriendsOnly bool      `json:"presenceFriendsOnly"`
	DMFriendsOnly       bool      `json:"dmFriendsOnly"`
	UpdatedBy           string    `json:"updatedBy,omitempty"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

var (
	friendMatchPolicies = []string{"priority", "generous"}
	friendOverLimits    = []string{"refuse", "trim"}
)

func defaultFriendRules() friendRules {
	return friendRules{DailyRequests: 20, MatchPolicy: "priority", OverLimit: "refuse", NotifyDeclined: true, NotifyRemoved: true}
}

func loadFriendRules(ctx context.Context, q queryer) (friendRules, error) {
	r := defaultFriendRules()
	err := q.QueryRowContext(ctx, `SELECT daily_requests, max_friends, expiry_days, reject_cooldown_hours, requests_bonus_meta, friends_bonus_meta, match_policy, over_limit, notify_declined, notify_removed, presence_friends_only, dm_friends_only, updated_by, updated_at FROM base_friend_rules WHERE id=1`).
		Scan(&r.DailyRequests, &r.MaxFriends, &r.ExpiryDays, &r.RejectCooldownHours, &r.RequestsBonusMeta, &r.FriendsBonusMeta, &r.MatchPolicy, &r.OverLimit,
			&r.NotifyDeclined, &r.NotifyRemoved, &r.PresenceFriendsOnly, &r.DMFriendsOnly, &r.UpdatedBy, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultFriendRules(), nil
	}
	return r, err
}

func (r friendRules) clean() (friendRules, error) {
	r.RequestsBonusMeta = strings.TrimSpace(r.RequestsBonusMeta)
	r.FriendsBonusMeta = strings.TrimSpace(r.FriendsBonusMeta)
	if r.MatchPolicy == "" {
		r.MatchPolicy = "priority"
	}
	if r.OverLimit == "" {
		r.OverLimit = "refuse"
	}
	switch {
	case r.DailyRequests < 0 || r.DailyRequests > 1000:
		return r, badRequest("INVALID_RULE", "dailyRequests", "between 0 and 1000")
	case r.MaxFriends < 0 || r.MaxFriends > 10000:
		return r, badRequest("INVALID_RULE", "maxFriends", "between 0 and 10000")
	case r.ExpiryDays < 0 || r.ExpiryDays > 365:
		return r, badRequest("INVALID_RULE", "expiryDays", "between 0 and 365")
	case r.RejectCooldownHours < 0 || r.RejectCooldownHours > 8760:
		return r, badRequest("INVALID_RULE", "rejectCooldownHours", "between 0 and 8760")
	case !validKeyType(friendMatchPolicies, r.MatchPolicy):
		return r, badRequest("INVALID_RULE", "matchPolicy", "priority or generous")
	case !validKeyType(friendOverLimits, r.OverLimit):
		return r, badRequest("INVALID_RULE", "overLimit", "refuse or trim")
	}
	return r, nil
}

func (s *service) checkMetaKey(ctx context.Context, field, key string) error {
	if key == "" {
		return nil
	}
	var ok bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_meta WHERE key=$1)`, key).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return badRequest("META_NOT_FOUND", field, "no meta with this key")
	}
	return nil
}

func (s *service) friendRulesGet(w http.ResponseWriter, r *http.Request) {
	rules, err := loadFriendRules(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (s *service) friendRulesPut(w http.ResponseWriter, r *http.Request) {
	in := defaultFriendRules()
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	rules, err := in.clean()
	if err == nil {
		err = s.checkMetaKey(r.Context(), "requestsBonusMeta", rules.RequestsBonusMeta)
	}
	if err == nil {
		err = s.checkMetaKey(r.Context(), "friendsBonusMeta", rules.FriendsBonusMeta)
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	_, err = s.db.ExecContext(r.Context(), `INSERT INTO base_friend_rules (id, daily_requests, max_friends, expiry_days, reject_cooldown_hours, requests_bonus_meta, friends_bonus_meta, match_policy, over_limit, notify_declined, notify_removed, presence_friends_only, dm_friends_only, updated_by, updated_at)
		VALUES (1,$1,$2,$3,$4,$5,$6,$7,$8,$10,$11,$12,$13,$9,now())
		ON CONFLICT (id) DO UPDATE SET daily_requests=EXCLUDED.daily_requests, max_friends=EXCLUDED.max_friends, expiry_days=EXCLUDED.expiry_days,
		reject_cooldown_hours=EXCLUDED.reject_cooldown_hours, requests_bonus_meta=EXCLUDED.requests_bonus_meta, friends_bonus_meta=EXCLUDED.friends_bonus_meta,
		match_policy=EXCLUDED.match_policy, over_limit=EXCLUDED.over_limit, notify_declined=EXCLUDED.notify_declined, notify_removed=EXCLUDED.notify_removed,
		presence_friends_only=EXCLUDED.presence_friends_only, dm_friends_only=EXCLUDED.dm_friends_only, updated_by=EXCLUDED.updated_by, updated_at=now()`,
		rules.DailyRequests, rules.MaxFriends, rules.ExpiryDays, rules.RejectCooldownHours, rules.RequestsBonusMeta, rules.FriendsBonusMeta, rules.MatchPolicy, rules.OverLimit, consoleUser(r),
		rules.NotifyDeclined, rules.NotifyRemoved, rules.PresenceFriendsOnly, rules.DMFriendsOnly)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.friendRulesGet(w, r)
}

func projectZone(ctx context.Context, q queryer) *time.Location {
	var name string
	if err := q.QueryRowContext(ctx, `SELECT timezone FROM economy_event_settings WHERE id=1`).Scan(&name); err != nil {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

func dayBounds(now time.Time, zone *time.Location) (string, time.Time) {
	local := now.In(zone)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, zone)
	return start.Format("2006-01-02"), start.AddDate(0, 0, 1)
}

var friendMessages = map[string]localized.Text{
	CodeFriendDailyLimit: {
		"fa": "امروز بیشتر از {limit} درخواست دوستی نمی‌توانی بفرستی. از {reset} دوباره می‌توانی.",
		"en": "You can send at most {limit} friend requests a day. You can send more after {reset}.",
		"ar": "يمكنك إرسال {limit} طلب صداقة كحد أقصى في اليوم. يمكنك الإرسال مجدداً بعد {reset}.",
	},
	CodeFriendLimit: {
		"fa": "فهرست دوستانت پر است (بیشترین تعداد: {limit}). برای دوست تازه، یکی از دوستانت را حذف کن.",
		"en": "Your friend list is full ({limit} friends at most). Remove a friend to add a new one.",
		"ar": "قائمة أصدقائك ممتلئة (الحد الأقصى {limit}). احذف صديقاً لإضافة صديق جديد.",
	},
	CodeFriendTargetLimit: {
		"fa": "فهرست دوستان این بازیکن پر است.",
		"en": "This player's friend list is full.",
		"ar": "قائمة أصدقاء هذا اللاعب ممتلئة.",
	},
	CodeFriendCooldown: {
		"fa": "این بازیکن درخواستت را رد کرده است. از {retry} دوباره می‌توانی درخواست بدهی.",
		"en": "This player declined your request. You can ask again after {retry}.",
		"ar": "رفض هذا اللاعب طلبك. يمكنك الطلب مجدداً بعد {retry}.",
	},
	CodeFriendTooManyNames: {
		"fa": "در هر درخواست حداکثر {limit} بازیکن را می‌توانی اضافه کنی.",
		"en": "You can add at most {limit} players in one request.",
		"ar": "يمكنك إضافة {limit} لاعب كحد أقصى في طلب واحد.",
	},
}

func localDigits(lang, s string) string {
	if lang != "fa" {
		return s
	}
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return '۰' + (r - '0')
		}
		return r
	}, s)
}

type friendError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Lang      string `json:"lang"`
	Limit     int    `json:"limit,omitempty"`
	Remaining *int   `json:"remaining,omitempty"`
	ResetsAt  string `json:"resetsAt,omitempty"`
	RetryAt   string `json:"retryAt,omitempty"`
	UserID    string `json:"userId,omitempty"`
	grpc      codes.Code
}

func (e *friendError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (s *service) friendFailure(ctx context.Context, userID string, e *friendError, zone *time.Location) error {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, ctxLang(ctx))
	if err != nil {
		e.Lang = "en"
		e.Message = e.Code
	} else {
		e.Lang = lang
		text := langs.Resolve(friendMessages[e.Code], lang)
		if text == "" {
			text = friendMessages[e.Code]["en"]
		}
		when := func(iso string) string {
			at, perr := time.Parse(time.RFC3339, iso)
			if perr != nil {
				return iso
			}
			return localDigits(lang, at.In(zone).Format("2006-01-02 15:04"))
		}
		text = strings.ReplaceAll(text, "{limit}", localDigits(lang, strconv.Itoa(e.Limit)))
		if e.ResetsAt != "" {
			text = strings.ReplaceAll(text, "{reset}", when(e.ResetsAt))
		}
		if e.RetryAt != "" {
			text = strings.ReplaceAll(text, "{retry}", when(e.RetryAt))
		}
		e.Message = text
	}
	return runtime.NewError(e.Error(), int(e.grpc))
}

type friendTarget struct {
	ID       string
	Username string
	Mine     int
	Theirs   int
}

func (s *service) friendTargets(ctx context.Context, userID string, ids, usernames []string) ([]friendTarget, error) {
	valid := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, err := uuid.FromString(strings.TrimSpace(id)); err == nil {
			valid = append(valid, strings.TrimSpace(id))
		}
	}
	if usernames == nil {
		usernames = []string{}
	}
	rows, err := s.db.QueryContext(ctx, `SELECT u.id::text, u.username, COALESCE(mine.state, -1), COALESCE(theirs.state, -1) FROM users u
		LEFT JOIN user_edge mine ON mine.source_id=$1::uuid AND mine.destination_id=u.id
		LEFT JOIN user_edge theirs ON theirs.source_id=u.id AND theirs.destination_id=$1::uuid
		WHERE (u.id = ANY($2::uuid[]) OR u.username = ANY($3::text[])) AND u.id <> $1::uuid AND u.id <> '00000000-0000-0000-0000-000000000000'::uuid`, userID, valid, usernames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []friendTarget{}
	for rows.Next() {
		var t friendTarget
		if err := rows.Scan(&t.ID, &t.Username, &t.Mine, &t.Theirs); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *service) friendCount(ctx context.Context, userID string) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM user_edge WHERE source_id=$1::uuid AND state=0`, userID).Scan(&n)
	return n, err
}

func (s *service) sentToday(ctx context.Context, userID, day string) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE((SELECT count FROM base_friend_request_day WHERE user_id=$1::uuid AND day=$2::date), 0)`, userID, day).Scan(&n)
	return n, err
}

func (s *service) reserveRequests(ctx context.Context, userID, day string, want, limit int, trim bool) (int, error) {
	took := 0
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO base_friend_request_day (user_id, day, count) VALUES ($1::uuid, $2::date, 0) ON CONFLICT (user_id, day) DO NOTHING`, userID, day); err != nil {
			return err
		}
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT count FROM base_friend_request_day WHERE user_id=$1::uuid AND day=$2::date FOR UPDATE`, userID, day).Scan(&count); err != nil {
			return err
		}
		took = want
		if limit > 0 && count+want > limit {
			if !trim {
				took = 0
				return nil
			}
			took = max(limit-count, 0)
		}
		if took == 0 {
			return nil
		}
		_, err := tx.ExecContext(ctx, `UPDATE base_friend_request_day SET count = count + $3, updated_at = now() WHERE user_id=$1::uuid AND day=$2::date`, userID, day, took)
		return err
	})
	return took, err
}

func orderTargets(targets []friendTarget, in *api.AddFriendsRequest) {
	rank := map[string]int{}
	for i, id := range in.GetIds() {
		if _, ok := rank[strings.TrimSpace(id)]; !ok {
			rank[strings.TrimSpace(id)] = i
		}
	}
	for i, name := range in.GetUsernames() {
		if _, ok := rank["u:"+name]; !ok {
			rank["u:"+name] = len(in.GetIds()) + i
		}
	}
	pos := func(t friendTarget) int {
		if v, ok := rank[t.ID]; ok {
			return v
		}
		if v, ok := rank["u:"+t.Username]; ok {
			return v
		}
		return 1 << 30
	}
	sort.SliceStable(targets, func(i, j int) bool { return pos(targets[i]) < pos(targets[j]) })
}

func (s *service) checkFriendAdd(ctx context.Context, userID string, in *api.AddFriendsRequest) (*api.AddFriendsRequest, error) {
	if len(in.GetIds())+len(in.GetUsernames()) > friendLookupMax {
		return nil, &friendError{Code: CodeFriendTooManyNames, Limit: friendLookupMax, grpc: codes.InvalidArgument}
	}
	rules, err := loadFriendRules(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if rules.ExpiryDays > 0 {
		if _, err := s.expireInvites(ctx, userID, rules.ExpiryDays, friendLookupMax); err != nil {
			return nil, err
		}
	}
	targets, err := s.friendTargets(ctx, userID, in.GetIds(), in.GetUsernames())
	if err != nil {
		return nil, err
	}
	orderTargets(targets, in)
	invites := []string{}
	accepts := []string{}
	for _, t := range targets {
		switch {
		case t.Mine == edgeReceived:
			accepts = append(accepts, t.ID)
		case t.Mine == -1 && t.Theirs != edgeBlocked:
			invites = append(invites, t.ID)
		}
	}
	if len(invites) == 0 && len(accepts) == 0 {
		return in, nil
	}
	eff, err := s.effectiveFriendLimits(ctx, rules, userID)
	if err != nil {
		return nil, err
	}
	if eff.MaxFriends > 0 {
		n, err := s.friendCount(ctx, userID)
		if err != nil {
			return nil, err
		}
		if n+len(accepts) > eff.MaxFriends || (len(invites) > 0 && n >= eff.MaxFriends) {
			return nil, &friendError{Code: CodeFriendLimit, Limit: eff.MaxFriends, grpc: codes.FailedPrecondition}
		}
	}
	for _, id := range accepts {
		other, err := s.effectiveFriendLimits(ctx, rules, id)
		if err != nil {
			return nil, err
		}
		if other.MaxFriends == 0 {
			continue
		}
		n, err := s.friendCount(ctx, id)
		if err != nil {
			return nil, err
		}
		if n >= other.MaxFriends {
			return nil, &friendError{Code: CodeFriendTargetLimit, Limit: other.MaxFriends, UserID: id, grpc: codes.FailedPrecondition}
		}
	}
	if len(invites) == 0 {
		return in, nil
	}
	if eff.RejectCooldownHours > 0 {
		var rejecter string
		var at time.Time
		err := s.db.QueryRowContext(ctx, `SELECT rejecter_id::text, rejected_at FROM base_friend_rejection WHERE requester_id=$1::uuid AND rejecter_id = ANY($2::uuid[])
			AND rejected_at > now() - $3::INT * INTERVAL '1 hour' ORDER BY rejected_at DESC LIMIT 1`, userID, invites, eff.RejectCooldownHours).Scan(&rejecter, &at)
		if err == nil {
			retry := at.Add(time.Duration(eff.RejectCooldownHours) * time.Hour).UTC().Format(time.RFC3339)
			return nil, &friendError{Code: CodeFriendCooldown, RetryAt: retry, UserID: rejecter, grpc: codes.FailedPrecondition}
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	zone := projectZone(ctx, s.db)
	day, reset := dayBounds(s.now(), zone)
	took, err := s.reserveRequests(ctx, userID, day, len(invites), eff.DailyRequests, rules.OverLimit == "trim")
	if err != nil {
		return nil, err
	}
	if took == 0 {
		sent, err := s.sentToday(ctx, userID, day)
		if err != nil {
			return nil, err
		}
		left := max(eff.DailyRequests-sent, 0)
		return nil, &friendError{Code: CodeFriendDailyLimit, Limit: eff.DailyRequests, Remaining: &left, ResetsAt: reset.UTC().Format(time.RFC3339), grpc: codes.ResourceExhausted}
	}
	if took == len(invites) {
		return in, nil
	}
	ids := append(append([]string{}, accepts...), invites[:took]...)
	return &api.AddFriendsRequest{Ids: ids, Metadata: in.GetMetadata()}, nil
}

func (s *service) beforeAddFriends(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AddFriendsRequest) (*api.AddFriendsRequest, error) {
	userID := ctxUser(ctx)
	if userID == "" || in == nil {
		return in, nil
	}
	out, err := s.checkFriendAdd(ctx, userID, in)
	if err != nil {
		var fe *friendError
		if errors.As(err, &fe) {
			return nil, s.friendFailure(ctx, userID, fe, projectZone(ctx, s.db))
		}
		return nil, s.hookFailure(err)
	}
	return out, nil
}

func (s *service) afterAddFriends(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.AddFriendsRequest) error {
	userID := ctxUser(ctx)
	if userID == "" || in == nil {
		return nil
	}
	targets, err := s.friendTargets(ctx, userID, in.GetIds(), in.GetUsernames())
	if err != nil || len(targets) == 0 {
		return nil
	}
	ids := make([]string, 0, len(targets))
	for _, t := range targets {
		ids = append(ids, t.ID)
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM base_friend_rejection WHERE rejecter_id=$1::uuid AND requester_id = ANY($2::uuid[])`, userID, ids); err != nil {
		s.logger.Warn("Base friend rejections not cleared: %v", err)
	}
	s.dropSuggestionPairs(ctx, userID, ids)
	return nil
}

func (s *service) beforeDeleteFriends(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.DeleteFriendsRequest) (*api.DeleteFriendsRequest, error) {
	userID := ctxUser(ctx)
	if userID == "" || in == nil || len(in.GetIds())+len(in.GetUsernames()) > friendLookupMax {
		return in, nil
	}
	targets, err := s.friendTargets(ctx, userID, in.GetIds(), in.GetUsernames())
	if err != nil {
		return nil, s.hookFailure(err)
	}
	rejected := []string{}
	removed := []string{}
	for _, t := range targets {
		switch t.Mine {
		case edgeReceived:
			rejected = append(rejected, t.ID)
		case edgeFriend:
			removed = append(removed, t.ID)
		}
	}
	if len(rejected) > 0 {
		if _, err := s.db.ExecContext(ctx, `INSERT INTO base_friend_rejection (requester_id, rejecter_id, rejected_at)
			SELECT r, $1::uuid, now() FROM unnest($2::uuid[]) AS r ON CONFLICT (requester_id, rejecter_id) DO UPDATE SET rejected_at = now()`, userID, rejected); err != nil {
			return nil, s.hookFailure(err)
		}
	}
	s.friendEdgeNotices(ctx, userID, rejected, removed)
	return in, nil
}

func (s *service) beforeListFriends(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.ListFriendsRequest) (*api.ListFriendsRequest, error) {
	userID := ctxUser(ctx)
	if userID == "" {
		return in, nil
	}
	rules, err := loadFriendRules(ctx, s.db)
	if err != nil {
		return nil, s.hookFailure(err)
	}
	if rules.ExpiryDays > 0 {
		if _, err := s.expireInvites(ctx, userID, rules.ExpiryDays, friendLookupMax); err != nil {
			s.logger.Warn("Base friend invites not expired: %v", err)
		}
	}
	return in, nil
}

func (s *service) expireInvites(ctx context.Context, userID string, days, limit int) (int, error) {
	if days <= 0 {
		return 0, nil
	}
	var user any
	if userID != "" {
		user = userID
	}
	rows, err := s.db.QueryContext(ctx, `SELECT source_id::text, destination_id::text FROM user_edge
		WHERE state=1 AND update_time < now() - $1::INT * INTERVAL '1 day' AND ($2::uuid IS NULL OR source_id=$2::uuid OR destination_id=$2::uuid)
		ORDER BY update_time LIMIT $3`, days, user, limit)
	if err != nil {
		return 0, err
	}
	type pair struct{ from, to string }
	pairs := []pair{}
	for rows.Next() {
		var p pair
		if err := rows.Scan(&p.from, &p.to); err != nil {
			rows.Close()
			return 0, err
		}
		pairs = append(pairs, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	n := 0
	for _, p := range pairs {
		if err := s.nk.FriendsDelete(ctx, p.from, "", []string{p.to}, nil); err != nil {
			s.logger.Warn("Base friend invite %s -> %s not expired: %v", p.from, p.to, err)
			continue
		}
		n++
	}
	return n, nil
}

func (s *service) friendsLoop(ctx context.Context) {
	ticker := time.NewTicker(friendsTick)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			_ = clusterkit.ReleaseLease(context.Background(), s.db, friendsLease, holder)
			return
		case <-ticker.C:
		}
		if _, err := s.friendsSweep(ctx, holder); err != nil && ctx.Err() == nil {
			s.logger.Warn("Base friend sweep failed: %v", err)
		}
	}
}

func (s *service) friendsSweep(ctx context.Context, holder string) (int, error) {
	ok, err := clusterkit.AcquireLease(ctx, s.db, friendsLease, holder, friendsLeaseTTL)
	if err != nil || !ok {
		return 0, err
	}
	rules, err := loadFriendRules(ctx, s.db)
	if err != nil {
		return 0, err
	}
	n, err := s.expireInvites(ctx, "", rules.ExpiryDays, friendsSweepBatch)
	if err != nil {
		return n, err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM base_friend_request_day WHERE day < (now() - INTERVAL '3 days')::date`); err != nil {
		return n, err
	}
	keep := max(rules.RejectCooldownHours, 24*30)
	_, err = s.db.ExecContext(ctx, `DELETE FROM base_friend_rejection WHERE rejected_at < now() - $1::INT * INTERVAL '1 hour'`, keep)
	return n, err
}

func (s *service) rpcFriendRules(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	rules, err := loadFriendRules(ctx, s.db)
	if err != nil {
		return nil, err
	}
	eff, err := s.effectiveFriendLimits(ctx, rules, userID)
	if err != nil {
		return nil, err
	}
	zone := projectZone(ctx, s.db)
	day, reset := dayBounds(s.now(), zone)
	sent, err := s.sentToday(ctx, userID, day)
	if err != nil {
		return nil, err
	}
	friends, err := s.friendCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	var remaining, slotsLeft any
	if eff.DailyRequests > 0 {
		remaining = max(eff.DailyRequests-sent, 0)
	}
	if eff.MaxFriends > 0 {
		slotsLeft = max(eff.MaxFriends-friends, 0)
	}
	var tier any
	if eff.TierID != "" {
		tier = map[string]any{"id": eff.TierID, "name": eff.TierName}
	}
	return map[string]any{
		"dailyRequests":       eff.DailyRequests,
		"sentToday":           sent,
		"remainingToday":      remaining,
		"resetsAt":            reset.UTC().Format(time.RFC3339),
		"maxFriends":          eff.MaxFriends,
		"friends":             friends,
		"friendSlotsLeft":     slotsLeft,
		"expiryDays":          rules.ExpiryDays,
		"rejectCooldownHours": eff.RejectCooldownHours,
		"base":                map[string]any{"dailyRequests": eff.BaseDaily, "maxFriends": eff.BaseMax},
		"bonus":               map[string]any{"dailyRequests": eff.BonusDaily, "maxFriends": eff.BonusMax},
		"tier":                tier,
		"overLimit":           rules.OverLimit,
		"timeZone":            zone.String(),
		"presenceFriendsOnly": rules.PresenceFriendsOnly,
		"dmFriendsOnly":       rules.DMFriendsOnly,
	}, nil
}

func (s *service) registerFriendHooks(initializer runtime.Initializer) error {
	if err := initializer.RegisterBeforeAddFriends(hookchain.Before("AddFriends", s.beforeAddFriends)); err != nil {
		return err
	}
	if err := initializer.RegisterAfterAddFriends(hookchain.AfterIn("AddFriends", s.afterAddFriends)); err != nil {
		return err
	}
	if err := initializer.RegisterBeforeDeleteFriends(hookchain.Before("DeleteFriends", s.beforeDeleteFriends)); err != nil {
		return err
	}
	return initializer.RegisterBeforeListFriends(hookchain.Before("ListFriends", s.beforeListFriends))
}
