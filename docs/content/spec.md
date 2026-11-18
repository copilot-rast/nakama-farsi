# Content packs: Unity Addressables and plain files

A generic plugin (`integrations/content`, `content.so`) that stores versioned packs of files and serves them over HTTP. A pack belongs to one platform, has a version, an optional app version range (`minAppVersion`, `maxAppVersion`) and is active or not; the newest active pack whose range holds the app version is the one clients get. Files are stored once per SHA-256 (`content_blob`), uploaded in parts, in local storage or S3 (`internal/objectstore`).

## Pack kinds

| Kind | For | Upload rule | `content_catalog` answer |
|---|---|---|---|
| `addressables` (default) | Unity Addressables `ServerData/<platform>` folders | one root file matching `packs.catalogPattern` (`catalog*.json\|bin`), optional `.hash` next to it | `baseUrl`, `catalogUrl`, `hashUrl`, `catalogHash` |
| `files` | web clients (Rive `.riv`, `.webp`, `.ogg`, `.json`, fonts, …) | any files, no catalog | the same plus `files:[{path, url, sha256, size, contentType, version}]` (`catalogUrl` is empty) |

Migration `20270103000002-content-files.sql` adds `content_pack.kind` (default `addressables`, so old packs and old clients are unchanged). The kind is sent on `POST /packs` (`kind`, `PACK_INVALID` with field `kind` for anything else). Paths may start with the platform folder (`Web/img/a.webp` is stored as `img/a.webp`). Add the platform you use for the web (for example `Web`) to `packs.platforms`, or use `WebGL`.

## Client

| Call | In | Out |
|---|---|---|
| RPC `content_catalog` | `{platform, appVersion?}` | `{platform, appVersion, pack: null \| {id, kind, platform, version, minAppVersion, maxAppVersion, baseUrl, catalogUrl, catalogPath, hashUrl, catalogHash, fileCount, totalBytes, activatedAt, files?}}`; `PLATFORM_REQUIRED`, `APP_VERSION_REQUIRED` (setting), `PACK_PLATFORM_NOT_ALLOWED`, `PACK_INVALID` |
| `GET /v2/content/catalog?platform=&appVersion=` | same as the RPC, without a session when `packs.publicIndex` | same JSON, `ETag` + `Cache-Control: public, max-age=<packs.indexCacheSeconds>`, `304` on `If-None-Match` |
| `GET\|HEAD /v2/content/files/{packId}/{path}` | | the file. `ETag` = `"<sha256>"`, `304` on `If-None-Match`; `Cache-Control: public, max-age=31536000, immutable` (the URL holds the pack id, and a ready pack never changes), `max-age=300` for `.json`, `.hash`, `.bin`; `Access-Control-Allow-Origin: *`, `X-Content-Type-Options: nosniff`, range requests on local storage. With `packs.redirectToStorage` and public S3 it answers `302` to the bucket URL |

Content types: `.json` `application/json`, `.hash`/`.txt` text, `.xml`, `.webp`, `.png`, `.jpg`/`.jpeg`, `.gif`, `.avif`, `.ogg`/`.oga` `audio/ogg`, `.mp3`, `.wav`, `.m4a`, `.mp4`, `.webm`, `.woff2`, `.woff`, `.ttf`, `.otf`, `.wasm`; `.riv` and anything else `application/octet-stream` (Rive has no registered type). SVG is not served as an image on purpose (it can carry script). For S3 the type is stored with the blob when it is first uploaded.

A web client keeps the `files` list, compares `sha256` with what it cached and downloads only changed files; the browser cache does the rest through the immutable URLs.

## Console (`/v2/console/base/content`)

`settings` (GET/PUT), `packs` (GET list, POST create → `{pack, partBytes, missing}`), `packs/events`, `packs/preview` (what a platform and app version would get), `packs/{id}` (GET with files and the index, DELETE), `packs/{id}/blobs/{sha}/{part}` (PUT a part), `packs/{id}/complete` (`{activate}`; `PACK_FILES_MISSING`), `packs/{id}/activate`, `packs/{id}/deactivate`. The upload panel («بسته‌های محتوا») hashes files in the browser, uploads only the missing blobs and has a «نوع بسته» picker (Addressables or plain files); the pack page shows the kind.

## Settings (`content_settings.value.packs`)

`platforms` (allowed names), `partBytes` (1 MiB), `maxFileBytes` (256 MiB), `maxPackBytes` (2 GiB), `maxFiles` (5000), `uploadHours` (24), `keepInactive` (5 ready inactive packs kept), `singleActive` (false), `requireAppVersion` (false), `publicIndex` (true), `indexCacheSeconds` (60), `baseUrl` (empty = the server's public URL), `redirectToStorage` (true), `catalogPattern`, `requireHash` (false).
