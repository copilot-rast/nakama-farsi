package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

const (
	CodePlayerCodeNotFound = "PLAYER_CODE_NOT_FOUND"
	playerCodesMax         = 100
)

var playerCharsets = map[string]string{
	"alnum":   "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	"letters": "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	"digits":  "0123456789",
}

var playerCodePattern = regexp.MustCompile(`^[A-Z0-9]{4,20}$`)

type playerCodeSettings struct {
	Length   int    `json:"length"`
	Charset  string `json:"charset"`
	Alphabet string `json:"alphabet"`
}

func defaultPlayerCodeSettings() playerCodeSettings {
	return playerCodeSettings{Length: 6, Charset: "alnum", Alphabet: playerCharsets["alnum"]}
}

var playerCodeStore = settingsrow.New("base_player_code_settings", defaultPlayerCodeSettings)

func (p playerCodeSettings) clean() (playerCodeSettings, error) {
	if p.Charset == "" {
		p.Charset = "alnum"
	}
	alphabet, ok := playerCharsets[p.Charset]
	switch {
	case !ok:
		return p, badRequest("INVALID_RULE", "charset", "alnum, letters or digits")
	case p.Length < 4 || p.Length > 16:
		return p, badRequest("INVALID_RULE", "length", "between 4 and 16")
	case p.Charset == "digits" && p.Length < 6:
		return p, badRequest("INVALID_RULE", "length", "digit codes need at least 6 characters")
	}
	p.Alphabet = alphabet
	return p, nil
}

func normalizePlayerCode(v string) string {
	return strings.ToUpper(strings.TrimSpace(v))
}

func (s *service) playerCode(ctx context.Context, userID string) (string, error) {
	var code string
	err := s.db.QueryRowContext(ctx, `SELECT code FROM base_player_code WHERE user_id=$1`, userID).Scan(&code)
	if err == nil {
		return code, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	err = s.db.QueryRowContext(ctx, `INSERT INTO base_player_code (user_id, code) SELECT $1, base_new_player_code() WHERE EXISTS (SELECT 1 FROM users WHERE id=$1)
		ON CONFLICT (user_id) DO UPDATE SET user_id=EXCLUDED.user_id RETURNING code`, userID).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", notFound("USER_NOT_FOUND")
	}
	return code, err
}

func (s *service) codesFor(ctx context.Context, ids []string) (map[string]string, error) {
	out := map[string]string{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text, code FROM base_player_code WHERE user_id = ANY($1::uuid[])`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, code string
		if err := rows.Scan(&id, &code); err != nil {
			return nil, err
		}
		out[id] = code
	}
	return out, rows.Err()
}

func (s *service) idsForCodes(ctx context.Context, userID string, codes []string) ([]string, error) {
	clean := make([]string, 0, len(codes))
	for _, c := range codes {
		c = normalizePlayerCode(c)
		if !playerCodePattern.MatchString(c) {
			return nil, badRequest(CodePlayerCodeNotFound, "codes", "not a player code")
		}
		clean = append(clean, c)
	}
	if len(clean) == 0 {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text, code FROM base_player_code WHERE code = ANY($1::text[])`, clean)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	found := map[string]string{}
	for rows.Next() {
		var id, code string
		if err := rows.Scan(&id, &code); err != nil {
			return nil, err
		}
		found[code] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := []string{}
	for _, c := range clean {
		id, ok := found[c]
		if !ok {
			return nil, notFound(CodePlayerCodeNotFound)
		}
		if id == userID {
			return nil, badRequest("CANNOT_TARGET_SELF", "codes", "this is your own account")
		}
		out = append(out, id)
	}
	return out, nil
}

func (s *service) rpcMyCode(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	code, err := s.playerCode(ctx, userID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"code": code}, nil
}

func (s *service) rpcPlayerCodes(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		IDs []string `json:"ids"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if len(in.IDs) == 0 || len(in.IDs) > playerCodesMax {
		return nil, badRequest("INVALID_USER_ID", "ids", "1 to 100 user ids")
	}
	for _, id := range in.IDs {
		if _, err := uuid.FromString(id); err != nil {
			return nil, badRequest("INVALID_USER_ID", "ids", "not a user ID")
		}
	}
	codes, err := s.codesFor(ctx, in.IDs)
	if err != nil {
		return nil, err
	}
	return map[string]any{"codes": codes}, nil
}

type relationView struct {
	Relationship string `json:"relationship"`
	Online       *bool  `json:"online,omitempty"`
}

func (s *service) relationship(ctx context.Context, userID, otherID string, banned bool) (relationView, error) {
	if otherID == userID {
		on := true
		return relationView{Relationship: "self", Online: &on}, nil
	}
	if banned {
		return relationView{Relationship: "unavailable"}, nil
	}
	var mine, theirs int
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE((SELECT state FROM user_edge WHERE source_id=$1 AND destination_id=$2), -1),
		COALESCE((SELECT state FROM user_edge WHERE source_id=$2 AND destination_id=$1), -1)`, userID, otherID).Scan(&mine, &theirs)
	if err != nil {
		return relationView{}, err
	}
	out := relationView{Relationship: "none"}
	switch {
	case mine == edgeBlocked:
		out.Relationship = "blocked"
	case theirs == edgeBlocked:
		out.Relationship = "blocked_by"
	case mine == edgeFriend:
		out.Relationship = "friends"
		on := s.isOnline(otherID)
		out.Online = &on
	case mine == edgeSent:
		out.Relationship = "sent"
	case mine == edgeReceived:
		out.Relationship = "received"
	}
	return out, nil
}

func (s *service) isOnline(userID string) bool {
	if s.nk == nil {
		return false
	}
	presences, err := s.nk.StreamUserList(0, userID, "", "", true, true)
	return err == nil && len(presences) > 0
}

func (s *service) rpcUserByCode(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Code string `json:"code"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	code := normalizePlayerCode(in.Code)
	if !playerCodePattern.MatchString(code) {
		return nil, badRequest(CodePlayerCodeNotFound, "code", "not a player code")
	}
	var id, username, display, avatar string
	var banned bool
	err := s.db.QueryRowContext(ctx, `SELECT u.id::text, u.username, COALESCE(u.display_name,''), COALESCE(u.avatar_url,''), u.disable_time > '1970-01-01 00:00:00+00'
		FROM base_player_code c JOIN users u ON u.id=c.user_id WHERE c.code=$1`, code).Scan(&id, &username, &display, &avatar, &banned)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound(CodePlayerCodeNotFound)
	}
	if err != nil {
		return nil, err
	}
	rel, err := s.relationship(ctx, userID, id, banned)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"user":         map[string]any{"id": id, "username": username, "displayName": display, "avatarUrl": avatar, "code": code},
		"relationship": rel.Relationship,
	}
	if rel.Online != nil {
		out["online"] = *rel.Online
	}
	return out, nil
}

func (s *service) playerCodeSettingsGet(w http.ResponseWriter, r *http.Request) {
	v, meta, err := playerCodeStore.Read(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var total int64
	_ = s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM base_player_code`).Scan(&total)
	writeJSON(w, http.StatusOK, map[string]any{"length": v.Length, "charset": v.Charset, "codes": total, "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt})
}

func (s *service) playerCodeSettingsPut(w http.ResponseWriter, r *http.Request) {
	in := defaultPlayerCodeSettings()
	var raw map[string]json.RawMessage
	if err := readBody(r, &raw); err != nil {
		s.writeError(w, err)
		return
	}
	for _, k := range []string{"updatedBy", "updatedAt", "codes", "alphabet"} {
		delete(raw, k)
	}
	b, _ := json.Marshal(raw)
	if err := json.Unmarshal(b, &in); err != nil {
		s.writeError(w, badRequest("INVALID_JSON", "", "settings are not valid"))
		return
	}
	v, err := in.clean()
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := playerCodeStore.Save(r.Context(), s.db, v, consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	s.playerCodeSettingsGet(w, r)
}
