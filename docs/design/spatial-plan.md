# Spatial redesign: implementation plan

Applies `docs/design/spatial.md` to every console screen listed in `docs/design/spatial-inventory.md`. Architecture: a third overlay `console/ui/spatial` (inventory §5). Each step ends with the full e2e suite run twice (`e2e/README.md`), the overlay's `npm test` and `npm run check`, and a fa handoff note. Nothing pushes; commits are one line.

Ground rules for every step:

- Only `console/ui/spatial/**`, the regenerated `console/ui/dist` files, and the small economy hooks named in the step. No change to the native bundle.
- All spatial CSS under `html.sp`; logical properties only; tokens only (no raw colours outside `tokens.css` and the sky).
- Each new native selector goes into `selectors.json` in the same commit.
- Screens are checked by driving the real console (log in, open the page, use it), in fa and en, light and dark, 1440 and 390.

## D0. Coordination (before D1)

- Wait for the project settings window (`economy/src/settings.js`) to land, then re-run the inventory crawl so §3 of the inventory matches what is deployed.
- In economy, as two small commits owned by whoever holds economy: (1) replace `z-[60]` in `ui.js` (panel holder + shade) with a class that exists or an `economy.css` rule — today the sticky top bar and table head cover the open edit panel; (2) expose `window.NakamaConsole = { router, lang, t, sections, badges, onRoute }` from `shell.js`.

Acceptance: the economy edit panel is above the top bar on the live console; `window.NakamaConsole.router` is defined after login.

## D1. Tokens, background, glass, dark

Build `console/ui/spatial` skeleton (package, `build.mjs` with idempotent inject/restore like economy, `dist-manifest.json`, `selectors.json`, `--check` = dist up to date + every manifest class present in `dist/static/index-*.css`). Ship `tokens.css`, `sky.js`, the glass levels, `html.sp` on by default with `?spatial=0` / `localStorage['nakama.spatial']='0'` to turn it off. Map the native colour roles to tokens: `bg-white`/`bg-secondary-50|100` surfaces, `text-secondary-400…900` → `ink-3…ink-1`, success/warning/error text and tags → `--pos/--warn/--neg`, page background → transparent. `brand.js` writes contrast-picked `--sp-accent`, `--sp-accent-ink` (light/dark), and the brand preview shows a glass card on the sky.

Acceptance:
- Every page shows the sky and glass windows in light and dark; the toggle swaps both without reload; reduced transparency gives thick glass.
- Changing the accent in Settings → Brand recolours rail, buttons, toggles, checkboxes, links, chart line live; for the current green brand white-on-accent and accent text measure ≥ 4.5:1.
- `npm run check` fails if a class in `selectors.json` is removed from the native CSS (test with a fixture).
- No horizontal scroll at 1440 or 390.

e2e: new `specs/console/09-spatial-shell.spec.ts` (part 1): `html.sp` present, sky present and swapped on dark toggle, `?spatial=0` restores the stock sidebar, accent change propagates to `--sp-accent` and a primary button's computed background, contrast of accent tokens ≥ 4.5 (computed in page). Update `01-shell.spec.ts` dark-mode test to assert tokens instead of stock classes.

## D2. Rail, top bar, settings entry

`shell.js`: hide `div:has(> .x-drawer)` only after the rail rendered (fail open); rail collapsed 64 / pinned 232 (stored), groups and flyouts per spec §4.1 with section icons, badge sums, keyboard (Tab/Arrow/Escape, `aria-haspopup`, `aria-expanded`, `role=menu`), avatar menu (account, sign out, Heroic links from the old footer). Top bar per §4.2: breadcrumb from the route, section tabs for economy/publishing, search (opens the existing Ctrl+K search if present, otherwise a page jump list built from the rail), EN/FA segmented, theme button, server stats in a popover. Settings entry: rail settings item → flyout with "Console settings" and "Project settings" (opens the settings window), gear in the top bar. Login, reset, reset-mfa: no rail, sky + centered thick window.

Acceptance:
- Every route in the inventory is reachable from the rail in ≤ 2 clicks; the current page and its group are marked; badges match the old sidebar's numbers.
- Flyout opens on hover, focus and click; closes on leave (160ms), Escape, route change; works with keyboard only.
- Breadcrumb correct on player, group, leaderboard, storage index detail and our detail routes; LTR mirrors everything.
- Project settings window opens from both entries and closes back to the page.

e2e: `09-spatial-shell.spec.ts` (part 2): rail navigation to every main route and every group page, active states, flyout keyboard path, pinned state survives reload, breadcrumb texts, EN/FA and theme from the new top bar, settings window entry. Update `01-shell.spec.ts` (sidebar section assertions move to the rail; legacy hash redirects unchanged) and `05-brand-setup.spec.ts` (brand logo/name now in the rail head and login window).

## D3. Native pages, group by group

Skin rules in `skin.css`, one group per commit, each with its selectors in the manifest:

1. **Lists**: players, groups, storage, storage indexes, matches, leaderboards, notifications, purchases/subscriptions, chat, audit log, config, runtime — list window with header inside, table skin (§5.2), sticky head as thick glass offset under the floating top bar, row hover, pager capsules, empty-row message as empty state.
2. **Details with tabs**: player (all tabs incl. hiro/satori empty states, notes aside), group, leaderboard records/details — segmented tabs, windows, identity grid wells.
3. **Overlays**: side panels → floating editor windows (Vue transition classes redefined, exit 160ms), `x-modal`, v-popper menus, `.x-notifications` toasts, tooltips.
4. **Dashboard**: KPI strip window, node table, ECharts CSS theming (text, grid, series colour), period header.
5. **Tools**: API explorer (request tabs, Monaco on an opaque surface), settings general/data management (toggles, upload well), login/reset windows.

Acceptance per group: open each screen of the group in fa/en × light/dark, use its main action (search, open a row, switch a tab, open and cancel a panel, open a menu), compare with the inventory shot and the mockup recipe side by side; contrast ≥ 4.5 on every text run (measurement from D6 run locally); Monaco cursor and menus land where clicked; sticky heads stay under the top bar while scrolling.

e2e: `10-spatial-visual.spec.ts` grows by one `describe` per group (screens listed in inventory §2), each asserting: manifest selectors match, no horizontal scroll, the group's main action works with spatial on (the existing `01…05` specs keep covering behaviour), and a screenshot is attached. Add a Monaco check (type in the storage object panel, value saved) and an ECharts check (4 SVG charts render, axis text uses `--ink-3`).

## D4. Our overlay pages

Economy, publication, base, settings tabs/window, wizard, brand: switch `ui.js`/`form.js` recipes to spatial classes where the skin would otherwise fight them (list window header inside, filter segmented with counts, choice chips for payment methods, language segmented full width with status words, toggle row with explanation, editor footer danger/cancel/primary, mockup editor width 536). Overview becomes KPI strips + gateway bars. Wizard: sky + centered window, step bar segmented. Retire `src/eco.css` rules as economy adopts the classes.

Acceptance: products list + edit match `.local/redesign-spatial/shots/2-products-edit.png` recipe for recipe; every economy/publication/base page and the settings window pass the same checks as D3; existing behaviour specs pass unchanged except selectors that named old classes.

e2e: update `02-base`, `03-economy`, `04-publication`, `05-brand-setup` helpers in `lib/console.ts` only where they select by stock classes (`panel()`, `settlePanel()`, `rows()`, `toasts()` should select by role/test id, not class); add product edit language-tab and chip interactions to `10-spatial-visual.spec.ts`.

## D5. Mobile (≤ 767px)

Rail hidden; top bar keeps a menu button (start side) that opens the pinned rail as a full-height sheet (slide from the start side, `--d3`), title, search icon, theme. Windows full width with 12px gutter, radius 22; KPI strip becomes 2 columns; tables become stacked cards (header text copied to `data-label`); editor/settings windows become bottom sheets (radius 22 top corners, max-height 92vh); flyouts become sheet sections.

Acceptance: every inventory screen at 390×844 with no sideways scroll, every row readable as a card, panels and wizard fit, all actions reachable by thumb (44px targets).

e2e: extend `07-responsive.spec.ts` to native screens (players, player tabs, storage + object panel, runtime, settings) with spatial on; sheet open/close.

## D6. Motion, end-to-end and visual regression

Motion (spec §10) lands with the component it belongs to in D1–D5; this step audits it as a whole and adds what is global:

- Route enter on `router.afterEach` (page change only), capped stagger, no table row animation.
- Segmented indicator sliding with transform; flyouts; window/editor/settings open and close (exit faster); drawers and sheets; toasts; dropdowns/selects; toggles and ticks; language-tab cross-fade; skeleton shimmer mirroring each page; KPI count-up on first paint; chart draw-in once; focus ring fade.
- Reduced motion: every animation instant; reduced transparency: thick glass.
- Performance budget: no long task > 50ms caused by spatial on runtime (292 rows) and config (139 rows); scroll at 60fps on the dashboard (trace in CI is optional, local check required).

Visual regression: `10-spatial-visual.spec.ts` takes 1440×900 screenshots of every inventory screen in fa light, fa dark, en light and the 390 set, compares them to baselines in `e2e/visual/` with a small pixel tolerance (sky is static, data is seeded by the spec), and runs the contrast measurement (port of `contrast.py` to TypeScript: record text runs, hide text, sample the background, fail under 4.5:1).

Acceptance:
- Clicking during any animation acts immediately (test: click a rail item 30ms after hovering a group; click Save 50ms after opening an editor).
- With `reducedMotion: 'reduce'` (Playwright context option) no element has a running animation 50ms after any interaction.
- No `transform` on `#main`, any scroll container, or any ancestor of an open overlay/Monaco/popover (asserted in page).
- Baselines approved by the user once; later diffs need an explicit baseline update in the same commit.

e2e: new `11-spatial-motion.spec.ts` (route enter class added and removed, flyout/editor timings within ±40ms of tokens, exit faster than enter, reduced-motion instant, interaction during animation, no transformed ancestors), `10-spatial-visual.spec.ts` baselines + contrast, `README.md` table rows for 09–11.

## Status (2026-11-10)

All steps done, one local commit each (`git log --oneline` D0…D6), screenshots per step in `.local/shots-spatial/d0…d6/` and sent to Telegram.

| Step | Result |
|---|---|
| D0 | `window.NakamaConsole` hook in economy `shell.js`; `brand.js` writes `--sp-accent*` picked by contrast; `console/ui/spatial` package with build/check/test and `selectors.json` (the `z-[60]` fix had already landed in `economy.css`) |
| D1 | tokens, sky (rebuilt on theme change), glass levels, dark, native colour roles mapped, `?spatial=0` |
| D2 | floating rail (collapsed + pinned, flyouts with keyboard, badges, settings flyout with «تنظیمات پروژه»), glass top bar (breadcrumb, section tabs with sliding indicator, page search Ctrl K, EN/FA segment, theme, server status popover, project settings gear, user menu), project settings window, sign-in window |
| D3 | every native page: list windows with the header inside, tables, tabs with a sliding capsule, side panels as floating windows, modals, menus, dashboard KPI strip and chart theming, Monaco, contrast fixes (danger buttons, primary tags, moon) |
| D4 | our pages: list windows, KPI strips on the economy overview, editor window with full-width language segment and payment chips, toasts, dialogs, brand preview on the sky, setup wizard |
| D5 | ≤ 767px: rail as a sheet from the start side, compact top bar, native tables as labelled cards, editors as bottom sheets, KPI strips in two columns |
| D6 | motion audit (one easing family, four durations, `backwards` fill so nothing keeps a transform after entering, KPI count-up, popover/toast entries, reduced motion), specs 09/10/11, contrast gate in every visual spec, docs |
