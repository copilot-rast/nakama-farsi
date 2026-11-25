(function () {
'use strict';
var INS_TEXT = {
  fa: {
    'group.insights': 'آنالیتیکس بازی',
    'nav.insights/overview': 'نمای کلی', 'nav.insights/active': 'کاربران فعال', 'nav.insights/retention': 'ماندگاری و کوهورت', 'nav.insights/funnels': 'قیف‌ها',
    'nav.insights/explore': 'کاوش رویدادها', 'nav.insights/sessions': 'جلسه‌ها', 'nav.insights/catalog': 'فهرست رویدادها و تگ‌ها', 'nav.insights/storage': 'محل ذخیره',
    'nav.insights/imports': 'ورود داده', 'nav.insights/settings': 'تنظیمات آنالیتیکس',
    'desc.insights/overview': 'خلاصهٔ بازی در بازهٔ انتخابی: کاربران فعال، کاربران تازه، ماندگاری و پرتکرارترین رویدادها.',
    'desc.insights/active': 'کاربران فعال روزانه، هفتگی و ماهانه؛ کاربر تازه در برابر کاربر برگشتی.',
    'desc.insights/retention': 'چند درصد از کاربرانی که در یک روز آمدند، روز ۱، ۷ و ۳۰ برگشتند. هر ردیف یک روز ورود است.',
    'desc.insights/funnels': 'مرحله‌ها را به ترتیب تعریف کنید تا ببینید چند نفر هر مرحله را رد کردند و کجا ریزش دارید.',
    'desc.insights/explore': 'هر رویداد را با هر ترکیبی از تگ‌ها جمع بزنید: کل، بر اساس کاربر، ماشین، مأموریت یا هر ترکیب دیگر.',
    'desc.insights/sessions': 'جلسه‌های بازی هر بازیکن با طول، پلتفرم و نسخه. روی یک جلسه بزنید تا رویدادهایش را به ترتیب زمان ببینید.',
    'desc.insights/catalog': 'گروه‌های تگ (مثل ماشین و مأموریت) و رویدادهای مجاز را برای هر بازی تعریف کنید. سرور رویدادها را با همین فهرست می‌سنجد.',
    'desc.insights/storage': 'رویدادها کجا ذخیره شوند: پستگرس خود سرور، پستگرس جدا، MongoDB یا ClickHouse. صف ارسال اگر مقصد قطع باشد رویدادی گم نمی‌کند.',
    'desc.insights/imports': 'داده را از Google Analytics 4 و GameAnalytics بیاورید. سرور در زمان‌بندی خودش داده را می‌کشد و به همین استاندارد تبدیل می‌کند.',
    'desc.insights/settings': 'سقف‌ها، محدودیت سرعت، سیاست رویدادهای ناشناخته و نگهداری داده.',
    'nav.insights/context': 'تگ‌هایی که سرور پر می‌کند',
    'desc.insights/context': 'سرور مأموریت و ماشین هر رویداد را از تلاش باز بازیکن برمی‌دارد. بازی لازم نیست این تگ‌ها را بفرستد و مقدار ساختگی هم جای مقدار درست را نمی‌گیرد.',
    'ctx.policy': 'رفتار', 'ctx.enabled': 'سرور تگ‌ها را پر کند', 'ctx.conflict': 'اگر بازی هم همین تگ را فرستاد',
    'ctx.c.override': 'مقدار سرور بماند', 'ctx.c.keep_client': 'مقدار بازی بماند', 'ctx.c.reject': 'رویداد رد شود',
    'ctx.flagMismatch': 'رویدادی که مقدارش با سرور فرق دارد علامت بخورد',
    'ctx.flagHint': 'رویداد علامت‌خورده در صفحهٔ جلسه‌ها با برچسب context_mismatch دیده می‌شود. مقدار بازی در فیلد clientTag.<گروه> می‌ماند. در کاوش رویدادها این رویدادها فقط با تیک «رویدادهای علامت‌خورده هم حساب شوند» شمرده می‌شوند.',
    'ctx.keepClientOutside': 'بیرون از مرحله، مقدار بازی پذیرفته شود',
    'ctx.outsideHint': 'مثلاً رویداد فروشگاه دربارهٔ یک ماشین وقتی هیچ مرحله‌ای باز نیست. اگر خاموش باشد، سرور این تگ‌ها را بیرون از مرحله دور می‌ریزد.',
    'ctx.grace': 'مهلت بعد از پایان مرحله (ثانیه)', 'ctx.graceHint': 'رویدادی که تا این مدت بعد از پایان مرحله رسید هنوز مال همان مرحله است. پیش‌فرض: {v}',
    'ctx.fills': 'گروه تگ و منبع', 'ctx.fillsHint': 'هر ردیف می‌گوید سرور کدام گروه تگ را از کجا پر کند. زمان رویداد (یا زمان رسیدن به سرور) تعیین می‌کند کدام تلاش باز بوده است.',
    'ctx.group': 'گروه تگ', 'ctx.source': 'منبع', 'ctx.fillOn': 'روشن', 'ctx.off': 'خاموش', 'ctx.addFill': 'افزودن ردیف', 'ctx.reset': 'برگشت به پیش‌فرض',
    'ctx.defaults': 'پیش‌فرض: {v}', 'ctx.noProviders': 'هیچ افزونه‌ای منبع نداده است (مثلاً افزونهٔ مرحله‌ها بار نشده). تا آن موقع سرور تگی پر نمی‌کند و تگ‌های بازی دست نمی‌خورند.',
    'code.INVALID_TAG_GROUP': 'گروه تگ نامعتبر است (حروف کوچک انگلیسی، عدد و _)', 'code.INVALID_SOURCE': 'منبع نامعتبر است', 'code.DUPLICATE_TAG_GROUP': 'این گروه تگ دو بار آمده است',
    'save': 'ذخیره', 'saved': 'ذخیره شد', 'deleted': 'حذف شد', 'cancel': 'انصراف', 'close': 'بستن', 'delete': 'حذف', 'yes': 'بله', 'no': 'خیر', 'add': 'افزودن', 'edit': 'ویرایش',
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'page.first': 'صفحهٔ اول', 'page.last': 'صفحهٔ آخر', 'page.of': 'صفحهٔ {page} از {pages}', 'pager': 'صفحه‌بندی', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'none': '—', 'all': 'همه',
    'confirm.title': 'مطمئن هستید؟', 'lang.text': '{label} ({lang})', 'updated': 'آخرین تغییر: {by} · {at}', 'missing': 'پیدا نشد',
    'f.from': 'از', 'f.to': 'تا', 'f.range': 'بازهٔ تاریخ', 'f.game': 'بازی', 'f.platform': 'پلتفرم', 'f.version': 'نسخه', 'f.apply': 'اعمال', 'f.days': '{n} روز', 'f.flagged': 'رویدادهای علامت‌خورده هم حساب شوند',
    'f.allGames': 'همهٔ بازی‌ها', 'f.mainGame': 'بازی اصلی', 'f.any': 'هر کدام',
    'csv': 'خروجی CSV', 'store': 'منبع داده: {s}',
    'k.dau': 'فعال امروز', 'k.wau': 'فعال ۷ روز', 'k.mau': 'فعال ۳۰ روز', 'k.stick': 'چسبندگی (DAU/MAU)', 'k.new': 'کاربر تازه', 'k.events': 'رویدادها', 'k.users': 'کاربران بازه',
    'k.d': 'ماندگاری روز {n}',
    'c.dau': 'فعال روزانه', 'c.new': 'تازه', 'c.returning': 'برگشتی', 'c.events': 'رویداد', 'c.day': 'روز', 'c.users': 'کاربر', 'c.count': 'تعداد', 'c.value': 'مقدار',
    'c.name': 'رویداد', 'c.platform': 'پلتفرم', 'c.bucket': 'بازه', 'c.cohort': 'روز ورود', 'c.size': 'اندازه',
    'top.events': 'پرتکرارترین رویدادها', 'top.platforms': 'پلتفرم‌ها',
    'r.days': 'روزها', 'r.mode': 'روش', 'r.classic': 'دقیقاً همان روز', 'r.rolling': 'همان روز یا بعد', 'r.pending': 'هنوز نرسیده',
    'fn.new': 'قیف تازه', 'fn.name': 'نام قیف', 'fn.window': 'پنجرهٔ زمانی (دقیقه)', 'fn.steps': 'مرحله‌ها', 'fn.step': 'مرحلهٔ {n}', 'fn.event': 'نام رویداد', 'fn.tags': 'تگ‌ها (اختیاری)',
    'fn.addStep': 'مرحلهٔ دیگر', 'fn.pick': 'یک قیف را انتخاب کنید یا قیف تازه بسازید.', 'fn.empty': 'هنوز قیفی نساخته‌اید.', 'fn.fromFirst': 'از مرحلهٔ اول', 'fn.fromPrev': 'از مرحلهٔ قبل',
    'fn.drop': 'ریزش', 'fn.median': 'میانهٔ زمان', 'fn.hintTags': 'مثلاً level=1,2 ; mode=ranked', 'fn.windowV': 'پنجره: {n} دقیقه',
    'ex.names': 'رویدادها', 'ex.namesHint': 'با ویرگول جدا کنید؛ خالی یعنی همه', 'ex.metric': 'سنجه', 'ex.group': 'گروه‌بندی بر اساس', 'ex.bucket': 'تقسیم زمانی', 'ex.filters': 'فیلتر تگ',
    'ex.filtersHint': 'مثلاً car=345345 ; mission=1,2', 'ex.limit': 'حداکثر ردیف', 'ex.run': 'اجرا', 'ex.noGroup': 'بدون گروه‌بندی',
    'm.count': 'تعداد رویداد', 'm.sum': 'جمع مقدار', 'm.avg': 'میانگین مقدار', 'm.min': 'کمترین', 'm.max': 'بیشترین', 'm.users': 'کاربران یکتا',
    'b.none': 'بدون تقسیم', 'b.hour': 'ساعتی', 'b.day': 'روزانه', 'b.week': 'هفتگی',
    'g.name': 'نام رویداد', 'g.platform': 'پلتفرم', 'g.appVersion': 'نسخه', 'g.user': 'کاربر', 'g.source': 'منبع', 'g.game': 'بازی', 'g.tag': 'تگ: {g}',
    's.player': 'بازیکن', 's.started': 'شروع', 's.length': 'طول', 's.status': 'وضعیت', 's.open': 'باز', 's.ended': 'تمام‌شده', 's.userId': 'شناسهٔ بازیکن', 's.timeline': 'رویدادهای جلسه',
    's.byTime': 'این جلسه شناسهٔ جلسه نداشت؛ رویدادهای بازیکن در همان بازهٔ زمانی نشان داده شده.', 's.noEvents': 'رویدادی برای این جلسه ثبت نشده.', 's.source': 'منبع',
    'cat.groups': 'گروه‌های تگ', 'cat.events': 'رویدادها', 'cat.newGroup': 'گروه تازه', 'cat.newEvent': 'رویداد تازه', 'cat.key': 'کلید', 'cat.nameFa': 'نام فارسی', 'cat.nameEn': 'نام انگلیسی',
    'cat.desc': 'توضیح', 'cat.closed': 'فقط مقدارهای فهرست مجازند', 'cat.enabled': 'فعال', 'cat.values': 'مقدارها', 'cat.valuesHint': 'هر خط: شناسه,نام فارسی,نام انگلیسی',
    'cat.valuesN': '{n} مقدار', 'cat.open': 'باز', 'cat.closedTag': 'بسته', 'cat.name': 'نام رویداد', 'cat.label': 'برچسب', 'cat.path': 'گروه‌های مسیر قدیمی', 'cat.pathHint': 'برای Name:Id1:Id2 به ترتیب؛ مثلاً car, mission',
    'cat.required': 'تگ‌های لازم', 'cat.allowed': 'تگ‌های مجاز (خالی یعنی همه)', 'cat.valueRequired': 'مقدار لازم است', 'cat.min': 'کمترین مقدار', 'cat.max': 'بیشترین مقدار', 'cat.unit': 'واحد',
    'cat.keyHint': 'حرف کوچک انگلیسی، عدد و _', 'cat.game': 'بازی', 'cat.emptyGroups': 'هنوز گروه تگی تعریف نشده.', 'cat.emptyEvents': 'هنوز رویدادی تعریف نشده؛ رویدادهای ناشناخته طبق تنظیمات پذیرفته می‌شوند.',
    'st.driver': 'نوع ذخیره‌ساز', 'st.nakama': 'پستگرس خود سرور (پیش‌فرض)', 'st.postgres': 'PostgreSQL جدا', 'st.clickhouse': 'ClickHouse', 'st.mongodb': 'MongoDB',
    'st.host': 'میزبان', 'st.port': 'پورت', 'st.database': 'پایگاه داده', 'st.user': 'کاربر', 'st.sslMode': 'SSL', 'st.schema': 'schema', 'st.url': 'نشانی HTTP', 'st.table': 'جدول',
    'st.hosts': 'میزبان‌ها (با ویرگول)', 'st.authSource': 'پایگاه احراز هویت', 'st.tls': 'اتصال TLS', 'st.collection': 'collection', 'st.secret': 'رمز عبور',
    'st.secretKeep': 'خالی بگذارید تا رمز فعلی بماند', 'st.secretSet': 'رمز ذخیره شده', 'st.noKey': 'ECONOMY_SECRET_KEY روی سرور تنظیم نیست؛ رمز ذخیره نمی‌شود.',
    'st.test': 'آزمایش اتصال', 'st.testOk': 'اتصال برقرار شد ({ms} میلی‌ثانیه)', 'st.testFail': 'اتصال برقرار نشد: {e}', 'st.status': 'وضعیت فعلی', 'st.error': 'خطای باز کردن: {e}',
    'st.switchNote': 'با عوض کردن محل ذخیره، رویدادهای قبلی جابه‌جا نمی‌شوند؛ داشبوردها فقط محل فعلی را می‌خوانند.',
    'ob.title': 'صف ارسال', 'ob.pending': 'در صف', 'ob.rows': 'بسته‌ها', 'ob.dead': 'بسته‌های ناموفق', 'ob.oldest': 'قدیمی‌ترین', 'ob.retry': 'تلاش دوباره', 'ob.discard': 'دور ریختن ناموفق‌ها',
    'ob.failing': 'بسته‌های با خطا', 'ob.attempts': 'تلاش', 'ob.error': 'خطا', 'ob.next': 'تلاش بعدی', 'ob.ok': 'صف خالی است و همه چیز رسیده.', 'ob.discardQ': 'رویدادهای ناموفق برای همیشه حذف شوند؟',
    'im.new': 'منبع تازه', 'im.kind': 'نوع', 'im.name': 'نام', 'im.every': 'هر چند دقیقه', 'im.config': 'تنظیمات (JSON)', 'im.secret': 'کلید محرمانه', 'im.secretHint.ga': 'JSON حساب سرویس گوگل',
    'im.secretHint.s3': 'Secret key دسترسی S3', 'im.reset': 'از اول بخوان (پاک کردن نشانگر)', 'im.run': 'اجرا همین حالا', 'im.queued': 'در صف اجرا', 'im.runs': 'اجراهای اخیر',
    'im.last': 'آخرین اجرا', 'im.status': 'وضعیت', 'im.read': 'خوانده', 'im.written': 'نوشته', 'im.skipped': 'رد شده', 'im.cursor': 'نشانگر', 'im.running': 'در حال اجرا',
    'im.untested': 'این اتصال‌ها فقط با سرور آزمایشی تست شده‌اند؛ اولین اجرای واقعی را با دقت بررسی کنید.', 'im.empty': 'هنوز منبعی اضافه نشده.',
    'kind.ga4_bigquery': 'GA4 (BigQuery export)', 'kind.ga4_data': 'GA4 (Data API، تجمیعی)', 'kind.gameanalytics': 'GameAnalytics (raw export روی S3)',
    'set.events': 'رویدادها', 'set.limits': 'سقف‌ها', 'set.rate': 'محدودیت سرعت (برای هر بازیکن)', 'set.policy': 'سیاست‌ها', 'set.storage': 'نگهداری و صف', 'set.client': 'راهنمای کلاینت',
    'set.enabled': 'دریافت رویداد روشن باشد', 'set.maxBatch': 'حداکثر رویداد در هر بسته', 'set.maxTags': 'حداکثر تگ هر رویداد', 'set.maxNameLength': 'حداکثر طول نام', 'set.maxFields': 'حداکثر فیلد آزاد',
    'set.maxStringLength': 'حداکثر طول متن فیلد', 'set.valueMin': 'کمترین مقدار مجاز', 'set.valueMax': 'بیشترین مقدار مجاز', 'set.batchesPerMinute': 'بسته در دقیقه', 'set.eventsPerMinute': 'رویداد در دقیقه',
    'set.pastHours': 'زمان کلاینت تا چند ساعت عقب', 'set.futureMinutes': 'زمان کلاینت تا چند دقیقه جلو', 'set.receiptDays': 'نگهداری رسید بسته (روز)',
    'set.unknownEvents': 'رویداد ناشناخته', 'set.unknownTagGroups': 'گروه تگ ناشناخته', 'set.unknownTagValues': 'مقدار بیرون از فهرست بسته', 'set.valueOutOfRange': 'مقدار بیرون از بازه',
    'set.gameEvents': 'رویدادهای سرور (gameevents) ثبت شوند', 'set.sessionEvents': 'شروع و پایان جلسه به‌عنوان رویداد ثبت شوند', 'set.flushSeconds': 'فاصلهٔ ارسال کلاینت (ثانیه)',
    'set.maxQueue': 'حداکثر صف کلاینت', 'set.keepDays': 'نگهداری رویداد خام (روز، ۰ یعنی همیشه)', 'set.batchSize': 'بسته در هر برداشت صف', 'set.maxPendingEvents': 'حداکثر رویداد در صف',
    'set.deadAfter': 'ناموفق بعد از چند تلاش', 'p.accept': 'بپذیر', 'p.flag': 'بپذیر و علامت بزن', 'p.reject': 'رد کن', 'set.default': 'پیش‌فرض: {v}',
    'code.INVALID_QUERY': 'پرسش نامعتبر است', 'code.STORE_UNAVAILABLE': 'محل ذخیره در دسترس نیست', 'code.INVALID_KEY': 'کلید نامعتبر است', 'code.INVALID_EVENT_NAME': 'نام رویداد نامعتبر است',
    'code.INVALID_TAG_VALUE': 'شناسهٔ مقدار نامعتبر است', 'code.REQUIRED': 'این فیلد لازم است', 'code.VALUE_OUT_OF_RANGE': 'مقدار خارج از بازه است', 'code.SECRET_KEY_MISSING': 'ECONOMY_SECRET_KEY روی سرور تنظیم نیست',
    'code.INVALID_DRIVER': 'نوع ذخیره‌ساز نامعتبر است', 'code.INVALID_URL': 'نشانی نامعتبر است', 'code.INVALID_KIND': 'نوع منبع نامعتبر است', 'code.INVALID_JSON': 'JSON نامعتبر است', 'code.INVALID_DATE': 'تاریخ نامعتبر است'
  },
  en: {
    'group.insights': 'Game analytics',
    'nav.insights/overview': 'Overview', 'nav.insights/active': 'Active users', 'nav.insights/retention': 'Retention and cohorts', 'nav.insights/funnels': 'Funnels',
    'nav.insights/explore': 'Event explorer', 'nav.insights/sessions': 'Sessions', 'nav.insights/catalog': 'Events and tags', 'nav.insights/storage': 'Storage',
    'nav.insights/imports': 'Imports', 'nav.insights/settings': 'Analytics settings',
    'desc.insights/overview': 'The game in the chosen range: active users, new users, retention and the most frequent events.',
    'desc.insights/active': 'Daily, weekly and monthly active users; new against returning players.',
    'desc.insights/retention': 'Share of the players who first came on a day that came back on day 1, 7 and 30. Each row is a first day.',
    'desc.insights/funnels': 'Define ordered steps to see how many players pass each step and where they drop off.',
    'desc.insights/explore': 'Aggregate any event by any combination of tags: total, by user, car, mission or any mix.',
    'desc.insights/sessions': 'Play sessions per player with length, platform and version. Open a session to see its events in time order.',
    'desc.insights/catalog': 'Define tag groups (such as car and mission) and the allowed events for each game. The server checks events against this list.',
    'desc.insights/storage': 'Where events are stored: the server\'s own Postgres, a separate PostgreSQL, MongoDB or ClickHouse. The delivery queue loses nothing while the store is down.',
    'desc.insights/imports': 'Bring data from Google Analytics 4 and GameAnalytics. The server pulls on its schedule and maps it to the same standard.',
    'desc.insights/settings': 'Limits, rate limits, unknown event policies and data retention.',
    'nav.insights/context': 'Server-filled tags',
    'desc.insights/context': 'The server takes the mission and car of each event from the player\'s open attempt. The game does not need to send these tags, and a faked value cannot replace the real one.',
    'ctx.policy': 'Behaviour', 'ctx.enabled': 'The server fills the tags', 'ctx.conflict': 'When the game sends the same tag',
    'ctx.c.override': 'Keep the server value', 'ctx.c.keep_client': 'Keep the game value', 'ctx.c.reject': 'Reject the event',
    'ctx.flagMismatch': 'Flag events whose value differs from the server',
    'ctx.flagHint': 'Flagged events show up on the Sessions page with the context_mismatch label. The game value is kept in the field clientTag.<group>. The explorer counts them only with "Count flagged events too".',
    'ctx.keepClientOutside': 'Accept game values outside a stage',
    'ctx.outsideHint': 'For example a shop event about a car while no stage is open. When off, the server drops these tags outside a stage.',
    'ctx.grace': 'Grace after the stage ends (seconds)', 'ctx.graceHint': 'An event up to this long after the stage ended still belongs to it. Default: {v}',
    'ctx.fills': 'Tag groups and sources', 'ctx.fillsHint': 'Each row says which tag group the server fills and from where. The event time (or the server receive time) decides which attempt was open.',
    'ctx.group': 'Tag group', 'ctx.source': 'Source', 'ctx.fillOn': 'On', 'ctx.off': 'off', 'ctx.addFill': 'Add row', 'ctx.reset': 'Back to defaults',
    'ctx.defaults': 'Default: {v}', 'ctx.noProviders': 'No plugin offers a source yet (for example the campaign plugin is not loaded). Until then the server fills nothing and game tags are left alone.',
    'code.INVALID_TAG_GROUP': 'Invalid tag group (lowercase letters, digits and _)', 'code.INVALID_SOURCE': 'Invalid source', 'code.DUPLICATE_TAG_GROUP': 'This tag group is listed twice',
    'save': 'Save', 'saved': 'Saved', 'deleted': 'Deleted', 'cancel': 'Cancel', 'close': 'Close', 'delete': 'Delete', 'yes': 'Yes', 'no': 'No', 'add': 'Add', 'edit': 'Edit',
    'loading': 'Loading…', 'empty': 'Nothing found', 'page.first': 'First page', 'page.last': 'Last page', 'page.of': 'Page {page} of {pages}', 'pager': 'Pagination', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'none': '—', 'all': 'All',
    'confirm.title': 'Are you sure?', 'lang.text': '{label} ({lang})', 'updated': 'Last change: {by} · {at}', 'missing': 'Not found',
    'f.from': 'From', 'f.to': 'To', 'f.range': 'Date range', 'f.game': 'Game', 'f.platform': 'Platform', 'f.version': 'Version', 'f.apply': 'Apply', 'f.days': '{n} days', 'f.flagged': 'Count flagged events too',
    'f.allGames': 'All games', 'f.mainGame': 'Main game', 'f.any': 'Any',
    'csv': 'Export CSV', 'store': 'Data source: {s}',
    'k.dau': 'Active today', 'k.wau': 'Active 7 days', 'k.mau': 'Active 30 days', 'k.stick': 'Stickiness (DAU/MAU)', 'k.new': 'New users', 'k.events': 'Events', 'k.users': 'Users in range',
    'k.d': 'Day {n} retention',
    'c.dau': 'Daily active', 'c.new': 'New', 'c.returning': 'Returning', 'c.events': 'Events', 'c.day': 'Day', 'c.users': 'Users', 'c.count': 'Count', 'c.value': 'Value',
    'c.name': 'Event', 'c.platform': 'Platform', 'c.bucket': 'Period', 'c.cohort': 'First day', 'c.size': 'Size',
    'top.events': 'Top events', 'top.platforms': 'Platforms',
    'r.days': 'Days', 'r.mode': 'Method', 'r.classic': 'Exactly that day', 'r.rolling': 'That day or later', 'r.pending': 'Not yet',
    'fn.new': 'New funnel', 'fn.name': 'Funnel name', 'fn.window': 'Time window (minutes)', 'fn.steps': 'Steps', 'fn.step': 'Step {n}', 'fn.event': 'Event name', 'fn.tags': 'Tags (optional)',
    'fn.addStep': 'Another step', 'fn.pick': 'Pick a funnel or make a new one.', 'fn.empty': 'No funnels yet.', 'fn.fromFirst': 'From step 1', 'fn.fromPrev': 'From previous',
    'fn.drop': 'Drop-off', 'fn.median': 'Median time', 'fn.hintTags': 'e.g. level=1,2 ; mode=ranked', 'fn.windowV': 'Window: {n} minutes',
    'ex.names': 'Events', 'ex.namesHint': 'Comma separated; empty means all', 'ex.metric': 'Metric', 'ex.group': 'Group by', 'ex.bucket': 'Time split', 'ex.filters': 'Tag filter',
    'ex.filtersHint': 'e.g. car=345345 ; mission=1,2', 'ex.limit': 'Max rows', 'ex.run': 'Run', 'ex.noGroup': 'No grouping',
    'm.count': 'Event count', 'm.sum': 'Sum of value', 'm.avg': 'Average value', 'm.min': 'Minimum', 'm.max': 'Maximum', 'm.users': 'Unique users',
    'b.none': 'No split', 'b.hour': 'Hourly', 'b.day': 'Daily', 'b.week': 'Weekly',
    'g.name': 'Event name', 'g.platform': 'Platform', 'g.appVersion': 'Version', 'g.user': 'User', 'g.source': 'Source', 'g.game': 'Game', 'g.tag': 'Tag: {g}',
    's.player': 'Player', 's.started': 'Started', 's.length': 'Length', 's.status': 'Status', 's.open': 'Open', 's.ended': 'Ended', 's.userId': 'Player id', 's.timeline': 'Session events',
    's.byTime': 'Events had no session id; the player\'s events in the same time span are shown.', 's.noEvents': 'No events were recorded for this session.', 's.source': 'Source',
    'cat.groups': 'Tag groups', 'cat.events': 'Events', 'cat.newGroup': 'New group', 'cat.newEvent': 'New event', 'cat.key': 'Key', 'cat.nameFa': 'Persian name', 'cat.nameEn': 'English name',
    'cat.desc': 'Description', 'cat.closed': 'Only listed values are allowed', 'cat.enabled': 'Enabled', 'cat.values': 'Values', 'cat.valuesHint': 'One per line: id,Persian label,English label',
    'cat.valuesN': '{n} values', 'cat.open': 'Open', 'cat.closedTag': 'Closed', 'cat.name': 'Event name', 'cat.label': 'Label', 'cat.path': 'Legacy path groups', 'cat.pathHint': 'For Name:Id1:Id2 in order, e.g. car, mission',
    'cat.required': 'Required tags', 'cat.allowed': 'Allowed tags (empty means any)', 'cat.valueRequired': 'Value is required', 'cat.min': 'Minimum value', 'cat.max': 'Maximum value', 'cat.unit': 'Unit',
    'cat.keyHint': 'Lowercase letters, digits and _', 'cat.game': 'Game', 'cat.emptyGroups': 'No tag groups yet.', 'cat.emptyEvents': 'No events defined yet; unknown events follow the settings.',
    'st.driver': 'Store type', 'st.nakama': 'The server\'s own Postgres (default)', 'st.postgres': 'Separate PostgreSQL', 'st.clickhouse': 'ClickHouse', 'st.mongodb': 'MongoDB',
    'st.host': 'Host', 'st.port': 'Port', 'st.database': 'Database', 'st.user': 'User', 'st.sslMode': 'SSL', 'st.schema': 'Schema', 'st.url': 'HTTP URL', 'st.table': 'Table',
    'st.hosts': 'Hosts (comma separated)', 'st.authSource': 'Auth database', 'st.tls': 'TLS connection', 'st.collection': 'Collection', 'st.secret': 'Password',
    'st.secretKeep': 'Leave empty to keep the current password', 'st.secretSet': 'Password stored', 'st.noKey': 'ECONOMY_SECRET_KEY is not set on the server; passwords cannot be stored.',
    'st.test': 'Test connection', 'st.testOk': 'Connected ({ms} ms)', 'st.testFail': 'Could not connect: {e}', 'st.status': 'Current state', 'st.error': 'Open failed: {e}',
    'st.switchNote': 'Switching the store does not move earlier events; dashboards read only the current store.',
    'ob.title': 'Delivery queue', 'ob.pending': 'Queued', 'ob.rows': 'Batches', 'ob.dead': 'Failed batches', 'ob.oldest': 'Oldest', 'ob.retry': 'Retry', 'ob.discard': 'Discard failed',
    'ob.failing': 'Batches with errors', 'ob.attempts': 'Attempts', 'ob.error': 'Error', 'ob.next': 'Next try', 'ob.ok': 'The queue is empty; everything was delivered.', 'ob.discardQ': 'Delete the failed events for good?',
    'im.new': 'New source', 'im.kind': 'Kind', 'im.name': 'Name', 'im.every': 'Every (minutes)', 'im.config': 'Settings (JSON)', 'im.secret': 'Secret', 'im.secretHint.ga': 'Google service account JSON',
    'im.secretHint.s3': 'S3 secret access key', 'im.reset': 'Read from the start (reset cursor)', 'im.run': 'Run now', 'im.queued': 'Queued', 'im.runs': 'Recent runs',
    'im.last': 'Last run', 'im.status': 'Status', 'im.read': 'Read', 'im.written': 'Written', 'im.skipped': 'Skipped', 'im.cursor': 'Cursor', 'im.running': 'Running',
    'im.untested': 'These connectors were tested only against test servers; check the first real run closely.', 'im.empty': 'No sources yet.',
    'kind.ga4_bigquery': 'GA4 (BigQuery export)', 'kind.ga4_data': 'GA4 (Data API, aggregated)', 'kind.gameanalytics': 'GameAnalytics (raw export on S3)',
    'set.events': 'Events', 'set.limits': 'Limits', 'set.rate': 'Rate limits (per player)', 'set.policy': 'Policies', 'set.storage': 'Retention and queue', 'set.client': 'Client hints',
    'set.enabled': 'Accept events', 'set.maxBatch': 'Max events per batch', 'set.maxTags': 'Max tags per event', 'set.maxNameLength': 'Max name length', 'set.maxFields': 'Max free fields',
    'set.maxStringLength': 'Max field text length', 'set.valueMin': 'Lowest allowed value', 'set.valueMax': 'Highest allowed value', 'set.batchesPerMinute': 'Batches per minute', 'set.eventsPerMinute': 'Events per minute',
    'set.pastHours': 'Client time up to hours back', 'set.futureMinutes': 'Client time up to minutes ahead', 'set.receiptDays': 'Keep batch receipts (days)',
    'set.unknownEvents': 'Unknown event', 'set.unknownTagGroups': 'Unknown tag group', 'set.unknownTagValues': 'Value outside a closed group', 'set.valueOutOfRange': 'Value out of range',
    'set.gameEvents': 'Record server game events (gameevents)', 'set.sessionEvents': 'Record session start and end as events', 'set.flushSeconds': 'Client send interval (seconds)',
    'set.maxQueue': 'Client queue limit', 'set.keepDays': 'Keep raw events (days, 0 = forever)', 'set.batchSize': 'Batches per queue claim', 'set.maxPendingEvents': 'Max queued events',
    'set.deadAfter': 'Failed after attempts', 'p.accept': 'Accept', 'p.flag': 'Accept and flag', 'p.reject': 'Reject', 'set.default': 'Default: {v}',
    'code.INVALID_QUERY': 'The query is not valid', 'code.STORE_UNAVAILABLE': 'The store is not reachable', 'code.INVALID_KEY': 'Invalid key', 'code.INVALID_EVENT_NAME': 'Invalid event name',
    'code.INVALID_TAG_VALUE': 'Invalid value id', 'code.REQUIRED': 'This field is required', 'code.VALUE_OUT_OF_RANGE': 'Value out of range', 'code.SECRET_KEY_MISSING': 'ECONOMY_SECRET_KEY is not set on the server',
    'code.INVALID_DRIVER': 'Invalid store type', 'code.INVALID_URL': 'Invalid URL', 'code.INVALID_KIND': 'Invalid source kind', 'code.INVALID_JSON': 'Invalid JSON', 'code.INVALID_DATE': 'Invalid date'
  }
};

function insLang() {
  var nc = window.NakamaConsole;
  var l = nc && nc.lang ? nc.lang() : (document.documentElement.getAttribute('lang') || 'fa');
  return /^en/i.test(l) ? 'en' : 'fa';
}

function has(key) {
  return (INS_TEXT[insLang()] || INS_TEXT.fa)[key] != null;
}

function t(key, vars) {
  var dict = INS_TEXT[insLang()] || INS_TEXT.fa;
  var s = dict[key] != null ? dict[key] : (INS_TEXT.en[key] != null ? INS_TEXT.en[key] : key);
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] != null ? digits(String(vars[k])) : m; });
  return s;
}

function digits(s) {
  if (insLang() !== 'fa') return s;
  return String(s).replace(/[0-9]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'[+d]; }).replace(/,/g, '٬').replace(/\./g, '٫');
}

function fmtNum(n, frac) {
  if (n == null || n === '') return t('none');
  var s = Number(n).toLocaleString('en-US', { maximumFractionDigits: frac == null ? 2 : frac });
  return digits(s);
}

function fmtPct(r) {
  if (r == null) return t('none');
  return digits((Math.round(r * 1000) / 10).toFixed(1)) + (insLang() === 'fa' ? '٪' : '%');
}

function fmtDate(v) {
  if (!v) return '';
  if (window.NakamaDate) return window.NakamaDate.format(v);
  var d = new Date(v);
  if (isNaN(d)) return '';
  try {
    return d.toLocaleString(insLang() === 'fa' ? 'fa-IR' : 'en-GB', { calendar: insLang() === 'fa' ? 'persian' : 'gregory', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

function fmtTime(v) {
  if (!v) return '';
  if (window.NakamaDate) return window.NakamaDate.format(v, { parts: { hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23' } });
  var d = new Date(v);
  if (isNaN(d)) return '';
  try {
    return d.toLocaleTimeString(insLang() === 'fa' ? 'fa-IR' : 'en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

function fmtDay(s) {
  if (!s) return '';
  if (window.NakamaDate && /^\d{4}-\d{2}-\d{2}$/.test(s)) return window.NakamaDate.format(s, { parts: { month: 'short', day: 'numeric' } });
  var d = new Date(s + 'T12:00:00Z');
  if (isNaN(d)) return digits(s);
  try {
    return d.toLocaleDateString(insLang() === 'fa' ? 'fa-IR' : 'en-GB', { calendar: insLang() === 'fa' ? 'persian' : 'gregory', timeZone: 'UTC', month: 'short', day: 'numeric' });
  } catch (e) {
    return s;
  }
}

function fmtSeconds(s) {
  if (s == null) return t('none');
  s = Math.round(s);
  var h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60), x = s % 60;
  var out = h ? h + ':' + String(m).padStart(2, '0') + ':' + String(x).padStart(2, '0') : m + ':' + String(x).padStart(2, '0');
  return digits(out);
}

var INS_API = '/v2/console/base';
var INS_INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 px-3 py-2 text-sm';
var INS_BTN = 'ia-btn inline-flex items-center justify-center gap-1.5 font-medium whitespace-nowrap rounded-md border text-sm py-1.5 px-3';

function h(tag, props) {
  var el = document.createElement(tag);
  if (props) {
    for (var k in props) {
      var v = props[k];
      if (v === null || v === undefined || v === false) continue;
      if (k === 'class') el.setAttribute('class', v);
      else if (k === 'text') el.textContent = v;
      else if (k.indexOf('on') === 0 && typeof v === 'function') el.addEventListener(k.slice(2).toLowerCase(), v);
      else if (k === 'value') el.value = v;
      else if (k === 'checked') el.checked = !!v;
      else el.setAttribute(k, v === true ? '' : v);
    }
  }
  for (var i = 2; i < arguments.length; i++) add(el, arguments[i]);
  return el;
}

function add(el, child) {
  if (child === null || child === undefined || child === false) return;
  if (Array.isArray(child)) { child.forEach(function (c) { add(el, c); }); return; }
  el.appendChild(typeof child === 'object' ? child : document.createTextNode(String(child)));
}

function clear(el) {
  while (el.firstChild) el.removeChild(el.firstChild);
  return el;
}

function token() {
  var stores = [];
  try { stores.push(window.sessionStorage); } catch (e) { stores.push(null); }
  try { stores.push(window.localStorage); } catch (e) { stores.push(null); }
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

function InsError(status, code, message, field, meta) {
  this.meta = meta || null;
  this.status = status;
  this.code = code;
  this.message = message;
  this.field = field;
}

function api(method, path, body) {
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  var init = { method: method, headers: headers, credentials: 'same-origin' };
  if (body !== undefined) { headers['Content-Type'] = 'application/json'; init.body = JSON.stringify(body); }
  return fetch((path.indexOf('/v2/') === 0 ? '' : INS_API) + path, init).then(function (res) {
    if (res.status === 401 && !/^#\/login/.test(location.hash)) location.hash = '#/login';
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) throw new InsError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', data && data.field || '', data && data.meta);
      return data;
    });
  });
}

function qs(params) {
  var out = [];
  Object.keys(params || {}).forEach(function (k) {
    var v = params[k];
    if (v === undefined || v === null || v === '') return;
    out.push(encodeURIComponent(k) + '=' + encodeURIComponent(v));
  });
  return out.length ? '?' + out.join('&') : '';
}

function button(opts) {
  var b = h('button', { type: opts.type || 'button', class: INS_BTN + ' ia-btn--' + (opts.variant || 'ghost') + (opts.class ? ' ' + opts.class : ''), disabled: !!opts.disabled, title: opts.title, 'aria-label': opts.title, onclick: opts.onclick },
    h('span', { text: opts.label }));
  if (opts.data) Object.keys(opts.data).forEach(function (k) { b.setAttribute('data-' + k, opts.data[k]); });
  return b;
}

function busy(btn, on) {
  btn.disabled = !!on;
  btn.classList.toggle('opacity-60', !!on);
}

function notice(kind, text) {
  var color = kind === 'error' ? 'error' : kind === 'success' ? 'success' : kind === 'warning' ? 'warning' : 'secondary';
  return h('div', { class: 'eco-notice eco-notice--' + color + ' flex items-center gap-2 rounded-md px-3 py-2 text-sm', role: kind === 'error' ? 'alert' : 'status' }, h('span', { class: 'flex-1', text: text }));
}

function tag(text, color) {
  return h('span', { class: 'eco-tag eco-tag--' + (color || 'secondary') + ' inline-flex leading-tight px-2 py-0.5 text-xs rounded-md whitespace-nowrap', text: text });
}

function mono(text) {
  return h('span', { class: 'font-mono text-xs', dir: 'ltr', text: text == null ? '' : String(text) });
}

var insToastBox = null;
function toast(kind, text) {
  if (!insToastBox || !insToastBox.isConnected) {
    insToastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-ins': 'toasts' });
    document.body.appendChild(insToastBox);
  }
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  insToastBox.appendChild(el);
  setTimeout(function () { el.remove(); }, kind === 'error' ? 6000 : 3000);
}

function errorText(err) {
  if (!err) return '';
  var code = err.code || '';
  return (err.message && err.message !== code ? err.message + ' ' : '') + (code ? '(' + code + ')' : '');
}

function card(children, cls) {
  return h('div', { class: 'bg-white dark:bg-secondary-800 rounded-md shadow' + (cls ? ' ' + cls : '') }, children);
}

function field(label, control, hint, extra) {
  var err = h('p', { class: 'ia-field-error eco-tf-err text-xs text-error-600 dark:text-error-400', hidden: true });
  var wrap = h('label', { class: 'ia-field eco-tbf block mb-4' + (extra ? ' ' + extra : '') },
    label ? h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 eco-tf-label', text: label }) : null,
    control,
    hint ? h('span', { class: 'block text-xs text-secondary-500 dark:text-secondary-400 eco-tf-hint', text: hint }) : null,
    err);
  wrap.setError = function (msg) { err.textContent = msg || ''; err.hidden = !msg; };
  if (label && control && control.tagName === 'INPUT' && control.getAttribute('placeholder') === label) wrap.classList.add('eco-tbf-q');
  return wrap;
}

function input(value, opts) {
  opts = opts || {};
  var el = h(opts.multiline ? 'textarea' : 'input', { class: INS_INPUT + (opts.mono ? ' font-mono' : ''), type: opts.multiline ? null : (opts.type || 'text'), dir: opts.ltr ? 'ltr' : 'auto', rows: opts.multiline ? (opts.rows || 3) : null, placeholder: opts.placeholder, name: opts.name, autocomplete: 'off', 'aria-label': opts.label });
  el.value = value == null ? '' : value;
  return el;
}

function select(value, options, opts) {
  opts = opts || {};
  var el = h('select', { class: INS_INPUT + ' eco-select', name: opts.name, 'aria-label': opts.label });
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) opt.selected = true;
    el.appendChild(opt);
  });
  return el;
}

function check(checked, label, name) {
  var box = h('input', { type: 'checkbox', class: 'ia-check', name: name, checked: !!checked });
  var wrap = h('label', { class: 'ia-checkline flex items-start gap-2 mb-3 cursor-pointer text-sm text-secondary-800 dark:text-secondary-200' }, box, h('span', { text: label }));
  wrap.input = box;
  return wrap;
}

function numberValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? Math.round(n) : NaN;
}

var insPanel = null;
function closePanel() {
  if (insPanel) { insPanel.remove(); insPanel = null; document.removeEventListener('keydown', panelKey); }
}

function panelKey(e) { if (e.key === 'Escape') closePanel(); }

function sidePanel(title, body, footer, name) {
  closePanel();
  var holder = h('div', { class: 'ia-panel-holder' });
  var panel = h('div', { class: 'ia-panel bg-white dark:bg-secondary-900 shadow-lg', role: 'dialog', 'aria-modal': 'true', 'aria-label': title, 'data-ins-panel': name || '' },
    h('div', { class: 'ia-panel-head flex items-center justify-between border-b px-6 py-3' },
      h('h2', { class: 'text-lg font-semibold truncate', text: title }),
      button({ label: '×', title: t('close'), onclick: closePanel, class: 'ia-close' })),
    h('div', { class: 'ia-panel-body p-6' }, body),
    footer ? h('div', { class: 'ia-panel-foot flex gap-2 justify-end bg-secondary-50 dark:bg-secondary-800 p-4' }, footer) : null);
  holder.appendChild(panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) closePanel(); });
  insPanel = h('div', { 'data-ins': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, h('div', { class: 'ia-shade' }), holder);
  document.body.appendChild(insPanel);
  document.addEventListener('keydown', panelKey);
  return panel;
}

function confirmBox(text, okLabel) {
  return new Promise(function (resolve) {
    var root;
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-ins': 'confirm', 'data-l10n-skip': '', class: 'ia-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'ia-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('confirm.title') }),
        h('p', { class: 'text-sm mb-4', text: text }),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: okLabel || t('yes'), variant: 'danger', data: { 'ins-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root);
  });
}

function pageFrame(page, desc, actions) {
  var content = h('div', { class: 'flex flex-col gap-4' });
  var el = h('div', { class: 'ia-page flex flex-col flex-1 px-4 md:px-6 pb-6 max-w-[1600px] w-full mx-auto', 'data-ins-page': page },
    h('div', { class: 'flex flex-col sm:flex-row gap-3 justify-between items-start mb-4' },
      h('div', {},
        h('h1', { class: 'text-xl font-semibold', text: t('nav.' + page) }),
        h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mt-1 max-w-3xl', text: desc })),
      h('div', { class: 'flex flex-wrap gap-2' }, actions || null)),
    content);
  return { el: el, content: content };
}

function table(columns, rows, onrow) {
  var thead = h('thead', {}, h('tr', {}, columns.map(function (c) { return h('th', { class: 'p-2.5 text-start text-sm font-medium', text: c.label }); })));
  var tbody = h('tbody', {});
  rows.forEach(function (row) {
    var tr = h('tr', { class: 'border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' + (onrow ? ' cursor-pointer' : ''), 'data-ins-row': row.id || '' });
    if (onrow) {
      tr.tabIndex = 0;
      tr.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,select,label')) onrow(row.item); });
      tr.addEventListener('keydown', function (e) { if (e.key === 'Enter') onrow(row.item); });
    }
    row.cells.forEach(function (cell, i) { tr.appendChild(h('td', { class: 'px-3 py-2 align-middle text-sm', 'data-label': columns[i].label }, cell)); });
    tbody.appendChild(tr);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow rounded-md overflow-x-auto' }, h('table', { class: 'w-full eco-table ia-table' }, thead, tbody));
}

function pager(state, onpage) {
  var total = state.total || 0;
  var nav = h('nav', { class: 'eco-pager', 'aria-label': t('pager') });
  if (!total) { nav.hidden = true; return nav; }
  var from = (state.page - 1) * state.limit + 1;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  function btn(name, path, label, disabled, go) {
    var b = h('button', { type: 'button', class: 'eco-pager-btn', title: label, 'aria-label': label, disabled: !!disabled, onclick: go, 'data-page': name });
    b.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4" aria-hidden="true" data-l10n-icon="mirror">' + path + '</svg>';
    return b;
  }
  nav.appendChild(h('span', { class: 'eco-pager-range tabular-nums', 'data-ins-range': '', text: t('range', { from: from, to: to, total: total }) }));
  if (last > 1) {
    nav.appendChild(h('div', { class: 'eco-pager-nav' },
      btn('first', '<path d="M7 6v12M17 18l-6-6 6-6"/>', t('page.first'), state.page <= 1, function () { onpage(1); }),
      btn('prev', '<path d="M15 18l-6-6 6-6"/>', t('prev'), state.page <= 1, function () { onpage(state.page - 1); }),
      h('span', { class: 'eco-pager-page tabular-nums', 'aria-current': 'page', text: t('page.of', { page: state.page, pages: last }) }),
      btn('next', '<path d="M9 18l6-6-6-6"/>', t('next'), state.page >= last, function () { onpage(state.page + 1); }),
      btn('last', '<path d="M17 6v12M7 18l6-6-6-6"/>', t('page.last'), state.page >= last, function () { onpage(last); })));
  }
  return nav;
}

function loadSettings(base) {
  return api('GET', base + '/settings');
}

function floatValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).replace(/٫/g, '.').trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? n : NaN;
}

var insLanguages = null;
function loadLanguages() {
  if (insLanguages) return Promise.resolve(insLanguages);
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch('/v2/console/base/languages', { headers: headers, credentials: 'same-origin' }).then(function (r) { return r.ok ? r.json() : null; }).then(function (d) {
    var list = d && d.items ? d.items.filter(function (l) { return l.isActive; }).map(function (l) { return { code: l.code, name: l.nativeName || l.code, isDefault: !!l.isDefault, direction: l.direction }; }) : [];
    if (!list.length) list = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    list.sort(function (a, b) { return (b.isDefault ? 1 : 0) - (a.isDefault ? 1 : 0); });
    insLanguages = list;
    return list;
  }, function () {
    insLanguages = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    return insLanguages;
  });
}

function translated(label, name, value, opts) {
  opts = opts || {};
  var inputs = {};
  var box = h('div', { class: 'grid sm:grid-cols-2 gap-x-4', 'data-ins-text': name });
  (insLanguages || []).forEach(function (l) {
    var el = input(value && value[l.code] || '', { name: name + '-' + l.code, multiline: opts.multiline, rows: 2 });
    el.setAttribute('dir', l.direction === 'rtl' ? 'rtl' : 'ltr');
    inputs[l.code] = el;
    box.appendChild(field(t('lang.text', { label: label, lang: l.name }), el));
  });
  box.read = function () {
    var out = {};
    Object.keys(inputs).forEach(function (k) { var v = inputs[k].value.trim(); if (v) out[k] = v; });
    return out;
  };
  return box;
}

function textOf(m, fallback) {
  if (!m) return fallback || '';
  var l = insLang();
  return m[l] || m.fa || m.en || Object.keys(m).map(function (k) { return m[k]; }).filter(Boolean)[0] || fallback || '';
}

function fieldError(msgBox, err) {
  clear(msgBox).appendChild(notice('error', insErrText(err)));
}

function options(list, value) {
  return list.map(function (o) { return typeof o === 'string' ? { value: o, label: o } : o; });
}

function dateInput(value, name, type) {
  if (window.NakamaDate && (!type || type === 'date')) return window.NakamaDate.input({ mode: 'date', value: value || null, name: name, label: name, inputClass: INS_INPUT });
  var el = h('input', { type: type || 'date', class: INS_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  el.value = value || '';
  return el;
}

function isoDay(d) {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

function insErrText(err) {
  if (err && err.code && has('code.' + err.code)) return t('code.' + err.code, err.meta || null) + (err.field ? ' · ' + err.field : '');
  return errorText(err) + (err && err.field ? ' · ' + err.field : '');
}

function section(title, children, name) {
  return card(h('div', { class: 'p-4', 'data-ins-section': name || '' },
    h('h2', { class: 'font-semibold mb-3', text: title }), children));
}

function settingsPage(root, page, base, build) {
  var frame = pageFrame(page, t('desc.' + page));
  clear(root).appendChild(frame.el);
  frame.content.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  Promise.all([loadSettings(base), loadLanguages()]).then(function (res) {
    var d = res[0];
    var cur = JSON.parse(JSON.stringify(d.settings));
    var form = build(cur, d.defaults, d);
    var msg = h('div', { 'data-ins-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', 'data-ins-updated': '' });
    var paintMeta = function (x) { meta.textContent = x.updatedBy ? t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }) : ''; };
    paintMeta(d);
    var save = button({ label: t('save'), variant: 'primary', data: { 'ins-save': page }, onclick: function () {
      busy(save, true);
      clear(msg);
      var bad = form.read(cur);
      if (bad) { busy(save, false); msg.appendChild(notice('error', bad)); return; }
      api('PUT', base + '/settings', cur).then(function (out) {
        busy(save, false);
        cur = JSON.parse(JSON.stringify(out.settings));
        paintMeta(out);
        toast('success', t('saved'));
        if (form.after) form.after(out);
      }, function (err) {
        busy(save, false);
        msg.appendChild(notice('error', insErrText(err)));
      });
    } });
    clear(frame.content);
    form.el.forEach(function (el) { frame.content.appendChild(el); });
    frame.content.appendChild(h('div', { class: 'ia-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), save));
  }, function (err) {
    clear(frame.content).appendChild(notice('error', insErrText(err)));
  });
}

function numInput(value, name, label) {
  var el = input(value == null ? '' : digits(String(value)), { ltr: true, name: name, label: label });
  el.setAttribute('inputmode', 'numeric');
  return el;
}

function hashQuery() {
  var i = location.hash.indexOf('?');
  var out = {};
  if (i === -1) return out;
  location.hash.slice(i + 1).split('&').forEach(function (kv) {
    var p = kv.split('=');
    if (p[0]) out[decodeURIComponent(p[0])] = decodeURIComponent(p[1] || '');
  });
  return out;
}

function playerLink(id, name) {
  if (!id) return h('span', { class: 'text-secondary-400', text: '—' });
  return h('a', { href: '#/players/' + id, class: 'eco-link', 'data-ins-player-link': id, dir: 'auto', text: name || id.slice(0, 8) });
}

var INS_SVG = 'http://www.w3.org/2000/svg';
var INS_SERIES = 8;

function s(tag, attrs) {
  var el = document.createElementNS(INS_SVG, tag);
  Object.keys(attrs || {}).forEach(function (k) { if (attrs[k] != null) el.setAttribute(k, attrs[k]); });
  for (var i = 2; i < arguments.length; i++) if (arguments[i]) el.appendChild(arguments[i]);
  return el;
}

function seriesColor(i) {
  return 'var(--ins-s' + ((i % INS_SERIES) + 1) + ')';
}

function niceMax(v) {
  if (!(v > 0)) return 1;
  var p = Math.pow(10, Math.floor(Math.log10(v)));
  var n = v / p;
  var step = n <= 1 ? 1 : n <= 2 ? 2 : n <= 5 ? 5 : 10;
  return step * p;
}

function shortNum(v) {
  var a = Math.abs(v);
  var out = a >= 1e9 ? (v / 1e9).toFixed(1) + 'B' : a >= 1e6 ? (v / 1e6).toFixed(1) + 'M' : a >= 1e4 ? (v / 1e3).toFixed(0) + 'K' : String(Math.round(v * 100) / 100);
  return digits(out.replace('.0', ''));
}

function legend(series) {
  if (series.length < 2) return null;
  return h('div', { class: 'ia-legend flex flex-wrap gap-x-4 gap-y-1 text-xs text-secondary-600 dark:text-secondary-300', 'data-ins-legend': '' },
    series.map(function (sr, i) {
      return h('span', { class: 'inline-flex items-center gap-1.5' }, h('span', { class: 'ia-swatch', style: 'background:' + seriesColor(i) }), h('span', { dir: 'auto', text: sr.label }));
    }));
}

function tooltipBox() {
  return h('div', { class: 'ia-tip', role: 'status', hidden: true });
}

function lineChart(labels, series, opts) {
  opts = opts || {};
  var W = 720, H = opts.height || 240, L = 48, R = 12, T = 12, B = 28;
  var max = 0;
  var ints = true;
  series.forEach(function (sr) { sr.values.forEach(function (v) { if (v > max) max = v; if (v % 1) ints = false; }); });
  var step = niceMax(max / 4);
  if (ints && step < 1) step = 1;
  max = step * 4;
  var n = labels.length;
  var x = function (i) { return n <= 1 ? L + (W - L - R) / 2 : L + i * (W - L - R) / (n - 1); };
  var y = function (v) { return T + (H - T - B) * (1 - v / max); };
  var svg = s('svg', { viewBox: '0 0 ' + W + ' ' + H, class: 'ia-chart', role: 'img', 'aria-label': opts.label || '', preserveAspectRatio: 'none' });
  for (var g = 0; g <= 4; g++) {
    var gv = max * g / 4;
    svg.appendChild(s('line', { x1: L, x2: W - R, y1: y(gv), y2: y(gv), class: 'ia-grid' }));
    var tx = s('text', { x: L - 6, y: y(gv) + 4, class: 'ia-axis', 'text-anchor': 'end' });
    tx.textContent = shortNum(gv);
    svg.appendChild(tx);
  }
  var every = Math.max(1, Math.ceil(n / 8));
  labels.forEach(function (lb, i) {
    if (i % every !== 0 && i !== n - 1) return;
    var tx = s('text', { x: x(i), y: H - 8, class: 'ia-axis', 'text-anchor': 'middle' });
    tx.textContent = opts.fmtLabel ? opts.fmtLabel(lb) : lb;
    svg.appendChild(tx);
  });
  series.forEach(function (sr, si) {
    var d = sr.values.map(function (v, i) { return (i ? 'L' : 'M') + x(i).toFixed(1) + ' ' + y(v || 0).toFixed(1); }).join(' ');
    svg.appendChild(s('path', { d: d, class: 'ia-line', stroke: seriesColor(si), 'data-ins-series': sr.label }));
    if (n <= 40) sr.values.forEach(function (v, i) { svg.appendChild(s('circle', { cx: x(i), cy: y(v || 0), r: 3, fill: seriesColor(si), class: 'ia-dot' })); });
  });
  var cross = s('line', { y1: T, y2: H - B, class: 'ia-cross', visibility: 'hidden' });
  svg.appendChild(cross);
  var tip = tooltipBox();
  var wrap = h('div', { class: 'ia-chartwrap relative', 'data-ins-chart': opts.name || 'line' }, svg, tip);
  svg.addEventListener('mousemove', function (e) {
    var r = svg.getBoundingClientRect();
    var px = (e.clientX - r.left) / r.width * W;
    var i = n <= 1 ? 0 : Math.max(0, Math.min(n - 1, Math.round((px - L) / ((W - L - R) / (n - 1)))));
    cross.setAttribute('x1', x(i));
    cross.setAttribute('x2', x(i));
    cross.setAttribute('visibility', 'visible');
    clear(tip).appendChild(h('div', { class: 'font-medium mb-1', text: opts.fmtLabel ? opts.fmtLabel(labels[i]) : labels[i] }));
    series.forEach(function (sr, si) {
      tip.appendChild(h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'ia-swatch', style: 'background:' + seriesColor(si) }), h('span', { class: 'flex-1', dir: 'auto', text: sr.label }), h('span', { class: 'font-medium', text: fmtNum(sr.values[i]) })));
    });
    tip.hidden = false;
    var left = x(i) / W * r.width;
    tip.style.left = Math.min(Math.max(0, left + 12), r.width - 180) + 'px';
    tip.style.top = '8px';
  });
  svg.addEventListener('mouseleave', function () { tip.hidden = true; cross.setAttribute('visibility', 'hidden'); });
  return h('div', { class: 'flex flex-col gap-2' }, legend(series), wrap);
}

function barList(rows, opts) {
  opts = opts || {};
  var max = 0;
  rows.forEach(function (r) { if (r.value > max) max = r.value; });
  var tip = tooltipBox();
  var list = h('div', { class: 'flex flex-col gap-1.5 relative', 'data-ins-chart': opts.name || 'bars', role: 'list' });
  rows.forEach(function (r) {
    var pct = max > 0 ? Math.max(1, r.value / max * 100) : 0;
    var row = h('div', { class: 'ia-barrow', role: 'listitem', tabindex: '0' },
      h('div', { class: 'ia-barlabel text-sm truncate', dir: 'auto', title: r.label, text: r.label }),
      h('div', { class: 'ia-bartrack' }, h('div', { class: 'ia-bar', style: 'width:' + pct + '%;background:' + (r.color || seriesColor(0)) })),
      h('div', { class: 'ia-barvalue text-sm font-medium', text: opts.fmt ? opts.fmt(r.value) : fmtNum(r.value) }));
    var show = function () {
      clear(tip).appendChild(h('div', { class: 'font-medium', dir: 'auto', text: r.label }));
      (r.tip || []).forEach(function (line) { tip.appendChild(h('div', { text: line })); });
      tip.hidden = false;
      tip.style.top = (row.offsetTop + row.offsetHeight) + 'px';
      tip.style.left = '0px';
    };
    row.addEventListener('mouseenter', show);
    row.addEventListener('focus', show);
    row.addEventListener('mouseleave', function () { tip.hidden = true; });
    row.addEventListener('blur', function () { tip.hidden = true; });
    list.appendChild(row);
  });
  list.appendChild(tip);
  return list;
}

function kpis(items) {
  return h('div', { class: 'ia-kpis grid gap-3', 'data-ins-kpis': '' }, items.map(function (k) {
    return card(h('div', { class: 'p-4', 'data-ins-kpi': k.key || '' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: k.label }),
      h('div', { class: 'text-2xl font-semibold mt-1', text: k.value }),
      k.sub ? h('div', { class: 'text-xs text-secondary-500 mt-1', text: k.sub }) : null));
  }));
}

function csvDownload(name, columns, rows) {
  var esc = function (v) {
    var x = v == null ? '' : String(v);
    return /[",\n]/.test(x) ? '"' + x.replace(/"/g, '""') + '"' : x;
  };
  var text = '﻿' + [columns.map(esc).join(',')].concat(rows.map(function (r) { return r.map(esc).join(','); })).join('\n');
  var url = URL.createObjectURL(new Blob([text], { type: 'text/csv;charset=utf-8' }));
  var a = h('a', { href: url, download: name + '.csv' });
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(function () { URL.revokeObjectURL(url); }, 1000);
}

function csvButton(name, get) {
  return button({ label: t('csv'), data: { 'ins-csv': name }, onclick: function () { var d = get(); csvDownload(name, d.columns, d.rows); } });
}

function panelCard(title, body, actions, name) {
  return card(h('div', { class: 'p-4 flex flex-col gap-3', 'data-ins-section': name || '' },
    h('div', { class: 'flex items-center justify-between gap-2 flex-wrap' }, h('h2', { class: 'font-semibold', text: title }), actions ? h('div', { class: 'flex gap-2' }, actions) : null),
    body));
}

function insQuery() {
  var q = hashQuery();
  var today = new Date();
  if (window.NakamaDate && (!q.to || !q.from)) {
    var D = window.NakamaDate.days;
    var tj = D.fromDay(window.NakamaDate.today());
    if (!q.to) q.to = D.toDay(tj);
    if (!q.from) q.from = D.toDay(tj - 29);
  }
  if (!q.to) q.to = isoDay(today);
  if (!q.from) { var f = new Date(today); f.setDate(f.getDate() - 29); q.from = isoDay(f); }
  return q;
}

function setQuery(next) {
  var page = (location.hash.split('?')[0] || '#/insights/overview');
  var parts = [];
  Object.keys(next).forEach(function (k) { if (next[k] !== '' && next[k] != null) parts.push(encodeURIComponent(k) + '=' + encodeURIComponent(next[k])); });
  location.hash = page + (parts.length ? '?' + parts.join('&') : '');
}

function filterQS(q, extra) {
  var p = { from: q.from, to: q.to, platform: q.platform, appVersion: q.appVersion, includeFlagged: q.includeFlagged };
  if (q.game != null && q.game !== '*' && q.game !== undefined) p.game = q.game === '~' ? '' : q.game;
  var out = qs(Object.assign(p, extra || {}));
  if (q.game === '~') out += (out ? '&' : '?') + 'game=';
  return out;
}

var insGames = null;
function loadGames() {
  if (insGames) return Promise.resolve(insGames);
  return api('GET', '/insights/events/catalog').then(function (d) { insGames = d.games || []; return insGames; }, function () { insGames = []; return insGames; });
}

function filterBar(q, opts) {
  opts = opts || {};
  var range = window.NakamaDate ? window.NakamaDate.input({ mode: 'range', value: { from: q.from, to: q.to }, name: 'range', label: t('f.range'), inputClass: INS_INPUT, max: window.NakamaDate.today(), required: true }) : null;
  var from = range ? { get value() { return (range.value && range.value.from) || ''; } } : dateInput(q.from, 'from');
  var to = range ? { get value() { return (range.value && range.value.to) || ''; } } : dateInput(q.to, 'to');
  if (range) {
    range.setAttribute('data-ins-range', '');
    range.input.addEventListener('change', function () { if (range.value && range.value.from && range.value.to && (range.value.from !== q.from || range.value.to !== q.to)) apply(); });
  }
  var game = select(q.game || '*', [{ value: '*', label: t('f.allGames') }, { value: '~', label: t('f.mainGame') }], { name: 'game', label: t('f.game') });
  loadGames().then(function (list) { list.forEach(function (g) { if (g) game.appendChild(h('option', { value: g, text: g, selected: g === q.game })); }); });
  var platform = input(q.platform || '', { ltr: true, name: 'platform', placeholder: t('f.any') });
  var version = input(q.appVersion || '', { ltr: true, name: 'appVersion', placeholder: t('f.any') });
  var flagged = check(q.includeFlagged === 'true', t('f.flagged'), 'includeFlagged');
  var apply = function (extra) {
    setQuery(Object.assign({}, q, { from: from.value, to: to.value, game: game.value === '*' ? '' : game.value, platform: platform.value.trim(), appVersion: version.value.trim(), includeFlagged: flagged.input.checked ? 'true' : '' }, extra || {}));
  };
  var preset = function (n) {
    return button({ label: t('f.days', { n: n }), data: { 'ins-preset': n }, onclick: function () {
      var e = new Date();
      var f = new Date(e);
      f.setDate(f.getDate() - (n - 1));
      from.value = isoDay(f);
      to.value = isoDay(e);
      apply();
    } });
  };
  return card(h('div', { class: 'ia-filters p-3 flex flex-wrap items-end gap-3', 'data-ins-filters': '' },
    range ? field(t('f.range'), range) : field(t('f.from'), from),
    range ? null : field(t('f.to'), to),
    range ? null : h('div', { class: 'flex gap-1 mb-0.5' }, preset(7), preset(30), preset(90)),
    opts.noGame ? null : field(t('f.game'), game), field(t('f.platform'), platform), field(t('f.version'), version),
    opts.flagged ? flagged : null,
    button({ label: t('f.apply'), variant: 'primary', data: { 'ins-apply': '' }, onclick: function () { apply(); } })));
}

function loadingInto(el) {
  clear(el).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
}

function failInto(el, err) {
  clear(el).appendChild(notice('error', insErrText(err)));
}

function storeNote(kind) {
  return h('p', { class: 'text-xs text-secondary-500', 'data-ins-store': kind || '', text: t('store', { s: t('st.' + (kind || 'nakama')) }) });
}

function dataTable(columns, rows, name) {
  if (!rows.length) return h('p', { class: 'text-sm text-secondary-500 p-2', 'data-ins-empty': name || '', text: t('empty') });
  var el = table(columns.map(function (c) { return { label: c }; }), rows.map(function (r, i) { return { id: String(i), cells: r }; }));
  if (name) el.setAttribute('data-ins-table', name);
  return el;
}

function dashFrame(root, page, actions) {
  var frame = pageFrame(page, t('desc.' + page), actions);
  clear(root).appendChild(frame.el);
  return frame;
}

function overviewPage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/overview');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(filterBar(q));
  frame.content.appendChild(body);
  loadingInto(body);
  Promise.all([api('GET', '/insights/activity' + filterQS(q)), api('GET', '/insights/retention' + filterQS(q))]).then(function (res) {
    var d = res[0], r = res[1], a = d.activity;
    clear(body);
    var tiles = [
      { key: 'dau', label: t('k.dau'), value: fmtNum(a.dau) }, { key: 'wau', label: t('k.wau'), value: fmtNum(a.wau) }, { key: 'mau', label: t('k.mau'), value: fmtNum(a.mau) },
      { key: 'stick', label: t('k.stick'), value: fmtPct(d.stickiness) }, { key: 'new', label: t('k.new'), value: fmtNum(a.newUsers) }, { key: 'events', label: t('k.events'), value: fmtNum(a.events) }];
    (r.summary || []).forEach(function (x) { tiles.push({ key: 'd' + x.day, label: t('k.d', { n: x.day }), value: fmtPct(x.rate), sub: fmtNum(x.returned) + ' / ' + fmtNum(x.cohort) }); });
    body.appendChild(kpis(tiles));
    var days = a.days.map(function (x) { return x.day; });
    body.appendChild(panelCard(t('c.dau'), lineChart(days, [{ label: t('c.dau'), values: a.days.map(function (x) { return x.dau; }) }, { label: t('c.new'), values: a.days.map(function (x) { return x.newUsers; }) }], { fmtLabel: fmtDay, name: 'overview-dau' }), null, 'overview-chart'));
    var grid = h('div', { class: 'grid gap-4 lg:grid-cols-2' });
    grid.appendChild(panelCard(t('top.events'), d.topEvents.length ? barList(d.topEvents.map(function (x) { return { label: x.keys[0], value: x.value, tip: [t('c.users') + ': ' + fmtNum(x.users)] }; }), { name: 'top-events' }) : h('p', { class: 'text-sm text-secondary-500', text: t('empty') }),
      csvButton('top-events', function () { return { columns: ['event', 'count', 'users'], rows: d.topEvents.map(function (x) { return [x.keys[0], x.value, x.users]; }) }; }), 'top-events'));
    grid.appendChild(panelCard(t('top.platforms'), d.platforms.filter(function (x) { return x.keys[0]; }).length ? barList(d.platforms.filter(function (x) { return x.keys[0]; }).map(function (x, i) { return { label: x.keys[0] || t('none'), value: x.value, color: seriesColor(i) }; }), { name: 'platforms' }) : h('p', { class: 'text-sm text-secondary-500', text: t('empty') }), null, 'platforms'));
    body.appendChild(grid);
    body.appendChild(storeNote(d.store));
  }, function (err) { failInto(body, err); });
}

function activePage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/active');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(filterBar(q));
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/activity' + filterQS(q)).then(function (d) {
    var a = d.activity;
    clear(body);
    body.appendChild(kpis([{ key: 'dau', label: t('k.dau'), value: fmtNum(a.dau) }, { key: 'wau', label: t('k.wau'), value: fmtNum(a.wau) }, { key: 'mau', label: t('k.mau'), value: fmtNum(a.mau) },
      { key: 'stick', label: t('k.stick'), value: fmtPct(d.stickiness) }, { key: 'users', label: t('k.users'), value: fmtNum(a.users) }, { key: 'new', label: t('k.new'), value: fmtNum(a.newUsers) }]));
    var labels = a.days.map(function (x) { return x.day; });
    body.appendChild(panelCard(t('nav.insights/active'), lineChart(labels, [
      { label: t('c.dau'), values: a.days.map(function (x) { return x.dau; }) },
      { label: t('c.new'), values: a.days.map(function (x) { return x.newUsers; }) },
      { label: t('c.returning'), values: a.days.map(function (x) { return x.returning; }) }], { fmtLabel: fmtDay, name: 'active' }), null, 'active-chart'));
    var rows = a.days.slice().reverse();
    body.appendChild(panelCard(t('c.day'), dataTable([t('c.day'), t('c.dau'), t('c.new'), t('c.returning'), t('c.events')], rows.map(function (x) { return [fmtDay(x.day), fmtNum(x.dau), fmtNum(x.newUsers), fmtNum(x.returning), fmtNum(x.events)]; }), 'active'),
      csvButton('active-users', function () { return { columns: ['day', 'dau', 'new', 'returning', 'events'], rows: rows.map(function (x) { return [x.day, x.dau, x.newUsers, x.returning, x.events]; }) }; }), 'active-table'));
    body.appendChild(storeNote(d.store));
  }, function (err) { failInto(body, err); });
}

function heat(rate) {
  if (rate == null) return '';
  var a = Math.min(1, Math.max(0.06, rate));
  return 'background:color-mix(in srgb, var(--ins-s1) ' + Math.round(a * 85) + '%, transparent)';
}

function retentionPage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/retention');
  var days = input(q.days || '', { ltr: true, name: 'days', placeholder: '1,7,30' });
  var mode = select(q.mode || 'classic', [{ value: 'classic', label: t('r.classic') }, { value: 'rolling', label: t('r.rolling') }], { name: 'mode', label: t('r.mode') });
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(filterBar(q));
  frame.content.appendChild(card(h('div', { class: 'ia-filters p-3 flex flex-wrap items-end gap-3' }, field(t('r.days'), days), field(t('r.mode'), mode),
    button({ label: t('f.apply'), variant: 'primary', data: { 'ins-retention-apply': '' }, onclick: function () { setQuery(Object.assign({}, q, { days: days.value.replace(/[۰-۹]/g, function (c) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(c)); }).replace(/\s/g, ''), mode: mode.value })); } }))));
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/retention' + filterQS(q, { days: q.days, mode: q.mode })).then(function (d) {
    clear(body);
    body.appendChild(kpis(d.summary.map(function (x) { return { key: 'd' + x.day, label: t('k.d', { n: x.day }), value: fmtPct(x.rate), sub: fmtNum(x.returned) + ' / ' + fmtNum(x.cohort) }; })));
    var cols = [t('c.cohort'), t('c.size')].concat(d.days.map(function (n) { return t('k.d', { n: n }); }));
    var rows = d.cohorts.map(function (c) {
      return [fmtDay(c.day), fmtNum(c.size)].concat(c.cells.map(function (cell) {
        return cell.returned == null ? h('span', { class: 'text-secondary-400', text: t('r.pending') }) : h('span', { class: 'ia-heat', style: heat(cell.rate), title: fmtNum(cell.returned), text: fmtPct(cell.rate) });
      }));
    });
    body.appendChild(panelCard(t('nav.insights/retention'), dataTable(cols, rows, 'cohorts'),
      csvButton('retention', function () { return { columns: ['first_day', 'size'].concat(d.days.map(function (n) { return 'd' + n; })), rows: d.cohorts.map(function (c) { return [c.day, c.size].concat(c.cells.map(function (x) { return x.rate == null ? '' : x.rate; })); }) }; }), 'cohorts'));
    body.appendChild(storeNote(d.store));
  }, function (err) { failInto(body, err); });
}

function parseTagText(text) {
  var out = {};
  String(text || '').split(/[;\n]/).forEach(function (part) {
    var kv = part.split('=');
    if (kv.length < 2) return;
    var g = kv[0].trim();
    var ids = kv.slice(1).join('=').split(',').map(function (x) { return x.trim(); }).filter(Boolean);
    if (g && ids.length) out[g] = ids;
  });
  return out;
}

function tagText(tags) {
  return Object.keys(tags || {}).map(function (g) { return g + '=' + tags[g].join(','); }).join(' ; ');
}

function namesList(el) {
  var dl = h('datalist', { id: 'ins-names-' + Math.random().toString(36).slice(2) });
  el.setAttribute('list', dl.id);
  api('GET', '/insights/values' + filterQS(insQuery(), { kind: 'name', limit: 200 })).then(function (d) { (d.items || []).forEach(function (v) { dl.appendChild(h('option', { value: v.value })); }); }, function () {});
  return dl;
}

function explorePage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/explore');
  var names = input(q.name || '', { ltr: true, name: 'name', placeholder: 'DistanceOverall' });
  var metric = select(q.metric || 'sum', ['count', 'sum', 'avg', 'min', 'max', 'users'].map(function (m) { return { value: m, label: t('m.' + m) }; }), { name: 'metric', label: t('ex.metric') });
  var bucket = select(q.bucket || 'none', ['none', 'hour', 'day', 'week'].map(function (b) { return { value: b, label: t('b.' + b) }; }), { name: 'bucket', label: t('ex.bucket') });
  var groupsWanted = (q.groupBy || '').split(',').filter(Boolean);
  var builtins = ['name', 'user', 'platform', 'appVersion', 'source', 'game'];
  var groupSelects = [0, 1, 2, 3].map(function (i) {
    var el = select(groupsWanted[i] || '', [{ value: '', label: t('ex.noGroup') }].concat(builtins.map(function (b) { return { value: b, label: t('g.' + b) }; })), { name: 'groupBy' + i, label: t('ex.group') });
    if (groupsWanted[i] && groupsWanted[i].indexOf('tag:') === 0) el.appendChild(h('option', { value: groupsWanted[i], text: t('g.tag', { g: groupsWanted[i].slice(4) }), selected: true }));
    return el;
  });
  api('GET', '/insights/values' + filterQS(q, { kind: 'tags', limit: 100 })).then(function (d) {
    (d.items || []).forEach(function (v) {
      groupSelects.forEach(function (el, i) {
        if (groupsWanted[i] === 'tag:' + v.value) return;
        el.appendChild(h('option', { value: 'tag:' + v.value, text: t('g.tag', { g: v.value }) }));
      });
    });
  }, function () {});
  var tagF = {};
  Object.keys(q).forEach(function (k) { if (k.indexOf('tag.') === 0) tagF[k.slice(4)] = q[k].split(','); });
  var tags = input(tagText(tagF), { ltr: true, name: 'tags', placeholder: t('ex.filtersHint') });
  var limit = numInput(q.limit || 100, 'limit', t('ex.limit'));
  var body = h('div', { class: 'flex flex-col gap-4' });
  var run = function () {
    var next = { from: q.from, to: q.to, game: q.game, platform: q.platform, appVersion: q.appVersion, includeFlagged: q.includeFlagged, name: names.value.replace(/\s/g, ''), metric: metric.value, bucket: bucket.value,
      groupBy: groupSelects.map(function (el) { return el.value; }).filter(Boolean).filter(function (v, i, a) { return a.indexOf(v) === i; }).join(','), limit: numberValue(limit) || 100 };
    var tf = parseTagText(tags.value);
    Object.keys(tf).forEach(function (g) { next['tag.' + g] = tf[g].join(','); });
    setQuery(next);
  };
  frame.content.appendChild(filterBar(q, { flagged: true }));
  frame.content.appendChild(card(h('div', { class: 'p-3 flex flex-col gap-3', 'data-ins-explore-form': '' },
    h('div', { class: 'ia-filters flex flex-wrap items-end gap-3' }, field(t('ex.names'), names, t('ex.namesHint')), namesList(names), field(t('ex.metric'), metric), field(t('ex.bucket'), bucket), field(t('ex.limit'), limit)),
    h('div', { class: 'ia-filters flex flex-wrap items-end gap-3' }, groupSelects.map(function (el) { return field(t('ex.group'), el); })),
    h('div', { class: 'ia-filters flex flex-wrap items-end gap-3' }, field(t('ex.filters'), tags), button({ label: t('ex.run'), variant: 'primary', data: { 'ins-run': '' }, onclick: run })))));
  frame.content.appendChild(body);
  var params = { name: q.name, metric: q.metric || 'sum', bucket: q.bucket || 'none', groupBy: q.groupBy, limit: q.limit };
  Object.keys(tagF).forEach(function (g) { params['tag.' + g] = tagF[g].join(','); });
  loadingInto(body);
  api('GET', '/insights/explore' + filterQS(q, params)).then(function (d) {
    clear(body);
    var gl = d.groupBy.map(function (g) { return g.indexOf('tag:') === 0 ? t('g.tag', { g: g.slice(4) }) : t('g.' + g); });
    var keyOf = function (r) { return r.keys.length ? r.keys.map(function (k) { return k === '' ? t('none') : k; }).join(' · ') : t('m.' + d.metric); };
    var chart;
    if (d.bucket !== 'none') {
      var labels = [];
      var seriesMap = {};
      var totals = {};
      d.rows.forEach(function (r) {
        if (labels.indexOf(r.bucket) === -1) labels.push(r.bucket);
        var k = keyOf(r);
        totals[k] = (totals[k] || 0) + r.value;
      });
      labels.sort();
      var top = Object.keys(totals).sort(function (a, b) { return totals[b] - totals[a]; }).slice(0, INS_SERIES);
      top.forEach(function (k) { seriesMap[k] = labels.map(function () { return 0; }); });
      d.rows.forEach(function (r) { var k = keyOf(r); if (seriesMap[k]) seriesMap[k][labels.indexOf(r.bucket)] = r.value; });
      chart = lineChart(labels, top.map(function (k) { return { label: k, values: seriesMap[k] }; }), { fmtLabel: d.bucket === 'hour' ? function (x) { return digits(x.slice(5).replace('T', ' ')); } : fmtDay, name: 'explore' });
    } else {
      chart = barList(d.rows.slice(0, 25).map(function (r) { return { label: keyOf(r), value: r.value, tip: [t('c.count') + ': ' + fmtNum(r.count), t('c.users') + ': ' + fmtNum(r.users)] }; }), { name: 'explore' });
    }
    var cols = (d.bucket !== 'none' ? [t('c.bucket')] : []).concat(gl, [t('m.' + d.metric), t('c.count'), t('c.users')]);
    var rows = d.rows.map(function (r) {
      return (d.bucket !== 'none' ? [digits(r.bucket)] : []).concat(r.keys.map(function (k) { return h('span', { dir: 'auto', text: k || t('none') }); }), [fmtNum(r.value), fmtNum(r.count), fmtNum(r.users)]);
    });
    body.appendChild(panelCard(t('nav.insights/explore'), d.rows.length ? chart : h('p', { class: 'text-sm text-secondary-500', text: t('empty') }), null, 'explore-chart'));
    body.appendChild(panelCard(t('c.value'), dataTable(cols, rows, 'explore'),
      csvButton('explore', function () { return { columns: (d.bucket !== 'none' ? ['bucket'] : []).concat(d.groupBy, [d.metric, 'count', 'users']), rows: d.rows.map(function (r) { return (d.bucket !== 'none' ? [r.bucket] : []).concat(r.keys, [r.value, r.count, r.users]); }) }; }), 'explore-table'));
    body.appendChild(storeNote(d.store));
  }, function (err) { failInto(body, err); });
}

function funnelPanel(fn, after) {
  fn = fn || { name: '', game: '', windowSeconds: 86400, steps: [{ name: '' }, { name: '' }] };
  var name = input(fn.name, { name: 'name' });
  var game = input(fn.game, { ltr: true, name: 'game' });
  var win = numInput(Math.round(fn.windowSeconds / 60), 'window', t('fn.window'));
  var rows = h('div', { class: 'ia-rows', 'data-ins-steps': '' });
  var addRow = function (st) {
    var ev = input(st.name || '', { ltr: true, name: 'step-name' });
    var tg = input(tagText(st.tags), { ltr: true, name: 'step-tags', placeholder: t('fn.hintTags') });
    var row = h('div', { class: 'ia-row', 'data-ins-step': '' }, field(t('fn.event'), ev), namesList(ev), field(t('fn.tags'), tg),
      button({ label: t('delete'), variant: 'danger', onclick: function () { if (rows.children.length > 2) row.remove(); } }));
    row.read = function () { var o = { name: ev.value.trim() }; var tags = parseTagText(tg.value); if (Object.keys(tags).length) o.tags = tags; return o; };
    rows.appendChild(row);
  };
  fn.steps.forEach(addRow);
  var msg = h('div', {});
  var save = button({ label: t('save'), variant: 'primary', data: { 'ins-funnel-save': '' }, onclick: function () {
    busy(save, true);
    var body = { name: name.value, game: game.value.trim(), windowSeconds: (numberValue(win) || 0) * 60, steps: Array.prototype.map.call(rows.children, function (r) { return r.read(); }) };
    api(fn.id ? 'PUT' : 'POST', '/insights/funnels' + (fn.id ? '/' + fn.id : ''), body).then(function (out) { closePanel(); toast('success', t('saved')); after(out); }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  var foot = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (fn.id) foot.unshift(button({ label: t('delete'), variant: 'danger', data: { 'ins-funnel-delete': '' }, onclick: function () {
    confirmBox(fn.name).then(function (ok) { if (ok) api('DELETE', '/insights/funnels/' + fn.id).then(function () { closePanel(); toast('success', t('deleted')); after(null); }, function (err) { fieldError(msg, err); }); });
  } }));
  sidePanel(fn.id ? fn.name : t('fn.new'), h('div', {}, field(t('fn.name'), name), field(t('cat.game'), game), field(t('fn.window'), win), h('h3', { class: 'font-semibold mb-2', text: t('fn.steps') }), rows,
    h('div', { class: 'mt-2' }, button({ label: t('fn.addStep'), data: { 'ins-add-step': '' }, onclick: function () { if (rows.children.length < 12) addRow({ name: '' }); } })), msg), foot, 'funnel');
}

function funnelsPage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/funnels', [button({ label: t('fn.new'), variant: 'primary', data: { 'ins-funnel-new': '' }, onclick: function () { funnelPanel(null, function (f) { setQuery(Object.assign({}, q, { funnel: f ? f.id : '' })); }); } })]);
  var layout = h('div', { class: 'ia-split' });
  var list = h('div', { class: 'flex flex-col gap-1', 'data-ins-funnels': '' });
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(filterBar(q));
  layout.appendChild(card(h('div', { class: 'p-3' }, list)));
  layout.appendChild(body);
  frame.content.appendChild(layout);
  loadingInto(list);
  api('GET', '/insights/funnels').then(function (d) {
    clear(list);
    if (!d.items.length) list.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('fn.empty') }));
    d.items.forEach(function (f) {
      list.appendChild(h('a', { href: '#/insights/funnels?' + filterQS(q, { funnel: f.id }).slice(1), class: 'ia-navitem rounded-md px-3 py-2 text-sm' + (q.funnel === f.id ? ' ia-navitem--on' : ''), 'data-ins-funnel': f.id, dir: 'auto', text: f.name }));
    });
    var fn = d.items.filter(function (f) { return f.id === q.funnel; })[0];
    if (!fn) { clear(body).appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('fn.pick') }))); return; }
    loadingInto(body);
    api('GET', '/insights/funnels/' + fn.id + '/run' + filterQS(q)).then(function (r) {
      clear(body);
      var first = r.steps.length ? r.steps[0].users : 0;
      var bars = h('div', { class: 'flex flex-col gap-3', 'data-ins-chart': 'funnel' }, r.steps.map(function (st, i) {
        var pct = first ? st.users / first * 100 : 0;
        return h('div', { class: 'ia-step', 'data-ins-step-result': i },
          h('div', { class: 'flex justify-between gap-2 text-sm mb-1' }, h('span', { class: 'font-medium', dir: 'auto', text: t('fn.step', { n: i + 1 }) + ' · ' + st.name }), h('span', { text: fmtNum(st.users) + ' · ' + fmtPct(st.fromFirst) })),
          h('div', { class: 'ia-bartrack ia-bartrack--tall' }, h('div', { class: 'ia-bar', style: 'width:' + Math.max(pct, st.users ? 1 : 0) + '%;background:' + seriesColor(0) })),
          i ? h('div', { class: 'text-xs text-secondary-500 mt-1', text: t('fn.fromPrev') + ': ' + fmtPct(st.fromPrevious) + ' · ' + t('fn.drop') + ': ' + fmtNum(st.dropOff) + ' · ' + t('fn.median') + ': ' + fmtSeconds(st.medianSeconds) }) : null);
      }));
      body.appendChild(panelCard(fn.name, h('div', { class: 'flex flex-col gap-3' }, h('p', { class: 'text-xs text-secondary-500', text: t('fn.windowV', { n: Math.round(r.windowSeconds / 60) }) }), bars),
        [button({ label: t('edit'), data: { 'ins-funnel-edit': '' }, onclick: function () { funnelPanel(fn, function (f) { setQuery(Object.assign({}, q, { funnel: f ? f.id : '' })); funnelsPage(root); }); } }),
          csvButton('funnel', function () { return { columns: ['step', 'event', 'users', 'from_first', 'from_previous', 'drop_off', 'median_seconds'], rows: r.steps.map(function (st, i) { return [i + 1, st.name, st.users, st.fromFirst, st.fromPrevious == null ? '' : st.fromPrevious, st.dropOff, st.medianSeconds == null ? '' : st.medianSeconds]; }) }; })], 'funnel-result'));
      body.appendChild(storeNote(r.store));
    }, function (err) { failInto(body, err); });
  }, function (err) { failInto(list, err); });
}

function timelinePanel(sess) {
  var body = h('div', { class: 'flex flex-col gap-3' });
  sidePanel(t('s.timeline'), body, [button({ label: t('close'), onclick: closePanel })], 'timeline');
  loadingInto(body);
  api('GET', '/insights/sessions/' + sess.id + '/timeline').then(function (d) {
    clear(body);
    var se = d.session;
    body.appendChild(h('dl', { class: 'ia-dl text-sm' },
      h('dt', { text: t('s.player') }), h('dd', {}, playerLink(se.userId, sess.username || sess.displayName)),
      h('dt', { text: t('s.started') }), h('dd', { text: fmtDate(se.startedAt) }),
      h('dt', { text: t('s.length') }), h('dd', { text: fmtSeconds(se.seconds) }),
      h('dt', { text: t('f.platform') }), h('dd', { dir: 'ltr', text: [se.platform, se.appVersion].filter(Boolean).join(' · ') || t('none') })));
    if (d.matchedBy === 'time' && d.events.length) body.appendChild(notice('warning', t('s.byTime')));
    if (!d.events.length) { body.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('s.noEvents') })); return; }
    body.appendChild(h('ol', { class: 'ia-timeline', 'data-ins-timeline': '' }, d.events.map(function (e) {
      return h('li', { class: 'ia-tl-item', 'data-ins-event': e.name },
        h('span', { class: 'ia-tl-time text-xs text-secondary-500', text: fmtTime(e.at) }),
        h('div', { class: 'flex flex-col gap-1 min-w-0' },
          h('div', { class: 'flex flex-wrap items-center gap-2' }, h('span', { class: 'font-medium font-mono text-sm', dir: 'ltr', text: e.name }), e.value != null ? tag(fmtNum(e.value), 'primary') : null, tag(e.source, 'secondary'), e.flagged ? tag(e.flagReason || 'flagged', 'warning') : null),
          Object.keys(e.tags || {}).length ? h('div', { class: 'flex flex-wrap gap-1' }, Object.keys(e.tags).map(function (g) { return h('span', { class: 'ia-chip', dir: 'ltr', text: g + ': ' + e.tags[g] }); })) : null,
          e.fields ? h('div', { class: 'text-xs text-secondary-500 font-mono', dir: 'ltr', text: JSON.stringify(e.fields) }) : null));
    })));
  }, function (err) { failInto(body, err); });
}

function sessionsPage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/sessions');
  var user = input(q.userId || '', { ltr: true, name: 'userId', placeholder: t('s.userId') });
  var status = select(q.status || '', [{ value: '', label: t('all') }, { value: 'open', label: t('s.open') }, { value: 'ended', label: t('s.ended') }], { name: 'status', label: t('s.status') });
  var platform = input(q.platform || '', { ltr: true, name: 'platform', placeholder: t('f.any') });
  var body = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx ia-filters' }, field(t('s.userId'), user), field(t('s.status'), status), field(t('f.platform'), platform),
    button({ label: t('f.apply'), variant: 'primary', data: { 'ins-apply': '' }, onclick: function () { setQuery({ userId: user.value.trim(), status: status.value, platform: platform.value.trim(), page: '' }); } })));
  frame.content.appendChild(body);
  var page = Number(q.page) || 1;
  loadingInto(body);
  api('GET', '/insights/sessions' + qs({ userId: q.userId, status: q.status, platform: q.platform, page: page, limit: 25 })).then(function (d) {
    clear(body);
    var rows = d.items.map(function (x) {
      return { id: x.id, item: x, cells: [playerLink(x.userId, x.username || x.displayName), h('span', { dir: 'ltr', text: [x.platform, x.appVersion].filter(Boolean).join(' · ') || t('none') }), fmtDate(x.startedAt), fmtSeconds(x.seconds),
        tag(x.endedAt ? t('s.ended') : t('s.open'), x.endedAt ? 'secondary' : 'success')] };
    });
    if (!rows.length) body.appendChild(h('p', { class: 'text-sm text-secondary-500', 'data-ins-empty': 'sessions', text: t('empty') }));
    else {
      var tb = table([{ label: t('s.player') }, { label: t('f.platform') }, { label: t('s.started') }, { label: t('s.length') }, { label: t('s.status') }], rows, timelinePanel);
      tb.setAttribute('data-ins-table', 'sessions');
      body.appendChild(tb);
    }
    body.appendChild(h('div', { class: 'flex justify-between items-center gap-2' },
      csvButton('sessions', function () { return { columns: ['session', 'user', 'platform', 'version', 'started', 'seconds', 'ended'], rows: d.items.map(function (x) { return [x.sessionId, x.userId, x.platform, x.appVersion, x.startedAt, x.seconds == null ? '' : x.seconds, x.endedAt || '']; }) }; }),
      pager({ page: d.page, limit: d.limit, total: d.total }, function (p) { setQuery(Object.assign({}, q, { page: p })); })));
  }, function (err) { failInto(body, err); });
}

function listText(v) { return (v || []).join(', '); }

function listOf(text) { return String(text || '').split(',').map(function (x) { return x.trim(); }).filter(Boolean); }

function gameQS(game) { return '?game=' + encodeURIComponent(game || ''); }

function groupPanel(game, g, after) {
  g = g || { key: '', nameFa: '', nameEn: '', description: '', closed: false, enabled: true, values: [] };
  var key = input(g.key, { ltr: true, name: 'key' });
  if (g.key) key.disabled = true;
  var nameFa = input(g.nameFa, { name: 'nameFa' });
  var nameEn = input(g.nameEn, { ltr: true, name: 'nameEn' });
  var desc = input(g.description, { name: 'description', multiline: true, rows: 2 });
  var closed = check(g.closed, t('cat.closed'), 'closed');
  var enabled = check(g.enabled, t('cat.enabled'), 'enabled');
  var values = input(g.values.map(function (v) { return [v.id, v.labelFa, v.labelEn].join(','); }).join('\n'), { ltr: true, name: 'values', multiline: true, rows: 8 });
  var msg = h('div', {});
  var save = button({ label: t('save'), variant: 'primary', data: { 'ins-group-save': '' }, onclick: function () {
    busy(save, true);
    var vals = values.value.split('\n').map(function (l) { return l.trim(); }).filter(Boolean).map(function (l) { var p = l.split(','); return { id: p[0].trim(), labelFa: (p[1] || '').trim(), labelEn: (p[2] || '').trim() }; });
    api('PUT', '/insights/events/catalog/groups/' + encodeURIComponent(key.value.trim()) + gameQS(game), { nameFa: nameFa.value, nameEn: nameEn.value, description: desc.value, closed: closed.input.checked, enabled: enabled.input.checked, values: vals })
      .then(function () { closePanel(); toast('success', t('saved')); after(); }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  var foot = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (g.key) foot.unshift(button({ label: t('delete'), variant: 'danger', data: { 'ins-group-delete': '' }, onclick: function () {
    confirmBox(g.key).then(function (ok) { if (ok) api('DELETE', '/insights/events/catalog/groups/' + encodeURIComponent(g.key) + gameQS(game)).then(function () { closePanel(); toast('success', t('deleted')); after(); }, function (err) { fieldError(msg, err); }); });
  } }));
  sidePanel(g.key || t('cat.newGroup'), h('div', {}, field(t('cat.key'), key, t('cat.keyHint')), h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('cat.nameFa'), nameFa), field(t('cat.nameEn'), nameEn)),
    field(t('cat.desc'), desc), closed, enabled, field(t('cat.values'), values, t('cat.valuesHint')), msg), foot, 'group');
}

function eventPanel(game, e, after) {
  e = e || { name: '', labelFa: '', labelEn: '', description: '', enabled: true, pathGroups: [], requiredTags: [], allowedTags: [], value: {} };
  var v = e.value || {};
  var name = input(e.name, { ltr: true, name: 'name' });
  if (e.name) name.disabled = true;
  var labelFa = input(e.labelFa, { name: 'labelFa' });
  var labelEn = input(e.labelEn, { ltr: true, name: 'labelEn' });
  var desc = input(e.description, { name: 'description', multiline: true, rows: 2 });
  var enabled = check(e.enabled, t('cat.enabled'), 'enabled');
  var path = input(listText(e.pathGroups), { ltr: true, name: 'pathGroups' });
  var req = input(listText(e.requiredTags), { ltr: true, name: 'requiredTags' });
  var allowed = input(listText(e.allowedTags), { ltr: true, name: 'allowedTags' });
  var vreq = check(v.required, t('cat.valueRequired'), 'valueRequired');
  var vmin = input(v.min == null ? '' : digits(String(v.min)), { ltr: true, name: 'min' });
  var vmax = input(v.max == null ? '' : digits(String(v.max)), { ltr: true, name: 'max' });
  var unit = input(v.unit || '', { name: 'unit' });
  var msg = h('div', {});
  var save = button({ label: t('save'), variant: 'primary', data: { 'ins-event-save': '' }, onclick: function () {
    busy(save, true);
    var mn = floatValue(vmin), mx = floatValue(vmax);
    api('PUT', '/insights/events/catalog/events/' + encodeURIComponent(name.value.trim()) + gameQS(game), { labelFa: labelFa.value, labelEn: labelEn.value, description: desc.value, enabled: enabled.input.checked,
      pathGroups: listOf(path.value), requiredTags: listOf(req.value), allowedTags: listOf(allowed.value), value: { required: vreq.input.checked, min: mn, max: mx, unit: unit.value } })
      .then(function () { closePanel(); toast('success', t('saved')); after(); }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  var foot = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (e.name) foot.unshift(button({ label: t('delete'), variant: 'danger', data: { 'ins-event-delete': '' }, onclick: function () {
    confirmBox(e.name).then(function (ok) { if (ok) api('DELETE', '/insights/events/catalog/events/' + encodeURIComponent(e.name) + gameQS(game)).then(function () { closePanel(); toast('success', t('deleted')); after(); }, function (err) { fieldError(msg, err); }); });
  } }));
  sidePanel(e.name || t('cat.newEvent'), h('div', {}, field(t('cat.name'), name), h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('cat.nameFa'), labelFa), field(t('cat.nameEn'), labelEn)), field(t('cat.desc'), desc), enabled,
    field(t('cat.path'), path, t('cat.pathHint')), field(t('cat.required'), req), field(t('cat.allowed'), allowed), vreq,
    h('div', { class: 'grid sm:grid-cols-3 gap-x-4' }, field(t('cat.min'), vmin), field(t('cat.max'), vmax), field(t('cat.unit'), unit)), msg), foot, 'event');
}

function catalogPage(root) {
  var q = hashQuery();
  var game = q.game || '';
  var frame = dashFrame(root, 'insights/catalog');
  var gameIn = input(game, { ltr: true, name: 'game', placeholder: t('f.mainGame') });
  var body = h('div', { class: 'flex flex-col gap-4' });
  var reload = function () { catalogPage(root); };
  frame.content.appendChild(card(h('div', { class: 'ia-filters p-3 flex flex-wrap items-end gap-3' }, field(t('cat.game'), gameIn),
    button({ label: t('f.apply'), variant: 'primary', data: { 'ins-apply': '' }, onclick: function () { setQuery({ game: gameIn.value.trim() }); } }))));
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/events/catalog' + gameQS(game)).then(function (d) {
    clear(body);
    var gl = d.groups.map(function (g) { return { id: g.key, item: g, cells: [mono(g.key), h('span', { dir: 'auto', text: (insLang() === 'fa' ? g.nameFa || g.nameEn : g.nameEn || g.nameFa) || t('none') }), tag(g.closed ? t('cat.closedTag') : t('cat.open'), g.closed ? 'warning' : 'secondary'), t('cat.valuesN', { n: g.values.length }), tag(g.enabled ? t('cat.enabled') : t('no'), g.enabled ? 'success' : 'secondary')] }; });
    var gt = gl.length ? table([{ label: t('cat.key') }, { label: t('cat.label') }, { label: t('cat.closed') }, { label: t('cat.values') }, { label: t('cat.enabled') }], gl, function (g) { groupPanel(game, g, reload); }) : h('p', { class: 'text-sm text-secondary-500', text: t('cat.emptyGroups') });
    if (gl.length) gt.setAttribute('data-ins-table', 'groups');
    body.appendChild(panelCard(t('cat.groups'), gt, [button({ label: t('cat.newGroup'), variant: 'primary', data: { 'ins-group-new': '' }, onclick: function () { groupPanel(game, null, reload); } })], 'groups'));
    var range = function (v) { v = v || {}; return v.min == null && v.max == null ? t('none') : digits((v.min == null ? '…' : v.min) + ' – ' + (v.max == null ? '…' : v.max)) + (v.unit ? ' ' + v.unit : ''); };
    var el = d.events.map(function (e) { return { id: e.name, item: e, cells: [mono(e.name), h('span', { dir: 'auto', text: (insLang() === 'fa' ? e.labelFa || e.labelEn : e.labelEn || e.labelFa) || t('none') }), mono(listText(e.pathGroups) || t('none')), mono(listText(e.requiredTags) || t('none')), range(e.value), tag(e.enabled ? t('cat.enabled') : t('no'), e.enabled ? 'success' : 'secondary')] }; });
    var et = el.length ? table([{ label: t('cat.name') }, { label: t('cat.label') }, { label: t('cat.path') }, { label: t('cat.required') }, { label: t('c.value') }, { label: t('cat.enabled') }], el, function (e) { eventPanel(game, e, reload); }) : h('p', { class: 'text-sm text-secondary-500', text: t('cat.emptyEvents') });
    if (el.length) et.setAttribute('data-ins-table', 'events');
    body.appendChild(panelCard(t('cat.events'), et, [button({ label: t('cat.newEvent'), variant: 'primary', data: { 'ins-event-new': '' }, onclick: function () { eventPanel(game, null, reload); } })], 'events'));
  }, function (err) { failInto(body, err); });
}

var INS_DRIVER_FIELDS = {
  postgres: [['host', 'st.host'], ['port', 'st.port', 'num'], ['database', 'st.database'], ['user', 'st.user'], ['sslMode', 'st.sslMode', ['require', 'verify-full', 'verify-ca', 'prefer', 'disable']], ['schema', 'st.schema']],
  clickhouse: [['url', 'st.url'], ['database', 'st.database'], ['table', 'st.table'], ['user', 'st.user']],
  mongodb: [['hosts', 'st.hosts'], ['database', 'st.database'], ['collection', 'st.collection'], ['user', 'st.user'], ['authSource', 'st.authSource'], ['tls', 'st.tls', 'bool']]
};

function storagePage(root) {
  var frame = dashFrame(root, 'insights/storage');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(body);
  loadingInto(body);
  Promise.all([api('GET', '/insights/storage'), api('GET', '/insights/outbox')]).then(function (res) {
    var st = res[0], ob = res[1];
    clear(body);
    var driver = select(st.driver, st.drivers.map(function (d) { return { value: d, label: t('st.' + d) }; }), { name: 'driver', label: t('st.driver') });
    var formBox = h('div', {});
    var controls = {};
    var secret = h('input', { type: 'password', class: INS_INPUT, name: 'secret', dir: 'ltr', autocomplete: 'new-password', placeholder: st.hasSecret ? t('st.secretKeep') : '' });
    var paint = function () {
      clear(formBox);
      controls = {};
      var fields = INS_DRIVER_FIELDS[driver.value];
      if (!fields) return;
      var cfg = (st.config || {})[driver.value] || {};
      var grid = h('div', { class: 'grid sm:grid-cols-2 gap-x-4' });
      fields.forEach(function (f) {
        var el;
        if (f[2] === 'bool') { el = check(cfg[f[0]], t(f[1]), f[0]); controls[f[0]] = { read: function () { return el.input.checked; } }; grid.appendChild(el); return; }
        if (Array.isArray(f[2])) el = select(cfg[f[0]] || f[2][0], f[2], { name: f[0], label: t(f[1]) });
        else el = input(cfg[f[0]] == null ? '' : String(cfg[f[0]]), { ltr: true, name: f[0] });
        controls[f[0]] = { read: function () { return f[2] === 'num' ? numberValue(el) || 0 : el.value.trim(); } };
        grid.appendChild(field(t(f[1]), el));
      });
      formBox.appendChild(grid);
      formBox.appendChild(field(t('st.secret'), secret, st.hasSecret ? t('st.secretSet') + ' · ' + t('st.secretKeep') : null));
      if (!st.secretKey) formBox.appendChild(notice('warning', t('st.noKey')));
    };
    driver.addEventListener('change', paint);
    paint();
    var body2 = function () {
      var cfg = {};
      Object.keys(controls).forEach(function (k) { cfg[k] = controls[k].read(); });
      var out = { driver: driver.value, config: {} };
      if (INS_DRIVER_FIELDS[driver.value]) out.config[driver.value] = cfg;
      if (secret.value) out.secret = secret.value;
      return out;
    };
    var msg = h('div', { 'data-ins-msg': '' });
    var test = button({ label: t('st.test'), data: { 'ins-storage-test': '' }, onclick: function () {
      busy(test, true);
      clear(msg);
      api('POST', '/insights/storage/test', body2()).then(function (r) { busy(test, false); msg.appendChild(notice(r.ok ? 'success' : 'error', r.ok ? t('st.testOk', { ms: r.millis }) : t('st.testFail', { e: r.error }))); }, function (err) { busy(test, false); fieldError(msg, err); });
    } });
    var save = button({ label: t('save'), variant: 'primary', data: { 'ins-storage-save': '' }, onclick: function () {
      busy(save, true);
      clear(msg);
      api('PUT', '/insights/storage', body2()).then(function () { toast('success', t('saved')); storagePage(root); }, function (err) { busy(save, false); fieldError(msg, err); });
    } });
    var status = st.status === 'error' ? notice('error', t('st.error', { e: st.error })) : null;
    body.appendChild(panelCard(t('st.driver'), h('div', {}, field(t('st.driver'), driver), formBox, h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('st.switchNote') }), status, msg,
      h('div', { class: 'flex gap-2 justify-end' }, test, save), st.updatedBy ? h('p', { class: 'text-xs text-secondary-500', text: t('updated', { by: st.updatedBy, at: fmtDate(st.updatedAt) }) }) : null), null, 'storage'));
    var tiles = kpis([{ key: 'pending', label: t('ob.pending'), value: fmtNum(ob.pending), sub: t('set.maxPendingEvents') + ': ' + fmtNum(ob.maxPendingEvents) }, { key: 'rows', label: t('ob.rows'), value: fmtNum(ob.rows) },
      { key: 'dead', label: t('ob.dead'), value: fmtNum(ob.dead), sub: fmtNum(ob.deadEvents) + ' ' + t('c.events') }, { key: 'oldest', label: t('ob.oldest'), value: ob.oldest ? fmtDate(ob.oldest) : t('none') }]);
    var failing = ob.failing.length ? dataTable(['#', t('s.source'), t('c.events'), t('ob.attempts'), t('ob.error'), t('ob.next')], ob.failing.map(function (f) {
      return [digits(String(f.id)), f.source, fmtNum(f.count), fmtNum(f.attempts), h('span', { class: 'text-xs font-mono', dir: 'ltr', text: f.lastError }), f.dead ? tag(t('ob.dead'), 'error') : fmtDate(f.nextAt)];
    }), 'outbox') : notice('success', t('ob.ok'));
    body.appendChild(panelCard(t('ob.title'), h('div', { class: 'flex flex-col gap-3' }, tiles, failing), [
      button({ label: t('ob.retry'), data: { 'ins-outbox-retry': '' }, onclick: function () { api('POST', '/insights/outbox/retry').then(function () { toast('success', t('saved')); storagePage(root); }, function (err) { toast('error', insErrText(err)); }); } }),
      button({ label: t('ob.discard'), variant: 'danger', disabled: !ob.dead, data: { 'ins-outbox-discard': '' }, onclick: function () { confirmBox(t('ob.discardQ')).then(function (ok) { if (ok) api('POST', '/insights/outbox/discard').then(function () { storagePage(root); }); }); } })], 'outbox'));
  }, function (err) { failInto(body, err); });
}

var INS_CONNECTOR_TEMPLATES = {
  ga4_bigquery: { projectId: '', dataset: 'analytics_000000000', startDate: '', daysPerRun: 3, lagDays: 1, paramTags: [], valueParam: '', nameMap: {}, includeEvents: [] },
  ga4_data: { propertyId: '', startDate: '', daysPerRun: 3, lagDays: 1, dimensions: ['platform'], nameMap: {}, includeEvents: [] },
  gameanalytics: { endpoint: 'https://s3.amazonaws.com', bucket: '', prefix: '', region: 'us-east-1', accessKey: '', filesPerRun: 20, skipSessions: false }
};

function connectorPanel(c, kinds, after) {
  var isNew = !c;
  c = c || { kind: kinds[0], name: '', enabled: false, everyMinutes: 60, config: INS_CONNECTOR_TEMPLATES[kinds[0]], hasSecret: false };
  var kind = select(c.kind, kinds.map(function (k) { return { value: k, label: t('kind.' + k) }; }), { name: 'kind', label: t('im.kind') });
  if (!isNew) kind.disabled = true;
  var name = input(c.name, { name: 'name' });
  var enabled = check(c.enabled, t('cat.enabled'), 'enabled');
  var every = numInput(c.everyMinutes, 'everyMinutes', t('im.every'));
  var config = input(JSON.stringify(c.config, null, 2), { ltr: true, mono: true, multiline: true, rows: 12, name: 'config' });
  var secret = input('', { ltr: true, mono: true, multiline: true, rows: 4, name: 'secret', placeholder: c.hasSecret ? t('st.secretKeep') : '' });
  var reset = check(false, t('im.reset'), 'resetCursor');
  kind.addEventListener('change', function () { config.value = JSON.stringify(INS_CONNECTOR_TEMPLATES[kind.value], null, 2); });
  var msg = h('div', {});
  var save = button({ label: t('save'), variant: 'primary', data: { 'ins-connector-save': '' }, onclick: function () {
    var cfg;
    try { cfg = JSON.parse(config.value); } catch (e) { fieldError(msg, { code: 'INVALID_JSON', field: 'config' }); return; }
    busy(save, true);
    var body = { kind: kind.value, name: name.value, enabled: enabled.input.checked, everyMinutes: numberValue(every) || 60, config: cfg, resetCursor: reset.input.checked };
    if (secret.value.trim()) body.secret = secret.value.trim();
    api(isNew ? 'POST' : 'PUT', '/insights/connectors' + (isNew ? '' : '/' + c.id), body).then(function () { closePanel(); toast('success', t('saved')); after(); }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  var runs = h('div', {});
  var foot = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (!isNew) {
    foot.unshift(button({ label: t('delete'), variant: 'danger', data: { 'ins-connector-delete': '' }, onclick: function () { confirmBox(c.name).then(function (ok) { if (ok) api('DELETE', '/insights/connectors/' + c.id).then(function () { closePanel(); after(); }); }); } }));
    foot.unshift(button({ label: t('im.run'), data: { 'ins-connector-run': '' }, onclick: function () { api('POST', '/insights/connectors/' + c.id + '/run').then(function () { toast('success', t('im.queued')); }, function (err) { fieldError(msg, err); }); } }));
    api('GET', '/insights/connectors/' + c.id + '/runs').then(function (d) {
      runs.appendChild(h('h3', { class: 'font-semibold my-2', text: t('im.runs') }));
      runs.appendChild(dataTable([t('s.started'), t('im.status'), t('im.read'), t('im.written'), t('im.skipped'), t('ob.error')], d.items.map(function (r) {
        return [fmtDate(r.startedAt), tag(r.status, r.status === 'ok' ? 'success' : r.status === 'error' ? 'error' : 'warning'), fmtNum(r.read), fmtNum(r.written), fmtNum(r.skipped), h('span', { class: 'text-xs font-mono', dir: 'ltr', text: r.error })];
      }), 'runs'));
    });
  }
  var hint = kind.value === 'gameanalytics' ? t('im.secretHint.s3') : t('im.secretHint.ga');
  sidePanel(isNew ? t('im.new') : c.name, h('div', {}, notice('warning', t('im.untested')), h('div', { class: 'h-3' }), field(t('im.kind'), kind), field(t('im.name'), name), enabled, field(t('im.every'), every),
    field(t('im.config'), config), field(t('im.secret'), secret, (c.hasSecret ? t('st.secretSet') + ' · ' : '') + hint), isNew ? null : reset, msg, runs), foot, 'connector');
}

function importsPage(root) {
  var frame = dashFrame(root, 'insights/imports');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/connectors').then(function (d) {
    clear(body);
    var reload = function () { importsPage(root); };
    var rows = d.items.map(function (c) {
      return { id: c.id, item: c, cells: [h('span', { dir: 'auto', text: c.name }), t('kind.' + c.kind), tag(c.enabled ? t('cat.enabled') : t('no'), c.enabled ? 'success' : 'secondary'), fmtNum(c.everyMinutes),
        c.lastRunAt ? fmtDate(c.lastRunAt) : t('none'), c.running ? tag(t('im.running'), 'warning') : c.lastStatus ? tag(c.lastStatus, c.lastStatus === 'ok' ? 'success' : 'error') : t('none'), h('span', { class: 'text-xs font-mono', dir: 'ltr', text: c.lastError })] };
    });
    var tb = rows.length ? table([{ label: t('im.name') }, { label: t('im.kind') }, { label: t('cat.enabled') }, { label: t('im.every') }, { label: t('im.last') }, { label: t('im.status') }, { label: t('ob.error') }], rows, function (c) { connectorPanel(c, d.kinds, reload); })
      : h('p', { class: 'text-sm text-secondary-500', text: t('im.empty') });
    if (rows.length) tb.setAttribute('data-ins-table', 'connectors');
    if (!d.secretKey) body.appendChild(notice('warning', t('st.noKey')));
    body.appendChild(panelCard(t('nav.insights/imports'), tb, [button({ label: t('im.new'), variant: 'primary', data: { 'ins-connector-new': '' }, onclick: function () { connectorPanel(null, d.kinds, reload); } })], 'connectors'));
  }, function (err) { failInto(body, err); });
}

var INS_SETTINGS = {
  events: [
    ['set.events', [['enabled', 'bool'], ['gameEvents', 'bool'], ['sessionEvents', 'bool']]],
    ['set.limits', [['maxBatch'], ['maxTags'], ['maxNameLength'], ['maxFields'], ['maxStringLength'], ['valueMin', 'float'], ['valueMax', 'float'], ['pastHours'], ['futureMinutes'], ['receiptDays']]],
    ['set.rate', [['batchesPerMinute'], ['eventsPerMinute']]],
    ['set.policy', [['unknownEvents', ['accept', 'flag', 'reject']], ['unknownTagGroups', ['accept', 'flag', 'reject']], ['unknownTagValues', ['flag', 'reject']], ['valueOutOfRange', ['flag', 'reject']]]],
    ['set.client', [['flushSeconds'], ['maxQueue']]]
  ],
  storage: [['set.storage', [['keepDays'], ['batchSize'], ['maxPendingEvents'], ['deadAfter']]]]
};

function analyticsSettingsPage(root) {
  var frame = dashFrame(root, 'insights/settings');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/events/settings').then(function (d) {
    clear(body);
    var readers = [];
    ['events', 'storage'].forEach(function (block) {
      INS_SETTINGS[block].forEach(function (sec) {
        var grid = h('div', { class: 'grid sm:grid-cols-2 lg:grid-cols-3 gap-x-4' });
        sec[1].forEach(function (f) {
          var key = f[0], kind = f[1], cur = d[block][key], def = d.defaults[block][key];
          var el;
          if (kind === 'bool') { el = check(cur, t('set.' + key), key); readers.push(function (out) { out[block][key] = el.input.checked; }); grid.appendChild(el); return; }
          if (Array.isArray(kind)) {
            el = select(cur, kind.map(function (p) { return { value: p, label: t('p.' + p) }; }), { name: key, label: t('set.' + key) });
            readers.push(function (out) { out[block][key] = el.value; });
            grid.appendChild(field(t('set.' + key), el, t('set.default', { v: t('p.' + def) })));
            return;
          }
          el = input(digits(String(cur)), { ltr: true, name: key });
          readers.push(function (out) { var v = kind === 'float' ? floatValue(el) : numberValue(el); out[block][key] = v == null ? def : v; });
          grid.appendChild(field(t('set.' + key), el, t('set.default', { v: def })));
        });
        body.appendChild(panelCard(t(sec[0]), grid, null, 'settings-' + sec[0]));
      });
    });
    var msg = h('div', { 'data-ins-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', text: d.updatedBy ? t('updated', { by: d.updatedBy, at: fmtDate(d.updatedAt) }) : '' });
    var save = button({ label: t('save'), variant: 'primary', data: { 'ins-settings-save': '' }, onclick: function () {
      var out = { events: Object.assign({}, d.events), storage: Object.assign({}, d.storage) };
      readers.forEach(function (r) { r(out); });
      busy(save, true);
      clear(msg);
      api('PUT', '/insights/events/settings', out).then(function (x) { busy(save, false); d = x; meta.textContent = t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }); toast('success', t('saved')); }, function (err) { busy(save, false); fieldError(msg, err); });
    } });
    body.appendChild(h('div', { class: 'ia-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), save));
  }, function (err) { failInto(body, err); });
}

function contextFillPage(root) {
  var frame = dashFrame(root, 'insights/context');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/events/settings').then(function (d) {
    clear(body);
    var c = d.context || {}, def = (d.defaults || {}).context || {};
    var sources = (d.sources || []).map(function (s) { return { value: s.key, label: s.key + ' · ' + ((insLang() === 'fa' ? s.labelFa || s.labelEn : s.labelEn || s.labelFa) || s.key) }; });
    if (!d.providers) body.appendChild(notice('warning', t('ctx.noProviders')));
    var enabled = check(c.enabled, t('ctx.enabled'), 'enabled');
    var conflict = select(c.conflict, ['override', 'keep_client', 'reject'].map(function (p) { return { value: p, label: t('ctx.c.' + p) }; }), { name: 'conflict', label: t('ctx.conflict') });
    var flag = check(c.flagMismatch, t('ctx.flagMismatch'), 'flagMismatch');
    var outside = check(c.keepClientOutside, t('ctx.keepClientOutside'), 'keepClientOutside');
    var grace = input(digits(String(c.graceSeconds == null ? def.graceSeconds : c.graceSeconds)), { ltr: true, name: 'graceSeconds' });
    body.appendChild(panelCard(t('ctx.policy'), h('div', {}, enabled,
      h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('ctx.conflict'), conflict, t('set.default', { v: t('ctx.c.' + def.conflict) })), field(t('ctx.grace'), grace, t('ctx.graceHint', { v: def.graceSeconds }))),
      flag, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('ctx.flagHint') }), outside, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-1', text: t('ctx.outsideHint') })), null, 'context-policy'));
    var rows = h('div', { class: 'flex flex-col gap-2', 'data-ins-fills': '' });
    var readers = [];
    var addRow = function (f) {
      var opts = sources.slice();
      if (f.source && !opts.some(function (o) { return o.value === f.source; })) opts.push({ value: f.source, label: f.source });
      if (!opts.length) opts.push({ value: '', label: t('none') });
      var on = check(f.enabled, t('ctx.fillOn'), 'fillEnabled');
      var group = input(f.group, { ltr: true, name: 'group', label: t('ctx.group') });
      var src = select(f.source, opts, { name: 'source', label: t('ctx.source') });
      var row = h('div', { class: 'ia-fill flex flex-wrap items-end gap-x-4 border-b border-secondary-100 dark:border-secondary-700 pb-2', 'data-ins-fill': f.group || '' },
        h('div', { class: 'pb-1' }, on), field(t('ctx.group'), group, null, 'flex-1 min-w-[8rem] !mb-1'), field(t('ctx.source'), src, null, 'flex-[2] min-w-[12rem] !mb-1'));
      var reader = function () { return { group: group.value.trim(), source: src.value, enabled: on.input.checked }; };
      readers.push(reader);
      row.appendChild(button({ label: t('delete'), variant: 'danger', data: { 'ins-fill-delete': '' }, onclick: function () { readers.splice(readers.indexOf(reader), 1); row.remove(); } }));
      rows.appendChild(row);
    };
    (c.fills || []).forEach(addRow);
    body.appendChild(panelCard(t('ctx.fills'), h('div', {}, h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 mb-3', text: t('ctx.fillsHint') }), rows,
      h('p', { class: 'text-xs text-secondary-500 mt-3', text: t('ctx.defaults', { v: (def.fills || []).map(function (f) { return f.group + ' ← ' + f.source + (f.enabled ? '' : ' (' + t('ctx.off') + ')'); }).join('، ') }) })),
      [button({ label: t('ctx.addFill'), data: { 'ins-fill-add': '' }, onclick: function () { addRow({ group: '', source: sources.length ? sources[0].value : '', enabled: true }); } })], 'context-fills'));
    var msg = h('div', { 'data-ins-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', text: d.updatedBy ? t('updated', { by: d.updatedBy, at: fmtDate(d.updatedAt) }) : '' });
    var reset = button({ label: t('ctx.reset'), data: { 'ins-context-reset': '' }, onclick: function () {
      clear(rows); readers.length = 0; (def.fills || []).forEach(addRow);
      enabled.input.checked = def.enabled; conflict.value = def.conflict; flag.input.checked = def.flagMismatch; outside.input.checked = def.keepClientOutside; grace.value = digits(String(def.graceSeconds));
    } });
    var save = button({ label: t('save'), variant: 'primary', data: { 'ins-context-save': '' }, onclick: function () {
      var g = numberValue(grace);
      var out = { context: { enabled: enabled.input.checked, conflict: conflict.value, flagMismatch: flag.input.checked, keepClientOutside: outside.input.checked, graceSeconds: g == null ? def.graceSeconds : g,
        fills: readers.map(function (r) { return r(); }).filter(function (f) { return f.group || f.source; }) } };
      busy(save, true);
      clear(msg);
      api('PUT', '/insights/events/settings', out).then(function (x) { busy(save, false); d = x; meta.textContent = t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }); toast('success', t('saved')); }, function (err) { busy(save, false); fieldError(msg, err); });
    } });
    body.appendChild(h('div', { class: 'ia-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), h('div', { class: 'flex gap-2' }, reset, save)));
  }, function (err) { failInto(body, err); });
}

var INS_ROOT = 'ins-root';
var INS_GROUPS = { insights: ['overview', 'active', 'retention', 'funnels', 'explore', 'sessions', 'catalog', 'storage', 'imports', 'settings', 'context'] };
var INS_ORDER = ['insights'];
var INS_RENDER = { 'insights/overview': overviewPage, 'insights/active': activePage, 'insights/retention': retentionPage, 'insights/funnels': funnelsPage, 'insights/explore': explorePage, 'insights/sessions': sessionsPage, 'insights/catalog': catalogPage, 'insights/storage': storagePage, 'insights/imports': importsPage, 'insights/settings': analyticsSettingsPage, 'insights/context': contextFillPage };
var INS_HASH = /^#\/(insights)(?:\/([^/?]*))?/;
var insPending = null;
var insRouter = null;
var insMounted = null;
var insScheduled = false;

function insRoute(hash) {
  var m = INS_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  var pages = INS_GROUPS[m[1]];
  return m[1] + '/' + (pages.indexOf(m[2]) !== -1 ? m[2] : pages[0]);
}

function insAllowed() {
  return !!token();
}

function insTitle() {
  var page = insRoute();
  return page ? t('group.' + page.split('/')[0]) + ' - ' + t('nav.' + page) : '';
}

function insAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || insRouter) return !!insRouter;
  try {
    var layout = r.resolve('/').matched[0];
    INS_ORDER.forEach(function (g) {
      if (r.hasRoute(g)) return;
      r.addRoute({ path: '/' + g + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: g, component: { name: 'InsightsPage', render: function () { return null; } }, meta: { get title() { return insTitle(); } } }] });
    });
    insRouter = r;
  } catch (e) {
    return false;
  }
  if (insPending) {
    var target = insPending;
    insPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function insNavLink(g, page, active) {
  var id = g + '/' + page;
  var on = id === active;
  return h('li', {}, h('a', { href: '#/' + id, 'data-ins-nav': id, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + id) }))));
}

function insSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelectorAll('[data-ins-group]');
  if (!group || !insAllowed()) { Array.prototype.forEach.call(existing, function (e) { e.remove(); }); return; }
  var target = document.querySelector('.x-drawer [data-cup-group]') || document.querySelector('.x-drawer [data-clan-group="trophies"]') || document.querySelector('.x-drawer [data-quiz-group]') || document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = insLang() + '|' + active;
  var mine = document.querySelector('[data-ins-group="insights"]');
  if (mine && existing.length === 1 && mine.getAttribute('data-key') === key && mine.nextElementSibling === target) return;
  Array.prototype.forEach.call(existing, function (e) { e.remove(); });
  INS_ORDER.forEach(function (g) {
    target.before(h('div', { 'data-ins-group': g, 'data-plane-group': g, 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
      h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group.' + g) }),
      h('nav', { 'aria-label': t('group.' + g) }, h('ul', { class: 'space-y-0.5' }, INS_GROUPS[g].map(function (p) { return insNavLink(g, p, active); })))));
  });
}

function insShowMains() {
  var hidden = document.querySelectorAll('[data-ins-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-ins-hidden'); }
}

function insUnmount() {
  var root = document.getElementById(INS_ROOT);
  if (root) root.remove();
  insShowMains();
  closePanel();
  insMounted = null;
}

function insSync() {
  insAddRoutes();
  var page = insRoute();
  insSyncNav(page);
  if (!page) { if (insMounted || document.getElementById(INS_ROOT)) insUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-ins-hidden', ''); }
  }
  var root = document.getElementById(INS_ROOT);
  if (!root) {
    root = h('div', { id: INS_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-ins': 'root' });
    insMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + insLang() + '|' + location.hash;
  if (insMounted !== key) {
    insMounted = key;
    closePanel();
    INS_RENDER[page](root);
  }
}

function insSchedule() {
  if (insScheduled) return;
  insScheduled = true;
  requestAnimationFrame(function () { insScheduled = false; insSync(); });
}

function insBoot() {
  window.addEventListener('hashchange', insSchedule);
  window.addEventListener('popstate', insSchedule);
  window.addEventListener('nakama-console:ready', insSchedule);
  new MutationObserver(insSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { insMounted = null; insSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (insAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  insSync();
}

window.NakamaInsights = { version: 1, pages: function () {
  var out = [];
  INS_ORDER.forEach(function (g) { INS_GROUPS[g].forEach(function (p) { out.push({ id: g + '/' + p, group: g, href: '#/' + g + '/' + p, label: t('nav.' + g + '/' + p) }); }); });
  return out;
} };

if (INS_HASH.test(location.hash)) {
  insPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', insBoot);
else insBoot();
})();
