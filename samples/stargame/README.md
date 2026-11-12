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


## Events, levels and the daily reward (phase 4)

- The seed also adds the metas `xp` and `level`, the event types `round_finished` (server only, value = score, 0–400) and `golden_star` (sent by the game, value 1, 20 a minute), eight coin prizes (10 … 200), four event prizes (three rounds a day → 20 coins automatically; ten golden stars a day → 15 coins to collect; a 100-point round → 50 coins once; every 1000 points in total → 100 coins automatically, at most 20 times), two meta rules (round score → `xp`, golden star → 2 `xp`), the level table `star_levels` (`xp` → `level`, six levels from «تازه‌کار» to «افسانهٔ کهکشان», a prize to collect from level 2) and the streak `daily_login` (sign-in, 7-day cycle, one grace day, 10/15/20/25/30/40/100 coins to collect).
- `stargame_finish` records `round_finished` through `internal/playerevents` and returns the result as `events`; the round result then shows a level-up line and new or automatic rewards. Each golden star caught sends `economy_event_record` from the game.
- Home shows the level card (badge, title, experience bar, «جایزهٔ سطح … را بگیر») and the daily sign-in panel (seven days, collect today's prize, notes for today, the grace day and a broken run). «جایزه‌ها» lists every goal with its progress for today, this week or all time, and a collect button; the menu shows a dot with the number of rewards waiting.

## Support and growth (phase 5)

- The seed adds the prizes `referral_inviter` (100 coins) and `referral_invited` (50 coins), sets them on the referral rule with the goal «one finished round» when the rule has no prizes or goal yet, and adds two config keys: `stargame.banner` (translated text; a `vip` rule for the `star-vip` tag) and `stargame.golden_rush` (on/off, on by default).
- Home shows the config banner (`base_remote_config` with platform `web` and the game version, fetched at start, on language change and every 30 s on home) and a «طوفان طلایی» badge; while golden rush is on, golden stars fall twice as often. The banner element carries `data-variant` with the rule that matched.
- «پشتیبانی»: live chat card (online state from `base_support_info`), the new-ticket form (department with its description, subject, text, one picture or pdf at a time uploaded in chunks) and «تیکت‌های من» with status and a «پاسخ تازه» mark; the menu shows a dot with unread replies. A ticket shows the thread with pictures, a reply box with attachments, «مشکلم حل شد» and a five-star rating once solved. Replies arrive live (notice 1210) and in the inbox; the inbox action opens the ticket.
- «گفتگوی زنده»: messages, «پشتیبان دارد می‌نویسد…», the agent's name, «پایان گفتگو» or «پایان و ساختن تیکت», rating, «گفتگوی تازه»; live notices (1211) plus a 3-second poll.
- «کد معرف»: my code with copy, the rewards for both sides, the goal, my invites, and the form to enter a friend's code in the first week (Persian digits and lower case are fine).
- «فعالیت من» (`base_my_activity`): my sign-ins, purchases, prizes, tickets, referrals and restrictions with kind tabs and «قدیمی‌ترها».
- The new code is in `web/growth.js`, loaded after `app.js`.

## Missions (tasks)

- The seed adds two task categories, five tasks, a starter task with three steps and the group «زنجیرهٔ شکارچی» (see `seed_tasks.go`). All are collected with a button, so balances in the other screens and tests do not change on their own. Rows are matched by key and never updated.
- «مأموریت‌ها» on the home screen (`missions.js`) calls `economy_tasks_list`: the chain at the top with its progress («۱ از ۳ مأموریت») and the current task, then the daily, weekly and story tasks with a progress bar, «امروز» / «این هفته», the end time, the prize and «جایزه را بگیر». A locked task says which task to finish first; the starter shows its steps. The menu card has a dot with the number of prizes to collect.
- A finished round that completes a task adds «یک مأموریت تمام شد!» to the round result. «شروع ستاره‌چین» sends an in-app message whose button opens the missions screen.

## Inventory and product rules

- The seed adds four items (see `seed_items.go`): «سبد چوبی» (starter skin, everyone has it and it is equipped first), «سبد یاقوتی» (sold by the product `star_item_ruby` for 60 coins; the product refuses a second purchase because the player owns everything in it), «سبد یخی» (120 coins, bought straight from the inventory) and «قاب طلایی» (a frame, free at level 2). It also adds the one-time «بستهٔ شروع» (`star_starter_pack`: 150 coins for 30 coins or 5000 toman, only up to level 2) with a purchase message. Rows are matched by key and never updated.
- «انبار» on the home screen (`inventory.js`) calls `economy_inventory`: items grouped by type with rarity, «مال توست» / «در بازی», and «خرید با … سکه», «باز کن», «انتخاب» or «کنار بگذار». Buying sends a fresh `requestId`. The equipped skin's `metadata.art` (`skin_ruby`, `skin_ice`, `default`) is the basket drawn in the round.
- The shop shows «یک بار» on one-time products, «مال توست» with a link to the inventory on item products the player owns, «از سطح …» on level-locked ones, and the purchase message after a coin purchase. Bought one-time products and products above the player's level disappear.

## Achievements

- The seed adds two achievement categories and five achievements (see `seed_achievements.go`): «اولین دور», the tiered «ستاره‌چین پرکار» (3 / 10 / 30 rounds for 20 / 50 / 100 coins), «شکارچی کامل» (finish the hunter chain), «هوادار ستاره‌چین» (from the gift code `STARFAN`, which also gives 20 coins and an in-app note) and the hidden «شکار افسانه‌ای» (a 300-point round). Every tier prize is collected with a button. Rows are matched by key and never updated.
- «دستاوردها» on the home screen (`achievements.js`) calls `economy_achievements`: points and count at the top, then a card per achievement with its badge (rarity colour, a lock while locked, a question mark for a hidden one with only its hint), state, tier («ردهٔ ۱ از ۳، برنزی»), a progress bar toward the next tier, the tier prizes with their state and «جایزه را بگیر» (`economy_achievement_claim`). The menu card has a dot with the number of tier prizes to collect, and a round that unlocks something adds «یک دستاورد تازه باز شد!» to the round result.
