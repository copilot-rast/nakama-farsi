(function () {
'use strict';
var MESSAGES = {
  fa: {
    'theme.toDark': 'حالت تیره',
    'theme.toLight': 'حالت روشن',
    'section': 'اقتصاد',
    'section.short': 'اقتصاد',
    'nav.overview': 'نمای کلی',
    'nav.currencies': 'ارزها',
    'nav.metas': 'متاها',
    'nav.prizes': 'جایزه‌ها',
    'nav.products': 'فروشگاه',
    'nav.gifts': 'هدیه‌ها',
    'nav.payments': 'پرداخت‌ها',
    'tab.payment-methods': 'درگاه‌های پرداخت',
    'tab.sms': 'پیامک',
    'tab.email': 'ایمیل',
    'settings': 'تنظیمات',
    'nav.categories': 'دسته‌بندی‌ها',
    'nav.tags': 'تگ‌ها',

    'desc.overview': 'خلاصهٔ اقتصاد بازی در یک نگاه.',
    'desc.currencies': 'ارزهای بازی و مقدار اولیه‌ای که هر بازیکن می‌گیرد.',
    'desc.metas': 'مقدارهای عددی هر بازیکن، مثل سطح یا امتیاز تجربه.',
    'desc.prizes': 'بسته‌هایی از ارز، متا و آیتم که به بازیکن داده می‌شود.',
    'desc.products': 'محصولاتی که بازیکن با پول یا سکه می‌خرد.',
    'desc.gifts': 'کدهای هدیه و تخفیف.',
    'desc.payments': 'همهٔ خریدها و وضعیت هرکدام.',
    'desc.payment-methods': 'درگاه‌ها و فروشگاه‌ها را تنظیم و فعال کنید.',

    'new.currencies': 'ارز جدید',
    'new.metas': 'متای جدید',
    'new.prizes': 'جایزهٔ جدید',
    'new.products': 'محصول جدید',
    'new.gifts': 'هدیهٔ جدید',
    'new.categories': 'دسته‌بندی جدید',
    'new.tags': 'تگ جدید',
    'edit.currencies': 'ویرایش ارز',
    'edit.metas': 'ویرایش متا',
    'edit.prizes': 'ویرایش جایزه',
    'edit.products': 'ویرایش محصول',
    'edit.gifts': 'ویرایش هدیه',
    'edit.categories': 'ویرایش دسته‌بندی',
    'edit.tags': 'ویرایش تگ',

    'search': 'جستجو در کلید یا نام…',
    'search.payments': 'جستجو در شناسه، کد پیگیری یا محصول…',
    'range': '{from} تا {to} از {total}',
    'range.none': 'موردی وجود ندارد',
    'page.first': 'صفحهٔ اول',
    'page.prev': 'صفحهٔ قبل',
    'page.next': 'صفحهٔ بعد',
    'refresh': 'تازه‌سازی',
    'all': 'همه',
    'none': 'هیچ‌کدام',
    'yes': 'بله',
    'no': 'خیر',
    'create': 'ایجاد',
    'save': 'ذخیره',
    'saving': 'در حال ذخیره…',
    'cancel': 'انصراف',
    'close': 'بستن',
    'delete': 'حذف',
    'delete.confirm': 'حذف شود؟',
    'delete.yes': 'بله، حذف شود',
    'saved': 'ذخیره شد.',
    'deleted': 'حذف شد.',
    'back': 'بازگشت',
    'add.row': 'افزودن ردیف',
    'remove': 'حذف',
    'upload': 'بارگذاری',
    'uploading': 'در حال بارگذاری…',
    'choose': 'انتخاب کنید',
    'search.short': 'جستجو…',
    'no.results': 'نتیجه‌ای پیدا نشد.',
    'loading': 'در حال بارگذاری…',
    'optional': 'اختیاری',

    'empty.title': 'هنوز چیزی اینجا نیست',
    'empty.search': 'با این جستجو چیزی پیدا نشد.',
    'empty.hint.currencies': 'اولین ارز بازی را بسازید، مثلاً سکه یا الماس.',
    'empty.hint.metas': 'اولین متا را بسازید، مثلاً سطح بازیکن.',
    'empty.hint.prizes': 'جایزه‌ها را برای محصولات و هدیه‌ها لازم دارید.',
    'empty.hint.products': 'برای ساخت محصول، اول یک جایزه بسازید.',
    'empty.hint.gifts': 'کد هدیه یا تخفیف بسازید و به بازیکنان بدهید.',
    'empty.hint.payments': 'هنوز خریدی انجام نشده است.',
    'empty.hint.categories': 'دسته‌بندی‌ها به مرتب‌شدن فروشگاه کمک می‌کنند.',
    'empty.hint.tags': 'تگ‌ها برای گروه‌بندی و فیلتر به کار می‌روند.',

    'err.network': 'اتصال به سرور برقرار نشد. دوباره تلاش کنید.',
    'err.forbidden': 'دسترسی به این بخش را ندارید.',
    'err.notfound': 'این بخش روی سرور فعال نیست. ماژول مربوط بارگذاری شده است؟',
    'err.item.notfound': 'این مورد پیدا نشد.',
    'err.generic': 'خطایی رخ داد. دوباره تلاش کنید.',
    'err.load': 'بارگذاری انجام نشد.',
    'err.required': 'این فیلد لازم است.',
    'err.key': 'فقط حروف کوچک انگلیسی، عدد و _ (۲ تا ۳۲ کاراکتر).',
    'err.key.general': 'با حرف کوچک انگلیسی یا عدد شروع کنید؛ بعد حرف، عدد، _ . یا - (تا ۶۴ کاراکتر).',
    'err.code': 'کد باید ۳ تا ۳۲ کاراکتر از A تا Z، عدد، _ یا - باشد.',
    'err.number': 'یک عدد معتبر وارد کنید.',
    'err.json': 'JSON معتبر نیست.',
    'err.form': 'چند فیلد نیاز به اصلاح دارد.',
    'err.upload': 'فایل بارگذاری نشد. فقط png، jpg، webp یا svg تا ۲ مگابایت.',
    'try.again': 'تلاش دوباره',

    'code.IN_USE': 'این مورد جای دیگری استفاده شده و حذف نمی‌شود.',
    'code.PAYMENT_METHOD_NOT_CONFIGURED': 'برای فعال‌کردن، اول فیلدهای لازم را پر کنید.',
    'code.DUPLICATE_KEY': 'این کلید قبلاً استفاده شده است.',
    'code.ALREADY_EXISTS': 'این مقدار قبلاً استفاده شده است.',
    'code.NOT_FOUND': 'پیدا نشد.',
    'code.INVALID_STATE': 'در این وضعیت این کار ممکن نیست.',
    'code.KEY_IMMUTABLE': 'کلید بعد از ساخت عوض نمی‌شود.',
    'code.REQUIRED': 'این فیلد لازم است.',
    'code.INVALID_KEY': 'این کلید معتبر نیست.',
    'code.INVALID_CODE': 'کد باید ۳ تا ۳۲ کاراکتر از A تا Z، عدد، _ یا - باشد.',
    'code.CODE_REQUIRED': 'هدیهٔ تخفیفی کد لازم دارد.',
    'code.ITEMS_REQUIRED': 'دست‌کم یک آیتم اضافه کنید.',
    'code.TOO_MANY_ITEMS': 'حداکثر ۵۰ آیتم مجاز است.',
    'code.INVALID_VALUE': 'این مقدار معتبر نیست.',
    'code.INVALID_ACTION': 'این عمل برای این نوع آیتم مجاز نیست.',
    'code.INVALID_ITEM_TYPE': 'نوع آیتم معتبر نیست.',
    'code.CURRENCY_NOT_FOUND': 'این ارز وجود ندارد.',
    'code.META_NOT_FOUND': 'این متا وجود ندارد.',
    'code.PRIZE_NOT_FOUND': 'این جایزه وجود ندارد.',
    'code.CATEGORY_NOT_FOUND': 'این دسته‌بندی وجود ندارد.',
    'code.TAG_NOT_FOUND': 'یکی از تگ‌ها دیگر وجود ندارد.',
    'code.REFERENCE_NOT_FOUND': 'یکی از موارد انتخاب‌شده دیگر وجود ندارد.',
    'code.PRICE_REQUIRED': 'درگاه‌ها قیمت دست‌کم ۱۰۰ تومان لازم دارند.',
    'code.PRICE_TOO_LOW': 'قیمت بعد از تخفیف باید صفر یا دست‌کم ۱۰۰ تومان باشد.',
    'code.INVALID_PRICE': 'قیمت معتبر نیست.',
    'code.COIN_PRICE_REQUIRED': 'برای پرداخت با سکه، قیمت سکه لازم است.',
    'code.SKU_REQUIRED': 'SKU این فروشگاه لازم است.',
    'code.INVALID_SKU': 'SKU معتبر نیست.',
    'code.SUBSCRIPTION_NOT_SUPPORTED': 'مایکت اشتراک را پشتیبانی نمی‌کند.',
    'code.INVALID_DISCOUNT': 'درصد تخفیف معتبر نیست.',
    'code.INVALID_LIMIT': 'سقف باید دست‌کم ۱ باشد.',
    'code.INVALID_WINDOW': 'پایان اعتبار باید بعد از شروع باشد.',
    'code.INVALID_TIME': 'زمان معتبر نیست.',
    'code.INVALID_URL': 'نشانی باید با http یا https یا / شروع شود.',
    'code.INVALID_OBJECT': 'باید یک شیء JSON باشد.',
    'code.OBJECT_TOO_LARGE': 'این JSON بیش از حد بزرگ است.',
    'code.INVALID_NAME': 'نام لازم است (تا ۱۲۸ کاراکتر).',
    'code.INVALID_KIND': 'کاربرد معتبر نیست.',
    'code.INVALID_TYPE': 'نوع معتبر نیست.',
    'code.INVALID_PAYMENT_METHOD': 'روش پرداخت ناشناخته است.',
    'code.PAYMENT_NOT_PENDING': 'فقط پرداخت‌های در انتظار استعلام می‌شوند.',
    'code.REWARD_ALREADY_GRANTED': 'جایزه قبلاً داده شده است.',
    'code.PAYMENT_NOT_SUCCESSFUL': 'فقط برای پرداخت موفق می‌شود جایزه داد.',
    'code.PAYMENT_NOT_FOUND': 'این پرداخت پیدا نشد.',
    'code.PAYMENT_METHOD_NOT_FOUND': 'این روش پرداخت پیدا نشد.',
    'code.GATEWAY_REQUEST_FAILED': 'درگاه پاسخ نداد. کمی بعد دوباره امتحان کنید.',
    'code.STORE_UNAVAILABLE': 'فروشگاه در دسترس نیست. دوباره امتحان کنید.',
    'code.STORE_CONFIGURATION_REJECTED': 'فروشگاه تنظیمات سرور را نپذیرفت.',
    'code.UNSUPPORTED_FILE_TYPE': 'فقط png، jpg، webp یا svg.',
    'code.FILE_TOO_LARGE': 'حجم فایل حداکثر ۲ مگابایت است.',

    'f.key': 'کلید',
    'f.key.hint': 'شناسهٔ ثابت در کد بازی؛ بعداً عوضش نکنید.',
    'f.key.locked': 'کلید بعد از ساخت عوض نمی‌شود.',
    'f.type': 'نوع',
    'f.sort': 'ترتیب',
    'f.sort.hint': 'عدد کمتر، بالاتر نمایش داده می‌شود.',
    'f.icon': 'آیکون',
    'f.cover': 'تصویر کاور',
    'f.metadata': 'متادیتا (JSON)',
    'f.settings': 'تنظیمات (JSON)',
    'f.defaultValue': 'مقدار اولیه',
    'f.defaultValue.currency.hint': 'یک بار به هر بازیکن داده می‌شود.',
    'f.defaultValue.meta.hint': 'اگر بازیکن مقداری نداشته باشد، همین حساب می‌شود.',
    'f.serverOnly': 'فقط سرور',
    'f.serverOnly.hint': 'بازی این مقدار را نمی‌بیند.',
    'f.category': 'دسته‌بندی',
    'f.tags': 'تگ‌ها',
    'f.isRepeatable': 'تکرارپذیر',
    'f.isRepeatable.hint': 'اگر خاموش باشد، هر بازیکن فقط یک بار این جایزه را می‌گیرد.',
    'f.items': 'آیتم‌های جایزه',
    'f.items.empty': 'هنوز آیتمی اضافه نشده است.',
    'f.item.type': 'نوع',
    'f.item.key': 'کلید',
    'f.item.action': 'عمل',
    'f.item.value': 'مقدار',
    'f.isActive': 'فعال',
    'f.priceToman': 'قیمت (تومان)',
    'f.priceToman.hint': 'برای درگاه‌ها دست‌کم ۱۰۰ تومان.',
    'f.coinPrice': 'قیمت با سکه',
    'f.coinPrice.currency': 'ارز',
    'f.coinPrice.amount': 'مقدار',
    'f.discountPercent': 'درصد تخفیف',
    'f.paymentMethods': 'روش‌های پرداخت',
    'f.sku.bazaar': 'SKU بازار',
    'f.sku.myket': 'SKU مایکت',
    'f.isSubscription': 'اشتراکی',
    'f.isSubscription.hint': 'فقط در بازار پشتیبانی می‌شود.',
    'f.prize': 'جایزه',
    'f.prize.product.hint': 'بعد از پرداخت موفق به بازیکن داده می‌شود.',
    'f.minLevel': 'حداقل سطح',
    'f.gift.name': 'نام',
    'f.code': 'کد',
    'f.code.hint': 'با حروف بزرگ ذخیره می‌شود. خالی بگذارید تا کدی نداشته باشد.',
    'f.isPromo': 'تبلیغاتی',
    'f.perUserLimit': 'سقف برای هر بازیکن',
    'f.maxTotalUses': 'سقف کل استفاده',
    'f.maxTotalUses.hint': 'خالی یعنی بدون سقف.',
    'f.totalUsed': 'استفاده‌شده',
    'f.validFrom': 'شروع اعتبار',
    'f.expireAt': 'پایان اعتبار',
    'f.uploadUrl': 'نشانی فایل',

    'col.key': 'کلید',
    'col.name': 'نام',
    'col.type': 'نوع',
    'col.default': 'مقدار اولیه',
    'col.sort': 'ترتیب',
    'col.items': 'آیتم‌ها',
    'col.price': 'قیمت',
    'col.methods': 'روش‌ها',
    'col.status': 'وضعیت',
    'col.code': 'کد',
    'col.usage': 'استفاده',
    'col.expire': 'پایان اعتبار',
    'col.updated': 'به‌روزرسانی',
    'col.created': 'زمان',
    'col.user': 'کاربر',
    'col.product': 'محصول',
    'col.method': 'روش',
    'col.amount': 'مبلغ',
    'col.reward': 'جایزه',

    'type.soft': 'نرم',
    'type.hard': 'سخت',
    'kind.product': 'محصول',
    'kind.prize': 'جایزه',
    'kind.gift': 'هدیه',
    'kind.currency': 'ارز',
    'gift.reward': 'جایزه',
    'gift.discount': 'تخفیف',
    'item.currency': 'ارز',
    'item.meta': 'متا',
    'item.storage': 'آیتم ذخیره‌سازی',
    'action.add': 'افزودن',
    'action.set': 'تنظیم',
    'action.give': 'دادن',
    'active': 'فعال',
    'inactive': 'غیرفعال',
    'repeatable': 'تکرارپذیر',
    'once': 'یک‌باره',
    'server.only': 'فقط سرور',
    'unlimited': 'بی‌سقف',
    'toman': '{n} تومان',
    'percent': '{n}٪',
    'amount.of': '{n} {name}',

    'method.zarinpal': 'زرین‌پال',
    'method.zibal': 'زیبال',
    'method.bazaar': 'بازار',
    'method.myket': 'مایکت',
    'method.in_app': 'سکهٔ بازی',
    'status.pending': 'در انتظار',
    'status.success': 'موفق',
    'status.failed': 'ناموفق',
    'status.expired': 'منقضی',

    'filter.status': 'وضعیت',
    'filter.method': 'روش',
    'filter.user': 'شناسهٔ کاربر',
    'filter.active': 'وضعیت',

    'pay.title': 'پرداخت',
    'pay.amounts': 'مبلغ',
    'pay.original': 'قیمت اصلی',
    'pay.final': 'مبلغ نهایی',
    'pay.rial': 'مبلغ به ریال',
    'pay.coin': 'پرداخت با سکه',
    'pay.discount': 'تخفیف',
    'pay.gateway': 'درگاه و فروشگاه',
    'pay.sandbox': 'آزمایشی',
    'pay.authority': 'Authority',
    'pay.trackId': 'Track ID',
    'pay.refId': 'کد پیگیری',
    'pay.storeToken': 'توکن فروشگاه',
    'pay.cardMask': 'کارت',
    'pay.outcome': 'نتیجه',
    'pay.error': 'خطا',
    'pay.reward': 'جایزه',
    'pay.reward.granted': 'داده شد',
    'pay.reward.notGranted': 'داده نشده',
    'pay.reward.error': 'خطای جایزه',
    'pay.inquiries': 'تعداد استعلام',
    'pay.times': 'زمان‌ها',
    'pay.createdAt': 'ایجاد',
    'pay.updatedAt': 'آخرین تغییر',
    'pay.verifiedAt': 'تأیید',
    'pay.inquiredAt': 'آخرین استعلام',
    'pay.user': 'بازیکن',
    'pay.product': 'محصول',
    'pay.logs': 'گزارش تبادل با درگاه',
    'pay.logs.empty': 'گزارشی ثبت نشده است.',
    'pay.inquire': 'استعلام دوباره',
    'pay.inquire.done': 'استعلام انجام شد.',
    'pay.retry': 'دادن دوبارهٔ جایزه',
    'pay.retry.done': 'جایزه دوباره داده شد.',
    'pay.request': 'درخواست',
    'pay.response': 'پاسخ',
    'pay.duration': '{n} میلی‌ثانیه',
    'log.request': 'درخواست',
    'log.callback': 'بازگشت از درگاه',
    'log.verify': 'تأیید',
    'log.inquiry': 'استعلام',
    'log.store_validate': 'بررسی فروشگاه',

    'pm.configured': 'تنظیم شده',
    'pm.notConfigured': 'تنظیم نشده',
    'pm.active': 'فعال',
    'pm.noFields': 'این روش تنظیمی ندارد؛ فقط فعال یا غیرفعالش کنید.',
    'pm.secret.hint': 'برای نگه‌داشتن مقدار فعلی، دست نزنید.',
    'pm.empty': 'هیچ روش پرداختی از سرور برنگشت.',
    'pm.field.merchantId': 'کد مرچنت',
    'pm.field.merchant': 'کد مرچنت',
    'pm.field.sandbox': 'حالت آزمایشی',
    'pm.field.packageName': 'نام بسته',
    'pm.field.apiSecret': 'کلید API',
    'pm.field.accessToken': 'توکن دسترسی',
    'pm.field.rsaPublicKey': 'کلید عمومی RSA',

    'ov.catalog': 'کاتالوگ',
    'ov.revenue30': 'درآمد ۳۰ روز اخیر',
    'ov.paymentsToday': 'پرداخت‌های امروز',
    'ov.grantErrors': 'جایزه‌های داده‌نشده',
    'ov.methods': 'درگاه‌های پرداخت',
    'ov.sandbox': 'آزمایشی',
    'ov.payments': 'پرداخت‌ها',
    'ov.revenue': 'درآمد',
    'ov.open': 'مشاهده',
    'ov.today': 'امروز',
    'ov.total': 'کل',

    'section.pub': 'انتشار',
    'section.pub.short': 'انتشار',
    'nav.faqs': 'سؤال‌های متداول',
    'nav.news': 'اخبار',
    'nav.comments': 'کامنت‌ها',
    'nav.notifications': 'اعلان‌ها',
    'desc.faqs': 'پرسش و پاسخ‌هایی که بازیکن در بخش راهنما می‌بیند.',
    'desc.news': 'خبرها و اطلاعیه‌هایی که در بازی نمایش داده می‌شوند.',
    'desc.comments': 'کامنت‌های بازیکنان روی خبرها را بررسی و تأیید کنید.',
    'desc.notifications': 'اعلان درون بازی، پیامک و ایمیل به بازیکنان بفرستید.',
    'desc.sms': 'سرویس‌های پیامک را تنظیم کنید. در حالت آزمایشی هیچ پیامکی واقعاً فرستاده نمی‌شود.',
    'desc.email': 'سرویس‌های ایمیل را تنظیم کنید. در حالت آزمایشی هیچ ایمیلی واقعاً فرستاده نمی‌شود.',
    'new.faqs': 'سؤال جدید',
    'new.news': 'خبر جدید',
    'edit.faqs': 'ویرایش سؤال',
    'edit.news': 'ویرایش خبر',
    'search.faqs': 'جستجو در سؤال یا پاسخ…',
    'search.news': 'جستجو در عنوان یا خلاصه…',
    'search.comments': 'جستجو در متن کامنت، نام کاربری یا خبر…',
    'search.notifications': 'جستجو در عنوان، متن یا شناسه…',
    'empty.hint.faqs': 'سؤال‌های پرتکرار بازیکنان را با پاسخشان اینجا بنویسید.',
    'empty.hint.news': 'اولین خبر را بنویسید تا در بازی دیده شود.',
    'empty.hint.comments': 'کامنتی برای بررسی نیست.',
    'empty.hint.notifications': 'هنوز اعلانی نفرستاده‌اید.',

    'kind.faq': 'سؤال متداول',
    'kind.news': 'خبر',
    'filter.category': 'دسته‌بندی',
    'filter.pinned': 'سنجاق',
    'filter.news': 'خبر',
    'filter.channel': 'کانال',

    'col.question': 'سؤال',
    'col.category': 'دسته‌بندی',
    'col.title': 'عنوان',
    'col.comments': 'کامنت‌ها',
    'col.published': 'زمان انتشار',
    'col.comment': 'کامنت',
    'col.news': 'خبر',
    'col.actions': 'کارها',
    'col.notification': 'اعلان',
    'col.channel': 'کانال',
    'col.target': 'گیرندگان',
    'col.delivery': 'ارسال',
    'col.address': 'نشانی',
    'col.error': 'خطا',
    'col.attempts': 'تلاش',
    'col.time': 'زمان',

    'f.faq.active.hint': 'سؤال غیرفعال در بازی دیده نمی‌شود.',
    'f.publishedAt': 'زمان انتشار',
    'f.publishedAt.hint': 'خالی یعنی همین حالا. اگر زمانی در آینده بدهید، خبر از همان موقع دیده می‌شود.',
    'f.news.active.hint': 'خبر غیرفعال در بازی دیده نمی‌شود.',
    'f.isPinned': 'سنجاق‌شده',
    'f.isPinned.hint': 'بالای فهرست خبرها می‌ماند.',
    'f.comments': 'کامنت‌ها',
    'f.allowComment': 'کامنت باز است',
    'f.allowComment.hint': 'بازیکنان می‌توانند زیر این خبر کامنت بگذارند.',
    'f.freeComment': 'انتشار بدون بررسی',
    'f.freeComment.hint': 'اگر خاموش باشد، هر کامنت اول باید در صفحهٔ کامنت‌ها تأیید شود.',
    'f.actions': 'دکمه‌های خبر',

    'md.write': 'نوشتن',
    'md.preview': 'پیش‌نمایش',
    'md.mode': 'حالت ویرایشگر',
    'md.empty': 'متنی برای پیش‌نمایش نیست.',
    'md.hint': 'قالب‌بندی:',
    'chars': '{n} از {max} نویسه',

    'act.add': 'افزودن دکمه',
    'act.empty': 'دکمه‌ای اضافه نشده است.',
    'act.type': 'نوع',
    'act.url': 'باز کردن پیوند',
    'act.goto': 'رفتن به بخشی از بازی',
    'act.none': 'بدون دکمه',
    'act.value': 'مقصد',
    'act.url.hint': 'پیوند در زبانهٔ تازه باز می‌شود.',
    'act.goto.hint': 'نام بخش بازی، مثل shop یا play.',

    'news.pinned': 'سنجاق‌شده',
    'news.notPinned': 'بدون سنجاق',
    'news.scheduled': 'زمان‌بندی‌شده',
    'news.commentsOff': 'بسته',

    'cm.pending': 'در انتظار تأیید',
    'cm.accepted': 'تأییدشده',
    'cm.pendingN': '{n} در انتظار',
    'cm.accept': 'تأیید',
    'cm.reject': 'رد کردن',
    'cm.reject.hint': 'کامنت از دید بازیکنان پنهان می‌شود.',
    'cm.accepted.done': 'کامنت تأیید شد.',
    'cm.rejected.done': 'کامنت رد شد.',
    'cm.reply': 'پاسخ',
    'cm.replies': '{n} پاسخ',

    'ch.in_app': 'درون بازی',
    'ch.sms': 'پیامک',
    'ch.email': 'ایمیل',
    'nst.queued': 'در صف',
    'nst.sending': 'در حال ارسال',
    'nst.done': 'انجام شد',
    'nst.failed': 'ناموفق',
    'rst.pending': 'در انتظار',
    'rst.sent': 'ارسال‌شده',
    'rst.skipped': 'بدون نشانی',
    'rst.failed': 'ناموفق',

    'tg.all': 'همهٔ کاربران',
    'tg.users': 'کاربران مشخص',
    'tg.group': 'یک گروه',
    'tg.usersN': '{n} کاربر',
    'tg.pickUsers': 'کاربران',
    'tg.searchUsers': 'جستجو و افزودن کاربر…',
    'tg.searchUsers.hint': 'نام کاربری، نام نمایشی یا شناسه را بنویسید.',
    'tg.users.hint': 'تا ۱۰۰۰ کاربر.',
    'tg.pickGroup': 'انتخاب گروه',
    'tg.group.hint': 'به همهٔ اعضای گروه فرستاده می‌شود، نه درخواست‌های عضویت.',
    'tg.all.hint': 'به همهٔ کاربرانی که حسابشان غیرفعال نشده فرستاده می‌شود.',
    'tg.none': 'هنوز گیرنده‌ای انتخاب نشده است.',
    'tg.withPhone': '{n} نفر شماره دارند',
    'tg.withEmail': '{n} نفر ایمیل دارند',
    'tg.skipped': '{n} نفر نشانی ندارند و کنار گذاشته می‌شوند.',

    'nt.title': 'اعلان',
    'nt.new': 'ارسال اعلان',
    'nt.new.desc': 'پیام را بنویسید، گیرندگان را انتخاب کنید و بفرستید.',
    'nt.s.channel': 'کانال',
    'nt.s.content': 'متن پیام',
    'nt.s.action': 'دکمهٔ اعلان',
    'nt.s.target': 'گیرندگان',
    'nt.summary': 'خلاصه',
    'nt.f.title': 'عنوان',
    'nt.f.body': 'متن',
    'nt.f.provider': 'سرویس ارسال',
    'nt.sms.hint': 'پیامک عنوان ندارد. متن ساده بنویسید.',
    'nt.sms.tooLong': 'متن پیامک حداکثر ۶۰۰ نویسه است.',
    'nt.action.hint': 'اختیاری. با زدن اعلان، پیوند باز می‌شود یا بازیکن به آن بخش بازی می‌رود.',
    'nt.inapp.hint': 'در صندوق پیام بازیکن ذخیره می‌شود و اگر بازی باز باشد، همان لحظه نشان داده می‌شود.',
    'nt.noProvider': 'سرویس فعالی برای این کانال نیست.',
    'nt.goSettings.sms': 'تنظیمات پیامک',
    'nt.goSettings.email': 'تنظیمات ایمیل',
    'nt.sandbox.note': 'این سرویس در حالت آزمایشی است؛ پیامی واقعاً فرستاده نمی‌شود.',
    'nt.live.note': 'این سرویس آزمایشی نیست؛ پیام واقعاً برای کاربران فرستاده می‌شود.',
    'nt.recipientsWord': 'گیرنده',
    'nt.send': 'ارسال اعلان',
    'nt.sending': 'در حال ارسال…',
    'nt.queued': 'اعلان در صف ارسال قرار گرفت.',
    'nt.confirmAll.title': 'برای همه فرستاده شود؟',
    'nt.confirmAll.text': 'این اعلان برای همهٔ {n} کاربر فرستاده می‌شود و پس‌گرفتنی نیست.',
    'nt.confirmAll.textNoCount': 'این اعلان برای همهٔ کاربران فرستاده می‌شود و پس‌گرفتنی نیست.',
    'nt.confirmAll.ok': 'بله، برای همه بفرست',
    'nt.confirmLive.title': 'ارسال واقعی',
    'nt.confirmLive.text': 'سرویس انتخاب‌شده آزمایشی نیست و پیام واقعاً فرستاده می‌شود.',
    'nt.info': 'مشخصات',
    'nt.content': 'متن',
    'nt.delivery': 'وضعیت ارسال',
    'nt.recipientsList': 'گیرندگان',
    'nt.recipients': 'گیرندگان',
    'nt.seen': 'دیده‌شده',
    'nt.live': 'خودکار تازه می‌شود',
    'nt.action': 'دکمه',
    'nt.createdBy': 'فرستنده',
    'nt.noRecipients': 'گیرنده‌ای ندارد.',
    'nt.sentAt': 'ارسال: {t}',
    'nt.seenAt': 'دیده شد: {t}',
    'nt.sentOf': '{sent} از {total}',
    'nt.skippedN': '{n} بی‌نشانی',
    'nt.failedN': '{n} ناموفق',

    'pv.kind': 'نوع سرویس',
    'pv.kind.kavenegar': 'کاوه‌نگار',
    'pv.kind.smsir': 'sms.ir',
    'pv.kind.twilio': 'Twilio',
    'pv.kind.smtp': 'SMTP',
    'pv.kind.resend': 'Resend',
    'pv.add': 'افزودن سرویس',
    'pv.added': 'سرویس اضافه شد. حالت آزمایشی روشن است.',
    'pv.allAdded': 'همهٔ سرویس‌های این کانال اضافه شده‌اند.',
    'pv.default': 'پیش‌فرض',
    'pv.empty.sms': 'هنوز سرویس پیامکی ندارید.',
    'pv.empty.email': 'هنوز سرویس ایمیلی ندارید.',
    'pv.empty.hint': 'از فهرست بالا یک سرویس اضافه کنید.',
    'pv.sandbox': 'حالت آزمایشی',
    'pv.sandbox.short': 'آزمایشی',
    'pv.sandbox.on': 'روشن است: هیچ پیامی واقعاً فرستاده نمی‌شود و فقط در گزارش سرور ثبت می‌شود.',
    'pv.sandbox.off': 'خاموش است: پیام‌ها واقعاً برای کاربران فرستاده می‌شوند.',
    'pv.live.title': 'حالت آزمایشی خاموش شود؟',
    'pv.live.text': 'از این به بعد پیام‌های این سرویس واقعاً برای کاربران فرستاده می‌شوند و هزینه دارند.',
    'pv.live.ok': 'بله، خاموش شود',
    'pv.field.apiKey': 'کلید API',
    'pv.field.sender': 'شمارهٔ فرستنده',
    'pv.field.lineNumber': 'شمارهٔ خط',
    'pv.field.accountSid': 'Account SID',
    'pv.field.authToken': 'Auth Token',
    'pv.field.from': 'فرستنده',
    'pv.field.host': 'میزبان',
    'pv.field.port': 'پورت',
    'pv.field.security': 'رمزگذاری',
    'pv.field.username': 'نام کاربری',
    'pv.field.password': 'رمز عبور',
    'pv.field.fromName': 'نام فرستنده',
    'pv.hint.host': 'فقط نام میزبان، بدون http و پورت.',
    'pv.hint.from': 'شماره یا نشانی‌ای که پیام از آن فرستاده می‌شود.',
    'pv.opt.starttls': 'STARTTLS',
    'pv.opt.tls': 'TLS',
    'pv.opt.none': 'بدون رمزگذاری',
    'pv.test': 'ارسال آزمایشی',
    'pv.test.hint': 'با تنظیمات ذخیره‌شده یک پیام می‌فرستد. در حالت آزمایشی چیزی ارسال نمی‌شود.',
    'pv.test.to': 'گیرنده',
    'pv.test.subject': 'موضوع',
    'pv.test.subject.ph': 'پیام آزمایشی',
    'pv.test.message': 'متن',
    'pv.test.message.default': 'این یک پیام آزمایشی است.',
    'pv.test.send': 'ارسال آزمایشی',
    'pv.test.sandbox': 'حالت آزمایشی: چیزی ارسال نشد. پاسخ:',
    'pv.test.ok': 'ارسال شد.',
    'pv.test.failed': 'ارسال نشد:',

    'code.TARGET_USERS_REQUIRED': 'دست‌کم یک کاربر انتخاب کنید.',
    'code.USERS_NOT_FOUND': 'بعضی از کاربران انتخاب‌شده وجود ندارند.',
    'code.TOO_MANY_USERS': 'حداکثر ۱۰۰۰ کاربر.',
    'code.GROUP_NOT_FOUND': 'این گروه پیدا نشد.',
    'code.INVALID_TARGET': 'گیرندگان معتبر نیستند.',
    'code.INVALID_CHANNEL': 'کانال معتبر نیست.',
    'code.PROVIDER_NOT_FOUND': 'این سرویس پیدا نشد.',
    'code.PROVIDER_NOT_CONFIGURED': 'سرویس فعال و کاملی برای این کانال نیست؛ در تنظیمات پروژه، فیلدهای لازم را پر کنید.',
    'code.PROVIDER_INACTIVE': 'این سرویس فعال نیست.',
    'code.PROVIDER_CHANNEL_MISMATCH': 'این سرویس برای این کانال نیست.',
    'code.PROVIDER_KIND_FIXED': 'نوع سرویس بعد از ساخت عوض نمی‌شود.',
    'code.INVALID_PROVIDER_KIND': 'این نوع سرویس شناخته‌شده نیست.',
    'code.INVALID_PHONE': 'شماره را به شکل ۰۹۱۲۳۴۵۶۷۸۹ یا با کد کشور بنویسید.',
    'code.INVALID_EMAIL': 'نشانی ایمیل معتبر نیست.',
    'code.INVALID_MESSAGE': 'متن لازم است (حداکثر ۶۰۰ نویسه).',
    'code.TOO_MANY_ACTIONS': 'حداکثر ۵ دکمه.',
    'code.COMMENT_NOT_FOUND': 'این کامنت پیدا نشد.',
    'code.NOTIFICATION_NOT_FOUND': 'این اعلان پیدا نشد.',
    'code.INVALID_FILTER': 'فیلتر معتبر نیست.',
    'code.QUERY_TOO_LONG': 'متن جستجو بیش از حد بلند است.',
    'code.URL_TOO_LONG': 'نشانی بیش از حد بلند است.',
    'code.PROVIDER_NOT_ALLOWED': 'اعلان درون بازی سرویس ارسال ندارد.',
    'section.base': 'عمومی',
    'section.base.short': 'عمومی',
    'desc.categories': 'دسته‌بندی‌های همهٔ بخش‌ها؛ هر دسته‌بندی برای یک نوع کلید است، مثل محصول یا خبر.',
    'desc.tags': 'برچسب‌هایی برای گروه‌بندی؛ تگ «عمومی» همه‌جا قابل استفاده است.',
    'tab.languages': 'زبان‌ها',
    'desc.languages': 'زبان‌هایی که محتوای بازی به آن‌ها ترجمه می‌شود. زبان پیش‌فرض همیشه فعال است.',
    'new.languages': 'زبان جدید',
    'edit.languages': 'ویرایش زبان',
    'search.languages': 'جستجو در کد یا نام زبان…',
    'empty.hint.languages': 'اولین زبان را اضافه کنید.',
    'filter.keyType': 'نوع کلید',
    'col.keyType': 'نوع کلید',
    'f.keyType': 'نوع کلید',
    'f.keyType.category.hint': 'مشخص می‌کند این دسته‌بندی در کدام بخش دیده شود.',
    'f.keyType.tag.hint': 'تگ «عمومی» برای همهٔ بخش‌ها قابل انتخاب است.',
    'kind.general': 'عمومی',
    'kind.user': 'بازیکن',
    'ps.title': 'تنظیمات پروژه',
    'ps.sections': 'بخش‌های تنظیمات پروژه',
    'ps.noAccess': 'به هیچ بخشی از تنظیمات پروژه دسترسی ندارید.',
    'ps.noAccess.hint': 'از مدیر کنسول بخواهید دسترسی لازم را برایتان باز کند.',
    'qc.add.categories': 'دسته‌بندی تازه',
    'qc.add.tags': 'تگ تازه',
    'qc.name': 'نام',
    'qc.name.ph.categories': 'مثلاً بسته‌های ویژه',
    'qc.name.ph.tags': 'مثلاً پرفروش',
    'qc.hint': 'برای «{type}»، به {lang}. کلید خودکار ساخته می‌شود.',
    'qc.create': 'بساز',
    'qc.err.name': 'یک نام بنویسید.',
    'qc.created': '«{name}» ساخته و انتخاب شد.',
    'f.name': 'نام',
    'f.message': 'پیام',
    'f.description': 'توضیح',
    'f.question': 'سؤال',
    'f.answer': 'پاسخ',
    'f.title': 'عنوان',
    'f.summary': 'خلاصه',
    'f.body': 'متن',
    'act.label': 'متن دکمه',
    'nt.f.sms': 'متن پیامک',
    'nt.action.label.hint': 'اگر خالی بماند، بازی متن خودش را نشان می‌دهد.',
    'tr.default': 'پیش‌فرض',
    'tr.empty': 'خالی',
    'lang.default': 'پیش‌فرض',
    'lang.activated': '{name} فعال شد.',
    'lang.deactivated': '{name} غیرفعال شد.',
    'col.language': 'زبان',
    'col.direction': 'جهت نوشتن',
    'dir.rtl': 'راست به چپ',
    'dir.ltr': 'چپ به راست',
    'f.lang.code': 'کد زبان',
    'f.lang.code.hint': 'کد کوتاه با حروف کوچک، مثل fa، en، ar یا pt-br. بعد از ساخت عوض نمی‌شود.',
    'f.lang.nativeName': 'نام بومی',
    'f.lang.nativeName.hint': 'نام زبان به خط خودش، مثل العربية.',
    'f.lang.name': 'نام ترجمه‌شده',
    'f.lang.name.hint': 'نام زبان در هر زبان دیگر؛ مقدار زبان پیش‌فرض لازم است.',
    'f.lang.direction': 'جهت نوشتن',
    'f.lang.active.hint': 'فقط زبان‌های فعال در فرم‌ها و بازی دیده می‌شوند.',
    'f.lang.default': 'زبان پیش‌فرض',
    'f.lang.default.hint': 'متن پیش‌فرض همه‌جا لازم است و وقتی ترجمه نباشد نشان داده می‌شود.',
    'cov.title': 'پوشش ترجمه',
    'cov.desc': 'تعداد مواردی که متنی دارند ولی در هر زبان ترجمه نشده‌اند. روی هر عدد بزنید تا به همان صفحه بروید.',
    'cov.resource': 'بخش',
    'cov.total': 'کل',
    'cov.missingN': '{n} ناقص',
    'cov.complete': 'کامل',
    'cov.open': 'باز کردن صفحه',
    'cov.module.base': 'عمومی',
    'cov.module.economy': 'اقتصاد',
    'cov.module.publication': 'انتشار',
    'code.DEFAULT_LANGUAGE_REQUIRED': 'متن به زبان پیش‌فرض ({lang}) لازم است.',
    'code.INVALID_LANGUAGE': 'یکی از زبان‌ها دیگر وجود ندارد؛ فرم را ببندید و دوباره باز کنید.',
    'code.TEXT_TOO_LONG': 'متن {lang} بیش از حد بلند است.',
    'code.CATEGORY_KIND_MISMATCH': 'نوع کلید این دسته‌بندی با این بخش جور نیست.',
    'code.TAG_KIND_MISMATCH': 'یکی از تگ‌ها برای این بخش نیست؛ فقط تگ همین بخش یا عمومی را انتخاب کنید.',
    'code.INVALID_KEY_TYPE': 'نوع کلید معتبر نیست.',
    'code.INVALID_LANGUAGE_CODE': 'کد زبان را با حروف کوچک بنویسید، مثل fa، en یا pt-br.',
    'code.LANGUAGE_NOT_FOUND': 'این زبان پیدا نشد.',
    'code.DEFAULT_LANGUAGE_LOCKED': 'زبان پیش‌فرض حذف یا غیرفعال نمی‌شود؛ اول زبان دیگری را پیش‌فرض کنید.',
    'code.LANGUAGE_ACTIVE': 'زبان فعال حذف نمی‌شود؛ اول آن را غیرفعال کنید.',
    'code.INVALID_DIRECTION': 'جهت نوشتن باید راست به چپ یا چپ به راست باشد.',
    'code.INVALID_NATIVE_NAME': 'نام بومی لازم است (تا ۶۴ نویسه).',
    'code.DEFAULT_MUST_BE_ACTIVE': 'فقط مورد فعال می‌تواند پیش‌فرض باشد.',
    'code.LANGUAGE_NOT_AVAILABLE': 'این زبان در دسترس نیست.',
    'tab.brand': 'برند',
    'tab.setup': 'راه‌اندازی',
    'desc.brand': 'نام، لوگو و رنگ اصلی کنسول. صفحهٔ ورود و بازی‌ها هم همین را نشان می‌دهند.',
    'desc.setup': 'راهنمای راه‌اندازی کنسول را هر وقت خواستید از اینجا باز کنید.',
    'brand.signIn': 'برای ادامه به داشبورد {name} وارد شوید',
    'brand.f.name': 'نام پروژه',
    'brand.f.name.hint': 'در نوار کناری، صفحهٔ ورود و عنوان مرورگر دیده می‌شود.',
    'brand.f.shortName': 'نام کوتاه',
    'brand.f.shortName.hint': 'برای جاهای تنگ؛ حداکثر ۲۴ نویسه.',
    'brand.f.logo': 'لوگو برای حالت روشن',
    'brand.f.logo.hint': 'افقی و با پس‌زمینهٔ شفاف بهتر است.',
    'brand.f.logoDark': 'لوگو برای حالت تیره',
    'brand.f.logoDark.hint': 'اگر خالی بماند، همان لوگوی روشن نشان داده می‌شود.',
    'brand.f.mark': 'نشان و فاوآیکن',
    'brand.f.mark.hint': 'تصویر مربعی کوچک؛ در زبانهٔ مرورگر و نوار جمع‌شده دیده می‌شود.',
    'brand.f.accent': 'رنگ اصلی',
    'brand.f.accent.hint': 'رنگ دکمه‌ها، پیوندها و منوی فعال. خالی یعنی رنگ پیش‌فرض.',
    'brand.f.accent.pick': 'انتخاب رنگ',
    'brand.f.accent.presets': 'رنگ‌های آماده',
    'brand.f.accent.reset': 'رنگ پیش‌فرض',
    'brand.f.welcome': 'پیام خوش‌آمد صفحهٔ ورود',
    'brand.f.welcome.hint': 'زیر عنوان صفحهٔ ورود نشان داده می‌شود.',
    'brand.contrast.onWhite': 'دکمه با متن سفید',
    'brand.contrast.onDark': 'پیوند روی پس‌زمینهٔ تیره',
    'brand.contrast.good': 'خوانا',
    'brand.contrast.large': 'فقط برای متن درشت',
    'brand.contrast.poor': 'کم‌کنتراست',
    'brand.contrast.tip': 'خواندن متن با این رنگ سخت است. رنگی تیره‌تر یا پررنگ‌تر انتخاب کنید.',
    'brand.preview': 'پیش‌نمایش',
    'brand.preview.light': 'روشن',
    'brand.preview.dark': 'تیره',
    'brand.preview.nav': 'داشبورد',
    'brand.preview.link': 'یک پیوند',
    'brand.err.short': 'نام کوتاه حداکثر ۲۴ نویسه است.',
    'brand.err.color': 'رنگ را به شکل هگز وارد کنید، مثل #6b59ee.',
    'code.INVALID_COLOR': 'رنگ را به شکل هگز وارد کنید، مثل #6b59ee.',
    'code.INVALID_STEP': 'این مرحله وجود ندارد.',
    'code.INVALID_STEP_STATUS': 'وضعیت این مرحله درست نیست.',
    'code.INVALID_STATUS': 'وضعیت راه‌اندازی درست نیست.',
    'setup.title': 'راه‌اندازی کنسول',
    'setup.subtitle': 'چند قدم کوتاه تا آماده‌شدن سرور',
    'setup.steps': 'مرحله‌های راه‌اندازی',
    'setup.later': 'بعداً',
    'setup.stay': 'ادامه می‌دهم',
    'setup.later.title': 'بقیه را بعداً انجام می‌دهید؟',
    'setup.later.text': 'هرچه ذخیره کرده‌اید می‌ماند. هر وقت خواستید از تنظیمات پروژه › راه‌اندازی ادامه دهید.',
    'setup.later.toast': 'هر وقت خواستید از تنظیمات پروژه › راه‌اندازی ادامه دهید.',
    'setup.unsaved.title': 'تغییرات ذخیره نشده‌اند',
    'setup.unsaved.text': 'اگر از این مرحله بروید، تغییرات آن از بین می‌رود.',
    'setup.unsaved.ok': 'بدون ذخیره برو',
    'setup.back': 'قبلی',
    'setup.skip': 'رد کردن این مرحله',
    'setup.saveNext': 'ذخیره و ادامه',
    'setup.finish': 'تمام',
    'setup.done.toast': 'راه‌اندازی تمام شد.',
    'setup.stepOf': 'مرحلهٔ {n} از {total}',
    'setup.step.brand': 'برند',
    'setup.step.languages': 'زبان‌ها',
    'setup.step.currency': 'ارز',
    'setup.step.payment': 'پرداخت',
    'setup.step.sms': 'پیامک',
    'setup.step.email': 'ایمیل',
    'setup.step.review': 'مرور نهایی',
    'setup.step.brand.title': 'برند پروژه',
    'setup.step.brand.desc': 'نام، لوگو و رنگی که مدیران و بازیکنان می‌بینند.',
    'setup.step.languages.title': 'زبان‌های بازی',
    'setup.step.languages.desc': 'زبان‌هایی را که بازی دارد روشن کنید و یکی را پیش‌فرض بگذارید.',
    'setup.step.currency.title': 'ارز اصلی بازی',
    'setup.step.currency.desc': 'سکه یا الماسی که بازیکن با آن خرید می‌کند و موجودی اول هر بازیکن.',
    'setup.step.payment.title': 'پرداخت در بازی',
    'setup.step.payment.desc': 'درگاه‌ها و فروشگاه‌هایی که بازیکن از راه آن‌ها پول می‌دهد.',
    'setup.step.sms.title': 'ارسال پیامک',
    'setup.step.sms.desc': 'برای کد ورود و اعلان‌های پیامکی.',
    'setup.step.email.title': 'ارسال ایمیل',
    'setup.step.email.desc': 'برای بازیابی حساب، خبرها و اعلان‌ها.',
    'setup.step.review.title': 'مرور نهایی',
    'setup.step.review.desc': 'همه‌چیز را یک بار ببینید. هر جا لازم بود برگردید و درستش کنید.',
    'setup.st.done': 'انجام شد',
    'setup.st.skipped': 'رد شد',
    'setup.st.pending': 'دست‌نخورده',
    'setup.status.pending': 'شروع نشده',
    'setup.status.in_progress': 'در حال انجام',
    'setup.status.completed': 'تمام شده',
    'setup.status.skipped': 'برای بعد',
    'setup.no': 'نه',
    'setup.err.answer': 'اول به این سؤال جواب دهید.',
    'setup.sandbox.on': 'روشن است: همه‌چیز آزمایشی است و هیچ پیام یا پرداختی واقعی نیست.',
    'setup.sandbox.off': 'خاموش است: پیام‌ها و پرداخت‌ها واقعی‌اند.',
    'setup.pay.need': 'پرداخت لازم دارید؟',
    'setup.pay.no': 'باشد. هر وقت لازم شد، از تنظیمات پروژه › درگاه‌های پرداخت روشنش کنید.',
    'setup.pay.pickOne': 'دست‌کم یک درگاه را انتخاب کنید.',
    'setup.pay.secret.sandbox': 'در حالت آزمایشی لازم نیست.',
    'setup.sms.need': 'پیامک لازم دارید؟',
    'setup.sms.no': 'باشد. هر وقت لازم شد، از تنظیمات پروژه › پیامک اضافه‌اش کنید.',
    'setup.email.need': 'ایمیل لازم دارید؟',
    'setup.email.no': 'باشد. هر وقت لازم شد، از تنظیمات پروژه › ایمیل اضافه‌اش کنید.',
    'setup.pv.kind': 'سرویس',
    'setup.pv.kind.hint': 'همین سرویس، سرویس پیش‌فرض این کانال می‌شود.',
    'setup.pv.new': 'این سرویس هنوز ساخته نشده است. با ذخیره، در حالت آزمایشی ساخته می‌شود.',
    'setup.test': 'ارسال آزمایشی',
    'setup.test.hint': 'اول تنظیمات را ذخیره می‌کند، بعد پیام آزمایشی را می‌فرستد. در حالت آزمایشی چیزی واقعاً ارسال نمی‌شود.',
    'setup.test.send': 'ذخیره و ارسال آزمایشی',
    'setup.test.liveBlocked': 'در راهنما فقط در حالت آزمایشی می‌شود پیام آزمایشی فرستاد.',
    'setup.lang.list': 'زبان‌های فعلی',
    'setup.lang.active': 'فعال بودن {name}',
    'setup.lang.makeDefault': 'پیش‌فرض شود',
    'setup.lang.defaultSet': '{name} زبان پیش‌فرض شد.',
    'setup.lang.added': '{name} اضافه و فعال شد.',
    'setup.lang.quick': 'افزودن سریع',
    'setup.lang.quick.hint': 'روی هر زبان بزنید تا اضافه و فعال شود.',
    'setup.lang.quick.none': 'همهٔ زبان‌های رایج اضافه شده‌اند.',
    'setup.cur.pick': 'ارز اصلی',
    'setup.cur.pick.hint': 'ارزی را که می‌خواهید ویرایش کنید انتخاب کنید یا ارز تازه بسازید.',
    'setup.cur.new': '+ ارز تازه',
    'setup.cur.none': 'هنوز ارزی ندارید. اولین ارز را همین‌جا بسازید.',
    'setup.cur.defaultValue': 'موجودی اول',
    'setup.cur.defaultValue.hint': 'هر بازیکن تازه با همین مقدار شروع می‌کند.',
    'setup.rv.brand.default': 'ظاهر پیش‌فرض',
    'setup.rv.brand.custom': 'برند سفارشی',
    'setup.rv.none': 'هیچ',
    'setup.rv.sep': '، ',
    'setup.rv.balance': 'موجودی اول {n}',
    'setup.rv.edit': 'ویرایش',
    'setup.rv.editStep': 'ویرایش {step}',
    'setup.rv.left': '{n} مرحله هنوز دست‌نخورده است. می‌توانید همین حالا تمام کنید یا برگردید.',
    'setup.rv.ready': 'همه‌چیز آماده است.',
    'setup.open': 'باز کردن راهنما',
    'setup.progress': '{n} از {total} مرحله',
    'setup.nudge.title': 'راه‌اندازی پروژه هنوز تمام نشده',
    'setup.nudge.skipped': 'راه‌اندازی پروژه را برای بعد گذاشته‌اید',
    'setup.nudge.desc': 'چند قدم کوتاه مانده تا برند، زبان‌ها، سکه و پرداخت آماده شوند.',
    'setup.nudge.go': 'ادامهٔ راه‌اندازی',
    'account': 'حساب کاربری',
    'notFound': 'صفحهٔ {path} پیدا نشد؛ به داشبورد برگشتید.',
    'setup.meta.opened': 'اولین باز شدن',
    'setup.meta.completed': 'پایان',
    'tab.countries': 'کشورها',
    'tab.accounts': 'قوانین حساب',
    'tab.otp': 'ورود با پیامک',
    'tab.app-versions': 'نسخه‌های برنامه',
    'desc.countries': 'کشورهایی که بازیکن می‌تواند انتخاب کند و شماره‌اش را با پیش‌شمارهٔ آن‌ها وارد کند.',
    'desc.app-versions': 'آخرین نسخه و کمترین نسخهٔ مجاز هر پلتفرم. نسخه‌های قدیمی‌تر از کمترین نسخه باید به‌روز شوند.',
    'desc.accounts': 'قانون‌های ساختن حساب. سرور همین‌ها را برای نام کاربری، رمز و نام نمایشی چک می‌کند.',
    'desc.otp': 'آخرین درخواست‌های کد. خود کد هیچ‌جا نشان داده نمی‌شود.',
    'new.countries': 'کشور تازه',
    'edit.countries': 'ویرایش کشور',
    'new.app-versions': 'نسخهٔ تازه',
    'edit.app-versions': 'ویرایش نسخه',
    'empty.hint.countries': 'هنوز کشوری نساخته‌اید.',
    'empty.hint.app-versions': 'برای هر پلتفرم یک ردیف بسازید تا بازی بتواند نسخه‌اش را چک کند.',
    'empty.hint.otp': 'هنوز کسی کد ورود نخواسته است.',
    'search.countries': 'جستجو در نام، کد یا پیش‌شماره',
    'search.app-versions': 'جستجو در پلتفرم یا نسخه',
    'search.otp': 'جستجوی شماره',
    'col.country': 'کشور',
    'col.dialCode': 'پیش‌شماره',
    'col.platform': 'پلتفرم',
    'col.latest': 'آخرین نسخه',
    'col.minimum': 'کمترین نسخه',
    'col.stores': 'فروشگاه‌ها',
    'col.phone': 'شماره',
    'col.player': 'بازیکن',
    'ct.activated': '{name} فعال شد',
    'ct.deactivated': '{name} غیرفعال شد',
    'f.ct.code': 'کد کشور',
    'f.ct.code.hint': 'دو حرف لاتین، مثل IR یا DE.',
    'f.ct.dialCode': 'پیش‌شماره',
    'f.ct.dialCode.hint': 'مثل ‎+98‎. ورود با پیامک فقط شماره‌های کشورهای فعال را می‌پذیرد.',
    'f.ct.name': 'نام کشور',
    'f.ct.flag': 'تصویر پرچم',
    'f.ct.flag.hint': 'اگر خالی بماند، پرچم از روی کد کشور ساخته می‌شود.',
    'f.ct.active.hint': 'کشورهای غیرفعال در بازی نشان داده نمی‌شوند.',
    'f.ct.sort.hint': 'عدد کمتر، بالاتر در فهرست بازی می‌آید.',
    'code.INVALID_COUNTRY_CODE': 'کد کشور دو حرف لاتین است، مثل IR.',
    'code.INVALID_DIAL_CODE': 'پیش‌شماره با + و حداکثر پنج رقم است، مثل ‎+98‎.',
    'code.COUNTRY_NOT_FOUND': 'این کشور پیدا نشد.',
    'platform.android': 'اندروید',
    'platform.ios': 'آی‌اواس',
    'platform.web': 'وب',
    'platform.windows': 'ویندوز',
    'platform.macos': 'مک',
    'platform.linux': 'لینوکس',
    'store.google_play': 'گوگل‌پلی',
    'store.app_store': 'اپ‌استور',
    'store.bazaar': 'کافه‌بازار',
    'store.myket': 'مایکت',
    'store.microsoft_store': 'مایکروسافت استور',
    'store.steam': 'استیم',
    'store.web': 'وب‌سایت',
    'store.direct': 'دانلود مستقیم',
    'code.INVALID_VERSION': 'نسخه را با عدد و نقطه بنویسید، مثل 1.4.2.',
    'code.VERSION_REQUIRED': 'آخرین نسخه را بنویسید.',
    'code.MIN_ABOVE_LATEST': 'کمترین نسخه نمی‌تواند از آخرین نسخه تازه‌تر باشد.',
    'code.DUPLICATE_STORE': 'هر فروشگاه فقط یک بار می‌آید.',
    'code.INVALID_PLATFORM': 'این پلتفرم پشتیبانی نمی‌شود.',
    'code.INVALID_STORE': 'این فروشگاه پشتیبانی نمی‌شود.',
    'code.TOO_MANY_LINKS': 'حداکثر هشت لینک.',
    'av.link.add': 'افزودن لینک فروشگاه',
    'av.link.empty': 'هنوز لینکی نیست. بازیکن برای به‌روزرسانی به همین لینک‌ها می‌رود.',
    'av.link.store': 'فروشگاه',
    'av.link.url': 'لینک',
    'av.link.label': 'نام نمایشی',
    'av.link.hint': 'لینک https یا لینک خود فروشگاه ({schemes}). نام نمایشی اختیاری است.',
    'f.av.platform': 'پلتفرم',
    'f.av.active.hint': 'خاموش باشد، بازی همیشه «همه‌چیز درست است» می‌گیرد.',
    'f.av.latest': 'آخرین نسخه',
    'f.av.latest.hint': 'نسخه‌های قدیمی‌تر پیام «نسخهٔ تازه آمده» می‌گیرند.',
    'f.av.min': 'کمترین نسخه',
    'f.av.min.hint': 'نسخه‌های قدیمی‌تر از این باید به‌روز شوند (به‌روزرسانی اجباری).',
    'f.av.stores': 'لینک فروشگاه‌ها',
    'f.av.messages': 'پیام‌ها',
    'f.av.updateMessage': 'پیام به‌روزرسانی',
    'f.av.updateMessage.hint': 'وقتی نسخهٔ تازه‌تری هست ولی اجباری نیست.',
    'f.av.forceMessage': 'پیام به‌روزرسانی اجباری',
    'f.av.forceMessage.hint': 'وقتی نسخهٔ بازیکن از کمترین نسخه قدیمی‌تر است.',
    'otp.status.pending': 'منتظر کد',
    'otp.status.verified': 'وارد شد',
    'otp.status.expired': 'منقضی',
    'otp.status.failed': 'تلاش زیاد',
    'otp.status.replaced': 'کد تازه آمد',
    'otp.status.send_failed': 'ارسال نشد',
    'otp.sandbox': 'آزمایشی',
    'otp.recent': 'درخواست‌های اخیر',
    'otp.code': 'کد',
    'otp.limits': 'محدودیت‌ها',
    'otp.message': 'متن پیامک',
    'otp.provider.default': 'سرویس پیش‌فرض پیامک',
    'otp.mode.off': 'ورود با پیامک خاموش است.',
    'otp.mode.sandbox': 'حالت آزمایشی: پیامکی فرستاده نمی‌شود و کد فقط در لاگ سرور نوشته می‌شود.',
    'otp.mode.noProvider': 'هیچ سرویس پیامک فعالی نیست؛ کد فرستاده نمی‌شود.',
    'otp.mode.setup': 'تنظیم پیامک',
    'otp.mode.live': 'کدها با {name} فرستاده می‌شوند.',
    'f.otp.enabled': 'ورود با پیامک روشن باشد',
    'f.otp.enabled.hint': 'خاموش باشد، بازی نمی‌تواند کد بخواهد.',
    'f.otp.sandbox': 'حالت آزمایشی',
    'f.otp.sandbox.hint': 'پیامک واقعی فرستاده نمی‌شود؛ کد در لاگ سرور نوشته می‌شود.',
    'f.otp.provider': 'سرویس پیامک',
    'f.otp.provider.hint': 'سرویس‌ها را در بخش پیامک اضافه کنید.',
    'f.otp.codeLength': 'تعداد رقم کد',
    'f.otp.codeLength.hint': 'بین ۴ تا ۸ رقم.',
    'f.otp.ttl': 'اعتبار کد (ثانیه)',
    'f.otp.ttl.hint': 'بعد از این زمان کد کار نمی‌کند.',
    'f.otp.attempts': 'تعداد تلاش برای هر کد',
    'f.otp.attempts.hint': 'بعد از این تعداد کد اشتباه، باید کد تازه خواست.',
    'f.otp.resend': 'فاصلهٔ ارسال دوباره (ثانیه)',
    'f.otp.resend.hint': 'بازیکن تا این زمان نمی‌تواند کد تازه بخواهد.',
    'f.otp.phoneHourly': 'کد در ساعت برای هر شماره',
    'f.otp.phoneDaily': 'کد در روز برای هر شماره',
    'f.otp.ipHourly': 'کد در ساعت برای هر IP',
    'f.otp.ipHourly.hint': 'جلوی ساختن انبوه کد از یک جا را می‌گیرد.',
    'f.otp.message': 'متن پیامک',
    'f.otp.message.hint': '{code} جای کد است و باید در هر زبان باشد. {minutes} زمان اعتبار و {app} نام برند است.',
    'code.DAILY_BELOW_HOURLY': 'سقف روزانه نمی‌تواند از سقف ساعتی کمتر باشد.',
    'code.OTP_MESSAGE_NEEDS_CODE': 'متن هر زبان باید {code} داشته باشد.',
    'code.PROVIDER_NOT_SMS': 'یک سرویس پیامک انتخاب کنید.',
    'code.INVALID_RULE': 'این عدد مجاز نیست.',
    'ar.username': 'نام کاربری',
    'ar.password': 'رمز',
    'ar.display': 'نام نمایشی',
    'f.ar.usernameMin': 'کمترین طول',
    'f.ar.usernameMax': 'بیشترین طول',
    'f.ar.charset': 'حروف مجاز',
    'f.ar.charset.hint': 'با «فقط لاتین» فقط A تا Z، رقم و _ . - پذیرفته می‌شود.',
    'ar.charset.any': 'هر حرفی (فارسی هم)',
    'ar.charset.latin': 'فقط لاتین',
    'f.ar.passwordMin': 'کمترین طول رمز',
    'f.ar.passwordMin.hint': 'نکاما رمز کمتر از ۸ حرف را نمی‌پذیرد.',
    'f.ar.passwordLetter': 'رمز حرف داشته باشد',
    'f.ar.passwordDigit': 'رمز عدد داشته باشد',
    'f.ar.displayNameMax': 'بیشترین طول نام نمایشی',
    'code.MAX_BELOW_MIN': 'بیشترین طول نمی‌تواند از کمترین طول کمتر باشد.',
    'ar.try': 'امتحان کنید',
    'ar.try.hint': 'یک نام کاربری و رمز بنویسید تا ببینید با همین عددهای فرم پذیرفته می‌شود یا نه.',
    'ar.try.username': 'نام کاربری نمونه',
    'ar.try.password': 'رمز نمونه',
    'ar.try.ok': 'پذیرفته می‌شود.',
    'code.USERNAME_TOO_SHORT': 'نام کاربری دست‌کم {n} حرف است.',
    'code.USERNAME_TOO_LONG': 'نام کاربری حداکثر {n} حرف است.',
    'code.USERNAME_INVALID_CHARACTERS': 'نام کاربری فقط حرف لاتین، رقم و _ . - دارد.',
    'code.PASSWORD_TOO_SHORT': 'رمز دست‌کم {n} حرف است.',
    'code.PASSWORD_NEEDS_LETTER': 'رمز باید حرف داشته باشد.',
    'code.PASSWORD_NEEDS_DIGIT': 'رمز باید عدد داشته باشد.',
    'section.mod': 'بازیکنان و گفتگو',
    'section.mod.short': 'گفتگو',
    'nav.players': 'مدیریت بازیکنان',
    'nav.reports': 'گزارش‌ها',
    'nav.restrictions': 'تعلیق و سکوت',
    'nav.chat-words': 'فیلتر کلمات',
    'desc.players': 'برچسب بازیکنان را ببینید، گروهی برچسب بزنید و بازیکنان معلق یا بی‌صدا را پیدا کنید.',
    'desc.reports': 'پیام‌هایی که بازیکنان گزارش کرده‌اند یا فیلتر علامت زده است.',
    'desc.restrictions': 'همهٔ تعلیق‌ها، سکوت‌ها و اخطارها، با امکان برداشتن زودتر از موعد.',
    'desc.chat-words': 'کلمه‌هایی که در گفتگو پنهان، رد یا برای بررسی علامت زده می‌شوند.',
    'desc.report-titles': 'دلیل‌هایی که بازیکن هنگام گزارش یک پیام انتخاب می‌کند.',
    'tab.report-titles': 'عنوان‌های گزارش',
    'new.chat-words': 'کلمهٔ تازه',
    'edit.chat-words': 'ویرایش کلمه',
    'new.report-titles': 'عنوان تازه',
    'edit.report-titles': 'ویرایش عنوان گزارش',
    'empty.hint.players': 'هنوز بازیکنی ثبت‌نام نکرده است.',
    'empty.hint.reports': 'گزارش تازه‌ای نیست. گزارش‌ها اینجا نشان داده می‌شوند.',
    'empty.hint.restrictions': 'هنوز کسی تعلیق یا بی‌صدا نشده است.',
    'empty.hint.chat-words': 'اولین کلمه را اضافه کنید تا گفتگو پاکیزه بماند.',
    'empty.hint.report-titles': 'یک عنوان بسازید تا بازیکنان بتوانند پیام‌ها را گزارش کنند.',
    'search.players': 'جست‌وجوی نام کاربری، نام یا شناسه',
    'search.reports': 'جست‌وجوی متن پیام یا نام بازیکن',
    'search.restrictions': 'جست‌وجوی بازیکن یا دلیل',
    'search.chat-words': 'جست‌وجوی کلمه',
    'rp.openN': '{n} گزارش باز',
    'dur.hours': '{n} ساعت',
    'dur.days': '{n} روز',
    'dur.custom': 'تاریخ دلخواه',
    'rs.until': 'تا {date}',
    'rs.duration': 'مدت',
    'rs.duration.suspend.hint': 'بعد از این زمان، تعلیق خودش برداشته می‌شود.',
    'rs.duration.mute.hint': 'بعد از این زمان، بازیکن دوباره می‌تواند پیام بفرستد.',
    'rs.reason': 'پیامی که بازیکن می‌بیند',
    'rs.reason.hint.suspend': 'بازیکن هنگام ورود این پیام را با زمان پایان تعلیق می‌بیند.',
    'rs.reason.hint.mute': 'بازیکن وقتی پیام می‌فرستد این را با زمان پایان سکوت می‌بیند.',
    'rs.reason.hint.warn': 'این پیام همان لحظه برای بازیکن فرستاده می‌شود.',
    'rs.reason.default.suspend': 'حسابت به خاطر زیر پا گذاشتن قوانین بازی موقتاً بسته شده است.',
    'rs.reason.default.mute': 'به خاطر رعایت نکردن قوانین گفتگو، فعلاً نمی‌توانی پیام بفرستی.',
    'rs.reason.default.warn': 'لطفاً در گفتگو محترمانه رفتار کن. دفعهٔ بعد ممکن است بی‌صدا شوی.',
    'rs.note': 'یادداشت داخلی',
    'rs.note.ph': 'فقط مدیران این را می‌بینند',
    'rs.note.hint': 'برای همکاران؛ به بازیکن نشان داده نمی‌شود.',
    'rs.liftNote': 'چرا زودتر برداشته می‌شود؟',
    'rs.kind.suspend': 'تعلیق',
    'rs.kind.mute': 'سکوت',
    'rs.kind.warn': 'اخطار',
    'rs.status.active': 'فعال',
    'rs.status.expired': 'تمام شده',
    'rs.status.lifted': 'برداشته شده',
    'rs.liftedAt': 'برداشته شد در {date}',
    'rs.do.suspend': 'تعلیق موقت',
    'rs.do.mute': 'بی‌صدا کردن',
    'rs.do.warn': 'اخطار',
    'rs.dialog.suspend': 'تعلیق موقت {name}',
    'rs.dialog.mute': 'بی‌صدا کردن {name}',
    'rs.dialog.warn': 'اخطار به {name}',
    'rs.dialog.suspend.text': 'بازیکن تا پایان این مدت نمی‌تواند وارد بازی شود و از بازی بیرون می‌رود.',
    'rs.dialog.mute.text': 'بازیکن تا پایان این مدت نمی‌تواند در گفتگو پیام بفرستد.',
    'rs.dialog.warn.text': 'بازیکن یک پیام اخطار می‌گیرد و در سابقه‌اش ثبت می‌شود.',
    'rs.done.suspend': '{name} تعلیق شد.',
    'rs.done.mute': '{name} بی‌صدا شد.',
    'rs.done.warn': 'اخطار برای {name} فرستاده شد.',
    'rs.lift': 'برداشتن',
    'rs.lift.suspend': 'برداشتن تعلیق',
    'rs.lift.mute': 'برداشتن سکوت',
    'rs.lift.text': 'محدودیت {name} همین حالا برداشته می‌شود.',
    'rs.lift.ok': 'بله، بردار',
    'rs.lifted': 'محدودیت برداشته شد.',
    'code.END_REQUIRED': 'زمان پایان را انتخاب کنید.',
    'code.END_IN_PAST': 'زمان پایان باید در آینده باشد.',
    'code.END_TOO_FAR': 'حداکثر ده سال؛ برای همیشه از مسدود کردن خود Nakama استفاده کنید.',
    'code.INVALID_END': 'تاریخ درست نیست.',
    'code.INVALID_RESTRICTION_KIND': 'نوع محدودیت درست نیست.',
    'code.NOT_ACTIVE': 'این محدودیت قبلاً تمام شده است.',
    'code.REPORT_CLOSED': 'این گزارش قبلاً بررسی شده است.',
    'code.INVALID_MOD_ACTION': 'این کار مجاز نیست.',
    'code.INVALID_MATCH': 'نوع جست‌وجو درست نیست.',
    'code.INVALID_WORD': 'کلمه باید ۱ تا ۶۴ حرف داشته باشد و دست‌کم یک حرف الفبا.',
    'code.TAGS_REQUIRED': 'دست‌کم یک برچسب انتخاب کنید.',
    'code.TOO_MANY_TAGS': 'برچسب‌ها بیش از حد مجاز است.',
    'code.USERS_REQUIRED': 'دست‌کم یک بازیکن انتخاب کنید.',
    'code.USER_NOT_FOUND': 'این بازیکن پیدا نشد.',
    'code.RESTRICTION_NOT_FOUND': 'این محدودیت پیدا نشد.',
    'code.REPORT_NOT_FOUND': 'این گزارش پیدا نشد.',
    'code.MESSAGE_NOT_FOUND': 'این پیام پیدا نشد.',
    'code.INVALID_FILTER': 'این فیلتر درست نیست.',
    'code.INVALID_OP': 'این کار مجاز نیست.',
    'code.TARGET_TAGS_REQUIRED': 'دست‌کم یک برچسب بازیکن انتخاب کنید.',
    'tag.public': 'دیدنی برای بازیکن',
    'tag.playersN': '{n} بازیکن',
    'f.tag.public': 'بازیکن این برچسب را ببیند',
    'f.tag.public.hint': 'اگر روشن باشد، بازی می‌تواند این برچسب را به خود بازیکن نشان دهد؛ مثلاً «عضو ویژه».',
    'f.audience': 'فقط برای این بازیکنان',
    'f.audience.hint': 'اگر برچسبی انتخاب کنید، این محصول فقط به بازیکنانی نشان داده می‌شود که دست‌کم یکی از این برچسب‌ها را دارند.',
    'audience.only': 'ویژهٔ برچسب',
    'tg.tags': 'برچسب بازیکن',
    'tg.tagsN': '{n} برچسب',
    'tg.pickTags': 'برچسب‌ها',
    'tg.tags.hint': 'بازیکنانی که دست‌کم یکی از این برچسب‌ها را دارند پیام را می‌گیرند.',
    'pl.banned': 'مسدود دائمی',
    'pl.suspendedUntil': 'معلق تا {date}',
    'pl.mutedUntil': 'بی‌صدا تا {date}',
    'pl.ok': 'بدون محدودیت',
    'pl.status.suspended': 'معلق',
    'pl.status.muted': 'بی‌صدا',
    'pl.status.banned': 'مسدود دائمی',
    'pl.status.tagged': 'برچسب‌دار',
    'pl.select': 'انتخاب {name}',
    'pl.bulk.tags': 'برچسب‌ها',
    'pl.bulk.pick': 'انتخاب برچسب',
    'pl.bulk.add': 'افزودن برچسب',
    'pl.bulk.remove': 'برداشتن برچسب',
    'pl.bulk.page': 'انتخاب همهٔ این صفحه',
    'pl.bulk.clear': 'لغو انتخاب',
    'pl.bulk.selected': '{n} بازیکن انتخاب شده',
    'pl.bulk.none': 'بازیکنی انتخاب نشده؛ از ستون اول انتخاب کنید.',
    'pl.bulk.added': 'برچسب برای {n} بازیکن ثبت شد.',
    'pl.bulk.removed': 'برچسب از {n} بازیکن برداشته شد.',
    'rp.title': 'بررسی گزارش',
    'rp.gone': 'بازیکن حذف شده',
    'rp.deleted': 'پیام پاک شده',
    'rp.source.player': 'گزارش بازیکن',
    'rp.source.filter': 'فیلتر کلمات',
    'rp.noTitle': 'بدون عنوان',
    'rp.byFilter': 'فیلتر خودکار',
    'rp.reportsN': '{n} گزارش',
    'rp.thisMessage': 'پیام گزارش‌شده',
    'rp.noContext': 'گفتگویی پیدا نشد.',
    'rp.conversation': 'گفتگو',
    'rp.report': 'گزارش',
    'rp.reason': 'دلیل',
    'rp.playerNote': 'توضیح بازیکن',
    'rp.matched': 'کلمه‌های پیدا شده',
    'rp.others': 'گزارش‌های دیگر',
    'rp.channel': 'اتاق',
    'rp.sender': 'فرستنده',
    'rp.history': 'سابقه',
    'rp.links': 'پیوندها',
    'rp.seeHistory': 'سابقهٔ محدودیت‌ها',
    'rp.seeReports': 'همهٔ گزارش‌های او',
    'rp.seeRestriction': 'دیدن محدودیت',
    'rp.count.reports': '{n} گزارش',
    'rp.count.warn': '{n} اخطار',
    'rp.count.mute': '{n} سکوت',
    'rp.count.suspend': '{n} تعلیق',
    'rp.decide': 'تصمیم',
    'rp.decision': 'نتیجه',
    'rp.action': 'کار',
    'rp.action.dismiss': 'رد گزارش',
    'rp.action.delete': 'پاک کردن پیام',
    'rp.action.warn': 'اخطار',
    'rp.action.mute': 'بی‌صدا',
    'rp.action.suspend': 'تعلیق',
    'rp.action.dismiss.hint': 'پیام مشکلی ندارد؛ گزارش بسته می‌شود.',
    'rp.action.delete.hint': 'پیام از گفتگو پاک می‌شود.',
    'rp.action.warn.hint': 'فرستنده یک اخطار می‌گیرد.',
    'rp.action.mute.hint': 'فرستنده تا زمانی که انتخاب می‌کنید نمی‌تواند پیام بفرستد.',
    'rp.action.suspend.hint': 'فرستنده تا زمانی که انتخاب می‌کنید نمی‌تواند وارد بازی شود.',
    'rp.alsoDelete': 'پیام هم پاک شود',
    'rp.note.hint': 'برای همکاران ثبت می‌شود.',
    'rp.submit': 'ثبت تصمیم',
    'rp.submit.dismiss': 'رد گزارش',
    'rp.submit.delete': 'پاک کردن پیام',
    'rp.submit.warn': 'فرستادن اخطار',
    'rp.submit.mute': 'بی‌صدا کردن',
    'rp.submit.suspend': 'تعلیق کردن',
    'rp.done.dismiss': 'گزارش رد شد.',
    'rp.done.delete': 'پیام پاک شد.',
    'rp.done.warn': 'اخطار فرستاده شد.',
    'rp.done.mute': 'فرستنده بی‌صدا شد.',
    'rp.done.suspend': 'فرستنده تعلیق شد.',
    'rp.status.open': 'باز',
    'rp.status.dismissed': 'رد شده',
    'rp.status.resolved': 'بررسی شده',
    'rp.messageDeleted': 'پیام پاک شد',
    'rp.restriction': 'محدودیت',
    'rp.resolvedBy': 'بررسی‌کننده',
    'rp.resolvedAt': 'زمان بررسی',
    'cw.action.mask': 'پنهان با ستاره',
    'cw.action.block': 'رد پیام',
    'cw.action.flag': 'علامت برای بررسی',
    'cw.match.word': 'کلمهٔ کامل',
    'cw.match.contains': 'هر جای متن',
    'cw.try': 'امتحان فیلتر',
    'cw.try.hint': 'یک جمله بنویسید تا ببینید بازیکنان چه می‌بینند. نیم‌فاصله، کشیده، اعراب، ي و ك عربی، فاصله بین حروف و حرف‌های تکراری هم شناخته می‌شوند.',
    'cw.try.label': 'جملهٔ نمونه',
    'cw.try.ph': 'مثلاً: تو اح‌ــمق هستی',
    'cw.try.blocked': 'این پیام فرستاده نمی‌شود.',
    'cw.try.masked': 'بخشی از پیام با ستاره پنهان می‌شود.',
    'cw.try.flagged': 'پیام فرستاده می‌شود و برای بررسی در گزارش‌ها می‌آید.',
    'cw.try.maskedFlagged': 'بخشی پنهان می‌شود و پیام برای بررسی در گزارش‌ها می‌آید.',
    'cw.try.clean': 'این پیام بدون تغییر فرستاده می‌شود.',
    'cw.try.shown': 'آنچه دیگران می‌بینند',
    'cw.try.notSent': 'فرستاده نمی‌شود',
    'cw.try.normalized': 'شکل یکسان‌شده:',
    'f.cw.word': 'کلمه یا عبارت',
    'f.cw.word.hint': 'شکل‌های دیگر همین کلمه (با ي عربی، نیم‌فاصله یا حرف تکراری) خودشان پیدا می‌شوند.',
    'f.cw.action': 'چه شود؟',
    'f.cw.action.hint': 'پنهان: با ستاره. رد: پیام فرستاده نمی‌شود. علامت: فرستاده می‌شود و برای بررسی می‌آید.',
    'f.cw.match': 'کجا پیدا شود؟',
    'f.cw.match.hint': 'کلمهٔ کامل پسوندهایی مثل «ها» و «ی» را هم می‌گیرد. «هر جای متن» برای کلمه‌های چسبیده به هم است.',
    'f.cw.active.hint': 'کلمهٔ غیرفعال نادیده گرفته می‌شود.',
    'f.cw.note': 'یادداشت',
    'f.cw.note.hint': 'فقط مدیران می‌بینند.',
    'f.rt.key.hint': 'شناسهٔ ثابت برای بازی، مثل spam.',
    'f.rt.name': 'عنوان',
    'f.rt.name.hint': 'بازیکن این را در فهرست دلیل‌ها می‌بیند.',
    'f.rt.description': 'توضیح',
    'f.rt.description.hint': 'یک جملهٔ کوتاه زیر عنوان.',
    'f.rt.active.hint': 'عنوان غیرفعال به بازیکن نشان داده نمی‌شود.',
    'pc.title': 'برچسب‌ها و محدودیت‌ها',
    'pc.tags': 'برچسب‌ها',
    'pc.tags.edit': 'ویرایش برچسب‌ها',
    'pc.tags.edit.text': 'برچسب‌های بازیکن برای پیام‌ها و پیشنهادهای ویژه به کار می‌روند.',
    'pc.tags.hint': 'فقط برچسب‌های نوع «بازیکن» اینجا می‌آیند.',
    'pc.tags.saved': 'برچسب‌ها ذخیره شد.',
    'pc.status': 'وضعیت',
    'pc.history': 'سابقه',
    'pc.historyN': '{n} مورد در سابقه',
    'pc.reportsN': '{n} گزارش علیه او',
    'pc.lastWarn': 'آخرین اخطار: {date}',
    'filter.tag': 'برچسب',
    'filter.kind': 'نوع',
    'filter.playerId': 'شناسهٔ بازیکن',
    'filter.source': 'منبع',
    'filter.reportTitle': 'عنوان گزارش',
    'filter.action': 'کار',
    'col.select': 'انتخاب',
    'col.tags': 'برچسب‌ها',
    'col.kind': 'نوع',
    'col.reason': 'دلیل',
    'col.by': 'به دست',
    'col.message': 'پیام',
    'col.reporter': 'گزارش‌دهنده',
    'col.word': 'کلمه',
    'col.action': 'کار',
    'col.match': 'جست‌وجو',
    'col.used': 'استفاده',
    'nav.event-types': 'رویدادها',
    'nav.event-rules': 'جایزهٔ رویدادها',
    'nav.meta-rules': 'قانون‌های متا',
    'nav.level-tables': 'جدول سطح‌ها',
    'nav.streaks': 'روزهای پیاپی',
    'nav.event-log': 'گزارش رویدادها',
    'desc.event-types': 'کارهایی که برای هر بازیکن ثبت می‌شود، مثل ورود به بازی، پایان یک دور یا خرید. رویدادی که اینجا تعریف نشده باشد پذیرفته نمی‌شود.',
    'desc.event-rules': 'وقتی بازیکن در یک رویداد به عدد مشخصی برسد جایزه می‌گیرد؛ خودکار یا با دکمهٔ «دریافت» در بازی.',
    'desc.meta-rules': 'هر رویداد چه مقداری به متاهای بازیکن، مثل امتیاز تجربه، اضافه کند.',
    'desc.level-tables': 'با رسیدن یک متا به هر آستانه، سطح بازیکن بالا می‌رود و جایزهٔ همان سطح آماده می‌شود.',
    'desc.streaks': 'جایزه برای چند روز پشت‌سرهم آمدن یا انجام یک کار. این بخش به Hiro وابسته نیست.',
    'desc.event-log': 'رویدادهای اخیر بازیکنان. رویدادهای قدیمی‌تر از مدت نگه‌داری پاک می‌شوند، ولی جمع‌ها می‌مانند.',
    'desc.events': 'منطقهٔ زمانی و شروع هفته برای شمارش روزانه و هفتگی، مدت نگه‌داری گزارش رویدادها و سقف رویدادهای هر بازیکن.',
    'tab.events': 'رویدادها',
    'new.event-types': 'رویداد جدید',
    'edit.event-types': 'ویرایش رویداد',
    'new.event-rules': 'جایزهٔ رویداد جدید',
    'edit.event-rules': 'ویرایش جایزهٔ رویداد',
    'new.meta-rules': 'قانون متای جدید',
    'edit.meta-rules': 'ویرایش قانون متا',
    'new.level-tables': 'جدول سطح جدید',
    'edit.level-tables': 'ویرایش جدول سطح',
    'new.streaks': 'روزهای پیاپی جدید',
    'edit.streaks': 'ویرایش روزهای پیاپی',
    'empty.hint.event-types': 'اولین رویداد را بسازید، مثلاً پایان یک دور.',
    'empty.hint.event-rules': 'مثلاً: هر روز سه دور بازی کن و ۲۰ سکه بگیر.',
    'empty.hint.meta-rules': 'مثلاً: امتیاز هر دور به امتیاز تجربه اضافه شود.',
    'empty.hint.level-tables': 'یک جدول بسازید تا امتیاز تجربه به سطح تبدیل شود.',
    'empty.hint.streaks': 'مثلاً یک جایزهٔ هفت‌روزه برای ورود روزانه بسازید.',
    'empty.hint.event-log': 'هنوز رویدادی ثبت نشده است.',
    'search.event-log': 'جستجو در نام کاربری یا کلید رویداد',
    'rs.when.count': '«{event}» {period} {n} بار',
    'rs.when.sum': 'جمع «{event}» {period} به {n} برسد',
    'rs.when.max': 'بیشترین مقدار «{event}» {period} به {n} برسد',
    'rs.period.total': 'در کل',
    'rs.period.day': 'در یک روز',
    'rs.period.week': 'در یک هفته',
    'rs.repeat.once': 'فقط یک بار',
    'rs.repeat.period.day': 'هر روز یک بار',
    'rs.repeat.period.week': 'هر هفته یک بار',
    'rs.repeat.period.total': 'یک بار',
    'rs.repeat.every': 'به ازای هر {n}',
    'rs.max': 'حداکثر {n} بار',
    'grant.auto': 'خودکار',
    'grant.claim': 'با دکمهٔ دریافت',
    'ev.window.always': 'همیشه',
    'ev.window.from': 'از {date}',
    'ev.window.to': 'تا {date}',
    'rows.empty': 'هنوز ردیفی اضافه نشده.',
    'code.DUPLICATE_LEVEL': 'هر سطح فقط یک بار می‌آید.',
    'code.DUPLICATE_DAY': 'هر روز فقط یک بار می‌آید.',
    'code.THRESHOLDS_NOT_ORDERED': 'سطح بالاتر باید آستانهٔ برابر یا بیشتر داشته باشد.',
    'code.SAME_META': 'سطح باید در متای دیگری نوشته شود.',
    'code.INVALID_RANGE': 'بیشترین مقدار نباید از کمترین کوچک‌تر باشد.',
    'code.REPEAT_NEEDS_PERIOD': '«یک بار در هر بازه» فقط با بازهٔ روزانه یا هفتگی کار می‌کند.',
    'code.EVERY_NEEDS_TOTAL': '«به ازای هر هدف» با تعداد یا جمع کار می‌کند، نه با بیشترین مقدار.',
    'code.INVALID_EVENT_KEY': 'از ۲ تا ۴۸ حرف کوچک لاتین، رقم، _ یا نقطه استفاده کنید و با حرف شروع کنید.',
    'code.EVENT_TYPE_NOT_FOUND': 'این رویداد تعریف نشده یا غیرفعال است.',
    'code.EVENT_NOT_ALLOWED': 'این رویداد را فقط سرور ثبت می‌کند.',
    'code.VALUE_OUT_OF_RANGE': 'مقدار بیرون از بازهٔ مجاز این رویداد است.',
    'code.METADATA_TOO_LARGE': 'اطلاعات همراه رویداد بیش از اندازه بزرگ است.',
    'code.TOO_MANY_EVENTS': 'رویدادها بیش از حد زیاد شد؛ یک دقیقهٔ دیگر دوباره امتحان کنید.',
    'code.SYSTEM_EVENT': 'رویدادهای داخلی حذف نمی‌شوند؛ می‌توانید غیرفعالشان کنید.',
    'code.SYSTEM_EVENT_SERVER_ONLY': 'رویدادهای داخلی را فقط سرور ثبت می‌کند.',
    'code.INVALID_RATE_LIMIT': 'سقف باید بین ۱ تا ۱۰۰٬۰۰۰ در دقیقه باشد.',
    'code.INVALID_THRESHOLD': 'هدف معتبر نیست.',
    'code.INVALID_MAX_CLAIMS': 'دست‌کم ۱، یا خالی برای بی‌سقف.',
    'code.INVALID_GRANT_MODE': 'روش دادن جایزه معتبر نیست.',
    'code.INVALID_MEASURE': 'نوع شمارش معتبر نیست.',
    'code.INVALID_PERIOD': 'بازه معتبر نیست.',
    'code.INVALID_REPEAT': 'تکرار معتبر نیست.',
    'code.INVALID_META_ACTION': 'تغییر متا معتبر نیست.',
    'code.INVALID_VALUE_SOURCE': 'منبع مقدار معتبر نیست.',
    'code.STEPS_REQUIRED': 'دست‌کم یک سطح اضافه کنید.',
    'code.TOO_MANY_STEPS': 'حداکثر ۲۰۰ سطح.',
    'code.INVALID_LEVEL': 'سطح‌ها از ۱ شروع می‌شوند.',
    'code.INVALID_PERIOD_DAYS': 'بین ۱ تا ۳۰ روز.',
    'code.INVALID_GRACE': 'بین ۰ تا ۳۰.',
    'code.INVALID_MISS_ACTION': 'رفتار بعد از غیبت معتبر نیست.',
    'code.INVALID_CYCLE': 'بین ۱ تا ۳۶۵ روز، یا خالی.',
    'code.TOO_MANY_REWARDS': 'حداکثر ۳۶۵ روز جایزه.',
    'code.INVALID_DAY': 'روزها از ۱ شروع می‌شوند.',
    'code.DAY_OUTSIDE_CYCLE': 'این روز بعد از پایان دوره است.',
    'code.INVALID_TIMEZONE': 'منطقهٔ زمانی معتبر نیست.',
    'code.INVALID_WEEK_START': 'روز شروع هفته معتبر نیست.',
    'code.INVALID_RETENTION': 'بین ۱ تا ۳۶۵۰ روز.',
    'code.CLAIM_NOT_FOUND': 'این جایزه پیدا نشد.',
    'code.ALREADY_CLAIMED': 'این جایزه قبلاً داده شده است.',
    'code.STREAK_NOT_FOUND': 'این بازیکن هنوز در این روزهای پیاپی نیست.',
    'filter.ev.source': 'ثبت‌کننده',
    'ev.source.client': 'بازی هم',
    'ev.source.server': 'فقط سرور',
    'ev.system': 'داخلی',
    'col.ev.source': 'ثبت‌کننده',
    'col.ev.range': 'بازهٔ مقدار',
    'ev.range.any': 'هر مقدار',
    'col.ev.rate': 'سقف',
    'ev.perMinute': '{n} در دقیقه',
    'ev.usedBy': '{n} قانون',
    'ev.system.notice': 'این رویداد داخلی است و خود سرور آن را ثبت می‌کند. نام، توضیح و روشن یا خاموش بودنش را می‌توانید عوض کنید.',
    'f.ev.key.hint': 'همان کلیدی که بازی می‌فرستد، مثل round_finished. بعداً عوض نمی‌شود.',
    'f.ev.client': 'بازی هم بتواند ثبتش کند',
    'f.ev.client.hint': 'روشن یعنی خود بازی این رویداد را می‌فرستد. برای کارهای مهم مثل خرید یا بُرد خاموش بگذارید تا فقط سرور ثبت کند.',
    'f.ev.valueMin': 'کمترین مقدار',
    'f.ev.valueMax': 'بیشترین مقدار',
    'f.ev.rate': 'سقف برای هر بازیکن در دقیقه',
    'f.ev.rate.hint': 'بیشتر از این، رویدادِ بازی پذیرفته نمی‌شود. رویدادهای سرور سقف ندارند.',
    'f.ev.active.hint': 'رویداد غیرفعال پذیرفته نمی‌شود و قانون‌هایش اجرا نمی‌شوند.',
    'filter.ev.event': 'رویداد',
    'col.ev.when': 'شرط',
    'col.ev.prize': 'جایزه',
    'col.ev.claims': 'جایزه‌ها',
    'ev.ended': 'تمام‌شده',
    'f.ev.rule.name': 'نام برای بازیکن',
    'f.ev.rule.name.hint': 'در بازی کنار نوار پیشرفت دیده می‌شود.',
    'f.ev.rule.description': 'توضیح برای بازیکن',
    'ev.h.when': 'چه وقت',
    'f.ev.event': 'رویداد',
    'f.ev.measure': 'شمارش',
    'f.ev.measure.hint': 'تعداد دفعه‌ها، جمع مقدارها یا بیشترین مقدار در یک بار.',
    'measure.count': 'تعداد دفعه‌ها',
    'measure.sum': 'جمع مقدارها',
    'measure.max': 'بیشترین مقدار',
    'f.ev.period': 'بازه',
    'period.total': 'از اول',
    'period.day': 'هر روز',
    'period.week': 'هر هفته',
    'f.ev.threshold': 'هدف',
    'f.ev.repeat': 'تکرار',
    'repeat.once': 'فقط یک بار',
    'repeat.period': 'یک بار در هر بازه',
    'repeat.every': 'به ازای هر هدف',
    'f.ev.maxClaims': 'سقف جایزه برای هر بازیکن',
    'f.ev.maxClaims.hint': 'خالی یعنی بی‌سقف.',
    'ev.h.prize': 'جایزه',
    'f.ev.grantMode': 'روش دادن',
    'f.ev.grantMode.hint': 'خودکار یعنی همان لحظه به حساب بازیکن می‌رود؛ با دکمهٔ دریافت یعنی بازیکن خودش آن را برمی‌دارد.',
    'ev.h.who': 'چه کسی و چه زمانی',
    'f.ev.audience': 'فقط برای این بازیکنان',
    'f.ev.audience.hint': 'خالی یعنی همهٔ بازیکنان. با برچسب، فقط بازیکنانی که دست‌کم یکی از برچسب‌ها را دارند.',
    'f.ev.startsAt': 'شروع',
    'f.ev.endsAt': 'پایان',
    'filter.ev.meta': 'متا',
    'col.ev.event': 'رویداد',
    'col.ev.change': 'تغییر',
    'mr.change.add.fixed': '+ {n}',
    'mr.change.add.event': '+ مقدار × {n}',
    'mr.change.set.fixed': '= {n}',
    'mr.change.set.event': '= مقدار × {n}',
    'mr.change.max.fixed': 'دست‌کم {n}',
    'mr.change.max.event': 'بیشترینِ مقدار × {n}',
    'f.mr.sort.hint': 'قانون‌ها به این ترتیب اجرا می‌شوند.',
    'f.mr.name': 'نام',
    'f.mr.name.hint': 'برای پیدا کردن قانون در کنسول.',
    'f.mr.meta': 'متا',
    'f.mr.action': 'تغییر',
    'mr.action.add': 'اضافه کن',
    'mr.action.set': 'برابر کن',
    'mr.action.max': 'بیشترین را نگه دار',
    'f.mr.source': 'مقدار از',
    'f.mr.source.hint': 'عدد ثابت، یا مقدار خود رویداد ضرب در عدد پایین.',
    'mr.source.fixed': 'عدد ثابت',
    'mr.source.event': 'مقدار رویداد',
    'f.mr.amount': 'عدد',
    'f.mr.amount.hint': 'برای «مقدار رویداد» ضریب است؛ مثلاً ۱ یعنی همان امتیاز دور.',
    'col.lv.meta': 'متا',
    'col.lv.steps': 'سطح‌ها',
    'lv.levelsN': '{n} سطح',
    'lv.top': 'سطح {level} با {n}',
    'col.ev.players': 'بازیکن',
    'f.lv.meta': 'متای امتیاز',
    'f.lv.meta.hint': 'مثلاً امتیاز تجربه. هر بار که این متا تغییر کند سطح دوباره حساب می‌شود.',
    'f.lv.levelMeta': 'سطح در این متا نوشته شود',
    'f.lv.levelMeta.hint': 'مثلاً متای «سطح»، تا بازی و محصولات با حداقل سطح از آن استفاده کنند.',
    'lv.h.steps': 'سطح‌ها و جایزه‌ها',
    'lv.add': 'افزودن سطح',
    'lv.empty': 'دست‌کم یک سطح اضافه کنید.',
    'f.lv.level': 'سطح',
    'f.lv.threshold': 'آستانه',
    'f.lv.title': 'عنوان سطح',
    'col.sk.rule': 'قاعده',
    'sk.everyDay': 'هر روز',
    'sk.everyDays': 'هر {n} روز',
    'sk.graceN': '{n} روز فرصت جبران',
    'sk.noGrace': 'بدون فرصت جبران',
    'sk.miss.reset': 'شروع دوباره از روز اول',
    'sk.miss.step_back': 'یک روز عقب برای هر غیبت',
    'col.sk.rewards': 'جایزه‌ها',
    'sk.rewardsN': '{n} روز جایزه',
    'sk.cycleN': 'دورهٔ {n} روزه',
    'sk.hiro.notice': 'این بخش مال خود این سرور است و به Hiro نیازی ندارد. صفحهٔ «زنجیره‌ها»ی Hiro جدا و دست‌نخورده می‌ماند.',
    'f.sk.event': 'رویداد شمارنده',
    'f.sk.event.hint': 'مثلاً «ورود به بازی». هر بازه فقط یک بار شمرده می‌شود.',
    'f.sk.periodDays': 'طول هر بازه (روز)',
    'f.sk.periodDays.hint': 'معمولاً ۱ روز.',
    'f.sk.cycle': 'طول دوره',
    'f.sk.cycle.hint': 'بعد از این تعداد روز، جایزه‌ها از روز اول تکرار می‌شوند. خالی یعنی بدون تکرار.',
    'f.sk.grace': 'روزهای فرصت جبران',
    'f.sk.grace.hint': 'چند روز غیبت بخشیده شود تا شمارش از دست نرود.',
    'f.sk.miss': 'اگر بیشتر غیبت کند',
    'sk.h.rewards': 'جایزهٔ هر روز',
    'sk.add': 'افزودن روز',
    'sk.empty': 'هنوز جایزه‌ای تعیین نشده.',
    'f.sk.day': 'روز',
    'filter.ev.from': 'از طرف',
    'ev.from.client': 'بازی',
    'ev.from.server': 'سرور',
    'ev.from.admin': 'کنسول',
    'col.ev.value': 'مقدار',
    'col.ev.from': 'از طرف',
    'col.ev.metadata': 'اطلاعات همراه',
    'f.es.timezone': 'منطقهٔ زمانی',
    'f.es.timezone.hint': 'روز و هفته با ساعت این منطقه شروع می‌شوند.',
    'f.es.weekStart': 'شروع هفته',
    'f.es.weekStart.hint': 'برای جایزه‌های هفتگی.',
    'es.now': 'اکنون در این منطقه: {time}',
    'f.es.retention': 'نگه‌داری گزارش (روز)',
    'f.es.retention.hint': 'رویدادهای قدیمی‌تر پاک می‌شوند. جمع هر بازیکن و پیشرفت جایزه‌ها می‌ماند.',
    'f.es.rate': 'سقف رویدادهای بازی برای هر بازیکن در دقیقه',
    'f.es.rate.hint': 'جمع همهٔ رویدادهایی که بازی می‌فرستد.',
    'claim.status.granted': 'داده شد',
    'claim.status.failed': 'خطا',
    'claim.status.waiting': 'منتظر دریافت',
    'claim.status.ready': 'در صف',
    'f.ev.value': 'مقدار',
    'f.ev.value.hint': 'برای شمارش ساده ۱ بگذارید.',
    'pp.record': 'ثبت رویداد',
    'pp.record.text': 'رویداد برای این بازیکن ثبت می‌شود و قانون‌های جایزه، متا و روزهای پیاپی مثل یک رویداد واقعی اجرا می‌شوند.',
    'pp.record.ok': 'ثبت',
    'pp.title': 'پیشرفت بازیکن',
    'pp.recorded': 'رویداد ثبت شد.',
    'pp.log': 'گزارش رویدادها',
    'pp.level': 'سطح {level}',
    'pp.meta': '{n} {name}',
    'pp.toNext': 'تا سطح {level}: {n} مانده',
    'pp.broken': 'قطع شده',
    'pp.today': 'امروز شمرده شد',
    'pp.streak': 'روز {n} · بهترین: {best}',
    'pp.reset': 'صفر کردن',
    'pp.reset.title': 'روزهای پیاپی صفر شود؟',
    'pp.reset.text': 'شمارش «{name}» برای این بازیکن از اول شروع می‌شود. جایزه‌های گرفته‌شده می‌مانند.',
    'pp.reset.done': 'صفر شد.',
    'pp.times': '{n} بار',
    'pp.noEvents': 'هنوز رویدادی ندارد.',
    'pp.levels': 'سطح',
    'pp.streaks': 'روزهای پیاپی',
    'pp.events': 'رویدادها',
    'pp.claims': 'جایزه‌های اخیر',
    'pp.readyN': '{n} منتظر',
    'claim.kind.event': 'رویداد',
    'claim.kind.level': 'سطح',
    'claim.kind.streak': 'روز پیاپی',
    'pp.noClaims': 'هنوز جایزه‌ای نگرفته.',
    'ov.events': 'رویدادها و جایزه‌ها',
    'ov.eventsToday': 'رویدادهای ۲۴ ساعت اخیر',
    'ov.claimsReady': 'جایزه‌های منتظر',
    'ov.claimsFailed': 'جایزه‌های ناموفق',
    'nav.tasks': "تسک‌ها",
    'nav.task-groups': "گروه‌های تسک",
    'desc.tasks': "کارهایی که بازیکن در بازی انجام می‌دهد و برایش جایزه می‌گیرد، مثل «امروز سه دور بازی کن». پیشرفت هر تسک از رویدادها می‌آید.",
    'desc.task-groups': "چند تسک که به بازیکن داده می‌شوند؛ یکی‌یکی یا چندتا با هم، به ترتیب یا تصادفی.",
    'new.tasks': "تسک جدید",
    'edit.tasks': "ویرایش تسک",
    'new.task-groups': "گروه تسک جدید",
    'edit.task-groups': "ویرایش گروه تسک",
    'empty.hint.tasks': "مثلاً: «امروز سه دور بازی کن» با ۲۰ سکه جایزه.",
    'empty.hint.task-groups': "مثلاً یک زنجیرهٔ سه‌تایی که تسک‌ها را به ترتیب می‌دهد.",
    'col.tk.when': "شرط",
    'col.tk.place': "جایگاه",
    'col.tk.done': "انجام‌شده",
    'col.tg.rule': "روش دادن",
    'col.tg.tasks': "تسک‌ها",
    'filter.tk.level': "نوع",
    'filter.tk.group': "گروه",
    'filter.tg.handout': "روش دادن",
    'tk.level.top': "تسک اصلی",
    'tk.level.sub': "زیرتسک",
    'tk.when.count': "«{event}» {n} بار",
    'tk.when.sum_value': "جمع «{event}» به {n} برسد",
    'tk.when.high_value': "یک «{event}» با مقدار {n} یا بیشتر",
    'tk.when.children': "وقتی همهٔ زیرتسک‌ها انجام شوند",
    'tk.filter.min': "فقط مقدار {n} به بالا",
    'tk.filter.max': "فقط مقدار تا {n}",
    'tk.filter.range': "فقط مقدار {min} تا {max}",
    'tk.repeat.once': "یک بار",
    'tk.repeat.daily': "هر روز",
    'tk.repeat.weekly': "هر هفته",
    'tk.repeat.interval': "هر {time} یک بار",
    'tk.repeatopt.once': "یک بار",
    'tk.repeatopt.daily': "هر روز از نو",
    'tk.repeatopt.weekly': "هر هفته از نو",
    'tk.repeatopt.interval': "بعد از یک فاصلهٔ زمانی",
    'tk.max': "حداکثر {n} بار",
    'tk.minutes': "{n} دقیقه",
    'tk.hours': "{n} ساعت",
    'tk.days': "{n} روز",
    'tk.noPrize': "بدون جایزه",
    'tk.sub': "زیرتسک",
    'tk.inGroup': "گروه",
    'tk.childrenN': "{n} زیرتسک",
    'tk.depsN': "بعد از {n} تسک دیگر",
    'tk.standalone': "مستقل",
    'tk.doneN': "{n} بار · {p} بازیکن",
    'f.tk.key.hint': "بازی با این کلید تسک را می‌شناسد و بعداً عوض نمی‌شود.",
    'f.tk.name': "نام تسک",
    'f.tk.name.hint': "همین نام در بازی دیده می‌شود.",
    'f.tk.description': "توضیح برای بازیکن",
    'tk.h.when': "کی انجام می‌شود",
    'f.tk.condition': "نوع شرط",
    'f.tk.condition.hint': "یا با یک رویداد پیش می‌رود، یا با تمام شدن زیرتسک‌هایش.",
    'tk.condition.event': "با رویداد",
    'tk.condition.children': "با زیرتسک‌ها",
    'f.tk.metric': "شمارش",
    'f.tk.metric.hint': "تعداد رویدادها، جمع مقدارها یا بیشترین مقدار.",
    'tk.metric.count': "تعداد دفعه‌ها",
    'tk.metric.sum_value': "جمع مقدارها",
    'tk.metric.high_value': "بیشترین مقدار",
    'f.tk.target': "هدف",
    'f.tk.valueMin': "کمترین مقدار قابل‌قبول",
    'f.tk.valueMax': "بیشترین مقدار قابل‌قبول",
    'f.tk.valueFilter.hint': "رویدادی که مقدارش بیرون این بازه باشد شمرده نمی‌شود.",
    'f.tk.parent': "تسک اصلی",
    'f.tk.parent.hint': "اگر این تسک بخشی از یک تسک بزرگ‌تر است، آن را انتخاب کنید. زیرتسک تکرار، زمان و مخاطبِ تسک اصلی را دارد.",
    'tk.h.repeat': "تکرار",
    'f.tk.repeat': "تکرار",
    'f.tk.repeatMinutes': "فاصله (دقیقه)",
    'f.tk.repeatMinutes.hint': "بعد از انجام شدن، این مدت صبر می‌کند تا دوباره شروع شود.",
    'f.tk.maxCompletions': "حداکثر دفعه‌ها",
    'f.tk.maxCompletions.hint': "خالی یعنی بی‌حد.",
    'f.tk.group': "گروه",
    'f.tk.group.hint': "تسکِ گروه فقط وقتی دیده می‌شود که گروه آن را به بازیکن بدهد.",
    'f.tk.dependsOn': "بعد از این تسک‌ها",
    'f.tk.dependsOn.hint': "تا این تسک‌ها انجام نشوند، این تسک قفل است.",
    'f.tk.prize.hint': "بدون جایزه هم می‌شود؛ مثلاً برای زیرتسک‌ها.",
    'f.tk.notify': "خبر دادن در بازی",
    'f.tk.notify.hint': "وقتی تسک تمام شد، یک پیام درون‌برنامه‌ای برای بازیکن می‌رود.",
    'f.tk.levelTable': "جدول سطح",
    'f.tk.levelTable.hint': "برای محدود کردن بر اساس سطح بازیکن.",
    'f.tk.minLevel': "از سطح",
    'f.tk.maxLevel': "تا سطح",
    'f.tg.name': "نام گروه",
    'tg.h.handout': "دادن تسک‌ها",
    'f.tg.activeCount': "تسک‌های هم‌زمان",
    'f.tg.activeCount.hint': "بازیکن هم‌زمان چند تسک از این گروه دارد.",
    'f.tg.orderMode': "ترتیب",
    'tg.order.ordered': "به ترتیب",
    'tg.order.random': "تصادفی",
    'f.tg.handout': "تسک بعدی کی بیاید",
    'f.tg.handout.hint': "بعد از گرفتن جایزهٔ تسک قبلی.",
    'tg.handout.auto': "تسک بعدی خودکار",
    'tg.handout.request': "تسک بعدی با درخواست بازیکن",
    'tg.handout.timer.after': "تسک بعدی {time} بعد",
    'tg.handoutopt.auto': "خودکار",
    'tg.handoutopt.request': "با درخواست بازیکن",
    'tg.handoutopt.timer': "بعد از مدتی",
    'f.tg.handoutMinutes': "صبر (دقیقه)",
    'f.tg.repeatable': "تکرار گروه",
    'f.tg.repeatable.hint': "وقتی همهٔ تسک‌ها تمام شد، گروه از اول شروع می‌شود.",
    'tg.h.tasks': "تسک‌های گروه",
    'f.tg.tasks': "تسک‌ها",
    'f.tg.tasks.hint': "به همان ترتیبی که اضافه می‌کنید داده می‌شوند. فقط تسک‌های یک‌باره.",
    'tg.active': "هم‌زمان {n} تسک",
    'tg.repeatable': "تکرارشونده",
    'tg.tasksN': "{n} تسک",
    'tk.card': "تسک‌های بازیکن",
    'tk.all': "همهٔ تسک‌ها",
    'tk.none': "تسکی برای این بازیکن نیست.",
    'tk.hiddenN': "{n} تسک دیگر برای این بازیکن دیده نمی‌شود (خاموش، تمام‌شده یا مخصوص بازیکنان دیگر).",
    'tk.state.active': "در جریان",
    'tk.state.completed': "تمام شده",
    'tk.state.claimed': "جایزه گرفته",
    'tk.state.locked': "قفل",
    'tk.state.cooldown': "منتظر نوبت بعد",
    'tk.state.done': "به سقف رسید",
    'tk.state.hidden': "پنهان",
    'tk.state.finished': "انجام شد",
    'tk.reason.inactive': "خاموش",
    'tk.reason.ended': "زمانش تمام شده",
    'tk.reason.notStarted': "هنوز شروع نشده",
    'tk.reason.notTargeted': "برای این بازیکن نیست",
    'tk.reason.notHanded': "هنوز داده نشده",
    'tk.reason.parent': "تسک اصلی پنهان است",
    'tk.reason.groupClosed': "گروه بسته است",
    'tk.reason.groupInactive': "گروه خاموش است",
    'tk.progress': "{n} از {total}",
    'tk.progress.children': "{n} از {total} زیرتسک",
    'tk.completionsN': "{n} بار انجام شده",
    'tk.reset': "از نو",
    'tk.reset.title': "پیشرفت این تسک صفر شود؟",
    'tk.reset.text': "پیشرفت «{name}» برای این بازیکن پاک می‌شود. جایزه‌های گرفته‌شده می‌مانند.",
    'tk.reset.done': "پیشرفت صفر شد.",
    'tk.complete': "تمامش کن",
    'tk.complete.title': "این تسک تمام شود؟",
    'tk.complete.text': "«{name}» برای این بازیکن تمام‌شده حساب می‌شود و جایزه‌اش آماده می‌شود.",
    'tk.complete.done': "تسک تمام شد.",
    'tk.claim': "جایزه را بده",
    'tk.claim.done': "جایزه به بازیکن رسید.",
    'tg.round': "دور {n}",
    'tg.progress': "{n} از {total} انجام شده",
    'tg.nextAt': "بعدی: {date}",
    'tg.next': "تسک بعدی را بده",
    'tg.next.done': "تسک بعدی داده شد.",
    'tg.empty': "الان تسکی از این گروه ندارد.",
    'claim.kind.task': "تسک",
    'kind.task': "تسک",
    'code.INVALID_METRIC': "روش شمارش درست نیست.",
    'code.INVALID_CONDITION': "نوع شرط درست نیست.",
    'code.INVALID_REPEAT_MINUTES': "فاصله باید بین ۱ دقیقه و یک سال باشد.",
    'code.INVALID_MAX_COMPLETIONS': "حداقل ۱، یا خالی برای بی‌حد.",
    'code.HAS_SUBTASKS': "این تسک زیرتسک دارد؛ اول زیرتسک‌ها را پاک یا جابه‌جا کنید.",
    'code.SELF_PARENT': "تسک نمی‌تواند زیرتسکِ خودش باشد.",
    'code.NESTED_SUBTASK': "زیرتسک فقط یک لایه دارد.",
    'code.PARENT_NOT_GROUPING': "تسک اصلی باید «با زیرتسک‌ها» تمام شود.",
    'code.SUBTASK_NEEDS_EVENT': "زیرتسک باید با رویداد پیش برود.",
    'code.SUBTASK_IN_GROUP': "زیرتسک دنبال تسک اصلی است؛ تسک اصلی را در گروه بگذارید.",
    'code.GROUP_TASK_REPEAT': "تسکِ گروه در هر دورِ گروه یک بار انجام می‌شود؛ تکرار را «یک بار» بگذارید.",
    'code.SELF_DEPENDENCY': "تسک نمی‌تواند منتظر خودش باشد.",
    'code.TOO_MANY_DEPENDENCIES': "حداکثر ۱۰ تسک.",
    'code.DEPENDENCY_CYCLE': "این تسک‌ها منتظر همدیگر می‌مانند و هیچ‌وقت باز نمی‌شوند.",
    'code.LEVEL_TABLE_REQUIRED': "جدول سطح را انتخاب کنید.",
    'code.INVALID_LEVEL_RANGE': "«تا سطح» نباید کمتر از «از سطح» باشد.",
    'code.LEVEL_TABLE_NOT_FOUND': "جدول سطح پیدا نشد.",
    'code.TASK_NOT_FOUND': "تسک پیدا نشد.",
    'code.TASK_GROUP_NOT_FOUND': "گروه تسک پیدا نشد.",
    'code.INVALID_ACTIVE_COUNT': "بین ۱ تا ۲۰.",
    'code.INVALID_ORDER_MODE': "ترتیب درست نیست.",
    'code.INVALID_HANDOUT': "روش دادن درست نیست.",
    'code.INVALID_HANDOUT_MINUTES': "صبر باید بین ۱ دقیقه و یک سال باشد.",
    'code.TASK_NOT_AVAILABLE': "این تسک الان برای بازیکن باز نیست.",
    'code.TASK_ALREADY_COMPLETED': "این تسک قبلاً تمام شده.",
    'code.TASK_ALREADY_CLAIMED': "جایزهٔ این تسک قبلاً داده شده.",
    'code.TASK_NOT_STARTED': "بازیکن هنوز این تسک را شروع نکرده.",
    'code.TASK_NOT_COMPLETED': "این تسک هنوز تمام نشده.",
    'code.TASK_HAS_NO_PRIZE': "این تسک جایزه ندارد.",
    'code.CURRENT_TASK_MUST_BE_DONE': "اول تسک‌های فعلی باید تمام شوند.",
    'code.DURATION_NOT_ELAPSED': "هنوز وقتِ تسک بعدی نرسیده.",
    'code.NO_TASK_AVAILABLE': "الان تسکی برای دادن نیست.",
    'code.TASK_GROUP_FINISHED': "همهٔ تسک‌های این گروه تمام شده.",
    'code.HANDOUT_AUTOMATIC': "تسک‌های این گروه خودکار داده می‌شوند.",
    'section.sup': "پشتیبانی",
    'section.sup.short': "پشتیبانی",
    'nav.tickets': "تیکت‌ها",
    'nav.live-chat': "گفتگوی زنده",
    'nav.activity': "فعالیت بازیکنان",
    'nav.referrals': "کدهای معرف",
    'nav.remote-config': "پیکربندی بازی",
    'desc.tickets': "درخواست‌های پشتیبانی بازیکنان؛ تیکت‌های منتظر پاسخ اول نشان داده می‌شوند.",
    'desc.live-chat': "صف گفتگوها و گفتگوهای شما. یک گفتگو را بردارید تا به بازیکن جواب بدهید.",
    'desc.activity': "سابقهٔ کارهای هر بازیکن: ورود، خرید، جایزه، تیکت، محدودیت و بقیه.",
    'desc.referrals': "چه کسی چه کسی را دعوت کرده، وضعیت جایزه‌ها و موردهای مشکوک.",
    'desc.remote-config': "مقدارهایی که بازی هنگام اجرا می‌خواند؛ برای هر پلتفرم، نسخه، برچسب یا درصدی از بازیکنان جدا تنظیم می‌شوند.",
    'desc.remote-config.edit': "مقدار پیش‌فرض و قانون‌ها را تنظیم کنید. اولین قانونی که به بازیکن بخورد برنده است.",
    'desc.support-departments': "بخش‌هایی که بازیکن هنگام ساختن تیکت انتخاب می‌کند.",
    'desc.support-canned': "متن‌های آماده که پشتیبان با یک کلیک در پاسخ می‌گذارد.",
    'desc.support-settings': "محدودیت‌های تیکت، گفتگوی زنده و مدت نگه‌داشتن سابقهٔ فعالیت.",
    'desc.referral-rules': "جایزهٔ دعوت‌کننده و دعوت‌شده، شرط گرفتن جایزه و قانون‌های جلوگیری از سوءاستفاده.",
    'tab.support-departments': "بخش‌های پشتیبانی",
    'tab.support-canned': "پاسخ‌های آماده",
    'tab.support-settings': "تنظیمات پشتیبانی",
    'tab.referral-rules': "قوانین معرف",
    'new.support-departments': "بخش تازه",
    'edit.support-departments': "ویرایش بخش",
    'new.support-canned': "پاسخ آمادهٔ تازه",
    'edit.support-canned': "ویرایش پاسخ آماده",
    'empty.hint.support-departments': "یک بخش بسازید تا بازیکنان تیکت بفرستند.",
    'empty.hint.support-canned': "یک پاسخ آماده بسازید تا سریع‌تر جواب بدهید.",
    'empty.hint.tickets': "تیکتی با این وضعیت نیست.",
    'empty.hint.activity': "هنوز فعالیتی ثبت نشده است.",
    'empty.hint.referrals': "هنوز کسی کد معرف وارد نکرده است.",
    'empty.hint.remote-config': "یک کلید بسازید تا بازی آن را بخواند.",
    'search.tickets': "جست‌وجوی شماره، موضوع یا نام بازیکن",
    'search.activity': "جست‌وجو در شماره، آی‌پی یا جزئیات",
    'search.referrals': "جست‌وجوی نام بازیکن یا کد",
    'search.remote-config': "جست‌وجوی کلید یا توضیح",
    'tk.waitingN': "{n} تیکت منتظر پاسخ",
    'lc.waitingN': "{n} گفتگو در صف",
    'tk.status.open': "تازه",
    'tk.status.awaiting_staff': "منتظر پشتیبانی",
    'tk.status.awaiting_player': "منتظر بازیکن",
    'tk.status.resolved': "حل شده",
    'tk.status.closed': "بسته",
    'tk.priority': "اولویت",
    'tk.priority.low': "کم",
    'tk.priority.normal': "معمولی",
    'tk.priority.high': "زیاد",
    'tk.priority.urgent': "فوری",
    'tk.filter.waiting': "منتظر پاسخ ما",
    'tk.filter.active': "همهٔ باز‌ها",
    'tk.mine': "مال من",
    'tk.unassigned': "بی‌مسئول",
    'tk.overdue': "دیر شده",
    'tk.overdueOnly': "فقط دیرشده‌ها",
    'tk.by.player': "آخرین پیام از بازیکن",
    'tk.by.staff': "آخرین پیام از پشتیبانی",
    'tk.stat.waiting': "منتظر پاسخ",
    'tk.stat.overdue': "دیر شده",
    'tk.stat.unassigned': "بی‌مسئول",
    'tk.stat.mine': "مال من",
    'tk.stat.firstResponse': "میانگین اولین پاسخ",
    'tk.stat.30days': "در ۳۰ روز گذشته",
    'tk.stat.rating': "میانگین رضایت",
    'tk.stat.rated': "{n} امتیاز",
    'tk.new': "تیکت تازه",
    'tk.new.text': "برای بازیکنی که از راه دیگری تماس گرفته، تیکت بسازید. پیام شما اولین پیام تیکت است.",
    'tk.new.ok': "ساختن و فرستادن",
    'tk.new.hint': "بازیکن این پیام را می‌بیند و اعلان می‌گیرد.",
    'tk.pickPlayer': "بازیکن را پیدا کنید",
    'tk.subject': "موضوع",
    'tk.message': "پیام",
    'tk.detail': "تیکت",
    'tk.player': "بازیکن",
    'tk.staff': "پشتیبانی",
    'tk.system': "سیستم",
    'tk.note': "یادداشت داخلی",
    'tk.note.by': "یادداشت {name}",
    'tk.reply': "پاسخ",
    'tk.reply.ph': "پاسخ را بنویسید… (Ctrl+Enter برای فرستادن)",
    'tk.note.ph': "یادداشتی برای همکاران بنویسید…",
    'tk.canned': "پاسخ آماده",
    'tk.canned.pick': "گذاشتن پاسخ آماده…",
    'tk.canned.none': "پاسخ آماده‌ای نیست",
    'tk.attach': "پیوست",
    'tk.uploading': "در حال بارگذاری…",
    'tk.internal': "یادداشت داخلی (بازیکن نمی‌بیند)",
    'tk.send': "فرستادن",
    'tk.saveNote': "ذخیرهٔ یادداشت",
    'tk.after': "بعد از پاسخ",
    'tk.after.awaiting_player': "منتظر بازیکن",
    'tk.after.resolved': "حل شد",
    'tk.after.closed': "بستن",
    'tk.sent': "پاسخ فرستاده شد و بازیکن اعلان گرفت.",
    'tk.noteSaved': "یادداشت ذخیره شد.",
    'tk.closedNote': "این تیکت بسته است. بازیکن دیگر نمی‌تواند پاسخ بدهد؛ شما می‌توانید یادداشت بگذارید یا آن را دوباره باز کنید.",
    'tk.move.resolved': "حل شد",
    'tk.move.closed': "بستن تیکت",
    'tk.move.awaiting_staff': "باز کردن دوباره",
    'tk.moved': "وضعیت تیکت عوض شد.",
    'tk.assigned': "مسئول تیکت عوض شد.",
    'tk.takeIt': "خودم برمی‌دارم",
    'tk.playerCounts': "{tickets} تیکت · {chats} گفتگو",
    'tk.playerTickets': "تیکت‌های این بازیکن",
    'tk.times': "زمان‌ها",
    'tk.created': "ساخته شده",
    'tk.due': "مهلت پاسخ",
    'tk.firstResponse': "اولین پاسخ",
    'tk.notYet': "هنوز نه",
    'tk.resolvedAt': "حل شده در",
    'tk.closedAt': "بسته شده در",
    'tk.rating': "رضایت بازیکن",
    'tk.ratingN': "{n} از ۵",
    'tk.fromChat': "از گفتگوی زنده",
    'tk.openChat': "دیدن گفتگو",
    'tk.source.chat': "از گفتگوی زنده",
    'tk.source.staff': "ساختهٔ پشتیبانی",
    'tk.ev.status_changed': "{who}: {from} ← {to}",
    'tk.ev.status_changed.auto': "بستن خودکار پس از چند روز بی‌پاسخی",
    'tk.ev.assigned': "{who} تیکت را به {to} داد",
    'tk.ev.assigned.auto': "تیکت خودکار به {to} رسید",
    'tk.ev.assigned.none': "{who} مسئول تیکت را برداشت",
    'tk.ev.department_changed': "{who} بخش تیکت را عوض کرد",
    'tk.ev.priority_changed': "{who} اولویت را {to} کرد",
    'tk.ev.rated': "بازیکن {n} از ۵ امتیاز داد",
    'tk.ev.created_from_chat': "این تیکت از یک گفتگوی زنده ساخته شد",
    'col.subject': "موضوع",
    'col.assignee': "مسئول",
    'col.department': "بخش",
    'col.lastMessage': "آخرین پیام",
    'col.firstResponse': "مهلت اولین پاسخ",
    'col.agents': "پشتیبان‌ها",
    'col.open': "باز",
    'col.what': "چه شد",
    'col.where': "کجا",
    'col.referral': "دعوت‌کننده ← دعوت‌شده",
    'col.goal': "شرط جایزه",
    'col.value': "مقدار",
    'col.rules': "قانون‌ها",
    'col.version': "نسخه",
    'filter.priority': "اولویت",
    'filter.assignee': "مسئول",
    'filter.department': "بخش",
    'filter.overdue': "مهلت",
    'filter.platform': "پلتفرم",
    'filter.country': "کد کشور",
    'filter.from': "از تاریخ (۲۰۲۶-۱۱-۰۱)",
    'filter.to': "تا تاریخ",
    'filter.flagged': "مشکوک",
    'filter.referrerId': "شناسهٔ دعوت‌کننده",
    'filter.type': "نوع",
    'dur.minutes': "{n} دقیقه",
    'size.b': "{n} بایت",
    'size.kb': "{n} کیلوبایت",
    'size.mb': "{n} مگابایت",
    'lc.queue': "صف انتظار",
    'lc.mine': "گفتگوهای من",
    'lc.all': "همه",
    'lc.closed': "بسته‌شده‌ها",
    'lc.views': "نمایش گفتگوها",
    'lc.available': "آمادهٔ جواب دادن",
    'lc.nowOnline': "شما آنلاین هستید و بازیکنان را می‌بینید.",
    'lc.nowAway': "شما دور از دسترس هستید.",
    'lc.empty.queue': "کسی در صف نیست",
    'lc.empty.queue.hint': "گفتگوهای تازه اینجا می‌آیند.",
    'lc.empty.mine': "گفتگوی بازی ندارید",
    'lc.empty.all': "گفتگوی بازی نیست",
    'lc.empty.closed': "گفتگوی بسته‌ای نیست",
    'lc.take': "برداشتن",
    'lc.taken': "گفتگو مال شما شد.",
    'lc.pick': "یک گفتگو را انتخاب کنید",
    'lc.pick.hint': "از صف یکی را بردارید تا شروع کنید.",
    'lc.waitSec': "{n} ثانیه در صف",
    'lc.waitMin': "{n} دقیقه در صف",
    'lc.typing': "در حال نوشتن…",
    'lc.playerTypingNow': "بازیکن دارد می‌نویسد…",
    'lc.playerOnline': "بازیکن آنلاین است",
    'lc.playerOffline': "بازیکن آنلاین نیست",
    'lc.online': "آنلاین",
    'lc.offline': "آفلاین",
    'lc.activeN': "{n} گفتگوی باز",
    'lc.status.waiting': "در صف",
    'lc.status.active': "در جریان",
    'lc.status.closed': "بسته",
    'lc.withAgent': "با {name}",
    'lc.ticketN': "تیکت {n}",
    'lc.reply.ph': "پیام را بنویسید… (Ctrl+Enter برای فرستادن)",
    'lc.someoneElse': "این گفتگو دست {name} است.",
    'lc.closedNote': "این گفتگو بسته شده است.",
    'lc.closedRated': "گفتگو بسته شد. بازیکن {n} از ۵ امتیاز داد.",
    'lc.transfer': "سپردن به همکار",
    'lc.transfer.text': "گفتگو را به یک همکار بدهید یا به صف برگردانید.",
    'lc.transfer.ok': "سپردن",
    'lc.transferTo': "به چه کسی",
    'lc.backToQueue': "برگرداندن به صف",
    'lc.transferred': "گفتگو سپرده شد.",
    'lc.close': "بستن گفتگو",
    'lc.close.title': "این گفتگو بسته شود؟",
    'lc.close.text': "بازیکن پیامی می‌بیند که گفتگو بسته شد و می‌تواند امتیاز بدهد.",
    'lc.toTicket': "گفتگو به تیکت تبدیل شود",
    'lc.toTicket.hint': "متن گفتگو در یک تیکت تازه می‌ماند تا بعداً پیگیری شود.",
    'lc.closedDone': "گفتگو بسته شد.",
    'lc.closedTicket': "گفتگو بسته شد و تیکت {n} ساخته شد.",
    'lc.ev.accepted': "{agent} گفتگو را برداشت",
    'lc.ev.transferred': "گفتگو از {from} به {to} رسید",
    'lc.ev.closed': "گفتگو بسته شد ({by})",
    'lc.ev.note': "پیام سیستم",
    'code.ALREADY_TAKEN': "این گفتگو را {name} زودتر برداشت.",
    'code.NOT_YOUR_CHAT': "این گفتگو دست همکار دیگری است.",
    'code.CHAT_CLOSED': "این گفتگو بسته شده است.",
    'code.CHAT_EMPTY': "گفتگو هنوز پیامی ندارد که تیکت شود.",
    'code.CHAT_NOT_FOUND': "گفتگو پیدا نشد.",
    'code.TICKET_NOT_FOUND': "تیکت پیدا نشد.",
    'code.TICKET_CLOSED': "این تیکت بسته است.",
    'code.INVALID_TRANSITION': "این تغییر وضعیت ممکن نیست.",
    'code.INVALID_PRIORITY': "اولویت معتبر نیست.",
    'code.INVALID_SUBJECT': "موضوع ۱ تا ۱۵۰ حرف است.",
    'code.BODY_REQUIRED': "پیام را بنویسید.",
    'code.BODY_TOO_LONG': "پیام خیلی بلند است.",
    'code.ATTACHMENT_NOT_FOUND': "پیوست پیدا نشد.",
    'code.TOO_MANY_ATTACHMENTS': "پیوست‌ها از حد مجاز بیشتر است.",
    'code.ATTACHMENT_TOO_LARGE': "فایل بزرگ‌تر از حد مجاز است.",
    'code.ATTACHMENT_TYPE': "فقط عکس png، jpg، webp و فایل pdf مجاز است.",
    'code.ATTACHMENTS_UNAVAILABLE': "پیوست فعلاً در دسترس نیست.",
    'code.DEPARTMENT_NOT_FOUND': "این بخش پیدا نشد.",
    'code.INVALID_AGENT': "این همکار پیدا نشد.",
    'code.TOO_MANY_AGENTS': "حداکثر ۵۰ پشتیبان در هر بخش.",
    'code.INVALID_JSON': "متن JSON درست نیست.",
    'code.INVALID_PERCENT': "درصد بین ۰ تا ۱۰۰ است.",
    'code.VALUE_REQUIRED': "مقدار لازم است.",
    'code.VALUE_TOO_LONG': "مقدار خیلی بلند است.",
    'code.TOO_MANY_RULES': "حداکثر ۲۰ قانون.",
    'code.INVALID_RULE_NAME': "نام قانون کوتاه، با حروف کوچک انگلیسی و غیر از default باشد.",
    'code.DUPLICATE_RULE_NAME': "نام قانون‌ها باید فرق کند.",
    'code.MIN_ABOVE_MAX': "کمترین نسخه از بیشترین نسخه بالاتر است.",
    'code.TYPE_LOCKED': "نوع کلید عوض نمی‌شود؛ یک کلید تازه بسازید.",
    'code.VERSION_NOT_FOUND': "این نسخه در تاریخچه نیست.",
    'code.DESCRIPTION_TOO_LONG': "توضیح حداکثر ۵۰۰ حرف است.",
    'code.REFERRAL_NOT_FOUND': "این دعوت پیدا نشد.",
    'code.REFERRAL_DECIDED': "دربارهٔ این دعوت قبلاً تصمیم گرفته شده است.",
    'code.NOT_HELD': "فقط دعوت‌های نگه‌داشته‌شده تأیید می‌خواهند.",
    'dp.everyone': "همه",
    'cn.allDepts': "همهٔ بخش‌ها",
    'f.dp.key.hint': "شناسهٔ ثابت بخش، مثل payments.",
    'f.dp.name': "نام بخش",
    'f.dp.name.hint': "بازیکن این نام را در فرم تیکت می‌بیند.",
    'f.dp.description': "توضیح",
    'f.dp.description.hint': "یک جمله که بگوید چه سؤال‌هایی به این بخش می‌آید.",
    'f.dp.hours': "مهلت اولین پاسخ (ساعت)",
    'f.dp.hours.hint': "تیکت بعد از این مدت بی‌پاسخ «دیر شده» می‌شود.",
    'f.dp.agents': "پشتیبان‌های این بخش",
    'f.dp.agents.hint': "تیکت تازه خودکار به کم‌کارترین نفر از این فهرست می‌رسد. خالی یعنی بدون تقسیم خودکار.",
    'f.dp.active.hint': "بخش خاموش در فرم بازیکن نشان داده نمی‌شود.",
    'f.cn.key.hint': "شناسهٔ ثابت، مثل greeting.",
    'f.cn.title': "عنوان",
    'f.cn.title.hint': "فقط پشتیبان‌ها این عنوان را می‌بینند.",
    'f.cn.body': "متن پاسخ",
    'f.cn.body.hint': "به زبان تیکت در پاسخ گذاشته می‌شود؛ قبل از فرستادن می‌توانید آن را عوض کنید.",
    'f.cn.department': "فقط برای بخش",
    'f.cn.department.hint': "خالی یعنی برای همهٔ بخش‌ها.",
    'f.cn.active.hint': "پاسخ خاموش در فهرست پاسخ‌ها نمی‌آید.",
    'ss.tickets': "تیکت‌ها",
    'ss.chat': "گفتگوی زنده",
    'ss.activity': "سابقهٔ فعالیت",
    'f.ss.maxOpen': "حداکثر تیکت باز هر بازیکن",
    'f.ss.maxOpen.hint': "بعد از این تعداد، بازیکن باید یکی را ببندد.",
    'f.ss.perHour': "تیکت تازه در هر ساعت",
    'f.ss.perHour.hint': "جلوی فرستادن پشت‌سرهم را می‌گیرد.",
    'f.ss.messages': "پیام در هر ساعت",
    'f.ss.messages.hint': "برای پاسخ‌های بازیکن در همهٔ تیکت‌ها.",
    'f.ss.autoClose': "بستن خودکار پس از (روز)",
    'f.ss.autoClose.hint': "تیکت منتظر بازیکن یا حل‌شده بعد از این مدت بسته می‌شود.",
    'f.ss.attachments': "حداکثر پیوست هر پیام",
    'f.ss.attachments.hint': "۰ یعنی بدون پیوست.",
    'f.ss.attachmentKb': "حجم هر پیوست (کیلوبایت)",
    'f.ss.attachmentKb.hint': "فقط png، jpg، webp و pdf.",
    'f.ss.chatEnabled': "گفتگوی زنده روشن باشد",
    'f.ss.chatEnabled.hint': "خاموش باشد، بازیکن فقط تیکت می‌فرستد.",
    'f.ss.greeting': "پیام خوش‌آمد",
    'f.ss.greeting.hint': "اول هر گفتگو برای بازیکن نوشته می‌شود.",
    'f.ss.offline': "پیام وقتی کسی آنلاین نیست",
    'f.ss.offline.hint': "وقتی هیچ پشتیبانی آنلاین نیست، به بازیکن نشان داده می‌شود.",
    'f.ss.idle': "بستن گفتگوی بی‌حرکت پس از (دقیقه)",
    'f.ss.idle.hint': "وقتی نه پیامی آمده و نه بازیکن در صفحه است.",
    'f.ss.chatRate': "پیام بازیکن در هر دقیقه",
    'f.ss.chatRate.hint': "جلوی پیام‌های پشت‌سرهم را می‌گیرد.",
    'f.ss.retention': "نگه‌داشتن سابقهٔ فعالیت (روز)",
    'f.ss.retention.hint': "فعالیت‌های قدیمی‌تر هر چند دقیقه یک بار پاک می‌شوند.",
    'act.signin': "ورود",
    'act.purchase': "خرید",
    'act.prize': "جایزه",
    'act.event': "رویداد",
    'act.restriction': "محدودیت",
    'act.report': "گزارش",
    'act.ticket': "تیکت",
    'act.chat': "گفتگوی پشتیبانی",
    'act.referral': "معرف",
    'act.config': "پیکربندی",
    'act.method.device': "دستگاه",
    'act.method.custom': "شناسهٔ اختصاصی",
    'act.method.email': "ایمیل",
    'act.method.phone': "شمارهٔ موبایل",
    'act.method.google': "گوگل",
    'act.method.apple': "اپل",
    'act.method.facebook': "فیس‌بوک",
    'act.method.steam': "استیم",
    'act.method.gamecenter': "گیم‌سنتر",
    'act.signin.device': "با {method} وارد شد",
    'act.signin.custom': "با {method} وارد شد",
    'act.signin.email': "با {method} وارد شد",
    'act.signin.phone': "با {method} وارد شد",
    'act.signin.google': "با {method} وارد شد",
    'act.signin.apple': "با {method} وارد شد",
    'act.signin.facebook': "با {method} وارد شد",
    'act.signin.steam': "با {method} وارد شد",
    'act.signin.gamecenter': "با {method} وارد شد",
    'act.purchase.completed': "خرید {productKey} با {method}",
    'act.prize.granted': "جایزهٔ {prize} را گرفت",
    'act.prize.gift_redeemed': "کد هدیهٔ {code} را وارد کرد",
    'act.event.level_up': "به سطح {to} رسید",
    'act.restriction.suspend': "تعلیق شد تا {endsAt}",
    'act.restriction.mute': "بی‌صدا شد تا {endsAt}",
    'act.restriction.warn': "اخطار گرفت",
    'act.restriction.lifted': "{kind} برداشته شد",
    'act.report.filed': "یک پیام را گزارش کرد",
    'act.report.received': "پیامش گزارش شد",
    'act.ticket.created': "تیکت {number} را ساخت: {subject}",
    'act.ticket.created_by_staff': "پشتیبانی تیکت {number} را برایش ساخت",
    'act.ticket.staff_reply': "پشتیبانی به تیکت {number} پاسخ داد",
    'act.ticket.resolved': "تیکت {number} حل شد",
    'act.ticket.closed': "تیکت {number} بسته شد",
    'act.ticket.resolved_by_player': "خودش تیکت {number} را حل‌شده زد",
    'act.ticket.rated': "به تیکت {number} امتیاز {rating} داد",
    'act.chat.opened': "گفتگوی زنده باز کرد",
    'act.chat.taken': "{agent} گفتگو را برداشت",
    'act.chat.transferred': "گفتگو از {from} به {to} رسید",
    'act.chat.closed': "گفتگوی زنده بسته شد",
    'act.referral.entered': "کد معرف {referrer} را وارد کرد",
    'act.referral.joined': "{referee} با کد او آمد",
    'act.referral.completed': "شرط معرف کامل شد",
    'act.referral.approved': "دعوت تأیید شد",
    'act.referral.rejected': "دعوت رد شد",
    'act.config.changed': "پیکربندی بازی برایش عوض شد: {variants}",
    'act.actor.player': "خود بازیکن",
    'act.actor.system': "سیستم",
    'act.actor.console': "کنسول ({name})",
    'act.open': "باز کردن",
    'act.all': "همهٔ فعالیت‌ها",
    'act.card': "فعالیت اخیر",
    'act.empty': "فعالیتی نیست.",
    'act.playerLink': "فعالیت‌های این بازیکن",
    'act.retention': "سابقه {n} روز نگه داشته می‌شود.",
    'pf.android': "اندروید",
    'pf.ios': "iOS",
    'pf.web': "وب",
    'pf.windows': "ویندوز",
    'pf.macos': "مک",
    'pf.linux': "لینوکس",
    'rf.status.pending': "منتظر شرط",
    'rf.status.held': "نگه‌داشته",
    'rf.status.completed': "کامل",
    'rf.status.rejected': "رد شده",
    'rf.flag.burst': "دعوت زیاد در یک روز",
    'rf.flag.same_ip': "آی‌پی مشترک با دعوت‌کننده",
    'rf.flag.ip_burst': "چند دعوت از یک آی‌پی",
    'rf.flag.shared_device': "دستگاه تکراری",
    'rf.flag.shared_phone': "شمارهٔ تکراری",
    'rf.flaggedOnly': "فقط مشکوک‌ها",
    'rf.noGoal': "بدون شرط",
    'rf.progress': "{n} از {total}",
    'rf.tree': "درخت دعوت",
    'rf.tree.title': "درخت دعوت",
    'rf.tree.up': "دعوت‌کننده‌ها",
    'rf.tree.down': "{n} نفر را دعوت کرده",
    'rf.tree.more': "{n} دعوت",
    'rf.tree.none': "هنوز کسی را دعوت نکرده است.",
    'rf.approve': "تأیید",
    'rf.reject': "رد",
    'rf.approve.title': "این دعوت تأیید شود؟",
    'rf.approve.text': "دعوت {from} از {to}. بعد از تأیید، اگر شرط کامل باشد، جایزه‌ها داده می‌شوند.",
    'rf.approve.done': "دعوت تأیید شد.",
    'rf.reject.title': "این دعوت رد شود؟",
    'rf.reject.text': "دعوت {from} از {to} رد می‌شود و جایزه‌ای داده نمی‌شود.",
    'rf.reject.done': "دعوت رد شد.",
    'rf.stat.total': "همهٔ دعوت‌ها",
    'rf.stat.today': "{n} در ۲۴ ساعت گذشته",
    'rf.stat.flagged': "{n} مشکوک",
    'rf.top': "بیشترین دعوت‌ها",
    'rf.top.line': "{n} دعوت · {done} کامل",
    'rf.top.flagged': "{n} مشکوک",
    'claim.kind.referral': "معرف",
    'rr.rewards': "جایزه‌ها",
    'rr.rules': "قانون‌ها",
    'rr.measure.count': "تعداد بار",
    'rr.measure.sum': "جمع مقدار",
    'rr.measure.max': "بیشترین مقدار",
    'f.rr.enabled': "کد معرف روشن باشد",
    'f.rr.enabled.hint': "خاموش باشد، بازیکن نمی‌تواند کد وارد کند.",
    'f.rr.referrerPrize': "جایزهٔ دعوت‌کننده",
    'f.rr.referrerPrize.hint': "وقتی دوستش شرط را کامل کند.",
    'f.rr.refereePrize': "جایزهٔ دعوت‌شده",
    'f.rr.refereePrize.hint': "برای بازیکنی که کد را وارد کرده است.",
    'f.rr.goal': "رویداد شرط",
    'f.rr.goal.hint': "جایزه‌ها بعد از این رویداد داده می‌شوند. خالی یعنی همان لحظهٔ وارد کردن کد.",
    'f.rr.measure': "سنجه",
    'f.rr.goalValue': "مقدار لازم",
    'f.rr.goalValue.hint': "مثلاً ۳ یعنی سه بار انجام رویداد.",
    'f.rr.days': "فقط حساب‌های تازه (روز)",
    'f.rr.days.hint': "بازیکن تا این چند روز بعد از ساختن حساب می‌تواند کد وارد کند. ۰ یعنی بدون محدودیت.",
    'f.rr.max': "حداکثر دعوت هر بازیکن",
    'f.rr.max.hint': "بعد از این تعداد، کد او دیگر پذیرفته نمی‌شود.",
    'f.rr.daily': "دعوت عادی در روز",
    'f.rr.daily.hint': "بیشتر از این، دعوت مشکوک علامت می‌خورد.",
    'f.rr.device': "یک دعوت برای هر دستگاه",
    'f.rr.device.hint': "دستگاهی که قبلاً با کدی آمده، دوباره پذیرفته نمی‌شود.",
    'f.rr.phone': "یک دعوت برای هر شماره",
    'f.rr.phone.hint': "شماره‌ای که قبلاً با کدی آمده، دوباره پذیرفته نمی‌شود.",
    'f.rr.hold': "دعوت مشکوک نگه داشته شود",
    'f.rr.hold.hint': "جایزهٔ دعوت‌های مشکوک تا تأیید شما داده نمی‌شود.",
    'cf.new': "کلید تازه",
    'cf.edit': "ویرایش کلید",
    'cf.key': "کلید",
    'cf.key.hint': "نامی که بازی می‌خواند، مثل stargame.banner.",
    'cf.type': "نوع",
    'cf.type.hint': "بعد از ساختن عوض نمی‌شود.",
    'cf.type.locked': "نوع کلید ثابت است.",
    'cf.type.string': "متن ساده",
    'cf.type.number': "عدد",
    'cf.type.bool': "روشن/خاموش",
    'cf.type.json': "JSON",
    'cf.type.text': "متن ترجمه‌شده",
    'cf.description': "توضیح",
    'cf.description.hint': "برای همکاران؛ بازی آن را نمی‌بیند.",
    'cf.default': "مقدار پیش‌فرض",
    'cf.default.hint': "وقتی هیچ قانونی به بازیکن نخورد.",
    'cf.value': "مقدار",
    'cf.valueOn': "روشن",
    'cf.active': "کلید روشن است",
    'cf.on': "روشن",
    'cf.off': "خاموش",
    'cf.everyone': "برای همه",
    'cf.rulesN': "{n} قانون",
    'cf.v': "نسخهٔ {n}",
    'cf.rules': "قانون‌ها",
    'cf.rules.hint': "به ترتیب بررسی می‌شوند؛ اولین قانونی که همهٔ شرط‌هایش بخورد مقدارش را می‌دهد.",
    'cf.rule': "قانون",
    'cf.rule.add': "قانون تازه",
    'cf.rule.none': "قانونی نیست؛ همه مقدار پیش‌فرض را می‌گیرند.",
    'cf.rule.name': "نام قانون",
    'cf.rule.name.hint': "کوتاه و انگلیسی، مثل vip یا android-new.",
    'cf.rule.percent': "درصد بازیکنان",
    'cf.rule.percent.hint': "خالی یعنی همه. هر بازیکن همیشه در همان گروه می‌ماند.",
    'cf.rule.platforms': "پلتفرم",
    'cf.rule.platforms.hint': "هیچ‌کدام یعنی همهٔ پلتفرم‌ها.",
    'cf.rule.min': "از نسخهٔ",
    'cf.rule.max': "تا نسخهٔ",
    'cf.rule.tags': "برچسب بازیکن",
    'cf.rule.countries': "کشورها",
    'cf.rule.countries.hint': "کد دوحرفی، با ویرگول جدا.",
    'cf.rule.value': "مقدار این قانون",
    'cf.rule.value.hint': "بازیکنانی که شرط‌ها را دارند این را می‌گیرند.",
    'cf.rule.up': "بالاتر",
    'cf.rule.down': "پایین‌تر",
    'cf.delete.title': "این کلید حذف شود؟",
    'cf.delete.text': "«{key}» دیگر به بازی داده نمی‌شود. از تاریخچه می‌توانید برش گردانید.",
    'cf.history': "تاریخچه",
    'cf.recent': "تغییرهای اخیر",
    'cf.action.create': "ساخت",
    'cf.action.update': "ویرایش",
    'cf.action.delete': "حذف",
    'cf.action.rollback': "برگشت",
    'cf.rollback': "برگشت به این نسخه",
    'cf.rollback.title': "برگشت به نسخهٔ {n}؟",
    'cf.rollback.text': "مقدار و قانون‌های آن نسخه به صورت نسخهٔ تازه ذخیره می‌شوند.",
    'cf.rolledBack': "به نسخهٔ {n} برگشت.",
    'cf.preview': "امتحان برای یک بازیکن",
    'cf.preview.hint': "ببینید بازی برای این بازیکن و دستگاه چه مقدارهایی می‌گیرد.",
    'cf.preview.player': "شناسهٔ بازیکن",
    'cf.preview.player.ph': "خالی برای بازیکن ناشناس",
    'cf.preview.version': "نسخهٔ برنامه",
    'cf.preview.country': "کد کشور",
    'cf.preview.run': "امتحان کن",
    'cf.variant.default': "پیش‌فرض",
    'cf.bucket': "جایگاه {n}٪",
    'nav.items': "آیتم‌ها",
    'desc.items': "چیزهایی که بازیکن در بازی دارد، مثل پوسته، قاب و آواتار. بازیکن آن‌ها را می‌خرد، با رسیدن به سطح باز می‌کند یا جایزه می‌گیرد.",
    'new.items': "آیتم جدید",
    'edit.items': "ویرایش آیتم",
    'search.items': "جست‌وجو در کلید و نام آیتم",
    'kind.item': "آیتم",
    'kind.item_type': "نوع آیتم",
    'col.it.type': "نوع",
    'col.it.rarity': "کمیابی",
    'col.it.get': "راه گرفتن",
    'col.it.owners': "دارندگان",
    'filter.it.type': "نوع",
    'filter.it.rarity': "کمیابی",
    'rarity.common': "معمولی",
    'rarity.uncommon': "کمی کمیاب",
    'rarity.rare': "کمیاب",
    'rarity.epic': "حماسی",
    'rarity.legendary': "افسانه‌ای",
    'it.get.starter': "از اول دارد",
    'it.get.level': "از سطح {n}",
    'it.get.prize': "فقط با جایزه",
    'it.get.requires': "پیش‌نیاز: {key}",
    'it.upgradesN': "{n} سطح ارتقا",
    'it.ownersN': "{n} بازیکن",
    'it.equippedN': "{n} نفر انتخاب کرده‌اند",
    'f.it.type': "نوع آیتم",
    'f.it.type.hint': "بازیکن از هر نوع فقط یک آیتم را انتخاب می‌کند، مثلاً یک پوسته.",
    'f.it.rarity': "کمیابی",
    'h.it.get': "بازیکن چطور آن را می‌گیرد",
    'f.it.starter': "همه از اول دارند",
    'f.it.starter.hint': "به هر بازیکن خودکار داده می‌شود و نمی‌شود پسش گرفت.",
    'f.it.price.currency': "ارز قیمت",
    'f.it.price.amount': "قیمت",
    'f.it.price.hint': "خالی بگذارید تا آیتم فقط با جایزه یا سطح به دست بیاید.",
    'f.it.levelTable': "جدول سطح",
    'f.it.levelTable.hint': "خالی یعنی جدول سطح اصلی بازی.",
    'f.it.unlockLevel': "باز شدن از سطح",
    'f.it.unlockLevel.hint': "بدون قیمت: بازیکن در این سطح آن را رایگان باز می‌کند. با قیمت: از این سطح خریدنی می‌شود.",
    'f.it.parent': "آیتم پایه",
    'f.it.requires': "پیش‌نیاز",
    'f.it.requires.hint': "بازیکن اول باید این آیتم را داشته باشد.",
    'h.it.upgrades': "ارتقاها",
    'it.up.add': "سطح ارتقای تازه",
    'it.up.empty': "این آیتم ارتقا ندارد.",
    'f.it.up.currency': "ارز",
    'f.it.up.amount': "قیمت این سطح",
    'h.it.services': "خدمت‌ها",
    'it.sv.add': "خدمت تازه",
    'it.sv.empty': "این آیتم خدمتی ندارد.",
    'f.it.sv.key': "کلید",
    'f.it.sv.currency': "ارز",
    'f.it.sv.price': "قیمت هر واحد",
    'f.it.sv.max': "بیشترین واحد",
    'f.it.sv.name': "نام خدمت",
    'f.it.image': "تصویر بزرگ",
    'item.item': "آیتم",
    'item.item.once': "یک بار به بازیکن داده می‌شود",
    'h.pr.rules': "قانون‌های خرید",
    'f.maxLevel': "حداکثر سطح",
    'f.maxLevel.hint': "بازیکن بالاتر از این سطح، این محصول را نمی‌بیند.",
    'f.pr.levelTable': "جدول سطح",
    'f.pr.levelTable.hint': "خالی یعنی جدول سطح اصلی بازی.",
    'f.isOneTime': "فقط یک بار خریدنی",
    'f.isOneTime.hint': "بعد از خرید، از فروشگاه همان بازیکن برداشته می‌شود.",
    'f.checkContents': "اگر بازیکن همه را دارد، نفروش",
    'f.checkContents.hint': "وقتی بازیکن همهٔ آیتم‌های جایزهٔ این محصول را دارد، خرید رد می‌شود.",
    'f.purchaseMessage': "پیام بعد از خرید",
    'pr.oneTime': "یک بار",
    'pr.levels': "سطح {min} تا {max}",
    'pr.levelFrom': "از سطح {n}",
    'pr.levelTo': "تا سطح {n}",
    'inv.card': "انبار بازیکن",
    'inv.counts': "{n} از {total} آیتم",
    'inv.equipped': "انتخاب‌شده",
    'inv.slot': "{slot}: {name}",
    'inv.source.starter': "از اول",
    'inv.source.buy': "خرید",
    'inv.source.unlock': "با سطح",
    'inv.source.prize': "جایزه",
    'inv.source.admin': "از کنسول",
    'inv.upgrade': "ارتقای {n} از {max}",
    'inv.inactive': "غیرفعال",
    'inv.grant': "دادن آیتم",
    'inv.grant.pick': "آیتمی که می‌خواهید بدهید",
    'inv.grant.done': "آیتم به بازیکن داده شد.",
    'inv.revoke': "پس گرفتن",
    'inv.revoke.title': "آیتم پس گرفته شود؟",
    'inv.revoke.text': "«{name}» از انبار این بازیکن برداشته می‌شود و پولی برنمی‌گردد.",
    'inv.revoke.done': "آیتم پس گرفته شد.",
    'inv.none': "این بازیکن هنوز آیتمی ندارد.",
    'inv.all': "همهٔ آیتم‌ها",
    'act.item': "آیتم",
    'act.item.bought': "«{item}» را خرید",
    'act.item.unlocked': "«{item}» را با سطح باز کرد",
    'act.item.upgraded': "«{item}» را به ارتقای {level} رساند",
    'act.item.granted': "«{item}» از کنسول به او داده شد",
    'act.item.revoked': "«{item}» از کنسول پس گرفته شد",
    'act.item.starter': "آیتم‌های شروع را گرفت",
    'code.ITEM_NOT_FOUND': "آیتم پیدا نشد.",
    'nav.achievements': 'دستاوردها',
    'desc.achievements': 'نشان‌هایی که بازیکن با بازی کردن، تمام کردن تسک‌ها یا گرفتن جایزه باز می‌کند. هر دستاورد می‌تواند چند رده داشته باشد و هر رده جایزهٔ خودش را.',
    'new.achievements': 'دستاورد جدید',
    'edit.achievements': 'ویرایش دستاورد',
    'search.achievements': 'جست‌وجو در کلید و نام دستاورد',
    'empty.achievements': 'هنوز دستاوردی نساخته‌اید.',
    'kind.achievement': 'دستاورد',
    'claim.kind.achievement': 'دستاورد',
    'item.achievement': 'دستاورد',
    'item.achievement.once': 'دستاورد برای بازیکن باز می‌شود',
    'col.ac.when': 'کی باز می‌شود',
    'col.ac.tiers': 'رده‌ها',
    'col.ac.rarity': 'کمیابی و امتیاز',
    'col.ac.players': 'باز کرده‌اند',
    'filter.ac.source': 'راه باز شدن',
    'filter.ac.hidden': 'پنهان',
    'ac.source.event': 'با رویداد',
    'ac.source.task': 'با تسک',
    'ac.source.manual': 'با جایزه یا کنسول',
    'ac.when.count': '«{event}» {n} بار',
    'ac.when.sum_value': 'جمع مقدار «{event}» به {n} برسد',
    'ac.when.high_value': 'یک «{event}» با مقدار {n} یا بیشتر',
    'ac.when.task': 'تمام شدن تسک «{name}»',
    'ac.when.group': 'تمام شدن همهٔ تسک‌های «{name}»',
    'ac.when.manual': 'فقط با جایزه یا از کنسول',
    'ac.filter.min': 'فقط مقدار {n} یا بیشتر',
    'ac.filter.max': 'فقط مقدار {n} یا کمتر',
    'ac.filter.range': 'فقط مقدار {min} تا {max}',
    'ac.tiersN': '{n} رده',
    'ac.tier.n': 'ردهٔ {n}',
    'ac.tier.at': '{name}: {n}',
    'ac.pointsN': '{n} امتیاز',
    'ac.unlockedN': '{n} بازیکن',
    'ac.hidden': 'پنهان',
    'ac.secret': 'مخفی تا باز شدن',
    'ac.noPrize': 'بدون جایزه',
    'ac.h.look': 'ظاهر و ارزش',
    'ac.h.when': 'کی باز می‌شود',
    'ac.h.tiers': 'رده‌ها و جایزه‌ها',
    'ac.h.prize': 'جایزهٔ باز شدن',
    'f.ac.name': 'نام دستاورد',
    'f.hint': 'راهنمای پیش از باز شدن',
    'f.ac.hint.hint': 'وقتی دستاورد هنوز قفل است به بازیکن نشان داده می‌شود. برای دستاورد پنهان، تنها چیزی است که بازیکن می‌بیند.',
    'f.ac.rarity': 'کمیابی',
    'f.ac.points': 'امتیاز',
    'f.ac.points.hint': 'جمع امتیاز دستاوردهای بازیکن در پروفایلش دیده می‌شود.',
    'f.ac.hidden': 'پنهان تا باز شدن',
    'f.ac.hidden.hint': 'بازیکن نام و شرطش را نمی‌بیند تا وقتی بازش کند.',
    'f.ac.notify': 'پیام درون‌برنامه بعد از باز شدن',
    'f.ac.notify.hint': 'بعد از باز شدن و هر ردهٔ تازه، پیامی با دکمهٔ «دیدن دستاوردها» برای بازیکن می‌رود.',
    'f.ac.source': 'راه باز شدن',
    'f.ac.source.hint': 'هر دستاوردی را جایزه‌ها و کنسول هم می‌توانند بدهند.',
    'f.ac.metric': 'شمارش',
    'f.ac.task': 'تسک',
    'f.ac.task.hint': 'با تمام شدن این تسک باز می‌شود.',
    'f.ac.group': 'یا گروه تسک',
    'f.ac.group.hint': 'وقتی همهٔ تسک‌های یک دور از گروه تمام شود باز می‌شود.',
    'f.ac.icon': 'نشان',
    'f.ac.lockedIcon': 'نشان قفل',
    'f.ac.tier.name': 'نام رده',
    'f.ac.tier.threshold': 'هدف',
    'ac.tier.add': 'ردهٔ تازه',
    'ac.tier.empty': 'دست‌کم یک رده با هدفش لازم است.',
    'ac.tiers.hint': 'رده‌ها به ترتیب باز می‌شوند. هر ردهٔ بالاتر هدف بزرگ‌تری دارد و جایزه‌اش فقط یک بار داده می‌شود.',
    'pac.card': 'دستاوردهای بازیکن',
    'pac.counts': '{n} از {total} دستاورد · {p} امتیاز',
    'pac.state.locked': 'قفل',
    'pac.state.unlocked': 'باز شده',
    'pac.state.completed': 'همهٔ رده‌ها',
    'pac.state.hidden': 'پنهان',
    'pac.progress': '{n} از {total}',
    'pac.tier': 'ردهٔ {n} از {max}',
    'pac.by.event': 'با بازی',
    'pac.by.progress': 'با بازی',
    'pac.by.task': 'با تسک',
    'pac.by.prize': 'با جایزه',
    'pac.by.admin': 'از کنسول ({name})',
    'pac.at': '{date}',
    'pac.claim': 'جایزه را بده',
    'pac.claim.done': 'جایزهٔ رده داده شد.',
    'pac.revoke': 'پس گرفتن',
    'pac.revoke.title': 'دستاورد پس گرفته شود؟',
    'pac.revoke.text': '«{name}» و پیشرفتش برای این بازیکن پاک می‌شود. جایزه‌هایی که گرفته پس گرفته نمی‌شود و دوباره هم داده نمی‌شود.',
    'pac.revoke.done': 'دستاورد پس گرفته شد.',
    'pac.give': 'دادن دستاورد',
    'pac.give.pick': 'دستاوردی که می‌خواهید بدهید',
    'pac.give.tier': 'رده',
    'pac.give.done': 'دستاورد به بازیکن داده شد.',
    'pac.none': 'هنوز دستاوردی تعریف نشده.',
    'pac.all': 'همهٔ دستاوردها',
    'pac.log': 'سابقهٔ دستاوردها',
    'pac.waiting': '{n} جایزه منتظر',
    'act.achievement': 'دستاورد',
    'act.achievement.unlocked': 'دستاورد «{achievement}» را باز کرد',
    'act.achievement.tier': 'به ردهٔ {tier} از «{achievement}» رسید',
    'act.achievement.revoked': 'دستاورد «{achievement}» از کنسول پس گرفته شد',
    'code.ACHIEVEMENT_NOT_FOUND': 'دستاورد پیدا نشد.',
    'code.ACHIEVEMENT_LOCKED': 'این دستاورد هنوز باز نشده.',
    'code.ALREADY_UNLOCKED': 'بازیکن این دستاورد را در همین رده دارد.',
    'code.NOT_UNLOCKED': 'بازیکن این دستاورد را ندارد.',
    'code.TIER_NOT_REACHED': 'بازیکن هنوز به این رده نرسیده.',
    'code.TIER_NOT_FOUND': 'این رده وجود ندارد.',
    'code.TIER_HAS_NO_PRIZE': 'این رده جایزه ندارد.',
    'code.TIER_IN_USE': 'بازیکن‌ها جایزهٔ این رده را گرفته‌اند؛ نمی‌شود حذفش کرد.',
    'code.NOTHING_TO_CLAIM': 'جایزه‌ای برای گرفتن نمانده.',
    'code.TIERS_REQUIRED': 'دست‌کم یک رده با هدفش لازم است.',
    'code.TIERS_NEED_EVENT': 'فقط دستاوردی که با رویداد شمرده می‌شود چند رده دارد.',
    'code.TOO_MANY_TIERS': 'حداکثر ۱۰ رده.',
    'code.TIERS_NOT_CONSECUTIVE': 'رده‌ها باید پشت‌سرهم باشند.',
    'code.DUPLICATE_TIER': 'هر رده یک بار.',
    'code.INVALID_TIER': 'رده باید بین ۱ تا ۱۰ باشد.',
    'code.INVALID_POINTS': 'امتیاز باید بین ۰ تا ۱۰۰٬۰۰۰ باشد.',
    'code.INVALID_SOURCE': 'راه باز شدن درست نیست.',
    'code.TASK_REQUIRED': 'تسک یا گروه تسکی را که دستاورد را باز می‌کند انتخاب کنید.',
    'code.TASK_AND_GROUP': 'یا تسک را انتخاب کنید یا گروه تسک را، نه هر دو.',
    'ac.thresholds.order': 'هر ردهٔ بالاتر هدف بزرگ‌تری لازم دارد.',
    'code.ITEM_TYPE_NOT_FOUND': "نوع آیتم پیدا نشد.",
    'code.SELF_REFERENCE': "آیتم نمی‌تواند پیش‌نیاز خودش باشد.",
    'code.INVALID_RARITY': "کمیابی را از فهرست انتخاب کنید.",
    'code.STARTER_HAS_LEVEL': "آیتم شروع از اول باز است و سطح نمی‌خواهد.",
    'code.TOO_MANY_UPGRADES': "حداکثر ۵۰ سطح ارتقا.",
    'code.TOO_MANY_SERVICES': "حداکثر ۲۰ خدمت.",
    'code.DUPLICATE_SERVICE': "کلید خدمت تکراری است.",
    'code.INVALID_UNITS': "تعداد واحد درست نیست.",
    'code.ALREADY_OWNED': "بازیکن این آیتم را دارد.",
    'code.NOT_OWNED': "بازیکن این آیتم را ندارد.",
    'code.STARTER_ITEM': "آیتم شروع را نمی‌شود پس گرفت.",
    'code.LEVEL_TOO_LOW': "سطح بازیکن کافی نیست.",
    'code.LEVEL_TOO_HIGH': "سطح بازیکن از حد این محصول بالاتر است.",
    'code.PRODUCT_ALREADY_BOUGHT': "این محصول فقط یک بار خریدنی است و بازیکن آن را خریده.",
    'code.PRIZE_ALREADY_OWNED': "بازیکن همهٔ چیزهای این جایزه را دارد.",
    'code.ONE_TIME_SUBSCRIPTION': "اشتراک نمی‌تواند یک‌بار خرید باشد."
  },
  en: {
    'theme.toDark': 'Dark mode',
    'theme.toLight': 'Light mode',
    'section': 'Economy',
    'section.short': 'Econ',
    'nav.overview': 'Overview',
    'nav.currencies': 'Currencies',
    'nav.metas': 'Metas',
    'nav.prizes': 'Prizes',
    'nav.products': 'Shop',
    'nav.gifts': 'Gifts',
    'nav.payments': 'Payments',
    'tab.payment-methods': 'Payment gateways',
    'tab.sms': 'SMS',
    'tab.email': 'Email',
    'settings': 'Settings',
    'nav.categories': 'Categories',
    'nav.tags': 'Tags',

    'desc.overview': 'Your game economy at a glance.',
    'desc.currencies': 'In-game currencies and the starting amount each player gets.',
    'desc.metas': 'Numeric values per player, such as level or XP.',
    'desc.prizes': 'Bundles of currencies, metas and items given to players.',
    'desc.products': 'Products players buy with money or coins.',
    'desc.gifts': 'Gift and discount codes.',
    'desc.payments': 'Every purchase and its status.',
    'desc.payment-methods': 'Configure and enable gateways and stores.',

    'new.currencies': 'New currency',
    'new.metas': 'New meta',
    'new.prizes': 'New prize',
    'new.products': 'New product',
    'new.gifts': 'New gift',
    'new.categories': 'New category',
    'new.tags': 'New tag',
    'edit.currencies': 'Edit currency',
    'edit.metas': 'Edit meta',
    'edit.prizes': 'Edit prize',
    'edit.products': 'Edit product',
    'edit.gifts': 'Edit gift',
    'edit.categories': 'Edit category',
    'edit.tags': 'Edit tag',

    'search': 'Search by key or name…',
    'search.payments': 'Search by ID, reference or product…',
    'range': '{from} to {to} of {total}',
    'range.none': 'No items',
    'page.first': 'First page',
    'page.prev': 'Previous page',
    'page.next': 'Next page',
    'refresh': 'Refresh',
    'all': 'All',
    'none': 'None',
    'yes': 'Yes',
    'no': 'No',
    'create': 'Create',
    'save': 'Save',
    'saving': 'Saving…',
    'cancel': 'Cancel',
    'close': 'Close',
    'delete': 'Delete',
    'delete.confirm': 'Delete this?',
    'delete.yes': 'Yes, delete',
    'saved': 'Saved.',
    'deleted': 'Deleted.',
    'back': 'Back',
    'add.row': 'Add row',
    'remove': 'Remove',
    'upload': 'Upload',
    'uploading': 'Uploading…',
    'choose': 'Choose',
    'search.short': 'Search…',
    'no.results': 'No results.',
    'loading': 'Loading…',
    'optional': 'optional',

    'empty.title': 'Nothing here yet',
    'empty.search': 'Nothing matches this search.',
    'empty.hint.currencies': 'Create your first currency, such as coins or gems.',
    'empty.hint.metas': 'Create your first meta, such as player level.',
    'empty.hint.prizes': 'Products and gifts need prizes.',
    'empty.hint.products': 'Create a prize first, then a product that grants it.',
    'empty.hint.gifts': 'Create a gift or discount code for players.',
    'empty.hint.payments': 'No purchases yet.',
    'empty.hint.categories': 'Categories keep the shop organized.',
    'empty.hint.tags': 'Tags help group and filter items.',

    'err.network': 'Could not reach the server. Try again.',
    'err.forbidden': 'You do not have access to this section.',
    'err.notfound': 'This endpoint is not available. Is the module loaded?',
    'err.item.notfound': 'This item was not found.',
    'err.generic': 'Something went wrong. Try again.',
    'err.load': 'Could not load.',
    'err.required': 'This field is required.',
    'err.key': 'Lowercase letters, digits and _ only (2 to 32 characters).',
    'err.key.general': 'Start with a lowercase letter or digit, then letters, digits, _ . or - (up to 64).',
    'err.code': 'Use 3 to 32 characters: A-Z, digits, _ or -.',
    'err.number': 'Enter a valid number.',
    'err.json': 'This is not valid JSON.',
    'err.form': 'Some fields need fixing.',
    'err.upload': 'Upload failed. Use png, jpg, webp or svg up to 2 MB.',
    'try.again': 'Try again',

    'code.IN_USE': 'This item is used elsewhere and cannot be deleted.',
    'code.PAYMENT_METHOD_NOT_CONFIGURED': 'Fill in the required fields before enabling.',
    'code.DUPLICATE_KEY': 'This key is already taken.',
    'code.ALREADY_EXISTS': 'This value is already taken.',
    'code.NOT_FOUND': 'Not found.',
    'code.INVALID_STATE': 'Not possible in the current state.',
    'code.KEY_IMMUTABLE': 'The key cannot change after creation.',
    'code.REQUIRED': 'This field is required.',
    'code.INVALID_KEY': 'This key is not valid.',
    'code.INVALID_CODE': 'Use 3 to 32 characters: A-Z, digits, _ or -.',
    'code.CODE_REQUIRED': 'A discount needs a code.',
    'code.ITEMS_REQUIRED': 'Add at least one item.',
    'code.TOO_MANY_ITEMS': 'At most 50 items.',
    'code.INVALID_VALUE': 'This value is not valid.',
    'code.INVALID_ACTION': 'This action is not allowed for this item type.',
    'code.INVALID_ITEM_TYPE': 'Unknown item type.',
    'code.CURRENCY_NOT_FOUND': 'This currency does not exist.',
    'code.META_NOT_FOUND': 'This meta does not exist.',
    'code.PRIZE_NOT_FOUND': 'This prize does not exist.',
    'code.CATEGORY_NOT_FOUND': 'This category does not exist.',
    'code.TAG_NOT_FOUND': 'One of the tags no longer exists.',
    'code.REFERENCE_NOT_FOUND': 'A selected item no longer exists.',
    'code.PRICE_REQUIRED': 'Gateways need a price of at least 100 Toman.',
    'code.PRICE_TOO_LOW': 'The discounted price must be 0 or at least 100 Toman.',
    'code.INVALID_PRICE': 'This price is not valid.',
    'code.COIN_PRICE_REQUIRED': 'In-game coins need a coin price.',
    'code.SKU_REQUIRED': 'This store needs a SKU.',
    'code.INVALID_SKU': 'This SKU is not valid.',
    'code.SUBSCRIPTION_NOT_SUPPORTED': 'Myket does not support subscriptions.',
    'code.INVALID_DISCOUNT': 'This discount is not valid.',
    'code.INVALID_LIMIT': 'The limit must be at least 1.',
    'code.INVALID_WINDOW': 'The end must be after the start.',
    'code.INVALID_TIME': 'This time is not valid.',
    'code.INVALID_URL': 'The URL must start with http, https or /.',
    'code.INVALID_OBJECT': 'This must be a JSON object.',
    'code.OBJECT_TOO_LARGE': 'This JSON is too large.',
    'code.INVALID_NAME': 'A name is required (up to 128 characters).',
    'code.INVALID_KIND': 'Unknown kind.',
    'code.INVALID_TYPE': 'Unknown type.',
    'code.INVALID_PAYMENT_METHOD': 'Unknown payment method.',
    'code.PAYMENT_NOT_PENDING': 'Only pending payments can be checked.',
    'code.REWARD_ALREADY_GRANTED': 'The reward was already granted.',
    'code.PAYMENT_NOT_SUCCESSFUL': 'Only successful payments can be granted.',
    'code.PAYMENT_NOT_FOUND': 'This payment was not found.',
    'code.PAYMENT_METHOD_NOT_FOUND': 'This payment method was not found.',
    'code.GATEWAY_REQUEST_FAILED': 'The gateway did not answer. Try again shortly.',
    'code.STORE_UNAVAILABLE': 'The store is unreachable. Try again.',
    'code.STORE_CONFIGURATION_REJECTED': 'The store rejected the server credentials.',
    'code.UNSUPPORTED_FILE_TYPE': 'Only png, jpg, webp or svg.',
    'code.FILE_TOO_LARGE': 'Files must be at most 2 MB.',

    'f.key': 'Key',
    'f.key.hint': 'Stable ID used by the game code; do not change it later.',
    'f.key.locked': 'The key cannot change after creation.',
    'f.type': 'Type',
    'f.sort': 'Sort order',
    'f.sort.hint': 'Lower numbers show first.',
    'f.icon': 'Icon',
    'f.cover': 'Cover image',
    'f.metadata': 'Metadata (JSON)',
    'f.settings': 'Settings (JSON)',
    'f.defaultValue': 'Default value',
    'f.defaultValue.currency.hint': 'Given once to every player.',
    'f.defaultValue.meta.hint': 'Used when a player has no value yet.',
    'f.serverOnly': 'Server only',
    'f.serverOnly.hint': 'Hidden from the game client.',
    'f.category': 'Category',
    'f.tags': 'Tags',
    'f.isRepeatable': 'Repeatable',
    'f.isRepeatable.hint': 'When off, each player gets this prize at most once.',
    'f.items': 'Prize items',
    'f.items.empty': 'No items yet.',
    'f.item.type': 'Type',
    'f.item.key': 'Key',
    'f.item.action': 'Action',
    'f.item.value': 'Value',
    'f.isActive': 'Active',
    'f.priceToman': 'Price (Toman)',
    'f.priceToman.hint': 'Gateways need at least 100 Toman.',
    'f.coinPrice': 'Coin price',
    'f.coinPrice.currency': 'Currency',
    'f.coinPrice.amount': 'Amount',
    'f.discountPercent': 'Discount %',
    'f.paymentMethods': 'Payment methods',
    'f.sku.bazaar': 'Bazaar SKU',
    'f.sku.myket': 'Myket SKU',
    'f.isSubscription': 'Subscription',
    'f.isSubscription.hint': 'Supported on Bazaar only.',
    'f.prize': 'Prize',
    'f.prize.product.hint': 'Granted after a successful payment.',
    'f.minLevel': 'Minimum level',
    'f.gift.name': 'Name',
    'f.code': 'Code',
    'f.code.hint': 'Stored in upper case. Leave empty for no code.',
    'f.isPromo': 'Promo',
    'f.perUserLimit': 'Limit per player',
    'f.maxTotalUses': 'Total use limit',
    'f.maxTotalUses.hint': 'Empty means no limit.',
    'f.totalUsed': 'Used',
    'f.validFrom': 'Valid from',
    'f.expireAt': 'Expires at',
    'f.uploadUrl': 'File URL',

    'col.key': 'Key',
    'col.name': 'Name',
    'col.type': 'Type',
    'col.default': 'Default',
    'col.sort': 'Sort',
    'col.items': 'Items',
    'col.price': 'Price',
    'col.methods': 'Methods',
    'col.status': 'Status',
    'col.code': 'Code',
    'col.usage': 'Usage',
    'col.expire': 'Expires',
    'col.updated': 'Updated',
    'col.created': 'Time',
    'col.user': 'User',
    'col.product': 'Product',
    'col.method': 'Method',
    'col.amount': 'Amount',
    'col.reward': 'Reward',

    'type.soft': 'Soft',
    'type.hard': 'Hard',
    'kind.product': 'Product',
    'kind.prize': 'Prize',
    'kind.gift': 'Gift',
    'kind.currency': 'Currency',
    'gift.reward': 'Reward',
    'gift.discount': 'Discount',
    'item.currency': 'Currency',
    'item.meta': 'Meta',
    'item.storage': 'Storage item',
    'action.add': 'Add',
    'action.set': 'Set',
    'action.give': 'Give',
    'active': 'Active',
    'inactive': 'Inactive',
    'repeatable': 'Repeatable',
    'once': 'Once',
    'server.only': 'Server only',
    'unlimited': 'Unlimited',
    'toman': '{n} Toman',
    'percent': '{n}%',
    'amount.of': '{n} {name}',

    'method.zarinpal': 'ZarinPal',
    'method.zibal': 'Zibal',
    'method.bazaar': 'Bazaar',
    'method.myket': 'Myket',
    'method.in_app': 'In-game coins',
    'status.pending': 'Pending',
    'status.success': 'Success',
    'status.failed': 'Failed',
    'status.expired': 'Expired',

    'filter.status': 'Status',
    'filter.method': 'Method',
    'filter.user': 'User ID',
    'filter.active': 'Status',

    'pay.title': 'Payment',
    'pay.amounts': 'Amount',
    'pay.original': 'Original price',
    'pay.final': 'Final amount',
    'pay.rial': 'Amount in Rial',
    'pay.coin': 'Paid with coins',
    'pay.discount': 'Discount',
    'pay.gateway': 'Gateway and store',
    'pay.sandbox': 'Sandbox',
    'pay.authority': 'Authority',
    'pay.trackId': 'Track ID',
    'pay.refId': 'Reference ID',
    'pay.storeToken': 'Store token',
    'pay.cardMask': 'Card',
    'pay.outcome': 'Outcome',
    'pay.error': 'Error',
    'pay.reward': 'Reward',
    'pay.reward.granted': 'Granted',
    'pay.reward.notGranted': 'Not granted',
    'pay.reward.error': 'Reward error',
    'pay.inquiries': 'Inquiries',
    'pay.times': 'Timeline',
    'pay.createdAt': 'Created',
    'pay.updatedAt': 'Updated',
    'pay.verifiedAt': 'Verified',
    'pay.inquiredAt': 'Last inquiry',
    'pay.user': 'Player',
    'pay.product': 'Product',
    'pay.logs': 'Gateway log',
    'pay.logs.empty': 'No log entries.',
    'pay.inquire': 'Inquire now',
    'pay.inquire.done': 'Inquiry done.',
    'pay.retry': 'Retry reward',
    'pay.retry.done': 'Reward granted again.',
    'pay.request': 'Request',
    'pay.response': 'Response',
    'pay.duration': '{n} ms',
    'log.request': 'Request',
    'log.callback': 'Callback',
    'log.verify': 'Verify',
    'log.inquiry': 'Inquiry',
    'log.store_validate': 'Store check',

    'pm.configured': 'Configured',
    'pm.notConfigured': 'Not configured',
    'pm.active': 'Active',
    'pm.noFields': 'Nothing to configure; just turn it on or off.',
    'pm.secret.hint': 'Leave as is to keep the stored value.',
    'pm.empty': 'The server returned no payment methods.',
    'pm.field.merchantId': 'Merchant ID',
    'pm.field.merchant': 'Merchant',
    'pm.field.sandbox': 'Sandbox mode',
    'pm.field.packageName': 'Package name',
    'pm.field.apiSecret': 'API secret',
    'pm.field.accessToken': 'Access token',
    'pm.field.rsaPublicKey': 'RSA public key',

    'ov.catalog': 'Catalog',
    'ov.revenue30': 'Revenue, last 30 days',
    'ov.paymentsToday': 'Payments today',
    'ov.grantErrors': 'Rewards not granted',
    'ov.methods': 'Payment gateways',
    'ov.sandbox': 'Sandbox',
    'ov.payments': 'Payments',
    'ov.revenue': 'Revenue',
    'ov.open': 'Open',
    'ov.today': 'Today',
    'ov.total': 'Total',

    'section.pub': 'Publishing',
    'section.pub.short': 'Pub',
    'nav.faqs': 'FAQ',
    'nav.news': 'News',
    'nav.comments': 'Comments',
    'nav.notifications': 'Notifications',
    'desc.faqs': 'Questions and answers players see in the help section.',
    'desc.news': 'News and announcements shown in the game.',
    'desc.comments': 'Review and approve player comments on news.',
    'desc.notifications': 'Send in-game notifications, SMS and email to players.',
    'desc.sms': 'Set up SMS services. In sandbox mode no SMS is actually sent.',
    'desc.email': 'Set up email services. In sandbox mode no email is actually sent.',
    'new.faqs': 'New question',
    'new.news': 'New article',
    'edit.faqs': 'Edit question',
    'edit.news': 'Edit article',
    'search.faqs': 'Search questions or answers…',
    'search.news': 'Search titles or summaries…',
    'search.comments': 'Search comment text, username or article…',
    'search.notifications': 'Search title, text or ID…',
    'empty.hint.faqs': 'Write the questions players ask most, with their answers.',
    'empty.hint.news': 'Write your first article so players can see it in the game.',
    'empty.hint.comments': 'Nothing to review.',
    'empty.hint.notifications': 'You have not sent any notifications yet.',

    'kind.faq': 'FAQ',
    'kind.news': 'News',
    'filter.category': 'Category',
    'filter.pinned': 'Pinned',
    'filter.news': 'Article',
    'filter.channel': 'Channel',

    'col.question': 'Question',
    'col.category': 'Category',
    'col.title': 'Title',
    'col.comments': 'Comments',
    'col.published': 'Published',
    'col.comment': 'Comment',
    'col.news': 'Article',
    'col.actions': 'Actions',
    'col.notification': 'Notification',
    'col.channel': 'Channel',
    'col.target': 'Recipients',
    'col.delivery': 'Delivery',
    'col.address': 'Address',
    'col.error': 'Error',
    'col.attempts': 'Attempts',
    'col.time': 'Time',

    'f.faq.active.hint': 'Inactive questions are hidden in the game.',
    'f.publishedAt': 'Publish time',
    'f.publishedAt.hint': 'Empty means now. A future time keeps the article hidden until then.',
    'f.news.active.hint': 'Inactive articles are hidden in the game.',
    'f.isPinned': 'Pinned',
    'f.isPinned.hint': 'Stays at the top of the news list.',
    'f.comments': 'Comments',
    'f.allowComment': 'Comments open',
    'f.allowComment.hint': 'Players can comment on this article.',
    'f.freeComment': 'Publish without review',
    'f.freeComment.hint': 'When off, each comment must be approved on the Comments page first.',
    'f.actions': 'Article buttons',

    'md.write': 'Write',
    'md.preview': 'Preview',
    'md.mode': 'Editor mode',
    'md.empty': 'Nothing to preview.',
    'md.hint': 'Formatting:',
    'chars': '{n} / {max} characters',

    'act.add': 'Add button',
    'act.empty': 'No buttons yet.',
    'act.type': 'Type',
    'act.url': 'Open a link',
    'act.goto': 'Go to a game screen',
    'act.none': 'No button',
    'act.value': 'Target',
    'act.url.hint': 'The link opens in a new tab.',
    'act.goto.hint': 'A game screen name, like shop or play.',

    'news.pinned': 'Pinned',
    'news.notPinned': 'Not pinned',
    'news.scheduled': 'Scheduled',
    'news.commentsOff': 'Closed',

    'cm.pending': 'Awaiting review',
    'cm.accepted': 'Approved',
    'cm.pendingN': '{n} awaiting',
    'cm.accept': 'Approve',
    'cm.reject': 'Reject',
    'cm.reject.hint': 'The comment is hidden from players.',
    'cm.accepted.done': 'Comment approved.',
    'cm.rejected.done': 'Comment rejected.',
    'cm.reply': 'Reply',
    'cm.replies': '{n} replies',

    'ch.in_app': 'In-game',
    'ch.sms': 'SMS',
    'ch.email': 'Email',
    'nst.queued': 'Queued',
    'nst.sending': 'Sending',
    'nst.done': 'Done',
    'nst.failed': 'Failed',
    'rst.pending': 'Pending',
    'rst.sent': 'Sent',
    'rst.skipped': 'No address',
    'rst.failed': 'Failed',

    'tg.all': 'All users',
    'tg.users': 'Chosen users',
    'tg.group': 'A group',
    'tg.usersN': '{n} users',
    'tg.pickUsers': 'Users',
    'tg.searchUsers': 'Search and add users…',
    'tg.searchUsers.hint': 'Type a username, display name or ID.',
    'tg.users.hint': 'Up to 1,000 users.',
    'tg.pickGroup': 'Choose a group',
    'tg.group.hint': 'Sent to all group members, not to join requests.',
    'tg.all.hint': 'Sent to every user whose account is not disabled.',
    'tg.none': 'No recipients chosen yet.',
    'tg.withPhone': '{n} have a phone number',
    'tg.withEmail': '{n} have an email',
    'tg.skipped': '{n} have no address and will be skipped.',

    'nt.title': 'Notification',
    'nt.new': 'Send notification',
    'nt.new.desc': 'Write the message, choose recipients and send.',
    'nt.s.channel': 'Channel',
    'nt.s.content': 'Message',
    'nt.s.action': 'Notification button',
    'nt.s.target': 'Recipients',
    'nt.summary': 'Summary',
    'nt.f.title': 'Title',
    'nt.f.body': 'Body',
    'nt.f.provider': 'Delivery service',
    'nt.sms.hint': 'SMS has no title. Write plain text.',
    'nt.sms.tooLong': 'SMS text can be at most 600 characters.',
    'nt.action.hint': 'Optional. Tapping the notification opens the link or takes the player to that screen.',
    'nt.inapp.hint': 'Saved to the player\'s inbox and shown right away if the game is open.',
    'nt.noProvider': 'There is no active service for this channel.',
    'nt.goSettings.sms': 'SMS settings',
    'nt.goSettings.email': 'Email settings',
    'nt.sandbox.note': 'This service is in sandbox mode; nothing is actually sent.',
    'nt.live.note': 'This service is not in sandbox mode; messages really go to users.',
    'nt.recipientsWord': 'recipients',
    'nt.send': 'Send notification',
    'nt.sending': 'Sending…',
    'nt.queued': 'Notification queued for delivery.',
    'nt.confirmAll.title': 'Send to everyone?',
    'nt.confirmAll.text': 'This notification goes to all {n} users and cannot be taken back.',
    'nt.confirmAll.textNoCount': 'This notification goes to all users and cannot be taken back.',
    'nt.confirmAll.ok': 'Yes, send to everyone',
    'nt.confirmLive.title': 'Real delivery',
    'nt.confirmLive.text': 'The chosen service is not in sandbox mode, so messages really go out.',
    'nt.info': 'Details',
    'nt.content': 'Message',
    'nt.delivery': 'Delivery',
    'nt.recipientsList': 'Recipients',
    'nt.recipients': 'Recipients',
    'nt.seen': 'Seen',
    'nt.live': 'Refreshing automatically',
    'nt.action': 'Button',
    'nt.createdBy': 'Sent by',
    'nt.noRecipients': 'No recipients.',
    'nt.sentAt': 'Sent: {t}',
    'nt.seenAt': 'Seen: {t}',
    'nt.sentOf': '{sent} of {total}',
    'nt.skippedN': '{n} no address',
    'nt.failedN': '{n} failed',

    'pv.kind': 'Service type',
    'pv.kind.kavenegar': 'Kavenegar',
    'pv.kind.smsir': 'sms.ir',
    'pv.kind.twilio': 'Twilio',
    'pv.kind.smtp': 'SMTP',
    'pv.kind.resend': 'Resend',
    'pv.add': 'Add service',
    'pv.added': 'Service added. Sandbox mode is on.',
    'pv.allAdded': 'All services for this channel are added.',
    'pv.default': 'Default',
    'pv.empty.sms': 'No SMS service yet.',
    'pv.empty.email': 'No email service yet.',
    'pv.empty.hint': 'Add one from the list above.',
    'pv.sandbox': 'Sandbox mode',
    'pv.sandbox.short': 'Sandbox',
    'pv.sandbox.on': 'On: nothing is actually sent; messages are only written to the server log.',
    'pv.sandbox.off': 'Off: messages really go to users.',
    'pv.live.title': 'Turn off sandbox mode?',
    'pv.live.text': 'From now on this service sends real messages to users, and they cost money.',
    'pv.live.ok': 'Yes, turn it off',
    'pv.field.apiKey': 'API key',
    'pv.field.sender': 'Sender number',
    'pv.field.lineNumber': 'Line number',
    'pv.field.accountSid': 'Account SID',
    'pv.field.authToken': 'Auth token',
    'pv.field.from': 'From',
    'pv.field.host': 'Host',
    'pv.field.port': 'Port',
    'pv.field.security': 'Encryption',
    'pv.field.username': 'Username',
    'pv.field.password': 'Password',
    'pv.field.fromName': 'Sender name',
    'pv.hint.host': 'Host name only, without http or port.',
    'pv.hint.from': 'The number or address messages are sent from.',
    'pv.opt.starttls': 'STARTTLS',
    'pv.opt.tls': 'TLS',
    'pv.opt.none': 'None',
    'pv.test': 'Test send',
    'pv.test.hint': 'Sends one message with the saved settings. In sandbox mode nothing is sent.',
    'pv.test.to': 'To',
    'pv.test.subject': 'Subject',
    'pv.test.subject.ph': 'Test message',
    'pv.test.message': 'Text',
    'pv.test.message.default': 'This is a test message.',
    'pv.test.send': 'Send test',
    'pv.test.sandbox': 'Sandbox: nothing was sent. Reply:',
    'pv.test.ok': 'Sent.',
    'pv.test.failed': 'Not sent:',

    'code.TARGET_USERS_REQUIRED': 'Pick at least one user.',
    'code.USERS_NOT_FOUND': 'Some chosen users do not exist.',
    'code.TOO_MANY_USERS': 'At most 1,000 users.',
    'code.GROUP_NOT_FOUND': 'Group not found.',
    'code.INVALID_TARGET': 'The recipients are not valid.',
    'code.INVALID_CHANNEL': 'The channel is not valid.',
    'code.PROVIDER_NOT_FOUND': 'Service not found.',
    'code.PROVIDER_NOT_CONFIGURED': 'No active, complete service for this channel; fill in the required fields in Project settings.',
    'code.PROVIDER_INACTIVE': 'This service is not active.',
    'code.PROVIDER_CHANNEL_MISMATCH': 'This service is for another channel.',
    'code.PROVIDER_KIND_FIXED': 'The service type cannot change after creation.',
    'code.INVALID_PROVIDER_KIND': 'Unknown service type.',
    'code.INVALID_PHONE': 'Write the number like 09123456789 or with a country code.',
    'code.INVALID_EMAIL': 'The email address is not valid.',
    'code.INVALID_MESSAGE': 'Text is required (at most 600 characters).',
    'code.TOO_MANY_ACTIONS': 'At most 5 buttons.',
    'code.COMMENT_NOT_FOUND': 'Comment not found.',
    'code.NOTIFICATION_NOT_FOUND': 'Notification not found.',
    'code.INVALID_FILTER': 'The filter is not valid.',
    'code.QUERY_TOO_LONG': 'The search text is too long.',
    'code.URL_TOO_LONG': 'The URL is too long.',
    'code.PROVIDER_NOT_ALLOWED': 'In-game notifications do not use a delivery service.',
    'section.base': 'General',
    'section.base.short': 'General',
    'desc.categories': 'Categories for every section; each one belongs to a key type, such as product or news.',
    'desc.tags': 'Labels for grouping; a General tag can be used everywhere.',
    'tab.languages': 'Languages',
    'desc.languages': 'Languages the game content is translated into. The default language is always active.',
    'new.languages': 'New language',
    'edit.languages': 'Edit language',
    'search.languages': 'Search code or name…',
    'empty.hint.languages': 'Add your first language.',
    'filter.keyType': 'Key type',
    'col.keyType': 'Key type',
    'f.keyType': 'Key type',
    'f.keyType.category.hint': 'Where this category can be used.',
    'f.keyType.tag.hint': 'A General tag can be picked in every section.',
    'kind.general': 'General',
    'kind.user': 'Player',
    'ps.title': 'Project settings',
    'ps.sections': 'Project settings sections',
    'ps.noAccess': "You don't have access to any project settings section.",
    'ps.noAccess.hint': 'Ask a console admin to give you access.',
    'qc.add.categories': 'New category',
    'qc.add.tags': 'New tag',
    'qc.name': 'Name',
    'qc.name.ph.categories': 'e.g. Special packs',
    'qc.name.ph.tags': 'e.g. Best seller',
    'qc.hint': 'For {type}, in {lang}. The key is made automatically.',
    'qc.create': 'Create',
    'qc.err.name': 'Type a name.',
    'qc.created': '“{name}” was created and selected.',
    'f.name': 'Name',
    'f.message': 'Message',
    'f.description': 'Description',
    'f.question': 'Question',
    'f.answer': 'Answer',
    'f.title': 'Title',
    'f.summary': 'Summary',
    'f.body': 'Body',
    'act.label': 'Button text',
    'nt.f.sms': 'SMS text',
    'nt.action.label.hint': 'Leave it empty and the game shows its own text.',
    'tr.default': 'Default',
    'tr.empty': 'empty',
    'lang.default': 'Default',
    'lang.activated': '{name} is active.',
    'lang.deactivated': '{name} is inactive.',
    'col.language': 'Language',
    'col.direction': 'Direction',
    'dir.rtl': 'Right to left',
    'dir.ltr': 'Left to right',
    'f.lang.code': 'Language code',
    'f.lang.code.hint': 'A short lower-case code, like fa, en, ar or pt-br. It cannot change later.',
    'f.lang.nativeName': 'Native name',
    'f.lang.nativeName.hint': 'The name in its own script, like العربية.',
    'f.lang.name': 'Translated name',
    'f.lang.name.hint': 'The language name in other languages; the default language value is required.',
    'f.lang.direction': 'Direction',
    'f.lang.active.hint': 'Only active languages show in forms and in the game.',
    'f.lang.default': 'Default language',
    'f.lang.default.hint': 'Default text is required everywhere and shows when a translation is missing.',
    'cov.title': 'Translation coverage',
    'cov.desc': 'How many items have text but no translation in each language. Click a number to open that page.',
    'cov.resource': 'Section',
    'cov.total': 'Total',
    'cov.missingN': '{n} missing',
    'cov.complete': 'Complete',
    'cov.open': 'Open page',
    'cov.module.base': 'General',
    'cov.module.economy': 'Economy',
    'cov.module.publication': 'Publishing',
    'code.DEFAULT_LANGUAGE_REQUIRED': 'Text in the default language ({lang}) is required.',
    'code.INVALID_LANGUAGE': 'One of the languages no longer exists; close the form and open it again.',
    'code.TEXT_TOO_LONG': 'The {lang} text is too long.',
    'code.CATEGORY_KIND_MISMATCH': 'This category has a different key type.',
    'code.TAG_KIND_MISMATCH': 'A tag does not belong to this section; pick tags of this section or General.',
    'code.INVALID_KEY_TYPE': 'Unknown key type.',
    'code.INVALID_LANGUAGE_CODE': 'Use a lower-case language code, like fa, en or pt-br.',
    'code.LANGUAGE_NOT_FOUND': 'This language was not found.',
    'code.DEFAULT_LANGUAGE_LOCKED': 'The default language cannot be deleted or deactivated; make another language the default first.',
    'code.LANGUAGE_ACTIVE': 'An active language cannot be deleted; deactivate it first.',
    'code.INVALID_DIRECTION': 'Direction must be right to left or left to right.',
    'code.INVALID_NATIVE_NAME': 'A native name is required (up to 64 characters).',
    'code.DEFAULT_MUST_BE_ACTIVE': 'Only an active item can be the default.',
    'code.LANGUAGE_NOT_AVAILABLE': 'This language is not available.',
    'tab.brand': 'Brand',
    'tab.setup': 'Setup',
    'desc.brand': 'The console name, logo and main color. The sign-in page and games show them too.',
    'desc.setup': 'Open the console setup guide from here whenever you need it.',
    'brand.signIn': 'Sign in to continue to the {name} dashboard',
    'brand.f.name': 'Project name',
    'brand.f.name.hint': 'Shown in the sidebar, on the sign-in page and in the browser tab.',
    'brand.f.shortName': 'Short name',
    'brand.f.shortName.hint': 'For tight spaces; up to 24 characters.',
    'brand.f.logo': 'Logo for light mode',
    'brand.f.logo.hint': 'A wide logo on a transparent background works best.',
    'brand.f.logoDark': 'Logo for dark mode',
    'brand.f.logoDark.hint': 'Leave empty to use the light logo.',
    'brand.f.mark': 'Mark and favicon',
    'brand.f.mark.hint': 'A small square image for the browser tab and the collapsed sidebar.',
    'brand.f.accent': 'Main color',
    'brand.f.accent.hint': 'Used for buttons, links and the active menu item. Empty means the default color.',
    'brand.f.accent.pick': 'Pick a color',
    'brand.f.accent.presets': 'Preset colors',
    'brand.f.accent.reset': 'Default color',
    'brand.f.welcome': 'Sign-in welcome text',
    'brand.f.welcome.hint': 'Shown under the sign-in title.',
    'brand.contrast.onWhite': 'Button with white text',
    'brand.contrast.onDark': 'Link on a dark background',
    'brand.contrast.good': 'Readable',
    'brand.contrast.large': 'Large text only',
    'brand.contrast.poor': 'Low contrast',
    'brand.contrast.tip': 'Text in this color is hard to read. Pick a darker or stronger color.',
    'brand.preview': 'Preview',
    'brand.preview.light': 'Light',
    'brand.preview.dark': 'Dark',
    'brand.preview.nav': 'Dashboard',
    'brand.preview.link': 'A link',
    'brand.err.short': 'The short name can be at most 24 characters.',
    'brand.err.color': 'Enter a hex color like #6b59ee.',
    'code.INVALID_COLOR': 'Enter a hex color like #6b59ee.',
    'code.INVALID_STEP': 'This step does not exist.',
    'code.INVALID_STEP_STATUS': 'This step status is not valid.',
    'code.INVALID_STATUS': 'This setup status is not valid.',
    'setup.title': 'Console setup',
    'setup.subtitle': 'A few short steps to get the server ready',
    'setup.steps': 'Setup steps',
    'setup.later': 'Later',
    'setup.stay': 'Keep going',
    'setup.later.title': 'Finish the rest later?',
    'setup.later.text': 'Everything you saved stays. Continue any time from Project settings › Setup.',
    'setup.later.toast': 'Continue any time from Project settings › Setup.',
    'setup.unsaved.title': 'Unsaved changes',
    'setup.unsaved.text': 'If you leave this step, its changes are lost.',
    'setup.unsaved.ok': 'Leave without saving',
    'setup.back': 'Back',
    'setup.skip': 'Skip this step',
    'setup.saveNext': 'Save and continue',
    'setup.finish': 'Finish',
    'setup.done.toast': 'Setup is complete.',
    'setup.stepOf': 'Step {n} of {total}',
    'setup.step.brand': 'Brand',
    'setup.step.languages': 'Languages',
    'setup.step.currency': 'Currency',
    'setup.step.payment': 'Payments',
    'setup.step.sms': 'SMS',
    'setup.step.email': 'Email',
    'setup.step.review': 'Review',
    'setup.step.brand.title': 'Project brand',
    'setup.step.brand.desc': 'The name, logo and color that admins and players see.',
    'setup.step.languages.title': 'Game languages',
    'setup.step.languages.desc': 'Turn on the languages your game supports and pick a default.',
    'setup.step.currency.title': 'Main game currency',
    'setup.step.currency.desc': 'The coin or gem players spend, and the balance every new player starts with.',
    'setup.step.payment.title': 'In-game payments',
    'setup.step.payment.desc': 'The gateways and stores players pay through.',
    'setup.step.sms.title': 'Sending SMS',
    'setup.step.sms.desc': 'For sign-in codes and SMS notifications.',
    'setup.step.email.title': 'Sending email',
    'setup.step.email.desc': 'For account recovery, news and notifications.',
    'setup.step.review.title': 'Review',
    'setup.step.review.desc': 'Check everything once. Go back and fix anything that needs it.',
    'setup.st.done': 'Done',
    'setup.st.skipped': 'Skipped',
    'setup.st.pending': 'Not touched',
    'setup.status.pending': 'Not started',
    'setup.status.in_progress': 'In progress',
    'setup.status.completed': 'Completed',
    'setup.status.skipped': 'Postponed',
    'setup.no': 'No',
    'setup.err.answer': 'Answer this question first.',
    'setup.sandbox.on': 'On: everything is a test and no message or payment is real.',
    'setup.sandbox.off': 'Off: messages and payments are real.',
    'setup.pay.need': 'Do you need payments?',
    'setup.pay.no': 'OK. Turn them on any time from Project settings › Payment Gateways.',
    'setup.pay.pickOne': 'Pick at least one gateway.',
    'setup.pay.secret.sandbox': 'Not needed in test mode.',
    'setup.sms.need': 'Do you need SMS?',
    'setup.sms.no': 'OK. Add it any time from Project settings › SMS.',
    'setup.email.need': 'Do you need email?',
    'setup.email.no': 'OK. Add it any time from Project settings › Email.',
    'setup.pv.kind': 'Service',
    'setup.pv.kind.hint': 'This service becomes the default for the channel.',
    'setup.pv.new': 'This service does not exist yet. Saving creates it in test mode.',
    'setup.test': 'Test send',
    'setup.test.hint': 'Saves the settings first, then sends a test message. Nothing is really sent in test mode.',
    'setup.test.send': 'Save and send a test',
    'setup.test.liveBlocked': 'The setup guide only sends tests in test mode.',
    'setup.lang.list': 'Current languages',
    'setup.lang.active': '{name} is active',
    'setup.lang.makeDefault': 'Make default',
    'setup.lang.defaultSet': '{name} is now the default language.',
    'setup.lang.added': '{name} was added and turned on.',
    'setup.lang.quick': 'Quick add',
    'setup.lang.quick.hint': 'Tap a language to add it and turn it on.',
    'setup.lang.quick.none': 'All common languages are already added.',
    'setup.cur.pick': 'Main currency',
    'setup.cur.pick.hint': 'Pick the currency to edit, or create a new one.',
    'setup.cur.new': '+ New currency',
    'setup.cur.none': 'No currencies yet. Create the first one here.',
    'setup.cur.defaultValue': 'Starting balance',
    'setup.cur.defaultValue.hint': 'Every new player starts with this amount.',
    'setup.rv.brand.default': 'Default look',
    'setup.rv.brand.custom': 'Custom brand',
    'setup.rv.none': 'None',
    'setup.rv.sep': ', ',
    'setup.rv.balance': 'starting balance {n}',
    'setup.rv.edit': 'Edit',
    'setup.rv.editStep': 'Edit {step}',
    'setup.rv.left': '{n} steps are still untouched. You can finish now or go back.',
    'setup.rv.ready': 'Everything is ready.',
    'setup.open': 'Open the guide',
    'setup.progress': '{n} of {total} steps',
    'setup.nudge.title': 'Project setup is not finished',
    'setup.nudge.skipped': 'You left project setup for later',
    'setup.nudge.desc': 'A few short steps are left to set up the brand, languages, currency and payments.',
    'setup.nudge.go': 'Continue setup',
    'account': 'Account',
    'notFound': 'Page {path} was not found; you are back on the dashboard.',
    'setup.meta.opened': 'First opened',
    'setup.meta.completed': 'Completed',
    'tab.countries': 'Countries',
    'tab.accounts': 'Account rules',
    'tab.otp': 'SMS login',
    'tab.app-versions': 'App versions',
    'desc.countries': 'Countries players can pick, and whose dialling codes they can use to sign in.',
    'desc.app-versions': 'The latest and the lowest allowed version per platform. Anything older than the lowest version must update.',
    'desc.accounts': 'Rules for new accounts. The server checks usernames, passwords and display names against them.',
    'desc.otp': 'Latest code requests. The code itself is never shown.',
    'new.countries': 'New country',
    'edit.countries': 'Edit country',
    'new.app-versions': 'New version',
    'edit.app-versions': 'Edit version',
    'empty.hint.countries': 'No countries yet.',
    'empty.hint.app-versions': 'Add one row per platform so the game can check its version.',
    'empty.hint.otp': 'Nobody has asked for a login code yet.',
    'search.countries': 'Search name, code or dialling code',
    'search.app-versions': 'Search platform or version',
    'search.otp': 'Search number',
    'col.country': 'Country',
    'col.dialCode': 'Dialling code',
    'col.platform': 'Platform',
    'col.latest': 'Latest',
    'col.minimum': 'Lowest allowed',
    'col.stores': 'Stores',
    'col.phone': 'Number',
    'col.player': 'Player',
    'ct.activated': '{name} turned on',
    'ct.deactivated': '{name} turned off',
    'f.ct.code': 'Country code',
    'f.ct.code.hint': 'Two Latin letters, like IR or DE.',
    'f.ct.dialCode': 'Dialling code',
    'f.ct.dialCode.hint': 'Like +98. SMS login only accepts numbers from active countries.',
    'f.ct.name': 'Country name',
    'f.ct.flag': 'Flag image',
    'f.ct.flag.hint': 'Leave empty to use the flag made from the country code.',
    'f.ct.active.hint': 'Inactive countries are hidden in the game.',
    'f.ct.sort.hint': 'Lower numbers show first in the game list.',
    'code.INVALID_COUNTRY_CODE': 'A country code is two Latin letters, like IR.',
    'code.INVALID_DIAL_CODE': 'A dialling code is + and up to five digits, like +98.',
    'code.COUNTRY_NOT_FOUND': 'This country was not found.',
    'platform.android': 'Android',
    'platform.ios': 'iOS',
    'platform.web': 'Web',
    'platform.windows': 'Windows',
    'platform.macos': 'macOS',
    'platform.linux': 'Linux',
    'store.google_play': 'Google Play',
    'store.app_store': 'App Store',
    'store.bazaar': 'Cafe Bazaar',
    'store.myket': 'Myket',
    'store.microsoft_store': 'Microsoft Store',
    'store.steam': 'Steam',
    'store.web': 'Website',
    'store.direct': 'Direct download',
    'code.INVALID_VERSION': 'Write the version with numbers and dots, like 1.4.2.',
    'code.VERSION_REQUIRED': 'Enter the latest version.',
    'code.MIN_ABOVE_LATEST': 'The lowest version cannot be newer than the latest one.',
    'code.DUPLICATE_STORE': 'Each store can be listed once.',
    'code.INVALID_PLATFORM': 'This platform is not supported.',
    'code.INVALID_STORE': 'This store is not supported.',
    'code.TOO_MANY_LINKS': 'At most eight links.',
    'av.link.add': 'Add store link',
    'av.link.empty': 'No links yet. Players go to these links to update.',
    'av.link.store': 'Store',
    'av.link.url': 'Link',
    'av.link.label': 'Display name',
    'av.link.hint': 'An https link or the store\'s own link ({schemes}). The display name is optional.',
    'f.av.platform': 'Platform',
    'f.av.active.hint': 'When off, the game always gets "all good".',
    'f.av.latest': 'Latest version',
    'f.av.latest.hint': 'Older versions get the "update available" message.',
    'f.av.min': 'Lowest allowed version',
    'f.av.min.hint': 'Versions older than this must update (forced update).',
    'f.av.stores': 'Store links',
    'f.av.messages': 'Messages',
    'f.av.updateMessage': 'Update message',
    'f.av.updateMessage.hint': 'When a newer version exists but updating is optional.',
    'f.av.forceMessage': 'Forced update message',
    'f.av.forceMessage.hint': 'When the player\'s version is older than the lowest allowed one.',
    'otp.status.pending': 'Waiting',
    'otp.status.verified': 'Signed in',
    'otp.status.expired': 'Expired',
    'otp.status.failed': 'Too many tries',
    'otp.status.replaced': 'Replaced',
    'otp.status.send_failed': 'Not sent',
    'otp.sandbox': 'Sandbox',
    'otp.recent': 'Recent requests',
    'otp.code': 'Code',
    'otp.limits': 'Limits',
    'otp.message': 'SMS text',
    'otp.provider.default': 'Default SMS service',
    'otp.mode.off': 'SMS login is off.',
    'otp.mode.sandbox': 'Sandbox: no SMS is sent; the code is only written to the server log.',
    'otp.mode.noProvider': 'No active SMS service; codes cannot be sent.',
    'otp.mode.setup': 'Set up SMS',
    'otp.mode.live': 'Codes are sent with {name}.',
    'f.otp.enabled': 'SMS login is on',
    'f.otp.enabled.hint': 'When off, the game cannot ask for codes.',
    'f.otp.sandbox': 'Sandbox',
    'f.otp.sandbox.hint': 'No real SMS is sent; the code goes to the server log.',
    'f.otp.provider': 'SMS service',
    'f.otp.provider.hint': 'Add services in the SMS section.',
    'f.otp.codeLength': 'Code digits',
    'f.otp.codeLength.hint': 'Between 4 and 8.',
    'f.otp.ttl': 'Code lifetime (seconds)',
    'f.otp.ttl.hint': 'After this the code stops working.',
    'f.otp.attempts': 'Tries per code',
    'f.otp.attempts.hint': 'After this many wrong codes, a new code is needed.',
    'f.otp.resend': 'Wait before resending (seconds)',
    'f.otp.resend.hint': 'The player cannot ask for a new code before this.',
    'f.otp.phoneHourly': 'Codes per number per hour',
    'f.otp.phoneDaily': 'Codes per number per day',
    'f.otp.ipHourly': 'Codes per IP per hour',
    'f.otp.ipHourly.hint': 'Stops one place from asking for many codes.',
    'f.otp.message': 'SMS text',
    'f.otp.message.hint': '{code} is the code and must be in every language. {minutes} is the lifetime and {app} the brand name.',
    'code.DAILY_BELOW_HOURLY': 'The daily limit cannot be lower than the hourly limit.',
    'code.OTP_MESSAGE_NEEDS_CODE': 'Every language must contain {code}.',
    'code.PROVIDER_NOT_SMS': 'Choose an SMS service.',
    'code.INVALID_RULE': 'This number is not allowed.',
    'ar.username': 'Username',
    'ar.password': 'Password',
    'ar.display': 'Display name',
    'f.ar.usernameMin': 'Minimum length',
    'f.ar.usernameMax': 'Maximum length',
    'f.ar.charset': 'Allowed letters',
    'f.ar.charset.hint': 'With "Latin only", only A to Z, digits and _ . - are accepted.',
    'ar.charset.any': 'Any letters (Persian too)',
    'ar.charset.latin': 'Latin only',
    'f.ar.passwordMin': 'Minimum password length',
    'f.ar.passwordMin.hint': 'Nakama never accepts passwords shorter than 8.',
    'f.ar.passwordLetter': 'Password needs a letter',
    'f.ar.passwordDigit': 'Password needs a digit',
    'f.ar.displayNameMax': 'Maximum display name length',
    'code.MAX_BELOW_MIN': 'The maximum cannot be lower than the minimum.',
    'ar.try': 'Try it',
    'ar.try.hint': 'Type a username and a password to see if the numbers in this form accept them.',
    'ar.try.username': 'Sample username',
    'ar.try.password': 'Sample password',
    'ar.try.ok': 'Accepted.',
    'code.USERNAME_TOO_SHORT': 'The username needs at least {n} letters.',
    'code.USERNAME_TOO_LONG': 'The username can have at most {n} letters.',
    'code.USERNAME_INVALID_CHARACTERS': 'The username can only have Latin letters, digits and _ . -',
    'code.PASSWORD_TOO_SHORT': 'The password needs at least {n} characters.',
    'code.PASSWORD_NEEDS_LETTER': 'The password needs a letter.',
    'code.PASSWORD_NEEDS_DIGIT': 'The password needs a digit.',
    'section.mod': 'Players & chat',
    'section.mod.short': 'Chat',
    'nav.players': 'Player tags',
    'nav.reports': 'Reports',
    'nav.restrictions': 'Suspensions & mutes',
    'nav.chat-words': 'Word filter',
    'desc.players': 'See player tags, tag many players at once and find suspended or muted players.',
    'desc.reports': 'Messages that players reported or the filter flagged.',
    'desc.restrictions': 'Every suspension, mute and warning, with early lifting.',
    'desc.chat-words': 'Words that are hidden, refused or flagged for review in chat.',
    'desc.report-titles': 'Reasons a player picks when reporting a message.',
    'tab.report-titles': 'Report reasons',
    'new.chat-words': 'New word',
    'edit.chat-words': 'Edit word',
    'new.report-titles': 'New reason',
    'edit.report-titles': 'Edit report reason',
    'empty.hint.players': 'No player has signed up yet.',
    'empty.hint.reports': 'No new reports. They show up here.',
    'empty.hint.restrictions': 'Nobody has been suspended or muted yet.',
    'empty.hint.chat-words': 'Add the first word to keep chat clean.',
    'empty.hint.report-titles': 'Create a reason so players can report messages.',
    'search.players': 'Search username, name or ID',
    'search.reports': 'Search message text or player',
    'search.restrictions': 'Search player or reason',
    'search.chat-words': 'Search words',
    'rp.openN': '{n} open reports',
    'dur.hours': '{n} h',
    'dur.days': '{n} d',
    'dur.custom': 'Pick a date',
    'rs.until': 'until {date}',
    'rs.duration': 'How long',
    'rs.duration.suspend.hint': 'The suspension lifts by itself after this.',
    'rs.duration.mute.hint': 'After this the player can chat again.',
    'rs.reason': 'Message the player sees',
    'rs.reason.hint.suspend': 'Shown at sign-in together with the end time.',
    'rs.reason.hint.mute': 'Shown when the player tries to chat, with the end time.',
    'rs.reason.hint.warn': 'Sent to the player right away.',
    'rs.reason.default.suspend': 'Your account is closed for a while because you broke the game rules.',
    'rs.reason.default.mute': 'You cannot send messages for now because you broke the chat rules.',
    'rs.reason.default.warn': 'Please be respectful in chat. Next time you may be muted.',
    'rs.note': 'Internal note',
    'rs.note.ph': 'Only admins see this',
    'rs.note.hint': 'For your team; never shown to the player.',
    'rs.liftNote': 'Why lift it early?',
    'rs.kind.suspend': 'Suspension',
    'rs.kind.mute': 'Mute',
    'rs.kind.warn': 'Warning',
    'rs.status.active': 'Active',
    'rs.status.expired': 'Ended',
    'rs.status.lifted': 'Lifted',
    'rs.liftedAt': 'lifted {date}',
    'rs.do.suspend': 'Suspend',
    'rs.do.mute': 'Mute',
    'rs.do.warn': 'Warn',
    'rs.dialog.suspend': 'Suspend {name}',
    'rs.dialog.mute': 'Mute {name}',
    'rs.dialog.warn': 'Warn {name}',
    'rs.dialog.suspend.text': 'The player cannot sign in until it ends and is signed out now.',
    'rs.dialog.mute.text': 'The player cannot send chat messages until it ends.',
    'rs.dialog.warn.text': 'The player gets a warning message and it is kept in the history.',
    'rs.done.suspend': '{name} is suspended.',
    'rs.done.mute': '{name} is muted.',
    'rs.done.warn': 'Warning sent to {name}.',
    'rs.lift': 'Lift',
    'rs.lift.suspend': 'Lift suspension',
    'rs.lift.mute': 'Lift mute',
    'rs.lift.text': 'The restriction on {name} ends now.',
    'rs.lift.ok': 'Yes, lift it',
    'rs.lifted': 'Restriction lifted.',
    'code.END_REQUIRED': 'Pick when it ends.',
    'code.END_IN_PAST': 'The end must be in the future.',
    'code.END_TOO_FAR': 'At most ten years; use the Nakama ban for longer.',
    'code.INVALID_END': 'This date is not valid.',
    'code.INVALID_RESTRICTION_KIND': 'Unknown restriction type.',
    'code.NOT_ACTIVE': 'This restriction has already ended.',
    'code.REPORT_CLOSED': 'This report was already handled.',
    'code.INVALID_MOD_ACTION': 'This action is not allowed.',
    'code.INVALID_MATCH': 'Unknown match type.',
    'code.INVALID_WORD': 'A word is 1 to 64 characters with at least one letter.',
    'code.TAGS_REQUIRED': 'Pick at least one tag.',
    'code.TOO_MANY_TAGS': 'Too many tags.',
    'code.USERS_REQUIRED': 'Pick at least one player.',
    'code.USER_NOT_FOUND': 'Player not found.',
    'code.RESTRICTION_NOT_FOUND': 'Restriction not found.',
    'code.REPORT_NOT_FOUND': 'Report not found.',
    'code.MESSAGE_NOT_FOUND': 'Message not found.',
    'code.INVALID_FILTER': 'This filter is not valid.',
    'code.INVALID_OP': 'This action is not allowed.',
    'code.TARGET_TAGS_REQUIRED': 'Pick at least one player tag.',
    'tag.public': 'Visible to the player',
    'tag.playersN': '{n} players',
    'f.tag.public': 'The player can see this tag',
    'f.tag.public.hint': 'When on, the game can show this tag to the player, for example "VIP member".',
    'f.audience': 'Only for these players',
    'f.audience.hint': 'With tags picked, only players who have at least one of them see this product.',
    'audience.only': 'Tagged players only',
    'tg.tags': 'Player tags',
    'tg.tagsN': '{n} tags',
    'tg.pickTags': 'Tags',
    'tg.tags.hint': 'Players with at least one of these tags get the message.',
    'pl.banned': 'Banned',
    'pl.suspendedUntil': 'Suspended until {date}',
    'pl.mutedUntil': 'Muted until {date}',
    'pl.ok': 'No restrictions',
    'pl.status.suspended': 'Suspended',
    'pl.status.muted': 'Muted',
    'pl.status.banned': 'Banned',
    'pl.status.tagged': 'Tagged',
    'pl.select': 'Select {name}',
    'pl.bulk.tags': 'Tags',
    'pl.bulk.pick': 'Pick tags',
    'pl.bulk.add': 'Add tags',
    'pl.bulk.remove': 'Remove tags',
    'pl.bulk.page': 'Select this page',
    'pl.bulk.clear': 'Clear selection',
    'pl.bulk.selected': '{n} players selected',
    'pl.bulk.none': 'No player selected; use the first column.',
    'pl.bulk.added': 'Tags added to {n} players.',
    'pl.bulk.removed': 'Tags removed from {n} players.',
    'rp.title': 'Review report',
    'rp.gone': 'Deleted player',
    'rp.deleted': 'Message deleted',
    'rp.source.player': 'Player report',
    'rp.source.filter': 'Word filter',
    'rp.noTitle': 'No reason',
    'rp.byFilter': 'Automatic filter',
    'rp.reportsN': '{n} reports',
    'rp.thisMessage': 'Reported message',
    'rp.noContext': 'No conversation found.',
    'rp.conversation': 'Conversation',
    'rp.report': 'Report',
    'rp.reason': 'Reason',
    'rp.playerNote': 'Player comment',
    'rp.matched': 'Matched words',
    'rp.others': 'Other reports',
    'rp.channel': 'Room',
    'rp.sender': 'Sender',
    'rp.history': 'History',
    'rp.links': 'Links',
    'rp.seeHistory': 'Restriction history',
    'rp.seeReports': 'All reports about them',
    'rp.seeRestriction': 'See restriction',
    'rp.count.reports': '{n} reports',
    'rp.count.warn': '{n} warnings',
    'rp.count.mute': '{n} mutes',
    'rp.count.suspend': '{n} suspensions',
    'rp.decide': 'Decision',
    'rp.decision': 'Outcome',
    'rp.action': 'Action',
    'rp.action.dismiss': 'Dismiss',
    'rp.action.delete': 'Delete message',
    'rp.action.warn': 'Warn',
    'rp.action.mute': 'Mute',
    'rp.action.suspend': 'Suspend',
    'rp.action.dismiss.hint': 'Nothing is wrong; the report closes.',
    'rp.action.delete.hint': 'The message is removed from the chat.',
    'rp.action.warn.hint': 'The sender gets a warning.',
    'rp.action.mute.hint': 'The sender cannot chat until the time you pick.',
    'rp.action.suspend.hint': 'The sender cannot sign in until the time you pick.',
    'rp.alsoDelete': 'Also delete the message',
    'rp.note.hint': 'Kept for your team.',
    'rp.submit': 'Save decision',
    'rp.submit.dismiss': 'Dismiss report',
    'rp.submit.delete': 'Delete message',
    'rp.submit.warn': 'Send warning',
    'rp.submit.mute': 'Mute sender',
    'rp.submit.suspend': 'Suspend sender',
    'rp.done.dismiss': 'Report dismissed.',
    'rp.done.delete': 'Message deleted.',
    'rp.done.warn': 'Warning sent.',
    'rp.done.mute': 'Sender muted.',
    'rp.done.suspend': 'Sender suspended.',
    'rp.status.open': 'Open',
    'rp.status.dismissed': 'Dismissed',
    'rp.status.resolved': 'Handled',
    'rp.messageDeleted': 'Message deleted',
    'rp.restriction': 'Restriction',
    'rp.resolvedBy': 'Handled by',
    'rp.resolvedAt': 'Handled at',
    'cw.action.mask': 'Hide with stars',
    'cw.action.block': 'Refuse message',
    'cw.action.flag': 'Flag for review',
    'cw.match.word': 'Whole word',
    'cw.match.contains': 'Anywhere in text',
    'cw.try': 'Try the filter',
    'cw.try.hint': 'Type a sentence to see what players see. ZWNJ, tatweel, diacritics, Arabic ي and ك, spaced letters and repeated letters are caught too.',
    'cw.try.label': 'Sample sentence',
    'cw.try.ph': 'For example: you are an i.d.i.o.t',
    'cw.try.blocked': 'This message is not sent.',
    'cw.try.masked': 'Part of the message is hidden with stars.',
    'cw.try.flagged': 'The message is sent and goes to reports for review.',
    'cw.try.maskedFlagged': 'Part is hidden and the message goes to reports for review.',
    'cw.try.clean': 'This message is sent unchanged.',
    'cw.try.shown': 'What others see',
    'cw.try.notSent': 'Not sent',
    'cw.try.normalized': 'Normalised form:',
    'f.cw.word': 'Word or phrase',
    'f.cw.word.hint': 'Other spellings (Arabic ي, ZWNJ, repeated letters) are found automatically.',
    'f.cw.action': 'What happens',
    'f.cw.action.hint': 'Hide: stars. Refuse: the message is not sent. Flag: sent and queued for review.',
    'f.cw.match': 'Where to look',
    'f.cw.match.hint': 'Whole word also catches endings like -s or Persian «ها» and «ی». "Anywhere" catches words glued to others.',
    'f.cw.active.hint': 'An inactive word is ignored.',
    'f.cw.note': 'Note',
    'f.cw.note.hint': 'Only admins see this.',
    'f.rt.key.hint': 'A fixed id for the game, like spam.',
    'f.rt.name': 'Title',
    'f.rt.name.hint': 'Players see this in the list of reasons.',
    'f.rt.description': 'Description',
    'f.rt.description.hint': 'One short line under the title.',
    'f.rt.active.hint': 'Inactive reasons are not shown to players.',
    'pc.title': 'Tags and restrictions',
    'pc.tags': 'Tags',
    'pc.tags.edit': 'Edit tags',
    'pc.tags.edit.text': 'Player tags drive targeted messages and special offers.',
    'pc.tags.hint': 'Only tags of the "Player" type are listed.',
    'pc.tags.saved': 'Tags saved.',
    'pc.status': 'Status',
    'pc.history': 'History',
    'pc.historyN': '{n} items in history',
    'pc.reportsN': '{n} reports against them',
    'pc.lastWarn': 'Last warning: {date}',
    'filter.tag': 'Tag',
    'filter.kind': 'Type',
    'filter.playerId': 'Player ID',
    'filter.source': 'Source',
    'filter.reportTitle': 'Reason',
    'filter.action': 'Action',
    'col.select': 'Select',
    'col.tags': 'Tags',
    'col.kind': 'Type',
    'col.reason': 'Reason',
    'col.by': 'By',
    'col.message': 'Message',
    'col.reporter': 'Reporter',
    'col.word': 'Word',
    'col.action': 'Action',
    'col.match': 'Match',
    'col.used': 'Used',
    'nav.event-types': 'Events',
    'nav.event-rules': 'Event prizes',
    'nav.meta-rules': 'Meta rules',
    'nav.level-tables': 'Level tables',
    'nav.streaks': 'Daily streaks',
    'nav.event-log': 'Event log',
    'desc.event-types': 'Things recorded for each player, such as signing in, finishing a round or buying. Events not defined here are refused.',
    'desc.event-rules': 'When a player reaches a number in an event, they get a prize, either automatically or with a "Collect" button in the game.',
    'desc.meta-rules': 'How much each event adds to a player\'s metas, such as experience.',
    'desc.level-tables': 'When a meta reaches each threshold, the player levels up and that level\'s prize becomes ready.',
    'desc.streaks': 'Rewards for coming back or doing something several days in a row. It does not depend on Hiro.',
    'desc.event-log': 'Recent player events. Events older than the keep time are removed, but totals stay.',
    'desc.events': 'Time zone and first day of the week for daily and weekly counts, how long the event log is kept, and the event limit per player.',
    'tab.events': 'Events',
    'new.event-types': 'New event',
    'edit.event-types': 'Edit event',
    'new.event-rules': 'New event prize',
    'edit.event-rules': 'Edit event prize',
    'new.meta-rules': 'New meta rule',
    'edit.meta-rules': 'Edit meta rule',
    'new.level-tables': 'New level table',
    'edit.level-tables': 'Edit level table',
    'new.streaks': 'New daily streak',
    'edit.streaks': 'Edit daily streak',
    'empty.hint.event-types': 'Create your first event, such as finishing a round.',
    'empty.hint.event-rules': 'For example: play three rounds a day and get 20 coins.',
    'empty.hint.meta-rules': 'For example: add every round score to experience.',
    'empty.hint.level-tables': 'Create a table to turn experience into levels.',
    'empty.hint.streaks': 'For example, create a seven-day reward for signing in every day.',
    'empty.hint.event-log': 'No events have been recorded yet.',
    'search.event-log': 'Search by username or event key',
    'rs.when.count': '"{event}" {n} times {period}',
    'rs.when.sum': '"{event}" adds up to {n} {period}',
    'rs.when.max': 'Highest "{event}" reaches {n} {period}',
    'rs.period.total': 'in total',
    'rs.period.day': 'in one day',
    'rs.period.week': 'in one week',
    'rs.repeat.once': 'only once',
    'rs.repeat.period.day': 'once a day',
    'rs.repeat.period.week': 'once a week',
    'rs.repeat.period.total': 'once',
    'rs.repeat.every': 'for every {n}',
    'rs.max': 'at most {n} times',
    'grant.auto': 'Automatic',
    'grant.claim': 'Player collects',
    'ev.window.always': 'Always',
    'ev.window.from': 'From {date}',
    'ev.window.to': 'Until {date}',
    'rows.empty': 'No rows yet.',
    'code.DUPLICATE_LEVEL': 'Each level can appear only once.',
    'code.DUPLICATE_DAY': 'Each day can appear only once.',
    'code.THRESHOLDS_NOT_ORDERED': 'A higher level needs the same or a larger threshold.',
    'code.SAME_META': 'Write the level to a different meta.',
    'code.INVALID_RANGE': 'The largest value cannot be smaller than the smallest.',
    'code.REPEAT_NEEDS_PERIOD': '"Once per window" needs a daily or weekly window.',
    'code.EVERY_NEEDS_TOTAL': '"For every target" works with a count or a sum, not with the highest value.',
    'code.INVALID_EVENT_KEY': 'Use 2 to 48 lower-case letters, digits, _ or dots, starting with a letter.',
    'code.EVENT_TYPE_NOT_FOUND': 'This event is not defined or is turned off.',
    'code.EVENT_NOT_ALLOWED': 'Only the server records this event.',
    'code.VALUE_OUT_OF_RANGE': 'The value is outside the range allowed for this event.',
    'code.METADATA_TOO_LARGE': 'The event details are too large.',
    'code.TOO_MANY_EVENTS': 'Too many events; try again in a minute.',
    'code.SYSTEM_EVENT': 'Built-in events cannot be deleted; you can turn them off.',
    'code.SYSTEM_EVENT_SERVER_ONLY': 'Built-in events are recorded only by the server.',
    'code.INVALID_RATE_LIMIT': 'The limit must be between 1 and 100,000 per minute.',
    'code.INVALID_THRESHOLD': 'The target is not valid.',
    'code.INVALID_MAX_CLAIMS': 'At least 1, or empty for no limit.',
    'code.INVALID_GRANT_MODE': 'This way of giving the prize is not valid.',
    'code.INVALID_MEASURE': 'This way of counting is not valid.',
    'code.INVALID_PERIOD': 'This window is not valid.',
    'code.INVALID_REPEAT': 'This repeat is not valid.',
    'code.INVALID_META_ACTION': 'This meta change is not valid.',
    'code.INVALID_VALUE_SOURCE': 'This amount source is not valid.',
    'code.STEPS_REQUIRED': 'Add at least one level.',
    'code.TOO_MANY_STEPS': 'At most 200 levels.',
    'code.INVALID_LEVEL': 'Levels start at 1.',
    'code.INVALID_PERIOD_DAYS': 'Between 1 and 30 days.',
    'code.INVALID_GRACE': 'Between 0 and 30.',
    'code.INVALID_MISS_ACTION': 'This missed-day rule is not valid.',
    'code.INVALID_CYCLE': 'Between 1 and 365 days, or empty.',
    'code.TOO_MANY_REWARDS': 'At most 365 reward days.',
    'code.INVALID_DAY': 'Days start at 1.',
    'code.DAY_OUTSIDE_CYCLE': 'This day is after the end of the cycle.',
    'code.INVALID_TIMEZONE': 'This time zone is not valid.',
    'code.INVALID_WEEK_START': 'This first day of the week is not valid.',
    'code.INVALID_RETENTION': 'Between 1 and 3650 days.',
    'code.CLAIM_NOT_FOUND': 'This reward was not found.',
    'code.ALREADY_CLAIMED': 'This reward was already given.',
    'code.STREAK_NOT_FOUND': 'This player has not started this streak.',
    'filter.ev.source': 'Recorded by',
    'ev.source.client': 'Game too',
    'ev.source.server': 'Server only',
    'ev.system': 'Built-in',
    'col.ev.source': 'Recorded by',
    'col.ev.range': 'Value range',
    'ev.range.any': 'Any',
    'col.ev.rate': 'Limit',
    'ev.perMinute': '{n} per minute',
    'ev.usedBy': '{n} rules',
    'ev.system.notice': 'This is a built-in event recorded by the server. You can change its name, description and whether it is on.',
    'f.ev.key.hint': 'The key the game sends, such as round_finished. It cannot change later.',
    'f.ev.client': 'The game may record it',
    'f.ev.client.hint': 'On means the game client sends this event. Keep it off for important things like purchases or wins, so only the server records them.',
    'f.ev.valueMin': 'Smallest value',
    'f.ev.valueMax': 'Largest value',
    'f.ev.rate': 'Limit per player per minute',
    'f.ev.rate.hint': 'Above this, events from the game are refused. Server events have no limit.',
    'f.ev.active.hint': 'An inactive event is refused and its rules do not run.',
    'filter.ev.event': 'Event',
    'col.ev.when': 'Condition',
    'col.ev.prize': 'Prize',
    'col.ev.claims': 'Rewards',
    'ev.ended': 'Ended',
    'f.ev.rule.name': 'Name shown to players',
    'f.ev.rule.name.hint': 'Shown in the game next to the progress bar.',
    'f.ev.rule.description': 'Description for players',
    'ev.h.when': 'When',
    'f.ev.event': 'Event',
    'f.ev.measure': 'Count by',
    'f.ev.measure.hint': 'Number of times, sum of values or the highest single value.',
    'measure.count': 'Number of times',
    'measure.sum': 'Sum of values',
    'measure.max': 'Highest value',
    'f.ev.period': 'Window',
    'period.total': 'All time',
    'period.day': 'Each day',
    'period.week': 'Each week',
    'f.ev.threshold': 'Target',
    'f.ev.repeat': 'Repeat',
    'repeat.once': 'Only once',
    'repeat.period': 'Once per window',
    'repeat.every': 'For every target',
    'f.ev.maxClaims': 'Most rewards per player',
    'f.ev.maxClaims.hint': 'Empty means no limit.',
    'ev.h.prize': 'Prize',
    'f.ev.grantMode': 'How it is given',
    'f.ev.grantMode.hint': 'Automatic adds it right away; player collects means the player takes it with a button.',
    'ev.h.who': 'Who and when',
    'f.ev.audience': 'Only for these players',
    'f.ev.audience.hint': 'Empty means everyone. With tags, only players who have at least one of them.',
    'f.ev.startsAt': 'Starts',
    'f.ev.endsAt': 'Ends',
    'filter.ev.meta': 'Meta',
    'col.ev.event': 'Event',
    'col.ev.change': 'Change',
    'mr.change.add.fixed': '+ {n}',
    'mr.change.add.event': '+ value × {n}',
    'mr.change.set.fixed': '= {n}',
    'mr.change.set.event': '= value × {n}',
    'mr.change.max.fixed': 'at least {n}',
    'mr.change.max.event': 'highest value × {n}',
    'f.mr.sort.hint': 'Rules run in this order.',
    'f.mr.name': 'Name',
    'f.mr.name.hint': 'Helps you find the rule in the console.',
    'f.mr.meta': 'Meta',
    'f.mr.action': 'Change',
    'mr.action.add': 'Add',
    'mr.action.set': 'Set',
    'mr.action.max': 'Keep the highest',
    'f.mr.source': 'Amount from',
    'f.mr.source.hint': 'A fixed number, or the event value times the number below.',
    'mr.source.fixed': 'Fixed number',
    'mr.source.event': 'Event value',
    'f.mr.amount': 'Number',
    'f.mr.amount.hint': 'With "Event value" it is a multiplier; 1 means the round score as it is.',
    'col.lv.meta': 'Meta',
    'col.lv.steps': 'Levels',
    'lv.levelsN': '{n} levels',
    'lv.top': 'level {level} at {n}',
    'col.ev.players': 'Players',
    'f.lv.meta': 'Points meta',
    'f.lv.meta.hint': 'For example experience. The level is checked each time this meta changes.',
    'f.lv.levelMeta': 'Write the level to this meta',
    'f.lv.levelMeta.hint': 'For example a "level" meta, so the game and products with a minimum level can use it.',
    'lv.h.steps': 'Levels and prizes',
    'lv.add': 'Add level',
    'lv.empty': 'Add at least one level.',
    'f.lv.level': 'Level',
    'f.lv.threshold': 'Points needed',
    'f.lv.title': 'Level title',
    'col.sk.rule': 'Rule',
    'sk.everyDay': 'Every day',
    'sk.everyDays': 'Every {n} days',
    'sk.graceN': '{n} grace days',
    'sk.noGrace': 'No grace days',
    'sk.miss.reset': 'Start again from day one',
    'sk.miss.step_back': 'One day back per missed day',
    'col.sk.rewards': 'Rewards',
    'sk.rewardsN': '{n} reward days',
    'sk.cycleN': '{n}-day cycle',
    'sk.hiro.notice': 'This is built into this server and does not need Hiro. The Hiro Streaks page stays separate and untouched.',
    'f.sk.event': 'Counting event',
    'f.sk.event.hint': 'For example "Sign-in". Counted at most once per period.',
    'f.sk.periodDays': 'Period length (days)',
    'f.sk.periodDays.hint': 'Usually 1 day.',
    'f.sk.cycle': 'Cycle length',
    'f.sk.cycle.hint': 'After this many days the rewards repeat from day one. Empty means no repeat.',
    'f.sk.grace': 'Grace days',
    'f.sk.grace.hint': 'How many missed days are forgiven before the count is lost.',
    'f.sk.miss': 'If more days are missed',
    'sk.h.rewards': 'Reward for each day',
    'sk.add': 'Add day',
    'sk.empty': 'No rewards yet.',
    'f.sk.day': 'Day',
    'filter.ev.from': 'Source',
    'ev.from.client': 'Game',
    'ev.from.server': 'Server',
    'ev.from.admin': 'Console',
    'col.ev.value': 'Value',
    'col.ev.from': 'Source',
    'col.ev.metadata': 'Details',
    'f.es.timezone': 'Time zone',
    'f.es.timezone.hint': 'Days and weeks start by the clock of this zone.',
    'f.es.weekStart': 'Week starts on',
    'f.es.weekStart.hint': 'Used by weekly prizes.',
    'es.now': 'Now in this zone: {time}',
    'f.es.retention': 'Keep the log (days)',
    'f.es.retention.hint': 'Older events are removed. Player totals and prize progress are kept.',
    'f.es.rate': 'Game events per player per minute',
    'f.es.rate.hint': 'All events sent by the game together.',
    'claim.status.granted': 'Given',
    'claim.status.failed': 'Failed',
    'claim.status.waiting': 'Waiting for player',
    'claim.status.ready': 'Queued',
    'f.ev.value': 'Value',
    'f.ev.value.hint': 'Use 1 for a simple count.',
    'pp.record': 'Record event',
    'pp.record.text': 'The event is recorded for this player, and prize, meta and streak rules run as for a real one.',
    'pp.record.ok': 'Record',
    'pp.title': 'Player progress',
    'pp.recorded': 'Event recorded.',
    'pp.log': 'Event log',
    'pp.level': 'Level {level}',
    'pp.meta': '{n} {name}',
    'pp.toNext': '{n} to level {level}',
    'pp.broken': 'Broken',
    'pp.today': 'Counted today',
    'pp.streak': 'Day {n} · best {best}',
    'pp.reset': 'Reset',
    'pp.reset.title': 'Reset this streak?',
    'pp.reset.text': '"{name}" starts again from day one for this player. Rewards already given stay.',
    'pp.reset.done': 'Reset done.',
    'pp.times': '{n} times',
    'pp.noEvents': 'No events yet.',
    'pp.levels': 'Level',
    'pp.streaks': 'Daily streaks',
    'pp.events': 'Events',
    'pp.claims': 'Recent rewards',
    'pp.readyN': '{n} waiting',
    'claim.kind.event': 'Event',
    'claim.kind.level': 'Level',
    'claim.kind.streak': 'Streak',
    'pp.noClaims': 'No rewards yet.',
    'ov.events': 'Events and rewards',
    'ov.eventsToday': 'Events in the last 24 hours',
    'ov.claimsReady': 'Rewards waiting',
    'ov.claimsFailed': 'Failed rewards',
    'nav.tasks': "Tasks",
    'nav.task-groups': "Task groups",
    'desc.tasks': "Things a player does in the game for a prize, like \"play three rounds today\". Progress comes from events.",
    'desc.task-groups': "A set of tasks handed to the player one at a time or a few at once, in order or at random.",
    'new.tasks': "New task",
    'edit.tasks': "Edit task",
    'new.task-groups': "New task group",
    'edit.task-groups': "Edit task group",
    'empty.hint.tasks': "For example: \"Play three rounds today\" for 20 coins.",
    'empty.hint.task-groups': "For example a chain of three tasks handed out in order.",
    'col.tk.when': "Condition",
    'col.tk.place': "Place",
    'col.tk.done': "Done",
    'col.tg.rule': "Hand-out",
    'col.tg.tasks': "Tasks",
    'filter.tk.level': "Kind",
    'filter.tk.group': "Group",
    'filter.tg.handout': "Hand-out",
    'tk.level.top': "Main task",
    'tk.level.sub': "Sub-task",
    'tk.when.count': "\"{event}\" {n} times",
    'tk.when.sum_value': "\"{event}\" adds up to {n}",
    'tk.when.high_value': "one \"{event}\" of {n} or more",
    'tk.when.children': "when every sub-task is done",
    'tk.filter.min': "only values from {n}",
    'tk.filter.max': "only values up to {n}",
    'tk.filter.range': "only values {min} to {max}",
    'tk.repeat.once': "once",
    'tk.repeat.daily': "every day",
    'tk.repeat.weekly': "every week",
    'tk.repeat.interval': "once every {time}",
    'tk.repeatopt.once': "Once",
    'tk.repeatopt.daily': "Again every day",
    'tk.repeatopt.weekly': "Again every week",
    'tk.repeatopt.interval': "Again after a wait",
    'tk.max': "at most {n} times",
    'tk.minutes': "{n} min",
    'tk.hours': "{n} h",
    'tk.days': "{n} days",
    'tk.noPrize': "No prize",
    'tk.sub': "Sub-task",
    'tk.inGroup': "Group",
    'tk.childrenN': "{n} sub-tasks",
    'tk.depsN': "after {n} other tasks",
    'tk.standalone': "On its own",
    'tk.doneN': "{n} times · {p} players",
    'f.tk.key.hint': "The game finds the task by this key; it cannot change later.",
    'f.tk.name': "Task name",
    'f.tk.name.hint': "Players see this name in the game.",
    'f.tk.description': "Description for players",
    'tk.h.when': "When is it done",
    'f.tk.condition': "Condition type",
    'f.tk.condition.hint': "Either an event moves it forward, or it is done when its sub-tasks are done.",
    'tk.condition.event': "By an event",
    'tk.condition.children': "By sub-tasks",
    'f.tk.metric': "Count",
    'f.tk.metric.hint': "How many events, the sum of their values or the highest value.",
    'tk.metric.count': "Number of times",
    'tk.metric.sum_value': "Sum of values",
    'tk.metric.high_value': "Highest value",
    'f.tk.target': "Target",
    'f.tk.valueMin': "Lowest value that counts",
    'f.tk.valueMax': "Highest value that counts",
    'f.tk.valueFilter.hint': "An event with a value outside this range does not count.",
    'f.tk.parent': "Main task",
    'f.tk.parent.hint': "Pick it if this task is part of a bigger one. A sub-task uses the repeat, time and players of its main task.",
    'tk.h.repeat': "Repeat",
    'f.tk.repeat': "Repeat",
    'f.tk.repeatMinutes': "Wait (minutes)",
    'f.tk.repeatMinutes.hint': "After it is done, it waits this long before it starts again.",
    'f.tk.maxCompletions': "Most times",
    'f.tk.maxCompletions.hint': "Empty means no limit.",
    'f.tk.group': "Group",
    'f.tk.group.hint': "A group task shows only when the group hands it to the player.",
    'f.tk.dependsOn': "After these tasks",
    'f.tk.dependsOn.hint': "The task stays locked until these are done.",
    'f.tk.prize.hint': "A task can have no prize, for example a sub-task.",
    'f.tk.notify': "Tell the player in the game",
    'f.tk.notify.hint': "Sends an in-app message when the task is done.",
    'f.tk.levelTable': "Level table",
    'f.tk.levelTable.hint': "To limit the task by player level.",
    'f.tk.minLevel': "From level",
    'f.tk.maxLevel': "Up to level",
    'f.tg.name': "Group name",
    'tg.h.handout': "Handing out tasks",
    'f.tg.activeCount': "Tasks at once",
    'f.tg.activeCount.hint': "How many tasks of this group a player has at the same time.",
    'f.tg.orderMode': "Order",
    'tg.order.ordered': "in order",
    'tg.order.random': "at random",
    'f.tg.handout': "When the next task comes",
    'f.tg.handout.hint': "After the prize of the last one is collected.",
    'tg.handout.auto': "next task automatically",
    'tg.handout.request': "next task when the player asks",
    'tg.handout.timer.after': "next task {time} later",
    'tg.handoutopt.auto': "Automatically",
    'tg.handoutopt.request': "When the player asks",
    'tg.handoutopt.timer': "After a wait",
    'f.tg.handoutMinutes': "Wait (minutes)",
    'f.tg.repeatable': "Repeat the group",
    'f.tg.repeatable.hint': "When every task is done, the group starts again.",
    'tg.h.tasks': "Tasks in the group",
    'f.tg.tasks': "Tasks",
    'f.tg.tasks.hint': "Handed out in the order you add them. One-time tasks only.",
    'tg.active': "{n} at once",
    'tg.repeatable': "repeats",
    'tg.tasksN': "{n} tasks",
    'tk.card': "Player tasks",
    'tk.all': "All tasks",
    'tk.none': "No tasks for this player.",
    'tk.hiddenN': "{n} more tasks are not shown to this player (off, ended or for other players).",
    'tk.state.active': "In progress",
    'tk.state.completed': "Done",
    'tk.state.claimed': "Prize collected",
    'tk.state.locked': "Locked",
    'tk.state.cooldown': "Waiting for next round",
    'tk.state.done': "Limit reached",
    'tk.state.hidden': "Hidden",
    'tk.state.finished': "Done",
    'tk.reason.inactive': "Off",
    'tk.reason.ended': "Ended",
    'tk.reason.notStarted': "Not started yet",
    'tk.reason.notTargeted': "Not for this player",
    'tk.reason.notHanded': "Not handed out yet",
    'tk.reason.parent': "Main task hidden",
    'tk.reason.groupClosed': "Group closed",
    'tk.reason.groupInactive': "Group off",
    'tk.progress': "{n} of {total}",
    'tk.progress.children': "{n} of {total} sub-tasks",
    'tk.completionsN': "done {n} times",
    'tk.reset': "Start over",
    'tk.reset.title': "Clear this task’s progress?",
    'tk.reset.text': "Progress of \"{name}\" is cleared for this player. Collected prizes stay.",
    'tk.reset.done': "Progress cleared.",
    'tk.complete': "Mark done",
    'tk.complete.title': "Mark this task done?",
    'tk.complete.text': "\"{name}\" counts as done for this player and its prize gets ready.",
    'tk.complete.done': "Task marked done.",
    'tk.claim': "Give the prize",
    'tk.claim.done': "The player got the prize.",
    'tg.round': "Round {n}",
    'tg.progress': "{n} of {total} done",
    'tg.nextAt': "Next: {date}",
    'tg.next': "Hand out next task",
    'tg.next.done': "Next task handed out.",
    'tg.empty': "No task from this group right now.",
    'claim.kind.task': "Task",
    'kind.task': "Task",
    'code.INVALID_METRIC': "Unknown way of counting.",
    'code.INVALID_CONDITION': "Unknown condition type.",
    'code.INVALID_REPEAT_MINUTES': "The wait must be between 1 minute and one year.",
    'code.INVALID_MAX_COMPLETIONS': "At least 1, or empty for no limit.",
    'code.HAS_SUBTASKS': "This task has sub-tasks; delete or move them first.",
    'code.SELF_PARENT': "A task cannot be its own sub-task.",
    'code.NESTED_SUBTASK': "Sub-tasks go one level deep.",
    'code.PARENT_NOT_GROUPING': "The main task must be done \"by sub-tasks\".",
    'code.SUBTASK_NEEDS_EVENT': "A sub-task must move with an event.",
    'code.SUBTASK_IN_GROUP': "Sub-tasks follow their main task; put the main task in the group.",
    'code.GROUP_TASK_REPEAT': "A group task runs once per round of the group; set repeat to \"Once\".",
    'code.SELF_DEPENDENCY': "A task cannot wait for itself.",
    'code.TOO_MANY_DEPENDENCIES': "At most 10 tasks.",
    'code.DEPENDENCY_CYCLE': "These tasks would wait for each other forever.",
    'code.LEVEL_TABLE_REQUIRED': "Pick the level table.",
    'code.INVALID_LEVEL_RANGE': "\"Up to level\" must not be below \"From level\".",
    'code.LEVEL_TABLE_NOT_FOUND': "Level table not found.",
    'code.TASK_NOT_FOUND': "Task not found.",
    'code.TASK_GROUP_NOT_FOUND': "Task group not found.",
    'code.INVALID_ACTIVE_COUNT': "Between 1 and 20.",
    'code.INVALID_ORDER_MODE': "Unknown order.",
    'code.INVALID_HANDOUT': "Unknown hand-out.",
    'code.INVALID_HANDOUT_MINUTES': "The wait must be between 1 minute and one year.",
    'code.TASK_NOT_AVAILABLE': "This task is not open for the player right now.",
    'code.TASK_ALREADY_COMPLETED': "This task is already done.",
    'code.TASK_ALREADY_CLAIMED': "The prize of this task was already given.",
    'code.TASK_NOT_STARTED': "The player has not started this task.",
    'code.TASK_NOT_COMPLETED': "This task is not done yet.",
    'code.TASK_HAS_NO_PRIZE': "This task has no prize.",
    'code.CURRENT_TASK_MUST_BE_DONE': "The current tasks must be done first.",
    'code.DURATION_NOT_ELAPSED': "It is not time for the next task yet.",
    'code.NO_TASK_AVAILABLE': "No task to hand out right now.",
    'code.TASK_GROUP_FINISHED': "Every task of this group is done.",
    'code.HANDOUT_AUTOMATIC': "Tasks of this group are handed out automatically.",
    'section.sup': "Support",
    'section.sup.short': "Support",
    'nav.tickets': "Tickets",
    'nav.live-chat': "Live chat",
    'nav.activity': "Player activity",
    'nav.referrals': "Referral codes",
    'nav.remote-config': "Game config",
    'desc.tickets': "Player support requests; tickets waiting for our answer come first.",
    'desc.live-chat': "The waiting queue and your chats. Take a chat to answer the player.",
    'desc.activity': "What each player did: sign-ins, purchases, prizes, tickets, restrictions and more.",
    'desc.referrals': "Who invited whom, reward status and suspicious cases.",
    'desc.remote-config': "Values the game reads at run time, set apart per platform, app version, player tag or a share of players.",
    'desc.remote-config.edit': "Set the default value and rules. The first rule that fits the player wins.",
    'desc.support-departments': "Departments a player picks when opening a ticket.",
    'desc.support-canned': "Ready texts an agent drops into a reply with one click.",
    'desc.support-settings': "Ticket limits, live chat and how long player activity is kept.",
    'desc.referral-rules': "Rewards for both sides, when they are paid and the rules against abuse.",
    'tab.support-departments': "Support departments",
    'tab.support-canned': "Canned replies",
    'tab.support-settings': "Support settings",
    'tab.referral-rules': "Referral rules",
    'new.support-departments': "New department",
    'edit.support-departments': "Edit department",
    'new.support-canned': "New canned reply",
    'edit.support-canned': "Edit canned reply",
    'empty.hint.support-departments': "Create a department so players can open tickets.",
    'empty.hint.support-canned': "Create a canned reply to answer faster.",
    'empty.hint.tickets': "No tickets with this status.",
    'empty.hint.activity': "No activity recorded yet.",
    'empty.hint.referrals': "Nobody has entered a referral code yet.",
    'empty.hint.remote-config': "Create a key for the game to read.",
    'search.tickets': "Search number, subject or player",
    'search.activity': "Search number, IP or details",
    'search.referrals': "Search player name or code",
    'search.remote-config': "Search key or description",
    'tk.waitingN': "{n} tickets waiting",
    'lc.waitingN': "{n} chats in the queue",
    'tk.status.open': "New",
    'tk.status.awaiting_staff': "Waiting for support",
    'tk.status.awaiting_player': "Waiting for player",
    'tk.status.resolved': "Solved",
    'tk.status.closed': "Closed",
    'tk.priority': "Priority",
    'tk.priority.low': "Low",
    'tk.priority.normal': "Normal",
    'tk.priority.high': "High",
    'tk.priority.urgent': "Urgent",
    'tk.filter.waiting': "Waiting for us",
    'tk.filter.active': "All open",
    'tk.mine': "Mine",
    'tk.unassigned': "Unassigned",
    'tk.overdue': "Overdue",
    'tk.overdueOnly': "Overdue only",
    'tk.by.player': "Last message from the player",
    'tk.by.staff': "Last message from support",
    'tk.stat.waiting': "Waiting for answer",
    'tk.stat.overdue': "Overdue",
    'tk.stat.unassigned': "Unassigned",
    'tk.stat.mine': "Mine",
    'tk.stat.firstResponse': "Average first answer",
    'tk.stat.30days': "In the last 30 days",
    'tk.stat.rating': "Average rating",
    'tk.stat.rated': "{n} ratings",
    'tk.new': "New ticket",
    'tk.new.text': "Open a ticket for a player who reached you another way. Your message is the first message of the ticket.",
    'tk.new.ok': "Create and send",
    'tk.new.hint': "The player sees this message and gets a notification.",
    'tk.pickPlayer': "Find the player",
    'tk.subject': "Subject",
    'tk.message': "Message",
    'tk.detail': "Ticket",
    'tk.player': "Player",
    'tk.staff': "Support",
    'tk.system': "System",
    'tk.note': "Internal note",
    'tk.note.by': "Note by {name}",
    'tk.reply': "Reply",
    'tk.reply.ph': "Write your reply… (Ctrl+Enter sends)",
    'tk.note.ph': "Write a note for your team…",
    'tk.canned': "Canned reply",
    'tk.canned.pick': "Insert a canned reply…",
    'tk.canned.none': "No canned replies",
    'tk.attach': "Attach",
    'tk.uploading': "Uploading…",
    'tk.internal': "Internal note (the player does not see it)",
    'tk.send': "Send",
    'tk.saveNote': "Save note",
    'tk.after': "After the reply",
    'tk.after.awaiting_player': "Wait for player",
    'tk.after.resolved': "Solved",
    'tk.after.closed': "Close",
    'tk.sent': "Reply sent and the player was notified.",
    'tk.noteSaved': "Note saved.",
    'tk.closedNote': "This ticket is closed. The player cannot reply any more; you can add a note or reopen it.",
    'tk.move.resolved': "Solved",
    'tk.move.closed': "Close ticket",
    'tk.move.awaiting_staff': "Reopen",
    'tk.moved': "Ticket status changed.",
    'tk.assigned': "Ticket owner changed.",
    'tk.takeIt': "Take it myself",
    'tk.playerCounts': "{tickets} tickets · {chats} chats",
    'tk.playerTickets': "Tickets of this player",
    'tk.times': "Times",
    'tk.created': "Created",
    'tk.due': "Answer due",
    'tk.firstResponse': "First answer",
    'tk.notYet': "Not yet",
    'tk.resolvedAt': "Solved at",
    'tk.closedAt': "Closed at",
    'tk.rating': "Player rating",
    'tk.ratingN': "{n} of 5",
    'tk.fromChat': "From live chat",
    'tk.openChat': "Open chat",
    'tk.source.chat': "From live chat",
    'tk.source.staff': "Opened by support",
    'tk.ev.status_changed': "{who}: {from} → {to}",
    'tk.ev.status_changed.auto': "Closed automatically after days without an answer",
    'tk.ev.assigned': "{who} gave the ticket to {to}",
    'tk.ev.assigned.auto': "Ticket went to {to} automatically",
    'tk.ev.assigned.none': "{who} removed the owner",
    'tk.ev.department_changed': "{who} changed the department",
    'tk.ev.priority_changed': "{who} set the priority to {to}",
    'tk.ev.rated': "The player rated {n} of 5",
    'tk.ev.created_from_chat': "This ticket was made from a live chat",
    'col.subject': "Subject",
    'col.assignee': "Owner",
    'col.department': "Department",
    'col.lastMessage': "Last message",
    'col.firstResponse': "First answer within",
    'col.agents': "Agents",
    'col.open': "Open",
    'col.what': "What happened",
    'col.where': "Where",
    'col.referral': "Inviter → invited",
    'col.goal': "Reward goal",
    'col.value': "Value",
    'col.rules': "Rules",
    'col.version': "Version",
    'filter.priority': "Priority",
    'filter.assignee': "Owner",
    'filter.department': "Department",
    'filter.overdue': "Due",
    'filter.platform': "Platform",
    'filter.country': "Country code",
    'filter.from': "From date (2026-11-01)",
    'filter.to': "To date",
    'filter.flagged': "Suspicious",
    'filter.referrerId': "Inviter id",
    'filter.type': "Type",
    'dur.minutes': "{n} min",
    'size.b': "{n} B",
    'size.kb': "{n} KB",
    'size.mb': "{n} MB",
    'lc.queue': "Queue",
    'lc.mine': "My chats",
    'lc.all': "All",
    'lc.closed': "Closed",
    'lc.views': "Chat views",
    'lc.available': "Available",
    'lc.nowOnline': "You are online and players can see you.",
    'lc.nowAway': "You are away.",
    'lc.empty.queue': "Nobody is waiting",
    'lc.empty.queue.hint': "New chats show up here.",
    'lc.empty.mine': "You have no open chats",
    'lc.empty.all': "No open chats",
    'lc.empty.closed': "No closed chats",
    'lc.take': "Take",
    'lc.taken': "The chat is yours.",
    'lc.pick': "Pick a chat",
    'lc.pick.hint': "Take one from the queue to start.",
    'lc.waitSec': "{n} s in queue",
    'lc.waitMin': "{n} min in queue",
    'lc.typing': "Typing…",
    'lc.playerTypingNow': "The player is typing…",
    'lc.playerOnline': "Player is online",
    'lc.playerOffline': "Player is offline",
    'lc.online': "Online",
    'lc.offline': "Offline",
    'lc.activeN': "{n} open chats",
    'lc.status.waiting': "Waiting",
    'lc.status.active': "Active",
    'lc.status.closed': "Closed",
    'lc.withAgent': "with {name}",
    'lc.ticketN': "Ticket {n}",
    'lc.reply.ph': "Write a message… (Ctrl+Enter sends)",
    'lc.someoneElse': "{name} is handling this chat.",
    'lc.closedNote': "This chat is closed.",
    'lc.closedRated': "Chat closed. The player rated {n} of 5.",
    'lc.transfer': "Transfer",
    'lc.transfer.text': "Give the chat to a teammate or put it back in the queue.",
    'lc.transfer.ok': "Transfer",
    'lc.transferTo': "To whom",
    'lc.backToQueue': "Back to the queue",
    'lc.transferred': "Chat transferred.",
    'lc.close': "Close chat",
    'lc.close.title': "Close this chat?",
    'lc.close.text': "The player sees that the chat ended and can rate it.",
    'lc.toTicket': "Turn the chat into a ticket",
    'lc.toTicket.hint': "The conversation is kept in a new ticket for follow-up.",
    'lc.closedDone': "Chat closed.",
    'lc.closedTicket': "Chat closed and ticket {n} was created.",
    'lc.ev.accepted': "{agent} took the chat",
    'lc.ev.transferred': "Chat moved from {from} to {to}",
    'lc.ev.closed': "Chat closed ({by})",
    'lc.ev.note': "System message",
    'code.ALREADY_TAKEN': "{name} already took this chat.",
    'code.NOT_YOUR_CHAT': "A teammate is handling this chat.",
    'code.CHAT_CLOSED': "This chat is closed.",
    'code.CHAT_EMPTY': "The chat has no messages to turn into a ticket.",
    'code.CHAT_NOT_FOUND': "Chat not found.",
    'code.TICKET_NOT_FOUND': "Ticket not found.",
    'code.TICKET_CLOSED': "This ticket is closed.",
    'code.INVALID_TRANSITION': "This status change is not possible.",
    'code.INVALID_PRIORITY': "Priority is not valid.",
    'code.INVALID_SUBJECT': "The subject is 1 to 150 letters.",
    'code.BODY_REQUIRED': "Write a message.",
    'code.BODY_TOO_LONG': "The message is too long.",
    'code.ATTACHMENT_NOT_FOUND': "Attachment not found.",
    'code.TOO_MANY_ATTACHMENTS': "Too many attachments.",
    'code.ATTACHMENT_TOO_LARGE': "The file is too large.",
    'code.ATTACHMENT_TYPE': "Only png, jpg, webp images and pdf files are allowed.",
    'code.ATTACHMENTS_UNAVAILABLE': "Attachments are not available right now.",
    'code.DEPARTMENT_NOT_FOUND': "Department not found.",
    'code.INVALID_AGENT': "This teammate was not found.",
    'code.TOO_MANY_AGENTS': "At most 50 agents per department.",
    'code.INVALID_JSON': "The JSON is not valid.",
    'code.INVALID_PERCENT': "Percent is 0 to 100.",
    'code.VALUE_REQUIRED': "A value is required.",
    'code.VALUE_TOO_LONG': "The value is too long.",
    'code.TOO_MANY_RULES': "At most 20 rules.",
    'code.INVALID_RULE_NAME': "Rule names are short lower-case keys and not \"default\".",
    'code.DUPLICATE_RULE_NAME': "Rule names must be different.",
    'code.MIN_ABOVE_MAX': "The lowest version is above the highest.",
    'code.TYPE_LOCKED': "The key type cannot change; create a new key.",
    'code.VERSION_NOT_FOUND': "That version is not in the history.",
    'code.DESCRIPTION_TOO_LONG': "The description is at most 500 letters.",
    'code.REFERRAL_NOT_FOUND': "Referral not found.",
    'code.REFERRAL_DECIDED': "This referral was already decided.",
    'code.NOT_HELD': "Only held referrals need approval.",
    'dp.everyone': "Everyone",
    'cn.allDepts': "All departments",
    'f.dp.key.hint': "A fixed id, like payments.",
    'f.dp.name': "Department name",
    'f.dp.name.hint': "Players see this name in the ticket form.",
    'f.dp.description': "Description",
    'f.dp.description.hint': "One sentence about what this department handles.",
    'f.dp.hours': "First answer within (hours)",
    'f.dp.hours.hint': "A ticket with no answer after this time is overdue.",
    'f.dp.agents': "Agents of this department",
    'f.dp.agents.hint': "New tickets go to the least busy agent in this list. Empty means no automatic sharing.",
    'f.dp.active.hint': "An inactive department is hidden from players.",
    'f.cn.key.hint': "A fixed id, like greeting.",
    'f.cn.title': "Title",
    'f.cn.title.hint': "Only agents see this title.",
    'f.cn.body': "Reply text",
    'f.cn.body.hint': "Inserted in the ticket language; you can edit it before sending.",
    'f.cn.department': "Only for department",
    'f.cn.department.hint': "Empty means every department.",
    'f.cn.active.hint': "An inactive reply is not offered.",
    'ss.tickets': "Tickets",
    'ss.chat': "Live chat",
    'ss.activity': "Activity log",
    'f.ss.maxOpen': "Most open tickets per player",
    'f.ss.maxOpen.hint': "After this the player has to close one first.",
    'f.ss.perHour': "New tickets per hour",
    'f.ss.perHour.hint': "Stops a flood of tickets.",
    'f.ss.messages': "Messages per hour",
    'f.ss.messages.hint': "For player replies across all tickets.",
    'f.ss.autoClose': "Close automatically after (days)",
    'f.ss.autoClose.hint': "Tickets waiting for the player or solved are closed after this time.",
    'f.ss.attachments': "Most attachments per message",
    'f.ss.attachments.hint': "0 turns attachments off.",
    'f.ss.attachmentKb': "Size of each attachment (KB)",
    'f.ss.attachmentKb.hint': "png, jpg, webp and pdf only.",
    'f.ss.chatEnabled': "Live chat is on",
    'f.ss.chatEnabled.hint': "When off, players can only open tickets.",
    'f.ss.greeting': "Greeting",
    'f.ss.greeting.hint': "Written to the player at the start of every chat.",
    'f.ss.offline': "Message when nobody is online",
    'f.ss.offline.hint': "Shown to the player when no agent is online.",
    'f.ss.idle': "Close idle chats after (minutes)",
    'f.ss.idle.hint': "When there are no messages and the player has left.",
    'f.ss.chatRate': "Player messages per minute",
    'f.ss.chatRate.hint': "Stops message floods.",
    'f.ss.retention': "Keep player activity (days)",
    'f.ss.retention.hint': "Older activity is removed every few minutes.",
    'act.signin': "Sign-in",
    'act.purchase': "Purchase",
    'act.prize': "Prize",
    'act.event': "Event",
    'act.restriction': "Restriction",
    'act.report': "Report",
    'act.ticket': "Ticket",
    'act.chat': "Support chat",
    'act.referral': "Referral",
    'act.config': "Config",
    'act.method.device': "device",
    'act.method.custom': "custom id",
    'act.method.email': "email",
    'act.method.phone': "mobile number",
    'act.method.google': "Google",
    'act.method.apple': "Apple",
    'act.method.facebook': "Facebook",
    'act.method.steam': "Steam",
    'act.method.gamecenter': "Game Center",
    'act.signin.device': "Signed in with {method}",
    'act.signin.custom': "Signed in with {method}",
    'act.signin.email': "Signed in with {method}",
    'act.signin.phone': "Signed in with {method}",
    'act.signin.google': "Signed in with {method}",
    'act.signin.apple': "Signed in with {method}",
    'act.signin.facebook': "Signed in with {method}",
    'act.signin.steam': "Signed in with {method}",
    'act.signin.gamecenter': "Signed in with {method}",
    'act.purchase.completed': "Bought {productKey} with {method}",
    'act.prize.granted': "Got the prize {prize}",
    'act.prize.gift_redeemed': "Redeemed gift code {code}",
    'act.event.level_up': "Reached level {to}",
    'act.restriction.suspend': "Suspended until {endsAt}",
    'act.restriction.mute': "Muted until {endsAt}",
    'act.restriction.warn': "Got a warning",
    'act.restriction.lifted': "{kind} lifted",
    'act.report.filed': "Reported a message",
    'act.report.received': "A message of theirs was reported",
    'act.ticket.created': "Opened ticket {number}: {subject}",
    'act.ticket.created_by_staff': "Support opened ticket {number} for them",
    'act.ticket.staff_reply': "Support answered ticket {number}",
    'act.ticket.resolved': "Ticket {number} solved",
    'act.ticket.closed': "Ticket {number} closed",
    'act.ticket.resolved_by_player': "Marked ticket {number} as solved",
    'act.ticket.rated': "Rated ticket {number} with {rating}",
    'act.chat.opened': "Opened a live chat",
    'act.chat.taken': "{agent} took the chat",
    'act.chat.transferred': "Chat moved from {from} to {to}",
    'act.chat.closed': "Live chat closed",
    'act.referral.entered': "Entered the referral code of {referrer}",
    'act.referral.joined': "{referee} joined with their code",
    'act.referral.completed': "Referral goal reached",
    'act.referral.approved': "Referral approved",
    'act.referral.rejected': "Referral rejected",
    'act.config.changed': "Game config changed for them: {variants}",
    'act.actor.player': "The player",
    'act.actor.system': "System",
    'act.actor.console': "Console ({name})",
    'act.open': "Open",
    'act.all': "All activity",
    'act.card': "Recent activity",
    'act.empty': "No activity.",
    'act.playerLink': "Activity of this player",
    'act.retention': "Activity is kept for {n} days.",
    'pf.android': "Android",
    'pf.ios': "iOS",
    'pf.web': "Web",
    'pf.windows': "Windows",
    'pf.macos': "macOS",
    'pf.linux': "Linux",
    'rf.status.pending': "Waiting for goal",
    'rf.status.held': "Held",
    'rf.status.completed': "Completed",
    'rf.status.rejected': "Rejected",
    'rf.flag.burst': "Many invites in one day",
    'rf.flag.same_ip': "Same IP as the inviter",
    'rf.flag.ip_burst': "Several invites from one IP",
    'rf.flag.shared_device': "Repeated device",
    'rf.flag.shared_phone': "Repeated number",
    'rf.flaggedOnly': "Suspicious only",
    'rf.noGoal': "No goal",
    'rf.progress': "{n} of {total}",
    'rf.tree': "Referral tree",
    'rf.tree.title': "Referral tree",
    'rf.tree.up': "Invited by",
    'rf.tree.down': "Invited {n} players",
    'rf.tree.more': "{n} invites",
    'rf.tree.none': "Has not invited anyone yet.",
    'rf.approve': "Approve",
    'rf.reject': "Reject",
    'rf.approve.title': "Approve this referral?",
    'rf.approve.text': "{from} invited {to}. Once approved, the rewards are paid when the goal is met.",
    'rf.approve.done': "Referral approved.",
    'rf.reject.title': "Reject this referral?",
    'rf.reject.text': "{from} invited {to}; no reward will be paid.",
    'rf.reject.done': "Referral rejected.",
    'rf.stat.total': "All referrals",
    'rf.stat.today': "{n} in the last 24 hours",
    'rf.stat.flagged': "{n} suspicious",
    'rf.top': "Top inviters",
    'rf.top.line': "{n} invites · {done} completed",
    'rf.top.flagged': "{n} suspicious",
    'claim.kind.referral': "Referral",
    'rr.rewards': "Rewards",
    'rr.rules': "Rules",
    'rr.measure.count': "Times",
    'rr.measure.sum': "Total value",
    'rr.measure.max': "Highest value",
    'f.rr.enabled': "Referral codes are on",
    'f.rr.enabled.hint': "When off, players cannot enter a code.",
    'f.rr.referrerPrize': "Inviter reward",
    'f.rr.referrerPrize.hint': "When their friend meets the goal.",
    'f.rr.refereePrize': "Invited player reward",
    'f.rr.refereePrize.hint': "For the player who entered the code.",
    'f.rr.goal': "Goal event",
    'f.rr.goal.hint': "Rewards are paid after this event. Empty pays them when the code is entered.",
    'f.rr.measure': "Measure",
    'f.rr.goalValue': "Needed value",
    'f.rr.goalValue.hint': "For example 3 means doing the event three times.",
    'f.rr.days': "New accounts only (days)",
    'f.rr.days.hint': "A player can enter a code this many days after creating the account. 0 means no limit.",
    'f.rr.max': "Most invites per player",
    'f.rr.max.hint': "After this their code is not accepted.",
    'f.rr.daily': "Normal invites per day",
    'f.rr.daily.hint': "Above this an invite is marked suspicious.",
    'f.rr.device': "One referral per device",
    'f.rr.device.hint': "A device that already came with a code is refused.",
    'f.rr.phone': "One referral per number",
    'f.rr.phone.hint': "A number that already came with a code is refused.",
    'f.rr.hold': "Hold suspicious referrals",
    'f.rr.hold.hint': "Rewards of suspicious referrals wait for your approval.",
    'cf.new': "New key",
    'cf.edit': "Edit key",
    'cf.key': "Key",
    'cf.key.hint': "The name the game reads, like stargame.banner.",
    'cf.type': "Type",
    'cf.type.hint': "Cannot change after creating.",
    'cf.type.locked': "The key type is fixed.",
    'cf.type.string': "Plain text",
    'cf.type.number': "Number",
    'cf.type.bool': "On/off",
    'cf.type.json': "JSON",
    'cf.type.text': "Translated text",
    'cf.description': "Description",
    'cf.description.hint': "For your team; the game does not see it.",
    'cf.default': "Default value",
    'cf.default.hint': "When no rule fits the player.",
    'cf.value': "Value",
    'cf.valueOn': "On",
    'cf.active': "The key is on",
    'cf.on': "On",
    'cf.off': "Off",
    'cf.everyone': "Everyone",
    'cf.rulesN': "{n} rules",
    'cf.v': "v{n}",
    'cf.rules': "Rules",
    'cf.rules.hint': "Checked in order; the first rule whose conditions all fit gives its value.",
    'cf.rule': "Rule",
    'cf.rule.add': "New rule",
    'cf.rule.none': "No rules; everyone gets the default.",
    'cf.rule.name': "Rule name",
    'cf.rule.name.hint': "Short and in English, like vip or android-new.",
    'cf.rule.percent': "Share of players",
    'cf.rule.percent.hint': "Empty means everyone. Each player always stays in the same group.",
    'cf.rule.platforms': "Platform",
    'cf.rule.platforms.hint': "None means every platform.",
    'cf.rule.min': "From version",
    'cf.rule.max': "To version",
    'cf.rule.tags': "Player tag",
    'cf.rule.countries': "Countries",
    'cf.rule.countries.hint': "Two-letter codes, separated by commas.",
    'cf.rule.value': "Value of this rule",
    'cf.rule.value.hint': "Players who meet the conditions get this.",
    'cf.rule.up': "Move up",
    'cf.rule.down': "Move down",
    'cf.delete.title': "Delete this key?",
    'cf.delete.text': "\"{key}\" will no longer be sent to the game. You can bring it back from the history.",
    'cf.history': "History",
    'cf.recent': "Recent changes",
    'cf.action.create': "Created",
    'cf.action.update': "Edited",
    'cf.action.delete': "Deleted",
    'cf.action.rollback': "Rolled back",
    'cf.rollback': "Roll back to this",
    'cf.rollback.title': "Roll back to v{n}?",
    'cf.rollback.text': "The value and rules of that version are saved as a new version.",
    'cf.rolledBack': "Rolled back to v{n}.",
    'cf.preview': "Try for a player",
    'cf.preview.hint': "See what values the game gets for this player and device.",
    'cf.preview.player': "Player id",
    'cf.preview.player.ph': "Empty for an unknown player",
    'cf.preview.version': "App version",
    'cf.preview.country': "Country code",
    'cf.preview.run': "Try it",
    'cf.variant.default': "Default",
    'cf.bucket': "bucket {n}%",
    'nav.items': "Items",
    'desc.items': "Things a player owns in the game, like skins, frames and avatars. Players buy them, unlock them by level or get them as prizes.",
    'new.items': "New item",
    'edit.items': "Edit item",
    'search.items': "Search item key and name",
    'kind.item': "Item",
    'kind.item_type': "Item type",
    'col.it.type': "Type",
    'col.it.rarity': "Rarity",
    'col.it.get': "How to get",
    'col.it.owners': "Owners",
    'filter.it.type': "Type",
    'filter.it.rarity': "Rarity",
    'rarity.common': "Common",
    'rarity.uncommon': "Uncommon",
    'rarity.rare': "Rare",
    'rarity.epic': "Epic",
    'rarity.legendary': "Legendary",
    'it.get.starter': "Everyone has it",
    'it.get.level': "From level {n}",
    'it.get.prize': "Prize only",
    'it.get.requires': "Needs: {key}",
    'it.upgradesN': "{n} upgrade levels",
    'it.ownersN': "{n} players",
    'it.equippedN': "{n} equipped",
    'f.it.type': "Item type",
    'f.it.type.hint': "A player equips one item of each type, for example one skin.",
    'f.it.rarity': "Rarity",
    'h.it.get': "How a player gets it",
    'f.it.starter': "Everyone has it from the start",
    'f.it.starter.hint': "Given to every player automatically and cannot be taken back.",
    'f.it.price.currency': "Price currency",
    'f.it.price.amount': "Price",
    'f.it.price.hint': "Leave empty when the item comes only from prizes or levels.",
    'f.it.levelTable': "Level table",
    'f.it.levelTable.hint': "Empty means the game's main level table.",
    'f.it.unlockLevel': "Unlocks at level",
    'f.it.unlockLevel.hint': "Without a price the player unlocks it for free at this level. With a price it can be bought from this level.",
    'f.it.parent': "Base item",
    'f.it.requires': "Requires",
    'f.it.requires.hint': "The player must own this item first.",
    'h.it.upgrades': "Upgrades",
    'it.up.add': "New upgrade level",
    'it.up.empty': "This item has no upgrades.",
    'f.it.up.currency': "Currency",
    'f.it.up.amount': "Price of this level",
    'h.it.services': "Services",
    'it.sv.add': "New service",
    'it.sv.empty': "This item has no services.",
    'f.it.sv.key': "Key",
    'f.it.sv.currency': "Currency",
    'f.it.sv.price': "Price per unit",
    'f.it.sv.max': "Most units",
    'f.it.sv.name': "Service name",
    'f.it.image': "Large image",
    'item.item': "Item",
    'item.item.once': "Given to the player once",
    'h.pr.rules': "Purchase rules",
    'f.maxLevel': "Maximum level",
    'f.maxLevel.hint': "Players above this level do not see the product.",
    'f.pr.levelTable': "Level table",
    'f.pr.levelTable.hint': "Empty means the game's main level table.",
    'f.isOneTime': "Can be bought only once",
    'f.isOneTime.hint': "After a purchase it leaves that player's shop.",
    'f.checkContents': "Don't sell when the player owns it all",
    'f.checkContents.hint': "When the player already owns every item of the product's prize, the purchase is refused.",
    'f.purchaseMessage': "Message after purchase",
    'pr.oneTime': "Once",
    'pr.levels': "Level {min} to {max}",
    'pr.levelFrom': "From level {n}",
    'pr.levelTo': "Up to level {n}",
    'inv.card': "Player inventory",
    'inv.counts': "{n} of {total} items",
    'inv.equipped': "Equipped",
    'inv.slot': "{slot}: {name}",
    'inv.source.starter': "Starter",
    'inv.source.buy': "Bought",
    'inv.source.unlock': "Level",
    'inv.source.prize': "Prize",
    'inv.source.admin': "Console",
    'inv.upgrade': "Upgrade {n} of {max}",
    'inv.inactive': "Inactive",
    'inv.grant': "Give item",
    'inv.grant.pick': "Item to give",
    'inv.grant.done': "Item given to the player.",
    'inv.revoke': "Take back",
    'inv.revoke.title': "Take this item back?",
    'inv.revoke.text': "\"{name}\" leaves this player's inventory. Nothing is refunded.",
    'inv.revoke.done': "Item taken back.",
    'inv.none': "This player has no items yet.",
    'inv.all': "All items",
    'act.item': "Item",
    'act.item.bought': "Bought \"{item}\"",
    'act.item.unlocked': "Unlocked \"{item}\" by level",
    'act.item.upgraded': "Upgraded \"{item}\" to {level}",
    'act.item.granted': "Got \"{item}\" from the console",
    'act.item.revoked': "\"{item}\" was taken back in the console",
    'act.item.starter': "Got the starter items",
    'code.ITEM_NOT_FOUND': "Item not found.",
    'nav.achievements': 'Achievements',
    'desc.achievements': 'Badges players unlock by playing, finishing missions or getting prizes. An achievement can have several tiers, each with its own prize.',
    'new.achievements': 'New achievement',
    'edit.achievements': 'Edit achievement',
    'search.achievements': 'Search achievement key and name',
    'empty.achievements': 'No achievements yet.',
    'kind.achievement': 'Achievement',
    'claim.kind.achievement': 'Achievement',
    'item.achievement': 'Achievement',
    'item.achievement.once': 'Unlocks the achievement for the player',
    'col.ac.when': 'Unlocks when',
    'col.ac.tiers': 'Tiers',
    'col.ac.rarity': 'Rarity and points',
    'col.ac.players': 'Unlocked by',
    'filter.ac.source': 'Unlocked by',
    'filter.ac.hidden': 'Hidden',
    'ac.source.event': 'Event',
    'ac.source.task': 'Mission',
    'ac.source.manual': 'Prize or console',
    'ac.when.count': '“{event}” {n} times',
    'ac.when.sum_value': '“{event}” values add up to {n}',
    'ac.when.high_value': 'one “{event}” of {n} or more',
    'ac.when.task': 'mission “{name}” is done',
    'ac.when.group': 'every mission of “{name}” is done',
    'ac.when.manual': 'only by a prize or the console',
    'ac.filter.min': 'values of {n} or more',
    'ac.filter.max': 'values of {n} or less',
    'ac.filter.range': 'values from {min} to {max}',
    'ac.tiersN': '{n} tiers',
    'ac.tier.n': 'Tier {n}',
    'ac.tier.at': '{name}: {n}',
    'ac.pointsN': '{n} points',
    'ac.unlockedN': '{n} players',
    'ac.hidden': 'Hidden',
    'ac.secret': 'Secret until unlocked',
    'ac.noPrize': 'No prize',
    'ac.h.look': 'Look and value',
    'ac.h.when': 'Unlocks when',
    'ac.h.tiers': 'Tiers and prizes',
    'ac.h.prize': 'Unlock prize',
    'f.ac.name': 'Achievement name',
    'f.hint': 'Hint while locked',
    'f.ac.hint.hint': 'Shown while the achievement is locked. For a hidden one it is all the player sees.',
    'f.ac.rarity': 'Rarity',
    'f.ac.points': 'Points',
    'f.ac.points.hint': 'The player profile shows the total points of unlocked achievements.',
    'f.ac.hidden': 'Hidden until unlocked',
    'f.ac.hidden.hint': 'Players do not see its name or goal until they unlock it.',
    'f.ac.notify': 'In-app message on unlock',
    'f.ac.notify.hint': 'After the unlock and every new tier the player gets a message with “Open achievements”.',
    'f.ac.source': 'Unlocked by',
    'f.ac.source.hint': 'Prizes and the console can give any achievement too.',
    'f.ac.metric': 'Count',
    'f.ac.task': 'Mission',
    'f.ac.task.hint': 'Unlocks when this mission is done.',
    'f.ac.group': 'Or mission group',
    'f.ac.group.hint': 'Unlocks when every mission of a round of the group is done.',
    'f.ac.icon': 'Badge',
    'f.ac.lockedIcon': 'Locked badge',
    'f.ac.tier.name': 'Tier name',
    'f.ac.tier.threshold': 'Target',
    'ac.tier.add': 'New tier',
    'ac.tier.empty': 'Add at least one tier with its target.',
    'ac.tiers.hint': 'Tiers unlock in order. Each higher tier needs a larger target and pays its prize once.',
    'pac.card': 'Player achievements',
    'pac.counts': '{n} of {total} achievements · {p} points',
    'pac.state.locked': 'Locked',
    'pac.state.unlocked': 'Unlocked',
    'pac.state.completed': 'All tiers',
    'pac.state.hidden': 'Hidden',
    'pac.progress': '{n} of {total}',
    'pac.tier': 'Tier {n} of {max}',
    'pac.by.event': 'by playing',
    'pac.by.progress': 'by playing',
    'pac.by.task': 'by a mission',
    'pac.by.prize': 'by a prize',
    'pac.by.admin': 'from the console ({name})',
    'pac.at': '{date}',
    'pac.claim': 'Give the prize',
    'pac.claim.done': 'Tier prize given.',
    'pac.revoke': 'Take back',
    'pac.revoke.title': 'Take the achievement back?',
    'pac.revoke.text': '“{name}” and its progress are removed for this player. Prizes already received stay and are not paid again.',
    'pac.revoke.done': 'Achievement taken back.',
    'pac.give': 'Give achievement',
    'pac.give.pick': 'Achievement to give',
    'pac.give.tier': 'Tier',
    'pac.give.done': 'Achievement given to the player.',
    'pac.none': 'No achievements are defined yet.',
    'pac.all': 'All achievements',
    'pac.log': 'Achievement history',
    'pac.waiting': '{n} prizes waiting',
    'act.achievement': 'Achievement',
    'act.achievement.unlocked': 'Unlocked “{achievement}”',
    'act.achievement.tier': 'Reached tier {tier} of “{achievement}”',
    'act.achievement.revoked': '“{achievement}” was taken back in the console',
    'code.ACHIEVEMENT_NOT_FOUND': 'Achievement not found.',
    'code.ACHIEVEMENT_LOCKED': 'This achievement is not unlocked yet.',
    'code.ALREADY_UNLOCKED': 'The player already has this achievement at this tier.',
    'code.NOT_UNLOCKED': 'The player does not have this achievement.',
    'code.TIER_NOT_REACHED': 'The player has not reached this tier yet.',
    'code.TIER_NOT_FOUND': 'This tier does not exist.',
    'code.TIER_HAS_NO_PRIZE': 'This tier has no prize.',
    'code.TIER_IN_USE': 'Players received this tier’s prize, so it cannot be removed.',
    'code.NOTHING_TO_CLAIM': 'There is no prize left to collect.',
    'code.TIERS_REQUIRED': 'Add at least one tier with its target.',
    'code.TIERS_NEED_EVENT': 'Only an achievement counted from events can have several tiers.',
    'code.TOO_MANY_TIERS': 'At most 10 tiers.',
    'code.TIERS_NOT_CONSECUTIVE': 'Tiers must follow each other.',
    'code.DUPLICATE_TIER': 'Each tier once.',
    'code.INVALID_TIER': 'Tiers go from 1 to 10.',
    'code.INVALID_POINTS': 'Points must be between 0 and 100,000.',
    'code.INVALID_SOURCE': 'Pick how it unlocks.',
    'code.TASK_REQUIRED': 'Pick the mission or the mission group that unlocks it.',
    'code.TASK_AND_GROUP': 'Pick a mission or a mission group, not both.',
    'ac.thresholds.order': 'Each higher tier needs a larger target.',
    'code.ITEM_TYPE_NOT_FOUND': "Item type not found.",
    'code.SELF_REFERENCE': "An item cannot require itself.",
    'code.INVALID_RARITY': "Pick a rarity from the list.",
    'code.STARTER_HAS_LEVEL': "A starter item is open from the start and needs no level.",
    'code.TOO_MANY_UPGRADES': "At most 50 upgrade levels.",
    'code.TOO_MANY_SERVICES': "At most 20 services.",
    'code.DUPLICATE_SERVICE': "Service keys must be unique.",
    'code.INVALID_UNITS': "The number of units is not valid.",
    'code.ALREADY_OWNED': "The player already owns this item.",
    'code.NOT_OWNED': "The player does not own this item.",
    'code.STARTER_ITEM': "Starter items cannot be taken back.",
    'code.LEVEL_TOO_LOW': "The player's level is too low.",
    'code.LEVEL_TOO_HIGH': "The player's level is above this product's limit.",
    'code.PRODUCT_ALREADY_BOUGHT': "This product can be bought once and the player already bought it.",
    'code.PRIZE_ALREADY_OWNED': "The player already owns everything in this prize.",
    'code.ONE_TIME_SUBSCRIPTION': "A subscription cannot be a one-time product."
  }
};

var API_BASE = '/v2/console/economy';
var PUB_BASE = '/v2/console/publication';
var BASE_BASE = '/v2/console/base';
var CONSOLE_BASE = '/v2/console';
var PERSIAN = '۰۱۲۳۴۵۶۷۸۹';
var SVG_NS = 'http://www.w3.org/2000/svg';

function lang() {
  return /^fa/i.test(document.documentElement.getAttribute('lang') || '') ? 'fa' : 'en';
}

function locale() {
  return lang() === 'fa' ? 'fa-IR' : 'en-US';
}

function toLatinDigits(s) {
  return String(s == null ? '' : s)
    .replace(/[۰-۹]/g, function (d) { return String(d.charCodeAt(0) - 0x06F0); })
    .replace(/[٠-٩]/g, function (d) { return String(d.charCodeAt(0) - 0x0660); })
    .replace(/[٬،]/g, ',')
    .replace(/٫/g, '.');
}

function localDigits(s) {
  s = String(s == null ? '' : s);
  if (lang() !== 'fa') return s;
  return s.replace(/\d/g, function (d) { return PERSIAN[+d]; });
}

function fmtNum(n, opts) {
  if (n === null || n === undefined || n === '') return '';
  var v = Number(n);
  if (!isFinite(v)) return String(n);
  return new Intl.NumberFormat(locale(), opts || { maximumFractionDigits: 4 }).format(v);
}

function fmtDate(v, withTime) {
  if (!v) return '';
  var d = new Date(v);
  if (isNaN(d.getTime())) return String(v);
  var o = withTime === false ? { dateStyle: 'medium' } : { dateStyle: 'medium', timeStyle: 'short' };
  return new Intl.DateTimeFormat(locale(), o).format(d);
}

function t(key, vars) {
  var table = MESSAGES[lang()];
  var s = table[key];
  if (s === undefined) s = MESSAGES.en[key];
  if (s === undefined) return key;
  if (vars) {
    s = s.replace(/\{(\w+)\}/g, function (m, k) {
      var v = vars[k];
      if (typeof v === 'number') return fmtNum(v);
      return v === undefined ? m : String(v);
    });
  }
  return s;
}

function has(key) {
  return MESSAGES.en[key] !== undefined;
}

var LANGS = { items: [], active: [], codes: [], def: 'fa', loaded: false, promise: null, seq: 0, at: 0 };
var FALLBACK_LANGS = [
  { code: 'fa', nativeName: 'فارسی', name: { fa: 'فارسی', en: 'Persian' }, direction: 'rtl', isActive: true, isDefault: true, sort: 0 },
  { code: 'en', nativeName: 'English', name: { fa: 'انگلیسی', en: 'English' }, direction: 'ltr', isActive: true, isDefault: false, sort: 1 }
];

function setLanguages(items) {
  LANGS.items = items;
  LANGS.active = items.filter(function (l) { return l.isActive; });
  LANGS.codes = items.map(function (l) { return l.code; });
  var d = items.filter(function (l) { return l.isDefault; })[0] || LANGS.active[0] || items[0];
  LANGS.def = d ? d.code : 'fa';
  if (d) LANGS.active.sort(function (a, b) { return (b.code === d.code) - (a.code === d.code); });
}
setLanguages(FALLBACK_LANGS);

function loadLanguages(force) {
  if (LANGS.promise && !force) return LANGS.promise;
  var my = ++LANGS.seq;
  LANGS.promise = baseApi('GET', '/languages?limit=100').then(function (res) {
    if (my !== LANGS.seq) return LANGS;
    var items = (res && res.items) || [];
    if (items.length) setLanguages(items);
    LANGS.loaded = true;
    LANGS.at = Date.now();
    return LANGS;
  }, function () {
    if (my === LANGS.seq) LANGS.promise = null;
    return LANGS;
  });
  return LANGS.promise;
}

function freshLanguages() {
  return loadLanguages(!LANGS.loaded || Date.now() - LANGS.at > 30000);
}

function langInfo(code) {
  return LANGS.items.filter(function (l) { return l.code === code; })[0] || null;
}

function langDir(code) {
  var l = langInfo(code);
  if (l) return l.direction === 'rtl' ? 'rtl' : 'ltr';
  return /^(fa|ar|he|ur|ps|ckb|ku|sd|dv|yi)(-|$)/.test(code || '') ? 'rtl' : 'ltr';
}

function langName(code) {
  var l = langInfo(code);
  if (!l) return code;
  var n = l.name || {};
  return n[lang()] || l.nativeName || n[LANGS.def] || code;
}

function i18nText(v) {
  if (v == null) return '';
  if (typeof v === 'string') return v;
  if (typeof v !== 'object') return String(v);
  var l = lang();
  if (v[l]) return v[l];
  if (v[LANGS.def]) return v[LANGS.def];
  for (var i = 0; i < LANGS.active.length; i++) if (v[LANGS.active[i].code]) return v[LANGS.active[i].code];
  var keys = Object.keys(v).sort();
  for (var j = 0; j < keys.length; j++) if (typeof v[keys[j]] === 'string' && v[keys[j]]) return v[keys[j]];
  return '';
}

function getPath(obj, path) {
  var parts = path.split('.');
  for (var i = 0; i < parts.length; i++) {
    if (obj == null) return undefined;
    obj = obj[parts[i]];
  }
  return obj;
}

function setPath(obj, path, value) {
  var parts = path.split('.');
  for (var i = 0; i < parts.length - 1; i++) {
    if (obj[parts[i]] == null || typeof obj[parts[i]] !== 'object') obj[parts[i]] = {};
    obj = obj[parts[i]];
  }
  obj[parts[parts.length - 1]] = value;
}

function clone(v) {
  return v === undefined ? undefined : JSON.parse(JSON.stringify(v));
}

function h(tag, props) {
  var el = tag.indexOf('svg:') === 0 ? document.createElementNS(SVG_NS, tag.slice(4)) : document.createElement(tag);
  if (props) {
    for (var k in props) {
      var v = props[k];
      if (v === null || v === undefined || v === false) continue;
      if (k === 'class') el.setAttribute('class', v);
      else if (k === 'text') el.textContent = v;
      else if (k === 'html') el.innerHTML = v;
      else if (k.indexOf('on') === 0 && typeof v === 'function') el.addEventListener(k.slice(2).toLowerCase(), v);
      else if (k === 'value') el.value = v;
      else if (k === 'checked') el.checked = !!v;
      else el.setAttribute(k, v === true ? '' : v);
    }
  }
  for (var i = 2; i < arguments.length; i++) append(el, arguments[i]);
  return el;
}

function append(el, child) {
  if (child === null || child === undefined || child === false) return;
  if (Array.isArray(child)) { child.forEach(function (c) { append(el, c); }); return; }
  el.appendChild(typeof child === 'object' ? child : document.createTextNode(String(child)));
}

function clear(el) {
  while (el.firstChild) el.removeChild(el.firstChild);
  return el;
}

function ApiError(status, code, message, field) {
  this.status = status;
  this.code = code;
  this.message = message;
  this.field = field;
}

function token() {
  var stores = [];
  try { stores.push(window.localStorage); } catch (e) { stores.push(null); }
  try { stores.push(window.sessionStorage); } catch (e) { stores.push(null); }
  for (var i = 0; i < stores.length; i++) {
    try {
      if (!stores[i]) continue;
      var raw = stores[i].getItem('SESSION_TOKENS');
      if (!raw) continue;
      var j = JSON.parse(raw);
      if (j && j.accessToken) return j.accessToken;
    } catch (e) { continue; }
  }
  return null;
}

function toLogin() {
  if (!/^#\/login/.test(location.hash)) location.hash = '#/login';
}

function qs(params) {
  var out = [];
  Object.keys(params || {}).forEach(function (k) {
    var v = params[k];
    if (v === undefined || v === null || v === '') return;
    out.push(encodeURIComponent(k) + '=' + encodeURIComponent(toLatinDigits(v)));
  });
  return out.length ? '?' + out.join('&') : '';
}

function request(base, method, path, body, isForm) {
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  var init = { method: method, headers: headers, credentials: 'same-origin' };
  if (body !== undefined) {
    if (isForm) init.body = body;
    else { headers['Content-Type'] = 'application/json'; init.body = JSON.stringify(body); }
  }
  return fetch(base + path, init).then(function (res) {
    if (res.status === 401) {
      toLogin();
      throw new ApiError(401, 'UNAUTHENTICATED', '');
    }
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) {
        var code = data && (typeof data.error === 'string' ? data.error : (data.code && typeof data.code === 'string' ? data.code : null));
        var msg = data && (data.message || (data.error && data.error.message)) || '';
        if (!code && /^[A-Z][A-Z0-9_]+$/.test(msg)) code = msg;
        if (code === 'INVALID_LANGUAGE') loadLanguages(true);
        throw new ApiError(res.status, code, msg, data && data.field);
      }
      return data;
    });
  }, function () {
    throw new ApiError(0, 'NETWORK', '');
  });
}

function api(method, path, body, isForm) {
  return request(API_BASE, method, path, body, isForm);
}

function pubApi(method, path, body, isForm) {
  return request(PUB_BASE, method, path, body, isForm);
}

function baseApi(method, path, body, isForm) {
  return request(BASE_BASE, method, path, body, isForm);
}

function consoleApi(method, path) {
  return request(CONSOLE_BASE, method, path);
}

function errorText(err, langCode) {
  if (!err) return t('err.generic');
  if (err.code && has('code.' + err.code)) return t('code.' + err.code, { lang: langName(langCode || LANGS.def) });
  if (err.status === 0) return t('err.network');
  if (err.status === 403) return t('err.forbidden');
  if (err.status === 404 && !err.code) return t('err.notfound');
  if (err.status === 404) return t('err.item.notfound');
  if (err.message && lang() === 'en') return err.message;
  if (err.message && /[؀-ۿ]/.test(err.message)) return err.message;
  return t('err.generic') + (err.code ? ' \u2068(' + err.code + ')\u2069' : '');
}

var ICONS = {
  overview: '<path d="M3 3v18h18"/><path d="M8 17v-4M13 17V8M18 17v-7"/>',
  currencies: '<circle cx="9" cy="9" r="6"/><path d="M18.1 10.4a6 6 0 1 1-7.7 7.7"/><path d="M8 7h1.5v4"/>',
  metas: '<path d="M21 4h-7M10 4H3M21 12h-9M8 12H3M21 20h-5M12 20H3M14 2v4M8 10v4M16 18v4"/>',
  prizes: '<rect x="3" y="8" width="18" height="4" rx="1"/><path d="M12 8v13M19 12v7a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2v-7"/><path d="M7.5 8a2.5 2.5 0 0 1 0-5C10 3 12 8 12 8s2-5 4.5-5a2.5 2.5 0 0 1 0 5"/>',
  products: '<path d="M3 9l1.5-5h15L21 9"/><path d="M3 9a3 3 0 0 0 6 0a3 3 0 0 0 6 0a3 3 0 0 0 6 0"/><path d="M5 12v8h14v-8M10 20v-5h4v5"/>',
  gifts: '<path d="M2 9a3 3 0 0 1 0 6v2a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-2a3 3 0 0 1 0-6V7a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2z"/><path d="M13 5v2M13 11v2M13 17v2"/>',
  payments: '<path d="M4 2v20l2-1 2 1 2-1 2 1 2-1 2 1 2-1 2 1V2l-2 1-2-1-2 1-2-1-2 1-2-1-2 1z"/><path d="M16 8H8M16 12H8M13 16H8"/>',
  categories: '<rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><circle cx="17.5" cy="17.5" r="3.5"/>',
  languages: '<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18"/>',
  tags: '<path d="M12.6 2.6A2 2 0 0 0 11.2 2H4a2 2 0 0 0-2 2v7.2a2 2 0 0 0 .6 1.4l8.7 8.7a2.4 2.4 0 0 0 3.4 0l6.6-6.6a2.4 2.4 0 0 0 0-3.4z"/><circle cx="7.5" cy="7.5" r="1"/>',
  search: '<circle cx="11" cy="11" r="7"/><path d="M21 21l-4.3-4.3"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  close: '<path d="M18 6L6 18M6 6l12 12"/>',
  first: '<path d="M7 6v12M17 18l-6-6 6-6"/>',
  prev: '<path d="M15 18l-6-6 6-6"/>',
  next: '<path d="M9 18l6-6-6-6"/>',
  back: '<path d="M19 12H5M12 19l-7-7 7-7"/>',
  refresh: '<path d="M21 12a9 9 0 1 1-2.6-6.4L21 8"/><path d="M21 3v5h-5"/>',
  trash: '<path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6"/>',
  chevron: '<path d="M6 9l6 6 6-6"/>',
  upload: '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>',
  empty: '<path d="M22 12h-6l-2 3h-4l-2-3H2"/><path d="M5.5 5.1L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.5-6.9A2 2 0 0 0 16.8 4H7.2a2 2 0 0 0-1.7 1.1z"/>',
  alert: '<circle cx="12" cy="12" r="9"/><path d="M12 8v4M12 16h.01"/>',
  check: '<path d="M20 6L9 17l-5-5"/>',
  faqs: '<circle cx="12" cy="12" r="9"/><path d="M9.1 9a3 3 0 0 1 5.8 1c0 2-3 3-3 3M12 17h.01"/>',
  news: '<path d="M4 22h16a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v16a2 2 0 0 1-2 2zm0 0a2 2 0 0 1-2-2v-9c0-1.1.9-2 2-2h2"/><path d="M18 14h-8M15 18h-5M10 6h8v4h-8z"/>',
  comments: '<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>',
  notifications: '<path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9"/><path d="M10.3 21a1.94 1.94 0 0 0 3.4 0"/>',
  send: '<path d="M22 2L11 13"/><path d="M22 2l-7 20-4-9-9-4z"/>',
  eye: '<path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/>',
  reply: '<path d="M9 17l-5-5 5-5"/><path d="M20 18v-2a4 4 0 0 0-4-4H4"/>',
  pin: '<path d="M12 17v5M9 10.8a2 2 0 0 1-1.1 1.8l-1.8.9A2 2 0 0 0 5 15.2V17h14v-1.8a2 2 0 0 0-1.1-1.8l-1.8-.9A2 2 0 0 1 15 10.8V7a1 1 0 0 1 1-1 2 2 0 0 0 0-4H8a2 2 0 0 0 0 4 1 1 0 0 1 1 1z"/>',
  users: '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.9M16 3.1a4 4 0 0 1 0 7.8"/>',
  flask: '<path d="M9 3h6M10 3v6.5L4.5 19a1.5 1.5 0 0 0 1.3 2h12.4a1.5 1.5 0 0 0 1.3-2L14 9.5V3"/><path d="M7 15h10"/>',
  image: '<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="M21 15l-3.1-3.1a2 2 0 0 0-2.8 0L6 21"/>',
  gear: '<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/>',
  mail: '<rect x="2" y="4" width="20" height="16" rx="2"/><path d="M22 7l-10 6L2 7"/>',
  flag: '<path d="M4 22V4a1 1 0 0 1 .4-.8A6 6 0 0 1 8 2c3 0 5 2 7.3 2a6 6 0 0 0 4.1-1.6.5.5 0 0 1 .6.4V14a1 1 0 0 1-.4.8A6 6 0 0 1 16 16c-3 0-5-2-8-2a6 6 0 0 0-4 1.5"/>',
  phone: '<rect x="6" y="2" width="12" height="20" rx="2"/><path d="M11 18h2"/><path d="M9.5 9.5l1.5 1.5 3.5-3.5"/>',
  versions: '<rect x="5" y="2" width="14" height="20" rx="2"/><path d="M12 7v7M9 11l3 3 3-3M10 18h4"/>',
  ban: '<circle cx="12" cy="12" r="10"/><path d="m4.9 4.9 14.2 14.2"/>',
  players: '<path d="M18 21a8 8 0 0 0-16 0"/><circle cx="10" cy="8" r="5"/><path d="M22 20c0-3.4-2-6.5-4-8a5 5 0 0 0-.4-8.7"/>',
  reports: '<path d="M7.9 20A9 9 0 1 0 4 16.1L2 22z"/><path d="M12 8v4M12 16h.01"/>',
  restrictions: '<path d="M20 13c0 5-3.5 7.5-7.7 9a1 1 0 0 1-.6 0C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.2-2.7a1.2 1.2 0 0 1 1.6 0C14.5 3.8 17 5 19 5a1 1 0 0 1 1 1z"/><path d="m9 12 2 2 4-4"/>',
  'chat-words': '<path d="M22 3H2l8 9.5V19l4 2v-8.5z"/>',
  'report-titles': '<path d="M3 5h.01M3 12h.01M3 19h.01M8 5h13M8 12h13M8 19h13"/>',
  tickets: '<path d="M2 9a3 3 0 0 1 0 6v2a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-2a3 3 0 0 1 0-6V7a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2Z"/><path d="M13 5v2M13 17v2M13 11v2"/>',
  'live-chat': '<path d="M14 9a2 2 0 0 1-2 2H6l-4 4V4a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2z"/><path d="M18 9h2a2 2 0 0 1 2 2v11l-4-4h-6a2 2 0 0 1-2-2v-1"/>',
  activity: '<path d="M22 12h-4l-3 9L9 3l-3 9H2"/>',
  referrals: '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M19 8v6M22 11h-6"/>',
  'remote-config': '<rect x="2" y="6" width="20" height="12" rx="6"/><circle cx="16" cy="12" r="2"/>',
  'support-departments': '<path d="M3 21h18M5 21V7l7-4 7 4v14M9 9h1M14 9h1M9 13h1M14 13h1M9 17h6"/>',
  'support-canned': '<path d="M13 2 3 14h9l-1 8 10-12h-9l1-8z"/>',
  'support-settings': '<path d="M3 14h3a2 2 0 0 1 2 2v3a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-7a9 9 0 0 1 18 0v7a2 2 0 0 1-2 2h-1a2 2 0 0 1-2-2v-3a2 2 0 0 1 2-2h3"/>',
  'referral-rules': '<path d="M20 12v10H4V12M2 7h20v5H2zM12 22V7M12 7H7.5a2.5 2.5 0 0 1 0-5C11 2 12 7 12 7zM12 7h4.5a2.5 2.5 0 0 0 0-5C13 2 12 7 12 7z"/>',
  clip: '<path d="m21.44 11.05-9.19 9.19a6 6 0 0 1-8.49-8.49l8.57-8.57A4 4 0 1 1 18 8.84l-8.59 8.57a2 2 0 0 1-2.83-2.83l8.49-8.48"/>',
  star: '<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>',
  history: '<path d="M3 12a9 9 0 1 0 3-6.7L3 8"/><path d="M3 3v5h5M12 7v5l3 2"/>'
};
var MIRROR = { first: 1, prev: 1, next: 1, back: 1, reply: 1 };

var PH_ICONS = {
  items: '<path d="M168,40.58V32A24,24,0,0,0,144,8H112A24,24,0,0,0,88,32v8.58A56.09,56.09,0,0,0,40,96V216a16,16,0,0,0,16,16H200a16,16,0,0,0,16-16V96A56.09,56.09,0,0,0,168,40.58ZM112,24h32a8,8,0,0,1,8,8v8H104V32A8,8,0,0,1,112,24Zm56,136H88v-8a8,8,0,0,1,8-8h64a8,8,0,0,1,8,8ZM88,176h48v8a8,8,0,0,0,16,0v-8h16v40H88Zm112,40H184V152a24,24,0,0,0-24-24H96a24,24,0,0,0-24,24v64H56V96A40,40,0,0,1,96,56h64a40,40,0,0,1,40,40V216ZM152,88a8,8,0,0,1-8,8H112a8,8,0,0,1,0-16h32A8,8,0,0,1,152,88Z"/>',
  overview: '<path d="M104,40H56A16,16,0,0,0,40,56v48a16,16,0,0,0,16,16h48a16,16,0,0,0,16-16V56A16,16,0,0,0,104,40Zm0,64H56V56h48v48Zm96-64H152a16,16,0,0,0-16,16v48a16,16,0,0,0,16,16h48a16,16,0,0,0,16-16V56A16,16,0,0,0,200,40Zm0,64H152V56h48v48Zm-96,32H56a16,16,0,0,0-16,16v48a16,16,0,0,0,16,16h48a16,16,0,0,0,16-16V152A16,16,0,0,0,104,136Zm0,64H56V152h48v48Zm96-64H152a16,16,0,0,0-16,16v48a16,16,0,0,0,16,16h48a16,16,0,0,0,16-16V152A16,16,0,0,0,200,136Zm0,64H152V152h48v48Z"/>',
  currencies: '<path d="M184,89.57V84c0-25.08-37.83-44-88-44S8,58.92,8,84v40c0,20.89,26.25,37.49,64,42.46V172c0,25.08,37.83,44,88,44s88-18.92,88-44V132C248,111.3,222.58,94.68,184,89.57ZM232,132c0,13.22-30.79,28-72,28-3.73,0-7.43-.13-11.08-.37C170.49,151.77,184,139,184,124V105.74C213.87,110.19,232,122.27,232,132ZM72,150.25V126.46A183.74,183.74,0,0,0,96,128a183.74,183.74,0,0,0,24-1.54v23.79A163,163,0,0,1,96,152,163,163,0,0,1,72,150.25Zm96-40.32V124c0,8.39-12.41,17.4-32,22.87V123.5C148.91,120.37,159.84,115.71,168,109.93ZM96,56c41.21,0,72,14.78,72,28s-30.79,28-72,28S24,97.22,24,84,54.79,56,96,56ZM24,124V109.93c8.16,5.78,19.09,10.44,32,13.57v23.37C36.41,141.4,24,132.39,24,124Zm64,48v-4.17c2.63.1,5.29.17,8,.17,3.88,0,7.67-.13,11.39-.35A121.92,121.92,0,0,0,120,171.41v23.46C100.41,189.4,88,180.39,88,172Zm48,26.25V174.4a179.48,179.48,0,0,0,24,1.6,183.74,183.74,0,0,0,24-1.54v23.79a165.45,165.45,0,0,1-48,0Zm64-3.38V171.5c12.91-3.13,23.84-7.79,32-13.57V172C232,180.39,219.59,189.4,200,194.87Z"/>',
  metas: '<path d="M232,208a8,8,0,0,1-8,8H32a8,8,0,0,1-8-8V48a8,8,0,0,1,16,0V156.69l50.34-50.35a8,8,0,0,1,11.32,0L128,132.69,180.69,80H160a8,8,0,0,1,0-16h40a8,8,0,0,1,8,8v40a8,8,0,0,1-16,0V91.31l-58.34,58.35a8,8,0,0,1-11.32,0L96,123.31l-56,56V200H224A8,8,0,0,1,232,208Z"/>',
  prizes: '<path d="M216,72H180.92c.39-.33.79-.65,1.17-1A29.53,29.53,0,0,0,192,49.57,32.62,32.62,0,0,0,158.44,16,29.53,29.53,0,0,0,137,25.91a54.94,54.94,0,0,0-9,14.48,54.94,54.94,0,0,0-9-14.48A29.53,29.53,0,0,0,97.56,16,32.62,32.62,0,0,0,64,49.57,29.53,29.53,0,0,0,73.91,71c.38.33.78.65,1.17,1H40A16,16,0,0,0,24,88v32a16,16,0,0,0,16,16v64a16,16,0,0,0,16,16H200a16,16,0,0,0,16-16V136a16,16,0,0,0,16-16V88A16,16,0,0,0,216,72ZM149,36.51a13.69,13.69,0,0,1,10-4.5h.49A16.62,16.62,0,0,1,176,49.08a13.69,13.69,0,0,1-4.5,10c-9.49,8.4-25.24,11.36-35,12.4C137.7,60.89,141,45.5,149,36.51Zm-64.09.36A16.63,16.63,0,0,1,96.59,32h.49a13.69,13.69,0,0,1,10,4.5c8.39,9.48,11.35,25.2,12.39,34.92-9.72-1-25.44-4-34.92-12.39a13.69,13.69,0,0,1-4.5-10A16.6,16.6,0,0,1,84.87,36.87ZM40,88h80v32H40Zm16,48h64v64H56Zm144,64H136V136h64Zm16-80H136V88h80v32Z"/>',
  products: '<path d="M230.14,58.87A8,8,0,0,0,224,56H62.68L56.6,22.57A8,8,0,0,0,48.73,16H24a8,8,0,0,0,0,16h18L67.56,172.29a24,24,0,0,0,5.33,11.27,28,28,0,1,0,44.4,8.44h45.42A27.75,27.75,0,0,0,160,204a28,28,0,1,0,28-28H91.17a8,8,0,0,1-7.87-6.57L80.13,152h116a24,24,0,0,0,23.61-19.71l12.16-66.86A8,8,0,0,0,230.14,58.87ZM104,204a12,12,0,1,1-12-12A12,12,0,0,1,104,204Zm96,0a12,12,0,1,1-12-12A12,12,0,0,1,200,204Zm4-74.57A8,8,0,0,1,196.1,136H77.22L65.59,72H214.41Z"/>',
  gifts: '<path d="M232,104a8,8,0,0,0,8-8V64a16,16,0,0,0-16-16H32A16,16,0,0,0,16,64V96a8,8,0,0,0,8,8,24,24,0,0,1,0,48,8,8,0,0,0-8,8v32a16,16,0,0,0,16,16H224a16,16,0,0,0,16-16V160a8,8,0,0,0-8-8,24,24,0,0,1,0-48ZM32,167.2a40,40,0,0,0,0-78.4V64H88V192H32Zm192,0V192H104V64H224V88.8a40,40,0,0,0,0,78.4Z"/>',
  payments: '<path d="M224,48H32A16,16,0,0,0,16,64V192a16,16,0,0,0,16,16H224a16,16,0,0,0,16-16V64A16,16,0,0,0,224,48Zm0,16V88H32V64Zm0,128H32V104H224v88Zm-16-24a8,8,0,0,1-8,8H168a8,8,0,0,1,0-16h32A8,8,0,0,1,208,168Zm-64,0a8,8,0,0,1-8,8H120a8,8,0,0,1,0-16h16A8,8,0,0,1,144,168Z"/>',
  categories: '<path d="M224,64H154.67L126.93,43.2a16.12,16.12,0,0,0-9.6-3.2H72A16,16,0,0,0,56,56V72H40A16,16,0,0,0,24,88V200a16,16,0,0,0,16,16H192.89A15.13,15.13,0,0,0,208,200.89V184h16.89A15.13,15.13,0,0,0,240,168.89V80A16,16,0,0,0,224,64ZM192,200H40V88H85.33l29.87,22.4A8,8,0,0,0,120,112h72Zm32-32H208V112a16,16,0,0,0-16-16H122.67L94.93,75.2a16.12,16.12,0,0,0-9.6-3.2H72V56h45.33L147.2,78.4A8,8,0,0,0,152,80h72Z"/>',
  languages: '<path d="M247.15,212.42l-56-112a8,8,0,0,0-14.31,0l-21.71,43.43A88,88,0,0,1,108,126.93,103.65,103.65,0,0,0,135.69,64H160a8,8,0,0,0,0-16H104V32a8,8,0,0,0-16,0V48H32a8,8,0,0,0,0,16h87.63A87.76,87.76,0,0,1,96,116.35a87.74,87.74,0,0,1-19-31,8,8,0,1,0-15.08,5.34A103.63,103.63,0,0,0,84,127a87.55,87.55,0,0,1-52,17,8,8,0,0,0,0,16,103.46,103.46,0,0,0,64-22.08,104.18,104.18,0,0,0,51.44,21.31l-26.6,53.19a8,8,0,0,0,14.31,7.16L148.94,192h70.11l13.79,27.58A8,8,0,0,0,240,224a8,8,0,0,0,7.15-11.58ZM156.94,176,184,121.89,211.05,176Z"/>',
  tags: '<path d="M243.31,136,144,36.69A15.86,15.86,0,0,0,132.69,32H40a8,8,0,0,0-8,8v92.69A15.86,15.86,0,0,0,36.69,144L136,243.31a16,16,0,0,0,22.63,0l84.68-84.68a16,16,0,0,0,0-22.63Zm-96,96L48,132.69V48h84.69L232,147.31ZM96,84A12,12,0,1,1,84,72,12,12,0,0,1,96,84Z"/>',
  search: '<path d="M229.66,218.34l-50.07-50.06a88.11,88.11,0,1,0-11.31,11.31l50.06,50.07a8,8,0,0,0,11.32-11.32ZM40,112a72,72,0,1,1,72,72A72.08,72.08,0,0,1,40,112Z"/>',
  plus: '<path d="M224,128a8,8,0,0,1-8,8H136v80a8,8,0,0,1-16,0V136H40a8,8,0,0,1,0-16h80V40a8,8,0,0,1,16,0v80h80A8,8,0,0,1,224,128Z"/>',
  close: '<path d="M205.66,194.34a8,8,0,0,1-11.32,11.32L128,139.31,61.66,205.66a8,8,0,0,1-11.32-11.32L116.69,128,50.34,61.66A8,8,0,0,1,61.66,50.34L128,116.69l66.34-66.35a8,8,0,0,1,11.32,11.32L139.31,128Z"/>',
  first: '<path d="M197.66,202.34a8,8,0,0,1-11.32,11.32l-80-80a8,8,0,0,1,0-11.32l80-80a8,8,0,0,1,11.32,11.32L123.31,128ZM72,40a8,8,0,0,0-8,8V208a8,8,0,0,0,16,0V48A8,8,0,0,0,72,40Z"/>',
  prev: '<path d="M165.66,202.34a8,8,0,0,1-11.32,11.32l-80-80a8,8,0,0,1,0-11.32l80-80a8,8,0,0,1,11.32,11.32L91.31,128Z"/>',
  next: '<path d="M181.66,133.66l-80,80a8,8,0,0,1-11.32-11.32L164.69,128,90.34,53.66a8,8,0,0,1,11.32-11.32l80,80A8,8,0,0,1,181.66,133.66Z"/>',
  back: '<path d="M224,128a8,8,0,0,1-8,8H59.31l58.35,58.34a8,8,0,0,1-11.32,11.32l-72-72a8,8,0,0,1,0-11.32l72-72a8,8,0,0,1,11.32,11.32L59.31,120H216A8,8,0,0,1,224,128Z"/>',
  refresh: '<path d="M224,48V96a8,8,0,0,1-8,8H168a8,8,0,0,1,0-16h28.69L182.06,73.37a79.56,79.56,0,0,0-56.13-23.43h-.45A79.52,79.52,0,0,0,69.59,72.71,8,8,0,0,1,58.41,61.27a96,96,0,0,1,135,.79L208,76.69V48a8,8,0,0,1,16,0ZM186.41,183.29a80,80,0,0,1-112.47-.66L59.31,168H88a8,8,0,0,0,0-16H40a8,8,0,0,0-8,8v48a8,8,0,0,0,16,0V179.31l14.63,14.63A95.43,95.43,0,0,0,130,222.06h.53a95.36,95.36,0,0,0,67.07-27.33,8,8,0,0,0-11.18-11.44Z"/>',
  trash: '<path d="M216,48H176V40a24,24,0,0,0-24-24H104A24,24,0,0,0,80,40v8H40a8,8,0,0,0,0,16h8V208a16,16,0,0,0,16,16H192a16,16,0,0,0,16-16V64h8a8,8,0,0,0,0-16ZM96,40a8,8,0,0,1,8-8h48a8,8,0,0,1,8,8v8H96Zm96,168H64V64H192ZM112,104v64a8,8,0,0,1-16,0V104a8,8,0,0,1,16,0Zm48,0v64a8,8,0,0,1-16,0V104a8,8,0,0,1,16,0Z"/>',
  chevron: '<path d="M213.66,101.66l-80,80a8,8,0,0,1-11.32,0l-80-80A8,8,0,0,1,53.66,90.34L128,164.69l74.34-74.35a8,8,0,0,1,11.32,11.32Z"/>',
  upload: '<path d="M224,144v64a8,8,0,0,1-8,8H40a8,8,0,0,1-8-8V144a8,8,0,0,1,16,0v56H208V144a8,8,0,0,1,16,0ZM93.66,77.66,120,51.31V144a8,8,0,0,0,16,0V51.31l26.34,26.35a8,8,0,0,0,11.32-11.32l-40-40a8,8,0,0,0-11.32,0l-40,40A8,8,0,0,0,93.66,77.66Z"/>',
  empty: '<path d="M208,32H48A16,16,0,0,0,32,48V208a16,16,0,0,0,16,16H208a16,16,0,0,0,16-16V48A16,16,0,0,0,208,32Zm0,16V152h-28.7A15.86,15.86,0,0,0,168,156.69L148.69,176H107.31L88,156.69A15.86,15.86,0,0,0,76.69,152H48V48Zm0,160H48V168H76.69L96,187.31A15.86,15.86,0,0,0,107.31,192h41.38A15.86,15.86,0,0,0,160,187.31L179.31,168H208v40Z"/>',
  alert: '<path d="M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm0,192a88,88,0,1,1,88-88A88.1,88.1,0,0,1,128,216Zm-8-80V80a8,8,0,0,1,16,0v56a8,8,0,0,1-16,0Zm20,36a12,12,0,1,1-12-12A12,12,0,0,1,140,172Z"/>',
  check: '<path d="M229.66,77.66l-128,128a8,8,0,0,1-11.32,0l-56-56a8,8,0,0,1,11.32-11.32L96,188.69,218.34,66.34a8,8,0,0,1,11.32,11.32Z"/>',
  faqs: '<path d="M140,180a12,12,0,1,1-12-12A12,12,0,0,1,140,180ZM128,72c-22.06,0-40,16.15-40,36v4a8,8,0,0,0,16,0v-4c0-11,10.77-20,24-20s24,9,24,20-10.77,20-24,20a8,8,0,0,0-8,8v8a8,8,0,0,0,16,0v-.72c18.24-3.35,32-17.9,32-35.28C168,88.15,150.06,72,128,72Zm104,56A104,104,0,1,1,128,24,104.11,104.11,0,0,1,232,128Zm-16,0a88,88,0,1,0-88,88A88.1,88.1,0,0,0,216,128Z"/>',
  news: '<path d="M88,112a8,8,0,0,1,8-8h80a8,8,0,0,1,0,16H96A8,8,0,0,1,88,112Zm8,40h80a8,8,0,0,0,0-16H96a8,8,0,0,0,0,16ZM232,64V184a24,24,0,0,1-24,24H32A24,24,0,0,1,8,184.11V88a8,8,0,0,1,16,0v96a8,8,0,0,0,16,0V64A16,16,0,0,1,56,48H216A16,16,0,0,1,232,64Zm-16,0H56V184a23.84,23.84,0,0,1-1.37,8H208a8,8,0,0,0,8-8Z"/>',
  comments: '<path d="M168,112a8,8,0,0,1-8,8H96a8,8,0,0,1,0-16h64A8,8,0,0,1,168,112Zm-8,24H96a8,8,0,0,0,0,16h64a8,8,0,0,0,0-16Zm72-8A104,104,0,0,1,79.12,219.82L45.07,231.17a16,16,0,0,1-20.24-20.24l11.35-34.05A104,104,0,1,1,232,128Zm-16,0A88,88,0,1,0,51.81,172.06a8,8,0,0,1,.66,6.54L40,216,77.4,203.53a7.85,7.85,0,0,1,2.53-.42,8,8,0,0,1,4,1.08A88,88,0,0,0,216,128Z"/>',
  notifications: '<path d="M221.8,175.94C216.25,166.38,208,139.33,208,104a80,80,0,1,0-160,0c0,35.34-8.26,62.38-13.81,71.94A16,16,0,0,0,48,200H88.81a40,40,0,0,0,78.38,0H208a16,16,0,0,0,13.8-24.06ZM128,216a24,24,0,0,1-22.62-16h45.24A24,24,0,0,1,128,216ZM48,184c7.7-13.24,16-43.92,16-80a64,64,0,1,1,128,0c0,36.05,8.28,66.73,16,80Z"/>',
  send: '<path d="M227.32,28.68a16,16,0,0,0-15.66-4.08l-.15,0L19.57,82.84a16,16,0,0,0-2.49,29.8L102,154l41.3,84.87A15.86,15.86,0,0,0,157.74,248q.69,0,1.38-.06a15.88,15.88,0,0,0,14-11.51l58.2-191.94c0-.05,0-.1,0-.15A16,16,0,0,0,227.32,28.68ZM157.83,231.85l-.05.14,0-.07-40.06-82.3,48-48a8,8,0,0,0-11.31-11.31l-48,48L24.08,98.25l-.07,0,.14,0L216,40Z"/>',
  eye: '<path d="M247.31,124.76c-.35-.79-8.82-19.58-27.65-38.41C194.57,61.26,162.88,48,128,48S61.43,61.26,36.34,86.35C17.51,105.18,9,124,8.69,124.76a8,8,0,0,0,0,6.5c.35.79,8.82,19.57,27.65,38.4C61.43,194.74,93.12,208,128,208s66.57-13.26,91.66-38.34c18.83-18.83,27.3-37.61,27.65-38.4A8,8,0,0,0,247.31,124.76ZM128,192c-30.78,0-57.67-11.19-79.93-33.25A133.47,133.47,0,0,1,25,128,133.33,133.33,0,0,1,48.07,97.25C70.33,75.19,97.22,64,128,64s57.67,11.19,79.93,33.25A133.46,133.46,0,0,1,231.05,128C223.84,141.46,192.43,192,128,192Zm0-112a48,48,0,1,0,48,48A48.05,48.05,0,0,0,128,80Zm0,80a32,32,0,1,1,32-32A32,32,0,0,1,128,160Z"/>',
  reply: '<path d="M232,200a8,8,0,0,1-16,0,88.1,88.1,0,0,0-88-88H51.31l34.35,34.34a8,8,0,0,1-11.32,11.32l-48-48a8,8,0,0,1,0-11.32l48-48A8,8,0,0,1,85.66,61.66L51.31,96H128A104.11,104.11,0,0,1,232,200Z"/>',
  pin: '<path d="M235.32,81.37,174.63,20.69a16,16,0,0,0-22.63,0L98.37,74.49c-10.66-3.34-35-7.37-60.4,13.14a16,16,0,0,0-1.29,23.78L85,159.71,42.34,202.34a8,8,0,0,0,11.32,11.32L96.29,171l48.29,48.29A16,16,0,0,0,155.9,224c.38,0,.75,0,1.13,0a15.93,15.93,0,0,0,11.64-6.33c19.64-26.1,17.75-47.32,13.19-60L235.33,104A16,16,0,0,0,235.32,81.37ZM224,92.69h0l-57.27,57.46a8,8,0,0,0-1.49,9.22c9.46,18.93-1.8,38.59-9.34,48.62L48,100.08c12.08-9.74,23.64-12.31,32.48-12.31A40.13,40.13,0,0,1,96.81,91a8,8,0,0,0,9.25-1.51L163.32,32,224,92.68Z"/>',
  users: '<path d="M117.25,157.92a60,60,0,1,0-66.5,0A95.83,95.83,0,0,0,3.53,195.63a8,8,0,1,0,13.4,8.74,80,80,0,0,1,134.14,0,8,8,0,0,0,13.4-8.74A95.83,95.83,0,0,0,117.25,157.92ZM40,108a44,44,0,1,1,44,44A44.05,44.05,0,0,1,40,108Zm210.14,98.7a8,8,0,0,1-11.07-2.33A79.83,79.83,0,0,0,172,168a8,8,0,0,1,0-16,44,44,0,1,0-16.34-84.87,8,8,0,1,1-5.94-14.85,60,60,0,0,1,55.53,105.64,95.83,95.83,0,0,1,47.22,37.71A8,8,0,0,1,250.14,206.7Z"/>',
  flask: '<path d="M221.69,199.77,160,96.92V40h8a8,8,0,0,0,0-16H88a8,8,0,0,0,0,16h8V96.92L34.31,199.77A16,16,0,0,0,48,224H208a16,16,0,0,0,13.72-24.23ZM110.86,103.25A7.93,7.93,0,0,0,112,99.14V40h32V99.14a7.93,7.93,0,0,0,1.14,4.11L183.36,167c-12,2.37-29.07,1.37-51.75-10.11-15.91-8.05-31.05-12.32-45.22-12.81ZM48,208l28.54-47.58c14.25-1.74,30.31,1.85,47.82,10.72,19,9.61,35,12.88,48,12.88a69.89,69.89,0,0,0,19.55-2.7L208,208Z"/>',
  image: '<path d="M216,40H40A16,16,0,0,0,24,56V200a16,16,0,0,0,16,16H216a16,16,0,0,0,16-16V56A16,16,0,0,0,216,40Zm0,16V158.75l-26.07-26.06a16,16,0,0,0-22.63,0l-20,20-44-44a16,16,0,0,0-22.62,0L40,149.37V56ZM40,172l52-52,80,80H40Zm176,28H194.63l-36-36,20-20L216,181.38V200ZM144,100a12,12,0,1,1,12,12A12,12,0,0,1,144,100Z"/>',
  gear: '<path d="M128,80a48,48,0,1,0,48,48A48.05,48.05,0,0,0,128,80Zm0,80a32,32,0,1,1,32-32A32,32,0,0,1,128,160Zm109.94-52.79a8,8,0,0,0-3.89-5.4l-29.83-17-.12-33.62a8,8,0,0,0-2.83-6.08,111.91,111.91,0,0,0-36.72-20.67,8,8,0,0,0-6.46.59L128,41.85,97.88,25a8,8,0,0,0-6.47-.6A112.1,112.1,0,0,0,54.73,45.15a8,8,0,0,0-2.83,6.07l-.15,33.65-29.83,17a8,8,0,0,0-3.89,5.4,106.47,106.47,0,0,0,0,41.56,8,8,0,0,0,3.89,5.4l29.83,17,.12,33.62a8,8,0,0,0,2.83,6.08,111.91,111.91,0,0,0,36.72,20.67,8,8,0,0,0,6.46-.59L128,214.15,158.12,231a7.91,7.91,0,0,0,3.9,1,8.09,8.09,0,0,0,2.57-.42,112.1,112.1,0,0,0,36.68-20.73,8,8,0,0,0,2.83-6.07l.15-33.65,29.83-17a8,8,0,0,0,3.89-5.4A106.47,106.47,0,0,0,237.94,107.21Zm-15,34.91-28.57,16.25a8,8,0,0,0-3,3c-.58,1-1.19,2.06-1.81,3.06a7.94,7.94,0,0,0-1.22,4.21l-.15,32.25a95.89,95.89,0,0,1-25.37,14.3L134,199.13a8,8,0,0,0-3.91-1h-.19c-1.21,0-2.43,0-3.64,0a8.08,8.08,0,0,0-4.1,1l-28.84,16.1A96,96,0,0,1,67.88,201l-.11-32.2a8,8,0,0,0-1.22-4.22c-.62-1-1.23-2-1.8-3.06a8.09,8.09,0,0,0-3-3.06l-28.6-16.29a90.49,90.49,0,0,1,0-28.26L61.67,97.63a8,8,0,0,0,3-3c.58-1,1.19-2.06,1.81-3.06a7.94,7.94,0,0,0,1.22-4.21l.15-32.25a95.89,95.89,0,0,1,25.37-14.3L122,56.87a8,8,0,0,0,4.1,1c1.21,0,2.43,0,3.64,0a8.08,8.08,0,0,0,4.1-1l28.84-16.1A96,96,0,0,1,188.12,55l.11,32.2a8,8,0,0,0,1.22,4.22c.62,1,1.23,2,1.8,3.06a8.09,8.09,0,0,0,3,3.06l28.6,16.29A90.49,90.49,0,0,1,222.9,142.12Z"/>',
  mail: '<path d="M224,48H32a8,8,0,0,0-8,8V192a16,16,0,0,0,16,16H216a16,16,0,0,0,16-16V56A8,8,0,0,0,224,48ZM203.43,64,128,133.15,52.57,64ZM216,192H40V74.19l82.59,75.71a8,8,0,0,0,10.82,0L216,74.19V192Z"/>',
  flag: '<path d="M42.76,50A8,8,0,0,0,40,56V224a8,8,0,0,0,16,0V179.77c26.79-21.16,49.87-9.75,76.45,3.41,16.4,8.11,34.06,16.85,53,16.85,13.93,0,28.54-4.75,43.82-18a8,8,0,0,0,2.76-6V56A8,8,0,0,0,218.76,50c-28,24.23-51.72,12.49-79.21-1.12C111.07,34.76,78.78,18.79,42.76,50ZM216,172.25c-26.79,21.16-49.87,9.74-76.45-3.41-25-12.35-52.81-26.13-83.55-8.4V59.79c26.79-21.16,49.87-9.75,76.45,3.4,25,12.35,52.82,26.13,83.55,8.4Z"/>',
  phone: '<path d="M176,16H80A24,24,0,0,0,56,40V216a24,24,0,0,0,24,24h96a24,24,0,0,0,24-24V40A24,24,0,0,0,176,16ZM72,64H184V192H72Zm8-32h96a8,8,0,0,1,8,8v8H72V40A8,8,0,0,1,80,32Zm96,192H80a8,8,0,0,1-8-8v-8H184v8A8,8,0,0,1,176,224Z"/>',
  versions: '<path d="M176,16H80A24,24,0,0,0,56,40V216a24,24,0,0,0,24,24h96a24,24,0,0,0,24-24V40A24,24,0,0,0,176,16Zm8,200a8,8,0,0,1-8,8H80a8,8,0,0,1-8-8V40a8,8,0,0,1,8-8h96a8,8,0,0,1,8,8ZM140,60a12,12,0,1,1-12-12A12,12,0,0,1,140,60Z"/>',
  ban: '<path d="M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm88,104a87.56,87.56,0,0,1-20.41,56.28L71.72,60.4A88,88,0,0,1,216,128ZM40,128A87.56,87.56,0,0,1,60.41,71.72L184.28,195.6A88,88,0,0,1,40,128Z"/>',
  players: '<path d="M230.92,212c-15.23-26.33-38.7-45.21-66.09-54.16a72,72,0,1,0-73.66,0C63.78,166.78,40.31,185.66,25.08,212a8,8,0,1,0,13.85,8c18.84-32.56,52.14-52,89.07-52s70.23,19.44,89.07,52a8,8,0,1,0,13.85-8ZM72,96a56,56,0,1,1,56,56A56.06,56.06,0,0,1,72,96Z"/>',
  reports: '<path d="M42.76,50A8,8,0,0,0,40,56V224a8,8,0,0,0,16,0V179.77c26.79-21.16,49.87-9.75,76.45,3.41,16.4,8.11,34.06,16.85,53,16.85,13.93,0,28.54-4.75,43.82-18a8,8,0,0,0,2.76-6V56A8,8,0,0,0,218.76,50c-28,24.23-51.72,12.49-79.21-1.12C111.07,34.76,78.78,18.79,42.76,50ZM216,172.25c-26.79,21.16-49.87,9.74-76.45-3.41-25-12.35-52.81-26.13-83.55-8.4V59.79c26.79-21.16,49.87-9.75,76.45,3.4,25,12.35,52.82,26.13,83.55,8.4Z"/>',
  restrictions: '<path d="M213.92,218.62l-160-176A8,8,0,0,0,42.08,53.38L80,95.09V128a48,48,0,0,0,69.11,43.12l11.1,12.2A63.41,63.41,0,0,1,128,192a64.07,64.07,0,0,1-64-64,8,8,0,0,0-16,0,80.11,80.11,0,0,0,72,79.6V240a8,8,0,0,0,16,0V207.59a78.83,78.83,0,0,0,35.16-12.22l30.92,34a8,8,0,1,0,11.84-10.76ZM128,160a32,32,0,0,1-32-32V112.69l41.66,45.82A32,32,0,0,1,128,160Zm57.52-3.91A63.32,63.32,0,0,0,192,128a8,8,0,0,1,16,0,79.16,79.16,0,0,1-8.11,35.12,8,8,0,0,1-7.19,4.49,7.88,7.88,0,0,1-3.51-.82A8,8,0,0,1,185.52,156.09ZM84,44.87A48,48,0,0,1,176,64v64a49.19,49.19,0,0,1-.26,5,8,8,0,0,1-8,7.17,8.13,8.13,0,0,1-.84,0,8,8,0,0,1-7.12-8.79c.11-1.1.17-2.24.17-3.36V64A32,32,0,0,0,98.64,51.25,8,8,0,1,1,84,44.87Z"/>',
  'chat-words': '<path d="M208,56V88a8,8,0,0,1-16,0V64H136V192h24a8,8,0,0,1,0,16H96a8,8,0,0,1,0-16h24V64H64V88a8,8,0,0,1-16,0V56a8,8,0,0,1,8-8H200A8,8,0,0,1,208,56Z"/>',
  'report-titles': '<path d="M80,64a8,8,0,0,1,8-8H216a8,8,0,0,1,0,16H88A8,8,0,0,1,80,64Zm136,56H88a8,8,0,0,0,0,16H216a8,8,0,0,0,0-16Zm0,64H88a8,8,0,0,0,0,16H216a8,8,0,0,0,0-16ZM44,52A12,12,0,1,0,56,64,12,12,0,0,0,44,52Zm0,64a12,12,0,1,0,12,12A12,12,0,0,0,44,116Zm0,64a12,12,0,1,0,12,12A12,12,0,0,0,44,180Z"/>',
  tickets: '<path d="M232,104a8,8,0,0,0,8-8V64a16,16,0,0,0-16-16H32A16,16,0,0,0,16,64V96a8,8,0,0,0,8,8,24,24,0,0,1,0,48,8,8,0,0,0-8,8v32a16,16,0,0,0,16,16H224a16,16,0,0,0,16-16V160a8,8,0,0,0-8-8,24,24,0,0,1,0-48ZM32,167.2a40,40,0,0,0,0-78.4V64H88V192H32Zm192,0V192H104V64H224V88.8a40,40,0,0,0,0,78.4Z"/>',
  'live-chat': '<path d="M201.89,54.66A103.43,103.43,0,0,0,128.79,24H128A104,104,0,0,0,24,128v56a24,24,0,0,0,24,24H64a24,24,0,0,0,24-24V144a24,24,0,0,0-24-24H40.36A88.12,88.12,0,0,1,190.54,65.93,87.39,87.39,0,0,1,215.65,120H192a24,24,0,0,0-24,24v40a24,24,0,0,0,24,24h24a24,24,0,0,1-24,24H136a8,8,0,0,0,0,16h56a40,40,0,0,0,40-40V128A103.41,103.41,0,0,0,201.89,54.66ZM64,136a8,8,0,0,1,8,8v40a8,8,0,0,1-8,8H48a8,8,0,0,1-8-8V136Zm128,56a8,8,0,0,1-8-8V144a8,8,0,0,1,8-8h24v56Z"/>',
  activity: '<path d="M240,128a8,8,0,0,1-8,8H204.94l-37.78,75.58A8,8,0,0,1,160,216h-.4a8,8,0,0,1-7.08-5.14L95.35,60.76,63.28,131.31A8,8,0,0,1,56,136H24a8,8,0,0,1,0-16H50.85L88.72,36.69a8,8,0,0,1,14.76.46l57.51,151,31.85-63.71A8,8,0,0,1,200,120h32A8,8,0,0,1,240,128Z"/>',
  referrals: '<path d="M176,160a39.89,39.89,0,0,0-28.62,12.09l-46.1-29.63a39.8,39.8,0,0,0,0-28.92l46.1-29.63a40,40,0,1,0-8.66-13.45l-46.1,29.63a40,40,0,1,0,0,55.82l46.1,29.63A40,40,0,1,0,176,160Zm0-128a24,24,0,1,1-24,24A24,24,0,0,1,176,32ZM64,152a24,24,0,1,1,24-24A24,24,0,0,1,64,152Zm112,72a24,24,0,1,1,24-24A24,24,0,0,1,176,224Z"/>',
  'remote-config': '<path d="M40,88H73a32,32,0,0,0,62,0h81a8,8,0,0,0,0-16H135a32,32,0,0,0-62,0H40a8,8,0,0,0,0,16Zm64-24A16,16,0,1,1,88,80,16,16,0,0,1,104,64ZM216,168H199a32,32,0,0,0-62,0H40a8,8,0,0,0,0,16h97a32,32,0,0,0,62,0h17a8,8,0,0,0,0-16Zm-48,24a16,16,0,1,1,16-16A16,16,0,0,1,168,192Z"/>',
  'support-departments': '<path d="M240,208H224V96a16,16,0,0,0-16-16H144V32a16,16,0,0,0-24.88-13.32L39.12,72A16,16,0,0,0,32,85.34V208H16a8,8,0,0,0,0,16H240a8,8,0,0,0,0-16ZM208,96V208H144V96ZM48,85.34,128,32V208H48ZM112,112v16a8,8,0,0,1-16,0V112a8,8,0,1,1,16,0Zm-32,0v16a8,8,0,0,1-16,0V112a8,8,0,1,1,16,0Zm0,56v16a8,8,0,0,1-16,0V168a8,8,0,0,1,16,0Zm32,0v16a8,8,0,0,1-16,0V168a8,8,0,0,1,16,0Z"/>',
  'support-canned': '<path d="M172,112a8,8,0,0,1-8,8H96a8,8,0,0,1,0-16h68A8,8,0,0,1,172,112Zm-8,24H96a8,8,0,0,0,0,16h68a8,8,0,0,0,0-16Zm68-12A100.11,100.11,0,0,1,132,224H48a16,16,0,0,1-16-16V124a100,100,0,0,1,200,0Zm-16,0a84,84,0,0,0-168,0v84h84A84.09,84.09,0,0,0,216,124Z"/>',
  'support-settings': '<path d="M128,80a48,48,0,1,0,48,48A48.05,48.05,0,0,0,128,80Zm0,80a32,32,0,1,1,32-32A32,32,0,0,1,128,160Zm109.94-52.79a8,8,0,0,0-3.89-5.4l-29.83-17-.12-33.62a8,8,0,0,0-2.83-6.08,111.91,111.91,0,0,0-36.72-20.67,8,8,0,0,0-6.46.59L128,41.85,97.88,25a8,8,0,0,0-6.47-.6A112.1,112.1,0,0,0,54.73,45.15a8,8,0,0,0-2.83,6.07l-.15,33.65-29.83,17a8,8,0,0,0-3.89,5.4,106.47,106.47,0,0,0,0,41.56,8,8,0,0,0,3.89,5.4l29.83,17,.12,33.62a8,8,0,0,0,2.83,6.08,111.91,111.91,0,0,0,36.72,20.67,8,8,0,0,0,6.46-.59L128,214.15,158.12,231a7.91,7.91,0,0,0,3.9,1,8.09,8.09,0,0,0,2.57-.42,112.1,112.1,0,0,0,36.68-20.73,8,8,0,0,0,2.83-6.07l.15-33.65,29.83-17a8,8,0,0,0,3.89-5.4A106.47,106.47,0,0,0,237.94,107.21Zm-15,34.91-28.57,16.25a8,8,0,0,0-3,3c-.58,1-1.19,2.06-1.81,3.06a7.94,7.94,0,0,0-1.22,4.21l-.15,32.25a95.89,95.89,0,0,1-25.37,14.3L134,199.13a8,8,0,0,0-3.91-1h-.19c-1.21,0-2.43,0-3.64,0a8.08,8.08,0,0,0-4.1,1l-28.84,16.1A96,96,0,0,1,67.88,201l-.11-32.2a8,8,0,0,0-1.22-4.22c-.62-1-1.23-2-1.8-3.06a8.09,8.09,0,0,0-3-3.06l-28.6-16.29a90.49,90.49,0,0,1,0-28.26L61.67,97.63a8,8,0,0,0,3-3c.58-1,1.19-2.06,1.81-3.06a7.94,7.94,0,0,0,1.22-4.21l.15-32.25a95.89,95.89,0,0,1,25.37-14.3L122,56.87a8,8,0,0,0,4.1,1c1.21,0,2.43,0,3.64,0a8.08,8.08,0,0,0,4.1-1l28.84-16.1A96,96,0,0,1,188.12,55l.11,32.2a8,8,0,0,0,1.22,4.22c.62,1,1.23,2,1.8,3.06a8.09,8.09,0,0,0,3,3.06l28.6,16.29A90.49,90.49,0,0,1,222.9,142.12Z"/>',
  'referral-rules': '<path d="M254.3,107.91,228.78,56.85a16,16,0,0,0-21.47-7.15L182.44,62.13,130.05,48.27a8.14,8.14,0,0,0-4.1,0L73.56,62.13,48.69,49.7a16,16,0,0,0-21.47,7.15L1.7,107.9a16,16,0,0,0,7.15,21.47l27,13.51,55.49,39.63a8.06,8.06,0,0,0,2.71,1.25l64,16a8,8,0,0,0,7.6-2.1l55.07-55.08,26.42-13.21a16,16,0,0,0,7.15-21.46Zm-54.89,33.37L165,113.72a8,8,0,0,0-10.68.61C136.51,132.27,116.66,130,104,122L147.24,80h31.81l27.21,54.41ZM41.53,64,62,74.22,36.43,125.27,16,115.06Zm116,119.13L99.42,168.61l-49.2-35.14,28-56L128,64.28l9.8,2.59-45,43.68-.08.09a16,16,0,0,0,2.72,24.81c20.56,13.13,45.37,11,64.91-5L188,152.66Zm62-57.87-25.52-51L214.47,64,240,115.06Zm-87.75,92.67a8,8,0,0,1-7.75,6.06,8.13,8.13,0,0,1-1.95-.24L80.41,213.33a7.89,7.89,0,0,1-2.71-1.25L51.35,193.26a8,8,0,0,1,9.3-13l25.11,17.94L126,208.24A8,8,0,0,1,131.82,217.94Z"/>',
  clip: '<path d="M209.66,122.34a8,8,0,0,1,0,11.32l-82.05,82a56,56,0,0,1-79.2-79.21L147.67,35.73a40,40,0,1,1,56.61,56.55L105,193A24,24,0,1,1,71,159L154.3,74.38A8,8,0,1,1,165.7,85.6L82.39,170.31a8,8,0,1,0,11.27,11.36L192.93,81A24,24,0,1,0,159,47L59.76,147.68a40,40,0,1,0,56.53,56.62l82.06-82A8,8,0,0,1,209.66,122.34Z"/>',
  star: '<path d="M239.18,97.26A16.38,16.38,0,0,0,224.92,86l-59-4.76L143.14,26.15a16.36,16.36,0,0,0-30.27,0L90.11,81.23,31.08,86a16.46,16.46,0,0,0-9.37,28.86l45,38.83L53,211.75a16.38,16.38,0,0,0,24.5,17.82L128,198.49l50.53,31.08A16.4,16.4,0,0,0,203,211.75l-13.76-58.07,45-38.83A16.43,16.43,0,0,0,239.18,97.26Zm-15.34,5.47-48.7,42a8,8,0,0,0-2.56,7.91l14.88,62.8a.37.37,0,0,1-.17.48c-.18.14-.23.11-.38,0l-54.72-33.65a8,8,0,0,0-8.38,0L69.09,215.94c-.15.09-.19.12-.38,0a.37.37,0,0,1-.17-.48l14.88-62.8a8,8,0,0,0-2.56-7.91l-48.7-42c-.12-.1-.23-.19-.13-.5s.18-.27.33-.29l63.92-5.16A8,8,0,0,0,103,91.86l24.62-59.61c.08-.17.11-.25.35-.25s.27.08.35.25L153,91.86a8,8,0,0,0,6.75,4.92l63.92,5.16c.15,0,.24,0,.33.29S224,102.63,223.84,102.73Z"/>',
  history: '<path d="M136,80v43.47l36.12,21.67a8,8,0,0,1-8.24,13.72l-40-24A8,8,0,0,1,120,128V80a8,8,0,0,1,16,0Zm-8-48A95.44,95.44,0,0,0,60.08,60.15C52.81,67.51,46.35,74.59,40,82V64a8,8,0,0,0-16,0v40a8,8,0,0,0,8,8H72a8,8,0,0,0,0-16H49c7.15-8.42,14.27-16.35,22.39-24.57a80,80,0,1,1,1.66,114.75,8,8,0,1,0-11,11.64A96,96,0,1,0,128,32Z"/>',
  'event-types': '<path d="M215.79,118.17a8,8,0,0,0-5-5.66L153.18,90.9l14.66-73.33a8,8,0,0,0-13.69-7l-112,120a8,8,0,0,0,3,13l57.63,21.61L88.16,238.43a8,8,0,0,0,13.69,7l112-120A8,8,0,0,0,215.79,118.17ZM109.37,214l10.47-52.38a8,8,0,0,0-5-9.06L62,132.71l84.62-90.66L136.16,94.43a8,8,0,0,0,5,9.06l52.8,19.8Z"/>',
  'event-rules': '<path d="M232,64H208V48a8,8,0,0,0-8-8H56a8,8,0,0,0-8,8V64H24A16,16,0,0,0,8,80V96a40,40,0,0,0,40,40h3.65A80.13,80.13,0,0,0,120,191.61V216H96a8,8,0,0,0,0,16h64a8,8,0,0,0,0-16H136V191.58c31.94-3.23,58.44-25.64,68.08-55.58H208a40,40,0,0,0,40-40V80A16,16,0,0,0,232,64ZM48,120A24,24,0,0,1,24,96V80H48v32q0,4,.39,8Zm144-8.9c0,35.52-29,64.64-64,64.9a64,64,0,0,1-64-64V56H192ZM232,96a24,24,0,0,1-24,24h-.5a81.81,81.81,0,0,0,.5-8.9V80h24Z"/>',
  'meta-rules': '<path d="M64,105V40a8,8,0,0,0-16,0v65a32,32,0,0,0,0,62v49a8,8,0,0,0,16,0V167a32,32,0,0,0,0-62Zm-8,47a16,16,0,1,1,16-16A16,16,0,0,1,56,152Zm80-95V40a8,8,0,0,0-16,0V57a32,32,0,0,0,0,62v97a8,8,0,0,0,16,0V119a32,32,0,0,0,0-62Zm-8,47a16,16,0,1,1,16-16A16,16,0,0,1,128,104Zm104,64a32.06,32.06,0,0,0-24-31V40a8,8,0,0,0-16,0v97a32,32,0,0,0,0,62v17a8,8,0,0,0,16,0V199A32.06,32.06,0,0,0,232,168Zm-32,16a16,16,0,1,1,16-16A16,16,0,0,1,200,184Z"/>',
  'level-tables': '<path d="M215.52,213.26,164.51,73l9.09-25H184a8,8,0,0,0,0-16H88a8,8,0,0,0,0,16h4.58L32.48,213.26a8,8,0,0,0,4.79,10.26A8.14,8.14,0,0,0,40,224a8,8,0,0,0,7.52-5.27L57.24,192h47l-7.74,21.26a8,8,0,0,0,4.79,10.26A8.14,8.14,0,0,0,104,224a8,8,0,0,0,7.52-5.27L130,168H182l18.45,50.73A8,8,0,0,0,208,224a8.14,8.14,0,0,0,2.73-.48A8,8,0,0,0,215.52,213.26Zm-88-85.26h-47L92.15,96h47Zm29.09-80L144.94,80H98L109.6,48ZM63.06,176,74.7,144h47L110,176Zm72.72-24L156,96.41,176.21,152Z"/>',
  streaks: '<path d="M173.79,51.48a221.25,221.25,0,0,0-41.67-34.34,8,8,0,0,0-8.24,0A221.25,221.25,0,0,0,82.21,51.48C54.59,80.48,40,112.47,40,144a88,88,0,0,0,176,0C216,112.47,201.41,80.48,173.79,51.48ZM96,184c0-27.67,22.53-47.28,32-54.3,9.48,7,32,26.63,32,54.3a32,32,0,0,1-64,0Zm77.27,15.93A47.8,47.8,0,0,0,176,184c0-44-42.09-69.79-43.88-70.86a8,8,0,0,0-8.24,0C122.09,114.21,80,140,80,184a47.8,47.8,0,0,0,2.73,15.93A71.88,71.88,0,0,1,56,144c0-34.41,20.4-63.15,37.52-81.19A216.21,216.21,0,0,1,128,33.54a215.77,215.77,0,0,1,34.48,29.27C193.49,95.5,200,125,200,144A71.88,71.88,0,0,1,173.27,199.93Z"/>',
  'event-log': '<path d="M80,64a8,8,0,0,1,8-8H216a8,8,0,0,1,0,16H88A8,8,0,0,1,80,64Zm136,56H88a8,8,0,0,0,0,16H216a8,8,0,0,0,0-16Zm0,64H88a8,8,0,0,0,0,16H216a8,8,0,0,0,0-16ZM44,52A12,12,0,1,0,56,64,12,12,0,0,0,44,52Zm0,64a12,12,0,1,0,12,12A12,12,0,0,0,44,116Zm0,64a12,12,0,1,0,12,12A12,12,0,0,0,44,180Z"/>',
  tasks: '<path d="M224,128a8,8,0,0,1-8,8H128a8,8,0,0,1,0-16h88A8,8,0,0,1,224,128ZM128,72h88a8,8,0,0,0,0-16H128a8,8,0,0,0,0,16Zm88,112H128a8,8,0,0,0,0,16h88a8,8,0,0,0,0-16ZM82.34,42.34,56,68.69,45.66,58.34A8,8,0,0,0,34.34,69.66l16,16a8,8,0,0,0,11.32,0l32-32A8,8,0,0,0,82.34,42.34Zm0,64L56,132.69,45.66,122.34a8,8,0,0,0-11.32,11.32l16,16a8,8,0,0,0,11.32,0l32-32a8,8,0,0,0-11.32-11.32Zm0,64L56,196.69,45.66,186.34a8,8,0,0,0-11.32,11.32l16,16a8,8,0,0,0,11.32,0l32-32a8,8,0,0,0-11.32-11.32Z"/>',
  'task-groups': '<path d="M230.91,172A8,8,0,0,1,228,182.91l-96,56a8,8,0,0,1-8.06,0l-96-56A8,8,0,0,1,36,169.09l92,53.65,92-53.65A8,8,0,0,1,230.91,172ZM220,121.09l-92,53.65L36,121.09A8,8,0,0,0,28,134.91l96,56a8,8,0,0,0,8.06,0l96-56A8,8,0,1,0,220,121.09ZM24,80a8,8,0,0,1,4-6.91l96-56a8,8,0,0,1,8.06,0l96,56a8,8,0,0,1,0,13.82l-96,56a8,8,0,0,1-8.06,0l-96-56A8,8,0,0,1,24,80Zm23.88,0L128,126.74,208.12,80,128,33.26Z"/>'
};

function icon(name, cls) {
  var el = document.createElementNS(SVG_NS, 'svg');
  el.setAttribute('xmlns', SVG_NS);
  var ph = PH_ICONS[name];
  el.setAttribute('class', 'shrink-0 x-icon inline shrink-0 ' + (ph ? '' : 'stroke-2 ') + (cls || 'h-4 w-4'));
  if (ph) {
    el.setAttribute('fill', 'currentColor');
    el.setAttribute('viewBox', '0 0 256 256');
  } else {
    el.setAttribute('stroke-linejoin', 'round');
    el.setAttribute('stroke-linecap', 'round');
    el.setAttribute('stroke', 'currentColor');
    el.setAttribute('fill', 'none');
    el.setAttribute('viewBox', '0 0 24 24');
  }
  el.setAttribute('aria-hidden', 'true');
  el.setAttribute('data-l10n-icon', MIRROR[name] ? 'mirror' : '');
  el.innerHTML = ph || ICONS[name] || '';
  return el;
}

function img(src, fallbackIcon) {
  var el = h('img', { src: src, alt: '', class: 'h-full w-full object-cover', loading: 'lazy' });
  el.addEventListener('error', function () { el.replaceWith(icon(fallbackIcon || 'image', 'h-4 w-4')); });
  return el;
}

function debounce(fn, ms) {
  var timer = null;
  return function () {
    var args = arguments;
    var self = this;
    clearTimeout(timer);
    timer = setTimeout(function () { fn.apply(self, args); }, ms);
  };
}

var BTN = 'x-button eco-btn relative transition duration-150 focus-visible:outline-secondary-300 outline-transparent outline outline-1 outline-offset-2 inline-flex items-center justify-center font-medium whitespace-nowrap overflow-hidden align-middle active:!shadow-none border appearance-none shrink-0 rounded-md';
var BTN_STYLE = {
  primary: '--x-button-text: #fff; --x-button-border: rgb(var(--primary-color-600)); --x-button-dark-text: #fff; --x-button-dark-border: rgb(var(--primary-color-500)); --x-button-bg: rgb(var(--primary-color-500)); --x-button-bg-hover: rgb(var(--primary-color-600)); --x-button-bg-active: rgb(var(--primary-color-800)); --x-button-dark-bg: rgb(var(--primary-color-600)); --x-button-dark-bg-hover: rgb(var(--primary-color-700)); --x-button-dark-bg-active: rgb(var(--primary-color-800));',
  ghost: '--x-button-border: transparent; --x-button-text: #1f2937; --x-button-dark-text: white; --x-button-bg-hover: #f3f4f6; --x-button-bg-active: #e5e7eb; --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  muted: '--x-button-border: transparent; --x-button-text: #475569; --x-button-dark-text: #64748b; --x-button-bg-hover: #f8fafc; --x-button-bg-active: #f1f5f9; --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  link: '--x-button-border: transparent; --x-button-text: rgb(var(--primary-color-600)); --x-button-dark-text: rgb(var(--primary-color-400)); --x-button-bg-hover: rgb(var(--primary-color-50)); --x-button-bg-active: rgb(var(--primary-color-100)); --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  danger: '--x-button-border: transparent; --x-button-text: rgb(var(--error-color-600)); --x-button-dark-text: rgb(var(--error-color-400)); --x-button-bg-hover: rgb(var(--error-color-50)); --x-button-bg-active: rgb(var(--error-color-100)); --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  dangerSolid: '--x-button-text: #fff; --x-button-border: rgb(var(--error-color-600)); --x-button-bg: rgb(var(--error-color-600)); --x-button-bg-hover: rgb(var(--error-color-700)); --x-button-bg-active: rgb(var(--error-color-800)); --x-button-dark-text: #fff; --x-button-dark-border: rgb(var(--error-color-600)); --x-button-dark-bg: rgb(var(--error-color-600)); --x-button-dark-bg-hover: rgb(var(--error-color-500));',
  disabled: '--x-button-bg: #f9fafb; --x-button-text: #d1d5db; --x-button-border: #e5e7eb; --x-button-dark-bg: #111827; --x-button-dark-text: #4b5563; --x-button-dark-border: #374151;'
};
var BTN_SIZE = {
  page: 'shadow-sm py-1.5 text-sm px-3',
  footer: 'py-2 px-4',
  sm: 'py-1.5 text-sm px-3',
  icon: 'leading-none p-1',
  iconLg: 'leading-none p-1.5'
};

function button(opts) {
  var variant = opts.disabled ? 'disabled' : (opts.variant || 'ghost');
  var size = BTN_SIZE[opts.size || 'footer'];
  if (variant === 'primary' && opts.size === 'footer') size += ' shadow-sm';
  var b = h('button', {
    type: opts.type || 'button',
    class: BTN + ' ' + size + ' ' + (opts.disabled ? 'cursor-not-allowed' : 'cursor-pointer') + (opts.class ? ' ' + opts.class : ''),
    style: BTN_STYLE[variant],
    disabled: !!opts.disabled,
    title: opts.title,
    'aria-label': opts.title,
    onclick: opts.onclick
  });
  var inner = h('div', { class: 'inline-flex items-center justify-center' });
  if (opts.icon) inner.appendChild(icon(opts.icon, (opts.iconClass || 'h-4 w-4') + (opts.label ? ' mr-2' : ' m-0.5')));
  if (opts.label) inner.appendChild(h('span', { text: opts.label }));
  b.appendChild(inner);
  return b;
}

function setBusy(btn, busy, label) {
  btn.disabled = !!busy;
  var span = btn.querySelector('span');
  if (busy) {
    btn.dataset.label = span ? span.textContent : '';
    if (span && label) span.textContent = label;
    btn.classList.add('opacity-60');
  } else {
    if (span && btn.dataset.label) span.textContent = btn.dataset.label;
    btn.classList.remove('opacity-60');
  }
}

var INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 outline-transparent outline outline-2 outline-offset-[-1px] transition-all duration-150 ease-in-out border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md hover:border-secondary-400 dark:hover:border-secondary-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 focus:outline-[color:var(--x-input-border)]';
var INPUT_SM = INPUT + ' px-2 py-1.5 text-sm';
var INPUT_MD = INPUT + ' px-3 py-2';
var INPUT_VARS = '--x-input-border: rgb(var(--primary-color-400));';
var SELECT_MD = INPUT_MD + ' eco-select';
var SELECT_SM = INPUT_SM + ' eco-select';

function searchInput(value, placeholder, oninput) {
  var input = h('input', {
    class: INPUT_SM + ' !pl-10',
    type: 'text',
    value: value || '',
    placeholder: placeholder,
    'aria-label': placeholder
  });
  input.addEventListener('input', oninput);
  return h('label', { class: 'x-label block relative align-bottom focus:outline-none x-input eco-search', style: INPUT_VARS },
    h('div', { class: 'relative' },
      icon('search', 'h-4 w-4 ml-2 left-1 text-secondary-600 dark:text-secondary-300 absolute my-auto inset-y-0'),
      input));
}

function filterSelect(label, value, options, onchange) {
  var sel = h('select', { class: SELECT_SM, 'aria-label': label, onchange: function () { onchange(sel.value); } });
  sel.appendChild(h('option', { value: '', text: label + ': ' + t('all') }));
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value || '')) opt.selected = true;
    sel.appendChild(opt);
  });
  return h('label', { class: 'x-label block relative x-input', style: INPUT_VARS }, sel);
}

function toggle(checked, label, onchange, size) {
  var on = !!checked;
  var track = h('div', { class: '', style: '--x-toggle-bg: rgb(var(--primary-color-500)); --x-toggle-dark-bg: rgb(var(--primary-color-600));' });
  var knob = h('div', { class: '' });
  var input = h('input', { type: 'checkbox', class: 'hidden', role: 'switch' });
  var sm = size === 'sm';
  function paint() {
    track.className = 'flex items-center rounded-full transition-colors duration-300 border-[3px] shrink-0 !border-transparent ' + (on ? 'bg-[color:var(--x-toggle-bg)] dark:bg-[color:var(--x-toggle-dark-bg)]' : 'bg-secondary-300 dark:bg-secondary-500');
    knob.className = 'rounded-full shadow transform transition duration-150 shrink-0 ' + (sm ? 'h-3 w-3' : 'h-4 w-4') + ' bg-white dark:bg-secondary-800' + (on ? ' translate-x-full' : '');
    input.checked = on;
    input.setAttribute('aria-checked', on ? 'true' : 'false');
  }
  track.appendChild(h('div', { class: 'relative shrink-0 ' + (sm ? 'w-6' : 'w-8') }, input, knob));
  var wrap = h('label', { class: 'inline-block x-toggle cursor-pointer', tabindex: '0' },
    h('div', { class: 'flex items-center' }, track,
      label ? h('span', { class: 'ml-2 font-medium text-secondary-800 dark:text-secondary-200 ' + (sm ? 'text-xs' : 'text-sm'), text: label }) : null));
  function flip(e) {
    e.preventDefault();
    on = !on;
    paint();
    onchange(on);
  }
  wrap.addEventListener('click', flip);
  wrap.addEventListener('keydown', function (e) { if (e.key === ' ' || e.key === 'Enter') flip(e); });
  wrap.set = function (v) { on = !!v; paint(); };
  paint();
  return wrap;
}

function checkbox(checked, label, onchange) {
  var on = !!checked;
  var box = h('div', {});
  var mark = document.createElementNS(SVG_NS, 'svg');
  mark.setAttribute('viewBox', '0 0 20 20');
  mark.innerHTML = '<path d="M0 11l2-2 5 5L18 3l2 2L7 18z"/>';
  box.appendChild(mark);
  var wrap = h('label', { class: 'x-checkbox inline-block relative cursor-pointer align-middle', tabindex: '0', role: 'checkbox' });
  function paint() {
    wrap.setAttribute('style', on
      ? '--x-checkbox-bg: rgb(var(--primary-color-500)); --x-checkbox-border: rgb(var(--primary-color-500)); --x-checkbox-dark-bg: rgb(var(--primary-color-500)); --x-checkbox-dark-border: rgb(var(--primary-color-500));'
      : '--x-checkbox-bg: #fff; --x-checkbox-border: #d1d5db; --x-checkbox-dark-bg: #111827; --x-checkbox-dark-border: #9ca3af;');
    box.setAttribute('class', 'rounded flex justify-center items-center shrink-0 border border-[color:var(--x-checkbox-border)] bg-[color:var(--x-checkbox-bg)] dark:border-[color:var(--x-checkbox-dark-border)] dark:bg-[color:var(--x-checkbox-dark-bg)] h-4 w-4');
    mark.setAttribute('class', 'fill-current text-white h-2.5 w-2.5' + (on ? '' : ' opacity-0'));
    wrap.setAttribute('aria-checked', on ? 'true' : 'false');
  }
  wrap.appendChild(h('div', { class: 'flex items-center' }, box,
    h('div', { class: 'inline-block text-secondary-800 dark:text-secondary-200 pl-2 font-medium' }, h('span', { text: label }))));
  function flip(e) { e.preventDefault(); on = !on; paint(); onchange(on); }
  wrap.addEventListener('click', flip);
  wrap.addEventListener('keydown', function (e) { if (e.key === ' ' || e.key === 'Enter') flip(e); });
  paint();
  return wrap;
}

function tag(text, color, extra) {
  return h('span', { class: 'eco-tag eco-tag--' + (color || 'secondary') + ' inline-flex leading-tight max-w-full px-2 py-0.5 text-xs rounded-md whitespace-nowrap' + (extra ? ' ' + extra : ''), text: text });
}

function mono(text) {
  return h('span', { class: 'font-mono text-xs', dir: 'ltr', text: text == null ? '' : String(text) });
}

function bar(width, extra) {
  return h('div', { class: 'animate-pulse bg-secondary-300 dark:bg-secondary-600 rounded-md eco-bar ' + (extra || 'h-3'), style: 'width:' + width });
}

function pageFrame(title, desc, actions) {
  var content = h('div', { class: 'flex flex-col flex-1 min-w-0 px-4 md:px-6' });
  var actionsBox = h('div', { class: 'flex items-center gap-1.5 justify-end w-full sm:w-auto' }, actions || null);
  var frame = h('div', { class: 'flex flex-col flex-1' },
    h('div', { class: 'w-full min-w-0 flex flex-col flex-1 pb-4 max-w-[1600px] mx-auto overflow-x-auto' },
      h('div', { class: 'mb-4 relative px-4 md:px-6' },
        h('div', { class: 'flex flex-col sm:flex-row gap-4 justify-between items-start' },
          h('div', { class: 'flex items-center gap-1' },
            h('div', { class: 'flex gap-4 items-center' },
              h('div', {},
                h('div', { class: 'flex items-center gap-2 mb-0.5' },
                  h('h1', { class: 'text-xl font-semibold truncate max-w-xs md:max-w-sm xl:max-w-xl mr-2' }, h('span', { text: title }))),
                desc ? h('div', { class: 'flex items-center gap-2' },
                  h('div', { class: 'text-secondary-500 -mt-0.5 space-x-2' }, h('span', { text: desc }))) : null))),
          actionsBox)),
      content));
  return { el: frame, content: content, actions: actionsBox };
}

function tabFrame(desc, actions) {
  var content = h('div', { class: 'flex flex-col flex-1 min-w-0' });
  var actionsBox = h('div', { class: 'flex flex-wrap items-center gap-2' }, actions || null);
  var frame = h('div', { class: 'flex flex-col flex-1 min-w-0' },
    h('div', { class: 'flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4' },
      h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: desc }),
      actionsBox),
    content);
  return { el: frame, content: content, actions: actionsBox };
}

function notice(kind, text, action) {
  var color = kind === 'error' ? 'error' : kind === 'success' ? 'success' : kind === 'warning' ? 'warning' : 'secondary';
  return h('div', { class: 'eco-notice eco-notice--' + color + ' flex items-center gap-2 rounded-md px-3 py-2 text-sm', role: kind === 'error' ? 'alert' : 'status' },
    icon(kind === 'success' ? 'check' : 'alert', 'h-4 w-4'),
    h('span', { class: 'flex-1', text: text }),
    action || null);
}

function emptyBlock(title, hint, action) {
  return h('div', { class: 'flex flex-col items-center justify-center text-center gap-2 py-12 px-4' },
    h('div', { class: 'text-secondary-400 dark:text-secondary-500' }, icon('empty', 'h-8 w-8')),
    h('div', { class: 'font-medium text-secondary-800 dark:text-secondary-200', text: title }),
    hint ? h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: hint }) : null,
    action ? h('div', { class: 'mt-2' }, action) : null);
}

var toastBox = null;
function toast(kind, text) {
  if (!toastBox || !toastBox.isConnected) {
    toastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-economy': '' });
    document.body.appendChild(toastBox);
  }
  toastBox.setAttribute('dir', document.documentElement.getAttribute('dir') || 'ltr');
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  toastBox.appendChild(el);
  setTimeout(function () { el.remove(); }, kind === 'error' ? 6000 : 3000);
}

function card(children, extra) {
  return h('div', { class: 'bg-white dark:bg-secondary-800 overflow-hidden rounded-md shadow' + (extra ? ' ' + extra : '') }, children);
}

function pager(state, onpage, onrefresh) {
  var total = state.total || 0;
  var from = total ? (state.page - 1) * state.limit + 1 : 0;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  var info = h('div', { class: 'text-xs text-secondary-500 shrink-0' });
  if (state.loading && !state.loaded) info.appendChild(bar('5rem'));
  else if (!total) info.appendChild(h('div', { text: t('range.none') }));
  else info.appendChild(h('div', { class: 'flex items-center gap-1' }, h('span', { class: 'tabular-nums', text: t('range', { from: from, to: to, total: total }) })));
  return h('div', { class: 'flex items-center gap-2 justify-end shrink-0' },
    info,
    onrefresh ? button({ variant: 'muted', size: 'icon', icon: 'refresh', iconClass: 'h-3 w-3', title: t('refresh'), onclick: onrefresh }) : null,
    button({ variant: 'muted', size: 'icon', icon: 'first', iconClass: 'h-3 w-3', title: t('page.first'), disabled: state.page <= 1, onclick: function () { onpage(1); } }),
    button({ variant: 'muted', size: 'icon', icon: 'prev', iconClass: 'h-3 w-3', title: t('page.prev'), disabled: state.page <= 1, onclick: function () { onpage(state.page - 1); } }),
    button({ variant: 'muted', size: 'icon', icon: 'next', iconClass: 'h-3 w-3', title: t('page.next'), disabled: state.page >= last, onclick: function () { onpage(state.page + 1); } }));
}

function table(columns, rows, opts) {
  opts = opts || {};
  var thead = h('thead', { class: 'x-table-head align-bottom bg-secondary-100 dark:bg-secondary-700 sticky top-0 z-10' });
  var tr = h('tr', { class: 'text-sm text-secondary-700 dark:text-secondary-200 border-b' });
  columns.forEach(function (c) {
    tr.appendChild(h('th', { class: 'x-table-header relative p-2.5 font-medium', width: c.width },
      h('div', { class: 'flex items-center gap-1 select-none ' + (c.align === 'center' ? 'justify-center' : 'justify-start') }, h('div', { class: 'flex items-center gap-1', text: c.label }))));
  });
  thead.appendChild(tr);
  var tbody = h('tbody', {});
  rows.forEach(function (row) {
    var r = h('tr', { class: 'x-table-row border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700 align-top' + (opts.onrow ? ' cursor-pointer' : '') });
    if (opts.onrow) {
      r.tabIndex = 0;
      r.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,label')) opts.onrow(row.item); });
      r.addEventListener('keydown', function (e) { if (e.key === 'Enter') opts.onrow(row.item); });
    }
    columns.forEach(function (c, i) {
      var cell = row.cells[i];
      r.appendChild(h('td', { class: 'relative x-table-cell px-3 h-9 py-1.5 align-middle' + (c.align === 'center' ? ' text-center' : ''), 'data-label': c.label }, cell));
    });
    tbody.appendChild(r);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow border-t rounded-md overflow-hidden' },
    h('div', { class: 'x-table w-full h-full overflow-x-auto' },
      h('div', { class: 'relative' },
        h('table', { class: 'w-full relative eco-table' }, thead, tbody)),
      opts.after || null));
}

function skeletonRows(columns, n) {
  var widths = ['70%', '45%', '60%', '35%', '50%'];
  var rows = [];
  for (var i = 0; i < n; i++) {
    rows.push({ item: null, cells: columns.map(function (c, j) { return bar(c.skeleton || widths[(i + j) % widths.length]); }) });
  }
  return rows;
}

function field(label, control, hint, opts) {
  opts = opts || {};
  var err = h('p', { class: 'text-error-600 dark:text-error-400', hidden: true });
  var wrap = h(opts.tag || 'div', { class: 'x-label block relative align-bottom focus:outline-none mb-5 x-input eco-field', style: INPUT_VARS },
    label ? h('p', { class: 'flex items-center gap-2 font-medium text-secondary-800 dark:text-secondary-200 mb-1' },
      h('span', { text: label }),
      opts.optional ? h('span', { class: 'text-xs font-normal text-secondary-500', text: '(' + t('optional') + ')' }) : null) : null,
    h('div', { class: 'relative' }, control),
    h('div', { class: 'x-input-footer text-xs mt-1' },
      hint ? h('p', { class: 'text-secondary-500 dark:text-secondary-400', text: hint }) : null,
      err));
  wrap.setError = function (msg) {
    err.textContent = msg || '';
    err.hidden = !msg;
    wrap.classList.toggle('eco-invalid', !!msg);
  };
  return wrap;
}

function sidePanel(title, body, footer, onclose) {
  var shade = h('div', { class: 'fixed z-[60] inset-0 opacity-60 dark:opacity-70 dark:bg-black bg-secondary-900 w-full h-full transition duration-200' });
  var panel = h('div', { class: 'x-form relative flex flex-col z-10 dark:bg-secondary-900 rounded-l-md shadow-lg overflow-hidden transition-all duration-300 h-full w-full bg-white sm:max-w-3xl', role: 'dialog', 'aria-modal': 'true', 'aria-label': title },
    h('div', { class: 'text-lg font-semibold' }, h('div', { class: 'pl-6 py-[0.69rem] border-b pr-12 truncate', text: title })),
    h('div', { class: 'relative overflow-y-auto w-full flex-1' }, h('div', { class: 'p-6' }, body)),
    h('div', { class: '!absolute flex items-center top-2 z-10 right-2' },
      button({ variant: 'ghost', size: 'iconLg', icon: 'close', title: t('close'), onclick: onclose })),
    footer ? h('div', { class: 'flex gap-4 bg-secondary-50 dark:bg-secondary-800 p-4 justify-end items-center' }, footer) : null);
  var holder = h('div', { class: 'fixed z-[60] inset-0 overflow-y-auto flex justify-end h-screen overflow-hidden transition' }, panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) onclose(); });
  var root = h('div', { 'data-economy': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, shade, holder);
  return root;
}

function segmented(options, value, onchange, label, size) {
  var wrap = h('div', { class: 'eco-seg inline-flex flex-wrap items-center gap-0.5 rounded-md p-0.5', role: 'radiogroup', 'aria-label': label || null });
  var btns = options.map(function (o) {
    return h('button', { type: 'button', role: 'radio', class: 'eco-seg-btn inline-flex items-center gap-1.5 rounded font-medium ' + (size === 'md' ? 'px-3 py-1.5 text-sm' : 'px-2 py-1 text-xs'), 'data-value': o.value,
      onclick: function () { if (String(value) === String(o.value)) return; pick(o.value); onchange(o.value); } },
      o.icon ? icon(o.icon, 'h-4 w-4') : null, h('span', { text: o.label }));
  });
  function pick(v) {
    value = v;
    btns.forEach(function (b) {
      var on = b.getAttribute('data-value') === String(v);
      b.setAttribute('aria-checked', on ? 'true' : 'false');
      b.classList.toggle('eco-seg-on', on);
    });
  }
  btns.forEach(function (b) { wrap.appendChild(b); });
  pick(value);
  wrap.set = pick;
  return wrap;
}

function confirmModal(opts) {
  return new Promise(function (resolve) {
    var done = false;
    function finish(v) {
      if (done) return;
      done = true;
      root.remove();
      document.removeEventListener('keydown', esc, true);
      resolve(v);
    }
    function esc(e) { if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); finish(false); } }
    var ok = button({ variant: opts.danger ? 'dangerSolid' : 'primary', size: 'footer', label: opts.ok || t('yes'), onclick: function () { finish(true); } });
    var box = h('div', { class: 'relative w-full eco-dialog bg-white dark:bg-secondary-900 rounded-md shadow-lg overflow-hidden', role: 'alertdialog', 'aria-modal': 'true', 'aria-label': opts.title },
      h('div', { class: 'p-6 flex gap-3' },
        h('div', { class: 'shrink-0 ' + (opts.danger ? 'text-error-600 dark:text-error-400' : 'eco-selected') }, icon(opts.icon || 'alert', 'h-6 w-6')),
        h('div', { class: 'min-w-0' },
          h('h2', { class: 'text-lg font-semibold mb-1', text: opts.title }),
          opts.text ? h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: opts.text }) : null,
          opts.body || null)),
      h('div', { class: 'flex gap-3 justify-end bg-secondary-50 dark:bg-secondary-800 px-6 py-3' },
        button({ variant: 'ghost', size: 'footer', label: opts.cancel || t('cancel'), onclick: function () { finish(false); } }), ok));
    var holder = h('div', { class: 'fixed z-[70] inset-0 flex items-center justify-center p-4' }, box);
    holder.addEventListener('mousedown', function (e) { if (e.target === holder) finish(false); });
    var root = h('div', { 'data-economy': 'dialog', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'fixed z-[70] inset-0 opacity-60 dark:opacity-70 dark:bg-black bg-secondary-900' }), holder);
    document.body.appendChild(root);
    document.addEventListener('keydown', esc, true);
    ok.focus();
  });
}

var KEY_RE = /^[a-z0-9][a-z0-9_.-]{0,63}$/;
var CURRENCY_KEY_RE = /^[a-z0-9_]{2,32}$/;
var CODE_RE = /^[A-Z0-9_-]{3,32}$/;
var refCache = {};

function refLabel(item) {
  if (!item) return '';
  var name = i18nText(item.name) || i18nText(item.title) || i18nText(item.question);
  return name || item.key || item.username || item.id || '';
}

function textInput(value, opts) {
  opts = opts || {};
  var el = h(opts.multiline ? 'textarea' : 'input', {
    class: INPUT_MD + (opts.mono ? ' font-mono text-sm' : '') + (opts.multiline ? ' eco-textarea' : ''),
    type: opts.multiline ? null : 'text',
    dir: opts.ltr ? 'ltr' : 'auto',
    'data-ltr': opts.ltr ? '' : null,
    'data-ph-rtl': opts.ltr && /[\u0600-\u06FF]/.test(opts.placeholder || '') ? '' : null,
    rows: opts.multiline ? (opts.rows || 4) : null,
    inputmode: opts.inputmode,
    placeholder: opts.placeholder,
    autocomplete: 'off',
    spellcheck: opts.ltr ? 'false' : null,
    'aria-label': opts.label
  });
  el.value = value == null ? '' : value;
  if (opts.ltr) el.addEventListener('input', function () { latinize(el); });
  if (opts.oninput) el.addEventListener('input', function () { opts.oninput(el.value); });
  return el;
}

function latinize(el) {
  if (!/[۰-۹٠-٩]/.test(el.value)) return;
  var pos = el.selectionStart;
  el.value = el.value.replace(/[۰-۹]/g, function (d) { return String(d.charCodeAt(0) - 0x06F0); }).replace(/[٠-٩]/g, function (d) { return String(d.charCodeAt(0) - 0x0660); });
  try { el.setSelectionRange(pos, pos); } catch (e) { return; }
}

function numberInput(value, opts) {
  var shown = value === null || value === undefined || value === '' ? '' : localDigits(String(value));
  var el = textInput(shown, { inputmode: opts.int ? 'numeric' : 'decimal', label: opts.label, placeholder: opts.placeholder, oninput: function () {
    var raw = toLatinDigits(el.value).replace(/,/g, '').trim();
    opts.oninput(raw);
  } });
  el.classList.add('tabular-nums');
  el.addEventListener('blur', function () {
    var raw = toLatinDigits(el.value).replace(/,/g, '').trim();
    if (raw !== '' && isFinite(Number(raw))) el.value = localDigits(raw);
  });
  return el;
}

function parseNumber(raw, int) {
  if (raw === null || raw === undefined || raw === '') return null;
  var s = toLatinDigits(String(raw)).replace(/,/g, '').trim();
  if (s === '') return null;
  if (int ? !/^-?\d+$/.test(s) : !/^-?\d+(\.\d+)?$/.test(s)) return NaN;
  return Number(s);
}

function selectInput(value, options, onchange, opts) {
  opts = opts || {};
  var sel = h('select', { class: SELECT_MD, 'aria-label': opts.label });
  if (opts.placeholder !== false) sel.appendChild(h('option', { value: '', text: opts.placeholder || t('choose') }));
  options.forEach(function (o) {
    var op = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) op.selected = true;
    sel.appendChild(op);
  });
  sel.addEventListener('change', function () { onchange(sel.value); });
  return sel;
}

function refPicker(opts) {
  var call = opts.api || api;
  var value = opts.value == null ? '' : opts.value;
  var cacheKey = function (v) { return opts.resource + ':' + (opts.valueKey || 'id') + ':' + v; };
  var boxLabel = h('span', { class: 'flex items-center gap-2 min-w-0 truncate' });
  var box = h('button', { type: 'button', class: SELECT_MD + ' eco-start flex items-center gap-2 cursor-pointer', 'aria-haspopup': 'listbox', 'aria-label': opts.label });
  box.appendChild(boxLabel);
  var clearBtn = opts.required ? null : button({ variant: 'muted', size: 'icon', icon: 'close', iconClass: 'h-3 w-3', title: t('remove'), class: 'eco-ref-clear', onclick: function (e) { e.stopPropagation(); set('', null); } });
  var list = h('div', { class: 'overflow-y-auto max-h-64', role: 'listbox' });
  var search = h('input', { class: INPUT_SM, type: 'text', placeholder: t('search.short'), 'aria-label': t('search.short') });
  var pop = h('div', { class: 'eco-pop absolute z-30 w-full mt-1 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-md shadow-lg p-1', hidden: true },
    h('div', { class: 'p-1 mb-0.5' }, search), list);
  var wrap = h('div', { class: 'relative eco-ref' }, box, clearBtn, pop);
  var seq = 0;

  function paint(item) {
    clear(boxLabel);
    if (!value) {
      boxLabel.appendChild(h('span', { class: 'text-secondary-400', text: opts.placeholder || t('choose') }));
    } else if (item) {
      boxLabel.appendChild(h('span', { text: refLabel(item) }));
      if (item.key && refLabel(item) !== item.key) boxLabel.appendChild(h('span', { class: 'font-mono text-xs text-secondary-500', dir: 'ltr', text: item.key }));
    } else {
      boxLabel.appendChild(h('span', { class: 'font-mono text-xs', dir: 'ltr', text: value }));
    }
    if (clearBtn) clearBtn.hidden = !value;
  }

  function set(v, item) {
    value = v || '';
    if (item) refCache[cacheKey(value)] = item;
    paint(item || refCache[cacheKey(value)]);
    opts.onchange(value || null, item || null);
  }

  function load(q) {
    var my = ++seq;
    clear(list);
    for (var i = 0; i < 3; i++) list.appendChild(h('div', { class: 'px-2 py-2' }, bar(['60%', '40%', '50%'][i])));
    var filters = Array.isArray(opts.filter) ? opts.filter : [opts.filter || {}];
    Promise.all(filters.map(function (f) { return call('GET', '/' + opts.resource + qs(Object.assign({ q: q, page: 1, limit: 20 }, f))); })).then(function (all) {
      if (my !== seq) return;
      clear(list);
      var seen = {};
      var items = [];
      all.forEach(function (res) { ((res && res.items) || []).forEach(function (it) { var k = it[opts.valueKey || 'id']; if (!seen[k]) { seen[k] = 1; items.push(it); } }); });
      if (!items.length) list.appendChild(h('div', { class: 'px-2 py-2 text-sm text-secondary-500', text: !q && opts.emptyText ? opts.emptyText : t('no.results') }));
      items.forEach(function (it) {
        var v = it[opts.valueKey || 'id'];
        refCache[cacheKey(v)] = it;
        var row = h('button', { type: 'button', role: 'option', 'aria-selected': String(v) === String(value) ? 'true' : 'false',
          class: 'w-full eco-start px-2 py-1.5 rounded-md hover:bg-secondary-100 dark:hover:bg-secondary-700 flex items-center justify-between gap-2 text-sm' + (String(v) === String(value) ? ' eco-selected font-medium' : ''),
          onclick: function () { close(); set(v, it); } },
          h('span', { class: 'truncate', text: refLabel(it) }),
          it.key ? h('span', { class: 'font-mono text-xs text-secondary-500', dir: 'ltr', text: it.key }) : null);
        list.appendChild(row);
      });
    }, function (err) {
      if (my !== seq) return;
      clear(list);
      list.appendChild(h('div', { class: 'px-2 py-2 text-sm text-error-600 dark:text-error-400', text: errorText(err) }));
    });
  }

  function outside(e) { if (!wrap.contains(e.target)) close(); }
  function open() {
    pop.hidden = false;
    search.value = '';
    load('');
    search.focus();
    document.addEventListener('mousedown', outside, true);
  }
  function close() {
    pop.hidden = true;
    document.removeEventListener('mousedown', outside, true);
  }
  box.addEventListener('click', function () { if (pop.hidden) open(); else close(); });
  search.addEventListener('input', debounce(function () { load(toLatinDigits(search.value).trim()); }, 250));
  wrap.addEventListener('keydown', function (e) { if (e.key === 'Escape' && !pop.hidden) { e.stopPropagation(); close(); box.focus(); } });

  wrap.select = function (item) { set(item[opts.valueKey || 'id'], item); };
  paint(refCache[cacheKey(value)]);
  if (value && !refCache[cacheKey(value)]) {
    var req = (opts.valueKey || 'id') === 'id'
      ? call('GET', '/' + opts.resource + '/' + encodeURIComponent(value))
      : call('GET', '/' + opts.resource + qs(Object.assign({ q: value, limit: 20 }, Array.isArray(opts.filter) ? {} : opts.filter || {}))).then(function (res) {
        return ((res && res.items) || []).filter(function (it) { return it[opts.valueKey] === value; })[0] || null;
      });
    req.then(function (item) { if (item) { refCache[cacheKey(value)] = item; paint(item); } }, function () { return null; });
  }
  return wrap;
}

function refMulti(opts) {
  var values = (opts.value || []).slice();
  var chips = h('div', { class: 'flex flex-wrap gap-1.5 mb-2' });
  function paint() {
    clear(chips);
    chips.hidden = !values.length;
    values.forEach(function (v) {
      var item = refCache[opts.resource + ':id:' + v];
      chips.appendChild(h('span', { class: 'eco-tag eco-tag--primary inline-flex items-center gap-1 leading-tight px-2 py-0.5 text-xs rounded-md' },
        h('span', { text: item ? refLabel(item) : v }),
        h('button', { type: 'button', class: 'eco-chip-x', 'aria-label': t('remove'), onclick: function () {
          values = values.filter(function (x) { return x !== v; });
          paint();
          opts.onchange(values.slice());
        } }, icon('close', 'h-3 w-3'))));
    });
  }
  var call = opts.api || api;
  var picker = refPicker({ resource: opts.resource, api: opts.api, emptyText: opts.emptyText, filter: opts.filter, label: opts.label, placeholder: opts.placeholder || t('choose'), value: '', onchange: function (v) {
    if (v && values.indexOf(v) === -1) {
      values.push(v);
      paint();
      opts.onchange(values.slice());
    }
    picker.querySelector('.eco-ref-clear') && picker.querySelector('.eco-ref-clear').click();
  } });
  function add(item) {
    refCache[opts.resource + ':id:' + item.id] = item;
    if (values.indexOf(item.id) !== -1) return;
    values.push(item.id);
    paint();
    opts.onchange(values.slice());
  }
  var missing = values.filter(function (v) { return !refCache[opts.resource + ':id:' + v]; });
  missing.forEach(function (v) {
    call('GET', '/' + opts.resource + '/' + encodeURIComponent(v)).then(function (it) {
      if (!it) return;
      refCache[opts.resource + ':id:' + v] = it;
      paint();
    }, function () { return null; });
  });
  paint();
  var out = h('div', {}, chips, opts.quick ? h('div', { class: 'eco-ref-row' }, picker, quickCreate({ resource: opts.resource, keyType: opts.quick, onCreated: add })) : picker);
  out.select = add;
  return out;
}

var FA_LATIN = { 'ا': 'a', 'آ': 'a', 'أ': 'a', 'إ': 'e', 'ب': 'b', 'پ': 'p', 'ت': 't', 'ث': 's', 'ج': 'j', 'چ': 'ch', 'ح': 'h', 'خ': 'kh', 'د': 'd', 'ذ': 'z', 'ر': 'r', 'ز': 'z', 'ژ': 'zh', 'س': 's', 'ش': 'sh', 'ص': 's', 'ض': 'z', 'ط': 't', 'ظ': 'z', 'ع': 'a', 'غ': 'gh', 'ف': 'f', 'ق': 'gh', 'ک': 'k', 'ك': 'k', 'گ': 'g', 'ل': 'l', 'م': 'm', 'ن': 'n', 'و': 'v', 'ؤ': 'v', 'ه': 'h', 'ة': 'h', 'ی': 'i', 'ي': 'i', 'ئ': 'i', 'ى': 'i', 'ء': '' };

function slugKey(text, prefix) {
  var s = toLatinDigits(text).toLowerCase().split('').map(function (c) { return FA_LATIN[c] !== undefined ? FA_LATIN[c] : c; }).join('');
  if (s.normalize) s = s.normalize('NFKD').replace(/[\u0300-\u036f]/g, '');
  s = s.replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '').slice(0, 48).replace(/-+$/, '');
  return s || prefix + '-' + Date.now().toString(36).slice(-5);
}

function quickCreate(opts) {
  var title = t('qc.add.' + opts.resource);
  var btn = button({ variant: 'ghost', size: 'iconLg', icon: 'plus', title: title, class: 'eco-qc-btn border-secondary-300 dark:border-secondary-700' });
  btn.setAttribute('aria-haspopup', 'dialog');
  btn.setAttribute('aria-expanded', 'false');
  btn.setAttribute('data-quick-create', opts.resource);
  var input = textInput('', { label: t('qc.name'), placeholder: t('qc.name.ph.' + opts.resource) });
  input.classList.add('text-sm');
  var hint = h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1' });
  var err = h('p', { class: 'eco-qc-err text-xs text-error-600 dark:text-error-400 mt-1', 'aria-live': 'assertive', hidden: true });
  var create = button({ variant: 'primary', size: 'sm', label: t('qc.create'), onclick: function () { submit(); } });
  create.setAttribute('data-qc-create', '');
  var cancel = button({ variant: 'ghost', size: 'sm', label: t('cancel'), onclick: function () { close(true); } });
  var pop = h('div', { class: 'eco-qc-pop bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-md shadow-lg p-3', role: 'dialog', 'aria-label': title, hidden: true },
    h('div', { class: 'text-sm font-medium text-secondary-800 dark:text-secondary-100 mb-2', text: title }),
    h('label', { class: 'x-label block x-input', style: INPUT_VARS }, input),
    hint, err,
    h('div', { class: 'flex justify-end gap-2 mt-3' }, cancel, create));
  var wrap = h('div', { class: 'eco-qc relative shrink-0' }, btn, pop);
  var busy = false;
  function showErr(msg) {
    err.textContent = msg || '';
    err.hidden = !msg;
    input.setAttribute('aria-invalid', msg ? 'true' : 'false');
  }
  function outside(e) { if (!wrap.contains(e.target)) close(false); }
  function open() {
    pop.hidden = false;
    btn.setAttribute('aria-expanded', 'true');
    input.value = '';
    hint.textContent = t('qc.hint', { type: keyTypeLabel(opts.resource, opts.keyType), lang: langName(LANGS.def) });
    input.setAttribute('dir', langDir(LANGS.def));
    showErr('');
    input.focus();
    document.addEventListener('mousedown', outside, true);
  }
  function close(refocus) {
    if (pop.hidden) return;
    pop.hidden = true;
    btn.setAttribute('aria-expanded', 'false');
    document.removeEventListener('mousedown', outside, true);
    if (refocus) btn.focus();
  }
  function submit() {
    if (busy) return;
    var text = input.value.trim();
    if (!text) { showErr(t('qc.err.name')); input.focus(); return; }
    showErr('');
    busy = true;
    setBusy(create, true, t('saving'));
    attempt(slugKey(text, opts.keyType), 1, text);
  }
  function attempt(base, n, text) {
    var name = {};
    name[LANGS.def] = text;
    var key = n === 1 ? base : base + '-' + n;
    baseApi('POST', '/' + opts.resource, { keyType: opts.keyType, key: key, name: name, description: {}, iconUrl: '', metadata: {}, sort: 0 }).then(function (item) {
      busy = false;
      setBusy(create, false);
      close(true);
      opts.onCreated(item);
      toast('success', t('qc.created', { name: text }));
    }, function (e) {
      if (e.code === 'ALREADY_EXISTS' && n < 9) return attempt(base, n + 1, text);
      busy = false;
      setBusy(create, false);
      showErr(errorText(e));
      input.focus();
    });
  }
  btn.addEventListener('click', function () { if (pop.hidden) open(); else close(true); });
  input.addEventListener('keydown', function (e) { if (e.key === 'Enter') { e.preventDefault(); submit(); } });
  input.addEventListener('input', function () { if (!err.hidden) showErr(''); });
  wrap.addEventListener('keydown', function (e) { if (e.key === 'Escape' && !pop.hidden) { e.preventDefault(); e.stopPropagation(); close(true); } });
  return wrap;
}

function imageInput(value, onchange) {
  var url = value || '';
  var preview = h('div', { class: 'h-16 w-16 shrink-0 rounded-md border border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-800 flex items-center justify-center overflow-hidden text-secondary-400' });
  var urlInput = textInput(url, { ltr: true, label: t('f.uploadUrl'), placeholder: 'https://', oninput: function (v) { url = v.trim(); paint(); onchange(url || null); } });
  urlInput.classList.add('text-sm');
  var file = h('input', { type: 'file', accept: 'image/png,image/jpeg,image/webp,image/svg+xml', class: 'hidden' });
  var err = h('p', { class: 'text-xs text-error-600 dark:text-error-400 mt-1', hidden: true });
  var up = button({ variant: 'link', size: 'sm', icon: 'upload', label: t('upload'), onclick: function () { file.click(); } });
  function paint() {
    clear(preview);
    if (url) preview.appendChild(img(url));
    else preview.appendChild(icon('image', 'h-6 w-6'));
  }
  file.addEventListener('change', function () {
    var f = file.files && file.files[0];
    if (!f) return;
    err.hidden = true;
    var fd = new FormData();
    fd.append('file', f, f.name);
    setBusy(up, true, t('uploading'));
    api('POST', '/uploads', fd, true).then(function (res) {
      url = (res && res.url) || '';
      urlInput.value = url;
      paint();
      onchange(url || null);
    }, function (e) {
      err.textContent = e && e.status === 400 ? t('err.upload') : errorText(e);
      err.hidden = false;
    }).then(function () { setBusy(up, false); file.value = ''; });
  });
  paint();
  return h('div', {}, h('div', { class: 'flex items-center gap-3' }, preview, h('div', { class: 'flex-1 min-w-0 flex flex-col gap-1.5' }, urlInput, h('div', {}, up))), file, err);
}

function toLocalInput(v) {
  if (!v) return '';
  var d = new Date(v);
  if (isNaN(d.getTime())) return '';
  var p = function (n) { return (n < 10 ? '0' : '') + n; };
  return d.getFullYear() + '-' + p(d.getMonth() + 1) + '-' + p(d.getDate()) + 'T' + p(d.getHours()) + ':' + p(d.getMinutes());
}

function dateInput(value, onchange) {
  var el = h('input', { type: 'datetime-local', class: INPUT_MD + ' tabular-nums', dir: 'ltr', 'data-ltr': '' });
  el.value = toLocalInput(value);
  var shown = h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1' });
  function paint() { shown.textContent = el.value ? fmtDate(new Date(el.value).toISOString()) : ''; shown.hidden = !el.value || lang() !== 'fa'; }
  el.addEventListener('input', function () {
    paint();
    onchange(el.value ? new Date(el.value).toISOString() : null);
  });
  paint();
  return h('div', {}, el, shown);
}

function trValue(v) {
  var out = {};
  if (v && typeof v === 'object') Object.keys(v).forEach(function (k) { if (typeof v[k] === 'string') out[k] = v[k]; });
  else if (typeof v === 'string' && v) out[LANGS.def] = v;
  return out;
}

function trClean(v) {
  var out = {};
  Object.keys(v || {}).forEach(function (k) {
    if (LANGS.codes.indexOf(k) === -1) return;
    var s = String(v[k] == null ? '' : v[k]).trim();
    if (s) out[k] = s;
  });
  return out;
}

function trCheck(v, opts) {
  var c = trClean(v);
  var any = Object.keys(c).length > 0;
  if ((opts.required || any) && !c[LANGS.def]) return { code: 'DEFAULT_LANGUAGE_REQUIRED', lang: LANGS.def };
  if (opts.max) {
    for (var i = 0; i < LANGS.items.length; i++) {
      var code = LANGS.items[i].code;
      if (c[code] && Array.from(c[code]).length > opts.max) return { code: 'TEXT_TOO_LONG', lang: code };
    }
  }
  return null;
}

function trServerLang(err, v, opts) {
  if (!err) return LANGS.def;
  if (err.code === 'INVALID_LANGUAGE') {
    var m = /language\s+(\S+)/.exec(err.message || '');
    return m && LANGS.active.some(function (l) { return l.code === m[1]; }) ? m[1] : LANGS.def;
  }
  if (err.code === 'TEXT_TOO_LONG') {
    var hit = trCheck(v, { max: opts.max });
    if (hit && hit.code === 'TEXT_TOO_LONG') return hit.lang;
    var best = LANGS.def;
    var len = -1;
    Object.keys(v || {}).forEach(function (k) { var n = Array.from(String(v[k] || '')).length; if (n > len && LANGS.active.some(function (l) { return l.code === k; })) { len = n; best = k; } });
    return best;
  }
  return LANGS.def;
}

function translatedInput(value, opts) {
  opts = opts || {};
  var values = trValue(value);
  var langs = LANGS.active.slice();
  var current = langs.length ? langs[0].code : LANGS.def;
  var tabs = h('div', { class: 'eco-tr-tabs', role: 'tablist', 'aria-label': opts.label || null });
  var panels = h('div', { class: 'eco-tr-panels' });
  var parts = {};
  langs.forEach(function (l) {
    var code = l.code;
    var dir = langDir(code);
    var dot = h('span', { class: 'eco-tr-dot', 'aria-hidden': 'true' });
    var tab = h('button', { type: 'button', role: 'tab', class: 'eco-tr-tab', 'data-lang': code, 'aria-selected': 'false', title: langName(code),
      onclick: function () { select(code); } },
      h('span', { dir: dir, text: l.nativeName || code }),
      code === LANGS.def ? h('span', { class: 'eco-tr-def', text: t('tr.default') }) : null,
      dot);
    var label = (opts.label ? opts.label + ' - ' : '') + langName(code);
    var ctl;
    var onin = function (v) {
      values[code] = v;
      paintTab(code);
      tab.classList.remove('eco-tr-bad');
      if (opts.oninput) opts.oninput(Object.assign({}, values));
    };
    if (opts.markdown) {
      ctl = markdownInput(values[code] || '', { ltr: dir === 'ltr', rows: opts.rows, max: opts.max, label: label, oninput: onin });
      var ta = ctl.querySelector('textarea');
      if (ta) ta.setAttribute('dir', dir);
    } else {
      var input = textInput(values[code] || '', { multiline: !!opts.multiline, rows: opts.rows || 3, label: label, oninput: function (v) { onin(v); if (count) paintCount(); } });
      input.setAttribute('dir', dir);
      var count = opts.counter && opts.max ? h('span', { class: 'text-xs tabular-nums text-secondary-500' }) : null;
      var paintCount = function () {
        var n = Array.from(input.value).length;
        count.textContent = t('chars', { n: n, max: opts.max });
        count.classList.toggle('text-error-600', n > opts.max);
      };
      if (count) paintCount();
      ctl = count ? h('div', {}, input, h('div', { class: 'flex justify-end mt-1' }, count)) : input;
    }
    var panel = h('div', { role: 'tabpanel', 'data-lang': code, hidden: true }, ctl);
    parts[code] = { tab: tab, dot: dot, panel: panel, ctl: ctl };
    tabs.appendChild(tab);
    panels.appendChild(panel);
  });
  function paintTab(code) {
    var p = parts[code];
    if (!p) return;
    var empty = !String(values[code] || '').trim();
    p.dot.hidden = !empty;
    p.tab.classList.toggle('eco-tr-empty', empty);
    p.tab.setAttribute('aria-label', langName(code) + (code === LANGS.def ? ' (' + t('tr.default') + ')' : '') + (empty ? ' (' + t('tr.empty') + ')' : ''));
  }
  function select(code) {
    if (!parts[code]) return;
    current = code;
    Object.keys(parts).forEach(function (k) {
      parts[k].panel.hidden = k !== code;
      parts[k].tab.setAttribute('aria-selected', k === code ? 'true' : 'false');
      parts[k].tab.classList.toggle('eco-tr-on', k === code);
    });
  }
  Object.keys(parts).forEach(paintTab);
  select(current);
  var el = h('div', { class: 'eco-tr', 'data-translated': opts.name || '' }, tabs, panels);
  el.select = select;
  el.markLang = function (code) {
    Object.keys(parts).forEach(function (k) { parts[k].tab.classList.toggle('eco-tr-bad', k === code); });
    if (code && parts[code]) {
      select(code);
      var input = parts[code].panel.querySelector('input,textarea');
      if (input) setTimeout(function () { try { input.focus({ preventScroll: true }); } catch (e) { input.focus(); } }, 0);
    }
  };
  el.value = function () { return Object.assign({}, values); };
  el.current = function () { return current; };
  return el;
}

function mdEscape(s) {
  return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&#34;').replace(/'/g, '&#39;');
}

function mdSafeUrl(u) {
  if (/[\s"'<>`]/.test(u)) return false;
  if (/^https?:\/\//i.test(u)) return true;
  return u.charAt(0) === '/' && u.charAt(1) !== '/' && u.charAt(1) !== '\\';
}

function mdLink(s, start) {
  var depth = 0;
  var i = start + 1;
  for (; i < s.length; i++) {
    if (s[i] === '[') depth++;
    else if (s[i] === ']') { if (depth === 0) break; depth--; }
  }
  if (i >= s.length - 1 || s[i + 1] !== '(') return null;
  var close = s.indexOf(')', i + 2);
  if (close < 0) return null;
  return { text: s.slice(start + 1, i), url: s.slice(i + 2, close).trim(), end: close + 1 };
}

function mdInline(s, links) {
  var out = '';
  var i = 0;
  while (i < s.length) {
    var m;
    if (s.slice(i, i + 2) === '![' && (m = mdLink(s, i + 1)) && mdSafeUrl(m.url)) {
      out += '<img src="' + mdEscape(m.url) + '" alt="' + mdEscape(m.text) + '">';
      i = m.end;
      continue;
    }
    if (s[i] === '[' && links && (m = mdLink(s, i)) && mdSafeUrl(m.url)) {
      out += '<a href="' + mdEscape(m.url) + '" rel="noopener noreferrer" target="_blank">' + mdInline(m.text, false) + '</a>';
      i = m.end;
      continue;
    }
    if (s.slice(i, i + 2) === '**') {
      var e2 = s.indexOf('**', i + 2);
      if (e2 > i + 2) { out += '<strong>' + mdInline(s.slice(i + 2, e2), links) + '</strong>'; i = e2 + 2; continue; }
    }
    if (s[i] === '*' && i + 1 < s.length && s[i + 1] !== ' ' && s[i + 1] !== '*') {
      var e1 = s.indexOf('*', i + 1);
      if (e1 > i + 1) { out += '<em>' + mdInline(s.slice(i + 1, e1), links) + '</em>'; i = e1 + 1; continue; }
    }
    var j = i + 1;
    while (j < s.length && s[j] !== '*' && s[j] !== '[' && s[j] !== '!') j++;
    out += mdEscape(s.slice(i, j));
    i = j;
  }
  return out;
}

function mdHtml(src) {
  var blocks = [];
  var cur = null;
  function flush() { if (cur && cur.lines.length) blocks.push(cur); cur = null; }
  String(src || '').replace(/\r\n?/g, '\n').split('\n').forEach(function (line) {
    var tl = line.trim();
    if (!tl) { flush(); return; }
    if (/^### /.test(tl)) { flush(); blocks.push({ kind: 'h3', lines: [tl.slice(4).trim()] }); return; }
    if (/^## /.test(tl)) { flush(); blocks.push({ kind: 'h2', lines: [tl.slice(3).trim()] }); return; }
    var kind = /^[-*] /.test(tl) ? 'ul' : /^[0-9]{1,3}[.)] /.test(tl) ? 'ol' : 'p';
    if (!cur || cur.kind !== kind) { flush(); cur = { kind: kind, lines: [] }; }
    cur.lines.push(kind === 'ul' ? tl.slice(2).trim() : kind === 'ol' ? tl.replace(/^[0-9]{1,3}[.)] /, '').trim() : tl);
  });
  flush();
  return blocks.map(function (b) {
    if (b.kind === 'h2' || b.kind === 'h3') return '<' + b.kind + '>' + mdInline(b.lines[0], true) + '</' + b.kind + '>';
    if (b.kind === 'ul' || b.kind === 'ol') return '<' + b.kind + '>' + b.lines.map(function (l) { return '<li>' + mdInline(l, true) + '</li>'; }).join('') + '</' + b.kind + '>';
    return '<p>' + b.lines.map(function (l) { return mdInline(l, true); }).join('<br>') + '</p>';
  }).join('');
}

function mdPlain(s) {
  return String(s || '').replace(/!\[[^\]]*\]\([^)]*\)/g, '').replace(/\[([^\]]*)\]\([^)]*\)/g, '$1').replace(/^\s*(#{2,3}|[-*]|\d{1,3}[.)])\s+/gm, '').replace(/\*\*|\*/g, '').replace(/\s+/g, ' ').trim();
}

function mdPreview(src, ltr) {
  var box = h('div', { class: 'eco-md text-sm text-secondary-800 dark:text-secondary-200', dir: ltr ? 'ltr' : 'rtl' });
  var html = mdHtml(src);
  if (html) box.innerHTML = html;
  else box.appendChild(h('p', { class: 'text-secondary-400', text: t('md.empty') }));
  return box;
}

function markdownInput(value, opts) {
  var text = value == null ? '' : String(value);
  var ta = textInput(text, { multiline: true, rows: opts.rows || 6, ltr: opts.ltr, label: opts.label, oninput: function (v) { text = v; opts.oninput(v); if (counter) paintCount(); } });
  var preview = h('div', { class: 'eco-md-box border border-secondary-200 dark:border-secondary-700 rounded-md px-3 py-2 bg-secondary-50 dark:bg-secondary-900', hidden: true });
  var counter = opts.max ? h('span', { class: 'text-xs tabular-nums text-secondary-500' }) : null;
  function paintCount() {
    var n = Array.from(text).length;
    counter.textContent = t('chars', { n: n, max: opts.max });
    counter.classList.toggle('text-error-600', n > opts.max);
  }
  var seg = segmented([{ value: 'write', label: t('md.write') }, { value: 'preview', label: t('md.preview') }], 'write', function (v) {
    ta.hidden = v === 'preview';
    preview.hidden = v !== 'preview';
    if (v === 'preview') clear(preview).appendChild(mdPreview(text, opts.ltr));
  }, t('md.mode'));
  if (counter) paintCount();
  return h('div', { class: 'eco-md-input' },
    h('div', { class: 'flex items-center justify-between gap-2 mb-2' }, seg, counter),
    ta, preview,
    h('div', { class: 'flex flex-wrap items-center gap-1 text-xs text-secondary-500 dark:text-secondary-400 mt-1' },
      h('span', { text: t('md.hint') }),
      ['**bold**', '*italic*', '[text](https://…)', '## heading', '- list', '![](https://…)'].map(function (s) { return h('code', { class: 'eco-md-token', dir: 'ltr', text: s }); })));
}

var GOTO_SCREENS = ['home', 'play', 'shop', 'gift', 'payments', 'news', 'faq', 'inbox'];

function actionsEditor(value, onchange) {
  var rows = (value || []).map(function (r) { return { type: r.type || 'url', value: r.value || '', label: trValue(r.label) }; });
  var listEl = h('div', { class: 'flex flex-col gap-2' });
  var errors = [];
  var listId = 'eco-goto-' + Math.random().toString(36).slice(2, 8);
  var datalist = h('datalist', { id: listId }, GOTO_SCREENS.map(function (s) { return h('option', { value: s }); }));
  function emit() { onchange(rows.map(function (r) { return { type: r.type, value: r.value, label: Object.assign({}, r.label) }; })); }
  var add = button({ variant: 'link', size: 'sm', icon: 'plus', label: t('act.add'), onclick: function () {
    rows.push({ type: 'url', value: '', label: {} });
    render();
    emit();
  } });
  function render() {
    clear(listEl);
    errors = [];
    add.hidden = rows.length >= 5;
    if (!rows.length) listEl.appendChild(h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400 border border-dashed border-secondary-300 dark:border-secondary-700 rounded-md px-3 py-4 text-center', text: t('act.empty') }));
    rows.forEach(function (row, idx) {
      var cell = function (labelKey, ctl) {
        return h('div', { class: 'min-w-0' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t(labelKey) }), ctl);
      };
      var typeSel = selectInput(row.type, [{ value: 'url', label: t('act.url') }, { value: 'goto', label: t('act.goto') }], function (v) { row.type = v; render(); emit(); }, { placeholder: false, label: t('act.type') });
      var val = textInput(row.value, { ltr: true, mono: true, label: t('act.value'), placeholder: row.type === 'url' ? 'https://' : 'shop', oninput: function (v) { row.value = v.trim(); emit(); } });
      if (row.type === 'goto') val.setAttribute('list', listId);
      var label = translatedInput(row.label, { label: t('act.label'), oninput: function (v) { row.label = v; emit(); } });
      var err = h('p', { class: 'text-xs text-error-600 dark:text-error-400 eco-items-err', hidden: true });
      errors.push({ row: row, el: err, label: label });
      listEl.appendChild(h('div', { class: 'border border-secondary-200 dark:border-secondary-700 rounded-md p-3', 'data-action-row': '' },
        h('div', { class: 'eco-actions-row' },
          cell('act.type', typeSel),
          cell('act.value', val),
          h('div', { class: 'min-w-0 eco-actions-label' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('act.label') }), label),
          h('div', { class: 'eco-items-remove' }, button({ variant: 'danger', size: 'iconLg', icon: 'trash', title: t('remove'), onclick: function () { rows.splice(idx, 1); render(); emit(); } }))),
        h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1.5', text: t(row.type === 'url' ? 'act.url.hint' : 'act.goto.hint') }),
        err));
    });
  }
  render();
  var el = h('div', {}, listEl, datalist, h('div', { class: 'mt-2' }, add));
  function show(e, msg, code) {
    e.el.textContent = msg;
    e.el.hidden = !msg;
    e.el.parentElement.classList.toggle('eco-row-invalid', !!msg);
    e.label.markLang(code || null);
  }
  el.validate = function () {
    var bad = false;
    errors.forEach(function (e) {
      var r = e.row;
      var msg = '';
      var hit = trCheck(r.label, { required: true, max: 64 });
      if (hit) msg = t('act.label') + ': ' + t('code.' + hit.code, { lang: langName(hit.lang) });
      else if (!r.value) msg = t('act.value') + ': ' + t('err.required');
      else if (r.type === 'url' && !/^https?:\/\/\S+$/i.test(r.value)) msg = t('code.INVALID_ACTION');
      show(e, msg, hit ? hit.lang : null);
      if (msg) bad = true;
    });
    return bad ? t('err.form') : null;
  };
  el.showServerError = function (err) {
    if (!err || !/label$/.test(String(err.field || ''))) return false;
    var target = errors.filter(function (e) { return trCheck(e.row.label, { required: true, max: 64 }); })[0] || errors[0];
    if (!target) return false;
    var code = trServerLang(err, target.row.label, { max: 64 });
    show(target, t('act.label') + ': ' + errorText(err, code), code);
    target.el.scrollIntoView({ block: 'center', behavior: 'smooth' });
    return true;
  };
  return el;
}

var ITEM_ACTIONS = { currency: ['add'], meta: ['add', 'set'], storage: ['give'], item: ['give'], achievement: ['give'] };

function itemsEditor(value, onchange) {
  var rows = (value || []).map(function (r) { return Object.assign({}, r); });
  var listEl = h('div', { class: 'flex flex-col gap-2' });
  var errors = [];
  function emit() { onchange(rows.map(function (r) { return Object.assign({}, r); })); }
  function render() {
    clear(listEl);
    errors = [];
    if (!rows.length) listEl.appendChild(h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400 border border-dashed border-secondary-300 dark:border-secondary-700 rounded-md px-3 py-4 text-center', text: t('f.items.empty') }));
    rows.forEach(function (row, idx) {
      var typeSel = selectInput(row.type, ['currency', 'meta', 'item', 'achievement', 'storage'].map(function (k) { return { value: k, label: t('item.' + k) }; }), function (v) {
        row.type = v;
        row.key = '';
        row.action = (ITEM_ACTIONS[v] || ['add'])[0];
        row.value = v === 'storage' ? {} : v === 'item' || v === 'achievement' ? 1 : 0;
        render();
        emit();
      }, { placeholder: false, label: t('f.item.type') });
      var keyCtl;
      if (row.type === 'storage') {
        keyCtl = textInput(row.key, { ltr: true, mono: true, label: t('f.item.key'), oninput: function (v) { row.key = v.trim(); emit(); } });
      } else if (row.type === 'item') {
        keyCtl = refPicker({ resource: 'items', valueKey: 'key', value: row.key, required: true, label: t('f.item.key'), onchange: function (v) { row.key = v || ''; emit(); } });
        keyCtl.setAttribute('data-item-picker', '');
      } else if (row.type === 'achievement') {
        keyCtl = refPicker({ resource: 'achievements', valueKey: 'key', value: row.key, required: true, label: t('f.item.key'), onchange: function (v) { row.key = v || ''; emit(); } });
        keyCtl.setAttribute('data-achievement-picker', '');
      } else {
        keyCtl = refPicker({ resource: row.type === 'meta' ? 'metas' : 'currencies', valueKey: 'key', value: row.key, required: true, label: t('f.item.key'), onchange: function (v) { row.key = v || ''; emit(); } });
      }
      var actSel = selectInput(row.action, (ITEM_ACTIONS[row.type] || ['add']).map(function (k) { return { value: k, label: t('action.' + k) }; }), function (v) { row.action = v; emit(); }, { placeholder: false, label: t('f.item.action') });
      var valCtl;
      if (row.type === 'item' || row.type === 'achievement') {
        valCtl = h('div', { class: 'text-sm text-secondary-600 dark:text-secondary-300 py-2', text: t('item.' + row.type + '.once') });
      } else if (row.type === 'storage') {
        valCtl = textInput(row.value == null ? '{}' : JSON.stringify(row.value), { ltr: true, mono: true, label: t('f.item.value'), oninput: function (v) {
          try { row.value = JSON.parse(v); row._bad = false; } catch (e) { row._bad = true; }
          emit();
        } });
      } else {
        valCtl = numberInput(row.value, { int: row.type === 'currency', label: t('f.item.value'), oninput: function (v) { row._raw = v; row.value = parseNumber(v, row.type === 'currency'); emit(); } });
      }
      var err = h('p', { class: 'text-xs text-error-600 dark:text-error-400 eco-items-err', hidden: true });
      errors.push({ row: row, el: err });
      var cell = function (labelKey, ctl) {
        return h('div', { class: 'min-w-0' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t(labelKey) }), ctl);
      };
      listEl.appendChild(h('div', { class: 'border border-secondary-200 dark:border-secondary-700 rounded-md p-3' },
        h('div', { class: 'eco-items-row' },
          cell('f.item.type', typeSel),
          cell('f.item.key', keyCtl),
          cell('f.item.action', actSel),
          cell('f.item.value', valCtl),
          h('div', { class: 'eco-items-remove' }, button({ variant: 'danger', size: 'iconLg', icon: 'trash', title: t('remove'), onclick: function () { rows.splice(idx, 1); render(); emit(); } }))),
        err));
    });
  }
  var add = button({ variant: 'link', size: 'sm', icon: 'plus', label: t('add.row'), onclick: function () {
    rows.push({ type: 'currency', key: '', action: 'add', value: 0 });
    render();
    emit();
  } });
  render();
  var el = h('div', {}, listEl, h('div', { class: 'mt-2' }, add));
  el.validate = function () {
    var bad = false;
    errors.forEach(function (e) {
      var r = e.row;
      var msg = '';
      if (!r.key) msg = t('f.item.key') + ': ' + t('err.required');
      else if (r.type === 'storage' && (r._bad || r.value === null || typeof r.value !== 'object')) msg = t('err.json');
      else if (r.type !== 'storage' && r.type !== 'item' && r.type !== 'achievement' && (r.value === null || r.value === undefined || isNaN(r.value))) msg = t('f.item.value') + ': ' + t('err.number');
      e.el.textContent = msg;
      e.el.hidden = !msg;
      e.el.parentElement.classList.toggle('eco-row-invalid', !!msg);
      if (msg) bad = true;
    });
    return bad ? t('err.form') : null;
  };
  return el;
}

function buildForm(fields, values, ctx) {
  var root = h('div', {});
  var entries = [];

  function addField(f, parent) {
    if (f.type === 'row') {
      var row = h('div', { class: 'eco-form-row' });
      f.fields.forEach(function (sub) { addField(sub, row); });
      var entryRow = { def: f, el: row, row: true };
      entries.push(entryRow);
      parent.appendChild(row);
      return;
    }
    if (f.type === 'heading') {
      var hd = h('div', { class: 'text-sm font-semibold text-secondary-800 dark:text-secondary-200 mb-3 mt-1 pt-4 border-t border-secondary-200 dark:border-secondary-700', text: t(f.label) });
      entries.push({ def: f, el: hd, row: true });
      parent.appendChild(hd);
      return;
    }
    var cur = getPath(values, f.name);
    var change = function (v) {
      setPath(values, f.name, v);
      if (entry.wrap.setError) entry.wrap.setError('');
      refresh();
    };
    var label = f.label ? t(f.label) : '';
    var control;
    var t2 = f.type || 'text';
    if (t2 === 'text' || t2 === 'key' || t2 === 'textarea' || t2 === 'code') {
      control = textInput(cur, { ltr: t2 === 'key' || t2 === 'code' || f.ltr, mono: t2 === 'key' || t2 === 'code', multiline: t2 === 'textarea', label: label, oninput: function (v) {
        if (t2 === 'key') v = v.trim();
        if (t2 === 'code') v = v.trim().toUpperCase();
        change(v);
      } });
      if (t2 === 'code') control.classList.add('uppercase');
      if (f.lockOnEdit && values.id) {
        control.readOnly = true;
        control.classList.add('eco-readonly');
      }
    } else if (t2 === 'json') {
      control = textInput(cur == null ? '' : JSON.stringify(cur, null, 2), { ltr: true, mono: true, multiline: true, rows: 5, label: label, oninput: function (v) {
        entry.raw = v;
        setPath(values, f.name, v);
        if (entry.wrap.setError) entry.wrap.setError('');
      } });
      entry = { raw: cur == null ? '' : JSON.stringify(cur) };
    } else if (t2 === 'number') {
      control = numberInput(cur, { int: f.int !== false, label: label, oninput: function (v) { change(v === '' ? null : v); } });
    } else if (t2 === 'bool') {
      control = toggle(cur, f.inline ? label : null, function (v) { change(v); });
    } else if (t2 === 'select') {
      control = selectInput(cur, f.options(), function (v) { change(f.noPlaceholder ? v : v || null); }, { label: label, placeholder: f.required || f.noPlaceholder ? false : t('none') });
      if (f.required && (cur == null || cur === '')) { setPath(values, f.name, f.options()[0].value); }
    } else if (t2 === 'multi') {
      var set = (cur || []).slice();
      control = h('div', { class: 'flex flex-wrap gap-4 eco-choices' });
      f.options().forEach(function (o) {
        control.appendChild(checkbox(set.indexOf(o.value) !== -1, o.label, function (on) {
          set = set.filter(function (x) { return x !== o.value; });
          if (on) set.push(o.value);
          var order = f.options().map(function (x) { return x.value; });
          set.sort(function (a, b) { return order.indexOf(a) - order.indexOf(b); });
          change(set.slice());
        }));
      });
    } else if (t2 === 'i18n') {
      control = translatedInput(cur, { name: f.name, label: label, multiline: f.multiline, markdown: f.markdown, rows: f.rows, max: f.max, oninput: function (v) {
        setPath(values, f.name, v);
        if (entry.wrap.setError) entry.wrap.setError('');
        control.markLang(null);
        if (ctx && ctx.onchange) ctx.onchange(values);
      } });
    } else if (t2 === 'ref') {
      control = refPicker({ resource: f.resource, api: f.api, valueKey: f.valueKey, filter: f.filter, value: cur, required: f.required, label: label, onchange: function (v) { change(v); } });
      if (f.quick) {
        var single = control;
        control = h('div', { class: 'eco-ref-row' }, single, quickCreate({ resource: f.resource, keyType: f.quick, onCreated: function (item) { single.select(item); } }));
      }
    } else if (t2 === 'refMulti') {
      control = refMulti({ resource: f.resource, api: f.api, filter: f.filter, value: cur, label: label, quick: f.quick, onchange: function (v) { change(v); } });
    } else if (t2 === 'image') {
      control = imageInput(cur, function (v) { change(v); });
    } else if (t2 === 'datetime') {
      control = dateInput(cur, function (v) { change(v); });
    } else if (t2 === 'items') {
      control = itemsEditor(cur, function (v) { setPath(values, f.name, v); if (entry.wrap.setError) entry.wrap.setError(''); });
    } else if (t2 === 'markdown') {
      control = markdownInput(cur, { ltr: f.ltr, rows: f.rows, max: f.max, label: label, oninput: function (v) { change(v); } });
    } else if (t2 === 'actions') {
      control = actionsEditor(cur, function (v) { setPath(values, f.name, v); if (entry.wrap.setError) entry.wrap.setError(''); });
    } else if (t2 === 'rows') {
      control = rowsEditor(cur, f, function (v) { setPath(values, f.name, v); if (entry.wrap.setError) entry.wrap.setError(''); if (f.refresh) refresh(); });
    } else if (t2 === 'links') {
      control = storeLinksEditor(cur, function (v) { setPath(values, f.name, v); if (entry.wrap.setError) entry.wrap.setError(''); });
    } else if (t2 === 'static') {
      control = h('div', { class: 'py-2 text-secondary-700 dark:text-secondary-200' }, f.render(values));
    }
    var entry = entry || {};
    entry.def = f;
    entry.control = control;
    if (t2 === 'bool' && f.inline) {
      entry.wrap = h('div', { class: 'mb-5 eco-field' }, control,
        f.hint ? h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: t(f.hint) }) : null);
      entry.wrap.setError = function () {};
    } else {
      var hintKey = f.lockOnEdit && values.id ? 'f.key.locked' : f.hint;
      entry.wrap = field(label, control, hintKey ? t(hintKey) : null, { optional: f.optional });
    }
    entries.push(entry);
    parent.appendChild(entry.wrap);
  }

  function refresh() {
    entries.forEach(function (e) {
      if (e.def.live && e.control) append(clear(e.control), e.def.render(values));
      var visible = !e.def.when || e.def.when(values);
      var el = e.row ? e.el : e.wrap;
      el.hidden = !visible;
      if (!visible && !e.hidden && !e.row && e.wrap && e.wrap.setError) e.wrap.setError(null);
      e.hidden = !visible;
    });
    if (ctx && ctx.onchange) ctx.onchange(values);
  }

  function validate() {
    var first = null;
    var errs = 0;
    entries.forEach(function (e) {
      if (e.row || e.hidden) return;
      var f = e.def;
      var v = getPath(values, f.name);
      var msg = null;
      var t2 = f.type || 'text';
      if (t2 === 'json') {
        var raw = typeof v === 'string' ? v : (v == null ? '' : JSON.stringify(v));
        if (raw.trim() === '') setPath(values, f.name, f.empty !== undefined ? f.empty : {});
        else { try { setPath(values, f.name, JSON.parse(raw)); } catch (x) { msg = t('err.json'); } }
      } else if (t2 === 'number') {
        var n = parseNumber(v, f.int !== false);
        if (n === null) { if (f.required) msg = t('err.required'); setPath(values, f.name, null); }
        else if (isNaN(n) || (f.min !== undefined && n < f.min) || (f.max !== undefined && n > f.max)) msg = t('err.number');
        else setPath(values, f.name, n);
      } else if (t2 === 'items' || t2 === 'actions' || t2 === 'links' || t2 === 'rows') {
        msg = e.control.validate();
      } else if (t2 === 'i18n') {
        var hit = trCheck(v, f);
        e.control.markLang(hit ? hit.lang : null);
        if (hit) msg = t('code.' + hit.code, { lang: langName(hit.lang) });
      } else if (f.required && (v === null || v === undefined || v === '' || (Array.isArray(v) && !v.length))) {
        msg = t('err.required');
      } else if (t2 === 'key' && v && !(f.keyRe || KEY_RE).test(v)) {
        msg = t(f.keyErr || (f.keyRe ? 'err.key' : 'err.key.general'));
      } else if (t2 === 'code' && v && !CODE_RE.test(v)) {
        msg = t('err.code');
      }
      if (!msg && f.check) msg = f.check(v, values);
      e.wrap.setError(msg);
      if (msg) { errs++; if (!first) first = e.wrap; }
    });
    if (first) first.scrollIntoView({ block: 'center', behavior: 'smooth' });
    return errs === 0;
  }

  function showServerError(err) {
    if (!err || !err.field) return false;
    var name = String(err.field).replace(/\[\d+\]/g, '').replace(/_([a-z])/g, function (m, c) { return c.toUpperCase(); });
    var best = null;
    entries.forEach(function (e) {
      if (e.row || e.hidden || !e.def.name) return;
      if (name === e.def.name || name.indexOf(e.def.name + '.') === 0 || e.def.name.indexOf(name + '.') === 0) { if (!best) best = e; }
    });
    if (!best) return false;
    if ((best.def.type === 'actions' || best.def.type === 'rows') && best.control.showServerError) return best.control.showServerError(err);
    if (best.def.type === 'i18n') {
      var code = trServerLang(err, getPath(values, best.def.name), best.def);
      best.control.markLang(code);
      best.wrap.setError(errorText(err, code));
      best.wrap.scrollIntoView({ block: 'center', behavior: 'smooth' });
      return true;
    }
    best.wrap.setError(errorText(err));
    best.wrap.scrollIntoView({ block: 'center', behavior: 'smooth' });
    return true;
  }

  function payload() {
    var out = {};
    fields.forEach(function walk(f) {
      if (f.type === 'row') { f.fields.forEach(walk); return; }
      if (!f.name || f.readonly || f.type === 'static' || f.type === 'heading') return;
      var visible = !f.when || f.when(values);
      var v = getPath(values, f.name);
      if (!visible) v = f.empty !== undefined ? f.empty : null;
      if (f.type === 'actions' && Array.isArray(v)) v = v.map(function (r) { return { type: r.type, value: r.value, label: trClean(r.label) }; });
      if (f.type === 'i18n') v = trClean(v);
      if (f.type === 'links' && Array.isArray(v)) v = v.map(function (r) { return { store: r.store, url: r.url, label: trClean(r.label) }; });
      if (f.type === 'items' && Array.isArray(v)) v = v.map(function (r) { return { type: r.type, key: r.key, action: r.action, value: r.value }; });
      setPath(out, f.name, v === undefined ? null : v);
    });
    return out;
  }

  fields.forEach(function (f) { addField(f, root); });
  refresh();
  return { el: root, validate: validate, payload: payload, showServerError: showServerError };
}

var METHODS = ['zarinpal', 'zibal', 'bazaar', 'myket', 'in_app'];
var GATEWAYS = ['zarinpal', 'zibal'];
var STORES = ['bazaar', 'myket'];
var CATEGORY_KEY_TYPES = ['product', 'prize', 'gift', 'currency', 'faq', 'news', 'task'];
var TAG_KEY_TYPES = ['product', 'prize', 'gift', 'currency', 'faq', 'news', 'task', 'general', 'user'];
var LANG_CODE_RE = /^[a-z]{2,3}(-[a-z0-9]{2,8})?$/;
var keyTypeCache = {};
var keyTypePromise = null;

function loadKeyTypes() {
  if (keyTypePromise) return keyTypePromise;
  keyTypePromise = Promise.all(['categories', 'tags'].map(function (r) {
    return baseApi('GET', '/' + r + '/key-types').then(function (res) { if (res && res.items && res.items.length) keyTypeCache[r] = res.items; }, function () { return null; });
  })).then(function () { return keyTypeCache; }, function () { keyTypePromise = null; return keyTypeCache; });
  return keyTypePromise;
}

function keyTypeLabel(resource, key) {
  var hit = (keyTypeCache[resource] || []).filter(function (k) { return k.key === key; })[0];
  return hit && hit.name && i18nText(hit.name) ? i18nText(hit.name) : has('kind.' + key) ? t('kind.' + key) : key;
}

function keyTypeOptions(resource) {
  return function () {
    var list = keyTypeCache[resource] ? keyTypeCache[resource].map(function (k) { return k.key; }) : resource === 'tags' ? TAG_KEY_TYPES : CATEGORY_KEY_TYPES;
    return list.map(function (k) { return { value: k, label: keyTypeLabel(resource, k) }; });
  };
}

function categoryField(keyType) {
  return { name: 'categoryId', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: keyType }, quick: keyType, label: 'f.category', optional: true };
}

function tagsField(keyType) {
  return { name: 'tagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: [{ keyType: keyType }, { keyType: 'general' }], quick: keyType, label: 'f.tags', optional: true, empty: [] };
}

function tr(name, opts) {
  return Object.assign({ name: name, type: 'i18n', label: 'f.' + name }, opts || {});
}
var STATUS_COLOR = { pending: 'warning', success: 'success', failed: 'error', expired: 'secondary' };

function opts(prefix, keys) {
  return function () { return keys.map(function (k) { return { value: k, label: t(prefix + k) }; }); };
}

function hasAny(list, wanted) {
  return (list || []).some(function (m) { return wanted.indexOf(m) !== -1; });
}

function nameCell(it, withIcon) {
  var name = i18nText(it.name) || it.name || '';
  var text = h('div', { class: 'min-w-0' },
    h('div', { class: 'font-medium truncate', text: typeof name === 'string' ? name : '' }),
    it.key || it.code ? h('div', { class: 'text-secondary-500' }, mono(it.key || it.code)) : null);
  if (!withIcon) return text;
  var pic = h('div', { class: 'h-8 w-8 shrink-0 rounded-md bg-secondary-100 dark:bg-secondary-700 overflow-hidden flex items-center justify-center text-secondary-400' });
  if (it.iconUrl) pic.appendChild(img(it.iconUrl));
  else pic.appendChild(icon('image', 'h-4 w-4'));
  return h('div', { class: 'flex items-center gap-3 min-w-0' }, pic, text);
}

var taxonomyNames = { categories: {}, tags: {} };
var taxonomyPending = { categories: {}, tags: {} };

function taxonomyLookup(kind, id) {
  if (taxonomyNames[kind][id]) return Promise.resolve(taxonomyNames[kind][id]);
  if (!taxonomyPending[kind][id]) {
    taxonomyPending[kind][id] = baseApi('GET', '/' + kind + '/' + encodeURIComponent(id)).then(function (res) {
      taxonomyNames[kind][id] = res;
      return res;
    }, function () { delete taxonomyPending[kind][id]; return null; });
  }
  return taxonomyPending[kind][id];
}

function taxonomyCell(it) {
  var box = h('div', { class: 'flex flex-wrap gap-1', 'data-taxonomy': '' });
  var ids = (it.tagIds || []).slice(0, 3);
  function paint() {
    clear(box);
    var cat = it.categoryId ? taxonomyNames.categories[it.categoryId] : null;
    if (cat) box.appendChild(tag(i18nText(cat.name) || cat.key, 'secondary'));
    ids.forEach(function (id) {
      var tg = taxonomyNames.tags[id];
      if (tg) box.appendChild(tag(i18nText(tg.name) || tg.key, 'primary'));
    });
    if ((it.tagIds || []).length > 3) box.appendChild(tag('+' + fmtNum(it.tagIds.length - 3), 'secondary'));
    if (!box.childNodes.length) box.appendChild(h('span', { class: 'text-secondary-400', text: t('none') }));
  }
  var jobs = [];
  if (it.categoryId && !taxonomyNames.categories[it.categoryId]) jobs.push(taxonomyLookup('categories', it.categoryId));
  ids.forEach(function (id) { if (!taxonomyNames.tags[id]) jobs.push(taxonomyLookup('tags', id)); });
  if (jobs.length) {
    box.appendChild(bar('4rem', 'h-4'));
    Promise.all(jobs).then(paint);
  } else paint();
  return box;
}

function numCell(v) {
  return h('span', { class: 'tabular-nums', text: v === null || v === undefined ? '' : fmtNum(v) });
}

function dateCell(v) {
  return h('span', { class: 'text-secondary-600 dark:text-secondary-300 whitespace-nowrap', text: fmtDate(v) });
}

function activeTag(on) {
  return tag(on ? t('active') : t('inactive'), on ? 'success' : 'secondary');
}

var KEY_NAMES = { data: null, promise: null, at: 0 };

function keyNames() {
  if (KEY_NAMES.data && Date.now() - KEY_NAMES.at < 30000) return Promise.resolve(KEY_NAMES.data);
  if (KEY_NAMES.promise) return KEY_NAMES.promise;
  var pick = function (res) {
    var out = {};
    ((res && res.items) || []).forEach(function (it) { out[it.key] = it.name; });
    return out;
  };
  KEY_NAMES.promise = Promise.all([
    api('GET', '/currencies?limit=100').then(pick, function () { return {}; }),
    api('GET', '/metas?limit=100').then(pick, function () { return {}; })
  ]).then(function (r) {
    KEY_NAMES.data = { currency: r[0], meta: r[1] };
    KEY_NAMES.at = Date.now();
    KEY_NAMES.promise = null;
    return KEY_NAMES.data;
  });
  return KEY_NAMES.promise;
}

function keyName(kind, key) {
  var names = KEY_NAMES.data && KEY_NAMES.data[kind];
  var name = names && names[key] ? i18nText(names[key]) : '';
  return name || key || '';
}

function withKeyNames(build) {
  var box = h('div', { class: 'contents' });
  var paint = function () { clear(box); append(box, build()); };
  paint();
  if (!KEY_NAMES.data || Date.now() - KEY_NAMES.at >= 30000) keyNames().then(function () { if (box.isConnected || box.parentNode) paint(); });
  return box;
}

function priceText(it) {
  var parts = [];
  if (it.priceToman !== null && it.priceToman !== undefined) parts.push(t('toman', { n: it.priceToman }));
  if (it.coinPrice && it.coinPrice.amount !== undefined && it.coinPrice.amount !== null) parts.push(t('amount.of', { n: fmtNum(it.coinPrice.amount), name: keyName('currency', it.coinPrice.currencyKey) }));
  return parts;
}

function itemSummary(items) {
  return (items || []).map(function (r) {
    if (r.type === 'storage') return r.key;
    if (r.type === 'item') return t('item.item') + ': ' + r.key;
    if (r.type === 'achievement') return t('item.achievement') + ': ' + r.key;
    var sign = r.action === 'add' && Number(r.value) > 0 ? '+' : '';
    return t('amount.of', { n: (r.action === 'set' ? '= ' : sign) + fmtNum(r.value), name: keyName(r.type === 'meta' ? 'meta' : 'currency', r.key) });
  });
}

function languageCell(it) {
  var pic = h('div', { class: 'h-8 w-8 shrink-0 rounded-md bg-secondary-100 dark:bg-secondary-700 overflow-hidden flex items-center justify-center text-secondary-400' });
  pic.appendChild(it.iconUrl ? img(it.iconUrl, 'languages') : icon('languages', 'h-4 w-4'));
  var trName = i18nText(it.name);
  return h('div', { class: 'flex items-center gap-3 min-w-0' }, pic,
    h('div', { class: 'min-w-0' },
      h('div', { class: 'font-medium truncate flex items-center gap-2' }, h('span', { dir: it.direction === 'rtl' ? 'rtl' : 'ltr', text: it.nativeName || it.code }), it.isDefault ? tag(t('lang.default'), 'primary') : null),
      h('div', { class: 'text-secondary-500 flex items-center gap-2' }, mono(it.code), trName && trName !== it.nativeName ? h('span', { class: 'text-xs truncate', text: trName }) : null)));
}

function languageBody(it, patch) {
  return Object.assign({ name: it.name || {}, nativeName: it.nativeName, direction: it.direction, iconUrl: it.iconUrl || '', isActive: !!it.isActive, isDefault: !!it.isDefault, sort: it.sort || 0 }, patch || {});
}

function languageActive(it) {
  var box = h('div', { class: 'flex items-center gap-2' });
  var sw = toggle(it.isActive, null, function (on) {
    baseApi('PUT', '/languages/' + encodeURIComponent(it.code), languageBody(it, { isActive: on })).then(function () {
      it.isActive = on;
      label.textContent = on ? t('active') : t('inactive');
      toast('success', t(on ? 'lang.activated' : 'lang.deactivated', { name: it.nativeName || it.code }));
      languagesChanged();
    }, function (err) {
      sw.set(!on);
      toast('error', errorText(err));
    });
  }, 'sm');
  sw.setAttribute('aria-label', t('f.isActive'));
  sw.setAttribute('title', t('f.isActive'));
  var label = h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: it.isActive ? t('active') : t('inactive') });
  box.appendChild(sw);
  box.appendChild(label);
  return box;
}

var languageHooks = [];
function languagesChanged() {
  loadLanguages(true).then(function () { languageHooks.forEach(function (fn) { fn(); }); });
}

var RESOURCES = {
  currencies: {
    filters: [{ name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'currency' } }],
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '60%' },
      { label: 'col.type', render: function (it) { return tag(t('type.' + it.type), it.type === 'hard' ? 'warning' : 'primary'); }, width: 110 },
      { label: 'col.category', render: taxonomyCell, skeleton: '30%' },
      { label: 'col.default', render: function (it) { return numCell(it.defaultValue); }, width: 130 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, keyRe: CURRENCY_KEY_RE, lockOnEdit: true },
        { name: 'type', type: 'select', label: 'f.type', required: true, options: opts('type.', ['soft', 'hard']) }
      ] },
      tr('name', { required: true, max: 128 }),
      { type: 'row', fields: [categoryField('currency'), tagsField('currency')] },
      { type: 'row', fields: [
        { name: 'defaultValue', type: 'number', label: 'f.defaultValue', hint: 'f.defaultValue.currency.hint', min: 0 },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function () { return { type: 'soft', name: {}, categoryId: null, tagIds: [], defaultValue: 0, sort: 0, metadata: {} }; }
  },
  metas: {
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '55%' },
      { label: 'col.default', render: function (it) { return numCell(it.defaultValue); }, width: 140 },
      { label: 'col.type', render: function (it) { return it.serverOnly ? tag(t('server.only'), 'warning') : ''; }, width: 130 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      tr('name', { required: true, max: 128 }),
      { name: 'defaultValue', type: 'number', int: false, label: 'f.defaultValue', hint: 'f.defaultValue.meta.hint' },
      { name: 'serverOnly', type: 'bool', inline: true, label: 'f.serverOnly', hint: 'f.serverOnly.hint' },
      { name: 'settings', type: 'json', label: 'f.settings', optional: true, empty: {} }
    ],
    defaults: function () { return { name: {}, defaultValue: 0, serverOnly: false, settings: {} }; }
  },
  prizes: {
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '55%' },
      { label: 'col.items', render: function (it) {
        return h('div', { class: 'flex flex-wrap gap-1' }, withKeyNames(function () {
          var s = itemSummary(it.items);
          return [s.slice(0, 3).map(function (x) { return tag(x, 'secondary'); }), s.length > 3 ? tag('+' + fmtNum(s.length - 3), 'secondary') : null];
        }));
      }, skeleton: '40%' },
      { label: 'col.type', render: function (it) { return tag(it.isRepeatable ? t('repeatable') : t('once'), it.isRepeatable ? 'primary' : 'secondary'); }, width: 120 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
      tr('name', { required: true, max: 128 }),
      tr('message', { multiline: true, max: 1000, optional: true }),
      { type: 'row', fields: [categoryField('prize'), tagsField('prize')] },
      { name: 'isRepeatable', type: 'bool', inline: true, label: 'f.isRepeatable', hint: 'f.isRepeatable.hint' },
      { name: 'items', type: 'items', label: 'f.items' },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function () { return { name: {}, message: {}, tagIds: [], isRepeatable: false, items: [{ type: 'currency', key: '', action: 'add', value: 0 }], metadata: {} }; }
  },
  products: {
    filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '60%' },
      { label: 'col.price', render: function (it) { return h('div', { class: 'flex flex-col tabular-nums whitespace-nowrap' }, withKeyNames(function () { return priceText(it).map(function (p) { return h('span', { text: p }); }); }), it.discountPercent ? h('span', { class: 'text-xs text-success-600', text: t('percent', { n: it.discountPercent }) }) : null); }, width: 160 },
      { label: 'col.methods', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, (it.paymentMethods || []).map(function (m) { return tag(t('method.' + m), 'secondary'); }), (it.audienceTagIds || []).length ? tag(t('audience.only'), 'primary') : null); }, skeleton: '45%' },
      { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 80 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
        { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
      ] },
      tr('name', { required: true, max: 128 }),
      tr('description', { multiline: true, max: 4000, optional: true }),
      { type: 'row', fields: [categoryField('product'), tagsField('product')] },
      { name: 'audienceTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.audience', hint: 'f.audience.hint', optional: true, empty: [] },
      { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', hint: 'f.prize.product.hint', required: true },
      { type: 'heading', label: 'f.paymentMethods' },
      { name: 'paymentMethods', type: 'multi', required: true, options: opts('method.', METHODS) },
      { type: 'row', fields: [
        { name: 'priceToman', type: 'number', label: 'f.priceToman', hint: 'f.priceToman.hint', min: 0,
          when: function (v) { return hasAny(v.paymentMethods, GATEWAYS.concat(STORES)); },
          check: function (n, v) { return hasAny(v.paymentMethods, GATEWAYS) && (n === null || n < 100) ? t('f.priceToman.hint') : null; } },
        { name: 'discountPercent', type: 'number', label: 'f.discountPercent', min: 0, max: 100, empty: 0,
          when: function (v) { return (v.paymentMethods || []).length > 0; } }
      ] },
      { type: 'row', when: function (v) { return hasAny(v.paymentMethods, ['in_app']); }, fields: [
        { name: 'coinPrice.currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.coinPrice.currency', required: true, when: function (v) { return hasAny(v.paymentMethods, ['in_app']); } },
        { name: 'coinPrice.amount', type: 'number', label: 'f.coinPrice.amount', required: true, min: 1, when: function (v) { return hasAny(v.paymentMethods, ['in_app']); } }
      ] },
      { type: 'row', when: function (v) { return hasAny(v.paymentMethods, STORES); }, fields: [
        { name: 'sku.bazaar', type: 'text', ltr: true, label: 'f.sku.bazaar', required: true, when: function (v) { return hasAny(v.paymentMethods, ['bazaar']); } },
        { name: 'sku.myket', type: 'text', ltr: true, label: 'f.sku.myket', required: true, when: function (v) { return hasAny(v.paymentMethods, ['myket']); } }
      ] },
      { name: 'isSubscription', type: 'bool', inline: true, label: 'f.isSubscription', hint: 'f.isSubscription.hint', empty: false, when: function (v) { return hasAny(v.paymentMethods, ['bazaar']); } },
      { type: 'heading', label: 'f.icon' },
      { type: 'row', fields: [
        { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
        { name: 'coverUrl', type: 'image', label: 'f.cover', optional: true }
      ] },
      { type: 'row', fields: [
        { name: 'minLevel', type: 'number', label: 'f.minLevel', min: 0, empty: 0 },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function () { return { isActive: true, name: {}, description: {}, tagIds: [], audienceTagIds: [], paymentMethods: ['zarinpal'], discountPercent: 0, sku: {}, coinPrice: null, isSubscription: false, minLevel: 0, sort: 0, metadata: {} }; },
    fixPayload: function (p) {
      if (p.coinPrice && p.coinPrice.currencyKey == null && p.coinPrice.amount == null) p.coinPrice = null;
      var sku = {};
      if (p.sku && p.sku.bazaar) sku.bazaar = p.sku.bazaar;
      if (p.sku && p.sku.myket) sku.myket = p.sku.myket;
      p.sku = sku;
      return p;
    }
  },
  gifts: {
    filters: [
      { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'gift' } },
      { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }
    ],
    columns: [
      { label: 'col.name', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) }), it.code ? h('div', { class: 'text-secondary-500' }, mono(it.code)) : null); }, skeleton: '50%' },
      { label: 'col.type', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('gift.' + it.type), it.type === 'discount' ? 'warning' : 'primary'), it.type === 'discount' && it.discountPercent ? tag(t('percent', { n: it.discountPercent }), 'secondary') : null); }, width: 150 },
      { label: 'col.category', render: taxonomyCell, skeleton: '30%' },
      { label: 'col.usage', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: fmtNum(it.totalUsed || 0) + ' / ' + (it.maxTotalUses ? fmtNum(it.maxTotalUses) : t('unlimited')) }); }, width: 130 },
      { label: 'col.expire', render: function (it) { return dateCell(it.expireAt); }, width: 190 },
      { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
    ],
    fields: [
      tr('name', { label: 'f.gift.name', required: true, max: 128 }),
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' },
      { type: 'row', fields: [categoryField('gift'), tagsField('gift')] },
      { type: 'row', fields: [
        { name: 'type', type: 'select', label: 'f.type', required: true, options: opts('gift.', ['reward', 'discount']) },
        { name: 'code', type: 'code', label: 'f.code', hint: 'f.code.hint', optional: true, check: function (c, v) { return v.type === 'discount' && !c ? t('code.CODE_REQUIRED') : null; } }
      ] },
      { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', required: true, when: function (v) { return v.type === 'reward'; } },
      { name: 'discountPercent', type: 'number', label: 'f.discountPercent', required: true, min: 1, max: 100, when: function (v) { return v.type === 'discount'; } },
      { type: 'row', fields: [
        { name: 'perUserLimit', type: 'number', label: 'f.perUserLimit', required: true, min: 1 },
        { name: 'maxTotalUses', type: 'number', label: 'f.maxTotalUses', hint: 'f.maxTotalUses.hint', min: 1, optional: true }
      ] },
      { type: 'row', fields: [
        { name: 'validFrom', type: 'datetime', label: 'f.validFrom', optional: true },
        { name: 'expireAt', type: 'datetime', label: 'f.expireAt', optional: true }
      ] },
      { name: 'isPromo', type: 'bool', inline: true, label: 'f.isPromo' },
      { name: 'totalUsed', type: 'static', label: 'f.totalUsed', readonly: true, when: function (v) { return !!v.id; }, render: function (v) { return numCell(v.totalUsed || 0); } }
    ],
    defaults: function (query) { return { name: {}, type: 'reward', isActive: true, isPromo: false, perUserLimit: 1, maxTotalUses: null, code: '', categoryId: (query && query.categoryId) || null, tagIds: [] }; },
    fixPayload: function (p) { if (p.code === '') p.code = null; return p; }
  },
  categories: {
    api: 'base',
    filters: [{ name: 'keyType', label: 'filter.keyType', options: keyTypeOptions('categories') }],
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '55%' },
      { label: 'col.keyType', render: function (it) { return tag(keyTypeLabel('categories', it.keyType), 'secondary'); }, width: 140 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
        { name: 'keyType', type: 'select', label: 'f.keyType', hint: 'f.keyType.category.hint', required: true, options: keyTypeOptions('categories') }
      ] },
      tr('name', { required: true, max: 128 }),
      tr('description', { multiline: true, max: 1000, optional: true }),
      { type: 'row', fields: [
        { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function (query) { return { keyType: (query && query.keyType) || 'product', name: {}, description: {}, iconUrl: '', metadata: {}, sort: 0 }; },
    fixPayload: function (p) { if (!p.iconUrl) p.iconUrl = ''; if (p.sort == null) p.sort = 0; return p; }
  },
  tags: {
    api: 'base',
    filters: [{ name: 'keyType', label: 'filter.keyType', options: keyTypeOptions('tags') }],
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '55%' },
      { label: 'col.keyType', render: function (it) {
        return h('div', { class: 'flex flex-wrap gap-1' }, tag(keyTypeLabel('tags', it.keyType), it.keyType === 'general' ? 'primary' : 'secondary'),
          it.keyType === 'user' && it.isPublic ? tag(t('tag.public'), 'success') : null,
          it.keyType === 'user' && it.players ? h('a', { href: hashFor('players', null, { tagId: it.id }), class: 'eco-link text-xs whitespace-nowrap', text: t('tag.playersN', { n: it.players }) }) : null);
      }, width: 200 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
        { name: 'keyType', type: 'select', label: 'f.keyType', hint: 'f.keyType.tag.hint', required: true, options: keyTypeOptions('tags') }
      ] },
      tr('name', { required: true, max: 128 }),
      tr('description', { multiline: true, max: 1000, optional: true }),
      { name: 'isPublic', type: 'bool', inline: true, label: 'f.tag.public', hint: 'f.tag.public.hint', empty: false, when: function (v) { return v.keyType === 'user'; } },
      { type: 'row', fields: [
        { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function (query) { return { isPublic: false, keyType: (query && query.keyType) || 'general', name: {}, description: {}, iconUrl: '', metadata: {}, sort: 0 }; },
    fixPayload: function (p) { if (!p.iconUrl) p.iconUrl = ''; if (p.sort == null) p.sort = 0; return p; }
  },
  languages: {
    api: 'base',
    filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
    columns: [
      { label: 'col.language', render: languageCell, skeleton: '60%' },
      { label: 'col.status', render: languageActive, width: 150 },
      { label: 'col.direction', render: function (it) { return tag(t('dir.' + it.direction), 'secondary'); }, width: 150 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'code', type: 'key', label: 'f.lang.code', hint: 'f.lang.code.hint', required: true, keyRe: LANG_CODE_RE, keyErr: 'code.INVALID_LANGUAGE_CODE', lockOnEdit: true },
        { name: 'nativeName', label: 'f.lang.nativeName', hint: 'f.lang.nativeName.hint', required: true, check: function (v) { return v && Array.from(v.trim()).length > 64 ? t('code.INVALID_NATIVE_NAME') : null; } }
      ] },
      tr('name', { label: 'f.lang.name', hint: 'f.lang.name.hint', required: true, max: 64 }),
      { type: 'row', fields: [
        { name: 'direction', type: 'select', label: 'f.lang.direction', required: true, options: opts('dir.', ['rtl', 'ltr']) },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { type: 'row', fields: [
        { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.lang.active.hint' },
        { name: 'isDefault', type: 'bool', inline: true, label: 'f.lang.default', hint: 'f.lang.default.hint' }
      ] }
    ],
    defaults: function () { return { code: '', name: {}, nativeName: '', direction: 'ltr', iconUrl: '', isActive: true, isDefault: false, sort: 0 }; },
    fixPayload: function (p) { p.code = String(p.code || '').trim().toLowerCase().replace(/_/g, '-'); if (!p.iconUrl) p.iconUrl = ''; if (p.sort == null) p.sort = 0; return p; },
    afterSave: function () { languagesChanged(); }
  },
  faqs: {
    api: 'publication',
    filters: [
      { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'faq' } },
      { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }
    ],
    columns: [
      { label: 'col.question', render: function (it) {
        return h('div', { class: 'min-w-0' },
          h('div', { class: 'font-medium eco-clamp', text: i18nText(it.question) }),
          h('div', { class: 'text-xs text-secondary-500 eco-clamp', text: mdPlain(i18nText(it.answer)) }));
      }, skeleton: '70%' },
      { label: 'col.category', render: function (it) { return it.categoryName ? tag(i18nText(it.categoryName), 'secondary') : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 150 },
      { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 80 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      tr('question', { required: true, max: 500 }),
      tr('answer', { markdown: true, required: true, max: 10000 }),
      { type: 'row', fields: [
        categoryField('faq'),
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.faq.active.hint' }
    ],
    defaults: function (query) { return { question: {}, answer: {}, categoryId: (query && query.categoryId) || null, isActive: true, sort: 0 }; }
  },
  news: {
    api: 'publication',
    filters: [
      { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'news' } },
      { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } },
      { name: 'isPinned', label: 'filter.pinned', options: function () { return [{ value: 'true', label: t('news.pinned') }, { value: 'false', label: t('news.notPinned') }]; } }
    ],
    columns: [
      { label: 'col.title', render: function (it) {
        var pic = h('div', { class: 'h-10 w-16 shrink-0 rounded-md bg-secondary-100 dark:bg-secondary-700 overflow-hidden flex items-center justify-center text-secondary-400' });
        pic.appendChild(it.coverUrl ? img(it.coverUrl) : icon('image', 'h-4 w-4'));
        return h('div', { class: 'flex items-center gap-3 min-w-0' }, pic,
          h('div', { class: 'min-w-0' },
            h('div', { class: 'font-medium flex items-center gap-1.5 min-w-0' }, it.isPinned ? h('span', { class: 'eco-text-warning', title: t('news.pinned') }, icon('pin', 'h-3.5 w-3.5')) : null, h('span', { class: 'eco-clamp', text: i18nText(it.title) })),
            i18nText(it.summary) ? h('div', { class: 'text-xs text-secondary-500 eco-clamp', text: i18nText(it.summary) }) : null));
      }, skeleton: '70%' },
      { label: 'col.category', render: function (it) { return it.categoryName ? tag(i18nText(it.categoryName), 'secondary') : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 130 },
      { label: 'col.comments', render: function (it) {
        if (!it.allowComment) return tag(t('news.commentsOff'), 'secondary');
        return h('div', { class: 'flex flex-wrap gap-1' },
          h('span', { class: 'tabular-nums text-sm', text: fmtNum(it.commentCount || 0) }),
          it.pendingComments ? h('a', { href: hashFor('comments', null, { status: 'pending', newsId: it.id }), class: 'eco-tag eco-tag--warning inline-flex leading-tight px-2 py-0.5 text-xs rounded-md whitespace-nowrap', text: t('cm.pendingN', { n: it.pendingComments }) }) : null);
      }, width: 130 },
      { label: 'col.published', render: function (it) {
        var future = it.publishedAt && new Date(it.publishedAt).getTime() > Date.now();
        return h('div', { class: 'flex flex-col items-start gap-0.5' }, dateCell(it.publishedAt), future ? tag(t('news.scheduled'), 'primary') : null);
      }, width: 190 },
      { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
    ],
    fields: [
      tr('title', { required: true, max: 200 }),
      tr('summary', { multiline: true, max: 1000, optional: true }),
      tr('body', { markdown: true, required: true, max: 50000, rows: 8 }),
      { name: 'coverUrl', type: 'image', label: 'f.cover', optional: true },
      { type: 'row', fields: [
        categoryField('news'),
        { name: 'publishedAt', type: 'datetime', label: 'f.publishedAt', hint: 'f.publishedAt.hint', optional: true }
      ] },
      { type: 'row', fields: [
        { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.news.active.hint' },
        { name: 'isPinned', type: 'bool', inline: true, label: 'f.isPinned', hint: 'f.isPinned.hint' }
      ] },
      { type: 'heading', label: 'f.comments' },
      { type: 'row', fields: [
        { name: 'allowComment', type: 'bool', inline: true, label: 'f.allowComment', hint: 'f.allowComment.hint' },
        { name: 'freeComment', type: 'bool', inline: true, label: 'f.freeComment', hint: 'f.freeComment.hint', empty: false, when: function (v) { return !!v.allowComment; } }
      ] },
      { type: 'heading', label: 'f.actions' },
      { name: 'actions', type: 'actions', empty: [] },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function (query) { return { title: {}, summary: {}, body: {}, coverUrl: '', categoryId: (query && query.categoryId) || null, publishedAt: null, isActive: true, isPinned: false, allowComment: true, freeComment: false, actions: [], metadata: {} }; },
    fixPayload: function (p) { if (!p.coverUrl) p.coverUrl = ''; if (!p.publishedAt) delete p.publishedAt; return p; }
  }
};

var SECTIONS = {
  economy: ['overview', 'currencies', 'metas', 'prizes', 'products', 'items', 'achievements', 'gifts', 'event-types', 'event-rules', 'meta-rules', 'level-tables', 'streaks', 'tasks', 'task-groups', 'event-log', 'referrals', 'payments'],
  publication: ['faqs', 'news', 'comments', 'notifications', 'remote-config'],
  moderation: ['players', 'reports', 'restrictions', 'chat-words', 'activity'],
  support: ['tickets', 'live-chat']
};
var SECTION_ORDER = ['economy', 'publication', 'moderation', 'support'];
var SECTION_LABEL = { economy: 'section', publication: 'section.pub', moderation: 'section.mod', support: 'section.sup' };
var PAGE_ACL = { players: 'ACCOUNT', restrictions: 'ACCOUNT', reports: 'CHANNEL_MESSAGE', 'chat-words': 'CHANNEL_MESSAGE', activity: 'ACCOUNT', tickets: 'ACCOUNT_NOTES', 'live-chat': 'ACCOUNT_NOTES', 'remote-config': 'SETTINGS' };
var SECTION_ACL = { economy: 'HIRO_ECONOMY', publication: 'NOTIFICATION' };
var LIMIT = 20;
var lastQuery = { comments: { status: 'pending' }, reports: { status: 'open' }, tickets: { status: 'waiting' } };

function sectionOf(page) {
  if (psSection(page)) return PS_ROUTE;
  if (SECTIONS.support.indexOf(page) !== -1) return 'support';
  if (SECTIONS.moderation.indexOf(page) !== -1) return 'moderation';
  return SECTIONS.publication.indexOf(page) !== -1 ? 'publication' : 'economy';
}

function apiFor(page) {
  var def = RESOURCES[page] || LISTS[page];
  if (def && def.api === 'base') return baseApi;
  if (def && def.api === 'publication') return pubApi;
  var sec = sectionOf(page);
  return sec === 'publication' ? pubApi : api;
}

function hashFor(page, id, query) {
  return '#/' + sectionOf(page) + '/' + page + (id ? '/' + encodeURIComponent(id) : '') + qs(query || {});
}

function go(page, id, query) {
  location.hash = hashFor(page, id, query);
}

function errorBlock(err, retry) {
  return h('div', { class: 'py-6' }, notice('error', errorText(err), retry ? button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: retry }) : null));
}

function listPage(container, name, query, ctx) {
  var def = RESOURCES[name] || LISTS[name];
  var call = apiFor(name);
  var columns = def.columns.map(function (c) { return { label: t(c.label), width: c.width, align: c.align, skeleton: c.skeleton }; });
  var state = { q: query.q || '', page: Math.max(1, parseInt(query.page, 10) || 1), limit: LIMIT, filters: {}, total: 0, loading: true, loaded: false };
  (def.filters || []).forEach(function (f) { if (query[f.name]) state.filters[f.name] = query[f.name]; });
  lastQuery[name] = currentQuery();

  var create = def.fields ? button({ variant: 'primary', size: 'page', icon: 'plus', label: t('new.' + name), onclick: function () { go(name, 'new', currentQuery()); } }) : def.create ? def.create() : null;
  var frame = def.settings || sectionOf(name) === PS_ROUTE ? tabFrame(t('desc.' + name), create) : pageFrame(t('nav.' + name), t('desc.' + name), create);
  var pagerTop = h('div', { class: 'relative flex items-center justify-end' });
  var pagerBottom = h('div', { class: 'flex gap-2 py-4 items-center justify-end' });
  var tableBox = h('div', {});
  var search = searchInput(state.q, t(has('search.' + name) ? 'search.' + name : 'search'), debounce(function (e) {
    state.q = toLatinDigits(e.target.value).trim();
    state.page = 1;
    sync();
    load();
  }, 300));
  var filters = (def.filters || []).map(function (f) {
    var opts2 = typeof f.options === 'function' ? f.options() : f.options;
    if (f.type === 'ref') {
      var picker = refPicker({ resource: f.resource, api: typeof f.api === 'function' ? f.api : f.api === 'publication' ? pubApi : api, filter: f.filter, value: state.filters[f.name] || '', label: t(f.label), placeholder: t(f.label) + ': ' + t('all'), onchange: function (v) {
        state.filters[f.name] = v || '';
        state.page = 1;
        sync();
        load();
      } });
      return h('div', { class: 'x-label x-input eco-filter-ref', style: INPUT_VARS }, picker);
    }
    if (f.type === 'text') {
      var inp = h('input', { class: INPUT_SM + ' font-mono eco-filter-text', dir: 'ltr', 'data-ltr': '', 'data-ph-rtl': /[\u0600-\u06FF]/.test(t(f.label)) ? '' : null, type: 'text', value: state.filters[f.name] || '', placeholder: t(f.label), 'aria-label': t(f.label) });
      inp.addEventListener('input', debounce(function () { state.filters[f.name] = toLatinDigits(inp.value).trim(); state.page = 1; sync(); load(); }, 300));
      return h('label', { class: 'x-label block relative x-input', style: INPUT_VARS }, inp);
    }
    return filterSelect(t(f.label), state.filters[f.name], opts2, function (v) { state.filters[f.name] = v; state.page = 1; sync(); load(); });
  });
  if (def.top) frame.content.appendChild(def.top());
  frame.content.appendChild(h('div', {},
    h('div', { class: 'flex flex-col items-start pb-2 gap-2' },
      filters.length > 3 ? h('div', { class: 'flex flex-col gap-2 w-full' },
        h('div', { class: 'flex flex-col lg:flex-row lg:items-center gap-2 w-full' },
          search,
          h('div', { class: 'flex-grow hidden lg:block' }),
          pagerTop),
        h('div', { class: 'eco-filters eco-filters-wrap flex flex-wrap items-center gap-2' }, filters))
      : h('div', { class: 'flex flex-col lg:flex-row lg:items-center gap-2 w-full' },
        search,
        filters.length ? h('div', { class: 'eco-filters flex flex-wrap items-center gap-2' }, filters) : null,
        h('div', { class: 'flex-grow hidden lg:block' }),
        pagerTop)),
    tableBox,
    pagerBottom));
  clear(container).appendChild(frame.el);

  function currentQuery() {
    var q = Object.assign({}, state.filters);
    if (state.q) q.q = state.q;
    if (state.page > 1) q.page = state.page;
    return q;
  }
  function sync() {
    lastQuery[name] = currentQuery();
    ctx.query = lastQuery[name];
    replaceHash(hashFor(name, ctx.id, lastQuery[name]));
  }
  function paintPager() {
    clear(pagerTop).appendChild(pager(state, function (p) { state.page = p; sync(); load(); }, function () { load(); }));
    clear(pagerBottom).appendChild(pager(state, function (p) { state.page = p; sync(); load(); window.scrollTo(0, 0); }));
  }
  function open(item) {
    if (def.rowHref) { location.hash = def.rowHref(item); return; }
    go(name, item.id, currentQuery());
  }
  var seq = 0;
  function load() {
    var my = ++seq;
    state.loading = true;
    if (!state.loaded) {
      clear(tableBox).appendChild(table(columns, skeletonRows(def.columns, 8)));
    } else {
      tableBox.classList.add('opacity-60');
    }
    paintPager();
    var params = Object.assign({ q: state.q, page: state.page, limit: state.limit }, state.filters);
    return call('GET', '/' + (def.path || name) + qs(params)).then(function (res) {
      if (my !== seq) return;
      state.loading = false;
      state.loaded = true;
      state.total = (res && res.total) || 0;
      var items = (res && res.items) || [];
      if (def.onItems) def.onItems(items);
      tableBox.classList.remove('opacity-60');
      if (!items.length && state.page > 1 && state.total > 0) {
        state.page = Math.max(1, Math.ceil(state.total / state.limit));
        sync();
        return load();
      }
      clear(tableBox);
      if (!items.length) {
        var filtered = state.q || Object.keys(state.filters).some(function (k) { return state.filters[k]; });
        tableBox.appendChild(card(emptyBlock(filtered ? t('empty.search') : t('empty.title'), filtered ? null : t('empty.hint.' + name),
          !filtered && def.fields ? button({ variant: 'primary', size: 'page', icon: 'plus', label: t('new.' + name), onclick: function () { go(name, 'new'); } }) : !filtered && def.create ? def.create() : null)));
      } else {
        tableBox.appendChild(table(columns, items.map(function (it) {
          return { item: it, cells: def.columns.map(function (c) { return c.render(it); }) };
        }), { onrow: def.noRow ? null : open }));
      }
      paintPager();
    }, function (err) {
      if (my !== seq) return;
      state.loading = false;
      tableBox.classList.remove('opacity-60');
      clear(tableBox).appendChild(errorBlock(err, load));
      paintPager();
    });
  }
  load();
  return { reload: load, name: name };
}

function formSkeleton(def) {
  var box = h('div', {});
  var count = 0;
  def.fields.forEach(function walk(f) {
    if (count > 7) return;
    if (f.type === 'row') { box.appendChild(h('div', { class: 'eco-form-row' }, f.fields.map(function () { return skeletonField(); }))); count += 1; return; }
    if (f.type === 'heading') return;
    box.appendChild(skeletonField(f.type === 'textarea' || f.type === 'json' || f.type === 'items' ? 'h-24' : 'h-10'));
    count += 1;
  });
  return box;
}

function skeletonField(height) {
  return h('div', { class: 'mb-5' }, bar('7rem', 'h-3 mb-2'), bar('100%', height || 'h-10'));
}

var panelState = null;

function closePanel() {
  if (panelState) {
    panelState.el.remove();
    document.removeEventListener('keydown', panelState.esc, true);
    panelState = null;
  }
}

function openEditor(name, id, query, onSaved) {
  var def = RESOURCES[name];
  if (!def) return;
  if (panelState && panelState.key === name + '/' + id) return;
  closePanel();
  var isNew = id === 'new';
  var title = t((isNew ? 'new.' : 'edit.') + name);
  var body = h('div', {});
  var footer = h('div', { class: 'flex gap-4 w-full items-center' });
  var back = function () { go(name, null, lastQuery[name] || {}); };
  var el = sidePanel(title, body, footer, back);
  var esc = function (e) { if (e.key === 'Escape' && panelState && !document.querySelector('[data-economy="panel"] .eco-pop:not([hidden]), [data-economy="panel"] .eco-qc-pop:not([hidden])')) { e.preventDefault(); back(); } };
  document.addEventListener('keydown', esc, true);
  panelState = { el: el, key: name + '/' + id, esc: esc };
  document.body.appendChild(el);
  body.appendChild(formSkeleton(def));

  var call = apiFor(name);
  var item = isNew ? Promise.resolve(def.defaults(query)) : call('GET', '/' + (def.path || name) + '/' + encodeURIComponent(id));
  Promise.all([item, freshLanguages(), loadKeyTypes(), def.prepare ? def.prepare() : null]).then(function (all) { return all[0]; }).then(function (data) {
    if (!panelState || panelState.el !== el) return;
    var values = clone(data) || {};
    if (def.toForm) values = def.toForm(values);
    var form = buildForm(def.fields, values);
    var top = h('div', {});
    clear(body).appendChild(h('form', { novalidate: true, onsubmit: function (e) { e.preventDefault(); save(); } }, top, form.el, h('button', { type: 'submit', hidden: true })));
    var saveBtn = button({ variant: 'primary', size: 'footer', type: 'button', label: isNew ? t('create') : t('save'), onclick: function () { save(); } });
    var cancelBtn = button({ variant: 'ghost', size: 'footer', label: t('cancel'), onclick: back });
    var del = null;
    if (!isNew) {
      del = h('div', { class: 'flex items-center gap-2' });
      var askDelete = function () {
        clear(del).appendChild(h('span', { class: 'text-sm text-secondary-700 dark:text-secondary-200', text: t('delete.confirm') }));
        del.appendChild(button({ variant: 'dangerSolid', size: 'sm', label: t('delete.yes'), onclick: doDelete }));
        del.appendChild(button({ variant: 'ghost', size: 'sm', label: t('no'), onclick: resetDelete }));
      };
      var resetDelete = function () { clear(del).appendChild(button({ variant: 'danger', size: 'footer', icon: 'trash', label: t('delete'), onclick: askDelete })); };
      var doDelete = function () {
        call('DELETE', '/' + (def.path || name) + '/' + encodeURIComponent(id)).then(function () {
          toast('success', t('deleted'));
          if (def.afterSave) def.afterSave();
          back();
          if (onSaved) onSaved();
        }, function (err) {
          resetDelete();
          clear(top).appendChild(h('div', { class: 'mb-5' }, notice('error', errorText(err))));
          top.scrollIntoView({ block: 'start' });
        });
      };
      resetDelete();
    }
    clear(footer).appendChild(h('div', { class: 'flex-1 flex items-center' }, del));
    footer.appendChild(cancelBtn);
    footer.appendChild(saveBtn);

    function save() {
      clear(top);
      if (!form.validate()) return;
      var payload = form.payload();
      if (def.fixPayload) payload = def.fixPayload(payload);
      setBusy(saveBtn, true, t('saving'));
      call(isNew ? 'POST' : 'PUT', '/' + (def.path || name) + (isNew ? '' : '/' + encodeURIComponent(id)), payload).then(function () {
        toast('success', t('saved'));
        if (def.afterSave) def.afterSave();
        back();
        if (onSaved) onSaved();
      }, function (err) {
        setBusy(saveBtn, false);
        if (!form.showServerError(err)) {
          clear(top).appendChild(h('div', { class: 'mb-5' }, notice('error', errorText(err))));
          top.scrollIntoView({ block: 'start' });
        }
      });
    }
    var firstInput = body.querySelector('input,textarea,select');
    if (firstInput) firstInput.focus();
  }, function (err) {
    if (!panelState || panelState.el !== el) return;
    clear(body).appendChild(errorBlock(err, function () { closePanel(); openEditor(name, id, query, onSaved); }));
  });
}

var PAYMENT_LIST = {
  filters: [
    { name: 'status', label: 'filter.status', options: opts('status.', ['pending', 'success', 'failed', 'expired']) },
    { name: 'method', label: 'filter.method', options: opts('method.', METHODS) },
    { name: 'userId', label: 'filter.user', type: 'text' }
  ],
  columns: [
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 170 },
    { label: 'col.user', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: it.username || '' }), h('div', { class: 'text-secondary-500 truncate' }, mono(it.userId))); }, skeleton: '50%' },
    { label: 'col.product', render: function (it) { var p = it.product || {}; return h('div', { class: 'min-w-0' }, h('div', { class: 'truncate', text: i18nText(p.name || it.productName) || '' }), h('div', { class: 'text-secondary-500' }, mono(p.key || it.productKey || ''))); }, skeleton: '45%' },
    { label: 'col.method', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('method.' + it.method), 'secondary'), it.sandbox ? tag(t('pay.sandbox'), 'warning') : null); }, width: 150 },
    { label: 'col.amount', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: it.method === 'in_app' ? fmtNum(it.coinAmount) + ' ' + (it.coinCurrency || '') : t('toman', { n: it.finalToman || 0 }) }); }, width: 150 },
    { label: 'col.status', render: function (it) { return paymentStatus(it); }, width: 150 }
  ]
};

function paymentStatus(it) {
  return h('div', { class: 'flex flex-wrap gap-1' },
    tag(t('status.' + it.status), STATUS_COLOR[it.status]),
    it.status === 'success' && !it.rewardGranted ? tag(t('pay.reward.notGranted'), 'error') : null);
}

function dl(rows) {
  return h('dl', { class: 'eco-dl' }, rows.filter(function (r) { return r && r[1] !== null && r[1] !== undefined && r[1] !== ''; }).map(function (r) {
    return [h('dt', { class: 'text-secondary-500 dark:text-secondary-400', text: t(r[0]) }), h('dd', { class: 'text-secondary-800 dark:text-secondary-200 min-w-0 break-words' }, r[1])];
  }));
}

function section(titleKey, body, actions) {
  return card(h('div', { class: 'p-4' },
    h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, h('h2', { class: 'text-lg font-medium', text: t(titleKey) }), actions || null),
    body));
}

function jsonBlock(v) {
  if (v === null || v === undefined || v === '') return null;
  var text = typeof v === 'string' ? v : JSON.stringify(v, null, 2);
  return h('pre', { class: 'eco-pre font-mono text-xs bg-secondary-50 dark:bg-secondary-900 border border-secondary-200 dark:border-secondary-700 rounded-md p-2 overflow-x-auto', dir: 'ltr', text: text });
}

function paymentDetail(container, id) {
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('payments', null, lastQuery.payments || {}); } });
  var frame = pageFrame(t('pay.title'), null, null);
  var titleBox = frame.el.querySelector('h1');
  frame.actions.appendChild(backBtn);
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-detail-grid' });
  frame.content.appendChild(grid);

  function skeleton() {
    clear(grid);
    [5, 6, 4, 4].forEach(function (n) {
      var rows = [];
      for (var i = 0; i < n; i++) rows.push([h('dt', {}, bar('6rem')), h('dd', {}, bar(['70%', '50%', '40%'][i % 3]))]);
      grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), h('dl', { class: 'eco-dl' }, rows))));
    });
    var logRows = [];
    for (var j = 0; j < 3; j++) logRows.push(h('div', { class: 'flex items-center gap-3 py-2' }, bar('5rem'), bar('40%'), bar('4rem')));
    grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), logRows), 'eco-span-all'));
  }

  function render(p) {
    clear(grid);
    clear(titleBox).appendChild(h('span', { text: t('pay.title') + ' ' }));
    titleBox.appendChild(h('span', { class: 'font-mono text-base text-secondary-500', dir: 'ltr', text: String(p.id || id).slice(0, 8) }));
    var user = p.user || {};
    var product = p.product || {};
    var gateway = GATEWAYS.indexOf(p.method) !== -1;
    var actions = h('div', { class: 'flex flex-wrap gap-2' });
    if (p.status === 'pending' && gateway) actions.appendChild(button({ variant: 'primary', size: 'page', icon: 'refresh', label: t('pay.inquire'), onclick: function (e) { act(e.currentTarget, 'inquire', 'pay.inquire.done'); } }));
    if (p.status === 'success' && (!p.rewardGranted || p.rewardError)) actions.appendChild(button({ variant: 'primary', size: 'page', icon: 'prizes', label: t('pay.retry'), onclick: function (e) { act(e.currentTarget, 'retry-grant', 'pay.retry.done'); } }));
    clear(frame.actions).appendChild(actions);
    frame.actions.appendChild(backBtn);

    grid.appendChild(section('col.status', dl([
      ['col.status', paymentStatus(p)],
      ['col.method', h('div', { class: 'flex flex-wrap gap-1' }, tag(t('method.' + p.method), 'secondary'), p.sandbox ? tag(t('pay.sandbox'), 'warning') : null)],
      ['pay.reward', tag(p.rewardGranted ? t('pay.reward.granted') : t('pay.reward.notGranted'), p.rewardGranted ? 'success' : 'secondary')],
      ['pay.reward.error', p.rewardError ? h('span', { class: 'text-error-600 dark:text-error-400', text: p.rewardError }) : null],
      ['pay.error', p.error ? h('span', { class: 'text-error-600 dark:text-error-400', text: p.error }) : null],
      ['pay.outcome', p.outcome ? mono(p.outcome) : null]
    ])));
    grid.appendChild(section('pay.amounts', dl([
      ['pay.original', p.originalToman != null ? h('span', { class: 'tabular-nums', text: t('toman', { n: p.originalToman }) }) : null],
      ['pay.discount', p.discountPercent ? h('span', { text: t('percent', { n: p.discountPercent }) }) : null],
      ['pay.final', p.finalToman != null ? h('span', { class: 'tabular-nums font-semibold', text: t('toman', { n: p.finalToman }) }) : null],
      ['pay.rial', p.amountRial ? h('span', { class: 'tabular-nums', text: fmtNum(p.amountRial) }) : null],
      ['pay.coin', p.coinAmount != null && p.coinCurrency ? h('span', { class: 'tabular-nums', text: fmtNum(p.coinAmount) + ' ' + p.coinCurrency }) : null]
    ])));
    grid.appendChild(section('pay.user', dl([
      ['col.user', h('a', { href: '#/players/' + encodeURIComponent(p.userId || user.id || ''), class: 'eco-link', text: user.username || p.username || p.userId || '' })],
      ['filter.user', mono(p.userId || user.id)],
      ['pay.product', h('div', {}, h('div', { text: i18nText(product.name) || '' }), product.key ? mono(product.key) : null)]
    ])));
    grid.appendChild(section('pay.gateway', dl([
      ['pay.authority', p.authority ? mono(p.authority) : null],
      ['pay.trackId', p.trackId ? mono(p.trackId) : null],
      ['pay.refId', p.refId ? mono(p.refId) : null],
      ['pay.storeToken', p.storeToken ? mono(p.storeToken) : null],
      ['pay.cardMask', p.cardMask ? mono(p.cardMask) : null],
      ['pay.inquiries', h('span', { class: 'tabular-nums', text: fmtNum(p.inquiryCount || 0) }) ],
      ['pay.createdAt', p.createdAt ? dateCell(p.createdAt) : null],
      ['pay.verifiedAt', p.verifiedAt ? dateCell(p.verifiedAt) : null],
      ['pay.inquiredAt', p.inquiredAt ? dateCell(p.inquiredAt) : null],
      ['pay.updatedAt', p.updatedAt ? dateCell(p.updatedAt) : null]
    ])));
    var logs = p.logs || [];
    var logBody = h('div', { class: 'flex flex-col' });
    if (!logs.length) logBody.appendChild(h('div', { class: 'text-sm text-secondary-500 py-2', text: t('pay.logs.empty') }));
    logs.forEach(function (l) {
      var ok = l.httpStatus && l.httpStatus < 400;
      var detail = h('div', { class: 'eco-log-detail mt-2', hidden: true },
        l.url ? h('div', { class: 'mb-2' }, mono(l.url)) : null,
        h('div', { class: 'eco-log-grid' },
          h('div', { class: 'min-w-0' }, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pay.request') }), jsonBlock(l.request) || h('div', { class: 'text-xs text-secondary-400', text: '…' })),
          h('div', { class: 'min-w-0' }, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pay.response') }), jsonBlock(l.response) || h('div', { class: 'text-xs text-secondary-400', text: '…' }))));
      var head = h('button', { type: 'button', class: 'w-full flex flex-wrap items-center gap-2 eco-start', 'aria-expanded': 'false', onclick: function () {
        detail.hidden = !detail.hidden;
        head.setAttribute('aria-expanded', detail.hidden ? 'false' : 'true');
        chev.classList.toggle('rotate-180', !detail.hidden);
      } });
      var chev = icon('chevron', 'h-4 w-4 text-secondary-500 transition');
      head.appendChild(tag(t('log.' + l.action), 'primary'));
      if (l.httpStatus) head.appendChild(tag(localDigits(String(l.httpStatus)), ok ? 'success' : 'error'));
      if (l.durationMs != null) head.appendChild(h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: t('pay.duration', { n: l.durationMs }) }));
      head.appendChild(h('span', { class: 'flex-1' }));
      head.appendChild(h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: fmtDate(l.createdAt) }));
      head.appendChild(chev);
      logBody.appendChild(h('div', { class: 'eco-log py-2 border-b border-secondary-200 dark:border-secondary-700' }, head, detail));
    });
    var sec = section('pay.logs', logBody);
    sec.classList.add('eco-span-all');
    grid.appendChild(sec);
  }

  function act(btn, path, doneKey) {
    setBusy(btn, true);
    api('POST', '/payments/' + encodeURIComponent(id) + '/' + path, {}).then(function () {
      toast('success', t(doneKey));
      load();
    }, function (err) {
      setBusy(btn, false);
      toast('error', errorText(err));
    });
  }

  function load() {
    skeleton();
    api('GET', '/payments/' + encodeURIComponent(id)).then(render, function (err) { clear(grid).appendChild(h('div', { class: 'eco-span-all' }, errorBlock(err, load))); });
  }
  load();
}

function paymentMethodsPage(container) {
  var frame = tabFrame(t('desc.payment-methods'));
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-card-grid' });
  frame.content.appendChild(grid);

  function skeleton() {
    clear(grid);
    for (var i = 0; i < 5; i++) {
      grid.appendChild(card(h('div', { class: 'p-4' },
        h('div', { class: 'flex items-center justify-between mb-4' }, bar('7rem', 'h-4'), bar('2rem', 'h-5 rounded-full')),
        skeletonField(), skeletonField(),
        h('div', { class: 'flex justify-end' }, bar('5rem', 'h-8')))));
    }
  }

  function methodCard(m) {
    var cfg = clone(m.config || {}) || {};
    var active = !!m.isActive;
    var errTop = h('div', {});
    var fieldEls = {};
    var body = h('div', {});
    (m.fields || []).forEach(function (f) {
      var label = has('pm.field.' + f.name) ? t('pm.field.' + f.name) : f.name;
      var ctl;
      if (f.type === 'bool') {
        ctl = toggle(cfg[f.name], label, function (v) { cfg[f.name] = v; }, 'sm');
        body.appendChild(h('div', { class: 'mb-4' }, ctl));
        fieldEls[f.name] = { setError: function () {} };
        return;
      }
      ctl = textInput(cfg[f.name] == null ? '' : cfg[f.name], { ltr: true, mono: f.type !== 'textarea' || true, multiline: f.type === 'textarea', rows: 3, label: label, oninput: function (v) { cfg[f.name] = v; } });
      if (f.type === 'secret') { ctl.setAttribute('autocomplete', 'off'); ctl.classList.add('eco-secret'); }
      ctl.classList.add('text-sm');
      var wrap = field(label, ctl, f.type === 'secret' && cfg[f.name] ? t('pm.secret.hint') : null, { optional: !f.required });
      fieldEls[f.name] = wrap;
      body.appendChild(wrap);
    });
    if (!(m.fields || []).length) body.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mb-4', text: t('pm.noFields') }));
    var saveBtn = button({ variant: 'primary', size: 'sm', label: t('save'), onclick: function () {
      clear(errTop);
      Object.keys(fieldEls).forEach(function (k) { fieldEls[k].setError(''); });
      setBusy(saveBtn, true, t('saving'));
      api('PUT', '/payment-methods/' + encodeURIComponent(m.method), { isActive: active, config: cfg }).then(function (res) {
        toast('success', t('saved'));
        if (res && res.method) el.replaceWith(methodCard(res));
        else setBusy(saveBtn, false);
      }, function (err) {
        setBusy(saveBtn, false);
        if (err.field && fieldEls[err.field]) fieldEls[err.field].setError(errorText(err));
        else errTop.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
      });
    } });
    var el = card(h('div', { class: 'p-4 flex flex-col h-full' },
      h('div', { class: 'flex items-start justify-between gap-3 mb-4' },
        h('div', {},
          h('div', { class: 'text-lg font-medium', text: has('method.' + m.method) ? t('method.' + m.method) : m.method }),
          h('div', { class: 'flex items-center gap-1.5 mt-1' }, mono(m.method), tag(m.configured ? t('pm.configured') : t('pm.notConfigured'), m.configured ? 'success' : 'secondary'))),
        toggle(active, t('pm.active'), function (v) { active = v; }, 'sm')),
      errTop,
      h('div', { class: 'flex-1' }, body),
      h('div', { class: 'flex justify-end pt-2' }, saveBtn)));
    el.setAttribute('data-method', m.method);
    return el;
  }

  function load() {
    skeleton();
    api('GET', '/payment-methods').then(function (res) {
      var list = Array.isArray(res) ? res : (res && (res.items || res.methods)) || [];
      clear(grid);
      if (!list.length) { grid.appendChild(h('div', { class: 'eco-span-all' }, card(emptyBlock(t('pm.empty'))))); return; }
      list.forEach(function (m) { grid.appendChild(methodCard(m)); });
    }, function (err) { clear(grid).appendChild(h('div', { class: 'eco-span-all' }, errorBlock(err, load))); });
  }
  load();
}

var OVERVIEW_CATALOG = ['currencies', 'metas', 'prizes', 'products', 'gifts', 'categories', 'tags'];

var COVERAGE_PAGES = { languages: 'languages', categories: 'categories', tags: 'tags', currencies: 'currencies', metas: 'metas', prizes: 'prizes', products: 'products', gifts: 'gifts', faqs: 'faqs', news: 'news' };

function coverageLabel(resource) {
  return resource === 'languages' ? t('tab.languages') : has('nav.' + resource) ? t('nav.' + resource) : resource;
}

function coveragePanel() {
  var body = h('div', {});
  var box = section('cov.title', h('div', {}, h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mb-4', text: t('cov.desc') }), body),
    button({ variant: 'muted', size: 'icon', icon: 'refresh', iconClass: 'h-3 w-3', title: t('refresh'), onclick: function () { load(); } }));
  box.classList.add('mt-6');
  box.setAttribute('data-coverage', '');
  var seq = 0;
  function cols(langs) {
    return [{ label: t('cov.resource'), skeleton: '55%' }, { label: t('cov.total'), width: 90 }].concat(langs.map(function (code) { return { label: langName(code), width: 120 }; }));
  }
  function skeleton() {
    var langs = LANGS.active.map(function (l) { return l.code; });
    var chips = h('div', { class: 'flex flex-wrap gap-2 mb-4' }, langs.map(function () { return bar('8rem', 'h-6'); }));
    clear(body).appendChild(h('div', {}, chips, table(cols(langs), skeletonRows(cols(langs), 6))));
  }
  function render(c) {
    var langs = (c && c.languages) || [];
    var missing = (c && c.missing) || {};
    var chips = h('div', { class: 'flex flex-wrap gap-2 mb-4' }, langs.map(function (code) {
      var n = missing[code] || 0;
      return h('span', { class: 'eco-tag eco-tag--' + (n ? 'warning' : 'success') + ' inline-flex items-center gap-1.5 px-2 py-1 text-xs rounded-md', 'data-cov-lang': code, 'data-cov-missing': String(n) },
        h('span', { class: 'font-medium', text: langName(code) + (code === c.default ? ' (' + t('tr.default') + ')' : '') }),
        h('span', { class: 'tabular-nums', text: n ? t('cov.missingN', { n: n }) : t('cov.complete') }));
    }));
    var rows = ((c && c.items) || []).map(function (it) {
      var page = COVERAGE_PAGES[it.resource];
      var href = page ? hashFor(page) : null;
      var name = h('div', { class: 'min-w-0' },
        href ? h('a', { href: href, class: 'eco-link font-medium', text: coverageLabel(it.resource) }) : h('span', { class: 'font-medium', text: coverageLabel(it.resource) }),
        h('div', { class: 'text-xs text-secondary-500', text: t('cov.module.' + it.module) }));
      var cells = [name, h('span', { class: 'tabular-nums', text: fmtNum(it.total || 0) })];
      langs.forEach(function (code) {
        var n = (it.missing || {})[code] || 0;
        cells.push(n
          ? h('a', { href: href || '#', class: 'eco-tag eco-tag--warning inline-flex leading-tight px-2 py-0.5 text-xs rounded-md whitespace-nowrap tabular-nums', 'data-cov': it.resource + ':' + code, text: t('cov.missingN', { n: n }), title: t('cov.open') })
          : h('span', { class: 'eco-text-success inline-flex items-center gap-1 text-xs', 'data-cov': it.resource + ':' + code }, icon('check', 'h-3.5 w-3.5'), h('span', { text: t('cov.complete') })));
      });
      return { item: it, cells: cells };
    });
    clear(body).appendChild(h('div', {}, chips, table(cols(langs), rows)));
  }
  function load() {
    var my = ++seq;
    skeleton();
    baseApi('GET', '/languages/coverage').then(function (c) { if (my === seq) render(c); }, function (err) { if (my === seq) clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
  box.reload = load;
  return box;
}

function languagesPage(container, query, ctx) {
  var list = listPage(container, 'languages', query, ctx);
  var cov = coveragePanel();
  container.appendChild(cov);
  var hook = function () {
    if (!cov.isConnected) { languageHooks.splice(languageHooks.indexOf(hook), 1); return; }
    cov.reload();
    list.reload();
  };
  languageHooks.push(hook);
  return list;
}

function statCard(label, value, sub, href, opts) {
  opts = opts || {};
  var inner = h('div', { class: 'p-4 flex flex-col gap-1 h-full' },
    h('div', { class: 'font-medium text-secondary-700 dark:text-secondary-200', text: label }),
    h('div', { class: 'text-3xl font-semibold tabular-nums' + (opts.alert ? ' text-error-600 dark:text-error-400' : ''), text: fmtNum(value || 0) }),
    sub ? h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: sub }) : null);
  if (!href) return card(inner);
  return h('a', { href: href, class: 'block bg-white dark:bg-secondary-800 overflow-hidden rounded-md shadow hover:bg-secondary-50 dark:hover:bg-secondary-700 transition' }, inner);
}

function overviewPage(container) {
  var frame = pageFrame(t('nav.overview'), t('desc.overview'));
  clear(container).appendChild(frame.el);
  var body = h('div', { class: 'flex flex-col gap-6' });
  frame.content.appendChild(body);

  function skeleton() {
    clear(body);
    [OVERVIEW_CATALOG.length, 6].forEach(function (n, k) {
      var grid = h('div', { class: 'eco-stat-grid' });
      for (var i = 0; i < n; i++) grid.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem'), bar('3rem', 'h-7'), k ? bar('8rem') : null)));
      body.appendChild(h('div', {}, h('div', { class: 'mb-3' }, bar('7rem', 'h-4')), grid));
    });
  }

  function render(s) {
    s = s || {};
    var c = s.counts || {};
    var by = s.paymentsByStatus || {};
    clear(body);
    var catalog = h('div', { class: 'eco-stat-grid' });
    OVERVIEW_CATALOG.forEach(function (name) {
      var active = name === 'products' ? c.activeProducts : name === 'gifts' ? c.activeGifts : undefined;
      catalog.appendChild(statCard(t('nav.' + name), c[name], active !== undefined ? t('active') + ': ' + fmtNum(active) : null, hashFor(name)));
    });
    body.appendChild(h('div', {}, h('h2', { class: 'text-lg font-medium mb-3', text: t('ov.catalog') }), catalog));

    if (c.eventsToday !== undefined) {
      var events = h('div', { class: 'eco-stat-grid', 'data-ov-events': '' });
      events.appendChild(statCard(t('ov.eventsToday'), c.eventsToday, null, hashFor('event-log')));
      events.appendChild(statCard(t('ov.claimsReady'), c.claimsReady, null, null));
      if (c.claimsFailed) events.appendChild(statCard(t('ov.claimsFailed'), c.claimsFailed, null, null, { alert: true }));
      body.appendChild(h('div', {}, h('h2', { class: 'text-lg font-medium mb-3', text: t('ov.events') }), events));
    }

    var payments = h('div', { class: 'eco-stat-grid' });
    payments.appendChild(statCard(t('ov.revenue30'), c.revenueToman30, t('toman', { n: '' }).trim(), null));
    payments.appendChild(statCard(t('ov.paymentsToday'), c.paymentsToday, t('ov.total') + ': ' + fmtNum(c.payments || 0), hashFor('payments')));
    ['pending', 'success', 'failed', 'expired'].forEach(function (st) {
      payments.appendChild(statCard(t('status.' + st), by[st], null, hashFor('payments', null, { status: st })));
    });
    if (c.grantErrors) payments.appendChild(statCard(t('ov.grantErrors'), c.grantErrors, null, hashFor('payments', null, { status: 'success' }), { alert: true }));
    body.appendChild(h('div', {}, h('h2', { class: 'text-lg font-medium mb-3', text: t('ov.payments') }), payments));

    var methods = s.paymentMethods || [];
    if (methods.length) {
      var list = h('div', { class: 'flex flex-col' });
      methods.forEach(function (m) {
        list.appendChild(h('a', { href: hashFor('payment-methods'), class: 'eco-log flex flex-wrap items-center gap-2 px-4 py-3 border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' },
          h('span', { class: 'font-medium flex-1', text: has('method.' + m.method) ? t('method.' + m.method) : m.method }),
          m.sandbox ? tag(t('ov.sandbox'), 'warning') : null,
          tag(m.configured ? t('pm.configured') : t('pm.notConfigured'), m.configured ? 'secondary' : 'error'),
          tag(m.isActive ? t('active') : t('inactive'), m.isActive ? 'success' : 'secondary')));
      });
      body.appendChild(h('div', {}, h('h2', { class: 'text-lg font-medium mb-3', text: t('ov.methods') }), card(list)));
    }
  }

  function load() {
    skeleton();
    api('GET', '/summary').then(render, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}

var CHANNELS = ['in_app', 'sms', 'email'];
var CHANNEL_ICON = { in_app: 'notifications', sms: 'comments', email: 'send' };
var NOTIF_STATUS = ['queued', 'sending', 'done', 'failed'];
var NOTIF_COLOR = { queued: 'secondary', sending: 'primary', done: 'success', failed: 'error' };
var RECIPIENT_STATUS = ['pending', 'sent', 'skipped', 'failed'];
var RECIPIENT_COLOR = { pending: 'secondary', sent: 'success', skipped: 'warning', failed: 'error' };
var SMS_MAX = 600;
var pendingComments = null;
var badgeAt = 0;

function kindLabel(kind) {
  return has('pv.kind.' + kind) ? t('pv.kind.' + kind) : kind;
}

function quietGet(path) {
  var tk = token();
  if (!tk) return Promise.reject(new ApiError(401, 'UNAUTHENTICATED', ''));
  return fetch(PUB_BASE + path, { headers: { Accept: 'application/json', Authorization: 'Bearer ' + tk }, credentials: 'same-origin' }).then(function (res) {
    if (!res.ok) throw new ApiError(res.status, null, '');
    return res.json();
  });
}

function refreshBadge(force) {
  var now = Date.now();
  if (!force && now - badgeAt < 60000) return;
  badgeAt = now;
  quietGet('/comments?status=pending&limit=1').then(function (res) {
    var n = (res && res.total) || 0;
    if (n !== pendingComments) {
      pendingComments = n;
      schedule();
    }
  }, function () { return null; });
}

function navBadge(page, collapsed) {
  var n = page === 'comments' ? pendingComments : page === 'reports' ? pendingReports : supportBadge(page);
  if (!n) return null;
  var label = t(page === 'comments' ? 'cm.pendingN' : page === 'reports' ? 'rp.openN' : page === 'tickets' ? 'tk.waitingN' : 'lc.waitingN', { n: n });
  if (collapsed) return h('span', { class: 'eco-nav-dot', 'aria-label': label });
  return h('span', { class: 'eco-nav-badge tabular-nums', 'aria-label': label, text: fmtNum(n) });
}

function userCell(id, username, displayName) {
  return h('div', { class: 'min-w-0' },
    h('a', { href: '#/players/' + encodeURIComponent(id || ''), class: 'eco-link font-medium truncate block', text: displayName || username || String(id || '').slice(0, 8) }),
    h('div', { class: 'text-xs text-secondary-500 truncate' }, mono(username && displayName ? username : String(id || '').slice(0, 8))));
}

function commentAct(btn, id, accepted) {
  setBusy(btn, true);
  pubApi('PUT', '/comments/' + encodeURIComponent(id), { isAccepted: accepted }).then(function () {
    toast('success', t(accepted ? 'cm.accepted.done' : 'cm.rejected.done'));
    refreshBadge(true);
    if (rendered && rendered.list) rendered.list.reload();
  }, function (err) {
    setBusy(btn, false);
    toast('error', errorText(err));
  });
}

function commentActions(it) {
  var box = h('div', { class: 'flex flex-wrap items-center gap-1 justify-end eco-row-actions' });
  function reset() {
    clear(box);
    if (!it.isAccepted) box.appendChild(button({ variant: 'primary', size: 'sm', icon: 'check', label: t('cm.accept'), onclick: function (e) { commentAct(e.currentTarget, it.id, true); } }));
    else box.appendChild(button({ variant: 'ghost', size: 'sm', icon: 'close', label: t('cm.reject'), title: t('cm.reject.hint'), onclick: function (e) { commentAct(e.currentTarget, it.id, false); } }));
    box.appendChild(button({ variant: 'danger', size: 'sm', icon: 'trash', title: t('delete'), onclick: ask }));
  }
  function ask() {
    clear(box);
    box.appendChild(h('span', { class: 'text-sm text-secondary-700 dark:text-secondary-200', text: t('delete.confirm') }));
    box.appendChild(button({ variant: 'dangerSolid', size: 'sm', label: t('delete.yes'), onclick: function (e) {
      setBusy(e.currentTarget, true);
      pubApi('DELETE', '/comments/' + encodeURIComponent(it.id)).then(function () {
        toast('success', t('deleted'));
        refreshBadge(true);
        if (rendered && rendered.list) rendered.list.reload();
      }, function (err) { reset(); toast('error', errorText(err)); });
    } }));
    box.appendChild(button({ variant: 'ghost', size: 'sm', label: t('no'), onclick: reset }));
  }
  reset();
  return box;
}

var COMMENT_LIST = {
  noRow: true,
  filters: [
    { name: 'status', label: 'filter.status', options: opts('cm.', ['pending', 'accepted']) },
    { name: 'newsId', label: 'filter.news', type: 'ref', resource: 'news', api: 'publication' }
  ],
  columns: [
    { label: 'col.comment', render: function (it) {
      return h('div', { class: 'min-w-0 flex flex-col gap-1' },
        h('div', { class: 'flex items-start gap-1.5' },
          it.parentId ? h('span', { class: 'text-secondary-400 mt-0.5', title: t('cm.reply') }, icon('reply', 'h-3.5 w-3.5')) : null,
          h('div', { class: 'eco-clamp-3 eco-pre-line break-words', dir: 'auto', text: it.message })),
        h('div', { class: 'flex flex-wrap items-center eco-meta text-xs text-secondary-500' },
          h('a', { href: '#/players/' + encodeURIComponent(it.userId), class: 'eco-link font-medium', text: it.displayName || it.username || String(it.userId).slice(0, 8) }),
          h('span', { text: fmtDate(it.createdAt) }),
          it.parentId ? h('span', { text: t('cm.reply') }) : null,
          it.replyCount ? h('span', { text: t('cm.replies', { n: it.replyCount }) }) : null));
    }, skeleton: '75%' },
    { label: 'col.news', render: function (it) { return h('a', { href: hashFor('news', it.newsId), class: 'eco-link eco-clamp', text: i18nText(it.newsTitle) }); }, width: 220, skeleton: '60%' },
    { label: 'col.status', render: function (it) { return tag(t(it.isAccepted ? 'cm.accepted' : 'cm.pending'), it.isAccepted ? 'success' : 'warning'); }, width: 120 },
    { label: 'col.actions', render: commentActions, width: 210, skeleton: '80%' }
  ]
};

function targetText(tg) {
  tg = tg || {};
  if (tg.mode === 'all') return t('tg.all');
  if (tg.mode === 'group') return t('tg.group');
  if (tg.mode === 'tags') return t('tg.tagsN', { n: (tg.tagIds || []).length });
  return t('tg.usersN', { n: (tg.userIds || []).length });
}

function notifTitle(it) {
  return i18nText(it.title) || mdPlain(i18nText(it.body));
}

var NOTIFICATION_LIST = {
  create: function () { return button({ variant: 'primary', size: 'page', icon: 'send', label: t('nt.new'), onclick: function () { go('notifications', 'new'); } }); },
  filters: [
    { name: 'channel', label: 'filter.channel', options: opts('ch.', CHANNELS) },
    { name: 'status', label: 'filter.status', options: opts('nst.', NOTIF_STATUS) }
  ],
  columns: [
    { label: 'col.notification', render: function (it) {
      return h('div', { class: 'min-w-0' },
        h('div', { class: 'font-medium eco-clamp', dir: 'auto', text: notifTitle(it) }),
        h('div', { class: 'text-xs text-secondary-500 flex flex-wrap eco-meta' }, h('span', { text: fmtDate(it.createdAt) }), it.createdBy ? h('span', { text: it.createdBy }) : null));
    }, skeleton: '65%' },
    { label: 'col.channel', render: function (it) {
      return h('div', { class: 'flex flex-wrap items-center gap-1' }, tag(t('ch.' + it.channel), it.channel === 'in_app' ? 'primary' : 'secondary'), it.providerKind ? h('span', { class: 'text-xs text-secondary-500', text: kindLabel(it.providerKind) }) : null);
    }, width: 170 },
    { label: 'col.target', render: function (it) { return h('span', { class: 'whitespace-nowrap', text: targetText(it.target) }); }, width: 120 },
    { label: 'col.delivery', render: function (it) {
      return h('div', { class: 'flex flex-wrap items-center gap-1' },
        h('span', { class: 'tabular-nums whitespace-nowrap text-sm', text: t('nt.sentOf', { sent: it.sentCount || 0, total: it.recipientCount || 0 }) }),
        it.skippedCount ? tag(t('nt.skippedN', { n: it.skippedCount }), 'warning') : null,
        it.failedCount ? tag(t('nt.failedN', { n: it.failedCount }), 'error') : null);
    }, width: 210 },
    { label: 'col.status', render: function (it) { return tag(t('nst.' + it.status), NOTIF_COLOR[it.status]); }, width: 120 }
  ]
};

var LISTS = { payments: PAYMENT_LIST, comments: COMMENT_LIST, notifications: NOTIFICATION_LIST };

function statTile(label, value, color) {
  return h('div', { class: 'rounded-md border border-secondary-200 dark:border-secondary-700 px-3 py-2' },
    h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: label }),
    h('div', { class: 'text-2xl font-semibold tabular-nums' + (color ? ' eco-text-' + color : ''), text: fmtNum(value || 0) }));
}

function i18nBlock(v, markdown) {
  v = v || {};
  var out = h('div', { class: 'flex flex-col gap-3' });
  var order = LANGS.items.map(function (l) { return l.code; });
  Object.keys(v).sort().forEach(function (k) { if (order.indexOf(k) === -1) order.push(k); });
  order.forEach(function (l) {
    if (!v[l]) return;
    var ltr = langDir(l) === 'ltr';
    out.appendChild(h('div', { 'data-lang': l },
      h('div', { class: 'text-xs text-secondary-500 mb-1', text: langName(l) + (l === LANGS.def ? ' (' + t('tr.default') + ')' : '') }),
      markdown ? mdPreview(v[l], ltr) : h('div', { class: 'text-sm eco-pre-line break-words', dir: ltr ? 'ltr' : 'rtl', text: v[l] })));
  });
  return out;
}

function notificationDetail(container, id) {
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('notifications', null, lastQuery.notifications || {}); } });
  var frame = pageFrame(t('nt.title'), null, backBtn);
  var titleBox = frame.el.querySelector('h1');
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-detail-grid' });
  frame.content.appendChild(grid);
  var rState = { page: 1, limit: 20, status: '', total: 0, loading: true, loaded: false };
  var recBox = h('div', {});
  var recPagerTop = h('div', { class: 'flex items-center justify-end' });
  var recPagerBottom = h('div', { class: 'flex items-center justify-end eco-pt-3' });
  var timer = null;
  var built = false;
  var statsBox = h('div', { class: 'eco-tile-grid' });
  var statusBox = h('div', {});

  function alive() { return container.isConnected && rendered && rendered.root === container; }

  function skeleton() {
    clear(grid);
    [5, 4].forEach(function (n) {
      var rows = [];
      for (var i = 0; i < n; i++) rows.push([h('dt', {}, bar('6rem')), h('dd', {}, bar(['70%', '50%', '40%'][i % 3]))]);
      grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), h('dl', { class: 'eco-dl' }, rows))));
    });
    var tiles = [];
    for (var k = 0; k < 6; k++) tiles.push(h('div', { class: 'rounded-md border border-secondary-200 dark:border-secondary-700 px-3 py-2' }, bar('4rem'), bar('2.5rem', 'h-7 mt-2')));
    grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), h('div', { class: 'eco-tile-grid' }, tiles)), 'eco-span-all'));
    grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), table(recColumns(), skeletonRows(recColumns(), 5))), 'eco-span-all'));
  }

  function recColumns() {
    return [
      { label: t('col.user'), skeleton: '50%' },
      { label: t('col.status'), width: 120 },
      { label: t('col.address'), width: 200, skeleton: '70%' },
      { label: t('col.error'), skeleton: '40%' },
      { label: t('col.attempts'), width: 80 },
      { label: t('col.time'), width: 180 }
    ];
  }

  function paintStats(n) {
    clear(statusBox).appendChild(h('div', { class: 'flex flex-wrap items-center gap-1.5' },
      tag(t('nst.' + n.status), NOTIF_COLOR[n.status]),
      n.status === 'queued' || n.status === 'sending' ? h('span', { class: 'text-xs text-secondary-500', text: t('nt.live') }) : null));
    clear(statsBox);
    statsBox.appendChild(statTile(t('nt.recipients'), n.recipientCount));
    statsBox.appendChild(statTile(t('rst.sent'), n.sentCount, 'success'));
    statsBox.appendChild(statTile(t('rst.skipped'), n.skippedCount, n.skippedCount ? 'warning' : null));
    statsBox.appendChild(statTile(t('rst.failed'), n.failedCount, n.failedCount ? 'error' : null));
    statsBox.appendChild(statTile(t('rst.pending'), n.pendingCount));
    if (n.channel === 'in_app') statsBox.appendChild(statTile(t('nt.seen'), n.seenCount));
  }

  function build(n) {
    built = true;
    clear(grid);
    clear(titleBox).appendChild(h('span', { text: t('nt.title') + ' ' }));
    titleBox.appendChild(h('span', { class: 'font-mono text-base text-secondary-500', dir: 'ltr', text: String(n.id).slice(0, 8) }));
    var action = n.action && n.action.type ? h('div', { class: 'flex flex-wrap items-center gap-1.5' }, tag(t('act.' + n.action.type), 'secondary'), mono(n.action.value), n.action.label && i18nText(n.action.label) ? h('span', { class: 'text-sm', text: i18nText(n.action.label) }) : null) : null;
    grid.appendChild(section('nt.info', dl([
      ['col.status', statusBox],
      ['col.channel', h('div', { class: 'flex flex-wrap items-center gap-1.5' }, tag(t('ch.' + n.channel), n.channel === 'in_app' ? 'primary' : 'secondary'), n.providerKind ? h('span', { class: 'text-sm', text: kindLabel(n.providerKind) }) : null)],
      ['col.target', h('span', { text: targetText(n.target) })],
      ['nt.action', action],
      ['nt.createdBy', n.createdBy || null],
      ['pay.createdAt', dateCell(n.createdAt)],
      ['pay.updatedAt', dateCell(n.updatedAt)]
    ])));
    var content = h('div', { class: 'flex flex-col gap-4' });
    var hasTitle = !!i18nText(n.title);
    if (hasTitle) content.appendChild(h('div', {}, h('div', { class: 'text-sm font-medium mb-1', text: t('nt.f.title') }), i18nBlock(n.title)));
    content.appendChild(h('div', {}, hasTitle ? h('div', { class: 'text-sm font-medium mb-1', text: t('nt.f.body') }) : null, i18nBlock(n.body, n.channel !== 'sms')));
    grid.appendChild(section('nt.content', content));
    var stats = section('nt.delivery', statsBox);
    stats.classList.add('eco-span-all');
    grid.appendChild(stats);
    var filter = filterSelect(t('filter.status'), rState.status, opts('rst.', RECIPIENT_STATUS)(), function (v) { rState.status = v; rState.page = 1; loadRecipients(); });
    var recs = section('nt.recipientsList', h('div', {}, recBox, recPagerBottom), h('div', { class: 'flex flex-wrap items-center gap-2 justify-end' }, filter, recPagerTop));
    recs.classList.add('eco-span-all');
    grid.appendChild(recs);
  }

  function paintRecPager() {
    clear(recPagerTop).appendChild(pager(rState, function (p) { rState.page = p; loadRecipients(); }));
    clear(recPagerBottom).appendChild(pager(rState, function (p) { rState.page = p; loadRecipients(); }));
  }

  function renderRecipients(r) {
    var items = (r && r.items) || [];
    rState.total = (r && r.total) || 0;
    rState.loaded = true;
    rState.loading = false;
    recBox.classList.remove('opacity-60');
    clear(recBox);
    if (!items.length) {
      recBox.appendChild(h('div', { class: 'text-sm text-secondary-500 py-6 text-center', text: rState.status ? t('empty.search') : t('nt.noRecipients') }));
    } else {
      var cols = recColumns();
      recBox.appendChild(table(cols, items.map(function (it) {
        var sandbox = it.error === 'sandbox';
        return { item: it, cells: [
          userCell(it.userId, it.username, it.displayName),
          h('div', { class: 'flex flex-wrap gap-1' }, tag(t('rst.' + it.status), RECIPIENT_COLOR[it.status]), sandbox ? tag(t('pv.sandbox.short'), 'warning') : null, it.lang ? h('span', { 'data-recipient-lang': it.lang }, tag(langName(it.lang), 'secondary')) : null),
          it.address ? mono(it.address) : h('span', {}),
          it.error && !sandbox && it.error !== 'no address' ? h('span', { class: 'text-xs text-error-600 dark:text-error-400 break-words', dir: 'auto', text: it.error }) : h('span', {}),
          h('span', { class: 'tabular-nums', text: fmtNum(it.attempts || 0) }),
          h('div', { class: 'flex flex-col text-xs text-secondary-600 dark:text-secondary-300' },
            it.sentAt ? h('span', { class: 'whitespace-nowrap', text: t('nt.sentAt', { t: fmtDate(it.sentAt) }) }) : null,
            it.seenAt ? h('span', { class: 'whitespace-nowrap', text: t('nt.seenAt', { t: fmtDate(it.seenAt) }) }) : null)
        ] };
      })));
    }
    paintRecPager();
  }

  function fetchDetail() {
    return pubApi('GET', '/notifications/' + encodeURIComponent(id) + qs({ rPage: rState.page, rLimit: rState.limit, rStatus: rState.status }));
  }

  function loadRecipients() {
    rState.loading = true;
    recBox.classList.add('opacity-60');
    paintRecPager();
    fetchDetail().then(function (n) {
      if (!alive()) return;
      paintStats(n);
      renderRecipients(n.recipients);
    }, function (err) { recBox.classList.remove('opacity-60'); toast('error', errorText(err)); });
  }

  function poll() {
    clearTimeout(timer);
    timer = setTimeout(function () {
      if (!alive()) return;
      fetchDetail().then(function (n) {
        if (!alive()) return;
        paintStats(n);
        renderRecipients(n.recipients);
        if (n.status === 'queued' || n.status === 'sending') poll();
      }, function () { poll(); });
    }, 2500);
  }

  function load() {
    skeleton();
    fetchDetail().then(function (n) {
      if (!alive()) return;
      build(n);
      paintStats(n);
      renderRecipients(n.recipients);
      if (n.status === 'queued' || n.status === 'sending') poll();
    }, function (err) { clear(grid).appendChild(h('div', { class: 'eco-span-all' }, errorBlock(err, load))); });
  }
  load();
}

function groupsApi(method, path) {
  var m = /^\/groups\/([^?]+)$/.exec(path);
  if (m) return consoleApi('GET', '/group/' + m[1]).then(function (g) { return g && g.id ? { id: g.id, name: g.name, key: null } : null; });
  var q = new URLSearchParams(path.split('?')[1] || '').get('q') || '';
  return consoleApi('GET', '/group' + qs({ filter: q })).then(function (res) {
    return { items: ((res && res.groups) || []).map(function (g) { return { id: g.id, name: g.name }; }) };
  });
}

function usersApi(method, path) {
  var m = /^\/users\/([^?]+)$/.exec(path);
  var q = m ? decodeURIComponent(m[1]) : new URLSearchParams(path.split('?')[1] || '').get('q') || '';
  if (!q) return Promise.resolve(m ? null : { items: [] });
  return pubApi('GET', '/users/search' + qs({ q: q })).then(function (list) {
    var items = (list || []).map(function (u) { return { id: u.id, name: u.displayName || u.username, key: u.username }; });
    return m ? items.filter(function (u) { return u.id === q; })[0] || null : { items: items };
  });
}

function composePage(container, query) {
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('notifications', null, lastQuery.notifications || {}); } });
  var frame = pageFrame(t('nt.new'), t('nt.new.desc'), backBtn);
  clear(container).appendChild(frame.el);
  var layout = h('div', { class: 'eco-compose' });
  frame.content.appendChild(layout);
  var v = {
    channel: CHANNELS.indexOf(query.channel) !== -1 ? query.channel : 'in_app',
    providerId: '',
    title: {},
    body: {},
    action: { type: '', value: '', label: {} },
    target: { mode: query.tagId ? 'tags' : 'users', userIds: query.userId ? [query.userId] : [], groupId: '', tagIds: query.tagId ? [query.tagId] : [] }
  };
  var providers = [];
  var preview = null;
  var previewSeq = 0;

  function skeleton() {
    clear(layout);
    var main = h('div', { class: 'flex flex-col gap-4 min-w-0' });
    [2, 3, 1, 2].forEach(function (n) {
      var f = [];
      for (var i = 0; i < n; i++) f.push(skeletonField(i === 1 ? 'h-24' : 'h-10'));
      main.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), f)));
    });
    layout.appendChild(main);
    layout.appendChild(h('div', {}, card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem', 'h-4'), bar('4rem', 'h-9'), bar('80%'), bar('100%', 'h-9')))));
  }

  function build() {
    clear(layout);
    var main = h('div', { class: 'flex flex-col gap-4 min-w-0' });
    var aside = h('div', { class: 'eco-compose-aside' });
    layout.appendChild(main);
    layout.appendChild(aside);
    var errTop = h('div', {});
    var fields = {};

    var providerBox = h('div', {});
    var channelSeg = segmented(CHANNELS.map(function (c) { return { value: c, label: t('ch.' + c), icon: CHANNEL_ICON[c] }; }), v.channel, function (c) {
      v.channel = c;
      v.providerId = '';
      paintChannel();
      paintCount();
    }, t('col.channel'), 'md');
    main.appendChild(section('nt.s.channel', h('div', {}, h('div', { class: 'mb-4' }, channelSeg), providerBox)));

    var titleRow = h('div', {});
    var titleCtl = translatedInput(v.title, { label: t('nt.f.title'), oninput: function (x) { v.title = x; fields.title.setError(''); titleCtl.markLang(null); } });
    fields.title = field(t('nt.f.title'), titleCtl);
    fields.title.ctl = titleCtl;
    titleRow.appendChild(fields.title);
    var bodyBox = h('div', {});
    main.appendChild(section('nt.s.content', h('div', {}, titleRow, bodyBox)));

    var actType = selectInput('', [{ value: 'url', label: t('act.url') }, { value: 'goto', label: t('act.goto') }], function (x) {
      v.action.type = x;
      actValue.hidden = !x;
      actValue.placeholder = x === 'url' ? 'https://' : 'shop';
      if (x === 'goto') actValue.setAttribute('list', 'eco-nt-goto'); else actValue.removeAttribute('list');
      fields.action.setError('');
    }, { label: t('act.type'), placeholder: t('act.none') });
    var actValue = textInput('', { ltr: true, mono: true, label: t('act.value'), oninput: function (x) { v.action.value = x.trim(); fields.action.setError(''); } });
    actValue.hidden = true;
    fields.action = field(null, h('div', { class: 'eco-form-row' }, actType, actValue, h('datalist', { id: 'eco-nt-goto' }, GOTO_SCREENS.map(function (s) { return h('option', { value: s }); }))), t('nt.action.hint'));
    var actLabelCtl = translatedInput(v.action.label, { label: t('act.label'), oninput: function (x) { v.action.label = x; fields.actionLabel.setError(''); actLabelCtl.markLang(null); } });
    fields.actionLabel = field(t('act.label'), actLabelCtl, t('nt.action.label.hint'), { optional: true });
    fields.actionLabel.ctl = actLabelCtl;
    fields.actionLabel.hidden = true;
    var actType0 = actType;
    actType0.addEventListener('change', function () { fields.actionLabel.hidden = !actType0.value; });
    var actionSection = section('nt.s.action', h('div', {}, fields.action, fields.actionLabel));
    main.appendChild(actionSection);

    var targetBox = h('div', {});
    var modeSeg = segmented([{ value: 'users', label: t('tg.users'), icon: 'users' }, { value: 'group', label: t('tg.group'), icon: 'categories' }, { value: 'tags', label: t('tg.tags'), icon: 'tags' }, { value: 'all', label: t('tg.all'), icon: 'notifications' }], v.target.mode, function (m) {
      v.target.mode = m;
      paintTarget();
      refreshPreview();
    }, t('col.target'), 'md');
    main.appendChild(section('nt.s.target', h('div', {}, h('div', { class: 'mb-4' }, modeSeg), targetBox)));

    var countBox = h('div', { class: 'flex flex-col gap-1', 'aria-live': 'polite' });
    var sendBtn = button({ variant: 'primary', size: 'footer', icon: 'send', label: t('nt.send'), class: 'w-full', onclick: function () { send(); } });
    aside.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' },
      h('h2', { class: 'text-lg font-medium', text: t('nt.summary') }),
      countBox, errTop, sendBtn)));

    function paintChannel() {
      clear(providerBox);
      titleRow.hidden = v.channel === 'sms';
      actionSection.hidden = v.channel === 'sms';
      clear(bodyBox);
      var max = v.channel === 'sms' ? SMS_MAX : v.channel === 'email' ? 20000 : 2000;
      var bodyCtl = translatedInput(v.body, { label: t(v.channel === 'sms' ? 'nt.f.sms' : 'nt.f.body'), markdown: v.channel !== 'sms', multiline: true, rows: 4, max: max, counter: v.channel === 'sms', oninput: function (x) { v.body = x; fields.body.setError(''); bodyCtl.markLang(null); } });
      fields.body = field(t(v.channel === 'sms' ? 'nt.f.sms' : 'nt.f.body'), bodyCtl, v.channel === 'sms' ? t('nt.sms.hint') : null);
      fields.body.ctl = bodyCtl;
      bodyBox.appendChild(fields.body);
      if (v.channel === 'in_app') {
        providerBox.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: t('nt.inapp.hint') }));
        return;
      }
      var list = providers.filter(function (p) { return p.channel === v.channel && p.isActive; });
      if (!list.length) {
        providerBox.appendChild(notice('error', t('nt.noProvider'), h('a', { href: hashFor(v.channel), class: 'eco-link text-sm font-medium', text: t('nt.goSettings.' + v.channel) })));
        return;
      }
      var def = list.filter(function (p) { return p.isDefault; })[0] || list[0];
      var options = list.map(function (p) { return { value: p.id, label: kindLabel(p.kind) + (p.id === def.id ? ' (' + t('pv.default') + ')' : '') + (p.sandbox ? ' · ' + t('pv.sandbox.short') : '') }; });
      if (!v.providerId) v.providerId = def.id;
      var sandboxNote = h('div', {});
      var paintNote = function () {
        var p = list.filter(function (x) { return x.id === v.providerId; })[0];
        clear(sandboxNote);
        if (!p) return;
        if (p.sandbox) sandboxNote.appendChild(h('div', { class: 'eco-sandbox-note rounded-md px-3 py-2 text-sm flex items-center gap-2' }, icon('flask', 'h-4 w-4 shrink-0'), h('span', { text: t('nt.sandbox.note') })));
        else sandboxNote.appendChild(notice('error', t('nt.live.note')));
      };
      fields.provider = field(t('nt.f.provider'), selectInput(v.providerId, options, function (x) { v.providerId = x; paintNote(); }, { placeholder: false, label: t('nt.f.provider') }));
      providerBox.appendChild(fields.provider);
      providerBox.appendChild(sandboxNote);
      paintNote();
    }

    function paintTarget() {
      clear(targetBox);
      fields.target = null;
      if (v.target.mode === 'users') {
        fields.target = field(t('tg.pickUsers'), refMulti({ resource: 'users', api: usersApi, value: v.target.userIds, label: t('tg.pickUsers'), placeholder: t('tg.searchUsers'), emptyText: t('tg.searchUsers.hint'), onchange: function (ids) {
          v.target.userIds = ids;
          fields.target.setError('');
          refreshPreview();
        } }), t('tg.users.hint'));
      } else if (v.target.mode === 'group') {
        fields.target = field(t('tg.pickGroup'), refPicker({ resource: 'groups', api: groupsApi, value: v.target.groupId, required: false, label: t('tg.pickGroup'), placeholder: t('tg.pickGroup'), onchange: function (id) {
          v.target.groupId = id || '';
          fields.target.setError('');
          refreshPreview();
        } }), t('tg.group.hint'));
      } else if (v.target.mode === 'tags') {
        fields.target = field(t('tg.pickTags'), refMulti({ resource: 'tags', api: baseApi, filter: { keyType: 'user' }, value: v.target.tagIds, label: t('tg.pickTags'), placeholder: t('tg.pickTags'), onchange: function (ids) {
          v.target.tagIds = ids;
          fields.target.setError('');
          refreshPreview();
        } }), t('tg.tags.hint'));
      } else {
        targetBox.appendChild(notice('secondary', t('tg.all.hint')));
      }
      if (fields.target) targetBox.appendChild(fields.target);
    }

    function paintCount() {
      clear(countBox);
      if (preview === 'loading') {
        countBox.appendChild(bar('4rem', 'h-9'));
        countBox.appendChild(bar('80%'));
        return;
      }
      if (!preview || preview.error) {
        countBox.appendChild(h('div', { class: 'text-sm text-secondary-500', text: preview && preview.error ? errorText(preview.error) : t('tg.none') }));
        return;
      }
      countBox.appendChild(h('div', { class: 'flex items-end gap-2' },
        h('span', { class: 'text-3xl font-semibold tabular-nums', 'data-recipient-count': String(preview.count), text: fmtNum(preview.count) }),
        h('span', { class: 'text-sm text-secondary-500', text: t('nt.recipientsWord') })));
      var reach = v.channel === 'sms' ? preview.withPhone : v.channel === 'email' ? preview.withEmail : preview.count;
      countBox.appendChild(h('div', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: t('tg.withPhone', { n: preview.withPhone }) + '، ' + t('tg.withEmail', { n: preview.withEmail }) }));
      if (v.channel !== 'in_app' && reach < preview.count) {
        countBox.appendChild(h('div', { class: 'text-sm eco-text-warning', text: t('tg.skipped', { n: preview.count - reach }) }));
      }
    }

    var runPreview = debounce(function () {
      var my = ++previewSeq;
      var tg = v.target;
      if ((tg.mode === 'users' && !tg.userIds.length) || (tg.mode === 'group' && !tg.groupId) || (tg.mode === 'tags' && !tg.tagIds.length)) { preview = null; paintCount(); return; }
      pubApi('GET', '/notifications/preview-target' + qs({ mode: tg.mode, groupId: tg.mode === 'group' ? tg.groupId : '', userIds: tg.mode === 'users' ? tg.userIds.join(',') : '', tagIds: tg.mode === 'tags' ? tg.tagIds.join(',') : '' })).then(function (res) {
        if (my !== previewSeq) return;
        preview = res;
        paintCount();
      }, function (err) {
        if (my !== previewSeq) return;
        preview = { error: err };
        paintCount();
      });
    }, 250);

    function refreshPreview() {
      preview = 'loading';
      paintCount();
      runPreview();
    }

    function validate() {
      var ok = true;
      var setErr = function (f, msg) { if (f) f.setError(msg); if (msg) ok = false; };
      [fields.title, fields.body, fields.action, fields.actionLabel, fields.target].forEach(function (f) { if (f) f.setError(''); });
      var max = v.channel === 'sms' ? SMS_MAX : v.channel === 'email' ? 20000 : 2000;
      var trErr = function (f, val, o) {
        var hit = trCheck(val, o);
        f.ctl.markLang(hit ? hit.lang : null);
        if (hit) setErr(f, hit.code === 'TEXT_TOO_LONG' && v.channel === 'sms' ? t('nt.sms.tooLong') : t('code.' + hit.code, { lang: langName(hit.lang) }));
      };
      if (v.channel !== 'sms') trErr(fields.title, v.title, { required: true, max: 200 });
      trErr(fields.body, v.body, { required: true, max: max });
      if (v.channel !== 'sms' && v.action.type) trErr(fields.actionLabel, v.action.label, { max: 64 });
      if (v.channel !== 'sms' && v.action.type) {
        if (!v.action.value) setErr(fields.action, t('act.value') + ': ' + t('err.required'));
        else if (v.action.type === 'url' && !/^https?:\/\/\S+$/i.test(v.action.value)) setErr(fields.action, t('code.INVALID_ACTION'));
      }
      if (v.target.mode === 'users' && !v.target.userIds.length) setErr(fields.target, t('code.TARGET_USERS_REQUIRED'));
      if (v.target.mode === 'group' && !v.target.groupId) setErr(fields.target, t('err.required'));
      if (v.target.mode === 'tags' && !v.target.tagIds.length) setErr(fields.target, t('code.TARGET_TAGS_REQUIRED'));
      if (v.channel !== 'in_app' && !providers.some(function (p) { return p.channel === v.channel && p.isActive; })) ok = false;
      return ok;
    }

    function payload() {
      var p = { channel: v.channel, body: trClean(v.body), target: { mode: v.target.mode } };
      if (v.channel !== 'sms') {
        p.title = trClean(v.title);
        if (v.action.type) {
          p.action = { type: v.action.type, value: v.action.value };
          var lbl = trClean(v.action.label);
          if (Object.keys(lbl).length) p.action.label = lbl;
        }
      }
      if (v.target.mode === 'users') p.target.userIds = v.target.userIds.slice();
      if (v.target.mode === 'group') p.target.groupId = v.target.groupId;
      if (v.target.mode === 'tags') p.target.tagIds = v.target.tagIds.slice();
      if (v.channel !== 'in_app' && v.providerId) p.providerId = v.providerId;
      return p;
    }

    function send() {
      clear(errTop);
      if (!validate()) {
        var bad = layout.querySelector('.eco-invalid');
        if (bad) bad.scrollIntoView({ block: 'center', behavior: 'smooth' });
        return;
      }
      var prov = providers.filter(function (p) { return p.id === v.providerId; })[0];
      var live = v.channel !== 'in_app' && prov && !prov.sandbox;
      var count = preview && !preview.error && preview !== 'loading' ? preview.count : null;
      var ask = Promise.resolve(true);
      if (v.target.mode === 'all' || live) {
        ask = confirmModal({
          title: v.target.mode === 'all' ? t('nt.confirmAll.title') : t('nt.confirmLive.title'),
          text: (v.target.mode === 'all' ? (count !== null ? t('nt.confirmAll.text', { n: count }) : t('nt.confirmAll.textNoCount')) : '') + (live ? (v.target.mode === 'all' ? ' ' : '') + t('nt.confirmLive.text') : ''),
          ok: v.target.mode === 'all' ? t('nt.confirmAll.ok') : t('nt.send'),
          danger: live,
          icon: 'send'
        });
      }
      ask.then(function (yes) {
        if (!yes) return;
        setBusy(sendBtn, true, t('nt.sending'));
        pubApi('POST', '/notifications', payload()).then(function (res) {
          toast('success', t('nt.queued'));
          go('notifications', res.id);
        }, function (err) {
          setBusy(sendBtn, false);
          var map = { title: fields.title, body: fields.body, action: fields.action, 'action.label': fields.actionLabel, 'target.userIds': fields.target, 'target.groupId': fields.target, 'target.tagIds': fields.target, providerId: fields.provider };
          var f = err && err.field && map[err.field];
          var vals = { title: v.title, body: v.body, 'action.label': v.action.label };
          var code = f && f.ctl ? trServerLang(err, vals[err.field], { max: err.field === 'title' ? 200 : err.field === 'action.label' ? 64 : v.channel === 'sms' ? SMS_MAX : v.channel === 'email' ? 20000 : 2000 }) : null;
          if (f && f.ctl) f.ctl.markLang(code);
          if (f) { f.hidden = false; f.setError(errorText(err, code)); f.scrollIntoView({ block: 'center', behavior: 'smooth' }); }
          else errTop.appendChild(notice('error', errorText(err)));
        });
      });
    }

    paintChannel();
    paintTarget();
    paintCount();
    if (v.target.userIds.length || v.target.tagIds.length) refreshPreview();
  }

  function load() {
    skeleton();
    Promise.all([pubApi('GET', '/providers'), freshLanguages()]).then(function (all) {
      var res = all[0];
      if (!container.isConnected) return;
      providers = (res && res.items) || [];
      build();
    }, function (err) { clear(layout).appendChild(h('div', { class: 'eco-span-all' }, errorBlock(err, load))); });
  }
  load();
}

function providerField(f, cfg, onchange) {
  var label = has('pv.field.' + f.name) ? t('pv.field.' + f.name) : f.name;
  var ctl;
  if (f.type === 'select') {
    ctl = selectInput(cfg[f.name] || f.default || '', (f.options || []).map(function (o) { return { value: o, label: has('pv.opt.' + o) ? t('pv.opt.' + o) : o }; }), function (x) { cfg[f.name] = x; onchange(); }, { placeholder: false, label: label });
  } else {
    ctl = textInput(cfg[f.name] == null ? '' : toLatinDigits(String(cfg[f.name])), { ltr: true, mono: true, label: label, inputmode: f.type === 'number' ? 'numeric' : f.type === 'email' ? 'email' : null, placeholder: f.default ? toLatinDigits(f.default) : null, oninput: function (x) { cfg[f.name] = f.type === 'number' ? toLatinDigits(x).trim() : x; onchange(); } });
    ctl.classList.add('text-sm');
    if (f.type === 'secret') ctl.classList.add('eco-secret');
  }
  var hintKey = f.type === 'secret' && cfg[f.name] ? 'pm.secret.hint' : has('pv.hint.' + f.name) ? 'pv.hint.' + f.name : null;
  return field(label, ctl, hintKey ? t(hintKey) : null, { optional: !f.required });
}

function providersPage(container, channel) {
  var frame = tabFrame(t('desc.' + channel));
  clear(container).appendChild(frame.el);
  var body = h('section', { 'data-channel': channel });
  frame.content.appendChild(body);
  var catalog = [];

  function skeleton() {
    clear(frame.actions).appendChild(bar('12rem', 'h-8'));
    var grid = h('div', { class: 'eco-card-grid' });
    for (var i = 0; i < (channel === 'sms' ? 3 : 2); i++) {
      grid.appendChild(card(h('div', { class: 'p-4' },
        h('div', { class: 'flex items-center justify-between mb-4' }, bar('7rem', 'h-4'), bar('2rem', 'h-5 rounded-full')),
        bar('100%', 'h-14 mb-4'),
        skeletonField(), skeletonField(),
        h('div', { class: 'flex justify-end' }, bar('5rem', 'h-8')))));
    }
    clear(body).appendChild(grid);
  }

  function testBox(p) {
    var email = p.channel === 'email';
    var to = textInput('', { ltr: true, mono: true, label: t('pv.test.to'), placeholder: email ? 'name@example.com' : '09xxxxxxxxx', inputmode: email ? 'email' : 'tel' });
    to.classList.add('text-sm');
    var subject = email ? textInput('', { label: t('pv.test.subject'), placeholder: t('pv.test.subject.ph') }) : null;
    if (subject) subject.classList.add('text-sm');
    var msg = textInput(t('pv.test.message.default'), { multiline: true, rows: 2, label: t('pv.test.message') });
    msg.classList.add('text-sm');
    var out = h('div', { 'aria-live': 'polite' });
    var sendBtn = button({ variant: 'link', size: 'sm', icon: 'send', label: t('pv.test.send'), onclick: function () {
      clear(out);
      setBusy(sendBtn, true, t('nt.sending'));
      var payload = { to: toLatinDigits(to.value).trim(), message: msg.value };
      if (subject) payload.subject = subject.value;
      pubApi('POST', '/providers/' + encodeURIComponent(p.id) + '/test', payload).then(function (r) {
        setBusy(sendBtn, false);
        if (r && r.ok && r.sandbox) out.appendChild(h('div', { class: 'eco-sandbox-note rounded-md px-3 py-2 text-sm flex items-start gap-2', 'data-test-result': 'sandbox' }, icon('flask', 'h-4 w-4 shrink-0 mt-0.5'), h('span', {}, h('span', { text: t('pv.test.sandbox') + ' ' }), r.reply ? mono(r.reply) : null)));
        else if (r && r.ok) out.appendChild(h('div', { 'data-test-result': 'ok' }, notice('success', t('pv.test.ok') + (r.reply ? ' ' + r.reply : ''))));
        else out.appendChild(h('div', { 'data-test-result': 'error' }, notice('error', t('pv.test.failed') + ' ' + ((r && r.error) || ''))));
      }, function (err) {
        setBusy(sendBtn, false);
        out.appendChild(notice('error', errorText(err)));
      });
    } });
    var details = h('details', { class: 'eco-test border-t border-secondary-200 dark:border-secondary-700 mt-4 eco-pt-3' },
      h('summary', { class: 'cursor-pointer text-sm font-medium text-secondary-700 dark:text-secondary-200 select-none', text: t('pv.test') }),
      h('div', { class: 'mt-3' },
        h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: t('pv.test.hint') }),
        field(t('pv.test.to'), to),
        subject ? field(t('pv.test.subject'), subject, null, { optional: true }) : null,
        field(t('pv.test.message'), msg),
        h('div', { class: 'flex justify-end -mt-2 mb-2' }, sendBtn),
        out));
    return details;
  }

  function providerCard(p) {
    var cfg = clone(p.config || {}) || {};
    var st = { isActive: !!p.isActive, isDefault: !!p.isDefault, sandbox: !!p.sandbox };
    var errTop = h('div', {});
    var fieldEls = {};
    var fieldsBox = h('div', {});
    (p.fields || []).forEach(function (f) {
      var w = providerField(f, cfg, function () { if (fieldEls[f.name]) fieldEls[f.name].setError(''); });
      fieldEls[f.name] = w;
      fieldsBox.appendChild(w);
    });
    var sandboxText = h('p', { class: 'text-xs mt-1.5' });
    var sandboxBox = h('div', { class: 'rounded-md px-3 py-2.5 mb-4', 'data-sandbox-box': '' });
    var sandboxToggle = toggle(st.sandbox, t('pv.sandbox'), function (on) { st.sandbox = on; paintSandbox(); });
    sandboxBox.appendChild(h('div', { class: 'flex items-center gap-2' }, icon('flask', 'h-4 w-4 shrink-0'), sandboxToggle));
    sandboxBox.appendChild(sandboxText);
    function paintSandbox() {
      sandboxBox.className = 'rounded-md px-3 py-2.5 mb-4 ' + (st.sandbox ? 'eco-sandbox-note' : 'eco-live-note');
      sandboxText.textContent = st.sandbox ? t('pv.sandbox.on') : t('pv.sandbox.off');
    }
    paintSandbox();
    var defToggle = toggle(st.isDefault, t('pv.default'), function (on) { st.isDefault = on; if (on && !st.isActive) { st.isActive = true; activeToggle.set(true); } }, 'sm');
    var activeToggle = toggle(st.isActive, t('pm.active'), function (on) { st.isActive = on; if (!on && st.isDefault) { st.isDefault = false; defToggle.set(false); } }, 'sm');

    var saveBtn = button({ variant: 'primary', size: 'sm', label: t('save'), onclick: function () { save(); } });
    var del = h('div', { class: 'flex items-center gap-2' });
    function resetDelete() { clear(del).appendChild(button({ variant: 'danger', size: 'sm', icon: 'trash', title: t('delete'), onclick: askDelete })); }
    function askDelete() {
      clear(del);
      del.appendChild(h('span', { class: 'text-sm text-secondary-700 dark:text-secondary-200', text: t('delete.confirm') }));
      del.appendChild(button({ variant: 'dangerSolid', size: 'sm', label: t('delete.yes'), onclick: function () {
        pubApi('DELETE', '/providers/' + encodeURIComponent(p.id)).then(function () { toast('success', t('deleted')); load(); }, function (err) { resetDelete(); clear(errTop).appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err)))); });
      } }));
      del.appendChild(button({ variant: 'ghost', size: 'sm', label: t('no'), onclick: resetDelete }));
    }
    resetDelete();

    function save() {
      clear(errTop);
      Object.keys(fieldEls).forEach(function (k) { fieldEls[k].setError(''); });
      var goingLive = p.sandbox && !st.sandbox;
      var ask = goingLive ? confirmModal({ title: t('pv.live.title'), text: t('pv.live.text'), ok: t('pv.live.ok'), danger: true, icon: 'alert' }) : Promise.resolve(true);
      ask.then(function (yes) {
        if (!yes) return;
        var out = {};
        (p.fields || []).forEach(function (f) { if (cfg[f.name] !== undefined && cfg[f.name] !== null) out[f.name] = String(cfg[f.name]); });
        setBusy(saveBtn, true, t('saving'));
        pubApi('PUT', '/providers/' + encodeURIComponent(p.id), { isActive: st.isActive, isDefault: st.isDefault, sandbox: st.sandbox, config: out }).then(function () {
          toast('success', t('saved'));
          load();
        }, function (err) {
          setBusy(saveBtn, false);
          if (err.field && fieldEls[err.field]) fieldEls[err.field].setError(errorText(err));
          else errTop.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
        });
      });
    }

    var el = card(h('div', { class: 'p-4 flex flex-col h-full' },
      h('div', { class: 'flex items-start justify-between gap-3 mb-4' },
        h('div', { class: 'min-w-0' },
          h('div', { class: 'text-lg font-medium', text: kindLabel(p.kind) }),
          h('div', { class: 'flex flex-wrap items-center gap-1.5 mt-1' }, mono(p.kind),
            tag(p.configured ? t('pm.configured') : t('pm.notConfigured'), p.configured ? 'success' : 'secondary'),
            p.isDefault ? tag(t('pv.default'), 'primary') : null)),
        h('div', { class: 'flex flex-col items-end gap-2 shrink-0' }, activeToggle, defToggle)),
      sandboxBox,
      errTop,
      h('div', { class: 'flex-1' }, fieldsBox),
      h('div', { class: 'flex items-center justify-between gap-2 pt-2' }, del, saveBtn),
      testBox(p)));
    el.setAttribute('data-provider', p.kind);
    return el;
  }

  function addControl(existing) {
    var free = catalog.filter(function (k) { return k.channel === channel && existing.indexOf(k.kind) === -1; });
    if (!free.length) return h('span', { class: 'text-xs text-secondary-500', text: t('pv.allAdded') });
    var kind = free[0].kind;
    var sel = h('select', { class: SELECT_SM, 'aria-label': t('pv.kind'), onchange: function () { kind = sel.value; } }, free.map(function (k) { return h('option', { value: k.kind, text: kindLabel(k.kind) }); }));
    var add = button({ variant: 'primary', size: 'page', icon: 'plus', label: t('pv.add'), onclick: function () {
      setBusy(add, true);
      pubApi('POST', '/providers', { channel: channel, kind: kind, sandbox: true, isActive: false, isDefault: false, config: {} }).then(function () {
        toast('success', t('pv.added'));
        load();
      }, function (err) { setBusy(add, false); toast('error', errorText(err)); });
    } });
    return h('div', { class: 'flex items-center gap-2' }, h('label', { class: 'x-label block relative x-input', style: INPUT_VARS }, sel), add);
  }

  function load() {
    skeleton();
    pubApi('GET', '/providers').then(function (res) {
      if (!container.isConnected) return;
      var items = (res && res.items) || [];
      catalog = (res && res.catalog) || [];
      var list = items.filter(function (p) { return p.channel === channel; });
      var grid = h('div', { class: 'eco-card-grid' });
      if (!list.length) grid.appendChild(h('div', { class: 'eco-span-all' }, card(emptyBlock(t('pv.empty.' + channel), t('pv.empty.hint')))));
      list.forEach(function (p) { grid.appendChild(providerCard(p)); });
      clear(frame.actions).appendChild(addControl(list.map(function (p) { return p.kind; })));
      clear(body).appendChild(grid);
    }, function (err) { clear(frame.actions); clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}

var APP_PLATFORMS = ['android', 'ios', 'web', 'windows', 'macos', 'linux'];
var APP_STORES = ['google_play', 'app_store', 'bazaar', 'myket', 'microsoft_store', 'steam', 'web', 'direct'];
var VERSION_RE = /^v?\d{1,9}(\.\d{1,9}){0,3}([-+][0-9A-Za-z.-]{1,40})?$/;
var STORE_URL_RE = /^(https?|market|bazaar|myket|itms-apps):\/\/\S+$/i;
var COUNTRY_RE = /^[A-Za-z]{2}$/;
var OTP_STATUS_COLOR = { pending: 'primary', verified: 'success', expired: 'secondary', failed: 'error', replaced: 'secondary', send_failed: 'error' };

function flagOf(code) {
  code = String(code || '').toUpperCase();
  if (!/^[A-Z]{2}$/.test(code)) return '';
  return String.fromCodePoint(code.charCodeAt(0) - 65 + 0x1F1E6, code.charCodeAt(1) - 65 + 0x1F1E6);
}

function flagBox(it) {
  var box = h('div', { class: 'eco-flag', 'aria-hidden': 'true' });
  if (it.flagUrl) box.appendChild(img(it.flagUrl, 'flag'));
  else box.appendChild(h('span', { text: it.flag || flagOf(it.code) }));
  return box;
}

function countryCell(it) {
  return h('div', { class: 'flex items-center gap-3 min-w-0' }, flagBox(it),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'font-medium truncate', text: i18nText(it.name) || it.code }),
      h('div', { class: 'text-secondary-500' }, mono(it.code))));
}

function countryBody(it, patch) {
  return Object.assign({ name: trClean(it.name || {}), dialCode: it.dialCode || '', flagUrl: it.flagUrl || '', isActive: !!it.isActive, sort: it.sort == null ? 100 : it.sort }, patch || {});
}

function countryActive(it) {
  var box = h('div', { class: 'flex items-center gap-2', 'data-country-toggle': it.code });
  var label = h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: it.isActive ? t('active') : t('inactive') });
  var sw = toggle(it.isActive, null, function (on) {
    baseApi('PUT', '/countries/' + encodeURIComponent(it.code), countryBody(it, { isActive: on })).then(function () {
      it.isActive = on;
      label.textContent = on ? t('active') : t('inactive');
      toast('success', t(on ? 'ct.activated' : 'ct.deactivated', { name: i18nText(it.name) || it.code }));
    }, function (err) {
      sw.set(!on);
      toast('error', errorText(err));
    });
  }, 'sm');
  sw.setAttribute('aria-label', t('f.isActive') + ': ' + (i18nText(it.name) || it.code));
  box.appendChild(sw);
  box.appendChild(label);
  return box;
}

function providerName(kind) {
  return has('pv.kind.' + kind) ? t('pv.kind.' + kind) : kind;
}

function platformLabel(p) {
  return has('platform.' + p) ? t('platform.' + p) : p;
}

function storeLabel(s) {
  return has('store.' + s) ? t('store.' + s) : s;
}

function versionCell(v) {
  return v ? h('span', { class: 'font-mono text-sm', dir: 'ltr', text: v }) : h('span', { class: 'text-secondary-400', text: t('none') });
}

function versionCheck(v) {
  return v && !VERSION_RE.test(String(v).trim()) ? t('code.INVALID_VERSION') : null;
}

function versionParts(v) {
  var s = String(v || '').trim().replace(/^v/, '');
  var pre = '';
  var m = /^([0-9.]+)(?:-([0-9A-Za-z.-]+))?/.exec(s);
  if (!m) return null;
  pre = m[2] || '';
  return { nums: m[1].split('.').map(Number), pre: pre };
}

function compareVersions(a, b) {
  var x = versionParts(a);
  var y = versionParts(b);
  if (!x || !y) return 0;
  for (var i = 0; i < 4; i++) {
    var p = x.nums[i] || 0;
    var q = y.nums[i] || 0;
    if (p !== q) return p < q ? -1 : 1;
  }
  if (x.pre === y.pre) return 0;
  if (!x.pre) return 1;
  if (!y.pre) return -1;
  return x.pre < y.pre ? -1 : 1;
}

function richText(str, nodes) {
  return String(str).split(/(\{\w+\})/).map(function (part) {
    var m = /^\{(\w+)\}$/.exec(part);
    return m && nodes[m[1]] ? nodes[m[1]] : part;
  });
}

function storeLinksEditor(value, onchange) {
  var rows = (value || []).map(function (r) { return { store: r.store || 'web', url: r.url || '', label: trValue(r.label) }; });
  var listEl = h('div', { class: 'flex flex-col gap-2' });
  var errors = [];
  function emit() { onchange(rows.map(function (r) { return { store: r.store, url: r.url, label: Object.assign({}, r.label) }; })); }
  var add = button({ variant: 'link', size: 'sm', icon: 'plus', label: t('av.link.add'), onclick: function () {
    var used = rows.map(function (r) { return r.store; });
    var next = APP_STORES.filter(function (s) { return used.indexOf(s) === -1; })[0] || 'web';
    rows.push({ store: next, url: '', label: {} });
    render();
    emit();
    var inputs = listEl.querySelectorAll('[data-link-url]');
    if (inputs.length) inputs[inputs.length - 1].focus();
  } });
  function render() {
    clear(listEl);
    errors = [];
    add.hidden = rows.length >= 8;
    if (!rows.length) listEl.appendChild(h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400 border border-dashed border-secondary-300 dark:border-secondary-700 rounded-md px-3 py-4 text-center', text: t('av.link.empty') }));
    rows.forEach(function (row, idx) {
      var cell = function (labelKey, ctl) {
        return h('div', { class: 'min-w-0' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t(labelKey) }), ctl);
      };
      var storeSel = selectInput(row.store, APP_STORES.map(function (s) { return { value: s, label: storeLabel(s) }; }), function (v) { row.store = v; emit(); }, { placeholder: false, label: t('av.link.store') });
      var url = textInput(row.url, { ltr: true, mono: true, label: t('av.link.url'), placeholder: 'https://', oninput: function (v) { row.url = v.trim(); emit(); } });
      url.setAttribute('data-link-url', '');
      var label = translatedInput(row.label, { label: t('av.link.label'), oninput: function (v) { row.label = v; emit(); } });
      var err = h('p', { class: 'text-xs text-error-600 dark:text-error-400 eco-items-err', hidden: true });
      errors.push({ row: row, el: err, label: label });
      listEl.appendChild(h('div', { class: 'border border-secondary-200 dark:border-secondary-700 rounded-md p-3', 'data-store-row': '' },
        h('div', { class: 'eco-actions-row eco-store-row' },
          cell('av.link.store', storeSel),
          cell('av.link.url', url),
          h('div', { class: 'min-w-0 eco-actions-label' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('av.link.label') }), label),
          h('div', { class: 'eco-items-remove' }, button({ variant: 'danger', size: 'iconLg', icon: 'trash', title: t('remove'), onclick: function () { rows.splice(idx, 1); render(); emit(); } }))),
        h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1.5' }, richText(t('av.link.hint'), { schemes: h('span', { dir: 'ltr', class: 'font-mono', text: 'bazaar://, myket://, market://' }) })),
        err));
    });
  }
  render();
  var el = h('div', {}, listEl, h('div', { class: 'mt-2' }, add));
  el.validate = function () {
    var bad = false;
    var seen = {};
    errors.forEach(function (e) {
      var r = e.row;
      var msg = '';
      var hit = trCheck(r.label, { max: 64 });
      if (!r.url) msg = t('av.link.url') + ': ' + t('err.required');
      else if (!STORE_URL_RE.test(r.url)) msg = t('code.INVALID_URL');
      else if (seen[r.store]) msg = t('code.DUPLICATE_STORE');
      else if (hit) msg = t('av.link.label') + ': ' + t('code.' + hit.code, { lang: langName(hit.lang) });
      seen[r.store] = true;
      e.el.textContent = msg;
      e.el.hidden = !msg;
      e.el.parentElement.classList.toggle('eco-row-invalid', !!msg);
      e.label.markLang(hit ? hit.lang : null);
      if (msg) bad = true;
    });
    return bad ? t('err.form') : null;
  };
  return el;
}

RESOURCES.countries = {
  api: 'base',
  filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
  columns: [
    { label: 'col.country', render: countryCell, skeleton: '60%' },
    { label: 'col.dialCode', render: function (it) { return it.dialCode ? h('span', { class: 'font-mono text-sm tabular-nums', dir: 'ltr', text: it.dialCode }) : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 130 },
    { label: 'col.status', render: countryActive, width: 150 },
    { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'code', type: 'key', label: 'f.ct.code', hint: 'f.ct.code.hint', required: true, keyRe: COUNTRY_RE, keyErr: 'code.INVALID_COUNTRY_CODE', lockOnEdit: true },
      { name: 'dialCode', type: 'text', ltr: true, label: 'f.ct.dialCode', hint: 'f.ct.dialCode.hint', optional: true, check: function (v) { return v && !/^\+?[1-9][0-9]{0,4}$/.test(String(v).trim()) ? t('code.INVALID_DIAL_CODE') : null; } }
    ] },
    tr('name', { label: 'f.ct.name', required: true, max: 100 }),
    { name: 'flagUrl', type: 'image', label: 'f.ct.flag', hint: 'f.ct.flag.hint', optional: true },
    { type: 'row', fields: [
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.ct.active.hint' },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.ct.sort.hint' }
    ] }
  ],
  defaults: function () { return { code: '', name: {}, dialCode: '', flagUrl: '', isActive: true, sort: 100 }; },
  fixPayload: function (p) {
    p.code = String(p.code || '').trim().toUpperCase();
    p.dialCode = String(p.dialCode || '').trim();
    if (p.dialCode && p.dialCode[0] !== '+') p.dialCode = '+' + p.dialCode;
    if (!p.flagUrl) p.flagUrl = '';
    if (p.sort == null) p.sort = 100;
    return p;
  }
};

RESOURCES['app-versions'] = {
  api: 'base',
  filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
  columns: [
    { label: 'col.platform', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium', text: platformLabel(it.platform) }), h('div', { class: 'text-secondary-500' }, mono(it.platform))); }, skeleton: '45%' },
    { label: 'col.latest', render: function (it) { return versionCell(it.latestVersion); }, width: 130 },
    { label: 'col.minimum', render: function (it) { return versionCell(it.minVersion); }, width: 130 },
    { label: 'col.stores', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, (it.storeLinks || []).map(function (l) { return tag(i18nText(l.label) || storeLabel(l.store), 'secondary'); })); }, skeleton: '40%' },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'platform', type: 'select', label: 'f.av.platform', required: true, options: function () { return APP_PLATFORMS.map(function (p) { return { value: p, label: platformLabel(p) }; }); } },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.av.active.hint' }
    ] },
    { type: 'row', fields: [
      { name: 'latestVersion', type: 'text', ltr: true, label: 'f.av.latest', hint: 'f.av.latest.hint', required: true, check: versionCheck },
      { name: 'minVersion', type: 'text', ltr: true, label: 'f.av.min', hint: 'f.av.min.hint', optional: true, check: function (v, vals) {
        var bad = versionCheck(v);
        if (bad) return bad;
        if (v && vals.latestVersion && VERSION_RE.test(vals.latestVersion) && compareVersions(v, vals.latestVersion) > 0) return t('code.MIN_ABOVE_LATEST');
        return null;
      } }
    ] },
    { type: 'heading', label: 'f.av.stores' },
    { name: 'storeLinks', type: 'links', empty: [] },
    { type: 'heading', label: 'f.av.messages' },
    tr('updateMessage', { label: 'f.av.updateMessage', hint: 'f.av.updateMessage.hint', multiline: true, rows: 2, max: 500, optional: true }),
    tr('forceMessage', { label: 'f.av.forceMessage', hint: 'f.av.forceMessage.hint', multiline: true, rows: 2, max: 500, optional: true })
  ],
  defaults: function () { return { platform: 'android', latestVersion: '1.0.0', minVersion: '', storeLinks: [], updateMessage: {}, forceMessage: {}, isActive: true }; },
  fixPayload: function (p) {
    p.latestVersion = String(p.latestVersion || '').trim();
    p.minVersion = String(p.minVersion || '').trim();
    p.storeLinks = p.storeLinks || [];
    return p;
  }
};

LISTS.otp = {
  api: 'publication',
  noRow: true,
  filters: [{ name: 'status', label: 'filter.status', options: function () { return ['pending', 'verified', 'expired', 'failed', 'replaced', 'send_failed'].map(function (s) { return { value: s, label: t('otp.status.' + s) }; }); } }],
  columns: [
    { label: 'col.phone', render: function (it) { return h('span', { class: 'font-mono text-sm tabular-nums', dir: 'ltr', text: it.phone }); }, skeleton: '40%' },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('otp.status.' + it.status), OTP_STATUS_COLOR[it.status]), it.sandbox ? tag(t('otp.sandbox'), 'warning') : tag(providerName(it.provider), 'secondary')); }, width: 200 },
    { label: 'col.attempts', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: fmtNum(it.attempts || 0) + ' / ' + fmtNum(it.maxAttempts || 0) }); }, width: 110 },
    { label: 'col.player', render: function (it) { return it.userId ? h('a', { href: '#/players/' + encodeURIComponent(it.userId), class: 'eco-link truncate', text: it.username || it.userId }) : h('span', { class: 'text-secondary-400', text: t('none') }); }, skeleton: '30%' },
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 180 }
  ]
};

function apiPage(call, path, title, fields, opts) {
  opts = opts || {};
  var box = card(h('div', { class: 'p-4 md:p-6' }, formSkeleton({ fields: typeof fields === 'function' ? fields({}) : fields })));
  box.setAttribute('data-settings-form', opts.id || '');
  var body = box.firstChild;
  var api2 = {};
  function load() {
    Promise.all([call('GET', path), freshLanguages()]).then(function (all) {
      if (!box.isConnected) return;
      paint(all[0] || {});
    }, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  function paint(data) {
    var values = clone(data) || {};
    var top = h('div', {});
    var extra = h('div', {});
    var form = buildForm(typeof fields === 'function' ? fields(values) : fields, values, { onchange: function (v) { if (opts.onchange) opts.onchange(v, extra); } });
    var lead = form.el.firstElementChild;
    if (lead && lead.classList.contains('border-t')) lead.classList.add('eco-flush');
    var saveBtn = button({ variant: 'primary', size: 'sm', label: t('save'), onclick: function () { save(); } });
    saveBtn.setAttribute('data-settings-save', '');
    clear(body).appendChild(h('form', { novalidate: true, onsubmit: function (e) { e.preventDefault(); save(); } },
      title ? h('h2', { class: 'text-lg font-medium mb-1', text: t(title) }) : null,
      opts.desc ? h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mb-4', text: t(opts.desc) }) : null,
      top, form.el, extra,
      h('div', { class: 'flex justify-end pt-2' }, saveBtn),
      h('button', { type: 'submit', hidden: true })));
    if (opts.onchange) opts.onchange(values, extra);
    function save() {
      clear(top);
      if (!form.validate()) return;
      var payload = form.payload();
      if (opts.fixPayload) payload = opts.fixPayload(payload);
      setBusy(saveBtn, true, t('saving'));
      call('PUT', path, payload).then(function (res) {
        toast('success', t('saved'));
        paint(res || payload);
        if (opts.afterSave) opts.afterSave(res);
      }, function (err) {
        setBusy(saveBtn, false);
        if (!form.showServerError(err)) {
          clear(top).appendChild(h('div', { class: 'mb-5' }, notice('error', errorText(err))));
          top.scrollIntoView({ block: 'start' });
        }
      });
    }
  }
  api2.el = box;
  api2.load = load;
  return api2;
}

var RULE_FIELDS = [
  { type: 'heading', label: 'ar.username' },
  { type: 'row', fields: [
    { name: 'usernameMin', type: 'number', label: 'f.ar.usernameMin', required: true, min: 1, max: 128 },
    { name: 'usernameMax', type: 'number', label: 'f.ar.usernameMax', required: true, min: 1, max: 128, check: function (v, vals) { return v != null && vals.usernameMin != null && Number(v) < Number(vals.usernameMin) ? t('code.MAX_BELOW_MIN') : null; } }
  ] },
  { name: 'usernameCharset', type: 'select', label: 'f.ar.charset', hint: 'f.ar.charset.hint', required: true, options: function () { return [{ value: 'any', label: t('ar.charset.any') }, { value: 'latin', label: t('ar.charset.latin') }]; } },
  { type: 'heading', label: 'ar.password' },
  { name: 'passwordMin', type: 'number', label: 'f.ar.passwordMin', hint: 'f.ar.passwordMin.hint', required: true, min: 8, max: 128 },
  { type: 'row', fields: [
    { name: 'passwordLetter', type: 'bool', inline: true, label: 'f.ar.passwordLetter' },
    { name: 'passwordDigit', type: 'bool', inline: true, label: 'f.ar.passwordDigit' }
  ] },
  { type: 'heading', label: 'ar.display' },
  { name: 'displayNameMax', type: 'number', label: 'f.ar.displayNameMax', required: true, min: 1, max: 255 }
];

function ruleProblems(r, username, password) {
  var out = [];
  var n = Array.from(username || '').length;
  if (username) {
    if (n < r.usernameMin) out.push(t('code.USERNAME_TOO_SHORT', { n: r.usernameMin }));
    else if (n > r.usernameMax) out.push(t('code.USERNAME_TOO_LONG', { n: r.usernameMax }));
    else if (r.usernameCharset === 'latin' && !/^[A-Za-z0-9_.-]+$/.test(username)) out.push(t('code.USERNAME_INVALID_CHARACTERS'));
  }
  if (password) {
    if (Array.from(password).length < r.passwordMin) out.push(t('code.PASSWORD_TOO_SHORT', { n: r.passwordMin }));
    else if (r.passwordLetter && !/\p{L}/u.test(password)) out.push(t('code.PASSWORD_NEEDS_LETTER'));
    else if (r.passwordDigit && !/\p{Nd}/u.test(password)) out.push(t('code.PASSWORD_NEEDS_DIGIT'));
  }
  return out;
}

function rulesTry(extra, values) {
  if (!extra.__try) {
    var result = h('div', { class: 'flex flex-col gap-1 text-sm', 'aria-live': 'polite', 'data-rules-result': '' });
    var user = textInput('', { label: t('ar.try.username'), placeholder: t('ar.try.username') });
    var pass = textInput('', { ltr: true, label: t('ar.try.password'), placeholder: t('ar.try.password') });
    user.setAttribute('data-rules-try', 'username');
    pass.setAttribute('data-rules-try', 'password');
    extra.__try = { result: result, user: user, pass: pass };
    var paintTry = function () { rulesPaint(extra); };
    user.addEventListener('input', paintTry);
    pass.addEventListener('input', paintTry);
    extra.appendChild(h('div', { class: 'eco-try mt-2 mb-4 rounded-md border border-secondary-200 dark:border-secondary-700 p-4' },
      h('div', { class: 'text-sm font-semibold text-secondary-800 dark:text-secondary-200 mb-1', text: t('ar.try') }),
      h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: t('ar.try.hint') }),
      h('div', { class: 'eco-form-row' }, field(t('ar.try.username'), user), field(t('ar.try.password'), pass)),
      result));
  }
  extra.__rules = values;
  rulesPaint(extra);
}

function rulesPaint(extra) {
  var tr2 = extra.__try;
  var v = extra.__rules || {};
  var r = { usernameMin: Number(toLatinDigits(String(v.usernameMin == null ? 1 : v.usernameMin))) || 1, usernameMax: Number(toLatinDigits(String(v.usernameMax == null ? 128 : v.usernameMax))) || 128,
    usernameCharset: v.usernameCharset || 'any', passwordMin: Number(toLatinDigits(String(v.passwordMin == null ? 8 : v.passwordMin))) || 8, passwordLetter: !!v.passwordLetter, passwordDigit: !!v.passwordDigit };
  clear(tr2.result);
  if (!tr2.user.value && !tr2.pass.value) return;
  var problems = ruleProblems(r, tr2.user.value.trim(), tr2.pass.value);
  if (!problems.length) tr2.result.appendChild(h('span', { class: 'eco-text-success inline-flex items-center gap-1.5', 'data-rules-ok': '' }, icon('check', 'h-4 w-4'), h('span', { text: t('ar.try.ok') })));
  problems.forEach(function (p) { tr2.result.appendChild(h('span', { class: 'eco-text-error inline-flex items-center gap-1.5', 'data-rules-bad': '' }, icon('alert', 'h-4 w-4'), h('span', { text: p }))); });
}

function accountRulesPage(container) {
  var frame = tabFrame(t('desc.accounts'));
  clear(container).appendChild(frame.el);
  var page = apiPage(baseApi, '/account-rules', null, RULE_FIELDS, { id: 'accounts', onchange: function (v, extra) { rulesTry(extra, v); } });
  frame.content.appendChild(page.el);
  page.load();
}

function otpModeNotice(v, providers) {
  var list = providers || [];
  var chosen = v.providerId ? list.filter(function (p) { return p.id === v.providerId; })[0] : list.filter(function (p) { return p.isActive; })[0];
  if (!v.isEnabled) return notice('secondary', t('otp.mode.off'));
  if (v.sandbox || (chosen && chosen.sandbox)) return notice('warning', t('otp.mode.sandbox'));
  if (!chosen || !chosen.isActive) return notice('error', t('otp.mode.noProvider'), h('a', { href: hashFor('sms'), class: 'eco-link text-sm', text: t('otp.mode.setup') }));
  return notice('success', t('otp.mode.live', { name: providerName(chosen.kind) }));
}

function otpFields(values) {
  var providers = values.providers || [];
  return [
    { type: 'row', fields: [
      { name: 'isEnabled', type: 'bool', inline: true, label: 'f.otp.enabled', hint: 'f.otp.enabled.hint' },
      { name: 'sandbox', type: 'bool', inline: true, label: 'f.otp.sandbox', hint: 'f.otp.sandbox.hint' }
    ] },
    { name: 'providerId', type: 'select', label: 'f.otp.provider', hint: 'f.otp.provider.hint', options: function () {
      return [{ value: '', label: t('otp.provider.default') }].concat(providers.map(function (p) {
        var name = providerName(p.kind);
        var flags = [];
        if (!p.isActive) flags.push(t('inactive'));
        if (p.sandbox) flags.push(t('otp.sandbox'));
        return { value: p.id, label: name + (flags.length ? ' (' + flags.join('، ') + ')' : '') };
      }));
    }, noPlaceholder: true },
    { type: 'heading', label: 'otp.code' },
    { type: 'row', fields: [
      { name: 'codeLength', type: 'number', label: 'f.otp.codeLength', hint: 'f.otp.codeLength.hint', required: true, min: 4, max: 8 },
      { name: 'ttlSeconds', type: 'number', label: 'f.otp.ttl', hint: 'f.otp.ttl.hint', required: true, min: 30, max: 1800 }
    ] },
    { type: 'row', fields: [
      { name: 'maxAttempts', type: 'number', label: 'f.otp.attempts', hint: 'f.otp.attempts.hint', required: true, min: 1, max: 10 },
      { name: 'resendSeconds', type: 'number', label: 'f.otp.resend', hint: 'f.otp.resend.hint', required: true, min: 10, max: 600 }
    ] },
    { type: 'heading', label: 'otp.limits' },
    { type: 'row', fields: [
      { name: 'phoneHourly', type: 'number', label: 'f.otp.phoneHourly', required: true, min: 1, max: 100 },
      { name: 'phoneDaily', type: 'number', label: 'f.otp.phoneDaily', required: true, min: 1, max: 500, check: function (v, vals) { return v != null && vals.phoneHourly != null && Number(v) < Number(vals.phoneHourly) ? t('code.DAILY_BELOW_HOURLY') : null; } }
    ] },
    { name: 'ipHourly', type: 'number', label: 'f.otp.ipHourly', hint: 'f.otp.ipHourly.hint', required: true, min: 1, max: 10000 },
    { type: 'heading', label: 'otp.message' },
    tr('message', { label: 'f.otp.message', hint: 'f.otp.message.hint', multiline: true, rows: 3, required: true, max: 300 })
  ];
}

function otpPage(container, query, ctx) {
  var list = listPage(container, 'otp', query, ctx);
  var providers = [];
  var head = h('div', { class: 'mb-3' });
  var page = apiPage(pubApi, '/otp/settings', null, function (values) { providers = values.providers || []; return otpFields(values); }, {
    id: 'otp',
    onchange: function (v) { clear(head).appendChild(otpModeNotice(v, providers)); },
    fixPayload: function (p) { if (!p.providerId) p.providerId = ''; return p; },
    afterSave: function () { list.reload(); }
  });
  var heading = h('h2', { class: 'text-lg font-medium mt-6 mb-2', text: t('otp.recent') });
  container.insertBefore(heading, container.firstChild);
  container.insertBefore(page.el, heading);
  container.insertBefore(head, page.el);
  page.load();
  return list;
}

function appVersionsPage(container, query, ctx) {
  return listPage(container, 'app-versions', query, ctx);
}

function countriesPage(container, query, ctx) {
  return listPage(container, 'countries', query, ctx);
}

var RESTRICT_KINDS = ['suspend', 'mute', 'warn'];
var RESTRICT_COLOR = { suspend: 'error', mute: 'warning', warn: 'primary' };
var RESTRICT_STATUS_COLOR = { active: 'error', expired: 'secondary', lifted: 'success' };
var REPORT_STATUS_COLOR = { open: 'warning', dismissed: 'secondary', resolved: 'success' };
var WORD_ACTION_COLOR = { mask: 'primary', block: 'error', flag: 'warning' };
var DURATIONS = [60, 1440, 4320, 10080, 43200];
var REPORT_ACTIONS = ['dismiss', 'delete', 'warn', 'mute', 'suspend'];
var pendingReports = null;
var reportBadgeAt = 0;
var playerSel = { ids: {}, items: {}, bar: null, page: [] };

function refreshReportBadge(force) {
  var now = Date.now();
  if (!force && now - reportBadgeAt < 60000) return;
  reportBadgeAt = now;
  if (!aclCan('CHANNEL_MESSAGE', 0)) return;
  quietBase('GET', '/reports/summary').then(function (res) {
    var n = res && typeof res.open === 'number' ? res.open : null;
    if (n !== pendingReports) {
      pendingReports = n;
      schedule();
    }
  });
}

function durationLabel(min) {
  if (min < 1440) return t('dur.hours', { n: localDigits(min / 60) });
  return t('dur.days', { n: localDigits(min / 1440) });
}

function untilText(iso) {
  return t('rs.until', { date: fmtDate(iso, true) });
}

function playerLink(id, name, extra) {
  if (!id) return h('span', { class: 'text-secondary-400', text: name || t('rp.gone') });
  return h('a', { href: '#/players/' + encodeURIComponent(id), class: 'eco-link truncate' + (extra ? ' ' + extra : ''), text: name || String(id).slice(0, 8) });
}

function playerTags(tags, max) {
  var list = tags || [];
  var box = h('div', { class: 'flex flex-wrap gap-1', 'data-player-tags': '' });
  list.slice(0, max || 6).forEach(function (tg) {
    var chip = tag(i18nText(tg.name) || tg.key, tg.isPublic ? 'success' : 'primary');
    if (tg.isPublic) chip.setAttribute('title', t('tag.public'));
    box.appendChild(chip);
  });
  if (max && list.length > max) box.appendChild(tag('+' + fmtNum(list.length - max), 'secondary'));
  if (!list.length) box.appendChild(h('span', { class: 'text-secondary-400', text: t('none') }));
  return box;
}

function playerStatus(it) {
  var box = h('div', { class: 'flex flex-col items-start gap-1', 'data-player-status': '' });
  if (it.banned) box.appendChild(tag(t('pl.banned'), 'error'));
  if (it.suspension) box.appendChild(h('span', { class: 'eco-tag eco-tag--error inline-flex leading-tight max-w-full px-2 py-0.5 text-xs rounded-md whitespace-nowrap', 'data-status': 'suspended', title: untilText(it.suspension.endsAt), text: t('pl.suspendedUntil', { date: fmtDate(it.suspension.endsAt, true) }) }));
  if (it.mute) box.appendChild(h('span', { class: 'eco-tag eco-tag--warning inline-flex leading-tight max-w-full px-2 py-0.5 text-xs rounded-md whitespace-nowrap', 'data-status': 'muted', title: untilText(it.mute.endsAt), text: t('pl.mutedUntil', { date: fmtDate(it.mute.endsAt, true) }) }));
  if (!box.childNodes.length) box.appendChild(tag(t('pl.ok'), 'success'));
  return box;
}

function durationPicker(onchange) {
  var state = { minutes: 1440, endsAt: null };
  var custom = h('div', { class: 'mt-2', hidden: true });
  var date = dateInput(null, function (v) { state.endsAt = v; emit(); });
  custom.appendChild(date);
  var options = DURATIONS.map(function (m) { return { value: String(m), label: durationLabel(m) }; }).concat([{ value: 'custom', label: t('dur.custom') }]);
  var seg = segmented(options, '1440', function (v) {
    custom.hidden = v !== 'custom';
    if (v === 'custom') { state.minutes = 0; var inp = custom.querySelector('input'); if (inp) inp.focus(); }
    else { state.minutes = Number(v); state.endsAt = null; }
    emit();
  }, t('rs.duration'), 'md');
  seg.setAttribute('data-duration', '');
  function emit() { if (onchange) onchange(value()); }
  function value() { return state.minutes ? { minutes: state.minutes } : { endsAt: state.endsAt }; }
  var el = h('div', {}, seg, custom);
  el.value = value;
  el.check = function () {
    if (state.minutes) return null;
    if (!state.endsAt) return t('code.END_REQUIRED');
    if (new Date(state.endsAt).getTime() <= Date.now() + 30000) return t('code.END_IN_PAST');
    return null;
  };
  return el;
}

function defaultReason(kind) {
  var out = {};
  LANGS.active.forEach(function (l) {
    var cat = MESSAGES[l.code];
    if (cat && cat['rs.reason.default.' + kind]) out[l.code] = cat['rs.reason.default.' + kind];
  });
  return out;
}

function restrictionForm(kind, opts) {
  opts = opts || {};
  var v = { reason: defaultReason(kind), note: '', deleteMessage: false };
  var fields = {};
  var box = h('div', { 'data-restrict-form': kind });
  var dur = null;
  if (kind === 'suspend' || kind === 'mute') {
    dur = durationPicker(function () { fields.duration.setError(''); });
    fields.duration = field(t('rs.duration'), dur, t(kind === 'suspend' ? 'rs.duration.suspend.hint' : 'rs.duration.mute.hint'));
    box.appendChild(fields.duration);
  }
  var reason = translatedInput(v.reason, { label: t('rs.reason'), multiline: true, rows: 2, max: 500, name: 'reason', oninput: function (x) { v.reason = x; fields.reason.setError(''); } });
  fields.reason = field(t('rs.reason'), reason, t('rs.reason.hint.' + kind));
  box.appendChild(fields.reason);
  if (opts.withDelete) {
    box.appendChild(h('div', { class: 'mb-5' }, checkbox(false, t('rp.alsoDelete'), function (on) { v.deleteMessage = on; })));
  }
  var note = textInput('', { multiline: true, rows: 2, label: t('rs.note'), placeholder: t('rs.note.ph'), oninput: function (x) { v.note = x; } });
  fields.note = field(t('rs.note'), note, t('rs.note.hint'), { optional: true });
  box.appendChild(fields.note);
  return {
    el: box,
    validate: function () {
      var ok = true;
      if (dur) {
        var bad = dur.check();
        fields.duration.setError(bad || '');
        if (bad) ok = false;
      }
      var hit = trCheck(v.reason, { required: true, max: 500 });
      if (hit) {
        fields.reason.setError(t('code.' + hit.code, { lang: langName(hit.lang) }));
        reason.markLang(hit.lang);
        ok = false;
      }
      return ok;
    },
    payload: function () {
      var p = { kind: kind, reason: trClean(v.reason), note: v.note.trim() };
      if (dur) Object.assign(p, dur.value());
      if (opts.withDelete) p.deleteMessage = v.deleteMessage;
      return p;
    },
    showError: function (err) {
      var f = err && err.field;
      if (f === 'endsAt' && fields.duration) { fields.duration.setError(errorText(err)); return true; }
      if (f === 'reason') { fields.reason.setError(errorText(err)); return true; }
      if (f === 'note') { fields.note.setError(errorText(err)); return true; }
      return false;
    }
  };
}

function formDialog(opts) {
  return new Promise(function (resolve) {
    var done = false;
    var top = h('div', {});
    function finish(v) {
      if (done) return;
      done = true;
      root.remove();
      document.removeEventListener('keydown', esc, true);
      resolve(v);
    }
    function esc(e) {
      if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); finish(null); return; }
      if (e.key === 'Tab') trapTab(e, box);
    }
    var ok = button({ variant: opts.danger ? 'dangerSolid' : 'primary', size: 'footer', label: opts.ok, onclick: function () { submit(); } });
    ok.setAttribute('data-dialog-ok', '');
    function submit() {
      clear(top);
      if (opts.validate && !opts.validate()) return;
      setBusy(ok, true, t('saving'));
      opts.submit().then(function (res) { finish(res || true); }, function (err) {
        setBusy(ok, false);
        if (!opts.showError || !opts.showError(err)) top.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
      });
    }
    var box = h('div', { class: 'relative w-full eco-dialog eco-dialog--form bg-white dark:bg-secondary-900 rounded-md shadow-lg overflow-hidden flex flex-col', role: 'dialog', 'aria-modal': 'true', 'aria-label': opts.title, 'data-form-dialog': opts.id || '' },
      h('div', { class: 'p-6 overflow-y-auto' },
        h('div', { class: 'flex gap-3 mb-4' },
          h('div', { class: 'shrink-0 ' + (opts.danger ? 'text-error-600 dark:text-error-400' : 'eco-selected') }, icon(opts.icon || 'alert', 'h-6 w-6')),
          h('div', { class: 'min-w-0' },
            h('h2', { class: 'text-lg font-semibold mb-1', text: opts.title }),
            opts.text ? h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: opts.text }) : null)),
        top, opts.body),
      h('div', { class: 'flex gap-3 justify-end bg-secondary-50 dark:bg-secondary-800 px-6 py-3' },
        button({ variant: 'ghost', size: 'footer', label: t('cancel'), onclick: function () { finish(null); } }), ok));
    var holder = h('div', { class: 'fixed z-[70] inset-0 flex items-center justify-center p-4' }, box);
    holder.addEventListener('mousedown', function (e) { if (e.target === holder) finish(null); });
    var root = h('div', { 'data-economy': 'dialog', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'fixed z-[70] inset-0 opacity-60 dark:opacity-70 dark:bg-black bg-secondary-900' }), holder);
    document.body.appendChild(root);
    document.addEventListener('keydown', esc, true);
    var first = box.querySelector('[data-duration] button, textarea, input');
    (first || ok).focus();
  });
}

function restrictDialog(player, kind) {
  return freshLanguages().then(function () {
    var form = restrictionForm(kind);
    var name = player.displayName || player.username || String(player.id).slice(0, 8);
    return formDialog({
      id: 'restrict-' + kind,
      title: t('rs.dialog.' + kind, { name: name }),
      text: t('rs.dialog.' + kind + '.text'),
      icon: kind === 'warn' ? 'alert' : 'ban',
      danger: kind === 'suspend',
      ok: t('rs.do.' + kind),
      body: form.el,
      validate: form.validate,
      showError: form.showError,
      submit: function () { return baseApi('POST', '/players/' + encodeURIComponent(player.id) + '/restrictions', form.payload()); }
    });
  }).then(function (res) {
    if (res) toast('success', t('rs.done.' + kind, { name: player.displayName || player.username || '' }));
    return res;
  });
}

function liftDialog(r) {
  var note = textInput('', { multiline: true, rows: 2, label: t('rs.liftNote'), placeholder: t('rs.note.ph') });
  return formDialog({
    id: 'lift',
    title: t('rs.lift.' + r.kind),
    text: t('rs.lift.text', { name: r.displayName || r.username || '' }),
    icon: 'check',
    ok: t('rs.lift.ok'),
    body: field(t('rs.liftNote'), note, null, { optional: true }),
    submit: function () { return baseApi('POST', '/restrictions/' + encodeURIComponent(r.id) + '/lift', { note: note.value.trim() }); }
  }).then(function (res) {
    if (res) toast('success', t('rs.lifted'));
    return res;
  });
}

function bulkTagBar(reload) {
  var tagIds = [];
  var count = h('span', { class: 'text-sm font-medium tabular-nums', 'aria-live': 'polite' });
  var picker = refMulti({ resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', value: [], label: t('pl.bulk.tags'), placeholder: t('pl.bulk.pick'), onchange: function (v) { tagIds = v; } });
  function run(op, btn) {
    var ids = Object.keys(playerSel.ids);
    if (!ids.length) return;
    if (!tagIds.length) { toast('error', t('code.TAGS_REQUIRED')); return; }
    setBusy(btn, true, t('saving'));
    baseApi('POST', '/players/tags', { userIds: ids, tagIds: tagIds, op: op }).then(function (res) {
      setBusy(btn, false);
      toast('success', t(op === 'add' ? 'pl.bulk.added' : 'pl.bulk.removed', { n: res.players, changed: res.changed }));
      reload();
    }, function (err) {
      setBusy(btn, false);
      toast('error', errorText(err));
    });
  }
  var add = button({ variant: 'primary', size: 'sm', icon: 'plus', label: t('pl.bulk.add'), onclick: function (e) { run('add', e.currentTarget); } });
  var remove = button({ variant: 'danger', size: 'sm', icon: 'trash', label: t('pl.bulk.remove'), onclick: function (e) { run('remove', e.currentTarget); } });
  var pageAll = button({ variant: 'link', size: 'sm', label: t('pl.bulk.page'), onclick: function () {
    playerSel.page.forEach(function (it) { playerSel.ids[it.id] = true; });
    paint();
    reload();
  } });
  var none = button({ variant: 'ghost', size: 'sm', label: t('pl.bulk.clear'), onclick: function () { playerSel.ids = {}; paint(); reload(); } });
  add.setAttribute('data-bulk', 'add');
  remove.setAttribute('data-bulk', 'remove');
  var el = h('div', { class: 'eco-bulk rounded-md border border-secondary-200 dark:border-secondary-700 p-3 mb-3', 'data-bulk-bar': '' },
    h('div', { class: 'flex flex-wrap items-center gap-2 mb-2' }, count, pageAll, none),
    h('div', { class: 'eco-bulk-row' }, h('div', { class: 'min-w-0 flex-1' }, picker), h('div', { class: 'flex flex-wrap gap-2' }, add, remove)));
  function paint() {
    var n = Object.keys(playerSel.ids).length;
    count.textContent = n ? t('pl.bulk.selected', { n: n }) : t('pl.bulk.none');
    add.disabled = remove.disabled = !n;
  }
  el.paint = paint;
  paint();
  return el;
}

function playerCheck(it) {
  var box = checkbox(!!playerSel.ids[it.id], '', function (on) {
    if (on) playerSel.ids[it.id] = true;
    else delete playerSel.ids[it.id];
    if (playerSel.bar) playerSel.bar.paint();
  });
  box.setAttribute('aria-label', t('pl.select', { name: it.displayName || it.username }));
  box.setAttribute('data-select', it.id);
  return box;
}

LISTS.players = {
  api: 'base',
  rowHref: function (it) { return '#/players/' + encodeURIComponent(it.id); },
  onItems: function (items) { playerSel.page = items; },
  filters: [
    { name: 'tagId', label: 'filter.tag', type: 'ref', resource: 'tags', api: baseApi, filter: { keyType: 'user' } },
    { name: 'status', label: 'filter.status', options: function () { return ['suspended', 'muted', 'banned', 'tagged'].map(function (s) { return { value: s, label: t('pl.status.' + s) }; }); } }
  ],
  columns: [
    { label: 'col.select', render: playerCheck, width: 48 },
    { label: 'col.player', render: function (it) {
      return h('div', { class: 'min-w-0' },
        h('div', { class: 'font-medium truncate', text: it.displayName || it.username }),
        h('div', { class: 'text-secondary-500 flex items-center gap-2 min-w-0' }, it.displayName ? h('span', { class: 'truncate', text: it.username }) : null, mono(String(it.id).slice(0, 8))));
    }, skeleton: '50%' },
    { label: 'col.tags', render: function (it) { return playerTags(it.tags, 4); }, skeleton: '40%' },
    { label: 'col.status', render: playerStatus, width: 220 },
    { label: 'col.created', render: function (it) { return dateCell(it.createTime); }, width: 170 }
  ]
};

function playersPage(container, query, ctx) {
  playerSel.ids = {};
  var list = listPage(container, 'players', query, ctx);
  var bar = bulkTagBar(function () { list.reload(); });
  playerSel.bar = bar;
  var table = container.querySelector('.eco-filters');
  var anchor = table ? table.closest('.flex.flex-col.items-start') : null;
  if (anchor && anchor.parentNode) anchor.parentNode.insertBefore(bar, anchor.nextSibling);
  return list;
}

function restrictionReason(it) {
  return h('div', { class: 'min-w-0' },
    h('div', { class: 'eco-clamp', text: i18nText(it.reason) || '' }),
    it.note ? h('div', { class: 'text-xs text-secondary-500 eco-clamp', text: it.note }) : null);
}

LISTS.restrictions = {
  api: 'base',
  noRow: true,
  filters: [
    { name: 'kind', label: 'filter.kind', options: function () { return RESTRICT_KINDS.map(function (k) { return { value: k, label: t('rs.kind.' + k) }; }); } },
    { name: 'status', label: 'filter.status', options: function () { return ['active', 'expired', 'lifted'].map(function (s) { return { value: s, label: t('rs.status.' + s) }; }); } },
    { name: 'userId', label: 'filter.playerId', type: 'text' }
  ],
  columns: [
    { label: 'col.player', render: function (it) { return h('div', { class: 'min-w-0' }, playerLink(it.userId, it.displayName || it.username, 'font-medium block'), mono(String(it.userId || '').slice(0, 8))); }, skeleton: '40%' },
    { label: 'col.kind', render: function (it) { return tag(t('rs.kind.' + it.kind), RESTRICT_COLOR[it.kind]); }, width: 110 },
    { label: 'col.status', render: function (it) {
      if (it.kind === 'warn') return h('span', { class: 'text-secondary-400', text: t('none') });
      return h('div', { class: 'flex flex-col items-start gap-0.5' },
        tag(t('rs.status.' + it.status), RESTRICT_STATUS_COLOR[it.status]),
        h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: it.status === 'lifted' && it.liftedAt ? t('rs.liftedAt', { date: fmtDate(it.liftedAt, true) }) : untilText(it.endsAt) }));
    }, width: 190 },
    { label: 'col.reason', render: restrictionReason, skeleton: '60%' },
    { label: 'col.by', render: function (it) { return h('div', { class: 'flex flex-col' }, h('span', { class: 'text-sm', text: it.createdBy || t('none') }), dateCell(it.createdAt)); }, width: 170 },
    { label: 'col.actions', render: function (it) {
      if (it.kind === 'warn' || it.status !== 'active') return h('span', {});
      var b = button({ variant: 'link', size: 'sm', icon: 'check', label: t('rs.lift'), onclick: function () {
        liftDialog(it).then(function (res) { if (res) psOrPageReload('restrictions'); });
      } });
      b.setAttribute('data-lift', it.id);
      return b;
    }, width: 140 }
  ]
};

var listReloads = {};

function psOrPageReload(name) {
  if (listReloads[name]) listReloads[name]();
}

function restrictionsPage(container, query, ctx) {
  var list = listPage(container, 'restrictions', query, ctx);
  listReloads.restrictions = list.reload;
  return list;
}

function reportMessage(it) {
  return h('div', { class: 'min-w-0' },
    h('div', { class: 'eco-clamp', dir: 'auto', text: it.content || '' }),
    h('div', { class: 'text-xs text-secondary-500 flex items-center gap-1.5 min-w-0' }, icon('users', 'h-3 w-3'), h('span', { class: 'truncate', text: it.senderName || it.senderUsername || t('rp.gone') }), it.messageDeleted ? tag(t('rp.deleted'), 'secondary') : null));
}

function reportReason(it) {
  if (it.source === 'filter') {
    return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('rp.source.filter'), 'warning'), (it.matched || []).slice(0, 3).map(function (m) { return tag(m.word, 'secondary'); }));
  }
  return h('div', { class: 'flex flex-wrap gap-1' }, tag(i18nText(it.title) || it.titleKey || t('rp.noTitle'), 'primary'));
}

LISTS.reports = {
  api: 'base',
  filters: [
    { name: 'status', label: 'filter.status', options: function () { return ['open', 'dismissed', 'resolved'].map(function (s) { return { value: s, label: t('rp.status.' + s) }; }); } },
    { name: 'source', label: 'filter.source', options: function () { return ['player', 'filter'].map(function (s) { return { value: s, label: t('rp.source.' + s) }; }); } },
    { name: 'titleId', label: 'filter.reportTitle', type: 'ref', resource: 'report-titles', api: baseApi },
    { name: 'senderId', label: 'filter.playerId', type: 'text' }
  ],
  columns: [
    { label: 'col.message', render: reportMessage, skeleton: '70%' },
    { label: 'col.reason', render: reportReason, width: 200 },
    { label: 'col.reporter', render: function (it) { return it.source === 'filter' ? h('span', { class: 'text-secondary-500', text: t('rp.byFilter') }) : playerLink(it.reporterId, it.reporterName || it.reporterUsername); }, width: 160 },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-col items-start gap-0.5' }, tag(t('rp.status.' + it.status), REPORT_STATUS_COLOR[it.status]), it.reports > 1 ? h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: t('rp.reportsN', { n: it.reports }) }) : null); }, width: 130 },
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 170 }
  ]
};

function contextView(lines, reportedId) {
  var box = h('ol', { class: 'eco-chat flex flex-col gap-2', 'data-report-context': '' });
  (lines || []).forEach(function (l) {
    var hit = l.reported || l.id === reportedId;
    box.appendChild(h('li', { class: 'eco-chat-line' + (hit ? ' eco-chat-hit' : ''), 'data-reported': hit ? 'true' : null },
      h('div', { class: 'flex items-center gap-2 text-xs text-secondary-500 mb-0.5' },
        playerLink(l.senderId, l.username, 'font-medium'),
        h('span', { class: 'whitespace-nowrap', text: fmtDate(l.createdAt, true) }),
        hit ? tag(t('rp.thisMessage'), 'warning') : null),
      h('div', { class: 'text-sm text-secondary-800 dark:text-secondary-200 break-words', dir: 'auto', text: l.content })));
  });
  if (!box.childNodes.length) box.appendChild(h('li', { class: 'text-sm text-secondary-500', text: t('rp.noContext') }));
  return box;
}

function reportDetail(container, id) {
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('reports', null, lastQuery.reports || {}); } });
  var frame = pageFrame(t('rp.title'), null, null);
  frame.actions.appendChild(backBtn);
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-detail-grid', 'data-report': id });
  frame.content.appendChild(grid);

  function skeleton() {
    clear(grid);
    [6, 5, 5, 4].forEach(function (n, i) {
      var rows = [];
      for (var j = 0; j < n; j++) rows.push(h('div', { class: 'py-1.5' }, bar(['70%', '50%', '40%'][j % 3])));
      grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), rows), i === 0 ? 'eco-span-all' : null));
    });
  }

  function decision(r) {
    if (r.status !== 'open') {
      return section('rp.decision', dl([
        ['col.status', tag(t('rp.status.' + r.status), REPORT_STATUS_COLOR[r.status])],
        ['rp.action', r.action ? tag(t('rp.action.' + r.action), r.action === 'dismiss' ? 'secondary' : 'primary') : null],
        ['rp.messageDeleted', r.messageDeleted ? tag(t('yes'), 'secondary') : null],
        ['rp.restriction', r.restrictionId ? h('a', { href: hashFor('restrictions', null, { userId: r.senderId }), class: 'eco-link', text: t('rp.seeRestriction') }) : null],
        ['rp.resolvedBy', r.resolvedBy ? h('span', { text: r.resolvedBy }) : null],
        ['rp.resolvedAt', r.resolvedAt ? dateCell(r.resolvedAt) : null],
        ['rs.note', r.resolutionNote ? h('span', { class: 'break-words', text: r.resolutionNote }) : null]
      ]));
    }
    var current = 'dismiss';
    var formBox = h('div', { class: 'mt-4' });
    var form = null;
    var note = null;
    var seg = segmented(REPORT_ACTIONS.map(function (a) { return { value: a, label: t('rp.action.' + a) }; }), current, function (a) { current = a; paintForm(); }, t('rp.action'), 'md');
    seg.setAttribute('data-report-actions', '');
    var hint = h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mt-2' });
    var top = h('div', {});
    var submitBtn = button({ variant: 'primary', size: 'footer', label: t('rp.submit'), onclick: function () { submit(); } });
    submitBtn.setAttribute('data-report-submit', '');
    function paintForm() {
      clear(formBox);
      clear(top);
      hint.textContent = t('rp.action.' + current + '.hint');
      form = null;
      note = null;
      if (current === 'warn' || current === 'mute' || current === 'suspend') {
        form = restrictionForm(current, { withDelete: true });
        formBox.appendChild(form.el);
      } else {
        note = textInput('', { multiline: true, rows: 2, label: t('rs.note'), placeholder: t('rs.note.ph') });
        formBox.appendChild(field(t('rs.note'), note, t('rp.note.hint'), { optional: true }));
      }
      submitBtn.querySelector('span').textContent = t('rp.submit.' + current);
    }
    function submit() {
      clear(top);
      if (form && !form.validate()) return;
      var body = form ? form.payload() : { note: note.value.trim() };
      body.action = current;
      delete body.kind;
      setBusy(submitBtn, true, t('saving'));
      baseApi('POST', '/reports/' + encodeURIComponent(id) + '/resolve', body).then(function () {
        toast('success', t('rp.done.' + current));
        refreshReportBadge(true);
        load();
      }, function (err) {
        setBusy(submitBtn, false);
        if (!form || !form.showError(err)) top.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
      });
    }
    paintForm();
    return card(h('div', { class: 'p-4', 'data-report-decide': '' },
      h('h2', { class: 'text-lg font-medium mb-3', text: t('rp.decide') }),
      top, seg, hint, formBox,
      h('div', { class: 'flex justify-end' }, submitBtn)), 'eco-span-all');
  }

  function render(r) {
    clear(grid);
    var head = h('div', { class: 'flex flex-wrap items-center gap-2' },
      tag(t('rp.status.' + r.status), REPORT_STATUS_COLOR[r.status]),
      r.source === 'filter' ? tag(t('rp.source.filter'), 'warning') : tag(t('rp.source.player'), 'secondary'),
      r.messageDeleted ? tag(t('rp.deleted'), 'secondary') : null);
    grid.appendChild(section('rp.conversation', h('div', {}, h('div', { class: 'mb-3' }, head), contextView(r.context, r.messageId)), null));
    grid.lastChild.classList.add('eco-span-all');
    var others = (r.others || []).filter(function (o) { return o.source === 'player'; });
    grid.appendChild(section('rp.report', dl([
      ['rp.reason', reportReason(r)],
      ['col.reporter', r.source === 'filter' ? h('span', { text: t('rp.byFilter') }) : playerLink(r.reporterId, r.reporterName || r.reporterUsername)],
      ['rp.playerNote', r.note ? h('span', { class: 'break-words', dir: 'auto', text: r.note }) : null],
      ['rp.matched', (r.matched || []).length ? h('div', { class: 'flex flex-wrap gap-1' }, r.matched.map(function (m) { return tag(m.word, WORD_ACTION_COLOR[m.action] || 'secondary'); })) : null],
      ['rp.others', others.length ? h('div', { class: 'flex flex-col gap-1' }, others.map(function (o) { return h('div', { class: 'text-sm' }, playerLink(o.reporterId, o.reporterName || o.reporterUsername), h('span', { class: 'text-secondary-500', text: ' · ' + (i18nText(o.title) || o.titleKey || '') })); })) : null],
      ['col.created', dateCell(r.createdAt)],
      ['rp.channel', r.channelLabel ? mono(r.channelLabel) : null]
    ])));
    var s = r.sender;
    grid.appendChild(section('rp.sender', s ? dl([
      ['col.player', playerLink(s.id, s.displayName || s.username, 'font-medium')],
      ['col.tags', playerTags(s.tags)],
      ['col.status', playerStatus(s)],
      ['rp.history', h('div', { class: 'flex flex-wrap gap-1' },
        tag(t('rp.count.reports', { n: s.counts.reports }), 'secondary'),
        tag(t('rp.count.warn', { n: s.counts.warn }), 'primary'),
        tag(t('rp.count.mute', { n: s.counts.mute }), 'warning'),
        tag(t('rp.count.suspend', { n: s.counts.suspend }), 'error'))],
      ['rp.links', h('div', { class: 'flex flex-wrap gap-3' },
        h('a', { href: hashFor('restrictions', null, { userId: s.id }), class: 'eco-link text-sm', text: t('rp.seeHistory') }),
        h('a', { href: hashFor('reports', null, { senderId: s.id }), class: 'eco-link text-sm', text: t('rp.seeReports') }))]
    ]) : notice('secondary', t('rp.gone'))));
    grid.appendChild(decision(r));
  }

  function load() {
    skeleton();
    baseApi('GET', '/reports/' + encodeURIComponent(id)).then(function (r) {
      if (!grid.isConnected) return;
      Promise.resolve(freshLanguages()).then(function () { render(r); });
    }, function (err) {
      clear(grid).appendChild(errorBlock(err, load));
    });
  }
  load();
}

function wordTry() {
  var input = textInput('', { multiline: true, rows: 2, label: t('cw.try.label'), placeholder: t('cw.try.ph') });
  input.setAttribute('data-word-try', '');
  var result = h('div', { class: 'flex flex-col gap-2 mt-3 text-sm', 'aria-live': 'polite', 'data-word-result': '' });
  var seq = 0;
  var run = debounce(function () {
    var text = input.value;
    var my = ++seq;
    clear(result);
    if (!text.trim()) return;
    result.appendChild(bar('60%'));
    baseApi('POST', '/chat-words/test', { text: text }).then(function (res) {
      if (my !== seq) return;
      clear(result);
      var verdict = res.blocked ? notice('error', t('cw.try.blocked')) : res.flagged ? notice('warning', res.changed ? t('cw.try.maskedFlagged') : t('cw.try.flagged')) : res.changed ? notice('success', t('cw.try.masked')) : notice('secondary', t('cw.try.clean'));
      result.appendChild(verdict);
      result.appendChild(h('div', { class: 'eco-try-out rounded-md px-3 py-2' },
        h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('cw.try.shown') }),
        h('div', { class: 'break-words', dir: 'auto', 'data-word-shown': '', text: res.blocked ? t('cw.try.notSent') : res.text })));
      result.appendChild(h('div', { class: 'text-xs text-secondary-500' }, h('span', { text: t('cw.try.normalized') + ' ' }), h('span', { class: 'font-mono', dir: 'auto', 'data-word-normalized': '', text: res.normalized || '' })));
      if ((res.hits || []).length) result.appendChild(h('div', { class: 'flex flex-wrap gap-1' }, res.hits.map(function (hh) { return tag(hh.word + ' · ' + t('cw.action.' + hh.action), WORD_ACTION_COLOR[hh.action]); })));
    }, function (err) {
      if (my !== seq) return;
      clear(result).appendChild(notice('error', errorText(err)));
    });
  }, 300);
  input.addEventListener('input', run);
  return card(h('div', { class: 'p-4' },
    h('h2', { class: 'text-base font-medium mb-1', text: t('cw.try') }),
    h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mb-3', text: t('cw.try.hint') }),
    input, result), 'mb-4 eco-word-try');
}

RESOURCES['chat-words'] = {
  api: 'base',
  filters: [
    { name: 'action', label: 'filter.action', options: function () { return ['mask', 'block', 'flag'].map(function (a) { return { value: a, label: t('cw.action.' + a) }; }); } },
    { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }
  ],
  columns: [
    { label: 'col.word', render: function (it) {
      return h('div', { class: 'min-w-0' },
        h('div', { class: 'font-medium truncate', dir: 'auto', text: it.word }),
        it.normalized && it.normalized !== it.word ? h('div', { class: 'text-xs text-secondary-500 font-mono truncate', dir: 'auto', text: it.normalized }) : null);
    }, skeleton: '40%' },
    { label: 'col.action', render: function (it) { return tag(t('cw.action.' + it.action), WORD_ACTION_COLOR[it.action]); }, width: 140 },
    { label: 'col.match', render: function (it) { return h('span', { class: 'text-sm', text: t('cw.match.' + it.match) }); }, width: 160 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 },
    { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 170 }
  ],
  fields: [
    { name: 'word', type: 'text', label: 'f.cw.word', hint: 'f.cw.word.hint', required: true, check: function (v) { return v && Array.from(v.trim()).length > 64 ? t('code.INVALID_WORD') : null; } },
    { type: 'row', fields: [
      { name: 'action', type: 'select', label: 'f.cw.action', hint: 'f.cw.action.hint', required: true, options: function () { return ['mask', 'block', 'flag'].map(function (a) { return { value: a, label: t('cw.action.' + a) }; }); } },
      { name: 'match', type: 'select', label: 'f.cw.match', hint: 'f.cw.match.hint', required: true, options: function () { return ['word', 'contains'].map(function (a) { return { value: a, label: t('cw.match.' + a) }; }); } }
    ] },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.cw.active.hint' },
    { name: 'note', type: 'textarea', label: 'f.cw.note', hint: 'f.cw.note.hint', optional: true }
  ],
  defaults: function () { return { word: '', action: 'mask', match: 'word', isActive: true, note: '' }; },
  fixPayload: function (p) { p.word = String(p.word || '').trim(); p.note = String(p.note || '').trim(); return p; }
};

function chatWordsPage(container, query, ctx) {
  var list = listPage(container, 'chat-words', query, ctx);
  var frameContent = container.querySelector('.flex.flex-col.flex-1.min-w-0.px-4');
  if (frameContent) frameContent.insertBefore(wordTry(), frameContent.firstChild);
  return list;
}

RESOURCES['report-titles'] = {
  api: 'base',
  settings: true,
  filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, false); }, skeleton: '50%' },
    { label: 'col.used', render: function (it) { return it.used ? h('a', { href: hashFor('reports', null, { titleId: it.id }), class: 'eco-link tabular-nums', text: t('rp.reportsN', { n: it.used }) }) : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 140 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 },
    { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.rt.key.hint', required: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { label: 'f.rt.name', hint: 'f.rt.name.hint', required: true, max: 100 }),
    tr('description', { label: 'f.rt.description', hint: 'f.rt.description.hint', multiline: true, max: 300, optional: true }),
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.rt.active.hint' }
  ],
  defaults: function () { return { key: '', name: {}, description: {}, isActive: true, sort: 100 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 100; return p; }
};

var PLAYER_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function playerCardRoute() {
  var m = /^#\/players\/([0-9a-f-]{36})(?:\/account)?(?:\?.*)?$/.exec(location.hash || '');
  return m ? m[1] : null;
}

function loadPlayerCard(id) {
  PLAYER_CARD.loading = true;
  return baseApi('GET', '/players/' + encodeURIComponent(id)).then(function (res) {
    PLAYER_CARD.loading = false;
    PLAYER_CARD.at = Date.now();
    PLAYER_CARD.data = res;
    PLAYER_CARD.err = null;
    var old = document.querySelector('[data-player-mod]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    PLAYER_CARD.loading = false;
    PLAYER_CARD.at = Date.now();
    PLAYER_CARD.err = err;
    schedule();
  });
}

function editPlayerTags(p) {
  var ids = (p.tags || []).map(function (x) { return x.id; });
  (p.tags || []).forEach(function (x) { refCache['tags:id:' + x.id] = x; });
  var picker = refMulti({ resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', value: ids, label: t('pc.tags'), placeholder: t('pl.bulk.pick'), onchange: function (v) { ids = v; } });
  return formDialog({
    id: 'player-tags',
    title: t('pc.tags.edit'),
    text: t('pc.tags.edit.text'),
    icon: 'tags',
    ok: t('save'),
    body: field(t('pc.tags'), picker, t('pc.tags.hint')),
    submit: function () { return baseApi('PUT', '/players/' + encodeURIComponent(p.id) + '/tags', { tagIds: ids }); }
  }).then(function (res) {
    if (res) toast('success', t('pc.tags.saved'));
    return res;
  });
}

function playerModCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-mod]');
  if (!id || !token() || !aclCan('ACCOUNT', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (PLAYER_CARD.id !== id) { PLAYER_CARD.id = id; PLAYER_CARD.data = null; PLAYER_CARD.err = null; PLAYER_CARD.at = 0; }
  if (!PLAYER_CARD.loading && Date.now() - PLAYER_CARD.at > 20000) loadPlayerCard(id);
  var p = PLAYER_CARD.data;
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify([p.tags, p.suspension, p.mute, p.banned, (p.history || []).length]) : PLAYER_CARD.err ? 'err' : 'loading');
  if (existing && existing.getAttribute('data-key') === key && existing.parentNode === grid && grid.firstChild === existing) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-mod': '', 'data-key': key, 'data-economy': 'player-mod', 'data-l10n-skip': '', 'aria-labelledby': 'player-mod-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-mod-title', class: 'font-medium', text: t('pc.title') });
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(PLAYER_CARD.err ? notice('error', errorText(PLAYER_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { PLAYER_CARD.at = 0; loadPlayerCard(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    grid.insertBefore(cardEl, grid.firstChild);
    return;
  }
  var write = aclCan('ACCOUNT', 1);
  var after = function (res) { if (res) { PLAYER_CARD.at = 0; loadPlayerCard(id); } };
  var actions = h('div', { class: 'flex flex-wrap items-center gap-2' });
  if (write) {
    if (p.suspension) actions.appendChild(liftButton(p.suspension.id, 'suspend', p, after));
    else actions.appendChild(actionButton('suspend', 'dangerSolid', function () { restrictDialog(p, 'suspend').then(after); }));
    if (p.mute) actions.appendChild(liftButton(p.mute.id, 'mute', p, after));
    else actions.appendChild(actionButton('mute', 'ghost', function () { restrictDialog(p, 'mute').then(after); }));
    actions.appendChild(actionButton('warn', 'ghost', function () { restrictDialog(p, 'warn').then(after); }));
  }
  var editTags = write ? button({ variant: 'link', size: 'sm', icon: 'tags', label: t('pc.tags.edit'), onclick: function () { editPlayerTags(p).then(after); } }) : null;
  if (editTags) editTags.setAttribute('data-player-tags-edit', '');
  var lastWarn = (p.history || []).filter(function (x) { return x.kind === 'warn'; })[0];
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-3' }, title, actions));
  body.appendChild(h('div', { class: 'eco-pc-grid' },
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('pc.tags') }),
      h('div', { class: 'flex flex-wrap items-center gap-2' }, playerTags(p.tags), editTags)),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('pc.status') }),
      playerStatus(p),
      p.suspension && i18nText(p.suspension.reason) ? h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300 mt-1 eco-clamp', text: i18nText(p.suspension.reason) }) : null,
      p.mute && i18nText(p.mute.reason) ? h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300 mt-1 eco-clamp', text: i18nText(p.mute.reason) }) : null),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('pc.history') }),
      h('div', { class: 'flex flex-wrap gap-3 text-sm' },
        h('a', { href: hashFor('restrictions', null, { userId: id }), class: 'eco-link', 'data-player-history': '', text: t('pc.historyN', { n: (p.history || []).length }) }),
        h('a', { href: hashFor('reports', null, { senderId: id }), class: 'eco-link', text: t('pc.reportsN', { n: p.reports ? p.reports.against : 0 }) })),
      lastWarn ? h('p', { class: 'text-xs text-secondary-500 mt-1', text: t('pc.lastWarn', { date: fmtDate(lastWarn.createdAt, true) }) }) : null)));
  grid.insertBefore(cardEl, grid.firstChild);
}

function actionButton(kind, variant, onclick) {
  var b = button({ variant: variant, size: 'sm', icon: kind === 'warn' ? 'alert' : 'ban', label: t('rs.do.' + kind), onclick: onclick });
  b.setAttribute('data-restrict', kind);
  return b;
}

function liftButton(rid, kind, p, after) {
  var b = button({ variant: 'link', size: 'sm', icon: 'check', label: t('rs.lift.' + kind), onclick: function () {
    liftDialog({ id: rid, kind: kind, username: p.username, displayName: p.displayName }).then(after);
  } });
  b.setAttribute('data-lift-kind', kind);
  return b;
}

ICONS['event-types'] = '<path d="M13 2 3 14h9l-1 8 10-12h-9z"/>';
ICONS['event-rules'] = '<circle cx="12" cy="8" r="6"/><path d="M8.2 13.2 7 22l5-3 5 3-1.2-8.8"/>';
ICONS['meta-rules'] = '<path d="M3 3v18h18"/><path d="m7 15 4-4 3 3 5-6"/>';
ICONS['level-tables'] = '<path d="M4 20h4v-6H4zM10 20h4V9h-4zM16 20h4V4h-4z"/>';
ICONS.streaks = '<path d="M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.4-.5-2-1-3-1.1-2.1-.2-4 2-6 .5 2.5 2 4.9 4 6.5 2 1.6 3 3.5 3 5.5a7 7 0 1 1-14 0c0-1.2.4-2.3 1-3.4A2.5 2.5 0 0 0 8.5 14.5z"/>';
ICONS['event-log'] = '<path d="M12 8v4l3 2"/><circle cx="12" cy="12" r="9"/>';

var EVENT_KEY_RE = /^[a-z][a-z0-9_.]{1,47}$/;
var REF_NAMES = {};

function refName(resource, id, call) {
  var key = resource + ':' + id;
  if (!REF_NAMES[key]) {
    REF_NAMES[key] = { item: undefined };
    REF_NAMES[key].p = (call || api)('GET', '/' + resource + '/' + encodeURIComponent(id)).then(function (it) {
      REF_NAMES[key].item = it;
      refCache[resource + ':id:' + id] = it;
      return it;
    }, function () { REF_NAMES[key].item = null; return null; });
  }
  return REF_NAMES[key];
}

function refNameCell(resource, id, opts) {
  opts = opts || {};
  var box = h('span', { class: 'min-w-0 truncate inline-flex items-center gap-1.5', 'data-ref-name': resource });
  if (!id) { box.appendChild(h('span', { class: 'text-secondary-400', text: t('none') })); return box; }
  var ref = refName(resource, id, opts.api);
  function paint(it) {
    clear(box);
    var name = it ? (i18nText(it.name) || it.key) : String(id).slice(0, 8);
    box.appendChild(h('span', { class: 'truncate', text: name }));
    if (opts.withKey && it && it.key) box.appendChild(h('span', { class: 'font-mono text-xs text-secondary-500', dir: 'ltr', text: it.key }));
  }
  if (ref.item !== undefined) paint(ref.item);
  else {
    box.appendChild(bar('5rem'));
    ref.p.then(paint);
  }
  return box;
}

function forgetRefs(resource) {
  Object.keys(REF_NAMES).forEach(function (k) { if (k.indexOf(resource + ':') === 0) delete REF_NAMES[k]; });
}

function numLocal(v) {
  return v === null || v === undefined || v === '' ? '' : fmtNum(v);
}

function ruleSentence(it, eventName) {
  var parts = [t('rs.when.' + it.measure, { event: eventName, n: numLocal(it.threshold), period: t('rs.period.' + it.period) })];
  parts.push(it.repeat === 'every' ? t('rs.repeat.every', { n: numLocal(it.threshold) }) : t('rs.repeat.' + it.repeat + (it.repeat === 'period' ? '.' + it.period : '')));
  if (it.maxClaims) parts.push(t('rs.max', { n: it.maxClaims }));
  return parts.join(' · ');
}

function ruleSentenceCell(it) {
  var box = h('div', { class: 'text-sm text-secondary-700 dark:text-secondary-200 eco-clamp', 'data-rule-sentence': '' });
  if (!it.eventTypeId) return box;
  var ref = refName('event-types', it.eventTypeId);
  function paint(et) { clear(box).appendChild(document.createTextNode(ruleSentence(it, et ? i18nText(et.name) || et.key : '…'))); }
  if (ref.item !== undefined) paint(ref.item);
  else { box.appendChild(bar('70%')); ref.p.then(paint); }
  return box;
}

function grantTag(mode) {
  return tag(t('grant.' + mode), mode === 'auto' ? 'success' : 'primary');
}

function windowCell(it) {
  if (!it.startsAt && !it.endsAt) return h('span', { class: 'text-secondary-400', text: t('ev.window.always') });
  return h('div', { class: 'flex flex-col text-xs text-secondary-600 dark:text-secondary-300 whitespace-nowrap' },
    it.startsAt ? h('span', { text: t('ev.window.from', { date: fmtDate(it.startsAt) }) }) : null,
    it.endsAt ? h('span', { text: t('ev.window.to', { date: fmtDate(it.endsAt) }) }) : null);
}

function rowsEditor(value, f, onchange) {
  var rows = (value || []).map(function (r) { return Object.assign({}, r); });
  var listEl = h('div', { class: 'flex flex-col gap-2' });
  var errors = [];
  var narrow = f.columns.filter(function (c) { return !c.wide; }).length;
  function clean(r) {
    var o = {};
    f.columns.forEach(function (c) {
      var v = r[c.name];
      if (c.type === 'i18n') v = trClean(v);
      o[c.name] = v === undefined || v === '' ? null : v;
    });
    if (r.id) o.id = r.id;
    return o;
  }
  function emit() { onchange(rows.map(clean)); }
  var add = button({ variant: 'link', size: 'sm', icon: 'plus', label: t(f.addLabel || 'add.row'), onclick: function () {
    rows.push(f.newRow ? f.newRow(rows) : {});
    render();
    emit();
    var last = listEl.lastElementChild;
    var first = last && last.querySelector('input,select,button');
    if (first) first.focus();
  } });
  add.setAttribute('data-rows-add', f.name);
  function control(c, row) {
    var label = t(c.label);
    if (c.type === 'number') return numberInput(row[c.name], { int: c.int !== false, label: label, oninput: function (v) { row[c.name] = v === '' ? null : parseNumber(v, c.int !== false); emit(); } });
    if (c.type === 'ref') return refPicker({ resource: c.resource, api: c.api, filter: c.filter, valueKey: c.valueKey, value: row[c.name], required: c.required, label: label, onchange: function (v) { row[c.name] = v; emit(); } });
    if (c.type === 'text') return textInput(row[c.name] || '', { ltr: c.ltr, mono: c.mono, label: label, oninput: function (v) { row[c.name] = v.trim(); emit(); } });
    if (c.type === 'select') {
      if (row[c.name] == null) row[c.name] = c.options()[0].value;
      return selectInput(row[c.name], c.options(), function (v) { row[c.name] = v; emit(); }, { placeholder: false, label: label });
    }
    var ti = translatedInput(row[c.name], { label: label, max: c.max, oninput: function (v) { row[c.name] = v; emit(); } });
    row._tr = ti;
    return ti;
  }
  function render() {
    clear(listEl);
    errors = [];
    add.hidden = !!(f.max && rows.length >= f.max);
    if (!rows.length) listEl.appendChild(h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400 border border-dashed border-secondary-300 dark:border-secondary-700 rounded-md px-3 py-4 text-center', text: t(f.emptyText || 'rows.empty') }));
    rows.forEach(function (row, idx) {
      var cells = f.columns.map(function (c) {
        return h('div', { class: 'min-w-0' + (c.wide ? ' eco-rows-wide' : ''), 'data-col': c.name },
          h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t(c.label) + (c.optional ? ' (' + t('optional') + ')' : '') }),
          control(c, row));
      });
      var err = h('p', { class: 'text-xs text-error-600 dark:text-error-400 eco-items-err', hidden: true });
      errors.push({ row: row, el: err });
      listEl.appendChild(h('div', { class: 'border border-secondary-200 dark:border-secondary-700 rounded-md p-3', 'data-row': String(idx) },
        h('div', { class: 'eco-rows-row', style: '--eco-rows-cols:' + narrow },
          cells,
          h('div', { class: 'eco-items-remove' }, button({ variant: 'danger', size: 'iconLg', icon: 'trash', title: t('remove'), onclick: function () { rows.splice(idx, 1); render(); emit(); } }))),
        err));
    });
  }
  render();
  var el = h('div', { 'data-rows': f.name }, listEl, h('div', { class: 'mt-2' }, add));
  function show(i, msg) {
    var e = errors[i];
    if (!e) return;
    e.el.textContent = msg || '';
    e.el.hidden = !msg;
    e.el.parentElement.classList.toggle('eco-row-invalid', !!msg);
  }
  el.validate = function () {
    var bad = false;
    errors.forEach(function (e, i) {
      var msg = '';
      f.columns.forEach(function (c) {
        if (msg) return;
        var v = e.row[c.name];
        if (c.type === 'i18n') {
          var hit = trCheck(v, { max: c.max });
          if (hit) msg = t(c.label) + ': ' + t('code.' + hit.code, { lang: langName(hit.lang) });
          return;
        }
        if (v === null || v === undefined || v === '') { if (c.required) msg = t(c.label) + ': ' + t('err.required'); return; }
        if (c.type === 'number' && (isNaN(v) || (c.min !== undefined && v < c.min) || (c.max !== undefined && v > c.max))) msg = t(c.label) + ': ' + t('err.number');
      });
      show(i, msg);
      if (msg) bad = true;
    });
    if (!bad && f.check) {
      var hit2 = f.check(rows);
      if (hit2) { show(hit2.index, hit2.message); bad = true; }
    }
    if (!bad && f.min && rows.length < f.min) return t(f.emptyText || 'rows.empty');
    return bad ? t('err.form') : null;
  };
  el.showServerError = function (err) {
    var m = /\[(\d+)\]/.exec(String(err && err.field || ''));
    if (!m) return false;
    show(Number(m[1]), errorText(err));
    var e = errors[Number(m[1])];
    if (e) e.el.scrollIntoView({ block: 'center', behavior: 'smooth' });
    return !!e;
  };
  return el;
}

function duplicateCheck(name, code) {
  return function (rows) {
    var seen = {};
    for (var i = 0; i < rows.length; i++) {
      var v = rows[i][name];
      if (v === null || v === undefined) continue;
      if (seen[v]) return { index: i, message: t('code.' + code) };
      seen[v] = 1;
    }
    return null;
  };
}

function grantOptions() {
  return opts('grant.', ['claim', 'auto'])();
}

function activeFilter() {
  return { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } };
}

function playerTagsField() {
  return { name: 'audienceTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.ev.audience', hint: 'f.ev.audience.hint', optional: true, empty: [] };
}

function windowRow() {
  return { type: 'row', fields: [
    { name: 'startsAt', type: 'datetime', label: 'f.ev.startsAt', optional: true },
    { name: 'endsAt', type: 'datetime', label: 'f.ev.endsAt', optional: true, check: function (v, vals) { return v && vals.startsAt && new Date(v) <= new Date(vals.startsAt) ? t('code.INVALID_WINDOW') : null; } }
  ] };
}

function afterEventSave(resource) {
  return function () { forgetRefs(resource); };
}

RESOURCES['event-types'] = {
  filters: [activeFilter(), { name: 'clientAllowed', label: 'filter.ev.source', options: function () { return [{ value: 'true', label: t('ev.source.client') }, { value: 'false', label: t('ev.source.server') }]; } }],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '55%' },
    { label: 'col.ev.source', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, tag(t(it.clientAllowed ? 'ev.source.client' : 'ev.source.server'), it.clientAllowed ? 'warning' : 'secondary'), it.isSystem ? tag(t('ev.system'), 'primary') : null);
    }, width: 170 },
    { label: 'col.ev.range', render: function (it) {
      if (it.valueMin == null && it.valueMax == null) return h('span', { class: 'text-secondary-400', text: t('ev.range.any') });
      return h('span', { class: 'tabular-nums whitespace-nowrap', text: (it.valueMin == null ? '…' : fmtNum(it.valueMin)) + ' – ' + (it.valueMax == null ? '…' : fmtNum(it.valueMax)) });
    }, width: 120 },
    { label: 'col.ev.rate', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: t('ev.perMinute', { n: it.rateLimit }) }); }, width: 130 },
    { label: 'col.used', render: function (it) { return it.usedBy ? h('span', { class: 'tabular-nums', text: t('ev.usedBy', { n: it.usedBy }) }) : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 120 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { name: 'isSystem', type: 'static', readonly: true, when: function (v) { return !!v.isSystem; }, render: function () { return notice('primary', t('ev.system.notice')); } },
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.ev.key.hint', required: true, keyRe: EVENT_KEY_RE, keyErr: 'code.INVALID_EVENT_KEY', lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { name: 'clientAllowed', type: 'bool', inline: true, label: 'f.ev.client', hint: 'f.ev.client.hint', empty: false, when: function (v) { return !v.isSystem; } },
    { type: 'row', fields: [
      { name: 'valueMin', type: 'number', int: false, label: 'f.ev.valueMin', optional: true },
      { name: 'valueMax', type: 'number', int: false, label: 'f.ev.valueMax', optional: true, check: function (v, vals) { return v != null && vals.valueMin != null && Number(v) < Number(vals.valueMin) ? t('code.INVALID_RANGE') : null; } }
    ] },
    { name: 'rateLimit', type: 'number', label: 'f.ev.rate', hint: 'f.ev.rate.hint', required: true, min: 1, max: 100000 },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.ev.active.hint' }
  ],
  defaults: function () { return { key: '', name: {}, description: {}, clientAllowed: false, valueMin: null, valueMax: null, rateLimit: 30, isActive: true, sort: 0 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 0; return p; },
  afterSave: afterEventSave('event-types')
};

RESOURCES['event-rules'] = {
  filters: [
    { name: 'eventTypeId', label: 'filter.ev.event', type: 'ref', resource: 'event-types' },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '40%' },
    { label: 'col.ev.when', render: ruleSentenceCell, skeleton: '60%' },
    { label: 'col.ev.prize', render: function (it) { return h('div', { class: 'flex flex-col gap-1 min-w-0' }, refNameCell('prizes', it.prizeId), h('div', {}, grantTag(it.grantMode))); }, width: 170 },
    { label: 'col.ev.claims', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.claims || 0) }); }, width: 100 },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive), (it.audienceTagIds || []).length ? tag(t('audience.only'), 'primary') : null, it.endsAt && new Date(it.endsAt) < new Date() ? tag(t('ev.ended'), 'secondary') : null); }, width: 130 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { label: 'f.ev.rule.name', hint: 'f.ev.rule.name.hint', required: true, max: 128 }),
    tr('description', { label: 'f.ev.rule.description', multiline: true, max: 1000, optional: true }),
    { type: 'heading', label: 'ev.h.when' },
    { name: 'eventTypeId', type: 'ref', resource: 'event-types', label: 'f.ev.event', required: true },
    { type: 'row', fields: [
      { name: 'measure', type: 'select', label: 'f.ev.measure', hint: 'f.ev.measure.hint', required: true, options: opts('measure.', ['count', 'sum', 'max']) },
      { name: 'period', type: 'select', label: 'f.ev.period', required: true, options: opts('period.', ['total', 'day', 'week']) }
    ] },
    { type: 'row', fields: [
      { name: 'threshold', type: 'number', int: false, label: 'f.ev.threshold', required: true, min: 0.0001 },
      { name: 'repeat', type: 'select', label: 'f.ev.repeat', required: true, options: opts('repeat.', ['once', 'period', 'every']),
        check: function (v, vals) { return v === 'period' && vals.period === 'total' ? t('code.REPEAT_NEEDS_PERIOD') : v === 'every' && vals.measure === 'max' ? t('code.EVERY_NEEDS_TOTAL') : null; } }
    ] },
    { name: 'maxClaims', type: 'number', label: 'f.ev.maxClaims', hint: 'f.ev.maxClaims.hint', optional: true, min: 1, when: function (v) { return v.repeat !== 'once'; } },
    { name: 'preview', type: 'static', readonly: true, live: true, render: function (v) {
      if (!v.eventTypeId || !v.measure || !v.period || !v.repeat || v.threshold == null || v.threshold === '') return h('span', {});
      var item = { measure: v.measure, period: v.period, repeat: v.repeat, threshold: parseNumber(v.threshold, false), maxClaims: v.repeat === 'once' ? null : v.maxClaims };
      return h('div', { class: 'eco-rule-preview text-sm rounded-md px-3 py-2', 'data-rule-preview': '' }, ruleSentenceCell(Object.assign({ eventTypeId: v.eventTypeId }, item)));
    } },
    { type: 'heading', label: 'ev.h.prize' },
    { type: 'row', fields: [
      { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', required: true },
      { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', hint: 'f.ev.grantMode.hint', required: true, options: grantOptions }
    ] },
    { type: 'heading', label: 'ev.h.who' },
    playerTagsField(),
    windowRow(),
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
  ],
  defaults: function (query) { return { key: '', name: {}, description: {}, eventTypeId: (query && query.eventTypeId) || null, measure: 'count', period: 'day', threshold: 1, repeat: 'period', maxClaims: null, grantMode: 'claim', audienceTagIds: [], startsAt: null, endsAt: null, isActive: true, sort: 0 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 0; if (p.repeat === 'once') p.maxClaims = null; return p; }
};

RESOURCES['meta-rules'] = {
  filters: [
    { name: 'eventTypeId', label: 'filter.ev.event', type: 'ref', resource: 'event-types' },
    { name: 'metaId', label: 'filter.ev.meta', type: 'ref', resource: 'metas' },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '40%' },
    { label: 'col.ev.event', render: function (it) { return refNameCell('event-types', it.eventTypeId); }, width: 170 },
    { label: 'col.ev.change', render: function (it) {
      return h('div', { class: 'flex items-center gap-1.5 text-sm min-w-0', 'data-meta-change': '' }, refNameCell('metas', it.metaId),
        h('span', { class: 'tabular-nums whitespace-nowrap text-secondary-600 dark:text-secondary-300', text: t('mr.change.' + it.action + '.' + it.valueSource, { n: numLocal(it.amount) }) }));
    }, skeleton: '45%' },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive), (it.audienceTagIds || []).length ? tag(t('audience.only'), 'primary') : null); }, width: 130 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.mr.sort.hint' }
    ] },
    tr('name', { label: 'f.mr.name', hint: 'f.mr.name.hint', required: true, max: 128 }),
    { type: 'row', fields: [
      { name: 'eventTypeId', type: 'ref', resource: 'event-types', label: 'f.ev.event', required: true },
      { name: 'metaId', type: 'ref', resource: 'metas', label: 'f.mr.meta', required: true }
    ] },
    { type: 'row', fields: [
      { name: 'action', type: 'select', label: 'f.mr.action', required: true, options: opts('mr.action.', ['add', 'set', 'max']) },
      { name: 'valueSource', type: 'select', label: 'f.mr.source', hint: 'f.mr.source.hint', required: true, options: opts('mr.source.', ['fixed', 'event']) }
    ] },
    { name: 'amount', type: 'number', int: false, label: 'f.mr.amount', hint: 'f.mr.amount.hint', required: true },
    playerTagsField(),
    windowRow(),
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
  ],
  defaults: function () { return { key: '', name: {}, eventTypeId: null, metaId: null, action: 'add', valueSource: 'event', amount: 1, audienceTagIds: [], startsAt: null, endsAt: null, isActive: true, sort: 0 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 0; return p; }
};

RESOURCES['level-tables'] = {
  filters: [{ name: 'metaId', label: 'filter.ev.meta', type: 'ref', resource: 'metas' }, activeFilter()],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '40%' },
    { label: 'col.lv.meta', render: function (it) {
      return h('div', { class: 'flex items-center gap-1.5 text-sm min-w-0' }, refNameCell('metas', it.metaId), it.levelMetaId ? h('span', { class: 'text-secondary-400', text: '←' }) : null, it.levelMetaId ? refNameCell('metas', it.levelMetaId) : null);
    }, skeleton: '35%' },
    { label: 'col.lv.steps', render: function (it) {
      var steps = it.steps || [];
      var top = steps[steps.length - 1];
      return h('div', { class: 'flex flex-col text-sm' }, h('span', { class: 'tabular-nums', text: t('lv.levelsN', { n: steps.length }) }),
        top ? h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: t('lv.top', { level: top.level, n: numLocal(top.threshold) }) }) : null);
    }, width: 160 },
    { label: 'col.ev.players', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.players || 0) }); }, width: 100 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { type: 'row', fields: [
      { name: 'metaId', type: 'ref', resource: 'metas', label: 'f.lv.meta', hint: 'f.lv.meta.hint', required: true },
      { name: 'levelMetaId', type: 'ref', resource: 'metas', label: 'f.lv.levelMeta', hint: 'f.lv.levelMeta.hint', optional: true,
        check: function (v, vals) { return v && v === vals.metaId ? t('code.SAME_META') : null; } }
    ] },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' },
    { type: 'heading', label: 'lv.h.steps' },
    { name: 'steps', type: 'rows', addLabel: 'lv.add', emptyText: 'lv.empty', min: 1, max: 200,
      newRow: function (rows) {
        var last = rows[rows.length - 1];
        return { level: last ? Number(last.level || 0) + 1 : 1, threshold: last ? Number(last.threshold || 0) * 2 || 100 : 0, name: {}, prizeId: null, grantMode: 'claim' };
      },
      check: function (rows) {
        var dup = duplicateCheck('level', 'DUPLICATE_LEVEL')(rows);
        if (dup) return dup;
        var sorted = rows.map(function (r, i) { return { r: r, i: i }; }).sort(function (a, b) { return a.r.level - b.r.level; });
        for (var i = 1; i < sorted.length; i++) if (Number(sorted[i].r.threshold) < Number(sorted[i - 1].r.threshold)) return { index: sorted[i].i, message: t('code.THRESHOLDS_NOT_ORDERED') };
        return null;
      },
      columns: [
        { name: 'level', type: 'number', label: 'f.lv.level', required: true, min: 1 },
        { name: 'threshold', type: 'number', int: false, label: 'f.lv.threshold', required: true, min: 0 },
        { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', optional: true },
        { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', options: grantOptions },
        { name: 'name', type: 'i18n', label: 'f.lv.title', max: 64, wide: true }
      ] }
  ],
  defaults: function () { return { key: '', name: {}, description: {}, metaId: null, levelMetaId: null, isActive: true, sort: 0, steps: [{ level: 1, threshold: 0, name: {}, prizeId: null, grantMode: 'claim' }, { level: 2, threshold: 100, name: {}, prizeId: null, grantMode: 'claim' }] }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 0; return p; }
};

RESOURCES.streaks = {
  filters: [{ name: 'eventTypeId', label: 'filter.ev.event', type: 'ref', resource: 'event-types' }, activeFilter()],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '40%' },
    { label: 'col.ev.event', render: function (it) { return refNameCell('event-types', it.eventTypeId); }, width: 170 },
    { label: 'col.sk.rule', render: function (it) {
      return h('div', { class: 'flex flex-col text-sm gap-0.5' },
        h('span', { text: it.periodDays === 1 ? t('sk.everyDay') : t('sk.everyDays', { n: it.periodDays }) }),
        h('span', { class: 'text-xs text-secondary-500', text: (it.gracePeriods ? t('sk.graceN', { n: it.gracePeriods }) : t('sk.noGrace')) + ' · ' + t('sk.miss.' + it.missAction) }));
    }, skeleton: '40%' },
    { label: 'col.sk.rewards', render: function (it) {
      var n = (it.rewards || []).length;
      return h('div', { class: 'flex flex-col text-sm' }, h('span', { class: 'tabular-nums', text: t('sk.rewardsN', { n: n }) }),
        it.cycleLength ? h('span', { class: 'text-xs text-secondary-500', text: t('sk.cycleN', { n: it.cycleLength }) }) : null);
    }, width: 150 },
    { label: 'col.ev.players', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.players || 0) }); }, width: 100 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { name: 'hiro', type: 'static', readonly: true, render: function () { return notice('secondary', t('sk.hiro.notice')); } },
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { name: 'eventTypeId', type: 'ref', resource: 'event-types', label: 'f.sk.event', hint: 'f.sk.event.hint', required: true },
    { type: 'row', fields: [
      { name: 'periodDays', type: 'number', label: 'f.sk.periodDays', hint: 'f.sk.periodDays.hint', required: true, min: 1, max: 30 },
      { name: 'cycleLength', type: 'number', label: 'f.sk.cycle', hint: 'f.sk.cycle.hint', optional: true, min: 1, max: 365 }
    ] },
    { type: 'row', fields: [
      { name: 'gracePeriods', type: 'number', label: 'f.sk.grace', hint: 'f.sk.grace.hint', required: true, min: 0, max: 30 },
      { name: 'missAction', type: 'select', label: 'f.sk.miss', required: true, options: opts('sk.miss.', ['reset', 'step_back']) }
    ] },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' },
    { type: 'heading', label: 'sk.h.rewards' },
    { name: 'rewards', type: 'rows', addLabel: 'sk.add', emptyText: 'sk.empty', max: 365,
      newRow: function (rows) {
        var last = rows[rows.length - 1];
        return { day: last ? Number(last.day || 0) + 1 : 1, prizeId: last ? last.prizeId : null, grantMode: last ? last.grantMode : 'claim' };
      },
      check: function (rows) { return duplicateCheck('day', 'DUPLICATE_DAY')(rows); },
      columns: [
        { name: 'day', type: 'number', label: 'f.sk.day', required: true, min: 1 },
        { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', required: true },
        { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', options: grantOptions }
      ] }
  ],
  defaults: function () { return { key: '', name: {}, description: {}, eventTypeId: null, periodDays: 1, cycleLength: 7, gracePeriods: 0, missAction: 'reset', isActive: true, sort: 0, rewards: [] }; },
  fixPayload: function (p) {
    if (p.sort == null) p.sort = 0;
    return p;
  }
};

LISTS['event-log'] = {
  noRow: true,
  filters: [
    { name: 'typeId', label: 'filter.ev.event', type: 'ref', resource: 'event-types' },
    { name: 'source', label: 'filter.ev.from', options: opts('ev.from.', ['client', 'server', 'admin']) },
    { name: 'userId', label: 'filter.user', type: 'text' }
  ],
  columns: [
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 170 },
    { label: 'col.player', render: function (it) {
      return h('div', { class: 'min-w-0' }, h('a', { href: '#/players/' + encodeURIComponent(it.userId), class: 'eco-link font-medium truncate block', text: it.displayName || it.username || String(it.userId).slice(0, 8) }),
        it.username && it.displayName ? h('div', { class: 'text-xs text-secondary-500 truncate', dir: 'ltr', text: it.username }) : null);
    }, skeleton: '40%' },
    { label: 'col.ev.event', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'truncate', text: i18nText(it.name) || it.key }), h('div', { class: 'text-secondary-500' }, mono(it.key))); }, skeleton: '40%' },
    { label: 'col.ev.value', render: function (it) { return numCell(it.value); }, width: 100 },
    { label: 'col.ev.from', render: function (it) { return tag(t('ev.from.' + it.source), it.source === 'client' ? 'warning' : it.source === 'admin' ? 'primary' : 'secondary'); }, width: 110 },
    { label: 'col.ev.metadata', render: function (it) {
      var keys = it.metadata ? Object.keys(it.metadata) : [];
      if (!keys.length) return h('span', { class: 'text-secondary-400', text: t('none') });
      return h('span', { class: 'font-mono text-xs text-secondary-600 dark:text-secondary-300 eco-clamp', dir: 'ltr', text: JSON.stringify(it.metadata) });
    }, skeleton: '30%' }
  ]
};

OVERVIEW_CATALOG.push('event-types', 'event-rules', 'level-tables', 'streaks');

var TIMEZONES = ['Asia/Tehran', 'UTC', 'Asia/Dubai', 'Asia/Baghdad', 'Asia/Kabul', 'Asia/Baku', 'Asia/Yerevan', 'Europe/Istanbul', 'Europe/London', 'Europe/Berlin', 'Europe/Moscow', 'America/New_York', 'America/Los_Angeles', 'America/Toronto', 'Australia/Sydney', 'Asia/Tokyo', 'Asia/Kolkata', 'Asia/Riyadh'];

function weekdayName(i) {
  var d = new Date(Date.UTC(2026, 10, 8 + i));
  return new Intl.DateTimeFormat(locale(), { weekday: 'long', timeZone: 'UTC' }).format(d);
}

function zoneNow(zone) {
  try {
    return new Intl.DateTimeFormat(locale(), { weekday: 'long', hour: '2-digit', minute: '2-digit', timeZone: zone }).format(new Date());
  } catch (e) {
    return '';
  }
}

var EVENT_SETTINGS_FIELDS = [
  { type: 'row', fields: [
    { name: 'timezone', type: 'select', label: 'f.es.timezone', hint: 'f.es.timezone.hint', required: true, options: function () {
      return TIMEZONES.map(function (z) { return { value: z, label: z === 'UTC' ? 'UTC' : z.replace('_', ' ') }; });
    } },
    { name: 'weekStart', type: 'select', label: 'f.es.weekStart', hint: 'f.es.weekStart.hint', required: true, noPlaceholder: true, options: function () {
      return [6, 0, 1, 2, 3, 4, 5].map(function (i) { return { value: String(i), label: weekdayName(i) }; });
    } }
  ] },
  { name: 'clock', type: 'static', readonly: true, live: true, render: function (v) {
    var now = zoneNow(v.timezone || 'UTC');
    return now ? h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 -mt-2 mb-4', 'data-zone-now': '' }, t('es.now', { time: now })) : h('span', {});
  } },
  { type: 'row', fields: [
    { name: 'retentionDays', type: 'number', label: 'f.es.retention', hint: 'f.es.retention.hint', required: true, min: 1, max: 3650 },
    { name: 'playerRateLimit', type: 'number', label: 'f.es.rate', hint: 'f.es.rate.hint', required: true, min: 1, max: 100000 }
  ] }
];

function eventSettingsPage(container) {
  var frame = tabFrame(t('desc.events'));
  clear(container).appendChild(frame.el);
  var page = apiPage(api, '/event-settings', null, EVENT_SETTINGS_FIELDS, { id: 'events', fixPayload: function (p) {
    p.weekStart = Number(p.weekStart);
    return p;
  } });
  frame.content.appendChild(page.el);
  page.load();
}

var PROGRESS_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadProgressCard(id) {
  PROGRESS_CARD.loading = true;
  return api('GET', '/users/' + encodeURIComponent(id) + '/progress' + qs({ lang: lang() })).then(function (res) {
    PROGRESS_CARD.loading = false;
    PROGRESS_CARD.at = Date.now();
    PROGRESS_CARD.data = res;
    PROGRESS_CARD.err = null;
    var old = document.querySelector('[data-player-progress]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    PROGRESS_CARD.loading = false;
    PROGRESS_CARD.at = Date.now();
    PROGRESS_CARD.err = err;
    schedule();
  });
}

function claimStatusTag(c) {
  var color = c.status === 'granted' ? 'success' : c.status === 'failed' ? 'error' : 'warning';
  return tag(t('claim.status.' + (c.status === 'ready' ? (c.grantMode === 'claim' ? 'waiting' : 'ready') : c.status)), color);
}

function recordEventDialog(userId) {
  var values = { key: '', value: 1 };
  var picker = refPicker({ resource: 'event-types', valueKey: 'key', value: '', required: true, label: t('f.ev.event'), filter: { isActive: 'true' }, onchange: function (v) { values.key = v || ''; } });
  var val = numberInput(1, { int: false, label: t('f.ev.value'), oninput: function (v) { values.value = v === '' ? null : parseNumber(v, false); } });
  var keyWrap = field(t('f.ev.event'), picker, null);
  var valWrap = field(t('f.ev.value'), val, t('f.ev.value.hint'));
  return formDialog({
    id: 'record-event',
    title: t('pp.record'),
    text: t('pp.record.text'),
    icon: 'event-types',
    ok: t('pp.record.ok'),
    body: h('div', {}, keyWrap, valWrap),
    validate: function () {
      keyWrap.setError(values.key ? '' : t('err.required'));
      valWrap.setError(values.value === null || isNaN(values.value) ? t('err.number') : '');
      return !!values.key && values.value !== null && !isNaN(values.value);
    },
    submit: function () { return api('POST', '/users/' + encodeURIComponent(userId) + '/events', { key: values.key, value: values.value }); }
  });
}

function progressSection(titleKey, body, extra) {
  return h('div', { class: 'min-w-0' }, h('div', { class: 'flex items-center justify-between gap-2 text-xs text-secondary-500 dark:text-secondary-400 mb-1' }, h('span', { text: t(titleKey) }), extra || null), body);
}

function playerProgressCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-progress]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (PROGRESS_CARD.id !== id) { PROGRESS_CARD.id = id; PROGRESS_CARD.data = null; PROGRESS_CARD.err = null; PROGRESS_CARD.at = 0; }
  if (!PROGRESS_CARD.loading && Date.now() - PROGRESS_CARD.at > 20000) loadProgressCard(id);
  var p = PROGRESS_CARD.data;
  var anchor = grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify([p.totals, p.streaks, p.levels, p.claims, p.readyClaims]) : PROGRESS_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-progress': '', 'data-key': key, 'data-economy': 'player-progress', 'data-l10n-skip': '', 'aria-labelledby': 'player-progress-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-progress-title', class: 'font-medium', text: t('pp.title') });
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(PROGRESS_CARD.err ? notice('error', errorText(PROGRESS_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { PROGRESS_CARD.at = 0; loadProgressCard(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var write = aclCan('HIRO_ECONOMY', 1);
  var reload = function (res) { if (res) { PROGRESS_CARD.at = 0; loadProgressCard(id); } };
  var actions = h('div', { class: 'flex flex-wrap items-center gap-2' });
  if (write) {
    var rec = button({ variant: 'ghost', size: 'sm', icon: 'event-types', label: t('pp.record'), onclick: function () { recordEventDialog(id).then(function (res) { if (res) toast('success', t('pp.recorded')); reload(res); }); } });
    rec.setAttribute('data-record-event', '');
    actions.appendChild(rec);
  }
  actions.appendChild(h('a', { href: hashFor('event-log', null, { userId: id }), class: 'eco-link text-sm', text: t('pp.log') }));
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-3' }, title, actions));

  var levels = h('div', { class: 'flex flex-col gap-2' });
  (p.levels || []).forEach(function (lv) {
    var next = lv.next;
    levels.appendChild(h('div', { class: 'min-w-0', 'data-pp-level': lv.key },
      h('div', { class: 'font-medium truncate', text: lv.name }),
      h('div', { class: 'text-sm', text: t('pp.level', { level: lv.level }) + (lv.title ? ' · ' + lv.title : '') }),
      h('div', { class: 'text-xs text-secondary-500 tabular-nums', text: t('pp.meta', { n: numLocal(lv.value), name: lv.metaName }) + (next ? ' · ' + t('pp.toNext', { level: next.level, n: numLocal(Math.max(0, next.threshold - lv.value)) }) : '') })));
  });
  if (!(p.levels || []).length) levels.appendChild(h('span', { class: 'text-sm text-secondary-400', text: t('none') }));

  var streaks = h('div', { class: 'flex flex-col gap-2' });
  (p.streaks || []).forEach(function (sk) {
    var row = h('div', { class: 'min-w-0', 'data-pp-streak': sk.key },
      h('div', { class: 'flex items-center gap-2 min-w-0' }, h('span', { class: 'font-medium truncate', text: sk.name }), sk.broken ? tag(t('pp.broken'), 'secondary') : sk.countedThisPeriod ? tag(t('pp.today'), 'success') : null),
      h('div', { class: 'text-sm tabular-nums', text: t('pp.streak', { n: sk.count, best: sk.best }) }));
    if (write && (sk.count > 0 || sk.countedThisPeriod)) {
      var reset = button({ variant: 'link', size: 'sm', label: t('pp.reset'), onclick: function () {
        confirmModal({ title: t('pp.reset.title'), text: t('pp.reset.text', { name: sk.name }), ok: t('pp.reset'), danger: true }).then(function (yes) {
          if (!yes) return;
          api('POST', '/users/' + encodeURIComponent(id) + '/streaks/' + encodeURIComponent(sk.id) + '/reset', {}).then(function () { toast('success', t('pp.reset.done')); reload(true); }, function (err) { toast('error', errorText(err)); });
        });
      } });
      reset.setAttribute('data-streak-reset', sk.key);
      row.appendChild(reset);
    }
    streaks.appendChild(row);
  });
  if (!(p.streaks || []).length) streaks.appendChild(h('span', { class: 'text-sm text-secondary-400', text: t('none') }));

  var totals = h('div', { class: 'flex flex-col gap-1' });
  (p.totals || []).slice(0, 6).forEach(function (e) {
    totals.appendChild(h('div', { class: 'flex items-center justify-between gap-2 text-sm', 'data-pp-total': e.key },
      h('span', { class: 'truncate', text: i18nText(e.name) || e.key }),
      h('span', { class: 'tabular-nums text-secondary-600 dark:text-secondary-300 whitespace-nowrap', text: t('pp.times', { n: e.count }) })));
  });
  if (!(p.totals || []).length) totals.appendChild(h('span', { class: 'text-sm text-secondary-400', text: t('pp.noEvents') }));

  body.appendChild(h('div', { class: 'eco-pc-grid' }, progressSection('pp.levels', levels), progressSection('pp.streaks', streaks), progressSection('pp.events', totals)));

  var claims = h('div', { class: 'flex flex-col mt-3 border-t border-secondary-200 dark:border-secondary-700 pt-4', 'data-pp-claims': '' });
  claims.appendChild(h('div', { class: 'flex items-center justify-between gap-2 text-xs text-secondary-500 dark:text-secondary-400 mb-1' }, h('span', { text: t('pp.claims') }), p.readyClaims ? tag(t('pp.readyN', { n: p.readyClaims }), 'warning') : null));
  (p.claims || []).slice(0, 8).forEach(function (c) {
    var line = h('div', { class: 'flex flex-wrap items-center gap-2 py-1 text-sm', 'data-pp-claim': c.claimId },
      tag(t('claim.kind.' + c.kind), 'secondary'),
      h('span', { class: 'truncate', text: c.sourceName || c.source }),
      c.level ? h('span', { class: 'text-xs text-secondary-500', text: t('pp.level', { level: c.level }) }) : null,
      h('span', { class: 'text-secondary-400', text: '←' }),
      h('span', { class: 'font-medium truncate', text: c.prize && c.prize.name }),
      claimStatusTag(c),
      h('span', { class: 'flex-1' }),
      h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: fmtDate(c.grantedAt || c.createdAt) }));
    if (write && c.status === 'failed') {
      line.appendChild(button({ variant: 'link', size: 'sm', icon: 'refresh', label: t('pay.retry'), onclick: function () {
        api('POST', '/event-claims/' + encodeURIComponent(c.claimId) + '/retry', {}).then(function () { toast('success', t('pay.retry.done')); reload(true); }, function (err) { toast('error', errorText(err)); });
      } }));
    }
    claims.appendChild(line);
  });
  if (!(p.claims || []).length) claims.appendChild(h('span', { class: 'text-sm text-secondary-400', text: t('pp.noClaims') }));
  body.appendChild(claims);
  place();
}

ICONS.tasks = '<path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>';
ICONS['task-groups'] = '<rect x="3" y="4" width="18" height="5" rx="1"/><rect x="3" y="11" width="18" height="5" rx="1"/><path d="M7 20h10"/>';

function minutesText(n) {
  n = Number(n) || 0;
  if (n && n % 1440 === 0) return t('tk.days', { n: fmtNum(n / 1440) });
  if (n && n % 60 === 0) return t('tk.hours', { n: fmtNum(n / 60) });
  return t('tk.minutes', { n: fmtNum(n) });
}

function taskSentence(it, eventName) {
  var parts = [];
  if (it.condition === 'children') parts.push(t('tk.when.children'));
  else {
    parts.push(t('tk.when.' + (it.metric || 'count'), { event: eventName, n: numLocal(it.target) }));
    if (it.valueMin != null && it.valueMax != null) parts.push(t('tk.filter.range', { min: numLocal(it.valueMin), max: numLocal(it.valueMax) }));
    else if (it.valueMin != null) parts.push(t('tk.filter.min', { n: numLocal(it.valueMin) }));
    else if (it.valueMax != null) parts.push(t('tk.filter.max', { n: numLocal(it.valueMax) }));
  }
  if (!it.parentId) {
    parts.push(it.repeat === 'interval' ? t('tk.repeat.interval', { time: minutesText(it.repeatMinutes) }) : t('tk.repeat.' + (it.repeat || 'once')));
    if (it.repeat !== 'once' && it.maxCompletions) parts.push(t('tk.max', { n: fmtNum(it.maxCompletions) }));
  }
  return parts.join(' · ');
}

function taskSentenceCell(it) {
  var box = h('div', { class: 'text-sm text-secondary-700 dark:text-secondary-200 eco-clamp', 'data-task-sentence': '' });
  if (it.condition === 'children' || !it.eventTypeId) { box.appendChild(document.createTextNode(taskSentence(it, ''))); return box; }
  var ref = refName('event-types', it.eventTypeId);
  function paint(et) { clear(box).appendChild(document.createTextNode(taskSentence(it, et ? i18nText(et.name) || et.key : '…'))); }
  if (ref.item !== undefined) paint(ref.item);
  else { box.appendChild(bar('70%')); ref.p.then(paint); }
  return box;
}

function handoutSentence(it) {
  var parts = [t('tg.active', { n: fmtNum(it.activeCount || 1) }), t('tg.order.' + it.orderMode)];
  parts.push(it.handout === 'timer' ? t('tg.handout.timer.after', { time: minutesText(it.handoutMinutes) }) : t('tg.handout.' + it.handout));
  if (it.isRepeatable) parts.push(t('tg.repeatable'));
  return parts.join(' · ');
}

function levelRow() {
  return { type: 'row', fields: [
    { name: 'levelTableId', type: 'ref', resource: 'level-tables', label: 'f.tk.levelTable', hint: 'f.tk.levelTable.hint', optional: true,
      check: function (v, vals) { return !v && (vals.minLevel != null || vals.maxLevel != null) ? t('code.LEVEL_TABLE_REQUIRED') : null; } },
    { name: 'minLevel', type: 'number', label: 'f.tk.minLevel', optional: true, min: 0 },
    { name: 'maxLevel', type: 'number', label: 'f.tk.maxLevel', optional: true, min: 0,
      check: function (v, vals) { return v != null && vals.minLevel != null && Number(v) < Number(vals.minLevel) ? t('code.INVALID_LEVEL_RANGE') : null; } }
  ] };
}

function isEvent(v) { return v.condition !== 'children'; }
function isTop(v) { return !v.parentId; }

RESOURCES.tasks = {
  filters: [
    { name: 'level', label: 'filter.tk.level', options: function () { return [{ value: 'top', label: t('tk.level.top') }, { value: 'sub', label: t('tk.level.sub') }]; } },
    { name: 'groupId', label: 'filter.tk.group', type: 'ref', resource: 'task-groups' },
    { name: 'eventTypeId', label: 'filter.ev.event', type: 'ref', resource: 'event-types' },
    { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'task' } },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '40%' },
    { label: 'col.tk.when', render: taskSentenceCell, skeleton: '55%' },
    { label: 'col.ev.prize', render: function (it) {
      if (!it.prizeId) return h('span', { class: 'text-secondary-400', text: t('tk.noPrize') });
      return h('div', { class: 'flex flex-col gap-1 min-w-0' }, refNameCell('prizes', it.prizeId), h('div', {}, grantTag(it.grantMode)));
    }, width: 170 },
    { label: 'col.tk.place', render: function (it) {
      var box = h('div', { class: 'flex flex-col gap-1 min-w-0 text-sm' });
      if (it.parentId) box.appendChild(h('div', { class: 'flex items-center gap-1 min-w-0' }, tag(t('tk.sub'), 'secondary'), refNameCell('tasks', it.parentId)));
      else if (it.groupId) box.appendChild(h('div', { class: 'flex items-center gap-1 min-w-0' }, tag(t('tk.inGroup'), 'primary'), refNameCell('task-groups', it.groupId)));
      if (it.children) box.appendChild(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('tk.childrenN', { n: fmtNum(it.children) }) }));
      if ((it.dependsOn || []).length) box.appendChild(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('tk.depsN', { n: fmtNum(it.dependsOn.length) }) }));
      if (!box.childNodes.length) box.appendChild(h('span', { class: 'text-secondary-400', text: t('tk.standalone') }));
      return box;
    }, width: 170 },
    { label: 'col.tk.done', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: t('tk.doneN', { n: fmtNum(it.completed || 0), p: fmtNum(it.players || 0) }) }); }, width: 120 },
    { label: 'col.status', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive),
        (it.audienceTagIds || []).length || it.minLevel != null || it.maxLevel != null ? tag(t('audience.only'), 'primary') : null,
        it.endsAt && new Date(it.endsAt) < new Date() ? tag(t('ev.ended'), 'secondary') : null);
    }, width: 130 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.tk.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { label: 'f.tk.name', hint: 'f.tk.name.hint', required: true, max: 128 }),
    tr('description', { label: 'f.tk.description', multiline: true, max: 1000, optional: true }),
    { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
    { type: 'row', fields: [categoryField('task'), tagsField('task')] },
    { type: 'heading', label: 'tk.h.when' },
    { name: 'condition', type: 'select', label: 'f.tk.condition', hint: 'f.tk.condition.hint', required: true, noPlaceholder: true, options: opts('tk.condition.', ['event', 'children']) },
    { name: 'eventTypeId', type: 'ref', resource: 'event-types', label: 'f.ev.event', required: true, when: isEvent },
    { type: 'row', when: isEvent, fields: [
      { name: 'metric', type: 'select', label: 'f.tk.metric', hint: 'f.tk.metric.hint', required: true, noPlaceholder: true, options: opts('tk.metric.', ['count', 'sum_value', 'high_value']), when: isEvent },
      { name: 'target', type: 'number', int: false, label: 'f.tk.target', required: true, min: 0.0001, when: isEvent }
    ] },
    { type: 'row', when: isEvent, fields: [
      { name: 'valueMin', type: 'number', int: false, label: 'f.tk.valueMin', hint: 'f.tk.valueFilter.hint', optional: true, when: isEvent },
      { name: 'valueMax', type: 'number', int: false, label: 'f.tk.valueMax', optional: true, when: isEvent,
        check: function (v, vals) { return v != null && vals.valueMin != null && Number(v) < Number(vals.valueMin) ? t('code.INVALID_RANGE') : null; } }
    ] },
    { name: 'parentId', type: 'ref', resource: 'tasks', filter: { level: 'top' }, label: 'f.tk.parent', hint: 'f.tk.parent.hint', optional: true, when: isEvent },
    { name: 'preview', type: 'static', readonly: true, live: true, render: function (v) {
      if (isEvent(v) && (!v.eventTypeId || v.target == null || v.target === '')) return h('span', {});
      var item = { condition: v.condition, eventTypeId: v.eventTypeId, metric: v.metric, target: parseNumber(v.target, false), valueMin: v.valueMin, valueMax: v.valueMax, repeat: v.repeat, repeatMinutes: v.repeatMinutes, maxCompletions: v.maxCompletions, parentId: v.parentId };
      return h('div', { class: 'eco-rule-preview text-sm rounded-md px-3 py-2', 'data-task-preview': '' }, taskSentenceCell(item));
    } },
    { type: 'heading', label: 'tk.h.repeat', when: isTop },
    { type: 'row', when: isTop, fields: [
      { name: 'repeat', type: 'select', label: 'f.tk.repeat', required: true, noPlaceholder: true, options: opts('tk.repeatopt.', ['once', 'daily', 'weekly', 'interval']), when: isTop,
        check: function (v, vals) { return vals.groupId && v !== 'once' ? t('code.GROUP_TASK_REPEAT') : null; } },
      { name: 'repeatMinutes', type: 'number', label: 'f.tk.repeatMinutes', hint: 'f.tk.repeatMinutes.hint', required: true, min: 1, max: 525600, when: function (v) { return isTop(v) && v.repeat === 'interval'; } },
      { name: 'maxCompletions', type: 'number', label: 'f.tk.maxCompletions', hint: 'f.tk.maxCompletions.hint', optional: true, min: 1, when: function (v) { return isTop(v) && v.repeat !== 'once'; } }
    ] },
    { name: 'groupId', type: 'ref', resource: 'task-groups', label: 'f.tk.group', hint: 'f.tk.group.hint', optional: true, when: isTop },
    { name: 'dependsOn', type: 'refMulti', resource: 'tasks', label: 'f.tk.dependsOn', hint: 'f.tk.dependsOn.hint', optional: true, empty: [] },
    { type: 'heading', label: 'ev.h.prize' },
    { type: 'row', fields: [
      { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', hint: 'f.tk.prize.hint', optional: true },
      { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', hint: 'f.ev.grantMode.hint', required: true, noPlaceholder: true, options: grantOptions }
    ] },
    { name: 'notify', type: 'bool', inline: true, label: 'f.tk.notify', hint: 'f.tk.notify.hint', empty: false },
    { type: 'heading', label: 'ev.h.who', when: isTop },
    { name: 'audienceTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.ev.audience', hint: 'f.ev.audience.hint', optional: true, empty: [], when: isTop },
    Object.assign(levelRow(), { when: isTop }),
    Object.assign(windowRow(), { when: isTop }),
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
  ],
  defaults: function (query) {
    return { key: '', name: {}, description: {}, iconUrl: '', categoryId: null, tagIds: [], condition: 'event', eventTypeId: (query && query.eventTypeId) || null, metric: 'count', target: 1, valueMin: null, valueMax: null,
      repeat: 'once', repeatMinutes: null, maxCompletions: null, parentId: (query && query.parentId) || null, dependsOn: [], groupId: (query && query.groupId) || null, prizeId: null, grantMode: 'claim', notify: false,
      audienceTagIds: [], levelTableId: null, minLevel: null, maxLevel: null, startsAt: null, endsAt: null, isActive: true, sort: 0 };
  },
  fixPayload: function (p) {
    if (p.sort == null) p.sort = 0;
    if (p.condition === 'children') { p.eventTypeId = null; p.parentId = null; p.valueMin = null; p.valueMax = null; p.target = 1; p.metric = 'count'; }
    if (p.parentId) { p.groupId = null; p.repeat = 'once'; p.audienceTagIds = []; p.levelTableId = null; p.minLevel = null; p.maxLevel = null; p.startsAt = null; p.endsAt = null; }
    if (p.repeat !== 'interval') p.repeatMinutes = null;
    if (p.repeat === 'once') p.maxCompletions = null;
    if (p.minLevel == null && p.maxLevel == null) p.levelTableId = null;
    p.dependsOn = p.dependsOn || [];
    return p;
  },
  afterSave: afterEventSave('tasks')
};

RESOURCES['task-groups'] = {
  filters: [
    { name: 'handout', label: 'filter.tg.handout', options: opts('tg.handoutopt.', ['auto', 'request', 'timer']) },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '40%' },
    { label: 'col.tg.rule', render: function (it) { return h('div', { class: 'text-sm text-secondary-700 dark:text-secondary-200 eco-clamp', 'data-group-sentence': '', text: handoutSentence(it) }); }, skeleton: '55%' },
    { label: 'col.tg.tasks', render: function (it) { return h('span', { class: 'tabular-nums', text: t('tg.tasksN', { n: fmtNum((it.taskIds || []).length) }) }); }, width: 120 },
    { label: 'col.ev.players', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.players || 0) }); }, width: 100 },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive), (it.audienceTagIds || []).length || it.minLevel != null || it.maxLevel != null ? tag(t('audience.only'), 'primary') : null); }, width: 130 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.tk.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { label: 'f.tg.name', required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
    categoryField('task'),
    { type: 'heading', label: 'tg.h.handout' },
    { type: 'row', fields: [
      { name: 'activeCount', type: 'number', label: 'f.tg.activeCount', hint: 'f.tg.activeCount.hint', required: true, min: 1, max: 20 },
      { name: 'orderMode', type: 'select', label: 'f.tg.orderMode', required: true, noPlaceholder: true, options: opts('tg.order.', ['ordered', 'random']) }
    ] },
    { type: 'row', fields: [
      { name: 'handout', type: 'select', label: 'f.tg.handout', hint: 'f.tg.handout.hint', required: true, noPlaceholder: true, options: opts('tg.handoutopt.', ['auto', 'request', 'timer']) },
      { name: 'handoutMinutes', type: 'number', label: 'f.tg.handoutMinutes', required: true, min: 1, max: 525600, when: function (v) { return v.handout === 'timer'; } }
    ] },
    { name: 'isRepeatable', type: 'bool', inline: true, label: 'f.tg.repeatable', hint: 'f.tg.repeatable.hint', empty: false },
    { type: 'heading', label: 'tg.h.tasks' },
    { name: 'taskIds', type: 'refMulti', resource: 'tasks', filter: { level: 'top' }, label: 'f.tg.tasks', hint: 'f.tg.tasks.hint', optional: true, empty: [] },
    { type: 'heading', label: 'ev.h.who' },
    playerTagsField(),
    levelRow(),
    windowRow(),
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
  ],
  defaults: function () {
    return { key: '', name: {}, description: {}, iconUrl: '', categoryId: null, activeCount: 1, orderMode: 'ordered', handout: 'auto', handoutMinutes: 60, isRepeatable: false, taskIds: [],
      audienceTagIds: [], levelTableId: null, minLevel: null, maxLevel: null, startsAt: null, endsAt: null, isActive: true, sort: 0 };
  },
  fixPayload: function (p) {
    if (p.sort == null) p.sort = 0;
    if (p.handout !== 'timer') p.handoutMinutes = 0;
    if (p.minLevel == null && p.maxLevel == null) p.levelTableId = null;
    p.taskIds = p.taskIds || [];
    return p;
  },
  afterSave: function () { forgetRefs('task-groups'); forgetRefs('tasks'); }
};

OVERVIEW_CATALOG.push('tasks', 'task-groups');

var TASK_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadTaskCard(id) {
  TASK_CARD.loading = true;
  return api('GET', '/users/' + encodeURIComponent(id) + '/tasks' + qs({ lang: lang() })).then(function (res) {
    TASK_CARD.loading = false;
    TASK_CARD.at = Date.now();
    TASK_CARD.data = res;
    TASK_CARD.err = null;
    var old = document.querySelector('[data-player-tasks]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    TASK_CARD.loading = false;
    TASK_CARD.at = Date.now();
    TASK_CARD.err = err;
    schedule();
  });
}

function taskStateTag(it) {
  var s = it.state;
  var color = s === 'claimed' ? 'success' : s === 'completed' ? 'warning' : s === 'active' ? 'primary' : 'secondary';
  var label = t('tk.state.' + (s === 'claimed' && !it.prize ? 'finished' : s));
  if (s === 'hidden' && it.reason) label = t('tk.reason.' + it.reason);
  return tag(label, color);
}

function taskAction(userId, it, what) {
  var done = { reset: 'tk.reset.done', complete: 'tk.complete.done', claim: 'tk.claim.done' }[what];
  function run() {
    api('POST', '/users/' + encodeURIComponent(userId) + '/tasks/' + encodeURIComponent(it.id) + '/' + what, {}).then(function () {
      toast('success', t(done));
      TASK_CARD.at = 0;
      loadTaskCard(userId);
    }, function (err) { toast('error', errorText(err)); });
  }
  if (what === 'claim') { run(); return; }
  confirmModal({ title: t('tk.' + what + '.title'), text: t('tk.' + what + '.text', { name: it.name }), ok: t('tk.' + what), danger: what === 'reset' }).then(function (yes) { if (yes) run(); });
}

function taskMeter(it) {
  var total = Number(it.target) || 0;
  var pct = it.state === 'completed' || it.state === 'claimed' ? 100 : total > 0 ? Math.max(0, Math.min(100, Math.round((Number(it.value) || 0) * 100 / total))) : 0;
  return h('span', { class: 'eco-tk-meter', 'aria-hidden': 'true', 'data-done': pct >= 100 ? '' : null }, h('span', { style: 'inline-size:' + pct + '%' }));
}

function taskLine(userId, it, write, depth) {
  var progress = it.condition === 'children' ? t('tk.progress.children', { n: fmtNum(it.value), total: fmtNum(it.target) }) : t('tk.progress', { n: numLocal(it.value), total: numLocal(it.target) });
  var line = h('div', { class: 'flex flex-wrap items-center gap-2 py-1.5 text-sm' + (depth ? ' eco-tk-sub' : ''), 'data-pp-task': it.key, 'data-state': it.state },
    h('span', { class: 'font-medium truncate min-w-0', text: it.name || it.key }),
    taskStateTag(it),
    taskMeter(it),
    h('span', { class: 'tabular-nums text-secondary-600 dark:text-secondary-300 whitespace-nowrap', 'data-task-progress': '', text: progress }),
    it.repeat && it.repeat !== 'once' && !depth ? h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: t('tk.repeat.' + it.repeat) + (it.completions ? (lang() === 'fa' ? '، ' : ' · ') + t('tk.completionsN', { n: fmtNum(it.completions) }) : '') }) : null,
    h('span', { class: 'flex-1' }));
  if (write && it.state !== 'hidden' && it.state !== 'locked') {
    var acts = h('div', { class: 'flex flex-wrap items-center gap-1' });
    if (it.state === 'active' && it.status) acts.appendChild(Object.assign(button({ variant: 'link', size: 'sm', label: t('tk.reset'), onclick: function () { taskAction(userId, it, 'reset'); } }), { title: t('tk.reset') }));
    if (it.state === 'completed' && it.claimable) acts.appendChild(button({ variant: 'link', size: 'sm', label: t('tk.claim'), onclick: function () { taskAction(userId, it, 'claim'); } }));
    if (it.state === 'completed' && !it.claimable) acts.appendChild(button({ variant: 'link', size: 'sm', label: t('tk.reset'), onclick: function () { taskAction(userId, it, 'reset'); } }));
    if (it.state === 'active') acts.appendChild(button({ variant: 'link', size: 'sm', label: t('tk.complete'), onclick: function () { taskAction(userId, it, 'complete'); } }));
    Array.prototype.forEach.call(acts.querySelectorAll('button'), function (b) { b.setAttribute('data-task-action', b.textContent.trim()); });
    line.appendChild(acts);
  }
  var wrap = h('div', { class: 'min-w-0' }, line);
  (it.children || []).forEach(function (c) { wrap.appendChild(taskLine(userId, c, write, 1)); });
  return wrap;
}

function playerTasksCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-tasks]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (TASK_CARD.id !== id) { TASK_CARD.id = id; TASK_CARD.data = null; TASK_CARD.err = null; TASK_CARD.at = 0; }
  if (!TASK_CARD.loading && Date.now() - TASK_CARD.at > 20000) loadTaskCard(id);
  var p = TASK_CARD.data;
  var anchor = grid.querySelector('[data-player-activity]') || grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify([p.items, p.groups]) : TASK_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-tasks': '', 'data-key': key, 'data-economy': 'player-tasks', 'data-l10n-skip': '', 'aria-labelledby': 'player-tasks-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-tasks-title', class: 'font-medium', text: t('tk.card') });
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(TASK_CARD.err ? notice('error', errorText(TASK_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { TASK_CARD.at = 0; loadTaskCard(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var write = aclCan('HIRO_ECONOMY', 1);
  var visible = (p.items || []).filter(function (it) { return it.isActive && it.state !== 'hidden'; });
  var hidden = (p.items || []).length - visible.length;
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' }, title,
    h('a', { href: hashFor('tasks'), class: 'eco-link text-sm', text: t('tk.all') })));
  var list = h('div', { class: 'flex flex-col eco-tk-list', 'data-pp-tasks': '' });
  visible.forEach(function (it) { list.appendChild(taskLine(id, it, write, 0)); });
  if (!visible.length) list.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('tk.none') }));
  body.appendChild(list);
  if (hidden) body.appendChild(h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: t('tk.hiddenN', { n: fmtNum(hidden) }) }));
  (p.groups || []).filter(function (g) { return g.isActive; }).forEach(function (g) {
    var head = h('div', { class: 'flex flex-wrap items-center gap-2 text-sm' },
      h('span', { class: 'font-medium', text: g.name || g.key }),
      tag(g.open ? t('tg.round', { n: fmtNum(g.cycle) }) : t('tk.reason.' + (g.reason || 'groupClosed')), g.open ? 'primary' : 'secondary'),
      h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums', text: t('tg.progress', { n: fmtNum(g.done), total: fmtNum(g.total) }) }),
      g.nextAt ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('tg.nextAt', { date: fmtDate(g.nextAt, true) }) }) : null,
      h('span', { class: 'flex-1' }));
    if (write && g.open && g.canRequest) {
      var nx = button({ variant: 'link', size: 'sm', label: t('tg.next'), onclick: function () {
        api('POST', '/users/' + encodeURIComponent(id) + '/task-groups/' + encodeURIComponent(g.id) + '/next', {}).then(function () { toast('success', t('tg.next.done')); TASK_CARD.at = 0; loadTaskCard(id); }, function (err) { toast('error', errorText(err)); });
      } });
      nx.setAttribute('data-group-next', g.key);
      head.appendChild(nx);
    }
    var box = h('div', { class: 'mt-3 border-t border-secondary-200 dark:border-secondary-700 pt-4', 'data-pp-group': g.key }, head);
    (g.tasks || []).forEach(function (it) { box.appendChild(taskLine(id, it, write, 0)); });
    if (g.open && !(g.tasks || []).length) box.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('tg.empty') }));
    body.appendChild(box);
  });
  place();
}

var TICKET_STATUSES = ['open', 'awaiting_staff', 'awaiting_player', 'resolved', 'closed'];
var TICKET_STATUS_COLOR = { open: 'warning', awaiting_staff: 'warning', awaiting_player: 'primary', resolved: 'success', closed: 'secondary' };
var TICKET_PRIORITIES = ['low', 'normal', 'high', 'urgent'];
var PRIORITY_COLOR = { low: 'secondary', normal: 'secondary', high: 'warning', urgent: 'error' };
var SUPPORT = { summary: null, at: 0, loading: false, agents: null, agentsAt: 0, beat: null, away: false };
try { SUPPORT.away = localStorage.getItem('economy.support.away') === '1'; } catch (e) { SUPPORT.away = false; }

function supportBadge(page) {
  var s = SUPPORT.summary;
  if (!s) return 0;
  if (page === 'tickets') return s.awaitingStaff || 0;
  if (page === 'live-chat') return s.waitingChats || 0;
  return 0;
}

function supportBadgeKey() {
  return supportBadge('tickets') + ':' + supportBadge('live-chat');
}

function setSummary(res) {
  if (!res || typeof res.awaitingStaff !== 'number') return;
  var before = supportBadgeKey();
  SUPPORT.summary = res;
  if (supportBadgeKey() !== before) schedule();
}

function refreshSupportBadge(force) {
  var now = Date.now();
  if (SUPPORT.loading || (!force && now - SUPPORT.at < 30000)) return;
  if (!token() || !aclCan('ACCOUNT_NOTES', 0)) return;
  SUPPORT.at = now;
  SUPPORT.loading = true;
  quietBase('GET', '/support/summary').then(function (res) { SUPPORT.loading = false; setSummary(res); }, function () { SUPPORT.loading = false; });
}

function onSupportPage() {
  return /^#\/support\//.test(location.hash || '');
}

function sendBeat() {
  if (!token() || !aclCan('ACCOUNT_NOTES', 1)) return Promise.resolve(null);
  return quietBase('POST', '/support/heartbeat', { away: SUPPORT.away }).then(function (res) { setSummary(res); return res; });
}

function supportBeat() {
  if (SUPPORT.beat) return;
  sendBeat();
  SUPPORT.beat = setInterval(function () {
    if (!onSupportPage() || !token()) { clearInterval(SUPPORT.beat); SUPPORT.beat = null; return; }
    sendBeat();
  }, 20000);
}

function loadAgents(force) {
  if (!force && SUPPORT.agents && Date.now() - SUPPORT.agentsAt < 15000) return Promise.resolve(SUPPORT.agents);
  return baseApi('GET', '/support/agents').then(function (res) {
    SUPPORT.agents = (res && res.items) || [];
    SUPPORT.me = res && res.me;
    SUPPORT.agentsAt = Date.now();
    return SUPPORT.agents;
  }, function () { return SUPPORT.agents || []; });
}

function ticketStatusTag(status) {
  var el = tag(t('tk.status.' + status), TICKET_STATUS_COLOR[status]);
  el.setAttribute('data-ticket-status', status);
  return el;
}

function priorityTag(p) {
  var el = tag(t('tk.priority.' + p), PRIORITY_COLOR[p]);
  el.setAttribute('data-priority', p);
  return el;
}

function deptName(d) {
  return d ? (i18nText(d.name) || d.key) : t('none');
}

function agentLabel(name) {
  return name ? name : t('tk.unassigned');
}

function minutesText(n) {
  if (n == null) return '—';
  if (n < 60) return t('dur.minutes', { n: localDigits(n) });
  if (n < 1440) return t('dur.hours', { n: localDigits(Math.round(n / 60)) });
  return t('dur.days', { n: localDigits(Math.round(n / 1440)) });
}

function ticketSubjectCell(it) {
  return h('div', { class: 'min-w-0' },
    h('div', { class: 'flex items-center gap-2 min-w-0' },
      h('span', { class: 'font-medium truncate', dir: 'auto', text: it.subject }),
      it.overdue ? tag(t('tk.overdue'), 'error') : null),
    h('div', { class: 'text-xs text-secondary-500 flex items-center gap-2 min-w-0' }, mono(it.number), h('span', { class: 'truncate', dir: 'auto', text: it.lastPreview || '' })));
}

LISTS.tickets = {
  api: 'base',
  path: 'support/tickets',
  top: function () { return ticketsTop(); },
  create: function () { return ticketsCreate(); },
  filters: [
    { name: 'status', label: 'filter.status', options: function () { return [{ value: 'waiting', label: t('tk.filter.waiting') }, { value: 'active', label: t('tk.filter.active') }].concat(TICKET_STATUSES.map(function (s) { return { value: s, label: t('tk.status.' + s) }; })); } },
    { name: 'priority', label: 'filter.priority', options: function () { return TICKET_PRIORITIES.map(function (p) { return { value: p, label: t('tk.priority.' + p) }; }); } },
    { name: 'assignee', label: 'filter.assignee', options: function () { return [{ value: 'me', label: t('tk.mine') }, { value: 'none', label: t('tk.unassigned') }].concat((SUPPORT.agents || []).map(function (a) { return { value: a.username, label: a.username }; })); } },
    { name: 'departmentId', label: 'filter.department', type: 'ref', resource: 'support/departments', api: baseApi },
    { name: 'overdue', label: 'filter.overdue', options: function () { return [{ value: 'true', label: t('tk.overdueOnly') }]; } }
  ],
  columns: [
    { label: 'col.subject', render: ticketSubjectCell, skeleton: '60%' },
    { label: 'col.player', render: function (it) { return playerLink(it.userId, it.displayName || it.username); }, width: 150 },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-col items-start gap-1' }, ticketStatusTag(it.status), it.priority !== 'normal' ? priorityTag(it.priority) : null); }, width: 150 },
    { label: 'col.assignee', render: function (it) { return h('span', { class: it.assignee ? '' : 'text-secondary-400', text: agentLabel(it.assignee) }); }, width: 130 },
    { label: 'col.department', render: function (it) { return h('span', { class: 'truncate', text: deptName(it.department) }); }, width: 150 },
    { label: 'col.lastMessage', render: function (it) { return h('div', { class: 'flex flex-col' }, dateCell(it.lastMessageAt), h('span', { class: 'text-xs text-secondary-500', text: t('tk.by.' + it.lastMessageBy) })); }, width: 170 }
  ]
};

function supportStats(s) {
  var grid = h('div', { class: 'eco-stat-grid mb-4', 'data-support-stats': '' });
  grid.appendChild(statCard(t('tk.stat.waiting'), s.awaitingStaff, null, hashFor('tickets', null, { status: 'waiting' })));
  grid.appendChild(statCard(t('tk.stat.overdue'), s.overdue, null, hashFor('tickets', null, { status: 'waiting', overdue: 'true' }), { alert: s.overdue > 0 }));
  grid.appendChild(statCard(t('tk.stat.unassigned'), s.unassigned, null, hashFor('tickets', null, { status: 'active', assignee: 'none' })));
  grid.appendChild(statCard(t('tk.stat.mine'), s.mine, null, hashFor('tickets', null, { status: 'waiting', assignee: 'me' })));
  grid.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-1 h-full' },
    h('div', { class: 'font-medium text-secondary-700 dark:text-secondary-200', text: t('tk.stat.firstResponse') }),
    h('div', { class: 'text-2xl font-semibold tabular-nums', text: minutesText(s.avgFirstResponseMinutes) }),
    h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('tk.stat.30days') }))));
  grid.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-1 h-full' },
    h('div', { class: 'font-medium text-secondary-700 dark:text-secondary-200', text: t('tk.stat.rating') }),
    h('div', { class: 'text-2xl font-semibold tabular-nums', text: s.avgRating ? localDigits(s.avgRating) + ' / ' + localDigits(5) : '—' }),
    h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('tk.stat.rated', { n: s.rated || 0 }) }))));
  return grid;
}

function ticketsPage(container, query, ctx) {
  loadAgents();
  return listPage(container, 'tickets', query, ctx);
}

function ticketsTop() {
  var holder = h('div', { 'data-support-top': '' }, h('div', { class: 'eco-stat-grid mb-4' }, [1, 2, 3, 4].map(function () { return card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem'), bar('3rem', 'h-7'))); })));
  baseApi('GET', '/support/summary').then(function (s) { setSummary(s); clear(holder).appendChild(supportStats(s)); }, function () { clear(holder); });
  return holder;
}

function ticketsCreate() {
  if (!aclCan('ACCOUNT_NOTES', 1)) return null;
  var btn = button({ variant: 'primary', size: 'page', icon: 'plus', label: t('tk.new'), onclick: function () { newTicketDialog().then(function (res) { if (res && res.id) go('tickets', res.id); }); } });
  btn.setAttribute('data-ticket-new', '');
  return btn;
}

function playerPicker(onchange) {
  return refPicker({ resource: 'support/players', api: baseApi, required: true, label: t('col.player'), placeholder: t('tk.pickPlayer'), onchange: onchange });
}

function departmentPicker(value, onchange) {
  return refPicker({ resource: 'support/departments', api: baseApi, value: value || '', label: t('col.department'), filter: { isActive: 'true' }, onchange: onchange });
}

function newTicketDialog() {
  var v = { userId: '', subject: '', departmentId: '', priority: 'normal', body: '' };
  var fields = {};
  fields.player = field(t('col.player'), playerPicker(function (x) { v.userId = x || ''; fields.player.setError(''); }));
  fields.subject = field(t('tk.subject'), textInput('', { label: t('tk.subject'), oninput: function (x) { v.subject = x; fields.subject.setError(''); } }));
  fields.dept = field(t('col.department'), departmentPicker('', function (x) { v.departmentId = x || ''; }), null, { optional: true });
  fields.priority = field(t('tk.priority'), segmented(TICKET_PRIORITIES.map(function (p) { return { value: p, label: t('tk.priority.' + p) }; }), 'normal', function (x) { v.priority = x; }, t('tk.priority')));
  fields.body = field(t('tk.message'), textInput('', { multiline: true, rows: 4, label: t('tk.message'), oninput: function (x) { v.body = x; fields.body.setError(''); } }), t('tk.new.hint'));
  return formDialog({
    id: 'ticket-new', title: t('tk.new'), text: t('tk.new.text'), icon: 'tickets', ok: t('tk.new.ok'),
    body: h('div', {}, fields.player, fields.subject, fields.dept, fields.priority, fields.body),
    validate: function () {
      var ok = true;
      if (!v.userId) { fields.player.setError(t('err.required')); ok = false; }
      if (!v.subject.trim()) { fields.subject.setError(t('err.required')); ok = false; }
      if (!v.body.trim()) { fields.body.setError(t('err.required')); ok = false; }
      return ok;
    },
    submit: function () { return baseApi('POST', '/support/tickets', v); }
  });
}

function fileSize(n) {
  if (n < 1024) return t('size.b', { n: localDigits(n) });
  if (n < 1048576) return t('size.kb', { n: localDigits(Math.round(n / 1024)) });
  return t('size.mb', { n: localDigits((n / 1048576).toFixed(1)) });
}

function attachmentList(list) {
  if (!list || !list.length) return null;
  return h('div', { class: 'flex flex-wrap gap-2 mt-2', 'data-attachments': '' }, list.map(function (a) {
    var isImg = /^image\//.test(a.type || '');
    return h('a', { href: a.url || '#', target: '_blank', rel: 'noopener noreferrer', class: 'eco-att', 'data-attachment': a.key, title: a.name },
      isImg && a.url ? h('img', { src: a.url, alt: a.name, class: 'eco-att-img', loading: 'lazy' }) : icon('clip', 'h-4 w-4'),
      h('span', { class: 'eco-att-name truncate', dir: 'auto', text: a.name }),
      h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: fileSize(a.size || 0) }));
  }));
}

function eventLine(m) {
  var d = m.eventData || {};
  var vars = {
    from: d.from ? (has('tk.status.' + d.from) ? t('tk.status.' + d.from) : has('tk.priority.' + d.from) ? t('tk.priority.' + d.from) : d.from) : '',
    to: d.to ? (has('tk.status.' + d.to) ? t('tk.status.' + d.to) : has('tk.priority.' + d.to) ? t('tk.priority.' + d.to) : d.to) : '',
    who: m.authorType === 'system' ? t('tk.system') : m.authorType === 'player' ? t('tk.player') : (m.author || ''),
    n: d.rating || 0
  };
  if (m.event === 'department_changed') vars.to = '';
  var key = 'tk.ev.' + m.event + (m.event === 'assigned' && !d.to ? '.none' : '') + (d.auto ? '.auto' : '');
  if (!has(key)) key = 'tk.ev.' + m.event;
  return h('li', { class: 'eco-msg-event', 'data-event': m.event }, h('span', { text: t(key, vars) }), h('span', { class: 'text-secondary-400 whitespace-nowrap', text: fmtDate(m.createdAt, true) }));
}

function messageBubble(m, opts) {
  opts = opts || {};
  if (m.kind === 'event') return eventLine(m);
  var staff = m.authorType === 'staff' || m.sender === 'agent';
  var note = m.kind === 'note';
  var cls = 'eco-msg' + (staff ? ' eco-msg--staff' : ' eco-msg--player') + (note ? ' eco-msg--note' : '') + (m.sender === 'system' ? ' eco-msg--system' : '');
  var who = note ? t('tk.note.by', { name: m.author || '' }) : staff ? (m.author || t('tk.staff')) : (opts.playerName || t('tk.player'));
  return h('li', { class: cls, 'data-msg-kind': m.kind || m.sender || '', 'data-seq': m.seq != null ? String(m.seq) : null },
    h('div', { class: 'eco-msg-head' }, h('span', { class: 'font-medium', text: who }), note ? tag(t('tk.note'), 'warning') : null, h('span', { class: 'text-secondary-500 whitespace-nowrap', text: fmtDate(m.createdAt, true) })),
    m.body ? h('div', { class: 'eco-msg-body', dir: 'auto', text: m.body }) : null,
    attachmentList(m.attachments));
}

function cannedPicker(getLang, departmentId, onpick) {
  var sel = h('select', { class: SELECT_SM + ' eco-canned', 'aria-label': t('tk.canned') }, h('option', { value: '', text: t('tk.canned.pick') }));
  var items = [];
  baseApi('GET', '/support/canned' + qs({ isActive: 'true', limit: 100, departmentId: departmentId || '' })).then(function (res) {
    items = (res && res.items) || [];
    items.forEach(function (c, i) { sel.appendChild(h('option', { value: String(i), text: i18nText(c.title) || c.key })); });
    if (!items.length) sel.appendChild(h('option', { value: '', disabled: true, text: t('tk.canned.none') }));
  }, function () { return null; });
  sel.addEventListener('change', function () {
    var c = items[Number(sel.value)];
    sel.value = '';
    if (!c) return;
    var l = getLang();
    var body = (c.body && (c.body[l] || c.body[LANGS.def])) || i18nText(c.body) || '';
    onpick(body);
  });
  return sel;
}

function insertAtCursor(el, text) {
  var start = el.selectionStart == null ? el.value.length : el.selectionStart;
  var end = el.selectionEnd == null ? el.value.length : el.selectionEnd;
  var before = el.value.slice(0, start);
  var glue = before && !/\s$/.test(before) ? '\n' : '';
  el.value = before + glue + text + el.value.slice(end);
  var pos = (before + glue + text).length;
  el.focus();
  try { el.setSelectionRange(pos, pos); } catch (e) { pos = 0; }
  el.dispatchEvent(new Event('input', { bubbles: true }));
}

function uploadButton(onfile) {
  var input = h('input', { type: 'file', accept: 'image/png,image/jpeg,image/webp,application/pdf', class: 'sr-only', 'data-attach-input': '' });
  var btn = button({ variant: 'ghost', size: 'sm', icon: 'clip', label: t('tk.attach'), onclick: function () { input.click(); } });
  input.addEventListener('change', function () {
    var f = input.files && input.files[0];
    input.value = '';
    if (!f) return;
    var fd = new FormData();
    fd.append('file', f);
    setBusy(btn, true, t('tk.uploading'));
    baseApi('POST', '/support/uploads', fd, true).then(function (a) { setBusy(btn, false); onfile(a); }, function (err) { setBusy(btn, false); toast('error', errorText(err)); });
  });
  return h('span', {}, input, btn);
}

function composer(opts) {
  var state = { body: '', internal: false, status: 'awaiting_player', files: [] };
  var area = textInput('', { multiline: true, rows: 4, label: t('tk.reply'), placeholder: t(opts.placeholder || 'tk.reply.ph'), oninput: function (v) { state.body = v; if (opts.oninput) opts.oninput(v); } });
  area.setAttribute('data-composer', '');
  var files = h('div', { class: 'flex flex-wrap gap-2' });
  function paintFiles() {
    clear(files);
    state.files.forEach(function (a, i) {
      files.appendChild(h('span', { class: 'eco-att', 'data-pending-attachment': a.key }, icon('clip', 'h-4 w-4'), h('span', { class: 'eco-att-name truncate', text: a.name }),
        button({ variant: 'muted', size: 'icon', icon: 'close', iconClass: 'h-3 w-3', title: t('remove'), onclick: function () { state.files.splice(i, 1); paintFiles(); } })));
    });
  }
  var tools = h('div', { class: 'flex flex-wrap items-center gap-2' }, cannedPicker(opts.lang, opts.departmentId, function (txt) { insertAtCursor(area, txt); }));
  if (opts.attachments) tools.appendChild(uploadButton(function (a) { state.files.push(a); paintFiles(); }));
  var after = null;
  if (opts.statusAfter) {
    after = segmented([{ value: 'awaiting_player', label: t('tk.after.awaiting_player') }, { value: 'resolved', label: t('tk.after.resolved') }, { value: 'closed', label: t('tk.after.closed') }], 'awaiting_player', function (v) { state.status = v; }, t('tk.after'));
    after.setAttribute('data-reply-status', '');
  }
  var noteToggle = opts.notes ? toggle(false, t('tk.internal'), function (v) { state.internal = v; box.classList.toggle('eco-composer--note', v); if (after) after.hidden = v; send.querySelector('span').textContent = v ? t('tk.saveNote') : t(opts.sendLabel || 'tk.send'); }, 'sm') : null;
  if (noteToggle) noteToggle.setAttribute('data-internal-toggle', '');
  var send = button({ variant: 'primary', size: 'sm', icon: 'send', label: t(opts.sendLabel || 'tk.send'), onclick: function () { submit(); } });
  send.setAttribute('data-send', '');
  var err = h('div', {});
  function submit() {
    clear(err);
    if (!state.body.trim() && !state.files.length) { err.appendChild(h('p', { class: 'text-sm text-error-600 dark:text-error-400 mb-2', text: t('err.required') })); area.focus(); return; }
    setBusy(send, true, t('saving'));
    opts.onsend({ body: state.body, internal: state.internal, status: state.status, attachments: state.files.map(function (f) { return f.key; }) }).then(function () {
      setBusy(send, false);
      state.body = '';
      area.value = '';
      state.files = [];
      paintFiles();
    }, function (e) {
      setBusy(send, false);
      err.appendChild(h('div', { class: 'mb-2' }, notice('error', errorText(e))));
    });
  }
  area.addEventListener('keydown', function (e) { if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) { e.preventDefault(); submit(); } });
  var box = h('div', { class: 'eco-composer', 'data-composer-box': '' }, err, area, files,
    h('div', { class: 'flex flex-wrap items-center gap-3 justify-between' }, tools, h('div', { class: 'flex flex-wrap items-center gap-3' }, noteToggle, after, send)));
  box.focus = function () { area.focus(); };
  return box;
}

function infoRow(label, value) {
  return h('div', { class: 'eco-info-row' }, h('dt', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: label }), h('dd', { class: 'text-sm' }, value));
}

function ticketDetail(container, id) {
  loadAgents();
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('tickets', null, lastQuery.tickets || {}); } });
  var frame = pageFrame(t('tk.detail'), '', backBtn);
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-ticket', 'data-ticket': id });
  frame.content.appendChild(grid);
  var write = aclCan('ACCOUNT_NOTES', 1);

  function skeleton() {
    clear(grid).appendChild(h('div', { class: 'eco-ticket-main' }, card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('50%', 'h-5'), bar('90%', 'h-16'), bar('80%', 'h-16'), bar('100%', 'h-24')))));
    grid.appendChild(h('div', { class: 'eco-ticket-side' }, card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('40%'), bar('70%'), bar('60%'), bar('50%')))));
  }

  function act(path, body) {
    return baseApi('POST', '/support/tickets/' + encodeURIComponent(id) + path, body).then(function (res) { render(res); refreshSupportBadge(true); return res; });
  }

  function render(tk) {
    var h1 = frame.el.querySelector('h1 span');
    if (h1) h1.textContent = tk.subject;
    clear(grid);
    var thread = h('ol', { class: 'eco-thread', 'data-thread': '' });
    (tk.messages || []).forEach(function (m) { thread.appendChild(messageBubble(m, { playerName: tk.displayName || tk.username })); });
    var main = h('div', { class: 'eco-ticket-main flex flex-col gap-4' },
      card(h('div', { class: 'p-4' },
        h('div', { class: 'flex flex-wrap items-center gap-2 mb-3' }, mono(tk.number), ticketStatusTag(tk.status), priorityTag(tk.priority), tk.overdue ? tag(t('tk.overdue'), 'error') : null, tk.source !== 'player' ? tag(t('tk.source.' + tk.source), 'secondary') : null),
        thread)));
    if (write && tk.status !== 'closed') {
      main.appendChild(card(h('div', { class: 'p-4' }, composer({ notes: true, statusAfter: true, attachments: true, departmentId: tk.departmentId, lang: function () { return tk.lang || LANGS.def; },
        onsend: function (v) { return act('/messages', v).then(function () { toast('success', v.internal ? t('tk.noteSaved') : t('tk.sent')); }); } }))));
    } else if (write) {
      main.appendChild(notice('secondary', t('tk.closedNote')));
      var noteOnly = composer({ notes: false, statusAfter: false, attachments: false, departmentId: tk.departmentId, lang: function () { return tk.lang || LANGS.def; }, sendLabel: 'tk.saveNote', placeholder: 'tk.note.ph',
        onsend: function (v) { return act('/messages', { body: v.body, internal: true }); } });
      main.appendChild(card(h('div', { class: 'p-4' }, noteOnly)));
    }
    grid.appendChild(main);

    var side = h('dl', { class: 'flex flex-col gap-3' });
    side.appendChild(infoRow(t('col.player'), h('div', { class: 'flex flex-col' }, playerLink(tk.userId, tk.displayName || tk.username), h('span', { class: 'text-xs text-secondary-500', text: t('tk.playerCounts', { tickets: (tk.player || {}).tickets || 0, chats: (tk.player || {}).chats || 0 }) }),
      h('a', { href: hashFor('tickets', null, { userId: tk.userId }), class: 'eco-link text-xs', text: t('tk.playerTickets') }),
      aclCan('ACCOUNT', 0) ? h('a', { href: hashFor('activity', null, { userId: tk.userId }), class: 'eco-link text-xs', text: t('act.playerLink') }) : null)));
    if (write) {
      var statusBtns = h('div', { class: 'flex flex-wrap gap-2', 'data-status-actions': '' });
      var moves = tk.status === 'closed' ? [] : tk.status === 'resolved' ? ['closed', 'awaiting_staff'] : ['resolved', 'closed'];
      if (tk.status === 'closed') moves = ['awaiting_staff'];
      moves.forEach(function (s) {
        var b = button({ variant: s === 'resolved' ? 'primary' : 'ghost', size: 'sm', label: t('tk.move.' + s), onclick: function () { setBusy(b, true); act('/status', { status: s }).then(function () { toast('success', t('tk.moved')); }, function (e) { setBusy(b, false); toast('error', errorText(e)); }); } });
        b.setAttribute('data-move', s);
        statusBtns.appendChild(b);
      });
      side.appendChild(infoRow(t('col.status'), h('div', { class: 'flex flex-col gap-2 items-start' }, ticketStatusTag(tk.status), statusBtns)));
      var agentSel = selectInput(tk.assignee || '', [{ value: '', label: t('tk.unassigned') }].concat((SUPPORT.agents || []).map(function (a) { return { value: a.username, label: a.username + (a.online ? ' • ' + t('lc.online') : '') }; })).concat(tk.assignee && !(SUPPORT.agents || []).some(function (a) { return a.username === tk.assignee; }) ? [{ value: tk.assignee, label: tk.assignee }] : []), function (v) {
        act('/assign', { assignee: v }).then(function () { toast('success', t('tk.assigned')); }, function (e) { toast('error', errorText(e)); });
      }, { label: t('col.assignee'), placeholder: false });
      agentSel.setAttribute('data-assignee', '');
      var mine = !tk.assignee || tk.assignee !== SUPPORT.me ? button({ variant: 'link', size: 'sm', label: t('tk.takeIt'), onclick: function () { act('/assign', { assignee: 'me' }).then(function () { toast('success', t('tk.assigned')); }); } }) : null;
      side.appendChild(infoRow(t('col.assignee'), h('div', { class: 'flex flex-col gap-1 items-start' }, agentSel, mine)));
      side.appendChild(infoRow(t('col.department'), departmentPicker(tk.departmentId, function (v) { if (v) act('/assign', { departmentId: v }).then(function () { toast('success', t('saved')); }, function (e) { toast('error', errorText(e)); }); })));
      var prio = segmented(TICKET_PRIORITIES.map(function (p) { return { value: p, label: t('tk.priority.' + p) }; }), tk.priority, function (v) { act('/assign', { priority: v }).then(function () { toast('success', t('saved')); }, function (e) { toast('error', errorText(e)); }); }, t('tk.priority'));
      prio.setAttribute('data-priority-pick', '');
      side.appendChild(infoRow(t('tk.priority'), prio));
    } else {
      side.appendChild(infoRow(t('col.status'), ticketStatusTag(tk.status)));
      side.appendChild(infoRow(t('col.assignee'), h('span', { text: agentLabel(tk.assignee) })));
      side.appendChild(infoRow(t('col.department'), h('span', { text: deptName(tk.department) })));
    }
    var times = h('div', { class: 'flex flex-col gap-1 text-sm', 'data-sla': '' },
      h('div', {}, t('tk.created') + ': ', h('span', { class: 'tabular-nums', text: fmtDate(tk.createdAt, true) })),
      tk.dueAt ? h('div', { class: tk.overdue ? 'text-error-600 dark:text-error-400' : '' }, t('tk.due') + ': ', h('span', { class: 'tabular-nums', text: fmtDate(tk.dueAt, true) })) : null,
      h('div', {}, t('tk.firstResponse') + ': ', h('span', { class: 'tabular-nums', text: tk.firstResponseAt ? fmtDate(tk.firstResponseAt, true) : t('tk.notYet') })),
      tk.resolvedAt ? h('div', {}, t('tk.resolvedAt') + ': ', h('span', { class: 'tabular-nums', text: fmtDate(tk.resolvedAt, true) })) : null,
      tk.closedAt ? h('div', {}, t('tk.closedAt') + ': ', h('span', { class: 'tabular-nums', text: fmtDate(tk.closedAt, true) })) : null);
    side.appendChild(infoRow(t('tk.times'), times));
    if (tk.rating) {
      side.appendChild(infoRow(t('tk.rating'), h('div', { 'data-rating': String(tk.rating) }, h('div', { class: 'eco-stars', 'aria-label': t('tk.ratingN', { n: tk.rating }) }, [1, 2, 3, 4, 5].map(function (i) { return h('span', { class: i <= tk.rating ? 'eco-star-on' : 'eco-star-off' }, icon('star', 'h-4 w-4')); })),
        tk.ratingComment ? h('p', { class: 'text-sm mt-1', dir: 'auto', text: tk.ratingComment }) : null)));
    }
    if (tk.chatId) side.appendChild(infoRow(t('tk.fromChat'), h('a', { href: hashFor('live-chat', tk.chatId), class: 'eco-link text-sm', text: t('tk.openChat') })));
    grid.appendChild(h('div', { class: 'eco-ticket-side' }, card(h('div', { class: 'p-4' }, side))));
  }

  function load() {
    skeleton();
    baseApi('GET', '/support/tickets/' + encodeURIComponent(id)).then(render, function (err) { clear(grid).appendChild(errorBlock(err, load)); });
  }
  load();
}

function waitText(sec) {
  if (sec < 60) return t('lc.waitSec', { n: localDigits(Math.max(1, sec)) });
  return t('lc.waitMin', { n: localDigits(Math.floor(sec / 60)) });
}

function liveChatPage(container, id, query) {
  loadAgents();
  supportBeat();
  var view = query.view || (id ? 'mine' : 'queue');
  var me = { id: id, seq: 0, stamp: '', poll: 0, chat: null, typingAt: 0 };
  var awayToggle = toggle(!SUPPORT.away, t('lc.available'), function (v) {
    SUPPORT.away = !v;
    try { localStorage.setItem('economy.support.away', SUPPORT.away ? '1' : '0'); } catch (e) { SUPPORT.away = !v; }
    sendBeat().then(function () { toast('success', t(v ? 'lc.nowOnline' : 'lc.nowAway')); });
  }, 'sm');
  awayToggle.setAttribute('data-available', '');
  var frame = pageFrame(t('nav.live-chat'), t('desc.live-chat'), awayToggle);
  clear(container).appendChild(frame.el);
  var listBox = h('div', { class: 'eco-lc-list', 'data-lc-list': '' });
  var tabs = segmented([{ value: 'queue', label: t('lc.queue') }, { value: 'mine', label: t('lc.mine') }, { value: 'all', label: t('lc.all') }, { value: 'closed', label: t('lc.closed') }], view, function (v) {
    view = v;
    replaceHash(hashFor('live-chat', me.id, v === 'queue' && !me.id ? {} : { view: v }));
    loadList();
  }, t('lc.views'));
  tabs.setAttribute('data-lc-tabs', '');
  var pane = h('div', { class: 'eco-lc-pane', 'data-lc-pane': '' });
  frame.content.appendChild(h('div', { class: 'eco-lc' }, h('div', { class: 'eco-lc-side' }, card(h('div', { class: 'p-3 flex flex-col gap-3' }, tabs, listBox))), pane));

  function alive() { return container.isConnected && frame.el.isConnected; }

  function chatItem(c) {
    var on = c.id === me.id;
    var row = h('div', { class: 'eco-lc-item' + (on ? ' eco-lc-item--on' : ''), 'data-chat': c.id, 'data-chat-status': c.status, role: 'button', tabindex: '0',
      onclick: function (e) { if (!e.target.closest('button')) openChat(c.id); },
      onkeydown: function (e) { if (e.key === 'Enter') openChat(c.id); } },
      h('div', { class: 'flex items-center gap-2 min-w-0' },
        h('span', { class: 'eco-dot' + (c.playerOnline ? ' eco-dot--on' : ''), title: c.playerOnline ? t('lc.playerOnline') : t('lc.playerOffline') }),
        h('span', { class: 'font-medium truncate', text: c.displayName || c.username }),
        c.unread && c.agent === SUPPORT.me ? h('span', { class: 'eco-nav-badge tabular-nums', text: fmtNum(c.unread) }) : null,
        h('span', { class: 'flex-1' }),
        c.status === 'waiting' ? h('span', { class: 'text-xs text-warning-700 eco-wait whitespace-nowrap tabular-nums', text: waitText(c.waitingSeconds || 0) }) : c.agent ? h('span', { class: 'text-xs text-secondary-500 truncate', text: c.agent }) : null),
      h('div', { class: 'text-xs text-secondary-500 truncate', dir: 'auto', text: c.playerTyping ? t('lc.typing') : (c.topic ? c.topic + ' · ' : '') + (c.lastPreview || '') }));
    if (c.status === 'waiting' && aclCan('ACCOUNT_NOTES', 1)) {
      var take = button({ variant: 'primary', size: 'sm', label: t('lc.take'), onclick: function () { takeChat(c.id, take); } });
      take.setAttribute('data-take', c.id);
      row.appendChild(h('div', { class: 'mt-2' }, take));
    }
    return row;
  }

  var listSeq = 0;
  function loadList() {
    var my = ++listSeq;
    var params = { view: view === 'all' ? '' : view, limit: 50 };
    baseApi('GET', '/support/chats' + qs(params)).then(function (res) {
      if (my !== listSeq || !alive()) return;
      var items = (res && res.items) || [];
      clear(listBox);
      if (!items.length) listBox.appendChild(emptyBlock(t('lc.empty.' + view), view === 'queue' ? t('lc.empty.queue.hint') : null));
      items.forEach(function (c) { listBox.appendChild(chatItem(c)); });
    }, function (err) { if (my === listSeq) clear(listBox).appendChild(errorBlock(err, loadList)); });
  }

  function takeChat(cid, btn) {
    if (btn) setBusy(btn, true);
    baseApi('POST', '/support/chats/' + encodeURIComponent(cid) + '/take', {}).then(function () {
      toast('success', t('lc.taken'));
      view = 'mine';
      tabs.set('mine');
      refreshSupportBadge(true);
      openChat(cid);
      loadList();
    }, function (err) {
      if (btn) setBusy(btn, false);
      toast('error', err.code === 'ALREADY_TAKEN' ? t('code.ALREADY_TAKEN', { name: err.field === 'agent' ? err.message : err.message }) : errorText(err));
      loadList();
    });
  }

  function openChat(cid) {
    me.id = cid;
    me.seq = 0;
    me.stamp = '';
    me.chat = null;
    me.poll++;
    replaceHash(hashFor('live-chat', cid, view === 'queue' ? {} : { view: view }));
    Array.prototype.forEach.call(listBox.querySelectorAll('[data-chat]'), function (el) { el.classList.toggle('eco-lc-item--on', el.getAttribute('data-chat') === cid); });
    paneSkeleton();
    poll(me.poll, 0);
  }

  function paneSkeleton() {
    clear(pane).appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('40%', 'h-5'), bar('70%', 'h-10'), bar('60%', 'h-10'), bar('100%', 'h-20'))));
  }

  var thread = null;
  var header = null;
  var typingEl = null;
  function renderChat(c, fresh) {
    if (fresh || !thread) {
      clear(pane);
      header = h('div', { class: 'eco-lc-head' });
      thread = h('ol', { class: 'eco-thread eco-lc-thread', 'data-thread': '', 'aria-live': 'polite' });
      typingEl = h('div', { class: 'eco-typing', 'data-typing': '', hidden: true }, h('span', { class: 'eco-typing-dots', 'aria-hidden': 'true' }, h('i', {}), h('i', {}), h('i', {})), h('span', { text: t('lc.playerTypingNow') }));
      var foot = h('div', { 'data-lc-foot': '' });
      pane.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, header, thread, typingEl, foot)));
      pane.__foot = foot;
    }
    paintHeader(c);
    (c.messages || []).forEach(function (m) {
      if (thread.querySelector('[data-seq="' + m.seq + '"]')) return;
      thread.appendChild(chatLine(m, c));
    });
    if ((c.messages || []).length) thread.scrollTop = thread.scrollHeight;
    typingEl.hidden = !c.playerTyping;
    paintFoot(c);
  }

  function chatLine(m, c) {
    if (m.sender === 'system') {
      var d = m.eventData || {};
      var key = 'lc.ev.' + (m.event || 'note');
      return h('li', { class: 'eco-msg-event', 'data-seq': String(m.seq), 'data-event': m.event || '' }, h('span', { dir: 'auto', text: m.event === 'greeting' ? m.body : t(key, { agent: d.agent || d.to || '', from: d.from || '', to: d.to || t('lc.queue'), by: d.by === 'player' ? t('tk.player') : d.by === 'system' ? t('tk.system') : (d.by || '') }) }), h('span', { class: 'text-secondary-400 whitespace-nowrap', text: fmtDate(m.createdAt, true) }));
    }
    return messageBubble({ kind: 'message', sender: m.sender, authorType: m.sender === 'agent' ? 'staff' : 'player', author: m.author, body: m.body, createdAt: m.createdAt, seq: m.seq }, { playerName: c.displayName || c.username });
  }

  function paintHeader(c) {
    clear(header);
    header.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2' },
      h('span', { class: 'eco-dot' + (c.playerOnline ? ' eco-dot--on' : ''), 'data-player-online': c.playerOnline ? 'true' : 'false' }),
      playerLink(c.userId, c.displayName || c.username, 'font-medium'),
      h('span', { class: 'text-xs text-secondary-500', text: c.playerOnline ? t('lc.playerOnline') : t('lc.playerOffline') }),
      tag(t('lc.status.' + c.status), c.status === 'waiting' ? 'warning' : c.status === 'active' ? 'success' : 'secondary'),
      c.agent ? h('span', { class: 'text-xs text-secondary-500', text: t('lc.withAgent', { name: c.agent }) }) : null,
      c.topic ? h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300', dir: 'auto', text: '· ' + c.topic }) : null,
      h('span', { class: 'flex-1' }),
      c.ticketId ? h('a', { href: hashFor('tickets', c.ticketId), class: 'eco-link text-sm', 'data-chat-ticket': '', text: t('lc.ticketN', { n: c.ticketNumber || '' }) }) : null));
  }

  function paintFoot(c) {
    var foot = pane.__foot;
    var key = c.status + '|' + c.agent + '|' + (SUPPORT.me || '');
    if (foot.__key === key) return;
    foot.__key = key;
    clear(foot);
    var write = aclCan('ACCOUNT_NOTES', 1);
    if (!write) return;
    if (c.status === 'waiting') {
      var take = button({ variant: 'primary', size: 'sm', label: t('lc.take'), onclick: function () { takeChat(c.id, take); } });
      take.setAttribute('data-take', c.id);
      foot.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2' }, take, button({ variant: 'ghost', size: 'sm', label: t('lc.close'), onclick: function () { closeDialog(c); } })));
      return;
    }
    if (c.status === 'closed') {
      foot.appendChild(notice('secondary', c.rating ? t('lc.closedRated', { n: c.rating }) : t('lc.closedNote')));
      return;
    }
    if (c.agent !== SUPPORT.me) {
      foot.appendChild(notice('secondary', t('lc.someoneElse', { name: c.agent })));
      return;
    }
    var comp = composer({ notes: false, statusAfter: false, attachments: false, lang: function () { return c.lang || LANGS.def; }, departmentId: c.departmentId, placeholder: 'lc.reply.ph',
      oninput: function () {
        var now = Date.now();
        if (now - me.typingAt < 2500) return;
        me.typingAt = now;
        baseApi('POST', '/support/chats/' + encodeURIComponent(c.id) + '/typing', {}).then(null, function () { return null; });
      },
      onsend: function (v) { return baseApi('POST', '/support/chats/' + encodeURIComponent(c.id) + '/messages', { body: v.body }).then(function (res) { merge(res); }); } });
    var transfer = button({ variant: 'ghost', size: 'sm', icon: 'users', label: t('lc.transfer'), onclick: function () { transferDialog(c); } });
    transfer.setAttribute('data-transfer', '');
    var close = button({ variant: 'danger', size: 'sm', label: t('lc.close'), onclick: function () { closeDialog(c); } });
    close.setAttribute('data-close-chat', '');
    foot.appendChild(comp);
    foot.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 mt-2' }, transfer, close));
  }

  function merge(res) {
    if (!res || res.id !== me.id) return;
    var fresh = !me.chat;
    me.chat = res;
    me.stamp = res.stamp || me.stamp;
    var msgs = res.messages || [];
    if (msgs.length) me.seq = Math.max(me.seq, msgs[msgs.length - 1].seq);
    renderChat(res, fresh);
    if (msgs.length && res.agent === SUPPORT.me && res.status === 'active') baseApi('POST', '/support/chats/' + encodeURIComponent(res.id) + '/read', { seq: me.seq }).then(null, function () { return null; });
  }

  function poll(token2, wait) {
    if (token2 !== me.poll || !alive() || !me.id) return;
    baseApi('GET', '/support/chats/' + encodeURIComponent(me.id) + qs({ after: me.seq, wait: wait, stamp: me.stamp })).then(function (res) {
      if (token2 !== me.poll || !alive()) return;
      merge(res);
      setTimeout(function () { poll(token2, 15); }, 150);
    }, function (err) {
      if (token2 !== me.poll || !alive()) return;
      if (!me.chat) { clear(pane).appendChild(errorBlock(err, function () { openChat(me.id); })); return; }
      setTimeout(function () { poll(token2, 15); }, 3000);
    });
  }

  function transferDialog(c) {
    var to = '';
    return loadAgents(true).then(function (agents) {
      var options = [{ value: '', label: t('lc.backToQueue') }].concat(agents.filter(function (a) { return a.username !== SUPPORT.me; }).map(function (a) { return { value: a.username, label: a.username + ' · ' + (a.online ? t('lc.online') : t('lc.offline')) + ' · ' + t('lc.activeN', { n: a.activeChats || 0 }) }; }));
      var sel = selectInput('', options, function (v) { to = v; }, { label: t('lc.transferTo'), placeholder: false });
      sel.setAttribute('data-transfer-to', '');
      return formDialog({ id: 'chat-transfer', title: t('lc.transfer'), text: t('lc.transfer.text'), icon: 'users', ok: t('lc.transfer.ok'), body: field(t('lc.transferTo'), sel),
        submit: function () { return baseApi('POST', '/support/chats/' + encodeURIComponent(c.id) + '/transfer', { to: to }); } });
    }).then(function (res) {
      if (!res) return;
      toast('success', t('lc.transferred'));
      me.stamp = '';
      if (pane.__foot) pane.__foot.__key = '';
      loadList();
    });
  }

  function closeDialog(c) {
    var toTicket = false;
    var box = checkbox(false, t('lc.toTicket'), function (on) { toTicket = on; });
    box.setAttribute('data-to-ticket', '');
    return formDialog({ id: 'chat-close', title: t('lc.close.title'), text: t('lc.close.text'), icon: 'live-chat', ok: t('lc.close'), danger: true, body: h('div', {}, box, h('p', { class: 'text-xs text-secondary-500 mt-1', text: t('lc.toTicket.hint') })),
      submit: function () { return baseApi('POST', '/support/chats/' + encodeURIComponent(c.id) + '/close', { toTicket: toTicket }); } }).then(function (res) {
      if (!res) return;
      toast('success', res.ticketId ? t('lc.closedTicket', { n: res.ticketNumber || '' }) : t('lc.closedDone'));
      if (pane.__foot) pane.__foot.__key = '';
      merge(res);
      loadList();
      refreshSupportBadge(true);
    });
  }

  loadList();
  var listTimer = setInterval(function () { if (!alive()) { clearInterval(listTimer); return; } loadList(); }, 4000);
  if (id) openChat(id);
  else clear(pane).appendChild(card(emptyBlock(t('lc.pick'), t('lc.pick.hint'))));
}

RESOURCES['support-departments'] = {
  api: 'base',
  path: 'support/departments',
  settings: true,
  filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, false); }, skeleton: '50%' },
    { label: 'col.firstResponse', render: function (it) { return h('span', { class: 'tabular-nums', text: t('dur.hours', { n: localDigits(it.firstResponseHours) }) }); }, width: 130 },
    { label: 'col.agents', render: function (it) { return (it.agents || []).length ? h('div', { class: 'flex flex-wrap gap-1' }, it.agents.slice(0, 4).map(function (a) { return tag(a, 'secondary'); })) : h('span', { class: 'text-secondary-400', text: t('dp.everyone') }); }, width: 200 },
    { label: 'col.open', render: function (it) { return it.open ? h('a', { href: hashFor('tickets', null, { departmentId: it.id, status: 'active' }), class: 'eco-link tabular-nums', text: fmtNum(it.open) }) : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 100 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.dp.key.hint', required: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { label: 'f.dp.name', hint: 'f.dp.name.hint', required: true, max: 100 }),
    tr('description', { label: 'f.dp.description', hint: 'f.dp.description.hint', multiline: true, max: 300, optional: true }),
    { name: 'firstResponseHours', type: 'number', label: 'f.dp.hours', hint: 'f.dp.hours.hint', required: true, min: 1, max: 720 },
    { name: 'agents', type: 'multi', label: 'f.dp.agents', hint: 'f.dp.agents.hint', options: function () {
      var names = (SUPPORT.agents || []).map(function (a) { return a.username; });
      return names.map(function (n) { return { value: n, label: n }; });
    } },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.dp.active.hint' }
  ],
  prepare: function () { return loadAgents(); },
  defaults: function () { return { key: '', name: {}, description: {}, firstResponseHours: 24, agents: [], isActive: true, sort: 100 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 100; if (!p.agents) p.agents = []; return p; }
};

RESOURCES['support-canned'] = {
  api: 'base',
  path: 'support/canned',
  settings: true,
  filters: [
    { name: 'departmentId', label: 'filter.department', type: 'ref', resource: 'support/departments', api: baseApi },
    { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }
  ],
  columns: [
    { label: 'col.title', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.title) || it.key }), h('div', { class: 'text-xs text-secondary-500 eco-clamp', dir: 'auto', text: i18nText(it.body) })); }, skeleton: '60%' },
    { label: 'col.department', render: function (it) { return h('span', { class: it.department ? '' : 'text-secondary-400', text: it.department ? deptName(it.department) : t('cn.allDepts') }); }, width: 170 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.cn.key.hint', required: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('title', { label: 'f.cn.title', hint: 'f.cn.title.hint', required: true, max: 100 }),
    tr('body', { label: 'f.cn.body', hint: 'f.cn.body.hint', required: true, multiline: true, rows: 5, max: 3000 }),
    { name: 'departmentId', type: 'ref', resource: 'support/departments', api: baseApi, label: 'f.cn.department', hint: 'f.cn.department.hint', optional: true },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.cn.active.hint' }
  ],
  defaults: function () { return { key: '', title: {}, body: {}, departmentId: '', isActive: true, sort: 100 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 100; if (!p.departmentId) p.departmentId = ''; return p; }
};

var SUPPORT_SETTINGS_FIELDS = [
  { type: 'heading', label: 'ss.tickets' },
  { type: 'row', fields: [
    { name: 'maxOpenTickets', type: 'number', label: 'f.ss.maxOpen', hint: 'f.ss.maxOpen.hint', required: true, min: 1, max: 100 },
    { name: 'ticketsPerHour', type: 'number', label: 'f.ss.perHour', hint: 'f.ss.perHour.hint', required: true, min: 1, max: 1000 }
  ] },
  { type: 'row', fields: [
    { name: 'messagesPerHour', type: 'number', label: 'f.ss.messages', hint: 'f.ss.messages.hint', required: true, min: 1, max: 10000 },
    { name: 'autoCloseDays', type: 'number', label: 'f.ss.autoClose', hint: 'f.ss.autoClose.hint', required: true, min: 1, max: 365 }
  ] },
  { type: 'row', fields: [
    { name: 'maxAttachments', type: 'number', label: 'f.ss.attachments', hint: 'f.ss.attachments.hint', required: true, min: 0, max: 10 },
    { name: 'maxAttachmentKb', type: 'number', label: 'f.ss.attachmentKb', hint: 'f.ss.attachmentKb.hint', required: true, min: 16, max: 10240 }
  ] },
  { type: 'heading', label: 'ss.chat' },
  { name: 'chatEnabled', type: 'bool', inline: true, label: 'f.ss.chatEnabled', hint: 'f.ss.chatEnabled.hint' },
  tr('chatGreeting', { label: 'f.ss.greeting', hint: 'f.ss.greeting.hint', multiline: true, rows: 2, max: 500, optional: true }),
  tr('chatOffline', { label: 'f.ss.offline', hint: 'f.ss.offline.hint', multiline: true, rows: 2, max: 500, optional: true }),
  { type: 'row', fields: [
    { name: 'chatIdleMinutes', type: 'number', label: 'f.ss.idle', hint: 'f.ss.idle.hint', required: true, min: 5, max: 1440 },
    { name: 'chatMessagesPerMinute', type: 'number', label: 'f.ss.chatRate', hint: 'f.ss.chatRate.hint', required: true, min: 1, max: 600 }
  ] },
  { type: 'heading', label: 'ss.activity' },
  { name: 'activityRetentionDays', type: 'number', label: 'f.ss.retention', hint: 'f.ss.retention.hint', required: true, min: 7, max: 3650 }
];

function supportSettingsPage(container) {
  var frame = tabFrame(t('desc.support-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(baseApi, '/support/settings', null, SUPPORT_SETTINGS_FIELDS, { id: 'support-settings' });
  frame.content.appendChild(page.el);
  page.load();
}

var ACTIVITY_KINDS = ['signin', 'purchase', 'prize', 'event', 'restriction', 'report', 'ticket', 'chat', 'referral', 'config'];
var ACTIVITY_COLOR = { signin: 'secondary', purchase: 'success', prize: 'success', event: 'primary', restriction: 'error', report: 'warning', ticket: 'primary', chat: 'primary', referral: 'success', config: 'secondary' };
var ACTIVITY_ICON = { signin: 'users', purchase: 'payments', prize: 'prizes', event: 'event-types', restriction: 'ban', report: 'reports', ticket: 'tickets', chat: 'live-chat', referral: 'referrals', config: 'remote-config' };
var CONFIG_TYPES = ['string', 'number', 'bool', 'json', 'text'];
var PLATFORMS = ['android', 'ios', 'web', 'windows', 'macos', 'linux'];
var REFERRAL_STATUS_COLOR = { pending: 'warning', held: 'error', completed: 'success', rejected: 'secondary' };

function activityText(it) {
  var d = it.data || {};
  var vars = {};
  Object.keys(d).forEach(function (k) {
    var v = d[k];
    vars[k] = typeof v === 'number' ? fmtNum(v) : Array.isArray(v) ? v.join('، ') : v == null ? '' : String(v);
  });
  if (d.method && has('method.' + d.method)) vars.method = t('method.' + d.method);
  if (d.status && has('tk.status.' + d.status)) vars.status = t('tk.status.' + d.status);
  if (d.kind && has('rs.kind.' + d.kind)) vars.kind = t('rs.kind.' + d.kind);
  if (it.kind === 'referral' && d.status && has('rf.status.' + d.status)) vars.status = t('rf.status.' + d.status);
  if (it.kind === 'config' && Array.isArray(d.variants)) vars.variants = d.variants.join('، ');
  if (it.kind === 'signin' && has('act.method.' + it.action)) vars.method = t('act.method.' + it.action);
  var key = 'act.' + it.kind + '.' + it.action;
  return has(key) ? t(key, vars) : t('act.' + it.kind) + ' · ' + it.action;
}

function activityWhat(it) {
  return h('div', { class: 'flex items-start gap-2 min-w-0', 'data-activity': it.kind + ':' + it.action },
    h('span', { class: 'eco-act-icon eco-act-icon--' + it.kind }, icon(ACTIVITY_ICON[it.kind] || 'activity', 'h-4 w-4')),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'flex items-center gap-2 flex-wrap' }, tag(t('act.' + it.kind), ACTIVITY_COLOR[it.kind]), activityLink(it)),
      h('div', { class: 'text-sm', dir: 'auto', text: activityText(it) })));
}

function activityLink(it) {
  if (it.kind === 'ticket' && it.refId && aclCan('ACCOUNT_NOTES', 0)) return h('a', { href: hashFor('tickets', it.refId), class: 'eco-link text-xs', text: t('act.open') });
  if (it.kind === 'chat' && it.refId && aclCan('ACCOUNT_NOTES', 0)) return h('a', { href: hashFor('live-chat', it.refId, { view: 'all' }), class: 'eco-link text-xs', text: t('act.open') });
  if (it.kind === 'purchase' && it.refId) return h('a', { href: hashFor('payments', it.refId), class: 'eco-link text-xs', text: t('act.open') });
  if (it.kind === 'report' && it.refId && aclCan('CHANNEL_MESSAGE', 0)) return h('a', { href: hashFor('reports', it.refId), class: 'eco-link text-xs', text: t('act.open') });
  if (it.kind === 'referral' && it.refId) return h('a', { href: hashFor('referrals', null, { q: it.username || '' }), class: 'eco-link text-xs', text: t('act.open') });
  return null;
}

function activityWhere(it) {
  var parts = [];
  if (it.country) parts.push(h('span', { class: 'font-mono text-xs', dir: 'ltr', text: it.country }));
  if (it.platform) parts.push(tag(has('pf.' + it.platform) ? t('pf.' + it.platform) : it.platform, 'secondary'));
  if (it.ip) parts.push(mono(it.ip));
  if (!parts.length) return h('span', { class: 'text-secondary-400', text: '—' });
  return h('div', { class: 'flex flex-wrap items-center gap-1' }, parts);
}

function actorText(a) {
  if (!a) return t('act.actor.system');
  if (a === 'player') return t('act.actor.player');
  if (a === 'system') return t('act.actor.system');
  if (a.indexOf('console:') === 0) return t('act.actor.console', { name: a.slice(8) || 'admin' });
  return a;
}

LISTS.activity = {
  api: 'base',
  path: 'activity',
  noRow: true,
  top: function () { return activityTop(); },
  filters: [
    { name: 'kind', label: 'filter.kind', options: function () { return ACTIVITY_KINDS.map(function (k) { return { value: k, label: t('act.' + k) }; }); } },
    { name: 'userId', label: 'filter.playerId', type: 'text' },
    { name: 'platform', label: 'filter.platform', options: function () { return PLATFORMS.map(function (p) { return { value: p, label: has('pf.' + p) ? t('pf.' + p) : p }; }); } },
    { name: 'country', label: 'filter.country', type: 'text' },
    { name: 'from', label: 'filter.from', type: 'text' },
    { name: 'to', label: 'filter.to', type: 'text' }
  ],
  columns: [
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 160 },
    { label: 'col.player', render: function (it) { return playerLink(it.userId, it.displayName || it.username); }, width: 150 },
    { label: 'col.what', render: activityWhat, skeleton: '70%' },
    { label: 'col.where', render: activityWhere, width: 200 },
    { label: 'col.by', render: function (it) { return h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: actorText(it.actor) }); }, width: 140 }
  ]
};

function activityTop() {
  var box = h('div', { class: 'mb-3', 'data-activity-top': '' });
  baseApi('GET', '/activity/summary').then(function (s) {
    if (!s) return;
    box.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: t('act.retention', { n: localDigits(s.retentionDays) }) }));
  }, function () { return null; });
  return box;
}

function activityPage(container, query, ctx) {
  return listPage(container, 'activity', query, ctx);
}

var ACTIVITY_CARD = { id: null, data: null, loading: false, at: 0, err: null, kind: '' };

function loadActivityCard(id) {
  ACTIVITY_CARD.loading = true;
  return baseApi('GET', '/activity' + qs({ userId: id, limit: 8, kind: ACTIVITY_CARD.kind })).then(function (res) {
    ACTIVITY_CARD.loading = false;
    ACTIVITY_CARD.at = Date.now();
    ACTIVITY_CARD.data = res;
    ACTIVITY_CARD.err = null;
    var old = document.querySelector('[data-player-activity]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    ACTIVITY_CARD.loading = false;
    ACTIVITY_CARD.at = Date.now();
    ACTIVITY_CARD.err = err;
    schedule();
  });
}

function playerActivityCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-activity]');
  if (!id || !token() || !aclCan('ACCOUNT', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (ACTIVITY_CARD.id !== id) { ACTIVITY_CARD.id = id; ACTIVITY_CARD.data = null; ACTIVITY_CARD.err = null; ACTIVITY_CARD.at = 0; ACTIVITY_CARD.kind = ''; }
  if (!ACTIVITY_CARD.loading && Date.now() - ACTIVITY_CARD.at > 20000) loadActivityCard(id);
  var p = ACTIVITY_CARD.data;
  var anchor = grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + ACTIVITY_CARD.kind + '|' + (p ? JSON.stringify((p.items || []).map(function (x) { return x.id; })) : ACTIVITY_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-activity': '', 'data-key': key, 'data-economy': 'player-activity', 'data-l10n-skip': '', 'aria-labelledby': 'player-activity-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  var title = h('h2', { id: 'player-activity-title', class: 'font-medium', text: t('act.card') });
  var chips = segmented([{ value: '', label: t('all') }].concat(['signin', 'purchase', 'prize', 'ticket', 'restriction'].map(function (k) { return { value: k, label: t('act.' + k) }; })), ACTIVITY_CARD.kind, function (v) {
    ACTIVITY_CARD.kind = v;
    ACTIVITY_CARD.at = 0;
    loadActivityCard(id);
  }, t('filter.kind'));
  chips.setAttribute('data-activity-chips', '');
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-3' }, title,
    h('div', { class: 'flex flex-wrap items-center gap-3' }, chips,
      h('a', { href: hashFor('activity', null, { userId: id }), class: 'eco-link text-sm', 'data-activity-all': '', text: t('act.all') }),
      aclCan('ACCOUNT_NOTES', 0) ? h('a', { href: hashFor('tickets', null, { userId: id, status: '' }), class: 'eco-link text-sm', text: t('tk.playerTickets') }) : null)));
  if (!p) {
    body.appendChild(ACTIVITY_CARD.err ? notice('error', errorText(ACTIVITY_CARD.err)) : h('div', { class: 'flex flex-col gap-2' }, bar('50%'), bar('70%'), bar('40%')));
    place();
    return;
  }
  var list = h('ol', { class: 'eco-timeline', 'data-timeline': '' });
  (p.items || []).forEach(function (it) {
    list.appendChild(h('li', { class: 'eco-timeline-item', 'data-activity-item': it.kind },
      h('span', { class: 'eco-act-icon eco-act-icon--' + it.kind }, icon(ACTIVITY_ICON[it.kind] || 'activity', 'h-4 w-4')),
      h('div', { class: 'min-w-0 flex-1' },
        h('div', { class: 'text-sm', dir: 'auto', text: activityText(it) }),
        h('div', { class: 'text-xs text-secondary-500 flex flex-wrap gap-2' }, h('span', { text: fmtDate(it.createdAt, true) }), it.country ? h('span', { dir: 'ltr', text: it.country }) : null, it.platform ? h('span', { text: has('pf.' + it.platform) ? t('pf.' + it.platform) : it.platform }) : null)),
      activityLink(it)));
  });
  if (!(p.items || []).length) list.appendChild(h('li', { class: 'text-sm text-secondary-400', text: t('act.empty') }));
  body.appendChild(list);
  place();
}

function referralStatusTag(it) {
  var el = tag(t('rf.status.' + it.status), REFERRAL_STATUS_COLOR[it.status]);
  el.setAttribute('data-referral-status', it.status);
  return el;
}

function referralFlags(flags) {
  if (!flags || !flags.length) return null;
  return h('div', { class: 'flex flex-wrap gap-1 mt-1' }, flags.map(function (f) { var el = tag(has('rf.flag.' + f) ? t('rf.flag.' + f) : f, 'warning'); el.setAttribute('data-flag', f); return el; }));
}

function referralGoal(it) {
  if (!it.goalEventKey) return h('span', { class: 'text-secondary-400', text: t('rf.noGoal') });
  var done = Number(it.goalProgress || 0);
  var need = Number(it.goalValue || 0);
  return h('div', { class: 'text-sm' }, h('span', { class: 'font-mono text-xs', dir: 'ltr', text: it.goalEventKey }), h('div', { class: 'tabular-nums text-xs text-secondary-500', text: t('rf.progress', { n: fmtNum(Math.min(done, need)), total: fmtNum(need) }) }));
}

function referralActions(it, reload) {
  var box = h('div', { class: 'flex flex-wrap gap-1' });
  var tree = button({ variant: 'link', size: 'sm', label: t('rf.tree'), onclick: function () { referralTree(it.referrerId); } });
  tree.setAttribute('data-tree', it.referrerId);
  box.appendChild(tree);
  if (!aclCan('HIRO_ECONOMY', 1)) return box;
  if (it.status === 'held') {
    var ok = button({ variant: 'primary', size: 'sm', label: t('rf.approve'), onclick: function () { decide(it, 'approve', reload); } });
    ok.setAttribute('data-approve', it.id);
    box.appendChild(ok);
  }
  if (it.status === 'held' || it.status === 'pending') {
    var no = button({ variant: 'danger', size: 'sm', label: t('rf.reject'), onclick: function () { decide(it, 'reject', reload); } });
    no.setAttribute('data-reject', it.id);
    box.appendChild(no);
  }
  return box;
}

function decide(it, what, reload) {
  var note = textInput('', { multiline: true, rows: 2, label: t('rs.note'), placeholder: t('rs.note.ph') });
  var v = { note: '' };
  note.addEventListener('input', function () { v.note = note.value; });
  return formDialog({ id: 'referral-' + what, title: t('rf.' + what + '.title'), text: t('rf.' + what + '.text', { from: it.referrerName, to: it.refereeName }), icon: what === 'approve' ? 'check' : 'ban', danger: what === 'reject', ok: t('rf.' + what),
    body: field(t('rs.note'), note, null, { optional: true }),
    submit: function () { return api('POST', '/referrals/' + encodeURIComponent(it.id) + '/' + what, v); } }).then(function (res) {
    if (res) { toast('success', t('rf.' + what + '.done')); if (reload) reload(); }
  });
}

function referralTree(userId) {
  var body = h('div', { class: 'flex flex-col gap-3', 'data-referral-tree': userId }, bar('60%'), bar('80%'));
  confirmModal({ title: t('rf.tree.title'), icon: 'referrals', body: body, ok: t('close'), cancel: t('back') });
  api('GET', '/referrals/tree/' + encodeURIComponent(userId)).then(function (tr2) {
    clear(body);
    var up = (tr2.referredBy || []).slice().reverse();
    var chain = h('div', { class: 'flex flex-wrap items-center gap-1 text-sm' });
    up.forEach(function (u) { chain.appendChild(playerLink(u.userId, u.username)); chain.appendChild(h('span', { class: 'text-secondary-400', text: '←' })); });
    chain.appendChild(h('span', { class: 'font-semibold', text: tr2.user.username }));
    chain.appendChild(mono(tr2.user.code || ''));
    body.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('rf.tree.up') }), chain));
    var list = h('ul', { class: 'flex flex-col gap-1 eco-tree-list' });
    (tr2.children || []).forEach(function (c) {
      list.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-tree-child': c.userId }, playerLink(c.userId, c.username), referralStatusTag(c),
        c.referrals ? button({ variant: 'link', size: 'sm', label: t('rf.tree.more', { n: c.referrals }), onclick: function () { referralTree(c.userId); } }) : null,
        referralFlags(c.flags), h('span', { class: 'text-xs text-secondary-500', text: fmtDate(c.createdAt) })));
    });
    if (!(tr2.children || []).length) list.appendChild(h('li', { class: 'text-sm text-secondary-400', text: t('rf.tree.none') }));
    body.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('rf.tree.down', { n: (tr2.children || []).length }) }), list));
  }, function (err) { clear(body).appendChild(notice('error', errorText(err))); });
}

var referralReload = null;

LISTS.referrals = {
  path: 'referrals',
  noRow: true,
  top: function () { return referralTop(); },
  filters: [
    { name: 'status', label: 'filter.status', options: function () { return ['pending', 'held', 'completed', 'rejected'].map(function (s) { return { value: s, label: t('rf.status.' + s) }; }); } },
    { name: 'flagged', label: 'filter.flagged', options: function () { return [{ value: 'true', label: t('rf.flaggedOnly') }]; } },
    { name: 'referrerId', label: 'filter.referrerId', type: 'text' }
  ],
  columns: [
    { label: 'col.referral', render: function (it) { return h('div', { class: 'flex flex-wrap items-center gap-1 min-w-0' }, playerLink(it.referrerId, it.referrerName), h('span', { class: 'text-secondary-400', text: '←' }), playerLink(it.refereeId, it.refereeName)); }, skeleton: '60%' },
    { label: 'col.code', render: function (it) { return mono(it.code); }, width: 110 },
    { label: 'col.status', render: function (it) { return h('div', {}, referralStatusTag(it), referralFlags(it.flags)); }, width: 170 },
    { label: 'col.goal', render: referralGoal, width: 150 },
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 150 },
    { label: 'col.actions', render: function (it) { return referralActions(it, function () { if (referralReload) referralReload(); }); }, width: 200 }
  ]
};

function referralTop() {
  var box = h('div', { 'data-referral-top': '' }, h('div', { class: 'eco-stat-grid mb-4' }, [1, 2, 3, 4].map(function () { return card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem'), bar('3rem', 'h-7'))); })));
  api('GET', '/referrals/summary').then(function (s) {
    clear(box);
    var grid = h('div', { class: 'eco-stat-grid mb-4' });
    grid.appendChild(statCard(t('rf.stat.total'), s.total, t('rf.stat.today', { n: fmtNum(s.today) }), null));
    grid.appendChild(statCard(t('rf.status.completed'), s.completed, null, hashFor('referrals', null, { status: 'completed' })));
    grid.appendChild(statCard(t('rf.status.pending'), s.pending, null, hashFor('referrals', null, { status: 'pending' })));
    grid.appendChild(statCard(t('rf.status.held'), s.held, t('rf.stat.flagged', { n: fmtNum(s.flagged) }), hashFor('referrals', null, { status: 'held' }), { alert: s.held > 0 }));
    box.appendChild(grid);
    if ((s.topReferrers || []).length) {
      var top = h('ol', { class: 'flex flex-col', 'data-top-referrers': '' });
      s.topReferrers.forEach(function (r, i) {
        top.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 px-4 py-2 border-b border-secondary-200 dark:border-secondary-700 text-sm' },
          h('span', { class: 'tabular-nums text-secondary-500 w-6', text: localDigits(i + 1) }), playerLink(r.userId, r.username, 'font-medium'),
          h('span', { class: 'tabular-nums', text: t('rf.top.line', { n: fmtNum(r.total), done: fmtNum(r.completed) }) }),
          r.flagged ? tag(t('rf.top.flagged', { n: fmtNum(r.flagged) }), 'warning') : null,
          h('span', { class: 'flex-1' }),
          button({ variant: 'link', size: 'sm', label: t('rf.tree'), onclick: function () { referralTree(r.userId); } })));
      });
      box.appendChild(h('div', { class: 'mb-4' }, h('h2', { class: 'text-lg font-medium mb-2', text: t('rf.top') }), card(top)));
    }
  }, function () { clear(box); });
  return box;
}

function referralsPage(container, query, ctx) {
  var list = listPage(container, 'referrals', query, ctx);
  referralReload = list.reload;
  return list;
}

var REFERRAL_FIELDS = [
  { name: 'isEnabled', type: 'bool', inline: true, label: 'f.rr.enabled', hint: 'f.rr.enabled.hint' },
  { type: 'heading', label: 'rr.rewards' },
  { type: 'row', fields: [
    { name: 'referrerPrizeId', type: 'ref', resource: 'prizes', label: 'f.rr.referrerPrize', hint: 'f.rr.referrerPrize.hint', optional: true },
    { name: 'refereePrizeId', type: 'ref', resource: 'prizes', label: 'f.rr.refereePrize', hint: 'f.rr.refereePrize.hint', optional: true }
  ] },
  { name: 'goalEventTypeId', type: 'ref', resource: 'event-types', label: 'f.rr.goal', hint: 'f.rr.goal.hint', optional: true },
  { type: 'row', fields: [
    { name: 'goalMeasure', type: 'select', label: 'f.rr.measure', required: true, noPlaceholder: true, options: function () { return ['count', 'sum', 'max'].map(function (m) { return { value: m, label: t('rr.measure.' + m) }; }); } },
    { name: 'goalValue', type: 'number', int: false, label: 'f.rr.goalValue', hint: 'f.rr.goalValue.hint', required: true }
  ] },
  { type: 'heading', label: 'rr.rules' },
  { type: 'row', fields: [
    { name: 'newAccountDays', type: 'number', label: 'f.rr.days', hint: 'f.rr.days.hint', required: true, min: 0, max: 3650 },
    { name: 'maxPerReferrer', type: 'number', label: 'f.rr.max', hint: 'f.rr.max.hint', required: true, min: 1, max: 100000 }
  ] },
  { name: 'dailyPerReferrer', type: 'number', label: 'f.rr.daily', hint: 'f.rr.daily.hint', required: true, min: 1, max: 100000 },
  { name: 'onePerDevice', type: 'bool', inline: true, label: 'f.rr.device', hint: 'f.rr.device.hint' },
  { name: 'onePerPhone', type: 'bool', inline: true, label: 'f.rr.phone', hint: 'f.rr.phone.hint' },
  { name: 'holdFlagged', type: 'bool', inline: true, label: 'f.rr.hold', hint: 'f.rr.hold.hint' }
];

function referralRulesPage(container) {
  var frame = tabFrame(t('desc.referral-rules'));
  clear(container).appendChild(frame.el);
  var page = apiPage(api, '/referral-settings', null, REFERRAL_FIELDS, { id: 'referral-rules', fixPayload: function (p) {
    ['referrerPrizeId', 'refereePrizeId', 'goalEventTypeId'].forEach(function (k) { if (!p[k]) p[k] = ''; });
    p.goalValue = Number(p.goalValue) || 1;
    return p;
  } });
  frame.content.appendChild(page.el);
  page.load();
}

function configValuePreview(it) {
  var v = it.value;
  if (it.type === 'bool') return tag(v ? t('cf.on') : t('cf.off'), v ? 'success' : 'secondary');
  if (it.type === 'text') return h('span', { class: 'truncate', dir: 'auto', text: i18nText(v) || '' });
  if (it.type === 'json') return h('span', { class: 'font-mono text-xs truncate', dir: 'ltr', text: JSON.stringify(v).slice(0, 60) });
  if (it.type === 'number') return h('span', { class: 'tabular-nums', text: fmtNum(v) });
  return h('span', { class: 'truncate', dir: 'auto', text: String(v == null ? '' : v) });
}

LISTS['remote-config'] = {
  api: 'base',
  path: 'remote-config',
  rowHref: function (it) { return hashFor('remote-config', it.id); },
  create: function () {
    if (!aclCan('SETTINGS', 1)) return null;
    var b = button({ variant: 'primary', size: 'page', icon: 'plus', label: t('cf.new'), onclick: function () { go('remote-config', 'new'); } });
    b.setAttribute('data-config-new', '');
    return b;
  },
  filters: [
    { name: 'type', label: 'filter.type', options: function () { return CONFIG_TYPES.map(function (k) { return { value: k, label: t('cf.type.' + k) }; }); } },
    { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }
  ],
  columns: [
    { label: 'col.key', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-mono text-sm', dir: 'ltr', text: it.key }), it.description ? h('div', { class: 'text-xs text-secondary-500 truncate', dir: 'auto', text: it.description }) : null); }, skeleton: '50%' },
    { label: 'col.type', render: function (it) { return tag(t('cf.type.' + it.type), 'secondary'); }, width: 110 },
    { label: 'col.value', render: configValuePreview, width: 240 },
    { label: 'col.rules', render: function (it) { return it.ruleCount ? h('span', { class: 'tabular-nums', text: t('cf.rulesN', { n: it.ruleCount }) }) : h('span', { class: 'text-secondary-400', text: t('cf.everyone') }); }, width: 120 },
    { label: 'col.version', render: function (it) { return h('span', { class: 'tabular-nums', text: t('cf.v', { n: it.version }) }); }, width: 90 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ]
};

function configPage(container, query, ctx) {
  var list = listPage(container, 'remote-config', query, ctx);
  var recent = h('div', { class: 'mt-6', 'data-config-recent': '' });
  container.appendChild(h('div', { class: 'px-4 md:px-6 max-w-[1600px] mx-auto w-full' }, recent));
  baseApi('GET', '/remote-config/history?limit=8').then(function (res) {
    var items = (res && res.items) || [];
    if (!items.length) return;
    var ol = h('ol', { class: 'flex flex-col' });
    items.forEach(function (x) {
      ol.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 px-4 py-2 border-b border-secondary-200 dark:border-secondary-700 text-sm' },
        tag(t('cf.action.' + x.action), x.action === 'delete' ? 'error' : x.action === 'rollback' ? 'warning' : 'secondary'),
        x.exists ? h('a', { href: hashFor('remote-config', x.keyId), class: 'eco-link font-mono', dir: 'ltr', text: x.key }) : h('span', { class: 'font-mono', dir: 'ltr', text: x.key }),
        h('span', { class: 'tabular-nums text-secondary-500', text: t('cf.v', { n: x.version }) }),
        h('span', { class: 'text-secondary-500', text: x.changedBy }),
        h('span', { class: 'flex-1' }),
        h('span', { class: 'text-xs text-secondary-500', text: fmtDate(x.createdAt, true) })));
    });
    recent.appendChild(h('h2', { class: 'text-lg font-medium mb-2', text: t('cf.recent') }));
    recent.appendChild(card(ol));
  }, function () { return null; });
  return list;
}

function configValueInput(type, value, onchange) {
  if (type === 'bool') return toggle(!!value, t('cf.valueOn'), function (v) { onchange(v); }, 'sm');
  if (type === 'number') return numberInput(value, { int: false, label: t('cf.value'), oninput: function (raw) { onchange(raw === '' ? null : parseNumber(raw, false)); } });
  if (type === 'text') {
    var tri = translatedInput(value || {}, { label: t('cf.value'), multiline: true, rows: 2, max: 2000, oninput: function (v) { onchange(v); } });
    return tri;
  }
  if (type === 'json') {
    var ta = textInput(value == null ? '' : JSON.stringify(value, null, 2), { ltr: true, mono: true, multiline: true, rows: 5, label: t('cf.value'), oninput: function (raw) {
      try { onchange(JSON.parse(raw)); ta.classList.remove('eco-bad'); } catch (e) { onchange({ __invalid: raw }); ta.classList.add('eco-bad'); }
    } });
    return ta;
  }
  return textInput(value == null ? '' : value, { label: t('cf.value'), oninput: function (v) { onchange(v); } });
}

function configValueValid(type, v) {
  if (v && typeof v === 'object' && v.__invalid !== undefined) return t('code.INVALID_JSON');
  if (type === 'number' && (v === null || isNaN(v))) return t('err.number');
  if (type === 'text') { var hit = trCheck(v, { required: true, max: 2000 }); if (hit) return t('code.' + hit.code, { lang: langName(hit.lang) }); }
  if (type === 'string' && (v == null)) return t('err.required');
  return null;
}

function ruleEditor(type, rule, onremove, onmove) {
  var errBox = h('p', { class: 'text-xs text-error-600 dark:text-error-400', hidden: true });
  var name = textInput(rule.name || '', { ltr: true, mono: true, label: t('cf.rule.name'), oninput: function (v) { rule.name = v.trim(); } });
  var plats = h('div', { class: 'flex flex-wrap gap-3 eco-choices' }, PLATFORMS.map(function (p) {
    return checkbox((rule.platforms || []).indexOf(p) !== -1, has('pf.' + p) ? t('pf.' + p) : p, function (on) {
      rule.platforms = (rule.platforms || []).filter(function (x) { return x !== p; });
      if (on) rule.platforms.push(p);
    });
  }));
  var minV = textInput(rule.minVersion || '', { ltr: true, mono: true, label: t('cf.rule.min'), placeholder: '1.0.0', oninput: function (v) { rule.minVersion = v.trim(); } });
  var maxV = textInput(rule.maxVersion || '', { ltr: true, mono: true, label: t('cf.rule.max'), placeholder: '2.0.0', oninput: function (v) { rule.maxVersion = v.trim(); } });
  var tags = refMulti({ resource: 'tags', api: baseApi, filter: { keyType: 'user' }, value: rule.tagIds || [], label: t('cf.rule.tags'), onchange: function (v) { rule.tagIds = v; } });
  var countries = textInput((rule.countries || []).join(', '), { ltr: true, mono: true, label: t('cf.rule.countries'), placeholder: 'IR, DE', oninput: function (v) { rule.countries = v.split(/[\s,،]+/).map(function (c) { return c.trim().toUpperCase(); }).filter(Boolean); } });
  var pct = numberInput(rule.percent == null ? '' : rule.percent, { int: false, label: t('cf.rule.percent'), placeholder: '100', oninput: function (raw) { rule.percent = raw === '' ? null : parseNumber(raw, false); } });
  var valueBox = configValueInput(type, rule.value, function (v) { rule.value = v; });
  var row = h('div', { class: 'eco-rule', 'data-rule': '' },
    h('div', { class: 'flex flex-wrap items-center gap-2 mb-3' },
      h('span', { class: 'font-medium text-sm', text: t('cf.rule') }),
      h('span', { class: 'flex-1' }),
      button({ variant: 'muted', size: 'icon', icon: 'prev', iconClass: 'h-3 w-3 eco-rot90', title: t('cf.rule.up'), onclick: function () { onmove(-1); } }),
      button({ variant: 'muted', size: 'icon', icon: 'next', iconClass: 'h-3 w-3 eco-rot90', title: t('cf.rule.down'), onclick: function () { onmove(1); } }),
      button({ variant: 'danger', size: 'icon', icon: 'trash', iconClass: 'h-3 w-3', title: t('remove'), onclick: onremove })),
    h('div', { class: 'eco-form-row' }, field(t('cf.rule.name'), name, t('cf.rule.name.hint')), field(t('cf.rule.percent'), pct, t('cf.rule.percent.hint'), { optional: true })),
    field(t('cf.rule.platforms'), plats, t('cf.rule.platforms.hint'), { optional: true }),
    h('div', { class: 'eco-form-row' }, field(t('cf.rule.min'), minV, null, { optional: true }), field(t('cf.rule.max'), maxV, null, { optional: true })),
    h('div', { class: 'eco-form-row' }, field(t('cf.rule.tags'), tags, null, { optional: true }), field(t('cf.rule.countries'), countries, t('cf.rule.countries.hint'), { optional: true })),
    field(t('cf.rule.value'), valueBox, t('cf.rule.value.hint')),
    errBox);
  row.setError = function (msg) { errBox.textContent = msg || ''; errBox.hidden = !msg; };
  return row;
}

function configEditor(container, id, query) {
  var isNew = id === 'new';
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('remote-config', null, lastQuery['remote-config'] || {}); } });
  var frame = pageFrame(t(isNew ? 'cf.new' : 'cf.edit'), t('desc.remote-config.edit'), backBtn);
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-config', 'data-config-editor': id });
  frame.content.appendChild(grid);
  var write = aclCan('SETTINGS', 1);

  function skeleton() {
    clear(grid).appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('40%', 'h-5'), bar('100%', 'h-10'), bar('100%', 'h-20'), bar('60%', 'h-10'))));
  }

  function render(item) {
    var v = clone(item);
    v.rules = v.rules || [];
    clear(grid);
    var top = h('div', {});
    var f = {};
    var key = textInput(v.key || '', { ltr: true, mono: true, label: t('cf.key'), oninput: function (x) { v.key = x.trim(); f.key.setError(''); } });
    key.setAttribute('data-config-key', '');
    f.key = field(t('cf.key'), key, t('cf.key.hint'));
    var typeSel = selectInput(v.type || 'string', CONFIG_TYPES.map(function (k) { return { value: k, label: t('cf.type.' + k) }; }), function (x) { v.type = x; v.value = x === 'bool' ? false : x === 'text' ? {} : x === 'json' ? {} : x === 'number' ? 0 : ''; v.rules = []; paintValue(); paintRules(); }, { label: t('cf.type'), placeholder: false });
    typeSel.setAttribute('data-config-type', '');
    if (!isNew) { typeSel.disabled = true; typeSel.classList.add('eco-readonly'); }
    f.type = field(t('cf.type'), typeSel, isNew ? t('cf.type.hint') : t('cf.type.locked'));
    var desc = textInput(v.description || '', { label: t('cf.description'), oninput: function (x) { v.description = x; } });
    f.desc = field(t('cf.description'), desc, t('cf.description.hint'), { optional: true });
    var valueHolder = h('div', {});
    f.value = field(t('cf.default'), valueHolder, t('cf.default.hint'));
    var active = toggle(v.isActive !== false, t('cf.active'), function (x) { v.isActive = x; }, 'sm');
    var rulesBox = h('div', { class: 'flex flex-col gap-3', 'data-rules': '' });
    var addRule = button({ variant: 'ghost', size: 'sm', icon: 'plus', label: t('cf.rule.add'), onclick: function () { v.rules.push({ name: 'rule-' + (v.rules.length + 1), value: v.type === 'bool' ? true : clone(v.value), platforms: [], countries: [], tagIds: [], percent: null }); paintRules(); } });
    addRule.setAttribute('data-rule-add', '');
    var rows = [];
    function paintValue() { clear(valueHolder).appendChild(configValueInput(v.type, v.value, function (x) { v.value = x; f.value.setError(''); })); }
    function paintRules() {
      clear(rulesBox);
      rows = v.rules.map(function (r, i) {
        var row = ruleEditor(v.type, r, function () { v.rules.splice(i, 1); paintRules(); }, function (dir) {
          var j = i + dir;
          if (j < 0 || j >= v.rules.length) return;
          var tmp = v.rules[i]; v.rules[i] = v.rules[j]; v.rules[j] = tmp;
          paintRules();
        });
        rulesBox.appendChild(row);
        return row;
      });
      if (!v.rules.length) rulesBox.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('cf.rule.none') }));
    }
    paintValue();
    paintRules();
    var save = button({ variant: 'primary', size: 'sm', label: isNew ? t('create') : t('save'), onclick: function () { doSave(); } });
    save.setAttribute('data-config-save', '');
    function doSave() {
      clear(top);
      var ok = true;
      if (!/^[a-z][a-z0-9_.]{1,63}$/.test(v.key || '')) { f.key.setError(t('code.INVALID_KEY')); ok = false; }
      var bad = configValueValid(v.type, v.value);
      if (bad) { f.value.setError(bad); ok = false; }
      v.rules.forEach(function (r, i) {
        var e = configValueValid(v.type, r.value);
        if (!e && r.percent != null && (isNaN(r.percent) || r.percent < 0 || r.percent > 100)) e = t('code.INVALID_PERCENT');
        rows[i].setError(e || '');
        if (e) ok = false;
      });
      if (!ok) return;
      var payload = { key: v.key, type: v.type, description: v.description || '', value: v.type === 'text' ? trClean(v.value) : v.value, isActive: v.isActive !== false,
        rules: v.rules.map(function (r) { return { name: r.name, value: v.type === 'text' ? trClean(r.value) : r.value, platforms: r.platforms || [], minVersion: r.minVersion || '', maxVersion: r.maxVersion || '', tagIds: r.tagIds || [], countries: r.countries || [], percent: r.percent == null ? null : r.percent }; }) };
      setBusy(save, true, t('saving'));
      baseApi(isNew ? 'POST' : 'PUT', '/remote-config' + (isNew ? '' : '/' + encodeURIComponent(id)), payload).then(function (res) {
        toast('success', t('saved'));
        if (isNew) go('remote-config', res.id);
        else render(res);
      }, function (err) {
        setBusy(save, false);
        if (err.field === 'key') f.key.setError(errorText(err));
        else top.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
      });
    }
    var del = !isNew && write ? button({ variant: 'danger', size: 'sm', icon: 'trash', label: t('delete'), onclick: function () {
      confirmModal({ title: t('cf.delete.title'), text: t('cf.delete.text', { key: v.key }), ok: t('delete'), danger: true }).then(function (yes) {
        if (!yes) return;
        baseApi('DELETE', '/remote-config/' + encodeURIComponent(id)).then(function () { toast('success', t('deleted')); go('remote-config'); }, function (err) { toast('error', errorText(err)); });
      });
    } }) : null;
    var formCard = card(h('div', { class: 'p-4 md:p-6' }, top, h('div', { class: 'eco-form-row' }, f.key, f.type), f.desc, f.value, h('div', { class: 'mb-5' }, active),
      h('div', { class: 'border-t border-secondary-200 dark:border-secondary-700 pt-4 mt-1' },
        h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-1' }, h('h2', { class: 'font-medium', text: t('cf.rules') }), addRule),
        h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mb-3', text: t('cf.rules.hint') }), rulesBox),
      write ? h('div', { class: 'flex justify-between gap-2 pt-4' }, del || h('span', {}), save) : null));
    var side = h('div', { class: 'eco-config-side flex flex-col gap-4' });
    if (!isNew) {
      side.appendChild(previewCard());
      side.appendChild(historyCard(item, function (res) { render(res); }));
    }
    grid.appendChild(h('div', { class: 'eco-config-main' }, formCard));
    grid.appendChild(side);
  }

  function load() {
    skeleton();
    if (isNew) { freshLanguages().then(function () { render({ key: query.key || '', type: 'string', description: '', value: '', rules: [], isActive: true }); }); return; }
    Promise.all([baseApi('GET', '/remote-config/' + encodeURIComponent(id)), freshLanguages()]).then(function (all) { render(all[0]); }, function (err) { clear(grid).appendChild(errorBlock(err, load)); });
  }
  load();
}

function previewCard() {
  var v = { userId: '', platform: 'web', version: '', country: '' };
  var out = h('div', { class: 'flex flex-col gap-1 mt-2', 'data-preview-out': '' });
  var uid = textInput('', { ltr: true, mono: true, label: t('cf.preview.player'), placeholder: t('cf.preview.player.ph'), oninput: function (x) { v.userId = x.trim(); } });
  var plat = selectInput('web', PLATFORMS.map(function (p) { return { value: p, label: has('pf.' + p) ? t('pf.' + p) : p }; }), function (x) { v.platform = x; }, { label: t('cf.rule.platforms'), placeholder: false });
  var ver = textInput('', { ltr: true, mono: true, label: t('cf.preview.version'), placeholder: '1.4.0', oninput: function (x) { v.version = x.trim(); } });
  var ctry = textInput('', { ltr: true, mono: true, label: t('cf.preview.country'), placeholder: 'IR', oninput: function (x) { v.country = x.trim().toUpperCase(); } });
  var run = button({ variant: 'primary', size: 'sm', label: t('cf.preview.run'), onclick: function () {
    setBusy(run, true);
    baseApi('POST', '/remote-config/preview', Object.assign({ lang: lang() }, v)).then(function (res) {
      setBusy(run, false);
      clear(out);
      Object.keys(res.values || {}).sort().forEach(function (k) {
        var val = res.values[k];
        out.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-preview-key': k },
          h('span', { class: 'font-mono text-xs', dir: 'ltr', text: k }), h('span', { class: 'text-secondary-400', text: '=' }),
          h('span', { class: 'font-medium', dir: 'auto', text: typeof val === 'object' ? JSON.stringify(val) : String(val) }),
          tag(res.variants[k] === 'default' ? t('cf.variant.default') : res.variants[k], res.variants[k] === 'default' ? 'secondary' : 'primary'),
          res.buckets && res.buckets[k] != null ? h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: t('cf.bucket', { n: localDigits((res.buckets[k] / 100).toFixed(2)) }) }) : null));
      });
    }, function (err) { setBusy(run, false); clear(out).appendChild(notice('error', errorText(err))); });
  } });
  run.setAttribute('data-preview-run', '');
  return card(h('div', { class: 'p-4', 'data-config-preview': '' }, h('h2', { class: 'font-medium mb-1', text: t('cf.preview') }), h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('cf.preview.hint') }),
    field(t('cf.preview.player'), uid, null, { optional: true }), h('div', { class: 'eco-form-row' }, field(t('cf.rule.platforms'), plat), field(t('cf.preview.version'), ver, null, { optional: true })), field(t('cf.preview.country'), ctry, null, { optional: true }), run, out));
}

function historyCard(item, onrestored) {
  var list = h('ol', { class: 'flex flex-col', 'data-config-history': '' }, h('li', { class: 'p-3' }, bar('70%')));
  baseApi('GET', '/remote-config/' + encodeURIComponent(item.id) + '/history?limit=20').then(function (res) {
    clear(list);
    ((res && res.items) || []).forEach(function (x) {
      var snap = x.snapshot || {};
      var li = h('li', { class: 'flex flex-wrap items-center gap-2 px-3 py-2 border-b border-secondary-200 dark:border-secondary-700 text-sm', 'data-history-version': String(x.version) },
        h('span', { class: 'tabular-nums font-medium', text: t('cf.v', { n: x.version }) }),
        tag(t('cf.action.' + x.action), 'secondary'),
        h('span', { class: 'text-xs text-secondary-500', text: (x.changedBy || '') + ' · ' + fmtDate(x.createdAt, true) }),
        h('span', { class: 'flex-1' }),
        h('span', { class: 'text-xs truncate eco-hist-val', dir: 'auto', text: configValuePreview({ type: item.type, value: snap.value }).textContent }));
      if (x.version !== item.version && x.action !== 'delete' && aclCan('SETTINGS', 1)) {
        var rb = button({ variant: 'link', size: 'sm', icon: 'history', label: t('cf.rollback'), onclick: function () {
          confirmModal({ title: t('cf.rollback.title', { n: x.version }), text: t('cf.rollback.text'), ok: t('cf.rollback') }).then(function (yes) {
            if (!yes) return;
            baseApi('POST', '/remote-config/' + encodeURIComponent(item.id) + '/rollback', { version: x.version }).then(function (r2) { toast('success', t('cf.rolledBack', { n: x.version })); onrestored(r2); }, function (err) { toast('error', errorText(err)); });
          });
        } });
        rb.setAttribute('data-rollback', String(x.version));
        li.appendChild(rb);
      }
      list.appendChild(li);
    });
  }, function (err) { clear(list).appendChild(h('li', { class: 'p-3' }, notice('error', errorText(err)))); });
  return card(h('div', {}, h('h2', { class: 'font-medium px-4 pt-4 pb-2', text: t('cf.history') }), list));
}

var RARITIES = ['common', 'uncommon', 'rare', 'epic', 'legendary'];
var RARITY_COLOR = { common: 'secondary', uncommon: 'success', rare: 'primary', epic: 'warning', legendary: 'error' };

if (ACTIVITY_KINDS.indexOf('item') === -1) ACTIVITY_KINDS.push('item');
if (CATEGORY_KEY_TYPES.indexOf('item') === -1) CATEGORY_KEY_TYPES.push('item', 'item_type');
if (TAG_KEY_TYPES.indexOf('item') === -1) TAG_KEY_TYPES.splice(TAG_KEY_TYPES.indexOf('general'), 0, 'item');

function rarityTag(r) {
  return tag(t('rarity.' + (r || 'common')), RARITY_COLOR[r] || 'secondary');
}

function priceLabel(p) {
  if (!p || p.amount == null) return '';
  return t('amount.of', { n: fmtNum(p.amount), name: keyName('currency', p.currencyKey) });
}

function itemGetCell(it) {
  return h('div', { class: 'flex flex-col items-start gap-1 min-w-0 text-sm' }, withKeyNames(function () {
    var out = [];
    if (it.isStarter) out.push(tag(t('it.get.starter'), 'success'));
    if (it.price) out.push(h('span', { class: 'tabular-nums whitespace-nowrap', text: priceLabel(it.price) }));
    if (it.unlockLevel) out.push(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('it.get.level', { n: fmtNum(it.unlockLevel) }) }));
    if (!it.isStarter && !it.price && !it.unlockLevel) out.push(h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: t('it.get.prize') }));
    if (it.requiresKey || it.parentKey) out.push(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 truncate', text: t('it.get.requires', { key: it.requiresKey || it.parentKey }) }));
    if ((it.upgrades || []).length) out.push(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('it.upgradesN', { n: fmtNum(it.upgrades.length) }) }));
    return out;
  }));
}

RESOURCES.items = {
  filters: [
    { name: 'typeId', label: 'filter.it.type', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'item_type' } },
    { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'item' } },
    { name: 'rarity', label: 'filter.it.rarity', options: opts('rarity.', RARITIES) },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '45%' },
    { label: 'col.it.type', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, tag(i18nText(it.typeName) || it.typeKey, 'secondary'), rarityTag(it.rarity)); }, width: 170 },
    { label: 'col.it.get', render: itemGetCell, skeleton: '40%', width: 190 },
    { label: 'col.it.owners', render: function (it) {
      return h('div', { class: 'flex flex-col text-sm' }, h('span', { class: 'tabular-nums', text: t('it.ownersN', { n: fmtNum(it.owners || 0) }) }),
        it.equippedBy ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('it.equippedN', { n: fmtNum(it.equippedBy) }) }) : null);
    }, width: 150 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    tr('name', { required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { type: 'row', fields: [
      { name: 'typeId', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'item_type' }, quick: 'item_type', label: 'f.it.type', hint: 'f.it.type.hint', required: true },
      { name: 'rarity', type: 'select', label: 'f.it.rarity', required: true, options: opts('rarity.', RARITIES) }
    ] },
    { type: 'row', fields: [categoryField('item'), tagsField('item')] },
    { type: 'heading', label: 'h.it.get' },
    { name: 'isStarter', type: 'bool', inline: true, label: 'f.it.starter', hint: 'f.it.starter.hint' },
    { type: 'row', when: function (v) { return !v.isStarter; }, fields: [
      { name: 'price.currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.it.price.currency', optional: true,
        check: function (c, v) { return !c && v.price && v.price.amount != null ? t('err.required') : null; } },
      { name: 'price.amount', type: 'number', label: 'f.it.price.amount', hint: 'f.it.price.hint', optional: true, min: 1,
        check: function (n, v) { return n == null && v.price && v.price.currencyKey ? t('err.required') : null; } }
    ] },
    { type: 'row', when: function (v) { return !v.isStarter; }, fields: [
      { name: 'unlockLevel', type: 'number', label: 'f.it.unlockLevel', hint: 'f.it.unlockLevel.hint', optional: true, min: 1 },
      { name: 'levelTableId', type: 'ref', resource: 'level-tables', label: 'f.it.levelTable', hint: 'f.it.levelTable.hint', optional: true }
    ] },
    { type: 'row', fields: [
      { name: 'requiresKey', type: 'ref', resource: 'items', valueKey: 'key', label: 'f.it.requires', hint: 'f.it.requires.hint', optional: true,
        check: function (v, vals) { return v && v === vals.key ? t('code.SELF_REFERENCE') : null; } },
      { name: 'parentKey', type: 'ref', resource: 'items', valueKey: 'key', label: 'f.it.parent', optional: true,
        check: function (v, vals) { return v && v === vals.key ? t('code.SELF_REFERENCE') : null; } }
    ] },
    { type: 'heading', label: 'h.it.upgrades' },
    { name: 'upgrades', type: 'rows', addLabel: 'it.up.add', emptyText: 'it.up.empty', max: 50,
      newRow: function (rows) { var last = rows[rows.length - 1]; return { currencyKey: last ? last.currencyKey : null, amount: last ? Number(last.amount || 0) * 2 || 10 : 10 }; },
      columns: [
        { name: 'currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.it.up.currency', required: true },
        { name: 'amount', type: 'number', label: 'f.it.up.amount', required: true, min: 1 }
      ] },
    { type: 'heading', label: 'h.it.services' },
    { name: 'services', type: 'rows', addLabel: 'it.sv.add', emptyText: 'it.sv.empty', max: 20,
      newRow: function () { return { key: '', currencyKey: null, unitPrice: 1, maxUnits: 1, name: {} }; },
      check: duplicateCheck('key', 'DUPLICATE_SERVICE'),
      columns: [
        { name: 'key', type: 'text', ltr: true, mono: true, label: 'f.it.sv.key', required: true },
        { name: 'currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.it.sv.currency', required: true },
        { name: 'unitPrice', type: 'number', label: 'f.it.sv.price', required: true, min: 1 },
        { name: 'maxUnits', type: 'number', label: 'f.it.sv.max', required: true, min: 1, max: 1000 },
        { name: 'name', type: 'i18n', label: 'f.it.sv.name', max: 64, wide: true }
      ] },
    { type: 'heading', label: 'f.icon' },
    { type: 'row', fields: [
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { name: 'imageUrl', type: 'image', label: 'f.it.image', optional: true }
    ] },
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' },
    { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
  ],
  defaults: function (query) {
    return { key: '', isActive: true, name: {}, description: {}, typeId: (query && query.typeId) || null, rarity: 'common', categoryId: null, tagIds: [], isStarter: false,
      price: null, unlockLevel: null, levelTableId: null, requiresKey: null, parentKey: null, upgrades: [], services: [], iconUrl: '', imageUrl: '', sort: 0, metadata: {} };
  },
  fixPayload: function (p) {
    if (p.isStarter || !p.price || (p.price.currencyKey == null && p.price.amount == null)) p.price = null;
    if (p.isStarter) { p.unlockLevel = null; p.levelTableId = null; }
    if (!p.iconUrl) p.iconUrl = '';
    if (!p.imageUrl) p.imageUrl = '';
    if (p.sort == null) p.sort = 0;
    p.requiresKey = p.requiresKey || '';
    p.parentKey = p.parentKey || '';
    p.levelTableId = p.levelTableId || '';
    p.categoryId = p.categoryId || '';
    p.upgrades = (p.upgrades || []).map(function (r) { return { currencyKey: r.currencyKey, amount: r.amount }; });
    p.services = (p.services || []).map(function (r) { return { key: r.key, currencyKey: r.currencyKey, unitPrice: r.unitPrice, maxUnits: r.maxUnits, name: r.name || {} }; });
    return p;
  },
  afterSave: function () { forgetRefs('items'); }
};

OVERVIEW_CATALOG.push('items');

(function productRules() {
  var def = RESOURCES.products;
  var at = -1;
  def.fields.forEach(function (f, i) { if (f.type === 'row' && f.fields.some(function (x) { return x.name === 'minLevel'; })) at = i; });
  var rules = [
    { type: 'heading', label: 'h.pr.rules' },
    { type: 'row', fields: [
      { name: 'minLevel', type: 'number', label: 'f.minLevel', min: 0, empty: 0 },
      { name: 'maxLevel', type: 'number', label: 'f.maxLevel', hint: 'f.maxLevel.hint', optional: true, min: 0,
        check: function (v, vals) { return v != null && Number(v) < Number(vals.minLevel || 0) ? t('code.INVALID_LEVEL_RANGE') : null; } }
    ] },
    { name: 'levelTableId', type: 'ref', resource: 'level-tables', label: 'f.pr.levelTable', hint: 'f.pr.levelTable.hint', optional: true },
    { type: 'row', fields: [
      { name: 'isOneTime', type: 'bool', inline: true, label: 'f.isOneTime', hint: 'f.isOneTime.hint', empty: false },
      { name: 'checkContents', type: 'bool', inline: true, label: 'f.checkContents', hint: 'f.checkContents.hint' }
    ] },
    tr('purchaseMessage', { multiline: true, max: 500, optional: true }),
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
  ];
  if (at !== -1) def.fields.splice(at, 1);
  var meta = -1;
  def.fields.forEach(function (f, i) { if (f.name === 'metadata') meta = i; });
  Array.prototype.splice.apply(def.fields, [meta === -1 ? def.fields.length : meta, 0].concat(rules));
  var baseDefaults = def.defaults;
  def.defaults = function (q) { return Object.assign(baseDefaults(q), { maxLevel: null, levelTableId: null, isOneTime: false, checkContents: true, purchaseMessage: {} }); };
  var baseFix = def.fixPayload;
  def.fixPayload = function (p) { p = baseFix(p); p.levelTableId = p.levelTableId || ''; if (p.isOneTime == null) p.isOneTime = false; if (p.checkContents == null) p.checkContents = true; return p; };
  def.columns.splice(3, 0, { label: 'h.pr.rules', render: function (it) {
    var box = h('div', { class: 'flex flex-wrap gap-1', 'data-product-rules': '' });
    if (it.isOneTime) box.appendChild(tag(t('pr.oneTime'), 'warning'));
    if (it.minLevel && it.maxLevel != null) box.appendChild(tag(t('pr.levels', { min: fmtNum(it.minLevel), max: fmtNum(it.maxLevel) }), 'secondary'));
    else if (it.minLevel) box.appendChild(tag(t('pr.levelFrom', { n: fmtNum(it.minLevel) }), 'secondary'));
    else if (it.maxLevel != null) box.appendChild(tag(t('pr.levelTo', { n: fmtNum(it.maxLevel) }), 'secondary'));
    if (!box.childNodes.length) box.appendChild(h('span', { class: 'text-secondary-400', text: t('none') }));
    return box;
  }, width: 150 });
})();

var INV_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadInvCard(id) {
  INV_CARD.loading = true;
  return api('GET', '/users/' + encodeURIComponent(id) + '/items' + qs({ lang: lang() })).then(function (res) {
    INV_CARD.loading = false;
    INV_CARD.at = Date.now();
    INV_CARD.data = res;
    INV_CARD.err = null;
    var old = document.querySelector('[data-player-items]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    INV_CARD.loading = false;
    INV_CARD.at = Date.now();
    INV_CARD.err = err;
    schedule();
  });
}

function reloadInv(id) { INV_CARD.at = 0; loadInvCard(id); }

function invLine(userId, it, write) {
  var pic = h('div', { class: 'h-8 w-8 shrink-0 rounded-md bg-secondary-100 dark:bg-secondary-700 overflow-hidden flex items-center justify-center text-secondary-500' });
  pic.appendChild(it.iconUrl ? img(it.iconUrl, 'items') : icon('items', 'h-4 w-4'));
  var line = h('div', { class: 'flex flex-wrap items-center gap-2 py-1.5 text-sm', 'data-inv-item': it.key, 'data-equipped': it.equipped ? '' : null },
    pic,
    h('span', { class: 'font-medium truncate min-w-0', text: it.name || it.key }),
    tag(it.typeName || it.type, 'secondary'),
    rarityTag(it.rarity),
    it.equipped ? tag(t('inv.equipped'), 'success') : null,
    !it.isActive ? tag(t('inv.inactive'), 'secondary') : null,
    h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 whitespace-nowrap', text: t('inv.source.' + it.source) }),
    it.maxUpgradeLevel ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 whitespace-nowrap tabular-nums', text: t('inv.upgrade', { n: fmtNum(it.upgradeLevel || 0), max: fmtNum(it.maxUpgradeLevel) }) }) : null,
    h('span', { class: 'flex-1' }));
  if (write && !it.isStarter) {
    var rb = button({ variant: 'link', size: 'sm', label: t('inv.revoke'), onclick: function () {
      confirmModal({ title: t('inv.revoke.title'), text: t('inv.revoke.text', { name: it.name || it.key }), ok: t('inv.revoke'), danger: true }).then(function (yes) {
        if (!yes) return;
        api('DELETE', '/users/' + encodeURIComponent(userId) + '/items/' + encodeURIComponent(it.id)).then(function () { toast('success', t('inv.revoke.done')); reloadInv(userId); }, function (err) { toast('error', errorText(err)); });
      });
    } });
    rb.setAttribute('data-inv-revoke', it.key);
    line.appendChild(rb);
  }
  return line;
}

function playerItemsCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-items]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (INV_CARD.id !== id) { INV_CARD.id = id; INV_CARD.data = null; INV_CARD.err = null; INV_CARD.at = 0; }
  if (!INV_CARD.loading && Date.now() - INV_CARD.at > 20000) loadInvCard(id);
  var p = INV_CARD.data;
  var anchor = grid.querySelector('[data-player-tasks]') || grid.querySelector('[data-player-activity]') || grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify([p.items, p.equipped]) : INV_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-items': '', 'data-key': key, 'data-economy': 'player-items', 'data-l10n-skip': '', 'aria-labelledby': 'player-items-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-items-title', class: 'font-medium flex items-center gap-2' }, icon('items', 'h-4 w-4'), h('span', { text: t('inv.card') }));
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(INV_CARD.err ? notice('error', errorText(INV_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { reloadInv(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var write = aclCan('HIRO_ECONOMY', 1);
  var owned = (p.items || []).filter(function (it) { return it.owned; });
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' }, title,
    h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300 tabular-nums', 'data-inv-counts': '', text: t('inv.counts', { n: fmtNum(owned.length), total: fmtNum((p.items || []).length) }) }),
    h('a', { href: hashFor('items'), class: 'eco-link text-sm', text: t('inv.all') })));
  var eq = Object.keys(p.equipped || {});
  if (eq.length) {
    var byKey = {};
    (p.items || []).forEach(function (it) { byKey[it.key] = it; });
    body.appendChild(h('div', { class: 'flex flex-wrap gap-1 mb-2', 'data-inv-equipped': '' }, eq.sort().map(function (slot) {
      var it = byKey[p.equipped[slot]];
      return tag(t('inv.slot', { slot: it ? it.typeName || slot : slot, name: it ? it.name || it.key : p.equipped[slot] }), 'success');
    })));
  }
  var list = h('div', { class: 'flex flex-col', 'data-inv-list': '' });
  owned.forEach(function (it) { list.appendChild(invLine(id, it, write)); });
  if (!owned.length) list.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('inv.none') }));
  body.appendChild(list);
  if (write) {
    var pick = '';
    var picker = refPicker({ resource: 'items', value: '', label: t('inv.grant.pick'), placeholder: t('inv.grant.pick'), onchange: function (v) { pick = v || ''; } });
    var give = button({ variant: 'primary', size: 'sm', icon: 'plus', label: t('inv.grant'), onclick: function () {
      if (!pick) { toast('error', t('inv.grant.pick')); return; }
      api('POST', '/users/' + encodeURIComponent(id) + '/items', { itemId: pick }).then(function () { toast('success', t('inv.grant.done')); reloadInv(id); }, function (err) { toast('error', errorText(err)); });
    } });
    give.setAttribute('data-inv-grant', '');
    body.appendChild(h('div', { class: 'mt-3 border-t border-secondary-200 dark:border-secondary-700 pt-4 flex flex-wrap items-center gap-2', 'data-inv-give': '' },
      h('div', { class: 'min-w-0 flex-1 eco-inv-pick' }, picker), give));
  }
  place();
}

PH_ICONS.achievements = '<path d="M216,96A88,88,0,1,0,72,163.83V240a8,8,0,0,0,11.58,7.16L128,225l44.43,22.21A8.07,8.07,0,0,0,176,248a8,8,0,0,0,8-8V163.83A87.85,87.85,0,0,0,216,96ZM56,96a72,72,0,1,1,72,72A72.08,72.08,0,0,1,56,96ZM168,227.06l-36.43-18.21a8,8,0,0,0-7.16,0L88,227.06V174.37a87.89,87.89,0,0,0,80,0ZM128,152A56,56,0,1,0,72,96,56.06,56.06,0,0,0,128,152Zm0-96A40,40,0,1,1,88,96,40,40,0,0,1,128,56Z"/>';
PH_ICONS.medalDuo = '<path d="M176,96a48,48,0,1,1-48-48A48,48,0,0,1,176,96Z" opacity="0.2"/>' + PH_ICONS.achievements;
PH_ICONS.lockDuo = '<path d="M208,96V208a8,8,0,0,1-8,8H56a8,8,0,0,1-8-8V96a8,8,0,0,1,8-8H200A8,8,0,0,1,208,96Z" opacity="0.2"/><path d="M208,80H176V56a48,48,0,0,0-96,0V80H48A16,16,0,0,0,32,96V208a16,16,0,0,0,16,16H208a16,16,0,0,0,16-16V96A16,16,0,0,0,208,80ZM96,56a32,32,0,0,1,64,0V80H96ZM208,208H48V96H208V208Zm-68-56a12,12,0,1,1-12-12A12,12,0,0,1,140,152Z"/>';

if (ACTIVITY_KINDS.indexOf('achievement') === -1) ACTIVITY_KINDS.push('achievement');
ACTIVITY_COLOR.achievement = 'warning';
ACTIVITY_ICON.achievement = 'achievements';
if (CATEGORY_KEY_TYPES.indexOf('achievement') === -1) CATEGORY_KEY_TYPES.push('achievement');
if (TAG_KEY_TYPES.indexOf('achievement') === -1) TAG_KEY_TYPES.splice(TAG_KEY_TYPES.indexOf('general'), 0, 'achievement');

var ACH_SOURCES = ['event', 'task', 'manual'];



function isEventAch(v) { return v.source === 'event'; }

function achTierLine(it) {
  var tiers = it.tiers || [];
  if (it.source !== 'event') return '';
  return tiers.map(function (r, i) { return t('ac.tier.at', { name: i18nText(r.name) || t('ac.tier.n', { n: fmtNum(r.tier || i + 1) }), n: numLocal(r.threshold) }); }).join(lang() === 'fa' ? '، ' : ' · ');
}

function achSentence(it, name) {
  if (it.source === 'task') return t(it.taskGroupId ? 'ac.when.group' : 'ac.when.task', { name: name });
  if (it.source !== 'event') return t('ac.when.manual');
  var tiers = it.tiers || [];
  var first = tiers.length ? tiers[0].threshold : '';
  var parts = [t('ac.when.' + (it.metric || 'count'), { event: name, n: numLocal(first) })];
  if (it.valueMin != null && it.valueMax != null) parts.push(t('ac.filter.range', { min: numLocal(it.valueMin), max: numLocal(it.valueMax) }));
  else if (it.valueMin != null) parts.push(t('ac.filter.min', { n: numLocal(it.valueMin) }));
  else if (it.valueMax != null) parts.push(t('ac.filter.max', { n: numLocal(it.valueMax) }));
  return parts.join(' · ');
}

function achSentenceCell(it) {
  var box = h('div', { class: 'text-sm text-secondary-700 dark:text-secondary-200 eco-clamp', 'data-ach-sentence': '' });
  var res = it.source === 'event' ? ['event-types', it.eventTypeId] : it.source === 'task' ? (it.taskGroupId ? ['task-groups', it.taskGroupId] : ['tasks', it.taskId]) : null;
  if (!res || !res[1]) { box.appendChild(document.createTextNode(achSentence(it, ''))); return box; }
  var ref = refName(res[0], res[1]);
  function paint(x) { clear(box).appendChild(document.createTextNode(achSentence(it, x ? i18nText(x.name) || x.key : '…'))); }
  if (ref.item !== undefined) paint(ref.item);
  else { box.appendChild(bar('70%')); ref.p.then(paint); }
  return box;
}

function achBadge(it, unlocked) {
  var tile = h('div', { class: 'h-9 w-9 shrink-0 rounded-md overflow-hidden flex items-center justify-center eco-ach-tile', 'data-rarity': it.rarity || 'common', 'data-locked': unlocked === false ? '' : null });
  var src = unlocked === false && it.lockedIconUrl ? it.lockedIconUrl : it.iconUrl;
  if (src) tile.appendChild(img(src, 'achievements'));
  else tile.appendChild(icon(unlocked === false ? 'lockDuo' : 'medalDuo', 'h-5 w-5'));
  return tile;
}

RESOURCES.achievements = {
  filters: [
    { name: 'source', label: 'filter.ac.source', options: opts('ac.source.', ACH_SOURCES) },
    { name: 'rarity', label: 'filter.it.rarity', options: opts('rarity.', RARITIES) },
    { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'achievement' } },
    { name: 'isHidden', label: 'filter.ac.hidden', options: function () { return [{ value: 'true', label: t('yes') }, { value: 'false', label: t('no') }]; } },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) {
      var name = i18nText(it.name) || '';
      return h('div', { class: 'flex items-center gap-3 min-w-0' }, achBadge(it),
        h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: name }), h('div', { class: 'text-secondary-500' }, mono(it.key))));
    }, skeleton: '40%' },
    { label: 'col.ac.when', render: function (it) {
      var box = h('div', { class: 'flex flex-col gap-1 min-w-0' }, achSentenceCell(it));
      var line = achTierLine(it);
      if (line) box.appendChild(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums', 'data-ach-tiers': '', text: line }));
      return box;
    }, skeleton: '55%' },
    { label: 'col.ac.tiers', render: function (it) {
      var tiers = it.tiers || [];
      var prizes = tiers.filter(function (r) { return r.prizeId; });
      var box = h('div', { class: 'flex flex-col gap-1 min-w-0 text-sm' }, h('span', { class: 'tabular-nums', text: t('ac.tiersN', { n: fmtNum(tiers.length) }) }));
      if (!prizes.length) box.appendChild(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ac.noPrize') }));
      else box.appendChild(h('div', { class: 'min-w-0' }, refNameCell('prizes', prizes[0].prizeId)));
      return box;
    }, width: 170 },
    { label: 'col.ac.rarity', render: function (it) {
      return h('div', { class: 'flex flex-wrap items-center gap-1' }, rarityTag(it.rarity), h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums whitespace-nowrap', text: t('ac.pointsN', { n: fmtNum(it.points || 0) }) }));
    }, width: 160 },
    { label: 'col.ac.players', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: t('ac.unlockedN', { n: fmtNum(it.unlocked || 0) }) }); }, width: 110 },
    { label: 'col.status', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive), it.isHidden ? tag(t('ac.hidden'), 'secondary') : null, tag(t('ac.source.' + it.source), it.source === 'event' ? 'primary' : it.source === 'task' ? 'success' : 'secondary'));
    }, width: 160 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    tr('name', { label: 'f.ac.name', required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    tr('hint', { label: 'f.hint', hint: 'f.ac.hint.hint', max: 300, optional: true }),
    { type: 'heading', label: 'ac.h.look' },
    { type: 'row', fields: [
      { name: 'rarity', type: 'select', label: 'f.ac.rarity', required: true, noPlaceholder: true, options: opts('rarity.', RARITIES) },
      { name: 'points', type: 'number', label: 'f.ac.points', hint: 'f.ac.points.hint', min: 0, max: 100000, empty: 0 }
    ] },
    { type: 'row', fields: [categoryField('achievement'), tagsField('achievement')] },
    { type: 'row', fields: [
      { name: 'iconUrl', type: 'image', label: 'f.ac.icon', optional: true },
      { name: 'lockedIconUrl', type: 'image', label: 'f.ac.lockedIcon', optional: true }
    ] },
    { name: 'isHidden', type: 'bool', inline: true, label: 'f.ac.hidden', hint: 'f.ac.hidden.hint', empty: false },
    { type: 'heading', label: 'ac.h.when' },
    { name: 'source', type: 'select', label: 'f.ac.source', hint: 'f.ac.source.hint', required: true, noPlaceholder: true, options: opts('ac.source.', ACH_SOURCES) },
    { name: 'eventTypeId', type: 'ref', resource: 'event-types', label: 'f.ev.event', required: true, when: isEventAch },
    { type: 'row', when: isEventAch, fields: [
      { name: 'metric', type: 'select', label: 'f.ac.metric', hint: 'f.tk.metric.hint', required: true, noPlaceholder: true, options: opts('tk.metric.', ['count', 'sum_value', 'high_value']), when: isEventAch },
      { name: 'valueMin', type: 'number', int: false, label: 'f.tk.valueMin', hint: 'f.tk.valueFilter.hint', optional: true, when: isEventAch },
      { name: 'valueMax', type: 'number', int: false, label: 'f.tk.valueMax', optional: true, when: isEventAch,
        check: function (v, vals) { return v != null && vals.valueMin != null && Number(v) < Number(vals.valueMin) ? t('code.INVALID_RANGE') : null; } }
    ] },
    { type: 'row', when: function (v) { return v.source === 'task'; }, fields: [
      { name: 'taskId', type: 'ref', resource: 'tasks', filter: { level: 'top' }, label: 'f.ac.task', hint: 'f.ac.task.hint', optional: true, when: function (v) { return v.source === 'task'; },
        check: function (v, vals) { return !v && !vals.taskGroupId ? t('code.TASK_REQUIRED') : v && vals.taskGroupId ? t('code.TASK_AND_GROUP') : null; } },
      { name: 'taskGroupId', type: 'ref', resource: 'task-groups', label: 'f.ac.group', hint: 'f.ac.group.hint', optional: true, when: function (v) { return v.source === 'task'; } }
    ] },
    { name: 'preview', type: 'static', readonly: true, live: true, render: function (v) {
      if (v.source === 'event' && !v.eventTypeId) return h('span', {});
      var item = { source: v.source, eventTypeId: v.eventTypeId, metric: v.metric, valueMin: v.valueMin, valueMax: v.valueMax, taskId: v.taskId, taskGroupId: v.taskGroupId, tiers: v.source === 'event' ? v.tiers : [] };
      return h('div', { class: 'eco-rule-preview text-sm rounded-md px-3 py-2 flex flex-col gap-1', 'data-ach-preview': '' }, achSentenceCell(item),
        v.source === 'event' && (v.tiers || []).length ? h('span', { class: 'text-xs tabular-nums', text: achTierLine(item) }) : null);
    } },
    { type: 'heading', label: 'ac.h.tiers', when: isEventAch },
    { name: 'tiers', type: 'rows', addLabel: 'ac.tier.add', emptyText: 'ac.tier.empty', min: 1, max: 10, when: isEventAch, empty: [], refresh: true,
      newRow: function (rows) {
        var last = rows[rows.length - 1];
        return { threshold: last ? Number(last.threshold || 0) * 3 || 10 : 1, name: {}, prizeId: null, grantMode: 'claim' };
      },
      check: function (rows) {
        for (var i = 1; i < rows.length; i++) if (Number(rows[i].threshold) <= Number(rows[i - 1].threshold)) return { index: i, message: t('ac.thresholds.order') };
        return null;
      },
      columns: [
        { name: 'threshold', type: 'number', int: false, label: 'f.ac.tier.threshold', required: true, min: 0.0001 },
        { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', optional: true },
        { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', options: grantOptions },
        { name: 'name', type: 'i18n', label: 'f.ac.tier.name', max: 64, wide: true }
      ] },
    { name: 'tiersHint', type: 'static', readonly: true, when: isEventAch, render: function () { return h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('ac.tiers.hint') }); } },
    { type: 'heading', label: 'ac.h.prize', when: function (v) { return !isEventAch(v); } },
    { type: 'row', when: function (v) { return !isEventAch(v); }, fields: [
      { name: 'unlockPrizeId', type: 'ref', resource: 'prizes', label: 'f.prize', optional: true, when: function (v) { return !isEventAch(v); } },
      { name: 'unlockGrantMode', type: 'select', label: 'f.ev.grantMode', hint: 'f.ev.grantMode.hint', noPlaceholder: true, options: grantOptions, when: function (v) { return !isEventAch(v); } }
    ] },
    { name: 'notify', type: 'bool', inline: true, label: 'f.ac.notify', hint: 'f.ac.notify.hint', empty: false },
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
  ],
  defaults: function (query) {
    return { key: '', isActive: true, name: {}, description: {}, hint: {}, rarity: 'common', points: 10, categoryId: null, tagIds: [], iconUrl: '', lockedIconUrl: '', isHidden: false,
      source: (query && query.source) || 'event', eventTypeId: (query && query.eventTypeId) || null, metric: 'count', valueMin: null, valueMax: null, taskId: null, taskGroupId: null,
      tiers: [{ threshold: 1, name: {}, prizeId: null, grantMode: 'claim' }], unlockPrizeId: null, unlockGrantMode: 'claim', notify: false, sort: 0 };
  },
  toForm: function (v) {
    var first = (v.tiers || [])[0] || {};
    v.unlockPrizeId = first.prizeId || null;
    v.unlockGrantMode = first.grantMode || 'claim';
    return v;
  },
  fixPayload: function (p) {
    if (p.sort == null) p.sort = 0;
    if (p.points == null) p.points = 0;
    if (!p.iconUrl) p.iconUrl = '';
    if (!p.lockedIconUrl) p.lockedIconUrl = '';
    p.categoryId = p.categoryId || '';
    p.tagIds = p.tagIds || [];
    if (p.source === 'event') {
      p.taskId = ''; p.taskGroupId = '';
      p.tiers = (p.tiers || []).map(function (r, i) { return { tier: i + 1, threshold: r.threshold, name: r.name || {}, prizeId: r.prizeId || '', grantMode: r.grantMode || 'claim' }; });
    } else {
      p.eventTypeId = ''; p.valueMin = null; p.valueMax = null; p.metric = 'count';
      if (p.source !== 'task') { p.taskId = ''; p.taskGroupId = ''; }
      p.taskId = p.taskId || ''; p.taskGroupId = p.taskGroupId || '';
      p.tiers = [{ tier: 1, threshold: 1, name: {}, prizeId: p.unlockPrizeId || '', grantMode: p.unlockGrantMode || 'claim' }];
    }
    delete p.unlockPrizeId;
    delete p.unlockGrantMode;
    delete p.preview;
    delete p.tiersHint;
    return p;
  },
  afterSave: function () { forgetRefs('achievements'); }
};

OVERVIEW_CATALOG.push('achievements');

var PAC_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadAchCard(id) {
  PAC_CARD.loading = true;
  return api('GET', '/users/' + encodeURIComponent(id) + '/achievements' + qs({ lang: lang() })).then(function (res) {
    PAC_CARD.loading = false;
    PAC_CARD.at = Date.now();
    PAC_CARD.data = res;
    PAC_CARD.err = null;
    var old = document.querySelector('[data-player-achievements]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    PAC_CARD.loading = false;
    PAC_CARD.at = Date.now();
    PAC_CARD.err = err;
    schedule();
  });
}

function reloadAch(id) { PAC_CARD.at = 0; loadAchCard(id); }

function achStateTag(it) {
  var color = it.state === 'completed' ? 'success' : it.state === 'unlocked' ? 'primary' : 'secondary';
  return tag(t('pac.state.' + it.state), color);
}

function achByText(it) {
  if (!it.unlocked || !it.unlockedBy) return '';
  var by = t('pac.by.' + it.unlockedBy, { name: it.unlockedRef || '' });
  return it.unlockedAt ? by + (lang() === 'fa' ? '، ' : ' · ') + fmtDate(it.unlockedAt, true) : by;
}

function achLine(userId, it, write) {
  var tierText = it.tiered || it.maxTier > 1 ? t('pac.tier', { n: fmtNum(it.tier || 0), max: fmtNum(it.maxTier || 1) }) : '';
  var pct = Math.round((Number(it.progress) || 0) * 100);
  var progress = it.source === 'event' && it.next ? t('pac.progress', { n: numLocal(it.value), total: numLocal(it.next.threshold) }) : '';
  var line = h('div', { class: 'flex flex-wrap items-center gap-2 py-2 text-sm', 'data-pac': it.key, 'data-state': it.state },
    achBadge(it, !!it.unlocked),
    h('div', { class: 'min-w-0 flex flex-col' },
      h('span', { class: 'font-medium truncate', text: it.name || it.key }),
      achByText(it) ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', 'data-pac-by': '', text: achByText(it) }) : null),
    achStateTag(it),
    rarityTag(it.rarity),
    it.hidden ? tag(t('ac.secret'), 'secondary') : null,
    it.isActive === false ? tag(t('inv.inactive'), 'secondary') : null,
    tierText ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums whitespace-nowrap', 'data-pac-tier': '', text: tierText }) : null,
    it.source === 'event' && it.next ? h('span', { class: 'eco-tk-meter', 'aria-hidden': 'true' }, h('span', { style: 'inline-size:' + pct + '%' })) : null,
    progress ? h('span', { class: 'tabular-nums text-secondary-600 dark:text-secondary-300 whitespace-nowrap', 'data-pac-progress': '', text: progress }) : null,
    h('span', { class: 'flex-1' }));
  if (write) {
    var acts = h('div', { class: 'flex flex-wrap items-center gap-1' });
    if ((it.claimable || []).length) {
      var cb = button({ variant: 'link', size: 'sm', label: t('pac.claim'), onclick: function () {
        api('POST', '/users/' + encodeURIComponent(userId) + '/achievements/' + encodeURIComponent(it.id) + '/claim', {}).then(function () { toast('success', t('pac.claim.done')); reloadAch(userId); }, function (err) { toast('error', errorText(err)); });
      } });
      cb.setAttribute('data-pac-claim', it.key);
      acts.appendChild(cb);
    }
    if (it.unlocked) {
      var rb = button({ variant: 'link', size: 'sm', label: t('pac.revoke'), onclick: function () {
        confirmModal({ title: t('pac.revoke.title'), text: t('pac.revoke.text', { name: it.name || it.key }), ok: t('pac.revoke'), danger: true }).then(function (yes) {
          if (!yes) return;
          api('DELETE', '/users/' + encodeURIComponent(userId) + '/achievements/' + encodeURIComponent(it.id)).then(function () { toast('success', t('pac.revoke.done')); reloadAch(userId); }, function (err) { toast('error', errorText(err)); });
        });
      } });
      rb.setAttribute('data-pac-revoke', it.key);
      acts.appendChild(rb);
    }
    if (acts.childNodes.length) line.appendChild(acts);
  }
  return line;
}

function playerAchievementsCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-achievements]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (PAC_CARD.id !== id) { PAC_CARD.id = id; PAC_CARD.data = null; PAC_CARD.err = null; PAC_CARD.at = 0; }
  if (!PAC_CARD.loading && Date.now() - PAC_CARD.at > 20000) loadAchCard(id);
  var p = PAC_CARD.data;
  var anchor = grid.querySelector('[data-player-items]') || grid.querySelector('[data-player-tasks]') || grid.querySelector('[data-player-activity]') || grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify(p.items) : PAC_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-achievements': '', 'data-key': key, 'data-economy': 'player-achievements', 'data-l10n-skip': '', 'aria-labelledby': 'player-achievements-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-achievements-title', class: 'font-medium flex items-center gap-2' }, icon('achievements', 'h-4 w-4'), h('span', { text: t('pac.card') }));
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(PAC_CARD.err ? notice('error', errorText(PAC_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { reloadAch(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var write = aclCan('HIRO_ECONOMY', 1);
  var items = p.items || [];
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' }, title,
    h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300 tabular-nums', 'data-pac-counts': '', text: t('pac.counts', { n: fmtNum(p.unlocked || 0), total: fmtNum(p.total || 0), p: fmtNum(p.points || 0) }) }),
    p.ready ? tag(t('pac.waiting', { n: fmtNum(p.ready) }), 'warning') : null,
    h('div', { class: 'flex flex-wrap items-center gap-3' },
      aclCan('ACCOUNT', 0) ? h('a', { href: hashFor('activity', null, { userId: id, kind: 'achievement' }), class: 'eco-link text-sm', 'data-pac-log': '', text: t('pac.log') }) : null,
      h('a', { href: hashFor('achievements'), class: 'eco-link text-sm', text: t('pac.all') }))));
  var sorted = items.slice().sort(function (a, b) { return (b.unlocked ? 1 : 0) - (a.unlocked ? 1 : 0); });
  var list = h('div', { class: 'flex flex-col eco-tk-list', 'data-pac-list': '' });
  sorted.forEach(function (it) { list.appendChild(achLine(id, it, write)); });
  if (!items.length) list.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('pac.none') }));
  body.appendChild(list);
  if (write && items.length) {
    var pick = '';
    var tierNum = null;
    var picker = refPicker({ resource: 'achievements', value: '', label: t('pac.give.pick'), placeholder: t('pac.give.pick'), onchange: function (v) { pick = v || ''; } });
    var tierIn = numberInput(null, { int: true, label: t('pac.give.tier'), oninput: function (v) { tierNum = v === '' ? null : parseNumber(v, true); } });
    tierIn.setAttribute('data-pac-tier-input', '');
    var give = button({ variant: 'primary', size: 'sm', icon: 'plus', label: t('pac.give'), onclick: function () {
      if (!pick) { toast('error', t('pac.give.pick')); return; }
      var bodyIn = { achievementId: pick };
      if (tierNum) bodyIn.tier = tierNum;
      api('POST', '/users/' + encodeURIComponent(id) + '/achievements', bodyIn).then(function () { toast('success', t('pac.give.done')); reloadAch(id); }, function (err) { toast('error', errorText(err)); });
    } });
    give.setAttribute('data-pac-give', '');
    body.appendChild(h('div', { class: 'mt-3 border-t border-secondary-200 dark:border-secondary-700 pt-4 flex flex-wrap items-end gap-2', 'data-pac-give-row': '' },
      h('div', { class: 'min-w-0 flex-1 eco-inv-pick' }, picker),
      h('div', { style: 'inline-size:6rem' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('pac.give.tier') + ' (' + t('optional') + ')' }), tierIn),
      give));
  }
  place();
}

var BRAND_KEY = 'nakama-brand';
var BRAND_SHADES = { 50: 0.9, 100: 0.8, 200: 0.6, 300: 0.4, 400: 0.1, 500: 0, 600: -0.1, 700: -0.25, 800: -0.58, 900: -0.7, 950: -0.8 };
var BRAND_DEFAULT_SHADES = { '244,244,253': 50, '233,232,252': 100, '209,205,248': 200, '181,175,245': 300, '130,117,240': 400, '107,89,238': 500, '102,84,226': 600, '93,77,206': 700, '69,58,154': 800, '59,49,130': 900, '48,40,106': 950 };
var BRAND_DEFAULT_ACCENT = '#6b59ee';
var BRAND_RGB_RE = /rgb\(\s*(\d{1,3})[\s,]+(\d{1,3})[\s,]+(\d{1,3})\s*\)/g;
var BRAND_HEX_RE = /^#?([0-9a-f]{3}|[0-9a-f]{6})$/i;
var BRAND = { data: null, loaded: false, promise: null, applied: '', titleName: '' };
var recolorObserver = null;
var nativeTitle = Object.getOwnPropertyDescriptor(Document.prototype, 'title');

function brandHex(v) {
  var m = BRAND_HEX_RE.exec(String(v || '').trim());
  if (!m) return '';
  var x = m[1].toLowerCase();
  if (x.length === 3) x = x[0] + x[0] + x[1] + x[1] + x[2] + x[2];
  return '#' + x;
}

function hexRgb(hex) {
  var x = brandHex(hex).slice(1);
  return [parseInt(x.slice(0, 2), 16), parseInt(x.slice(2, 4), 16), parseInt(x.slice(4, 6), 16)];
}

function rgbHex(rgb) {
  return '#' + rgb.map(function (c) { return ('0' + Math.round(c).toString(16)).slice(-2); }).join('');
}

function brandPalette(hex) {
  var base = hexRgb(hex);
  var out = {};
  Object.keys(BRAND_SHADES).forEach(function (k) {
    var a = BRAND_SHADES[k];
    out[k] = base.map(function (c) { return Math.round(a >= 0 ? c + (255 - c) * a : c * (1 + a)); });
  });
  return out;
}

function luminance(rgb) {
  var v = rgb.map(function (c) {
    c = c / 255;
    return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
  });
  return 0.2126 * v[0] + 0.7152 * v[1] + 0.0722 * v[2];
}

function contrastRatio(a, b) {
  var la = luminance(a);
  var lb = luminance(b);
  return (Math.max(la, lb) + 0.05) / (Math.min(la, lb) + 0.05);
}

function brandContrast(hex) {
  var p = brandPalette(hex);
  return {
    light: contrastRatio(p[500], [255, 255, 255]),
    dark: contrastRatio(p[400], [17, 24, 39])
  };
}

function contrastLevel(ratio) {
  if (ratio >= 4.5) return 'good';
  if (ratio >= 3) return 'large';
  return 'poor';
}

function brandText(map) {
  if (!map || typeof map !== 'object') return '';
  var l = lang();
  if (map[l]) return map[l];
  if (map[LANGS.def]) return map[LANGS.def];
  var keys = Object.keys(map).sort();
  for (var i = 0; i < keys.length; i++) if (map[keys[i]]) return map[keys[i]];
  return '';
}

function brandName() {
  return BRAND.data ? brandText(BRAND.data.name) : '';
}

function brandCustom(b) {
  b = b || BRAND.data;
  return !!(b && (b.logoUrl || b.logoDarkUrl || b.markUrl || b.accent || b.shortName || brandText(b.name) || brandText(b.welcome)));
}

function brandSet(data) {
  BRAND.data = data && typeof data === 'object' ? data : null;
  BRAND.loaded = true;
  try {
    if (brandCustom(BRAND.data)) localStorage.setItem(BRAND_KEY, JSON.stringify(BRAND.data));
    else localStorage.removeItem(BRAND_KEY);
  } catch (e) {}
  BRAND.applied = '';
  applyBrand();
}

function brandLoad(force) {
  if (BRAND.promise && !force) return BRAND.promise;
  BRAND.promise = fetch(BASE_BASE + '/brand/public', { cache: 'no-cache', credentials: 'omit', headers: { Accept: 'application/json' } }).then(function (res) {
    if (!res.ok) throw new Error('brand ' + res.status);
    return res.json();
  }).then(function (data) {
    brandSet(data);
    return BRAND.data;
  }, function () {
    BRAND.promise = null;
    return BRAND.data;
  });
  return BRAND.promise;
}

var SP_INK_LIGHT_BG = [255, 255, 255];
var SP_INK_DARK_BG = [36, 37, 41];

var SP_INK_MIN = 5.5;

function pickShade(p, shades, bg, min) {
  for (var i = 0; i < shades.length; i++) if (contrastRatio(p[shades[i]], bg) >= (min || 4.5)) return shades[i];
  return shades[shades.length - 1];
}

function spatialAccent(hex) {
  var p = brandPalette(hex);
  var a = pickShade(p, [500, 600, 700, 800, 900], [255, 255, 255]);
  var ink = pickShade(p, [600, 700, 800, 900, 950], SP_INK_LIGHT_BG, SP_INK_MIN);
  var inkDark = pickShade(p, [300, 200, 100, 50], SP_INK_DARK_BG, SP_INK_MIN);
  return {
    accent: p[a], ink: p[ink], inkDark: p[inkDark],
    shades: { accent: +a, ink: +ink, inkDark: +inkDark },
    ratios: { onAccent: contrastRatio(p[a], [255, 255, 255]), ink: contrastRatio(p[ink], SP_INK_LIGHT_BG), inkDark: contrastRatio(p[inkDark], SP_INK_DARK_BG) }
  };
}

function brandCss(hex) {
  var p = brandPalette(hex);
  var sp = spatialAccent(hex);
  return 'html:root{' + Object.keys(p).map(function (k) { return '--primary-color-' + k + ':' + p[k].join(' ') + ';'; }).join('') +
    '--pl-accent:rgb(' + sp.accent.join(' ') + ');--pl-accent-ink:rgb(' + sp.ink.join(' ') + ');--pl-accent-ink-dark:rgb(' + sp.inkDark.join(' ') + ');}';
}

function recolorEl(el) {
  if (!el || el.nodeType !== 1) return;
  var s = el.getAttribute('style');
  if (!s || s.indexOf('rgb(') === -1) return;
  var out = s.replace(BRAND_RGB_RE, function (m, r, g, b) {
    var n = BRAND_DEFAULT_SHADES[+r + ',' + +g + ',' + +b];
    return n ? 'rgb(var(--primary-color-' + n + '))' : m;
  });
  if (out !== s) el.setAttribute('style', out);
}

function recolorTree(node) {
  if (!node || node.nodeType !== 1) return;
  recolorEl(node);
  var list = node.querySelectorAll('[style*="rgb("]');
  for (var i = 0; i < list.length; i++) recolorEl(list[i]);
}

function startRecolor() {
  if (recolorObserver) return;
  recolorTree(document.documentElement);
  recolorObserver = new MutationObserver(function (records) {
    for (var i = 0; i < records.length; i++) {
      var r = records[i];
      if (r.type === 'attributes') recolorEl(r.target);
      else for (var j = 0; j < r.addedNodes.length; j++) recolorTree(r.addedNodes[j]);
    }
  });
  recolorObserver.observe(document.documentElement, { subtree: true, childList: true, attributes: true, attributeFilter: ['style'] });
}

function applyAccent(hex) {
  var style = document.getElementById('brand-accent');
  hex = brandHex(hex);
  if (!hex) {
    if (style) style.remove();
    return;
  }
  var css = brandCss(hex);
  if (!style) {
    style = document.createElement('style');
    style.id = 'brand-accent';
  }
  if (style.textContent !== css) style.textContent = css;
  if (style.parentNode !== document.head) document.head.appendChild(style);
  startRecolor();
}

function applyFavicon(url) {
  var links = document.querySelectorAll('link[rel~="icon"]');
  if (!links.length && url) {
    var link = document.createElement('link');
    link.rel = 'icon';
    document.head.appendChild(link);
    links = [link];
  }
  for (var i = 0; i < links.length; i++) {
    var l = links[i];
    if (!l.hasAttribute('data-brand-default')) l.setAttribute('data-brand-default', l.getAttribute('href') || '/favicon.ico');
    var next = url || l.getAttribute('data-brand-default');
    if (l.getAttribute('href') !== next) l.setAttribute('href', next);
  }
}

function brandTitle(v) {
  v = String(v == null ? '' : v);
  if (BRAND.titleName) v = v.split(BRAND.titleName).join('Nakama');
  var name = brandName();
  BRAND.titleName = name;
  if (!name) return v;
  var parts = v.split(' | ');
  var last = parts.length - 1;
  parts[last] = parts[last].replace(/\bNakama\b/, name);
  return parts.join(' | ');
}

function hookTitle() {
  if (!nativeTitle || !nativeTitle.set) return;
  Object.defineProperty(document, 'title', {
    configurable: true,
    get: function () { return nativeTitle.get.call(document); },
    set: function (v) { nativeTitle.set.call(document, brandTitle(sectionPageTitle(v))); }
  });
}

function refreshTitle() {
  var cur = nativeTitle.get.call(document);
  var next = brandTitle(cur);
  if (next !== cur) nativeTitle.set.call(document, next);
}

function brandLogoEl(b, sizeCls, compact) {
  var wrap = h('span', { class: 'brand-logo ' + sizeCls, 'data-brand-logo': '', 'data-l10n-skip': '' });
  var name = brandText(b.name);
  var alt = name || b.shortName || '';
  if (b.logoUrl && !compact) {
    wrap.appendChild(h('img', { src: b.logoUrl, alt: alt, class: 'brand-logo-img' + (b.logoDarkUrl ? ' brand-only-light' : '') }));
    if (b.logoDarkUrl) wrap.appendChild(h('img', { src: b.logoDarkUrl, alt: alt, class: 'brand-logo-img brand-only-dark' }));
    return wrap;
  }
  var mark = b.markUrl;
  if (mark) wrap.appendChild(h('img', { src: mark, alt: compact ? alt : '', class: 'brand-logo-mark' }));
  var label = compact ? (b.shortName || name).slice(0, 2) : (name || b.shortName);
  if (!mark || !compact) wrap.appendChild(h('span', { class: 'brand-logo-name', text: label || 'Nakama' }));
  return wrap;
}

function placeLogo(anchor, b, compact) {
  var natives = [];
  for (var i = 0; i < anchor.children.length; i++) if (!anchor.children[i].hasAttribute('data-brand-logo')) natives.push(anchor.children[i]);
  var ours = anchor.querySelector(':scope > [data-brand-logo]');
  var show = b && (b.logoUrl || b.markUrl || brandText(b.name) || b.shortName);
  if (!show) {
    if (ours) ours.remove();
    natives.forEach(function (n) { if (n.hasAttribute('data-brand-hidden')) { n.style.display = ''; n.removeAttribute('data-brand-hidden'); } });
    return;
  }
  var img = anchor.querySelector('img:not([data-brand-logo] img)');
  var size = img && /\bh-(\d+)\b/.exec(img.className) ? 'brand-h-' + /\bh-(\d+)\b/.exec(img.className)[1] : 'brand-h-7';
  var key = lang() + '|' + compact + '|' + size + '|' + JSON.stringify([b.name, b.shortName, b.logoUrl, b.logoDarkUrl, b.markUrl]);
  if (!ours || ours.getAttribute('data-key') !== key) {
    var built = brandLogoEl(b, size, compact);
    built.setAttribute('data-key', key);
    if (ours) ours.replaceWith(built);
    else anchor.insertBefore(built, anchor.firstChild);
  }
  natives.forEach(function (n) {
    if (n.style.display !== 'none') n.style.display = 'none';
    n.setAttribute('data-brand-hidden', '');
  });
}

function applyLogos(b) {
  var drawers = document.querySelectorAll('.x-drawer a[href="#/"]');
  for (var i = 0; i < drawers.length; i++) {
    var a = drawers[i];
    if (!a.querySelector('img') || a.closest('nav')) continue;
    var drawer = a.closest('.x-drawer');
    var compact = !!drawer && drawer.getBoundingClientRect().width > 0 && drawer.getBoundingClientRect().width < 100;
    placeLogo(a, b, compact);
  }
  var login = document.querySelectorAll('a[href="/"]');
  for (var j = 0; j < login.length; j++) if (login[j].querySelector('img')) placeLogo(login[j], b, false);
}

function applyLogin(b) {
  var input = document.querySelector('form input[name="username"]');
  var heads = document.querySelectorAll('h1[data-brand-hidden]');
  var native = null;
  if (input) {
    var box = input.closest('form');
    for (var k = 0; k < 4 && box && !native; k++) {
      box = box.parentElement;
      native = box && box.querySelector('h1:not([data-brand-signin])');
    }
  }
  var name = b ? brandText(b.name) || b.shortName : '';
  var welcome = b ? brandText(b.welcome) : '';
  var ours = document.querySelector('[data-brand-signin]');
  if (!native || (!name && !welcome)) {
    if (ours) ours.remove();
    for (var i = 0; i < heads.length; i++) { heads[i].style.display = ''; heads[i].removeAttribute('data-brand-hidden'); }
    return;
  }
  var key = lang() + '|' + name + '|' + welcome;
  if (!ours || ours.getAttribute('data-key') !== key || ours.previousElementSibling !== native) {
    var built = h('div', { 'data-brand-signin': '', 'data-l10n-skip': '', 'data-key': key, class: 'brand-signin', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('h1', { class: native.className, text: name ? t('brand.signIn', { name: name }) : native.textContent }),
      welcome ? h('p', { class: 'brand-welcome', text: welcome }) : null);
    if (ours) ours.remove();
    native.after(built);
  }
  if (native.style.display !== 'none') native.style.display = 'none';
  native.setAttribute('data-brand-hidden', '');
}

function applyBrand() {
  var b = BRAND.data;
  var key = lang() + '|' + JSON.stringify(b);
  if (BRAND.applied !== key) {
    BRAND.applied = key;
    applyAccent(b && b.accent);
    applyFavicon(b && b.markUrl ? b.markUrl : '');
    refreshTitle();
  }
  applyLogos(b);
  applyLogin(b);
}

(function () {
  try {
    var cached = localStorage.getItem(BRAND_KEY);
    if (cached) BRAND.data = JSON.parse(cached);
  } catch (e) { BRAND.data = null; }
  hookTitle();
  if (BRAND.data) {
    applyAccent(BRAND.data.accent);
    applyFavicon(BRAND.data.markUrl || '');
    refreshTitle();
  }
  brandLoad();
})();

var BRAND_SWATCHES = ['#6b59ee', '#2563eb', '#0891b2', '#059669', '#d97706', '#dc2626', '#db2777', '#475569'];

function contrastLine(labelKey, ratio) {
  var level = contrastLevel(ratio);
  var color = level === 'good' ? 'success' : level === 'large' ? 'warning' : 'error';
  return h('div', { class: 'flex items-center justify-between gap-3 text-sm', 'data-contrast': labelKey },
    h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: t(labelKey) }),
    h('span', { class: 'flex items-center gap-2' },
      h('span', { class: 'tabular-nums', dir: 'ltr', text: fmtNum(ratio, { minimumFractionDigits: 1, maximumFractionDigits: 1 }) + ':' + fmtNum(1) }),
      tag(t('brand.contrast.' + level), color)));
}

function brandPreview(state) {
  var box = h('div', { class: 'brand-preview', 'aria-hidden': 'true' });
  function paint() {
    clear(box);
    var hex = brandHex(state.accent) || BRAND_DEFAULT_ACCENT;
    var p = brandPalette(hex);
    ['light', 'dark'].forEach(function (mode) {
      var dark = mode === 'dark';
      var logoUrl = dark ? state.logoDarkUrl || state.logoUrl : state.logoUrl;
      var name = brandText(state.name) || state.shortName || 'Nakama';
      var head = h('div', { class: 'brand-preview-head' });
      if (logoUrl) head.appendChild(h('img', { src: logoUrl, alt: '', class: 'brand-preview-logo' }));
      else {
        if (state.markUrl) head.appendChild(h('img', { src: state.markUrl, alt: '', class: 'brand-preview-mark' }));
        head.appendChild(h('span', { class: 'brand-preview-name', text: name }));
      }
      var sp = spatialAccent(hex);
      var accent = 'rgb(' + sp.accent.join(',') + ')';
      var link = 'rgb(' + (dark ? sp.inkDark : sp.ink).join(',') + ')';
      box.appendChild(h('div', { class: 'brand-preview-pane brand-preview-' + mode },
        h('div', { class: 'brand-preview-mode', text: t(dark ? 'brand.preview.dark' : 'brand.preview.light') }),
        head,
        h('div', { class: 'brand-preview-nav', style: 'border-inline-start-color:' + link + ';color:' + link + ';background-color:rgba(' + sp.accent.join(',') + ',' + (dark ? '.17' : '.11') + ')' }, t('brand.preview.nav')),
        h('div', { class: 'brand-preview-row' },
          h('span', { class: 'brand-preview-btn', style: 'background-color:' + accent, text: t('save') }),
          h('span', { class: 'brand-preview-link', style: 'color:' + link, text: t('brand.preview.link') }))));
    });
  }
  paint();
  box.paint = paint;
  return box;
}

function brandImage(value, extra, onchange) {
  var el = imageInput(value, onchange);
  el.classList.add('brand-img');
  if (extra) el.classList.add(extra);
  return el;
}

function brandEditor(data) {
  data = data || {};
  var state = {
    name: trValue(data.name), shortName: data.shortName || '', logoUrl: data.logoUrl || '', logoDarkUrl: data.logoDarkUrl || '',
    markUrl: data.markUrl || '', accent: brandHex(data.accent), welcome: trValue(data.welcome)
  };
  var initial = null;
  var preview = brandPreview(state);
  var onchange = function () { preview.paint(); paintContrast(); };

  var nameCtl = translatedInput(state.name, { name: 'name', label: t('brand.f.name'), max: 64, oninput: function (v) { state.name = v; nameField.setError(''); nameCtl.markLang(null); onchange(); } });
  var nameField = field(t('brand.f.name'), nameCtl, t('brand.f.name.hint'), { optional: true });
  var shortCtl = textInput(state.shortName, { label: t('brand.f.shortName'), oninput: function (v) { state.shortName = v; shortField.setError(''); onchange(); } });
  var shortField = field(t('brand.f.shortName'), shortCtl, t('brand.f.shortName.hint'), { optional: true });
  var logoField = field(t('brand.f.logo'), brandImage(state.logoUrl, '', function (v) { state.logoUrl = v || ''; logoField.setError(''); onchange(); }), t('brand.f.logo.hint'), { optional: true });
  var darkField = field(t('brand.f.logoDark'), brandImage(state.logoDarkUrl, 'brand-img--dark', function (v) { state.logoDarkUrl = v || ''; darkField.setError(''); onchange(); }), t('brand.f.logoDark.hint'), { optional: true });
  var markField = field(t('brand.f.mark'), brandImage(state.markUrl, '', function (v) { state.markUrl = v || ''; markField.setError(''); onchange(); }), t('brand.f.mark.hint'), { optional: true });

  var picker = h('input', { type: 'color', class: 'brand-color-input', 'aria-label': t('brand.f.accent.pick'), value: state.accent || BRAND_DEFAULT_ACCENT });
  var hexInput = textInput(state.accent, { ltr: true, mono: true, label: t('brand.f.accent'), placeholder: BRAND_DEFAULT_ACCENT });
  hexInput.classList.add('text-sm');
  hexInput.setAttribute('maxlength', '7');
  var swatches = h('div', { class: 'brand-swatches', role: 'group', 'aria-label': t('brand.f.accent.presets') });
  BRAND_SWATCHES.forEach(function (c) {
    swatches.appendChild(h('button', { type: 'button', class: 'brand-swatch', style: 'background-color:' + c, title: c, 'aria-label': c, 'data-color': c, onclick: function () { setAccent(c); } }));
  });
  var resetBtn = button({ variant: 'link', size: 'sm', label: t('brand.f.accent.reset'), onclick: function () { setAccent(''); } });
  var contrastBox = h('div', { class: 'brand-contrast', 'aria-live': 'polite' });
  function setAccent(v, fromText) {
    state.accent = v ? brandHex(v) || v : '';
    if (!fromText) hexInput.value = state.accent;
    if (brandHex(state.accent)) picker.value = brandHex(state.accent);
    accentField.setError('');
    Array.prototype.forEach.call(swatches.children, function (s) { s.setAttribute('aria-pressed', s.getAttribute('data-color') === brandHex(state.accent) ? 'true' : 'false'); });
    onchange();
  }
  picker.addEventListener('input', function () { setAccent(picker.value); });
  hexInput.addEventListener('input', function () {
    var v = hexInput.value.trim();
    if (v && v[0] !== '#') v = '#' + v;
    setAccent(v, true);
  });
  function paintContrast() {
    clear(contrastBox);
    var hex = brandHex(state.accent);
    if (state.accent && !hex) return;
    var c = brandContrast(hex || BRAND_DEFAULT_ACCENT);
    contrastBox.appendChild(contrastLine('brand.contrast.onWhite', c.light));
    contrastBox.appendChild(contrastLine('brand.contrast.onDark', c.dark));
    if (contrastLevel(c.light) === 'poor' || contrastLevel(c.dark) === 'poor') contrastBox.appendChild(h('p', { class: 'text-xs eco-text-warning mt-1', text: t('brand.contrast.tip') }));
  }
  var accentField = field(t('brand.f.accent'), h('div', { class: 'flex flex-col gap-2' },
    h('div', { class: 'flex items-center gap-2' }, picker, h('div', { class: 'brand-hex' }, hexInput), resetBtn),
    swatches), t('brand.f.accent.hint'), { optional: true });

  var welcomeCtl = translatedInput(state.welcome, { name: 'welcome', label: t('brand.f.welcome'), multiline: true, rows: 2, max: 280, counter: true, oninput: function (v) { state.welcome = v; welcomeField.setError(''); welcomeCtl.markLang(null); } });
  var welcomeField = field(t('brand.f.welcome'), welcomeCtl, t('brand.f.welcome.hint'), { optional: true });

  var fieldsEl = { name: nameField, shortName: shortField, logoUrl: logoField, logoDarkUrl: darkField, markUrl: markField, accent: accentField, welcome: welcomeField };
  setAccent(state.accent);
  var el = h('div', { class: 'brand-editor' },
    h('div', { class: 'brand-editor-main' },
      nameField, shortField,
      h('div', { class: 'eco-form-row' }, logoField, darkField),
      markField, accentField, contrastBox, welcomeField),
    h('aside', { class: 'brand-editor-aside' },
      h('div', { class: 'text-sm font-medium text-secondary-700 dark:text-secondary-200 mb-2', text: t('brand.preview') }),
      preview));

  function values() {
    return { name: trClean(state.name), shortName: state.shortName.trim(), logoUrl: (state.logoUrl || '').trim(), logoDarkUrl: (state.logoDarkUrl || '').trim(), markUrl: (state.markUrl || '').trim(), accent: brandHex(state.accent), welcome: trClean(state.welcome) };
  }
  initial = JSON.stringify(values());
  function validate() {
    var ok = true;
    var hit = trCheck(state.name, { max: 64 });
    if (hit) { nameCtl.markLang(hit.lang); nameField.setError(t('code.' + hit.code, { lang: langName(hit.lang) })); ok = false; }
    var wh = trCheck(state.welcome, { max: 280 });
    if (wh) { welcomeCtl.markLang(wh.lang); welcomeField.setError(t('code.' + wh.code, { lang: langName(wh.lang) })); ok = false; }
    if (Array.from(state.shortName.trim()).length > 24) { shortField.setError(t('brand.err.short')); ok = false; }
    if (state.accent && !brandHex(state.accent)) { accentField.setError(t('brand.err.color')); ok = false; }
    if (!ok) { var bad = el.querySelector('.eco-invalid'); if (bad) bad.scrollIntoView({ block: 'center', behavior: 'smooth' }); }
    return ok;
  }
  function showError(err) {
    var f = err && err.field && fieldsEl[err.field];
    if (!f) return false;
    if (err.field === 'name') nameCtl.markLang(trServerLang(err, state.name, { max: 64 }));
    if (err.field === 'welcome') welcomeCtl.markLang(trServerLang(err, state.welcome, { max: 280 }));
    f.setError(err.code === 'INVALID_COLOR' ? t('brand.err.color') : errorText(err));
    f.scrollIntoView({ block: 'center', behavior: 'smooth' });
    return true;
  }
  return {
    el: el, values: values, validate: validate, showError: showError,
    dirty: function () { return JSON.stringify(values()) !== initial; },
    clean: function () { initial = JSON.stringify(values()); }
  };
}

function brandFormSkeleton() {
  return h('div', { class: 'brand-editor' },
    h('div', { class: 'brand-editor-main' }, skeletonField('h-16'), skeletonField(),
      h('div', { class: 'eco-form-row' }, skeletonField('h-16'), skeletonField('h-16')), skeletonField('h-16'), skeletonField(), skeletonField('h-16')),
    h('aside', { class: 'brand-editor-aside' }, bar('5rem', 'h-3 mb-2'), bar('100%', 'h-36 mb-3'), bar('100%', 'h-36')));
}

function saveBrand(editor) {
  if (!editor.validate()) return Promise.resolve(false);
  return baseApi('PUT', '/brand', editor.values()).then(function (res) {
    brandSet(res);
    editor.clean();
    return true;
  }, function (err) {
    if (!editor.showError(err)) toast('error', errorText(err));
    return false;
  });
}

function brandPage(container) {
  var frame = tabFrame(t('desc.brand'));
  clear(container).appendChild(frame.el);
  var body = h('div', { class: 'p-4 md:p-6' }, brandFormSkeleton());
  frame.content.appendChild(card(body));
  function load() {
    Promise.all([baseApi('GET', '/brand'), freshLanguages()]).then(function (all) {
      if (!container.isConnected) return;
      var editor = brandEditor(all[0]);
      var saveBtn = button({ variant: 'primary', size: 'sm', label: t('save'), onclick: function () {
        setBusy(saveBtn, true, t('saving'));
        saveBrand(editor).then(function (ok) { setBusy(saveBtn, false); if (ok) toast('success', t('saved')); });
      } });
      clear(body).appendChild(editor.el);
      body.appendChild(h('div', { class: 'flex justify-end pt-2' }, saveBtn));
    }, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}

var SETUP_ORDER = ['brand', 'languages', 'currency', 'payment', 'sms', 'email', 'review'];
var SETUP_ICONS = { brand: 'image', languages: 'languages', currency: 'currencies', payment: 'payments', sms: 'comments', email: 'send', review: 'check' };
var SETUP_STATUS_COLOR = { pending: 'warning', in_progress: 'primary', completed: 'success', skipped: 'secondary' };
var STEP_COLOR = { done: 'success', skipped: 'secondary', pending: 'warning' };
var QUICK_LANGS = [
  { code: 'fa', native: 'فارسی', dir: 'rtl', fa: 'فارسی', en: 'Persian' },
  { code: 'en', native: 'English', dir: 'ltr', fa: 'انگلیسی', en: 'English' },
  { code: 'ar', native: 'العربية', dir: 'rtl', fa: 'عربی', en: 'Arabic' },
  { code: 'tr', native: 'Türkçe', dir: 'ltr', fa: 'ترکی', en: 'Turkish' },
  { code: 'ru', native: 'Русский', dir: 'ltr', fa: 'روسی', en: 'Russian' },
  { code: 'de', native: 'Deutsch', dir: 'ltr', fa: 'آلمانی', en: 'German' },
  { code: 'fr', native: 'Français', dir: 'ltr', fa: 'فرانسوی', en: 'French' },
  { code: 'es', native: 'Español', dir: 'ltr', fa: 'اسپانیایی', en: 'Spanish' },
  { code: 'ur', native: 'اردو', dir: 'rtl', fa: 'اردو', en: 'Urdu' },
  { code: 'zh', native: '中文', dir: 'ltr', fa: 'چینی', en: 'Chinese' }
];
var setupClaimFor = null;
var wiz = null;

function setupStepStatus(state, k) {
  return (state && state.steps && state.steps[k]) || 'pending';
}

function quietBase(method, path, body) {
  var tk = token();
  if (!tk) return Promise.resolve(null);
  var headers = { Accept: 'application/json', Authorization: 'Bearer ' + tk };
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  return fetch(BASE_BASE + path, { method: method, headers: headers, body: body === undefined ? undefined : JSON.stringify(body) }).then(function (res) {
    return res.ok ? res.json() : null;
  }, function () { return null; });
}

function maybeClaimSetup() {
  var tk = token();
  if (!tk || wiz || setupClaimFor === tk || /^#\/login/.test(location.hash)) return;
  if (!document.querySelector('.x-drawer a[href="#/players"]')) return;
  setupClaimFor = tk;
  quietBase('POST', '/setup/claim').then(function (res) {
    if (res && res.claimed && res.setup && !wiz) openWizard(res.setup, true);
  });
}

function putSetup(body) {
  return baseApi('PUT', '/setup', body).then(function (st) {
    if (wiz) { wiz.state = st; paintBar(); }
    return st;
  });
}

function focusables(root) {
  return Array.prototype.filter.call(root.querySelectorAll('a[href],button:not([disabled]),input:not([disabled]):not([type="hidden"]):not(.hidden),select:not([disabled]),textarea:not([disabled]),summary,[tabindex]:not([tabindex="-1"])'), function (el) {
    return el.offsetParent !== null || el === document.activeElement;
  });
}

function openWizard(state, auto, step) {
  if (wiz) { goStep(step || wiz.step); return; }
  var dir = document.documentElement.getAttribute('dir') || 'ltr';
  var bar = h('ol', { class: 'setup-steps', 'aria-label': t('setup.steps') });
  var body = h('div', { class: 'setup-body' });
  var footer = h('div', { class: 'setup-footer' });
  var later = button({ variant: 'ghost', size: 'sm', label: t('setup.later'), onclick: function () { closeWizard(true); } });
  later.setAttribute('data-setup-later', '');
  var title = h('div', { class: 'setup-title' },
    h('div', { class: 'setup-title-mark' }, icon('flask', 'h-5 w-5')),
    h('div', { class: 'min-w-0' },
      h('h1', { id: 'setup-title', class: 'text-lg font-semibold truncate', text: t('setup.title') }),
      h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 truncate', text: t('setup.subtitle') })));
  var el = h('div', { class: 'setup-overlay', 'data-economy': 'setup', 'data-l10n-skip': '', dir: dir, role: 'dialog', 'aria-modal': 'true', 'aria-labelledby': 'setup-title' },
    h('header', { class: 'setup-header' }, title, later),
    h('nav', { class: 'setup-nav', 'aria-label': t('setup.steps') }, bar),
    h('main', { class: 'setup-main' }, h('div', { class: 'setup-inner' }, body)),
    footer);
  wiz = { el: el, state: state, step: null, ctl: null, bar: bar, body: body, footer: footer, prevFocus: document.activeElement, auto: !!auto };
  document.addEventListener('keydown', trapKeys, true);
  document.addEventListener('keydown', wizardEsc);
  document.documentElement.classList.add('setup-open');
  document.body.appendChild(el);
  if (!auto && state && (state.status === 'pending' || state.status === 'skipped')) putSetup({ status: 'in_progress' }).then(null, function () {});
  goStep(step || (state && state.nextStep) || 'brand', true);
}

function trapKeys(e) {
  if (e.key !== 'Tab' || !wiz || document.querySelector('[data-economy="dialog"]')) return;
  trapTab(e, wiz.el);
}

function trapTab(e, root) {
  var list = focusables(root).filter(function (el) { return el.getAttribute('tabindex') !== '-1'; });
  if (!list.length) return;
  var first = list[0];
  var last = list[list.length - 1];
  var active = document.activeElement;
  if (active && root.contains(active) && list.indexOf(active) === -1) {
    var after = list.some(function (el) { return active.compareDocumentPosition(el) & Node.DOCUMENT_POSITION_FOLLOWING; });
    var before = list.some(function (el) { return active.compareDocumentPosition(el) & Node.DOCUMENT_POSITION_PRECEDING; });
    if (!e.shiftKey && !after) { e.preventDefault(); first.focus(); }
    else if (e.shiftKey && !before) { e.preventDefault(); last.focus(); }
    return;
  }
  if (e.shiftKey && (document.activeElement === first || !root.contains(document.activeElement))) { e.preventDefault(); last.focus(); }
  else if (!e.shiftKey && (document.activeElement === last || !root.contains(document.activeElement))) { e.preventDefault(); first.focus(); }
}

function wizardEsc(e) {
  if (e.key !== 'Escape' || !wiz || e.defaultPrevented) return;
  if (document.querySelector('[data-economy="dialog"]')) return;
  e.preventDefault();
  confirmModal({ title: t('setup.later.title'), text: t('setup.later.text'), ok: t('setup.later'), cancel: t('setup.stay'), icon: 'alert' }).then(function (yes) {
    if (yes) closeWizard(true);
  });
}

function closeWizard(later, done) {
  if (!wiz) return;
  var w = wiz;
  var go = function () {
    w.el.remove();
    document.removeEventListener('keydown', wizardEsc);
    document.removeEventListener('keydown', trapKeys, true);
    document.documentElement.classList.remove('setup-open');
    wiz = null;
    if (w.prevFocus && w.prevFocus.isConnected) { try { w.prevFocus.focus(); } catch (e) {} }
    psRefresh('setup');
  };
  if (done) { go(); return; }
  if (later && (!w.ctl || !w.ctl.dirty || !w.ctl.dirty())) {
    putSetup({ status: 'skipped' }).then(null, function () {});
    toast('success', t('setup.later.toast'));
    go();
    return;
  }
  confirmModal({ title: t('setup.unsaved.title'), text: t('setup.unsaved.text'), ok: t('setup.unsaved.ok'), danger: true }).then(function (yes) {
    if (!yes) return;
    putSetup({ status: 'skipped' }).then(null, function () {});
    toast('success', t('setup.later.toast'));
    go();
  });
}

function paintBar() {
  if (!wiz) return;
  clear(wiz.bar);
  SETUP_ORDER.forEach(function (k, i) {
    var st = k === 'review' ? (wiz.state && wiz.state.status === 'completed' ? 'done' : 'pending') : setupStepStatus(wiz.state, k);
    var on = wiz.step === k;
    var mark = st === 'done' ? icon('check', 'h-3.5 w-3.5') : st === 'skipped' ? h('span', { text: '–' }) : h('span', { text: localDigits(i + 1) });
    var b = h('button', { type: 'button', class: 'setup-step setup-step--' + st + (on ? ' setup-step--on' : ''), 'data-step': k, 'aria-current': on ? 'step' : null,
      'aria-label': localDigits(i + 1) + '. ' + t('setup.step.' + k) + ' (' + t('setup.st.' + st) + ')',
      onclick: function () { goStep(k); } },
      h('span', { class: 'setup-step-num', 'aria-hidden': 'true' }, mark),
      h('span', { class: 'setup-step-label', text: t('setup.step.' + k) }));
    wiz.bar.appendChild(h('li', {}, b));
  });
  var cur = wiz.bar.querySelector('.setup-step--on');
  if (cur && cur.scrollIntoView) { try { cur.scrollIntoView({ block: 'nearest', inline: 'center' }); } catch (e) {} }
}

function goStep(k, force) {
  if (!wiz) return;
  if (!force && wiz.step === k) return;
  var leave = !force && wiz.ctl && wiz.ctl.dirty && wiz.ctl.dirty() ? confirmModal({ title: t('setup.unsaved.title'), text: t('setup.unsaved.text'), ok: t('setup.unsaved.ok'), danger: true }) : Promise.resolve(true);
  leave.then(function (yes) {
    if (!yes || !wiz) return;
    wiz.step = k;
    paintBar();
    renderStep(k);
  });
}

function nextOf(k) {
  var i = SETUP_ORDER.indexOf(k);
  return SETUP_ORDER[Math.min(i + 1, SETUP_ORDER.length - 1)];
}

function prevOf(k) {
  var i = SETUP_ORDER.indexOf(k);
  return SETUP_ORDER[Math.max(i - 1, 0)];
}

function renderStep(k) {
  var w = wiz;
  var i = SETUP_ORDER.indexOf(k);
  var st = k === 'review' ? null : setupStepStatus(w.state, k);
  var content = h('div', { class: 'setup-content' });
  clear(w.body).appendChild(h('section', { class: 'setup-section', 'aria-labelledby': 'setup-step-title', 'data-setup-step': k },
    h('div', { class: 'setup-section-head' },
      h('div', { class: 'setup-section-icon' }, icon(SETUP_ICONS[k], 'h-5 w-5')),
      h('div', { class: 'min-w-0 flex-1' },
        h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('setup.stepOf', { n: localDigits(i + 1), total: localDigits(SETUP_ORDER.length) }) }),
        h('h2', { id: 'setup-step-title', class: 'text-xl font-semibold', tabindex: '-1', text: t('setup.step.' + k + '.title') }),
        h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 mt-1', text: t('setup.step.' + k + '.desc') })),
      st ? tag(t('setup.st.' + st), STEP_COLOR[st], 'shrink-0') : null),
    content));
  var ctl = STEP_BUILDERS[k](content, w);
  w.ctl = ctl;
  paintFooter(k, ctl);
  var hd = w.body.querySelector('#setup-step-title');
  if (hd) { try { hd.focus({ preventScroll: true }); } catch (e) { hd.focus(); } }
  var main = w.el.querySelector('.setup-main');
  if (main) main.scrollTop = 0;
}

function paintFooter(k, ctl) {
  var w = wiz;
  clear(w.footer);
  var first = k === SETUP_ORDER[0];
  var review = k === 'review';
  var back = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('setup.back'), disabled: first, onclick: function () { goStep(prevOf(k)); } });
  var skip = review ? null : button({ variant: 'muted', size: 'sm', label: t('setup.skip'), onclick: function () {
    setBusy(skip, true);
    putSetup({ steps: stepPatch(k, 'skipped') }).then(function () { goStep(nextOf(k), true); }, function (err) { setBusy(skip, false); toast('error', errorText(err)); });
  } });
  if (skip) skip.setAttribute('data-setup-skip', '');
  var primary = button({ variant: 'primary', size: 'sm', icon: review ? 'check' : null, label: review ? t('setup.finish') : t('setup.saveNext'), onclick: function () {
    setBusy(primary, true, t('saving'));
    Promise.resolve(ctl.save()).then(function (result) {
      if (!result || !wiz) { setBusy(primary, false); return; }
      if (review) {
        return putSetup({ status: 'completed' }).then(function () {
          toast('success', t('setup.done.toast'));
          closeWizard(false, true);
        });
      }
      return putSetup({ steps: stepPatch(k, result) }).then(function () { goStep(nextOf(k), true); });
    }).then(null, function (err) {
      setBusy(primary, false);
      toast('error', errorText(err));
    });
  } });
  primary.setAttribute('data-setup-next', '');
  w.footer.appendChild(h('div', { class: 'setup-footer-inner' },
    back,
    h('div', { class: 'flex items-center gap-2' }, skip, primary)));
}

function stepPatch(k, v) {
  var o = {};
  o[k] = v;
  return o;
}

function stepSkeleton(rows) {
  var box = h('div', { class: 'setup-card' });
  for (var i = 0; i < rows; i++) box.appendChild(h('div', { class: 'flex items-center gap-3 py-3 border-b border-secondary-200 dark:border-secondary-700' }, bar('2rem', 'h-8'), h('div', { class: 'flex-1' }, bar('40%', 'h-3 mb-2'), bar('25%', 'h-3')), bar('3rem', 'h-5 rounded-full')));
  return box;
}

function questionSkeleton(cards) {
  var box = h('div', {}, h('div', { class: 'setup-card mb-4' }, bar('12rem', 'h-4 mb-3'), h('div', { class: 'flex gap-2' }, bar('6rem', 'h-9'), bar('6rem', 'h-9'))));
  for (var i = 0; i < cards; i++) box.appendChild(h('div', { class: 'setup-card mb-3' }, bar('8rem', 'h-4 mb-4'), bar('100%', 'h-12 mb-4'), skeletonField(), skeletonField()));
  return box;
}

function card2(children, extra) {
  return h('div', { class: 'setup-card' + (extra ? ' ' + extra : '') }, children);
}

function failBox(container, err, retry) {
  clear(container).appendChild(errorBlock(err, retry));
}

function stepBrand(box) {
  var editor = null;
  box.appendChild(card2(brandFormSkeleton()));
  function load() {
    Promise.all([baseApi('GET', '/brand'), freshLanguages()]).then(function (all) {
      if (!box.isConnected) return;
      editor = brandEditor(all[0]);
      clear(box).appendChild(card2(editor.el));
    }, function (err) { failBox(box, err, load); });
  }
  load();
  return {
    save: function () { return editor ? saveBrand(editor).then(function (ok) { return ok ? 'done' : false; }) : false; },
    dirty: function () { return !!editor && editor.dirty(); }
  };
}

function quickLangBody(q) {
  var name = { fa: q.fa, en: q.en };
  if (LANGS.def !== 'fa' && LANGS.def !== 'en') name[LANGS.def] = q.native;
  Object.keys(name).forEach(function (k) { if (LANGS.codes.indexOf(k) === -1 && k !== q.code) delete name[k]; });
  name[q.code] = q.native;
  return { code: q.code, name: name, nativeName: q.native, direction: q.dir, isActive: true, isDefault: false, sort: LANGS.items.length + 1 };
}

function stepLanguages(box) {
  box.appendChild(stepSkeleton(3));
  var busy = false;
  function load() {
    baseApi('GET', '/languages?limit=100').then(function (res) {
      if (!box.isConnected) return;
      paint((res && res.items) || []);
    }, function (err) { failBox(box, err, load); });
  }
  function act(p, done) {
    if (busy) return;
    busy = true;
    p.then(function () {
      busy = false;
      languagesChanged();
      if (done) toast('success', done);
      load();
    }, function (err) {
      busy = false;
      toast('error', errorText(err));
      load();
    });
  }
  function paint(items) {
    var list = h('ul', { class: 'setup-list', 'aria-label': t('setup.lang.list') });
    items.forEach(function (it) {
      var sw = toggle(it.isActive, null, function (on) {
        act(baseApi('PUT', '/languages/' + encodeURIComponent(it.code), languageBody(it, { isActive: on })), t(on ? 'lang.activated' : 'lang.deactivated', { name: it.nativeName || it.code }));
      }, 'sm');
      sw.setAttribute('aria-label', t('setup.lang.active', { name: it.nativeName || it.code }));
      if (it.isDefault) { sw.classList.add('opacity-60', 'pointer-events-none'); sw.setAttribute('aria-disabled', 'true'); sw.removeAttribute('tabindex'); }
      var def = it.isDefault ? null : button({ variant: 'link', size: 'sm', label: t('setup.lang.makeDefault'), onclick: function () {
        act(baseApi('PUT', '/languages/' + encodeURIComponent(it.code), languageBody(it, { isActive: true, isDefault: true })), t('setup.lang.defaultSet', { name: it.nativeName || it.code }));
      } });
      list.appendChild(h('li', { class: 'setup-list-row', 'data-lang': it.code },
        h('div', { class: 'min-w-0 flex-1' }, languageCell(it)),
        h('div', { class: 'flex items-center gap-3 shrink-0' }, def, h('div', { class: 'flex items-center gap-2 text-xs text-secondary-600 dark:text-secondary-300' }, sw, h('span', { text: it.isActive ? t('active') : t('inactive') })))));
    });
    var have = items.map(function (x) { return x.code; });
    var chips = h('div', { class: 'setup-chips' });
    QUICK_LANGS.filter(function (q) { return have.indexOf(q.code) === -1; }).forEach(function (q) {
      chips.appendChild(h('button', { type: 'button', class: 'setup-chip', 'data-add-lang': q.code, onclick: function () {
        act(baseApi('POST', '/languages', quickLangBody(q)), t('setup.lang.added', { name: q.native }));
      } }, icon('plus', 'h-3.5 w-3.5'), h('span', { dir: q.dir, text: q.native }), h('span', { class: 'setup-chip-code', dir: 'ltr', text: q.code })));
    });
    clear(box).appendChild(card2(list));
    box.appendChild(card2([
      h('div', { class: 'text-sm font-medium mb-1', text: t('setup.lang.quick') }),
      h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: t('setup.lang.quick.hint') }),
      chips.children.length ? chips : h('p', { class: 'text-sm text-secondary-500', text: t('setup.lang.quick.none') })
    ], 'mt-4'));
  }
  load();
  return { save: function () { return 'done'; }, dirty: function () { return false; } };
}

var SETUP_CURRENCY_FIELDS = [
  { type: 'row', fields: [
    { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, keyRe: CURRENCY_KEY_RE, lockOnEdit: true },
    { name: 'type', type: 'select', label: 'f.type', required: true, options: opts('type.', ['soft', 'hard']) }
  ] },
  { name: 'name', type: 'i18n', label: 'f.name', required: true, max: 128 },
  { name: 'defaultValue', type: 'number', label: 'setup.cur.defaultValue', hint: 'setup.cur.defaultValue.hint', min: 0 },
  { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true }
];

function stepCurrency(box) {
  var form = null;
  var current = null;
  var dirty = false;
  var items = [];
  var top = h('div', {});
  box.appendChild(card2([bar('10rem', 'h-4 mb-3'), bar('100%', 'h-10 mb-5'), h('div', { class: 'eco-form-row' }, skeletonField(), skeletonField()), skeletonField('h-16'), skeletonField(), skeletonField('h-16')]));
  function build(item) {
    current = item;
    dirty = false;
    var values = clone(item) || { type: 'soft', name: {}, defaultValue: 0, sort: 0, metadata: {} };
    form = buildForm(SETUP_CURRENCY_FIELDS, values, { onchange: function () { dirty = true; } });
    dirty = false;
    return form.el;
  }
  function paint() {
    var formBox = h('div', {});
    var pick = null;
    if (items.length) {
      var options = items.map(function (c) { return { value: c.id, label: (i18nText(c.name) || c.key) + ' (' + c.key + ')' }; });
      options.push({ value: '', label: t('setup.cur.new') });
      pick = selectInput(current ? current.id : '', options, function (v) {
        var next = items.filter(function (c) { return c.id === v; })[0] || null;
        clear(formBox).appendChild(build(next));
        clear(top);
      }, { label: t('setup.cur.pick'), placeholder: false });
    }
    clear(box).appendChild(card2([
      pick ? field(t('setup.cur.pick'), pick, t('setup.cur.pick.hint')) : h('div', { class: 'mb-4' }, notice('info', t('setup.cur.none'))),
      top,
      formBox
    ]));
    formBox.appendChild(build(current));
  }
  function load() {
    api('GET', '/currencies?limit=100').then(function (res) {
      if (!box.isConnected) return;
      items = ((res && res.items) || []).slice().sort(function (a, b) { return (a.sort || 0) - (b.sort || 0) || String(a.key).localeCompare(String(b.key)); });
      current = items[0] || null;
      paint();
    }, function (err) { failBox(box, err, load); });
  }
  load();
  return {
    save: function () {
      if (!form) return false;
      clear(top);
      if (!form.validate()) return false;
      var payload = form.payload();
      var body = current ? { key: current.key, type: current.type, name: current.name, defaultValue: current.defaultValue, iconUrl: current.iconUrl, categoryId: current.categoryId || null, tagIds: current.tagIds || [], metadata: current.metadata || {}, sort: current.sort || 0 } : { metadata: {}, sort: 0 };
      Object.keys(payload).forEach(function (k) { body[k] = payload[k]; });
      if (body.iconUrl == null) body.iconUrl = '';
      return api(current ? 'PUT' : 'POST', '/currencies' + (current ? '/' + encodeURIComponent(current.id) : ''), body).then(function (saved) {
        toast('success', t('saved'));
        dirty = false;
        if (saved && saved.id) {
          var idx = items.map(function (c) { return c.id; }).indexOf(saved.id);
          if (idx === -1) items.push(saved); else items[idx] = saved;
          current = saved;
        }
        return 'done';
      }, function (err) {
        if (!form.showServerError(err)) top.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
        return false;
      });
    },
    dirty: function () { return dirty; }
  };
}

function needQuestion(labelKey, value, onchange) {
  var seg = segmented([{ value: 'yes', label: t('yes') }, { value: 'no', label: t('setup.no') }], value, onchange, t(labelKey), 'md');
  seg.setAttribute('data-need', '');
  return card2(h('div', { class: 'flex flex-col sm:flex-row sm:items-center justify-between gap-3' },
    h('div', { class: 'font-medium', text: t(labelKey) }), seg), 'mb-4');
}

function sandboxBox(on, onchange, confirmOff) {
  var text = h('p', { class: 'text-xs mt-1.5' });
  var boxEl = h('div', { 'data-sandbox-box': '' });
  var sw = toggle(on, t('pv.sandbox'), function (v) {
    if (!v && confirmOff) {
      confirmModal({ title: t('pv.live.title'), text: t('pv.live.text'), ok: t('pv.live.ok'), danger: true, icon: 'alert' }).then(function (yes) {
        if (!yes) { sw.set(true); return; }
        on = false; paint(); onchange(false);
      });
      return;
    }
    on = v; paint(); onchange(v);
  });
  boxEl.appendChild(h('div', { class: 'flex items-center gap-2' }, icon('flask', 'h-4 w-4 shrink-0'), sw));
  boxEl.appendChild(text);
  function paint() {
    boxEl.className = 'rounded-md px-3 py-2.5 mb-4 setup-sandbox ' + (on ? 'eco-sandbox-note' : 'eco-live-note');
    text.textContent = on ? t('setup.sandbox.on') : t('setup.sandbox.off');
  }
  paint();
  return boxEl;
}

function stepPayment(box) {
  var methods = [];
  var picked = {};
  var cfgs = {};
  var els = {};
  var need = null;
  var dirty = false;
  var top = h('div', {});
  box.appendChild(questionSkeleton(2));
  function methodCard(m) {
    var cfg = cfgs[m.method];
    var fieldEls = {};
    var fieldsBox = h('div', { class: 'setup-method-fields' });
    var sandboxField = (m.fields || []).filter(function (f) { return f.name === 'sandbox'; })[0];
    if (sandboxField) fieldsBox.appendChild(sandboxBox(cfg.sandbox !== false, function (v) { cfg.sandbox = v; dirty = true; }, true));
    (m.fields || []).forEach(function (f) {
      if (f.type === 'bool') return;
      var label = has('pm.field.' + f.name) ? t('pm.field.' + f.name) : f.name;
      var ctl = textInput(cfg[f.name] == null ? '' : cfg[f.name], { ltr: true, mono: true, multiline: f.type === 'textarea', rows: 3, label: label, oninput: function (v) { cfg[f.name] = v; dirty = true; fieldEls[f.name].setError(''); } });
      ctl.classList.add('text-sm');
      if (f.type === 'secret') ctl.classList.add('eco-secret');
      var hint = f.type === 'secret' && cfg[f.name] ? t('pm.secret.hint') : f.type === 'secret' && sandboxField ? t('setup.pay.secret.sandbox') : null;
      fieldEls[f.name] = field(label, ctl, hint, { optional: !f.required });
      fieldsBox.appendChild(fieldEls[f.name]);
    });
    if (!(m.fields || []).length) fieldsBox.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: t('pm.noFields') }));
    var errBox = h('div', {});
    var check = checkbox(picked[m.method], has('method.' + m.method) ? t('method.' + m.method) : m.method, function (on) {
      picked[m.method] = on;
      dirty = true;
      fieldsBox.hidden = !on;
      el.classList.toggle('setup-method--on', on);
    });
    check.setAttribute('data-pick-method', m.method);
    fieldsBox.hidden = !picked[m.method];
    var el = card2([
      h('div', { class: 'flex items-center justify-between gap-3' }, check,
        h('div', { class: 'flex items-center gap-1.5' }, mono(m.method), tag(m.configured ? t('pm.configured') : t('pm.notConfigured'), m.configured ? 'success' : 'secondary'))),
      errBox,
      fieldsBox
    ], 'setup-method mb-3' + (picked[m.method] ? ' setup-method--on' : ''));
    el.setAttribute('data-method', m.method);
    els[m.method] = { el: el, fields: fieldEls, err: errBox };
    return el;
  }
  function paint() {
    var list = h('div', { 'data-methods': '' });
    methods.forEach(function (m) { list.appendChild(methodCard(m)); });
    list.hidden = need !== 'yes';
    var noBox = h('div', { hidden: need !== 'no' }, notice('info', t('setup.pay.no')));
    clear(box).appendChild(needQuestion('setup.pay.need', need, function (v) { need = v; dirty = true; list.hidden = v !== 'yes'; noBox.hidden = v !== 'no'; }));
    box.appendChild(top);
    box.appendChild(noBox);
    box.appendChild(list);
  }
  function load() {
    api('GET', '/payment-methods').then(function (res) {
      if (!box.isConnected) return;
      methods = Array.isArray(res) ? res : (res && (res.items || res.methods)) || [];
      methods.forEach(function (m) {
        picked[m.method] = !!m.isActive;
        cfgs[m.method] = clone(m.config || {}) || {};
        if (!m.configured && !m.isActive && (m.fields || []).some(function (f) { return f.name === 'sandbox'; })) cfgs[m.method].sandbox = true;
      });
      need = methods.some(function (m) { return m.isActive; }) ? 'yes' : null;
      paint();
    }, function (err) { failBox(box, err, load); });
  }
  load();
  return {
    save: function () {
      clear(top);
      if (!need) { top.appendChild(h('div', { class: 'mb-4' }, notice('error', t('setup.err.answer')))); return false; }
      if (need === 'no') return 'skipped';
      var changed = methods.filter(function (m) { return picked[m.method] !== !!m.isActive || (picked[m.method] && JSON.stringify(cfgs[m.method]) !== JSON.stringify(m.config || {})); });
      if (!methods.some(function (m) { return picked[m.method]; })) { top.appendChild(h('div', { class: 'mb-4' }, notice('error', t('setup.pay.pickOne')))); return false; }
      var failed = false;
      return changed.reduce(function (p, m) {
        return p.then(function () {
          var e = els[m.method];
          clear(e.err);
          Object.keys(e.fields).forEach(function (k) { e.fields[k].setError(''); });
          return api('PUT', '/payment-methods/' + encodeURIComponent(m.method), { isActive: !!picked[m.method], config: cfgs[m.method] }).then(function (res) {
            if (res && res.method) { m.isActive = res.isActive; m.config = res.config; m.configured = res.configured; cfgs[m.method] = clone(res.config || {}) || {}; }
          }, function (err) {
            failed = true;
            if (err.field && e.fields[err.field]) e.fields[err.field].setError(errorText(err));
            else e.err.appendChild(h('div', { class: 'mt-3' }, notice('error', errorText(err))));
            e.el.scrollIntoView({ block: 'center', behavior: 'smooth' });
          });
        });
      }, Promise.resolve()).then(function () {
        if (failed) return false;
        dirty = false;
        if (changed.length) toast('success', t('saved'));
        return 'done';
      });
    },
    dirty: function () { return dirty; }
  };
}

function stepProvider(channel) {
  return function (box) {
    var catalog = [];
    var items = [];
    var need = null;
    var kind = null;
    var cfg = {};
    var st = { sandbox: true };
    var fieldEls = {};
    var dirty = false;
    var top = h('div', {});
    var formBox = h('div', {});
    var testBoxEl = h('div', {});
    box.appendChild(questionSkeleton(1));
    function current() { return items.filter(function (p) { return p.kind === kind; })[0] || null; }
    function buildForm2() {
      var p = current();
      var def = catalog.filter(function (k) { return k.channel === channel && k.kind === kind; })[0] || { fields: [] };
      cfg = p ? clone(p.config || {}) || {} : {};
      st = { sandbox: p ? !!p.sandbox : true };
      fieldEls = {};
      clear(formBox);
      formBox.appendChild(sandboxBox(st.sandbox, function (v) { st.sandbox = v; dirty = true; paintTest(); }, true));
      (def.fields || []).forEach(function (f) {
        var w = providerField(f, cfg, function () { dirty = true; if (fieldEls[f.name]) fieldEls[f.name].setError(''); });
        fieldEls[f.name] = w;
        formBox.appendChild(w);
      });
      if (!p) formBox.appendChild(h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 -mt-2 mb-2', text: t('setup.pv.new') }));
      paintTest();
    }
    function saveProvider() {
      clear(top);
      Object.keys(fieldEls).forEach(function (k) { fieldEls[k].setError(''); });
      var p = current();
      var def = catalog.filter(function (k) { return k.channel === channel && k.kind === kind; })[0] || { fields: [] };
      var out = {};
      (def.fields || []).forEach(function (f) { if (cfg[f.name] !== undefined && cfg[f.name] !== null) out[f.name] = String(cfg[f.name]); });
      var body = { isActive: true, isDefault: true, sandbox: st.sandbox, config: out };
      var call = p ? pubApi('PUT', '/providers/' + encodeURIComponent(p.id), body) : pubApi('POST', '/providers', Object.assign({ channel: channel, kind: kind }, body));
      return call.then(function (saved) {
        var idx = items.map(function (x) { return x.id; }).indexOf(saved.id);
        if (idx === -1) items.push(saved); else items[idx] = saved;
        items.forEach(function (x) { if (x.id !== saved.id) x.isDefault = false; });
        dirty = false;
        buildForm2();
        return true;
      }, function (err) {
        if (err.field && fieldEls[err.field]) { fieldEls[err.field].setError(errorText(err)); fieldEls[err.field].scrollIntoView({ block: 'center', behavior: 'smooth' }); }
        else top.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
        return false;
      });
    }
    function paintTest() {
      clear(testBoxEl);
      var email = channel === 'email';
      var p = current();
      var to = textInput('', { ltr: true, mono: true, label: t('pv.test.to'), placeholder: email ? 'name@example.com' : '09xxxxxxxxx', inputmode: email ? 'email' : 'tel' });
      to.classList.add('text-sm');
      var subject = email ? textInput('', { label: t('pv.test.subject'), placeholder: t('pv.test.subject.ph') }) : null;
      if (subject) subject.classList.add('text-sm');
      var msg = textInput(t('pv.test.message.default'), { multiline: true, rows: 2, label: t('pv.test.message') });
      msg.classList.add('text-sm');
      var out = h('div', { 'aria-live': 'polite' });
      var live = !st.sandbox;
      var sendBtn = button({ variant: 'link', size: 'sm', icon: 'send', label: t('setup.test.send'), disabled: live, onclick: function () {
        clear(out);
        setBusy(sendBtn, true, t('nt.sending'));
        var before = dirty || !current() ? saveProvider() : Promise.resolve(true);
        before.then(function (ok) {
          var cur = current();
          if (!ok || !cur) { setBusy(sendBtn, false); return; }
          if (!cur.sandbox) { setBusy(sendBtn, false); out.appendChild(notice('error', t('setup.test.liveBlocked'))); return; }
          var payload = { to: toLatinDigits(to.value).trim(), message: msg.value };
          if (subject) payload.subject = subject.value;
          return pubApi('POST', '/providers/' + encodeURIComponent(cur.id) + '/test', payload).then(function (r) {
            setBusy(sendBtn, false);
            if (r && r.ok && r.sandbox) out.appendChild(h('div', { class: 'eco-sandbox-note rounded-md px-3 py-2 text-sm flex items-start gap-2', 'data-test-result': 'sandbox' }, icon('flask', 'h-4 w-4 shrink-0 mt-0.5'), h('span', {}, h('span', { text: t('pv.test.sandbox') + ' ' }), r.reply ? mono(r.reply) : null)));
            else if (r && r.ok) out.appendChild(h('div', { 'data-test-result': 'ok' }, notice('success', t('pv.test.ok'))));
            else out.appendChild(h('div', { 'data-test-result': 'error' }, notice('error', t('pv.test.failed') + ' ' + ((r && r.error) || ''))));
          }, function (err) {
            setBusy(sendBtn, false);
            var bad = err.field === 'to';
            out.appendChild(h('div', { 'data-test-result': 'error' }, notice('error', errorText(err))));
            if (bad) to.focus();
          });
        });
      } });
      testBoxEl.appendChild(card2([
        h('div', { class: 'flex items-center gap-2 font-medium mb-1' }, icon('send', 'h-4 w-4'), h('span', { text: t('setup.test') })),
        h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: live ? t('setup.test.liveBlocked') : t('setup.test.hint') }),
        field(t('pv.test.to'), to),
        subject ? field(t('pv.test.subject'), subject, null, { optional: true }) : null,
        field(t('pv.test.message'), msg),
        h('div', { class: 'flex justify-end -mt-2' }, sendBtn),
        out
      ], 'mt-4'));
      if (p) testBoxEl.setAttribute('data-provider-id', p.id);
    }
    function paint() {
      var kinds = catalog.filter(function (k) { return k.channel === channel; });
      var seg = segmented(kinds.map(function (k) { return { value: k.kind, label: kindLabel(k.kind) }; }), kind, function (v) { kind = v; dirty = true; buildForm2(); clear(top); }, t('pv.kind'), 'md');
      seg.setAttribute('data-kind', '');
      var yesBox = h('div', { hidden: need !== 'yes' },
        card2([
          field(t('setup.pv.kind'), seg, t('setup.pv.kind.hint')),
          top,
          formBox
        ]),
        testBoxEl);
      var noBox = h('div', { hidden: need !== 'no' }, notice('info', t('setup.' + channel + '.no')));
      clear(box).appendChild(needQuestion('setup.' + channel + '.need', need, function (v) { need = v; dirty = true; yesBox.hidden = v !== 'yes'; noBox.hidden = v !== 'no'; }));
      box.appendChild(noBox);
      box.appendChild(yesBox);
      buildForm2();
    }
    function load() {
      pubApi('GET', '/providers').then(function (res) {
        if (!box.isConnected) return;
        catalog = (res && res.catalog) || [];
        items = ((res && res.items) || []).filter(function (p) { return p.channel === channel; });
        var def = items.filter(function (p) { return p.isDefault; })[0] || items.filter(function (p) { return p.isActive; })[0] || items[0];
        kind = def ? def.kind : (catalog.filter(function (k) { return k.channel === channel; })[0] || {}).kind;
        need = items.some(function (p) { return p.isActive; }) ? 'yes' : null;
        paint();
        dirty = false;
      }, function (err) { failBox(box, err, load); });
    }
    load();
    return {
      save: function () {
        clear(top);
        if (!need) { top.appendChild(h('div', { class: 'mb-4' }, notice('error', t('setup.err.answer')))); var q = box.querySelector('[data-need]'); if (q) q.scrollIntoView({ block: 'center' }); return false; }
        if (need === 'no') return 'skipped';
        return saveProvider().then(function (ok) { if (ok) toast('success', t('saved')); return ok ? 'done' : false; });
      },
      dirty: function () { return dirty; }
    };
  };
}

function stepReview(box, w) {
  box.appendChild(stepSkeleton(6));
  function load() {
    Promise.all([
      baseApi('GET', '/setup'),
      baseApi('GET', '/brand'),
      baseApi('GET', '/languages?limit=100'),
      api('GET', '/currencies?limit=100').then(null, function () { return null; }),
      api('GET', '/payment-methods').then(null, function () { return null; }),
      pubApi('GET', '/providers').then(null, function () { return null; })
    ]).then(function (all) {
      if (!box.isConnected || !wiz) return;
      wiz.state = all[0];
      paintBar();
      var b = all[1] || {};
      var langs = (all[2] && all[2].items) || [];
      var curs = ((all[3] && all[3].items) || []).slice().sort(function (a, c) { return (a.sort || 0) - (c.sort || 0); });
      var pms = Array.isArray(all[4]) ? all[4] : [];
      var pvs = (all[5] && all[5].items) || [];
      var summary = {
        brand: brandCustom(b) ? h('span', { class: 'flex items-center gap-2 min-w-0' }, b.accent ? h('span', { class: 'setup-dot', style: 'background-color:' + b.accent }) : null, h('span', { class: 'truncate', text: brandText(b.name) || b.shortName || t('setup.rv.brand.custom') })) : t('setup.rv.brand.default'),
        languages: langs.filter(function (l) { return l.isActive; }).map(function (l) { return (l.nativeName || l.code) + (l.isDefault ? ' (' + t('lang.default') + ')' : ''); }).join(t('setup.rv.sep')),
        currency: curs.length ? (i18nText(curs[0].name) || curs[0].key) + ' · ' + t('setup.rv.balance', { n: localDigits(fmtNum(curs[0].defaultValue || 0)) }) : t('setup.rv.none'),
        payment: pms.filter(function (m) { return m.isActive; }).map(function (m) { return t('method.' + m.method) + (m.config && m.config.sandbox ? ' (' + t('pv.sandbox.short') + ')' : ''); }).join(t('setup.rv.sep')) || t('setup.rv.none'),
        sms: provSummary(pvs, 'sms'),
        email: provSummary(pvs, 'email')
      };
      var list = h('ul', { class: 'setup-list' });
      SETUP_ORDER.slice(0, -1).forEach(function (k) {
        var s = setupStepStatus(wiz.state, k);
        list.appendChild(h('li', { class: 'setup-list-row', 'data-review': k },
          h('div', { class: 'setup-section-icon setup-section-icon--sm' }, icon(SETUP_ICONS[k], 'h-4 w-4')),
          h('div', { class: 'min-w-0 flex-1' },
            h('div', { class: 'font-medium', text: t('setup.step.' + k) }),
            h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400 truncate' }, summary[k])),
          h('div', { class: 'flex items-center gap-2 shrink-0' },
            tag(t('setup.st.' + s), STEP_COLOR[s]),
            button({ variant: 'link', size: 'sm', label: t('setup.rv.edit'), title: t('setup.rv.editStep', { step: t('setup.step.' + k) }), onclick: function () { goStep(k); } }))));
      });
      var left = SETUP_ORDER.slice(0, -1).filter(function (k) { return setupStepStatus(wiz.state, k) === 'pending'; }).length;
      clear(box).appendChild(card2(list));
      box.appendChild(h('div', { class: 'mt-4' }, notice(left ? 'info' : 'success', left ? t('setup.rv.left', { n: localDigits(left) }) : t('setup.rv.ready'))));
    }, function (err) { failBox(box, err, load); });
  }
  load();
  return { save: function () { return 'done'; }, dirty: function () { return false; } };
}

function provSummary(pvs, channel) {
  var list = pvs.filter(function (p) { return p.channel === channel && p.isActive; });
  if (!list.length) return t('setup.rv.none');
  var d = list.filter(function (p) { return p.isDefault; })[0] || list[0];
  return kindLabel(d.kind) + (d.sandbox ? ' (' + t('pv.sandbox.short') + ')' : '');
}

var STEP_BUILDERS = {
  brand: stepBrand,
  languages: stepLanguages,
  currency: stepCurrency,
  payment: stepPayment,
  sms: stepProvider('sms'),
  email: stepProvider('email'),
  review: stepReview
};

function setupPage(container) {
  var openBtn = button({ variant: 'primary', size: 'page', icon: 'flask', label: t('setup.open'), onclick: function () { openFromPage(); } });
  openBtn.setAttribute('data-setup-open', '');
  var frame = tabFrame(t('desc.setup'), openBtn);
  clear(container).appendChild(frame.el);
  var body = h('div', {});
  frame.content.appendChild(body);
  var state = null;
  function openFromPage(step) {
    openWizard(state, false, step);
  }
  function skeleton() {
    clear(body).appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'flex items-center justify-between mb-4' }, bar('10rem', 'h-5'), bar('5rem', 'h-5 rounded-full')), bar('100%', 'h-2 mb-4 rounded-full'), stepSkeleton(6))));
  }
  function load() {
    skeleton();
    baseApi('GET', '/setup').then(function (st) {
      if (!container.isConnected) return;
      state = st;
      var steps = SETUP_ORDER.slice(0, -1);
      var finished = steps.filter(function (k) { return setupStepStatus(st, k) !== 'pending'; }).length;
      var list = h('ul', { class: 'setup-list' });
      steps.forEach(function (k, i) {
        var s = setupStepStatus(st, k);
        list.appendChild(h('li', { class: 'setup-list-row', 'data-setup-row': k },
          h('div', { class: 'setup-section-icon setup-section-icon--sm' }, icon(SETUP_ICONS[k], 'h-4 w-4')),
          h('div', { class: 'min-w-0 flex-1' },
            h('div', { class: 'font-medium', text: localDigits(i + 1) + '. ' + t('setup.step.' + k) }),
            h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('setup.step.' + k + '.desc') })),
          h('div', { class: 'flex items-center gap-2 shrink-0' }, tag(t('setup.st.' + s), STEP_COLOR[s]),
            button({ variant: 'link', size: 'sm', label: t('setup.rv.edit'), title: t('setup.rv.editStep', { step: t('setup.step.' + k) }), onclick: function () { openFromPage(k); } }))));
      });
      var meta = [];
      var who = function (user, at) { return [user ? h('bdi', { text: user }) : null, user ? ' · ' : null, h('bdi', { text: fmtDate(at) })]; };
      if (st.firstOpenedAt) meta.push([t('setup.meta.opened'), who(st.firstOpenedBy, st.firstOpenedAt)]);
      if (st.completedAt) meta.push([t('setup.meta.completed'), who(st.completedBy, st.completedAt)]);
      clear(body).appendChild(card(h('div', { class: 'p-4' },
        h('div', { class: 'flex flex-wrap items-center justify-between gap-3 mb-3' },
          h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'text-lg font-medium', text: t('setup.title') }), tag(t('setup.status.' + st.status), SETUP_STATUS_COLOR[st.status] || 'secondary')),
          h('span', { class: 'text-sm text-secondary-500 tabular-nums', text: t('setup.progress', { n: localDigits(finished), total: localDigits(steps.length) }) })),
        h('div', { class: 'setup-progress', role: 'progressbar', 'aria-valuemin': '0', 'aria-valuemax': String(steps.length), 'aria-valuenow': String(finished), 'aria-label': t('setup.progress', { n: localDigits(finished), total: localDigits(steps.length) }) },
          h('div', { class: 'setup-progress-bar', style: 'width:' + Math.round(finished / steps.length * 100) + '%' })),
        meta.length ? h('dl', { class: 'eco-dl mt-4 mb-2' }, meta.map(function (m) { return [h('dt', { class: 'text-secondary-500', text: m[0] }), h('dd', {}, m[1])]; })) : null,
        list)));
    }, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}

var PS_ROUTE = 'project-settings';
var PS_SECTIONS = [
  { id: 'brand', label: 'tab.brand', icon: 'image', acl: 'SETTINGS', render: function (box) { brandPage(box); } },
  { id: 'setup', label: 'tab.setup', icon: 'flask', acl: 'SETTINGS', render: function (box) { setupPage(box); } },
  { id: 'languages', label: 'tab.languages', icon: 'languages', acl: 'SETTINGS', render: function (box, route, ctx) { return languagesPage(box, route.query, ctx); } },
  { id: 'countries', label: 'tab.countries', icon: 'flag', acl: 'SETTINGS', render: function (box, route, ctx) { return countriesPage(box, route.query, ctx); } },
  { id: 'accounts', label: 'tab.accounts', icon: 'users', acl: 'SETTINGS', render: function (box) { accountRulesPage(box); } },
  { id: 'otp', label: 'tab.otp', icon: 'phone', acl: 'NOTIFICATION', render: function (box, route, ctx) { return otpPage(box, route.query, ctx); } },
  { id: 'app-versions', label: 'tab.app-versions', icon: 'versions', acl: 'SETTINGS', render: function (box, route, ctx) { return appVersionsPage(box, route.query, ctx); } },
  { id: 'categories', label: 'nav.categories', icon: 'categories', acl: 'HIRO_ECONOMY', render: function (box, route, ctx) { return listPage(box, 'categories', route.query, ctx); } },
  { id: 'tags', label: 'nav.tags', icon: 'tags', acl: 'HIRO_ECONOMY', render: function (box, route, ctx) { return listPage(box, 'tags', route.query, ctx); } },
  { id: 'report-titles', label: 'tab.report-titles', icon: 'report-titles', acl: 'CHANNEL_MESSAGE', render: function (box, route, ctx) { return listPage(box, 'report-titles', route.query, ctx); } },
  { id: 'events', label: 'tab.events', icon: 'event-types', acl: 'HIRO_ECONOMY', render: function (box) { eventSettingsPage(box); } },
  { id: 'referral-rules', label: 'tab.referral-rules', icon: 'referral-rules', acl: 'HIRO_ECONOMY', render: function (box) { referralRulesPage(box); } },
  { id: 'support-departments', label: 'tab.support-departments', icon: 'support-departments', acl: 'ACCOUNT_NOTES', render: function (box, route, ctx) { loadAgents(); return listPage(box, 'support-departments', route.query, ctx); } },
  { id: 'support-canned', label: 'tab.support-canned', icon: 'support-canned', acl: 'ACCOUNT_NOTES', render: function (box, route, ctx) { return listPage(box, 'support-canned', route.query, ctx); } },
  { id: 'support-settings', label: 'tab.support-settings', icon: 'support-settings', acl: 'ACCOUNT_NOTES', render: function (box) { supportSettingsPage(box); } },
  { id: 'payment-methods', label: 'tab.payment-methods', icon: 'payments', acl: 'HIRO_ECONOMY', render: function (box) { paymentMethodsPage(box); } },
  { id: 'sms', label: 'tab.sms', icon: 'comments', acl: 'NOTIFICATION', render: function (box) { providersPage(box, 'sms'); } },
  { id: 'email', label: 'tab.email', icon: 'mail', acl: 'NOTIFICATION', render: function (box) { providersPage(box, 'email'); } }
];
var PS_LEGACY = {
  'settings/brand': 'brand',
  'settings/setup': 'setup',
  'settings/languages': 'languages',
  'settings/payment-methods': 'payment-methods',
  'settings/sms': 'sms',
  'settings/email': 'email',
  'base/categories': 'categories',
  'base/tags': 'tags',
  'economy/categories': 'categories',
  'economy/tags': 'tags',
  'economy/payment-methods': 'payment-methods',
  'publication/providers': 'sms'
};
var ACL_RESOURCES = { ACCOUNT: 0, ACCOUNT_NOTES: 5, CHANNEL_MESSAGE: 13, NOTIFICATION: 20, SETTINGS: 22, HIRO_ECONOMY: 27 };
var PS_RETURN_KEY = 'economy.ps.return';
var aclCache = { tk: null, bits: null };
var ps = null;
var psReturn = null;
try { psReturn = sessionStorage.getItem(PS_RETURN_KEY); } catch (e) { psReturn = null; }

function psSection(id) {
  for (var i = 0; i < PS_SECTIONS.length; i++) if (PS_SECTIONS[i].id === id) return PS_SECTIONS[i];
  return null;
}

function b64urlBytes(s) {
  s = String(s || '').replace(/-/g, '+').replace(/_/g, '/');
  while (s.length % 4) s += '=';
  var bin = atob(s);
  var out = [];
  for (var i = 0; i < bin.length; i++) out.push(bin.charCodeAt(i));
  return out;
}

function consoleAcl() {
  var tk = token();
  if (!tk) return null;
  if (aclCache.tk === tk) return aclCache.bits;
  var bits = null;
  try {
    var payload = JSON.parse(new TextDecoder().decode(new Uint8Array(b64urlBytes(tk.split('.')[1]))));
    if (payload && typeof payload.acl === 'string' && payload.acl) bits = b64urlBytes(payload.acl);
  } catch (e) { bits = null; }
  aclCache = { tk: tk, bits: bits };
  return bits;
}

function aclCan(resource, level) {
  var bits = consoleAcl();
  if (!bits || !bits.length) return true;
  if (bits.every(function (b) { return b === 0xFF; })) return true;
  var idx = ACL_RESOURCES[resource] * 3 + (level || 0);
  return !!(bits[idx >> 3] & (1 << (7 - (idx & 7))));
}

function psAllowed() {
  return PS_SECTIONS.filter(function (s) { return aclCan(s.acl, 0); });
}

function psEntryHash() {
  var allowed = psAllowed();
  return hashFor(allowed.length ? allowed[0].id : PS_SECTIONS[0].id);
}

function psRemember(hash) {
  if (!hash || /^#\/login/.test(hash)) return;
  psReturn = hash;
  try { sessionStorage.setItem(PS_RETURN_KEY, hash); } catch (e) { return; }
}

function psNavigate(hash, replace) {
  if (router) {
    var p = replace ? router.replace(hash.slice(1)) : router.push(hash.slice(1));
    Promise.resolve(p).then(schedule, schedule);
    return;
  }
  if (replace) history.replaceState(history.state, '', location.pathname + location.search + hash);
  else location.hash = hash;
  schedule();
}

function psEnter(e) {
  if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
  e.preventDefault();
  if (ps) { if (ps.list) { var tab = ps.list.querySelector('[aria-selected="true"]'); if (tab) tab.focus(); } return; }
  psNavigate(e.currentTarget.getAttribute('href'), false);
}

function psExit() {
  var back = psReturn || '#/';
  var st = history.state;
  if (st && typeof st.back === 'string' && '#' + st.back === back) history.back();
  else psNavigate(back, true);
}

function psRefresh(page) {
  if (ps && ps.ctl && ps.ctl.key && ps.ctl.key.indexOf('|' + page + '|') !== -1) { ps.ctl.key = ''; schedule(); }
}

function psKeys(e) {
  if (!ps || wiz || document.querySelector('[data-economy="dialog"]') || panelState) return;
  if (e.key === 'Tab') { trapTab(e, ps.win); return; }
  if (e.key !== 'Escape' || e.defaultPrevented) return;
  if (ps.el.querySelector('.eco-pop:not([hidden]), .eco-qc-pop:not([hidden])')) return;
  e.preventDefault();
  psExit();
}

function psFocusIn(e) {
  if (!ps || wiz || panelState || document.querySelector('[data-economy="dialog"]')) return;
  if (ps.el.contains(e.target) || (e.target.closest && e.target.closest('.eco-toasts'))) return;
  var list = focusables(ps.win).filter(function (el) { return el.getAttribute('tabindex') !== '-1'; });
  var target = ps.list.querySelector('[aria-selected="true"]') || list[0];
  if (target) target.focus({ preventScroll: true });
}

function psListKeys(e) {
  var items = Array.prototype.slice.call(ps.list.querySelectorAll('[role="tab"]'));
  var i = items.indexOf(document.activeElement);
  if (i === -1) return;
  var rtl = (document.documentElement.getAttribute('dir') || 'ltr') === 'rtl';
  var next = null;
  if (e.key === 'ArrowDown' || e.key === (rtl ? 'ArrowLeft' : 'ArrowRight')) next = items[(i + 1) % items.length];
  else if (e.key === 'ArrowUp' || e.key === (rtl ? 'ArrowRight' : 'ArrowLeft')) next = items[(i - 1 + items.length) % items.length];
  else if (e.key === 'Home') next = items[0];
  else if (e.key === 'End') next = items[items.length - 1];
  if (!next) return;
  e.preventDefault();
  next.focus();
  psNavigate(next.getAttribute('href'), true);
}

function psMount() {
  var list = h('div', { class: 'ps-list', role: 'tablist', 'aria-orientation': 'vertical' });
  list.addEventListener('keydown', psListKeys);
  var headIcon = h('span', { class: 'ps-head-icon' });
  var headTitle = h('h3', { id: 'ps-section-title', class: 'ps-head-title' });
  var body = h('div', { id: 'project-settings-body', class: 'ps-body', role: 'tabpanel', 'aria-labelledby': 'ps-section-title', 'data-l10n-skip': '' });
  var scroller = h('div', { class: 'ps-scroll' }, body);
  var title = h('h2', { id: 'ps-title', class: 'ps-title' });
  var close = button({ variant: 'ghost', size: 'iconLg', icon: 'close', title: t('close'), class: 'ps-close', onclick: psExit });
  close.setAttribute('data-ps-close', '');
  var win = h('div', { class: 'ps-window', role: 'dialog', 'aria-modal': 'true', 'aria-labelledby': 'ps-title' },
    h('aside', { class: 'ps-side' },
      h('div', { class: 'ps-side-head' }, h('span', { class: 'ps-mark' }, icon('gear', 'h-5 w-5')), title),
      list),
    h('section', { class: 'ps-main' },
      h('header', { class: 'ps-head' }, headIcon, headTitle),
      scroller),
    close);
  var holder = h('div', { class: 'ps-holder' }, win);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) psExit(); });
  var el = h('div', { 'data-economy': 'project-settings', 'data-l10n-skip': '' }, h('div', { class: 'ps-shade' }), holder);
  ps = { el: el, win: win, list: list, body: body, scroller: scroller, title: title, headIcon: headIcon, headTitle: headTitle, ctl: null, listKey: null, page: null, prevFocus: document.activeElement };
  if (rendered) rendered.id = undefined;
  closePanel();
  document.body.appendChild(el);
  document.documentElement.classList.add('ps-open');
  document.documentElement.setAttribute('data-ps-open', '');
  document.addEventListener('keydown', psKeys, true);
  document.addEventListener('focusin', psFocusIn, true);
}

function psHide() {
  if (!ps) return;
  var old = ps;
  ps = null;
  closePanel();
  old.el.remove();
  document.removeEventListener('keydown', psKeys, true);
  document.removeEventListener('focusin', psFocusIn, true);
  document.documentElement.classList.remove('ps-open');
  document.documentElement.removeAttribute('data-ps-open');
  if (old.prevFocus && old.prevFocus.isConnected && !old.el.contains(old.prevFocus)) { try { old.prevFocus.focus({ preventScroll: true }); } catch (e) { return; } }
}

function psPaintList(allowed, page) {
  var key = lang() + '|' + page + '|' + allowed.map(function (s) { return s.id; }).join(',');
  if (ps.listKey === key) return;
  ps.listKey = key;
  ps.title.textContent = t('ps.title');
  ps.list.setAttribute('aria-label', t('ps.sections'));
  ps.list.setAttribute('aria-orientation', window.matchMedia && window.matchMedia('(max-width: 767px)').matches ? 'horizontal' : 'vertical');
  var focused = ps.list.contains(document.activeElement);
  var have = Array.prototype.map.call(ps.list.children, function (a) { return a.getAttribute('data-ps-section'); }).join(',');
  if (have && have === allowed.map(function (s) { return s.id; }).join(',')) {
    var cur = null;
    Array.prototype.forEach.call(ps.list.children, function (a) {
      var id = a.getAttribute('data-ps-section');
      var on = id === page;
      a.setAttribute('aria-selected', on ? 'true' : 'false');
      a.setAttribute('tabindex', on ? '0' : '-1');
      a.querySelector('span').textContent = t(psSection(id).label);
      if (on) cur = a;
    });
    if (cur && focused && document.activeElement !== cur) cur.focus({ preventScroll: true });
    if (cur && cur.scrollIntoView && ps.list.scrollWidth > ps.list.clientWidth) cur.scrollIntoView({ block: 'nearest', inline: 'nearest' });
    return;
  }
  clear(ps.list);
  var current = null;
  allowed.forEach(function (s) {
    var on = s.id === page;
    var a = h('a', { href: hashFor(s.id), role: 'tab', id: 'ps-tab-' + s.id, class: 'ps-item', 'aria-selected': on ? 'true' : 'false', 'aria-controls': 'project-settings-body', tabindex: on ? '0' : '-1', 'data-ps-section': s.id },
      icon(s.icon, 'h-4 w-4'), h('span', { class: 'truncate', text: t(s.label) }));
    a.addEventListener('click', function (e) {
      if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
      e.preventDefault();
      if (s.id !== (ps && ps.page)) psNavigate(hashFor(s.id), true);
    });
    if (on) current = a;
    ps.list.appendChild(a);
  });
  if (current) {
    if (focused) current.focus({ preventScroll: true });
    if (current.scrollIntoView && ps.list.scrollWidth > ps.list.clientWidth) current.scrollIntoView({ block: 'nearest', inline: 'nearest' });
  }
}

function psShow(route) {
  var allowed = psAllowed();
  var sec = psSection(route.page);
  if (allowed.length && allowed.indexOf(sec) === -1) {
    var to = '#/' + PS_ROUTE + '/' + allowed[0].id;
    if (router) router.replace(to.slice(1)).then(schedule, schedule);
    else { history.replaceState(history.state, '', location.pathname + location.search + to); schedule(); }
    return;
  }
  var fresh = !ps;
  if (fresh) psMount();
  ps.el.setAttribute('dir', document.documentElement.getAttribute('dir') || 'ltr');
  psPaintList(allowed, route.page);
  if (!allowed.length) {
    if (ps.page !== null || fresh) {
      ps.page = null;
      ps.ctl = null;
      clear(ps.headIcon);
      ps.headTitle.textContent = t('ps.title');
      clear(ps.body).appendChild(card(emptyBlock(t('ps.noAccess'), t('ps.noAccess.hint'))));
    }
    if (fresh) ps.win.querySelector('[data-ps-close]').focus();
    return;
  }
  if (ps.page !== route.page || !ps.ctl || ps.headTitle.textContent !== t(sec.label)) {
    clear(ps.headIcon).appendChild(icon(sec.icon, 'h-5 w-5'));
    ps.headTitle.textContent = t(sec.label);
  }
  ps.page = route.page;
  ps.ctl = pageCtl(ps.ctl, ps.body, route, function (ctl) {
    ps.scroller.scrollTop = 0;
    ps.ctl = ctl;
  });
  if (fresh) {
    var tab = ps.list.querySelector('[aria-selected="true"]');
    if (tab) tab.focus({ preventScroll: true });
  }
}

function placePsNav(drawer, collapsed, on, key) {
  var scroll = drawer.querySelector(':scope > .x-scroll');
  var existing = drawer.querySelector(':scope > [data-ps-nav]');
  if (!scroll || !psAllowed().length) {
    if (existing) existing.remove();
    return null;
  }
  if (existing && existing.getAttribute('data-key') === key && existing.previousElementSibling === scroll) return existing;
  var a = h('a', { href: psEntryHash(), class: (on ? NAV_ACTIVE + ' ' : '') + NAV_LINK, 'aria-current': on ? 'page' : null, title: collapsed ? t('ps.title') : null, 'data-ps-entry': 'sidebar', onclick: psEnter });
  a.appendChild(collapsed ? h('div', { class: 'x-tooltip flex' }, icon('gear', 'h-4 w-4')) : icon('gear', 'h-4 w-4'));
  a.appendChild(h('div', { class: 'font-medium relative truncate' + (collapsed ? '' : ' flex-1') }, h('span', { text: collapsed ? ' ' : t('ps.title') })));
  var box = h('div', { class: 'ps-nav', 'data-ps-nav': '', 'data-l10n-skip': '', 'data-key': key }, h('nav', { 'aria-label': t('ps.title') }, a));
  if (existing) existing.replaceWith(box);
  else scroll.after(box);
  return box;
}

var psButton = null;

function dockPsButton() {
  if (!psButton) {
    psButton = h('a', { class: 'economy-ps', 'data-l10n-skip': '', 'data-ps-entry': 'topbar', onclick: psEnter }, icon('gear', 'h-4 w-4'));
  }
  var counter = document.querySelector('number-flow-vue');
  var bar = counter && counter.closest('.justify-end');
  var menu = bar && bar.querySelector(':scope > .x-popover');
  var show = !!menu && !!token() && psAllowed().length > 0;
  if (!show) {
    if (psButton.parentElement) psButton.remove();
    return;
  }
  var href = psEntryHash();
  if (psButton.getAttribute('href') !== href) psButton.setAttribute('href', href);
  if (psButton.getAttribute('aria-label') !== t('ps.title')) {
    psButton.setAttribute('title', t('ps.title'));
    psButton.setAttribute('aria-label', t('ps.title'));
  }
  if (ps) psButton.setAttribute('aria-current', 'page');
  else psButton.removeAttribute('aria-current');
  if (psButton.nextElementSibling !== menu) menu.before(psButton);
}

document.addEventListener('keydown', function (e) {
  if (e.key && e.key.length > 1 && !/^(Shift|Control|Alt|Meta|CapsLock)$/.test(e.key)) document.documentElement.setAttribute('data-eco-kbd', '');
}, true);
document.addEventListener('pointerdown', function () { document.documentElement.removeAttribute('data-eco-kbd'); }, true);

var NATIVE_PURPLE = /^(#6b59ee|rgb\(107, ?89, ?238\))$/i;
var CHART_AXIS = '20px';
var CHART_EDGE = '24px';
var CHART = { key: '', patched: false };
var MONACO = { node: null, len: -1, el: null };

function vueFind(test) {
  var app = document.getElementById('app');
  var out = [];
  var seen = new Set();
  function walkVnode(v, depth) {
    if (!v || depth > 800) return;
    if (v.component) walkInstance(v.component, depth + 1);
    if (Array.isArray(v.children)) {
      for (var i = 0; i < v.children.length; i++) {
        var c = v.children[i];
        if (c && typeof c === 'object') walkVnode(c, depth + 1);
      }
    }
    if (v.suspense && v.suspense.activeBranch) walkVnode(v.suspense.activeBranch, depth + 1);
  }
  function walkInstance(inst, depth) {
    if (seen.has(inst)) return;
    seen.add(inst);
    if (test(inst)) out.push(inst);
    walkVnode(inst.subTree, depth + 1);
  }
  walkVnode(app && app._vnode, 0);
  return out;
}

function faDigits(s) {
  return String(s).replace(/(\d),(?=\d{3}\b)/g, '$1٬').replace(/(\d)\.(?=\d)/g, '$1٫').replace(/\d/g, function (d) { return PERSIAN[+d]; });
}

function compactFa(v) {
  var s = String(v);
  var m = /^(-?\d+(?:\.\d+)?)\s?([kKM])$/.exec(s);
  if (m) s = String(Number(m[1])) + ' ' + (m[2] === 'M' ? 'میلیون' : 'هزار');
  return faDigits(s);
}

function chartAccent() {
  var b = BRAND.data;
  return b && b.accent ? brandHex(b.accent) : null;
}

function asList(v) {
  return Array.isArray(v) ? v : [v];
}

function mapAxis(v, fn) {
  if (!v) return v;
  return Array.isArray(v) ? v.map(fn) : fn(v);
}

function wrapFn(fn, make) {
  if (typeof fn !== 'function' || fn.__nk) return fn;
  var out = make(fn);
  out.__nk = true;
  return out;
}

function htmlDigits(html) {
  return String(html).replace(/>([^<]+)</g, function (m, text) { return '>' + faDigits(text) + '<'; });
}

function accentGradient(hex) {
  var rgb = hexRgb(hex).join(',');
  return { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(' + rgb + ',0.3)' }, { offset: 1, color: 'rgba(' + rgb + ',0)' }] };
}

var CHART_SECONDS = new Intl.DateTimeFormat(undefined, { hour: 'numeric', minute: '2-digit', second: '2-digit' });

function seriesSpan(series) {
  var min = Infinity;
  var max = -Infinity;
  asList(series || []).forEach(function (s) {
    (s && Array.isArray(s.data) ? s.data : []).forEach(function (d) {
      var v = Array.isArray(d) ? d[0] : d && d.value ? d.value[0] : null;
      if (typeof v === 'string') v = Date.parse(v);
      if (typeof v !== 'number' || !isFinite(v)) return;
      if (v < 1e11) v *= 1000;
      if (v < min) min = v;
      if (v > max) max = v;
    });
  });
  return max > min ? max - min : 0;
}

function chartOption(opt) {
  if (!opt || typeof opt !== 'object' || Array.isArray(opt)) return opt;
  var rtl = document.documentElement.getAttribute('dir') === 'rtl';
  var fa = lang() === 'fa';
  var accent = chartAccent();
  var o = Object.assign({}, opt);
  if (o.xAxis) {
    o.xAxis = mapAxis(o.xAxis, function (a) {
      var x = Object.assign({}, a, { inverse: rtl });
      var span = seriesSpan(o.series);
      if (x.type === 'time' && span > 0 && span < 900000 && x.axisLabel && typeof x.axisLabel.formatter === 'function' && !x.axisLabel.formatter.__nkSec) {
        var sec = function (v) { return CHART_SECONDS.format(new Date(v)); };
        sec.__nk = true;
        sec.__nkSec = true;
        x.axisLabel = Object.assign({}, x.axisLabel, { formatter: sec });
      }
      return x;
    });
  }
  if (o.yAxis) {
    o.yAxis = mapAxis(o.yAxis, function (a) {
      var y = Object.assign({}, a, { position: rtl ? 'right' : 'left' });
      if (fa && y.axisLabel && typeof y.axisLabel.formatter === 'function') {
        y.axisLabel = Object.assign({}, y.axisLabel, { formatter: wrapFn(y.axisLabel.formatter, function (f) { return function () { return compactFa(f.apply(this, arguments)); }; }) });
      }
      return y;
    });
  }
  if (o.grid) {
    o.grid = mapAxis(o.grid, function (g) {
      if (!g || !g.containLabel) return g;
      return Object.assign({}, g, rtl ? { left: CHART_EDGE, right: CHART_AXIS } : { left: CHART_AXIS, right: CHART_EDGE });
    });
  }
  if (o.title && rtl) {
    o.title = mapAxis(o.title, function (t) { return Object.assign({}, t, { left: 'right' }); });
  }
  if (o.legend && rtl) {
    o.legend = mapAxis(o.legend, function (l) { return Object.assign({}, l, { left: 'right' }); });
  }
  if (o.tooltip) {
    o.tooltip = mapAxis(o.tooltip, function (tt) {
      var x = Object.assign({}, tt);
      if (rtl) x.extraCssText = 'direction:rtl;text-align:right;';
      if (fa && typeof x.formatter === 'function') x.formatter = wrapFn(x.formatter, function (f) { return function () { return htmlDigits(f.apply(this, arguments)); }; });
      return x;
    });
  }
  if (accent && o.series) {
    o.series = asList(o.series).map(function (s) {
      if (!s || typeof s !== 'object') return s;
      var line = s.lineStyle && s.lineStyle.color;
      var item = s.itemStyle && s.itemStyle.color;
      if (!(typeof line === 'string' && NATIVE_PURPLE.test(line)) && !(typeof item === 'string' && NATIVE_PURPLE.test(item))) return s;
      var x = Object.assign({}, s);
      if (x.lineStyle) x.lineStyle = Object.assign({}, x.lineStyle, { color: accent });
      if (x.itemStyle) x.itemStyle = Object.assign({}, x.itemStyle, { color: accent });
      if (x.areaStyle) x.areaStyle = Object.assign({}, x.areaStyle, { color: accentGradient(accent) });
      x.emphasis = Object.assign({}, x.emphasis, { lineStyle: Object.assign({}, x.emphasis && x.emphasis.lineStyle, { color: accent }), itemStyle: Object.assign({}, x.emphasis && x.emphasis.itemStyle, { color: accent }) });
      return x;
    });
  }
  return o;
}

function patchCharts(proto) {
  if (!proto || proto.__nkCharts) return;
  var setOption = proto.setOption;
  Object.defineProperty(proto, '__nkCharts', { value: true });
  proto.setOption = function (opt) {
    var args = Array.prototype.slice.call(arguments);
    args[0] = chartOption(opt);
    return setOption.apply(this, args);
  };
}

function restyleChart(chart) {
  if (!chart || chart.isDisposed && chart.isDisposed()) return;
  var o;
  try { o = chart.getOption(); } catch (e) { return; }
  if (!o) return;
  var part = {};
  ['xAxis', 'yAxis', 'grid', 'title', 'legend', 'tooltip', 'series'].forEach(function (k) { if (o[k] && o[k].length) part[k] = o[k]; });
  chart.setOption(part);
}

function syncCharts() {
  var els = document.querySelectorAll('[_echarts_instance_]');
  if (!els.length) return;
  var key = Array.prototype.map.call(els, function (e) { return e.getAttribute('_echarts_instance_'); }).join(',') + '|' + chartAccent() + '|' + lang();
  if (key === CHART.key) return;
  CHART.key = key;
  vueFind(function (inst) { return inst.type && inst.type.name === 'echarts'; }).forEach(function (inst) {
    var chart = inst.setupState && inst.setupState.chart;
    if (!chart || typeof chart.setOption !== 'function') return;
    patchCharts(Object.getPrototypeOf(chart));
    restyleChart(chart);
  });
}

function hexHsl(hex) {
  var c = hexRgb(hex).map(function (v) { return v / 255; });
  var max = Math.max(c[0], c[1], c[2]);
  var min = Math.min(c[0], c[1], c[2]);
  var l = (max + min) / 2;
  var d = max - min;
  var hue = 0;
  var s = d === 0 ? 0 : d / (1 - Math.abs(2 * l - 1));
  if (d) {
    if (max === c[0]) hue = ((c[1] - c[2]) / d) % 6;
    else if (max === c[1]) hue = (c[2] - c[0]) / d + 2;
    else hue = (c[0] - c[1]) / d + 4;
  }
  return [(hue * 60 + 360) % 360, s, l];
}

function darkToken(hex) {
  var hsl = hexHsl(hex);
  if (hsl[2] >= 0.6) return hex;
  var l = Math.min(0.86, 0.7 + (0.6 - hsl[2]) * 0.3);
  var s = hsl[1] < 0.15 ? hsl[1] : Math.min(0.85, Math.max(0.5, hsl[1]));
  return 'hsl(' + Math.round(hsl[0]) + ' ' + Math.round(s * 100) + '% ' + Math.round(l * 100) + '%)';
}

function syncMonaco() {
  var node = document.querySelector('style.monaco-colors');
  if (!node) return;
  var len = node.textContent.length;
  if (node === MONACO.node && len === MONACO.len && MONACO.el && MONACO.el.isConnected) return;
  MONACO.node = node;
  MONACO.len = len;
  var rules = [];
  var re = /\.(mtk\d+) \{ color: (#[0-9a-fA-F]{6}); \}/g;
  var m;
  while ((m = re.exec(node.textContent))) rules.push('html.dark .monaco-editor .' + m[1] + '{color:' + darkToken(m[2]) + '}');
  if (!MONACO.el) MONACO.el = h('style', { id: 'nk-monaco-dark' });
  MONACO.el.textContent = rules.join('\n');
  if (MONACO.el.parentNode !== document.head || MONACO.el.nextSibling) document.head.appendChild(MONACO.el);
}

var SETUP_NUDGE = { state: null, at: 0, loading: false };

function dashboardSetup() {
  var path = (location.hash || '#/').replace(/^#/, '').split('?')[0];
  var card = document.querySelector('[data-setup-nudge]');
  if (path !== '/' && path !== '') { if (card) card.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.max-w-\\[1600px\\] .grid.grid-cols-1.gap-6');
  if (!grid || !token()) return;
  if (!SETUP_NUDGE.loading && Date.now() - SETUP_NUDGE.at > 15000) {
    SETUP_NUDGE.loading = true;
    quietBase('GET', '/setup').then(function (st) {
      SETUP_NUDGE.loading = false;
      SETUP_NUDGE.at = Date.now();
      SETUP_NUDGE.state = st;
      schedule();
    });
  }
  var st = SETUP_NUDGE.state;
  var show = !!st && st.status !== 'completed' && !wiz;
  if (!show) { if (card) card.remove(); return; }
  var steps = SETUP_ORDER.slice(0, -1);
  var done = steps.filter(function (k) { return setupStepStatus(st, k) !== 'pending'; }).length;
  var key = lang() + '|' + st.status + '|' + done;
  if (card && card.getAttribute('data-key') === key && card.parentNode === grid && grid.firstChild === card) return;
  if (card) card.remove();
  var progress = t('setup.progress', { n: localDigits(done), total: localDigits(steps.length) });
  var go = button({ variant: 'primary', size: 'sm', icon: 'flask', label: t('setup.nudge.go'), onclick: function () { openWizard(SETUP_NUDGE.state, false); SETUP_NUDGE.at = 0; } });
  go.setAttribute('data-setup-nudge-go', '');
  card = h('section', { 'data-setup-nudge': '', 'data-key': key, 'data-economy': 'nudge', 'data-l10n-skip': '', class: 'setup-nudge', 'aria-labelledby': 'setup-nudge-title', dir: document.documentElement.getAttribute('dir') || 'ltr' },
    h('div', { class: 'setup-nudge-mark' }, icon('flask', 'h-5 w-5')),
    h('div', { class: 'setup-nudge-text' },
      h('h2', { id: 'setup-nudge-title', class: 'font-medium', text: t(st.status === 'skipped' ? 'setup.nudge.skipped' : 'setup.nudge.title') }),
      h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400' }, t('setup.nudge.desc') + ' ', h('span', { class: 'tabular-nums', text: progress }))),
    h('div', { class: 'setup-nudge-bar', role: 'progressbar', 'aria-valuemin': '0', 'aria-valuemax': String(steps.length), 'aria-valuenow': String(done), 'aria-label': progress },
      h('div', { class: 'setup-progress-bar', style: 'width:' + Math.round(done / steps.length * 100) + '%' })),
    go);
  grid.insertBefore(card, grid.firstChild);
}

var AUDIT_LABELS = /^(base|economy|publication) (create|update|delete|invoke) /;

function auditLogSources() {
  if (!/^#\/settings\/audit-log/.test(location.hash || '')) return;
  var rows = document.querySelectorAll('#main table tbody tr');
  for (var i = 0; i < rows.length; i++) {
    var row = rows[i];
    var msg = row.getAttribute('data-nk-audit');
    if (msg === null) {
      var cells = row.cells;
      msg = '';
      for (var c = 0; c < cells.length; c++) {
        var txt = cells[c].textContent.replace(/[⁦-⁩]/g, '').trim();
        if (AUDIT_LABELS.test(txt) || /^(پایه|اقتصاد|انتشار) · /.test(txt)) { msg = 'module'; break; }
      }
      row.setAttribute('data-nk-audit', msg);
    }
    if (msg !== 'module') continue;
    var walker = document.createTreeWalker(row, NodeFilter.SHOW_TEXT);
    var n;
    while ((n = walker.nextNode())) {
      var v = n.nodeValue.trim();
      if (v === 'Hiro Economy' || v === 'اقتصاد Hiro') n.nodeValue = n.nodeValue.replace(v, t('section'));
    }
  }
}

var SATORI = { state: null, promise: null };

function satoriState() {
  if (SATORI.state !== null) return Promise.resolve(SATORI.state);
  if (SATORI.promise) return SATORI.promise;
  var tk = token();
  if (!tk) return Promise.resolve(true);
  SATORI.promise = fetch(CONSOLE_BASE + '/extensions', { headers: { Accept: 'application/json', Authorization: 'Bearer ' + tk } }).then(function (res) { return res.ok ? res.json() : null; }).then(function (ext) {
    SATORI.promise = null;
    SATORI.state = ext ? !!ext.satori : true;
    return SATORI.state;
  }, function () { SATORI.promise = null; return true; });
  return SATORI.promise;
}

function satoriGuard(to) {
  if (!to || !/^\/players\/[^/]+\/satori(\/|$)/.test(to.path)) return true;
  return satoriState().then(function (on) { return on ? true : to.path.replace(/\/satori(\/.*)?$/, ''); });
}

function satoriRoutes() {
  if (!/^#\/players\/[^/?]+\/satori(\/|\?|$)/.test(location.hash || '')) return;
  satoriState().then(function (on) {
    if (on || !router) return;
    var m = /^#(\/players\/[^/?]+)\/satori/.exec(location.hash);
    if (m) router.replace(m[1]);
  });
}

var NOT_FOUND = { last: '', at: 0 };

function notFoundHook(arg) {
  var path = null;
  if (arg && arg.newURL) path = (new URL(arg.newURL).hash || '#/').replace(/^#/, '').split('?')[0];
  else if (arg && arg.redirectedFrom) path = arg.redirectedFrom.path;
  else path = (location.hash || '#/').replace(/^#/, '').split('?')[0];
  checkNotFound(path || '/');
}

function checkNotFound(path) {
  if (!router || !token()) return;
  var r;
  try { r = router.resolve(path); } catch (e) { return; }
  if (!r || r.name !== '404') return;
  if (NOT_FOUND.last === path && Date.now() - NOT_FOUND.at < 1500) return;
  NOT_FOUND.last = path;
  NOT_FOUND.at = Date.now();
  toast('info', t('notFound', { path: '\u2068' + path + '\u2069' }));
  setTimeout(function () {
    var now = (location.hash || '#/').replace(/^#/, '').split('?')[0];
    if (now === path) history.replaceState(history.state, '', location.pathname + location.search + '#/');
  }, 0);
}

window.addEventListener('hashchange', notFoundHook);

function nativeLabels() {
  var acct = document.querySelector('#main > div.sticky:first-child .justify-end > .x-popover > button.x-button');
  if (acct && acct.getAttribute('aria-label') !== t('account')) {
    acct.setAttribute('aria-label', t('account'));
    acct.setAttribute('title', t('account'));
  }
  if (document.getElementById('main')) return;
  var links = document.querySelectorAll('a[href="/"]');
  for (var i = 0; i < links.length; i++) {
    var a = links[i];
    if (!a.querySelector('img')) continue;
    var label = brandName() || 'Nakama';
    if (a.getAttribute('aria-label') !== label) a.setAttribute('aria-label', label);
    if (a.getAttribute('tabindex') !== '-1') a.setAttribute('tabindex', '-1');
  }
}

function nativeTweaks() {
  nativeLabels();
  syncCharts();
  syncMonaco();
  dashboardSetup();
  auditLogSources();
  satoriRoutes();
  playerModCard();
  playerProgressCard();
  playerActivityCard();
  playerTasksCard();
  playerItemsCard();
  playerAchievementsCard();
}

var ROOT_ID = 'economy-root';
var NAV_LINK = 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent dark:border-transparent';
var NAV_ACTIVE = '!text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700';
var rendered = null;
var router = null;
var pendingHash = null;
var EmptyPage = { name: 'EconomyPage', render: function () { return null; } };

function captureRouter(app) {
  if (router) return;
  try {
    var r = app && app.config && app.config.globalProperties && app.config.globalProperties.$router;
    if (!r || typeof r.addRoute !== 'function') return;
    var layout = r.resolve('/').matched[0];
    if (!layout || !layout.components) return;
    SECTION_ORDER.forEach(function (sec) {
      if (r.hasRoute(sec)) return;
      r.addRoute({ path: '/' + sec + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: sec, component: EmptyPage, meta: { get title() { return sectionTitle(sec); } } }] });
    });
    if (!r.hasRoute(PS_ROUTE)) r.addRoute({ path: '/' + PS_ROUTE + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: PS_ROUTE, component: EmptyPage, meta: { get title() { return psTitle(); } } }] });
    Object.keys(PS_LEGACY).forEach(function (from) {
      var name = 'legacy-' + from.replace('/', '-');
      if (!r.hasRoute(name)) r.addRoute({ path: '/' + from + '/:rest(.*)*', name: name, redirect: function (to) { return to.fullPath.replace('/' + from, '/' + PS_ROUTE + '/' + PS_LEGACY[from]); } });
    });
    router = r;
    r.beforeEach(satoriGuard);
    r.afterEach(function (to, from) { fireRoute(to, from); });
  } catch (e) {
    router = null;
    return;
  }
  try { window.dispatchEvent(new CustomEvent('nakama-console:ready')); } catch (e) { router = router || null; }
  if (pendingHash) {
    var target = pendingHash;
    pendingHash = null;
    router.replace(target.slice(1));
  }
}

function sectionTitle(sec) {
  var r = parseRoute(location.hash);
  var page = r && r.section === sec ? r.page : null;
  return t(SECTION_LABEL[sec]) + (page ? ' - ' + t('nav.' + page) : '');
}

function sectionPageTitle(v) {
  var r = parseRoute(location.hash);
  if (!r || r.section === PS_ROUTE || !SECTIONS[r.section]) return v;
  var parts = String(v == null ? '' : v).split(' | ');
  if (parts.length < 2) return v;
  parts[0] = sectionTitle(r.section);
  return parts.join(' | ');
}

function psTitle() {
  var r = parseRoute(location.hash);
  var sec = r && psSection(r.page);
  return t('ps.title') + (sec ? ' - ' + t(sec.label) : '');
}

function legacyHash(hash) {
  var m = /^#\/([^/?]+\/[^/?]+)(.*)$/.exec(hash || '');
  return m && PS_LEGACY[m[1]] ? '#/' + PS_ROUTE + '/' + PS_LEGACY[m[1]] + m[2] : null;
}

function trapApp() {
  var proto = Element.prototype;
  if (Object.getOwnPropertyDescriptor(proto, '__vue_app__')) return;
  Object.defineProperty(proto, '__vue_app__', {
    configurable: true,
    get: function () { return undefined; },
    set: function (app) {
      Object.defineProperty(this, '__vue_app__', { value: app, writable: true, configurable: true });
      delete proto.__vue_app__;
      captureRouter(app);
    }
  });
}

function replaceHash(hash) {
  var r = parseRoute(hash);
  if (r && r.section === PS_ROUTE) { if (ps && ps.ctl) ps.ctl.key = keyOf(r); }
  else if (rendered) rendered.key = keyOf(r);
  if (router) router.replace(hash.slice(1));
  else history.replaceState(history.state, '', location.pathname + location.search + hash);
}

function parseRoute(hash) {
  var m = /^#\/(economy|publication|moderation|support|project-settings)(?:\/([^/?]*))?(?:\/([^?]*))?(?:\?(.*))?$/.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  var section = m[1];
  var page;
  if (section === PS_ROUTE) page = psSection(m[2]) ? m[2] : null;
  else {
    page = m[2] || SECTIONS[section][0];
    if (SECTIONS[section].indexOf(page) === -1) page = SECTIONS[section][0];
  }
  var query = {};
  (m[4] || '').split('&').forEach(function (pair) {
    if (!pair) return;
    var i = pair.indexOf('=');
    var k = decodeURIComponent(i === -1 ? pair : pair.slice(0, i));
    var v = i === -1 ? '' : decodeURIComponent(pair.slice(i + 1).replace(/\+/g, ' '));
    query[k] = v;
  });
  return { section: section, page: page, id: m[3] ? decodeURIComponent(m[3]) : null, query: query };
}

function keyOf(r) {
  if (!r) return '';
  var detail = (r.page === 'payments' || r.page === 'notifications' || r.page === 'reports' || r.page === 'tickets' || r.page === 'live-chat' || r.page === 'remote-config' || r.page === 'referrals') && r.id ? '/' + r.id : '';
  return lang() + '|' + r.page + detail + '|' + JSON.stringify(r.query);
}

function navPages(section) {
  return SECTIONS[section].filter(function (p) { var need = PAGE_ACL[p] || SECTION_ACL[section]; return !need || aclCan(need, 0); });
}

function buildNav(section, collapsed, active) {
  var ul = h('ul', { class: 'space-y-0.5' });
  navPages(section).forEach(function (p) {
    var on = p === active;
    var a = h('a', { href: hashFor(p, null, p === active ? null : lastQuery[p]), class: (on ? NAV_ACTIVE + ' ' : '') + NAV_LINK, 'aria-current': on ? 'page' : null, title: collapsed ? t('nav.' + p) : null, 'data-economy-page': p });
    a.appendChild(collapsed ? h('div', { class: 'x-tooltip flex' }, icon(p, 'h-4 w-4')) : icon(p, 'h-4 w-4'));
    a.appendChild(h('div', { class: 'font-medium relative truncate' + (collapsed ? '' : ' flex-1') }, h('span', { text: collapsed ? ' ' : t('nav.' + p) })));
    var badge = navBadge(p, collapsed);
    if (badge) a.appendChild(badge);
    ul.appendChild(h('li', {}, a));
  });
  var title = SECTION_LABEL[section];
  return h('div', { 'data-economy-nav': section, 'data-l10n-skip': '', class: section === SECTION_ORDER[0] ? null : 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 ' + (collapsed ? 'text-[8px] text-center' : 'px-3 text-[10px]'), text: collapsed ? t(title + '.short') : t(title) }),
    h('nav', { 'aria-label': t(title) }, ul));
}

function placeNav(prev, section, collapsed, active, key) {
  var next = prev.nextElementSibling;
  var existing = next && next.getAttribute('data-economy-nav') === section ? next : null;
  if (!existing || existing.getAttribute('data-key') !== key) {
    var built = buildNav(section, collapsed, active);
    built.setAttribute('data-key', key);
    if (existing) existing.replaceWith(built);
    else prev.after(built);
    existing = built;
  }
  return existing;
}

function isCollapsed(anchor) {
  var drawer = anchor.closest('.x-drawer');
  if (drawer && drawer.getBoundingClientRect().width > 0 && drawer.getBoundingClientRect().width < 100) return true;
  var span = anchor.querySelector('span');
  return !!span && span.textContent === ' ';
}

function syncSidebar(route, psOn) {
  var active = route ? route.page : null;
  var anchors = [];
  var hrefs = ['#/players', '#/chat', '#/notifications', '#/groups', '#/storage', '#/'];
  for (var hi = 0; hi < hrefs.length && !anchors.length; hi++) {
    anchors = Array.prototype.filter.call(document.querySelectorAll('.x-drawer a[href="' + hrefs[hi] + '"]'), function (x) { return !!x.closest('nav') && !x.closest('[data-economy-nav]'); });
  }
  var keep = [];
  for (var i = 0; i < anchors.length; i++) {
    var a = anchors[i];
    var drawer = a.closest('.x-drawer');
    if (!drawer || a.closest('#' + ROOT_ID)) continue;
    var nav = a.closest('nav');
    var group = nav && nav.parentElement;
    if (!group || !group.parentElement) continue;
    var collapsed = isCollapsed(a);
    var key = lang() + '|' + collapsed + '|' + active + '|' + pendingComments + '|' + pendingReports + '|' + supportBadgeKey() + '|' + JSON.stringify(lastQuery) + '|' + token();
    var prev = group;
    for (var si = 0; si < SECTION_ORDER.length; si++) {
      if (!navPages(SECTION_ORDER[si]).length) continue;
      prev = placeNav(prev, SECTION_ORDER[si], collapsed, active, key);
      keep.push(prev);
    }
    var psNav = placePsNav(drawer, collapsed, !!psOn, lang() + '|' + collapsed + '|' + !!psOn + '|' + psEntryHash());
    if (psNav) keep.push(psNav);
    refreshBadge(false);
    refreshReportBadge(false);
    refreshSupportBadge(false);
    var vueActive = group.parentElement.querySelectorAll('a[aria-current="page"]:not([data-economy-page])');
    for (var j = 0; j < vueActive.length && (active || psOn); j++) {
      if (vueActive[j].getAttribute('href') === '#/') vueActive[j].setAttribute('data-economy-dimmed', '');
    }
  }
  var all = document.querySelectorAll('[data-economy-nav],[data-ps-nav]');
  for (var k = 0; k < all.length; k++) if (keep.indexOf(all[k]) === -1) all[k].remove();
  document.documentElement.toggleAttribute('data-economy-active', !!active || !!psOn);
}

function showMains() {
  var hidden = document.querySelectorAll('[data-economy-hidden]');
  for (var i = 0; i < hidden.length; i++) {
    hidden[i].style.display = '';
    hidden[i].removeAttribute('data-economy-hidden');
  }
}

function unmount() {
  var root = document.getElementById(ROOT_ID);
  if (root) root.remove();
  showMains();
  closePanel();
  rendered = null;
}

function mountPage(root, route, ctl) {
  var ps2 = psSection(route.page);
  var ctx = { id: route.id, query: route.query };
  if (ps2) { ctl.list = ps2.render(root, route, ctx) || null; return; }
  if (route.page === 'comments') refreshBadge(true);
  if (route.page === 'reports') refreshReportBadge(true);
  if (route.page === 'tickets' || route.page === 'live-chat') { refreshSupportBadge(true); supportBeat(); }
  if (route.page === 'overview') overviewPage(root);
  else if (route.page === 'tickets' && route.id) ticketDetail(root, route.id);
  else if (route.page === 'tickets') ctl.list = ticketsPage(root, route.query, ctx);
  else if (route.page === 'live-chat') liveChatPage(root, route.id, route.query);
  else if (route.page === 'activity') ctl.list = activityPage(root, route.query, ctx);
  else if (route.page === 'referrals') ctl.list = referralsPage(root, route.query, ctx);
  else if (route.page === 'remote-config' && route.id) configEditor(root, route.id, route.query);
  else if (route.page === 'remote-config') ctl.list = configPage(root, route.query, ctx);
  else if (route.page === 'reports' && route.id) reportDetail(root, route.id);
  else if (route.page === 'players') ctl.list = playersPage(root, route.query, ctx);
  else if (route.page === 'restrictions') ctl.list = restrictionsPage(root, route.query, ctx);
  else if (route.page === 'chat-words') ctl.list = chatWordsPage(root, route.query, ctx);
  else if (route.page === 'payments' && route.id) paymentDetail(root, route.id);
  else if (route.page === 'notifications' && route.id === 'new') composePage(root, route.query);
  else if (route.page === 'notifications' && route.id) notificationDetail(root, route.id);
  else ctl.list = listPage(root, route.page, route.query, ctx);
}

function pageCtl(prev, root, route, onNew) {
  var key = keyOf(route);
  var ctl = prev && prev.root === root ? prev : null;
  if (!ctl || ctl.key !== key) {
    closePanel();
    ctl = { root: root, key: key, id: undefined, list: null };
    onNew(ctl);
    mountPage(root, route, ctl);
  }
  if (RESOURCES[route.page] && ctl.id !== route.id) {
    ctl.id = route.id;
    var mine = ctl;
    if (route.id) openEditor(route.page, route.id, route.query, function () { if (mine.list) mine.list.reload(); });
    else closePanel();
  }
  return ctl;
}

function render(root, route) {
  pageCtl(rendered, root, route, function (ctl) {
    window.scrollTo(0, 0);
    var scroller = document.getElementById('main');
    if (scroller) scroller.scrollTop = 0;
    rendered = ctl;
  });
}

function sync() {
  dockTheme();
  applyBrand();
  maybeClaimSetup();
  nativeTweaks();
  var hash = location.hash;
  var moved = legacyHash(hash);
  if (moved) {
    hash = moved;
    if (router) router.replace(moved.slice(1));
    else history.replaceState(history.state, '', location.pathname + location.search + moved);
  }
  var route = parseRoute(hash);
  if (route && !LANGS.tried && token()) {
    LANGS.tried = true;
    loadLanguages().then(function () {
      if (!LANGS.loaded || LANGS.def === 'fa') return;
      if (rendered) rendered.key = '';
      if (ps && ps.ctl) ps.ctl.key = '';
      schedule();
    });
    loadKeyTypes();
  }
  if (route && route.section === PS_ROUTE) {
    syncSidebar(null, true);
    psShow(route);
    dockPsButton();
    return;
  }
  if (ps) psHide();
  psRemember(hash || '#/');
  syncSidebar(route, false);
  dockPsButton();
  if (!route) { if (rendered || document.getElementById(ROOT_ID)) unmount(); return; }
  var root = document.getElementById(ROOT_ID);
  var main = document.querySelector('#main > main');
  if (!main) return;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') {
      mains[i].style.display = 'none';
      mains[i].setAttribute('data-economy-hidden', '');
    }
  }
  if (!root) {
    root = h('div', { id: ROOT_ID, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-economy': 'root' });
    rendered = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  render(root, route);
}

var THEME_KEY = 'color-mode';
var THEME_ICONS = {
  moon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M21.5 14.1A9.5 9.5 0 0 1 9.9 2.5a9.5 9.5 0 1 0 11.6 11.6z"/></svg>',
  sun: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" aria-hidden="true"><circle cx="12" cy="12" r="4.5" fill="currentColor"/><path d="M12 1.5v2.5M12 20v2.5M4.6 4.6l1.8 1.8M17.6 17.6l1.8 1.8M1.5 12H4M20 12h2.5M4.6 19.4l1.8-1.8M17.6 6.4l1.8-1.8"/></svg>'
};
var themeButton = null;

function storedTheme() {
  try {
    var v = localStorage.getItem(THEME_KEY);
    if (v === 'dark' || v === 'light') return v;
  } catch (e) {}
  return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function isDarkTheme() {
  return document.documentElement.classList.contains('dark');
}

function paintThemeButton() {
  if (!themeButton) return;
  var dark = isDarkTheme();
  var label = t(dark ? 'theme.toLight' : 'theme.toDark');
  themeButton.innerHTML = dark ? THEME_ICONS.sun : THEME_ICONS.moon;
  themeButton.setAttribute('title', label);
  themeButton.setAttribute('aria-label', label);
  themeButton.setAttribute('aria-pressed', dark ? 'true' : 'false');
}

function toggleTheme() {
  var next = isDarkTheme() ? 'light' : 'dark';
  try { localStorage.setItem(THEME_KEY, next); } catch (e) {}
  document.documentElement.classList.toggle('dark', next === 'dark');
  paintThemeButton();
}

function dockTheme() {
  if (!themeButton) {
    themeButton = document.createElement('button');
    themeButton.type = 'button';
    themeButton.className = 'l10n-switcher economy-theme';
    themeButton.setAttribute('data-l10n-skip', '');
    themeButton.addEventListener('click', toggleTheme);
    paintThemeButton();
  }
  var switcher = document.querySelector('.l10n-switcher:not(.economy-theme)');
  if (switcher) {
    if (themeButton.nextElementSibling !== switcher) switcher.before(themeButton);
    themeButton.classList.toggle('l10n-switcher--docked', switcher.classList.contains('l10n-switcher--docked'));
  } else if (!themeButton.parentElement) {
    document.body.appendChild(themeButton);
  }
}

document.documentElement.classList.toggle('dark', storedTheme() === 'dark');

var scheduled = false;
function schedule() {
  if (scheduled) return;
  scheduled = true;
  requestAnimationFrame(function () {
    scheduled = false;
    sync();
  });
}

function boot() {
  var app = document.getElementById('app');
  if (!router && app && app.__vue_app__) captureRouter(app.__vue_app__);
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    var el = document.getElementById('app');
    if (!router && el && el.__vue_app__) captureRouter(el.__vue_app__);
    if (router || waited >= 5000) {
      clearInterval(poll);
      if (pendingHash) { location.hash = pendingHash; pendingHash = null; }
    }
  }, 100);
  window.addEventListener('hashchange', sync);
  window.addEventListener('popstate', schedule);
  new MutationObserver(schedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { if (rendered) rendered.key = ''; if (ps) { ps.listKey = null; if (ps.ctl) ps.ctl.key = ''; } paintThemeButton(); schedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  sync();
}

var routeHooks = [notFoundHook];

function fireRoute(to, from) {
  for (var i = 0; i < routeHooks.length; i++) {
    try { routeHooks[i](to, from); } catch (e) { continue; }
  }
}

function consoleSections() {
  return SECTION_ORDER.filter(function (sec) { return navPages(sec).length > 0; }).map(function (sec) {
    return {
      id: sec,
      label: t(SECTION_LABEL[sec]),
      pages: navPages(sec).map(function (p) { return { id: p, href: hashFor(p), label: t('nav.' + p), badge: p === 'comments' ? pendingComments || 0 : p === 'reports' ? pendingReports || 0 : supportBadge(p) }; })
    };
  });
}

window.NakamaConsole = {
  version: 1,
  get router() { return router; },
  lang: function () { return lang(); },
  t: function (key, vars) { return t(key, vars); },
  sections: consoleSections,
  badges: function () { return { comments: pendingComments || 0, reports: pendingReports || 0, tickets: supportBadge('tickets'), chats: supportBadge('live-chat') }; },
  onRoute: function (fn) {
    if (typeof fn !== 'function') return function () {};
    routeHooks.push(fn);
    return function () { var i = routeHooks.indexOf(fn); if (i !== -1) routeHooks.splice(i, 1); };
  },
  projectSettings: {
    allowed: function () { return !!token() && psAllowed().length > 0; },
    href: function () { return psEntryHash(); },
    isOpen: function () { return !!ps; },
    open: function () { if (!ps) psNavigate(psEntryHash(), false); },
    sections: function () { return psAllowed().map(function (s) { return { id: s.id, href: hashFor(s.id), label: t(s.label) }; }); }
  },
  brand: function () {
    var b = BRAND.data;
    if (!b) return null;
    return { name: brandName(), shortName: b.shortName || '', markUrl: b.markUrl || '', logoUrl: b.logoUrl || '', logoDarkUrl: b.logoDarkUrl || '' };
  },
  toggleTheme: function () { toggleTheme(); },
  isDark: function () { return isDarkTheme(); }
};

if (legacyHash(location.hash) || /^#\/(economy|publication|moderation|support|project-settings)(\/|\?|$)/.test(location.hash)) {
  pendingHash = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
trapApp();
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot);
else boot();
})();
