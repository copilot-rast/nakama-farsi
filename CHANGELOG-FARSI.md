# nakama-farsi change log

Changes of this fork on top of heroiclabs/nakama. The upstream `CHANGELOG.md` stays as heroiclabs writes it.

Releases are tagged on `main` as `v3.42-fa.N`: the upstream version this fork is based on, then a running number (`v3.42-fa.1`, `v3.42-fa.2`, …; after merging upstream 3.43 the next one is `v3.43-fa.1`). Each tag gets a GitHub release whose notes copy the entry below. Game forks see these tags and releases first on their «به‌روزرسانی‌ها» page, so write every entry for the people who merge it: what changed, new migrations, new environment variables, and anything they must do by hand.

## [Unreleased] — on `dev`, based on heroiclabs/nakama `37618ad4` (3.42.0 + master of 2026-11-12)

### Added
- Persian console: fa-IR with RTL and Persian digits, the Plane look (rail, sidebar per section, Yekan Bakh, Phosphor icons, light and dark), Ctrl K page search, branding and a first-login setup wizard, one project settings window with every section.
- Accounts: phone and SMS login with sandbox codes, account rules, app versions with forced update, countries.
- Players and chat: player tags, suspensions, mutes and warnings, a Persian-aware word filter, player reports.
- Events and economy: currencies, metas, prizes, shop with ZarinPal/Bazaar/Myket, gift codes, player events, event prizes, meta rules, level tables, streaks.
- Streak milestones: safe days a missed day never drops below, one-time milestone prizes, an option to pay day rewards again after a fall.
- Tasks and task groups with sub-tasks, repeats and dependencies.
- Items and inventory with buy, unlock, upgrade, equip and services; product rules.
- Achievements with tiers, progress, hidden badges and per-tier prizes.
- Banners and banner placements with schedules and player targeting.
- Chance game (wheel, loot box, slot) and prize groups.
- Sponsors with sponsored banners, a report and CSV export.
- Notifications: in-app inbox, push (FCM HTTP v1 and a log sandbox), campaigns, automatic notifications, categories and opt-out.
- Support and growth: tickets, live support chat, referral codes, remote config, player activity.
- Games: many games on one server, game settings with rooms, entry fees and prizes, matchmaking hooks, the Star Catcher sample with tic-tac-toe.
- Friend request rules: requests per day, most friends, invite expiry and a wait after a rejection, per player tag, with extra requests and slots from economy metas, refuse or trim over the limit, translated JSON errors and `base_friend_rules`.
- Player reports not tied to a chat message, with a review page, internal notes and the same decisions as chat reports, admin-set report rules; `base_report_player`, `base_my_reports`.
- Voice chat tokens: `voice_token` for matches, parties and open rooms, LiveKit and Rastaar comms providers chosen in «سرور ویس» with sealed secrets and a connection test, per-game switch and room size, allowed and listen-only player tags.
- Friend gifts: give a friend an item or currency from the inventory or buy them a shop product (coins, gateways, stores), only things they do not have; giftable per item, item type and product; friends, friendship age, mutual friends, blocks, levels and account ages; daily sender, recipient and pair limits per player tag with purchasable bonus metas; instant or accept/decline delivery with expiry and returns or refunds; a filtered message; inbox, push and live notices; `Gift:Sent`/`Gift:Received`; admin log, reverse, report and settings in «هدیه به دوست»; client RPCs `economy_friend_gift_options`, `economy_friend_gift_send`, `economy_friend_gifts`, `economy_friend_gift_accept`, `economy_friend_gift_decline`, `economy_friend_gift_cancel`. Gift purchases no longer count as the sender's own purchases of the product.
- Personal offers: an admin pool of offers over shop products; each player gets their own daily set (admin period, reset time and zone, count, weights, tags, levels, metas, no-repeat window, bought cooldown, free or paid rerolls) and entry offers (first entry, level, meta, tag, account age or a return after a break) with a personal countdown; bought through the normal shop flow at the offer price with per-offer limits; client RPCs `economy_offers`, `economy_offer_buy`, `economy_offers_reroll`, `economy_offer_seen`; console «پیشنهادهای ویژه» with preview, report, settings and a player card.
- Meta relations: when a player meta reaches a value (>, ≥, =, ≤, <), change other metas (add or set) and give a prize or one from a prize group, once, again after a drop, or every time with a cap; checked on every meta change with an admin chain depth; `MetaRelation:Completed` event; client RPC `economy_meta_relations`; console tab «رابطهٔ متاها», a meta studio side panel that shows and edits everything feeding and fed by a meta, «تغییر متا» on the player card. Purchases, gift codes, admin and game prizes that give metas now also update level tables.
- Console navigation: sidebar groups that fold and remember their state, and tabbed pages (metas, tasks, prizes, events, banners, notifications) with one address per tab; the old addresses such as `#/economy/prize-groups` redirect to the matching tab, and Ctrl K finds tabs too.
- Cluster mode on Redis (off by default).
- Updates page: this server's version and the parent repo's new commits, releases, changed files and warnings, with a copy-only update guide; build info is embedded by `build/buildinfo.sh` and the base commit is recorded in `build/upstream.json` by `build/upstream-sync.sh record`.

### Migrations
- Many new `migrate/sql/2026*` files for the base, economy and publication plugins; run `nakama migrate up` (the Docker image does it on start unless `NAKAMA_MIGRATE_ON_START=false`).

### Environment
- `ECONOMY_SECRET_KEY` (32+ characters) seals provider secrets and the optional GitHub token; storage, payment and game variables are listed in `build/nakama-start.sh`.
- `BASE_UPDATES_API_URL` (optional) replaces the GitHub API address, for tests only.
