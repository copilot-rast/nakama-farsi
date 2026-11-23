package objectstore

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	amzDateFormat  = "20060102T150405Z"
	amzDayFormat   = "20060102"
	sigAlgorithm   = "AWS4-HMAC-SHA256"
	emptySHA256Hex = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

type s3Store struct {
	cfg    Config
	client *http.Client
	now    func() time.Time
}

func (s *s3Store) Driver() string { return DriverS3 }

func (s *s3Store) fullKey(key string) string { return s.cfg.Prefix + key }

func (s *s3Store) PublicURL(key string) string {
	if s.cfg.PublicBaseURL == "" || !validKey(key) {
		return ""
	}
	return s.cfg.PublicBaseURL + "/" + encodePath(s.fullKey(key))
}

func (s *s3Store) objectURL(key string) (*url.URL, error) {
	if !validKey(key) {
		return nil, fmt.Errorf("invalid object key %q", key)
	}
	u, err := url.Parse(s.cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	full := s.fullKey(key)
	if s.cfg.PathStyle {
		u.Path = "/" + s.cfg.Bucket + "/" + full
		u.RawPath = "/" + s.cfg.Bucket + "/" + encodePath(full)
	} else {
		u.Host = s.cfg.Bucket + "." + u.Host
		u.Path = "/" + full
		u.RawPath = "/" + encodePath(full)
	}
	return u, nil
}

func (s *s3Store) do(ctx context.Context, method, key string, body []byte, header http.Header) (*http.Response, error) {
	return s.doQuery(ctx, method, key, nil, body, header)
}

func s3Error(op string, resp *http.Response) error {
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	code := ""
	if i := bytes.Index(raw, []byte("<Code>")); i >= 0 {
		rest := raw[i+6:]
		if j := bytes.Index(rest, []byte("</Code>")); j >= 0 {
			code = string(rest[:j])
		}
	}
	if code == "" {
		code = http.StatusText(resp.StatusCode)
	}
	return fmt.Errorf("s3 %s failed: %d %s", op, resp.StatusCode, code)
}

func (s *s3Store) Put(ctx context.Context, key string, data []byte, contentType string) error {
	if data == nil {
		data = []byte{}
	}
	h := http.Header{}
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	resp, err := s.do(ctx, http.MethodPut, key, data, h)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return s3Error("put", resp)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (s *s3Store) Get(ctx context.Context, key string) (*Object, error) {
	if !validKey(key) {
		return nil, ErrNotFound
	}
	resp, err := s.do(ctx, http.MethodGet, key, nil, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, s3Error("get", resp)
	}
	obj := &Object{Body: resp.Body, Size: resp.ContentLength, ContentType: resp.Header.Get("Content-Type")}
	if t, err := http.ParseTime(resp.Header.Get("Last-Modified")); err == nil {
		obj.ModTime = t
	}
	return obj, nil
}

func (s *s3Store) Delete(ctx context.Context, key string) error {
	resp, err := s.do(ctx, http.MethodDelete, key, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 && resp.StatusCode != http.StatusNotFound {
		return s3Error("delete", resp)
	}
	return nil
}

func SignV4(req *http.Request, payloadHash, accessKey, secretKey, region, service string, at time.Time) {
	at = at.UTC()
	amzDate := at.Format(amzDateFormat)
	day := at.Format(amzDayFormat)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	values := map[string]string{"host": host}
	for k, v := range req.Header {
		name := strings.ToLower(k)
		if name == "authorization" {
			continue
		}
		parts := make([]string, len(v))
		for i, item := range v {
			parts[i] = strings.Join(strings.Fields(item), " ")
		}
		values[name] = strings.Join(parts, ",")
	}
	names := make([]string, 0, len(values))
	for k := range values {
		names = append(names, k)
	}
	sort.Strings(names)
	var canonHeaders strings.Builder
	for _, n := range names {
		canonHeaders.WriteString(n + ":" + values[n] + "\n")
	}
	signed := strings.Join(names, ";")
	path := req.URL.Path
	if path == "" {
		path = "/"
	}
	canonical := strings.Join([]string{req.Method, encodePath(path), canonicalQuery(req.URL.Query()), canonHeaders.String(), signed, payloadHash}, "\n")
	scope := day + "/" + region + "/" + service + "/aws4_request"
	sum := sha256.Sum256([]byte(canonical))
	toSign := sigAlgorithm + "\n" + amzDate + "\n" + scope + "\n" + hex.EncodeToString(sum[:])
	key := hmacSHA256([]byte("AWS4"+secretKey), day)
	key = hmacSHA256(key, region)
	key = hmacSHA256(key, service)
	key = hmacSHA256(key, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(key, toSign))
	req.Header.Set("Authorization", sigAlgorithm+" Credential="+accessKey+"/"+scope+", SignedHeaders="+signed+", Signature="+signature)
}

func hmacSHA256(key []byte, data string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(data))
	return m.Sum(nil)
}

func canonicalQuery(q url.Values) string {
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := []string{}
	for _, k := range keys {
		vals := append([]string(nil), q[k]...)
		sort.Strings(vals)
		for _, v := range vals {
			parts = append(parts, uriEncode(k, true)+"="+uriEncode(v, true))
		}
	}
	return strings.Join(parts, "&")
}

func encodePath(p string) string { return uriEncode(p, false) }

func uriEncode(s string, encodeSlash bool) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-', c == '_', c == '.', c == '~':
			b.WriteByte(c)
		case c == '/' && !encodeSlash:
			b.WriteByte(c)
		default:
			b.WriteString("%" + strings.ToUpper(strconv.FormatInt(int64(c)|0x100, 16)[1:]))
		}
	}
	return b.String()
}
