package apierr

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/internal/settings"
	"google.golang.org/grpc/codes"
)

func ForModule(ctx context.Context, db *sql.DB, lang string, bare error, e *Error) error {
	var rt *runtime.Error
	if bare == nil || !errors.As(bare, &rt) {
		return bare
	}
	if cfg, _ := settings.Load(ctx, db); cfg.Errors.ModuleMode != settings.ModeEnvelope {
		return bare
	}
	if e == nil {
		code := strings.TrimSpace(rt.Message)
		if !CodePattern.MatchString(code) {
			return bare
		}
		e = New(StatusFromGRPC(codes.Code(rt.Code)), code)
	}
	userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	env := Build(ctx, db, ResolveLang(ctx, db, userID, lang, headerValue(ctx, "Accept-Language")), e)
	raw, err := json.Marshal(env)
	if err != nil {
		return bare
	}
	return runtime.NewError(string(raw), int(rt.Code))
}

func PayloadLang(payload string) string {
	return payloadLang(payload)
}
