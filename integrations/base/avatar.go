package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"google.golang.org/grpc/codes"
)

const (
	avatarMaxOptions = 300
	avatarMaxIDs     = 100
)

type avatarSettings struct {
	Enabled   bool      `json:"enabled"`
	UpdatedBy string    `json:"updatedBy,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

func loadAvatarSettings(ctx context.Context, q queryer) (avatarSettings, error) {
	st := avatarSettings{Enabled: true}
	var raw, by string
	var at time.Time
	err := q.QueryRowContext(ctx, `SELECT settings::text, updated_by, updated_at FROM base_avatar_settings WHERE id=1`).Scan(&raw, &by, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return avatarSettings{Enabled: true}, nil
	}
	st.UpdatedBy, st.UpdatedAt = by, at
	return st, nil
}

type avatarOption struct {
	Key       string
	Name      localized.Text
	ImageURL  string
	IconURL   string
	Value     string
	Access    string
	ItemKey   string
	IsDefault bool
}

type avatarPart struct {
	ID       string
	Key      string
	Name     localized.Text
	Layer    int
	Required bool
	Options  []avatarOption
}

func (p avatarPart) option(key string) (avatarOption, bool) {
	for _, o := range p.Options {
		if o.Key == key {
			return o, true
		}
	}
	return avatarOption{}, false
}

func (p avatarPart) fallback() (avatarOption, bool) {
	for _, o := range p.Options {
		if o.IsDefault && o.Access == "free" {
			return o, true
		}
	}
	if p.Required {
		for _, o := range p.Options {
			if o.Access == "free" {
				return o, true
			}
		}
	}
	return avatarOption{}, false
}

func loadAvatarCatalog(ctx context.Context, q *sql.DB) ([]avatarPart, error) {
	rows, err := q.QueryContext(ctx, `SELECT p.id::text, p.key, p.name::text, p.layer, p.required FROM base_avatar_part p WHERE p.is_active ORDER BY p.sort, p.key`)
	if err != nil {
		return nil, err
	}
	parts := []avatarPart{}
	at := map[string]int{}
	for rows.Next() {
		var p avatarPart
		var name string
		if err := rows.Scan(&p.ID, &p.Key, &name, &p.Layer, &p.Required); err != nil {
			rows.Close()
			return nil, err
		}
		p.Name = localized.Parse(name)
		at[p.ID] = len(parts)
		parts = append(parts, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return parts, nil
	}
	orows, err := q.QueryContext(ctx, `SELECT o.part_id::text, o.key, o.name::text, o.image_url, o.icon_url, o.value, o.access, o.item_key, o.is_default
		FROM base_avatar_option o JOIN base_avatar_part p ON p.id=o.part_id WHERE p.is_active AND o.is_active ORDER BY o.sort, o.key`)
	if err != nil {
		return nil, err
	}
	defer orows.Close()
	for orows.Next() {
		var pid, name string
		var o avatarOption
		if err := orows.Scan(&pid, &o.Key, &name, &o.ImageURL, &o.IconURL, &o.Value, &o.Access, &o.ItemKey, &o.IsDefault); err != nil {
			return nil, err
		}
		o.Name = localized.Parse(name)
		if i, ok := at[pid]; ok {
			parts[i].Options = append(parts[i].Options, o)
		}
	}
	return parts, orows.Err()
}

func (s *service) ownedItemKeys(ctx context.Context, userID string, keys []string) (map[string]bool, error) {
	out := map[string]bool{}
	if len(keys) == 0 {
		return out, nil
	}
	if _, err := uuid.FromString(userID); err != nil {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT i.key FROM economy_user_item ui JOIN economy_item i ON i.id=ui.item_id WHERE ui.user_id=$1::uuid AND i.key = ANY($2::text[])`, userID, keys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		out[k] = true
	}
	return out, rows.Err()
}

func avatarItemKeys(parts []avatarPart) []string {
	seen := map[string]bool{}
	for _, p := range parts {
		for _, o := range p.Options {
			if o.Access == "item" {
				seen[o.ItemKey] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

type avatarLayer struct {
	Part     string `json:"part"`
	Option   string `json:"option"`
	ImageURL string `json:"imageUrl"`
	Value    string `json:"value"`
	Layer    int    `json:"layer"`
}

type avatarLook struct {
	Parts  map[string]string `json:"parts"`
	Layers []avatarLayer     `json:"layers"`
}

func resolveLook(parts []avatarPart, saved map[string]string, owned map[string]bool) avatarLook {
	out := avatarLook{Parts: map[string]string{}, Layers: []avatarLayer{}}
	for _, p := range parts {
		o, ok := p.option(saved[p.Key])
		if ok && o.Access == "item" && !owned[o.ItemKey] {
			ok = false
		}
		if !ok {
			if _, chosen := saved[p.Key]; chosen && saved[p.Key] == "" && !p.Required {
				continue
			}
			o, ok = p.fallback()
		}
		if !ok {
			continue
		}
		out.Parts[p.Key] = o.Key
		out.Layers = append(out.Layers, avatarLayer{Part: p.Key, Option: o.Key, ImageURL: o.ImageURL, Value: o.Value, Layer: p.Layer})
	}
	sort.SliceStable(out.Layers, func(i, j int) bool { return out.Layers[i].Layer < out.Layers[j].Layer })
	return out
}

func (s *service) savedLooks(ctx context.Context, ids []string) (map[string]map[string]string, error) {
	out := map[string]map[string]string{}
	valid := []string{}
	for _, id := range ids {
		if _, err := uuid.FromString(id); err == nil {
			valid = append(valid, id)
		}
	}
	if len(valid) == 0 {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text, look::text FROM base_avatar_look WHERE user_id::text = ANY($1::text[])`, valid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, raw string
		if err := rows.Scan(&id, &raw); err != nil {
			return nil, err
		}
		m := map[string]string{}
		_ = json.Unmarshal([]byte(raw), &m)
		out[id] = m
	}
	return out, rows.Err()
}

func avatarFail(code string, grpc codes.Code, meta map[string]any) error {
	raw, _ := json.Marshal(map[string]any{"code": code, "meta": meta})
	return runtime.NewError(string(raw), int(grpc))
}

func (s *service) rpcAvatar(ctx context.Context, userID, payload string) (any, error) {
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
	st, err := loadAvatarSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"lang": lang, "enabled": st.Enabled, "parts": []any{}, "look": avatarLook{Parts: map[string]string{}, Layers: []avatarLayer{}}}
	if !st.Enabled {
		return out, nil
	}
	parts, err := loadAvatarCatalog(ctx, s.db)
	if err != nil {
		return nil, err
	}
	keys := avatarItemKeys(parts)
	owned, err := s.ownedItemKeys(ctx, userID, keys)
	if err != nil {
		return nil, err
	}
	items := map[string]map[string]any{}
	if len(keys) > 0 {
		rows, err := s.db.QueryContext(ctx, `SELECT key, name::text, icon_url, COALESCE(price::text,'null') FROM economy_item WHERE key = ANY($1::text[])`, keys)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var k, name, icon, price string
			if err := rows.Scan(&k, &name, &icon, &price); err != nil {
				rows.Close()
				return nil, err
			}
			items[k] = map[string]any{"key": k, "name": langs.Resolve(localized.Parse(name), lang), "iconUrl": icon, "price": json.RawMessage(price)}
		}
		rows.Close()
	}
	list := make([]map[string]any, 0, len(parts))
	for _, p := range parts {
		opts := make([]map[string]any, 0, len(p.Options))
		for _, o := range p.Options {
			v := map[string]any{"key": o.Key, "name": langs.Resolve(o.Name, lang), "imageUrl": o.ImageURL, "iconUrl": o.IconURL, "value": o.Value, "access": o.Access, "isDefault": o.IsDefault, "owned": o.Access == "free" || owned[o.ItemKey]}
			if o.Access == "item" {
				v["itemKey"] = o.ItemKey
				if it, ok := items[o.ItemKey]; ok {
					v["item"] = it
				}
			}
			opts = append(opts, v)
		}
		list = append(list, map[string]any{"key": p.Key, "name": langs.Resolve(p.Name, lang), "layer": p.Layer, "required": p.Required, "options": opts})
	}
	saved, err := s.savedLooks(ctx, []string{userID})
	if err != nil {
		return nil, err
	}
	out["parts"] = list
	out["look"] = resolveLook(parts, saved[userID], owned)
	return out, nil
}

func (s *service) rpcAvatarSave(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Look map[string]string `json:"look"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	st, err := loadAvatarSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if !st.Enabled {
		return nil, avatarFail("AVATAR_DISABLED", codes.FailedPrecondition, nil)
	}
	parts, err := loadAvatarCatalog(ctx, s.db)
	if err != nil {
		return nil, err
	}
	byKey := map[string]avatarPart{}
	for _, p := range parts {
		byKey[p.Key] = p
	}
	owned, err := s.ownedItemKeys(ctx, userID, avatarItemKeys(parts))
	if err != nil {
		return nil, err
	}
	clean := map[string]string{}
	keys := make([]string, 0, len(in.Look))
	for k := range in.Look {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, partKey := range keys {
		optKey := strings.TrimSpace(in.Look[partKey])
		p, ok := byKey[partKey]
		if !ok {
			return nil, avatarFail("AVATAR_PART_NOT_FOUND", codes.NotFound, map[string]any{"part": partKey})
		}
		if optKey == "" {
			if p.Required {
				return nil, avatarFail("AVATAR_PART_REQUIRED", codes.InvalidArgument, map[string]any{"part": partKey})
			}
			clean[partKey] = ""
			continue
		}
		o, ok := p.option(optKey)
		if !ok {
			return nil, avatarFail("AVATAR_OPTION_NOT_FOUND", codes.NotFound, map[string]any{"part": partKey, "option": optKey})
		}
		if o.Access == "item" && !owned[o.ItemKey] {
			return nil, avatarFail("AVATAR_OPTION_LOCKED", codes.PermissionDenied, map[string]any{"part": partKey, "option": optKey, "itemKey": o.ItemKey})
		}
		clean[partKey] = optKey
	}
	raw, _ := json.Marshal(clean)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO base_avatar_look (user_id, look, updated_at) VALUES ($1::uuid, $2::jsonb, now())
		ON CONFLICT (user_id) DO UPDATE SET look=EXCLUDED.look, updated_at=now()`, userID, string(raw)); err != nil {
		return nil, err
	}
	return map[string]any{"look": resolveLook(parts, clean, owned)}, nil
}

func (s *service) rpcAvatarGet(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		UserIDs []string `json:"userIds"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if len(in.UserIDs) == 0 || len(in.UserIDs) > avatarMaxIDs {
		return nil, badRequest("INVALID_USER_IDS", "userIds", "1 to 100 user ids")
	}
	ids := []string{}
	seen := map[string]bool{}
	for _, id := range in.UserIDs {
		id = strings.TrimSpace(id)
		if id == "" || len(id) > 128 || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	st, err := loadAvatarSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	looks := map[string]avatarLook{}
	if !st.Enabled {
		for _, id := range ids {
			looks[id] = avatarLook{Parts: map[string]string{}, Layers: []avatarLayer{}}
		}
		return map[string]any{"enabled": false, "looks": looks}, nil
	}
	parts, err := loadAvatarCatalog(ctx, s.db)
	if err != nil {
		return nil, err
	}
	saved, err := s.savedLooks(ctx, ids)
	if err != nil {
		return nil, err
	}
	keys := avatarItemKeys(parts)
	for _, id := range ids {
		owned := map[string]bool{}
		if len(saved[id]) > 0 {
			if owned, err = s.ownedItemKeys(ctx, id, keys); err != nil {
				return nil, err
			}
		}
		looks[id] = resolveLook(parts, saved[id], owned)
	}
	return map[string]any{"enabled": true, "looks": looks}, nil
}

func (s *service) avatarSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := loadAvatarSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *service) avatarSettingsPut(w http.ResponseWriter, r *http.Request) {
	in := avatarSettings{Enabled: true}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	raw, _ := json.Marshal(map[string]bool{"enabled": in.Enabled})
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO base_avatar_settings (id, settings, updated_by, updated_at) VALUES (1, $1::jsonb, $2, now())
		ON CONFLICT (id) DO UPDATE SET settings=EXCLUDED.settings, updated_by=EXCLUDED.updated_by, updated_at=now()`, string(raw), consoleUser(r)); err != nil {
		s.writeError(w, err)
		return
	}
	s.avatarSettingsGet(w, r)
}

const avatarPartJSON = `jsonb_build_object('id',p.id,'key',p.key,'name',p.name,'layer',p.layer,'required',p.required,'sort',p.sort,'isActive',p.is_active,
	'createdBy',p.created_by,'createdAt',p.created_at,'updatedAt',p.updated_at,
	'count',(SELECT count(*) FROM base_avatar_option o WHERE o.part_id=p.id),
	'players',(SELECT count(*) FROM base_avatar_look l WHERE l.look ? p.key AND l.look->>p.key <> ''),
	'options',(SELECT COALESCE(jsonb_agg(jsonb_build_object('key',o.key,'name',o.name,'imageUrl',o.image_url,'iconUrl',o.icon_url,'value',o.value,'access',o.access,'itemKey',o.item_key,'isDefault',o.is_default,'isActive',o.is_active) ORDER BY o.sort, o.key),'[]'::jsonb) FROM base_avatar_option o WHERE o.part_id=p.id))`

func (s *service) avatarPartList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		where.add("(p.key ILIKE ? OR "+translatedSearch("p.name", "?")+")", likePattern(p.Q), likePattern(p.Q))
	}
	if v := r.URL.Query().Get("isActive"); v == "true" || v == "false" {
		where.add("p.is_active = ?", v == "true")
	}
	out, err := s.listRaw(r, ` FROM base_avatar_part p`, avatarPartJSON, "p.sort, p.key", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func fetchAvatarPart(ctx context.Context, q queryer, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+avatarPartJSON+`::text FROM base_avatar_part p WHERE p.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func (s *service) avatarPartGet(w http.ResponseWriter, r *http.Request) {
	raw, err := fetchAvatarPart(r.Context(), s.db, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

type avatarOptionIn struct {
	Key       string         `json:"key"`
	Name      localized.Text `json:"name"`
	ImageURL  string         `json:"imageUrl"`
	IconURL   string         `json:"iconUrl"`
	Value     string         `json:"value"`
	Access    string         `json:"access"`
	ItemKey   string         `json:"itemKey"`
	IsDefault bool           `json:"isDefault"`
	IsActive  *bool          `json:"isActive"`
}

func (s *service) avatarPartWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Key      string           `json:"key"`
			Name     localized.Text   `json:"name"`
			Layer    int              `json:"layer"`
			Required bool             `json:"required"`
			Sort     int              `json:"sort"`
			IsActive *bool            `json:"isActive"`
			Options  []avatarOptionIn `json:"options"`
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
		if err == nil && (in.Layer < -1000 || in.Layer > 1000) {
			err = badRequest("INVALID_LAYER", "layer", "between -1000 and 1000")
		}
		if err == nil && (in.Sort < -100000 || in.Sort > 100000) {
			err = badRequest("INVALID_SORT", "sort", "between -100000 and 100000")
		}
		if err == nil && (len(in.Options) == 0 || len(in.Options) > avatarMaxOptions) {
			err = badRequest("INVALID_OPTIONS", "options", "1 to 300 options")
		}
		seen := map[string]bool{}
		defaults, free := 0, 0
		itemKeys := []string{}
		for i := 0; err == nil && i < len(in.Options); i++ {
			o := &in.Options[i]
			field := "options[" + itoa(i) + "]"
			if o.Key, err = cleanKey(field+".key", o.Key); err != nil {
				break
			}
			if seen[o.Key] {
				err = badRequest("DUPLICATE_KEY", field+".key", "each option needs its own key")
				break
			}
			seen[o.Key] = true
			if o.Name, err = cleanText(langs, field+".name", o.Name, false, 100); err != nil {
				break
			}
			if o.ImageURL, err = cleanURL(field+".imageUrl", o.ImageURL); err != nil {
				break
			}
			if o.IconURL, err = cleanURL(field+".iconUrl", o.IconURL); err != nil {
				break
			}
			o.Value = strings.TrimSpace(o.Value)
			if len([]rune(o.Value)) > 64 {
				err = badRequest("INVALID_VALUE", field+".value", "at most 64 letters")
				break
			}
			if o.ImageURL == "" && o.Value == "" {
				err = badRequest("OPTION_LOOK_REQUIRED", field+".imageUrl", "give the option a picture or a value")
				break
			}
			o.ItemKey = strings.TrimSpace(o.ItemKey)
			switch o.Access {
			case "", "free":
				o.Access, o.ItemKey = "free", ""
			case "item":
				if o.ItemKey == "" {
					err = badRequest("ITEM_REQUIRED", field+".itemKey", "pick the item that opens this option")
				}
				itemKeys = append(itemKeys, o.ItemKey)
			default:
				err = badRequest("INVALID_ACCESS", field+".access", "free or item")
			}
			if err != nil {
				break
			}
			active := o.IsActive == nil || *o.IsActive
			if o.IsDefault {
				if o.Access != "free" || !active {
					err = badRequest("DEFAULT_MUST_BE_FREE", field+".isDefault", "the default option must be free and active")
					break
				}
				defaults++
			}
			if o.Access == "free" && active {
				free++
			}
		}
		if err == nil && defaults > 1 {
			err = badRequest("ONE_DEFAULT", "options", "only one option can be the default")
		}
		if err == nil && in.Required && free == 0 {
			err = badRequest("REQUIRED_NEEDS_FREE", "required", "a required part needs at least one free active option")
		}
		if err == nil && len(itemKeys) > 0 {
			uniq := map[string]bool{}
			list := []string{}
			for _, k := range itemKeys {
				if !uniq[k] {
					uniq[k] = true
					list = append(list, k)
				}
			}
			err = s.checkKeysExist(ctx, "economy_item", "itemKey", list)
		}
		if err != nil {
			s.writeError(w, err)
			return
		}
		active := in.IsActive == nil || *in.IsActive
		name, _ := json.Marshal(in.Name)
		var out json.RawMessage
		err = s.inTx(ctx, func(tx *sql.Tx) error {
			if create {
				var taken bool
				if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_avatar_part WHERE key=$1)`, key).Scan(&taken); err != nil {
					return err
				}
				if taken {
					return conflict("KEY_EXISTS", "another avatar part has this key")
				}
				if _, err := tx.ExecContext(ctx, `INSERT INTO base_avatar_part (id, key, name, layer, required, sort, is_active, created_by) VALUES ($1,$2,$3::jsonb,$4,$5,$6,$7,$8)`,
					id, key, string(name), in.Layer, in.Required, in.Sort, active, consoleUser(r)); err != nil {
					return err
				}
			} else {
				res, err := tx.ExecContext(ctx, `UPDATE base_avatar_part SET name=$2::jsonb, layer=$3, required=$4, sort=$5, is_active=$6, updated_at=now() WHERE id=$1`,
					id, string(name), in.Layer, in.Required, in.Sort, active)
				if err != nil {
					return err
				}
				if n, _ := res.RowsAffected(); n == 0 {
					return notFound("NOT_FOUND")
				}
			}
			keys := make([]string, 0, len(in.Options))
			for i, o := range in.Options {
				on, _ := json.Marshal(o.Name)
				act := o.IsActive == nil || *o.IsActive
				if _, err := tx.ExecContext(ctx, `INSERT INTO base_avatar_option (id, part_id, key, name, image_url, icon_url, value, access, item_key, is_default, sort, is_active)
					VALUES ($1,$2,$3,$4::jsonb,$5,$6,$7,$8,$9,$10,$11,$12)
					ON CONFLICT (part_id, key) DO UPDATE SET name=EXCLUDED.name, image_url=EXCLUDED.image_url, icon_url=EXCLUDED.icon_url, value=EXCLUDED.value, access=EXCLUDED.access,
					item_key=EXCLUDED.item_key, is_default=EXCLUDED.is_default, sort=EXCLUDED.sort, is_active=EXCLUDED.is_active`,
					uuid.Must(uuid.NewV4()).String(), id, o.Key, string(on), o.ImageURL, o.IconURL, o.Value, o.Access, o.ItemKey, o.IsDefault, i, act); err != nil {
					return err
				}
				keys = append(keys, o.Key)
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM base_avatar_option WHERE part_id=$1 AND NOT (key = ANY($2::text[]))`, id, keys); err != nil {
				return err
			}
			out, err = fetchAvatarPart(ctx, tx, id)
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

func (s *service) avatarPartDelete(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM base_avatar_part WHERE id=$1`, id)
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

func (s *service) playerAvatar(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if err := s.playerExists(r.Context(), id); err != nil {
		s.writeError(w, err)
		return
	}
	parts, err := loadAvatarCatalog(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	saved, err := s.savedLooks(r.Context(), []string{id})
	if err != nil {
		s.writeError(w, err)
		return
	}
	owned, err := s.ownedItemKeys(r.Context(), id, avatarItemKeys(parts))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"saved": saved[id] != nil, "look": resolveLook(parts, saved[id], owned)})
}
