package objectstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DriverEnv        = "STORAGE_DRIVER"
	S3EndpointEnv    = "STORAGE_S3_ENDPOINT"
	S3BucketEnv      = "STORAGE_S3_BUCKET"
	S3RegionEnv      = "STORAGE_S3_REGION"
	S3AccessKeyEnv   = "STORAGE_S3_ACCESS_KEY"
	S3SecretKeyEnv   = "STORAGE_S3_SECRET_KEY"
	S3PathStyleEnv   = "STORAGE_S3_PATH_STYLE"
	S3PublicBaseEnv  = "STORAGE_S3_PUBLIC_BASE_URL"
	S3PrefixEnv      = "STORAGE_S3_PREFIX"
	DriverLocal      = "local"
	DriverS3         = "s3"
	defaultS3Region  = "us-east-1"
	defaultS3Timeout = 30 * time.Second
)

var ErrNotFound = errors.New("object not found")

type Object struct {
	Body        io.ReadCloser
	Size        int64
	ContentType string
	ModTime     time.Time
}

type Store interface {
	Driver() string
	Put(ctx context.Context, key string, data []byte, contentType string) error
	Get(ctx context.Context, key string) (*Object, error)
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}

type Config struct {
	Driver         string
	LocalDir       string
	Endpoint       string
	Bucket         string
	Region         string
	AccessKey      string
	SecretKey      string
	PathStyle      bool
	PublicBaseURL  string
	Prefix         string
	MultipartBytes int64
}

type Namespace struct {
	LocalDir string
	Prefix   string
}

func ConfigFromEnv(env map[string]string, ns Namespace) (Config, error) {
	get := func(k string) string { return strings.TrimSpace(env[k]) }
	cfg := Config{Driver: strings.ToLower(get(DriverEnv)), LocalDir: ns.LocalDir}
	if cfg.Driver == "" {
		cfg.Driver = DriverLocal
	}
	switch cfg.Driver {
	case DriverLocal:
		if cfg.LocalDir == "" {
			return cfg, errors.New("local storage needs a directory")
		}
		return cfg, nil
	case DriverS3:
	default:
		return cfg, fmt.Errorf("%s must be %q or %q", DriverEnv, DriverLocal, DriverS3)
	}
	cfg.Endpoint = strings.TrimRight(get(S3EndpointEnv), "/")
	cfg.Bucket = get(S3BucketEnv)
	cfg.Region = get(S3RegionEnv)
	cfg.AccessKey = get(S3AccessKeyEnv)
	cfg.SecretKey = get(S3SecretKeyEnv)
	cfg.PublicBaseURL = strings.TrimRight(get(S3PublicBaseEnv), "/")
	cfg.PathStyle = true
	switch strings.ToLower(get(S3PathStyleEnv)) {
	case "", "true", "1", "yes":
	case "false", "0", "no":
		cfg.PathStyle = false
	default:
		return cfg, fmt.Errorf("%s must be true or false", S3PathStyleEnv)
	}
	if cfg.Region == "" {
		cfg.Region = defaultS3Region
	}
	u, err := url.Parse(cfg.Endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || (u.Path != "" && u.Path != "/") {
		return cfg, fmt.Errorf("%s must be an http(s) URL without a path", S3EndpointEnv)
	}
	if cfg.Bucket == "" || strings.ContainsAny(cfg.Bucket, "/ ") {
		return cfg, fmt.Errorf("%s is required and must not contain / or spaces", S3BucketEnv)
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return cfg, fmt.Errorf("%s and %s are required", S3AccessKeyEnv, S3SecretKeyEnv)
	}
	if cfg.PublicBaseURL != "" {
		if p, err := url.Parse(cfg.PublicBaseURL); err != nil || (p.Scheme != "http" && p.Scheme != "https") || p.Host == "" {
			return cfg, fmt.Errorf("%s must be an http(s) URL", S3PublicBaseEnv)
		}
	}
	cfg.Prefix = strings.Trim(get(S3PrefixEnv), "/")
	if cfg.Prefix != "" {
		cfg.Prefix += "/"
	}
	cfg.Prefix += strings.TrimLeft(ns.Prefix, "/")
	return cfg, nil
}

func New(cfg Config, client *http.Client) (Store, error) {
	switch cfg.Driver {
	case DriverLocal, "":
		if cfg.LocalDir == "" {
			return nil, errors.New("local storage needs a directory")
		}
		return &localStore{dir: cfg.LocalDir}, nil
	case DriverS3:
		if client == nil {
			client = &http.Client{Timeout: defaultS3Timeout}
		}
		return &s3Store{cfg: cfg, client: client, now: time.Now}, nil
	}
	return nil, fmt.Errorf("unknown storage driver %q", cfg.Driver)
}

func FromEnv(env map[string]string, ns Namespace) (Store, error) {
	cfg, err := ConfigFromEnv(env, ns)
	if err != nil {
		return nil, err
	}
	return New(cfg, nil)
}

func validKey(key string) bool {
	if key == "" || strings.HasPrefix(key, "/") || strings.Contains(key, "\\") || strings.ContainsRune(key, 0) {
		return false
	}
	for _, part := range strings.Split(key, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

type localStore struct {
	dir string
}

func (l *localStore) Driver() string { return DriverLocal }

func (l *localStore) PublicURL(string) string { return "" }

func (l *localStore) path(key string) (string, error) {
	if !validKey(key) {
		return "", fmt.Errorf("invalid object key %q", key)
	}
	return filepath.Join(l.dir, filepath.FromSlash(key)), nil
}

func (l *localStore) Put(_ context.Context, key string, data []byte, _ string) error {
	p, err := l.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return err
	}
	tmp := filepath.Join(filepath.Dir(p), "."+filepath.Base(p)+".tmp")
	if err := os.WriteFile(tmp, data, 0o640); err != nil {
		return err
	}
	if err := os.Rename(tmp, p); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (l *localStore) Get(_ context.Context, key string) (*Object, error) {
	p, err := l.path(key)
	if err != nil {
		return nil, ErrNotFound
	}
	f, err := os.Open(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		f.Close()
		return nil, ErrNotFound
	}
	return &Object{Body: f, Size: info.Size(), ModTime: info.ModTime()}, nil
}

func (l *localStore) Delete(_ context.Context, key string) error {
	p, err := l.path(key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
