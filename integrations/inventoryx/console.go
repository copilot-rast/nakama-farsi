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
	"github.com/heroiclabs/nakama/v3/integrations/inventoryx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

func (s *service) registerConsoleHTTP(initializer runtime.Initializer) error {
	type route struct {
		path, method string
		handler      http.HandlerFunc
	}
	p := consolePrefix
	routes := []route{
		{p + "/settings", http.MethodGet, s.settingsGet},
		{p + "/settings", http.MethodPut, s.settingsPut},
		{p + "/items", http.MethodGet, s.itemList},
		{p + "/items/{id}", http.MethodGet, s.itemGet},
		{p + "/items/{id}", http.MethodPut, s.itemPut},
		{p + "/items/{id}", http.MethodDelete, s.itemDelete},
		{p + "/import", http.MethodPost, s.importItems},
		{p + "/import/reports", http.MethodGet, s.importReports},
		{p + "/import/reports/{rid}", http.MethodGet, s.importReports},
		{p + "/players/{id}", http.MethodGet, s.playerGet},
		{p + "/players/{id}/tracks", http.MethodPut, s.playerSetTrack},
		{p + "/players/{id}/unlocks", http.MethodPost, s.playerUnlock},
		{p + "/players/{id}/unlocks/{itemId}", http.MethodDelete, s.playerRelock},
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) settingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, meta, err := settings.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	_, lvErr := s.level(r.Context(), uuid.Nil.String(), "")
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"settings": cfg, "defaults": settings.Defaults(), "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt, "economy": lvErr == nil})
}

func (s *service) settingsPut(w http.ResponseWriter, r *http.Request) {
	in := settings.Defaults()
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if field, code := in.Validate(); code != "" {
		s.fail(w, pluginkit.Bad(code, field, ""))
		return
	}
	if err := settings.Save(r.Context(), s.db, in, pluginkit.ConsoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	s.settingsGet(w, r)
}

func resolveName(langs *localized.Set, lang, raw, fallback string) string {
	if langs != nil {
		if v := langs.Resolve(localized.Parse(raw), lang); v != "" {
			return v
		}
	}
	return fallback
}

func (s *service) itemList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, limit := pluginkit.Page(r, 20, 100)
	q := r.URL.Query()
	args := &pluginkit.Args{}
	where := []string{"TRUE"}
	if v := strings.TrimSpace(q.Get("q")); v != "" {
		like := args.Add(pluginkit.LikePattern(v))
		where = append(where, `(i.key ILIKE `+like+` OR i.name::text ILIKE `+like+`)`)
	}
	if v := strings.TrimSpace(q.Get("typeId")); v != "" {
		where = append(where, `i.type_id::text = `+args.Add(v))
	}
	switch q.Get("configured") {
	case "true":
		where = append(where, `x.item_id IS NOT NULL`)
	case "false":
		where = append(where, `x.item_id IS NULL`)
	case "unlock":
		where = append(where, `jsonb_array_length(x.unlock_prices) > 0`)
	case "tracks":
		where = append(where, `jsonb_array_length(x.tracks) > 0`)
	}
	cond := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM economy_item i LEFT JOIN inventoryx_item x ON x.item_id=i.id WHERE `+cond, args.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	lim, off := args.Add(limit), args.Add((page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT i.id::text, i.key, i.name::text, i.icon_url, COALESCE(c.key,''), COALESCE(c.name::text,'{}'), i.is_active, i.is_starter, COALESCE(i.unlock_level,0), i.price::text,
		COALESCE(x.unlock_prices,'[]')::text, COALESCE(x.tracks,'[]')::text, COALESCE(x.updated_by,''), x.updated_at
		FROM economy_item i LEFT JOIN base_category c ON c.id=i.type_id LEFT JOIN inventoryx_item x ON x.item_id=i.id
		WHERE `+cond+` ORDER BY i.sort, i.key LIMIT `+lim+` OFFSET `+off, args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	langs, lang, _ := localized.ForUser(ctx, s.db, "", q.Get("lang"))
	items := []map[string]any{}
	for rows.Next() {
		var id, key, name, icon, typeKey, typeName, up, tr, by string
		var priceRaw sql.NullString
		var active, starter bool
		var unlockLevel int
		var at sql.NullTime
		if err := rows.Scan(&id, &key, &name, &icon, &typeKey, &typeName, &active, &starter, &unlockLevel, &priceRaw, &up, &tr, &by, &at); err != nil {
			s.fail(w, err)
			return
		}
		var c itemConfig
		_ = json.Unmarshal([]byte(up), &c.UnlockPrices)
		_ = json.Unmarshal([]byte(tr), &c.Tracks)
		var p any
		if priceRaw.Valid {
			_ = json.Unmarshal([]byte(priceRaw.String), &p)
		}
		row := map[string]any{"id": id, "key": key, "name": resolveName(langs, lang, name, key), "iconUrl": icon, "type": typeKey, "typeName": resolveName(langs, lang, typeName, typeKey),
			"isActive": active, "isStarter": starter, "unlockLevel": unlockLevel, "price": p, "unlockPrices": pricesView(c.UnlockPrices), "tracks": c.Tracks, "updatedBy": by}
		if c.Tracks == nil {
			row["tracks"] = []track{}
		}
		if at.Valid {
			row["updatedAt"] = at.Time.UTC().Format(time.RFC3339)
		}
		items = append(items, row)
	}
	if err := rows.Err(); err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) itemID(r *http.Request) (string, string, error) {
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		return "", "", pluginkit.Missing("ITEM_NOT_FOUND")
	}
	var key string
	err := s.db.QueryRowContext(r.Context(), `SELECT key FROM economy_item WHERE id=$1`, id).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", pluginkit.Missing("ITEM_NOT_FOUND")
	}
	return id, key, err
}

func (s *service) itemGet(w http.ResponseWriter, r *http.Request) {
	id, key, err := s.itemID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	cfgs, err := loadConfigs(r.Context(), s.db, []string{id})
	if err != nil {
		s.fail(w, err)
		return
	}
	c := cfgs[id]
	if c.UnlockPrices == nil {
		c.UnlockPrices = []price{}
	}
	if c.Tracks == nil {
		c.Tracks = []track{}
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"id": id, "key": key, "unlockPrices": c.UnlockPrices, "tracks": c.Tracks})
}

func (s *service) cleaner(ctx context.Context) (func(itemConfig) (itemConfig, error), *localized.Set, error) {
	cfg := s.settings(ctx)
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		return nil, nil, err
	}
	cur, err := currencyKeys(ctx, s.db)
	if err != nil {
		return nil, nil, err
	}
	return func(in itemConfig) (itemConfig, error) { return cleanConfig(cfg, langs, cur, in) }, langs, nil
}

func (s *service) itemPut(w http.ResponseWriter, r *http.Request) {
	id, _, err := s.itemID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	var in itemConfig
	if err := pluginkit.ReadBody(r, &in, 512<<10); err != nil {
		s.fail(w, err)
		return
	}
	clean, _, err := s.cleaner(r.Context())
	if err != nil {
		s.fail(w, err)
		return
	}
	c, err := clean(in)
	if err != nil {
		s.fail(w, err)
		return
	}
	if err := saveConfig(r.Context(), s.db, id, c, pluginkit.ConsoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	s.itemGet(w, r)
}

func (s *service) itemDelete(w http.ResponseWriter, r *http.Request) {
	id, _, err := s.itemID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	if _, err := s.db.ExecContext(r.Context(), `DELETE FROM inventoryx_item WHERE item_id=$1`, id); err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) playerID(r *http.Request) (string, error) {
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		return "", pluginkit.Missing("PLAYER_NOT_FOUND")
	}
	var ok bool
	if err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1)`, id).Scan(&ok); err != nil {
		return "", err
	}
	if !ok {
		return "", pluginkit.Missing("PLAYER_NOT_FOUND")
	}
	return id, nil
}

func (s *service) playerGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.playerID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	langs, lang, _ := localized.ForUser(ctx, s.db, "", r.URL.Query().Get("lang"))
	unlocks := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT i.id::text, i.key, i.name::text, u.paid::text, u.actor, u.created_at FROM inventoryx_user_unlock u JOIN economy_item i ON i.id=u.item_id WHERE u.user_id=$1 ORDER BY u.created_at DESC`, userID)
	if err != nil {
		s.fail(w, err)
		return
	}
	for rows.Next() {
		var id, key, name, paid, actor string
		var at time.Time
		if err := rows.Scan(&id, &key, &name, &paid, &actor, &at); err != nil {
			rows.Close()
			s.fail(w, err)
			return
		}
		var p []price
		_ = json.Unmarshal([]byte(paid), &p)
		unlocks = append(unlocks, map[string]any{"itemId": id, "itemKey": key, "name": resolveName(langs, lang, name, key), "paid": pricesView(p), "actor": actor, "at": at.UTC().Format(time.RFC3339)})
	}
	rows.Close()
	tracks := []map[string]any{}
	rows, err = s.db.QueryContext(ctx, `SELECT i.id::text, i.key, i.name::text, x.tracks::text, EXISTS (SELECT 1 FROM economy_user_item o WHERE o.user_id=$1 AND o.item_id=i.id) OR i.is_starter
		FROM inventoryx_item x JOIN economy_item i ON i.id=x.item_id
		WHERE jsonb_array_length(x.tracks) > 0 AND (EXISTS (SELECT 1 FROM economy_user_item o WHERE o.user_id=$1 AND o.item_id=i.id) OR i.is_starter OR EXISTS (SELECT 1 FROM inventoryx_user_track t WHERE t.user_id=$1 AND t.item_id=i.id))
		ORDER BY i.sort, i.key`, userID)
	if err != nil {
		s.fail(w, err)
		return
	}
	type itemTracks struct {
		id, key, name string
		owned         bool
		tracks        []track
	}
	list := []itemTracks{}
	ids := []string{}
	for rows.Next() {
		var it itemTracks
		var raw string
		if err := rows.Scan(&it.id, &it.key, &it.name, &raw, &it.owned); err != nil {
			rows.Close()
			s.fail(w, err)
			return
		}
		_ = json.Unmarshal([]byte(raw), &it.tracks)
		list = append(list, it)
		ids = append(ids, it.id)
	}
	rows.Close()
	st, err := s.loadPlayer(ctx, userID, ids)
	if err != nil {
		s.fail(w, err)
		return
	}
	for _, it := range list {
		levels := st.levels[it.id]
		if levels == nil {
			levels = map[string]int{}
		}
		for _, t := range it.tracks {
			lv := levels[t.Key]
			name := t.Key
			if n := langs.Resolve(t.Name, lang); n != "" {
				name = n
			}
			tracks = append(tracks, map[string]any{"itemId": it.id, "itemKey": it.key, "itemName": resolveName(langs, lang, it.name, it.key), "owned": it.owned,
				"track": t.Key, "trackName": name, "level": lv, "maxLevel": len(t.Levels), "value": t.valueAt(lv), "unit": t.Unit})
		}
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"unlocks": unlocks, "tracks": tracks})
}

func (s *service) playerSetTrack(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.playerID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	var in struct {
		ItemID string `json:"itemId"`
		Track  string `json:"track"`
		Level  int    `json:"level"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if _, err := uuid.FromString(in.ItemID); err != nil {
		s.fail(w, pluginkit.Missing("ITEM_NOT_FOUND"))
		return
	}
	cfgs, err := loadConfigs(ctx, s.db, []string{in.ItemID})
	if err != nil {
		s.fail(w, err)
		return
	}
	t, ok := cfgs[in.ItemID].track(strings.TrimSpace(in.Track))
	if !ok {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusNotFound, Code: "TRACK_NOT_FOUND", Field: "track"})
		return
	}
	if in.Level < 0 || in.Level > len(t.Levels) {
		s.fail(w, pluginkit.Bad("INVALID_LEVEL", "level", "level must be 0 to the track's last level"))
		return
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO inventoryx_user_track (user_id, item_id, track, level) VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id, item_id, track) DO UPDATE SET level=EXCLUDED.level, updated_at=now()`, userID, in.ItemID, t.Key, in.Level); err != nil {
		s.fail(w, err)
		return
	}
	var key string
	_ = s.db.QueryRowContext(ctx, `SELECT key FROM economy_item WHERE id=$1`, in.ItemID).Scan(&key)
	admin := pluginkit.ConsoleUser(r)
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "upgraded", RefID: key, Actor: "admin", Data: map[string]any{"item": key, "track": t.Key, "level": in.Level, "admin": admin}})
	s.playerGet(w, r)
}

func (s *service) playerUnlock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.playerID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	var in struct {
		ItemID string `json:"itemId"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	var key string
	if _, err := uuid.FromString(in.ItemID); err == nil {
		err = s.db.QueryRowContext(ctx, `SELECT key FROM economy_item WHERE id=$1`, in.ItemID).Scan(&key)
	}
	if key == "" {
		s.fail(w, pluginkit.Missing("ITEM_NOT_FOUND"))
		return
	}
	admin := pluginkit.ConsoleUser(r)
	res, err := s.db.ExecContext(ctx, `INSERT INTO inventoryx_user_unlock (user_id, item_id, actor) VALUES ($1,$2,$3) ON CONFLICT (user_id, item_id) DO NOTHING`, userID, in.ItemID, "console:"+admin)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "ALREADY_UNLOCKED"})
		return
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "unlocked", RefID: key, Actor: "admin", Data: map[string]any{"item": key, "early": true, "admin": admin}})
	s.playerGet(w, r)
}

func (s *service) playerRelock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.playerID(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	itemID := pluginkit.PathVar(r, "itemId")
	if _, err := uuid.FromString(itemID); err != nil {
		s.fail(w, pluginkit.Missing("ITEM_NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM inventoryx_user_unlock WHERE user_id=$1 AND item_id=$2`, userID, itemID)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusNotFound, Code: "NOT_UNLOCKED"})
		return
	}
	s.playerGet(w, r)
}
