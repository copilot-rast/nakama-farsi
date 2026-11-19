---
title: توکن کاربر comms (چت، تماس، جلسه)
titleEn: comms user token (chat, calls, meetings)
summary: گرفتن توکن کوتاه‌عمر با comms_user_token و عوض کردنش با نشست comms
---

برای چت، تماس و جلسه در سرویس comms، اپ یک توکن کاربر کوتاه‌عمر لازم دارد. این توکن را سرور بازی برای بازیکنِ واردشده امضا می‌کند. اپ هیچ کلید یا رمز comms ندارد.

## RPC

```text
comms_user_token
```

بدنه خالی است: `{}`. بازیکن از نشست خوانده می‌شود.

جواب وقتی روشن است:

```json
{
  "enabled": true,
  "url": "https://comms.example.com",
  "appId": "0b5c2a4e-6f0d-4e5b-9a43-1d2c3b4a5f60",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expiresAt": "2026-11-19T06:32:04Z"
}
```

| فیلد | معنی |
| --- | --- |
| `url` | نشانی عمومی comms. همهٔ درخواست‌های comms به همین نشانی می‌روند |
| `appId` | شناسهٔ برنامه در comms. فقط برای نمایش یا لاگ |
| `token` | توکن کاربر. یک بار مصرف است و فقط برای ساختن نشست |
| `expiresAt` | پایان اعتبار توکن (UTC). پیش‌فرض ۵ دقیقه، حداکثر ۱۰ دقیقه |

جواب وقتی خاموش است (خطا نیست):

```json
{ "enabled": false, "reason": "COMMS_TOKENS_DISABLED" }
```

| `reason` | کار اپ |
| --- | --- |
| `COMMS_TOKENS_DISABLED` | ادمین این قابلیت را خاموش کرده. بخش چت/تماس comms را نشان ندهید |
| `COMMS_NOT_CONFIGURED` | تنظیم سرور کامل نیست. مثل خاموش رفتار کنید |

خطا:

| کد | کار اپ |
| --- | --- |
| `COMMS_TOKEN_RATE_LIMITED` (429) | درخواست زیاد شد. چند ثانیه بعد دوباره بگیرید؛ در حلقه صدا نزنید |

## گرفتن نشست comms

توکن را در `POST <url>/v1/sessions` با سرآیند `Authorization: Bearer <token>` و بدون بدنه بفرستید. جواب `201` است با `session_token`، `resume_token` و `user`. از اینجا به بعد فقط با نشست comms کار کنید و برای تازه کردن، `POST <url>/v1/sessions/refresh` با `resume_token`.

با SDK وب comms (`@comms/sdk`) این کار خودکار است؛ فقط `getUserToken` را به همین RPC وصل کنید:

```ts
import { createClient } from '@comms/sdk'

async function commsToken(): Promise<string> {
  const r = await nakama.rpc(session, 'comms_user_token', {})
  const data = r.payload as { enabled: boolean; token?: string; reason?: string }
  if (!data.enabled || !data.token) throw new Error(data.reason || 'COMMS_TOKENS_DISABLED')
  return data.token
}

const info = (await nakama.rpc(session, 'comms_user_token', {})).payload as { enabled: boolean; url?: string }
if (info.enabled) {
  const comms = createClient({ baseUrl: info.url!, getUserToken: commsToken, session: { store: 'device' } })
  comms.start()
}
```

## نکته‌ها

* RPC را روی سوکت صدا بزنید و اگر سوکت وصل نیست روی HTTP (`POST /v2/rpc/comms_user_token?unwrap=true`).
* توکن را ذخیره نکنید. هر بار که SDK یا اپ نشست تازه می‌خواهد، یک توکن تازه بگیرید. توکن مصرف‌شده دوباره پذیرفته نمی‌شود.
* شناسهٔ کاربر در comms همان شناسهٔ بازیکن (Nakama user id) است و نام نمایشی همان نام نمایشی بازیکن (یا نام کاربری).
* این توکن به چت صوتی بازی (`voice_token`) ربطی ندارد. ویس مسابقه همچنان با `voice_token` کار می‌کند، حتی وقتی این بخش خاموش است و برعکس.
* نشانی صفحهٔ وب شما باید در `allowed_origins` برنامهٔ comms باشد؛ این را مدیر comms تنظیم می‌کند.
