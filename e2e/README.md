# End-to-end suite

Playwright specs that drive the real console overlay and the Star Catcher sample (`samples/stargame`) against a running server. Every phase adds specs here and runs the whole suite before it is reported done.

## Run

```sh
cd e2e
export PATH=~/.nvm/versions/node/v22.18.0/bin:$PATH
npm ci                 # once; browsers are not downloaded, the system Chrome is used
npm test               # everything
npm run test:console   # console specs only
npm run test:game      # game specs only
npm run report         # open the HTML report from the last run (e2e/playwright-report)
```

One spec or one test: `npx playwright test specs/console/03-economy.spec.ts -g "gift codes"`.

## Settings

Read from the environment first, then from `~/.config/nakama-farsi/mac-dev.env` (`E2E_ENV_FILE` points elsewhere). Credentials are never printed.

| Variable | Default |
|---|---|
| `CONSOLE_URL` | `http://100.106.1.79:7451` |
| `API_URL` | `http://100.106.1.79:7450` |
| `GAME_URL` | `$API_URL/stargame/` |
| `CONSOLE_USER`, `CONSOLE_PASSWORD` | from the env file |
| `SERVER_KEY` | from the env file (device sign-in for test players) |
| `DB_PASSWORD`, `E2E_DB_HOST`, `E2E_DB_PORT`, `E2E_DB_NAME`, `PSQL` | `127.0.0.1`, `5481`, `nakama`, Homebrew `psql@16` |
| `E2E_SERVER_LOG` | `../logs/nakama.log` (the server's output; SMS login specs read the sandbox code from it) |
| `CHROME_PATH` | `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome` |

## How it is built

- `specs/global.setup.ts` (project `setup`) backs up the shared config rows into the schema `zz_e2e_backup` (`sql/backup.sql`), records their hashes (`sql/snapshot.sql`), marks a `pending` setup as completed so the wizard does not cover other specs, and signs in once. The session is saved to `.auth/admin.json` and injected into every console page.
- `specs/global.teardown.ts` (project `teardown`, runs after everything) fails when a spec changed config without restoring it or left `e2e_…` rows behind, then restores the backup (`sql/restore.sql`), checks the hashes match the start exactly, removes leftovers (`sql/cleanup.sql`) and drops the backup schema. If a run is killed, the next setup restores the old backup first.
- Specs that must change shared config (payment methods, providers, languages, brand, setup state) call `restoreConfig()` from `lib/db.ts` when they finish. Everything else creates its own rows with unique keys (`uid()` → `e2e_<time><rand>`, device ids `e2e-…`, test languages `xa-…`) through the console API and deletes them afterwards. Test players are deleted through the console API; their payments stay as history.
- Sandbox is never turned off. The publication spec refuses to run if any active SMS/email provider is live. The SMS login specs refuse to run unless SMS login is on and in sandbox; they read the code from the server log (`lib/accounts.ts` `waitCode`), use numbers `+98999…` (`testPhone()`), clear the `otp:` rate-limit counters before sending, and delete their phone accounts and code rows afterwards.
- One worker, one browser per test. This Mac's headless Chrome crashes (SIGSEGV) as soon as any page or context is closed, which also breaks Playwright's `storageState` loading. `lib/fixtures.ts` therefore launches a fresh Chrome for each test (`page`) and for each extra page (`spawn()`), never closes pages, and closes the whole browser at the end. Do not use the built-in `browser` fixture or `context.close()`.
- Every page fails its test on an uncaught page error.
- Bank payments return to `ECONOMY_WEB_REDIRECT_URL`; on this Mac that is a static server on `:7470` that is not running, so the ZarinPal test maps that redirect to `GAME_URL` with the same query. Gateway tests skip with a reason when the sandbox host is unreachable.

## Specs

| File | Covers |
|---|---|
| `console/01-shell.spec.ts` | sign-in in fa/RTL, our two sidebar sections + the project settings entry, active states, EN/FA switch, dark mode over reload, reload/back/forward with query, every legacy hash redirects to its project settings section, Nakama's own settings tabs untouched |
| `console/02-base.spec.ts` | add/deactivate/delete a language and its tab in forms, default locked, coverage numbers move after a translation, categories and tags CRUD with key types, list filters, `IN_USE` delete, pickers filtered by key type (+ `general`) |
| `console/03-economy.spec.ts` | currency (Persian digits), meta, prize item rows, product with fa/en/ar names + methods + SKU + coin price, reward and discount gift codes, category and tag pickers on currencies and gifts (key-type filtering, quick-create, list chips, category filter, `economy_catalog` currency keys), masked secrets on payment methods, payments list and detail |
| `console/04-publication.spec.ts` | FAQ and news (markdown preview, cover upload, actions) shown in the game, comment moderation round trip, live in-app notification (toast, bell, inbox, seen), SMS/email sandbox to users with and without addresses, provider test-send |
| `console/05-brand-setup.spec.ts` | wizard auto-open once, every step, step bar, skip, «بعداً» and resume, reopen from settings, brand on sign-in/sidebar/title/favicon/accent, reset, claim-once with two admins |
| `console/08-project-settings.spec.ts` | project settings window from the top bar and the sidebar, Esc/close back to the page underneath, every section, deep link + reload + back, arrows/Home/End/Tab trap, Esc closes the editor first, ACL hiding for limited console users (created and deleted in the test), quick-create from product, prize, FAQ and news forms (empty name, duplicate names, key type), player tags kept out of content pickers, EN + dark, 390px sheet and popover |
| `console/07-responsive.spec.ts` | every overlay page at 390px: no sideways scroll, tables as cards; drawer and wizard fit; native pages with spatial on (labelled cards, 36px top-bar targets), rail sheet, bottom-sheet editor |
| `console/09-spatial-shell.spec.ts` | spatial on: sky and dark swap, `?spatial=0`/`=1`, brand accent picked by contrast, every main page from the rail, every group page through its flyout, keyboard flyout, pinned rail over reload, breadcrumbs, section tabs, badges, EN/FA and theme from the top bar, project settings from rail and gear, Ctrl K page search, sign-in window |
| `console/10-spatial-visual.spec.ts` | every native and overlay page in fa light and dark (and en): no sideways scroll, measured text contrast ≥ 4.5:1 (`lib/contrast.ts`: text runs vs the pixels behind them), screenshots attached; overlays, manifest selectors on their routes, dashboard strip + charts, Monaco typing, editor segment and chips, wizard |
| `console/11-spatial-motion.spec.ts` | page enter on route change only, durations from tokens, exits faster, flyout fade and editor scale reach rest, clicks during animations, no transform on ancestors of overlays/Monaco/popovers, KPI count-up, reduced motion |
| `console/12-accounts.spec.ts` | countries (seeded list, flags, search, inline toggle hides a country from `base_countries`, filter, create/edit/delete with field errors), account rules (field errors, try-it box, saved rules refuse a short or non-Latin username and a weak e-mail sign-up), SMS login settings (sandbox notice, limits and `{code}` validation, recent requests never show the code, verified row links to the player), app versions (editor with store links and messages, field errors, RPC answers forced/optional/ok, inactive row), a NOTIFICATION-only console user sees SMS login, SMS and email only |
| `console/15-players-chat.spec.ts` | player tags from the native player page card and its dialog, public tag through `base_my_tags`, tag filter and bulk add/remove on «مدیریت بازیکنان», notification preview and compose count by tag, product audience (catalog, purchase refused, `TAG_KIND_MISMATCH`, editor picker); suspension from the card (custom-date validation, translated reason, note) refuses device sign-in and session refresh with the JSON notice and end time, logs the old token out, shows in the player list filter and the restrictions page, early lift with a note; chat over a real socket (`lib/chat.ts`): Persian evasions masked, `CHAT_BLOCKED`, flag report, try box, new word block → mask, duplicate by normalised form; player reports (duplicate, own, unknown reason), review page context and mute decision, `CHAT_MUTED` notice with end time, delete decision removes the message; report reasons section drives `base_report_titles`; a chat-only console user sees reports and the word filter only |
| `console/14-ui-fixes.spec.ts` | regression checks from the 2026-11-11 UI audit: EN console on a fa-IR browser (Latin digits, KPIs, status popover inside the screen), Persian native delete confirmation and modal dim, Monaco dark background and Persian font, LTR fields, runtime paths, chart axis direction per language, translated tab titles, coin price names, no failing chat/Satori request, not-found toast, palette ranking, audit-log module rows, dashboard setup card, setup meta line, 390px config button, storage IDs and framed economy cards |
| `game/06-game.spec.ts` | 50 starting coins, a steered 30 s round pays coins, replay refused, STAR100 once, coin skin purchase, `INSUFFICIENT_BALANCE`, Bazaar/Myket simulators, token reuse refused, restore, HALF50 on Zibal and ZarinPal sandbox pages, ZarinPal «پرداخت» back to the game, fa/en/ar switch, news/FAQ/inbox screens |
| `game/07-responsive.spec.ts` | game screens at 390px, including the account screen |
| `game/16-chat.spec.ts` | two Star Catcher players in a test room (`?room=`): live chat with masking, banned word refused with a toast, report dialog (reason required, note, duplicate); live mute banner with reason and end time, still muted after a reload, lifted live; warning dialog; live suspension screen with reason and end time, still suspended after a reload, back after the lift; public player tag on home and account and the VIP pack only for the tagged player |
| `game/13-accounts.spec.ts` | guest links a phone (wrong code, Persian digits, resend wait), stays signed in over a reload, the same number on a second device reaches the same account, sign-out back to the guest; new phone account without keeping progress; country picker saves my country, dialling code follows it, a disabled country disappears; username rules from the console in the account screen; optional-update banner and «بعداً», forced-update screen that blocks navigation, both driven by the console row |

## Adding specs for a new phase

1. Put console journeys in `specs/console/NN-<phase>.spec.ts`, game journeys in `specs/game/`. Import `test`/`expect` from `lib/fixtures.ts`.
2. Use `open(page, '#/…')`/`go()`/`openSettings(page, section)` and the helpers in `lib/console.ts` (`root` is the project settings body while the window is open, else `#economy-root`; `settlePanel`, `fillTranslated`, `pickRef`, `savePanel`, `search`, `rows`, `toasts`), `lib/game.ts` for the game and `lib/api.ts` for setup and checks through the APIs.
3. Create data with `uid()` keys and delete it in `afterAll`. Chat specs use rooms named `e2e-…` and delete their reports and messages afterwards (`base_chat_report.channel_label`, `message.stream_label`); `base_chat_word` and `base_report_title` are backed up and hashed like the other shared config tables. If the phase adds a new shared config table, add it to `sql/backup.sql`, `sql/restore.sql` and `sql/snapshot.sql`; if it adds new keyed rows, add them to `sql/leftovers.sql` and `sql/cleanup.sql`.
4. Assert what a user sees (text, state, numbers in the page), then cross-check through the API.
5. Run `npm test` twice in a row before reporting.
