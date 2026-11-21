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
    'save': 'ذخیره', 'saved': 'ذخیره شد', 'deleted': 'حذف شد', 'cancel': 'انصراف', 'close': 'بستن', 'delete': 'حذف', 'yes': 'بله', 'no': 'خیر', 'add': 'افزودن', 'edit': 'ویرایش',
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'none': '—', 'all': 'همه',
    'confirm.title': 'مطمئن هستید؟', 'lang.text': '{label} ({lang})', 'updated': 'آخرین تغییر: {by} · {at}', 'missing': 'پیدا نشد',
    'f.from': 'از', 'f.to': 'تا', 'f.game': 'بازی', 'f.platform': 'پلتفرم', 'f.version': 'نسخه', 'f.apply': 'اعمال', 'f.days': '{n} روز', 'f.flagged': 'رویدادهای علامت‌خورده هم حساب شوند',
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
    'save': 'Save', 'saved': 'Saved', 'deleted': 'Deleted', 'cancel': 'Cancel', 'close': 'Close', 'delete': 'Delete', 'yes': 'Yes', 'no': 'No', 'add': 'Add', 'edit': 'Edit',
    'loading': 'Loading…', 'empty': 'Nothing found', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'none': '—', 'all': 'All',
    'confirm.title': 'Are you sure?', 'lang.text': '{label} ({lang})', 'updated': 'Last change: {by} · {at}', 'missing': 'Not found',
    'f.from': 'From', 'f.to': 'To', 'f.game': 'Game', 'f.platform': 'Platform', 'f.version': 'Version', 'f.apply': 'Apply', 'f.days': '{n} days', 'f.flagged': 'Count flagged events too',
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
  var d = new Date(v);
  if (isNaN(d)) return '';
  try {
    return d.toLocaleString(insLang() === 'fa' ? 'fa-IR' : 'en-GB', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

function fmtTime(v) {
  if (!v) return '';
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
  var d = new Date(s + 'T12:00:00Z');
  if (isNaN(d)) return digits(s);
  try {
    return d.toLocaleDateString(insLang() === 'fa' ? 'fa-IR' : 'en-GB', { month: 'short', day: 'numeric' });
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
