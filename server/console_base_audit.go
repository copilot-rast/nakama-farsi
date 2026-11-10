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

var baseTaxonomyAudit = consoleModuleAudit{
	prefix:   "/v2/console/base/",
	label:    "base",
	resource: console.AclResources_HIRO_ECONOMY,
}

func consoleBaseAuditLog(ctx context.Context, logger *zap.Logger, db *sql.DB, method, pattern, urlPath string, body []byte) {
	mod := baseTaxonomyAudit
	if strings.HasPrefix(pattern, "/v2/console/base/languages") || strings.HasPrefix(pattern, "/v2/console/base/brand") || strings.HasPrefix(pattern, "/v2/console/base/setup") {
		mod = baseLanguagesAudit
	}
	consoleModuleAuditLog(ctx, logger, db, mod, method, pattern, urlPath, body)
}
