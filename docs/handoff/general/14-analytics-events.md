---
title: رویدادهای آنالیتیکس (Analytics.Send)
titleEn: Analytics events (Analytics.Send)
summary: یک استاندارد برای همهٔ رویدادها؛ ارسال دسته‌ای با سوکت از insights_send
---

هر رویداد بازی یک **نام**، چند **تگ** (گروه تگ + شناسه) و یک **مقدار** عددی اختیاری دارد. سرور هر رویداد را یک بار با شناسهٔ بازیکن، جلسه، زمان، پلتفرم و نسخه ذخیره می‌کند و پنل می‌تواند آن را با هر ترکیبی از تگ‌ها جمع بزند: کل، بر اساس بازیکن، ماشین، مأموریت، یا ماشین + مأموریت + بازیکن.

```text
Analytics.Send(name: "DistanceOverall",
               tags: [{tagGroup: "mission", id: "3434343"}, {tagGroup: "car", id: "345345"}],
               value: 150)
```

## یک رویداد

```json
{"name": "DistanceOverall", "tags": [{"tagGroup": "mission", "id": "3434343"}, {"tagGroup": "car", "id": "345345"}], "value": 150, "time": 1795280400000}
```

| فیلد | قانون |
| --- | --- |
| `name` | با حرف انگلیسی شروع شود؛ حرف، عدد، `_ . -`؛ حداکثر ۶۴ نویسه. بزرگی و کوچکی حرف مهم است؛ PascalCase پیشنهاد می‌شود. |
| `tags` | فهرست `{tagGroup, id}`؛ یا شیء `{"mission": "3434343"}`. کلید گروه: حرف کوچک انگلیسی، عدد و `_`. شناسه تا ۶۴ نویسه و بدون `:`. هر گروه یک بار؛ حداکثر ۸ تگ. |
| `value` | عدد اختیاری. بدون مقدار، رویداد فقط شمرده می‌شود. |
| `time` | زمان دستگاه (میلی‌ثانیهٔ یونیکس یا ISO). اگر بیش از ۷۲ ساعت عقب یا ۵ دقیقه جلو باشد، زمان سرور ثبت می‌شود و رویداد علامت می‌خورد. |
| `fields` | شیء آزاد کوچک (حداکثر ۱۶ کلید، متن تا ۲۵۶ نویسه). در جدول زمانی دیده می‌شود ولی جمع زده نمی‌شود. |
| `path` | شکل قدیمی به جای `name` + `tags`: `"DistanceOverall:345345:3434343"`. ترتیب گروه‌ها را ادمین در پنل برای هر رویداد تعیین می‌کند. |

در طول مرحله لازم نیست تگ `mission` و `car` را بفرستید؛ سرور آن‌ها را از تلاش باز بازیکن پر می‌کند (بخش ۲۲، «سرور تگ مأموریت و ماشین را خودش پر می‌کند»). شناسه‌های عددی بالا فقط نمونهٔ شکل‌اند؛ مقدار واقعی کلید مرحله و کلید ماشین است، مثل `chapter-close-beta-level-01` و `tiggo8`.

رویدادهای GameAnalytics هم پذیرفته می‌شوند (`type`: `design`، `progression`، `resource`، `business`، `error`) با همان فیلدهای GameAnalytics؛ سرور آن‌ها را به همین استاندارد تبدیل می‌کند.

## ارسال: `insights_send` (سوکت اول)

رویدادها را در صف محلی نگه دارید و دسته‌ای بفرستید: هر `flushSeconds` ثانیه (پیش‌فرض ۳۰)، وقتی صف به `maxBatch` رسید (پیش‌فرض ۱۰۰)، و وقتی بازی به پس‌زمینه می‌رود. با سوکت باز `socket.rpc("insights_send", payload)`؛ اگر سوکت نیست، `POST /v2/rpc/insights_send`.

```json
{
  "batchId": "b-01JD3W7T2Q9V7Z3M2K4X",
  "sessionId": "s-8f2c1a",
  "platform": "android",
  "appVersion": "1.4.0",
  "game": "",
  "events": [
    {"name": "DistanceOverall", "tags": [{"tagGroup": "car", "id": "345345"}], "value": 150, "time": 1795280400000},
    {"path": "DistanceOverall:345345:3434343", "value": 20}
  ]
}
```

- `batchId` لازم است (۸ تا ۶۴ نویسه، `A-Z a-z 0-9 _ . : -`). برای هر دسته یک شناسهٔ تازه بسازید و **همان دسته را با همان شناسه** دوباره بفرستید تا جواب بگیرید؛ سرور دوباره نمی‌شمارد و `duplicate: true` برمی‌گرداند.
- `sessionId` همان شناسه‌ای است که به `insights_session` می‌دهید (شروع، پینگ، پایان جلسه).
- `game` فقط وقتی سرور چند بازی دارد (کلید بازی از ادمین)؛ وگرنه خالی.

جواب:

```json
{"batchId": "b-01JD3W7T2Q9V7Z3M2K4X", "duplicate": false, "received": 2, "accepted": 2, "flagged": 0, "rejected": 0, "dropped": 0,
 "results": [{"index": 0, "name": "DistanceOverall", "status": "stored"}, {"index": 1, "name": "DistanceOverall", "status": "stored"}]}
```

`status`: `stored`، `flagged` (ذخیره شد ولی علامت خورد)، `rejected` (ذخیره نشد؛ `code` دلیل را می‌گوید)، `dropped` (رویداد را ادمین خاموش کرده). بعد از هر جواب ۲۰۰ دسته را از صف بردارید؛ رویدادهای `rejected` را دوباره نفرستید.

| خطای کل دسته | کار اپ |
| --- | --- |
| `ANALYTICS_RATE_LIMITED` (429) | دسته را نگه دارید و بعد از `retryAfterSeconds` بفرستید. |
| `ANALYTICS_BACKLOG_FULL` (503) | سرور پشت سر هم کار دارد؛ دسته را نگه دارید و ۳۰ ثانیه بعد بفرستید. |
| `ANALYTICS_DISABLED` (503) | ارسال را تا اجرای بعدی متوقف کنید. |
| `ANALYTICS_BATCH_ID_REQUIRED`، `ANALYTICS_INVALID_BATCH_ID`، `ANALYTICS_BATCH_EMPTY`، `ANALYTICS_BATCH_TOO_LARGE` | خطای برنامه است؛ دسته را دور بریزید. |
| خطای شبکه یا 5xx دیگر | دسته را با همان `batchId` دوباره بفرستید. |

کدهای هر رویداد: `ANALYTICS_INVALID_EVENT_NAME`، `ANALYTICS_INVALID_TAG`، `ANALYTICS_TOO_MANY_TAGS`، `ANALYTICS_DUPLICATE_TAG_GROUP`، `ANALYTICS_UNKNOWN_EVENT`، `ANALYTICS_UNKNOWN_TAG_GROUP`، `ANALYTICS_UNKNOWN_TAG_VALUE`، `ANALYTICS_MISSING_TAG`، `ANALYTICS_INVALID_VALUE`، `ANALYTICS_VALUE_REQUIRED`، `ANALYTICS_VALUE_OUT_OF_RANGE`، `ANALYTICS_INVALID_FIELDS`، `ANALYTICS_CLIENT_TIME_OUT_OF_RANGE`، `ANALYTICS_INVALID_TYPED_EVENT`، `ANALYTICS_PATH_GROUPS`.

## تنظیمات از سرور: `insights_config`

یک بار بعد از ورود بخوانید (`{"game": ""}`):

```json
{"events": {"enabled": true, "maxBatch": 100, "maxTags": 8, "maxNameLength": 64, "flushSeconds": 30, "maxQueue": 1000, "game": "",
  "events": [{"name": "DistanceOverall", "pathGroups": ["car", "mission"], "requiredTags": []}],
  "groups": [{"key": "car", "closed": true}, {"key": "mission", "closed": false}]},
 "sessions": {"source": "client", "pingSeconds": 60, "timeoutSeconds": 300}}
```

اگر صف محلی از `maxQueue` گذشت، قدیمی‌ترین رویدادها را دور بریزید. `enabled: false` یعنی چیزی نفرستید.

## نمونه با nakama-js

```js
const queue = [];
function send(name, tags, value) { queue.push({ name, tags, value, time: Date.now() }); }
async function flush(socket, sessionId) {
  if (!queue.length) return;
  const events = queue.splice(0, 100);
  const batch = { batchId: 'b-' + crypto.randomUUID(), sessionId, platform: 'web', appVersion: '1.0.0', events };
  for (let i = 0; i < 5; i++) {
    try { await socket.rpc('insights_send', JSON.stringify(batch)); return; }
    catch (e) { await new Promise((r) => setTimeout(r, 2000 * (i + 1))); }
  }
  queue.unshift(...events);
}
send('DistanceOverall', [{ tagGroup: 'mission', id: '3434343' }, { tagGroup: 'car', id: '345345' }], 150);
```

## یونیتی

پکیج `com.nakamafarsi.unity` هنوز این بخش را ندارد. چیزهایی که باید اضافه شود:

- `Analytics.Send(string name, IEnumerable<Tag> tags = null, double? value = null, IDictionary<string, object> fields = null)` و `Analytics.SendPath(string path, double? value = null)`؛ `Tag { TagGroup, Id }`.
- صف روی دیسک (`Application.persistentDataPath`) تا رویدادها با بسته شدن بازی گم نشوند؛ سقف `maxQueue`.
- ارسال دسته‌ای هر `flushSeconds` و در `OnApplicationPause(true)`؛ `batchId` تازه برای هر دسته که تا جواب ۲۰۰ ثابت می‌ماند؛ سوکت اول، HTTP پشتیبان؛ رعایت `retryAfterSeconds`.
- خواندن `insights_config` بعد از ورود و اعتبارسنجی محلی نام و تگ‌ها.
- وصل کردن به `insights_session` (شروع، پینگ، پایان) و فرستادن همان `sessionId` در دسته‌ها.
- آداپتر برای کدهایی که امروز GameAnalytics SDK صدا می‌زنند (`design`، `progression`، `resource`، `business`، `error`).
