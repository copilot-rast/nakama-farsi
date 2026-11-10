# Roadmap (agreed 2026-11-10)

Order is fixed. Each phase is verified end to end with the stargame sample. Commits stay local (no push) until the owner asks. The console keeps its current look until a redesign direction is picked.

| Phase | Scope |
|---|---|
| 0 | Branding + first-login setup wizard (done) |
| 0b | Cluster readiness (runs right after phase 0, before phase 1): audit every plugin for single-node assumptions and fix them now so the cluster phase needs no rework: uploads behind a storage interface (local disk default, S3/MinIO option), background workers safe on many nodes (claims with SKIP LOCKED or a lease, never double-run), no in-memory state that must be shared (caches get a short TTL or DB-backed invalidation, rate limits in the DB), config only from env. From now on every phase follows these rules. Database stays as it is. Done 2026-11-10: audit, rules and results in `docs/cluster-readiness.md`. |
| 1 | Project settings window like the asp dashboard: one large window with a section list (brand, setup, languages, categories, tags, payment gateways, SMS, email, and later sections), opened from the top bar and the sidebar, deep-linkable, sections gated by admin ACL. Existing settings tabs and the «عمومی» sidebar section move into it; Nakama's own settings stay. Categories and tags follow asp: key type per item, pickers filtered by key type (+ `general` for tags), a «+» quick-create next to every picker, new key type `user` for player tags. Done 2026-11-10: `#/project-settings/<section>`, details in `docs/base/spec.md` (Console pages). |
| 2 | Accounts: phone + SMS OTP login on the existing SMS providers (rate limits, sandbox shows the code in the log), account rules (username/password length), app versions per platform with forced update and store links + client RPC, countries with flags. |
| 3 | Players and chat: player tags (and targeting notifications/offers by tag), temporary suspension with end date and a message shown to the player, chat moderation (Persian-aware word filter, message reports with report titles, temporary mute, review page). |
| 4 | Events and economy phase 2: player event tracking, event → prize rules, meta rules, streaks. |
| 5 | Support and growth: tickets, departments, canned replies, live support chat; referral codes with economy rewards; remote content/config; player activity log. |
| 6 | Cluster (owner's choice: path 2): port the cluster layer of doublemo/nakama-plus (memberlist gossip, discovery, cross-node routing for sessions/presence/streams, matches, matchmaker, parties, status, leaderboard rank cache) onto this fork, keeping the official module path and heroiclabs/nakama-common so all plugins and the console overlay stay as they are. Start with a k6 load test of one node, then a spike on a separate branch with 2–3 nodes, then the full port with multi-node e2e. Database untouched. |

Out of scope (asp-specific): Workspace, Partner, SdkRegistry, Type, ModuleLink. Backups belong to the server/Dokploy, not the console.
