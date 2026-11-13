Object.assign(STRINGS.fa, {
  games: "بازی‌ها",
  gamesHint: "یک بازی و یک اتاق انتخاب کن. ورودی هر اتاق از کیف پولت کم می‌شود و برنده جایزه را می‌برد.",
  gamesNone: "فعلاً بازی‌ای نیست.",
  gamePlayers: (a, b) => (a === b ? `${num(a)} نفره` : `${num(a)} تا ${num(b)} نفر`),
  gameSolo: "تک‌نفره",
  gamePlay: "بازی کن",
  gameRoundTime: (n) => `هر دور ${num(n)} ثانیه`,
  gameTurn: (n) => `هر نوبت ${num(n)} ثانیه`,
  gameMaintenance: "در حال تعمیر",
  gameFree: "رایگان",
  gameFee: (n, name) => `ورودی ${num(n)} ${name}`,
  gamePrize: (name) => `جایزه: ${name}`,
  gameMinLevel: (n) => `از سطح ${num(n)}`,
  gameEnter: "ورود",
  gameSearching: "دنبال حریف می‌گردیم…",
  gameCancel: "لغو",
  gameFound: "حریف پیدا شد!",
  tttTitle: "دوز",
  tttYou: (m) => `تو: ${m}`,
  tttYourTurn: "نوبت توست",
  tttTheirTurn: "نوبت حریف است",
  tttLeft: (n) => `${num(n)} ثانیه`,
  tttWon: "بردی!",
  tttLost: "باختی.",
  tttDraw: "مساوی شد.",
  tttCancelled: "حریف نیامد؛ ورودی‌ات برگشت.",
  tttReason: { line: "سه خانه در یک خط", timeout: "وقت نوبت تمام شد", left: "حریف بازی را ترک کرد", draw: "همهٔ خانه‌ها پر شد", time_limit: "زمان بازی تمام شد", server_stop: "سرور بازی را بست" },
  tttPrize: (name) => `جایزه‌ات رسید: ${name}`,
  tttBack: "برگشت به بازی‌ها",
  tttCell: (n) => `خانهٔ ${num(n)}`,
});

Object.assign(STRINGS.en, {
  games: "Games",
  gamesHint: "Pick a game and a room. The room's entry comes out of your wallet and the winner takes the prize.",
  gamesNone: "No games right now.",
  gamePlayers: (a, b) => (a === b ? `${num(a)} players` : `${num(a)} to ${num(b)} players`),
  gameSolo: "Solo",
  gamePlay: "Play",
  gameRoundTime: (n) => `${num(n)} s a round`,
  gameTurn: (n) => `${num(n)} s a turn`,
  gameMaintenance: "Under maintenance",
  gameFree: "Free",
  gameFee: (n, name) => `Entry ${num(n)} ${name}`,
  gamePrize: (name) => `Prize: ${name}`,
  gameMinLevel: (n) => `From level ${num(n)}`,
  gameEnter: "Enter",
  gameSearching: "Looking for an opponent…",
  gameCancel: "Cancel",
  gameFound: "Opponent found!",
  tttTitle: "Tic-tac-toe",
  tttYou: (m) => `You: ${m}`,
  tttYourTurn: "Your turn",
  tttTheirTurn: "Opponent's turn",
  tttLeft: (n) => `${num(n)} s`,
  tttWon: "You won!",
  tttLost: "You lost.",
  tttDraw: "It's a draw.",
  tttCancelled: "No opponent came; your entry is back.",
  tttReason: { line: "Three in a row", timeout: "The turn timed out", left: "The opponent left", draw: "The board is full", time_limit: "The match ran out of time", server_stop: "The server closed the match" },
  tttPrize: (name) => `Your prize arrived: ${name}`,
  tttBack: "Back to games",
  tttCell: (n) => `Cell ${num(n)}`,
});

Object.assign(STRINGS.fa.errors, {
  GAME_DISABLED: "این بازی فعلاً خاموش است.",
  GAME_NOT_FOUND: "این بازی پیدا نشد.",
  ROOM_NOT_FOUND: "این اتاق پیدا نشد.",
  ROOM_REQUIRED: "یک اتاق انتخاب کن.",
  LEVEL_TOO_LOW: "سطحت برای این اتاق کم است.",
  PLATFORM_NOT_SUPPORTED: "این بازی روی این دستگاه نیست.",
  UPDATE_REQUIRED: "برای این بازی باید بازی را به‌روز کنی.",
  MATCH_STARTED: "این بازی شروع شده است.",
  MATCH_CANCELLED: "این بازی لغو شده است.",
  NOT_YOUR_TURN: "نوبت تو نیست.",
  CELL_TAKEN: "این خانه پر است.",
});

Object.assign(STRINGS.en.errors, {
  GAME_DISABLED: "This game is switched off for now.",
  GAME_NOT_FOUND: "This game was not found.",
  ROOM_NOT_FOUND: "This room was not found.",
  ROOM_REQUIRED: "Pick a room.",
  LEVEL_TOO_LOW: "Your level is too low for this room.",
  PLATFORM_NOT_SUPPORTED: "This game is not on this device.",
  UPDATE_REQUIRED: "Update the game to play this.",
  MATCH_STARTED: "This match has already started.",
  MATCH_CANCELLED: "This match was cancelled.",
  NOT_YOUR_TURN: "It's not your turn.",
  CELL_TAKEN: "That cell is taken.",
});

for (const s of ["games", "ttt"]) if (!SCREENS.includes(s)) SCREENS.push(s);

const lobby = { games: [], ticket: null, matchId: null, me: null, view: null, seq: 0, timer: null, searchFor: null };

function lobbyLabel(text, cls) {
  return el("span", { class: "chip" + (cls ? " " + cls : ""), text });
}

function lobbyRoomCard(g, r) {
  const fee = r.entryFee ? t("gameFee", r.entryFee.amount, r.entryFee.currencyName || r.entryFee.currencyKey) : t("gameFree");
  const btn = el("button", { class: "btn btn-primary", type: "button", "data-room": r.key, text: t("gameEnter"), disabled: !g.isActive || r.locked, onclick: () => lobbyJoin(g, r) });
  return el("div", { class: "room-card" + (r.locked ? " locked" : ""), "data-room-card": r.key }, [
    el("b", { class: "room-name", text: r.name }),
    el("span", { class: "room-fee", "data-fee": r.entryFee ? String(r.entryFee.amount) : "0", text: fee }),
    r.prize ? el("span", { class: "muted small room-prize", text: t("gamePrize", r.prize.name) }) : null,
    r.minLevel ? el("span", { class: "muted small", "data-min-level": String(r.minLevel), text: t("gameMinLevel", r.minLevel) }) : null,
    btn,
  ]);
}

function lobbyCard(g) {
  const head = el("div", { class: "game-head" }, [
    g.iconUrl ? el("img", { class: "game-icon", src: g.iconUrl, alt: "", width: "56", height: "56" }) : el("span", { class: "game-icon" }),
    el("div", { class: "game-title" }, [
      el("h2", { text: g.name }),
      el("div", { class: "chips" }, [
        g.category ? lobbyLabel(g.category.name) : null,
        lobbyLabel(g.maxPlayers > 1 ? t("gamePlayers", g.minPlayers, g.maxPlayers) : t("gameSolo")),
        g.turnSeconds ? lobbyLabel(t("gameTurn", g.turnSeconds)) : null,
        g.key === "stargame" && g.settings && g.settings.roundSeconds ? lobbyLabel(t("gameRoundTime", g.settings.roundSeconds), "chip-gold") : null,
        !g.isActive ? lobbyLabel(t("gameMaintenance"), "chip-bad") : null,
      ]),
    ]),
  ]);
  const body = [head, g.description ? el("p", { class: "muted small", text: g.description }) : null];
  if (!g.isActive && g.maintenance) body.push(el("p", { class: "game-maint", role: "status", text: g.maintenance }));
  if (g.key === "stargame") {
    body.push(el("button", { class: "btn btn-primary", type: "button", "data-play": g.key, text: t("gamePlay"), disabled: !g.isActive, onclick: () => startGame() }));
  } else if (g.rooms && g.rooms.length) {
    body.push(el("div", { class: "rooms" }, g.rooms.map((r) => lobbyRoomCard(g, r))));
  }
  return el("article", { class: "panel game-card", "data-game": g.key, "data-active": g.isActive ? "true" : "false" }, body);
}

async function loadLobby() {
  const box = $("#games-list");
  const my = ++lobby.seq;
  box.textContent = "";
  box.append(el("div", { class: "spinner" }));
  try {
    const r = await rpc("economy_games", { lang: i18n.lang, platform: "web", version: gameVersion() });
    if (my !== lobby.seq) return;
    lobby.games = r.items || [];
    box.textContent = "";
    if (!lobby.games.length) box.append(el("p", { class: "muted center", text: t("gamesNone") }));
    for (const g of lobby.games) box.append(lobbyCard(g));
  } catch (err) {
    if (my !== lobby.seq) return;
    box.textContent = "";
    box.append(el("p", { class: "muted center", text: errorText(err) }));
  }
}

function lobbySearching(g, r) {
  const box = $("#games-search");
  box.textContent = "";
  box.hidden = !g;
  if (!g) return;
  box.append(el("div", { class: "spinner" }), el("b", { text: t("gameSearching") }), el("span", { class: "muted small", text: `${g.name} · ${r.name}` }),
    el("button", { class: "btn btn-ghost", type: "button", id: "games-cancel", text: t("gameCancel"), onclick: lobbyCancel }));
}

async function lobbyJoin(g, r) {
  try {
    const ticket = await rpc("economy_game_join", { key: g.key, room: r.key, lang: i18n.lang, platform: "web", version: gameVersion() });
    if (!(await socketReady())) throw new ApiError("NETWORK", 0);
    lobbySearching(g, r);
    const res = await rtSend({ matchmaker_add: {
      min_count: ticket.minCount, max_count: ticket.maxCount, query: ticket.query,
      string_properties: Object.assign({ platform: "web", version: gameVersion(), lang: i18n.lang }, ticket.stringProperties),
      numeric_properties: ticket.numericProperties,
    } });
    lobby.ticket = res.matchmaker_ticket && res.matchmaker_ticket.ticket;
    lobby.searchFor = { game: g, room: r };
    document.body.dataset.matchmaking = "searching";
  } catch (err) {
    lobbySearching(null);
    const notice = parseNotice(err && err.message);
    toast(errorText(notice ? new ApiError(notice.code, 0) : err), "bad");
  }
}

async function lobbyCancel() {
  const ticket = lobby.ticket;
  lobby.ticket = null;
  lobbySearching(null);
  document.body.dataset.matchmaking = "";
  if (ticket) await rtSend({ matchmaker_remove: { ticket } }).catch(() => {});
}

function lobbyB64(text) {
  return btoa(String.fromCharCode(...new TextEncoder().encode(text)));
}

function lobbyUnb64(data) {
  try {
    return JSON.parse(new TextDecoder().decode(Uint8Array.from(atob(data || ""), (c) => c.charCodeAt(0))));
  } catch (e) {
    return {};
  }
}

async function lobbyMatched(m) {
  lobby.ticket = null;
  document.body.dataset.matchmaking = "matched";
  toast(t("gameFound"));
  try {
    const res = await rtSend({ match_join: m.match_id ? { match_id: m.match_id } : { token: m.token } });
    lobby.matchId = res.match && res.match.match_id;
    lobby.me = res.match && res.match.self ? res.match.self.user_id : null;
    lobby.view = null;
    lobbySearching(null);
    show("ttt");
  } catch (err) {
    lobbySearching(null);
    const notice = parseNotice(err && err.message);
    toast(errorText(notice ? new ApiError(notice.code, 0) : new ApiError(typeof (err && err.message) === "string" && /^[A-Z_]+$/.test(err.message) ? err.message : "NETWORK", 0)), "bad");
  }
}

function tttPaint() {
  const v = lobby.view;
  const board = $("#ttt-board");
  const status = $("#ttt-status");
  const result = $("#ttt-result");
  board.textContent = "";
  status.textContent = "";
  result.textContent = "";
  result.hidden = true;
  if (!v) {
    status.append(el("div", { class: "spinner" }), el("span", { text: t("gameSearching") }));
    return;
  }
  const mine = v.marks ? v.marks[lobby.me] : "";
  const myTurn = !v.winner && v.reason === undefined && v.turn === lobby.me;
  board.dataset.mark = mine || "";
  board.dataset.turn = v.reason !== undefined ? "over" : myTurn ? "me" : "them";
  (v.board || []).forEach((c, i) => {
    board.append(el("button", { class: "ttt-cell" + (c ? " mark-" + c.toLowerCase() : ""), type: "button", "data-cell": String(i), "aria-label": t("tttCell", i + 1) + (c ? " " + c : ""),
      disabled: !myTurn || !!c, text: c || "", onclick: () => tttMove(i) }));
  });
  if (v.reason === undefined) {
    status.append(el("b", { text: mine ? t("tttYou", mine) : "" }), el("span", { "data-ttt-turn": myTurn ? "me" : "them", text: myTurn ? t("tttYourTurn") : t("tttTheirTurn") }),
      el("span", { class: "muted small", id: "ttt-clock", text: t("tttLeft", lobby.left != null ? lobby.left : v.turnLeft) }));
    return;
  }
  result.hidden = false;
  const outcome = v.reason === "cancelled" ? "cancelled" : !v.winner ? "draw" : v.winner === lobby.me ? "won" : "lost";
  result.dataset.outcome = outcome;
  result.append(el("h2", { text: outcome === "won" ? t("tttWon") : outcome === "lost" ? t("tttLost") : outcome === "draw" ? t("tttDraw") : t("tttCancelled") }));
  if (v.reason && t("tttReason")[v.reason]) result.append(el("p", { class: "muted small", text: t("tttReason")[v.reason] }));
  const mineOut = (v.payouts || []).find((p) => p.userId === lobby.me);
  if (mineOut && mineOut.rewards && mineOut.rewards.length) {
    const coins = mineOut.rewards.filter((r) => r.type === "currency").reduce((a, r) => a + Number(r.value || 0), 0);
    result.append(el("p", { class: "ttt-prize", "data-coins": String(coins), text: t("tttPrize", t("coins", coins)) }));
  }
  result.append(el("button", { class: "btn btn-primary", type: "button", id: "ttt-back", text: t("tttBack"), onclick: () => show("games") }));
  refreshBalances().catch(() => {});
}

function tttMove(cell) {
  if (!lobby.matchId) return;
  rtSend({ match_data_send: { match_id: lobby.matchId, op_code: 2, data: lobbyB64(JSON.stringify({ cell })) } }).catch(() => {});
}

function tttTick() {
  clearInterval(lobby.timer);
  lobby.timer = setInterval(() => {
    if ($("#screen-ttt").hidden || !lobby.view || lobby.view.reason !== undefined) return;
    lobby.left = Math.max(0, (lobby.left != null ? lobby.left : lobby.view.turnLeft) - 1);
    const c = $("#ttt-clock");
    if (c) c.textContent = t("tttLeft", lobby.left);
  }, 1000);
}

function lobbySocket(msg) {
  if (msg && msg.matchmaker_matched) {
    lobbyMatched(msg.matchmaker_matched);
    return true;
  }
  const d = msg && msg.match_data;
  if (!d || d.match_id !== lobby.matchId) return false;
  const op = Number(d.op_code);
  const body = lobbyUnb64(d.data);
  if (op === 1 || op === 3 || op === 4) {
    if (op === 4 && body.reason === "cancelled" && !body.board) body.board = [];
    lobby.view = body;
    lobby.left = body.turnLeft;
    if (!$("#screen-ttt").hidden) tttPaint();
    tttTick();
  } else if (op === 5) {
    toast(errorText(new ApiError(body.code || "INTERNAL_ERROR", 0)), "bad");
  }
  return true;
}

function lobbyShow(name) {
  if (name === "games") loadLobby();
  if (name === "ttt") tttPaint();
  if (name !== "ttt" && name !== "games" && lobby.matchId && lobby.view && lobby.view.reason !== undefined) {
    rtSend({ match_leave: { match_id: lobby.matchId } }).catch(() => {});
    lobby.matchId = null;
  }
}

window.__lobby = lobby;
