package upstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultAPI     = "https://api.github.com"
	maxBody        = 8 << 20
	userAgent      = "nakama-farsi-updates"
	requestTimeout = 20 * time.Second
)

var repoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,100}/[A-Za-z0-9_.-]{1,100}$`)

func NormalizeRepo(v string) (string, bool) {
	v = strings.TrimSpace(v)
	v = strings.TrimSuffix(v, "/")
	v = strings.TrimSuffix(v, ".git")
	for _, p := range []string{"https://github.com/", "http://github.com/", "https://www.github.com/", "git@github.com:", "github.com/"} {
		if strings.HasPrefix(strings.ToLower(v), p) {
			v = v[len(p):]
			break
		}
	}
	if !repoPattern.MatchString(v) || strings.Contains(v, "..") {
		return "", false
	}
	return v, true
}

var branchPattern = regexp.MustCompile(`^[A-Za-z0-9._/-]{1,200}$`)

func ValidBranch(v string) bool {
	return branchPattern.MatchString(v) && !strings.Contains(v, "..") && !strings.HasPrefix(v, "/") && !strings.HasSuffix(v, "/")
}

type Rate struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}

type RateLimitError struct {
	Status int
	Until  time.Time
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("github rate limit reached (HTTP %d), retry after %s", e.Status, e.Until.UTC().Format(time.RFC3339))
}

type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("github answered HTTP %d", e.Status)
	}
	return fmt.Sprintf("github answered HTTP %d: %s", e.Status, e.Message)
}

func IsNotFound(err error) bool {
	var h *HTTPError
	return errors.As(err, &h) && h.Status == http.StatusNotFound
}

type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client
	Now     func() time.Time
}

type response struct {
	notModified bool
	etag        string
	body        []byte
	rate        Rate
}

func (c *Client) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

func (c *Client) base() string {
	b := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	if b == "" {
		return DefaultAPI
	}
	return b
}

func (c *Client) get(ctx context.Context, path, etag string) (response, error) {
	var out response
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base()+path, nil)
	if err != nil {
		return out, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", userAgent)
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	hc := c.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: requestTimeout}
	}
	res, err := hc.Do(req)
	if err != nil {
		var ue *url.Error
		if errors.As(err, &ue) {
			return out, fmt.Errorf("github request failed: %v", ue.Err)
		}
		return out, fmt.Errorf("github request failed: %v", err)
	}
	defer res.Body.Close()
	out.rate = readRate(res.Header)
	out.etag = res.Header.Get("ETag")
	body, err := io.ReadAll(io.LimitReader(res.Body, maxBody+1))
	if err != nil {
		return out, fmt.Errorf("github read failed: %v", err)
	}
	if len(body) > maxBody {
		return out, errors.New("github answer is too large")
	}
	switch {
	case res.StatusCode == http.StatusNotModified:
		out.notModified = true
		return out, nil
	case res.StatusCode == http.StatusTooManyRequests || (res.StatusCode == http.StatusForbidden && (res.Header.Get("X-RateLimit-Remaining") == "0" || res.Header.Get("Retry-After") != "")):
		return out, &RateLimitError{Status: res.StatusCode, Until: c.retryAt(res.Header, out.rate)}
	case res.StatusCode >= 200 && res.StatusCode < 300:
		out.body = body
		return out, nil
	}
	var msg struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &msg)
	m := msg.Message
	if len(m) > 200 {
		m = m[:200]
	}
	return out, &HTTPError{Status: res.StatusCode, Message: m}
}

func (c *Client) retryAt(h http.Header, r Rate) time.Time {
	now := c.now()
	if v := strings.TrimSpace(h.Get("Retry-After")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return now.Add(time.Duration(n) * time.Second)
		}
	}
	if !r.Reset.IsZero() && r.Reset.After(now) {
		return r.Reset
	}
	return now.Add(time.Hour)
}

func readRate(h http.Header) Rate {
	var r Rate
	r.Limit, _ = strconv.Atoi(h.Get("X-RateLimit-Limit"))
	r.Remaining, _ = strconv.Atoi(h.Get("X-RateLimit-Remaining"))
	if n, err := strconv.ParseInt(h.Get("X-RateLimit-Reset"), 10, 64); err == nil && n > 0 {
		r.Reset = time.Unix(n, 0).UTC()
	}
	return r
}
