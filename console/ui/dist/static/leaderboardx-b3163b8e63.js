(function () {
'use strict';
var LB_TEXT = {
  fa: {
    'group': 'جدول امتیاز',
    'nav.boards': 'جدول‌ها',
    'nav.standings': 'رتبه‌های زنده',
    'nav.winners': 'برنده‌های هر دوره',
    'nav.report': 'گزارش جدول‌ها',
    'nav.settings': 'تنظیمات جدول‌ها',
    'desc.boards': 'هر جدول یک لیدربورد Nakama است: دورهٔ ریست، ترتیب، این‌که امتیاز از کجا می‌آید، سطح‌ها و جایزهٔ پایان دوره.',
    'desc.standings': 'رتبهٔ بازیکن‌ها در دورهٔ جاری، با سطح هر نفر. امتیاز را می‌شود دستی عوض کرد یا دوره را همین حالا بست.',
    'desc.winners': 'هر دورهٔ بسته‌شده با رتبه‌ها، سطح‌ها و وضعیت پرداخت جایزهٔ هر بازیکن.',
    'desc.report': 'امتیازهای ثبت‌شده، امتیازهای ردشده، بازیکن‌های دورهٔ جاری و جایزه‌های پرداخت‌شده در روزهای اخیر.',
    'desc.settings': 'تعداد ردیف‌ها، روش پرداخت جایزه، سقف ارسال امتیاز، نگهداری تاریخچه و رویدادهای پایان دوره.',
    'save': 'ذخیره', 'saved': 'ذخیره شد', 'cancel': 'انصراف', 'close': 'بستن', 'remove': 'حذف', 'delete': 'حذف', 'all': 'همه',
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'yes': 'بله', 'no': 'خیر',
    'inactive': 'غیرفعال', 'active': 'فعال', 'confirm.title': 'مطمئن هستید؟', 'lang.text': '{label} ({lang})', 'updated': 'آخرین تغییر: {by} · {at}',
    'none': '—', 'new': 'جدول تازه', 'search': 'جستجو',
    'f.key': 'کلید', 'f.name': 'نام', 'f.description': 'توضیح', 'f.icon': 'آدرس آیکون', 'f.sort': 'ترتیب نمایش', 'f.active': 'فعال باشد',
    'f.period': 'دورهٔ ریست', 'f.cron': 'CRON دلخواه (UTC)', 'f.startsAt': 'شروع', 'f.endsAt': 'پایان', 'f.maxPeriods': 'تعداد دوره (۰ = بی‌پایان)',
    'f.sortOrder': 'ترتیب', 'f.operator': 'روش ثبت امتیاز', 'f.source': 'امتیاز از کجا می‌آید', 'f.event': 'کلید رویداد', 'f.context': 'فقط این زمینه (مثلاً کلید مرحله)',
    'f.field': 'فیلد عددی رویداد', 'f.calc': 'محاسبه', 'f.multiplier': 'ضریب', 'f.clientSubmit': 'بازی بتواند امتیاز بفرستد (leaderboardx_submit)',
    'f.maxPerWrite': 'بیشترین امتیاز هر ارسال (۰ = سقف کلی در تنظیمات)', 'f.perMinute': 'سقف ارسال هر بازیکن در دقیقه (۰ = بدون سقف)', 'f.onExceed': 'وقتی بیشتر از سقف بود',
    'f.tierBasis': 'سطح‌بندی بر اساس', 'f.rewardMinScore': 'کمترین امتیاز برای جایزه (۰ = بدون شرط)', 'f.payMode': 'پرداخت جایزه',
    'hint.key': 'همان شناسهٔ لیدربورد در Nakama؛ بعداً عوض نمی‌شود. اگر لیدربوردی با این شناسه هست، همان به کار می‌رود.',
    'hint.cron': 'پنج بخش: دقیقه ساعت روز ماه روز‌هفته، به وقت UTC. مثال: 0 0 * * 5 یعنی هر جمعه ساعت ۰.',
    'hint.window': 'بیرون از این بازه امتیاز گرفته نمی‌شود. با پایان بازه، دورهٔ آخر بسته و جایزه‌اش داده می‌شود.',
    'hint.multiplier': 'مقدار رویداد در این عدد ضرب و گرد می‌شود (امتیاز Nakama عدد صحیح است).',
    'period.none': 'بدون ریست', 'period.daily': 'روزانه (۰۰:۰۰ UTC)', 'period.weekly': 'هفتگی (دوشنبه ۰۰:۰۰ UTC)', 'period.monthly': 'ماهانه (روز اول)', 'period.custom': 'CRON دلخواه',
    'sort.desc': 'امتیاز بیشتر بهتر است', 'sort.asc': 'امتیاز کمتر بهتر است (مثل زمان)',
    'op.incr': 'جمع شود', 'op.best': 'بهترین امتیاز بماند', 'op.set': 'آخرین امتیاز بماند', 'op.decr': 'کم شود',
    'src.economy_event': 'رویداد اقتصاد (بازی یا سرور می‌فرستد)', 'src.game_event': 'رویداد سرور افزونه‌ها (مثل پایان مرحله)', 'src.client': 'فقط ارسال مستقیم بازی', 'src.server': 'فقط کد سرور و ادمین',
    'calc.value': 'مقدار رویداد', 'calc.count': 'هر رویداد یک امتیاز',
    'exceed.reject': 'رد شود', 'exceed.clamp': 'به سقف کم شود',
    'basis.none': 'بدون سطح', 'basis.score': 'امتیاز', 'basis.rank': 'رتبه', 'basis.percent': 'درصد بالای جدول',
    'pay.': 'پیش‌فرض تنظیمات', 'pay.auto': 'خودکار در پایان دوره', 'pay.claim': 'بازیکن خودش بگیرد',
    'sec.period': 'دوره و بازهٔ زمانی', 'sec.scoring': 'امتیاز', 'sec.limits': 'جلوگیری از تقلب', 'sec.tiers': 'سطح‌ها', 'sec.rewards': 'جایزهٔ پایان دوره',
    'tier.add': 'سطح تازه', 'tier.key': 'کلید سطح', 'tier.threshold.score': 'از امتیاز', 'tier.threshold.rank': 'تا رتبهٔ', 'tier.threshold.percent': 'درصد بالا',
    'tier.s.score': 'امتیاز {v} یا بهتر', 'tier.s.rank': 'رتبهٔ {from} تا {v}', 'tier.s.percent': '{v}٪ بالای جدول', 'tier.hint': 'بهترین سطح اول. هر بازیکن در اولین سطحی است که شرطش را دارد.',
    'rw.add': 'ردیف جایزهٔ تازه', 'rw.tier': 'برای', 'rw.allRanks': 'رتبهٔ کل جدول', 'rw.inTier': 'داخل سطح {t}', 'rw.from': 'از رتبهٔ', 'rw.to': 'تا رتبهٔ',
    'rw.currency': 'ارز', 'rw.amount': 'مقدار', 'rw.addCurrency': 'ارز', 'rw.prize': 'جایزهٔ اقتصاد', 'rw.group': 'گروه جایزه', 'rw.noPrize': 'بدون جایزه', 'rw.missing': 'پیدا نشد',
    'rw.s.rank': 'رتبهٔ {from} تا {to}', 'rw.s.one': 'رتبهٔ {from}', 'rw.s.tierAll': 'همهٔ سطح {t}', 'rw.s.tier': 'رتبهٔ {from} تا {to} سطح {t}',
    'rw.hint': 'هر بازیکن در هر دوره یک بار جایزه می‌گیرد. اگر چند ردیف به او برسد، تنظیمات می‌گوید اولی یا همه.',
    'b.delete': 'جدول {key} حذف شود؟ لیدربورد Nakama و امتیازهایش هم پاک می‌شود؛ تاریخچهٔ دوره‌ها می‌ماند.', 'b.inUse': 'جایزه‌های این جدول هنوز پرداخت یا گرفته نشده است.',
    'b.players': '{n} بازیکن', 'b.ends': 'پایان دوره: {at}', 'b.noEnd': 'دوره پایان ندارد', 'b.tiers': '{n} سطح', 'b.rewards': '{n} ردیف جایزه',
    'b.restart': 'این تغییر (ترتیب، روش یا دورهٔ ریست) لیدربورد Nakama را از نو می‌سازد. دورهٔ جاری {n} بازیکن دارد و اول بسته و بایگانی می‌شود.',
    'b.restartPay': 'جایزهٔ دورهٔ جاری داده شود', 'b.restartGo': 'بستن دوره و ذخیره',
    'st.upcoming': 'هنوز شروع نشده', 'st.active': 'فعال', 'st.ended': 'تمام شده', 'st.inactive': 'غیرفعال',
    'sd.board': 'جدول', 'sd.find': 'نام کاربری یا شناسهٔ بازیکن', 'sd.rank': 'رتبه', 'sd.player': 'بازیکن', 'sd.score': 'امتیاز', 'sd.tier': 'سطح',
    'sd.edit': 'تغییر امتیاز', 'sd.mode': 'روش', 'sd.mode.set': 'همین عدد بشود', 'sd.mode.add': 'اضافه شود', 'sd.removeScore': 'امتیاز {name} در این دوره پاک شود؟',
    'sd.close': 'بستن دوره همین حالا', 'sd.closePay': 'جایزهٔ این دوره داده شود', 'sd.closeConfirm': 'دورهٔ جاری بسته و بایگانی می‌شود و جدول از صفر شروع می‌کند.',
    'sd.closed': 'دوره بسته شد', 'sd.period': 'دورهٔ جاری: {from} تا {to}', 'sd.noBoards': 'هنوز جدولی ساخته نشده است.',
    'wn.end': 'پایان دوره', 'wn.reason': 'علت', 'wn.players': 'بازیکن', 'wn.rewarded': 'جایزه‌دار', 'wn.status': 'پرداخت', 'wn.retry': 'تلاش دوباره',
    'wn.retried': '{n} جایزه دوباره در صف رفت، {p} پرداخت شد', 'wn.allBoards': 'همهٔ جدول‌ها', 'wn.onlyRewarded': 'فقط جایزه‌دارها', 'wn.error': 'خطا',
    'reason.reset': 'ریست خودکار', 'reason.manual': 'بستن دستی', 'reason.restart': 'تغییر جدول', 'reason.end': 'پایان بازه',
    'rs.none': 'بدون جایزه', 'rs.pending': 'در صف پرداخت', 'rs.claimable': 'منتظر گرفتن', 'rs.paid': 'پرداخت شد', 'rs.failed': 'ناموفق', 'rs.expired': 'مهلت گذشت',
    'rp.days': 'بازه', 'rp.d': '{n} روز اخیر', 'rp.kpi.boards': 'جدول فعال', 'rp.kpi.players': 'بازیکن دورهٔ جاری', 'rp.kpi.writes': 'امتیاز ثبت‌شده', 'rp.kpi.refused': 'امتیاز ردشده',
    'rp.kpi.paid': 'جایزهٔ پرداخت‌شده', 'rp.kpi.open': 'جایزهٔ باز', 'rp.kpi.failed': 'پرداخت ناموفق', 'rp.writers': 'بازیکن‌های فعال', 'rp.periods': 'دورهٔ بسته', 'rp.daily': 'امتیازهای هر روز',
    'se.display': 'نمایش', 'se.topLimit': 'ردیف‌های پیش‌فرض هر جدول', 'se.maxLimit': 'بیشترین ردیف در یک درخواست', 'se.aroundLimit': 'ردیف‌های اطراف بازیکن', 'se.showDisplayName': 'نام نمایشی و آواتار در ردیف‌ها باشد',
    'se.showUpcoming': 'جدول‌هایی که هنوز شروع نشده‌اند نشان داده شوند', 'se.showEndedDays': 'جدول تمام‌شده تا چند روز دیده شود',
    'se.rewards': 'جایزه', 'se.payMode': 'روش پرداخت پیش‌فرض', 'se.claimDays': 'مهلت گرفتن جایزه (روز)', 'se.stack': 'اگر چند ردیف جایزه به یک بازیکن برسد', 'se.stack.first': 'فقط اولین ردیف', 'se.stack.all': 'همهٔ ردیف‌ها جمع شوند',
    'se.groupMode': 'گروه جایزه', 'se.groupMode.draw': 'یکی به قید قرعه', 'se.groupMode.all': 'همهٔ جایزه‌های گروه', 'se.maxTries': 'دفعه‌های تلاش برای پرداخت', 'se.catchUpDays': 'دوره‌ای که بیش از این روز پیش تمام شده (سرور خاموش بوده) بدون جایزه بایگانی شود',
    'se.scores': 'ارسال امتیاز', 'se.requireRequestId': 'ارسال بازی requestId داشته باشد (ارسال تکراری دو بار حساب نشود)', 'se.playerPerMinute': 'سقف ارسال هر بازیکن در دقیقه در همهٔ جدول‌ها', 'se.logDays': 'نگهداری گزارش امتیازها (روز)', 'se.clientMaxPerWrite': 'سقف هر امتیاز که بازی می‌فرستد، برای جدولی که سقف خودش را ندارد (۰ = بدون سقف)',
    'se.history': 'تاریخچه', 'se.keepDays': 'نگهداری دوره‌ها (روز، ۰ = همیشه)', 'se.playerLimit': 'ردیف‌های تاریخچهٔ هر بازیکن', 'se.winnersLimit': 'ردیف‌های فهرست برنده‌ها',
    'se.events': 'رویدادهای پایان دوره', 'se.economyEvent': 'رویداد اقتصاد برای جایزه‌دارها (مقدار = رتبه)', 'se.analyticsEvent': 'رویداد آنالیتیکس (زمینه = کلید جدول)', 'se.notify': 'به بازیکن اعلان فرستاده شود', 'se.notifyCode': 'کد اعلان (۰ = ۲۳۰۱)',
    'se.noPrizes': 'افزونهٔ اقتصاد بارگذاری نشده؛ جایزه‌های اقتصاد پرداخت نمی‌شوند (ارزها پرداخت می‌شوند).',
    'pc.title': 'جدول امتیاز', 'pc.current': 'رتبه در دورهٔ جاری', 'pc.history': 'دوره‌های گذشته', 'pc.none': 'در هیچ جدولی امتیاز ندارد.', 'pc.refused': '{n} امتیاز ردشده در ۳۰ روز اخیر',
    'pc.rank': 'رتبهٔ {r}', 'pc.score': '{s} امتیاز',
    'code.KEY_EXISTS': 'این کلید قبلاً استفاده شده است.', 'code.INVALID_KEY': 'کلید درست نیست (حروف انگلیسی، رقم، _ . : -).', 'code.VALUE_OUT_OF_RANGE': 'عدد خارج از بازهٔ مجاز است.',
    'code.INVALID_POLICY': 'یکی از گزینه‌ها معتبر نیست.', 'code.DEFAULT_LANGUAGE_REQUIRED': 'نام به زبان پیش‌فرض لازم است.', 'code.INVALID_FIELD': 'نام فیلد درست نیست.',
    'f.scope': 'رتبه‌بندی چه کسی', 'scope.player': 'بازیکن‌ها', 'scope.clan': 'کلن‌ها', 'hint.scope': 'در جدول کلن هر ردیف یک کلن است و امتیاز عضوها برای کلنشان جمع می‌شود.',
    'hint.clanNoRewards': 'جدول کلن جایزه ندارد و بازی نمی‌تواند مستقیم امتیاز بفرستد.', 'b.clan': 'کلن',
    'code.LEADERBOARD_CLAN_NO_REWARDS': 'جدول کلن نمی‌تواند جایزه داشته باشد.', 'code.LEADERBOARD_CLAN_NO_SUBMIT': 'در جدول کلن بازی نمی‌تواند مستقیم امتیاز بفرستد.',
    'code.LEADERBOARD_INVALID_SCHEDULE': 'دورهٔ ریست یا CRON درست نیست.', 'code.LEADERBOARD_INVALID_WINDOW': 'پایان باید بعد از شروع باشد.',
    'code.LEADERBOARD_COUNT_NEEDS_INCR': 'برای «هر رویداد یک امتیاز» روش ثبت باید «جمع شود» باشد.', 'code.LEADERBOARD_TIER_ORDER': 'سطح‌ها باید از بهترین به پایین مرتب باشند.',
    'code.LEADERBOARD_TIER_NOT_FOUND': 'این سطح در جدول نیست.', 'code.LEADERBOARD_INVALID_RANKS': 'بازهٔ رتبه درست نیست (بدون سطح، «تا رتبه» لازم است).',
    'code.LEADERBOARD_PRIZE_NOT_FOUND': 'جایزهٔ انتخاب‌شده پیدا نشد.', 'code.LEADERBOARD_EMPTY_REWARD': 'ردیف جایزه خالی است.', 'code.LEADERBOARD_KEY_IS_TOURNAMENT': 'این شناسه مال یک تورنمنت است.',
    'code.LEADERBOARD_RESTART_REQUIRED': 'برای این تغییر باید دورهٔ جاری بسته شود.', 'code.TOO_MANY_ROWS': 'ردیف‌ها بیش از حد مجاز است.', 'code.IN_USE': 'جایزه‌های این جدول هنوز تسویه نشده است.',
    'code.USER_NOT_FOUND': 'بازیکن پیدا نشد.', 'code.INVALID_JSON': 'دادهٔ فرستاده‌شده درست نیست.', 'code.VALUE_TOO_LONG': 'متن بیش از حد بلند است.'
  },
  en: {
    'group': 'Leaderboards',
    'nav.boards': 'Boards',
    'nav.standings': 'Live standings',
    'nav.winners': 'Winners by period',
    'nav.report': 'Leaderboard report',
    'nav.settings': 'Leaderboard settings',
    'desc.boards': 'Each board is a Nakama leaderboard: reset period, sort order, where scores come from, tiers and the period-end rewards.',
    'desc.standings': 'Players of the running period with their tiers. Change a score by hand or close the period now.',
    'desc.winners': 'Every closed period with ranks, tiers and the payment state of each reward.',
    'desc.report': 'Scores written and refused, players of the running periods and rewards paid in the last days.',
    'desc.settings': 'Row counts, how rewards are paid, submit caps, history retention and period-end events.',
    'save': 'Save', 'saved': 'Saved', 'cancel': 'Cancel', 'close': 'Close', 'remove': 'Remove', 'delete': 'Delete', 'all': 'All',
    'loading': 'Loading…', 'empty': 'Nothing found', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'yes': 'Yes', 'no': 'No',
    'inactive': 'Inactive', 'active': 'Active', 'confirm.title': 'Are you sure?', 'lang.text': '{label} ({lang})', 'updated': 'Last change: {by} · {at}',
    'none': '—', 'new': 'New board', 'search': 'Search',
    'f.key': 'Key', 'f.name': 'Name', 'f.description': 'Description', 'f.icon': 'Icon URL', 'f.sort': 'Display order', 'f.active': 'Active',
    'f.period': 'Reset period', 'f.cron': 'Custom CRON (UTC)', 'f.startsAt': 'Starts', 'f.endsAt': 'Ends', 'f.maxPeriods': 'Number of periods (0 = endless)',
    'f.sortOrder': 'Order', 'f.operator': 'Score write', 'f.source': 'Where scores come from', 'f.event': 'Event key', 'f.context': 'Only this context (e.g. a stage key)',
    'f.field': 'Number field of the event', 'f.calc': 'Calculation', 'f.multiplier': 'Multiplier', 'f.clientSubmit': 'The game may send scores (leaderboardx_submit)',
    'f.maxPerWrite': 'Largest score per write (0 = the cap in settings)', 'f.perMinute': 'Writes per player per minute (0 = no cap)', 'f.onExceed': 'Above the cap',
    'f.tierBasis': 'Tiers by', 'f.rewardMinScore': 'Lowest score for a reward (0 = any)', 'f.payMode': 'Reward payment',
    'hint.key': 'The Nakama leaderboard id; it cannot change later. An existing leaderboard with this id is used.',
    'hint.cron': 'Five parts: minute hour day month weekday, in UTC. Example: 0 0 * * 5 = every Friday at 00:00.',
    'hint.window': 'No scores outside this window. When it ends, the last period is closed and paid.',
    'hint.multiplier': 'The event value is multiplied and rounded (Nakama scores are whole numbers).',
    'period.none': 'Never resets', 'period.daily': 'Daily (00:00 UTC)', 'period.weekly': 'Weekly (Monday 00:00 UTC)', 'period.monthly': 'Monthly (1st)', 'period.custom': 'Custom CRON',
    'sort.desc': 'Higher is better', 'sort.asc': 'Lower is better (like a time)',
    'op.incr': 'Add up', 'op.best': 'Keep the best', 'op.set': 'Keep the latest', 'op.decr': 'Subtract',
    'src.economy_event': 'Economy event (sent by the game or the server)', 'src.game_event': 'Server event of a plugin (e.g. a stage result)', 'src.client': 'Only direct submits from the game', 'src.server': 'Only server code and admins',
    'calc.value': 'Event value', 'calc.count': 'One point per event',
    'exceed.reject': 'Refuse', 'exceed.clamp': 'Lower to the cap',
    'basis.none': 'No tiers', 'basis.score': 'Score', 'basis.rank': 'Rank', 'basis.percent': 'Top percent',
    'pay.': 'Settings default', 'pay.auto': 'Automatically at period end', 'pay.claim': 'The player claims it',
    'sec.period': 'Period and window', 'sec.scoring': 'Scoring', 'sec.limits': 'Anti-cheat', 'sec.tiers': 'Tiers', 'sec.rewards': 'Period-end rewards',
    'tier.add': 'New tier', 'tier.key': 'Tier key', 'tier.threshold.score': 'From score', 'tier.threshold.rank': 'Down to rank', 'tier.threshold.percent': 'Top percent',
    'tier.s.score': 'Score {v} or better', 'tier.s.rank': 'Ranks {from}–{v}', 'tier.s.percent': 'Top {v}%', 'tier.hint': 'Best tier first. A player is in the first tier whose condition holds.',
    'rw.add': 'New reward row', 'rw.tier': 'For', 'rw.allRanks': 'Overall rank', 'rw.inTier': 'Inside tier {t}', 'rw.from': 'From rank', 'rw.to': 'To rank',
    'rw.currency': 'Currency', 'rw.amount': 'Amount', 'rw.addCurrency': 'Currency', 'rw.prize': 'Economy prize', 'rw.group': 'Prize group', 'rw.noPrize': 'No prize', 'rw.missing': 'missing',
    'rw.s.rank': 'Ranks {from}–{to}', 'rw.s.one': 'Rank {from}', 'rw.s.tierAll': 'All of tier {t}', 'rw.s.tier': 'Ranks {from}–{to} of tier {t}',
    'rw.hint': 'A player is paid once per period. When several rows match, the settings say the first or all.',
    'b.delete': 'Delete board {key}? Its Nakama leaderboard and scores are removed; the period history stays.', 'b.inUse': 'Rewards of this board are not settled yet.',
    'b.players': '{n} players', 'b.ends': 'Period ends {at}', 'b.noEnd': 'No period end', 'b.tiers': '{n} tiers', 'b.rewards': '{n} reward rows',
    'b.restart': 'This change (order, score write or reset period) recreates the Nakama leaderboard. The running period has {n} players; it is closed and archived first.',
    'b.restartPay': 'Pay the running period', 'b.restartGo': 'Close the period and save',
    'st.upcoming': 'Not started', 'st.active': 'Active', 'st.ended': 'Ended', 'st.inactive': 'Inactive',
    'sd.board': 'Board', 'sd.find': 'Username or player id', 'sd.rank': 'Rank', 'sd.player': 'Player', 'sd.score': 'Score', 'sd.tier': 'Tier',
    'sd.edit': 'Change score', 'sd.mode': 'How', 'sd.mode.set': 'Set to this number', 'sd.mode.add': 'Add', 'sd.removeScore': 'Remove the score of {name} in this period?',
    'sd.close': 'Close the period now', 'sd.closePay': 'Pay this period', 'sd.closeConfirm': 'The running period is closed and archived and the board starts from zero.',
    'sd.closed': 'Period closed', 'sd.period': 'Running period: {from} – {to}', 'sd.noBoards': 'No board yet.',
    'wn.end': 'Period end', 'wn.reason': 'Why', 'wn.players': 'Players', 'wn.rewarded': 'Rewarded', 'wn.status': 'Payment', 'wn.retry': 'Retry',
    'wn.retried': '{n} rewards queued again, {p} paid', 'wn.allBoards': 'All boards', 'wn.onlyRewarded': 'Rewarded only', 'wn.error': 'Error',
    'reason.reset': 'Automatic reset', 'reason.manual': 'Closed by hand', 'reason.restart': 'Board changed', 'reason.end': 'Window ended',
    'rs.none': 'No reward', 'rs.pending': 'Queued', 'rs.claimable': 'Waiting for claim', 'rs.paid': 'Paid', 'rs.failed': 'Failed', 'rs.expired': 'Expired',
    'rp.days': 'Range', 'rp.d': 'Last {n} days', 'rp.kpi.boards': 'Active boards', 'rp.kpi.players': 'Players now', 'rp.kpi.writes': 'Scores written', 'rp.kpi.refused': 'Scores refused',
    'rp.kpi.paid': 'Rewards paid', 'rp.kpi.open': 'Open rewards', 'rp.kpi.failed': 'Failed payments', 'rp.writers': 'Active players', 'rp.periods': 'Closed periods', 'rp.daily': 'Scores per day',
    'se.display': 'Display', 'se.topLimit': 'Default rows per board', 'se.maxLimit': 'Most rows per request', 'se.aroundLimit': 'Rows around the player', 'se.showDisplayName': 'Rows carry display name and avatar',
    'se.showUpcoming': 'List boards that have not started', 'se.showEndedDays': 'Days an ended board stays listed',
    'se.rewards': 'Rewards', 'se.payMode': 'Default payment', 'se.claimDays': 'Days to claim a reward', 'se.stack': 'When several reward rows match a player', 'se.stack.first': 'Only the first row', 'se.stack.all': 'Add up every row',
    'se.groupMode': 'Prize groups', 'se.groupMode.draw': 'One drawn by weight', 'se.groupMode.all': 'Every prize of the group', 'se.maxTries': 'Payment attempts', 'se.catchUpDays': 'A period that ended more days ago than this (server down) is archived without rewards',
    'se.scores': 'Score submits', 'se.requireRequestId': 'Game submits need a requestId (a resend never counts twice)', 'se.playerPerMinute': 'Submits per player per minute over all boards', 'se.logDays': 'Days to keep the score log', 'se.clientMaxPerWrite': 'Largest score the game may send to a board without its own cap (0 = no cap)',
    'se.history': 'History', 'se.keepDays': 'Days to keep periods (0 = forever)', 'se.playerLimit': 'Rows of a player\'s history', 'se.winnersLimit': 'Rows of the winners list',
    'se.events': 'Period-end events', 'se.economyEvent': 'Economy event for rewarded players (value = rank)', 'se.analyticsEvent': 'Analytics event (context = board key)', 'se.notify': 'Send the player a notification', 'se.notifyCode': 'Notification code (0 = 2301)',
    'se.noPrizes': 'The economy plugin is not loaded; economy prizes are not paid (currencies are).',
    'pc.title': 'Leaderboards', 'pc.current': 'Rank in the running period', 'pc.history': 'Past periods', 'pc.none': 'No score on any board.', 'pc.refused': '{n} refused scores in the last 30 days',
    'pc.rank': 'Rank {r}', 'pc.score': '{s} points',
    'code.KEY_EXISTS': 'This key is already used.', 'code.INVALID_KEY': 'Invalid key (English letters, digits, _ . : -).', 'code.VALUE_OUT_OF_RANGE': 'The number is out of range.',
    'code.INVALID_POLICY': 'One of the choices is not valid.', 'code.DEFAULT_LANGUAGE_REQUIRED': 'A name in the default language is required.', 'code.INVALID_FIELD': 'Invalid field name.',
    'f.scope': 'Who is ranked', 'scope.player': 'Players', 'scope.clan': 'Clans', 'hint.scope': 'On a clan board each row is a clan and members\' scores add up for their clan.',
    'hint.clanNoRewards': 'Clan boards have no rewards and the game cannot submit scores directly.', 'b.clan': 'Clan',
    'code.LEADERBOARD_CLAN_NO_REWARDS': 'A clan board cannot have rewards.', 'code.LEADERBOARD_CLAN_NO_SUBMIT': 'On a clan board the game cannot submit scores directly.',
    'code.LEADERBOARD_INVALID_SCHEDULE': 'The reset period or CRON is not valid.', 'code.LEADERBOARD_INVALID_WINDOW': 'The end must be after the start.',
    'code.LEADERBOARD_COUNT_NEEDS_INCR': '“One point per event” needs the “Add up” score write.', 'code.LEADERBOARD_TIER_ORDER': 'Tiers must go from the best down.',
    'code.LEADERBOARD_TIER_NOT_FOUND': 'This tier is not on the board.', 'code.LEADERBOARD_INVALID_RANKS': 'Invalid rank range (without a tier, “to rank” is needed).',
    'code.LEADERBOARD_PRIZE_NOT_FOUND': 'The prize was not found.', 'code.LEADERBOARD_EMPTY_REWARD': 'The reward row is empty.', 'code.LEADERBOARD_KEY_IS_TOURNAMENT': 'This id belongs to a tournament.',
    'code.LEADERBOARD_RESTART_REQUIRED': 'This change needs the running period to be closed.', 'code.TOO_MANY_ROWS': 'Too many rows.', 'code.IN_USE': 'Rewards of this board are not settled.',
    'code.USER_NOT_FOUND': 'Player not found.', 'code.INVALID_JSON': 'Invalid data.', 'code.VALUE_TOO_LONG': 'The text is too long.'
  }
};

function lbLang() {
  var nc = window.NakamaConsole;
  var l = nc && nc.lang ? nc.lang() : (document.documentElement.getAttribute('lang') || 'fa');
  return /^en/i.test(l) ? 'en' : 'fa';
}

function has(key) {
  return (LB_TEXT[lbLang()] || LB_TEXT.fa)[key] != null;
}

function t(key, vars) {
  var dict = LB_TEXT[lbLang()] || LB_TEXT.fa;
  var s = dict[key] != null ? dict[key] : (LB_TEXT.en[key] != null ? LB_TEXT.en[key] : key);
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] != null ? digits(String(vars[k])) : m; });
  return s;
}

function digits(s) {
  if (lbLang() !== 'fa') return s;
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
    return d.toLocaleString(lbLang() === 'fa' ? 'fa-IR' : 'en-GB', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

function tr(key, vars, raw) {
  var s = t(key, vars);
  Object.keys(raw || {}).forEach(function (k) { s = s.split('{' + k + '}').join(String(raw[k])); });
  return s;
}

var LB_API = '/v2/console/base';
var LB_INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 px-3 py-2 text-sm';
var LB_BTN = 'lb-btn inline-flex items-center justify-center gap-1.5 font-medium whitespace-nowrap rounded-md border text-sm py-1.5 px-3';

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

function LbError(status, code, message, field, meta) {
  this.meta = meta || {};
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
  return fetch((path.indexOf('/v2/') === 0 ? '' : LB_API) + path, init).then(function (res) {
    if (res.status === 401 && !/^#\/login/.test(location.hash)) location.hash = '#/login';
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) throw new LbError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', data && data.field || '', data && data.meta);
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
  var b = h('button', { type: opts.type || 'button', class: LB_BTN + ' lb-btn--' + (opts.variant || 'ghost') + (opts.class ? ' ' + opts.class : ''), disabled: !!opts.disabled, title: opts.title, 'aria-label': opts.title, onclick: opts.onclick },
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

var lbToastBox = null;
function toast(kind, text) {
  if (!lbToastBox || !lbToastBox.isConnected) {
    lbToastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-lbx': 'toasts' });
    document.body.appendChild(lbToastBox);
  }
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  lbToastBox.appendChild(el);
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
  var err = h('p', { class: 'lb-field-error text-xs text-error-600 dark:text-error-400', hidden: true });
  var wrap = h('label', { class: 'lb-field block mb-4' + (extra ? ' ' + extra : '') },
    label ? h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1', text: label }) : null,
    control,
    hint ? h('span', { class: 'block text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: hint }) : null,
    err);
  wrap.setError = function (msg) { err.textContent = msg || ''; err.hidden = !msg; };
  return wrap;
}

function input(value, opts) {
  opts = opts || {};
  var el = h(opts.multiline ? 'textarea' : 'input', { class: LB_INPUT + (opts.mono ? ' font-mono' : ''), type: opts.multiline ? null : (opts.type || 'text'), dir: opts.ltr ? 'ltr' : 'auto', rows: opts.multiline ? (opts.rows || 3) : null, placeholder: opts.placeholder, name: opts.name, autocomplete: 'off', 'aria-label': opts.label });
  el.value = value == null ? '' : value;
  return el;
}

function select(value, options, opts) {
  opts = opts || {};
  var el = h('select', { class: LB_INPUT + ' eco-select', name: opts.name, 'aria-label': opts.label });
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) opt.selected = true;
    el.appendChild(opt);
  });
  return el;
}

function check(checked, label, name) {
  var box = h('input', { type: 'checkbox', class: 'lb-check', name: name, checked: !!checked });
  var wrap = h('label', { class: 'lb-checkline flex items-start gap-2 mb-3 cursor-pointer text-sm text-secondary-800 dark:text-secondary-200' }, box, h('span', { text: label }));
  wrap.input = box;
  return wrap;
}

function numberValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? Math.round(n) : NaN;
}

var lbPanel = null;
function closePanel() {
  if (lbPanel) { lbPanel.remove(); lbPanel = null; document.removeEventListener('keydown', panelKey); }
}

function panelKey(e) { if (e.key === 'Escape') closePanel(); }

function sidePanel(title, body, footer, name) {
  closePanel();
  var holder = h('div', { class: 'lb-panel-holder' });
  var panel = h('div', { class: 'lb-panel bg-white dark:bg-secondary-900 shadow-lg', role: 'dialog', 'aria-modal': 'true', 'aria-label': title, 'data-lbx-panel': name || '' },
    h('div', { class: 'lb-panel-head flex items-center justify-between border-b px-6 py-3' },
      h('h2', { class: 'text-lg font-semibold truncate', text: title }),
      button({ label: '×', title: t('close'), onclick: closePanel, class: 'lb-close' })),
    h('div', { class: 'lb-panel-body p-6' }, body),
    footer ? h('div', { class: 'lb-panel-foot flex gap-2 justify-end bg-secondary-50 dark:bg-secondary-800 p-4' }, footer) : null);
  holder.appendChild(panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) closePanel(); });
  lbPanel = h('div', { 'data-lbx': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, h('div', { class: 'lb-shade' }), holder);
  document.body.appendChild(lbPanel);
  document.addEventListener('keydown', panelKey);
  return panel;
}

function confirmBox(text, okLabel) {
  return new Promise(function (resolve) {
    var root;
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-lbx': 'confirm', 'data-l10n-skip': '', class: 'lb-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'lb-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('confirm.title') }),
        h('p', { class: 'text-sm mb-4', text: text }),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: okLabel || t('yes'), variant: 'danger', data: { 'lbx-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root);
  });
}

function pageFrame(page, desc, actions) {
  var content = h('div', { class: 'flex flex-col gap-4' });
  var el = h('div', { class: 'lb-page flex flex-col flex-1 px-4 md:px-6 pb-6 max-w-[1600px] w-full mx-auto', 'data-lbx-page': page },
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
    var tr = h('tr', { class: 'border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' + (onrow ? ' cursor-pointer' : ''), 'data-lbx-row': row.id || '' });
    if (onrow) {
      tr.tabIndex = 0;
      tr.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,select,label')) onrow(row.item); });
      tr.addEventListener('keydown', function (e) { if (e.key === 'Enter') onrow(row.item); });
    }
    row.cells.forEach(function (cell, i) { tr.appendChild(h('td', { class: 'px-3 py-2 align-middle text-sm', 'data-label': columns[i].label }, cell)); });
    tbody.appendChild(tr);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow rounded-md overflow-x-auto' }, h('table', { class: 'w-full eco-table lb-table' }, thead, tbody));
}

function pager(state, onpage) {
  var total = state.total || 0;
  var from = total ? (state.page - 1) * state.limit + 1 : 0;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  return h('div', { class: 'flex items-center justify-end gap-2 text-xs text-secondary-500' },
    h('span', { 'data-lbx-range': '', text: total ? t('range', { from: from, to: to, total: total }) : t('empty') }),
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

var lbLanguages = null;
function loadLanguages() {
  if (lbLanguages) return Promise.resolve(lbLanguages);
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch('/v2/console/base/languages', { headers: headers, credentials: 'same-origin' }).then(function (r) { return r.ok ? r.json() : null; }).then(function (d) {
    var list = d && d.items ? d.items.filter(function (l) { return l.isActive; }).map(function (l) { return { code: l.code, name: l.nativeName || l.code, isDefault: !!l.isDefault, direction: l.direction }; }) : [];
    if (!list.length) list = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    list.sort(function (a, b) { return (b.isDefault ? 1 : 0) - (a.isDefault ? 1 : 0); });
    lbLanguages = list;
    return list;
  }, function () {
    lbLanguages = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    return lbLanguages;
  });
}

// translated renders one input per active language; .read() returns {lang: text}.
function translated(label, name, value, opts) {
  opts = opts || {};
  var inputs = {};
  var box = h('div', { class: 'grid sm:grid-cols-2 gap-x-4', 'data-lbx-text': name });
  (lbLanguages || []).forEach(function (l) {
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
  var l = lbLang();
  return m[l] || m.fa || m.en || Object.keys(m).map(function (k) { return m[k]; }).filter(Boolean)[0] || fallback || '';
}

function fieldError(msgBox, err) {
  clear(msgBox).appendChild(notice('error', errorText(err) + (err && err.field ? ' · ' + err.field : '')));
}

function options(list, value) {
  return list.map(function (o) { return typeof o === 'string' ? { value: o, label: o } : o; });
}

function dateInput(value, name, type) {
  var el = h('input', { type: type || 'date', class: LB_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  el.value = value || '';
  return el;
}

function isoDay(d) {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

function lbErrText(err) {
  if (err && err.code && has('code.' + err.code)) return t('code.' + err.code) + (err.field ? ' · ' + err.field : '');
  return errorText(err) + (err && err.field ? ' · ' + err.field : '');
}

function section(title, children, name) {
  return card(h('div', { class: 'p-4', 'data-lbx-section': name || '' },
    h('h2', { class: 'font-semibold mb-3', text: title }), children));
}

// settingsPage loads a plugin's settings once, lets build add its sections and saves everything with one button.
function settingsPage(root, page, base, build) {
  var frame = pageFrame(page, t('desc.' + page));
  clear(root).appendChild(frame.el);
  frame.content.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  Promise.all([loadSettings(base), loadLanguages()]).then(function (res) {
    var d = res[0];
    var cur = JSON.parse(JSON.stringify(d.settings));
    var form = build(cur, d.defaults, d);
    var msg = h('div', { 'data-lbx-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', 'data-lbx-updated': '' });
    var paintMeta = function (x) { meta.textContent = x.updatedBy ? t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }) : ''; };
    paintMeta(d);
    var save = button({ label: t('save'), variant: 'primary', data: { 'lbx-save': page }, onclick: function () {
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
        msg.appendChild(notice('error', lbErrText(err)));
      });
    } });
    clear(frame.content);
    form.el.forEach(function (el) { frame.content.appendChild(el); });
    frame.content.appendChild(h('div', { class: 'lb-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), save));
  }, function (err) {
    clear(frame.content).appendChild(notice('error', lbErrText(err)));
  });
}

// numInput accepts Persian and Latin digits (type=number refuses Persian ones).
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
  return h('a', { href: '#/players/' + id, class: 'eco-link', 'data-lbx-player-link': id, dir: 'auto', text: name || id.slice(0, 8) });
}

// Shared editor parts: economy key lists, currency rows, prize pickers, tags.
var lbRefs = null;
function loadRefs(force) {
  if (lbRefs && !force) return Promise.resolve(lbRefs);
  return api('GET', '/leaderboardx/refs' + qs({ lang: lbLang() })).then(function (r) { lbRefs = r || {}; return lbRefs; }, function () { lbRefs = {}; return lbRefs; });
}

function refOptions(kind, value, emptyLabel) {
  var list = (lbRefs && lbRefs[kind]) || [];
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

// currencyRows edits {key: amount}; .read() returns the map.
function currencyRows(values, name) {
  var box = h('div', { class: 'lb-rows', 'data-lbx-currencies': name || '' });
  var rows = [];
  function addRow(k, v) {
    var key = refSelect('currencies', k || '', 'reward-currency', '—');
    var amt = numInput(v == null ? '' : v, 'reward-amount', t('rw.amount'));
    var row = h('div', { class: 'lb-row' }, field(t('rw.currency'), key), field(t('rw.amount'), amt),
      h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r.row !== row; }); box.dispatchEvent(new Event('input', { bubbles: true })); } })));
    rows.push({ row: row, key: key, amt: amt });
    box.appendChild(row);
  }
  Object.keys(values || {}).forEach(function (k) { addRow(k, values[k]); });
  var wrap = h('div', {}, box, button({ label: '+ ' + t('rw.addCurrency'), data: { 'lbx-action': 'add-currency' }, onclick: function () { addRow(); } }));
  wrap.read = function () {
    var out = {};
    rows.forEach(function (r) { var k = r.key.value; var n = numberValue(r.amt); if (k && n) out[k] = n; });
    return out;
  };
  return wrap;
}

// prizePickers edits {prizeKey, prizeGroupKey}.
function prizePickers(value) {
  value = value || {};
  var prize = refSelect('prizes', value.prizeKey, 'reward-prize', t('rw.noPrize'));
  var group = refSelect('prizeGroups', value.prizeGroupKey, 'reward-prize-group', t('rw.noPrize'));
  var el = h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('rw.prize'), prize), field(t('rw.group'), group));
  el.read = function () { var o = {}; if (prize.value) o.prizeKey = prize.value; if (group.value) o.prizeGroupKey = group.value; return o; };
  return el;
}

function rewardText(cur, prizes) {
  var parts = Object.keys(cur || {}).map(function (k) { return fmtNum(cur[k]) + ' ' + k; });
  (prizes || []).forEach(function (p) { if (p) parts.push(p); });
  if (!parts.length) return t('none');
  return '⁦' + parts.join(' + ') + '⁩';
}

// rewardsText describes a result's rewards {currencies:[{currencyKey, amount}], prizes:[{prizeKey, prizeGroupKey}]}.
function rewardsText(r) {
  r = r || {};
  var cur = {};
  (r.currencies || []).forEach(function (c) { cur[c.currencyKey] = c.amount; });
  return rewardText(cur, (r.prizes || []).map(function (p) { return p.prizeKey || p.prizeGroupKey; }));
}

function statusTag(st) {
  return tag(t('st.' + st), st === 'active' ? 'success' : st === 'upcoming' ? 'primary' : st === 'ended' ? 'warning' : 'secondary');
}

function resultTag(st) {
  var color = { paid: 'success', pending: 'primary', claimable: 'warning', failed: 'error', expired: 'secondary', none: 'secondary' }[st] || 'secondary';
  var el = tag(t('rs.' + st), color);
  el.setAttribute('data-lbx-status', st);
  return el;
}

function tierName(b, key) {
  if (!key) return t('none');
  var tier = (b && b.tiers || []).filter(function (x) { return x.key === key; })[0];
  return tier ? textOf(tier.name, tier.key) : key;
}

function tierSentence(basis, th, prev) {
  if (basis === 'score') return t('tier.s.score', { v: fmtNum(th) });
  if (basis === 'rank') return t('tier.s.rank', { from: prev ? prev + 1 : 1, v: th });
  if (basis === 'percent') return t('tier.s.percent', { v: th });
  return '';
}

function rewardSentence(r, b) {
  var tn = r.tier ? tierName(b, r.tier) : '';
  if (r.tier && !r.toRank) return tr('rw.s.tierAll', null, { t: tn });
  if (r.tier) return tr('rw.s.tier', { from: r.fromRank || 1, to: r.toRank }, { t: tn });
  if ((r.fromRank || 1) === r.toRank) return t('rw.s.one', { from: r.toRank });
  return t('rw.s.rank', { from: r.fromRank || 1, to: r.toRank });
}

function localInput(iso, name) {
  var el = h('input', { type: 'datetime-local', class: LB_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  if (iso) {
    var d = new Date(iso);
    if (!isNaN(d)) el.value = new Date(d.getTime() - d.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
  }
  el.read = function () { return el.value ? new Date(el.value).toISOString() : null; };
  return el;
}

function boardsPage(root) {
  var frame = pageFrame('boards', t('desc.boards'), [button({ label: t('new'), variant: 'primary', data: { 'lbx-action': 'new-board' }, onclick: function () { boardEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  function reload() {
    clear(frame.content).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/leaderboardx/boards'), loadLanguages(), loadRefs(true)]).then(function (res) {
      var list = res[0].items || [];
      clear(frame.content);
      if (!list.length) { frame.content.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      frame.content.appendChild(h('div', { 'data-lbx-list': 'boards' }, table([
        { label: t('f.name') }, { label: t('f.key') }, { label: t('f.period') }, { label: t('f.source') }, { label: t('sec.tiers') }, { label: t('sec.rewards') }, { label: t('wn.players') }, { label: '' }
      ], list.map(function (x) {
        var b = x.board;
        var src = t('src.' + b.source.type) + (b.source.event ? ' · ' + b.source.event : '');
        return { id: b.key, item: x, cells: [
          h('div', { class: 'flex flex-wrap items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(b.name, b.key) }), statusTag(x.status), b.scope === 'clan' ? h('span', { class: 'eco-tag eco-tag--primary inline-flex leading-tight px-2 py-0.5 text-xs rounded-md whitespace-nowrap', 'data-lbx-scope': b.key, text: t('b.clan') }) : null),
          mono(b.key),
          h('div', { class: 'flex flex-col text-xs' }, h('span', { text: t('period.' + b.period) }), h('span', { class: 'text-secondary-500', text: x.periodEnd ? t('b.ends', { at: fmtDate(x.periodEnd) }) : t('b.noEnd') })),
          h('span', { class: 'text-xs', 'data-lbx-source': b.key, text: src }),
          h('span', { class: 'text-xs', text: b.tierBasis === 'none' ? t('basis.none') : t('basis.' + b.tierBasis) + ' · ' + t('b.tiers', { n: b.tiers.length }) }),
          h('span', { class: 'text-xs', 'data-lbx-rewards': b.key, text: b.rewards.length ? b.rewards.map(function (r) { return rewardSentence(r, b); }).join(' · ') : t('none') }),
          h('a', { href: '#/leaderboardx/standings?board=' + encodeURIComponent(b.key), class: 'eco-link', 'data-lbx-players': b.key, text: t('b.players', { n: x.players }) }),
          button({ label: t('delete'), variant: 'danger', data: { 'lbx-delete': b.key }, onclick: function () {
            confirmBox(tr('b.delete', null, { key: b.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/leaderboardx/boards/' + b.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', err.code === 'IN_USE' ? t('b.inUse') : lbErrText(err)); });
            });
          } })
        ] };
      }), function (x) { boardEditor(x.board, reload); })));
    }, function (err) { clear(frame.content).appendChild(notice('error', lbErrText(err))); });
  }
  reload();
}

function tierRows(b, basisSel) {
  var box = h('div', { class: 'flex flex-col gap-3', 'data-lbx-tiers': '' });
  var rows = [];
  function paint() {
    var basis = basisSel.value;
    rows.forEach(function (r, i) {
      r.label.textContent = t('tier.threshold.' + (basis === 'none' ? 'score' : basis));
      var prev = i > 0 ? numberValue(rows[i - 1].th) : 0;
      r.sentence.textContent = tierSentence(basis, floatValue(r.th) || 0, prev);
    });
    wrap.hidden = basis === 'none';
  }
  function addRow(tier) {
    tier = tier || { key: '', name: {}, threshold: '' };
    var key = input(tier.key, { name: 'tier-key', ltr: true, mono: true });
    var th = numInput(tier.threshold, 'tier-threshold');
    var icon = input(tier.iconUrl, { name: 'tier-icon', ltr: true });
    var name = translated(t('f.name'), 'tier-name', tier.name);
    var label = h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1' });
    var sentence = h('div', { class: 'text-xs text-secondary-500 lb-sentence', 'data-lbx-tier-sentence': '' });
    var el = h('div', { class: 'lb-track', 'data-lbx-tier': '' },
      h('div', { class: 'grid sm:grid-cols-3 gap-x-4 items-end' }, field(t('tier.key'), key), h('label', { class: 'lb-field block mb-4' }, label, th), field(t('f.icon'), icon)),
      name, h('div', { class: 'flex items-center justify-between gap-2' }, sentence,
        button({ label: t('remove'), variant: 'danger', onclick: function () { el.remove(); rows = rows.filter(function (r) { return r.el !== el; }); paint(); } })));
    var row = { el: el, th: th, label: label, sentence: sentence, read: function () { return { key: key.value.trim(), name: name.read(), iconUrl: icon.value.trim(), threshold: floatValue(th) || 0 }; } };
    th.addEventListener('input', paint);
    rows.push(row);
    box.appendChild(el);
    paint();
  }
  var wrap = h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('tier.hint') }), box, button({ label: '+ ' + t('tier.add'), data: { 'lbx-action': 'add-tier' }, onclick: function () { addRow(); } }));
  (b.tiers || []).forEach(addRow);
  basisSel.addEventListener('change', paint);
  paint();
  wrap.read = function () { return basisSel.value === 'none' ? [] : rows.map(function (r) { return r.read(); }); };
  wrap.keys = function () { return rows.map(function (r) { var o = r.read(); return { key: o.key, name: o.name }; }).filter(function (x) { return x.key; }); };
  return wrap;
}

function rewardRows(b, tiers) {
  var box = h('div', { class: 'flex flex-col gap-3', 'data-lbx-reward-rows': '' });
  var rows = [];
  function tierOptions(value) {
    var opts = [{ value: '', label: t('rw.allRanks') }];
    tiers.keys().forEach(function (x) { opts.push({ value: x.key, label: tr('rw.inTier', null, { t: textOf(x.name, x.key) }) }); });
    if (value && !opts.some(function (o) { return o.value === value; })) opts.push({ value: value, label: value });
    return opts;
  }
  function addRow(r) {
    r = r || { fromRank: 1, toRank: 1 };
    var tierSel = select(r.tier || '', tierOptions(r.tier), { name: 'reward-tier', label: t('rw.tier') });
    tierSel.addEventListener('focus', function () {
      var v = tierSel.value;
      clear(tierSel);
      tierOptions(v).forEach(function (o) { var opt = h('option', { value: o.value, text: o.label }); if (o.value === v) opt.selected = true; tierSel.appendChild(opt); });
    });
    var from = numInput(r.fromRank || 1, 'reward-from', t('rw.from'));
    var to = numInput(r.toRank || '', 'reward-to', t('rw.to'));
    var cur = currencyRows(r.currencies, 'reward');
    var prizes = prizePickers(r);
    var sentence = h('div', { class: 'text-xs text-secondary-500 lb-sentence', 'data-lbx-reward-sentence': '' });
    function paint() { sentence.textContent = rewardSentence({ tier: tierSel.value, fromRank: numberValue(from) || 1, toRank: numberValue(to) || 0 }, { tiers: tiers.keys() }); }
    [tierSel, from, to].forEach(function (el) { el.addEventListener('input', paint); el.addEventListener('change', paint); });
    var el = h('div', { class: 'lb-track', 'data-lbx-reward': '' },
      h('div', { class: 'grid sm:grid-cols-3 gap-x-4 items-end' }, field(t('rw.tier'), tierSel), field(t('rw.from'), from), field(t('rw.to'), to)),
      cur, prizes, h('div', { class: 'flex items-center justify-between gap-2 mt-2' }, sentence,
        button({ label: t('remove'), variant: 'danger', onclick: function () { el.remove(); rows = rows.filter(function (x) { return x.el !== el; }); } })));
    rows.push({ el: el, read: function () { var o = prizes.read(); o.tier = tierSel.value; o.fromRank = numberValue(from) || 1; o.toRank = numberValue(to) || 0; o.currencies = cur.read(); return o; } });
    box.appendChild(el);
    paint();
  }
  (b.rewards || []).forEach(addRow);
  var wrap = h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('rw.hint') }), box, button({ label: '+ ' + t('rw.add'), data: { 'lbx-action': 'add-reward' }, onclick: function () { addRow(); } }));
  wrap.read = function () { return rows.map(function (r) { return r.read(); }); };
  return wrap;
}

function boardEditor(b, done) {
  var creating = !b;
  b = b || { name: {}, description: {}, isActive: true, sortOrder: 'desc', operator: 'incr', period: 'weekly', resetSchedule: '', tierBasis: 'none', tiers: [], rewards: [],
    source: { type: 'economy_event', calc: 'value', multiplier: 1 }, limits: { onExceed: 'reject' }, payMode: '', metadata: {} };
  var opt = function (prefix, list) { return list.map(function (v) { return { value: v, label: t(prefix + v) }; }); };
  var key = input(b.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = translated(t('f.name'), 'name', b.name);
  var desc = translated(t('f.description'), 'description', b.description, { multiline: true });
  var icon = input(b.iconUrl, { name: 'iconUrl', ltr: true });
  var sort = numInput(b.sort || 0, 'sort');
  var active = check(b.isActive, t('f.active'), 'isActive');
  var scope = select(b.scope || 'player', opt('scope.', ['player', 'clan']), { name: 'scope' });
  var clanNote = h('p', { class: 'text-xs text-warning-700 dark:text-warning-400 mb-3', 'data-lbx-clan-note': '', text: t('hint.clanNoRewards') });
  var period = select(b.period, opt('period.', ['none', 'daily', 'weekly', 'monthly', 'custom']), { name: 'period' });
  var cron = input(b.period === 'custom' ? b.resetSchedule : '', { name: 'resetSchedule', ltr: true, mono: true, placeholder: '0 0 * * 5' });
  var cronField = field(t('f.cron'), cron, t('hint.cron'));
  var starts = localInput(b.startsAt, 'startsAt');
  var ends = localInput(b.endsAt, 'endsAt');
  var maxPeriods = numInput(b.maxPeriods || 0, 'maxPeriods');
  var sortOrder = select(b.sortOrder, opt('sort.', ['desc', 'asc']), { name: 'sortOrder' });
  var operator = select(b.operator, opt('op.', ['incr', 'best', 'set', 'decr']), { name: 'operator' });
  var srcType = select(b.source.type, opt('src.', ['economy_event', 'game_event', 'client', 'server']), { name: 'source.type' });
  var event = input(b.source.event, { name: 'source.event', ltr: true, mono: true });
  var events = h('datalist', { id: 'lbx-events' });
  ((lbRefs && lbRefs.economyEvents) || []).forEach(function (e) { events.appendChild(h('option', { value: e.key, text: e.name })); });
  event.setAttribute('list', 'lbx-events');
  var context = input(b.source.context, { name: 'source.context', ltr: true, mono: true });
  var fieldIn = input(b.source.field, { name: 'source.field', ltr: true, mono: true, placeholder: 'stars' });
  var calc = select(b.source.calc, opt('calc.', ['value', 'count']), { name: 'source.calc' });
  var mult = input(b.source.multiplier == null ? 1 : digits(String(b.source.multiplier)), { name: 'source.multiplier', ltr: true });
  var eventBox = h('div', {}, field(t('f.event'), event), events);
  var gameBox = h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.context'), context), field(t('f.field'), fieldIn));
  var clientSubmit = check(b.limits.clientSubmit, t('f.clientSubmit'), 'limits.clientSubmit');
  var maxPerWrite = numInput(b.limits.maxPerWrite || 0, 'limits.maxPerWrite');
  var perMinute = numInput(b.limits.perMinute || 0, 'limits.perMinute');
  var onExceed = select(b.limits.onExceed || 'reject', opt('exceed.', ['reject', 'clamp']), { name: 'limits.onExceed' });
  var basis = select(b.tierBasis, opt('basis.', ['none', 'score', 'rank', 'percent']), { name: 'tierBasis' });
  var tiers = tierRows(b, basis);
  var rewards = rewardRows(b, tiers);
  var minScore = numInput(b.rewardMinScore || 0, 'rewardMinScore');
  var payMode = select(b.payMode || '', opt('pay.', ['', 'auto', 'claim']), { name: 'payMode' });
  var clientOpt = srcType.querySelector('option[value="client"]');
  var rewardsSec = null;
  function paint() {
    var clan = scope.value === 'clan';
    if (clan && srcType.value === 'client') srcType.value = 'economy_event';
    if (clientOpt) clientOpt.disabled = clan;
    cronField.hidden = period.value !== 'custom';
    eventBox.hidden = srcType.value !== 'economy_event' && srcType.value !== 'game_event';
    gameBox.hidden = srcType.value !== 'game_event';
    if (clan) { clientSubmit.input.checked = false; clientSubmit.input.disabled = true; }
    else if (srcType.value === 'client') { clientSubmit.input.checked = true; clientSubmit.input.disabled = true; } else clientSubmit.input.disabled = false;
    clanNote.hidden = !clan;
    if (rewardsSec) rewardsSec.hidden = clan;
  }
  [period, srcType, scope].forEach(function (el) { el.addEventListener('change', paint); });
  paint();
  var msg = h('div', { 'data-lbx-msg': '' });
  var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
  rewardsSec = section(t('sec.rewards'), h('div', {}, rewards, h('div', { class: 'h-3' }), grid(field(t('f.rewardMinScore'), minScore), field(t('f.payMode'), payMode))), 'rewards');
  paint();
  var body = h('div', { 'data-lbx-form': 'board' }, field(t('f.key'), key, t('hint.key')), name, desc, grid(field(t('f.icon'), icon), field(t('f.sort'), sort)), active,
    field(t('f.scope'), scope, t('hint.scope')), clanNote,
    section(t('sec.period'), h('div', {}, field(t('f.period'), period), cronField, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.window') }),
      grid(field(t('f.startsAt'), starts), field(t('f.endsAt'), ends)), field(t('f.maxPeriods'), maxPeriods, null, 'lb-num')), 'period'), h('div', { class: 'h-3' }),
    section(t('sec.scoring'), h('div', {}, grid(field(t('f.sortOrder'), sortOrder), field(t('f.operator'), operator)), field(t('f.source'), srcType), eventBox, gameBox,
      grid(field(t('f.calc'), calc), field(t('f.multiplier'), mult, t('hint.multiplier')))), 'scoring'), h('div', { class: 'h-3' }),
    section(t('sec.limits'), h('div', {}, clientSubmit, grid(field(t('f.maxPerWrite'), maxPerWrite), field(t('f.perMinute'), perMinute)), field(t('f.onExceed'), onExceed)), 'limits'), h('div', { class: 'h-3' }),
    section(t('sec.tiers'), h('div', {}, field(t('f.tierBasis'), basis), tiers), 'tiers'), h('div', { class: 'h-3' }),
    rewardsSec, msg);
  function read() {
    var clan = scope.value === 'clan';
    return { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), sort: numberValue(sort) || 0, isActive: active.input.checked, scope: scope.value,
      period: period.value, resetSchedule: period.value === 'custom' ? cron.value.trim() : '', startsAt: starts.read(), endsAt: ends.read(), maxPeriods: numberValue(maxPeriods) || 0,
      sortOrder: sortOrder.value, operator: operator.value,
      source: { type: srcType.value, event: event.value.trim(), context: context.value.trim(), field: fieldIn.value.trim(), calc: calc.value, multiplier: floatValue(mult) || 1 },
      limits: { clientSubmit: !clan && clientSubmit.input.checked, maxPerWrite: numberValue(maxPerWrite) || 0, perMinute: numberValue(perMinute) || 0, onExceed: onExceed.value },
      tierBasis: basis.value, tiers: tiers.read(), rewards: clan ? [] : rewards.read(), rewardMinScore: numberValue(minScore) || 0, payMode: payMode.value, metadata: b.metadata || {} };
  }
  function send(extra) {
    busy(save, true);
    clear(msg);
    var out = read();
    Object.keys(extra || {}).forEach(function (k) { out[k] = extra[k]; });
    api(creating ? 'POST' : 'PUT', '/leaderboardx/boards' + (creating ? '' : '/' + b.id), out).then(function () {
      closePanel(); toast('success', t('saved')); done();
    }, function (err) {
      busy(save, false);
      if (err.code === 'LEADERBOARD_RESTART_REQUIRED') {
        var pay = check(false, t('b.restartPay'), 'payOnRestart');
        msg.appendChild(h('div', { class: 'eco-notice eco-notice--warning rounded-md px-3 py-2 text-sm flex flex-col gap-2', 'data-lbx-restart': '' },
          h('span', { text: t('b.restart', { n: (err.meta && err.meta.players) || 0 }) }), pay,
          h('div', {}, button({ label: t('b.restartGo'), variant: 'primary', data: { 'lbx-action': 'restart' }, onclick: function () { send({ restart: true, payOnRestart: pay.input.checked }); } }))));
        return;
      }
      msg.appendChild(notice('error', lbErrText(err)));
    });
  }
  var save = button({ label: t('save'), variant: 'primary', data: { 'lbx-action': 'save-board' }, onclick: function () { send(); } });
  sidePanel(creating ? t('new') : textOf(b.name, b.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'board');
}

// boardPicker loads the boards and calls onpick with the chosen one (kept in the hash as ?board=key).
function boardPicker(list, current, onpick, allLabel) {
  var opts = allLabel ? [{ value: '', label: allLabel }] : [];
  list.forEach(function (x) { opts.push({ value: x.board.key, label: textOf(x.board.name, x.board.key) + ' · ' + x.board.key }); });
  var sel = select(current || '', opts, { name: 'board', label: t('sd.board') });
  sel.setAttribute('dir', 'auto');
  sel.addEventListener('change', function () { onpick(sel.value); });
  return sel;
}

function standingsPage(root) {
  var q = hashQuery();
  var state = { board: q.board || '', find: '', cursor: '', stack: [] };
  var frame = pageFrame('standings', t('desc.standings'));
  clear(root).appendChild(frame.el);
  var head = h('div', { class: 'flex flex-wrap gap-2 items-end lb-filters' });
  var body = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(head);
  frame.content.appendChild(body);
  var boards = [];
  function current() { return boards.filter(function (x) { return x.board.key === state.board; })[0]; }
  api('GET', '/leaderboardx/boards').then(function (d) {
    boards = d.items || [];
    if (!boards.length) { body.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('sd.noBoards') }))); return; }
    if (!current()) state.board = boards[0].board.key;
    var find = input('', { name: 'find', ltr: true, placeholder: t('sd.find') });
    find.addEventListener('keydown', function (e) { if (e.key === 'Enter') { state.find = find.value.trim(); state.cursor = ''; state.stack = []; load(); } });
    var closeBtn = button({ label: t('sd.close'), variant: 'danger', data: { 'lbx-action': 'close-period' }, onclick: closePeriod });
    head.appendChild(field(t('sd.board'), boardPicker(boards, state.board, function (v) { state.board = v; state.cursor = ''; state.stack = []; load(); })));
    head.appendChild(field(t('search'), find));
    head.appendChild(h('div', { class: 'mb-4' }, closeBtn));
    load();
  }, function (err) { body.appendChild(notice('error', lbErrText(err))); });
  function closePeriod() {
    var x = current();
    if (!x) return;
    var pay = check(true, t('sd.closePay'), 'closePay');
    var root2;
    function done(ok) {
      root2.remove();
      if (!ok) return;
      api('POST', '/leaderboardx/boards/' + x.board.id + '/close', { pay: pay.input.checked }).then(function () { toast('success', t('sd.closed')); load(); }, function (err) { toast('error', lbErrText(err)); });
    }
    root2 = h('div', { 'data-lbx': 'confirm', 'data-l10n-skip': '', class: 'lb-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'lb-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('sd.close') }), h('p', { class: 'text-sm mb-3', text: t('sd.closeConfirm') }), pay,
        h('div', { class: 'flex gap-2 justify-end' }, button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: t('sd.close'), variant: 'danger', data: { 'lbx-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root2);
  }
  function load() {
    var x = current();
    if (!x) return;
    clear(body).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', '/leaderboardx/boards/' + x.board.id + '/standings' + qs({ limit: 50, cursor: state.cursor, q: state.find })).then(function (d) {
      clear(body);
      var b = d.board;
      body.appendChild(h('div', { class: 'flex flex-wrap items-center gap-3 text-sm', 'data-lbx-standing-head': b.key },
        statusTag(d.status), h('span', { 'data-lbx-total': '', text: t('b.players', { n: d.players }) }),
        h('span', { class: 'text-secondary-500', text: d.periodEnd ? t('sd.period', { from: fmtDate(d.periodStart), to: fmtDate(d.periodEnd) }) : t('b.noEnd') })));
      var items = d.items || [];
      if (!items.length) { body.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      body.appendChild(h('div', { 'data-lbx-list': 'standings' }, table([{ label: t('sd.rank') }, { label: t('sd.player') }, { label: t('sd.score') }, { label: t('sd.tier') }, { label: '' }],
        items.map(function (r) {
          return { id: r.userId, item: r, cells: [
            h('span', { class: 'font-semibold', 'data-lbx-rank': r.userId, text: fmtNum(r.rank) }),
            playerLink(r.userId, r.displayName || r.username),
            h('span', { 'data-lbx-score': r.userId, text: fmtNum(r.score) + (r.subscore ? ' · ' + fmtNum(r.subscore) : '') }),
            r.tier ? tag(tierName(b, r.tier), 'primary') : h('span', { text: t('none') }),
            h('div', { class: 'flex gap-1' }, button({ label: t('sd.edit'), data: { 'lbx-edit': r.userId }, onclick: function () { scoreEditor(b, r, load); } }),
              button({ label: t('remove'), variant: 'danger', data: { 'lbx-remove': r.userId }, onclick: function () {
                confirmBox(tr('sd.removeScore', null, { name: r.username || r.userId }), t('remove')).then(function (ok) {
                  if (!ok) return;
                  api('DELETE', '/leaderboardx/boards/' + b.id + '/records/' + r.userId).then(function () { toast('success', t('saved')); load(); }, function (err) { toast('error', lbErrText(err)); });
                });
              } }))
          ] };
        }))));
      if (!state.find) {
        body.appendChild(h('div', { class: 'flex justify-end gap-2' },
          button({ label: t('prev'), disabled: !state.stack.length, onclick: function () { state.cursor = state.stack.pop() || ''; load(); } }),
          button({ label: t('next'), disabled: !d.nextCursor, onclick: function () { state.stack.push(state.cursor); state.cursor = d.nextCursor; load(); } })));
      }
    }, function (err) { clear(body).appendChild(notice('error', lbErrText(err))); });
  }
}

function scoreEditor(b, r, done) {
  var score = numInput(r.score, 'score', t('sd.score'));
  var mode = select('set', [{ value: 'set', label: t('sd.mode.set') }, { value: 'add', label: t('sd.mode.add') }], { name: 'mode' });
  var msg = h('div', {});
  var save = button({ label: t('save'), variant: 'primary', data: { 'lbx-action': 'save-score' }, onclick: function () {
    busy(save, true);
    api('PUT', '/leaderboardx/boards/' + b.id + '/records/' + r.userId, { score: numberValue(score) || 0, mode: mode.value }).then(function () {
      closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); clear(msg).appendChild(notice('error', lbErrText(err))); });
  } });
  sidePanel(t('sd.edit') + ' · ' + (r.username || r.userId), h('div', { 'data-lbx-form': 'score' }, field(t('sd.score'), score), field(t('sd.mode'), mode), msg), [button({ label: t('cancel'), onclick: closePanel }), save], 'score');
}

function winnersPage(root) {
  var q = hashQuery();
  var state = { board: q.board || '', page: 1, limit: 20 };
  var frame = pageFrame('winners', t('desc.winners'));
  clear(root).appendChild(frame.el);
  var head = h('div', { class: 'flex flex-wrap gap-2 items-end lb-filters' });
  var body = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(head);
  frame.content.appendChild(body);
  var boards = [];
  Promise.all([api('GET', '/leaderboardx/boards'), loadLanguages()]).then(function (res) {
    boards = res[0].items || [];
    head.appendChild(field(t('sd.board'), boardPicker(boards, state.board, function (v) { state.board = v; state.page = 1; load(); }, t('wn.allBoards'))));
    load();
  }, function (err) { body.appendChild(notice('error', lbErrText(err))); });
  function counts(c) {
    var box = h('div', { class: 'flex flex-wrap gap-1' });
    ['paid', 'pending', 'claimable', 'failed', 'expired'].forEach(function (k) {
      if (c[k]) { var el = resultTag(k); el.textContent = t('rs.' + k) + ' ' + digits(String(c[k])); box.appendChild(el); }
    });
    return box.childNodes.length ? box : h('span', { text: t('none') });
  }
  function load() {
    clear(body).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', '/leaderboardx/periods' + qs({ boardKey: state.board, page: state.page, limit: state.limit })).then(function (d) {
      clear(body);
      var items = d.items || [];
      if (!items.length) { body.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      body.appendChild(h('div', { 'data-lbx-list': 'periods' }, table([
        { label: t('sd.board') }, { label: t('wn.end') }, { label: t('wn.reason') }, { label: t('wn.players') }, { label: t('wn.rewarded') }, { label: t('wn.status') }, { label: '' }
      ], items.map(function (p) {
        return { id: p.id, item: p, cells: [
          h('div', { class: 'flex flex-col' }, h('span', { class: 'font-medium', text: textOf(p.boardName, p.boardKey) }), mono(p.boardKey)),
          h('span', { class: 'text-xs', text: fmtDate(p.periodEnd) }),
          tag(t('reason.' + p.reason), p.reason === 'reset' ? 'primary' : 'secondary'),
          h('span', { 'data-lbx-period-players': p.id, text: fmtNum(p.players) }),
          h('span', { text: fmtNum(p.rewarded) }),
          h('div', { 'data-lbx-period-counts': p.id }, counts(p.counts || {})),
          p.counts && p.counts.failed ? button({ label: t('wn.retry'), variant: 'primary', data: { 'lbx-retry': p.id }, onclick: function () {
            api('POST', '/leaderboardx/periods/' + p.id + '/retry', {}).then(function (out) { toast('success', t('wn.retried', { n: out.requeued, p: out.paid })); load(); }, function (err) { toast('error', lbErrText(err)); });
          } }) : null
        ] };
      }), function (p) { periodPanel(p, boards); })));
      body.appendChild(pager({ page: state.page, limit: state.limit, total: d.total }, function (n) { state.page = n; load(); }));
    }, function (err) { clear(body).appendChild(notice('error', lbErrText(err))); });
  }
}

function periodPanel(p, boards) {
  var x = boards.filter(function (b) { return b.board.key === p.boardKey; })[0];
  var b = x ? x.board : { tiers: [] };
  var only = check(false, t('wn.onlyRewarded'), 'onlyRewarded');
  var list = h('div', { 'data-lbx-results': p.id });
  var state = { page: 1, limit: 50 };
  function load() {
    clear(list).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', '/leaderboardx/periods/' + p.id + qs({ status: only.input.checked ? 'rewarded' : '', page: state.page, limit: state.limit })).then(function (d) {
      clear(list);
      var items = d.items || [];
      if (!items.length) { list.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('empty') })); return; }
      var ol = h('ol', { class: 'flex flex-col gap-2' });
      items.forEach(function (r) {
        ol.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-sm border-b border-secondary-200 dark:border-secondary-700 pb-2', 'data-lbx-result': r.userId },
          h('span', { class: 'font-semibold min-w-[2rem]', text: fmtNum(r.rank) }), playerLink(r.userId, r.username),
          h('span', { class: 'text-xs', text: fmtNum(r.score) }), r.tier ? tag(tierName(b, r.tier), 'primary') : null,
          h('span', { class: 'text-xs', text: rewardsText(r.rewards) }), resultTag(r.status),
          r.lastError ? h('span', { class: 'text-xs text-error-600', text: t('wn.error') + ': ' + r.lastError }) : null));
      });
      list.appendChild(ol);
      list.appendChild(pager({ page: state.page, limit: state.limit, total: d.total }, function (n) { state.page = n; load(); }));
    }, function (err) { clear(list).appendChild(notice('error', lbErrText(err))); });
  }
  only.input.addEventListener('change', function () { state.page = 1; load(); });
  sidePanel(textOf(p.boardName, p.boardKey) + ' · ' + fmtDate(p.periodEnd), h('div', {}, only, list), [button({ label: t('close'), onclick: closePanel })], 'period');
  load();
}

function reportPage(root) {
  var q = hashQuery();
  var state = { days: Number(q.days) || 30 };
  var frame = pageFrame('report', t('desc.report'));
  clear(root).appendChild(frame.el);
  var daysSel = select(String(state.days), [7, 30, 90].map(function (n) { return { value: String(n), label: t('rp.d', { n: n }) }; }), { name: 'days', label: t('rp.days') });
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end lb-filters' }, field(t('rp.days'), daysSel)));
  frame.content.appendChild(body);
  daysSel.addEventListener('change', function () { state.days = Number(daysSel.value); reload(); });
  function kpi(name, value) {
    return card(h('div', { class: 'p-4', 'data-lbx-kpi': name }, h('div', { class: 'text-xs text-secondary-500', text: t('rp.kpi.' + name) }), h('div', { class: 'text-xl font-semibold mt-1', text: fmtNum(value || 0) })));
  }
  function reload() {
    clear(body).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', '/leaderboardx/report' + qs({ days: state.days })).then(function (d) {
      clear(body);
      var tot = d.totals || {};
      body.appendChild(h('div', { class: 'grid grid-cols-2 md:grid-cols-4 gap-3' }, kpi('boards', tot.boards), kpi('players', tot.players), kpi('writes', tot.writes), kpi('refused', tot.refused),
        kpi('paid', tot.paid), kpi('open', tot.open), kpi('failed', tot.failed)));
      var daily = d.daily || [];
      if (daily.length) {
        var peak = Math.max.apply(null, daily.map(function (x) { return x.writes + x.refused; }).concat([1]));
        var bars = h('ol', { class: 'flex flex-col gap-1', 'data-lbx-daily': '' });
        daily.forEach(function (x) {
          bars.appendChild(h('li', { class: 'flex items-center gap-2 text-xs' }, h('span', { class: 'w-24 shrink-0', dir: 'ltr', text: x.day }),
            h('span', { class: 'lb-bar flex-1' }, h('span', { style: 'width:' + Math.round((x.writes / peak) * 100) + '%' })),
            h('span', { class: 'w-20 text-end', text: fmtNum(x.writes) + (x.refused ? ' / ' + fmtNum(x.refused) : '') })));
        });
        body.appendChild(section(t('rp.daily'), bars, 'daily'));
      }
      var items = d.items || [];
      if (!items.length) { body.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      body.appendChild(h('div', { 'data-lbx-list': 'report' }, table([
        { label: t('sd.board') }, { label: t('rp.kpi.players') }, { label: t('rp.kpi.writes') }, { label: t('rp.kpi.refused') }, { label: t('rp.writers') }, { label: t('rp.periods') }, { label: t('wn.status') }
      ], items.map(function (r) {
        var rw = r.rewards || {};
        return { id: r.key, item: r, cells: [
          h('div', { class: 'flex flex-col' }, h('span', { class: 'font-medium', text: textOf(r.name, r.key) }), mono(r.key)),
          h('span', { 'data-lbx-report-players': r.key, text: fmtNum(r.players) }),
          h('span', { 'data-lbx-report-writes': r.key, text: fmtNum(r.writes) }),
          h('span', { 'data-lbx-report-refused': r.key, text: fmtNum(r.refused) }),
          h('span', { text: fmtNum(r.writers) }), h('span', { text: fmtNum(r.periods) }),
          h('span', { class: 'text-xs', 'data-lbx-report-rewards': r.key, text: t('rs.paid') + ' ' + fmtNum(rw.paid) + ' · ' + t('rs.claimable') + ' ' + fmtNum(rw.claimable) + ' · ' + t('rs.failed') + ' ' + fmtNum(rw.failed) })
        ] };
      }))));
    }, function (err) { clear(body).appendChild(notice('error', lbErrText(err))); });
  }
  reload();
}

function lbSettingsPage(root) {
  settingsPage(root, 'settings', '/leaderboardx', function (cur, defs, d) {
    var opt = function (prefix, list) { return list.map(function (v) { return { value: v, label: t(prefix + v) }; }); };
    var notes = h('div', { class: 'flex flex-col gap-2' });
    if (d.prizes === false) notes.appendChild(notice('warning', t('se.noPrizes')));
    var ds = cur.display, r = cur.rewards, s = cur.scores, hi = cur.history, e = cur.events;
    var topLimit = numInput(ds.topLimit, 'display.topLimit'), maxLimit = numInput(ds.maxLimit, 'display.maxLimit'), around = numInput(ds.aroundLimit, 'display.aroundLimit');
    var showName = check(ds.showDisplayName, t('se.showDisplayName'), 'display.showDisplayName');
    var showUpcoming = check(ds.showUpcoming, t('se.showUpcoming'), 'display.showUpcoming');
    var endedDays = numInput(ds.showEndedDays, 'display.showEndedDays');
    var payMode = select(r.payMode, opt('pay.', ['auto', 'claim']), { name: 'rewards.payMode' });
    var claimDays = numInput(r.claimDays, 'rewards.claimDays');
    var stack = select(r.stack, opt('se.stack.', ['first', 'all']), { name: 'rewards.stack' });
    var groupMode = select(r.prizeGroupMode, opt('se.groupMode.', ['draw', 'all']), { name: 'rewards.prizeGroupMode' });
    var maxTries = numInput(r.maxTries, 'rewards.maxTries'), catchUp = numInput(r.catchUpDays, 'rewards.catchUpDays');
    var reqId = check(s.requireRequestId, t('se.requireRequestId'), 'scores.requireRequestId');
    var perMinute = numInput(s.playerPerMinute, 'scores.playerPerMinute'), logDays = numInput(s.logDays, 'scores.logDays'), clientMax = numInput(s.clientMaxPerWrite, 'scores.clientMaxPerWrite');
    var keepDays = numInput(hi.keepDays, 'history.keepDays'), playerLimit = numInput(hi.playerLimit, 'history.playerLimit'), winnersLimit = numInput(hi.winnersLimit, 'history.winnersLimit');
    var ecoEvent = input(e.economyEvent, { name: 'events.economyEvent', ltr: true, mono: true });
    var anEvent = input(e.analyticsEvent, { name: 'events.analyticsEvent', ltr: true, mono: true });
    var notify = check(e.notify, t('se.notify'), 'events.notify');
    var code = numInput(e.notifyCode, 'events.notifyCode');
    var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
    return {
      el: [notes,
        section(t('se.display'), h('div', {}, grid(field(t('se.topLimit'), topLimit), field(t('se.maxLimit'), maxLimit)), grid(field(t('se.aroundLimit'), around), field(t('se.showEndedDays'), endedDays)), showName, showUpcoming), 'display'),
        section(t('se.rewards'), h('div', {}, grid(field(t('se.payMode'), payMode), field(t('se.claimDays'), claimDays)), grid(field(t('se.stack'), stack), field(t('se.groupMode'), groupMode)),
          grid(field(t('se.maxTries'), maxTries), field(t('se.catchUpDays'), catchUp))), 'rewards'),
        section(t('se.scores'), h('div', {}, reqId, grid(field(t('se.playerPerMinute'), perMinute), field(t('se.logDays'), logDays)), field(t('se.clientMaxPerWrite'), clientMax, null, 'lb-num')), 'scores'),
        section(t('se.history'), h('div', {}, field(t('se.keepDays'), keepDays, null, 'lb-num'), grid(field(t('se.playerLimit'), playerLimit), field(t('se.winnersLimit'), winnersLimit))), 'history'),
        section(t('se.events'), h('div', {}, grid(field(t('se.economyEvent'), ecoEvent), field(t('se.analyticsEvent'), anEvent)), notify, field(t('se.notifyCode'), code, null, 'lb-num')), 'events')],
      read: function (c) {
        c.display = { topLimit: numberValue(topLimit), maxLimit: numberValue(maxLimit), aroundLimit: numberValue(around), showDisplayName: showName.input.checked, showUpcoming: showUpcoming.input.checked, showEndedDays: numberValue(endedDays) };
        c.rewards = { payMode: payMode.value, claimDays: numberValue(claimDays), stack: stack.value, prizeGroupMode: groupMode.value, maxTries: numberValue(maxTries), catchUpDays: numberValue(catchUp) };
        c.scores = { requireRequestId: reqId.input.checked, playerPerMinute: numberValue(perMinute), logDays: numberValue(logDays), clientMaxPerWrite: numberValue(clientMax) || 0 };
        c.history = { keepDays: numberValue(keepDays), playerLimit: numberValue(playerLimit), winnersLimit: numberValue(winnersLimit) };
        c.events = { economyEvent: ecoEvent.value.trim(), analyticsEvent: anEvent.value.trim(), notify: notify.input.checked, notifyCode: numberValue(code) || 0 };
        return null;
      }
    };
  });
}

var LB_PLAYER = { id: null, data: null, loading: false, at: 0 };

function lbPlayerRoute() {
  var m = /^#\/players\/([0-9a-f-]{36})(?:\?.*)?$/.exec(location.hash || '');
  return m ? m[1] : null;
}

function lbLoadPlayer(id) {
  LB_PLAYER.loading = true;
  return api('GET', '/leaderboardx/players/' + id).then(function (d) { return d; }, function (err) { return { error: err }; }).then(function (d) {
    LB_PLAYER.loading = false;
    LB_PLAYER.at = Date.now();
    if (LB_PLAYER.id === id) LB_PLAYER.data = d;
    lbSchedule();
  });
}

function lbPlayerBody(d) {
  if (d.error) return notice('warning', lbErrText(d.error));
  var box = h('div', { class: 'flex flex-col gap-3' });
  var cur = d.current || [];
  var hist = d.history || [];
  if (!cur.length && !hist.length) { box.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('pc.none') })); return box; }
  if (cur.length) {
    var ol = h('ol', { class: 'flex flex-col gap-1', 'data-lbx-player-current': '' });
    cur.forEach(function (c) {
      ol.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-lbx-player-board': c.boardKey },
        h('a', { href: '#/leaderboardx/standings?board=' + encodeURIComponent(c.boardKey), class: 'eco-link font-medium', text: textOf(c.name, c.boardKey) }),
        h('span', { 'data-lbx-player-rank': c.boardKey, text: t('pc.rank', { r: c.rank }) }), h('span', { class: 'text-xs text-secondary-500', text: t('pc.score', { s: fmtNum(c.score) }) }),
        c.tier ? tag(c.tier, 'primary') : null, statusTag(c.status)));
    });
    box.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pc.current') }), ol));
  }
  if (hist.length) {
    var hl = h('ol', { class: 'flex flex-col gap-1', 'data-lbx-player-history': '' });
    hist.forEach(function (r) {
      hl.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-xs', 'data-lbx-player-result': r.periodId },
        h('span', { class: 'font-medium', text: textOf(r.boardName, r.boardKey) }), h('span', { class: 'text-secondary-500', text: fmtDate(r.periodEnd) }),
        h('span', { text: t('pc.rank', { r: r.rank }) }), r.tier ? tag(r.tier, 'primary') : null, h('span', { text: rewardsText(r.rewards) }),
        r.status !== 'none' ? resultTag(r.status) : null));
    });
    box.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pc.history') }), hl));
  }
  if (d.refused30d) box.appendChild(h('p', { class: 'text-xs text-error-600', 'data-lbx-player-refused': '', text: t('pc.refused', { n: d.refused30d }) }));
  return box;
}

function lbPlayerCard() {
  var id = lbPlayerRoute();
  var existing = document.querySelector('[data-lbx-player]');
  if (!id || !token()) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (LB_PLAYER.id !== id) { LB_PLAYER.id = id; LB_PLAYER.data = null; LB_PLAYER.at = 0; }
  if (!LB_PLAYER.loading && Date.now() - LB_PLAYER.at > 20000) lbLoadPlayer(id);
  var d = LB_PLAYER.data;
  var key = lbLang() + '|' + id + '|' + (d ? JSON.stringify(d) : 'loading');
  if (existing && existing.getAttribute('data-key') === key && existing.parentNode === grid) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' }, h('h2', { id: 'lbx-player-title', class: 'font-medium mb-3', text: t('pc.title') }));
  body.appendChild(d ? lbPlayerBody(d) : h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  grid.appendChild(h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-lbx-player': '', 'data-key': key, 'data-l10n-skip': '',
    'aria-labelledby': 'lbx-player-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body));
}

var LB_ROOT = 'lbx-root';
var LB_PAGES = ['boards', 'standings', 'winners', 'report', 'settings'];
var LB_RENDER = { boards: boardsPage, standings: standingsPage, winners: winnersPage, report: reportPage, settings: lbSettingsPage };
var LB_HASH = /^#\/leaderboardx(?:\/([^/?]*))?/;
var lbPending = null;
var lbRouter = null;
var lbMounted = null;
var lbScheduled = false;

function lbRoute(hash) {
  var m = LB_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  return LB_PAGES.indexOf(m[1]) !== -1 ? m[1] : LB_PAGES[0];
}

function lbAllowed() {
  return !!token();
}

function lbTitle() {
  var page = lbRoute();
  return t('group') + (page ? ' - ' + t('nav.' + page) : '');
}

function lbAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || lbRouter) return !!lbRouter;
  try {
    if (!r.hasRoute('leaderboardx')) {
      var layout = r.resolve('/').matched[0];
      r.addRoute({ path: '/leaderboardx/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'leaderboardx', component: { name: 'LeaderboardxPage', render: function () { return null; } }, meta: { get title() { return lbTitle(); } } }] });
    }
    lbRouter = r;
  } catch (e) {
    return false;
  }
  if (lbPending) {
    var target = lbPending;
    lbPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function lbNavLink(page, active) {
  var on = page === active;
  return h('li', {}, h('a', { href: '#/leaderboardx/' + page, 'data-lbx-nav': page, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + page) }))));
}

function lbSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelector('[data-lbx-group]');
  if (!group || !lbAllowed()) { if (existing) existing.remove(); return; }
  var target = document.querySelector('.x-drawer [data-assets-group]') || document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = lbLang() + '|' + active;
  if (existing && existing.getAttribute('data-key') === key && existing.nextElementSibling === target) return;
  var built = h('div', { 'data-lbx-group': '', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group') }),
    h('nav', { 'aria-label': t('group') }, h('ul', { class: 'space-y-0.5' }, LB_PAGES.map(function (p) { return lbNavLink(p, active); }))));
  if (existing) existing.remove();
  target.before(built);
}

function lbShowMains() {
  var hidden = document.querySelectorAll('[data-lbx-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-lbx-hidden'); }
}

function lbUnmount() {
  var root = document.getElementById(LB_ROOT);
  if (root) root.remove();
  lbShowMains();
  closePanel();
  lbMounted = null;
}

function lbSync() {
  lbAddRoutes();
  var page = lbRoute();
  lbSyncNav(page);
  lbPlayerCard();
  if (!page) { if (lbMounted || document.getElementById(LB_ROOT)) lbUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-lbx-hidden', ''); }
  }
  var root = document.getElementById(LB_ROOT);
  if (!root) {
    root = h('div', { id: LB_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-lbx': 'root' });
    lbMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + lbLang() + (page === 'standings' || page === 'winners' || page === 'report' ? '|' + location.hash : '');
  if (lbMounted !== key) {
    lbMounted = key;
    closePanel();
    LB_RENDER[page](root);
  }
}

function lbSchedule() {
  if (lbScheduled) return;
  lbScheduled = true;
  requestAnimationFrame(function () { lbScheduled = false; lbSync(); });
}

function lbBoot() {
  window.addEventListener('hashchange', lbSchedule);
  window.addEventListener('popstate', lbSchedule);
  window.addEventListener('nakama-console:ready', lbSchedule);
  new MutationObserver(lbSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { lbMounted = null; lbSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (lbAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  lbSync();
}

window.NakamaLeaderboardx = { version: 1, pages: function () { return LB_PAGES.map(function (p) { return { id: p, href: '#/leaderboardx/' + p, label: t('nav.' + p) }; }); } };

if (LB_HASH.test(location.hash)) {
  lbPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', lbBoot);
else lbBoot();
})();
