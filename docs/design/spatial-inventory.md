# Spatial redesign: screen inventory and technique proof

Every screen of the console as it is on 2026-11-10 (live `http://100.106.1.79:7451`, branch `dev`), what it is built from, what changes in the spatial design (`docs/design/spatial.md`) and what can go wrong. Screenshots are 1440×900, fa, light, JPEG, in `.local/spatial-inventory/` (git-excluded; regenerate with the crawl described at the end).

## 1. How the console is put together

```
#app
└ div.h-screen.flex-col
  └ div.flex-1.flex.overflow-hidden
    ├ div.flex.border-r.z-20.dark            ← sidebar wrapper, always dark
    │ └ .x-drawer (220px)                    ← logo row · .x-scroll › .grid of groups (nav + label + <nav>) · footer grid (Cloud, Satori, Hiro links)
    └ #main.flex-col.overflow-y-auto.bg-secondary-50   ← the scroll container
      ├ div.sticky.top-0.z-30.bg-white.min-h-[51px]   ← top bar: breadcrumb (detail pages) · server stats · gear (new) · EN · theme · user
      ├ main.flex-col.pt-4                     ← native page (hidden while an overlay page shows)
      │ └ div.max-w-[1600px].mx-auto › header (div.mb-4.px-6: h1 + description + actions) › body (div.px-6 › .grid.gap-6)
      └ #economy-root.flex-col.pt-4            ← our pages (base, economy, publication, settings tabs)
body
├ .x-notifications (toasts)
├ teleported: scrim div.fixed.inset-0.bg-secondary-900.opacity-60 + holder div.fixed.inset-0.justify-end › .x-form side panel
├ teleported: x-modal holders, v-popper poppers
└ .context-view, .monaco-aria-container (Monaco)
```

Native class recipes the skin keys on (all found on live pages):

| Component | Recipe |
|---|---|
| Card / window | `div.flex.flex-col.rounded-md.shadow.bg-white` (`dark:bg-secondary-800`) |
| Table card | `div.bg-white.shadow.border-t.rounded-md.overflow-hidden` › `div.w-full.relative.overflow-x-scroll` › `table` |
| Table head | `thead.x-table-head.bg-secondary-100.sticky.top-0.z-10` › `th.x-table-header.p-2.5.font-medium` |
| Table cell | `td.x-table-cell.px-3.h-9.py-0.5.align-middle` (`.text-center` variant); rows `tr.x-table-row` (some lists have no class on `tr`) |
| Button | `button.x-button._button_omddn_2.inline-flex.font-medium` with colour from inline `--x-button-*` variables |
| Input | `input.appearance-none.block.w-full.border-secondary-300.rounded-md.bg-white` (+ `outline` focus recipe); wrapper `.x-input`, label `.x-label`, footer `.x-input-footer` |
| Select | `.x-select` (button + list), native `select` inside some filters |
| Toggle | `.x-toggle` › hidden checkbox + track `.rounded-full.border-[3px]` + knob `.translate-x-full` when on |
| Tabs | `.x-tab-group` › `a.x-tab.x-link.router-link-(exact-)active` (underline style); API explorer uses `div.x-tab` |
| Side panel | Vue `<Transition enter-from-class="translate-x-full" leave-active-class="duration-400">` › `div.fixed.z-30.inset-0.flex.justify-end` › `.x-form.rounded-l-md.shadow-lg.bg-white.sm:max-w-3xl` (teleported to `body`) |
| Modal | `x-modal`: `relative flex flex-col z-10 bg-white rounded-md shadow-lg transform transition-all`, visible state `ease-out duration-200 opacity-100 translate-y-0 sm:scale-100` |
| Popover / menu | `.v-popper.v-popper--theme-dropdown.x-popover` (row "⋮" menus, user menu) |
| Badge / tag | `.x-badge`, `.x-tag` |
| Alert | `.x-alert` (config page) |
| Upload | `.x-upload` (data management) |
| Empty box | `div.border-dashed.rounded-md` |
| Skeleton | `.animate-pulse` blocks |
| Charts | ECharts, SVG renderer (`[_echarts_instance_]`), 4 on the dashboard |
| Code editor | Monaco (`.monaco-editor`), in the API explorer and in side panels (storage object, player metadata) |
| Colours | `--primary-color-*`, `--secondary-color-*`, `--error/success/warning-color-*`, comma-separated `r, g, b`, set by the bundle at run time; brand.js overrides `--primary-color-*` space-separated |

Our overlay pages reuse these same recipes (economy `ui.js` copies the class strings), so one skin covers both, plus a few `eco-*` classes (`eco-tag`, `eco-notice`, `eco-toasts`, `eco-select`, `eco-tab-on`, `eco-nav-badge`, `eco-dialog`, `setup-overlay`, `ps-window`).

## 2. Native screens

Legend for "changes": **W** = cards become glass windows; **T** = table skin (§5.2 of the spec); **F** = form controls (§5.3); **S** = segmented tabs (§5.4); **P** = side panel → floating editor window (§6.1).

| Screen | Route | Built from | Changes in spatial | Risks | Shot |
|---|---|---|---|---|---|
| Dashboard | `#/` | 5 KPI cards (`grid xl:grid-cols-5`), node table, 4 ECharts cards, live badge, top-bar server stats | KPI cards → one KPI strip window; node table **T**; charts in windows with CSS-only theming; date/period header like the mockup | ECharts colours and LTR time axis are not reachable (§8 of spec); native digit roll animation already present; chart title contrast 1.94 in the proof (overlaps plot) | `dashboard-fa-light.jpg` |
| Players | `#/players` | header + "learn more" link + ⋮, search input, deleted-accounts **x-toggle**, pager buttons, table (29 rows), row ⋮ popover | list window with header inside, search as well, **T** with clickable rows, pager as capsules | row click navigates (no anchor) — keep the whole row clickable; popover teleported | `players-fa-light.jpg`, `players-row-menu-fa-light.jpg` |
| Player › Account | `#/players/:id` | avatar tile, id title, "show details" chip, edit (primary) + ⋮, **x-tab-group** (account, wallet, storage, friends, groups, payments, notifications), 2 cards (profile dl + JSON metadata, identity grid with device `x-select` + unlink), notes aside (`border-dashed` empty box) | breadcrumb in top bar; tabs **S**; cards **W**; notes aside becomes a narrow glass window; identity grid cells → wells | aside uses `bg-secondary-50` full height (looked like a slab in the proof until mapped to transparent) | `player-profile-fa-light.jpg` |
| Player › Edit | (panel) | **P** with 6 inputs + Monaco metadata | floating editor window | Monaco inside: no transform on ancestors after mount | `player-edit-panel-fa-light.jpg` |
| Player › Wallet | `#/players/:id/wallet` | wallet card + ledger table | **W**, **T** | ledger JSON opens in a panel with Monaco | `player-wallet-fa-light.jpg` |
| Player › Storage | `…/storage` | collection `x-select` + table | **F**, **T** | — | `player-storage-fa-light.jpg` |
| Player › Friends / Groups / Payments / Notifications | `…/friends`, `…/groups`, `…/payments`, `…/notifications` | search + table each | **T** | empty tables show a one-row message; style as empty state | `player-friends-…`, `player-groups-…`, `player-payments-…`, `player-notifications-fa-light.jpg` |
| Player › Hiro (+ inventory, progressions), Satori messages | `…/hiro`, `…/hiro/inventory`, `…/hiro/progressions`, `…/satori/messages` | tab hidden in the bar; page shows a dashed "only these systems are supported" box | empty state recipe | routes reachable only by URL; keep them working | `player-hiro-…`, `player-hiro-inventory-…`, `player-hiro-progressions-…`, `player-satori-fa-light.jpg` |
| Groups | `#/groups` | search + table | list window, **T** | — | `groups-fa-light.jpg` |
| Group › Details / Members | `#/groups/:id`, `…/members` | tabs, detail card, members table | **S**, **W**, **T** | — | `group-details-…`, `group-members-fa-light.jpg` |
| Storage | `#/storage` | collection/key/user filters (`x-input`, `x-select`), table (28 rows) | list window, filters as wells | — | `storage-fa-light.jpg` |
| Storage object | `#/storage` + row click | **P** with Monaco value editor, permission selects | floating editor window (wide, 768) | Monaco + `.context-view` in body | `storage-object-fa-light.jpg` |
| Storage indexes / entries | `#/storage-indexes`, `#/storage-indexes/:name` | table, "none registered" row; entries page | **T**, empty state | no index on this server: entries page only seen as not-found | `storage-indexes-…`, `storage-index-entries-fa-light.jpg` |
| Matches | `#/matches` | type select + search, table | **T**, **F** | live data refresh: no row animation | `matches-fa-light.jpg` |
| Leaderboards / records / details | `#/leaderboards`, `#/leaderboards/:id`, `…/details` | table; records table + detail card | **T**, **W** | no leaderboard on this server; detail pages only seen as not-found, re-check when one exists | `leaderboards-…`, `leaderboard-records-…`, `leaderboard-details-fa-light.jpg` |
| Notifications | `#/notifications` | search + table (20 rows) | **T** | — | `notifications-fa-light.jpg` |
| Payments › Purchases / Subscriptions | `#/payments`, `#/payments/subscriptions` | tabs + search + table | **S**, **T** | name clash with our economy "payments" in the rail: native is under main items, ours under Economy | `purchases-…`, `subscriptions-fa-light.jpg` |
| Chat messages | `#/chat` | `x-form` filter (type select, ids), table | **F**, **T** | — | `chat-fa-light.jpg` |
| API explorer | `#/api-explorer` | request tabs (`div.x-tab`, closable), endpoint `x-select`, user id input, send, body/session-vars tabs, Monaco | **S** for request tabs, **F**, Monaco on an opaque thick surface radius 16 | Monaco + 3 canvases (minimap); no transforms on its ancestors | `api-explorer-fa-light.jpg` |
| Runtime modules | `#/runtime` | modules table, RPC filter + table (292 rows, "API explorer" outline buttons) | two windows, **T** | longest page: blur only on floating layers, no row animation | `runtime-fa-light.jpg` |
| Settings › Users | `#/settings`, `#/settings/users` | settings tabs (native 5 + our 6 tabs today), users table, add user **P** | tabs **S**; **T**; **P** | live build shows two active tab underlines on `#/settings/brand` (audit log + brand) — check against the settings-window work before D3 | `settings-fa-light.jpg`, `settings-users-…`, `settings-users-add-fa-light.jpg` |
| Settings › General | `#/settings/general` | card with toggles | **W**, toggles | — | `settings-general-fa-light.jpg` |
| Settings › Configuration | `#/settings/config` | server version, `x-alert`, filter, config table (139 rows) | **T**, alert as a notice card | long table | `settings-config-fa-light.jpg` |
| Settings › Data management | `#/settings/data-management` | `x-form`, `x-upload` | upload drop zone as a dashed-free well | — | `settings-data-management-fa-light.jpg` |
| Settings › Audit log | `#/settings/audit-log` | table (40 rows) | **T** | — | `settings-audit-log-fa-light.jpg` |
| Login | `#/login` | split: dark RSS releases column (`rss-releases-panel`) + form (`x-form`), EN + theme | form in a centered thick window on the sky; releases as a glass list at the start side or dropped | not inside `#main`: the shell must not render the rail/top bar here; brand logo/welcome from brand.js | `login-fa-light.jpg` |
| Password reset / MFA reset | `#/reset`, `#/reset-mfa` | message + link (token required) | same window as login | only the invalid-token state is reachable without a mail | `reset-…`, `reset-mfa-fa-light.jpg` |
| 404 | `/:pathMatch(.*)*` | native not-found | empty state | — | (as `storage-index-entries`) |

## 3. Our overlay screens

All render into `#economy-root` with the native recipes above, or into overlays of their own.

| Screen | Route | Built from | Changes in spatial | Risks | Shot |
|---|---|---|---|---|---|
| Economy › Overview | `#/economy/overview` | catalog stat cards (`eco-stat-grid`), payment stats, gateway list with `eco-tag`s | KPI strip windows like the mockup dashboard, gateway rows with bars | active rail item for our sections renders white-on-white in the stock dark sidebar today (`NAV_ACTIVE` `!bg-secondary-50`) — disappears with the rail | `economy-overview-fa-light.jpg` |
| Economy › Currencies, Metas, Prizes, Shop, Gifts, Payments | `#/economy/<page>` | list page: header + new button, search + status select + pager, table with image tile / key mono / chips / status tags | the mockup's products page: list window with header inside, filter segmented with counts, **T** with selection, status as dot + text | none beyond the panel bug below | `economy-currencies-…`, `economy-metas-…`, `economy-prizes-…`, `economy-products-…`, `economy-gifts-…`, `economy-payments-fa-light.jpg` |
| Economy edit panel | `#/economy/<page>/<id>` | `ui.js` panel: `div.fixed.z-[60].inset-0.justify-end` › `.x-form` with translated inputs (language chips), reference pickers, checkboxes, toggle | the mockup's floating editor window: language segmented full width, payment chips, toggle row, footer danger / cancel / save | **bug today:** `z-[60]` is not in the native CSS bundle, so the holder and its shade have `z-index: auto`; the sticky top bar (z-30) and the sticky table head (z-10) paint over the open panel (see shot). Fix with an existing class or a CSS rule in `economy.css` | `economy-product-panel-fa-light.jpg` |
| Payment detail | `#/economy/payments/<id>` | `eco-detail-grid`, logs (`eco-log`), `eco-pre` JSON | windows + dl + log list | — | (route seen through e2e) |
| Base › Categories, Tags | `#/base/categories`, `#/base/tags` | list pages | list window | moving into the project settings window (`settings.js`, `PS_LEGACY`) | `base-categories-…`, `base-tags-fa-light.jpg` |
| Publication › FAQ, News, Comments, Notifications, Compose | `#/publication/<page>`, `…/notifications/new` | list pages; markdown preview, cover upload; comment moderation; compose form | list windows, editor windows, compose as a full page form window | markdown preview typography on glass | `pub-faqs-…`, `pub-news-…`, `pub-comments-…`, `pub-notifications-…`, `pub-compose-fa-light.jpg` |
| Settings tabs (ours) | `#/settings/payment-methods`, `sms`, `email`, `languages`, `brand`, `setup` | cloned native `.x-tab` items (`eco-tab`), forms with toggles and masked secrets, languages coverage table, brand form + light/dark preview, setup state | move into the project settings window (§6.3 of spec); brand preview must show the glass look (sky + glass card) instead of the flat slate boxes | being moved now by the settings-window work: re-inventory after it lands | `settings-payment-methods-…`, `settings-sms-…`, `settings-email-…`, `settings-languages-…`, `settings-brand-…`, `settings-setup-fa-light.jpg` |
| Project settings window | `#/project-settings/<id>` (new, `settings.js`) | `.ps-shade` z 55 + `.ps-holder` z 55 › `.ps-window`, section list + body | thick window per spec §6.3 | not deployed on the live server yet | — |
| Setup wizard | auto-open once; `#/settings/setup` | `.setup-overlay` (z 65, full screen, solid `secondary-50`), step bar, forms | sky + one centered thick window, step bar as segmented | must stay above the rail (z 70) | (05-brand-setup spec) |
| Toasts | — | `.eco-toasts` (z 80) and native `.x-notifications` | thick toasts | two systems, style both | — |

## 4. Technique proof

Throwaway, nothing committed, nothing under `console/ui/` touched. A prototype stylesheet and script were injected into the live console with Playwright `addStyleTag`/`addScriptTag` (sources kept in `.local/spatial-inventory/proof/src/`: `proof.css`, `proof.js`, runner `proof.mjs`).

What it does: puts `html.sp` on the page, draws the mockup's sky (SVG, rebuilt when `html.dark` changes), hides `div:has(> .x-drawer)`, builds a 64px rail from the hidden drawer's links (direct items for the first two groups, group buttons with flyouts for Economy / Publishing / Developer, settings at the bottom, badge sums from `eco-nav-badge`), restyles the native top bar as a glass bar, turns cards, table cards, inputs, tabs and side panels into the spatial recipes, and adds motion: route enter on `router.afterEach` (router read from `#app.__vue_app__`), flyout open/close, window scale+fade, hover/press.

Screenshots (`.local/spatial-inventory/proof/`):

| Page | Before (fa light) | After light | After dark |
|---|---|---|---|
| Dashboard | `dashboard-before-light.jpg` | `dashboard-after-light.jpg` | `dashboard-after-dark.jpg` |
| Players list | `players-before-light.jpg` | `players-after-light.jpg` | `players-after-dark.jpg` |
| Player detail (tabs) | `player-before-light.jpg` | `player-after-light.jpg` | `player-after-dark.jpg` |
| Economy shop | `products-before-light.jpg` | `products-after-light.jpg` (+ `products-after-light-en.jpg`, LTR with flyout) | `products-after-dark.jpg` |
| Motion | `motion-flyout-70ms.jpg` → `motion-flyout-open.jpg`; `motion-editor-90ms.jpg` → `motion-editor-open.jpg` / `motion-editor-open-dark.jpg`; `motion-route-110ms.jpg` (page mid-fade over the native skeleton); `motion-row-hover.jpg` | | |

Size of the overrides: 132 lines of CSS, 201 lines of JS. 37 distinct native classes used in selectors (most used: `.fixed`/`.inset-0`/`.justify-end` for the panel holder, `.appearance-none` for inputs, `.bg-white`, `.rounded-md`, `.shadow`, `.x-tab`, `.x-form`, `.x-button`, `.x-table-head`, `.text-secondary-400…900`). `!important`: 5 (2 for the scrim, which carries `opacity-60` and a utility background; 3 in the reduced-motion block). Everything else wins by specificity (`html.sp #main …` beats single-class utilities), no load-order tricks.

What broke, and the fix that worked:

1. **The sky vanished behind `body`** when given `z-index: -1` — `body`'s background paints over it. Fix: background on `html`, `body` transparent.
2. **Giving `#app` a z-index put the page above our own edit panel**, which exposed a real bug: economy `ui.js` uses `z-[60]`, a class that does not exist in the native bundle, so the panel is `z-index: auto` and the sticky top bar and table head already cover it on the live console (`economy-product-panel-fa-light.jpg`). Fix in the proof: explicit `z-index` tokens on the holder and scrim; the real fix belongs in economy.
3. **Route animation on the editor's own route** (`/economy/products/<id>`) replayed the page and transformed the panel's ancestor. Fix: animate only when the page changes, not the id; remove the class on `animationend`.
4. **Contrast**: brand green on white buttons 3.81, links 3.48–3.75, success tags in dark 3.00, accent text on the active dark tab 2.67, native top-bar stats 2.34 (fixed by mapping `text-secondary-400` to `ink-3`). Token text passes everywhere; fixed shades fail. Hence contrast-picked accent shades in spatial.md §1.4.
5. **Native charts** keep their own colours and LTR time axis (instance not reachable).
6. **Dark tab** active label in accent colour failed; selected segments use `ink-1`.
7. The native side panel's slide-in became a scale+fade by redefining `.translate-x-full` / `.duration-400` on the holder — Vue still drives the transition. Our own panel has no transition, so it gets a mount keyframe.
8. The flyout's page count used a hard-coded Persian word; the real rail must use the catalog.

Worked without changes: RTL and LTR (the rail moves to the start side in both through logical properties; the flyout mirrors with a `[dir=ltr]` variable), l10n's logical stylesheet, the brand accent (all accent tokens derive from `--primary-color-*`, so the green brand coloured the rail, buttons, toggles and checkboxes), the dark toggle (tokens keyed on `html.dark`), economy's sidebar injection (the rail reads it), no horizontal scroll on any of the four pages.

## 5. Recommended architecture

**A third overlay, `console/ui/spatial`**, built and injected exactly like `economy` and `l10n`:

```
console/ui/spatial/
  package.json            build / check / test (node --test)
  scripts/build.mjs       bundles src → dist/static/spatial-<hash>.{css,js}(+.gz), injects a
                          <!-- nakama-spatial:start/end --> block, writes dist-manifest.json;
                          --check also runs the selector check below
  selectors.json          every native hook: { id, selector, cssClasses[], routes[], min }
  src/tokens.css          §1 tokens, light + dark, z scale, motion tokens
  src/sky.js              background layer
  src/skin.css            native component overrides, all under html.sp, keyed off the recipes in §1
  src/eco.css             overrides for eco-* classes until economy adopts spatial classes
  src/shell.js            rail (collapsed/pinned/flyouts), top bar, user menu, mobile sheet,
                          route motion, KPI count-up, th→data-label for mobile cards
  src/i18n.js             fa/en strings for the shell
  test/                   build idempotency, selector manifest, token contrast table
```

Why not extend economy: economy is large, changes daily (it is being edited right now) and owns pages; the skin is cross-cutting, must be switchable as a whole (`html.sp` on/off, `?spatial=0` for support), and must be testable against upstream bundle updates on its own. Economy still changes in D4 to use spatial classes directly for its own pages, and gets two small hooks: `window.NakamaConsole = { router, lang, t, sections, badges, onRoute }` (economy already captures the router through the `__vue_app__` trap and deletes the trap afterwards, so a second overlay cannot trap it again), and the `z-[60]` fix.

Coexistence:

- **Load order** in `index.html`: l10n runtime (sync) → economy (defer) → spatial (defer). Stylesheets: l10n.css, economy.css, native logical CSS, **spatial.css last** (its block is anchored after the `data-l10n-source` link). Specificity does the real work; order only settles ties.
- **l10n**: spatial uses logical properties only; strings it adds come from its own catalog keyed by `lang()`; its DOM carries `data-l10n-skip`. The rail reuses labels from the drawer, which l10n has already translated.
- **Brand accent**: tokens derive from `--primary-color-*` with `color-mix` (works for both comma and space forms). `brand.js` additionally writes `--sp-accent`, `--sp-accent-ink` (light/dark) picked by contrast (§1.4) into its existing `#brand-accent` style tag.
- **Dark toggle**: economy's toggle stays the single switch (`html.dark` + stored mode); spatial only reads the class.
- **Upstream robustness**: (a) `selectors.json` lists every class spatial depends on; `npm run check` fails when one of them no longer exists in `dist/static/index-*.css` (no server needed — this would have caught `z-[60]`); (b) an e2e spec visits every route of §2/§3 and fails when a manifest selector matches fewer elements than `min` on its routes; (c) skin rules never use positional selectors (`:nth-child` of native layout) except the documented `#main > div.sticky:first-child` top bar, which is also in the manifest; (d) if the shell cannot find `.x-drawer` it does not hide the stock sidebar (fail open to the stock UI, never to an empty screen).

## 6. Regenerating the inventory

Scripts in `.local/spatial-inventory/proof/src/` (git-excluded; they write a session file next to themselves, set `SP`/`OUT` in `common.mjs` first): `common.mjs` (login from `~/.config/nakama-farsi/mac-dev.env`, one Chrome per run, closes the browser not pages), `routes.mjs` (router dump), `ids.mjs`/`ids2.mjs` (sample player/group ids by clicking rows), `crawl.mjs` (screens + component census), `extras.mjs` (login, panels, menus), `proof.mjs` (before/after/motion + contrast input), contrast with `contrast.py` (needs Pillow). D6 turns the useful parts into `e2e/specs/console/10-spatial-visual.spec.ts`.

## 7. As built (D0–D6, 2026-11-10)

The architecture of §5 was built as described: `console/ui/spatial` (sources `src/*.js` + `src/*.css`, `scripts/build.mjs`, `selectors.json`, `test/`), injected last before `</head>` as a synchronous script so `html.sp` is set before the first paint, `?spatial=0` / `?spatial=1` persisted in `localStorage['nakama.spatial']`. Economy exposes `window.NakamaConsole` (`router`, `lang()`, `t()`, `sections()`, `badges()`, `onRoute(fn)`, `projectSettings.{allowed,href,isOpen,open,sections}`, `brand()`, `toggleTheme()`, `isDark()`) and fires `nakama-console:ready`; spatial never traps the router itself. `brand.js` writes `--sp-accent`, `--sp-accent-ink`, `--sp-accent-ink-dark` picked by contrast (green brand: white on shade 600 = 4.53, links on 700 = 4.73, dark links on 200 = 5.04).

`npm run check` in `console/ui/spatial` fails when (a) a class listed in `selectors.json` disappears from the console bundle (CSS, or a component name `av('Name')`/`` `XName` `` in the JS for `x-*` markers), or (b) spatial CSS uses a native class that is not listed. `e2e/specs/console/10-spatial-visual.spec.ts` checks every manifest selector on its routes.

### Native parts reached after the UI audit (2026-11-11)

| What | How |
|---|---|
| Native chart time axis runs right to left in fa | `economy/src/native.js` finds the vue-echarts instances through the Vue vnode tree (`#app._vnode`), patches `setOption` on their prototype once and rewrites every option: `xAxis.inverse` and `yAxis.position: 'right'` in RTL, title on the right, tooltip `direction: rtl`, Persian digits and «هزار/میلیون» on the value axis, second-level labels when the data spans under 15 minutes, brand accent on the default purple series |
| Monaco followed the light theme in dark mode | `economy.css` overrides the `--vscode-*` colours under `html.dark`; `native.js` reads Monaco's `style.monaco-colors` and writes lighter `.mtkN` token colours for dark mode; minimap and overview ruler canvases are inverted. Persian in JSON uses Vazirmatn instead of Courier New's monospaced Arabic |
| Server stats in English read `NaN` / `–` | the l10n `Intl` patch now runs for English too, so number-flow parses Latin digits when the browser default is fa-IR |
| Native dialogs kept English strings (`Confirm Deletion`) | the strings live in the confirm composable as `.value=` assignments and `||` fallbacks; `scripts/audit.mjs` now extracts both and the catalog has the `upstream` and `auditMessages` namespaces |
| Native tab indicator | the native 2px underline animates `left/width`; it is hidden and spatial draws its own capsule moved by `transform` (`.sp-tabind`) |

### Deviations from `spatial.md`, decided while building

- Without a brand accent the default accent is the console's own purple (`#6b59ee` palette, shades picked by the same contrast rule), not the mockup blue: the native components keep using `--primary-color-*`, and two different accents on one screen looked wrong. A brand accent replaces both.
- `--z-settings` is 55 (economy's own value): the editor panel opened from inside the settings window must sit above it (60).
- Our editor window is 680px wide, not 536: product and prize forms have five-column item rows and reference pickers whose popovers were clipped at 536.
- The dark-mode moon is `rgb(255 236 200 / .30)` (spec `.55`): text sitting on the sky right under it measured 3.91:1.
- `--pos` light is `#146638` (spec `#17703f`) and dark `--ink-2` is `#c5cad4` (spec `#bec3ce`) so status text and secondary text stay ≥ 4.5:1 on every glass level over the darkest hill.
- Danger buttons use a fixed `--danger-fill: #b42318` (native `#ef4444` with white text measured 3.76:1).
- Visual regression is screenshots attached to every `10-spatial-visual` test (fa light, fa dark, en, 390) plus the measured contrast gate; pixel baselines were left out because seeded data (dates, uptime, notifications created by other specs) changes between runs.
