package apierr

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"google.golang.org/grpc/codes"
)

type Envelope struct {
	Code              string         `json:"code"`
	Status            int            `json:"status"`
	Message           string         `json:"message"`
	Meta              map[string]any `json:"meta,omitempty"`
	RetryAfterSeconds *int           `json:"retryAfterSeconds,omitempty"`
	Field             string         `json:"field,omitempty"`
	Detail            string         `json:"detail,omitempty"`
	Lang              string         `json:"lang,omitempty"`
}

func Normalize(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := As(err); ok {
		return e
	}
	var rt *runtime.Error
	if errors.As(err, &rt) {
		code := strings.TrimSpace(rt.Message)
		if CodePattern.MatchString(code) {
			return New(StatusFromGRPC(codes.Code(rt.Code)), code)
		}
		return New(StatusFromGRPC(codes.Code(rt.Code)), "INTERNAL_ERROR").WithDetail(rt.Message)
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return New(http.StatusGatewayTimeout, "TIMEOUT")
	case errors.Is(err, context.Canceled):
		return New(499, "CANCELED")
	}
	return New(http.StatusInternalServerError, "INTERNAL_ERROR")
}

func Build(ctx context.Context, db *sql.DB, lang string, err error) Envelope {
	e := Normalize(err)
	cfg, _ := settings.Load(ctx, db)
	env := Envelope{Code: e.Code, Status: e.Status, Meta: e.Meta, Field: e.Field, Lang: lang}
	if cfg.Errors.IncludeDetail && e.Status < 500 {
		env.Detail = e.Detail
	}
	entries, _, cerr := Catalog(ctx, db)
	var entry Entry
	var known bool
	if cerr == nil {
		entry, known = entries[e.Code]
	}
	if env.Status == 0 {
		env.Status = entry.Status
		if env.Status == 0 {
			env.Status = http.StatusBadRequest
		}
	}
	if e.RetryAfter > 0 {
		r := e.RetryAfter
		env.RetryAfterSeconds = &r
	} else if known && entry.RetryAfter != nil && *entry.RetryAfter > 0 {
		r := *entry.RetryAfter
		env.RetryAfterSeconds = &r
	}
	meta := map[string]any{}
	for k, v := range e.Meta {
		meta[k] = v
	}
	if env.RetryAfterSeconds != nil {
		meta["retryAfter"] = *env.RetryAfterSeconds
	}
	if known {
		env.Message = Fill(entry.MessageFor(lang, FallbackLang(ctx, db)), meta, cfg.Errors.LocalizeDigits && UsesPersianDigits(lang))
	}
	if env.Message == "" {
		if e.Detail != "" && e.Status < 500 {
			env.Message = e.Detail
		} else {
			env.Message = e.Code
		}
	}
	return env
}

func FallbackLang(ctx context.Context, db *sql.DB) string {
	cfg, _ := settings.Load(ctx, db)
	if cfg.Errors.FallbackLang != "" {
		return cfg.Errors.FallbackLang
	}
	if db == nil {
		return "en"
	}
	if set, err := localized.Load(ctx, db); err == nil && set.Default != "" {
		return set.Default
	}
	return "en"
}

func headerValue(ctx context.Context, name string) string {
	h, _ := ctx.Value(runtime.RUNTIME_CTX_HEADERS).(map[string][]string)
	for k, v := range h {
		if strings.EqualFold(k, name) && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}

func acceptLanguage(raw string) []string {
	out := []string{}
	for _, part := range strings.Split(raw, ",") {
		tag := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		if tag != "" && tag != "*" {
			out = append(out, tag)
		}
	}
	return out
}

func ResolveLang(ctx context.Context, db *sql.DB, userID, requested string, header string) string {
	cfg, _ := settings.Load(ctx, db)
	if db == nil {
		if requested != "" {
			return localized.Normalize(requested)
		}
		return "en"
	}
	set, err := localized.Load(ctx, db)
	if err != nil || set == nil {
		if requested != "" {
			return localized.Normalize(requested)
		}
		return "en"
	}
	candidates := []string{requested}
	account := ""
	if _, err := uuid.FromString(userID); err == nil {
		account = localized.UserLangTag(ctx, db, userID)
	}
	if cfg.Errors.PreferAccountLanguage {
		candidates = append(candidates, account)
		candidates = append(candidates, acceptLanguage(header)...)
	} else {
		candidates = append(candidates, acceptLanguage(header)...)
		candidates = append(candidates, account)
	}
	if cfg.Errors.FallbackLang != "" {
		candidates = append(candidates, cfg.Errors.FallbackLang)
	}
	return set.Choose(candidates...)
}

func ToRuntime(ctx context.Context, db *sql.DB, userID, requestedLang string, err error) error {
	if err == nil {
		return nil
	}
	e := Normalize(err)
	cfg, _ := settings.Load(ctx, db)
	grpc := int(GRPCCode(e.Status))
	if e.Status == 0 {
		grpc = int(codes.InvalidArgument)
	}
	if cfg.Errors.Mode == settings.ModeCode {
		return runtime.NewError(e.Code, grpc)
	}
	lang := ResolveLang(ctx, db, userID, requestedLang, headerValue(ctx, "Accept-Language"))
	env := Build(ctx, db, lang, e)
	raw, mErr := json.Marshal(env)
	if mErr != nil {
		return runtime.NewError(e.Code, grpc)
	}
	return runtime.NewError(string(raw), int(GRPCCode(env.Status)))
}

type Handler func(ctx context.Context, userID, payload string) (any, error)

type RPCOptions struct {
	MaxPayload     int
	AllowAnonymous bool
	Logger         runtime.Logger
}

func RPC(db *sql.DB, opts RPCOptions, fn Handler) func(context.Context, runtime.Logger, *sql.DB, runtime.NakamaModule, string) (string, error) {
	if opts.MaxPayload == 0 {
		opts.MaxPayload = 16384
	}
	return func(ctx context.Context, logger runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
		userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
		lang := payloadLang(payload)
		if _, err := uuid.FromString(userID); err != nil && !opts.AllowAnonymous {
			return "", ToRuntime(ctx, db, "", lang, Unauthorized("UNAUTHENTICATED"))
		}
		if len(payload) > opts.MaxPayload {
			return "", ToRuntime(ctx, db, userID, lang, BadRequest("REQUEST_TOO_LARGE").With("max", opts.MaxPayload))
		}
		out, err := fn(ctx, userID, payload)
		if err != nil {
			e := Normalize(err)
			if e.Status >= 500 {
				l := opts.Logger
				if l == nil {
					l = logger
				}
				l.Error("RPC failed: %v", err)
			}
			return "", ToRuntime(ctx, db, userID, lang, e)
		}
		body, err := json.Marshal(out)
		if err != nil {
			return "", ToRuntime(ctx, db, userID, lang, New(http.StatusInternalServerError, "INTERNAL_ERROR"))
		}
		return string(body), nil
	}
}

func payloadLang(payload string) string {
	if !strings.Contains(payload, `"lang"`) {
		return ""
	}
	var in struct {
		Lang string `json:"lang"`
	}
	_ = json.Unmarshal([]byte(payload), &in)
	return in.Lang
}

func WriteHTTP(w http.ResponseWriter, r *http.Request, db *sql.DB, err error) {
	e := Normalize(err)
	lang := ResolveLang(r.Context(), db, "", r.URL.Query().Get("lang"), r.Header.Get("Accept-Language"))
	env := Build(r.Context(), db, lang, e)
	if env.RetryAfterSeconds != nil {
		w.Header().Set("Retry-After", strconv.Itoa(*env.RetryAfterSeconds))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(env.Status)
	_ = json.NewEncoder(w).Encode(env)
}

func Decode(payload string, out any) error {
	if strings.TrimSpace(payload) == "" {
		payload = "{}"
	}
	if err := json.Unmarshal([]byte(payload), out); err != nil {
		return BadRequest("INVALID_JSON").WithDetail("request body is not valid JSON for this call")
	}
	return nil
}
