# Publication module (phase 1)

Port of asp `Modules/Publication` (Faq, News, Notification) as a Go runtime plugin with console pages in the existing console overlay, plus real SMS and email delivery (asp only logs them). Push (Firebase Cloud Messaging) came in phase 13, see «Push, campaigns and automatic notifications» below.

References:
- asp: `/Users/mazyar/Projects/asp/backend/src/PolicePatrol.Api/Modules/Publication/{Faq,News,Notification}`, provider fields in `Modules/Base/Sms/Services/SmsProviderCatalog.cs` and `Modules/Base/Email/Services/EmailSenderCatalog.cs`, admin UI in `/Users/mazyar/Projects/asp/frontend/src/features/{faqs,news,notifications,smsProviders,emailSender}`.
- Economy module in this repo (same conventions): `docs/economy/spec.md`, `integrations/economy` (crypto.go, resources.go generic CRUD, admin.go route registration, payment_methods.go masking), `console/ui/economy` (overlay).

## Layout

| Part | Path |
|---|---|
| Go plugin | `integrations/publication/` → `publication.so` |
| Migrations | `migrate/sql/2026103100xxxx-publication-*.sql` |
| Console auth for `/v2/console/publication/` | `server/console.go` (same wrapper branch as economy), `console/acl/acl.go` (GET → `NOTIFICATION` read, others → `NOTIFICATION` write) |
| Console pages | extend the existing overlay in `console/ui/economy/` with a second sidebar section «انتشار» / "Publishing" (do not create a second overlay) |

Same rules as economy: PostgreSQL 16, re-runnable Up, Down raises, no code comments, idempotent writes, each console handler registered with one method.

## Shared pieces

- Categories: `base_category` with `key_type` `faq` / `news` (Base module, `docs/base/spec.md`). Publication reads that table directly; `20261101000001-base-move-categories.sql` moved the old `economy_category` rows (ids kept) and repointed `category_id`.
- Languages and translated text: `base_language` + the shared helper `internal/localized` (validation and resolution rules in `docs/base/spec.md`).
- Images: reuse `POST /v2/console/economy/uploads` and `GET /v2/economy/files/{name}`.
- Secrets: `ECONOMY_SECRET_KEY` with a different HKDF info string (`publication-provider-config`).

## Data model (tables prefixed `publication_`)

Text fields that users see are `jsonb` maps keyed by any existing language code (`{"fa", "en", "ar", …}`); the default language value is required for required fields and whenever any value is given (400 `DEFAULT_LANGUAGE_REQUIRED`, unknown code 400 `INVALID_LANGUAGE`). Rich text (FAQ answers, news bodies, email bodies) is a restricted Markdown subset: paragraphs, `**bold**`, `*italic*`, links, lists, headings `##`/`###`, images `![](url)` — rendered to HTML server-side with escaping for email, returned raw + `html` to clients.

- `faq(id, category_id null, question jsonb, answer jsonb (markdown per lang), is_active, sort, created_at, updated_at)`.
- `news(id, title jsonb, summary jsonb, body jsonb, cover_url, category_id null, is_active, is_pinned, published_at, allow_comment bool, free_comment bool, actions jsonb [{type:'url'|'goto', label:{<lang>: text} (required), value}], metadata jsonb, created_at, updated_at)`. Only `is_active` and `published_at <= now()` are visible to clients.
- `news_seen(news_id, user_id, seen_at, primary key both)`.
- `news_comment(id, news_id, user_id, parent_id null, message text 1..1000, is_accepted bool, created_at, updated_at)`. `free_comment` → accepted on create, else pending. One reply level (a reply to a reply attaches to the root). Clients see accepted comments plus their own pending ones. Rate limit: 5 comments per user per minute.
- `notification(id, channel text check in ('in_app','sms','email'), title jsonb, body jsonb, action jsonb null {type, value, label?: {<lang>: text}}, target jsonb {mode:'all'|'users'|'group'|'tags', userIds?, groupId?, tagIds?}, provider text null, status text check in ('queued','sending','done','failed'), recipient_count int, sent_count int, failed_count int, skipped_count int, created_by text, created_at, updated_at)`.
- `notification_recipient(notification_id, user_id, status text check in ('pending','sent','failed','skipped'), address text null, error text null, lang text null (language the recipient got), message text null (sandbox only: the text that would have been sent, email = subject + newline + plain text, max 1000), attempts int, seen_at null, sent_at null, primary key(notification_id, user_id))`.
- `provider(id, channel text check in ('sms','email'), kind text, is_active bool, is_default bool, sandbox bool, config text /*AES-GCM*/, updated_at)`, unique `(channel, kind)`; at most one default per channel.

## Targeting and addresses

- `all`: every non-disabled Nakama user; `users`: explicit ids (max 1000); `group`: members of a Nakama group (state ≤ member, not join requests).
- Phone: account metadata `phone` (accept `09xxxxxxxxx`, `+989…`, `989…`; normalize to `+98…` for Twilio, `09…` for Kavenegar/sms.ir). Email: account `email`, else metadata `email`. Users without an address are `skipped`.
- Recipients are materialized when the notification is created, in batches; delivery runs in a background worker started in `InitModule` (every 5 s, batch 50, up to 3 attempts with backoff, then `failed`). The notification row tracks counts and moves to `done`. Every node runs the worker: recipients are claimed with `FOR UPDATE SKIP LOCKED` (the claim bumps `attempts` and sets a 2-minute lease in `next_attempt_at`), and each external SMS/email send first renews its own claim (`attempts` unchanged); if another node took the row after the lease ran out, the send is skipped, so a slow batch never sends twice. In-app sends go through `NotificationsSend`: the inbox row is always stored, but the live socket push reaches only users connected to the sending node until the cluster phase routes it.

## Channels

- Every recipient gets text in its own language: account `lang_tag` if it is an active language, else the default language; per field, missing values fall back to the default language. The chosen language is stored in the recipient row (`lang`).
- `in_app`: for each recipient call `nk.NotificationsSend` with `persistent: true`, `code` 1001, subject = resolved title, content `{notificationId, title, body, action: {type, value, label?} | null, lang}`. Seen tracking via our table.
- `sms` providers (fields; secret fields masked like payment methods):
  - `kavenegar`: `apiKey` (secret), `sender`. `POST https://api.kavenegar.com/v1/{apiKey}/sms/send.json` form `receptor`, `sender`, `message`; `return.status == 200` ok.
  - `smsir`: `apiKey` (secret), `lineNumber`. `POST https://api.sms.ir/v1/send/bulk` header `X-API-KEY`, JSON `{lineNumber, messageText, mobiles:[…]}`; `status == 1` ok.
  - `twilio`: `accountSid`, `authToken` (secret), `from`. `POST https://api.twilio.com/2010-04-01/Accounts/{sid}/Messages.json` basic auth, form `To`, `From`, `Body`; 2xx with `sid` ok.
- `email` providers:
  - `smtp`: `host`, `port`, `username`, `password` (secret), `from`, `fromName`, `security` (`starttls`|`tls`|`none`). Use `net/smtp`; send `multipart/alternative` (text + HTML), UTF-8, `Subject` RFC 2047 encoded.
  - `resend`: `apiKey` (secret), `from`. `POST https://api.resend.com/emails` Bearer, `{from, to:[…], subject, html, text}`.
- Email HTML uses `lang` and the language's `direction` from `base_language`.
- **Sandbox** (per provider, default ON for new providers): nothing is sent; the message is written to the recipient row as `sent` with `error = "sandbox"`, `lang` and `message` (the resolved text), and to the server log. This protects real users while testing.
- Verify the exact request/response shapes of every provider against its current official docs before coding; if something differs from this file, follow the docs and note it.
- Test send: `POST /v2/console/publication/providers/{id}/test {to, message}` sends one message (honours sandbox) and returns the provider's reply summary.

## Client RPCs (user session)

| RPC | Input | Output |
|---|---|---|
| `publication_faq_list` | `{categoryKey?, q?, lang?}` | `{lang, categories:[{key,name}], items:[{id,question,answer,answerHtml,categoryKey}]}` |
| `publication_news_list` | `{limit?, cursor?, categoryKey?, lang?}` (phase 25 adds tags, expiry, unread, reactions and SEO; see below) | `{lang, items:[{id,title,summary,coverUrl,publishedAt,isPinned,seen,commentCount}], cursor}` pinned first |
| `publication_news_get` | `{id, lang?}` | full news incl. `bodyHtml`, `actions`, `allowComment`, `seen` |
| `publication_news_seen` | `{id}` | `{}` idempotent |
| `publication_news_comments` | `{newsId, limit?, cursor?}` | `{items:[{id,userId,username,message,parentId,isAccepted,mine,createdAt,replies:[…]}], cursor}` |
| `publication_news_comment` | `{newsId, message, parentId?}` | comment (with `isAccepted`) |
| `publication_inbox` | `{limit?, cursor?, lang?}` | `{lang, items:[{id,title,body,bodyHtml,action,createdAt,seen}], unread, cursor}` in_app only |
| `publication_inbox_seen` | `{ids:[…]}` or `{all:true}` | `{unread}` |

Every RPC with text accepts `lang` and returns resolved strings plus `lang` (the language used): requested `lang` if active → account `lang_tag` if active → default language; unknown `lang` is not an error. `faq_list.q` searches every translation. Inbox `action` is `{type, value, label?}` with the label resolved.

Errors: runtime errors with UPPER_SNAKE codes.

## Console API (`/v2/console/publication/`)

- CRUD (`GET list ?q&page&limit[&categoryId][&isActive]`, `GET /{id}`, `POST`, `PUT /{id}`, `DELETE /{id}`) for `faqs`, `news`.
- `GET /comments?newsId&status=pending|accepted&page&limit` (with username, news title), `PUT /comments/{id}` `{isAccepted}`, `DELETE /comments/{id}`.
- `GET /notifications?channel&status&page&limit`, `GET /notifications/{id}` (with recipient page `?rPage&rStatus`; each recipient has `lang` and `message`), `POST /notifications` `{channel, title, body, action?: {type, value, label?}, target, providerId?}` → created (queued). Validation: sms body ≤ 600 chars, title not needed for sms; email needs title; `users` ids must exist; group must exist. Response includes the resolved `recipientCount`.
- `GET /notifications/preview-target` `?mode&groupId&userIds&tagIds` → `{count, withPhone, withEmail}`.
- Phase 3: mode `tags` targets every player who has at least one of `tagIds` (player tags, key type `user`, at most 50; banned players skipped); errors `TARGET_TAGS_REQUIRED`, `TOO_MANY_TAGS`, `TAG_NOT_FOUND`. The compose page has a fourth target «برچسب بازیکن» with a player-tag picker and the live recipient count; `#/publication/notifications/new?tagId=<id>` opens it preselected. Phone login refuses a suspended player with the `ACCOUNT_SUSPENDED` notice (see `docs/base/spec.md`, «Players and chat»).
- `GET /providers` → `[{id, channel, kind, isActive, isDefault, sandbox, configured, fields:[…], config:{masked}}]` plus the catalog of available kinds; `POST /providers`, `PUT /providers/{id}`, `DELETE /providers/{id}`, `POST /providers/{id}/test`.
- `GET /users/search?q` → `[{id, username, displayName}]` for the target picker (q matches username/display name/id; max 20).
- `GET /summary` → counts for an overview card.

## Console pages

New sidebar section «انتشار» / "Publishing" after «اقتصاد»: سؤال‌های متداول، اخبار، کامنت‌ها (with pending badge), اعلان‌ها (list + «ارسال اعلان» form with channel, target picker with live count from preview-target, markdown body with preview). Delivery services (provider cards like payment methods, with test send) live in the console's Settings page as the «پیامک» (`#/settings/sms`) and «ایمیل» (`#/settings/email`) tabs; `#/publication/providers` redirects to the SMS tab. Same look, fa/en, dark, mobile cards, server-side search/paging, layout-matching loading states.

## Phone login (phase 2)

Players sign in with a mobile number and a one-time SMS code sent through the providers above. Code in `integrations/publication/otp.go`, migration `20261110040000-publication-otp.sql` (Up re-runnable, Down raises).

### Data

- `publication_otp_settings` (single row `id = 1`, seeded): `is_enabled` (true), `sandbox` (true), `provider_id` (null = the default active SMS provider), `code_length` 4–8 (5), `ttl_seconds` 30–1800 (120), `max_attempts` 1–10 (5), `resend_seconds` 10–600 (60), `phone_hourly` (5), `phone_daily` (10), `ip_hourly` (30), `message` (translated SMS text, ≤300 per language, every translation must contain `{code}`; `{minutes}` = lifetime rounded up, `{app}` = brand name or short name).
- `publication_otp_code`: one row per request: `phone` (E.164), `code_hash`, `status` (`pending`, `verified`, `expired`, `failed` = too many wrong codes, `replaced` = a newer code was asked for, `send_failed`), `attempts`, `max_attempts`, `sandbox`, `provider`, `lang`, `error`, `client_ip`, `requested_by`, `user_id`, `created_at`, `expires_at`, `verified_at`. Codes are never stored: `code_hash` = HMAC-SHA256 over `id|phone|code` with a key derived from `ECONOMY_SECRET_KEY` (HKDF info `publication-otp-code`), so a database copy cannot be brute-forced offline. Rows older than 30 days are deleted every 10 minutes by one node at a time (`clusterkit` lease `publication-otp-sweep`), which also sweeps expired rate-limit counters.
- `publication_phone_account (phone primary key, user_id unique → users on delete cascade, login_key, created_at, verified_at)`.

### Client RPCs (user session; a game signs in as a device guest first)

| RPC | Input | Output |
|---|---|---|
| `publication_otp_request` | `{phone, country?, lang?}` | `{requestId, phone, expiresIn, resendAfter, codeLength, sandbox}` |
| `publication_otp_verify` | `{phone, country?, code, link?}` | `{token, expiresAt, loginKey, userId, username, created, linked, phone}` |
| `publication_phone` | `{}` | `{phone, otpEnabled, codeLength}` (the caller's number, or empty) |

- Number: Persian/Arabic digits, spaces, dashes and brackets are accepted. With `country` (ISO code) and no `+`/`00`, the leading zeros are dropped and the country's dialling code is put in front; Iranian forms (`09…`, `9…`, `989…`) work without it. The result must be E.164 (8–15 digits) and start with the dialling code of an active country (`base_country`), else `INVALID_PHONE` / `PHONE_COUNTRY_NOT_ALLOWED`.
- Limits (database, `clusterkit.Allow`, the same on every node): one request per number per `resend_seconds` (`OTP_TOO_SOON`), `phone_hourly` and `phone_daily` per number, `ip_hourly` per client IP and per calling user (`OTP_RATE_LIMITED`, HTTP 429 / gRPC ResourceExhausted). A new request marks the number's pending code `replaced`.
- Sending: the chosen provider if it is active, else the default active SMS provider. **Sandbox** when the OTP setting or that provider is in sandbox: nothing is sent and the server log gets `Publication OTP sandbox code <code> for <phone> (request <id>, <lang>)`. Live: one synchronous send (25 s timeout, idempotency key `otp-<id>`); a failure marks the row `send_failed` and returns `OTP_SEND_FAILED`. Live without any active provider → `OTP_UNAVAILABLE`; settings off → `OTP_DISABLED`. The text is the translated `message` in the player's language (requested `lang` → account `lang_tag` → default).
- Verify, in one transaction with an advisory lock on the number: the latest code row must be `pending` (`OTP_NOT_FOUND` otherwise, `OTP_TOO_MANY_ATTEMPTS` after a `failed` one) and not expired (`OTP_EXPIRED`, row marked `expired`). A wrong code counts one try and is committed; the last allowed wrong try marks it `failed` (`OTP_INVALID_CODE` / `OTP_TOO_MANY_ATTEMPTS`). A right code marks it `verified`, so it works once.
- Account: a number that already has an account signs into it (`ACCOUNT_DISABLED` for a banned account). Otherwise, with `link: true` the number is attached to the caller's own account (a guest keeps its progress; `ACCOUNT_HAS_PHONE` when it already has another number); without `link` a new Nakama account is created with `AuthenticateCustom`. The account metadata gets `phone` (E.164), which publication SMS notifications already use.
- Session: `token` is a normal session token from `AuthenticateTokenGenerate` (default lifetime). `loginKey` (`ph_` + 40 hex) is the account's custom id when this module set it; a game may keep it and call `authenticateCustom(loginKey, create=false)` for a full session with a refresh token and a silent sign-in next time (treat it like a device id). It is empty for a linked account that already had its own custom id.

### Console

`GET/PUT /v2/console/publication/otp/settings` (full replace; the GET also lists the SMS providers for the picker; errors `INVALID_VALUE`, `PROVIDER_NOT_FOUND`, `PROVIDER_NOT_SMS`, `OTP_MESSAGE_NEEDS_CODE`, translated-text errors) and `GET /v2/console/publication/otp?q&status&page&limit` (recent requests with username, plus `phoneAccounts` and `last24h` counts; `q` matches the number with or without `+`/leading zero; the hash is never returned). ACL `NOTIFICATION` like the rest of publication. The project settings section «ورود با پیامک» is described in `docs/base/spec.md` (Accounts).

## In-app notifications from other plugins (phase 5)

`internal/inapp` holds one sender that the publication plugin registers at start (`inapp.SetSender`). Base calls `inapp.Send` for support replies: the message becomes a normal `in_app` notification to the listed users (translated title and body, optional action), created by `support:<agent>`, delivered by the worker and shown in the inbox like any other. When the publication plugin is not loaded the support code only sends its live notice. A message may ask for `Channel: "push_in_app"` (friend gifts use it when the admin picks «صندوق پیام و پوش»); anything else is `in_app`.

## Banners and placements (phase 10)

Ported from rakhtkan `publishing/banners` and `banner-lists` (placement key and name, banner name, content tags, title, subtitle, description, cover, one action button), extended with scheduling, targeting, ordering, per-language content, in-game actions and counters. Migration `20261117000000-publication-banners.sql`.

### Data

- `publication_banner_list` (placement): `key` (`^[a-z][a-z0-9_]{1,63}$`, the game asks by it, locked after creation), translated `name`/`description`, `layout` (`carousel`, `single`, `grid`, `strip`), `max_items` (1–20), `rotate_seconds` (0–60, 0 = by hand), `is_active`, `sort`. A placement with banners cannot be deleted (`IN_USE`).
- `publication_banner`: `key` (locked), `list_id`, internal `name`, content `tag_ids` (tag key type `banner` or `general`, new key type `banner` in base), translated `title` (required), `subtitle`, `description`, per-language `image_url` (a language without one falls back to the default language), `bg_color` (`#rrggbb`), action (`none`, `screen` + screen name, `product` + product key, `achievement` + achievement key, `url` + per-language http(s) URL) with a translated button `label`, `priority` (−1000…1000, higher first, then oldest first).
- Schedule: `starts_at`/`ends_at` (instants; the console enters and shows them in the project zone from `economy_event_settings.timezone`), `days` (0 = Sunday … 6 = Saturday, empty = every day) and `day_start`/`day_end` (minutes in the project zone, `HH:MM` on the wire, `24:00` allowed as an end).
- Targeting: `audience_tag_ids` (player tags, any one is enough), `exclude_tag_ids` (never shown to these; a tag cannot be in both), `level_table_id` + `min_level`/`max_level` (empty table = the main level table, else the `level` meta), `platforms`, `min_version`/`max_version`, `countries` (ISO codes; the player's saved country, else the request country), `languages` (the player's resolved language), `player_age` `all`/`new`/`returning` against `new_days` (account age).
- A banner whose product is inactive or not visible to the player (product audience tags) or whose achievement is inactive is not shown.
- Counters: `publication_banner_seen` (banner, player, day in the project zone, `viewed_at`, `clicked_at`; no foreign key to `users`, swept after 45 days and for deleted players under the lease `publication-banner-sweep`) and `publication_banner_day` (views and clicks per banner per day). Each player counts at most once per banner per day for a view and once for a click: one statement inserts the seen row with `ON CONFLICT … WHERE viewed_at IS NULL` and bumps the day row only when that insert or update happened, so parallel calls and many nodes never double count.

### Client RPCs (user session)

| RPC | Payload | Answer |
|---|---|---|
| `publication_banners` | `placement`, optional `platform`, `version`, `lang` | `placement` (key, name, layout, rotateSeconds, maxItems) and `items` (id, key, title, subtitle, description, imageUrl, bgColor, priority, endsAt, `action` {type, value or url, label}) for banners the caller may see right now, at most `max_items`. Unknown or inactive placement: `BANNER_LIST_NOT_FOUND`. |
| `publication_banner_view` | `bannerIds` (1–20), `platform`, `version`, `lang` | `counted` (ids counted now), `skipped` (not eligible now), `day` |
| `publication_banner_click` | `bannerId`, `platform`, `version`, `lang` | `counted`, `bannerId` and the `action` from the server; a banner the caller may not see now: `BANNER_NOT_AVAILABLE` |

Each RPC re-checks eligibility on the server. Rate limits in the database (`cluster_rate_limit`): 60 list calls and 60 count calls a minute per player (`TOO_MANY_REQUESTS`). Counting never grants anything; the client gets the action from the server.

### Console API and pages

Generic resources `banner-lists` and `banners` (list with search and filters `isActive`, `layout`, `key`; `listId`, `state` = `live`/`scheduled`/`ended`/`off`, `actionType`, `tagId`), plus `GET /banners/settings` (project zone), `GET /banners/{id}/stats?days=` (daily series in the project zone, totals) and `GET /banner-lists/{id}/preview?userId=&platform=&version=&lang=&at=` (what that player gets now, and the reason for every hidden banner). Writes go to the console audit log like every `/v2/console/publication/` call.

Pages «بنرها» and «جایگاه‌های بنر» in the «انتشار» section (sidebar, Ctrl K): banner list with picture, placement and priority, state and window, a short targeting line and 7-day views/clicks/click rate; the docked editor has a live preview card (follows typing, preview language switch, placement layout), per-language image upload through the economy uploads store, product and achievement pickers, zone-aware start/end inputs, week days and hours, player targeting, and a 14-day views/clicks chart. The placement editor lists its banners in display order and has the «یک بازیکن چه می‌بیند» check.

## Sponsors (phase 12)

Ported from rakhtkan `base/sponsor` (name, key, colour, URL, metadata), `base/ad_position` (key, name, tags) and `base/ads` (an image in several positions with a sponsor, its URL falling back to the sponsor's), and its sponsor wizard (sponsor → positions → ads). rakhtkan ads are banners without schedule, targeting or counting, and its client could only list sponsors; here the banner placements are the ad positions and a banner with a sponsor is the sponsored content, so nothing is built twice. rakhtkan has no sponsored prize, so prizes are not tied to sponsors. Migration `20261119000000-publication-sponsors.sql`.

### Data

- `publication_sponsor`: `key` (locked), translated `name` (required) and `description`, `logo_url`, brand `color` (`#rrggbb`), `website_url` (absolute http(s)), `links` (up to 8: `type` in `website`, `instagram`, `telegram`, `x`, `youtube`, `aparat`, `whatsapp`, `app`, `other`, http(s) `url`, translated `label`), `tag_ids` (tag key type `sponsor` or `general`), private contact (`contact_name`, `contact_email` lower-cased, `contact_phone` normalised like SMS numbers, `notes`; never sent to clients), contract window `starts_at`/`ends_at` (instants, entered in the project zone), `is_active`, `sort`. A sponsor is live when it is active and inside its window.
- `publication_banner.sponsor_id` and `publication_banner_list.sponsor_id` (both optional). A banner whose sponsor is not live is hidden (`sponsor_off` in the player check). New banner action `sponsor`: the button opens the sponsor's website (needs a sponsor with a website: `SPONSOR_REQUIRED`, `SPONSOR_NO_WEBSITE`), like rakhtkan's `resolvedUrl`. A sponsored placement shows «با حمایت …» over all its banners while the sponsor is live.
- Counters: `publication_sponsor_seen` (sponsor, player, day in the project zone, `viewed_at`, `clicked_at`; no foreign key to `users`; swept with the banner rows after 45 days and for deleted players under the lease `publication-banner-sweep`) and `publication_sponsor_day` (views and clicks per sponsor per day). The same single-statement `INSERT … ON CONFLICT … WHERE viewed_at IS NULL` as banners counts each player once per sponsor per day for a view and once for a click, on any number of nodes. A sponsor that is not live is never counted.

### Client RPCs (user session)

| RPC | Payload | Answer |
|---|---|---|
| `publication_sponsors` | optional `lang`, `tagKey` | live sponsors in `sort` order: id, key, name, description, logoUrl, color, websiteUrl, links (type, url, label), endsAt |
| `publication_sponsor_view` | `sponsorIds` (1–20) | `counted`, `skipped` (not live), `day` |
| `publication_sponsor_click` | `sponsorId`, optional `link` (a link type; default the website), `lang` | `counted`, the `url` from the server and the sponsor; `SPONSOR_NOT_AVAILABLE`, `SPONSOR_LINK_NOT_FOUND` |

`publication_banners` items carry `sponsor` (id, key, name, logoUrl, color, websiteUrl) and the placement carries its live `sponsor`; a `sponsor` action has `url` and `sponsorId`. `publication_banner_view` also counts the banner's and the placement's live sponsors (`sponsorsCounted`) and `publication_banner_click` the banner's sponsor (`sponsorCounted`). Rate limits in `cluster_rate_limit`: 60 list and 60 count calls a minute per player (`spn:` keys, `TOO_MANY_REQUESTS`).

### Console API and pages

Generic resource `sponsors` (search in key, name, contact and website; filters `isActive`, `state` = `live`/`scheduled`/`ended`/`off`, `tagId`, `key`; list rows carry 7-day and all-time views/clicks and the banner and placement counts; delete refused while a banner or placement names it: `IN_USE`), `banners` and `banner-lists` take `sponsorId` (and filter by it). `GET /sponsors/{id}/report?days=|from=&to=` (days in the project zone, at most 366: views, clicks and players per day, totals, distinct players, all-time totals, the sponsored banners with their own views/clicks and the hosted placements; `INVALID_DATE`, `INVALID_DATE_RANGE`) and `GET /sponsors/{id}/report.csv` (UTF-8 with BOM: `day,views,clicks,players,ctr_percent`, a `total` line, then the banner rows; cells that start with `= + - @` are prefixed with `'`). Every write goes to the console audit log.

Page «اسپانسرها» in the «انتشار» section (sidebar with the handshake icon, Ctrl K): list with logo, state and window, «۲ بنر / میزبان ۱ جایگاه» (links to the banner list filtered by the sponsor) and 7-day views/clicks; the docked editor has name, description, logo upload, brand colour, website, the link rows, sponsor tags, the private contact block, the contract window in the project zone and the report (7/14/30/90 days, views, clicks, click rate and players tiles, daily bars, banners and placements, «دریافت CSV»). The banner editor has a sponsor picker and the action «باز کردن وب‌سایت اسپانسر»; the placement editor has «اسپانسر جایگاه»; banner and placement rows show «با حمایت …».

## Push, campaigns and automatic notifications (phase 13)

Ported from rakhtkan `base/fcm` (user devices with platform and token, Firebase templates, campaigns with schedule and status, delivery rows, `delivered`/`opened` events, the push ↔ inbox share key) and its admin pages, without its holes: the client can no longer trigger sends or events, every limit is in the database, and every send and every automatic trigger is counted once per player and occurrence on any number of nodes. Push is a fourth channel of the existing notifications (same table, worker, history page and detail page), not a second system. Migration `20261120000000-publication-push.sql`.

### Devices and players' choices

- `publication_push_device`: one row per token (unique; a token that shows up for another player moves to that player), `platform` `android`/`ios`/`web`, `app_version`, optional `device_key`, `is_active`, `fail_count`, `last_error`, `disabled_at`, `last_seen_at`; no foreign key to `users`. At most 10 active devices per player (older ones are switched off); a newer token for the same `device_key` switches the old one off; a token that the provider calls dead (FCM `UNREGISTERED`, `SENDER_ID_MISMATCH`, an `INVALID_ARGUMENT` about the token, or a `sandbox-dead…` token in the sandbox) is switched off at once, five transient failures in a row also switch it off. Rows of deleted players and devices switched off for 30 days are swept under the lease `publication-banner-sweep`.
- `publication_notify_category` (seeded: `system` «پیام‌های مهم» always on, `offers`, `rewards`, `reminders`, `news`): `key`, translated `name`/`description`, `can_opt_out`, `default_on` (a category players cannot turn off must be on by default), `is_active`, `sort`. `publication_notify_pref (user, category, enabled)` holds each player's choice; campaigns and automatic rules skip players who turned the category off (`skipped`, `opted_out`). Test sends ignore the choice.

### Client RPCs (user session)

| RPC | Payload | Answer |
|---|---|---|
| `publication_push_register` | `token` (16–4096, no spaces), `platform`, optional `appVersion`, `deviceId` | `deviceId`, `platform`, `active` |
| `publication_push_unregister` | `token` (only the caller's own) | `removed` |
| `publication_push_devices` | — | the caller's devices with the token masked |
| `publication_notify_prefs` | optional `lang` | active categories with `enabled`, `canOptOut`, and `pushDevices` |
| `publication_notify_prefs_set` | `prefs` {categoryKey: bool} (1–50), `lang` | the same list; `CATEGORY_LOCKED`, `CATEGORY_NOT_FOUND` |
| `publication_notification_opened` | `notificationId` | `counted` (once per player and notification), `unread`; `NOTIFICATION_NOT_FOUND` when the caller did not get it |

Register and unregister share 20 calls a minute per player, prefs 30, opened 60 (`push:` keys in `cluster_rate_limit`). `publication_inbox` now lists `in_app` and `push_in_app` messages, fills the template values, and adds `imageUrl`, `category` {key, name} and `opened`.

### Push providers

`publication_provider` takes channel `push` with two kinds: `fcm` (field `serviceAccount`, the Firebase service account JSON, sealed with the provider key like every other secret, masked as `********` on read, never logged; the JSON is checked on save: `type` = `service_account`, `project_id`, `client_email` and an RSA `private_key`, else `INVALID_SERVICE_ACCOUNT`; the console shows only the project id) and `log` (never sends, always sandbox; the migration adds it active and default). FCM HTTP v1: an RS256 JWT for `https://www.googleapis.com/auth/firebase.messaging` is exchanged at `https://oauth2.googleapis.com/token` (the `token_uri` in the JSON is ignored), the access token is cached in memory until a minute before it expires and refreshed once on 401, and each device gets `POST https://fcm.googleapis.com/v1/projects/{project_id}/messages:send` with `notification` (title, body, https image), `data` (notificationId, lang, actionType, actionValue, category), `android.priority: HIGH`, `apns` mutable content and image, and `webpush.fcm_options.link` for URL actions. 429, 5xx, `UNAVAILABLE`, `INTERNAL` and `QUOTA_EXCEEDED` are retried by the worker with backoff. A provider in sandbox (and `log`) only writes «Publication push sandbox …» with the token masked to the log and stores the text in the recipient row. Tests and the e2e suite only use the sandbox; the FCM code is tested against a local fake server. Test send from the push settings card goes to the devices of one player (username or ID).

### Campaigns

`POST /notifications` takes, besides the old fields, channel `push` or `push_in_app` («پوش و صندوق»: the same message in the device and in the inbox, like rakhtkan's share key), `name` (internal), `imageUrl`, `categoryId` (required for push channels), `audience` (the banner player targeting: player tags in and out, level range with a level table, platforms and app version range matched against the player's registered devices, countries, languages, new or returning, plus `inactiveDays` from the last sign-in), `scheduledAt` (an instant entered in the project zone, at most a year ahead, `SCHEDULE_IN_PAST`) and an action `goto`/`screen`, `product`, `achievement` or `url` (product and achievement keys must exist). A notification with a push channel, a category, filters or a schedule is a `campaign`; automatic ones are `auto`, test sends `test`. `in_app` campaigns take the same filters and schedule; SMS and email do not (`AUDIENCE_NOT_ALLOWED`, `IMAGE_NOT_ALLOWED`).

- Recipients are chosen when the campaign starts: at once, or for a scheduled one by the worker on any node (`status='scheduled' AND scheduled_at <= now()` claimed with `FOR UPDATE SKIP LOCKED`, recipients written and the row moved to `queued` in the same transaction, so it starts once). Players filtered out are not recipients; opted-out players are `skipped` (`opted_out`); a `push` campaign skips players without a device (`no device`).
- Delivery runs in the existing worker (claims with a lease, renewed before each real send). Before a push goes to a device the worker inserts `publication_push_delivery (notification, device)` as `sending` (only a `failed` row may be taken again), so a device never gets the same campaign twice even when a claim expires on a slow node; the row ends as `sent`, `failed` or `dead`. `push_in_app` writes the inbox first (all of a batch in one `NotificationsSend`), then the pushes; it counts as sent even without a device.
- Counts on the notification: recipients, sent, skipped, failed, pending, `openedCount` (from `publication_notification_opened`), `pushSent`/`pushFailed` (devices) and, in the detail, `pushDevices` by delivery status. `POST /notifications/{id}/cancel` stops a scheduled, queued or sending one (`NOT_CANCELLABLE` otherwise): pending recipients become `skipped` (`cancelled`), sent ones stay. `POST /notifications/test` sends the same payload to one player now (`kind` `test`). `POST /notifications/preview` answers the live count for the compose page: `count` (after filters), `targeted`, `optedOut`, `withPush`, `withPhone`, `withEmail`. `GET /notifications` filters by `kind`, `ruleId` and `categoryId` too; `GET /push-devices?userId&platform&active` lists devices with masked tokens.

### Automatic notifications

`publication_auto_rule`: `key`, internal `name`, `trigger`, translated `title`/`body` with `{name}` values, `image_url`, action, channel `push`/`in_app`/`push_in_app`, category (required), `cooldown_minutes` per player (0–525600), quiet hours `quiet_start`/`quiet_end` in the project zone (may cross midnight), `is_active`. Triggers:

| Trigger | Fires | Occurrence key | Values |
|---|---|---|---|
| `event` | for each `economy_player_event` of `event_key` recorded after the rule was made, optionally only when `metadata.match_field = match_value` (works for client, server and built-in events such as `achievement_unlocked`, `ChanceGame:Played`, `level_up`, `login`, `item_bought`; the server records them, the client cannot fake server-only events) | `ev:<event id>` | `{value}` and the event's text/number metadata |
| `streak_at_risk` | in the last `hoursBefore` (1–23, default 4) hours of a streak period (project zone) for players whose streak would break if they miss this period | `streak:<key>:<period>` | `{streak}`, `{count}`, `{hours}` |
| `chance_ready` | when a chance game's free turn has come back (looked back one day) | `chance:<game>:<last free play>` | `{game}` |
| `inactive` | when the last sign-in (or the account creation) is `days` (1–365, default 3) to `days + 2` days ago | `inactive:<days>:<last sign-in>` | `{days}` |

Every trigger also fills `{player}`. One node at a time (lease `publication-auto`, every `PUBLICATION_AUTO_EVERY_SECONDS`, default 10 s) scans: events from a cursor in `publication_auto_state` with two minutes of overlap, the other triggers by query; each hit is inserted into `publication_auto_fire (rule, player, occurrence)` with `ON CONFLICT DO NOTHING`, so a rescan or a second node never adds a second message. Due fires are then taken with `FOR UPDATE SKIP LOCKED` in batches of 500: in quiet hours they wait until the end of the window; a player who opted out (`opted_out`) or got a message from this rule within the cooldown (`cooldown`, also two hits in the same batch) is skipped; the rest become one `auto` notification per rule and batch, with the values stored per recipient (`vars`) and filled in for the inbox, the in-app message and the push. Fires are kept 90 days. Console: generic resource `auto-rules` (sent, skipped and pending in 7 days, opened, last sent), `GET /auto-rules/{id}/fires?status=` and `POST /auto-rules/{id}/test {userId}` (sample values, kind `test`).

### Console pages

- «اعلان‌ها (ارسال و سابقه)» gains: channels «پوش» and «پوش و صندوق» with the push provider and its sandbox note, «دسته و تصویر» (internal name, category picker, picture upload), more actions (product, achievement), «فیلتر بازیکنان» (the banner targeting fields plus «دست‌کم این تعداد روز نیامده»), «زمان ارسال» (now or a time in the project zone, with a confirmation), the live count with opted-out and device counts, and «ارسال آزمایشی» to one player. The list filters by kind and shows scheduled times and opens; the detail shows kind, name, category, rule, filters, picture, opened and push tiles, recipients' opened time and pushes, and «لغو ارسال».
- «اعلان‌های خودکار»: list with the rule sentence, channel, category, repeat gap and quiet hours, 7-day numbers; the docked editor has the trigger fields with a live sentence, the message with the values hint, picture and action, limits, the numbers, the latest fires with their reasons and the test send.
- «دسته‌های اعلان»: categories with the player choice, opted-out count and use.
- Project settings section «پوش» (`#/project-settings/push`, `NOTIFICATION`): provider cards like SMS and email; the FCM card takes the JSON in a text area and shows only the project id afterwards.

## News: workflow, webhooks, reactions and SEO (phase 25)

The news section is generic: admins write articles in the console and other services push them through per-source webhooks. Code in `integrations/publication/news*.go`, migrations `20261130000000-publication-news-ingest.sql` and `20261130000001-publication-news-seo.sql` (Up re-runnable, Down raises). Every policy below is an admin setting with the default in brackets.

### Data

- `publication_news` gains `status` (`draft` | `pending_review` | `published` | `archived` | `rejected`; old inactive rows became `draft`, `is_active` stays as `status='published'` for old readers), `expires_at`, `source` (`console` or `webhook:<key>`), `source_id`, `external_id` (unique with `source`), `source_url`, `content_hash`, `lang` (language the item arrived in), `body_format` (`markdown` for old rows, `html` for the rich editor and webhooks), `tag_ids` (news and general tags; `tagUsedSQL` counts them), `allow_reaction`, `notify_mode` (`none` | `in_app` | `push` | `both`), `notify_audience` (the campaign audience: player tags in and out), `notified_at` / `notify_claimed_until` / `notification_id`, `approved_by` / `approved_at`, `reject_reason`, `created_by` / `updated_by`, and SEO: `slug` (unique), `meta_title`, `meta_description` (translated), `keywords`, `canonical_url`, `og_image`, `noindex`.
- Visible to players: `status='published' AND published_at <= now()` and not expired (expired ones only when «خبر تمام‌شده قابل دیدن» is on).
- Rich bodies: the console editor is Tiptap 2.27.2 (`@tiptap/core` directly, the overlay is vanilla JS, no React; bundled once by `console/ui/tiptap` with esbuild into `console/ui/economy/vendor/tiptap.js`, no CDN). Bodies are stored as HTML per language and cleaned on the server (`news_html.go`) to exactly the editor's allowlist: `p`, `h1`–`h6` (only `style="text-align: left|right|center|justify"`), `strong`, `em`, `u`, `s`, `code`, `pre`, `blockquote`, `hr`, `br`, lists (`ol start`), task lists (`ul data-type="taskList"`, `li data-type="taskItem" data-checked`, a disabled checkbox), tables (`table`, `thead`, `tbody`, `tr`, `th`/`td` with `colspan`, `rowspan`, `colwidth`), links (`http`, `https`, `mailto`, `tel`, `/…`, always `target="_blank" rel="noopener noreferrer nofollow"`). Scripts, styles, frames, images, forms and every other attribute are removed; unknown wrappers keep their text. Old markdown bodies render as before and open in the editor converted to HTML.
- `publication_news_source`: `key` (in the address, locked), `name`, `is_active`, `auth_mode` [`token`], header names [`X-News-Token`, `X-News-Signature`, `X-News-Timestamp`], `skew_seconds` [300], `secret` (sealed with the publication key, generated when empty, only a hint is listed), `items_path`, `mapping` (JSON paths per field: title, summary, body, cover, category key, tag keys, lang, publishedAt, expiresAt, externalId, url, slug, metaTitle, metaDescription, keywords, canonicalUrl, ogImage, noindex), `body_format` [`auto`], default category, language and tags, `publish_mode` [`review`], `update_mode` [`update`], `published_at_mode` [`source`], `rate_limit_per_minute` [60], `copy_images` [off], `notify_mode` [`none`], `allow_comment`, `allow_reaction`.
- `publication_news_ingest_log` (every request and console test with per-item results and the body up to the configured size), `publication_news_replay` (HMAC signatures for twice the skew), `publication_news_reaction_type` (seeded `like` 👍, `love` ❤️, `haha` 😂, `wow` 😮, `sad` 😢), `publication_news_reaction` (one per player and article), `publication_news_settings` (single row).

### Webhook ingest

`POST /v2/publication/news/ingest/{key}` on the client API port (no session). Order: active source (else 404) → body ≤ 1 MB (413) → per-source rate limit in `cluster_rate_limit` (429 with `Retry-After`, not logged) → auth (401, logged without the body) → JSON → items (`items_path`, or the body itself, or its array; at most 50) → each item in its own transaction.

- `token`: the secret in the token header (`Authorization: Bearer …` when the header is `Authorization`), constant-time compare. `INVALID_TOKEN`.
- `hmac`: `hex(HMAC-SHA256(secret, timestamp + "." + body))` in the signature header (optional `sha256=` prefix), unix seconds (or ms) in the timestamp header within ±skew (`TIMESTAMP_OUT_OF_WINDOW`), each signature accepted once (`REPLAYED`, 409).
- Mapping: title required (`ITEM_TITLE_REQUIRED`); HTML bodies are cleaned to the allowlist, markdown is rendered then cleaned, text becomes paragraphs; the incoming language maps to an active `base_language` (`fa-IR` → `fa`) or the source default; with «پر کردن زبان پیش‌فرض» [on] the default language gets the same text so admins can edit it. Unknown category or tag keys are noted (`CATEGORY_UNKNOWN`, `TAGS_UNKNOWN`), never created.
- Dedupe: `(source, external_id)`; without an id the content hash is the id. Same hash → `unchanged`; new content → `updated` (status kept) or `ignored` by `update_mode`; new → `created` as `pending_review` or `published` by `publish_mode`. An advisory lock per item makes parallel deliveries safe.
- Remote cover copy (per source, off): only http(s), public addresses (private, loopback and link-local refused), at most 2 MB, png/jpg/webp, stored through `internal/objectstore` in the economy upload space and served from `/v2/economy/files/`. A failure keeps the original URL and notes `COVER_COPY_FAILED`.
- Answer: `{results:[{index, externalId, status, newsId, code, notes}], created, updated, unchanged, ignored, failed}`.

### Notify on publish

A loop on every node (every 5 s) claims up to 20 visible, not yet notified articles with `FOR UPDATE SKIP LOCKED` and a two-minute `notify_claimed_until` lease, then per article locks the row, re-checks `notified_at`, creates one notification (kind `news`, `news_id`, target all players with the article's audience, channel from `notify_mode`, the category from settings [`news`], title and summary or body text, action `goto news/<id>`, cover as push picture) and sets `notified_at` in the same transaction. So the notice goes out exactly once, when the publish time passes or on approval; a refused campaign (for example no push provider) is logged and the article is still marked. After commit, with «پیام زنده هنگام انتشار» [on], `NotificationSendAll` sends a non-persistent socket notice (code `1320`, subject `news`, `{event:"news_published", newsId, title}`) so online clients refresh. The same loop sweeps the ingest log [30 days] and old replay rows under the lease `publication-news-sweep`.

### Client RPCs (user session, all plain RPCs)

| RPC | Payload | Answer |
|---|---|---|
| `publication_news_list` | `limit`, `cursor`, `lang`, `categoryKey`, `tagKey` / `tagKeys` + `tagMode` (`any`/`all`), `expired` (`exclude`/`include`/`only`, the last two only when the admin allows), `unreadOnly`, `includeTags` [true], `includeReactions` [true], `includeSeo` [false] | items with `expiresAt`, `expired`, `slug`, `tags:[{key,name,iconUrl}]`, `reactions:{key:n}`, `reactionCount`, `myReaction`, `allowReaction`, `allowComment`, `sourceUrl`, optional `seo`; `reactionTypes` |
| `publication_news_get` / `publication_news_by_slug` | `id` or `slug`, `lang` | the article with `bodyFormat`, `bodyHtml`, tags, reactions, `seo` |
| `publication_news_unread` | `categoryKey`, `tagKeys` | `unread` (from sign-up or all, by setting) |
| `publication_news_seen` | `id`, `ids` (≤ 200) or `all:true` | `unread` |
| `publication_news_react` | `newsId`, `reaction` (key; empty removes) | `reactions`, `reactionCount`, `myReaction`; `REACTION_NOT_FOUND`, `REACTION_NOT_ALLOWED`, `NEWS_EXPIRED`, `REACTION_RATE_LIMITED` [30/min per player on all nodes] |
| `publication_news_reaction_types` | `lang` | active reaction types |
| `publication_news_comment_delete` | `id` | deletes the caller's own comment |

Comments keep their behaviour; the per-minute limit is now a setting [5].

`seo` = `{slug, metaTitle (falls back to the title), metaDescription (summary, then the first 160 characters of the body), keywords, canonicalUrl, ogImage (falls back to the cover), noindex}`.

### Public pages (off by default)

With «صفحه‌های عمومی اخبار» on and a website address set: `GET /v2/publication/public/news?lang&limit&cursor&categoryKey&tagKey`, `GET /v2/publication/public/news/{slug}?lang` (body HTML), `GET /v2/publication/public/sitemap.xml` (`<loc>` = canonical or base + path template [`/news/{slug}`], `<lastmod>`) and, when RSS is on, `GET /v2/publication/public/rss.xml` (latest 50, site title from settings). Only published, running, indexable (`noindex` off) articles; `Cache-Control: public, max-age=<cache>` [300]; 404 while off. Slugs are made from the default-language title on create when empty, in the admin style [`native`: Persian letters kept, Arabic ي/ك normalized, Persian digits to Latin; `latin`: Persian transliterated; `id`: short id], made unique with `-2`, `-3`; a manual slug must be letters and digits joined by `-` (`INVALID_SLUG`) and unique (`SLUG_EXISTS`, 409).

### Console API (`/v2/console/publication/`)

`news` CRUD gains the fields above and the filters `status`, `source`, `sourceId`, `tagId`, `expired`, `slug` (search covers slug and external id); `POST /news-review {ids, action: approve|reject|archive, reason}`; `GET|PUT /news-settings`; `news-sources` CRUD, `POST /news-sources/{id}/test {payload, save}` (preview without saving, or a real ingest marked as a test), `POST /news-sources/{id}/secret/reveal`, `POST /news-sources/{id}/secret/rotate` (both logged); `GET /news-ingest-log?sourceId&status&isTest`, `GET /news-ingest-log/{id}` (with body); `news-reactions` CRUD (`IN_USE` once players used one; switch it off instead).

### Console pages

«اخبار» is a tabbed page: خبرها (status, source, tags, expiry, reactions and comments in the list; filters for status, category, tag, ending and pinned; the editor with the Tiptap toolbar (bold, italic, underline, strike, headings, lists, task list, link, alignment, table and its row/column/header tools, undo/redo, fa/en labels, RTL or LTR per language tab), status and reject reason, publish and end time, comments and reactions switches, the publish notice with player tags, an SEO section with a Google-style result preview, slug, meta title and description with counters, keywords, canonical, share image and noindex, and the origin of webhook items), صف بررسی (checkboxes, bulk approve and reject with a reason), منبع‌ها (endpoint and a curl sample per auth mode, secret reveal and rotate, mapping, defaults, publishing, rate limit, image copy, test payload with per-item preview), گزارش دریافت (filters, counts, side panel with results and body), واکنش‌ها, تنظیمات.
