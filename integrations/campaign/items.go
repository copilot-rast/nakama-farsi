package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
)

func itemParam(payload string, cfg settings.Items) string {
	var raw map[string]any
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		return ""
	}
	for _, name := range []string{cfg.Param, "itemKey"} {
		if v, ok := raw[name].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func valuesAt(v any, path []string) []string {
	if len(path) == 0 {
		switch t := v.(type) {
		case string:
			if s := strings.TrimSpace(t); s != "" {
				return []string{s}
			}
		case []any:
			out := []string{}
			for _, x := range t {
				out = append(out, valuesAt(x, nil)...)
			}
			return out
		}
		return nil
	}
	switch t := v.(type) {
	case map[string]any:
		return valuesAt(t[path[0]], path[1:])
	case []any:
		out := []string{}
		for _, x := range t {
			out = append(out, valuesAt(x, path)...)
		}
		return out
	}
	return nil
}

func stageValues(st map[string]any, paths []string) []string {
	out := []string{}
	for _, p := range paths {
		out = append(out, valuesAt(st, strings.Split(p, "."))...)
	}
	return out
}

func hasFold(list []string, names ...string) bool {
	for _, x := range list {
		for _, n := range names {
			if n != "" && strings.EqualFold(x, n) {
				return true
			}
		}
	}
	return false
}

type gateItem struct {
	ID     string
	Key    string
	Type   string
	GameID string
	Class  string
}

func (s *service) loadGateItem(ctx context.Context, key, classField string) (*gateItem, error) {
	var it gateItem
	q := `SELECT i.id::text, i.key, c.key, COALESCE(i.metadata->>'gameId',''), '' FROM economy_item i JOIN base_category c ON c.id=i.type_id WHERE (i.key=$1 OR lower(i.metadata->>'gameId')=lower($1)) AND i.is_active ORDER BY (i.key=$1) DESC LIMIT 1`
	err := s.db.QueryRowContext(ctx, q, key).Scan(&it.ID, &it.Key, &it.Type, &it.GameID, &it.Class)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var has bool
	if err := s.db.QueryRowContext(ctx, `SELECT to_regclass('itemkind_profile') IS NOT NULL`).Scan(&has); err == nil && has {
		var gid, class string
		if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(data->>'gameId',''), COALESCE(data->>$2,'') FROM itemkind_profile WHERE item_id=$1 AND status<>'deleted'`, it.ID, classField).Scan(&gid, &class); err == nil {
			if it.GameID == "" {
				it.GameID = gid
			}
			it.Class = class
		}
	}
	return &it, nil
}

func (s *service) checkStageItem(ctx context.Context, userID string, st *stage, cfg settings.Items, key string) (map[string]any, error) {
	if !cfg.Enforce {
		return nil, nil
	}
	allow := stageValues(st.Settings, cfg.Allow)
	deny := stageValues(st.Settings, cfg.Deny)
	classes := stageValues(st.Settings, cfg.Classes)
	ruled := len(allow) > 0 || len(deny) > 0 || len(classes) > 0
	if key == "" {
		if ruled && cfg.RequireItem {
			return nil, apierr.BadRequest("STAGE_ITEM_REQUIRED").WithField(cfg.Param)
		}
		return nil, nil
	}
	it, err := s.loadGateItem(ctx, key, cfg.ClassField)
	if err != nil {
		return nil, err
	}
	if it == nil {
		return nil, apierr.NotFound("ITEM_NOT_FOUND").WithField(cfg.Param)
	}
	refuse := func(why string) error {
		return apierr.Forbidden("STAGE_ITEM_NOT_ALLOWED").WithField(cfg.Param).With("itemKey", it.Key).With("reason", why)
	}
	if cfg.ItemType != "" && it.Type != cfg.ItemType {
		return nil, refuse("type")
	}
	if cfg.RequireOwned {
		var owned bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_user_item WHERE user_id=$1 AND item_id=$2)`, userID, it.ID).Scan(&owned); err != nil {
			return nil, err
		}
		if !owned {
			return nil, apierr.Forbidden("ITEM_NOT_OWNED").WithField(cfg.Param).With("itemKey", it.Key)
		}
	}
	if len(allow) > 0 && !hasFold(allow, it.Key, it.GameID) {
		return nil, refuse("not_in_allowed")
	}
	if hasFold(deny, it.Key, it.GameID) {
		return nil, refuse("excluded")
	}
	if len(classes) > 0 && !hasFold(classes, it.Class) {
		return nil, refuse("class")
	}
	return map[string]any{"itemKey": it.Key, "itemType": it.Type}, nil
}
