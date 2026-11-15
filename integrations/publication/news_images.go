package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/objectstore"
)

const remoteImageMax = 2 << 20

var remoteImageTypes = map[string]string{"image/png": "png", "image/jpeg": "jpg", "image/webp": "webp"}

func dataDirFromArgs(args []string) string {
	for i, arg := range args {
		for _, flag := range []string{"--data_dir", "-data_dir"} {
			if arg == flag && i+1 < len(args) {
				return args[i+1]
			}
			if strings.HasPrefix(arg, flag+"=") {
				return strings.TrimPrefix(arg, flag+"=")
			}
		}
	}
	return "./data"
}

func (s *service) setupImageStore(env map[string]string) {
	base := strings.TrimRight(strings.TrimSpace(env["ECONOMY_PUBLIC_BASE_URL"]), "/")
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		return
	}
	dir := strings.TrimSpace(env["ECONOMY_UPLOAD_DIR"])
	if dir == "" {
		dir = filepath.Join(dataDirFromArgs(os.Args), "economy-uploads")
	}
	cfg, err := objectstore.ConfigFromEnv(env, objectstore.Namespace{LocalDir: dir, Prefix: "economy/"})
	if err != nil {
		s.logger.Warn("Publication news image copy is off: %v", err)
		return
	}
	store, err := objectstore.New(cfg, nil)
	if err != nil {
		s.logger.Warn("Publication news image copy is off: %v", err)
		return
	}
	s.imageStore = store
	s.imageBase = base
}

func publicAddr(ip net.IP) bool {
	return !(ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() || ip.IsInterfaceLocalMulticast())
}

func (s *service) fetchClient() *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second, Control: func(network, address string, _ syscall.RawConn) error {
		if s.fetchPrivate {
			return nil
		}
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}
		if ip := net.ParseIP(host); ip == nil || !publicAddr(ip) {
			return errors.New("private address refused")
		}
		return nil
	}}
	return &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{DialContext: dialer.DialContext, Proxy: nil, MaxIdleConns: 2, IdleConnTimeout: 30 * time.Second},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return errors.New("too many redirects")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return errors.New("redirect to a non-http address")
			}
			return nil
		}}
}

func (s *service) copyRemoteImage(ctx context.Context, url string) (string, error) {
	if s.imageStore == nil {
		return "", errors.New("image storage is not configured (ECONOMY_PUBLIC_BASE_URL)")
	}
	lower := strings.ToLower(url)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return "", errors.New("only http(s) images are copied")
	}
	if strings.HasPrefix(url, s.imageBase+"/v2/economy/files/") {
		return url, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "image/png,image/jpeg,image/webp")
	resp, err := s.fetchClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("image answered " + resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, remoteImageMax+1))
	if err != nil {
		return "", err
	}
	if len(data) == 0 || len(data) > remoteImageMax {
		return "", errors.New("image is empty or larger than 2 MB")
	}
	ext, ok := remoteImageTypes[http.DetectContentType(data)]
	if !ok {
		return "", errors.New("only png, jpg and webp images are copied")
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	name := hex.EncodeToString(raw) + "." + ext
	if err := s.imageStore.Put(ctx, name, data, "image/"+map[string]string{"png": "png", "jpg": "jpeg", "webp": "webp"}[ext]); err != nil {
		return "", err
	}
	return s.imageBase + "/v2/economy/files/" + name, nil
}
