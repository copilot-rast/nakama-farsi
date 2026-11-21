---
title: تنظیمات مرحله از پنل، قالب ستاره و جزئیات نتیجه
titleEn: Stage settings from the console, star templates and result details
summary: فیلد settings هر مرحله و فصل، قانون ستاره و کمینهٔ زمان در فهرست مرحله‌ها، رتبهٔ هر معیار در نتیجه، timesSeen، مرحله‌ای که بعد از برد تکرار نمی‌شود، شناسهٔ ثابت پله و یک راه برای گشت آزاد
---

همهٔ این‌ها در پنل تنظیم می‌شوند و بازی فقط می‌خواند. RPCها همان قبلی‌اند: `campaign_chapters`، `campaign_attempt_start` و `campaign_attempt_complete`. روی سوکت با `callRpc` بفرستید.

## کی چه چیزی را صدا بزنید

| زمان | RPC | ورودی |
| --- | --- | --- |
| باز شدن صفحهٔ نقشه یا فهرست مرحله‌ها | `campaign_chapters` | `{ "lang": "fa" }` |
| قبل از شروع تایمر مأموریت | `campaign_attempt_start` | `{ "stageKey": "..." }` |
| بعد از برد یا باخت | `campaign_attempt_complete` | `{ "attemptId", "stageKey", "success", "elapsedSeconds", "healthPercent", "fields", "lang" }` |

مقدار `healthPercent` را همیشه بفرستید. اگر نفرستید و تنظیم «اگر فیلد نیامد» روی رد باشد، برد فقط ۱ ستاره می‌گیرد.

## فیلدهای تازه در `campaign_chapters`

هر مرحله در `chapters[].stages[]` این فیلدها را هم دارد:

| فیلد | معنی |
| --- | --- |
| `settings` | تنظیمات مأموریت که ادمین در فرم پنل پر کرده. هر فیلدی که خالی مانده با مقدار پیش‌فرضش می‌آید. متن‌های چندزبانه به زبان `lang` تبدیل شده‌اند |
| `starRules` | قانون ستاره‌ای که سرور با آن حساب می‌کند |
| `starTemplate` | کلید قالب ستاره (`base`، `race`، `stealth`، `quest`) یا خالی |
| `minSeconds` | کمترین زمانی که نتیجه پذیرفته می‌شود |
| `noReplayAfterWin` | اگر `true` باشد، بعد از برد دیگر شروع نمی‌شود |
| `replayable` | آیا همین بازیکن الان می‌تواند شروعش کند |

فصل هم `settings` خودش را دارد. هر پله در `milestones[]` یک `id` ثابت دارد.

شکل `settings` را خود پروژه در پنل تعریف می‌کند («تنظیمات مرحله‌ها ← فیلدهای مرحله و فصل»). نمونه:

```json
{
  "key": "chapter-close-beta-level-02",
  "minSeconds": 35,
  "starTemplate": "",
  "replayable": true,
  "starRules": [
    { "stars": 3, "conditions": [
      { "field": "timeSavedPercent", "op": ">=", "value": 30 },
      { "field": "healthPercent", "op": ">=", "value": 75 },
      { "field": "timesSeen", "op": "<=", "value": 0, "missing": "pass" } ] }
  ],
  "settings": {
    "unityLevelId": "M_DangerousChemicalC2",
    "brief": "عملیات شبیخون",
    "missions": [ { "mode": "capture", "timer": 240, "showArrowGuidance": true, "enemies": [ { "speedMax": 140 } ] } ]
  }
}
```

اگر بازی یک فیلد را نمی‌شناسد، از آن رد شود. فیلدی که نیامده یعنی ادمین تعریفش نکرده است.

## قالب‌های ستاره

| قالب | معیارها |
| --- | --- |
| `base` | زمان، سلامت و دفعهٔ دیده شدن |
| `race` | زمان و دفعهٔ دیده شدن |
| `stealth` | سلامت و دفعهٔ دیده شدن |
| `quest` | مثل `base` |

دفعهٔ دیده شدن را با نام `timesSeen` در `fields` بفرستید، مثلاً `"fields": { "timesSeen": 1 }`. اگر مأموریت تعقیب پنهانی نیست، لازم نیست بفرستید؛ نیامدنش قبول حساب می‌شود.

## جواب `campaign_attempt_complete`

علاوه بر فیلدهای قبلی:

| فیلد | معنی |
| --- | --- |
| `criteria` | رتبهٔ هر معیار برای صفحهٔ نتیجه: `[{ "field", "value", "stars" }]` |
| `starRules`، `minSeconds` | همان قانونی که حساب شد |
| `timeSavedPercent` | درصد زمانی که باقی ماند |
| `milestoneIds` | شناسهٔ پله‌هایی که همین الان پرداخت شدند |
| `replayable` | آیا بعد از این نتیجه باز هم شروع می‌شود |

```json
{
  "stars": 2,
  "criteria": [
    { "field": "timeSavedPercent", "value": 94.17, "stars": 3 },
    { "field": "timesSeen", "value": 1, "stars": 2 }
  ],
  "rewards": [ { "currencyKey": "coin", "amount": 80000 } ],
  "milestoneIds": ["1"],
  "replayable": true
}
```

ستارهٔ نهایی همان `stars` است. `criteria` فقط برای نمایش است.

## خطای تازه

| کد | کی | کار اپ |
| --- | --- | --- |
| `STAGE_ALREADY_WON` (`status` 403) | شروع مرحله‌ای که `noReplayAfterWin` دارد و بازیکن یک بار برده | دکمهٔ شروع را خاموش کنید؛ `replayable` همین را از قبل می‌گوید |
| `STAGE_FINISHED_TOO_FAST` (`meta: { minSeconds }`) | زمان کمتر از `minSeconds` | نتیجه پذیرفته نشد؛ پیام بدهید |

## پله‌های فصل

هر پله `id` ثابت دارد. اگر ادمین درصد پله را عوض کند، کسی که آن پله را گرفته دوباره جایزه نمی‌گیرد. برای علامت روی نوار پیشرفت از `reached` و `claimed` همان پله استفاده کنید.

## گشت آزاد

گشت آزاد فقط از مرحله‌های تکرارشوندهٔ یک فصل (مثلاً فصل `freeride`) می‌گذرد: همان `campaign_attempt_start` و `campaign_attempt_complete`. جدول جدای «جایزهٔ گشت آزاد» به‌طور پیش‌فرض خاموش است. در این حالت `freeride[]` در `campaign_chapters` خالی است و `campaign_freeride_claim` جواب `FREERIDE_NOT_FOUND` می‌دهد.
