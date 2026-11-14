package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/restrict"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
	"github.com/heroiclabs/nakama/v3/internal/voice"
	"google.golang.org/grpc/codes"
)

const (
	voiceTokenLimit  = 30
	voiceTokenWindow = time.Minute
	voiceTestLimit   = 10
	voiceTestWindow  = time.Minute
	voicePlayersMax  = 100

	streamModeMatchRelayed       = 5
	streamModeMatchAuthoritative = 6
	streamModeParty              = 7
)

var (
	matchIDPattern = regexp.MustCompile(`^[0-9a-fA-F-]{36}\.[A-Za-z0-9_.-]{0,64}$`)
	roomKeyPattern = regexp.MustCompile(`^[a-z0-9_-]{1,48}$`)
	gameKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)
)

type voiceSettings struct {
	Enabled    bool              `json:"enabled"`
	Provider   string            `json:"provider"`
	Config     map[string]string `json:"config"`
	Secrets    map[string]string `json:"secrets"`
	TTLSeconds int               `json:"ttlSeconds"`
	OpenRooms  bool              `json:"openRooms"`
	CanPublish bool              `json:"canPublish"`
	UpdatedBy  string            `json:"updatedBy"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}

func defaultVoiceSettings() voiceSettings {
	return voiceSettings{Config: map[string]string{}, Secrets: map[string]string{}, TTLSeconds: 600, CanPublish: true}
}

func (s *service) configureVoice(env map[string]string) {
	s.voiceHTTP = &http.Client{Timeout: 5 * time.Second}
	if secret, err := secretbox.Secret(env); err == nil {
		if key, err := secretbox.BaseVoice.DeriveKey(secret, secretbox.BaseVoiceInfo); err == nil {
			s.voiceKey = key
		}
	}
}

func (s *service) loadVoiceSettings(ctx context.Context, q queryer) (voiceSettings, error) {
	out := defaultVoiceSettings()
	var cfg, sealed string
	err := q.QueryRowContext(ctx, `SELECT enabled, provider, config::text, secrets_sealed, ttl_seconds, open_rooms, can_publish, updated_by, updated_at FROM base_voice_settings WHERE id=1`).
		Scan(&out.Enabled, &out.Provider, &cfg, &sealed, &out.TTLSeconds, &out.OpenRooms, &out.CanPublish, &out.UpdatedBy, &out.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	_ = json.Unmarshal([]byte(cfg), &out.Config)
	if out.Config == nil {
		out.Config = map[string]string{}
	}
	if sealed != "" {
		if s.voiceKey == nil {
			return out, errVoiceSeal
		}
		raw, err := secretbox.BaseVoice.Open(s.voiceKey, sealed)
		if err != nil {
			return out, err
		}
		_ = json.Unmarshal(raw, &out.Secrets)
	}
	if out.Secrets == nil {
		out.Secrets = map[string]string{}
	}
	return out, nil
}

var errVoiceSeal = errors.New("voice secrets are sealed but ECONOMY_SECRET_KEY is missing")

func (v voiceSettings) view() map[string]any {
	masked := map[string]string{}
	for k, val := range v.Secrets {
		if val != "" {
			masked[k] = secretMask
		}
	}
	return map[string]any{"enabled": v.Enabled, "provider": v.Provider, "config": v.Config, "secrets": masked, "ttlSeconds": v.TTLSeconds,
		"openRooms": v.OpenRooms, "canPublish": v.CanPublish, "updatedBy": v.UpdatedBy, "updatedAt": v.UpdatedAt}
}

func voiceProviders() []map[string]any {
	out := []map[string]any{}
	for _, spec := range voice.Kinds() {
		out = append(out, map[string]any{"kind": spec.Kind, "fields": spec.Fields})
	}
	return out
}

func (s *service) voiceSettingsGet(w http.ResponseWriter, r *http.Request) {
	v, err := s.loadVoiceSettings(r.Context(), s.db)
	if err != nil && !errors.Is(err, errVoiceSeal) {
		s.writeError(w, err)
		return
	}
	out := v.view()
	out["providers"] = voiceProviders()
	out["canSeal"] = s.voiceKey != nil
	writeJSON(w, http.StatusOK, out)
}

func (s *service) voiceMerge(old voiceSettings, in voiceSettings) (voiceSettings, error) {
	in.Provider = strings.TrimSpace(in.Provider)
	out := voiceSettings{Enabled: in.Enabled, Provider: in.Provider, Config: map[string]string{}, Secrets: map[string]string{}, TTLSeconds: in.TTLSeconds, OpenRooms: in.OpenRooms, CanPublish: in.CanPublish}
	if out.TTLSeconds < 60 || out.TTLSeconds > 86400 {
		return out, badRequest("INVALID_RULE", "ttlSeconds", "between 60 and 86400 seconds")
	}
	if out.Provider == "" {
		if out.Enabled {
			return out, badRequest("VOICE_PROVIDER_REQUIRED", "provider", "pick a voice server first")
		}
		return out, nil
	}
	spec, ok := voice.Lookup(out.Provider)
	if !ok {
		return out, badRequest("INVALID_VOICE_PROVIDER", "provider", "unknown voice provider")
	}
	for _, f := range spec.Fields {
		if f.Secret {
			v := strings.TrimSpace(in.Secrets[f.Name])
			if v == secretMask {
				if old.Provider == out.Provider {
					v = old.Secrets[f.Name]
				} else {
					v = ""
				}
			}
			if v != "" {
				out.Secrets[f.Name] = v
			}
			continue
		}
		if v := strings.TrimSpace(in.Config[f.Name]); v != "" {
			out.Config[f.Name] = v
		}
	}
	if err := voice.Validate(out.Provider, voice.Config{Values: out.Config, Secrets: out.Secrets}); err != nil {
		var fe *voice.FieldError
		if errors.As(err, &fe) {
			return out, badRequest(fe.Code, fe.Field, fe.Message)
		}
		return out, err
	}
	if len(out.Secrets) > 0 && s.voiceKey == nil {
		return out, badRequest("VOICE_SEAL_UNAVAILABLE", "provider", "ECONOMY_SECRET_KEY is needed to keep provider secrets")
	}
	return out, nil
}

func readVoiceInput(r *http.Request) (voiceSettings, error) {
	in := defaultVoiceSettings()
	if err := readBody(r, &in); err != nil {
		return in, err
	}
	if in.Config == nil {
		in.Config = map[string]string{}
	}
	if in.Secrets == nil {
		in.Secrets = map[string]string{}
	}
	return in, nil
}

func (s *service) voiceSettingsPut(w http.ResponseWriter, r *http.Request) {
	in, err := readVoiceInput(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	old, err := s.loadVoiceSettings(r.Context(), s.db)
	if err != nil && !errors.Is(err, errVoiceSeal) {
		s.writeError(w, err)
		return
	}
	next, err := s.voiceMerge(old, in)
	if err != nil {
		s.writeError(w, err)
		return
	}
	sealed := ""
	if len(next.Secrets) > 0 {
		raw, _ := json.Marshal(next.Secrets)
		if sealed, err = secretbox.BaseVoice.Seal(s.voiceKey, raw); err != nil {
			s.writeError(w, err)
			return
		}
	}
	cfg, _ := json.Marshal(next.Config)
	_, err = s.db.ExecContext(r.Context(), `INSERT INTO base_voice_settings (id, enabled, provider, config, secrets_sealed, ttl_seconds, open_rooms, can_publish, updated_by, updated_at)
		VALUES (1,$1,$2,$3,$4,$5,$6,$7,$8,now())
		ON CONFLICT (id) DO UPDATE SET enabled=EXCLUDED.enabled, provider=EXCLUDED.provider, config=EXCLUDED.config, secrets_sealed=EXCLUDED.secrets_sealed,
		ttl_seconds=EXCLUDED.ttl_seconds, open_rooms=EXCLUDED.open_rooms, can_publish=EXCLUDED.can_publish, updated_by=EXCLUDED.updated_by, updated_at=now()`,
		next.Enabled, next.Provider, string(cfg), sealed, next.TTLSeconds, next.OpenRooms, next.CanPublish, consoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.voiceSettingsGet(w, r)
}

func (s *service) voiceProvider(v voiceSettings) (voice.Provider, error) {
	spec, ok := voice.Lookup(v.Provider)
	if !ok {
		return nil, &voice.ProviderError{Code: "VOICE_NOT_CONFIGURED"}
	}
	return spec.New(voice.Config{Values: v.Config, Secrets: v.Secrets}, s.voiceHTTP)
}

func voiceErrorCode(err error) (string, int) {
	var pe *voice.ProviderError
	if errors.As(err, &pe) {
		return pe.Code, pe.Status
	}
	if errors.Is(err, voice.ErrUnavailable) {
		return "VOICE_PROVIDER_UNAVAILABLE", 0
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "VOICE_PROVIDER_TIMEOUT", 0
	}
	return "VOICE_PROVIDER_ERROR", 0
}

func (s *service) voiceTest(w http.ResponseWriter, r *http.Request) {
	in, err := readVoiceInput(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	d, err := clusterkit.Allow(r.Context(), s.db, "voice-test:"+consoleUser(r), voiceTestLimit, voiceTestWindow)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if !d.Allowed {
		s.writeError(w, fail(http.StatusTooManyRequests, "TOO_MANY_TESTS", "", "too many connection tests, wait a minute"))
		return
	}
	old, err := s.loadVoiceSettings(r.Context(), s.db)
	if err != nil && !errors.Is(err, errVoiceSeal) {
		s.writeError(w, err)
		return
	}
	target := old
	if strings.TrimSpace(in.Provider) != "" {
		in.Enabled = true
		if target, err = s.voiceMerge(old, in); err != nil {
			s.writeError(w, err)
			return
		}
	}
	if target.Provider == "" {
		s.writeError(w, badRequest("VOICE_PROVIDER_REQUIRED", "provider", "pick a voice server first"))
		return
	}
	start := time.Now()
	p, err := s.voiceProvider(target)
	if err == nil {
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		err = p.Test(ctx)
		cancel()
	}
	out := map[string]any{"provider": target.Provider, "ms": time.Since(start).Milliseconds(), "ok": err == nil}
	if err != nil {
		code, status := voiceErrorCode(err)
		out["code"] = code
		if status != 0 {
			out["status"] = status
		}
		s.logger.Warn("Base voice test with %s failed: %v", target.Provider, err)
	}
	writeJSON(w, http.StatusOK, out)
}

type voiceTarget struct {
	room    string
	gameKey string
	players []voice.Player
}

func (s *service) matchTarget(ctx context.Context, userID, matchID string) (voiceTarget, error) {
	t := voiceTarget{room: matchID}
	var gameKey string
	known := false
	err := s.db.QueryRowContext(ctx, `SELECT game_key FROM economy_game_match WHERE id=$1 AND status IN ('open','started')`, matchID).Scan(&gameKey)
	if err == nil {
		known = true
		rows, err := s.db.QueryContext(ctx, `SELECT e.user_id::text, COALESCE(NULLIF(u.display_name,''), u.username, '') FROM economy_game_entry e LEFT JOIN users u ON u.id=e.user_id
			WHERE e.match_id=$1 AND e.status IN ('pending','paid','free') ORDER BY e.created_at LIMIT $2`, matchID, voicePlayersMax)
		if err != nil {
			return t, err
		}
		defer rows.Close()
		member := false
		for rows.Next() {
			var p voice.Player
			if err := rows.Scan(&p.ID, &p.Name); err != nil {
				return t, err
			}
			p.Name = clip(p.Name, 100)
			member = member || p.ID == userID
			t.players = append(t.players, p)
		}
		if err := rows.Err(); err != nil {
			return t, err
		}
		if member {
			t.gameKey = gameKey
			return t, nil
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return t, err
	}
	m, err := s.nk.MatchGet(ctx, matchID)
	if err != nil || m == nil {
		if known {
			return t, fail(http.StatusForbidden, "NOT_IN_MATCH", "", "you are not in this room")
		}
		return t, notFound("MATCH_NOT_FOUND")
	}
	parts := strings.SplitN(matchID, ".", 2)
	mode := uint8(streamModeMatchRelayed)
	if m.Authoritative {
		mode = streamModeMatchAuthoritative
		t.gameKey = m.HandlerName
	}
	if gameKey != "" {
		t.gameKey = gameKey
	}
	presences, err := s.nk.StreamUserList(mode, parts[0], "", parts[1], true, true)
	if err != nil {
		return t, err
	}
	return t, presencePlayers(&t, presences, userID, "NOT_IN_MATCH")
}

func presencePlayers(t *voiceTarget, presences []runtime.Presence, userID, code string) error {
	seen := map[string]bool{}
	t.players = nil
	member := false
	for _, p := range presences {
		if seen[p.GetUserId()] || len(t.players) >= voicePlayersMax {
			continue
		}
		seen[p.GetUserId()] = true
		member = member || p.GetUserId() == userID
		t.players = append(t.players, voice.Player{ID: p.GetUserId(), Name: clip(p.GetUsername(), 100)})
	}
	if !member {
		return fail(http.StatusForbidden, code, "", "you are not in this room")
	}
	return nil
}

func clip(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

func (s *service) rpcVoiceToken(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		MatchID string `json:"matchId"`
		PartyID string `json:"partyId"`
		Room    string `json:"room"`
		GameKey string `json:"gameKey"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	in.MatchID, in.PartyID, in.Room, in.GameKey = strings.TrimSpace(in.MatchID), strings.TrimSpace(in.PartyID), strings.TrimSpace(in.Room), strings.TrimSpace(in.GameKey)
	n := 0
	for _, v := range []string{in.MatchID, in.PartyID, in.Room} {
		if v != "" {
			n++
		}
	}
	if n != 1 {
		return nil, badRequest("VOICE_ROOM_REQUIRED", "matchId", "send exactly one of matchId, partyId or room")
	}
	if in.GameKey != "" && !gameKeyPattern.MatchString(in.GameKey) {
		return nil, badRequest("INVALID_GAME_KEY", "gameKey", "not a game key")
	}
	settings, err := s.loadVoiceSettings(ctx, s.db)
	if err != nil {
		if errors.Is(err, errVoiceSeal) {
			s.logger.Warn("Base voice: %v", err)
			return map[string]any{"enabled": false, "reason": "VOICE_NOT_CONFIGURED"}, nil
		}
		return nil, err
	}
	if !settings.Enabled || settings.Provider == "" {
		return map[string]any{"enabled": false, "reason": "VOICE_DISABLED"}, nil
	}
	d, err := clusterkit.Allow(ctx, s.db, "voice-token:"+userID, voiceTokenLimit, voiceTokenWindow)
	if err != nil {
		return nil, err
	}
	if !d.Allowed {
		return nil, fail(http.StatusTooManyRequests, "VOICE_RATE_LIMITED", "", "too many voice requests, wait a minute")
	}
	var target voiceTarget
	switch {
	case in.MatchID != "":
		if !matchIDPattern.MatchString(in.MatchID) {
			return nil, badRequest("MATCH_NOT_FOUND", "matchId", "not a match id")
		}
		if target, err = s.matchTarget(ctx, userID, in.MatchID); err != nil {
			return nil, err
		}
	case in.PartyID != "":
		if !matchIDPattern.MatchString(in.PartyID) {
			return nil, badRequest("PARTY_NOT_FOUND", "partyId", "not a party id")
		}
		parts := strings.SplitN(in.PartyID, ".", 2)
		presences, err := s.nk.StreamUserList(streamModeParty, parts[0], "", parts[1], true, true)
		if err != nil {
			return nil, err
		}
		target = voiceTarget{room: "party." + in.PartyID, gameKey: in.GameKey}
		if err := presencePlayers(&target, presences, userID, "NOT_IN_PARTY"); err != nil {
			return nil, err
		}
	default:
		if !settings.OpenRooms {
			return nil, fail(http.StatusForbidden, "ROOM_NOT_ALLOWED", "room", "open voice rooms are off")
		}
		if !roomKeyPattern.MatchString(in.Room) {
			return nil, badRequest("INVALID_ROOM", "room", "1 to 48 lower-case letters, digits, - or _")
		}
		if in.GameKey == "" {
			return nil, badRequest("GAME_KEY_REQUIRED", "gameKey", "an open room belongs to a game")
		}
		target = voiceTarget{room: "room." + in.GameKey + "." + in.Room, gameKey: in.GameKey}
	}
	if target.gameKey != "" {
		var voiceOn, active bool
		err := s.db.QueryRowContext(ctx, `SELECT voice_enabled, is_active FROM economy_game WHERE key=$1`, target.gameKey).Scan(&voiceOn, &active)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if in.Room != "" {
				return nil, notFound("GAME_NOT_FOUND")
			}
		case err != nil:
			return nil, err
		case !voiceOn:
			return map[string]any{"enabled": false, "reason": "GAME_VOICE_DISABLED", "gameKey": target.gameKey}, nil
		case !active && in.Room != "":
			return map[string]any{"enabled": false, "reason": "GAME_DISABLED", "gameKey": target.gameKey}, nil
		}
	}
	if !voice.ValidRoom(target.room) {
		return nil, badRequest("INVALID_ROOM", "room", "room name is not valid")
	}
	canPublish := settings.CanPublish
	if mute, err := restrict.Active(ctx, s.db, userID, restrict.KindMute); err == nil && mute != nil {
		canPublish = false
	}
	var name string
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(display_name,''), username) FROM users WHERE id=$1`, userID).Scan(&name)
	p, err := s.voiceProvider(settings)
	if err != nil {
		code, _ := voiceErrorCode(err)
		return nil, runtime.NewError(code, int(codes.Unavailable))
	}
	grant, err := p.Token(ctx, voice.Request{Room: target.room, Identity: userID, Name: clip(name, 100), GameKey: target.gameKey, Players: target.players,
		TTL: time.Duration(settings.TTLSeconds) * time.Second, CanPublish: canPublish})
	if err != nil {
		code, status := voiceErrorCode(err)
		s.logger.Warn("Base voice token from %s failed (%s %d): %v", settings.Provider, code, status, err)
		return nil, runtime.NewError(code, int(codes.Unavailable))
	}
	return map[string]any{"enabled": true, "provider": grant.Provider, "url": grant.URL, "token": grant.Token, "room": grant.Room, "identity": grant.Identity,
		"expiresAt": grant.ExpiresAt.UTC().Format(time.RFC3339), "canPublish": canPublish, "gameKey": target.gameKey}, nil
}
