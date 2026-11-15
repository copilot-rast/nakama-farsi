package main

import (
	"context"
	"strings"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func (s *service) clientRPCs() map[string]clientRPC {
	rpcs := map[string]clientRPC{
		"base_languages":      s.rpcLanguages,
		"base_set_language":   s.rpcSetLanguage,
		"base_categories":     s.rpcCategories,
		"base_tags":           s.rpcTags,
		"base_brand":          s.rpcBrand,
		"base_account_rules":  s.rpcAccountRules,
		"base_app_version":    s.rpcAppVersion,
		"base_countries":      s.rpcCountries,
		"base_set_country":    s.rpcSetCountry,
		"base_my_tags":        s.rpcMyTags,
		"base_my_status":      s.rpcMyStatus,
		"base_report_titles":  s.rpcReportTitles,
		"base_report_message": s.rpcReportMessage,
		"base_friend_rules":   s.rpcFriendRules,
		"base_chat_quota":     s.rpcChatQuota,
		"base_report_player":  s.rpcReportPlayer,
		"base_my_reports":     s.rpcMyReports,
		"voice_token":         s.rpcVoiceToken,

		"base_support_info":           s.rpcSupportInfo,
		"base_support_tickets":        s.rpcTicketList,
		"base_support_ticket_create":  s.rpcTicketCreate,
		"base_support_ticket_get":     s.rpcTicketGet,
		"base_support_ticket_reply":   s.rpcTicketReply,
		"base_support_ticket_resolve": s.rpcTicketResolve,
		"base_support_ticket_rate":    s.rpcTicketRate,
		"base_support_upload_begin":   s.rpcUploadBegin,
		"base_support_upload_part":    s.rpcUploadPart,
		"base_support_upload_finish":  s.rpcUploadFinish,
		"base_support_chat_open":      s.rpcChatOpen,
		"base_support_chat_get":       s.rpcChatGet,
		"base_support_chat_send":      s.rpcChatSend,
		"base_support_chat_typing":    s.rpcChatTyping,
		"base_support_chat_close":     s.rpcChatClose,
		"base_support_chat_rate":      s.rpcChatRate,
		"base_remote_config":          s.rpcRemoteConfig,
		"base_my_activity":            s.rpcMyActivity,
	}
	for name, fn := range s.socialRPCs() {
		rpcs[name] = fn
	}
	return rpcs
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
		return s.taxonomyItems(ctx, userID, in.Lang, "base_tag", "key_type <> 'user'")
	}
	if !validKeyType(clientTagKeyTypes, kt) {
		return nil, badRequest("INVALID_KEY_TYPE", "keyType", "keyType must be one of "+strings.Join(clientTagKeyTypes, ", "))
	}
	return s.taxonomyItems(ctx, userID, in.Lang, "base_tag", "key_type IN ($1, 'general')", kt)
}
