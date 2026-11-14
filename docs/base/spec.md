# Base module: languages, categories, tags, translations

Port of asp `Modules/Base/{Language,LanguageItem,Category,Tag}` (`/Users/mazyar/Projects/asp/backend/src/PolicePatrol.Api/Modules/Base`). Goal: any game built on this server can be fully multilingual. Every user-facing title, name, description or body in every module is translatable into every active language.

## Decisions

- New Go plugin `integrations/base` → `base.so`, owning languages, categories and tags. Economy and Publication read its tables (same DB) and drop their own category/tag tables after migrating the data.
- Translations live on the owning row as `jsonb` maps keyed by language code (`{"fa": "…", "en": "…", "ar": "…"}`), not in a separate LanguageItem table. Same capability as asp, atomic with the row, no orphan rows.
- The console's own UI language (EN/FA button) is unrelated and stays as it is.
- Jazebe content is out of scope.

## Languages (`base_language`)

`code text primary key` (BCP 47 primary subtag or `xx-YY`, lower-case, e.g. `fa`, `en`, `ar`, `tr`, `pt-br`), `name jsonb` (translated names, e.g. `{"fa":"فارسی","en":"Persian"}`), `native_name text` (`فارسی`), `direction text check in ('rtl','ltr')`, `icon_url`, `is_active bool`, `is_default bool`, `sort int`, `created_at`, `updated_at`.

- Exactly one default, and it must be active. Seed: `fa` (default, rtl) and `en` (ltr), both active.
- The default cannot be deactivated or deleted. A language can be deleted only when inactive; its translations stay in the rows (harmless) and are dropped on the next save of each row.
- Changing the default never deletes data; rows missing a value in the new default are reported by `GET /v2/console/base/languages/coverage`.

## Translated text rules (shared Go helper used by every module)

- Type `LocalizedText = map[string]string`. Validation on write: keys must be existing language codes, values trimmed, the default language value required for required fields, per-field max length per language.
- Resolution for clients: requested `lang` → user's account `lang_tag` (only if it is an existing active language) → default language → first non-empty value. Clients may send `lang` on every RPC; responses return resolved strings plus `lang` (the language actually used for that field set is not needed per field).
- Console API always reads and writes the full maps.
- Delivery (in_app, SMS, email): each recipient gets text in its own resolved language (`lang_tag` → default). Nakama sets `lang_tag` to `en` on new accounts; a client RPC `base_set_language {lang}` updates the caller's `lang_tag` (must be an active language) so a game can store the player's choice.

## Fields that become `LocalizedText`

Already maps (keep, but generalise from fixed fa/en to any language): currency name, meta name, prize name + message, product name + description, category name, tag name, faq question + answer, news title + summary + body + action labels, notification title + body + action label, language name.

Newly translatable: gift `name` (was plain text) and category/tag `description` (new). Rich text (faq answer, news body, notification body, email) stays the restricted Markdown, per language.

## Categories and tags (`base_category`, `base_tag`), like asp

`id uuid, key text, key_type text, name jsonb, description jsonb, icon_url, metadata jsonb, sort int, created_at, updated_at`, unique `(key_type, key)`.

- `key_type` (asp KeyType) says which module/entity uses it. Registered key types: `product`, `prize`, `gift`, `currency`, `faq`, `news`, `task`, `item`, `item_type`, `achievement` (categories; `item_type` categories are the item types of the economy inventory, since `20261115000000-economy-items.sql`; `achievement` since `20261116000000-economy-achievements.sql`, console label «دستاورد» / "Achievement"; tags also take `task`, `item`, `achievement` and `banner` (since `20261117000000-publication-banners.sql`, «بنر» / "Banner") and `sponsor` (since `20261119000000-publication-sponsors.sql`, «اسپانسر» / "Sponsor")); tags may use `product`, `prize`, `gift`, `currency` (since `20261110050000-economy-gift-currency-taxonomy.sql`), `news`, `faq`, `general` (usable everywhere) or `user` (player tags, phase 1; migration `20261110020000-base-tag-user-key-type.sql` widens the CHECK). `GET /v2/console/base/{categories|tags}/key-types` returns the list with translated labels (`user` = «بازیکن» / "Player"). Player tags never reach game clients: `base_tags` refuses `keyType: user` and leaves them out of the unfiltered list; economy refuses them on products and prizes (`TAG_KIND_MISMATCH`).
- Migration copies `economy_category` (kind → key_type) and `economy_tag` (key_type `general`) preserving ids, repoints references (product/prize/gift/faq/news `category_id`, `tag_ids`), then drops the economy tables. Re-runnable; Down raises.
- Delete returns 409 `IN_USE` when referenced, except tags, which are removed from `tag_ids` (current economy behaviour).

## Console API (`/v2/console/base/`, console session, audit-logged)

ACL: `languages` → `SETTINGS` read/write; `categories`, `tags` → `HIRO_ECONOMY` read/write (same as economy today). Same generic CRUD shape as economy (`GET list ?q&page&limit[&keyType]`, `GET /{id}`, `POST`, `PUT /{id}`, `DELETE /{id}`) for `languages` (id = code), `categories`, `tags`. Plus `GET languages/coverage` → per module/resource count of rows missing each active language.

The old `/v2/console/economy/{categories,tags}` endpoints are removed; the console moves to the new ones.

## Client RPCs

| RPC | Input | Output |
|---|---|---|
| `base_languages` | `{}` | `{default, items:[{code, name (resolved), nativeName, direction, iconUrl}]}` active only |
| `base_set_language` | `{lang}` | `{lang}` |
| `base_categories` | `{keyType, lang?}` | `{items:[{key, name, description, iconUrl}]}` |
| `base_tags` | `{keyType?, lang?}` | same shape |

All existing economy and publication client RPCs accept `lang` and resolve every translated field with the rules above.

## Console pages

- Project settings window («تنظیمات پروژه» / "Project settings", phase 1, like the asp `SettingsModal`): one large dialog (`min(1400px, 97vw)` × `86vh`, RTL, current console look) with the section list on the start side (icon + label, `role="tablist"`) and the panel on the other. Sections come from one registry, `PS_SECTIONS` in `console/ui/economy/src/settings.js` (id, label, icon, ACL resource, render); later phases add a line there. Current order: brand, setup, languages, countries, account rules, SMS login, app versions, categories, tags, payment gateways, SMS, email. Each section reuses the page it replaced (no copies).
  - Route: `#/project-settings/<section>[/<id>][?query]` (a Vue route like the other overlay sections, so it survives a reload; `<id>` opens the item editor, the query keeps list filters). Opening pushes one history entry; switching sections replaces it, so browser back closes the window. Close (×, Esc, click outside) goes back to the page the window was opened from (kept in `sessionStorage` for reloads); an unknown section goes to the first allowed one.
  - Entry points: «تنظیمات پروژه» with a gear at the bottom of the sidebar (above the Hiro/Satori links, icon-only when collapsed) and a gear button in the top bar next to the account menu. Both open the first section the user may see.
  - ACL: the overlay decodes the `acl` bitmap in the console session token (bit `resource*3 + level`, all-`0xFF` = admin; a token without the claim shows everything, the server still enforces). Brand, setup, languages, countries, account rules and app versions need `SETTINGS` read; categories, tags, payment gateways need `HIRO_ECONOMY` read; SMS login, SMS and email need `NOTIFICATION` read. Hidden sections are not listed and their deep links redirect; with no section allowed both entries disappear and a deep link shows a "no access" note.
  - Keyboard and a11y: focus moves to the active section, Tab is trapped in the window, arrows/Home/End move between sections, Esc closes the item editor or a picker first and the window last. At ≤767px the window is a full-screen sheet with the section list as a sideways scroller of chips.
  - Moved here: the settings tabs we used to inject into Nakama's `#/settings` (payment gateways, SMS, email, languages, brand, setup) and the «عمومی» sidebar section (categories, tags). Nakama's own settings tabs are untouched. Old hashes redirect (replace) with their id and query: `#/settings/{payment-methods|sms|email|languages|brand|setup}`, `#/base/{categories|tags}`, `#/economy/{categories|tags|payment-methods}`, `#/publication/providers` (→ SMS). Overview, coverage and compose links and the wizard copy point to the new place.
- Languages section: list with default badge, active toggle, direction, order; form with code, native name, translated names, icon; coverage panel (how many items lack each language, linking to the page).
- Categories and tags sections: key type filter and field (tags include «بازیکن»), icon upload, description.
- Pickers (asp `TaxonomyPickers`): every category picker asks the server for its form's key type, every tag picker for the key type plus `general`, with server-side search. Next to each picker (products, prizes, FAQ, news) a «+» opens a small quick-create popover: type a name in the default language, «بساز» creates it with that key type and a key slugged from the name (Persian letters transliterated, a numbered suffix on `ALREADY_EXISTS`, a fallback key for names with no letters), selects it at once and toasts; errors stay inline. Gifts and currencies have no category field in the backend, so they have no picker.
- Every translatable field in every form shows one tab per active language (default first, marked), a dot on tabs whose value is empty, RTL/LTR input direction per language, and the Markdown preview per language. List pages show the name in the console's UI language if present, else the default language.
- Fix: digits inside inputs follow the field, not the UI language (codes, phone numbers, sender numbers stay Latin in inputs); read-only display follows the UI language consistently.

## Implementation (as built)

| Part | Path |
|---|---|
| Go plugin | `integrations/base/` → `base.so` (no runtime env needed) |
| Shared translated-text helper | `internal/localized` (`Text`, `Set.Clean`, `Set.Resolve`, `Set.Choose`, `ForUser`); imported by base, economy and publication |
| Migrations | `20261101000000-base-schema.sql` (tables + seed fa/en), `20261101000001-base-move-categories.sql` (copy economy categories/tags with ids, repoint `category_id` FKs of prize/product/faq/news, drop `economy_category`/`economy_tag`), `20261101000002-economy-gift-name-localized.sql`, `20261101000003-publication-recipient-lang.sql` (recipient `lang`, `message`). Up re-runnable, Down raises. |
| Console auth/audit | `server/console.go` (`/v2/console/base/` branch), `server/console_base_audit.go` (languages → `SETTINGS`, categories/tags → `HIRO_ECONOMY`), `console/acl/acl.go` |

Language codes: `^[a-z]{2,3}(-[a-z0-9]{2,8})?$`, input is lower-cased and `_` becomes `-`. `name` max 64 per language (default language required), `nativeName` required (max 64), `direction` `rtl`|`ltr` (default `ltr`). Category/tag `key` `^[a-z0-9][a-z0-9_.-]{0,63}$`, `name` max 128 (required), `description` max 1000, `iconUrl` http(s) or `/…`, `metadata` object.

Console JSON:

- Language: `{id (= code), code, name:{<lang>:text}, nativeName, direction, iconUrl, isActive, isDefault, sort, createdAt, updatedAt}`. `POST /languages` body `{code, name, nativeName, direction, iconUrl, isActive (default true), isDefault, sort}`; `PUT /languages/{code}` same body without `code` (full replace; send `isDefault: true` for the current default). Making a language default unsets the old one in the same transaction. List `?q&page&limit&isActive`, ordered default first, then `sort`, `code`.
- Errors: `INVALID_LANGUAGE_CODE`, `ALREADY_EXISTS` (409), `LANGUAGE_NOT_FOUND` (404), `DEFAULT_LANGUAGE_LOCKED` (409: deactivate or delete the default), `DEFAULT_LANGUAGE_REQUIRED` (400 on `isDefault:false` for the default, field `isDefault`), `DEFAULT_MUST_BE_ACTIVE`, `LANGUAGE_ACTIVE` (409: delete an active language), `INVALID_DIRECTION`, `INVALID_NATIVE_NAME`.
- `GET /languages/coverage` → `{default, languages:[active codes, default first], missing:{<lang>: total}, items:[{module, resource, fields:[…], total, missing:{<lang>: count}}]}`. A row counts as missing a language when any of its translated fields has some value but none in that language. Resources: base languages/categories/tags, economy currencies/metas/prizes/products/gifts, publication faqs/news.
- Category/tag: `{id, key, keyType, name, description, iconUrl, metadata, sort, createdAt, updatedAt}`; list `?q&page&limit&keyType` (q matches key and every translation), newest first. Errors `INVALID_KEY_TYPE`, `INVALID_KEY`, `ALREADY_EXISTS` (409, unique per key type), `IN_USE` (409 delete of a used category; 400 key type change of a used record, field `keyType`; a tag may always move to `general`).
- `GET /{categories|tags}/key-types` → `{items:[{key, name:{fa,en}}]}`.

Translated-text errors on every module: 400 `INVALID_LANGUAGE` (key is not an existing language code), `DEFAULT_LANGUAGE_REQUIRED`, `TEXT_TOO_LONG`, each with `field`. Clients must send only codes from `GET /languages` (drop keys of deleted languages before saving).

Client RPC outputs: `base_languages` → `{default, lang, items:[{code, name, nativeName, direction, iconUrl, isDefault}]}` (`name` falls back to `nativeName`); `base_set_language` → `{lang}` or `LANGUAGE_NOT_AVAILABLE`; `base_categories` (`keyType` required, else `INVALID_KEY_TYPE`) and `base_tags` (`keyType` optional; a key type also returns `general` tags) → `{lang, items:[{key, keyType, name, description, iconUrl, metadata, sort}]}` ordered by `sort`, `key`.

## Branding and setup wizard

Per-server console branding plus a first-login setup wizard. Both live in the base plugin; the wizard writes only through the existing APIs (base languages, economy currencies and payment methods, publication providers), so everything it sets stays editable from the normal pages.

### Data (`20261110000000-base-brand-setup.sql`, Up re-runnable, Down raises)

- `base_brand` (single row, `id = 1`): `name jsonb` (translated), `short_name`, `logo_url`, `logo_dark_url`, `mark_url` (favicon and collapsed-sidebar mark), `accent` (`''` or `#rrggbb`), `welcome jsonb` (translated sign-in text), `updated_by`, `updated_at`. Empty row = the stock Nakama look.
- `base_setup` (single row, `id = 1`): `status` (`pending` | `in_progress` | `completed` | `skipped`), `steps jsonb` (`{brand|languages|currency|payment|sms|email: "done"|"skipped"}`; missing = not touched), `first_opened_by/at`, `completed_by/at`, `updated_by/at`. Seeded `pending`, so an upgraded server shows the wizard once to the first admin who signs in.

### Console API

| Route | ACL | Notes |
|---|---|---|
| `GET /v2/console/base/brand` | `SETTINGS` read | full brand incl. `updatedBy` |
| `PUT /v2/console/base/brand` | `SETTINGS` write | full replace; `{}` resets to defaults. Body `{name, shortName (≤24), logoUrl, logoDarkUrl, markUrl, accent, welcome}`; `name` ≤64 and `welcome` ≤280 per language, translated-text rules apply (default language required once any value is set). Errors `INVALID_COLOR`, `INVALID_URL`, `TEXT_TOO_LONG`, `INVALID_LANGUAGE`, `DEFAULT_LANGUAGE_REQUIRED` with `field`. Accent accepts `#rgb`/`#rrggbb`, stored lower-case `#rrggbb`. |
| `GET /v2/console/base/brand/public` | none | the only unauthenticated console route: `server/console_public.go` allow-lists exact path + GET/HEAD and strips identity headers. Non-secret fields only, `Cache-Control: public, max-age=60, stale-while-revalidate=300`, `ETag` + `If-None-Match` → 304. |
| `GET /v2/console/base/setup` | `SETTINGS` read | `{status, steps, nextStep, stepOrder, firstOpenedBy/At, completedBy/At, updatedBy/At}`; `nextStep` is the first step not done/skipped, else `review`. |
| `PUT /v2/console/base/setup` | `SETTINGS` write | `{status?, steps?}`; steps merge (`"pending"` clears one). `completed` stamps `completedBy/At`; `skipped` and `in_progress` never downgrade a `completed` setup; `pending` is a full reset (clears steps and stamps, so the wizard auto-opens again). Any non-pending write on a never-opened setup stamps `firstOpenedBy/At`. |
| `POST /v2/console/base/setup/claim` | `SETTINGS` write | atomic `UPDATE … WHERE status='pending' RETURNING`: exactly one caller ever gets `{claimed: true}`; everyone else gets `{claimed: false, setup}`. Audited as `INVOKE`. |

Brand and setup writes are audit-logged under `SETTINGS`.

### Client RPC

`base_brand {lang?}` → `{lang, custom, name, shortName, logoUrl, logoDarkUrl, markUrl, accent, welcome, updatedAt}` with `name`/`welcome` resolved by the usual language rules; `custom: false` and empty strings when nothing is set (games fall back to their own defaults).

### Console behaviour

- Brand is read from the public endpoint (cached in `localStorage` for a flash-free start) and applied before and after sign-in: sidebar logo (light/dark variants, mark when collapsed), sign-in logo, sign-in heading «برای ادامه به داشبورد {name} وارد شوید» plus the welcome text, `document.title` (`Page | {name} (host)`), favicon (= mark), and the accent. The accent overrides `--primary-color-50…950` (palette mixed from the hex like the UI kit does) and rewrites inline `--x-*` colors the native bundle hardcodes from the default palette to `rgb(var(--primary-color-N))`; the overlay's own buttons, toggles, tabs and inputs use the variables directly. Native ECharts series colors keep the stock purple.
- The brand form shows a live light/dark preview and WCAG contrast for white text on the accent and the accent link on a dark background (≥4.5 readable, ≥3 large text only).
- Project settings sections «برند» (`#/project-settings/brand`) and «راه‌اندازی» (`#/project-settings/setup`: status, progress, per-step state, who opened/finished, «باز کردن راهنما», per-step edit links). The wizard opens above the settings window and refreshes the setup section when it closes.
- Auto-open: once per page load per session token, when the signed-in shell is visible, the console calls `setup/claim`; only `claimed: true` opens the wizard. Opening it by hand on a `pending`/`skipped` setup marks it `in_progress`, so it never auto-opens afterwards.
- Wizard: full-screen overlay (`data-economy="setup"`), clickable step bar (compact numbers on mobile), «قبلی», «رد کردن این مرحله», «ذخیره و ادامه», «تمام» on review, «بعداً» (status `skipped`, progress kept; Esc asks first). Focus is trapped in the overlay; leaving a step with unsaved changes asks first. Steps: brand; languages (toggle active, make default, quick add fa/en/ar/tr/ru/de/fr/es/ur/zh, all saved at once); currency (pick existing or new, key/type/name/starting balance/icon; other fields kept on save; Latin digits in the number input); payments («پرداخت لازم دارید؟», pick gateways, sandbox box first, new gateways default to sandbox); SMS and email («… لازم دارید؟», service, sandbox, fields with masked secrets, saved as active + default; test-send saves first and is only allowed in sandbox); review (status per step, summaries, edit links). Answering «نه» marks the step skipped.

## Accounts (phase 2)

Account rules, app versions and countries live in the base plugin; phone + SMS login lives in the publication plugin because it owns the SMS providers and their secrets (`docs/publication/spec.md`, «Phone login»). Migration `20261110030000-base-accounts.sql` (Up re-runnable, Down raises) creates the tables and seeds the defaults, so a deploy needs no manual step.

### Account rules (`base_account_rules`, single row `id = 1`)

`username_min` (1), `username_max` (128), `username_charset` (`any` | `latin` = `[A-Za-z0-9_.-]`), `password_min` (8, never below Nakama's own 8), `password_letter`, `password_digit`, `display_name_max` (255), `updated_by/at`. The defaults are Nakama's own limits, so an upgraded server refuses nothing new until an admin tightens the rules. Lengths count letters (runes), not bytes.

Enforced by before-hooks registered in `integrations/base/rules.go`:

| Hook | Checks |
|---|---|
| `AuthenticateEmail` | only when the email has no account yet (sign-up): password rules, then the username if one is sent |
| `AuthenticateDevice/Custom/Google/Apple/Facebook/Steam/GameCenter` | the username, only if it is sent and nobody has it yet (an existing account signing in is never refused) |
| `UpdateAccount` | a changed username; the display name length |
| `LinkEmail` | password rules |

Errors are runtime errors with codes `USERNAME_TOO_SHORT`, `USERNAME_TOO_LONG`, `USERNAME_INVALID_CHARACTERS`, `PASSWORD_TOO_SHORT`, `PASSWORD_NEEDS_LETTER`, `PASSWORD_NEEDS_DIGIT`, `DISPLAY_NAME_TOO_LONG` (gRPC InvalidArgument, HTTP 400 with the code as `message`). Nakama-generated usernames (sign-ups without a username, phone accounts) are not checked. Only one before-hook per message exists in Nakama, so since phase 3 these hooks are registered through `internal/hookchain` (see «Players and chat»); a game plugin that registers its own hook through `hookchain.Before` runs after these checks instead of replacing them.

Console: `GET/PUT /v2/console/base/account-rules` (`SETTINGS` read/write, full replace, `INVALID_RULE` with `field`). Client RPC `base_account_rules {}` → `{usernameMin, usernameMax, usernameCharset, passwordMin, passwordLetter, passwordDigit, displayNameMax}`.

### App versions (`base_app_version`, one row per platform)

`id`, `platform` (`android`, `ios`, `web`, `windows`, `macos`, `linux`; unique), `latest_version`, `min_version` (optional, never newer than latest), `store_links jsonb [{store, url, label:{<lang>: text}}]` (stores `google_play`, `app_store`, `bazaar`, `myket`, `microsoft_store`, `steam`, `web`, `direct`; each once, max 8; url `http(s)://`, `market://`, `bazaar://`, `myket://` or `itms-apps://`), `update_message` and `force_message` (translated, ≤500), `is_active`.

Versions are 1–4 numeric parts with an optional `-pre` or `+build` (`^v?\d+(\.\d+){0,3}([-+][0-9A-Za-z.-]+)?$`); compared part by part (missing parts are 0), a pre-release is older than its release, build metadata is ignored.

Console: CRUD `/v2/console/base/app-versions` (`SETTINGS`; list `?q&page&limit&isActive`, ordered by platform) and `GET /app-versions/catalog` (platforms and stores with fa/en names). Errors `INVALID_PLATFORM`, `INVALID_VERSION`, `VERSION_REQUIRED`, `MIN_ABOVE_LATEST`, `INVALID_STORE`, `DUPLICATE_STORE`, `INVALID_URL`, `TOO_MANY_LINKS`, `ALREADY_EXISTS` (409, platform), translated-text errors.

Client RPC `base_app_version {platform, version, store?, lang?}` → `{lang, platform, current, status, latest, minimum, message, storeUrl, stores:[{store, name, url}]}`. `status`: `update_required` when `version < min`, `update_available` when `version < latest`, else `ok` (also when the platform has no active row). `message` is `force_message` (falls back to `update_message`) or `update_message`, resolved by the usual language rules; `name` is the link label or the built-in store name; `storeUrl` is the link for `store`, else the first link. `INVALID_PLATFORM` / `INVALID_VERSION` on bad input.

### Countries (`base_country`)

`code` (ISO 3166 alpha-2, upper-case, primary key), `name jsonb` (translated), `dial_code` (`+98`), `flag_url` (optional image; otherwise the flag emoji made from the code), `is_active`, `sort` (100; Iran is seeded with 1). The migration seeds 243 countries with fa, en and ar names (from CLDR) and dialling codes (libphonenumber metadata), all active. A name key for a language that does not exist is dropped on save.

Console: CRUD `/v2/console/base/countries` (`SETTINGS`; id = code; list `?q&page&limit&isActive`, `q` matches code, dialling code and every translation; items carry `flag`). Errors `INVALID_COUNTRY_CODE`, `INVALID_DIAL_CODE`, `COUNTRY_NOT_FOUND`, `ALREADY_EXISTS`.

Client RPCs: `base_countries {lang?}` → `{lang, selected, items:[{code, name, dialCode, flag, flagUrl}]}` (active only, by `sort` then name; `selected` = the player's account metadata `country`); `base_set_country {code}` → `{code}` stores the choice in account metadata `country` (`COUNTRY_NOT_AVAILABLE` for an unknown or inactive country). SMS login accepts only numbers whose dialling code belongs to an active country.

### Console sections

Four new project settings sections (registry `PS_SECTIONS`, code in `console/ui/economy/src/accounts.js`), in this order after languages:

- «کشورها» (`#/project-settings/countries[/<code>]`, `SETTINGS`): list with flag, name, code, dialling code, inline active toggle and order; active filter and search; editor with code (locked on edit), translated name, dialling code, flag image, active, order.
- «قوانین حساب» (`#/project-settings/accounts`, `SETTINGS`): one form (username lengths and letters, password length and letter/digit, display name length) with a «امتحان کنید» box that checks a sample username and password against the numbers in the form before saving.
- «ورود با پیامک» (`#/project-settings/otp`, `NOTIFICATION`, same as SMS): a notice with the current mode (off / sandbox / live with which service / no service), the settings form (on, sandbox, SMS service, code digits, lifetime, tries, resend wait, per-number hourly and daily limits, per-IP hourly limit, translated SMS text with `{code}`, `{minutes}`, `{app}`) and «درخواست‌های اخیر» (number, status, sandbox or service, tries, player link, time; search by number, status filter). The code itself is never shown.
- «نسخه‌های برنامه» (`#/project-settings/app-versions[/<id>]`, `SETTINGS`): list with platform, latest, lowest allowed, store tags and status; editor with platform, versions, a store-link list (store, link, optional translated label) and the two translated messages.

## Players and chat (phase 3)

Player tags, temporary suspensions and mutes, warnings and chat moderation live in the base plugin. Migration `20261111000000-base-players-chat.sql` (Up re-runnable, Down raises) adds the tables and seeds six report reasons and ten filter words, so a deploy needs no manual step; `20261111000001-economy-product-audience.sql` adds `economy_product.audience_tag_ids`.

### Data

| Table | What |
|---|---|
| `base_tag.is_public` | new column; only meaningful for key type `user`. A public player tag can be read by the player it is given to. |
| `base_user_tag(user_id, tag_id, created_by, created_at)` | tags given to players; both foreign keys cascade on delete |
| `base_restriction` | one row per suspension, mute or warning: `kind` (`suspend`, `mute`, `warn`), `status` (`active`, `lifted`, `expired`), `starts_at`, `ends_at`, `reason` (translated, shown to the player), `note` (internal), `report_id`, `created_by`, `lifted_by/at`, `lift_note`, `ended_at`. At most one active suspension and one active mute per player (partial unique index); a new one lifts the old one (`lift_note = 'replaced'`). A warning is stored as `expired` with `ends_at = starts_at`. |
| `base_chat_word` | filter words: `word` (1–64 letters), `action` (`mask`, `block`, `flag`), `match` (`word`, `contains`), `is_active`, `note` |
| `base_report_title` | report reasons: `key`, translated `name` and `description`, `is_active`, `sort`. Seeded: abuse, harassment, spam, cheating, personal-info, other (fa, en, ar). |
| `base_chat_report` | one row per report: `source` (`player` or `filter`), `status` (`open`, `dismissed`, `resolved`), the message id, channel id and room label, a snapshot of the sender, the message text and up to five messages before and after it (`context`), the reporter, the reason (id and a snapshot of its name), the player's note, matched filter words, and the decision (`action`, `restriction_id`, `message_deleted`, `resolved_by/at`, `resolution_note`). A player reports a message once; the filter flags a message once. |

Restrictions are enforced by `status = 'active' AND ends_at > now()`, so a restriction ends exactly on time on every node even before any worker runs. Every node runs `restrictionLoop` (5 s): it closes the live sessions of players suspended in the last two minutes (`SessionLogout` plus `SessionDisconnect` of the presences on that node, so every node cleans up its own sockets without shared memory), and every 30 s the node that holds the `base-restriction-expire` lease (`internal/clusterkit`) marks ended rows `expired`. Nakama's own permanent ban (`disable_time`) is untouched and shown as «مسدود دائمی».

### Enforcement

- Sign-in: before-hooks for `AuthenticateDevice` (by device id), `AuthenticateCustom` (by custom id), `AuthenticateEmail` (by email or username) and `SessionRefresh` (user id read from the refresh token) refuse a suspended player. Phone login (`publication_otp_verify`) refuses too. Google, Apple, Facebook, Steam and Game Center sign-ins cannot be checked before Nakama verifies the token, so their after-hooks log the new session out at once (the client then gets 401 without the message; next device or custom sign-in shows it).
- The refusal is a runtime error with gRPC code 7 (HTTP 403) whose message is JSON: `{"code":"ACCOUNT_SUSPENDED","kind":"suspend","until":"<RFC 3339>","reason":"<in the player's language>","lang":"fa"}`.
- Chat: a before-hook on `ChannelMessageSend` and `ChannelMessageUpdate` refuses muted (and suspended) players with the same JSON shape and code `CHAT_MUTED`, then runs the word filter over every string in the message content (JSON object). `block` refuses the message with the plain message `CHAT_BLOCKED`; `mask` replaces the matched letters with `*`; `flag` lets the message through and the after-hook stores a `filter` report with the matched words. The word list is read from the database for each message (no cache).
- Hooks are composed through `internal/hookchain`: every plugin that registers a hook through it adds to one chain per message instead of replacing the previous hook (Nakama keeps one hook per message). Base registers its account-rule, suspension and chat hooks this way; a game plugin that wants its own `ChannelMessageSend` or `AuthenticateDevice` hook should call `hookchain.Before` too.
- Live notices (code, content): 1102 suspended (`{code, kind, until, reason, lang, restrictionId}`, not persisted, sent just before the sockets are closed), 1103 muted (not persisted), 1104 warning (persisted, `code: "WARNED"`), 1105 lifted (`{code: "LIFTED", kind}`).

### Word filter (`internal/chatfilter`)

Before matching, both the words and the text are normalised: NFKD with combining marks removed (Arabic diacritics, madda and hamza above, Latin accents, presentation forms such as `ﺍﺣﻤﻖ`, full-width Latin), ZWNJ/ZWJ/bidi marks/soft hyphen/tatweel removed, `ي ى` → `ی`, `ك` → `ک`, `ة ە ھ` → `ه`, Persian and Arabic digits → Latin, lower case, a few Cyrillic and Greek look-alikes → Latin, leet before a Latin letter (`0→o 1→i 3→e 4→a 5→s 7→t @→a $→s !→i`), repeated letters collapsed (a word letter repeated n times matches n or more in the text). Text is split two ways (by spaces and punctuation, and by spaces only with punctuation and digits dropped), and runs of single letters are joined (`ا ح م ق`, `i.d.i.o.t`).

- `word` match: the whole token (or two adjacent tokens, for `بي شعور` or `id iot`), optionally followed by a common suffix (Persian `ی ای ها های ان ت ش م مان تان شان تر ترین ه`, English `s es ed er ers ing y ish`). Multi-word entries (`خفه شو`) match consecutive tokens.
- `contains` match: anywhere in the text with separators removed (`حرام‌زاده‌ای`, `ح ر ا م ز ا د ه`).
- Masking keeps spaces and punctuation and drops invisible marks inside the match, so `تو اح‌ـمق هستی` becomes `تو **** هستی`.

### Console API (`/v2/console/base/`)

ACL: `players`, `restrictions` → `ACCOUNT` read/write; `chat-words`, `report-titles`, `reports` → `CHANNEL_MESSAGE` read/write (resolving a report with mute or suspend needs only `CHANNEL_MESSAGE` write). Audit-logged under those resources (`/lift`, `/resolve` and `/chat-words/test` as invoke).

| Route | Notes |
|---|---|
| `GET /players?q&tagId&status&page&limit` | `q` matches username, display name or id; `status` = `suspended`, `muted`, `banned`, `tagged`. Items `{id, username, displayName, avatarUrl, createTime, banned, tags:[{id,key,name,isPublic}], suspension:{id,endsAt,reason}|null, mute:…}` |
| `GET /players/{id}` | the same plus `history` (last 50 restrictions) and `reports {against, by, open}` |
| `PUT /players/{id}/tags` | `{tagIds}` full replace; only key type `user` (`TAG_KIND_MISMATCH`), at most 50 |
| `POST /players/tags` | bulk `{userIds (≤1000), tagIds, op: add|remove}` → `{changed, players}`; `USERS_NOT_FOUND`, `TAGS_REQUIRED`, `INVALID_OP` |
| `POST /players/{id}/restrictions` | `{kind, minutes | endsAt (RFC 3339), reason (translated, required, ≤500), note}`; `END_REQUIRED`, `END_IN_PAST`, `END_TOO_FAR` (10 years; use the ban for longer), `INVALID_RESTRICTION_KIND`, `USER_NOT_FOUND`. Sends the live notice and, for a suspension, closes the player's sessions. |
| `GET /restrictions?q&kind&status&userId` | `status` = `active`, `expired` (also active rows past their end), `lifted`; `activeTotal` |
| `GET /restrictions/{id}`, `POST /restrictions/{id}/lift` | `{note}`; `NOT_ACTIVE` (409) once ended |
| `chat-words` CRUD | `{word, action, match, isActive, note}`; items carry `normalized`; duplicates by normalised form → `ALREADY_EXISTS` (409, field `word`); `INVALID_WORD`, `INVALID_MOD_ACTION`, `INVALID_MATCH` |
| `POST /chat-words/test` | `{text}` → `{text (as others see it), normalized, changed, blocked, flagged, hits:[{id, word, action}]}` |
| `report-titles` CRUD | `{key, name, description, isActive, sort}`; items carry `used` |
| `GET /reports?q&status&source&titleId&senderId`, `GET /reports/summary` | list with `openTotal`; summary `{open, player, filter}` |
| `GET /reports/{id}` | report plus `context`, `others` (other reports on the same message), `sender` (player view plus `counts`) |
| `POST /reports/{id}/resolve` | `{action: dismiss|delete|warn|mute|suspend, deleteMessage, minutes|endsAt, reason, note}`; closes every open report on that message; `delete` (or `deleteMessage`) removes the message with `ChannelMessageRemove`; `REPORT_CLOSED` (409) when already handled |

Tag definitions stay under `/tags` (`HIRO_ECONOMY`); a tag of key type `user` accepts `isPublic` and its list items carry `players`. A user tag that players have cannot change its key type (`IN_USE`); deleting a tag removes it from players and from product audiences.

### Client RPCs

| RPC | Input | Output |
|---|---|---|
| `base_my_tags` | `{lang?}` | `{lang, items:[{key, name, description, iconUrl, …}]}` public player tags of the caller only |
| `base_my_status` | `{lang?}` | `{lang, suspension: notice|null, mute: notice|null, warnings:[{id, reason, createdAt}]}` (warnings of the last 7 days, at most 5) |
| `base_report_titles` | `{lang?}` | `{lang, items:[{key, name, description}]}` active reasons |
| `base_report_message` | `{messageId, titleKey, note? (≤500)}` | `{reportId, duplicate}`; `REPORT_TITLE_NOT_FOUND`, `MESSAGE_NOT_FOUND` (also when the caller cannot see the channel), `CANNOT_REPORT_OWN`, `TOO_MANY_REPORTS` (20 per hour per player, `clusterkit.Allow`) |

### Console pages

- New sidebar section «بازیکنان و گفتگو» / "Players & chat" (rail group with a shield icon, top-bar tabs, Ctrl K page search), pages hidden by ACL:
  - «مدیریت بازیکنان» (`#/moderation/players`, `ACCOUNT`): search, tag filter, status filter, a checkbox per row and a bulk bar (select this page, add or remove tags); a row opens the native player page.
  - «گزارش‌ها» (`#/moderation/reports[/<id>]`, `CHANNEL_MESSAGE`, open-count badge): status, source, reason and player filters; the review page shows the conversation around the message with the reported line marked, the report (reason, reporter, note, matched words, other reports), the sender (tags, status, counts, links to history and reports) and the decision: dismiss, delete message, warn, mute or suspend with duration chips (1 h, 1 d, 3 d, 7 d, 30 d or a date), a translated message for the player prefilled per action, «پیام هم پاک شود» and an internal note.
  - «تعلیق و سکوت» (`#/moderation/restrictions`, `ACCOUNT`): kind, status and player filters, reason, who and when, «برداشتن» with a note.
  - «فیلتر کلمات» (`#/moderation/chat-words[/<id>]`, `CHANNEL_MESSAGE`): «امتحان فیلتر» box (what others see, the normalised form, the hits) above the word list; editor with word, action, match, active and note.
- Project settings section «عنوان‌های گزارش» (`#/project-settings/report-titles`, `CHANNEL_MESSAGE`).
- Native player page (`#/players/<id>`, account tab): a «برچسب‌ها و محدودیت‌ها» card above the account card with the tags («ویرایش برچسب‌ها» dialog), the current status, «تعلیق موقت», «بی‌صدا کردن», «اخطار» or «برداشتن …» buttons (ACCOUNT write) and links to the history and reports.
- Tag editor: «بازیکن این برچسب را ببیند» for player tags; the tag list shows «دیدنی برای بازیکن» and a player count linking to the filtered player list.
- Sidebar placement no longer depends on the native «بازیکنان» link, so a console user without `ACCOUNT` access (for example a chat moderator) still gets our sections.

## Support, remote config and player activity (phase 5)

Ported from asp `Modules/Base/Support` (tickets, departments, canned replies, live chat) with the same statuses, priorities, SLA timestamps and access rules; asp has no remote config and only an HTTP audit log, so those two are new here. Migrations `20261112000000-base-support.sql` and `20261112000001-base-remote-config-activity.sql` (Up re-runnable, Down raises) seed two departments (`general`, `payments`), three canned replies and the support settings row, so a deploy needs no manual step. Star Catcher seeds two config keys.

### Tickets

| Table | What |
|---|---|
| `base_support_settings` (row 1) | open tickets per player (10), new tickets per hour (6), player messages per hour (40), attachments per message (5) and size (2048 KB), auto-close days (7), live chat on/off, greeting and offline text (translated), idle close minutes (30), chat messages per minute (12), activity retention days (180) |
| `base_support_department` | `key`, translated `name`/`description`, `first_response_hours` (SLA, 24), `agents` (console usernames; a new ticket goes to the least busy one), `is_active`, `sort` |
| `base_support_canned` | `key`, translated `title`/`body`, optional `department_id` (null = every department), `is_active`, `sort` |
| `base_support_ticket` | `number` (`T-` + 6 letters/digits), player, subject, department, `priority` (`low`, `normal`, `high`, `urgent`), `status` (`open`, `awaiting_staff`, `awaiting_player`, `resolved`, `closed`), `assignee`, `source` (`player`, `chat`, `staff`), counters and last message, `last_player_message_at`, `last_staff_message_at`, `player_seen_at`, `due_at`, `first_response_at`, `resolved_at`, `closed_at`, `rating` 1–5, `rating_comment`, `rated_at` |
| `base_support_message` | `kind` (`message`, `note` = internal, `event` = audit line), `author_type` (`player`, `staff`, `system`), body (≤5000), `attachments` `[{key, name, type, size}]`, `event` (`status_changed`, `assigned`, `department_changed`, `priority_changed`, `rated`, `created_from_chat`) |
| `base_support_upload`, `base_support_upload_part`, `base_support_file` | chunked player uploads (rows in the database until finished), then the file in `internal/objectstore` (`support/` prefix; local default `<data_dir>/support-uploads`) |

Rules (from asp): a player reply keeps `open` or moves to `awaiting_staff` (reopens a solved ticket); a staff reply moves to `awaiting_player` by default, or `resolved`/`closed`, assigns the replier when nobody owns the ticket and sets `first_response_at`; a note changes nothing the player sees. Staff moves: `resolved`/`closed` from anything but closed, `awaiting_staff` (reopen) only from solved or closed, `awaiting_player` only from open or waiting for staff (`INVALID_TRANSITION`). `due_at` = last player message (or creation) + department hours while the ticket waits for staff; overdue when past. A rating is allowed once, after the ticket is solved or closed. Every node runs `supportLoop` (30 s); the node that holds the `base-support-autoclose` lease closes idle chats, and every 10 minutes closes tickets waiting for the player or solved for `auto_close_days`, removes unfinished uploads and files never attached after a day, and sweeps old activity.

Attachments: png, jpg, webp and pdf found by content (`ATTACHMENT_TYPE`), at most `max_attachment_kb` (`ATTACHMENT_TOO_LARGE`) and `max_attachments` per message (`TOO_MANY_ATTACHMENTS`). Players upload in 96 KB chunks through RPCs (the RPC body limit is 256 KB); agents upload with multipart `POST /support/uploads`. Files are private: every response carries a URL `/v2/support/files/<key>?e=<unix>&s=<hmac>` signed with a key derived from `ECONOMY_SECRET_KEY` (HKDF info `support-file-url`), valid for an hour and served with `nosniff` and a sandbox CSP; a player can attach only their own files. Without `ECONOMY_SECRET_KEY` attachments are off (`ATTACHMENTS_UNAVAILABLE`).

Player notification: a staff reply, solve or close sends a live notice (code 1210, `{type:"ticket", ticketId, number, event}`, not persisted) and, through `internal/inapp`, an in-app notification in the publication inbox (translated title and body, action `goto` `support/ticket/<id>`, created by `support:<agent>`).

### Live support chat

`base_live_chat` (one open chat per player by a partial unique index: `waiting` → `active` → `closed`) and `base_live_chat_message` (`seq` per chat; `player`, `agent` or `system` with an event: `greeting`, `accepted`, `transferred`, `closed`). `base_support_agent` keeps each console user's last heartbeat and away flag; an agent counts as online for 75 s after a heartbeat.

- Take: one `UPDATE … WHERE status='waiting'`; the losers get `ALREADY_TAKEN` (409, `field` = the winner). Messages, typing and transfer need the chat to be active and owned (`NOT_YOUR_CHAT`). Transfer goes to another known agent or back to the queue, under the chat row lock, so two transfers cannot both win.
- Close by the agent, the player or the idle sweep; «تبدیل به تیکت» makes a `chat` ticket whose first message is the transcript, assigned to the agent. The player can rate a closed chat once if an agent took it.
- Realtime: every state lives in the database. The player gets live notices (code 1211, `{type:"chat", chatId, event: message|typing|status, seq}`) through Nakama notifications and also polls `base_support_chat_get` every 3 s while the chat screen is open. Console agents long-poll `GET /support/chats/{id}?after=<seq>&wait=15&stamp=<stamp>`: the handler re-reads the row every 600 ms and answers as soon as the stamp (last seq, status, owner, typing, online, read seq) changes, so it works on any node. Typing shows for 5 s after the last keystroke; a player counts as online for 30 s after the last chat call.

### Client RPCs

| RPC | Input | Output / errors |
|---|---|---|
| `base_support_info` | `{lang?}` | departments (translated), attachment limits and types, `chatEnabled`, `agentsOnline`, greeting and offline text, `unread` (tickets with a staff reply not seen), `openTickets`, `chatId` |
| `base_support_tickets` | `{lang?}` | the caller's last 50 tickets, open first, with `unread` and `canRate` |
| `base_support_ticket_create` | `{subject, departmentKey?, body, attachments?, lang?}` | the ticket with messages; `INVALID_SUBJECT`, `BODY_REQUIRED`, `DEPARTMENT_NOT_FOUND`, `TOO_MANY_OPEN_TICKETS`, `TOO_MANY_TICKETS` |
| `base_support_ticket_get` | `{id, lang?}` | ticket and messages (no notes, no assignment events); marks it seen; `TICKET_NOT_FOUND` for someone else's ticket |
| `base_support_ticket_reply`, `_resolve`, `_rate` | `{id, body, attachments}`, `{id}`, `{id, rating, comment}` | `TICKET_CLOSED`, `TOO_MANY_MESSAGES`, `TICKET_NOT_FINISHED`, `ALREADY_RATED`, `INVALID_RATING` |
| `base_support_upload_begin`, `_part`, `_finish` | `{name, size}`, `{uploadId, index, data (base64 ≤96 KB)}`, `{uploadId}` | `{uploadId, chunkBytes, parts}`, `{index}`, the attachment `{key, name, type, size, url}`; `UPLOAD_NOT_FOUND`, `UPLOAD_INCOMPLETE`, `TOO_MANY_UPLOADS` (30 an hour) |
| `base_support_chat_open`, `_get`, `_send`, `_typing`, `_close`, `_rate` | `{topic?, body?, departmentKey?}`, `{chatId?, after?}`, `{chatId, body}`, `{chatId}`, `{chatId, toTicket?}`, `{chatId, rating, comment?}` | `{chat: {id, status, agent, agentTyping, lastSeq, messages, ticketId, ticketNumber, canRate, rating}, agentsOnline}`; `CHAT_DISABLED`, `CHAT_CLOSED`, `TOO_MANY_CHATS`, `CANNOT_RATE` |
| `base_remote_config` | `{platform?, version?, keys?, lang?}` | `{lang, values, variants, hash}` |
| `base_my_activity` | `{limit?, kind?, before?}` | the caller's own timeline (sign-ins, purchases, prizes, events, tickets, chats, referrals, restrictions) |

### Remote config

`base_config_key`: `key` (`^[a-z][a-z0-9_.]{1,63}$`), `type` (`string`, `number`, `bool`, `json`, `text` = translated), `description`, default `value`, ordered `rules` `[{name, value, platforms, minVersion, maxVersion, tagIds (player tags), countries, percent}]`, `is_active`, `version`. The first rule whose conditions all fit gives its value (variant = rule name), otherwise the default (variant `default`). Country is the account's `country` (phase 2), else a `CF-IPCountry`-style header. Percent rollout uses `fnv64a(key + ":" + userId) % 10000 < percent × 100`, so a player stays in the same bucket, raising the percent only adds players, and different keys bucket independently. Every create, edit, delete and rollback writes a full snapshot to `base_config_history`; rollback restores a version as a new version (a deleted key comes back with its id). The type cannot change (`TYPE_LOCKED`). When a player's resolved values change between two full fetches, a `config/changed` activity row lists the changed `key:variant` pairs (the first fetch is not a change).

### Player activity

`base_activity` (one row per thing a player did or that happened to them: `kind`, `action`, `ref_id`, `data`, `actor` (`player`, `system`, `console:<name>`), `ip`, `country`, `platform`, time). Written through `internal/activity` by every plugin: sign-ins (after-hooks for device, custom, email, Google, Apple, Facebook, Steam, Game Center, and phone login in publication; platform from `X-Client-Platform` or the user agent, country from a CDN header or the account), purchases, prizes from events, levels, streaks, referrals and gift codes, level ups, restrictions and lifts, reports filed and received, tickets, live chats, referrals and config changes. Rows older than `activity_retention_days` are deleted in batches of 5000 under the `base-support-autoclose` lease.

### Console API (`/v2/console/base/`)

ACL: `support/*` → `ACCOUNT_NOTES` read/write (so a support agent needs no economy or player management access; the native console has no support resource, and account notes are the closest one); `remote-config*` → `SETTINGS`; `activity*` → `ACCOUNT`. Audit-logged under those resources (`take`, `transfer`, `close`, `status`, `assign` as invoke; heartbeats, typing and read marks are not logged).

| Route | Notes |
|---|---|
| `GET /support/summary`, `POST /support/heartbeat {away?}` | badge counts (`awaitingStaff`, `waitingChats`, `overdue`, `unassigned`, `mine`, `myChats`, `agentsOnline`, `avgFirstResponseMinutes` (30 days), `avgRating`, `rated`); the heartbeat also returns them |
| `GET /support/agents`, `GET /support/players?q` | console users with online state and open work; player search for a new ticket |
| `support/departments`, `support/canned` CRUD, `GET/PUT /support/settings` | department delete refuses while it has open tickets (`IN_USE`) |
| `GET /support/tickets?q&status(waiting|active|…)&priority&departmentId&assignee(me|none|name)&userId&overdue&sort(due|priority)` | list |
| `POST /support/tickets`, `GET /support/tickets/{id}` | staff ticket for a player (`awaiting_player`, first response set); detail with notes, events and player counts |
| `POST /support/tickets/{id}/messages {body, internal, status, attachments}`, `/status {status}`, `/assign {assignee, departmentId, priority}` | |
| `POST /support/uploads` | multipart `file` |
| `GET /support/chats?view(queue|mine|active|closed)&userId`, `GET /support/chats/{id}?after&wait&stamp` | queue ordered by waiting time; long poll |
| `POST /support/chats/{id}/take`, `/messages`, `/typing`, `/read {seq}`, `/transfer {to}`, `/close {toTicket}` | |
| `remote-config` CRUD, `GET /remote-config/history`, `GET /remote-config/{id}/history`, `POST /remote-config/{id}/rollback {version}`, `POST /remote-config/preview {userId?, platform, version, country, tagIds, lang}` | preview returns values, variants and the player's bucket per key |
| `GET /activity?userId&kind(comma list)&action&country&platform&from&to&q`, `GET /activity/summary` | |

### Console pages

- New sidebar section «پشتیبانی» / "Support" (rail group with a headset icon, top-bar tabs, Ctrl K page search; badges: tickets waiting for staff, chats in the queue):
  - «تیکت‌ها» (`#/support/tickets[/<id>]`): stat strip (waiting, overdue, unassigned, mine, average first answer, average rating), filters, «تیکت تازه» dialog with a player search. The ticket page shows the thread (player and staff bubbles, yellow internal notes, event lines, attachments with previews), the composer (canned reply picker in the ticket language, attachment, «یادداشت داخلی», status after the reply), and a side card (player links, status moves, owner, «خودم برمی‌دارم», department, priority, SLA times, rating, the chat it came from).
  - «گفتگوی زنده» (`#/support/live-chat[/<id>]`): «آمادهٔ جواب دادن» switch (heartbeat with away), views queue / mine / all / closed refreshed every 4 s with waiting time, unread counts and «برداشتن»; the conversation with online dot, typing dots, composer with canned replies, «سپردن به همکار» (agent list or back to the queue) and «بستن گفتگو» with «گفتگو به تیکت تبدیل شود».
- «فعالیت بازیکنان» (`#/moderation/activity`, `ACCOUNT`): kind, player, platform, country and date filters, search, one row per activity with a translated sentence, where (country, platform, IP) and who.
- «پیکربندی بازی» (`#/publication/remote-config[/<id>]`, `SETTINGS`): key list with value preview, rule count and version, recent changes; the editor has the typed value input, the rule list (name, share of players, platforms, version range, player tags, countries, value; reorder), history with «برگشت به این نسخه» and «امتحان برای یک بازیکن».
- Project settings: «بخش‌های پشتیبانی», «پاسخ‌های آماده», «تنظیمات پشتیبانی» (`ACCOUNT_NOTES`).
- Native player page: «فعالیت اخیر» card (last eight items, kind chips, links to the full log and the player's tickets).
- Economy and publication pages now need `HIRO_ECONOMY` / `NOTIFICATION` read to show in the sidebar (they were shown to everyone and answered 403), so a support agent sees only «پشتیبانی».

## Friends and player reports (phase 17)

### Friend request rules (17a)

Nakama core keeps friends in `user_edge` (state 0 friend, 1 invite sent, 2 invite received, 3 blocked). The base plugin adds rules on top through hooks composed with `internal/hookchain` (a game plugin that adds its own friend hooks through `hookchain` runs after these). Migration `20261124000000-base-friend-rules.sql` (Up re-runnable, Down raises) creates the tables and the single settings row, so a deploy needs no manual step.

| Table | What |
|---|---|
| `base_friend_rules` (row 1) | `daily_requests` (20; 0 = no limit, ≤1000), `max_friends` (0 = no limit, ≤10000), `expiry_days` (0 = never, ≤365), `reject_cooldown_hours` (0 = off, ≤8760), `updated_by/at`. Only the daily limit is on by default. |
| `base_friend_request_day(user_id, day, count)` | new requests per player per day; `day` is the date in the project time zone (`economy_event_settings.timezone`, UTC when missing), so the count restarts at local midnight |
| `base_friend_rejection(requester_id, rejecter_id, rejected_at)` | written when a player deletes a friend edge that was an invite they received (state 2), i.e. declines it |

Hooks (`integrations/base/friends.go`):

| Hook | What |
|---|---|
| before `AddFriends` | resolves `ids` and `usernames` (at most 100, `FRIEND_TOO_MANY_USERS`) and sorts them into new invites (no edge yet, not blocked by the other side), accepts (the caller holds a received invite) and no-ops (already friends, already invited, blocked). Expires the caller's old invites first. Then: max friends (the caller cannot send while full or accept past the limit; accepting also checks the sender's list, `FRIEND_TARGET_LIMIT_REACHED`), the cooldown after a rejection by any new target, and the daily limit as one atomic `INSERT … ON CONFLICT DO UPDATE … WHERE count + n <= limit` on the day row, so parallel calls on many nodes can never pass the limit. A request with more new invites than are left is refused whole. Accepting an invite and no-ops never count. |
| after `AddFriends` | clears rejections the caller made of these players (reaching out lifts the other side's cooldown) |
| before `DeleteFriends` | records a rejection for every received invite being deleted |
| before `ListFriends` | expires the caller's sent and received invites older than `expiry_days` |

Expiry is lazy on read and also swept: every node runs `friendsLoop` (1 min), the node that holds the `base.friends.expire` lease (`internal/clusterkit`) removes up to 500 expired invites through `nk.FriendsDelete` (so Nakama keeps its edge counts right), day rows older than three days and rejections older than the cooldown (at least 30 days). Removing an expired invite is not a rejection. Known limit: max friends is read before Nakama writes the edge, so two accepts by the same player at the same instant could pass it by one.

Errors are runtime errors whose message is JSON the client can show as is: `{"code", "message" (in the player's language: fa, en or ar built in, with Persian digits in fa), "lang", "limit"?, "remaining"?, "resetsAt"?, "retryAt"?, "userId"?}`.

| Code | gRPC / HTTP | Extra fields |
|---|---|---|
| `FRIEND_DAILY_LIMIT` | 8 / 429 | `limit`, `remaining`, `resetsAt` (next local midnight, RFC 3339) |
| `FRIEND_LIMIT_REACHED` | 9 / 400 | `limit` |
| `FRIEND_TARGET_LIMIT_REACHED` | 9 / 400 | `limit`, `userId` |
| `FRIEND_REQUEST_COOLDOWN` | 9 / 400 | `retryAt`, `userId` |
| `FRIEND_TOO_MANY_USERS` | 3 / 400 | `limit` (100) |

Console: `GET/PUT /v2/console/base/friend-rules` (`SETTINGS` read/write, audit-logged, full replace, `INVALID_RULE` with `field`). Client RPC `base_friend_rules {lang?}` → `{dailyRequests, sentToday, remainingToday (null when unlimited), resetsAt, maxFriends, friends, expiryDays, rejectCooldownHours, timeZone}`.

Project settings section «قوانین دوستی» (`#/project-settings/friends`, `SETTINGS`), after «قوانین حساب»: requests per day, most friends, request expiry in days, wait after a rejection in hours, each with «صفر یعنی بی‌حد» hints, and a live summary sentence under the form.

### Player reports (17b)

A player can report another player, not tied to a chat message, from any game or screen. The reasons are the same report titles as chat reports (`base_report_title`, «عنوان‌های گزارش»; the title's `used` count adds both kinds). Migration `20261124000001-base-player-reports.sql` (Up re-runnable, Down raises).

`base_player_report`: reporter and reported player (ids kept with a username snapshot; `ON DELETE SET NULL`), reason (`title_id`, `title_key` and a snapshot of the translated name), the player's `note` (≤1000 letters), optional context `match_id`, `content_id` (≤128) and `game_key` (≤64), `status` (`open`, `reviewing`, `resolved`, `dismissed`), `reviewer` and `review_started_at`, internal `notes` (`[{by, at, text}]`, at most 50), the decision (`action` `dismiss|warn|mute|suspend`, `restriction_id`, `resolved_by/at`, `resolution_note`).

Rules on the server: no self-report (`CANNOT_REPORT_SELF`), the reported player must exist (`USER_NOT_FOUND`), the reason must be active (`REPORT_TITLE_NOT_FOUND`), context without control characters (`INVALID_CONTEXT`). One open or reviewing report per reporter, player and reason within 24 hours: under a transaction advisory lock on that triple, a repeat answers the existing id with `duplicate: true` (parallel repeats on many nodes make one row). New reports are limited to 10 per hour per reporter through `clusterkit.Allow` inside the same transaction (`TOO_MANY_REPORTS`, 429); a duplicate does not use the limit. Each new report adds a `report` activity row for both players (`filed`, `received`, data `kind: player`).

Client RPCs:

| RPC | Input | Output |
|---|---|---|
| `base_report_player` | `{userId, titleKey, note?, matchId?, contentId?, gameKey?}` | `{reportId, duplicate}` |
| `base_my_reports` | `{lang?, page?, limit? (≤50)}` | `{lang, items:[{id, userId, username, displayName, titleKey, title, note, matchId, contentId, gameKey, status, createdAt, closedAt}], total, page, limit}`; the reporter sees the status, never the action or the notes |

Console API (`/v2/console/base/player-reports`, `CHANNEL_MESSAGE` like chat reports, audit-logged; `/status`, `/notes` and `/resolve` as invoke):

| Route | Notes |
|---|---|
| `GET /player-reports?q&status&titleId&targetId&reporterId&gameKey` | `status` also takes `active` (open or reviewing); `openTotal` counts active reports |
| `GET /player-reports/summary` | `{open, reviewing, active}` |
| `GET /player-reports/{id}` | the report, `others` (last 20 reports against the same player) and `target` (player view plus `counts {warn, mute, suspend, chatReports, reports}`) |
| `POST /player-reports/{id}/status` | `{status: reviewing|open}` takes the report for review (records the console user) or puts it back; `INVALID_REPORT_STATUS`, `REPORT_CLOSED` (409) |
| `POST /player-reports/{id}/notes` | `{text}` (≤1000) appends an internal note; `TEXT_REQUIRED`, `TOO_MANY_NOTES` |
| `POST /player-reports/{id}/resolve` | `{action: dismiss|warn|mute|suspend, minutes|endsAt, reason, note, closeOthers}`; warn, mute and suspend go through the same restrictions as chat reports (translated reason, live notice, sign-out for a suspension) with `report_id` set to this report; `closeOthers` closes every other open or reviewing report against the same player with the same decision; `REPORT_CLOSED` (409) when already handled |

`GET /players/{id}` adds `playerReports {against, by, open}`.

Console: page «گزارش بازیکن‌ها» (`#/moderation/player-reports[/<id>]`, `CHANNEL_MESSAGE`, badge with open and reviewing reports) after «گزارش‌ها» in «بازیکنان و گفتگو»: filters (status with «باز یا در حال بررسی» by default, reason, player id), rows with the reported player, the player's note, context chips (game, match, content), reason with the count against the player, reporter and status; the review page shows the report, the reported player (tags, status, counts, links to restrictions, player reports and chat reports), the internal notes with «افزودن یادداشت», «برمی‌دارم برای بررسی» / «برگرداندن به صف», and the decision (dismiss, warn, mute or suspend with the duration chips and translated message used for chat reports, «گزارش‌های باز دیگر علیه این بازیکن هم با همین تصمیم بسته شوند»). The native player page card links to the player reports against the player.
