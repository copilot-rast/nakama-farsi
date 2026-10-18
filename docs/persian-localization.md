# Persian (fa-IR) Localization of the Nakama Console

The Nakama Console in this fork opens in Persian, right-to-left, with Solar
Hijri dates and Persian digits. English is one click away and is unchanged
from upstream.

The console is shipped upstream as a prebuilt bundle (`console/ui/dist`), so
localization is an overlay: the vendor files are untouched, and a small
runtime plus generated stylesheets are loaded alongside them. The audit that
led to this design is in [persian-localization-audit.md](persian-localization-audit.md).

## Layout

```
console/ui/l10n/
  config.json            locales, default locale, directional icons, fonts
  locales/en.json        English source messages (what the console renders)
  locales/fa-IR.json     Persian translations, same namespaces and keys
  locales/catalog.schema.json
  src/runtime.js         localization runtime (DOM, Intl, title, mailto)
  src/l10n.css           typography and RTL refinements
  fonts/                 Vazirmatn (SIL OFL 1.1) and its licence
  scripts/build.mjs      builds into console/ui/dist and injects index.html
  scripts/logical-css.mjs  physical -> logical CSS conversion
  scripts/audit.mjs      translation completeness audit
  scripts/catalog.mjs    catalog loading and validation
  audit-ignore.json      strings intentionally left untranslated, by category
  dist-manifest.json     files the build generated in console/ui/dist
  test/                  Node test suite (node:test + jsdom)
console/ui_l10n_test.go  Go test for the embedded, served assets
```

Generated files are committed, so `go build` needs no Node toolchain.

## How it works

`scripts/build.mjs` writes content-hashed files into `console/ui/dist/static`
(with the `.gz` copies the console server serves for `/static/*.js|css`) and
adds a marked block to `index.html`:

```html
<html lang="fa-IR" dir="rtl">
  ...
  <!-- nakama-l10n:start -->
  <link rel="preload" href="/static/vazirmatn-nl-wght-<hash>.woff2" as="font" type="font/woff2" crossorigin>
  <link rel="stylesheet" href="/static/l10n-<hash>.css">
  <script src="/static/l10n-<hash>.js"></script>
  <!-- nakama-l10n:end -->
  <script type="module" crossorigin src="/static/index-<upstream>.js"></script>
  <link rel="stylesheet" crossorigin href="/static/index-<upstream>.logical-<hash>.css" data-l10n-source="/static/index-<upstream>.css">
```

The build is idempotent: it first removes its previous block, so it can run
again after an upstream update.

The runtime (a classic script, so it runs before the module bundle):

1. **Chooses the locale**: `?lang=en|fa` (remembered), then the stored
   preference (`localStorage["nakama.console.locale"]`), then `fa-IR`. It sets
   `<html lang dir>` and keeps them when the console tries to reset `lang`.
2. **Formats with the locale**: `Intl.DateTimeFormat`, `Intl.NumberFormat`,
   `Intl.RelativeTimeFormat` and `Date#toLocale*String` calls *without an
   explicit locale* use `fa-IR`, whose default calendar is Solar Hijri and
   whose digits are Persian. Calls with an explicit locale are left alone,
   because the date picker and number inputs parse that output. Exceptions,
   both display-only: currency formatting becomes Persian (the currency is
   kept, nothing is converted), and `formatToParts` keeps Latin digits for the
   animated counters, which parse digits; their Persian glyphs come from the
   Vazirmatn FD font instead.
3. **Translates the DOM**: a `MutationObserver` translates text nodes and the
   `placeholder`, `title`, `aria-label`, `alt` attributes as Vue renders and
   updates them, and the document title per breadcrumb segment. It skips
   `pre`, `code`, `textarea`, editors (Monaco, ProseMirror), `contenteditable`,
   and anything marked `translate="no"`, `.notranslate` or `data-l10n-skip`.
4. **Keeps mixed text readable**: interpolated Latin values are isolated
   (FSI…PDI); digit groups with separators in Persian sentences (UUIDs,
   ranges) are kept in order (LRI…PDI); untranslated English sentences
   (external feed, server text) become LTR paragraphs; untranslated
   multi-token values such as cron expressions keep their order. Single tokens
   like IDs are left untouched so copying them stays exact.
5. **Mirrors directional icons**: SVGs whose markup matches an icon listed in
   `config.json` (`mirrorIcons`, extracted from the bundle's icon registry at
   build time) get `data-l10n-icon="mirror"`; logos, media, clocks and charts
   are not mirrored.
6. **Localizes the password-reset email**: `window.open("mailto:…")` subjects
   and bodies are translated before the mail client opens.

Debug helpers in the browser console:

```js
NakamaL10n.locale              // "fa-IR"
NakamaL10n.setLocale('en')     // switch and reload
localStorage.setItem('nakama.console.l10n.debug', '1')  // then reload
NakamaL10n.missing()           // English strings seen on screen without a translation
```

## Translation catalogs

Both catalogs are namespaced JSON with identical keys:

`locales/en.json`:

```json
"players": {
  "ban": "Ban",
  "areYouSureYouWantTo2": "Are you sure you want to ban the player '{name}'?"
},
"time": {
  "countMinutes": "1 minute | {count} minutes"
}
```

`locales/fa-IR.json`:

```json
"players": {
  "ban": "مسدود کردن",
  "areYouSureYouWantTo2": "آیا مطمئن هستید که می‌خواهید بازیکن «{name}» را مسدود کنید؟"
},
"time": {
  "countMinutes": "{count} دقیقه"
}
```

Keys are derived from the English text and only need to be unique within a
namespace; they never appear in the UI.

Namespaces: `common`, `nav`, `auth`, `dashboard`, `components`, `time`,
`validation`, `players`, `groups`, `storage`, `leaderboards`, `matches`,
`chat`, `notifications`, `payments`, `runtime`, `apiExplorer`, `settings`,
`hiro`, `satori`, `server` (console API errors returned by the Go server).

Matching rules:

- The **English value is what the console renders**; matching ignores
  surrounding and repeated whitespace (which is preserved in the output).
- **Placeholders** `{name}` capture runtime values. The name sets what may be
  captured: `count, n, from, to, min, max, amount, seconds, percent, row,
  step, factor` match numbers only; `field, id, provider, version, word,
  path, status` match one token; `time` must itself be a catalog phrase or
  already localized; any other name matches free text. Numeric values are
  shown with Persian digits. Prefer the narrowest type, so short patterns
  cannot swallow unrelated sentences.
- **Plurals**: `"1 minute | {count} minutes"`. A form without the placeholder
  treats its `1` as the count. The translation may have one form or
  `one | other` (chosen with `Intl.PluralRules("fa-IR")`); Persian usually
  needs one.
- **Context**: `{"message": "read", "context": ".x-toggle"}` applies only
  inside elements matching the selector. Context entries win over plain ones.
- **Fragments**: text split by links or `<strong>` arrives as separate text
  nodes. Translate each fragment so the Persian sentence reads in order around
  the inner element; a fragment that starts with punctuation attaches to the
  previous element (the source's leading space is dropped).
- Exact matches win over patterns; longer patterns win over shorter ones.

### Adding or changing a translation

1. Find the English text as rendered (the debug list above, or `npm run audit -- --list`).
2. Add the key to the right namespace in **both** `locales/en.json` and
   `locales/fa-IR.json` (keys: lowerCamelCase).
3. Run `npm run build && npm test` in `console/ui/l10n`, then commit the
   catalogs together with the regenerated `console/ui/dist` files.

The build fails on missing translations, unknown keys, placeholder mismatches,
Arabic ي/ك instead of Persian ی/ک, misplaced zero-width non-joiners, and
duplicate source messages.

### Persian style used

- Persian ی and ک; half-space (ZWNJ) in compounds: می‌شود، ثبت‌نام،
  به‌روزرسانی، پیش‌فرض، جدول‌های امتیازات.
- Guillemets « » for quoted values; Persian punctuation ، ؛ ؟.
- Words the user must type stay Latin (`delete`, `update`).
- Product and protocol names stay in Latin script: Nakama, Hiro, Satori,
  Heroic Labs, API, RPC, JSON, MFA, UTC.
- Glossary: Player بازیکن · Group گروه · Leaderboard جدول امتیازات ·
  Tournament تورنمنت · Match مسابقه · Authoritative سرورمحور · Session نشست ·
  Presence حضور · Party پارتی · Storage ذخیره‌سازی · Storage object شیء
  ذخیره‌سازی · Collection کالکشن · Wallet کیف پول · Notification اعلان ·
  Runtime زمان اجرا · Endpoint نقطهٔ پایانی · Currency ارز · Inventory موجودی ·
  Reward جایزه · Achievement دستاورد · Item set مجموعهٔ آیتم.

## RTL conventions

- **Layout**: `scripts/logical-css.mjs` rewrites the console stylesheet's
  physical properties to logical ones (`margin-left` → `margin-inline-start`,
  `left` → `inset-inline-start`, `text-align: right` → `end`, asymmetric
  4-value `margin`/`padding`/`inset`/`border-*`/`border-radius`). This is
  identical in LTR and mirrors in RTL without cascade conflicts. Horizontal
  translations (`--tw-translate-x`, `translateX`) get negated
  `html[dir="rtl"]` overrides. Monaco editor rules and `@keyframes` are left
  alone.
- **Always LTR**: `pre`, `code`, `kbd`, `samp`, Monaco and code previews
  (`pre.shiki`), the `hh:mm` time input, and chart containers (time flows left
  to right; Persian labels inside charts are embedded as RTL runs).
- **Inputs**: the console's inputs default to `dir="ltr"`. In RTL they take
  `direction: rtl` (so icons and paddings match the mirrored layout) and
  `unicode-bidi: plaintext` (text order follows the content: Persian
  right-to-left, emails/IDs/URLs left-to-right). Monaco's internal textarea is
  excluded.
- **New CSS** in `src/l10n.css` should use logical properties and scope RTL
  fixes with `html[dir='rtl']`.
- **Icons**: add an icon name to `config.json` → `mirrorIcons` only if it
  points in the reading direction (arrows, chevrons, back, logout, send).

## Fonts

Vazirmatn v33.003 by Saber Rastikerdar, SIL Open Font License 1.1
(`fonts/OFL.txt`, published at `/static/vazirmatn-OFL.txt`), from the official
release at https://github.com/rastikerdar/vazirmatn.

- `Vazirmatn-NL-wght.woff2` (variable, 100–900, Non-Latin cut, 48 KB, preloaded)
  covers Persian/Arabic only (`unicode-range`), so Latin text keeps the
  console's Inter and monospace stays Roboto Mono.
- `Vazirmatn-FD-NL-*` (Farsi-Digits cut, 400/500/700) is used only for the
  animated dashboard counters, whose DOM keeps Latin digits.
- Stack: `'Vazirmatn', 'Inter Variable', Tahoma, Arial, sans-serif` via
  `--font-family-sans`; `font-display: swap`. Letter-spacing is reset for
  Persian; 10 px sidebar labels are raised to 11 px.

## Dates, times and numbers

- Displayed dates use `fa-IR` → Solar Hijri, e.g. `۱ فروردین ۱۴۰۳، ۱۶:۰۰`.
  Time zones are unchanged: the browser zone, or UTC when the console's
  "Show dates in UTC timezone" setting is on. Nothing assumes `Asia/Tehran`.
- Timestamps sent to and stored by the API are untouched (ISO 8601/Unix).
- Relative times and durations: `۵ دقیقه پیش`, `۴۷ دقیقه`.
- Numbers: `۱٬۲۳۴٬۵۶۷٫۵`; pure-number text in the UI is shown with Persian
  digits, except inside `.font-mono`, `[dir="ltr"]`, `[data-l10n-latn]`, code
  and inputs. IDs, versions (`3.41.0`) and long digit strings keep Latin digits.
- Currency: amounts keep their currency (USD stays USD); only the presentation
  is Persian. Nothing is converted to rials or tomans.
- The date picker (vue-datepicker) is localized but its grid is Gregorian;
  see limitations.

## Backend messages

Go error messages, config warnings and notification subjects are unchanged:
they are API contracts that game clients and SDKs may rely on. Their Persian
text lives in the `server` and `notifications` namespaces and is applied when
the console displays them. Machine-readable codes are untouched.

## Testing

```sh
cd console/ui/l10n
npm ci
npm test            # catalogs, matcher, DOM runtime in jsdom, Intl, CSS, build --check, audit --strict
npm run audit -- --list
go test -mod=vendor ./console/   # from the repository root: embedded index and assets
```

The Node suite covers: Persian as default and RTL root, catalog completeness
and key parity, interpolation, typed placeholders, pluralization, context
entries, Persian digits, Solar Hijri formatting (Nowruz and the 1403 leap
day), explicit-locale passthrough, relative-time compatibility, LTR handling
of code and identifiers, icon mirroring, title translation, locale switching
and persistence, fallback for unknown locales, English left untouched, the
reset email, idempotent `index.html` injection, and that the committed `dist`
matches the sources. CI runs it in `.github/workflows/console-l10n.yaml`.

## Updating to a new upstream console

When an upstream release replaces `console/ui/dist`:

```sh
cd console/ui/l10n
npm ci
npm run build                 # regenerates assets, re-injects index.html, removes stale files
npm run audit -- --list       # new or reworded strings
# add missing strings to locales/en.json + locales/fa-IR.json, rebuild
npm test
```

Then check the main screens in both languages. If a directional icon was
added upstream, add its name to `mirrorIcons`.

## Known limitations

See section 5 of the [audit](persian-localization-audit.md#5-known-limitations).
