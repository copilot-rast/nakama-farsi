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
| `CHROME_PATH` | `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome` |

## How it is built

- `specs/global.setup.ts` (project `setup`) backs up the shared config rows into the schema `zz_e2e_backup` (`sql/backup.sql`), records their hashes (`sql/snapshot.sql`), marks a `pending` setup as completed so the wizard does not cover other specs, and signs in once. The session is saved to `.auth/admin.json` and injected into every console page.
- `specs/global.teardown.ts` (project `teardown`, runs after everything) fails when a spec changed config without restoring it or left `e2e_…` rows behind, then restores the backup (`sql/restore.sql`), checks the hashes match the start exactly, removes leftovers (`sql/cleanup.sql`) and drops the backup schema. If a run is killed, the next setup restores the old backup first.
- Specs that must change shared config (payment methods, providers, languages, brand, setup state) call `restoreConfig()` from `lib/db.ts` when they finish. Everything else creates its own rows with unique keys (`uid()` → `e2e_<time><rand>`, device ids `e2e-…`, test languages `xa-…`) through the console API and deletes them afterwards. Test players are deleted through the console API; their payments stay as history.
- Sandbox is never turned off. The publication spec refuses to run if any active SMS/email provider is live.
- One worker, one browser per test. This Mac's headless Chrome crashes (SIGSEGV) as soon as any page or context is closed, which also breaks Playwright's `storageState` loading. `lib/fixtures.ts` therefore launches a fresh Chrome for each test (`page`) and for each extra page (`spawn()`), never closes pages, and closes the whole browser at the end. Do not use the built-in `browser` fixture or `context.close()`.
- Every page fails its test on an uncaught page error.
- Bank payments return to `ECONOMY_WEB_REDIRECT_URL`; on this Mac that is a static server on `:7470` that is not running, so the ZarinPal test maps that redirect to `GAME_URL` with the same query. Gateway tests skip with a reason when the sandbox host is unreachable.

## Specs

| File | Covers |
|---|---|
| `console/01-shell.spec.ts` | sign-in in fa/RTL, our three sidebar sections and active states, EN/FA switch, dark mode over reload, reload/back/forward with query, legacy hash redirects, one active settings tab |
| `console/02-base.spec.ts` | add/deactivate/delete a language and its tab in forms, default locked, coverage numbers move after a translation, categories and tags CRUD with key types, list filters, `IN_USE` delete, pickers filtered by key type (+ `general`) |
| `console/03-economy.spec.ts` | currency (Persian digits), meta, prize item rows, product with fa/en/ar names + methods + SKU + coin price, reward and discount gift codes, masked secrets on payment methods, payments list and detail |
| `console/04-publication.spec.ts` | FAQ and news (markdown preview, cover upload, actions) shown in the game, comment moderation round trip, live in-app notification (toast, bell, inbox, seen), SMS/email sandbox to users with and without addresses, provider test-send |
| `console/05-brand-setup.spec.ts` | wizard auto-open once, every step, step bar, skip, «بعداً» and resume, reopen from settings, brand on sign-in/sidebar/title/favicon/accent, reset, claim-once with two admins |
| `console/07-responsive.spec.ts` | every overlay page at 390px: no sideways scroll, tables as cards; drawer and wizard fit |
| `game/06-game.spec.ts` | 50 starting coins, a steered 30 s round pays coins, replay refused, STAR100 once, coin skin purchase, `INSUFFICIENT_BALANCE`, Bazaar/Myket simulators, token reuse refused, restore, HALF50 on Zibal and ZarinPal sandbox pages, ZarinPal «پرداخت» back to the game, fa/en/ar switch, news/FAQ/inbox screens |
| `game/07-responsive.spec.ts` | game screens at 390px |

## Adding specs for a new phase

1. Put console journeys in `specs/console/NN-<phase>.spec.ts`, game journeys in `specs/game/`. Import `test`/`expect` from `lib/fixtures.ts`.
2. Use `open(page, '#/…')`/`go()` and the helpers in `lib/console.ts` (`settlePanel`, `fillTranslated`, `pickRef`, `savePanel`, `search`, `rows`, `toasts`), `lib/game.ts` for the game and `lib/api.ts` for setup and checks through the APIs.
3. Create data with `uid()` keys and delete it in `afterAll`. If the phase adds a new shared config table, add it to `sql/backup.sql`, `sql/restore.sql` and `sql/snapshot.sql`; if it adds new keyed rows, add them to `sql/leftovers.sql` and `sql/cleanup.sql`.
4. Assert what a user sees (text, state, numbers in the page), then cross-check through the API.
5. Run `npm test` twice in a row before reporting.
