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

func consoleEconomyAuditLog(ctx context.Context, logger *zap.Logger, db *sql.DB, method, pattern, urlPath string, body []byte) {
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
		if strings.HasSuffix(pattern, "/inquire") || strings.HasSuffix(pattern, "/retry-grant") || strings.HasSuffix(pattern, "/grant") {
			action = console.AuditLogAction_INVOKE
		}
	default:
		action = console.AuditLogAction_UPDATE
	}

	metadata := map[string]any{"method": method, "path": urlPath}
	if !strings.HasSuffix(pattern, "/uploads") && len(body) > 0 {
		var parsed any
		if err := json.Unmarshal(body, &parsed); err == nil {
			if strings.HasPrefix(pattern, "/v2/console/economy/payment-methods/") {
				parsed = maskEconomyMethodBody(parsed)
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
		Resource:  console.AclResources_HIRO_ECONOMY.String(),
		Action:    action.String(),
		Timestamp: time.Now(),
		Metadata:  string(raw),
		Message:   "economy " + strings.ToLower(action.String()) + " " + strings.TrimPrefix(urlPath, "/v2/console/economy/"),
	}
	if err := auditLogAdd(logger, db, entry); err != nil {
		logger.Error("Failed to add economy audit log entry", zap.Error(err))
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
