package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

type resultTemplate struct {
	ID         string `json:"id"`
	Key        string `json:"key"`
	Name       string `json:"name"`
	MissionID  string `json:"missionId"`
	MissionKey string `json:"missionKey"`
	Body       string `json:"body"`
	WonText    string `json:"wonText"`
	LostText   string `json:"lostText"`
	WebhookURL string `json:"webhookUrl"`
	HasSecret  bool   `json:"hasSecret"`
	Active     bool   `json:"isActive"`
	UpdatedAt  string `json:"updatedAt"`
	secret     string
}

const templateCols = `t.id::text, t.key, t.name, COALESCE(t.mission_id::text,''), COALESCE(m.key,''), t.body, t.won_text, t.lost_text, t.webhook_url, t.webhook_secret, t.is_active, t.updated_at`

const templateFrom = ` FROM quiz_result_template t LEFT JOIN quiz_mission m ON m.id=t.mission_id`

func scanTemplate(sc scanner) (*resultTemplate, error) {
	t := &resultTemplate{}
	var at sql.NullTime
	if err := sc.Scan(&t.ID, &t.Key, &t.Name, &t.MissionID, &t.MissionKey, &t.Body, &t.WonText, &t.LostText, &t.WebhookURL, &t.secret, &t.Active, &at); err != nil {
		return nil, err
	}
	t.HasSecret = t.secret != ""
	t.UpdatedAt = stamp(at)
	return t, nil
}

func num(f float64) string {
	return strconv.FormatFloat(math.Round(f*100)/100, 'f', -1, 64)
}

func moment(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04")
}

func (s *service) resultValues(ctx context.Context, t *resultTemplate, userID string) (map[string]string, error) {
	v := map[string]string{"user.id": userID}
	var username, display sql.NullString
	_ = s.db.QueryRowContext(ctx, `SELECT username, display_name FROM users WHERE id=$1`, userID).Scan(&username, &display)
	v["user.username"], v["user.displayName"] = username.String, display.String
	w, err := loadWorld(ctx, s.db)
	if err != nil {
		return nil, err
	}
	p, err := loadProgress(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}
	langs, lang, _ := localized.ForUser(ctx, s.db, userID, "")
	m := w.missions[t.MissionID]
	var missions []*mission
	var packs []*pack
	if m != nil {
		if m.Kind == "game" {
			missions = w.missionsOf(m.ID)
		} else {
			missions = []*mission{m}
		}
		for _, x := range missions {
			packs = append(packs, w.packsOf(x.ID)...)
		}
		u := p.missions[m.ID]
		v["mission.name"], v["mission.description"] = text(langs, lang, m.Name), text(langs, lang, m.Description)
		if m.TimeSeconds > 0 {
			v["mission.time"] = strconv.Itoa(m.TimeSeconds)
		}
		if u.hasRow {
			v["mission.startedAt"] = moment(&u.Started)
		}
		v["mission.completedAt"] = moment(u.Completed)
		v["status"] = t.LostText
		if u.Completed != nil {
			v["status"] = t.WonText
		}
	} else {
		v["status"] = t.LostText
	}
	completed, names := 0, []string{}
	for _, x := range missions {
		if p.completed(x.ID) {
			completed++
		}
		names = append(names, text(langs, lang, x.Name))
	}
	v["missions.total"], v["missions.completed"], v["missions.names"] = strconv.Itoa(len(missions)), strconv.Itoa(completed), strings.Join(names, "، ")
	played, won, total, right := 0, 0, 0, 0
	var scoreSum float64
	scored := 0
	pnames := []string{}
	for _, k := range packs {
		u := p.packs[k.ID]
		pnames = append(pnames, text(langs, lang, k.Name))
		if u.Plays > 0 {
			played++
		}
		if u.Wins > 0 {
			won++
		}
		total += u.LastTotal
		right += u.LastCorrect
		if u.LastTotal > 0 {
			scoreSum += u.Best
			scored++
		}
	}
	v["pack.total"], v["pack.played"], v["pack.won"], v["pack.names"] = strconv.Itoa(len(packs)), strconv.Itoa(played), strconv.Itoa(won), strings.Join(pnames, "، ")
	v["quiz.total"], v["quiz.correct"], v["quiz.wrong"] = strconv.Itoa(total), strconv.Itoa(right), strconv.Itoa(total-right)
	score := 0.0
	if scored > 0 {
		score = scoreSum / float64(scored)
	}
	v["score"] = num(score)
	return v, nil
}

func (s *service) rpcResultReport(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		TemplateKey string `json:"templateKey"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	t, err := scanTemplate(s.db.QueryRowContext(ctx, `SELECT `+templateCols+templateFrom+` WHERE t.key=$1 AND t.is_active`, strings.TrimSpace(in.TemplateKey)))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierr.NotFound("QUIZ_REPORT_TEMPLATE_NOT_FOUND").WithField("templateKey")
	}
	if err != nil {
		return nil, err
	}
	values, err := s.resultValues(ctx, t, userID)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	out := map[string]any{"templateKey": t.Key, "name": t.Name, "missionKey": t.MissionKey, "body": render(t.Body, values), "values": values, "createdAt": now.Format(time.RFC3339)}
	if t.WebhookURL != "" {
		body := map[string]any{"report": out, "user": map[string]string{"id": userID, "username": values["user.username"], "displayName": values["user.displayName"]}}
		if _, err := s.db.ExecContext(ctx, `INSERT INTO quiz_result_delivery (id, template_id, user_id, url, payload, next_attempt_at) VALUES ($1,$2,$3,$4,$5,$6)`,
			uuid.Must(uuid.NewV4()).String(), t.ID, userID, t.WebhookURL, jsonOf(body), now); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *service) deliverWebhooks(ctx context.Context, limit int) {
	cfg := s.settings(ctx)
	rows, err := s.db.QueryContext(ctx, `SELECT d.id::text, d.url, d.payload::text, d.attempts, t.webhook_secret FROM quiz_result_delivery d JOIN quiz_result_template t ON t.id=d.template_id
		WHERE d.status='pending' AND d.next_attempt_at <= now() ORDER BY d.next_attempt_at LIMIT $1`, limit)
	if err != nil {
		return
	}
	type job struct {
		id, url, payload, secret string
		attempts                 int
	}
	var jobs []job
	for rows.Next() {
		var j job
		if rows.Scan(&j.id, &j.url, &j.payload, &j.attempts, &j.secret) == nil {
			jobs = append(jobs, j)
		}
	}
	rows.Close()
	for _, j := range jobs {
		headers := map[string]string{"X-Quiz-Report-Delivery": j.id}
		if j.secret != "" && s.secretKey != nil {
			if plain, err := webhookBox.Open(s.secretKey, j.secret); err == nil {
				mac := hmac.New(sha256.New, plain)
				mac.Write([]byte(j.payload))
				headers["X-Quiz-Report-Signature"] = "sha256=" + hex.EncodeToString(mac.Sum(nil))
			}
		}
		status, err := s.send(ctx, j.url, []byte(j.payload), headers)
		attempts := j.attempts + 1
		switch {
		case err == nil && status >= 200 && status < 300:
			_, _ = s.db.ExecContext(ctx, `UPDATE quiz_result_delivery SET status='sent', attempts=$2, delivered_at=now(), last_error='' WHERE id=$1`, j.id, attempts)
		default:
			reason := fmt.Sprintf("remote returned %d", status)
			if err != nil {
				reason = err.Error()
			}
			terminal := errors.Is(err, errBlockedURL) || attempts >= cfg.Retention.WebhookAttempts
			next := "pending"
			if terminal {
				next = "failed"
			}
			delay := time.Duration(math.Pow(3, float64(attempts))*10) * time.Second
			_, _ = s.db.ExecContext(ctx, `UPDATE quiz_result_delivery SET status=$2, attempts=$3, last_error=$4, next_attempt_at=now()+$5::interval WHERE id=$1`,
				j.id, next, attempts, clip(reason, 300), fmt.Sprintf("%d seconds", int(delay.Seconds())))
		}
	}
}

func clip(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

var errBlockedURL = errors.New("webhook url is not allowed")

func (s *service) postWebhook(ctx context.Context, raw string, body []byte, headers map[string]string) (int, error) {
	if !webhookAllowed(raw) {
		return 0, errBlockedURL
	}
	u, _ := url.Parse(raw)
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	transport := &http.Transport{DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, ip := range ips {
			if privateIP(ip.IP) {
				return nil, errBlockedURL
			}
		}
		if len(ips) == 0 {
			return nil, errBlockedURL
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
	}}
	client := &http.Client{Timeout: 10 * time.Second, Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), errBlockedURL.Error()) {
			return 0, errBlockedURL
		}
		return 0, err
	}
	_ = res.Body.Close()
	return res.StatusCode, nil
}
