package apierr

import (
	"errors"
	"net/http"
	"regexp"

	"google.golang.org/grpc/codes"
)

var CodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,79}$`)

type Error struct {
	Code       string
	Status     int
	Detail     string
	Field      string
	Meta       map[string]any
	RetryAfter int
}

func (e *Error) Error() string {
	if e.Detail != "" {
		return e.Code + ": " + e.Detail
	}
	return e.Code
}

func New(status int, code string) *Error {
	return &Error{Code: code, Status: status}
}

func BadRequest(code string) *Error   { return New(http.StatusBadRequest, code) }
func Unauthorized(code string) *Error { return New(http.StatusUnauthorized, code) }
func Forbidden(code string) *Error    { return New(http.StatusForbidden, code) }
func NotFound(code string) *Error     { return New(http.StatusNotFound, code) }
func Conflict(code string) *Error     { return New(http.StatusConflict, code) }
func Unavailable(code string) *Error  { return New(http.StatusServiceUnavailable, code) }

func TooManyRequests(code string, retryAfterSeconds int) *Error {
	return New(http.StatusTooManyRequests, code).Retry(retryAfterSeconds)
}

func (e *Error) With(key string, value any) *Error {
	if e.Meta == nil {
		e.Meta = map[string]any{}
	}
	e.Meta[key] = value
	return e
}

func (e *Error) WithDetail(detail string) *Error {
	e.Detail = detail
	return e
}

func (e *Error) WithField(field string) *Error {
	e.Field = field
	return e
}

func (e *Error) Retry(seconds int) *Error {
	if seconds < 0 {
		seconds = 0
	}
	e.RetryAfter = seconds
	return e
}

func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

func GRPCCode(status int) codes.Code {
	switch status {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusPaymentRequired, http.StatusPreconditionFailed, http.StatusLocked:
		return codes.FailedPrecondition
	case http.StatusServiceUnavailable, http.StatusBadGateway:
		return codes.Unavailable
	case http.StatusGatewayTimeout, http.StatusRequestTimeout:
		return codes.DeadlineExceeded
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusRequestEntityTooLarge:
		return codes.InvalidArgument
	default:
		if status >= 200 && status < 300 {
			return codes.OK
		}
		return codes.Internal
	}
}

func StatusFromGRPC(c codes.Code) int {
	switch c {
	case codes.InvalidArgument, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Canceled:
		return 499
	case codes.Unimplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}
