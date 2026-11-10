# Star Catcher (ستاره‌چین)

A small web game for testing the economy, publication and base modules end to end. Plain HTML, CSS and JavaScript on a canvas; it talks to the Nakama REST API and WebSocket directly. The coin reward for a round is the Go plugin in `main.go` (`stargame_start`, `stargame_finish`).

## Play it

The plugin serves the game itself at `/stargame/` on the Nakama API port, so after a deploy it is just there:

- https://nakama.dastar.dev/stargame/
- locally: http://100.106.1.79:7450/stargame/

Served from there, the game talks to the same origin and reads the server key from `/stargame/config.json` (the `STARGAME_SERVER_KEY` runtime env, which compose and `.local/run.sh` set to the server key). The files are embedded in `stargame.so` with `go:embed`, so a change under `web/` needs a plugin rebuild.

## Run the files from a plain static server

```sh
cd samples/stargame/web
python3 -m http.server 7470
```

Open http://localhost:7470. Outside `/stargame/` it uses the Nakama API on the same host, port 7450, with server key `defaultkey`.

To play against another server, pass it once in the URL. The game remembers it in localStorage:

```
http://localhost:7470/?api=https://nakama.dastar.dev&key=<NAKAMA_SERVER_KEY>
```

`?api=reset&key=reset` goes back to the defaults.

## Shop data (automatic)

On every start the plugin creates what the game needs if it is missing, in one transaction with `ON CONFLICT DO NOTHING`, so several nodes can start together:

- languages fa, en, ar; currency `coin` (50 for new players)
- product categories `star_coin_packs`, `star_skins`; prizes `star_coins_100/500/1200`, `star_skin_gold/neon/galaxy`
- products `star_pack_small/medium/large` (Toman, ZarinPal, Zibal, Bazaar, Myket) and `star_skin_gold/neon/galaxy` (coins, in_app)
- gift codes `STAR100` (100 coins) and `HALF50` (50% off a bank payment)
- public player tag `star-vip` («عضو ویژه») and product `star_pack_vip` (1200 coins at half price) that only players with that tag see
- app versions for `web` and `android` (latest `1.4.0`, lowest `1.0.0`, store links, fa/en/ar update messages)
- payment methods, only while a method was never saved in the console: in_app active; zarinpal and zibal active in sandbox; bazaar and myket active in sandbox with package `ir.nakama.stargame`

Rows are matched by key and never updated: names, prices or methods changed in the console stay as they are. A seeded row that an admin deletes comes back on the next start. The log shows one line, `Star Catcher seed: created …` or `nothing new`. Set the runtime env `STARGAME_SEED=false` (compose: `STARGAME_SEED=false` in the environment) to turn it off.

## Server side

- The Docker image (`build/Dockerfile.local`) ships `stargame.so` next to the other plugins.
- Bank payments return to `ECONOMY_WEB_REDIRECT_URL`; compose defaults it to `https://nakama.dastar.dev/stargame/?payment={paymentId}&success={success}`. For a static server use `http://localhost:7470/?payment={paymentId}&success={success}`.
- `seed.py` is the manual alternative. Unlike the plugin it overwrites the sample rows with its own values:

```sh
CONSOLE_URL=https://nakama-console.dastar.dev CONSOLE_USER=admin CONSOLE_PASSWORD=... python3 samples/stargame/seed.py
```

## Accounts (phase 2)

- **Phone sign-in** («حساب من»): the player starts as a device guest. Pick the country (dialling code), type the number, «کد را بفرست», type the code. «پیشرفت این دستگاه را نگه دار» (on by default) attaches the number to the guest account; off, a new account is made. The game stores the returned `loginKey` and signs in with `authenticateCustom` next time; «خروج از حساب موبایل» goes back to the device guest. In sandbox (the default) no SMS is sent: the code is in the server log (`OTP sandbox code …`). Wrong codes, expiry, resend wait and limits show as messages.
- **Country picker**: «کشور من» opens a searchable list with flags (`base_countries`), saved with `base_set_country`; the phone form uses the same list for the dialling code.
- **Username**: the account screen shows the rules from `base_account_rules` and saves with Nakama's account update, so the server-side rules apply.
- **Version check**: on start the game calls `base_app_version` with platform `web` and its version (`GAME_VERSION` in `app.js`, shown at the bottom of the page). An optional update shows a banner with the store link («بعداً» hides it until the next version); a forced update replaces every screen with «باید بازی را به‌روز کنی» and the store buttons. `?v=1.3.0` pretends to be another version, for testing.

## Chat and moderation (phase 3)

- «گفتگو» on the home screen joins the room `lobby` over the WebSocket (`?room=<name>` picks another room, which the e2e suite uses) and loads the last 30 messages. Bad words come back masked from the server; a banned word is refused with a toast and the text stays in the box.
- Every message from another player has a ⚑ button: the dialog lists the reasons from `base_report_titles` in the player's language, takes an optional note and calls `base_report_message`.
- A muted player sees a banner with the reason and end time (from `base_my_status`, the live notice 1103 or the `CHAT_MUTED` error) and the input is disabled until the mute ends or is lifted (notice 1105).
- A warning (notice 1104) opens a dialog. A suspension (notice 1102, or `ACCOUNT_SUSPENDED` at sign-in or after a 401) shows the «حسابت موقتاً بسته است» screen with the reason and end time; «دوباره امتحان کن» signs in again.
- Public player tags (`base_my_tags`) show as gold chips under the greeting and in «حساب من»; products only for the player's tags carry «ویژهٔ تو».

