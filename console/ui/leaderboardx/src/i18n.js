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
    'f.maxPerWrite': 'بیشترین امتیاز هر ارسال (۰ = بدون سقف)', 'f.perMinute': 'سقف ارسال هر بازیکن در دقیقه (۰ = بدون سقف)', 'f.onExceed': 'وقتی بیشتر از سقف بود',
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
    'se.scores': 'ارسال امتیاز', 'se.requireRequestId': 'ارسال بازی requestId داشته باشد (ارسال تکراری دو بار حساب نشود)', 'se.playerPerMinute': 'سقف ارسال هر بازیکن در دقیقه در همهٔ جدول‌ها', 'se.logDays': 'نگهداری گزارش امتیازها (روز)',
    'se.history': 'تاریخچه', 'se.keepDays': 'نگهداری دوره‌ها (روز، ۰ = همیشه)', 'se.playerLimit': 'ردیف‌های تاریخچهٔ هر بازیکن', 'se.winnersLimit': 'ردیف‌های فهرست برنده‌ها',
    'se.events': 'رویدادهای پایان دوره', 'se.economyEvent': 'رویداد اقتصاد برای جایزه‌دارها (مقدار = رتبه)', 'se.analyticsEvent': 'رویداد آنالیتیکس (زمینه = کلید جدول)', 'se.notify': 'به بازیکن اعلان فرستاده شود', 'se.notifyCode': 'کد اعلان',
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
    'f.maxPerWrite': 'Largest score per write (0 = no cap)', 'f.perMinute': 'Writes per player per minute (0 = no cap)', 'f.onExceed': 'Above the cap',
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
    'se.scores': 'Score submits', 'se.requireRequestId': 'Game submits need a requestId (a resend never counts twice)', 'se.playerPerMinute': 'Submits per player per minute over all boards', 'se.logDays': 'Days to keep the score log',
    'se.history': 'History', 'se.keepDays': 'Days to keep periods (0 = forever)', 'se.playerLimit': 'Rows of a player\'s history', 'se.winnersLimit': 'Rows of the winners list',
    'se.events': 'Period-end events', 'se.economyEvent': 'Economy event for rewarded players (value = rank)', 'se.analyticsEvent': 'Analytics event (context = board key)', 'se.notify': 'Send the player a notification', 'se.notifyCode': 'Notification code',
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
