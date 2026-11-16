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
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/tagdiscount"
	"google.golang.org/grpc/codes"
)

const (
	CodeStickersDisabled      = "STICKERS_DISABLED"
	CodeStickerNotAllowedHere = "STICKER_NOT_ALLOWED_HERE"
	CodeStickerNotFound       = "STICKER_NOT_FOUND"
	CodeStickerLocked         = "STICKER_LOCKED"
	stickerMaxPerPack         = 200
)

type stickerSettings struct {
	Enabled    bool      `json:"enabled"`
	Direct     bool      `json:"direct"`
	Groups     bool      `json:"groups"`
	Rooms      bool      `json:"rooms"`
	CountQuota bool      `json:"countQuota"`
	UpdatedBy  string    `json:"updatedBy,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt,omitempty"`
}

func defaultStickerSettings() stickerSettings {
	return stickerSettings{Enabled: true, Direct: true, Groups: true, Rooms: true, CountQuota: true}
}

func loadStickerSettings(ctx context.Context, q queryer) (stickerSettings, error) {
	st := defaultStickerSettings()
	var raw, by string
	var at time.Time
	err := q.QueryRowContext(ctx, `SELECT settings::text, updated_by, updated_at FROM base_sticker_settings WHERE id=1`).Scan(&raw, &by, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return defaultStickerSettings(), nil
	}
	st.UpdatedBy, st.UpdatedAt = by, at
	return st, nil
}

func (st stickerSettings) allows(kind string) bool {
	switch kind {
	case "dm":
		return st.Direct
	case "group":
		return st.Groups
	case "room":
		return st.Rooms
	}
	return false
}

var stickerMessages = map[string]localized.Text{
	CodeStickersDisabled:      {"fa": "ارسال استیکر فعلاً بسته است.", "en": "Stickers are turned off.", "ar": "الملصقات متوقفة حالياً."},
	CodeStickerNotAllowedHere: {"fa": "در این گفتگو نمی‌توانی استیکر بفرستی.", "en": "Stickers are not allowed in this chat.", "ar": "لا يمكنك إرسال ملصق في هذه المحادثة."},
	CodeStickerNotFound:       {"fa": "این استیکر پیدا نشد.", "en": "This sticker was not found.", "ar": "لم يتم العثور على هذا الملصق."},
	CodeStickerLocked:         {"fa": "این بستهٔ استیکر را هنوز نداری.", "en": "You do not own this sticker pack yet.", "ar": "لا تملك حزمة الملصقات هذه بعد."},
}

type stickerError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Lang    string         `json:"lang"`
	Meta    map[string]any `json:"meta,omitempty"`
}

func (s *service) stickerRefusal(ctx context.Context, userID, code string, grpc codes.Code, meta map[string]any) error {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, "")
	text := ""
	if err == nil && langs != nil {
		text = langs.Resolve(stickerMessages[code], lang)
	} else {
		lang = "en"
	}
	if text == "" {
		text = stickerMessages[code]["en"]
	}
	raw, _ := json.Marshal(stickerError{Code: code, Message: text, Lang: lang, Meta: meta})
	return runtime.NewError(string(raw), int(grpc))
}

type stickerRef struct {
	Pack, Key, ImageURL, Access, ItemKey string
}

func splitStickerRef(v string) (string, string, bool) {
	pack, key, ok := strings.Cut(strings.TrimSpace(v), "/")
	if !ok || !keyPattern.MatchString(pack) || !keyPattern.MatchString(key) {
		return "", "", false
	}
	return pack, key, true
}

func (s *service) ownsItem(ctx context.Context, userID, itemKey string) (bool, error) {
	var owns bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_user_item ui JOIN economy_item i ON i.id=ui.item_id WHERE ui.user_id=$1::uuid AND i.key=$2)`, userID, itemKey).Scan(&owns)
	return owns, err
}

func (s *service) stickerGate(ctx context.Context, userID, channelID, content string) (string, bool, stickerSettings, error) {
	st := defaultStickerSettings()
	if userID == "" || !validChatContent(content) {
		return content, false, st, nil
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return content, false, st, nil
	}
	rawRef, ok := doc["sticker"]
	if !ok {
		return content, false, st, nil
	}
	st, err := loadStickerSettings(ctx, s.db)
	if err != nil {
		return "", true, st, s.hookFailure(err)
	}
	if !st.Enabled {
		return "", true, st, s.stickerRefusal(ctx, userID, CodeStickersDisabled, codes.PermissionDenied, nil)
	}
	t, ok := parseChatChannel(channelID, userID)
	if !ok {
		return content, true, st, nil
	}
	if !st.allows(t.Kind) {
		return "", true, st, s.stickerRefusal(ctx, userID, CodeStickerNotAllowedHere, codes.PermissionDenied, map[string]any{"chat": t.Kind})
	}
	var ref string
	_ = json.Unmarshal(rawRef, &ref)
	pack, key, ok := splitStickerRef(ref)
	if !ok {
		return "", true, st, s.stickerRefusal(ctx, userID, CodeStickerNotFound, codes.NotFound, map[string]any{"sticker": ref})
	}
	var r stickerRef
	err = s.db.QueryRowContext(ctx, `SELECT p.key, s.key, s.image_url, p.access, p.item_key FROM base_sticker s JOIN base_sticker_pack p ON p.id=s.pack_id
		WHERE p.key=$1 AND s.key=$2 AND p.is_active AND s.is_active`, pack, key).Scan(&r.Pack, &r.Key, &r.ImageURL, &r.Access, &r.ItemKey)
	if errors.Is(err, sql.ErrNoRows) {
		return "", true, st, s.stickerRefusal(ctx, userID, CodeStickerNotFound, codes.NotFound, map[string]any{"sticker": ref})
	}
	if err != nil {
		return "", true, st, s.hookFailure(err)
	}
	if r.Access == "item" {
		owns, err := s.ownsItem(ctx, userID, r.ItemKey)
		if err != nil {
			return "", true, st, s.hookFailure(err)
		}
		if !owns {
			return "", true, st, s.stickerRefusal(ctx, userID, CodeStickerLocked, codes.PermissionDenied, map[string]any{"pack": r.Pack, "itemKey": r.ItemKey})
		}
	}
	doc["sticker"], _ = json.Marshal(r.Pack + "/" + r.Key)
	doc["stickerPack"], _ = json.Marshal(r.Pack)
	doc["stickerUrl"], _ = json.Marshal(r.ImageURL)
	out, err := json.Marshal(doc)
	if err != nil {
		return "", true, st, s.hookFailure(err)
	}
	return string(out), true, st, nil
}

type clientSticker struct {
	Key      string `json:"key"`
	Ref      string `json:"ref"`
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
}

type clientStickerPack struct {
	Key         string          `json:"key"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	IconURL     string          `json:"iconUrl"`
	Access      string          `json:"access"`
	Owned       bool            `json:"owned"`
	ItemKey     string          `json:"itemKey,omitempty"`
	Item        map[string]any  `json:"item,omitempty"`
	Stickers    []clientSticker `json:"stickers"`
}

func (s *service) rpcStickers(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	st, err := loadStickerSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"lang": lang, "enabled": st.Enabled, "chats": map[string]bool{"direct": st.Direct, "groups": st.Groups, "rooms": st.Rooms}, "countQuota": st.CountQuota, "packs": []clientStickerPack{}}
	if !st.Enabled {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.id::text, p.key, p.name::text, p.description::text, p.icon_url, p.access, p.item_key,
		COALESCE((SELECT jsonb_build_object('key',i.key,'name',i.name,'iconUrl',i.icon_url,'price',i.price,'typeId',i.type_id::text,
			'productIds',(SELECT COALESCE(jsonb_agg(pr.id::text ORDER BY pr.sort, pr.created_at),'[]'::jsonb) FROM economy_product pr JOIN economy_prize z ON z.id=pr.prize_id
				WHERE pr.is_active AND z.items @> jsonb_build_array(jsonb_build_object('type','item','key',i.key))))
			FROM economy_item i WHERE i.key=p.item_key AND p.access='item'), 'null'::jsonb)::text,
		p.access='free' OR EXISTS (SELECT 1 FROM economy_user_item ui JOIN economy_item i ON i.id=ui.item_id WHERE ui.user_id=$1::uuid AND i.key=p.item_key)
		FROM base_sticker_pack p WHERE p.is_active ORDER BY p.sort, p.key`, userID)
	if err != nil {
		return nil, err
	}
	packs := []clientStickerPack{}
	ids := []string{}
	disc, derr := tagdiscount.ForUser(ctx, s.db, userID)
	if derr != nil {
		disc = &tagdiscount.Set{Settings: tagdiscount.Defaults()}
	}
	for rows.Next() {
		var p clientStickerPack
		var id, name, desc, item string
		if err := rows.Scan(&id, &p.Key, &name, &desc, &p.IconURL, &p.Access, &p.ItemKey, &item, &p.Owned); err != nil {
			rows.Close()
			return nil, err
		}
		p.Name = langs.Resolve(localized.Parse(name), lang)
		p.Description = langs.Resolve(localized.Parse(desc), lang)
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
		p.Stickers = []clientSticker{}
		packs = append(packs, p)
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) > 0 {
		at := map[string]int{}
		for i, id := range ids {
			at[id] = i
		}
		srows, err := s.db.QueryContext(ctx, `SELECT pack_id::text, key, name::text, image_url FROM base_sticker WHERE pack_id::text = ANY($1::text[]) AND is_active ORDER BY sort, key`, ids)
		if err != nil {
			return nil, err
		}
		for srows.Next() {
			var pid, name string
			var c clientSticker
			if err := srows.Scan(&pid, &c.Key, &name, &c.ImageURL); err != nil {
				srows.Close()
				return nil, err
			}
			p := &packs[at[pid]]
			c.Ref = p.Key + "/" + c.Key
			c.Name = langs.Resolve(localized.Parse(name), lang)
			p.Stickers = append(p.Stickers, c)
		}
		srows.Close()
		if err := srows.Err(); err != nil {
			return nil, err
		}
	}
	out["packs"] = packs
	return out, nil
}

func (s *service) stickerSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := loadStickerSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *service) stickerSettingsPut(w http.ResponseWriter, r *http.Request) {
	in := defaultStickerSettings()
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	in.UpdatedBy, in.UpdatedAt = "", time.Time{}
	raw, _ := json.Marshal(map[string]bool{"enabled": in.Enabled, "direct": in.Direct, "groups": in.Groups, "rooms": in.Rooms, "countQuota": in.CountQuota})
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO base_sticker_settings (id, settings, updated_by, updated_at) VALUES (1, $1::jsonb, $2, now())
		ON CONFLICT (id) DO UPDATE SET settings=EXCLUDED.settings, updated_by=EXCLUDED.updated_by, updated_at=now()`, string(raw), consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	s.stickerSettingsGet(w, r)
}

const stickerPackJSON = `jsonb_build_object('id',p.id,'key',p.key,'name',p.name,'description',p.description,'iconUrl',p.icon_url,'access',p.access,'itemKey',p.item_key,
	'item',(SELECT jsonb_build_object('key',i.key,'name',i.name,'iconUrl',i.icon_url) FROM economy_item i WHERE i.key=p.item_key AND p.access='item'),
	'sort',p.sort,'isActive',p.is_active,'createdBy',p.created_by,'createdAt',p.created_at,'updatedAt',p.updated_at,
	'count',(SELECT count(*) FROM base_sticker s WHERE s.pack_id=p.id),
	'stickers',(SELECT COALESCE(jsonb_agg(jsonb_build_object('key',s.key,'name',s.name,'imageUrl',s.image_url,'isActive',s.is_active) ORDER BY s.sort, s.key),'[]'::jsonb) FROM base_sticker s WHERE s.pack_id=p.id))`

func (s *service) stickerPackList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		where.add("(p.key ILIKE ? OR "+translatedSearch("p.name", "?")+")", likePattern(p.Q), likePattern(p.Q))
	}
	if v := r.URL.Query().Get("isActive"); v == "true" || v == "false" {
		where.add("p.is_active = ?", v == "true")
	}
	if v := r.URL.Query().Get("access"); v == "free" || v == "item" {
		where.add("p.access = ?", v)
	}
	out, err := s.listRaw(r, ` FROM base_sticker_pack p`, stickerPackJSON, "p.sort, p.key", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func fetchStickerPack(ctx context.Context, q queryer, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+stickerPackJSON+`::text FROM base_sticker_pack p WHERE p.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func (s *service) stickerPackGet(w http.ResponseWriter, r *http.Request) {
	raw, err := fetchStickerPack(r.Context(), s.db, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

type stickerIn struct {
	Key      string         `json:"key"`
	Name     localized.Text `json:"name"`
	ImageURL string         `json:"imageUrl"`
	IsActive *bool          `json:"isActive"`
}

func (s *service) stickerPackWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Key         string         `json:"key"`
			Name        localized.Text `json:"name"`
			Description localized.Text `json:"description"`
			IconURL     string         `json:"iconUrl"`
			Access      string         `json:"access"`
			ItemKey     string         `json:"itemKey"`
			Sort        int            `json:"sort"`
			IsActive    *bool          `json:"isActive"`
			Stickers    []stickerIn    `json:"stickers"`
		}
		if err := readBody(r, &in); err != nil {
			s.writeError(w, err)
			return
		}
		ctx := r.Context()
		id := pathVar(r, "id")
		if create {
			id = uuid.Must(uuid.NewV4()).String()
		} else if _, err := uuid.FromString(id); err != nil {
			s.writeError(w, notFound("NOT_FOUND"))
			return
		}
		langs, err := localized.Load(ctx, s.db)
		if err != nil {
			s.writeError(w, err)
			return
		}
		key, err := cleanKey("key", in.Key)
		if err == nil {
			in.Name, err = cleanText(langs, "name", in.Name, true, 100)
		}
		if err == nil {
			in.Description, err = cleanText(langs, "description", in.Description, false, 500)
		}
		if err == nil {
			in.IconURL, err = cleanURL("iconUrl", in.IconURL)
		}
		if err == nil && (in.Sort < -100000 || in.Sort > 100000) {
			err = badRequest("INVALID_SORT", "sort", "between -100000 and 100000")
		}
		in.ItemKey = strings.TrimSpace(in.ItemKey)
		if err == nil {
			switch in.Access {
			case "", "free":
				in.Access, in.ItemKey = "free", ""
			case "item":
				if in.ItemKey == "" {
					err = badRequest("ITEM_REQUIRED", "itemKey", "pick the item that opens this pack")
				} else {
					err = s.checkKeysExist(ctx, "economy_item", "itemKey", []string{in.ItemKey})
				}
			default:
				err = badRequest("INVALID_ACCESS", "access", "free or item")
			}
		}
		if err == nil && (len(in.Stickers) == 0 || len(in.Stickers) > stickerMaxPerPack) {
			err = badRequest("INVALID_STICKERS", "stickers", "1 to 200 stickers")
		}
		seen := map[string]bool{}
		for i := 0; err == nil && i < len(in.Stickers); i++ {
			st := &in.Stickers[i]
			field := "stickers[" + itoa(i) + "]"
			if st.Key, err = cleanKey(field+".key", st.Key); err != nil {
				break
			}
			if seen[st.Key] {
				err = badRequest("DUPLICATE_KEY", field+".key", "each sticker needs its own key")
				break
			}
			seen[st.Key] = true
			if st.Name, err = cleanText(langs, field+".name", st.Name, false, 100); err != nil {
				break
			}
			if st.ImageURL, err = cleanURL(field+".imageUrl", st.ImageURL); err == nil && st.ImageURL == "" {
				err = badRequest("IMAGE_REQUIRED", field+".imageUrl", "upload the sticker image")
			}
		}
		if err != nil {
			s.writeError(w, err)
			return
		}
		active := in.IsActive == nil || *in.IsActive
		name, _ := json.Marshal(in.Name)
		desc, _ := json.Marshal(in.Description)
		var out json.RawMessage
		err = s.inTx(ctx, func(tx *sql.Tx) error {
			if create {
				var taken bool
				if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_sticker_pack WHERE key=$1)`, key).Scan(&taken); err != nil {
					return err
				}
				if taken {
					return conflict("KEY_EXISTS", "another sticker pack has this key")
				}
				if _, err := tx.ExecContext(ctx, `INSERT INTO base_sticker_pack (id, key, name, description, icon_url, access, item_key, sort, is_active, created_by)
					VALUES ($1,$2,$3::jsonb,$4::jsonb,$5,$6,$7,$8,$9,$10)`, id, key, string(name), string(desc), in.IconURL, in.Access, in.ItemKey, in.Sort, active, consoleUser(r)); err != nil {
					return err
				}
			} else {
				res, err := tx.ExecContext(ctx, `UPDATE base_sticker_pack SET name=$2::jsonb, description=$3::jsonb, icon_url=$4, access=$5, item_key=$6, sort=$7, is_active=$8, updated_at=now() WHERE id=$1`,
					id, string(name), string(desc), in.IconURL, in.Access, in.ItemKey, in.Sort, active)
				if err != nil {
					return err
				}
				if n, _ := res.RowsAffected(); n == 0 {
					return notFound("NOT_FOUND")
				}
			}
			keys := make([]string, 0, len(in.Stickers))
			for i, st := range in.Stickers {
				sn, _ := json.Marshal(st.Name)
				on := st.IsActive == nil || *st.IsActive
				if _, err := tx.ExecContext(ctx, `INSERT INTO base_sticker (id, pack_id, key, name, image_url, sort, is_active) VALUES ($1,$2,$3,$4::jsonb,$5,$6,$7)
					ON CONFLICT (pack_id, key) DO UPDATE SET name=EXCLUDED.name, image_url=EXCLUDED.image_url, sort=EXCLUDED.sort, is_active=EXCLUDED.is_active`,
					uuid.Must(uuid.NewV4()).String(), id, st.Key, string(sn), st.ImageURL, i, on); err != nil {
					return err
				}
				keys = append(keys, st.Key)
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM base_sticker WHERE pack_id=$1 AND NOT (key = ANY($2::text[]))`, id, keys); err != nil {
				return err
			}
			out, err = fetchStickerPack(ctx, tx, id)
			return err
		})
		if err != nil {
			s.writeError(w, err)
			return
		}
		status := http.StatusOK
		if create {
			status = http.StatusCreated
		}
		writeRaw(w, status, out)
	}
}

func (s *service) stickerPackDelete(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM base_sticker_pack WHERE id=$1`, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func stickerItemPrice(m map[string]any, disc *tagdiscount.Set) {
	typeID, _ := m["typeId"].(string)
	delete(m, "typeId")
	key, _ := m["key"].(string)
	m["finalPrice"] = m["price"]
	m["discountPercent"] = 0
	price, ok := m["price"].(map[string]any)
	if !ok {
		return
	}
	amount, ok := price["amount"].(float64)
	if !ok {
		return
	}
	match := disc.Item(key, typeID)
	if match.Percent <= 0 {
		return
	}
	m["finalPrice"] = map[string]any{"currencyKey": price["currencyKey"], "amount": tagdiscount.Apply(int64(amount), match.Percent)}
	m["discountPercent"] = match.Percent
}
