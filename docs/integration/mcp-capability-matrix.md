# MCP capability matrix

## Game and referral reads (2026-10-22)

**112 tools**, one resource and one prompt are registered. Six read-only game tools expose caller-scoped active games and brand-scoped list, get, catalog, metadata and report through native authorization checks. `jazebe_referral_stats` exposes only the authenticated caller's own invitation counts. New `jazebe_profile_overview` and `jazebe_profile_public` delegate to owner and planet-scoped native read guards; MCP mock-upstream discovery/invocation passes. Game entry, rewards, referral acceptance, profile mutations, operator game changes and workers remain native-only. Remote OAuth and live Nakama interoperability remain NOT_RUN.


## Brand medal inspection (2026-10-22)

**103 tools**, one resource and one prompt are registered. Five new read-only tools are `jazebe_medal_wall`, `jazebe_brand_medals`, `jazebe_brand_medal_meta`, `jazebe_brand_medal_library`, and `jazebe_brand_medal_backfill_status`. Each delegates to a Nakama RPC requiring the caller's active brand economy role. Medal add/update/status and operator library mutation are native-only until remote MCP OAuth scopes are implemented. MCP SDK discovery and module tests/vet pass; a live Nakama/PostgreSQL call is NOT_RUN.

## Brand policy read (2026-10-22)

**98 tools**, one resource and one prompt are registered. `jazebe_brand_platform_policy` calls the native brand policy read RPC and requires an active brand administrator for the provided brand UUID. Platform policy mutation and operator medal library operations remain native-only until remote MCP OAuth scopes and privileged delegation are implemented. SDK discovery and mock-upstream tests pass; live Nakama/PostgreSQL interoperability is NOT_RUN.

## Medal library exposure decision (2026-10-22)

The eight new operator medal-library RPCs are native-only. They require operator or super-operator identity and cannot safely be offered over the current loopback MCP session adapter without remote OAuth scope delegation. Existing MCP count remains 97 tools, one resource and one prompt. No MCP medal acquisition tool exists because awards must arise from server-verified activity.

## Progression side effects (2026-10-22)

Tool count remains **97**, with one resource and one prompt. Existing caller-scoped shop purchase and group-join MCP tools now trigger the native XP side effect through their RPC handlers. No separate XP grant tool is exposed because XP is earned only through authorized domain activity. Medal evaluation and scoped remote OAuth are missing; remote HTTP stays loopback-only. No live Nakama/PostgreSQL invocation was run.

## Current shop read additions (2026-10-22)

**97 tools**, one resource and one prompt are registered. New brand-scoped read tools cover product list/detail, eligible lotteries, coupon inventory counts and aggregate sales. The underlying native RPC enforces active brand administrator access. Raw coupon code listing, product writes and operator moderation remain native-only because scoped remote OAuth is not implemented. Client store tools cover all three instant-delivery product types. MCP SDK mock-upstream discovery and Streamable HTTP tests pass; live Nakama/PostgreSQL interoperability does not. Remote HTTP remains loopback-only.

## Current instant-delivery shop tools (2026-10-22)

**92 tools** are registered, plus one resource and one prompt. Caller-scoped `jazebe_shop_purchase`, `jazebe_shop_coupon_purchase` and `jazebe_shop_digital_purchase` use the real native purchase RPCs; the catalog list/detail now covers ticket, coupon and digital products. Buyer archives continue to return only the caller's orders. Brand coupon/digital creation and raw coupon generation remain native-only because scoped remote OAuth is not implemented. MCP SDK discovery and mock-upstream protocol tests pass; live Nakama/PostgreSQL interoperability is pending. Remote HTTP remains loopback-only.

## Current shop additions (2026-10-22)

Current registration count: **89 tools**, one resource, one prompt. New caller-scoped tools `jazebe_shop_purchases`, `jazebe_shop_products`, and `jazebe_shop_product` call the new native RPCs. Archive retrieval is restricted to the Nakama session owner; catalog reads require an active planet/brand and deny suspended members. Catalog output currently covers ticket products only. Pure native tests, MCP SDK discovery with mocked upstream and Streamable HTTP tests pass; live Nakama/PostgreSQL access and remote OAuth authorization remain unverified. Remote HTTP stays loopback only.

Remote HTTP is loopback only and accepts Nakama sessions; compliant OAuth and scopes for remote production use are pending. SDK client discovery and invocation tests pass over in-memory and Streamable HTTP transports against a mock Nakama upstream; live service interoperability remains unverified.

Current registered tool count: **86**. Brand lottery metadata, winner reads, cancellation and other source-backed operations, active/finished raffle lists and client detail, and two shop ticket tools are exposed. Ten group management tools and five member chat tools remain. Draw, delivery, replay, staff moderation and realtime subscribe remain native only. Native RPC authorization remains authoritative. Remote OAuth and scope delegation are still absent, so remote exposure remains loopback restricted.

| Capability | Native service | MCP primitive | Authorization | Status |
|---|---|---|---|---|
| Brand lottery draft, update, activation and reads | Six `jazebe_lottery_*` RPCs | Six matching lottery tools | Active brand administrator and Nakama caller session | SDK discovery and mocked read invocation pass; live DB and OAuth pending; no ticket minting tool |
| Brand lottery cancellation | `jazebe_lottery_brand_cancel` | Matching cancellation tool | Active brand administrator, reasoned cancellation | Pure guards pass; live refunds/transfers unverified |
| Active raffle client cards | `jazebe_lottery_client_active` | Matching client tool | Authenticated caller, active planet and brand, suspended member denial | SDK array invocation with mock Nakama passes; live DB pending |
| Shop ticket product and purchase | `jazebe_shop_ticket_create`, `jazebe_shop_ticket_purchase` | Two corresponding tools | Brand administrator for creation; active group member with points for purchase | Pure idempotency and product validation tests pass; SQL and live upstream pending |
| Nakama account | Nakama `/v2/account` | `nakama_get_account` tool; `nakama://me/account` resource; `inspect_my_account` prompt | Caller session | SDK test passed; live Nakama pending |
| Brand creation | `jazebe_brand_create` RPC | `jazebe_create_brand` tool | Caller session and `JAZEBE_OPERATOR_USER_IDS` allowlist | Partial business parity; deny-by-default unit test |
| Brand team list | `jazebe_brand_team_list` RPC | `jazebe_list_brand_team` tool | Active main brand administrator | Role unit test and MCP discovery pass; DB pending |
| Brand list | `jazebe_brand_list` RPC | `jazebe_list_brands` tool | Caller session | Mock-upstream invocation tested; live DB pending |
| Planet creation | `jazebe_planet_create` RPC | `jazebe_create_planet` tool | Caller session and brand ownership | Setup draft and default settings implemented; DB test pending |
| Brand planet drafts | `jazebe_brand_planets` RPC | `jazebe_list_brand_planets` tool | Caller session and brand ownership | Mock-upstream invocation tested; live DB pending |
| Brand planet detail | `jazebe_brand_planet_get` RPC | `jazebe_get_brand_planet` tool | Caller session and brand ownership | Mock-upstream invocation tested; live DB pending |
| Brand planet update | `jazebe_brand_planet_update` RPC | `jazebe_update_brand_planet` tool | Caller session and brand ownership | Mock-upstream invocation tested; live DB pending |
| Planet publish readiness | `jazebe_planet_publish_check` RPC | `jazebe_check_planet_publish` tool | Caller session and brand ownership | Pure rule and mock-upstream invocation tests pass; live DB pending |
| Planet publish request | `jazebe_planet_publish_request` RPC | `jazebe_request_planet_publish` tool | Caller session and brand ownership | Source requirement check; media mission path exists, DB pending |
| Planet publish withdrawal | `jazebe_planet_publish_withdraw` RPC | `jazebe_withdraw_planet_publish` tool | Caller session and brand ownership | Native code compiles; DB pending |
| Media mission draft | `jazebe_media_mission_draft_create` RPC | `jazebe_create_media_mission_draft` tool | Caller session and brand ownership | Draft validation and MCP discovery tests pass; DB pending |
| Media mission activation | `jazebe_media_mission_activate` RPC | `jazebe_activate_media_mission` tool | Active brand administrator | Native build and MCP discovery pass; DB pending |
| Media mission start and completion | Native media start/complete RPCs | `jazebe_start_media_mission`, `jazebe_complete_media_mission` tools | Caller session, active planet and mission | Server watch rule tested in unit; mock MCP start invocation passes; DB pending |
| Media review | `jazebe_media_mission_review` RPC | `jazebe_review_media_mission` tool | Active brand administrator | Review validation test passes; DB pending |
| Point balance | `jazebe_point_balance` RPC | `jazebe_get_point_balance` tool | Caller session, own wallet only | MCP discovery pass; DB pending |
| Media catalog/history/withdrawal | Native media list/history/withdraw RPCs | Three corresponding media MCP tools | Caller session and own completion records | MCP discovery and native build pass; DB pending |
| Code-entry mission draft/activation | Native code draft/activate RPCs | `jazebe_create_code_mission_draft`, `jazebe_activate_code_mission` tools | Active brand administrator | Native guards and MCP discovery pass; DB pending |
| Generated code stock | `jazebe_code_mission_generate` RPC | `jazebe_generate_mission_codes` tool | Active brand administrator; returned codes sensitive | Crypto format unit test and MCP discovery pass; DB pending |
| Code redemption | `jazebe_code_mission_redeem` RPC | `jazebe_redeem_mission_code` tool | Caller session, active mission, valid unused code | Native build and MCP discovery pass; DB concurrency pending |
| Code stock and actionable catalog | Native code stock and combined mission list RPCs | `jazebe_list_mission_codes`, `jazebe_list_missions` tools | Brand administrator for stock; caller session for catalog | MCP discovery passes; live DB pending |
| Mission status | `jazebe_mission_status` RPC | `jazebe_set_mission_status` tool | Active brand administrator and last solo mission guard | Native build and MCP discovery pass; DB pending |
| Wallet summary and history | Native wallet summary and transaction RPCs | `jazebe_wallet_summary`, `jazebe_wallet_transactions` tools | Caller session, own wallet only | MCP discovery passes; DB and localization verification pending |
| Content evidence mission | Native content draft, activation and submission RPCs | Three content mission tools accepting text/code/image/video | Brand administrator for draft/activation; caller session for submission | Pure guards and MCP discovery pass; live media/DB pending |
| Mission media upload and withdrawal | Native upload begin/delete RPCs and ticketed HTTP byte transfer | `jazebe_media_upload_begin`, `jazebe_media_upload_delete` tools | Nakama session; native owner/planet/purpose checks; upload ticket for HTTP chunks | SDK discovery and pure upload guards pass; live HTTP/DB pending |
| Group creation, policy and name reservation | Three native group RPCs | `jazebe_group_create`, `jazebe_group_meta`, `jazebe_group_name_availability` tools | Nakama session; native planet policy and membership constraints | Pure guards and SDK discovery pass; activity side effects and live DB pending |
| Group join, request, cancellation and leave | Four native group RPCs | `jazebe_group_join`, `jazebe_group_membership_state`, `jazebe_group_cancel_request`, `jazebe_group_leave` tools | Nakama session; caller-scoped membership and request rows | Pure guards and SDK discovery pass; live DB pending |
| Group manager request queue and decisions | Native queue, single and bulk decision RPCs | Three manager-scoped tools | Nakama session plus active creator/admin membership in the specific group | Pure validation/formatting and SDK discovery pass; live DB pending |
| Group invitation links | Native mint, preview, join and revoke RPCs | Four invite tools | Nakama session; native group membership, token status, capacity and ownership checks | Pure token guards and SDK discovery pass; invite-mission links and live DB pending |
| Generic review/history/withdrawal | Native aliases of supported mission handlers | Three generic mission tools | Brand administrator for review; own records for history/withdrawal | SDK discovery passes; live DB pending |
| Mission detail and completion detail | Native `jazebe_mission_detail` and `jazebe_mission_history_item` RPCs | Two caller-scoped MCP tools | Authenticated session; own completion only | SDK discovery passes; live DB pending |
| Mission display conditions | Native mission drafts, reads and action handlers | Optional display conditions on three mission-draft MCP tools | Brand administrator for drafts; caller-specific visibility on reads/actions | Pure Tehran/level/prerequisite tests pass; live DB pending |
| Periodic media/code/text mission | Native drafts, start/redemption/submission and completion | Existing mission MCP tools now accept repeat pattern and period | Brand administrator for draft; member session for action | Pure Tehran cycle tests and SDK discovery pass; live DB pending |
| Operator mission moderation | Four native operator RPCs | No MCP exposure | `JAZEBE_OPERATOR_USER_IDS` allowlist | Deny-by-default unit test; live DB pending |
| Active planet list | `jazebe_planet_list` RPC | `jazebe_list_planets` tool | Caller session | Partial business parity; discovery tested |
| Planet button join | `jazebe_planet_join` RPC | `jazebe_join_planet` tool | Caller session | Partial business parity; discovery tested |
| Caller memberships | `jazebe_planet_memberships` RPC | `jazebe_list_my_planets` tool | Caller session | Mock-upstream invocation tested; live DB pending |

All other Jazebe capabilities remain unexposed until their underlying services and access rules exist. Operator planet transitions are not exposed through MCP. Brand administration tools enforce the native role checks; remote HTTP remains loopback only until OAuth authorization is implemented.
