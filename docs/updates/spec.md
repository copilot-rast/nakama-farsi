# Updates («به‌روزرسانی‌ها»)

Read-only watcher for the parent repo (owner's design, 2026-11-13). The console shows this server's version and what changed in the parent since the commit this build is based on. Nothing is merged, pulled or run from the console.

## Repo chain

```
heroiclabs/nakama  ->  copilot-rast/nakama-farsi (main, stage, dev)  ->  per-game forks (GitHub or GitLab)
```

Each repo watches only its own parent:

| Repo | «ریپوی بالایی» setting | `build/upstream.json` records |
|---|---|---|
| nakama-farsi | `heroiclabs/nakama`, branch `master` (default) | the newest heroiclabs commit merged into HEAD |
| a game fork | `copilot-rast/nakama-farsi`, branch `main`, `stage` or `dev` | the newest nakama-farsi commit merged into HEAD |

A fork never needs heroiclabs. When a fork merges nakama-farsi, `build/upstream.json` conflicts (nakama-farsi's file points at heroiclabs): resolve it by running `build/upstream-sync.sh record nakama-farsi <branch>` (the remote name the fork uses), which rewrites the file for the fork's own parent.

## Base commit and build info

The base commit is tracked in Git, so a build needs no network:

- `build/upstream.json`: `provider`, `repo`, `remote`, `branch`, `commit` (base), `date`, `subject`, `recordedAt`. Current value: heroiclabs/nakama `master` at `37618ad4b927a9dd0671599ffc7387a7386f70fa` (2026-11-12, «Fix matchmaker remainder eviction…», the last upstream merge `00f58d08`).
- `build/upstream-sync.sh record [remote] [branch]`: run after merging the parent; finds `git merge-base HEAD <remote>/<branch>` and writes the file (defaults: the remote and branch already in the file, else `upstream master`). Commit the file. `build/upstream-sync.sh show` prints it and how many parent commits are not merged yet (as of the last fetch).
- `build/buildinfo.sh` prints Go `-ldflags` that set `internal/buildinfo` (`Commit`, `Branch`, `Date`, `Dirty`, `OwnAhead` = own commits on top of the base, `Upstream` = base64 of `build/upstream.json`) and `main.commitID`. Pass the same flags to the server and to every plugin:

```sh
LD="$(build/buildinfo.sh)"
go build -o nakama -ldflags "$LD" .
go build -buildmode=plugin -ldflags "$LD" -o modules/base.so ./integrations/base
```

The base plugin reads `internal/buildinfo`; the server binary does not import it, so a plugin built without the flags reports «this build has no base commit» (`no_base`).

- `build/Dockerfile.local` runs the script once into `/go/buildinfo.ldflags` and uses it for the server and all plugins. With `.git` in the build context (Hamravesh and Dokploy clone the repo, `.dockerignore` keeps `.git`) the script reads Git; `Dirty` ignores files the context left out. Without `.git`, or to override, pass build args: `GIT_COMMIT`, `GIT_BRANCH`, `GIT_DIRTY`, `GIT_OWN_AHEAD`, `BUILD_DATE` (each one overrides Git only for its own field). `build/upstream.json` is copied with the source, so the base is always known.

## Settings

Project settings section «ریپوی بالایی» (`#/project-settings/updates-source`, ACL `SETTINGS`), table `base_updates_settings` (one row; no row = defaults):

| Field | Rule |
|---|---|
| `enabled` | on by default; off stops the worker, the badge and «بررسی دوباره» (`UPDATES_OFF`) |
| `provider` | `github` only (`UNSUPPORTED_PROVIDER`); a GitLab parent is not built |
| `repo` | `owner/name` or a github.com URL, stored as `owner/name` (`INVALID_REPO`); default = `build/upstream.json` repo, else `heroiclabs/nakama` |
| `branch` | free text (`INVALID_BRANCH`); suggestions `main`, `stage`, `dev` for nakama-farsi, `master` for heroiclabs |
| `intervalHours` | 1–168, default 6 (`INVALID_INTERVAL`) |
| `token` | optional read-only GitHub token, sealed with `secretbox.Base` (key from `ECONOMY_SECRET_KEY`, HKDF info `base-updates-token`), shown as `********`; `********` keeps it, empty removes it (`INVALID_GITHUB_TOKEN`, `SECRET_KEY_MISSING`). Never logged; the console audit log stores the body with the token masked. The repo is public, so no token is needed: unauthenticated GitHub allows 60 requests an hour and a check uses 3. |

Changing the repo, branch or token, or switching checks back on, queues a check at once.

## Worker

`integrations/base/updates.go`, started in `InitModule`:

- Every node ticks every 5 s and reads `base_updates_state` (one row). A check is due when the interval has passed (`next_at`), a check is queued (`requested_at > checked_at`), or the repo, branch or base differs from the last check. Only then does the node take the `clusterkit` lease `base.updates.check` (90 s); it re-reads the row under the lease and takes the lease again before writing, so one node checks and writes, on any number of nodes. The lease is released on shutdown.
- The check (`internal/upstream`) calls GitHub REST with `Accept: application/vnd.github+json`, `X-GitHub-Api-Version`, a fixed User-Agent and the token if set: `compare/{base}...{branch}` (ahead/behind, up to 250 commits, up to 300 files with patches), `releases?per_page=30`, `tags?per_page=50`. Each request sends the stored `If-None-Match`; a 304 reuses the stored part (and does not use rate limit). ETags are kept per request path, so a new repo, branch or base starts fresh.
- Results: `ok` (stored, next check after the interval); `rate_limited` (403 with `X-RateLimit-Remaining: 0`, any 403 with `Retry-After`, or 429: next check at `Retry-After` or `X-RateLimit-Reset`, at least 1 minute; queued checks wait for it); `base_missing` (404: base commit or branch not in the repo); `error` (network, 5xx, bad JSON: retry after 30 minutes or the interval if shorter); `base_mismatch` (the setting's repo is not the repo the base was recorded for: no request is made); `no_base` (the build has no base commit). A failed check keeps the last good result when the repo, branch and base are unchanged.
- Never on the request path: pages read the cached row only.
- `BASE_UPDATES_API_URL` (runtime env, passed by `build/nakama-start.sh`) replaces `https://api.github.com`; tests and e2e point it at a fake.

## Analysis

- Releases (drafts skipped) published after the base date or tagged on a new commit, newest first, up to 10, bodies up to 8000 characters; new tags = tags whose commit is one of the new commits. Latest release and latest tag overall are shown too. The page shows release notes and tags before the commit list.
- Commits newest first, up to 100 (with a link to the full compare when there are more), merge and breaking flags.
- Files grouped by area: `migrations` (`migrate/sql/`), `core` (`server/`, `internal/`, `main.go`, `apigrpc/`, `social/`, `iap/`, `se/`, `flags/`, `migrate/`), `plugins` (`integrations/`, `samples/`, `sample_go_module/`, `data/modules/`), `console` (`console/`), `deps` (`go.mod`, `go.sum`, `vendor/`, at most 20 vendor files listed), `build` (`build/`, `.github/`, Dockerfiles, compose, `buf.*`, Makefile), `e2e` (`e2e/`, `*_test.go`, `testdata/`), `docs` (`docs/`, `*.md`), `other`.
- Warnings: new migrations; changed old migrations; new environment variables (names added in patches of Go `Getenv`/`LookupEnv`, shell, compose, YAML, env files and docs, minus names that were removed); Go version changes (`go`/`toolchain` in `go.mod`, `FROM golang:` in Dockerfiles, shown as old → new); dependency changes (`go.mod`/`vendor`, with added modules); breaking changes (commit messages or release notes with «breaking», «backwards incompatible», «deprecated», a conventional `type!:` subject or a «### Removed» heading).

## Console API

All under `/v2/console/base`, ACL `SETTINGS` (read for GET, write otherwise), audited as `base`:

| Route | |
|---|---|
| `GET /updates` | build info, settings (token masked), state, `current` (the cached result is for the current repo, branch and base), result |
| `GET /updates/summary` | `newCommits`, `status`, `latestRelease`, `releases`, `warnings` for the badges (0 when switched off or not current) |
| `POST /updates/check` | queue a check (202); at most 5 per 10 minutes for the whole project (`TOO_MANY_CHECKS`, 429 with `retryAfter`) |
| `GET`/`PUT /updates/settings` | the «ریپوی بالایی» section |

## Console

- «سامانه» sidebar section (rail mode with the `cloudDown` icon) with «به‌روزرسانی‌ها» (`#/system/updates`), also in Ctrl K.
- Page: status notice; «این سرور» card (commit, dirty tag, branch, build date, own commits on the base, base commit with link and date, base repo, when it was recorded, how many commits behind); «ریپوی بالایی» card (repo link, status, latest release, last and next check, interval, GitHub quota, new commits, compare link, «بررسی دوباره» that queues a check and polls until it is done); warnings; releases and notes; commits with links; changed files by area (migrations open); «راهنمای به‌روزرسانی» with copyable commands (add or fix the remote, fetch, merge, record, build, migrate) and the fork note — display only. Phosphor Duotone icons, Yekan Bakh, light/dark, stacked cards at 390px, fa/en with Persian digits in fa (hashes and commands stay Latin).
- Badges: sidebar count on «به‌روزرسانی‌ها», a «به‌روزرسانی N» pill in the top bar (count only on phones) that opens the page, and a row in the dashboard «نیاز به رسیدگی» card. The summary is read at most every 5 minutes.

## Release notes for forks

nakama-farsi keeps its own notes in `CHANGELOG-FARSI.md` (the upstream `CHANGELOG.md` is left as heroiclabs writes it, so merges do not conflict). Releases are tagged on `main` as `v3.42-fa.N` (the upstream version, then a running number) with a GitHub release whose notes copy the changelog entry; forks see these tags and releases first on their updates page.

## Tests

- `internal/upstream`: a fake GitHub (`httptest`) for compare, releases and tags, headers, ETag 304 reuse and per-path ETags, 403 and 429 rate limits, 404, 401 (token never in errors), 502, bad JSON, network errors, repo/branch rules, areas and breaking keywords.
- `integrations/base/updates_postgres_test.go` (`BASE_DISPOSABLE_TEST_DB_URL`): defaults, a stored check, no check before the interval, a queued check with ETags, six nodes ticking at once check once, a held lease blocks other nodes until it expires, settings rules, sealed token (sent as `Bearer`, kept by `********`, cleared by empty), base mismatch makes no request, switched off, `SECRET_KEY_MISSING`, rate limit keeps the result and blocks queued checks, 500 and 404 states, `no_base`, the check limit.
- e2e `console/41-updates.spec.ts` against `e2e/fake-github/server.mjs` (started by Playwright `webServer` on `127.0.0.1:7489`; the local server runs with `BASE_UPDATES_API_URL=http://127.0.0.1:7489`). The real GitHub is never called in tests.

## Limits

- GitHub parents only; a fork hosted on GitLab can still watch nakama-farsi on GitHub, but a GitLab parent is not supported.
- Compare returns at most 250 commits and 300 files; beyond that the page links to GitHub. Patches GitHub leaves out (large files) are not scanned for environment variables.
- Environment variables and breaking changes are found by patterns and keywords, so they are hints, not a full review.
- No Telegram alert.
