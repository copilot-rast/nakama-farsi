# Economy module (phase 1)

Port of the asp `Modules/Economy` (Currency, Meta, Prize, Shop, Payment, Gift) plus uploads, as a Go runtime plugin with admin pages inside the existing Nakama console. Gamification is out of scope. Phase 4 (2026-11-11) adds player events, event prizes, meta rules with level tables and streaks (section «Events, levels and streaks» below); Achievement is not built.

References:
- asp: `/Users/mazyar/Projects/asp/backend/src/PolicePatrol.Api/Modules/Economy` (ZarinPal, Bazaar, Myket, shop, prize, gift, admin UI in `/Users/mazyar/Projects/asp/frontend/src/features/{currencies,metas,prizes,shops,gifts,paymentMethods,payments,categories,tags}`)
- afz: `/Users/mazyar/Projects/afz/new-codebase/backend/src/Academia.Api/Modules/Commerce/Payment` (Zibal, signed callback tickets, reconciliation)

## Layout

| Part | Path |
|---|---|
| Go plugin (package main, `InitModule`) | `integrations/economy/` → `economy.so` |
| Migrations (embedded in Nakama's migrator) | `migrate/sql/2026103000xxxx-economy-*.sql`, `20261101000002-economy-gift-name-localized.sql` |
| Languages, categories, tags | Base module (`docs/base/spec.md`, `integrations/base`); economy reads `base_language`, `base_category`, `base_tag` |
| Translated text helper | `internal/localized` (shared with publication and base) |
| Secret sealing (HKDF + AES-GCM) | `internal/secretbox` (shared with publication and `samples/stargame`) |
| Sample game and its seed data | `samples/stargame` (served at `/stargame/`, seeds its shop on start; see its README) |
| Console auth for `/v2/console/economy/` | `server/console.go`, `console/acl/acl.go` |
| Console pages (overlay, like the l10n layer) | `console/ui/economy/` → injected into `console/ui/dist/index.html` |

Rules: SQL must run on PostgreSQL 16 (CockroachDB is legacy, but avoid Postgres-only features where a portable form exists; no advisory locks, use `SELECT … FOR UPDATE`). Up sections re-runnable (`IF NOT EXISTS`), Down raises an error. No code comments. Every money/grant path idempotent.

## Runtime env (`--runtime.env KEY=VALUE`)

| Key | Meaning | Default |
|---|---|---|
| `ECONOMY_SECRET_KEY` | 32+ chars; AES-GCM key for gateway secrets and HMAC for callback tickets (derive two keys with HKDF, salt `nakama-economy`, info `economy-config-seal` / `economy-callback-ticket`; helpers in `internal/secretbox`, shared with publication and the stargame sample). Changing it makes stored secrets unreadable | required, module refuses to start without it; compose falls back to `NAKAMA_SESSION_ENCRYPTION_KEY` |
| `ECONOMY_PUBLIC_BASE_URL` | Public base of the API port, used for callback and file URLs | required |
| `ECONOMY_DEEP_LINK` | Redirect after a gateway payment; `{success}` and `{paymentId}` are replaced | `nakama://payment?success={success}&paymentId={paymentId}` |
| `ECONOMY_WEB_REDIRECT_URL` | If set, used instead of the deep link (same placeholders) | empty; compose sets `https://nakama.dastar.dev/stargame/?payment={paymentId}&success={success}` |
| `ECONOMY_LOCAL_STRIPE_URL`, `ECONOMY_LOCAL_GOOGLE_PLAY_URL`, `ECONOMY_LOCAL_APPLE_URL`, `ECONOMY_LOCAL_BAZAAR_URL`, `ECONOMY_LOCAL_MYKET_URL` | Local test only: send Stripe, Google Play, Apple, Cafe Bazaar and Myket calls to a fake store on this machine. Only localhost or loopback addresses are accepted; anything else stops the plugin | empty (the real stores) |
| `ECONOMY_PENDING_TTL_MINUTES` | Pending payment expiry | 45 |
| `ECONOMY_UPLOAD_DIR` | Where admin uploads are written with the `local` driver | `<data_dir>/economy-uploads` |
| `STORAGE_DRIVER` | `local` (files in `ECONOMY_UPLOAD_DIR`, one node only) or `s3` (S3-compatible bucket shared by all nodes; `internal/objectstore`, SigV4 with the standard library) | `local` |
| `STORAGE_S3_ENDPOINT` | `http(s)://host[:port]` without a path, e.g. MinIO `http://minio:9000` | required for `s3` |
| `STORAGE_S3_BUCKET`, `STORAGE_S3_ACCESS_KEY`, `STORAGE_S3_SECRET_KEY` | Bucket (must exist) and credentials | required for `s3` |
| `STORAGE_S3_REGION` | Signing region | `us-east-1` |
| `STORAGE_S3_PATH_STYLE` | `true`: `endpoint/bucket/key` (MinIO); `false`: `bucket.endpoint/key` | `true` |
| `STORAGE_S3_PREFIX` | Optional folder in the bucket; economy objects are `<prefix>/economy/<name>` | empty |
| `STORAGE_S3_PUBLIC_BASE_URL` | Empty: Nakama proxies `GET /v2/economy/files/{name}` from the bucket with the same headers as `local`. Set: that URL answers `302` to `<base>/<prefix>/economy/<name>` (the bucket or CDN must then serve safe headers for SVG itself) | empty |

The module refuses to start when the storage settings are invalid. File URLs stay `<ECONOMY_PUBLIC_BASE_URL>/v2/economy/files/<name>` with every driver. Moving from `local` to `s3`: copy `ECONOMY_UPLOAD_DIR/*` to `<prefix>/economy/` in the bucket first.

The payment reconciler runs on every node: each pass claims one pending payment at a time with `FOR UPDATE SKIP LOCKED` and stamps `inquired_at`, so two nodes never inquire or verify the same payment in the same minute; settlement stays guarded by the `status='pending'` transitions.

## Data model (tables prefixed `economy_`)

Names, descriptions and messages are `jsonb` maps keyed by language code (`{"fa": "...", "en": "...", "ar": "..."}`), validated by `internal/localized` against `base_language`: keys must be existing language codes (active or inactive), values are trimmed and empty values dropped, the default language value is required for required fields and whenever any value is given, max length per language. Errors: 400 `INVALID_LANGUAGE`, `DEFAULT_LANGUAGE_REQUIRED`, `TEXT_TOO_LONG` with `field`.

- Categories and tags live in the Base module: `base_category` (`key_type` = `product` | `prize` | `gift` | `currency`) and `base_tag` (`key_type` = `product` | `prize` | `gift` | `currency` | `general`). Products, prizes, gifts and currencies all carry `category_id` + `tag_ids` (gifts and currencies since `20261110050000-economy-gift-currency-taxonomy.sql`, which also widens the `base_tag` key-type CHECK). `category_id` must point at a category whose `key_type` equals the entity (`CATEGORY_KIND_MISMATCH`); `tag_ids` must be tags whose `key_type` is the entity or `general` (`TAG_NOT_FOUND`, `TAG_KIND_MISMATCH`). The old `economy_category` / `economy_tag` tables were copied (ids kept) and dropped by `20261101000001-base-move-categories.sql`.
- `currency(id, key unique [a-z0-9_]{2,32}, name jsonb, type text check in ('soft','hard'), default_value bigint >=0, icon_url, category_id null (key type `currency`), tag_ids text[] (`currency` or `general`), metadata jsonb, sort, created_at, updated_at)`. A user's balance is the Nakama wallet entry named by `key`. Default values are granted once per user per currency (`user_currency_init(user_id, currency_id, primary key both)`), lazily on the first client economy RPC.
- `meta(id, key unique, name jsonb, default_value numeric(20,4), server_only bool, settings jsonb, created_at, updated_at)`; `user_meta(user_id, meta_id, value numeric(20,4), updated_at, primary key(user_id, meta_id))`. Missing row means `default_value`. `server_only` metas are hidden from clients.
- `prize(id, key unique, name jsonb, message jsonb, category_id null, tag_ids text[], is_repeatable bool, items jsonb, metadata jsonb, created_at, updated_at)`. `items` is an array of `{type: 'currency'|'meta'|'storage', key, action: 'add'|'set'|'give', value}`:
  - `currency` + `add` (value integer, may be negative only if balance allows) via Nakama `WalletUpdate` with ledger metadata `{module:"economy", reason, sourceType, sourceId}`.
  - `meta` + `add`|`set` (numeric).
  - `storage` + `give`: writes `value` (object) to collection `economy_items`, key = item key, merging is not done (overwrite).
- `user_prize(id, user_id, prize_id, source_type text check in ('payment','gift','admin'), source_id text, claimed bool, error text, created_at)`, unique `(source_type, source_id, prize_id)`. Granting is all-or-nothing in one transaction (wallet update through `nk.WalletUpdate` happens after the DB part commits only if you cannot do both atomically; prefer `nk.MultiUpdate` for wallet + storage, then mark claimed; on failure record `error` so admin can retry). A non-repeatable prize is granted at most once per user.
- `product(id, key unique, name jsonb, description jsonb, is_active bool, category_id null, tag_ids text[], price_toman bigint null, coin_price jsonb null {currency_key, amount}, discount_percent int 0..100, price_minor bigint null, price_currency text null (ISO code, lower case; the Stripe price in the smallest unit), payment_methods text[] subset of ('zarinpal','zibal','stripe','bazaar','myket','google_play','apple','in_app'), sku jsonb {bazaar?, myket?, google_play?, apple?}, is_subscription bool, prize_id not null, icon_url, cover_url, min_level int 0, metadata jsonb, sort int, created_at, updated_at)`. Validation: gateway methods need `price_toman >= 100`; `stripe` needs `price_minor` (currency defaults to `usd`) and no subscription; `in_app` needs `coin_price`; store methods need the matching `sku`.
- `gift(id, name jsonb (translated; existing text names were migrated to `{default: text}`), type text check in ('reward','discount'), code text unique null (stored upper-case), prize_id null, discount_percent null 1..100, is_promo bool, per_user_limit int default 1, max_total_uses int null, total_used int, valid_from, expire_at, is_active, category_id null (key type `gift`), tag_ids text[] (`gift` or `general`), created_at, updated_at)`. `reward`: redeem grants `prize_id`. `discount`: usable as `gift_code` on a gateway purchase only.
- `gift_use(id, gift_id, user_id, payment_id null, status text check in ('reserved','used','released'), created_at, updated_at)`. A discount is reserved when a payment starts and becomes `used` on success or `released` on fail/expiry; limits count reserved + used.
- `payment_method(method pk, is_active bool, config text /* AES-GCM ciphertext of a JSON object */, updated_at)`.
- `payment(id uuid, user_id, product_id, method, sandbox bool, status text check in ('pending','success','failed','expired'), original_toman bigint, final_toman bigint, amount_rial bigint, coin_currency text null, coin_amount bigint null, gift_id null, discount_percent int, authority text null, track_id text null, ref_id text null, store_token text null, card_mask text null, outcome text null, error text null, reward_granted bool, reward_error text null, inquiry_count int, inquired_at, verified_at, created_at, updated_at)`. Unique `(method, authority)`, `(method, track_id)`, `(method, store_token)` where not null. Indexes `(status, created_at)`, `(user_id, created_at)`.
- `payment` also has `currency text null, amount_minor bigint null` (Stripe) and, for `google_play`/`apple`, a unique `(method, ref_id)` among successful payments (Google `orderId`, Apple `transactionId`).
- `payment_webhook(method, event_id, type, payment_id null, result, created_at)`, primary key `(method, event_id)`: processed Stripe events.
- `payment_log(id, payment_id, action text check in ('request','callback','verify','inquiry','store_validate','webhook'), url, request jsonb, response jsonb, http_status int, duration_ms int, created_at)`. Secrets masked as `********`.

## Payment methods

Config fields per method (secret fields are returned as `********`; sending `********` back keeps the stored value; `""` clears):

| method | fields |
|---|---|
| zarinpal | `merchantId` (secret, required unless sandbox), `sandbox` bool |
| zibal | `merchant` (secret, required unless sandbox → literal `zibal`), `sandbox` bool |
| bazaar | `packageName`, `apiSecret` (secret), `rsaPublicKey`, `sandbox` bool |
| myket | `packageName`, `accessToken` (secret), `rsaPublicKey`, `sandbox` bool |
| google_play | `packageName`, `serviceAccountJson` (secret, type `secretJson`), `sandbox` bool |
| apple | `bundleId`, `sharedSecret` (secret, legacy receipts), `issuerId`, `keyId`, `privateKeyP8` (secret, type `secretText`), `sandbox` bool; needs `sharedSecret` or all three API keys |
| stripe | `secretKey` (secret), `webhookSecret` (secret), `publishableKey`, `currency` (default `usd`), `testMode` bool |
| in_app | none; active flag only |

Amounts: prices are Toman; gateways get Rial = Toman × 10.

- **ZarinPal** (asp `Gateways/ZarinpalGateway.cs`): `POST https://api.zarinpal.com/pg/v4/payment/request.json` (sandbox host `sandbox.zarinpal.com`) `{merchant_id, amount (Rial), currency:"IRR", callback_url, description}`; success `data.code==100` + `data.authority`; redirect `https://www.zarinpal.com/pg/StartPay/{authority}` (sandbox `https://sandbox.zarinpal.com/pg/StartPay/{authority}`). Verify `POST …/verify.json {merchant_id, amount, authority}`; `100` or `101` = paid; read `ref_id`, `card_pan`.
- **Zibal** (afz `Gateways/ZibalGateway.cs`): `POST https://gateway.zibal.ir/v1/request {merchant, amount, callbackUrl, description, orderId}` → `result==100` + `trackId`; redirect `https://gateway.zibal.ir/start/{trackId}`. Verify `/v1/verify {merchant, trackId}`: `100`/`201` paid, `202` failed, else unknown. Inquiry `/v1/inquiry`: `(100,1)` paid, `(100,2)` paid-not-verified, `(100,-1)` pending, `(100, 3..12|15|16|18)` failed.
- **Bazaar** (asp `Gateways/StoreBillingValidator.cs`): `GET https://pardakht.cafebazaar.ir/devapi/v2/api/validate/{package}/inapp/{sku}/purchases/{token}/`, header `CAFEBAZAAR-PISHKHAN-API-SECRET`. `purchaseState==0` valid, `developerPayload` must equal the payment id. Subscriptions: `GET …/api/applications/{package}/subscriptions/{sku}/purchases/{token}`, active while `validUntilTimestampMsec` is in the future; transaction key `"{linkedSubscriptionToken ?? token}#{validUntilMs}"`.
- **Myket**: `POST https://developer.myket.ir/api/partners/applications/{package}/purchases/products/{sku}/verify` body `{"tokenId": token}`, header `X-Access-Token`. Same `purchaseState`/`developerPayload` rules. No subscriptions.
- Store reply classes (`classifyStore` in `gateways.go`): 2xx accepted; 401/403, or an `error` such as `invalid_credentials`/`access_forbidden` → configuration rejected (reason `credentials`); Bazaar 404 `not_found` ("The requested purchase is not found!") → rejected (`purchase_not_found`); Bazaar 404 `invalid_value` → configuration rejected when `error_description` names the package ("Package name is invalid", reason `package`) or the product ("Product is not found.", reason `product`), rejected when it names the token ("Invalid token!", reason `token_invalid`), configuration rejected (`unknown_value`) otherwise; Myket `messageCode` with `Sku`/`Product` (`SkuIdNotFound`) → configuration (`product`), `Package`/`Application` → configuration (`package`), other 400/404/410 → rejected; Google `applicationNotFound`/`projectNotLinked` → configuration (`package`), `productNotFound` → configuration (`product`), other 400/404/410 → rejected; Apple API 404 → rejected, 401 → configuration; 429, other statuses, network errors → `STORE_UNAVAILABLE`. Configuration rejected answers `STORE_CONFIGURATION_REJECTED` with status 412 (gRPC 9 FAILED_PRECONDITION, permanent until an admin fixes the method; the payment stays pending with that error); `METHOD_UNAVAILABLE` and `STORE_MODE_CHANGED` are 412 too; only `STORE_UNAVAILABLE` stays 503 (gRPC 14, retry). Escape every path segment.
- Store health (console): every live store call records the last failure (`economy_payment_method.store_last_failure`: time, HTTP status, class, reason, store `error`/`error_description` clipped to 300 chars, never the secret or the token) and the last good answer (`store_last_ok_at`). `GET /payment-methods` returns `health {lastFailure, lastOkAt, lastTest, stuckPayments}` for store methods. `POST /payment-methods/{method}/test {sku?}` calls the live store with a made-up token (sku = given, else the first product sku of the method, else a placeholder) and answers `result`: `credentials_ok` (the store answered "purchase not found"), `credentials_rejected`, `package_not_found`, `product_not_found`, `config_rejected`, `unreachable`, `not_configured`, `not_testable` (Apple without App Store Server API keys); stored in `store_last_test`. `POST /payments/{id}/recheck` sends the last token the game sent for a pending or expired store payment (from its `store_validate` log) through the normal settle path (answer = payment detail + `recheck {status, error?}`; `STORE_TOKEN_UNKNOWN` when no token was ever sent, `PAYMENT_NOT_PENDING` for settled ones). `POST /payment-methods/{method}/recheck` does the same for up to 100 pending or expired payments of the last 30 days with a logged token (`checked`, `delivered`, `stillPending`, `closed`, `items`). Migration `20270204000000-economy-store-checks.sql`.
- **Store sandbox**: no network call; token `sandbox-<paymentId>-<anything>` is valid with payload = paymentId.
- **Gateway sandbox**: ZarinPal sandbox host; Zibal merchant `zibal`.

Flow and safety (union of asp + afz):
1. Start: validate product active + method allowed + method active/configured; apply `gift_code` discount (reserve); `final_toman` 0 settles immediately; insert `pending` payment; for gateways call request and store authority/trackId; failure → `failed` + error.
2. Callback: `GET|POST /v2/economy/payments/callback/{ticket}` on the API port (`RegisterHttp`, no session). Ticket = base64url(`paymentId|expiresUnix`) + `.` + HMAC-SHA256, constant-time compare, 48h expiry. Payment id only from the ticket. The authority/trackId in the query must match the stored one or it is ignored (no state change). Then verify server-side and redirect (302, `Cache-Control: no-store`) to the deep link / web URL.
3. Verify: not `pending` → no-op. Amount from the gateway must equal `amount_rial` (missing amount → treat as unknown, stay pending). Paid → transition `pending→success` with a conditional update; the winner grants the prize (`source_type=payment`, `source_id=payment.id`) and marks the gift use `used`. Grant failure keeps `success` with `reward_error`; admin can retry.
4. Reconciler (`nk` background goroutine started in `InitModule`, every 60 s, batch 20): pending gateway payments older than 3 min → Zibal inquiry / ZarinPal verify; older than `ECONOMY_PENDING_TTL_MINUTES` → `expired`, release gift reservation.
5. Stores: start returns `{paymentId, sku, developerPayload: paymentId}`; submit validates token, binds payload, settles with `store_token` unique (duplicate token → `TOKEN_ALREADY_USED`). Review/restore endpoint for lost submits.
6. `in_app`: debit `coin_price` from the wallet and grant in one call path; insufficient → `INSUFFICIENT_BALANCE`, nothing changes.

## Google Play, Apple and Stripe (2026-11-18)

All three are OFF by default; nothing is seeded. Secrets are sealed with `ECONOMY_SECRET_KEY` like the other methods and read back as `********`. The method view also has `sandboxField` (`sandbox`, or `testMode` for Stripe) and `sandboxNeedsKeys` (true for the three: their test mode still calls the real store, so secrets stay required). Migration `20270103000003-economy-store-gateways.sql`.

Store methods `google_play` and `apple` use the bazaar/myket flow unchanged: `economy_purchase_start` (once_key, `PAYMENT_IN_PROGRESS`), `economy_purchase_submit {paymentId, token}`, `economy_purchase_review`, `economy_purchase_status`, the reward retry worker and grant on next call, `SANDBOX_OFF` / `PAYMENT_EXPIRED` / `STORE_MODE_CHANGED`. The payment id goes into the purchase: Google `obfuscatedProfileId` (or `obfuscatedAccountId`), Apple `appAccountToken` (the payment id is a UUID).

- **Google Play**: OAuth2 service-account JWT (scope `androidpublisher`, token cached per key). Products: `GET https://androidpublisher.googleapis.com/androidpublisher/v3/applications/{package}/purchases/products/{sku}/tokens/{token}`; `purchaseState` 0 valid, 1 → `PURCHASE_INVALID`, 2 → `PURCHASE_PENDING` (payment stays pending); `productId` must be the sku (`PRODUCT_MISMATCH`); `purchaseType==0` is a test purchase; `consumptionState==1` follows `storeConsumed`; `orderId` must be new (`TOKEN_ALREADY_USED`); after the grant an unacknowledged purchase is acknowledged (`…:acknowledge`) when `googleAcknowledge`. Subscriptions: `purchases/subscriptionsv2/tokens/{token}`, the line item with the sku, valid while ACTIVE / IN_GRACE_PERIOD / CANCELED and `expiryTime` is ahead; `testPurchase` marks a test; store key `{linkedPurchaseToken ?? token}#{expiryMs}`.
- **Apple**: token forms: digits = `transactionId` → App Store Server API `GET https://api.storekit.itunes.apple.com/inApps/v1/transactions/{id}` (sandbox host `api.storekit-sandbox.itunes.apple.com` when `sandbox`), ES256 JWT `{iss: issuerId, aud: appstoreconnect-v1, bid: bundleId}` with `kid: keyId`; a StoreKit 2 JWS → its `transactionId` is looked up the same way, or, without API keys, the JWS is verified locally; anything else is a base64 receipt → `verifyReceipt` with `sharedSecret` (production first, sandbox on 21007; 21004 → `STORE_CONFIGURATION_REJECTED`, 21005/21009/211xx → `STORE_UNAVAILABLE`, others → `PURCHASE_INVALID`). Every signed transaction (also the API answer) is verified: x5c chain to Apple Root CA G3 (the one in `iap.AppleRootPEM`), the App Store leaf (1.2.840.113635.100.6.11.1) and intermediate (1.2.840.113635.100.6.2.1) markers, ES256 signature. Checks: `bundleId` (`BUNDLE_MISMATCH`), `productId` = sku (`PRODUCT_MISMATCH`), `revocationDate` → `PURCHASE_INVALID`, `environment` vs the sandbox flag. Store key = `transactionId` (subscriptions `{originalTransactionId}#{expiresMs}`), so a transaction is used once (`TOKEN_ALREADY_USED`). Receipts may be up to 64 KB (other stores 1024).
- **Test purchases**: with `sandbox` on only test purchases (Google `purchaseType 0` / `testPurchase`, Apple `environment != Production`) are accepted, a live one is `STORE_ENVIRONMENT_MISMATCH`; with `sandbox` off a test purchase is `TEST_PURCHASE_REFUSED` unless `liveAcceptsTestPurchases`. These codes, `BUNDLE_MISMATCH`, `PRODUCT_MISMATCH` and `PURCHASE_CONSUMED` close the payment as failed. Not-found Google purchases are not acknowledged, so Google refunds them after three days.
- **Stripe** (`stripe`, a web gateway): start checks the keys match `testMode` (`sk_test_`/`rk_test_` in test mode, `sk_live_`/`rk_live_` live; `pk_` and `whsec_` prefixes on save, error `STRIPE_KEY_MODE`), needs a product price in the method currency, refuses offers, friend gifts and discount codes. `POST https://api.stripe.com/v1/checkout/sessions` (form) `mode=payment`, one line item with `price_data {currency, unit_amount = discounted(price_minor), product_data.name}`, `client_reference_id` + `metadata[paymentId]` + `payment_intent_data[metadata][paymentId]` = payment id, `success_url`/`cancel_url` = `{ECONOMY_PUBLIC_BASE_URL}/v2/economy/stripe/return/{ticket}?result=success|cancel`, `expires_at` = now + pending TTL (31 min to 23 h), header `Idempotency-Key: economy-{paymentId}` (per payment, so a replaced one-time payment cannot get the old session back). The session id is kept in `authority`.
- **Stripe return**: `GET /v2/economy/stripe/return/{ticket}` changes nothing; it redirects (302) to `ECONOMY_WEB_REDIRECT_URL` / `ECONOMY_DEEP_LINK` with `success=true` for `result=success` unless the payment already failed or expired. The client then polls `economy_purchase_status`, which stays `pending` until the webhook.
- **Stripe webhook**: `POST /v2/economy/stripe/webhook` (public, API port). `Stripe-Signature` `t=…,v1=…`: HMAC-SHA256 of `{t}.{raw body}` with `webhookSecret`, any `v1` may match, `|now − t| <= stripeToleranceSeconds` (400 `INVALID_SIGNATURE` / `SIGNATURE_TOO_OLD`); `livemode` must match `testMode` (400 `MODE_MISMATCH`); an event id already in `economy_payment_webhook` answers `{received:true, duplicate:true}`. `checkout.session.completed` with `payment_status=paid` or `checkout.session.async_payment_succeeded`: session id, `amount_total` and `currency` must match the payment (else failed `AMOUNT_MISMATCH`); then `pending|expired|failed → success` (the money was taken) and the normal grant + reward retry. `completed` unpaid → pending note `ASYNC_PAYMENT_PENDING`; `async_payment_failed` → failed; `checkout.session.expired` → expired. Unknown payments answer 200 `unknown_payment`; database errors 500 so Stripe retries. Answer `{received:true, result}`.

Payment settings (`GET/PUT /payment-settings`, «پرداخت‌ها ← تنظیمات») added for these methods:

| key | default | meaning |
|---|---|---|
| `liveAcceptsTestPurchases` | `false` | accept Google/Apple test purchases while `sandbox` is off (Apple App Review) |
| `appStorePayload` | `optional` | `required`: a Google/Apple purchase without the payment id is refused (`DEVELOPER_PAYLOAD_MISMATCH`); `optional`: it binds to the payment it is submitted with (review binds to the latest pending/expired payment of that product) |
| `storeConsumed` | `accept` | `refuse`: a Google purchase that is already consumed but unknown here → `PURCHASE_CONSUMED` |
| `googleAcknowledge` | `true` | acknowledge Google purchases after the grant |
| `stripeToleranceSeconds` | `300` | oldest accepted webhook signature, 30..3600 |

## Client RPCs (game, user session)

All take/return JSON. Errors use Nakama runtime errors with a stable `message` code (UPPER_SNAKE).

Language: every RPC that returns text accepts `lang` and returns resolved strings plus `lang` (the language used). Resolution: requested `lang` if it is an active language (`en-US` matches `en`) → the account `lang_tag` if active → default language; per field, a missing value falls back to the default language, then the first non-empty value. An unknown `lang` is not an error.

| RPC | Input | Output |
|---|---|---|
| `economy_catalog` | `{categoryKey?, lang?}` | `{lang, currencies:[{key,name,type,iconUrl,categoryKey,tags}], categories:[{key,name,description,iconUrl,sort}], products:[{id,key,name,description,iconUrl,coverUrl,priceToman,finalToman,coinPrice,discountPercent,paymentMethods,sku,isSubscription,categoryKey,tags}]}` (active only, methods filtered to active methods) |
| `economy_balances` | `{}` | `{wallet:{key:amount}, metas:{key:value} (non server_only), items:[…storage economy_items], regen:[regen view]}` (regen currencies are settled first, so `wallet` already holds the gain) |
| `economy_purchase_start` | `{productId, method, giftCode?, offerId?}` (with `offerId` the player's personal offer is sold at its frozen price, see «Personal offers») | gateway: `{paymentId, status:"redirect", redirectUrl}`; stripe: `{paymentId, status:"redirect", redirectUrl, sessionId, currency, amountMinor, publishableKey?}`; store: `{paymentId, status:"store", sku, developerPayload}`; in_app or free: `{paymentId, status:"success", rewards}` |
| `economy_purchase_submit` | `{paymentId, token}` | `{paymentId, status, rewards?}` |
| `economy_purchase_review` | `{method, sku, token}` | same as submit |
| `economy_purchase_status` | `{paymentId}` | `{paymentId, status, rewardGranted, error?}` (own payments only) |
| `economy_store_config` | `{method}` | `{method, sandbox, rsaPublicKey, packageName}` (+ `bundleId` for `apple`) |
| `economy_gift_redeem` | `{code}` | `{rewards, gift:{name, categoryKey, tags, lang}}` (`tags` are keys of the gift's `gift`/`general` tags) |
| `economy_payments_mine` | `{limit?, cursor?, lang?}` | `{lang, items:[{…, productName (resolved string)}], cursor}` |

`rewards` = the prize items actually applied.

## Console API (console port, console session token, `/v2/console/economy/`)

Wired through `customHttpAuthFunc` + audit log + mux params (extend the `/v2/console/hiro/` branch in `server/console.go` to also cover `/v2/console/economy/`). ACL in `console/acl/acl.go`: GET needs `HIRO_ECONOMY` read, other methods need `HIRO_ECONOMY` write (admins pass). Each handler registered with exactly one method so the ACL/audit gets it.

Generic CRUD for `currencies, metas, prizes, products, gifts` (categories and tags moved to `/v2/console/base/{categories,tags}`; the old economy endpoints are removed):
- `GET /v2/console/economy/{resource}?q=&page=1&limit=20[&type=][&categoryId=][&isActive=]` → `{items, total, page, limit}` (server-side search on key and every translation of the name, newest first)
- `GET /v2/console/economy/{resource}/{id}` → item
- `POST /v2/console/economy/{resource}` → created item
- `PUT /v2/console/economy/{resource}/{id}` → updated item (full replace of editable fields)
- `DELETE /v2/console/economy/{resource}/{id}` → `{}`; refuse with 409 `IN_USE` when referenced (prize used by product/gift, currency used in coin_price or prize items).

JSON fields are camelCase mirrors of the columns (`name:{<lang>: text}` full maps, also for gifts, `tagIds`, `priceToman`, `coinPrice:{currencyKey,amount}`, `paymentMethods`, `sku:{bazaar,myket}`, `prizeId`, `items:[…]`, …). Validation errors: 400 `{"error": CODE, "message": text, "field": name}`.

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
- `GET /v2/console/economy/summary` → counts for the overview page (`categories`/`tags` count the economy key types in `base_category`/`base_tag`). Besides the catalog counts it returns active counts (`activeProducts`, `activeGifts`, `activeOffers`, `activeItems`, `activeTasks`, `activeAchievements`, `activeStreaks`, `activeLevelTables`, `activeMetaRelations`, `activeEventTypes`, `activeEventRules`, `activeChanceGames`, `activeGames`), `meta-rules`, `games`, `friend-gifts` / `friendGiftsPending`, `referrals` / `referralsCompleted`, the previous periods for the deltas (`revenueTomanPrev30` = live revenue 30–60 days ago, `paymentsPrevDay` and `eventsPrevDay` = 24–48 hours ago) and `revenueDaily`: the last 30 calendar days oldest first, `{day, toman, success, payments}` (live gateway and store revenue, successful and all payments that day).

Overview page (`#/economy/overview`): a KPI row (30-day revenue with a sparkline and the change against the previous 30 days, payments in 24 hours against the 24 hours before, events in 24 hours, waiting prizes with the failed count); «کاتالوگ» in the sidebar groups «ارز و پیشرفت», «فروش», «جایزه و بازی», «رویداد و رشد» with one tile per page or tab (count, «N فعال» or another short line, «هنوز چیزی ساخته نشده» at zero) that opens it, categories and tags as chips; «پرداخت‌ها» with a 30-day bar chart (revenue or payment count, time runs in the reading direction, native tooltips and an `aria-label` summary), a status bar with a legend that links to the filtered payment list, the grant-error alert and the gateways. Empty states: no active gateway → «تنظیم درگاه پرداخت» (project settings), no payments yet → the shop. Tiles and sections the admin cannot open are left out.

## Console pages

Sidebar section «اقتصاد» / "Economy" after «منوی اصلی»: نمای کلی، ارزها، متاها، جایزه‌ها، فروشگاه، هدیه‌ها، پرداخت‌ها. دسته‌بندی‌ها و تگ‌ها move to the «عمومی» / "General" section (Base module). Routes `#/economy/<page>[/<id>]`. Payment methods live in the console's Settings page as the «درگاه‌های پرداخت» tab (`#/settings/payment-methods`); `#/economy/payment-methods` redirects there. Pages look like the console's own pages (same Tailwind classes, light/dark, fa/en, Persian digits in fa). Lists: server-side search + paging; on mobile tables become stacked cards. Forms in a side drawer or a detail page; prize item editor is a repeatable row editor; product form shows only the fields relevant to the chosen payment methods.

Product audience (phase 3, `20261111000001-economy-product-audience.sql`): `economy_product.audience_tag_ids text[]` holds player tags (key type `user`, else `TAG_KIND_MISMATCH` on `audienceTagIds`, at most 20). Empty = every player. Otherwise `economy_catalog` lists the product only for players who have at least one of the tags (and marks it `exclusive: true`), and `economy_purchase_start` answers `PRODUCT_NOT_FOUND` to everyone else; restoring a store purchase is not limited. The product editor has «فقط برای این بازیکنان» (player-tag picker with quick-create) and the product list shows «ویژهٔ برچسب». Deleting a tag removes it from audiences.

Console: the currency and gift lists show a «دسته‌بندی» column (category chip and up to three tag chips) and a category filter; their editors have the category picker (key type `currency` / `gift`) and the tag picker (that key type plus `general`), each with the «+» quick-create, like products and prizes. Deleting a category used by a gift or currency is refused (`IN_USE`); deleting a tag removes it from their `tag_ids`. The setup wizard's currency step keeps a currency's category and tags when it saves.

## Events, levels and streaks (phase 4)

Port of the useful parts of asp `Modules/Base/Event` (user events with count/sum/high value), `Modules/Economy/EventPrizeRelation` (event → prize when a value is reached, repeatable or once), `Resource/MetaRelation` (a meta crossing a threshold sets other metas and grants a prize, once per player) and `Resource/Streak` (per-player value with freeze tokens, rewards per streak value, repeatable or once). asp Gamification was not used. Changes against asp: events are admin-defined types (unknown keys are refused), periods are calendar days/weeks in a configured time zone instead of rolling seconds, a streak counts at most once per period (asp added on every event), asp freeze tokens became grace periods per run, asp MetaRelation became ordered level tables, and every reward goes through one claims table so a grant is never doubled. Migration `20261111020000-economy-events.sql` (Up re-runnable, Down raises) adds the tables, widens `economy_user_prize.source_type` with `event`, `level`, `streak`, and seeds four built-in event types.

### Tables

| Table | What |
|---|---|
| `economy_event_settings` (one row) | `timezone` (IANA, default `Asia/Tehran`), `week_start` (0 Sunday … 6 Saturday, default 6), `retention_days` (raw log, default 30), `player_rate_limit` (game events per player per minute, all types, default 120) |
| `economy_event_type` | `key` (`^[a-z][a-z0-9_.]{1,47}$`, immutable), translated `name`/`description`, `client_allowed`, `value_min`/`value_max`, `rate_limit` (per player per minute), `is_active`, `is_system`, `sort`. Seeded built-ins (`is_system`, server only, cannot be deleted): `login`, `purchase_completed` (value = final toman, or the coin amount for coin purchases), `gift_redeemed`, `level_up` (value = new level) |
| `economy_player_event` | raw log: user, type, `value numeric(20,4)`, `metadata` (object ≤ 2 KB), `source` (`client`, `server`, `admin`), time. Rows older than `retention_days` are removed |
| `economy_player_event_total`, `economy_player_event_day` | per player and type: count, sum, max (all time, and per local day; day rows kept `max(retention, 35)` days) |
| `economy_event_rule` | event prize: `event_type_id`, `measure` (`count`, `sum`, `max`), `period` (`total`, `day`, `week`), `threshold > 0`, `repeat` (`once` = one reward ever, `period` = one per day/week window, `every` = one per multiple of the threshold; `every` needs count or sum, `period` needs day or week), `max_claims`, `prize_id`, `grant_mode` (`auto`, `claim`), `audience_tag_ids` (player tags, empty = everyone), `starts_at`/`ends_at`, translated name/description shown to players |
| `economy_event_rule_progress` | per rule, player and window start (`1970-01-01` for total): count, sum, max and `steps` (rewards already issued in that window) |
| `economy_meta_rule` | event → meta: `action` (`add`, `set`, `max`), `value_source` (`fixed` = `amount`; `event` = event value × `amount`), audience and window |
| `economy_level_table`, `economy_level_step`, `economy_user_level` | a source meta (e.g. `xp`), an optional level meta written with the reached level, and steps `(level, threshold, translated title, prize, grant_mode)` with thresholds non-decreasing by level; the reached level per player never goes down |
| `economy_streak`, `economy_streak_reward`, `economy_user_streak` | counting event, `period_days` (1–30), `grace_periods` (missed periods forgiven per run), `miss_action` (`reset` = new run from day one, `step_back` = one day back per missed period beyond the grace), `cycle_length` (rewards repeat every N days, empty = no repeat); rewards per day with prize and grant mode; per player `count`, `best`, `run`, `grace_used`, `last_period` |
| `economy_reward_claim` | one row per reward a player earned: `kind` (`event`, `level`, `streak`), `owner_id` (rule, table or streak), `ref_id` (rule, step or reward), `scope` (`<window>#<n>`, `L<level>`, `r<run>:<day>`, `r<run>f<falls>:<day>`, `m<day>`), prize, grant mode, `status` (`ready`, `granted`, `failed`), attempts. Unique `(kind, ref_id, user_id, scope)` |

### Recording

`recordEvent` (one transaction): check the type (unknown or inactive → `EVENT_TYPE_NOT_FOUND`; a client sending a server-only type → `EVENT_NOT_ALLOWED`; value outside the range → `VALUE_OUT_OF_RANGE`; metadata over 2 KB → `METADATA_TOO_LARGE`), then for client events two `clusterkit.Allow` limits (`ev:<user>:<type>` at the type's rate, `ev:<user>` at the player limit → `TOO_MANY_EVENTS`, 429). It writes the raw row and the totals, then for every active rule of the type in its window and audience upserts the progress row (the upsert row lock serialises the same player on any node), computes the rewards due and inserts claims with `ON CONFLICT DO NOTHING`; then meta rules; then streaks (`FOR UPDATE` on the player's streak row, at most one count per period). After the commit, `auto` claims are granted, and level tables whose source meta changed are re-checked (each in its own transaction under a `FOR UPDATE` on `economy_user_level`); a level-up writes the level meta, inserts level claims, and records `level_up`. Prizes with meta items re-check levels too. Cascades stop at depth 6.

Grants use the existing prize path: `source_type` = claim kind, `source_id` = claim id, so a claim pays exactly once even when two nodes grant it at the same moment; the claim is then marked `granted`. Every node runs `eventLoop` (30 s): auto claims left `ready`/`failed` for over a minute are claimed one at a time with `FOR UPDATE SKIP LOCKED` (an `attempted_at` lease of 2 minutes, at most 8 tries) and granted; every 10 minutes the holder of the `economy-event-sweep` lease removes old raw events, day totals and window progress.

Periods: a day is the calendar day in `timezone`; a week starts on `week_start`; a streak period is `floor(local day number / period_days)`. Changing the zone moves future windows only.

Sources: login after-hooks on every authenticate call (device, custom, email, Google, Apple, Facebook, Steam, Game Center) through `internal/hookchain.After` (base's suspension after-hooks now use the same chain), skipped for suspended players; publication phone login; purchases (once per payment, when `reward_granted` first turns true) and gift codes in economy. Other plugins call `internal/playerevents.Record(ctx, Event{UserID, Key, Value, Metadata})`; economy registers the recorder at start, so the call returns `ErrUnavailable` when economy is not loaded.

### Client RPCs

| RPC | Input | Output |
|---|---|---|
| `economy_event_record` | `{key, value? (default 1), metadata?, lang?}` | `{lang, eventId, key, value, rules:[{key, period, periodStart, periodEndsAt, measure, value, threshold, reached, newClaims}], metas (visible ones changed), streaks:[{key, count, best, advanced, reset}], levels:[{key, from, to}], claims:[{claimId, kind, grantMode, status, prize:{id,key,name,message}}]}` |
| `economy_events_progress` | `{lang?}` | `{lang, timezone, items:[{key, name, description, eventKey, eventName, measure, period, repeat, threshold, target, current, claims, maxClaims, grantMode, claimable:[claimId], done, periodEndsAt, endsAt?, prize}], ready:[…]}` (rules the player's tags allow, in their window) |
| `economy_levels` | `{lang?}` | `{lang, items:[{key, name, description, metaKey, metaName, value, level, title, levelThreshold, next, steps:[{level, threshold, title, prize?, grantMode, status: locked|reached|claimable|pending|claimed, claimId?}]}]}` (tables whose source meta is not server-only) |
| `economy_streaks` | `{lang?}` | `{lang, items:[{key, name, description, eventKey, periodDays, count, best, run, countedThisPeriod, broken, graceLeft, gracePeriods, missAction, cycleLength, nextPeriodAt, breaksAt, claimable, days:[{day, status: locked|today|done|claimable|pending|claimed, prize?, grantMode?, claimId?}]}]}`. A missed streak shows `broken` and count 0 before the next event resets it |
| `economy_reward_claim` | `{claimId, lang?}` | `{claimId, kind, rewards, prize, levels}`; `CLAIM_NOT_FOUND` (also someone else's), `ALREADY_CLAIMED` (409) |
| `economy_rewards_mine` | `{status? (ready|granted|failed), limit?, lang?}` | `{lang, items:[{claimId, kind, source, sourceName, level?, prize, status, grantMode, createdAt, grantedAt?}]}` |

### Console API (`/v2/console/economy/`, `HIRO_ECONOMY`)

CRUD `event-types`, `event-rules`, `meta-rules`, `level-tables` (with `steps`), `streaks` (with `rewards`) through the generic resource handlers (search on key and every name translation; filters `isActive`, `eventTypeId`, `metaId`, `prizeId`, `clientAllowed`). Deleting a type with history or rules, a rule/table/streak with claims, or a built-in type is refused (`IN_USE`, `SYSTEM_EVENT`); level steps and streak rewards are upserted by level/day so claim references stay. Errors: `INVALID_EVENT_KEY`, `SYSTEM_EVENT_SERVER_ONLY`, `INVALID_RANGE`, `INVALID_RATE_LIMIT`, `INVALID_MEASURE`, `INVALID_PERIOD`, `INVALID_REPEAT`, `REPEAT_NEEDS_PERIOD`, `EVERY_NEEDS_TOTAL`, `INVALID_THRESHOLD`, `INVALID_MAX_CLAIMS`, `INVALID_GRANT_MODE`, `INVALID_WINDOW`, `INVALID_META_ACTION`, `INVALID_VALUE_SOURCE`, `SAME_META`, `STEPS_REQUIRED`, `DUPLICATE_LEVEL`, `THRESHOLDS_NOT_ORDERED`, `INVALID_PERIOD_DAYS`, `INVALID_GRACE`, `INVALID_CYCLE`, `DUPLICATE_DAY`, `DAY_OUTSIDE_CYCLE`.

Other routes: `GET|PUT /event-settings` (`INVALID_TIMEZONE`, `INVALID_WEEK_START`, `INVALID_RETENTION`), `GET /event-log?q&typeId&userId&source`, `GET /event-claims?userId&ownerId&kind&status`, `POST /event-claims/{id}/retry`, `GET /users/{id}/progress?lang` (totals, streaks, levels, last 20 rewards, waiting count), `POST /users/{id}/events {key, value, metadata}` (records as `admin`, any active type), `POST /users/{id}/streaks/{streakId}/reset` (new run from zero). `/retry`, `/events`, `/reset` are audit-logged as invoke. The summary has `event-types`, `event-rules`, `level-tables`, `streaks`, `eventsToday`, `claimsReady`, `claimsFailed`.

### Console pages

Economy section, after «هدیه‌ها»: «رویدادها» (`#/economy/event-types`), «جایزهٔ رویدادها» (`event-rules`, list with a readable condition sentence such as «پایان یک دور» در یک روز ۳ بار · هر روز یک بار, and a live preview in the editor), «قانون‌های متا» (`meta-rules`), «جدول سطح‌ها» (`level-tables`, a row editor for levels with translated titles), «روزهای پیاپی» (`streaks`, a row editor for reward days and a note that this is independent of Hiro; the native Hiro «زنجیره‌ها» pages are untouched), «گزارش رویدادها» (`event-log`). Project settings section «رویدادها» (`#/project-settings/events`): zone with the current local weekday and time, week start, log retention, game events per player per minute. The overview shows events of the last 24 hours and waiting/failed rewards. The native player page gets a «پیشرفت بازیکن» card under «برچسب‌ها و محدودیت‌ها» (levels, streaks with «صفر کردن», event totals, recent rewards with retry for failed ones, «ثبت رویداد» dialog, link to the player's log). All pages are in the Ctrl K search; the spatial section tabs now scroll and keep the current page in view.

### Streak milestones (safe days, phase 14)

Owner's request 2026-11-13. Migration `20261121000000-economy-streak-milestones.sql` (Up re-runnable, Down raises) adds `economy_streak.repay_after_fall` (default `false`), the table `economy_streak_milestone` (`streak_id`, `day` 1–3650 unique per streak, optional `prize_id`, `grant_mode` `auto`/`claim`) and two player columns on `economy_user_streak`: `peak` (highest count in the current run; existing rows get `peak = count`) and `falls` (fall-backs in the current run, default 0). A streak without milestones and with the switch off behaves exactly as before: same counts, same claim scopes.

- **Floor.** The floor is the highest milestone day not above `peak`. On a miss beyond the grace the count never goes below it, with either `miss_action`: `reset` goes to the floor instead of a new run (run, `peak` and `best` stay; the grace is fresh), so with a milestone at 20 a player on day 25 who misses days is at 20 and the next counted day is 21; `step_back` steps back one day per missed period as before but stops at the floor. Below the first milestone (`floor = 0`) both actions work as before (`reset` starts run + 1 from day one). Every lowering of the count after a miss adds one to `falls`.
- **Claim scopes.** Day rewards keep `r<run>:<count>`. With `repay_after_fall` on and `falls > 0` the scope is `r<run>f<falls>:<count>`, so each fall starts a new chain and the days re-reached pay again; with it off (the default) the old scope is used, the claim already exists and `ON CONFLICT DO NOTHING` pays nothing, so missing days on purpose to collect days 21–25 again earns nothing. Turning the switch on later pays re-reached days from the next fall; turning it off stops paying any day already paid in the run. Existing claims are untouched by the migration and keep matching.
- **Milestone prize.** Optional, once per player and streak, ever: a claim of kind `streak` with `owner_id = ref_id = streak id` and the fixed scope `m<day>`. It is inserted on each counted day for every milestone with a prize and `day <= count` (so a milestone added later pays on the player's next counted day once they are past it); a new run, a fall-back or a console reset never pays it again, and deleting and re-adding the same day does not either.
- **Cycle.** Milestones are absolute streak days, not cycle days. With `cycle_length` 7 and a milestone at 20, a fall gives count 20, the next day is 21 and pays cycle day `((21-1) mod 7) + 1 = 7`. Because day scopes use the absolute count, the cycle still repeats rewards for new counts; counts re-reached after a fall are blocked or paid by the switch like any other day.
- **Best.** `best` is the highest count ever and is never lowered by a fall or a reset. `peak` is per run and decides the floor; a new run (only possible below the first milestone) and the console reset set it to 0.
- **Console reset.** `POST /users/{id}/streaks/{streakId}/reset` (player card «صفر کردن») sets run + 1, count 0, `peak` 0, `falls` 0, so the floor is cleared; milestone prizes already won stay won.
- **Parallel events.** The `FOR UPDATE` on the player's streak row serialises events of the same player on any node: ten parallel events after a fall count once from the floor and record one fall.

API: the console streak resource takes and returns `repayAfterFall` and `milestones:[{id, day, prizeId|null, grantMode}]` (rows upserted by day; `INVALID_MILESTONE_DAY`, `DUPLICATE_MILESTONE`, `TOO_MANY_MILESTONES` (50), `PRIZE_NOT_FOUND`, `INVALID_GRANT_MODE`). `economy_streaks` items (and the console player progress) add `floor`, `peak`, `nextMilestone` (first milestone above `peak`, or null), `floored` (true when the streak is broken and the floor stops the fall), `repayAfterFall` and `milestones:[{day, reached (day <= peak), safe (day = floor), status: locked|reached|claimable|pending|claimed, prize?, grantMode?, claimId?}]`; a broken streak shows the floor as its count instead of 0, and the day tiles show re-reached days as claimed when they will not pay again. The `economy_event_record` streak entries add `floored`, `floor` and `nextMilestone`.

Console: the «روزهای پیاپی» editor has a «نقطه‌های امن» row editor (day, optional one-time prize, grant mode) with a short note, and the switch «جایزهٔ روزها دوباره داده شود» (off by default, with the farming warning); the list shows «نقطهٔ امن: روز ۳، ۱۰». The player card shows «تا روز ۲۰ امن است · نقطهٔ امن بعدی: روز ۵۰». Star Catcher seeds a milestone at day 7 with 100 coins to collect on `daily_login` (only when the streak has no milestone yet) and shows «تا روز ۷ امن است», the next safe day with the days left, the milestone prize chip, and «… به روز ۷ برگشتی، نه روز اول» after a miss.

### Boosters, special days, week start and the infinity phase (2026-11-21)

- **Timed booster** prize item: `{type: "booster", key: <meta key>, action: give|set, value: {multiplier 1.01-100, minutes 1-525600}}`. Granting upserts `economy_booster (user_id, meta_id, multiplier, expires_at)`: `give` adds the minutes to a running booster (or starts one) and overwrites the multiplier (legacy: no stacking, time extends), `set` restarts it now. While active, positive `add` meta rules on that meta are multiplied (prize meta items, `set` and `max` rules are not). `economy_boosters {lang}` lists the player's active boosters `{metaKey, metaName, multiplier, expiresAt, secondsLeft}`. The console prize editor has the type «افزایش‌دهندهٔ زمان‌دار» with a meta picker, multiplier and minutes. Legacy: `xp-booster-2x-1h`, 2× for 60 minutes.
- **Special days** (`economy_streak_extra (streak_id, day, prize_id, grant_mode)`, `extras` in the streak resource, max 100): an extra claim (scope `x<day scope>`) on track day `count`, only when the period advanced without a miss. Shown in `economy_streaks` as `extras[] {day, prize, grantMode, status, claimId}`. Legacy: day 8, 15, 22 and 28 chests.
- **`missAction: "week_start"`** with `weekLength` (1–365, default 7): after the grace periods, a break sends the count to the first day of the current block, `((count-1)/weekLength)*weekLength + 1` (missing on day 12 → day 8; finishing day 7 and missing → day 1), never below the milestone floor. Counts as a fall (`falls+1`, the run stays).
- **Infinity phase** (`infinity`, `infinityCurrency`, `infinityMin`, `infinityMax`; needs `cycleLength`, `INFINITY_NEEDS_CYCLE`, `INVALID_INFINITY_RANGE`): once the peak reached `cycleLength`, every later advance pays a uniform amount in `[min, max]` of the currency straight to the wallet (grant id `streak:<id>:r<run>:<count>`, reason `streak_infinity`), gives no day rewards or extras, and a missed period no longer breaks the streak. `economy_streaks` has `phase: track|infinity` and `infinity {currencyKey, min, max}`; streak changes in event results carry `phase`, `infinityCurrency`, `infinityAmount`. Legacy: 28-day track, 150–180 coins.

## Meta relations and the meta studio (phase 20)

Owner's request 2026-11-14: the legacy Rakhtkan `meta_relation` module (backend `economy/resource/meta_relation`, panel `economy/subModules/meta-relations` and `meta-relation-studio-sheet.tsx`) next to the level tables, which stay as they are. Migration `20261125000000-economy-meta-relations.sql` (Up re-runnable, Down raises) adds the tables, widens `economy_reward_claim.kind` and `economy_user_prize.source_type` with `relation`, adds two policy columns to `economy_event_settings` and seeds the built-in server event `MetaRelation:Completed`.

### Tables

| Table | What |
|---|---|
| `economy_meta_relation` | `key` (immutable), translated `name`/`description`, source `meta_id`, `operator` (`gt`, `gte`, `lt`, `lte`, `eq`), `value numeric(20,4)`, `targets` (JSON `[{metaId, action: add\|set, amount}]`, at most 20, each meta once per action, adding zero refused), `prize_type` (`none`, `prize`, `prizeGroup`) with `prize_id` or `prize_group_id`, `grant_mode` (`auto` default, `claim`), `repeat` (`once` default, `rearm`, `every`), `max_fires` (for `rearm`/`every`, empty = no cap), `audience_tag_ids`, `client_visible` (default true), `starts_at`/`ends_at`, `is_active`, `sort`. A relation needs a target or a prize (`RELATION_NEEDS_RESULT`) |
| `economy_user_meta_relation` | per relation and player: `fires`, `armed`, `first_done_at`, `last_done_at` |
| `economy_event_settings` | `relation_chain_depth` (0–25, default 8): how many relations may trigger one another in a row, 0 = a relation's targets never run another relation; `relation_prize_missing` (`complete` default, `wait`): what a relation does when its prize group has nothing to draw |

### Evaluation

The meta-change hook (`evaluateLevels`) now runs the meta relations of the changed metas first, then their level tables. It is reached from every path that writes a player meta: event meta rules, prizes from claims, tasks and achievements, and (new) every other `grant()` (purchases, gift codes, admin prize grants, game prizes), the new admin meta change, level metas written by a level-up, and relation targets themselves. Relations are not cut off by the event cascade depth (6); their own chain depth applies.

For each active relation of a changed meta, in `sort, created_at` order, inside its window and audience, one transaction locks the player's row (`FOR UPDATE` on `economy_user_meta_relation`, so parallel changes on many nodes fire it once), reads the current value (stored value or the meta default) and compares it at 4 decimals. `once` fires when it never fired; `rearm` fires while armed and re-arms when the condition is checked and no longer holds (the classic «100 XP → level + 1, XP − 100»); `every` fires on every change while the condition holds; `max_fires` caps both. A fire applies all targets in the same transaction (`add` starts from the meta default for a new row, `set`), draws a group member by weight when the prize is a group, inserts a `relation` claim (scope `F<n>@<time>`) and counts the fire. After the commit: `auto` claims are granted, the changed target metas are evaluated one chain step deeper, `MetaRelation:Completed` is recorded (value = fires, metadata `entityId` = relation id and `key`, so event prizes, tasks and achievements can hook it), and the activity log gets `meta_relation`. A relation fires at most once per cascade (loops such as a → b → a stop there), but the cascade can still re-arm it.

### Client RPC

| RPC | Input | Output |
|---|---|---|
| `economy_meta_relations` | `{completedOnly?, lang?}` | `{lang, items:[{key, name, description, repeat, fires, maxFires?, completed, done, active, grantMode, prize? or prizeGroup?, condition?:{meta, metaName, operator, value, current, holds}, claims:[{claimId, status: claimable\|pending\|claimed, grantMode, prize, createdAt}], lastDoneAt?, endsAt?}], ready:[{claimId, kind: relation, source, sourceName, prize, createdAt}]}`: visible relations the player completed plus, unless `completedOnly`, the active ones their tags allow. The condition is left out when the source meta is server-only |

Relation prizes are claimed with `economy_reward_claim` and listed by `economy_rewards_mine` like every other reward; the `economy_event_record` answer carries `relations:[{relationId, key, fires, claimId?}]`.

### Console API (`/v2/console/economy/`, `HIRO_ECONOMY`)

CRUD `meta-relations` (filters `metaId`, `targetMetaId`, `isActive`, `prizeId`; list rows carry `players` and `claims`; delete refused with claims, `IN_USE`). Metas used by a relation (source or target), and prizes and prize groups used by one, cannot be deleted. New filters for the studio: `prizes?metaKey=`, `level-tables?levelMetaId=`. `POST /users/{id}/metas {metaId, action: add|set, value, note?}` changes a player meta and runs the relations and level tables (activity `meta_changed`); `GET /users/{id}/meta-relations` and `DELETE /users/{id}/meta-relations/{relationId}` (forget a completion so it can fire again; received prizes stay, uncollected ones are removed). `/users/{id}/progress` adds `relations`. `PUT /event-settings` accepts a partial body and checks `relationChainDepth` (`INVALID_CHAIN_DEPTH`) and `relationPrizeMissing` (`INVALID_POLICY`). Errors: `INVALID_OPERATOR`, `RELATION_NEEDS_RESULT`, `TOO_MANY_TARGETS`, `DUPLICATE_TARGET`, `ZERO_AMOUNT`, `INVALID_PRIZE_TYPE`, `PRIZE_GROUP_NOT_FOUND`, `INVALID_REPEAT`, `INVALID_MAX_FIRES`.

### Console pages

«متاها» gets a tab «رابطهٔ متاها» (`#/economy/metas/relations`; `#/economy/meta-relations` redirects): list with the condition and results as chips, players, repeat and status; editor with the condition, a row editor for target metas, prize or prize group with grant mode, repeat and cap, a live preview, player tags, window and visibility. Clicking a meta in «متاها» opens the meta studio (`#/economy/metas?studio=<id>`), a side panel with the whole flow of that meta: inputs (event meta rules, relations that change it, prizes that give it, level tables that write their level into it), the meta itself, and outputs (relations with a condition on it, level tables on it). Rules, relations and level tables are created and edited inside the panel with the meta locked; Esc goes back, then closes. Project settings «رویدادها» gets the two relation policies. The player progress card gets «تغییر متا» and «رابطه‌های انجام‌شده» with «از نو».

### Differences from the legacy module

Five comparators instead of three (`more`, `less`, `equal`); repeat modes and a cap (legacy fired once per player only); prize groups; window, audience, order and visibility (legacy ordered by creation); chain depth and the empty-group policy in the settings instead of a hard-coded 25; all targets and the claim in one transaction (legacy applied targets best effort, each committing on its own); prizes through the shared claims table so they pay once on any node; a client view of active relations with progress and claims, not only completions.

## Personal offers: daily offers and entry offers (phase 21)

Owner's request 2026-11-14 (Ludo-style games). Before this phase the shop had a product discount (`discount_percent`), product audience tags, level limits and one-time products, and gift codes; there was no per-player offer, no time-limited offer and no timer. Offers are built on top of the shop instead of beside it: an offer is a template over an existing product (the product keeps its prize, its rules and its payment methods), and buying an offer is a normal shop purchase with the offer's frozen price. Migration `20261126000000-economy-offers.sql` (Up re-runnable, Down raises) adds the tables below, `economy_payment.user_offer_id` and the built-in server event `Offer:Bought`.

### Tables

| Table | What |
|---|---|
| `economy_offer` | the admin pool. `key` (immutable), `kind` (`daily` or `entry`), translated `name` (empty = the product name), `description`, `badge`; `product_id`; price: `price_mode` `discount` (`discount_percent` 0–100 on the product's Toman and coin price; 0 keeps the product price and its own discount) or `price` (special `price_toman` 0 or ≥ 100, and/or special `coin_price`; an empty one keeps the product's), `sku` (offer SKUs for Bazaar and Myket); `purchase_limit` per player and copy (default 1); `weight` (daily draw, default 10, 0 = never drawn); audience: `audience_tag_ids` (any of), `exclude_tag_ids`, `min_level`/`max_level` on `level_table_id` (empty = main table), `meta_conditions` (`[{metaId, operator, value}]`, up to 10, same comparators as meta relations), `min_account_days`/`max_account_days`; entry only: `trigger_type` (`first_entry`, `level`, `meta`, `tag`, `account_days`, `lapsed`) with `trigger_value` and `trigger_ref` (level table, meta or player tag), `duration_minutes` (default 1440), `repeat` (`once` default, `cooldown`), `cooldown_hours` (default 168), `max_triggers` (empty = no cap); `starts_at`/`ends_at`, `is_active`, `sort` (for entry offers: priority) |
| `economy_offer_settings` | one JSON row with the policies below; a missing row or field means the default |
| `economy_user_offer` | a personal copy: player, template, kind, `period_start` and `slot` (daily), status (`active`, `replaced` by a reroll, `closed` by an admin), the price frozen when the copy was made (`original_toman`, `final_toman`, `coin_currency`, `coin_amount`, `coin_original`, `discount_percent`, `sku`), `purchase_limit`, `bought`, `reroll` number, `forgotten` (admin reset of an entry offer), `triggered_at`, `shown_at`, `expires_at`, `last_bought_at` |
| `economy_user_offer_period` | one row per player and daily period: the reroll count and whether the set was drawn; its row lock keeps two parallel calls from drawing twice |
| `economy_user_offer_state` | the player's last two login times (written by the login hook) for the lapsed-return trigger |

### Settings (`GET/PUT /offer-settings`, partial body)

| Setting | Default | Meaning |
|---|---|---|
| `dailyEnabled` | true | daily offers on |
| `timezone`, `resetHour`, `resetMinute` | Asia/Tehran, 0, 0 | when a period starts |
| `periodDays` | 1 | period length 1–30 days; periods are aligned on 1970-01-01 in the zone so every node and player agrees |
| `count`, `countByTag` | 3, [] | offers per player and period; a player tag in `countByTag` raises it (largest wins) |
| `weighting` | `weighted` | `weighted` (by weight, without replacement), `uniform` (weight ignored), `priority` (heaviest first, ties at random) |
| `noRepeatDays` | 1 | a template the player got in the last N days (before this period) is drawn only after the fresh ones; 0 = off |
| `fillWithRepeats` | true | when fresh templates run out, fill the slots with recent ones |
| `boughtCooldownDays` | 7 | a template the player bought is not drawn for N days; −1 = never again; 0 = off |
| `excludeOwned` | true | skip a template whose product only gives items the player owns |
| `respectProductRules` | true | skip a template whose product the player cannot buy now (level, one-time bought, owned contents, missing required item, non-repeatable prize received) |
| `inactiveProducts` | `allow` | `allow`: products switched off in the shop can still be sold as offers (offer-only packs); `skip` |
| `showSoldOut` | true | a bought-out offer stays listed with `soldOut` until it ends |
| `giftCodes` | false | gateway discount codes also apply on top of an offer price |
| `rerollEnabled`, `rerollLimit` | true, 1 | rerolls per period (0 = no limit) |
| `rerollCost`, `rerollCostStep` | null, 0 | `{currencyKey, amount}` or null for free; each reroll costs `step` more |
| `rerollAvoidCurrent` | true | the new draw leaves out the offers it replaces |
| `entryEnabled` | true | entry offers on |
| `entryMaxActive` | 1 | entry offers open at once per player; the rest wait, highest priority (`sort`) first |
| `entryTimerStart` | `list` | when a player's countdown starts: `list` (the first `economy_offers` that returns it), `seen` (the game calls `economy_offer_seen`), `trigger` (the moment the trigger held) |

### Selection (daily)

The first `economy_offers` call of a period draws the player's set once: active templates in their window, minus the ones the player's tags, levels, metas or account age rule out, the product checks above and templates no active payment method can sell (`no_method`), the bought cooldown and weight 0. The draw is seeded with the player, the period and the reroll number, so a preview made by the admin shows exactly what the player will get; fresh templates first, then (with `fillWithRepeats`) the recent ones. The copies keep their price for the whole period even if the admin edits the template; switching a template off hides its copies at once. Settings changes apply from the next period (or after «قرعهٔ دوباره» on the player card).

A reroll (`economy_offers_reroll`) checks the limit, takes the cost from the wallet, then in one transaction replaces every unbought copy of the period (status `replaced`) with a new draw (bought ones stay and keep their slot; the replaced templates are left out when `rerollAvoidCurrent`); when the draw is empty (`NO_OFFERS_LEFT`) or anything fails the cost is refunded.

### Entry offers and timers

On every `economy_offers` call (entry offers on) the player's entry templates are judged in `sort` order: the template rules, the product checks and the trigger (`first_entry` always; `level` ≥ value on the chosen table; `meta` ≥ value; `tag` held; `account_days` ≥ value; `lapsed`: the previous login is at least value days before the latest one). A template opens a copy when its trigger holds, no copy is open, `max_triggers` is not reached, and either it never opened or `repeat` is `cooldown`, the cooldown since the previous copy ended (expired, bought out or closed) has passed and, for `lapsed`, the return is newer than the previous copy. At most `entryMaxActive` copies are open; the state row lock serialises parallel calls. The countdown is the player's own: `expires_at = start + duration_minutes`, where start follows `entryTimerStart`; an expired copy is gone for good. A copy bought out stays listed (with `showSoldOut`) until its timer ends.

### Buying

`economy_offer_buy {offerId, method, giftCode?, lang?}` (same as `economy_purchase_start` with `offerId`) checks that the copy is the caller's, active, its template live, not expired (`OFFER_EXPIRED`) and not bought out (`OFFER_SOLD_OUT`); a not yet started entry timer starts. The product is then sold through the normal flow with the frozen price: the method must be one of the product's methods the offer can sell (`in_app` needs a coin price, gateways a Toman price of 0 or ≥ 100, stores an offer SKU, or the product SKU when the Toman price did not change), the product rules and non-repeatable prizes still apply, and the offer's own audience replaces the product audience (and, with `inactiveProducts: allow`, the product's shop switch). The payment row carries `user_offer_id` and the offer price; coin purchases reserve the limit before the coins move (released when the purchase fails), gateway and store purchases count when their prize is delivered (so two gateway payments started in parallel for a limit of one can both be delivered after payment). Store checks (`economy_purchase_submit`, `economy_purchase_review`) use the offer SKU. A delivered offer records `Offer:Bought` (value = Toman paid; metadata `key`, `kind`, `offerId`, `paymentId`) besides `purchase_completed`.

Behaviour of the existing shop: unchanged. `economy_catalog` lists products at their own price, `economy_purchase_start` without `offerId` works as before, and purchase counts of offers count as purchases of their product (so one-time products stay one-time).

### Client RPCs

| RPC | Input | Output |
|---|---|---|
| `economy_offers` | `{lang?}` | `{lang, now, daily:{enabled, periodStart, periodEnd, secondsLeft, items:[offer], reroll:{enabled, used, limit, left?, cost:{currencyKey, amount}\|null}}, entry:{enabled, timerStart, items:[offer]}}`; `offer` = `{offerId, key, kind, name, description, badge, product:{id, key, name, iconUrl, coverUrl, items}, originalToman, finalToman, coinPrice:{currencyKey, amount}\|null, coinOriginal, discountPercent, paymentMethods, sku, purchaseLimit, bought, remaining, soldOut, started, expiresAt, secondsLeft, shownAt?, slot, durationMinutes? (entry)}` |
| `economy_offer_buy` | `{offerId, method, giftCode?, lang?}` | as `economy_purchase_start` |
| `economy_offers_reroll` | `{lang?}` | the `economy_offers` answer plus `paid` (`{currencyKey, amount}` or null); errors `REROLL_DISABLED`, `REROLL_LIMIT`, `NOTHING_TO_REROLL`, `NO_OFFERS_LEFT`, `INSUFFICIENT_BALANCE`, `REROLL_BUSY` |
| `economy_offer_seen` | `{offerId}` | `{offerId, expiresAt, secondsLeft}`; starts an entry countdown in timer mode `seen` (marks a daily copy shown) |

`economy_purchase_start` takes an optional `offerId` (then `productId` may be left out; a different one is `OFFER_PRODUCT_MISMATCH`).

### Console API (`/v2/console/economy/`, `HIRO_ECONOMY`)

CRUD `offers` (daily) and `entry-offers` (entry), same fields; list rows carry `productKey`, `productName`, `productActive`, `made` and `bought`; filters `isActive`, `productId`; delete refused once a copy was bought (`IN_USE`; switch it off instead). `GET/PUT /offer-settings` (`{settings, periodStart, periodEnd}`; errors `INVALID_TIMEZONE`, `INVALID_SETTING`, `INVALID_POLICY`, `TAG_NOT_FOUND`, `DUPLICATE_TAG`, `CURRENCY_NOT_FOUND`). `GET /offer-preview?userId=` (current copies of the period, the draw with `reason` per template: `picked`, `recent`, `bought_recently`, `weight_zero`, `inactive`, `audience`, `excluded_tag`, `level_low`, `level_high`, `meta`, `account_new`, `account_old`, `product_inactive`, `owned`, `product_level`, `product_bought`, `product_requires`, `no_method`; entry templates with `holds`, `due`, `open`, `count`). `GET /offer-report?days=30&kind=` (per template: made, shown, players, buyers, purchases, conversion = buyers / shown, Toman revenue, coins by currency; totals with rerolls). Player: `GET /users/{id}/offers`, `POST /users/{id}/offers/reset-daily` (close this period's copies and draw again on the next open), `POST /users/{id}/offers/{copyId}/expire`, `POST /users/{id}/entry-offers/{offerId}/reset` (forget the history so the trigger can open it again; purchases stay in the report). Admin actions go to the activity log. Template errors: `PRODUCT_NOT_FOUND`, `INVALID_PRICE_MODE`, `INVALID_DISCOUNT`, `INVALID_PRICE`, `PRICE_REQUIRED`, `CURRENCY_NOT_FOUND`, `INVALID_SKU`, `INVALID_LIMIT`, `INVALID_WEIGHT`, `INVALID_LEVEL`, `LEVEL_TABLE_NOT_FOUND`, `TOO_MANY_CONDITIONS`, `META_NOT_FOUND`, `INVALID_OPERATOR`, `INVALID_DAYS`, `INVALID_TRIGGER`, `TAG_NOT_FOUND`, `INVALID_DURATION`, `INVALID_REPEAT`, `INVALID_COOLDOWN`, `INVALID_MAX_TRIGGERS`, `INVALID_WINDOW`.

### Console pages

Sidebar «فروش» gets «پیشنهادهای ویژه» (`#/economy/offers`) with tabs «روزانه» (pool list with price, weight, bought/made and status; editor with product and price, offer SKUs, limit, weight, audience, level, meta conditions, account age and window), «پیشنهاد ورود» (`/entry`; the editor adds the trigger with a live sentence such as «وقتی بازیکن بعد از دست‌کم ۷ روز غیبت برگردد. بعد از آن ۲ ساعت ۳۰ دقیقه فرصت دارد.», duration and repeat), «پیش‌نمایش» (`/preview?userId=`: pick a player or paste an ID; current copies with their countdown, the draw with the reason for every template, entry offers with their state), «گزارش» (`/report`: range and kind filters, totals and per-offer shown, buyers, purchases, conversion and revenue) and «تنظیمات» (`/settings`, every setting above with the current period). The player page gets an «پیشنهادهای ویژه» card: today's copies and the entry offers with price, state and time left, «قرعهٔ دوباره», «بستن» per copy, «از نو» per entry offer and a link to the preview.

## Friend gifts (phase 22)

Owner's request 2026-11-14 (Ludo-style games). A player gives a friend something: an item or currency from their own inventory, or a shop product they pay for. Before this phase the only "gift" was the gift code (`economy_gift`, page «هدیه‌ها»); friend gifts are a separate feature with their own names everywhere (`economy_friend_gift*`, RPCs `economy_friend_gift_*`, console «هدیه به دوست», `#/economy/friend-gifts`). Migration `20261127000000-economy-friend-gifts.sql` (Up re-runnable, Down raises) adds the tables below, `economy_item.giftable` and `economy_product.giftable` (NULL = follow the defaults), `economy_payment.friend_gift_id`, the `gift` source on `economy_user_item`, the `friend_gift` source on `economy_user_prize` and the built-in server events `Gift:Sent` and `Gift:Received`.

### Tables

| Table | What |
|---|---|
| `economy_friend_gift` | one gift and its audit trail: `sender_id`, `recipient_id`, `source` (`inventory`, `shop`), `kind` (`item`, `currency`, `product`), `item_id`/`item_key` and `item_state` (the sender's copy: source, upgrade level, equipped), `currency_key`/`amount`, `product_id`/`prize_id`/`payment_id`/`method`, `message` and `message_flagged`, `delivery` (`instant`, `accept`, frozen at send time), `status`, `outcome`, `error`, `request_id` (idempotency, unique per sender), `day` (project day it counts against), `expires_at`, `delivered_at`, `responded_at`, `reversed_at`/`reversed_by`/`reverse_note` |
| `economy_friend_gift_rule` | limits per player tag: `name`, `tag_ids`, `priority`, `per_sender_daily`, `per_recipient_daily`, `per_pair_daily` (NULL keeps the setting, 0 = no limit), `can_send`, `is_active` |
| `economy_friend_gift_settings` | one JSON row with the policies below; a missing row or field means the default |

Status: `sending` (currency between the row and the wallet), `paying` (shop gift waiting for its payment), `pending` (waiting for the recipient), `delivered`, `declined`, `expired`, `cancelled` (taken back by the sender), `reversed` (admin), `failed` (nothing moved: payment closed, not enough coins). Outcome: `delivered`, `returned` (back to the sender), `refunded` (coins back), `sender_got` (the sender got the product), `refund_due` (money to refund by hand), `kept` (nobody got it); empty while the wallet or prize step is still to run (a sweep every 30 s retries it).

### Settings (`GET/PUT /friend-gift-settings`, partial body)

| Setting | Default | Meaning |
|---|---|---|
| `enabled` | true | off: no new gifts; pending ones stay |
| `recipients` | `friends` | `friends` (both accepted) or `any` player |
| `minFriendDays` | 0 | friends for at least N days (from the accept time) |
| `minMutualFriends` | 0 | sender and recipient share at least N friends |
| `minSenderAccountDays`, `minRecipientAccountDays` | 0, 0 | account age |
| `minSenderLevel`, `levelTableId` | 0, main table | sender level |
| `inventoryEnabled` | true | items from the inventory |
| `itemsDefault`, `itemTypes` | true, [] | default giftability, per item type `[{typeId, giftable}]`; an item's own `giftable` wins; starter items never |
| `equipped` | `block` | `block` (take it off first) or `unequip` |
| `upgradeLevel` | `keep` | `keep` or `reset` the item's upgrade level on the recipient |
| `regift` | `block` | an item received as a gift can (`allow`) or cannot be given again |
| `itemRequirements` | `check` | the recipient must own the required/parent item and reach the unlock level (`ignore`) |
| `currencyEnabled`, `currencies` | false, [] | currency gifts, `[{currencyKey, min, max}]` (max 0 = no limit) |
| `shopEnabled`, `productsDefault` | true, true | shop gifts; a product's own `giftable` wins; subscriptions never |
| `methods` | `in_app`, `gateway`, `store` | payment kinds allowed for shop gifts (only the product's active methods) |
| `partialOwned` | `block` | a product whose items the recipient owns in part: `block` or `allow` (they get the rest) |
| `stackableProducts` | `allow` | products without items (coins, metas) |
| `productRules` | `recipient` | whose level is checked against the product min/max level: `recipient`, `sender`, `both`, `none` |
| `discountCodes` | false | gateway discount codes on gift purchases |
| `perSenderDaily`, `perRecipientDaily`, `perPairDaily` | 5, 10, 2 | per project day (events time zone); 0 = no limit |
| `matchPolicy` | `priority` | several tag rules: the first by priority, or `generous` (largest limit, 0 wins; may send if any rule allows) |
| `sendBonusMeta`, `receiveBonusMeta` | none | a meta whose value adds to the sender/recipient limit (sell it in the shop for purchasable quota) |
| `delivery` | `instant` | `instant` or `accept` (the server holds the gift until the recipient answers) |
| `expiryHours` | 72 | time to accept |
| `onExpiry` | `return` | `return` or `accept` automatically |
| `coinReturn` | `refund` | a declined/expired/refused coin purchase: `refund`, `sender` (product to the sender), `none` |
| `moneyReturn` | `sender` | the same for a money purchase: `sender`, `refund_due`, `none` (money is never refunded automatically) |
| `senderCancel` | true | the sender can take back a pending gift |
| `messageEnabled`, `messageMaxLength` | true, 120 | optional note, letters |
| `messageFilter` | `rules` | the chat word list: `rules` (each word's mask/block/flag), `block` (any listed word refuses), `off` |
| `notifyChannel` | `in_app` | recipient notice in the inbox, `push_in_app` (falls back to inbox when push cannot be queued) or `none` |
| `socketNotify` | true | live notification (code 1310, subject `friend_gift`, content `{type, event, giftId, status, kind, senderId, recipientId}`; events `received`, `pending`, `accepted`, `declined`, `cancelled`, `expired`, `reversed`) |
| `notifySender` | true | the sender hears about accept, decline and returns (socket; inbox when something came back) |

### Rules

Who: never yourself; the recipient must exist and not be disabled; a block on either side (`user_edge` state 3) always refuses (`GIFT_BLOCKED`), even with `recipients: any`; then the friend, friendship age, mutual friends, account ages and sender level. Limits: the sender's tag rules give the sender and pair limits and `can_send`, the recipient's tag rules give the receive limit; the bonus metas add to non-zero limits. Gifts count against the day they were sent unless they never left the sender (`failed`, `cancelled`); a shop gift counts from the moment its payment starts. Every send runs in a transaction that holds advisory locks on both players, so parallel sends cannot pass a limit or give the same item twice.

What: only what the recipient lacks. An item the recipient owns or is already getting (pending) is refused (`RECIPIENT_OWNS_ITEM`, `GIFT_ALREADY_PENDING`). A product is refused when its non-repeatable prize was already received, a one-time product was bought by or given to the recipient, all its items are owned, or (with `partialOwned: block`) some are; another gift of the same product to the recipient that is pending or being paid also blocks (`GIFT_ALREADY_PENDING`). The checks run again at delivery: a recipient who got the thing meanwhile makes the gift go back per the return settings.

Inventory items move inside the send transaction (out of the sender, into the recipient with source `gift` or into escrow for `accept`). Currency moves through the wallet ledger right after (`grantId` per step, so retries never move twice). A shop gift creates the gift row (`paying`) and starts a normal payment paid by the sender with `friend_gift_id` on it: coins are taken at once; gateways redirect; stores answer `{status: 'store', sku, developerPayload}` and finish with `economy_purchase_submit`. When the payment is paid (callback, submit, review or reconcile), the prize is granted to the recipient (`economy_user_prize.source_type = friend_gift`) or the gift waits for the recipient. A payment that fails or expires fails the gift. Gift payments are not one-time purchases of the sender and do not count in the sender's own purchase counts.

Accept mode: accept delivers; decline (recipient), cancel (sender) and expiry give it back: items and currency always to the sender, products per `coinReturn`/`moneyReturn` (a product the sender cannot receive becomes a refund for coins or `refund_due` for money).

Events: `Gift:Sent` on the sender when the gift leaves them (inventory: at send; shop: when paid), `Gift:Received` on the recipient at delivery; value = amount (1 for items and products), metadata `giftId, kind, source, key, senderId, recipientId` plus `itemKey`, `currency`/`amount` or `productKey`/`method`. Activity lines `prize/gift_sent`, `gift_received`, `gift_returned`, `gift_reversed`.

Admin reverse (`POST /friend-gifts/{id}/reverse {returnTo, note}`): a pending gift is cancelled; a delivered one is taken back from the recipient (the item if they still have it, else `ITEM_GONE`; the currency, or the product's coins, items and added metas; `INSUFFICIENT_BALANCE` when spent) and goes to `returnTo`: `sender`, `refund` (coin purchases) or `none`. The reversal is idempotent per gift and logged on both players.

### Client RPCs

| RPC | Input | Output |
|---|---|---|
| `economy_friend_gift_options` | `{friendId, lang?}` | `{lang, friend:{id, username, displayName, isFriend}, canSend, reason, day, limits:{sender, recipient, pair: {used, limit, remaining?, bonus?}}, delivery, expiryHours, messageEnabled, messageMaxLength, discountCodes, items:[item + {upgradeLevel, equipped, recipientOwns, giftable, reason}], currencies:[{key, name, iconUrl, min, max, balance}], products:[{id, key, name, description, iconUrl, coverUrl, priceToman, finalToman, coinPrice, discountPercent, paymentMethods, sku, items, recipientOwns, giftable, reason}]}` |
| `economy_friend_gift_send` | `{friendId, itemKey \| currencyKey+amount \| productId+method(+giftCode), message?, requestId?, lang?}` | `{gift, replayed?}`; shop gifts add the purchase answer (`paymentId`, `status` `success`/`redirect`/`store`, `redirectUrl`, `sku`, `developerPayload`) |
| `economy_friend_gifts` | `{box: received\|sent\|all, status?, limit?, cursor?, lang?}` | `{lang, items:[gift], cursor, pendingCount}` |
| `economy_friend_gift_accept` / `_decline` | `{giftId, lang?}` | `{gift}` (recipient) |
| `economy_friend_gift_cancel` | `{giftId, lang?}` | `{gift}` (sender) |

`gift` = `{id, box, kind, source, status, outcome, delivery, name, iconUrl, itemKey, currencyKey, amount, productKey, items?, message, createdAt, expiresAt, secondsLeft?, deliveredAt, respondedAt, from:{id, username, displayName}, to:{…}}`; the sender also sees `method`, `paymentId`, `paymentStatus` (the recipient never sees the price). `economy_purchase_status` adds `friendGiftId` and `friendGiftStatus` for gift payments; `economy_payments_mine` rows carry `friendGiftId`.

Errors (Nakama RPC error message = code): `FRIEND_GIFTS_OFF`, `INVENTORY_GIFTS_OFF`, `SHOP_GIFTS_OFF`, `GIFT_SELF`, `GIFT_SOURCE_REQUIRED`, `RECIPIENT_NOT_FOUND`, `NOT_FRIENDS`, `FRIENDSHIP_TOO_NEW`, `NOT_ENOUGH_MUTUAL_FRIENDS`, `GIFT_BLOCKED`, `SENDER_ACCOUNT_TOO_NEW`, `RECIPIENT_ACCOUNT_TOO_NEW`, `SENDER_LEVEL_TOO_LOW`, `GIFT_SENDING_BLOCKED`, `GIFT_SENDER_DAILY_LIMIT`, `GIFT_RECIPIENT_DAILY_LIMIT`, `GIFT_PAIR_DAILY_LIMIT` (429), `ITEM_NOT_FOUND`, `ITEM_NOT_OWNED`, `STARTER_ITEM`, `ITEM_NOT_GIFTABLE`, `ITEM_EQUIPPED`, `REGIFT_BLOCKED`, `RECIPIENT_OWNS_ITEM`, `GIFT_ALREADY_PENDING`, `RECIPIENT_REQUIRES_ITEM`, `RECIPIENT_LEVEL_TOO_LOW`, `CURRENCY_NOT_GIFTABLE`, `INVALID_AMOUNT`, `INSUFFICIENT_BALANCE`, `PRODUCT_NOT_FOUND`, `PRODUCT_NOT_GIFTABLE`, `METHOD_NOT_ALLOWED`, `METHOD_UNAVAILABLE`, `RECIPIENT_OWNS_PRODUCT`, `RECIPIENT_LEVEL_TOO_HIGH`, `LEVEL_TOO_LOW`, `LEVEL_TOO_HIGH`, `GIFT_NOT_APPLICABLE`, `MESSAGE_TOO_LONG`, `MESSAGE_BLOCKED`, `GIFT_MESSAGES_OFF`, `INVALID_REQUEST_ID`, `REQUEST_ID_REUSED`, `GIFT_NOT_FOUND`, `GIFT_NOT_PENDING`, `GIFT_EXPIRED`, `GIFT_CANCEL_OFF`, plus the payment errors of `economy_purchase_start`.

### Console API (`/v2/console/economy/`, `HIRO_ECONOMY`)

`GET /friend-gifts` (filters `status`, `kind`, `source`, `outcome`, `userId` (either side), `senderId`, `recipientId`, `itemId`, `productId`, `flagged=true`, `q` on usernames, keys and message), `GET /friend-gifts/{id}` (row plus `trail` from the activity log), `POST /friend-gifts/{id}/reverse` (`GIFT_NOT_REVERSIBLE`, `GIFT_CHANGED`, `ITEM_GONE`, `INSUFFICIENT_BALANCE`, `INVALID_POLICY`). CRUD `friend-gift-rules` (`INVALID_NAME`, `INVALID_PRIORITY`, `INVALID_LIMIT`, `TAGS_REQUIRED`, `TAG_KIND_MISMATCH`, `TOO_MANY_TAGS`). `GET/PUT /friend-gift-settings` (`{settings, day}`; `INVALID_SETTING`, `INVALID_POLICY`, `LEVEL_TABLE_NOT_FOUND`, `ITEM_TYPE_NOT_FOUND`, `DUPLICATE_TYPE`, `CURRENCY_NOT_FOUND`, `DUPLICATE_CURRENCY`, `INVALID_AMOUNT`, `META_NOT_FOUND`). `GET /friend-gift-report?days=30` (counts by status and kind, senders, recipients, money revenue, coins spent, currency moved, top senders, recipients and gifts, per-day series). `GET /users/{id}/friend-gifts` (sent, received, pending, today's counts, the sender limits with the matching rule and bonus, last 10 gifts). Items and products take `giftable` (`true`, `false` or `null`).

### Console pages

Sidebar «فروش» gets «هدیه به دوست» (`#/economy/friend-gifts`) with tabs «گزارش هدیه‌ها» (log with filters; a row opens the gift with players, message, payment, trail and «برگرداندن هدیه» with where it goes and a note), «قانون‌ها» (`/rules`, tag rules editor), «تنظیمات» (`/settings`, every setting above) and «آمار» (`/report`). The item and product editors get «هدیه به دوست» (as default, giftable, not giftable). The player page gets a «هدیه به دوست» card: sent, received, waiting, today against the limit, the matching rule and bonus, the last gifts and a link to the log.

## Referral codes (phase 5)

asp has only a lazily made invite code and "inviter set once, not yourself" (`UserInvite`), with the reward given by a system event. Here the code, the rules against abuse, the goal and the rewards are all server-side. Migration `20261112000002-economy-referral.sql` (Up re-runnable, Down raises) adds the tables, seeds the settings row, and lets `economy_reward_claim.kind` and `economy_user_prize.source_type` be `referral`; Star Catcher sets the default prizes (100 coins for the inviter, 50 for the invited player) and the goal (one finished round) when the rule has none.

### Data

- `economy_referral_settings` (row 1): `is_enabled`, `new_account_days` (7; 0 = no limit), `referrer_prize_id`, `referee_prize_id`, `goal_event_type_id` (null = pay at once), `goal_measure` (`count`, `sum`, `max`), `goal_value`, `max_per_referrer` (100), `daily_per_referrer` (10; above it the referral is flagged `burst`), `one_per_device`, `one_per_phone`, `hold_flagged`.
- `economy_referral_code(user_id, code)`: 8 letters/digits without look-alikes, made on first use, unique.
- `economy_referral`: `referrer_id`, `referee_id` (unique: a player enters one code), `code`, `status` (`pending` = waiting for the goal, `held` = flagged and waiting for a decision, `completed`, `rejected`), `flags`, a copy of the prizes and the goal at entry time (a later rule change does not touch existing referrals), `ip`, decision fields.
- `economy_referral_mark(kind, hash)`: SHA-256 of every device id and phone number that already came with a code; not deleted with the account, so a deleted and recreated account cannot enter a code again with the same device or number.

### Entering a code (`economy_referral_enter {code}`)

The code is upper-cased with Persian/Arabic digits, spaces, dashes and ZWNJ removed. In one transaction under an advisory lock on the player: `CODE_NOT_FOUND` (also for a banned inviter), `SELF_REFERRAL` (own code, or the same phone as the inviter), `ALREADY_REFERRED` (409), `ACCOUNT_TOO_OLD` (403, account older than `new_account_days`), `REFERRAL_CYCLE` (the inviter was invited by this player), `REFERRER_LIMIT` (409, `max_per_referrer` non-rejected referrals), `DEVICE_ALREADY_USED` / `PHONE_ALREADY_USED` (409, when `one_per_*`; otherwise the referral is flagged `shared_device` / `shared_phone`), `REFERRAL_DISABLED` (403), `TOO_MANY_TRIES` (10 an hour, `clusterkit.Allow`). Flags: `burst`, `same_ip` (the inviter signed in from the same IP in the last 30 days, from the activity log), `ip_burst` (two or more referrals of the inviter from this IP in a day). A flagged referral is `held` when `hold_flagged` is on.

Without a goal (and not held) the referral is completed at once; otherwise it waits. Every recorded event (`recordEvent`, inside its transaction) checks whether the player has a pending referral whose goal is reached by the player's totals and completes it under a row lock. Completion inserts one claim per side into `economy_reward_claim` (`kind='referral'`, `ref_id` = referral id, scope `referrer` / `referee`, automatic), granted through the prize path with `source_id` = claim id, so each reward is paid once on any number of nodes; failed grants are retried by the existing claim worker. Activity rows: `entered` (referee), `joined` (referrer), `completed` (both), `approved` / `rejected`, plus `prize/granted` for each reward.

`economy_referral_me {lang?}` → `{code, enabled, canEnter, enterUntil, newAccountDays, referredBy: {username, status, createdAt}|null, counts: {total, completed, pending}, recent: [{username, status, createdAt}], rewards: {referrer, referee} (prize views), goal: {eventKey, eventName, measure, value}|null}`; `economy_referral_enter` returns the same plus `entered: {id, status, claims}`.

### Console (`/v2/console/economy/`, `HIRO_ECONOMY`)

`GET /referrals?q&status&flagged&flag&referrerId&refereeId` (with goal progress and claims), `GET /referrals/summary` (counts and top ten inviters), `GET /referrals/tree/{userId}` (up to five inviters up the chain and every player invited), `GET /referrals/{id}`, `POST /referrals/{id}/approve {note}` (held → completed, or pending when the goal is not reached yet; `NOT_HELD`), `POST /referrals/{id}/reject {note}` (`REFERRAL_DECIDED` once completed or rejected), `GET/PUT /referral-settings`.

Pages: «کدهای معرف» (`#/economy/referrals`): stat strip, top inviters with their tree, filters (status, suspicious, inviter), rows with flags, goal progress, «تأیید» / «رد» with a note and «درخت دعوت»; project settings «قوانین معرف» (rewards, goal event and value, new-account days, limits, one per device/phone, hold suspicious).

## Tasks and task groups (missions)

Port of the general task system of rakhtkan (`gamification/task`, `task_group`) with the claim-under-lock idea of asp `Gamification/Tasks` (TaskDefinition, TaskGroup, TaskEventListener, RevertClaimAsync). The card, deck and submission tasks of rakhtkan are game-specific and were not ported, and neither was rakhtkan's task `version` (an edited task keeps its player progress). Targeting by level and player tags is new; no source had it. Migration `20261114000000-economy-tasks.sql` (Up re-runnable, Down raises) adds the tables, lets `economy_reward_claim.kind` and `economy_user_prize.source_type` be `task`, and adds the key type `task` to `base_category` and `base_tag` (console label «تسک» / "Task"; tag pickers show `task` + `general`).

### Decision: one engine, tasks next to event prizes

Tasks are not a second event engine. They run inside the same `recordEvent` transaction as event prizes, meta rules, streaks and referral goals (`applyTasks`, right after `applyEventRules` and before meta rules), so an event is checked once (type, server-only flag, value range, per-type and per-player rate limits) and counts for tasks exactly once, in the same commit as the raw event row. They reuse the calendar clock (day/week windows in the zone and week start of `economy_event_settings`), the claims table `economy_reward_claim` (kind `task`, `ref_id` = task, `scope` = task instance), the prize grant path (`source_type='task'`, `source_id` = claim id, so a claim pays once on any number of nodes) and the retry worker for automatic rewards.

What stays different: event prizes are invisible "when X happens, pay Y" rules; tasks are the player-facing layer with a lifecycle (`locked` → `active` → `completed` → `claimed`, plus `cooldown`, `done`), dependencies, sub-tasks, hand-out groups and targeting. Use an event prize when nothing needs to be shown; use a task when the player should see and collect it. Measures map as `count` = count, `sum_value` = sum, `high_value` = highest value (the event prize names `count`/`sum`/`max`).

A manual (`claim`) task does not put a row in `economy_reward_claim` until the player collects it, so the rewards dot and `economy_rewards_mine?status=ready` of earlier phases do not change; automatic tasks insert their claim at completion like event prizes.

### Tables

| Table | What |
|---|---|
| `economy_task` | `key` (immutable), translated `name`/`description`, `icon_url` (economy uploads), `category_id` (key type `task`), `tag_ids` (`task` or `general`); `condition` (`event` = an event moves it, `children` = done when every active sub-task is done); `event_type_id`, `metric` (`count`, `sum_value`, `high_value`), `target > 0`, optional `value_min`/`value_max` (events outside the range do not count); `repeat` (`once`, `daily`, `weekly` = a new instance every local day/week, `interval` = a new instance `repeat_minutes` after the last completion), `max_completions`; `parent_id` (one level of sub-tasks), `depends_on` (task ids, at most 10, no cycles; locked until each has been completed at least once), `group_id`; optional `prize_id`, `grant_mode` (`auto`, `claim`), `notify`; targeting `audience_tag_ids` (player tags), `level_table_id` + `min_level`/`max_level`; `starts_at`/`ends_at`, `is_active`, `sort` |
| `economy_task_group` | `key`, translated `name`/`description`, `icon_url`, `category_id`; `active_count` (1–20 tasks handed out at once), `order_mode` (`ordered` by the group order, `random`), `handout` (`auto` = the next task comes as soon as a place is free, `request` = when the player asks, `timer` = `handout_minutes` after a place frees; the very first hand-out is immediate), `is_repeatable` (start a new round when every task is done); the same targeting and window as tasks |
| `economy_user_task` | one row per task, player and instance `scope`: `once`, `d<date>`, `w<week start>`, `i<n>` (interval), `g<round>` (group task); `status` (`active`, `completed`, `claimed`), count, sum, max, `completed_at`, `completed_by` (`player`, `admin`), `claim_id`, `claimed_at`, last grant `error` |
| `economy_user_task_group` | per group and player: `cycle` (round), `handed` (task ids handed out this round, in order), `next_at` (timer) |

Rules the server enforces (errors in the console): a sub-task needs an event (`SUBTASK_NEEDS_EVENT`), its parent must be `children` (`PARENT_NOT_GROUPING`), no deeper nesting (`NESTED_SUBTASK`), no group (`SUBTASK_IN_GROUP`); a sub-task always follows its parent's instance, window and targeting, so its own repeat, window and targeting are cleared on save. A group task is `once` (`GROUP_TASK_REPEAT`): it runs once per round of its group. A task with sub-tasks cannot become an event task (`HAS_SUBTASKS`) or be deleted while it has them. Other errors: `INVALID_CONDITION`, `INVALID_METRIC`, `INVALID_TARGET`, `INVALID_RANGE`, `INVALID_REPEAT`, `INVALID_REPEAT_MINUTES`, `INVALID_MAX_COMPLETIONS`, `SELF_PARENT`, `SELF_DEPENDENCY`, `TOO_MANY_DEPENDENCIES`, `DEPENDENCY_CYCLE`, `LEVEL_TABLE_REQUIRED`, `INVALID_LEVEL`, `INVALID_LEVEL_RANGE`, `INVALID_ACTIVE_COUNT`, `INVALID_ORDER_MODE`, `INVALID_HANDOUT`, `INVALID_HANDOUT_MINUTES`, `KEY_IMMUTABLE`, `IN_USE` (a task with completions or claims, a group with players in it). Deleting a task removes it from other tasks' `depends_on`; deleting a group detaches its tasks.

### Progress

For every active event task of the event's type whose value is in range, `applyTasks` first decides which tasks are open for the player (active, in window, targeted, dependencies met, handed out if in a group, not in cooldown or at `max_completions`), then upserts the instance row with `ON CONFLICT … DO UPDATE … WHERE status='active'`. The upsert row lock serialises parallel events of the same player on any node; once the row is `completed` nothing more is counted. Deciding first means the event that finishes a task does not also count for a task it unlocks. Reaching the target sets `completed` (or `claimed` straight away when the task has no prize), creates the automatic claim, and checks the parent under a lock on the parent's instance row, so two sub-tasks finishing at the same moment complete the parent once. Level targeting reads `economy_user_level` (or the level the source meta already reaches). Daily and weekly windows are the same calendar windows as event prizes; the sweep removes old `active` daily/weekly rows (finished rows stay, they count for `max_completions` and dependencies).

After the commit, groups touched by a finished task are refilled, and tasks with `notify` send an in-app message through `internal/inapp` (title «مأموریت انجام شد», body with the task name, action `goto tasks`).

### Claims

`claimTask` (client RPC, the generic `economy_reward_claim` with a task claim id, and the console): the instance's stable `claim_id` is generated at completion. The claim row is written (committed) first, then a transaction locks it `FOR UPDATE`, grants the prize through `grant` and marks both the claim and the instance as granted/claimed. A second claim waits on the lock and gets `ALREADY_CLAIMED`. If the grant fails, the transaction rolls back and the claim row it created is deleted again (asp's RevertClaim), so the task stays `completed` and collectable; the error is kept on the instance. The lock is never held on a row that an event transaction or Nakama's wallet update needs: holding the instance row, or inserting a claim row (a key-share lock on `users`) while waiting for the wallet's `FOR UPDATE` on the same player, deadlocked across connections in the first version, and the tests now cover parallel claims and events.

### Client RPCs

| RPC | Input | Output |
|---|---|---|
| `economy_tasks_list` | `{lang?, category?}` | `{lang, timezone, items:[task], groups:[group], ready}`; tasks not in a group and visible to the player, sub-tasks nested in `children`; `ready` = collectable tasks |
| `economy_task_claim` | `{taskId? | key? | claimId?, lang?}` | `{lang, claimId, kind:'task', rewards, prize, levels, task}`; `TASK_NOT_COMPLETED`, `ALREADY_CLAIMED`, `TASK_NOT_FOUND`, `CLAIM_NOT_FOUND` |
| `economy_task_group_current` | `{key, lang?}` | `{lang, group, handed:[keys handed out by this call]}` (hands out due tasks for `auto` and `timer`) |
| `economy_task_group_next` | `{key, lang?}` | same; for `request` groups (and due `timer` groups): `CURRENT_TASK_MUST_BE_DONE`, `DURATION_NOT_ELAPSED`, `NO_TASK_AVAILABLE`, `TASK_GROUP_FINISHED`, `HANDOUT_AUTOMATIC` |

A task: `{id, key, name, description, iconUrl, categoryKey, tags, condition, eventKey, eventName, metric, valueMin?, valueMax?, target, value, progress (0–1), state (locked|active|completed|claimed|cooldown|done), claimable, claimId?, pending (older finished instances still to collect), grantMode, prize, repeat, repeatMinutes, maxCompletions, completions, scope, periodEndsAt, expiresAt, nextAt, dependsOn:[{id,key,name,done}], children, groupKey?, parentKey?}`. A group: `{id, key, name, description, iconUrl, categoryKey, cycle, activeCount, orderMode, handout, handoutMinutes, repeatable, total, done, remaining, tasks, nextAt, canRequest, finished, expiresAt}`. Text is translated by `lang` with the usual `internal/localized` fallback. `economy_event_record` answers also carry `tasks:[{taskId, key, scope, value, target, completed, status, claimId?, grantMode?}]`, which `stargame_finish` passes through.

### Console API (`/v2/console/economy/`, `HIRO_ECONOMY`)

CRUD `tasks` (filters `isActive`, `groupId`, `eventTypeId`, `categoryId`, `prizeId`, `parentId`, `repeat`, `level=top|sub`; list rows carry `children`, `players`, `completed`) and `task-groups` (filters `isActive`, `handout`, `categoryId`; `taskIds` in the body sets the member tasks and their order). Player tools: `GET /users/{id}/tasks?lang` (every task with its state and the reason it is hidden: `inactive`, `ended`, `notStarted`, `notTargeted`, `notHanded`, `parent`, `groupClosed`), `POST /users/{id}/tasks/{taskId}/reset` (deletes the latest unclaimed instance, its unpaid claim and its sub-task rows; `TASK_ALREADY_CLAIMED`, `TASK_NOT_STARTED`), `/complete` (`TASK_NOT_AVAILABLE`, `TASK_ALREADY_COMPLETED`), `/claim`, `POST /users/{id}/task-groups/{groupId}/next`; these are audit-logged as invoke. `GET /event-claims?kind=task` and the player progress card list task rewards with the task name. The summary counts `tasks` and `task-groups`.

### Console pages

Economy section after «روزهای پیاپی»: «تسک‌ها» (`#/economy/tasks`: search, filters for kind, group, event, category and status; a condition sentence such as «پایان یک دور» ۲ بار · هر روز; prize and mode; place: sub-task of, group, number of sub-tasks or dependencies; completions and players; cards at 390px) with the floating editor (condition type, event, count, target, value range, parent task, live sentence, repeat with wait and limit, group, dependencies, prize, mode, in-app note, player tags, level range, window), and «گروه‌های تسک» (`#/economy/task-groups`, a sentence such as هم‌زمان ۱ تسک · به ترتیب · تسک بعدی با درخواست بازیکن; editor with tasks picked in hand-out order). The native player page gets «تسک‌های بازیکن» after the activity card: every visible task with state and progress, sub-tasks indented, «از نو», «تمامش کن», «جایزه را بده», and each group with its round, progress, next time and «تسک بعدی را بده». Both pages are in Ctrl K. A field that a form hides now also drops its old error message.

### Entry fee timing (2026-11-21)

Room field `feeAt`: `seat` (default, stored as empty) or `start` (`INVALID_FEE_AT`). `seat` keeps today's behaviour: `Enter` takes the fee, `Leave` before the start and `Cancel` give it back. `start` (Mench, as the legacy MatchFunding): `Enter` and `economy_game_join` only check the balance (`INSUFFICIENT_BALANCE`, also with `overdraft` on), the entry stays `free`; `Start` takes the fee from every `free` entry (bots only with `botsPay`) in one transaction, with `overdraft` letting any wallet, bots included, go below zero, and marks them `paid`. A lobby leave or a cancel before the start moves no money; `Void` of a started match and `refunded` results in `Settle` give the fee back, once. Console: «زمان گرفتن ورودی» in the room play box.

### Action log retention (2026-11-21)

The game setting `actionLogDays` (integer in the game's `settings`; a game plugin lists it in its schema to show it in the console) says how long `economy_game_action` rows of that game's matches are kept: missing = 60 days (today's behaviour), 0 = forever (Mench, as the legacy match_actions table). The sweep deletes at most 5000 rows per pass and per retention.

### Entry requirements (phase 24)

Owner's request (2026-11-14, Ludo-style games): a room, or a whole game, can ask for items, player tags, a level and meta values, optionally using up a ticket.

- Data: `economy_game.requirements` (one set for the whole game, `20261129000000-economy-game-requirements.sql`) and `rooms[].requirements` (one set per room) hold `gamecatalog.Requirements`: `itemIds` with `itemsMode` `any` (default) or `all`, `consumeItems` (default off: owning is enough; on: each entry deletes one listed item, or all of them in `all` mode), `tagIds` with `tagsMode` `any`/`all`, `excludeTagIds`, `minLevel` (default level table; rooms keep their own `minLevel` and its `LEVEL_TOO_LOW` as before), `metas` `[{metaId, operator gt|gte|eq|lte|lt, value, consume}]` (a `consume` above 0 makes the meta a counted ticket: the player needs at least that much and it is taken on entry), `productIds` (products offered on a refusal; empty = found automatically from active products whose prize gives a missing item or adds a missing meta), `paths` (`matchmaking`, `bot`, `friend`; all = stored as none), `refund` `return` (default: tickets come back when the player leaves before the start or the match is cancelled, once) or `keep`, `lock` `show` (default: listed as locked with what is missing) or `hide` (left out of `economy_games` for players who do not qualify), `inherit` (rooms only, default true: the game's set also applies) and `message` (refusal text per language, default «برای ورود به این اتاق شرط‌هایش را نداری.»). An empty set is stored as nothing. Limits: 20 items, 20 tags each, 10 metas, 10 products. Errors carry the field `requirements.…` or `rooms[i].requirements.…`: `ITEM_NOT_FOUND`, `TAG_NOT_FOUND`, `TAG_KIND_MISMATCH`, `TAG_BOTH_WAYS`, `META_NOT_FOUND`, `PRODUCT_NOT_FOUND`, `INVALID_OPERATOR`, `INVALID_VALUE`, `INVALID_CONSUME`, `INVALID_PATH`, `INVALID_REQUIREMENT`, `TOO_MANY_REQUIREMENTS`, `STARTER_TICKET` (starter items cannot be used up). Items, metas and products named in a set cannot be deleted (`IN_USE`); a deleted player tag simply never matches.
- Paths: `economy_game_join {…, path}` (default `matchmaking`) and the matchmaker hook (always `matchmaking`) check the sets of that path; `gamecatalog.Entry.Path` tells `Enter` how the player came in (`""` = matchmaking), so a bot match or a friend room passes `bot` or `friend`. `gamecatalog.Check(ctx, Entry)` answers the same refusal without entering; a fork's bot or friend-room RPC calls it before it makes the match. Providers that do not implement `gamecatalog.Checker` answer nil.
- Refusal: `REQUIREMENTS_NOT_MET` (HTTP 403, gRPC 7 in the matchmaker hook) with `meta {game, room, path, message, missing, products}`. `missing` items: `{type: level, minLevel, level}`, `{type: item, mode, items: [{key, name, iconUrl}]}` (in `all` mode only the lacking ones), `{type: tag, mode, tags: [{key, name}]}`, `{type: excludedTag, tags}` (only public tag names are shown), `{type: meta, key, name, operator, value, have, consume}`. `products`: `[{id, key, name, iconUrl, priceToman, coinPrice, satisfies: ["item:key", "meta:key"]}]`.
- Tickets: spent in `Enter` inside the entry transaction after the fee, under the player's item and meta row locks, so parallel joins spend once; a rejoin is free. The entry keeps `path` and `tickets` (`{items: [{id, key, source, sourceRef, upgradeLevel}], metas: [{id, key, amount}], refund, returned}`); `Leave` before the start and `Cancel` give them back once (items with their old source, starter → `game`). Activity rows `item/game_ticket` and `item/game_ticket_returned`. `EntryResult.tickets` is `{items, metas, refund}`; `GET /game-matches/{id}` entries show `path` and `tickets`.
- `economy_games {…, path}` / `economy_game`: each room adds `requirements` (its own set: `items {mode, consume, list: [{key, name, iconUrl, owned}]}`, `tags {mode, list: [{key, name, has}]}`, `minLevel`, `metas [{key, name, operator, value, consume, have}]`, `excludesTags`, `paths`, `consumes`, `refund`), `inheritsGame`, `eligible` (for the asked path), `eligibleBy {matchmaking, bot, friend}`, `missing`, `products` (when not eligible), `consumes`, `levelLocked` and `locked` (level or requirements). The game adds `requirements`, `eligible`, `eligibleBy`, `missing`, `products`, `consumes` and `path` for its own set. A game whose own set hides it, or whose rooms are all hidden, is left out.
- Console: the game editor has «شرط ورود» for the whole game and each room row has «شرط ورود به اتاق»: a folded box with a one-line summary («۱ آیتم (یکی کافی است) · بلیت مصرف می‌شود · حریف‌یابی، اتاق دوستانه») and «ویرایش شرط‌ها» that opens the nested form (item and tag pickers with quick-create, any/all, ticket switch, excluded tags, lowest level for the game, meta rows with «مصرف با ورود», offered products, paths, locked or hidden, ticket refund, inherit, refusal message per language). Room chips in the list get «· شرط‌دار». fa and en, RTL, 390px.
- Star Catcher lobby shows «لازم: …» under a locked room and «یک بلیت مصرف می‌شود» for ticket rooms.

### Friend rooms and bot matches (2026-11-18)

Any realtime game (registered in `internal/gamecatalog` with `Realtime: true` and a Nakama match handler under the game key) gets private rooms and bot matches without its own RPCs. The match handler reads its params with `internal/friendroom.FromParams` and follows the convention below; Soltan's `soltan_duel` is the first user (flickkick keeps its own RPCs).

- `economy_game_friend_create {key, room?, friendId?, lang?, platform?, version?}` → `{matchId, code, key, room, friendId, existing, expiresInSeconds}`. Checks the game like `economy_game_join` with path `friend` (fee balance, level, requirements). With `friendId` (a mutual friend) only that player may take the seat and gets a persistent notification `game_invite` (`{game, matchId, code, room, hostId, hostName, expiresAt}`, code = game setting `inviteNotifyCode`, default 2310); without it the first player with the code joins. One waiting room per host and game: a second call answers it (`existing: true`), or cancels it first when it was for another friend. Errors: `GAME_KEY_REQUIRED`, `GAME_NOT_REALTIME`, `FRIEND_ROOMS_DISABLED`, `CANNOT_INVITE_SELF`, `NOT_FRIENDS`, `MATCH_CREATE_FAILED` + the join check errors.
- `economy_game_friend_join {code}` or `{matchId}` → `{matchId, key, room, hostId, expiresInSeconds}`; the client then joins the match on the socket. `INVALID_CODE`, `ROOM_NOT_FOUND`, `OWN_ROOM`, `NOT_INVITED`, `CODE_JOIN_DISABLED` + join check errors (path `friend`).
- `economy_game_friend_cancel {matchId}` (host) / `economy_game_friend_decline {matchId}` (invited player) → `{matchId, done}`; `NOT_CANCELLABLE` / `NOT_DECLINABLE` (409, `meta.reason`) once the match started or for anyone else.
- `economy_game_bot {key, room?, …}` → `{matchId, key, room, botAfterSeconds}` when the game has `botsAllowed`; path `bot`. `BOTS_DISABLED`.
- Game settings read (add them to the game's schema to show them in the console): `friendRooms` (default true), `friendWaitSeconds` (120), `friendCodeJoin` (true), `inviteNotifyCode` (2310), `botAfterSeconds` (20, the client waits this long in the matchmaker before it asks for a bot).
- Match convention: params `{game, room, mode: matchmaking|friend|bot, host, code, invited, users, expiresAt (unix ms)}` (matchmaker matches send `{game, room, users}`); `Params.MayJoin(userId)`; label from `Params.Label(status, extra)` with `mode`, `status` (`waiting`, `playing`, `finished`), `open`, and for friend rooms `host`, `code`, `invited`, `expiresAt` (unix seconds); `MatchSignal` data `{"action":"cancel"|"decline","userId"}` answered with `friendroom.Answer(ok, reason)`.

### Star Catcher

The seed adds the task categories `star_missions_daily` and `star_missions_story`, daily «امروز دو دور بازی کن» (20 coins) and «پنج ستارهٔ طلایی بگیر» (15), weekly «این هفته ۱۰۰۰ امتیاز جمع کن» (100), «شروع ستاره‌چین» with three sub-tasks (one round, three golden stars, a round of 100+; 200 coins, in-app note), «کهنه‌کار: ده دور بازی کن» (after the starter, from level 2; 100) and the group «زنجیرهٔ شکارچی» (three tasks in order, automatic hand-out, 10/15/30 coins). Every seeded task is collected with a button, so no earlier balance changes.

## Items, inventory and product rules (phase 8)

Port of asp `Economy/Inventory` (InventoryItem, PlayerItem, ServiceRequest, InventoryPlayerService, InventoryUseService) and the shop rules of asp `Economy/Shop` (ProductShop MinLevel/MaxLevel/IsOneTime/CheckContents/PurchaseMessages, PlayerShopService Owned/LevelLocked, PaymentProductResolver checks). Migration `20261115000000-economy-items.sql` (Up re-runnable, Down raises) adds the tables, the key types `item` (categories and tags of items) and `item_type` (categories used as item types; seeded `avatar`, `frame`, `skin`, `board`, `piece`, `ball`, `trail`, `emote`, `other`, editable in project settings «دسته‌بندی‌ها»), five server-only built-in event types (`item_bought` value = price, `item_unlocked` value = level, `item_upgraded` value = new upgrade level, `item_equipped`, `item_used` value = units) and the product rule columns.

### Decisions

- Item types are base categories of key type `item_type`, not an enum, so a game adds its own types without a release. The type is also the equip slot: a player equips at most one item per type.
- Ownership is one row per player and item (`economy_user_item`, primary key `(user_id, item_id)`); items are not stackable. Every grant is `INSERT … ON CONFLICT DO NOTHING`, so a second grant of an owned item is a no-op, never a duplicate.
- The ownership row has no foreign key to `users`: inserting a referencing row takes a key-share lock on the player row, and Nakama's wallet update takes `FOR UPDATE` on the same row from its own connection, which deadlocked the task claims before (see «Claims» in tasks). Rows of deleted players are removed by the event sweep (200 players per pass) and by the e2e cleanup.
- Buying, upgrading and services debit through Nakama's wallet with a deterministic `grantId` in the ledger metadata (UUIDv5 of player and `requestId`). The ownership/upgrade row is written (and locked) in the plugin transaction *before* the debit, so a parallel second buy waits on that row and gets `ALREADY_OWNED` without paying. If the transaction fails after the wallet committed, a retry with the same `requestId` finds the ledger row and finishes without a second debit. Without a `requestId` the server makes one, so the call is still safe against double charging (ownership is unique) but cannot be replayed.
- `requestId` results are kept 24 hours in `economy_item_request (user_id, request_id)`; the same id for another action, item or detail answers `REQUEST_ID_REUSED`. Required for services (`economy_item_use`), optional for buy and upgrade.
- Starter items are granted lazily on the first economy call of a player (`ensureDefaults`, in its own small transaction, `ON CONFLICT DO NOTHING`), and the first starter of each type is equipped when that slot is empty. Starters cannot be taken back in the console (`STARTER_ITEM`), so they never come back after a revoke.
- Unlock by level: an item with `unlock_level` and no price is opened free with `economy_item_unlock` once the player reaches the level; an item with both is buyable from that level (`LEVEL_TOO_LOW` before). asp's "pay to unlock early" was not ported.
- Level everywhere (items, product rules) comes from the level tables: the item's or product's `level_table_id`, else the first active table by `sort`; the reached level is the higher of `economy_user_level` and the level the source meta already reaches (same rule as tasks). Without any level table the old `level` meta is read, so older setups keep working. This replaces the hard-coded `level` meta read of `economy_purchase_start`.
- Not copied from asp: the non-atomic claim, find-then-insert dedup (here unique keys decide), currency or meta writes from the client, the prize "send gift" transfer, the avatar buy outside a transaction, the best-effort ledger (here the wallet ledger is the record and the request id makes it idempotent) and enum-only types.

### Tables

| Table | What |
|---|---|
| `economy_item` | `key` (unique, immutable after create: `KEY_IMMUTABLE`), `type_id` (category of key type `item_type`), translated `name`/`description`, `icon_url`, `image_url, sound_url (a sound such as a goal sound, mp3/ogg/wav upload)` (economy uploads), `category_id` (key type `item`), `tag_ids` (`item` or `general`), `rarity` (`common`, `uncommon`, `rare`, `epic`, `legendary`), `parent_key` / `requires_key` (another item that must be owned first), `is_starter`, `level_table_id` + `unlock_level`, `price` `{currencyKey, amount}` (null = not for sale), `upgrades` `[{currencyKey, amount}]` (the price of level 1, 2, …; the length is the max upgrade level, at most 50), `services` `[{key, name, currencyKey, unitPrice, maxUnits}]` (at most 20), `metadata`, `sort`, `is_active` |
| `economy_user_item` | `user_id`, `item_id`, `source` (`starter`, `buy`, `unlock`, `prize`, `admin`), `source_ref` (grant row, request grant id, level or admin name), `owned_at`, `upgrade_level`, `equipped`, `slot` (type key), `version`. Unique `(user_id, slot) WHERE equipped` |
| `economy_item_request` | `(user_id, request_id)`, action, item key, detail, stored result, `expires_at` (24 h; swept with the events) |
| `economy_product` new columns | `max_level` (null = no limit), `level_table_id`, `is_one_time`, `check_contents` (default true), `purchase_message` (translated) |
| `economy_payment.once_key` | `user:product` for one-time products, unique where set |

### Prize item type `item`

`{type:'item', key, action:'give'}` (value is stored as `1`). It is granted inside the one `grant()` transaction of every prize source (payments, gifts, tasks, event prizes, streaks, level steps, referrals, admin grants), before the wallet part, with `INSERT … ON CONFLICT DO NOTHING` and `source_ref` = the grant row, so the existing claim / `once_key` design keeps every grant single. A prize with an unknown item is refused (`ITEM_NOT_FOUND`), and an item used in a prize cannot be deleted (`IN_USE`). `storage` items keep working unchanged.

### Product rules

- `min_level` / `max_level` against the level above. `economy_catalog` hides a product when the player is above `max_level` and marks `levelLocked` below `min_level`; `economy_purchase_start` answers `LEVEL_TOO_LOW` / `LEVEL_TOO_HIGH`. Levels are checked only when a purchase starts: a payment that is already paid is delivered even if the level changed meanwhile.
- `is_one_time`: the catalog hides it once the player has a successful purchase; start answers `PRODUCT_ALREADY_BOUGHT`. Settlement guard: the success transition sets `once_key` in the same `UPDATE … WHERE status='pending'`; when two store or gateway payments of the same player and one-time product settle at the same time, the unique index lets only one win. The loser stays `success` (the money was taken) with `outcome='duplicate'`, `reward_error='PRODUCT_ALREADY_BOUGHT'` and no grant, so it shows in «پرداخت‌ها» as a payment to refund; `retry-grant` refuses it. Coin (`in_app`) payments take the once key at insert, before the debit, and release it if they fail.
- `check_contents` (default on; it only matters when the prize gives items): start answers `PRIZE_ALREADY_OWNED` when the player owns every item of the prize and `REQUIRES_ITEM` when an item's `requires_key`/`parent_key` is missing and not in the same prize. At settlement the grant refuses with `PRIZE_ALREADY_OWNED` when none of the items could be added; for coin purchases this happens before the debit (the payment fails, no coins move), for paid payments it is kept as `reward_error` for a refund.
- `purchase_message`: translated, returned as `purchaseMessage` by the catalog, by a successful coin purchase and by `economy_purchase_status` once delivered. `economy_purchase_start` and `economy_purchase_status` accept `lang`.
- Purchase counts are successful payments that are not duplicates.

### Client RPCs

| RPC | Input | Output |
|---|---|---|
| `economy_inventory` | `{type?, lang?}` | `{lang, level, counts:{owned,total}, equipped:{slot:key}, items:[{id, key, type, typeName, name, description, iconUrl, imageUrl, soundUrl, rarity, categoryKey, tags, parentKey, requiresKey, isStarter, unlockLevel, price, maxUpgradeLevel, services, metadata, sort, isActive, owned, equipped, locked, playerLevel, requirementMet, canBuy, canUnlock, upgradeLevel, nextUpgradePrice, source?, ownedAt?, version?}]}` (active items) |
| `economy_item_buy` | `{key, requestId?, lang?}` | `{lang, item, paid:[{currencyKey,amount}], wallet, replayed}`; `ITEM_NOT_FOUND`, `NOT_FOR_SALE`, `LEVEL_TOO_LOW`, `REQUIRES_ITEM`, `ALREADY_OWNED`, `INSUFFICIENT_BALANCE`, `REQUEST_ID_REUSED` |
| `economy_item_unlock` | `{key, lang?}` | same; `NOT_UNLOCKABLE`, `LEVEL_TOO_LOW`, `REQUIRES_ITEM`, `ALREADY_OWNED` |
| `economy_item_upgrade` | `{key, fromLevel?, requestId?, lang?}` | same plus `level`; `NOT_OWNED`, `UPGRADE_CHANGED` (fromLevel differs), `UPGRADE_MAXED` |
| `economy_item_equip` / `economy_item_unequip` | `{key, lang?}` | `{lang, slot, item, unequipped:[keys]}`; `NOT_OWNED`. All owned rows of the type are locked in id order, so parallel equips leave exactly one equipped |
| `economy_item_use` | `{key, service, units?, requestId, lang?}` | same plus `service`, `units`; `REQUEST_ID_REQUIRED`, `SERVICE_NOT_FOUND`, `INVALID_UNITS`, `NOT_OWNED` |
| `economy_purchases_mine` | `{}` | `{items:[{productId, key, count, lastPurchasedAt}]}` |

`economy_catalog` products now also carry `maxLevel`, `isOneTime`, `owned`, `levelLocked`, `bought`, `purchaseCount`, `requiresKey`, `purchaseMessage` and `items` (item keys of the prize). Buy, unlock, upgrade and grant/revoke write player activity rows of the new kind `item` (`bought`, `unlocked`, `upgraded`, `granted`, `revoked`, `starter`).

### Console API (`/v2/console/economy/`, `HIRO_ECONOMY`)

CRUD `items` through the generic resource handlers (search on key and every name translation; filters `typeId`, `categoryId`, `isActive`, `isStarter`, `rarity`, `forSale`; rows carry `owners` and `equippedBy`). Errors: `ITEM_TYPE_NOT_FOUND`, `CATEGORY_KIND_MISMATCH` (field `typeId`), `INVALID_RARITY`, `SELF_REFERENCE`, `ITEM_NOT_FOUND`, `STARTER_HAS_LEVEL`, `INVALID_LEVEL`, `LEVEL_TABLE_NOT_FOUND`, `CURRENCY_NOT_FOUND`, `INVALID_PRICE`, `TOO_MANY_UPGRADES`, `TOO_MANY_SERVICES`, `DUPLICATE_SERVICE`, `INVALID_UNITS`, `KEY_IMMUTABLE`, `IN_USE` (owned, in a prize, or required by another item). Products: `INVALID_LEVEL_RANGE`, `LEVEL_TABLE_NOT_FOUND`, `ONE_TIME_SUBSCRIPTION`. Player tools: `GET /users/{id}/items?lang&type` (also inactive items the player owns), `POST /users/{id}/items {itemId}` (source `admin`, `ALREADY_OWNED`), `DELETE /users/{id}/items/{itemId}` (`NOT_OWNED`, `STARTER_ITEM`); both go to the console audit log and the player activity. The summary counts `items` and `ownedItems`.

### Console pages

Economy section after «فروشگاه»: «آیتم‌ها» (`#/economy/items`, Phosphor backpack icon): search, filters for type, category, rarity and status; rows with type and rarity chips, how a player gets it (starter, price, from level N, prize only, required item, upgrade levels) and owners; cards at 390px. The docked editor has type (with the «+» quick-create of an item type), rarity, category and tags, «همه از اول دارند», price, unlock level and table, required and base item, an upgrade row editor (currency and price per level), a service row editor, icon and large image. The prize item editor has the new type «آیتم» with an item picker. The product editor has a «قانون‌های خرید» block (min and max level, level table, «فقط یک بار خریدنی», «اگر بازیکن همه را دارد، نفروش», «پیام بعد از خرید») and the product list a rules column. The native player page gets «انبار بازیکن» after the tasks card: owned items with type, rarity, equipped mark, source and upgrade level, the equipped item per type, «پس گرفتن» with a confirmation and «دادن آیتم» with an item picker. The page is in Ctrl K.

### Star Catcher

The seed adds four items: «سبد چوبی» (skin, starter), «سبد یاقوتی» (skin, rare, sold only by the product `star_item_ruby` for 60 coins with `check_contents` on), «سبد یخی» (skin, epic, 120 coins straight from the inventory) and «قاب طلایی» (frame, free from level 2 of `star_levels`); and the one-time product «بستهٔ شروع» (`star_starter_pack`, 150 coins for 30 coins or 5000 toman, up to level 2, with a purchase message). The «انبار» screen lists the items by type with rarity, owned and equipped marks, buy, unlock and equip buttons; the equipped skin changes the basket in the game (`metadata.art`).

## Achievements (phase 9)

Port of asp `Economy/Resource/Achievement` (Achievement `Name`, `IconId`, `Key`, `TaskId`; `UserAchievement` unique per player; granted only by a prize item of type achievement or by an admin; `GET /client/achievements`) and its admin pages under `frontend/src/features/achievements`. No source had a rule engine, tiers or progress; they are added here on top of the existing engines. Migration `20261116000000-economy-achievements.sql` (Up re-runnable, Down raises) adds the tables, the key type `achievement` to `base_category` and `base_tag` (console label «دستاورد» / "Achievement"; tag pickers show `achievement` + `general`), lets `economy_reward_claim.kind` and `economy_user_prize.source_type` be `achievement`, and seeds the server-only built-in event type `achievement_unlocked`.

### Decisions

- No third engine. An achievement unlocks in one of four ways: (a) `source = 'event'`: a condition on the event pipeline with the task condition model (event type, `count` / `sum_value` / `high_value`, optional value range, all-time window only), counted in the same `recordEvent` transaction right after tasks, so an event counts once and the type, server-only flag, value range and rate limits apply; (b) `source = 'task'`: a linked task (`task_id`) or task group (`task_group_id`, done when every active top-level task of a round has a finished instance), checked inside the task completion code, so it works for events, sub-task parents and the console «تمامش کن»; (c) the prize item `{type:'achievement', key, action:'give'}` granted inside the one `grant()` transaction, so products, gifts, event prizes, streaks, levels, tasks, referrals and admin grants can give achievements; (d) admin give and revoke from the console. Every achievement can be given by (c) and (d), whatever its source; `source = 'manual'` means only by (c) and (d).
- Tiers: every achievement has 1–10 ordered tiers (`economy_achievement_tier`: `tier`, `threshold`, translated `name`, `icon_url`, `prize_id`, `grant_mode`). Several tiers need `source = 'event'` (`TIERS_NEED_EVENT`); task and manual achievements get one tier (made by the server when the body has none). Unlocked means tier 1 reached. Thresholds go strictly up (`THRESHOLDS_NOT_ORDERED`), tiers are numbered 1, 2, 3 … (`TIERS_NOT_CONSECUTIVE`). Tiers are upserted by number, so claim references stay; removing a tier that has claims is refused (`TIER_IN_USE`).
- One row per player and achievement (`economy_user_achievement`, primary key `(achievement_id, user_id)`) keeps the counters (`count`, `sum`, `max`), the reached `tier`, `unlocked_at`, `tier_at` and how it was unlocked (`source`: `progress`, `event`, `task`, `prize`, `admin`; `source_ref`: event id, task key, grant row or admin name). The reached tier never goes down by events; an admin give raises it.
- The player row has no foreign key to `users`, like `economy_user_item`, because the prize path writes it inside the grant transaction. In that transaction the achievement row is written after the wallet update, so the grant never holds an achievement row lock while it waits for Nakama's `FOR UPDATE` on the player (the deadlock that tasks and items hit before). Rows of deleted players are removed by the event sweep (200 players per pass) and by the e2e cleanup.
- Tier prizes are paid through `economy_reward_claim` (kind `achievement`, `owner_id` = achievement, `ref_id` = tier, scope `t<tier>`) with a deterministic claim id (UUIDv5 of player and tier), and through the prize path with `source_type='achievement'`, `source_id` = claim id. A tier pays once per player on any number of nodes, even after a revoke and a second unlock. `auto` tiers insert their claim right after the commit and grant it; failed ones are retried by the claim worker, and the list RPC re-creates a missing auto claim (a crash between commit and claim). `claim` tiers put no row in the claims table until the player collects them (like manual tasks): the claim row is committed first, then locked `FOR UPDATE` while the prize is granted; if the grant fails the row is deleted again.
- On unlock (once per achievement) the server records `achievement_unlocked` (value = the achievement's points, metadata `{key}`), so tasks, event prizes and other achievements can count unlocks. Cascades stop at depth 6, as for events.
- `is_hidden`: a locked hidden achievement shows only `{id, hidden, state:'hidden', rarity, points, hint, lockedIconUrl, categoryKey}`; once unlocked it is shown in full. Public profiles list only unlocked achievements, so a secret stays secret.
- `notify`: an in-app message through `internal/inapp` on unlock and on every new tier (title «دستاورد تازه», action `goto achievements`).
- Revoke deletes the player row (progress and tier) and the player's unpaid claims of that achievement. Prizes already paid are not taken back and are not paid again.
- Not copied from asp: the client-writable unlocks, the missing uniqueness under parallel grants (here the primary key and the claim key decide) and the task link as the only rule.

### Tables

| Table | What |
|---|---|
| `economy_achievement` | `key` (unique, immutable: `KEY_IMMUTABLE`), translated `name`, `description`, `hint`, `icon_url`, `locked_icon_url` (economy uploads), `category_id` (key type `achievement`), `tag_ids` (`achievement` or `general`), `rarity` (`common` … `legendary`), `points` (0–100000), `is_hidden`, `source` (`event`, `task`, `manual`), `event_type_id`, `metric`, `value_min`/`value_max`, `task_id` or `task_group_id` (not both), `notify`, `sort`, `is_active` |
| `economy_achievement_tier` | see above; unique `(achievement_id, tier)` |
| `economy_user_achievement` | see above |

### Client RPCs

| RPC | Input | Output |
|---|---|---|
| `economy_achievements` | `{lang?, category?}` | `{lang, points, unlocked, total, ready, items:[{id, key, name, description, hint, iconUrl, lockedIconUrl, displayIconUrl, categoryKey, tags, rarity, points, hidden, source, metric, eventKey?, eventName?, taskKey?, groupKey?, valueMin?, valueMax?, state (locked|unlocked|completed|hidden), unlocked, tier, maxTier, tierName, tiered, value, progress (0–1 toward the next tier), next:{tier, threshold, name}|null, tiers:[{tier, name, threshold, iconUrl, grantMode, prize?, status (locked|reached|claimable|pending|claimed), claimId?}], claimable:[tier], unlockedAt?, tierAt?}]}` (active achievements; `ready` = tiers to collect) |
| `economy_achievement_claim` | `{key? \| id?, tier?, lang?}` | `{lang, kind:'achievement', claims:[{tier, claimId, rewards, prize}], achievement, levels}`; without `tier` every collectable tier is paid. `ACHIEVEMENT_NOT_FOUND`, `ACHIEVEMENT_LOCKED`, `TIER_NOT_FOUND`, `TIER_NOT_REACHED`, `TIER_HAS_NO_PRIZE`, `ALREADY_CLAIMED` (409), `NOTHING_TO_CLAIM` |
| `economy_achievements_profile` | `{userId? \| username?, lang?}` (empty = me) | `{lang, user:{userId, username, displayName, avatarUrl, self}, points, unlocked, total, items:[{key, name, description, iconUrl, rarity, points, hidden, tier, maxTier, tierName, categoryKey, unlockedAt}]}`; `total` counts visible achievements plus unlocked hidden ones; disabled accounts answer `USER_NOT_FOUND` |

`economy_event_record` answers (and `stargame_finish`) carry `achievements:[{achievementId, key, from, to, value, unlocked, source}]`; auto tier claims appear in `claims`. `economy_rewards_mine` and the console claims list name the achievement and the tier. Activity rows of the new kind `achievement`: `unlocked`, `tier`, `revoked`.

### Console API (`/v2/console/economy/`, `HIRO_ECONOMY`)

CRUD `achievements` through the generic resource handlers (search on key and every name translation; filters `isActive`, `source`, `categoryId`, `rarity`, `isHidden`, `eventTypeId`, `taskId`, `taskGroupId`, `tiered`; rows carry `tiers` with `players` per tier, `unlocked`, `players`). Errors: `INVALID_SOURCE`, `EVENT_TYPE_NOT_FOUND`, `INVALID_METRIC`, `INVALID_RANGE`, `TASK_REQUIRED`, `TASK_AND_GROUP`, `TASK_NOT_FOUND`, `TASK_GROUP_NOT_FOUND`, `TIERS_REQUIRED`, `TOO_MANY_TIERS`, `TIERS_NEED_EVENT`, `INVALID_TIER`, `DUPLICATE_TIER`, `TIERS_NOT_CONSECUTIVE`, `INVALID_THRESHOLD`, `THRESHOLDS_NOT_ORDERED`, `TIER_IN_USE`, `INVALID_GRANT_MODE`, `PRIZE_NOT_FOUND`, `INVALID_RARITY`, `INVALID_POINTS`, `CATEGORY_KIND_MISMATCH`, `TAG_KIND_MISMATCH`, `KEY_IMMUTABLE`, `IN_USE` (unlocked by a player, with claims, or in a prize). A prize with an unknown achievement is refused (`ACHIEVEMENT_NOT_FOUND`); a task or task group an achievement waits for cannot be deleted (`IN_USE`). Player tools: `GET /users/{id}/achievements?lang` (also inactive ones, hidden ones in full, `unlockedBy`/`unlockedRef`), `POST /users/{id}/achievements {achievementId, tier?}` (`ALREADY_UNLOCKED`, `TIER_NOT_FOUND`), `DELETE /users/{id}/achievements/{achievementId}` (`NOT_UNLOCKED`), `POST /users/{id}/achievements/{achievementId}/claim {tier?}`; they go to the console audit log (create, delete, invoke) and the player activity. The summary counts `achievements` and `unlockedAchievements`.

### Console pages

Economy section after «آیتم‌ها»: «دستاوردها» (`#/economy/achievements`, Phosphor medal icon): search, filters for source, rarity, category, hidden and status; rows with the badge tile, a sentence such as «پایان یک دور» ۳ بار and the tier targets (برنزی: ۳، نقره‌ای: ۱۰، طلایی: ۳۰), tiers and first prize, rarity and points, players who unlocked it, status, hidden and source chips; cards at 390px. The docked editor has name, description, the locked hint, rarity, points, category and tags, badge and locked badge, «پنهان تا باز شدن», the source with the task condition fields (event, count, value range) or the task / task group pickers, a live sentence, the tier row editor (target, prize, mode, translated name; order checked) for event achievements and one unlock prize for the others, and the in-app note switch. The prize item editor has the new type «دستاورد» with an achievement picker. The native player page gets «دستاوردهای بازیکن» after the inventory card: every achievement with its badge, state, rarity, tier, a progress meter toward the next tier, who unlocked it and when, «جایزه را بده» for waiting tiers, «پس گرفتن» with a confirmation, «دادن دستاورد» with a picker and an optional tier, and a link to the player's achievement history in the activity log. The page is in Ctrl K; the plane sidebar then had 18 economy entries (20 since the chance game pages).

### Star Catcher

The seed adds the achievement categories `star_ach_play` and `star_ach_special` and five achievements: «اولین دور» (one round, 5 points), «ستاره‌چین پرکار» (rounds played, tiers برنزی ۳ / نقره‌ای ۱۰ / طلایی ۳۰ with 20 / 50 / 100 coins to collect), «شکارچی کامل» (the «زنجیرهٔ شکارچی» task group, 50 coins), «هوادار ستاره‌چین» (manual, given by the prize `star_fan_badge` of the gift code `STARFAN` with 20 coins, in-app note) and the hidden «شکار افسانه‌ای» (a 300-point round, 30 coins, hint «یک دور واقعاً بزرگ بازی کن.»). Every tier is collected with a button, so no balance changes on its own. Rows are matched by key and never updated; tiers are written only when the achievement row is created.

## Chance game (phase 11)

A wheel of fortune, loot box or slot machine: the player plays and wins one prize by chance, or nothing. The contract below is the owner's spec (2026-11-12), kept word for word; field names, error codes, HTTP-style status codes, response shapes and the cumulative pick are implemented as written. The rest of this section is what the spec did not say. Migration `20261118000000-economy-chance-games.sql` (Up re-runnable, Down raises) adds the tables, lets `economy_reward_claim.kind` and `economy_user_prize.source_type` be `chance`, and seeds the server-only built-in event type `ChanceGame:Played`. Reference behaviour: rakhtkan `app/modules/gamification/chance_games` (and `economy/prize_group`), without its holes.

### Owner's spec

```
Fields: id, name, description?, metadata JSON?, prizeType enum(prize, prizeGroup)=prize,
prize[]: {chance number 0-100, prizeId? | prizeGroupId?} (empty = nothing won),
intervalTime number (minutes)?, price?: {currencyId, amount}.

Wheel of fortune, loot box, slot machine: the player plays, wins one prize by chance (may be empty).
Sum of chances must be exactly 100 on create/update, else 400 TOTAL_CHANCE_MUST_BE_100 meta {totalChance}.

Admin: CRUD.
Client:
- intervalTime null → CanPlay, GetNextPlayTime, Play return 400 INTERVAL_TIME_IS_NULL.
- CanPlay(id): false if last recorded play is newer than intervalTime.
- GetNextPlayTime(id): intervalTime - (now - lastPlay), 0 if negative.
- Play(id): if waiting → 403 WAITING_FOR_TURN meta {nextPlayTime: intervalTime - (now - lastPlayDate)}; else random pick.
- PlayWithTicket(id): price null → 400 TICKET_PLAY_IS_NOT_SUPPORT; not enough currency → 409 CURRENCY_IS_NOT_ENOUGH; else deduct and play.
Pick: random number 0..100, walk prizes in order by cumulative chance (30/40/30: 25→1st, 50→2nd, 85→3rd).
Response 200 {code: SUCCESS, prizeId: id | null}.
Prize is recorded and claimed at once. Event "ChanceGame:Played", context entity id = chance game id, metadata null.
```

### Decisions

- **Pick.** Chances are stored with at most four decimals and compared as integers in units of 1/10000 (the sum must be exactly 1 000 000 units, so 33.3333 × 3 is refused with `totalChance` 99.9999). The roll is `crypto/rand` over `[0, 1 000 000)`, i.e. `0 ≤ roll < 100`; rows are walked in order and the first row whose running total is greater than the roll wins (30/40/30: 25 → row 1, 30 → row 2, 50 → row 2, 85 → row 3). A row with chance 0 can never win. A row without `prizeId`/`prizeGroupId` wins nothing. A prize that is not repeatable and that the player already has counts as nothing won for that play (the slot is still reported), so a free turn cannot be re-rolled by failing on an owned prize. The roll is stored with the play and returned as `roll` (0–100), so a client can land its wheel on the exact point the server drew.
- **Prize groups.** New `economy_prize_group` (key, translated name and description) with `economy_prize_group_member (group_id, prize_id, weight 1–1 000 000, sort)`, 1–100 prizes, each prize once. A group pays **one** member, drawn with `crypto/rand` by weight (all weights 1 = uniform, which is what rakhtkan's `get_random` did); paying every member was not chosen because a loot-box slot should stay one prize. The draw is a plain function (`drawFromPrizeGroup`) so other features can reuse groups later. An empty group cannot be saved (`PRIZE_GROUP_EMPTY`); a prize in a group or in a chance game cannot be deleted, and a group used by a chance game or a play cannot be deleted (`IN_USE`). rakhtkan cleaned dangling references silently on read; here references are refused at write time instead.
- **Turns.** `intervalTime` is whole minutes (1–525 600). The wait is counted from the player's last **free** play of that game; ticket plays do not move the free turn. `nextPlayTime` is **seconds**, rounded up (rakhtkan mixed minutes and seconds). `GetNextPlayTime`/`Play` also return `nextPlayAt` (RFC 3339) when there is a wait.
- **Race safety.** Every play first takes a row lock on `economy_chance_state (game_id, user_id)` (`INSERT … ON CONFLICT DO NOTHING` then `SELECT … FOR UPDATE`), reads the game `FOR SHARE`, checks the turn against the last free play, draws, and inserts the play row and its claim in the same transaction. Parallel `Play` calls for one player on any number of nodes therefore play at most once per interval (Postgres test: 16 parallel calls → 1 play, 15 `WAITING_FOR_TURN`). No player row has a foreign key to `users`.
- **Recorded and claimed at once.** A winning play inserts an `economy_reward_claim` (kind `chance`, `owner_id` = game, `ref_id` = play, scope `play`, a fresh claim id per play) in the decision transaction, then grants it right away through `grantClaim` → `grant()` (`source_type='chance'`, `source_id` = claim id). The play row is `pending` until the grant transaction marks it `done` (`After` hook), so the claim, the user-prize row and the play are one record. Any path that grants a `chance` claim (player claim RPC, console retry, the auto-claim worker) goes through the same hook, so it can never pay a ticket prize without taking the ticket.
- **Ticket spend + prize in one grant.** `PlayWithTicket` puts the price into `grant()`'s `Debit`; `grant()` now sends the debit as its own wallet update *before* the prize update in the same `nk.MultiUpdate` call (one Nakama transaction), so a prize of the same currency can never pay for its own ticket (Postgres test: 10 coins, price 10, prize 5, six parallel plays → exactly one play, 5 coins left). This also fixes the same net-changeset hole for coin (`in_app`) purchases, which now write two ledger rows (`part: debit` and `part: prize`). A ticket play that wins nothing debits through the item wallet helper with the play id as `grantId`. Both are idempotent by `grantId`; a short balance is checked first (`409 CURRENCY_IS_NOT_ENOUGH` with `meta {currencyKey, amount, balance}`) and again by the wallet itself. If the grant fails and the ledger shows no `chance` row for the play, the play, its claim and its user-prize row are deleted (no free prize, no lost coins, the free turn is given back). Plays left `pending` by a crash are finished or cancelled by `recoverChancePlays` on every node (every 30 s, `FOR UPDATE SKIP LOCKED`, after two minutes, at most 8 attempts).
- **Event.** After a play is `done` the server records `ChanceGame:Played` (value 1, server only) through `recordEvent`, so tasks, event prizes, achievements and streaks can count chance games. Nakama events have no separate context, so the entity id goes into the event metadata: `{entityId: <game id>, mode: "free"|"ticket", playId}`. It is recorded once per play (`economy_chance_play.event_at` guard). Built-in event keys may have the form `Word:Word` (`builtinEventKeyPattern`); admins still create keys with the lower-case pattern.
- **Not copied from rakhtkan:** the unlocked interval check, `random.uniform`, currency spend and prize grant in separate transactions, ticket plays without a recorded event, the silent cleanup of stale references, seconds vs minutes, and `WAITING_FOR_TURN` as 400.

### Status codes and Nakama RPC errors

Client RPCs answer like every economy RPC: the HTTP-style status of the spec is mapped onto a gRPC code with `grpcCode` (`errors.go`), and Nakama's HTTP gateway turns that into the HTTP status the client sees. Because the spec carries `meta`, every chance RPC error message is JSON (the same convention as the phase 3 suspension notice): `{"code":"WAITING_FOR_TURN","status":403,"meta":{"nextPlayTime":3540}}`; `status` is the spec's status even where the gateway sends another one.

| Spec | Code | gRPC code | HTTP from Nakama |
|---|---|---|---|
| 400 | `INTERVAL_TIME_IS_NULL`, `TICKET_PLAY_IS_NOT_SUPPORT`, `INVALID_JSON` | 3 `InvalidArgument` | 400 |
| 403 | `WAITING_FOR_TURN` (`meta.nextPlayTime` seconds) | 7 `PermissionDenied` | 403 |
| 404 | `CHANCE_GAME_NOT_FOUND` (unknown or inactive) | 5 `NotFound` | 404 |
| 409 | `CURRENCY_IS_NOT_ENOUGH` (`meta {currencyKey, amount, balance}`) | 9 `FailedPrecondition` (as `INSUFFICIENT_BALANCE`, `ALREADY_CLAIMED`) | 400 |

The console API answers with the real HTTP status and `{"error", "message", "field", "meta"}`, e.g. `400 {"error":"TOTAL_CHANCE_MUST_BE_100","field":"prize","meta":{"totalChance":90}}`.

### Tables

| Table | What |
|---|---|
| `economy_chance_game` | `key` (unique, immutable: `KEY_IMMUTABLE`), translated `name` / `description`, `metadata` (object), `prize_type` (`prize`, `prizeGroup`), `prize` (jsonb `[{chance, prizeId, prizeGroupId}]`, 1–50 rows), `interval_time` (minutes or null), `price` (jsonb `{currencyId, amount}` or null), `is_active`, `sort` |
| `economy_prize_group`, `economy_prize_group_member` | see above |
| `economy_chance_state` | `(game_id, user_id)` lock row with `plays` and `last_play_at` |
| `economy_chance_play` | one row per play: `mode` (`free`, `ticket`), `status` (`pending`, `done`), `roll` (0–999 999), `slot`, `prize_id`, `prize_group_id`, `claim_id`, `price` (`{currencyKey, amount}` taken), `error`, `attempts`, `created_at`, `finished_at`, `event_at` |

Save errors (400 unless noted, `field` points at the row as `prize[i]` / `members[i]`): `TOTAL_CHANCE_MUST_BE_100` (`meta.totalChance`), `PRIZE_LIST_REQUIRED`, `TOO_MANY_PRIZES`, `INVALID_PRIZE_CHANCE`, `INVALID_PRIZE_TYPE`, `PRIZE_TYPE_MISMATCH` (a `prizeGroupId` in a `prize` game or the other way round), `PRIZE_NOT_FOUND`, `PRIZE_GROUP_NOT_FOUND`, `INVALID_INTERVAL_TIME`, `INVALID_PRICE_PAYLOAD` (currency without amount or the other way round), `INVALID_PRICE_AMOUNT`, `CURRENCY_NOT_FOUND`, `INTERVAL_OR_PRICE_REQUIRED` (a game needs a free turn, a ticket price or both), `PRIZE_GROUP_EMPTY`, `DUPLICATE_PRIZE`, `INVALID_WEIGHT`, `KEY_IMMUTABLE`, 409 `IN_USE` (a game with plays; delete is refused, switch it off instead).

### Client RPCs

All answers carry `code: "SUCCESS"`. `id` may be replaced by `key`.

| Spec | RPC | Input | Output |
|---|---|---|---|
| list | `economy_chance_games` | `{lang?}` | `{code, lang, items:[{id, key, name, description, metadata, prizeType, prize:[{chance, prizeId, prizeGroupId, prize:{id,key,name,message}|null, group:{id,key,name,prizes:[{id,key,name,weight,share}]}|null}], intervalTime, price:{currencyId, currencyKey, currencyName, amount}|null, plays, canPlay?, nextPlayTime?, nextPlayAt?, lastPlayAt?}]}` (active games; the turn fields only when the game has a free turn) |
| get | `economy_chance_game` | `{id, lang?}` | `{code, lang, game}` |
| CanPlay | `economy_chance_can_play` | `{id}` | `{code, canPlay, nextPlayTime}` |
| GetNextPlayTime | `economy_chance_next_play_time` | `{id}` | `{code, nextPlayTime, nextPlayAt?}` |
| Play | `economy_chance_play` | `{id, lang?}` | `{code, prizeId, prize, prizeGroupId, group, playId, gameId, gameKey, mode, slot, roll, rewards, spent, status, nextPlayTime, nextPlayAt?, wallet, lang}` |
| PlayWithTicket | `economy_chance_play_with_ticket` | `{id, lang?}` | same, `spent: {currencyKey, amount}` |
| history | `economy_chance_history` | `{id?, limit?, cursor?, lang?}` | `{code, lang, items:[{playId, gameId, gameKey, gameName, mode, status, roll, slot, prizeId, prize, prizeGroupId, spent, rewards, createdAt}], cursor}` (own plays, newest first) |

`economy_rewards_mine` and the console claims list name chance claims with the game (`kind: "chance"`). A won prize also writes the usual `prize granted` activity row (source `chance`).

### Console API (`/v2/console/economy/`, `HIRO_ECONOMY`, every write in the console audit log)

CRUD `chance-games` (search on key and every name translation; filters `isActive`, `prizeType`, `mode=free|ticket`; rows carry `plays`, `players`, `wins`, `lastPlayAt`) and `prize-groups` (filter `prizeId`; rows carry `members` and `games`) through the generic resource handlers. `GET /chance-plays?gameId&userId&prizeId&mode&won&status&q&page&limit` and `GET /users/{userId}/chance-plays?…` list plays with game, player, prize, group, roll, price and status. A currency used as a ticket price cannot be deleted (`IN_USE`). The summary counts `chance-games`, `prize-groups` and `chancePlaysToday`.

### Console pages

Economy section after «دستاوردها»: «بازی شانس» (`#/economy/chance-games`, Phosphor dice-five duotone) and «گروه‌های جایزه» (`#/economy/prize-groups`, Phosphor package duotone), both in the plane sidebar (now 20 economy entries) and in Ctrl K. The game list shows a dice tile, a stacked chance bar (empty slots hatched) with «۴ خانه، ۳ با جایزه», the way to play («رایگان هر ۱ روز», «بلیت: ۱۵ سکه»), plays and players, status and prize type; cards at 390px. The docked editor has key, name, description, play stats, «هر خانه چه می‌دهد» (prize or prize group), the chances table (chance in %, prize or group picker by type, add/remove rows, a new row gets what is left to 100), a live sum «جمع شانس‌ها: ۹۰ از ۱۰۰» with «۱۰ مانده» / «درست است» and the bar (green when exactly 100, red otherwise; saving is blocked until it is 100), the free-turn interval, ticket currency and price (both or none), a warning when neither is set, metadata and sort. The group editor has key, name, description, the prize rows with weights and a live «شانس هر جایزه در گروه» list in percent. The native player page gets «بازی‌های شانس بازیکن» after the achievements card: the last plays with game, time, free or ticket, «برد: …» or «پوچ», the group it came from, the price taken and the roll, «بیشتر نشان بده». Light and dark use the plane tokens; Persian digits in fa.

### Star Catcher

The seed adds the prize groups «کیسهٔ سکه» (`star_box_coins`: 15 / 25 / 40 coins, weights 3 / 2 / 1) and «گنج بزرگ» (`star_box_big`: 100 / 200 coins, weights 2 / 1), the free game «گردونهٔ روزانه» (`star_wheel_daily`, once a day: 10 coins 35%, 20 coins 25%, 50 coins 10%, nothing 30%) and the ticket game «جعبهٔ شانس» (`star_lucky_box`, 15 coins: coin pouch 55%, big treasure 10%, nothing 35%). Rows are matched by key and never updated; group members are written only when the group is created. The «شانس» screen (`web/chance.js`) draws each game as a wheel from its chance table, spins it to the server's `roll`, shows the win or «پوچ», counts down to the next free turn, disables the ticket button when coins are short, lists the odds and «بازی‌های اخیر من»; the menu card has a dot with the free turns ready.

### Chest slots, price ladder, pity and reveal (2026-11-21)

Added for the Mench (Ludo) port; migration `20270108000000-economy-chance-boosters-streaks.sql` (Up re-runnable, Down raises). Existing games keep today's behaviour: every new column defaults off.

- **Chest with slots** (`prizeType: "slots"`). Each `prize` row is an independent slot `{chance 0-100, kind: currency|prize|prizeGroup, currencyId+min+max | prizeId | prizeGroupId}`; the chances do not add up to 100 (`SLOTS_ALL_ZERO` when none is above 0, `INVALID_SLOT_RANGE`, `INVALID_SLOT_KIND`). Every slot rolls on its own; a currency slot pays a uniform whole amount in `[min, max]`; a non-repeatable prize the player owns is a miss. `rerollEmpty` (default on, as legacy) rolls the whole chest again when every slot misses, up to 1000 times. The roll is stored on the play as `outcome` (JSONB). The ticket debit and the currency credits run in one transaction with ledger grant ids `<playId>` and `<playId>:slots` (idempotent); prize slots are granted with `source_type='chance_slot'`, `source_id='<playId>:<slot>'` (no reward claim row, so `chanceGrantExtras` is not involved). A pending play is finished by the existing recovery loop. Slots refuse a jackpot or pity (`JACKPOT_NOT_FOR_SLOTS`, `PITY_NOT_FOR_SLOTS`). Legacy seed chest L1: coin 70% 120–180, gem 30% 3–6.
- **Price ladder** (`priceLadder: [int]`, 1–10 steps, needs the ticket currency, `LADDER_NEEDS_PRICE`, `INVALID_PRICE_LADDER`). The nth ticket play of the project day costs `ladder[min(n, len-1)]`; `price.amount` in the views is the next price, plus `priceLadder` and `ticketsToday`. Both ticket RPCs take an optional `expectedCost`; a mismatch is `CHANCE_PRICE_CHANGED` (409, meta `{amount, expected}`) and charges nothing. Legacy default ladder `[10,15,20,25,30]` (gems) with one free spin a day (`intervalTime` 1440).
- **Pity.** One row may be `jackpot: true` (`ONE_JACKPOT`); `pityThreshold` (1–100000, needs the jackpot row, `PITY_NEEDS_JACKPOT`) forces the jackpot when `since_jackpot + 1 >= threshold`. `economy_chance_state.since_jackpot` resets on any jackpot (natural or forced) and not on the daily reset; forced plays have `economy_chance_play.forced`. The counter and threshold are never sent to the player. Legacy default 40.
- **Reveal ack** (`revealAck`, default off). The reward is granted at once; the newest done play without `acked_at` is `pendingReveal` in `economy_chance_game(s)`, and play answers carry `needsAck`. `economy_chance_ack {playId}` sets `acked_at` once (`PLAY_NOT_FOUND` for an unknown or foreign play). An unacked play does not block the next one.
- Console: «هر خانه چه می‌دهد» has «صندوق خانه‌دار» with its own slot rows (kind, currency, min, max, prize, group) and the re-roll switch; prize games get «ردیف جایزهٔ بزرگ» and «تضمین جایزهٔ بزرگ بعد از»; the play section gets «پلکان قیمت بلیت در روز» and «نمایش نتیجه تا تأیید بازیکن»; list rows show ladder and pity tags. Play log rows include `outcome`, `forced`, `ackedAt`.

## Games (phase 15)

Owner's idea (2026-11-13): one Nakama hosts many games (for example a «Hot Board» product with ten board games). Game logic stays code: authoritative match handlers in a Go plugin, one handler per game key (`hotboard_ludo`, …). There is no «add a game» in the panel that makes logic. Each game's settings live in the console, and both the client and the server read them.

### Decisions

- The feature lives in the economy plugin, next to the wallet, prizes, prize groups and level tables it needs, under `/v2/console/economy/` (`HIRO_ECONOMY`, every write in the console audit log).
- The shared package `internal/gamecatalog` is the contract between plugins, like `internal/playerevents`: a game plugin calls `gamecatalog.Register(Definition{Key, Name, Plugin, Realtime, Schema})` in `InitModule`; the economy plugin installs the provider (`SetProvider`). Match handlers use `gamecatalog.Settings`, `Enter`, `Leave`, `Start`, `Cancel` and `Finish`; errors are `*gamecatalog.Error` with a `Code` (`gamecatalog.Code(err)`). Plugins load in any order.
- The key is the match handler name, locked after creation (`KEY_IMMUTABLE`), 2–64 lower-case letters, digits or `_`, starting with a letter (`INVALID_GAME_KEY`).
- Extra settings are checked against the JSON schema the game's code registers. Supported subset: `type: object`, `properties` of type `integer`, `number`, `string` or `boolean` with `title` and `description` (text or per language), `default`, `minimum`, `maximum`, `minLength`, `maxLength`, `enum`, plus `required` and `additionalProperties: false`. Missing values get their defaults on save; `INVALID_SETTING`, `UNKNOWN_SETTING` and `SETTING_REQUIRED` carry the field `settings.<name>`. A game without registered code keeps its settings as a plain object.
- Settings cache: each node keeps every game row in memory and checks one version number (`economy_game_version`, bumped by a statement trigger on any change to `economy_game`, also direct SQL) on each read; rows are reloaded only when the version moved. A console save is seen by the next read on every node.
- Entry fee: spent through the wallet path (`walletChange`, ledger `grantId` = UUIDv5 of match, player, `fee` and the attempt number) inside a transaction that locks the match row and the player's entry row, so parallel joins on many nodes spend once. A player who cannot pay leaves no entry. Leaving before the start, or a cancelled match before the start, refunds once (`refund` grant id per attempt); joining again after a refund charges again. After `Start` leaving forfeits the fee; a rejoin of a paid player is free. Open matches older than 15 minutes are cancelled and refunded by the event loop on any node.
- Prize: a room pays a prize (must be repeatable, `PRIZE_NOT_REPEATABLE`) or a prize group (one member drawn by weight and stored on the entry before paying). `Finish(match, winners)` marks the match finished once, stores the winners, then pays each winner through `grant()` with source type `game` and source id UUIDv5(match, player, `prize`), so a repeated or parallel finish pays once. A draw is `Finish` with no winners. Winners get one point on the linked leaderboard.
- Matchmaking: `RegisterBeforeRt("MatchmakerAdd")` (through `internal/hookchain`) checks every ticket that has the string property `game` (switched on, platform and version, room, level, balance) and rewrites the query to `+properties.game:<key> +properties.room:<room>` with the game's player counts; tickets without `game` are left alone. `RegisterMatchmakerMatched` creates an authoritative match with the game key as module and `{game, room, users}` as params when the game's code is realtime; otherwise the match stays relayed. A game plugin must not register its own matchmaker-matched hook (the last one registered wins in Nakama).

### Tables

`economy_game` (key, on/off, name, description, maintenance message per language, icon, `game` category and tags, sort, min/max players, turn seconds, match seconds, bots allowed, rooms JSON `[{key, name, entryFee: {currencyKey, amount}, prizeId | prizeGroupId, minLevel}]` up to 20, leaderboard id, platforms, min/max app version, settings), `economy_game_version`, `economy_game_match` (match id, game, room, `open` → `started` → `finished`, or `cancelled`, players, winners), `economy_game_entry` (match, player, `pending`/`paid`/`free`/`refunded`/`left`, fee, attempt, prize, prize status). No foreign key from player rows to `users`; entries of deleted players are swept after the match ends. New key type `game` for categories and tags; `economy_user_prize.source_type` gets `game`.

Since phase 18 `economy_game.voice_enabled` (default true, added by `20261124000002-base-voice.sql`) switches voice chat for one game; the base plugin's `voice_token` reads it (`docs/base/spec.md` «Voice chat»). The editor shows it as «چت صوتی», the list as «بی‌ویس» when off, admin rows carry `voiceEnabled` and `voiceMaxPlayers` (per-game voice room size, NULL = the voice server setting; `INVALID_VOICE_PLAYERS`) and `economy_games`/`economy_game` items carry `voice`.

### Client RPCs (errors are the chance-game JSON `{code, status, meta}`)

- `economy_games {lang, platform, version}` → `{items, level}`: games for this platform and version in lobby order, names in the player's language, rooms with `entryFee.currencyName`, `prize` (`prize` or `prizeGroup` with its name), `minLevel` and `locked`, `settings`, `realtime`, `multiplayer`. A switched-off game is listed with `isActive: false` and `maintenance` only when it has a message; otherwise it is hidden.
- `economy_game {key, lang, platform, version}` → one game (`GAME_NOT_FOUND`, `PLATFORM_NOT_SUPPORTED`, `UPDATE_REQUIRED`, `VERSION_NOT_SUPPORTED`).
- `economy_game_join {key, room, platform, version, lang, path}` → the matchmaker ticket `{query, minCount, maxCount, stringProperties: {game, room}, numericProperties: {level}, entryFee, realtime, path, consumes}`; entry requirements refuse with `REQUIREMENTS_NOT_MET` (see «Entry requirements»). Refusals: `GAME_DISABLED` (`meta.message`), `SINGLE_PLAYER_GAME`, `ROOM_REQUIRED`, `ROOM_NOT_FOUND`, `LEVEL_TOO_LOW` (`meta.minLevel`), `INSUFFICIENT_BALANCE` (`meta.amount`, `meta.balance`). The same checks run again in the matchmaker hook, so a hand-made ticket cannot skip them. The fee itself is spent when the match handler lets the player in.

### Console API

CRUD `games` (list filters `categoryId`, `platform`, `isActive`; match counts in each row), `GET /games-registry` (registered definitions with `schema`, `configured`, plus `unconfigured` and `missingHandler`), `GET /game-matches` (`gameKey`, `status`, `userId`, `q`), `GET /game-matches/{id}` (entries), `POST /game-matches/{id}/cancel` (refunds an open match). Delete is refused (`IN_USE`) while the game has open or started matches.

### Console page

«بازی‌ها» (`#/economy/games`, Phosphor puzzle-piece duotone) in the economy sidebar (now 21 entries) and Ctrl K. Above the list: a warning per registered game without settings with «تنظیم کن» (opens the editor with the key and name filled), an error per configured game whose code is not on this server, or «همهٔ … بازیِ ثبت‌شده تنظیم دارند». Rows: icon, name and key, players and turn time, room chips («bronze: ۱۰ سکه · سطح ۲»), matches and players, on/off and the code tag («بازی زنده», «کد ثبت شده», «کد ندارد»); cards at 390px. The docked editor: key with a line telling whether code registers it, on/off, name, description, maintenance message, stats, category and tags (`game` key type, quick-create), icon and lobby order; players and time (fewest/most players, turn, match limit, bots); rooms row editor (key, entry currency and amount, lowest level, winner prize or prize group, room name per language); platforms and app versions, leaderboard id; «تنظیمات ویژهٔ این بازی» drawn from the registered schema (number, text, choice and switch fields with min/max/default hints, checked before saving and server errors on the field), or a JSON box when no code registers the key. fa and en strings, Persian digits in fa, light and dark.

### Star Catcher

`samples/stargame` registers two games: `stargame` (single player, schema `roundSeconds` 20–120 default 30 and `maxCoins` 0–100 default 30) and `stargame_tictactoe` (realtime, schema `firstMove` random/first_joined, `waitSeconds`, `forfeitOnLeave`) with the authoritative match handler in `ttt.go`. `stargame_start` reads the round time and coin cap from the settings (refuses with `GAME_DISABLED` when off; the shortest accepted round is the round time minus 5 s). The tic-tac-toe handler charges the entry on join, starts when two players are in, ends on three in a row, a full board, a turn timeout (turn seconds from the console), the match limit or a player leaving, pays through `Finish`, and cancels with refunds when the opponent does not come in `waitSeconds`. The seed adds the categories «آرکید» and «تخته‌ای», the two game rows (tic-tac-toe rooms: bronze 10 coins → 20 coins, silver 25 coins → 50 coins from level 2, gold 50 coins → «گنج بزرگ» from level 3) and the leaderboard `stargame_tictactoe_wins`; rows are matched by key and never updated. The «بازی‌ها» screen (`web/lobby.js`) lists the games with rooms, fees, prizes and locks; a room button calls `economy_game_join`, adds the matchmaker ticket over the socket, joins the match the server made and draws the board.

To add a second game in another plugin: `gamecatalog.Register(...)` and `initializer.RegisterMatch(key, …)` in its `InitModule`, then configure the key in «بازی‌ها». Inside the handler call `gamecatalog.Enter` in `MatchJoinAttempt` (reject with its code), `Start` when the match begins, `Leave` for players leaving before the start, `Finish(matchID, winners)` at the end and `Cancel` when it never started.

## Item sounds

Migration `20261210000000-economy-item-sound.sql` adds `economy_item.sound_url` (default empty). The item editor has a «صدا» field (upload or link, with a player preview); the console upload endpoint `POST /v2/console/economy/uploads` also takes mp3, ogg and wav (sniffed from the file, 2 MB like images) and serves them from `/v2/economy/files/<name>` with an audio content type. Client item views (`economy_inventory`, item answers) carry `soundUrl`. Games read it from the equipped item, for example a goal sound slot.

## Avatar designer

Migration `20261228000000-economy-avatar.sql`, code `integrations/economy/avatars*.go`. Replaces the base avatar builder; its data is copied (parts → categories with one slot, options → items, each saved look → one active avatar with its choices, `""` choices → cleared categories).

### Data

| Table | What |
|---|---|
| `economy_avatar_settings` (row 1) | `enabled` [true], `maxAvatars` [1, 1–50], `tagLimits` [`{tagId, maxAvatars}`, the largest matching player tag wins], `nameMaxLength` [24], `autoCreate` [true: a player with no avatar gets one on the first call], `keepLast` [true: the last avatar cannot be deleted], `slotFull` [`replace` or `refuse`, for categories with more than one slot], `keepBought` [true: a bought item stays wearable after its condition stops holding] |
| `economy_avatar_type` | avatar bases such as male/female: `key`, `name`, `image_url` (empty body), `icon_url`, `sort` (first active type is the default for new avatars), `is_active` |
| `economy_avatar_category` | `key`, `name`, `icon_url`, `layer` (drawing order), `required` (always one item), `slots` [1, 1–20], `sort`, `is_active` |
| `economy_avatar_item` | `category_id` (immutable), `key` (unique in the category), `name`, `image_url`, `icon_url`, `value`, `access` (`free`, `currency` with `price {currencyKey, amount}`, `product` with `product_id`: owned after a successful payment for it, `item` with `item_key`: owned with that economy item), `requirements` (the game entry-requirements shape: `minLevel`, `metas [{metaId, operator, value}]`, `tagIds`, `excludeTagIds`, `itemIds`, `lock` `show`/`hide`, `message`; nothing is consumed), `type_ids` (empty = every type), `is_default` (free, no condition, one per category), `sort`, `is_active` |
| `economy_avatar_unlock` | currency purchases `(user_id, item_id)` with `paid` and `request_id` |
| `economy_avatar` | `user_id`, `type_id`, `name`, `is_active` (one per player), `cleared` (categories the player emptied) |
| `economy_avatar_equip` | `(avatar_id, item_id)`, `category_id`, `equipped_at` |

A look is resolved per category: worn items that are active, fit the avatar type and are still owned (newest `slots` kept); if none and the category is not cleared, the default item, or for a required category the first free item without a condition that fits. Conditions are checked when buying and wearing.

### Client RPCs

`economy_avatar_catalog {lang?, avatarId?}`, `economy_avatar_buy {category, key, requestId?}`, `economy_avatar_list`, `economy_avatar_create {name?, type?, activate?}`, `economy_avatar_rename {id, name}`, `economy_avatar_delete {id}`, `economy_avatar_equip|unequip {id?, category, key}` (no id = the active avatar), `economy_avatar_set_active {id}`, `economy_avatar_looks {userIds}`. Errors: `AVATAR_DISABLED`, `AVATAR_NOT_FOUND`, `AVATAR_TYPE_NOT_FOUND`, `AVATAR_CATEGORY_NOT_FOUND`, `AVATAR_ITEM_NOT_FOUND`, `AVATAR_TYPE_NOT_ALLOWED`, `AVATAR_ITEM_LOCKED`, `AVATAR_ITEM_NOT_OWNED`, `AVATAR_ITEM_FREE`, `AVATAR_ALREADY_OWNED`, `AVATAR_PRODUCT_REQUIRED`, `AVATAR_SLOT_FULL`, `AVATAR_CATEGORY_REQUIRED`, `AVATAR_NOT_EQUIPPED`, `AVATAR_LIMIT`, `AVATAR_LAST`, `INVALID_NAME`, `INVALID_USER_IDS`, `INSUFFICIENT_BALANCE`. The old `base_avatar`, `base_avatar_save` and `base_avatar_get` keep their shapes and JSON `{code, meta}` errors on the active avatar.

### Console

«آواتارساز» hub in the economy sales group (`#/economy/avatar`): «دسته‌ها», «آیتم‌ها» (filters category, access, type, status; editor with look, how to get, allowed types and the unlock-condition panel), «نوع‌ها», «تنظیمات». Console API: CRUD `avatar-types`, `avatar-categories`, `avatar-items` (delete refused with `IN_USE` while bought or used), `GET/PUT /avatar-settings`, `GET /users/{userId}/avatars`.


## Regenerating currencies (lives, energy)

Any wallet currency can refill by itself over time. The setting lives on the currency row (`economy_currency.regen`, JSON, null = off), edited in the currency editor («پرشدن خودکار»). Migration `20270103000000-economy-regen.sql` (re-runnable) adds the column and three tables.

### Settings (`regen` on `POST/PUT /currencies`)

| Key | Default (editor) | Meaning |
|---|---|---|
| `enabled` | false | turn regen on |
| `max` | 5 | regen fills up to this (1–1 000 000). Purchases, prizes and refunds of other sources may go above it |
| `intervalSeconds` | 1800 | one unit per interval (1 s – 30 days) |
| `clamp` | false | on the next read a value above `max` is cut back to `max` (ledger reason `regen_clamp`) |
| `refillPrice` | null | `{currencyKey, amount}` for a full refill; another currency, amount ≥ 1; null = no refill |
| `unlimitedTag` | "" | a base tag (key type `user`); while the player holds it unexpired, spends of this currency are free |

A `PUT` without the `regen` field keeps the stored value (the setup wizard saves currencies without it); `regen: null` clears it. Errors: `INVALID_REGEN`, `INVALID_REGEN_MAX`, `INVALID_REGEN_INTERVAL`, `REFILL_SAME_CURRENCY`, `INVALID_KEY`, `INVALID_PRICE`, `CURRENCY_NOT_FOUND`, `TAG_NOT_FOUND`. The list filter `regen=true|false` and the list column «پرشدن خودکار» show it. `economy_catalog` currencies carry `regen: {max, intervalSeconds, refillPrice} | null`.

The "unlimited lives for N minutes" buff is a timed player tag from a prize (`{type:"tag", key, action:"give"|"set", value: days}`); days take 4 decimals, so 30 minutes is `0.0208` (29.95 min) and the step is about 9 seconds.

### Tables

| Table | What |
|---|---|
| `economy_user_regen` | per player and currency: `anchor_at` (when the next unit started counting) and `last_value` (the value at the last settle) |
| `economy_regen_spend` | one row per `(user_id, source_id)`: currency, amount, `free` (paid by the unlimited tag), `refunded`, `refunded_at` |
| `economy_regen_refill` | one row per `(user_id, request_id)` with the stored answer |

### Lazy settle

There is no cron. Every read or spend locks the player's `users` row and settles: `n = floor((now − anchor) / interval)`, `gain = min(n, max − value)`, the gain is written to the wallet in the same transaction (`pluginkit.WalletTx`, ledger reason `regen`, `sourceType: "regen"`), and the anchor moves by `n × interval` (or to now when the value reaches `max`). While the value is at or above `max` the anchor follows now, so the first unit after a spend comes one interval after that spend. If the wallet dropped below `max` outside regen (a plain debit by another feature) since the last settle, the timer starts at that read (no instant refill). Spends through the hook or RPCs below all run inside the same lock, so parallel calls cannot double-spend.

View: `{currencyKey, value, regen, max, intervalSeconds, nextAt, fullAt, unlimited, unlimitedUntil, refillPrice}`; `nextAt`/`fullAt` are null when full, `unlimitedUntil` is null for a tag without an end. A currency without regen answers `regen: false` with only `value` (and the unlimited flags if it has an unlimited tag).

### Client RPCs

| RPC | In | Out |
|---|---|---|
| `economy_regen_state` | `{currencies?: [key]}` (max 20; empty = every regen currency) | `{serverTime, currencies:[view]}`; `INVALID_KEY`, `CURRENCY_NOT_FOUND`, `TOO_MANY_CURRENCIES` |
| `economy_regen_refill` | `{currencyKey, requestId}` | `{currencyKey, filled, paid:{currencyKey, amount}, state: view, wallet, replayed}`; idempotent per `requestId` (same answer, `replayed: true`); `REQUEST_ID_REQUIRED`, `REQUEST_ID_REUSED`, `REGEN_OFF`, `REFILL_OFF`, `ALREADY_FULL`, `INSUFFICIENT_BALANCE` |

### Hooks for other plugins (`internal/economyhooks`)

```go
func RegenView(ctx context.Context, userID, currency string) (Regen, error)
func SpendRegen(ctx context.Context, userID, currency string, amount int64, sourceID string) (RegenSpend, error)
func RefundRegen(ctx context.Context, userID, currency, sourceID string) (RegenRefund, error)
func RegenAvailable() bool
```

`RegenSpend{View, Spent, Free, Replayed}`, `RegenRefund{View, Refunded, Replayed, Found}`. Without economy they return `ErrNoRegen`. `SpendRegen` works for any currency (a currency without regen is a plain idempotent debit), is idempotent per `(user, sourceID)` (`SOURCE_REUSED` when the source spent another currency) and refuses with `*Refusal{Status: 409, Code: "NOT_ENOUGH_CURRENCY", Meta: {currencyKey, value, amount, nextAt, fullAt, refillPrice?}}`; a refused spend writes nothing but the settle. `RefundRegen` gives the spent amount back once per source (ledger `regen_refund`), never above `max` for a regen currency, nothing for a free spend; an unknown source answers `Found: false`. Validation errors come back as `*Refusal` too. Economy registers them in `InitModule` (`SetRegen`). The campaign plugin uses them for its play cost (`docs/campaign/spec.md`).

Other plugins read the store and the wallet through `economyhooks.Catalog(ctx, userID, lang)` (the `economy_catalog` answer) and `economyhooks.Balances(ctx, userID)` (the `economy_balances` answer: defaults granted, regen settled). Without economy both return `ErrNoCatalog`.

`economyhooks.LevelProgressOf(ctx, userID, tableKey)` reads a player's place on an active level table by key: `LevelProgress{TableKey, Level, MaxLevel, Value, LevelStart, Next}` (`Value` = the table's meta value, `LevelStart` = threshold of the current level, `Next` = threshold of the next level or nil on the last one; a player who never moved is on the highest level whose threshold is 0, like `economy_levels`). Unknown or inactive table → `ErrNoLevelTable`; without economy → `ErrNoLevels`. Levels and their prizes still move only through meta changes (events, meta rules, prizes).

## Stealth bots

Per game, column `economy_game.stealth` (JSON, edited at `GET/PUT /v2/console/economy/game-stealth/{key}`, list at `GET /game-stealth`; console «ربات‌ها › بات پنهان»). Parsed by `internal/stealthbots` (`Parse`, `Of(game)`, defaults with `enabled: false`). Game code reads `stealthbots.Of(g)` from `gamecatalog.Settings`.

- Settings: `enabled`, `poolBots` (true), `syntheticFallback` (true), `levelRange` (2), `arriveMin/MaxSeconds` (4/14), `thinkMin/MaxMs` (700/2600), `pausePercent` (6) with `pauseMin/MaxMs` (2500/6000), `randomSide` (true), `onlineAfterMin/MaxMinutes` (3/25), `onlineIdlePercent` (4), outcomes `trophies`, `fullPrize`, `leaderboard`, `progress` (all true), guard `guard` (false), `guardDailyWins` (10), `guardBoost` (1). Helpers: `ArriveDelay`, `ThinkDelay` (pause chance included), `Side(n)`, `OnlineAfter`, `BoostFor(ctx, db, user, game)`.
- Pool bots: `gamecatalog.Bots` with `MatchID` leases bots in `economy_bot_lease` (one live match per bot; also skips bots seated in another open/started match; stale after 30 min) and picks by `Level ± LevelRange`, widening ×2, ×4, then any. Leases are released on settle, finish and cancel (`gamecatalog.ReleaseBots` for games that do not settle through economy); the bot stays online for `OnlineAfter`.
- Presence: `stealthbots.MarkUsers/MarkFriends` set `online` in `base_users_get`, `base_friends_list` and friend status for leased bots, bots in their online tail and a rotating `onlineIdlePercent` of idle bots; only while some game has stealth on.
- Client answers when on: `stealthbots.Scrub` removes bot fields from `economy_games`/`economy_game` (incl. `eligibleBy.bot` and `bot*` settings), `economy_game_queue` drops `humanSeats`/`botFillSeconds`, `economy_game_bot` drops `botAfterSeconds`, friendroom labels report `matchmaking` for bot rooms (`Params.Stealth`). `settings.stealth` is never sent to clients.
- Outcomes when on: bot entries get career stats (`progress`), bot winners are written to the game leaderboard (`leaderboard`), human wins against bots count in `economy_bot_guard` (`guard`).
- Accounts: `metadata.bot` is no longer written and is removed from existing bots (migration `20270206000000-economy-stealth-bots.sql`); `privatemeta` strips a `bot` key; bot accounts get a create time 20–420 days back. `economy_player_card` shows a bot's XP inside its level while stealth is active. Console bot list shows `inMatch`, `matchId`, `onlineUntil`, `accountCreatedAt`.
- Friend requests to bots: `base_friend_rules.bot_decline_min_minutes` / `bot_decline_max_minutes` (random per request, stable by hash); max 0 falls back to `bot_decline_hours`; all 0 leaves requests pending.
- Markers: `stealthbots.Markers(json)` lists bot markers in a JSON payload (used by the game tests to scan every message).
- Pool growth (`autoGrow` true, `poolMax` 2000, `growNearLevel` true, `growStats` true): when `leaseBots` finds no free bot near the level (or none at all with `growNearLevel` off) and the enabled pool is below `poolMax`, it makes real accounts on the spot under an advisory lock (`stealthbots.Identity` names, `GrowLevel` near the asked level, `Age` 20–420 days, wallet for the room fee, source `stealth-grow`), seeds `economy_game_stats` (and `trophy_player` when that table exists) from `SeedStats`, and leases them to the match.
- Made-up fallback: games call `gamecatalog.KeepGhost(ctx, game, bot)`; rows live in `economy_bot_ghost` for `ghostDays` (90). `base_users_get` adds ghost users for ids/usernames it did not find (only while stealth is active) and `economy_player_card` builds a card with stable stats from the id.
- Bot look (`lookByLevel` true, `lookSpread` 15): `stealthbots.RankLook` orders items of one type by starter, rarity, then price percentile per currency; `PickLook` takes the item at the level's place (±spread) that the level has unlocked. Grown bots copy the slots of a pool bot that has a look. The card falls back to equipped owned items when `economy_bot.look` is empty.
- One match per player (`oneLiveMatch` true): `gamecatalog.OneMatch(ctx, user, game, create)` keeps `economy_game_claim`; while the claimed match still runs (`MatchGet`), is not finished/cancelled and the player has not left, the same match id comes back. `economy_game_bot` uses it; server fills in games should too.
- Stats for games that finish with `Finish`: `gamecatalog.RecordStats(ctx, matchID, results)` writes `economy_game_stats` once per match (`economy_game_match.stats_done`, also set by `Settle`), statuses `won`, `lost`, `draw` (a game only), `forfeit`, `timeout`; bots only with stealth `progress`.
- Match presences (`presences` true): core `RuntimeGoMatchCore.VirtualPresenceJoin/Leave` adds a virtual presence (session id = UUIDv5 of match id and user) shown in the join answer and sent as `MatchPresenceEvent`; it never routes data or counts toward size. Plugins call `stealthbots.ShowPresence/HidePresence(dispatcher, ...)`; nothing changes when no virtual presence is added.
