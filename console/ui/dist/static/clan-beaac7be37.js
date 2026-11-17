(function () {
'use strict';
var CL_TEXT = {
  fa: {
    'group.trophies': 'جام و لیگ', 'group.clans': 'کلن‌ها',
    'nav.trophies/leagues': 'لیگ‌ها', 'nav.trophies/players': 'جام بازیکنان', 'nav.trophies/seasons': 'فصل‌ها', 'nav.trophies/settings': 'تنظیمات جام',
    'nav.clans/list': 'فهرست کلن‌ها', 'nav.clans/levels': 'سطح‌های کلن', 'nav.clans/wars': 'جنگ‌های کلن', 'nav.clans/rounds': 'دورهای جنگ', 'nav.clans/settings': 'تنظیمات کلن',
    'desc.trophies/leagues': 'لیگ‌ها پله‌های جام هستند. هر بازیکن در لیگی است که حداقل جامش را دارد. اینجا لیگ بسازید، جایزهٔ رسیدن به آن را بگذارید و بگویید بازیکن از لیگ پایین‌تر بیفتد یا نه.',
    'desc.trophies/players': 'جام هر بازیکن، بهترین رکورد و لیگش. روی یک ردیف بزنید تا تاریخچهٔ جام و جایزه‌های لیگ را ببینید و در صورت نیاز جامش را عوض کنید.',
    'desc.trophies/seasons': 'در پایان هر فصل جام بازیکن‌ها طبق قانون تنظیمات کم می‌شود. اینجا فصل‌های قبلی و زمان پایان فصل بعدی را می‌بینید و می‌توانید همین حالا فصل را تمام کنید.',
    'desc.trophies/settings': 'جام از نتیجهٔ بازی‌ها می‌آید. اینجا بگویید کدام رویداد نتیجهٔ بازی است، هر برد و باخت چند جام دارد، جایزه‌های لیگ چطور داده شوند و فصل کی تمام شود.',
    'desc.clans/list': 'همهٔ کلن‌ها با سطح، عضو، جام و نتیجهٔ جنگ‌ها. روی یک کلن بزنید تا پروفایلش را ویرایش کنید، نقش عضوها را عوض کنید، عضوی را بیرون کنید یا امتیاز تجربه بدهید.',
    'desc.clans/levels': 'جاده‌ی سطح کلن: هر سطح چقدر امتیاز تجربه می‌خواهد، چند عضو جا دارد و چه جایزه‌ای می‌دهد. سطح ۱ باید از صفر شروع شود و امتیاز هر سطح از قبلی بیشتر باشد.',
    'desc.clans/wars': 'قالب جنگ‌ها. جنگ امتیازی: هر کار عضوها امتیاز دارد. جنگ دوئل: عضوهای دو کلن با هم بازی می‌کنند و برد امتیاز دارد. با زمان‌بندی، دورها خودکار ساخته می‌شوند.',
    'desc.clans/rounds': 'هر دور یک بار اجرای یک قالب جنگ است: ثبت‌نام، شروع، جنگ‌ها و نتیجه. اینجا دور تازه بسازید، شروع کنید، زودتر تمامش کنید یا لغوش کنید.',
    'desc.clans/settings': 'قانون‌های کلن: ساختن کلن و هزینه‌اش، نشان‌ها، اختیار هر نقش، زمان‌های انتظار، راه‌های گرفتن امتیاز تجربه و رویدادها.',
    'save': 'ذخیره', 'saved': 'ذخیره شد', 'deleted': 'حذف شد', 'cancel': 'انصراف', 'close': 'بستن', 'remove': 'حذف', 'delete': 'حذف', 'search': 'جستجو', 'all': 'همه',
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'yes': 'بله', 'no': 'خیر',
    'active': 'فعال', 'inactive': 'غیرفعال', 'on': 'روشن', 'off': 'خاموش', 'confirm.title': 'مطمئن هستید؟', 'lang.text': '{label} ({lang})', 'updated': 'آخرین تغییر: {by} · {at}', 'none': '—',
    'missing': 'پیدا نشد', 'noPrizes': 'افزونهٔ جایزه بارگذاری نشده است؛ جایزه‌ها ثبت می‌شوند ولی پرداخت نمی‌شوند.',
    'f.name': 'نام', 'f.description': 'توضیح', 'f.key': 'کلید', 'f.icon': 'نشانی آیکون', 'f.color': 'رنگ', 'f.active': 'فعال باشد', 'f.state': 'وضعیت', 'f.metadata': 'داده‌ی اضافه (JSON)',
    'f.player': 'بازیکن', 'f.date': 'زمان', 'f.updated': 'آخرین تغییر', 'f.created': 'ساخته‌شده',
    'hint.key': 'حرف کوچک انگلیسی، عدد و _ . - ؛ بازی با همین کلید کار می‌کند.', 'hint.json': 'یک شیء JSON، مثل {"badge":"gold"}', 'hint.tags': 'با ویرگول جدا کنید', 'hint.zero': '۰ یعنی بدون محدودیت',
    'src.economy_event': 'رویداد اقتصاد', 'src.game_event': 'رویداد بازی',
    'tr.unranked': 'بدون لیگ',
    'lg.new': 'لیگ تازه', 'lg.min': 'حداقل جام', 'lg.from': 'از {n} جام', 'lg.rules': 'قانون‌ها', 'lg.prize': 'جایزه', 'lg.count': 'بازیکن', 'lg.players': '{n} بازیکن جام دارند', 'lg.unranked': '{n} بازیکن بدون لیگ',
    'lg.floorOn': 'پایین نمی‌افتد', 'lg.resetTo': 'آخر فصل: {n}', 'lg.empty': 'هنوز لیگی ساخته نشده است. همهٔ بازیکن‌ها «بدون لیگ» هستند.', 'lg.delete': 'لیگ {key} حذف شود؟ بازیکن‌های آن به لیگ پایین‌تر می‌روند.',
    'lg.hint': 'نام و آیکون همان چیزی است که بازیکن در بازی می‌بیند.', 'lg.floor': 'بازیکن از این لیگ پایین نیفتد', 'lg.hint.floor': 'وقتی بازیکن به این لیگ رسید، با باخت جامش از حداقل این لیگ کمتر نمی‌شود.',
    'lg.reset': 'جام آخر فصل', 'lg.hint.reset': 'فقط وقتی قانون فصل «برگشت به عدد لیگ» است. خالی یعنی تغییر نکند.', 'lg.hint.min': 'دو لیگ نمی‌توانند حداقل یکسان داشته باشند.',
    'lg.hint.prize': 'وقتی بازیکن اولین بار به این لیگ برسد داده می‌شود.',
    'tp.search': 'بازیکن', 'tp.hint.search': 'نام کاربری یا شناسهٔ کامل', 'tp.league': 'لیگ', 'tp.trophies': 'جام', 'tp.best': 'بهترین', 'tp.bestV': 'همیشه {b} · این فصل {s}', 'tp.record': 'برد / باخت / مساوی',
    'tp.wld': '{w} / {l} / {d}', 'tp.empty': 'بازیکنی با جام پیدا نشد.', 'tp.open': 'باز کردن این بازیکن', 'tp.retry': 'تلاش دوباره برای جایزه‌های ناموفق', 'tp.retried': '{n} جایزه دوباره در صف رفت',
    'tp.change': 'تغییر جام', 'tp.hint.change': 'تغییر در تاریخچه با نام شما ثبت می‌شود و لیگ بازیکن دوباره حساب می‌شود.', 'tp.amount': 'مقدار', 'tp.modeL': 'نوع تغییر', 'tp.mode.set': 'برابر شود با', 'tp.mode.add': 'اضافه یا کم شود',
    'tp.apply': 'اعمال', 'tp.log': 'تاریخچهٔ جام', 'tp.noLog': 'هنوز تغییری ثبت نشده است.', 'tp.reason': 'دلیل', 'tp.delta': 'تغییر', 'tp.after': 'قبل → بعد', 'tp.opponent': 'حریف',
    'tp.reason.win': 'برد', 'tp.reason.loss': 'باخت', 'tp.reason.draw': 'مساوی', 'tp.reason.admin': 'ادمین', 'tp.reason.season': 'پایان فصل', 'tp.move': '{a} ← {b}',
    'tp.prizes': 'جایزه‌های لیگ', 'tp.noPrizes': 'هنوز جایزهٔ لیگی نگرفته است.', 'tp.season': 'فصل', 'tp.ever': 'همیشگی',
    'pay.pending': 'در صف', 'pay.paid': 'پرداخت شد', 'pay.failed': 'ناموفق',
    'ss.reset': 'پایان فصل همین حالا', 'ss.confirm': 'فصل فعلی همین حالا تمام شود؟ جام بازیکن‌ها با قانون «{mode}» کم می‌شود و این کار برنمی‌گردد.', 'ss.done': 'فصل تمام شد',
    'ss.hint.reset': 'فصل تازه بلافاصله شروع می‌شود.', 'ss.auto': 'پایان خودکار', 'ss.schedule': 'زمان‌بندی', 'ss.modeL': 'قانون پایان فصل', 'ss.next': 'پایان بعدی', 'ss.manual': 'فقط دستی',
    'ss.mode.cut': 'کم کردن جام بالای یک عدد', 'ss.mode.league': 'برگشت به عدد هر لیگ', 'ss.mode.zero': 'صفر شدن همه',
    'ss.number': 'فصل', 'ss.n': 'فصل {n}', 'ss.started': 'شروع', 'ss.ends': 'پایان برنامه‌ریزی‌شده', 'ss.ended': 'پایان', 'ss.current': 'فصل جاری', 'ss.players': 'بازیکن کم‌شده', 'ss.actor': 'به دست',
    'ss.scheduler': 'زمان‌بند', 'ss.empty': 'هنوز فصلی ثبت نشده است.',
    'ts.general': 'کلی', 'ts.enabled': 'جام روشن باشد', 'ts.hint.enabled': 'خاموش که باشد نتیجهٔ بازی‌ها جام نمی‌دهد و نمی‌گیرد.',
    'ts.match': 'نتیجهٔ بازی', 'ts.hint.match': 'جام از رویدادی می‌آید که بازی بعد از هر مسابقه می‌فرستد. نام فیلدهای همان رویداد را بنویسید.',
    'ts.source': 'نوع رویداد', 'ts.event': 'نام رویداد', 'ts.hint.event': 'خالی یعنی هیچ رویدادی جام نمی‌دهد.', 'ts.won': 'فیلد برد', 'ts.hint.won': 'مقدار true یعنی بازیکن برده است.',
    'ts.draw': 'فیلد مساوی', 'ts.hint.draw': 'اختیاری', 'ts.matchField': 'فیلد شناسهٔ مسابقه', 'ts.hint.matchField': 'یک مسابقه دو بار حساب نمی‌شود.', 'ts.opponent': 'فیلد حریف',
    'ts.hint.opponent': 'شناسهٔ حریف برای حساب جام بر اساس اختلاف.', 'ts.modeField': 'فیلد حالت بازی', 'ts.hint.modeField': 'اختیاری', 'ts.modes': 'حالت‌هایی که جام دارند',
    'ts.hint.modes': 'خالی یعنی همهٔ حالت‌ها', 'ts.skip': 'فیلد رد کردن', 'ts.hint.skip': 'اگر این فیلد true باشد مسابقه جام ندارد (مثلاً بازی دوستانه).',
    'ts.amounts': 'مقدار جام', 'ts.start': 'جام شروع', 'ts.hint.start': 'جام بازیکن تازه', 'ts.win': 'جام برد', 'ts.loss': 'جام باخت', 'ts.drawA': 'جام مساوی', 'ts.hint.drawA': 'می‌تواند منفی باشد',
    'ts.scale': 'جام بر اساس اختلاف با حریف تغییر کند', 'ts.hint.scale': 'برد از حریف قوی‌تر جام بیشتر و باخت به حریف ضعیف‌تر جام بیشتری کم می‌کند.',
    'ts.scalePer': 'به ازای هر چند جام اختلاف', 'ts.scaleStep': 'چند جام تغییر', 'ts.minWin': 'کمترین جام برد', 'ts.maxWin': 'بیشترین جام برد', 'ts.minLoss': 'کمترین جام باخت', 'ts.maxLoss': 'بیشترین جام باخت',
    'ts.floors': 'کف لیگ‌ها رعایت شود', 'ts.hint.floors': 'در لیگی که «پایین نمی‌افتد» دارد، جام از حداقل لیگ کمتر نمی‌شود.',
    'ts.prizes': 'جایزه‌های لیگ', 'ts.oncePer': 'هر جایزه چند بار', 'ts.hint.oncePer': 'یک بار برای همیشه یا یک بار در هر فصل', 'ts.once.ever': 'یک بار برای همیشه', 'ts.once.season': 'یک بار در هر فصل',
    'ts.onlyRising': 'فقط وقتی با برد به لیگ می‌رسد (نه با تغییر ادمین)', 'ts.tries': 'دفعات تلاش', 'ts.hint.tries': 'اگر پرداخت جایزه خطا بدهد چند بار دوباره تلاش شود',
    'ts.season': 'فصل', 'ts.seasonOn': 'فصل خودکار تمام شود', 'ts.schedule': 'زمان‌بندی (cron)', 'ts.hint.schedule': 'مثلاً 0 0 1 * * یعنی اول هر ماه', 'ts.hint.sMode': 'با پایان فصل جام‌ها چطور کم شوند',
    'ts.above': 'بالای چند جام', 'ts.hint.above': 'فقط جام بیشتر از این عدد کم می‌شود', 'ts.keep': 'چند درصد بماند', 'ts.hint.keep': 'از جام بالای آن عدد',
    'ts.mm': 'هم‌تیمی‌یابی', 'ts.mmOn': 'حریف هم‌جام پیدا شود', 'ts.hint.mm': 'جام بازیکن به صف پیدا کردن حریف داده می‌شود.', 'ts.games': 'بازی‌ها', 'ts.hint.games': 'هیچ‌کدام یعنی همهٔ بازی‌ها',
    'ts.range': 'بازهٔ جام حریف', 'ts.hint.range': 'حریف حداکثر این قدر جام کمتر یا بیشتر دارد', 'ts.property': 'نام ویژگی در صف', 'ts.hint.property': 'نام عددی که به صف فرستاده می‌شود',
    'ts.display': 'نمایش', 'ts.hist': 'تعداد تاریخچه برای بازیکن', 'ts.hint.hist': 'بازیکن چند ردیف آخر را می‌بیند', 'ts.logDays': 'نگه‌داری تاریخچه (روز)', 'ts.hint.logDays': 'ردیف‌های قدیمی‌تر پاک می‌شوند',
    'ts.showOpponent': 'نام حریف در تاریخچه دیده شود',
    'ts.events': 'رویدادها و اعلان', 'ts.publish': 'تغییر را به سامانهٔ رویداد بفرست', 'ts.notify': 'به بازیکن اعلان بده', 'ts.nCode': 'کد اعلان تغییر جام', 'ts.hint.nCode': 'بازی با این کد اعلان را می‌شناسد',
    'ts.lCode': 'کد اعلان تغییر لیگ', 'ts.hint.lCode': 'وقتی لیگ بازیکن عوض می‌شود',
    'jt.open': 'آزاد', 'jt.closed': 'با درخواست', 'jt.invite': 'فقط با دعوت',
    'role.leader': 'رهبر', 'role.co_leader': 'معاون', 'role.elder': 'ریش‌سفید', 'role.member': 'عضو',
    'cl.search': 'کلن', 'cl.hint.search': 'نام، برچسب یا شناسه', 'cl.joinType': 'نوع عضویت', 'cl.sortL': 'مرتب‌سازی', 'cl.sort.trophies': 'بیشترین جام', 'cl.sort.level': 'بالاترین سطح',
    'cl.sort.members': 'بیشترین عضو', 'cl.sort.new': 'تازه‌ترین', 'cl.sort.wars': 'بیشترین برد جنگ', 'cl.clan': 'کلن', 'cl.level': 'سطح', 'cl.lv': 'سطح {n}', 'cl.members': 'عضو', 'cl.of': '{n} از {m}',
    'cl.wars': 'جنگ‌ها (برد / باخت / مساوی)', 'cl.empty': 'کلنی پیدا نشد.', 'cl.xpOf': '{x} از {n} امتیاز', 'cl.xpMax': '{x} امتیاز · آخرین سطح',
    'cl.tag': 'برچسب', 'cl.warStats': '{w} برد · {l} باخت · {d} مساوی · {p} امتیاز', 'cl.requests': 'درخواست‌های در انتظار',
    'cl.profile': 'پروفایل', 'cl.lang': 'زبان', 'cl.avatar': 'نشانی تصویر', 'cl.emblem': 'نماد', 'cl.frame': 'قاب', 'cl.minTrophies': 'حداقل جام برای ورود', 'cl.hint.minTrophies': 'بازیکن کم‌جام‌تر نمی‌تواند بیاید',
    'cl.xp': 'امتیاز تجربه', 'cl.hint.xp': 'امتیاز با نام ادمین در تاریخچهٔ کلن ثبت می‌شود و ممکن است کلن سطح بگیرد.', 'cl.xpAmount': 'مقدار امتیاز', 'cl.addXp': 'افزودن امتیاز', 'cl.xpAdded': '{n} امتیاز اضافه شد',
    'cl.membersT': 'عضوها ({n})', 'cl.noMembers': 'عضوی ندارد.', 'cl.memberLine': '{x} امتیاز · {p} امتیاز جنگ · {t} جام', 'cl.role': 'نقش', 'cl.setRole': 'تغییر نقش', 'cl.kick': 'بیرون کردن',
    'cl.kickConfirm': '{name} از کلن بیرون شود؟ تا مدتی نمی‌تواند برگردد.', 'cl.kicked': 'بیرون شد', 'cl.leaderConfirm': '{name} رهبر شود؟ رهبر فعلی معاون می‌شود.',
    'cl.xpLog': 'تاریخچهٔ امتیاز', 'cl.noXp': 'هنوز امتیازی نگرفته است.', 'cl.source': 'منبع', 'cl.amount': 'مقدار', 'cl.warsT': 'جنگ‌های اخیر', 'cl.noWars': 'هنوز در جنگی نبوده است.',
    'cl.delete': 'حذف کلن', 'cl.deleteConfirm': 'کلن «{name}» حذف شود؟ همهٔ عضوها بیرون می‌روند و این کار برنمی‌گردد.',
    'xs.admin': 'ادمین', 'xs.war_win': 'برد جنگ', 'xs.war_draw': 'مساوی جنگ', 'xs.war_lose': 'باخت جنگ', 'xs.war_bye': 'جنگ بدون حریف',
    'lv.hint': 'تغییر جاده روی همهٔ کلن‌ها اثر دارد: سطح هر کلن و ظرفیت عضوش دوباره حساب می‌شود.', 'lv.level': 'سطح', 'lv.xp': 'امتیاز لازم', 'lv.hint.xp': 'کل امتیاز از اول', 'lv.cap': 'ظرفیت عضو',
    'lv.prize': 'جایزهٔ رسیدن', 'lv.perks': 'امتیازهای ویژه (JSON)', 'lv.clans': '{n} کلن در این سطح', 'lv.new': 'سطح تازه', 'lv.add': 'افزودن سطح',
    'mode.points': 'امتیازی', 'mode.duel': 'دوئل',
    'wr.new': 'قالب جنگ تازه', 'wr.mode': 'نوع جنگ', 'wr.schedule': 'زمان‌بندی', 'wr.manual': 'دستی', 'wr.duration': 'مدت (دقیقه)', 'wr.minutes': '{n} دقیقه', 'wr.signup': 'ثبت‌نام',
    'wr.signup.auto': 'خودکار', 'wr.signup.leader': 'با رهبر', 'wr.next': 'شروع بعدی', 'wr.empty': 'هنوز قالب جنگی ساخته نشده است.', 'wr.delete': 'قالب جنگ {key} حذف شود؟',
    'wr.config': 'قالب جنگ', 'wr.hint': 'هر دور جنگ یک نسخه از این قانون‌ها را نگه می‌دارد؛ تغییر قالب روی دورهای در جریان اثر ندارد.',
    'wr.hint.mode': 'امتیازی: کارهای عضوها امتیاز دارد. دوئل: عضوها با عضوهای کلن حریف بازی می‌کنند.',
    'wr.timing': 'زمان', 'wr.hint.schedule': 'cron؛ خالی یعنی فقط دستی. مثال: 0 18 * * 5 یعنی هر جمعه ساعت ۱۸', 'wr.hint.duration': 'از شروع تا پایان جنگ', 'wr.hint.signup': 'خودکار: همهٔ کلن‌های مجاز. با رهبر: رهبر ثبت‌نام می‌کند.',
    'wr.signupMinutes': 'مهلت ثبت‌نام (دقیقه)', 'wr.hint.signupMinutes': 'چقدر قبل از شروع ثبت‌نام باز شود', 'wr.pairing': 'جفت کردن کلن‌ها', 'wr.pair.random': 'تصادفی', 'wr.pair.level': 'هم‌سطح', 'wr.pair.trophies': 'هم‌جام',
    'wr.pointsRules': 'قانون‌های امتیاز', 'wr.hint.points': 'هر قانون می‌گوید کدام رویداد چند امتیاز برای کلن دارد. با فیلد و مقدار می‌توانید فقط رویدادهای خاصی را حساب کنید.',
    'wr.source': 'نوع رویداد', 'wr.event': 'رویداد', 'wr.field': 'فیلد', 'wr.equals': 'برابر با', 'wr.points': 'امتیاز', 'wr.perValue': 'امتیاز ضرب در مقدار فیلد', 'wr.addPoint': 'افزودن قانون',
    'wr.duelRules': 'قانون دوئل', 'wr.hint.duel': 'رویداد نتیجهٔ مسابقه باید حریف را داشته باشد؛ فقط مسابقه با عضو کلن حریف حساب می‌شود.', 'wr.won': 'فیلد برد', 'wr.drawF': 'فیلد مساوی',
    'wr.opp': 'فیلد حریف', 'wr.hint.opp': 'شناسهٔ بازیکن حریف', 'wr.winPoints': 'امتیاز برد', 'wr.drawPoints': 'امتیاز مساوی',
    'wr.scoring': 'شمارش', 'wr.matchField': 'فیلد شناسهٔ مسابقه', 'wr.hint.matchField': 'یک مسابقه دو بار حساب نمی‌شود', 'wr.attacks': 'حمله برای هر عضو', 'wr.hint.attacks': 'چند رویداد هر عضو امتیاز دارد؛ ۰ یعنی بی‌حد',
    'wr.eligibility': 'شرط شرکت', 'wr.minMembers': 'حداقل عضو', 'wr.maxMembers': 'حداکثر عضو', 'wr.minLevel': 'حداقل سطح کلن', 'wr.minTrophies': 'حداقل جام کلن', 'wr.joined': 'ساعت عضویت',
    'wr.hint.joined': 'عضوی که کمتر از این مدت عضو است امتیاز نمی‌آورد',
    'wr.prizes': 'جایزه‌ها', 'wr.pWin': 'جایزهٔ برد', 'wr.pLose': 'جایزهٔ باخت', 'wr.pDraw': 'جایزهٔ مساوی', 'wr.pBye': 'جایزهٔ بدون حریف', 'wr.to': 'جایزه به چه کسی',
    'wr.to.participants': 'عضوهایی که شرکت کردند', 'wr.to.members': 'همهٔ عضوها', 'wr.byeResult': 'کلن بدون حریف', 'wr.hint.bye': 'وقتی تعداد کلن‌ها فرد است',
    'wr.bye.none': 'نتیجه‌ای نمی‌گیرد', 'wr.bye.win': 'برنده حساب شود',
    'rd.new': 'دور تازه', 'rd.create': 'ساختن دور', 'rd.created': 'دور ساخته شد', 'rd.hint': 'یک دور از روی قالب ساخته می‌شود. با «همین حالا شروع شود» کلن‌ها فوراً جفت می‌شوند.',
    'rd.startNow': 'همین حالا شروع شود', 'rd.hint.starts': 'تا این زمان ثبت‌نام باز است', 'rd.hint.duration': 'خالی یعنی مدت قالب', 'rd.needStart': 'زمان شروع را انتخاب کنید.',
    'rd.noConfigs': 'اول یک قالب جنگ بسازید.', 'rd.starts': 'شروع', 'rd.ends': 'پایان', 'rd.wars': 'جنگ', 'rd.signups': 'ثبت‌نام', 'rd.empty': 'دوری پیدا نشد.',
    'rd.st.signup': 'ثبت‌نام', 'rd.st.active': 'در جریان', 'rd.st.settled': 'تمام‌شده', 'rd.st.cancelled': 'لغوشده',
    'rd.detail': 'جزئیات دور', 'rd.start': 'شروع همین حالا', 'rd.settle': 'پایان و نتیجه', 'rd.cancel': 'لغو دور', 'rd.settleConfirm': 'دور همین حالا تمام شود و نتیجه‌ها و جایزه‌ها داده شوند؟',
    'rd.cancelConfirm': 'دور لغو شود؟ جنگ‌های در جریان بدون نتیجه بسته می‌شوند.', 'rd.done.start': 'دور شروع شد', 'rd.done.settle': 'دور تمام شد', 'rd.done.cancel': 'دور لغو شد',
    'rd.signupAt': 'باز شدن ثبت‌نام', 'rd.settledAt': 'بسته شد', 'rd.actor': 'به دست', 'rd.warsT': 'جنگ‌ها ({n})', 'rd.members': '{n} عضو', 'rd.wins': '{n} برد', 'rd.events': '{n} رویداد',
    'rd.noOpponent': 'بدون حریف', 'rd.winner': 'برنده: {name}', 'rd.notStarted': 'جنگ‌ها با شروع دور ساخته می‌شوند.', 'rd.noWars': 'جنگی ساخته نشد (کلن مجازی نبود).',
    'rd.signupsT': 'ثبت‌نام‌ها ({n})', 'rd.autoSignup': 'ثبت‌نام خودکار است؛ همهٔ کلن‌های مجاز در شروع می‌آیند.', 'rd.noSignups': 'هنوز کلنی ثبت‌نام نکرده است.', 'rd.score': 'امتیاز',
    'out.win': 'برد', 'out.lose': 'باخت', 'out.draw': 'مساوی', 'out.bye': 'بدون حریف', 'out.cancelled': 'بدون نتیجه',
    'cs.enabled': 'کلن روشن باشد', 'cs.hint.enabled': 'خاموش که باشد بازیکن‌ها نمی‌توانند کلن بسازند یا وارد شوند.',
    'cs.create': 'ساختن کلن', 'cs.createOn': 'بازیکن‌ها بتوانند کلن بسازند', 'cs.cost': 'هزینهٔ ساختن', 'cs.hint.cost': 'از کیف بازیکن کم می‌شود؛ بدون ردیف یعنی رایگان',
    'cs.currency': 'ارز', 'cs.amount': 'مقدار', 'cs.addCost': 'افزودن ارز', 'cs.minTrophies': 'حداقل جام سازنده', 'cs.hint.minTrophies': 'از افزونهٔ جام', 'cs.nameMin': 'کوتاه‌ترین نام', 'cs.nameMax': 'بلندترین نام',
    'cs.descMax': 'بلندترین توضیح', 'cs.defJoin': 'نوع عضویت پیش‌فرض', 'cs.maxMin': 'سقف «حداقل جام ورود»', 'cs.hint.maxMin': 'رهبر بیشتر از این نمی‌تواند بگذارد',
    'cs.badge': 'نشان کلن', 'cs.hint.badge': 'گزینه‌هایی که بازیکن برای نشان کلن دارد. خالی یعنی هر مقدار کوتاهی پذیرفته می‌شود.', 'cs.colors': 'رنگ‌ها',
    'cs.roles': 'اختیار نقش‌ها', 'cs.hint.roles': 'هر نقش چه کاری می‌تواند بکند. رهبر همیشه می‌تواند کلن را منتقل کند.', 'cs.action': 'کار',
    'cs.perm.accept': 'پذیرفتن درخواست', 'cs.perm.invite': 'دعوت کردن', 'cs.perm.kick': 'بیرون کردن', 'cs.perm.promote': 'بالا و پایین بردن نقش', 'cs.perm.edit': 'ویرایش پروفایل', 'cs.perm.war': 'ثبت‌نام در جنگ',
    'cs.permHint.accept': 'درخواست ورود به کلن', 'cs.permHint.invite': 'دعوت بازیکن به کلن', 'cs.permHint.kick': 'فقط نقش پایین‌تر', 'cs.permHint.promote': 'تا یک پله زیر خودش',
    'cs.permHint.edit': 'نام، توضیح، نشان، نوع عضویت', 'cs.permHint.war': 'وقتی ثبت‌نام جنگ با رهبر است',
    'cs.maxCo': 'حداکثر معاون', 'cs.leaderLeave': 'رفتن رهبر', 'cs.hint.leaderLeave': 'رهبر بخواهد برود چه شود', 'cs.leave.transfer': 'رهبری به عضو بعدی می‌رسد', 'cs.leave.block': 'اول باید رهبری را بدهد',
    'cs.cooldowns': 'زمان‌های انتظار', 'cs.joinAfter': 'ورود دوباره بعد از رفتن (دقیقه)', 'cs.hint.joinAfter': 'به هر کلنی', 'cs.rejoin': 'برگشت به کلن بعد از اخراج (ساعت)', 'cs.hint.rejoin': 'به همان کلن',
    'cs.leaveAfter': 'رفتن بعد از ورود (دقیقه)', 'cs.hint.leaveAfter': '۰ یعنی هر وقت', 'cs.invDays': 'اعتبار دعوت (روز)', 'cs.reqs': 'درخواست باز هر بازیکن', 'cs.hint.reqs': 'هم‌زمان به چند کلن',
    'cs.xp': 'امتیاز تجربه', 'cs.hint.sources': 'هر ردیف می‌گوید کدام رویدادِ عضوها برای کلن امتیاز دارد.', 'cs.context': 'زمینه', 'cs.xpAmount': 'امتیاز', 'cs.perValue': 'ضرب در مقدار فیلد',
    'cs.addSource': 'افزودن منبع امتیاز', 'cs.cap': 'سقف روزانهٔ هر عضو', 'cs.hint.cap': '۰ یعنی بی‌سقف', 'cs.warWin': 'امتیاز برد جنگ', 'cs.warDraw': 'امتیاز مساوی جنگ', 'cs.warLoss': 'امتیاز باخت جنگ',
    'cs.chat': 'گفتگو', 'cs.system': 'پیام سیستمی در گفتگوی کلن', 'cs.hint.system': 'مثل «فلانی وارد شد» یا «کلن سطح ۵ شد»',
    'cs.sLimit': 'نتیجهٔ جستجو', 'cs.hint.sLimit': 'چند کلن در هر جستجو', 'cs.hLimit': 'تاریخچه', 'cs.hint.hLimit': 'چند ردیف تاریخچه برای بازیکن',
    'cs.notify': 'به عضوها اعلان بده', 'cs.code': 'کد اعلان', 'cs.hint.code': 'بازی با این کد اعلان کلن را می‌شناسد',
    'code.TROPHY_LEAGUE_EXISTS': 'لیگی با همین کلید یا همین حداقل جام هست.', 'code.TROPHY_LEAGUE_NOT_FOUND': 'این لیگ پیدا نشد.',
    'code.CLAN_WAR_CONFIG_EXISTS': 'قالب جنگی با همین کلید هست.', 'code.CLAN_WAR_CONFIG_NOT_FOUND': 'این قالب جنگ پیدا نشد.', 'code.CLAN_WAR_ROUND_NOT_FOUND': 'این دور پیدا نشد.',
    'code.CLAN_WAR_ROUND_NOT_OPEN': 'این دور دیگر در ثبت‌نام نیست.', 'code.CLAN_WAR_ROUND_NOT_ACTIVE': 'این دور در جریان نیست.', 'code.CLAN_WAR_POINTS_REQUIRED': 'جنگ امتیازی دست‌کم یک قانون امتیاز لازم دارد.',
    'code.CLAN_LEVEL_ONE_REQUIRED': 'جاده باید با سطح ۱ و امتیاز صفر شروع شود.', 'code.CLAN_LEVEL_ORDER': 'سطح‌ها تکراری نباشند و امتیاز هر سطح از قبلی بیشتر باشد.',
    'code.CLAN_LEVELS_REQUIRED': 'دست‌کم یک سطح لازم است.', 'code.CLAN_IN_WAR': 'این کلن الان در جنگ است؛ بعد از پایان جنگ دوباره تلاش کنید.',
    'code.CLAN_LEADER_MUST_TRANSFER': 'رهبر را نمی‌شود پایین آورد؛ اول عضو دیگری را رهبر کنید.', 'code.CLAN_NOT_FOUND': 'این کلن پیدا نشد.', 'code.CLAN_MEMBER_NOT_FOUND': 'این بازیکن عضو این کلن نیست.',
    'code.CLAN_NAME_INVALID': 'نام کلن باید {min} تا {max} حرف باشد.', 'code.CLAN_NAME_TAKEN': 'این نام را کلن دیگری دارد.', 'code.CLAN_BADGE_INVALID': 'این نشان در فهرست نشان‌ها نیست.',
    'code.CLAN_DESCRIPTION_TOO_LONG': 'توضیح بلندتر از حد مجاز است.', 'code.CLAN_LANG_INVALID': 'کد زبان درست نیست.', 'code.CLAN_AVATAR_INVALID': 'نشانی تصویر درست نیست.',
    'code.CLAN_JOIN_TYPE_INVALID': 'نوع عضویت درست نیست.', 'code.CLAN_MIN_TROPHIES_INVALID': 'حداقل جام بیرون از بازهٔ مجاز است.', 'code.CLAN_TOO_MANY_CO_LEADERS': 'تعداد معاون‌ها به سقف رسیده است.',
    'code.IN_USE': 'در حال استفاده است؛ اول دورهای باز آن را تمام یا لغو کنید.', 'code.PRIZE_NOT_FOUND': 'این جایزه پیدا نشد.', 'code.INVALID_SCHEDULE': 'زمان‌بندی (cron) درست نیست.',
    'code.CONFIRM_REQUIRED': 'باید تأیید کنید.', 'code.VALUE_OUT_OF_RANGE': 'مقدار بیرون از بازهٔ مجاز است.',
    'code.INVALID_KEY': 'کلید یا نام درست نیست؛ با حرف یا عدد انگلیسی شروع شود و فقط حرف، عدد و _ . - داشته باشد.', 'code.INVALID_POLICY': 'یکی از گزینه‌ها درست نیست.',
    'code.INVALID_JSON': 'JSON درست نیست.', 'code.USER_NOT_FOUND': 'این بازیکن پیدا نشد.', 'code.INVALID_USER_ID': 'شناسهٔ بازیکن درست نیست.', 'code.NOT_FOUND': 'پیدا نشد.',
    'code.DEFAULT_LANGUAGE_REQUIRED': 'متن به زبان پیش‌فرض لازم است.', 'code.TEXT_TOO_LONG': 'متن بلندتر از حد مجاز است.'
  },
  en: {
    'group.trophies': 'Trophies', 'group.clans': 'Clans',
    'nav.trophies/leagues': 'Leagues', 'nav.trophies/players': 'Player trophies', 'nav.trophies/seasons': 'Seasons', 'nav.trophies/settings': 'Trophy settings',
    'nav.clans/list': 'Clans', 'nav.clans/levels': 'Clan levels', 'nav.clans/wars': 'Clan wars', 'nav.clans/rounds': 'War rounds', 'nav.clans/settings': 'Clan settings',
    'desc.trophies/leagues': 'Leagues are steps on the trophy road. A player sits in the highest league whose minimum they reach. Create leagues, set the prize for reaching one, and choose whether players can drop out of it.',
    'desc.trophies/players': 'Each player\'s trophies, best record and league. Open a row to see the trophy history and league prizes, and change the trophies if needed.',
    'desc.trophies/seasons': 'When a season ends, trophies go down by the rule in the settings. See past seasons and when the next one ends, or end the season now.',
    'desc.trophies/settings': 'Trophies come from match results. Choose which event is the match result, how many trophies a win or loss is worth, how league prizes are paid and when the season ends.',
    'desc.clans/list': 'All clans with their level, members, trophies and war results. Open a clan to edit its profile, change member roles, kick a member or give XP.',
    'desc.clans/levels': 'The clan level road: how much XP each level needs, how many members it holds and what prize it gives. Level 1 must start at 0 and each level needs more XP than the one before.',
    'desc.clans/wars': 'War templates. Points war: member actions earn points. Duel war: members of the two clans play each other and wins earn points. With a schedule, rounds are created automatically.',
    'desc.clans/rounds': 'A round is one run of a war template: signup, start, wars and results. Create a round, start it, end it early or cancel it.',
    'desc.clans/settings': 'Clan rules: creating a clan and its cost, badges, what each role may do, cooldowns, XP sources and events.',
    'save': 'Save', 'saved': 'Saved', 'deleted': 'Deleted', 'cancel': 'Cancel', 'close': 'Close', 'remove': 'Remove', 'delete': 'Delete', 'search': 'Search', 'all': 'All',
    'loading': 'Loading…', 'empty': 'Nothing found', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'yes': 'Yes', 'no': 'No',
    'active': 'Active', 'inactive': 'Inactive', 'on': 'On', 'off': 'Off', 'confirm.title': 'Are you sure?', 'lang.text': '{label} ({lang})', 'updated': 'Last change: {by} · {at}', 'none': '—',
    'missing': 'missing', 'noPrizes': 'The prize plugin is not loaded; prizes are recorded but not paid.',
    'f.name': 'Name', 'f.description': 'Description', 'f.key': 'Key', 'f.icon': 'Icon URL', 'f.color': 'Color', 'f.active': 'Active', 'f.state': 'Status', 'f.metadata': 'Extra data (JSON)',
    'f.player': 'Player', 'f.date': 'Time', 'f.updated': 'Last change', 'f.created': 'Created',
    'hint.key': 'Lowercase letters, digits and _ . - ; the game uses this key.', 'hint.json': 'A JSON object, like {"badge":"gold"}', 'hint.tags': 'Separate with commas', 'hint.zero': '0 means no limit',
    'src.economy_event': 'Economy event', 'src.game_event': 'Game event',
    'tr.unranked': 'No league',
    'lg.new': 'New league', 'lg.min': 'Minimum trophies', 'lg.from': 'From {n} trophies', 'lg.rules': 'Rules', 'lg.prize': 'Prize', 'lg.count': 'Players', 'lg.players': '{n} players have trophies', 'lg.unranked': '{n} players in no league',
    'lg.floorOn': 'No drop', 'lg.resetTo': 'Season end: {n}', 'lg.empty': 'No leagues yet. Every player is in "No league".', 'lg.delete': 'Delete league {key}? Its players move to the league below.',
    'lg.hint': 'The name and icon are what players see in the game.', 'lg.floor': 'Players cannot drop out of this league', 'lg.hint.floor': 'Once a player reaches this league, losses never take them below its minimum.',
    'lg.reset': 'Trophies at season end', 'lg.hint.reset': 'Only when the season rule is "Back to each league\'s number". Empty means no change.', 'lg.hint.min': 'Two leagues cannot share a minimum.',
    'lg.hint.prize': 'Given the first time a player reaches this league.',
    'tp.search': 'Player', 'tp.hint.search': 'Username or full ID', 'tp.league': 'League', 'tp.trophies': 'Trophies', 'tp.best': 'Best', 'tp.bestV': 'All time {b} · this season {s}', 'tp.record': 'Wins / losses / draws',
    'tp.wld': '{w} / {l} / {d}', 'tp.empty': 'No players with trophies found.', 'tp.open': 'Open this player', 'tp.retry': 'Retry failed prizes', 'tp.retried': '{n} prizes queued again',
    'tp.change': 'Change trophies', 'tp.hint.change': 'The change is logged under your name and the player\'s league is worked out again.', 'tp.amount': 'Amount', 'tp.modeL': 'Change', 'tp.mode.set': 'Set to', 'tp.mode.add': 'Add or subtract',
    'tp.apply': 'Apply', 'tp.log': 'Trophy history', 'tp.noLog': 'No changes yet.', 'tp.reason': 'Reason', 'tp.delta': 'Change', 'tp.after': 'Before → after', 'tp.opponent': 'Opponent',
    'tp.reason.win': 'Win', 'tp.reason.loss': 'Loss', 'tp.reason.draw': 'Draw', 'tp.reason.admin': 'Admin', 'tp.reason.season': 'Season end', 'tp.move': '{a} → {b}',
    'tp.prizes': 'League prizes', 'tp.noPrizes': 'No league prizes yet.', 'tp.season': 'Season', 'tp.ever': 'Once ever',
    'pay.pending': 'Queued', 'pay.paid': 'Paid', 'pay.failed': 'Failed',
    'ss.reset': 'End season now', 'ss.confirm': 'End the current season now? Trophies go down by the "{mode}" rule and this cannot be undone.', 'ss.done': 'Season ended',
    'ss.hint.reset': 'A new season starts right away.', 'ss.auto': 'Automatic end', 'ss.schedule': 'Schedule', 'ss.modeL': 'Season end rule', 'ss.next': 'Next end', 'ss.manual': 'Manual only',
    'ss.mode.cut': 'Cut trophies above a number', 'ss.mode.league': 'Back to each league\'s number', 'ss.mode.zero': 'Everyone back to zero',
    'ss.number': 'Season', 'ss.n': 'Season {n}', 'ss.started': 'Started', 'ss.ends': 'Planned end', 'ss.ended': 'Ended', 'ss.current': 'Current season', 'ss.players': 'Players cut', 'ss.actor': 'By',
    'ss.scheduler': 'Scheduler', 'ss.empty': 'No seasons yet.',
    'ts.general': 'General', 'ts.enabled': 'Trophies are on', 'ts.hint.enabled': 'When off, match results give and take no trophies.',
    'ts.match': 'Match result', 'ts.hint.match': 'Trophies come from the event the game sends after each match. Type the field names of that event.',
    'ts.source': 'Event type', 'ts.event': 'Event name', 'ts.hint.event': 'Empty means no event gives trophies.', 'ts.won': 'Win field', 'ts.hint.won': 'true means the player won.',
    'ts.draw': 'Draw field', 'ts.hint.draw': 'Optional', 'ts.matchField': 'Match ID field', 'ts.hint.matchField': 'A match is never counted twice.', 'ts.opponent': 'Opponent field',
    'ts.hint.opponent': 'The opponent\'s ID, for trophies by trophy gap.', 'ts.modeField': 'Game mode field', 'ts.hint.modeField': 'Optional', 'ts.modes': 'Modes that give trophies',
    'ts.hint.modes': 'Empty means every mode', 'ts.skip': 'Skip field', 'ts.hint.skip': 'If this field is true the match gives no trophies (a friendly, for example).',
    'ts.amounts': 'Trophy amounts', 'ts.start': 'Starting trophies', 'ts.hint.start': 'For a new player', 'ts.win': 'Trophies for a win', 'ts.loss': 'Trophies for a loss', 'ts.drawA': 'Trophies for a draw', 'ts.hint.drawA': 'Can be negative',
    'ts.scale': 'Scale by the trophy gap to the opponent', 'ts.hint.scale': 'Beating a stronger opponent gives more; losing to a weaker one takes more.',
    'ts.scalePer': 'For every N trophies of gap', 'ts.scaleStep': 'Change by', 'ts.minWin': 'Fewest for a win', 'ts.maxWin': 'Most for a win', 'ts.minLoss': 'Fewest for a loss', 'ts.maxLoss': 'Most for a loss',
    'ts.floors': 'Respect league floors', 'ts.hint.floors': 'In a "no drop" league, trophies never go below its minimum.',
    'ts.prizes': 'League prizes', 'ts.oncePer': 'How often', 'ts.hint.oncePer': 'Once ever or once each season', 'ts.once.ever': 'Once ever', 'ts.once.season': 'Once each season',
    'ts.onlyRising': 'Only when the player climbs into the league (not by an admin change)', 'ts.tries': 'Tries', 'ts.hint.tries': 'How many times a failed prize payment is retried',
    'ts.season': 'Season', 'ts.seasonOn': 'End seasons automatically', 'ts.schedule': 'Schedule (cron)', 'ts.hint.schedule': 'For example 0 0 1 * * is the first of every month', 'ts.hint.sMode': 'How trophies go down at season end',
    'ts.above': 'Above', 'ts.hint.above': 'Only trophies above this number are cut', 'ts.keep': 'Percent kept', 'ts.hint.keep': 'Of the trophies above that number',
    'ts.mm': 'Matchmaking', 'ts.mmOn': 'Match players with similar trophies', 'ts.hint.mm': 'The player\'s trophies are sent to the matchmaker.', 'ts.games': 'Games', 'ts.hint.games': 'None picked means every game',
    'ts.range': 'Trophy range', 'ts.hint.range': 'The opponent has at most this many more or fewer trophies', 'ts.property': 'Matchmaker property', 'ts.hint.property': 'The name of the number sent to the matchmaker',
    'ts.display': 'Display', 'ts.hist': 'History rows for the player', 'ts.hint.hist': 'How many latest rows the player sees', 'ts.logDays': 'Keep history (days)', 'ts.hint.logDays': 'Older rows are removed',
    'ts.showOpponent': 'Show the opponent in the history',
    'ts.events': 'Events and notifications', 'ts.publish': 'Send changes to the event system', 'ts.notify': 'Notify the player', 'ts.nCode': 'Trophy change code', 'ts.hint.nCode': 'The game knows the notification by this code',
    'ts.lCode': 'League change code', 'ts.hint.lCode': 'When the player\'s league changes',
    'jt.open': 'Open', 'jt.closed': 'Request to join', 'jt.invite': 'Invite only',
    'role.leader': 'Leader', 'role.co_leader': 'Co-leader', 'role.elder': 'Elder', 'role.member': 'Member',
    'cl.search': 'Clan', 'cl.hint.search': 'Name, tag or ID', 'cl.joinType': 'Joining', 'cl.sortL': 'Sort by', 'cl.sort.trophies': 'Most trophies', 'cl.sort.level': 'Highest level',
    'cl.sort.members': 'Most members', 'cl.sort.new': 'Newest', 'cl.sort.wars': 'Most war wins', 'cl.clan': 'Clan', 'cl.level': 'Level', 'cl.lv': 'Level {n}', 'cl.members': 'Members', 'cl.of': '{n} of {m}',
    'cl.wars': 'Wars (W / L / D)', 'cl.empty': 'No clans found.', 'cl.xpOf': '{x} of {n} XP', 'cl.xpMax': '{x} XP · top level',
    'cl.tag': 'Tag', 'cl.warStats': '{w} won · {l} lost · {d} drawn · {p} points', 'cl.requests': 'Pending requests',
    'cl.profile': 'Profile', 'cl.lang': 'Language', 'cl.avatar': 'Image URL', 'cl.emblem': 'Emblem', 'cl.frame': 'Frame', 'cl.minTrophies': 'Trophies needed to join', 'cl.hint.minTrophies': 'Players with fewer cannot join',
    'cl.xp': 'Experience (XP)', 'cl.hint.xp': 'The XP is logged under your name in the clan history and may level the clan up.', 'cl.xpAmount': 'XP amount', 'cl.addXp': 'Add XP', 'cl.xpAdded': '{n} XP added',
    'cl.membersT': 'Members ({n})', 'cl.noMembers': 'No members.', 'cl.memberLine': '{x} XP · {p} war points · {t} trophies', 'cl.role': 'Role', 'cl.setRole': 'Change role', 'cl.kick': 'Kick',
    'cl.kickConfirm': 'Kick {name} from the clan? They cannot come back for a while.', 'cl.kicked': 'Kicked', 'cl.leaderConfirm': 'Make {name} the leader? The current leader becomes a co-leader.',
    'cl.xpLog': 'XP history', 'cl.noXp': 'No XP yet.', 'cl.source': 'Source', 'cl.amount': 'Amount', 'cl.warsT': 'Recent wars', 'cl.noWars': 'Has not been in a war yet.',
    'cl.delete': 'Delete clan', 'cl.deleteConfirm': 'Delete the clan "{name}"? Every member is removed and this cannot be undone.',
    'xs.admin': 'Admin', 'xs.war_win': 'War win', 'xs.war_draw': 'War draw', 'xs.war_lose': 'War loss', 'xs.war_bye': 'War without opponent',
    'lv.hint': 'Changing the road affects every clan: each clan\'s level and member cap are worked out again.', 'lv.level': 'Level', 'lv.xp': 'XP needed', 'lv.hint.xp': 'Total XP from the start', 'lv.cap': 'Member cap',
    'lv.prize': 'Prize for reaching', 'lv.perks': 'Perks (JSON)', 'lv.clans': '{n} clans at this level', 'lv.new': 'New level', 'lv.add': 'Add level',
    'mode.points': 'Points', 'mode.duel': 'Duel',
    'wr.new': 'New war template', 'wr.mode': 'War type', 'wr.schedule': 'Schedule', 'wr.manual': 'Manual', 'wr.duration': 'Length (minutes)', 'wr.minutes': '{n} min', 'wr.signup': 'Signup',
    'wr.signup.auto': 'Automatic', 'wr.signup.leader': 'By the leader', 'wr.next': 'Next start', 'wr.empty': 'No war templates yet.', 'wr.delete': 'Delete war template {key}?',
    'wr.config': 'War template', 'wr.hint': 'Each round keeps its own copy of these rules; changing the template does not touch running rounds.',
    'wr.hint.mode': 'Points: member actions earn points. Duel: members play members of the other clan.',
    'wr.timing': 'Timing', 'wr.hint.schedule': 'cron; empty means manual only. Example: 0 18 * * 5 is every Friday at 18:00', 'wr.hint.duration': 'From start to end of the war', 'wr.hint.signup': 'Automatic: every clan that qualifies. By the leader: the leader signs up.',
    'wr.signupMinutes': 'Signup window (minutes)', 'wr.hint.signupMinutes': 'How long before the start signup opens', 'wr.pairing': 'Pairing', 'wr.pair.random': 'Random', 'wr.pair.level': 'Similar level', 'wr.pair.trophies': 'Similar trophies',
    'wr.pointsRules': 'Point rules', 'wr.hint.points': 'Each rule says which event is worth how many points for the clan. Use a field and value to count only some events.',
    'wr.source': 'Event type', 'wr.event': 'Event', 'wr.field': 'Field', 'wr.equals': 'Equals', 'wr.points': 'Points', 'wr.perValue': 'Points times the field value', 'wr.addPoint': 'Add rule',
    'wr.duelRules': 'Duel rules', 'wr.hint.duel': 'The match result event must name the opponent; only matches against a member of the other clan count.', 'wr.won': 'Win field', 'wr.drawF': 'Draw field',
    'wr.opp': 'Opponent field', 'wr.hint.opp': 'The opponent player\'s ID', 'wr.winPoints': 'Points for a win', 'wr.drawPoints': 'Points for a draw',
    'wr.scoring': 'Counting', 'wr.matchField': 'Match ID field', 'wr.hint.matchField': 'A match is never counted twice', 'wr.attacks': 'Attacks per member', 'wr.hint.attacks': 'How many events per member score; 0 means no limit',
    'wr.eligibility': 'Who can take part', 'wr.minMembers': 'Fewest members', 'wr.maxMembers': 'Most members', 'wr.minLevel': 'Lowest clan level', 'wr.minTrophies': 'Fewest clan trophies', 'wr.joined': 'Hours in the clan',
    'wr.hint.joined': 'Members who joined more recently do not score',
    'wr.prizes': 'Prizes', 'wr.pWin': 'Win prize', 'wr.pLose': 'Loss prize', 'wr.pDraw': 'Draw prize', 'wr.pBye': 'No-opponent prize', 'wr.to': 'Prize goes to',
    'wr.to.participants': 'Members who took part', 'wr.to.members': 'Every member', 'wr.byeResult': 'Clan without an opponent', 'wr.hint.bye': 'When the number of clans is odd',
    'wr.bye.none': 'Gets no result', 'wr.bye.win': 'Counts as a win',
    'rd.new': 'New round', 'rd.create': 'Create round', 'rd.created': 'Round created', 'rd.hint': 'A round is made from a template. With "Start now" clans are paired right away.',
    'rd.startNow': 'Start now', 'rd.hint.starts': 'Signup stays open until then', 'rd.hint.duration': 'Empty means the template length', 'rd.needStart': 'Pick a start time.',
    'rd.noConfigs': 'Create a war template first.', 'rd.starts': 'Starts', 'rd.ends': 'Ends', 'rd.wars': 'Wars', 'rd.signups': 'Signups', 'rd.empty': 'No rounds found.',
    'rd.st.signup': 'Signup', 'rd.st.active': 'Running', 'rd.st.settled': 'Finished', 'rd.st.cancelled': 'Cancelled',
    'rd.detail': 'Round details', 'rd.start': 'Start now', 'rd.settle': 'End and score', 'rd.cancel': 'Cancel round', 'rd.settleConfirm': 'End the round now and give out results and prizes?',
    'rd.cancelConfirm': 'Cancel the round? Running wars close with no result.', 'rd.done.start': 'Round started', 'rd.done.settle': 'Round ended', 'rd.done.cancel': 'Round cancelled',
    'rd.signupAt': 'Signup opens', 'rd.settledAt': 'Closed', 'rd.actor': 'By', 'rd.warsT': 'Wars ({n})', 'rd.members': '{n} members', 'rd.wins': '{n} wins', 'rd.events': '{n} events',
    'rd.noOpponent': 'No opponent', 'rd.winner': 'Winner: {name}', 'rd.notStarted': 'Wars are made when the round starts.', 'rd.noWars': 'No wars were made (no clan qualified).',
    'rd.signupsT': 'Signups ({n})', 'rd.autoSignup': 'Signup is automatic; every clan that qualifies joins at the start.', 'rd.noSignups': 'No clan has signed up yet.', 'rd.score': 'Score',
    'out.win': 'Won', 'out.lose': 'Lost', 'out.draw': 'Draw', 'out.bye': 'No opponent', 'out.cancelled': 'No result',
    'cs.enabled': 'Clans are on', 'cs.hint.enabled': 'When off, players cannot create or join clans.',
    'cs.create': 'Creating a clan', 'cs.createOn': 'Players can create clans', 'cs.cost': 'Creation cost', 'cs.hint.cost': 'Taken from the player\'s wallet; no rows means free',
    'cs.currency': 'Currency', 'cs.amount': 'Amount', 'cs.addCost': 'Add currency', 'cs.minTrophies': 'Creator\'s minimum trophies', 'cs.hint.minTrophies': 'From the trophy plugin', 'cs.nameMin': 'Shortest name', 'cs.nameMax': 'Longest name',
    'cs.descMax': 'Longest description', 'cs.defJoin': 'Default joining', 'cs.maxMin': 'Highest "trophies to join"', 'cs.hint.maxMin': 'A leader cannot set more than this',
    'cs.badge': 'Clan badge', 'cs.hint.badge': 'The choices players have for the clan badge. Empty means any short value is accepted.', 'cs.colors': 'Colors',
    'cs.roles': 'Role permissions', 'cs.hint.roles': 'What each role may do. The leader can always hand over the clan.', 'cs.action': 'Action',
    'cs.perm.accept': 'Accept requests', 'cs.perm.invite': 'Invite', 'cs.perm.kick': 'Kick', 'cs.perm.promote': 'Promote and demote', 'cs.perm.edit': 'Edit profile', 'cs.perm.war': 'Sign up for wars',
    'cs.permHint.accept': 'Requests to join the clan', 'cs.permHint.invite': 'Invite a player to the clan', 'cs.permHint.kick': 'Lower roles only', 'cs.permHint.promote': 'Up to one step below themselves',
    'cs.permHint.edit': 'Name, description, badge, joining', 'cs.permHint.war': 'When war signup is by the leader',
    'cs.maxCo': 'Most co-leaders', 'cs.leaderLeave': 'Leader leaving', 'cs.hint.leaderLeave': 'What happens when the leader leaves', 'cs.leave.transfer': 'Leadership passes to the next member', 'cs.leave.block': 'Must hand over first',
    'cs.cooldowns': 'Cooldowns', 'cs.joinAfter': 'Join again after leaving (minutes)', 'cs.hint.joinAfter': 'Any clan', 'cs.rejoin': 'Back after a kick (hours)', 'cs.hint.rejoin': 'The same clan',
    'cs.leaveAfter': 'Leave after joining (minutes)', 'cs.hint.leaveAfter': '0 means any time', 'cs.invDays': 'Invite valid for (days)', 'cs.reqs': 'Open requests per player', 'cs.hint.reqs': 'To how many clans at once',
    'cs.xp': 'Experience (XP)', 'cs.hint.sources': 'Each row says which member event earns XP for the clan.', 'cs.context': 'Context', 'cs.xpAmount': 'XP', 'cs.perValue': 'Times the field value',
    'cs.addSource': 'Add XP source', 'cs.cap': 'Daily cap per member', 'cs.hint.cap': '0 means no cap', 'cs.warWin': 'XP for a war win', 'cs.warDraw': 'XP for a war draw', 'cs.warLoss': 'XP for a war loss',
    'cs.chat': 'Chat', 'cs.system': 'System messages in the clan chat', 'cs.hint.system': 'Like "Sara joined" or "The clan reached level 5"',
    'cs.sLimit': 'Search results', 'cs.hint.sLimit': 'Clans per search', 'cs.hLimit': 'History', 'cs.hint.hLimit': 'History rows for the player',
    'cs.notify': 'Notify members', 'cs.code': 'Notification code', 'cs.hint.code': 'The game knows clan notifications by this code',
    'code.TROPHY_LEAGUE_EXISTS': 'A league with this key or this minimum already exists.', 'code.TROPHY_LEAGUE_NOT_FOUND': 'League not found.',
    'code.CLAN_WAR_CONFIG_EXISTS': 'A war template with this key already exists.', 'code.CLAN_WAR_CONFIG_NOT_FOUND': 'War template not found.', 'code.CLAN_WAR_ROUND_NOT_FOUND': 'Round not found.',
    'code.CLAN_WAR_ROUND_NOT_OPEN': 'This round is no longer in signup.', 'code.CLAN_WAR_ROUND_NOT_ACTIVE': 'This round is not running.', 'code.CLAN_WAR_POINTS_REQUIRED': 'A points war needs at least one point rule.',
    'code.CLAN_LEVEL_ONE_REQUIRED': 'The road must start with level 1 at 0 XP.', 'code.CLAN_LEVEL_ORDER': 'Levels must not repeat and each level needs more XP than the one before.',
    'code.CLAN_LEVELS_REQUIRED': 'At least one level is needed.', 'code.CLAN_IN_WAR': 'This clan is in a war right now; try again after it ends.',
    'code.CLAN_LEADER_MUST_TRANSFER': 'The leader cannot be demoted; make another member the leader first.', 'code.CLAN_NOT_FOUND': 'Clan not found.', 'code.CLAN_MEMBER_NOT_FOUND': 'This player is not a member of this clan.',
    'code.CLAN_NAME_INVALID': 'The clan name must be {min} to {max} characters.', 'code.CLAN_NAME_TAKEN': 'Another clan has this name.', 'code.CLAN_BADGE_INVALID': 'This badge is not in the badge lists.',
    'code.CLAN_DESCRIPTION_TOO_LONG': 'The description is too long.', 'code.CLAN_LANG_INVALID': 'The language code is not valid.', 'code.CLAN_AVATAR_INVALID': 'The image URL is not valid.',
    'code.CLAN_JOIN_TYPE_INVALID': 'The joining type is not valid.', 'code.CLAN_MIN_TROPHIES_INVALID': 'Trophies to join is out of range.', 'code.CLAN_TOO_MANY_CO_LEADERS': 'The clan already has the most co-leaders allowed.',
    'code.IN_USE': 'In use; finish or cancel its open rounds first.', 'code.PRIZE_NOT_FOUND': 'Prize not found.', 'code.INVALID_SCHEDULE': 'The schedule (cron) is not valid.',
    'code.CONFIRM_REQUIRED': 'Confirmation is required.', 'code.VALUE_OUT_OF_RANGE': 'The value is out of range.',
    'code.INVALID_KEY': 'The key or name is not valid; start with a letter or digit and use only letters, digits and _ . -', 'code.INVALID_POLICY': 'One of the choices is not valid.',
    'code.INVALID_JSON': 'The JSON is not valid.', 'code.USER_NOT_FOUND': 'Player not found.', 'code.INVALID_USER_ID': 'The player ID is not valid.', 'code.NOT_FOUND': 'Not found.',
    'code.DEFAULT_LANGUAGE_REQUIRED': 'Text in the default language is required.', 'code.TEXT_TOO_LONG': 'The text is too long.'
  }
};

function clLang() {
  var nc = window.NakamaConsole;
  var l = nc && nc.lang ? nc.lang() : (document.documentElement.getAttribute('lang') || 'fa');
  return /^en/i.test(l) ? 'en' : 'fa';
}

function has(key) {
  return (CL_TEXT[clLang()] || CL_TEXT.fa)[key] != null;
}

function t(key, vars) {
  var dict = CL_TEXT[clLang()] || CL_TEXT.fa;
  var s = dict[key] != null ? dict[key] : (CL_TEXT.en[key] != null ? CL_TEXT.en[key] : key);
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] != null ? digits(String(vars[k])) : m; });
  return s;
}

function digits(s) {
  if (clLang() !== 'fa') return s;
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
    return d.toLocaleString(clLang() === 'fa' ? 'fa-IR' : 'en-GB', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

function tr(key, vars, raw) {
  var s = t(key, vars);
  Object.keys(raw || {}).forEach(function (k) { s = s.split('{' + k + '}').join(String(raw[k])); });
  return s;
}

var CL_API = '/v2/console/base';
var CL_INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 px-3 py-2 text-sm';
var CL_BTN = 'cl-btn inline-flex items-center justify-center gap-1.5 font-medium whitespace-nowrap rounded-md border text-sm py-1.5 px-3';

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

function ClError(status, code, message, field, meta) {
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
  return fetch((path.indexOf('/v2/') === 0 ? '' : CL_API) + path, init).then(function (res) {
    if (res.status === 401 && !/^#\/login/.test(location.hash)) location.hash = '#/login';
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) throw new ClError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', data && data.field || '', data && data.meta);
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
  var b = h('button', { type: opts.type || 'button', class: CL_BTN + ' cl-btn--' + (opts.variant || 'ghost') + (opts.class ? ' ' + opts.class : ''), disabled: !!opts.disabled, title: opts.title, 'aria-label': opts.title, onclick: opts.onclick },
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

var clToastBox = null;
function toast(kind, text) {
  if (!clToastBox || !clToastBox.isConnected) {
    clToastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-clan': 'toasts' });
    document.body.appendChild(clToastBox);
  }
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  clToastBox.appendChild(el);
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
  var err = h('p', { class: 'cl-field-error text-xs text-error-600 dark:text-error-400', hidden: true });
  var wrap = h('label', { class: 'cl-field block mb-4' + (extra ? ' ' + extra : '') },
    label ? h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1', text: label }) : null,
    control,
    hint ? h('span', { class: 'block text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: hint }) : null,
    err);
  wrap.setError = function (msg) { err.textContent = msg || ''; err.hidden = !msg; };
  return wrap;
}

function input(value, opts) {
  opts = opts || {};
  var el = h(opts.multiline ? 'textarea' : 'input', { class: CL_INPUT + (opts.mono ? ' font-mono' : ''), type: opts.multiline ? null : (opts.type || 'text'), dir: opts.ltr ? 'ltr' : 'auto', rows: opts.multiline ? (opts.rows || 3) : null, placeholder: opts.placeholder, name: opts.name, autocomplete: 'off', 'aria-label': opts.label });
  el.value = value == null ? '' : value;
  return el;
}

function select(value, options, opts) {
  opts = opts || {};
  var el = h('select', { class: CL_INPUT + ' eco-select', name: opts.name, 'aria-label': opts.label });
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) opt.selected = true;
    el.appendChild(opt);
  });
  return el;
}

function check(checked, label, name) {
  var box = h('input', { type: 'checkbox', class: 'cl-check', name: name, checked: !!checked });
  var wrap = h('label', { class: 'cl-checkline flex items-start gap-2 mb-3 cursor-pointer text-sm text-secondary-800 dark:text-secondary-200' }, box, h('span', { text: label }));
  wrap.input = box;
  return wrap;
}

function numberValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? Math.round(n) : NaN;
}

var clPanel = null;
function closePanel() {
  if (clPanel) { clPanel.remove(); clPanel = null; document.removeEventListener('keydown', panelKey); }
}

function panelKey(e) { if (e.key === 'Escape') closePanel(); }

function sidePanel(title, body, footer, name) {
  closePanel();
  var holder = h('div', { class: 'cl-panel-holder' });
  var panel = h('div', { class: 'cl-panel bg-white dark:bg-secondary-900 shadow-lg', role: 'dialog', 'aria-modal': 'true', 'aria-label': title, 'data-clan-panel': name || '' },
    h('div', { class: 'cl-panel-head flex items-center justify-between border-b px-6 py-3' },
      h('h2', { class: 'text-lg font-semibold truncate', text: title }),
      button({ label: '×', title: t('close'), onclick: closePanel, class: 'cl-close' })),
    h('div', { class: 'cl-panel-body p-6' }, body),
    footer ? h('div', { class: 'cl-panel-foot flex gap-2 justify-end bg-secondary-50 dark:bg-secondary-800 p-4' }, footer) : null);
  holder.appendChild(panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) closePanel(); });
  clPanel = h('div', { 'data-clan': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, h('div', { class: 'cl-shade' }), holder);
  document.body.appendChild(clPanel);
  document.addEventListener('keydown', panelKey);
  return panel;
}

function confirmBox(text, okLabel) {
  return new Promise(function (resolve) {
    var root;
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-clan': 'confirm', 'data-l10n-skip': '', class: 'cl-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'cl-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('confirm.title') }),
        h('p', { class: 'text-sm mb-4', text: text }),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: okLabel || t('yes'), variant: 'danger', data: { 'clan-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root);
  });
}

function pageFrame(page, desc, actions) {
  var content = h('div', { class: 'flex flex-col gap-4' });
  var el = h('div', { class: 'cl-page flex flex-col flex-1 px-4 md:px-6 pb-6 max-w-[1600px] w-full mx-auto', 'data-clan-page': page },
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
    var tr = h('tr', { class: 'border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' + (onrow ? ' cursor-pointer' : ''), 'data-clan-row': row.id || '' });
    if (onrow) {
      tr.tabIndex = 0;
      tr.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,select,label')) onrow(row.item); });
      tr.addEventListener('keydown', function (e) { if (e.key === 'Enter') onrow(row.item); });
    }
    row.cells.forEach(function (cell, i) { tr.appendChild(h('td', { class: 'px-3 py-2 align-middle text-sm', 'data-label': columns[i].label }, cell)); });
    tbody.appendChild(tr);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow rounded-md overflow-x-auto' }, h('table', { class: 'w-full eco-table cl-table' }, thead, tbody));
}

function pager(state, onpage) {
  var total = state.total || 0;
  var from = total ? (state.page - 1) * state.limit + 1 : 0;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  return h('div', { class: 'flex items-center justify-end gap-2 text-xs text-secondary-500' },
    h('span', { 'data-clan-range': '', text: total ? t('range', { from: from, to: to, total: total }) : t('empty') }),
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

var clLanguages = null;
function loadLanguages() {
  if (clLanguages) return Promise.resolve(clLanguages);
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch('/v2/console/base/languages', { headers: headers, credentials: 'same-origin' }).then(function (r) { return r.ok ? r.json() : null; }).then(function (d) {
    var list = d && d.items ? d.items.filter(function (l) { return l.isActive; }).map(function (l) { return { code: l.code, name: l.nativeName || l.code, isDefault: !!l.isDefault, direction: l.direction }; }) : [];
    if (!list.length) list = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    list.sort(function (a, b) { return (b.isDefault ? 1 : 0) - (a.isDefault ? 1 : 0); });
    clLanguages = list;
    return list;
  }, function () {
    clLanguages = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    return clLanguages;
  });
}

function translated(label, name, value, opts) {
  opts = opts || {};
  var inputs = {};
  var box = h('div', { class: 'grid sm:grid-cols-2 gap-x-4', 'data-clan-text': name });
  (clLanguages || []).forEach(function (l) {
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
  var l = clLang();
  return m[l] || m.fa || m.en || Object.keys(m).map(function (k) { return m[k]; }).filter(Boolean)[0] || fallback || '';
}

function fieldError(msgBox, err) {
  clear(msgBox).appendChild(notice('error', clErrText(err)));
}

function options(list, value) {
  return list.map(function (o) { return typeof o === 'string' ? { value: o, label: o } : o; });
}

function dateInput(value, name, type) {
  var el = h('input', { type: type || 'date', class: CL_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  el.value = value || '';
  return el;
}

function isoDay(d) {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

function clErrText(err) {
  if (err && err.code && has('code.' + err.code)) return t('code.' + err.code, err.meta || null) + (err.field ? ' · ' + err.field : '');
  return errorText(err) + (err && err.field ? ' · ' + err.field : '');
}

function section(title, children, name) {
  return card(h('div', { class: 'p-4', 'data-clan-section': name || '' },
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
    var msg = h('div', { 'data-clan-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', 'data-clan-updated': '' });
    var paintMeta = function (x) { meta.textContent = x.updatedBy ? t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }) : ''; };
    paintMeta(d);
    var save = button({ label: t('save'), variant: 'primary', data: { 'clan-save': page }, onclick: function () {
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
        msg.appendChild(notice('error', clErrText(err)));
      });
    } });
    clear(frame.content);
    form.el.forEach(function (el) { frame.content.appendChild(el); });
    frame.content.appendChild(h('div', { class: 'cl-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), save));
  }, function (err) {
    clear(frame.content).appendChild(notice('error', clErrText(err)));
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
  return h('a', { href: '#/players/' + id, class: 'eco-link', 'data-clan-player-link': id, dir: 'auto', text: name || id.slice(0, 8) });
}

var clRefs = {};
function loadRefs(kind, force) {
  if (clRefs[kind] && !force) return Promise.resolve(clRefs[kind]);
  return api('GET', '/' + kind + '/refs' + qs({ lang: clLang() })).then(function (r) { clRefs[kind] = r || {}; return clRefs[kind]; }, function () { clRefs[kind] = {}; return clRefs[kind]; });
}

function refList(kind, list) {
  return ((clRefs[kind] && clRefs[kind][list]) || []).map(function (x) { return typeof x === 'string' ? { key: x, name: x } : x; });
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

var clListSeq = 0;
function eventInput(items, value, name, placeholder) {
  var el = input(value || '', { name: name, ltr: true, mono: true, placeholder: placeholder });
  var id = 'cl-list-' + (++clListSeq);
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
  var dot = h('span', { class: 'cl-dot' });
  if (color) dot.style.background = color;
  return dot;
}

function nameCell(title, key, icon, color) {
  return h('div', { class: 'flex items-center gap-2 min-w-0' }, icon ? h('img', { src: icon, alt: '', class: 'cl-icon' }) : colorDot(color),
    h('div', { class: 'flex flex-col min-w-0' }, h('span', { class: 'font-medium', dir: 'auto', text: title }), key ? mono(key) : null));
}

function dl(pairs, name) {
  var box = h('dl', { class: 'cl-dl text-sm', 'data-clan-dl': name || '' });
  pairs.forEach(function (p) {
    if (!p) return;
    box.appendChild(h('dt', { text: p[0] }));
    box.appendChild(h('dd', {}, p[1] == null || p[1] === '' ? t('none') : p[1]));
  });
  return box;
}

function deltaText(n) {
  var s = (n > 0 ? '+' : n < 0 ? '−' : '') + fmtNum(Math.abs(n));
  return h('span', { class: n > 0 ? 'cl-amount-in' : n < 0 ? 'cl-amount-out' : '', dir: 'ltr', text: s });
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
  var picker = h('input', { type: 'color', class: 'cl-color', name: 'color-picker', 'aria-label': t('f.color'), value: /^#[0-9A-Fa-f]{6}$/.test(value || '') ? value : '#f5b301' });
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
  var el = h('input', { type: 'datetime-local', class: CL_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  if (v) {
    var d = new Date(v);
    if (!isNaN(d)) el.value = new Date(d.getTime() - d.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
  }
  el.read = function () { if (!el.value) return null; var d = new Date(el.value); return isNaN(d) ? null : d.toISOString(); };
  return el;
}

var clLeagues = null;
function loadLeagues(force) {
  if (clLeagues && !force) return Promise.resolve(clLeagues);
  return api('GET', '/trophies/leagues').then(function (d) { clLeagues = d; return d; });
}

function leagueName(key) {
  if (!key) return t('tr.unranked');
  var l = clLeagues && (clLeagues.items || []).filter(function (x) { return x.key === key; })[0];
  return l ? textOf(l.name, l.key) : key;
}

function leaguesPage(root) {
  var frame = pageFrame('trophies/leagues', t('desc.trophies/leagues'), [
    button({ label: t('lg.new'), variant: 'primary', data: { 'clan-action': 'new-league' }, onclick: function () { leagueEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  var box = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(box);
  function reload() {
    clear(box).appendChild(loading());
    Promise.all([loadLeagues(true), loadLanguages(), loadRefs('trophies')]).then(function (res) {
      var d = res[0];
      var items = d.items || [];
      clear(box);
      box.appendChild(h('div', { class: 'flex flex-wrap gap-2', 'data-clan-summary': 'leagues' },
        tag(t('lg.players', { n: fmtNum(d.players || 0) }), 'primary'), tag(t('lg.unranked', { n: fmtNum(d.unranked || 0) }), 'secondary')));
      box.appendChild(h('div', { 'data-clan-list': 'leagues' }, table([
        { label: t('f.name') }, { label: t('lg.min') }, { label: t('lg.rules') }, { label: t('lg.prize') }, { label: t('lg.count') }, { label: t('f.state') }, { label: '' }
      ], items.map(function (l) {
        return { id: l.key, item: l, cells: [
          nameCell(textOf(l.name, l.key), l.key, l.iconUrl, l.color),
          h('span', { text: t('lg.from', { n: fmtNum(l.minTrophies) }) }),
          h('div', { class: 'flex flex-wrap gap-1' }, l.floor ? tag(t('lg.floorOn'), 'warning') : null,
            l.resetTo != null ? tag(t('lg.resetTo', { n: fmtNum(l.resetTo) }), 'secondary') : null, !l.floor && l.resetTo == null ? h('span', { class: 'text-secondary-400', text: t('none') }) : null),
          l.prizeKey ? mono(l.prizeKey) : h('span', { class: 'text-secondary-400', text: t('none') }),
          h('span', { text: fmtNum(l.players || 0) }),
          l.isActive ? tag(t('active'), 'success') : tag(t('inactive'), 'warning'),
          button({ label: t('delete'), variant: 'danger', data: { 'clan-delete': l.key }, onclick: function () {
            confirmBox(t('lg.delete', { key: l.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/trophies/leagues/' + l.id).then(function () { toast('success', t('deleted')); reload(); }, function (err) { toast('error', clErrText(err)); });
            });
          } })
        ] };
      }), function (l) { leagueEditor(l, reload); })));
      if (!items.length) box.appendChild(emptyCard(t('lg.empty')));
    }, function (err) { clear(box).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}

function leagueEditor(l, done) {
  var creating = !l;
  l = l || { name: {}, description: {}, isActive: true, floor: false, resetTo: null, minTrophies: 0, color: '', iconUrl: '', prizeKey: '', metadata: {} };
  var key = input(l.key || '', { name: 'key', ltr: true, mono: true });
  var name = translated(t('f.name'), 'name', l.name);
  var desc = translated(t('f.description'), 'description', l.description, { multiline: true });
  var icon = input(l.iconUrl, { name: 'iconUrl', ltr: true });
  var color = colorField(l.color);
  var min = numInput(l.minTrophies || 0, 'minTrophies');
  var floor = check(l.floor, t('lg.floor'), 'floor');
  var reset = numInput(l.resetTo == null ? '' : l.resetTo, 'resetTo');
  var prize = refSelect(refList('trophies', 'prizes'), l.prizeKey, 'prizeKey', t('none'));
  var active = check(l.isActive, t('f.active'), 'isActive');
  var meta = jsonInput(l.metadata, 'metadata');
  var msg = h('div', {});
  var body = h('div', { 'data-clan-form': 'league' },
    h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('lg.hint') }),
    field(t('f.key'), key, t('hint.key')), name, desc,
    grid(field(t('f.icon'), icon), field(t('f.color'), color)),
    grid(field(t('lg.min'), min, t('lg.hint.min')), field(t('lg.reset'), reset, t('lg.hint.reset'))),
    floor, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('lg.hint.floor') }),
    field(t('lg.prize'), prize, t('lg.hint.prize')), active,
    field(t('f.metadata'), meta, t('hint.json')), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'clan-action': 'save-league' }, onclick: function () {
    clear(msg);
    var m = meta.read();
    if (m === undefined) { msg.appendChild(notice('error', t('code.INVALID_JSON') + ' · metadata')); return; }
    var r = numberValue(reset);
    var out = { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), color: color.input.value.trim(), minTrophies: intOr(min, 0),
      floor: floor.input.checked, resetTo: r == null || isNaN(r) ? null : r, prizeKey: prize.value, isActive: active.input.checked, metadata: m };
    busy(save, true);
    api(creating ? 'POST' : 'PUT', '/trophies/leagues' + (creating ? '' : '/' + l.id), out).then(function (d) {
      clLeagues = d; closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  sidePanel(creating ? t('lg.new') : textOf(l.name, l.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'league');
}

function trPlayersPage(root) {
  var state = { page: 1, limit: 25, total: 0, q: '', league: '*' };
  var frame = pageFrame('trophies/players', t('desc.trophies/players'), [
    button({ label: t('tp.retry'), data: { 'clan-action': 'retry-prizes' }, onclick: function (e) {
      var b = e.currentTarget;
      busy(b, true);
      api('POST', '/trophies/prizes/retry', {}).then(function (d) { busy(b, false); toast('success', t('tp.retried', { n: fmtNum(d.retried || 0) })); }, function (err) { busy(b, false); toast('error', clErrText(err)); });
    } })]);
  clear(root).appendChild(frame.el);
  var search = searchBox(function (v) { state.q = v; state.page = 1; reload(); });
  var league = select('*', [{ value: '*', label: t('all') }], { name: 'league' });
  league.addEventListener('change', function () { state.league = league.value; state.page = 1; reload(); });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end cl-filters' }, field(t('tp.search'), search, t('tp.hint.search')), field(t('tp.league'), league)));
  frame.content.appendChild(listBox);
  loadLeagues(true).then(function (d) {
    clear(league);
    [{ value: '*', label: t('all') }, { value: '-', label: t('tr.unranked') }].concat((d.items || []).map(function (l) { return { value: l.key, label: textOf(l.name, l.key) }; })).forEach(function (o) {
      league.appendChild(h('option', { value: o.value, text: o.label }));
    });
    league.value = state.league;
  }, function () {});
  function reload() {
    clear(listBox).appendChild(loading());
    var q = qs({ q: state.q, page: state.page, limit: state.limit, league: state.league !== '*' && state.league !== '-' ? state.league : null });
    if (state.league === '-') q += (q ? '&' : '?') + 'league=';
    Promise.all([api('GET', '/trophies/players' + q), loadLeagues()]).then(function (res) {
      var d = res[0];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-clan-list': 'trophy-players' }, table([
        { label: t('f.player') }, { label: t('tp.trophies') }, { label: t('tp.best') }, { label: t('tp.league') }, { label: t('tp.record') }, { label: t('f.updated') }
      ], items.map(function (p) {
        return { id: p.userId, item: p, cells: [
          h('div', { class: 'flex flex-col' }, playerLink(p.userId, p.displayName || p.username), p.username ? mono(p.username) : null),
          h('span', { class: 'font-semibold', text: fmtNum(p.trophies) }),
          h('span', { text: t('tp.bestV', { b: fmtNum(p.best), s: fmtNum(p.seasonBest) }) }),
          p.league ? tag(leagueName(p.league), 'primary') : tag(t('tr.unranked'), 'secondary'),
          h('span', { text: t('tp.wld', { w: fmtNum(p.wins), l: fmtNum(p.losses), d: fmtNum(p.draws) }) }),
          h('span', { class: 'text-xs', text: fmtDate(p.updatedAt) })
        ] };
      }), function (p) { trophyPlayerPanel(p.userId, p.displayName || p.username, reload); })));
      if (!items.length) {
        listBox.appendChild(emptyCard(t('tp.empty')));
        if (isUUID(state.q)) listBox.appendChild(h('div', {}, button({ label: t('tp.open'), variant: 'primary', data: { 'clan-action': 'open-player' }, onclick: function () { trophyPlayerPanel(state.q.toLowerCase(), '', reload); } })));
      }
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}

function trophyPlayerPanel(userId, label, done) {
  var body = h('div', { 'data-clan-player': userId }, loading());
  sidePanel(label || userId.slice(0, 8), body, [button({ label: t('close'), onclick: closePanel })], 'trophy-player');
  function paint(d) {
    clear(body);
    var amount = numInput('', 'trophies', t('tp.amount'));
    var mode = select('set', opt('tp.mode.', ['set', 'add']), { name: 'mode' });
    var msg = h('div', {});
    var apply = button({ label: t('tp.apply'), variant: 'primary', data: { 'clan-action': 'set-trophies' }, onclick: function () {
      clear(msg);
      var n = numberValue(amount);
      if (n == null || isNaN(n)) { msg.appendChild(notice('error', t('code.VALUE_OUT_OF_RANGE') + ' · trophies')); return; }
      busy(apply, true);
      api('PUT', '/trophies/players/' + userId, { trophies: n, mode: mode.value }).then(function (out) {
        toast('success', t('saved')); paint(out); if (done) done();
      }, function (err) { busy(apply, false); fieldError(msg, err); });
    } });
    body.appendChild(h('div', { class: 'mb-4' }, dl([
      [t('f.player'), playerLink(userId, label)],
      [t('tp.trophies'), h('span', { class: 'font-semibold', 'data-clan-value': 'trophies', text: fmtNum(d.trophies) })],
      [t('tp.best'), t('tp.bestV', { b: fmtNum(d.best), s: fmtNum(d.seasonBest) })],
      [t('tp.league'), d.league ? leagueName(d.league) : t('tr.unranked')],
      [t('tp.record'), t('tp.wld', { w: fmtNum(d.wins), l: fmtNum(d.losses), d: fmtNum(d.draws) })]
    ], 'trophy-player')));
    body.appendChild(section(t('tp.change'), h('div', {},
      h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('tp.hint.change') }),
      h('div', { class: 'cl-row' }, field(t('tp.amount'), amount), field(t('tp.modeL'), mode), h('div', {}, apply)), msg), 'set'));
    body.appendChild(h('div', { class: 'h-3' }));
    var log = d.log || [];
    body.appendChild(section(t('tp.log'), log.length ? table([
      { label: t('f.date') }, { label: t('tp.reason') }, { label: t('tp.delta') }, { label: t('tp.after') }, { label: t('tp.league') }, { label: t('tp.opponent') }
    ], log.map(function (x) {
      return { id: String(x.id), cells: [
        h('span', { class: 'text-xs', text: fmtDate(x.createdAt) }),
        tag(has('tp.reason.' + x.reason) ? t('tp.reason.' + x.reason) : x.reason, x.reason === 'win' ? 'success' : x.reason === 'loss' ? 'error' : 'secondary'),
        deltaText(x.delta),
        h('span', { dir: 'ltr', text: digits(fmtNum(x.before) + ' → ' + fmtNum(x.after)) }),
        x.leagueBefore !== x.leagueAfter ? h('span', { text: tr('tp.move', null, { a: leagueName(x.leagueBefore), b: leagueName(x.leagueAfter) }) }) : h('span', { text: leagueName(x.leagueAfter) }),
        x.opponent ? playerLink(x.opponent.userId, x.opponent.displayName || x.opponent.username) : h('span', { class: 'text-secondary-400', text: t('none') })
      ] };
    })) : h('p', { class: 'text-sm text-secondary-500', text: t('tp.noLog') }), 'log'));
    body.appendChild(h('div', { class: 'h-3' }));
    var prizes = d.prizes || [];
    body.appendChild(section(t('tp.prizes'), prizes.length ? table([
      { label: t('tp.league') }, { label: t('tp.season') }, { label: t('lg.prize') }, { label: t('f.state') }, { label: t('f.date') }
    ], prizes.map(function (x) {
      return { id: x.claimKey, cells: [
        h('span', { text: leagueName(x.league) }),
        h('span', { text: x.season ? fmtNum(x.season) : t('tp.ever') }),
        mono(x.prizeKey),
        h('div', { class: 'flex flex-col gap-1' }, tag(t('pay.' + x.status), x.status === 'paid' ? 'success' : x.status === 'failed' ? 'error' : 'warning'),
          x.error ? h('span', { class: 'text-xs text-error-600', dir: 'ltr', text: x.error }) : null),
        h('span', { class: 'text-xs', text: fmtDate(x.paidAt || x.createdAt) })
      ] };
    })) : h('p', { class: 'text-sm text-secondary-500', text: t('tp.noPrizes') }), 'prizes'));
  }
  Promise.all([api('GET', '/trophies/players/' + userId), loadLeagues()]).then(function (res) { paint(res[0]); }, function (err) { clear(body).appendChild(notice('error', clErrText(err))); });
}

function seasonsPage(root) {
  var frame = pageFrame('trophies/seasons', t('desc.trophies/seasons'));
  clear(root).appendChild(frame.el);
  function paint(d) {
    clear(frame.content);
    var reset = button({ label: t('ss.reset'), variant: 'danger', class: 'cl-btn--outline', data: { 'clan-action': 'reset-season' }, onclick: function () {
      confirmBox(t('ss.confirm', { mode: t('ss.mode.' + d.mode) }), t('ss.reset')).then(function (ok) {
        if (!ok) return;
        busy(reset, true);
        api('POST', '/trophies/seasons/reset', { confirm: true }).then(function (out) { toast('success', t('ss.done')); paint(out); }, function (err) { busy(reset, false); toast('error', clErrText(err)); });
      });
    } });
    frame.content.appendChild(card(h('div', { class: 'p-4 flex flex-col sm:flex-row gap-4 justify-between', 'data-clan-section': 'season-info' },
      dl([
        [t('ss.auto'), d.enabled ? tag(t('on'), 'success') : tag(t('off'), 'secondary')],
        [t('ss.schedule'), d.schedule ? mono(d.schedule) : t('none')],
        [t('ss.modeL'), t('ss.mode.' + d.mode)],
        [t('ss.next'), d.enabled && d.nextRun ? h('span', { 'data-clan-value': 'next-run', text: fmtDate(d.nextRun) }) : t('ss.manual')]
      ], 'season'),
      h('div', { class: 'flex flex-col gap-2 items-start' }, reset, h('span', { class: 'text-xs text-secondary-500 max-w-xs', text: t('ss.hint.reset') })))));
    var items = d.items || [];
    frame.content.appendChild(h('div', { 'data-clan-list': 'seasons' }, table([
      { label: t('ss.number') }, { label: t('ss.started') }, { label: t('ss.ends') }, { label: t('ss.ended') }, { label: t('ss.players') }, { label: t('ss.modeL') }, { label: t('ss.actor') }
    ], items.map(function (s) {
      return { id: String(s.number), cells: [
        h('span', { class: 'font-semibold', text: t('ss.n', { n: s.number }) }),
        h('span', { class: 'text-xs', text: fmtDate(s.startedAt) }),
        h('span', { class: 'text-xs', text: s.endsAt ? fmtDate(s.endsAt) : t('none') }),
        s.endedAt ? h('span', { class: 'text-xs', text: fmtDate(s.endedAt) }) : tag(t('ss.current'), 'success'),
        h('span', { text: s.endedAt ? fmtNum(s.playersReset) : t('none') }),
        h('span', { text: s.mode ? t('ss.mode.' + s.mode) : t('none') }),
        h('span', { dir: 'auto', text: s.actor || (s.endedAt ? t('ss.scheduler') : t('none')) })
      ] };
    }))));
    if (!items.length) frame.content.appendChild(emptyCard(t('ss.empty')));
  }
  frame.content.appendChild(loading());
  api('GET', '/trophies/seasons').then(paint, function (err) { clear(frame.content).appendChild(notice('error', clErrText(err))); });
}

function checkList(items, chosen, name) {
  var box = h('div', { class: 'cl-scope-list', 'data-clan-checks': name });
  var all = items.slice();
  (chosen || []).forEach(function (k) { if (!all.some(function (x) { return x.key === k; })) all.push({ key: k, name: k + ' (' + t('missing') + ')' }); });
  var boxes = all.map(function (x) {
    var c = check((chosen || []).indexOf(x.key) !== -1, (x.name && x.name !== x.key ? x.name + ' · ' : '') + x.key, name);
    c.input.value = x.key;
    box.appendChild(c);
    return c;
  });
  if (!all.length) box.appendChild(h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('none') }));
  box.read = function () { return boxes.filter(function (c) { return c.input.checked; }).map(function (c) { return c.input.value; }); };
  return box;
}

function trophySettingsPage(root) {
  loadRefs('trophies', true).then(function () {
    settingsPage(root, 'trophies/settings', '/trophies', function (cur, defs, d) {
      var notes = h('div', { class: 'flex flex-col gap-2' });
      if (d.prizes === false) notes.appendChild(notice('warning', t('noPrizes')));
      var m = cur.match, a = cur.amounts, p = cur.prizes, s = cur.season, mm = cur.matchmaking, ds = cur.display, ev = cur.events;
      var enabled = check(cur.enabled, t('ts.enabled'), 'enabled');
      var source = select(m.source, opt('src.', ['economy_event', 'game_event']), { name: 'match.source' });
      var event = eventInput(refList('trophies', 'economyEvents'), m.event, 'match.event', 'match_end');
      var won = input(m.wonField, { name: 'match.wonField', ltr: true, mono: true });
      var draw = input(m.drawField, { name: 'match.drawField', ltr: true, mono: true });
      var matchF = input(m.matchField, { name: 'match.matchField', ltr: true, mono: true });
      var opp = input(m.opponentField, { name: 'match.opponentField', ltr: true, mono: true });
      var modeF = input(m.modeField, { name: 'match.modeField', ltr: true, mono: true });
      var modes = tagsInput(m.modes, 'match.modes');
      var skip = input(m.skipField, { name: 'match.skipField', ltr: true, mono: true });
      var start = numInput(a.start, 'amounts.start');
      var win = numInput(a.win, 'amounts.win');
      var loss = numInput(a.loss, 'amounts.loss');
      var drawA = numInput(a.draw, 'amounts.draw');
      var scale = check(a.scale, t('ts.scale'), 'amounts.scale');
      var scalePer = numInput(a.scalePer, 'amounts.scalePer');
      var scaleStep = numInput(a.scaleStep, 'amounts.scaleStep');
      var minWin = numInput(a.minWin, 'amounts.minWin');
      var maxWin = numInput(a.maxWin, 'amounts.maxWin');
      var minLoss = numInput(a.minLoss, 'amounts.minLoss');
      var maxLoss = numInput(a.maxLoss, 'amounts.maxLoss');
      var floors = check(a.floors, t('ts.floors'), 'amounts.floors');
      var oncePer = select(p.oncePer, opt('ts.once.', ['ever', 'season']), { name: 'prizes.oncePer' });
      var rising = check(p.onlyRising, t('ts.onlyRising'), 'prizes.onlyRising');
      var tries = numInput(p.maxTries, 'prizes.maxTries');
      var sOn = check(s.enabled, t('ts.seasonOn'), 'season.enabled');
      var sched = input(s.schedule, { name: 'season.schedule', ltr: true, mono: true, placeholder: '0 0 1 * *' });
      var sMode = select(s.mode, opt('ss.mode.', ['cut', 'league', 'zero']), { name: 'season.mode' });
      var above = numInput(s.above, 'season.above');
      var keep = numInput(s.keepPercent, 'season.keepPercent');
      var cutBox = grid(field(t('ts.above'), above, t('ts.hint.above')), field(t('ts.keep'), keep, t('ts.hint.keep')));
      var mmOn = check(mm.enabled, t('ts.mmOn'), 'matchmaking.enabled');
      var games = checkList(refList('trophies', 'games'), mm.games, 'matchmaking.games');
      var range = numInput(mm.range, 'matchmaking.range');
      var prop = input(mm.property, { name: 'matchmaking.property', ltr: true, mono: true });
      var hist = numInput(ds.historyLimit, 'display.historyLimit');
      var logDays = numInput(ds.logDays, 'display.logDays');
      var showOpp = check(ds.showOpponent, t('ts.showOpponent'), 'display.showOpponent');
      var pub = check(ev.publish, t('ts.publish'), 'events.publish');
      var notify = check(ev.notify, t('ts.notify'), 'events.notify');
      var nCode = numInput(ev.notifyCode, 'events.notifyCode');
      var lCode = numInput(ev.leagueCode, 'events.leagueCode');
      function paint() { cutBox.hidden = sMode.value !== 'cut'; }
      sMode.addEventListener('change', paint);
      paint();
      return {
        el: [notes,
          section(t('ts.general'), h('div', {}, enabled, h('p', { class: 'text-xs text-secondary-500', text: t('ts.hint.enabled') })), 'general'),
          section(t('ts.match'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('ts.hint.match') }),
            grid(field(t('ts.source'), source), field(t('ts.event'), event, t('ts.hint.event'))),
            grid(field(t('ts.won'), won, t('ts.hint.won')), field(t('ts.draw'), draw, t('ts.hint.draw'))),
            grid(field(t('ts.matchField'), matchF, t('ts.hint.matchField')), field(t('ts.opponent'), opp, t('ts.hint.opponent'))),
            grid(field(t('ts.modeField'), modeF, t('ts.hint.modeField')), field(t('ts.modes'), modes, t('ts.hint.modes'))),
            field(t('ts.skip'), skip, t('ts.hint.skip'))), 'match'),
          section(t('ts.amounts'), h('div', {}, grid3(field(t('ts.start'), start, t('ts.hint.start')), field(t('ts.win'), win), field(t('ts.loss'), loss)),
            grid3(field(t('ts.drawA'), drawA, t('ts.hint.drawA')), h('div', {}), h('div', {})), scale, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('ts.hint.scale') }),
            grid(field(t('ts.scalePer'), scalePer), field(t('ts.scaleStep'), scaleStep)),
            grid(field(t('ts.minWin'), minWin), field(t('ts.maxWin'), maxWin)),
            grid(field(t('ts.minLoss'), minLoss), field(t('ts.maxLoss'), maxLoss)), floors, h('p', { class: 'text-xs text-secondary-500 -mt-2', text: t('ts.hint.floors') })), 'amounts'),
          section(t('ts.prizes'), h('div', {}, grid(field(t('ts.oncePer'), oncePer, t('ts.hint.oncePer')), field(t('ts.tries'), tries, t('ts.hint.tries'))), rising), 'prizes'),
          section(t('ts.season'), h('div', {}, sOn, grid(field(t('ts.schedule'), sched, t('ts.hint.schedule')), field(t('ss.modeL'), sMode, t('ts.hint.sMode'))), cutBox), 'season'),
          section(t('ts.mm'), h('div', {}, mmOn, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('ts.hint.mm') }), field(t('ts.games'), games, t('ts.hint.games')),
            grid(field(t('ts.range'), range, t('ts.hint.range')), field(t('ts.property'), prop, t('ts.hint.property')))), 'matchmaking'),
          section(t('ts.display'), h('div', {}, grid(field(t('ts.hist'), hist, t('ts.hint.hist')), field(t('ts.logDays'), logDays, t('ts.hint.logDays'))), showOpp), 'display'),
          section(t('ts.events'), h('div', {}, pub, notify, grid(field(t('ts.nCode'), nCode, t('ts.hint.nCode')), field(t('ts.lCode'), lCode, t('ts.hint.lCode')))), 'events')],
        read: function (c) {
          c.enabled = enabled.input.checked;
          c.match = { source: source.value, event: event.input.value.trim(), wonField: won.value.trim(), drawField: draw.value.trim(), matchField: matchF.value.trim(),
            opponentField: opp.value.trim(), modeField: modeF.value.trim(), modes: modes.read(), skipField: skip.value.trim() };
          c.amounts = { start: intOr(start, 0), win: intOr(win, 0), loss: intOr(loss, 0), draw: intOr(drawA, 0), scale: scale.input.checked, scalePer: intOr(scalePer, 0),
            scaleStep: intOr(scaleStep, 0), minWin: intOr(minWin, 0), maxWin: intOr(maxWin, 0), minLoss: intOr(minLoss, 0), maxLoss: intOr(maxLoss, 0), floors: floors.input.checked };
          c.prizes = { oncePer: oncePer.value, onlyRising: rising.input.checked, maxTries: intOr(tries, 0) };
          c.season = { enabled: sOn.input.checked, schedule: sched.value.trim(), mode: sMode.value, above: intOr(above, 0), keepPercent: intOr(keep, 0) };
          c.matchmaking = { enabled: mmOn.input.checked, games: games.read(), range: intOr(range, 0), property: prop.value.trim() };
          c.display = { historyLimit: intOr(hist, 0), logDays: intOr(logDays, 0), showOpponent: showOpp.input.checked };
          c.events = { publish: pub.input.checked, notify: notify.input.checked, notifyCode: intOr(nCode, 0), leagueCode: intOr(lCode, 0) };
          return null;
        }
      };
    });
  });
}

var CL_ROLES = ['leader', 'co_leader', 'elder', 'member'];
var clClanSettings = null;
function loadClanSettings(force) {
  if (clClanSettings && !force) return Promise.resolve(clClanSettings);
  return api('GET', '/clans/settings').then(function (d) { clClanSettings = d.settings; return clClanSettings; }, function () { return null; });
}

function joinTag(j) {
  return tag(t('jt.' + j), j === 'open' ? 'success' : j === 'invite' ? 'primary' : 'warning');
}

function levelBar(c) {
  var span = (c.nextLevelXp || 0) - (c.levelXp || 0);
  var pct = span > 0 ? Math.max(0, Math.min(100, Math.round(((c.xp - c.levelXp) / span) * 100))) : 100;
  var bar = h('div', { class: 'cl-bar', role: 'progressbar', 'aria-valuenow': pct, 'aria-valuemin': 0, 'aria-valuemax': 100 }, h('span', {}));
  bar.firstChild.style.width = pct + '%';
  return h('div', { class: 'flex flex-col gap-1 min-w-[7rem]' }, h('span', { class: 'text-xs', text: c.nextLevelXp ? t('cl.xpOf', { x: fmtNum(c.xp), n: fmtNum(c.nextLevelXp) }) : t('cl.xpMax', { x: fmtNum(c.xp) }) }), bar);
}

function clansPage(root) {
  var state = { page: 1, limit: 25, total: 0, q: '', joinType: '', sort: 'trophies' };
  var frame = pageFrame('clans/list', t('desc.clans/list'));
  clear(root).appendChild(frame.el);
  var search = searchBox(function (v) { state.q = v; state.page = 1; reload(); });
  var jt = select('', [{ value: '', label: t('all') }].concat(opt('jt.', ['open', 'closed', 'invite'])), { name: 'joinType' });
  var sort = select('trophies', opt('cl.sort.', ['trophies', 'level', 'members', 'new', 'wars']), { name: 'sort' });
  jt.addEventListener('change', function () { state.joinType = jt.value; state.page = 1; reload(); });
  sort.addEventListener('change', function () { state.sort = sort.value; state.page = 1; reload(); });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end cl-filters' }, field(t('cl.search'), search, t('cl.hint.search')), field(t('cl.joinType'), jt), field(t('cl.sortL'), sort)));
  frame.content.appendChild(listBox);
  function reload() {
    clear(listBox).appendChild(loading());
    api('GET', '/clans/list' + qs({ q: state.q, joinType: state.joinType, sort: state.sort, page: state.page, limit: state.limit })).then(function (d) {
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-clan-list': 'clans' }, table([
        { label: t('cl.clan') }, { label: t('cl.level') }, { label: t('cl.members') }, { label: t('tp.trophies') }, { label: t('cl.wars') }, { label: t('cl.joinType') }, { label: t('f.created') }
      ], items.map(function (c) {
        return { id: c.clanId, item: c, cells: [
          nameCell(c.name, '#' + c.tag, c.avatarUrl, c.badge && c.badge.color),
          h('div', { class: 'flex items-center gap-2' }, tag(t('cl.lv', { n: c.level }), 'primary'), levelBar(c)),
          h('span', { text: t('cl.of', { n: fmtNum(c.members), m: fmtNum(c.maxMembers) }) }),
          h('span', { class: 'font-semibold', text: fmtNum(c.trophies) }),
          h('span', { text: t('tp.wld', { w: fmtNum(c.warWins), l: fmtNum(c.warLosses), d: fmtNum(c.warDraws) }) }),
          joinTag(c.joinType),
          h('span', { class: 'text-xs', text: fmtDate(c.createdAt) })
        ] };
      }), function (c) { clanPanel(c.clanId, c.name, reload); })));
      if (!items.length) listBox.appendChild(emptyCard(t('cl.empty')));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}

function badgeControl(list, value, name) {
  if (list && list.length) {
    var opts = [{ value: '', label: t('none') }].concat(list.map(function (v) { return { value: v, label: v }; }));
    if (value && list.indexOf(value) === -1) opts.push({ value: value, label: value + ' (' + t('missing') + ')' });
    var el = select(value || '', opts, { name: name, label: name });
    el.setAttribute('dir', 'ltr');
    return el;
  }
  return input(value || '', { name: name, ltr: true, mono: true });
}

function clanPanel(clanId, label, done) {
  var body = h('div', { 'data-clan-detail': clanId }, loading());
  var del = button({ label: t('cl.delete'), variant: 'danger', data: { 'clan-action': 'delete-clan' }, onclick: function () {
    confirmBox(t('cl.deleteConfirm', { name: label }), t('delete')).then(function (ok) {
      if (!ok) return;
      busy(del, true);
      api('DELETE', '/clans/list/' + clanId).then(function () { closePanel(); toast('success', t('deleted')); if (done) done(); }, function (err) { busy(del, false); toast('error', clErrText(err)); });
    });
  } });
  sidePanel(label, body, [del, h('div', { class: 'flex-1' }), button({ label: t('close'), onclick: closePanel })], 'clan');
  function changed(d) { paint(d); if (done) done(); }
  function paint(d) {
    var c = d.clan;
    var cfg = clClanSettings || { badge: {} };
    var badge = c.badge || {};
    label = c.name;
    clear(body);
    body.appendChild(h('div', { class: 'mb-4' }, dl([
      [t('cl.tag'), mono('#' + c.tag)],
      [t('cl.level'), h('div', { class: 'flex items-center gap-2' }, tag(t('cl.lv', { n: c.level }), 'primary'), levelBar(c))],
      [t('cl.members'), h('span', { 'data-clan-value': 'members', text: t('cl.of', { n: fmtNum(c.members), m: fmtNum(c.maxMembers) }) })],
      [t('tp.trophies'), fmtNum(c.trophies)],
      [t('cl.wars'), t('cl.warStats', { w: fmtNum(c.warWins), l: fmtNum(c.warLosses), d: fmtNum(c.warDraws), p: fmtNum(c.warPoints) })],
      [t('cl.requests'), fmtNum(d.requests || 0)],
      [t('f.created'), fmtDate(c.createdAt)]
    ], 'clan')));
    var name = input(c.name, { name: 'name' });
    var desc = input(c.description, { name: 'description', multiline: true });
    var avatar = input(c.avatarUrl, { name: 'avatarUrl', ltr: true });
    var lang = input(c.lang, { name: 'lang', ltr: true, mono: true, placeholder: 'fa' });
    var emblem = badgeControl(cfg.badge.emblems, badge.emblem, 'badge.emblem');
    var color = badgeControl(cfg.badge.colors, badge.color, 'badge.color');
    var frameB = badgeControl(cfg.badge.frames, badge.frame, 'badge.frame');
    var jt = select(c.joinType, opt('jt.', ['open', 'closed', 'invite']), { name: 'joinType' });
    var minT = numInput(c.minTrophies || 0, 'minTrophies');
    var pmsg = h('div', {});
    var psave = button({ label: t('save'), variant: 'primary', data: { 'clan-action': 'save-profile' }, onclick: function () {
      clear(pmsg);
      busy(psave, true);
      api('PUT', '/clans/list/' + clanId, { name: name.value.trim(), description: desc.value.trim(), avatarUrl: avatar.value.trim(), lang: lang.value.trim(),
        badge: { emblem: emblem.value.trim(), color: color.value.trim(), frame: frameB.value.trim() }, joinType: jt.value, minTrophies: intOr(minT, 0) }).then(function (out) {
        toast('success', t('saved')); changed(out);
      }, function (err) { busy(psave, false); fieldError(pmsg, err); });
    } });
    body.appendChild(section(t('cl.profile'), h('div', { 'data-clan-form': 'profile' }, grid(field(t('f.name'), name), field(t('cl.lang'), lang)), field(t('f.description'), desc),
      field(t('cl.avatar'), avatar), grid3(field(t('cl.emblem'), emblem), field(t('f.color'), color), field(t('cl.frame'), frameB)),
      grid(field(t('cl.joinType'), jt), field(t('cl.minTrophies'), minT, t('cl.hint.minTrophies'))), pmsg, h('div', { class: 'flex justify-end' }, psave)), 'profile'));
    body.appendChild(h('div', { class: 'h-3' }));
    var amount = numInput('', 'xp-amount', t('cl.xpAmount'));
    var xmsg = h('div', {});
    var xbtn = button({ label: t('cl.addXp'), variant: 'primary', data: { 'clan-action': 'add-xp' }, onclick: function () {
      clear(xmsg);
      var n = numberValue(amount);
      if (!n || isNaN(n) || n <= 0) { xmsg.appendChild(notice('error', t('code.VALUE_OUT_OF_RANGE') + ' · amount')); return; }
      busy(xbtn, true);
      api('POST', '/clans/list/' + clanId + '/xp', { amount: n }).then(function (out) { toast('success', t('cl.xpAdded', { n: fmtNum(n) })); changed(out); }, function (err) { busy(xbtn, false); fieldError(xmsg, err); });
    } });
    body.appendChild(section(t('cl.xp'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('cl.hint.xp') }),
      h('div', { class: 'cl-row' }, field(t('cl.xpAmount'), amount), h('div', {}, xbtn)), xmsg), 'add-xp'));
    body.appendChild(h('div', { class: 'h-3' }));
    var members = d.members || [];
    var mlist = h('div', { class: 'cl-rows' });
    members.forEach(function (m) {
      var role = select(m.role, opt('role.', CL_ROLES), { name: 'role', label: t('cl.role') });
      var rbtn = button({ label: t('cl.setRole'), data: { 'clan-action': 'role' }, onclick: function () {
        if (role.value === m.role) return;
        var go = role.value === 'leader' ? confirmBox(t('cl.leaderConfirm', { name: m.displayName || m.username }), t('cl.setRole')) : Promise.resolve(true);
        go.then(function (ok) {
          if (!ok) { role.value = m.role; return; }
          busy(rbtn, true);
          api('PUT', '/clans/list/' + clanId + '/members/' + m.userId, { role: role.value }).then(function (out) { toast('success', t('saved')); changed(out); }, function (err) { busy(rbtn, false); role.value = m.role; toast('error', clErrText(err)); });
        });
      } });
      var kick = button({ label: t('cl.kick'), variant: 'danger', data: { 'clan-action': 'kick' }, onclick: function () {
        confirmBox(t('cl.kickConfirm', { name: m.displayName || m.username }), t('cl.kick')).then(function (ok) {
          if (!ok) return;
          busy(kick, true);
          api('DELETE', '/clans/list/' + clanId + '/members/' + m.userId).then(function (out) { toast('success', t('cl.kicked')); changed(out); }, function (err) { busy(kick, false); toast('error', clErrText(err)); });
        });
      } });
      mlist.appendChild(h('div', { class: 'cl-member', 'data-clan-member': m.userId },
        h('div', { class: 'flex flex-col min-w-0 cl-member-who' }, playerLink(m.userId, m.displayName || m.username),
          h('span', { class: 'text-xs text-secondary-500', text: t('cl.memberLine', { x: fmtNum(m.xp), p: fmtNum(m.warPoints), t: fmtNum(m.trophies) }) + (m.joinedAt ? ' · ' + fmtDate(m.joinedAt) : '') })),
        h('div', { class: 'flex flex-wrap gap-2 items-center' }, tag(t('role.' + m.role), m.role === 'leader' ? 'warning' : m.role === 'member' ? 'secondary' : 'primary'), role, rbtn, kick)));
    });
    body.appendChild(section(t('cl.membersT', { n: fmtNum(members.length) }), members.length ? mlist : h('p', { class: 'text-sm text-secondary-500', text: t('cl.noMembers') }), 'members'));
    body.appendChild(h('div', { class: 'h-3' }));
    var xp = d.xp || [];
    body.appendChild(section(t('cl.xpLog'), xp.length ? table([{ label: t('f.date') }, { label: t('cl.source') }, { label: t('cl.amount') }, { label: t('f.player') }],
      xp.map(function (x, i) {
        return { id: String(i), cells: [h('span', { class: 'text-xs', text: fmtDate(x.createdAt) }), h('span', { text: has('xs.' + x.source) ? t('xs.' + x.source) : x.source }), deltaText(x.amount),
          x.userId ? playerLink(x.userId, x.username) : h('span', { class: 'text-secondary-400', text: t('none') })] };
      })) : h('p', { class: 'text-sm text-secondary-500', text: t('cl.noXp') }), 'xp'));
    body.appendChild(h('div', { class: 'h-3' }));
    var wars = d.wars || [];
    body.appendChild(section(t('cl.warsT'), wars.length ? table([{ label: t('wr.config') }, { label: t('wr.mode') }, { label: t('f.state') }, { label: t('rd.score') }, { label: t('rd.ends') }],
      wars.map(function (w) {
        var mine = w.clanA === clanId;
        return { id: w.warId, cells: [mono(w.configKey), h('span', { text: t('mode.' + w.mode) }),
          w.status === 'active' ? tag(t('rd.st.active'), 'warning') : !w.outcome ? tag(t('out.cancelled'), 'secondary') : tag(t('out.' + w.outcome), w.outcome === 'win' ? 'success' : w.outcome === 'lose' ? 'error' : 'secondary'),
          h('span', { dir: 'ltr', text: digits(String(mine ? w.scoreA : w.scoreB) + ' : ' + String(mine ? w.scoreB : w.scoreA)) }), h('span', { class: 'text-xs', text: fmtDate(w.endsAt) })] };
      })) : h('p', { class: 'text-sm text-secondary-500', text: t('cl.noWars') }), 'wars'));
  }
  Promise.all([api('GET', '/clans/list/' + clanId), loadClanSettings()]).then(function (res) { paint(res[0]); }, function (err) { clear(body).appendChild(notice('error', clErrText(err))); });
}

function levelsPage(root) {
  var frame = pageFrame('clans/levels', t('desc.clans/levels'));
  clear(root).appendChild(frame.el);
  frame.content.appendChild(loading());
  function paint(d) {
    clear(frame.content);
    var rows = [];
    var box = h('div', { class: 'cl-rows', 'data-clan-levels': '' });
    function addRow(l) {
      var lv = numInput(l.level, 'level', t('lv.level'));
      var xp = numInput(l.xp, 'xp', t('lv.xp'));
      var cap = numInput(l.memberCap, 'memberCap', t('lv.cap'));
      var prize = refSelect(refList('clans', 'prizes'), l.prizeKey, 'prizeKey', t('none'));
      var icon = input(l.iconUrl, { name: 'iconUrl', ltr: true });
      var perks = jsonInput(l.perks, 'perks');
      var names = {};
      var nameBox = h('div', { class: 'cl-row-full grid sm:grid-cols-2 gap-x-4' });
      (clLanguages || []).forEach(function (g) {
        var el = input((l.name || {})[g.code] || '', { name: 'name-' + g.code });
        el.setAttribute('dir', g.direction === 'rtl' ? 'rtl' : 'ltr');
        names[g.code] = el;
        nameBox.appendChild(field(t('lang.text', { label: t('f.name'), lang: g.name }), el));
      });
      var rec = { lv: lv, xp: xp, cap: cap, prize: prize, icon: icon, perks: perks, names: names };
      var row = h('div', { class: 'cl-row cl-level-row', 'data-clan-level': '' },
        field(t('lv.level'), lv), field(t('lv.xp'), xp, t('lv.hint.xp')), field(t('lv.cap'), cap), field(t('lv.prize'), prize),
        nameBox, field(t('f.icon'), icon, null, 'cl-row-full'), field(t('lv.perks'), perks, t('hint.json'), 'cl-row-full'),
        h('div', { class: 'cl-row-full flex items-center justify-between gap-2' },
          h('span', { class: 'text-xs text-secondary-500', text: l.clans != null ? t('lv.clans', { n: fmtNum(l.clans) }) : t('lv.new') }),
          button({ label: t('remove'), variant: 'danger', data: { 'clan-action': 'remove-level' }, onclick: function () { row.remove(); rows = rows.filter(function (r) { return r !== rec; }); } })));
      rec.row = row;
      rows.push(rec);
      box.appendChild(row);
    }
    (d.items || []).forEach(addRow);
    var msg = h('div', { 'data-clan-msg': '' });
    var addBtn = button({ label: '+ ' + t('lv.add'), data: { 'clan-action': 'add-level' }, onclick: function () {
      var last = rows[rows.length - 1];
      var lastLv = last ? intOr(last.lv, rows.length) : 0;
      var lastXp = last ? intOr(last.xp, 0) : -100;
      addRow({ level: lastLv + 1, xp: lastXp + 100 * (lastLv + 1), memberCap: last ? intOr(last.cap, 30) : 30, name: {}, perks: {} });
    } });
    var save = button({ label: t('save'), variant: 'primary', data: { 'clan-save': 'levels' }, onclick: function () {
      clear(msg);
      var items = [];
      for (var i = 0; i < rows.length; i++) {
        var r = rows[i];
        var p = r.perks.read();
        if (p === undefined) { msg.appendChild(notice('error', t('code.INVALID_JSON') + ' · ' + t('lv.perks') + ' ' + digits(String(i + 1)))); return; }
        var name = {};
        Object.keys(r.names).forEach(function (k) { var v = r.names[k].value.trim(); if (v) name[k] = v; });
        items.push({ level: intOr(r.lv, 0), xp: intOr(r.xp, 0), memberCap: intOr(r.cap, 0), prizeKey: r.prize.value, iconUrl: r.icon.value.trim(), name: name, perks: p });
      }
      busy(save, true);
      api('PUT', '/clans/levels', { items: items }).then(function (out) { toast('success', t('saved')); paint(out); }, function (err) { busy(save, false); msg.appendChild(notice('error', clErrText(err))); });
    } });
    frame.content.appendChild(notice('secondary', t('lv.hint')));
    frame.content.appendChild(box);
    frame.content.appendChild(h('div', {}, addBtn));
    frame.content.appendChild(h('div', { class: 'cl-savebar flex flex-wrap items-center justify-between gap-3' }, h('div', { class: 'flex-1' }, msg), save));
  }
  Promise.all([api('GET', '/clans/levels'), loadLanguages(), loadRefs('clans')]).then(function (res) { paint(res[0]); }, function (err) { clear(frame.content).appendChild(notice('error', clErrText(err))); });
}

var clWarDefaults = null;
var clWarConfigs = null;
function loadWarConfigs(force) {
  if (clWarConfigs && !force) return Promise.resolve(clWarConfigs);
  return api('GET', '/clans/wars/configs').then(function (d) { clWarConfigs = d.items || []; clWarDefaults = d.defaults || null; return clWarConfigs; });
}

function eventsFor(type) {
  return type === 'game_event' ? refList('clans', 'gameEvents') : refList('clans', 'economyEvents');
}

function warsPage(root) {
  var frame = pageFrame('clans/wars', t('desc.clans/wars'), [
    button({ label: t('wr.new'), variant: 'primary', data: { 'clan-action': 'new-war' }, onclick: function () { warEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  var box = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(box);
  function reload() {
    clear(box).appendChild(loading());
    Promise.all([loadWarConfigs(true), loadLanguages(), loadRefs('clans')]).then(function (res) {
      var items = res[0];
      clear(box);
      box.appendChild(h('div', { 'data-clan-list': 'war-configs' }, table([
        { label: t('f.name') }, { label: t('wr.mode') }, { label: t('wr.schedule') }, { label: t('wr.duration') }, { label: t('wr.signup') }, { label: t('wr.next') }, { label: t('f.state') }, { label: '' }
      ], items.map(function (c) {
        return { id: c.key, item: c, cells: [
          nameCell(textOf(c.name, c.key), c.key),
          tag(t('mode.' + c.mode), c.mode === 'duel' ? 'warning' : 'primary'),
          c.schedule ? mono(c.schedule) : h('span', { text: t('wr.manual') }),
          h('span', { text: t('wr.minutes', { n: fmtNum(c.durationMinutes) }) }),
          h('span', { text: t('wr.signup.' + c.signup) + (c.signupMinutes ? ' · ' + t('wr.minutes', { n: fmtNum(c.signupMinutes) }) : '') }),
          h('span', { class: 'text-xs', text: c.nextStart ? fmtDate(c.nextStart) : t('none') }),
          c.isActive ? tag(t('active'), 'success') : tag(t('inactive'), 'warning'),
          button({ label: t('delete'), variant: 'danger', data: { 'clan-delete': c.key }, onclick: function () {
            confirmBox(t('wr.delete', { key: c.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/clans/wars/configs/' + c.id).then(function () { toast('success', t('deleted')); reload(); }, function (err) { toast('error', clErrText(err)); });
            });
          } })
        ] };
      }), function (c) { warEditor(c, reload); })));
      if (!items.length) box.appendChild(emptyCard(t('wr.empty')));
    }, function (err) { clear(box).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}

function pointRows(list) {
  var box = h('div', { class: 'cl-rows', 'data-clan-rules': 'points' });
  var rows = [];
  function addRow(p) {
    p = p || { type: 'economy_event', event: '', points: 1 };
    var type = select(p.type, opt('src.', ['economy_event', 'game_event']), { name: 'point-type' });
    var ev = eventInput(eventsFor(p.type), p.event, 'point-event', 'match_end');
    type.addEventListener('change', function () {
      var list = document.getElementById(ev.input.getAttribute('list'));
      if (!list) return;
      clear(list);
      eventsFor(type.value).forEach(function (e) { list.appendChild(h('option', { value: e.key, text: e.name || e.key })); });
    });
    var fld = input(p.field, { name: 'point-field', ltr: true, mono: true, placeholder: 'won' });
    var eq = input(p.equals, { name: 'point-equals', ltr: true, mono: true, placeholder: 'true' });
    var pts = numInput(p.points, 'point-points', t('wr.points'));
    var per = check(p.perValue, t('wr.perValue'), 'point-perValue');
    var rec = { type: type, ev: ev, fld: fld, eq: eq, pts: pts, per: per };
    var row = h('div', { class: 'cl-row', 'data-clan-point': '' }, field(t('wr.source'), type), field(t('wr.event'), ev), field(t('wr.field'), fld), field(t('wr.equals'), eq),
      field(t('wr.points'), pts), h('div', { class: 'cl-row-full flex items-center justify-between gap-2' }, per,
        button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r !== rec; }); } })));
    rows.push(rec);
    box.appendChild(row);
  }
  (list || []).forEach(addRow);
  var wrap = h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('wr.hint.points') }), box,
    h('div', { class: 'mt-2' }, button({ label: '+ ' + t('wr.addPoint'), data: { 'clan-action': 'add-point' }, onclick: function () { addRow(); } })));
  wrap.read = function () {
    return rows.map(function (r) {
      var o = { type: r.type.value, event: r.ev.input.value.trim(), points: intOr(r.pts, 0) };
      if (r.fld.value.trim()) o.field = r.fld.value.trim();
      if (r.eq.value.trim()) o.equals = r.eq.value.trim();
      if (r.per.input.checked) o.perValue = true;
      return o;
    });
  };
  return wrap;
}

function warEditor(c, done) {
  var creating = !c;
  var defs = clWarDefaults || {};
  c = c ? JSON.parse(JSON.stringify(c)) : { name: {}, description: {}, mode: 'points', isActive: true, schedule: '', durationMinutes: 1440, signup: 'auto', signupMinutes: 0, pairing: 'random',
    rules: JSON.parse(JSON.stringify(defs.points || { points: [], duel: {}, eligibility: {}, prizes: {} })) };
  var r = c.rules || {};
  r.duel = r.duel && r.duel.type ? r.duel : JSON.parse(JSON.stringify((defs.duel && defs.duel.duel) || { type: 'economy_event', wonField: 'won', drawField: 'draw', opponentField: 'opponentId', winPoints: 1 }));
  r.eligibility = r.eligibility || {};
  r.prizes = r.prizes || {};
  var prizes = refList('clans', 'prizes');
  var key = input(c.key || '', { name: 'key', ltr: true, mono: true });
  var name = translated(t('f.name'), 'name', c.name);
  var desc = translated(t('f.description'), 'description', c.description, { multiline: true });
  var mode = select(c.mode, opt('mode.', ['points', 'duel']), { name: 'mode' });
  var active = check(c.isActive, t('f.active'), 'isActive');
  var sched = input(c.schedule, { name: 'schedule', ltr: true, mono: true, placeholder: '0 18 * * 5' });
  var dur = numInput(c.durationMinutes, 'durationMinutes');
  var signup = select(c.signup, opt('wr.signup.', ['auto', 'leader']), { name: 'signup' });
  var signupMin = numInput(c.signupMinutes, 'signupMinutes');
  var pairing = select(c.pairing, opt('wr.pair.', ['random', 'level', 'trophies']), { name: 'pairing' });
  var points = pointRows(r.points);
  var d = r.duel;
  var dType = select(d.type, opt('src.', ['economy_event', 'game_event']), { name: 'duel.type' });
  var dEvent = eventInput(eventsFor(d.type), d.event, 'duel.event', 'match_end');
  var dWon = input(d.wonField, { name: 'duel.wonField', ltr: true, mono: true });
  var dDraw = input(d.drawField, { name: 'duel.drawField', ltr: true, mono: true });
  var dOpp = input(d.opponentField, { name: 'duel.opponentField', ltr: true, mono: true });
  var dWin = numInput(d.winPoints, 'duel.winPoints');
  var dDrawP = numInput(d.drawPoints, 'duel.drawPoints');
  var duelBox = h('div', { 'data-clan-rules': 'duel' }, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('wr.hint.duel') }),
    grid(field(t('wr.source'), dType), field(t('wr.event'), dEvent)), grid3(field(t('wr.won'), dWon), field(t('wr.drawF'), dDraw), field(t('wr.opp'), dOpp, t('wr.hint.opp'))),
    grid(field(t('wr.winPoints'), dWin), field(t('wr.drawPoints'), dDrawP)));
  var matchF = input(r.matchField, { name: 'rules.matchField', ltr: true, mono: true });
  var attacks = numInput(r.attacksPerMember || 0, 'rules.attacksPerMember');
  var e = r.eligibility;
  var minM = numInput(e.minMembers || 1, 'eligibility.minMembers');
  var maxM = numInput(e.maxMembers || 0, 'eligibility.maxMembers');
  var minL = numInput(e.minClanLevel || 0, 'eligibility.minClanLevel');
  var minT = numInput(e.minTrophies || 0, 'eligibility.minTrophies');
  var joined = numInput(e.joinedHours || 0, 'eligibility.joinedHours');
  var pz = r.prizes;
  var pWin = refSelect(prizes, (pz.win || {}).prizeKey, 'prizes.win', t('none'));
  var pLose = refSelect(prizes, (pz.lose || {}).prizeKey, 'prizes.lose', t('none'));
  var pDraw = refSelect(prizes, (pz.draw || {}).prizeKey, 'prizes.draw', t('none'));
  var pBye = refSelect(prizes, (pz.bye || {}).prizeKey, 'prizes.bye', t('none'));
  var pTo = select(pz.to || 'participants', opt('wr.to.', ['participants', 'members']), { name: 'prizes.to' });
  var bye = select(r.byeResult || 'none', opt('wr.bye.', ['none', 'win']), { name: 'byeResult' });
  var pointsSec = section(t('wr.pointsRules'), points, 'points');
  var duelSec = section(t('wr.duelRules'), duelBox, 'duel');
  function paint() { pointsSec.hidden = mode.value !== 'points'; duelSec.hidden = mode.value !== 'duel'; }
  mode.addEventListener('change', paint);
  paint();
  var msg = h('div', {});
  var body = h('div', { 'data-clan-form': 'war' },
    h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('wr.hint') }),
    field(t('f.key'), key, t('hint.key')), name, desc,
    grid(field(t('wr.mode'), mode, t('wr.hint.mode')), h('div', { class: 'pt-6' }, active)),
    section(t('wr.timing'), h('div', {}, grid(field(t('wr.schedule'), sched, t('wr.hint.schedule')), field(t('wr.duration'), dur, t('wr.hint.duration'))),
      grid3(field(t('wr.signup'), signup, t('wr.hint.signup')), field(t('wr.signupMinutes'), signupMin, t('wr.hint.signupMinutes')), field(t('wr.pairing'), pairing))), 'timing'), h('div', { class: 'h-3' }),
    pointsSec, duelSec, h('div', { class: 'h-3' }),
    section(t('wr.scoring'), grid(field(t('wr.matchField'), matchF, t('wr.hint.matchField')), field(t('wr.attacks'), attacks, t('wr.hint.attacks'))), 'scoring'), h('div', { class: 'h-3' }),
    section(t('wr.eligibility'), h('div', {}, grid3(field(t('wr.minMembers'), minM), field(t('wr.maxMembers'), maxM, t('hint.zero')), field(t('wr.minLevel'), minL)),
      grid(field(t('wr.minTrophies'), minT), field(t('wr.joined'), joined, t('wr.hint.joined')))), 'eligibility'), h('div', { class: 'h-3' }),
    section(t('wr.prizes'), h('div', {}, grid(field(t('wr.pWin'), pWin), field(t('wr.pLose'), pLose)), grid(field(t('wr.pDraw'), pDraw), field(t('wr.pBye'), pBye)),
      grid(field(t('wr.to'), pTo), field(t('wr.byeResult'), bye, t('wr.hint.bye')))), 'prizes'), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'clan-action': 'save-war' }, onclick: function () {
    clear(msg);
    var rules = { points: points.read(), duel: { type: dType.value, event: dEvent.input.value.trim(), wonField: dWon.value.trim(), drawField: dDraw.value.trim(), opponentField: dOpp.value.trim(),
        winPoints: intOr(dWin, 0), drawPoints: intOr(dDrawP, 0) },
      matchField: matchF.value.trim(), attacksPerMember: intOr(attacks, 0),
      eligibility: { minMembers: intOr(minM, 1), maxMembers: intOr(maxM, 0), minClanLevel: intOr(minL, 0), minTrophies: intOr(minT, 0), joinedHours: intOr(joined, 0) },
      prizes: { win: { prizeKey: pWin.value }, lose: { prizeKey: pLose.value }, draw: { prizeKey: pDraw.value }, bye: { prizeKey: pBye.value }, to: pTo.value }, byeResult: bye.value };
    if (mode.value === 'points') rules.duel = { type: '', event: '', wonField: '', drawField: '', opponentField: '', winPoints: 0, drawPoints: 0 };
    var out = { key: key.value.trim(), name: name.read(), description: desc.read(), mode: mode.value, isActive: active.input.checked, schedule: sched.value.trim(), durationMinutes: intOr(dur, 0),
      signup: signup.value, signupMinutes: intOr(signupMin, 0), pairing: pairing.value, rules: rules };
    busy(save, true);
    api(creating ? 'POST' : 'PUT', '/clans/wars/configs' + (creating ? '' : '/' + c.id), out).then(function () {
      clWarConfigs = null; closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  sidePanel(creating ? t('wr.new') : textOf(c.name, c.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'war');
}

var CL_ROUND_COLOR = { signup: 'primary', active: 'warning', settled: 'success', cancelled: 'secondary' };

function roundTag(s) {
  return tag(t('rd.st.' + s), CL_ROUND_COLOR[s] || 'secondary');
}

function roundsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 25, total: 0, status: q.status || '', configKey: q.configKey || '' };
  var frame = pageFrame('clans/rounds', t('desc.clans/rounds'), [
    button({ label: t('rd.new'), variant: 'primary', data: { 'clan-action': 'new-round' }, onclick: function () { roundCreator(reload); } })]);
  clear(root).appendChild(frame.el);
  var status = select(state.status, [{ value: '', label: t('all') }].concat(opt('rd.st.', ['signup', 'active', 'settled', 'cancelled'])), { name: 'status' });
  var cfg = select(state.configKey, [{ value: '', label: t('all') }], { name: 'configKey' });
  status.addEventListener('change', function () { state.status = status.value; state.page = 1; reload(); });
  cfg.addEventListener('change', function () { state.configKey = cfg.value; state.page = 1; reload(); });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end cl-filters' }, field(t('f.state'), status), field(t('wr.config'), cfg)));
  frame.content.appendChild(listBox);
  loadWarConfigs(true).then(function (list) {
    list.forEach(function (c) { cfg.appendChild(h('option', { value: c.key, text: textOf(c.name, c.key) + ' · ' + c.key })); });
    cfg.value = state.configKey;
  }, function () {});
  function reload() {
    clear(listBox).appendChild(loading());
    Promise.all([api('GET', '/clans/wars/rounds' + qs({ status: state.status, configKey: state.configKey, page: state.page, limit: state.limit })), loadLanguages()]).then(function (res) {
      var d = res[0];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-clan-list': 'rounds' }, table([
        { label: t('f.name') }, { label: t('wr.mode') }, { label: t('f.state') }, { label: t('rd.starts') }, { label: t('rd.ends') }, { label: t('rd.wars') }, { label: t('rd.signups') }
      ], items.map(function (r) {
        return { id: r.id, item: r, cells: [
          nameCell(textOf(r.name, r.configKey), r.configKey),
          h('span', { text: t('mode.' + r.mode) }),
          roundTag(r.status),
          h('span', { class: 'text-xs', text: fmtDate(r.startsAt) }),
          h('span', { class: 'text-xs', text: fmtDate(r.endsAt) }),
          h('span', { text: fmtNum(r.wars || 0) }),
          h('span', { text: fmtNum(r.signups || 0) })
        ] };
      }), function (r) { roundPanel(r.id, reload); })));
      if (!items.length) listBox.appendChild(emptyCard(t('rd.empty')));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}

function roundCreator(done) {
  var body = h('div', { 'data-clan-form': 'round' }, loading());
  var create = button({ label: t('rd.create'), variant: 'primary', data: { 'clan-action': 'create-round' }, disabled: true });
  sidePanel(t('rd.new'), body, [button({ label: t('cancel'), onclick: closePanel }), create], 'new-round');
  loadWarConfigs(true).then(function (list) {
    clear(body);
    if (!list.length) { body.appendChild(notice('warning', t('rd.noConfigs'))); return; }
    var cfg = select(list[0].key, list.map(function (c) { return { value: c.key, label: textOf(c.name, c.key) + ' · ' + t('mode.' + c.mode) }; }), { name: 'configKey' });
    var starts = localInput(null, 'startsAt');
    var dur = numInput('', 'durationMinutes');
    var now = check(true, t('rd.startNow'), 'startNow');
    var msg = h('div', {});
    var startsField = field(t('rd.starts'), starts, t('rd.hint.starts'));
    function paint() { startsField.hidden = now.input.checked; }
    now.input.addEventListener('change', paint);
    paint();
    body.appendChild(h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('rd.hint') }));
    body.appendChild(field(t('wr.config'), cfg));
    body.appendChild(now);
    body.appendChild(startsField);
    body.appendChild(field(t('wr.duration'), dur, t('rd.hint.duration')));
    body.appendChild(msg);
    create.disabled = false;
    create.addEventListener('click', function () {
      clear(msg);
      var payload = { configKey: cfg.value, startNow: now.input.checked, durationMinutes: intOr(dur, 0) };
      if (!now.input.checked) {
        var at = starts.read();
        if (!at) { msg.appendChild(notice('error', t('rd.needStart'))); return; }
        payload.startsAt = at;
      }
      busy(create, true);
      api('POST', '/clans/wars/rounds', payload).then(function (out) { toast('success', t('rd.created')); done(); roundPanel(out.id, done, out); }, function (err) { busy(create, false); fieldError(msg, err); });
    });
  }, function (err) { clear(body).appendChild(notice('error', clErrText(err))); });
}

function roundPanel(id, done, first) {
  var body = h('div', { 'data-clan-round': id }, loading());
  var foot = h('div', { class: 'flex flex-wrap gap-2 justify-end w-full' });
  sidePanel(t('rd.detail'), body, [foot], 'round');
  function act(path, confirmText, okLabel, btn) {
    (confirmText ? confirmBox(confirmText, okLabel) : Promise.resolve(true)).then(function (ok) {
      if (!ok) return;
      busy(btn, true);
      api('POST', '/clans/wars/rounds/' + id + '/' + path, {}).then(function (out) { toast('success', t('rd.done.' + path)); paint(out); if (done) done(); }, function (err) { busy(btn, false); toast('error', clErrText(err)); });
    });
  }
  function paint(r) {
    clear(body);
    clear(foot);
    var open = r.status === 'signup' || r.status === 'active';
    if (r.status === 'signup') { var sb = button({ label: t('rd.start'), variant: 'primary', data: { 'clan-action': 'start-round' }, onclick: function () { act('start', null, null, sb); } }); foot.appendChild(sb); }
    if (r.status === 'active') { var st = button({ label: t('rd.settle'), variant: 'primary', data: { 'clan-action': 'settle-round' }, onclick: function () { act('settle', t('rd.settleConfirm'), t('rd.settle'), st); } }); foot.appendChild(st); }
    if (open) { var cb = button({ label: t('rd.cancel'), variant: 'danger', data: { 'clan-action': 'cancel-round' }, onclick: function () { act('cancel', t('rd.cancelConfirm'), t('rd.cancel'), cb); } }); foot.insertBefore(cb, foot.firstChild); }
    foot.appendChild(button({ label: t('close'), onclick: closePanel }));
    body.appendChild(h('div', { class: 'mb-4' }, dl([
      [t('f.name'), h('span', { dir: 'auto', text: textOf(r.name, r.configKey) })],
      [t('wr.config'), mono(r.configKey)],
      [t('f.state'), h('span', { 'data-clan-value': 'round-status' }, roundTag(r.status))],
      [t('wr.mode'), t('mode.' + r.mode)],
      [t('wr.signup'), t('wr.signup.' + r.signup)],
      [t('wr.pairing'), t('wr.pair.' + r.pairing)],
      r.signupAt ? [t('rd.signupAt'), fmtDate(r.signupAt)] : null,
      [t('rd.starts'), fmtDate(r.startsAt)],
      [t('rd.ends'), fmtDate(r.endsAt)],
      r.settledAt ? [t('rd.settledAt'), fmtDate(r.settledAt)] : null,
      [t('rd.actor'), h('span', { dir: 'auto', text: r.actor || t('ss.scheduler') })]
    ], 'round')));
    var wars = r.items || [];
    var duel = r.mode === 'duel';
    body.appendChild(section(t('rd.warsT', { n: fmtNum(wars.length) }), wars.length ? h('div', { class: 'cl-rows' }, wars.map(function (w) {
      var winA = w.result === 'a', winB = w.result === 'b';
      var res = w.status === 'active' ? tag(t('rd.st.active'), 'warning') : w.result === 'draw' ? tag(t('out.draw'), 'secondary') : w.result === 'bye' ? tag(t('out.bye'), 'secondary')
        : w.result ? tag(t('rd.winner', { name: winA ? w.nameA : w.nameB }), 'success') : tag(t('out.cancelled'), 'secondary');
      return h('div', { class: 'cl-war', 'data-clan-war': w.warId },
        h('div', { class: 'cl-war-side' + (winA ? ' cl-war-win' : '') }, h('span', { class: 'font-medium', dir: 'auto', text: w.nameA || w.clanA.slice(0, 8) }),
          h('span', { class: 'text-xs text-secondary-500', text: t('rd.members', { n: fmtNum(w.membersA) }) + (duel ? ' · ' + t('rd.wins', { n: fmtNum(w.winsA) }) : '') })),
        h('div', { class: 'cl-war-score', dir: 'ltr', 'data-clan-score': '', text: digits(String(w.scoreA)) + ' : ' + (w.clanB ? digits(String(w.scoreB)) : '–') }),
        h('div', { class: 'cl-war-side' + (winB ? ' cl-war-win' : '') }, w.clanB ? h('span', { class: 'font-medium', dir: 'auto', text: w.nameB || w.clanB.slice(0, 8) }) : h('span', { class: 'text-secondary-500', text: t('rd.noOpponent') }),
          w.clanB ? h('span', { class: 'text-xs text-secondary-500', text: t('rd.members', { n: fmtNum(w.membersB) }) + (duel ? ' · ' + t('rd.wins', { n: fmtNum(w.winsB) }) : '') }) : null),
        h('div', { class: 'cl-row-full flex flex-wrap gap-2 items-center' }, res, h('span', { class: 'text-xs text-secondary-500', text: t('rd.events', { n: fmtNum(w.events) }) })));
    })) : h('p', { class: 'text-sm text-secondary-500', text: r.status === 'signup' ? t('rd.notStarted') : t('rd.noWars') }), 'wars'));
    body.appendChild(h('div', { class: 'h-3' }));
    var su = r.signupList || [];
    body.appendChild(section(t('rd.signupsT', { n: fmtNum(su.length) }), su.length ? h('ul', { class: 'flex flex-col gap-1 text-sm' }, su.map(function (s) {
      return h('li', { class: 'flex justify-between gap-2' }, h('span', { dir: 'auto', text: s.name || s.clanId.slice(0, 8) }), h('span', { class: 'text-xs text-secondary-500', text: fmtDate(s.createdAt) }));
    })) : h('p', { class: 'text-sm text-secondary-500', text: r.signup === 'auto' ? t('rd.autoSignup') : t('rd.noSignups') }), 'signups'));
  }
  if (first) { loadLanguages().then(function () { paint(first); }); return; }
  Promise.all([api('GET', '/clans/wars/rounds/' + id), loadLanguages()]).then(function (res) { paint(res[0]); }, function (err) { clear(body).appendChild(notice('error', clErrText(err))); });
}

var CL_PERMS = ['accept', 'invite', 'kick', 'promote', 'edit', 'war'];

function costRows(values) {
  var box = h('div', { class: 'cl-rows', 'data-clan-cost': '' });
  var rows = [];
  function addRow(k, v) {
    var key = refSelect(refList('clans', 'currencies'), k || '', 'cost-currency', '—');
    var amt = numInput(v == null ? '' : v, 'cost-amount', t('cs.amount'));
    var rec = { key: key, amt: amt };
    var row = h('div', { class: 'cl-row' }, field(t('cs.currency'), key), field(t('cs.amount'), amt),
      h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r !== rec; }); } })));
    rows.push(rec);
    box.appendChild(row);
  }
  Object.keys(values || {}).forEach(function (k) { addRow(k, values[k]); });
  var wrap = h('div', {}, box, h('div', { class: 'mt-2' }, button({ label: '+ ' + t('cs.addCost'), data: { 'clan-action': 'add-cost' }, onclick: function () { addRow(); } })));
  wrap.read = function () {
    var out = {};
    rows.forEach(function (r) { var n = numberValue(r.amt); if (r.key.value && n) out[r.key.value] = n; });
    return out;
  };
  return wrap;
}

function xpSourceRows(list) {
  var box = h('div', { class: 'cl-rows', 'data-clan-xp-sources': '' });
  var rows = [];
  function addRow(s) {
    s = s || { type: 'game_event', event: '', amount: 1 };
    var type = select(s.type, opt('src.', ['economy_event', 'game_event']), { name: 'xp-type' });
    var ev = eventInput(eventsFor(s.type), s.event, 'xp-event', 'task_completed');
    var ctx = input(s.context, { name: 'xp-context', ltr: true, mono: true });
    var fld = input(s.field, { name: 'xp-field', ltr: true, mono: true });
    var eq = input(s.equals, { name: 'xp-equals', ltr: true, mono: true });
    var amt = numInput(s.amount, 'xp-amount', t('cs.xpAmount'));
    var per = check(s.perValue, t('cs.perValue'), 'xp-perValue');
    var rec = { type: type, ev: ev, ctx: ctx, fld: fld, eq: eq, amt: amt, per: per };
    var row = h('div', { class: 'cl-row', 'data-clan-xp-source': '' }, field(t('wr.source'), type), field(t('wr.event'), ev), field(t('cs.context'), ctx), field(t('wr.field'), fld),
      field(t('wr.equals'), eq), field(t('cs.xpAmount'), amt), h('div', { class: 'cl-row-full flex items-center justify-between gap-2' }, per,
        button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r !== rec; }); } })));
    rows.push(rec);
    box.appendChild(row);
  }
  (list || []).forEach(addRow);
  var wrap = h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('cs.hint.sources') }), box,
    h('div', { class: 'mt-2' }, button({ label: '+ ' + t('cs.addSource'), data: { 'clan-action': 'add-xp-source' }, onclick: function () { addRow(); } })));
  wrap.read = function () {
    return rows.map(function (r) {
      var o = { type: r.type.value, event: r.ev.input.value.trim(), amount: intOr(r.amt, 0) };
      ['ctx', 'fld', 'eq'].forEach(function (k, i) { var v = r[k].value.trim(); if (v) o[['context', 'field', 'equals'][i]] = v; });
      if (r.per.input.checked) o.perValue = true;
      return o;
    });
  };
  return wrap;
}

function permGrid(roles) {
  var boxes = {};
  var head = h('tr', {}, h('th', { class: 'p-2 text-start text-sm font-medium', text: t('cs.action') }), CL_ROLES.map(function (r) { return h('th', { class: 'p-2 text-center text-sm font-medium', text: t('role.' + r) }); }));
  var body = h('tbody', {});
  CL_PERMS.forEach(function (p) {
    boxes[p] = {};
    var tr = h('tr', { class: 'border-t border-secondary-200 dark:border-secondary-700', 'data-clan-perm': p }, h('td', { class: 'p-2 text-sm', 'data-label': t('cs.action') }, h('div', { class: 'flex flex-col' },
      h('span', { class: 'font-medium', text: t('cs.perm.' + p) }), h('span', { class: 'text-xs text-secondary-500', text: t('cs.permHint.' + p) }))));
    CL_ROLES.forEach(function (r) {
      var c = h('input', { type: 'checkbox', class: 'cl-check', name: 'roles.' + p + '.' + r, 'aria-label': t('cs.perm.' + p) + ' · ' + t('role.' + r), checked: (roles[p] || []).indexOf(r) !== -1 });
      boxes[p][r] = c;
      tr.appendChild(h('td', { class: 'p-2 text-center', 'data-label': t('role.' + r) }, c));
    });
    body.appendChild(tr);
  });
  var el = h('div', { class: 'overflow-x-auto' }, h('table', { class: 'w-full cl-perm-table', 'data-clan-perms': '' }, h('thead', {}, head), body));
  el.read = function () {
    var out = {};
    CL_PERMS.forEach(function (p) { out[p] = CL_ROLES.filter(function (r) { return boxes[p][r].checked; }); });
    return out;
  };
  return el;
}

function clanSettingsPage(root) {
  loadRefs('clans', true).then(function () {
    settingsPage(root, 'clans/settings', '/clans', function (cur) {
      var cr = cur.create, bd = cur.badge, ro = cur.roles, cd = cur.cooldowns, xp = cur.xp, ch = cur.chat, ds = cur.display, ev = cur.events;
      var enabled = check(cur.enabled, t('cs.enabled'), 'enabled');
      var crOn = check(cr.enabled, t('cs.createOn'), 'create.enabled');
      var cost = costRows(cr.cost);
      var crMin = numInput(cr.minTrophies, 'create.minTrophies');
      var nameMin = numInput(cr.nameMin, 'create.nameMin');
      var nameMax = numInput(cr.nameMax, 'create.nameMax');
      var descMax = numInput(cr.descriptionMax, 'create.descriptionMax');
      var defJoin = select(cr.defaultJoinType, opt('jt.', ['open', 'closed', 'invite']), { name: 'create.defaultJoinType' });
      var maxMin = numInput(cr.maxMinTrophies, 'create.maxMinTrophies');
      var emblems = tagsInput(bd.emblems, 'badge.emblems');
      var colors = tagsInput(bd.colors, 'badge.colors', '#E53935، #1E88E5');
      var frames = tagsInput(bd.frames, 'badge.frames');
      var perms = permGrid(ro);
      var maxCo = numInput(ro.maxCoLeaders, 'roles.maxCoLeaders');
      var leave = select(ro.leaderLeave, opt('cs.leave.', ['transfer', 'block']), { name: 'roles.leaderLeave' });
      var joinAfter = numInput(cd.joinAfterLeaveMinutes, 'cooldowns.joinAfterLeaveMinutes');
      var rejoin = numInput(cd.rejoinAfterKickHours, 'cooldowns.rejoinAfterKickHours');
      var leaveAfter = numInput(cd.leaveAfterJoinMinutes, 'cooldowns.leaveAfterJoinMinutes');
      var invDays = numInput(cd.inviteDays, 'cooldowns.inviteDays');
      var reqs = numInput(cd.requestsPerUser, 'cooldowns.requestsPerUser');
      var sources = xpSourceRows(xp.sources);
      var cap = numInput(xp.memberDailyCap, 'xp.memberDailyCap');
      var warWin = numInput(xp.warWin, 'xp.warWin');
      var warDraw = numInput(xp.warDraw, 'xp.warDraw');
      var warLoss = numInput(xp.warLoss, 'xp.warLoss');
      var sys = check(ch.systemMessages, t('cs.system'), 'chat.systemMessages');
      var sLimit = numInput(ds.searchLimit, 'display.searchLimit');
      var hLimit = numInput(ds.historyLimit, 'display.historyLimit');
      var pub = check(ev.publish, t('ts.publish'), 'events.publish');
      var notify = check(ev.notify, t('cs.notify'), 'events.notify');
      var code = numInput(ev.code, 'events.code');
      return {
        el: [
          section(t('ts.general'), h('div', {}, enabled, h('p', { class: 'text-xs text-secondary-500', text: t('cs.hint.enabled') })), 'general'),
          section(t('cs.create'), h('div', {}, crOn, field(t('cs.cost'), cost, t('cs.hint.cost')),
            grid3(field(t('cs.minTrophies'), crMin, t('cs.hint.minTrophies')), field(t('cs.nameMin'), nameMin), field(t('cs.nameMax'), nameMax)),
            grid3(field(t('cs.descMax'), descMax), field(t('cs.defJoin'), defJoin), field(t('cs.maxMin'), maxMin, t('cs.hint.maxMin')))), 'create'),
          section(t('cs.badge'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('cs.hint.badge') }), field(t('cl.emblem'), emblems), field(t('cs.colors'), colors), field(t('cl.frame'), frames)), 'badge'),
          section(t('cs.roles'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('cs.hint.roles') }), perms, h('div', { class: 'h-3' }),
            grid(field(t('cs.maxCo'), maxCo), field(t('cs.leaderLeave'), leave, t('cs.hint.leaderLeave')))), 'roles'),
          section(t('cs.cooldowns'), h('div', {}, grid(field(t('cs.joinAfter'), joinAfter, t('cs.hint.joinAfter')), field(t('cs.rejoin'), rejoin, t('cs.hint.rejoin'))),
            grid3(field(t('cs.leaveAfter'), leaveAfter, t('cs.hint.leaveAfter')), field(t('cs.invDays'), invDays), field(t('cs.reqs'), reqs, t('cs.hint.reqs')))), 'cooldowns'),
          section(t('cs.xp'), h('div', {}, sources, h('div', { class: 'h-3' }), grid(field(t('cs.cap'), cap, t('cs.hint.cap')), h('div', {})),
            grid3(field(t('cs.warWin'), warWin), field(t('cs.warDraw'), warDraw), field(t('cs.warLoss'), warLoss))), 'xp'),
          section(t('cs.chat'), h('div', {}, sys, h('p', { class: 'text-xs text-secondary-500 -mt-2', text: t('cs.hint.system') })), 'chat'),
          section(t('ts.display'), grid(field(t('cs.sLimit'), sLimit, t('cs.hint.sLimit')), field(t('cs.hLimit'), hLimit, t('cs.hint.hLimit'))), 'display'),
          section(t('ts.events'), h('div', {}, pub, notify, field(t('cs.code'), code, t('cs.hint.code'), 'cl-num')), 'events')],
        read: function (c) {
          c.enabled = enabled.input.checked;
          c.create = { enabled: crOn.input.checked, cost: cost.read(), minTrophies: intOr(crMin, 0), nameMin: intOr(nameMin, 0), nameMax: intOr(nameMax, 0), descriptionMax: intOr(descMax, 0),
            defaultJoinType: defJoin.value, maxMinTrophies: intOr(maxMin, 0) };
          c.badge = { emblems: emblems.read(), colors: colors.read(), frames: frames.read() };
          var p = perms.read();
          c.roles = { accept: p.accept, invite: p.invite, kick: p.kick, promote: p.promote, edit: p.edit, war: p.war, maxCoLeaders: intOr(maxCo, 0), leaderLeave: leave.value };
          c.cooldowns = { joinAfterLeaveMinutes: intOr(joinAfter, 0), rejoinAfterKickHours: intOr(rejoin, 0), leaveAfterJoinMinutes: intOr(leaveAfter, 0), inviteDays: intOr(invDays, 0), requestsPerUser: intOr(reqs, 0) };
          c.xp = { sources: sources.read(), memberDailyCap: intOr(cap, 0), warWin: intOr(warWin, 0), warDraw: intOr(warDraw, 0), warLoss: intOr(warLoss, 0) };
          c.chat = { systemMessages: sys.input.checked };
          c.display = { searchLimit: intOr(sLimit, 0), historyLimit: intOr(hLimit, 0) };
          c.events = { publish: pub.input.checked, notify: notify.input.checked, code: intOr(code, 0) };
          return null;
        },
        after: function (out) { clClanSettings = out.settings; }
      };
    });
  });
}

var CL_ROOT = 'clan-root';
var CL_GROUPS = { trophies: ['leagues', 'players', 'seasons', 'settings'], clans: ['list', 'levels', 'wars', 'rounds', 'settings'] };
var CL_ORDER = ['trophies', 'clans'];
var CL_RENDER = { 'trophies/leagues': leaguesPage, 'trophies/players': trPlayersPage, 'trophies/seasons': seasonsPage, 'trophies/settings': trophySettingsPage,
  'clans/list': clansPage, 'clans/levels': levelsPage, 'clans/wars': warsPage, 'clans/rounds': roundsPage, 'clans/settings': clanSettingsPage };
var CL_HASH = /^#\/(trophies|clans)(?:\/([^/?]*))?/;
var clPending = null;
var clRouter = null;
var clMounted = null;
var clScheduled = false;

function clRoute(hash) {
  var m = CL_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  var pages = CL_GROUPS[m[1]];
  return m[1] + '/' + (pages.indexOf(m[2]) !== -1 ? m[2] : pages[0]);
}

function clAllowed() {
  return !!token();
}

function clTitle() {
  var page = clRoute();
  return page ? t('group.' + page.split('/')[0]) + ' - ' + t('nav.' + page) : '';
}

function clAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || clRouter) return !!clRouter;
  try {
    var layout = r.resolve('/').matched[0];
    CL_ORDER.forEach(function (g) {
      if (r.hasRoute(g)) return;
      r.addRoute({ path: '/' + g + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: g, component: { name: g === 'clans' ? 'ClansPage' : 'TrophiesPage', render: function () { return null; } }, meta: { get title() { return clTitle(); } } }] });
    });
    clRouter = r;
  } catch (e) {
    return false;
  }
  if (clPending) {
    var target = clPending;
    clPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function clNavLink(g, page, active) {
  var id = g + '/' + page;
  var on = id === active;
  return h('li', {}, h('a', { href: '#/' + id, 'data-clan-nav': id, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + id) }))));
}

function clSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelectorAll('[data-clan-group]');
  if (!group || !clAllowed()) { Array.prototype.forEach.call(existing, function (e) { e.remove(); }); return; }
  var target = document.querySelector('.x-drawer [data-quiz-group]') || document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = clLang() + '|' + active;
  var first = document.querySelector('[data-clan-group="trophies"]');
  var second = document.querySelector('[data-clan-group="clans"]');
  if (first && second && first.getAttribute('data-key') === key && second.getAttribute('data-key') === key && first.nextElementSibling === second && second.nextElementSibling === target) return;
  Array.prototype.forEach.call(existing, function (e) { e.remove(); });
  CL_ORDER.forEach(function (g) {
    target.before(h('div', { 'data-clan-group': g, 'data-plane-group': g, 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
      h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group.' + g) }),
      h('nav', { 'aria-label': t('group.' + g) }, h('ul', { class: 'space-y-0.5' }, CL_GROUPS[g].map(function (p) { return clNavLink(g, p, active); })))));
  });
}

function clShowMains() {
  var hidden = document.querySelectorAll('[data-clan-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-clan-hidden'); }
}

function clUnmount() {
  var root = document.getElementById(CL_ROOT);
  if (root) root.remove();
  clShowMains();
  closePanel();
  clMounted = null;
}

function clSync() {
  clAddRoutes();
  var page = clRoute();
  clSyncNav(page);
  if (!page) { if (clMounted || document.getElementById(CL_ROOT)) clUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-clan-hidden', ''); }
  }
  var root = document.getElementById(CL_ROOT);
  if (!root) {
    root = h('div', { id: CL_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-clan': 'root' });
    clMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + clLang() + (page === 'clans/rounds' ? '|' + location.hash : '');
  if (clMounted !== key) {
    clMounted = key;
    closePanel();
    CL_RENDER[page](root);
  }
}

function clSchedule() {
  if (clScheduled) return;
  clScheduled = true;
  requestAnimationFrame(function () { clScheduled = false; clSync(); });
}

function clBoot() {
  window.addEventListener('hashchange', clSchedule);
  window.addEventListener('popstate', clSchedule);
  window.addEventListener('nakama-console:ready', clSchedule);
  new MutationObserver(clSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { clMounted = null; clSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (clAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  clSync();
}

window.NakamaClan = { version: 1, pages: function () {
  var out = [];
  CL_ORDER.forEach(function (g) { CL_GROUPS[g].forEach(function (p) { out.push({ id: g + '/' + p, group: g, href: '#/' + g + '/' + p, label: t('nav.' + g + '/' + p) }); }); });
  return out;
} };

if (CL_HASH.test(location.hash)) {
  clPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', clBoot);
else clBoot();
})();
