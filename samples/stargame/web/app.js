"use strict";

const API = `${location.protocol}//${location.hostname}:7450`;
const SERVER_KEY = "defaultkey";
const ROUND_SECONDS = 30;
const LS = {
  device: "stargame.device",
  skin: "stargame.skin",
  storeToken: "stargame.lastStoreToken",
};

const ERRORS = {
  UNAUTHENTICATED: "نشست تمام شده؛ دوباره وارد شو.",
  REQUEST_TOO_LARGE: "درخواست خیلی بزرگ است.",
  INVALID_JSON: "درخواست نامعتبر است.",
  INVALID_LANGUAGE: "زبان نامعتبر است.",
  INVALID_CURSOR: "صفحه‌بندی نامعتبر است؛ دوباره باز کن.",
  INVALID_METHOD: "روش پرداخت نامعتبر است.",
  INVALID_TOKEN: "توکن خرید خالی یا نامعتبر است.",
  INVALID_ID: "شناسه نامعتبر است.",
  INTERNAL_ERROR: "خطای سرور. کمی بعد دوباره امتحان کن.",
  TIMEOUT: "پاسخ سرور طول کشید. دوباره امتحان کن.",
  CANCELED: "درخواست لغو شد.",
  PRODUCT_NOT_FOUND: "این محصول پیدا نشد یا دیگر فروخته نمی‌شود.",
  PRIZE_NOT_FOUND: "جایزه‌ی این مورد پیدا نشد.",
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
  INSUFFICIENT_BALANCE: "سکه‌ات کافی نیست. چند دور بازی کن یا بسته‌ی سکه بخر.",
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
  COMMENT_RATE_LIMITED: "کمی صبر کن؛ در هر دقیقه فقط ۵ نظر می‌شود فرستاد.",
  COMMENT_MESSAGE_REQUIRED: "اول نظرت را بنویس.",
  COMMENT_TOO_LONG: "نظر حداکثر ۱۰۰۰ نویسه است.",
  PARENT_COMMENT_NOT_FOUND: "نظری که به آن جواب می‌دهی دیگر نیست.",
  QUERY_TOO_LONG: "متن جستجو خیلی بلند است.",
};

const SCREENS = ["home", "play", "shop", "gift", "payments", "news", "faq", "inbox"];
const SCREEN_LABEL = { home: "خانه", play: "بازی", shop: "فروشگاه", gift: "کد هدیه", payments: "پرداخت‌های من", news: "اخبار", faq: "سؤال‌های متداول", inbox: "صندوق پیام" };

const METHOD_LABEL = {
  zarinpal: "زرین‌پال",
  zibal: "زیبال",
  bazaar: "کافه‌بازار",
  myket: "مایکت",
  in_app: "پرداخت با سکه",
};

const STATUS_LABEL = {
  success: "موفق",
  pending: "در انتظار",
  failed: "ناموفق",
  expired: "منقضی",
};

const SKINS = {
  default: { name: "سبد ساده", body: ["#c98a52", "#8a5a30"], rim: "#e4b383", glow: null },
  skin_gold: { name: "سبد طلایی", body: ["#ffe27a", "#c98a00"], rim: "#fff4c2", glow: "rgba(255,201,60,0.55)" },
  skin_neon: { name: "سبد نئونی", body: ["#0b1f2a", "#0b1f2a"], rim: "#2ef2ff", glow: "rgba(46,242,255,0.8)", neon: true },
  skin_galaxy: { name: "سبد کهکشانی", body: ["#7b3fe4", "#1c1a6b"], rim: "#c6a8ff", glow: "rgba(150,110,255,0.6)", galaxy: true },
};

const nf = new Intl.NumberFormat("fa-IR");
const fa = (n) => nf.format(Number(n) || 0);
const faDigits = (s) => String(s).replace(/[0-9]/g, (d) => "۰۱۲۳۴۵۶۷۸۹"[d]);
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
  socket: null,
  socketRetry: 0,
};

class ApiError extends Error {
  constructor(code, status) {
    super(code);
    this.code = code;
    this.status = status;
  }
}

function errorText(err) {
  const code = err && err.code ? err.code : "INTERNAL_ERROR";
  return ERRORS[code] || `خطا: ${code}`;
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

async function login() {
  const body = await request("/v2/account/authenticate/device?create=true", {
    method: "POST",
    headers: { Authorization: "Basic " + btoa(SERVER_KEY + ":"), "Content-Type": "application/json" },
    body: JSON.stringify({ id: deviceId() }),
  });
  state.token = body.token;
  state.username = decodeJwt(body.token).usn || "";
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
      await login();
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
  if (name !== "play") stopGame();
  for (const s of document.querySelectorAll(".screen")) s.hidden = s.id !== "screen-" + name;
  window.scrollTo(0, 0);
  if (name === "home") {
    renderHome();
    if (state.token) refreshNewsDot();
  }
  if (name === "shop") loadShop();
  if (name === "payments") loadPayments(true);
  if (name === "news") loadNews(true);
  if (name === "faq") loadFaq();
  if (name === "inbox") {
    $("#note-toast").hidden = true;
    loadInbox(true);
  }
  if (name === "gift") {
    $("#gift-msg").textContent = "";
    $("#gift-code").focus();
  }
}

async function refreshBalances() {
  const b = await rpc("economy_balances", {});
  state.wallet = b.wallet || {};
  state.items = b.items || [];
  $("#balance").textContent = fa(state.wallet.coin || 0);
  const owned = selectedSkin();
  if (!ownsSkin(owned)) localStorage.setItem(LS.skin, "default");
  return b;
}

function ownsSkin(key) {
  return key === "default" || state.items.some((i) => i.key === key);
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
  $("#username").textContent = state.username || "بازیکن";
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
          toast(`${SKINS[key].name} انتخاب شد`, "ok");
        },
      }, [skinPreview(key), SKINS[key].name, el("span", { class: "tag", text: key === current ? "انتخاب‌شده" : "" })])
    );
  }
  const missing = Object.keys(SKINS).filter((k) => !ownsSkin(k)).length;
  if (missing) {
    box.append(
      el("button", { class: "skin", type: "button", onclick: () => { state.shopTab = "star_skins"; show("shop"); } }, [
        el("span", { class: "muted", text: "+" }),
        `${fa(missing)} سبد دیگر`,
        el("span", { class: "tag", text: "در فروشگاه" }),
      ])
    );
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
  $("#hud-score").textContent = fa(game.score);
  $("#hud-time").textContent = fa(Math.ceil(Math.max(0, game.time)));
}

async function startGame() {
  show("play");
  game.canvas = $("#canvas");
  game.ctx = game.canvas.getContext("2d");
  resizeCanvas();
  game.objects = [];
  game.floats = [];
  game.score = 0;
  game.time = ROUND_SECONDS;
  game.spawnIn = 0.4;
  game.basket.x = game.w / 2;
  game.basket.target = null;
  game.skin = selectedSkin();
  game.phase = "starting";
  game.roundId = null;
  updateHud();
  setOverlay(el("div", { class: "result" }, [el("div", { class: "spinner" }), el("p", { text: "آماده‌سازی دور…" })]));
  game.running = true;
  game.last = performance.now();
  cancelAnimationFrame(game.raf);
  game.raf = requestAnimationFrame(loop);
  try {
    const r = await rpc("stargame_start", {});
    if (!game.running) return;
    game.roundId = r.roundId;
  } catch (err) {
    if (!game.running) return;
    game.phase = "error";
    setOverlay(resultView("شروع نشد", errorText(err), null));
    return;
  }
  game.phase = "countdown";
  game.count = 3;
  setOverlay(el("div", { class: "countdown", text: fa(3) }));
}

function stopGame() {
  game.running = false;
  game.phase = "idle";
  cancelAnimationFrame(game.raf);
}

function spawn() {
  const r = Math.random();
  const kind = r < 0.1 ? "gold" : r < 0.3 ? "bomb" : "star";
  const size = kind === "gold" ? 17 : kind === "bomb" ? 15 : 13;
  const elapsed = ROUND_SECONDS - game.time;
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
      setOverlay(el("div", { class: "countdown", text: fa(next) }));
    }
    return;
  }
  if (game.phase !== "playing") return;

  game.time -= dt;
  game.spawnIn -= dt;
  const elapsed = ROUND_SECONDS - game.time;
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
      game.score += delta;
      game.floats.push({ x: o.x, y: by - 30, text: (delta > 0 ? "+" : "−") + fa(Math.abs(delta)), color: delta > 0 ? (delta > 1 ? "#ffc93c" : "#fff3b0") : "#ff6b6b", life: 0.9 });
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
  ctx.font = "700 20px Vazirmatn, Tahoma, sans-serif";
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
      el("button", { class: "btn btn-primary", type: "button", text: "دوباره بازی", onclick: startGame }),
      el("button", { class: "btn", type: "button", text: "خانه", onclick: () => show("home") }),
    ]),
  ]);
}

async function finishGame() {
  game.phase = "finishing";
  game.time = 0;
  updateHud();
  const score = game.score;
  setOverlay(el("div", { class: "result" }, [el("div", { class: "spinner" }), el("p", { text: "در حال ثبت امتیاز…" })]));
  try {
    const r = await rpc("stargame_finish", { roundId: game.roundId, score });
    state.wallet.coin = r.balance;
    $("#balance").textContent = fa(r.balance);
    game.phase = "done";
    const text = r.coins > 0 ? `${fa(r.coins)} سکه گرفتی. موجودی: ${fa(r.balance)} سکه` : `این دور سکه‌ای نگرفتی. موجودی: ${fa(r.balance)} سکه`;
    setOverlay(resultView("دور تمام شد!", text, el("div", { class: "big", text: fa(score) })));
    $("#stage-overlay").dataset.coins = String(r.coins);
  } catch (err) {
    game.phase = "done";
    setOverlay(resultView("امتیاز ثبت نشد", errorText(err), el("div", { class: "big", text: fa(score) })));
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
    if (r.type === "currency") return `${fa(r.value)} ${r.key === "coin" ? "سکه" : r.key}`;
    if (r.type === "storage") return (r.value && r.value.name) || (SKINS[r.key] && SKINS[r.key].name) || r.key;
    return `${r.key} ${fa(r.value)}`;
  }).join("، ");
}

function priceView(p) {
  if (p.coinPrice) return el("span", { class: "price" }, `${fa(p.coinPrice.amount)} سکه`);
  if (p.priceToman !== null && p.priceToman !== undefined) {
    if (p.finalToman !== p.priceToman) return el("span", { class: "price" }, [el("s", { text: fa(p.priceToman) }), `${fa(p.finalToman)} تومان`]);
    return el("span", { class: "price" }, `${fa(p.priceToman)} تومان`);
  }
  return el("span", { class: "price", text: "رایگان" });
}

async function loadShop() {
  const box = $("#products");
  box.textContent = "";
  box.append(el("div", { class: "center-block" }, [el("div", { class: "spinner" })]));
  try {
    const [catalog] = await Promise.all([rpc("economy_catalog", { lang: "fa" }), refreshBalances()]);
    state.catalog = catalog;
    renderShop();
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
    box.append(el("div", { class: "empty", text: "فعلاً چیزی برای فروش نیست." }));
    return;
  }
  for (const p of products) box.append(productCard(p));
}

function skinKeyOf(p) {
  const m = /^star_(skin_[a-z]+)$/.exec(p.key);
  return m && SKINS[m[1]] ? m[1] : null;
}

function productCard(p) {
  const skin = skinKeyOf(p);
  const owned = skin && ownsSkin(skin);
  const art = el("div", { class: "product-art" }, skin ? skinPreview(skin, 72, 36) : el("span", { class: "coin" }));
  const methods = el("div", { class: "methods" });
  if (owned) {
    const selected = selectedSkin() === skin;
    methods.append(el("span", { class: "owned", text: "مال توست" }));
    methods.append(el("button", { class: "btn", type: "button", disabled: selected, text: selected ? "انتخاب‌شده" : "انتخاب برای بازی", onclick: () => { localStorage.setItem(LS.skin, skin); toast(`${SKINS[skin].name} انتخاب شد`, "ok"); renderShop(); } }));
  } else {
    for (const m of p.paymentMethods) {
      methods.append(el("button", { class: m === "in_app" ? "btn btn-primary" : "btn", type: "button", "data-method": m, text: m === "in_app" ? "خرید با سکه" : METHOD_LABEL[m] || m, onclick: (e) => buy(p, m, e.currentTarget) }));
    }
  }
  return el("article", { class: "product", "data-key": p.key }, [
    el("div", { class: "product-top" }, [art, el("div", {}, [el("h3", { text: p.name }), p.description ? el("p", { class: "muted small", text: p.description }) : null])]),
    priceView(p),
    methods,
  ]);
}

async function buy(p, method, button) {
  if (button) button.disabled = true;
  try {
    if (method === "in_app") {
      const ok = await confirmDialog("خرید با سکه", `«${p.name}» را با ${fa(p.coinPrice.amount)} سکه می‌خری؟`, "بخر");
      if (!ok) return;
    }
    const payload = { productId: p.id, method };
    const code = $("#discount-code").value.trim().toUpperCase();
    if ((method === "zarinpal" || method === "zibal") && code) payload.giftCode = code;
    const r = await rpc("economy_purchase_start", payload);
    if (r.status === "redirect") {
      toast(`در حال رفتن به ${METHOD_LABEL[method]}…`);
      location.href = r.redirectUrl;
      return;
    }
    if (r.status === "store") {
      await storeSimulator(p, method, r);
      return;
    }
    await refreshBalances();
    toast(`خرید موفق! ${rewardText(r.rewards)}`, "ok");
    if (skinKeyOf(p)) localStorage.setItem(LS.skin, skinKeyOf(p));
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
        el("button", { class: "btn", type: "button", text: "انصراف", onclick: () => close(false) }),
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
    const pay = el("button", { class: "btn btn-ok", type: "button", id: "sim-pay", text: "پرداخت موفق" });
    pay.addEventListener("click", async () => {
      pay.disabled = true;
      msg.className = "form-msg";
      msg.textContent = "در حال تأیید خرید…";
      const used = input.value.trim();
      try {
        const r = await rpc("economy_purchase_submit", { paymentId: start.paymentId, token: used });
        localStorage.setItem(LS.storeToken, JSON.stringify({ method, sku: start.sku, token: used }));
        await refreshBalances();
        close(true);
        if (r.status === "success" && r.rewardGranted === false) toast(errorText({ code: r.error || "REWARD_PENDING" }), "bad");
        else toast(`خرید موفق! ${rewardText(r.rewards)}`, "ok");
        if (!$("#screen-shop").hidden) renderShop();
      } catch (err) {
        msg.className = "form-msg bad";
        msg.textContent = errorText(err);
        pay.disabled = false;
      }
    });
    body.append(
      el("span", { class: "sim-badge", text: "شبیه‌ساز فروشگاه (حالت تست)" }),
      el("div", { class: "sim-store" }, [el("div", { class: "sim-logo " + method, text: method === "bazaar" ? "ب" : "م" }), el("div", {}, [el("b", { text: METHOD_LABEL[method] }), el("div", { class: "muted small", text: "روی وب برنامه‌ی فروشگاه نداریم؛ این پنجره جای آن را می‌گیرد." })])]),
      el("dl", { class: "kv" }, [
        el("dt", { text: "محصول" }), el("dd", { text: p.name }),
        el("dt", { text: "مبلغ" }), el("dd", { text: p.finalToman !== null && p.finalToman !== undefined ? `${fa(p.finalToman)} تومان` : "—" }),
        el("dt", { text: "SKU" }), el("dd", { class: "mono", dir: "ltr", text: start.sku }),
        el("dt", { text: "پرداخت" }), el("dd", { class: "mono", dir: "ltr", text: start.paymentId }),
      ]),
      el("div", { class: "field" }, [el("label", { for: "sim-token", text: "توکن خرید آزمایشی (برای تست می‌توانی یک توکن قدیمی را جایگزین کنی)" }), input]),
      msg,
      el("div", { class: "dialog-actions" }, [pay, el("button", { class: "btn", type: "button", text: "انصراف", onclick: () => close(false) })])
    );
  });
}

function restoreDialog() {
  const last = JSON.parse(localStorage.getItem(LS.storeToken) || "null");
  const skus = state.catalog ? state.catalog.products.filter((p) => p.sku && (p.sku.bazaar || p.sku.myket)) : [];
  return openDialog((body, close) => {
    const method = el("select", { id: "restore-method" }, [el("option", { value: "bazaar", text: "کافه‌بازار" }), el("option", { value: "myket", text: "مایکت" })]);
    const sku = el("select", { id: "restore-sku", dir: "ltr" });
    const fillSkus = () => {
      sku.textContent = "";
      for (const p of skus) {
        const s = p.sku[method.value];
        if (s) sku.append(el("option", { value: s, text: `${s} — ${p.name}` }));
      }
      if (last && last.method === method.value) sku.value = last.sku;
    };
    method.addEventListener("change", fillSkus);
    if (last) method.value = last.method;
    fillSkus();
    const token = el("input", { type: "text", id: "restore-token", class: "mono", dir: "ltr", spellcheck: "false", value: last ? last.token : "" });
    const msg = el("p", { class: "form-msg", role: "status" });
    const go = el("button", { class: "btn btn-primary", type: "button", text: "بررسی و بازیابی" });
    go.addEventListener("click", async () => {
      go.disabled = true;
      msg.className = "form-msg";
      msg.textContent = "در حال بررسی…";
      try {
        const r = await rpc("economy_purchase_review", { method: method.value, sku: sku.value, token: token.value.trim() });
        await refreshBalances();
        msg.className = "form-msg ok";
        msg.textContent = r.status === "success" ? `خرید تأیید شد${r.rewards ? ": " + rewardText(r.rewards) : ""}. اگر قبلاً جایزه را گرفته بودی، دوباره داده نمی‌شود.` : `وضعیت: ${STATUS_LABEL[r.status] || r.status}`;
      } catch (err) {
        msg.className = "form-msg bad";
        msg.textContent = errorText(err);
      } finally {
        go.disabled = false;
      }
    });
    body.append(
      el("h2", { text: "بازیابی خرید فروشگاه" }),
      el("p", { class: "muted small", text: "اگر پولت از بازار یا مایکت کم شد ولی سکه نرسید، توکن خرید را اینجا بفرست." }),
      el("div", { class: "field" }, [el("label", { for: "restore-method", text: "فروشگاه" }), method]),
      el("div", { class: "field" }, [el("label", { for: "restore-sku", text: "محصول" }), sku]),
      el("div", { class: "field" }, [el("label", { for: "restore-token", text: "توکن خرید" }), token]),
      msg,
      el("div", { class: "dialog-actions" }, [go, el("button", { class: "btn", type: "button", text: "بستن", onclick: () => close(false) })])
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
  msg.textContent = "در حال بررسی…";
  try {
    const r = await rpc("economy_gift_redeem", { code });
    await refreshBalances();
    msg.className = "form-msg ok";
    msg.textContent = `هدیه گرفتی: ${rewardText(r.rewards)}`;
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
    return new Date(iso).toLocaleString("fa-IR", { dateStyle: "medium", timeStyle: "short" });
  } catch (e) {
    return faDigits(iso);
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
    const r = await rpc("economy_payments_mine", { limit: 20, cursor: state.paymentsCursor || undefined });
    if (reset) box.textContent = "";
    if (reset && !r.items.length) box.append(el("div", { class: "empty", text: "هنوز پرداختی نداری." }));
    for (const p of r.items) {
      const name = p.productName ? p.productName.fa || p.productName.en || p.productKey : p.productKey;
      const amount = p.method === "in_app" ? `${fa(p.coinAmount)} سکه` : `${fa(p.finalToman)} تومان`;
      box.append(el("div", { class: "pay", "data-status": p.status, "data-method": p.method }, [
        el("span", { class: "name", text: name }),
        el("span", { class: "amount", text: amount }),
        el("span", { class: "meta", text: `${METHOD_LABEL[p.method] || p.method}، ${formatDate(p.createdAt)}${p.refId ? "، کد پیگیری " + faDigits(p.refId) : ""}` }),
        el("span", { class: "status status-" + p.status, text: STATUS_LABEL[p.status] || p.status }),
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
  $("#loading-text").textContent = "در حال بررسی پرداخت…";
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
      el("h2", { text: ok ? "پرداخت موفق بود" : status && status.status === "pending" ? "پرداخت هنوز در انتظار است" : "پرداخت انجام نشد" }),
      el("p", { text: ok ? (status.rewardGranted ? `سکه‌ها به حسابت اضافه شد. موجودی: ${fa(state.wallet.coin || 0)} سکه` : errorText({ code: "REWARD_PENDING" })) : status && status.error ? errorText({ code: status.error }) : status && status.status === "pending" ? "چند دقیقه‌ی دیگر در «پرداخت‌های من» نگاه کن." : "پولی کم نشده باشد، برمی‌گردد." }),
      el("div", { class: "dialog-actions" }, [el("button", { class: "btn btn-primary", type: "button", text: "باشه", onclick: () => close(true) })])
    );
  });
  return true;
}

function navigate(target) {
  const name = String(target || "").trim();
  if (name === "play") return startGame();
  if (SCREENS.includes(name)) return show(name);
  toast("این بخش در بازی پیدا نشد.", "bad");
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
  const dot = $("#news-dot");
  dot.hidden = !n;
  dot.textContent = n ? fa(n) : "";
  $("#menu-news").setAttribute("aria-label", n ? `اخبار، ${fa(n)} خبر تازه` : "اخبار");
}

async function refreshNewsDot() {
  try {
    const r = await rpc("publication_news_list", { lang: "fa", limit: 50 });
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
    const r = await rpc("publication_news_list", { lang: "fa", limit: 10, cursor: state.newsCursor || undefined });
    if (reset) box.textContent = "";
    if (reset && !r.items.length) box.append(el("div", { class: "empty", text: "هنوز خبری منتشر نشده." }));
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
        n.isPinned ? el("span", { class: "chip chip-gold", text: "سنجاق‌شده" }) : null,
        n.seen ? null : el("span", { class: "chip chip-new" }, [el("span", { class: "dot", "aria-hidden": "true" }), "تازه"]),
      ]),
      el("h3", { text: n.title }),
      n.summary ? el("p", { class: "muted small", text: n.summary }) : null,
      el("div", { class: "news-meta muted small" }, [el("span", { text: formatDate(n.publishedAt) }), el("span", { text: `${fa(n.commentCount)} نظر` })]),
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
    const a = await rpc("publication_news_get", { id, lang: "fa" });
    state.article = a;
    box.textContent = "";
    if (a.coverUrl) box.append(el("img", { class: "article-cover", src: a.coverUrl, alt: "" }));
    box.append(
      el("div", { class: "news-tags" }, [a.isPinned ? el("span", { class: "chip chip-gold", text: "سنجاق‌شده" }) : null]),
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
    $("#comments-count").textContent = a.allowComment ? "" : "نظر دادن برای این خبر بسته است";
    loadComments(true);
  } catch (err) {
    box.textContent = "";
    box.append(el("div", { class: "empty", text: errorText(err) }));
  }
}

function commentView(c, root) {
  const node = el("div", { class: "comment" + (c.isAccepted ? "" : " pending") + (c.mine ? " mine" : ""), "data-comment": c.id }, [
    el("div", { class: "comment-head" }, [
      el("bdi", { class: "comment-user", text: c.displayName || c.username || "بازیکن" }),
      c.mine ? el("span", { class: "chip", text: "تو" }) : null,
      el("span", { class: "muted small", text: formatDate(c.createdAt) }),
      c.isAccepted ? null : el("span", { class: "chip chip-wait", text: "در انتظار تأیید" }),
    ]),
    el("p", { class: "comment-text", dir: "auto", text: c.message }),
  ]);
  if (root && c.isAccepted && state.article && state.article.allowComment) {
    node.append(el("button", { class: "link-btn", type: "button", text: "پاسخ", onclick: () => startReply(c) }));
  }
  if (root && c.replies && c.replies.length) {
    node.append(el("div", { class: "replies" }, c.replies.map((r) => commentView(r, false))));
  }
  return node;
}

function startReply(c) {
  state.replyTo = c;
  $("#comment-hint").textContent = "";
  $("#comment-hint").append(`در پاسخ به ${c.displayName || c.username || "بازیکن"} `, el("button", { class: "link-btn", type: "button", text: "لغو", onclick: () => { state.replyTo = null; $("#comment-hint").textContent = ""; } }));
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
    if (reset && !r.items.length) list.append(el("p", { class: "muted small center", id: "no-comments", text: state.article.allowComment ? "هنوز نظری نیست. اولین نفر باش!" : "نظری نیست." }));
    for (const c of r.items) list.append(commentView(c, true));
    state.commentsCursor = r.cursor || "";
    more.hidden = !r.cursor;
    if (state.article.allowComment) $("#comments-count").textContent = `${fa(state.article.commentCount)} نظر`;
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
    toast(c.isAccepted ? "نظرت منتشر شد." : "نظرت ثبت شد و بعد از تأیید برای بقیه نمایش داده می‌شود.", "ok");
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
    const r = await rpc("publication_faq_list", { lang: "fa", q: q || undefined, categoryKey: state.faqCategory || undefined });
    if (my !== faqSeq) return;
    const tabs = $("#faq-tabs");
    tabs.textContent = "";
    if (r.categories.length) {
      const all = [{ key: "", name: "همه" }].concat(r.categories);
      for (const c of all) {
        tabs.append(el("button", { class: "tab", type: "button", role: "tab", "aria-selected": c.key === state.faqCategory ? "true" : "false", "data-cat": c.key, onclick: () => { state.faqCategory = c.key; loadFaq(); } }, c.name));
      }
    }
    box.textContent = "";
    if (!r.items.length) {
      box.append(el("div", { class: "empty", text: q ? "با این جستجو سؤالی پیدا نشد." : "هنوز سؤالی اضافه نشده." }));
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
  b.textContent = state.unread > 99 ? "+۹۹" : fa(state.unread);
  $("#btn-inbox").setAttribute("aria-label", state.unread ? `صندوق پیام، ${fa(state.unread)} پیام نخوانده` : "صندوق پیام");
}

async function refreshUnread() {
  try {
    const r = await rpc("publication_inbox", { lang: "fa", limit: 1 });
    setUnread(r.unread);
  } catch (e) {
    return;
  }
}

function inboxItem(m) {
  const node = el("div", { class: "inbox-item" + (m.seen ? "" : " unread"), "data-inbox": m.id }, [
    el("div", { class: "inbox-head" }, [
      m.seen ? null : el("span", { class: "dot", "aria-label": "نخوانده" }),
      el("h3", { text: m.title || "پیام" }),
      el("span", { class: "muted small", text: formatDate(m.createdAt) }),
    ]),
    richText(m.bodyHtml),
  ]);
  if (m.action && m.action.value) {
    node.append(el("button", { class: "btn", type: "button", "data-action": m.action.type, text: m.action.type === "url" ? "باز کردن پیوند ↗" : `رفتن به ${SCREEN_LABEL[m.action.value] || m.action.value}`, onclick: () => {
      if (!m.seen) rpc("publication_inbox_seen", { ids: [m.id] }).then((r) => setUnread(r.unread)).catch(() => {});
      runAction(m.action);
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
    const r = await rpc("publication_inbox", { lang: "fa", limit: 20, cursor: state.inboxCursor || undefined });
    if (reset) box.textContent = "";
    if (reset && !r.items.length) box.append(el("div", { class: "empty", text: "صندوقت خالی است." }));
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
    toast("همه‌ی پیام‌ها خوانده شد.", "ok");
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
    el("span", { class: "note-text" }, [el("b", { text: n.title || "پیام تازه" }), n.body ? el("span", { class: "muted small", text: plainText(n.body) }) : null])
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
      const r = await rpc("publication_inbox", { lang: "fa", limit: 10 });
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
  const proto = location.protocol === "https:" ? "wss:" : "ws:";
  const ws = new WebSocket(`${proto}//${location.hostname}:7450/ws?token=${encodeURIComponent(state.token)}&format=json`);
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
    const list = msg && msg.notifications && msg.notifications.notifications;
    if (!list) return;
    for (const n of list) {
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
    if (state.socket !== ws) return;
    state.socket = null;
    const wait = Math.min(30000, 1000 * 2 ** state.socketRetry++);
    setTimeout(() => { if (!state.socket) connectSocket(); }, wait);
  };
}

async function boot() {
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
  try {
    await login();
    await refreshBalances();
  } catch (err) {
    $("#loading-text").textContent = errorText(err);
    document.querySelector("#screen-loading .spinner").hidden = true;
    return;
  }
  refreshUnread();
  refreshNewsDot();
  if (!(await handleReturn())) show("home");
}

boot();
