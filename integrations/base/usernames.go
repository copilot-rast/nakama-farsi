package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
	"google.golang.org/grpc/codes"
)

const (
	CodeUsernameRequired     = "USERNAME_REQUIRED"
	CodeUsernameUnchanged    = "USERNAME_UNCHANGED"
	CodeUsernameTaken        = "USERNAME_TAKEN"
	CodeUsernameCooldown     = "USERNAME_CHANGE_COOLDOWN"
	CodeUsernamePriceChanged = "USERNAME_PRICE_CHANGED"
	CodeCurrencyNotEnough    = "CURRENCY_IS_NOT_ENOUGH"
	usernameHistoryMax       = 50
	usernameLadderMax        = 20
)

type usernameSettings struct {
	Enabled       bool    `json:"enabled"`
	FirstFree     bool    `json:"firstFree"`
	Currency      string  `json:"currency"`
	Ladder        []int64 `json:"ladder"`
	CooldownHours int     `json:"cooldownHours"`
}

func defaultUsernameSettings() usernameSettings {
	return usernameSettings{Enabled: false, FirstFree: true, Currency: "gem", Ladder: []int64{50, 100, 200, 400, 800}, CooldownHours: 0}
}

var usernameStore = settingsrow.New("base_username_settings", defaultUsernameSettings)

func (u usernameSettings) clean() (usernameSettings, error) {
	u.Currency = strings.TrimSpace(u.Currency)
	if u.Ladder == nil {
		u.Ladder = []int64{}
	}
	switch {
	case u.CooldownHours < 0 || u.CooldownHours > 8760:
		return u, badRequest("INVALID_RULE", "cooldownHours", "between 0 and 8760")
	case len(u.Ladder) > usernameLadderMax:
		return u, badRequest("INVALID_RULE", "ladder", "at most 20 prices")
	case len(u.Ladder) > 0 && u.Currency == "":
		return u, badRequest("CURRENCY_REQUIRED", "currency", "pick the currency of the prices")
	}
	for _, v := range u.Ladder {
		if v < 1 || v > 1_000_000_000 {
			return u, badRequest("INVALID_RULE", "ladder", "every price between 1 and 1000000000")
		}
	}
	return u, nil
}

type usernameState struct {
	FreeUsed  bool
	PaidCount int
	Last      sql.NullTime
}

func loadUsernameState(ctx context.Context, q queryer, userID, lock string) (usernameState, error) {
	var st usernameState
	err := q.QueryRowContext(ctx, `SELECT free_used, paid_count, last_change_at FROM base_username_state WHERE user_id=$1`+lock, userID).Scan(&st.FreeUsed, &st.PaidCount, &st.Last)
	if errors.Is(err, sql.ErrNoRows) {
		return usernameState{}, nil
	}
	return st, err
}

type usernameQuote struct {
	Enabled        bool   `json:"enabled"`
	Free           bool   `json:"free"`
	Cost           int64  `json:"cost"`
	Currency       string `json:"currency"`
	CooldownHours  int    `json:"cooldownHours"`
	CooldownEndsAt string `json:"cooldownEndsAt,omitempty"`
	PaidChanges    int    `json:"paidChanges"`
}

func (u usernameSettings) quote(st usernameState, now time.Time) usernameQuote {
	q := usernameQuote{Enabled: u.Enabled, Currency: u.Currency, CooldownHours: u.CooldownHours, PaidChanges: st.PaidCount, Free: true}
	if !u.Enabled {
		return q
	}
	if !(u.FirstFree && !st.FreeUsed) && len(u.Ladder) > 0 {
		q.Free = false
		q.Cost = u.Ladder[min(st.PaidCount, len(u.Ladder)-1)]
	}
	if u.CooldownHours > 0 && st.Last.Valid {
		ends := st.Last.Time.Add(time.Duration(u.CooldownHours) * time.Hour)
		if ends.After(now) {
			q.CooldownEndsAt = ends.UTC().Format(time.RFC3339)
		}
	}
	return q
}

var usernameMessages = map[string]localized.Text{
	CodeUsernameRequired:     {"fa": "نام کاربری تازه را بنویس.", "en": "Type the new username.", "ar": "اكتب اسم المستخدم الجديد."},
	CodeUsernameUnchanged:    {"fa": "نام کاربری تازه با نام فعلی یکی است.", "en": "Choose a different username.", "ar": "اختر اسم مستخدم مختلفاً."},
	CodeUsernameTaken:        {"fa": "این نام کاربری را کس دیگری برداشته است.", "en": "This username is taken.", "ar": "اسم المستخدم هذا مأخوذ."},
	CodeUsernameCooldown:     {"fa": "نام کاربری را تا {until} نمی‌توانی دوباره عوض کنی.", "en": "You can change your username again after {until}.", "ar": "يمكنك تغيير اسم المستخدم مجدداً بعد {until}."},
	CodeUsernamePriceChanged: {"fa": "هزینهٔ تغییر نام عوض شده است: {cost}.", "en": "The price of a username change is now {cost}.", "ar": "تغيّر سعر تغيير الاسم: {cost}."},
	CodeCurrencyNotEnough:    {"fa": "برای تغییر نام {cost} لازم است.", "en": "A username change costs {cost}.", "ar": "تغيير الاسم يكلف {cost}."},
}

type socialError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Lang    string         `json:"lang"`
	Meta    map[string]any `json:"meta,omitempty"`
}

func (s *service) socialRefusal(ctx context.Context, userID, code string, grpc codes.Code, texts map[string]localized.Text, meta map[string]any, fill map[string]string) error {
	lang := "en"
	text := ""
	if langs, l, err := localized.ForUser(ctx, s.db, userID, ctxLang(ctx)); err == nil && langs != nil {
		lang = l
		text = langs.Resolve(texts[code], lang)
	}
	if text == "" {
		text = texts[code]["en"]
	}
	if text == "" {
		text = code
	}
	for k, v := range fill {
		text = strings.ReplaceAll(text, "{"+k+"}", localDigits(lang, v))
	}
	raw, _ := json.Marshal(socialError{Code: code, Message: text, Lang: lang, Meta: meta})
	return runtime.NewError(string(raw), int(grpc))
}

func (s *service) usernameRefusal(ctx context.Context, userID, code string, grpc codes.Code, meta map[string]any, fill map[string]string) error {
	return s.socialRefusal(ctx, userID, code, grpc, usernameMessages, meta, fill)
}

type usernameChangeIn struct {
	UserID   string
	Name     string
	Expected *int64
	Admin    string
}

type usernameChangeOut struct {
	Username    string        `json:"username"`
	OldUsername string        `json:"oldUsername"`
	Source      string        `json:"source"`
	Cost        int64         `json:"cost"`
	Currency    string        `json:"currency,omitempty"`
	Quote       usernameQuote `json:"quote"`
}

func (s *service) changeUsername(ctx context.Context, in usernameChangeIn) (*usernameChangeOut, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, s.usernameRefusal(ctx, in.UserID, CodeUsernameRequired, codes.InvalidArgument, nil, nil)
	}
	rules, err := loadRules(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if code := rules.usernameError(name); code != "" {
		return nil, ruleError(code)
	}
	settings, err := usernameStore.Load(ctx, s.db)
	if err != nil {
		return nil, err
	}
	admin := in.Admin != ""
	now := s.now().UTC()
	out := &usernameChangeOut{Username: name}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `SELECT username FROM users WHERE id=$1 FOR UPDATE`, in.UserID).Scan(&out.OldUsername); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return notFound("USER_NOT_FOUND")
			}
			return err
		}
		if out.OldUsername == name {
			return s.usernameRefusal(ctx, in.UserID, CodeUsernameUnchanged, codes.InvalidArgument, nil, nil)
		}
		var taken bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE username=$1 AND id<>$2)`, name, in.UserID).Scan(&taken); err != nil {
			return err
		}
		if taken {
			return s.usernameRefusal(ctx, in.UserID, CodeUsernameTaken, codes.AlreadyExists, nil, nil)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO base_username_state (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, in.UserID); err != nil {
			return err
		}
		st, err := loadUsernameState(ctx, tx, in.UserID, " FOR UPDATE")
		if err != nil {
			return err
		}
		q := settings.quote(st, now)
		out.Source = "free"
		if admin {
			out.Source = "admin"
		} else {
			if q.CooldownEndsAt != "" {
				ends, _ := time.Parse(time.RFC3339, q.CooldownEndsAt)
				return s.usernameRefusal(ctx, in.UserID, CodeUsernameCooldown, codes.FailedPrecondition, map[string]any{"cooldownEndsAt": q.CooldownEndsAt},
					map[string]string{"until": ends.In(projectZone(ctx, tx)).Format("2006-01-02 15:04")})
			}
			if in.Expected != nil && *in.Expected != q.Cost {
				return s.usernameRefusal(ctx, in.UserID, CodeUsernamePriceChanged, codes.FailedPrecondition, map[string]any{"cost": q.Cost, "currency": q.Currency, "expected": *in.Expected},
					map[string]string{"cost": itoa64(q.Cost) + " " + q.Currency})
			}
			if q.Cost > 0 {
				out.Source, out.Cost, out.Currency = "paid", q.Cost, q.Currency
			}
		}
		id := uuid.Must(uuid.NewV4()).String()
		if out.Cost > 0 {
			err := pluginkit.WalletTx(ctx, tx, in.UserID, "username:"+id, map[string]int64{out.Currency: -out.Cost}, map[string]any{"module": "base", "reason": "username_change", "sourceType": "username", "sourceId": id})
			var neg *pluginkit.ErrInsufficient
			if errors.As(err, &neg) {
				var balance int64
				_ = tx.QueryRowContext(ctx, `SELECT COALESCE((wallet->>$2)::BIGINT, 0) FROM users WHERE id=$1`, in.UserID, out.Currency).Scan(&balance)
				return s.usernameRefusal(ctx, in.UserID, CodeCurrencyNotEnough, codes.FailedPrecondition, map[string]any{"currencyKey": out.Currency, "amount": out.Cost, "balance": balance},
					map[string]string{"cost": itoa64(out.Cost) + " " + out.Currency})
			}
			if err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE users SET username=$2, update_time=now() WHERE id=$1`, in.UserID, name); err != nil {
			if _, unique := isUniqueViolation(err); unique {
				return s.usernameRefusal(ctx, in.UserID, CodeUsernameTaken, codes.AlreadyExists, nil, nil)
			}
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO base_username_change (id, user_id, old_username, new_username, source, cost, currency, admin) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			id, in.UserID, out.OldUsername, name, out.Source, out.Cost, out.Currency, in.Admin); err != nil {
			return err
		}
		if !admin && settings.Enabled {
			paid := 0
			if out.Source == "paid" {
				paid = 1
			}
			if _, err := tx.ExecContext(ctx, `UPDATE base_username_state SET free_used=true, paid_count=paid_count+$2, last_change_at=$3, updated_at=now() WHERE user_id=$1`, in.UserID, paid, now); err != nil {
				return err
			}
			st.FreeUsed, st.PaidCount, st.Last = true, st.PaidCount+paid, sql.NullTime{Time: now, Valid: true}
		}
		out.Quote = settings.quote(st, now)
		return nil
	})
	if err != nil {
		return nil, err
	}
	actor := "player"
	if admin {
		actor = in.Admin
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: in.UserID, Kind: activity.KindAccount, Action: "username_changed", Actor: actor,
		Data: map[string]any{"from": out.OldUsername, "to": out.Username, "source": out.Source, "cost": out.Cost, "currency": out.Currency}})
	return out, nil
}

func itoa64(v int64) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (s *service) usernameQuoteFor(ctx context.Context, userID string) (usernameQuote, error) {
	settings, err := usernameStore.Load(ctx, s.db)
	if err != nil {
		return usernameQuote{}, err
	}
	st, err := loadUsernameState(ctx, s.db, userID, "")
	if err != nil {
		return usernameQuote{}, err
	}
	return settings.quote(st, s.now().UTC()), nil
}

func (s *service) rpcUsernameQuote(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	return s.usernameQuoteFor(ctx, userID)
}

func (s *service) rpcUsernameChange(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Username     string `json:"username"`
		ExpectedCost *int64 `json:"expectedCost"`
		Lang         string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	return s.changeUsername(ctx, usernameChangeIn{UserID: userID, Name: in.Username, Expected: in.ExpectedCost})
}

func (s *service) usernameThroughPolicy(ctx context.Context, userID, name string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	settings, err := usernameStore.Load(ctx, s.db)
	if err != nil || !settings.Enabled {
		return false, err
	}
	if _, err := s.changeUsername(ctx, usernameChangeIn{UserID: userID, Name: name}); err != nil {
		return false, err
	}
	return true, nil
}

func (s *service) usernameSettingsGet(w http.ResponseWriter, r *http.Request) {
	v, meta, err := usernameStore.Read(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"enabled": v.Enabled, "firstFree": v.FirstFree, "currency": v.Currency, "ladder": v.Ladder, "cooldownHours": v.CooldownHours,
		"updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt})
}

func (s *service) usernameSettingsPut(w http.ResponseWriter, r *http.Request) {
	in := defaultUsernameSettings()
	var raw map[string]json.RawMessage
	if err := readBody(r, &raw); err != nil {
		s.writeError(w, err)
		return
	}
	delete(raw, "updatedBy")
	delete(raw, "updatedAt")
	b, _ := json.Marshal(raw)
	if err := json.Unmarshal(b, &in); err != nil {
		s.writeError(w, badRequest("INVALID_JSON", "", "settings are not valid"))
		return
	}
	v, err := in.clean()
	if err == nil && v.Currency != "" {
		err = s.checkCurrencyKey(r.Context(), v.Currency)
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := usernameStore.Save(r.Context(), s.db, v, consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	s.usernameSettingsGet(w, r)
}

func (s *service) checkCurrencyKey(ctx context.Context, key string) error {
	var ok bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, key).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return badRequest("CURRENCY_NOT_FOUND", "currency", "no currency with this key")
	}
	return nil
}

func (s *service) playerUsernameChanges(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if err := s.userExists(r.Context(), s.db, id); err != nil {
		s.writeError(w, err)
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT jsonb_build_object('id',id,'oldUsername',old_username,'newUsername',new_username,'source',source,'cost',cost,'currency',currency,'admin',admin,'createdAt',created_at)::text
		FROM base_username_change WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`, id, usernameHistoryMax)
	if err != nil {
		s.writeError(w, err)
		return
	}
	items, err := rawRows(rows)
	if err != nil {
		s.writeError(w, err)
		return
	}
	quote, err := s.usernameQuoteFor(r.Context(), id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "quote": quote})
}

func (s *service) playerUsernameRename(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if err := s.userExists(r.Context(), s.db, id); err != nil {
		s.writeError(w, err)
		return
	}
	var in struct {
		Username string `json:"username"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	who := consoleUser(r)
	if who == "" {
		who = "console"
	}
	out, err := s.changeUsername(r.Context(), usernameChangeIn{UserID: id, Name: in.Username, Admin: who})
	if err != nil {
		s.writeError(w, consoleRefusal(err))
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func consoleRefusal(err error) error {
	var rt *runtime.Error
	if !errors.As(err, &rt) {
		return err
	}
	var e socialError
	code := rt.Message
	if json.Unmarshal([]byte(rt.Message), &e) == nil && e.Code != "" {
		code = e.Code
	}
	status := http.StatusBadRequest
	switch codes.Code(rt.Code) {
	case codes.AlreadyExists:
		status = http.StatusConflict
	case codes.NotFound:
		status = http.StatusNotFound
	case codes.FailedPrecondition:
		status = http.StatusConflict
	}
	field := ""
	if strings.HasPrefix(code, "USERNAME") {
		field = "username"
	}
	return fail(status, code, field, e.Message)
}
