# Spatial design system (console)

The look the user approved on 2026-11-10 for the whole admin console: a soft sky behind everything, frosted-glass windows floating on it, a floating icon rail instead of the dark sidebar, and a glass top bar. It applies to every section: the native Nakama pages and our overlay pages (base, economy, publication, settings, wizard, brand).

Sources of truth:

- Approved mockup: `.local/redesign-spatial/{index.html,products.html,dashboard-dark.html}` and `shots/` (git-excluded). Every value below was taken from its generator (`gen.mjs`), not guessed.
- Proven recipes reused from `creative-platform` (`packages/ui/src/spatial.css`, `spatialPage.css`, `apps/web/src/features/shell/spatialRail.module.css`): blur only on floating layers, `prefers-reduced-transparency` fallback, capsule buttons, shimmer.
- Technique proof and native DOM recipes: `docs/design/spatial-inventory.md`. Order of work: `docs/design/spatial-plan.md`.

Built in D0–D6 (`console/ui/spatial`, see `spatial-plan.md` § Status). Where the build changed a value it is changed below too; the reasons are in `spatial-inventory.md` §7.

All CSS below is scoped under `html.sp` (the class the spatial overlay puts on `<html>`), so removing that class gives back the stock look.

## 1. Tokens

### 1.1 Shape, spacing, type, layers (both themes)

```css
html.sp {
  --r-window: 28px;   /* page windows, panels, KPI strip */
  --r-editor: 30px;   /* floating editor window, settings window */
  --r-panel: 22px;    /* rail, top bar */
  --r-flyout: 20px;
  --r-card: 16px;     /* cards inside a window */
  --r-field: 14px;    /* inputs inside an editor window */
  --r-control: 12px;  /* inputs on pages, rail items, icon tiles */
  --r-pill: 999px;    /* buttons, segmented controls, chips, badges */

  --gap: 16px;        /* outer gutter: screen edge to rail/top bar, between windows */
  --pad-window: 20px 24px 16px;
  --pad-list: 24px 24px 12px;
  --rail: 64px;  --rail-full: 232px;
  --topbar-h: 64px;
  --control-h: 40px;  --field-h: 44px;  --row-h: 52px;  --row-h-list: 56px;

  --font: "Vazirmatn", system-ui, sans-serif;
  --mono: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
  --fs-display: 28px/1.3 600;   /* page title (dashboard date) */
  --fs-title: 24px/1.3 600;     /* list window title */
  --fs-editor: 20px/1.4 600;    /* editor window title */
  --fs-h3: 16px/1.5 600;        /* panel title, top bar title */
  --fs-body: 14px/1.6 400;
  --fs-small: 13px;             /* table cells, nav, panel subtitles */
  --fs-meta: 12px;              /* th, hints, timestamps, badges */
  --fs-kpi: 32px/1.25 500;

  --z-sky: -1; --z-content: auto; --z-sticky: 10; --z-topbar: 30; --z-rail: 40; --z-flyout: 45;
  --z-scrim: 50; --z-window: 60; --z-settings: 55; --z-wizard: 70; --z-toast: 80;

  --blur: blur(32px) saturate(165%);
}
```

Body text is 14px, line-height 1.6, `font-feature-settings: "tnum"`, `-webkit-font-smoothing: antialiased`. Weights in use: 400, 500, 600 only. Machine text (keys, ids, SKUs, JSON, code) uses `--mono` at 12px with `direction: ltr; unicode-bidi: isolate`.

### 1.2 Colour (light)

```css
html.sp {
  color-scheme: light;
  --base: #e7ebf1;
  --ink-1: #12151c;  --ink-2: #414755;  --ink-3: #4f5665;
  --glass-top: rgb(255 255 255 / .70);  --glass-bot: rgb(255 255 255 / .56);
  --thick-top: rgb(255 255 255 / .90);  --thick-bot: rgb(255 255 255 / .84);
  --line: rgb(18 21 28 / .08);  --line-2: rgb(18 21 28 / .14);
  --well: rgb(18 21 28 / .05);  --well-2: rgb(18 21 28 / .08);
  --capsule: rgb(255 255 255 / .74);  --capsule-hi: rgb(255 255 255 / .96);
  --hl: inset 0 1px 0 rgb(255 255 255 / .95);
  --shade: 20 24 33;  --depth: .24;  --depth-lift: .36;
  --pos: #146638;  --neg: #ad3226;  --warn: #7f5300;  --danger-fill: #b42318;
  --star: #e5a50a;  --star-edge: #a26f00;
  --scrim: rgb(30 36 50 / .10);
}
```

### 1.3 Colour (dark, `html.sp.dark`)

```css
html.sp.dark {
  color-scheme: dark;
  --base: #0a0c12;
  --ink-1: #f1f2f6;  --ink-2: #c5cad4;  --ink-3: #a0a6b4;
  --glass-top: rgb(52 55 70 / .62);  --glass-bot: rgb(40 43 56 / .62);
  --thick-top: rgb(42 45 58 / .92);  --thick-bot: rgb(34 36 47 / .92);
  --line: rgb(255 255 255 / .09);  --line-2: rgb(255 255 255 / .15);
  --well: rgb(0 0 0 / .24);  --well-2: rgb(0 0 0 / .32);
  --capsule: rgb(70 74 90 / .55);  --capsule-hi: rgb(86 90 108 / .7);
  --hl: inset 0 1px 0 rgb(255 255 255 / .11);
  --shade: 0 0 0;  --depth: .6;  --depth-lift: .8;
  --pos: #71d9a0;  --neg: #ff9a8f;  --warn: #f0c463;
  --star: #f6bd2a;  --star-edge: #b98500;
  --scrim: rgb(0 0 0 / .34);
}
```

`ink-3` is the lightest text allowed anywhere (placeholders included). There is no lighter "disabled grey" for text; disabled controls use `ink-3` plus `opacity: .55` on the control's fill, never on its label.

### 1.4 Accent from the brand

The brand colour (Settings → Brand, `brand.js`) must drive the accent. `brand.js` already writes `--primary-color-50…950` on `html:root` as space-separated `r g b`; the native bundle writes the same variables comma-separated (`248, 250, 252`). So:

- Never write `rgb(var(--primary-color-600) / .11)` — it breaks on the comma form. Use `color-mix(in srgb, rgb(var(--primary-color-600)) 11%, transparent)`, which works for both.
- A fixed shade is not safe: the current brand (green) gives 3.81:1 for white text on shade 600 (measured in the proof). The accent tokens must be picked by contrast, in `brand.js`, every time the accent changes, and written next to the palette:

| Token | Use | Rule (light) | Rule (dark) |
|---|---|---|---|
| `--accent` | filled primary button, toggle on, checkbox on, rail/flyout badge, chart "today" line | first of shades 500, 600, 700, 800 with white text ≥ 4.5:1 | same rule |
| `--accent-ink` | links, active nav text, selected chip text, focus ring | first of 600, 700, 800, 900 with ≥ 4.5:1 on `#dfe4ee` (worst light glass over the hills) | first of 300, 200, 100 with ≥ 4.5:1 on `#4a4d5f` (`capsule-hi` over dark glass, the lightest surface accent text can sit on) |
| `--accent-soft` | active nav fill, selected row, chip on | `color-mix(in srgb, var(--accent) 11%, transparent)` | `color-mix(in srgb, var(--accent) 17%, transparent)` (light-mode `primary` tags use an opaque `color-mix(in srgb, var(--accent) 10%, #fff)` so their text keeps 4.5:1 wherever the glass sits on the sky) |
| `--accent-line` | chip-on border, input focus border | `color-mix(… 30%, transparent)` | `color-mix(… 36%, transparent)` |
| `--on-accent` | text on `--accent` | `#fff` | `#fff` |

Default accent when no brand is set: the console's own palette (`#6b59ee`) through the same rule: accent shade 500 (4.90:1), ink 600 (4.55:1), dark ink 200 (4.76:1). The mockup blue was dropped so native components (which keep `--primary-color-*`) and spatial never show two accents. Inline `rgb(...)` styles that the native app writes keep being recoloured by the existing `startRecolor()` observer.

## 2. Background (sky)

One fixed layer behind everything: `position: fixed; inset: 0; z-index: var(--z-sky); pointer-events: none`, an inline SVG `viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice"`. `html.sp` gets `background: var(--base)` and `body` becomes transparent (otherwise the body background paints over a `z-index: -1` sky — found in the proof).

Layers, bottom to top (same in both themes, colours differ):

| Layer | Light | Dark |
|---|---|---|
| vertical gradient 0 / .6 / 1 | `#dfe7f3` / `#e9edf3` / `#f4ece6` | `#0b0e1a` / `#111528` / `#1a1630` |
| wash 1, radial at (.82, .2) r .5 | `rgb(170 196 236 / .55)` | `rgb(70 96 190 / .30)` |
| wash 2, radial at (.35, 1) r .55 | `rgb(248 212 192 / .55)` | `rgb(120 70 150 / .18)` |
| glow, radial at (.22, .14) r .42 | `rgb(255 222 190 / .85)` | `rgb(255 214 150 / .20)` |
| stars | none | 140 dots, seeded PRNG (seed 7, ×16807 mod 2³¹−1), y < 560, r .7 (85%) or 1.2, opacity .25–.70 |
| sun/moon at (300, 128), blur 34 | r 48, `rgb(255 246 234 / .9)` | r 34, `rgb(255 236 200 / .30)` |
| three hills, blur 18 | `#d5dcea`, `#ccd5e6`, `#c5cfe2` | `#171a2e`, `#121424`, `#0d0f1a` |

The glow sits top-left in both directions (it is scenery, not layout; do not mirror it). The sky is rebuilt when `html.dark` changes (a `MutationObserver` on `<html>` class). With `prefers-reduced-transparency: reduce` the sky stays but every glass level falls back to its thick values.

## 3. Glass levels

| Level | Where | Recipe |
|---|---|---|
| **Base window** (`.glass`) | top bar, rail, page windows, KPI strip, panels, list window | `border: 1px solid var(--line); background: linear-gradient(180deg, var(--glass-top), var(--glass-bot)); box-shadow: var(--hl), 0 1px 2px rgb(var(--shade)/.06), 0 24px 56px -28px rgb(var(--shade)/var(--depth));` |
| **Raised / floating** (`.thick`) | editor window, settings window, modals, flyouts, dropdown menus, popovers, toasts, chart tooltip | `border: 1px solid var(--line); background: linear-gradient(180deg, var(--thick-top), var(--thick-bot)); box-shadow: var(--hl), 0 2px 6px rgb(var(--shade)/.08), 0 48px 96px -36px rgb(var(--shade)/var(--depth-lift));` |
| **Recessed** (`.well`) | inputs, search, selects, segmented-control track, event icon tiles, table hover | `background: var(--well); border: 1px solid var(--line);` (no shadow) |
| **Capsule** | secondary buttons, round icon buttons, EN/FA and period switches' selected segment | `background: var(--capsule); border: 1px solid var(--line); box-shadow: var(--hl);` selected segment uses `--capsule-hi` + `0 1px 3px rgb(var(--shade)/.14)` |

Blur (`backdrop-filter: var(--blur)`, with `-webkit-` prefix) only on layers that float over moving content: top bar, rail, flyouts, editor/settings windows, modals, toasts, sticky table heads. Page windows sit on the static sky and do not need it (the creative-platform `[data-blur]` lesson); dropping it there keeps long pages (runtime: 292 rows, config: 139) smooth. Nothing inside a window is blurred.

## 4. Layout shell

- `#main` gets `padding-inline-start: calc(var(--rail) + 2 * var(--gap))` (96px) and `padding-inline-end: var(--gap)`; with the rail pinned open, `calc(var(--rail-full) + 2 * var(--gap))` (264px).
- The stock dark sidebar (`div:has(> .x-drawer)`) is hidden, not removed: the economy overlay keeps injecting its sections into it and the rail reads them from there.
- The rail is on the **start** side (right in fa, left in en), `inset-inline-start: var(--gap)`.

### 4.1 Rail

Collapsed (default), 64px:

```
position: fixed; z-index: var(--z-rail); inset-block: var(--gap); inset-inline-start: var(--gap);
width: 64px; border-radius: var(--r-panel); padding: 10px 0; gap: 4px; flex column, centered; .glass
```

- Brand tile 40×40, radius 12, `--accent-soft` fill, brand mark (or the star) in `--accent-ink`. Under it a 28px divider (`border-top: 1px solid var(--line-2)`, margin 6px 0).
- Item 44×40, radius 12, icon 18px Lucide stroke 1.5, colour `ink-2`. Hover: `--capsule-hi` + `--hl` + `0 6px 16px -8px rgb(var(--shade)/.3)`, colour `ink-1`. Active (`aria-current`): `--accent-soft` fill, `--accent-ink` icon.
- Group item (a section with several pages): same tile plus a 3×10px "more" tick on the inline-end edge (`ink-3`, opacity .55) and, if any page has a count, a badge 16px high, radius 8, `--accent` fill, 10px/600 white, at the top inline-end corner. Active if any page of the group is active.
- Bottom: grow spacer, settings item, avatar 36px circle (`--well-2`, initial letter, 13px/600) that opens the user menu (account, sign out, Heroic links that used to live in the sidebar footer).

**Flyout** (group pages): `.thick`, width 248, radius 20, padding 8, `inset-inline-start: calc(var(--rail) + var(--gap) + 12px)`, top aligned to the hovered item (clamped to the viewport). Head row: group name 14px/600, page count 12px `ink-3` (Persian digits in fa). Rows 40px, radius 12, icon 16px `ink-2`, label 14px `ink-1`, optional count pill (`--accent`, 12px/600, line-height 18px, padding 0 7px). Hover row `--capsule-hi`; current page `--accent-soft` + `--accent-ink`. Opens on hover (no delay), on focus and on click; closes 160ms after the pointer leaves both, on Escape, on route change. `role="menu"`, items `role="menuitem"`, the group button has `aria-haspopup="menu"` and `aria-expanded`.

**Expanded (pinned), 232px:** padding 12px 10px, gap 2. Head: brand tile + name (15px/600) + game name (12px `ink-3`) + 32px collapse toggle (`--capsule`, radius 10). Group label rows 12px `ink-3`, padding 12px 10px 4px. Link rows 36px, radius 12, 14px `ink-2`; current `--accent-soft`/`--accent-ink`/500. A group shows as a row with chevron and a count chip (`--well-2`, 12px, radius 8); clicking it still opens the flyout (no nested accordion). Footer: user row (avatar, name 13px/500, role 12px `ink-3`) above a `border-top: 1px solid var(--line)`. The pinned state is remembered in `localStorage` (`nakama.spatial.rail`).

**What goes where:**

| Rail slot | Contents (route) |
|---|---|
| Direct items | Dashboard `#/`, Players `#/players`, Groups `#/groups`, Storage `#/storage`, Storage indexes `#/storage-indexes`, Matches `#/matches`, Leaderboards `#/leaderboards`, Notifications `#/notifications`, Payments `#/payments`, Chat `#/chat` |
| Divider, then groups | **General** (base: categories, tags — moving into the project settings window), **Economy** (overview, currencies, metas, prizes, shop, gifts, payments), **Publishing** (FAQ, news, comments + pending badge, notifications), **Developer** (API explorer `#/api-explorer`, runtime modules `#/runtime`) |
| Bottom | Settings: console settings `#/settings/*` (users, general, config, data management, audit log) and the **project settings window** (brand, setup, languages, categories, tags, payment methods, SMS, email — being built now in `economy/src/settings.js`, route `#/project-settings/<id>`); avatar menu |

The rail is built from the hidden `.x-drawer` (labels arrive already translated by l10n, our sections already injected by economy), so new native routes appear automatically. A group's icon is set per section (`megaphone` publishing, `coins` economy, `tag` general, `code` developer), not the first page's icon.

### 4.2 Top bar

`.glass`, height 64, radius 22, `margin-block-start: var(--gap)`, `padding-inline: 20px 12px`, sticky at `top: var(--gap)`, `z-index: var(--z-topbar)`, items gap 12.

Start → end: **title** (16px/600) or **breadcrumb** (`section ‹ page`: 14px `ink-2`, current 16px/600 `ink-1`, separator chevron 14px `ink-3`, mirrored by direction) · optional **live pill** (28px, `--well`, 12px `ink-2`, 8px dot `--pos` with a 3px 22% halo) · **section tabs** (segmented control, §6.4, for groups: economy, publishing, player tabs stay in the page) · spacer · **search** (pill, 40px, 300px wide (200 when tabs show), `--well`, search icon + placeholder `ink-3`, `Ctrl K` keycaps: mono 11px, line-height 20, `--capsule-hi`, border `--line-2`, bottom border 2px) · **EN/FA** segmented · **theme** round button 40px (moon in light, sun in dark) · the native server stats (sessions, presences, parties, matches) move into a compact popover behind a status dot, they are too low-contrast and too wide for the bar.

### 4.3 Page header

Inside `main`, before the windows: `display: flex; align-items: flex-end; justify-content: space-between; gap: 16px; padding: 28px 8px 20px`. Title 28px/600, subtitle 14px `ink-2` with 4px top margin, actions (buttons, period segmented) at the end. When the page is a single list window, the header moves inside the window (`.lhead`: title 24px/600, margin-bottom 20).

## 5. Content

### 5.1 Windows, cards, KPIs

- Page window: `.glass`, radius 28, padding `20px 24px 16px`. Panel head: title 16px/600, subtitle 13px `ink-2`, actions (link 13px/500 `--accent-ink` with a chevron) at the end, margin-bottom 12.
- Grid of windows: 2fr / 1fr rows, gap 16, `margin-top: 16px`.
- KPI strip: one window, padding 8, grid `1.05fr 1fr 1.35fr 1fr .9fr`; each KPI padding `16px 20px 18px`, separated by a 1px `--line-2` rule inset 20px top and bottom on the inline-start edge. Label 13px `ink-2` (optional status dot), value 32px/500 nowrap (unit 14px `ink-2`), change line 12px `ink-3` with `--pos`/`--neg` 500 for the number. Optional action pill at the top inline-end: 28px, `--accent-soft`, `--accent-ink`, 12px/500.
- Card inside a window (event list, gateway bars): no glass, just rows separated by `1px solid var(--line)`. Icon tile 32px, radius 10, `--well`, icon `ink-2` (`--warn` for warnings).
- Progress bar: 6px, radius 3, track `--well-2`, fill `ink-3` at .55 opacity, the leading item `--accent` at full opacity.

### 5.2 Tables

- Table sits directly in the window (no inner card). `font-size: 13px; border-collapse: collapse`.
- `th`: 12px/500 `ink-3`, `text-align: start`, padding `0 12px 10px`, no background. The native `x-table-head` (`bg-secondary-100 sticky top-0 z-10`) becomes transparent; when it is sticky it gets `.thick` + blur so rows do not show through, and `top` must equal `calc(var(--topbar-h) + 2 * var(--gap))` because the top bar now floats.
- `td`: height 52 (56 for list pages), padding `0 12px`, `border-top: 1px solid var(--line)`, `ink-1`. Secondary line in a cell: 12px `ink-3`, line-height 1.3; keys in mono.
- Status: 6px dot + 12px/500 text in `--pos`/`--warn`/`--neg`/`ink-3` (no filled pills for status).
- Row hover: `--well`. Selected row: `--accent-soft` on all cells, a 3px `--accent` inner bar on the start cell, rounded 12px ends, neighbouring borders hidden. Checkbox 18px, radius 6, border 1.5px `ink-3`; on: `--accent` fill, white check 12px.
- Amounts: 500 weight, unit 12px `ink-3` after the number.
- Footer: 12px `ink-3`, `padding: 14px 12px 4px`, `border-top: 1px solid var(--line)`: selection count at start, "N of M" at end.

**Mobile (< 768px):** every table becomes stacked cards (our pages already do this, `07-responsive.spec.ts`). Native tables have no labels on cells, so the shell copies each `th` text into a `data-label` on the matching `td` (once per render, via the observer) and CSS renders `td::before { content: attr(data-label) }`. Each row becomes a `--r-card` card on `--capsule`, gap 8.

### 5.3 Forms

| Control | Recipe |
|---|---|
| Label | 13px/500 `ink-2`, margin-bottom 6; helper on the same row at the end, 12px/400 `ink-3` (counter "۵۸ از ۲۰۰ نویسه", "optional", "cannot change after create") |
| Input / select / textarea | min-height 44 in editors (40 on pages), padding `0 14px`, radius 14 (12 on pages), `--well`, `1px solid var(--line)`, 14px `ink-1`, placeholder `ink-3`. Textarea padding `11px 14px`, min-height 72, line-height 1.7. Unit suffix 13px `ink-3` at the end |
| Focus | `--capsule-hi` fill, border `--accent`, ring `0 0 0 4px var(--accent-soft)`; caret `--accent` |
| Read-only | transparent, dashed `--line-2` border, `ink-2`, value aligned to the end, mono for keys |
| Select / reference picker | input recipe + chevron 16px `ink-3` at the end; secondary summary 12px `ink-3` before the chevron. The open list is a `.thick` popover (radius 16, padding 6, rows 40px radius 10) |
| Toggle | track 44×26 radius 13; off `--well-2` + `--line` border, on `--accent`; knob 20px white, `0 1px 3px rgb(0 0 0/.25)`, 3px inset. Label 500 `ink-1`, explanation 13px `ink-3` |
| Checkbox | 18px, radius 6 (see tables) |
| Choice chips (payment methods) | 36px pill, padding `0 14px 0 12px`, `--well`, `--line` border, 13px `ink-2`, 16px round tick at the start; on: `--accent-soft`, `--accent-line` border, `--accent-ink` 500, tick filled `--accent` with a white check |
| Two-column rows | `grid-template-columns: 1fr 1fr; gap: 12px`; one column under 640px |
| Field stack | gap 12 inside editors |

### 5.4 Segmented controls and tabs

One recipe for page section tabs, list filters ("all / active / inactive" with counts), period switch, EN/FA and content-language tabs: track `--well` + `--line` border, padding 3, radius pill, gap 2; segment 32px, padding `0 14px`, 13px `ink-2`; selected `--capsule-hi` + `--hl` + `0 1px 3px rgb(var(--shade)/.14)`, `ink-1`/500. Counts inside a segment: 12px `ink-3` (selected `ink-2`). In dark mode selected text stays `ink-1` (accent text on `capsule-hi` failed at 2.67:1 in the proof).

Language tabs in editors stretch to full width (`flex: 1`) and carry a status word: "default" `ink-3`, "complete" `ink-3`, "not translated" `--warn`.

The native `.x-tab-group` / `.x-tab` (player and group detail tabs, settings tabs, payments tabs, API explorer) keep their links; the skin removes the underline (`border-color: transparent` on the bar) and paints the active link (`.router-link-exact-active`, `[aria-current=page]`, `.eco-tab-on`) as the selected segment.

### 5.5 Buttons

- Base: 40px, padding `0 16px`, radius pill, `--capsule`, `--line` border, `--hl`, 14px/500 `ink-1`, icon 18px `ink-2`, gap 8.
- Primary: `--accent` fill, transparent border, white text/icon, `inset 0 1px 0 rgb(255 255 255/.18), 0 6px 16px -8px rgb(var(--shade)/.5)`.
- Danger (in editor footers): transparent, no border, `--neg` text and icon, padding `0 10px`.
- Round icon button: 40px circle, capsule recipe.
- Native `.x-button` keeps its `--x-button-*` colour variables (already brand-driven by economy `ui.js`); the skin only sets radius pill and motion.

## 6. Overlays

### 6.1 Drawers become floating editor windows

Every side panel (native `x-form … sm:max-w-3xl` teleported to `body`, our `ui.js` panel) becomes a floating window:

- holder: `position: fixed; inset: 0; z-index: var(--z-window); padding: var(--gap)`, panel at the inline-end side (left in fa, right in en, as today).
- window: `.thick`, radius 30, width 680 (our editors; 536 clipped the item rows and pickers) / up to 768 (native `max-w-3xl` panels keep their width), height `auto`, `max-height: 100%`, flex column. Head `padding: 20px 28px 14px` (kicker 12px `ink-3`, title 20px/600, status chip at the end). Body `padding: 14px 28px 0`, gap 12. Footer `padding: 14px 28px 16px`, `border-top: 1px solid var(--line)`: danger at start, spacer, cancel, primary.
- Under the window a handle bar (mockup): close button 28px round glass + a 132×6 grab bar `ink-2` at .38, 12px below the window. Optional; Escape and the close button must always work.
- The page behind recedes: scrim `--scrim` (z 50), the list window `transform: scale(.985)` (on the window, never on an ancestor of the overlay) and `inert` on `main`.
- Native panels that contain Monaco (storage object, player metadata, wallet ledger JSON) keep the editor on an opaque `--thick-top` surface (Monaco paints its own background).

### 6.2 Modals

Native `x-modal` (confirm dialogs, "add user"): centered, `.thick`, radius 28, padding 24, max-width per size (keep native sizes), title 16px/600, body 14px `ink-2`, footer buttons at the end. Our `eco-dialog` uses the same recipe.

### 6.3 Project settings window

The settings window being built in `economy/src/settings.js` (`.ps-window`, sections: brand, setup, languages, categories, tags, payment methods, SMS, email) is a centered `.thick` window, radius 30, `z-index: var(--z-settings)`, min(1040px, 100%)×min(760px, 100vh − 2×gap): a 236px section list (rows 40px radius 12, icon 16px, current `--accent-soft`) at the start and the section body on the right, own header with title 20px/600 and a 28px close. Entry points: the rail's settings item (flyout: "Console settings" → `#/settings/users`, "Project settings" → the window) and the gear in the top bar.

### 6.4 Popovers, menus, tooltips

`v-popper` dropdowns, row "⋮" menus, the reference picker list: `.thick`, radius 16, padding 6, rows 36–40px radius 10; tooltips: `--ink-1` background, `--base` text, 12px, radius 8, padding `4px 8px` (no glass on tooltips).

### 6.5 Toasts

Native `.x-notifications` and our `.eco-toasts`: each toast `.thick`, radius 16, padding `12px 16px`, 14px, min-width 280, an icon tile at the start coloured by kind (`--pos`, `--neg`, `--warn`, `--accent-ink`), stacked at the bottom inline-end with gap 8, `z-index: var(--z-toast)`.

### 6.6 Wizard

The setup wizard overlay (`.setup-overlay`, z 65 today → `--z-wizard`) keeps its full-screen layout but sits on the sky with one centered `.thick` window (radius 30, max-width 720) and the step bar as a segmented control.

## 7. Empty and loading states

- Loading skeletons mirror the page they replace (rule from `skeletons-mirror-the-page`): the KPI strip shows five KPI blocks, a list shows its header row and 8 rows with the real column widths, an editor shows its field stack. Blocks are `--well-2`, radius 6 (text) / 12 (controls), with one shimmer: `linear-gradient(90deg, transparent, rgb(255 255 255/.35), transparent)` (dark: `/.06`) moving across in 1.6s linear infinite, `background-attachment: fixed` so all blocks shimmer as one sweep. Native `.animate-pulse` skeletons get the same colours.
- Empty state inside a window: 40px icon tile `--well`, title 14px/500 `ink-1`, one line 13px `ink-2`, one primary action. Native dashed boxes (`border-dashed rounded-md`, e.g. player notes, hiro "not enabled") become a `--well` card with radius 16 and no dashes.
- Error state: same layout, icon `--neg`, a retry button.

## 8. Charts

- Our charts (dashboard mockup): SVG; grid `--line-2` 1px; axis labels 11px `ink-3`; today line `--accent` (dark `--accent-ink`) 2.25px, round joins; yesterday dashed `ink-3` 1.5px `3 4` at .75; event band `--star` at .09 with dashed `--star-edge` edges; "now" dot `--accent` with a 3px `--base` stroke; area gradient from `--accent` 16% (dark 20%) to 0; value tag `.thick` 12px/500.
- **Time runs right to left in fa** (now at the left edge, 00:00 at the right), the subtitle says so; left to right in en.
- Native dashboard charts are ECharts with the SVG renderer; their options are not reachable (the instance is not exposed). Skin them with CSS only: `[_echarts_instance_] text { fill: var(--ink-3) }`, grid/axis strokes to `--line-2`, series stroke to `--accent`. Their time axis stays left-to-right in both languages (documented exception) until we can hook the instance.

## 9. Focus

`:focus-visible` on every interactive element: `outline: 2px solid var(--accent-ink); outline-offset: 2px`; inputs use the 4px `--accent-soft` ring instead of an outline. The ring fades in (motion §10). Never remove focus styles from native controls; the native `outline-transparent … focus-visible:outline-secondary-300` recipe is overridden to the accent ring.

## 10. Motion

Light, smooth, everywhere, and never in the way.

### 10.1 Tokens

```css
html.sp {
  --ease: cubic-bezier(.2, .8, .2, 1);        /* enter + standard (one family) */
  --ease-exit: cubic-bezier(.4, 0, 1, 1);     /* exits only: same family, accelerating */
  --d1: 120ms;   /* hover, press, focus ring, row highlight, toggle knob */
  --d2: 200ms;   /* tabs, segmented indicator, flyouts, dropdowns, checkboxes, toasts out */
  --d3: 280ms;   /* route change, windows/editor/settings open, drawers, toasts in, list stagger */
  --d4: 480ms;   /* one-off: chart draw-in, KPI number count-up */
}
```

Exits are faster than enters: an exit uses the next shorter duration with `--ease-exit` (window close 160ms, flyout close `--d1`).

### 10.2 Rules

1. Animate only `transform` and `opacity` (plus `background-color`/`box-shadow`/`color` on hover, which do not move layout). Never `width`, `height`, `top`, `left`, `margin`, `inset`.
2. Never delay interaction: no `pointer-events: none` during animations, no waiting for `animationend` before a route or a click acts. Animations start from where things already are.
3. Never put `transform` on an ancestor of a fixed overlay, Monaco or a popover while the overlay is open: a transformed ancestor becomes the containing block of `position: fixed` children (Monaco's context menus, v-popper poppers). Route animations put the class on the content root and remove it on `animationend`.
4. Stagger at most 8 items, 30ms apart, capped at 120ms. Never animate rows of a table; a table appears as one block.
5. `prefers-reduced-motion: reduce` → durations 1ms, no delays, no shimmer, no count-up, no draw-in (the end state is shown at once).
6. Respect `prefers-reduced-transparency` (glass → thick) independently of motion.

### 10.3 Catalogue

| What | Enter | Exit | Notes |
|---|---|---|---|
| Route change (native `<main>` swap, economy root render) | content root: `opacity 0 → 1`, `translateY(6px) → 0`, `--d3 --ease`; first 8 children of the first grid/stack stagger in by 30ms from `translateY(8px)` | none (old page is gone at once) | Triggered by `router.afterEach` when the page changes, not when only the id or query changes (opening an editor changes the path to `/economy/products/<id>` and must not replay the page). No layout shift: transform only. Proven in the injected proof. |
| Rail flyout | `opacity 0 → 1`, `translateX(8px) scale(.97) → none` from the rail side (`-8px` in ltr), `--d2` | reverse, `--d1 --ease-exit`; `visibility` flips after the fade | Proven in the proof (frame at 70ms shows the fade). |
| Rail expand / collapse | rail does not animate width: the 232px panel is a second layer that fades + slides in (`translateX` from the rail side, `--d3`) and `#main` padding changes once | `--d2` | Width animation is forbidden (rule 1). |
| Top-bar tab indicator | one absolutely-positioned capsule behind the segments, moved with `transform: translateX(<offset>)` and scaled with `scaleX` on a fixed base width, `--d2` | — | Same for every segmented control (period, EN/FA, language tabs, filters). |
| Hover / press: buttons, rail items | hover `translateY(-1px)` + shadow, `--d1`; press `scale(.98)` (rail items `.94`) | — | `@media (hover: hover)` only. |
| Hover: table rows | `background-color` to `--well`, `--d1` | — | No transform on rows. |
| Hover: clickable cards/KPIs | `translateY(-2px)` + lifted shadow, `--d1` | — | Only on cards that are links. |
| Editor / settings window / modal open | `opacity 0 → 1`, `scale(.96) → 1` from the side it opens on (`transform-origin` at the inline-end edge for editors, center for modals), `--d3`; scrim fades `--d2` | `opacity → 0`, `scale(.98)`, 160ms `--ease-exit` | Native panels are a Vue `<Transition>` with `enter-from-class="translate-x-full"`, `leave-active-class="duration-400"`: the skin redefines `.translate-x-full` on the holder as `scale(.96); opacity: 0` and `.duration-400` as 160ms, so Vue still drives it and times the removal from our values. Our own panels animate with a keyframe on mount. Proven (frame at 90ms). |
| Drawers on mobile (rail sheet) | `translateX(±100%) → 0`, `--d3` | `--d2 --ease-exit` | |
| Toasts | `opacity 0 → 1`, `translateY(8px) → 0`, `--d3` | `opacity → 0`, `scale(.98)`, `--d2` | |
| Dropdowns, selects, popovers | `opacity`, `scale(.98) → 1` from the trigger edge, `--d2` | `--d1` | |
| Toggle | knob `transform: translateX(18px)` (mirrored in rtl), track colour, `--d1` | — | The native toggle already moves its knob with `translate-x-full`; l10n mirrors it. |
| Checkbox / chip tick | tick `scale(.6) → 1` + `opacity`, `--d2` | `--d1` | |
| Tab switch inside a form (language tabs) | panel content cross-fades `opacity` `--d2` | — | No height animation: the field keeps its size. |
| List items appearing (non-table lists, events, cards) | stagger rule 4 | — | |
| Skeleton shimmer | 1.6s linear infinite sweep | — | Off with reduced motion. |
| KPI numbers | count up from 0 (or the previous value) over `--d4` with `--ease`, digits in the page locale | — | The native dashboard already rolls its digits (seen mid-roll in the proof). Ours only on first paint and on period switch. |
| Charts | line `stroke-dashoffset` draw-in and area `opacity` once per page load, `--d4` | — | Never on data refresh. |
| Focus ring | `outline-color` / ring `box-shadow` from transparent, `--d1` | — | |

### 10.4 Risky native spots

- **Monaco** (storage object panel, player edit metadata, API explorer, runtime/config JSON): no transform or opacity animation on its ancestors after mount (it measures its box once; a scaled ancestor gives wrong cursor positions until the next layout). Animate the window, then let Monaco `layout()` (native already does on resize).
- **ECharts** (dashboard): do not scale its container; ECharts reads its size from the DOM and re-renders on resize. Fade only.
- **Sticky heads and the sticky top bar**: no transform on `#main` or on any scroll container, or `position: sticky` breaks.
- **v-popper poppers and `x-notifications`** are positioned by script; animate their inner box, not the positioned element.
- **Teleported panels** are created and destroyed by Vue; keep enter/exit inside Vue's transition classes (above) so it does not cut an exit animation short.

## 11. Direction, digits, fonts

- Only logical properties in spatial CSS (`inset-inline-start`, `padding-inline`, `border-start-start-radius`, `text-align: start`). Horizontal transforms read a direction variable (`--dir: 1` in ltr, `-1` in rtl) or use a `[dir=ltr]` override, as the proof's flyout does.
- l10n converts the native stylesheet to logical properties already; spatial loads after it and never reintroduces physical properties.
- Icons that mean direction (chevrons, back arrows, breadcrumb separator) are mirrored with `[dir=rtl] .ic-dir { transform: scaleX(-1) }`. Everything else is not mirrored.
- Digits follow the locale: Persian digits (۰–۹, `٬` thousands, `٫` decimal) in fa, Latin in en; Latin only on the wire and in machine text (keys, ids, JSON, code, `Ctrl K`). The proof's flyout count used a hard-coded Persian word; the real rail must use the i18n catalog.
- Font: Vazirmatn variable (already shipped by l10n as `/static/vazirmatn-nl-wght-*.woff2`, `font-display: block`) for both languages; system mono for machine text.

## 12. Contrast on glass

Every text run must be ≥ 4.5:1 against the pixels it actually sits on (glass over sky), measured, not computed from tokens. Method (reused from the mockup work): render the page, record every text run's colour and box, hide all text, screenshot, and compare each run's colour to the 10th-percentile background under its box (`contrast.py`).

Measured:

| Surface | Min | Fails |
|---|---|---|
| Mockup dashboard, light | 5.34 | 0 |
| Mockup products + editor, light | 5.23 (page behind the scrim 4.83) | 0 |
| Mockup dashboard, dark | 4.95 | 0 |
| Proof on the live console (4 pages × 2 themes) | 1.3–4.5 on a few runs | brand green accent (white on shade 600 = 3.81; links 3.48–3.75), success tag text in dark (3.00), native active tab in dark (2.67), native `text-secondary-400` stats (2.34, fixed in the proof by mapping to `ink-3`), a chart title overlapped by the plot, the decorative breadcrumb "/" |

So the token set passes; the failures come from fixed native colours and the fixed accent shade. That is why §1.4 picks accent shades by contrast and why D3 maps every native `text-secondary-*`/`text-success-*`/tag recipe to tokens. The visual-regression step (plan D6) runs this measurement on every page in both themes and fails under 4.5.
