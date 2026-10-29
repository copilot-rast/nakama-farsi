# Publication module (phase 1)

Port of asp `Modules/Publication` (Faq, News, Notification) as a Go runtime plugin with console pages in the existing console overlay, plus real SMS and email delivery (asp only logs them). Push (Firebase) is out of scope for now.

References:
- asp: `/Users/mazyar/Projects/asp/backend/src/PolicePatrol.Api/Modules/Publication/{Faq,News,Notification}`, provider fields in `Modules/Base/Sms/Services/SmsProviderCatalog.cs` and `Modules/Base/Email/Services/EmailSenderCatalog.cs`, admin UI in `/Users/mazyar/Projects/asp/frontend/src/features/{faqs,news,notifications,smsProviders,emailSender}`.
- Economy module in this repo (same conventions): `docs/economy/spec.md`, `integrations/economy` (crypto.go, resources.go generic CRUD, admin.go route registration, payment_methods.go masking), `console/ui/economy` (overlay).

## Layout

| Part | Path |
|---|---|
| Go plugin | `integrations/publication/` → `publication.so` |
| Migrations | `migrate/sql/2026103100xxxx-publication-*.sql` |
| Console auth for `/v2/console/publication/` | `server/console.go` (same wrapper branch as economy), `console/acl/acl.go` (GET → `NOTIFICATION` read, others → `NOTIFICATION` write) |
| Console pages | extend the existing overlay in `console/ui/economy/` with a second sidebar section «انتشار» / "Publishing" (do not create a second overlay) |

Same rules as economy: PostgreSQL 16, re-runnable Up, Down raises, no code comments, idempotent writes, each console handler registered with one method.

## Shared pieces

- Categories: reuse `economy_category`. A migration widens its `kind` check to also allow `'faq'` and `'news'`. Publication reads that table directly. The console «دسته‌بندی‌ها» page gets the two new kinds.
- Images: reuse `POST /v2/console/economy/uploads` and `GET /v2/economy/files/{name}`.
- Secrets: `ECONOMY_SECRET_KEY` with a different HKDF info string (`publication-provider-config`).

## Data model (tables prefixed `publication_`)

Text fields that users see are `jsonb {"fa", "en"}`, `fa` required. Rich text (FAQ answers, news bodies, email bodies) is a restricted Markdown subset: paragraphs, `**bold**`, `*italic*`, links, lists, headings `##`/`###`, images `![](url)` — rendered to HTML server-side with escaping for email, returned raw + `html` to clients.

- `faq(id, category_id null, question jsonb, answer jsonb (markdown per lang), is_active, sort, created_at, updated_at)`.
- `news(id, title jsonb, summary jsonb, body jsonb, cover_url, category_id null, is_active, is_pinned, published_at, allow_comment bool, free_comment bool, actions jsonb [{type:'url'|'goto', label:{fa,en}, value}], metadata jsonb, created_at, updated_at)`. Only `is_active` and `published_at <= now()` are visible to clients.
- `news_seen(news_id, user_id, seen_at, primary key both)`.
- `news_comment(id, news_id, user_id, parent_id null, message text 1..1000, is_accepted bool, created_at, updated_at)`. `free_comment` → accepted on create, else pending. One reply level (a reply to a reply attaches to the root). Clients see accepted comments plus their own pending ones. Rate limit: 5 comments per user per minute.
- `notification(id, channel text check in ('in_app','sms','email'), title jsonb, body jsonb, action jsonb null {type, value}, target jsonb {mode:'all'|'users'|'group', userIds?, groupId?}, provider text null, status text check in ('queued','sending','done','failed'), recipient_count int, sent_count int, failed_count int, skipped_count int, created_by text, created_at, updated_at)`.
- `notification_recipient(notification_id, user_id, status text check in ('pending','sent','failed','skipped'), address text null, error text null, attempts int, seen_at null, sent_at null, primary key(notification_id, user_id))`.
- `provider(id, channel text check in ('sms','email'), kind text, is_active bool, is_default bool, sandbox bool, config text /*AES-GCM*/, updated_at)`, unique `(channel, kind)`; at most one default per channel.

## Targeting and addresses

- `all`: every non-disabled Nakama user; `users`: explicit ids (max 1000); `group`: members of a Nakama group (state ≤ member, not join requests).
- Phone: account metadata `phone` (accept `09xxxxxxxxx`, `+989…`, `989…`; normalize to `+98…` for Twilio, `09…` for Kavenegar/sms.ir). Email: account `email`, else metadata `email`. Users without an address are `skipped`.
- Recipients are materialized when the notification is created, in batches; delivery runs in a background worker started in `InitModule` (every 5 s, batch 50, up to 3 attempts with backoff, then `failed`). The notification row tracks counts and moves to `done`.

## Channels

- `in_app`: for each recipient call `nk.NotificationsSend` with `persistent: true`, `code` 1001, subject = title in the user's lang (account `lang_tag`, default fa), content `{notificationId, title, body, action}`. Seen tracking via our table.
- `sms` providers (fields; secret fields masked like payment methods):
  - `kavenegar`: `apiKey` (secret), `sender`. `POST https://api.kavenegar.com/v1/{apiKey}/sms/send.json` form `receptor`, `sender`, `message`; `return.status == 200` ok.
  - `smsir`: `apiKey` (secret), `lineNumber`. `POST https://api.sms.ir/v1/send/bulk` header `X-API-KEY`, JSON `{lineNumber, messageText, mobiles:[…]}`; `status == 1` ok.
  - `twilio`: `accountSid`, `authToken` (secret), `from`. `POST https://api.twilio.com/2010-04-01/Accounts/{sid}/Messages.json` basic auth, form `To`, `From`, `Body`; 2xx with `sid` ok.
- `email` providers:
  - `smtp`: `host`, `port`, `username`, `password` (secret), `from`, `fromName`, `security` (`starttls`|`tls`|`none`). Use `net/smtp`; send `multipart/alternative` (text + HTML), UTF-8, `Subject` RFC 2047 encoded.
  - `resend`: `apiKey` (secret), `from`. `POST https://api.resend.com/emails` Bearer, `{from, to:[…], subject, html, text}`.
- **Sandbox** (per provider, default ON for new providers): nothing is sent; the message is written to the recipient row as `sent` with `error = "sandbox"` and to the server log. This protects real users while testing.
- Verify the exact request/response shapes of every provider against its current official docs before coding; if something differs from this file, follow the docs and note it.
- Test send: `POST /v2/console/publication/providers/{id}/test {to, message}` sends one message (honours sandbox) and returns the provider's reply summary.

## Client RPCs (user session)

| RPC | Input | Output |
|---|---|---|
| `publication_faq_list` | `{categoryKey?, q?, lang?}` | `{categories:[{key,name}], items:[{id,question,answer,answerHtml,categoryKey}]}` |
| `publication_news_list` | `{limit?, cursor?, categoryKey?, lang?}` | `{items:[{id,title,summary,coverUrl,publishedAt,isPinned,seen,commentCount}], cursor}` pinned first |
| `publication_news_get` | `{id, lang?}` | full news incl. `bodyHtml`, `actions`, `allowComment`, `seen` |
| `publication_news_seen` | `{id}` | `{}` idempotent |
| `publication_news_comments` | `{newsId, limit?, cursor?}` | `{items:[{id,userId,username,message,parentId,isAccepted,mine,createdAt,replies:[…]}], cursor}` |
| `publication_news_comment` | `{newsId, message, parentId?}` | comment (with `isAccepted`) |
| `publication_inbox` | `{limit?, cursor?, lang?}` | `{items:[{id,title,body,action,createdAt,seen}], unread, cursor}` in_app only |
| `publication_inbox_seen` | `{ids:[…]}` or `{all:true}` | `{unread}` |

Errors: runtime errors with UPPER_SNAKE codes.

## Console API (`/v2/console/publication/`)

- CRUD (`GET list ?q&page&limit[&categoryId][&isActive]`, `GET /{id}`, `POST`, `PUT /{id}`, `DELETE /{id}`) for `faqs`, `news`.
- `GET /comments?newsId&status=pending|accepted&page&limit` (with username, news title), `PUT /comments/{id}` `{isAccepted}`, `DELETE /comments/{id}`.
- `GET /notifications?channel&status&page&limit`, `GET /notifications/{id}` (with recipient page `?rPage&rStatus`), `POST /notifications` `{channel, title, body, action?, target, providerId?}` → created (queued). Validation: sms body ≤ 600 chars, title not needed for sms; email needs title; `users` ids must exist; group must exist. Response includes the resolved `recipientCount`.
- `GET /notifications/preview-target` `?mode&groupId&userIds` → `{count, withPhone, withEmail}`.
- `GET /providers` → `[{id, channel, kind, isActive, isDefault, sandbox, configured, fields:[…], config:{masked}}]` plus the catalog of available kinds; `POST /providers`, `PUT /providers/{id}`, `DELETE /providers/{id}`, `POST /providers/{id}/test`.
- `GET /users/search?q` → `[{id, username, displayName}]` for the target picker (q matches username/display name/id; max 20).
- `GET /summary` → counts for an overview card.

## Console pages

New sidebar section «انتشار» / "Publishing" after «اقتصاد»: سؤال‌های متداول، اخبار، کامنت‌ها (with pending badge), اعلان‌ها (list + «ارسال اعلان» form with channel, target picker with live count from preview-target, markdown body with preview). Delivery services (provider cards like payment methods, with test send) live in the console's Settings page as the «پیامک» (`#/settings/sms`) and «ایمیل» (`#/settings/email`) tabs; `#/publication/providers` redirects to the SMS tab. Same look, fa/en, dark, mobile cards, server-side search/paging, layout-matching loading states.
