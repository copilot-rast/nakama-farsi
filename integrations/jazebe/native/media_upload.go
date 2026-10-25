package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const mediaChunkBytes = 128 * 1024

func mediaUploadPerMinute() (int, error) {
	raw := strings.TrimSpace(os.Getenv("JAZEBE_MEDIA_UPLOAD_PER_MINUTE"))
	if raw == "" {
		return 20, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > 10000 {
		return 0, status.Error(codes.InvalidArgument, "invalid JAZEBE_MEDIA_UPLOAD_PER_MINUTE")
	}
	return limit, nil
}

type mediaUploadBeginInput struct {
	PlanetID      string `json:"planetId"`
	Purpose       string `json:"purpose"`
	FileName      string `json:"fileName"`
	ExpectedBytes int    `json:"expectedBytes"`
}

func mediaPurposeLimit(purpose string) int {
	switch purpose {
	case "evidence":
		return 25 * 1024 * 1024
	case "avatar":
		return 5 * 1024 * 1024
	case "mission_media":
		return 50 * 1024 * 1024
	default:
		return 0
	}
}

func validateMediaUploadBegin(input mediaUploadBeginInput) (mediaUploadBeginInput, error) {
	var err error
	input.PlanetID, err = requiredUUID(input.PlanetID)
	if err != nil {
		return input, err
	}
	limit := mediaPurposeLimit(input.Purpose)
	if limit == 0 || input.ExpectedBytes < 1 || input.ExpectedBytes > limit {
		return input, status.Error(codes.InvalidArgument, "invalid media purpose or size")
	}
	input.FileName = path.Base(strings.ReplaceAll(strings.TrimSpace(input.FileName), `\`, "/"))
	if input.FileName == "." || input.FileName == "/" || !utf8.ValidString(input.FileName) ||
		utf8.RuneCountInString(input.FileName) < 1 || utf8.RuneCountInString(input.FileName) > 200 ||
		strings.ContainsAny(input.FileName, "\x00\r\n") {
		return input, status.Error(codes.InvalidArgument, "invalid media file name")
	}
	return input, nil
}

// The authenticated RPC issues a short-lived capability for one upload. The
// custom HTTP routes do not inherit Nakama's RPC session middleware.
func mediaUploadBegin(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input mediaUploadBeginInput
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	input, err = validateMediaUploadBegin(input)
	if err != nil {
		return "", err
	}
	var allowed bool
	if input.Purpose == "mission_media" {
		err = db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet p JOIN jazebe_active_brand_admin a
			ON a.brand_id=p.brand_id WHERE p.id=$1 AND a.user_id=$2 AND a.is_active)`, input.PlanetID, user).Scan(&allowed)
	} else {
		err = db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet p JOIN jazebe_planet_membership m
			ON m.planet_id=p.id WHERE p.id=$1 AND p.status='active' AND m.user_id=$2)`, input.PlanetID, user).Scan(&allowed)
	}
	if err != nil {
		return "", status.Error(codes.Internal, "upload authorization unavailable")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "upload purpose is not available to this user")
	}
	limit, err := mediaUploadPerMinute()
	if err != nil {
		return "", err
	}
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return "", status.Error(codes.Internal, "upload ticket unavailable")
	}
	token := base64.RawURLEncoding.EncodeToString(secret[:])
	digest := sha256.Sum256([]byte(token))
	id := uuid.Must(uuid.NewV4()).String()
	expires := time.Now().UTC().Add(15 * time.Minute)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "upload ticket unavailable")
	}
	defer tx.Rollback()
	// Serialize ticket issuance per Nakama user across all Nakama nodes.
	if err := lockXact(ctx, tx, "774102:"+user); err != nil {
		return "", status.Error(codes.Internal, "upload rate limit unavailable")
	}
	var recent int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_media_upload
		WHERE owner_user_id=$1 AND created_at >= now() - interval '60 seconds'`, user).Scan(&recent); err != nil {
		return "", status.Error(codes.Internal, "upload rate limit unavailable")
	}
	if recent >= limit {
		return "", status.Error(codes.ResourceExhausted, "media upload rate limit exceeded")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_media_upload
		(id,planet_id,owner_user_id,purpose,token_hash,file_name,expected_bytes,expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, id, input.PlanetID, user, input.Purpose, digest[:], input.FileName, input.ExpectedBytes, expires)
	if err != nil {
		logger.Error("Media upload ticket insert: %v", err)
		return "", status.Error(codes.Internal, "upload ticket unavailable")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "upload ticket unavailable")
	}
	return encode(map[string]any{"uploadId": id, "token": token, "chunkBytes": mediaChunkBytes,
		"chunkPath": "/v2/jazebe/media/" + id + "/chunks", "completePath": "/v2/jazebe/media/" + id + "/complete",
		"expiresAt": expires})
}

// A member may withdraw only their own upload. Existing completion records
// retain their evidence metadata; the daily sweep preserves referenced bytes.
func mediaUploadDelete(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		ID string `json:"id"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	id, err := requiredUUID(input.ID)
	if err != nil {
		return "", err
	}
	result, err := db.ExecContext(ctx, `UPDATE jazebe_media_upload SET status='deleted',deleted_at=COALESCE(deleted_at,now()),token_hash=$3
		WHERE id=$1 AND owner_user_id=$2`, id, user, []byte{})
	if err != nil {
		return "", status.Error(codes.Internal, "media withdrawal failed")
	}
	count, err := result.RowsAffected()
	if err != nil {
		return "", status.Error(codes.Internal, "media withdrawal failed")
	}
	if count == 0 {
		return "", status.Error(codes.NotFound, "media not found")
	}
	return encode(map[string]any{"success": true})
}

func sniffMedia(bytes []byte) (mime, kind string) {
	if len(bytes) >= 8 && string(bytes[:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png", "image"
	}
	if len(bytes) >= 3 && bytes[0] == 0xff && bytes[1] == 0xd8 && bytes[2] == 0xff {
		return "image/jpeg", "image"
	}
	if len(bytes) >= 6 && (string(bytes[:6]) == "GIF87a" || string(bytes[:6]) == "GIF89a") {
		return "image/gif", "image"
	}
	if len(bytes) >= 12 && string(bytes[:4]) == "RIFF" && string(bytes[8:12]) == "WEBP" {
		return "image/webp", "image"
	}
	if len(bytes) >= 8 && string(bytes[4:8]) == "ftyp" {
		return "video/mp4", "video"
	}
	return "", ""
}

func mediaHTTPError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func mediaTicket(tx *sql.Tx, ctx context.Context, id, token string) (purpose string, expected, received, chunks int, err error) {
	if len(token) != 43 {
		return "", 0, 0, 0, sql.ErrNoRows
	}
	var digest []byte
	var expires time.Time
	var state string
	err = tx.QueryRowContext(ctx, `SELECT purpose,expected_bytes,received_bytes,received_chunks,token_hash,expires_at,status
		FROM jazebe_media_upload WHERE id=$1 FOR UPDATE`, id).Scan(&purpose, &expected, &received, &chunks, &digest, &expires, &state)
	if err != nil {
		return "", 0, 0, 0, err
	}
	want := sha256.Sum256([]byte(token))
	if subtle.ConstantTimeCompare(digest, want[:]) != 1 || !time.Now().UTC().Before(expires) || state != "open" {
		return "", 0, 0, 0, sql.ErrNoRows
	}
	return purpose, expected, received, chunks, nil
}

func mediaChunkHTTP(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		id, err := requiredUUID(mux.Vars(r)["id"])
		if err != nil {
			mediaHTTPError(w, http.StatusNotFound, "upload not found")
			return
		}
		index, err := strconv.Atoi(r.Header.Get("X-Chunk-Index"))
		if err != nil || index < 0 {
			mediaHTTPError(w, http.StatusBadRequest, "invalid chunk index")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, mediaChunkBytes)
		data, err := io.ReadAll(r.Body)
		if err != nil || len(data) == 0 || len(data) > mediaChunkBytes {
			mediaHTTPError(w, http.StatusRequestEntityTooLarge, "invalid chunk size")
			return
		}
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			mediaHTTPError(w, 500, "upload unavailable")
			return
		}
		defer tx.Rollback()
		_, expected, received, chunks, err := mediaTicket(tx, r.Context(), id, r.Header.Get("X-Jazebe-Upload-Token"))
		if err != nil {
			mediaHTTPError(w, http.StatusNotFound, "upload not found")
			return
		}
		if index != chunks || received+len(data) > expected {
			mediaHTTPError(w, http.StatusConflict, "unexpected chunk or size")
			return
		}
		if _, err := tx.ExecContext(r.Context(), `INSERT INTO jazebe_media_chunk(upload_id,seq,data,size_bytes)
			VALUES ($1,$2,$3,$4)`, id, index, data, len(data)); err != nil {
			mediaHTTPError(w, 500, "upload failed")
			return
		}
		if _, err := tx.ExecContext(r.Context(), `UPDATE jazebe_media_upload SET received_bytes=received_bytes+$2,
			received_chunks=received_chunks+1 WHERE id=$1`, id, len(data)); err != nil || tx.Commit() != nil {
			mediaHTTPError(w, 500, "upload failed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"receivedBytes": received + len(data), "nextChunk": index + 1})
	}
}

func mediaCompleteHTTP(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requiredUUID(mux.Vars(r)["id"])
		if err != nil {
			mediaHTTPError(w, 404, "upload not found")
			return
		}
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			mediaHTTPError(w, 500, "upload unavailable")
			return
		}
		defer tx.Rollback()
		purpose, expected, received, _, err := mediaTicket(tx, r.Context(), id, r.Header.Get("X-Jazebe-Upload-Token"))
		if err != nil {
			mediaHTTPError(w, 404, "upload not found")
			return
		}
		if received != expected {
			mediaHTTPError(w, 409, "upload incomplete")
			return
		}
		rows, err := tx.QueryContext(r.Context(), `SELECT data FROM jazebe_media_chunk WHERE upload_id=$1 ORDER BY seq`, id)
		if err != nil {
			mediaHTTPError(w, 500, "upload unavailable")
			return
		}
		hash := sha256.New()
		var first []byte
		var total int
		for rows.Next() {
			var data []byte
			if err := rows.Scan(&data); err != nil {
				rows.Close()
				mediaHTTPError(w, 500, "upload unavailable")
				return
			}
			if first == nil {
				first = append([]byte(nil), data...)
			}
			total += len(data)
			_, _ = hash.Write(data)
		}
		err = rows.Err()
		rows.Close()
		if err != nil || total != expected {
			mediaHTTPError(w, 500, "upload integrity failed")
			return
		}
		mime, kind := sniffMedia(first)
		if kind == "" || (purpose == "avatar" && kind != "image") {
			mediaHTTPError(w, 422, "unsupported media type")
			return
		}
		if _, err := tx.ExecContext(r.Context(), `UPDATE jazebe_media_upload SET status='complete',mime_type=$2,
			sha256=$3,completed_at=now(),token_hash=$4 WHERE id=$1`, id, mime, hex.EncodeToString(hash.Sum(nil)), []byte{}); err != nil || tx.Commit() != nil {
			mediaHTTPError(w, 500, "upload finalization failed")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "url": "/v2/jazebe/media/" + id,
			"mimeType": mime, "sizeBytes": expected, "purpose": purpose})
	}
}

func mediaReadHTTP(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := requiredUUID(mux.Vars(r)["id"])
		if err != nil {
			mediaHTTPError(w, 404, "media not found")
			return
		}
		var mime string
		var size int
		err = db.QueryRowContext(r.Context(), `SELECT mime_type,expected_bytes FROM jazebe_media_upload
			WHERE id=$1 AND status='complete' AND deleted_at IS NULL`, id).Scan(&mime, &size)
		if err != nil {
			mediaHTTPError(w, 404, "media not found")
			return
		}
		start, end, partial := 0, size-1, false
		if raw := r.Header.Get("Range"); raw != "" {
			if !strings.HasPrefix(raw, "bytes=") || strings.Count(raw, "-") != 1 || strings.Contains(raw, ",") {
				mediaHTTPError(w, 416, "invalid range")
				return
			}
			parts := strings.SplitN(strings.TrimPrefix(raw, "bytes="), "-", 2)
			if parts[0] != "" {
				start, err = strconv.Atoi(parts[0])
				if err != nil {
					mediaHTTPError(w, 416, "invalid range")
					return
				}
			}
			if parts[1] != "" {
				end, err = strconv.Atoi(parts[1])
				if err != nil {
					mediaHTTPError(w, 416, "invalid range")
					return
				}
			}
			if parts[0] == "" {
				suffix, e := strconv.Atoi(parts[1])
				if e != nil || suffix < 1 {
					mediaHTTPError(w, 416, "invalid range")
					return
				}
				start = size - suffix
				if start < 0 {
					start = 0
				}
				end = size - 1
			}
			if end >= size {
				end = size - 1
			}
			if start < 0 || start >= size || end < start {
				mediaHTTPError(w, 416, "invalid range")
				return
			}
			partial = true
		}
		rows, err := db.QueryContext(r.Context(), `SELECT data FROM jazebe_media_chunk WHERE upload_id=$1 ORDER BY seq`, id)
		if err != nil {
			mediaHTTPError(w, 500, "media unavailable")
			return
		}
		defer rows.Close()
		w.Header().Set("Content-Type", mime)
		w.Header().Set("Content-Disposition", "inline")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "private, max-age=86400")
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", strconv.Itoa(end-start+1))
		if partial {
			w.Header().Set("Content-Range", "bytes "+strconv.Itoa(start)+"-"+strconv.Itoa(end)+"/"+strconv.Itoa(size))
			w.WriteHeader(206)
		}
		position := 0
		for rows.Next() {
			var data []byte
			if rows.Scan(&data) != nil {
				return
			}
			chunkEnd := position + len(data) - 1
			if chunkEnd >= start && position <= end {
				from, to := 0, len(data)
				if start > position {
					from = start - position
				}
				if end < chunkEnd {
					to = end - position + 1
				}
				if _, err := w.Write(data[from:to]); err != nil {
					return
				}
			}
			position += len(data)
			if position > end {
				break
			}
		}
	}
}

func registerMediaHTTP(initializer runtime.Initializer, db *sql.DB) error {
	for _, route := range []struct {
		path, method string
		handler      http.HandlerFunc
	}{
		{"/v2/jazebe/media/{id}/chunks", http.MethodPost, mediaChunkHTTP(db)},
		{"/v2/jazebe/media/{id}/complete", http.MethodPost, mediaCompleteHTTP(db)},
		{"/v2/jazebe/media/{id}", http.MethodGet, mediaReadHTTP(db)},
	} {
		if err := initializer.RegisterHttp(route.path, route.handler, route.method); err != nil {
			return err
		}
	}
	return nil
}
