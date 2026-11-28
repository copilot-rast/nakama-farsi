(function () {
'use strict';
var CU_TEXT = {
  fa: {
    'group.cups': 'تورنمنت‌ها',
    'nav.cups/list': 'فهرست تورنمنت‌ها', 'nav.cups/templates': 'تورنمنت‌های تکراری', 'nav.cups/settings': 'تنظیمات تورنمنت', 'nav.cups/view': 'جدول تورنمنت',
    'desc.cups/list': 'تورنمنت‌های حذفی: بازیکن‌ها ثبت‌نام می‌کنند، سرور جدول را می‌چیند و برنده‌ها تا فینال بالا می‌روند. روی یک تورنمنت بزنید تا جدول زنده و کارهای مدیر را ببینید.',
    'desc.cups/templates': 'قالب تورنمنت‌هایی که خودشان تکرار می‌شوند. با زمان‌بندی CRON، سرور تورنمنت بعدی را خودش می‌سازد؛ هر قالب در هر زمان فقط یک تورنمنت باز دارد.',
    'desc.cups/settings': 'روشن و خاموش کردن تورنمنت‌ها، اعلان‌ها، سقف‌ها و قانون‌های پیش‌فرض تورنمنت تازه.',
    'desc.cups/view': '',
    'save': 'ذخیره', 'saved': 'ذخیره شد', 'deleted': 'حذف شد', 'done': 'انجام شد', 'cancel': 'انصراف', 'close': 'بستن', 'remove': 'حذف', 'delete': 'حذف', 'search': 'جستجو', 'all': 'همه',
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'page.first': 'صفحهٔ اول', 'page.last': 'صفحهٔ آخر', 'page.of': 'صفحهٔ {page} از {pages}', 'pager': 'صفحه‌بندی', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'yes': 'بله', 'no': 'خیر',
    'active': 'فعال', 'inactive': 'غیرفعال', 'confirm.title': 'مطمئن هستید؟', 'lang.text': '{label} ({lang})', 'updated': 'آخرین تغییر: {by} · {at}', 'none': '—', 'missing': 'پیدا نشد',
    'f.name': 'نام', 'f.description': 'توضیح', 'f.key': 'کلید', 'f.icon': 'نشانی آیکون', 'f.color': 'رنگ', 'f.state': 'وضعیت', 'f.player': 'بازیکن', 'f.date': 'زمان', 'f.updated': 'آخرین تغییر',
    'f.sort': 'ترتیب', 'f.id': 'شناسه',
    'hint.key': 'حرف کوچک انگلیسی، عدد و _ . - ؛ بازی با همین کلید کار می‌کند.', 'hint.tags': 'با ویرگول جدا کنید', 'hint.zero': '۰ یعنی بدون محدودیت',
    'st.draft': 'پیش‌نویس', 'st.upcoming': 'به‌زودی', 'st.signup': 'ثبت‌نام باز', 'st.running': 'در حال برگزاری', 'st.finished': 'تمام‌شده', 'st.cancelled': 'لغوشده',
    'mode.window': 'بازهٔ زمانی', 'mode.live': 'زنده',
    'ms.pending': 'منتظر دور قبل', 'ms.waiting': 'در انتظار شروع', 'ms.open': 'باز', 'ms.playing': 'در حال بازی', 'ms.finished': 'تمام‌شده', 'ms.bye': 'استراحت (بای)', 'ms.void': 'باطل',
    'ps.registered': 'ثبت‌نام‌شده', 'ps.playing': 'در جدول', 'ps.eliminated': 'حذف‌شده', 'ps.finished': 'تمام کرده', 'ps.disqualified': 'محروم', 'ps.withdrawn': 'انصراف داده', 'ps.removed': 'برداشته‌شده',
    'rs.pending': 'شروع نشده', 'rs.running': 'در جریان', 'rs.finished': 'تمام‌شده',
    'fs.none': 'بدون ورودی', 'fs.paid': 'پرداخت‌شده', 'fs.refunded': 'برگشت داده شد',
    'pay.pending': 'در صف', 'pay.paid': 'پرداخت شد', 'pay.failed': 'ناموفق',
    'po.champion': 'قهرمان', 'po.runner_up': 'نایب‌قهرمان', 'po.semi': 'نیمه‌نهایی', 'po.third': 'سوم', 'po.fourth': 'چهارم', 'po.reached': 'رسیدن به دور {n}',
    'pl.champion': 'قهرمان', 'pl.runnerUp': 'نایب‌قهرمان', 'pl.semi': 'نیمه‌نهایی‌ها', 'pl.third': 'سوم', 'pl.fourth': 'چهارم',
    'rk.final': 'فینال', 'rk.semi': 'نیمه‌نهایی', 'rk.quarter': 'یک‌چهارم نهایی', 'rk.third': 'رده‌بندی (مقام سوم)', 'rk.of': 'دور {n} نفره',
    'why.played': 'بازی شد', 'why.draw': 'مساوی', 'why.double_no_show': 'هیچ‌کدام نیامدند', 'why.forfeit': 'باخت با غیبت', 'why.unplayed': 'بازی تمام نشد', 'why.admin': 'تصمیم مدیر',
    'why.deadline': 'پایان وقت دور', 'why.disqualified': 'محرومیت حریف', 'why.no_show': 'غیبت حریف', 'why.walkover': 'بدون حریف', 'why.empty': 'هر دو جا خالی', 'why.timeout': 'نتیجه نیامد',
    'why.launch_failed': 'بازی ساخته نشد', 'why.bye': 'استراحت', 'why.underfill': 'بازیکن کافی ثبت‌نام نکرد', 'why.schedule': 'زمان‌بندی',
    'cu.new': 'تورنمنت تازه', 'cu.edit': 'ویرایش', 'cu.create': 'ساختن', 'cu.created': 'تورنمنت ساخته شد', 'cu.empty': 'هنوز تورنمنتی نساخته‌اید. با «تورنمنت تازه» شروع کنید.',
    'cu.hint.search': 'نام، کلید قالب یا شناسه', 'cu.byTemplate': 'تورنمنت‌های قالب', 'cu.allCups': 'همهٔ تورنمنت‌ها',
    'cu.players': 'بازیکن‌ها', 'cu.starts': 'شروع', 'cu.round': 'دور', 'cu.entrantsV': '{n} از {cap}', 'cu.roundV': 'دور {n} از {of}',
    'cu.signupAt': 'شروع ثبت‌نام', 'cu.startsAt': 'شروع تورنمنت', 'cu.announceAt': 'نمایش در فهرست از', 'cu.startedAt': 'شروع شد', 'cu.finishedAt': 'تمام شد',
    'cu.hint.signupAt': 'از این زمان بازیکن‌ها می‌توانند ثبت‌نام کنند.', 'cu.hint.startsAt': 'ثبت‌نام بسته می‌شود و جدول چیده می‌شود.', 'cu.hint.announceAt': 'خالی یعنی از همان شروع ثبت‌نام دیده شود.',
    'cu.publish': 'همین حالا منتشر شود', 'cu.lockedRules': 'تورنمنت شروع شده است؛ فقط نام، توضیح و آیکون عوض می‌شود.', 'cu.needName': 'دست‌کم یک نام بنویسید.', 'cu.needTimes': 'زمان ثبت‌نام و شروع را بگذارید.',
    'cu.back': '← فهرست تورنمنت‌ها', 'cu.liveOn': 'هر ۵ ثانیه تازه می‌شود', 'cu.refreshed': 'به‌روز در {at}',
    'cu.publishNow': 'انتشار', 'cu.published': 'منتشر شد', 'cu.startNow': 'شروع همین حالا', 'cu.startConfirm': 'تورنمنت با {n} بازیکن همین حالا شروع شود؟ ثبت‌نام بسته می‌شود.', 'cu.started': 'تورنمنت شروع شد',
    'cu.addPlayer': 'افزودن بازیکن', 'cu.added': 'بازیکن اضافه شد', 'cu.cancel': 'لغو و برگشت پول', 'cu.cancelTitle': 'لغو تورنمنت', 'cu.cancelled': 'تورنمنت لغو شد و ورودی‌ها برگشت داده شد',
    'cu.hint.cancel': 'تورنمنت لغو می‌شود، بازی‌های باز باطل می‌شوند و ورودی همهٔ بازیکن‌ها برمی‌گردد. جایزه‌ای که پرداخت شده برنمی‌گردد.', 'cu.cancelReason': 'دلیل لغو', 'cu.hint.reason': 'مثلاً مشکل سرور (اختیاری)',
    'cu.deleteConfirm': 'این تورنمنت برای همیشه حذف شود؟',
    'cu.size': 'اندازهٔ جدول', 'cu.sizeV': '{n} جایگاه · {rounds} دور', 'tp.template': 'قالب',
    'cu.bracket': 'جدول', 'cu.noBracket': 'جدول وقتی تورنمنت شروع شود چیده می‌شود.', 'cu.noBracketCancelled': 'این تورنمنت پیش از چیدن جدول لغو شد.',
    'cu.tbd': 'معلوم نیست', 'cu.ready': 'آماده است', 'cu.matchOf': '{a} در برابر {b}', 'cu.match': 'مسابقه', 'cu.sideA': 'بازیکن اول', 'cu.sideB': 'بازیکن دوم', 'cu.winsN': '{n} برد',
    'cu.winner': 'برنده', 'cu.reason': 'چطور تمام شد', 'cu.draws': 'مساوی‌ها', 'cu.opensAt': 'باز شد', 'cu.matchStarts': 'شروع', 'cu.closesAt': 'مهلت', 'cu.gameMatch': 'شناسهٔ بازی',
    'cu.games': 'بازی‌ها', 'cu.noGames': 'هنوز بازی‌ای انجام نشده است.', 'cu.gameN': 'بازی {n}', 'cu.draw': 'مساوی', 'cu.wonBy': 'برد {name}',
    'cu.force': 'نتیجه به دست مدیر', 'cu.hint.force': 'بازی باز باطل می‌شود و برنده‌ای که انتخاب کنید به دور بعد می‌رود.', 'cu.winIs': 'برنده: {name}',
    'cu.forceConfirm': '{name} برندهٔ این مسابقه شود؟', 'cu.forceOk': 'ثبت نتیجه', 'cu.forced': 'نتیجه ثبت شد',
    'cu.voidMatch': 'باطل کردن مسابقه', 'cu.voidConfirm': 'این مسابقه باطل شود؟ هیچ‌کدام بالا نمی‌روند و جای بعدی خالی می‌ماند.', 'cu.voided': 'مسابقه باطل شد',
    'cu.playersT': 'بازیکن‌ها ({n})', 'cu.seed': 'رتبه', 'cu.trophies': 'جام', 'cu.place': 'جایگاه', 'cu.fee': 'ورودی', 'cu.actions': 'کارها', 'cu.placeN': 'جایگاه {n}', 'cu.reachedV': 'تا دور {n}',
    'cu.noPlayersYet': 'هنوز کسی ثبت‌نام نکرده است.', 'cu.noPlayers': 'بازیکنی نیست.', 'cu.goneN': '{n} بازیکن انصراف داده یا برداشته شده است.',
    'cu.removeConfirm': '{name} از تورنمنت برداشته شود؟ ورودی‌اش کامل برمی‌گردد.', 'cu.removed': 'بازیکن برداشته شد',
    'cu.swap': 'جابه‌جایی', 'cu.swapTitle': 'جابه‌جایی {name}', 'cu.swapTo': 'بازیکن تازه', 'cu.swapped': 'بازیکن جابه‌جا شد',
    'cu.hint.swap': 'بازیکن تازه جای همین بازیکن را در جدول می‌گیرد. فقط پیش از شروع یا پیش از اولین بازی‌اش. ورودی بازیکن قبلی برمی‌گردد.',
    'cu.disqualify': 'محروم کردن', 'cu.dqTitle': 'محروم کردن {name}', 'cu.dqRefund': 'ورودی‌اش برگردد', 'cu.disqualified': 'بازیکن محروم شد',
    'cu.hint.dq': 'بازیکن از تورنمنت بیرون می‌رود و حریفش برنده می‌شود. جایزه‌ای نمی‌گیرد.',
    'cu.hint.who': 'نام کاربری یا شناسه', 'cu.hint.whoLong': 'نام کاربری دقیق یا شناسهٔ کامل بازیکن', 'cu.hint.add': 'بازیکن بدون پرداخت ورودی و بدون بررسی شرط‌ها اضافه می‌شود.',
    'cu.payouts': 'جایزه‌ها', 'cu.kind': 'برای', 'cu.amounts': 'مبلغ از صندوق', 'cu.retry': 'تلاش دوباره برای جایزه‌های ناموفق', 'cu.retried': '{n} جایزه دوباره در صف رفت',
    'tp.new': 'قالب تازه', 'tp.edit': 'ویرایش قالب', 'tp.empty': 'هنوز قالب تکراری ندارید.', 'tp.schedule': 'زمان‌بندی (CRON)', 'tp.nextStart': 'شروع بعدی', 'tp.openCups': 'تورنمنت باز', 'tp.manual': 'دستی',
    'tp.active': 'فعال باشد و تورنمنت‌ها را خودش بسازد', 'tp.signupMinutes': 'ثبت‌نام (دقیقه)', 'tp.announceMinutes': 'نمایش زودتر (دقیقه)',
    'tp.hint': 'در هر زمان زمان‌بندی، تورنمنت تازه‌ای با همین قانون‌ها ساخته می‌شود. ثبت‌نام آن چند دقیقه پیش از شروع باز می‌شود.',
    'tp.hint.schedule': 'پنج بخش: دقیقه ساعت روز ماه روز‌هفته، به وقت سرور. مثال: 0 18 * * 5 یعنی هر جمعه ساعت ۱۸.', 'tp.hint.signup': 'چند دقیقه پیش از شروع ثبت‌نام باز شود.',
    'tp.hint.announce': 'چند دقیقه پیش از ثبت‌نام در فهرست دیده شود.', 'tp.runNow': 'ساختن یکی همین حالا', 'tp.runConfirm': 'یک تورنمنت از این قالب همین حالا ساخته شود؟',
    'tp.deleteConfirm': 'قالب {key} حذف شود؟ تورنمنت‌هایی که از آن ساخته شده‌اند می‌مانند.',
    'cs.s.general': 'کلی', 'cs.enabled': 'تورنمنت‌ها روشن باشند', 'cs.hint.enabled': 'خاموش که باشد، بازیکن‌ها نمی‌توانند ثبت‌نام کنند ولی تورنمنت‌های در جریان تمام می‌شوند.',
    'cs.s.events': 'اعلان‌ها', 'cs.notify': 'اعلان سوکت برای بازیکن‌ها', 'cs.code': 'کد اعلان', 'cs.hint.code': 'کدی که اپ برای شناختن اعلان تورنمنت می‌خواند.',
    'cs.inApp': 'پیام درون‌برنامه برای ثبت‌نام، پایان، جایزه و لغو', 'cs.publish': 'رویداد بازی برای مسابقه و پایان (برای ماموریت‌ها و آمار)',
    'cs.s.limits': 'سقف‌ها', 'cs.maxCapacity': 'بیشترین ظرفیت', 'cs.hint.maxCapacity': 'بیشترین بازیکن یک تورنمنت.', 'cs.activePerPlayer': 'تورنمنت باز هم‌زمان برای هر بازیکن',
    'cs.tickSeconds': 'فاصلهٔ بررسی (ثانیه)', 'cs.hint.tick': 'هر چند ثانیه سرور زمان‌ها و مهلت‌ها را بررسی کند.', 'cs.listLimit': 'تعداد در فهرست', 'cs.historyLimit': 'تعداد در تاریخچه',
    'cs.finishedDays': 'نمایش تمام‌شده‌ها (روز)', 'cs.hint.finishedDays': 'تورنمنت تمام‌شده تا چند روز در فهرست بازیکن بماند.',
    'cs.s.defaults': 'قانون‌های پیش‌فرض', 'cs.hint.defaults': 'تورنمنت یا قالب تازه با این قانون‌ها باز می‌شود. روی تورنمنت‌های ساخته‌شده اثری ندارد.',
    'ru.s.game': 'بازی', 'ru.game': 'بازی', 'ru.external': 'بیرونی (نتیجه با cup_report)', 'ru.hint.game': 'بازی‌ای که مسابقه‌ها در آن انجام می‌شود.',
    'ru.rated': 'جام و امتیاز بازی هم حساب شود', 'ru.gameOptions': 'تنظیم بازی (JSON)', 'ru.hint.gameOptions': 'همین‌طور به بازی داده می‌شود، مثل {"goals":3}',
    'ru.s.players': 'بازیکن‌ها و جدول', 'ru.capacity': 'ظرفیت', 'ru.hint.capacity': 'هر عددی می‌شود؛ جای خالی جدول با استراحت (بای) پر می‌شود.',
    'ru.minPlayers': 'کمترین بازیکن', 'ru.hint.minPlayers': 'اگر موقع شروع کمتر باشد، تورنمنت لغو و ورودی‌ها برگشت داده می‌شود.',
    'ru.underfill': 'اگر پر نشد', 'ru.under.shrink': 'با همین تعداد برگزار شود', 'ru.under.cancel': 'لغو شود', 'ru.hint.underfill': 'وقتی بیشتر از کمترین ولی کمتر از ظرفیت ثبت‌نام کرده‌اند.',
    'ru.seeding': 'چیدن جدول', 'ru.seed.random': 'تصادفی', 'ru.seed.trophies': 'بر اساس جام', 'ru.hint.seeding': 'با جام، بهترین‌ها فقط در فینال به هم می‌رسند و استراحت‌ها به آن‌ها می‌رسد.',
    'ru.thirdPlace': 'بازی رده‌بندی برای مقام سوم',
    'ru.s.bestOf': 'تعداد بازی هر مسابقه', 'ru.bestOf': 'بازی در هر مسابقه', 'ru.bestOfN': 'بهترین از {n}', 'ru.hint.bestOf': 'هر کس زودتر به بیش از نصف بردها برسد بالا می‌رود.',
    'ru.roundBestOf': 'برای دورهای خاص', 'ru.hint.roundBestOf': 'مثلاً فینال بهترین از ۳.', 'ru.addRoundBestOf': 'افزودن دور', 'ru.round': 'دور', 'ru.roundN': 'دور {n}',
    'ru.s.timing': 'زمان‌بندی', 'ru.mode': 'نوع زمان‌بندی',
    'ru.hint.mode.window': 'هر دور مدتی باز است و دو بازیکن هر وقت هر دو آماده شدند بازی می‌کنند.',
    'ru.hint.mode.live': 'هر دور در ساعت مشخص شروع می‌شود؛ بازیکن باید پیش از آن حضورش را بزند.',
    'ru.windowMinutes': 'مدت هر دور (دقیقه)', 'ru.hint.windowMinutes': 'بعد از آن، دور با قانون غیبت بسته می‌شود.', 'ru.readySeconds': 'اعتبار آمادگی (ثانیه)', 'ru.hint.readySeconds': 'اگر حریف در این مدت آماده نشود، آمادگی باید دوباره زده شود.',
    'ru.checkInMinutes': 'حضور (دقیقه پیش از شروع)', 'ru.hint.checkInMinutes': 'از چند دقیقه پیش از هر دور حضور باز شود.', 'ru.roundMinutes': 'فاصلهٔ دورها (دقیقه)', 'ru.hint.roundMinutes': 'هر دور چند دقیقه بعد از قبلی شروع شود.',
    'ru.joinSeconds': 'مهلت ورود به بازی (ثانیه)', 'ru.hint.joinSeconds': 'اگر بازیکن در این مدت وارد بازی نشود، غایب است.', 'ru.gameMinutes': 'بیشترین مدت بازی (دقیقه)', 'ru.hint.gameMinutes': 'بازی‌ای که نتیجه‌اش تا این زمان نیاید باطل می‌شود.',
    'ru.noShow': 'اگر هیچ‌کدام نیامدند', 'ru.hint.noShow': 'وقتی دور تمام شد و هیچ‌کدام حاضر نشدند.', 'ru.unplayed': 'اگر آمدند ولی تمام نکردند', 'ru.hint.unplayed': 'وقتی هر دو آمدند ولی مسابقه تا پایان دور تمام نشد.',
    'ru.ns.both_out': 'هر دو حذف شوند', 'ru.ns.higher_seed': 'رتبهٔ بهتر برنده شود', 'ru.ns.random': 'قرعه',
    'ru.s.draw': 'مساوی', 'ru.drawRule': 'اگر بازی مساوی شد', 'ru.draw.replay': 'دوباره بازی کنند', 'ru.draw.higher_seed': 'رتبهٔ بهتر برنده شود', 'ru.draw.random': 'قرعه',
    'ru.maxReplays': 'بیشترین تکرار', 'ru.fallback': 'بعد از تکرارها', 'ru.gameDecides': 'بازی خودش مساوی را بشکند (مثل گل طلایی)',
    'ru.s.entry': 'ورود', 'ru.feeCurrency': 'ارز ورودی', 'ru.noCurrency': 'رایگان', 'ru.feeAmount': 'ورودی', 'ru.hint.fee': 'موقع ثبت‌نام کم می‌شود.',
    'ru.withdrawRefund': 'برگشت ورودی با انصراف (درصد)', 'ru.hint.withdrawRefund': 'اگر بازیکن پیش از شروع انصراف دهد.',
    'ru.minTrophies': 'کمترین جام', 'ru.maxTrophies': 'بیشترین جام', 'ru.leagues': 'فقط این لیگ‌ها', 'ru.hint.leagues': 'هیچ‌کدام یعنی همهٔ لیگ‌ها.',
    'ru.tagIds': 'برچسب لازم', 'ru.tagsMode': 'از برچسب‌های لازم', 'ru.tagsMode.any': 'یکی کافی است', 'ru.tagsMode.all': 'همه لازم است', 'ru.excludeTagIds': 'برچسب ممنوع', 'ru.noneYet': 'چیزی ساخته نشده است.',
    'ru.s.prizes': 'جایزه‌ها', 'ru.noPrize': 'بدون جایزه', 'ru.prize': 'جایزه', 'ru.hint.semi': 'به هر دو بازندهٔ نیمه‌نهایی.',
    'ru.reached': 'جایزهٔ رسیدن به دور', 'ru.hint.reached': 'به هر کس که دست‌کم به این دور برسد.', 'ru.reachedRound': 'از دور', 'ru.addReached': 'افزودن جایزهٔ دور',
    'ru.pool': 'صندوق جایزه', 'ru.hint.pool': 'درصدی از ورودی‌ها به‌اضافهٔ مبلغ پایه، با همان ارز ورودی، بین برنده‌ها پخش می‌شود. سهم‌ها درصد صندوق هستند.',
    'ru.poolPercent': 'درصد از ورودی‌ها', 'ru.poolBase': 'مبلغ پایه',
    'ru.badNumber': '«{field}» باید عدد باشد.', 'ru.badJson': 'تنظیم بازی باید یک شیء JSON باشد.', 'ru.badReached': 'برای هر جایزهٔ دور، دور و جایزه را انتخاب کنید.',
    'ru.capN': '{n} نفر', 'ru.free': 'رایگان', 'ru.feeV': 'ورودی {amount} {currency}',
    'code.CUP_ALREADY_OVER': "این تورنمنت تمام شده است.",
    'code.CUP_ALREADY_SIGNED_UP': "قبلاً در این تورنمنت ثبت‌نام کرده‌ای.",
    'code.CUP_BAD_STATUS': "فیلتر فهرست تورنمنت‌ها درست نیست.",
    'code.CUP_BUSY': "تورنمنت در حال به‌روزرسانی است. چند لحظه بعد دوباره امتحان کن.",
    'code.CUP_CAPACITY_BELOW_PLAYERS': "ظرفیت نمی‌تواند از تعداد ثبت‌نام‌شده‌ها کمتر باشد.",
    'code.CUP_CHECK_IN_NOT_OPEN': "زمان اعلام حضور این بازی هنوز نرسیده است.",
    'code.CUP_DELETE_ONLY_DRAFT_OR_CANCELLED': "فقط تورنمنت پیش‌نویس یا لغوشده را می‌شود حذف کرد.",
    'code.CUP_DISABLED': "تورنمنت‌ها فعلاً خاموش هستند.",
    'code.CUP_FULL': "ظرفیت این تورنمنت پر شده است.",
    'code.CUP_GAME_ID_REQUIRED': "شناسه بازی لازم است.",
    'code.CUP_GAME_NOT_FOUND': "این بازی به هیچ تورنمنتی تعلق ندارد.",
    'code.CUP_GAME_UNKNOWN': "این بازی مسابقه تورنمنت را پشتیبانی نمی‌کند.",
    'code.CUP_ID_REQUIRED': "شناسه تورنمنت لازم است.",
    'code.CUP_LEAGUE_REQUIRED': "لیگ تو اجازه ورود به این تورنمنت را ندارد.",
    'code.CUP_MATCH_NOT_FOUND': "بازی تورنمنت پیدا نشد.",
    'code.CUP_MATCH_NOT_READY': "هنوز هر دو بازیکن این بازی مشخص نشده‌اند.",
    'code.CUP_MATCH_OVER': "نتیجه این بازی قبلاً مشخص شده است.",
    'code.CUP_NOT_A_PLAYER': "تو در این تورنمنت بازی نمی‌کنی.",
    'code.CUP_NOT_DRAFT': "این تورنمنت قبلاً منتشر شده است.",
    'code.CUP_NOT_ENOUGH_CURRENCY': "برای ورودی تورنمنت {amount} {currencyKey} لازم داری.",
    'code.CUP_NOT_FOUND': "تورنمنت پیدا نشد.",
    'code.CUP_NOT_RUNNING': "این تورنمنت در حال برگزاری نیست.",
    'code.CUP_NOT_SIGNED_UP': "در این تورنمنت ثبت‌نام نکرده‌ای.",
    'code.CUP_NOT_STARTABLE': "این تورنمنت را الان نمی‌شود شروع کرد.",
    'code.CUP_NO_MATCH': "در این تورنمنت بازی‌ای نداری.",
    'code.CUP_REFUNDS_PENDING': "هنوز ورودی بعضی بازیکن‌ها برنگشته است.",
    'code.CUP_REMOVED': "از این تورنمنت کنار گذاشته شده‌ای.",
    'code.CUP_SCHEDULE_REQUIRED': "قالب فعال باید زمان‌بندی داشته باشد.",
    'code.CUP_SERVER_ONLY': "نتیجه تورنمنت را فقط سرور بازی می‌فرستد.",
    'code.CUP_SIGNUP_AFTER_START': "ثبت‌نام باید قبل از شروع تورنمنت باز شود.",
    'code.CUP_SIGNUP_CLOSED': "ثبت‌نام این تورنمنت بسته شده است.",
    'code.CUP_SIGNUP_NOT_OPEN': "ثبت‌نام از {signupAt} باز می‌شود.",
    'code.CUP_START_IN_PAST': "زمان شروع گذشته است.",
    'code.CUP_SWAP_TOO_LATE': "دیگر نمی‌شود این بازیکن را عوض کرد.",
    'code.CUP_TAG_EXCLUDED': "بازیکن‌های دارای برچسب تو نمی‌توانند وارد این تورنمنت شوند.",
    'code.CUP_TAG_REQUIRED': "برای این تورنمنت برچسب ویژه (مثلاً VIP) لازم است.",
    'code.CUP_TEMPLATE_NOT_FOUND': "قالب تورنمنت پیدا نشد.",
    'code.CUP_TIMES_REQUIRED': "زمان ثبت‌نام و شروع لازم است.",
    'code.CUP_TOO_MANY_ACTIVE': "الان در {max} تورنمنت هستی و بیشتر از این نمی‌شود.",
    'code.CUP_TROPHIES_TOO_HIGH': "برای این تورنمنت حداکثر {max} جام مجاز است.",
    'code.CUP_TROPHIES_TOO_LOW': "برای این تورنمنت دست‌کم {min} جام لازم است.",
    'code.CUP_USE_DISQUALIFY': "تورنمنت شروع شده؛ به‌جای حذف، بازیکن را رد صلاحیت کن.",
    'code.CUP_WAITING_FOR_OPPONENT': "حریف بعدی‌ات هنوز مشخص نشده است.",
    'code.CUP_WINNER_NOT_IN_MATCH': "برنده باید یکی از دو بازیکن باشد.",
    'code.CUP_WITHDRAW_CLOSED': "دیگر نمی‌شود از این تورنمنت انصراف داد.",
    'code.INTERNAL_ERROR': "مشکلی پیش آمد. دوباره امتحان کنید.",
    'code.INVALID_KEY': "کلید یا نام درست نیست.",
    'code.INVALID_POLICY': "مقدار انتخاب‌شده مجاز نیست.",
    'code.INVALID_SCHEDULE': "زمان‌بندی CRON درست نیست.",
    'code.KEY_TAKEN': "این کلید قبلاً استفاده شده است.",
    'code.PLAYER_NOT_FOUND': "بازیکن پیدا نشد.",
    'code.VALUE_OUT_OF_RANGE': "مقدار خارج از بازهٔ مجاز است."
  },
  en: {
    'group.cups': 'Tournaments',
    'nav.cups/list': 'Tournaments', 'nav.cups/templates': 'Recurring tournaments', 'nav.cups/settings': 'Tournament settings', 'nav.cups/view': 'Tournament bracket',
    'desc.cups/list': 'Knockout tournaments: players sign up, the server builds the bracket and winners move up to the final. Open a tournament to see its live bracket and the admin actions.',
    'desc.cups/templates': 'Tournaments that repeat by themselves. With a CRON schedule the server creates the next tournament; each template has at most one open tournament.',
    'desc.cups/settings': 'Turn tournaments on or off, notifications, limits and the default rules for a new tournament.',
    'desc.cups/view': '',
    'save': 'Save', 'saved': 'Saved', 'deleted': 'Deleted', 'done': 'Done', 'cancel': 'Cancel', 'close': 'Close', 'remove': 'Remove', 'delete': 'Delete', 'search': 'Search', 'all': 'All',
    'loading': 'Loading…', 'empty': 'Nothing found', 'page.first': 'First page', 'page.last': 'Last page', 'page.of': 'Page {page} of {pages}', 'pager': 'Pagination', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'yes': 'Yes', 'no': 'No',
    'active': 'Active', 'inactive': 'Inactive', 'confirm.title': 'Are you sure?', 'lang.text': '{label} ({lang})', 'updated': 'Last change: {by} · {at}', 'none': '—', 'missing': 'missing',
    'f.name': 'Name', 'f.description': 'Description', 'f.key': 'Key', 'f.icon': 'Icon URL', 'f.color': 'Colour', 'f.state': 'Status', 'f.player': 'Player', 'f.date': 'Time', 'f.updated': 'Last change',
    'f.sort': 'Order', 'f.id': 'ID',
    'hint.key': 'Lowercase letters, digits and _ . - ; the game uses this key.', 'hint.tags': 'Separate with commas', 'hint.zero': '0 means no limit',
    'st.draft': 'Draft', 'st.upcoming': 'Upcoming', 'st.signup': 'Sign-up open', 'st.running': 'Running', 'st.finished': 'Finished', 'st.cancelled': 'Cancelled',
    'mode.window': 'Time window', 'mode.live': 'Live',
    'ms.pending': 'Waiting', 'ms.waiting': 'Waiting to start', 'ms.open': 'Open', 'ms.playing': 'Playing', 'ms.finished': 'Finished', 'ms.bye': 'Bye', 'ms.void': 'Void',
    'ps.registered': 'Signed up', 'ps.playing': 'In the bracket', 'ps.eliminated': 'Knocked out', 'ps.finished': 'Finished', 'ps.disqualified': 'Disqualified', 'ps.withdrawn': 'Withdrew', 'ps.removed': 'Removed',
    'rs.pending': 'Not started', 'rs.running': 'In progress', 'rs.finished': 'Finished',
    'fs.none': 'No fee', 'fs.paid': 'Paid', 'fs.refunded': 'Refunded',
    'pay.pending': 'Queued', 'pay.paid': 'Paid', 'pay.failed': 'Failed',
    'po.champion': 'Champion', 'po.runner_up': 'Runner-up', 'po.semi': 'Semi-final', 'po.third': 'Third', 'po.fourth': 'Fourth', 'po.reached': 'Reached round {n}',
    'pl.champion': 'Champion', 'pl.runnerUp': 'Runner-up', 'pl.semi': 'Semi-finalists', 'pl.third': 'Third', 'pl.fourth': 'Fourth',
    'rk.final': 'Final', 'rk.semi': 'Semi-final', 'rk.quarter': 'Quarter-final', 'rk.third': 'Third-place match', 'rk.of': 'Round of {n}',
    'why.played': 'Played', 'why.draw': 'Draw', 'why.double_no_show': 'Neither showed up', 'why.forfeit': 'Forfeit', 'why.unplayed': 'Not finished', 'why.admin': 'Admin decision',
    'why.deadline': 'Round deadline', 'why.disqualified': 'Opponent disqualified', 'why.no_show': 'Opponent did not show up', 'why.walkover': 'Walkover', 'why.empty': 'Both slots empty', 'why.timeout': 'No result in time',
    'why.launch_failed': 'Game could not start', 'why.bye': 'Bye', 'why.underfill': 'Not enough players signed up', 'why.schedule': 'Schedule',
    'cu.new': 'New tournament', 'cu.edit': 'Edit', 'cu.create': 'Create', 'cu.created': 'Tournament created', 'cu.empty': 'No tournaments yet. Start with «New tournament».',
    'cu.hint.search': 'Name, template key or ID', 'cu.byTemplate': 'Tournaments of template', 'cu.allCups': 'All tournaments',
    'cu.players': 'Players', 'cu.starts': 'Starts', 'cu.round': 'Round', 'cu.entrantsV': '{n} of {cap}', 'cu.roundV': 'Round {n} of {of}',
    'cu.signupAt': 'Sign-up opens', 'cu.startsAt': 'Tournament starts', 'cu.announceAt': 'Shown in the list from', 'cu.startedAt': 'Started', 'cu.finishedAt': 'Finished',
    'cu.hint.signupAt': 'Players can sign up from this time.', 'cu.hint.startsAt': 'Sign-up closes and the bracket is built.', 'cu.hint.announceAt': 'Empty means it shows when sign-up opens.',
    'cu.publish': 'Publish now', 'cu.lockedRules': 'The tournament has started; only the name, description and icon can change.', 'cu.needName': 'Write at least one name.', 'cu.needTimes': 'Set the sign-up and start times.',
    'cu.back': '← Tournaments', 'cu.liveOn': 'Refreshes every 5 seconds', 'cu.refreshed': 'Updated {at}',
    'cu.publishNow': 'Publish', 'cu.published': 'Published', 'cu.startNow': 'Start now', 'cu.startConfirm': 'Start the tournament now with {n} players? Sign-up closes.', 'cu.started': 'Tournament started',
    'cu.addPlayer': 'Add player', 'cu.added': 'Player added', 'cu.cancel': 'Cancel and refund', 'cu.cancelTitle': 'Cancel tournament', 'cu.cancelled': 'Tournament cancelled and fees refunded',
    'cu.hint.cancel': 'The tournament is cancelled, open games are voided and every player gets the entry fee back. Prizes already paid are not taken back.', 'cu.cancelReason': 'Reason', 'cu.hint.reason': 'For example a server problem (optional)',
    'cu.deleteConfirm': 'Delete this tournament for good?',
    'cu.size': 'Bracket size', 'cu.sizeV': '{n} slots · {rounds} rounds', 'tp.template': 'Template',
    'cu.bracket': 'Bracket', 'cu.noBracket': 'The bracket is built when the tournament starts.', 'cu.noBracketCancelled': 'This tournament was cancelled before the bracket was built.',
    'cu.tbd': 'TBD', 'cu.ready': 'Ready', 'cu.matchOf': '{a} vs {b}', 'cu.match': 'Match', 'cu.sideA': 'First player', 'cu.sideB': 'Second player', 'cu.winsN': '{n} wins',
    'cu.winner': 'Winner', 'cu.reason': 'How it ended', 'cu.draws': 'Draws', 'cu.opensAt': 'Opened', 'cu.matchStarts': 'Starts', 'cu.closesAt': 'Deadline', 'cu.gameMatch': 'Game match ID',
    'cu.games': 'Games', 'cu.noGames': 'No games played yet.', 'cu.gameN': 'Game {n}', 'cu.draw': 'Draw', 'cu.wonBy': 'Won by {name}',
    'cu.force': 'Admin result', 'cu.hint.force': 'The open game is voided and the winner you pick moves on.', 'cu.winIs': 'Winner: {name}',
    'cu.forceConfirm': 'Make {name} the winner of this match?', 'cu.forceOk': 'Set result', 'cu.forced': 'Result saved',
    'cu.voidMatch': 'Void match', 'cu.voidConfirm': 'Void this match? Nobody moves on and the next slot stays empty.', 'cu.voided': 'Match voided',
    'cu.playersT': 'Players ({n})', 'cu.seed': 'Seed', 'cu.trophies': 'Trophies', 'cu.place': 'Place', 'cu.fee': 'Fee', 'cu.actions': 'Actions', 'cu.placeN': 'Place {n}', 'cu.reachedV': 'Round {n}',
    'cu.noPlayersYet': 'Nobody has signed up yet.', 'cu.noPlayers': 'No players.', 'cu.goneN': '{n} players withdrew or were removed.',
    'cu.removeConfirm': 'Remove {name} from the tournament? The fee is refunded in full.', 'cu.removed': 'Player removed',
    'cu.swap': 'Swap', 'cu.swapTitle': 'Swap {name}', 'cu.swapTo': 'New player', 'cu.swapped': 'Player swapped',
    'cu.hint.swap': 'The new player takes this player\'s slot. Only before the start or before their first game. The old player\'s fee is refunded.',
    'cu.disqualify': 'Disqualify', 'cu.dqTitle': 'Disqualify {name}', 'cu.dqRefund': 'Refund the fee', 'cu.disqualified': 'Player disqualified',
    'cu.hint.dq': 'The player leaves the tournament and the opponent wins. No prize is paid to them.',
    'cu.hint.who': 'Username or ID', 'cu.hint.whoLong': 'Exact username or full player ID', 'cu.hint.add': 'The player is added without a fee and without checking the requirements.',
    'cu.payouts': 'Prizes', 'cu.kind': 'For', 'cu.amounts': 'From the pool', 'cu.retry': 'Retry failed prizes', 'cu.retried': '{n} prizes queued again',
    'tp.new': 'New template', 'tp.edit': 'Edit template', 'tp.empty': 'No recurring templates yet.', 'tp.schedule': 'Schedule (CRON)', 'tp.nextStart': 'Next start', 'tp.openCups': 'Open tournaments', 'tp.manual': 'Manual',
    'tp.active': 'Active, creates tournaments by itself', 'tp.signupMinutes': 'Sign-up (minutes)', 'tp.announceMinutes': 'Show earlier (minutes)',
    'tp.hint': 'At each scheduled time a new tournament with these rules is created. Its sign-up opens some minutes before the start.',
    'tp.hint.schedule': 'Five parts: minute hour day month weekday, server time. Example: 0 18 * * 5 means every Friday at 18:00.', 'tp.hint.signup': 'How many minutes before the start sign-up opens.',
    'tp.hint.announce': 'How many minutes before sign-up it shows in the list.', 'tp.runNow': 'Create one now', 'tp.runConfirm': 'Create a tournament from this template now?',
    'tp.deleteConfirm': 'Delete template {key}? Tournaments already created from it stay.',
    'cs.s.general': 'General', 'cs.enabled': 'Tournaments are on', 'cs.hint.enabled': 'When off, players cannot sign up, but running tournaments still finish.',
    'cs.s.events': 'Notifications', 'cs.notify': 'Socket notifications to players', 'cs.code': 'Notification code', 'cs.hint.code': 'The code the app reads to know a tournament notification.',
    'cs.inApp': 'In-app messages for check-in, finish, prize and cancel', 'cs.publish': 'Game events for matches and finish (for missions and stats)',
    'cs.s.limits': 'Limits', 'cs.maxCapacity': 'Largest capacity', 'cs.hint.maxCapacity': 'Most players in one tournament.', 'cs.activePerPlayer': 'Open tournaments per player at once',
    'cs.tickSeconds': 'Check every (seconds)', 'cs.hint.tick': 'How often the server checks times and deadlines.', 'cs.listLimit': 'Items in the list', 'cs.historyLimit': 'Items in history',
    'cs.finishedDays': 'Show finished (days)', 'cs.hint.finishedDays': 'How many days a finished tournament stays in the player list.',
    'cs.s.defaults': 'Default rules', 'cs.hint.defaults': 'New tournaments and templates open with these rules. Existing tournaments do not change.',
    'ru.s.game': 'Game', 'ru.game': 'Game', 'ru.external': 'External (results by cup_report)', 'ru.hint.game': 'The game the matches are played in.',
    'ru.rated': 'Also count trophies and game rating', 'ru.gameOptions': 'Game options (JSON)', 'ru.hint.gameOptions': 'Passed to the game as is, for example {"goals":3}',
    'ru.s.players': 'Players and bracket', 'ru.capacity': 'Capacity', 'ru.hint.capacity': 'Any number; empty bracket slots become byes.',
    'ru.minPlayers': 'Fewest players', 'ru.hint.minPlayers': 'With fewer at the start the tournament is cancelled and fees are refunded.',
    'ru.underfill': 'If not full', 'ru.under.shrink': 'Play with the players we have', 'ru.under.cancel': 'Cancel', 'ru.hint.underfill': 'When more than the fewest but less than capacity signed up.',
    'ru.seeding': 'Seeding', 'ru.seed.random': 'Random', 'ru.seed.trophies': 'By trophies', 'ru.hint.seeding': 'By trophies, the best players meet only in the final and get the byes.',
    'ru.thirdPlace': 'Third-place match',
    'ru.s.bestOf': 'Games per match', 'ru.bestOf': 'Games per match', 'ru.bestOfN': 'Best of {n}', 'ru.hint.bestOf': 'Whoever first wins more than half moves on.',
    'ru.roundBestOf': 'For some rounds', 'ru.hint.roundBestOf': 'For example the final as best of 3.', 'ru.addRoundBestOf': 'Add round', 'ru.round': 'Round', 'ru.roundN': 'Round {n}',
    'ru.s.timing': 'Timing', 'ru.mode': 'Timing type',
    'ru.hint.mode.window': 'Each round stays open for a while and the two players play when both are ready.',
    'ru.hint.mode.live': 'Each round starts at a set time; players must check in before it.',
    'ru.windowMinutes': 'Round length (minutes)', 'ru.hint.windowMinutes': 'After that the round closes with the no-show rule.', 'ru.readySeconds': 'Ready lasts (seconds)', 'ru.hint.readySeconds': 'If the opponent is not ready in this time, ready must be pressed again.',
    'ru.checkInMinutes': 'Check-in (minutes before)', 'ru.hint.checkInMinutes': 'How many minutes before each round check-in opens.', 'ru.roundMinutes': 'Time between rounds (minutes)', 'ru.hint.roundMinutes': 'Each round starts this many minutes after the last.',
    'ru.joinSeconds': 'Time to join the game (seconds)', 'ru.hint.joinSeconds': 'A player who does not join in this time is absent.', 'ru.gameMinutes': 'Longest game (minutes)', 'ru.hint.gameMinutes': 'A game with no result by then is voided.',
    'ru.noShow': 'If neither showed up', 'ru.hint.noShow': 'When the round ends and neither player came.', 'ru.unplayed': 'If both came but did not finish', 'ru.hint.unplayed': 'When both came but the match did not finish by the end of the round.',
    'ru.ns.both_out': 'Both are out', 'ru.ns.higher_seed': 'Better seed wins', 'ru.ns.random': 'Coin toss',
    'ru.s.draw': 'Draws', 'ru.drawRule': 'If a game is a draw', 'ru.draw.replay': 'Play again', 'ru.draw.higher_seed': 'Better seed wins', 'ru.draw.random': 'Coin toss',
    'ru.maxReplays': 'Most replays', 'ru.fallback': 'After the replays', 'ru.gameDecides': 'The game breaks draws itself (like a golden goal)',
    'ru.s.entry': 'Entry', 'ru.feeCurrency': 'Fee currency', 'ru.noCurrency': 'Free', 'ru.feeAmount': 'Entry fee', 'ru.hint.fee': 'Taken at sign-up.',
    'ru.withdrawRefund': 'Refund on withdraw (percent)', 'ru.hint.withdrawRefund': 'If the player withdraws before the start.',
    'ru.minTrophies': 'Fewest trophies', 'ru.maxTrophies': 'Most trophies', 'ru.leagues': 'Only these leagues', 'ru.hint.leagues': 'None means every league.',
    'ru.tagIds': 'Required tags', 'ru.tagsMode': 'Of the required tags', 'ru.tagsMode.any': 'Any one is enough', 'ru.tagsMode.all': 'All are needed', 'ru.excludeTagIds': 'Blocked tags', 'ru.noneYet': 'Nothing created yet.',
    'ru.s.prizes': 'Prizes', 'ru.noPrize': 'No prize', 'ru.prize': 'Prize', 'ru.hint.semi': 'To both semi-final losers.',
    'ru.reached': 'Prize for reaching a round', 'ru.hint.reached': 'To everyone who reaches at least this round.', 'ru.reachedRound': 'From round', 'ru.addReached': 'Add round prize',
    'ru.pool': 'Prize pool', 'ru.hint.pool': 'A percent of the fees plus a base amount, in the fee currency, shared by the winners. Shares are percents of the pool.',
    'ru.poolPercent': 'Percent of fees', 'ru.poolBase': 'Base amount',
    'ru.badNumber': '«{field}» must be a number.', 'ru.badJson': 'Game options must be a JSON object.', 'ru.badReached': 'Pick a round and a prize for each round prize.',
    'ru.capN': '{n} players', 'ru.free': 'Free', 'ru.feeV': 'Fee {amount} {currency}',
    'code.CUP_ALREADY_OVER': "This cup is already over.",
    'code.CUP_ALREADY_SIGNED_UP': "You are already signed up for this cup.",
    'code.CUP_BAD_STATUS': "Unknown cup list filter.",
    'code.CUP_BUSY': "The cup is being updated. Try again in a moment.",
    'code.CUP_CAPACITY_BELOW_PLAYERS': "Capacity cannot be below the signed-up players.",
    'code.CUP_CHECK_IN_NOT_OPEN': "Check-in for this match is not open yet.",
    'code.CUP_DELETE_ONLY_DRAFT_OR_CANCELLED': "Only draft or cancelled cups can be deleted.",
    'code.CUP_DISABLED': "Cups are turned off.",
    'code.CUP_FULL': "This cup is full.",
    'code.CUP_GAME_ID_REQUIRED': "The game match id is required.",
    'code.CUP_GAME_NOT_FOUND': "This game does not belong to a cup.",
    'code.CUP_GAME_UNKNOWN': "This game cannot run cup matches.",
    'code.CUP_ID_REQUIRED': "The cup id is required.",
    'code.CUP_LEAGUE_REQUIRED': "Your league cannot enter this cup.",
    'code.CUP_MATCH_NOT_FOUND': "Cup match not found.",
    'code.CUP_MATCH_NOT_READY': "Both players of this match are not known yet.",
    'code.CUP_MATCH_OVER': "This match is already decided.",
    'code.CUP_NOT_A_PLAYER': "You are not playing in this cup.",
    'code.CUP_NOT_DRAFT': "This cup is already published.",
    'code.CUP_NOT_ENOUGH_CURRENCY': "You do not have enough {currencyKey} for the entry fee ({amount}).",
    'code.CUP_NOT_FOUND': "Cup not found.",
    'code.CUP_NOT_RUNNING': "This cup is not running.",
    'code.CUP_NOT_SIGNED_UP': "You are not signed up for this cup.",
    'code.CUP_NOT_STARTABLE': "This cup cannot be started now.",
    'code.CUP_NO_MATCH': "You have no match in this cup.",
    'code.CUP_REFUNDS_PENDING': "Some entry fees are not refunded yet.",
    'code.CUP_REMOVED': "You were removed from this cup.",
    'code.CUP_SCHEDULE_REQUIRED': "An active template needs a schedule.",
    'code.CUP_SERVER_ONLY': "Only the game server can report cup results.",
    'code.CUP_SIGNUP_AFTER_START': "Signup must open before the cup starts.",
    'code.CUP_SIGNUP_CLOSED': "Signup for this cup is closed.",
    'code.CUP_SIGNUP_NOT_OPEN': "Signup opens at {signupAt}.",
    'code.CUP_START_IN_PAST': "The start time is in the past.",
    'code.CUP_SWAP_TOO_LATE': "This player can no longer be swapped.",
    'code.CUP_TAG_EXCLUDED': "Players with your tag cannot enter this cup.",
    'code.CUP_TAG_REQUIRED': "This cup needs a special tag, such as VIP.",
    'code.CUP_TEMPLATE_NOT_FOUND': "Cup template not found.",
    'code.CUP_TIMES_REQUIRED': "Signup and start times are required.",
    'code.CUP_TOO_MANY_ACTIVE': "You are already in {max} cups.",
    'code.CUP_TROPHIES_TOO_HIGH': "You need at most {max} trophies for this cup.",
    'code.CUP_TROPHIES_TOO_LOW': "You need {min} trophies for this cup.",
    'code.CUP_USE_DISQUALIFY': "The cup has started; disqualify the player instead.",
    'code.CUP_WAITING_FOR_OPPONENT': "Your next opponent is not known yet.",
    'code.CUP_WINNER_NOT_IN_MATCH': "The winner must be one of the two players.",
    'code.CUP_WITHDRAW_CLOSED': "You can no longer withdraw from this cup.",
    'code.INTERNAL_ERROR': "Something went wrong. Please try again.",
    'code.INVALID_KEY': "The key or name is not valid.",
    'code.INVALID_POLICY': "This option is not allowed.",
    'code.INVALID_SCHEDULE': "The CRON schedule is not valid.",
    'code.KEY_TAKEN': "This key is already used.",
    'code.PLAYER_NOT_FOUND': "Player not found.",
    'code.VALUE_OUT_OF_RANGE': "The value is out of range."
  }
};

function cuLang() {
  var nc = window.NakamaConsole;
  var l = nc && nc.lang ? nc.lang() : (document.documentElement.getAttribute('lang') || 'fa');
  return /^en/i.test(l) ? 'en' : 'fa';
}

function has(key) {
  return (CU_TEXT[cuLang()] || CU_TEXT.fa)[key] != null;
}

function t(key, vars) {
  var dict = CU_TEXT[cuLang()] || CU_TEXT.fa;
  var s = dict[key] != null ? dict[key] : (CU_TEXT.en[key] != null ? CU_TEXT.en[key] : key);
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] != null ? digits(String(vars[k])) : m; });
  return s;
}

function digits(s) {
  if (cuLang() !== 'fa') return s;
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
    return d.toLocaleString(cuLang() === 'fa' ? 'fa-IR' : 'en-GB', { calendar: cuLang() === 'fa' ? 'persian' : 'gregory', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

function tr(key, vars, raw) {
  var s = t(key, vars);
  Object.keys(raw || {}).forEach(function (k) { s = s.split('{' + k + '}').join(String(raw[k])); });
  return s;
}

var CU_API = '/v2/console/base';
var CU_INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 px-3 py-2 text-sm';
var CU_BTN = 'cu-btn inline-flex items-center justify-center gap-1.5 font-medium whitespace-nowrap rounded-md border text-sm py-1.5 px-3';

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

function CuError(status, code, message, field, meta) {
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
  return fetch((path.indexOf('/v2/') === 0 ? '' : CU_API) + path, init).then(function (res) {
    if (res.status === 401 && !/^#\/login/.test(location.hash)) location.hash = '#/login';
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) throw new CuError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', data && data.field || '', data && data.meta);
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
  var b = h('button', { type: opts.type || 'button', class: CU_BTN + ' cu-btn--' + (opts.variant || 'ghost') + (opts.class ? ' ' + opts.class : ''), disabled: !!opts.disabled, title: opts.title, 'aria-label': opts.title, onclick: opts.onclick },
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

var cuToastBox = null;
function toast(kind, text) {
  if (!cuToastBox || !cuToastBox.isConnected) {
    cuToastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-cup': 'toasts' });
    document.body.appendChild(cuToastBox);
  }
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  cuToastBox.appendChild(el);
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
  var err = h('p', { class: 'cu-field-error eco-tf-err text-xs text-error-600 dark:text-error-400', hidden: true });
  var wrap = h('label', { class: 'cu-field eco-tbf block mb-4' + (extra ? ' ' + extra : '') },
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
  var el = h(opts.multiline ? 'textarea' : 'input', { class: CU_INPUT + (opts.mono ? ' font-mono' : ''), type: opts.multiline ? null : (opts.type || 'text'), dir: opts.ltr ? 'ltr' : 'auto', rows: opts.multiline ? (opts.rows || 3) : null, placeholder: opts.placeholder, name: opts.name, autocomplete: 'off', 'aria-label': opts.label });
  el.value = value == null ? '' : value;
  return el;
}

function select(value, options, opts) {
  opts = opts || {};
  var el = h('select', { class: CU_INPUT + ' eco-select', name: opts.name, 'aria-label': opts.label });
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) opt.selected = true;
    el.appendChild(opt);
  });
  return el;
}

function check(checked, label, name) {
  var box = h('input', { type: 'checkbox', class: 'cu-check', name: name, checked: !!checked });
  var wrap = h('label', { class: 'cu-checkline flex items-start gap-2 mb-3 cursor-pointer text-sm text-secondary-800 dark:text-secondary-200' }, box, h('span', { text: label }));
  wrap.input = box;
  return wrap;
}

function numberValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? Math.round(n) : NaN;
}

var cuPanel = null;
function closePanel() {
  if (cuPanel) { cuPanel.remove(); cuPanel = null; document.removeEventListener('keydown', panelKey); }
}

function panelKey(e) { if (e.key === 'Escape') closePanel(); }

function sidePanel(title, body, footer, name) {
  closePanel();
  var holder = h('div', { class: 'cu-panel-holder' });
  var panel = h('div', { class: 'cu-panel bg-white dark:bg-secondary-900 shadow-lg', role: 'dialog', 'aria-modal': 'true', 'aria-label': title, 'data-cup-panel': name || '' },
    h('div', { class: 'cu-panel-head flex items-center justify-between border-b px-6 py-3' },
      h('h2', { class: 'text-lg font-semibold truncate', text: title }),
      button({ label: '×', title: t('close'), onclick: closePanel, class: 'cu-close' })),
    h('div', { class: 'cu-panel-body p-6' }, body),
    footer ? h('div', { class: 'cu-panel-foot flex gap-2 justify-end bg-secondary-50 dark:bg-secondary-800 p-4' }, footer) : null);
  holder.appendChild(panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) closePanel(); });
  cuPanel = h('div', { 'data-cup': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, h('div', { class: 'cu-shade' }), holder);
  document.body.appendChild(cuPanel);
  document.addEventListener('keydown', panelKey);
  return panel;
}

function confirmBox(text, okLabel) {
  return new Promise(function (resolve) {
    var root;
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-cup': 'confirm', 'data-l10n-skip': '', class: 'cu-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'cu-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('confirm.title') }),
        h('p', { class: 'text-sm mb-4', text: text }),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: okLabel || t('yes'), variant: 'danger', data: { 'cup-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root);
  });
}

function pageFrame(page, desc, actions) {
  var content = h('div', { class: 'flex flex-col gap-4' });
  var el = h('div', { class: 'cu-page flex flex-col flex-1 px-4 md:px-6 pb-6 max-w-[1600px] w-full mx-auto', 'data-cup-page': page },
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
    var tr = h('tr', { class: 'border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' + (onrow ? ' cursor-pointer' : ''), 'data-cup-row': row.id || '' });
    if (onrow) {
      tr.tabIndex = 0;
      tr.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,select,label')) onrow(row.item); });
      tr.addEventListener('keydown', function (e) { if (e.key === 'Enter') onrow(row.item); });
    }
    row.cells.forEach(function (cell, i) { tr.appendChild(h('td', { class: 'px-3 py-2 align-middle text-sm', 'data-label': columns[i].label }, cell)); });
    tbody.appendChild(tr);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow rounded-md overflow-x-auto' }, h('table', { class: 'w-full eco-table cu-table' }, thead, tbody));
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
  nav.appendChild(h('span', { class: 'eco-pager-range tabular-nums', 'data-cup-range': '', text: t('range', { from: from, to: to, total: total }) }));
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

var cuLanguages = null;
function loadLanguages() {
  if (cuLanguages) return Promise.resolve(cuLanguages);
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch('/v2/console/base/languages', { headers: headers, credentials: 'same-origin' }).then(function (r) { return r.ok ? r.json() : null; }).then(function (d) {
    var list = d && d.items ? d.items.filter(function (l) { return l.isActive; }).map(function (l) { return { code: l.code, name: l.nativeName || l.code, isDefault: !!l.isDefault, direction: l.direction }; }) : [];
    if (!list.length) list = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    list.sort(function (a, b) { return (b.isDefault ? 1 : 0) - (a.isDefault ? 1 : 0); });
    cuLanguages = list;
    return list;
  }, function () {
    cuLanguages = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    return cuLanguages;
  });
}

function translated(label, name, value, opts) {
  opts = opts || {};
  var inputs = {};
  var box = h('div', { class: 'grid sm:grid-cols-2 gap-x-4', 'data-cup-text': name });
  (cuLanguages || []).forEach(function (l) {
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
  var l = cuLang();
  return m[l] || m.fa || m.en || Object.keys(m).map(function (k) { return m[k]; }).filter(Boolean)[0] || fallback || '';
}

function fieldError(msgBox, err) {
  clear(msgBox).appendChild(notice('error', cuErrText(err)));
}

function options(list, value) {
  return list.map(function (o) { return typeof o === 'string' ? { value: o, label: o } : o; });
}

function dateInput(value, name, type) {
  var el = h('input', { type: type || 'date', class: CU_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  el.value = value || '';
  return el;
}

function isoDay(d) {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

function cuErrText(err) {
  if (err && err.code && has('code.' + err.code)) return t('code.' + err.code, err.meta || null) + (err.field ? ' · ' + err.field : '');
  return errorText(err) + (err && err.field ? ' · ' + err.field : '');
}

function section(title, children, name) {
  return card(h('div', { class: 'p-4', 'data-cup-section': name || '' },
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
    var msg = h('div', { 'data-cup-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', 'data-cup-updated': '' });
    var paintMeta = function (x) { meta.textContent = x.updatedBy ? t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }) : ''; };
    paintMeta(d);
    var save = button({ label: t('save'), variant: 'primary', data: { 'cup-save': page }, onclick: function () {
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
        msg.appendChild(notice('error', cuErrText(err)));
      });
    } });
    clear(frame.content);
    form.el.forEach(function (el) { frame.content.appendChild(el); });
    frame.content.appendChild(h('div', { class: 'cu-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), save));
  }, function (err) {
    clear(frame.content).appendChild(notice('error', cuErrText(err)));
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
  return h('a', { href: '#/players/' + id, class: 'eco-link', 'data-cup-player-link': id, dir: 'auto', text: name || id.slice(0, 8) });
}

var cuRefs = {};
function loadRefs(kind, force) {
  if (cuRefs[kind] && !force) return Promise.resolve(cuRefs[kind]);
  return api('GET', '/' + kind + '/refs' + qs({ lang: cuLang() })).then(function (r) { cuRefs[kind] = r || {}; return cuRefs[kind]; }, function () { cuRefs[kind] = {}; return cuRefs[kind]; });
}

function refList(kind, list) {
  return ((cuRefs[kind] && cuRefs[kind][list]) || []).map(function (x) { return typeof x === 'string' ? { key: x, name: x } : x; });
}

function refOptions(items, value, emptyLabel) {
  var opts = emptyLabel != null ? [{ value: '', label: emptyLabel }] : [];
  items.forEach(function (r) { opts.push({ value: r.key, label: (r.name && r.name !== r.key ? r.name + ' · ' : '') + r.key }); });
  if (value && !items.some(function (r) { return r.key === value; })) opts.push({ value: value, label: value + ' (' + t('missing') + ')' });
  return opts;
}

function refSelect(items, value, name, emptyLabel) {
  var el = select(value || '', refOptions(items, value, emptyLabel), { name: name, label: name });
  el.setAttribute('dir', 'auto');
  return el;
}

var cuListSeq = 0;
function eventInput(items, value, name, placeholder) {
  var el = input(value || '', { name: name, ltr: true, mono: true, placeholder: placeholder });
  var id = 'cu-list-' + (++cuListSeq);
  var list = h('datalist', { id: id });
  items.forEach(function (e) { list.appendChild(h('option', { value: e.key, text: e.name || e.key })); });
  el.setAttribute('list', id);
  var wrap = h('div', {}, el, list);
  wrap.input = el;
  return wrap;
}

function grid() {
  return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments));
}

function grid3() {
  return h('div', { class: 'grid sm:grid-cols-3 gap-x-4' }, Array.prototype.slice.call(arguments));
}

function opt(prefix, list) {
  return list.map(function (v) { return { value: v, label: t(prefix + v) }; });
}

function tagsInput(list, name, placeholder) {
  var el = input((list || []).join('، '), { name: name, placeholder: placeholder || t('hint.tags'), ltr: true });
  el.read = function () { return el.value.split(/[,،\n]/).map(function (s) { return s.trim(); }).filter(Boolean); };
  return el;
}

function jsonInput(value, name) {
  var el = input(JSON.stringify(value || {}), { name: name, ltr: true, mono: true, multiline: true, rows: 2 });
  el.read = function () { try { var v = JSON.parse(el.value.trim() || '{}'); return v && typeof v === 'object' && !Array.isArray(v) ? v : undefined; } catch (e) { return undefined; } };
  return el;
}

function intOr(el, d) {
  var n = numberValue(el);
  return n == null || isNaN(n) ? d : n;
}

function colorDot(color) {
  var dot = h('span', { class: 'cu-dot' });
  if (color) dot.style.background = color;
  return dot;
}

function nameCell(title, key, icon, color) {
  return h('div', { class: 'flex items-center gap-2 min-w-0' }, icon ? h('img', { src: icon, alt: '', class: 'cu-icon' }) : colorDot(color),
    h('div', { class: 'flex flex-col min-w-0' }, h('span', { class: 'font-medium', dir: 'auto', text: title }), key ? mono(key) : null));
}

function dl(pairs, name) {
  var box = h('dl', { class: 'cu-dl text-sm', 'data-cup-dl': name || '' });
  pairs.forEach(function (p) {
    if (!p) return;
    box.appendChild(h('dt', { text: p[0] }));
    box.appendChild(h('dd', {}, p[1] == null || p[1] === '' ? t('none') : p[1]));
  });
  return box;
}

function deltaText(n) {
  var s = (n > 0 ? '+' : n < 0 ? '−' : '') + fmtNum(Math.abs(n));
  return h('span', { class: n > 0 ? 'cu-amount-in' : n < 0 ? 'cu-amount-out' : '', dir: 'ltr', text: s });
}

function searchBox(onchange) {
  var el = input('', { name: 'q', placeholder: t('search') });
  var timer = null;
  el.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { onchange(el.value.trim()); }, 300); });
  return el;
}

function loading() {
  return h('p', { class: 'text-sm text-secondary-500', text: t('loading') });
}

function emptyCard(text) {
  return card(h('p', { class: 'p-4 text-sm text-secondary-500', text: text }));
}

function colorField(value) {
  var color = input(value || '', { name: 'color', ltr: true, mono: true, placeholder: '#F5B301' });
  var picker = h('input', { type: 'color', class: 'cu-color', name: 'color-picker', 'aria-label': t('f.color'), value: /^#[0-9A-Fa-f]{6}$/.test(value || '') ? value : '#f5b301' });
  picker.addEventListener('input', function () { color.value = picker.value.toUpperCase(); });
  color.addEventListener('input', function () { if (/^#[0-9A-Fa-f]{6}$/.test(color.value.trim())) picker.value = color.value.trim(); });
  var wrap = h('div', { class: 'flex gap-2 items-center' }, picker, color);
  wrap.input = color;
  return wrap;
}

function isUUID(s) {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(s || '');
}

function localInput(v, name) {
  if (window.NakamaDate) return window.NakamaDate.input({ mode: 'datetime', value: v || null, name: name, label: name, inputClass: CU_INPUT });
  var el = h('input', { type: 'datetime-local', class: CU_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  if (v) {
    var d = new Date(v);
    if (!isNaN(d)) el.value = new Date(d.getTime() - d.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
  }
  el.read = function () { if (!el.value) return null; var d = new Date(el.value); return isNaN(d) ? null : d.toISOString(); };
  return el;
}

function group(label, control, hint) {
  return h('div', { class: 'cu-field block mb-4', role: 'group', 'aria-label': label },
    h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1', text: label }), control,
    hint ? h('span', { class: 'block text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: hint }) : null);
}

function checkList(items, chosen, name) {
  var box = h('div', { class: 'cu-scope-list', 'data-cup-checks': name });
  var all = items.slice();
  (chosen || []).forEach(function (k) { if (!all.some(function (x) { return x.key === k; })) all.push({ key: k, name: k + ' (' + t('missing') + ')' }); });
  var boxes = all.map(function (x) {
    var c = check((chosen || []).indexOf(x.key) !== -1, x.name && x.name !== x.key ? x.name : x.key, name);
    c.input.value = x.key;
    box.appendChild(c);
    return c;
  });
  if (!all.length) box.appendChild(h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('ru.noneYet') }));
  box.read = function () { return boxes.filter(function (c) { return c.input.checked; }).map(function (c) { return c.input.value; }); };
  return box;
}

function bestOfSelect(v, name) {
  return select(String(v || 1), [1, 3, 5, 7, 9].map(function (n) { return { value: String(n), label: t('ru.bestOfN', { n: n }) }; }), { name: name, label: t('ru.bestOf') });
}

function prizeSelect(v, name) {
  return refSelect(refList('cups', 'prizes'), v, name, t('ru.noPrize'));
}

function rowList(items, make, addLabel, name) {
  var box = h('div', { class: 'cu-rows', 'data-cup-rows': name });
  var rows = [];
  function push(it) {
    var r = make(it || null);
    var del = button({ label: t('remove'), variant: 'danger', data: { 'cup-remove': name }, onclick: function () { r.el.remove(); rows.splice(rows.indexOf(r), 1); } });
    r.el.appendChild(h('div', { class: 'flex items-end' }, del));
    rows.push(r);
    box.appendChild(r.el);
  }
  (items || []).forEach(push);
  var wrap = h('div', { class: 'mb-4' }, box, h('div', { class: 'mt-2' }, button({ label: addLabel, data: { 'cup-add': name }, onclick: function () { push(null); } })));
  wrap.read = function () { return rows.map(function (r) { return r.read(); }); };
  return wrap;
}

function roundOptions() {
  var out = [{ value: 'final', label: t('rk.final') }, { value: 'semi', label: t('rk.semi') }, { value: 'third', label: t('rk.third') }];
  for (var i = 1; i <= 7; i++) out.push({ value: String(i), label: t('ru.roundN', { n: i }) });
  return out;
}

function rulesForm(src, opts) {
  opts = opts || {};
  var r = JSON.parse(JSON.stringify(src || {}));
  r.timing = r.timing || {};
  r.draw = r.draw || {};
  r.entry = r.entry || {};
  r.entry.fee = r.entry.fee || {};
  r.entry.requirements = r.entry.requirements || {};
  r.prizes = r.prizes || {};
  r.prizes.pool = r.prizes.pool || {};
  r.prizes.pool.shares = r.prizes.pool.shares || {};
  var tm = r.timing, dr = r.draw, en = r.entry, rq = en.requirements, pz = r.prizes, pool = pz.pool, sh = pool.shares;
  var nums = [];
  function num(value, name, label) {
    var el = numInput(value, name, label);
    nums.push({ el: el, label: label });
    return el;
  }
  var gameOpts = refList('cups', 'games').map(function (g) { return { value: g.key, label: g.key ? (g.name && g.name !== g.key ? g.name + ' · ' + g.key : g.key) : t('ru.external') }; });
  if (r.game && !gameOpts.some(function (o) { return o.value === r.game; })) gameOpts.push({ value: r.game, label: r.game + ' (' + t('missing') + ')' });
  if (!gameOpts.length) gameOpts.push({ value: '', label: t('ru.external') });
  var game = select(r.game || '', gameOpts, { name: 'rules.game', label: t('ru.game') });
  var rated = check(r.rated, t('ru.rated'), 'rules.rated');
  var gameOptions = jsonInput(r.gameOptions, 'rules.gameOptions');
  var capacity = num(r.capacity, 'rules.capacity', t('ru.capacity'));
  var minPlayers = num(r.minPlayers, 'rules.minPlayers', t('ru.minPlayers'));
  var underfill = select(r.underfill || 'shrink', opt('ru.under.', ['shrink', 'cancel']), { name: 'rules.underfill', label: t('ru.underfill') });
  var seeding = select(r.seeding || 'random', opt('ru.seed.', ['random', 'trophies']), { name: 'rules.seeding', label: t('ru.seeding') });
  var third = check(r.thirdPlace, t('ru.thirdPlace'), 'rules.thirdPlace');
  var bestOf = bestOfSelect(r.bestOf, 'rules.bestOf');
  var roundBestOf = rowList(r.roundBestOf, function (it) {
    var round = select(it ? it.round : 'final', roundOptions(), { name: 'roundBestOf.round', label: t('ru.round') });
    var bo = bestOfSelect(it ? it.bestOf : 3, 'roundBestOf.bestOf');
    return { el: h('div', { class: 'cu-row' }, field(t('ru.round'), round), field(t('ru.bestOf'), bo)), read: function () { return { round: round.value, bestOf: Number(bo.value) }; } };
  }, t('ru.addRoundBestOf'), 'roundBestOf');
  var mode = select(tm.mode || 'window', opt('mode.', ['window', 'live']), { name: 'rules.timing.mode', label: t('ru.mode') });
  var windowMinutes = num(tm.windowMinutes, 'rules.timing.windowMinutes', t('ru.windowMinutes'));
  var readySeconds = num(tm.readySeconds, 'rules.timing.readySeconds', t('ru.readySeconds'));
  var checkInMinutes = num(tm.checkInMinutes, 'rules.timing.checkInMinutes', t('ru.checkInMinutes'));
  var roundMinutes = num(tm.roundMinutes, 'rules.timing.roundMinutes', t('ru.roundMinutes'));
  var joinSeconds = num(tm.joinSeconds, 'rules.timing.joinSeconds', t('ru.joinSeconds'));
  var gameMinutes = num(tm.gameMinutes, 'rules.timing.gameMinutes', t('ru.gameMinutes'));
  var noShow = select(tm.noShow || 'both_out', opt('ru.ns.', ['both_out', 'higher_seed', 'random']), { name: 'rules.timing.noShow', label: t('ru.noShow') });
  var unplayed = select(tm.unplayed || 'higher_seed', opt('ru.ns.', ['both_out', 'higher_seed', 'random']), { name: 'rules.timing.unplayed', label: t('ru.unplayed') });
  var windowBox = grid(field(t('ru.windowMinutes'), windowMinutes, t('ru.hint.windowMinutes')), field(t('ru.readySeconds'), readySeconds, t('ru.hint.readySeconds')));
  var liveBox = grid(field(t('ru.checkInMinutes'), checkInMinutes, t('ru.hint.checkInMinutes')), field(t('ru.roundMinutes'), roundMinutes, t('ru.hint.roundMinutes')));
  var modeHint = h('p', { class: 'text-xs text-secondary-500 mb-3', 'data-cup-mode-hint': '' });
  function paintMode() {
    windowBox.hidden = mode.value !== 'window';
    liveBox.hidden = mode.value !== 'live';
    modeHint.textContent = t('ru.hint.mode.' + mode.value);
  }
  mode.addEventListener('change', paintMode);
  paintMode();
  var drawRule = select(dr.rule || 'replay', opt('ru.draw.', ['replay', 'higher_seed', 'random']), { name: 'rules.draw.rule', label: t('ru.drawRule') });
  var maxReplays = num(dr.maxReplays, 'rules.draw.maxReplays', t('ru.maxReplays'));
  var fallback = select(dr.fallback || 'higher_seed', opt('ru.draw.', ['higher_seed', 'random']), { name: 'rules.draw.fallback', label: t('ru.fallback') });
  var gameDecides = check(dr.gameDecides, t('ru.gameDecides'), 'rules.draw.gameDecides');
  var replayBox = grid(field(t('ru.maxReplays'), maxReplays), field(t('ru.fallback'), fallback));
  function paintDraw() { replayBox.hidden = drawRule.value !== 'replay'; }
  drawRule.addEventListener('change', paintDraw);
  paintDraw();
  var currencies = refList('cups', 'currencies');
  var feeCur = refSelect(currencies, en.fee.currencyKey || '', 'rules.entry.fee.currencyKey', t('ru.noCurrency'));
  var feeAmount = num(en.fee.amount || 0, 'rules.entry.fee.amount', t('ru.feeAmount'));
  var refund = num(en.withdrawRefundPercent, 'rules.entry.withdrawRefundPercent', t('ru.withdrawRefund'));
  var minTrophies = num(rq.minTrophies || 0, 'rules.entry.requirements.minTrophies', t('ru.minTrophies'));
  var maxTrophies = num(rq.maxTrophies || 0, 'rules.entry.requirements.maxTrophies', t('ru.maxTrophies'));
  var leagues = checkList(refList('cups', 'leagues'), rq.leagues, 'rules.entry.requirements.leagues');
  var tags = refList('cups', 'tags');
  var tagIds = checkList(tags, rq.tagIds, 'rules.entry.requirements.tagIds');
  var tagsMode = select(rq.tagsMode || 'any', opt('ru.tagsMode.', ['any', 'all']), { name: 'rules.entry.requirements.tagsMode', label: t('ru.tagsMode') });
  var excludeTagIds = checkList(tags, rq.excludeTagIds, 'rules.entry.requirements.excludeTagIds');
  var pChampion = prizeSelect(pz.champion, 'rules.prizes.champion');
  var pRunner = prizeSelect(pz.runnerUp, 'rules.prizes.runnerUp');
  var pSemi = prizeSelect(pz.semi, 'rules.prizes.semi');
  var pThird = prizeSelect(pz.third, 'rules.prizes.third');
  var pFourth = prizeSelect(pz.fourth, 'rules.prizes.fourth');
  var reached = rowList(pz.reached, function (it) {
    var round = numInput(it ? it.round : 1, 'reached.round', t('ru.round'));
    var prize = prizeSelect(it ? it.prizeKey : '', 'reached.prizeKey');
    return { el: h('div', { class: 'cu-row' }, field(t('ru.reachedRound'), round), field(t('ru.prize'), prize)), read: function () { return { round: intOr(round, 0), prizeKey: prize.value }; } };
  }, t('ru.addReached'), 'reached');
  var poolPercent = num(pool.percent || 0, 'rules.prizes.pool.percent', t('ru.poolPercent'));
  var poolBase = num(pool.base || 0, 'rules.prizes.pool.base', t('ru.poolBase'));
  var shChampion = num(sh.champion || 0, 'rules.prizes.pool.shares.champion', t('pl.champion'));
  var shRunner = num(sh.runnerUp || 0, 'rules.prizes.pool.shares.runnerUp', t('pl.runnerUp'));
  var shSemi = num(sh.semi || 0, 'rules.prizes.pool.shares.semi', t('pl.semi'));
  var shThird = num(sh.third || 0, 'rules.prizes.pool.shares.third', t('pl.third'));
  var shFourth = num(sh.fourth || 0, 'rules.prizes.pool.shares.fourth', t('pl.fourth'));
  var thirdPrizes = grid(field(t('pl.third'), pThird), field(t('pl.fourth'), pFourth));
  var semiPrize = field(t('pl.semi'), pSemi, t('ru.hint.semi'));
  function paintThird() { thirdPrizes.hidden = !third.input.checked; semiPrize.hidden = third.input.checked; }
  third.input.addEventListener('change', paintThird);
  paintThird();
  var locked = !!opts.locked;
  var el = [
    section(t('ru.s.game'), [
      field(t('ru.game'), game, t('ru.hint.game')), rated,
      field(t('ru.gameOptions'), gameOptions, t('ru.hint.gameOptions'))
    ], 'game'),
    section(t('ru.s.players'), [
      grid(field(t('ru.capacity'), capacity, t('ru.hint.capacity')), field(t('ru.minPlayers'), minPlayers, t('ru.hint.minPlayers'))),
      grid(field(t('ru.underfill'), underfill, t('ru.hint.underfill')), field(t('ru.seeding'), seeding, t('ru.hint.seeding'))),
      third
    ], 'players'),
    section(t('ru.s.bestOf'), [
      field(t('ru.bestOf'), bestOf, t('ru.hint.bestOf')),
      group(t('ru.roundBestOf'), roundBestOf, t('ru.hint.roundBestOf'))
    ], 'bestOf'),
    section(t('ru.s.timing'), [
      field(t('ru.mode'), mode), modeHint, windowBox, liveBox,
      grid(field(t('ru.joinSeconds'), joinSeconds, t('ru.hint.joinSeconds')), field(t('ru.gameMinutes'), gameMinutes, t('ru.hint.gameMinutes'))),
      grid(field(t('ru.noShow'), noShow, t('ru.hint.noShow')), field(t('ru.unplayed'), unplayed, t('ru.hint.unplayed')))
    ], 'timing'),
    section(t('ru.s.draw'), [field(t('ru.drawRule'), drawRule), replayBox, gameDecides], 'draw'),
    section(t('ru.s.entry'), [
      grid(field(t('ru.feeCurrency'), feeCur, t('ru.hint.fee')), field(t('ru.feeAmount'), feeAmount)),
      field(t('ru.withdrawRefund'), refund, t('ru.hint.withdrawRefund')),
      grid(field(t('ru.minTrophies'), minTrophies, t('hint.zero')), field(t('ru.maxTrophies'), maxTrophies, t('hint.zero'))),
      group(t('ru.leagues'), leagues, t('ru.hint.leagues')),
      group(t('ru.tagIds'), tagIds), field(t('ru.tagsMode'), tagsMode),
      group(t('ru.excludeTagIds'), excludeTagIds)
    ], 'entry'),
    section(t('ru.s.prizes'), [
      grid(field(t('pl.champion'), pChampion), field(t('pl.runnerUp'), pRunner)),
      semiPrize, thirdPrizes,
      group(t('ru.reached'), reached, t('ru.hint.reached')),
      h('h3', { class: 'font-medium text-sm mt-2 mb-2', text: t('ru.pool') }),
      h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('ru.hint.pool') }),
      grid(field(t('ru.poolPercent'), poolPercent), field(t('ru.poolBase'), poolBase)),
      grid3(field(t('pl.champion'), shChampion), field(t('pl.runnerUp'), shRunner), field(t('pl.semi'), shSemi)),
      grid(field(t('pl.third'), shThird), field(t('pl.fourth'), shFourth))
    ], 'prizes')
  ];
  if (locked) el.forEach(function (s) { Array.prototype.forEach.call(s.querySelectorAll('input,select,textarea,button'), function (x) { x.disabled = true; }); });
  function read() {
    for (var i = 0; i < nums.length; i++) {
      var n = numberValue(nums[i].el);
      if (n === null || isNaN(n)) return { error: t('ru.badNumber', { field: nums[i].label }) };
    }
    var go = gameOptions.read();
    if (go === undefined) return { error: t('ru.badJson') };
    var rb = roundBestOf.read();
    var rc = reached.read();
    for (var j = 0; j < rc.length; j++) if (!rc[j].round || !rc[j].prizeKey) return { error: t('ru.badReached') };
    var out = JSON.parse(JSON.stringify(src || {}));
    out.game = game.value;
    out.rated = rated.input.checked;
    out.gameOptions = go;
    out.capacity = numberValue(capacity);
    out.minPlayers = numberValue(minPlayers);
    out.underfill = underfill.value;
    out.seeding = seeding.value;
    out.thirdPlace = third.input.checked;
    out.bestOf = Number(bestOf.value);
    out.roundBestOf = rb;
    out.timing = { mode: mode.value, windowMinutes: numberValue(windowMinutes), roundMinutes: numberValue(roundMinutes), checkInMinutes: numberValue(checkInMinutes),
      noShow: noShow.value, unplayed: unplayed.value, readySeconds: numberValue(readySeconds), joinSeconds: numberValue(joinSeconds), gameMinutes: numberValue(gameMinutes) };
    out.draw = { rule: drawRule.value, maxReplays: numberValue(maxReplays), fallback: fallback.value, gameDecides: gameDecides.input.checked };
    out.entry = { fee: { currencyKey: feeCur.value, amount: feeCur.value ? numberValue(feeAmount) : 0 }, withdrawRefundPercent: numberValue(refund),
      requirements: { minTrophies: numberValue(minTrophies), maxTrophies: numberValue(maxTrophies), leagues: leagues.read(), tagIds: tagIds.read(), tagsMode: tagsMode.value, excludeTagIds: excludeTagIds.read() } };
    out.prizes = { champion: pChampion.value, runnerUp: pRunner.value, semi: third.input.checked ? '' : pSemi.value, third: third.input.checked ? pThird.value : '', fourth: third.input.checked ? pFourth.value : '',
      reached: rc, pool: { percent: numberValue(poolPercent), base: numberValue(poolBase), shares: { champion: numberValue(shChampion), runnerUp: numberValue(shRunner), semi: numberValue(shSemi), third: numberValue(shThird), fourth: numberValue(shFourth) } } };
    return { rules: out };
  }
  return { el: el, read: read };
}

function rulesSummary(r) {
  if (!r) return '';
  var parts = [t('mode.' + (r.timing && r.timing.mode || 'window')), t('ru.bestOfN', { n: r.bestOf || 1 }), t('ru.capN', { n: r.capacity })];
  if (r.entry && r.entry.fee && r.entry.fee.amount > 0) parts.push(tr('ru.feeV', null, { amount: fmtNum(r.entry.fee.amount), currency: r.entry.fee.currencyKey }));
  else parts.push(t('ru.free'));
  return parts.join(' · ');
}

var CU_STATUS_COLOR = { draft: 'secondary', upcoming: 'primary', signup: 'primary', running: 'warning', finished: 'success', cancelled: 'secondary' };
var CU_STATUSES = ['draft', 'upcoming', 'signup', 'running', 'finished', 'cancelled'];

function statusTag(s) {
  return tag(t('st.' + s), CU_STATUS_COLOR[s] || 'secondary');
}

function roundName(key) {
  if (!key) return '';
  var m = /^r(\d+)$/.exec(key);
  return m ? t('rk.of', { n: Number(m[1]) }) : t('rk.' + key);
}

function cupName(c) {
  return textOf(c && c.name, c && (c.templateKey || (c.id || '').slice(0, 8)));
}

function viewHref(id) {
  return '#/cups/list/' + encodeURIComponent(id);
}

function cupsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 25, total: 0, status: CU_STATUSES.indexOf(q.status) !== -1 ? q.status : '', q: '', template: q.template || '' };
  var frame = pageFrame('cups/list', t('desc.cups/list'), [
    button({ label: t('cu.new'), variant: 'primary', data: { 'cup-action': 'new-cup' }, onclick: function () { cupEditor(null, function (c) { location.hash = viewHref(c.id); }); } })]);
  clear(root).appendChild(frame.el);
  var status = select(state.status, [{ value: '', label: t('all') }].concat(opt('st.', CU_STATUSES)), { name: 'status', label: t('f.state') });
  var search = searchBox(function (v) { state.q = v; state.page = 1; reload(); });
  status.addEventListener('change', function () { state.status = status.value; state.page = 1; reload(); });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx cu-filters' }, field(t('f.state'), status), field(t('search'), search, t('cu.hint.search'))));
  if (state.template) frame.content.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 text-sm' }, h('span', { text: t('cu.byTemplate') }), mono(state.template),
    h('a', { href: '#/cups/list', class: 'eco-link', 'data-cup-clear-template': '', text: t('cu.allCups') })));
  frame.content.appendChild(listBox);
  function reload() {
    clear(listBox).appendChild(loading());
    Promise.all([api('GET', '/cups/list' + qs({ status: state.status, q: state.q, template: state.template, page: state.page, limit: state.limit })), loadLanguages()]).then(function (res) {
      var d = res[0];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      if (!items.length) { listBox.appendChild(emptyCard(state.q || state.status ? t('empty') : t('cu.empty'))); return; }
      listBox.appendChild(h('div', { 'data-cup-list': 'cups' }, table([
        { label: t('f.name') }, { label: t('f.state') }, { label: t('ru.mode') }, { label: t('cu.players') }, { label: t('cu.starts') }, { label: t('cu.round') }
      ], items.map(function (c) {
        var r = c.rules || {};
        return { id: c.id, item: c, cells: [
          h('a', { href: viewHref(c.id), class: 'flex flex-col min-w-0', 'data-cup-open': c.id },
            h('span', { class: 'font-medium eco-link', dir: 'auto', text: cupName(c) }),
            h('span', { class: 'text-xs text-secondary-500', text: rulesSummary(r) })),
          statusTag(c.status),
          h('span', { text: t('mode.' + ((r.timing && r.timing.mode) || 'window')) }),
          h('span', { class: 'tabular-nums', 'data-cup-entrants': '', text: t('cu.entrantsV', { n: fmtNum(c.status === 'running' || c.status === 'finished' ? c.players : c.entrants || 0), cap: fmtNum(r.capacity || 0) }) }),
          h('span', { class: 'text-xs', text: fmtDate(c.startsAt) }),
          h('span', { class: 'text-xs', text: c.status === 'running' && c.currentRound ? t('cu.roundV', { n: c.currentRound, of: c.rounds }) : c.status === 'finished' ? t('st.finished') : t('none') })
        ] };
      }), function (c) { location.hash = viewHref(c.id); })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', cuErrText(err))); });
  }
  reload();
}

function soon(minutes) {
  var d = new Date(Date.now() + minutes * 60000);
  d.setSeconds(0, 0);
  return d.toISOString();
}

function cupEditor(c, done) {
  var editing = !!(c && c.id);
  var open = !editing || c.status === 'draft' || c.status === 'upcoming' || c.status === 'signup';
  var body = h('div', { 'data-cup-form': 'cup' }, loading());
  var save = button({ label: editing ? t('save') : t('cu.create'), variant: 'primary', data: { 'cup-action': 'save-cup' }, disabled: true });
  sidePanel(editing ? t('cu.edit') : t('cu.new'), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'cup');
  Promise.all([loadRefs('cups', true), loadLanguages()]).then(function (res) {
    var refs = res[0];
    clear(body);
    var name = translated(t('f.name'), 'name', c && c.name);
    var desc = translated(t('f.description'), 'description', c && c.description, { multiline: true });
    var icon = input(c && c.iconUrl || '', { name: 'iconUrl', ltr: true, placeholder: 'https://' });
    var signupAt = localInput(c ? c.signupAt : soon(0), 'signupAt');
    var startsAt = localInput(c ? c.startsAt : soon(60), 'startsAt');
    var announceAt = localInput(c ? c.announceAt : null, 'announceAt');
    var publish = editing ? null : check(false, t('cu.publish'), 'publish');
    var rules = rulesForm(c ? c.rules : refs.defaults, { locked: !open });
    var msg = h('div', { 'data-cup-msg': '' });
    body.appendChild(name);
    body.appendChild(desc);
    body.appendChild(field(t('f.icon'), icon));
    var times = grid(field(t('cu.signupAt'), signupAt, t('cu.hint.signupAt')), field(t('cu.startsAt'), startsAt, t('cu.hint.startsAt')));
    body.appendChild(times);
    body.appendChild(field(t('cu.announceAt'), announceAt, t('cu.hint.announceAt')));
    if (!open) {
      [signupAt, startsAt, announceAt].forEach(function (x) { x.disabled = true; });
      body.appendChild(notice('warning', t('cu.lockedRules')));
      body.appendChild(h('div', { class: 'h-3' }));
    }
    if (publish) body.appendChild(publish);
    rules.el.forEach(function (s) { body.appendChild(s); body.appendChild(h('div', { class: 'h-3' })); });
    body.appendChild(msg);
    save.disabled = false;
    save.addEventListener('click', function () {
      clear(msg);
      var payload = { name: name.read(), description: desc.read(), iconUrl: icon.value.trim() };
      if (!Object.keys(payload.name).length) { msg.appendChild(notice('error', t('cu.needName'))); return; }
      if (open) {
        var rr = rules.read();
        if (rr.error) { msg.appendChild(notice('error', rr.error)); return; }
        payload.rules = rr.rules;
        payload.signupAt = signupAt.read();
        payload.startsAt = startsAt.read();
        payload.announceAt = announceAt.read();
        if (!payload.signupAt || !payload.startsAt) { msg.appendChild(notice('error', t('cu.needTimes'))); return; }
      }
      if (publish) payload.publish = publish.input.checked;
      busy(save, true);
      api(editing ? 'PUT' : 'POST', '/cups/list' + (editing ? '/' + c.id : ''), payload).then(function (out) {
        closePanel();
        toast('success', editing ? t('saved') : t('cu.created'));
        if (done) done(out.cup, out);
      }, function (err) { busy(save, false); fieldError(msg, err); });
    });
  }, function (err) { clear(body).appendChild(notice('error', cuErrText(err))); });
}

var CU_MATCH_COLOR = { pending: 'secondary', waiting: 'primary', open: 'primary', playing: 'warning', finished: 'success', bye: 'secondary', void: 'secondary' };
var CU_PLAYER_COLOR = { registered: 'primary', playing: 'warning', eliminated: 'secondary', finished: 'success', disqualified: 'error', withdrawn: 'secondary', removed: 'secondary' };
var CU_POLL_MS = 5000;
var cuPoll = null;

function matchTag(s) {
  return tag(t('ms.' + s), CU_MATCH_COLOR[s] || 'secondary');
}

function playerTag(s) {
  return tag(t('ps.' + s), CU_PLAYER_COLOR[s] || 'secondary');
}

function reasonText(r) {
  if (!r) return '';
  var parts = String(r).split(':');
  var main = has('why.' + parts[0]) ? t('why.' + parts[0]) : parts[0];
  if (parts[1]) main += ' (' + (has('ru.ns.' + parts[1]) ? t('ru.ns.' + parts[1]) : parts[1]) + ')';
  return main;
}

function payoutKind(k) {
  var m = /^reached_(\d+)$/.exec(k || '');
  if (m) return t('po.reached', { n: Number(m[1]) });
  return has('po.' + k) ? t('po.' + k) : k;
}

function sideName(s) {
  if (!s) return '';
  return s.displayName || s.username || (s.userId || '').slice(0, 8);
}

function isOpen(c) {
  return c.status === 'draft' || c.status === 'upcoming' || c.status === 'signup';
}

function stopPoll() {
  if (cuPoll) { clearInterval(cuPoll); cuPoll = null; }
}

function resolveUser(text) {
  text = String(text || '').trim();
  if (!text) return Promise.reject(new CuError(400, 'PLAYER_NOT_FOUND', '', 'userId'));
  if (isUUID(text)) return Promise.resolve(text);
  return api('GET', '/v2/console/account' + qs({ filter: text })).then(function (d) {
    var list = (d && d.users) || [];
    var hit = list.filter(function (u) { var x = u.user || u; return (x.username || '').toLowerCase() === text.toLowerCase(); })[0] || (list.length === 1 ? list[0] : null);
    if (!hit) throw new CuError(404, 'PLAYER_NOT_FOUND', '', 'userId');
    return (hit.user || hit).id;
  });
}

function cupViewPage(root) {
  stopPoll();
  var hm = /^#\/cups\/list\/([^/?]+)/.exec(location.hash);
  var id = hm ? decodeURIComponent(hm[1]) : '';
  var frame = pageFrame('cups/view', '', null);
  clear(root).appendChild(frame.el);
  var head = frame.el.querySelector('h1');
  var descEl = frame.el.querySelector('h1 + p');
  var actions = frame.el.querySelector('.flex.flex-wrap.gap-2');
  var back = h('a', { href: '#/cups/list', class: 'eco-link text-sm mb-2 inline-block', 'data-cup-back': '', text: t('cu.back') });
  frame.el.insertBefore(back, frame.el.firstChild);
  if (!id) { frame.content.appendChild(notice('error', t('code.CUP_NOT_FOUND'))); return; }
  var live = h('p', { class: 'text-xs text-secondary-500', 'data-cup-live': '' });
  var summary = h('div', { 'data-cup-summary': '' });
  var bracketBox = h('div', { 'data-cup-bracket-box': '' });
  var playersBox = h('div', { 'data-cup-players-box': '' });
  var payoutsBox = h('div', { 'data-cup-payouts-box': '' });
  frame.content.appendChild(h('div', { class: 'flex flex-col gap-4' }, live, summary, bracketBox, playersBox, payoutsBox));
  frame.content.appendChild(loading());
  var data = null;
  function busyUi() {
    return !!cuPanel || !!document.querySelector('[data-cup=confirm]');
  }
  function refresh(quiet) {
    return api('GET', '/cups/list/' + encodeURIComponent(id)).then(function (d) {
      if (!root.isConnected) { stopPoll(); return; }
      data = d;
      paint();
    }, function (err) {
      if (quiet) return;
      var l = frame.content.querySelector('p.text-secondary-500:last-child');
      if (l && l.textContent === t('loading')) l.remove();
      summary.appendChild(notice('error', cuErrText(err)));
    });
  }
  function act(path, body, confirmText, okLabel, doneText) {
    return (confirmText ? confirmBox(confirmText, okLabel) : Promise.resolve(true)).then(function (ok) {
      if (!ok) return false;
      return api(path.method || 'POST', '/cups/list/' + encodeURIComponent(id) + path.path, body).then(function (out) {
        toast('success', doneText || t('done'));
        if (out && out.cup) { data = out; paint(); } else refresh(true);
        return true;
      }, function (err) { toast('error', cuErrText(err)); return false; });
    });
  }
  function paint() {
    var c = data.cup;
    var l = frame.content.querySelector(':scope > p');
    if (l) l.remove();
    head.textContent = cupName(c);
    head.setAttribute('dir', 'auto');
    descEl.textContent = rulesSummary(c.rules);
    live.textContent = (c.status === 'running' || c.status === 'signup' || c.status === 'upcoming' ? t('cu.liveOn') + ' · ' : '') + t('cu.refreshed', { at: fmtDate(data.serverTime || new Date().toISOString()) });
    clear(actions);
    paintActions(c);
    clear(summary).appendChild(summaryCard(c));
    clear(bracketBox).appendChild(bracketCard(c, data.bracket));
    clear(playersBox).appendChild(playersCard(c, data.players || []));
    clear(payoutsBox).appendChild(payoutsCard(c, data.payouts || []));
  }
  function paintActions(c) {
    var open = isOpen(c);
    actions.appendChild(button({ label: t('cu.edit'), data: { 'cup-action': 'edit-cup' }, onclick: function () { cupEditor(c, function (nc, out) { data = out; paint(); }); } }));
    if (c.status === 'draft') actions.appendChild(button({ label: t('cu.publishNow'), variant: 'primary', data: { 'cup-action': 'publish' }, onclick: function () { act({ path: '/publish' }, {}, null, null, t('cu.published')); } }));
    if (open) actions.appendChild(button({ label: t('cu.startNow'), variant: 'primary', data: { 'cup-action': 'start' }, onclick: function () { act({ path: '/start' }, {}, t('cu.startConfirm', { n: c.players }), t('cu.startNow'), t('cu.started')); } }));
    if (open) actions.appendChild(button({ label: t('cu.addPlayer'), data: { 'cup-action': 'add-player' }, onclick: function () { addPlayerPanel(); } }));
    if (c.status !== 'finished' && c.status !== 'cancelled') actions.appendChild(button({ label: t('cu.cancel'), variant: 'danger', class: 'cu-btn--outline', data: { 'cup-action': 'cancel' }, onclick: function () { cancelPanel(c); } }));
    if (c.status === 'draft' || c.status === 'cancelled') actions.appendChild(button({ label: t('delete'), variant: 'danger', data: { 'cup-action': 'delete' }, onclick: function () {
      confirmBox(t('cu.deleteConfirm'), t('delete')).then(function (ok) {
        if (!ok) return;
        api('DELETE', '/cups/list/' + encodeURIComponent(id)).then(function () { toast('success', t('deleted')); location.hash = '#/cups/list'; }, function (err) { toast('error', cuErrText(err)); });
      });
    } }));
  }
  function summaryCard(c) {
    var r = c.rules || {};
    var champ = c.championId ? (data.players || []).filter(function (p) { return p.userId === c.championId; })[0] : null;
    var pool = c.pool || {};
    return card(h('div', { class: 'p-4 grid gap-4 md:grid-cols-2' },
      dl([
        [t('f.state'), h('span', { 'data-cup-value': 'status' }, statusTag(c.status))],
        [t('ru.mode'), h('span', { 'data-cup-value': 'mode', text: t('mode.' + ((r.timing && r.timing.mode) || 'window')) })],
        [t('ru.game'), r.game ? mono(r.game) : t('ru.external')],
        [t('cu.players'), h('span', { 'data-cup-value': 'players', text: t('cu.entrantsV', { n: fmtNum(c.status === 'running' || c.status === 'finished' ? c.players : (data.players || []).filter(function (p) { return p.status === 'registered'; }).length), cap: fmtNum(r.capacity) }) })],
        c.size ? [t('cu.size'), h('span', { 'data-cup-value': 'size', text: t('cu.sizeV', { n: c.size, rounds: c.rounds }) })] : null,
        c.status === 'running' ? [t('cu.round'), h('span', { 'data-cup-value': 'round', text: t('cu.roundV', { n: c.currentRound, of: c.rounds }) })] : null,
        champ ? [t('pl.champion'), h('span', { 'data-cup-value': 'champion' }, playerLink(champ.userId, champ.displayName || champ.username))] : null,
        c.cancelReason ? [t('cu.cancelReason'), h('span', { 'data-cup-value': 'cancel-reason', dir: 'auto', text: reasonText(c.cancelReason) })] : null
      ], 'cup'),
      dl([
        [t('cu.signupAt'), fmtDate(c.signupAt)],
        [t('cu.startsAt'), fmtDate(c.startsAt)],
        c.startedAt ? [t('cu.startedAt'), fmtDate(c.startedAt)] : null,
        c.finishedAt ? [t('cu.finishedAt'), fmtDate(c.finishedAt)] : null,
        [t('ru.feeAmount'), r.entry && r.entry.fee && r.entry.fee.amount > 0 ? h('span', {}, fmtNum(r.entry.fee.amount) + ' ', mono(r.entry.fee.currencyKey)) : t('ru.free')],
        pool.amount ? [t('ru.pool'), h('span', { 'data-cup-value': 'pool' }, fmtNum(pool.amount) + ' ', mono(pool.currencyKey || ''))] : null,
        c.templateKey ? [t('tp.template'), mono(c.templateKey)] : null,
        [t('f.id'), mono(c.id)]
      ], 'times')));
  }
  function bracketCard(c, br) {
    if (!br || !br.rounds || !br.rounds.length) return section(t('cu.bracket'), h('p', { class: 'text-sm text-secondary-500', 'data-cup-no-bracket': '', text: c.status === 'cancelled' ? t('cu.noBracketCancelled') : t('cu.noBracket') }), 'bracket');
    var cols = br.rounds.map(function (rd) {
      return h('div', { class: 'cu-round', 'data-cup-round': String(rd.round) },
        h('div', { class: 'cu-round-head' },
          h('span', { class: 'font-medium', text: roundName(rd.key) }),
          h('span', { class: 'text-xs text-secondary-500', text: t('ru.bestOfN', { n: rd.bestOf }) + ' · ' + t('rs.' + rd.status) })),
        h('div', { class: 'cu-round-body' }, rd.matches.map(function (m) { return matchCard(c, m); })));
    });
    var extra = br.third ? h('div', { class: 'mt-4' }, h('h3', { class: 'font-medium text-sm mb-2', text: t('rk.third') }), h('div', { class: 'cu-third' }, matchCard(c, br.third))) : null;
    return section(t('cu.bracket'), [h('div', { class: 'cu-bracket', 'data-cup-bracket': '', tabindex: '0', role: 'region', 'aria-label': t('cu.bracket') }, cols), extra], 'bracket');
  }
  function sideRow(m, s, wins) {
    if (!s) return h('div', { class: 'cu-side cu-side--empty' }, h('span', { class: 'text-secondary-400', text: m.status === 'bye' ? t('ms.bye') : t('cu.tbd') }));
    var won = m.winnerId && m.winnerId === s.userId;
    var lost = m.winnerId && m.winnerId !== s.userId;
    return h('div', { class: 'cu-side' + (won ? ' cu-side--won' : '') + (lost ? ' cu-side--lost' : ''), 'data-cup-side': s.userId },
      h('span', { class: 'cu-seed tabular-nums', text: s.seed ? digits(String(s.seed)) : '' }),
      h('span', { class: 'cu-side-name', dir: 'auto', text: sideName(s) }),
      s.ready && !done(m) ? h('span', { class: 'cu-ready', title: t('cu.ready'), 'aria-label': t('cu.ready'), text: '●' }) : null,
      h('span', { class: 'cu-wins tabular-nums', text: m.status === 'bye' || (!(m.games || []).length && m.status !== 'playing') ? '' : digits(String(wins || 0)) }));
  }
  function done(m) { return m.status === 'finished' || m.status === 'bye' || m.status === 'void'; }
  function matchCard(c, m) {
    var btn = h('button', { type: 'button', class: 'cu-match', 'data-cup-match': m.id, 'data-cup-status': m.status, 'aria-label': t('cu.matchOf', { a: sideName(m.a) || t('cu.tbd'), b: sideName(m.b) || t('cu.tbd') }) },
      sideRow(m, m.a, m.a && m.a.wins), sideRow(m, m.b, m.b && m.b.wins),
      h('div', { class: 'cu-match-foot' }, matchTag(m.status), m.reason && m.status !== 'bye' ? h('span', { class: 'text-xs text-secondary-500 truncate', text: reasonText(m.reason) }) : null));
    btn.addEventListener('click', function () { matchPanel(c, m); });
    return btn;
  }
  function matchPanel(c, m) {
    var body = h('div', { 'data-cup-match-panel': m.id });
    var foot = h('div', { class: 'flex flex-wrap gap-2 justify-end w-full' });
    sidePanel(t('cu.match') + ' · ' + (m.kind === 'third' ? t('rk.third') : roundName(m.roundKey)), body, [foot], 'match');
    body.appendChild(dl([
      [t('f.state'), matchTag(m.status)],
      [t('ru.bestOf'), t('ru.bestOfN', { n: m.bestOf })],
      [t('cu.sideA'), m.a ? h('span', {}, playerLink(m.a.userId, sideName(m.a)), ' · ', t('cu.winsN', { n: m.a.wins })) : t('cu.tbd')],
      [t('cu.sideB'), m.b ? h('span', {}, playerLink(m.b.userId, sideName(m.b)), ' · ', t('cu.winsN', { n: m.b.wins })) : (m.status === 'bye' ? t('ms.bye') : t('cu.tbd'))],
      m.winnerId ? [t('cu.winner'), h('span', { 'data-cup-value': 'winner', dir: 'auto', text: sideName(m.winnerId === (m.a && m.a.userId) ? m.a : m.b) })] : null,
      m.reason ? [t('cu.reason'), reasonText(m.reason)] : null,
      m.draws ? [t('cu.draws'), fmtNum(m.draws)] : null,
      m.opensAt ? [t('cu.opensAt'), fmtDate(m.opensAt)] : null,
      m.startsAt ? [t('cu.matchStarts'), fmtDate(m.startsAt)] : null,
      m.closesAt ? [t('cu.closesAt'), fmtDate(m.closesAt)] : null,
      m.finishedAt ? [t('cu.finishedAt'), fmtDate(m.finishedAt)] : null,
      m.gameMatchId ? [t('cu.gameMatch'), mono(m.gameMatchId)] : null,
      [t('f.id'), mono(m.id)]
    ], 'match'));
    var games = m.games || [];
    body.appendChild(h('div', { class: 'h-3' }));
    body.appendChild(section(t('cu.games'), games.length ? h('div', { class: 'cu-rows' }, games.map(function (g) {
      var w = g.winnerId ? (m.a && m.a.userId === g.winnerId ? sideName(m.a) : m.b && m.b.userId === g.winnerId ? sideName(m.b) : g.winnerId.slice(0, 8)) : '';
      return h('div', { class: 'cu-game', 'data-cup-game': String(g.no) },
        h('span', { class: 'font-medium', text: t('cu.gameN', { n: g.no }) }), matchTag(g.status === 'launching' ? 'waiting' : g.status),
        h('span', { class: 'text-sm', dir: 'auto', text: g.draw ? t('cu.draw') : w ? t('cu.wonBy', { name: w }) : reasonText(g.reason) }),
        h('span', { class: 'text-xs text-secondary-500', text: fmtDate(g.finishedAt || g.createdAt) }));
    })) : h('p', { class: 'text-sm text-secondary-500', text: t('cu.noGames') }), 'games'));
    var canForce = c.status === 'running' && !done(m) && m.a && m.b;
    if (canForce) {
      body.appendChild(h('div', { class: 'h-3' }));
      body.appendChild(section(t('cu.force'), [h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('cu.hint.force') }),
        h('div', { class: 'flex flex-wrap gap-2' },
          [m.a, m.b].map(function (s) {
            return button({ label: t('cu.winIs', { name: sideName(s) }), variant: 'primary', data: { 'cup-force': s.userId }, onclick: function () {
              act({ path: '/matches/' + m.id + '/result' }, { winnerId: s.userId }, t('cu.forceConfirm', { name: sideName(s) }), t('cu.forceOk'), t('cu.forced')).then(function (ok) { if (ok) closePanel(); });
            } });
          }),
          button({ label: t('cu.voidMatch'), variant: 'danger', data: { 'cup-force': 'void' }, onclick: function () {
            act({ path: '/matches/' + m.id + '/result' }, { void: true }, t('cu.voidConfirm'), t('cu.voidMatch'), t('cu.voided')).then(function (ok) { if (ok) closePanel(); });
          } }))], 'force'));
    }
    foot.appendChild(button({ label: t('close'), onclick: closePanel }));
  }
  function playersCard(c, players) {
    var open = isOpen(c);
    var running = c.status === 'running';
    var list = players.filter(function (p) { return p.status !== 'removed' && p.status !== 'withdrawn'; });
    var gone = players.length - list.length;
    var rows = list.slice().sort(function (a, b) { return (a.seed || 999) - (b.seed || 999) || String(a.joinedAt).localeCompare(String(b.joinedAt)); });
    var tbl = rows.length ? table([
      { label: t('cu.seed') }, { label: t('f.player') }, { label: t('f.state') }, { label: t('cu.trophies') }, { label: t('cu.place') }, { label: t('cu.fee') }, { label: t('cu.actions') }
    ], rows.map(function (p) {
      var acts = [];
      var activeP = p.status === 'registered' || p.status === 'playing';
      if (open && p.status === 'registered') acts.push(button({ label: t('remove'), variant: 'danger', data: { 'cup-player-action': 'remove' }, onclick: function () {
        act({ method: 'DELETE', path: '/players/' + p.userId }, undefined, t('cu.removeConfirm', { name: p.username }), t('remove'), t('cu.removed'));
      } }));
      if ((open || running) && activeP) acts.push(button({ label: t('cu.swap'), data: { 'cup-player-action': 'swap' }, onclick: function () { swapPanel(p); } }));
      if (running && activeP) acts.push(button({ label: t('cu.disqualify'), variant: 'danger', data: { 'cup-player-action': 'disqualify' }, onclick: function () { disqualifyPanel(p); } }));
      return { id: p.userId, item: p, cells: [
        h('span', { class: 'tabular-nums', text: p.seed ? digits(String(p.seed)) : t('none') }),
        playerLink(p.userId, p.displayName || p.username),
        playerTag(p.status),
        h('span', { class: 'tabular-nums', text: fmtNum(p.trophies || 0) }),
        h('span', { text: p.place ? placeText(p.place) : p.reached ? t('cu.reachedV', { n: p.reached }) : t('none') }),
        h('span', { class: 'text-xs', text: feeText(p) }),
        h('div', { class: 'flex flex-wrap gap-1' }, acts.length ? acts : t('none'))
      ] };
    })) : emptyCard(open ? t('cu.noPlayersYet') : t('cu.noPlayers'));
    return section(t('cu.playersT', { n: fmtNum(list.length) }), [tbl, gone ? h('p', { class: 'text-xs text-secondary-500 mt-2', text: t('cu.goneN', { n: gone }) }) : null], 'players');
  }
  function placeText(n) {
    return n === 1 ? t('pl.champion') : n === 2 ? t('pl.runnerUp') : n === 3 ? t('pl.third') : n === 4 ? t('pl.fourth') : t('cu.placeN', { n: n });
  }
  function feeText(p) {
    var keys = Object.keys(p.fee || {});
    var amount = keys.map(function (k) { return fmtNum(p.fee[k]) + ' ' + k; }).join('، ');
    return (amount ? amount + ' · ' : '') + t('fs.' + (p.feeState || 'none'));
  }
  function payoutsCard(c, payouts) {
    if (!payouts.length) return h('div', {});
    var failed = payouts.some(function (p) { return p.status === 'failed'; });
    var retry = failed ? button({ label: t('cu.retry'), data: { 'cup-action': 'retry-payouts' }, onclick: function () {
      busy(retry, true);
      api('POST', '/cups/payouts/retry', { cupId: id }).then(function (o) { toast('success', t('cu.retried', { n: o.requeued || 0 })); refresh(true); }, function (err) { busy(retry, false); toast('error', cuErrText(err)); });
    } }) : null;
    return section(t('cu.payouts'), [retry ? h('div', { class: 'mb-3' }, retry) : null, table([
      { label: t('f.player') }, { label: t('cu.kind') }, { label: t('ru.prize') }, { label: t('cu.amounts') }, { label: t('f.state') }, { label: t('f.date') }
    ], payouts.map(function (p) {
      var am = Object.keys(p.amounts || {}).map(function (k) { return fmtNum(p.amounts[k]) + ' ' + k; }).join('، ');
      return { id: p.id, item: p, cells: [
        playerLink(p.userId, p.username),
        h('span', { text: payoutKind(p.kind) }),
        p.prizeKey ? mono(p.prizeKey) : t('none'),
        h('span', { dir: 'auto', class: 'text-xs', text: am || t('none') }),
        h('span', { title: p.lastError || '' }, tag(t('pay.' + p.status), p.status === 'paid' ? 'success' : p.status === 'failed' ? 'error' : 'warning')),
        h('span', { class: 'text-xs', text: fmtDate(p.paidAt || p.createdAt) })
      ] };
    }))], 'payouts');
  }
  function addPlayerPanel() {
    var who = input('', { name: 'userId', ltr: true, placeholder: t('cu.hint.who') });
    var msg = h('div', {});
    var ok = button({ label: t('cu.addPlayer'), variant: 'primary', data: { 'cup-action': 'save-player' } });
    sidePanel(t('cu.addPlayer'), [field(t('f.player'), who, t('cu.hint.whoLong')), h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('cu.hint.add') }), msg], [button({ label: t('cancel'), onclick: closePanel }), ok], 'add-player');
    who.focus();
    ok.addEventListener('click', function () {
      clear(msg);
      busy(ok, true);
      resolveUser(who.value).then(function (uid) {
        return api('POST', '/cups/list/' + encodeURIComponent(id) + '/players', { userId: uid }).then(function (out) { closePanel(); toast('success', t('cu.added')); data = out; paint(); });
      }).catch(function (err) { busy(ok, false); fieldError(msg, err); });
    });
  }
  function swapPanel(p) {
    var who = input('', { name: 'to', ltr: true, placeholder: t('cu.hint.who') });
    var msg = h('div', {});
    var ok = button({ label: t('cu.swap'), variant: 'primary', data: { 'cup-action': 'save-swap' } });
    sidePanel(t('cu.swapTitle', { name: p.username }), [h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('cu.hint.swap') }), field(t('cu.swapTo'), who, t('cu.hint.whoLong')), msg], [button({ label: t('cancel'), onclick: closePanel }), ok], 'swap');
    who.focus();
    ok.addEventListener('click', function () {
      clear(msg);
      busy(ok, true);
      resolveUser(who.value).then(function (uid) {
        return api('POST', '/cups/list/' + encodeURIComponent(id) + '/swap', { from: p.userId, to: uid }).then(function (out) { closePanel(); toast('success', t('cu.swapped')); data = out; paint(); });
      }).catch(function (err) { busy(ok, false); fieldError(msg, err); });
    });
  }
  function disqualifyPanel(p) {
    var refund = check(false, t('cu.dqRefund'), 'refund');
    var msg = h('div', {});
    var ok = button({ label: t('cu.disqualify'), variant: 'danger', class: 'cu-btn--outline', data: { 'cup-action': 'save-disqualify' } });
    sidePanel(t('cu.dqTitle', { name: p.username }), [h('p', { class: 'text-sm mb-3', text: t('cu.hint.dq') }), refund, msg], [button({ label: t('cancel'), onclick: closePanel }), ok], 'disqualify');
    ok.addEventListener('click', function () {
      clear(msg);
      busy(ok, true);
      api('POST', '/cups/list/' + encodeURIComponent(id) + '/players/' + p.userId + '/disqualify', { refund: refund.input.checked }).then(function (out) { closePanel(); toast('success', t('cu.disqualified')); data = out; paint(); },
        function (err) { busy(ok, false); fieldError(msg, err); });
    });
  }
  function cancelPanel(c) {
    var reason = input('', { name: 'reason', placeholder: t('cu.hint.reason') });
    var msg = h('div', {});
    var ok = button({ label: t('cu.cancel'), variant: 'danger', class: 'cu-btn--outline', data: { 'cup-action': 'save-cancel' } });
    sidePanel(t('cu.cancelTitle'), [h('p', { class: 'text-sm mb-3', text: t('cu.hint.cancel') }), field(t('cu.cancelReason'), reason), msg], [button({ label: t('close'), onclick: closePanel }), ok], 'cancel');
    ok.addEventListener('click', function () {
      clear(msg);
      busy(ok, true);
      api('POST', '/cups/list/' + encodeURIComponent(id) + '/cancel', { reason: reason.value.trim() }).then(function (out) { closePanel(); toast('success', t('cu.cancelled')); data = out; paint(); },
        function (err) { busy(ok, false); fieldError(msg, err); });
    });
  }
  Promise.all([loadLanguages(), loadRefs('cups')]).then(function () { return refresh(false); }).then(function () {
    cuPoll = setInterval(function () {
      if (!root.isConnected || !document.querySelector('[data-cup-page="cups/view"]') || location.hash.indexOf('#/cups/list/' + id) !== 0) { stopPoll(); return; }
      if (!data || busyUi() || document.hidden) return;
      var st = data.cup.status;
      if (st === 'finished' || st === 'cancelled' || st === 'draft') return;
      refresh(true);
    }, CU_POLL_MS);
  });
}

function templatesPage(root) {
  var frame = pageFrame('cups/templates', t('desc.cups/templates'), [
    button({ label: t('tp.new'), variant: 'primary', data: { 'cup-action': 'new-template' }, onclick: function () { templateEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(listBox);
  function reload() {
    clear(listBox).appendChild(loading());
    Promise.all([api('GET', '/cups/templates'), loadLanguages()]).then(function (res) {
      var items = res[0].items || [];
      clear(listBox);
      if (!items.length) { listBox.appendChild(emptyCard(t('tp.empty'))); return; }
      listBox.appendChild(h('div', { 'data-cup-list': 'templates' }, table([
        { label: t('f.name') }, { label: t('f.state') }, { label: t('tp.schedule') }, { label: t('tp.nextStart') }, { label: t('tp.openCups') }, { label: t('f.updated') }
      ], items.map(function (x) {
        return { id: x.key, item: x, cells: [
          h('div', { class: 'flex flex-col min-w-0' }, h('span', { class: 'font-medium', dir: 'auto', text: textOf(x.name, x.key) }), mono(x.key), h('span', { class: 'text-xs text-secondary-500', text: rulesSummary(x.rules) })),
          tag(x.active ? t('active') : t('inactive'), x.active ? 'success' : 'secondary'),
          x.schedule ? mono(x.schedule) : h('span', { class: 'text-secondary-500', text: t('tp.manual') }),
          h('span', { class: 'text-xs', text: x.active && x.nextStart ? fmtDate(x.nextStart) : t('none') }),
          x.openCups ? h('a', { href: '#/cups/list?template=' + encodeURIComponent(x.key), class: 'eco-link tabular-nums', text: fmtNum(x.openCups) }) : h('span', { text: fmtNum(0) }),
          h('span', { class: 'text-xs', dir: 'auto', text: (x.updatedBy ? x.updatedBy + ' · ' : '') + fmtDate(x.updatedAt) })
        ] };
      }), function (x) { templateEditor(x, reload); })));
    }, function (err) { clear(listBox).appendChild(notice('error', cuErrText(err))); });
  }
  reload();
}

function templateEditor(x, done) {
  var editing = !!(x && x.id);
  var body = h('div', { 'data-cup-form': 'template' }, loading());
  var save = button({ label: t('save'), variant: 'primary', data: { 'cup-action': 'save-template' }, disabled: true });
  var foot = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (editing) {
    foot.unshift(button({ label: t('tp.runNow'), data: { 'cup-action': 'run-template' }, onclick: function () {
      confirmBox(t('tp.runConfirm'), t('tp.runNow')).then(function (ok) {
        if (!ok) return;
        api('POST', '/cups/templates/' + x.id + '/run', {}).then(function (out) { closePanel(); toast('success', t('cu.created')); location.hash = viewHref(out.cup.id); }, function (err) { toast('error', cuErrText(err)); });
      });
    } }));
    foot.unshift(button({ label: t('delete'), variant: 'danger', data: { 'cup-action': 'delete-template' }, onclick: function () {
      confirmBox(t('tp.deleteConfirm', { key: x.key }), t('delete')).then(function (ok) {
        if (!ok) return;
        api('DELETE', '/cups/templates/' + x.id).then(function () { closePanel(); toast('success', t('deleted')); done(); }, function (err) { toast('error', cuErrText(err)); });
      });
    } }));
  }
  sidePanel(editing ? t('tp.edit') : t('tp.new'), body, foot, 'template');
  Promise.all([loadRefs('cups', true), loadLanguages()]).then(function (res) {
    var refs = res[0];
    clear(body);
    var key = input(x ? x.key : '', { name: 'key', ltr: true, mono: true, placeholder: 'weekly_cup' });
    var name = translated(t('f.name'), 'name', x && x.name);
    var desc = translated(t('f.description'), 'description', x && x.description, { multiline: true });
    var icon = input(x && x.iconUrl || '', { name: 'iconUrl', ltr: true, placeholder: 'https://' });
    var active = check(x ? x.active : false, t('tp.active'), 'active');
    var schedule = input(x ? x.schedule : '0 18 * * 5', { name: 'schedule', ltr: true, mono: true, placeholder: '0 18 * * 5' });
    var signup = numInput(x ? x.signupMinutes : 1440, 'signupMinutes', t('tp.signupMinutes'));
    var announce = numInput(x ? x.announceMinutes : 0, 'announceMinutes', t('tp.announceMinutes'));
    var sort = numInput(x ? x.sort : 0, 'sort', t('f.sort'));
    var rules = rulesForm(x ? x.rules : refs.defaults);
    var msg = h('div', { 'data-cup-msg': '' });
    body.appendChild(h('p', { class: 'text-sm text-secondary-500 mb-4', text: t('tp.hint') }));
    body.appendChild(field(t('f.key'), key, t('hint.key')));
    body.appendChild(name);
    body.appendChild(desc);
    body.appendChild(field(t('f.icon'), icon));
    body.appendChild(active);
    body.appendChild(field(t('tp.schedule'), schedule, t('tp.hint.schedule')));
    body.appendChild(grid3(field(t('tp.signupMinutes'), signup, t('tp.hint.signup')), field(t('tp.announceMinutes'), announce, t('tp.hint.announce')), field(t('f.sort'), sort)));
    rules.el.forEach(function (s) { body.appendChild(s); body.appendChild(h('div', { class: 'h-3' })); });
    body.appendChild(msg);
    save.disabled = false;
    save.addEventListener('click', function () {
      clear(msg);
      var rr = rules.read();
      if (rr.error) { msg.appendChild(notice('error', rr.error)); return; }
      var n = [signup, announce, sort].map(function (el) { return numberValue(el); });
      if (n.some(function (v) { return v === null || isNaN(v); })) { msg.appendChild(notice('error', t('ru.badNumber', { field: t('tp.signupMinutes') }))); return; }
      var payload = { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), active: active.input.checked, schedule: schedule.value.trim(),
        signupMinutes: n[0], announceMinutes: n[1], sort: n[2], rules: rr.rules };
      busy(save, true);
      api(editing ? 'PUT' : 'POST', '/cups/templates' + (editing ? '/' + x.id : ''), payload).then(function () { closePanel(); toast('success', t('saved')); done(); }, function (err) { busy(save, false); fieldError(msg, err); });
    });
  }, function (err) { clear(body).appendChild(notice('error', cuErrText(err))); });
}

function cupSettingsPage(root) {
  loadRefs('cups', true).then(function () {
    settingsPage(root, 'cups/settings', '/cups', function (cur) {
      var ev = cur.events, ds = cur.display, lm = cur.limits;
      var enabled = check(cur.enabled, t('cs.enabled'), 'enabled');
      var notify = check(ev.notify, t('cs.notify'), 'events.notify');
      var code = numInput(ev.code, 'events.code', t('cs.code'));
      var inApp = check(ev.inApp, t('cs.inApp'), 'events.inApp');
      var publish = check(ev.publish, t('cs.publish'), 'events.publish');
      var listLimit = numInput(ds.listLimit, 'display.listLimit', t('cs.listLimit'));
      var historyLimit = numInput(ds.historyLimit, 'display.historyLimit', t('cs.historyLimit'));
      var finishedDays = numInput(ds.finishedDays, 'display.finishedDays', t('cs.finishedDays'));
      var maxCapacity = numInput(lm.maxCapacity, 'limits.maxCapacity', t('cs.maxCapacity'));
      var activePer = numInput(lm.activePerPlayer, 'limits.activePerPlayer', t('cs.activePerPlayer'));
      var tick = numInput(lm.tickSeconds, 'limits.tickSeconds', t('cs.tickSeconds'));
      var rules = rulesForm(cur.defaults);
      var nums = [[code, 'cs.code'], [listLimit, 'cs.listLimit'], [historyLimit, 'cs.historyLimit'], [finishedDays, 'cs.finishedDays'], [maxCapacity, 'cs.maxCapacity'], [activePer, 'cs.activePerPlayer'], [tick, 'cs.tickSeconds']];
      return {
        el: [
          section(t('cs.s.general'), [enabled, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('cs.hint.enabled') })], 'general'),
          section(t('cs.s.events'), [notify, field(t('cs.code'), code, t('cs.hint.code')), inApp, publish], 'events'),
          section(t('cs.s.limits'), [grid3(field(t('cs.maxCapacity'), maxCapacity, t('cs.hint.maxCapacity')), field(t('cs.activePerPlayer'), activePer, t('hint.zero')), field(t('cs.tickSeconds'), tick, t('cs.hint.tick'))),
            grid3(field(t('cs.listLimit'), listLimit), field(t('cs.historyLimit'), historyLimit), field(t('cs.finishedDays'), finishedDays, t('cs.hint.finishedDays')))], 'limits'),
          h('h2', { class: 'font-semibold mt-2', text: t('cs.s.defaults') }),
          h('p', { class: 'text-sm text-secondary-500 -mt-2', text: t('cs.hint.defaults') })
        ].concat(rules.el),
        read: function (c) {
          for (var i = 0; i < nums.length; i++) {
            var n = numberValue(nums[i][0]);
            if (n === null || isNaN(n)) return t('ru.badNumber', { field: t(nums[i][1]) });
          }
          var rr = rules.read();
          if (rr.error) return rr.error;
          c.enabled = enabled.input.checked;
          c.events = { notify: notify.input.checked, code: numberValue(code), inApp: inApp.input.checked, publish: publish.input.checked };
          c.display = { listLimit: numberValue(listLimit), historyLimit: numberValue(historyLimit), finishedDays: numberValue(finishedDays) };
          c.limits = { maxCapacity: numberValue(maxCapacity), activePerPlayer: numberValue(activePer), tickSeconds: numberValue(tick) };
          c.defaults = rr.rules;
          return null;
        }
      };
    });
  });
}

var CU_ROOT = 'cup-root';
var CU_GROUPS = { cups: ['list', 'templates', 'settings'] };
var CU_ORDER = ['cups'];
var CU_RENDER = { 'cups/list': cupsPage, 'cups/view': cupViewPage, 'cups/templates': templatesPage, 'cups/settings': cupSettingsPage };
var CU_HASH = /^#\/(cups)(?:\/([^/?]*))?(?:\/([^/?]+))?/;
var cuPending = null;
var cuRouter = null;
var cuMounted = null;
var cuScheduled = false;

function cuRoute(hash) {
  var m = CU_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  var pages = CU_GROUPS[m[1]];
  if (m[2] === 'list' && m[3]) return m[1] + '/view';
  return m[1] + '/' + (pages.indexOf(m[2]) !== -1 ? m[2] : pages[0]);
}

function cuAllowed() {
  return !!token();
}

function cuTitle() {
  var page = cuRoute();
  return page ? t('group.' + page.split('/')[0]) + ' - ' + t('nav.' + page) : '';
}

function cuAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || cuRouter) return !!cuRouter;
  try {
    var layout = r.resolve('/').matched[0];
    CU_ORDER.forEach(function (g) {
      if (r.hasRoute(g)) return;
      r.addRoute({ path: '/' + g + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: g, component: { name: 'CupsPage', render: function () { return null; } }, meta: { get title() { return cuTitle(); } } }] });
    });
    cuRouter = r;
  } catch (e) {
    return false;
  }
  if (cuPending) {
    var target = cuPending;
    cuPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function cuNavLink(g, page, active) {
  var id = g + '/' + page;
  var on = id === active || (page === 'list' && active === g + '/view');
  return h('li', {}, h('a', { href: '#/' + id, 'data-cup-nav': id, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + id) }))));
}

function cuModOn(id) {
  var M = window.NakamaConsole && window.NakamaConsole.modules;
  return !M || !M.on || M.on(id);
}

function cuSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelectorAll('[data-cup-group]');
  if (!group || !cuAllowed() || !cuModOn('cups')) { Array.prototype.forEach.call(existing, function (e) { e.remove(); }); return; }
  var target = document.querySelector('.x-drawer [data-clan-group="trophies"]') || document.querySelector('.x-drawer [data-quiz-group]') || document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = cuLang() + '|' + active;
  var mine = document.querySelector('[data-cup-group="cups"]');
  if (mine && existing.length === 1 && mine.getAttribute('data-key') === key && mine.nextElementSibling === target) return;
  Array.prototype.forEach.call(existing, function (e) { e.remove(); });
  CU_ORDER.forEach(function (g) {
    target.before(h('div', { 'data-cup-group': g, 'data-plane-group': g, 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
      h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group.' + g) }),
      h('nav', { 'aria-label': t('group.' + g) }, h('ul', { class: 'space-y-0.5' }, CU_GROUPS[g].map(function (p) { return cuNavLink(g, p, active); })))));
  });
}

function cuShowMains() {
  var hidden = document.querySelectorAll('[data-cup-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-cup-hidden'); }
}

function cuUnmount() {
  var root = document.getElementById(CU_ROOT);
  if (root) root.remove();
  cuShowMains();
  closePanel();
  stopPoll();
  cuMounted = null;
}

function cuSync() {
  cuAddRoutes();
  var page = cuRoute();
  cuSyncNav(page);
  if (!page) { if (cuMounted || document.getElementById(CU_ROOT)) cuUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-cup-hidden', ''); }
  }
  var root = document.getElementById(CU_ROOT);
  if (!root) {
    root = h('div', { id: CU_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-cup': 'root' });
    cuMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + cuLang() + '|' + cuModOn('cups') + (page === 'cups/view' || page === 'cups/list' ? '|' + location.hash : '');
  if (cuMounted !== key) {
    cuMounted = key;
    closePanel();
    if (cuModOn('cups')) CU_RENDER[page](root);
    else window.NakamaConsole.modules.offPage(root, 'cups');
  }
}

function cuSchedule() {
  if (cuScheduled) return;
  cuScheduled = true;
  requestAnimationFrame(function () { cuScheduled = false; cuSync(); });
}

function cuBoot() {
  window.addEventListener('hashchange', cuSchedule);
  window.addEventListener('popstate', cuSchedule);
  window.addEventListener('nakama-console:ready', cuSchedule);
  window.addEventListener('nakama-console:modules', cuSchedule);
  new MutationObserver(cuSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { cuMounted = null; cuSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (cuAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  cuSync();
}

window.NakamaCup = { version: 1, pages: function () {
  var out = [];
  CU_ORDER.forEach(function (g) { CU_GROUPS[g].forEach(function (p) { out.push({ id: g + '/' + p, group: g, href: '#/' + g + '/' + p, label: t('nav.' + g + '/' + p) }); }); });
  return out;
} };

if (CU_HASH.test(location.hash)) {
  cuPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', cuBoot);
else cuBoot();
})();
