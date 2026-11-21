---
title: کد بازیکن
titleEn: Player code
summary: کد کوتاه عمومی هر حساب، پیدا کردن بازیکن با کد و افزودن دوست با کد
---

هر حساب یک کد کوتاه عمومی دارد (پیش‌فرض ۶ نویسه از حرف لاتین بزرگ و رقم، مثل `K7Q2ZD`). کد هنگام ساخت حساب ساخته می‌شود و عوض نمی‌شود. بازیکن آن را به دیگران می‌دهد تا پیدایش کنند. شناسهٔ اصلی بازیکن (user id) را به کاربر نشان ندهید؛ کد را نشان دهید.

## کد خودم

```text
base_my_code
```

بدنه: `{}` ← `{ "code": "K7Q2ZD" }`

## کد چند بازیکن

برای نشان دادن کد در فهرست دوستان یا تابلوی امتیاز:

```text
base_player_codes
```

```json
{ "ids": ["<user id>", "<user id>"] }
```

حداکثر ۱۰۰ شناسه. جواب: `{ "codes": { "<user id>": "K7Q2ZD" } }`

## پیدا کردن بازیکن با کد

```text
base_user_by_code
```

```json
{ "code": "k7q2zd" }
```

کوچک و بزرگ بودن حروف و فاصلهٔ دو طرف مهم نیست.

```json
{
  "user": { "id": "<user id>", "username": "arash", "displayName": "آرش", "avatarUrl": "", "code": "K7Q2ZD" },
  "relationship": "friends",
  "online": true
}
```

| `relationship` | معنی |
| --- | --- |
| `self` | خود بازیکن |
| `friends` | دوست هستید. فقط در این حالت و `self` فیلد `online` می‌آید |
| `sent` | شما درخواست فرستاده‌اید |
| `received` | او درخواست فرستاده؛ می‌توانید بپذیرید |
| `blocked` | شما او را بلاک کرده‌اید |
| `blocked_by` | او شما را بلاک کرده |
| `unavailable` | حساب بسته شده است |
| `none` | رابطه‌ای نیست |

خطا: `PLAYER_CODE_NOT_FOUND` (۴۰۴).

## افزودن دوست با کد

`base_friends_add`، `base_friends_remove` و `base_friends_block` حالا فیلد `codes` هم می‌گیرند (کنار `ids` و `usernames`):

```json
{ "codes": ["K7Q2ZD"] }
```

خطاها: `PLAYER_CODE_NOT_FOUND`، و `CANNOT_TARGET_SELF` اگر کد خود بازیکن باشد.
