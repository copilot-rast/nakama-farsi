# Jazebe HTTP route declaration inventory

This lists controller decorators from the read-only Jazebe NestJS source. A decorator is an entry point, not proof that its operation is integrated with Nakama identity or MCP. Global API prefix and controller path are applied at runtime. Empty method paths use the controller root.

The machine-readable audit covers all **389 source route identities and handlers**. Contract-level classification covers **28 routes**: 19 PARTIAL, 5 NOT_STARTED and 4 BLOCKED. The remaining 361 have not undergone behavioral contract review. PostgreSQL and live runtime verification remain NOT_RUN. Registered RPCs are not counted as route parity.

| ID | Method | Controller path | Method path | Source | Status |
|---|---|---|---|---|---|
| JZ-ROUTE-001 | POST | admin/languages |  | nest-backend/back-end/apps/api/src/modules/base/language/language.controller.ts | Native parity pending |
| JZ-ROUTE-002 | PUT | admin/languages | :id | nest-backend/back-end/apps/api/src/modules/base/language/language.controller.ts | Native parity pending |
| JZ-ROUTE-003 | DELETE | admin/languages | :id | nest-backend/back-end/apps/api/src/modules/base/language/language.controller.ts | Native parity pending |
| JZ-ROUTE-004 | GET | admin/languages |  | nest-backend/back-end/apps/api/src/modules/base/language/language.controller.ts | Native parity pending |
| JZ-ROUTE-005 | GET | admin/languages | :id | nest-backend/back-end/apps/api/src/modules/base/language/language.controller.ts | Native parity pending |
| JZ-ROUTE-006 | GET | client/languages |  | nest-backend/back-end/apps/api/src/modules/base/language/language.controller.ts | Native parity pending |
| JZ-ROUTE-007 | GET | client/languages | :id | nest-backend/back-end/apps/api/src/modules/base/language/language.controller.ts | Native parity pending |
| JZ-ROUTE-008 | GET | admin/language-items |  | nest-backend/back-end/apps/api/src/modules/base/language-item/language-item.controller.ts | Native parity pending |
| JZ-ROUTE-009 | POST | admin/language-items | add | nest-backend/back-end/apps/api/src/modules/base/language-item/language-item.controller.ts | Native parity pending |
| JZ-ROUTE-010 | PUT | admin/language-items | set | nest-backend/back-end/apps/api/src/modules/base/language-item/language-item.controller.ts | Native parity pending |
| JZ-ROUTE-011 | PATCH | admin/language-items | update | nest-backend/back-end/apps/api/src/modules/base/language-item/language-item.controller.ts | Native parity pending |
| JZ-ROUTE-012 | GET | admin/language-items | :id | nest-backend/back-end/apps/api/src/modules/base/language-item/language-item.controller.ts | Native parity pending |
| JZ-ROUTE-013 | DELETE | admin/language-items | :id | nest-backend/back-end/apps/api/src/modules/base/language-item/language-item.controller.ts | Native parity pending |
| JZ-ROUTE-014 | GET | admin/types |  | nest-backend/back-end/apps/api/src/modules/base/type/type.controller.ts | Native parity pending |
| JZ-ROUTE-015 | POST | admin/types | add | nest-backend/back-end/apps/api/src/modules/base/type/type.controller.ts | Native parity pending |
| JZ-ROUTE-016 | PATCH | admin/types | update | nest-backend/back-end/apps/api/src/modules/base/type/type.controller.ts | Native parity pending |
| JZ-ROUTE-017 | GET | admin/types | :id | nest-backend/back-end/apps/api/src/modules/base/type/type.controller.ts | Native parity pending |
| JZ-ROUTE-018 | DELETE | admin/types | :id | nest-backend/back-end/apps/api/src/modules/base/type/type.controller.ts | Native parity pending |
| JZ-ROUTE-019 | GET | client/types |  | nest-backend/back-end/apps/api/src/modules/base/type/type.controller.ts | Native parity pending |
| JZ-ROUTE-020 | GET | client/types | :id | nest-backend/back-end/apps/api/src/modules/base/type/type.controller.ts | Native parity pending |
| JZ-ROUTE-021 | GET | health |  | nest-backend/back-end/apps/api/src/modules/health/health.controller.ts | Native parity pending |
| JZ-ROUTE-022 | GET | health | ready | nest-backend/back-end/apps/api/src/modules/health/health.controller.ts | Native parity pending |
| JZ-ROUTE-023 | GET | health | details | nest-backend/back-end/apps/api/src/modules/health/health.controller.ts | Native parity pending |
| JZ-ROUTE-024 | GET | metrics |  | nest-backend/back-end/apps/api/src/modules/health/metrics.controller.ts | Native parity pending |
| JZ-ROUTE-025 | POST | client/auth | sign-in | nest-backend/back-end/apps/api/src/modules/identity/controllers/auth.controller.ts | Native parity pending |
| JZ-ROUTE-026 | POST | client/auth | sign-in/verify | nest-backend/back-end/apps/api/src/modules/identity/controllers/auth.controller.ts | Native parity pending |
| JZ-ROUTE-027 | POST | client/auth | refresh-token | nest-backend/back-end/apps/api/src/modules/identity/controllers/auth.controller.ts | Native parity pending |
| JZ-ROUTE-028 | POST | client/auth | logout | nest-backend/back-end/apps/api/src/modules/identity/controllers/auth.controller.ts | Native parity pending |
| JZ-ROUTE-029 | GET | client/profile | me | nest-backend/back-end/apps/api/src/modules/identity/controllers/profile.controller.ts | Native parity pending |
| JZ-ROUTE-030 | PATCH | client/profile | me | nest-backend/back-end/apps/api/src/modules/identity/controllers/profile.controller.ts | Native parity pending |
| JZ-ROUTE-031 | POST | client/account | deactivate/request-code | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/account.controller.ts | BLOCKED: trusted registered phone, OTP and SMS delivery adapter absent; source checks group creator |
| JZ-ROUTE-032 | POST | client/account | deactivate | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/account.controller.ts | BLOCKED: OTP confirmation and immediate Nakama session revocation absent |
| JZ-ROUTE-033 | GET | admin/jazebeh/users |  | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | PARTIAL: `jazebe_operator_user_list`; trusted phone/name fields and SQL/HTTP verification absent |
| JZ-ROUTE-034 | GET | admin/jazebeh/users | :userId | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | PARTIAL: `jazebe_operator_user_get`; sealed KYC detail and SQL/HTTP verification absent |
| JZ-ROUTE-035 | GET | admin/jazebeh/users | :userId/audit | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | PARTIAL: `jazebe_operator_user_audit`; only Jazebe audit records, SQL/runtime unverified |
| JZ-ROUTE-036 | POST | admin/jazebeh/users | :userId/suspend | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | PARTIAL: `jazebe_operator_user_suspend` persists state, core ban and audit atomically; local token/socket revocation and retry added; cross-node revocation and SQL/runtime behavior unverified |
| JZ-ROUTE-037 | POST | admin/jazebeh/users | :userId/unsuspend | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | PARTIAL: `jazebe_operator_user_unsuspend` conditionally clears only the recorded Jazebe core ban; SQL/runtime and source HTTP contract unverified |
| JZ-ROUTE-038 | POST | admin/jazebeh/users | :userId/reactivate | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | PARTIAL: `jazebe_operator_user_reactivate` guards deactivated state and ban provenance; self-deactivation OTP path absent, SQL/runtime unverified |
| JZ-ROUTE-039 | GET | admin/jazebeh/users | :userId/constellation | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | PARTIAL: `jazebe_operator_user_constellation` projects memberships, wallet, XP, medals and active group; source HTTP and PostgreSQL behavior unverified |
| JZ-ROUTE-040 | GET | admin/jazebeh/users | :userId/kyc | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | BLOCKED: sealed trusted KYC identity record and provider integration absent |
| JZ-ROUTE-041 | POST | admin/jazebeh/users | :userId/kyc/decision | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | NOT_STARTED: managerial decision over review-state KYC record absent |
| JZ-ROUTE-042 | PATCH | admin/jazebeh/users | :userId/kyc | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | BLOCKED: sealed identifiers and super-level correction workflow absent |
| JZ-ROUTE-043 | POST | admin/jazebeh/users | :userId/purge | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | NOT_STARTED: audited legal-purge workflow absent; no production deletion authorized |
| JZ-ROUTE-044 | GET | admin/jazebeh/kyc-queue |  | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/admin-users.controller.ts | NOT_STARTED: operator review queue over trusted KYC records absent |
| JZ-ROUTE-045 | GET | client/auth | display-name-availability | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/auth-profile.controller.ts | PARTIAL: `jazebe_profile_name_available`, source HTTP/suggestion parity unverified |
| JZ-ROUTE-046 | GET | client/auth | profile | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/auth-profile.controller.ts | PARTIAL: `jazebe_profile_bootstrap`, brand-domain auto-membership absent |
| JZ-ROUTE-047 | POST | client/auth | terms | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/auth-profile.controller.ts | PARTIAL: `jazebe_profile_terms_accept`, PostgreSQL/runtime unverified |
| JZ-ROUTE-048 | POST | client/auth | profile | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/auth-profile.controller.ts | PARTIAL: `jazebe_profile_setup`, signup listener and HTTP behavior unverified |
| JZ-ROUTE-049 | GET | client/profile |  | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/profile.controller.ts | PARTIAL: `jazebe_profile_overview`, trusted phone and translations absent |
| JZ-ROUTE-050 | PATCH | client/profile |  | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/profile.controller.ts | PARTIAL: `jazebe_profile_rename`, live persistence unverified |
| JZ-ROUTE-051 | PATCH | client/profile | avatar | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/profile.controller.ts | PARTIAL: `jazebe_profile_avatar_set`, live media lifecycle unverified |
| JZ-ROUTE-052 | GET | client/profile | verification | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/profile.controller.ts | NOT_STARTED: masked KYC detail endpoint absent |
| JZ-ROUTE-053 | POST | client/profile | verification | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/profile.controller.ts | NOT_STARTED: trusted phone, Shahkar and IBAN provider dependencies absent |
| JZ-ROUTE-054 | GET | client/users | :userId/public-profile | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/users.controller.ts | PARTIAL: `jazebe_profile_public`, source HTTP/runtime parity unverified |
| JZ-ROUTE-055 | POST | client/users | :userId/reports | nest-backend/back-end/apps/api/src/modules/jazebeh/account/controllers/users.controller.ts | PARTIAL: `jazebe_profile_report` validates caller, planet, reason and duplicate open report; source HTTP/SQL/runtime unverified |
| JZ-ROUTE-056 | GET | brand/analytics | planets | nest-backend/back-end/apps/api/src/modules/jazebeh/analytics/controllers/analytics.controller.ts | Native parity pending |
| JZ-ROUTE-057 | GET | brand/analytics | lottery/winners | nest-backend/back-end/apps/api/src/modules/jazebeh/analytics/controllers/analytics.controller.ts | Native parity pending |
| JZ-ROUTE-058 | GET | brand/analytics | exports | nest-backend/back-end/apps/api/src/modules/jazebeh/analytics/controllers/analytics.controller.ts | Native parity pending |
| JZ-ROUTE-059 | POST | brand/analytics | exports | nest-backend/back-end/apps/api/src/modules/jazebeh/analytics/controllers/analytics.controller.ts | Native parity pending |
| JZ-ROUTE-060 | GET | brand/analytics | exports/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/analytics/controllers/analytics.controller.ts | Native parity pending |
| JZ-ROUTE-061 | GET | brand/analytics | :report | nest-backend/back-end/apps/api/src/modules/jazebeh/analytics/controllers/analytics.controller.ts | Native parity pending |
| JZ-ROUTE-062 | GET | brand/analytics/exports/download | :token | nest-backend/back-end/apps/api/src/modules/jazebeh/analytics/controllers/analytics.controller.ts | Native parity pending |
| JZ-ROUTE-063 | GET | brand | me | nest-backend/back-end/apps/api/src/modules/jazebeh/brand/controllers/brand-me.controller.ts | Native parity pending |
| JZ-ROUTE-064 | GET | brand/team |  | nest-backend/back-end/apps/api/src/modules/jazebeh/brand/controllers/brand-team.controller.ts | Native parity pending |
| JZ-ROUTE-065 | GET | brand/team | lookup | nest-backend/back-end/apps/api/src/modules/jazebeh/brand/controllers/brand-team.controller.ts | Native parity pending |
| JZ-ROUTE-066 | POST | brand/team |  | nest-backend/back-end/apps/api/src/modules/jazebeh/brand/controllers/brand-team.controller.ts | Native parity pending |
| JZ-ROUTE-067 | PATCH | brand/team | :id/role | nest-backend/back-end/apps/api/src/modules/jazebeh/brand/controllers/brand-team.controller.ts | Native parity pending |
| JZ-ROUTE-068 | POST | brand/team | :id/deactivate | nest-backend/back-end/apps/api/src/modules/jazebeh/brand/controllers/brand-team.controller.ts | Native parity pending |
| JZ-ROUTE-069 | POST | brand/team | :id/reactivate | nest-backend/back-end/apps/api/src/modules/jazebeh/brand/controllers/brand-team.controller.ts | Native parity pending |
| JZ-ROUTE-070 | GET | admin/jazebeh/shop/products |  | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-admin.controller.ts | Native parity pending |
| JZ-ROUTE-071 | GET | admin/jazebeh/shop/products | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-admin.controller.ts | Native parity pending |
| JZ-ROUTE-072 | POST | admin/jazebeh/shop/products | :id/suspend | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-admin.controller.ts | Native parity pending |
| JZ-ROUTE-073 | POST | admin/jazebeh/shop/products | :id/unsuspend | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-admin.controller.ts | Native parity pending |
| JZ-ROUTE-074 | GET | brand/planets/:planetId/shop | products | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-brand.controller.ts | Native parity pending |
| JZ-ROUTE-075 | POST | brand/planets/:planetId/shop | products | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-brand.controller.ts | Native parity pending |
| JZ-ROUTE-076 | GET | brand/planets/:planetId/shop | lotteries | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-brand.controller.ts | Native parity pending |
| JZ-ROUTE-077 | GET | brand/planets/:planetId/shop | sales | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-brand.controller.ts | Native parity pending |
| JZ-ROUTE-078 | GET | brand/planets/:planetId/shop | orders | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-brand.controller.ts | Native parity pending |
| JZ-ROUTE-079 | GET | brand/planets/:planetId/shop | products/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-brand.controller.ts | Native parity pending |
| JZ-ROUTE-080 | PATCH | brand/planets/:planetId/shop | products/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-brand.controller.ts | Native parity pending |
| JZ-ROUTE-081 | POST | brand/planets/:planetId/shop | products/:id/status | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-brand.controller.ts | Native parity pending |
| JZ-ROUTE-082 | GET | brand/planets/:planetId/shop | products/:id/coupons | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-brand.controller.ts | Native parity pending |
| JZ-ROUTE-083 | POST | brand/planets/:planetId/shop | products/:id/coupons | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-brand.controller.ts | Native parity pending |
| JZ-ROUTE-084 | GET | brand/planets/:planetId/shop | products/:id/orders | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/shop-brand.controller.ts | Native parity pending |
| JZ-ROUTE-085 | GET | client/store | products | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/store.controller.ts | Native parity pending |
| JZ-ROUTE-086 | GET | client/store | products/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/store.controller.ts | Native parity pending |
| JZ-ROUTE-087 | POST | client/store | products/:id/purchase | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/store.controller.ts | Native parity pending |
| JZ-ROUTE-088 | GET | client/store | purchases | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/store.controller.ts | Native parity pending |
| JZ-ROUTE-089 | GET | client/wallet |  | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/wallet.controller.ts | Native parity pending |
| JZ-ROUTE-090 | GET | client/wallet | transactions | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/wallet.controller.ts | Native parity pending |
| JZ-ROUTE-091 | GET | admin/jazebeh/users/:userId/wallets |  | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/wallet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-092 | GET | admin/jazebeh/users/:userId/wallets | :planetId/transactions | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/wallet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-093 | POST | admin/jazebeh/users/:userId/wallets | :planetId/adjust | nest-backend/back-end/apps/api/src/modules/jazebeh/economy/controllers/wallet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-094 | GET | client/games |  | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game.controller.ts | Native parity pending |
| JZ-ROUTE-095 | POST | client/games | :id/plays | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game.controller.ts | Native parity pending |
| JZ-ROUTE-096 | POST | client/games | :id/rewards | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game.controller.ts | Native parity pending |
| JZ-ROUTE-097 | GET | admin/jazebeh | game-catalog | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-admin.controller.ts | Native parity pending |
| JZ-ROUTE-098 | POST | admin/jazebeh | game-catalog | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-admin.controller.ts | Native parity pending |
| JZ-ROUTE-099 | GET | admin/jazebeh | game-catalog/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-admin.controller.ts | Native parity pending |
| JZ-ROUTE-100 | PATCH | admin/jazebeh | game-catalog/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-admin.controller.ts | Native parity pending |
| JZ-ROUTE-101 | POST | admin/jazebeh | game-catalog/:id/suspend | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-admin.controller.ts | Native parity pending |
| JZ-ROUTE-102 | POST | admin/jazebeh | game-catalog/:id/unsuspend | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-admin.controller.ts | Native parity pending |
| JZ-ROUTE-103 | GET | admin/jazebeh | games | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-admin.controller.ts | Native parity pending |
| JZ-ROUTE-104 | GET | admin/jazebeh | games/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-admin.controller.ts | Native parity pending |
| JZ-ROUTE-105 | POST | admin/jazebeh | games/:id/suspend | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-admin.controller.ts | Native parity pending |
| JZ-ROUTE-106 | POST | admin/jazebeh | games/:id/unsuspend | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-admin.controller.ts | Native parity pending |
| JZ-ROUTE-107 | GET | brand/planets/:planetId/games |  | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-brand.controller.ts | Native parity pending |
| JZ-ROUTE-108 | POST | brand/planets/:planetId/games |  | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-brand.controller.ts | Native parity pending |
| JZ-ROUTE-109 | GET | brand/planets/:planetId/games | catalog | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-brand.controller.ts | Native parity pending |
| JZ-ROUTE-110 | GET | brand/planets/:planetId/games | meta | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-brand.controller.ts | Native parity pending |
| JZ-ROUTE-111 | GET | brand/planets/:planetId/games | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-brand.controller.ts | Native parity pending |
| JZ-ROUTE-112 | PATCH | brand/planets/:planetId/games | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-brand.controller.ts | Native parity pending |
| JZ-ROUTE-113 | POST | brand/planets/:planetId/games | :id/status | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-brand.controller.ts | Native parity pending |
| JZ-ROUTE-114 | GET | brand/planets/:planetId/games | :id/report | nest-backend/back-end/apps/api/src/modules/jazebeh/game/controllers/game-brand.controller.ts | Native parity pending |
| JZ-ROUTE-115 | GET | brand/moderation | reports | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/brand-moderation.controller.ts | Native parity pending |
| JZ-ROUTE-116 | PATCH | brand/moderation | reports/:reportId/dismiss | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/brand-moderation.controller.ts | Native parity pending |
| JZ-ROUTE-117 | GET | brand/moderation | groups | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/brand-moderation.controller.ts | Native parity pending |
| JZ-ROUTE-118 | DELETE | brand/moderation | groups/:groupId/messages/:seq | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/brand-moderation.controller.ts | Native parity pending |
| JZ-ROUTE-119 | POST | brand/moderation | groups/:groupId/suspend | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/brand-moderation.controller.ts | Native parity pending |
| JZ-ROUTE-120 | POST | brand/moderation | groups/:groupId/unsuspend | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/brand-moderation.controller.ts | Native parity pending |
| JZ-ROUTE-121 | POST | brand/moderation | groups/:groupId/members/:userId/remove | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/brand-moderation.controller.ts | Native parity pending |
| JZ-ROUTE-122 | GET | brand/moderation | suspended-members | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/brand-moderation.controller.ts | Native parity pending |
| JZ-ROUTE-123 | POST | brand/moderation | planets/:planetId/members/:userId/suspend | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/brand-moderation.controller.ts | Native parity pending |
| JZ-ROUTE-124 | POST | brand/moderation | planets/:planetId/members/:userId/unsuspend | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/brand-moderation.controller.ts | Native parity pending |
| JZ-ROUTE-125 | GET | client/groups/:id/messages |  | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/chat.controller.ts | Native parity pending |
| JZ-ROUTE-126 | POST | client/groups/:id/messages |  | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/chat.controller.ts | Native parity pending |
| JZ-ROUTE-127 | POST | client/groups/:id/messages | read | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/chat.controller.ts | Native parity pending |
| JZ-ROUTE-128 | DELETE | client/groups/:id/messages | :seq | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/chat.controller.ts | Native parity pending |
| JZ-ROUTE-129 | POST | client/groups/:id/messages | :seq/report | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/chat.controller.ts | Native parity pending |
| JZ-ROUTE-130 | GET | client/groups |  | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-131 | GET | client/groups | meta | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-132 | GET | client/groups | mine | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-133 | GET | client/groups | my-request | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-134 | GET | client/groups | name-availability | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-135 | POST | client/groups |  | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-136 | GET | client/groups | invites/:token | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-137 | POST | client/groups | join-by-invite | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-138 | GET | client/groups | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-139 | PATCH | client/groups | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-140 | DELETE | client/groups | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-141 | GET | client/groups | :id/members | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-142 | PATCH | client/groups | :id/members/:memberId | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-143 | GET | client/groups | :id/leadership-offer | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-144 | POST | client/groups | :id/leadership-offer/accept | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-145 | POST | client/groups | :id/leadership-offer/decline | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-146 | DELETE | client/groups | :id/leadership-offer | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-147 | DELETE | client/groups | :id/members/:memberId | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-148 | GET | client/groups | :id/activity | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-149 | GET | client/groups | :id/history | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-150 | GET | client/groups | :id/membership | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-151 | POST | client/groups | :id/membership | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-152 | DELETE | client/groups | :id/membership | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-153 | POST | client/groups | :id/leave | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-154 | GET | client/groups | :id/requests | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-155 | POST | client/groups | :id/requests | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-156 | PATCH | client/groups | :id/requests | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-157 | PATCH | client/groups | :id/requests/:requestId | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-158 | POST | client/groups | :id/invites | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-159 | DELETE | client/groups | :id/invites/:token | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group.controller.ts | Native parity pending |
| JZ-ROUTE-160 | GET | admin/jazebeh | groups | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-161 | GET | admin/jazebeh | groups/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-162 | GET | admin/jazebeh | groups/:id/members | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-163 | GET | admin/jazebeh | groups/:id/requests | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-164 | GET | admin/jazebeh | groups/:id/activity | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-165 | GET | admin/jazebeh | groups/:id/history | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-166 | POST | admin/jazebeh | groups/:id/suspend | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-167 | POST | admin/jazebeh | groups/:id/unsuspend | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-168 | POST | admin/jazebeh | groups/:id/disband | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-169 | GET | admin/jazebeh | groups/:id/messages | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-170 | DELETE | admin/jazebeh | groups/:id/messages/:seq | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-171 | GET | admin/jazebeh | reports | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-172 | PATCH | admin/jazebeh | reports/:id/dismiss | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-admin.controller.ts | Native parity pending |
| JZ-ROUTE-173 | GET | brand/planets/:planetId | groups | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-brand.controller.ts | Native parity pending |
| JZ-ROUTE-174 | GET | brand/planets/:planetId | groups/:groupId | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-brand.controller.ts | Native parity pending |
| JZ-ROUTE-175 | GET | brand/planets/:planetId | groups/:groupId/members | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-brand.controller.ts | Native parity pending |
| JZ-ROUTE-176 | GET | brand/planets/:planetId | groups/:groupId/requests | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-brand.controller.ts | Native parity pending |
| JZ-ROUTE-177 | GET | brand/planets/:planetId | groups/:groupId/activity | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-brand.controller.ts | Native parity pending |
| JZ-ROUTE-178 | GET | brand/planets/:planetId | groups/:groupId/history | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-brand.controller.ts | Native parity pending |
| JZ-ROUTE-179 | GET | brand/planets/:planetId | groups/:groupId/messages | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-brand.controller.ts | Native parity pending |
| JZ-ROUTE-180 | DELETE | brand/planets/:planetId | groups/:groupId/messages/:seq | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-brand.controller.ts | Native parity pending |
| JZ-ROUTE-181 | GET | brand/planets/:planetId | reports | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-brand.controller.ts | Native parity pending |
| JZ-ROUTE-182 | PATCH | brand/planets/:planetId | reports/:reportId/dismiss | nest-backend/back-end/apps/api/src/modules/jazebeh/group/controllers/group-brand.controller.ts | Native parity pending |
| JZ-ROUTE-183 | GET | brand/integrations | meta | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-184 | GET | brand/integrations | api-keys | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-185 | POST | brand/integrations | api-keys | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-186 | POST | brand/integrations | api-keys/:id/rotate | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-187 | POST | brand/integrations | api-keys/:id/revoke | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-188 | GET | brand/integrations | webhooks | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-189 | POST | brand/integrations | webhooks | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-190 | GET | brand/integrations | webhooks/deliveries | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-191 | POST | brand/integrations | webhooks/deliveries/:id/replay | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-192 | PATCH | brand/integrations | webhooks/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-193 | POST | brand/integrations | webhooks/:id/rotate-secret | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-194 | DELETE | brand/integrations | webhooks/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-195 | GET | brand/integrations | events/catalog | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-196 | GET | brand/integrations | events | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-197 | POST | brand/integrations | events/:id/reprocess | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/integration-brand.controller.ts | Native parity pending |
| JZ-ROUTE-198 | POST | public | events | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/public-api.controller.ts | Native parity pending |
| JZ-ROUTE-199 | GET | public | users/:ref/summary | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/public-api.controller.ts | Native parity pending |
| JZ-ROUTE-200 | POST | public | points/grant | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/public-api.controller.ts | Native parity pending |
| JZ-ROUTE-201 | GET | client/referrals | :token | nest-backend/back-end/apps/api/src/modules/jazebeh/integration/controllers/referral.controller.ts | Native parity pending |
| JZ-ROUTE-202 | GET | client/raffles |  | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery.controller.ts | Native parity pending |
| JZ-ROUTE-203 | GET | client/raffles | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery.controller.ts | Native parity pending |
| JZ-ROUTE-204 | GET | admin/jazebeh/lotteries |  | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-admin.controller.ts | Native parity pending |
| JZ-ROUTE-205 | GET | admin/jazebeh/lotteries | :id/audit | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-admin.controller.ts | Native parity pending |
| JZ-ROUTE-206 | GET | admin/jazebeh/lotteries | :id/tickets | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-admin.controller.ts | Native parity pending |
| JZ-ROUTE-207 | POST | admin/jazebeh/lotteries | :id/replay | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-admin.controller.ts | Native parity pending |
| JZ-ROUTE-208 | GET | admin/jazebeh/lotteries | :id/report | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-admin.controller.ts | Native parity pending |
| JZ-ROUTE-209 | GET | brand/planets/:planetId/lotteries |  | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-brand.controller.ts | Native parity pending |
| JZ-ROUTE-210 | GET | brand/planets/:planetId/lotteries | meta | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-brand.controller.ts | Native parity pending |
| JZ-ROUTE-211 | POST | brand/planets/:planetId/lotteries |  | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-brand.controller.ts | Native parity pending |
| JZ-ROUTE-212 | GET | brand/planets/:planetId/lotteries | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-brand.controller.ts | Native parity pending |
| JZ-ROUTE-213 | PATCH | brand/planets/:planetId/lotteries | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-brand.controller.ts | Native parity pending |
| JZ-ROUTE-214 | POST | brand/planets/:planetId/lotteries | :id/activate | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-brand.controller.ts | Native parity pending |
| JZ-ROUTE-215 | POST | brand/planets/:planetId/lotteries | :id/draw | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-brand.controller.ts | Native parity pending |
| JZ-ROUTE-216 | POST | brand/planets/:planetId/lotteries | :id/cancel | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-brand.controller.ts | Native parity pending |
| JZ-ROUTE-217 | GET | brand/planets/:planetId/lotteries | :id/pool | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-brand.controller.ts | Native parity pending |
| JZ-ROUTE-218 | GET | brand/planets/:planetId/lotteries | :id/winners | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-brand.controller.ts | Native parity pending |
| JZ-ROUTE-219 | PATCH | brand/planets/:planetId/lotteries | :id/winners/:winnerId/delivery | nest-backend/back-end/apps/api/src/modules/jazebeh/lottery/controllers/lottery-brand.controller.ts | Native parity pending |
| JZ-ROUTE-220 | GET | client/missions |  | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission.controller.ts | Native parity pending |
| JZ-ROUTE-221 | GET | client/missions | history | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission.controller.ts | Native parity pending |
| JZ-ROUTE-222 | GET | client/missions | history/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission.controller.ts | Native parity pending |
| JZ-ROUTE-223 | DELETE | client/missions | history/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission.controller.ts | Native parity pending |
| JZ-ROUTE-224 | GET | client/missions | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission.controller.ts | Native parity pending |
| JZ-ROUTE-225 | POST | client/missions | :id/media/start | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission.controller.ts | Native parity pending |
| JZ-ROUTE-226 | POST | client/missions | :id/complete | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission.controller.ts | Native parity pending |
| JZ-ROUTE-227 | POST | client/missions | :id/code | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission.controller.ts | Native parity pending |
| JZ-ROUTE-228 | POST | client/missions | :id/evidence | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission.controller.ts | Native parity pending |
| JZ-ROUTE-229 | DELETE | client/missions | :id/evidence | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission.controller.ts | Native parity pending |
| JZ-ROUTE-230 | GET | admin/jazebeh/missions |  | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-admin.controller.ts | Native parity pending |
| JZ-ROUTE-231 | GET | admin/jazebeh/missions | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-admin.controller.ts | Native parity pending |
| JZ-ROUTE-232 | POST | admin/jazebeh/missions | :id/suspend | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-admin.controller.ts | Native parity pending |
| JZ-ROUTE-233 | POST | admin/jazebeh/missions | :id/unsuspend | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-admin.controller.ts | Native parity pending |
| JZ-ROUTE-234 | GET | brand/planets/:planetId/missions |  | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-brand.controller.ts | Native parity pending |
| JZ-ROUTE-235 | POST | brand/planets/:planetId/missions |  | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-brand.controller.ts | Native parity pending |
| JZ-ROUTE-236 | GET | brand/planets/:planetId/missions | meta | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-brand.controller.ts | Native parity pending |
| JZ-ROUTE-237 | GET | brand/planets/:planetId/missions | review | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-brand.controller.ts | Native parity pending |
| JZ-ROUTE-238 | POST | brand/planets/:planetId/missions | review | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-brand.controller.ts | Native parity pending |
| JZ-ROUTE-239 | POST | brand/planets/:planetId/missions | review/:completionId | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-brand.controller.ts | Native parity pending |
| JZ-ROUTE-240 | GET | brand/planets/:planetId/missions | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-brand.controller.ts | Native parity pending |
| JZ-ROUTE-241 | PATCH | brand/planets/:planetId/missions | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-brand.controller.ts | Native parity pending |
| JZ-ROUTE-242 | POST | brand/planets/:planetId/missions | :id/status | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-brand.controller.ts | Native parity pending |
| JZ-ROUTE-243 | GET | brand/planets/:planetId/missions | :id/codes | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-brand.controller.ts | Native parity pending |
| JZ-ROUTE-244 | POST | brand/planets/:planetId/missions | :id/codes | nest-backend/back-end/apps/api/src/modules/jazebeh/mission/controllers/mission-brand.controller.ts | Native parity pending |
| JZ-ROUTE-245 | GET | admin/jazebeh/user-reports |  | nest-backend/back-end/apps/api/src/modules/jazebeh/moderation/controllers/user-report-admin.controller.ts | PARTIAL: `jazebe_operator_report_list`; SQL/HTTP contract unverified |
| JZ-ROUTE-246 | POST | admin/jazebeh/user-reports | :id/review | nest-backend/back-end/apps/api/src/modules/jazebeh/moderation/controllers/user-report-admin.controller.ts | PARTIAL: `jazebe_operator_report_review`; SQL/runtime unverified |
| JZ-ROUTE-247 | POST | admin/jazebeh/user-reports | :id/resolve | nest-backend/back-end/apps/api/src/modules/jazebeh/moderation/controllers/user-report-admin.controller.ts | PARTIAL: `jazebe_operator_report_resolve`; SQL/runtime unverified |
| JZ-ROUTE-248 | GET | brand/broadcasts |  | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-249 | GET | brand/broadcasts | planets | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-250 | POST | brand/broadcasts |  | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-251 | GET | brand/broadcasts | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-252 | GET | brand/broadcasts | :id/audience | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-253 | POST | brand/broadcasts | :id/confirm | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-254 | POST | brand/broadcasts | :id/cancel | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-255 | GET | admin/jazebeh | broadcasts | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-256 | POST | admin/jazebeh | broadcasts | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-257 | GET | admin/jazebeh | broadcasts/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-258 | GET | admin/jazebeh | broadcasts/:id/audience | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-259 | POST | admin/jazebeh | broadcasts/:id/confirm | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-260 | POST | admin/jazebeh | broadcasts/:id/cancel | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-261 | GET | admin/jazebeh | delivery-failures | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-262 | POST | admin/jazebeh | delivery-failures/:id/retry | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/broadcast.controller.ts | Native parity pending |
| JZ-ROUTE-263 | GET | notifications |  | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-264 | GET | notifications | unread-count | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-265 | POST | notifications | read-all | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-266 | PATCH | notifications | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-267 | DELETE | notifications | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-268 | DELETE | notifications |  | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-269 | GET | settings | preferences | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-270 | PATCH | settings | preferences | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-271 | GET | settings | planet-order | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-272 | PATCH | settings | planet-order | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-273 | PUT | settings | planet-order | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-274 | GET | settings | invite-code | nest-backend/back-end/apps/api/src/modules/jazebeh/notification/controllers/notification.controller.ts | Native parity pending |
| JZ-ROUTE-275 | GET | admin/jazebeh/operations | queues | nest-backend/back-end/apps/api/src/modules/jazebeh/operations/controllers/operations.controller.ts | Native parity pending |
| JZ-ROUTE-276 | POST | admin/jazebeh/operations | queues/:name/retry-failed | nest-backend/back-end/apps/api/src/modules/jazebeh/operations/controllers/operations.controller.ts | Native parity pending |
| JZ-ROUTE-277 | POST | admin/jazebeh/operations | jobs/run | nest-backend/back-end/apps/api/src/modules/jazebeh/operations/controllers/operations.controller.ts | Native parity pending |
| JZ-ROUTE-278 | GET | admin/jazebeh/operations | dead-letters | nest-backend/back-end/apps/api/src/modules/jazebeh/operations/controllers/operations.controller.ts | Native parity pending |
| JZ-ROUTE-279 | DELETE | admin/jazebeh/operations | dead-letters/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/operations/controllers/operations.controller.ts | Native parity pending |
| JZ-ROUTE-280 | GET | admin/jazebeh/operations | failed-deliveries | nest-backend/back-end/apps/api/src/modules/jazebeh/operations/controllers/operations.controller.ts | Native parity pending |
| JZ-ROUTE-281 | POST | admin/jazebeh/operations | failed-deliveries/:kind/:id/replay | nest-backend/back-end/apps/api/src/modules/jazebeh/operations/controllers/operations.controller.ts | Native parity pending |
| JZ-ROUTE-282 | GET | admin/jazebeh/operations | reconciliation | nest-backend/back-end/apps/api/src/modules/jazebeh/operations/controllers/operations.controller.ts | Native parity pending |
| JZ-ROUTE-283 | GET | admin/jazebeh/operations | exports | nest-backend/back-end/apps/api/src/modules/jazebeh/operations/controllers/operations.controller.ts | Native parity pending |
| JZ-ROUTE-284 | GET | admin/jazebeh | operator/me | nest-backend/back-end/apps/api/src/modules/jazebeh/operator/controllers/operator.controller.ts | Native parity pending |
| JZ-ROUTE-285 | GET | admin/jazebeh | operators | nest-backend/back-end/apps/api/src/modules/jazebeh/operator/controllers/operator.controller.ts | Native parity pending |
| JZ-ROUTE-286 | PATCH | admin/jazebeh | operators/:id/level | nest-backend/back-end/apps/api/src/modules/jazebeh/operator/controllers/operator.controller.ts | Native parity pending |
| JZ-ROUTE-287 | GET | admin/jazebeh | audit-log | nest-backend/back-end/apps/api/src/modules/jazebeh/operator/controllers/operator.controller.ts | Native parity pending |
| JZ-ROUTE-288 | GET | admin/jazebeh | audit-log/exports | nest-backend/back-end/apps/api/src/modules/jazebeh/operator/controllers/operator.controller.ts | Native parity pending |
| JZ-ROUTE-289 | POST | admin/jazebeh | audit-log/exports | nest-backend/back-end/apps/api/src/modules/jazebeh/operator/controllers/operator.controller.ts | Native parity pending |
| JZ-ROUTE-290 | GET | admin/jazebeh | audit-log/exports/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/operator/controllers/operator.controller.ts | Native parity pending |
| JZ-ROUTE-291 | GET | admin/jazebeh/audit-log/exports/download | :token | nest-backend/back-end/apps/api/src/modules/jazebeh/operator/controllers/operator.controller.ts | Native parity pending |
| JZ-ROUTE-292 | GET | client/brand-domain |  | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/brand-domain.controller.ts | Native parity pending |
| JZ-ROUTE-293 | GET | client | planet | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet.controller.ts | Native parity pending |
| JZ-ROUTE-294 | GET | client | brands | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet.controller.ts | Native parity pending |
| JZ-ROUTE-295 | GET | client | brands/:planet/missions/preview | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet.controller.ts | Native parity pending |
| JZ-ROUTE-296 | GET | client | brands/:planet/groups/preview | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet.controller.ts | Native parity pending |
| JZ-ROUTE-297 | GET | client | brands/:planet/raffle/preview | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet.controller.ts | Native parity pending |
| JZ-ROUTE-298 | GET | client | planets/discover | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet.controller.ts | Native parity pending |
| JZ-ROUTE-299 | POST | client | planets/:planet/membership | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet.controller.ts | Native parity pending |
| JZ-ROUTE-300 | GET | admin/jazebeh | brands | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-301 | POST | admin/jazebeh | brands | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-302 | GET | admin/jazebeh | brands/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-303 | PATCH | admin/jazebeh | brands/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-304 | POST | admin/jazebeh | brands/:id/contract | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-305 | POST | admin/jazebeh | brands/:id/admins | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-306 | PATCH | admin/jazebeh | brands/:id/admins/:adminId/role | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-307 | POST | admin/jazebeh | brands/:id/admins/:adminId/deactivate | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-308 | POST | admin/jazebeh | brands/:id/admins/:adminId/reactivate | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-309 | GET | admin/jazebeh | planets | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-310 | GET | admin/jazebeh | planets/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-311 | POST | admin/jazebeh | planets/:id/approve | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-312 | POST | admin/jazebeh | planets/:id/reject | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-313 | POST | admin/jazebeh | planets/:id/suspend | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-314 | POST | admin/jazebeh | planets/:id/unsuspend | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-315 | POST | admin/jazebeh | planets/:id/archive | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-316 | GET | admin/jazebeh | planets/:id/domain | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-317 | GET | admin/jazebeh | planet-domains/check | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-318 | PUT | admin/jazebeh | planets/:id/domain | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-319 | DELETE | admin/jazebeh | planets/:id/domain | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-320 | GET | admin/jazebeh | planet-categories | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-321 | POST | admin/jazebeh | planet-categories | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-322 | PATCH | admin/jazebeh | planet-categories/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-323 | DELETE | admin/jazebeh | planet-categories/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native parity pending |
| JZ-ROUTE-324 | GET | admin/jazebeh | platform-policy | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native RPC implemented; PostgreSQL/Nakama unverified |
| JZ-ROUTE-325 | PATCH | admin/jazebeh | platform-policy | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-admin.controller.ts | Native RPC implemented; PostgreSQL/Nakama unverified |
| JZ-ROUTE-326 | GET | brand | planets | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native parity pending |
| JZ-ROUTE-327 | POST | brand | planets | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native parity pending |
| JZ-ROUTE-328 | GET | brand | planets/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native parity pending |
| JZ-ROUTE-329 | PATCH | brand | planets/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native parity pending |
| JZ-ROUTE-330 | GET | brand | planets/:id/publish-check | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native parity pending |
| JZ-ROUTE-331 | POST | brand | planets/:id/publish | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native parity pending |
| JZ-ROUTE-332 | POST | brand | planets/:id/withdraw | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native parity pending |
| JZ-ROUTE-333 | POST | brand | planets/:id/suspend | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native parity pending |
| JZ-ROUTE-334 | GET | brand | planets/:id/settings | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native parity pending |
| JZ-ROUTE-335 | PUT | brand | planets/:id/settings | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native parity pending |
| JZ-ROUTE-336 | GET | brand | planet-categories | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native parity pending |
| JZ-ROUTE-337 | GET | brand | platform-policy | nest-backend/back-end/apps/api/src/modules/jazebeh/planet/controllers/planet-brand.controller.ts | Native RPC implemented; PostgreSQL/Nakama unverified |
| JZ-ROUTE-338 | GET | admin/jazebeh/medal-library |  | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/medal-library.controller.ts | Native RPC implemented; PostgreSQL/Nakama unverified |
| JZ-ROUTE-339 | GET | admin/jazebeh/medal-library | meta | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/medal-library.controller.ts | Native RPC implemented; PostgreSQL/Nakama unverified |
| JZ-ROUTE-340 | POST | admin/jazebeh/medal-library |  | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/medal-library.controller.ts | Native RPC implemented; PostgreSQL/Nakama unverified |
| JZ-ROUTE-341 | GET | admin/jazebeh/medal-library | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/medal-library.controller.ts | Native RPC implemented; PostgreSQL/Nakama unverified |
| JZ-ROUTE-342 | PATCH | admin/jazebeh/medal-library | :id | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/medal-library.controller.ts | Native RPC implemented; PostgreSQL/Nakama unverified |
| JZ-ROUTE-343 | POST | admin/jazebeh/medal-library | :id/activate | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/medal-library.controller.ts | Native RPC implemented; PostgreSQL/Nakama unverified |
| JZ-ROUTE-344 | POST | admin/jazebeh/medal-library | :id/deactivate | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/medal-library.controller.ts | Native RPC implemented; PostgreSQL/Nakama unverified |
| JZ-ROUTE-345 | GET | admin/jazebeh/medal-library | :id/history | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/medal-library.controller.ts | Native RPC implemented; PostgreSQL/Nakama unverified |
| JZ-ROUTE-346 | GET | brand/planets/:planetId/progression | xp | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/progression-brand.controller.ts | Native parity pending |
| JZ-ROUTE-347 | PUT | brand/planets/:planetId/progression | levels | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/progression-brand.controller.ts | Native parity pending |
| JZ-ROUTE-348 | PUT | brand/planets/:planetId/progression | xp | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/progression-brand.controller.ts | Native parity pending |
| JZ-ROUTE-349 | GET | brand/planets/:planetId/progression | medals | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/progression-brand.controller.ts | Native RPC implemented; SQL/runtime and HTTP contract unverified |
| JZ-ROUTE-350 | GET | brand/planets/:planetId/progression | medals/meta | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/progression-brand.controller.ts | Native RPC implemented; SQL/runtime and HTTP contract unverified |
| JZ-ROUTE-351 | GET | brand/planets/:planetId/progression | medals/library | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/progression-brand.controller.ts | Native RPC implemented; SQL/runtime and HTTP contract unverified |
| JZ-ROUTE-352 | POST | brand/planets/:planetId/progression | medals | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/progression-brand.controller.ts | Native RPC implemented; SQL/runtime and HTTP contract unverified |
| JZ-ROUTE-353 | PATCH | brand/planets/:planetId/progression | medals/:id | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/progression-brand.controller.ts | Native RPC implemented; SQL/runtime and HTTP contract unverified |
| JZ-ROUTE-354 | POST | brand/planets/:planetId/progression | medals/:id/status | nest-backend/back-end/apps/api/src/modules/jazebeh/progression/controllers/progression-brand.controller.ts | Native RPC implemented; SQL/runtime and HTTP contract unverified |
| JZ-ROUTE-355 | GET | platform/products/:productId/members |  | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-356 | POST | platform/products/:productId/members | invite | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-357 | PATCH | platform/products/:productId/members | :memberUserId | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-358 | DELETE | platform/products/:productId/members | :memberUserId | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-359 | POST | platform/invitations | accept | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-360 | POST | platform/workspaces |  | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-361 | GET | platform/workspaces |  | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-362 | GET | platform/workspaces | :id | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-363 | DELETE | platform/workspaces | :id | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-364 | POST | platform/workspaces | :id/products/:productId | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-365 | DELETE | platform/workspaces | products/:productId | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-366 | POST | platform/transfers |  | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-367 | POST | platform/transfers | :id/verify | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-368 | GET | platform/admin | me | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-369 | GET | platform/admin | admins | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-370 | POST | platform/admin | admins | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-371 | PATCH | platform/admin | admins/:id | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-org.controller.ts | Native parity pending |
| JZ-ROUTE-372 | GET | platform/admin | system/overview | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-373 | GET | platform/admin | system/products | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-374 | GET | platform/admin | system/users | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-375 | GET | platform/admin | system/users/:id | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-376 | PATCH | platform/admin | system/users/:id/grants | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-377 | GET | platform/admin | system/workspaces | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-378 | POST | platform/admin | system/products/:id/approve | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-379 | POST | platform/admin | system/products/:id/reject | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-380 | PATCH | platform/admin | system/products/:id/active | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-381 | PATCH | platform/admin | system/workspaces/:id/active | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-382 | POST | platform/admin | system/workspaces/:id/approve | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-383 | POST | platform/admin | system/workspaces/:id/reject | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-384 | GET | platform/admin | settings | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-385 | PATCH | platform/admin | settings | nest-backend/back-end/apps/api/src/modules/platform/controllers/platform-system.controller.ts | Native parity pending |
| JZ-ROUTE-386 | POST | platform/products |  | nest-backend/back-end/apps/api/src/modules/platform/controllers/product.controller.ts | Native parity pending |
| JZ-ROUTE-387 | GET | platform/products |  | nest-backend/back-end/apps/api/src/modules/platform/controllers/product.controller.ts | Native parity pending |
| JZ-ROUTE-388 | GET | platform/products | :id | nest-backend/back-end/apps/api/src/modules/platform/controllers/product.controller.ts | Native parity pending |
| JZ-ROUTE-389 | PATCH | platform/products | :id | nest-backend/back-end/apps/api/src/modules/platform/controllers/product.controller.ts | Native parity pending |

Total: 389 route decorators. Guards, DTO validation, service calls, and tenant policies must be traced in source before any adapter is declared complete.
