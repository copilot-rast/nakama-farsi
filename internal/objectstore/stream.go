package objectstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	MinMultipartBytes     = 5 << 20
	DefaultMultipartBytes = 8 << 20
	maxMultipartParts     = 10000
)

type StreamPutter interface {
	PutStream(ctx context.Context, key string, r io.Reader, contentType string) error
}

func PutStream(ctx context.Context, store Store, key string, r io.Reader, contentType string) error {
	if sp, ok := store.(StreamPutter); ok {
		return sp.PutStream(ctx, key, r, contentType)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return store.Put(ctx, key, data, contentType)
}

func (l *localStore) PutStream(_ context.Context, key string, r io.Reader, _ string) error {
	p, err := l.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(p), "."+filepath.Base(p)+".*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = io.Copy(f, r)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(tmp, 0o640)
	}
	if err == nil {
		err = os.Rename(tmp, p)
	}
	if err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (s *s3Store) multipartBytes() int {
	n := s.cfg.MultipartBytes
	if n <= 0 {
		n = DefaultMultipartBytes
	}
	if n < MinMultipartBytes {
		n = MinMultipartBytes
	}
	return int(n)
}

func (s *s3Store) PutStream(ctx context.Context, key string, r io.Reader, contentType string) error {
	if !validKey(key) {
		return fmt.Errorf("invalid object key %q", key)
	}
	buf := make([]byte, s.multipartBytes())
	n, err := io.ReadFull(r, buf)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return s.Put(ctx, key, buf[:n], contentType)
	}
	if err != nil {
		return err
	}
	id, err := s.createMultipart(ctx, key, contentType)
	if err != nil {
		return err
	}
	etags := []string{}
	fail := func(err error) error {
		actx, cancel := context.WithTimeout(context.WithoutCancel(ctx), defaultS3Timeout)
		defer cancel()
		if aerr := s.abortMultipart(actx, key, id); aerr != nil {
			return fmt.Errorf("%w (abort: %v)", err, aerr)
		}
		return err
	}
	for {
		if len(etags) >= maxMultipartParts {
			return fail(errors.New("s3 multipart upload has too many parts"))
		}
		etag, err := s.uploadPart(ctx, key, id, len(etags)+1, buf[:n])
		if err != nil {
			return fail(err)
		}
		etags = append(etags, etag)
		n, err = io.ReadFull(r, buf)
		if n == 0 && errors.Is(err, io.EOF) {
			break
		}
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			return fail(err)
		}
	}
	if err := s.completeMultipart(ctx, key, id, etags); err != nil {
		return fail(err)
	}
	return nil
}

func (s *s3Store) doQuery(ctx context.Context, method, key string, query url.Values, body []byte, header http.Header) (*http.Response, error) {
	u, err := s.objectURL(key)
	if err != nil {
		return nil, err
	}
	u.RawQuery = canonicalQuery(query)
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header[k] = v
	}
	hash := emptySHA256Hex
	if body != nil {
		sum := sha256.Sum256(body)
		hash = hex.EncodeToString(sum[:])
	}
	SignV4(req, hash, s.cfg.AccessKey, s.cfg.SecretKey, s.cfg.Region, "s3", s.now())
	return s.client.Do(req)
}

func s3BodyError(op string, raw []byte) error {
	var e struct {
		XMLName xml.Name `xml:"Error"`
		Code    string   `xml:"Code"`
	}
	if xml.Unmarshal(raw, &e) == nil && e.XMLName.Local == "Error" {
		return fmt.Errorf("s3 %s failed: 200 %s", op, e.Code)
	}
	return nil
}

func (s *s3Store) createMultipart(ctx context.Context, key, contentType string) (string, error) {
	h := http.Header{}
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	resp, err := s.doQuery(ctx, http.MethodPost, key, url.Values{"uploads": {""}}, nil, h)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return "", s3Error("create multipart", resp)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return "", err
	}
	var out struct {
		UploadID string `xml:"UploadId"`
	}
	if err := xml.Unmarshal(raw, &out); err != nil || out.UploadID == "" {
		return "", errors.New("s3 create multipart failed: no upload id")
	}
	return out.UploadID, nil
}

func (s *s3Store) uploadPart(ctx context.Context, key, id string, number int, data []byte) (string, error) {
	q := url.Values{"partNumber": {strconv.Itoa(number)}, "uploadId": {id}}
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}
		resp, err := s.doQuery(ctx, http.MethodPut, key, q, data, nil)
		if err != nil {
			last = err
			continue
		}
		if resp.StatusCode/100 != 2 {
			last = s3Error("upload part", resp)
			resp.Body.Close()
			if resp.StatusCode/100 == 4 {
				return "", last
			}
			continue
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		etag := resp.Header.Get("ETag")
		if etag == "" {
			return "", errors.New("s3 upload part failed: no ETag")
		}
		return etag, nil
	}
	return "", last
}

type completePart struct {
	PartNumber int    `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

type completeUpload struct {
	XMLName xml.Name       `xml:"CompleteMultipartUpload"`
	Parts   []completePart `xml:"Part"`
}

func (s *s3Store) completeMultipart(ctx context.Context, key, id string, etags []string) error {
	in := completeUpload{}
	for i, e := range etags {
		in.Parts = append(in.Parts, completePart{PartNumber: i + 1, ETag: e})
	}
	body, err := xml.Marshal(in)
	if err != nil {
		return err
	}
	h := http.Header{}
	h.Set("Content-Type", "application/xml")
	resp, err := s.doQuery(ctx, http.MethodPost, key, url.Values{"uploadId": {id}}, body, h)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return s3Error("complete multipart", resp)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return err
	}
	return s3BodyError("complete multipart", raw)
}

func (s *s3Store) abortMultipart(ctx context.Context, key, id string) error {
	resp, err := s.doQuery(ctx, http.MethodDelete, key, url.Values{"uploadId": {id}}, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 && resp.StatusCode != http.StatusNotFound {
		return s3Error("abort multipart", resp)
	}
	return nil
}
