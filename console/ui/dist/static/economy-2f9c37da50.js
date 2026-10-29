(function () {
'use strict';
var MESSAGES = {
  fa: {
    'section': 'اقتصاد',
    'section.short': 'اقتصاد',
    'nav.overview': 'نمای کلی',
    'nav.currencies': 'ارزها',
    'nav.metas': 'متاها',
    'nav.prizes': 'جایزه‌ها',
    'nav.products': 'فروشگاه',
    'nav.gifts': 'هدیه‌ها',
    'nav.payments': 'پرداخت‌ها',
    'nav.payment-methods': 'روش‌های پرداخت',
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
    'desc.categories': 'دسته‌بندی محصولات، جایزه‌ها، هدیه‌ها و ارزها.',
    'desc.tags': 'برچسب‌هایی برای گروه‌بندی محصولات و جایزه‌ها.',

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
    'err.notfound': 'این بخش روی سرور فعال نیست. ماژول اقتصاد بارگذاری شده است؟',
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
    'code.FA_REQUIRED': 'نام فارسی لازم است.',
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
    'code.CATEGORY_KIND_MISMATCH': 'این دسته‌بندی برای این بخش نیست.',
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
    'code.TEXT_TOO_LONG': 'متن بیش از حد بلند است.',
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
    'f.name': 'نام',
    'f.name.fa': 'نام فارسی',
    'f.name.en': 'نام انگلیسی',
    'f.message.fa': 'پیام فارسی',
    'f.message.en': 'پیام انگلیسی',
    'f.description.fa': 'توضیح فارسی',
    'f.description.en': 'توضیح انگلیسی',
    'f.type': 'نوع',
    'f.kind': 'کاربرد',
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
    'col.kind': 'کاربرد',
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
    'filter.kind': 'کاربرد',
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
    'ov.methods': 'روش‌های پرداخت',
    'ov.sandbox': 'آزمایشی',
    'ov.payments': 'پرداخت‌ها',
    'ov.revenue': 'درآمد',
    'ov.open': 'مشاهده',
    'ov.today': 'امروز',
    'ov.total': 'کل'
  },
  en: {
    'section': 'Economy',
    'section.short': 'Econ',
    'nav.overview': 'Overview',
    'nav.currencies': 'Currencies',
    'nav.metas': 'Metas',
    'nav.prizes': 'Prizes',
    'nav.products': 'Shop',
    'nav.gifts': 'Gifts',
    'nav.payments': 'Payments',
    'nav.payment-methods': 'Payment Methods',
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
    'desc.categories': 'Categories for products, prizes, gifts and currencies.',
    'desc.tags': 'Labels to group products and prizes.',

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
    'err.notfound': 'This endpoint is not available. Is the economy module loaded?',
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
    'code.FA_REQUIRED': 'A Persian name is required.',
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
    'code.CATEGORY_KIND_MISMATCH': 'This category is for another section.',
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
    'code.TEXT_TOO_LONG': 'This text is too long.',
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
    'f.name': 'Name',
    'f.name.fa': 'Persian name',
    'f.name.en': 'English name',
    'f.message.fa': 'Persian message',
    'f.message.en': 'English message',
    'f.description.fa': 'Persian description',
    'f.description.en': 'English description',
    'f.type': 'Type',
    'f.kind': 'Used for',
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
    'col.kind': 'Used for',
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
    'filter.kind': 'Used for',
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
    'ov.methods': 'Payment methods',
    'ov.sandbox': 'Sandbox',
    'ov.payments': 'Payments',
    'ov.revenue': 'Revenue',
    'ov.open': 'Open',
    'ov.today': 'Today',
    'ov.total': 'Total'
  }
};

var API_BASE = '/v2/console/economy';
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

function i18nText(v) {
  if (v == null) return '';
  if (typeof v === 'string') return v;
  var l = lang();
  return v[l] || v.fa || v.en || '';
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

function api(method, path, body, isForm) {
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  var init = { method: method, headers: headers, credentials: 'same-origin' };
  if (body !== undefined) {
    if (isForm) init.body = body;
    else { headers['Content-Type'] = 'application/json'; init.body = JSON.stringify(body); }
  }
  return fetch(API_BASE + path, init).then(function (res) {
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
        throw new ApiError(res.status, code, msg, data && data.field);
      }
      return data;
    });
  }, function () {
    throw new ApiError(0, 'NETWORK', '');
  });
}

function errorText(err) {
  if (!err) return t('err.generic');
  if (err.code && has('code.' + err.code)) return t('code.' + err.code);
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
  'payment-methods': '<rect x="2" y="5" width="20" height="14" rx="2"/><path d="M2 10h20M6 15h4"/>',
  categories: '<rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><circle cx="17.5" cy="17.5" r="3.5"/>',
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
  image: '<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="M21 15l-3.1-3.1a2 2 0 0 0-2.8 0L6 21"/>'
};
var MIRROR = { first: 1, prev: 1, next: 1, back: 1 };

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
  primary: '--x-button-text: #fff; --x-button-border: rgb(102,84,226); --x-button-dark-text: #fff; --x-button-dark-border: rgb(107 89 238); --x-button-bg: rgb(107 89 238); --x-button-bg-hover: rgb(102,84,226); --x-button-bg-active: rgb(69,58,154); --x-button-dark-bg: rgb(102,84,226); --x-button-dark-bg-hover: rgb(93,77,206); --x-button-dark-bg-active: rgb(69,58,154);',
  ghost: '--x-button-border: transparent; --x-button-text: #1f2937; --x-button-dark-text: white; --x-button-bg-hover: #f3f4f6; --x-button-bg-active: #e5e7eb; --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  muted: '--x-button-border: transparent; --x-button-text: #475569; --x-button-dark-text: #64748b; --x-button-bg-hover: #f8fafc; --x-button-bg-active: #f1f5f9; --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  link: '--x-button-border: transparent; --x-button-text: rgb(102,84,226); --x-button-dark-text: rgb(130,117,240); --x-button-bg-hover: rgb(244,244,253); --x-button-bg-active: rgb(233,232,252); --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
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
var INPUT_VARS = '--x-input-border: rgb(130,117,240);';
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
  var track = h('div', { class: '', style: '--x-toggle-bg: rgb(107 89 238); --x-toggle-dark-bg: rgb(102,84,226);' });
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
      ? '--x-checkbox-bg: rgb(107 89 238); --x-checkbox-border: rgb(107 89 238); --x-checkbox-dark-bg: rgb(107 89 238); --x-checkbox-dark-border: rgb(107 89 238);'
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

function notice(kind, text, action) {
  var color = kind === 'error' ? 'error' : kind === 'success' ? 'success' : 'secondary';
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

var KEY_RE = /^[a-z0-9][a-z0-9_.-]{0,63}$/;
var CURRENCY_KEY_RE = /^[a-z0-9_]{2,32}$/;
var CODE_RE = /^[A-Z0-9_-]{3,32}$/;
var refCache = {};

function refLabel(item) {
  if (!item) return '';
  var name = i18nText(item.name);
  return name || item.key || item.id || '';
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
  if (opts.oninput) el.addEventListener('input', function () { opts.oninput(el.value); });
  return el;
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
    var params = Object.assign({ q: q, page: 1, limit: 20 }, opts.filter || {});
    api('GET', '/' + opts.resource + qs(params)).then(function (res) {
      if (my !== seq) return;
      clear(list);
      var items = (res && res.items) || [];
      if (!items.length) list.appendChild(h('div', { class: 'px-2 py-2 text-sm text-secondary-500', text: t('no.results') }));
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

  paint(refCache[cacheKey(value)]);
  if (value && !refCache[cacheKey(value)]) {
    var req = (opts.valueKey || 'id') === 'id'
      ? api('GET', '/' + opts.resource + '/' + encodeURIComponent(value))
      : api('GET', '/' + opts.resource + qs(Object.assign({ q: value, limit: 20 }, opts.filter || {}))).then(function (res) {
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
  var picker = refPicker({ resource: opts.resource, filter: opts.filter, label: opts.label, placeholder: t('choose'), value: '', onchange: function (v) {
    if (v && values.indexOf(v) === -1) {
      values.push(v);
      paint();
      opts.onchange(values.slice());
    }
    picker.querySelector('.eco-ref-clear') && picker.querySelector('.eco-ref-clear').click();
  } });
  var missing = values.filter(function (v) { return !refCache[opts.resource + ':id:' + v]; });
  missing.forEach(function (v) {
    api('GET', '/' + opts.resource + '/' + encodeURIComponent(v)).then(function (it) {
      refCache[opts.resource + ':id:' + v] = it;
      paint();
    }, function () { return null; });
  });
  paint();
  return h('div', {}, chips, picker);
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
      control = selectInput(cur, f.options(), function (v) { change(v || null); }, { label: label, placeholder: f.required ? false : t('none') });
      if (f.required && (cur == null || cur === '')) { setPath(values, f.name, f.options()[0].value); }
    } else if (t2 === 'multi') {
      var set = (cur || []).slice();
      control = h('div', { class: 'flex flex-wrap gap-4' });
      f.options().forEach(function (o) {
        control.appendChild(checkbox(set.indexOf(o.value) !== -1, o.label, function (on) {
          set = set.filter(function (x) { return x !== o.value; });
          if (on) set.push(o.value);
          var order = f.options().map(function (x) { return x.value; });
          set.sort(function (a, b) { return order.indexOf(a) - order.indexOf(b); });
          change(set.slice());
        }));
      });
    } else if (t2 === 'ref') {
      control = refPicker({ resource: f.resource, valueKey: f.valueKey, filter: f.filter, value: cur, required: f.required, label: label, onchange: function (v) { change(v); } });
    } else if (t2 === 'refMulti') {
      control = refMulti({ resource: f.resource, filter: f.filter, value: cur, label: label, onchange: function (v) { change(v); } });
    } else if (t2 === 'image') {
      control = imageInput(cur, function (v) { change(v); });
    } else if (t2 === 'datetime') {
      control = dateInput(cur, function (v) { change(v); });
    } else if (t2 === 'items') {
      control = itemsEditor(cur, function (v) { setPath(values, f.name, v); if (entry.wrap.setError) entry.wrap.setError(''); });
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
      } else if (t2 === 'items') {
        msg = e.control.validate();
      } else if (f.required && (v === null || v === undefined || v === '' || (Array.isArray(v) && !v.length))) {
        msg = t('err.required');
      } else if (t2 === 'key' && v && !(f.keyRe || KEY_RE).test(v)) {
        msg = t(f.keyRe ? 'err.key' : 'err.key.general');
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
      if (f.type === 'items' && Array.isArray(v)) v = v.map(function (r) { return { type: r.type, key: r.key, action: r.action, value: r.value }; });
      if (/\.(fa|en)$/.test(f.name)) {
        var parentPath = f.name.slice(0, -3);
        if (typeof getPath(out, parentPath) !== 'object' || getPath(out, parentPath) === null) setPath(out, parentPath, {});
        if (v === null || v === undefined || String(v).trim() === '') return;
        v = String(v).trim();
      }
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

var nameRow = function (base, required) {
  return { type: 'row', fields: [
    { name: base + '.fa', label: 'f.' + base + '.fa', required: required !== false },
    { name: base + '.en', label: 'f.' + base + '.en', ltr: true, optional: true }
  ] };
};
var nameTextareas = function (base) {
  return { type: 'row', fields: [
    { name: base + '.fa', type: 'textarea', label: 'f.' + base + '.fa', optional: true },
    { name: base + '.en', type: 'textarea', label: 'f.' + base + '.en', ltr: true, optional: true }
  ] };
};

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
      nameRow('name'),
      { type: 'row', fields: [
        { name: 'defaultValue', type: 'number', label: 'f.defaultValue', hint: 'f.defaultValue.currency.hint', min: 0 },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function () { return { type: 'soft', name: { fa: '', en: '' }, defaultValue: 0, sort: 0, metadata: {} }; }
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
      nameRow('name'),
      { name: 'defaultValue', type: 'number', int: false, label: 'f.defaultValue', hint: 'f.defaultValue.meta.hint' },
      { name: 'serverOnly', type: 'bool', inline: true, label: 'f.serverOnly', hint: 'f.serverOnly.hint' },
      { name: 'settings', type: 'json', label: 'f.settings', optional: true, empty: {} }
    ],
    defaults: function () { return { name: { fa: '', en: '' }, defaultValue: 0, serverOnly: false, settings: {} }; }
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
      nameRow('name'),
      nameTextareas('message'),
      { type: 'row', fields: [
        { name: 'categoryId', type: 'ref', resource: 'categories', filter: { kind: 'prize' }, label: 'f.category', optional: true },
        { name: 'tagIds', type: 'refMulti', resource: 'tags', label: 'f.tags', optional: true, empty: [] }
      ] },
      { name: 'isRepeatable', type: 'bool', inline: true, label: 'f.isRepeatable', hint: 'f.isRepeatable.hint' },
      { name: 'items', type: 'items', label: 'f.items' },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function () { return { name: { fa: '', en: '' }, message: { fa: '', en: '' }, tagIds: [], isRepeatable: false, items: [{ type: 'currency', key: '', action: 'add', value: 0 }], metadata: {} }; }
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
      nameRow('name'),
      nameTextareas('description'),
      { type: 'row', fields: [
        { name: 'categoryId', type: 'ref', resource: 'categories', filter: { kind: 'product' }, label: 'f.category', optional: true },
        { name: 'tagIds', type: 'refMulti', resource: 'tags', label: 'f.tags', optional: true, empty: [] }
      ] },
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
    defaults: function () { return { isActive: true, name: { fa: '', en: '' }, description: { fa: '', en: '' }, tagIds: [], paymentMethods: ['zarinpal'], discountPercent: 0, sku: {}, coinPrice: null, isSubscription: false, minLevel: 0, sort: 0, metadata: {} }; },
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
      { label: 'col.name', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: it.name || '' }), it.code ? h('div', { class: 'text-secondary-500' }, mono(it.code)) : null); }, skeleton: '50%' },
      { label: 'col.type', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('gift.' + it.type), it.type === 'discount' ? 'warning' : 'primary'), it.type === 'discount' && it.discountPercent ? tag(t('percent', { n: it.discountPercent }), 'secondary') : null); }, width: 150 },
      { label: 'col.usage', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: fmtNum(it.totalUsed || 0) + ' / ' + (it.maxTotalUses ? fmtNum(it.maxTotalUses) : t('unlimited')) }); }, width: 130 },
      { label: 'col.expire', render: function (it) { return dateCell(it.expireAt); }, width: 190 },
      { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'name', label: 'f.gift.name', required: true },
        { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
      ] },
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
    defaults: function () { return { type: 'reward', isActive: true, isPromo: false, perUserLimit: 1, maxTotalUses: null, code: '' }; },
    fixPayload: function (p) { if (p.code === '') p.code = null; return p; }
  },
  categories: {
    filters: [{ name: 'kind', label: 'filter.kind', options: opts('kind.', ['product', 'prize', 'gift', 'currency']) }],
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '55%' },
      { label: 'col.kind', render: function (it) { return tag(t('kind.' + it.kind), 'secondary'); }, width: 130 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
        { name: 'kind', type: 'select', label: 'f.kind', required: true, options: opts('kind.', ['product', 'prize', 'gift', 'currency']) }
      ] },
      nameRow('name'),
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ],
    defaults: function (query) { return { kind: (query && query.kind) || 'product', name: { fa: '', en: '' }, sort: 0 }; }
  },
  tags: {
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '55%' },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
      nameRow('name')
    ],
    defaults: function () { return { name: { fa: '', en: '' } }; }
  }
};

var PAGES = ['overview', 'currencies', 'metas', 'prizes', 'products', 'gifts', 'payments', 'payment-methods', 'categories', 'tags'];
var LIMIT = 20;
var lastQuery = {};

function hashFor(page, id, query) {
  return '#/economy/' + page + (id ? '/' + encodeURIComponent(id) : '') + qs(query || {});
}

function go(page, id, query) {
  location.hash = hashFor(page, id, query);
}

function errorBlock(err, retry) {
  return h('div', { class: 'py-6' }, notice('error', errorText(err), retry ? button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: retry }) : null));
}

function listPage(container, name, query, ctx) {
  var def = RESOURCES[name] || PAYMENT_LIST;
  var columns = def.columns.map(function (c) { return { label: t(c.label), width: c.width, align: c.align, skeleton: c.skeleton }; });
  var state = { q: query.q || '', page: Math.max(1, parseInt(query.page, 10) || 1), limit: LIMIT, filters: {}, total: 0, loading: true, loaded: false };
  (def.filters || []).forEach(function (f) { if (query[f.name]) state.filters[f.name] = query[f.name]; });
  lastQuery[name] = currentQuery();

  var create = def.fields ? button({ variant: 'primary', size: 'page', icon: 'plus', label: t('new.' + name), onclick: function () { go(name, 'new', currentQuery()); } }) : null;
  var frame = pageFrame(t('nav.' + name), t('desc.' + name), create);
  var pagerTop = h('div', { class: 'relative flex items-center justify-end' });
  var pagerBottom = h('div', { class: 'flex gap-2 py-4 items-center justify-end' });
  var tableBox = h('div', {});
  var search = searchInput(state.q, t(name === 'payments' ? 'search.payments' : 'search'), debounce(function (e) {
    state.q = toLatinDigits(e.target.value).trim();
    state.page = 1;
    sync();
    load();
  }, 300));
  var filters = (def.filters || []).map(function (f) {
    var opts2 = typeof f.options === 'function' ? f.options() : f.options;
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
    return api('GET', '/' + name + qs(params)).then(function (res) {
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
          !filtered && def.fields ? button({ variant: 'primary', size: 'page', icon: 'plus', label: t('new.' + name), onclick: function () { go(name, 'new'); } }) : null)));
      } else {
        tableBox.appendChild(table(columns, items.map(function (it) {
          return { item: it, cells: def.columns.map(function (c) { return c.render(it); }) };
        }), { onrow: open }));
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
  var esc = function (e) { if (e.key === 'Escape' && panelState && !document.querySelector('[data-economy="panel"] .eco-pop:not([hidden])')) { e.preventDefault(); back(); } };
  document.addEventListener('keydown', esc, true);
  panelState = { el: el, key: name + '/' + id, esc: esc };
  document.body.appendChild(el);
  body.appendChild(formSkeleton(def));

  var item = isNew ? Promise.resolve(def.defaults(query)) : api('GET', '/' + name + '/' + encodeURIComponent(id));
  item.then(function (data) {
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
        api('DELETE', '/' + name + '/' + encodeURIComponent(id)).then(function () {
          toast('success', t('deleted'));
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
      api(isNew ? 'POST' : 'PUT', '/' + name + (isNew ? '' : '/' + encodeURIComponent(id)), payload).then(function () {
        toast('success', t('saved'));
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
  var frame = pageFrame(t('nav.payment-methods'), t('desc.payment-methods'));
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

var ROOT_ID = 'economy-root';
var NAV_LINK = 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent dark:border-transparent';
var NAV_ACTIVE = '!border-primary-600 !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700';
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
    if (!r.hasRoute('economy')) {
      r.addRoute({ path: '/economy/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'economy', component: EmptyPage, meta: { title: 'Economy' } }] });
    }
    router = r;
  } catch (e) {
    router = null;
    return;
  }
  if (pendingHash) {
    var target = pendingHash;
    pendingHash = null;
    router.replace(target.slice(1));
  }
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
  if (rendered) rendered.key = keyOf(parseRoute(hash));
  if (router) router.replace(hash.slice(1));
  else history.replaceState(history.state, '', location.pathname + location.search + hash);
}

function parseRoute(hash) {
  var m = /^#\/economy(?:\/([^/?]*))?(?:\/([^?]*))?(?:\?(.*))?$/.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  var page = m[1] || 'overview';
  if (PAGES.indexOf(page) === -1) page = 'overview';
  var query = {};
  (m[3] || '').split('&').forEach(function (pair) {
    if (!pair) return;
    var i = pair.indexOf('=');
    var k = decodeURIComponent(i === -1 ? pair : pair.slice(0, i));
    var v = i === -1 ? '' : decodeURIComponent(pair.slice(i + 1).replace(/\+/g, ' '));
    query[k] = v;
  });
  return { page: page, id: m[2] ? decodeURIComponent(m[2]) : null, query: query };
}

function keyOf(r) {
  if (!r) return '';
  var detail = r.page === 'payments' && r.id ? '/' + r.id : '';
  return lang() + '|' + r.page + detail + '|' + JSON.stringify(r.query);
}

function navIcon(name) {
  var svg = icon(name, 'h-4 w-4');
  return svg;
}

function buildNav(collapsed, active) {
  var ul = h('ul', { class: 'space-y-0.5' });
  PAGES.forEach(function (p) {
    var on = p === active;
    var a = h('a', { href: hashFor(p, null, p === active ? null : lastQuery[p]), class: (on ? NAV_ACTIVE + ' ' : '') + NAV_LINK, 'aria-current': on ? 'page' : null, title: collapsed ? t('nav.' + p) : null, 'data-economy-page': p });
    a.appendChild(collapsed ? h('div', { class: 'x-tooltip flex' }, navIcon(p)) : navIcon(p));
    a.appendChild(h('div', { class: 'font-medium relative truncate' }, h('span', { text: collapsed ? ' ' : t('nav.' + p) })));
    ul.appendChild(h('li', {}, a));
  });
  return h('div', { 'data-economy-nav': '', 'data-l10n-skip': '' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 ' + (collapsed ? 'text-[8px] text-center' : 'px-3 text-[10px]'), text: collapsed ? t('section.short') : t('section') }),
    h('nav', { 'aria-label': t('section') }, ul));
}

function isCollapsed(anchor) {
  var drawer = anchor.closest('.x-drawer');
  if (drawer && drawer.getBoundingClientRect().width > 0 && drawer.getBoundingClientRect().width < 100) return true;
  var span = anchor.querySelector('span');
  return !!span && span.textContent === ' ';
}

function syncSidebar(route) {
  var active = route ? route.page : null;
  var anchors = document.querySelectorAll('a[href="#/players"]');
  var keep = [];
  for (var i = 0; i < anchors.length; i++) {
    var a = anchors[i];
    if (!a.closest('.x-drawer') || a.closest('#' + ROOT_ID)) continue;
    var nav = a.closest('nav');
    var group = nav && nav.parentElement;
    if (!group || !group.parentElement) continue;
    var collapsed = isCollapsed(a);
    var key = lang() + '|' + collapsed + '|' + active + '|' + JSON.stringify(lastQuery);
    var next = group.nextElementSibling;
    var existing = next && next.hasAttribute('data-economy-nav') ? next : null;
    if (!existing || existing.getAttribute('data-key') !== key) {
      var built = buildNav(collapsed, active);
      built.setAttribute('data-key', key);
      if (existing) existing.replaceWith(built);
      else group.after(built);
      existing = built;
    }
    keep.push(existing);
    var vueActive = group.parentElement.querySelectorAll('a[aria-current="page"]:not([data-economy-page])');
    for (var j = 0; j < vueActive.length && active; j++) {
      if (vueActive[j].getAttribute('href') === '#/') vueActive[j].setAttribute('data-economy-dimmed', '');
    }
  }
  var all = document.querySelectorAll('[data-economy-nav]');
  for (var k = 0; k < all.length; k++) if (keep.indexOf(all[k]) === -1) all[k].remove();
  document.documentElement.toggleAttribute('data-economy-active', !!active);
}

function unmount() {
  var root = document.getElementById(ROOT_ID);
  if (root) root.remove();
  var hidden = document.querySelectorAll('[data-economy-hidden]');
  for (var i = 0; i < hidden.length; i++) {
    hidden[i].style.display = '';
    hidden[i].removeAttribute('data-economy-hidden');
  }
  closePanel();
  rendered = null;
}

function render(root, route) {
  var key = keyOf(route);
  var ctl = rendered && rendered.root === root ? rendered : null;
  if (!ctl || ctl.key !== key) {
    closePanel();
    window.scrollTo(0, 0);
    var scroller = document.getElementById('main');
    if (scroller) scroller.scrollTop = 0;
    ctl = { root: root, key: key, id: undefined, list: null };
    rendered = ctl;
    var ctx = { id: route.id, query: route.query };
    if (route.page === 'overview') overviewPage(root);
    else if (route.page === 'payment-methods') paymentMethodsPage(root);
    else if (route.page === 'payments' && route.id) paymentDetail(root, route.id);
    else ctl.list = listPage(root, route.page, route.query, ctx);
  }
  if (RESOURCES[route.page] && ctl.id !== route.id) {
    ctl.id = route.id;
    if (route.id) openEditor(route.page, route.id, route.query, function () { if (ctl.list) ctl.list.reload(); });
    else closePanel();
  }
}

function sync() {
  var route = parseRoute();
  syncSidebar(route);
  if (!route) { if (rendered || document.getElementById(ROOT_ID)) unmount(); return; }
  var main = document.querySelector('#main > main');
  if (!main) return;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') {
      mains[i].style.display = 'none';
      mains[i].setAttribute('data-economy-hidden', '');
    }
  }
  var root = document.getElementById(ROOT_ID);
  if (!root) {
    root = h('div', { id: ROOT_ID, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-economy': 'root' });
    rendered = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  render(root, route);
}

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
  new MutationObserver(function () { if (rendered) rendered.key = ''; schedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  sync();
}

if (/^#\/economy/.test(location.hash)) {
  pendingHash = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
trapApp();
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot);
else boot();
})();
