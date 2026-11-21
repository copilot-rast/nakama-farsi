(function () {
'use strict';
var QZ_TEXT = {
  fa: {
    'group': 'آزمون‌ها',
    'nav.packs': 'پک‌های آزمون', 'nav.questions': 'سوال‌ها', 'nav.missions': 'ماموریت‌ها و بازی‌ها', 'nav.reports': 'گزارش بازیکنان', 'nav.results': 'گزارش نتیجه',
    'nav.import': 'ورود و خروج', 'nav.stats': 'آمار آزمون‌ها', 'nav.settings': 'تنظیمات آزمون',
    'nav.duels': 'دوئل‌ها', 'nav.duel-categories': 'دسته‌های دوئل',
    'desc.duels': 'دوئل نوبتی دو بازیکن: هر دور یک دسته و چند سوال دارد و بازیکن‌ها به نوبت جواب می‌دهند. اینجا دوئل‌ها را ببینید، جزئیات هر جواب را بررسی کنید و در صورت نیاز دوئل باز را لغو کنید (ورودی برمی‌گردد).',
    'desc.duel-categories': 'دسته‌هایی که در دوئل به بازیکن پیشنهاد می‌شوند. سوال‌ها با همان کلید در فیلد «دسته» به دسته وصل می‌شوند. دسته‌ای که سوال کافی نداشته باشد پیشنهاد نمی‌شود.',
    'dc.new': 'دستهٔ تازه', 'dc.questions': 'سوال', 'dc.picks': 'بار انتخاب', 'dc.state': 'وضعیت', 'dc.count': '{n} سوال فعال · {e} مناسب دوئل', 'dc.playable': 'قابل بازی',
    'dc.notEnough': 'کمتر از {n} سوال', 'dc.inDuel': 'در دوئل', 'dc.notInDuel': 'بیرون از دوئل', 'dc.empty': 'هنوز دسته‌ای ساخته نشده است.', 'dc.delete': 'دستهٔ {key} حذف شود؟',
    'dc.inUse': 'سوال‌ها یا دوئل‌هایی از این دسته استفاده کرده‌اند. بهتر است خاموشش کنید. با این حال حذف شود؟', 'dc.color': 'رنگ', 'dc.duelOn': 'در دوئل پیشنهاد شود',
    'dc.hint': 'نام و آیکون همان چیزی است که بازیکن هنگام انتخاب دسته می‌بیند.', 'dc.hint.key': 'همین کلید را در فیلد «دسته»ی سوال‌ها بنویسید. بعداً عوض نمی‌شود.',
    'du.status': 'وضعیت', 'du.mode': 'نوع', 'du.player': 'بازیکن', 'du.playerHint': 'نام کاربری یا شناسه', 'du.playerA': 'شروع‌کننده', 'du.playerB': 'حریف', 'du.turn': 'نوبت',
    'du.updated': 'آخرین تغییر', 'du.empty': 'دوئلی پیدا نشد.', 'du.noOpponent': 'هنوز حریفی نیامده', 'du.winner': 'برنده', 'du.until': 'تا {at}',
    'du.st.waiting': 'در انتظار حریف', 'du.st.invited': 'دعوت دوست', 'du.st.active': 'در جریان', 'du.st.finished': 'تمام‌شده', 'du.st.expired': 'منقضی', 'du.st.declined': 'ردشده', 'du.st.cancelled': 'لغوشده',
    'du.mode.random': 'حریف تصادفی', 'du.mode.friend': 'با دوست',
    'du.reason.score': 'با امتیاز', 'du.reason.draw': 'مساوی', 'du.reason.timeout': 'تمام شدن وقت نوبت', 'du.reason.forfeit': 'انصراف', 'du.reason.expired': 'حریف پیدا نشد',
    'du.reason.declined': 'دعوت رد شد', 'du.reason.admin': 'لغو با ادمین',
    'du.kpi.started': 'دوئل در ۳۰ روز', 'du.kpi.finished': 'تمام‌شده', 'du.kpi.open': 'در جریان', 'du.kpi.waiting': 'در انتظار حریف', 'du.kpi.timeouts': 'باخت با وقت', 'du.kpi.players': 'بازیکن',
    'du.detail': 'جزئیات دوئل', 'du.cancel': 'لغو دوئل', 'du.cancelConfirm': 'دوئل لغو شود؟ ورودی هر دو بازیکن برمی‌گردد و جایزه‌ای داده نمی‌شود.', 'du.cancelled': 'دوئل لغو شد',
    'du.created': 'شروع', 'du.finished': 'پایان', 'du.deadline': 'مهلت نوبت', 'du.entryFee': 'ورودی', 'du.feeV': '{n} {c}', 'du.paid': 'جایزه‌ها پرداخت شد',
    'du.fee.none': 'بدون ورودی', 'du.fee.paid': 'ورودی پرداخت شد', 'du.fee.refunded': 'ورودی برگشت', 'du.round': 'دور {n}', 'du.notPicked': 'هنوز انتخاب نشده',
    'du.pickedBy': 'انتخاب: {name}', 'du.offers': 'پیشنهادها', 'du.notAnswered': 'جواب نداده', 'du.secs': '{n} ثانیه', 'du.served': 'نمایش: {at}',
    'du.out.correct': 'درست', 'du.out.wrong': 'غلط', 'du.out.late': 'دیر', 'du.out.skip': 'بی‌جواب', 'du.out.pending': 'در حال جواب',
    'ds.title': 'دوئل', 'ds.hint': 'این مقدارها هنگام شروع هر دوئل در خودش ثبت می‌شوند؛ تغییرشان روی دوئل‌های در جریان اثری ندارد.', 'ds.enabled': 'دوئل روشن باشد',
    'ds.rounds': 'تعداد دور', 'ds.per': 'سوال در هر دور', 'ds.choices': 'دسته‌های پیشنهادی در هر دور', 'ds.hint.choices': 'بازیکن از میان این تعداد دسته یکی را انتخاب می‌کند.',
    'ds.source': 'سوال‌ها از', 'ds.source.any': 'همهٔ سوال‌های فعال دسته', 'ds.source.bank': 'فقط بانک سوال (بدون پک)',
    'ds.qsec': 'زمان هر سوال (ثانیه)', 'ds.grace': 'مهلت اضافه (ثانیه)', 'ds.hint.grace': 'برای تأخیر شبکه؛ جواب بعد از آن «دیر» و غلط حساب می‌شود.',
    'ds.avoid': 'تکرار نشدن سوال (روز)', 'ds.hint.avoid': 'سوال‌هایی که هر کدام از دو بازیکن در این مدت دیده‌اند آخر از همه انتخاب می‌شوند. ۰ یعنی خاموش.',
    'ds.shuffle': 'جابه‌جایی گزینه‌ها (برای هر دو بازیکن یکسان)', 'ds.serveNext': 'سوال بعدی همراه جواب فرستاده شود (زمانش از همان لحظه شروع می‌شود)',
    'ds.turn': 'مهلت هر نوبت (ساعت)', 'ds.hint.turn': 'اگر بازیکن در این مدت نوبتش را تمام نکند، طبق قانون زیر بازنده می‌شود.',
    'ds.timeout': 'نتیجهٔ تمام شدن وقت', 'ds.timeout.opponent_wins': 'حریف برنده می‌شود', 'ds.timeout.score': 'با امتیاز تا همان لحظه',
    'ds.waiting': 'انتظار برای حریف تصادفی (ساعت)', 'ds.hint.waiting': 'بعد از آن دوئل منقضی می‌شود و ورودی برمی‌گردد.', 'ds.invite': 'اعتبار دعوت دوست (ساعت)',
    'ds.friendsOnly': 'فقط دوست‌ها را می‌شود دعوت کرد', 'ds.maxActive': 'سقف دوئل باز هر بازیکن', 'ds.daily': 'سقف شروع دوئل در روز',
    'ds.feeCur': 'ارز ورودی', 'ds.feeAmt': 'مقدار ورودی', 'ds.hint.fee': 'خالی یعنی رایگان. در لغو، انقضا و رد دعوت برمی‌گردد.',
    'ds.win': 'جایزهٔ برنده', 'ds.lose': 'جایزهٔ بازنده', 'ds.draw': 'جایزهٔ مساوی', 'ds.pcCur': 'ارز هر جواب درست', 'ds.pcAmt': 'مقدار برای هر جواب درست',
    'ds.hint.pc': 'در پایان دوئل به هر بازیکن به ازای جواب‌های درستش داده می‌شود.', 'ds.evPlayed': 'رویداد بازی دوئل', 'ds.evWon': 'رویداد برد دوئل', 'ds.evCorrect': 'رویداد جواب درست',
    'ds.notify': 'اعلان زنده برای دعوت، نوبت و پایان', 'ds.history': 'نگه‌داری دوئل‌های تمام‌شده (روز)',
    'code.QUIZ_DUEL_OVER': 'این دوئل تمام شده است.', 'code.INVALID_COLOR': 'رنگ باید به شکل #RRGGBB باشد.', 'code.KEY_EXISTS': 'این کلید قبلاً استفاده شده است.', 'code.PRIZE_NOT_FOUND': 'این جایزه پیدا نشد.',
    'desc.packs': 'هر پک یک آزمون است: سوال‌های خودش یا سوال‌هایی که از بانک یک دسته کشیده می‌شوند، شرط قبولی، زمان، جایزهٔ قبولی و اتصال به مرحلهٔ یک‌بارهٔ کمپین.',
    'desc.questions': 'سوال‌های چندگزینه‌ای و تشریحی با متن ترجمه‌شده، گزینهٔ درست، توضیح، رسانه، سختی، برچسب و راهنما. پاسخ درست هیچ‌وقت پیش از جواب دادن به بازی فرستاده نمی‌شود.',
    'desc.missions': 'ماموریت چند پک را کنار هم می‌گذارد و بازی چند ماموریت را؛ به ترتیب یا همه با هم، با زمان و جایزهٔ یک‌باره برای تمام کردن.',
    'desc.reports': 'گزارش‌هایی که بازیکنان دربارهٔ سوال‌ها فرستاده‌اند. بررسی کنید، سوال را اصلاح یا خاموش کنید و گزارش را ببندید.',
    'desc.results': 'متن گزارش نتیجه با جای‌خالی‌هایی مثل {score} و {status} که برای هر بازیکن پر می‌شود و در صورت نیاز به یک وب‌هوک فرستاده می‌شود.',
    'desc.import': 'سوال‌ها، پک‌ها و ماموریت‌ها را یکجا با JSON وارد کنید یا خروجی بگیرید. اگر حتی یک خطا باشد هیچ چیز ثبت نمی‌شود.',
    'desc.stats': 'بازی، قبولی، میانگین امتیاز و زمان هر پک در روزهای اخیر، و سخت‌ترین سوال‌ها.',
    'desc.settings': 'سیاست‌های آزمون: ارسال سوال‌ها، زمان، قبولی، نمایش پاسخ، جایزه، سقف‌ها، راهنما، گزارش، رویدادها و اتصال به کمپین.',
    'save': 'ذخیره', 'saved': 'ذخیره شد', 'cancel': 'انصراف', 'close': 'بستن', 'remove': 'حذف', 'delete': 'حذف', 'search': 'جستجو', 'all': 'همه', 'edit': 'ویرایش',
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'yes': 'بله', 'no': 'خیر',
    'inactive': 'غیرفعال', 'active': 'فعال', 'confirm.title': 'مطمئن هستید؟', 'lang.text': '{label} ({lang})', 'updated': 'آخرین تغییر: {by} · {at}', 'none': '—', 'default': 'پیش‌فرض تنظیمات',
    'f.key': 'کلید', 'f.name': 'نام', 'f.description': 'توضیح', 'f.icon': 'آدرس آیکون', 'f.sort': 'ترتیب', 'f.active': 'فعال باشد', 'f.mission': 'ماموریت', 'f.pack': 'پک',
    'f.category': 'دسته', 'f.tags': 'برچسب‌ها', 'f.language': 'زبان', 'f.difficulty': 'سختی', 'f.reward': 'جایزه', 'f.metadata': 'دادهٔ اضافه (JSON)',
    'hint.key': 'شناسه‌ای که بازی می‌فرستد؛ حرف و عدد انگلیسی، بعداً عوض نمی‌شود.', 'hint.tags': 'با ویرگول جدا کنید', 'hint.language': 'خالی یعنی همهٔ زبان‌ها',
    'hint.count': 'خالی یا ۰ یعنی همهٔ سوال‌ها. با عدد کمتر، هر بار به همین تعداد سوال تصادفی کشیده می‌شود.',
    'hint.pool': 'اگر پر باشد سوال‌ها از همهٔ سوال‌های فعال این دسته (بانک سوال) کشیده می‌شوند، نه از سوال‌های خود پک.',
    'hint.qtime': 'خالی یعنی بدون زمان.', 'hint.qtimeDefault': 'خالی یعنی پیش‌فرض تنظیمات ({n} ثانیه).', 'hint.ptime': 'زمان کل از شروع؛ بعد از آن پاسخ پذیرفته نمی‌شود و نتیجه ثبت می‌شود.',
    'hint.sendAnswers': 'فقط وقتی کار می‌کند که در تنظیمات «اجازهٔ فرستادن پاسخ همراه سوال» روشن باشد.',
    'hint.passCorrect': 'اگر پر باشد بر درصد مقدم است؛ مثلاً ۹ از ۱۰.', 'hint.passPercent': 'خالی یعنی پیش‌فرض تنظیمات ({n}٪).',
    'hint.packReward': 'با قبولی پرداخت می‌شود؛ یک بار یا هر بار طبق تنظیمات. ارزها به کیف پول و جایزهٔ اقتصاد از همان مسیر جایزه‌ها.',
    'hint.stage': 'با قبولی در این پک، تلاش همین مرحلهٔ کمپین روی سرور تمام می‌شود و بازی دیگر نمی‌تواند نتیجه‌اش را خودش بفرستد.',
    'hint.missionTime': 'از اولین شروع بازیکن حساب می‌شود. خالی یعنی بدون زمان.', 'hint.missionReward': 'وقتی همهٔ پک‌ها (یا ماموریت‌های بازی) قبول شدند، فقط یک بار پرداخت می‌شود.',
    'hint.bank': 'سوال بدون پک در بانک می‌ماند و پک‌هایی که دسته‌اش را بکشند از آن استفاده می‌کنند.', 'hint.media': 'آدرس تصویر یا ویدیو', 'hint.category': 'برای بانک سوال و گزارش',
    'hint.options': 'گزینه‌ها به همین ترتیب ذخیره می‌شوند و بسته به تنظیمات برای هر بازیکن جابه‌جا می‌شوند. دست‌کم یک گزینهٔ درست لازم است.',
    'hint.accepted': 'هر پاسخ قابل قبول در یک خط', 'hint.qtimeOwn': 'خالی یعنی زمان پک', 'hint.qkey': 'اختیاری؛ ورود دوباره با همین کلید سوال را به‌روز می‌کند.',
    'hint.questionReward': 'برای پاسخ درست؛ یک بار برای هر بازیکن یا هر بار طبق تنظیمات. همراه نتیجهٔ آزمون پرداخت می‌شود.',
    'hint.hints': 'بازیکن هر راهنما را یک بار می‌خرد؛ قیمت از کیف پول کم می‌شود و تکرار درخواست دوباره پول نمی‌گیرد.',
    'hint.rsMission': 'آمار گزارش از پک‌های این ماموریت یا ماموریت‌های این بازی ساخته می‌شود.', 'hint.webhook': 'اختیاری؛ فقط آدرس عمومی http یا https. هر گزارش در صف ارسال می‌رود.',
    'hint.secret': 'با این کلید امضای HMAC-SHA256 در سرآیند X-Quiz-Report-Signature فرستاده می‌شود. خالی گذاشتن، کلید قبلی را نگه می‌دارد.',
    'hint.importPack': 'اگر انتخاب شود همهٔ سوال‌ها به همین پک می‌روند.', 'hint.importLang': 'زبان متن سطرهای جدولی (rows)',
    'pk.new': 'پک تازه', 'pk.noMission': 'بدون ماموریت', 'pk.questions': 'سوال', 'pk.pass': 'شرط قبولی', 'pk.time': 'زمان', 'pk.stageTag': 'مرحله: {k}',
    'pk.fromBank': '{n} از بانک {c}', 'pk.passCorrect': '{n} پاسخ درست', 'pk.passPercent': '{n}٪', 'pk.asked': '{n} سوال در هر بار', 'pk.perQuestion': '{n} ثانیه برای هر سوال', 'pk.whole': 'کل: {n} ثانیه',
    'pk.delete': 'پک {key} حذف شود؟', 'pk.inUse': 'این پک بازی شده است. بهتر است خاموشش کنید. با این حال حذف شود (نتیجه‌های بازیکنان هم پاک می‌شود)؟',
    'pk.play': 'بازی', 'pk.count': 'تعداد سوال در هر بار', 'pk.pool': 'کشیدن از بانک دستهٔ', 'pk.qtime': 'زمان هر سوال (ثانیه)', 'pk.ptime': 'زمان کل پک (ثانیه)',
    'pk.shuffleQ': 'جابه‌جایی سوال‌ها', 'pk.shuffleO': 'جابه‌جایی گزینه‌ها', 'pk.sendAnswers': 'پاسخ درست همراه سوال فرستاده شود', 'pk.passCorrectF': 'تعداد پاسخ درست لازم', 'pk.passPercentF': 'درصد قبولی',
    'pk.reward': 'جایزهٔ قبولی', 'pk.campaign': 'اتصال به کمپین', 'pk.stage': 'مرحلهٔ کمپین (یک‌باره)', 'pk.noStage': 'بدون مرحله', 'pk.openQuestions': 'سوال‌های این پک',
    'pk.s.players': '{n} بازیکن', 'pk.s.plays': '{n} بار بازی', 'pk.s.wins': '{n} قبولی', 'pk.s.rate': 'نرخ قبولی {n}٪', 'pk.s.best': 'میانگین بهترین امتیاز {n}',
    'qs.new': 'سوال تازه', 'qs.edit': 'ویرایش سوال', 'qs.bank': 'بانک (بدون پک)', 'qs.type': 'نوع', 'qs.type.options': 'چندگزینه‌ای', 'qs.type.text': 'تشریحی (متن)',
    'qs.reportedOnly': 'فقط سوال‌های گزارش‌شده', 'qs.question': 'سوال', 'qs.answer': 'پاسخ درست', 'qs.hints': 'راهنما', 'qs.reports': '{n} گزارش باز',
    'qs.delete': 'این سوال حذف شود؟', 'qs.inUse': 'بازیکنان به این سوال پاسخ داده‌اند. بهتر است خاموشش کنید. با این حال حذف شود؟',
    'qs.option': 'گزینه', 'qs.optionN': 'گزینهٔ {n}', 'qs.correct': 'درست', 'qs.media': 'آدرس تصویر', 'qs.mediaF': 'رسانهٔ سوال', 'qs.addOption': 'گزینهٔ تازه', 'qs.options': 'گزینه‌ها',
    'qs.accepted': 'پاسخ‌های قابل قبول', 'qs.explanation': 'توضیح بعد از پاسخ', 'qs.time': 'زمان این سوال (ثانیه)', 'qs.reward': 'جایزهٔ پاسخ درست',
    'qs.s.answers': '{n} پاسخ', 'qs.s.correct': '{n}٪ درست',
    'hn.title': 'راهنماها', 'hn.add': 'راهنمای تازه', 'hn.afterSave': 'بعد از ذخیرهٔ سوال می‌توانید راهنما اضافه کنید.', 'hn.free': 'رایگان', 'hn.bought': '{n} خرید',
    'hn.delete': 'این راهنما حذف شود؟ خریدهای قبلی هم پاک می‌شوند.', 'hn.text': 'متن راهنما', 'hn.currency': 'ارز قیمت', 'hn.amount': 'قیمت',
    'ms.new': 'ماموریت تازه', 'ms.newGame': 'بازی تازه', 'ms.kind': 'نوع', 'ms.kind.mission': 'ماموریت', 'ms.kind.game': 'بازی', 'ms.game': 'بازی', 'ms.noGame': 'بدون بازی',
    'ms.order': 'ترتیب', 'ms.order.all_at_once': 'همه با هم', 'ms.order.sequential': 'به ترتیب', 'ms.time': 'زمان (ثانیه)', 'ms.timeF': 'زمان از شروع (ثانیه)', 'ms.items': 'محتوا',
    'ms.packs': '{n} پک', 'ms.missions': '{n} ماموریت', 'ms.reward': 'جایزهٔ تمام کردن', 'ms.delete': '{key} حذف شود؟', 'ms.inUse': 'هنوز پک یا ماموریتی به این وصل است.',
    'ms.empty': 'هنوز ماموریتی نساخته‌اید. پک‌ها بدون ماموریت هم بازی می‌شوند.',
    'ms.hint.mission': 'پک‌ها را در صفحهٔ پک‌ها به این ماموریت وصل کنید.', 'ms.hint.game': 'ماموریت‌ها را در فرم ماموریت به این بازی وصل کنید.',
    'ms.orderHint.mission': 'به ترتیب: هر پک بعد از قبولی در پک قبلی باز می‌شود.', 'ms.orderHint.game': 'به ترتیب: هر ماموریت بعد از تمام شدن ماموریت قبلی باز می‌شود.',
    'rp.reasons': 'دلیل‌های گزارش', 'rp.reasonsHint': 'بازیکن هنگام گزارش یکی از دلیل‌های فعال را انتخاب می‌کند.', 'rp.status': 'وضعیت', 'rp.reason': 'دلیل', 'rp.openCount': '{n} گزارش باز',
    'rp.date': 'زمان', 'rp.player': 'بازیکن', 'rp.note': 'توضیح بازیکن', 'rp.review': 'بررسی', 'rp.empty': 'گزارشی نیست.', 'rp.same': 'همهٔ گزارش‌های باز همین سوال ({n} مورد)',
    'rp.disable': 'سوال خاموش شود', 'rp.editQuestion': 'باز کردن سوال برای اصلاح', 'rp.resolution': 'نتیجهٔ بررسی', 'rp.do.resolved': 'رسیدگی شد', 'rp.do.dismissed': 'رد شد', 'rp.do.open': 'باز شود',
    'rp.newReason': 'دلیل تازه', 'rp.reasonInUse': 'گزارش‌هایی این دلیل را دارند؛ به جای حذف خاموشش کنید.',
    'st.open': 'باز', 'st.resolved': 'رسیدگی‌شده', 'st.dismissed': 'ردشده', 'st.done': 'تمام', 'st.paying': 'در حال پرداخت', 'st.closed': 'بسته', 'st.pending': 'در صف', 'st.sent': 'فرستاده شد', 'st.failed': 'ناموفق',
    'rs.new': 'گزارش تازه', 'rs.mission': 'ماموریت یا بازی', 'rs.webhook': 'وب‌هوک', 'rs.deliveries': 'ارسال‌ها', 'rs.delete': 'گزارش {key} حذف شود؟', 'rs.empty': 'هنوز گزارش نتیجه‌ای نساخته‌اید.',
    'rs.body': 'متن گزارش', 'rs.placeholders': 'برای افزودن، روی جای‌خالی بزنید:', 'rs.insert': 'افزودن به متن', 'rs.won': 'متن قبولی', 'rs.lost': 'متن مردودی', 'rs.url': 'آدرس وب‌هوک',
    'rs.secret': 'کلید امضا', 'rs.secretKept': 'کلید قبلی نگه داشته می‌شود', 'rs.noSecrets': 'ECONOMY_SECRET_KEY تنظیم نشده؛ کلید امضا ذخیره نمی‌شود.', 'rs.attempts': 'تلاش', 'rs.error': 'خطا',
    'im.title': 'ورود', 'im.help': 'فرمت: {"missions":[…],"packs":[…],"questions":[…],"rows":[…]}. سوال با key موجود به‌روز می‌شود. rows همان ستون‌های جدول قدیمی است (Pack Key، Question، Option 1 تا 6، Correct Options، Accepted Answers، Answer Language، Difficulty، Time). «نمونه» را ببینید.',
    'im.pack': 'پک مقصد', 'im.packFromFile': 'از خود فایل (packKey)', 'im.lang': 'زبان سطرها', 'im.file': 'فایل JSON', 'im.json': 'متن JSON', 'im.sample': 'نمونه', 'im.dry': 'بررسی بدون ثبت', 'im.go': 'ورود',
    'im.dryOk': 'فایل درست است.', 'im.done': 'وارد شد.', 'im.count.missions': 'ماموریت: {c} تازه، {u} به‌روز', 'im.count.packs': 'پک: {c} تازه، {u} به‌روز', 'im.count.questions': 'سوال: {c} تازه، {u} به‌روز',
    'im.section': 'بخش', 'im.row': 'ردیف', 'im.field': 'فیلد', 'im.error': 'خطا', 'im.exportTitle': 'خروجی', 'im.exportHelp': 'خروجی همان فرمت ورود است، با پاسخ‌های درست و راهنماها.',
    'im.exportPack': 'پک', 'im.export': 'دریافت JSON', 'im.exported': '{n} سوال در فایل',
    'sx.days': 'بازهٔ زمانی', 'sx.d': '{n} روز', 'sx.players': 'بازیکن', 'sx.plays': 'بازی', 'sx.finished': 'تمام‌شده', 'sx.passes': 'قبولی', 'sx.rate': 'نرخ قبولی', 'sx.avg': 'میانگین امتیاز',
    'sx.secs': 'میانگین زمان (ثانیه)', 'sx.hardest': 'سخت‌ترین سوال‌ها', 'sx.answers': 'پاسخ', 'sx.correctRate': 'درصد درست', 'sx.skipped': 'بی‌پاسخ یا دیر', 'sx.noHard': 'هنوز دادهٔ کافی نیست.',
    'sx.kpi.players': 'بازیکن‌ها', 'sx.kpi.plays': 'بازی‌ها', 'sx.kpi.finished': 'تمام‌شده', 'sx.kpi.rate': 'نرخ قبولی', 'sx.kpi.hints': 'راهنمای خریده‌شده', 'sx.kpi.reports': 'گزارش باز',
    'rw.currency': 'ارز', 'rw.amount': 'مقدار', 'rw.addCurrency': 'ارز', 'rw.prize': 'جایزهٔ اقتصاد', 'rw.group': 'گروه جایزه', 'rw.noPrize': 'بدون جایزه', 'rw.missing': 'پیدا نشد',
    'se.noPrizes': 'افزونهٔ اقتصاد بارگذاری نشده؛ جایزه‌های اقتصاد پرداخت نمی‌شوند (ارزها پرداخت می‌شوند).', 'se.noCampaign': 'افزونهٔ کمپین بارگذاری نشده؛ اتصال پک به مرحله کاری نمی‌کند.',
    'se.play': 'بازی', 'se.delivery': 'ارسال سوال‌ها', 'se.delivery.all': 'همه در شروع (مثل قبل)', 'se.delivery.one_by_one': 'یکی‌یکی، سوال بعد پس از پاسخ',
    'se.hint.delivery': 'در حالت یکی‌یکی زمان هر سوال از لحظهٔ فرستادنش حساب می‌شود و بازیکن سوال‌های بعد را از قبل نمی‌بیند.',
    'se.shuffleQ': 'جابه‌جایی سوال‌ها (پک‌هایی که انتخاب خودشان را ندارند)', 'se.shuffleO': 'جابه‌جایی گزینه‌ها', 'se.restart': 'بازی دوبارهٔ پکی که آزمون بازش هست',
    'se.restart.resume': 'همان آزمون ادامه پیدا کند', 'se.restart.discard': 'آزمون قبلی بدون جایزه بسته شود', 'se.hours': 'عمر آزمون باز بدون زمان (ساعت)', 'se.hint.hours': 'بعد از آن آزمون با همان پاسخ‌ها نمره می‌گیرد.',
    'se.autoFinish': 'با پاسخ به آخرین سوال، نتیجه خودکار ثبت شود',
    'se.timing': 'زمان', 'se.qsec': 'زمان پیش‌فرض هر سوال (ثانیه)', 'se.hint.qsec': '۰ یعنی بدون زمان.', 'se.grace': 'حاشیهٔ تأخیر شبکه (ثانیه)', 'se.late': 'پاسخ بعد از زمان سوال',
    'se.late.wrong': 'غلط حساب شود (دیر)', 'se.late.reject': 'پذیرفته نشود',
    'se.scoring': 'قبولی و بررسی پاسخ', 'se.pass': 'درصد قبولی پیش‌فرض', 'se.hint.pass': 'پک می‌تواند تعداد یا درصد خودش را داشته باشد.', 'se.multi': 'سوال با چند گزینهٔ درست',
    'se.multi.any': 'انتخاب یکی از گزینه‌های درست کافی است', 'se.multi.all': 'باید همهٔ گزینه‌های درست و فقط آن‌ها انتخاب شوند', 'se.match': 'مقایسهٔ پاسخ متنی',
    'se.match.normalized': 'یکسان‌سازی (ی/ي، ک/ك، اعداد فارسی، فاصله و نیم‌فاصله)', 'se.match.exact': 'دقیق (فقط بزرگی و کوچکی حروف مهم نیست)', 'se.hint.match': 'پاسخ در هر زبانی که تعریف شده پذیرفته می‌شود.',
    'se.reveal': 'نمایش پاسخ درست', 'se.revealMode': 'پاسخ درست کی فرستاده شود', 'se.reveal.after_answer': 'بعد از هر پاسخ', 'se.reveal.after_finish': 'فقط در نتیجهٔ آخر', 'se.reveal.never': 'هرگز',
    'se.allowBefore': 'اجازهٔ فرستادن پاسخ همراه سوال برای پک‌ها و ماموریت‌هایی که «فرستادن پاسخ» دارند', 'se.explanation': 'توضیح سوال هم همراه پاسخ فرستاده شود',
    'se.rewards': 'جایزه', 'se.packMode': 'جایزهٔ قبولی پک', 'se.packMode.first_pass': 'فقط اولین قبولی', 'se.packMode.every_pass': 'هر قبولی', 'se.packMode.none': 'هیچ',
    'se.qMode': 'جایزهٔ پاسخ درست', 'se.qMode.first_correct': 'یک بار برای هر سوال', 'se.qMode.every_correct': 'هر بار', 'se.qMode.none': 'هیچ',
    'se.groupMode': 'پرداخت گروه جایزه', 'se.groupMode.draw': 'یک جایزه به قید قرعه', 'se.groupMode.all': 'همهٔ جایزه‌های گروه', 'se.zone': 'منطقهٔ زمانی شروع روز (سقف‌های روزانه)',
    'se.limits': 'سقف‌ها', 'se.dailyPlays': 'سقف بازی هر بازیکن در روز', 'se.dailyPasses': 'سقف قبولی جایزه‌دار در روز', 'se.hint.dailyPasses': 'بعد از سقف، قبولی ثبت می‌شود ولی جایزه ندارد. ۰ یعنی بدون سقف.',
    'se.maxPlays': 'سقف دفعات بازی هر پک', 'se.cooldown': 'فاصلهٔ دو بازی یک پک (ثانیه)', 'se.replay': 'پک قبول‌شده دوباره بازی شود', 'se.hint.zero': '۰ یعنی بدون سقف.',
    'se.hints': 'راهنما', 'se.pricing': 'قیمت راهنما', 'se.pricing.per_hint': 'قیمت خود هر راهنما', 'se.pricing.fixed': 'یک قیمت ثابت برای همه', 'se.pricing.free': 'رایگان',
    'se.keep': 'راهنمای خریده‌شده', 'se.keep.forever': 'برای همیشه مال بازیکن است', 'se.keep.session': 'فقط در همان آزمون', 'se.fixedCur': 'ارز قیمت ثابت', 'se.fixedAmt': 'قیمت ثابت', 'se.maxHints': 'سقف راهنمای هر سوال',
    'se.reports': 'گزارش سوال', 'se.reportsOn': 'بازیکنان بتوانند سوال را گزارش کنند', 'se.requireSeen': 'فقط سوالی که به بازیکن داده شده', 'se.noteRequired': 'نوشتن توضیح لازم باشد',
    'se.repDaily': 'سقف گزارش هر بازیکن در روز', 'se.repAuto': 'خاموش شدن خودکار سوال با این تعداد گزارش باز', 'se.hint.repAuto': '۰ یعنی هرگز.',
    'se.events': 'رویدادها', 'se.hint.events': 'کلید رویدادهای بازیکن در اقتصاد (برای کارها و دستاوردها) و نام رویداد تحلیل؛ خالی یعنی فرستاده نشود.',
    'se.evCompleted': 'تمام کردن آزمون (مقدار = امتیاز)', 'se.evPassed': 'قبولی (مقدار ۱)', 'se.evCorrect': 'پاسخ‌های درست (مقدار = تعداد)', 'se.evAnalytics': 'رویداد تحلیل (کلید پک = context)', 'se.onFailure': 'برای مردودی هم فرستاده شود',
    'se.campaign': 'کمپین', 'se.hint.campaign': 'پکی که به مرحلهٔ کمپین وصل است، با شروع، تلاش مرحله را باز می‌کند و با نتیجه، همان تلاش را روی سرور تمام می‌کند.',
    'se.campaignOn': 'اتصال پک به مرحلهٔ کمپین فعال باشد', 'se.sendFailures': 'مردودی هم به‌عنوان باخت به کمپین فرستاده شود',
    'se.retention': 'نگه‌داری', 'se.keepDays': 'نگه‌داری آزمون‌های تمام‌شده (روز)', 'se.hint.zeroKeep': '۰ یعنی همیشه.', 'se.delDays': 'نگه‌داری ارسال‌های وب‌هوک (روز)', 'se.attempts': 'دفعات تلاش وب‌هوک',
    'pc.title': 'آزمون‌ها', 'pc.noQuiz': 'افزونهٔ آزمون پاسخ نداد.', 'pc.summary': '{packs} پک بازی‌شده · {hints} راهنمای خریده‌شده · {reports} گزارش', 'pc.passed': 'قبول', 'pc.notPassed': 'هنوز قبول نشده',
    'pc.failed': 'مردود', 'pc.plays': '{p} بازی، {w} قبولی، بهترین {b}', 'pc.reset': 'از نو', 'pc.reset.confirm': 'پیشرفت پک «{name}» پاک شود؟',
    'pc.reset.rewards': 'جایزهٔ گرفته‌شده هم فراموش شود (قبولی بعدی دوباره جایزه می‌دهد)', 'pc.none': 'هنوز آزمونی بازی نکرده است.', 'pc.missions': 'ماموریت‌ها', 'pc.completed': 'تمام شده',
    'pc.started': 'شروع {at}', 'pc.rewarded': 'جایزه داده شد', 'pc.sessions': 'آزمون‌های اخیر', 'pc.score': '{c} از {n} · امتیاز {s}',
    'code.INVALID_KEY': 'کلید درست نیست؛ با حرف یا عدد انگلیسی شروع شود و فقط حرف، عدد و _ . : - داشته باشد.', 'code.ALREADY_EXISTS': 'این کلید از قبل هست.', 'code.IN_USE': 'در حال استفاده است.',
    'code.VALUE_OUT_OF_RANGE': 'مقدار بیرون از بازهٔ مجاز است.', 'code.INVALID_TIMEZONE': 'منطقهٔ زمانی شناخته نشد.', 'code.INVALID_POLICY': 'یکی از گزینه‌ها درست نیست.',
    'code.DEFAULT_LANGUAGE_REQUIRED': 'متن به زبان پیش‌فرض لازم است.', 'code.INVALID_JSON': 'JSON درست نیست.', 'code.INVALID_PRIZE_KEY': 'کلید جایزه درست نیست.', 'code.ONE_PRIZE_ONLY': 'یا جایزه یا گروه جایزه.',
    'code.QUIZ_OPTIONS_REQUIRED': 'دست‌کم دو گزینه لازم است.', 'code.QUIZ_CORRECT_OPTION_REQUIRED': 'دست‌کم یک گزینه را درست علامت بزنید.', 'code.QUIZ_OPTION_TEXT_REQUIRED': 'هر گزینه متن یا تصویر لازم دارد.',
    'code.DUPLICATE_OPTION': 'شناسهٔ گزینه‌ها تکراری است.', 'code.TOO_MANY_OPTIONS': 'حداکثر ۱۲ گزینه.', 'code.QUIZ_TEXT_ANSWERS_REQUIRED': 'دست‌کم یک پاسخ قابل قبول بنویسید.',
    'code.QUIZ_QUESTION_TEXT_OR_CONTENT_REQUIRED': 'متن سوال یا رسانه لازم است.', 'code.QUIZ_TYPE_IS_INVALID': 'نوع سوال درست نیست.', 'code.QUIZ_TYPE_INVALID': 'نوع سوال درست نیست.',
    'code.QUIZ_PACK_NOT_FOUND': 'پک پیدا نشد.', 'code.QUIZ_MISSION_NOT_FOUND': 'ماموریت پیدا نشد.', 'code.QUIZ_GAME_NOT_FOUND': 'بازی پیدا نشد.', 'code.STAGE_NOT_FOUND': 'مرحلهٔ کمپین پیدا نشد.',
    'code.QUIZ_PACK_WIN_THRESHOLD_INVALID': 'شرط قبولی درست نیست (بیشتر از تعداد سوال یا درصد بیرون از ۰ تا ۱۰۰).', 'code.QUIZ_MISSION_ORDER_INVALID': 'ترتیب درست نیست.',
    'code.QUIZ_HINT_PRICE_INVALID': 'قیمت راهنما درست نیست.', 'code.QUIZ_HINT_PRICE_REQUIRES_CURRENCY_ID': 'برای قیمت، ارز را انتخاب کنید.', 'code.QUIZ_HINT_VALUE_REQUIRED': 'متن یا تصویر راهنما لازم است.',
    'code.QUIZ_REPORT_TEMPLATE_NAME_IS_REQUIRED': 'نام گزارش لازم است.', 'code.QUIZ_REPORT_TEMPLATE_BODY_IS_REQUIRED': 'متن گزارش لازم است.',
    'code.QUIZ_REPORT_WEBHOOK_URL_IS_INVALID': 'آدرس وب‌هوک باید http یا https عمومی باشد (نه آدرس داخلی یا localhost).', 'code.QUIZ_SECRET_UNAVAILABLE': 'ECONOMY_SECRET_KEY تنظیم نشده؛ کلید امضا ذخیره نمی‌شود.',
    'code.QUIZ_IMPORT_HAS_ERRORS': 'فایل خطا دارد و هیچ چیز وارد نشد.', 'code.QUIZ_IMPORT_FILE_IS_REQUIRED': 'فایل خالی است.', 'code.QUIZ_IMPORT_TOO_LARGE': 'فایل خیلی بزرگ است (حداکثر ۵۰۰۰ سوال).',
    'code.QUIZ_OPTION_INDEX_NOT_FOUND': 'شمارهٔ گزینهٔ درست بیرون از گزینه‌هاست.', 'code.DUPLICATE_KEY': 'کلید تکراری در فایل.', 'code.INVALID_LANGUAGE': 'کد زبان درست نیست.',
    'code.TEXT_TOO_LONG': 'متن خیلی بلند است.', 'code.TOO_MANY_TAGS': 'حداکثر ۲۰ برچسب.', 'code.INVALID_STATUS': 'وضعیت درست نیست.', 'code.INVALID_KIND': 'نوع درست نیست.'
  },
  en: {
    'group': 'Quizzes',
    'nav.packs': 'Quiz packs', 'nav.questions': 'Questions', 'nav.missions': 'Missions and games', 'nav.reports': 'Player reports', 'nav.results': 'Result reports',
    'nav.import': 'Import and export', 'nav.stats': 'Quiz report', 'nav.settings': 'Quiz settings',
    'nav.duels': 'Duels', 'nav.duel-categories': 'Duel categories',
    'desc.duels': 'Turn-based duels between two players: each round has a category and a few questions, and the players answer in turns. See every duel, check each answer and cancel an open duel (entry fees are refunded).',
    'desc.duel-categories': 'Categories offered to players in a duel. Questions join a category through the same key in their Category field. A category without enough questions is not offered.',
    'dc.new': 'New category', 'dc.questions': 'Questions', 'dc.picks': 'Picks', 'dc.state': 'State', 'dc.count': '{n} active · {e} fit for duels', 'dc.playable': 'Playable',
    'dc.notEnough': 'Fewer than {n} questions', 'dc.inDuel': 'In duels', 'dc.notInDuel': 'Not in duels', 'dc.empty': 'No categories yet.', 'dc.delete': 'Delete category {key}?',
    'dc.inUse': 'Questions or duels use this category. Switching it off is safer. Delete anyway?', 'dc.color': 'Color', 'dc.duelOn': 'Offer it in duels',
    'dc.hint': 'The name and icon are what players see when they pick a category.', 'dc.hint.key': 'Write this key in the Category field of questions. It cannot change later.',
    'du.status': 'Status', 'du.mode': 'Mode', 'du.player': 'Player', 'du.playerHint': 'Username or id', 'du.playerA': 'Starter', 'du.playerB': 'Opponent', 'du.turn': 'Turn',
    'du.updated': 'Last change', 'du.empty': 'No duels found.', 'du.noOpponent': 'No opponent yet', 'du.winner': 'Winner', 'du.until': 'until {at}',
    'du.st.waiting': 'Waiting for opponent', 'du.st.invited': 'Friend invited', 'du.st.active': 'In progress', 'du.st.finished': 'Finished', 'du.st.expired': 'Expired', 'du.st.declined': 'Declined', 'du.st.cancelled': 'Cancelled',
    'du.mode.random': 'Random opponent', 'du.mode.friend': 'With a friend',
    'du.reason.score': 'By score', 'du.reason.draw': 'Draw', 'du.reason.timeout': 'Turn timed out', 'du.reason.forfeit': 'Forfeit', 'du.reason.expired': 'No opponent found',
    'du.reason.declined': 'Invite declined', 'du.reason.admin': 'Cancelled by an admin',
    'du.kpi.started': 'Duels in 30 days', 'du.kpi.finished': 'Finished', 'du.kpi.open': 'In progress', 'du.kpi.waiting': 'Waiting for opponent', 'du.kpi.timeouts': 'Lost on time', 'du.kpi.players': 'Players',
    'du.detail': 'Duel details', 'du.cancel': 'Cancel duel', 'du.cancelConfirm': 'Cancel this duel? Both entry fees are refunded and no prize is paid.', 'du.cancelled': 'Duel cancelled',
    'du.created': 'Started', 'du.finished': 'Ended', 'du.deadline': 'Turn deadline', 'du.entryFee': 'Entry fee', 'du.feeV': '{n} {c}', 'du.paid': 'Prizes paid',
    'du.fee.none': 'No entry fee', 'du.fee.paid': 'Entry fee paid', 'du.fee.refunded': 'Entry fee refunded', 'du.round': 'Round {n}', 'du.notPicked': 'Not picked yet',
    'du.pickedBy': 'Picked by {name}', 'du.offers': 'Offered', 'du.notAnswered': 'Not answered', 'du.secs': '{n} s', 'du.served': 'Shown {at}',
    'du.out.correct': 'Right', 'du.out.wrong': 'Wrong', 'du.out.late': 'Late', 'du.out.skip': 'Skipped', 'du.out.pending': 'Answering',
    'ds.title': 'Duel', 'ds.hint': 'These values are copied into each duel when it starts; changing them does not affect duels in progress.', 'ds.enabled': 'Duels are on',
    'ds.rounds': 'Rounds', 'ds.per': 'Questions per round', 'ds.choices': 'Categories offered per round', 'ds.hint.choices': 'The player picks one of this many categories.',
    'ds.source': 'Questions from', 'ds.source.any': 'Every active question of the category', 'ds.source.bank': 'Question bank only (no pack)',
    'ds.qsec': 'Time per question (seconds)', 'ds.grace': 'Grace (seconds)', 'ds.hint.grace': 'For network delay; an answer after it counts as late and wrong.',
    'ds.avoid': 'Avoid repeats (days)', 'ds.hint.avoid': 'Questions either player saw in this time are picked last. 0 turns it off.',
    'ds.shuffle': 'Shuffle options (the same order for both players)', 'ds.serveNext': 'Send the next question with each answer (its clock starts right away)',
    'ds.turn': 'Time per turn (hours)', 'ds.hint.turn': 'A player who does not finish the turn in time loses by the rule below.',
    'ds.timeout': 'When a turn times out', 'ds.timeout.opponent_wins': 'The opponent wins', 'ds.timeout.score': 'Decide by the score so far',
    'ds.waiting': 'Wait for a random opponent (hours)', 'ds.hint.waiting': 'After that the duel expires and the entry fee is refunded.', 'ds.invite': 'Friend invite lasts (hours)',
    'ds.friendsOnly': 'Only friends can be invited', 'ds.maxActive': 'Open duels per player', 'ds.daily': 'Duel starts per day',
    'ds.feeCur': 'Entry fee currency', 'ds.feeAmt': 'Entry fee amount', 'ds.hint.fee': 'Empty means free. Refunded on cancel, expiry and a declined invite.',
    'ds.win': 'Winner prize', 'ds.lose': 'Loser prize', 'ds.draw': 'Draw prize', 'ds.pcCur': 'Currency per right answer', 'ds.pcAmt': 'Amount per right answer',
    'ds.hint.pc': 'Paid to each player at the end for their right answers.', 'ds.evPlayed': 'Duel played event', 'ds.evWon': 'Duel won event', 'ds.evCorrect': 'Right answers event',
    'ds.notify': 'Live notifications for invites, turns and results', 'ds.history': 'Keep finished duels (days)',
    'code.QUIZ_DUEL_OVER': 'This duel is over.', 'code.INVALID_COLOR': 'Use a color like #RRGGBB.', 'code.KEY_EXISTS': 'This key is already used.', 'code.PRIZE_NOT_FOUND': 'Prize not found.',
    'desc.packs': 'Each pack is a quiz: its own questions or questions drawn from a category bank, a pass rule, timing, a pass reward and an optional link to a one-time campaign stage.',
    'desc.questions': 'Options and text questions with translated texts, right answers, explanation, media, difficulty, tags and hints. The right answer never reaches the game before the player answers.',
    'desc.missions': 'A mission groups packs and a game groups missions, in order or all at once, with a time limit and a one-time reward for finishing.',
    'desc.reports': 'Reports players sent about questions. Review them, fix or switch off the question and close the report.',
    'desc.results': 'Result report texts with placeholders such as {score} and {status}, filled for each player and optionally posted to a webhook.',
    'desc.import': 'Import questions, packs and missions as JSON, or export them. One error and nothing is written.',
    'desc.stats': 'Plays, passes, average score and time per pack over the last days, and the hardest questions.',
    'desc.settings': 'Quiz policies: question delivery, timing, passing, answer reveal, rewards, limits, hints, reports, events and the campaign link.',
    'save': 'Save', 'saved': 'Saved', 'cancel': 'Cancel', 'close': 'Close', 'remove': 'Remove', 'delete': 'Delete', 'search': 'Search', 'all': 'All', 'edit': 'Edit',
    'loading': 'Loading…', 'empty': 'Nothing found', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'yes': 'Yes', 'no': 'No',
    'inactive': 'Inactive', 'active': 'Active', 'confirm.title': 'Are you sure?', 'lang.text': '{label} ({lang})', 'updated': 'Last change: {by} · {at}', 'none': '—', 'default': 'Settings default',
    'f.key': 'Key', 'f.name': 'Name', 'f.description': 'Description', 'f.icon': 'Icon address', 'f.sort': 'Order', 'f.active': 'Active', 'f.mission': 'Mission', 'f.pack': 'Pack',
    'f.category': 'Category', 'f.tags': 'Tags', 'f.language': 'Language', 'f.difficulty': 'Difficulty', 'f.reward': 'Reward', 'f.metadata': 'Extra data (JSON)',
    'hint.key': 'The id the game sends; English letters and digits, cannot change later.', 'hint.tags': 'Separate with commas', 'hint.language': 'Empty = every language',
    'hint.count': 'Empty or 0 = every question. A smaller number draws that many random questions each play.',
    'hint.pool': 'When set, questions are drawn from every active question of this category (question bank) instead of the pack\'s own.',
    'hint.qtime': 'Empty = no time.', 'hint.qtimeDefault': 'Empty = settings default ({n} s).', 'hint.ptime': 'Total time from the start; after it answers are refused and the result is saved.',
    'hint.sendAnswers': 'Works only when "send answers with questions" is allowed in the settings.',
    'hint.passCorrect': 'Wins over the percent when set, e.g. 9 of 10.', 'hint.passPercent': 'Empty = settings default ({n}%).',
    'hint.packReward': 'Paid on a pass, once or every time per the settings. Currencies go to the wallet; economy prizes through economy.',
    'hint.stage': 'Passing this pack completes the attempt of this campaign stage on the server; the game can no longer send that stage\'s result itself.',
    'hint.missionTime': 'Counted from the player\'s first start. Empty = no time.', 'hint.missionReward': 'Paid once when every pack (or every mission of the game) is passed.',
    'hint.bank': 'A question without a pack stays in the bank; packs drawing its category use it.', 'hint.media': 'Image or video address', 'hint.category': 'For question banks and reports',
    'hint.options': 'Options are stored in this order and shuffled per player depending on the settings. At least one right option is required.',
    'hint.accepted': 'One accepted answer per line', 'hint.qtimeOwn': 'Empty = the pack time', 'hint.qkey': 'Optional; importing again with this key updates the question.',
    'hint.questionReward': 'For a right answer, once per player or every time per the settings. Paid with the quiz result.',
    'hint.hints': 'A player buys each hint once; the price is taken from the wallet and a resend never charges twice.',
    'hint.rsMission': 'Report numbers come from the packs of this mission or the missions of this game.', 'hint.webhook': 'Optional; public http or https only. Each report is queued for delivery.',
    'hint.secret': 'An HMAC-SHA256 signature with this key is sent in X-Quiz-Report-Signature. Leave empty to keep the current key.',
    'hint.importPack': 'When chosen, every question goes into this pack.', 'hint.importLang': 'Language of spreadsheet rows (rows)',
    'pk.new': 'New pack', 'pk.noMission': 'No mission', 'pk.questions': 'Questions', 'pk.pass': 'Pass rule', 'pk.time': 'Time', 'pk.stageTag': 'Stage: {k}',
    'pk.fromBank': '{n} from bank {c}', 'pk.passCorrect': '{n} right answers', 'pk.passPercent': '{n}%', 'pk.asked': '{n} per play', 'pk.perQuestion': '{n} s per question', 'pk.whole': 'total {n} s',
    'pk.delete': 'Delete pack {key}?', 'pk.inUse': 'This pack was played. Switching it off is better. Delete anyway (player results are removed too)?',
    'pk.play': 'Play', 'pk.count': 'Questions per play', 'pk.pool': 'Draw from category bank', 'pk.qtime': 'Time per question (s)', 'pk.ptime': 'Total pack time (s)',
    'pk.shuffleQ': 'Shuffle questions', 'pk.shuffleO': 'Shuffle options', 'pk.sendAnswers': 'Send the right answers with the questions', 'pk.passCorrectF': 'Right answers needed', 'pk.passPercentF': 'Pass percent',
    'pk.reward': 'Pass reward', 'pk.campaign': 'Campaign link', 'pk.stage': 'Campaign stage (once)', 'pk.noStage': 'No stage', 'pk.openQuestions': 'Questions of this pack',
    'pk.s.players': '{n} players', 'pk.s.plays': '{n} plays', 'pk.s.wins': '{n} passes', 'pk.s.rate': 'pass rate {n}%', 'pk.s.best': 'avg best score {n}',
    'qs.new': 'New question', 'qs.edit': 'Edit question', 'qs.bank': 'Bank (no pack)', 'qs.type': 'Type', 'qs.type.options': 'Options', 'qs.type.text': 'Text answer',
    'qs.reportedOnly': 'Reported questions only', 'qs.question': 'Question', 'qs.answer': 'Right answer', 'qs.hints': 'Hints', 'qs.reports': '{n} open reports',
    'qs.delete': 'Delete this question?', 'qs.inUse': 'Players answered this question. Switching it off is better. Delete anyway?',
    'qs.option': 'Option', 'qs.optionN': 'Option {n}', 'qs.correct': 'Right', 'qs.media': 'Image address', 'qs.mediaF': 'Question media', 'qs.addOption': 'New option', 'qs.options': 'Options',
    'qs.accepted': 'Accepted answers', 'qs.explanation': 'Explanation after answering', 'qs.time': 'Time of this question (s)', 'qs.reward': 'Right answer reward',
    'qs.s.answers': '{n} answers', 'qs.s.correct': '{n}% right',
    'hn.title': 'Hints', 'hn.add': 'New hint', 'hn.afterSave': 'Save the question to add hints.', 'hn.free': 'Free', 'hn.bought': '{n} bought',
    'hn.delete': 'Delete this hint? Past purchases are removed too.', 'hn.text': 'Hint text', 'hn.currency': 'Price currency', 'hn.amount': 'Price',
    'ms.new': 'New mission', 'ms.newGame': 'New game', 'ms.kind': 'Kind', 'ms.kind.mission': 'Mission', 'ms.kind.game': 'Game', 'ms.game': 'Game', 'ms.noGame': 'No game',
    'ms.order': 'Order', 'ms.order.all_at_once': 'All at once', 'ms.order.sequential': 'In order', 'ms.time': 'Time (s)', 'ms.timeF': 'Time from start (s)', 'ms.items': 'Content',
    'ms.packs': '{n} packs', 'ms.missions': '{n} missions', 'ms.reward': 'Finish reward', 'ms.delete': 'Delete {key}?', 'ms.inUse': 'Packs or missions still belong to it.',
    'ms.empty': 'No missions yet. Packs can be played without one.',
    'ms.hint.mission': 'Link packs to this mission on the packs page.', 'ms.hint.game': 'Link missions to this game in the mission form.',
    'ms.orderHint.mission': 'In order: each pack opens after the previous one is passed.', 'ms.orderHint.game': 'In order: each mission opens after the previous one is finished.',
    'rp.reasons': 'Report reasons', 'rp.reasonsHint': 'Players pick one of the active reasons when they report.', 'rp.status': 'Status', 'rp.reason': 'Reason', 'rp.openCount': '{n} open reports',
    'rp.date': 'Time', 'rp.player': 'Player', 'rp.note': 'Player note', 'rp.review': 'Review', 'rp.empty': 'No reports.', 'rp.same': 'Every open report of this question ({n})',
    'rp.disable': 'Switch the question off', 'rp.editQuestion': 'Open the question to fix it', 'rp.resolution': 'Review result', 'rp.do.resolved': 'Resolved', 'rp.do.dismissed': 'Dismiss', 'rp.do.open': 'Reopen',
    'rp.newReason': 'New reason', 'rp.reasonInUse': 'Reports use this reason; switch it off instead.',
    'st.open': 'Open', 'st.resolved': 'Resolved', 'st.dismissed': 'Dismissed', 'st.done': 'Done', 'st.paying': 'Paying', 'st.closed': 'Closed', 'st.pending': 'Queued', 'st.sent': 'Sent', 'st.failed': 'Failed',
    'rs.new': 'New report', 'rs.mission': 'Mission or game', 'rs.webhook': 'Webhook', 'rs.deliveries': 'Deliveries', 'rs.delete': 'Delete report {key}?', 'rs.empty': 'No result reports yet.',
    'rs.body': 'Report text', 'rs.placeholders': 'Tap a placeholder to insert it:', 'rs.insert': 'Insert into the text', 'rs.won': 'Pass text', 'rs.lost': 'Fail text', 'rs.url': 'Webhook address',
    'rs.secret': 'Signing key', 'rs.secretKept': 'The current key is kept', 'rs.noSecrets': 'ECONOMY_SECRET_KEY is not set; a signing key cannot be stored.', 'rs.attempts': 'Attempts', 'rs.error': 'Error',
    'im.title': 'Import', 'im.help': 'Format: {"missions":[…],"packs":[…],"questions":[…],"rows":[…]}. A question with a known key is updated. rows takes the legacy spreadsheet columns (Pack Key, Question, Option 1–6, Correct Options, Accepted Answers, Answer Language, Difficulty, Time). See "Sample".',
    'im.pack': 'Target pack', 'im.packFromFile': 'From the file (packKey)', 'im.lang': 'Row language', 'im.file': 'JSON file', 'im.json': 'JSON text', 'im.sample': 'Sample', 'im.dry': 'Check only', 'im.go': 'Import',
    'im.dryOk': 'The file is valid.', 'im.done': 'Imported.', 'im.count.missions': 'missions: {c} new, {u} updated', 'im.count.packs': 'packs: {c} new, {u} updated', 'im.count.questions': 'questions: {c} new, {u} updated',
    'im.section': 'Section', 'im.row': 'Row', 'im.field': 'Field', 'im.error': 'Error', 'im.exportTitle': 'Export', 'im.exportHelp': 'The export uses the import format, with right answers and hints.',
    'im.exportPack': 'Pack', 'im.export': 'Download JSON', 'im.exported': '{n} questions in the file',
    'sx.days': 'Period', 'sx.d': '{n} days', 'sx.players': 'Players', 'sx.plays': 'Plays', 'sx.finished': 'Finished', 'sx.passes': 'Passes', 'sx.rate': 'Pass rate', 'sx.avg': 'Avg score',
    'sx.secs': 'Avg time (s)', 'sx.hardest': 'Hardest questions', 'sx.answers': 'Answers', 'sx.correctRate': 'Right %', 'sx.skipped': 'Skipped or late', 'sx.noHard': 'Not enough data yet.',
    'sx.kpi.players': 'Players', 'sx.kpi.plays': 'Plays', 'sx.kpi.finished': 'Finished', 'sx.kpi.rate': 'Pass rate', 'sx.kpi.hints': 'Hints bought', 'sx.kpi.reports': 'Open reports',
    'rw.currency': 'Currency', 'rw.amount': 'Amount', 'rw.addCurrency': 'Currency', 'rw.prize': 'Economy prize', 'rw.group': 'Prize group', 'rw.noPrize': 'No prize', 'rw.missing': 'not found',
    'se.noPrizes': 'The economy plugin is not loaded; economy prizes are not paid (currencies are).', 'se.noCampaign': 'The campaign plugin is not loaded; pack-stage links do nothing.',
    'se.play': 'Play', 'se.delivery': 'Question delivery', 'se.delivery.all': 'All at the start (legacy)', 'se.delivery.one_by_one': 'One by one, the next after an answer',
    'se.hint.delivery': 'One by one: each question\'s clock starts when it is sent and players cannot read ahead.',
    'se.shuffleQ': 'Shuffle questions (packs without their own choice)', 'se.shuffleO': 'Shuffle options', 'se.restart': 'Playing a pack whose session is open',
    'se.restart.resume': 'Resume that session', 'se.restart.discard': 'Close it without a reward', 'se.hours': 'Open session lifetime without a time limit (h)', 'se.hint.hours': 'Afterwards the session is scored with its answers.',
    'se.autoFinish': 'Save the result automatically after the last answer',
    'se.timing': 'Timing', 'se.qsec': 'Default time per question (s)', 'se.hint.qsec': '0 = no time.', 'se.grace': 'Network grace (s)', 'se.late': 'An answer after the question time',
    'se.late.wrong': 'Counts as wrong (late)', 'se.late.reject': 'Is refused',
    'se.scoring': 'Passing and checking', 'se.pass': 'Default pass percent', 'se.hint.pass': 'A pack can have its own count or percent.', 'se.multi': 'Questions with several right options',
    'se.multi.any': 'Picking one right option is enough', 'se.multi.all': 'Every right option and nothing else', 'se.match': 'Text answer matching',
    'se.match.normalized': 'Normalized (Arabic/Persian letters, digits, spaces and half spaces)', 'se.match.exact': 'Exact (case does not matter)', 'se.hint.match': 'An answer in any defined language is accepted.',
    'se.reveal': 'Right answer reveal', 'se.revealMode': 'When the right answer is sent', 'se.reveal.after_answer': 'After each answer', 'se.reveal.after_finish': 'Only in the final result', 'se.reveal.never': 'Never',
    'se.allowBefore': 'Allow packs and missions marked "send answers" to send right answers with the questions', 'se.explanation': 'Send the explanation with the answer',
    'se.rewards': 'Rewards', 'se.packMode': 'Pack pass reward', 'se.packMode.first_pass': 'First pass only', 'se.packMode.every_pass': 'Every pass', 'se.packMode.none': 'None',
    'se.qMode': 'Right answer reward', 'se.qMode.first_correct': 'Once per question', 'se.qMode.every_correct': 'Every time', 'se.qMode.none': 'None',
    'se.groupMode': 'Prize group payout', 'se.groupMode.draw': 'One prize drawn by weight', 'se.groupMode.all': 'Every prize of the group', 'se.zone': 'Day start time zone (daily limits)',
    'se.limits': 'Limits', 'se.dailyPlays': 'Plays per player per day', 'se.dailyPasses': 'Rewarded passes per day', 'se.hint.dailyPasses': 'After the cap a pass is recorded but pays nothing. 0 = no cap.',
    'se.maxPlays': 'Plays per pack', 'se.cooldown': 'Wait between two plays of a pack (s)', 'se.replay': 'A passed pack can be played again', 'se.hint.zero': '0 = no limit.',
    'se.hints': 'Hints', 'se.pricing': 'Hint price', 'se.pricing.per_hint': 'Each hint\'s own price', 'se.pricing.fixed': 'One fixed price', 'se.pricing.free': 'Free',
    'se.keep': 'A bought hint', 'se.keep.forever': 'Stays the player\'s for good', 'se.keep.session': 'Only in that session', 'se.fixedCur': 'Fixed price currency', 'se.fixedAmt': 'Fixed price', 'se.maxHints': 'Hints per question',
    'se.reports': 'Question reports', 'se.reportsOn': 'Players can report questions', 'se.requireSeen': 'Only questions the player was asked', 'se.noteRequired': 'A note is required',
    'se.repDaily': 'Reports per player per day', 'se.repAuto': 'Switch a question off at this many open reports', 'se.hint.repAuto': '0 = never.',
    'se.events': 'Events', 'se.hint.events': 'Economy player event keys (for tasks and achievements) and the analytics event name; empty = not sent.',
    'se.evCompleted': 'Quiz finished (value = score)', 'se.evPassed': 'Passed (value 1)', 'se.evCorrect': 'Right answers (value = count)', 'se.evAnalytics': 'Analytics event (pack key = context)', 'se.onFailure': 'Send for failed quizzes too',
    'se.campaign': 'Campaign', 'se.hint.campaign': 'A pack linked to a campaign stage starts the stage attempt when played and completes it on the server with the result.',
    'se.campaignOn': 'Pack-stage links are on', 'se.sendFailures': 'Send failed quizzes to the campaign as lost runs',
    'se.retention': 'Retention', 'se.keepDays': 'Keep finished sessions (days)', 'se.hint.zeroKeep': '0 = forever.', 'se.delDays': 'Keep webhook deliveries (days)', 'se.attempts': 'Webhook attempts',
    'pc.title': 'Quizzes', 'pc.noQuiz': 'The quiz plugin did not answer.', 'pc.summary': '{packs} packs played · {hints} hints bought · {reports} reports', 'pc.passed': 'Passed', 'pc.notPassed': 'Not passed yet',
    'pc.failed': 'Failed', 'pc.plays': '{p} plays, {w} passes, best {b}', 'pc.reset': 'Reset', 'pc.reset.confirm': 'Clear progress on "{name}"?',
    'pc.reset.rewards': 'Forget the paid reward too (the next pass pays again)', 'pc.none': 'No quiz played yet.', 'pc.missions': 'Missions', 'pc.completed': 'Finished',
    'pc.started': 'started {at}', 'pc.rewarded': 'Rewarded', 'pc.sessions': 'Recent sessions', 'pc.score': '{c} of {n} · score {s}',
    'code.INVALID_KEY': 'Invalid key: start with an English letter or digit; only letters, digits and _ . : -.', 'code.ALREADY_EXISTS': 'This key already exists.', 'code.IN_USE': 'In use.',
    'code.VALUE_OUT_OF_RANGE': 'The value is out of range.', 'code.INVALID_TIMEZONE': 'Unknown time zone.', 'code.INVALID_POLICY': 'One of the choices is not valid.',
    'code.DEFAULT_LANGUAGE_REQUIRED': 'A text in the default language is required.', 'code.INVALID_JSON': 'Invalid JSON.', 'code.INVALID_PRIZE_KEY': 'Invalid prize key.', 'code.ONE_PRIZE_ONLY': 'A prize or a prize group, not both.',
    'code.QUIZ_OPTIONS_REQUIRED': 'At least two options are required.', 'code.QUIZ_CORRECT_OPTION_REQUIRED': 'Mark at least one option as right.', 'code.QUIZ_OPTION_TEXT_REQUIRED': 'Each option needs a text or an image.',
    'code.DUPLICATE_OPTION': 'Duplicate option ids.', 'code.TOO_MANY_OPTIONS': 'At most 12 options.', 'code.QUIZ_TEXT_ANSWERS_REQUIRED': 'Write at least one accepted answer.',
    'code.QUIZ_QUESTION_TEXT_OR_CONTENT_REQUIRED': 'A question text or media is required.', 'code.QUIZ_TYPE_IS_INVALID': 'Invalid question type.', 'code.QUIZ_TYPE_INVALID': 'Invalid question type.',
    'code.QUIZ_PACK_NOT_FOUND': 'Pack not found.', 'code.QUIZ_MISSION_NOT_FOUND': 'Mission not found.', 'code.QUIZ_GAME_NOT_FOUND': 'Game not found.', 'code.STAGE_NOT_FOUND': 'Campaign stage not found.',
    'code.QUIZ_PACK_WIN_THRESHOLD_INVALID': 'Invalid pass rule (more than the questions asked, or a percent outside 0–100).', 'code.QUIZ_MISSION_ORDER_INVALID': 'Invalid order.',
    'code.QUIZ_HINT_PRICE_INVALID': 'Invalid hint price.', 'code.QUIZ_HINT_PRICE_REQUIRES_CURRENCY_ID': 'Pick a currency for the price.', 'code.QUIZ_HINT_VALUE_REQUIRED': 'A hint text or image is required.',
    'code.QUIZ_REPORT_TEMPLATE_NAME_IS_REQUIRED': 'A report name is required.', 'code.QUIZ_REPORT_TEMPLATE_BODY_IS_REQUIRED': 'A report text is required.',
    'code.QUIZ_REPORT_WEBHOOK_URL_IS_INVALID': 'The webhook must be a public http or https address (not internal or localhost).', 'code.QUIZ_SECRET_UNAVAILABLE': 'ECONOMY_SECRET_KEY is not set; a signing key cannot be stored.',
    'code.QUIZ_IMPORT_HAS_ERRORS': 'The file has errors; nothing was imported.', 'code.QUIZ_IMPORT_FILE_IS_REQUIRED': 'The file is empty.', 'code.QUIZ_IMPORT_TOO_LARGE': 'The file is too large (at most 5000 questions).',
    'code.QUIZ_OPTION_INDEX_NOT_FOUND': 'The right option number is outside the options.', 'code.DUPLICATE_KEY': 'Duplicate key in the file.', 'code.INVALID_LANGUAGE': 'Invalid language code.',
    'code.TEXT_TOO_LONG': 'The text is too long.', 'code.TOO_MANY_TAGS': 'At most 20 tags.', 'code.INVALID_STATUS': 'Invalid status.', 'code.INVALID_KIND': 'Invalid kind.'
  }
};

function qzLang() {
  var nc = window.NakamaConsole;
  var l = nc && nc.lang ? nc.lang() : (document.documentElement.getAttribute('lang') || 'fa');
  return /^en/i.test(l) ? 'en' : 'fa';
}

function has(key) {
  return (QZ_TEXT[qzLang()] || QZ_TEXT.fa)[key] != null;
}

function t(key, vars) {
  var dict = QZ_TEXT[qzLang()] || QZ_TEXT.fa;
  var s = dict[key] != null ? dict[key] : (QZ_TEXT.en[key] != null ? QZ_TEXT.en[key] : key);
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] != null ? digits(String(vars[k])) : m; });
  return s;
}

function digits(s) {
  if (qzLang() !== 'fa') return s;
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
    return d.toLocaleString(qzLang() === 'fa' ? 'fa-IR' : 'en-GB', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

function tr(key, vars, raw) {
  var s = t(key, vars);
  Object.keys(raw || {}).forEach(function (k) { s = s.split('{' + k + '}').join(String(raw[k])); });
  return s;
}

var QZ_API = '/v2/console/base';
var QZ_INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 px-3 py-2 text-sm';
var QZ_BTN = 'qz-btn inline-flex items-center justify-center gap-1.5 font-medium whitespace-nowrap rounded-md border text-sm py-1.5 px-3';

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

function QzError(status, code, message, field, meta) {
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
  return fetch((path.indexOf('/v2/') === 0 ? '' : QZ_API) + path, init).then(function (res) {
    if (res.status === 401 && !/^#\/login/.test(location.hash)) location.hash = '#/login';
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) throw new QzError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', data && data.field || '', data && data.meta);
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
  var b = h('button', { type: opts.type || 'button', class: QZ_BTN + ' qz-btn--' + (opts.variant || 'ghost') + (opts.class ? ' ' + opts.class : ''), disabled: !!opts.disabled, title: opts.title, 'aria-label': opts.title, onclick: opts.onclick },
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

var qzToastBox = null;
function toast(kind, text) {
  if (!qzToastBox || !qzToastBox.isConnected) {
    qzToastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-quiz': 'toasts' });
    document.body.appendChild(qzToastBox);
  }
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  qzToastBox.appendChild(el);
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
  var err = h('p', { class: 'qz-field-error text-xs text-error-600 dark:text-error-400', hidden: true });
  var wrap = h('label', { class: 'qz-field block mb-4' + (extra ? ' ' + extra : '') },
    label ? h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1', text: label }) : null,
    control,
    hint ? h('span', { class: 'block text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: hint }) : null,
    err);
  wrap.setError = function (msg) { err.textContent = msg || ''; err.hidden = !msg; };
  return wrap;
}

function input(value, opts) {
  opts = opts || {};
  var el = h(opts.multiline ? 'textarea' : 'input', { class: QZ_INPUT + (opts.mono ? ' font-mono' : ''), type: opts.multiline ? null : (opts.type || 'text'), dir: opts.ltr ? 'ltr' : 'auto', rows: opts.multiline ? (opts.rows || 3) : null, placeholder: opts.placeholder, name: opts.name, autocomplete: 'off', 'aria-label': opts.label });
  el.value = value == null ? '' : value;
  return el;
}

function select(value, options, opts) {
  opts = opts || {};
  var el = h('select', { class: QZ_INPUT + ' eco-select', name: opts.name, 'aria-label': opts.label });
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) opt.selected = true;
    el.appendChild(opt);
  });
  return el;
}

function check(checked, label, name) {
  var box = h('input', { type: 'checkbox', class: 'qz-check', name: name, checked: !!checked });
  var wrap = h('label', { class: 'qz-checkline flex items-start gap-2 mb-3 cursor-pointer text-sm text-secondary-800 dark:text-secondary-200' }, box, h('span', { text: label }));
  wrap.input = box;
  return wrap;
}

function numberValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? Math.round(n) : NaN;
}

var qzPanel = null;
function closePanel() {
  if (qzPanel) { qzPanel.remove(); qzPanel = null; document.removeEventListener('keydown', panelKey); }
}

function panelKey(e) { if (e.key === 'Escape') closePanel(); }

function sidePanel(title, body, footer, name) {
  closePanel();
  var holder = h('div', { class: 'qz-panel-holder' });
  var panel = h('div', { class: 'qz-panel bg-white dark:bg-secondary-900 shadow-lg', role: 'dialog', 'aria-modal': 'true', 'aria-label': title, 'data-quiz-panel': name || '' },
    h('div', { class: 'qz-panel-head flex items-center justify-between border-b px-6 py-3' },
      h('h2', { class: 'text-lg font-semibold truncate', text: title }),
      button({ label: '×', title: t('close'), onclick: closePanel, class: 'qz-close' })),
    h('div', { class: 'qz-panel-body p-6' }, body),
    footer ? h('div', { class: 'qz-panel-foot flex gap-2 justify-end bg-secondary-50 dark:bg-secondary-800 p-4' }, footer) : null);
  holder.appendChild(panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) closePanel(); });
  qzPanel = h('div', { 'data-quiz': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, h('div', { class: 'qz-shade' }), holder);
  document.body.appendChild(qzPanel);
  document.addEventListener('keydown', panelKey);
  return panel;
}

function confirmBox(text, okLabel) {
  return new Promise(function (resolve) {
    var root;
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-quiz': 'confirm', 'data-l10n-skip': '', class: 'qz-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'qz-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('confirm.title') }),
        h('p', { class: 'text-sm mb-4', text: text }),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: okLabel || t('yes'), variant: 'danger', data: { 'quiz-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root);
  });
}

function pageFrame(page, desc, actions) {
  var content = h('div', { class: 'flex flex-col gap-4' });
  var el = h('div', { class: 'qz-page flex flex-col flex-1 px-4 md:px-6 pb-6 max-w-[1600px] w-full mx-auto', 'data-quiz-page': page },
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
    var tr = h('tr', { class: 'border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' + (onrow ? ' cursor-pointer' : ''), 'data-quiz-row': row.id || '' });
    if (onrow) {
      tr.tabIndex = 0;
      tr.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,select,label')) onrow(row.item); });
      tr.addEventListener('keydown', function (e) { if (e.key === 'Enter') onrow(row.item); });
    }
    row.cells.forEach(function (cell, i) { tr.appendChild(h('td', { class: 'px-3 py-2 align-middle text-sm', 'data-label': columns[i].label }, cell)); });
    tbody.appendChild(tr);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow rounded-md overflow-x-auto' }, h('table', { class: 'w-full eco-table qz-table' }, thead, tbody));
}

function pager(state, onpage) {
  var total = state.total || 0;
  var from = total ? (state.page - 1) * state.limit + 1 : 0;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  return h('div', { class: 'flex items-center justify-end gap-2 text-xs text-secondary-500' },
    h('span', { 'data-quiz-range': '', text: total ? t('range', { from: from, to: to, total: total }) : t('empty') }),
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

var qzLanguages = null;
function loadLanguages() {
  if (qzLanguages) return Promise.resolve(qzLanguages);
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch('/v2/console/base/languages', { headers: headers, credentials: 'same-origin' }).then(function (r) { return r.ok ? r.json() : null; }).then(function (d) {
    var list = d && d.items ? d.items.filter(function (l) { return l.isActive; }).map(function (l) { return { code: l.code, name: l.nativeName || l.code, isDefault: !!l.isDefault, direction: l.direction }; }) : [];
    if (!list.length) list = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    list.sort(function (a, b) { return (b.isDefault ? 1 : 0) - (a.isDefault ? 1 : 0); });
    qzLanguages = list;
    return list;
  }, function () {
    qzLanguages = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    return qzLanguages;
  });
}

function translated(label, name, value, opts) {
  opts = opts || {};
  var inputs = {};
  var box = h('div', { class: 'grid sm:grid-cols-2 gap-x-4', 'data-quiz-text': name });
  (qzLanguages || []).forEach(function (l) {
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
  var l = qzLang();
  return m[l] || m.fa || m.en || Object.keys(m).map(function (k) { return m[k]; }).filter(Boolean)[0] || fallback || '';
}

function fieldError(msgBox, err) {
  clear(msgBox).appendChild(notice('error', errorText(err) + (err && err.field ? ' · ' + err.field : '')));
}

function options(list, value) {
  return list.map(function (o) { return typeof o === 'string' ? { value: o, label: o } : o; });
}

function dateInput(value, name, type) {
  var el = h('input', { type: type || 'date', class: QZ_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  el.value = value || '';
  return el;
}

function isoDay(d) {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

function qzErrText(err) {
  if (err && err.code && has('code.' + err.code)) return t('code.' + err.code) + (err.field ? ' · ' + err.field : '');
  return errorText(err) + (err && err.field ? ' · ' + err.field : '');
}

function section(title, children, name) {
  return card(h('div', { class: 'p-4', 'data-quiz-section': name || '' },
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
    var msg = h('div', { 'data-quiz-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', 'data-quiz-updated': '' });
    var paintMeta = function (x) { meta.textContent = x.updatedBy ? t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }) : ''; };
    paintMeta(d);
    var save = button({ label: t('save'), variant: 'primary', data: { 'quiz-save': page }, onclick: function () {
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
        msg.appendChild(notice('error', qzErrText(err)));
      });
    } });
    clear(frame.content);
    form.el.forEach(function (el) { frame.content.appendChild(el); });
    frame.content.appendChild(h('div', { class: 'qz-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), save));
  }, function (err) {
    clear(frame.content).appendChild(notice('error', qzErrText(err)));
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
  return h('a', { href: '#/players/' + id, class: 'eco-link', 'data-quiz-player-link': id, dir: 'auto', text: name || id.slice(0, 8) });
}

var qzRefs = null;
function loadRefs(force) {
  if (qzRefs && !force) return Promise.resolve(qzRefs);
  return api('GET', '/quiz/refs' + qs({ lang: qzLang() })).then(function (r) { qzRefs = r || {}; return qzRefs; }, function () { qzRefs = {}; return qzRefs; });
}

var qzSettingsCache = null;
function loadQzSettings() {
  if (qzSettingsCache) return Promise.resolve(qzSettingsCache);
  return api('GET', '/quiz/settings').then(function (d) { qzSettingsCache = d; return d; });
}

function refOptions(kind, value, emptyLabel) {
  var list = (qzRefs && qzRefs[kind]) || [];
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
  var box = h('div', { class: 'qz-rows', 'data-quiz-currencies': name || '' });
  var rows = [];
  function addRow(k, v) {
    var key = refSelect('currencies', k || '', 'reward-currency', '—');
    var amt = numInput(v == null ? '' : v, 'reward-amount', t('rw.amount'));
    var row = h('div', { class: 'qz-row' }, field(t('rw.currency'), key), field(t('rw.amount'), amt),
      h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r.row !== row; }); } })));
    rows.push({ row: row, key: key, amt: amt });
    box.appendChild(row);
  }
  Object.keys(values || {}).forEach(function (k) { addRow(k, values[k]); });
  var wrap = h('div', {}, box, button({ label: '+ ' + t('rw.addCurrency'), data: { 'quiz-action': 'add-currency-' + (name || '') }, onclick: function () { addRow(); } }));
  wrap.read = function () {
    var out = {};
    rows.forEach(function (r) { var k = r.key.value; var n = numberValue(r.amt); if (k && n) out[k] = n; });
    return out;
  };
  return wrap;
}

function rewardEditor(value, name) {
  value = value || {};
  var cur = currencyRows(value.currencies, name);
  var prize = refSelect('prizes', value.prizeKey, name + '-prize', t('rw.noPrize'));
  var group = refSelect('prizeGroups', value.prizeGroupKey, name + '-prize-group', t('rw.noPrize'));
  var el = h('div', { 'data-quiz-reward': name }, cur, h('div', { class: 'grid sm:grid-cols-2 gap-x-4 mt-3' }, field(t('rw.prize'), prize), field(t('rw.group'), group)));
  el.read = function () {
    var o = { currencies: cur.read() };
    if (prize.value) o.prizeKey = prize.value;
    if (group.value) o.prizeGroupKey = group.value;
    return o;
  };
  return el;
}

function rewardText(r) {
  r = r || {};
  var cur = r.currencies || {};
  var parts = Object.keys(cur).map(function (k) { return fmtNum(cur[k]) + ' ' + k; });
  if (r.prizeKey) parts.push('🎁 ' + r.prizeKey);
  if (r.prizeGroupKey) parts.push('🎁 ' + r.prizeGroupKey);
  return parts.length ? '⁦' + parts.join(' + ') + '⁩' : t('none');
}

function rewardListText(list) {
  if (!list || !list.length) return t('none');
  return '⁦' + list.map(function (r) { return fmtNum(r.amount) + ' ' + r.currencyKey; }).join(' + ') + '⁩';
}

function tagsInput(list, name) {
  var el = input((list || []).join('، '), { name: name, placeholder: t('hint.tags') });
  el.read = function () { return el.value.split(/[,،\n]/).map(function (s) { return s.trim(); }).filter(Boolean); };
  return el;
}

function triSelect(value, name) {
  var el = select(value === true ? 'yes' : value === false ? 'no' : '', [{ value: '', label: t('default') }, { value: 'yes', label: t('yes') }, { value: 'no', label: t('no') }], { name: name, label: name });
  el.read = function () { return el.value === 'yes' ? true : el.value === 'no' ? false : null; };
  return el;
}

function jsonInput(value, name) {
  var el = input(JSON.stringify(value || {}), { name: name, ltr: true, mono: true, multiline: true, rows: 2 });
  el.read = function () { try { return JSON.parse(el.value || '{}'); } catch (e) { return undefined; } };
  return el;
}

function optNumber(el) {
  var n = floatValue(el);
  return n == null || isNaN(n) ? null : n;
}

function statusTag(s) {
  return tag(t('st.' + s), s === 'open' ? 'warning' : s === 'resolved' || s === 'done' || s === 'sent' ? 'success' : s === 'failed' ? 'error' : 'secondary');
}

function passText(p) {
  if (p.passCorrect != null) return t('pk.passCorrect', { n: p.passCorrect });
  if (p.passPercent != null) return t('pk.passPercent', { n: p.passPercent });
  return t('default');
}

function packsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 50, total: 0, missionId: q.missionId || '', q: '' };
  var missions = [];
  var frame = pageFrame('packs', t('desc.packs'), [
    button({ label: t('pk.new'), variant: 'primary', data: { 'quiz-action': 'new-pack' }, onclick: function () { packEditor(null, missions, reload); } })]);
  clear(root).appendChild(frame.el);
  var missionSel = h('select', { class: QZ_INPUT + ' eco-select', name: 'missionId', 'aria-label': t('f.mission') });
  var search = input('', { name: 'q', placeholder: t('search') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end qz-filters' }, field(t('f.mission'), missionSel), field(t('search'), search)));
  frame.content.appendChild(listBox);
  missionSel.addEventListener('change', function () { state.missionId = missionSel.value; state.page = 1; reload(); });
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.q = search.value.trim(); state.page = 1; reload(); }, 300); });

  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/missions?kind=mission'), api('GET', '/quiz/packs' + qs({ missionId: state.missionId, q: state.q, page: state.page, limit: state.limit })), loadLanguages(), loadRefs(true), loadQzSettings()]).then(function (res) {
      missions = res[0].items || [];
      clear(missionSel);
      [{ value: '', label: t('all') }, { value: 'none', label: t('pk.noMission') }].concat(missions.map(function (m) { return { value: m.id, label: textOf(m.name, m.key) }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.missionId) opt.selected = true;
        missionSel.appendChild(opt);
      });
      var d = res[1];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'packs' }, table([
        { label: t('f.name') }, { label: t('f.mission') }, { label: t('pk.questions') }, { label: t('pk.pass') }, { label: t('pk.time') }, { label: t('f.reward') }, { label: '' }
      ], items.map(function (p) {
        return { id: p.key, item: p, cells: [
          h('div', { class: 'flex flex-col' }, h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(p.name, p.key) }), p.isActive ? null : tag(t('inactive'), 'warning'),
            p.campaignStageKey ? tag(t('pk.stageTag', { k: p.campaignStageKey }), 'primary') : null), mono(p.key)),
          h('span', { text: p.missionKey || t('none') }),
          h('a', { class: 'eco-link', href: '#/quiz/questions?packId=' + p.id, 'data-quiz-questions-link': p.key, text: p.poolCategory ? t('pk.fromBank', { n: fmtNum(p.questions), c: p.poolCategory }) : fmtNum(p.questions) }),
          h('span', { text: passText(p) + (p.questionCount ? ' · ' + t('pk.asked', { n: p.questionCount }) : '') }),
          h('span', { text: (p.questionSeconds ? t('pk.perQuestion', { n: p.questionSeconds }) : '') + (p.timeSeconds ? ' ' + t('pk.whole', { n: p.timeSeconds }) : '') || t('none') }),
          h('span', { class: 'text-xs', text: rewardText(p.reward) }),
          button({ label: t('delete'), variant: 'danger', data: { 'quiz-delete': p.key }, onclick: function () { deletePack(p, reload); } })
        ] };
      }), function (p) { packEditor(p, missions, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function deletePack(p, done) {
  confirmBox(t('pk.delete', { key: p.key }), t('delete')).then(function (ok) {
    if (!ok) return;
    api('DELETE', '/quiz/packs/' + p.id).then(function () { toast('success', t('saved')); done(); }, function (err) {
      if (err.code !== 'IN_USE') { toast('error', qzErrText(err)); return; }
      confirmBox(t('pk.inUse'), t('delete')).then(function (yes) {
        if (yes) api('DELETE', '/quiz/packs/' + p.id + '?force=true').then(function () { toast('success', t('saved')); done(); }, function (e2) { toast('error', qzErrText(e2)); });
      });
    });
  });
}

function packEditor(p, missions, done) {
  var creating = !p;
  var show = function (stats) {
    p = p || { name: {}, description: {}, isActive: true, reward: {}, tags: [], metadata: {} };
    var key = input(p.key || '', { name: 'key', ltr: true, mono: true });
    if (!creating) key.disabled = true;
    var mission = select(p.missionId || '', [{ value: '', label: t('pk.noMission') }].concat(missions.map(function (m) { return { value: m.id, label: textOf(m.name, m.key) }; })), { name: 'missionId', label: t('f.mission') });
    var name = translated(t('f.name'), 'name', p.name);
    var desc = translated(t('f.description'), 'description', p.description, { multiline: true });
    var category = input(p.category, { name: 'category' });
    var tags = tagsInput(p.tags, 'tags');
    var lang = input(p.language, { name: 'language', ltr: true, placeholder: 'fa' });
    var difficulty = numInput(p.difficulty || '', 'difficulty');
    var icon = input(p.iconUrl, { name: 'iconUrl', ltr: true });
    var sort = numInput(p.sort || 0, 'sort');
    var time = numInput(p.timeSeconds || '', 'timeSeconds');
    var qtime = numInput(p.questionSeconds || '', 'questionSeconds');
    var pct = numInput(p.passPercent == null ? '' : p.passPercent, 'passPercent');
    var correct = numInput(p.passCorrect == null ? '' : p.passCorrect, 'passCorrect');
    var count = numInput(p.questionCount || '', 'questionCount');
    var pool = input(p.poolCategory, { name: 'poolCategory' });
    var shuffleQ = triSelect(p.shuffleQuestions, 'shuffleQuestions');
    var shuffleO = triSelect(p.shuffleOptions, 'shuffleOptions');
    var send = check(p.sendAnswers, t('pk.sendAnswers'), 'sendAnswers');
    var reward = rewardEditor(p.reward, 'pack');
    var stage = refSelect('stages', p.campaignStageKey, 'campaignStageKey', t('pk.noStage'));
    var active = check(p.isActive, t('f.active'), 'isActive');
    var meta = jsonInput(p.metadata, 'metadata');
    var msg = h('div', {});
    var cfg = qzSettingsCache && qzSettingsCache.settings;
    var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
    var statsBox = stats ? h('div', { class: 'flex flex-wrap gap-2 mb-4', 'data-quiz-pack-stats': '' },
      tag(t('pk.s.players', { n: fmtNum(stats.players) })), tag(t('pk.s.plays', { n: fmtNum(stats.plays) })), tag(t('pk.s.wins', { n: fmtNum(stats.wins) }), 'success'),
      tag(t('pk.s.rate', { n: fmtNum(stats.winRate) })), tag(t('pk.s.best', { n: fmtNum(stats.avgBestScore) }))) : null;
    var body = h('div', { 'data-quiz-form': 'pack' }, statsBox,
      grid(field(t('f.key'), key, t('hint.key')), field(t('f.mission'), mission)), name, desc,
      grid(field(t('f.category'), category), field(t('f.tags'), tags)),
      grid(field(t('f.language'), lang, t('hint.language')), field(t('f.difficulty'), difficulty)),
      grid(field(t('f.icon'), icon), field(t('f.sort'), sort)), active,
      section(t('pk.play'), h('div', {},
        grid(field(t('pk.count'), count, t('hint.count')), field(t('pk.pool'), pool, t('hint.pool'))),
        grid(field(t('pk.qtime'), qtime, cfg && cfg.timing.questionSeconds ? t('hint.qtimeDefault', { n: cfg.timing.questionSeconds }) : t('hint.qtime')), field(t('pk.ptime'), time, t('hint.ptime'))),
        grid(field(t('pk.shuffleQ'), shuffleQ), field(t('pk.shuffleO'), shuffleO)), send, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-2', text: t('hint.sendAnswers') })), 'play'),
      h('div', { class: 'h-3' }), section(t('pk.pass'), h('div', {}, grid(field(t('pk.passCorrectF'), correct, t('hint.passCorrect')), field(t('pk.passPercentF'), pct, cfg ? t('hint.passPercent', { n: cfg.scoring.passPercent }) : null))), 'pass'),
      h('div', { class: 'h-3' }), section(t('pk.reward'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.packReward') }), reward), 'reward'),
      h('div', { class: 'h-3' }), section(t('pk.campaign'), h('div', {}, field(t('pk.stage'), stage, t('hint.stage'))), 'campaign'),
      h('div', { class: 'h-3' }), field(t('f.metadata'), meta), msg);
    var save = button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-pack' }, onclick: function () {
      busy(save, true);
      clear(msg);
      var md = meta.read();
      if (md === undefined) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_JSON') + ' · metadata')); return; }
      var out = { key: key.value.trim(), missionId: mission.value, name: name.read(), description: desc.read(), category: category.value.trim(), tags: tags.read(), language: lang.value.trim(),
        difficulty: numberValue(difficulty) || 0, iconUrl: icon.value.trim(), sort: numberValue(sort) || 0, timeSeconds: numberValue(time) || 0, questionSeconds: numberValue(qtime) || 0,
        passPercent: optNumber(pct), passCorrect: optNumber(correct), questionCount: numberValue(count) || 0, poolCategory: pool.value.trim(), shuffleQuestions: shuffleQ.read(),
        shuffleOptions: shuffleO.read(), sendAnswers: send.input.checked, reward: reward.read(), campaignStageKey: stage.value, isActive: active.input.checked, metadata: md };
      api(creating ? 'POST' : 'PUT', '/quiz/packs' + (creating ? '' : '/' + p.id), out).then(function () {
        closePanel(); toast('success', t('saved')); qzRefs = null; done();
      }, function (err) { busy(save, false); msg.appendChild(notice('error', qzErrText(err))); });
    } });
    var footer = [button({ label: t('cancel'), onclick: closePanel }), save];
    if (!creating) footer.unshift(h('a', { class: QZ_BTN + ' qz-btn qz-btn--ghost', href: '#/quiz/questions?packId=' + p.id, onclick: closePanel, text: t('pk.openQuestions') }));
    sidePanel(creating ? t('pk.new') : textOf(p.name, p.key), body, footer, 'pack');
  };
  if (creating) { show(null); return; }
  api('GET', '/quiz/packs/' + p.id).then(function (d) { p = d.pack; show(d.stats); }, function (err) { toast('error', qzErrText(err)); });
}

function questionsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 50, total: 0, packId: q.packId || '', type: '', category: '', q: '', reported: q.reported || '', active: '' };
  var packs = [];
  var frame = pageFrame('questions', t('desc.questions'), [
    button({ label: t('qs.new'), variant: 'primary', data: { 'quiz-action': 'new-question' }, onclick: function () { questionEditor(null, packs, reload, state.packId); } })]);
  clear(root).appendChild(frame.el);
  var packSel = h('select', { class: QZ_INPUT + ' eco-select', name: 'packId', 'aria-label': t('f.pack') });
  var typeSel = select('', [{ value: '', label: t('all') }, { value: 'options', label: t('qs.type.options') }, { value: 'text', label: t('qs.type.text') }], { name: 'type', label: t('qs.type') });
  var catSel = h('select', { class: QZ_INPUT + ' eco-select', name: 'category', 'aria-label': t('f.category') });
  var activeSel = select('', [{ value: '', label: t('all') }, { value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }], { name: 'active', label: t('f.active') });
  var reported = check(state.reported === 'true', t('qs.reportedOnly'), 'reported');
  var search = input('', { name: 'q', placeholder: t('search') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end qz-filters' }, field(t('f.pack'), packSel), field(t('qs.type'), typeSel), field(t('f.category'), catSel),
    field(t('f.active'), activeSel), field(t('search'), search), reported));
  frame.content.appendChild(listBox);
  packSel.addEventListener('change', function () { state.packId = packSel.value; state.page = 1; reload(); });
  typeSel.addEventListener('change', function () { state.type = typeSel.value; state.page = 1; reload(); });
  catSel.addEventListener('change', function () { state.category = catSel.value; state.page = 1; reload(); });
  activeSel.addEventListener('change', function () { state.active = activeSel.value; state.page = 1; reload(); });
  reported.input.addEventListener('change', function () { state.reported = reported.input.checked ? 'true' : ''; state.page = 1; reload(); });
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.q = search.value.trim(); state.page = 1; reload(); }, 300); });

  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/packs?limit=200'), api('GET', '/quiz/questions' + qs({ packId: state.packId, type: state.type, category: state.category, active: state.active, reported: state.reported, q: state.q, page: state.page, limit: state.limit })),
      loadLanguages(), loadRefs(true)]).then(function (res) {
      packs = res[0].items || [];
      clear(packSel);
      [{ value: '', label: t('all') }, { value: 'none', label: t('qs.bank') }].concat(packs.map(function (p) { return { value: p.id, label: textOf(p.name, p.key) }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.packId) opt.selected = true;
        packSel.appendChild(opt);
      });
      clear(catSel);
      [{ value: '', label: t('all') }].concat((qzRefs.categories || []).map(function (c) { return { value: c.key, label: c.key }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.category) opt.selected = true;
        catSel.appendChild(opt);
      });
      var d = res[1];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'questions' }, table([
        { label: t('qs.question') }, { label: t('f.pack') }, { label: t('qs.type') }, { label: t('f.difficulty') }, { label: t('qs.answer') }, { label: t('qs.hints') }, { label: '' }
      ], items.map(function (x) {
        var right = (x.options || []).filter(function (o) { return o.correct; }).map(function (o) { return textOf(o.text, '#' + o.id); });
        return { id: x.id, item: x, cells: [
          h('div', { class: 'flex flex-col gap-1' }, h('span', { class: 'font-medium qz-sentence', dir: 'auto', text: textOf(x.question, x.mediaUrl || x.id) }),
            h('div', { class: 'flex flex-wrap gap-1' }, x.key ? mono(x.key) : null, x.isActive ? null : tag(t('inactive'), 'warning'), x.openReports ? tag(t('qs.reports', { n: x.openReports }), 'error') : null,
              x.category ? tag(x.category) : null)),
          h('span', { text: x.packKey || t('qs.bank') }),
          h('span', { text: t('qs.type.' + x.type) }),
          h('span', { text: fmtNum(x.difficulty) }),
          h('span', { class: 'text-xs', dir: 'auto', text: x.type === 'options' ? right.join('، ') : Object.keys(x.answers || {}).map(function (l) { return (x.answers[l] || []).join(' | '); }).join(' · ') }),
          h('span', { text: fmtNum(x.hints) }),
          button({ label: t('delete'), variant: 'danger', data: { 'quiz-delete': x.id }, onclick: function () { deleteQuestion(x, reload); } })
        ] };
      }), function (x) { questionEditor(x, packs, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function deleteQuestion(x, done) {
  confirmBox(t('qs.delete'), t('delete')).then(function (ok) {
    if (!ok) return;
    api('DELETE', '/quiz/questions/' + x.id).then(function () { toast('success', t('saved')); done(); }, function (err) {
      if (err.code !== 'IN_USE') { toast('error', qzErrText(err)); return; }
      confirmBox(t('qs.inUse'), t('delete')).then(function (yes) {
        if (yes) api('DELETE', '/quiz/questions/' + x.id + '?force=true').then(function () { toast('success', t('saved')); done(); }, function (e2) { toast('error', qzErrText(e2)); });
      });
    });
  });
}

function optionEditor(list) {
  var box = h('div', { class: 'qz-rows', 'data-quiz-options': '' });
  var rows = [];
  var nextId = 0;
  (list || []).forEach(function (o) { if (o.id >= nextId) nextId = o.id + 1; });
  function addRow(o) {
    o = o || { id: nextId++, text: {}, correct: false };
    var texts = {};
    var right = check(o.correct, t('qs.correct'), 'option-correct');
    var media = input(o.mediaUrl, { name: 'option-media', ltr: true, placeholder: t('qs.media') });
    var inputs = (qzLanguages || []).map(function (l) {
      var el = input(o.text && o.text[l.code] || '', { name: 'option-' + l.code, label: l.name });
      el.setAttribute('dir', l.direction === 'rtl' ? 'rtl' : 'ltr');
      texts[l.code] = el;
      return field(t('lang.text', { label: t('qs.option'), lang: l.name }), el);
    });
    var row = h('div', { class: 'qz-track', 'data-quiz-option': String(o.id) },
      h('div', { class: 'flex items-center justify-between gap-2 mb-2' }, h('span', { class: 'text-xs text-secondary-500', text: t('qs.optionN', { n: rows.length + 1 }) }), right,
        button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r.row !== row; }); } })),
      h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, inputs), field(t('qs.media'), media));
    rows.push({ row: row, id: o.id, texts: texts, right: right, media: media });
    box.appendChild(row);
  }
  (list || []).forEach(addRow);
  var wrap = h('div', {}, box, button({ label: '+ ' + t('qs.addOption'), data: { 'quiz-action': 'add-option' }, onclick: function () { addRow(); } }));
  wrap.add = addRow;
  wrap.count = function () { return rows.length; };
  wrap.read = function () {
    return rows.map(function (r) {
      var txt = {};
      Object.keys(r.texts).forEach(function (k) { var v = r.texts[k].value.trim(); if (v) txt[k] = v; });
      return { id: r.id, text: txt, mediaUrl: r.media.value.trim(), correct: r.right.input.checked };
    });
  };
  return wrap;
}

function answerEditor(value) {
  var inputs = {};
  var box = h('div', { class: 'grid sm:grid-cols-2 gap-x-4', 'data-quiz-answers': '' });
  (qzLanguages || []).forEach(function (l) {
    var el = input(((value || {})[l.code] || []).join('\n'), { name: 'answers-' + l.code, multiline: true, rows: 3 });
    el.setAttribute('dir', l.direction === 'rtl' ? 'rtl' : 'ltr');
    inputs[l.code] = el;
    box.appendChild(field(t('lang.text', { label: t('qs.accepted'), lang: l.name }), el, t('hint.accepted')));
  });
  box.read = function () {
    var out = {};
    Object.keys(inputs).forEach(function (k) {
      var list = inputs[k].value.split('\n').map(function (s) { return s.trim(); }).filter(Boolean);
      if (list.length) out[k] = list;
    });
    return out;
  };
  return box;
}

function questionEditor(x, packs, done, packId) {
  var creating = !x;
  var show = function (stats) {
    x = x || { type: 'options', question: {}, explanation: {}, isActive: true, reward: {}, tags: [], packId: packId && packId !== 'none' ? packId : '', metadata: {},
      options: [{ id: 0, text: {}, correct: true }, { id: 1, text: {} }, { id: 2, text: {} }, { id: 3, text: {} }] };
    var key = input(x.key || '', { name: 'key', ltr: true, mono: true });
    var pack = select(x.packId || '', [{ value: '', label: t('qs.bank') }].concat(packs.map(function (p) { return { value: p.id, label: textOf(p.name, p.key) }; })), { name: 'packId', label: t('f.pack') });
    var type = select(x.type, [{ value: 'options', label: t('qs.type.options') }, { value: 'text', label: t('qs.type.text') }], { name: 'type', label: t('qs.type') });
    var text = translated(t('qs.question'), 'question', x.question, { multiline: true });
    var media = input(x.mediaUrl, { name: 'mediaUrl', ltr: true });
    var opts = optionEditor(x.options);
    var answers = answerEditor(x.answers);
    var expl = translated(t('qs.explanation'), 'explanation', x.explanation, { multiline: true });
    var difficulty = numInput(x.difficulty || '', 'difficulty');
    var category = input(x.category, { name: 'category' });
    var tags = tagsInput(x.tags, 'tags');
    var time = numInput(x.timeSeconds == null ? '' : x.timeSeconds, 'timeSeconds');
    var sort = numInput(x.sort || 0, 'sort');
    var active = check(x.isActive, t('f.active'), 'isActive');
    var reward = rewardEditor(x.reward, 'question');
    var meta = jsonInput(x.metadata, 'metadata');
    var optBox = section(t('qs.options'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.options') }), opts), 'options');
    var ansBox = section(t('qs.accepted'), answers, 'answers');
    var hintsBox = h('div', {});
    var msg = h('div', {});
    var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
    function paint() { optBox.hidden = type.value !== 'options'; ansBox.hidden = type.value !== 'text'; }
    type.addEventListener('change', paint);
    var statsBox = stats ? h('div', { class: 'flex flex-wrap gap-2 mb-4', 'data-quiz-question-stats': '' }, tag(t('qs.s.answers', { n: fmtNum(stats.answers) })),
      tag(t('qs.s.correct', { n: stats.answers ? fmtNum(Math.round(stats.correct * 1000 / stats.answers) / 10) : 0 }), 'success'), x.openReports ? tag(t('qs.reports', { n: x.openReports }), 'error') : null) : null;
    var body = h('div', { 'data-quiz-form': 'question' }, statsBox,
      grid(field(t('f.pack'), pack, t('hint.bank')), field(t('qs.type'), type)), text, field(t('qs.mediaF'), media, t('hint.media')),
      optBox, ansBox, h('div', { class: 'h-3' }), expl,
      grid(field(t('f.difficulty'), difficulty), field(t('qs.time'), time, t('hint.qtimeOwn'))),
      grid(field(t('f.category'), category, t('hint.category')), field(t('f.tags'), tags)),
      grid(field(t('f.key'), key, t('hint.qkey')), field(t('f.sort'), sort)), active,
      section(t('qs.reward'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.questionReward') }), reward), 'reward'),
      h('div', { class: 'h-3' }), hintsBox, h('div', { class: 'h-3' }), field(t('f.metadata'), meta), msg);
    paint();
    if (!creating) hintList(x.id, hintsBox);
    else hintsBox.appendChild(h('p', { class: 'text-xs text-secondary-500', text: t('hn.afterSave') }));
    var save = button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-question' }, onclick: function () {
      busy(save, true);
      clear(msg);
      var md = meta.read();
      if (md === undefined) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_JSON'))); return; }
      var tsec = numberValue(time);
      var out = { key: key.value.trim(), packId: pack.value, type: type.value, question: text.read(), mediaUrl: media.value.trim(), options: type.value === 'options' ? opts.read() : [],
        answers: type.value === 'text' ? answers.read() : {}, explanation: expl.read(), difficulty: numberValue(difficulty) || 0, category: category.value.trim(), tags: tags.read(),
        timeSeconds: tsec == null || isNaN(tsec) ? null : tsec, sort: numberValue(sort) || 0, isActive: active.input.checked, reward: reward.read(), metadata: md };
      api(creating ? 'POST' : 'PUT', '/quiz/questions' + (creating ? '' : '/' + x.id), out).then(function (d) {
        toast('success', t('saved'));
        done();
        if (creating) { closePanel(); questionEditor(d.question, packs, done); } else closePanel();
      }, function (err) { busy(save, false); msg.appendChild(notice('error', qzErrText(err))); });
    } });
    sidePanel(creating ? t('qs.new') : t('qs.edit'), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'question');
  };
  if (creating) { show(null); return; }
  api('GET', '/quiz/questions/' + x.id).then(function (d) { x = d.question; show(d.stats); }, function (err) { toast('error', qzErrText(err)); });
}

function hintList(questionId, box) {
  clear(box).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  api('GET', '/quiz/questions/' + questionId + '/hints').then(function (d) {
    var list = h('div', { class: 'flex flex-col gap-2', 'data-quiz-hints': '' });
    (d.items || []).forEach(function (hn) {
      list.appendChild(h('div', { class: 'qz-track flex flex-wrap items-center gap-2', 'data-quiz-hint': hn.id },
        h('span', { class: 'font-medium', text: textOf(hn.name, '') }), hn.isActive ? null : tag(t('inactive'), 'warning'),
        tag(hn.price && hn.price.amount ? '⁦' + fmtNum(hn.price.amount) + ' ' + hn.price.currencyKey + '⁩' : t('hn.free'), 'primary'),
        h('span', { class: 'text-xs text-secondary-500', text: t('hn.bought', { n: fmtNum(hn.bought) }) }), h('div', { class: 'flex-1' }),
        button({ label: t('edit'), data: { 'quiz-hint-edit': hn.id }, onclick: function () { hintForm(questionId, hn, box); } }),
        button({ label: t('delete'), variant: 'danger', onclick: function () {
          confirmBox(t('hn.delete'), t('delete')).then(function (ok) {
            if (ok) api('DELETE', '/quiz/hints/' + hn.id).then(function () { hintList(questionId, box); }, function (err) { toast('error', qzErrText(err)); });
          });
        } })));
    });
    clear(box).appendChild(section(t('hn.title'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.hints') }), list,
      h('div', { class: 'mt-2' }, button({ label: '+ ' + t('hn.add'), data: { 'quiz-action': 'add-hint' }, onclick: function () { hintForm(questionId, null, box); } }))), 'hints'));
  }, function (err) { clear(box).appendChild(notice('error', qzErrText(err))); });
}

function hintForm(questionId, hn, box) {
  var creating = !hn;
  hn = hn || { name: {}, text: {}, price: {}, isActive: true };
  var name = translated(t('f.name'), 'hint-name', hn.name);
  var text = translated(t('hn.text'), 'hint-text', hn.text, { multiline: true });
  var media = input(hn.mediaUrl, { name: 'hint-media', ltr: true });
  var cur = refSelect('currencies', (hn.price || {}).currencyKey, 'hint-currency', t('hn.free'));
  var amount = numInput((hn.price || {}).amount || '', 'hint-amount');
  var active = check(hn.isActive, t('f.active'), 'hint-active');
  var msg = h('div', {});
  var form = h('div', { class: 'qz-track', 'data-quiz-form': 'hint' }, name, text, field(t('qs.media'), media),
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('hn.currency'), cur), field(t('hn.amount'), amount)), active, msg,
    h('div', { class: 'flex gap-2 justify-end' }, button({ label: t('cancel'), onclick: function () { hintList(questionId, box); } }),
      button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-hint' }, onclick: function () {
        clear(msg);
        var out = { name: name.read(), text: text.read(), mediaUrl: media.value.trim(), price: { currencyKey: cur.value, amount: cur.value ? numberValue(amount) || 0 : 0 }, isActive: active.input.checked };
        api(creating ? 'POST' : 'PUT', creating ? '/quiz/questions/' + questionId + '/hints' : '/quiz/hints/' + hn.id, out).then(function () { toast('success', t('saved')); hintList(questionId, box); },
          function (err) { msg.appendChild(notice('error', qzErrText(err))); });
      } })));
  clear(box).appendChild(section(t('hn.title'), form, 'hints'));
}

function missionsPage(root) {
  var items = [];
  var frame = pageFrame('missions', t('desc.missions'), [
    button({ label: t('ms.newGame'), data: { 'quiz-action': 'new-game' }, onclick: function () { missionEditor(null, 'game', items, reload); } }),
    button({ label: t('ms.new'), variant: 'primary', data: { 'quiz-action': 'new-mission' }, onclick: function () { missionEditor(null, 'mission', items, reload); } })]);
  clear(root).appendChild(frame.el);
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(listBox);
  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/missions'), loadLanguages(), loadRefs(true)]).then(function (res) {
      items = res[0].items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'missions' }, table([
        { label: t('f.name') }, { label: t('ms.kind') }, { label: t('ms.game') }, { label: t('ms.order') }, { label: t('ms.time') }, { label: t('ms.items') }, { label: t('f.reward') }, { label: '' }
      ], items.map(function (m) {
        return { id: m.key, item: m, cells: [
          h('div', { class: 'flex flex-col' }, h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(m.name, m.key) }), m.isActive ? null : tag(t('inactive'), 'warning')), mono(m.key)),
          tag(t('ms.kind.' + m.kind), m.kind === 'game' ? 'primary' : 'secondary'),
          h('span', { text: m.parentKey || t('none') }),
          h('span', { text: t('ms.order.' + m.orderMode) }),
          h('span', { text: m.timeSeconds ? fmtNum(m.timeSeconds) : t('none') }),
          m.kind === 'mission' ? h('a', { class: 'eco-link', href: '#/quiz/packs?missionId=' + m.id, text: t('ms.packs', { n: fmtNum(m.items) }) }) : h('span', { text: t('ms.missions', { n: fmtNum(m.items) }) }),
          h('span', { class: 'text-xs', text: rewardText(m.reward) }),
          button({ label: t('delete'), variant: 'danger', data: { 'quiz-delete': m.key }, onclick: function () {
            confirmBox(t('ms.delete', { key: m.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/quiz/missions/' + m.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', err.code === 'IN_USE' ? t('ms.inUse') : qzErrText(err)); });
            });
          } })
        ] };
      }), function (m) { missionEditor(m, m.kind, items, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('ms.empty') })));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function missionEditor(m, kind, all, done) {
  var creating = !m;
  m = m || { kind: kind, name: {}, description: {}, isActive: true, reward: {}, orderMode: 'all_at_once', metadata: {} };
  var games = all.filter(function (x) { return x.kind === 'game'; });
  var key = input(m.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var parent = select(m.parentId || '', [{ value: '', label: t('ms.noGame') }].concat(games.map(function (g) { return { value: g.id, label: textOf(g.name, g.key) }; })), { name: 'parentId', label: t('ms.game') });
  var name = translated(t('f.name'), 'name', m.name);
  var desc = translated(t('f.description'), 'description', m.description, { multiline: true });
  var order = select(m.orderMode, ['all_at_once', 'sequential'].map(function (o) { return { value: o, label: t('ms.order.' + o) }; }), { name: 'orderMode', label: t('ms.order') });
  var time = numInput(m.timeSeconds || '', 'timeSeconds');
  var send = check(m.sendAnswers, t('pk.sendAnswers'), 'sendAnswers');
  var icon = input(m.iconUrl, { name: 'iconUrl', ltr: true });
  var sort = numInput(m.sort || 0, 'sort');
  var active = check(m.isActive, t('f.active'), 'isActive');
  var reward = rewardEditor(m.reward, 'mission');
  var meta = jsonInput(m.metadata, 'metadata');
  var msg = h('div', {});
  var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
  var body = h('div', { 'data-quiz-form': 'mission' },
    h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('ms.hint.' + m.kind) }),
    grid(field(t('f.key'), key, t('hint.key')), m.kind === 'mission' ? field(t('ms.game'), parent) : h('div', {})), name, desc,
    grid(field(t('ms.order'), order, t('ms.orderHint.' + m.kind)), field(t('ms.timeF'), time, t('hint.missionTime'))),
    m.kind === 'mission' ? send : null, grid(field(t('f.icon'), icon), field(t('f.sort'), sort)), active,
    section(t('ms.reward'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.missionReward') }), reward), 'reward'),
    h('div', { class: 'h-3' }), field(t('f.metadata'), meta), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-mission' }, onclick: function () {
    busy(save, true);
    clear(msg);
    var md = meta.read();
    if (md === undefined) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_JSON'))); return; }
    var out = { key: key.value.trim(), kind: m.kind, parentId: m.kind === 'mission' ? parent.value : '', name: name.read(), description: desc.read(), orderMode: order.value,
      timeSeconds: numberValue(time) || 0, sendAnswers: send.input.checked, iconUrl: icon.value.trim(), sort: numberValue(sort) || 0, isActive: active.input.checked, reward: reward.read(), metadata: md };
    api(creating ? 'POST' : 'PUT', '/quiz/missions' + (creating ? '' : '/' + m.id), out).then(function () {
      closePanel(); toast('success', t('saved')); qzRefs = null; done();
    }, function (err) { busy(save, false); msg.appendChild(notice('error', qzErrText(err))); });
  } });
  sidePanel(creating ? t(m.kind === 'game' ? 'ms.newGame' : 'ms.new') : textOf(m.name, m.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'mission');
}

function reportsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 50, total: 0, status: q.status != null ? q.status : 'open', reasonKey: '', questionId: q.questionId || '' };
  var frame = pageFrame('reports', t('desc.reports'), [button({ label: t('rp.reasons'), data: { 'quiz-action': 'reasons' }, onclick: reasonsPanel })]);
  clear(root).appendChild(frame.el);
  var statusSel = select(state.status, [{ value: '', label: t('all') }].concat(['open', 'resolved', 'dismissed'].map(function (s) { return { value: s, label: t('st.' + s) }; })), { name: 'status', label: t('rp.status') });
  var reasonSel = h('select', { class: QZ_INPUT + ' eco-select', name: 'reasonKey', 'aria-label': t('rp.reason') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end qz-filters' }, field(t('rp.status'), statusSel), field(t('rp.reason'), reasonSel)));
  frame.content.appendChild(listBox);
  statusSel.addEventListener('change', function () { state.status = statusSel.value; state.page = 1; reload(); });
  reasonSel.addEventListener('change', function () { state.reasonKey = reasonSel.value; state.page = 1; reload(); });
  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/reports' + qs({ status: state.status, reasonKey: state.reasonKey, questionId: state.questionId, page: state.page, limit: state.limit })), api('GET', '/quiz/reasons'), loadLanguages()]).then(function (res) {
      var d = res[0];
      var reasons = res[1].items || [];
      clear(reasonSel);
      [{ value: '', label: t('all') }].concat(reasons.map(function (r) { return { value: r.key, label: textOf(r.name, r.key) }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.reasonKey) opt.selected = true;
        reasonSel.appendChild(opt);
      });
      state.total = d.total || 0;
      clear(listBox);
      listBox.appendChild(h('div', { class: 'flex flex-wrap gap-2' }, tag(t('rp.openCount', { n: fmtNum(d.open) }), d.open ? 'warning' : 'success')));
      var items = d.items || [];
      listBox.appendChild(h('div', { 'data-quiz-list': 'reports' }, table([
        { label: t('rp.date') }, { label: t('rp.player') }, { label: t('qs.question') }, { label: t('rp.reason') }, { label: t('rp.note') }, { label: t('rp.status') }, { label: '' }
      ], items.map(function (r) {
        return { id: r.id, item: r, cells: [
          h('span', { class: 'text-xs', text: fmtDate(r.createdAt) }),
          playerLink(r.userId, r.username),
          h('div', { class: 'flex flex-col gap-1' }, h('span', { dir: 'auto', class: 'qz-sentence', text: textOf(r.question, r.questionId) }),
            h('div', { class: 'flex flex-wrap gap-1' }, r.packKey ? mono(r.packKey) : null, r.questionActive ? null : tag(t('inactive'), 'warning'), r.openReports > 1 ? tag(t('qs.reports', { n: r.openReports }), 'error') : null)),
          h('span', { text: textOf(r.reasonName, r.reasonKey) }),
          h('span', { class: 'text-xs qz-sentence', dir: 'auto', text: r.note || t('none') }),
          h('div', { class: 'flex flex-col gap-1' }, statusTag(r.status), r.resolution ? h('span', { class: 'text-xs text-secondary-500', dir: 'auto', text: r.resolution }) : null),
          button({ label: t('rp.review'), data: { 'quiz-review': r.id }, onclick: function () { reviewPanel(r, reload); } })
        ] };
      }), function (r) { reviewPanel(r, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('rp.empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function reviewPanel(r, done) {
  var resolution = input(r.resolution, { name: 'resolution', multiline: true, rows: 3 });
  var same = check(r.openReports > 1, t('rp.same', { n: r.openReports }), 'sameQuestion');
  var disable = check(false, t('rp.disable'), 'disableQuestion');
  var msg = h('div', {});
  var body = h('div', { 'data-quiz-form': 'review' },
    h('p', { class: 'font-medium mb-2', dir: 'auto', text: textOf(r.question, r.questionId) }),
    h('p', { class: 'text-sm mb-1', text: t('rp.reason') + ': ' + textOf(r.reasonName, r.reasonKey) }),
    r.note ? h('p', { class: 'text-sm mb-3', dir: 'auto', text: r.note }) : null,
    h('p', { class: 'mb-3' }, h('a', { class: 'eco-link', href: '#/quiz/questions?q=' + encodeURIComponent(r.questionId), 'data-quiz-edit-question': r.questionId, onclick: function (e) {
      e.preventDefault(); closePanel(); api('GET', '/quiz/packs?limit=200').then(function (d) { questionEditor({ id: r.questionId }, d.items || [], function () {}); });
    }, text: t('rp.editQuestion') })),
    field(t('rp.resolution'), resolution), same, disable, msg);
  var act = function (status) {
    return button({ label: t('rp.do.' + status), variant: status === 'resolved' ? 'primary' : 'ghost', data: { 'quiz-action': 'review-' + status }, onclick: function () {
      clear(msg);
      api('PUT', '/quiz/reports/' + r.id, { status: status, resolution: resolution.value.trim(), sameQuestion: same.input.checked, disableQuestion: disable.input.checked }).then(function () {
        closePanel(); toast('success', t('saved')); done();
      }, function (err) { msg.appendChild(notice('error', qzErrText(err))); });
    } });
  };
  sidePanel(t('rp.review'), body, [button({ label: t('cancel'), onclick: closePanel }), r.status !== 'open' ? act('open') : null, act('dismissed'), act('resolved')].filter(Boolean), 'review');
}

function reasonsPanel() {
  var body = h('div', { 'data-quiz-form': 'reasons' }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  sidePanel(t('rp.reasons'), body, [button({ label: t('close'), onclick: closePanel })], 'reasons');
  function paint() {
    Promise.all([api('GET', '/quiz/reasons'), loadLanguages()]).then(function (res) {
      clear(body);
      body.appendChild(h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('rp.reasonsHint') }));
      (res[0].items || []).forEach(function (r) { body.appendChild(reasonRow(r, paint)); });
      body.appendChild(h('div', { class: 'h-3' }));
      body.appendChild(reasonRow(null, paint));
    }, function (err) { clear(body).appendChild(notice('error', qzErrText(err))); });
  }
  paint();
}

function reasonRow(r, done) {
  var creating = !r;
  r = r || { key: '', name: {}, sort: 100, isActive: true };
  var key = input(r.key, { name: 'reason-key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = translated(t('f.name'), 'reason-name', r.name);
  var sort = numInput(r.sort, 'reason-sort');
  var active = check(r.isActive, t('f.active'), 'reason-active');
  var msg = h('div', {});
  return h('div', { class: 'qz-track mb-2', 'data-quiz-reason': r.key || 'new' },
    creating ? h('p', { class: 'font-medium mb-2', text: t('rp.newReason') }) : null,
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.key'), key), field(t('f.sort'), sort)), name, active, msg,
    h('div', { class: 'flex gap-2 justify-end' },
      creating ? null : button({ label: t('delete'), variant: 'danger', onclick: function () {
        api('DELETE', '/quiz/reasons/' + r.id).then(done, function (err) { msg.appendChild(notice('error', err.code === 'IN_USE' ? t('rp.reasonInUse') : qzErrText(err))); });
      } }),
      button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-reason' }, onclick: function () {
        clear(msg);
        api(creating ? 'POST' : 'PUT', '/quiz/reasons' + (creating ? '' : '/' + r.id), { key: key.value.trim(), name: name.read(), sort: numberValue(sort) || 0, isActive: active.input.checked }).then(function () {
          toast('success', t('saved')); done();
        }, function (err) { msg.appendChild(notice('error', qzErrText(err))); });
      } })));
}

function resultsPage(root) {
  var data = null;
  var frame = pageFrame('results', t('desc.results'), [button({ label: t('rs.new'), variant: 'primary', data: { 'quiz-action': 'new-template' }, onclick: function () { templateEditor(null, data, reload); } })]);
  clear(root).appendChild(frame.el);
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(listBox);
  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/templates'), api('GET', '/quiz/missions'), loadRefs(true)]).then(function (res) {
      data = res[0];
      data.missions = res[1].items || [];
      var items = data.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'templates' }, table([
        { label: t('f.name') }, { label: t('rs.mission') }, { label: t('rs.webhook') }, { label: '' }
      ], items.map(function (x) {
        return { id: x.key, item: x, cells: [
          h('div', { class: 'flex flex-col' }, h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: x.name }), x.isActive ? null : tag(t('inactive'), 'warning')), mono(x.key)),
          h('span', { text: x.missionKey || t('none') }),
          h('span', { class: 'text-xs', dir: 'ltr', text: x.webhookUrl || t('none') }),
          h('div', { class: 'flex gap-1' }, button({ label: t('rs.deliveries'), data: { 'quiz-deliveries': x.key }, onclick: function () { deliveriesPanel(x); } }),
            button({ label: t('delete'), variant: 'danger', data: { 'quiz-delete': x.key }, onclick: function () {
              confirmBox(t('rs.delete', { key: x.key }), t('delete')).then(function (ok) {
                if (ok) api('DELETE', '/quiz/templates/' + x.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', qzErrText(err)); });
              });
            } }))
        ] };
      }), function (x) { templateEditor(x, data, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('rs.empty') })));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function templateEditor(x, data, done) {
  var creating = !x;
  x = x || { name: '', body: '', wonText: '', lostText: '', isActive: true };
  var key = input(x.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = input(x.name, { name: 'name' });
  var mission = select(x.missionId || '', [{ value: '', label: t('none') }].concat((data.missions || []).map(function (m) { return { value: m.id, label: t('ms.kind.' + m.kind) + ' · ' + textOf(m.name, m.key) }; })), { name: 'missionId', label: t('rs.mission') });
  var bodyIn = input(x.body, { name: 'body', multiline: true, rows: 8 });
  var won = input(x.wonText, { name: 'wonText', placeholder: 'won' });
  var lost = input(x.lostText, { name: 'lostText', placeholder: 'lost' });
  var url = input(x.webhookUrl, { name: 'webhookUrl', ltr: true, placeholder: 'https://' });
  var secret = input('', { name: 'webhookSecret', ltr: true, type: 'password', placeholder: x.hasSecret ? t('rs.secretKept') : '' });
  var active = check(x.isActive, t('f.active'), 'isActive');
  var chips = h('div', { class: 'flex flex-wrap gap-1 mb-4', 'data-quiz-placeholders': '' }, (data.placeholders || []).map(function (p) {
    return h('button', { type: 'button', class: 'qz-chip', title: t('rs.insert'), onclick: function () {
      var at = bodyIn.selectionStart || bodyIn.value.length;
      bodyIn.value = bodyIn.value.slice(0, at) + '{' + p.key + '}' + bodyIn.value.slice(at);
      bodyIn.focus();
    }, text: '{' + p.key + '}' });
  }));
  var msg = h('div', {});
  var body = h('div', { 'data-quiz-form': 'template' },
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.key'), key), field(t('f.name'), name)), field(t('rs.mission'), mission, t('hint.rsMission')),
    field(t('rs.body'), bodyIn), h('p', { class: 'text-xs text-secondary-500 mb-1', text: t('rs.placeholders') }), chips,
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('rs.won'), won), field(t('rs.lost'), lost)),
    section(t('rs.webhook'), h('div', {}, field(t('rs.url'), url, t('hint.webhook')), field(t('rs.secret'), secret, data.secrets ? t('hint.secret') : t('rs.noSecrets'))), 'webhook'),
    h('div', { class: 'h-3' }), active, msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-template' }, onclick: function () {
    busy(save, true);
    clear(msg);
    var out = { key: key.value.trim(), name: name.value.trim(), missionId: mission.value, body: bodyIn.value, wonText: won.value.trim(), lostText: lost.value.trim(), webhookUrl: url.value.trim(), isActive: active.input.checked };
    if (secret.value) out.webhookSecret = secret.value;
    api(creating ? 'POST' : 'PUT', '/quiz/templates' + (creating ? '' : '/' + x.id), out).then(function () {
      closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); msg.appendChild(notice('error', qzErrText(err))); });
  } });
  sidePanel(creating ? t('rs.new') : x.name, body, [button({ label: t('cancel'), onclick: closePanel }), save], 'template');
}

function deliveriesPanel(x) {
  var body = h('div', { 'data-quiz-deliveries-list': x.key }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  sidePanel(t('rs.deliveries') + ' · ' + x.name, body, [button({ label: t('close'), onclick: closePanel })], 'deliveries');
  api('GET', '/quiz/templates/' + x.id + '/deliveries').then(function (d) {
    var items = d.items || [];
    clear(body).appendChild(items.length ? table([{ label: t('rp.date') }, { label: t('rp.player') }, { label: t('rp.status') }, { label: t('rs.attempts') }, { label: t('rs.error') }],
      items.map(function (it) {
        return { id: it.id, cells: [h('span', { class: 'text-xs', text: fmtDate(it.createdAt) }), playerLink(it.userId), statusTag(it.status), h('span', { text: fmtNum(it.attempts) }),
          h('span', { class: 'text-xs', dir: 'ltr', text: it.lastError || '' })] };
      })) : h('p', { class: 'text-sm text-secondary-500', text: t('empty') }));
  }, function (err) { clear(body).appendChild(notice('error', qzErrText(err))); });
}

var QZ_IMPORT_SAMPLE = {
  packs: [{ key: 'general-1', name: { fa: 'دانش عمومی', en: 'General knowledge' }, questionCount: 10, passCorrect: 8, questionSeconds: 30 }],
  questions: [
    { key: 'gk-001', packKey: 'general-1', type: 'options', question: { fa: 'پایتخت فرانسه کجاست؟', en: 'Capital of France?' },
      options: [{ id: 0, text: { fa: 'لندن', en: 'London' } }, { id: 1, text: { fa: 'پاریس', en: 'Paris' }, correct: true }, { id: 2, text: { fa: 'رم', en: 'Rome' } }],
      explanation: { fa: 'پاریس پایتخت فرانسه است.' }, difficulty: 1, hints: [{ name: { fa: 'راهنما' }, text: { fa: 'شهر برج ایفل' }, price: { currencyKey: 'coin', amount: 50 } }] },
    { packKey: 'general-1', type: 'text', question: { en: 'Largest ocean on Earth?' }, answers: { en: ['Pacific', 'Pacific Ocean'], fa: ['آرام', 'اقیانوس آرام'] } }],
  rows: [{ 'Pack Key': 'general-1', 'Type': 'options', 'Question': 'Which numbers are even?', 'Option 1': '1', 'Option 2': '2', 'Option 3': '3', 'Option 4': '4', 'Correct Options': '2,4' }]
};

function importPage(root) {
  var frame = pageFrame('import', t('desc.import'));
  clear(root).appendChild(frame.el);
  var text = input('', { name: 'import-json', ltr: true, mono: true, multiline: true, rows: 14, placeholder: JSON.stringify(QZ_IMPORT_SAMPLE, null, 2) });
  var file = h('input', { type: 'file', accept: '.json,application/json', name: 'import-file', class: 'text-sm' });
  file.addEventListener('change', function () {
    var f = file.files && file.files[0];
    if (!f) return;
    var reader = new FileReader();
    reader.onload = function () { text.value = String(reader.result || ''); };
    reader.readAsText(f);
  });
  var packSel = h('select', { class: QZ_INPUT + ' eco-select', name: 'import-pack', 'aria-label': t('im.pack') });
  var lang = input('fa', { name: 'import-lang', ltr: true });
  var result = h('div', { 'data-quiz-import-result': '' });
  var exportPack = h('select', { class: QZ_INPUT + ' eco-select', name: 'export-pack', 'aria-label': t('im.exportPack') });
  function run(dry, btn) {
    clear(result);
    var doc;
    try { doc = JSON.parse(text.value || '{}'); } catch (e) { result.appendChild(notice('error', t('code.INVALID_JSON'))); return; }
    if (Array.isArray(doc)) doc = { rows: doc };
    if (packSel.value) doc.packKey = packSel.value;
    if (lang.value.trim()) doc.lang = lang.value.trim();
    doc.dryRun = dry;
    busy(btn, true);
    api('POST', '/quiz/import', doc).then(function (out) {
      busy(btn, false);
      var c = out.counts || {};
      var line = ['missions', 'packs', 'questions'].map(function (k) { var x = c[k] || {}; return t('im.count.' + k, { c: fmtNum(x.created || 0), u: fmtNum(x.updated || 0) }); }).join(' · ');
      result.appendChild(notice('success', (dry ? t('im.dryOk') : t('im.done')) + ' ' + line));
      qzRefs = null;
    }, function (err) {
      busy(btn, false);
      result.appendChild(notice('error', qzErrText(err)));
      var errs = err.meta && err.meta.errors;
      if (errs && errs.length) {
        result.appendChild(table([{ label: t('im.section') }, { label: t('im.row') }, { label: t('im.field') }, { label: t('im.error') }], errs.map(function (e, i) {
          return { id: String(i), cells: [h('span', { text: e.section }), h('span', { text: digits(String(e.index + 1)) }), mono(e.field), h('span', { text: has('code.' + e.code) ? t('code.' + e.code) : e.code })] };
        })));
      }
    });
  }
  var dryBtn = button({ label: t('im.dry'), data: { 'quiz-action': 'import-dry' }, onclick: function () { run(true, dryBtn); } });
  var goBtn = button({ label: t('im.go'), variant: 'primary', data: { 'quiz-action': 'import' }, onclick: function () { run(false, goBtn); } });
  var sampleBtn = button({ label: t('im.sample'), onclick: function () { text.value = JSON.stringify(QZ_IMPORT_SAMPLE, null, 2); } });
  var expBtn = button({ label: t('im.export'), data: { 'quiz-action': 'export' }, onclick: function () {
    busy(expBtn, true);
    api('GET', '/quiz/export' + qs({ packId: exportPack.value })).then(function (d) {
      busy(expBtn, false);
      var blob = new Blob([JSON.stringify(d, null, 2)], { type: 'application/json' });
      var a = h('a', { href: URL.createObjectURL(blob), download: 'quiz-export.json' });
      document.body.appendChild(a); a.click(); a.remove();
      toast('success', t('im.exported', { n: fmtNum((d.questions || []).length) }));
    }, function (err) { busy(expBtn, false); toast('error', qzErrText(err)); });
  } });
  frame.content.appendChild(section(t('im.title'), h('div', {},
    h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('im.help') }),
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('im.pack'), packSel, t('hint.importPack')), field(t('im.lang'), lang, t('hint.importLang'))),
    field(t('im.file'), file), field(t('im.json'), text), h('div', { class: 'flex flex-wrap gap-2 mb-3' }, sampleBtn, dryBtn, goBtn), result), 'import'));
  frame.content.appendChild(section(t('im.exportTitle'), h('div', {}, h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('im.exportHelp') }),
    h('div', { class: 'flex flex-wrap gap-2 items-end' }, field(t('im.exportPack'), exportPack), expBtn)), 'export'));
  Promise.all([api('GET', '/quiz/packs?limit=200'), loadLanguages()]).then(function (res) {
    var packs = res[0].items || [];
    [{ value: '', label: t('im.packFromFile') }].concat(packs.map(function (p) { return { value: p.key, label: textOf(p.name, p.key) }; })).forEach(function (o) { packSel.appendChild(h('option', { value: o.value, text: o.label })); });
    [{ value: '', label: t('all') }].concat(packs.map(function (p) { return { value: p.id, label: textOf(p.name, p.key) }; })).forEach(function (o) { exportPack.appendChild(h('option', { value: o.value, text: o.label })); });
  });
}

function statsPage(root) {
  var q = hashQuery();
  var days = Number(q.days) || 30;
  var frame = pageFrame('stats', t('desc.stats'));
  clear(root).appendChild(frame.el);
  var daySel = select(String(days), [7, 30, 90].map(function (d) { return { value: String(d), label: t('sx.d', { n: d }) }; }), { name: 'days', label: t('sx.days') });
  daySel.addEventListener('change', function () { location.hash = '#/quiz/stats?days=' + daySel.value; });
  var box = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end qz-filters' }, field(t('sx.days'), daySel)));
  frame.content.appendChild(box);
  box.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  api('GET', '/quiz/stats' + qs({ days: days, lang: qzLang() })).then(function (d) {
    clear(box);
    var tot = d.totals || {};
    var rate = tot.finished ? Math.round(tot.passes * 1000 / tot.finished) / 10 : 0;
    var kpi = function (k, v) {
      return card(h('div', { class: 'p-4', 'data-quiz-kpi': k }, h('div', { class: 'text-xs text-secondary-500', text: t('sx.kpi.' + k) }), h('div', { class: 'text-2xl font-semibold mt-1', text: v })));
    };
    box.appendChild(h('div', { class: 'grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-3' }, kpi('players', fmtNum(tot.players)), kpi('plays', fmtNum(tot.plays)), kpi('finished', fmtNum(tot.finished)),
      kpi('rate', fmtNum(rate) + '٪'), kpi('hints', fmtNum(tot.hintsBought)), kpi('reports', fmtNum(tot.openReports))));
    var packs = d.packs || [];
    box.appendChild(h('div', { 'data-quiz-list': 'stats' }, table([
      { label: t('f.pack') }, { label: t('f.mission') }, { label: t('sx.players') }, { label: t('sx.plays') }, { label: t('sx.finished') }, { label: t('sx.passes') }, { label: t('sx.rate') }, { label: t('sx.avg') }, { label: t('sx.secs') }
    ], packs.map(function (p) {
      return { id: p.key, cells: [h('div', { class: 'flex flex-col' }, h('span', { class: 'font-medium', text: p.name }), mono(p.key)), h('span', { text: p.missionKey || t('none') }),
        h('span', { text: fmtNum(p.players) }), h('span', { text: fmtNum(p.plays) }), h('span', { text: fmtNum(p.finished) }), h('span', { text: fmtNum(p.passes) }),
        h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'qz-bar flex-1' }, h('span', { style: 'width:' + (p.passRate || 0) + '%' })), h('span', { class: 'text-xs', text: fmtNum(p.passRate) + '٪' })),
        h('span', { text: fmtNum(p.avgScore) }), h('span', { text: fmtNum(p.avgSeconds) })] };
    }))));
    if (!packs.length) box.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') })));
    var hard = d.hardest || [];
    box.appendChild(section(t('sx.hardest'), hard.length ? table([{ label: t('qs.question') }, { label: t('f.pack') }, { label: t('sx.answers') }, { label: t('sx.correctRate') }, { label: t('sx.skipped') }],
      hard.map(function (x) {
        return { id: x.id, cells: [h('span', { dir: 'auto', class: 'qz-sentence', text: x.question }), h('span', { text: x.packKey || t('qs.bank') }), h('span', { text: fmtNum(x.answers) }),
          h('span', { text: fmtNum(x.correctRate) + '٪' }), h('span', { text: fmtNum(x.skipped) })] };
      })) : h('p', { class: 'text-sm text-secondary-500', text: t('sx.noHard') }), 'hardest'));
  }, function (err) { clear(box).appendChild(notice('error', qzErrText(err))); });
}

function colorDot(c) {
  return h('span', { class: 'qz-dot', style: c ? 'background:' + c : '', 'aria-hidden': 'true' });
}

function duelCategoriesPage(root) {
  var state = { page: 1, limit: 50, total: 0, q: '' };
  var frame = pageFrame('duel-categories', t('desc.duel-categories'), [
    button({ label: t('dc.new'), variant: 'primary', data: { 'quiz-action': 'new-duel-category' }, onclick: function () { duelCategoryEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  var search = input('', { name: 'q', placeholder: t('search') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end qz-filters' }, field(t('search'), search)));
  frame.content.appendChild(listBox);
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.q = search.value.trim(); state.page = 1; reload(); }, 300); });
  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/duel-categories' + qs({ q: state.q, page: state.page, limit: state.limit })), loadLanguages()]).then(function (res) {
      var d = res[0];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'duel-categories' }, table([
        { label: t('f.name') }, { label: t('dc.questions') }, { label: t('dc.picks') }, { label: t('dc.state') }, { label: '' }
      ], items.map(function (c) {
        return { id: c.key, item: c, cells: [
          h('div', { class: 'flex items-center gap-2' }, c.iconUrl ? h('img', { src: c.iconUrl, alt: '', class: 'qz-cat-icon' }) : colorDot(c.color),
            h('div', { class: 'flex flex-col' }, h('span', { class: 'font-medium', dir: 'auto', text: textOf(c.name, c.key) }), mono(c.key))),
          h('div', { class: 'flex flex-col gap-1' }, h('span', { text: t('dc.count', { n: fmtNum(c.questions), e: fmtNum(c.eligible) }) }),
            c.playable ? tag(t('dc.playable'), 'success') : tag(t('dc.notEnough', { n: fmtNum(d.perRound) }), 'warning')),
          h('span', { text: fmtNum(c.rounds) }),
          h('div', { class: 'flex flex-wrap gap-1' }, c.isActive ? tag(t('active'), 'success') : tag(t('inactive'), 'warning'), c.duel ? tag(t('dc.inDuel'), 'primary') : tag(t('dc.notInDuel'), 'secondary')),
          button({ label: t('delete'), variant: 'danger', data: { 'quiz-delete': c.key }, onclick: function () {
            confirmBox(t('dc.delete', { key: c.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/quiz/duel-categories/' + c.id).then(function () { toast('success', t('saved')); reload(); }, function (err) {
                if (err.code !== 'IN_USE') { toast('error', qzErrText(err)); return; }
                confirmBox(t('dc.inUse'), t('delete')).then(function (force) {
                  if (force) api('DELETE', '/quiz/duel-categories/' + c.id + '?force=true').then(function () { toast('success', t('saved')); reload(); }, function (e2) { toast('error', qzErrText(e2)); });
                });
              });
            });
          } })
        ] };
      }), function (c) { duelCategoryEditor(c, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('dc.empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function duelCategoryEditor(c, done) {
  var creating = !c;
  c = c || { name: {}, description: {}, isActive: true, duel: true, color: '', sort: 0, iconUrl: '' };
  var key = input(c.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = translated(t('f.name'), 'name', c.name);
  var desc = translated(t('f.description'), 'description', c.description, { multiline: true });
  var icon = input(c.iconUrl, { name: 'iconUrl', ltr: true });
  var color = input(c.color || '', { name: 'color', ltr: true, mono: true, placeholder: '#22AA66' });
  var picker = h('input', { type: 'color', class: 'qz-color', name: 'color-picker', 'aria-label': t('dc.color'), value: c.color || '#3b82f6' });
  picker.addEventListener('input', function () { color.value = picker.value.toUpperCase(); });
  color.addEventListener('input', function () { if (/^#[0-9A-Fa-f]{6}$/.test(color.value.trim())) picker.value = color.value.trim(); });
  var sort = numInput(c.sort || 0, 'sort');
  var active = check(c.isActive, t('f.active'), 'isActive');
  var duel = check(c.duel, t('dc.duelOn'), 'duel');
  var msg = h('div', {});
  var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
  var body = h('div', { 'data-quiz-form': 'duel-category' },
    h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('dc.hint') }),
    field(t('f.key'), key, t('dc.hint.key')), name, desc,
    grid(field(t('f.icon'), icon), field(t('dc.color'), h('div', { class: 'flex gap-2 items-center' }, picker, color))),
    grid(field(t('f.sort'), sort), h('div', {})), active, duel, msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-duel-category' }, onclick: function () {
    busy(save, true);
    clear(msg);
    var out = { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), color: color.value.trim(), sort: numberValue(sort) || 0,
      isActive: active.input.checked, duel: duel.input.checked };
    api(creating ? 'POST' : 'PUT', '/quiz/duel-categories' + (creating ? '' : '/' + c.id), out).then(function () {
      closePanel(); toast('success', t('saved')); qzRefs = null; done();
    }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  sidePanel(creating ? t('dc.new') : textOf(c.name, c.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'duel-category');
}

function duelStatusTag(s) {
  var color = s === 'active' ? 'primary' : s === 'finished' ? 'success' : s === 'waiting' || s === 'invited' ? 'warning' : 'secondary';
  return tag(t('du.st.' + s), color);
}

function duelPlayer(p, winner) {
  if (!p) return h('span', { class: 'text-xs text-secondary-500', text: t('du.noOpponent') });
  return h('div', { class: 'flex items-center gap-1.5' }, playerLink(p.userId, p.displayName || p.username), h('span', { class: 'font-semibold', text: fmtNum(p.score) }),
    winner && winner === p.userId ? tag(t('du.winner'), 'success') : null);
}

function duelsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 20, total: 0, status: q.status || 'all', mode: '', player: q.player || '' };
  var frame = pageFrame('duels', t('desc.duels'));
  clear(root).appendChild(frame.el);
  var tiles = h('div', { class: 'grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-3', 'data-quiz-kpis': 'duels' });
  var statusSel = select(state.status, ['all', 'open', 'active', 'waiting', 'invited', 'finished', 'expired', 'declined', 'cancelled'].map(function (s) { return { value: s, label: s === 'all' ? t('all') : t('du.st.' + s) }; }), { name: 'status', label: t('du.status') });
  var modeSel = select('', [{ value: '', label: t('all') }, { value: 'random', label: t('du.mode.random') }, { value: 'friend', label: t('du.mode.friend') }], { name: 'mode', label: t('du.mode') });
  var player = input(state.player, { name: 'player', placeholder: t('du.playerHint') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(tiles);
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end qz-filters' }, field(t('du.status'), statusSel), field(t('du.mode'), modeSel), field(t('du.player'), player)));
  frame.content.appendChild(listBox);
  statusSel.addEventListener('change', function () { state.status = statusSel.value; state.page = 1; reload(); });
  modeSel.addEventListener('change', function () { state.mode = modeSel.value; state.page = 1; reload(); });
  var timer = null;
  player.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.player = player.value.trim(); state.page = 1; reload(); }, 300); });
  api('GET', '/quiz/duel-stats' + qs({ days: 30, lang: qzLang() })).then(function (s) {
    var kpi = function (k, v) {
      return card(h('div', { class: 'p-4', 'data-quiz-kpi': k }, h('div', { class: 'text-xs text-secondary-500', text: t('du.kpi.' + k) }), h('div', { class: 'text-2xl font-semibold mt-1', text: v })));
    };
    clear(tiles);
    [kpi('started', fmtNum(s.started)), kpi('finished', fmtNum(s.finished)), kpi('open', fmtNum(s.open)), kpi('waiting', fmtNum(s.waiting)), kpi('timeouts', fmtNum(s.timeouts)),
      kpi('players', fmtNum(s.players))].forEach(function (x) { tiles.appendChild(x); });
  }, function () { clear(tiles); });
  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', '/quiz/duels' + qs({ status: state.status, mode: state.mode, player: state.player, page: state.page, limit: state.limit })).then(function (d) {
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'duels' }, table([
        { label: t('du.playerA') }, { label: t('du.playerB') }, { label: t('du.status') }, { label: t('du.mode') }, { label: t('du.turn') }, { label: t('du.updated') }
      ], items.map(function (x) {
        var turn = x.turnUser ? (x.playerA && x.turnUser === x.playerA.userId ? x.playerA : x.playerB) : null;
        return { id: x.id, item: x, cells: [
          duelPlayer(x.playerA, x.winner), duelPlayer(x.playerB, x.winner),
          h('div', { class: 'flex flex-col gap-1' }, duelStatusTag(x.status), x.reason ? h('span', { class: 'text-xs text-secondary-500', text: t('du.reason.' + x.reason) }) : null),
          h('span', { text: t('du.mode.' + x.mode) }),
          turn ? h('div', { class: 'flex flex-col' }, h('span', { dir: 'auto', text: turn.displayName || turn.username }), x.turnDeadline ? h('span', { class: 'text-xs text-secondary-500', text: t('du.until', { at: fmtDate(x.turnDeadline) }) }) : null) : h('span', { text: t('none') }),
          h('span', { class: 'text-xs', text: fmtDate(x.updatedAt) })
        ] };
      }), function (x) { duelPanel(x.id, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('du.empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function duelAnswerCell(a, options) {
  if (!a) return h('span', { class: 'text-xs text-secondary-400', text: t('du.notAnswered') });
  var opt = (options || []).filter(function (o) { return o.id === a.optionId; })[0];
  var color = a.outcome === 'correct' ? 'success' : a.outcome === 'pending' ? 'warning' : 'error';
  return h('div', { class: 'flex flex-col gap-0.5' }, tag(t('du.out.' + a.outcome), color),
    opt ? h('span', { class: 'text-xs', dir: 'auto', text: opt.text }) : null,
    a.answeredAt ? h('span', { class: 'text-xs text-secondary-500', text: t('du.secs', { n: fmtNum(a.seconds) }) }) : h('span', { class: 'text-xs text-secondary-500', text: t('du.served', { at: fmtDate(a.servedAt) }) }));
}

function duelPanel(id, done) {
  var body = h('div', { 'data-quiz-form': 'duel' }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  var cancel = button({ label: t('du.cancel'), variant: 'danger', data: { 'quiz-action': 'cancel-duel' }, disabled: true });
  sidePanel(t('du.detail'), body, [button({ label: t('close'), onclick: closePanel }), cancel], 'duel');
  api('GET', '/quiz/duels/' + id + qs({ lang: qzLang() })).then(function (d) {
    clear(body);
    var nameOf = function (p) { return p ? (p.displayName || p.username) : t('du.noOpponent'); };
    var open = d.status === 'waiting' || d.status === 'invited' || d.status === 'active';
    cancel.disabled = !open;
    cancel.onclick = function () {
      confirmBox(t('du.cancelConfirm'), t('du.cancel')).then(function (ok) {
        if (!ok) return;
        busy(cancel, true);
        api('POST', '/quiz/duels/' + id + '/cancel', {}).then(function () { closePanel(); toast('success', t('du.cancelled')); done(); }, function (err) { busy(cancel, false); toast('error', qzErrText(err)); });
      });
    };
    var fee = d.rules && d.rules.entryFee && d.rules.entryFee.amount ? t('du.feeV', { n: fmtNum(d.rules.entryFee.amount), c: d.rules.entryFee.currencyKey }) : t('none');
    body.appendChild(h('div', { class: 'flex flex-wrap gap-2 mb-3' }, duelStatusTag(d.status), tag(t('du.mode.' + d.mode), 'secondary'), d.reason ? tag(t('du.reason.' + d.reason), 'secondary') : null));
    body.appendChild(h('div', { class: 'grid sm:grid-cols-2 gap-3 mb-4' },
      card(h('div', { class: 'p-3' }, h('div', { class: 'text-xs text-secondary-500', text: t('du.playerA') }), duelPlayer(d.playerA, d.winner), h('div', { class: 'text-xs text-secondary-500 mt-1', text: t('du.fee.' + (d.feeA || 'none')) }))),
      card(h('div', { class: 'p-3' }, h('div', { class: 'text-xs text-secondary-500', text: t('du.playerB') }), duelPlayer(d.playerB, d.winner), h('div', { class: 'text-xs text-secondary-500 mt-1', text: t('du.fee.' + (d.feeB || 'none')) })))));
    body.appendChild(h('dl', { class: 'qz-dl text-sm mb-4' },
      h('dt', { text: t('du.created') }), h('dd', { text: fmtDate(d.createdAt) }),
      h('dt', { text: t('du.finished') }), h('dd', { text: d.finishedAt ? fmtDate(d.finishedAt) : t('none') }),
      h('dt', { text: t('du.deadline') }), h('dd', { text: d.turnDeadline ? fmtDate(d.turnDeadline) : t('none') }),
      h('dt', { text: t('du.entryFee') }), h('dd', { text: fee }),
      h('dt', { text: t('du.paid') }), h('dd', { text: d.status === 'finished' ? (d.paidA && d.paidB ? t('yes') : t('no')) : t('none') })));
    if (d.payError) body.appendChild(notice('error', d.payError));
    (d.rounds || []).forEach(function (r) {
      var picker = r.picker === (d.playerA && d.playerA.userId) ? d.playerA : d.playerB;
      var head = h('div', { class: 'flex flex-wrap items-center gap-2 mb-2' }, h('span', { class: 'font-semibold', text: t('du.round', { n: r.round }) }),
        r.category ? h('span', { class: 'flex items-center gap-1' }, colorDot(r.category.color), h('span', { dir: 'auto', text: r.category.name })) : tag(t('du.notPicked'), 'secondary'),
        h('span', { class: 'text-xs text-secondary-500', dir: 'auto', text: t('du.pickedBy', { name: nameOf(picker) }) }));
      var offers = r.offers && r.offers.length ? h('p', { class: 'text-xs text-secondary-500 mb-2', dir: 'auto', text: t('du.offers') + ': ' + r.offers.map(function (o) { return o.name; }).join('، ') }) : null;
      var rows = (r.questions || []).map(function (x, i) {
        return h('div', { class: 'qz-row', 'data-quiz-duel-question': x.questionId },
          h('div', { class: 'qz-row-full flex flex-col gap-1' }, h('span', { class: 'font-medium qz-sentence', dir: 'auto', text: fmtNum(i + 1) + '. ' + (x.text || x.questionId) }),
            h('div', { class: 'flex flex-wrap gap-1' }, (x.options || []).map(function (o) { return o.correct ? tag(o.text, 'success') : h('span', { class: 'text-xs text-secondary-500', dir: 'auto', text: o.text }); }))),
          h('div', {}, h('div', { class: 'text-xs text-secondary-500', dir: 'auto', text: nameOf(d.playerA) }), duelAnswerCell(x.a, x.options)),
          h('div', {}, h('div', { class: 'text-xs text-secondary-500', dir: 'auto', text: nameOf(d.playerB) }), duelAnswerCell(x.b, x.options)));
      });
      body.appendChild(section(t('du.round', { n: r.round }), h('div', {}, head, offers, h('div', { class: 'qz-rows' }, rows)), 'round-' + r.round));
    });
  }, function (err) { clear(body).appendChild(notice('error', qzErrText(err))); });
}

function duelSettingsSection(cur) {
  var d = cur.duel || {};
  var ev = d.events || {};
  var fee = d.entryFee || {};
  var pc = d.perCorrect || {};
  var opt = function (prefix, list) { return list.map(function (v) { return { value: v, label: t(prefix + v) }; }); };
  var enabled = check(d.enabled, t('ds.enabled'), 'duel.enabled');
  var rounds = numInput(d.rounds, 'duel.rounds');
  var per = numInput(d.questionsPerRound, 'duel.questionsPerRound');
  var choices = numInput(d.categoryChoices, 'duel.categoryChoices');
  var qsec = numInput(d.questionSeconds, 'duel.questionSeconds');
  var grace = numInput(d.graceSeconds, 'duel.graceSeconds');
  var source = select(d.questionSource, opt('ds.source.', ['any', 'bank']), { name: 'duel.questionSource' });
  var avoid = numInput(d.avoidRepeatDays, 'duel.avoidRepeatDays');
  var shuffle = check(d.shuffleOptions, t('ds.shuffle'), 'duel.shuffleOptions');
  var serveNext = check(d.serveNextOnAnswer, t('ds.serveNext'), 'duel.serveNextOnAnswer');
  var turn = numInput(d.turnHours, 'duel.turnHours');
  var timeout = select(d.timeoutResult, opt('ds.timeout.', ['opponent_wins', 'score']), { name: 'duel.timeoutResult' });
  var waiting = numInput(d.waitingHours, 'duel.waitingHours');
  var invite = numInput(d.inviteHours, 'duel.inviteHours');
  var friends = check(d.friendsOnly, t('ds.friendsOnly'), 'duel.friendsOnly');
  var maxActive = numInput(d.maxActive, 'duel.maxActive');
  var daily = numInput(d.dailyStarts, 'duel.dailyStarts');
  var feeCur = refSelect('currencies', fee.currencyKey, 'duel.entryFee.currencyKey', t('none'));
  var feeAmt = numInput(fee.amount || 0, 'duel.entryFee.amount');
  var win = refSelect('prizes', d.winPrize, 'duel.winPrize', t('none'));
  var lose = refSelect('prizes', d.losePrize, 'duel.losePrize', t('none'));
  var draw = refSelect('prizes', d.drawPrize, 'duel.drawPrize', t('none'));
  var pcCur = refSelect('currencies', pc.currencyKey, 'duel.perCorrect.currencyKey', t('none'));
  var pcAmt = numInput(pc.amount || 0, 'duel.perCorrect.amount');
  var evPlayed = input(ev.played, { name: 'duel.events.played', ltr: true, mono: true });
  var evWon = input(ev.won, { name: 'duel.events.won', ltr: true, mono: true });
  var evCorrect = input(ev.correct, { name: 'duel.events.correct', ltr: true, mono: true });
  var notify = check(d.notify, t('ds.notify'), 'duel.notify');
  var history = numInput(d.historyDays, 'duel.historyDays');
  var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
  return {
    el: section(t('ds.title'), h('div', { 'data-quiz-settings': 'duel' },
      h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('ds.hint') }), enabled,
      grid(field(t('ds.rounds'), rounds), field(t('ds.per'), per)),
      grid(field(t('ds.choices'), choices, t('ds.hint.choices')), field(t('ds.source'), source)),
      grid(field(t('ds.qsec'), qsec), field(t('ds.grace'), grace, t('ds.hint.grace'))),
      grid(field(t('ds.avoid'), avoid, t('ds.hint.avoid')), h('div', {})), shuffle, serveNext,
      grid(field(t('ds.turn'), turn, t('ds.hint.turn')), field(t('ds.timeout'), timeout)),
      grid(field(t('ds.waiting'), waiting, t('ds.hint.waiting')), field(t('ds.invite'), invite)), friends,
      grid(field(t('ds.maxActive'), maxActive), field(t('ds.daily'), daily, t('se.hint.zero'))),
      grid(field(t('ds.feeCur'), feeCur, t('ds.hint.fee')), field(t('ds.feeAmt'), feeAmt)),
      grid(field(t('ds.win'), win), field(t('ds.lose'), lose)),
      grid(field(t('ds.draw'), draw), h('div', {})),
      grid(field(t('ds.pcCur'), pcCur, t('ds.hint.pc')), field(t('ds.pcAmt'), pcAmt)),
      grid(field(t('ds.evPlayed'), evPlayed), field(t('ds.evWon'), evWon)),
      grid(field(t('ds.evCorrect'), evCorrect), field(t('ds.history'), history, t('se.hint.zeroKeep'))), notify), 'duel'),
    read: function (c) {
      c.duel = { enabled: enabled.input.checked, rounds: numberValue(rounds) || 0, questionsPerRound: numberValue(per) || 0, categoryChoices: numberValue(choices) || 0,
        questionSeconds: numberValue(qsec) || 0, graceSeconds: numberValue(grace) || 0, questionSource: source.value, avoidRepeatDays: numberValue(avoid) || 0,
        shuffleOptions: shuffle.input.checked, serveNextOnAnswer: serveNext.input.checked, turnHours: numberValue(turn) || 0, timeoutResult: timeout.value,
        waitingHours: numberValue(waiting) || 0, inviteHours: numberValue(invite) || 0, friendsOnly: friends.input.checked, maxActive: numberValue(maxActive) || 0,
        dailyStarts: numberValue(daily) || 0, entryFee: { currencyKey: feeCur.value, amount: feeCur.value ? numberValue(feeAmt) || 0 : 0 },
        winPrize: win.value, losePrize: lose.value, drawPrize: draw.value, perCorrect: { currencyKey: pcCur.value, amount: pcCur.value ? numberValue(pcAmt) || 0 : 0 },
        events: { played: evPlayed.value.trim(), won: evWon.value.trim(), correct: evCorrect.value.trim() }, notify: notify.input.checked, historyDays: numberValue(history) || 0 };
    }
  };
}

function quizSettingsPage(root) {
  loadRefs(true).then(function () {
    settingsPage(root, 'settings', '/quiz', function (cur, defs, d) {
      var opt = function (prefix, list) { return list.map(function (v) { return { value: v, label: t(prefix + v) }; }); };
      var notes = h('div', { class: 'flex flex-col gap-2' });
      if (d.prizes === false) notes.appendChild(notice('warning', t('se.noPrizes')));
      if (d.campaign === false) notes.appendChild(notice('secondary', t('se.noCampaign')));
      var p = cur.play, tm = cur.timing, sc = cur.scoring, rv = cur.reveal, rw = cur.rewards, lm = cur.limits, hn = cur.hints, rp = cur.reports, ev = cur.events, cp = cur.campaign, rt = cur.retention;
      var delivery = select(p.delivery, opt('se.delivery.', ['all', 'one_by_one']), { name: 'play.delivery' });
      var shuffleQ = check(p.shuffleQuestions, t('se.shuffleQ'), 'play.shuffleQuestions');
      var shuffleO = check(p.shuffleOptions, t('se.shuffleO'), 'play.shuffleOptions');
      var restart = select(p.onRestart, opt('se.restart.', ['resume', 'discard']), { name: 'play.onRestart' });
      var hours = numInput(p.sessionHours, 'play.sessionHours');
      var autoFinish = check(p.autoFinish, t('se.autoFinish'), 'play.autoFinish');
      var qsec = numInput(tm.questionSeconds, 'timing.questionSeconds');
      var grace = numInput(tm.graceSeconds, 'timing.graceSeconds');
      var late = select(tm.lateAnswer, opt('se.late.', ['wrong', 'reject']), { name: 'timing.lateAnswer' });
      var pass = numInput(sc.passPercent, 'scoring.passPercent');
      var multi = select(sc.multiCorrect, opt('se.multi.', ['any', 'all']), { name: 'scoring.multiCorrect' });
      var match = select(sc.textMatch, opt('se.match.', ['normalized', 'exact']), { name: 'scoring.textMatch' });
      var reveal = select(rv.mode, opt('se.reveal.', ['after_answer', 'after_finish', 'never']), { name: 'reveal.mode' });
      var before = check(rv.allowBeforeAnswer, t('se.allowBefore'), 'reveal.allowBeforeAnswer');
      var expl = check(rv.explanation, t('se.explanation'), 'reveal.explanation');
      var packMode = select(rw.packMode, opt('se.packMode.', ['first_pass', 'every_pass', 'none']), { name: 'rewards.packMode' });
      var qMode = select(rw.questionMode, opt('se.qMode.', ['first_correct', 'every_correct', 'none']), { name: 'rewards.questionMode' });
      var groupMode = select(rw.prizeGroupMode, opt('se.groupMode.', ['draw', 'all']), { name: 'rewards.prizeGroupMode' });
      var zone = input(rw.dayZone, { name: 'rewards.dayZone', ltr: true });
      var dailyPlays = numInput(lm.dailyPlays, 'limits.dailyPlays');
      var dailyPasses = numInput(lm.dailyRewardedPasses, 'limits.dailyRewardedPasses');
      var maxPlays = numInput(lm.maxPlaysPerPack, 'limits.maxPlaysPerPack');
      var replay = check(lm.replayAfterPass, t('se.replay'), 'limits.replayAfterPass');
      var cooldown = numInput(lm.retryCooldownSeconds, 'limits.retryCooldownSeconds');
      var pricing = select(hn.pricing, opt('se.pricing.', ['per_hint', 'fixed', 'free']), { name: 'hints.pricing' });
      var fixedCur = refSelect('currencies', (hn.fixedPrice || {}).currencyKey, 'hints.fixedPrice.currencyKey', t('none'));
      var fixedAmt = numInput((hn.fixedPrice || {}).amount || 0, 'hints.fixedPrice.amount');
      var keep = select(hn.keep, opt('se.keep.', ['forever', 'session']), { name: 'hints.keep' });
      var maxHints = numInput(hn.maxPerQuestion, 'hints.maxPerQuestion');
      var repOn = check(rp.enabled, t('se.reportsOn'), 'reports.enabled');
      var repDaily = numInput(rp.dailyLimit, 'reports.dailyLimit');
      var repSeen = check(rp.requireSeen, t('se.requireSeen'), 'reports.requireSeen');
      var repNote = check(rp.noteRequired, t('se.noteRequired'), 'reports.noteRequired');
      var repAuto = numInput(rp.autoDisableAt, 'reports.autoDisableAt');
      var evCompleted = input(ev.completed, { name: 'events.completed', ltr: true, mono: true });
      var evPassed = input(ev.passed, { name: 'events.passed', ltr: true, mono: true });
      var evCorrect = input(ev.correct, { name: 'events.correct', ltr: true, mono: true });
      var evAnalytics = input(ev.analytics, { name: 'events.analytics', ltr: true, mono: true });
      var evFail = check(ev.onFailure, t('se.onFailure'), 'events.onFailure');
      var cpOn = check(cp.enabled, t('se.campaignOn'), 'campaign.enabled');
      var cpFail = check(cp.sendFailures, t('se.sendFailures'), 'campaign.sendFailures');
      var keepDays = numInput(rt.sessionDays, 'retention.sessionDays');
      var delDays = numInput(rt.deliveryDays, 'retention.deliveryDays');
      var attempts = numInput(rt.webhookAttempts, 'retention.webhookAttempts');
      var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
      var duel = duelSettingsSection(cur);
      return {
        el: [notes, duel.el,
          section(t('se.play'), h('div', {}, grid(field(t('se.delivery'), delivery, t('se.hint.delivery')), field(t('se.restart'), restart)), shuffleQ, shuffleO,
            grid(field(t('se.hours'), hours, t('se.hint.hours')), h('div', {})), autoFinish), 'play'),
          section(t('se.timing'), h('div', {}, grid(field(t('se.qsec'), qsec, t('se.hint.qsec')), field(t('se.grace'), grace)), field(t('se.late'), late)), 'timing'),
          section(t('se.scoring'), h('div', {}, grid(field(t('se.pass'), pass, t('se.hint.pass')), field(t('se.multi'), multi)), field(t('se.match'), match, t('se.hint.match'))), 'scoring'),
          section(t('se.reveal'), h('div', {}, field(t('se.revealMode'), reveal), before, expl), 'reveal'),
          section(t('se.rewards'), h('div', {}, grid(field(t('se.packMode'), packMode), field(t('se.qMode'), qMode)), grid(field(t('se.groupMode'), groupMode), field(t('se.zone'), zone))), 'rewards'),
          section(t('se.limits'), h('div', {}, grid(field(t('se.dailyPlays'), dailyPlays, t('se.hint.zero')), field(t('se.dailyPasses'), dailyPasses, t('se.hint.dailyPasses'))),
            grid(field(t('se.maxPlays'), maxPlays, t('se.hint.zero')), field(t('se.cooldown'), cooldown)), replay), 'limits'),
          section(t('se.hints'), h('div', {}, grid(field(t('se.pricing'), pricing), field(t('se.keep'), keep)), grid(field(t('se.fixedCur'), fixedCur), field(t('se.fixedAmt'), fixedAmt)),
            field(t('se.maxHints'), maxHints, t('se.hint.zero'), 'qz-num')), 'hints'),
          section(t('se.reports'), h('div', {}, repOn, repSeen, repNote, grid(field(t('se.repDaily'), repDaily), field(t('se.repAuto'), repAuto, t('se.hint.repAuto')))), 'reports'),
          section(t('se.events'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('se.hint.events') }), grid(field(t('se.evCompleted'), evCompleted), field(t('se.evPassed'), evPassed)),
            grid(field(t('se.evCorrect'), evCorrect), field(t('se.evAnalytics'), evAnalytics)), evFail), 'events'),
          section(t('se.campaign'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('se.hint.campaign') }), cpOn, cpFail), 'campaign'),
          section(t('se.retention'), grid(field(t('se.keepDays'), keepDays, t('se.hint.zeroKeep')), field(t('se.delDays'), delDays), field(t('se.attempts'), attempts)), 'retention')],
        read: function (c) {
          c.play = { delivery: delivery.value, shuffleQuestions: shuffleQ.input.checked, shuffleOptions: shuffleO.input.checked, onRestart: restart.value, sessionHours: numberValue(hours), autoFinish: autoFinish.input.checked };
          c.timing = { questionSeconds: numberValue(qsec) || 0, graceSeconds: numberValue(grace) || 0, lateAnswer: late.value };
          c.scoring = { passPercent: floatValue(pass) == null ? 0 : floatValue(pass), multiCorrect: multi.value, textMatch: match.value };
          c.reveal = { mode: reveal.value, allowBeforeAnswer: before.input.checked, explanation: expl.input.checked };
          c.rewards = { packMode: packMode.value, questionMode: qMode.value, prizeGroupMode: groupMode.value, dayZone: zone.value.trim() };
          c.limits = { dailyPlays: numberValue(dailyPlays) || 0, dailyRewardedPasses: numberValue(dailyPasses) || 0, maxPlaysPerPack: numberValue(maxPlays) || 0, replayAfterPass: replay.input.checked,
            retryCooldownSeconds: numberValue(cooldown) || 0 };
          c.hints = { pricing: pricing.value, fixedPrice: { currencyKey: fixedCur.value, amount: numberValue(fixedAmt) || 0 }, keep: keep.value, maxPerQuestion: numberValue(maxHints) || 0 };
          c.reports = { enabled: repOn.input.checked, dailyLimit: numberValue(repDaily) || 0, requireSeen: repSeen.input.checked, noteRequired: repNote.input.checked, autoDisableAt: numberValue(repAuto) || 0 };
          c.events = { completed: evCompleted.value.trim(), passed: evPassed.value.trim(), correct: evCorrect.value.trim(), analytics: evAnalytics.value.trim(), onFailure: evFail.input.checked };
          c.campaign = { enabled: cpOn.input.checked, sendFailures: cpFail.input.checked };
          duel.read(c);
          c.retention = { sessionDays: numberValue(keepDays) || 0, deliveryDays: numberValue(delDays) || 0, webhookAttempts: numberValue(attempts) || 1 };
          return null;
        },
        after: function () { qzSettingsCache = null; }
      };
    });
  });
}

var QZ_PLAYER = { id: null, data: null, loading: false, at: 0, resetRewards: false };

function qzPlayerRoute() {
  var m = /^#\/players\/([0-9a-f-]{36})(?:\?.*)?$/.exec(location.hash || '');
  return m ? m[1] : null;
}

function qzLoadPlayer(id) {
  QZ_PLAYER.loading = true;
  return api('GET', '/quiz/players/' + id + qs({ lang: qzLang() })).then(function (d) { return d; }, function (err) { return { error: err }; }).then(function (d) {
    QZ_PLAYER.loading = false;
    QZ_PLAYER.at = Date.now();
    if (QZ_PLAYER.id === id) QZ_PLAYER.data = d;
    qzSchedule();
  });
}

function qzReloadPlayer() {
  QZ_PLAYER.at = 0;
  if (QZ_PLAYER.id) qzLoadPlayer(QZ_PLAYER.id);
}

function qzPlayerBody(id, d) {
  if (d.error) return notice('warning', t('pc.noQuiz'));
  var box = h('div', { class: 'flex flex-col gap-3' });
  box.appendChild(h('p', { class: 'text-sm', 'data-quiz-player-summary': '', text: t('pc.summary', { packs: (d.packs || []).length, hints: d.hintsBought, reports: d.reports }) }));
  var rewards = check(QZ_PLAYER.resetRewards, t('pc.reset.rewards'), 'reset-rewards');
  rewards.input.addEventListener('change', function () { QZ_PLAYER.resetRewards = rewards.input.checked; });
  if ((d.packs || []).length) {
    var list = h('ol', { class: 'flex flex-col gap-1', 'data-quiz-player-packs': '' });
    d.packs.forEach(function (p) {
      list.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-quiz-player-pack': p.key },
        h('span', { class: 'font-medium', text: p.name }), p.firstPassAt ? tag(t('pc.passed'), 'success') : tag(t('pc.notPassed'), 'secondary'),
        h('span', { class: 'text-xs text-secondary-500', text: t('pc.plays', { p: p.plays, w: p.wins, b: p.bestScore }) }),
        p.lastPlayedAt ? h('span', { class: 'text-xs text-secondary-500', text: fmtDate(p.lastPlayedAt) }) : null,
        button({ label: t('pc.reset'), variant: 'danger', data: { 'quiz-reset': p.key }, onclick: function () {
          confirmBox(t('pc.reset.confirm', { name: p.name }), t('pc.reset')).then(function (ok) {
            if (!ok) return;
            api('DELETE', '/quiz/players/' + id + '/packs/' + p.id + (QZ_PLAYER.resetRewards ? '?rewards=reset' : '')).then(function () { toast('success', t('saved')); qzReloadPlayer(); },
              function (err) { toast('error', qzErrText(err)); });
          });
        } })));
    });
    box.appendChild(list);
  } else box.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('pc.none') }));
  if ((d.missions || []).length) {
    var ml = h('ol', { class: 'flex flex-col gap-1', 'data-quiz-player-missions': '' });
    d.missions.forEach(function (m) {
      ml.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-xs' }, tag(t('ms.kind.' + m.kind)), h('span', { class: 'font-medium', text: m.name }),
        m.completedAt ? tag(t('pc.completed'), 'success') : tag(t('pc.started', { at: fmtDate(m.startedAt) })), m.rewarded ? tag(t('pc.rewarded'), 'primary') : null));
    });
    box.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pc.missions') }), ml));
  }
  if ((d.sessions || []).length) {
    var sl = h('ol', { class: 'flex flex-col gap-1', 'data-quiz-player-sessions': '' });
    d.sessions.forEach(function (s) {
      sl.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-xs' }, h('span', { class: 'text-secondary-500', text: fmtDate(s.startedAt) }), mono(s.packKey), statusTag(s.status),
        h('span', { text: t('pc.score', { c: s.correct, n: s.total, s: s.score }) }), s.status === 'done' ? tag(s.passed ? t('pc.passed') : t('pc.failed'), s.passed ? 'success' : 'error') : null));
    });
    box.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pc.sessions') }), sl));
  }
  box.appendChild(rewards);
  return box;
}

function qzPlayerCard() {
  var id = qzPlayerRoute();
  var existing = document.querySelector('[data-quiz-player]');
  if (!id || !token()) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (QZ_PLAYER.id !== id) { QZ_PLAYER.id = id; QZ_PLAYER.data = null; QZ_PLAYER.at = 0; }
  if (!QZ_PLAYER.loading && Date.now() - QZ_PLAYER.at > 20000) qzLoadPlayer(id);
  var d = QZ_PLAYER.data;
  var key = qzLang() + '|' + id + '|' + (d ? JSON.stringify(d) : 'loading');
  if (existing && existing.getAttribute('data-key') === key && existing.parentNode === grid) return;
  if (existing && existing.contains(document.activeElement) && existing.parentNode === grid) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' }, h('h2', { id: 'quiz-player-title', class: 'font-medium mb-3', text: t('pc.title') }));
  body.appendChild(d ? qzPlayerBody(id, d) : h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  grid.appendChild(h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-quiz-player': '', 'data-key': key, 'data-l10n-skip': '',
    'aria-labelledby': 'quiz-player-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body));
}

var QZ_ROOT = 'quiz-root';
var QZ_PAGES = ['packs', 'questions', 'missions', 'duels', 'duel-categories', 'reports', 'results', 'import', 'stats', 'settings'];
var QZ_RENDER = { packs: packsPage, questions: questionsPage, missions: missionsPage, duels: duelsPage, 'duel-categories': duelCategoriesPage, reports: reportsPage, results: resultsPage, 'import': importPage, stats: statsPage, settings: quizSettingsPage };
var QZ_HASH = /^#\/quiz(?:\/([^/?]*))?/;
var qzPending = null;
var qzRouter = null;
var qzMounted = null;
var qzScheduled = false;

function qzRoute(hash) {
  var m = QZ_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  return QZ_PAGES.indexOf(m[1]) !== -1 ? m[1] : QZ_PAGES[0];
}

function qzAllowed() {
  return !!token();
}

function qzTitle() {
  var page = qzRoute();
  return t('group') + (page ? ' - ' + t('nav.' + page) : '');
}

function qzAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || qzRouter) return !!qzRouter;
  try {
    if (!r.hasRoute('quiz')) {
      var layout = r.resolve('/').matched[0];
      r.addRoute({ path: '/quiz/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'quiz', component: { name: 'QuizPage', render: function () { return null; } }, meta: { get title() { return qzTitle(); } } }] });
    }
    qzRouter = r;
  } catch (e) {
    return false;
  }
  if (qzPending) {
    var target = qzPending;
    qzPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function qzNavLink(page, active) {
  var on = page === active;
  return h('li', {}, h('a', { href: '#/quiz/' + page, 'data-quiz-nav': page, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + page) }))));
}

function qzSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelector('[data-quiz-group]');
  if (!group || !qzAllowed()) { if (existing) existing.remove(); return; }
  var target = document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = qzLang() + '|' + active;
  if (existing && existing.getAttribute('data-key') === key && existing.nextElementSibling === target) return;
  var built = h('div', { 'data-quiz-group': '', 'data-plane-group': 'quiz', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group') }),
    h('nav', { 'aria-label': t('group') }, h('ul', { class: 'space-y-0.5' }, QZ_PAGES.map(function (p) { return qzNavLink(p, active); }))));
  if (existing) existing.remove();
  target.before(built);
}

function qzShowMains() {
  var hidden = document.querySelectorAll('[data-quiz-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-quiz-hidden'); }
}

function qzUnmount() {
  var root = document.getElementById(QZ_ROOT);
  if (root) root.remove();
  qzShowMains();
  closePanel();
  qzMounted = null;
}

function qzSync() {
  qzAddRoutes();
  var page = qzRoute();
  qzSyncNav(page);
  qzPlayerCard();
  if (!page) { if (qzMounted || document.getElementById(QZ_ROOT)) qzUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-quiz-hidden', ''); }
  }
  var root = document.getElementById(QZ_ROOT);
  if (!root) {
    root = h('div', { id: QZ_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-quiz': 'root' });
    qzMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + qzLang() + (page === 'questions' || page === 'reports' || page === 'stats' || page === 'duels' ? '|' + location.hash : '');
  if (qzMounted !== key) {
    qzMounted = key;
    closePanel();
    QZ_RENDER[page](root);
  }
}

function qzSchedule() {
  if (qzScheduled) return;
  qzScheduled = true;
  requestAnimationFrame(function () { qzScheduled = false; qzSync(); });
}

function qzBoot() {
  window.addEventListener('hashchange', qzSchedule);
  window.addEventListener('popstate', qzSchedule);
  window.addEventListener('nakama-console:ready', qzSchedule);
  new MutationObserver(qzSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { qzMounted = null; qzSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (qzAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  qzSync();
}

window.NakamaQuiz = { version: 1, pages: function () { return QZ_PAGES.map(function (p) { return { id: p, href: '#/quiz/' + p, label: t('nav.' + p) }; }); } };

if (QZ_HASH.test(location.hash)) {
  qzPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', qzBoot);
else qzBoot();
})();
