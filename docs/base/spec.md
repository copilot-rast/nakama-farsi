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
- Sidebar groups and tabbed pages (owner's design 2026-11-14). One registry in `console/ui/economy/src/pages.js`: `NAV` (per section, an ungrouped page or a group `{ id, items }`, labels `navg.<id>`) and `HUBS` (a page made of tabs: `[page, slug]` pairs, the first tab has no slug). Both the native drawer and the plane sidebar draw from it, and `NakamaConsole.sections()` / `pageTabs()` expose it.
  - اقتصاد: «نمای کلی»; «ارز و پیشرفت»: ارزها، متاها (متاها | قانون‌های متا | جدول سطح‌ها; room for a later «رابطهٔ متاها» tab), تسک‌ها (تسک‌ها | گروه‌های تسک)، دستاوردها، روزهای پیاپی; «فروش»: فروشگاه، آیتم‌ها، هدیه‌ها، پرداخت‌ها; «جایزه و بازی»: جایزه‌ها (جایزه‌ها | گروه‌های جایزه)، بازی‌ها، بازی شانس; «رویداد و رشد»: رویدادها (رویدادها | جایزهٔ رویدادها | گزارش)، کدهای معرف.
  - انتشار: «محتوا»: سؤال‌های متداول، اخبار، کامنت‌ها; «بنر و اسپانسر»: بنرها (بنرها | جایگاه‌های بنر)، اسپانسرها; «پیام به بازیکن»: اعلان‌ها (اعلان‌ها | دسته‌های اعلان)، اعلان‌های خودکار; then «پیکربندی بازی» on its own under a divider. بازیکنان و گفتگو: «بازیکنان»: مدیریت بازیکنان، تعلیق و سکوت، فعالیت بازیکنان; «گزارش و گفتگو»: گزارش‌ها، گزارش بازیکن‌ها، فیلتر کلمات. Support and system are short and stay flat.
  - Groups fold with a caret button (`aria-expanded`); the state is kept per admin and browser in `localStorage` `nakama-console:nav-groups` (only closed groups are stored, a broken value counts as all open). Going to a page inside a closed group opens it again. A collapsed group shows the sum of its badges.
  - A tabbed page has the page name as its heading and a tab strip under the description; each tab is a link with its own address: `#/economy/metas[/rules|/levels]`, `#/economy/tasks[/groups]`, `#/economy/prizes[/groups]`, `#/economy/events[/prizes|/report]`, `#/publication/banners[/placements]`, `#/publication/notifications[/categories]`, then `/<id>` and the query as before, so reload, back and deep links keep the tab. The sidebar entry returns to the tab last opened in this session. Breadcrumbs show page › tab, the top-bar segment shows the pages of the current group.
  - Old addresses redirect (replace) with their id and query: `#/economy/meta-rules` → `metas/rules`, `level-tables` → `metas/levels`, `task-groups` → `tasks/groups`, `prize-groups` → `prizes/groups`, `event-types` → `events`, `event-rules` → `events/prizes`, `event-log` → `events/report`, `#/publication/banner-lists` → `banners/placements`, `notify-categories` → `notifications/categories`.
  - ACL: a tab follows its page's ACL and is hidden without it, a tabbed page with no allowed tab and a group with no visible entry are hidden. Ctrl K lists every page with «بخش · گروه» and every tab with «بخش · صفحه».
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
- Wizard: a centered window over the canvas (`data-economy="setup"`, full screen on phones) with the steps on the side (name, state, «مرحلهٔ n از ۷», a progress bar of finished steps; on phones a row of numbered steps with the current name), one step per screen with the title on top and the actions at the bottom, sized to fit 1280×720 without scrolling (on phones only the step body scrolls); extras sit under a closed «تنظیمات بیشتر» (brand: short name, dark logo, mark, welcome; currency: key, type, icon) that opens itself on an error inside; a new currency starts as `coin`, soft, «سکه»/Coins, 100; payment gateways are an accordion with one gateway's settings open at a time; SMS and email put the test-send box beside the form; the step before review reads «ذخیره و مرور نهایی»; clickable step list, «قبلی», «رد کردن این مرحله», «ذخیره و ادامه», «تمام» on review, «بعداً» (status `skipped`, progress kept; Esc asks first). Focus is trapped in the overlay; leaving a step with unsaved changes asks first. Steps: brand; languages (toggle active, make default, quick add fa/en/ar/tr/ru/de/fr/es/ur/zh, all saved at once); currency (pick existing or new, key/type/name/starting balance/icon; other fields kept on save; Latin digits in the number input); payments («پرداخت لازم دارید؟», pick gateways, sandbox box first, new gateways default to sandbox); SMS and email («… لازم دارید؟», service, sandbox, fields with masked secrets, saved as active + default; test-send saves first and is only allowed in sandbox); review (status per step, summaries, edit links). Answering «نه» marks the step skipped.

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

#### Limits per player type and bought limits (owner's follow-up 2026-11-14)

Migration `20261124000003-base-friend-tiers.sql`. The row above holds the defaults; on top of it:

- `base_friend_rule` (tag rules): `name`, `tag_ids` (player tags, key type `user`, 1–20), `priority` (−1000…1000), and optional `daily_requests`, `max_friends`, `reject_cooldown_hours` (NULL = keep the default, 0 = unlimited / no wait), `is_active`. A rule applies to a player who has any of its tags. Console CRUD `/v2/console/base/friend-rules/tiers` (`SETTINGS`, audit-logged; `TAGS_REQUIRED`, `TAG_KIND_MISMATCH`, `TAG_NOT_FOUND`, `LIMIT_REQUIRED`, `INVALID_RULE`); deleting a tag removes it from the rules.
- `match_policy` (admin choice): `priority` (default: the matching rule with the highest priority overrides the defaults field by field) or `generous` (for each field the most generous value of all matching rules: unlimited beats any number, then the largest number; the shortest wait).
- Bonus metas (admin choice): `requests_bonus_meta` and `friends_bonus_meta` name economy metas (`META_NOT_FOUND` on save). The player's meta value (rounded down, never below 0) is added to the daily requests and to the friend limit, so a shop product, prize, event rule or admin grant that raises the meta raises the limit. Effective limit = tag-rule (or default) base + bonus; an unlimited base stays unlimited. A deleted meta counts as 0.
- `over_limit` (admin choice): `refuse` (default: a request with more new invites than are left is refused whole) or `trim` (the first invites in the order sent go through up to what is left, the rest are dropped from the request).
- The daily counter is now taken with a row lock (`SELECT … FOR UPDATE` on the day row) so both modes stay exact under parallel calls on many nodes. Accepting checks the sender's effective friend limit. Invite expiry stays global (one sweeper).

`base_friend_rules` also returns `{friendSlotsLeft, base: {dailyRequests, maxFriends}, bonus: {dailyRequests, maxFriends}, tier: {id, name} | null, overLimit}`; `dailyRequests`, `maxFriends` and `rejectCooldownHours` are the caller's effective values.

The section «قوانین دوستی» gained «وقتی درخواست‌ها از سقف بیشتر شد», the two meta pickers («سقف بیشتر با متا»), «وقتی چند قانون به یک بازیکن بخورد» and below the form the list «قانون‌ها برای برچسب‌های بازیکن» (name and tag chips, daily requests, most friends with «پیش‌فرض»/«بی‌حد», priority with the player count, status) with the drawer editor `#/project-settings/friends/<id>`.

Friend gifts (economy, phase 22, `docs/economy/spec.md` «Friend gifts») read the same `user_edge` rows: accepted friends both ways, `update_time` of the accepted edge as the friendship start, and a block (state 3) on either side refuses any gift. Their gift limits per player tag follow the same pattern as these friend rules (tag rule with priority, `priority` or `generous`, bonus metas) but live in economy (`economy_friend_gift_rule`).

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

### Friends and profile as RPCs (socket transport)

Nakama has friends, own account and user lookup only as HTTP API calls; the realtime socket has no messages for them. So that a client can send everything over the socket, `integrations/base/social_rpc.go` adds RPCs that do the same work through the runtime (`nk.Friends*`, `nk.Account*`, `nk.Users*`). Runtime calls skip API hooks, so each RPC runs the before/after hooks registered through `internal/hookchain` for the matching API call (`hookchain.RunBefore`, `RunAfter`, `RunAfterIn`): friend limits, invite expiry, rejection cooldown, account rules. Hooks registered only in Lua or JavaScript are not run. Answers use Nakama's own JSON (`protojson`, proto names, enum numbers), the same as the HTTP API. An optional `lang` in any body sets the language of error texts (a socket opened without `lang` is English).

| RPC | Input | Output |
|---|---|---|
| `base_friends_list` | `{state? (0-3), limit? (1-100, default 100), cursor?}` | `{friends: [Friend], cursor}` |
| `base_friends_add` | `{ids?, usernames?}` (1-100 together) | `{friends: [Friend]}`: the targets' edges after the call (no row: nothing changed, e.g. the target blocked the caller) |
| `base_friends_remove` | same | same (also declines, cancels, unfriends and unblocks) |
| `base_friends_block` | same | same |
| `base_account_get` | `{}` | `Account` without `disable_time` |
| `base_account_update` | `{username?, display_name?, avatar_url?, lang_tag?, location?, timezone?}`; empty or missing = unchanged | `Account` after the update |
| `base_users_get` | `{ids?, usernames?}` (≤100) | `{users: [User]}` (no `users` when nothing matched) |

Errors besides the friend and account-rule codes: `NO_USERS`, `INVALID_USER_ID`, `INVALID_USERNAME`, `CANNOT_TARGET_SELF`, `USER_NOT_FOUND` (no given username exists), `INVALID_LIMIT`, `INVALID_STATE`, `INVALID_CURSOR`, `USERNAME_TAKEN` (gRPC 6), `USERNAME_TOO_LONG`. Bodies stay far below the 16 KB RPC limit and the 32 KB socket read limit (100 ids ≈ 4 KB); answers are not limited by the socket read limit.

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

Rules are admin-set (owner's follow-up 2026-11-14, migration `20261124000004-base-report-rules.sql`, row `base_report_rules`): `per_hour` (10; 0 = no limit), `duplicate_hours` (24; 0 = no duplicate check), `min_account_hours` (0; newer accounts get `ACCOUNT_TOO_NEW`, 403), `note_required` (false; `NOTE_REQUIRED`), `trusted_tag_ids` with `trusted_per_hour` (30; 0 = no limit) for players with any of those player tags. Console `GET/PUT /v2/console/base/report-rules` (`CHANNEL_MESSAGE`, audit-logged) and the project settings section «قوانین گزارش بازیکن» (`#/project-settings/report-rules`) after «عنوان‌های گزارش». `base_my_reports` also returns `limits {perHour, remaining (null when unlimited), trusted, noteRequired, minAccountHours, duplicateHours, noteMax}`.

Console: page «گزارش بازیکن‌ها» (`#/moderation/player-reports[/<id>]`, `CHANNEL_MESSAGE`, badge with open and reviewing reports) after «گزارش‌ها» in «بازیکنان و گفتگو»: filters (status with «باز یا در حال بررسی» by default, reason, player id), rows with the reported player, the player's note, context chips (game, match, content), reason with the count against the player, reporter and status; the review page shows the report, the reported player (tags, status, counts, links to restrictions, player reports and chat reports), the internal notes with «افزودن یادداشت», «برمی‌دارم برای بررسی» / «برگرداندن به صف», and the decision (dismiss, warn, mute or suspend with the duration chips and translated message used for chat reports, «گزارش‌های باز دیگر علیه این بازیکن هم با همین تصمیم بسته شوند»). The native player page card links to the player reports against the player.

## Voice chat (phase 18)

Generic voice chat tokens for any game, with the voice service chosen in the console, not in the environment. Code: `internal/voice` (provider registry, LiveKit and comms providers) and `integrations/base/voice.go` (settings, console API, `voice_token`). Migration `20261124000002-base-voice.sql` (Up re-runnable, Down raises) creates `base_voice_settings` (row 1, voice off) and adds `economy_game.voice_enabled` (default true).

### Providers (`internal/voice`)

A provider is one file that calls `voice.Register(voice.Spec{Kind, Fields, New})` in `init`. `Fields` (`name`, `kind` `text|url|wsurl`, `secret`, `required`) drive the console form, `voice.Validate` and which values are sealed; `New(cfg, httpClient)` returns a `Provider` with `Token(ctx, Request) (Grant, error)` and `Test(ctx) error`. Adding a third provider is one new file plus its labels (`vc.provider.<kind>`, `vc.about.<kind>`, `f.vc.<kind>.<field>` and `.hint` in `console/ui/economy/src/i18n.js`); nothing else changes.

| Kind | Fields | Token | Test connection |
|---|---|---|---|
| `livekit` | `url` (ws/wss/http/https), `apiKey`, `apiSecret` (secret) | signed here: HS256 JWT with `iss` = API key, `sub` = Nakama user id, `name` = display name, `nbf` (−10 s), `iat`, `exp` = now + lifetime, `jti`, `video {roomJoin: true, room, canPublish, canSubscribe: true, canPublishData: false}`, `metadata` `{"gameKey"}`; `url` is returned as ws/wss | `POST <http url>/twirp/livekit.RoomService/ListRooms` with a one-minute `roomList` token: 200 ok, 401/403 `VOICE_BAD_CREDENTIALS` |
| `comms` (Rastaar comms) | `url` (server to server), `publicUrl` (for games, defaults to `url`), `serverKey` (secret, the app's `ck_…` key) | `POST /v1/server/game-rooms/{room}/players/{userId}/token` with `Authorization: Bearer <serverKey>` → `{token, expires_at}`; on 404 or `room_closed` it opens the room (`POST /v1/server/game-rooms {match_id, players:[{id,name}], video:false, ttl_seconds}`, retried without `ttl_seconds` when comms refuses it) and asks again; on `not_listed` it adds the player (`PUT …/players/{userId} {name}`) and asks again. The token is comms' one-time join token; the game exchanges it at `POST <publicUrl>/v1/game/join` (the comms SDK does this) | `GET /healthz`, then `GET /v1/server/game-rooms/<random>` with the key: 404 ok, 401/403 `VOICE_BAD_CREDENTIALS` |

Provider errors map to `VOICE_BAD_CREDENTIALS`, `VOICE_ROOM_FULL`, `VOICE_PROVIDER_BUSY`, `VOICE_PROVIDER_UNAVAILABLE`, `VOICE_PROVIDER_TIMEOUT`, `VOICE_PROVIDER_ERROR`. HTTP calls use a 5 s client.

### Settings (`base_voice_settings`, row 1)

`enabled` (false), `provider` (`''`, `livekit`, `comms`), `config` (the provider's non-secret fields), `secrets_sealed` (the secret fields as JSON sealed with `internal/secretbox` box `BaseVoice`, HKDF info `base-voice-secrets` from `ECONOMY_SECRET_KEY`), `ttl_seconds` (600, 60–86400), `open_rooms` (false), `can_publish` (true), `updated_by/at`. Secrets are never returned: the console gets `********` for a saved secret, sends it back unchanged to keep it (only while the provider stays the same) and an empty value to clear it. Without `ECONOMY_SECRET_KEY` a secret cannot be saved (`VOICE_SEAL_UNAVAILABLE`).

Console API (`SETTINGS`, audit-logged with secrets masked; the test as invoke):

| Route | Notes |
|---|---|
| `GET /v2/console/base/voice/settings` | settings with masked `secrets`, `providers` (kind and fields from the registry), `canSeal` |
| `PUT /v2/console/base/voice/settings` | `{enabled, provider, config, secrets, ttlSeconds, openRooms, canPublish}`; unknown fields of a provider are dropped; `VOICE_PROVIDER_REQUIRED` (on without a provider), `INVALID_VOICE_PROVIDER`, `VOICE_FIELD_REQUIRED`, `INVALID_URL`, `VOICE_FIELD_TOO_LONG` (with `field` = the provider field), `INVALID_RULE` (`ttlSeconds`) |
| `POST /v2/console/base/voice/test` | empty body tests the saved settings; a body like PUT tests those values (masked secrets are taken from the saved ones). Answers 200 `{ok, provider, ms, code?, status?}`; 10 tests a minute per console user (`TOO_MANY_TESTS`) |

### Client RPC `voice_token`

Input: exactly one of `matchId`, `partyId` or `room` (`VOICE_ROOM_REQUIRED`), and an optional `gameKey`. Output: `{enabled: true, provider, url, token, room, identity, expiresAt, canPublish, gameKey}`, or `{enabled: false, reason}` with no error when voice is off (`VOICE_DISABLED`), not usable (`VOICE_NOT_CONFIGURED`, e.g. sealed secrets without the key), off for the game (`GAME_VOICE_DISABLED`) or the game is off for an open room (`GAME_DISABLED`). `enabled` and `token` keep the shape the Rakhtkan client already reads (`{enabled, token}` from the legacy `voice.go`).

Who may join:

- `matchId` (`<uuid>.<node>`): a player of a games-registry match (`economy_game_match` open or started, with an `economy_game_entry` that is pending, paid or free; read from the database, so it works on any node), else a presence in the match stream (`MatchGet` then the authoritative or relayed match stream). The room is the match id; the game key comes from the registry match or the authoritative handler name, never from the client. `NOT_IN_MATCH` (403), `MATCH_NOT_FOUND` (404).
- `partyId`: a presence in the party stream; room `party.<partyId>`; `NOT_IN_PARTY`.
- `room` + `gameKey`: only when «اتاق‌های باز هر بازی» is on (`ROOM_NOT_ALLOWED`), room key `^[a-z0-9_-]{1,48}$` (`INVALID_ROOM`), the game must exist (`GAME_NOT_FOUND`, `GAME_KEY_REQUIRED`); room `room.<gameKey>.<room>`.

Admin policies (owner's follow-up 2026-11-14, migration `20261124000005-base-voice-policies.sql`): `allowed_tag_ids` (empty = everyone; others get `{enabled: false, reason: "VOICE_NOT_FOR_PLAYER"}`), `listen_only_tag_ids` (those players get `canPublish: false`), `tokens_per_minute` (30, 1–600) and `max_room_players` (0 = no limit) with a per-game override `economy_game.voice_max_players` (NULL = the global number): in a match or party only the first players up to the limit (registry entries in join order, otherwise stream order) get a token, others `VOICE_ROOM_FULL` (409), and comms gets only those players. Open rooms are not counted (the providers know who is inside, Nakama does not).

Then the game's `voice_enabled` is checked, a muted player (`base_restriction`) gets `canPublish: false` (listen only), and `tokens_per_minute` tokens a minute per player are allowed (`clusterkit.Allow`, `VOICE_RATE_LIMITED`, 429). The identity is the Nakama user id and the name the display name (or username). Stream presence checks read this node's tracker; in cluster mode use registry matches (database) for reliable checks across nodes.

### Console

Project settings section «سرور ویس» (`#/project-settings/voice`, `SETTINGS`), before «ریپوی بالایی»: state line, «چت صوتی روشن باشد», provider select (fields of the chosen provider with a short note, secrets as password fields shown as `********` once saved), token lifetime, «بازیکن بتواند حرف بزند», «اتاق‌های باز هر بازی», «امتحان اتصال» (result with the time in ms or the translated error) and save. The section also has «توکن در دقیقه برای هر بازیکن», «بیشترین بازیکن در اتاق ویس», «فقط این برچسب‌ها ویس بگیرند» and «این برچسب‌ها فقط بشنوند». Games editor: «چت صوتی» switch with «بیشترین بازیکن در ویس»; games list chip «بی‌ویس».

## Private chat quota (phase 23)

A daily quota on private (direct) messages, with message packs sold in the shop and unlimited players. Off by default: until an admin turns it on nothing is counted or stored and direct messages behave as before. Migration `20261128000000-base-chat-quota.sql` (Up re-runnable, Down raises). Code: `integrations/base/chatquota.go` (engine, hook, RPC, sweep) and `chatquota_admin.go` (console).

### Where it hooks

Nakama sends chat only over the realtime socket (`ChannelMessageSend` in `server/pipeline_channel.go`); there is no REST send. The base plugin's existing before hook `rt:ChannelMessageSend` (composed with `internal/hookchain`) first runs moderation (suspension, mute, word filter) and then `chatQuotaGate`, so a refused or masked message never costs quota. The gate:

1. loads the settings (`base_chat_quota_settings`, row 1); off → pass;
2. parses the channel id: `4.<a>.<b>.` is a direct message (the peer is the other id), `3.<group>..` a group, `2...<label>` a room; groups and rooms are counted only when the admin turns them on;
3. leaves anything Nakama will refuse anyway to Nakama: a bad channel id, content that is not a JSON object, or a session that has not joined the channel (`nk.StreamUserGet` on the caller's session) — those never cost quota;
4. decides and records the message in one transaction that locks the player's day row (`SELECT … FOR UPDATE`), so parallel sends on any node never pass the quota.

Messages sent by server code (`nk.ChannelMessageSend` from a runtime module) are not counted. `ChannelMessageUpdate` (editing) is not counted.

### Data

| Table | What |
|---|---|
| `base_chat_quota_settings` (row 1) | `settings` JSONB (below), `updated_by/at` |
| `base_chat_quota_rule` | tag rules: `name`, `tag_ids` (player tags, 1–20), `priority` (−1000…1000), `unlimited`, optional `daily_messages`, `daily_conversations`, `friend_daily_messages`, `friend_daily_conversations` (NULL keeps the settings value, 0 = no free quota), `is_active` |
| `base_chat_quota_day(user_id, day)` | per player and quota day: `messages`, `conversations`, `friend_messages`, `friend_conversations` (free units used), `pack_used`, `free_sent` (free replies or friend messages), `unlimited_sent`, `total`, `blocked`, `notified`; kept 400 days |
| `base_chat_quota_target(user_id, target)` | `first_at`, `last_at` of the player's messages to a target (`u:<peer>`, `g:<group>`, `r:<label>`); used for new chats and free replies |
| `base_chat_quota_lot` | message packs: `amount`, `remaining`, `source` (`meta` = bought or won, `admin` = console), `note`, `expires_at` (NULL = never), `created_by/at`; spent oldest-expiry first; used or expired lots are deleted after 180 days |

### Settings (all admin choices, defaults in brackets)

| Field | Values |
|---|---|
| `enabled` | [false] |
| `countMode` | [`messages`] each message costs one unit; `conversations` only the first message of a new chat costs a unit; `both` messages cost units and new chats also count against a hard daily cap (packs add messages only) |
| `dailyMessages`, `dailyConversations` | free units a day [20, 5], 0…100000; 0 = no free quota, packs only |
| `conversationScope` | [`day`] a new chat is the first message today to that target; `ever` the first message ever |
| `resetTime`, `timeZone` | [`00:00`, `Asia/Tehran`] start of the quota day (`INVALID_SETTING`, `INVALID_TIMEZONE`) |
| `countGroups`, `countRooms` | [false, false] also count group or room chat (same quota) |
| `replyFree` | [`off`]; `started` messages in a chat the other player started are free; `reply` one free reply after each of their messages |
| `friendPolicy` | [`same`]; `separate` accepted friends use `friendDailyMessages` / `friendDailyConversations` [50, 10]; `free` messages to friends are free |
| `packMeta` | economy meta key [none]; whole units of the player's value move into a pack lot (the meta keeps only the fraction), on the next send, status call or the 1-minute sweep; the lot's start is the meta's `updated_at` (the purchase) |
| `packExpiryDays` | [0 = never] 0…3650, from the purchase |
| `order` | [`free_first`] or `packs_first` |
| `packProducts` | product keys sent to the client when the quota runs out [empty = every active product whose prize adds to `packMeta`] |
| `bonusMeta` | meta whose value is added to the free daily quota every day and not spent [none] |
| `matchPolicy` | [`priority`] the highest-priority matching tag rule wins field by field; `generous` takes the largest value of each field and unlimited if any rule is unlimited |
| `unlimitedItems` | item keys; owning any → unlimited [none] |
| `unlimitedMeta`, `unlimitedMetaMode` | meta that makes the player unlimited: [`positive`] value > 0, or `until` value is a Unix time in the future (a subscription end) |
| `message` | translated text that replaces the built-in refusal text (`{reset}`, `{limit}`) [built-in fa/en/ar] |
| `notice` | [`none`]; `inbox` also sends a persistent notification (code 1120) on the first refusal of the day |

The pack meta cannot also be the bonus or unlimited meta (`SAME_META`). Unknown metas, items or products are refused (`META_NOT_FOUND`, `ITEM_NOT_FOUND`, `PRODUCT_NOT_FOUND`).

Selling packs needs no new purchase code: the admin makes a meta (for example `dm_pack`), a prize that adds to it (`+50`) and a product with that prize; coins, gateways and stores work as for any product, and gift codes, event prizes or meta relations that add to the meta give messages the same way.

### Refusal

A refused send is a runtime error on the socket whose message is JSON (gRPC 8, `RESOURCE_EXHAUSTED`):

`{"code": "DM_QUOTA_EXCEEDED" | "DM_CONVERSATION_LIMIT", "message" (player's saved language, fa/en/ar built in or the admin text, Persian digits in fa), "lang", "unit": "messages" | "conversations", "limit", "used", "remaining": 0, "packsLeft", "resetsAt" (RFC 3339), "friend", "rule": {id, name} | null, "products": [{id, key, name, amount}], "productIds": [...]}`

`DM_CONVERSATION_LIMIT` is only used in `both` mode (the new chat cap is reached; packs do not help). The refusal increments `blocked` for the report.

### Client RPC `base_chat_quota {lang?}`

`{enabled, countMode, unlimited, unlimitedBy ("rule" | "item" | "meta" | ""), rule, freeLeft (free units left today in the main pool), packsLeft, packsExpireAt, order, day, resetsAt, timeZone, messages: {limit, used, left}, conversations: {…}, friends?: {messages, conversations} (when separate), friendPolicy, replyFree, bonus, today: {sent, fromPacks, free, unlimited, blocked}, products, productIds, counts: {direct, groups, rooms}}`.

### Console API (`/v2/console/base/`, `CHANNEL_MESSAGE` read/write, audit-logged)

| Route | What |
|---|---|
| `GET/PUT /chat-quota/settings` | settings (full replace); GET adds `detectedProducts` |
| `GET/POST /chat-quota/rules`, `GET/PUT/DELETE /chat-quota/rules/{id}` | tag rules (`LIMIT_REQUIRED` unless unlimited or a limit is set, `TAGS_REQUIRED`, `TAG_KIND_MISMATCH`); deleting a tag removes it from the rules |
| `GET /players/{id}/chat-quota` | the RPC answer plus `lots` (20) and `days` (14) |
| `POST /players/{id}/chat-quota/grant {amount, expiresInDays?, note?}` | an admin pack; `expiresInDays` missing → settings expiry, 0 → never |
| `POST /players/{id}/chat-quota/reset` | today's free counters back to 0 (packs untouched) |
| `POST /players/{id}/chat-quota/lots/{lot}/revoke` | a pack's remaining messages to 0 |
| `GET /chat-quota/report?days=` | totals (sent, blocked, blocked players, from packs, free, unlimited), daily series, top refused players, pack sales from `economy_payment` (count, Toman, coins, per product) and packs given by purchase/prize vs console |

### Console pages

«سهمیهٔ پیام خصوصی» in «بازیکنان و گفتگو» › «گزارش و گفتگو» with tabs «قانون‌ها» (`#/moderation/chat-quota`, tag rules), «تنظیمات» (with a summary line and the products that sell packs) and «آمار». The player page gets a «سهمیهٔ پیام خصوصی» card: free left, packs left and next expiry, today's sends, unlimited and rule chips, «پیام اضافه» (grant dialog), «صفر کردن امروز», packs with «پس گرفتن», and the last days.
