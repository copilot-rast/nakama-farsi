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

Rules: SQL must run on PostgreSQL 16 (CockroachDB is legacy, but avoid Postgres-only features where a portable form exists; no advisory locks, use `SELECT … FOR UPDATE`). Up sections re-runnable (`IF NOT EXISTS`), Down raises like the Jazebe migrations. No code comments. Every money/grant path idempotent.

## Runtime env (`--runtime.env KEY=VALUE`)

| Key | Meaning | Default |
|---|---|---|
| `ECONOMY_SECRET_KEY` | 32+ chars; AES-GCM key for gateway secrets and HMAC for callback tickets (derive two keys with HKDF, salt `nakama-economy`, info `economy-config-seal` / `economy-callback-ticket`; helpers in `internal/secretbox`, shared with publication and the stargame sample). Changing it makes stored secrets unreadable | required, module refuses to start without it; compose falls back to `NAKAMA_SESSION_ENCRYPTION_KEY` |
| `ECONOMY_PUBLIC_BASE_URL` | Public base of the API port, used for callback and file URLs | required |
| `ECONOMY_DEEP_LINK` | Redirect after a gateway payment; `{success}` and `{paymentId}` are replaced | `nakama://payment?success={success}&paymentId={paymentId}` |
| `ECONOMY_WEB_REDIRECT_URL` | If set, used instead of the deep link (same placeholders) | empty; compose sets `https://nakama.dastar.dev/stargame/?payment={paymentId}&success={success}` |
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
- `product(id, key unique, name jsonb, description jsonb, is_active bool, category_id null, tag_ids text[], price_toman bigint null, coin_price jsonb null {currency_key, amount}, discount_percent int 0..100, payment_methods text[] subset of ('zarinpal','zibal','bazaar','myket','in_app'), sku jsonb {bazaar?, myket?}, is_subscription bool, prize_id not null, icon_url, cover_url, min_level int 0, metadata jsonb, sort int, created_at, updated_at)`. Validation: gateway methods need `price_toman >= 100`; `in_app` needs `coin_price`; store methods need the matching `sku`.
- `gift(id, name jsonb (translated; existing text names were migrated to `{default: text}`), type text check in ('reward','discount'), code text unique null (stored upper-case), prize_id null, discount_percent null 1..100, is_promo bool, per_user_limit int default 1, max_total_uses int null, total_used int, valid_from, expire_at, is_active, category_id null (key type `gift`), tag_ids text[] (`gift` or `general`), created_at, updated_at)`. `reward`: redeem grants `prize_id`. `discount`: usable as `gift_code` on a gateway purchase only.
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

Language: every RPC that returns text accepts `lang` and returns resolved strings plus `lang` (the language used). Resolution: requested `lang` if it is an active language (`en-US` matches `en`) → the account `lang_tag` if active → default language; per field, a missing value falls back to the default language, then the first non-empty value. An unknown `lang` is not an error.

| RPC | Input | Output |
|---|---|---|
| `economy_catalog` | `{categoryKey?, lang?}` | `{lang, currencies:[{key,name,type,iconUrl,categoryKey,tags}], categories:[{key,name,description,iconUrl,sort}], products:[{id,key,name,description,iconUrl,coverUrl,priceToman,finalToman,coinPrice,discountPercent,paymentMethods,sku,isSubscription,categoryKey,tags}]}` (active only, methods filtered to active methods) |
| `economy_balances` | `{}` | `{wallet:{key:amount}, metas:{key:value} (non server_only), items:[…storage economy_items]}` |
| `economy_purchase_start` | `{productId, method, giftCode?}` | gateway: `{paymentId, status:"redirect", redirectUrl}`; store: `{paymentId, status:"store", sku, developerPayload}`; in_app or free: `{paymentId, status:"success", rewards}` |
| `economy_purchase_submit` | `{paymentId, token}` | `{paymentId, status, rewards?}` |
| `economy_purchase_review` | `{method, sku, token}` | same as submit |
| `economy_purchase_status` | `{paymentId}` | `{paymentId, status, rewardGranted, error?}` (own payments only) |
| `economy_store_config` | `{method}` | `{method, sandbox, rsaPublicKey, packageName}` |
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
- `GET /v2/console/economy/summary` → counts for the overview page (`categories`/`tags` count the economy key types in `base_category`/`base_tag`).

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
| `economy_reward_claim` | one row per reward a player earned: `kind` (`event`, `level`, `streak`), `owner_id` (rule, table or streak), `ref_id` (rule, step or reward), `scope` (`<window>#<n>`, `L<level>`, `r<run>:<day>`), prize, grant mode, `status` (`ready`, `granted`, `failed`), attempts. Unique `(kind, ref_id, user_id, scope)` |

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

### Star Catcher

The seed adds the task categories `star_missions_daily` and `star_missions_story`, daily «امروز دو دور بازی کن» (20 coins) and «پنج ستارهٔ طلایی بگیر» (15), weekly «این هفته ۱۰۰۰ امتیاز جمع کن» (100), «شروع ستاره‌چین» with three sub-tasks (one round, three golden stars, a round of 100+; 200 coins, in-app note), «کهنه‌کار: ده دور بازی کن» (after the starter, from level 2; 100) and the group «زنجیرهٔ شکارچی» (three tasks in order, automatic hand-out, 10/15/30 coins). Every seeded task is collected with a button, so no earlier balance changes.
