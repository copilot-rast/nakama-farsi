package server

import (
	"context"
	"database/sql"
	"strings"

	"github.com/heroiclabs/nakama/v3/console"
	"go.uber.org/zap"
)

var baseLanguagesAudit = consoleModuleAudit{
	prefix:   "/v2/console/base/",
	label:    "base",
	resource: console.AclResources_SETTINGS,
	invokes:  []string{"/setup/claim"},
}

var basePlayersAudit = consoleModuleAudit{
	prefix:   "/v2/console/base/",
	label:    "base",
	resource: console.AclResources_ACCOUNT,
	invokes:  []string{"/lift"},
}

var baseChatAudit = consoleModuleAudit{
	prefix:   "/v2/console/base/",
	label:    "base",
	resource: console.AclResources_CHANNEL_MESSAGE,
	invokes:  []string{"/resolve", "/chat-words/test"},
}

var baseSupportAudit = consoleModuleAudit{
	prefix:   "/v2/console/base/",
	label:    "base",
	resource: console.AclResources_ACCOUNT_NOTES,
	invokes:  []string{"/heartbeat", "/take", "/typing", "/read", "/transfer", "/close", "/status", "/assign"},
	skipBody: func(pattern string) bool { return strings.HasSuffix(pattern, "/uploads") },
}

var baseUpdatesAudit = consoleModuleAudit{
	prefix:   "/v2/console/base/",
	label:    "base",
	resource: console.AclResources_SETTINGS,
	invokes:  []string{"/updates/check"},
	mask:     maskBaseUpdatesBody,
}

func maskBaseUpdatesBody(_ string, body any) any {
	m, ok := body.(map[string]any)
	if !ok {
		return body
	}
	if v, ok := m["token"].(string); ok && v != "" {
		m["token"] = "********"
	}
	return m
}

var baseTaxonomyAudit = consoleModuleAudit{
	prefix:   "/v2/console/base/",
	label:    "base",
	resource: console.AclResources_HIRO_ECONOMY,
}

func consoleBaseAuditLog(ctx context.Context, logger *zap.Logger, db *sql.DB, method, pattern, urlPath string, body []byte) {
	mod := baseTaxonomyAudit
	for _, prefix := range []string{"/v2/console/base/languages", "/v2/console/base/brand", "/v2/console/base/setup", "/v2/console/base/account-rules", "/v2/console/base/friend-rules", "/v2/console/base/app-versions", "/v2/console/base/countries", "/v2/console/base/remote-config"} {
		if strings.HasPrefix(pattern, prefix) {
			mod = baseLanguagesAudit
		}
	}
	if strings.HasPrefix(pattern, "/v2/console/base/updates") {
		mod = baseUpdatesAudit
	}
	if strings.HasPrefix(pattern, "/v2/console/base/players") || strings.HasPrefix(pattern, "/v2/console/base/restrictions") || strings.HasPrefix(pattern, "/v2/console/base/activity") {
		mod = basePlayersAudit
	}
	for _, prefix := range []string{"/v2/console/base/chat-words", "/v2/console/base/report-titles", "/v2/console/base/reports"} {
		if strings.HasPrefix(pattern, prefix) {
			mod = baseChatAudit
		}
	}
	if strings.HasPrefix(pattern, "/v2/console/base/support/") {
		if strings.HasSuffix(pattern, "/heartbeat") || strings.HasSuffix(pattern, "/typing") || strings.HasSuffix(pattern, "/read") {
			return
		}
		mod = baseSupportAudit
	}
	consoleModuleAuditLog(ctx, logger, db, mod, method, pattern, urlPath, body)
}
