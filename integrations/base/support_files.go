package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
)

const (
	uploadChunkBytes = 96 * 1024
	fileURLTTL       = time.Hour
	supportFileInfo  = "support-file-url"
)

var (
	supportFileTypes   = map[string]string{"png": "image/png", "jpg": "image/jpeg", "webp": "image/webp", "pdf": "application/pdf"}
	supportFilePattern = regexp.MustCompile(`^[a-f0-9]{32}\.(png|jpg|webp|pdf)$`)
	baseBox            = secretbox.Box{Salt: "nakama-base", AAD: "base-support"}
)

type attachment struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Type string `json:"type"`
	Size int    `json:"size"`
	URL  string `json:"url,omitempty"`
}

func (s *service) configureFiles(env map[string]string, dataDir string) error {
	secret, err := secretbox.Secret(env)
	if err != nil {
		return err
	}
	key, err := baseBox.DeriveKey(secret, supportFileInfo)
	if err != nil {
		return err
	}
	cfg, err := objectstore.ConfigFromEnv(env, objectstore.Namespace{LocalDir: filepath.Join(dataDir, "support-uploads"), Prefix: "support/"})
	if err != nil {
		return err
	}
	store, err := objectstore.New(cfg, nil)
	if err != nil {
		return err
	}
	s.files, s.fileKey = store, key
	s.publicBase = strings.TrimRight(strings.TrimSpace(env["ECONOMY_PUBLIC_BASE_URL"]), "/")
	return nil
}

func sniffSupportFile(data []byte) (string, bool) {
	switch http.DetectContentType(data) {
	case "image/png":
		return "png", true
	case "image/jpeg":
		return "jpg", true
	case "image/webp":
		return "webp", true
	case "application/pdf":
		return "pdf", true
	}
	return "", false
}

func cleanFileName(name, ext string) string {
	name = strings.TrimSpace(filepath.Base(strings.ReplaceAll(name, "\\", "/")))
	name = strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune(`<>:"/\|?*`, r) {
			return -1
		}
		return r
	}, name)
	if name == "" || name == "." {
		name = "file"
	}
	for utf8.RuneCountInString(name) > 80 {
		_, size := utf8.DecodeLastRuneInString(name)
		name = name[:len(name)-size]
	}
	if !strings.HasSuffix(strings.ToLower(name), "."+ext) && !(ext == "jpg" && strings.HasSuffix(strings.ToLower(name), ".jpeg")) {
		name += "." + ext
	}
	return name
}

func (s *service) fileSig(key string, exp int64) string {
	mac := hmac.New(sha256.New, s.fileKey)
	mac.Write([]byte(key + "|" + strconv.FormatInt(exp, 10)))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *service) fileURL(key string) string {
	if s.fileKey == nil || !supportFilePattern.MatchString(key) {
		return ""
	}
	exp := s.now().Add(fileURLTTL).Unix()
	return s.publicBase + "/v2/support/files/" + key + "?e=" + strconv.FormatInt(exp, 10) + "&s=" + s.fileSig(key, exp)
}

func (s *service) signAttachments(list []attachment) []attachment {
	for i := range list {
		list[i].URL = s.fileURL(list[i].Key)
	}
	return list
}

func (s *service) fileHTTP(w http.ResponseWriter, r *http.Request) {
	key := pathVar(r, "key")
	exp, _ := strconv.ParseInt(r.URL.Query().Get("e"), 10, 64)
	sig := r.URL.Query().Get("s")
	if s.fileKey == nil || !supportFilePattern.MatchString(key) || exp < s.now().Unix() || !hmac.Equal([]byte(sig), []byte(s.fileSig(key, exp))) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var name, ctype string
	if err := s.db.QueryRowContext(r.Context(), `SELECT name, content_type FROM base_support_file WHERE key=$1`, key).Scan(&name, &ctype); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	obj, err := s.files.Get(r.Context(), key)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer obj.Body.Close()
	data, err := io.ReadAll(obj.Body)
	if err != nil {
		http.Error(w, "read failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'; img-src 'self'; style-src 'unsafe-inline'")
	w.Header().Set("Cache-Control", "private, max-age=3600")
	disp := "inline"
	if ctype == "application/pdf" {
		disp = "attachment"
	}
	w.Header().Set("Content-Disposition", disp+"; filename*=UTF-8''"+urlPathEscape(name))
	http.ServeContent(w, r, "", obj.ModTime, bytes.NewReader(data))
}

func urlPathEscape(s string) string {
	var b strings.Builder
	for _, c := range []byte(s) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '_' {
			b.WriteByte(c)
		} else {
			b.WriteString("%" + strings.ToUpper(hex.EncodeToString([]byte{c})))
		}
	}
	return b.String()
}

func (s *service) storeFile(ctx context.Context, userID *string, by, name string, data []byte, maxKB int) (attachment, error) {
	if s.files == nil || s.fileKey == nil {
		return attachment{}, fail(http.StatusServiceUnavailable, "ATTACHMENTS_UNAVAILABLE", "", "attachments need ECONOMY_SECRET_KEY")
	}
	if len(data) == 0 || len(data) > maxKB*1024 {
		return attachment{}, badRequest("ATTACHMENT_TOO_LARGE", "file", "file is too large")
	}
	ext, ok := sniffSupportFile(data)
	if !ok {
		return attachment{}, badRequest("ATTACHMENT_TYPE", "file", "only png, jpg, webp and pdf are allowed")
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return attachment{}, err
	}
	key := hex.EncodeToString(raw) + "." + ext
	if err := s.files.Put(ctx, key, data, supportFileTypes[ext]); err != nil {
		return attachment{}, err
	}
	a := attachment{Key: key, Name: cleanFileName(name, ext), Type: supportFileTypes[ext], Size: len(data)}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO base_support_file (key, user_id, uploaded_by, name, content_type, size) VALUES ($1,$2,$3,$4,$5,$6)`, key, userID, by, a.Name, a.Type, a.Size); err != nil {
		_ = s.files.Delete(ctx, key)
		return attachment{}, err
	}
	a.URL = s.fileURL(key)
	return a, nil
}

func (s *service) rpcUploadBegin(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Name string `json:"name"`
		Size int    `json:"size"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	st, err := loadSupportSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if st.MaxAttachments == 0 || s.files == nil {
		return nil, fail(http.StatusServiceUnavailable, "ATTACHMENTS_UNAVAILABLE", "", "attachments are turned off")
	}
	if in.Size <= 0 || in.Size > st.MaxAttachmentKB*1024 {
		return nil, badRequest("ATTACHMENT_TOO_LARGE", "size", "file is too large")
	}
	if err := s.allow(ctx, "sup-up:"+userID, 30, time.Hour, "TOO_MANY_UPLOADS"); err != nil {
		return nil, err
	}
	parts := (in.Size + uploadChunkBytes - 1) / uploadChunkBytes
	id := uuid.Must(uuid.NewV4()).String()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO base_support_upload (id, user_id, uploaded_by, name, size, parts) VALUES ($1,$2,'player',$3,$4,$5)`, id, userID, strings.TrimSpace(in.Name), in.Size, parts); err != nil {
		return nil, err
	}
	return map[string]any{"uploadId": id, "chunkBytes": uploadChunkBytes, "parts": parts}, nil
}

func (s *service) rpcUploadPart(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		UploadID string `json:"uploadId"`
		Index    int    `json:"index"`
		Data     string `json:"data"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(in.Data)
	if err != nil || len(data) == 0 || len(data) > uploadChunkBytes {
		return nil, badRequest("INVALID_CHUNK", "data", "chunk must be base64 of at most 96 KB")
	}
	if _, err := uuid.FromString(in.UploadID); err != nil {
		return nil, notFound("UPLOAD_NOT_FOUND")
	}
	var parts int
	err = s.db.QueryRowContext(ctx, `SELECT parts FROM base_support_upload WHERE id=$1 AND user_id=$2 AND created_at > now() - INTERVAL '1 hour'`, in.UploadID, userID).Scan(&parts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("UPLOAD_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	if in.Index < 0 || in.Index >= parts {
		return nil, badRequest("INVALID_CHUNK", "index", "chunk index is out of range")
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO base_support_upload_part (upload_id, idx, data) VALUES ($1,$2,$3) ON CONFLICT (upload_id, idx) DO UPDATE SET data=EXCLUDED.data`, in.UploadID, in.Index, data); err != nil {
		return nil, err
	}
	return map[string]any{"index": in.Index}, nil
}

func (s *service) rpcUploadFinish(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		UploadID string `json:"uploadId"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if _, err := uuid.FromString(in.UploadID); err != nil {
		return nil, notFound("UPLOAD_NOT_FOUND")
	}
	st, err := loadSupportSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	var name string
	var size, parts int
	err = s.db.QueryRowContext(ctx, `SELECT name, size, parts FROM base_support_upload WHERE id=$1 AND user_id=$2`, in.UploadID, userID).Scan(&name, &size, &parts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("UPLOAD_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT idx, data FROM base_support_upload_part WHERE upload_id=$1 ORDER BY idx`, in.UploadID)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	n := 0
	for rows.Next() {
		var idx int
		var data []byte
		if err := rows.Scan(&idx, &data); err != nil {
			rows.Close()
			return nil, err
		}
		if idx != n {
			rows.Close()
			return nil, badRequest("UPLOAD_INCOMPLETE", "", "some chunks are missing")
		}
		buf.Write(data)
		n++
	}
	rows.Close()
	if n != parts || buf.Len() != size {
		return nil, badRequest("UPLOAD_INCOMPLETE", "", "some chunks are missing")
	}
	uid := userID
	a, err := s.storeFile(ctx, &uid, "player", name, buf.Bytes(), st.MaxAttachmentKB)
	if err != nil {
		return nil, err
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM base_support_upload WHERE id=$1`, in.UploadID)
	return a, nil
}

func (s *service) consoleSupportUpload(w http.ResponseWriter, r *http.Request) {
	st, err := loadSupportSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	limit := int64(st.MaxAttachmentKB) * 1024
	r.Body = http.MaxBytesReader(w, r.Body, limit+64*1024)
	if err := r.ParseMultipartForm(limit + 64*1024); err != nil {
		s.writeError(w, badRequest("ATTACHMENT_TOO_LARGE", "file", "file is too large"))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		s.writeError(w, badRequest("FILE_REQUIRED", "file", "the file field is required"))
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		s.writeError(w, badRequest("INVALID_UPLOAD", "file", "cannot read file"))
		return
	}
	a, err := s.storeFile(r.Context(), nil, "staff:"+agentName(r), header.Filename, data, st.MaxAttachmentKB)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (s *service) claimAttachments(ctx context.Context, tx *sql.Tx, keys []string, userID string, staff bool, max int) ([]attachment, error) {
	out := []attachment{}
	seen := map[string]bool{}
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		if !supportFilePattern.MatchString(k) {
			return nil, badRequest("ATTACHMENT_NOT_FOUND", "attachments", "attachment does not exist")
		}
		out = append(out, attachment{Key: k})
	}
	if len(out) > max {
		return nil, badRequest("TOO_MANY_ATTACHMENTS", "attachments", "too many attachments")
	}
	for i := range out {
		cond := `user_id=$2`
		args := []any{out[i].Key, userID}
		if staff {
			cond = `user_id IS NULL AND uploaded_by LIKE 'staff:%' AND $2::text <> ''`
		}
		err := tx.QueryRowContext(ctx, `UPDATE base_support_file SET attached=true WHERE key=$1 AND NOT attached AND `+cond+` RETURNING name, content_type, size`, args...).Scan(&out[i].Name, &out[i].Type, &out[i].Size)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, badRequest("ATTACHMENT_NOT_FOUND", "attachments", "attachment does not exist")
		}
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *service) sweepSupportFiles(ctx context.Context) (int, error) {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM base_support_upload WHERE created_at < now() - INTERVAL '1 day'`); err != nil {
		return 0, err
	}
	rows, err := s.db.QueryContext(ctx, `DELETE FROM base_support_file WHERE NOT attached AND created_at < now() - INTERVAL '1 day' RETURNING key`)
	if err != nil {
		return 0, err
	}
	keys := []string{}
	for rows.Next() {
		var k string
		if rows.Scan(&k) == nil {
			keys = append(keys, k)
		}
	}
	rows.Close()
	for _, k := range keys {
		if s.files != nil {
			_ = s.files.Delete(ctx, k)
		}
	}
	return len(keys), nil
}
