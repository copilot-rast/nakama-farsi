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
- Cluster mode on Redis (off by default).
- Updates page: this server's version and the parent repo's new commits, releases, changed files and warnings, with a copy-only update guide; build info is embedded by `build/buildinfo.sh` and the base commit is recorded in `build/upstream.json` by `build/upstream-sync.sh record`.

### Migrations
- Many new `migrate/sql/2026*` files for the base, economy and publication plugins; run `nakama migrate up` (the Docker image does it on start unless `NAKAMA_MIGRATE_ON_START=false`).

### Environment
- `ECONOMY_SECRET_KEY` (32+ characters) seals provider secrets and the optional GitHub token; storage, payment and game variables are listed in `build/nakama-start.sh`.
- `BASE_UPDATES_API_URL` (optional) replaces the GitHub API address, for tests only.
