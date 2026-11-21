package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/inventoryx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

var requestNS = uuid.Must(uuid.FromString("8b6d2f0e-5c3a-4e7b-9f41-2a7d9c1e6b53"))

type result struct {
	ItemKey  string  `json:"itemKey"`
	Action   string  `json:"action"`
	Track    string  `json:"track,omitempty"`
	Level    int     `json:"level,omitempty"`
	Value    float64 `json:"value,omitempty"`
	Owned    bool    `json:"owned,omitempty"`
	Paid     []price `json:"paid"`
	Replayed bool    `json:"replayed,omitempty"`
	fresh    bool
}

func cleanRequestID(raw string, required bool) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		if required {
			return "", apierr.BadRequest("REQUEST_ID_REQUIRED").WithField("requestId")
		}
		return uuid.Must(uuid.NewV4()).String(), nil
	}
	if len(id) > 64 {
		return "", apierr.BadRequest("REQUEST_ID_REQUIRED").WithField("requestId")
	}
	return id, nil
}

func grantID(userID, requestID string) string {
	return uuid.NewV5(requestNS, userID+":"+requestID).String()
}

func claim(ctx context.Context, tx *sql.Tx, userID, requestID, action, itemKey, detail string, hours int) (*result, error) {
	if _, err := tx.ExecContext(ctx, `DELETE FROM inventoryx_request WHERE user_id=$1 AND request_id=$2 AND expires_at < now()`, userID, requestID); err != nil {
		return nil, err
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO inventoryx_request (user_id, request_id, action, item_key, detail, expires_at) VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (user_id, request_id) DO NOTHING`, userID, requestID, action, itemKey, detail, time.Now().Add(time.Duration(hours)*time.Hour))
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n > 0 {
		return nil, nil
	}
	var a, k, d, raw string
	if err := tx.QueryRowContext(ctx, `SELECT action, item_key, detail, result::text FROM inventoryx_request WHERE user_id=$1 AND request_id=$2`, userID, requestID).Scan(&a, &k, &d, &raw); err != nil {
		return nil, err
	}
	if a != action || k != itemKey || d != detail {
		return nil, apierr.Conflict("REQUEST_ID_REUSED").WithField("requestId")
	}
	var r result
	_ = json.Unmarshal([]byte(raw), &r)
	if r.Paid == nil {
		r.Paid = []price{}
	}
	r.Replayed = true
	return &r, nil
}

func keep(ctx context.Context, tx *sql.Tx, userID, requestID string, r result) error {
	b, _ := json.Marshal(r)
	_, err := tx.ExecContext(ctx, `UPDATE inventoryx_request SET result=$3 WHERE user_id=$1 AND request_id=$2`, userID, requestID, string(b))
	return err
}

func loadItem(ctx context.Context, tx *sql.Tx, key string, stoppedOK bool) (itemRow, itemConfig, error) {
	rows, err := tx.QueryContext(ctx, itemSelect+` WHERE i.key=$1`, strings.TrimSpace(key))
	if err != nil {
		return itemRow{}, itemConfig{}, err
	}
	items, err := scanItems(rows)
	if err != nil {
		return itemRow{}, itemConfig{}, err
	}
	if len(items) == 0 {
		return itemRow{}, itemConfig{}, apierr.NotFound("ITEM_NOT_FOUND").WithField("key")
	}
	if !items[0].Active && !stoppedOK {
		return itemRow{}, itemConfig{}, apierr.Conflict("ITEM_SALE_STOPPED").WithField("key")
	}
	cfgs, err := loadConfigs(ctx, tx, []string{items[0].ID})
	if err != nil {
		return itemRow{}, itemConfig{}, err
	}
	return items[0], cfgs[items[0].ID], nil
}

func debitError(err error) error {
	var ins *pluginkit.ErrInsufficient
	if errors.As(err, &ins) {
		return apierr.New(http.StatusPaymentRequired, "INSUFFICIENT_BALANCE").With("currency", ins.Path)
	}
	return err
}

func priceMap(list []price) map[string]int64 {
	out := map[string]int64{}
	for _, p := range list {
		out[p.CurrencyKey] += p.Amount
	}
	return out
}

func owns(ctx context.Context, tx *sql.Tx, userID, key string) (bool, error) {
	var ok bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_user_item u JOIN economy_item i ON i.id=u.item_id WHERE u.user_id=$1 AND i.key=$2)
		OR EXISTS (SELECT 1 FROM economy_item WHERE key=$2 AND is_starter AND is_active)`, userID, key).Scan(&ok)
	return ok, err
}

func (s *service) unlock(ctx context.Context, userID, key, rawRequest, currency string) (result, error) {
	cfg := s.settings(ctx)
	requestID, err := cleanRequestID(rawRequest, cfg.Unlock.RequireRequestID)
	if err != nil {
		return result{}, err
	}
	key = strings.TrimSpace(key)
	currency = strings.TrimSpace(currency)
	gid := grantID(userID, requestID)
	var out result
	err = pluginkit.InTx(ctx, s.db, func(tx *sql.Tx) error {
		prev, err := claim(ctx, tx, userID, requestID, "unlock", key, currency, cfg.RequestHours)
		if err != nil || prev != nil {
			if prev != nil {
				out = *prev
			}
			return err
		}
		it, c, err := loadItem(ctx, tx, key, false)
		if err != nil {
			return err
		}
		if len(c.UnlockPrices) == 0 {
			return apierr.Conflict("UNLOCK_NOT_AVAILABLE").WithDetail("this item cannot be unlocked early")
		}
		var owned bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_user_item WHERE user_id=$1 AND item_id=$2)`, userID, it.ID).Scan(&owned); err != nil {
			return err
		}
		if owned || it.Starter {
			return apierr.Conflict("ALREADY_OWNED")
		}
		level := 0
		if it.UnlockLevel > 0 {
			if level, err = s.level(ctx, userID, it.TableID); err != nil {
				return err
			}
		}
		if it.UnlockLevel == 0 || level >= it.UnlockLevel {
			return apierr.Conflict("ALREADY_UNLOCKED").With("level", level).With("unlockLevel", it.UnlockLevel)
		}
		if cfg.Unlock.CheckRequirement && it.Needs != "" {
			ok, err := owns(ctx, tx, userID, it.Needs)
			if err != nil {
				return err
			}
			if !ok {
				return apierr.Conflict("REQUIRES_ITEM").With("requiresKey", it.Needs)
			}
		}
		pay, err := pickPrices(cfg, c.UnlockPrices, currency)
		if err != nil {
			return err
		}
		paid, _ := json.Marshal(pay)
		res, err := tx.ExecContext(ctx, `INSERT INTO inventoryx_user_unlock (user_id, item_id, paid, grant_id) VALUES ($1,$2,$3,$4) ON CONFLICT (user_id, item_id) DO NOTHING`, userID, it.ID, string(paid), gid)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return apierr.Conflict("ALREADY_UNLOCKED").With("level", level).With("unlockLevel", it.UnlockLevel)
		}
		if err := pluginkit.Debit(ctx, s.nk, tx, userID, gid, priceMap(pay), map[string]any{"module": owner, "reason": "item_unlock", "sourceType": "item", "sourceId": it.Key, "itemKey": it.Key, "requestId": requestID}); err != nil {
			return debitError(err)
		}
		gave := false
		if !it.ForSale && cfg.Unlock.OwnFreeItems {
			res, err := tx.ExecContext(ctx, `INSERT INTO economy_user_item (user_id, item_id, source, source_ref) VALUES ($1,$2,'unlock',$3) ON CONFLICT (user_id, item_id) DO NOTHING`, userID, it.ID, gid)
			if err != nil {
				return err
			}
			n, _ := res.RowsAffected()
			gave = n > 0
		}
		out = result{ItemKey: it.Key, Action: "unlock", Paid: pay, Owned: gave, Level: it.UnlockLevel, fresh: true}
		return keep(ctx, tx, userID, requestID, out)
	})
	if err == nil && out.fresh {
		activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "unlocked", RefID: out.ItemKey, Actor: "player",
			Data: map[string]any{"item": out.ItemKey, "early": true, "paid": out.Paid, "owned": out.Owned}}))
		s.publish(ctx, gid, userID, "item_unlocked", map[string]any{"itemKey": out.ItemKey, "early": true}, out)
	}
	return out, err
}

func pickPrices(cfg settings.All, list []price, currency string) ([]price, error) {
	if cfg.Unlock.PriceMode == settings.PriceAll {
		return list, nil
	}
	if currency == "" {
		if len(list) == 1 {
			return list, nil
		}
		keys := []string{}
		for _, p := range list {
			keys = append(keys, p.CurrencyKey)
		}
		return nil, apierr.BadRequest("CURRENCY_REQUIRED").WithField("currencyKey").With("currencies", keys)
	}
	for _, p := range list {
		if p.CurrencyKey == currency {
			return []price{p}, nil
		}
	}
	return nil, apierr.BadRequest("CURRENCY_NOT_ACCEPTED").WithField("currencyKey")
}

func (s *service) upgrade(ctx context.Context, userID, key, trackKey string, fromLevel *int, rawRequest string) (result, error) {
	cfg := s.settings(ctx)
	requestID, err := cleanRequestID(rawRequest, cfg.Upgrade.RequireRequestID)
	if err != nil {
		return result{}, err
	}
	if fromLevel == nil && cfg.Upgrade.RequireFromLevel {
		return result{}, apierr.BadRequest("FROM_LEVEL_REQUIRED").WithField("fromLevel")
	}
	key = strings.TrimSpace(key)
	trackKey = strings.TrimSpace(trackKey)
	detail := trackKey + "#"
	if fromLevel != nil {
		detail += strconv.Itoa(*fromLevel)
	}
	gid := grantID(userID, requestID)
	var out result
	err = pluginkit.InTx(ctx, s.db, func(tx *sql.Tx) error {
		prev, err := claim(ctx, tx, userID, requestID, "upgrade", key, detail, cfg.RequestHours)
		if err != nil || prev != nil {
			if prev != nil {
				out = *prev
			}
			return err
		}
		it, c, err := loadItem(ctx, tx, key, cfg.Upgrade.AllowsStopped())
		if err != nil {
			return err
		}
		t, ok := c.track(trackKey)
		if !ok {
			return apierr.NotFound("TRACK_NOT_FOUND").WithField("track").With("track", trackKey)
		}
		if it.Starter {
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_item (user_id, item_id, source, source_ref) VALUES ($1,$2,'starter','') ON CONFLICT (user_id, item_id) DO NOTHING`, userID, it.ID); err != nil {
				return err
			}
		}
		var one int
		err = tx.QueryRowContext(ctx, `SELECT 1 FROM economy_user_item WHERE user_id=$1 AND item_id=$2 FOR UPDATE`, userID, it.ID).Scan(&one)
		if errors.Is(err, sql.ErrNoRows) {
			return apierr.Conflict("NOT_OWNED")
		}
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO inventoryx_user_track (user_id, item_id, track) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, userID, it.ID, t.Key); err != nil {
			return err
		}
		levels := map[string]int{}
		rows, err := tx.QueryContext(ctx, `SELECT track, level FROM inventoryx_user_track WHERE user_id=$1 AND item_id=$2 ORDER BY track FOR UPDATE`, userID, it.ID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var k string
			var lv int
			if err := rows.Scan(&k, &lv); err != nil {
				rows.Close()
				return err
			}
			levels[k] = lv
		}
		rows.Close()
		level := levels[t.Key]
		if fromLevel != nil && *fromLevel != level {
			return apierr.Conflict("UPGRADE_CHANGED").With("level", level).With("track", t.Key)
		}
		if level >= len(t.Levels) {
			return apierr.Conflict("UPGRADE_MAXED").With("level", level).With("maxLevel", len(t.Levels)).With("track", t.Key)
		}
		next := t.Levels[level]
		if next.MinPlayerLevel > 0 {
			pl, err := s.level(ctx, userID, it.TableID)
			if err != nil {
				return err
			}
			if pl < next.MinPlayerLevel {
				return apierr.Forbidden("LEVEL_TOO_LOW").With("level", pl).With("unlockLevel", next.MinPlayerLevel).With("track", t.Key)
			}
		}
		for _, r := range next.Requires {
			if levels[r.Track] < r.Level {
				return apierr.Conflict("TRACK_REQUIREMENT").With("track", r.Track).With("level", r.Level).With("current", levels[r.Track])
			}
		}
		paid := []price{}
		if next.Price > 0 {
			paid = append(paid, price{CurrencyKey: t.CurrencyKey, Amount: next.Price})
			if err := pluginkit.Debit(ctx, s.nk, tx, userID, gid, priceMap(paid), map[string]any{"module": owner, "reason": "item_track_upgrade", "sourceType": "item", "sourceId": it.Key,
				"itemKey": it.Key, "track": t.Key, "level": level + 1, "requestId": requestID}); err != nil {
				return debitError(err)
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE inventoryx_user_track SET level=$4, updated_at=now() WHERE user_id=$1 AND item_id=$2 AND track=$3`, userID, it.ID, t.Key, level+1); err != nil {
			return err
		}
		out = result{ItemKey: it.Key, Action: "upgrade", Track: t.Key, Level: level + 1, Value: t.valueAt(level + 1), Paid: paid, fresh: true}
		return keep(ctx, tx, userID, requestID, out)
	})
	if err == nil && out.fresh {
		activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindItem, Action: "upgraded", RefID: out.ItemKey, Actor: "player",
			Data: map[string]any{"item": out.ItemKey, "track": out.Track, "level": out.Level, "paid": out.Paid}}))
		s.publish(ctx, gid, userID, "item_upgraded", map[string]any{"itemKey": out.ItemKey, "track": out.Track, "level": out.Level}, out)
	}
	return out, err
}

func (s *service) publish(ctx context.Context, id, userID, name string, fields map[string]any, r result) {
	if !gameevents.Has() || id == "" {
		return
	}
	typed := []gameevents.Typed{}
	for _, p := range r.Paid {
		if p.Amount > 0 && p.CurrencyKey != "" {
			typed = append(typed, gameevents.Resource("Sink", p.CurrencyKey, float64(p.Amount), r.Action, r.ItemKey))
		}
	}
	if _, errs := gameevents.Publish(ctx, gameevents.Event{ID: id, UserID: userID, Name: name, Fields: fields, Source: owner, Typed: typed}); len(errs) > 0 && s.logger != nil {
		s.logger.Debug("Inventoryx game event %s not recorded: %v", name, errs)
	}
}
