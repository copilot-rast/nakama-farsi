package economyhooks

import (
	"context"
	"database/sql"
)

type ItemBuy struct {
	UserID    string
	ItemID    string
	ItemKey   string
	TypeKey   string
	ParentKey string
}

type (
	ItemBuyCheck func(ctx context.Context, tx *sql.Tx, b ItemBuy) error
	SlotResolver func(ctx context.Context, itemID, typeKey string) string
)

var (
	buyChecks  = map[string]ItemBuyCheck{}
	slotByHook = map[string]SlotResolver{}
)

func SetItemBuyCheck(owner string, fn ItemBuyCheck) { set(buyChecks, owner, fn) }
func SetSlotResolver(owner string, fn SlotResolver) { set(slotByHook, owner, fn) }

func CheckItemBuy(ctx context.Context, tx *sql.Tx, b ItemBuy) error {
	for _, fn := range sorted(buyChecks) {
		if fn == nil {
			continue
		}
		if err := fn(ctx, tx, b); err != nil {
			return err
		}
	}
	return nil
}

func Slot(ctx context.Context, itemID, typeKey string) string {
	for _, fn := range sorted(slotByHook) {
		if fn == nil {
			continue
		}
		if s := fn(ctx, itemID, typeKey); s != "" {
			return s
		}
	}
	return ""
}

type optionsKey struct{}

func WithOptions(ctx context.Context, names ...string) context.Context {
	if len(names) == 0 {
		return ctx
	}
	set := map[string]bool{}
	if prev, ok := ctx.Value(optionsKey{}).(map[string]bool); ok {
		for k := range prev {
			set[k] = true
		}
	}
	for _, n := range names {
		if n != "" {
			set[n] = true
		}
	}
	return context.WithValue(ctx, optionsKey{}, set)
}

func HasOption(ctx context.Context, name string) bool {
	set, _ := ctx.Value(optionsKey{}).(map[string]bool)
	return set[name]
}
