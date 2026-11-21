(function () {
'use strict';
var CG_TEXT = {
  fa: {
    'group': 'مرحله‌ها',
    'nav.chapters': 'فصل‌ها',
    'nav.stages': 'مرحله‌ها',
    'nav.freeride': 'جایزهٔ گشت آزاد',
    'nav.report': 'گزارش مرحله‌ها',
    'nav.settings': 'تنظیمات مرحله‌ها',
    'desc.chapters': 'فصل‌ها به ترتیب بازی. هر فصل شرط باز شدن، آیکون، آدرس Addressables و جایزه‌های پیشرفت ستاره دارد.',
    'desc.stages': 'مرحله‌های هر فصل با زمان مجاز، کمینهٔ زمان، قانون ستاره، جایزهٔ کامل (سه‌ستاره)، XP و جایزهٔ هر ستاره. تکرار مرحله فقط اختلاف را می‌دهد.',
    'desc.freeride': 'جایزه‌هایی که بازیکن بیرون از مرحله (در گشت آزاد) می‌گیرد، با زمان انتظار و سقف روزانه. گروه جایزهٔ مشترک با مرحله‌های تکرارشونده یک سقف دارند.',
    'desc.report': 'شروع، برد، نرخ تمام کردن، میانگین ستاره و ریزش هر مرحله در روزهای اخیر.',
    'desc.settings': 'قانون‌های پیش‌فرض ستاره و ضریب جایزه، باز شدن مرحله‌ها، زمان تلاش، XP و رویدادهایی که بعد از هر مرحله فرستاده می‌شوند.',
    'save': 'ذخیره', 'saved': 'ذخیره شد', 'cancel': 'انصراف', 'close': 'بستن', 'remove': 'حذف', 'delete': 'حذف', 'search': 'جستجو', 'all': 'همه',
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'yes': 'بله', 'no': 'خیر',
    'inactive': 'غیرفعال', 'active': 'فعال', 'confirm.title': 'مطمئن هستید؟', 'lang.text': '{label} ({lang})', 'updated': 'آخرین تغییر: {by} · {at}',
    'new': 'تازه', 'edit': 'ویرایش', 'up': 'بالا', 'down': 'پایین', 'none': '—', 'default': 'پیش‌فرض تنظیمات', 'own': 'قانون خود این مرحله',
    'f.key': 'کلید', 'f.name': 'نام', 'f.description': 'توضیح', 'f.icon': 'آدرس آیکون', 'f.address': 'آدرس Addressables', 'f.sort': 'ترتیب', 'f.active': 'فعال باشد',
    'f.chapter': 'فصل', 'f.mode': 'نوع', 'f.timeLimit': 'زمان مجاز (ثانیه)', 'f.minSeconds': 'کمینهٔ زمان (ثانیه)', 'f.xp': 'XP کامل', 'f.rewardGroup': 'گروه جایزه',
    'f.cooldown': 'زمان انتظار بین دو جایزه (ثانیه)', 'f.daily': 'سقف جایزه در روز', 'f.xpWhenLimited': 'وقتی جایزه محدود شد XP بدهد', 'f.minLevel': 'کمترین سطح',
    'f.metadata': 'دادهٔ اضافه (JSON)',
    'hint.key': 'همان شناسهٔ مرحله در بازی؛ بعداً عوض نمی‌شود.', 'hint.rewardGroup': 'خالی یعنی سقف خود همین مرحله. با یک نام مشترک، چند مرحله و جایزهٔ گشت آزاد یک سقف دارند.',
    'hint.minSeconds': 'نتیجه‌ای که زودتر از این تمام شود رد می‌شود. خالی یعنی کمترین زمان پیش‌فرض تنظیمات؛ ۰ یعنی بدون کمینه.',
    'mode.ordered': 'داستانی (به ترتیب)', 'mode.repeatable': 'تکرارشونده', 'mode.once': 'یک‌باره (آزمون)',
    'mode.hint.ordered': 'با ستاره‌های مرحله‌های قبل باز می‌شود؛ تکرار فقط اختلاف را می‌دهد.',
    'mode.hint.repeatable': 'همیشه باز؛ هر برد جایزه دارد، با زمان انتظار و سقف روزانه.',
    'mode.hint.once': 'همیشه باز؛ فقط اولین قبولی جایزه دارد.',
    'ch.new': 'فصل تازه', 'ch.stages': '{n} مرحله', 'ch.unlock': 'شرط باز شدن فصل', 'ch.milestones': 'جایزهٔ پیشرفت ستاره', 'ch.addMilestone': 'پله تازه',
    'ch.percent': 'درصد ستاره', 'ch.delete': 'فصل {key} حذف شود؟', 'ch.inUse': 'این فصل هنوز مرحله دارد؛ اول مرحله‌ها را جابه‌جا یا حذف کنید.',
    'ch.noUnlock': 'شرطی ندارد (یا شرط پیش‌فرض تنظیمات)',
    'st.new': 'مرحلهٔ تازه', 'st.filter.chapter': 'همهٔ فصل‌ها', 'st.players': '{n} بازیکن', 'st.delete': 'مرحلهٔ {key} حذف شود؟',
    'st.inUse': 'بازیکنان این مرحله را بازی کرده‌اند؛ به جای حذف خاموشش کنید.', 'st.unlock': 'شرط باز شدن', 'st.unlockDefault': 'پیش‌فرض: مرحله‌های داستانی قبلی این فصل',
    'st.stars': 'قانون ستاره', 'st.multipliers': 'ضریب جایزه برای ۱، ۲ و ۳ ستاره', 'st.rewards': 'جایزهٔ کامل (سه‌ستاره)', 'st.tierPrizes': 'جایزهٔ رسیدن به ستاره',
    'st.addTier': 'جایزهٔ ستاره', 'st.limits': 'زمان انتظار و سقف روزانه', 'st.timing': 'زمان', 'st.preview': 'پیش‌نمایش پرداخت',
    'st.preview.line': '{n} ستاره: {pay}', 'st.preview.replay': 'تکرار با ستارهٔ کمتر یا برابر: هیچ', 'st.moveUp': 'جابه‌جایی به بالا', 'st.moveDown': 'جابه‌جایی به پایین',
    'rule.add': 'شرط تازه', 'rule.type': 'نوع شرط', 'rule.key': 'کلید', 'rule.value': 'مقدار', 'rule.table': 'جدول سطح',
    'rule.level': 'سطح بازیکن', 'rule.meta': 'متا', 'rule.item': 'داشتن آیتم', 'rule.stage': 'ستارهٔ مرحله', 'rule.chapter': 'پیشرفت فصل',
    'rule.s.level': 'سطح ≥ {v}', 'rule.s.meta': '{k} ≥ {v}', 'rule.s.item': 'آیتم {k}', 'rule.s.stage': 'مرحلهٔ {k}: دست‌کم {v} ستاره', 'rule.s.stage0': 'مرحلهٔ {k} تمام شده باشد',
    'rule.s.chapter': 'فصل {k}: دست‌کم {v}٪ ستاره', 'rule.s.chapter0': 'فصل {k} تمام شده باشد',
    'star.add': 'قانون ستاره', 'star.stars': 'ستاره', 'star.cond': 'شرط', 'star.addCond': 'شرط تازه', 'star.field': 'فیلد نتیجه', 'star.op': 'عمل', 'star.value': 'مقدار',
    'star.f.timeSavedPercent': 'درصد زمان صرفه‌جویی‌شده', 'star.f.healthPercent': 'درصد سلامت', 'star.f.elapsedSeconds': 'زمان (ثانیه)', 'star.f.custom': 'فیلد دلخواه بازی…',
    'star.sentence': '{stars} ستاره وقتی {conds}', 'star.and': ' و ', 'star.always': 'همیشه', 'star.useDefault': 'از قانون پیش‌فرض تنظیمات استفاده شود',
    'star.hint': 'بالاترین قانونی که همهٔ شرط‌هایش برقرار باشد ستاره را می‌دهد. برد بدون قانون برقرار، کمترین ستارهٔ تنظیمات را می‌گیرد.',
    'rw.currency': 'ارز', 'rw.amount': 'مقدار', 'rw.addCurrency': 'ارز', 'rw.prize': 'جایزهٔ اقتصاد', 'rw.group': 'گروه جایزه', 'rw.noPrize': 'بدون جایزه', 'rw.missing': 'پیدا نشد',
    'fr.new': 'جایزهٔ تازه', 'fr.delete': 'جایزهٔ {key} حذف شود؟', 'fr.limits': '{c} · {d}', 'fr.cooldown': 'هر {n} ثانیه', 'fr.daily': '{n} بار در روز', 'fr.noLimit': 'بدون سقف',
    'rp.days': 'بازهٔ زمانی', 'rp.d': '{n} روز', 'rp.stage': 'مرحله', 'rp.players': 'بازیکن', 'rp.attempts': 'تلاش', 'rp.wins': 'برد', 'rp.rate': 'نرخ تمام کردن',
    'rp.avgStars': 'میانگین ستاره', 'rp.drop': 'ریزش از مرحلهٔ قبل', 'rp.rejected': 'نتیجهٔ ردشده', 'rp.kpi.players': 'بازیکن‌ها', 'rp.kpi.attempts': 'تلاش‌ها', 'rp.kpi.wins': 'بردها', 'rp.kpi.rejected': 'نتیجه‌های ردشده',
    'se.attempts': 'تلاش و زمان', 'se.hours': 'مهلت ارسال نتیجه (ساعت)', 'se.grace': 'حاشیهٔ زمان (ثانیه)', 'se.minSeconds': 'کمترین زمان مرحله (ثانیه، برای مرحله‌ای که زمان خودش را ندارد)', 'se.closeOthers': 'شروع مرحلهٔ تازه، تلاش‌های باز دیگر را ببندد', 'se.serverClosesOthers': 'شروع مرحله از سمت سرور (مثل بستهٔ آزمون) هم تلاش‌های باز دیگر را ببندد',
    'se.keepDays': 'نگه‌داری تلاش‌های تمام‌شده (روز، ۰ = همیشه)', 'se.stars': 'ستاره و ضریب جایزه', 'se.maxStars': 'بیشترین ستاره', 'se.minOnSuccess': 'ستارهٔ برد بدون قانون برقرار',
    'se.missing': 'شرط روی فیلدی که بازی نفرستاده', 'se.missing.pass': 'برقرار حساب شود', 'se.missing.fail': 'برقرار نباشد', 'se.multipliers': 'ضریب‌های مرحلهٔ داستانی و یک‌باره',
    'se.repeatMultipliers': 'ضریب‌های مرحلهٔ تکرارشونده', 'se.rules': 'قانون‌های پیش‌فرض ستاره', 'se.unlock': 'باز شدن', 'se.stageDefault': 'مرحلهٔ داستانی بی‌شرط باز می‌شود وقتی',
    'se.stageDefault.all_previous': 'همهٔ مرحله‌های داستانی قبل ستاره گرفته باشند', 'se.stageDefault.previous': 'مرحلهٔ داستانی قبلی ستاره گرفته باشد', 'se.stageDefault.none': 'همیشه باز',
    'se.chapterDefault': 'فصل بی‌شرط', 'se.chapterDefault.none': 'همیشه باز', 'se.chapterDefault.previous': 'وقتی فصل قبل تمام شده باشد', 'se.minStars': 'ستارهٔ لازم برای گذشتن از مرحله',
    'se.showLocked': 'فصل و مرحلهٔ قفل هم با شرط‌هایش به بازی فرستاده شود', 'se.rewards': 'پرداخت', 'se.replay': 'تکرار مرحلهٔ داستانی و یک‌باره',
    'se.replay.difference': 'فقط اختلاف با بهترین نتیجه', 'se.replay.none': 'فقط اولین برد', 'se.replay.full': 'هر برد جایزهٔ کامل', 'se.zone': 'منطقهٔ زمانی شروع روز (سقف روزانه)',
    'se.groupMode': 'پرداخت گروه جایزه', 'se.groupMode.draw': 'یک جایزه به قید قرعه', 'se.groupMode.all': 'همهٔ جایزه‌های گروه',
    'se.milestoneModes': 'نوع مرحله‌هایی که در درصد ستارهٔ فصل حساب می‌شوند', 'se.xp': 'XP و سطح', 'se.metaKey': 'متای XP (اقتصاد)', 'se.metaKey.none': 'XP فقط در همین افزونه',
    'se.levelTable': 'جدول سطح', 'se.levelTable.main': 'جدول اصلی', 'se.metaMissing': 'متای {k} در اقتصاد نیست؛ XP فقط در همین افزونه جمع می‌شود.',
    'se.events': 'رویدادها', 'se.economyEvent': 'رویداد اقتصاد بعد از هر مرحله', 'se.analyticsEvent': 'رویداد تحلیل (کلید مرحله = context)', 'se.onFailure': 'برای باخت هم فرستاده شود',
    'se.play': 'هزینهٔ بازی', 'se.play.hint': 'هر شروع مرحله این مقدار از ارز را کم می‌کند (مثلاً یک جان). مرحله با metadata.playCost می‌تواند مقدار خودش را داشته باشد؛ ۰ یعنی رایگان.',
    'se.costCurrency': 'ارز هزینه', 'se.cost.none': 'بدون هزینه', 'se.costAmount': 'مقدار هر شروع', 'se.refundOnWin': 'اگر برد، هزینه برگردد',
    'se.costMissing': 'ارز {k} در اقتصاد نیست؛ شروع مرحله رد می‌شود.', 'se.costNoRegen': 'ارز {k} پرشدن خودکار ندارد؛ هزینه مثل یک کسر ساده از کیف پول است.',
    'code.NOT_ENOUGH_PLAY_COST': 'ارز کافی برای شروع مرحله نیست.',
    'se.freeride': 'گشت آزاد', 'se.requireRequestId': 'requestId برای گرفتن جایزه لازم باشد', 'se.requestHours': 'نگه‌داری جواب هر requestId (ساعت)',
    'se.noEconomy': 'افزونهٔ اقتصاد بارگذاری نشده؛ سطح و جایزه‌های اقتصاد کار نمی‌کنند.',
    'pc.title': 'پیشرفت مرحله‌ها', 'pc.xp': 'XP مرحله‌ها: {xp} · سطح {level}', 'pc.reset': 'از نو', 'pc.reset.confirm': 'پیشرفت مرحلهٔ «{name}» پاک شود؟',
    'pc.reset.claims': 'جایزه‌های گرفته‌شده هم پاک شود (بردهای بعدی دوباره جایزه می‌دهند)', 'pc.notPlayed': 'بازی نشده', 'pc.locked': 'قفل',
    'pc.wins': '{w} برد از {c}', 'pc.attempts': 'تلاش‌های اخیر', 'pc.noCampaign': 'افزونهٔ مرحله‌ها پاسخ نداد.',
    'att.open': 'باز', 'att.closed': 'بسته', 'att.paying': 'در حال پرداخت', 'att.done': 'تمام',
    'code.INVALID_KEY': 'کلید درست نیست؛ با حرف یا عدد انگلیسی شروع شود و فقط حرف، عدد و _ . : - داشته باشد.',
    'code.ALREADY_EXISTS': 'این کلید از قبل هست.', 'code.IN_USE': 'در حال استفاده است.', 'code.CHAPTER_NOT_FOUND': 'فصل پیدا نشد.', 'code.STAGE_NOT_FOUND': 'مرحله پیدا نشد.',
    'code.VALUE_OUT_OF_RANGE': 'مقدار بیرون از بازهٔ مجاز است.', 'code.INVALID_MULTIPLIERS': 'برای هر ستاره یک ضریب بین ۰ و ۱۰ لازم است.', 'code.INVALID_RULE': 'نوع شرط درست نیست.',
    'code.INVALID_OPERATOR': 'عمل شرط درست نیست.', 'code.INVALID_FIELD': 'نام فیلد درست نیست (حرف انگلیسی، عدد و _).', 'code.REWARD_REQUIRED': 'یک جایزه انتخاب کنید.',
    'code.DUPLICATE_MILESTONE': 'هر درصد یک بار.', 'code.INVALID_TIMEZONE': 'منطقهٔ زمانی شناخته نشد.', 'code.INVALID_POLICY': 'یکی از گزینه‌ها درست نیست.',
    'code.DEFAULT_LANGUAGE_REQUIRED': 'نام به زبان پیش‌فرض لازم است.', 'code.INVALID_MODE': 'نوع مرحله درست نیست.', 'code.INVALID_JSON': 'JSON درست نیست.',
    'code.INVALID_PRIZE_KEY': 'کلید جایزه درست نیست.',
    'fd.default': 'پیش‌فرض: {v}', 'fd.defaultOpt': 'پیش‌فرض ({v})', 'fd.tagsHint': 'چند مقدار را با ویرگول جدا کنید.', 'fd.item': '{label} {n}', 'fd.add': '{label}',
    'fd.none': 'برای این فصل فیلدی تعریف نشده است. فیلدها را در «تنظیمات مرحله‌ها ← فیلدهای مرحله و فصل» تعریف کنید.',
    'tpl.none': 'قانون پیش‌فرض تنظیمات', 'tpl.label': 'قالب قانون ستاره', 'tpl.hint': 'قالب‌ها در تنظیمات مرحله‌ها تعریف می‌شوند. قانون پایین همان چیزی است که سرور حساب می‌کند.',
    'st.noReplay': 'بعد از برد دوباره بازی نشود', 'st.noReplayHint': 'بازیکنی که یک بار برده باشد این مرحله را دوباره شروع نمی‌کند (STAGE_ALREADY_WON).',
    'st.settings': 'تنظیمات مأموریت', 'st.settingsHint': 'بازی این مقدارها را در campaign_chapters و در فیلد settings هر مرحله می‌خواند. خالی یعنی پیش‌فرض فیلد.',
    'st.advanced': 'پیشرفته', 'hint.metadata': 'دادهٔ آزاد برای بازی. فیلدهایی که فرم دارند را این‌جا ننویسید.',
    'ch.stepId': 'شناسهٔ پله: {id}', 'ch.milestoneHint': 'هر پله شناسهٔ ثابت دارد. اگر درصد را عوض کنید، بازیکنی که این پله را گرفته دوباره جایزه نمی‌گیرد.',
    'ch.settings': 'تنظیمات فصل', 'fr.off': 'این مسیر خاموش است و بازی آن را نمی‌بیند. گشت آزاد را با مرحله‌های تکرارشونده بسازید یا در تنظیمات روشنش کنید.',
    'se.tableEnabled': 'جدول جدای جایزهٔ گشت آزاد روشن باشد', 'se.tableEnabledHint': 'خاموش یعنی گشت آزاد فقط از مسیر مرحله‌های تکرارشونده می‌گذرد و campaign_freeride_claim پاسخ FREERIDE_NOT_FOUND می‌دهد.',
    'se.templates': 'قالب‌های قانون ستاره', 'se.templatesHint': 'هر مرحله می‌تواند یکی از این قالب‌ها را بگیرد. پایه: زمان و سلامت و دفعهٔ دیده شدن. مسابقه: بدون سلامت. مخفی: بدون زمان. کوئست: مثل پایه.',
    'se.addTemplate': 'قالب تازه', 'se.fields': 'فیلدهای مرحله و فصل', 'se.fieldsJson': 'تعریف فیلدها (JSON)',
    'se.fieldsHint': 'از روی این تعریف، فرم «تنظیمات مأموریت» در هر مرحله و «تنظیمات فصل» ساخته می‌شود. نوع‌ها: number، integer، bool، text، i18n، select، tags، list. با chapters فیلد فقط در همان فصل‌ها دیده می‌شود و با kinds فقط در فصل‌هایی که metadata.kind آن‌ها در فهرست است (مثلاً همهٔ فصل‌های POI).',
    'star.missing': 'اگر بازی نفرستاد', 'star.missing.default': 'طبق تنظیمات', 'star.missing.pass': 'قبول', 'star.missing.fail': 'رد',
    'star.f.timesSeen': 'دفعهٔ دیده شدن', 'star.ifMissing': ' (اگر نیامد: قبول)',
    'code.INVALID_FIELD_VALUE': 'مقدار این فیلد درست نیست.', 'code.UNKNOWN_FIELD': 'این فیلد تعریف نشده است.', 'code.INVALID_TEMPLATE': 'این قالب ستاره پیدا نشد.',
    'code.TOO_MANY_ITEMS': 'تعداد ردیف‌ها از سقف بیشتر است.', 'code.INVALID_OPTION': 'این گزینه در فهرست نیست.', 'code.FIELDS_TOO_DEEP': 'فهرست تودرتو حداکثر سه سطح دارد.',
    'code.INVALID_FIELD_TYPE': 'نوع فیلد درست نیست.', 'code.OPTIONS_REQUIRED': 'فیلد انتخابی گزینه لازم دارد.', 'code.DUPLICATE_KEY': 'این کلید تکراری است.',
    'code.TOO_MANY_TEMPLATES': 'تعداد قالب‌ها زیاد است.', 'code.TEXT_TOO_LONG': 'متن خیلی بلند است.', 'code.FIELDS_REQUIRED': 'فهرست باید فیلد داشته باشد.',
    'code.STAGE_ALREADY_WON': 'این مرحله برده شده و دوباره بازی نمی‌شود.'
  },
  en: {
    'group': 'Stages',
    'nav.chapters': 'Chapters',
    'nav.stages': 'Stages',
    'nav.freeride': 'Free ride rewards',
    'nav.report': 'Stage report',
    'nav.settings': 'Stage settings',
    'desc.chapters': 'Chapters in play order. Each has unlock rules, an icon, an Addressables address and star progress rewards.',
    'desc.stages': 'Stages of each chapter with a time limit, a minimum time, star rules, the full (3-star) reward, XP and per-star prizes. A replay pays only the difference.',
    'desc.freeride': 'Rewards a player claims outside stages (free ride), with a cooldown and a daily cap. A reward group shared with repeatable stages shares one cap.',
    'desc.report': 'Starts, wins, completion rate, average stars and drop-off of each stage over the last days.',
    'desc.settings': 'Default star rules and reward multipliers, how stages unlock, attempt timing, XP and the events sent after each stage.',
    'save': 'Save', 'saved': 'Saved', 'cancel': 'Cancel', 'close': 'Close', 'remove': 'Remove', 'delete': 'Delete', 'search': 'Search', 'all': 'All',
    'loading': 'Loading…', 'empty': 'Nothing found', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'yes': 'Yes', 'no': 'No',
    'inactive': 'Inactive', 'active': 'Active', 'confirm.title': 'Are you sure?', 'lang.text': '{label} ({lang})', 'updated': 'Last change: {by} · {at}',
    'new': 'New', 'edit': 'Edit', 'up': 'Up', 'down': 'Down', 'none': '—', 'default': 'Settings default', 'own': 'This stage\'s own rules',
    'f.key': 'Key', 'f.name': 'Name', 'f.description': 'Description', 'f.icon': 'Icon address', 'f.address': 'Addressables address', 'f.sort': 'Order', 'f.active': 'Active',
    'f.chapter': 'Chapter', 'f.mode': 'Kind', 'f.timeLimit': 'Time limit (seconds)', 'f.minSeconds': 'Minimum time (seconds)', 'f.xp': 'Full XP', 'f.rewardGroup': 'Reward group',
    'f.cooldown': 'Cooldown between rewards (seconds)', 'f.daily': 'Rewards per day', 'f.xpWhenLimited': 'Give XP when the reward is limited', 'f.minLevel': 'Lowest level',
    'f.metadata': 'Extra data (JSON)',
    'hint.key': 'The stage id in the game; it cannot change later.', 'hint.rewardGroup': 'Empty means this stage\'s own cap. With a shared name, stages and free ride rewards share one cap.',
    'hint.minSeconds': 'A result finished faster than this is refused. Empty uses the default minimum from the settings; 0 means no minimum.',
    'mode.ordered': 'Story (in order)', 'mode.repeatable': 'Repeatable', 'mode.once': 'One time (quiz)',
    'mode.hint.ordered': 'Unlocks with the stars of earlier stages; a replay pays only the difference.',
    'mode.hint.repeatable': 'Always open; every win pays, with a cooldown and a daily cap.',
    'mode.hint.once': 'Always open; only the first pass pays.',
    'ch.new': 'New chapter', 'ch.stages': '{n} stages', 'ch.unlock': 'Chapter unlock rules', 'ch.milestones': 'Star progress rewards', 'ch.addMilestone': 'Add step',
    'ch.percent': 'Star percent', 'ch.delete': 'Delete chapter {key}?', 'ch.inUse': 'This chapter still has stages; move or delete them first.',
    'ch.noUnlock': 'No rules (or the settings default)',
    'st.new': 'New stage', 'st.filter.chapter': 'All chapters', 'st.players': '{n} players', 'st.delete': 'Delete stage {key}?',
    'st.inUse': 'Players have played this stage; turn it off instead.', 'st.unlock': 'Unlock rules', 'st.unlockDefault': 'Default: the earlier story stages of this chapter',
    'st.stars': 'Star rules', 'st.multipliers': 'Reward multiplier for 1, 2 and 3 stars', 'st.rewards': 'Full reward (3 stars)', 'st.tierPrizes': 'Prizes for reaching stars',
    'st.addTier': 'Star prize', 'st.limits': 'Cooldown and daily cap', 'st.timing': 'Timing', 'st.preview': 'Payout preview',
    'st.preview.line': '{n} stars: {pay}', 'st.preview.replay': 'Replay with the same or fewer stars: nothing', 'st.moveUp': 'Move up', 'st.moveDown': 'Move down',
    'rule.add': 'Add rule', 'rule.type': 'Rule', 'rule.key': 'Key', 'rule.value': 'Value', 'rule.table': 'Level table',
    'rule.level': 'Player level', 'rule.meta': 'Meta', 'rule.item': 'Owns item', 'rule.stage': 'Stage stars', 'rule.chapter': 'Chapter progress',
    'rule.s.level': 'level ≥ {v}', 'rule.s.meta': '{k} ≥ {v}', 'rule.s.item': 'item {k}', 'rule.s.stage': 'stage {k}: at least {v} stars', 'rule.s.stage0': 'stage {k} passed',
    'rule.s.chapter': 'chapter {k}: at least {v}% stars', 'rule.s.chapter0': 'chapter {k} finished',
    'star.add': 'Add star rule', 'star.stars': 'Stars', 'star.cond': 'Condition', 'star.addCond': 'Add condition', 'star.field': 'Result field', 'star.op': 'Op', 'star.value': 'Value',
    'star.f.timeSavedPercent': 'Time saved %', 'star.f.healthPercent': 'Health %', 'star.f.elapsedSeconds': 'Time (seconds)', 'star.f.custom': 'Custom game field…',
    'star.sentence': '{stars} stars when {conds}', 'star.and': ' and ', 'star.always': 'always', 'star.useDefault': 'Use the default rules of the settings',
    'star.hint': 'The highest rule whose conditions all hold gives the stars. A win that matches no rule gets the settings\' lowest stars.',
    'rw.currency': 'Currency', 'rw.amount': 'Amount', 'rw.addCurrency': 'Currency', 'rw.prize': 'Economy prize', 'rw.group': 'Prize group', 'rw.noPrize': 'No prize', 'rw.missing': 'not found',
    'fr.new': 'New reward', 'fr.delete': 'Delete reward {key}?', 'fr.limits': '{c} · {d}', 'fr.cooldown': 'every {n} s', 'fr.daily': '{n} a day', 'fr.noLimit': 'no cap',
    'rp.days': 'Period', 'rp.d': '{n} days', 'rp.stage': 'Stage', 'rp.players': 'Players', 'rp.attempts': 'Attempts', 'rp.wins': 'Wins', 'rp.rate': 'Completion rate',
    'rp.avgStars': 'Average stars', 'rp.drop': 'Drop-off from the previous stage', 'rp.rejected': 'Refused results', 'rp.kpi.players': 'Players', 'rp.kpi.attempts': 'Attempts', 'rp.kpi.wins': 'Wins', 'rp.kpi.rejected': 'Refused results',
    'se.attempts': 'Attempts and timing', 'se.hours': 'Time to send a result (hours)', 'se.grace': 'Time grace (seconds)', 'se.minSeconds': 'Shortest stage time (seconds, for stages without their own)', 'se.closeOthers': 'Starting a stage closes the player\'s other open attempts', 'se.serverClosesOthers': 'A stage started by the server (such as a quiz pack) also closes the other open attempts',
    'se.keepDays': 'Keep finished attempts (days, 0 = forever)', 'se.stars': 'Stars and reward multipliers', 'se.maxStars': 'Most stars', 'se.minOnSuccess': 'Stars of a win that matches no rule',
    'se.missing': 'A condition on a field the game did not send', 'se.missing.pass': 'holds', 'se.missing.fail': 'does not hold', 'se.multipliers': 'Multipliers of story and one-time stages',
    'se.repeatMultipliers': 'Multipliers of repeatable stages', 'se.rules': 'Default star rules', 'se.unlock': 'Unlocking', 'se.stageDefault': 'A story stage without rules opens when',
    'se.stageDefault.all_previous': 'every earlier story stage has stars', 'se.stageDefault.previous': 'the story stage before it has stars', 'se.stageDefault.none': 'always open',
    'se.chapterDefault': 'A chapter without rules', 'se.chapterDefault.none': 'is always open', 'se.chapterDefault.previous': 'opens when the chapter before is finished', 'se.minStars': 'Stars needed to pass a stage',
    'se.showLocked': 'Send locked chapters and stages to the game too, with their rules', 'se.rewards': 'Payouts', 'se.replay': 'Replaying story and one-time stages',
    'se.replay.difference': 'pays only the difference to the best result', 'se.replay.none': 'pays only the first win', 'se.replay.full': 'pays the full reward every win', 'se.zone': 'Zone where a day starts (daily caps)',
    'se.groupMode': 'Prize group payout', 'se.groupMode.draw': 'One prize drawn by weight', 'se.groupMode.all': 'Every prize of the group',
    'se.milestoneModes': 'Stage kinds counted in a chapter\'s star percent', 'se.xp': 'XP and level', 'se.metaKey': 'XP meta (economy)', 'se.metaKey.none': 'XP only in this plugin',
    'se.levelTable': 'Level table', 'se.levelTable.main': 'Main table', 'se.metaMissing': 'Economy has no meta {k}; XP is kept only in this plugin.',
    'se.events': 'Events', 'se.economyEvent': 'Economy event after each stage', 'se.analyticsEvent': 'Analytics event (stage key = context)', 'se.onFailure': 'Send them for lost runs too',
    'se.play': 'Play cost', 'se.play.hint': 'Every stage start takes this much of the currency (a life, for example). A stage can set its own amount with metadata.playCost; 0 means free.',
    'se.costCurrency': 'Cost currency', 'se.cost.none': 'No cost', 'se.costAmount': 'Amount per start', 'se.refundOnWin': 'Give it back on a win',
    'se.costMissing': 'Currency {k} is not in economy; stage starts will be refused.', 'se.costNoRegen': 'Currency {k} does not refill over time; the cost is a plain wallet debit.',
    'code.NOT_ENOUGH_PLAY_COST': 'Not enough currency to start the stage.',
    'se.freeride': 'Free ride', 'se.requireRequestId': 'A requestId is required to claim', 'se.requestHours': 'Keep the answer of each requestId (hours)',
    'se.noEconomy': 'The economy plugin is not loaded; levels and economy prizes do not work.',
    'pc.title': 'Stage progress', 'pc.xp': 'Stage XP: {xp} · level {level}', 'pc.reset': 'Reset', 'pc.reset.confirm': 'Clear the progress of “{name}”?',
    'pc.reset.claims': 'Also forget the rewards already paid (the next wins pay again)', 'pc.notPlayed': 'not played', 'pc.locked': 'locked',
    'pc.wins': '{w} wins of {c}', 'pc.attempts': 'Recent attempts', 'pc.noCampaign': 'The stages plugin did not answer.',
    'att.open': 'open', 'att.closed': 'closed', 'att.paying': 'paying', 'att.done': 'done',
    'code.INVALID_KEY': 'Invalid key; start with an English letter or digit and use letters, digits and _ . : - only.',
    'code.ALREADY_EXISTS': 'This key already exists.', 'code.IN_USE': 'In use.', 'code.CHAPTER_NOT_FOUND': 'Chapter not found.', 'code.STAGE_NOT_FOUND': 'Stage not found.',
    'code.VALUE_OUT_OF_RANGE': 'The value is out of range.', 'code.INVALID_MULTIPLIERS': 'One multiplier between 0 and 10 per star is needed.', 'code.INVALID_RULE': 'Invalid rule type.',
    'code.INVALID_OPERATOR': 'Invalid operator.', 'code.INVALID_FIELD': 'Invalid field name (English letters, digits and _).', 'code.REWARD_REQUIRED': 'Pick a reward.',
    'code.DUPLICATE_MILESTONE': 'Each percent once.', 'code.INVALID_TIMEZONE': 'Unknown time zone.', 'code.INVALID_POLICY': 'One of the choices is not valid.',
    'code.DEFAULT_LANGUAGE_REQUIRED': 'A name in the default language is required.', 'code.INVALID_MODE': 'Invalid stage kind.', 'code.INVALID_JSON': 'Invalid JSON.',
    'code.INVALID_PRIZE_KEY': 'Invalid prize key.',
    'fd.default': 'Default: {v}', 'fd.defaultOpt': 'Default ({v})', 'fd.tagsHint': 'Separate values with commas.', 'fd.item': '{label} {n}', 'fd.add': '{label}',
    'fd.none': 'No fields are defined for this chapter. Define them in Stage settings > Stage and chapter fields.',
    'tpl.none': 'Default rules of the settings', 'tpl.label': 'Star rule template', 'tpl.hint': 'Templates are defined in the stage settings. The rules below are what the server scores.',
    'st.noReplay': 'No replay after a win', 'st.noReplayHint': 'A player who has won once cannot start this stage again (STAGE_ALREADY_WON).',
    'st.settings': 'Mission settings', 'st.settingsHint': 'The game reads these values from campaign_chapters, in each stage\'s settings field. Empty means the field default.',
    'st.advanced': 'Advanced', 'hint.metadata': 'Free data for the game. Do not put fields that have a form here.',
    'ch.stepId': 'Step id: {id}', 'ch.milestoneHint': 'Each step has a fixed id. Changing its percent never pays a player who already got it.',
    'ch.settings': 'Chapter settings', 'fr.off': 'This path is off and the game does not see it. Build free ride with repeatable stages, or turn it on in the settings.',
    'se.tableEnabled': 'Separate free ride reward table on', 'se.tableEnabledHint': 'Off means free ride only goes through repeatable stages and campaign_freeride_claim answers FREERIDE_NOT_FOUND.',
    'se.templates': 'Star rule templates', 'se.templatesHint': 'Each stage can pick one template. Base: time, health and times seen. Race: no health. Stealth: no time. Quest: like Base.',
    'se.addTemplate': 'New template', 'se.fields': 'Stage and chapter fields', 'se.fieldsJson': 'Field definitions (JSON)',
    'se.fieldsHint': 'The Mission settings form of each stage and the Chapter settings form are built from this. Types: number, integer, bool, text, i18n, select, tags, list. With chapters a field shows only in those chapters, with kinds only in chapters whose metadata.kind is listed (for example every POI chapter).',
    'star.missing': 'If the game does not send it', 'star.missing.default': 'As in the settings', 'star.missing.pass': 'Pass', 'star.missing.fail': 'Fail',
    'star.f.timesSeen': 'Times seen', 'star.ifMissing': ' (missing: pass)',
    'code.INVALID_FIELD_VALUE': 'This field value is not valid.', 'code.UNKNOWN_FIELD': 'This field is not defined.', 'code.INVALID_TEMPLATE': 'This star template was not found.',
    'code.TOO_MANY_ITEMS': 'Too many rows.', 'code.INVALID_OPTION': 'This option is not in the list.', 'code.FIELDS_TOO_DEEP': 'Nested lists go three levels deep at most.',
    'code.INVALID_FIELD_TYPE': 'Invalid field type.', 'code.OPTIONS_REQUIRED': 'A select field needs options.', 'code.DUPLICATE_KEY': 'This key is used twice.',
    'code.TOO_MANY_TEMPLATES': 'Too many templates.', 'code.TEXT_TOO_LONG': 'The text is too long.', 'code.FIELDS_REQUIRED': 'A list needs fields.',
    'code.STAGE_ALREADY_WON': 'This stage is won and cannot be played again.'
  }
};

function cgLang() {
  var nc = window.NakamaConsole;
  var l = nc && nc.lang ? nc.lang() : (document.documentElement.getAttribute('lang') || 'fa');
  return /^en/i.test(l) ? 'en' : 'fa';
}

function has(key) {
  return (CG_TEXT[cgLang()] || CG_TEXT.fa)[key] != null;
}

function t(key, vars) {
  var dict = CG_TEXT[cgLang()] || CG_TEXT.fa;
  var s = dict[key] != null ? dict[key] : (CG_TEXT.en[key] != null ? CG_TEXT.en[key] : key);
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] != null ? digits(String(vars[k])) : m; });
  return s;
}

function digits(s) {
  if (cgLang() !== 'fa') return s;
  return String(s).replace(/[0-9]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'[+d]; }).replace(/,/g, '٬');
}

function fmtNum(n) {
  if (n == null || n === '') return '';
  var s = Number(n).toLocaleString('en-US', { maximumFractionDigits: 2 });
  return digits(s);
}

function fmtDate(v) {
  if (!v) return '';
  var d = new Date(v);
  if (isNaN(d)) return '';
  try {
    return d.toLocaleString(cgLang() === 'fa' ? 'fa-IR' : 'en-GB', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

function tr(key, vars, raw) {
  var s = t(key, vars);
  Object.keys(raw || {}).forEach(function (k) { s = s.split('{' + k + '}').join(String(raw[k])); });
  return s;
}

var CG_API = '/v2/console/base';
var CG_INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 px-3 py-2 text-sm';
var CG_BTN = 'cg-btn inline-flex items-center justify-center gap-1.5 font-medium whitespace-nowrap rounded-md border text-sm py-1.5 px-3';

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

function CgError(status, code, message, field) {
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
  return fetch((path.indexOf('/v2/') === 0 ? '' : CG_API) + path, init).then(function (res) {
    if (res.status === 401 && !/^#\/login/.test(location.hash)) location.hash = '#/login';
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) throw new CgError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', data && data.field || '');
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
  var b = h('button', { type: opts.type || 'button', class: CG_BTN + ' cg-btn--' + (opts.variant || 'ghost') + (opts.class ? ' ' + opts.class : ''), disabled: !!opts.disabled, title: opts.title, 'aria-label': opts.title, onclick: opts.onclick },
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

var cgToastBox = null;
function toast(kind, text) {
  if (!cgToastBox || !cgToastBox.isConnected) {
    cgToastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-campaign': 'toasts' });
    document.body.appendChild(cgToastBox);
  }
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  cgToastBox.appendChild(el);
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
  var err = h('p', { class: 'cg-field-error text-xs text-error-600 dark:text-error-400', hidden: true });
  var wrap = h('label', { class: 'cg-field block mb-4' + (extra ? ' ' + extra : '') },
    label ? h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1', text: label }) : null,
    control,
    hint ? h('span', { class: 'block text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: hint }) : null,
    err);
  wrap.setError = function (msg) { err.textContent = msg || ''; err.hidden = !msg; };
  return wrap;
}

function input(value, opts) {
  opts = opts || {};
  var el = h(opts.multiline ? 'textarea' : 'input', { class: CG_INPUT + (opts.mono ? ' font-mono' : ''), type: opts.multiline ? null : (opts.type || 'text'), dir: opts.ltr ? 'ltr' : 'auto', rows: opts.multiline ? (opts.rows || 3) : null, placeholder: opts.placeholder, name: opts.name, autocomplete: 'off', 'aria-label': opts.label });
  el.value = value == null ? '' : value;
  return el;
}

function select(value, options, opts) {
  opts = opts || {};
  var el = h('select', { class: CG_INPUT + ' eco-select', name: opts.name, 'aria-label': opts.label });
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) opt.selected = true;
    el.appendChild(opt);
  });
  return el;
}

function check(checked, label, name) {
  var box = h('input', { type: 'checkbox', class: 'cg-check', name: name, checked: !!checked });
  var wrap = h('label', { class: 'cg-checkline flex items-start gap-2 mb-3 cursor-pointer text-sm text-secondary-800 dark:text-secondary-200' }, box, h('span', { text: label }));
  wrap.input = box;
  return wrap;
}

function numberValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? Math.round(n) : NaN;
}

var cgPanel = null;
function closePanel() {
  if (cgPanel) { cgPanel.remove(); cgPanel = null; document.removeEventListener('keydown', panelKey); }
}

function panelKey(e) { if (e.key === 'Escape') closePanel(); }

function sidePanel(title, body, footer, name) {
  closePanel();
  var holder = h('div', { class: 'cg-panel-holder' });
  var panel = h('div', { class: 'cg-panel bg-white dark:bg-secondary-900 shadow-lg', role: 'dialog', 'aria-modal': 'true', 'aria-label': title, 'data-campaign-panel': name || '' },
    h('div', { class: 'cg-panel-head flex items-center justify-between border-b px-6 py-3' },
      h('h2', { class: 'text-lg font-semibold truncate', text: title }),
      button({ label: '×', title: t('close'), onclick: closePanel, class: 'cg-close' })),
    h('div', { class: 'cg-panel-body p-6' }, body),
    footer ? h('div', { class: 'cg-panel-foot flex gap-2 justify-end bg-secondary-50 dark:bg-secondary-800 p-4' }, footer) : null);
  holder.appendChild(panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) closePanel(); });
  cgPanel = h('div', { 'data-campaign': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, h('div', { class: 'cg-shade' }), holder);
  document.body.appendChild(cgPanel);
  document.addEventListener('keydown', panelKey);
  return panel;
}

function confirmBox(text, okLabel) {
  return new Promise(function (resolve) {
    var root;
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-campaign': 'confirm', 'data-l10n-skip': '', class: 'cg-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'cg-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('confirm.title') }),
        h('p', { class: 'text-sm mb-4', text: text }),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: okLabel || t('yes'), variant: 'danger', data: { 'campaign-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root);
  });
}

function pageFrame(page, desc, actions) {
  var content = h('div', { class: 'flex flex-col gap-4' });
  var el = h('div', { class: 'cg-page flex flex-col flex-1 px-4 md:px-6 pb-6 max-w-[1600px] w-full mx-auto', 'data-campaign-page': page },
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
    var tr = h('tr', { class: 'border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' + (onrow ? ' cursor-pointer' : ''), 'data-campaign-row': row.id || '' });
    if (onrow) {
      tr.tabIndex = 0;
      tr.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,select,label')) onrow(row.item); });
      tr.addEventListener('keydown', function (e) { if (e.key === 'Enter') onrow(row.item); });
    }
    row.cells.forEach(function (cell, i) { tr.appendChild(h('td', { class: 'px-3 py-2 align-middle text-sm', 'data-label': columns[i].label }, cell)); });
    tbody.appendChild(tr);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow rounded-md overflow-x-auto' }, h('table', { class: 'w-full eco-table cg-table' }, thead, tbody));
}

function pager(state, onpage) {
  var total = state.total || 0;
  var from = total ? (state.page - 1) * state.limit + 1 : 0;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  return h('div', { class: 'flex items-center justify-end gap-2 text-xs text-secondary-500' },
    h('span', { 'data-campaign-range': '', text: total ? t('range', { from: from, to: to, total: total }) : t('empty') }),
    button({ label: t('prev'), disabled: state.page <= 1, onclick: function () { onpage(state.page - 1); } }),
    button({ label: t('next'), disabled: state.page >= last, onclick: function () { onpage(state.page + 1); } }));
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

var cgLanguages = null;
function loadLanguages() {
  if (cgLanguages) return Promise.resolve(cgLanguages);
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch('/v2/console/base/languages', { headers: headers, credentials: 'same-origin' }).then(function (r) { return r.ok ? r.json() : null; }).then(function (d) {
    var list = d && d.items ? d.items.filter(function (l) { return l.isActive; }).map(function (l) { return { code: l.code, name: l.nativeName || l.code, isDefault: !!l.isDefault, direction: l.direction }; }) : [];
    if (!list.length) list = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    list.sort(function (a, b) { return (b.isDefault ? 1 : 0) - (a.isDefault ? 1 : 0); });
    cgLanguages = list;
    return list;
  }, function () {
    cgLanguages = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    return cgLanguages;
  });
}

function translated(label, name, value, opts) {
  opts = opts || {};
  var inputs = {};
  var box = h('div', { class: 'grid sm:grid-cols-2 gap-x-4', 'data-campaign-text': name });
  (cgLanguages || []).forEach(function (l) {
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
  var l = cgLang();
  return m[l] || m.fa || m.en || Object.keys(m).map(function (k) { return m[k]; }).filter(Boolean)[0] || fallback || '';
}

function fieldError(msgBox, err) {
  clear(msgBox).appendChild(notice('error', errorText(err) + (err && err.field ? ' · ' + err.field : '')));
}

function options(list, value) {
  return list.map(function (o) { return typeof o === 'string' ? { value: o, label: o } : o; });
}

function dateInput(value, name, type) {
  var el = h('input', { type: type || 'date', class: CG_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  el.value = value || '';
  return el;
}

function isoDay(d) {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

function cgErrText(err) {
  if (err && err.code && has('code.' + err.code)) return t('code.' + err.code) + (err.field ? ' · ' + err.field : '');
  return errorText(err) + (err && err.field ? ' · ' + err.field : '');
}

function section(title, children, name) {
  return card(h('div', { class: 'p-4', 'data-campaign-section': name || '' },
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
    var msg = h('div', { 'data-campaign-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', 'data-campaign-updated': '' });
    var paintMeta = function (x) { meta.textContent = x.updatedBy ? t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }) : ''; };
    paintMeta(d);
    var save = button({ label: t('save'), variant: 'primary', data: { 'campaign-save': page }, onclick: function () {
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
        msg.appendChild(notice('error', cgErrText(err)));
      });
    } });
    clear(frame.content);
    form.el.forEach(function (el) { frame.content.appendChild(el); });
    frame.content.appendChild(h('div', { class: 'cg-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), save));
  }, function (err) {
    clear(frame.content).appendChild(notice('error', cgErrText(err)));
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
  return h('a', { href: '#/players/' + id, class: 'eco-link', 'data-campaign-player-link': id, dir: 'auto', text: name || id.slice(0, 8) });
}

var cgRefs = null;
function loadRefs(force) {
  if (cgRefs && !force) return Promise.resolve(cgRefs);
  return api('GET', '/campaign/refs' + qs({ lang: cgLang() })).then(function (r) { cgRefs = r || {}; return cgRefs; }, function () { cgRefs = {}; return cgRefs; });
}

function refOptions(kind, value, emptyLabel) {
  var list = (cgRefs && cgRefs[kind]) || [];
  var opts = emptyLabel != null ? [{ value: '', label: emptyLabel }] : [];
  list.forEach(function (r) { opts.push({ value: r.key, label: (r.name && r.name !== r.key ? r.name + ' · ' : '') + r.key }); });
  if (value && !list.some(function (r) { return r.key === value; })) opts.push({ value: value, label: value + ' (' + t('rw.missing') + ')' });
  return opts;
}

function refSelect(kind, value, name, emptyLabel) {
  var el = select(value || '', refOptions(kind, value, emptyLabel), { name: name, label: name });
  el.setAttribute('dir', 'auto');
  return el;
}

function currencyRows(values, name) {
  var box = h('div', { class: 'cg-rows', 'data-campaign-currencies': name || '' });
  var rows = [];
  function addRow(k, v) {
    var key = refSelect('currencies', k || '', 'reward-currency', '—');
    var amt = numInput(v == null ? '' : v, 'reward-amount', t('rw.amount'));
    var row = h('div', { class: 'cg-row' }, field(t('rw.currency'), key), field(t('rw.amount'), amt),
      h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r.row !== row; }); box.dispatchEvent(new Event('input', { bubbles: true })); } })));
    rows.push({ row: row, key: key, amt: amt });
    box.appendChild(row);
  }
  Object.keys(values || {}).forEach(function (k) { addRow(k, values[k]); });
  var wrap = h('div', {}, box, button({ label: '+ ' + t('rw.addCurrency'), data: { 'campaign-action': 'add-currency' }, onclick: function () { addRow(); } }));
  wrap.read = function () {
    var out = {};
    rows.forEach(function (r) { var k = r.key.value; var n = numberValue(r.amt); if (k && n) out[k] = n; });
    return out;
  };
  return wrap;
}

function prizePickers(value) {
  value = value || {};
  var prize = refSelect('prizes', value.prizeKey, 'reward-prize', t('rw.noPrize'));
  var group = refSelect('prizeGroups', value.prizeGroupKey, 'reward-prize-group', t('rw.noPrize'));
  var el = h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('rw.prize'), prize), field(t('rw.group'), group));
  el.read = function () { var o = {}; if (prize.value) o.prizeKey = prize.value; if (group.value) o.prizeGroupKey = group.value; return o; };
  return el;
}

var CG_RULE_TYPES = ['level', 'meta', 'item', 'stage', 'chapter'];

function ruleSentence(r) {
  var v = r.value || 0;
  switch (r.type) {
    case 'level': return t('rule.s.level', { v: v }) + (r.table ? ' (' + r.table + ')' : '');
    case 'meta': return tr('rule.s.meta', { v: v }, { k: r.key });
    case 'item': return tr('rule.s.item', null, { k: r.key });
    case 'stage': return v ? tr('rule.s.stage', { v: v }, { k: r.key }) : tr('rule.s.stage0', null, { k: r.key });
    case 'chapter': return v ? tr('rule.s.chapter', { v: v }, { k: r.key }) : tr('rule.s.chapter0', null, { k: r.key });
  }
  return r.type;
}

function ruleEditor(rules, name) {
  var box = h('div', { class: 'cg-rows', 'data-campaign-rules': name || '' });
  var rows = [];
  function keyKind(type) { return { meta: 'metas', item: 'items', stage: 'stages', chapter: 'chapters' }[type]; }
  function addRow(r) {
    r = r || { type: 'level', value: 1 };
    var type = select(r.type, CG_RULE_TYPES.map(function (x) { return { value: x, label: t('rule.' + x) }; }), { name: 'rule-type', label: t('rule.type') });
    var keyBox = h('div', {});
    var value = numInput(r.value == null ? '' : r.value, 'rule-value', t('rule.value'));
    var sentence = h('div', { class: 'text-xs text-secondary-500 cg-sentence', 'data-campaign-rule-sentence': '' });
    var key = null;
    var row = h('div', { class: 'cg-row' });
    function paintKey(cur) {
      clear(keyBox);
      var kind = type.value === 'level' ? 'levelTables' : keyKind(type.value);
      key = refSelect(kind, cur, type.value === 'level' ? 'rule-table' : 'rule-key', type.value === 'level' ? t('se.levelTable.main') : '—');
      keyBox.appendChild(field(type.value === 'level' ? t('rule.table') : t('rule.key'), key));
      value.disabled = type.value === 'item';
      key.addEventListener('change', paint);
    }
    function read() {
      var o = { type: type.value, value: type.value === 'item' ? 0 : (floatValue(value) || 0) };
      if (type.value === 'level') { if (key.value) o.table = key.value; } else o.key = key.value;
      return o;
    }
    function paint() { sentence.textContent = ruleSentence(read()); }
    type.addEventListener('change', function () { paintKey(''); paint(); });
    value.addEventListener('input', paint);
    paintKey(r.type === 'level' ? r.table : r.key);
    row.appendChild(field(t('rule.type'), type));
    row.appendChild(keyBox);
    row.appendChild(field(t('rule.value'), value));
    row.appendChild(h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (x) { return x.row !== row; }); } })));
    row.appendChild(h('div', { class: 'cg-row-full' }, sentence));
    rows.push({ row: row, read: read });
    box.appendChild(row);
    paint();
  }
  (rules || []).forEach(addRow);
  var wrap = h('div', {}, box, button({ label: '+ ' + t('rule.add'), data: { 'campaign-action': 'add-rule' }, onclick: function () { addRow(); } }));
  wrap.read = function () { return rows.map(function (r) { return r.read(); }); };
  return wrap;
}

var CG_FIELDS = ['timeSavedPercent', 'healthPercent', 'timesSeen', 'elapsedSeconds'];
var CG_OPS = ['>=', '>', '<=', '<', '==', '!='];

function starSentence(r) {
  var conds = (r.conditions || []).map(function (c) {
    var label = has('star.f.' + c.field) ? t('star.f.' + c.field) : c.field;
    return label + ' ' + c.op + ' ' + fmtNum(c.value) + (c.missing === 'pass' ? t('star.ifMissing') : '');
  });
  return tr('star.sentence', { stars: r.stars }, { conds: conds.length ? conds.join(t('star.and')) : t('star.always') });
}

function starEditor(rules, maxStars) {
  var box = h('div', { class: 'flex flex-col gap-3', 'data-campaign-star-rules': '' });
  var items = [];
  function addRule(r) {
    r = r || { stars: maxStars || 3, conditions: [{ field: 'timeSavedPercent', op: '>=', value: 0 }] };
    var stars = numInput(r.stars, 'star-stars', t('star.stars'));
    var sentence = h('p', { class: 'text-sm font-medium cg-sentence', 'data-campaign-star-sentence': '' });
    var condBox = h('div', { class: 'cg-rows' });
    var conds = [];
    var card_ = h('div', { class: 'cg-track', 'data-campaign-star-rule': '' });
    function addCond(c) {
      c = c || { field: 'healthPercent', op: '>=', value: 0 };
      var custom = CG_FIELDS.indexOf(c.field) === -1;
      var fieldSel = select(custom ? '__custom' : c.field, CG_FIELDS.map(function (f) { return { value: f, label: t('star.f.' + f) }; }).concat([{ value: '__custom', label: t('star.f.custom') }]), { name: 'cond-field', label: t('star.field') });
      var customIn = input(custom ? c.field : '', { name: 'cond-custom', ltr: true, mono: true, placeholder: 'arrests' });
      customIn.hidden = !custom;
      var op = select(c.op, CG_OPS.map(function (o) { return { value: o, label: o }; }), { name: 'cond-op', label: t('star.op') });
      var val = numInput(c.value, 'cond-value', t('star.value'));
      var miss = select(c.missing || '', [{ value: '', label: t('star.missing.default') }, { value: 'pass', label: t('star.missing.pass') }, { value: 'fail', label: t('star.missing.fail') }], { name: 'cond-missing', label: t('star.missing') });
      var row = h('div', { class: 'cg-row' }, field(t('star.field'), h('div', { class: 'flex flex-col gap-1' }, fieldSel, customIn)), field(t('star.op'), op), field(t('star.value'), val), field(t('star.missing'), miss),
        h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); conds = conds.filter(function (x) { return x.row !== row; }); paint(); } })));
      fieldSel.addEventListener('change', function () { customIn.hidden = fieldSel.value !== '__custom'; paint(); });
      [customIn, op, val, miss].forEach(function (el) { el.addEventListener('input', paint); el.addEventListener('change', paint); });
      conds.push({ row: row, read: function () { var o = { field: fieldSel.value === '__custom' ? customIn.value.trim() : fieldSel.value, op: op.value, value: floatValue(val) || 0 }; if (miss.value) o.missing = miss.value; return o; } });
      condBox.appendChild(row);
    }
    function read() { return { stars: numberValue(stars) || 0, conditions: conds.map(function (c) { return c.read(); }) }; }
    function paint() { sentence.textContent = starSentence(read()); }
    stars.addEventListener('input', paint);
    (r.conditions || []).forEach(addCond);
    card_.appendChild(h('div', { class: 'flex flex-wrap items-end gap-3' }, field(t('star.stars'), stars, null, 'cg-num'), h('div', { class: 'flex-1' }, sentence),
      button({ label: t('remove'), variant: 'danger', onclick: function () { card_.remove(); items = items.filter(function (x) { return x.el !== card_; }); } })));
    card_.appendChild(condBox);
    card_.appendChild(button({ label: '+ ' + t('star.addCond'), data: { 'campaign-action': 'add-cond' }, onclick: function () { addCond(); paint(); } }));
    items.push({ el: card_, read: read });
    box.appendChild(card_);
    paint();
  }
  (rules || []).forEach(addRule);
  var wrap = h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('star.hint') }), box,
    button({ label: '+ ' + t('star.add'), data: { 'campaign-action': 'add-star-rule' }, onclick: function () { addRule(); } }));
  wrap.read = function () { return items.map(function (i) { return i.read(); }); };
  return wrap;
}

function multiplierInputs(values, maxStars, name) {
  var inputs = [];
  var box = h('div', { class: 'flex flex-wrap gap-3', 'data-campaign-multipliers': name || '' });
  for (var i = 0; i < (maxStars || 3); i++) {
    var el = input(values && values[i] != null ? digits(String(values[i])) : '', { ltr: true, name: (name || 'mult') + '-' + (i + 1), label: String(i + 1) });
    el.setAttribute('inputmode', 'decimal');
    el.classList.add('cg-num');
    inputs.push(el);
    box.appendChild(field('★'.repeat(i + 1), el, null, 'cg-num'));
  }
  box.read = function () { return inputs.map(function (el) { var v = floatValue(el); return v == null || isNaN(v) ? 0 : v; }); };
  box.inputs = inputs;
  return box;
}

function rewardText(cur) {
  var keys = Object.keys(cur || {});
  if (!keys.length) return t('none');
  return '\u2066' + keys.map(function (k) { return fmtNum(cur[k]) + ' ' + k; }).join(' + ') + '\u2069';
}

function rewardListText(list) {
  if (!list || !list.length) return t('none');
  return '\u2066' + list.map(function (r) { return fmtNum(r.amount) + ' ' + r.currencyKey; }).join(' + ') + '\u2069';
}

function modeTag(mode) {
  return tag(t('mode.' + mode), mode === 'ordered' ? 'primary' : mode === 'once' ? 'warning' : 'success');
}

function stars(n, max) {
  var s = '';
  for (var i = 0; i < (max || 3); i++) s += i < n ? '★' : '☆';
  return h('span', { class: 'cg-stars', 'data-campaign-stars': String(n), 'aria-label': String(n), text: s });
}

function cgAppliesTo(d, scope) {
  var keys = d.chapters || [], kinds = d.kinds || [];
  if (!keys.length && !kinds.length) return true;
  return keys.indexOf(scope.key) !== -1 || (!!scope.kind && kinds.indexOf(scope.kind) !== -1);
}

function cgScope(c) {
  return { key: (c && c.key) || '', kind: (c && c.metadata && c.metadata.kind) || '' };
}

function cgDefaultText(d) {
  if (d.default == null) return '';
  if (d.type === 'bool') return d.default ? t('yes') : t('no');
  if (d.type === 'tags') return (d.default || []).join(', ');
  if (typeof d.default === 'number') return fmtNum(d.default);
  return String(d.default);
}

function fieldsForm(defs, values, scope, prefix) {
  values = values || {};
  prefix = prefix || 'settings';
  var box = h('div', { class: 'cg-fields', 'data-campaign-fields': prefix });
  var readers = [];
  (defs || []).forEach(function (d) {
    if (scope && !cgAppliesTo(d, scope)) return;
    var label = textOf(d.label, d.key);
    var hint = textOf(d.hint, '');
    var name = prefix + '.' + d.key;
    var has_ = Object.prototype.hasOwnProperty.call(values, d.key) && values[d.key] != null;
    var v = values[d.key];
    var def = cgDefaultText(d);
    var withDef = def ? (hint ? hint + ' · ' : '') + t('fd.default', { v: def }) : hint;
    var el, read;
    switch (d.type) {
      case 'number':
      case 'integer':
        el = input(has_ ? digits(String(v)) : '', { ltr: true, name: name, label: label, placeholder: def });
        el.setAttribute('inputmode', d.type === 'integer' ? 'numeric' : 'decimal');
        el.classList.add('cg-num');
        read = function () { var n = d.type === 'integer' ? numberValue(el) : floatValue(el); return n == null ? undefined : n; };
        box.appendChild(field(label, el, withDef, 'cg-num-field'));
        break;
      case 'bool':
        var start = has_ ? !!v : !!d.default;
        el = check(start, label, name);
        read = function () { var c = el.input.checked; return has_ || c !== !!d.default ? c : undefined; };
        box.appendChild(h('div', {}, el, hint ? h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: hint }) : null));
        break;
      case 'select':
        el = select(has_ ? v : '', [{ value: '', label: def ? t('fd.defaultOpt', { v: def }) : t('none') }].concat((d.options || []).map(function (o) { return { value: o, label: has('opt.' + o) ? t('opt.' + o) + ' · ' + o : o }; })), { name: name, label: label });
        read = function () { return el.value || undefined; };
        box.appendChild(field(label, el, hint));
        break;
      case 'tags':
        el = input(has_ ? (v || []).join(', ') : '', { ltr: true, mono: true, name: name, label: label, placeholder: def });
        read = function () {
          var raw = el.value.trim();
          if (!raw) return has_ ? [] : undefined;
          return raw.split(/[,،]/).map(function (x) { return x.trim(); }).filter(Boolean);
        };
        box.appendChild(field(label, el, withDef ? withDef + ' · ' + t('fd.tagsHint') : t('fd.tagsHint')));
        break;
      case 'i18n':
        el = translated(label, name, has_ ? v : {}, { multiline: true });
        read = function () { var o = el.read(); return Object.keys(o).length ? o : undefined; };
        box.appendChild(h('div', {}, el, hint ? h('p', { class: 'text-xs text-secondary-500 -mt-3 mb-3', text: hint }) : null));
        break;
      case 'list':
        var items = [];
        var list = h('div', { class: 'cg-rows', 'data-campaign-list-field': d.key });
        var addItem = function (item) {
          var sub = fieldsForm(d.fields, item || {}, null, d.key);
          var row = h('div', { class: 'cg-track', 'data-campaign-list-item': d.key });
          var entry = { row: row, sub: sub };
          row.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-2' },
            h('span', { class: 'text-sm font-medium', 'data-campaign-list-title': '' }),
            button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); items = items.filter(function (x) { return x !== entry; }); renumber(); } })));
          row.appendChild(sub);
          items.push(entry);
          list.appendChild(row);
          renumber();
        };
        var renumber = function () { items.forEach(function (x, i) { x.row.querySelector('[data-campaign-list-title]').textContent = t('fd.item', { label: label, n: i + 1 }); }); };
        (has_ ? v : []).forEach(addItem);
        var max = d.maxItems || 20;
        read = function () { var out = items.map(function (x) { return x.sub.read(); }); return has_ || out.length ? out : undefined; };
        box.appendChild(h('div', { class: 'mb-4', 'data-campaign-list': d.key },
          h('div', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1', text: label }),
          hint ? h('p', { class: 'text-xs text-secondary-500 mb-2', text: hint }) : null, list,
          button({ label: '+ ' + t('fd.add', { label: label }), data: { 'campaign-action': 'add-' + d.key }, onclick: function () { if (items.length < max) addItem({}); else toast('error', t('code.TOO_MANY_ITEMS')); } })));
        break;
      default:
        el = input(has_ ? v : '', { name: name, label: label, placeholder: def, ltr: /^[A-Za-z0-9_.:\-\s]*$/.test(String(has_ ? v : def)) });
        read = function () { var s = el.value.trim(); return s ? s : (has_ ? '' : undefined); };
        box.appendChild(field(label, el, withDef));
    }
    readers.push({ key: d.key, read: read });
  });
  box.read = function () {
    var out = {};
    readers.forEach(function (r) { var v = r.read(); if (v !== undefined && !(typeof v === 'number' && isNaN(v))) out[r.key] = v; });
    return out;
  };
  box.invalid = function () {
    var bad = null;
    readers.forEach(function (r) { var v = r.read(); if (typeof v === 'number' && isNaN(v)) bad = bad || r.key; });
    return bad;
  };
  box.empty = !readers.length;
  return box;
}

function chaptersPage(root) {
  var frame = pageFrame('chapters', t('desc.chapters'), [button({ label: t('ch.new'), variant: 'primary', data: { 'campaign-action': 'new-chapter' }, onclick: function () { chapterEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  function reload() {
    clear(frame.content).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/campaign/chapters'), loadLanguages(), loadRefs(true)]).then(function (res) {
      var list = res[0].items || [];
      clear(frame.content);
      if (!list.length) { frame.content.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      var move = function (i, d) {
        var ids = list.map(function (c) { return c.id; });
        var j = i + d;
        if (j < 0 || j >= ids.length) return;
        var x = ids[i]; ids[i] = ids[j]; ids[j] = x;
        api('PUT', '/campaign/chapters/order', { ids: ids }).then(reload, function (err) { toast('error', cgErrText(err)); });
      };
      frame.content.appendChild(h('div', { 'data-campaign-list': 'chapters' }, table([
        { label: t('f.sort') }, { label: t('f.name') }, { label: t('f.key') }, { label: t('nav.stages') }, { label: t('ch.unlock') }, { label: t('ch.milestones') }, { label: '' }
      ], list.map(function (c, i) {
        return { id: c.key, item: c, cells: [
          h('div', { class: 'flex gap-1' }, button({ label: '↑', title: t('st.moveUp'), disabled: i === 0, data: { 'campaign-up': c.key }, onclick: function () { move(i, -1); } }),
            button({ label: '↓', title: t('st.moveDown'), disabled: i === list.length - 1, data: { 'campaign-down': c.key }, onclick: function () { move(i, 1); } })),
          h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(c.name, c.key) }), c.isActive ? null : tag(t('inactive'), 'warning')),
          mono(c.key),
          h('a', { href: '#/campaign/stages?chapterId=' + c.id, class: 'eco-link', text: t('ch.stages', { n: c.stageCount }) }),
          h('span', { class: 'text-xs', text: (c.unlock || []).length ? c.unlock.map(ruleSentence).join(' · ') : t('ch.noUnlock') }),
          h('span', { class: 'text-xs', text: (c.milestones || []).map(function (m) { return digits(String(m.percent)) + '٪'; }).join(' · ') || t('none') }),
          button({ label: t('delete'), variant: 'danger', data: { 'campaign-delete': c.key }, onclick: function () {
            confirmBox(t('ch.delete', { key: c.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/campaign/chapters/' + c.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', err.code === 'IN_USE' ? t('ch.inUse') : cgErrText(err)); });
            });
          } })
        ] };
      }), function (c) { chapterEditor(c, reload); })));
    }, function (err) { clear(frame.content).appendChild(notice('error', cgErrText(err))); });
  }
  reload();
}

function milestoneRows(list) {
  var box = h('div', { class: 'flex flex-col gap-3', 'data-campaign-milestones': '' });
  var rows = [];
  function addRow(m) {
    m = m || { percent: 100 };
    var pct = numInput(m.percent, 'milestone-percent', t('ch.percent'));
    var cur = currencyRows(m.currencies, 'milestone');
    var prizes = prizePickers(m);
    var el = h('div', { class: 'cg-track', 'data-campaign-milestone': '' }, h('div', { class: 'flex items-end gap-3' }, field(t('ch.percent'), pct, null, 'cg-num'), h('div', { class: 'flex-1' }),
      button({ label: t('remove'), variant: 'danger', onclick: function () { el.remove(); rows = rows.filter(function (r) { return r.el !== el; }); } })), cur, prizes);
    if (m.id) el.querySelector('.flex').insertBefore(h('span', { class: 'text-xs text-secondary-500 self-center', 'data-campaign-milestone-id': m.id, text: t('ch.stepId', { id: m.id }) }), el.querySelector('.flex').children[1]);
    rows.push({ el: el, read: function () { var o = prizes.read(); if (m.id) o.id = m.id; o.percent = floatValue(pct) || 0; o.currencies = cur.read(); return o; } });
    box.appendChild(el);
  }
  (list || []).forEach(addRow);
  var wrap = h('div', {}, box, button({ label: '+ ' + t('ch.addMilestone'), data: { 'campaign-action': 'add-milestone' }, onclick: function () { addRow(); } }));
  wrap.read = function () { return rows.map(function (r) { return r.read(); }); };
  return wrap;
}

function chapterEditor(c, done) {
  loadCgSettings().then(function (cfgRes) { chapterForm(c, done, cfgRes); }, function () { chapterForm(c, done, null); });
}

function chapterForm(c, done, cfgRes) {
  var creating = !c;
  c = c || { name: {}, description: {}, unlock: [], milestones: [], isActive: true, metadata: {} };
  var key = input(c.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = translated(t('f.name'), 'name', c.name);
  var desc = translated(t('f.description'), 'description', c.description, { multiline: true });
  var icon = input(c.iconUrl, { name: 'iconUrl', ltr: true });
  var address = input(c.address, { name: 'address', ltr: true, mono: true });
  var active = check(c.isActive, t('f.active'), 'isActive');
  var rules = ruleEditor(c.unlock, 'chapter');
  var milestones = milestoneRows(c.milestones);
  var defs = (cfgRes && cfgRes.settings.fields && cfgRes.settings.fields.chapter) || [];
  var setForm = fieldsForm(defs, c.settings || {}, cgScope(c));
  var meta = input(JSON.stringify(c.metadata || {}, null, 2), { name: 'metadata', ltr: true, mono: true, multiline: true, rows: 4 });
  var msg = h('div', {});
  var body = h('div', { 'data-campaign-form': 'chapter' }, field(t('f.key'), key), name, desc, h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.icon'), icon), field(t('f.address'), address)), active,
    section(t('ch.unlock'), rules, 'unlock'), h('div', { class: 'h-3' }), section(t('ch.milestones'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('ch.milestoneHint') }), milestones), 'milestones'),
    setForm.empty ? null : h('div', { class: 'h-3' }), setForm.empty ? null : section(t('ch.settings'), setForm, 'settings'),
    h('div', { class: 'h-3' }), section(t('st.advanced'), field(t('f.metadata'), meta, t('hint.metadata')), 'advanced'), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'campaign-action': 'save-chapter' }, onclick: function () {
    busy(save, true);
    clear(msg);
    var md;
    try { md = JSON.parse(meta.value || '{}'); } catch (e) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_JSON') + ' · metadata')); return; }
    var badField = setForm.invalid();
    if (badField) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_FIELD_VALUE') + ' · ' + badField)); return; }
    var out = { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), address: address.value.trim(), sort: c.sort || 0,
      isActive: active.input.checked, unlock: rules.read(), milestones: milestones.read(), metadata: md, settings: setForm.read() };
    if (creating) out.sort = 1000;
    api(creating ? 'POST' : 'PUT', '/campaign/chapters' + (creating ? '' : '/' + c.id), out).then(function () {
      closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); msg.appendChild(notice('error', cgErrText(err))); });
  } });
  sidePanel(creating ? t('ch.new') : textOf(c.name, c.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'chapter');
}

var cgSettingsCache = null;
function loadCgSettings() {
  if (cgSettingsCache) return Promise.resolve(cgSettingsCache);
  return api('GET', '/campaign/settings').then(function (d) { cgSettingsCache = d; return d; });
}

function stagesPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 50, total: 0, chapterId: q.chapterId || '', q: q.q || '', mode: '' };
  var chapters = [];
  var frame = pageFrame('stages', t('desc.stages'), [button({ label: t('st.new'), variant: 'primary', data: { 'campaign-action': 'new-stage' }, onclick: function () { stageEditor(null, chapters, reload, state.chapterId); } })]);
  clear(root).appendChild(frame.el);
  var chapterSel = h('select', { class: CG_INPUT + ' eco-select', name: 'chapterId', 'aria-label': t('f.chapter') });
  var modeSel = select('', [{ value: '', label: t('all') }].concat(['ordered', 'repeatable', 'once'].map(function (m) { return { value: m, label: t('mode.' + m) }; })), { name: 'mode', label: t('f.mode') });
  var search = input(state.q, { name: 'q', placeholder: t('search') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end cg-filters' }, field(t('f.chapter'), chapterSel), field(t('f.mode'), modeSel), field(t('search'), search)));
  frame.content.appendChild(listBox);
  chapterSel.addEventListener('change', function () { state.chapterId = chapterSel.value; state.page = 1; reload(); });
  modeSel.addEventListener('change', function () { state.mode = modeSel.value; state.page = 1; reload(); });
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.q = search.value.trim(); state.page = 1; reload(); }, 300); });

  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/campaign/chapters'), api('GET', '/campaign/stages' + qs({ chapterId: state.chapterId, mode: state.mode, q: state.q, page: state.page, limit: state.limit })), loadLanguages(), loadRefs(true), loadCgSettings()]).then(function (res) {
      chapters = res[0].items || [];
      clear(chapterSel);
      [{ value: '', label: t('st.filter.chapter') }].concat(chapters.map(function (c) { return { value: c.id, label: textOf(c.name, c.key) }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.chapterId) opt.selected = true;
        chapterSel.appendChild(opt);
      });
      var d = res[1];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      var max = (cgSettingsCache && cgSettingsCache.settings.stars.maxStars) || 3;
      var move = function (i, dir) {
        var same = items.filter(function (x) { return x.chapterId === items[i].chapterId; });
        var ids = same.map(function (x) { return x.id; });
        var at = ids.indexOf(items[i].id);
        var j = at + dir;
        if (j < 0 || j >= ids.length) return;
        var x = ids[at]; ids[at] = ids[j]; ids[j] = x;
        api('PUT', '/campaign/stages/order', { ids: ids }).then(reload, function (err) { toast('error', cgErrText(err)); });
      };
      listBox.appendChild(h('div', { 'data-campaign-list': 'stages' }, table([
        { label: t('f.sort') }, { label: t('f.name') }, { label: t('f.chapter') }, { label: t('f.mode') }, { label: t('st.rewards') }, { label: t('f.timeLimit') }, { label: t('rp.players') }, { label: '' }
      ], items.map(function (s, i) {
        return { id: s.key, item: s, cells: [
          state.chapterId ? h('div', { class: 'flex gap-1' }, button({ label: '↑', title: t('st.moveUp'), data: { 'campaign-up': s.key }, onclick: function () { move(i, -1); } }),
            button({ label: '↓', title: t('st.moveDown'), data: { 'campaign-down': s.key }, onclick: function () { move(i, 1); } })) : h('span', { text: digits(String(s.sort)) }),
          h('div', { class: 'flex flex-col' }, h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(s.name, s.key) }), s.isActive ? null : tag(t('inactive'), 'warning')), mono(s.key)),
          h('span', { text: s.chapterKey }),
          modeTag(s.mode),
          h('span', { class: 'text-xs', 'data-campaign-reward': s.key, text: rewardText((s.rewards || {}).currencies) + (s.xp ? ' · \u2066' + fmtNum(s.xp) + ' XP\u2069' : '') }),
          h('span', { text: s.timeLimitSeconds ? fmtNum(s.timeLimitSeconds) : t('none') }),
          h('span', { text: fmtNum(s.players) }),
          button({ label: t('delete'), variant: 'danger', data: { 'campaign-delete': s.key }, onclick: function () {
            confirmBox(t('st.delete', { key: s.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/campaign/stages/' + s.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', err.code === 'IN_USE' ? t('st.inUse') : cgErrText(err)); });
            });
          } })
        ] };
      }), function (s) { stageEditor(s, chapters, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
      void max;
    }, function (err) { clear(listBox).appendChild(notice('error', cgErrText(err))); });
  }
  reload();
}

function previewLines(cur, xp, mult, max) {
  var lines = [];
  for (var n = 1; n <= max; n++) {
    var m = mult[n - 1] || 0;
    var parts = Object.keys(cur || {}).map(function (k) { return fmtNum(Math.floor(cur[k] * m + 1e-9)) + ' ' + k; });
    if (xp) parts.push(fmtNum(Math.floor(xp * m + 1e-9)) + ' XP');
    lines.push(tr('st.preview.line', { n: n }, { pay: parts.length ? '\u2066' + parts.join(' + ') + '\u2069' : t('none') }));
  }
  return lines;
}

function stageEditor(s, chapters, done, chapterId) {
  var creating = !s;
  loadCgSettings().then(function (cfgRes) {
    var cfg = cfgRes.settings;
    var max = cfg.stars.maxStars;
    s = s || { name: {}, description: {}, mode: 'ordered', isActive: true, rewards: {}, tierPrizes: [], chapterId: chapterId || (chapters[0] && chapters[0].id) || '', metadata: {}, ownUnlock: false, ownStarRules: false, ownMultipliers: false };
    var key = input(s.key || '', { name: 'key', ltr: true, mono: true });
    if (!creating) key.disabled = true;
    var chapter = select(s.chapterId, chapters.map(function (c) { return { value: c.id, label: textOf(c.name, c.key) }; }), { name: 'chapterId', label: t('f.chapter') });
    var mode = select(s.mode, ['ordered', 'repeatable', 'once'].map(function (m) { return { value: m, label: t('mode.' + m) }; }), { name: 'mode', label: t('f.mode') });
    var modeHint = h('p', { class: 'text-xs text-secondary-500 -mt-3 mb-3', 'data-campaign-mode-hint': '' });
    var name = translated(t('f.name'), 'name', s.name);
    var desc = translated(t('f.description'), 'description', s.description, { multiline: true });
    var icon = input(s.iconUrl, { name: 'iconUrl', ltr: true });
    var address = input(s.address, { name: 'address', ltr: true, mono: true });
    var active = check(s.isActive, t('f.active'), 'isActive');
    var limit = numInput(s.timeLimitSeconds, 'timeLimitSeconds', t('f.timeLimit'));
    var minSec = numInput(s.minSeconds, 'minSeconds', t('f.minSeconds'));
    var ownUnlock = check(s.ownUnlock, t('own'), 'ownUnlock');
    var rules = ruleEditor(s.ownUnlock ? s.unlock : [], 'stage');
    var unlockBox = h('div', {}, ownUnlock, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('st.unlockDefault') }), rules);
    var ownStars = check(!s.ownStarRules, t('star.useDefault'), 'useDefaultStars');
    var templates = cfg.stars.templates || [];
    var template = select(s.starTemplate || '', [{ value: '', label: t('tpl.none') }].concat(templates.map(function (x) { return { value: x.key, label: textOf(x.name, x.key) + ' · ' + x.key }; })), { name: 'starTemplate', label: t('tpl.label') });
    var templateBox = h('div', {}, field(t('tpl.label'), template, t('tpl.hint')), h('ul', { class: 'text-sm flex flex-col gap-1 mb-3', 'data-campaign-template-rules': '' }));
    var noReplay = check(s.noReplayAfterWin, t('st.noReplay'), 'noReplayAfterWin');
    var scopeOf = function (id) { return cgScope(chapters.filter(function (x) { return x.id === id; })[0]); };
    var settingsHolder = h('div', { 'data-campaign-stage-settings': '' });
    var settingsForm = null;
    var stageDefs = (cfg.fields && cfg.fields.stage) || [];
    function paintSettings() {
      var keep = settingsForm ? settingsForm.read() : (s.settings || {});
      settingsForm = fieldsForm(stageDefs, keep, scopeOf(chapter.value));
      clear(settingsHolder).appendChild(settingsForm.empty ? h('p', { class: 'text-sm text-secondary-500', text: t('fd.none') }) : settingsForm);
    }
    chapter.addEventListener('change', paintSettings);
    paintSettings();
    var starRules = starEditor(s.ownStarRules ? s.starRules : cfg.stars.rules, max);
    var ownMult = check(!s.ownMultipliers, t('star.useDefault'), 'useDefaultMultipliers');
    var mult = multiplierInputs(s.ownMultipliers ? s.starMultipliers : (s.mode === 'repeatable' ? cfg.stars.repeatableMultipliers : cfg.stars.multipliers), max, 'mult');
    var cur = currencyRows((s.rewards || {}).currencies, 'stage');
    var xp = numInput(s.xp || '', 'xp', t('f.xp'));
    var tiers = h('div', { class: 'cg-rows', 'data-campaign-tiers': '' });
    var tierRows = [];
    function addTier(tp) {
      tp = tp || { stars: max };
      var st = numInput(tp.stars, 'tier-stars', t('star.stars'));
      var pickers = prizePickers(tp);
      var row = h('div', { class: 'cg-track', 'data-campaign-tier': '' }, h('div', { class: 'flex items-end gap-3' }, field(t('star.stars'), st, null, 'cg-num'), h('div', { class: 'flex-1' }),
        button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); tierRows = tierRows.filter(function (r) { return r.row !== row; }); } })), pickers);
      tierRows.push({ row: row, read: function () { var o = pickers.read(); o.stars = numberValue(st) || 0; return o; } });
      tiers.appendChild(row);
    }
    (s.tierPrizes || []).forEach(addTier);
    var group = input(s.rewardGroup, { name: 'rewardGroup', ltr: true, mono: true });
    var cooldown = numInput(s.cooldownSeconds || '', 'cooldownSeconds', t('f.cooldown'));
    var daily = numInput(s.dailyLimit || '', 'dailyLimit', t('f.daily'));
    var xpLimited = check(s.xpWhenLimited, t('f.xpWhenLimited'), 'xpWhenLimited');
    var meta = input(JSON.stringify(s.metadata || {}, null, 2), { name: 'metadata', ltr: true, mono: true, multiline: true, rows: 4 });
    var preview = h('ul', { class: 'text-sm flex flex-col gap-1', 'data-campaign-preview': '' });
    function paint() {
      modeHint.textContent = t('mode.hint.' + mode.value);
      rules.hidden = !ownUnlock.input.checked;
      starRules.hidden = ownStars.input.checked;
      templateBox.hidden = !ownStars.input.checked;
      var tplRules = template.value ? ((templates.filter(function (x) { return x.key === template.value; })[0] || {}).rules || []) : cfg.stars.rules;
      var tplList = templateBox.querySelector('[data-campaign-template-rules]');
      clear(tplList);
      tplRules.forEach(function (rr) { tplList.appendChild(h('li', { class: 'text-secondary-600 dark:text-secondary-300', text: starSentence(rr) })); });
      mult.hidden = false;
      mult.inputs.forEach(function (el) { el.disabled = ownMult.input.checked; });
      if (ownMult.input.checked) {
        var def = mode.value === 'repeatable' ? cfg.stars.repeatableMultipliers : cfg.stars.multipliers;
        mult.inputs.forEach(function (el, i) { el.value = digits(String(def[i])); });
      }
      clear(preview);
      previewLines(cur.read(), numberValue(xp) || 0, mult.read(), max).forEach(function (l) { preview.appendChild(h('li', { text: l })); });
      if (mode.value !== 'repeatable') preview.appendChild(h('li', { class: 'text-secondary-500', text: t('st.preview.replay') }));
    }
    var msg = h('div', {});
    var body = h('div', { 'data-campaign-form': 'stage' },
      h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.key'), key, t('hint.key')), field(t('f.chapter'), chapter)),
      field(t('f.mode'), mode), modeHint, name, desc,
      h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.icon'), icon), field(t('f.address'), address)), active,
      section(t('st.timing'), h('div', {}, h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.timeLimit'), limit), field(t('f.minSeconds'), minSec, t('hint.minSeconds'))), noReplay,
        h('p', { class: 'text-xs text-secondary-500 -mt-2', text: t('st.noReplayHint') })), 'timing'),
      h('div', { class: 'h-3' }), section(t('st.settings'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('st.settingsHint') }), settingsHolder), 'settings'),
      h('div', { class: 'h-3' }), section(t('st.unlock'), unlockBox, 'unlock'),
      h('div', { class: 'h-3' }), section(t('st.stars'), h('div', {}, ownStars, templateBox, starRules), 'stars'),
      h('div', { class: 'h-3' }), section(t('st.rewards'), h('div', {}, cur, h('div', { class: 'mt-3' }, field(t('f.xp'), xp, null, 'cg-num')),
        h('div', { class: 'text-sm font-medium mt-2 mb-1', text: t('st.multipliers') }), ownMult, mult,
        h('div', { class: 'text-sm font-medium mt-3 mb-1', text: t('st.preview') }), preview), 'rewards'),
      h('div', { class: 'h-3' }), section(t('st.tierPrizes'), h('div', {}, tiers, button({ label: '+ ' + t('st.addTier'), data: { 'campaign-action': 'add-tier' }, onclick: function () { addTier(); } })), 'tiers'),
      h('div', { class: 'h-3' }), section(t('st.limits'), h('div', {}, field(t('f.rewardGroup'), group, t('hint.rewardGroup')),
        h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.cooldown'), cooldown), field(t('f.daily'), daily)), xpLimited), 'limits'),
      h('div', { class: 'h-3' }), section(t('st.advanced'), field(t('f.metadata'), meta, t('hint.metadata')), 'advanced'), msg);
    body.addEventListener('input', paint);
    body.addEventListener('change', paint);
    body.addEventListener('click', function (e) { if (e.target.closest('button')) setTimeout(paint, 0); });
    var save = button({ label: t('save'), variant: 'primary', data: { 'campaign-action': 'save-stage' }, onclick: function () {
      busy(save, true);
      clear(msg);
      var md;
      try { md = JSON.parse(meta.value || '{}'); } catch (e) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_JSON') + ' · metadata')); return; }
      var lim = floatValue(limit), mn = floatValue(minSec);
      var badField = settingsForm.invalid();
      if (badField) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_FIELD_VALUE') + ' · ' + badField)); return; }
      var out = { key: key.value.trim(), chapterId: chapter.value, mode: mode.value, name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), address: address.value.trim(),
        sort: creating ? 1000 : s.sort, isActive: active.input.checked, timeLimitSeconds: lim == null ? null : lim, minSeconds: mn == null ? null : mn,
        unlock: ownUnlock.input.checked ? rules.read() : null, starRules: ownStars.input.checked ? null : starRules.read(), starMultipliers: ownMult.input.checked ? null : mult.read(),
        rewards: { currencies: cur.read() }, xp: numberValue(xp) || 0, tierPrizes: tierRows.map(function (r) { return r.read(); }),
        rewardGroup: group.value.trim(), cooldownSeconds: numberValue(cooldown) || 0, dailyLimit: numberValue(daily) || 0, xpWhenLimited: xpLimited.input.checked, metadata: md,
        settings: settingsForm.read(), starTemplate: ownStars.input.checked ? template.value : '', noReplayAfterWin: noReplay.input.checked };
      api(creating ? 'POST' : 'PUT', '/campaign/stages' + (creating ? '' : '/' + s.id), out).then(function () {
        closePanel(); toast('success', t('saved')); done();
      }, function (err) { busy(save, false); msg.appendChild(notice('error', cgErrText(err))); });
    } });
    sidePanel(creating ? t('st.new') : textOf(s.name, s.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'stage');
    paint();
  }, function (err) { toast('error', cgErrText(err)); });
}

function freeridePage(root) {
  var frame = pageFrame('freeride', t('desc.freeride'), [button({ label: t('fr.new'), variant: 'primary', data: { 'campaign-action': 'new-freeride' }, onclick: function () { freerideEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  function reload() {
    clear(frame.content).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/campaign/freeride'), loadLanguages(), loadRefs(true), loadCgSettings()]).then(function (res) {
      var list = res[0].items || [];
      clear(frame.content);
      if (!res[3].settings.freeride.tableEnabled) frame.content.appendChild(notice('warning', t('fr.off')));
      if (!list.length) { frame.content.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      frame.content.appendChild(h('div', { 'data-campaign-list': 'freeride' }, table([
        { label: t('f.name') }, { label: t('f.key') }, { label: t('st.rewards') }, { label: t('st.limits') }, { label: t('f.rewardGroup') }, { label: '' }
      ], list.map(function (f) {
        var r = f.rewards || {};
        var prize = [r.prizeKey, r.prizeGroupKey].filter(Boolean).join(' · ');
        return { id: f.key, item: f, cells: [
          h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(f.name, f.key) }), f.isActive ? null : tag(t('inactive'), 'warning')),
          mono(f.key),
          h('span', { class: 'text-xs', text: rewardText(r.currencies) + (prize ? ' · ' + prize : '') + (f.xp ? ' · ' + fmtNum(f.xp) + ' XP' : '') }),
          h('span', { class: 'text-xs', 'data-campaign-limits': f.key, text: tr('fr.limits', null, { c: f.cooldownSeconds ? t('fr.cooldown', { n: f.cooldownSeconds }) : t('fr.noLimit'), d: f.dailyLimit ? t('fr.daily', { n: f.dailyLimit }) : t('fr.noLimit') }) }),
          f.rewardGroup ? mono(f.rewardGroup) : h('span', { text: t('none') }),
          button({ label: t('delete'), variant: 'danger', data: { 'campaign-delete': f.key }, onclick: function () {
            confirmBox(t('fr.delete', { key: f.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/campaign/freeride/' + f.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', cgErrText(err)); });
            });
          } })
        ] };
      }), function (f) { freerideEditor(f, reload); })));
    }, function (err) { clear(frame.content).appendChild(notice('error', cgErrText(err))); });
  }
  reload();
}

function freerideEditor(f, done) {
  var creating = !f;
  f = f || { name: {}, description: {}, rewards: {}, isActive: true };
  var key = input(f.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = translated(t('f.name'), 'name', f.name);
  var desc = translated(t('f.description'), 'description', f.description, { multiline: true });
  var icon = input(f.iconUrl, { name: 'iconUrl', ltr: true });
  var active = check(f.isActive, t('f.active'), 'isActive');
  var cur = currencyRows((f.rewards || {}).currencies, 'freeride');
  var prizes = prizePickers(f.rewards || {});
  var xp = numInput(f.xp || '', 'xp', 'XP');
  var group = input(f.rewardGroup, { name: 'rewardGroup', ltr: true, mono: true });
  var cooldown = numInput(f.cooldownSeconds || '', 'cooldownSeconds', t('f.cooldown'));
  var daily = numInput(f.dailyLimit || '', 'dailyLimit', t('f.daily'));
  var minLevel = numInput(f.minLevel || '', 'minLevel', t('f.minLevel'));
  var msg = h('div', {});
  var body = h('div', { 'data-campaign-form': 'freeride' }, field(t('f.key'), key), name, desc, field(t('f.icon'), icon), active,
    section(t('st.rewards'), h('div', {}, cur, prizes, field('XP', xp, null, 'cg-num')), 'rewards'), h('div', { class: 'h-3' }),
    section(t('st.limits'), h('div', {}, field(t('f.rewardGroup'), group, t('hint.rewardGroup')), h('div', { class: 'grid sm:grid-cols-3 gap-x-4' }, field(t('f.cooldown'), cooldown), field(t('f.daily'), daily), field(t('f.minLevel'), minLevel))), 'limits'), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'campaign-action': 'save-freeride' }, onclick: function () {
    busy(save, true);
    clear(msg);
    var rw = prizes.read();
    rw.currencies = cur.read();
    var out = { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), isActive: active.input.checked, sort: f.sort || 0, rewards: rw,
      xp: numberValue(xp) || 0, rewardGroup: group.value.trim(), cooldownSeconds: numberValue(cooldown) || 0, dailyLimit: numberValue(daily) || 0, minLevel: numberValue(minLevel) || 0 };
    api(creating ? 'POST' : 'PUT', '/campaign/freeride' + (creating ? '' : '/' + f.id), out).then(function () {
      closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); msg.appendChild(notice('error', cgErrText(err))); });
  } });
  sidePanel(creating ? t('fr.new') : textOf(f.name, f.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'freeride');
}

function reportPage(root) {
  var q = hashQuery();
  var state = { days: Number(q.days) || 30, chapterId: q.chapterId || '' };
  var frame = pageFrame('report', t('desc.report'));
  clear(root).appendChild(frame.el);
  var daysSel = select(String(state.days), [7, 30, 90].map(function (n) { return { value: String(n), label: t('rp.d', { n: n }) }; }), { name: 'days', label: t('rp.days') });
  var chapterSel = h('select', { class: CG_INPUT + ' eco-select', name: 'chapterId', 'aria-label': t('f.chapter') });
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end cg-filters' }, field(t('rp.days'), daysSel), field(t('f.chapter'), chapterSel)));
  frame.content.appendChild(body);
  daysSel.addEventListener('change', function () { state.days = Number(daysSel.value); reload(); });
  chapterSel.addEventListener('change', function () { state.chapterId = chapterSel.value; reload(); });
  function kpi(name, value) {
    return card(h('div', { class: 'p-4', 'data-campaign-kpi': name }, h('div', { class: 'text-xs text-secondary-500', text: t('rp.kpi.' + name) }), h('div', { class: 'text-xl font-semibold mt-1', text: fmtNum(value || 0) })));
  }
  function bar(pct, name) {
    return h('span', { class: 'flex items-center gap-2 min-w-[8rem]' }, h('span', { class: 'cg-bar flex-1' }, h('span', { style: 'width:' + Math.max(0, Math.min(100, pct || 0)) + '%' })),
      h('span', { class: 'text-xs whitespace-nowrap', 'data-campaign-rate': name, text: fmtNum(pct || 0) + '٪' }));
  }
  function reload() {
    clear(body).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/campaign/chapters'), api('GET', '/campaign/report' + qs({ days: state.days, chapterId: state.chapterId, lang: cgLang() }))]).then(function (res) {
      clear(chapterSel);
      [{ value: '', label: t('st.filter.chapter') }].concat((res[0].items || []).map(function (c) { return { value: c.id, label: textOf(c.name, c.key) }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.chapterId) opt.selected = true;
        chapterSel.appendChild(opt);
      });
      var d = res[1];
      clear(body);
      var tot = d.totals || {};
      body.appendChild(h('div', { class: 'grid grid-cols-2 md:grid-cols-4 gap-3' }, kpi('players', tot.players), kpi('attempts', tot.attempts), kpi('wins', tot.wins), kpi('rejected', tot.rejected)));
      var items = d.items || [];
      if (!items.length) { body.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      body.appendChild(h('div', { 'data-campaign-list': 'report' }, table([
        { label: t('rp.stage') }, { label: t('f.mode') }, { label: t('rp.players') }, { label: t('rp.attempts') }, { label: t('rp.wins') }, { label: t('rp.rate') }, { label: t('rp.avgStars') }, { label: t('rp.drop') }, { label: t('rp.rejected') }
      ], items.map(function (r) {
        return { id: r.key, item: r, cells: [
          h('div', { class: 'flex flex-col' }, h('span', { class: 'font-medium', text: r.name }), mono(r.chapterKey + ' / ' + r.key)), modeTag(r.mode),
          h('span', { 'data-campaign-players': r.key, text: fmtNum(r.players) }), h('span', { text: fmtNum(r.attempts) }), h('span', { text: fmtNum(r.wins) }),
          bar(r.completionRate, r.key), h('span', { 'data-campaign-avg': r.key, text: fmtNum(r.avgBestStars) }),
          h('span', { 'data-campaign-drop': r.key, text: r.dropOff == null ? t('none') : fmtNum(r.dropOff) + '٪' }), h('span', { text: fmtNum(r.rejected) })
        ] };
      }))));
    }, function (err) { clear(body).appendChild(notice('error', cgErrText(err))); });
  }
  reload();
}

function campaignSettingsPage(root) {
  Promise.all([loadRefs(true)]).then(function () {
    settingsPage(root, 'settings', '/campaign', function (cur, defs, d) {
      var opt = function (prefix, list) { return list.map(function (v) { return { value: v, label: t(prefix + v) }; }); };
      var notes = h('div', { class: 'flex flex-col gap-2' });
      if (d.economy === false) notes.appendChild(notice('warning', t('se.noEconomy')));
      if (d.xpMetaFound === false) notes.appendChild(notice('warning', t('se.metaMissing', { k: cur.xp.metaKey })));
      var pl = cur.play || defs.play || { costCurrency: '', costAmount: 1, refundOnWin: true };
      if (pl.costCurrency && d.costCurrencyFound === false) notes.appendChild(notice('warning', t('se.costMissing', { k: pl.costCurrency })));
      else if (pl.costCurrency && d.costCurrencyRegen === false) notes.appendChild(notice('info', t('se.costNoRegen', { k: pl.costCurrency })));
      var costCur = refSelect('currencies', pl.costCurrency, 'play.costCurrency', t('se.cost.none'));
      var costAmount = numInput(pl.costAmount, 'play.costAmount');
      var refundOnWin = check(pl.refundOnWin, t('se.refundOnWin'), 'play.refundOnWin');
      var a = cur.attempts, s = cur.stars, u = cur.unlock, r = cur.rewards, x = cur.xp, e = cur.events, f = cur.freeride;
      var hours = numInput(a.hours, 'attempts.hours'), grace = numInput(a.graceSeconds, 'attempts.graceSeconds'), keep = numInput(a.keepDays, 'attempts.keepDays'), minSec = numInput(a.defaultMinSeconds, 'attempts.defaultMinSeconds');
      var closeOthers = check(a.closeOthers, t('se.closeOthers'), 'attempts.closeOthers'), serverCloses = check(a.serverClosesOthers, t('se.serverClosesOthers'), 'attempts.serverClosesOthers');
      var maxStars = numInput(s.maxStars, 'stars.maxStars'), minOn = numInput(s.minOnSuccess, 'stars.minOnSuccess');
      var missing = select(s.missingField, opt('se.missing.', ['pass', 'fail']), { name: 'stars.missingField' });
      var mult = multiplierInputs(s.multipliers, s.maxStars, 'multipliers');
      var rmult = multiplierInputs(s.repeatableMultipliers, s.maxStars, 'repeatableMultipliers');
      var rules = starEditor(s.rules, s.maxStars);
      var stageDefault = select(u.stageDefault, opt('se.stageDefault.', ['all_previous', 'previous', 'none']), { name: 'unlock.stageDefault' });
      var chapterDefault = select(u.chapterDefault, opt('se.chapterDefault.', ['none', 'previous']), { name: 'unlock.chapterDefault' });
      var minStars = numInput(u.minStars, 'unlock.minStars');
      var showLocked = check(u.showLocked, t('se.showLocked'), 'unlock.showLocked');
      var replay = select(r.replayMode, opt('se.replay.', ['difference', 'none', 'full']), { name: 'rewards.replayMode' });
      var zone = input(r.dayZone, { name: 'rewards.dayZone', ltr: true });
      var groupMode = select(r.prizeGroupMode, opt('se.groupMode.', ['draw', 'all']), { name: 'rewards.prizeGroupMode' });
      var modes = ['ordered', 'repeatable', 'once'].map(function (m) { return check((r.milestoneModes || []).indexOf(m) !== -1, t('mode.' + m), 'milestone-' + m); });
      var metaKey = refSelect('metas', x.metaKey, 'xp.metaKey', t('se.metaKey.none'));
      var table_ = refSelect('levelTables', x.levelTableKey, 'xp.levelTableKey', t('se.levelTable.main'));
      var ecoEvent = input(e.economyEvent, { name: 'events.economyEvent', ltr: true, mono: true });
      var anEvent = input(e.analyticsEvent, { name: 'events.analyticsEvent', ltr: true, mono: true });
      var onFailure = check(e.onFailure, t('se.onFailure'), 'events.onFailure');
      var tableOn = check(f.tableEnabled, t('se.tableEnabled'), 'freeride.tableEnabled');
      var reqId = check(f.requireRequestId, t('se.requireRequestId'), 'freeride.requireRequestId');
      var tplBox = h('div', { class: 'flex flex-col gap-3', 'data-campaign-templates': '' });
      var tplItems = [];
      var addTpl = function (tp) {
        tp = tp || { key: '', name: {}, rules: [] };
        var k = input(tp.key, { name: 'template-key', ltr: true, mono: true, label: t('f.key') });
        var nm = translated(t('f.name'), 'template-name', tp.name);
        var ed = starEditor(tp.rules, s.maxStars);
        var el = h('div', { class: 'cg-track', 'data-campaign-template': tp.key || '' },
          h('div', { class: 'flex items-end gap-3' }, field(t('f.key'), k, null, 'flex-1'),
            button({ label: t('remove'), variant: 'danger', onclick: function () { el.remove(); tplItems = tplItems.filter(function (x) { return x.el !== el; }); } })), nm, ed);
        tplItems.push({ el: el, read: function () { return { key: k.value.trim(), name: nm.read(), rules: ed.read() }; } });
        tplBox.appendChild(el);
      };
      (s.templates || []).forEach(addTpl);
      var fieldsRaw = input(JSON.stringify(cur.fields || { stage: [], chapter: [] }, null, 2), { name: 'fields', ltr: true, mono: true, multiline: true, rows: 10 });
      var reqHours = numInput(f.requestHours, 'freeride.requestHours');
      var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
      return {
        el: [notes,
          section(t('se.play'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: t('se.play.hint') }),
            grid(field(t('se.costCurrency'), costCur), field(t('se.costAmount'), costAmount)), refundOnWin), 'play'),
          section(t('se.attempts'), h('div', {}, grid(field(t('se.hours'), hours), field(t('se.grace'), grace)), grid(field(t('se.keepDays'), keep), field(t('se.minSeconds'), minSec)), closeOthers, serverCloses), 'attempts'),
          section(t('se.stars'), h('div', {}, grid(field(t('se.maxStars'), maxStars), field(t('se.minOnSuccess'), minOn)), field(t('se.missing'), missing),
            h('div', { class: 'text-sm font-medium mb-1', text: t('se.multipliers') }), mult, h('div', { class: 'text-sm font-medium mt-3 mb-1', text: t('se.repeatMultipliers') }), rmult,
            h('div', { class: 'text-sm font-medium mt-3 mb-1', text: t('se.rules') }), rules,
            h('div', { class: 'text-sm font-medium mt-4 mb-1', text: t('se.templates') }), h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('se.templatesHint') }), tplBox,
            button({ label: '+ ' + t('se.addTemplate'), data: { 'campaign-action': 'add-template' }, onclick: function () { addTpl(); } })), 'stars'),
          section(t('se.unlock'), h('div', {}, grid(field(t('se.stageDefault'), stageDefault), field(t('se.chapterDefault'), chapterDefault)), field(t('se.minStars'), minStars, null, 'cg-num'), showLocked), 'unlock'),
          section(t('se.rewards'), h('div', {}, grid(field(t('se.replay'), replay), field(t('se.groupMode'), groupMode)), field(t('se.zone'), zone),
            h('div', { class: 'text-sm font-medium mb-1', text: t('se.milestoneModes') }), h('div', { class: 'flex flex-wrap gap-4' }, modes)), 'rewards'),
          section(t('se.xp'), grid(field(t('se.metaKey'), metaKey), field(t('se.levelTable'), table_)), 'xp'),
          section(t('se.events'), h('div', {}, grid(field(t('se.economyEvent'), ecoEvent), field(t('se.analyticsEvent'), anEvent)), onFailure), 'events'),
          section(t('se.freeride'), h('div', {}, tableOn, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('se.tableEnabledHint') }), reqId, field(t('se.requestHours'), reqHours, null, 'cg-num')), 'freeride'),
          section(t('se.fields'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('se.fieldsHint') }), field(t('se.fieldsJson'), fieldsRaw)), 'fields')],
        read: function (c) {
          c.play = { costCurrency: costCur.value || '', costAmount: numberValue(costAmount), refundOnWin: refundOnWin.input.checked };
          c.attempts = { hours: numberValue(hours), graceSeconds: numberValue(grace), keepDays: numberValue(keep), defaultMinSeconds: numberValue(minSec), closeOthers: closeOthers.input.checked, serverClosesOthers: serverCloses.input.checked };
          c.stars = { maxStars: numberValue(maxStars), minOnSuccess: numberValue(minOn), missingField: missing.value, multipliers: mult.read(), repeatableMultipliers: rmult.read(), rules: rules.read(),
            templates: tplItems.map(function (x) { return x.read(); }) };
          try { c.fields = JSON.parse(fieldsRaw.value || '{}'); } catch (err) { return t('code.INVALID_JSON') + ' · fields'; }
          c.unlock = { stageDefault: stageDefault.value, chapterDefault: chapterDefault.value, minStars: numberValue(minStars), showLocked: showLocked.input.checked };
          c.rewards = { replayMode: replay.value, dayZone: zone.value.trim(), prizeGroupMode: groupMode.value,
            milestoneModes: ['ordered', 'repeatable', 'once'].filter(function (m, i) { return modes[i].input.checked; }) };
          c.xp = { metaKey: metaKey.value, levelTableKey: table_.value };
          c.events = { economyEvent: ecoEvent.value.trim(), analyticsEvent: anEvent.value.trim(), onFailure: onFailure.input.checked };
          c.freeride = { tableEnabled: tableOn.input.checked, requireRequestId: reqId.input.checked, requestHours: numberValue(reqHours) };
          return null;
        },
        after: function () { cgSettingsCache = null; cgNavSettings = false; cgMounted = null; cgSchedule(); }
      };
    });
  });
}

var CG_PLAYER = { id: null, data: null, loading: false, at: 0, resetClaims: false };

function cgPlayerRoute() {
  var m = /^#\/players\/([0-9a-f-]{36})(?:\?.*)?$/.exec(location.hash || '');
  return m ? m[1] : null;
}

function cgLoadPlayer(id) {
  CG_PLAYER.loading = true;
  return api('GET', '/campaign/players/' + id + qs({ lang: cgLang() })).then(function (d) { return d; }, function (err) { return { error: err }; }).then(function (d) {
    CG_PLAYER.loading = false;
    CG_PLAYER.at = Date.now();
    if (CG_PLAYER.id === id) CG_PLAYER.data = d;
    cgSchedule();
  });
}

function cgReloadPlayer() {
  CG_PLAYER.at = 0;
  if (CG_PLAYER.id) cgLoadPlayer(CG_PLAYER.id);
}

function cgPlayerBody(id, d) {
  if (d.error) return notice('warning', t('pc.noCampaign'));
  var box = h('div', { class: 'flex flex-col gap-3' });
  box.appendChild(h('p', { class: 'text-sm', 'data-campaign-player-xp': '', text: t('pc.xp', { xp: fmtNum(d.xp), level: d.level }) }));
  var claims = check(CG_PLAYER.resetClaims, t('pc.reset.claims'), 'reset-claims');
  claims.input.addEventListener('change', function () { CG_PLAYER.resetClaims = claims.input.checked; });
  (d.chapters || []).forEach(function (ch) {
    var played = (ch.stages || []).filter(function (s) { return s.played; });
    var list = h('ol', { class: 'flex flex-col gap-1' });
    (ch.stages || []).forEach(function (s) {
      var reset = s.played ? button({ label: t('pc.reset'), variant: 'danger', data: { 'campaign-reset': s.key }, onclick: function () {
        confirmBox(t('pc.reset.confirm', { name: s.name }), t('pc.reset')).then(function (ok) {
          if (!ok) return;
          api('DELETE', '/campaign/players/' + id + '/stages/' + s.id + (CG_PLAYER.resetClaims ? '?claims=reset' : '')).then(function () { toast('success', t('saved')); cgReloadPlayer(); },
            function (err) { toast('error', cgErrText(err)); });
        });
      } }) : null;
      list.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-campaign-player-stage': s.key },
        h('span', { class: 'font-medium', text: s.name }), modeTag(s.mode), stars(s.bestStars, 3),
        h('span', { class: 'text-xs text-secondary-500', text: s.played ? t('pc.wins', { w: s.wins, c: s.completions }) : (s.unlocked ? t('pc.notPlayed') : t('pc.locked')) }),
        s.lastCompletedAt ? h('span', { class: 'text-xs text-secondary-500', text: fmtDate(s.lastCompletedAt) }) : null, reset));
    });
    box.appendChild(h('div', { 'data-campaign-player-chapter': ch.key },
      h('div', { class: 'flex items-center gap-2 mb-1' }, h('span', { class: 'text-xs font-semibold text-secondary-600 dark:text-secondary-300', text: ch.name }),
        h('span', { class: 'cg-bar flex-1 max-w-[10rem]' }, h('span', { style: 'width:' + (ch.starPercent || 0) + '%' })),
        h('span', { class: 'text-xs', text: digits(String(ch.stars)) + '/' + digits(String(ch.maxStars)) + ' ★' })),
      played.length || (ch.stages || []).length ? list : null));
  });
  if ((d.attempts || []).length) {
    var ol = h('ol', { class: 'flex flex-col gap-1', 'data-campaign-player-attempts': '' });
    d.attempts.forEach(function (a) {
      ol.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-xs' }, h('span', { class: 'text-secondary-500', text: fmtDate(a.startedAt) }), mono(a.stageKey),
        tag(t('att.' + a.status), a.status === 'done' ? 'success' : a.status === 'open' ? 'primary' : 'secondary'), a.status === 'done' ? stars(a.stars, 3) : null,
        a.rejected ? tag(a.lastReject + ' ×' + digits(String(a.rejected)), 'error') : null));
    });
    box.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pc.attempts') }), ol));
  }
  box.appendChild(claims);
  return box;
}

function cgPlayerCard() {
  var id = cgPlayerRoute();
  var existing = document.querySelector('[data-campaign-player]');
  if (!id || !token()) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (CG_PLAYER.id !== id) { CG_PLAYER.id = id; CG_PLAYER.data = null; CG_PLAYER.at = 0; }
  if (!CG_PLAYER.loading && Date.now() - CG_PLAYER.at > 20000) cgLoadPlayer(id);
  var d = CG_PLAYER.data;
  var key = cgLang() + '|' + id + '|' + (d ? JSON.stringify(d) : 'loading');
  if (existing && existing.getAttribute('data-key') === key && existing.parentNode === grid) return;
  if (existing && existing.contains(document.activeElement) && existing.parentNode === grid) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' }, h('h2', { id: 'campaign-player-title', class: 'font-medium mb-3', text: t('pc.title') }));
  body.appendChild(d ? cgPlayerBody(id, d) : h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  grid.appendChild(h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-campaign-player': '', 'data-key': key, 'data-l10n-skip': '',
    'aria-labelledby': 'campaign-player-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body));
}

var CG_ROOT = 'campaign-root';
var CG_PAGES = ['chapters', 'stages', 'freeride', 'report', 'settings'];
var CG_RENDER = { chapters: chaptersPage, stages: stagesPage, freeride: freeridePage, report: reportPage, settings: campaignSettingsPage };
var CG_HASH = /^#\/campaign(?:\/([^/?]*))?/;
var cgPending = null;
var cgRouter = null;
var cgMounted = null;
var cgScheduled = false;

function cgRoute(hash) {
  var m = CG_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  return CG_PAGES.indexOf(m[1]) !== -1 ? m[1] : CG_PAGES[0];
}

function cgAllowed() {
  return !!token();
}

function cgTitle() {
  var page = cgRoute();
  return t('group') + (page ? ' - ' + t('nav.' + page) : '');
}

function cgAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || cgRouter) return !!cgRouter;
  try {
    if (!r.hasRoute('campaign')) {
      var layout = r.resolve('/').matched[0];
      r.addRoute({ path: '/campaign/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'campaign', component: { name: 'CampaignPage', render: function () { return null; } }, meta: { get title() { return cgTitle(); } } }] });
    }
    cgRouter = r;
  } catch (e) {
    return false;
  }
  if (cgPending) {
    var target = cgPending;
    cgPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function cgNavLink(page, active) {
  var on = page === active;
  return h('li', {}, h('a', { href: '#/campaign/' + page, 'data-campaign-nav': page, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + page) }))));
}

var cgNavSettings = false;
function cgVisiblePages(active) {
  if (!cgNavSettings && token()) {
    cgNavSettings = true;
    loadCgSettings().then(cgSchedule, function () { cgNavSettings = false; });
  }
  var off = cgSettingsCache && cgSettingsCache.settings && cgSettingsCache.settings.freeride && !cgSettingsCache.settings.freeride.tableEnabled;
  return CG_PAGES.filter(function (p) { return p !== 'freeride' || !off || active === 'freeride'; });
}

function cgSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelector('[data-campaign-group]');
  if (!group || !cgAllowed()) { if (existing) existing.remove(); return; }
  var target = document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = cgLang() + '|' + active + '|' + cgVisiblePages(active).length;
  if (existing && existing.getAttribute('data-key') === key && existing.nextElementSibling === target) return;
  var built = h('div', { 'data-campaign-group': '', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group') }),
    h('nav', { 'aria-label': t('group') }, h('ul', { class: 'space-y-0.5' }, cgVisiblePages(active).map(function (p) { return cgNavLink(p, active); }))));
  if (existing) existing.remove();
  target.before(built);
}

function cgShowMains() {
  var hidden = document.querySelectorAll('[data-campaign-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-campaign-hidden'); }
}

function cgUnmount() {
  var root = document.getElementById(CG_ROOT);
  if (root) root.remove();
  cgShowMains();
  closePanel();
  cgMounted = null;
}

function cgSync() {
  cgAddRoutes();
  var page = cgRoute();
  cgSyncNav(page);
  cgPlayerCard();
  if (!page) { if (cgMounted || document.getElementById(CG_ROOT)) cgUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-campaign-hidden', ''); }
  }
  var root = document.getElementById(CG_ROOT);
  if (!root) {
    root = h('div', { id: CG_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-campaign': 'root' });
    cgMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + cgLang() + (page === 'stages' || page === 'report' ? '|' + location.hash : '');
  if (cgMounted !== key) {
    cgMounted = key;
    closePanel();
    CG_RENDER[page](root);
  }
}

function cgSchedule() {
  if (cgScheduled) return;
  cgScheduled = true;
  requestAnimationFrame(function () { cgScheduled = false; cgSync(); });
}

function cgBoot() {
  window.addEventListener('hashchange', cgSchedule);
  window.addEventListener('popstate', cgSchedule);
  window.addEventListener('nakama-console:ready', cgSchedule);
  new MutationObserver(cgSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { cgMounted = null; cgSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (cgAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  cgSync();
}

window.NakamaCampaign = { version: 1, pages: function () { return cgVisiblePages('').map(function (p) { return { id: p, href: '#/campaign/' + p, label: t('nav.' + p) }; }); } };

if (CG_HASH.test(location.hash)) {
  cgPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', cgBoot);
else cgBoot();
})();
