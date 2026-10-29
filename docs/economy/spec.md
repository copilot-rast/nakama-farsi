# Economy module (phase 1)

Port of the asp `Modules/Economy` (Currency, Meta, Prize, Shop, Payment, Gift) plus the Base pieces they need (Category, Tag, uploads), as a Go runtime plugin with admin pages inside the existing Nakama console. Gamification is out of scope. Events, EventPrizeRelation, MetaRelation, Streak and Achievement are phase 2.

References:
- asp: `/Users/mazyar/Projects/asp/backend/src/PolicePatrol.Api/Modules/Economy` (ZarinPal, Bazaar, Myket, shop, prize, gift, admin UI in `/Users/mazyar/Projects/asp/frontend/src/features/{currencies,metas,prizes,shops,gifts,paymentMethods,payments,categories,tags}`)
- afz: `/Users/mazyar/Projects/afz/new-codebase/backend/src/Academia.Api/Modules/Commerce/Payment` (Zibal, signed callback tickets, reconciliation)

## Layout

| Part | Path |
|---|---|
| Go plugin (package main, `InitModule`) | `integrations/economy/` → `economy.so` |
| Migrations (embedded in Nakama's migrator) | `migrate/sql/2026103000xxxx-economy-*.sql` |
| Console auth for `/v2/console/economy/` | `server/console.go`, `console/acl/acl.go` |
| Console pages (overlay, like the l10n layer) | `console/ui/economy/` → injected into `console/ui/dist/index.html` |

Rules: SQL must run on PostgreSQL 16 (CockroachDB is legacy, but avoid Postgres-only features where a portable form exists; no advisory locks, use `SELECT … FOR UPDATE`). Up sections re-runnable (`IF NOT EXISTS`), Down raises like the Jazebe migrations. No code comments. Every money/grant path idempotent.

## Runtime env (`--runtime.env KEY=VALUE`)

| Key | Meaning | Default |
|---|---|---|
| `ECONOMY_SECRET_KEY` | 32+ chars; AES-GCM key for gateway secrets and HMAC for callback tickets (derive two keys with HKDF) | required, module refuses to start without it |
| `ECONOMY_PUBLIC_BASE_URL` | Public base of the API port, used for callback and file URLs | required |
| `ECONOMY_DEEP_LINK` | Redirect after a gateway payment; `{success}` and `{paymentId}` are replaced | `nakama://payment?success={success}&paymentId={paymentId}` |
| `ECONOMY_WEB_REDIRECT_URL` | If set, used instead of the deep link (same placeholders) | empty |
| `ECONOMY_PENDING_TTL_MINUTES` | Pending payment expiry | 45 |
| `ECONOMY_UPLOAD_DIR` | Where admin uploads are written | `<data_dir>/economy-uploads` |

## Data model (tables prefixed `economy_`)

Names and descriptions are `jsonb` `{"fa": "...", "en": "..."}`; `fa` required.

- `category(id uuid, kind text check in ('product','prize','gift','currency'), key text, name jsonb, sort int, created_at, updated_at)`, unique `(kind,key)`.
- `tag(id, key unique, name jsonb, created_at, updated_at)`.
- `currency(id, key unique [a-z0-9_]{2,32}, name jsonb, type text check in ('soft','hard'), default_value bigint >=0, icon_url, metadata jsonb, sort, created_at, updated_at)`. A user's balance is the Nakama wallet entry named by `key`. Default values are granted once per user per currency (`user_currency_init(user_id, currency_id, primary key both)`), lazily on the first client economy RPC.
- `meta(id, key unique, name jsonb, default_value numeric(20,4), server_only bool, settings jsonb, created_at, updated_at)`; `user_meta(user_id, meta_id, value numeric(20,4), updated_at, primary key(user_id, meta_id))`. Missing row means `default_value`. `server_only` metas are hidden from clients.
- `prize(id, key unique, name jsonb, message jsonb, category_id null, tag_ids text[], is_repeatable bool, items jsonb, metadata jsonb, created_at, updated_at)`. `items` is an array of `{type: 'currency'|'meta'|'storage', key, action: 'add'|'set'|'give', value}`:
  - `currency` + `add` (value integer, may be negative only if balance allows) via Nakama `WalletUpdate` with ledger metadata `{module:"economy", reason, sourceType, sourceId}`.
  - `meta` + `add`|`set` (numeric).
  - `storage` + `give`: writes `value` (object) to collection `economy_items`, key = item key, merging is not done (overwrite).
- `user_prize(id, user_id, prize_id, source_type text check in ('payment','gift','admin'), source_id text, claimed bool, error text, created_at)`, unique `(source_type, source_id, prize_id)`. Granting is all-or-nothing in one transaction (wallet update through `nk.WalletUpdate` happens after the DB part commits only if you cannot do both atomically; prefer `nk.MultiUpdate` for wallet + storage, then mark claimed; on failure record `error` so admin can retry). A non-repeatable prize is granted at most once per user.
- `product(id, key unique, name jsonb, description jsonb, is_active bool, category_id null, tag_ids text[], price_toman bigint null, coin_price jsonb null {currency_key, amount}, discount_percent int 0..100, payment_methods text[] subset of ('zarinpal','zibal','bazaar','myket','in_app'), sku jsonb {bazaar?, myket?}, is_subscription bool, prize_id not null, icon_url, cover_url, min_level int 0, metadata jsonb, sort int, created_at, updated_at)`. Validation: gateway methods need `price_toman >= 100`; `in_app` needs `coin_price`; store methods need the matching `sku`.
- `gift(id, name text, type text check in ('reward','discount'), code text unique null (stored upper-case), prize_id null, discount_percent null 1..100, is_promo bool, per_user_limit int default 1, max_total_uses int null, total_used int, valid_from, expire_at, is_active, created_at, updated_at)`. `reward`: redeem grants `prize_id`. `discount`: usable as `gift_code` on a gateway purchase only.
- `gift_use(id, gift_id, user_id, payment_id null, status text check in ('reserved','used','released'), created_at, updated_at)`. A discount is reserved when a payment starts and becomes `used` on success or `released` on fail/expiry; limits count reserved + used.
- `payment_method(method pk, is_active bool, config text /* AES-GCM ciphertext of a JSON object */, updated_at)`.
- `payment(id uuid, user_id, product_id, method, sandbox bool, status text check in ('pending','success','failed','expired'), original_toman bigint, final_toman bigint, amount_rial bigint, coin_currency text null, coin_amount bigint null, gift_id null, discount_percent int, authority text null, track_id text null, ref_id text null, store_token text null, card_mask text null, outcome text null, error text null, reward_granted bool, reward_error text null, inquiry_count int, inquired_at, verified_at, created_at, updated_at)`. Unique `(method, authority)`, `(method, track_id)`, `(method, store_token)` where not null. Indexes `(status, created_at)`, `(user_id, created_at)`.
- `payment_log(id, payment_id, action text check in ('request','callback','verify','inquiry','store_validate'), url, request jsonb, response jsonb, http_status int, duration_ms int, created_at)`. Secrets masked as `********`.

## Payment methods

Config fields per method (secret fields are returned as `********`; sending `********` back keeps the stored value; `""` clears):

| method | fields |
|---|---|
| zarinpal | `merchantId` (secret, required unless sandbox), `sandbox` bool |
| zibal | `merchant` (secret, required unless sandbox → literal `zibal`), `sandbox` bool |
| bazaar | `packageName`, `apiSecret` (secret), `rsaPublicKey`, `sandbox` bool |
| myket | `packageName`, `accessToken` (secret), `rsaPublicKey`, `sandbox` bool |
| in_app | none; active flag only |

Amounts: prices are Toman; gateways get Rial = Toman × 10.

- **ZarinPal** (asp `Gateways/ZarinpalGateway.cs`): `POST https://api.zarinpal.com/pg/v4/payment/request.json` (sandbox host `sandbox.zarinpal.com`) `{merchant_id, amount (Rial), currency:"IRR", callback_url, description}`; success `data.code==100` + `data.authority`; redirect `https://www.zarinpal.com/pg/StartPay/{authority}` (sandbox `https://sandbox.zarinpal.com/pg/StartPay/{authority}`). Verify `POST …/verify.json {merchant_id, amount, authority}`; `100` or `101` = paid; read `ref_id`, `card_pan`.
- **Zibal** (afz `Gateways/ZibalGateway.cs`): `POST https://gateway.zibal.ir/v1/request {merchant, amount, callbackUrl, description, orderId}` → `result==100` + `trackId`; redirect `https://gateway.zibal.ir/start/{trackId}`. Verify `/v1/verify {merchant, trackId}`: `100`/`201` paid, `202` failed, else unknown. Inquiry `/v1/inquiry`: `(100,1)` paid, `(100,2)` paid-not-verified, `(100,-1)` pending, `(100, 3..12|15|16|18)` failed.
- **Bazaar** (asp `Gateways/StoreBillingValidator.cs`): `GET https://pardakht.cafebazaar.ir/devapi/v2/api/validate/{package}/inapp/{sku}/purchases/{token}/`, header `CAFEBAZAAR-PISHKHAN-API-SECRET`. `purchaseState==0` valid, `developerPayload` must equal the payment id. Subscriptions: `GET …/api/applications/{package}/subscriptions/{sku}/purchases/{token}`, active while `validUntilTimestampMsec` is in the future; transaction key `"{linkedSubscriptionToken ?? token}#{validUntilMs}"`.
- **Myket**: `POST https://developer.myket.ir/api/partners/applications/{package}/purchases/products/{sku}/verify` body `{"tokenId": token}`, header `X-Access-Token`. Same `purchaseState`/`developerPayload` rules. No subscriptions.
- Store reply classes (asp `StoreReplyClassifier`): 2xx accepted; 401/403 or 404 with `error=="invalid_value"` → configuration rejected (payment stays pending, error `STORE_CONFIGURATION_REJECTED`); 400/404/410 rejected → failed; other/network → `STORE_UNAVAILABLE`, payment stays pending. Escape every path segment.
- **Store sandbox**: no network call; token `sandbox-<paymentId>-<anything>` is valid with payload = paymentId.
- **Gateway sandbox**: ZarinPal sandbox host; Zibal merchant `zibal`.

Flow and safety (union of asp + afz):
1. Start: validate product active + method allowed + method active/configured; apply `gift_code` discount (reserve); `final_toman` 0 settles immediately; insert `pending` payment; for gateways call request and store authority/trackId; failure → `failed` + error.
2. Callback: `GET|POST /v2/economy/payments/callback/{ticket}` on the API port (`RegisterHttp`, no session). Ticket = base64url(`paymentId|expiresUnix`) + `.` + HMAC-SHA256, constant-time compare, 48h expiry. Payment id only from the ticket. The authority/trackId in the query must match the stored one or it is ignored (no state change). Then verify server-side and redirect (302, `Cache-Control: no-store`) to the deep link / web URL.
3. Verify: not `pending` → no-op. Amount from the gateway must equal `amount_rial` (missing amount → treat as unknown, stay pending). Paid → transition `pending→success` with a conditional update; the winner grants the prize (`source_type=payment`, `source_id=payment.id`) and marks the gift use `used`. Grant failure keeps `success` with `reward_error`; admin can retry.
4. Reconciler (`nk` background goroutine started in `InitModule`, every 60 s, batch 20): pending gateway payments older than 3 min → Zibal inquiry / ZarinPal verify; older than `ECONOMY_PENDING_TTL_MINUTES` → `expired`, release gift reservation.
5. Stores: start returns `{paymentId, sku, developerPayload: paymentId}`; submit validates token, binds payload, settles with `store_token` unique (duplicate token → `TOKEN_ALREADY_USED`). Review/restore endpoint for lost submits.
6. `in_app`: debit `coin_price` from the wallet and grant in one call path; insufficient → `INSUFFICIENT_BALANCE`, nothing changes.

## Client RPCs (game, user session)

All take/return JSON. Errors use Nakama runtime errors with a stable `message` code (UPPER_SNAKE).

| RPC | Input | Output |
|---|---|---|
| `economy_catalog` | `{categoryKey?, lang?}` | `{currencies:[{key,name,type,iconUrl}], categories:[…], products:[{id,key,name,description,iconUrl,coverUrl,priceToman,finalToman,coinPrice,discountPercent,paymentMethods,sku,isSubscription,categoryKey,tags}]}` (active only, methods filtered to active methods) |
| `economy_balances` | `{}` | `{wallet:{key:amount}, metas:{key:value} (non server_only), items:[…storage economy_items]}` |
| `economy_purchase_start` | `{productId, method, giftCode?}` | gateway: `{paymentId, status:"redirect", redirectUrl}`; store: `{paymentId, status:"store", sku, developerPayload}`; in_app or free: `{paymentId, status:"success", rewards}` |
| `economy_purchase_submit` | `{paymentId, token}` | `{paymentId, status, rewards?}` |
| `economy_purchase_review` | `{method, sku, token}` | same as submit |
| `economy_purchase_status` | `{paymentId}` | `{paymentId, status, rewardGranted, error?}` (own payments only) |
| `economy_store_config` | `{method}` | `{method, sandbox, rsaPublicKey, packageName}` |
| `economy_gift_redeem` | `{code}` | `{rewards}` |
| `economy_payments_mine` | `{limit?, cursor?}` | `{items, cursor}` |

`rewards` = the prize items actually applied.

## Console API (console port, console session token, `/v2/console/economy/`)

Wired through `customHttpAuthFunc` + audit log + mux params (extend the `/v2/console/hiro/` branch in `server/console.go` to also cover `/v2/console/economy/`). ACL in `console/acl/acl.go`: GET needs `HIRO_ECONOMY` read, other methods need `HIRO_ECONOMY` write (admins pass). Each handler registered with exactly one method so the ACL/audit gets it.

Generic CRUD for `categories, tags, currencies, metas, prizes, products, gifts`:
- `GET /v2/console/economy/{resource}?q=&page=1&limit=20[&kind=][&categoryId=][&isActive=]` → `{items, total, page, limit}` (server-side search on key/name, newest first)
- `GET /v2/console/economy/{resource}/{id}` → item
- `POST /v2/console/economy/{resource}` → created item
- `PUT /v2/console/economy/{resource}/{id}` → updated item (full replace of editable fields)
- `DELETE /v2/console/economy/{resource}/{id}` → `{}`; refuse with 409 `IN_USE` when referenced (prize used by product/gift, category used, currency used in coin_price or prize items).

JSON fields are camelCase mirrors of the columns (`name:{fa,en}`, `tagIds`, `priceToman`, `coinPrice:{currencyKey,amount}`, `paymentMethods`, `sku:{bazaar,myket}`, `prizeId`, `items:[…]`, …). Validation errors: 400 `{"error": CODE, "message": text, "field": name}`.

Other:
- `GET /v2/console/economy/payment-methods` → `[{method, isActive, configured, fields:[{name, type:'text'|'secret'|'bool'|'textarea', required}], config:{…masked}}]`
- `PUT /v2/console/economy/payment-methods/{method}` `{isActive, config}` → same item; activating without required fields → 400 `PAYMENT_METHOD_NOT_CONFIGURED`.
- `GET /v2/console/economy/payments?status=&method=&userId=&q=&page=&limit=` → list with `username`.
- `GET /v2/console/economy/payments/{id}` → payment + `logs` + product + user.
- `POST /v2/console/economy/payments/{id}/inquire` → re-check a pending gateway payment now.
- `POST /v2/console/economy/payments/{id}/retry-grant` → retry a failed grant.
- `GET /v2/console/economy/users/{userId}` → `{wallet, metas, prizes (latest 50)}`.
- `POST /v2/console/economy/users/{userId}/grant` `{prizeId}` → `{rewards}` (`source_type=admin`, source id = new uuid).
- `POST /v2/console/economy/uploads` multipart `file` (png/jpg/webp/svg, ≤ 2 MB, sniffed) → `{url}` served publicly at `GET /v2/economy/files/{name}` on the API port.
- `GET /v2/console/economy/summary` → counts for the overview page.

## Console pages

Sidebar section «اقتصاد» / "Economy" after «منوی اصلی»: نمای کلی، ارزها، متاها، جایزه‌ها، فروشگاه، هدیه‌ها، پرداخت‌ها، روش‌های پرداخت، دسته‌بندی‌ها، تگ‌ها. Routes `#/economy/<page>[/<id>]`. Pages look like the console's own pages (same Tailwind classes, light/dark, fa/en, Persian digits in fa). Lists: server-side search + paging; on mobile tables become stacked cards. Forms in a side drawer or a detail page; prize item editor is a repeatable row editor; product form shows only the fields relevant to the chosen payment methods.
