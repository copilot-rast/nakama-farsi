"use strict";

const SERVER = (() => {
  const q = new URLSearchParams(location.search);
  for (const name of ["api", "key"]) {
    const v = q.get(name);
    if (v === "reset") localStorage.removeItem("stargame." + name);
    else if (v) localStorage.setItem("stargame." + name, v);
  }
  const served = location.pathname === "/stargame" || location.pathname.startsWith("/stargame/");
  const fallback = served ? location.origin : `${location.protocol}//${location.hostname}:7450`;
  const api = (localStorage.getItem("stargame.api") || fallback).replace(/\/+$/, "");
  return { api, key: localStorage.getItem("stargame.key") || "", fromConfig: served && api === location.origin };
})();
const API = SERVER.api;
let SERVER_KEY = SERVER.key;

async function loadServerKey() {
  if (SERVER_KEY) return;
  if (SERVER.fromConfig) {
    try {
      const res = await fetch("/stargame/config.json", { cache: "no-store" });
      if (res.ok) SERVER_KEY = (await res.json()).serverKey || "";
    } catch {}
  }
  if (!SERVER_KEY) SERVER_KEY = "defaultkey";
}
const ROUND_SECONDS = 30;
const CHAT_ROOM = (() => {
  const r = new URLSearchParams(location.search).get("room") || "";
  return /^[a-z0-9-]{1,40}$/.test(r) ? r : "lobby";
})();
const GAME_VERSION = "1.4.0";
const PLATFORM = "web";
const LS = {
  device: "stargame.device",
  skin: "stargame.skin",
  storeToken: "stargame.lastStoreToken",
  lang: "stargame.lang",
  langDir: "stargame.langDir",
  loginKey: "stargame.loginKey",
  skipUpdate: "stargame.skipUpdate",
};

const STRINGS = {
  fa: {
    appName: "ستاره‌چین",
    home: "خانه",
    play: "بازی",
    shop: "فروشگاه",
    gift: "کد هدیه",
    payments: "پرداخت‌های من",
    news: "اخبار",
    faq: "سؤال‌های متداول",
    inbox: "صندوق پیام",
    coinUnit: "سکه",
    language: "زبان",
    loggingIn: "در حال ورود…",
    retry: "دوباره امتحان کن",
    hello: "سلام",
    heroTitle: "ستاره‌ها را بگیر، از بمب‌ها فرار کن",
    heroText: () => `هر دور ${num(30)} ثانیه است. هر ستاره <b>${num(1)} امتیاز</b> دارد و ستارهٔ طلایی <b>${num(5)} امتیاز</b>. هر بمب <b>${num(3)} امتیاز</b> کم می‌کند. هر ${num(5)} امتیاز یک سکه، تا سقف ${num(30)} سکه در هر دور.`,
    startGame: "شروع بازی",
    mySkins: "سبدهای من",
    mySkinsHint: "سبدی را که می‌خواهی در بازی داشته باشی انتخاب کن",
    score: "امتیاز",
    time: "زمان",
    quit: "خروج",
    canvasLabel: "صفحهٔ بازی",
    playHint: "با موس، لمس یا کلیدهای جهت‌نما سبد را حرکت بده.",
    back: "بازگشت",
    restore: "بازیابی خرید",
    discountLabel: "کد تخفیف (اختیاری، فقط زرین‌پال و زیبال)",
    giftPrompt: "کد هدیه‌ات را وارد کن",
    giftSubmit: "دریافت هدیه",
    more: "بیشتر",
    olderNews: "خبرهای قدیمی‌تر",
    allNews: "همهٔ خبرها",
    comments: "نظرها",
    commentLabel: "نظرت را بنویس",
    commentPlaceholder: "نظرت را بنویس…",
    sendComment: "ارسال نظر",
    moreComments: "نظرهای بیشتر",
    faqSearch: "جستجو در سؤال‌ها",
    faqSearchPlaceholder: "جستجو در سؤال‌ها…",
    seenAll: "همه را خواندم",
    olderMessages: "پیام‌های قدیمی‌تر",
    player: "بازیکن",
    skinSelected: (name) => `${name} انتخاب شد`,
    selected: "انتخاب‌شده",
    moreSkins: (n) => `${num(n)} سبد دیگر`,
    inShop: "در فروشگاه",
    preparingRound: "آماده‌سازی دور…",
    startFailed: "شروع نشد",
    rewards: "جایزه‌ها",
    rewardsHint: "هدف‌های روزانه و رکوردها؛ با هر کدام جایزه بگیر.",
    rewardsEmpty: "فعلاً جایزه‌ای تعریف نشده.",
    levelTitle: (n, title) => `سطح ${num(n)}` + (title ? ` · ${title}` : ""),
    levelXp: (v, next, name) => `${num(v)} از ${num(next)} ${name}`,
    levelMax: (v, name) => `${num(v)} ${name} · بالاترین سطح`,
    claimLevel: (n, prize) => `جایزهٔ سطح ${num(n)} را بگیر: ${prize}`,
    streakDay: (n) => `روز ${num(n)}`,
    streakInfo: (count, best) => `${num(count)} روز پشت‌سرهم · بهترین: ${num(best)}`,
    streakToday: "امروز شمرده شد؛ فردا برگرد تا روز بعد را بگیری.",
    streakStart: "امروز وارد شدی؛ جایزهٔ امروزت آماده است.",
    streakBroken: "چند روز نیامدی؛ از روز اول شروع کن.",
    streakGrace: (n) => `اگر ${num(n)} روز جا بماند، روزهایت از دست نمی‌رود.`,
    streakSafe: (n) => `تا روز ${num(n)} امن است`,
    streakNextSafe: (n, left) => `نقطهٔ امن بعدی: روز ${num(n)} (${num(left)} روز دیگر)`,
    streakFloored: (n) => `چند روز نیامدی، ولی به روز ${num(n)} برگشتی، نه روز اول.`,
    streakSafePrize: (n, prize) => `روز ${num(n)}: ${prize}`,
    claim: "بگیر",
    claimed: "گرفتی",
    locked: "بعداً",
    todayLabel: "امروز",
    pendingLabel: "در راه",
    rewardClaimed: (text) => `جایزه گرفتی: ${text}`,
    rewardDone: "انجام شد",
    rewardAuto: "خودکار می‌رسد",
    periodDay: "امروز",
    periodWeek: "این هفته",
    progressOf: (c, target) => `${num(c)} از ${num(target)}`,
    levelUp: (n) => `سطحت بالا رفت: سطح ${num(n)}!`,
    newRewards: (n) => `${num(n)} جایزهٔ تازه منتظر توست.`,
    autoRewards: (n) => `${num(n)} جایزه خودکار به حسابت آمد.`,
    playAgain: "دوباره بازی",
    savingScore: "در حال ثبت امتیاز…",
    roundCoins: (c, b) => `${num(c)} سکه گرفتی. موجودی: ${num(b)} سکه`,
    roundNoCoins: (b) => `این دور سکه‌ای نگرفتی. موجودی: ${num(b)} سکه`,
    roundOver: "دور تمام شد!",
    scoreNotSaved: "امتیاز ثبت نشد",
    coins: (n) => `${num(n)} سکه`,
    toman: (n) => `${num(n)} تومان`,
    free: "رایگان",
    listSep: "، ",
    emptyShop: "فعلاً چیزی برای فروش نیست.",
    owned: "مال توست",
    useInGame: "انتخاب برای بازی",
    buyWithCoins: "خرید با سکه",
    confirmBuy: (name, n) => `«${name}» را با ${num(n)} سکه می‌خری؟`,
    buy: "بخر",
    goingTo: (m) => `در حال رفتن به ${m}…`,
    buyOk: (r) => `خرید موفق! ${r}`,
    cancel: "انصراف",
    simPay: "پرداخت موفق",
    confirmingPurchase: "در حال تأیید خرید…",
    simBadge: "شبیه‌ساز فروشگاه (حالت تست)",
    simNote: "روی وب برنامهٔ فروشگاه نداریم؛ این پنجره جای آن را می‌گیرد.",
    product: "محصول",
    amount: "مبلغ",
    payment: "پرداخت",
    simTokenLabel: "توکن خرید آزمایشی (برای تست می‌توانی یک توکن قدیمی را جایگزین کنی)",
    checkRestore: "بررسی و بازیابی",
    checking: "در حال بررسی…",
    restoreOk: (r) => `خرید تأیید شد${r ? ": " + r : ""}. اگر قبلاً جایزه را گرفته بودی، دوباره داده نمی‌شود.`,
    statusIs: (s) => `وضعیت: ${s}`,
    restoreTitle: "بازیابی خرید فروشگاه",
    restoreHelp: "اگر پولت از بازار یا مایکت کم شد ولی سکه نرسید، توکن خرید را اینجا بفرست.",
    store: "فروشگاه",
    purchaseToken: "توکن خرید",
    close: "بستن",
    giftGot: (r) => `هدیه گرفتی: ${r}`,
    noPayments: "هنوز پرداختی نداری.",
    refId: "کد پیگیری",
    checkingPayment: "در حال بررسی پرداخت…",
    payOk: "پرداخت موفق بود",
    payPending: "پرداخت هنوز در انتظار است",
    payFailed: "پرداخت انجام نشد",
    payOkText: (b) => `سکه‌ها به حسابت اضافه شد. موجودی: ${num(b)} سکه`,
    payPendingText: "چند دقیقهٔ دیگر در «پرداخت‌های من» نگاه کن.",
    payFailedText: "پولی کم نشده باشد، برمی‌گردد.",
    ok: "باشه",
    sectionNotFound: "این بخش در بازی پیدا نشد.",
    newsAria: (n) => (n ? `اخبار، ${num(n)} خبر تازه` : "اخبار"),
    noNews: "هنوز خبری منتشر نشده.",
    pinned: "سنجاق‌شده",
    fresh: "تازه",
    commentCount: (n) => `${num(n)} نظر`,
    commentsClosed: "نظر دادن برای این خبر بسته است",
    you: "تو",
    awaitingApproval: "در انتظار تأیید",
    reply: "پاسخ",
    replyingTo: (name) => `در پاسخ به ${name} `,
    noCommentsYet: "هنوز نظری نیست. اولین نفر باش!",
    noComments: "نظری نیست.",
    commentLive: "نظرت منتشر شد.",
    commentQueued: "نظرت ثبت شد و بعد از تأیید برای بقیه نمایش داده می‌شود.",
    all: "همه",
    noFaqMatch: "با این جستجو سؤالی پیدا نشد.",
    noFaq: "هنوز سؤالی اضافه نشده.",
    inboxAria: (n) => (n ? `صندوق پیام، ${num(n)} پیام نخوانده` : "صندوق پیام"),
    unread: "نخوانده",
    message: "پیام",
    openLink: "باز کردن پیوند ↗",
    goTo: (s) => `رفتن به ${s}`,
    inboxEmpty: "صندوقت خالی است.",
    allRead: "همهٔ پیام‌ها خوانده شد.",
    newMessage: "پیام تازه",
    account: "حساب من",
    guest: "مهمان",
    guestHint: "پیشرفتت فقط روی همین دستگاه است. با شماره موبایل وارد شو تا روی هر دستگاهی داشته باشی‌اش.",
    signedInAs: (p) => `با شمارهٔ ${p} وارد شده‌ای`,
    signedInHint: "با همین شماره روی هر دستگاهی وارد شو.",
    myCountry: "کشور من",
    myCountryHint: "برای نشان دادن پرچم و پیش‌شماره",
    chooseCountry: "انتخاب کشور",
    searchCountry: "جستجوی کشور…",
    noCountry: "کشوری پیدا نشد.",
    countrySaved: (n) => `کشورت شد ${n}`,
    otpTitle: "ورود با شماره موبایل",
    otpHint: "کد یک‌بار مصرف برایت پیامک می‌شود.",
    phone: "شماره موبایل",
    otpKeep: "پیشرفت این دستگاه را نگه دار",
    sendCode: "کد را بفرست",
    sending: "در حال فرستادن…",
    codeSent: (p) => `کد به ${p} فرستاده شد.`,
    codeSandbox: "حالت آزمایشی: پیامکی نمی‌آید؛ کد در لاگ سرور است.",
    code: "کد",
    codeDigits: (n) => `کد ${num(n)} رقمی`,
    verify: "ورود",
    verifying: "در حال بررسی…",
    resendIn: (n) => `ارسال دوباره تا ${num(n)} ثانیه`,
    resend: "ارسال دوباره",
    changeNumber: "شمارهٔ دیگر",
    welcomeBack: (p) => `خوش آمدی! با ${p} وارد شدی.`,
    newAccount: (p) => `حساب تازه با ${p} ساخته شد.`,
    linked: (p) => `شمارهٔ ${p} به همین حساب وصل شد.`,
    signOut: "خروج از حساب موبایل",
    signedOut: "از حساب موبایل خارج شدی. حالا مهمانی.",
    username: "نام کاربری",
    saveName: "ذخیره",
    nameSaved: "نام کاربری ذخیره شد.",
    nameRule: (min, max, latin) => `بین ${num(min)} تا ${num(max)} حرف${latin ? "، فقط حرف لاتین، رقم و _ . -" : ""}.`,
    updateAvailable: "نسخهٔ تازه آمده",
    updateRequired: "باید بازی را به‌روز کنی",
    updateNow: "به‌روزرسانی",
    later: "بعداً",
    updateDefault: "برای ادامه، بازی را به‌روز کن.",
    versions: (cur, min) => `نسخهٔ تو: ${cur} — کمترین نسخهٔ مجاز: ${min}`,
    versionLabel: (v) => `نسخهٔ ${v}`,
    errorCode: (code) => `خطا: ${code}`,
    methods: { zarinpal: "زرین‌پال", zibal: "زیبال", bazaar: "کافه‌بازار", myket: "مایکت", in_app: "پرداخت با سکه" },
    statuses: { success: "موفق", pending: "در انتظار", failed: "ناموفق", expired: "منقضی" },
    skins: { default: "سبد ساده", skin_gold: "سبد طلایی", skin_neon: "سبد نئونی", skin_galaxy: "سبد کهکشانی" },
    chat: "گفتگو",
    chatTitle: "گفتگوی بازیکنان",
    chatPlaceholder: "پیامت را بنویس",
    send: "بفرست",
    chatRules: "کلمه‌های زشت با ستاره پنهان می‌شوند. اگر پیامی آزارت داد، گزارشش کن.",
    chatEmpty: "هنوز پیامی نیست. اولین نفر باش!",
    chatJoining: "در حال ورود به اتاق گفتگو…",
    chatOffline: "به گفتگو وصل نیستیم. چند لحظه بعد دوباره امتحان کن.",
    chatYou: "تو",
    report: "گزارش",
    reportMessage: (name) => `گزارش پیام ${name}`,
    reportTitle: "چرا این پیام را گزارش می‌کنی؟",
    reportNote: "توضیح (اختیاری)",
    reportSend: "فرستادن گزارش",
    reportDone: "گزارشت ثبت شد. ممنونیم!",
    reportDuplicate: "این پیام را قبلاً گزارش کرده‌ای.",
    reportPick: "یک دلیل انتخاب کن.",
    mutedUntil: (d) => `تا ${d} نمی‌توانی در گفتگو پیام بفرستی.`,
    mutedTitle: "فعلاً بی‌صدا هستی",
    warnedTitle: "اخطار",
    liftedMute: "حالا دوباره می‌توانی پیام بفرستی.",
    suspendedTitle: "حسابت موقتاً بسته است",
    suspendedUntil: (d) => `تا ${d} نمی‌توانی وارد بازی شوی.`,
    myTags: "برچسب‌های من",
    myTagsHint: "برچسب‌هایی که بازی به تو داده",
    exclusive: "ویژهٔ تو",
    messageRemoved: "این پیام پاک شد.",
    errors: {
      CLAIM_NOT_FOUND: "این جایزه پیدا نشد.",
      ALREADY_CLAIMED: "این جایزه را قبلاً گرفته‌ای.",
      TOO_MANY_EVENTS: "کمی آرام‌تر؛ یک دقیقهٔ دیگر دوباره امتحان کن.",
      CHAT_BLOCKED: "این پیام کلمهٔ ممنوع داشت و فرستاده نشد.",
      CHAT_MUTED: "فعلاً نمی‌توانی پیام بفرستی.",
      ACCOUNT_SUSPENDED: "حسابت موقتاً بسته است.",
      REPORT_TITLE_NOT_FOUND: "یک دلیل انتخاب کن.",
      CANNOT_REPORT_OWN: "پیام خودت را نمی‌توانی گزارش کنی.",
      MESSAGE_NOT_FOUND: "این پیام دیگر وجود ندارد.",
      TOO_MANY_REPORTS: "گزارش زیادی فرستادی؛ کمی بعد دوباره امتحان کن.",
      TEXT_TOO_LONG: "متن خیلی بلند است.",
      OTP_DISABLED: "ورود با پیامک فعلاً خاموش است.",
      OTP_UNAVAILABLE: "سرویس پیامک آماده نیست. کمی بعد امتحان کن.",
      OTP_SEND_FAILED: "پیامک فرستاده نشد. دوباره امتحان کن.",
      OTP_TOO_SOON: "کمی صبر کن و بعد کد تازه بخواه.",
      OTP_RATE_LIMITED: "برای این شماره کد زیادی خواسته شده. بعداً امتحان کن.",
      OTP_NOT_FOUND: "اول کد بخواه.",
      OTP_EXPIRED: "کد منقضی شد. کد تازه بخواه.",
      OTP_INVALID_CODE: "کد درست نیست.",
      OTP_TOO_MANY_ATTEMPTS: "چند بار کد اشتباه زدی. کد تازه بخواه.",
      INVALID_PHONE: "شمارهٔ موبایل درست نیست.",
      PHONE_COUNTRY_NOT_ALLOWED: "شماره‌های این کشور پذیرفته نمی‌شوند.",
      ACCOUNT_DISABLED: "این حساب بسته شده است.",
      ACCOUNT_HAS_PHONE: "این حساب شمارهٔ دیگری دارد.",
      COUNTRY_NOT_AVAILABLE: "این کشور در دسترس نیست.",
      USERNAME_TOO_SHORT: "نام کاربری کوتاه است.",
      USERNAME_TOO_LONG: "نام کاربری بلند است.",
      USERNAME_INVALID_CHARACTERS: "نام کاربری فقط حرف لاتین، رقم و _ . - می‌تواند داشته باشد.",
      USERNAME_TAKEN: "این نام کاربری را کس دیگری دارد.",
      PASSWORD_TOO_SHORT: "رمز کوتاه است.",
      PASSWORD_NEEDS_LETTER: "رمز باید حرف داشته باشد.",
      PASSWORD_NEEDS_DIGIT: "رمز باید عدد داشته باشد.",
      DISPLAY_NAME_TOO_LONG: "نام نمایشی بلند است.",
      UNAUTHENTICATED: "نشست تمام شده؛ دوباره وارد شو.",
      REQUEST_TOO_LARGE: "درخواست خیلی بزرگ است.",
      INVALID_JSON: "درخواست نامعتبر است.",
      INVALID_LANGUAGE: "زبان نامعتبر است.",
      LANGUAGE_NOT_AVAILABLE: "این زبان در دسترس نیست.",
      INVALID_CURSOR: "صفحه‌بندی نامعتبر است؛ دوباره باز کن.",
      INVALID_METHOD: "روش پرداخت نامعتبر است.",
      INVALID_TOKEN: "توکن خرید خالی یا نامعتبر است.",
      INVALID_ID: "شناسه نامعتبر است.",
      INTERNAL_ERROR: "خطای سرور. کمی بعد دوباره امتحان کن.",
      TIMEOUT: "پاسخ سرور طول کشید. دوباره امتحان کن.",
      CANCELED: "درخواست لغو شد.",
      PRODUCT_NOT_FOUND: "این محصول پیدا نشد یا دیگر فروخته نمی‌شود.",
      PRIZE_NOT_FOUND: "جایزهٔ این مورد پیدا نشد.",
      PAYMENT_NOT_FOUND: "این پرداخت پیدا نشد.",
      METHOD_NOT_ALLOWED: "این محصول با این روش خریدنی نیست.",
      METHOD_UNAVAILABLE: "این روش پرداخت الان در دسترس نیست.",
      LEVEL_TOO_LOW: "سطحت برای این محصول کافی نیست.",
      PRIZE_ALREADY_CLAIMED: "این را قبلاً گرفته‌ای.",
      PRIZE_OWNER_MISMATCH: "این جایزه مال حساب دیگری است.",
      GIFT_NOT_APPLICABLE: "این کد تخفیف اینجا قابل استفاده نیست.",
      GIFT_NOT_FOUND: "این کد وجود ندارد.",
      GIFT_INACTIVE: "این کد فعال نیست.",
      GIFT_NOT_STARTED: "زمان استفاده از این کد هنوز نرسیده.",
      GIFT_EXPIRED: "این کد منقضی شده.",
      GIFT_ALREADY_USED: "این کد را قبلاً استفاده کرده‌ای.",
      GIFT_EXHAUSTED: "ظرفیت این کد تمام شده.",
      GIFT_NOT_REDEEMABLE: "این کد تخفیف است؛ موقع خرید با درگاه واردش کن.",
      PRICE_TOO_LOW: "با این تخفیف، مبلغ نهایی کمتر از حد مجاز می‌شود.",
      INSUFFICIENT_BALANCE: "سکه‌ات کافی نیست. چند دور بازی کن یا بستهٔ سکه بخر.",
      SKU_REQUIRED: "این محصول در این فروشگاه کد ندارد.",
      GATEWAY_REQUEST_FAILED: "درگاه پرداخت جواب نداد. دوباره امتحان کن.",
      GATEWAY_UNAVAILABLE: "درگاه پرداخت در دسترس نیست.",
      GATEWAY_REJECTED: "درگاه پرداخت را رد کرد.",
      REFERENCE_NOT_SAVED: "ثبت پرداخت کامل نشد. دوباره امتحان کن.",
      PAYMENT_ALREADY_SETTLED: "این پرداخت قبلاً تمام شده.",
      PAYMENT_CLOSED: "این پرداخت بسته شده؛ یک خرید تازه شروع کن.",
      PAYMENT_NOT_SUCCESSFUL: "پرداخت موفق نبود.",
      STORE_CONFIGURATION_REJECTED: "تنظیمات فروشگاه روی سرور رد شد؛ پرداخت در انتظار ماند.",
      STORE_UNAVAILABLE: "فروشگاه در دسترس نیست؛ کمی بعد دوباره امتحان کن.",
      PURCHASE_INVALID: "فروشگاه این خرید را تأیید نکرد.",
      SANDBOX_TOKEN_INVALID: "توکن آزمایشی نامعتبر است.",
      DEVELOPER_PAYLOAD_MISMATCH: "این توکن مال این پرداخت یا این حساب نیست.",
      TOKEN_ALREADY_USED: "این توکن خرید قبلاً استفاده شده.",
      TOKEN_NOT_VERIFIED: "فروشگاه مدت اشتراک را برنگرداند.",
      SUBSCRIPTION_OWNED_BY_ANOTHER_USER: "این اشتراک مال حساب دیگری است.",
      SUBSCRIPTION_NOT_SUPPORTED: "این فروشگاه اشتراک ندارد.",
      REWARD_PENDING: "پرداخت موفق بود ولی جایزه هنوز ثبت نشده؛ پشتیبانی بررسی می‌کند.",
      AMOUNT_MISMATCH: "مبلغ پرداخت با سفارش یکی نبود.",
      AMOUNT_UNKNOWN: "مبلغ پرداخت مشخص نشد.",
      NOT_VERIFIED: "پرداخت تأیید نشد.",
      CANCELED_BY_USER: "پرداخت لغو شد.",
      EXPIRED: "مهلت پرداخت تمام شد.",
      TIMED_OUT: "مهلت پرداخت تمام شد.",
      REFUNDED: "این پرداخت برگشت خورده.",
      ROUND_NOT_FOUND: "این دور بازی پیدا نشد.",
      ROUND_ALREADY_FINISHED: "این دور قبلاً ثبت شده.",
      ROUND_TOO_SHORT: "دور خیلی زود تمام شد و ثبت نشد.",
      ROUND_EXPIRED: "این دور خیلی قدیمی شده و ثبت نشد.",
      SCORE_TOO_HIGH: "امتیاز باورکردنی نیست و ثبت نشد.",
      NETWORK: "به سرور وصل نشدیم. اینترنت یا آدرس را بررسی کن.",
      NEWS_NOT_FOUND: "این خبر پیدا نشد یا دیگر منتشر نمی‌شود.",
      COMMENT_NOT_ALLOWED: "نظر دادن برای این خبر بسته است.",
      COMMENT_RATE_LIMITED: () => `کمی صبر کن؛ در هر دقیقه فقط ${num(5)} نظر می‌شود فرستاد.`,
      COMMENT_MESSAGE_REQUIRED: "اول نظرت را بنویس.",
      COMMENT_TOO_LONG: () => `نظر حداکثر ${num(1000)} نویسه است.`,
      PARENT_COMMENT_NOT_FOUND: "نظری که به آن جواب می‌دهی دیگر نیست.",
      QUERY_TOO_LONG: "متن جستجو خیلی بلند است.",
    },
  },
  en: {
    appName: "Star Catcher",
    home: "Home",
    play: "Game",
    shop: "Shop",
    gift: "Gift code",
    payments: "My payments",
    news: "News",
    faq: "FAQ",
    inbox: "Inbox",
    coinUnit: "coins",
    language: "Language",
    loggingIn: "Signing in…",
    retry: "Try again",
    hello: "Hi",
    heroTitle: "Catch the stars, dodge the bombs",
    heroText: () => `Each round lasts ${num(30)} seconds. A star is worth <b>${num(1)} point</b> and a golden star <b>${num(5)} points</b>. A bomb takes away <b>${num(3)} points</b>. Every ${num(5)} points is one coin, up to ${num(30)} coins per round.`,
    startGame: "Play",
    mySkins: "My baskets",
    mySkinsHint: "Pick the basket you want to play with",
    score: "Score",
    time: "Time",
    quit: "Quit",
    canvasLabel: "Game board",
    playHint: "Move the basket with the mouse, touch, or the arrow keys.",
    back: "Back",
    restore: "Restore purchase",
    discountLabel: "Discount code (optional, Zarinpal and Zibal only)",
    giftPrompt: "Enter your gift code",
    giftSubmit: "Redeem",
    more: "Load more",
    olderNews: "Older news",
    allNews: "All news",
    comments: "Comments",
    commentLabel: "Write a comment",
    commentPlaceholder: "Write a comment…",
    sendComment: "Post comment",
    moreComments: "More comments",
    faqSearch: "Search questions",
    faqSearchPlaceholder: "Search questions…",
    seenAll: "Mark all read",
    olderMessages: "Older messages",
    player: "Player",
    skinSelected: (name) => `${name} selected`,
    selected: "Selected",
    moreSkins: (n) => `${num(n)} more ${n === 1 ? "basket" : "baskets"}`,
    inShop: "In the shop",
    preparingRound: "Getting the round ready…",
    startFailed: "Could not start",
    rewards: "Rewards",
    rewardsHint: "Daily goals and records. Each one has a prize.",
    rewardsEmpty: "No rewards right now.",
    levelTitle: (n, title) => `Level ${num(n)}` + (title ? ` · ${title}` : ""),
    levelXp: (v, next, name) => `${num(v)} of ${num(next)} ${name}`,
    levelMax: (v, name) => `${num(v)} ${name} · top level`,
    claimLevel: (n, prize) => `Collect the level ${num(n)} prize: ${prize}`,
    streakDay: (n) => `Day ${num(n)}`,
    streakInfo: (count, best) => `${num(count)} days in a row · best: ${num(best)}`,
    streakToday: "Counted today. Come back tomorrow for the next day.",
    streakStart: "You signed in today; today's prize is ready.",
    streakBroken: "You missed some days; start again from day one.",
    streakGrace: (n) => `Missing ${num(n)} ${n === 1 ? "day" : "days"} will not break it.`,
    streakSafe: (n) => `Safe up to day ${num(n)}`,
    streakNextSafe: (n, left) => `Next safe day: day ${num(n)} (${num(left)} more ${left === 1 ? "day" : "days"})`,
    streakFloored: (n) => `You missed some days, but you went back to day ${num(n)}, not day one.`,
    streakSafePrize: (n, prize) => `Day ${num(n)}: ${prize}`,
    claim: "Collect",
    claimed: "Collected",
    locked: "Later",
    todayLabel: "Today",
    pendingLabel: "On the way",
    rewardClaimed: (text) => `You got ${text}`,
    rewardDone: "Done",
    rewardAuto: "Comes automatically",
    periodDay: "Today",
    periodWeek: "This week",
    progressOf: (c, target) => `${num(c)} of ${num(target)}`,
    levelUp: (n) => `Level up: level ${num(n)}!`,
    newRewards: (n) => `${num(n)} new ${n === 1 ? "reward is" : "rewards are"} waiting.`,
    autoRewards: (n) => `${num(n)} ${n === 1 ? "reward was" : "rewards were"} added for you.`,
    playAgain: "Play again",
    savingScore: "Saving your score…",
    roundCoins: (c, b) => `You got ${num(c)} ${c === 1 ? "coin" : "coins"}. Balance: ${num(b)} ${b === 1 ? "coin" : "coins"}`,
    roundNoCoins: (b) => `No coins this round. Balance: ${num(b)} ${b === 1 ? "coin" : "coins"}`,
    roundOver: "Round over!",
    scoreNotSaved: "Score not saved",
    coins: (n) => `${num(n)} ${n === 1 ? "coin" : "coins"}`,
    toman: (n) => `${num(n)} toman`,
    free: "Free",
    listSep: ", ",
    emptyShop: "Nothing for sale right now.",
    owned: "Yours",
    useInGame: "Use in game",
    buyWithCoins: "Buy with coins",
    confirmBuy: (name, n) => `Buy "${name}" for ${num(n)} ${n === 1 ? "coin" : "coins"}?`,
    buy: "Buy",
    goingTo: (m) => `Going to ${m}…`,
    buyOk: (r) => `Purchase complete! ${r}`,
    cancel: "Cancel",
    simPay: "Pay",
    confirmingPurchase: "Confirming the purchase…",
    simBadge: "Store simulator (test mode)",
    simNote: "There is no store app on the web, so this window stands in for it.",
    product: "Product",
    amount: "Amount",
    payment: "Payment",
    simTokenLabel: "Test purchase token (to test, you can paste an old token here)",
    checkRestore: "Check and restore",
    checking: "Checking…",
    restoreOk: (r) => `Purchase confirmed${r ? ": " + r : ""}. If you already got the reward, you will not get it again.`,
    statusIs: (s) => `Status: ${s}`,
    restoreTitle: "Restore a store purchase",
    restoreHelp: "If Bazaar or Myket charged you but the coins did not arrive, send the purchase token here.",
    store: "Store",
    purchaseToken: "Purchase token",
    close: "Close",
    giftGot: (r) => `Gift received: ${r}`,
    noPayments: "No payments yet.",
    refId: "Ref",
    checkingPayment: "Checking the payment…",
    payOk: "Payment successful",
    payPending: "The payment is still pending",
    payFailed: "The payment did not go through",
    payOkText: (b) => `The coins were added to your account. Balance: ${num(b)} ${b === 1 ? "coin" : "coins"}`,
    payPendingText: "Check \"My payments\" again in a few minutes.",
    payFailedText: "If you were charged, the money will be returned.",
    ok: "OK",
    sectionNotFound: "This section was not found in the game.",
    newsAria: (n) => (n ? `News, ${num(n)} new` : "News"),
    noNews: "No news yet.",
    pinned: "Pinned",
    fresh: "New",
    commentCount: (n) => `${num(n)} ${n === 1 ? "comment" : "comments"}`,
    commentsClosed: "Comments are closed for this news item",
    you: "You",
    awaitingApproval: "Awaiting approval",
    reply: "Reply",
    replyingTo: (name) => `Replying to ${name} `,
    noCommentsYet: "No comments yet. Be the first!",
    noComments: "No comments.",
    commentLive: "Your comment is live.",
    commentQueued: "Your comment was saved. Others will see it after approval.",
    all: "All",
    noFaqMatch: "No questions match this search.",
    noFaq: "No questions yet.",
    inboxAria: (n) => (n ? `Inbox, ${num(n)} unread` : "Inbox"),
    unread: "Unread",
    message: "Message",
    openLink: "Open link ↗",
    goTo: (s) => `Go to ${s}`,
    inboxEmpty: "Your inbox is empty.",
    allRead: "All messages marked as read.",
    newMessage: "New message",
    account: "My account",
    guest: "Guest",
    guestHint: "Your progress lives on this device only. Sign in with your phone to keep it on any device.",
    signedInAs: (p) => `Signed in with ${p}`,
    signedInHint: "Use this number to sign in on any device.",
    myCountry: "My country",
    myCountryHint: "Used for your flag and dialling code",
    chooseCountry: "Choose a country",
    searchCountry: "Search countries…",
    noCountry: "No country found.",
    countrySaved: (n) => `Your country is now ${n}`,
    otpTitle: "Sign in with your phone",
    otpHint: "We text you a one-time code.",
    phone: "Mobile number",
    otpKeep: "Keep the progress on this device",
    sendCode: "Send code",
    sending: "Sending…",
    codeSent: (p) => `We sent a code to ${p}.`,
    codeSandbox: "Sandbox: no SMS arrives; the code is in the server log.",
    code: "Code",
    codeDigits: (n) => `${num(n)}-digit code`,
    verify: "Sign in",
    verifying: "Checking…",
    resendIn: (n) => `Resend in ${num(n)} s`,
    resend: "Resend",
    changeNumber: "Use another number",
    welcomeBack: (p) => `Welcome back! Signed in with ${p}.`,
    newAccount: (p) => `New account created with ${p}.`,
    linked: (p) => `${p} is now linked to this account.`,
    signOut: "Sign out of the phone account",
    signedOut: "Signed out. You are a guest now.",
    username: "Username",
    saveName: "Save",
    nameSaved: "Username saved.",
    nameRule: (min, max, latin) => `${num(min)} to ${num(max)} characters${latin ? "; Latin letters, digits and _ . - only" : ""}.`,
    updateAvailable: "A new version is out",
    updateRequired: "Please update the game",
    updateNow: "Update",
    later: "Later",
    updateDefault: "Update the game to keep playing.",
    versions: (cur, min) => `Your version: ${cur} — lowest allowed: ${min}`,
    versionLabel: (v) => `Version ${v}`,
    errorCode: (code) => `Error: ${code}`,
    methods: { zarinpal: "Zarinpal", zibal: "Zibal", bazaar: "Cafe Bazaar", myket: "Myket", in_app: "Paid with coins" },
    statuses: { success: "Successful", pending: "Pending", failed: "Failed", expired: "Expired" },
    skins: { default: "Plain basket", skin_gold: "Golden basket", skin_neon: "Neon basket", skin_galaxy: "Galaxy basket" },
    chat: "Chat",
    chatTitle: "Player chat",
    chatPlaceholder: "Write a message",
    send: "Send",
    chatRules: "Bad words are hidden with stars. If a message hurts you, report it.",
    chatEmpty: "No messages yet. Say hi first!",
    chatJoining: "Joining the chat room…",
    chatOffline: "Chat is not connected. Try again in a moment.",
    chatYou: "You",
    report: "Report",
    reportMessage: (name) => `Report ${name}'s message`,
    reportTitle: "Why are you reporting this message?",
    reportNote: "Comment (optional)",
    reportSend: "Send report",
    reportDone: "Thanks, your report was sent!",
    reportDuplicate: "You already reported this message.",
    reportPick: "Pick a reason.",
    mutedUntil: (d) => `You cannot send chat messages until ${d}.`,
    mutedTitle: "You are muted for now",
    warnedTitle: "Warning",
    liftedMute: "You can send messages again.",
    suspendedTitle: "Your account is closed for now",
    suspendedUntil: (d) => `You cannot play until ${d}.`,
    myTags: "My tags",
    myTagsHint: "Tags the game gave you",
    exclusive: "Just for you",
    messageRemoved: "This message was removed.",
    errors: {
      CLAIM_NOT_FOUND: "This reward was not found.",
      ALREADY_CLAIMED: "You already collected this reward.",
      TOO_MANY_EVENTS: "Slow down a little and try again in a minute.",
      CHAT_BLOCKED: "This message had a banned word and was not sent.",
      CHAT_MUTED: "You cannot send messages right now.",
      ACCOUNT_SUSPENDED: "Your account is closed for now.",
      REPORT_TITLE_NOT_FOUND: "Pick a reason.",
      CANNOT_REPORT_OWN: "You cannot report your own message.",
      MESSAGE_NOT_FOUND: "This message no longer exists.",
      TOO_MANY_REPORTS: "Too many reports; try again later.",
      TEXT_TOO_LONG: "The text is too long.",
      OTP_DISABLED: "Phone sign-in is turned off for now.",
      OTP_UNAVAILABLE: "The SMS service is not ready. Try again later.",
      OTP_SEND_FAILED: "The SMS could not be sent. Try again.",
      OTP_TOO_SOON: "Wait a little before asking for a new code.",
      OTP_RATE_LIMITED: "Too many codes for this number. Try again later.",
      OTP_NOT_FOUND: "Ask for a code first.",
      OTP_EXPIRED: "The code has expired. Ask for a new one.",
      OTP_INVALID_CODE: "That code is not right.",
      OTP_TOO_MANY_ATTEMPTS: "Too many wrong codes. Ask for a new one.",
      INVALID_PHONE: "That mobile number is not valid.",
      PHONE_COUNTRY_NOT_ALLOWED: "Numbers from this country are not accepted.",
      ACCOUNT_DISABLED: "This account is closed.",
      ACCOUNT_HAS_PHONE: "This account already has another number.",
      COUNTRY_NOT_AVAILABLE: "This country is not available.",
      USERNAME_TOO_SHORT: "The username is too short.",
      USERNAME_TOO_LONG: "The username is too long.",
      USERNAME_INVALID_CHARACTERS: "The username can only use Latin letters, digits and _ . -",
      USERNAME_TAKEN: "Someone else has this username.",
      PASSWORD_TOO_SHORT: "The password is too short.",
      PASSWORD_NEEDS_LETTER: "The password needs a letter.",
      PASSWORD_NEEDS_DIGIT: "The password needs a digit.",
      DISPLAY_NAME_TOO_LONG: "The display name is too long.",
      UNAUTHENTICATED: "Your session ended. Sign in again.",
      REQUEST_TOO_LARGE: "The request is too large.",
      INVALID_JSON: "The request is not valid.",
      INVALID_LANGUAGE: "The language is not valid.",
      LANGUAGE_NOT_AVAILABLE: "This language is not available.",
      INVALID_CURSOR: "The page is not valid. Open it again.",
      INVALID_METHOD: "The payment method is not valid.",
      INVALID_TOKEN: "The purchase token is empty or not valid.",
      INVALID_ID: "The id is not valid.",
      INTERNAL_ERROR: "Server error. Try again in a moment.",
      TIMEOUT: "The server took too long. Try again.",
      CANCELED: "The request was canceled.",
      PRODUCT_NOT_FOUND: "This product was not found or is no longer for sale.",
      PRIZE_NOT_FOUND: "The prize for this item was not found.",
      PAYMENT_NOT_FOUND: "This payment was not found.",
      METHOD_NOT_ALLOWED: "This product can't be bought with this method.",
      METHOD_UNAVAILABLE: "This payment method is not available right now.",
      LEVEL_TOO_LOW: "Your level is too low for this product.",
      PRIZE_ALREADY_CLAIMED: "You already got this.",
      PRIZE_OWNER_MISMATCH: "This prize belongs to another account.",
      GIFT_NOT_APPLICABLE: "This discount code can't be used here.",
      GIFT_NOT_FOUND: "This code does not exist.",
      GIFT_INACTIVE: "This code is not active.",
      GIFT_NOT_STARTED: "This code can't be used yet.",
      GIFT_EXPIRED: "This code has expired.",
      GIFT_ALREADY_USED: "You already used this code.",
      GIFT_EXHAUSTED: "This code has no uses left.",
      GIFT_NOT_REDEEMABLE: "This is a discount code. Enter it when you pay through a gateway.",
      PRICE_TOO_LOW: "With this discount the final price is below the minimum.",
      INSUFFICIENT_BALANCE: "Not enough coins. Play a few rounds or buy a coin pack.",
      SKU_REQUIRED: "This product has no code in this store.",
      GATEWAY_REQUEST_FAILED: "The payment gateway did not answer. Try again.",
      GATEWAY_UNAVAILABLE: "The payment gateway is not available.",
      GATEWAY_REJECTED: "The payment gateway declined the payment.",
      REFERENCE_NOT_SAVED: "The payment could not be recorded. Try again.",
      PAYMENT_ALREADY_SETTLED: "This payment is already complete.",
      PAYMENT_CLOSED: "This payment is closed. Start a new purchase.",
      PAYMENT_NOT_SUCCESSFUL: "The payment was not successful.",
      STORE_CONFIGURATION_REJECTED: "The server rejected the store settings. The payment is still pending.",
      STORE_UNAVAILABLE: "The store is not available. Try again in a moment.",
      PURCHASE_INVALID: "The store did not confirm this purchase.",
      SANDBOX_TOKEN_INVALID: "The test token is not valid.",
      DEVELOPER_PAYLOAD_MISMATCH: "This token is not for this payment or this account.",
      TOKEN_ALREADY_USED: "This purchase token was already used.",
      TOKEN_NOT_VERIFIED: "The store did not return the subscription period.",
      SUBSCRIPTION_OWNED_BY_ANOTHER_USER: "This subscription belongs to another account.",
      SUBSCRIPTION_NOT_SUPPORTED: "This store has no subscriptions.",
      REWARD_PENDING: "The payment went through but the reward is not recorded yet. Support will check it.",
      AMOUNT_MISMATCH: "The paid amount did not match the order.",
      AMOUNT_UNKNOWN: "The paid amount is unknown.",
      NOT_VERIFIED: "The payment was not verified.",
      CANCELED_BY_USER: "The payment was canceled.",
      EXPIRED: "The payment time ran out.",
      TIMED_OUT: "The payment time ran out.",
      REFUNDED: "This payment was refunded.",
      ROUND_NOT_FOUND: "This round was not found.",
      ROUND_ALREADY_FINISHED: "This round was already recorded.",
      ROUND_TOO_SHORT: "The round ended too early and was not recorded.",
      ROUND_EXPIRED: "This round is too old and was not recorded.",
      SCORE_TOO_HIGH: "The score is not believable and was not recorded.",
      NETWORK: "Can't reach the server. Check your internet or the address.",
      NEWS_NOT_FOUND: "This news item was not found or is no longer published.",
      COMMENT_NOT_ALLOWED: "Comments are closed for this news item.",
      COMMENT_RATE_LIMITED: () => `Wait a moment. You can post only ${num(5)} comments a minute.`,
      COMMENT_MESSAGE_REQUIRED: "Write your comment first.",
      COMMENT_TOO_LONG: () => `A comment can be at most ${num(1000)} characters.`,
      PARENT_COMMENT_NOT_FOUND: "The comment you are replying to is gone.",
      QUERY_TOO_LONG: "The search text is too long.",
    },
  },
};

const SCREENS = ["home", "play", "shop", "gift", "payments", "news", "faq", "inbox", "account", "update", "chat", "rewards", "support", "ticket", "livechat", "referral", "activity", "notify"];

const SKINS = {
  default: { body: ["#c98a52", "#8a5a30"], rim: "#e4b383", glow: null },
  skin_gold: { body: ["#ffe27a", "#c98a00"], rim: "#fff4c2", glow: "rgba(255,201,60,0.55)" },
  skin_neon: { body: ["#0b1f2a", "#0b1f2a"], rim: "#2ef2ff", glow: "rgba(46,242,255,0.8)", neon: true },
  skin_galaxy: { body: ["#7b3fe4", "#1c1a6b"], rim: "#c6a8ff", glow: "rgba(150,110,255,0.6)", galaxy: true },
  skin_ruby: { body: ["#ff5a6e", "#9b0f2a"], rim: "#ffc2cb", glow: "rgba(255,70,100,0.55)" },
  skin_ice: { body: ["#d8f6ff", "#5fb8e8"], rim: "#ffffff", glow: "rgba(150,220,255,0.7)" },
};

const LOCALES = { fa: "fa-IR", en: "en-US" };
const i18n = { lang: "fa", dir: "rtl", ui: STRINGS.fa, nf: new Intl.NumberFormat("fa-IR"), locale: "fa-IR" };

function setLocale(code, direction) {
  i18n.lang = code;
  i18n.dir = direction === "ltr" ? "ltr" : "rtl";
  i18n.ui = STRINGS[code] || STRINGS.en;
  i18n.locale = LOCALES[code] || code;
  try {
    i18n.nf = new Intl.NumberFormat(i18n.locale);
  } catch (e) {
    i18n.locale = "en-US";
    i18n.nf = new Intl.NumberFormat("en-US");
  }
}

function num(n) {
  return i18n.nf.format(Number(n) || 0);
}

function t(key, ...args) {
  const v = i18n.ui[key] !== undefined ? i18n.ui[key] : STRINGS.en[key];
  return typeof v === "function" ? v(...args) : v;
}

function skinName(key) {
  return i18n.ui.skins[key] || STRINGS.en.skins[key] || key;
}

function methodLabel(m) {
  return i18n.ui.methods[m] || m;
}

function statusLabel(s) {
  return i18n.ui.statuses[s] || s;
}

const $ = (sel) => document.querySelector(sel);

const state = {
  token: null,
  username: "",
  wallet: {},
  items: [],
  catalog: null,
  shopTab: null,
  paymentsCursor: "",
  newsCursor: "",
  newsLoading: false,
  article: null,
  commentsCursor: "",
  replyTo: null,
  faqCategory: "",
  inboxCursor: "",
  unread: 0,
  newsNew: 0,
  languages: [],
  socket: null,
  socketRetry: 0,
  phone: "",
  countries: [],
  country: "",
  version: null,
  rules: null,
  suspended: null,
  myTags: [],
  rt: { cid: 0, pending: new Map() },
  chat: { channelId: "", joined: false, ids: new Set(), mute: null, titles: null },
  otp: { step: "phone", phone: "", country: "", until: 0, timer: 0, codeLength: 5, sandbox: false },
};

class ApiError extends Error {
  constructor(code, status, detail) {
    super(code);
    this.code = code;
    this.status = status;
    this.detail = detail || null;
  }
}

function parseNotice(message) {
  if (typeof message !== "string" || message[0] !== "{") return null;
  try {
    const v = JSON.parse(message);
    return v && typeof v.code === "string" ? v : null;
  } catch (e) {
    return null;
  }
}

function errorText(err) {
  const code = err && err.code ? err.code : "INTERNAL_ERROR";
  const v = i18n.ui.errors[code] !== undefined ? i18n.ui.errors[code] : STRINGS.en.errors[code];
  if (v === undefined) return t("errorCode", code);
  return typeof v === "function" ? v() : v;
}

function deviceId() {
  let id = localStorage.getItem(LS.device);
  if (!id) {
    const bytes = new Uint8Array(16);
    crypto.getRandomValues(bytes);
    id = "star-" + Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
    localStorage.setItem(LS.device, id);
  }
  return id;
}

async function request(path, options) {
  let res;
  try {
    res = await fetch(API + path, options);
  } catch (e) {
    throw new ApiError("NETWORK", 0);
  }
  const text = await res.text();
  let body = null;
  try {
    body = text ? JSON.parse(text) : null;
  } catch (e) {
    body = null;
  }
  if (!res.ok) {
    const notice = body ? parseNotice(body.message) : null;
    if (notice) throw new ApiError(notice.code, res.status, notice);
    const code = body && typeof body.message === "string" && /^[A-Z_]+$/.test(body.message) ? body.message : res.status === 401 ? "UNAUTHENTICATED" : "INTERNAL_ERROR";
    throw new ApiError(code, res.status);
  }
  return body;
}

function decodeJwt(token) {
  try {
    const part = token.split(".")[1].replace(/-/g, "+").replace(/_/g, "/");
    const json = decodeURIComponent(Array.from(atob(part), (c) => "%" + c.charCodeAt(0).toString(16).padStart(2, "0")).join(""));
    return JSON.parse(json);
  } catch (e) {
    return {};
  }
}

async function authenticate(path, id) {
  return request(path, {
    method: "POST",
    headers: { Authorization: "Basic " + btoa(SERVER_KEY + ":"), "Content-Type": "application/json" },
    body: JSON.stringify({ id }),
  });
}

async function login() {
  let body = null;
  const key = localStorage.getItem(LS.loginKey);
  if (key) {
    try {
      body = await authenticate("/v2/account/authenticate/custom?create=false", key);
    } catch (err) {
      if (err.code === "NETWORK") throw err;
      localStorage.removeItem(LS.loginKey);
      state.phone = "";
    }
  }
  if (!body) body = await authenticate("/v2/account/authenticate/device?create=true", deviceId());
  useSession(body.token);
}

function useSession(token) {
  state.token = token;
  state.username = decodeJwt(token).usn || "";
  connectSocket();
}

async function rpc(id, payload, retry = true) {
  try {
    return await request(`/v2/rpc/${id}?unwrap`, {
      method: "POST",
      headers: { Authorization: "Bearer " + state.token, "Content-Type": "application/json" },
      body: JSON.stringify(payload || {}),
    });
  } catch (err) {
    if (retry && err.status === 401) {
      try {
        await login();
      } catch (e) {
        if (e.code === "ACCOUNT_SUSPENDED") showSuspended(e.detail);
        throw e;
      }
      return rpc(id, payload, false);
    }
    throw err;
  }
}

let toastTimer = null;
function toast(text, kind = "") {
  const el = $("#toast");
  el.textContent = text;
  el.className = "toast" + (kind ? " " + kind : "");
  el.hidden = false;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => (el.hidden = true), 4200);
}

function el(tag, attrs = {}, children = []) {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (v === null || v === undefined || v === false) continue;
    if (k === "class") node.className = v;
    else if (k === "text") node.textContent = v;
    else if (k.startsWith("on")) node.addEventListener(k.slice(2), v);
    else node.setAttribute(k, v === true ? "" : v);
  }
  for (const c of [].concat(children)) {
    if (c === null || c === undefined || c === false) continue;
    node.append(c instanceof Node ? c : document.createTextNode(String(c)));
  }
  return node;
}

function show(name) {
  if (state.version && state.version.status === "update_required" && name !== "update") name = "update";
  if (state.suspended && name !== "suspended") name = "suspended";
  if (name !== "play") stopGame();
  for (const s of document.querySelectorAll(".screen")) s.hidden = s.id !== "screen-" + name;
  window.scrollTo(0, 0);
  if (name === "home") {
    renderHome();
    if (state.token) refreshNewsDot();
  }
  if (name === "shop") loadShop();
  if (name === "rewards") loadRewards();
  if (name === "payments") loadPayments(true);
  if (name === "news") loadNews(true);
  if (name === "faq") loadFaq();
  if (name === "inbox") {
    $("#note-toast").hidden = true;
    loadInbox(true);
  }
  if (name === "account") renderAccount();
  if (name === "chat") openChat();
  if (name === "suspended") renderSuspended();
  if (name === "update") renderUpdate();
  if (name === "gift") {
    $("#gift-msg").textContent = "";
    $("#gift-code").focus();
  }
  if (typeof growthShow === "function") growthShow(name);
  if (typeof missionsShow === "function") missionsShow(name);
  if (typeof inventoryShow === "function") inventoryShow(name);
  if (typeof achievementsShow === "function") achievementsShow(name);
  if (typeof chanceShow === "function") chanceShow(name);
  if (typeof lobbyShow === "function") lobbyShow(name);
  if (typeof bannersShow === "function") bannersShow(name);
  if (typeof notifyShow === "function") notifyShow(name);
  if (typeof sponsorsShow === "function") sponsorsShow(name);
}

async function refreshBalances() {
  const b = await rpc("economy_balances", {});
  state.wallet = b.wallet || {};
  state.items = b.items || [];
  $("#balance").textContent = num(state.wallet.coin || 0);
  const owned = selectedSkin();
  if (!ownsSkin(owned)) localStorage.setItem(LS.skin, "default");
  return b;
}

function ownsSkin(key) {
  return key === "default" || state.items.some((i) => i.key === key) || (state.invArts || []).includes(key);
}

function selectedSkin() {
  const key = localStorage.getItem(LS.skin) || "default";
  return SKINS[key] ? key : "default";
}

function drawBasket(ctx, x, y, w, h, skinKey, t = 0) {
  const s = SKINS[skinKey] || SKINS.default;
  ctx.save();
  if (s.glow) {
    ctx.shadowColor = s.glow;
    ctx.shadowBlur = s.neon ? 18 + Math.sin(t * 6) * 6 : 14;
  }
  const top = y - h / 2;
  const bottom = y + h / 2;
  const inset = w * 0.12;
  ctx.beginPath();
  ctx.moveTo(x - w / 2, top);
  ctx.lineTo(x + w / 2, top);
  ctx.lineTo(x + w / 2 - inset, bottom);
  ctx.lineTo(x - w / 2 + inset, bottom);
  ctx.closePath();
  const g = ctx.createLinearGradient(0, top, 0, bottom);
  g.addColorStop(0, s.body[0]);
  g.addColorStop(1, s.body[1]);
  ctx.fillStyle = g;
  ctx.fill();
  ctx.shadowBlur = 0;
  ctx.lineWidth = s.neon ? 3 : 2;
  ctx.strokeStyle = s.rim;
  ctx.stroke();
  ctx.beginPath();
  ctx.clip();
  ctx.lineWidth = 1.5;
  ctx.strokeStyle = s.neon ? "rgba(46,242,255,0.45)" : "rgba(0,0,0,0.22)";
  for (let i = 1; i < 6; i++) {
    const lx = x - w / 2 + (w / 6) * i;
    ctx.beginPath();
    ctx.moveTo(lx, top);
    ctx.lineTo(lx + (lx < x ? inset * 0.6 : -inset * 0.6), bottom);
    ctx.stroke();
  }
  ctx.beginPath();
  ctx.moveTo(x - w / 2, y);
  ctx.lineTo(x + w / 2, y);
  ctx.stroke();
  if (s.galaxy) {
    ctx.fillStyle = "rgba(255,255,255,0.9)";
    for (let i = 0; i < 9; i++) {
      const px = x - w / 2 + ((i * 37) % 100) / 100 * w;
      const py = top + ((i * 53) % 100) / 100 * h;
      ctx.globalAlpha = 0.5 + 0.5 * Math.sin(t * 3 + i);
      ctx.fillRect(px, py, 2, 2);
    }
    ctx.globalAlpha = 1;
  }
  ctx.restore();
  ctx.save();
  ctx.fillStyle = s.rim;
  ctx.fillRect(x - w / 2 - 2, top - 3, w + 4, 4);
  ctx.restore();
}

function skinPreview(key, w = 96, h = 48) {
  const c = document.createElement("canvas");
  const dpr = window.devicePixelRatio || 1;
  c.width = w * dpr;
  c.height = h * dpr;
  const ctx = c.getContext("2d");
  ctx.scale(dpr, dpr);
  drawBasket(ctx, w / 2, h / 2 + 4, w * 0.8, h * 0.5, key, 0.3);
  return c;
}

function renderHome() {
  $("#username").textContent = state.username || t("player");
  refreshProgress();
  paintTags($("#my-tags"));
  const box = $("#skins");
  box.textContent = "";
  const current = selectedSkin();
  for (const key of Object.keys(SKINS)) {
    if (!ownsSkin(key)) continue;
    box.append(
      el("button", {
        class: "skin",
        type: "button",
        "aria-pressed": key === current ? "true" : "false",
        "data-skin": key,
        onclick: () => {
          localStorage.setItem(LS.skin, key);
          renderHome();
          toast(t("skinSelected", skinName(key)), "ok");
        },
      }, [skinPreview(key), skinName(key), el("span", { class: "tag", text: key === current ? t("selected") : "" })])
    );
  }
  const missing = Object.keys(SKINS).filter((k) => !ownsSkin(k)).length;
  if (missing) {
    box.append(
      el("button", { class: "skin", type: "button", onclick: () => { state.shopTab = "star_skins"; show("shop"); } }, [
        el("span", { class: "muted", text: "+" }),
        t("moreSkins", missing),
        el("span", { class: "tag", text: t("inShop") }),
      ])
    );
  }
}

function roundNotes(ev) {
  if (!ev) return null;
  const lines = [];
  for (const l of ev.levels || []) if (l.to > Math.max(l.from, 1)) lines.push(t("levelUp", l.to));
  const claims = ev.claims || [];
  const waiting = claims.filter((c) => c.grantMode === "claim" && c.status !== "granted").length;
  const auto = claims.filter((c) => c.grantMode === "auto" && c.status === "granted").length;
  if (auto) lines.push(t("autoRewards", auto));
  if (waiting) lines.push(t("newRewards", waiting));
  const missions = typeof missionNotes === "function" ? missionNotes(ev) : null;
  if (missions) lines.push(missions);
  const achs = typeof achievementNotes === "function" ? achievementNotes(ev) : null;
  if (achs) lines.push(achs);
  if (!lines.length) return null;
  return el("div", { class: "round-notes", "data-round-notes": "" }, lines.map((x) => el("p", { text: x })));
}

let goldenQueue = Promise.resolve();
function recordGolden() {
  if (!state.token) return;
  goldenQueue = goldenQueue.then(() => rpc("economy_event_record", { key: "golden_star", value: 1 }, false)).catch(() => {});
}

let progressSeq = 0;
async function refreshProgress() {
  if (!state.token) return;
  const my = ++progressSeq;
  try {
    const [lv, sk] = await Promise.all([rpc("economy_levels", { lang: i18n.lang }), rpc("economy_streaks", { lang: i18n.lang })]);
    if (my !== progressSeq) return;
    state.levels = lv.items || [];
    state.streaks = sk.items || [];
    renderLevelCard();
    renderStreakCard();
    refreshRewardsDot();
  } catch (e) {
    if (my === progressSeq) {
      $("#level-card").hidden = true;
      $("#streak-card").hidden = true;
    }
  }
}

async function refreshRewardsDot() {
  try {
    const res = await rpc("economy_rewards_mine", { status: "ready", limit: 50, lang: i18n.lang });
    const n = (res.items || []).filter((c) => c.grantMode === "claim").length;
    const dot = $("#rewards-dot");
    dot.hidden = !n;
    dot.textContent = n ? num(n) : "";
  } catch (e) {}
}

async function claimReward(id, btn) {
  if (btn) btn.disabled = true;
  try {
    const r = await rpc("economy_reward_claim", { claimId: id, lang: i18n.lang });
    toast(t("rewardClaimed", rewardText(r.rewards) || (r.prize && r.prize.name) || ""), "ok");
    await refreshBalances().catch(() => {});
    await refreshProgress();
    if (currentScreen() === "rewards") loadRewards();
  } catch (err) {
    toast(errorText(err), "bad");
    if (btn) btn.disabled = false;
  }
}

function renderLevelCard() {
  const box = $("#level-card");
  const lv = (state.levels || [])[0];
  box.textContent = "";
  if (!lv) { box.hidden = true; return; }
  box.hidden = false;
  box.dataset.level = String(lv.level);
  const next = lv.next;
  const from = Number(lv.levelThreshold) || 0;
  const pct = next ? Math.max(0, Math.min(100, ((lv.value - from) / Math.max(1, next.threshold - from)) * 100)) : 100;
  box.append(
    el("div", { class: "level-top" }, [
      el("span", { class: "level-badge", "aria-hidden": "true", text: num(lv.level) }),
      el("div", { class: "level-name" }, [
        el("b", { id: "level-title", text: t("levelTitle", lv.level, lv.title) }),
        el("span", { class: "muted small", id: "level-xp", text: next ? t("levelXp", lv.value, next.threshold, lv.metaName) : t("levelMax", lv.value, lv.metaName) }),
      ]),
    ]),
    el("div", { class: "bar", role: "progressbar", "aria-valuemin": "0", "aria-valuemax": "100", "aria-valuenow": String(Math.round(pct)), "aria-label": lv.name }, [el("i", { style: `width:${pct.toFixed(1)}%` })])
  );
  for (const st of lv.steps || []) {
    if (st.status !== "claimable" || !st.claimId) continue;
    const b = el("button", { class: "btn btn-primary level-claim", type: "button", "data-claim": st.claimId, "data-level": String(st.level), text: t("claimLevel", st.level, st.prize ? st.prize.name : "") });
    b.addEventListener("click", () => claimReward(st.claimId, b));
    box.append(b);
  }
}

function renderStreakCard() {
  const box = $("#streak-card");
  const sk = (state.streaks || [])[0];
  box.textContent = "";
  if (!sk) { box.hidden = true; return; }
  box.hidden = false;
  box.dataset.streak = sk.key;
  box.dataset.count = String(sk.count);
  box.append(el("div", { class: "panel-head" }, [
    el("h2", { id: "streak-title", text: sk.name }),
    el("span", { class: "muted small", id: "streak-info", text: t("streakInfo", sk.count, sk.best) }),
  ]));
  const days = el("div", { class: "streak-days" });
  for (const d of sk.days || []) {
    const label = { claimed: t("claimed"), claimable: t("claim"), done: t("claimed"), today: t("todayLabel"), pending: t("pendingLabel") }[d.status] || t("locked");
    const tile = el(d.status === "claimable" ? "button" : "div", { class: "streak-day " + d.status, "data-day": String(d.day), "data-status": d.status, type: d.status === "claimable" ? "button" : null }, [
      el("span", { class: "day-n", text: t("streakDay", d.day) }),
      el("span", { class: "day-prize", text: d.prize ? d.prize.name : "" }),
      el("span", { class: "day-state", text: label }),
    ]);
    if (d.status === "claimable" && d.claimId) tile.addEventListener("click", () => claimReward(d.claimId, tile));
    days.append(tile);
  }
  box.append(days);
  const safe = streakSafeRow(sk);
  if (safe) box.append(safe);
  const note = sk.broken && sk.floored ? t("streakFloored", sk.floor) : sk.broken ? t("streakBroken") : sk.countedThisPeriod ? (sk.claimable && sk.claimable.length ? t("streakStart") : t("streakToday")) : sk.gracePeriods ? t("streakGrace", sk.gracePeriods) : "";
  if (note) box.append(el("p", { class: "muted small streak-note", text: note }));
}

function streakSafeRow(sk) {
  const ms = sk.milestones || [];
  if (!ms.length) return null;
  const row = el("div", { class: "streak-safe", id: "streak-safe", "data-floor": String(sk.floor || 0) });
  if (sk.floor) row.append(el("span", { class: "safe-floor", id: "streak-floor", text: t("streakSafe", sk.floor) }));
  if (sk.nextMilestone) row.append(el("span", { class: "safe-next", id: "streak-next", text: t("streakNextSafe", sk.nextMilestone, Math.max(0, sk.nextMilestone - sk.count)) }));
  for (const m of ms) {
    if (!m.prize) continue;
    const label = { claimed: t("claimed"), claimable: t("claim"), pending: t("pendingLabel") }[m.status] || (m.reached ? t("claimed") : t("locked"));
    const chip = el(m.status === "claimable" ? "button" : "span", { class: "safe-prize " + m.status, "data-milestone": String(m.day), "data-status": m.status, type: m.status === "claimable" ? "button" : null }, [
      el("span", { text: t("streakSafePrize", m.day, m.prize.name) }),
      el("span", { class: "day-state", text: label }),
    ]);
    if (m.status === "claimable" && m.claimId) chip.addEventListener("click", () => claimReward(m.claimId, chip));
    row.append(chip);
  }
  return row;
}

async function loadRewards() {
  const box = $("#rewards-list");
  box.textContent = "";
  box.append(el("div", { class: "center-block" }, [el("div", { class: "spinner" })]));
  try {
    const res = await rpc("economy_events_progress", { lang: i18n.lang });
    box.textContent = "";
    const items = res.items || [];
    if (!items.length) { box.append(el("p", { class: "muted center", text: t("rewardsEmpty") })); return; }
    for (const it of items) {
      const ready = (it.claimable || []).length > 0;
      const target = Number(it.target) || Number(it.threshold) || 1;
      const current = Number(it.current) || 0;
      const pct = it.done ? 100 : Math.max(0, Math.min(100, (current / target) * 100));
      const period = it.period === "day" ? t("periodDay") : it.period === "week" ? t("periodWeek") : "";
      const foot = el("div", { class: "reward-foot" }, [el("span", { class: "reward-prize", text: it.prize ? it.prize.name : "" })]);
      if (ready) {
        for (const id of it.claimable) {
          const b = el("button", { class: "btn btn-primary", type: "button", "data-claim": id, text: t("claim") });
          b.addEventListener("click", () => claimReward(id, b));
          foot.append(b);
        }
      } else if (it.done) foot.append(el("span", { class: "chip", text: t("rewardDone") }));
      else if (it.grantMode === "auto") foot.append(el("span", { class: "muted small", text: t("rewardAuto") }));
      box.append(el("article", { class: "reward" + (ready ? " ready" : it.done ? " done" : ""), "data-rule": it.key }, [
        el("div", { class: "reward-head" }, [el("b", { text: it.name }), el("span", { class: "muted small", "data-progress": "", text: (period ? period + " · " : "") + t("progressOf", Math.min(current, target), target) })]),
        it.description ? el("p", { class: "muted small", text: it.description }) : null,
        el("div", { class: "bar", role: "progressbar", "aria-valuemin": "0", "aria-valuemax": "100", "aria-valuenow": String(Math.round(pct)), "aria-label": it.name }, [el("i", { style: `width:${pct.toFixed(1)}%` })]),
        foot,
      ]));
    }
  } catch (err) {
    box.textContent = "";
    box.append(el("p", { class: "muted center", text: errorText(err) }));
  }
}

const game = {
  running: false,
  raf: 0,
  canvas: null,
  ctx: null,
  w: 0,
  h: 0,
  basket: { x: 0, w: 90, h: 26, target: null },
  objects: [],
  floats: [],
  sky: [],
  keys: { left: false, right: false },
  score: 0,
  time: ROUND_SECONDS,
  spawnIn: 0,
  last: 0,
  roundId: null,
  phase: "idle",
  skin: "default",
  clock: 0,
};

window.__stargame = {
  get phase() { return game.phase; },
  get width() { return game.w; },
  get basket() { return { x: game.basket.x, w: game.basket.w, y: game.h - 34 }; },
  get objects() { return game.objects.map((o) => ({ x: o.x, y: o.y, kind: o.kind })); },
  get score() { return game.score; },
};

function resizeCanvas() {
  const c = game.canvas;
  const rect = c.parentElement.getBoundingClientRect();
  const w = Math.floor(rect.width);
  const avail = window.innerHeight - c.getBoundingClientRect().top - 60;
  const h = Math.max(340, Math.min(620, avail, Math.round(w * 1.25)));
  const dpr = window.devicePixelRatio || 1;
  c.style.height = h + "px";
  c.width = Math.round(w * dpr);
  c.height = Math.round(h * dpr);
  game.ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  const oldW = game.w || w;
  game.w = w;
  game.h = h;
  game.basket.w = Math.max(70, Math.min(130, w * 0.18));
  game.basket.x = game.basket.x ? (game.basket.x / oldW) * w : w / 2;
  game.sky = Array.from({ length: Math.round((w * h) / 4000) }, () => ({ x: Math.random() * w, y: Math.random() * h, r: Math.random() * 1.3 + 0.2, p: Math.random() * 6 }));
}

function setOverlay(content) {
  const o = $("#stage-overlay");
  o.textContent = "";
  if (!content) {
    o.hidden = true;
    return;
  }
  o.append(content);
  o.hidden = false;
}

function updateHud() {
  $("#hud-score").textContent = num(game.score);
  $("#hud-time").textContent = num(Math.ceil(Math.max(0, game.time)));
}

async function startGame() {
  show("play");
  game.canvas = $("#canvas");
  game.ctx = game.canvas.getContext("2d");
  resizeCanvas();
  game.objects = [];
  game.floats = [];
  game.score = 0;
  game.roundSeconds = ROUND_SECONDS;
  game.time = ROUND_SECONDS;
  game.spawnIn = 0.4;
  game.basket.x = game.w / 2;
  game.basket.target = null;
  game.skin = selectedSkin();
  game.phase = "starting";
  game.roundId = null;
  updateHud();
  setOverlay(el("div", { class: "result" }, [el("div", { class: "spinner" }), el("p", { text: t("preparingRound") })]));
  game.running = true;
  game.last = performance.now();
  cancelAnimationFrame(game.raf);
  game.raf = requestAnimationFrame(loop);
  try {
    const r = await rpc("stargame_start", {});
    if (!game.running) return;
    game.roundId = r.roundId;
    if (r.seconds > 0) {
      game.roundSeconds = r.seconds;
      game.time = r.seconds;
      updateHud();
    }
  } catch (err) {
    if (!game.running) return;
    game.phase = "error";
    setOverlay(resultView(t("startFailed"), errorText(err), null));
    return;
  }
  game.phase = "countdown";
  game.count = 3;
  setOverlay(el("div", { class: "countdown", text: num(3) }));
}

function stopGame() {
  game.running = false;
  game.phase = "idle";
  cancelAnimationFrame(game.raf);
}

function spawn() {
  const r = Math.random();
  const gold = typeof goldenRush === "function" && goldenRush() ? 0.2 : 0.1;
  const kind = r < gold ? "gold" : r < gold + 0.2 ? "bomb" : "star";
  const size = kind === "gold" ? 17 : kind === "bomb" ? 15 : 13;
  const elapsed = (game.roundSeconds || ROUND_SECONDS) - game.time;
  const speed = (game.h * (0.38 + Math.random() * 0.22)) * (1 + elapsed / 45);
  game.objects.push({ kind, x: size + Math.random() * (game.w - size * 2), y: -size, r: size, vy: speed, rot: Math.random() * Math.PI, spin: (Math.random() - 0.5) * 3 });
}

function loop(now) {
  if (!game.running) return;
  const dt = Math.min(0.1, (now - game.last) / 1000);
  game.last = now;
  game.clock += dt;
  update(dt);
  draw();
  game.raf = requestAnimationFrame(loop);
}

function update(dt) {
  const b = game.basket;
  const speed = game.w * 1.1;
  if (game.keys.left) b.x -= speed * dt;
  if (game.keys.right) b.x += speed * dt;
  if (b.target !== null && !game.keys.left && !game.keys.right) b.x += (b.target - b.x) * Math.min(1, dt * 18);
  b.x = Math.max(b.w / 2, Math.min(game.w - b.w / 2, b.x));

  if (game.phase === "countdown") {
    const prev = Math.ceil(game.count);
    game.count -= dt;
    const next = Math.ceil(game.count);
    if (game.count <= 0) {
      game.phase = "playing";
      setOverlay(null);
    } else if (next !== prev) {
      setOverlay(el("div", { class: "countdown", text: num(next) }));
    }
    return;
  }
  if (game.phase !== "playing") return;

  game.time -= dt;
  game.spawnIn -= dt;
  const elapsed = (game.roundSeconds || ROUND_SECONDS) - game.time;
  if (game.spawnIn <= 0) {
    spawn();
    game.spawnIn = Math.max(0.28, 0.55 - elapsed * 0.009) * (0.7 + Math.random() * 0.6);
  }
  const by = game.h - 34;
  for (const o of game.objects) {
    o.y += o.vy * dt;
    o.rot += o.spin * dt;
    if (!o.done && o.y + o.r >= by - b.h / 2 && o.y - o.r <= by + b.h / 2 && Math.abs(o.x - b.x) <= b.w / 2 + o.r * 0.4) {
      o.done = true;
      const delta = o.kind === "gold" ? 5 : o.kind === "bomb" ? -3 : 1;
      if (o.kind === "gold") recordGolden();
      game.score += delta;
      game.floats.push({ x: o.x, y: by - 30, text: (delta > 0 ? "+" : "−") + num(Math.abs(delta)), color: delta > 0 ? (delta > 1 ? "#ffc93c" : "#fff3b0") : "#ff6b6b", life: 0.9 });
    }
  }
  game.objects = game.objects.filter((o) => !o.done && o.y - o.r < game.h);
  for (const f of game.floats) {
    f.y -= 50 * dt;
    f.life -= dt;
  }
  game.floats = game.floats.filter((f) => f.life > 0);
  updateHud();
  if (game.time <= 0) finishGame();
}

function drawStar(ctx, x, y, r, rot, fill, glow) {
  ctx.save();
  ctx.translate(x, y);
  ctx.rotate(rot);
  ctx.beginPath();
  for (let i = 0; i < 10; i++) {
    const rr = i % 2 === 0 ? r : r * 0.45;
    const a = (Math.PI / 5) * i - Math.PI / 2;
    ctx.lineTo(Math.cos(a) * rr, Math.sin(a) * rr);
  }
  ctx.closePath();
  ctx.shadowColor = glow;
  ctx.shadowBlur = r;
  ctx.fillStyle = fill;
  ctx.fill();
  ctx.restore();
}

function drawBomb(ctx, o) {
  ctx.save();
  ctx.translate(o.x, o.y);
  ctx.beginPath();
  ctx.arc(0, 0, o.r, 0, Math.PI * 2);
  const g = ctx.createRadialGradient(-o.r * 0.35, -o.r * 0.35, 1, 0, 0, o.r);
  g.addColorStop(0, "#6b6f86");
  g.addColorStop(1, "#14162a");
  ctx.fillStyle = g;
  ctx.fill();
  ctx.strokeStyle = "#ff6b6b";
  ctx.lineWidth = 1.5;
  ctx.stroke();
  ctx.strokeStyle = "#c9b28a";
  ctx.lineWidth = 2;
  ctx.beginPath();
  ctx.moveTo(o.r * 0.4, -o.r * 0.8);
  ctx.quadraticCurveTo(o.r * 0.9, -o.r * 1.4, o.r * 1.1, -o.r * 1.2);
  ctx.stroke();
  ctx.fillStyle = Math.sin(game.clock * 30) > 0 ? "#ffcf4a" : "#ff6b3d";
  ctx.beginPath();
  ctx.arc(o.r * 1.1, -o.r * 1.2, 3, 0, Math.PI * 2);
  ctx.fill();
  ctx.restore();
}

function draw() {
  const ctx = game.ctx;
  const { w, h } = game;
  ctx.clearRect(0, 0, w, h);
  const bg = ctx.createLinearGradient(0, 0, 0, h);
  bg.addColorStop(0, "#070b22");
  bg.addColorStop(1, "#1a1550");
  ctx.fillStyle = bg;
  ctx.fillRect(0, 0, w, h);
  for (const s of game.sky) {
    ctx.globalAlpha = 0.35 + 0.35 * Math.sin(game.clock * 2 + s.p);
    ctx.fillStyle = "#fff";
    ctx.fillRect(s.x, s.y, s.r, s.r);
  }
  ctx.globalAlpha = 1;
  ctx.fillStyle = "rgba(124,140,255,0.12)";
  ctx.fillRect(0, h - 12, w, 12);
  for (const o of game.objects) {
    if (o.kind === "bomb") drawBomb(ctx, o);
    else if (o.kind === "gold") drawStar(ctx, o.x, o.y, o.r, o.rot, "#ffc93c", "rgba(255,201,60,0.9)");
    else drawStar(ctx, o.x, o.y, o.r, o.rot, "#fff3b0", "rgba(255,243,176,0.6)");
  }
  drawBasket(ctx, game.basket.x, h - 34, game.basket.w, game.basket.h, game.skin, game.clock);
  ctx.textAlign = "center";
  ctx.direction = "ltr";
  ctx.font = "700 20px " + getComputedStyle(document.body).fontFamily;
  for (const f of game.floats) {
    ctx.globalAlpha = Math.max(0, f.life / 0.9);
    ctx.fillStyle = f.color;
    ctx.fillText(f.text, f.x, f.y);
  }
  ctx.globalAlpha = 1;
}

function resultView(title, text, extra) {
  return el("div", { class: "result" }, [
    el("h2", { text: title }),
    extra,
    text ? el("p", { class: "muted", text }) : null,
    el("div", { class: "row" }, [
      el("button", { class: "btn btn-primary", type: "button", text: t("playAgain"), onclick: startGame }),
      el("button", { class: "btn", type: "button", text: t("home"), onclick: () => show("home") }),
    ]),
  ]);
}

async function finishGame() {
  game.phase = "finishing";
  game.time = 0;
  updateHud();
  const score = game.score;
  setOverlay(el("div", { class: "result" }, [el("div", { class: "spinner" }), el("p", { text: t("savingScore") })]));
  try {
    const r = await rpc("stargame_finish", { roundId: game.roundId, score });
    state.wallet.coin = r.balance;
    $("#balance").textContent = num(r.balance);
    game.phase = "done";
    const text = r.coins > 0 ? t("roundCoins", r.coins, r.balance) : t("roundNoCoins", r.balance);
    const notes = roundNotes(r.events);
    setOverlay(resultView(t("roundOver"), text, el("div", {}, [el("div", { class: "big", text: num(score) }), notes])));
    if (r.events) {
      refreshBalances().catch(() => {});
      refreshProgress();
    }
    $("#stage-overlay").dataset.coins = String(r.coins);
  } catch (err) {
    game.phase = "done";
    setOverlay(resultView(t("scoreNotSaved"), errorText(err), el("div", { class: "big", text: num(score) })));
  }
}

function pointerToBasket(e) {
  if (!game.canvas) return;
  const rect = game.canvas.getBoundingClientRect();
  game.basket.target = e.clientX - rect.left;
}

function setupInput() {
  const c = $("#canvas");
  c.addEventListener("pointermove", (e) => {
    if (e.pointerType === "mouse" || e.buttons || e.pressure > 0) pointerToBasket(e);
  });
  c.addEventListener("pointerdown", (e) => {
    c.setPointerCapture(e.pointerId);
    pointerToBasket(e);
  });
  window.addEventListener("keydown", (e) => {
    if ($("#screen-play").hidden) return;
    if (e.key === "ArrowLeft" || e.key === "a" || e.key === "A") { game.keys.left = true; game.basket.target = null; e.preventDefault(); }
    if (e.key === "ArrowRight" || e.key === "d" || e.key === "D") { game.keys.right = true; game.basket.target = null; e.preventDefault(); }
  });
  window.addEventListener("keyup", (e) => {
    if (e.key === "ArrowLeft" || e.key === "a" || e.key === "A") game.keys.left = false;
    if (e.key === "ArrowRight" || e.key === "d" || e.key === "D") game.keys.right = false;
  });
  window.addEventListener("blur", () => { game.keys.left = game.keys.right = false; });
  window.addEventListener("resize", () => { if (game.running) resizeCanvas(); });
}

function rewardText(rewards) {
  if (!rewards || !rewards.length) return "";
  return rewards.map((r) => {
    if (r.type === "currency") return r.key === "coin" ? t("coins", r.value) : `${num(r.value)} ${r.key}`;
    if (r.type === "storage") return SKINS[r.key] ? skinName(r.key) : (r.value && r.value.name) || r.key;
    if (r.type === "item") return typeof itemName === "function" ? itemName(r.key) : r.key;
    return `${r.key} ${num(r.value)}`;
  }).join(t("listSep"));
}

function priceView(p) {
  if (p.coinPrice) return el("span", { class: "price" }, t("coins", p.coinPrice.amount));
  if (p.priceToman !== null && p.priceToman !== undefined) {
    if (p.finalToman !== p.priceToman) return el("span", { class: "price" }, [el("s", { text: num(p.priceToman) }), t("toman", p.finalToman)]);
    return el("span", { class: "price" }, t("toman", p.priceToman));
  }
  return el("span", { class: "price", text: t("free") });
}

async function loadShop() {
  const box = $("#products");
  box.textContent = "";
  box.append(el("div", { class: "center-block" }, [el("div", { class: "spinner" })]));
  try {
    const [catalog] = await Promise.all([rpc("economy_catalog", { lang: i18n.lang }), refreshBalances(), typeof loadInvArts === "function" ? loadInvArts() : null]);
    state.catalog = catalog;
    renderShop();
    if (typeof bannerShopHook === "function") bannerShopHook();
  } catch (err) {
    box.textContent = "";
    box.append(el("div", { class: "empty", text: errorText(err) }));
  }
}

function renderShop() {
  const cat = state.catalog;
  const used = new Set(cat.products.map((p) => p.categoryKey));
  const cats = cat.categories.filter((c) => used.has(c.key));
  if (!cats.some((c) => c.key === state.shopTab)) state.shopTab = cats.length ? cats[0].key : null;
  const tabs = $("#shop-tabs");
  tabs.textContent = "";
  for (const c of cats) {
    tabs.append(el("button", { class: "tab", type: "button", role: "tab", "aria-selected": c.key === state.shopTab ? "true" : "false", "data-cat": c.key, onclick: () => { state.shopTab = c.key; renderShop(); } }, c.name));
  }
  const products = cat.products.filter((p) => p.categoryKey === state.shopTab);
  $("#discount-row").hidden = !products.some((p) => p.paymentMethods.some((m) => m === "zarinpal" || m === "zibal"));
  const box = $("#products");
  box.textContent = "";
  if (!products.length) {
    box.append(el("div", { class: "empty", text: t("emptyShop") }));
    return;
  }
  for (const p of products) box.append(productCard(p));
}

function skinKeyOf(p) {
  const m = /^star_(skin_[a-z]+)$/.exec(p.key);
  return m && SKINS[m[1]] ? m[1] : null;
}

function productCard(p) {
  const itemArt = typeof productItemArt === "function" ? productItemArt(p) : null;
  const skin = skinKeyOf(p) || itemArt;
  const owned = p.owned || (skinKeyOf(p) && ownsSkin(skin));
  const art = el("div", { class: "product-art" }, skin ? skinPreview(skin, 72, 36) : el("span", { class: "coin" }));
  const methods = el("div", { class: "methods" });
  if (owned && !skinKeyOf(p)) {
    methods.append(el("span", { class: "owned", "data-owned": "", text: t("owned") }));
    methods.append(el("button", { class: "btn", type: "button", text: t("inventory"), onclick: () => show("inventory") }));
  } else if (p.levelLocked) {
    methods.append(el("span", { class: "chip chip-wait", "data-level-locked": "", text: t("fromLevel", p.minLevel) }));
  } else if (owned) {
    const selected = selectedSkin() === skin;
    methods.append(el("span", { class: "owned", text: t("owned") }));
    methods.append(el("button", { class: "btn", type: "button", disabled: selected, text: selected ? t("selected") : t("useInGame"), onclick: () => { localStorage.setItem(LS.skin, skin); toast(t("skinSelected", skinName(skin)), "ok"); renderShop(); } }));
  } else {
    for (const m of p.paymentMethods) {
      methods.append(el("button", { class: m === "in_app" ? "btn btn-primary" : "btn", type: "button", "data-method": m, text: m === "in_app" ? t("buyWithCoins") : methodLabel(m), onclick: (e) => buy(p, m, e.currentTarget) }));
    }
  }
  return el("article", { class: "product" + (p.exclusive ? " exclusive" : ""), "data-key": p.key }, [
    el("div", { class: "product-top" }, [art, el("div", {}, [el("h3", {}, [p.name, p.exclusive ? el("span", { class: "badge-exclusive", text: t("exclusive") }) : null, p.isOneTime ? el("span", { class: "badge-exclusive", "data-one-time": "", text: t("oneTime") }) : null]), p.description ? el("p", { class: "muted small", text: p.description }) : null])]),
    priceView(p),
    methods,
  ]);
}

async function buy(p, method, button) {
  if (button) button.disabled = true;
  try {
    if (method === "in_app") {
      const ok = await confirmDialog(t("buyWithCoins"), t("confirmBuy", p.name, p.coinPrice.amount), t("buy"));
      if (!ok) return;
    }
    const payload = { productId: p.id, method, lang: i18n.lang };
    const code = $("#discount-code").value.trim().toUpperCase();
    if ((method === "zarinpal" || method === "zibal") && code) payload.giftCode = code;
    const r = await rpc("economy_purchase_start", payload);
    if (r.status === "redirect") {
      toast(t("goingTo", methodLabel(method)));
      location.href = r.redirectUrl;
      return;
    }
    if (r.status === "store") {
      await storeSimulator(p, method, r);
      return;
    }
    await refreshBalances();
    toast(r.purchaseMessage || t("buyOk", rewardText(r.rewards)), "ok");
    if (skinKeyOf(p)) localStorage.setItem(LS.skin, skinKeyOf(p));
    if (p.isOneTime || (p.items || []).length) {
      state.catalog = await rpc("economy_catalog", { lang: i18n.lang });
      if (typeof loadInvArts === "function") loadInvArts();
    }
    renderShop();
  } catch (err) {
    toast(errorText(err), "bad");
  } finally {
    if (button && button.isConnected) button.disabled = false;
  }
}

function openDialog(build) {
  const d = $("#dialog");
  const body = $("#dialog-body");
  body.textContent = "";
  return new Promise((resolve) => {
    const close = (v) => {
      d.close();
      resolve(v);
    };
    build(body, close);
    d.onclose = () => resolve(undefined);
    d.showModal();
  });
}

function confirmDialog(title, text, okText) {
  return openDialog((body, close) => {
    body.append(
      el("h2", { text: title }),
      el("p", { text }),
      el("div", { class: "dialog-actions" }, [
        el("button", { class: "btn btn-primary", type: "button", text: okText, onclick: () => close(true) }),
        el("button", { class: "btn", type: "button", text: t("cancel"), onclick: () => close(false) }),
      ])
    );
  });
}

function randomPart() {
  const bytes = new Uint8Array(6);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
}

function storeSimulator(p, method, start) {
  const token = `sandbox-${start.paymentId}-${randomPart()}`;
  return openDialog((body, close) => {
    const input = el("input", { type: "text", id: "sim-token", class: "mono", dir: "ltr", spellcheck: "false", value: token });
    const msg = el("p", { class: "form-msg", role: "status" });
    const pay = el("button", { class: "btn btn-ok", type: "button", id: "sim-pay", text: t("simPay") });
    pay.addEventListener("click", async () => {
      pay.disabled = true;
      msg.className = "form-msg";
      msg.textContent = t("confirmingPurchase");
      const used = input.value.trim();
      try {
        const r = await rpc("economy_purchase_submit", { paymentId: start.paymentId, token: used });
        localStorage.setItem(LS.storeToken, JSON.stringify({ method, sku: start.sku, token: used }));
        await refreshBalances();
        close(true);
        if (r.status === "success" && r.rewardGranted === false) toast(errorText({ code: r.error || "REWARD_PENDING" }), "bad");
        else toast(t("buyOk", rewardText(r.rewards)), "ok");
        if (!$("#screen-shop").hidden) renderShop();
      } catch (err) {
        msg.className = "form-msg bad";
        msg.textContent = errorText(err);
        pay.disabled = false;
      }
    });
    body.append(
      el("span", { class: "sim-badge", text: t("simBadge") }),
      el("div", { class: "sim-store" }, [el("div", { class: "sim-logo " + method, text: method === "bazaar" ? "ب" : "م" }), el("div", {}, [el("b", { text: methodLabel(method) }), el("div", { class: "muted small", text: t("simNote") })])]),
      el("dl", { class: "kv" }, [
        el("dt", { text: t("product") }), el("dd", { text: p.name }),
        el("dt", { text: t("amount") }), el("dd", { text: p.finalToman !== null && p.finalToman !== undefined ? t("toman", p.finalToman) : "-" }),
        el("dt", { text: "SKU" }), el("dd", { class: "mono", dir: "ltr", text: start.sku }),
        el("dt", { text: t("payment") }), el("dd", { class: "mono", dir: "ltr", text: start.paymentId }),
      ]),
      el("div", { class: "field" }, [el("label", { for: "sim-token", text: t("simTokenLabel") }), input]),
      msg,
      el("div", { class: "dialog-actions" }, [pay, el("button", { class: "btn", type: "button", text: t("cancel"), onclick: () => close(false) })])
    );
  });
}

function restoreDialog() {
  const last = JSON.parse(localStorage.getItem(LS.storeToken) || "null");
  const skus = state.catalog ? state.catalog.products.filter((p) => p.sku && (p.sku.bazaar || p.sku.myket)) : [];
  return openDialog((body, close) => {
    const method = el("select", { id: "restore-method" }, [el("option", { value: "bazaar", text: methodLabel("bazaar") }), el("option", { value: "myket", text: methodLabel("myket") })]);
    const sku = el("select", { id: "restore-sku", dir: "ltr" });
    const fillSkus = () => {
      sku.textContent = "";
      for (const p of skus) {
        const s = p.sku[method.value];
        if (s) sku.append(el("option", { value: s, text: `${s} · ${p.name}` }));
      }
      if (last && last.method === method.value) sku.value = last.sku;
    };
    method.addEventListener("change", fillSkus);
    if (last) method.value = last.method;
    fillSkus();
    const token = el("input", { type: "text", id: "restore-token", class: "mono", dir: "ltr", spellcheck: "false", value: last ? last.token : "" });
    const msg = el("p", { class: "form-msg", role: "status" });
    const go = el("button", { class: "btn btn-primary", type: "button", text: t("checkRestore") });
    go.addEventListener("click", async () => {
      go.disabled = true;
      msg.className = "form-msg";
      msg.textContent = t("checking");
      try {
        const r = await rpc("economy_purchase_review", { method: method.value, sku: sku.value, token: token.value.trim() });
        await refreshBalances();
        msg.className = "form-msg ok";
        msg.textContent = r.status === "success" ? t("restoreOk", r.rewards ? rewardText(r.rewards) : "") : t("statusIs", statusLabel(r.status));
      } catch (err) {
        msg.className = "form-msg bad";
        msg.textContent = errorText(err);
      } finally {
        go.disabled = false;
      }
    });
    body.append(
      el("h2", { text: t("restoreTitle") }),
      el("p", { class: "muted small", text: t("restoreHelp") }),
      el("div", { class: "field" }, [el("label", { for: "restore-method", text: t("store") }), method]),
      el("div", { class: "field" }, [el("label", { for: "restore-sku", text: t("product") }), sku]),
      el("div", { class: "field" }, [el("label", { for: "restore-token", text: t("purchaseToken") }), token]),
      msg,
      el("div", { class: "dialog-actions" }, [go, el("button", { class: "btn", type: "button", text: t("close"), onclick: () => close(false) })])
    );
  });
}

async function redeemGift(e) {
  e.preventDefault();
  const input = $("#gift-code");
  const msg = $("#gift-msg");
  const btn = e.submitter || $("#gift-form button");
  const code = input.value.trim().toUpperCase();
  if (!code) return;
  btn.disabled = true;
  msg.className = "form-msg";
  msg.textContent = t("checking");
  try {
    const r = await rpc("economy_gift_redeem", { code });
    await refreshBalances();
    msg.className = "form-msg ok";
    msg.textContent = t("giftGot", rewardText(r.rewards));
    input.value = "";
  } catch (err) {
    msg.className = "form-msg bad";
    msg.textContent = errorText(err);
  } finally {
    btn.disabled = false;
  }
}

function formatDate(iso) {
  try {
    return new Date(iso).toLocaleString(i18n.locale, { dateStyle: "medium", timeStyle: "short" });
  } catch (e) {
    return String(iso);
  }
}

async function loadPayments(reset) {
  const box = $("#payments");
  const more = $("#btn-more");
  if (reset) {
    state.paymentsCursor = "";
    box.textContent = "";
    box.append(el("div", { class: "center-block" }, [el("div", { class: "spinner" })]));
  }
  try {
    const r = await rpc("economy_payments_mine", { lang: i18n.lang, limit: 20, cursor: state.paymentsCursor || undefined });
    if (reset) box.textContent = "";
    if (reset && !r.items.length) box.append(el("div", { class: "empty", text: t("noPayments") }));
    for (const p of r.items) {
      const name = typeof p.productName === "string" && p.productName ? p.productName : p.productKey;
      const amount = p.method === "in_app" ? t("coins", p.coinAmount) : t("toman", p.finalToman);
      box.append(el("div", { class: "pay", "data-status": p.status, "data-method": p.method }, [
        el("span", { class: "name", text: name }),
        el("span", { class: "amount", text: amount }),
        el("span", { class: "meta" }, [
          methodLabel(p.method),
          t("listSep"),
          el("bdi", { text: formatDate(p.createdAt) }),
          p.refId ? `${t("listSep")}${t("refId")} ` : null,
          p.refId ? el("bdi", { class: "mono", dir: "ltr", text: p.refId }) : null,
        ]),
        el("span", { class: "status status-" + p.status, text: statusLabel(p.status) }),
      ]));
    }
    state.paymentsCursor = r.cursor || "";
    more.hidden = !r.cursor;
  } catch (err) {
    if (reset) box.textContent = "";
    box.append(el("div", { class: "empty", text: errorText(err) }));
  }
}

async function handleReturn() {
  const params = new URLSearchParams(location.search);
  const paymentId = params.get("payment");
  if (!paymentId) return false;
  history.replaceState(null, "", location.pathname);
  show("loading");
  $("#loading-text").textContent = t("checkingPayment");
  let status = null;
  for (let i = 0; i < 30; i++) {
    try {
      status = await rpc("economy_purchase_status", { paymentId });
    } catch (err) {
      status = { status: "error", error: err.code };
      break;
    }
    if (status.status !== "pending") break;
    await new Promise((r) => setTimeout(r, 2000));
  }
  await refreshBalances().catch(() => {});
  show("home");
  const ok = status && status.status === "success";
  await openDialog((body, close) => {
    body.setAttribute("data-payment-result", status ? status.status : "unknown");
    body.append(
      el("h2", { text: ok ? t("payOk") : status && status.status === "pending" ? t("payPending") : t("payFailed") }),
      el("p", { text: ok ? (status.rewardGranted ? t("payOkText", state.wallet.coin || 0) : errorText({ code: "REWARD_PENDING" })) : status && status.error ? errorText({ code: status.error }) : status && status.status === "pending" ? t("payPendingText") : t("payFailedText") }),
      el("div", { class: "dialog-actions" }, [el("button", { class: "btn btn-primary", type: "button", text: t("ok"), onclick: () => close(true) })])
    );
  });
  return true;
}

function navigate(target) {
  const name = String(target || "").trim();
  if (name === "play") return startGame();
  if (name.startsWith("support/ticket/")) return openTicket(name.slice("support/ticket/".length));
  if (SCREENS.includes(name)) return show(name);
  toast(t("sectionNotFound"), "bad");
}

function runAction(action) {
  if (!action || !action.value) return;
  if (action.type === "url") {
    window.open(action.value, "_blank", "noopener,noreferrer");
    return;
  }
  navigate(action.value);
}

function richText(html) {
  const box = el("div", { class: "rich" });
  box.innerHTML = html || "";
  for (const a of box.querySelectorAll("a")) {
    a.target = "_blank";
    a.rel = "noopener noreferrer";
  }
  return box;
}

function skeletonList(n, cls) {
  return Array.from({ length: n }, () => el("div", { class: "skel " + cls, "aria-hidden": "true" }, [el("span", { class: "skel-block" }), el("span", { class: "skel-line" }), el("span", { class: "skel-line short" })]));
}

function setNewsDot(n) {
  state.newsNew = n;
  const dot = $("#news-dot");
  dot.hidden = !n;
  dot.textContent = n ? num(n) : "";
  $("#menu-news").setAttribute("aria-label", t("newsAria", n));
}

async function refreshNewsDot() {
  try {
    const r = await rpc("publication_news_list", { lang: i18n.lang, limit: 50 });
    setNewsDot(r.items.filter((n) => !n.seen).length);
  } catch (e) {
    setNewsDot(0);
  }
}

async function loadNews(reset) {
  if (state.newsLoading) return;
  state.newsLoading = true;
  const box = $("#news-list");
  const more = $("#btn-news-more");
  if (reset) {
    state.newsCursor = "";
    box.textContent = "";
    box.append(...skeletonList(3, "skel-news"));
  }
  try {
    const r = await rpc("publication_news_list", { lang: i18n.lang, limit: 10, cursor: state.newsCursor || undefined });
    if (reset) box.textContent = "";
    if (reset && !r.items.length) box.append(el("div", { class: "empty", text: t("noNews") }));
    for (const n of r.items) box.append(newsCard(n));
    state.newsCursor = r.cursor || "";
    more.hidden = !r.cursor;
    if (reset) setNewsDot(r.items.filter((n) => !n.seen).length);
  } catch (err) {
    if (reset) box.textContent = "";
    box.append(el("div", { class: "empty", text: errorText(err) }));
  } finally {
    state.newsLoading = false;
  }
}

function newsCard(n) {
  const cover = n.coverUrl ? el("img", { class: "news-cover", src: n.coverUrl, alt: "", loading: "lazy" }) : el("div", { class: "news-cover news-cover-empty", "aria-hidden": "true" }, "★");
  return el("button", { class: "news-card" + (n.seen ? "" : " unread") + (n.isPinned ? " pinned" : ""), type: "button", "data-news": n.id, onclick: () => openArticle(n.id) }, [
    cover,
    el("div", { class: "news-body" }, [
      el("div", { class: "news-tags" }, [
        n.isPinned ? el("span", { class: "chip chip-gold", text: t("pinned") }) : null,
        n.seen ? null : el("span", { class: "chip chip-new" }, [el("span", { class: "dot", "aria-hidden": "true" }), t("fresh")]),
      ]),
      el("h3", { text: n.title }),
      n.summary ? el("p", { class: "muted small", text: n.summary }) : null,
      el("div", { class: "news-meta muted small" }, [el("span", { text: formatDate(n.publishedAt) }), el("span", { text: t("commentCount", n.commentCount) })]),
    ]),
  ]);
}

async function openArticle(id) {
  for (const s of document.querySelectorAll(".screen")) s.hidden = s.id !== "screen-article";
  stopGame();
  window.scrollTo(0, 0);
  const box = $("#article");
  box.textContent = "";
  box.append(el("div", { class: "skel skel-article", "aria-hidden": "true" }, [el("span", { class: "skel-block" }), el("span", { class: "skel-line" }), el("span", { class: "skel-line" }), el("span", { class: "skel-line short" })]));
  $("#comments").hidden = true;
  state.replyTo = null;
  try {
    const a = await rpc("publication_news_get", { id, lang: i18n.lang });
    state.article = a;
    box.textContent = "";
    if (a.coverUrl) box.append(el("img", { class: "article-cover", src: a.coverUrl, alt: "" }));
    box.append(
      el("div", { class: "news-tags" }, [a.isPinned ? el("span", { class: "chip chip-gold", text: t("pinned") }) : null]),
      el("h1", { text: a.title }),
      el("p", { class: "muted small", text: formatDate(a.publishedAt) }),
      a.summary ? el("p", { class: "article-summary", text: a.summary }) : null,
      richText(a.bodyHtml)
    );
    if (a.actions && a.actions.length) {
      box.append(el("div", { class: "article-actions" }, a.actions.map((act, i) => el("button", {
        class: i === 0 ? "btn btn-primary" : "btn",
        type: "button",
        "data-action": act.type,
        onclick: () => runAction(act),
      }, [act.label || act.value, act.type === "url" ? el("span", { class: "ext", "aria-hidden": "true", text: "↗" }) : null]))));
    }
    if (!a.seen) rpc("publication_news_seen", { id: a.id }).then(() => refreshNewsDot()).catch(() => {});
    const c = $("#comments");
    c.hidden = false;
    $("#comment-form").hidden = !a.allowComment;
    $("#comment-hint").textContent = "";
    $("#comments-count").textContent = a.allowComment ? "" : t("commentsClosed");
    loadComments(true);
  } catch (err) {
    box.textContent = "";
    box.append(el("div", { class: "empty", text: errorText(err) }));
  }
}

function commentView(c, root) {
  const node = el("div", { class: "comment" + (c.isAccepted ? "" : " pending") + (c.mine ? " mine" : ""), "data-comment": c.id }, [
    el("div", { class: "comment-head" }, [
      el("bdi", { class: "comment-user", text: c.displayName || c.username || t("player") }),
      c.mine ? el("span", { class: "chip", text: t("you") }) : null,
      el("span", { class: "muted small", text: formatDate(c.createdAt) }),
      c.isAccepted ? null : el("span", { class: "chip chip-wait", text: t("awaitingApproval") }),
    ]),
    el("p", { class: "comment-text", dir: "auto", text: c.message }),
  ]);
  if (root && c.isAccepted && state.article && state.article.allowComment) {
    node.append(el("button", { class: "link-btn", type: "button", text: t("reply"), onclick: () => startReply(c) }));
  }
  if (root && c.replies && c.replies.length) {
    node.append(el("div", { class: "replies" }, c.replies.map((r) => commentView(r, false))));
  }
  return node;
}

function startReply(c) {
  state.replyTo = c;
  $("#comment-hint").textContent = "";
  $("#comment-hint").append(t("replyingTo", c.displayName || c.username || t("player")), el("button", { class: "link-btn", type: "button", text: t("cancel"), onclick: () => { state.replyTo = null; $("#comment-hint").textContent = ""; } }));
  $("#comment-text").focus();
}

async function loadComments(reset) {
  const list = $("#comment-list");
  const more = $("#btn-comments-more");
  if (!state.article) return;
  if (reset) {
    state.commentsCursor = "";
    list.textContent = "";
    list.append(...skeletonList(2, "skel-comment"));
  }
  try {
    const r = await rpc("publication_news_comments", { newsId: state.article.id, limit: 20, cursor: state.commentsCursor || undefined });
    if (reset) list.textContent = "";
    if (reset && !r.items.length) list.append(el("p", { class: "muted small center", id: "no-comments", text: state.article.allowComment ? t("noCommentsYet") : t("noComments") }));
    for (const c of r.items) list.append(commentView(c, true));
    state.commentsCursor = r.cursor || "";
    more.hidden = !r.cursor;
    if (state.article.allowComment) $("#comments-count").textContent = t("commentCount", state.article.commentCount);
  } catch (err) {
    if (reset) list.textContent = "";
    list.append(el("p", { class: "form-msg bad", text: errorText(err) }));
  }
}

async function postComment(e) {
  e.preventDefault();
  const input = $("#comment-text");
  const btn = $("#comment-form button[type=submit]");
  const message = input.value.trim();
  if (!message || !state.article) return;
  btn.disabled = true;
  try {
    const payload = { newsId: state.article.id, message };
    if (state.replyTo) payload.parentId = state.replyTo.id;
    const c = await rpc("publication_news_comment", payload);
    input.value = "";
    const parent = state.replyTo;
    state.replyTo = null;
    $("#comment-hint").textContent = "";
    const empty = $("#no-comments");
    if (empty) empty.remove();
    const view = commentView(c, !c.parentId);
    if (c.parentId) {
      const host = document.querySelector(`[data-comment="${c.parentId}"]`) || (parent && document.querySelector(`[data-comment="${parent.id}"]`));
      let replies = host && host.querySelector(".replies");
      if (host && !replies) {
        replies = el("div", { class: "replies" });
        host.append(replies);
      }
      if (replies) replies.append(view);
      else $("#comment-list").prepend(view);
    } else {
      $("#comment-list").prepend(view);
    }
    toast(c.isAccepted ? t("commentLive") : t("commentQueued"), "ok");
  } catch (err) {
    toast(errorText(err), "bad");
  } finally {
    btn.disabled = false;
  }
}

let faqTimer = null;
let faqSeq = 0;
async function loadFaq() {
  const my = ++faqSeq;
  const box = $("#faq-list");
  box.textContent = "";
  box.append(...skeletonList(4, "skel-faq"));
  try {
    const q = $("#faq-q").value.trim();
    const r = await rpc("publication_faq_list", { lang: i18n.lang, q: q || undefined, categoryKey: state.faqCategory || undefined });
    if (my !== faqSeq) return;
    const tabs = $("#faq-tabs");
    tabs.textContent = "";
    if (r.categories.length) {
      const all = [{ key: "", name: t("all") }].concat(r.categories);
      for (const c of all) {
        tabs.append(el("button", { class: "tab", type: "button", role: "tab", "aria-selected": c.key === state.faqCategory ? "true" : "false", "data-cat": c.key, onclick: () => { state.faqCategory = c.key; loadFaq(); } }, c.name));
      }
    }
    box.textContent = "";
    if (!r.items.length) {
      box.append(el("div", { class: "empty", text: q ? t("noFaqMatch") : t("noFaq") }));
      return;
    }
    for (const f of r.items) {
      box.append(el("details", { class: "faq", "data-faq": f.id }, [
        el("summary", {}, [el("span", { text: f.question }), el("span", { class: "chev", "aria-hidden": "true", text: "⌄" })]),
        richText(f.answerHtml),
      ]));
    }
  } catch (err) {
    if (my !== faqSeq) return;
    box.textContent = "";
    box.append(el("div", { class: "empty", text: errorText(err) }));
  }
}

function setUnread(n) {
  state.unread = Math.max(0, Number(n) || 0);
  const b = $("#inbox-badge");
  b.hidden = !state.unread;
  b.textContent = state.unread > 99 ? "+" + num(99) : num(state.unread);
  $("#btn-inbox").setAttribute("aria-label", t("inboxAria", state.unread));
}

async function refreshUnread() {
  try {
    const r = await rpc("publication_inbox", { lang: i18n.lang, limit: 1 });
    setUnread(r.unread);
  } catch (e) {
    return;
  }
}

function inboxItem(m) {
  const node = el("div", { class: "inbox-item" + (m.seen ? "" : " unread"), "data-inbox": m.id }, [
    el("div", { class: "inbox-head" }, [
      m.seen ? null : el("span", { class: "dot", "aria-label": t("unread") }),
      el("h3", { text: m.title || t("message") }),
      el("span", { class: "muted small", text: formatDate(m.createdAt) }),
    ]),
    m.category ? el("span", { class: "chip inbox-cat", "data-inbox-category": m.category.key, text: m.category.name }) : null,
    m.imageUrl ? el("img", { class: "inbox-img", src: m.imageUrl, alt: "", loading: "lazy", decoding: "async" }) : null,
    richText(m.bodyHtml),
  ]);
  if (m.action && m.action.value) {
    const label = m.action.label || (m.action.type === "url" ? t("openLink") : t("goTo", SCREENS.includes(m.action.value) ? t(m.action.value) : m.action.value));
    node.append(el("button", { class: "btn", type: "button", "data-action": m.action.type, text: label, onclick: () => {
      if (typeof markOpened === "function" && !m.opened) {
        m.opened = true;
        markOpened(m.id);
      } else if (!m.seen) rpc("publication_inbox_seen", { ids: [m.id] }).then((r) => setUnread(r.unread)).catch(() => {});
      if (typeof runNoteAction === "function" && ["screen", "product", "achievement"].includes(m.action.type)) runNoteAction(m.action);
      else runAction(m.action);
    } }));
  }
  return node;
}

async function loadInbox(reset) {
  const box = $("#inbox-list");
  const more = $("#btn-inbox-more");
  if (reset) {
    state.inboxCursor = "";
    box.textContent = "";
    box.append(...skeletonList(3, "skel-inbox"));
  }
  try {
    const r = await rpc("publication_inbox", { lang: i18n.lang, limit: 20, cursor: state.inboxCursor || undefined });
    if (reset) box.textContent = "";
    if (reset && !r.items.length) box.append(el("div", { class: "empty", text: t("inboxEmpty") }));
    for (const m of r.items) box.append(inboxItem(m));
    state.inboxCursor = r.cursor || "";
    more.hidden = !r.cursor;
    setUnread(r.unread);
    $("#btn-seen-all").disabled = !r.unread;
  } catch (err) {
    if (reset) box.textContent = "";
    box.append(el("div", { class: "empty", text: errorText(err) }));
  }
}

async function seenAll() {
  const btn = $("#btn-seen-all");
  btn.disabled = true;
  try {
    const r = await rpc("publication_inbox_seen", { all: true });
    setUnread(r.unread);
    for (const n of document.querySelectorAll(".inbox-item.unread")) {
      n.classList.remove("unread");
      const d = n.querySelector(".dot");
      if (d) d.remove();
    }
    toast(t("allRead"), "ok");
  } catch (err) {
    btn.disabled = false;
    toast(errorText(err), "bad");
  }
}

let noteTimer = null;
function showNote(n) {
  const t = $("#note-toast");
  t.textContent = "";
  t.append(
    el("span", { class: "note-icon", "aria-hidden": "true" }, "★"),
    el("span", { class: "note-text" }, [el("b", { text: n.title || t("newMessage") }), n.body ? el("span", { class: "muted small", text: plainText(n.body) }) : null])
  );
  t.hidden = false;
  t.classList.remove("in");
  void t.offsetWidth;
  t.classList.add("in");
  clearTimeout(noteTimer);
  noteTimer = setTimeout(() => (t.hidden = true), 6000);
}

function plainText(md) {
  return String(md || "").replace(/!\[[^\]]*\]\([^)]*\)/g, "").replace(/\[([^\]]*)\]\([^)]*\)/g, "$1").replace(/[*#]/g, "").replace(/\s+/g, " ").trim().slice(0, 140);
}

async function liveNotification(id, fallback) {
  let note = fallback;
  let unread = state.unread + 1;
  for (let i = 0; i < 4; i++) {
    try {
      const r = await rpc("publication_inbox", { lang: i18n.lang, limit: 10 });
      const found = r.items.find((m) => m.id === id);
      unread = Math.max(r.unread, found ? 0 : state.unread + 1);
      if (found) {
        note = { title: found.title, body: found.body };
        break;
      }
    } catch (e) {
      break;
    }
    await new Promise((r) => setTimeout(r, 500));
  }
  showNote(note);
  setUnread(unread);
  const bell = $("#btn-inbox");
  bell.classList.remove("ring");
  void bell.offsetWidth;
  bell.classList.add("ring");
  if (!$("#screen-inbox").hidden) loadInbox(true);
}

function connectSocket() {
  if (state.socket) {
    state.socket.onclose = null;
    state.socket.close();
    state.socket = null;
  }
  if (!state.token) return;
  const ws = new WebSocket(`${API.replace(/^http/, "ws")}/ws?token=${encodeURIComponent(state.token)}&format=json`);
  state.socket = ws;
  ws.onopen = () => {
    state.socketRetry = 0;
    document.body.dataset.socket = "open";
  };
  ws.onmessage = (ev) => {
    let msg;
    try {
      msg = JSON.parse(ev.data);
    } catch (e) {
      return;
    }
    if (msg && msg.cid && state.rt.pending.has(msg.cid)) {
      const p = state.rt.pending.get(msg.cid);
      state.rt.pending.delete(msg.cid);
      if (msg.error) p.reject(msg.error);
      else p.resolve(msg);
      return;
    }
    if (typeof lobbySocket === "function" && lobbySocket(msg)) return;
    if (msg && msg.channel_message) {
      chatIncoming(msg.channel_message);
      return;
    }
    const list = msg && msg.notifications && msg.notifications.notifications;
    if (!list) return;
    for (const n of list) {
      if (n.code >= 1102 && n.code <= 1105) {
        moderationNotice(n);
        continue;
      }
      if (n.code === 1210 || n.code === 1211) {
        if (typeof supportNotice === "function") supportNotice(n);
        continue;
      }
      if (n.code !== 1001) continue;
      let content = {};
      try {
        content = typeof n.content === "string" ? JSON.parse(n.content) : n.content || {};
      } catch (e) {
        content = {};
      }
      liveNotification(content.notificationId, { title: n.subject, body: "" });
    }
  };
  ws.onclose = () => {
    document.body.dataset.socket = "closed";
    for (const p of state.rt.pending.values()) p.reject({ code: 0, message: "NETWORK" });
    state.rt.pending.clear();
    state.chat.joined = false;
    if (state.socket !== ws) return;
    state.socket = null;
    if (state.token && !state.suspended) rpc("base_my_status", { lang: i18n.lang }).catch(() => {});
    const wait = Math.min(30000, 1000 * 2 ** state.socketRetry++);
    setTimeout(() => { if (!state.socket) connectSocket(); }, wait);
  };
}

function applyStatic() {
  const root = document.documentElement;
  root.lang = i18n.lang;
  root.dir = i18n.dir;
  root.classList.toggle("ui-fallback", !STRINGS[i18n.lang]);
  document.title = t("appName");
  for (const n of document.querySelectorAll("[data-i18n]")) n.textContent = t(n.dataset.i18n);
  for (const n of document.querySelectorAll("[data-i18n-html]")) n.innerHTML = t(n.dataset.i18nHtml);
  for (const n of document.querySelectorAll("[data-i18n-aria]")) n.setAttribute("aria-label", t(n.dataset.i18nAria));
  for (const n of document.querySelectorAll("[data-i18n-ph]")) n.setAttribute("placeholder", t(n.dataset.i18nPh));
  if (state.token) $("#balance").textContent = num(state.wallet.coin || 0);
  $("#hud-score").textContent = num(game.score);
  $("#hud-time").textContent = num(Math.ceil(Math.max(0, game.time)));
  setUnread(state.unread);
  setNewsDot(state.newsNew);
  renderLangPicker();
  $("#game-version").textContent = t("versionLabel", gameVersion());
  renderUpdateBanner();
  if (!$("#screen-account").hidden) renderAccount();
  if (!$("#screen-update").hidden) renderUpdate();
  if (typeof paintConfig === "function") paintConfig();
}

function renderLangPicker() {
  const box = $("#lang-picker");
  const list = $("#lang-options");
  list.textContent = "";
  box.hidden = state.languages.length < 2;
  for (const l of state.languages) {
    const current = l.code === i18n.lang;
    list.append(el("button", {
      class: "lang-opt",
      type: "button",
      lang: l.code,
      dir: l.direction,
      "data-lang": l.code,
      "aria-pressed": current ? "true" : "false",
      title: l.name || l.nativeName,
      onclick: () => changeLanguage(l.code),
    }, [current ? el("span", { class: "lang-check", "aria-hidden": "true", text: "✓" }) : null, l.nativeName || l.code]));
  }
}

function useLanguage(item) {
  setLocale(item.code, item.direction);
  localStorage.setItem(LS.lang, item.code);
  localStorage.setItem(LS.langDir, i18n.dir);
  applyStatic();
}

function pickLanguage(r) {
  const codes = r.items.map((l) => l.code);
  const saved = localStorage.getItem(LS.lang);
  if (saved && codes.includes(saved)) return saved;
  for (const raw of navigator.languages || [navigator.language || ""]) {
    const code = String(raw).toLowerCase();
    if (codes.includes(code)) return code;
    const primary = code.split("-")[0];
    if (codes.includes(primary)) return primary;
  }
  return codes.includes(r.default) ? r.default : codes[0];
}

async function initLanguage() {
  let r;
  try {
    r = await rpc("base_languages", localStorage.getItem(LS.lang) ? { lang: localStorage.getItem(LS.lang) } : {});
  } catch (e) {
    return;
  }
  state.languages = r.items || [];
  if (!state.languages.length) return;
  const code = pickLanguage(r);
  useLanguage(state.languages.find((l) => l.code === code));
  try {
    await rpc("base_set_language", { lang: code });
  } catch (e) {
    return;
  }
}

function currentScreen() {
  const s = Array.from(document.querySelectorAll(".screen")).find((n) => !n.hidden);
  return s ? s.id.replace("screen-", "") : "home";
}

async function changeLanguage(code) {
  const item = state.languages.find((l) => l.code === code);
  if (!item || code === i18n.lang) return;
  useLanguage(item);
  $("#toast").hidden = true;
  $("#note-toast").hidden = true;
  try {
    await rpc("base_set_language", { lang: code });
  } catch (err) {
    toast(errorText(err), "bad");
  }
  state.catalog = null;
  refreshUnread();
  refreshNewsDot();
  const screen = currentScreen();
  if (screen === "article" && state.article) openArticle(state.article.id);
  else if (screen === "home") refreshBalances().then(renderHome).catch(() => renderHome());
  if (screen === "home" && typeof loadMenuBanners === "function") loadMenuBanners(true);
  else if (["shop", "payments", "news", "faq", "inbox", "support", "referral", "activity", "livechat"].includes(screen)) show(screen);
  else refreshBalances().catch(() => {});
  checkVersion().then(() => { if (state.version && state.version.status === "update_required") show("update"); }).catch(() => {});
  loadCountries();
  if (typeof loadConfig === "function") loadConfig().then(() => { if (currentScreen() === "home") renderHome(); });
}

async function boot() {
  const saved = localStorage.getItem(LS.lang);
  if (saved) setLocale(saved, localStorage.getItem(LS.langDir));
  applyStatic();
  setupInput();
  for (const b of document.querySelectorAll("[data-go]")) b.addEventListener("click", () => show(b.dataset.go));
  $("#btn-play").addEventListener("click", startGame);
  $("#btn-quit").addEventListener("click", () => show("home"));
  $("#btn-restore").addEventListener("click", restoreDialog);
  $("#btn-more").addEventListener("click", () => loadPayments(false));
  $("#gift-form").addEventListener("submit", redeemGift);
  $("#btn-inbox").addEventListener("click", () => show("inbox"));
  $("#btn-seen-all").addEventListener("click", seenAll);
  $("#btn-inbox-more").addEventListener("click", () => loadInbox(false));
  $("#btn-news-more").addEventListener("click", () => loadNews(false));
  $("#btn-comments-more").addEventListener("click", () => loadComments(false));
  $("#comment-form").addEventListener("submit", postComment);
  $("#note-toast").addEventListener("click", () => { $("#note-toast").hidden = true; show("inbox"); });
  $("#faq-q").addEventListener("input", () => { clearTimeout(faqTimer); faqTimer = setTimeout(loadFaq, 300); });
  setupAccount();
  $("#btn-retry").addEventListener("click", connect);
  $("#btn-suspended-retry").addEventListener("click", () => { state.suspended = null; connect(); });
  $("#chat-form").addEventListener("submit", sendChat);
  await connect();
}

async function connect() {
  const retry = $("#btn-retry");
  retry.hidden = true;
  document.querySelector("#screen-loading .spinner").hidden = false;
  $("#loading-text").textContent = t("loggingIn");
  try {
    await loadServerKey();
    await login();
    await initLanguage();
    await checkVersion();
    if (state.version && state.version.status === "update_required") {
      show("update");
      return;
    }
    await refreshBalances();
  } catch (err) {
    if (err.code === "ACCOUNT_SUSPENDED") {
      showSuspended(err.detail);
      return;
    }
    $("#loading-text").textContent = errorText(err);
    document.querySelector("#screen-loading .spinner").hidden = true;
    retry.hidden = false;
    retry.focus();
    return;
  }
  loadMyTags();
  loadCountries();
  loadPhone();
  refreshUnread();
  if (typeof registerPush === "function") registerPush(false);
  refreshNewsDot();
  if (typeof loadConfig === "function") await loadConfig();
  if (typeof refreshSupportDot === "function") refreshSupportDot();
  if (!(await handleReturn())) show("home");
}


function gameVersion() {
  const v = new URLSearchParams(location.search).get("v");
  return v && /^\d{1,9}(\.\d{1,9}){0,3}$/.test(v) ? v : GAME_VERSION;
}

async function checkVersion() {
  try {
    state.version = await rpc("base_app_version", { platform: PLATFORM, version: gameVersion(), store: "web", lang: i18n.lang });
  } catch (err) {
    state.version = null;
  }
  renderUpdateBanner();
  return state.version;
}

function renderUpdateBanner() {
  const box = $("#update-banner");
  const v = state.version;
  const show = !!v && v.status === "update_available" && localStorage.getItem(LS.skipUpdate) !== v.latest;
  box.hidden = !show;
  if (!show) return;
  $("#update-banner-text").textContent = v.message || t("versions", gameVersion(), v.latest);
  const go = $("#update-banner-go");
  if (v.storeUrl) {
    go.href = v.storeUrl;
    go.hidden = false;
  } else go.hidden = true;
}

function renderUpdate() {
  const v = state.version || {};
  document.body.classList.toggle("blocked", v.status === "update_required");
  $("#update-text").textContent = v.message || t("updateDefault");
  $("#update-versions").textContent = t("versions", gameVersion(), v.minimum || "");
  const box = $("#update-stores");
  box.textContent = "";
  for (const st of v.stores || []) box.append(el("a", { class: "btn btn-primary", href: st.url, target: "_blank", rel: "noopener", "data-store": st.store, text: st.name }));
}

async function loadCountries() {
  try {
    const r = await rpc("base_countries", { lang: i18n.lang });
    state.countries = r.items || [];
    if (r.selected) state.country = r.selected;
  } catch (err) {
    state.countries = [];
  }
  if (!state.otp.country || !state.countries.some((c) => c.code === state.otp.country)) state.otp.country = defaultCountry();
  if (!$("#screen-account").hidden) renderAccount();
}

function defaultCountry() {
  const codes = state.countries.map((c) => c.code);
  if (state.country && codes.includes(state.country)) return state.country;
  if (codes.includes("IR")) return "IR";
  return codes[0] || "";
}

function countryOf(code) {
  return state.countries.find((c) => c.code === code) || null;
}

function flagNode(c) {
  if (c && c.flagUrl) return el("img", { class: "flag", src: c.flagUrl, alt: "", width: "24", height: "18" });
  return el("span", { class: "flag", "aria-hidden": "true", text: c ? c.flag : "🏳️" });
}

function paintCountryButton(btn, code, withDial) {
  const c = countryOf(code);
  btn.textContent = "";
  btn.dataset.country = code || "";
  btn.append(flagNode(c));
  if (!withDial) btn.append(el("span", { class: "name", text: c ? c.name : t("chooseCountry") }));
  if (withDial && c) btn.append(el("bdi", { class: "dial", dir: "ltr", text: c.dialCode }));
  btn.setAttribute("aria-label", t("chooseCountry") + (c ? ": " + c.name + " " + c.dialCode : ""));
}

function countryPicker(current) {
  return openDialog((body, close) => {
    const input = el("input", { type: "search", class: "country-search", id: "country-q", placeholder: t("searchCountry"), "aria-label": t("searchCountry"), autocomplete: "off" });
    const list = el("div", { class: "country-list", role: "listbox", id: "country-list", "aria-label": t("chooseCountry") });
    const paint = () => {
      const q = input.value.trim().toLowerCase().replace(/^\+/, "");
      list.textContent = "";
      const hits = state.countries.filter((c) => !q || c.name.toLowerCase().includes(q) || c.code.toLowerCase() === q || c.dialCode.replace("+", "").startsWith(q));
      for (const c of hits.slice(0, 80)) {
        list.append(el("button", { class: "country-opt", type: "button", role: "option", "aria-selected": c.code === current ? "true" : "false", "data-code": c.code, onclick: () => close(c.code) }, [
          flagNode(c), el("span", { class: "name", text: c.name }), el("bdi", { class: "dial", dir: "ltr", text: c.dialCode }),
        ]));
      }
      if (!hits.length) list.append(el("p", { class: "muted center", text: t("noCountry") }));
    };
    input.addEventListener("input", paint);
    body.append(el("h2", { text: t("chooseCountry") }), input, list, el("div", { class: "dialog-actions" }, [el("button", { class: "btn", type: "button", text: t("close"), onclick: () => close(undefined) })]));
    paint();
    setTimeout(() => input.focus(), 0);
  });
}

async function loadPhone() {
  try {
    const r = await rpc("publication_phone", {});
    state.phone = r.phone || "";
    state.otp.codeLength = r.codeLength || state.otp.codeLength;
  } catch (err) {
    state.phone = "";
  }
  if (!$("#screen-account").hidden) renderAccount();
}

async function loadRules() {
  if (state.rules) return state.rules;
  try {
    state.rules = await rpc("base_account_rules", {});
  } catch (err) {
    state.rules = null;
  }
  return state.rules;
}

function renderAccount() {
  paintTags($("#account-tags"));
  $("#account-tags-panel").hidden = !state.myTags.length;
  const st = $("#account-status");
  st.textContent = "";
  st.append(
    el("span", { class: "avatar", "aria-hidden": "true", text: state.phone ? "★" : "?" }),
    el("div", {}, [
      el("b", { id: "account-title", text: state.phone ? t("signedInAs", "⁦" + state.phone + "⁩") : t("guest") }),
      el("span", { class: "muted small", text: state.phone ? t("signedInHint") : t("guestHint") }),
    ])
  );
  paintCountryButton($("#btn-country"), state.country || "", false);
  paintCountryButton($("#otp-country"), state.otp.country, true);
  $("#otp-form").hidden = !!state.phone;
  $("#btn-signout").hidden = !state.phone;
  $("#otp-link").closest(".check-row").hidden = !!localStorage.getItem(LS.loginKey);
  const code = state.otp.step === "code";
  $("#otp-step-phone").hidden = code;
  $("#otp-step-code").hidden = !code;
  if (code) {
    $("#otp-sent").textContent = t("codeSent", "⁦" + state.otp.phone + "⁩") + (state.otp.sandbox ? " " + t("codeSandbox") : "");
    $("#otp-code").setAttribute("maxlength", String(state.otp.codeLength));
    $("#otp-code").setAttribute("placeholder", "•".repeat(state.otp.codeLength));
    $("#otp-code").setAttribute("aria-label", t("codeDigits", state.otp.codeLength));
  }
  paintResend();
  const name = $("#name-input");
  if (document.activeElement !== name) name.value = state.username;
  loadRules().then((r) => { $("#name-hint").textContent = r ? t("nameRule", r.usernameMin, r.usernameMax, r.usernameCharset === "latin") : ""; });
}

function paintResend() {
  const btn = $("#otp-resend");
  const left = Math.ceil((state.otp.until - Date.now()) / 1000);
  btn.disabled = left > 0;
  btn.textContent = left > 0 ? t("resendIn", left) : t("resend");
  clearTimeout(state.otp.timer);
  if (left > 0) state.otp.timer = setTimeout(paintResend, 1000);
}

function otpMsg(text, kind) {
  const m = $("#otp-msg");
  m.className = "form-msg" + (kind ? " " + kind : "");
  m.textContent = text || "";
}

async function sendCode() {
  const phone = $("#otp-phone").value.trim();
  if (!phone) {
    otpMsg(errorText({ code: "INVALID_PHONE" }), "bad");
    $("#otp-phone").focus();
    return;
  }
  const btn = state.otp.step === "code" ? $("#otp-resend") : $("#otp-send");
  btn.disabled = true;
  otpMsg(t("sending"));
  try {
    const r = await rpc("publication_otp_request", { phone, country: state.otp.country, lang: i18n.lang });
    Object.assign(state.otp, { step: "code", phone: r.phone, until: Date.now() + (r.resendAfter || 60) * 1000, codeLength: r.codeLength || 5, sandbox: !!r.sandbox });
    otpMsg("");
    $("#otp-code").value = "";
    renderAccount();
    $("#otp-code").focus();
  } catch (err) {
    otpMsg(errorText(err), "bad");
  } finally {
    btn.disabled = false;
    paintResend();
  }
}

async function verifyCode() {
  const code = $("#otp-code").value.replace(/[۰-۹]/g, (d) => String(d.charCodeAt(0) - 0x06f0)).replace(/\D/g, "");
  if (!code) {
    otpMsg(errorText({ code: "OTP_INVALID_CODE" }), "bad");
    return;
  }
  const btn = $("#otp-verify");
  btn.disabled = true;
  otpMsg(t("verifying"));
  try {
    const keep = !localStorage.getItem(LS.loginKey) && $("#otp-link").checked;
    const r = await rpc("publication_otp_verify", { phone: state.otp.phone, country: state.otp.country, code, link: keep });
    if (r.loginKey) localStorage.setItem(LS.loginKey, r.loginKey);
    useSession(r.token);
    state.phone = r.phone;
    rpc("base_set_language", { lang: i18n.lang }).catch(() => {});
    Object.assign(state.otp, { step: "phone", phone: "", until: 0 });
    otpMsg("");
    $("#otp-phone").value = "";
    const shown = "⁦" + r.phone + "⁩";
    toast(r.created ? t("newAccount", shown) : r.linked ? t("linked", shown) : t("welcomeBack", shown), "ok");
    await refreshBalances().catch(() => {});
    loadCountries();
    refreshUnread();
    refreshNewsDot();
    renderAccount();
  } catch (err) {
    otpMsg(errorText(err), "bad");
    if (err.code === "OTP_EXPIRED" || err.code === "OTP_TOO_MANY_ATTEMPTS" || err.code === "OTP_NOT_FOUND") state.otp.until = 0;
    paintResend();
  } finally {
    btn.disabled = false;
  }
}

async function signOut() {
  localStorage.removeItem(LS.loginKey);
  state.phone = "";
  try {
    await login();
    await refreshBalances();
  } catch (err) {
    toast(errorText(err), "bad");
  }
  toast(t("signedOut"), "ok");
  loadCountries();
  renderAccount();
}

async function saveName(e) {
  e.preventDefault();
  const name = $("#name-input").value.trim();
  const msg = $("#name-msg");
  msg.className = "form-msg";
  msg.textContent = "";
  try {
    await request("/v2/account", { method: "PUT", headers: { Authorization: "Bearer " + state.token, "Content-Type": "application/json" }, body: JSON.stringify({ username: name }) });
    await login();
    msg.className = "form-msg ok";
    msg.textContent = t("nameSaved");
  } catch (err) {
    msg.className = "form-msg bad";
    msg.textContent = errorText(err.status === 409 ? { code: "USERNAME_TAKEN" } : err);
  }
}

async function pickMyCountry() {
  const code = await countryPicker(state.country);
  if (!code || code === state.country) return;
  try {
    await rpc("base_set_country", { code });
    state.country = code;
    if (!state.otp.phone) state.otp.country = code;
    renderAccount();
    const c = countryOf(code);
    toast(t("countrySaved", c ? c.name : code), "ok");
  } catch (err) {
    toast(errorText(err), "bad");
  }
}

async function pickOtpCountry() {
  const code = await countryPicker(state.otp.country);
  if (!code) return;
  state.otp.country = code;
  renderAccount();
  $("#otp-phone").focus();
}

function setupAccount() {
  $("#btn-country").addEventListener("click", pickMyCountry);
  $("#otp-country").addEventListener("click", pickOtpCountry);
  $("#otp-form").addEventListener("submit", (e) => {
    e.preventDefault();
    if (state.otp.step === "code") verifyCode();
    else sendCode();
  });
  $("#otp-resend").addEventListener("click", sendCode);
  $("#otp-change").addEventListener("click", () => {
    Object.assign(state.otp, { step: "phone", until: 0 });
    otpMsg("");
    renderAccount();
    $("#otp-phone").focus();
  });
  $("#btn-signout").addEventListener("click", signOut);
  $("#name-form").addEventListener("submit", saveName);
  $("#update-banner-later").addEventListener("click", () => {
    if (state.version && state.version.latest) localStorage.setItem(LS.skipUpdate, state.version.latest);
    renderUpdateBanner();
  });
}

boot();

function paintTags(box) {
  if (!box) return;
  box.textContent = "";
  box.hidden = !state.myTags.length;
  for (const tg of state.myTags) box.append(el("span", { class: "my-tag", "data-tag": tg.key, text: tg.name || tg.key }));
}

async function loadMyTags() {
  try {
    const r = await rpc("base_my_tags", { lang: i18n.lang });
    state.myTags = r.items || [];
  } catch (e) {
    state.myTags = [];
  }
  paintTags($("#my-tags"));
  if (!$("#screen-account").hidden) renderAccount();
}

function showSuspended(detail) {
  state.suspended = detail || { code: "ACCOUNT_SUSPENDED" };
  if (state.socket) {
    state.socket.onclose = null;
    state.socket.close();
    state.socket = null;
  }
  state.token = null;
  show("suspended");
}

function renderSuspended() {
  const d = state.suspended || {};
  $("#suspended-reason").textContent = d.reason || t("suspendedTitle");
  $("#suspended-until").textContent = d.until ? t("suspendedUntil", formatDate(d.until)) : "";
}

function rtSend(body) {
  const ws = state.socket;
  if (!ws || ws.readyState !== WebSocket.OPEN) return Promise.reject({ code: 0, message: "NETWORK" });
  const cid = String(++state.rt.cid);
  return new Promise((resolve, reject) => {
    state.rt.pending.set(cid, { resolve, reject });
    ws.send(JSON.stringify(Object.assign({ cid }, body)));
    setTimeout(() => {
      if (state.rt.pending.has(cid)) {
        state.rt.pending.delete(cid);
        reject({ code: 0, message: "TIMEOUT" });
      }
    }, 10000);
  });
}

async function socketReady() {
  for (let i = 0; i < 50; i++) {
    if (state.socket && state.socket.readyState === WebSocket.OPEN) return true;
    if (!state.socket && state.token) connectSocket();
    await new Promise((r) => setTimeout(r, 100));
  }
  return false;
}

function chatErrorText(err) {
  const notice = parseNotice(err && err.message);
  if (notice) return notice.code === "CHAT_MUTED" ? t("mutedUntil", formatDate(notice.until)) : errorText(new ApiError(notice.code, 0));
  const code = err && typeof err.message === "string" && /^[A-Z_]+$/.test(err.message) ? err.message : "NETWORK";
  return errorText(new ApiError(code, 0));
}

function paintMute() {
  const banner = $("#chat-banner");
  const m = state.chat.mute;
  banner.textContent = "";
  banner.hidden = !m;
  $("#chat-input").disabled = !!m;
  $("#chat-send").disabled = !!m;
  if (!m) return;
  banner.append(el("b", { text: t("mutedTitle") }), el("span", { text: t("mutedUntil", formatDate(m.until)) }), m.reason ? el("span", { class: "muted small", text: m.reason }) : null);
}

async function refreshMute() {
  try {
    const r = await rpc("base_my_status", { lang: i18n.lang });
    state.chat.mute = r.mute || null;
  } catch (e) {
    state.chat.mute = null;
  }
  paintMute();
}

function chatText(content) {
  let c = content;
  if (typeof c === "string") {
    try {
      c = JSON.parse(c);
    } catch (e) {
      return c;
    }
  }
  return c && typeof c.text === "string" ? c.text : "";
}

function chatLine(m) {
  const mine = m.sender_id === decodeJwt(state.token || "").uid;
  const node = el("div", { class: "chat-msg" + (mine ? " mine" : ""), "data-message-id": m.message_id }, [
    el("div", { class: "chat-meta" }, [el("bdi", { class: "chat-name", text: mine ? t("chatYou") : m.username || "" }), el("span", { class: "muted small", text: formatDate(m.create_time) })]),
    el("p", { class: "chat-text", dir: "auto", text: chatText(m.content) }),
  ]);
  if (!mine) {
    node.append(el("button", { class: "chat-report", type: "button", "data-report": m.message_id, "aria-label": t("reportMessage", m.username || ""), title: t("report"), onclick: () => reportDialog(m) }, [el("span", { "aria-hidden": "true", text: "⚑" }), el("span", { class: "sr-only", text: t("report") })]));
  }
  return node;
}

function chatIncoming(m) {
  if (!state.chat.channelId || m.channel_id !== state.chat.channelId) return;
  const list = $("#chat-list");
  if (Number(m.code) === 2) {
    const node = list.querySelector(`[data-message-id="${CSS.escape(m.message_id)}"]`);
    if (node) {
      node.classList.add("removed");
      node.querySelector(".chat-text").textContent = t("messageRemoved");
      const btn = node.querySelector(".chat-report");
      if (btn) btn.remove();
    }
    return;
  }
  if (Number(m.code || 0) !== 0 || state.chat.ids.has(m.message_id)) return;
  state.chat.ids.add(m.message_id);
  const empty = list.querySelector(".empty");
  if (empty) empty.remove();
  list.append(chatLine(m));
  list.scrollTop = list.scrollHeight;
}

async function openChat() {
  const list = $("#chat-list");
  list.textContent = "";
  list.append(el("div", { class: "empty", text: t("chatJoining") }));
  state.chat.ids = new Set();
  refreshMute();
  if (!(await socketReady())) {
    list.textContent = "";
    list.append(el("div", { class: "empty", text: t("chatOffline") }));
    return;
  }
  try {
    const r = await rtSend({ channel_join: { target: CHAT_ROOM, type: 1, persistence: true, hidden: false } });
    state.chat.channelId = r.channel.id;
    state.chat.joined = true;
    const hist = await request(`/v2/channel/${encodeURIComponent(r.channel.id)}?limit=30&forward=false`, { headers: { Authorization: "Bearer " + state.token } });
    list.textContent = "";
    const msgs = ((hist && hist.messages) || []).filter((m) => Number(m.code || 0) === 0).reverse();
    for (const m of msgs) {
      state.chat.ids.add(m.message_id);
      list.append(chatLine(m));
    }
    if (!msgs.length) list.append(el("div", { class: "empty", text: t("chatEmpty") }));
    list.scrollTop = list.scrollHeight;
  } catch (err) {
    list.textContent = "";
    list.append(el("div", { class: "empty", text: err instanceof ApiError ? errorText(err) : chatErrorText(err) }));
  }
}

async function sendChat(e) {
  e.preventDefault();
  const input = $("#chat-input");
  const text = input.value.trim();
  if (!text || state.chat.mute) return;
  const btn = $("#chat-send");
  btn.disabled = true;
  try {
    if (!state.chat.joined) await openChat();
    await rtSend({ channel_message_send: { channel_id: state.chat.channelId, content: JSON.stringify({ text }) } });
    input.value = "";
  } catch (err) {
    const notice = parseNotice(err && err.message);
    if (notice && notice.code === "CHAT_MUTED") {
      state.chat.mute = notice;
      paintMute();
    }
    toast(chatErrorText(err), "bad");
  } finally {
    btn.disabled = !!state.chat.mute;
    input.focus();
  }
}

async function reportTitles() {
  if (state.chat.titles && state.chat.titles.lang === i18n.lang) return state.chat.titles.items;
  const r = await rpc("base_report_titles", { lang: i18n.lang });
  state.chat.titles = { lang: i18n.lang, items: r.items || [] };
  return state.chat.titles.items;
}

async function reportDialog(m) {
  let titles;
  try {
    titles = await reportTitles();
  } catch (err) {
    toast(errorText(err), "bad");
    return;
  }
  await openDialog((body, close) => {
    const msg = el("p", { class: "form-msg", role: "status" });
    const list = el("div", { class: "report-reasons", role: "radiogroup", "aria-label": t("reportTitle") });
    for (const r of titles) {
      list.append(el("label", { class: "report-reason" }, [el("input", { type: "radio", name: "report-title", value: r.key }), el("span", {}, [el("b", { text: r.name }), r.description ? el("span", { class: "muted small", text: r.description }) : null])]));
    }
    const note = el("textarea", { rows: "2", maxlength: "500", id: "report-note", "aria-label": t("reportNote"), placeholder: t("reportNote") });
    const send = el("button", { class: "btn btn-primary", type: "button", id: "report-send", text: t("reportSend"), onclick: async () => {
      const picked = list.querySelector("input:checked");
      if (!picked) {
        msg.textContent = t("reportPick");
        return;
      }
      send.disabled = true;
      try {
        const r = await rpc("base_report_message", { messageId: m.message_id, titleKey: picked.value, note: note.value.trim() });
        close(true);
        toast(r.duplicate ? t("reportDuplicate") : t("reportDone"), "ok");
      } catch (err) {
        send.disabled = false;
        msg.textContent = errorText(err);
      }
    } });
    body.append(
      el("h2", { text: t("reportTitle") }),
      el("blockquote", { class: "report-quote", dir: "auto", text: chatText(m.content) }),
      list,
      note,
      msg,
      el("div", { class: "dialog-actions" }, [send, el("button", { class: "btn", type: "button", text: t("cancel"), onclick: () => close(false) })])
    );
  });
}

function moderationNotice(n) {
  let content = {};
  try {
    content = typeof n.content === "string" ? JSON.parse(n.content) : n.content || {};
  } catch (e) {
    content = {};
  }
  if (n.code === 1102) {
    showSuspended(content);
    return;
  }
  if (n.code === 1103) {
    state.chat.mute = content;
    paintMute();
    toast(t("mutedUntil", formatDate(content.until)), "bad");
    return;
  }
  if (n.code === 1104) {
    openDialog((body, close) => {
      body.append(
        el("h2", { text: t("warnedTitle") }),
        el("p", { class: "warn-text", text: content.reason || n.subject || "" }),
        el("div", { class: "dialog-actions" }, [el("button", { class: "btn btn-primary", type: "button", text: t("ok"), onclick: () => close(true) })])
      );
    });
    return;
  }
  if (n.code === 1105 && content.kind === "mute") {
    state.chat.mute = null;
    paintMute();
    toast(t("liftedMute"), "ok");
  }
}
