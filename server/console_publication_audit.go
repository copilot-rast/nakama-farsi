package server

import (
	"context"
	"database/sql"
	"strings"

	"github.com/heroiclabs/nakama/v3/console"
	"go.uber.org/zap"
)

var publicationPlainConfigFields = map[string]bool{
	"sender": true, "lineNumber": true, "accountSid": true, "from": true, "fromName": true,
	"host": true, "port": true, "security": true, "username": true,
}

var publicationAudit = consoleModuleAudit{
	prefix:   "/v2/console/publication/",
	label:    "publication",
	resource: console.AclResources_NOTIFICATION,
	invokes:  []string{"/test"},
	mask: func(pattern string, body any) any {
		if strings.HasPrefix(pattern, "/v2/console/publication/providers") {
			return maskPublicationProviderBody(body)
		}
		return body
	},
}

func consolePublicationAuditLog(ctx context.Context, logger *zap.Logger, db *sql.DB, method, pattern, urlPath string, body []byte) {
	consoleModuleAuditLog(ctx, logger, db, publicationAudit, method, pattern, urlPath, body)
}

func maskPublicationProviderBody(body any) any {
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
		if publicationPlainConfigFields[k] {
			masked[k] = v
			continue
		}
		masked[k] = "********"
	}
	m["config"] = masked
	return m
}
