package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha1" //nolint:gosec // npm clients check the tarball with a SHA-1 shasum; SHA-512 integrity is sent as well
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apikey"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
)

// Unity package registry. Unity's Package Manager speaks the npm registry protocol: a scoped
// registry in Packages/manifest.json points at <server>/upm, ~/.upmconfig.toml holds an API key
// (sent as a Bearer token), and the registry answers the package document ("packument"), the
// tarballs and a search. Packages are published by CI with an API key (scope upm:publish).

const (
	upmPrefix       = "/upm"
	ScopeUpmRead    = "upm:read"
	ScopeUpmPublish = "upm:publish"
	// partBytes stays below Nakama's socket.max_request_size_bytes (256 KB by default), which
	// caps every request including plugin routes; larger archives are sent in parts.
	upmPartBytes = 192 * 1024
	upmMaxParts  = 2600
)

var (
	upmNamePattern    = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,213}$`)
	upmVersionPattern = regexp.MustCompile(`^(0|[1-9]\d{0,5})\.(0|[1-9]\d{0,5})\.(0|[1-9]\d{0,5})(?:-([0-9A-Za-z.-]{1,32}))?$`)
	upmUploadPattern  = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)
)

// semver is a parsed x.y.z[-pre] version (the legacy SdkSemVer rules).
type semver struct {
	major, minor, patch int
	pre                 string
}

func parseSemver(v string) (semver, bool) {
	m := upmVersionPattern.FindStringSubmatch(v)
	if m == nil {
		return semver{}, false
	}
	a, _ := strconv.Atoi(m[1])
	b, _ := strconv.Atoi(m[2])
	c, _ := strconv.Atoi(m[3])
	return semver{a, b, c, m[4]}, true
}

func (a semver) less(b semver) bool {
	if a.major != b.major {
		return a.major < b.major
	}
	if a.minor != b.minor {
		return a.minor < b.minor
	}
	if a.patch != b.patch {
		return a.patch < b.patch
	}
	if a.pre == b.pre {
		return false
	}
	if a.pre == "" {
		return false
	}
	if b.pre == "" {
		return true
	}
	return a.pre < b.pre
}

type packageVersion struct {
	Name        string          `json:"name"`
	Version     string          `json:"version"`
	Manifest    json.RawMessage `json:"-"`
	DisplayName string          `json:"displayName,omitempty"`
	Description string          `json:"description,omitempty"`
	Shasum      string          `json:"shasum"`
	Integrity   string          `json:"-"`
	Size        int64           `json:"size"`
	StorageKey  string          `json:"-"`
	PublishedBy string          `json:"publishedBy"`
	CreatedAt   time.Time       `json:"createdAt"`
}

func (p *packageVersion) fillFromManifest() {
	var m struct {
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
	}
	_ = json.Unmarshal(p.Manifest, &m)
	p.DisplayName, p.Description = m.DisplayName, m.Description
}

func upmFileName(name, version string) string { return name + "-" + version + ".tgz" }

// readManifest returns package/package.json of a gzipped npm tarball.
func readManifest(tarball []byte) (map[string]any, error) {
	gz, err := gzip.NewReader(bytes.NewReader(tarball))
	if err != nil {
		return nil, apierr.BadRequest("UPM_PACKAGE_INVALID").WithDetail("the file is not a gzipped npm tarball")
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil, apierr.BadRequest("UPM_PACKAGE_INVALID").WithDetail("the tarball has no package/package.json")
		}
		if err != nil {
			return nil, apierr.BadRequest("UPM_PACKAGE_INVALID").WithDetail("the file is not a gzipped npm tarball")
		}
		if strings.TrimLeft(h.Name, "./") != "package/package.json" {
			continue
		}
		if h.Size > 256*1024 {
			return nil, apierr.BadRequest("UPM_PACKAGE_INVALID").WithDetail("package.json is too large")
		}
		raw, err := io.ReadAll(io.LimitReader(tr, 256*1024+1))
		if err != nil {
			return nil, apierr.BadRequest("UPM_PACKAGE_INVALID").WithDetail("package.json could not be read")
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil || m == nil {
			return nil, apierr.BadRequest("UPM_PACKAGE_INVALID").WithDetail("package.json is not a JSON object")
		}
		return m, nil
	}
}

func (s *service) packagesConfig(ctx context.Context) (settings.Packages, error) {
	cfg, _ := settings.Load(ctx, s.db)
	if !cfg.Packages.Enabled {
		return cfg.Packages, apierr.NotFound("UPM_REGISTRY_DISABLED")
	}
	if s.store == nil {
		return cfg.Packages, apierr.Unavailable("UPM_STORAGE_UNAVAILABLE")
	}
	return cfg.Packages, nil
}

// publishPackage validates an archive, stores it and records the version.
func (s *service) publishPackage(ctx context.Context, tarball []byte, by string) (packageVersion, error) {
	cfg, err := s.packagesConfig(ctx)
	if err != nil {
		return packageVersion{}, err
	}
	if len(tarball) == 0 {
		return packageVersion{}, apierr.BadRequest("UPM_PACKAGE_EMPTY")
	}
	if int64(len(tarball)) > int64(cfg.MaxPackageMB)*1024*1024 {
		return packageVersion{}, apierr.BadRequest("UPM_PACKAGE_TOO_LARGE").With("maxMB", cfg.MaxPackageMB)
	}
	m, err := readManifest(tarball)
	if err != nil {
		return packageVersion{}, err
	}
	name, _ := m["name"].(string)
	version, _ := m["version"].(string)
	if !upmNamePattern.MatchString(name) {
		return packageVersion{}, apierr.BadRequest("UPM_PACKAGE_NAME_INVALID").WithField("name")
	}
	if len(cfg.NamePrefixes) > 0 {
		ok := false
		for _, p := range cfg.NamePrefixes {
			ok = ok || strings.HasPrefix(name, p)
		}
		if !ok {
			return packageVersion{}, apierr.BadRequest("UPM_PACKAGE_NAME_NOT_ALLOWED").With("prefixes", strings.Join(cfg.NamePrefixes, ", "))
		}
	}
	if _, ok := parseSemver(version); !ok {
		return packageVersion{}, apierr.BadRequest("UPM_PACKAGE_VERSION_INVALID").WithField("version")
	}
	var exists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM devkit_package_version WHERE name=$1 AND version=$2)`, name, version).Scan(&exists); err != nil {
		return packageVersion{}, err
	}
	if exists {
		return packageVersion{}, apierr.Conflict("UPM_PACKAGE_VERSION_EXISTS").With("name", name).With("version", version)
	}
	sum1 := sha1.Sum(tarball) //nolint:gosec
	sum512 := sha512.Sum512(tarball)
	key := "upm/" + name + "/" + upmFileName(name, version)
	if err := s.store.Put(ctx, key, tarball, "application/octet-stream"); err != nil {
		return packageVersion{}, err
	}
	raw, _ := json.Marshal(m)
	p := packageVersion{Name: name, Version: version, Manifest: raw, Shasum: hex.EncodeToString(sum1[:]), Integrity: "sha512-" + base64.StdEncoding.EncodeToString(sum512[:]),
		Size: int64(len(tarball)), StorageKey: key, PublishedBy: by}
	err = s.db.QueryRowContext(ctx, `INSERT INTO devkit_package_version (name, version, manifest, shasum, integrity, size, storage_key, published_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (name, version) DO NOTHING RETURNING created_at`, name, version, raw, p.Shasum, p.Integrity, p.Size, key, by).Scan(&p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return packageVersion{}, apierr.Conflict("UPM_PACKAGE_VERSION_EXISTS").With("name", name).With("version", version)
	}
	if err != nil {
		return packageVersion{}, err
	}
	p.fillFromManifest()
	if cfg.KeepVersions > 0 {
		s.trimVersions(ctx, name, cfg.KeepVersions)
	}
	return p, nil
}

// trimVersions removes the oldest versions of a package beyond keep (by semver).
func (s *service) trimVersions(ctx context.Context, name string, keep int) {
	list, err := s.packageVersions(ctx, name)
	if err != nil || len(list) <= keep {
		return
	}
	for _, p := range list[:len(list)-keep] {
		_ = s.deletePackage(ctx, p.Name, p.Version)
	}
}

func (s *service) deletePackage(ctx context.Context, name, version string) error {
	var key string
	err := s.db.QueryRowContext(ctx, `DELETE FROM devkit_package_version WHERE name=$1 AND version=$2 RETURNING storage_key`, name, version).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return apierr.NotFound("UPM_PACKAGE_NOT_FOUND")
	}
	if err != nil {
		return err
	}
	if err := s.store.Delete(ctx, key); err != nil && !errors.Is(err, objectstore.ErrNotFound) && s.logger != nil {
		s.logger.Warn("Package archive %s not deleted: %v", key, err)
	}
	return nil
}

// packageVersions lists versions (one package when name is set), oldest first by semver.
func (s *service) packageVersions(ctx context.Context, name string) ([]packageVersion, error) {
	q := `SELECT name, version, manifest, shasum, integrity, size, storage_key, published_by, created_at FROM devkit_package_version`
	args := []any{}
	if name != "" {
		q += ` WHERE name=$1`
		args = append(args, name)
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []packageVersion{}
	for rows.Next() {
		var p packageVersion
		if err := rows.Scan(&p.Name, &p.Version, &p.Manifest, &p.Shasum, &p.Integrity, &p.Size, &p.StorageKey, &p.PublishedBy, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.fillFromManifest()
		out = append(out, p)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		a, _ := parseSemver(out[i].Version)
		b, _ := parseSemver(out[j].Version)
		return a.less(b)
	})
	return out, rows.Err()
}

// latestOf is the highest release version, or the highest version when all are pre-releases.
func latestOf(list []packageVersion) packageVersion {
	for i := len(list) - 1; i >= 0; i-- {
		if v, _ := parseSemver(list[i].Version); v.pre == "" {
			return list[i]
		}
	}
	return list[len(list)-1]
}

func (s *service) registryURL(r *http.Request, cfg settings.Packages) string {
	if cfg.PublicURL != "" {
		return strings.TrimRight(cfg.PublicURL, "/")
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if p := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]); p == "http" || p == "https" {
		scheme = p
	}
	host := r.Host
	if h := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0]); h != "" {
		host = h
	}
	return scheme + "://" + host + upmPrefix
}

func packument(name string, list []packageVersion, base string) map[string]any {
	versions, times := map[string]any{}, map[string]any{}
	created, modified := list[0].CreatedAt, list[0].CreatedAt
	for _, p := range list {
		m := map[string]any{}
		_ = json.Unmarshal(p.Manifest, &m)
		m["_id"] = name + "@" + p.Version
		m["dist"] = map[string]any{"tarball": base + "/" + name + "/-/" + upmFileName(name, p.Version), "shasum": p.Shasum, "integrity": p.Integrity}
		versions[p.Version] = m
		times[p.Version] = p.CreatedAt.UTC().Format(time.RFC3339)
		if p.CreatedAt.Before(created) {
			created = p.CreatedAt
		}
		if p.CreatedAt.After(modified) {
			modified = p.CreatedAt
		}
	}
	times["created"], times["modified"] = created.UTC().Format(time.RFC3339), modified.UTC().Format(time.RFC3339)
	latest := latestOf(list)
	out := map[string]any{"_id": name, "name": name, "dist-tags": map[string]string{"latest": latest.Version}, "versions": versions, "time": times}
	if latest.Description != "" {
		out["description"] = latest.Description
	}
	return out
}

// upmRead runs next when the registry is on and the caller may read it.
func (s *service) upmRead(next func(w http.ResponseWriter, r *http.Request, cfg settings.Packages)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := s.packagesConfig(r.Context())
		if err != nil {
			upmError(w, r, s.db, err)
			return
		}
		if cfg.Access == settings.AccessPublic {
			next(w, r, cfg)
			return
		}
		all, _ := settings.Load(r.Context(), s.db)
		full := apikey.FromRequest(r)
		key, err := apikey.Verify(r.Context(), s.db, full, ScopeUpmRead, apikey.ClientIP(r, all.APIKeys.TrustProxyHeaders))
		if err != nil {
			// a publish key may read too
			if k2, err2 := apikey.Verify(r.Context(), s.db, full, ScopeUpmPublish, apikey.ClientIP(r, all.APIKeys.TrustProxyHeaders)); err2 == nil {
				key, err = k2, nil
			}
		}
		if err != nil {
			upmError(w, r, s.db, err)
			return
		}
		_ = key
		next(w, r, cfg)
	}
}

// upmError answers npm-style ({"error": "..."}) with the devkit code and message, and asks
// for a Bearer token on 401 so npm-style clients send the configured one.
func upmError(w http.ResponseWriter, r *http.Request, db *sql.DB, err error) {
	if ae, ok := apierr.As(err); ok && ae.Status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", `Bearer realm="packages"`)
	}
	apierr.WriteHTTP(w, r, db, err)
}

func (s *service) httpPackument(w http.ResponseWriter, r *http.Request, cfg settings.Packages) {
	name := pathVar(r, "name")
	list, err := s.packageVersions(r.Context(), name)
	if err != nil {
		upmError(w, r, s.db, err)
		return
	}
	if len(list) == 0 || !upmNamePattern.MatchString(name) {
		upmError(w, r, s.db, apierr.NotFound("UPM_PACKAGE_NOT_FOUND"))
		return
	}
	writeJSON(w, http.StatusOK, packument(name, list, s.registryURL(r, cfg)))
}

func (s *service) httpTarball(w http.ResponseWriter, r *http.Request, _ settings.Packages) {
	name, file := pathVar(r, "name"), pathVar(r, "file")
	prefix := name + "-"
	if !strings.HasPrefix(file, prefix) || !strings.HasSuffix(file, ".tgz") {
		upmError(w, r, s.db, apierr.NotFound("UPM_PACKAGE_NOT_FOUND"))
		return
	}
	version := strings.TrimSuffix(strings.TrimPrefix(file, prefix), ".tgz")
	var key string
	var size int64
	err := s.db.QueryRowContext(r.Context(), `SELECT storage_key, size FROM devkit_package_version WHERE name=$1 AND version=$2`, name, version).Scan(&key, &size)
	if errors.Is(err, sql.ErrNoRows) {
		upmError(w, r, s.db, apierr.NotFound("UPM_PACKAGE_NOT_FOUND"))
		return
	}
	if err != nil {
		upmError(w, r, s.db, err)
		return
	}
	obj, err := s.store.Get(r.Context(), key)
	if err != nil {
		if errors.Is(err, objectstore.ErrNotFound) {
			err = apierr.NotFound("UPM_PACKAGE_NOT_FOUND")
		}
		upmError(w, r, s.db, err)
		return
	}
	defer obj.Body.Close()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, obj.Body)
}

func (s *service) httpSearch(w http.ResponseWriter, r *http.Request, _ settings.Packages) {
	text := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("text")))
	list, err := s.packageVersions(r.Context(), "")
	if err != nil {
		upmError(w, r, s.db, err)
		return
	}
	byName := map[string][]packageVersion{}
	names := []string{}
	for _, p := range list {
		if text != "" && !strings.Contains(p.Name, text) {
			continue
		}
		if _, ok := byName[p.Name]; !ok {
			names = append(names, p.Name)
		}
		byName[p.Name] = append(byName[p.Name], p)
	}
	objects := []any{}
	for _, n := range names {
		l := latestOf(byName[n])
		objects = append(objects, map[string]any{"package": map[string]any{"name": n, "version": l.Version, "description": l.Description,
			"displayName": l.DisplayName, "date": l.CreatedAt.UTC().Format(time.RFC3339)}})
	}
	writeJSON(w, http.StatusOK, map[string]any{"objects": objects, "total": len(objects), "time": time.Now().UTC().Format(time.RFC3339)})
}

// httpPublish takes the archive as the body, or in parts: ?upload=<id>&part=<n>&parts=<count>
// (parts in order, each at most 192 KB; the last part publishes).
func (s *service) httpPublish(w http.ResponseWriter, r *http.Request, key *apikey.Key) {
	ctx := r.Context()
	cfg, err := s.packagesConfig(ctx)
	if err != nil {
		upmError(w, r, s.db, err)
		return
	}
	max := int64(cfg.MaxPackageMB) * 1024 * 1024
	body, err := io.ReadAll(io.LimitReader(r.Body, max+1))
	if err != nil {
		upmError(w, r, s.db, apierr.BadRequest("UPM_PACKAGE_INVALID").WithDetail(err.Error()))
		return
	}
	q := r.URL.Query()
	if upload := q.Get("upload"); upload != "" {
		part, err1 := strconv.Atoi(q.Get("part"))
		parts, err2 := strconv.Atoi(q.Get("parts"))
		if !upmUploadPattern.MatchString(upload) || err1 != nil || err2 != nil || parts < 1 || parts > upmMaxParts || part < 0 || part >= parts || len(body) > upmPartBytes || len(body) == 0 {
			upmError(w, r, s.db, apierr.BadRequest("UPM_PART_INVALID").With("partBytes", upmPartBytes))
			return
		}
		prefix := "upm/uploads/" + key.ID + "/" + upload + "/"
		if err := s.store.Put(ctx, prefix+strconv.Itoa(part), body, "application/octet-stream"); err != nil {
			upmError(w, r, s.db, err)
			return
		}
		if part < parts-1 {
			writeJSON(w, http.StatusAccepted, map[string]any{"received": part, "parts": parts})
			return
		}
		var joined bytes.Buffer
		for i := 0; i < parts; i++ {
			obj, err := s.store.Get(ctx, prefix+strconv.Itoa(i))
			if err != nil {
				upmError(w, r, s.db, apierr.BadRequest("UPM_PART_MISSING").With("part", i))
				return
			}
			_, err = io.Copy(&joined, obj.Body)
			obj.Body.Close()
			if err != nil {
				upmError(w, r, s.db, err)
				return
			}
			if int64(joined.Len()) > max {
				break
			}
		}
		for i := 0; i < parts; i++ {
			_ = s.store.Delete(ctx, prefix+strconv.Itoa(i))
		}
		body = joined.Bytes()
	}
	p, err := s.publishPackage(ctx, body, "apikey:"+key.Name)
	if err != nil {
		upmError(w, r, s.db, err)
		return
	}
	if s.logger != nil {
		s.logger.Info("Package %s@%s published by API key %s", p.Name, p.Version, key.Name)
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *service) httpPackageList(w http.ResponseWriter, r *http.Request, _ *apikey.Key) {
	list, err := s.packageVersions(r.Context(), strings.TrimSpace(r.URL.Query().Get("name")))
	if err != nil {
		upmError(w, r, s.db, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// registerPackageHTTP adds the registry (/upm) and the publish API (/v2/devkit/packages).
func (s *service) registerPackageHTTP(initializer runtime.Initializer) error {
	routes := []struct {
		path    string
		method  string
		handler http.HandlerFunc
	}{
		{upmPrefix + "/-/v1/search", http.MethodGet, s.upmRead(s.httpSearch)},
		{upmPrefix + "/{name}", http.MethodGet, s.upmRead(s.httpPackument)},
		{upmPrefix + "/{name}/-/{file}", http.MethodGet, s.upmRead(s.httpTarball)},
		{publicPrefix + "/packages", http.MethodGet, apikey.Require(s.db, ScopeUpmPublish, s.httpPackageList)},
		{publicPrefix + "/packages", http.MethodPost, apikey.Require(s.db, ScopeUpmPublish, s.httpPublish)},
	}
	for _, rt := range routes {
		if err := initializer.RegisterHttp(rt.path, rt.handler, rt.method); err != nil {
			return err
		}
	}
	return nil
}

// packageStore opens the object store for archives (STORAGE_* env; local: <data_dir>/packages).
func packageStore(env map[string]string) (objectstore.Store, error) {
	return objectstore.FromEnv(env, objectstore.Namespace{LocalDir: filepath.Join(dataDir(os.Args), "packages"), Prefix: "devkit/"})
}

func dataDir(args []string) string {
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

// Console: list, delete and the install snippets.

func (s *service) consolePackages(w http.ResponseWriter, r *http.Request) {
	cfg, _ := settings.Load(r.Context(), s.db)
	list, err := s.packageVersions(r.Context(), "")
	if err != nil {
		s.fail(w, err)
		return
	}
	url := s.registryURL(r, cfg.Packages)
	if cfg.Packages.PublicURL == "" && s.publicBase != "" {
		url = s.publicBase + upmPrefix
	}
	writeJSON(w, http.StatusOK, map[string]any{"versions": list, "registryUrl": url, "enabled": cfg.Packages.Enabled,
		"access": cfg.Packages.Access, "readScope": ScopeUpmRead, "publishScope": ScopeUpmPublish, "partBytes": upmPartBytes})
}

func (s *service) consolePackageDelete(w http.ResponseWriter, r *http.Request) {
	name, version := pathVar(r, "name"), pathVar(r, "version")
	if err := s.deletePackage(r.Context(), name, version); err != nil {
		if ae, ok := apierr.As(err); ok && ae.Status == http.StatusNotFound {
			s.fail(w, missing("UPM_PACKAGE_NOT_FOUND"))
			return
		}
		s.fail(w, err)
		return
	}
	if s.logger != nil {
		s.logger.Info("Package %s@%s deleted by %s", name, version, consoleUser(r))
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": fmt.Sprintf("%s@%s", name, version)})
}
