# Clan: clans on Nakama groups, XP road, clan wars

A generic plugin (`integrations/clan`, `clan.so`).

## Decisions

- **Native groups.** A clan is a Nakama group plus a `clan` row (tag, badge, join type, minimum trophies, level, XP, totals). Membership, join requests (state 3) and the chat channel (`3.<groupId>`) are Nakama's. Roles: leader = superadmin, co-leader = admin, elder and member = member state with `clan_member.role`. All changes go through `clan_*` RPCs; native UpdateGroup/DeleteGroup/Join/Leave/Add/Kick/Ban/Promote/Demote calls on a clan group are refused (`CLAN_USE_RPC`). One clan per player (unique index).
- **Rules are settings.** Who may accept, invite, kick, promote, edit and sign up for wars (per role), co-leader cap, leader leaving (`transfer` to the next member or `block`), join-after-leave and rejoin-after-kick cooldowns, leave-after-join wait, invite days, open requests per player, create cost, name and description limits, badge lists.
- **Clan chat** is the Nakama group channel, so the base chat quota (`countGroups`) and sticker rules apply. The plugin posts system messages (`content.type = clan_system`) for join, leave, kick, role, level and war events.
- **XP road.** `xp.sources` (economy or game events with optional field filter and per-value amounts, default `task_completed`, which economy now publishes when a task completes), war results (`xp.warWin/Draw/Loss`), daily cap per member. `clan_xp_log` is unique per source, so XP is idempotent. Levels (`clan_level`) set the member cap (the group's `max_count`), free-form perks and a prize for every member at level-up (once per clan and level).
- **Wars.** Admin configs (`clan_war_config`), each in one mode: `points` (members' events score points by rule) or `duel` (a match between members of the two clans; the winner's clan scores). Schedule (CRON or manual), duration, signup (`auto` or `leader`, with a signup window), pairing (`random`, `level`, `trophies`), attacks per member, eligibility (min members, max members, min clan level, min trophies, joined hours before start), prizes for win/lose/draw/bye to participants or all members, bye result. Rounds snapshot their rules and members.
- **Idempotent scoring.** `clan_war_event` is unique on `(war, user, source)`; points sources are `p:<event>:<matchId>`, duel results are counted once per match (`r:<matchId>`). Over-limit attacks are logged as not counted.
- **Cluster.** One worker under the `clan-worker` lease schedules rounds, starts and settles them; each round transition locks the round row with `FOR UPDATE SKIP LOCKED`; payouts (`clan_payout`) are paid with `SKIP LOCKED` and retried.
- **Leaderboards** are leaderboardx boards: the plugin publishes `clan_trophies`, `clan_xp`, `clan_war_won` (owner = clan id, board `scope = clan`) and `clan_member_xp`, `clan_war_points` (owner = member). `leaderboardx_board {clan:true}` ranks only the caller's clan members on any player board (internal clan leaderboard).
- **Hooks for other plugins.** `internal/clanhooks`: `ClanOf`, `Members`, `WarRival` (two players on opposite sides of a running war).

## Client RPCs

`clan_create`, `clan_get`, `clan_mine`, `clan_search`, `clan_update`, `clan_join`, `clan_leave`, `clan_requests`, `clan_request_accept`, `clan_request_reject`, `clan_invite`, `clan_invites`, `clan_invite_accept`, `clan_invite_decline`, `clan_kick`, `clan_promote`, `clan_demote`, `clan_transfer`, `clan_levels`, `clan_xp_history`, `clan_war_current`, `clan_war_history`, `clan_war_signup`, `clan_war_opponents`. Types are in `/docs/openapi.json`.

Socket notifications (code `events.code`, default 2410): `clan_join_request`, `clan_request_accepted`, `clan_request_rejected`, `clan_invite`, `clan_member_joined`, `clan_kicked`, `clan_role_changed`, `clan_level_up`, `clan_war_started`, `clan_war_score`, `clan_war_ended`, `clan_deleted`.

## Console (`/v2/console/base/clans`, ACL `GROUP`)

Clans (profile, members, roles, kick, XP, delete), levels, war configs, rounds (create, start now, settle now, cancel), settings. Pages under «کلن‌ها» (`#/clans/…`).

## Data

`clan_settings`, `clan`, `clan_member`, `clan_request`, `clan_cooldown`, `clan_level`, `clan_level_reached`, `clan_xp_log`, `clan_payout`, `clan_war_config`, `clan_war_round`, `clan_war_signup`, `clan_war`, `clan_war_member`, `clan_war_event` (migration `20261229000001-clan-schema.sql`); leaderboardx `scope` (`20261229000002-leaderboardx-scope.sql`).
