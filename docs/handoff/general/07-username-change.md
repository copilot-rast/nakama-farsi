---
title: تغییر نام کاربری با هزینه
titleEn: Paid username change
summary: قیمت تغییر بعدی با base_username_quote و تغییر نام با base_username_change
---

ادمین در کنسول (تنظیمات پروژه ← قوانین حساب ← «تغییر نام کاربری») تعیین می‌کند که تغییر نام رایگان است یا هزینه دارد. اگر این قانون خاموش باشد، تغییر نام مثل قبل آزاد است.

پیش‌فرض وقتی روشن شود: اولین تغییر رایگان، بعد ۵۰، ۱۰۰، ۲۰۰، ۴۰۰ و ۸۰۰ جم (بعد از پلهٔ آخر همان ۸۰۰). فاصله بین دو تغییر صفر ساعت.

## قیمت تغییر بعدی

```text
base_username_quote
```

بدنه: `{}`

```json
{
  "enabled": true,
  "free": false,
  "cost": 100,
  "currency": "gem",
  "cooldownHours": 0,
  "cooldownEndsAt": "2026-11-22T10:00:00Z",
  "paidChanges": 1
}
```

| فیلد | معنی |
| --- | --- |
| `enabled` | قانون روشن است. `false` یعنی تغییر رایگان و بی‌محدودیت |
| `free` | تغییر بعدی رایگان است |
| `cost`، `currency` | هزینهٔ تغییر بعدی و کلید ارز |
| `cooldownEndsAt` | فقط وقتی فعلاً نمی‌شود عوض کرد؛ زمان باز شدن (UTC) |
| `paidChanges` | تعداد تغییرهای پولی تا الان |

## تغییر نام

```text
base_username_change
```

```json
{ "username": "new_name", "expectedCost": 100 }
```

`expectedCost` اختیاری است. اگر بفرستید و قیمت عوض شده باشد، تغییر انجام نمی‌شود و `USERNAME_PRICE_CHANGED` برمی‌گردد. پیشنهاد: همان `cost` آخرین `base_username_quote` را بفرستید.

جواب:

```json
{
  "username": "new_name",
  "oldUsername": "old_name",
  "source": "paid",
  "cost": 100,
  "currency": "gem",
  "quote": { "enabled": true, "free": false, "cost": 200, "currency": "gem", "cooldownHours": 0, "paidChanges": 2 }
}
```

`source` یکی از `free`، `paid` یا `admin` است. `quote` قیمت تغییر بعدی است؛ لازم نیست دوباره `base_username_quote` بزنید.

تغییر نام با `UpdateAccount` (یا `base_account_update`) هم همین قانون را دارد و همین هزینه را کم می‌کند. برای نشان دادن قیمت قبل از تأیید، از RPCهای بالا استفاده کنید.

## خطاها

پیام خطا یک JSON است با `code`، `message` (به زبان بازیکن)، `lang` و `meta`.

| کد | `meta` | کار اپ |
| --- | --- | --- |
| `USERNAME_UNCHANGED` | | نام تازه با نام فعلی یکی است |
| `USERNAME_TAKEN` | | نام را کس دیگری دارد |
| `USERNAME_TOO_SHORT`، `USERNAME_TOO_LONG`، `USERNAME_INVALID_CHARACTERS` | | قوانین حساب (`base_account_rules`) |
| `USERNAME_CHANGE_COOLDOWN` | `cooldownEndsAt` | تا آن زمان دکمه را غیرفعال کنید |
| `USERNAME_PRICE_CHANGED` | `cost`، `currency`، `expected` | قیمت تازه را نشان دهید و دوباره تأیید بگیرید |
| `CURRENCY_IS_NOT_ENOUGH` | `currencyKey`، `amount`، `balance` | فروشگاه جم را پیشنهاد دهید |

## نکته‌ها

* همهٔ RPCها را روی سوکت صدا بزنید و اگر سوکت وصل نیست روی HTTP (`POST /v2/rpc/<id>?unwrap=true`).
* بعد از تغییر نام، توکن نشست هنوز نام قبلی را دارد. برای نمایش، نام تازهٔ جواب را به‌کار ببرید یا `base_account_get` بزنید.
* ادمین هم می‌تواند از کنسول نام بازیکن را بدون هزینه عوض کند. این تغییر قیمت و فاصلهٔ بعدی بازیکن را عوض نمی‌کند.
