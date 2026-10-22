# Native Jazebe integration architecture

The read-only source consists of `../jazebe/nest-backend` (NestJS API, worker, PostgreSQL/TypeORM, Redis/BullMQ, React dashboard) and `../jazebe/nextjs` (Persian Next.js app). The destination is this Go Nakama fork, baseline `f9a48a22`, with Persian Console localization. Nakama's server lifecycle, authentication, realtime, storage, and Console code remain intact.

`integrations/jazebe/native` is a Nakama Go runtime module. The existing Docker image build compiles it as a plugin and places it in Nakama's module directory. `InitModule` registers RPCs through `runtime.Initializer`. Every RPC reads the authenticated Nakama user ID from the runtime context. There is no Jazebe `platform_user` copy or parallel JWT.

The additive migration `migrate/sql/20261022120000-jazebe-brand-planet.sql` introduces brand, planet, and membership tables in Nakama's database. Brand ownership checks occur in the SQL write statement. Planet joins are idempotent and limited to active planets. The down migration retains data.

| RPC | Implemented behavior | Remaining source parity |
|---|---|---|
| `jazebe_brand_create` | User creates a named brand | Admin roles, settings, billing, domains |
| `jazebe_brand_list` | Lists caller-owned brands | Pagination, staff membership |
| `jazebe_planet_create` | Owner creates a named planet | Full settings, lifecycle, domains, themes |
| `jazebe_planet_list` | Lists active planets | Source filters and tenant visibility |
| `jazebe_planet_join` | User joins an active planet once | Consent, invites, progression, groups, notifications |

The original missions, economy, games, lottery, media, jobs, dashboard, and user app remain unported. Local source snapshots from the earlier audit may still exist under `integrations/jazebe/nest-backend` and `integrations/jazebe/nextjs`; they are ignored by Git and excluded from Docker builds. The original Jazebe repositories remain unchanged. There is no Jazebe sidecar in the deployment.
