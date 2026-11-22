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
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'page.first': 'صفحهٔ اول', 'page.last': 'صفحهٔ آخر', 'page.of': 'صفحهٔ {page} از {pages}', 'pager': 'صفحه‌بندی', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'yes': 'بله', 'no': 'خیر',
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
    'se.items': 'آیتم مجاز در مرحله', 'se.itemsHint': 'هر مرحله می‌تواند بگوید با کدام آیتم (مثلاً ماشین) بازی می‌شود. کلاینت کلید آیتم را با campaign_attempt_start می‌فرستد و سرور مالکیت و مجاز بودن را بررسی می‌کند.',
    'se.itemsEnforce': 'بررسی در شروع مرحله روشن باشد', 'se.itemsRequire': 'مرحله‌ای که قانون دارد، بدون آیتم شروع نشود', 'se.itemsOwned': 'بازیکن باید آیتم را داشته باشد',
    'se.itemsParam': 'نام فیلد در درخواست', 'se.itemsType': 'نوع آیتم (خالی یعنی همه)', 'se.itemsAllow': 'مسیر فهرست مجازها', 'se.itemsDeny': 'مسیر فهرست ممنوع‌ها',
    'se.itemsClasses': 'مسیر کلاس‌های مجاز', 'se.itemsClassField': 'فیلد کلاس در پروفایل آیتم', 'se.itemsPathHint': 'مسیر در تنظیمات مرحله، مثل missions.acceptedCars؛ چند مسیر را با ویرگول جدا کنید',
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
    'loading': 'Loading…', 'empty': 'Nothing found', 'page.first': 'First page', 'page.last': 'Last page', 'page.of': 'Page {page} of {pages}', 'pager': 'Pagination', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'yes': 'Yes', 'no': 'No',
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
    'se.items': 'Allowed item per stage', 'se.itemsHint': 'Each stage can say which item (for example a car) it is played with. The client sends the item key with campaign_attempt_start and the server checks ownership and the stage rules.',
    'se.itemsEnforce': 'Check at stage start', 'se.itemsRequire': 'A stage with rules cannot start without an item', 'se.itemsOwned': 'The player must own the item',
    'se.itemsParam': 'Request field name', 'se.itemsType': 'Item type (empty = any)', 'se.itemsAllow': 'Path of the allowed list', 'se.itemsDeny': 'Path of the excluded list',
    'se.itemsClasses': 'Path of the allowed classes', 'se.itemsClassField': 'Class field in the item profile', 'se.itemsPathHint': 'A path in the stage settings, like missions.acceptedCars; separate several with commas',
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
  if (window.NakamaDate) return window.NakamaDate.format(v);
  var d = new Date(v);
  if (isNaN(d)) return '';
  try {
    return d.toLocaleString(cgLang() === 'fa' ? 'fa-IR' : 'en-GB', { calendar: cgLang() === 'fa' ? 'persian' : 'gregory', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

function tr(key, vars, raw) {
  var s = t(key, vars);
  Object.keys(raw || {}).forEach(function (k) { s = s.split('{' + k + '}').join(String(raw[k])); });
  return s;
}
