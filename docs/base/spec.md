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

- `key_type` (asp KeyType) says which module/entity uses it. Registered key types: `product`, `prize`, `gift`, `currency`, `faq`, `news` (categories); tags may use `product`, `prize`, `gift`, `currency` (since `20261110050000-economy-gift-currency-taxonomy.sql`), `news`, `faq`, `general` (usable everywhere) or `user` (player tags, phase 1; migration `20261110020000-base-tag-user-key-type.sql` widens the CHECK). `GET /v2/console/base/{categories|tags}/key-types` returns the list with translated labels (`user` = «بازیکن» / "Player"). Player tags never reach game clients: `base_tags` refuses `keyType: user` and leaves them out of the unfiltered list; economy refuses them on products and prizes (`TAG_KIND_MISMATCH`).
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

Errors are runtime errors with codes `USERNAME_TOO_SHORT`, `USERNAME_TOO_LONG`, `USERNAME_INVALID_CHARACTERS`, `PASSWORD_TOO_SHORT`, `PASSWORD_NEEDS_LETTER`, `PASSWORD_NEEDS_DIGIT`, `DISPLAY_NAME_TOO_LONG` (gRPC InvalidArgument, HTTP 400 with the code as `message`). Nakama-generated usernames (sign-ups without a username, phone accounts) are not checked. Only one before-hook per message exists in Nakama: a game plugin that registers its own before-hook for these messages replaces this one.

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
