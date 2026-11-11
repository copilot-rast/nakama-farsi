# BlockWorld plugin (prototype, 2026-11-11)

`integrations/blockworld` → `blockworld.so`. Makes Nakama the game backend of BlockWorld (SpacetimeDB 2.12 world in `~/Projects/games/blockworld-backend`). Full design, flows and how to run: `blockworld-backend/docs/nakama-integration.md`.

## What it adds

- **OIDC issuer for the world.** `GET <issuer>/.well-known/openid-configuration` and `GET <issuer>/jwks.json` (RS256, `kid` = JWK thumbprint). SpacetimeDB validates world tokens against them; its identity is `blake3(iss|sub)`, so one Nakama account = one world identity.
- **Client RPCs** (user session):

| RPC | Input | Output / errors |
|---|---|---|
| `blockworld_token` | `{lang?}` | `{token, issuer, subject, name, username, expiresAt}` (1 h). Suspended → the phase 3 JSON notice (code 7); `ACCOUNT_BANNED`. Writes `blockworld_link`. |
| `blockworld_wallet` | `{}` | `{currency, coins, owned, prices}`; grants the currency default once (same table as economy). |
| `blockworld_buy_block` | `{blocks (1–64), expectedTotal}` | `{purchaseId, blocks, total, coins, actionIds}`; `BLOCK_NOT_SOLD`, `PRICE_CHANGED`, `ALREADY_OWNED:<name>`, `NOT_ENOUGH_COINS`. Debits with `WalletUpdate` (ledger `module: blockworld`), then purchase + outbox rows; refunds at once if those fail. |

- **Server-to-server RPCs** (`http_key`, refused with `SERVER_ONLY` from a session): `blockworld_bridge_pull {limit}` (unacked outbox, `FOR UPDATE SKIP LOCKED`, redelivered after `BLOCKWORLD_REDELIVERY_SECONDS`), `blockworld_bridge_ack {ids}`, `blockworld_bridge_report {eventKey, kind: game_won|wallet_migration, userId, data}` (once per `eventKey`; `game_won` records the player event `blockworld_game_won` through `internal/playerevents`), `blockworld_bridge_flags` (bool remote config keys `blockworld.feature.*`, default values).
- **Moderation sync** (every node, lease `blockworld-moderation-sync`): suspensions and mutes from `base_restriction` and permanent bans (`users.disable_time`) of linked players become `restrict` outbox actions; lifts and ends too.
- **Refund sweep** (lease `blockworld-purchase-refund`): a debit with no purchase row after 2 minutes is refunded once.
- **Seeds** (idempotent, on start; `BLOCKWORLD_SEED=0` turns them off): 51 block prices from BlockWorld's `shared/economy.ts`, currency `bw_coin` (default 10 000), prize `blockworld_win_100`, event type `blockworld_game_won` (server only), event rule `blockworld_every_win` (every win, auto), remote config `blockworld.feature.rps` (bool, true).
- **Tables** (`migrate/sql/20261113000000-blockworld.sql`, Down raises): `blockworld_link`, `blockworld_block`, `blockworld_purchase` (unique user + block), `blockworld_outbox`, `blockworld_restriction_sync`, `blockworld_world_event`, `blockworld_migration`.

## Runtime env

| Key | Default |
|---|---|
| `BLOCKWORLD_ISSUER` | `ECONOMY_PUBLIC_BASE_URL` + `/blockworld`; must be reachable from the SpacetimeDB host. Local: `http://127.0.0.1:7450/blockworld` (set in `.local/run.sh`) |
| `BLOCKWORLD_JWT_KEY_FILE` | `<data_dir>/blockworld/jwt-rs256.pem`, generated once (mode 600). In a cluster every node must get the same file from a secret. |
| `BLOCKWORLD_CURRENCY` | `bw_coin` |
| `BLOCKWORLD_START_COINS`, `BLOCKWORLD_WIN_PRIZE` | 10000, 100 (seed values only) |
| `BLOCKWORLD_TOKEN_TTL_SECONDS` | 3600 |

Tests: `go test ./integrations/blockworld/` (token signing and verification, JWKS, key file, config). The flows are tested end to end from the BlockWorld repo (`bridge/test/live`, `tests/e2e-nakama`).

Local only for now: `build/Dockerfile*` and the compose files do not build or ship `blockworld.so`, so a deploy of `dev` does not seed anything into a live database. Build it next to the others with `go build -buildmode=plugin -mod=vendor -o .local/build/modules/blockworld.so ./integrations/blockworld`.
