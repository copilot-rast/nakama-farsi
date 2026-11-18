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

## Add a changelog entry

Edit `general/handoff.json` (general) or `project/handoff.json` (fork) and add an entry at the top of `changelog`:

```json
{"date": "2026-11-18", "title": "متن فارسی", "titleEn": "English text", "section": "05-unity-sdk.md"}
```

`date` is YYYY-MM-DD, `section` is optional and must be a real source path. The file and the console panel merge both lists, newest first, so a merge from nakama-farsi brings new general entries into every fork.

## Settings (Server rules › Front-end handoff)

`handoff.enabled` (default on) and `handoff.audience` (`all` default, or `admins`). Stored in `base_server_rules`.

## Progress

- [x] Engine, endpoints, ACL, admin header, general content (nakama-farsi)
- [x] Console header button (plane topbar `data-pl-handoff`) + side panel (`console/ui/economy/src/handoff.js`), Server rules section, e2e `79-handoff`
- [ ] Forks: rakhtkan, combatrace, pp project parts; finteen, viralvanguard, soltan general only
