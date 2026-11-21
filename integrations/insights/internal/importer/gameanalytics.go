package importer

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/events"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
)

type eventsRaw = events.Raw

const emptyHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

type gaConfig struct {
	Endpoint     string `json:"endpoint"`
	Bucket       string `json:"bucket"`
	Prefix       string `json:"prefix"`
	Region       string `json:"region"`
	AccessKey    string `json:"accessKey"`
	FilesPerRun  int    `json:"filesPerRun"`
	SkipSessions bool   `json:"skipSessions"`
}

type gaCursor struct {
	After string `json:"after"`
	Line  int    `json:"line,omitempty"`
}

type gameAnalytics struct{}

func (gameAnalytics) Kind() string { return "gameanalytics" }

func (gameAnalytics) parse(raw json.RawMessage) (gaConfig, error) {
	var c gaConfig
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, bad("config", "INVALID_JSON")
	}
	var err error
	if c.Endpoint, err = CheckEndpoint(c.Endpoint, ""); err != nil || c.Endpoint == "" {
		return c, bad("endpoint", "INVALID_URL")
	}
	if c.Bucket = strings.TrimSpace(c.Bucket); c.Bucket == "" || strings.ContainsAny(c.Bucket, "/?#") {
		return c, bad("bucket", "REQUIRED")
	}
	c.Prefix = strings.TrimLeft(strings.TrimSpace(c.Prefix), "/")
	if c.Region == "" {
		c.Region = "us-east-1"
	}
	if strings.TrimSpace(c.AccessKey) == "" {
		return c, bad("accessKey", "REQUIRED")
	}
	if c.FilesPerRun == 0 {
		c.FilesPerRun = 20
	}
	if c.FilesPerRun < 1 || c.FilesPerRun > 500 {
		return c, bad("filesPerRun", "VALUE_OUT_OF_RANGE")
	}
	return c, nil
}

func (g gameAnalytics) Check(raw json.RawMessage, _ string) (json.RawMessage, error) {
	c, err := g.parse(raw)
	if err != nil {
		return nil, err
	}
	return encode(c), nil
}

type s3List struct {
	Contents []struct {
		Key  string `xml:"Key"`
		Size int64  `xml:"Size"`
	} `xml:"Contents"`
	IsTruncated bool `xml:"IsTruncated"`
}

func (c gaConfig) request(ctx context.Context, hc *http.Client, secret, path string, q url.Values, now time.Time) (*http.Response, error) {
	u := c.Endpoint + "/" + url.PathEscape(c.Bucket)
	if path != "" {
		parts := strings.Split(path, "/")
		for i := range parts {
			parts[i] = url.PathEscape(parts[i])
		}
		u += "/" + strings.Join(parts, "/")
	}
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	objectstore.SignV4(req, emptyHash, c.AccessKey, secret, c.Region, "s3", now)
	if hc == nil {
		hc = &http.Client{Timeout: 120 * time.Second}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		resp.Body.Close()
		return nil, fmt.Errorf("S3 HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return resp, nil
}

func (g gameAnalytics) Pull(ctx context.Context, env Env) (Outcome, error) {
	c, err := g.parse(env.Config)
	if err != nil {
		return Outcome{Cursor: env.Cursor}, err
	}
	cur := decodeCursor[gaCursor](env.Cursor)
	out := Outcome{}
	q := url.Values{"list-type": {"2"}, "max-keys": {strconv.Itoa(c.FilesPerRun)}}
	if c.Prefix != "" {
		q.Set("prefix", c.Prefix)
	}
	if cur.After != "" {
		q.Set("start-after", cur.After)
	}
	resp, err := c.request(ctx, env.HTTP, env.Secret, "", q, env.Now)
	if err != nil {
		out.Cursor = encode(cur)
		return out, err
	}
	var list s3List
	err = xml.NewDecoder(io.LimitReader(resp.Body, 8<<20)).Decode(&list)
	resp.Body.Close()
	if err != nil {
		out.Cursor = encode(cur)
		return out, errors.New("S3 listing is not XML")
	}
	keys := []string{}
	if cur.Line > 0 && cur.After != "" {
		keys = append(keys, cur.After)
	}
	for _, o := range list.Contents {
		if o.Key != cur.After && !strings.HasSuffix(o.Key, "/") {
			keys = append(keys, o.Key)
		}
	}
	for _, key := range keys {
		skip := 0
		if key == cur.After {
			skip = cur.Line
		}
		read, skipped, lines, err := c.file(ctx, env, key, skip)
		out.Read += read
		out.Skipped += skipped
		if err != nil {
			if lines > 0 {
				cur = gaCursor{After: key, Line: lines}
			}
			out.Cursor = encode(cur)
			return out, err
		}
		cur = gaCursor{After: key}
		if env.Renew != nil {
			if err := env.Renew(ctx); err != nil {
				out.Cursor = encode(cur)
				return out, err
			}
		}
	}
	out.Cursor = encode(cur)
	out.More = list.IsTruncated
	return out, nil
}

func (c gaConfig) file(ctx context.Context, env Env, key string, skip int) (int64, int64, int, error) {
	resp, err := c.request(ctx, env.HTTP, env.Secret, key, nil, env.Now)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()
	var r io.Reader = io.LimitReader(resp.Body, 2<<30)
	if strings.HasSuffix(key, ".gz") || resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r)
		if err != nil {
			return 0, 0, 0, err
		}
		defer gz.Close()
		r = gz
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 4<<20)
	var read, skipped int64
	line := 0
	batch := []Item{}
	flushed := skip
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := env.Emit(ctx, batch); err != nil {
			return err
		}
		batch = batch[:0]
		flushed = line
		return nil
	}
	for sc.Scan() {
		line++
		if line <= skip {
			continue
		}
		text := strings.TrimSpace(sc.Text())
		if text == "" {
			continue
		}
		read++
		it, ok := c.line([]byte(text), key, line)
		if !ok {
			skipped++
			continue
		}
		batch = append(batch, it)
		if len(batch) >= 500 {
			if err := flush(); err != nil {
				return read, skipped, flushed, err
			}
		}
	}
	if err := sc.Err(); err != nil {
		return read, skipped, flushed, err
	}
	if err := flush(); err != nil {
		return read, skipped, flushed, err
	}
	return read, skipped, 0, nil
}

type gaLine struct {
	Data map[string]any `json:"data"`
}

func num(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	}
	return 0, false
}

func (c gaConfig) line(raw []byte, key string, n int) (Item, bool) {
	var l gaLine
	if err := json.Unmarshal(raw, &l); err != nil || l.Data == nil {
		var flat map[string]any
		if json.Unmarshal(raw, &flat) != nil {
			return Item{}, false
		}
		l.Data = flat
	}
	d := l.Data
	s := func(k string) string { v, _ := d[k].(string); return strings.TrimSpace(v) }
	cat := strings.ToLower(s("category"))
	eventID := s("event_id")
	it := Item{UserID: s("user_id"), SessionID: s("session_id"), Platform: strings.ToLower(s("platform")), AppVersion: s("build")}
	if it.UserID != "" {
		it.UserID = "ga:" + it.UserID
	}
	ts, ok := num(d["client_ts"])
	if !ok {
		ts, ok = num(d["arrival_ts"])
	}
	if !ok || ts <= 0 {
		return Item{}, false
	}
	it.At = time.Unix(int64(ts), 0).UTC()
	parts := strings.Split(eventID, ":")
	part := func(i int) string {
		if i < len(parts) {
			return parts[i]
		}
		return ""
	}
	switch cat {
	case "design":
		if eventID == "" {
			return Item{}, false
		}
		it.Raw = events.Raw{Type: "design", ID: eventID}
		if v, ok := num(d["value"]); ok {
			it.Raw.Value = v
		}
	case "progression":
		it.Raw = events.Raw{Type: "progression", Status: part(0), P1: part(1), P2: part(2), P3: part(3)}
		if v, ok := num(d["score"]); ok {
			it.Raw.Score = v
		}
	case "resource":
		amount, _ := num(d["amount"])
		it.Raw = events.Raw{Type: "resource", Flow: part(0), Currency: part(1), ItemType: part(2), Item: part(3), Amount: amount}
	case "business":
		amount, _ := num(d["amount"])
		it.Raw = events.Raw{Type: "business", Currency: s("currency"), Amount: amount, ItemType: part(0), Item: part(1), CartType: s("cart_type")}
	case "error":
		it.Raw = events.Raw{Type: "error", Severity: s("severity"), Message: s("message")}
	case "user":
		if c.SkipSessions {
			return Item{}, false
		}
		it.Raw = events.Raw{Name: "session_start"}
	case "session_end":
		if c.SkipSessions {
			return Item{}, false
		}
		it.Raw = events.Raw{Name: "session_end"}
		if v, ok := num(d["length"]); ok {
			it.Raw.Value = v
		}
	default:
		return Item{}, false
	}
	it.Key = "ga:" + key + ":" + strconv.Itoa(n)
	return it, true
}

func init() {
	register(gameAnalytics{})
}
