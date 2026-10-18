# Persian (fa-IR) Localization Audit

This audit records what the repository contains, which parts are user-facing,
and how each was localized. It was produced by inspecting the repository, the
compiled console bundle, and the running console (Nakama built from this
repository, plus the official 3.41.0 image, both against PostgreSQL 16).

## 1. Repository findings

| Aspect | Finding |
| --- | --- |
| Languages | Go server (`main.go`, `server/`, `internal/`, `social/`, `iap/`), Lua/JavaScript/Go runtime APIs, protobuf/gRPC (`apigrpc/`, `console/console.proto`) |
| Admin UI | The Nakama Console, shipped only as a **prebuilt, minified Vue 3 bundle** in `console/ui/dist` (`static/index-*.js`, 7 MB, carrying a Heroic Labs proprietary notice). The UI source is not in this repository: the Angular console was removed in `61291d76`, and the current Vue console is developed upstream elsewhere. |
| Embedding | `console/ui.go` embeds `ui/dist/*` with `go:embed`. `server/console.go` serves `index.html` (replacing `{{nt}}`) and, for `/static/*.js` and `*.css`, the **`.gz` sibling** when the client accepts gzip. |
| Component library | Internal `x-*`/`hl-*` Vue components, Tailwind CSS v3 utilities, Monaco editor, ECharts (SVG renderer), vue-datepicker, number-flow, tiptap, Shiki, Segment analytics |
| Existing i18n | vue-i18n is bundled but only scaffolded (`en.json`/`pt.json` with a single `hello` key). Every UI string is hard-coded in compiled templates. No RTL support (no `[dir=rtl]` rule in 205 KB of CSS). |
| Dates and numbers | Mostly `Intl.DateTimeFormat`/`NumberFormat` with the browser default locale; relative times via `Intl.RelativeTimeFormat("en")` (the app strips `" ago"` for durations); date-fns/dayjs in a few places. vue-datepicker formats with an explicit `en-US` and parses that output. |
| Backend messages | ~220 `status.Error` messages in `server/console*.go` reach the console as toast text; config warnings; notification subjects (`"%v wants to add you as a friend"`). They are API contracts for game clients and SDKs. |
| Emails | The server sends no email. The console composes one: the password-reset `mailto:` link (subject and body built in JavaScript). |
| Public pages | None besides the console (`/login`, `/reset`, `/reset-mfa`). The sign-in page shows an external Heroic Labs blog feed. |
| Tests / CI | Go tests (`go test`), GitHub Actions `build.yaml` (binary build). No frontend tests (no frontend source). |

## 2. Approach decided by the findings

Because the console cannot be rebuilt from source, localization is an
**overlay** that leaves the vendor bundle byte-for-byte unchanged:

1. A runtime script loaded before the bundle translates the rendered DOM
   from a key-based catalog, sets `lang`/`dir`, and redirects default-locale
   `Intl` formatting to `fa-IR` (Solar Hijri dates, Persian digits).
2. A build step converts the console stylesheet to CSS logical properties (a
   faithful no-op in LTR, mirrored in RTL) and adds Persian typography.
3. Server messages stay unchanged in Go; their Persian text is in the catalog's
   `server` namespace and applied where the console displays them.

Details are in [persian-localization.md](persian-localization.md).

## 3. Localization inventory

Priorities: P1 = every session, P2 = common admin tasks, P3 = specialized or optional modules.

| Area | Routes / sources | Before | RTL | Status | Risks / notes | Priority |
| --- | --- | --- | --- | --- | --- | --- |
| Sign in, MFA, password and MFA reset | `/login`, `/reset`, `/reset-mfa`; MFA setup | English only | Mirrored; password and username inputs keep icons on the correct side | Translated, including error alerts split around links | Blog feed on the sign-in page is external English content (rendered as LTR paragraphs) | P1 |
| Shell: sidebar, top bar, user menu, breadcrumbs, titles | layout components, `document.title` | English | Sidebar on the right, drawer slides from the right, icons mirrored | Translated; language switcher in the top bar | | P1 |
| Dashboard | `/` | English, en-US numbers and times | Cards and tables mirrored; charts keep left-to-right time axes | Translated; Persian digits; Jalali/24h axis labels; uptime as a duration | Chart titles are SVG text and get explicit RTL isolation | P1 |
| Shared components | tables, pagination, filters, selects, uploads, empty states, confirm dialogs, side panels, toasts | English | Mirrored; pagination icons mirrored | Translated, including ARIA labels | | P1 |
| Players | `/players`, profile, account, wallet, friends, groups, payments, storage, notifications, notes, import/export, ban/unban, unlink | English | Mirrored | Translated, including enum labels (friend states) | User data (names, IDs, JSON) is never rewritten | P1 |
| Groups | `/groups`, details, members | English | Mirrored | Translated, including role labels | | P2 |
| Storage and storage indexes | `/storage`, `/storage-indexes` | English | Mirrored; JSON editors LTR | Translated, including permission labels | | P2 |
| Leaderboards and tournaments | `/leaderboards`, records, details | English | Mirrored; cron expressions kept LTR | Translated, including sort order and operator | | P2 |
| Matches, chat, notifications, payments | `/matches`, `/chat`, `/notifications`, `/payments` | English | Mirrored | Translated | Notification subjects created by the server are translated on display | P2 |
| API Explorer, runtime modules | `/api-explorer`, `/runtime` | English | Mirrored; Monaco editor LTR | Translated | Monaco's own menus stay English (third-party) | P2 |
| Settings | users, permissions templates, general, configuration, data management, audit log | English | Mirrored; config value previews LTR | Translated, including ACL resource names and audit actions | Confirmation words users must type (`delete`, `update`) stay Latin | P2 |
| Date and time pickers | vue-datepicker in filters | English, Gregorian | Mirrored grid and arrows; `hh:mm` LTR | Labels, weekdays, months, ARIA translated | Calendar grid stays Gregorian (third-party; no Jalali mode) | P2 |
| Validation and toasts | form rules, success/error toasts | English | n/a | Translated | | P1 |
| Backend (console API) messages | `server/console*.go` | English | n/a | Translated on display (`server` namespace, 223 messages) | Go strings intentionally unchanged (API contract) | P2 |
| Password-reset email | `mailto:` composed by the console | English | n/a | Subject and body translated when opened | Link kept free of bidi control characters | P3 |
| Hiro configuration editors | economy, energy, inventory, achievements, rewards, scheduling | English | Mirrored | Translated (~400 messages) | Requires a Hiro licence; verified by catalog/audit, not visually | P3 |
| Satori messages | player Satori tab | English | Mirrored | Translated | Requires Satori; not visually verified | P3 |
| Cron description sentences | cronstrue-style generator in Hiro scheduling | English | n/a | Not translated | Sentences are assembled from many fragments by the library | P3 |
| Monaco editor, analytics SDK, rich-text editor internals | vendor code | English | Editors forced LTR | Not translated (third-party) | | n/a |

## 4. Translation completeness (evidence)

`npm run audit` (in `console/ui/l10n`) recovers user-facing strings from the
compiled bundle by syntactic context (Vue text vnodes, element text, UI props,
toast calls), runs each through the runtime matcher, and classifies the rest
with `audit-ignore.json`. Result for `index-CNoeGISM.js`:

| | Count |
| --- | --- |
| User-facing candidates | 1204 |
| Translated | 1039 |
| Intentionally untranslated: technical (CSS classes, identifiers, expression syntax) | 49 |
| Intentionally untranslated: third-party (Monaco editor) | 50 |
| Intentionally untranslated: developer-only (console logs) | 43 |
| Covered by fragments (HTML/concatenated text translated piecewise) | 15 |
| Brand and protocol names (Google Authenticator, Authy, OneSignal, UTC, HTML, JSON, UID, `DIRECT_MESSAGE_SENT`) | 8 |
| **Untranslated user-facing strings** | **0** |

The catalog holds 1636 messages; besides the audited contexts it covers strings
the audit cannot see syntactically: enum labels the console title-cases at run
time, relative times, server messages, the reset email, and date-picker labels.
Crawling every route in a browser (desktop 1440×900, tablet 820×1180, phone
390×844), including modals, menus, validation errors and date pickers, found
no untranslated console text; the remaining English was data (IDs, usernames,
config keys), the external blog feed, and brand names.

## 5. Known limitations

1. **No console source.** Localization works on the rendered DOM. Strings are
   matched by their English text, so a console upgrade that rewords a string
   needs a catalog update (`npm run audit` lists such strings).
2. **User data equal to a UI string.** A value that exactly equals a catalog
   entry (for example a group literally named `Open`) is shown translated.
   Inputs, editors, code and multi-word user text are never rewritten.
3. **Date picker is Gregorian.** vue-datepicker has no Solar Hijri mode; its
   labels are Persian, but the grid follows the Gregorian calendar. All dates
   displayed elsewhere use the Solar Hijri calendar.
4. **Monaco editor UI** (context menu, command palette) stays English.
5. **External blog feed** on the sign-in page is English content from Heroic Labs.
6. **Hiro and Satori screens** need licensed modules and were verified by the
   catalog and audit, not by screenshots.
7. **Cron descriptions** generated by the scheduling library stay English.
8. **Search normalization.** Player and group search runs in SQL on the
   server; Arabic/Persian letter variants (ي/ی, ك/ک) are not normalized.

## 6. Upstream merge notes

Files added by the fork: `console/ui/l10n/**`, `console/ui_l10n_test.go`,
`docs/persian-localization*.md`, `.github/workflows/console-l10n.yaml`, and
generated files in `console/ui/dist/static` (listed in
`console/ui/l10n/dist-manifest.json`).

Upstream-owned files touched: `console/ui/dist/index.html` and `index.html.gz`
(a marked block and the stylesheet link). When an upstream release replaces
`console/ui/dist`, run `npm ci && npm run build && npm run audit` in
`console/ui/l10n` and commit the result. No Go source, API, schema or vendor
bundle was modified.
