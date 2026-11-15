package hookchain

import (
	"context"
	"database/sql"
	"reflect"
	"sync"

	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
)

type BeforeFn[T any] func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in T) (T, error)

type AfterFn[O, I any] func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out O, in I) error

type AfterRtFn func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out, in *rtapi.Envelope) error

var (
	mu     sync.Mutex
	chains = map[string]any{}
)

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Interface, reflect.Func:
		return rv.IsNil()
	}
	return false
}

func Before[T any](name string, fn BeforeFn[T]) BeforeFn[T] {
	mu.Lock()
	defer mu.Unlock()
	list, _ := chains["before:"+name].([]BeforeFn[T])
	list = append(append([]BeforeFn[T]{}, list...), fn)
	chains["before:"+name] = list
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in T) (T, error) {
		for _, f := range list {
			out, err := f(ctx, logger, db, nk, in)
			if err != nil {
				return out, err
			}
			if isNil(out) {
				return out, nil
			}
			in = out
		}
		return in, nil
	}
}

func AfterRt(name string, fn AfterRtFn) AfterRtFn {
	mu.Lock()
	defer mu.Unlock()
	list, _ := chains["after-rt:"+name].([]AfterRtFn)
	list = append(append([]AfterRtFn{}, list...), fn)
	chains["after-rt:"+name] = list
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out, in *rtapi.Envelope) error {
		var first error
		for _, f := range list {
			if err := f(ctx, logger, db, nk, out, in); err != nil && first == nil {
				first = err
			}
		}
		return first
	}
}

func After[O, I any](name string, fn AfterFn[O, I]) AfterFn[O, I] {
	mu.Lock()
	defer mu.Unlock()
	list, _ := chains["after:"+name].([]AfterFn[O, I])
	list = append(append([]AfterFn[O, I]{}, list...), fn)
	chains["after:"+name] = list
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out O, in I) error {
		var first error
		for _, f := range list {
			if err := f(ctx, logger, db, nk, out, in); err != nil && first == nil {
				first = err
			}
		}
		return first
	}
}

type AfterInFn[I any] func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in I) error

func AfterIn[I any](name string, fn AfterInFn[I]) AfterInFn[I] {
	mu.Lock()
	defer mu.Unlock()
	list, _ := chains["after-in:"+name].([]AfterInFn[I])
	list = append(append([]AfterInFn[I]{}, list...), fn)
	chains["after-in:"+name] = list
	return func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in I) error {
		var first error
		for _, f := range list {
			if err := f(ctx, logger, db, nk, in); err != nil && first == nil {
				first = err
			}
		}
		return first
	}
}

// RunBefore runs the before hooks registered so far under name, as the server
// would for the matching API call. With none registered it returns in. An RPC
// that does the work of a built-in API call uses it so the same rules apply.
func RunBefore[T any](ctx context.Context, name string, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in T) (T, error) {
	mu.Lock()
	list, _ := chains["before:"+name].([]BeforeFn[T])
	mu.Unlock()
	for _, f := range list {
		out, err := f(ctx, logger, db, nk, in)
		if err != nil {
			return out, err
		}
		if isNil(out) {
			return out, nil
		}
		in = out
	}
	return in, nil
}

// RunAfter runs the after hooks registered so far under name and returns the
// first error after running all of them.
func RunAfter[O, I any](ctx context.Context, name string, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out O, in I) error {
	mu.Lock()
	list, _ := chains["after:"+name].([]AfterFn[O, I])
	mu.Unlock()
	var first error
	for _, f := range list {
		if err := f(ctx, logger, db, nk, out, in); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// RunAfterIn is RunAfter for hooks that only receive the request.
func RunAfterIn[I any](ctx context.Context, name string, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in I) error {
	mu.Lock()
	list, _ := chains["after-in:"+name].([]AfterInFn[I])
	mu.Unlock()
	var first error
	for _, f := range list {
		if err := f(ctx, logger, db, nk, in); err != nil && first == nil {
			first = err
		}
	}
	return first
}
