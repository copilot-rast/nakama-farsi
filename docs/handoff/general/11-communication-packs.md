---
title: پیام سریع داخل مسابقه (بسته‌های ارتباطی)
titleEn: In-match quick messages (communication packs)
summary: فهرست بسته‌ها، علاقه‌مندی‌ها و فرستادن پیام بسته داخل مسابقه از راه سوکت
---

بسته‌های ارتباطی همان بسته‌های استیکر کنسول هستند با یک «نوع» و «جای استفاده». نوع یکی از `sticker`، `text`، `gif` یا `emoji` است. بسته‌ای که «در مسابقه» روشن دارد اینجا می‌آید. بستهٔ رایگان (`starter`) برای همه باز است؛ بستهٔ آیتمی با خرید یا گرفتن آیتمش باز می‌شود (فروش از فروشگاه معمول، `itemKey`).

همهٔ RPCها روی سوکت صدا زده شوند (`callRpc`). HTTP فقط جایگزین است.

## فهرست بسته‌ها

```text
base_comm_packs
```

بدنه: `{ "lang": "fa", "kind": "text", "mine": false }`. هر سه اختیاری‌اند. `mine: true` فقط بسته‌های بازِ بازیکن را می‌دهد.

```json
{
  "enabled": true,
  "lang": "fa",
  "cooldown": { "minIntervalSeconds": 4, "maxCount": 6, "windowSeconds": 30, "lockSeconds": 30 },
  "maxFavorites": 0,
  "packs": [
    {
      "key": "hello", "kind": "text", "name": "سلام‌ها", "description": "", "iconUrl": "",
      "access": "free", "starter": true, "owned": true,
      "items": [ { "key": "gg", "ref": "hello/gg", "name": "", "text": "بازی خوبی بود", "favorite": true } ]
    },
    {
      "key": "faces", "kind": "emoji", "name": "چهره‌ها", "access": "item", "starter": false, "owned": false,
      "itemKey": "faces_pack",
      "item": { "key": "faces_pack", "name": "چهره‌های ویژه", "price": {"currencyKey": "gem", "amount": 25}, "finalPrice": {"currencyKey": "gem", "amount": 25}, "discountPercent": 0, "productIds": ["…"] },
      "items": [ { "key": "angry", "ref": "faces/angry", "assetId": "emoji.angry", "favorite": false } ]
    }
  ]
}
```

| فیلد مورد | معنی |
| --- | --- |
| `text` | فقط در بستهٔ `text`، به زبان درخواست |
| `assetId` | شناسهٔ فایلی که داخل اپ است (استیکر، گیف، ایموجی). اگر هست، همان را نشان بدهید |
| `imageUrl` | تصویر از سرور، وقتی `assetId` نیست |
| `favorite` | در بستهٔ قفل همیشه `false` |

`enabled: false` یعنی پیام سریع خاموش است؛ دکمهٔ آن را نشان ندهید. `cooldown` همان قانونی است که سرور اجرا می‌کند؛ برای خاکستری کردن دکمه بعد از ارسال به کار ببرید.

## علاقه‌مندی‌ها

```text
base_comm_favorite    { "pack": "hello", "item": "gg", "favorite": true }
base_comm_favorites   { "lang": "fa", "kind": "", "pack": "" }
```

`favorite: false` حذف می‌کند و هیچ‌وقت خطا نمی‌دهد. افزودن فقط برای موردی که الان قابل فرستادن است:

| کد | معنی |
| --- | --- |
| `PACK_NOT_AVAILABLE` | بسته نیست یا برای مسابقه نیست |
| `ITEM_NOT_AVAILABLE` | مورد نیست یا خاموش است |
| `PACK_NOT_OWNED` | بسته را ندارد؛ صفحهٔ خرید آیتم را باز کنید |
| `FAVORITES_FULL` | سقف علاقه‌مندی پر است (وقتی ادمین سقف گذاشته) |

فهرست علاقه‌مندی‌ها جدیدترین اول است: `{ "items": [ { "pack", "packKind", "key", "ref", "name", "text?", "assetId?", "imageUrl?", "addedAt" } ] }`.

## فرستادن داخل مسابقه (سوکت)

پیام بسته داده‌ی خود مسابقه است و از دستگیرهٔ همان بازی رد می‌شود. opcode را هر بازی خودش تعیین می‌کند (در بازی نمونهٔ دوز: فرستادن `6`، دریافت `7`، خطا `5`).

فرستادن (`match_data_send`):

```json
{ "pack": "hello", "item": "gg", "clientMessageId": "c-17" }
```

`clientMessageId` اختیاری است (تا ۶۴ نویسه). تکرار همان شناسه تا یک دقیقه همان پیام قبلی را برمی‌گرداند و دوباره پخش نمی‌شود.

دریافت (برای همهٔ بازیکنان مسابقه، فرستنده هم):

```json
{
  "matchId": "…", "messageId": "9f2c…", "senderId": "<userId>",
  "pack": "hello", "packKind": "text", "item": "gg",
  "text": { "fa": "بازی خوبی بود", "en": "Good game" },
  "assetId": "", "imageUrl": "",
  "sentAt": "2026-11-21T10:00:00Z", "clientMessageId": "c-17"
}
```

`text` همهٔ زبان‌ها را دارد؛ زبان خود بازیکن را نشان بدهید. پیام‌ها ذخیره نمی‌شوند.

خطا فقط به فرستنده:

```json
{ "code": "COOLDOWN_ACTIVE", "retryAfterMs": 2800 }
```

| کد | کار اپ |
| --- | --- |
| `COOLDOWN_ACTIVE` | فاصلهٔ دو پیام رعایت نشد؛ تا `retryAfterMs` دکمه را خاموش کنید |
| `SPAM_LOCKED` | پیام زیاد شد؛ تا `retryAfterMs` گفتگو قفل است |
| `CHANNEL_CLOSED` | گفتگوی این مسابقه بسته شده (پایان مسابقه یا پشتیبانی) |
| `SEND_NOT_ALLOWED` | پشتیبانی ارسال این بازیکن را در این مسابقه قطع کرده |
| `COMMS_DISABLED` | پیام سریع خاموش است |
| `PACK_NOT_AVAILABLE`، `ITEM_NOT_AVAILABLE`، `PACK_NOT_OWNED` | مثل علاقه‌مندی‌ها |
| `INVALID_COMM_MESSAGE` | بدنهٔ پیام درست نیست |

## در چت معمولی

بسته‌ای که «در گفتگوی چت» روشن دارد مثل قبل در `base_stickers` می‌آید و با `{"sticker": "pack/key"}` فرستاده می‌شود. حالا هر بسته `kind` دارد و هر استیکر شاید `text` یا `assetId` داشته باشد. پیام ذخیره‌شده برای بستهٔ غیر استیکری `stickerKind`، و در صورت وجود `stickerText` (همهٔ زبان‌ها) و `stickerAsset` هم دارد.
