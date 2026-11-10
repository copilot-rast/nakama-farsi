package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/objectstore"
)

const maxUploadBytes = 2 << 20

var fileTypes = map[string]string{"png": "image/png", "jpg": "image/jpeg", "webp": "image/webp", "svg": "image/svg+xml"}

var (
	fileNamePattern = regexp.MustCompile(`^[a-f0-9]{32}\.(png|jpg|webp|svg)$`)
	svgDanger       = regexp.MustCompile(`(?i)<script|<foreignobject|\son[a-z]+\s*=|javascript:|<iframe|<embed|<object`)
)

func sniffImage(data []byte) (string, bool) {
	switch http.DetectContentType(data) {
	case "image/png":
		return "png", true
	case "image/jpeg":
		return "jpg", true
	case "image/webp":
		return "webp", true
	}
	head := strings.ToLower(string(data[:min(len(data), 1024)]))
	trimmed := strings.TrimSpace(strings.TrimPrefix(head, "\ufeff"))
	if (strings.HasPrefix(trimmed, "<?xml") || strings.HasPrefix(trimmed, "<svg") || strings.HasPrefix(trimmed, "<!--")) && bytes.Contains(bytes.ToLower(data), []byte("<svg")) {
		if svgDanger.Match(data) {
			return "", false
		}
		return "svg", true
	}
	return "", false
}

func (s *service) consoleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+64*1024)
	if err := r.ParseMultipartForm(maxUploadBytes + 64*1024); err != nil {
		s.writeError(w, badRequest("INVALID_UPLOAD", "file", "send multipart/form-data with a file field of at most 2 MB"))
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		s.writeError(w, badRequest("FILE_REQUIRED", "file", "the file field is required"))
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil {
		s.writeError(w, badRequest("INVALID_UPLOAD", "file", "cannot read file"))
		return
	}
	if len(data) == 0 || len(data) > maxUploadBytes {
		s.writeError(w, badRequest("FILE_TOO_LARGE", "file", "files must be at most 2 MB"))
		return
	}
	ext, ok := sniffImage(data)
	if !ok {
		s.writeError(w, badRequest("UNSUPPORTED_FILE_TYPE", "file", "only png, jpg, webp and safe svg are allowed"))
		return
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		s.writeError(w, err)
		return
	}
	name := hex.EncodeToString(raw) + "." + ext
	if err := s.store.Put(r.Context(), name, data, fileTypes[ext]); err != nil {
		s.logger.Error("Economy upload to %s storage failed: %v", s.store.Driver(), err)
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"url": s.cfg.publicBase + "/v2/economy/files/" + name, "name": name})
}

func (s *service) fileHTTP(w http.ResponseWriter, r *http.Request) {
	name := pathVar(r, "name")
	if !fileNamePattern.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	if target := s.store.PublicURL(name); target != "" {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.Redirect(w, r, target, http.StatusFound)
		return
	}
	obj, err := s.store.Get(r.Context(), name)
	if err != nil {
		if !errors.Is(err, objectstore.ErrNotFound) {
			s.logger.Warn("Economy file %s from %s storage failed: %v", name, s.store.Driver(), err)
		}
		http.NotFound(w, r)
		return
	}
	defer obj.Body.Close()
	content, ok := obj.Body.(io.ReadSeeker)
	if !ok {
		data, err := io.ReadAll(io.LimitReader(obj.Body, maxUploadBytes+1))
		if err != nil || len(data) > maxUploadBytes {
			http.NotFound(w, r)
			return
		}
		content = bytes.NewReader(data)
	}
	w.Header().Set("Content-Type", fileTypes[strings.TrimPrefix(path.Ext(name), ".")])
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(w, r, name, obj.ModTime, content)
}
