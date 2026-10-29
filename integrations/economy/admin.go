package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
)

const consolePrefix = "/v2/console/economy"

func (s *service) registerConsoleHTTP(initializer runtime.Initializer) error {
	type route struct {
		path, method string
		handler      http.HandlerFunc
	}
	routes := []route{
		{consolePrefix + "/summary", http.MethodGet, s.consoleSummary},
		{consolePrefix + "/payment-methods", http.MethodGet, s.consoleMethodsList},
		{consolePrefix + "/payment-methods/{method}", http.MethodPut, s.consoleMethodUpdate},
		{consolePrefix + "/payments", http.MethodGet, s.consolePaymentsList},
		{consolePrefix + "/payments/{id}", http.MethodGet, s.consolePaymentGet},
		{consolePrefix + "/payments/{id}/inquire", http.MethodPost, s.consolePaymentInquire},
		{consolePrefix + "/payments/{id}/retry-grant", http.MethodPost, s.consolePaymentRetryGrant},
		{consolePrefix + "/users/{userId}", http.MethodGet, s.consoleUserGet},
		{consolePrefix + "/users/{userId}/grant", http.MethodPost, s.consoleUserGrant},
		{consolePrefix + "/uploads", http.MethodPost, s.consoleUpload},
	}
	for _, res := range resources() {
		base := consolePrefix + "/" + res.name
		routes = append(routes,
			route{base, http.MethodGet, s.resourceList(res)},
			route{base, http.MethodPost, s.resourceWrite(res, true)},
			route{base + "/{id}", http.MethodGet, s.resourceGet(res)},
			route{base + "/{id}", http.MethodPut, s.resourceWrite(res, false)},
			route{base + "/{id}", http.MethodDelete, s.resourceDelete(res)},
		)
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) registerPublicHTTP(initializer runtime.Initializer) error {
	if err := initializer.RegisterHttp("/v2/economy/payments/callback/{ticket}", s.callbackHTTP, http.MethodGet, http.MethodPost); err != nil {
		return err
	}
	return initializer.RegisterHttp("/v2/economy/files/{name}", s.fileHTTP, http.MethodGet, http.MethodHead)
}

const paymentAdminJSON = `jsonb_build_object('id',p.id,'userId',p.user_id,'username',u.username,'productId',p.product_id,'productKey',pr.key,'productName',pr.name,
	'method',p.method,'sandbox',p.sandbox,'status',p.status,'originalToman',p.original_toman,'finalToman',p.final_toman,'amountRial',p.amount_rial,
	'coinCurrency',p.coin_currency,'coinAmount',p.coin_amount,'giftId',p.gift_id,'discountPercent',p.discount_percent,'authority',p.authority,'trackId',p.track_id,
	'refId',p.ref_id,'storeToken',p.store_token,'cardMask',p.card_mask,'outcome',p.outcome,'error',p.error,'rewardGranted',p.reward_granted,'rewardError',p.reward_error,
	'inquiryCount',p.inquiry_count,'inquiredAt',p.inquired_at,'verifiedAt',p.verified_at,'createdAt',p.created_at,'updatedAt',p.updated_at)`

const paymentAdminFrom = ` FROM economy_payment p LEFT JOIN users u ON u.id=p.user_id LEFT JOIN economy_product pr ON pr.id=p.product_id`

func (s *service) consolePaymentsList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	q := r.URL.Query()
	where := &sqlWhere{}
	if v := q.Get("status"); v != "" {
		where.add("p.status = ?", v)
	}
	if v := q.Get("method"); v != "" {
		where.add("p.method = ?", v)
	}
	if v := q.Get("userId"); v != "" {
		if _, err := uuid.FromString(v); err != nil {
			s.writeError(w, badRequest("INVALID_ID", "userId", "must be a UUID"))
			return
		}
		where.add("p.user_id = ?::uuid", v)
	}
	if v := q.Get("productId"); v != "" {
		if _, err := uuid.FromString(v); err != nil {
			s.writeError(w, badRequest("INVALID_ID", "productId", "must be a UUID"))
			return
		}
		where.add("p.product_id = ?::uuid", v)
	}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(p.id::text ILIKE "+n+" OR COALESCE(p.authority,'') ILIKE "+n+" OR COALESCE(p.track_id,'') ILIKE "+n+" OR COALESCE(p.ref_id,'') ILIKE "+n+
			" OR COALESCE(p.store_token,'') ILIKE "+n+" OR COALESCE(u.username,'') ILIKE "+n+" OR COALESCE(pr.key,'') ILIKE "+n+")")
	}
	var total int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*)`+paymentAdminFrom+where.sql(), where.args...).Scan(&total); err != nil {
		s.writeError(w, err)
		return
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+paymentAdminJSON+`::text`+paymentAdminFrom+where.sql()+` ORDER BY p.created_at DESC, p.id DESC LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2),
		append(append([]any{}, where.args...), p.Limit, p.Offset)...)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	items := []json.RawMessage{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			s.writeError(w, err)
			return
		}
		items = append(items, json.RawMessage(raw))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": p.Page, "limit": p.Limit})
}

func (s *service) paymentDetail(r *http.Request, id string) (map[string]any, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("PAYMENT_NOT_FOUND")
	}
	var raw string
	err := s.db.QueryRowContext(r.Context(), `SELECT `+paymentAdminJSON+`::text`+paymentAdminFrom+` WHERE p.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("PAYMENT_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT jsonb_build_object('id',id,'action',action,'url',url,'request',request,'response',response,'httpStatus',http_status,'durationMs',duration_ms,'createdAt',created_at)::text
		FROM economy_payment_log WHERE payment_id=$1 ORDER BY created_at, id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logs := []json.RawMessage{}
	for rows.Next() {
		var l string
		if err := rows.Scan(&l); err != nil {
			return nil, err
		}
		logs = append(logs, json.RawMessage(l))
	}
	out["logs"] = logs
	productID, _ := out["productId"].(string)
	if product, err := s.fetchResource(r.Context(), s.db, resourceByName("products"), productID); err == nil {
		out["product"] = product
	}
	userID, _ := out["userId"].(string)
	var user string
	if err := s.db.QueryRowContext(r.Context(), `SELECT jsonb_build_object('id',id,'username',username,'displayName',display_name)::text FROM users WHERE id=$1`, userID).Scan(&user); err == nil {
		out["user"] = json.RawMessage(user)
	} else {
		out["user"] = nil
	}
	return out, nil
}

func resourceByName(name string) resource {
	for _, r := range resources() {
		if r.name == name {
			return r
		}
	}
	return resource{}
}

func (s *service) consolePaymentGet(w http.ResponseWriter, r *http.Request) {
	out, err := s.paymentDetail(r, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) consolePaymentInquire(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	p, err := s.loadPayment(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if p.Status != "pending" {
		s.writeError(w, conflict("PAYMENT_NOT_PENDING", "only pending payments can be checked"))
		return
	}
	if _, err := s.reconcileOne(r.Context(), id); err != nil {
		s.writeError(w, err)
		return
	}
	s.consolePaymentGet(w, r)
}

func (s *service) consolePaymentRetryGrant(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	p, err := s.loadPayment(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if p.RewardGranted {
		s.writeError(w, conflict("REWARD_ALREADY_GRANTED", "the reward was already granted"))
		return
	}
	if p.Status != "success" {
		s.writeError(w, conflict("PAYMENT_NOT_SUCCESSFUL", "only successful payments can be granted"))
		return
	}
	if _, err := s.grantPayment(r.Context(), id); err != nil {
		s.writeError(w, err)
		return
	}
	s.consolePaymentGet(w, r)
}

func (s *service) consoleUserGet(w http.ResponseWriter, r *http.Request) {
	userID := pathVar(r, "userId")
	if _, err := uuid.FromString(userID); err != nil {
		s.writeError(w, notFound("USER_NOT_FOUND"))
		return
	}
	var username string
	if err := s.db.QueryRowContext(r.Context(), `SELECT username FROM users WHERE id=$1`, userID).Scan(&username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.writeError(w, notFound("USER_NOT_FOUND"))
			return
		}
		s.writeError(w, err)
		return
	}
	wallet, err := s.userWallet(r.Context(), userID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	metas, err := s.userMetas(r.Context(), userID, true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT jsonb_build_object('id',up.id,'prizeId',up.prize_id,'prizeKey',z.key,'prizeName',z.name,'sourceType',up.source_type,'sourceId',up.source_id,
		'claimed',up.claimed,'error',up.error,'rewards',up.rewards,'createdAt',up.created_at)::text
		FROM economy_user_prize up JOIN economy_prize z ON z.id=up.prize_id WHERE up.user_id=$1 ORDER BY up.created_at DESC LIMIT 50`, userID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	prizes := []json.RawMessage{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			s.writeError(w, err)
			return
		}
		prizes = append(prizes, json.RawMessage(raw))
	}
	writeJSON(w, http.StatusOK, map[string]any{"userId": userID, "username": username, "wallet": wallet, "metas": metas, "prizes": prizes})
}

func (s *service) consoleUserGrant(w http.ResponseWriter, r *http.Request) {
	userID := pathVar(r, "userId")
	if _, err := uuid.FromString(userID); err != nil {
		s.writeError(w, notFound("USER_NOT_FOUND"))
		return
	}
	var in struct {
		PrizeID string `json:"prizeId"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	prizeID, err := cleanUUID("prizeId", in.PrizeID, true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var exists bool
	if err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1)`, userID).Scan(&exists); err != nil || !exists {
		s.writeError(w, notFound("USER_NOT_FOUND"))
		return
	}
	rewards, err := s.grant(r.Context(), grantRequest{UserID: userID, PrizeID: *prizeID, SourceType: "admin", SourceID: uuid.Must(uuid.NewV4()).String(), Reason: "admin_grant"})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rewards": rewards})
}

func (s *service) consoleSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	counts := map[string]int64{}
	for name, query := range map[string]string{
		"currencies":     `SELECT count(*) FROM economy_currency`,
		"metas":          `SELECT count(*) FROM economy_meta`,
		"prizes":         `SELECT count(*) FROM economy_prize`,
		"products":       `SELECT count(*) FROM economy_product`,
		"activeProducts": `SELECT count(*) FROM economy_product WHERE is_active`,
		"gifts":          `SELECT count(*) FROM economy_gift`,
		"activeGifts":    `SELECT count(*) FROM economy_gift WHERE is_active`,
		"categories":     `SELECT count(*) FROM economy_category`,
		"tags":           `SELECT count(*) FROM economy_tag`,
		"grantErrors":    `SELECT count(*) FROM economy_payment WHERE status='success' AND NOT reward_granted`,
		"paymentsToday":  `SELECT count(*) FROM economy_payment WHERE created_at >= now() - interval '24 hours'`,
		"revenueToman30": `SELECT COALESCE(sum(final_toman),0) FROM economy_payment WHERE status='success' AND NOT sandbox AND method IN ('zarinpal','zibal','bazaar','myket') AND created_at >= now() - interval '30 days'`,
	} {
		var n int64
		if err := s.db.QueryRowContext(ctx, query).Scan(&n); err != nil {
			s.writeError(w, err)
			return
		}
		counts[name] = n
	}
	byStatus := map[string]int64{"pending": 0, "success": 0, "failed": 0, "expired": 0}
	rows, err := s.db.QueryContext(ctx, `SELECT status, count(*) FROM economy_payment GROUP BY status`)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var st string
		var n int64
		if err := rows.Scan(&st, &n); err != nil {
			s.writeError(w, err)
			return
		}
		byStatus[st] = n
	}
	methods := []map[string]any{}
	for _, m := range allMethods {
		pm, err := s.loadMethod(ctx, s.db, m)
		if err != nil {
			s.writeError(w, err)
			return
		}
		methods = append(methods, map[string]any{"method": m, "isActive": pm.Active, "configured": methodConfigured(m, pm.Config), "sandbox": pm.Config.flag("sandbox")})
	}
	counts["payments"] = byStatus["pending"] + byStatus["success"] + byStatus["failed"] + byStatus["expired"]
	writeJSON(w, http.StatusOK, map[string]any{"counts": counts, "paymentsByStatus": byStatus, "paymentMethods": methods})
}
