package importer

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	scopeBigQuery  = "https://www.googleapis.com/auth/bigquery.readonly"
	scopeAnalytics = "https://www.googleapis.com/auth/analytics.readonly"
	googleToken    = "https://oauth2.googleapis.com/token"
)

type serviceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
	ProjectID   string `json:"project_id"`
}

func parseServiceAccount(secret string) (*serviceAccount, *rsa.PrivateKey, error) {
	var sa serviceAccount
	if err := json.Unmarshal([]byte(secret), &sa); err != nil {
		return nil, nil, bad("secret", "INVALID_SERVICE_ACCOUNT")
	}
	if sa.ClientEmail == "" || sa.PrivateKey == "" {
		return nil, nil, bad("secret", "INVALID_SERVICE_ACCOUNT")
	}
	block, _ := pem.Decode([]byte(sa.PrivateKey))
	if block == nil {
		return nil, nil, bad("secret", "INVALID_SERVICE_ACCOUNT")
	}
	var key *rsa.PrivateKey
	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rk, ok := k.(*rsa.PrivateKey)
		if !ok {
			return nil, nil, bad("secret", "INVALID_SERVICE_ACCOUNT")
		}
		key = rk
	} else if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		key = k
	} else {
		return nil, nil, bad("secret", "INVALID_SERVICE_ACCOUNT")
	}
	return &sa, key, nil
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func signJWT(sa *serviceAccount, key *rsa.PrivateKey, scope, aud string, now time.Time) (string, error) {
	head := b64([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims, _ := json.Marshal(map[string]any{"iss": sa.ClientEmail, "scope": scope, "aud": aud, "iat": now.Unix(), "exp": now.Add(time.Hour).Unix()})
	unsigned := head + "." + b64(claims)
	sum := sha256.Sum256([]byte(unsigned))
	sig, err := rsa.SignPKCS1v15(nil, key, crypto.SHA256, sum[:])
	if err != nil {
		return "", err
	}
	return unsigned + "." + b64(sig), nil
}

func googleAccessToken(ctx context.Context, hc *http.Client, secret, scope, tokenOverride string, now time.Time) (string, error) {
	sa, key, err := parseServiceAccount(secret)
	if err != nil {
		return "", err
	}
	tokenURL := googleToken
	if tokenOverride != "" {
		tokenURL = tokenOverride
	} else if sa.TokenURI != "" && strings.HasPrefix(sa.TokenURI, "https://") {
		tokenURL = sa.TokenURI
	}
	jwt, err := signJWT(sa, key, scope, tokenURL, now)
	if err != nil {
		return "", err
	}
	form := url.Values{"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"}, "assertion": {jwt}}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var out struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if err := doJSON(hc, req, &out); err != nil {
		return "", fmt.Errorf("google token: %w", err)
	}
	if out.AccessToken == "" {
		return "", fmt.Errorf("google token: %s %s", out.Error, out.Description)
	}
	return out.AccessToken, nil
}

func doJSON(hc *http.Client, req *http.Request, out any) error {
	if hc == nil {
		hc = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		msg := string(raw)
		if len(msg) > 300 {
			msg = msg[:300]
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return errors.New("response is not JSON")
	}
	return nil
}

func postJSON(ctx context.Context, hc *http.Client, target, token string, body any, out any) error {
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	return doJSON(hc, req, out)
}

func getJSON(ctx context.Context, hc *http.Client, target, token string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return doJSON(hc, req, out)
}

type gaCommon struct {
	StartDate  string            `json:"startDate"`
	DaysPerRun int               `json:"daysPerRun"`
	LagDays    int               `json:"lagDays"`
	NameMap    map[string]string `json:"nameMap"`
	Include    []string          `json:"includeEvents"`
	APIBase    string            `json:"apiBase,omitempty"`
	TokenURL   string            `json:"tokenUrl,omitempty"`
}

func (c *gaCommon) check(now time.Time) error {
	if c.StartDate == "" {
		c.StartDate = now.AddDate(0, 0, -7).Format(time.DateOnly)
	}
	if _, err := time.Parse(time.DateOnly, c.StartDate); err != nil {
		return bad("startDate", "INVALID_DATE")
	}
	if c.DaysPerRun == 0 {
		c.DaysPerRun = 3
	}
	if c.DaysPerRun < 1 || c.DaysPerRun > 31 {
		return bad("daysPerRun", "VALUE_OUT_OF_RANGE")
	}
	if c.LagDays == 0 {
		c.LagDays = 1
	}
	if c.LagDays < 0 || c.LagDays > 7 {
		return bad("lagDays", "VALUE_OUT_OF_RANGE")
	}
	for k, v := range c.NameMap {
		if EventName(v) != v || v == "" || k == "" {
			return bad("nameMap", "INVALID_EVENT_NAME")
		}
	}
	if len(c.Include) > 200 {
		return bad("includeEvents", "VALUE_OUT_OF_RANGE")
	}
	var err error
	if c.TokenURL, err = CheckEndpoint(c.TokenURL, ""); err != nil {
		return bad("tokenUrl", "INVALID_URL")
	}
	return nil
}

func (c gaCommon) name(ga string) string {
	if v, ok := c.NameMap[ga]; ok {
		return v
	}
	return EventName(ga)
}

func (c gaCommon) wanted(ga string) bool {
	if len(c.Include) == 0 {
		return true
	}
	for _, n := range c.Include {
		if n == ga {
			return true
		}
	}
	return false
}

func (c gaCommon) lastDay(now time.Time) string {
	return now.AddDate(0, 0, -c.LagDays).Format(time.DateOnly)
}

type bqConfig struct {
	gaCommon
	ProjectID  string   `json:"projectId"`
	Dataset    string   `json:"dataset"`
	ParamTags  []string `json:"paramTags"`
	ValueParam string   `json:"valueParam"`
	Location   string   `json:"location"`
}

type bqCursor struct {
	Day       string `json:"day"`
	JobID     string `json:"jobId,omitempty"`
	Location  string `json:"location,omitempty"`
	PageToken string `json:"pageToken,omitempty"`
}

type ga4BigQuery struct{}

func (ga4BigQuery) Kind() string { return "ga4_bigquery" }

var bqIdent = func(s string) bool {
	if s == "" || len(s) > 1024 {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func (ga4BigQuery) parse(raw json.RawMessage, now time.Time) (bqConfig, error) {
	var c bqConfig
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, bad("config", "INVALID_JSON")
	}
	if !bqIdent(c.ProjectID) {
		return c, bad("projectId", "REQUIRED")
	}
	if !bqIdent(c.Dataset) {
		return c, bad("dataset", "REQUIRED")
	}
	if err := c.check(now); err != nil {
		return c, err
	}
	for _, p := range c.ParamTags {
		if GroupKey(p) == "" {
			return c, bad("paramTags", "INVALID_KEY")
		}
	}
	if len(c.ParamTags) > 8 {
		return c, bad("paramTags", "VALUE_OUT_OF_RANGE")
	}
	var err error
	if c.APIBase, err = CheckEndpoint(c.APIBase, "https://bigquery.googleapis.com"); err != nil {
		return c, bad("apiBase", "INVALID_URL")
	}
	return c, nil
}

func (g ga4BigQuery) Check(raw json.RawMessage, secret string) (json.RawMessage, error) {
	c, err := g.parse(raw, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if secret != "" {
		if _, _, err := parseServiceAccount(secret); err != nil {
			return nil, err
		}
	}
	return encode(c), nil
}

type bqResponse struct {
	JobComplete  bool `json:"jobComplete"`
	JobReference struct {
		JobID    string `json:"jobId"`
		Location string `json:"location"`
	} `json:"jobReference"`
	PageToken string `json:"pageToken"`
	Rows      []struct {
		F []struct {
			V any `json:"v"`
		} `json:"f"`
	} `json:"rows"`
}

func (g ga4BigQuery) Pull(ctx context.Context, env Env) (Outcome, error) {
	c, err := g.parse(env.Config, env.Now)
	if err != nil {
		return Outcome{Cursor: env.Cursor}, err
	}
	cur := decodeCursor[bqCursor](env.Cursor)
	if cur.Day == "" {
		cur.Day = c.StartDate
	}
	last := c.lastDay(env.Now)
	out := Outcome{}
	if cur.Day > last {
		out.Cursor = encode(cur)
		return out, nil
	}
	token, err := googleAccessToken(ctx, env.HTTP, env.Secret, scopeBigQuery, c.TokenURL, env.Now)
	if err != nil {
		out.Cursor = encode(cur)
		return out, err
	}
	days := 0
	for cur.Day <= last && days < c.DaysPerRun {
		var resp bqResponse
		base := c.APIBase + "/bigquery/v2/projects/" + url.PathEscape(c.ProjectID) + "/queries"
		if cur.JobID == "" {
			suffix := strings.ReplaceAll(cur.Day, "-", "")
			q := "SELECT CAST(event_timestamp AS STRING), event_name, IFNULL(user_id, ''), IFNULL(user_pseudo_id, ''), IFNULL(platform, ''), IFNULL(app_info.version, ''), CAST(IFNULL(event_bundle_sequence_id, 0) AS STRING), " +
				"CAST((SELECT value.int_value FROM UNNEST(event_params) WHERE key = 'ga_session_id') AS STRING), TO_JSON_STRING(event_params) " +
				"FROM `" + c.ProjectID + "." + c.Dataset + ".events_" + suffix + "` ORDER BY event_timestamp"
			body := map[string]any{"query": q, "useLegacySql": false, "maxResults": 1000, "timeoutMs": 30000}
			if c.Location != "" {
				body["location"] = c.Location
			}
			if err := postJSON(ctx, env.HTTP, base, token, body, &resp); err != nil {
				if strings.Contains(err.Error(), "HTTP 404") {
					cur = bqCursor{Day: nextDay(cur.Day)}
					days++
					continue
				}
				out.Cursor = encode(cur)
				return out, err
			}
			cur.JobID, cur.Location = resp.JobReference.JobID, resp.JobReference.Location
		} else {
			v := url.Values{"maxResults": {"1000"}, "timeoutMs": {"30000"}}
			if cur.PageToken != "" {
				v.Set("pageToken", cur.PageToken)
			}
			if cur.Location != "" {
				v.Set("location", cur.Location)
			}
			if err := getJSON(ctx, env.HTTP, base+"/"+url.PathEscape(cur.JobID)+"?"+v.Encode(), token, &resp); err != nil {
				out.Cursor = encode(cur)
				return out, err
			}
		}
		if !resp.JobComplete {
			out.Cursor, out.More = encode(cur), true
			return out, nil
		}
		items := make([]Item, 0, len(resp.Rows))
		for _, row := range resp.Rows {
			out.Read++
			it, ok := c.row(row.F)
			if !ok {
				out.Skipped++
				continue
			}
			items = append(items, it)
		}
		if len(items) > 0 {
			if err := env.Emit(ctx, items); err != nil {
				out.Cursor = encode(cur)
				return out, err
			}
		}
		if resp.PageToken != "" {
			cur.PageToken = resp.PageToken
		} else {
			cur = bqCursor{Day: nextDay(cur.Day)}
			days++
		}
		if env.Renew != nil {
			if err := env.Renew(ctx); err != nil {
				out.Cursor = encode(cur)
				return out, err
			}
		}
		if env.Max > 0 && out.Read >= int64(env.Max) {
			break
		}
	}
	out.Cursor = encode(cur)
	out.More = cur.Day <= last
	return out, nil
}

func nextDay(d string) string {
	t, err := time.Parse(time.DateOnly, d)
	if err != nil {
		return d
	}
	return t.AddDate(0, 0, 1).Format(time.DateOnly)
}

func str(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case nil:
		return ""
	}
	return fmt.Sprint(v)
}

func (c bqConfig) row(f []struct {
	V any `json:"v"`
}) (Item, bool) {
	if len(f) < 9 {
		return Item{}, false
	}
	tsRaw, gaName, uid, pseudo, platform, version, seq, sess, params := str(f[0].V), str(f[1].V), str(f[2].V), str(f[3].V), str(f[4].V), str(f[5].V), str(f[6].V), str(f[7].V), str(f[8].V)
	if !c.wanted(gaName) {
		return Item{}, false
	}
	us, err := strconv.ParseInt(tsRaw, 10, 64)
	if err != nil || us <= 0 {
		return Item{}, false
	}
	name := c.name(gaName)
	if name == "" {
		return Item{}, false
	}
	user := uid
	if user == "" && pseudo != "" {
		user = "ga4:" + pseudo
	}
	var list []struct {
		Key   string         `json:"key"`
		Value map[string]any `json:"value"`
	}
	_ = json.Unmarshal([]byte(params), &list)
	values := map[string]string{}
	nums := map[string]float64{}
	for _, p := range list {
		for _, k := range []string{"string_value", "int_value", "double_value", "float_value"} {
			v, ok := p.Value[k]
			if !ok || v == nil {
				continue
			}
			s := str(v)
			values[p.Key] = s
			if k != "string_value" {
				if n, err := strconv.ParseFloat(s, 64); err == nil {
					nums[p.Key] = n
				}
			}
			break
		}
	}
	tags := map[string]string{}
	for _, p := range c.ParamTags {
		if v := TagID(values[p]); v != "" {
			tags[GroupKey(p)] = v
		}
	}
	raw := rawEvent(name, tags)
	if c.ValueParam != "" {
		if n, ok := nums[c.ValueParam]; ok {
			raw.Value = n
		}
	}
	return Item{Key: "ga4:" + tsRaw + ":" + pseudo + ":" + gaName + ":" + seq, Raw: raw, UserID: user, SessionID: sess, At: time.UnixMicro(us).UTC(),
		Platform: strings.ToLower(platform), AppVersion: version}, true
}

func rawEvent(name string, tags map[string]string) eventsRaw {
	r := eventsRaw{Name: name}
	if len(tags) > 0 {
		r.Tags, _ = json.Marshal(tags)
	}
	return r
}

type dataConfig struct {
	gaCommon
	PropertyID string   `json:"propertyId"`
	Dimensions []string `json:"dimensions"`
}

type dataCursor struct {
	Day    string `json:"day"`
	Offset int    `json:"offset,omitempty"`
}

type ga4Data struct{}

func (ga4Data) Kind() string { return "ga4_data" }

func (ga4Data) parse(raw json.RawMessage, now time.Time) (dataConfig, error) {
	var c dataConfig
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, bad("config", "INVALID_JSON")
	}
	if _, err := strconv.ParseInt(c.PropertyID, 10, 64); err != nil {
		return c, bad("propertyId", "REQUIRED")
	}
	if err := c.check(now); err != nil {
		return c, err
	}
	if len(c.Dimensions) > 6 {
		return c, bad("dimensions", "VALUE_OUT_OF_RANGE")
	}
	for _, d := range c.Dimensions {
		if GroupKey(d) == "" || d == "date" || d == "eventName" {
			return c, bad("dimensions", "INVALID_KEY")
		}
	}
	var err error
	if c.APIBase, err = CheckEndpoint(c.APIBase, "https://analyticsdata.googleapis.com"); err != nil {
		return c, bad("apiBase", "INVALID_URL")
	}
	return c, nil
}

func (g ga4Data) Check(raw json.RawMessage, secret string) (json.RawMessage, error) {
	c, err := g.parse(raw, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if secret != "" {
		if _, _, err := parseServiceAccount(secret); err != nil {
			return nil, err
		}
	}
	return encode(c), nil
}

func (g ga4Data) Pull(ctx context.Context, env Env) (Outcome, error) {
	c, err := g.parse(env.Config, env.Now)
	if err != nil {
		return Outcome{Cursor: env.Cursor}, err
	}
	cur := decodeCursor[dataCursor](env.Cursor)
	if cur.Day == "" {
		cur.Day = c.StartDate
	}
	last := c.lastDay(env.Now)
	out := Outcome{}
	if cur.Day > last {
		out.Cursor = encode(cur)
		return out, nil
	}
	token, err := googleAccessToken(ctx, env.HTTP, env.Secret, scopeAnalytics, c.TokenURL, env.Now)
	if err != nil {
		out.Cursor = encode(cur)
		return out, err
	}
	dims := []map[string]string{{"name": "date"}, {"name": "eventName"}}
	for _, d := range c.Dimensions {
		dims = append(dims, map[string]string{"name": d})
	}
	days := 0
	for cur.Day <= last && days < c.DaysPerRun {
		body := map[string]any{"dateRanges": []map[string]string{{"startDate": cur.Day, "endDate": cur.Day}}, "dimensions": dims,
			"metrics": []map[string]string{{"name": "eventCount"}}, "limit": 10000, "offset": cur.Offset}
		var resp struct {
			RowCount int `json:"rowCount"`
			Rows     []struct {
				DimensionValues []struct {
					Value string `json:"value"`
				} `json:"dimensionValues"`
				MetricValues []struct {
					Value string `json:"value"`
				} `json:"metricValues"`
			} `json:"rows"`
		}
		if err := postJSON(ctx, env.HTTP, c.APIBase+"/v1beta/properties/"+c.PropertyID+":runReport", token, body, &resp); err != nil {
			out.Cursor = encode(cur)
			return out, err
		}
		at, _ := time.Parse(time.DateOnly, cur.Day)
		items := []Item{}
		for _, row := range resp.Rows {
			out.Read++
			if len(row.DimensionValues) < 2 || len(row.MetricValues) < 1 || !c.wanted(row.DimensionValues[1].Value) {
				out.Skipped++
				continue
			}
			n, err := strconv.ParseFloat(row.MetricValues[0].Value, 64)
			name := c.name(row.DimensionValues[1].Value)
			if err != nil || name == "" {
				out.Skipped++
				continue
			}
			tags := map[string]string{}
			keyParts := []string{"ga4d", cur.Day, row.DimensionValues[1].Value}
			for i, d := range c.Dimensions {
				if i+2 < len(row.DimensionValues) {
					v := TagID(row.DimensionValues[i+2].Value)
					keyParts = append(keyParts, v)
					if v != "" && v != "(not set)" {
						tags[GroupKey(d)] = v
					}
				}
			}
			raw := rawEvent(name, tags)
			raw.Value = n
			items = append(items, Item{Key: strings.Join(keyParts, ":"), Raw: raw, UserID: "ga4:aggregate", At: at.Add(12 * time.Hour)})
		}
		if len(items) > 0 {
			if err := env.Emit(ctx, items); err != nil {
				out.Cursor = encode(cur)
				return out, err
			}
		}
		if cur.Offset+len(resp.Rows) < resp.RowCount && len(resp.Rows) > 0 {
			cur.Offset += len(resp.Rows)
		} else {
			cur = dataCursor{Day: nextDay(cur.Day)}
			days++
		}
		if env.Renew != nil {
			if err := env.Renew(ctx); err != nil {
				out.Cursor = encode(cur)
				return out, err
			}
		}
	}
	out.Cursor = encode(cur)
	out.More = cur.Day <= last
	return out, nil
}

func init() {
	register(ga4BigQuery{})
	register(ga4Data{})
}
