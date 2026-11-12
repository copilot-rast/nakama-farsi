Object.assign(STRINGS.fa, {
  chance: "شانس",
  chanceHint: "گردونه را بچرخان یا جعبهٔ شانس را باز کن. سرور برنده را انتخاب می‌کند و جایزه همان لحظه به تو می‌رسد.",
  chanceFree: "بچرخان، رایگان",
  chanceTicket: (n, name) => `با بلیت · ${num(n)} ${name}`,
  chanceWait: (s) => `نوبت رایگان بعدی: ${s}`,
  chanceFreeReady: "نوبت رایگانت آماده است",
  chanceNoFree: "این بازی نوبت رایگان ندارد",
  chanceEvery: (s) => `رایگان، هر ${s} یک بار`,
  chanceSpinning: "در حال چرخیدن…",
  chanceWon: (name) => `بردی: ${name}`,
  chanceNothing: "این بار پوچ شد. دفعهٔ بعد!",
  chanceEmpty: "پوچ",
  chanceShort: (n, name) => `برای بلیت ${num(n)} ${name} لازم است.`,
  chanceHistory: "بازی‌های اخیر من",
  chanceHistoryEmpty: "هنوز بازی نکرده‌ای.",
  chanceNone: "فعلاً بازی شانسی نیست.",
  chancePlays: (n) => `${num(n)} بار بازی کرده‌ای`,
  chanceModeFree: "رایگان",
  chanceModeTicket: "بلیت",
  chanceFromGroup: (name) => `از ${name}`,
  chanceOdds: (n) => `${num(n)}٪`,
  chanceDays: (n) => `${num(n)} روز`,
  chanceHours: (n) => `${num(n)} ساعت`,
  chanceMinutes: (n) => `${num(n)} دقیقه`,
  chanceOddsTitle: "شانس هر خانه",
});

Object.assign(STRINGS.en, {
  chance: "Luck",
  chanceHint: "Spin the wheel or open the lucky box. The server picks the winner and the prize reaches you at once.",
  chanceFree: "Spin free",
  chanceTicket: (n, name) => `Ticket · ${num(n)} ${name}`,
  chanceWait: (s) => `Next free turn: ${s}`,
  chanceFreeReady: "Your free turn is ready",
  chanceNoFree: "This game has no free turn",
  chanceEvery: (s) => `Free once every ${s}`,
  chanceSpinning: "Spinning…",
  chanceWon: (name) => `You won: ${name}`,
  chanceNothing: "Nothing this time. Next time!",
  chanceEmpty: "Nothing",
  chanceShort: (n, name) => `A ticket needs ${num(n)} ${name}.`,
  chanceHistory: "My recent plays",
  chanceHistoryEmpty: "You have not played yet.",
  chanceNone: "No chance games right now.",
  chancePlays: (n) => `You played ${num(n)} times`,
  chanceModeFree: "Free",
  chanceModeTicket: "Ticket",
  chanceFromGroup: (name) => `from ${name}`,
  chanceOdds: (n) => `${num(n)}%`,
  chanceDays: (n) => `${num(n)} d`,
  chanceHours: (n) => `${num(n)} h`,
  chanceMinutes: (n) => `${num(n)} min`,
  chanceOddsTitle: "Odds of each slot",
});

Object.assign(STRINGS.fa.errors, {
  WAITING_FOR_TURN: "نوبت بعدی هنوز نرسیده.",
  CURRENCY_IS_NOT_ENOUGH: "سکه‌ات برای بلیت کافی نیست.",
  TICKET_PLAY_IS_NOT_SUPPORT: "این بازی با بلیت بازی نمی‌شود.",
  INTERVAL_TIME_IS_NULL: "این بازی نوبت رایگان ندارد.",
  CHANCE_GAME_NOT_FOUND: "این بازی شانس دیگر نیست.",
});

Object.assign(STRINGS.en.errors, {
  WAITING_FOR_TURN: "Your next turn is not here yet.",
  CURRENCY_IS_NOT_ENOUGH: "Not enough coins for a ticket.",
  TICKET_PLAY_IS_NOT_SUPPORT: "This game cannot be played with a ticket.",
  INTERVAL_TIME_IS_NULL: "This game has no free turn.",
  CHANCE_GAME_NOT_FOUND: "This chance game is gone.",
});

if (!SCREENS.includes("chance")) SCREENS.push("chance");

const CH_COLORS = ["#ffc93c", "#7c8cff", "#3ddc97", "#ff8a5c", "#c6a8ff", "#2ef2ff", "#ff6b9a", "#9be15d"];
const CH_EMPTY = "#2f3878";
const chanceState = { games: [], rotation: {}, timer: null, busy: false, seq: 0 };

function chanceReduced() {
  return window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

function chanceClock(sec) {
  sec = Math.max(0, Math.ceil(Number(sec) || 0));
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = sec % 60;
  const two = (v) => num(v).padStart(2, num(0));
  return (h ? num(h) + ":" : "") + two(m) + ":" + two(s);
}

function chanceSpan(min) {
  min = Number(min) || 0;
  if (min >= 1440 && min % 1440 === 0) return t("chanceDays", min / 1440);
  if (min >= 60 && min % 60 === 0) return t("chanceHours", min / 60);
  return t("chanceMinutes", min);
}

function chanceLabel(row) {
  if (row.prize) return row.prize.name || row.prize.key;
  if (row.group) return row.group.name || row.group.key;
  return t("chanceEmpty");
}

function chancePoint(r, deg) {
  const a = (deg * Math.PI) / 180;
  return [r * Math.sin(a), -r * Math.cos(a)];
}

function chanceWheel(game) {
  const ns = "http://www.w3.org/2000/svg";
  const svg = document.createElementNS(ns, "svg");
  svg.setAttribute("viewBox", "-100 -100 200 200");
  svg.setAttribute("class", "ch-wheel-svg");
  svg.setAttribute("aria-hidden", "true");
  let start = 0;
  let colour = 0;
  (game.prize || []).forEach((row, i) => {
    const size = Number(row.chance) * 3.6;
    if (size <= 0) return;
    const end = start + size;
    const fill = row.prizeId || row.prizeGroupId ? CH_COLORS[colour++ % CH_COLORS.length] : CH_EMPTY;
    let shape;
    if (size >= 359.999) {
      shape = document.createElementNS(ns, "circle");
      shape.setAttribute("r", "96");
    } else {
      const [x1, y1] = chancePoint(96, start);
      const [x2, y2] = chancePoint(96, end);
      shape = document.createElementNS(ns, "path");
      shape.setAttribute("d", `M0,0 L${x1.toFixed(2)},${y1.toFixed(2)} A96,96 0 ${size > 180 ? 1 : 0} 1 ${x2.toFixed(2)},${y2.toFixed(2)} Z`);
    }
    shape.setAttribute("fill", fill);
    shape.setAttribute("stroke", "#0b1026");
    shape.setAttribute("stroke-width", "1.5");
    shape.setAttribute("data-slice", String(i));
    svg.append(shape);
    const mid = start + size / 2;
    const label = document.createElementNS(ns, "text");
    const [lx, ly] = chancePoint(60, mid);
    label.setAttribute("x", lx.toFixed(2));
    label.setAttribute("y", ly.toFixed(2));
    label.setAttribute("text-anchor", "middle");
    label.setAttribute("dominant-baseline", "middle");
    label.setAttribute("transform", `rotate(${(mid > 90 && mid < 270 ? mid + 180 : mid).toFixed(2)} ${lx.toFixed(2)} ${ly.toFixed(2)})`);
    label.setAttribute("class", "ch-label" + (fill === CH_EMPTY ? " ch-label-empty" : ""));
    label.setAttribute("direction", i18n.dir);
    let text = chanceLabel(row);
    if (size < 40 && text.length > 8) text = text.slice(0, 7) + "…";
    label.textContent = text;
    svg.append(label);
    start = end;
  });
  const hub = document.createElementNS(ns, "circle");
  hub.setAttribute("r", "14");
  hub.setAttribute("class", "ch-hub");
  svg.append(hub);
  return svg;
}

function chanceOdds(game) {
  return el("ul", { class: "ch-odds", "data-chance-odds": "" }, (game.prize || []).filter((r) => Number(r.chance) > 0).map((row) =>
    el("li", {}, [
      el("span", { class: "ch-odds-name", text: chanceLabel(row) }),
      row.group && row.group.prizes ? el("span", { class: "muted small", text: row.group.prizes.map((p) => p.name).join(i18n.dir === "rtl" ? "، " : ", ") }) : null,
      el("b", { class: "ch-odds-pct", text: t("chanceOdds", row.chance) }),
    ])));
}

function chanceCard(game) {
  const rot = chanceState.rotation[game.id] || 0;
  const wheel = el("div", { class: "ch-wheel", "data-chance-wheel": "", "data-rotation": String(rot), style: `transform: rotate(${rot}deg)` }, [chanceWheel(game)]);
  const stage = el("div", { class: "ch-stage" }, [el("div", { class: "ch-pointer", "aria-hidden": "true" }), wheel]);
  const result = el("p", { class: "ch-result", "data-chance-result": "", role: "status", "aria-live": "polite", hidden: true });
  const buttons = el("div", { class: "ch-actions" });
  const status = el("p", { class: "muted small", "data-chance-status": "" });
  if (game.intervalTime) {
    const free = el("button", { class: "btn btn-primary", type: "button", "data-chance-free": game.key, text: t("chanceFree") });
    free.disabled = !game.canPlay;
    free.addEventListener("click", () => playChance(game, "free", card));
    buttons.append(free);
    if (game.canPlay) status.textContent = t("chanceFreeReady");
    else status.append(el("span", { "data-chance-wait": String(game.nextPlayTime || 0), text: t("chanceWait", chanceClock(game.nextPlayTime)) }));
  } else {
    status.textContent = t("chanceNoFree");
  }
  if (game.price) {
    const cur = game.price.currencyName || game.price.currencyKey;
    const tb = el("button", { class: "btn", type: "button", "data-chance-ticket": game.key, text: t("chanceTicket", game.price.amount, cur) });
    const have = Number((state.wallet || {})[game.price.currencyKey] || 0);
    if (have < game.price.amount) {
      tb.disabled = true;
      tb.setAttribute("aria-describedby", "ch-short-" + game.key);
    }
    tb.addEventListener("click", () => playChance(game, "ticket", card));
    buttons.append(tb);
    if (have < game.price.amount) buttons.append(el("span", { class: "muted small", id: "ch-short-" + game.key, "data-chance-short": "", text: t("chanceShort", game.price.amount, cur) }));
  }
  const meta = [];
  if (game.intervalTime) meta.push(el("span", { class: "chip chip-new", text: t("chanceEvery", chanceSpan(game.intervalTime)) }));
  if (game.plays) meta.push(el("span", { class: "chip", "data-chance-plays": "", text: t("chancePlays", game.plays) }));
  const card = el("article", { class: "reward ch-card", "data-chance-game": game.key, "data-can-play": game.canPlay ? "true" : "false" }, [
    el("div", { class: "reward-head" }, [el("b", { text: game.name }), el("div", { class: "inv-chips" }, meta)]),
    game.description ? el("p", { class: "muted small", text: game.description }) : null,
    stage,
    result,
    buttons,
    status,
    el("details", { class: "ch-details" }, [el("summary", { class: "small", text: t("chanceOddsTitle") }), chanceOdds(game)]),
  ]);
  return card;
}

function chanceTick() {
  clearInterval(chanceState.timer);
  chanceState.timer = setInterval(() => {
    if (currentScreen() !== "chance") { clearInterval(chanceState.timer); return; }
    let reload = false;
    for (const w of document.querySelectorAll("[data-chance-wait]")) {
      const left = Number(w.dataset.chanceWait) - 1;
      w.dataset.chanceWait = String(left);
      w.textContent = t("chanceWait", chanceClock(left));
      if (left <= 0) reload = true;
    }
    if (reload && !chanceState.busy) loadChance();
  }, 1000);
}

async function playChance(game, mode, card) {
  if (chanceState.busy) return;
  chanceState.busy = true;
  for (const b of card.querySelectorAll(".ch-actions button")) b.disabled = true;
  const result = card.querySelector("[data-chance-result]");
  result.hidden = false;
  result.className = "ch-result";
  result.textContent = t("chanceSpinning");
  try {
    const res = await rpc(mode === "free" ? "economy_chance_play" : "economy_chance_play_with_ticket", { id: game.id, lang: i18n.lang });
    const wheel = card.querySelector("[data-chance-wheel]");
    const prev = chanceState.rotation[game.id] || 0;
    const target = (360 - (Number(res.roll) * 3.6) % 360) % 360;
    const next = prev - (prev % 360) + 360 * 5 + target;
    chanceState.rotation[game.id] = next;
    wheel.dataset.rotation = String(next);
    wheel.dataset.slot = res.slot == null ? "" : String(res.slot);
    const wait = chanceReduced() ? 0 : 3200;
    wheel.style.transition = wait ? `transform ${wait}ms cubic-bezier(.17,.67,.21,1)` : "none";
    requestAnimationFrame(() => { wheel.style.transform = `rotate(${next}deg)`; });
    await new Promise((r) => setTimeout(r, wait + 60));
    card.dataset.lastPrize = res.prizeId || "";
    result.dataset.prizeId = res.prizeId || "";
    result.dataset.playId = res.playId;
    if (res.prizeId) {
      const name = rewardText(res.rewards) || (res.prize && res.prize.name) || "";
      result.className = "ch-result ch-won";
      result.textContent = t("chanceWon", name) + (res.group ? " " + t("chanceFromGroup", res.group.name) : "");
    } else {
      result.className = "ch-result ch-lost";
      result.textContent = t("chanceNothing");
    }
    if (res.wallet) state.wallet = res.wallet;
    await refreshBalances().catch(() => {});
  } catch (err) {
    result.className = "ch-result ch-lost";
    result.textContent = errorText(err);
    toast(errorText(err), "bad");
  }
  chanceState.busy = false;
  const keep = { key: game.key, text: result.textContent, cls: result.className, prize: result.dataset.prizeId || "", play: result.dataset.playId || "" };
  await loadChance(keep);
}

async function loadChanceHistory() {
  const box = $("#chance-history");
  try {
    const res = await rpc("economy_chance_history", { limit: 5, lang: i18n.lang });
    box.textContent = "";
    const items = res.items || [];
    if (!items.length) {
      box.append(el("p", { class: "muted small", text: t("chanceHistoryEmpty") }));
      return;
    }
    box.append(el("ul", { class: "ch-history" }, items.map((x) => el("li", { "data-chance-play": x.playId, "data-won": x.prizeId ? "true" : "false" }, [
      el("span", { text: x.gameName }),
      el("span", { class: "chip " + (x.mode === "ticket" ? "chip-new" : "chip-gold"), text: x.mode === "ticket" ? t("chanceModeTicket") : t("chanceModeFree") }),
      el("b", { class: x.prizeId ? "reward-prize" : "muted", text: x.prizeId ? (rewardText(x.rewards) || (x.prize && x.prize.name) || "") : t("chanceEmpty") }),
      el("span", { class: "muted small", text: fmtDay(x.createdAt) }),
    ]))));
  } catch (err) {
    box.textContent = "";
  }
}

async function refreshChanceDot() {
  if (!state.token) return;
  try {
    const res = await rpc("economy_chance_games", { lang: i18n.lang });
    const ready = (res.items || []).filter((g) => g.canPlay).length;
    const dot = $("#chance-dot");
    dot.hidden = !ready;
    dot.textContent = ready ? num(ready) : "";
  } catch (e) {}
}

async function loadChance(keep) {
  const box = $("#chance-list");
  const my = ++chanceState.seq;
  if (!box.childNodes.length) box.append(el("div", { class: "center-block" }, [el("div", { class: "spinner" })]));
  try {
    const res = await rpc("economy_chance_games", { lang: i18n.lang });
    if (my !== chanceState.seq) return;
    chanceState.games = res.items || [];
    box.textContent = "";
    const ready = chanceState.games.filter((g) => g.canPlay).length;
    const dot = $("#chance-dot");
    dot.hidden = !ready;
    dot.textContent = ready ? num(ready) : "";
    if (!chanceState.games.length) {
      box.append(el("p", { class: "muted center", text: t("chanceNone") }));
    } else {
      box.append(el("div", { class: "ch-grid" }, chanceState.games.map((g) => {
        const card = chanceCard(g);
        if (keep && keep.key === g.key) {
          const r = card.querySelector("[data-chance-result]");
          r.hidden = false;
          r.className = keep.cls;
          r.textContent = keep.text;
          r.dataset.prizeId = keep.prize;
          r.dataset.playId = keep.play;
        }
        return card;
      })));
    }
    chanceTick();
  } catch (err) {
    if (my !== chanceState.seq) return;
    box.textContent = "";
    box.append(el("p", { class: "muted center", text: errorText(err) }));
  }
  loadChanceHistory();
}

function chanceShow(name) {
  if (name === "home") refreshChanceDot();
  if (name === "chance") {
    $("#chance-list").textContent = "";
    loadChance();
  }
}
