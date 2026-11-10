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
    'setup.meta.opened': 'اولین باز شدن',
    'setup.meta.completed': 'پایان',
    'tab.countries': 'کشورها',
    'tab.accounts': 'قوانین حساب',
    'tab.otp': 'ورود با پیامک',
    'tab.app-versions': 'نسخه‌های برنامه',
    'desc.countries': 'کشورهایی که بازیکن می‌تواند انتخاب کند و شماره‌اش را با پیش‌شماره‌ی آن‌ها وارد کند.',
    'desc.app-versions': 'آخرین نسخه و کمترین نسخه‌ی مجاز هر پلتفرم. نسخه‌های قدیمی‌تر از کمترین نسخه باید به‌روز شوند.',
    'desc.accounts': 'قانون‌های ساختن حساب. سرور همین‌ها را برای نام کاربری، رمز و نام نمایشی چک می‌کند.',
    'desc.otp': 'آخرین درخواست‌های کد. خود کد هیچ‌جا نشان داده نمی‌شود.',
    'new.countries': 'کشور تازه',
    'edit.countries': 'ویرایش کشور',
    'new.app-versions': 'نسخه‌ی تازه',
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
    'av.link.hint': 'لینک https یا لینک خود فروشگاه (bazaar:// ، myket:// ، market://). نام نمایشی اختیاری است.',
    'f.av.platform': 'پلتفرم',
    'f.av.active.hint': 'خاموش باشد، بازی همیشه «همه‌چیز درست است» می‌گیرد.',
    'f.av.latest': 'آخرین نسخه',
    'f.av.latest.hint': 'نسخه‌های قدیمی‌تر پیام «نسخه‌ی تازه آمده» می‌گیرند.',
    'f.av.min': 'کمترین نسخه',
    'f.av.min.hint': 'نسخه‌های قدیمی‌تر از این باید به‌روز شوند (به‌روزرسانی اجباری).',
    'f.av.stores': 'لینک فروشگاه‌ها',
    'f.av.messages': 'پیام‌ها',
    'f.av.updateMessage': 'پیام به‌روزرسانی',
    'f.av.updateMessage.hint': 'وقتی نسخه‌ی تازه‌تری هست ولی اجباری نیست.',
    'f.av.forceMessage': 'پیام به‌روزرسانی اجباری',
    'f.av.forceMessage.hint': 'وقتی نسخه‌ی بازیکن از کمترین نسخه قدیمی‌تر است.',
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
    'f.otp.resend': 'فاصله‌ی ارسال دوباره (ثانیه)',
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
    'code.PASSWORD_NEEDS_DIGIT': 'رمز باید عدد داشته باشد.'
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
    'tab.payment-methods': 'Payment Gateways',
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

    'new.currencies': 'New Currency',
    'new.metas': 'New Meta',
    'new.prizes': 'New Prize',
    'new.products': 'New Product',
    'new.gifts': 'New Gift',
    'new.categories': 'New Category',
    'new.tags': 'New Tag',
    'edit.currencies': 'Edit Currency',
    'edit.metas': 'Edit Meta',
    'edit.prizes': 'Edit Prize',
    'edit.products': 'Edit Product',
    'edit.gifts': 'Edit Gift',
    'edit.categories': 'Edit Category',
    'edit.tags': 'Edit Tag',

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
    'av.link.hint': 'An https link or the store\'s own link (bazaar://, myket://, market://). The display name is optional.',
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
    'code.PASSWORD_NEEDS_DIGIT': 'The password needs a digit.'
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
  return t('err.generic') + (err.code ? ' (' + err.code + ')' : '');
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
  first: '<path d="M11 17l-5-5 5-5M18 17l-5-5 5-5"/>',
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
  versions: '<rect x="5" y="2" width="14" height="20" rx="2"/><path d="M12 7v7M9 11l3 3 3-3M10 18h4"/>'
};
var MIRROR = { first: 1, prev: 1, next: 1, back: 1, reply: 1 };

function icon(name, cls) {
  var el = document.createElementNS(SVG_NS, 'svg');
  el.setAttribute('xmlns', SVG_NS);
  el.setAttribute('class', 'shrink-0 x-icon inline shrink-0 stroke-2 ' + (cls || 'h-4 w-4'));
  el.setAttribute('stroke-linejoin', 'round');
  el.setAttribute('stroke-linecap', 'round');
  el.setAttribute('stroke', 'currentColor');
  el.setAttribute('fill', 'none');
  el.setAttribute('viewBox', '0 0 24 24');
  el.setAttribute('aria-hidden', 'true');
  el.setAttribute('data-l10n-icon', MIRROR[name] ? 'mirror' : '');
  el.innerHTML = ICONS[name] || '';
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
  var el = h('input', { type: 'datetime-local', class: INPUT_MD + ' tabular-nums', dir: 'ltr' });
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

var ITEM_ACTIONS = { currency: ['add'], meta: ['add', 'set'], storage: ['give'] };

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
      var typeSel = selectInput(row.type, ['currency', 'meta', 'storage'].map(function (k) { return { value: k, label: t('item.' + k) }; }), function (v) {
        row.type = v;
        row.key = '';
        row.action = (ITEM_ACTIONS[v] || ['add'])[0];
        row.value = v === 'storage' ? {} : 0;
        render();
        emit();
      }, { placeholder: false, label: t('f.item.type') });
      var keyCtl;
      if (row.type === 'storage') {
        keyCtl = textInput(row.key, { ltr: true, mono: true, label: t('f.item.key'), oninput: function (v) { row.key = v.trim(); emit(); } });
      } else {
        keyCtl = refPicker({ resource: row.type === 'meta' ? 'metas' : 'currencies', valueKey: 'key', value: row.key, required: true, label: t('f.item.key'), onchange: function (v) { row.key = v || ''; emit(); } });
      }
      var actSel = selectInput(row.action, (ITEM_ACTIONS[row.type] || ['add']).map(function (k) { return { value: k, label: t('action.' + k) }; }), function (v) { row.action = v; emit(); }, { placeholder: false, label: t('f.item.action') });
      var valCtl;
      if (row.type === 'storage') {
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
      else if (r.type !== 'storage' && (r.value === null || r.value === undefined || isNaN(r.value))) msg = t('f.item.value') + ': ' + t('err.number');
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
      var visible = !e.def.when || e.def.when(values);
      var el = e.row ? e.el : e.wrap;
      el.hidden = !visible;
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
      } else if (t2 === 'items' || t2 === 'actions' || t2 === 'links') {
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
    if (best.def.type === 'actions' && best.control.showServerError) return best.control.showServerError(err);
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
var CATEGORY_KEY_TYPES = ['product', 'prize', 'gift', 'currency', 'faq', 'news'];
var TAG_KEY_TYPES = ['product', 'prize', 'news', 'faq', 'general', 'user'];
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

function numCell(v) {
  return h('span', { class: 'tabular-nums', text: v === null || v === undefined ? '' : fmtNum(v) });
}

function dateCell(v) {
  return h('span', { class: 'text-secondary-600 dark:text-secondary-300 whitespace-nowrap', text: fmtDate(v) });
}

function activeTag(on) {
  return tag(on ? t('active') : t('inactive'), on ? 'success' : 'secondary');
}

function priceText(it) {
  var parts = [];
  if (it.priceToman !== null && it.priceToman !== undefined) parts.push(t('toman', { n: it.priceToman }));
  if (it.coinPrice && it.coinPrice.amount !== undefined && it.coinPrice.amount !== null) parts.push(fmtNum(it.coinPrice.amount) + ' ' + (it.coinPrice.currencyKey || ''));
  return parts;
}

function itemSummary(items) {
  return (items || []).map(function (r) {
    if (r.type === 'storage') return r.key;
    var sign = r.action === 'add' && Number(r.value) > 0 ? '+' : '';
    return (r.action === 'set' ? '= ' : sign) + fmtNum(r.value) + ' ' + r.key;
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
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '60%' },
      { label: 'col.type', render: function (it) { return tag(t('type.' + it.type), it.type === 'hard' ? 'warning' : 'primary'); }, width: 110 },
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
      { type: 'row', fields: [
        { name: 'defaultValue', type: 'number', label: 'f.defaultValue', hint: 'f.defaultValue.currency.hint', min: 0 },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function () { return { type: 'soft', name: {}, defaultValue: 0, sort: 0, metadata: {} }; }
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
        var s = itemSummary(it.items);
        return h('div', { class: 'flex flex-wrap gap-1' }, s.slice(0, 3).map(function (x) { return tag(x, 'secondary'); }), s.length > 3 ? tag('+' + fmtNum(s.length - 3), 'secondary') : null);
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
      { label: 'col.price', render: function (it) { return h('div', { class: 'flex flex-col tabular-nums whitespace-nowrap' }, priceText(it).map(function (p) { return h('span', { text: p }); }), it.discountPercent ? h('span', { class: 'text-xs text-success-600', text: t('percent', { n: it.discountPercent }) }) : null); }, width: 160 },
      { label: 'col.methods', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, (it.paymentMethods || []).map(function (m) { return tag(t('method.' + m), 'secondary'); })); }, skeleton: '45%' },
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
    defaults: function () { return { isActive: true, name: {}, description: {}, tagIds: [], paymentMethods: ['zarinpal'], discountPercent: 0, sku: {}, coinPrice: null, isSubscription: false, minLevel: 0, sort: 0, metadata: {} }; },
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
    filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
    columns: [
      { label: 'col.name', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) }), it.code ? h('div', { class: 'text-secondary-500' }, mono(it.code)) : null); }, skeleton: '50%' },
      { label: 'col.type', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('gift.' + it.type), it.type === 'discount' ? 'warning' : 'primary'), it.type === 'discount' && it.discountPercent ? tag(t('percent', { n: it.discountPercent }), 'secondary') : null); }, width: 150 },
      { label: 'col.usage', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: fmtNum(it.totalUsed || 0) + ' / ' + (it.maxTotalUses ? fmtNum(it.maxTotalUses) : t('unlimited')) }); }, width: 130 },
      { label: 'col.expire', render: function (it) { return dateCell(it.expireAt); }, width: 190 },
      { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
    ],
    fields: [
      tr('name', { label: 'f.gift.name', required: true, max: 128 }),
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' },
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
    defaults: function () { return { name: {}, type: 'reward', isActive: true, isPromo: false, perUserLimit: 1, maxTotalUses: null, code: '' }; },
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
      { label: 'col.keyType', render: function (it) { return tag(keyTypeLabel('tags', it.keyType), it.keyType === 'general' ? 'primary' : 'secondary'); }, width: 140 },
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
      { type: 'row', fields: [
        { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function (query) { return { keyType: (query && query.keyType) || 'general', name: {}, description: {}, iconUrl: '', metadata: {}, sort: 0 }; },
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
  economy: ['overview', 'currencies', 'metas', 'prizes', 'products', 'gifts', 'payments'],
  publication: ['faqs', 'news', 'comments', 'notifications']
};
var SECTION_ORDER = ['economy', 'publication'];
var LIMIT = 20;
var lastQuery = { comments: { status: 'pending' } };

function sectionOf(page) {
  if (psSection(page)) return PS_ROUTE;
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
      var inp = h('input', { class: INPUT_SM + ' font-mono eco-filter-text', dir: 'ltr', type: 'text', value: state.filters[f.name] || '', placeholder: t(f.label), 'aria-label': t(f.label) });
      inp.addEventListener('input', debounce(function () { state.filters[f.name] = toLatinDigits(inp.value).trim(); state.page = 1; sync(); load(); }, 300));
      return h('label', { class: 'x-label block relative x-input', style: INPUT_VARS }, inp);
    }
    return filterSelect(t(f.label), state.filters[f.name], opts2, function (v) { state.filters[f.name] = v; state.page = 1; sync(); load(); });
  });
  frame.content.appendChild(h('div', {},
    h('div', { class: 'flex flex-col items-start pb-2 gap-2' },
      h('div', { class: 'flex flex-col lg:flex-row lg:items-center gap-2 w-full' },
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
  function open(item) { go(name, item.id, currentQuery()); }
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
    return call('GET', '/' + name + qs(params)).then(function (res) {
      if (my !== seq) return;
      state.loading = false;
      state.loaded = true;
      state.total = (res && res.total) || 0;
      var items = (res && res.items) || [];
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
  var item = isNew ? Promise.resolve(def.defaults(query)) : call('GET', '/' + name + '/' + encodeURIComponent(id));
  Promise.all([item, freshLanguages(), loadKeyTypes()]).then(function (all) { return all[0]; }).then(function (data) {
    if (!panelState || panelState.el !== el) return;
    var values = clone(data) || {};
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
        call('DELETE', '/' + name + '/' + encodeURIComponent(id)).then(function () {
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
      call(isNew ? 'POST' : 'PUT', '/' + name + (isNew ? '' : '/' + encodeURIComponent(id)), payload).then(function () {
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
  if (page !== 'comments' || !pendingComments) return null;
  if (collapsed) return h('span', { class: 'eco-nav-dot', 'aria-label': t('cm.pendingN', { n: pendingComments }) });
  return h('span', { class: 'eco-nav-badge tabular-nums', 'aria-label': t('cm.pendingN', { n: pendingComments }), text: fmtNum(pendingComments) });
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
    target: { mode: 'users', userIds: query.userId ? [query.userId] : [], groupId: '' }
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
    var modeSeg = segmented([{ value: 'users', label: t('tg.users'), icon: 'users' }, { value: 'group', label: t('tg.group'), icon: 'categories' }, { value: 'all', label: t('tg.all'), icon: 'notifications' }], v.target.mode, function (m) {
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
      if ((tg.mode === 'users' && !tg.userIds.length) || (tg.mode === 'group' && !tg.groupId)) { preview = null; paintCount(); return; }
      pubApi('GET', '/notifications/preview-target' + qs({ mode: tg.mode, groupId: tg.mode === 'group' ? tg.groupId : '', userIds: tg.mode === 'users' ? tg.userIds.join(',') : '' })).then(function (res) {
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
          var map = { title: fields.title, body: fields.body, action: fields.action, 'action.label': fields.actionLabel, 'target.userIds': fields.target, 'target.groupId': fields.target, providerId: fields.provider };
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
    if (v.target.userIds.length) refreshPreview();
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
        h('div', { class: 'eco-actions-row' },
          cell('av.link.store', storeSel),
          cell('av.link.url', url),
          h('div', { class: 'min-w-0 eco-actions-label' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('av.link.label') }), label),
          h('div', { class: 'eco-items-remove' }, button({ variant: 'danger', size: 'iconLg', icon: 'trash', title: t('remove'), onclick: function () { rows.splice(idx, 1); render(); emit(); } }))),
        h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1.5', text: t('av.link.hint') }),
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

var SP_INK_LIGHT_BG = [223, 228, 238];
var SP_INK_DARK_BG = [74, 77, 95];

function pickShade(p, shades, bg) {
  for (var i = 0; i < shades.length; i++) if (contrastRatio(p[shades[i]], bg) >= 4.5) return shades[i];
  return shades[shades.length - 1];
}

function spatialAccent(hex) {
  var p = brandPalette(hex);
  var a = pickShade(p, [500, 600, 700, 800, 900], [255, 255, 255]);
  var ink = pickShade(p, [600, 700, 800, 900, 950], SP_INK_LIGHT_BG);
  var inkDark = pickShade(p, [300, 200, 100, 50], SP_INK_DARK_BG);
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
    '--sp-accent:rgb(' + sp.accent.join(' ') + ');--sp-accent-ink:rgb(' + sp.ink.join(' ') + ');--sp-accent-ink-dark:rgb(' + sp.inkDark.join(' ') + ');}';
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
    set: function (v) { nativeTitle.set.call(document, brandTitle(v)); }
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
    var num = form.el.querySelector('input[inputmode="numeric"]');
    if (num) {
      num.setAttribute('dir', 'ltr');
      num.value = toLatinDigits(num.value);
      num.addEventListener('input', function () { latinize(num); });
      num.addEventListener('blur', function () { num.value = toLatinDigits(num.value); });
    }
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
      var body = current ? { key: current.key, type: current.type, name: current.name, defaultValue: current.defaultValue, iconUrl: current.iconUrl, metadata: current.metadata || {}, sort: current.sort || 0 } : { metadata: {}, sort: 0 };
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
            h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 truncate', text: t('setup.step.' + k + '.desc') })),
          h('div', { class: 'flex items-center gap-2 shrink-0' }, tag(t('setup.st.' + s), STEP_COLOR[s]),
            button({ variant: 'link', size: 'sm', label: t('setup.rv.edit'), title: t('setup.rv.editStep', { step: t('setup.step.' + k) }), onclick: function () { openFromPage(k); } }))));
      });
      var meta = [];
      if (st.firstOpenedAt) meta.push([t('setup.meta.opened'), (st.firstOpenedBy || '') + ' · ' + fmtDate(st.firstOpenedAt)]);
      if (st.completedAt) meta.push([t('setup.meta.completed'), (st.completedBy || '') + ' · ' + fmtDate(st.completedAt)]);
      clear(body).appendChild(card(h('div', { class: 'p-4' },
        h('div', { class: 'flex flex-wrap items-center justify-between gap-3 mb-3' },
          h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'text-lg font-medium', text: t('setup.title') }), tag(t('setup.status.' + st.status), SETUP_STATUS_COLOR[st.status] || 'secondary')),
          h('span', { class: 'text-sm text-secondary-500 tabular-nums', text: t('setup.progress', { n: localDigits(finished), total: localDigits(steps.length) }) })),
        h('div', { class: 'setup-progress', role: 'progressbar', 'aria-valuemin': '0', 'aria-valuemax': String(steps.length), 'aria-valuenow': String(finished), 'aria-label': t('setup.progress', { n: localDigits(finished), total: localDigits(steps.length) }) },
          h('div', { class: 'setup-progress-bar', style: 'width:' + Math.round(finished / steps.length * 100) + '%' })),
        meta.length ? h('dl', { class: 'eco-dl mt-4 mb-2' }, meta.map(function (m) { return [h('dt', { class: 'text-secondary-500', text: m[0] }), h('dd', { text: m[1] })]; })) : null,
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
var ACL_RESOURCES = { NOTIFICATION: 20, SETTINGS: 22, HIRO_ECONOMY: 27 };
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

var ROOT_ID = 'economy-root';
var NAV_LINK = 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent dark:border-transparent';
var NAV_ACTIVE = '!text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700';
var rendered = null;
var router = null;
var pendingHash = null;
var EmptyPage = { name: 'EconomyPage', render: function () { return null; } };
var SECTION_TITLE = { economy: 'Economy', publication: 'Publishing' };

function captureRouter(app) {
  if (router) return;
  try {
    var r = app && app.config && app.config.globalProperties && app.config.globalProperties.$router;
    if (!r || typeof r.addRoute !== 'function') return;
    var layout = r.resolve('/').matched[0];
    if (!layout || !layout.components) return;
    SECTION_ORDER.forEach(function (sec) {
      if (r.hasRoute(sec)) return;
      r.addRoute({ path: '/' + sec + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: sec, component: EmptyPage, meta: { title: SECTION_TITLE[sec] } }] });
    });
    if (!r.hasRoute(PS_ROUTE)) r.addRoute({ path: '/' + PS_ROUTE + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: PS_ROUTE, component: EmptyPage, meta: { get title() { return psTitle(); } } }] });
    Object.keys(PS_LEGACY).forEach(function (from) {
      var name = 'legacy-' + from.replace('/', '-');
      if (!r.hasRoute(name)) r.addRoute({ path: '/' + from + '/:rest(.*)*', name: name, redirect: function (to) { return to.fullPath.replace('/' + from, '/' + PS_ROUTE + '/' + PS_LEGACY[from]); } });
    });
    router = r;
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
  var m = /^#\/(economy|publication|project-settings)(?:\/([^/?]*))?(?:\/([^?]*))?(?:\?(.*))?$/.exec(hash === undefined ? location.hash : hash);
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
  var detail = (r.page === 'payments' || r.page === 'notifications') && r.id ? '/' + r.id : '';
  return lang() + '|' + r.page + detail + '|' + JSON.stringify(r.query);
}

function buildNav(section, collapsed, active) {
  var ul = h('ul', { class: 'space-y-0.5' });
  SECTIONS[section].forEach(function (p) {
    var on = p === active;
    var a = h('a', { href: hashFor(p, null, p === active ? null : lastQuery[p]), class: (on ? NAV_ACTIVE + ' ' : '') + NAV_LINK, 'aria-current': on ? 'page' : null, title: collapsed ? t('nav.' + p) : null, 'data-economy-page': p });
    a.appendChild(collapsed ? h('div', { class: 'x-tooltip flex' }, icon(p, 'h-4 w-4')) : icon(p, 'h-4 w-4'));
    a.appendChild(h('div', { class: 'font-medium relative truncate' + (collapsed ? '' : ' flex-1') }, h('span', { text: collapsed ? ' ' : t('nav.' + p) })));
    var badge = navBadge(p, collapsed);
    if (badge) a.appendChild(badge);
    ul.appendChild(h('li', {}, a));
  });
  var title = section === 'economy' ? 'section' : 'section.pub';
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
  var anchors = document.querySelectorAll('a[href="#/players"]');
  var keep = [];
  for (var i = 0; i < anchors.length; i++) {
    var a = anchors[i];
    var drawer = a.closest('.x-drawer');
    if (!drawer || a.closest('#' + ROOT_ID)) continue;
    var nav = a.closest('nav');
    var group = nav && nav.parentElement;
    if (!group || !group.parentElement) continue;
    var collapsed = isCollapsed(a);
    var key = lang() + '|' + collapsed + '|' + active + '|' + pendingComments + '|' + JSON.stringify(lastQuery);
    var prev = group;
    for (var si = 0; si < SECTION_ORDER.length; si++) {
      prev = placeNav(prev, SECTION_ORDER[si], collapsed, active, key);
      keep.push(prev);
    }
    var psNav = placePsNav(drawer, collapsed, !!psOn, lang() + '|' + collapsed + '|' + !!psOn + '|' + psEntryHash());
    if (psNav) keep.push(psNav);
    refreshBadge(false);
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
  if (route.page === 'overview') overviewPage(root);
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

var routeHooks = [];

function fireRoute(to, from) {
  for (var i = 0; i < routeHooks.length; i++) {
    try { routeHooks[i](to, from); } catch (e) { continue; }
  }
}

function consoleSections() {
  return SECTION_ORDER.map(function (sec) {
    return {
      id: sec,
      label: t(sec === 'economy' ? 'section' : 'section.pub'),
      pages: SECTIONS[sec].map(function (p) { return { id: p, href: hashFor(p), label: t('nav.' + p), badge: p === 'comments' ? pendingComments || 0 : 0 }; })
    };
  });
}

window.NakamaConsole = {
  version: 1,
  get router() { return router; },
  lang: function () { return lang(); },
  t: function (key, vars) { return t(key, vars); },
  sections: consoleSections,
  badges: function () { return { comments: pendingComments || 0 }; },
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

if (legacyHash(location.hash) || /^#\/(economy|publication|project-settings)(\/|\?|$)/.test(location.hash)) {
  pendingHash = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
trapApp();
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot);
else boot();
})();
