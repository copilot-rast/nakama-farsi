---
title: مستندات API
titleEn: API docs
summary: صفحهٔ /docs، فایل openapi.json و ساخت تایپ‌ها
---

هر سرور مستندات خودش را دارد. در آن فقط RPCهایی هست که روی همان سرور فعال‌اند.

| چه چیزی | آدرس |
| --- | --- |
| صفحهٔ مستندات (Swagger UI) | `<آدرس سرور>/docs` |
| سند OpenAPI نسخهٔ ۳ | `<آدرس سرور>/docs/openapi.json` |

* صفحه و فایل‌هایش از خود سرور می‌آیند و به CDN بیرونی وصل نمی‌شوند.
* هر RPC یک مسیر جدا دارد: `POST /v2/rpc/<id>?unwrap=true` با هدر `Authorization: Bearer <session>`.
* برای هر RPC دو تایپ هست: `<Name>Request` و `<Name>Response`.
* کدهای خطای هر RPC در توضیح همان RPC و در فیلد `x-error-codes` آمده است.
* ورود و نشست زیر برچسب `nakama` است.

## اگر صفحه باز نشد

| جواب | یعنی |
| --- | --- |
| 404 | ادمین مستندات را خاموش کرده است |
| 401 | ادمین برای مستندات ورود کنسول خواسته است |

در هر دو حالت به تیم سرور خبر دهید.

## ساخت تایپ‌ها با Orval

```ts
import { defineConfig } from 'orval';

export default defineConfig({
  api: {
    input: { target: 'https://<API_HOST>/docs/openapi.json' },
    output: {
      mode: 'tags-split',
      target: 'src/api/generated/endpoints',
      schemas: 'src/api/generated/models',
      client: 'fetch',
      baseUrl: 'https://<API_HOST>',
      clean: true,
    },
  },
});
```

```bash
npx orval --config orval.config.ts
npx tsc --noEmit
```

* سند هنگام ساخت سرور از روی کد همان نسخه ساخته می‌شود. بعد از هر دیپلوی دوباره Orval را اجرا کنید.
* خطای `tsc` یعنی شکل یک جواب عوض شده است؛ کد همان صفحه را درست کنید.
* پوشهٔ `generated` را دستی ویرایش نکنید.
* فیلدهای آزاد مثل `metadata` در تایپ‌ها `unknown` هستند. قبل از استفاده شکلشان را بررسی کنید.
