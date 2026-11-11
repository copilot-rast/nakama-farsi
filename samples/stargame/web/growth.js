Object.assign(STRINGS.fa, {
  support: "پشتیبانی",
  referral: "کد معرف",
  activity: "فعالیت من",
  goldenRush: "طوفان طلایی: ستاره‌های طلایی دو برابر!",
  liveChat: "گفتگوی زنده",
  liveChatOpen: "گفتگو با پشتیبانی",
  liveChatResume: "ادامهٔ گفتگو",
  liveOnline: "یکی از پشتیبان‌ها آنلاین است",
  liveOffline: "الان کسی آنلاین نیست؛ پیامت می‌ماند یا تیکت بفرست",
  liveOff: "گفتگوی زنده فعلاً خاموش است",
  newTicket: "تیکت تازه",
  department: "بخش",
  subject: "موضوع",
  ticketBody: "چه شده؟",
  attach: "پیوست",
  attachHint: (n, kb) => `تا ${num(n)} فایل، هر کدام تا ${num(kb)} کیلوبایت (عکس یا pdf)`,
  uploading: (p) => `در حال بارگذاری… ${num(p)}٪`,
  ticketSend: "فرستادن تیکت",
  ticketSent: "تیکتت ثبت شد. جواب پشتیبانی را همین‌جا و در صندوق پیام می‌بینی.",
  myTickets: "تیکت‌های من",
  noTickets: "هنوز تیکتی نفرستاده‌ای.",
  newReply: "پاسخ تازه",
  yourReply: "پاسخ تو",
  ticketSolved: "مشکلم حل شد",
  ticketSolvedDone: "خوشحالیم که مشکلت حل شد.",
  ticketClosed: "این تیکت بسته شده است. اگر باز هم مشکلی هست، تیکت تازه بفرست.",
  rateTitle: "از پاسخ پشتیبانی راضی بودی؟",
  rateComment: "اگر خواستی، چند کلمه بنویس",
  rateSend: "ثبت امتیاز",
  rateThanks: "ممنون از امتیازت!",
  rated: (n) => `امتیاز تو: ${num(n)} از ${num(5)}`,
  star: (n) => `${num(n)} ستاره`,
  ticketStatus: {
    open: "ثبت شد",
    awaiting_staff: "منتظر پشتیبانی",
    awaiting_player: "منتظر جواب تو",
    resolved: "حل شد",
    closed: "بسته شد",
  },
  you: "تو",
  supportTeam: "پشتیبانی",
  systemMsg: "سیستم",
  agentTyping: "پشتیبان دارد می‌نویسد…",
  chatWaiting: "در صف هستی؛ به‌زودی یکی از پشتیبان‌ها جواب می‌دهد.",
  chatWith: (name) => `در حال گفتگو با ${name}`,
  chatEnded: "گفتگو تمام شد.",
  chatEnd: "پایان گفتگو",
  chatToTicket: "پایان و ساختن تیکت",
  chatTicketMade: (n) => `گفتگو به تیکت ${n} تبدیل شد.`,
  chatNew: "گفتگوی تازه",
  chatEvent: {
    accepted: (a) => `${a} گفتگو را شروع کرد`,
    transferred: (a) => (a ? `گفتگو به ${a} سپرده شد` : "گفتگو به صف برگشت"),
    closed: "گفتگو بسته شد",
  },
  myCodeHint: "این کد را به دوستت بده تا در هفتهٔ اول بازی واردش کند.",
  copy: "کپی",
  copied: "کد کپی شد.",
  refRewardBoth: (a, b) => `تو ${a} می‌گیری و دوستت ${b}.`,
  refGoal: (event, n) => `جایزه‌ها بعد از اینکه دوستت ${num(n)} بار «${event}» را انجام داد داده می‌شوند.`,
  refCounts: (total, done) => `${num(total)} دوست با کد تو آمده‌اند؛ ${num(done)} نفر شرط را کامل کرده‌اند.`,
  refStatus: { pending: "منتظر بازی", completed: "جایزه داده شد", rejected: "رد شد", held: "در حال بررسی" },
  enterCode: "کد دوستت را وارد کن",
  enterCodeHint: (date) => `فقط تا ${date} می‌توانی کد معرف وارد کنی.`,
  enterCodeOk: "ثبت کد",
  enterCodeDone: "کد ثبت شد!",
  enterCodeWait: "کد ثبت شد. جایزه‌ها بعد از کامل کردن شرط می‌رسد.",
  enterCodeHeld: "کد ثبت شد و در حال بررسی است.",
  referredBy: (name) => `تو با کد ${name} آمده‌ای.`,
  cantEnter: "دیگر نمی‌توانی کد معرف وارد کنی.",
  actAll: "همه",
  actEmpty: "هنوز فعالیتی نیست.",
  act: {
    signin: () => "وارد بازی شدی",
    purchase: (d) => `خرید ${d.productKey || ""}`,
    prize: (d, a) => (a === "gift_redeemed" ? `کد هدیهٔ ${d.code || ""} را وارد کردی` : `جایزهٔ «${d.prize || ""}» را گرفتی`),
    event: (d) => `به سطح ${num(d.to || 0)} رسیدی`,
    ticket: (d, a) => ({ created: `تیکت ${d.number || ""} را فرستادی`, staff_reply: `پشتیبانی به تیکت ${d.number || ""} جواب داد`, resolved: `تیکت ${d.number || ""} حل شد`, closed: `تیکت ${d.number || ""} بسته شد`, rated: `به تیکت ${d.number || ""} امتیاز دادی`, resolved_by_player: `تیکت ${d.number || ""} را حل‌شده زدی`, created_by_staff: `پشتیبانی تیکت ${d.number || ""} را برایت باز کرد` }[a] || "تیکت"),
    chat: (d, a) => ({ opened: "گفتگوی زنده را شروع کردی", taken: "پشتیبان به گفتگو آمد", transferred: "گفتگو به پشتیبان دیگری رسید", closed: "گفتگوی زنده تمام شد" }[a] || "گفتگوی پشتیبانی"),
    referral: (d, a) => ({ entered: `کد معرف ${d.referrer || ""} را وارد کردی`, joined: `${d.referee || ""} با کد تو آمد`, completed: "شرط کد معرف کامل شد", approved: "کد معرف تأیید شد", rejected: "کد معرف رد شد" }[a] || "کد معرف"),
    restriction: (d, a) => ({ suspend: "حسابت موقتاً بسته شد", mute: "موقتاً نمی‌توانستی پیام بدهی", warn: "اخطار گرفتی", lifted: "محدودیت برداشته شد" }[a] || "محدودیت"),
  },
  actKind: { signin: "ورود", purchase: "خرید", prize: "جایزه", event: "سطح", ticket: "پشتیبانی", chat: "گفتگو", referral: "معرف", restriction: "محدودیت" },
});
Object.assign(STRINGS.fa.errors, {
  TOO_MANY_OPEN_TICKETS: "چند تیکت باز داری؛ اول یکی را ببند.",
  TOO_MANY_TICKETS: "در یک ساعت تیکت زیادی فرستادی؛ کمی صبر کن.",
  TOO_MANY_MESSAGES: "پیام زیادی فرستادی؛ کمی صبر کن.",
  TOO_MANY_UPLOADS: "فایل زیادی فرستادی؛ کمی صبر کن.",
  TOO_MANY_CHATS: "گفتگوی زیادی باز کردی؛ کمی صبر کن.",
  INVALID_SUBJECT: "موضوع را بنویس (حداکثر ۱۵۰ حرف).",
  BODY_REQUIRED: "پیامت را بنویس.",
  BODY_TOO_LONG: "پیام خیلی بلند است.",
  TICKET_CLOSED: "این تیکت بسته شده است.",
  TICKET_NOT_FOUND: "این تیکت پیدا نشد.",
  TICKET_NOT_FINISHED: "بعد از حل شدن تیکت می‌توانی امتیاز بدهی.",
  ALREADY_RATED: "قبلاً امتیاز داده‌ای.",
  ATTACHMENT_TOO_LARGE: "فایل خیلی بزرگ است.",
  ATTACHMENT_TYPE: "فقط عکس png، jpg، webp یا فایل pdf.",
  TOO_MANY_ATTACHMENTS: "پیوست زیادی گذاشتی.",
  ATTACHMENTS_UNAVAILABLE: "فرستادن فایل فعلاً ممکن نیست.",
  CHAT_DISABLED: "گفتگوی زنده فعلاً خاموش است.",
  CHAT_CLOSED: "این گفتگو تمام شده است.",
  CHAT_EMPTY: "گفتگو هنوز پیامی ندارد.",
  CANNOT_RATE: "این گفتگو را نمی‌شود امتیاز داد.",
  CODE_NOT_FOUND: "این کد پیدا نشد. حروف را دوباره نگاه کن.",
  SELF_REFERRAL: "نمی‌توانی کد خودت را وارد کنی.",
  ALREADY_REFERRED: "قبلاً یک کد معرف وارد کرده‌ای.",
  ACCOUNT_TOO_OLD: "کد معرف فقط برای بازیکن‌های تازه است.",
  REFERRAL_CYCLE: "این دوستت را خودت دعوت کرده‌ای.",
  REFERRER_LIMIT: "این کد دیگر پذیرفته نمی‌شود.",
  DEVICE_ALREADY_USED: "روی این دستگاه قبلاً کد معرف استفاده شده است.",
  PHONE_ALREADY_USED: "با این شماره قبلاً کد معرف استفاده شده است.",
  REFERRAL_DISABLED: "کد معرف فعلاً خاموش است.",
  TOO_MANY_TRIES: "چند بار اشتباه زدی؛ کمی صبر کن.",
});
Object.assign(STRINGS.en, {
  support: "Support",
  referral: "Referral code",
  activity: "My activity",
  goldenRush: "Golden rush: twice the golden stars!",
  liveChat: "Live chat",
  liveChatOpen: "Chat with support",
  liveChatResume: "Continue the chat",
  liveOnline: "Someone from support is online",
  liveOffline: "Nobody is online right now; your message stays, or send a ticket",
  liveOff: "Live chat is off for now",
  newTicket: "New ticket",
  department: "Department",
  subject: "Subject",
  ticketBody: "What happened?",
  attach: "Attach",
  attachHint: (n, kb) => `Up to ${num(n)} files, ${num(kb)} KB each (image or pdf)`,
  uploading: (p) => `Uploading… ${num(p)}%`,
  ticketSend: "Send ticket",
  ticketSent: "Your ticket is in. You will see the answer here and in your inbox.",
  myTickets: "My tickets",
  noTickets: "You have not sent a ticket yet.",
  newReply: "New reply",
  yourReply: "Your reply",
  ticketSolved: "My problem is solved",
  ticketSolvedDone: "Glad it is solved.",
  ticketClosed: "This ticket is closed. If there is still a problem, send a new ticket.",
  rateTitle: "Were you happy with the answer?",
  rateComment: "Add a few words if you like",
  rateSend: "Send rating",
  rateThanks: "Thanks for rating!",
  rated: (n) => `Your rating: ${num(n)} of ${num(5)}`,
  star: (n) => `${num(n)} stars`,
  ticketStatus: { open: "Sent", awaiting_staff: "Waiting for support", awaiting_player: "Waiting for you", resolved: "Solved", closed: "Closed" },
  you: "You",
  supportTeam: "Support",
  systemMsg: "System",
  agentTyping: "Support is typing…",
  chatWaiting: "You are in the queue; someone from support will answer soon.",
  chatWith: (name) => `Chatting with ${name}`,
  chatEnded: "The chat ended.",
  chatEnd: "End chat",
  chatToTicket: "End and make a ticket",
  chatTicketMade: (n) => `The chat became ticket ${n}.`,
  chatNew: "New chat",
  chatEvent: {
    accepted: (a) => `${a} joined the chat`,
    transferred: (a) => (a ? `The chat went to ${a}` : "The chat went back to the queue"),
    closed: "The chat was closed",
  },
  myCodeHint: "Give this code to a friend to enter in their first week.",
  copy: "Copy",
  copied: "Code copied.",
  refRewardBoth: (a, b) => `You get ${a} and your friend gets ${b}.`,
  refGoal: (event, n) => `Rewards are paid after your friend does "${event}" ${num(n)} times.`,
  refCounts: (total, done) => `${num(total)} friends came with your code; ${num(done)} reached the goal.`,
  refStatus: { pending: "Waiting to play", completed: "Rewarded", rejected: "Rejected", held: "Being checked" },
  enterCode: "Enter your friend's code",
  enterCodeHint: (date) => `You can enter a referral code until ${date}.`,
  enterCodeOk: "Enter code",
  enterCodeDone: "Code entered!",
  enterCodeWait: "Code entered. The rewards come after the goal is met.",
  enterCodeHeld: "Code entered and being checked.",
  referredBy: (name) => `You came with the code of ${name}.`,
  cantEnter: "You can no longer enter a referral code.",
  actAll: "All",
  actEmpty: "No activity yet.",
  act: {
    signin: () => "You signed in",
    purchase: (d) => `Bought ${d.productKey || ""}`,
    prize: (d, a) => (a === "gift_redeemed" ? `You redeemed gift code ${d.code || ""}` : `You got the prize "${d.prize || ""}"`),
    event: (d) => `You reached level ${num(d.to || 0)}`,
    ticket: (d, a) => ({ created: `You sent ticket ${d.number || ""}`, staff_reply: `Support answered ticket ${d.number || ""}`, resolved: `Ticket ${d.number || ""} solved`, closed: `Ticket ${d.number || ""} closed`, rated: `You rated ticket ${d.number || ""}`, resolved_by_player: `You marked ticket ${d.number || ""} solved`, created_by_staff: `Support opened ticket ${d.number || ""} for you` }[a] || "Ticket"),
    chat: (d, a) => ({ opened: "You started a live chat", taken: "Support joined the chat", transferred: "The chat moved to another agent", closed: "The live chat ended" }[a] || "Support chat"),
    referral: (d, a) => ({ entered: `You entered the code of ${d.referrer || ""}`, joined: `${d.referee || ""} joined with your code`, completed: "Referral goal reached", approved: "Referral approved", rejected: "Referral rejected" }[a] || "Referral"),
    restriction: (d, a) => ({ suspend: "Your account was suspended for a while", mute: "You could not chat for a while", warn: "You got a warning", lifted: "A restriction was lifted" }[a] || "Restriction"),
  },
  actKind: { signin: "Sign-in", purchase: "Purchase", prize: "Prize", event: "Level", ticket: "Support", chat: "Chat", referral: "Referral", restriction: "Restriction" },
});
Object.assign(STRINGS.en.errors, {
  TOO_MANY_OPEN_TICKETS: "You have several open tickets; close one first.",
  TOO_MANY_TICKETS: "Too many tickets this hour; wait a little.",
  TOO_MANY_MESSAGES: "Too many messages; wait a little.",
  TOO_MANY_UPLOADS: "Too many files; wait a little.",
  TOO_MANY_CHATS: "Too many chats; wait a little.",
  INVALID_SUBJECT: "Write a subject (at most 150 letters).",
  BODY_REQUIRED: "Write your message.",
  BODY_TOO_LONG: "The message is too long.",
  TICKET_CLOSED: "This ticket is closed.",
  TICKET_NOT_FOUND: "This ticket was not found.",
  TICKET_NOT_FINISHED: "You can rate once the ticket is solved.",
  ALREADY_RATED: "You already rated this.",
  ATTACHMENT_TOO_LARGE: "The file is too large.",
  ATTACHMENT_TYPE: "Only png, jpg, webp images or pdf files.",
  TOO_MANY_ATTACHMENTS: "Too many attachments.",
  ATTACHMENTS_UNAVAILABLE: "Sending files is not possible right now.",
  CHAT_DISABLED: "Live chat is off for now.",
  CHAT_CLOSED: "This chat has ended.",
  CHAT_EMPTY: "The chat has no messages yet.",
  CANNOT_RATE: "This chat cannot be rated.",
  CODE_NOT_FOUND: "This code was not found. Check the letters again.",
  SELF_REFERRAL: "You cannot enter your own code.",
  ALREADY_REFERRED: "You already entered a referral code.",
  ACCOUNT_TOO_OLD: "Referral codes are only for new players.",
  REFERRAL_CYCLE: "You invited this friend yourself.",
  REFERRER_LIMIT: "This code is no longer accepted.",
  DEVICE_ALREADY_USED: "A referral code was already used on this device.",
  PHONE_ALREADY_USED: "A referral code was already used with this number.",
  REFERRAL_DISABLED: "Referral codes are off for now.",
  TOO_MANY_TRIES: "Too many tries; wait a little.",
});

const growth = {
  config: null,
  configAt: 0,
  info: null,
  files: [],
  replyFiles: [],
  ticketId: null,
  ticket: null,
  chat: null,
  chatSeq: 0,
  chatTimer: 0,
  typingAt: 0,
  actKind: "",
  actBefore: null,
};

async function loadConfig() {
  try {
    growth.config = await rpc("base_remote_config", { platform: PLATFORM, version: gameVersion(), lang: i18n.lang });
    growth.configAt = Date.now();
  } catch (e) {
    growth.config = growth.config || null;
  }
  paintConfig();
}

function configValue(key, fallback) {
  const v = growth.config && growth.config.values ? growth.config.values[key] : undefined;
  return v === undefined || v === null ? fallback : v;
}

function goldenRush() {
  return configValue("stargame.golden_rush", false) === true;
}

function paintConfig() {
  const banner = $("#config-banner");
  if (!banner) return;
  const text = configValue("stargame.banner", "");
  banner.textContent = typeof text === "string" ? text : "";
  banner.hidden = !banner.textContent;
  if (growth.config && growth.config.variants) banner.dataset.variant = growth.config.variants["stargame.banner"] || "";
  $("#rush-badge").hidden = !goldenRush();
}

function growthShow(name) {
  if (name === "home") {
    paintConfig();
    if (Date.now() - growth.configAt > 30000) loadConfig();
    refreshSupportDot();
  }
  if (name === "support") loadSupport();
  if (name === "livechat") openLiveChat();
  else stopChatPoll();
  if (name === "referral") loadReferral();
  if (name === "activity") loadActivity(true);
}

async function refreshSupportDot() {
  try {
    const info = await rpc("base_support_info", { lang: i18n.lang });
    growth.info = info;
    const dot = $("#support-dot");
    dot.hidden = !info.unread;
    dot.textContent = info.unread ? num(info.unread) : "";
  } catch (e) {}
}

function statusChip(status) {
  return el("span", { class: "status-chip s-" + status, "data-status": status, text: (t("ticketStatus") || {})[status] || status });
}

async function loadSupport() {
  const list = $("#ticket-list");
  list.textContent = "";
  list.append(el("div", { class: "center-block" }, [el("div", { class: "spinner" })]));
  try {
    const [info, mine] = await Promise.all([rpc("base_support_info", { lang: i18n.lang }), rpc("base_support_tickets", { lang: i18n.lang })]);
    growth.info = info;
    const sel = $("#ticket-dept");
    const keep = sel.value;
    sel.textContent = "";
    for (const d of info.departments || []) sel.append(el("option", { value: d.key, text: d.name }));
    if (keep && (info.departments || []).some((d) => d.key === keep)) sel.value = keep;
    paintDeptHint();
    $("#ticket-attach-hint").textContent = info.attachments ? t("attachHint", info.maxAttachments, info.maxAttachmentKb) : "";
    $("#ticket-form .attach-row").hidden = !info.attachments;
    const live = $("#live-state");
    live.textContent = "";
    live.append(el("span", { class: "live-dot" + (info.agentsOnline ? " on" : "") }), info.chatEnabled ? (info.agentsOnline ? t("liveOnline") : t("liveOffline")) : t("liveOff"));
    $("#btn-livechat").disabled = !info.chatEnabled;
    $("#btn-livechat").textContent = info.chatId ? t("liveChatResume") : t("liveChatOpen");
    list.textContent = "";
    const items = mine.items || [];
    if (!items.length) list.append(el("p", { class: "muted center", text: t("noTickets") }));
    for (const tk of items) {
      list.append(el("button", { class: "ticket-item", type: "button", "data-ticket": tk.id, onclick: () => openTicket(tk.id) }, [
        el("div", { class: "row" }, [el("b", { text: tk.subject, dir: "auto" }), statusChip(tk.status)]),
        el("div", { class: "row" }, [el("span", { class: "muted small mono", text: tk.number }), tk.unread ? el("span", { class: "new", text: t("newReply") }) : el("span", { class: "muted small", text: formatDate(tk.lastMessageAt) })]),
      ]));
    }
    const dot = $("#support-dot");
    dot.hidden = !info.unread;
    dot.textContent = info.unread ? num(info.unread) : "";
  } catch (err) {
    list.textContent = "";
    list.append(el("p", { class: "muted center", text: errorText(err) }));
  }
}

function paintDeptHint() {
  const d = ((growth.info && growth.info.departments) || []).find((x) => x.key === $("#ticket-dept").value);
  $("#ticket-dept-hint").textContent = d ? d.description || "" : "";
}

function b64(bytes) {
  let s = "";
  for (let i = 0; i < bytes.length; i += 0x8000) s += String.fromCharCode.apply(null, bytes.subarray(i, i + 0x8000));
  return btoa(s);
}

async function uploadFile(file, onprogress) {
  const data = new Uint8Array(await file.arrayBuffer());
  const begin = await rpc("base_support_upload_begin", { name: file.name, size: data.length });
  const size = begin.chunkBytes;
  for (let i = 0; i < begin.parts; i++) {
    await rpc("base_support_upload_part", { uploadId: begin.uploadId, index: i, data: b64(data.subarray(i * size, Math.min(data.length, (i + 1) * size))) });
    onprogress(Math.round(((i + 1) / begin.parts) * 100));
  }
  return rpc("base_support_upload_finish", { uploadId: begin.uploadId });
}

function paintFiles(box, list) {
  box.textContent = "";
  list.forEach((f, i) => {
    box.append(el("span", { class: "file-chip", "data-file": f.key }, [el("span", { text: f.name }), el("button", { type: "button", "aria-label": t("cancel"), text: "×", onclick: () => { list.splice(i, 1); paintFiles(box, list); } })]));
  });
}

function setupFileInput(input, box, list, msg) {
  input.addEventListener("change", async () => {
    const f = input.files && input.files[0];
    input.value = "";
    if (!f) return;
    const max = (growth.info && growth.info.maxAttachments) || 5;
    if (list.length >= max) { toast(errorText({ code: "TOO_MANY_ATTACHMENTS" }), "bad"); return; }
    if (growth.info && f.size > growth.info.maxAttachmentKb * 1024) { toast(errorText({ code: "ATTACHMENT_TOO_LARGE" }), "bad"); return; }
    try {
      const a = await uploadFile(f, (p) => { if (msg) msg.textContent = t("uploading", p); });
      if (msg) msg.textContent = "";
      list.push(a);
      paintFiles(box, list);
    } catch (err) {
      if (msg) msg.textContent = "";
      toast(errorText(err), "bad");
    }
  });
}

async function sendTicket(e) {
  e.preventDefault();
  const msg = $("#ticket-msg");
  msg.className = "form-msg";
  msg.textContent = "";
  const btn = $("#ticket-send");
  btn.disabled = true;
  try {
    const tk = await rpc("base_support_ticket_create", { departmentKey: $("#ticket-dept").value, subject: $("#ticket-subject").value, body: $("#ticket-body").value, attachments: growth.files.map((f) => f.key), lang: i18n.lang });
    $("#ticket-subject").value = "";
    $("#ticket-body").value = "";
    growth.files = [];
    paintFiles($("#ticket-files"), growth.files);
    toast(t("ticketSent"), "ok");
    openTicket(tk.id);
  } catch (err) {
    msg.className = "form-msg bad";
    msg.textContent = errorText(err);
  } finally {
    btn.disabled = false;
  }
}

function msgFiles(list) {
  if (!list || !list.length) return null;
  return el("div", { class: "msg-files" }, list.map((a) => el("a", { href: (a.url || "").startsWith("/") ? API + a.url : a.url, target: "_blank", rel: "noopener", "data-attachment": a.key }, [
    /^image\//.test(a.type) ? el("img", { src: (a.url || "").startsWith("/") ? API + a.url : a.url, alt: a.name }) : a.name,
  ])));
}

function ticketLine(m) {
  if (m.kind === "event") {
    const d = m.eventData || {};
    const text = m.event === "status_changed" ? (t("ticketStatus") || {})[d.to] || d.to : m.event === "rated" ? t("rated", d.rating || 0) : m.event === "created_from_chat" ? t("liveChat") : m.event;
    return el("div", { class: "chat-msg system-msg", text: text + " · " + formatDate(m.createdAt) });
  }
  const staff = m.authorType === "staff";
  return el("div", { class: "chat-msg " + (staff ? "staff" : "mine"), "data-msg": m.authorType }, [
    el("div", { class: "chat-meta" }, [el("span", { class: "chat-name", text: staff ? t("supportTeam") : t("you") }), el("span", { class: "muted small", text: formatDate(m.createdAt) })]),
    m.body ? el("p", { class: "chat-text", dir: "auto", text: m.body }) : null,
    msgFiles(m.attachments),
  ]);
}

async function openTicket(id) {
  growth.ticketId = id;
  show("ticket");
  const thread = $("#ticket-thread");
  thread.textContent = "";
  thread.append(el("div", { class: "center-block" }, [el("div", { class: "spinner" })]));
  try {
    if (!growth.info) growth.info = await rpc("base_support_info", { lang: i18n.lang });
    const tk = await rpc("base_support_ticket_get", { id, lang: i18n.lang });
    paintTicket(tk);
    refreshSupportDot();
  } catch (err) {
    thread.textContent = "";
    thread.append(el("div", { class: "empty", text: errorText(err) }));
  }
}

function paintTicket(tk) {
  growth.ticket = tk;
  $("#ticket-title").textContent = tk.subject;
  $("#ticket-number").textContent = tk.number;
  const meta = $("#ticket-meta");
  meta.textContent = "";
  meta.append(statusChip(tk.status), el("span", { class: "muted small", text: tk.departmentName || "" }));
  const thread = $("#ticket-thread");
  thread.textContent = "";
  for (const m of tk.messages || []) thread.append(ticketLine(m));
  thread.scrollTop = thread.scrollHeight;
  const form = $("#ticket-reply");
  form.hidden = tk.status === "closed";
  $("#ticket-resolve").hidden = tk.status === "resolved" || tk.status === "closed";
  form.querySelector(".attach-row").hidden = !(growth.info && growth.info.attachments);
  if (tk.status === "closed") meta.append(el("p", { class: "muted small", text: t("ticketClosed") }));
  paintRate(tk);
}

function starPicker(onpick) {
  let value = 0;
  const box = el("div", { class: "stars", role: "radiogroup", "aria-label": t("rateTitle") });
  const btns = [1, 2, 3, 4, 5].map((n) => el("button", { class: "star-btn", type: "button", role: "radio", "aria-pressed": "false", "aria-label": t("star", n), "data-star": String(n), text: "★", onclick: () => {
    value = n;
    btns.forEach((b, i) => { b.setAttribute("aria-pressed", i < n ? "true" : "false"); b.setAttribute("aria-checked", i === n - 1 ? "true" : "false"); });
    onpick(n);
  } }));
  box.append(...btns);
  box.value = () => value;
  return box;
}

function paintRate(tk) {
  const box = $("#ticket-rate");
  box.textContent = "";
  if (tk.rating) {
    box.hidden = false;
    box.append(el("p", { "data-rated": String(tk.rating), text: t("rated", tk.rating) }));
    return;
  }
  if (!tk.canRate) { box.hidden = true; return; }
  box.hidden = false;
  let rating = 0;
  const comment = el("input", { type: "text", maxlength: "500", placeholder: t("rateComment"), "aria-label": t("rateComment") });
  const send = el("button", { class: "btn btn-primary", type: "button", disabled: true, text: t("rateSend"), "data-rate-send": "" });
  const stars = starPicker((n) => { rating = n; send.disabled = false; });
  send.addEventListener("click", async () => {
    send.disabled = true;
    try {
      const res = await rpc("base_support_ticket_rate", { id: tk.id, rating, comment: comment.value, lang: i18n.lang });
      toast(t("rateThanks"), "ok");
      paintTicket(res);
    } catch (err) {
      toast(errorText(err), "bad");
      send.disabled = false;
    }
  });
  box.append(el("h2", { text: t("rateTitle") }), stars, comment, send);
}

async function sendReply(e) {
  e.preventDefault();
  const btn = $("#ticket-reply-send");
  btn.disabled = true;
  try {
    const res = await rpc("base_support_ticket_reply", { id: growth.ticketId, body: $("#ticket-reply-body").value, attachments: growth.replyFiles.map((f) => f.key), lang: i18n.lang });
    $("#ticket-reply-body").value = "";
    growth.replyFiles = [];
    paintFiles($("#ticket-reply-files"), growth.replyFiles);
    paintTicket(res);
  } catch (err) {
    toast(errorText(err), "bad");
  } finally {
    btn.disabled = false;
  }
}

async function resolveTicket() {
  try {
    const res = await rpc("base_support_ticket_resolve", { id: growth.ticketId, lang: i18n.lang });
    toast(t("ticketSolvedDone"), "ok");
    paintTicket(res);
  } catch (err) {
    toast(errorText(err), "bad");
  }
}

function supportNotice(n) {
  let c = {};
  try {
    c = typeof n.content === "string" ? JSON.parse(n.content) : n.content || {};
  } catch (e) {
    c = {};
  }
  if (c.type === "chat") {
    if (currentScreen() === "livechat" && growth.chat && growth.chat.id === c.chatId) pollChat(true);
    return;
  }
  if (c.type === "ticket") {
    if (currentScreen() === "ticket" && growth.ticketId === c.ticketId) openTicket(c.ticketId);
    else if (currentScreen() === "support") loadSupport();
    refreshSupportDot();
  }
}

function chatLineView(m) {
  if (m.sender === "system") {
    const d = m.eventData || {};
    const ev = t("chatEvent") || {};
    let text = m.body || "";
    if (m.event === "accepted") text = ev.accepted(d.agent || t("supportTeam"));
    else if (m.event === "transferred") text = ev.transferred(d.to || "");
    else if (m.event === "closed") text = ev.closed;
    return el("div", { class: "chat-msg system-msg", "data-seq": String(m.seq), "data-event": m.event || "", dir: "auto", text });
  }
  const agent = m.sender === "agent";
  return el("div", { class: "chat-msg " + (agent ? "staff" : "mine"), "data-seq": String(m.seq), "data-sender": m.sender }, [
    el("div", { class: "chat-meta" }, [el("span", { class: "chat-name", text: agent ? m.author || t("supportTeam") : t("you") }), el("span", { class: "muted small", text: formatDate(m.createdAt) })]),
    el("p", { class: "chat-text", dir: "auto", text: m.body }),
  ]);
}

function paintChat(res, reset) {
  const chat = res.chat;
  const thread = $("#lc-thread");
  if (reset) thread.textContent = "";
  growth.chat = chat;
  const banner = $("#lc-banner");
  const actions = $("#lc-actions");
  actions.textContent = "";
  $("#lc-typing").hidden = !(chat && chat.agentTyping);
  if (!chat) {
    growth.chatSeq = 0;
    banner.hidden = false;
    banner.textContent = res.agentsOnline ? t("liveOnline") : t("liveOffline");
    $("#lc-state").textContent = "";
    $("#lc-form").hidden = false;
    return;
  }
  for (const m of chat.messages || []) {
    if (thread.querySelector(`[data-seq="${m.seq}"]`)) continue;
    thread.append(chatLineView(m));
    growth.chatSeq = Math.max(growth.chatSeq, m.seq);
  }
  thread.scrollTop = thread.scrollHeight;
  $("#lc-state").textContent = chat.status === "active" ? t("chatWith", chat.agent || t("supportTeam")) : chat.status === "waiting" ? "" : t("chatEnded");
  $("#lc-state").dataset.status = chat.status;
  banner.hidden = chat.status !== "waiting";
  banner.textContent = chat.status === "waiting" ? (res.agentsOnline ? t("chatWaiting") : t("liveOffline")) : "";
  $("#lc-form").hidden = chat.status === "closed";
  if (chat.status === "closed") {
    if (chat.ticketNumber) actions.append(el("button", { class: "btn btn-ghost", type: "button", text: t("chatTicketMade", chat.ticketNumber), onclick: () => openTicket(chat.ticketId) }));
    if (chat.canRate) {
      let rating = 0;
      const send = el("button", { class: "btn btn-primary", type: "button", disabled: true, text: t("rateSend"), "data-rate-send": "" });
      const stars = starPicker((n) => { rating = n; send.disabled = false; });
      send.addEventListener("click", async () => {
        send.disabled = true;
        try {
          const r = await rpc("base_support_chat_rate", { chatId: chat.id, rating });
          toast(t("rateThanks"), "ok");
          paintChat(r, false);
        } catch (err) {
          toast(errorText(err), "bad");
          send.disabled = false;
        }
      });
      actions.append(el("div", { class: "rate-panel" }, [el("b", { text: t("rateTitle") }), stars, send]));
    } else if (chat.rating) actions.append(el("p", { class: "muted small", "data-rated": String(chat.rating), text: t("rated", chat.rating) }));
    actions.append(el("button", { class: "btn btn-primary", type: "button", "data-chat-new": "", text: t("chatNew"), onclick: async () => { growth.chat = null; growth.chatSeq = 0; $("#lc-thread").textContent = ""; paintChat({ chat: null, agentsOnline: res.agentsOnline }, true); } }));
    return;
  }
  actions.append(
    el("button", { class: "btn btn-ghost", type: "button", "data-chat-end": "", text: t("chatEnd"), onclick: () => endChat(false) }),
    el("button", { class: "btn btn-ghost", type: "button", "data-chat-ticket": "", text: t("chatToTicket"), onclick: () => endChat(true) })
  );
}

async function openLiveChat() {
  $("#lc-thread").textContent = "";
  growth.chatSeq = 0;
  try {
    paintChat(await rpc("base_support_chat_get", {}), true);
  } catch (err) {
    $("#lc-thread").append(el("div", { class: "empty", text: errorText(err) }));
  }
  startChatPoll();
}

function startChatPoll() {
  stopChatPoll();
  growth.chatTimer = setInterval(() => { if (currentScreen() !== "livechat") { stopChatPoll(); return; } pollChat(false); }, 3000);
}

function stopChatPoll() {
  if (growth.chatTimer) clearInterval(growth.chatTimer);
  growth.chatTimer = 0;
}

let chatPolling = false;
async function pollChat(force) {
  if (!growth.chat || (chatPolling && !force)) return;
  chatPolling = true;
  try {
    const res = await rpc("base_support_chat_get", { chatId: growth.chat.id, after: growth.chatSeq });
    paintChat(res, false);
  } catch (e) {} finally {
    chatPolling = false;
  }
}

async function sendLiveChat(e) {
  e.preventDefault();
  const input = $("#lc-input");
  const body = input.value.trim();
  if (!body) return;
  const btn = $("#lc-send");
  btn.disabled = true;
  try {
    let res;
    if (!growth.chat || growth.chat.status === "closed") res = await rpc("base_support_chat_open", { body, lang: i18n.lang });
    else res = await rpc("base_support_chat_send", { chatId: growth.chat.id, body, after: growth.chatSeq });
    input.value = "";
    paintChat(res, false);
    startChatPoll();
  } catch (err) {
    toast(errorText(err), "bad");
  } finally {
    btn.disabled = false;
    input.focus();
  }
}

function chatTyping() {
  if (!growth.chat || growth.chat.status === "closed") return;
  const now = Date.now();
  if (now - growth.typingAt < 2500) return;
  growth.typingAt = now;
  rpc("base_support_chat_typing", { chatId: growth.chat.id }).catch(() => {});
}

async function endChat(toTicket) {
  if (!growth.chat) return;
  try {
    const res = await rpc("base_support_chat_close", { chatId: growth.chat.id, toTicket });
    paintChat(res, false);
    if (toTicket && res.chat && res.chat.ticketNumber) toast(t("chatTicketMade", res.chat.ticketNumber), "ok");
  } catch (err) {
    toast(errorText(err), "bad");
  }
}

function paintReferral(r) {
  $("#ref-code").textContent = r.code;
  const rw = r.rewards || {};
  $("#ref-rewards").textContent = rw.referrer || rw.referee ? t("refRewardBoth", rw.referrer ? rw.referrer.name : "—", rw.referee ? rw.referee.name : "—") : "";
  $("#ref-goal").textContent = r.goal ? t("refGoal", r.goal.eventName, r.goal.value) : "";
  const c = r.counts || {};
  $("#ref-counts").textContent = t("refCounts", c.total || 0, c.completed || 0);
  const recent = $("#ref-recent");
  recent.textContent = "";
  for (const x of r.recent || []) recent.append(el("li", {}, [el("bdi", { text: x.username }), el("span", { class: "muted small", text: (t("refStatus") || {})[x.status] || x.status })]));
  const form = $("#ref-form");
  const by = $("#ref-by");
  form.hidden = !r.canEnter;
  by.hidden = r.canEnter;
  by.textContent = "";
  if (r.referredBy) by.append(el("p", { "data-referred-by": "", text: t("referredBy", r.referredBy.username) }), el("p", { class: "muted small", text: (t("refStatus") || {})[r.referredBy.status] || "" }));
  else by.append(el("p", { class: "muted", text: t("cantEnter") }));
  $("#ref-enter-hint").textContent = r.enterUntil ? t("enterCodeHint", formatDate(r.enterUntil)) : "";
}

async function loadReferral() {
  $("#ref-msg").textContent = "";
  try {
    paintReferral(await rpc("economy_referral_me", { lang: i18n.lang }));
  } catch (err) {
    $("#ref-counts").textContent = errorText(err);
  }
}

async function enterReferral(e) {
  e.preventDefault();
  const msg = $("#ref-msg");
  msg.className = "form-msg";
  msg.textContent = "";
  const btn = $("#ref-send");
  btn.disabled = true;
  try {
    const r = await rpc("economy_referral_enter", { code: $("#ref-input").value, lang: i18n.lang });
    const st = r.entered && r.entered.status;
    msg.className = "form-msg ok";
    msg.textContent = st === "completed" ? t("enterCodeDone") : st === "held" ? t("enterCodeHeld") : t("enterCodeWait");
    toast(msg.textContent, "ok");
    paintReferral(r);
    refreshBalances().catch(() => {});
  } catch (err) {
    msg.className = "form-msg bad";
    msg.textContent = errorText(err);
  } finally {
    btn.disabled = false;
  }
}

const ACT_KINDS = ["", "signin", "purchase", "prize", "ticket", "referral"];

function paintActKinds() {
  const box = $("#act-kinds");
  box.textContent = "";
  for (const k of ACT_KINDS) {
    box.append(el("button", { class: "tab", type: "button", "aria-selected": growth.actKind === k ? "true" : "false", "data-kind": k, text: k ? (t("actKind") || {})[k] || k : t("actAll"), onclick: () => { growth.actKind = k; loadActivity(true); } }));
  }
}

async function loadActivity(reset) {
  paintActKinds();
  const list = $("#act-list");
  if (reset) {
    growth.actBefore = null;
    list.textContent = "";
  }
  try {
    const res = await rpc("base_my_activity", { limit: 20, kind: growth.actKind, before: growth.actBefore || 0 });
    for (const it of res.items || []) {
      const fn = (t("act") || {})[it.kind];
      const text = typeof fn === "function" ? fn(it.data || {}, it.action) : it.kind;
      list.append(el("li", { class: "act-item", "data-act": it.kind + ":" + it.action }, [
        el("span", { class: "act-icon k-" + it.kind, "aria-hidden": "true", text: "●" }),
        el("div", {}, [el("div", { dir: "auto", text }), el("div", { class: "act-when", text: formatDate(it.createdAt) })]),
      ]));
    }
    growth.actBefore = res.before || null;
    $("#act-more").hidden = !res.before;
    if (!list.children.length) list.append(el("li", { class: "empty", text: t("actEmpty") }));
  } catch (err) {
    list.append(el("li", { class: "empty", text: errorText(err) }));
  }
}

function setupGrowth() {
  $("#ticket-form").addEventListener("submit", sendTicket);
  $("#ticket-dept").addEventListener("change", paintDeptHint);
  $("#ticket-reply").addEventListener("submit", sendReply);
  $("#ticket-resolve").addEventListener("click", resolveTicket);
  setupFileInput($("#ticket-file"), $("#ticket-files"), growth.files, $("#ticket-msg"));
  setupFileInput($("#ticket-reply-file"), $("#ticket-reply-files"), growth.replyFiles, null);
  $("#btn-livechat").addEventListener("click", () => show("livechat"));
  $("#lc-form").addEventListener("submit", sendLiveChat);
  $("#lc-input").addEventListener("input", chatTyping);
  $("#ref-form").addEventListener("submit", enterReferral);
  $("#ref-copy").addEventListener("click", () => {
    const code = $("#ref-code").textContent;
    (navigator.clipboard ? navigator.clipboard.writeText(code) : Promise.reject()).then(() => toast(t("copied"), "ok"), () => toast(code));
  });
  $("#act-more").addEventListener("click", () => loadActivity(false));
  applyStatic();
}

setupGrowth();
