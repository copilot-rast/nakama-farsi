package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/console"
	"go.uber.org/zap"
)

const consoleUsernameHeader = "X-Nakama-Console-Username"

type consoleModuleAudit struct {
	prefix   string
	label    string
	resource console.AclResources
	invokes  []string
	skipBody func(pattern string) bool
	mask     func(pattern string, body any) any
}

var economyAudit = consoleModuleAudit{
	prefix:   "/v2/console/economy/",
	label:    "economy",
	resource: console.AclResources_HIRO_ECONOMY,
	invokes:  []string{"/inquire", "/retry-grant", "/grant"},
	skipBody: func(pattern string) bool { return strings.HasSuffix(pattern, "/uploads") },
	mask: func(pattern string, body any) any {
		if strings.HasPrefix(pattern, "/v2/console/economy/payment-methods/") {
			return maskEconomyMethodBody(body)
		}
		return body
	},
}

func consoleEconomyAuditLog(ctx context.Context, logger *zap.Logger, db *sql.DB, method, pattern, urlPath string, body []byte) {
	consoleModuleAuditLog(ctx, logger, db, economyAudit, method, pattern, urlPath, body)
}

func consoleModuleAuditLog(ctx context.Context, logger *zap.Logger, db *sql.DB, mod consoleModuleAudit, method, pattern, urlPath string, body []byte) {
	var action console.AuditLogAction
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return
	case http.MethodPut, http.MethodPatch:
		action = console.AuditLogAction_UPDATE
	case http.MethodDelete:
		action = console.AuditLogAction_DELETE
	case http.MethodPost:
		action = console.AuditLogAction_CREATE
		for _, suffix := range mod.invokes {
			if strings.HasSuffix(pattern, suffix) {
				action = console.AuditLogAction_INVOKE
			}
		}
	default:
		action = console.AuditLogAction_UPDATE
	}

	metadata := map[string]any{"method": method, "path": urlPath}
	if (mod.skipBody == nil || !mod.skipBody(pattern)) && len(body) > 0 {
		var parsed any
		if err := json.Unmarshal(body, &parsed); err == nil {
			if mod.mask != nil {
				parsed = mod.mask(pattern, parsed)
			}
			metadata["body"] = parsed
		}
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		raw = []byte("{}")
	}

	var userId uuid.UUID
	if uid, ok := ctx.Value(ctxConsoleUserIdKey{}).(uuid.UUID); ok {
		userId = uid
	}
	username, _ := ctx.Value(ctxConsoleUsernameKey{}).(string)
	email, _ := ctx.Value(ctxConsoleEmailKey{}).(string)

	entry := &AuditLogEntry{
		UserID:    userId.String(),
		Username:  username,
		Email:     email,
		Resource:  mod.resource.String(),
		Action:    action.String(),
		Timestamp: time.Now(),
		Metadata:  string(raw),
		Message:   mod.label + " " + strings.ToLower(action.String()) + " " + strings.TrimPrefix(urlPath, mod.prefix),
	}
	if err := auditLogAdd(logger, db, entry); err != nil {
		logger.Error("Failed to add "+mod.label+" audit log entry", zap.Error(err))
	}
}

func maskEconomyMethodBody(body any) any {
	m, ok := body.(map[string]any)
	if !ok {
		return body
	}
	cfg, ok := m["config"].(map[string]any)
	if !ok {
		return m
	}
	masked := make(map[string]any, len(cfg))
	for k, v := range cfg {
		if b, isBool := v.(bool); isBool {
			masked[k] = b
			continue
		}
		if k == "packageName" {
			masked[k] = v
			continue
		}
		masked[k] = "********"
	}
	m["config"] = masked
	return m
}
