---
title: صندوق خانه‌دار، گردونهٔ پلکانی، افزایش‌دهنده و جایزهٔ روزانه
titleEn: Chests with slots, ladder wheel, boosters and the daily reward track
summary: نتیجهٔ صندوق چندخانه، قیمت پلکانی و تأیید نمایش نتیجهٔ گردونه، فهرست افزایش‌دهنده‌ها و روزهای ویژه و مرحلهٔ بی‌پایان زنجیره
---

همهٔ RPCهای این بخش را روی سوکت با `callRpc` بفرستید (HTTP فقط پشتیبان). جواب‌ها JSON هستند. خطاها همان شکل همیشگی بازی‌های شانس را دارند: `message` یک JSON است با `code`، `status` و گاهی `meta`.

## صندوق خانه‌دار (`prizeType: "slots"`)

در صندوق خانه‌دار هر خانه جدا قرعه می‌خورد؛ ممکن است چند خانه با هم بیایند. اگر هیچ خانه‌ای نیاید و ادمین «قرعهٔ دوباره» را روشن گذاشته باشد (پیش‌فرض)، سرور دوباره قرعه می‌کشد، پس صندوق خالی باز نمی‌شود.

بازی کردن مثل قبل است:

```text
economy_chance_play               { "id": "<gameId>", "lang": "fa" }
economy_chance_play_with_ticket   { "id": "<gameId>", "expectedCost": 10 }
```

در جواب، کنار فیلدهای همیشگی، `outcome` می‌آید:

```json
{
  "code": "SUCCESS",
  "playId": "8c0d…",
  "status": "done",
  "outcome": [
    { "slot": 0, "kind": "currency", "currencyKey": "coin", "amount": 154 },
    { "slot": 1, "kind": "prize", "prizeId": "41a7…", "prize": { "id": "41a7…", "key": "gems_3", "name": "۳ الماس" }, "rewards": [{ "type": "currency", "key": "gem", "action": "add", "value": 3 }] }
  ],
  "rewards": [
    { "type": "currency", "key": "coin", "action": "add", "value": 154 },
    { "type": "currency", "key": "gem", "action": "add", "value": 3 }
  ],
  "wallet": { "coin": 1154, "gem": 3 }
}
```

| فیلد خانه | معنی |
| --- | --- |
| `slot` | شمارهٔ خانه در فهرست `prize` بازی (از ۰) |
| `kind` | `currency` (مقدار تصادفی در بازه)، `prize` یا `prizeGroup` |
| `currencyKey`, `amount` | فقط برای `currency` |
| `prizeId`, `prize`, `rewards` | برای جایزه و گروه جایزه؛ در گروه، `prizeGroupId` هم هست |

در `economy_chance_game` و `economy_chance_games`، هر ردیف صندوق `kind` دارد و ردیف ارزی `currency: { currencyKey, currencyName, min, max }`. شانس هر ردیف، درصدِ آمدن همان خانه است و جمع‌شان ۱۰۰ نیست. `economy_chance_history` هم برای این بازی‌ها `outcome` برمی‌گرداند.

## گردونه: قیمت پلکانی، جایزهٔ بزرگ و تأیید نتیجه

### قیمت پلکانی

اگر بازی `priceLadder` داشته باشد، بلیت اولِ هر روز (به منطقهٔ زمانی پروژه) قیمت پلهٔ اول را دارد، دومی پلهٔ دوم و …؛ بعد از آخرین پله همان قیمت آخر می‌ماند. در نمای بازی:

```json
"price": { "currencyKey": "gem", "amount": 15 },
"priceLadder": [10, 15, 20, 25, 30],
"ticketsToday": 1
```

`price.amount` همیشه قیمت **بلیت بعدی** است. همین عدد را در `expectedCost` بفرستید. اگر قیمت در این فاصله عوض شده باشد، چیزی کم نمی‌شود و این خطا می‌آید:

| کد | کار اپ |
| --- | --- |
| `CHANCE_PRICE_CHANGED` (`status` 409، `meta: { amount, expected }`) | قیمت تازه (`meta.amount`) را نشان دهید و بعد از تأیید بازیکن دوباره بفرستید |

جواب بلیت `nextTicketCost` هم دارد: قیمت بلیت بعدی.

### جایزهٔ بزرگ

ردیف جایزهٔ بزرگ در `prize` با `"jackpot": true` مشخص است. اگر ادمین تضمین گذاشته باشد، بعد از چند بازی بی‌جایزهٔ بزرگ، سرور خودش جایزهٔ بزرگ را می‌دهد. تعداد و شمارنده به اپ داده نمی‌شود؛ اپ فقط نتیجه (`slot`) را نشان می‌دهد.

### تأیید دیدن نتیجه

وقتی `revealAck` در نمای بازی `true` است، جایزه همان لحظه داده می‌شود ولی آخرین نتیجهٔ تأییدنشده در `pendingReveal` می‌ماند تا اگر اپ وسط چرخیدن بسته شد، دفعهٔ بعد نتیجه را نشان دهد:

```json
"revealAck": true,
"pendingReveal": { "playId": "8c0d…", "mode": "ticket", "slot": 3, "prizeId": "…", "prize": { "name": "۲۵ الماس" }, "rewards": [ … ], "createdAt": "2026-11-21T09:12:04Z" }
```

جواب بازی در این حالت `"needsAck": true` دارد. بعد از تمام شدن نمایش:

```text
economy_chance_ack   { "playId": "8c0d…" }
```

```json
{ "code": "SUCCESS", "acknowledged": true, "playId": "8c0d…", "ackedAt": "2026-11-21T09:12:09Z" }
```

تأیید دوباره بی‌خطر است. شناسهٔ ناشناخته یا مال بازیکن دیگر: `PLAY_NOT_FOUND`. نتیجهٔ تأییدنشده جلوی بازی بعدی را نمی‌گیرد.

## افزایش‌دهندهٔ زمان‌دار (XP booster)

افزایش‌دهنده یک جایزه است (از فروشگاه، گردونه، صندوق یا هر جای دیگر). تا وقتی فعال است، امتیازهایی که رویدادها به آن متا (مثلاً `xp`) اضافه می‌کنند در ضریب ضرب می‌شوند. چند افزایش‌دهنده روی هم ضرب نمی‌شوند: گرفتن دوباره زمان را اضافه می‌کند و ضریبِ تازه جای قبلی را می‌گیرد.

در `rewards` به این شکل می‌آید:

```json
{ "type": "booster", "key": "xp", "action": "give", "value": { "multiplier": 2, "minutes": 60 } }
```

`action` یا `give` (به زمانِ باقی‌مانده اضافه می‌شود) یا `set` (از همین حالا دوباره شروع می‌شود).

فهرست افزایش‌دهنده‌های فعال بازیکن:

```text
economy_boosters   { "lang": "fa" }
```

```json
{ "code": "SUCCESS", "lang": "fa", "items": [
  { "metaKey": "xp", "metaName": "تجربه", "multiplier": 2, "expiresAt": "2026-11-21T10:12:04Z", "secondsLeft": 3540 }
] }
```

شمارش معکوس را با `expiresAt` بسازید و بعد از صفر شدن دوباره این RPC را بخوانید.

## جایزهٔ روزانه: روزهای ویژه، برگشت به اول هفته، مرحلهٔ بی‌پایان

`economy_streaks` برای هر زنجیره این فیلدهای تازه را دارد:

```json
{
  "missAction": "week_start",
  "weekLength": 7,
  "phase": "track",
  "extras": [ { "day": 8, "prize": { "name": "صندوق ۱" }, "grantMode": "claim", "status": "locked" } ],
  "infinity": { "currencyKey": "coin", "min": 150, "max": 180 }
}
```

| فیلد | معنی |
| --- | --- |
| `missAction: "week_start"` | با غیبت، بازیکن به روز اولِ همان هفته برمی‌گردد (با `weekLength` ۷: غیبت در روز ۱۲ یعنی روز ۸) |
| `extras` | جایزهٔ اضافهٔ روزهای ویژه؛ فقط وقتی داده می‌شود که بازیکن بدون غیبت به آن روز رسیده باشد. `status` مثل روزهای معمولی است (`locked`، `claimable`، `claimed`) و با `claimId` و `economy_reward_claim` گرفته می‌شود |
| `phase` | `track` یا `infinity`. در `infinity` روزها دیگر جایزهٔ روز ندارند و غیبت بازیکن را عقب نمی‌برد |
| `infinity` | اگر `null` نباشد: بعد از آخرین روز چرخه، هر روز مقداری تصادفی بین `min` و `max` از `currencyKey` خودکار به کیف پول می‌رود |

در نتیجهٔ رویداد (`economy_event_record` و همهٔ جاهایی که `streaks` برمی‌گردد) هر تغییر زنجیره در مرحلهٔ بی‌پایان این‌ها را هم دارد:

```json
{ "key": "daily_reward", "count": 31, "phase": "infinity", "infinityCurrency": "coin", "infinityAmount": 163 }
```
