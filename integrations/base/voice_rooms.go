package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/voice"
)

const voiceRoomsMax = 200

func voiceRoomKind(room string) string {
	switch {
	case strings.HasPrefix(room, "party."):
		return "party"
	case strings.HasPrefix(room, "room."):
		return "room"
	}
	return "match"
}

func (s *service) voiceMuted(ctx context.Context, room, userID string) (bool, error) {
	var muted bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_voice_mute mu JOIN base_voice_room r ON r.room=mu.room WHERE mu.room=$1 AND mu.user_id=$2::uuid AND r.expires_at > now())`, room, userID).Scan(&muted)
	return muted, err
}

func (s *service) recordVoiceGrant(ctx context.Context, provider, room, gameKey, userID, name string, canPublish bool, expires time.Time) {
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO base_voice_room (room, provider, kind, game_key, expires_at) VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (room) DO UPDATE SET provider=EXCLUDED.provider, game_key=EXCLUDED.game_key, last_token_at=now(), expires_at=GREATEST(base_voice_room.expires_at, EXCLUDED.expires_at)`,
			room, provider, voiceRoomKind(room), gameKey, expires); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO base_voice_member (room, user_id, name, can_publish) VALUES ($1,$2::uuid,$3,$4)
			ON CONFLICT (room, user_id) DO UPDATE SET name=EXCLUDED.name, can_publish=EXCLUDED.can_publish, last_token_at=now()`, room, userID, name, canPublish)
		return err
	})
	if err != nil {
		s.logger.Warn("Base voice room %s not recorded: %v", room, err)
	}
}

func (s *service) voiceModerator(ctx context.Context) (voice.Moderator, voiceSettings, error) {
	settings, err := s.loadVoiceSettings(ctx, s.db)
	if err != nil {
		return nil, settings, err
	}
	if settings.Provider == "" {
		return nil, settings, nil
	}
	p, err := s.voiceProvider(settings)
	if err != nil {
		return nil, settings, nil
	}
	m, _ := p.(voice.Moderator)
	return m, settings, nil
}

type voiceMemberView struct {
	UserID      string     `json:"userId"`
	Username    string     `json:"username"`
	DisplayName string     `json:"displayName"`
	Name        string     `json:"name"`
	CanPublish  bool       `json:"canPublish"`
	Muted       bool       `json:"muted"`
	MutedBy     string     `json:"mutedBy,omitempty"`
	MutedAt     *time.Time `json:"mutedAt,omitempty"`
	Connected   *bool      `json:"connected"`
	LastTokenAt time.Time  `json:"lastTokenAt"`
}

type voiceRoomView struct {
	Room        string            `json:"room"`
	Kind        string            `json:"kind"`
	Provider    string            `json:"provider"`
	GameKey     string            `json:"gameKey"`
	CreatedAt   time.Time         `json:"createdAt"`
	LastTokenAt time.Time         `json:"lastTokenAt"`
	ExpiresAt   time.Time         `json:"expiresAt"`
	Live        *bool             `json:"live"`
	Members     []voiceMemberView `json:"members"`
}

func (s *service) voiceRoomsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	rows, err := s.db.QueryContext(ctx, `SELECT room, kind, provider, game_key, created_at, last_token_at, expires_at FROM base_voice_room
		WHERE expires_at > now() AND ($1 = '' OR room ILIKE '%' || $1 || '%' OR game_key ILIKE '%' || $1 || '%'
			OR EXISTS (SELECT 1 FROM base_voice_member m LEFT JOIN users u ON u.id=m.user_id WHERE m.room=base_voice_room.room AND (u.username ILIKE '%' || $1 || '%' OR m.user_id::text = $1)))
		ORDER BY last_token_at DESC LIMIT $2`, q, voiceRoomsMax)
	if err != nil {
		s.writeError(w, err)
		return
	}
	rooms := []voiceRoomView{}
	at := map[string]int{}
	for rows.Next() {
		var v voiceRoomView
		if err := rows.Scan(&v.Room, &v.Kind, &v.Provider, &v.GameKey, &v.CreatedAt, &v.LastTokenAt, &v.ExpiresAt); err != nil {
			rows.Close()
			s.writeError(w, err)
			return
		}
		v.Members = []voiceMemberView{}
		at[v.Room] = len(rooms)
		rooms = append(rooms, v)
	}
	rows.Close()
	if len(rooms) > 0 {
		names := make([]string, 0, len(rooms))
		for _, v := range rooms {
			names = append(names, v.Room)
		}
		mrows, err := s.db.QueryContext(ctx, `SELECT m.room, m.user_id::text, COALESCE(u.username,''), COALESCE(u.display_name,''), m.name, m.can_publish, mu.user_id IS NOT NULL, COALESCE(mu.muted_by,''), mu.created_at, m.last_token_at
			FROM base_voice_member m LEFT JOIN users u ON u.id=m.user_id LEFT JOIN base_voice_mute mu ON mu.room=m.room AND mu.user_id=m.user_id
			WHERE m.room = ANY($1) ORDER BY m.first_token_at`, names)
		if err != nil {
			s.writeError(w, err)
			return
		}
		for mrows.Next() {
			var room string
			var m voiceMemberView
			var mutedAt sql.NullTime
			if err := mrows.Scan(&room, &m.UserID, &m.Username, &m.DisplayName, &m.Name, &m.CanPublish, &m.Muted, &m.MutedBy, &mutedAt, &m.LastTokenAt); err != nil {
				mrows.Close()
				s.writeError(w, err)
				return
			}
			if mutedAt.Valid {
				t := mutedAt.Time
				m.MutedAt = &t
			}
			rooms[at[room]].Members = append(rooms[at[room]].Members, m)
		}
		mrows.Close()
	}
	if r.URL.Query().Get("live") == "true" {
		if mod, _, err := s.voiceModerator(ctx); err == nil && mod != nil {
			for i := range rooms {
				if i >= 20 {
					break
				}
				cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
				list, err := mod.Participants(cctx, rooms[i].Room)
				cancel()
				if err != nil {
					continue
				}
				on := map[string]bool{}
				for _, p := range list {
					on[p.Identity] = p.Connected
				}
				live := len(list) > 0
				rooms[i].Live = &live
				for j := range rooms[i].Members {
					c := on[rooms[i].Members[j].UserID]
					rooms[i].Members[j].Connected = &c
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": rooms, "total": len(rooms)})
}

func (s *service) voiceMute(muted bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		room := pathVar(r, "room")
		var in struct {
			UserID string `json:"userId"`
		}
		if err := readBody(r, &in); err != nil {
			s.writeError(w, err)
			return
		}
		if _, err := uuid.FromString(in.UserID); err != nil {
			s.writeError(w, badRequest("INVALID_USER_ID", "userId", "not a user ID"))
			return
		}
		var exists bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_voice_room WHERE room=$1 AND expires_at > now())`, room).Scan(&exists); err != nil {
			s.writeError(w, err)
			return
		}
		if !exists {
			s.writeError(w, notFound("VOICE_ROOM_NOT_FOUND"))
			return
		}
		var member bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_voice_member WHERE room=$1 AND user_id=$2::uuid)`, room, in.UserID).Scan(&member); err != nil {
			s.writeError(w, err)
			return
		}
		if !member {
			s.writeError(w, notFound("VOICE_PARTICIPANT_NOT_FOUND"))
			return
		}
		var err error
		if muted {
			_, err = s.db.ExecContext(ctx, `INSERT INTO base_voice_mute (room, user_id, muted_by) VALUES ($1,$2::uuid,$3) ON CONFLICT (room, user_id) DO UPDATE SET muted_by=EXCLUDED.muted_by, created_at=now()`, room, in.UserID, consoleUser(r))
		} else {
			_, err = s.db.ExecContext(ctx, `DELETE FROM base_voice_mute WHERE room=$1 AND user_id=$2::uuid`, room, in.UserID)
		}
		if err != nil {
			s.writeError(w, err)
			return
		}
		out := map[string]any{"room": room, "userId": in.UserID, "muted": muted, "applied": false}
		mod, settings, err := s.voiceModerator(ctx)
		can := !muted && settings.CanPublish
		if err == nil && mod != nil {
			cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			perr := mod.SetPublish(cctx, room, in.UserID, can)
			cancel()
			switch {
			case perr == nil:
				out["applied"] = true
			case errors.Is(perr, voice.ErrNotInRoom):
				out["reason"] = "NOT_CONNECTED"
			default:
				code, _ := voiceErrorCode(perr)
				out["reason"] = code
				s.logger.Warn("Base voice mute of %s in %s not applied live: %v", in.UserID, room, perr)
			}
		} else {
			out["reason"] = "VOICE_NOT_CONFIGURED"
		}
		_, _ = s.db.ExecContext(ctx, `UPDATE base_voice_member SET can_publish=$3 WHERE room=$1 AND user_id=$2::uuid`, room, in.UserID, can)
		raw, _ := json.Marshal(out)
		writeRaw(w, http.StatusOK, raw)
	}
}

func (s *service) sweepVoiceRooms(ctx context.Context) {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM base_voice_mute WHERE room IN (SELECT room FROM base_voice_room WHERE expires_at < now() - INTERVAL '1 day')`); err != nil {
		s.logger.Warn("Base voice mute sweep failed: %v", err)
		return
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM base_voice_room WHERE expires_at < now() - INTERVAL '1 day'`); err != nil {
		s.logger.Warn("Base voice room sweep failed: %v", err)
	}
}
