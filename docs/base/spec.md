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

- `key_type` (asp KeyType) says which module/entity uses it. Registered key types: `product`, `prize`, `gift`, `currency`, `faq`, `news` (categories); tags may use `product`, `prize`, `news`, `faq`, or `general` (usable everywhere). `GET /v2/console/base/{categories|tags}/key-types` returns the list with translated labels.
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

- Settings tab «زبان‌ها» / "Languages" after «ایمیل» (`#/settings/languages`): list with default badge, active toggle, direction, order; form with code, native name, translated names, icon; coverage panel (how many items lack each language, linking to the page).
- New sidebar section «عمومی» / "General" before «اقتصاد», holding «دسته‌بندی‌ها» and «تگ‌ها» (moved out of «اقتصاد»; old hashes redirect). Key type filter and field, icon upload, description.
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
