package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/gofrs/uuid/v5"
)

const offerJSON = `jsonb_build_object('id',t.id,'key',t.key,'kind',t.kind,'name',t.name,'description',t.description,'badge',t.badge,'productId',t.product_id,
	'priceMode',t.price_mode,'discountPercent',t.discount_percent,'priceToman',t.price_toman,'coinPrice',t.coin_price,'sku',t.sku,'purchaseLimit',t.purchase_limit,'weight',t.weight,
	'audienceTagIds',to_jsonb(t.audience_tag_ids),'excludeTagIds',to_jsonb(t.exclude_tag_ids),'minLevel',t.min_level,'maxLevel',t.max_level,'levelTableId',t.level_table_id,
	'metaConditions',t.meta_conditions,'minAccountDays',t.min_account_days,'maxAccountDays',t.max_account_days,'triggerType',t.trigger_type,'triggerValue',t.trigger_value,
	'triggerRef',t.trigger_ref,'durationMinutes',t.duration_minutes,'repeat',t.repeat,'cooldownHours',t.cooldown_hours,'maxTriggers',t.max_triggers,
	'startsAt',t.starts_at,'endsAt',t.ends_at,'isActive',t.is_active,'sort',t.sort,'metadata',t.metadata,'createdAt',t.created_at,'updatedAt',t.updated_at)`

func loadOfferTemplates(ctx context.Context, q rowsQueryer, where string, args ...any) ([]offerTemplate, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+offerJSON+`::text FROM economy_offer t `+where+` ORDER BY t.sort, t.created_at, t.id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []offerTemplate{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var o offerTemplate
		if err := json.Unmarshal([]byte(raw), &o); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

type offerCtx struct {
	s        *service
	ctx      context.Context
	userID   string
	st       offerSettings
	now      time.Time
	facts    offerFacts
	products map[string]productView
	prizes   map[string]prizeRow
	counts   map[string]purchaseCount
	owned    map[string]bool
	methods  map[string]paymentMethod
	lc       *levelCache
	metas    map[string]float64
	err      error
}

func (s *service) newOfferCtx(ctx context.Context, userID string, st offerSettings) (*offerCtx, error) {
	oc := &offerCtx{s: s, ctx: ctx, userID: userID, st: st, now: s.now(), products: map[string]productView{}, prizes: map[string]prizeRow{}, lc: newLevelCache(s.db, userID)}
	var err error
	tags, err := s.userTagIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	if oc.counts, err = s.purchaseCounts(ctx, userID); err != nil {
		return nil, err
	}
	if oc.owned, err = s.ownedItemKeys(ctx, userID); err != nil {
		return nil, err
	}
	if oc.methods, err = s.activeMethods(ctx); err != nil {
		return nil, err
	}
	var created time.Time
	var last, prev sql.NullTime
	if err := s.db.QueryRowContext(ctx, `SELECT u.create_time, o.last_login_at, o.prev_login_at FROM users u LEFT JOIN economy_user_offer_state o ON o.user_id=u.id WHERE u.id=$1`, userID).Scan(&created, &last, &prev); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("USER_NOT_FOUND")
		}
		return nil, err
	}
	oc.facts = offerFacts{Tags: tags, Now: oc.now, AccountAge: oc.now.Sub(created)}
	if last.Valid {
		oc.facts.LastLogin = &last.Time
	}
	if prev.Valid {
		oc.facts.PrevLogin = &prev.Time
	}
	oc.facts.Level = func(table string) int {
		lv, err := oc.lc.level(ctx, table)
		if err != nil && oc.err == nil {
			oc.err = err
		}
		return lv
	}
	oc.facts.Games = func(gameKey string) int {
		var n int
		if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(sum(games),0) FROM economy_game_stats WHERE user_id=$1 AND ($2='' OR game_key=$2)`, userID, gameKey).Scan(&n); err != nil && oc.err == nil {
			oc.err = err
		}
		return n
	}
	oc.facts.Meta = func(metaID string) float64 {
		if oc.metas == nil {
			oc.metas = map[string]float64{}
			rows, err := s.db.QueryContext(ctx, `SELECT m.id::text, COALESCE(u.value, m.default_value)::float8 FROM economy_meta m LEFT JOIN economy_user_meta u ON u.meta_id=m.id AND u.user_id=$1`, userID)
			if err != nil {
				oc.err = err
				return 0
			}
			defer rows.Close()
			for rows.Next() {
				var id string
				var v float64
				if err := rows.Scan(&id, &v); err == nil {
					oc.metas[id] = v
				}
			}
		}
		return oc.metas[metaID]
	}
	return oc, nil
}

func (oc *offerCtx) product(id string) (productView, prizeRow, error) {
	p, ok := oc.products[id]
	if !ok {
		var err error
		if p, err = oc.s.loadProduct(oc.ctx, oc.s.db, id); err != nil {
			return p, prizeRow{}, err
		}
		oc.products[id] = p
	}
	prize, ok := oc.prizes[p.PrizeID]
	if !ok {
		var err error
		if prize, err = loadPrize(oc.ctx, oc.s.db, p.PrizeID); err != nil {
			return p, prize, err
		}
		oc.prizes[p.PrizeID] = prize
	}
	return p, prize, nil
}

func (oc *offerCtx) block(o offerTemplate) (string, error) {
	if r := offerBlock(o, oc.facts); r != "" || oc.err != nil {
		return r, oc.err
	}
	p, prize, err := oc.product(o.ProductID)
	if err != nil {
		return "", err
	}
	if !p.IsActive && oc.st.InactiveProducts == "skip" {
		return "product_inactive", nil
	}
	keys := prizeItemKeys(prize.Items)
	if oc.st.ExcludeOwned && len(keys) > 0 {
		all := true
		for _, k := range keys {
			if !oc.owned[k] {
				all = false
				break
			}
		}
		if all {
			return "owned", nil
		}
	}
	if oc.st.RespectProductRules {
		st, err := oc.s.productStateFor(oc.ctx, p, prize, oc.lc, oc.counts, oc.owned)
		if err != nil {
			return "", err
		}
		if st.LevelLocked || st.OverMax {
			return "product_level", nil
		}
		if p.IsOneTime && st.Count > 0 {
			return "product_bought", nil
		}
		if st.Owned {
			return "owned", nil
		}
		if st.Missing != "" {
			return "product_requires", nil
		}
		if !prize.Repeatable {
			claimed, err := prizeClaimedByUser(oc.ctx, oc.s.db, oc.userID, prize.ID)
			if err != nil {
				return "", err
			}
			if claimed {
				return "product_bought", nil
			}
		}
	}
	if len(offerMethods(p, priceOffer(o, p), oc.methods)) == 0 {
		return "no_method", nil
	}
	return "", nil
}

type offerDraw struct {
	Picks   []string
	Reasons map[string]string
	Pool    map[string]offerTemplate
	Count   int
}

func (oc *offerCtx) drawDaily(start time.Time, reroll int, keep []string, avoid map[string]bool) (offerDraw, error) {
	d := offerDraw{Reasons: map[string]string{}, Pool: map[string]offerTemplate{}}
	tpls, err := loadOfferTemplates(oc.ctx, oc.s.db, `WHERE t.kind='daily'`)
	if err != nil {
		return d, err
	}
	recent := map[string]bool{}
	if oc.st.NoRepeatDays > 0 {
		rows, err := oc.s.db.QueryContext(oc.ctx, `SELECT DISTINCT offer_id::text FROM economy_user_offer WHERE user_id=$1 AND kind='daily' AND period_start < $2 AND period_start >= $2 - $3::int * INTERVAL '1 day'`,
			oc.userID, start, oc.st.NoRepeatDays)
		if err != nil {
			return d, err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err == nil {
				recent[id] = true
			}
		}
		rows.Close()
	}
	exclude := map[string]bool{}
	if oc.st.BoughtCooldownDays != 0 {
		rows, err := oc.s.db.QueryContext(oc.ctx, `SELECT offer_id::text, max(last_bought_at) FROM economy_user_offer WHERE user_id=$1 AND kind='daily' AND bought > 0 GROUP BY offer_id`, oc.userID)
		if err != nil {
			return d, err
		}
		for rows.Next() {
			var id string
			var at sql.NullTime
			if err := rows.Scan(&id, &at); err != nil {
				continue
			}
			if oc.st.BoughtCooldownDays < 0 || (at.Valid && oc.now.Sub(at.Time) < time.Duration(oc.st.BoughtCooldownDays)*24*time.Hour) {
				exclude[id] = true
			}
		}
		rows.Close()
	}
	kept := map[string]bool{}
	for _, id := range keep {
		kept[id] = true
	}
	cands := []offerCandidate{}
	for _, o := range tpls {
		d.Pool[o.ID] = o
		reason, err := oc.block(o)
		if err != nil {
			return d, err
		}
		switch {
		case reason != "":
		case kept[o.ID]:
			reason = "kept"
		case avoid[o.ID]:
			reason = "replaced"
		case exclude[o.ID]:
			reason = "bought_recently"
		case o.Weight <= 0 && oc.st.Weighting != "scored":
			reason = "weight_zero"
		}
		if reason != "" {
			d.Reasons[o.ID] = reason
			continue
		}
		if recent[o.ID] {
			d.Reasons[o.ID] = "recent"
		}
		cands = append(cands, offerCandidate{ID: o.ID, Weight: o.Weight, Sort: o.Sort})
	}
	d.Count = oc.st.dailyCount(oc.facts.Tags)
	need := d.Count - len(keep)
	if oc.st.Weighting == "scored" {
		picks, err := oc.drawScored(start, reroll, cands, d.Pool, need)
		if err != nil {
			return d, err
		}
		d.Picks = picks
		for _, id := range d.Picks {
			d.Reasons[id] = "picked"
		}
		return d, nil
	}
	d.Picks = pickOffers(cands, offerPick{Count: need, Mode: oc.st.Weighting, Recent: recent, FillAgain: oc.st.FillWithRepeats, Seed: offerSeed(oc.userID, periodKey(start), strconv.Itoa(reroll))})
	for _, id := range d.Picks {
		d.Reasons[id] = "picked"
	}
	return d, nil
}

func (oc *offerCtx) drawScored(start time.Time, reroll int, cands []offerCandidate, pool map[string]offerTemplate, need int) ([]string, error) {
	if need <= 0 || len(cands) == 0 {
		return []string{}, nil
	}
	sc := oc.st.Score
	wallet, err := oc.s.userWallet(oc.ctx, oc.userID)
	if err != nil {
		return nil, err
	}
	last := map[string]time.Time{}
	rows, err := oc.s.db.QueryContext(oc.ctx, `SELECT offer_id::text, max(period_start) FROM economy_user_offer WHERE user_id=$1 AND kind='daily' AND period_start < $2 AND period_start >= $2 - $3::int * INTERVAL '1 day' GROUP BY offer_id`,
		oc.userID, start, sc.RecencyDays)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id string
		var at time.Time
		if err := rows.Scan(&id, &at); err == nil {
			last[id] = at
		}
	}
	rows.Close()
	list := make([]scoredCandidate, 0, len(cands))
	for _, c := range cands {
		o := pool[c.ID]
		p, _, err := oc.product(o.ProductID)
		if err != nil {
			return nil, err
		}
		sc := scoredCandidate{ID: o.ID, Key: o.Key, Rarity: offerRarity(o.Metadata), CreatedAt: offerCreatedAt(o.Metadata, o.CreatedAt)}
		if p.CoinPrice != nil && p.CoinPrice.CurrencyKey != "" {
			sc.Currency, sc.Price, sc.HasPrice = p.CoinPrice.CurrencyKey, p.CoinPrice.Amount, true
		}
		if at, ok := last[o.ID]; ok {
			t := at
			sc.LastOffered = &t
		}
		list = append(list, sc)
	}
	seed := oc.userID + ":" + start.In(oc.st.zone()).Format("2006-01-02")
	if reroll > 0 {
		seed += ":" + strconv.Itoa(reroll)
	}
	return pickScored(list, need, seed, wallet, start, oc.now, sc), nil
}

func (oc *offerCtx) insertOffer(tx *sql.Tx, o offerTemplate, periodStart *time.Time, slot, reroll int, expires *time.Time, shown bool) (string, error) {
	p, _, err := oc.product(o.ProductID)
	if err != nil {
		return "", err
	}
	price := priceOffer(o, p)
	if f := oc.st.CoinPriceFloor; f > 0 && price.CoinAmount != nil && *price.CoinAmount < f && (price.CoinOriginal == nil || *price.CoinOriginal >= f) {
		price.CoinAmount = i64(f)
	}
	sku, _ := json.Marshal(price.SKU)
	id := uuid.Must(uuid.NewV4()).String()
	var coin any
	if price.CoinCurrency != "" {
		coin = price.CoinCurrency
	}
	var shownAt any
	if shown {
		shownAt = oc.now
	}
	_, err = tx.ExecContext(oc.ctx, `INSERT INTO economy_user_offer (id, user_id, offer_id, kind, period_start, slot, original_toman, final_toman, coin_currency, coin_amount, coin_original, discount_percent, sku, purchase_limit, reroll, triggered_at, shown_at, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		id, oc.userID, o.ID, o.Kind, periodStart, slot, price.OriginalToman, price.FinalToman, coin, price.CoinAmount, price.CoinOriginal, price.DiscountPercent, string(sku), o.PurchaseLimit, reroll, oc.now, shownAt, expires)
	return id, err
}

func (oc *offerCtx) ensureDaily() (time.Time, time.Time, int, error) {
	start, end := oc.st.offerPeriod(oc.now)
	rerolls := 0
	if !oc.st.DailyEnabled {
		return start, end, 0, nil
	}
	err := oc.s.inTx(oc.ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(oc.ctx, `INSERT INTO economy_user_offer_period (user_id, period_start) VALUES ($1,$2) ON CONFLICT DO NOTHING`, oc.userID, start); err != nil {
			return err
		}
		var generated bool
		if err := tx.QueryRowContext(oc.ctx, `SELECT rerolls, generated FROM economy_user_offer_period WHERE user_id=$1 AND period_start=$2 FOR UPDATE`, oc.userID, start).Scan(&rerolls, &generated); err != nil {
			return err
		}
		if generated {
			return nil
		}
		d, err := oc.drawDaily(start, 0, nil, nil)
		if err != nil {
			return err
		}
		for i, id := range d.Picks {
			if _, err := oc.insertOffer(tx, d.Pool[id], &start, i, 0, &end, false); err != nil {
				return err
			}
		}
		_, err = tx.ExecContext(oc.ctx, `UPDATE economy_user_offer_period SET generated=true, updated_at=now() WHERE user_id=$1 AND period_start=$2`, oc.userID, start)
		return err
	})
	return start, end, rerolls, err
}

type entryState struct {
	tpl     offerTemplate
	hist    entryHistory
	holds   bool
	reason  string
	due     bool
	created bool
}

func (oc *offerCtx) entryStates(create bool) ([]entryState, error) {
	tpls, err := loadOfferTemplates(oc.ctx, oc.s.db, `WHERE t.kind='entry'`)
	if err != nil {
		return nil, err
	}
	var out []entryState
	run := func(q rowsQueryer, tx *sql.Tx) error {
		hist := map[string]*entryHistory{}
		rows, err := q.QueryContext(oc.ctx, `SELECT offer_id::text, count(*),
				bool_or(status='active' AND bought < purchase_limit AND (expires_at IS NULL OR expires_at > $2)),
				NULLIF(max(LEAST(COALESCE(expires_at, 'infinity'::timestamptz), CASE WHEN bought >= purchase_limit THEN COALESCE(last_bought_at, 'infinity'::timestamptz) ELSE 'infinity'::timestamptz END,
					CASE WHEN status <> 'active' THEN updated_at ELSE 'infinity'::timestamptz END)), 'infinity'::timestamptz),
				max(triggered_at)
			FROM economy_user_offer WHERE user_id=$1 AND kind='entry' AND NOT forgotten GROUP BY offer_id`, oc.userID, oc.now)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id string
			var h entryHistory
			var end, trig sql.NullTime
			if err := rows.Scan(&id, &h.Count, &h.Open, &end, &trig); err != nil {
				rows.Close()
				return err
			}
			if end.Valid {
				h.LastEnd = &end.Time
			}
			if trig.Valid {
				h.LastTriggered = &trig.Time
			}
			hist[id] = &h
		}
		rows.Close()
		open := 0
		for _, h := range hist {
			if h.Open {
				open++
			}
		}
		out = out[:0]
		for _, o := range tpls {
			e := entryState{tpl: o}
			if h := hist[o.ID]; h != nil {
				e.hist = *h
			}
			e.reason, err = oc.block(o)
			if err != nil {
				return err
			}
			e.holds = entryTriggerHolds(o, oc.facts)
			if oc.err != nil {
				return oc.err
			}
			e.due = e.reason == "" && entryDue(o, e.hist, e.holds, oc.facts)
			if create && e.due && oc.st.EntryEnabled && open < oc.st.EntryMaxActive && tx != nil {
				var expires *time.Time
				if oc.st.EntryTimerStart == "trigger" {
					t := oc.now.Add(time.Duration(o.DurationMinutes) * time.Minute)
					expires = &t
				}
				if _, err := oc.insertOffer(tx, o, nil, 0, 0, expires, false); err != nil {
					return err
				}
				e.created = true
				e.hist.Open = true
				e.hist.Count++
				open++
			}
			out = append(out, e)
		}
		return nil
	}
	if !create {
		return out, run(oc.s.db, nil)
	}
	err = oc.s.inTx(oc.ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(oc.ctx, `INSERT INTO economy_user_offer_state (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, oc.userID); err != nil {
			return err
		}
		var dummy string
		if err := tx.QueryRowContext(oc.ctx, `SELECT user_id::text FROM economy_user_offer_state WHERE user_id=$1 FOR UPDATE`, oc.userID).Scan(&dummy); err != nil {
			return err
		}
		return run(tx, tx)
	})
	return out, err
}

func (oc *offerCtx) startEntryTimers(ids []string) error {
	q := `UPDATE economy_user_offer u SET shown_at=COALESCE(u.shown_at, $2), expires_at=COALESCE(u.expires_at, $2::timestamptz + o.duration_minutes * INTERVAL '1 minute'), updated_at=now()
		FROM economy_offer o WHERE o.id=u.offer_id AND u.user_id=$1 AND u.kind='entry' AND u.status='active' AND (u.shown_at IS NULL OR u.expires_at IS NULL)`
	args := []any{oc.userID, oc.now}
	if ids != nil {
		q += ` AND u.id::text = ANY($3)`
		args = append(args, ids)
	}
	_, err := oc.s.db.ExecContext(oc.ctx, q, args...)
	return err
}

type userOffer struct {
	ID              string            `json:"id"`
	UserID          string            `json:"userId"`
	OfferID         string            `json:"offerId"`
	Kind            string            `json:"kind"`
	PeriodStart     *time.Time        `json:"periodStart"`
	Slot            int               `json:"slot"`
	Status          string            `json:"status"`
	OriginalToman   *int64            `json:"originalToman"`
	FinalToman      *int64            `json:"finalToman"`
	CoinCurrency    *string           `json:"coinCurrency"`
	CoinAmount      *int64            `json:"coinAmount"`
	CoinOriginal    *int64            `json:"coinOriginal"`
	DiscountPercent int               `json:"discountPercent"`
	SKU             map[string]string `json:"sku"`
	PurchaseLimit   int               `json:"purchaseLimit"`
	Bought          int               `json:"bought"`
	Reroll          int               `json:"reroll"`
	Forgotten       bool              `json:"forgotten"`
	TriggeredAt     time.Time         `json:"triggeredAt"`
	ShownAt         *time.Time        `json:"shownAt"`
	ExpiresAt       *time.Time        `json:"expiresAt"`
	LastBoughtAt    *time.Time        `json:"lastBoughtAt"`
	CreatedAt       time.Time         `json:"createdAt"`
	ProductName     i18n              `json:"productName"`
	Tpl             offerTemplate     `json:"offer"`
}

const userOfferJSON = `jsonb_build_object('id',u.id,'userId',u.user_id,'offerId',u.offer_id,'kind',u.kind,'periodStart',u.period_start,'slot',u.slot,'status',u.status,
	'originalToman',u.original_toman,'finalToman',u.final_toman,'coinCurrency',u.coin_currency,'coinAmount',u.coin_amount,'coinOriginal',u.coin_original,'discountPercent',u.discount_percent,
	'sku',u.sku,'purchaseLimit',u.purchase_limit,'bought',u.bought,'reroll',u.reroll,'forgotten',u.forgotten,'triggeredAt',u.triggered_at,'shownAt',u.shown_at,'expiresAt',u.expires_at,
	'lastBoughtAt',u.last_bought_at,'createdAt',u.created_at,'productName',(SELECT p.name FROM economy_product p WHERE p.id=t.product_id),'offer',` + offerJSON + `)`

func loadUserOffers(ctx context.Context, q rowsQueryer, where string, args ...any) ([]userOffer, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+userOfferJSON+`::text FROM economy_user_offer u JOIN economy_offer t ON t.id=u.offer_id `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []userOffer{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var u userOffer
		if err := json.Unmarshal([]byte(raw), &u); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (u userOffer) price() offerPrice {
	p := offerPrice{OriginalToman: u.OriginalToman, FinalToman: u.FinalToman, CoinAmount: u.CoinAmount, CoinOriginal: u.CoinOriginal, DiscountPercent: u.DiscountPercent, SKU: u.SKU}
	if u.CoinCurrency != nil {
		p.CoinCurrency = *u.CoinCurrency
	}
	if p.SKU == nil {
		p.SKU = map[string]string{}
	}
	return p
}

func (u userOffer) soldOut() bool { return u.Bought >= u.PurchaseLimit }

func (u userOffer) expired(now time.Time) bool { return u.ExpiresAt != nil && !u.ExpiresAt.After(now) }

func (p productView) withOffer(u userOffer) productView {
	price := u.price()
	p.PriceToman = price.OriginalToman
	p.DiscountPercent = price.DiscountPercent
	if price.CoinCurrency != "" && price.CoinAmount != nil {
		p.CoinPrice = &coinPrice{CurrencyKey: price.CoinCurrency, Amount: *price.CoinAmount}
	} else {
		p.CoinPrice = nil
	}
	p.SKU = price.SKU
	p.offer = &offerRef{ID: u.ID, Final: price.FinalToman}
	return p
}

type offerRef struct {
	ID    string
	Final *int64
}

func (p productView) offerID() any {
	if p.offer == nil {
		return nil
	}
	return p.offer.ID
}

func sortOffersBySlot(list []userOffer) {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Slot != list[j].Slot {
			return list[i].Slot < list[j].Slot
		}
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})
}

func (u userOffer) displayName() i18n {
	if len(u.Tpl.Name) > 0 {
		return u.Tpl.Name
	}
	return u.ProductName
}
