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
