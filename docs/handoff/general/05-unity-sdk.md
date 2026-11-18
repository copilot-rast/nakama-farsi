---
title: پکیج یونیتی
titleEn: Unity package
summary: نصب، تنظیم، شروع سریع، پرداخت و چت
---

پکیج `com.nakamafarsi.unity` بازی یونیتی را به همین سرور وصل می‌کند. نسخهٔ گسترش‌یافتهٔ کلاینت رسمی Nakama است و همه‌چیز در همین یک پکیج است. همین راهنما در صفحهٔ **SDK** کنسول هم هست.

## چه چیزهایی دارد

* ورود مهمان با شناسهٔ دستگاه و ورود پیامکی، با حفظ پیشرفت مهمان
* فراخوانی آمادهٔ همهٔ RPCهای سرور؛ اول روی سوکت و اگر سوکت قطع بود روی HTTP
* پرداخت با سکه، زرین‌پال و زیبال، و برگشت خودکار به بازی
* کافه‌بازار و مایکت: خرید، ثبت در سرور و مصرف خرید بعد از تأیید سرور
* چت متنی و استیکر، اعلان زنده و صندوق اعلان
* جدول امتیاز، وظایف، آیتم‌ها، فروشگاه و آواتار

## پیش‌نیاز

* یونیتی ۲۰۲۱.۳ یا تازه‌تر
* برنامهٔ Git روی سیستم توسعه‌دهنده. مخزن عمومی است و حساب GitHub لازم نیست

## نصب

این خط را در `Packages/manifest.json` داخل `dependencies` بگذارید:

```json
"com.nakamafarsi.unity": "https://github.com/copilot-rast/nakama-farsi-unity.git?path=/Packages/Nakama#dev"
```

اگر پکیج رسمی `com.heroiclabs.nakama-unity` نصب است، حذفش کنید. دو نسخه با هم کار نمی‌کنند.

## تنظیم

از منوی `Tools › Nakama Farsi › Create SDK Config` فایل تنظیمات را بسازید و این‌ها را پر کنید:

* آدرس سرور: آدرس API بازی، نه آدرس کنسول
* کلید سرور: از مدیر سرور بگیرید
* زبان: `fa`
* لینک برگشت پرداخت: فقط اگر زرین‌پال یا زیبال دارید، مثل `mygame://payment`

## شروع سریع

بعد از ورود، نشست ذخیره و خودکار تمدید می‌شود و سوکت خودش وصل می‌شود.

```csharp
var client = NakamaFarsiRunner.Initialize(NakamaFarsiConfig.Load());
await client.LoginWithDeviceAsync(SystemInfo.deviceUniqueIdentifier);

var boards = await client.Rpcs.Leaderboardx.ListAsync();
var tasks  = await client.Rpcs.Economy.TasksListAsync();
```

ورود پیامکی:

```csharp
var sent   = await client.Phone.RequestCodeAsync("09121234567");
var result = await client.Phone.VerifyAsync(sent.Phone, code);
```

## پرداخت

```csharp
await NakamaFarsiStore.SetupAsync(client);
var store = await NakamaFarsiStore.BuyAsync(productId);

var web = await NakamaFarsiGateway.BuyAsync(productId, PaymentMethods.ZarinPal);
var coins = await client.Payments.BuyAsync(productId, PaymentMethods.Coins);
```

* بار اول، پکیج پیشنهاد نصب Poolakey (کافه‌بازار) را می‌دهد. از منوی `Tools › Nakama Farsi › Store` هم می‌شود.
* برای عوض کردن استور بیلد: `Tools › Nakama Farsi › Store › Build for Bazaar` یا `Build for Myket`.
* وقتی پرداخت سرور آزمایشی است یا بازی روی گوشی اندروید اجرا نمی‌شود، یک استور آزمایشی جایش کار می‌کند.

## چت و استیکر

```csharp
var room = await client.Chat.JoinRoomAsync("lobby");
await client.Chat.SendTextAsync(room.Id, "salam");
await client.Chat.SendStickerAsync(room.Id, "happy/smile");
```

اگر سقف پیام روزانه پر شده باشد یا استیکر قفل باشد، سرور خطا می‌دهد.

## خطاها

هر خطای سرور یک `ApiException` است با `Code`، `Status`، `Message` و `Meta`. بعد از هر خطا، وضعیت را دوباره از سرور بخوانید.
