package main

import (
	"context"
	"errors"
	"math"
	"net/http"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

type playCost struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
	Spent       int64  `json:"spent"`
	Free        bool   `json:"free"`
	Via         string `json:"via"`
}

type refundRef struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
	Via         string `json:"via"`
	Done        bool   `json:"done"`
	Refunded    int64  `json:"refunded"`
}

func stageCost(cfg settings.All, st *stage) (string, int64) {
	if cfg.Play.CostCurrency == "" {
		return "", 0
	}
	amount := cfg.Play.CostAmount
	if v, ok := st.Metadata["playCost"].(float64); ok && v >= 0 && v <= 1000000 && v == math.Trunc(v) {
		amount = int64(v)
	}
	return cfg.Play.CostCurrency, amount
}

func costSource(attemptID string) string { return "campaign:" + attemptID }

func (s *service) takeCost(ctx context.Context, userID, attemptID, currency string, amount int64) (*playCost, *economyhooks.Regen, error) {
	pc := &playCost{CurrencyKey: currency, Amount: amount}
	res, err := s.spend(ctx, userID, currency, amount, costSource(attemptID))
	if err == nil {
		pc.Spent, pc.Free, pc.Via = res.Spent, res.Free, "economy"
		view := res.View
		return pc, &view, nil
	}
	if r, ok := economyhooks.AsRefusal(err); ok {
		if r.Code == "NOT_ENOUGH_CURRENCY" {
			e := apierr.Conflict("NOT_ENOUGH_PLAY_COST").With("currencyKey", currency).With("amount", amount)
			for _, k := range []string{"value", "nextAt", "fullAt", "refillPrice"} {
				if v, found := r.Meta[k]; found {
					e = e.With(k, v)
				}
			}
			return nil, nil, e
		}
		status := r.Status
		if status == 0 {
			status = http.StatusConflict
		}
		return nil, nil, apierr.New(status, r.Code)
	}
	if !errors.Is(err, economyhooks.ErrNoRegen) {
		return nil, nil, err
	}
	err = pluginkit.Debit(ctx, s.nk, s.db, userID, costSource(attemptID)+":cost", map[string]int64{currency: amount},
		map[string]any{"module": "campaign", "sourceType": sourceType, "sourceId": attemptID, "reason": "play_cost"})
	var short *pluginkit.ErrInsufficient
	if errors.As(err, &short) {
		return nil, nil, apierr.Conflict("NOT_ENOUGH_PLAY_COST").With("currencyKey", currency).With("amount", amount)
	}
	if err != nil {
		return nil, nil, err
	}
	pc.Spent, pc.Via = amount, "wallet"
	return pc, nil, nil
}

func (s *service) giveBack(ctx context.Context, userID, attemptID string, r *refundRef) error {
	if r.Done {
		return nil
	}
	if r.Via == "economy" {
		out, err := s.refund(ctx, userID, r.CurrencyKey, costSource(attemptID))
		if err != nil {
			return err
		}
		r.Done, r.Refunded = true, out.Refunded
		return nil
	}
	grant := costSource(attemptID) + ":refund"
	applied, err := pluginkit.WalletApplied(ctx, s.db, userID, grant)
	if err != nil {
		return err
	}
	if !applied && r.Amount > 0 {
		md := map[string]any{"grantId": grant, "module": "campaign", "sourceType": sourceType, "sourceId": attemptID, "reason": "play_cost_refund"}
		if _, _, err := s.nk.MultiUpdate(ctx, nil, nil, nil, []*runtime.WalletUpdate{{UserID: userID, Changeset: map[string]int64{r.CurrencyKey: r.Amount}, Metadata: md}}, true); err != nil {
			return err
		}
	}
	r.Done, r.Refunded = true, r.Amount
	return nil
}

func refundFor(cfg settings.All, success bool, raw []byte) *refundRef {
	if !success || !cfg.Play.RefundOnWin || len(raw) == 0 {
		return nil
	}
	var pc playCost
	unmarshal(raw, &pc)
	if pc.CurrencyKey == "" || pc.Spent <= 0 || pc.Free {
		return nil
	}
	return &refundRef{CurrencyKey: pc.CurrencyKey, Amount: pc.Spent, Via: pc.Via}
}
