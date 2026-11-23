package main

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/heroiclabs/nakama/v3/internal/configtransfer"
)

const paymentMethodTable = "economy_payment_method"

type paymentTransfer struct{ s *service }

func (s *service) transferAreas() []configtransfer.Area {
	areas := configtransfer.EconomyAreas()
	areas = append(areas, configtransfer.Area{
		Key: "payment_methods", Owner: "economy", Order: 200,
		Title:  map[string]string{"fa": "روش‌های پرداخت (بدون کلیدها)", "en": "Payment methods (without keys)"},
		Note:   map[string]string{"fa": "کلیدها و شناسه‌های پذیرنده هیچ‌وقت در فایل نمی‌آیند. روشی که روی این سرور کلید ندارد فعال نمی‌شود.", "en": "Keys and merchant ids never go into the file. A method without keys on this server is not switched on."},
		Custom: paymentTransfer{s: s},
	})
	return areas
}

func plainMethodConfig(method string, cfg methodConfig) map[string]any {
	out := map[string]any{}
	for _, f := range methodFields[method] {
		if secretType(f.Type) || configtransfer.SecretName(f.Name) {
			continue
		}
		if f.Type == "bool" {
			out[f.Name] = cfg.flag(f.Name)
			continue
		}
		out[f.Name] = cfg.str(f.Name)
	}
	return out
}

func (p paymentTransfer) Export(ctx context.Context, q configtransfer.Querier) ([]configtransfer.FileTable, error) {
	rows, err := q.QueryContext(ctx, `SELECT method FROM economy_payment_method ORDER BY method`)
	if err != nil {
		return nil, err
	}
	var methods []string
	for rows.Next() {
		var m string
		if err := rows.Scan(&m); err != nil {
			rows.Close()
			return nil, err
		}
		methods = append(methods, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ft := configtransfer.FileTable{Table: paymentMethodTable, Rows: []configtransfer.Row{}}
	for _, m := range methods {
		if !validMethod(m) {
			continue
		}
		pm, err := p.s.loadMethod(ctx, q, m)
		if err != nil {
			return nil, err
		}
		ft.Rows = append(ft.Rows, configtransfer.Row{Key: m, Values: map[string]any{"method": m, "is_active": pm.Active, "config": plainMethodConfig(m, pm.Config)}})
	}
	return []configtransfer.FileTable{ft}, nil
}

func (p paymentTransfer) Apply(ctx context.Context, q configtransfer.Querier, in []configtransfer.FileTable, pick func(table, key string) bool, by string) ([]configtransfer.RowResult, error) {
	var out []configtransfer.RowResult
	for _, ft := range in {
		for _, row := range ft.Rows {
			rr := configtransfer.RowResult{Table: ft.Table, Key: row.Key}
			method, _ := row.Values["method"].(string)
			if method == "" {
				method = row.Key
			}
			rr.Key = method
			switch {
			case ft.Table != paymentMethodTable || !validMethod(method):
				rr.Status, rr.Reason = configtransfer.StatusSkipped, "unknown_method"
			case !pick(ft.Table, method):
				rr.Status = configtransfer.StatusSkipped
			default:
				if err := p.applyOne(ctx, q, method, row.Values, &rr); err != nil {
					return nil, err
				}
			}
			out = append(out, rr)
		}
	}
	return out, nil
}

func (p paymentTransfer) applyOne(ctx context.Context, q configtransfer.Querier, method string, values map[string]any, rr *configtransfer.RowResult) error {
	old, err := p.s.loadMethod(ctx, q, method)
	if err != nil {
		rr.Status, rr.Reason = configtransfer.StatusError, "config_unreadable"
		return nil
	}
	active, _ := values["is_active"].(bool)
	given := map[string]any{}
	if cfg, ok := values["config"].(map[string]any); ok {
		for _, f := range methodFields[method] {
			if secretType(f.Type) || configtransfer.SecretName(f.Name) {
				continue
			}
			if v, ok := cfg[f.Name]; ok {
				given[f.Name] = v
			}
		}
	}
	next, err := mergeMethodConfig(method, old.Config, given)
	if err != nil {
		rr.Status, rr.Reason = configtransfer.StatusError, errCode(err)
		return nil
	}
	if active && !methodConfigured(method, next) {
		rr.Status, rr.Reason = configtransfer.StatusSecretNeeded, "keys_missing"
		return nil
	}
	before, after := plainMethodConfig(method, old.Config), plainMethodConfig(method, next)
	names := make([]string, 0, len(after))
	for k := range after {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		if jsonText(before[k]) != jsonText(after[k]) {
			rr.Changes = append(rr.Changes, configtransfer.Change{Field: "config." + k, From: before[k], To: after[k]})
		}
	}
	if old.Active != active {
		rr.Changes = append(rr.Changes, configtransfer.Change{Field: "is_active", From: old.Active, To: active})
	}
	if len(rr.Changes) == 0 {
		rr.Status = configtransfer.StatusSame
		return nil
	}
	rr.Status = configtransfer.StatusChanged
	var exists bool
	if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_payment_method WHERE method=$1)`, method).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		rr.Status = configtransfer.StatusNew
	}
	plain, _ := json.Marshal(next)
	sealed, err := seal(p.s.cfg.sealKey, plain)
	if err != nil {
		return err
	}
	_, err = q.ExecContext(ctx, `INSERT INTO economy_payment_method (method, is_active, config, updated_at) VALUES ($1,$2,$3,now())
		ON CONFLICT (method) DO UPDATE SET is_active=EXCLUDED.is_active, config=EXCLUDED.config, updated_at=now()`, method, active, sealed)
	return err
}

func jsonText(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func errCode(err error) string {
	if app, ok := asApp(err); ok {
		return app.Code
	}
	return "invalid"
}
