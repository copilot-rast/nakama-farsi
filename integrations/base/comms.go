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
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/commpacks"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/tagdiscount"
)

const (
	commLease    = "base.comms.sweep"
	commLeaseTTL = 5 * time.Minute
	commTick     = 10 * time.Minute
)

type commSettingsView struct {
	commpacks.Settings
	UpdatedBy string    `json:"updatedBy,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

func (s *service) commSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := commpacks.Load(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	out := commSettingsView{Settings: st}
	_ = s.db.QueryRowContext(r.Context(), `SELECT updated_by, updated_at FROM base_comm_settings WHERE id=1`).Scan(&out.UpdatedBy, &out.UpdatedAt)
	writeJSON(w, http.StatusOK, out)
}

func (s *service) commSettingsPut(w http.ResponseWriter, r *http.Request) {
	in := commpacks.Defaults()
	var body map[string]json.RawMessage
	if err := readBody(r, &body); err != nil {
		s.writeError(w, err)
		return
	}
	delete(body, "updatedBy")
	delete(body, "updatedAt")
	raw, _ := json.Marshal(body)
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		s.writeError(w, badRequest("INVALID_BODY", "", "unknown or wrong field"))
		return
	}
	if field, msg := in.Check(); field != "" {
		s.writeError(w, badRequest("INVALID_SETTINGS", field, msg))
		return
	}
	saved, _ := json.Marshal(in)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO base_comm_settings (id, settings, updated_by, updated_at) VALUES (1, $1::jsonb, $2, now())
		ON CONFLICT (id) DO UPDATE SET settings=EXCLUDED.settings, updated_by=EXCLUDED.updated_by, updated_at=now()`, string(saved), consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	s.commSettingsGet(w, r)
}

const commChannelJSON = `jsonb_build_object('matchId',c.match_id,'status',CASE WHEN c.status='open' AND c.expires_at <= now() THEN 'expired' ELSE c.status END,
	'closeReason',c.close_reason,'closedBy',c.closed_by,'closedAt',c.closed_at,'messages',c.messages,'lastAt',c.last_at,'createdAt',c.created_at,'expiresAt',c.expires_at,
	'gameKey',(SELECT m.game_key FROM economy_game_match m WHERE m.id=c.match_id),
	'members',(SELECT COALESCE(jsonb_agg(jsonb_build_object('userId',mb.user_id,'username',COALESCE(u.username,''),'displayName',COALESCE(u.display_name,''),'canSend',mb.can_send,'sent',mb.sent,'changedBy',mb.changed_by) ORDER BY mb.created_at),'[]'::jsonb)
		FROM base_comm_member mb LEFT JOIN users u ON u.id=mb.user_id WHERE mb.match_id=c.match_id))`

func (s *service) commChannelList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	switch r.URL.Query().Get("status") {
	case "", "open":
		where.add("c.status = ?", "open")
		where.parts = append(where.parts, "c.expires_at > now()")
	case "closed":
		where.add("(c.status = ? OR c.expires_at <= now())", "closed")
	case "all":
	default:
		s.writeError(w, badRequest("INVALID_STATUS", "status", "open, closed or all"))
		return
	}
	if p.Q != "" {
		where.add("(c.match_id ILIKE ? OR EXISTS (SELECT 1 FROM base_comm_member mb JOIN users u ON u.id=mb.user_id WHERE mb.match_id=c.match_id AND u.username ILIKE ?))", likePattern(p.Q), likePattern(p.Q))
	}
	out, err := s.listRaw(r, ` FROM base_comm_channel c`, commChannelJSON, "c.created_at DESC, c.match_id", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) commChannel(ctx context.Context, q queryer, id string) (json.RawMessage, error) {
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+commChannelJSON+`::text FROM base_comm_channel c WHERE c.match_id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("CHANNEL_NOT_FOUND")
	}
	return json.RawMessage(raw), err
}

func (s *service) commChannelGet(w http.ResponseWriter, r *http.Request) {
	raw, err := s.commChannel(r.Context(), s.db, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

func (s *service) commChannelStatus(open bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := pathVar(r, "id")
		var res sql.Result
		var err error
		if open {
			st, lerr := commpacks.Load(ctx, s.db)
			if lerr != nil {
				s.writeError(w, lerr)
				return
			}
			res, err = s.db.ExecContext(ctx, `UPDATE base_comm_channel SET status='open', close_reason='', closed_by='', closed_at=NULL,
				expires_at=GREATEST(expires_at, now() + $2::INT * INTERVAL '1 hour') WHERE match_id=$1`, id, st.ChannelHours)
		} else {
			res, err = s.db.ExecContext(ctx, `UPDATE base_comm_channel SET status='closed', close_reason='staff', closed_by=$2, closed_at=now() WHERE match_id=$1`, id, consoleUser(r))
		}
		if err != nil {
			s.writeError(w, err)
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			s.writeError(w, notFound("CHANNEL_NOT_FOUND"))
			return
		}
		s.commChannelGet(w, r)
	}
}

func (s *service) commChannelCanSend(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID  string `json:"userId"`
		CanSend *bool  `json:"canSend"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if _, err := uuid.FromString(in.UserID); err != nil {
		s.writeError(w, badRequest("INVALID_USER_ID", "userId", "not a user ID"))
		return
	}
	if in.CanSend == nil {
		s.writeError(w, badRequest("CAN_SEND_REQUIRED", "canSend", "true or false"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `UPDATE base_comm_member SET can_send=$3, changed_by=$4 WHERE match_id=$1 AND user_id=$2::uuid`, pathVar(r, "id"), in.UserID, *in.CanSend, consoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("NOT_A_MEMBER"))
		return
	}
	s.commChannelGet(w, r)
}

type clientCommItem struct {
	Key      string `json:"key"`
	Ref      string `json:"ref"`
	Name     string `json:"name"`
	Text     string `json:"text,omitempty"`
	AssetID  string `json:"assetId,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
	Favorite bool   `json:"favorite"`
}

type clientCommPack struct {
	Key         string           `json:"key"`
	Kind        string           `json:"kind"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	IconURL     string           `json:"iconUrl"`
	Access      string           `json:"access"`
	Starter     bool             `json:"starter"`
	Owned       bool             `json:"owned"`
	ItemKey     string           `json:"itemKey,omitempty"`
	Item        map[string]any   `json:"item,omitempty"`
	Items       []clientCommItem `json:"items"`
}

func (s *service) commPacks(ctx context.Context, userID, lang, kind string, mine bool) ([]clientCommPack, string, error) {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, lang)
	if err != nil {
		return nil, "", err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.id::text, p.key, p.kind, p.name::text, p.description::text, p.icon_url, p.access, p.item_key,
		COALESCE((SELECT jsonb_build_object('key',i.key,'name',i.name,'iconUrl',i.icon_url,'price',i.price,'typeId',i.type_id::text,
			'productIds',(SELECT COALESCE(jsonb_agg(pr.id::text ORDER BY pr.sort, pr.created_at),'[]'::jsonb) FROM economy_product pr JOIN economy_prize z ON z.id=pr.prize_id
				WHERE pr.is_active AND z.items @> jsonb_build_array(jsonb_build_object('type','item','key',i.key))))
			FROM economy_item i WHERE i.key=p.item_key AND p.access='item'), 'null'::jsonb)::text,
		p.access='free' OR EXISTS (SELECT 1 FROM economy_user_item ui JOIN economy_item i ON i.id=ui.item_id WHERE ui.user_id=$1::uuid AND i.key=p.item_key)
		FROM base_sticker_pack p WHERE p.is_active AND p.in_match AND ($2 = '' OR p.kind = $2) ORDER BY p.sort, p.key`, userID, kind)
	if err != nil {
		return nil, lang, err
	}
	packs := []clientCommPack{}
	ids := []string{}
	disc, derr := tagdiscount.ForUser(ctx, s.db, userID)
	if derr != nil {
		disc = &tagdiscount.Set{Settings: tagdiscount.Defaults()}
	}
	for rows.Next() {
		var p clientCommPack
		var id, name, desc, item string
		if err := rows.Scan(&id, &p.Key, &p.Kind, &name, &desc, &p.IconURL, &p.Access, &p.ItemKey, &item, &p.Owned); err != nil {
			rows.Close()
			return nil, lang, err
		}
		if mine && !p.Owned {
			continue
		}
		p.Name = langs.Resolve(localized.Parse(name), lang)
		p.Description = langs.Resolve(localized.Parse(desc), lang)
		p.Starter = p.Access == "free"
		if p.Access == "item" {
			var m map[string]any
			if json.Unmarshal([]byte(item), &m) == nil && m != nil {
				if raw, ok := m["name"]; ok {
					b, _ := json.Marshal(raw)
					m["name"] = langs.Resolve(localized.Parse(string(b)), lang)
				}
				stickerItemPrice(m, disc)
				p.Item = m
			}
		} else {
			p.ItemKey = ""
		}
		p.Items = []clientCommItem{}
		packs = append(packs, p)
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, lang, err
	}
	if len(ids) == 0 {
		return packs, lang, nil
	}
	at := map[string]int{}
	for i, id := range ids {
		at[id] = i
	}
	srows, err := s.db.QueryContext(ctx, `SELECT s.pack_id::text, s.key, s.name::text, s.text::text, s.asset_id, s.image_url,
		EXISTS (SELECT 1 FROM base_comm_favorite f WHERE f.user_id=$2::uuid AND f.pack_id=s.pack_id AND f.sticker_key=s.key)
		FROM base_sticker s WHERE s.pack_id::text = ANY($1::text[]) AND s.is_active ORDER BY s.sort, s.key`, ids, userID)
	if err != nil {
		return nil, lang, err
	}
	defer srows.Close()
	for srows.Next() {
		var pid, name, text string
		var c clientCommItem
		if err := srows.Scan(&pid, &c.Key, &name, &text, &c.AssetID, &c.ImageURL, &c.Favorite); err != nil {
			return nil, lang, err
		}
		p := &packs[at[pid]]
		c.Ref = p.Key + "/" + c.Key
		c.Name = langs.Resolve(localized.Parse(name), lang)
		c.Text = langs.Resolve(localized.Parse(text), lang)
		if !p.Owned {
			c.Favorite = false
		}
		p.Items = append(p.Items, c)
	}
	return packs, lang, srows.Err()
}

func (s *service) rpcCommPacks(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
		Kind string `json:"kind"`
		Mine bool   `json:"mine"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if in.Kind != "" && !validKeyType(stickerKinds, in.Kind) {
		return nil, badRequest("INVALID_KIND", "kind", "sticker, text, gif or emoji")
	}
	st, err := commpacks.Load(ctx, s.db)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"enabled": st.Enabled, "cooldown": st.Rule(), "maxFavorites": st.MaxFavorites, "packs": []clientCommPack{}}
	if !st.Enabled {
		return out, nil
	}
	packs, lang, err := s.commPacks(ctx, userID, in.Lang, in.Kind, in.Mine)
	if err != nil {
		return nil, err
	}
	out["lang"] = lang
	out["packs"] = packs
	return out, nil
}

func (s *service) rpcCommFavorites(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
		Kind string `json:"kind"`
		Pack string `json:"pack"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.key, p.kind, s.key, s.name::text, s.text::text, s.asset_id, s.image_url, f.created_at
		FROM base_comm_favorite f JOIN base_sticker_pack p ON p.id=f.pack_id JOIN base_sticker s ON s.pack_id=f.pack_id AND s.key=f.sticker_key
		WHERE f.user_id=$1::uuid AND p.is_active AND p.in_match AND s.is_active AND ($2 = '' OR p.kind=$2) AND ($3 = '' OR p.key=$3)
			AND (p.access='free' OR EXISTS (SELECT 1 FROM economy_user_item ui JOIN economy_item i ON i.id=ui.item_id WHERE ui.user_id=$1::uuid AND i.key=p.item_key))
		ORDER BY f.created_at DESC LIMIT 1000`, userID, in.Kind, in.Pack)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var pack, kind, key, name, text, asset, image string
		var at time.Time
		if err := rows.Scan(&pack, &kind, &key, &name, &text, &asset, &image, &at); err != nil {
			return nil, err
		}
		item := map[string]any{"pack": pack, "packKind": kind, "key": key, "ref": pack + "/" + key, "name": langs.Resolve(localized.Parse(name), lang), "addedAt": at.UTC().Format(time.RFC3339)}
		if t := langs.Resolve(localized.Parse(text), lang); t != "" {
			item["text"] = t
		}
		if asset != "" {
			item["assetId"] = asset
		}
		if image != "" {
			item["imageUrl"] = image
		}
		items = append(items, item)
	}
	return map[string]any{"lang": lang, "items": items}, rows.Err()
}

func (s *service) rpcCommFavorite(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Pack     string `json:"pack"`
		Item     string `json:"item"`
		Favorite *bool  `json:"favorite"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	in.Pack, in.Item = strings.TrimSpace(in.Pack), strings.TrimSpace(in.Item)
	if in.Pack == "" || in.Item == "" {
		return nil, badRequest("ITEM_REQUIRED", "item", "send pack and item")
	}
	fav := in.Favorite == nil || *in.Favorite
	out := map[string]any{"pack": in.Pack, "item": in.Item, "favorite": fav}
	if !fav {
		if _, err := s.db.ExecContext(ctx, `DELETE FROM base_comm_favorite f USING base_sticker_pack p WHERE f.pack_id=p.id AND p.key=$2 AND f.user_id=$1::uuid AND f.sticker_key=$3`, userID, in.Pack, in.Item); err != nil {
			return nil, err
		}
		return out, nil
	}
	var packID, access, itemKey string
	err := s.db.QueryRowContext(ctx, `SELECT id::text, access, item_key FROM base_sticker_pack WHERE key=$1 AND is_active AND in_match`, in.Pack).Scan(&packID, &access, &itemKey)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fail(http.StatusPreconditionFailed, commpacks.CodePackNotFound, "pack", "this pack is not available")
	}
	if err != nil {
		return nil, err
	}
	var active bool
	err = s.db.QueryRowContext(ctx, `SELECT is_active FROM base_sticker WHERE pack_id=$1 AND key=$2`, packID, in.Item).Scan(&active)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !active) {
		return nil, fail(http.StatusPreconditionFailed, commpacks.CodeItemNotFound, "item", "this item is not available")
	}
	if err != nil {
		return nil, err
	}
	if access == "item" {
		owns, err := s.ownsItem(ctx, userID, itemKey)
		if err != nil {
			return nil, err
		}
		if !owns {
			return nil, fail(http.StatusPreconditionFailed, commpacks.CodePackNotOwned, "pack", "you do not own this pack")
		}
	}
	st, err := commpacks.Load(ctx, s.db)
	if err != nil {
		return nil, err
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('base_comm_favorite:' || $1))`, userID); err != nil {
			return err
		}
		if st.MaxFavorites > 0 {
			var n int
			if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM base_comm_favorite WHERE user_id=$1::uuid`, userID).Scan(&n); err != nil {
				return err
			}
			var exists bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_comm_favorite WHERE user_id=$1::uuid AND pack_id=$2::uuid AND sticker_key=$3)`, userID, packID, in.Item).Scan(&exists); err != nil {
				return err
			}
			if !exists && n >= st.MaxFavorites {
				return conflict("FAVORITES_FULL", "the favorites list is full")
			}
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO base_comm_favorite (user_id, pack_id, sticker_key) VALUES ($1::uuid,$2::uuid,$3) ON CONFLICT DO NOTHING`, userID, packID, in.Item)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *service) commLoop(ctx context.Context) {
	ticker := time.NewTicker(commTick)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			_ = clusterkit.ReleaseLease(context.Background(), s.db, commLease, holder)
			return
		case <-ticker.C:
		}
		ok, err := clusterkit.AcquireLease(ctx, s.db, commLease, holder, commLeaseTTL)
		if err != nil || !ok {
			continue
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE base_comm_channel SET status='closed', close_reason='expired', closed_by='system', closed_at=now() WHERE status='open' AND expires_at <= now()`); err != nil {
			s.logger.Warn("Base comm channel expiry failed: %v", err)
		}
		if _, err := s.db.ExecContext(ctx, `DELETE FROM base_comm_channel WHERE created_at < now() - INTERVAL '7 days'`); err != nil {
			s.logger.Warn("Base comm channel sweep failed: %v", err)
		}
		s.sweepVoiceRooms(ctx)
	}
}
