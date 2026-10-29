package main

import (
	"context"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func (s *service) clientRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"base_languages":    s.rpcLanguages,
		"base_set_language": s.rpcSetLanguage,
		"base_categories":   s.rpcCategories,
		"base_tags":         s.rpcTags,
	}
}

func (s *service) rpcLanguages(ctx context.Context, userID, payload string) (any, error) {
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
	rows, err := s.db.QueryContext(ctx, `SELECT code, name::text, native_name, direction, icon_url FROM base_language WHERE is_active ORDER BY is_default DESC, sort, code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var code, name, native, dir, icon string
		if err := rows.Scan(&code, &name, &native, &dir, &icon); err != nil {
			return nil, err
		}
		resolved := langs.Resolve(localized.Parse(name), lang)
		if resolved == "" {
			resolved = native
		}
		items = append(items, map[string]any{"code": code, "name": resolved, "nativeName": native, "direction": dir, "iconUrl": icon, "isDefault": code == langs.Default})
	}
	return map[string]any{"default": langs.Default, "lang": lang, "items": items}, rows.Err()
}

func (s *service) rpcSetLanguage(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, err := s.languages(ctx, s.db)
	if err != nil {
		return nil, err
	}
	code := localized.Normalize(in.Lang)
	if code == "" || !langs.IsActive(code) {
		return nil, badRequest("LANGUAGE_NOT_AVAILABLE", "lang", "this language is not available")
	}
	if err := s.nk.AccountUpdateId(ctx, userID, "", nil, "", "", "", code, ""); err != nil {
		return nil, err
	}
	return map[string]any{"lang": code}, nil
}

func (s *service) taxonomyItems(ctx context.Context, userID, requested, table, where string, args ...any) (any, error) {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, requested)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT key, key_type, name::text, description::text, icon_url, metadata::text, sort FROM `+table+` WHERE `+where+` ORDER BY sort, key, key_type`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var key, kt, name, desc, icon, meta string
		var sort int
		if err := rows.Scan(&key, &kt, &name, &desc, &icon, &meta, &sort); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"key": key, "keyType": kt, "name": langs.Resolve(localized.Parse(name), lang), "description": langs.Resolve(localized.Parse(desc), lang),
			"iconUrl": icon, "metadata": rawJSON(meta), "sort": sort})
	}
	return map[string]any{"lang": lang, "items": items}, rows.Err()
}

func (s *service) rpcCategories(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		KeyType string `json:"keyType"`
		Lang    string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	kt := strings.TrimSpace(in.KeyType)
	if !validKeyType(categoryKeyTypes, kt) {
		return nil, badRequest("INVALID_KEY_TYPE", "keyType", "keyType must be one of "+strings.Join(categoryKeyTypes, ", "))
	}
	return s.taxonomyItems(ctx, userID, in.Lang, "base_category", "key_type = $1", kt)
}

func (s *service) rpcTags(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		KeyType string `json:"keyType"`
		Lang    string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	kt := strings.TrimSpace(in.KeyType)
	if kt == "" {
		return s.taxonomyItems(ctx, userID, in.Lang, "base_tag", "true")
	}
	if !validKeyType(tagKeyTypes, kt) {
		return nil, badRequest("INVALID_KEY_TYPE", "keyType", "keyType must be one of "+strings.Join(tagKeyTypes, ", "))
	}
	return s.taxonomyItems(ctx, userID, in.Lang, "base_tag", "key_type IN ($1, 'general')", kt)
}
