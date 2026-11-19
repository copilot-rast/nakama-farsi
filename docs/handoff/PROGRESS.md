# Front-end handoff: progress and how it works

## Layout

- `docs/handoff/general/` — the general part (nakama-farsi only). Forks never edit it.
- `docs/handoff/project/` — the fork's own part. Only forks have it; nakama-farsi never creates it, so merges do not conflict.
- `docs/handoff/assets/` — style.css, script.js, Vazirmatn woff2 fonts (embedded as base64 in the file).
- `docs/handoff/content.go` — `//go:embed *`; the base plugin embeds everything here.
- `internal/handoffkit/` — loader, Markdown renderer, HTML renderer, settings.
- `integrations/base/handoff.go` — `GET /v2/console/base/handoff` (meta + changelog) and `GET /v2/console/base/handoff/download` (the single HTML file).

## Add a section

Drop a `.md` (or `.html` body fragment) into `general/` or `project/`. Sections are sorted by path; subfolders are fine. Optional front matter:

```
---
title: عنوان بخش
titleEn: Section title
summary: یک خط زیر عنوان
---
```

Without `title` the first `# ` heading is the title. `{{CONSOLE_URL}}` in a section becomes the console address at download time (e.g. `{{CONSOLE_URL}}/#/system/sdk`). Links to other source files (`../phase1/README.md#anchor`) become in-file anchors.

Connection placeholders, filled from the running server at download time (sources keep only the placeholders, never a real key):

| Placeholder | Value |
| --- | --- |
| `{{SERVER_KEY}}` | `socket.server_key` of the running server; "ask the admin" when `handoff.showServerKey` is off |
| `{{SERVER_HOST}}`, `{{SERVER_PORT}}`, `{{SERVER_SSL}}` | `HANDOFF_PUBLIC_HOST` / `_PORT` / `_SSL`; without them the host of `ECONOMY_PUBLIC_BASE_URL`; port 443 with SSL, 80 without; SSL defaults to true |
| `{{SERVER_URL}}` | `http(s)://host[:port]` from the same values |

Inside code (`` `x` `` or code blocks) an unknown value becomes an ASCII marker (`ASK_THE_SERVER_ADMIN`, `API_HOST_FROM_ADMIN`); in text it becomes a Persian "ask the admin" note. `[text](#connection)` and `[text](#changes)` link to the connection box and the change list. A test fails if a source contains a 64-hex key.

## Add a changelog entry

Edit `general/handoff.json` (general) or `project/handoff.json` (fork) and add an entry at the top of `changelog`:

```json
{"date": "2026-11-18", "title": "متن فارسی", "titleEn": "English text", "section": "05-unity-sdk.md"}
```

`date` is YYYY-MM-DD, `section` is optional and must be a real source path. The file and the console panel merge both lists, newest first, so a merge from nakama-farsi brings new general entries into every fork.

## Settings (Server rules › Front-end handoff)

`handoff.enabled` (default on), `handoff.audience` (`all` default, or `admins`) and `handoff.showServerKey` (default on). Stored in `base_server_rules`. The console SDK page reads the same values from `GET /v2/console/base/handoff/connection` (any console login; the key only when `showServerKey` is on and the audience lets the user in).

## Progress

- [x] Engine, endpoints, ACL, admin header, general content (nakama-farsi)
- [x] Console header button (plane topbar `data-pl-handoff`) + side panel (`console/ui/economy/src/handoff.js`), Server rules section, e2e `79-handoff`
- [x] Forks merged at 723b1810 and recorded: rakhtkan, combatrace, pp (with project parts), finteen, viralvanguard, soltan (general only)
- [x] e2e 79 + shell/nav specs green on nakama-farsi, Rakhtkan and PP local servers

## Fork parts (sources moved in)

- rakhtkan: `project/01-phase1` … `06-cup`, `07-api-docs` from `rakhtkan/handoff/*-md` (README → 00-overview, REQUIRED-INFO → 10-required-info with the server key replaced by a link to the console SDK page, cup examples → `06-cup/05-examples.md`). Left out: dev-test-data (test logins), front-test-data-request, DEPLOY-*, rakhtkan.env, the duplicate levels-xp file.
- combatrace: `project/01-unity.md`, `02-heatmap-admin.md`. Left out: DEPLOY-FA (ops), gap-review-fa (internal, stale).
- pp: `project/00-overview.html` … `10-errors.html` from `pp/handoff/pp-unity-guide-work/src` (server key removed, links rewritten). Left out: the older Markdown guides (superseded), migration note, legacy parity audit, analytics design note.
