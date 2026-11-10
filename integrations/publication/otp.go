package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/restrict"
)

const (
	otpLoginKeyPrefix = "ph_"
	otpMessageMax     = 300
	otpKeepCodes      = 30 * 24 * time.Hour
)

var otpStatuses = []string{"pending", "verified", "expired", "failed", "replaced", "send_failed"}

type otpSettings struct {
	IsEnabled     bool      `json:"isEnabled"`
	Sandbox       bool      `json:"sandbox"`
	ProviderID    string    `json:"providerId"`
	CodeLength    int       `json:"codeLength"`
	TTLSeconds    int       `json:"ttlSeconds"`
	MaxAttempts   int       `json:"maxAttempts"`
	ResendSeconds int       `json:"resendSeconds"`
	PhoneHourly   int       `json:"phoneHourly"`
	PhoneDaily    int       `json:"phoneDaily"`
	IPHourly      int       `json:"ipHourly"`
	Message       i18n      `json:"message"`
	UpdatedBy     string    `json:"updatedBy,omitempty"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func defaultOTPSettings() otpSettings {
	return otpSettings{IsEnabled: true, Sandbox: true, CodeLength: 5, TTLSeconds: 120, MaxAttempts: 5, ResendSeconds: 60, PhoneHourly: 5, PhoneDaily: 10, IPHourly: 30,
		Message: i18n{"fa": "کد ورود شما: {code}\nاین کد تا {minutes} دقیقه معتبر است.", "en": "Your login code: {code}\nIt is valid for {minutes} minutes."}}
}

func loadOTPSettings(ctx context.Context, q queryer) (otpSettings, error) {
	st := defaultOTPSettings()
	var provider sql.NullString
	var msg string
	err := q.QueryRowContext(ctx, `SELECT is_enabled, sandbox, provider_id::text, code_length, ttl_seconds, max_attempts, resend_seconds, phone_hourly, phone_daily, ip_hourly, message::text, updated_by, updated_at
		FROM publication_otp_settings WHERE id=1`).Scan(&st.IsEnabled, &st.Sandbox, &provider, &st.CodeLength, &st.TTLSeconds, &st.MaxAttempts, &st.ResendSeconds,
		&st.PhoneHourly, &st.PhoneDaily, &st.IPHourly, &msg, &st.UpdatedBy, &st.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultOTPSettings(), nil
	}
	if err != nil {
		return st, err
	}
	st.ProviderID = provider.String
	st.Message = parseI18n(msg)
	return st, nil
}

func between(field string, v, lo, hi int) error {
	if v < lo || v > hi {
		return badRequest("INVALID_VALUE", field, "must be between "+strconv.Itoa(lo)+" and "+strconv.Itoa(hi))
	}
	return nil
}

func (s *service) cleanOTPSettings(ctx context.Context, in otpSettings) (otpSettings, error) {
	checks := []error{
		between("codeLength", in.CodeLength, 4, 8),
		between("ttlSeconds", in.TTLSeconds, 30, 1800),
		between("maxAttempts", in.MaxAttempts, 1, 10),
		between("resendSeconds", in.ResendSeconds, 10, 600),
		between("phoneHourly", in.PhoneHourly, 1, 100),
		between("phoneDaily", in.PhoneDaily, 1, 500),
		between("ipHourly", in.IPHourly, 1, 10000),
	}
	for _, err := range checks {
		if err != nil {
			return in, err
		}
	}
	if in.PhoneDaily < in.PhoneHourly {
		return in, badRequest("INVALID_VALUE", "phoneDaily", "the daily limit cannot be lower than the hourly limit")
	}
	in.ProviderID = strings.TrimSpace(in.ProviderID)
	if in.ProviderID != "" {
		p, err := s.loadProvider(ctx, s.db, in.ProviderID, false)
		if err != nil {
			return in, badRequest("PROVIDER_NOT_FOUND", "providerId", "choose an SMS service from the list")
		}
		if p.Channel != "sms" {
			return in, badRequest("PROVIDER_NOT_SMS", "providerId", "choose an SMS service")
		}
	}
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		return in, err
	}
	if in.Message, err = cleanI18n(langs, "message", in.Message, true, otpMessageMax); err != nil {
		return in, err
	}
	for _, v := range in.Message {
		if !strings.Contains(v, "{code}") {
			return in, badRequest("OTP_MESSAGE_NEEDS_CODE", "message", "every translation must contain {code}")
		}
	}
	return in, nil
}

func (s *service) smsProviderOptions(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, kind, is_active, is_default, sandbox FROM publication_provider WHERE channel='sms' ORDER BY is_default DESC, kind`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, kind string
		var active, def, sandbox bool
		if err := rows.Scan(&id, &kind, &active, &def, &sandbox); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "kind": kind, "isActive": active, "isDefault": def, "sandbox": sandbox})
	}
	return out, rows.Err()
}

func (s *service) consoleOTPSettingsView(ctx context.Context) (map[string]any, error) {
	st, err := loadOTPSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	providers, err := s.smsProviderOptions(ctx)
	if err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(st)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	out["providers"] = providers
	return out, nil
}

func (s *service) consoleOTPSettingsGet(w http.ResponseWriter, r *http.Request) {
	out, err := s.consoleOTPSettingsView(r.Context())
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) consoleOTPSettingsPut(w http.ResponseWriter, r *http.Request) {
	in := defaultOTPSettings()
	in.Message = nil
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	st, err := s.cleanOTPSettings(r.Context(), in)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var provider any
	if st.ProviderID != "" {
		provider = st.ProviderID
	}
	_, err = s.db.ExecContext(r.Context(), `INSERT INTO publication_otp_settings (id, is_enabled, sandbox, provider_id, code_length, ttl_seconds, max_attempts, resend_seconds, phone_hourly, phone_daily, ip_hourly, message, updated_by, updated_at)
		VALUES (1,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,now())
		ON CONFLICT (id) DO UPDATE SET is_enabled=EXCLUDED.is_enabled, sandbox=EXCLUDED.sandbox, provider_id=EXCLUDED.provider_id, code_length=EXCLUDED.code_length,
		ttl_seconds=EXCLUDED.ttl_seconds, max_attempts=EXCLUDED.max_attempts, resend_seconds=EXCLUDED.resend_seconds, phone_hourly=EXCLUDED.phone_hourly,
		phone_daily=EXCLUDED.phone_daily, ip_hourly=EXCLUDED.ip_hourly, message=EXCLUDED.message, updated_by=EXCLUDED.updated_by, updated_at=now()`,
		st.IsEnabled, st.Sandbox, provider, st.CodeLength, st.TTLSeconds, st.MaxAttempts, st.ResendSeconds, st.PhoneHourly, st.PhoneDaily, st.IPHourly, st.Message.JSON(), consoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleOTPSettingsGet(w, r)
}

const otpRequestJSON = `jsonb_build_object('id',o.id,'phone',o.phone,'status',CASE WHEN o.status='pending' AND o.expires_at <= now() THEN 'expired' ELSE o.status END,
	'attempts',o.attempts,'maxAttempts',o.max_attempts,'sandbox',o.sandbox,'provider',o.provider,'lang',o.lang,'error',o.error,'clientIp',o.client_ip,
	'userId',o.user_id,'username',u.username,'createdAt',o.created_at,'expiresAt',o.expires_at,'verifiedAt',o.verified_at)`

func (s *service) consoleOTPRequests(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		q := strings.TrimLeft(strings.NewReplacer(" ", "", "-", "").Replace(latinDigits(p.Q)), "+0")
		where.add("o.phone ILIKE ?", likePattern(q))
	}
	switch st := r.URL.Query().Get("status"); st {
	case "":
	case "expired":
		where.add("(o.status = 'expired' OR (o.status = 'pending' AND o.expires_at <= now()))")
	case "pending":
		where.add("o.status = 'pending' AND o.expires_at > now()")
	default:
		ok := false
		for _, v := range otpStatuses {
			ok = ok || v == st
		}
		if !ok {
			s.writeError(w, badRequest("INVALID_FILTER", "status", "unknown status"))
			return
		}
		where.add("o.status = ?", st)
	}
	from := ` FROM publication_otp_code o LEFT JOIN users u ON u.id = o.user_id`
	var total int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*)`+from+where.sql(), where.args...).Scan(&total); err != nil {
		s.writeError(w, err)
		return
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+otpRequestJSON+`::text`+from+where.sql()+` ORDER BY o.created_at DESC, o.id LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2),
		append(append([]any{}, where.args...), p.Limit, p.Offset)...)
	if err != nil {
		s.writeError(w, err)
		return
	}
	items, err := rawRows(rows)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var accounts, today int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT (SELECT count(*) FROM publication_phone_account), (SELECT count(*) FROM publication_otp_code WHERE created_at > now() - interval '24 hours')`).Scan(&accounts, &today); err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": p.Page, "limit": p.Limit, "phoneAccounts": accounts, "last24h": today})
}

func randomDigits(n int) (string, error) {
	var b strings.Builder
	for i := 0; i < n; i++ {
		d, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b.WriteByte(byte('0' + d.Int64()))
	}
	return b.String(), nil
}

func newLoginKey() (string, error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return otpLoginKeyPrefix + hex.EncodeToString(raw), nil
}

func (s *service) otpHash(id, phone, code string) string {
	mac := hmac.New(sha256.New, s.cfg.otpKey)
	mac.Write([]byte(id + "|" + phone + "|" + code))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *service) otpPhone(ctx context.Context, rawPhone, country string) (string, error) {
	raw := strings.TrimSpace(latinDigits(rawPhone))
	country = strings.ToUpper(strings.TrimSpace(country))
	compact := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", ".", "").Replace(raw)
	if country != "" && country != "IR" && !strings.HasPrefix(compact, "+") && !strings.HasPrefix(compact, "00") && compact != "" {
		var dial string
		err := s.db.QueryRowContext(ctx, `SELECT dial_code FROM base_country WHERE code=$1 AND is_active`, country).Scan(&dial)
		if errors.Is(err, sql.ErrNoRows) {
			return "", badRequest("PHONE_COUNTRY_NOT_ALLOWED", "country", "this country is not available")
		}
		if err != nil {
			return "", err
		}
		compact = dial + strings.TrimLeft(compact, "0")
	}
	phone, ok := normalizePhone(compact)
	if !ok {
		return "", badRequest("INVALID_PHONE", "phone", "enter a mobile number")
	}
	var allowed bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_country WHERE is_active AND dial_code <> '' AND $1 LIKE dial_code || '%')`, phone).Scan(&allowed); err != nil {
		return "", err
	}
	if !allowed {
		return "", badRequest("PHONE_COUNTRY_NOT_ALLOWED", "phone", "numbers from this country are not accepted")
	}
	return phone, nil
}

func clientIP(ctx context.Context) string {
	ip, _ := ctx.Value(runtime.RUNTIME_CTX_CLIENT_IP).(string)
	return strings.TrimSpace(ip)
}

func (s *service) otpLimit(ctx context.Context, key string, limit int, window time.Duration, code string) error {
	d, err := clusterkit.Allow(ctx, s.db, key, int64(limit), window)
	if err != nil {
		return err
	}
	if !d.Allowed {
		return fail(http.StatusTooManyRequests, code, "phone", "try again in "+strconv.Itoa(int(d.RetryAfter.Seconds())+1)+" seconds")
	}
	return nil
}

func (s *service) appName(ctx context.Context, langs *localized.Set, lang string) string {
	var name, short string
	if err := s.db.QueryRowContext(ctx, `SELECT name::text, short_name FROM base_brand WHERE id=1`).Scan(&name, &short); err != nil {
		return ""
	}
	if v := langs.Resolve(parseI18n(name), lang); v != "" {
		return v
	}
	return short
}

func otpText(template, code, app string, ttl int) string {
	minutes := (ttl + 59) / 60
	return strings.NewReplacer("{code}", code, "{minutes}", strconv.Itoa(minutes), "{app}", app).Replace(template)
}

func (s *service) rpcOTPRequest(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Phone   string `json:"phone"`
		Country string `json:"country"`
		Lang    string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	st, err := loadOTPSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if !st.IsEnabled {
		return nil, fail(http.StatusServiceUnavailable, "OTP_DISABLED", "", "SMS login is turned off")
	}
	phone, err := s.otpPhone(ctx, in.Phone, in.Country)
	if err != nil {
		return nil, err
	}
	ip := clientIP(ctx)
	if err := s.otpLimit(ctx, "otp:resend:"+phone, 1, time.Duration(st.ResendSeconds)*time.Second, "OTP_TOO_SOON"); err != nil {
		return nil, err
	}
	if err := s.otpLimit(ctx, "otp:phone:h:"+phone, st.PhoneHourly, time.Hour, "OTP_RATE_LIMITED"); err != nil {
		return nil, err
	}
	if err := s.otpLimit(ctx, "otp:phone:d:"+phone, st.PhoneDaily, 24*time.Hour, "OTP_RATE_LIMITED"); err != nil {
		return nil, err
	}
	if ip != "" {
		if err := s.otpLimit(ctx, "otp:ip:h:"+ip, st.IPHourly, time.Hour, "OTP_RATE_LIMITED"); err != nil {
			return nil, err
		}
	}
	if err := s.otpLimit(ctx, "otp:user:h:"+userID, st.IPHourly, time.Hour, "OTP_RATE_LIMITED"); err != nil {
		return nil, err
	}
	var p provider
	var haveProvider bool
	if st.ProviderID != "" {
		p, err = s.loadProvider(ctx, s.db, st.ProviderID, false)
		haveProvider = err == nil && p.Active
	}
	if !haveProvider {
		p, err = s.defaultProvider(ctx, s.db, "sms")
		haveProvider = err == nil
	}
	sandbox := st.Sandbox || (haveProvider && p.Sandbox)
	if !sandbox && !haveProvider {
		return nil, fail(http.StatusServiceUnavailable, "OTP_UNAVAILABLE", "", "no SMS service is set up")
	}
	langs, lang, err := s.clientLang(ctx, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	code, err := randomDigits(st.CodeLength)
	if err != nil {
		return nil, err
	}
	id := uuid.Must(uuid.NewV4()).String()
	kind := "sandbox"
	if haveProvider {
		kind = p.Kind
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `UPDATE publication_otp_code SET status='replaced' WHERE phone=$1 AND status='pending'`, phone); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO publication_otp_code (id, phone, code_hash, max_attempts, sandbox, provider, lang, client_ip, requested_by, expires_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9, now() + $10::INT8 * INTERVAL '1 second')`,
			id, phone, s.otpHash(id, phone, code), st.MaxAttempts, sandbox, kind, lang, ip, userID, st.TTLSeconds)
		return err
	})
	if err != nil {
		return nil, err
	}
	text := otpText(langs.Resolve(st.Message, lang), code, s.appName(ctx, langs, lang), st.TTLSeconds)
	if sandbox {
		s.logger.Info("Publication OTP sandbox code %s for %s (request %s, %s)", code, phone, id, lang)
	} else {
		sendCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
		reply, err := s.deliver(sendCtx, p, outMessage{To: phone, Text: text, Lang: lang, IdempotencyKey: "otp-" + id})
		cancel()
		if err != nil {
			_, _ = s.db.ExecContext(ctx, `UPDATE publication_otp_code SET status='send_failed', error=$2 WHERE id=$1`, id, truncate(err.Error(), 500))
			s.logger.Warn("Publication OTP send via %s to %s failed: %v", p.Kind, maskAddress(phone), err)
			return nil, fail(http.StatusBadGateway, "OTP_SEND_FAILED", "", "the code could not be sent; try again")
		}
		s.logger.Info("Publication OTP sent via %s to %s: %s", p.Kind, maskAddress(phone), truncate(reply, 200))
	}
	return map[string]any{"requestId": id, "phone": phone, "expiresIn": st.TTLSeconds, "resendAfter": st.ResendSeconds, "codeLength": st.CodeLength, "sandbox": sandbox}, nil
}

type otpRow struct {
	ID          string
	Hash        string
	Status      string
	Attempts    int
	MaxAttempts int
	Expired     bool
}

type otpLogin struct {
	UserID   string
	LoginKey string
	Created  bool
	Linked   bool
}

func (s *service) checkOTP(ctx context.Context, tx *sql.Tx, phone, code string) error {
	var row otpRow
	err := tx.QueryRowContext(ctx, `SELECT id::text, code_hash, status, attempts, max_attempts, expires_at <= now() FROM publication_otp_code WHERE phone=$1 ORDER BY created_at DESC LIMIT 1 FOR UPDATE`, phone).
		Scan(&row.ID, &row.Hash, &row.Status, &row.Attempts, &row.MaxAttempts, &row.Expired)
	if errors.Is(err, sql.ErrNoRows) {
		return badRequest("OTP_NOT_FOUND", "code", "ask for a new code")
	}
	if err != nil {
		return err
	}
	switch {
	case row.Status == "failed":
		return badRequest("OTP_TOO_MANY_ATTEMPTS", "code", "too many wrong codes; ask for a new one")
	case row.Status != "pending":
		return badRequest("OTP_NOT_FOUND", "code", "ask for a new code")
	case row.Expired:
		_, err := tx.ExecContext(ctx, `UPDATE publication_otp_code SET status='expired' WHERE id=$1`, row.ID)
		if err != nil {
			return err
		}
		return &otpVerdict{badRequest("OTP_EXPIRED", "code", "the code has expired; ask for a new one")}
	}
	if !hmac.Equal([]byte(s.otpHash(row.ID, phone, code)), []byte(row.Hash)) {
		row.Attempts++
		status := "pending"
		if row.Attempts >= row.MaxAttempts {
			status = "failed"
		}
		if _, err := tx.ExecContext(ctx, `UPDATE publication_otp_code SET attempts=$2, status=$3 WHERE id=$1`, row.ID, row.Attempts, status); err != nil {
			return err
		}
		if status == "failed" {
			return &otpVerdict{badRequest("OTP_TOO_MANY_ATTEMPTS", "code", "too many wrong codes; ask for a new one")}
		}
		return &otpVerdict{badRequest("OTP_INVALID_CODE", "code", strconv.Itoa(row.MaxAttempts-row.Attempts)+" tries left")}
	}
	_, err = tx.ExecContext(ctx, `UPDATE publication_otp_code SET status='verified', attempts=attempts+1, verified_at=now() WHERE id=$1`, row.ID)
	return err
}

type otpVerdict struct{ err *appError }

func (v *otpVerdict) Error() string { return v.err.Error() }

func (s *service) otpAccount(ctx context.Context, tx *sql.Tx, phone, caller string, link bool) (otpLogin, error) {
	var out otpLogin
	err := tx.QueryRowContext(ctx, `SELECT user_id::text, login_key FROM publication_phone_account WHERE phone=$1`, phone).Scan(&out.UserID, &out.LoginKey)
	if err == nil {
		var disabled bool
		if err := tx.QueryRowContext(ctx, `SELECT disable_time > '1970-01-01 00:00:00+00' FROM users WHERE id=$1`, out.UserID).Scan(&disabled); err != nil {
			return out, err
		}
		if disabled {
			return out, fail(http.StatusForbidden, "ACCOUNT_DISABLED", "", "this account is disabled")
		}
		_, err = tx.ExecContext(ctx, `UPDATE publication_phone_account SET verified_at=now() WHERE phone=$1`, phone)
		out.Linked = out.UserID == caller
		return out, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return out, err
	}
	if link {
		var hasPhone bool
		var customID string
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_phone_account WHERE user_id=$1), coalesce(custom_id,'') FROM users WHERE id=$1`, caller).Scan(&hasPhone, &customID); err != nil {
			return out, err
		}
		if hasPhone {
			return out, fail(http.StatusConflict, "ACCOUNT_HAS_PHONE", "link", "this account already has a phone number")
		}
		out.UserID, out.Linked = caller, true
		if customID == "" {
			key, err := newLoginKey()
			if err != nil {
				return out, err
			}
			if err := s.nk.LinkCustom(ctx, caller, key); err != nil {
				return out, err
			}
			out.LoginKey = key
		}
	} else {
		key, err := newLoginKey()
		if err != nil {
			return out, err
		}
		uid, _, created, err := s.nk.AuthenticateCustom(ctx, key, "", true)
		if err != nil {
			return out, err
		}
		out.UserID, out.LoginKey, out.Created = uid, key, created
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO publication_phone_account (phone, user_id, login_key) VALUES ($1,$2,$3)`, phone, out.UserID, out.LoginKey); err != nil {
		return out, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE users SET metadata = jsonb_set(coalesce(metadata,'{}'::jsonb), '{phone}', to_jsonb($2::text)), update_time=now() WHERE id=$1`, out.UserID, phone)
	return out, err
}

func (s *service) rpcOTPVerify(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Phone   string `json:"phone"`
		Country string `json:"country"`
		Code    string `json:"code"`
		Link    bool   `json:"link"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	st, err := loadOTPSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if !st.IsEnabled {
		return nil, fail(http.StatusServiceUnavailable, "OTP_DISABLED", "", "SMS login is turned off")
	}
	phone, err := s.otpPhone(ctx, in.Phone, in.Country)
	if err != nil {
		return nil, err
	}
	code := strings.TrimSpace(latinDigits(in.Code))
	if code == "" || len(code) > 8 || strings.Trim(code, "0123456789") != "" {
		return nil, badRequest("OTP_INVALID_CODE", "code", "enter the digits from the SMS")
	}
	var login otpLogin
	var verdict *otpVerdict
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('publication-otp:' || $1))`, phone); err != nil {
			return err
		}
		if err := s.checkOTP(ctx, tx, phone, code); err != nil {
			if errors.As(err, &verdict) {
				return nil
			}
			return err
		}
		var err error
		login, err = s.otpAccount(ctx, tx, phone, userID, in.Link)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE publication_otp_code SET user_id=$2 WHERE phone=$1 AND status='verified' AND user_id IS NULL`, phone, login.UserID)
		return err
	})
	if err != nil {
		return nil, err
	}
	if verdict != nil {
		return nil, verdict.err
	}
	if err := restrict.RefuseSuspended(ctx, s.db, login.UserID, ""); err != nil {
		return nil, err
	}
	var username string
	if err := s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id=$1`, login.UserID).Scan(&username); err != nil {
		return nil, err
	}
	token, exp, err := s.nk.AuthenticateTokenGenerate(login.UserID, username, 0, nil)
	if err != nil {
		return nil, err
	}
	return map[string]any{"token": token, "expiresAt": exp, "loginKey": login.LoginKey, "userId": login.UserID, "username": username,
		"created": login.Created, "linked": login.Linked, "phone": phone}, nil
}

func (s *service) rpcPhone(ctx context.Context, userID, payload string) (any, error) {
	var phone string
	err := s.db.QueryRowContext(ctx, `SELECT phone FROM publication_phone_account WHERE user_id=$1`, userID).Scan(&phone)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	st, err := loadOTPSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	return map[string]any{"phone": phone, "otpEnabled": st.IsEnabled, "codeLength": st.CodeLength}, nil
}

func (s *service) otpSweepLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ok, err := clusterkit.AcquireLease(ctx, s.db, "publication-otp-sweep", holder, 9*time.Minute)
			if err != nil || !ok {
				continue
			}
			if _, err := s.db.ExecContext(ctx, `DELETE FROM publication_otp_code WHERE created_at < now() - $1::INT8 * INTERVAL '1 second'`, int64(otpKeepCodes.Seconds())); err != nil && ctx.Err() == nil {
				s.logger.Error("Publication OTP sweep failed: %v", err)
			}
			if _, err := clusterkit.SweepRateLimits(ctx, s.db, time.Hour); err != nil && ctx.Err() == nil {
				s.logger.Error("Rate limit sweep failed: %v", err)
			}
		}
	}
}
