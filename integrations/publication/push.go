package main

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
)

const (
	fcmScope            = "https://www.googleapis.com/auth/firebase.messaging"
	pushDevicesPerUser  = 10
	pushRegisterPerMin  = 20
	sandboxDeadPrefix   = "sandbox-dead"
	pushDeadAfterErrors = 5
)

var pushPlatforms = []string{"android", "ios", "web"}

type serviceAccount struct {
	Type         string `json:"type"`
	ProjectID    string `json:"project_id"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
}

func parseServiceAccount(raw string) (serviceAccount, *rsa.PrivateKey, error) {
	var sa serviceAccount
	if err := json.Unmarshal([]byte(raw), &sa); err != nil {
		return sa, nil, errors.New("not valid JSON")
	}
	if sa.Type != "service_account" {
		return sa, nil, errors.New(`"type" must be "service_account"`)
	}
	if strings.TrimSpace(sa.ProjectID) == "" || strings.TrimSpace(sa.ClientEmail) == "" {
		return sa, nil, errors.New(`"project_id" and "client_email" are required`)
	}
	block, _ := pem.Decode([]byte(sa.PrivateKey))
	if block == nil {
		return sa, nil, errors.New(`"private_key" is not a PEM key`)
	}
	var key *rsa.PrivateKey
	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rk, ok := k.(*rsa.PrivateKey)
		if !ok {
			return sa, nil, errors.New(`"private_key" is not an RSA key`)
		}
		key = rk
	} else if rk, err2 := x509.ParsePKCS1PrivateKey(block.Bytes); err2 == nil {
		key = rk
	} else {
		return sa, nil, errors.New(`"private_key" cannot be read`)
	}
	return sa, key, nil
}

type fcmToken struct {
	value  string
	expiry time.Time
}

type tokenCache struct {
	mu   sync.Mutex
	byID map[string]fcmToken
}

var fcmTokens = &tokenCache{byID: map[string]fcmToken{}}

func cacheKey(p provider) string {
	sum := sha256.Sum256([]byte(p.ID + "\x00" + p.Config.str("serviceAccount")))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (c *tokenCache) get(key string, now time.Time) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	t, ok := c.byID[key]
	if !ok || now.After(t.expiry.Add(-time.Minute)) {
		return "", false
	}
	return t.value, true
}

func (c *tokenCache) put(key, value string, expiry time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.byID[key] = fcmToken{value: value, expiry: expiry}
}

func (c *tokenCache) drop(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.byID, key)
}

func signJWT(sa serviceAccount, key *rsa.PrivateKey, aud string, now time.Time) (string, error) {
	head, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT", "kid": sa.PrivateKeyID})
	claims, _ := json.Marshal(map[string]any{"iss": sa.ClientEmail, "scope": fcmScope, "aud": aud, "iat": now.Unix(), "exp": now.Add(time.Hour).Unix()})
	unsigned := base64.RawURLEncoding.EncodeToString(head) + "." + base64.RawURLEncoding.EncodeToString(claims)
	sum := sha256.Sum256([]byte(unsigned))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (s *service) fcmAccessToken(ctx context.Context, p provider, sa serviceAccount, key *rsa.PrivateKey) (string, error) {
	ck := cacheKey(p)
	now := s.now()
	if tok, ok := fcmTokens.get(ck, now); ok {
		return tok, nil
	}
	tokenURL := s.baseURLs["fcmauth"]
	assertion, err := signJWT(sa, key, "https://oauth2.googleapis.com/token", now)
	if err != nil {
		return "", permanentErr("cannot sign the FCM token request: %v", err)
	}
	form := url.Values{"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"}, "assertion": {assertion}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", permanentErr("bad token URL")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	status, body, err := s.doHTTP(req, "fcm token")
	if err != nil {
		return "", err
	}
	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
	}
	_ = json.Unmarshal(body, &out)
	if status != http.StatusOK || out.AccessToken == "" {
		if transientStatus(status) {
			return "", transientErr("fcm token request answered %d", status)
		}
		return "", permanentErr("fcm token request refused (%d %s); check the service account", status, out.Error)
	}
	if out.ExpiresIn <= 0 {
		out.ExpiresIn = 3600
	}
	fcmTokens.put(ck, out.AccessToken, now.Add(time.Duration(out.ExpiresIn)*time.Second))
	return out.AccessToken, nil
}

type pushMessage struct {
	Token    string
	Platform string
	Title    string
	Body     string
	Image    string
	Link     string
	Data     map[string]string
}

type pushResult struct {
	Dead  bool
	Reply string
}

func maskToken(t string) string {
	if len(t) <= 12 {
		return "****"
	}
	return t[:6] + "…" + t[len(t)-4:]
}

func (s *service) sendPush(ctx context.Context, p provider, m pushMessage) (pushResult, error) {
	if p.Kind == "log" || p.Sandbox {
		if strings.HasPrefix(m.Token, sandboxDeadPrefix) {
			s.logger.Info("Publication push sandbox via %s: token %s is unregistered (simulated)", p.Kind, maskToken(m.Token))
			return pushResult{Dead: true, Reply: "sandbox: UNREGISTERED"}, permanentErr("UNREGISTERED (sandbox)")
		}
		s.logger.Info("Publication push sandbox via %s to %s (%s): %s — %s", p.Kind, maskToken(m.Token), m.Platform, truncate(m.Title, 80), truncate(m.Body, 160))
		return pushResult{Reply: "sandbox"}, nil
	}
	if p.Kind != "fcm" {
		return pushResult{}, permanentErr("unknown push provider kind %q", p.Kind)
	}
	return s.sendFCM(ctx, p, m, true)
}

func (s *service) sendFCM(ctx context.Context, p provider, m pushMessage, retryAuth bool) (pushResult, error) {
	sa, key, err := parseServiceAccount(p.Config.str("serviceAccount"))
	if err != nil {
		return pushResult{}, permanentErr("service account: %v", err)
	}
	tok, err := s.fcmAccessToken(ctx, p, sa, key)
	if err != nil {
		return pushResult{}, err
	}
	msg := map[string]any{"token": m.Token}
	note := map[string]any{"title": m.Title, "body": m.Body}
	if m.Image != "" && strings.HasPrefix(strings.ToLower(m.Image), "https://") {
		note["image"] = m.Image
	}
	msg["notification"] = note
	if len(m.Data) > 0 {
		msg["data"] = m.Data
	}
	msg["android"] = map[string]any{"priority": "HIGH"}
	apns := map[string]any{"payload": map[string]any{"aps": map[string]any{"mutable-content": 1}}}
	if img, ok := note["image"]; ok {
		apns["fcm_options"] = map[string]any{"image": img}
	}
	msg["apns"] = apns
	if m.Link != "" && strings.HasPrefix(strings.ToLower(m.Link), "https://") {
		msg["webpush"] = map[string]any{"fcm_options": map[string]any{"link": m.Link}}
	}
	raw, _ := json.Marshal(map[string]any{"message": msg})
	endpoint := strings.TrimRight(s.baseURLs["fcm"], "/") + "/v1/projects/" + url.PathEscape(sa.ProjectID) + "/messages:send"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return pushResult{}, permanentErr("bad FCM URL")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	status, body, err := s.doHTTP(req, "fcm")
	if err != nil {
		return pushResult{}, err
	}
	if status == http.StatusOK {
		var ok struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(body, &ok)
		return pushResult{Reply: ok.Name}, nil
	}
	var fe struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
			Details []struct {
				Type      string `json:"@type"`
				ErrorCode string `json:"errorCode"`
			} `json:"details"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &fe)
	code := fe.Error.Status
	for _, d := range fe.Error.Details {
		if d.ErrorCode != "" {
			code = d.ErrorCode
		}
	}
	if status == http.StatusUnauthorized && retryAuth {
		fcmTokens.drop(cacheKey(p))
		return s.sendFCM(ctx, p, m, false)
	}
	switch {
	case code == "UNREGISTERED" || code == "SENDER_ID_MISMATCH" || (code == "INVALID_ARGUMENT" && strings.Contains(strings.ToLower(fe.Error.Message), "registration token")):
		return pushResult{Dead: true, Reply: code}, permanentErr("fcm: %s", code)
	case transientStatus(status) || code == "UNAVAILABLE" || code == "INTERNAL" || code == "QUOTA_EXCEEDED":
		return pushResult{}, transientErr("fcm answered %d %s", status, code)
	}
	return pushResult{}, permanentErr("fcm answered %d %s", status, code)
}

type pushDevice struct {
	ID         string
	UserID     string
	Token      string
	Platform   string
	AppVersion string
}

func (s *service) userDevices(ctx context.Context, ids []string) (map[string][]pushDevice, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, user_id::text, token, platform, app_version FROM publication_push_device WHERE user_id = ANY($1::uuid[]) AND is_active ORDER BY user_id, last_seen_at DESC`, uuidArray(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]pushDevice{}
	for rows.Next() {
		var d pushDevice
		if err := rows.Scan(&d.ID, &d.UserID, &d.Token, &d.Platform, &d.AppVersion); err != nil {
			return nil, err
		}
		out[d.UserID] = append(out[d.UserID], d)
	}
	return out, rows.Err()
}

func (s *service) claimDelivery(ctx context.Context, notificationID string, d pushDevice, providerID string) (bool, error) {
	var got string
	err := s.db.QueryRowContext(ctx, `INSERT INTO publication_push_delivery (notification_id, device_id, user_id, status, provider_id) VALUES ($1,$2,$3,'sending',$4)
		ON CONFLICT (notification_id, device_id) DO UPDATE SET status='sending', updated_at=now(), provider_id=EXCLUDED.provider_id WHERE publication_push_delivery.status='failed'
		RETURNING device_id::text`, notificationID, d.ID, d.UserID, providerID).Scan(&got)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (s *service) finishDelivery(ctx context.Context, notificationID string, d pushDevice, status, errText string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE publication_push_delivery SET status=$3, error=$4, updated_at=now() WHERE notification_id=$1 AND device_id=$2`, notificationID, d.ID, status, truncate(errText, 300))
	if err != nil {
		return err
	}
	switch status {
	case "dead":
		_, err = s.db.ExecContext(ctx, `UPDATE publication_push_device SET is_active=false, disabled_at=now(), last_error=$2, fail_count=fail_count+1, updated_at=now() WHERE id=$1`, d.ID, truncate(errText, 300))
	case "failed":
		_, err = s.db.ExecContext(ctx, `UPDATE publication_push_device SET fail_count=fail_count+1, last_error=$2, updated_at=now(),
			is_active = is_active AND fail_count+1 < $3, disabled_at = CASE WHEN fail_count+1 >= $3 THEN now() ELSE disabled_at END WHERE id=$1`, d.ID, truncate(errText, 300), pushDeadAfterErrors)
	case "sent":
		_, err = s.db.ExecContext(ctx, `UPDATE publication_push_device SET fail_count=0, last_error='' WHERE id=$1 AND fail_count > 0`, d.ID)
	}
	return err
}

func cleanPushToken(raw string) (string, error) {
	tok := strings.TrimSpace(raw)
	if len(tok) < 16 || len(tok) > 4096 || strings.ContainsAny(tok, " \t\r\n\"'<>`") {
		return "", badRequest("INVALID_TOKEN", "token", "push token must be 16 to 4096 characters without spaces")
	}
	return tok, nil
}

func (s *service) rpcPushRegister(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Token      string `json:"token"`
		Platform   string `json:"platform"`
		AppVersion string `json:"appVersion"`
		DeviceID   string `json:"deviceId"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	tok, err := cleanPushToken(in.Token)
	if err != nil {
		return nil, err
	}
	platform := strings.ToLower(strings.TrimSpace(in.Platform))
	if !containsStr(pushPlatforms, platform) {
		return nil, badRequest("INVALID_PLATFORM", "platform", "platform must be android, ios or web")
	}
	_, version, err := cleanClientTarget("", in.AppVersion)
	if err != nil {
		return nil, err
	}
	devKey := strings.TrimSpace(in.DeviceID)
	if len(devKey) > 128 {
		return nil, badRequest("INVALID_DEVICE", "deviceId", "device id is too long (max 128)")
	}
	if err := s.bannerLimit(ctx, "push:r:"+userID, pushRegisterPerMin); err != nil {
		return nil, err
	}
	var id string
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('push-device:' || $1))`, userID); err != nil {
			return err
		}
		if err := tx.QueryRowContext(ctx, `INSERT INTO publication_push_device (id, user_id, token, platform, app_version, device_key) VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (token) DO UPDATE SET user_id=EXCLUDED.user_id, platform=EXCLUDED.platform, app_version=EXCLUDED.app_version, device_key=EXCLUDED.device_key,
				is_active=true, fail_count=0, last_error='', disabled_at=NULL, last_seen_at=now(), updated_at=now()
			RETURNING id::text`, uuid.Must(uuid.NewV4()).String(), userID, tok, platform, version, devKey).Scan(&id); err != nil {
			return err
		}
		if devKey != "" {
			if _, err := tx.ExecContext(ctx, `UPDATE publication_push_device SET is_active=false, disabled_at=now(), last_error='replaced by a newer token', updated_at=now()
				WHERE user_id=$1 AND device_key=$2 AND id<>$3 AND is_active`, userID, devKey, id); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(ctx, `UPDATE publication_push_device SET is_active=false, disabled_at=now(), last_error='too many devices', updated_at=now()
			WHERE id IN (SELECT id FROM publication_push_device WHERE user_id=$1 AND is_active ORDER BY last_seen_at DESC, created_at DESC OFFSET $2)`, userID, pushDevicesPerUser)
		return err
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"deviceId": id, "platform": platform, "active": true}, nil
}

func (s *service) rpcPushUnregister(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Token string `json:"token"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	tok, err := cleanPushToken(in.Token)
	if err != nil {
		return nil, err
	}
	if err := s.bannerLimit(ctx, "push:r:"+userID, pushRegisterPerMin); err != nil {
		return nil, err
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM publication_push_device WHERE token=$1 AND user_id=$2`, tok, userID)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	return map[string]any{"removed": n == 1}, nil
}

func (s *service) rpcPushDevices(ctx context.Context, userID, payload string) (any, error) {
	var in struct{}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, token, platform, app_version, is_active, last_seen_at FROM publication_push_device WHERE user_id=$1 ORDER BY is_active DESC, last_seen_at DESC LIMIT 20`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, tok, platform, version string
		var active bool
		var seen time.Time
		if err := rows.Scan(&id, &tok, &platform, &version, &active, &seen); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"id": id, "token": maskToken(tok), "platform": platform, "appVersion": version, "active": active, "lastSeenAt": seen.UTC()})
	}
	return map[string]any{"items": items}, rows.Err()
}

func (s *service) sweepPushDevices(ctx context.Context) (int64, error) {
	var total int64
	for _, q := range []string{
		`DELETE FROM publication_push_device WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = publication_push_device.user_id)`,
		`DELETE FROM publication_push_device WHERE NOT is_active AND disabled_at < now() - interval '30 days'`,
		`DELETE FROM publication_notify_pref p WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = p.user_id)`,
		`DELETE FROM publication_auto_fire WHERE created_at < now() - interval '90 days' AND status <> 'pending'`,
		`DELETE FROM publication_auto_fire f WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = f.user_id)`,
	} {
		res, err := s.db.ExecContext(ctx, q)
		if err != nil {
			return total, err
		}
		n, _ := res.RowsAffected()
		total += n
	}
	return total, nil
}
