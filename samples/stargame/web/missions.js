Object.assign(STRINGS.fa, {
  tasks: "مأموریت‌ها",
  tasksHint: "کارهای کوچک بازی که هر کدام جایزه دارد. پیشرفتشان خودکار ثبت می‌شود.",
  tasksEmpty: "فعلاً مأموریتی نداری.",
  taskState: { active: "در جریان", completed: "تمام شد", claimed: "جایزه را گرفتی", locked: "قفل", cooldown: "منتظر دور بعد", done: "تمام شد" },
  taskClaim: "جایزه را بگیر",
  taskClaimed: (text) => `جایزهٔ مأموریت رسید: ${text}`,
  taskToday: "امروز",
  taskWeek: "این هفته",
  taskEnds: (date) => `تا ${date}`,
  taskNext: (date) => `دور بعد: ${date}`,
  taskLocked: (names) => `اول این را تمام کن: ${names}`,
  taskStepsOf: (n, total) => `${num(n)} از ${num(total)} قدم`,
  taskRepeatDaily: "هر روز از نو",
  taskRepeatWeekly: "هر هفته از نو",
  groupProgress: (n, total) => `${num(n)} از ${num(total)} مأموریت`,
  groupNext: "مأموریت بعدی را بده",
  groupNextAt: (date) => `مأموریت بعدی: ${date}`,
  groupFinished: "همهٔ مأموریت‌های این زنجیره را تمام کردی.",
  missionDone: (n) => (n === 1 ? "یک مأموریت تمام شد!" : `${num(n)} مأموریت تمام شد!`),
  goToTasks: "دیدن مأموریت‌ها",
});

Object.assign(STRINGS.en, {
  tasks: "Missions",
  tasksHint: "Small in-game jobs, each with a prize. Progress is saved for you.",
  tasksEmpty: "No missions for you right now.",
  taskState: { active: "In progress", completed: "Done", claimed: "Prize collected", locked: "Locked", cooldown: "Waiting for next round", done: "Done" },
  taskClaim: "Collect prize",
  taskClaimed: (text) => `Mission prize received: ${text}`,
  taskToday: "Today",
  taskWeek: "This week",
  taskEnds: (date) => `until ${date}`,
  taskNext: (date) => `Next round: ${date}`,
  taskLocked: (names) => `Finish this first: ${names}`,
  taskStepsOf: (n, total) => `${num(n)} of ${num(total)} steps`,
  taskRepeatDaily: "New every day",
  taskRepeatWeekly: "New every week",
  groupProgress: (n, total) => `${num(n)} of ${num(total)} missions`,
  groupNext: "Give me the next mission",
  groupNextAt: (date) => `Next mission: ${date}`,
  groupFinished: "You finished every mission of this chain.",
  missionDone: (n) => (n === 1 ? "A mission is done!" : `${num(n)} missions done!`),
  goToTasks: "Open missions",
});

Object.assign(STRINGS.fa.errors, {
  TASK_NOT_COMPLETED: "این مأموریت هنوز تمام نشده.",
  ALREADY_CLAIMED: STRINGS.fa.errors.ALREADY_CLAIMED || "این جایزه را قبلاً گرفته‌ای.",
  CURRENT_TASK_MUST_BE_DONE: "اول مأموریت فعلی را تمام کن.",
  DURATION_NOT_ELAPSED: "هنوز وقت مأموریت بعدی نرسیده.",
  NO_TASK_AVAILABLE: "الان مأموریت تازه‌ای نیست.",
  TASK_GROUP_FINISHED: "همهٔ مأموریت‌های این زنجیره تمام شده.",
});

Object.assign(STRINGS.en.errors, {
  TASK_NOT_COMPLETED: "This mission is not done yet.",
  ALREADY_CLAIMED: STRINGS.en.errors.ALREADY_CLAIMED || "You already collected this prize.",
  CURRENT_TASK_MUST_BE_DONE: "Finish the current mission first.",
  DURATION_NOT_ELAPSED: "The next mission is not ready yet.",
  NO_TASK_AVAILABLE: "No new mission right now.",
  TASK_GROUP_FINISHED: "Every mission of this chain is done.",
});

if (!SCREENS.includes("tasks")) SCREENS.push("tasks");

function missionTime(iso) {
  try {
    return new Intl.DateTimeFormat(i18n.locale, { weekday: "short", hour: "2-digit", minute: "2-digit" }).format(new Date(iso));
  } catch (e) {
    return iso;
  }
}

function missionNotes(ev) {
  const done = ((ev && ev.tasks) || []).filter((x) => x.completed).length;
  return done ? t("missionDone", done) : null;
}

async function refreshTasksDot() {
  if (!state.token) return;
  try {
    const res = await rpc("economy_tasks_list", { lang: i18n.lang });
    const dot = $("#tasks-dot");
    dot.hidden = !res.ready;
    dot.textContent = res.ready ? num(res.ready) : "";
  } catch (e) {}
}

async function claimMission(it, btn) {
  btn.disabled = true;
  try {
    const r = await rpc("economy_task_claim", it.claimId ? { claimId: it.claimId, lang: i18n.lang } : { taskId: it.id, lang: i18n.lang });
    toast(t("taskClaimed", rewardText(r.rewards) || (r.prize && r.prize.name) || ""), "ok");
    await refreshBalances().catch(() => {});
    refreshProgress();
  } catch (err) {
    toast(errorText(err), "bad");
  }
  loadTasks();
}

function missionCard(it, child) {
  const state = it.state;
  const ready = !!it.claimable;
  const target = Number(it.target) || 1;
  const value = Math.min(Number(it.value) || 0, target);
  const pct = state === "completed" || state === "claimed" ? 100 : Math.max(0, Math.min(100, (value / target) * 100));
  const meta = [];
  if (it.repeat === "daily") meta.push(t("taskToday"));
  if (it.repeat === "weekly") meta.push(t("taskWeek"));
  const sep = i18n.dir === "rtl" ? "، " : " · ";
  const progress = it.condition === "children" ? t("taskStepsOf", value, target) : t("progressOf", value, target);
  const foot = el("div", { class: "reward-foot" }, [el("span", { class: "reward-prize", text: it.prize ? it.prize.name : "" })]);
  if (ready) {
    const b = el("button", { class: "btn btn-primary", type: "button", "data-claim": it.key, text: t("taskClaim") });
    b.addEventListener("click", () => claimMission(it, b));
    foot.append(b);
  } else if (state === "claimed" || state === "done") foot.append(el("span", { class: "chip", text: (t("taskState") || {})[state === "claimed" && !it.prize ? "done" : state] }));
  const notes = [];
  if (state === "locked") {
    const names = (it.dependsOn || []).filter((d) => !d.done).map((d) => d.name).join(i18n.dir === "rtl" ? "، " : ", ");
    if (names) notes.push(t("taskLocked", names));
  }
  if (state === "cooldown" && it.nextAt) notes.push(t("taskNext", missionTime(it.nextAt)));
  else if (it.expiresAt && state !== "claimed") notes.push(t("taskEnds", missionTime(it.expiresAt)));
  if (!child && it.repeat === "daily") notes.push(t("taskRepeatDaily"));
  if (!child && it.repeat === "weekly") notes.push(t("taskRepeatWeekly"));
  const chipClass = state === "completed" ? "chip chip-gold" : state === "active" ? "chip chip-new" : state === "locked" ? "chip chip-wait" : "chip";
  const node = el("article", { class: "reward mission" + (ready ? " ready" : "") + (child ? " mission-step" : "") + (state === "locked" ? " locked" : ""), "data-task": it.key, "data-state": state }, [
    el("div", { class: "reward-head" }, [
      el("b", { text: it.name }),
      el("span", { class: chipClass, "data-task-state": "", text: (t("taskState") || {})[state === "claimed" && !it.prize ? "done" : state] || state }),
    ]),
    it.description ? el("p", { class: "muted small", text: it.description }) : null,
    el("div", { class: "mission-line" }, [
      el("div", { class: "bar", role: "progressbar", "aria-valuemin": "0", "aria-valuemax": "100", "aria-valuenow": String(Math.round(pct)), "aria-label": it.name }, [el("i", { style: `width:${pct.toFixed(1)}%` })]),
      el("span", { class: "muted small", "data-progress": "", text: (meta.length ? meta.join(sep) + sep : "") + progress }),
    ]),
    notes.length ? el("p", { class: "muted small", "data-task-note": "", text: notes.join(sep) }) : null,
    child ? null : foot,
  ]);
  if ((it.children || []).length) {
    const steps = el("div", { class: "mission-steps" }, it.children.map((c) => missionCard(c, true)));
    node.append(steps);
  }
  return node;
}

function groupCard(g) {
  const head = el("div", { class: "panel-head" }, [
    el("h2", { text: g.name }),
    el("span", { class: "muted small", "data-group-progress": "", text: t("groupProgress", g.done, g.total) }),
  ]);
  const body = el("div", { class: "rewards" }, (g.tasks || []).filter((x) => x.state !== "claimed").map((x) => missionCard(x, false)));
  const foot = el("div", { class: "group-foot" });
  if (g.finished) foot.append(el("p", { class: "muted small", text: t("groupFinished") }));
  else if (g.canRequest) {
    const b = el("button", { class: "btn btn-primary", type: "button", "data-group-next": g.key, text: t("groupNext") });
    b.addEventListener("click", async () => {
      b.disabled = true;
      try {
        await rpc("economy_task_group_next", { key: g.key, lang: i18n.lang });
      } catch (err) {
        toast(errorText(err), "bad");
      }
      loadTasks();
    });
    foot.append(b);
  } else if (g.nextAt) foot.append(el("p", { class: "muted small", text: t("groupNextAt", missionTime(g.nextAt)) }));
  return el("section", { class: "panel mission-group", "data-group": g.key }, [head, g.description ? el("p", { class: "muted small", text: g.description }) : null, body, foot.childNodes.length ? foot : null]);
}

let tasksSeq = 0;
async function loadTasks() {
  const box = $("#tasks-list");
  const my = ++tasksSeq;
  if (!box.childNodes.length) box.append(el("div", { class: "center-block" }, [el("div", { class: "spinner" })]));
  try {
    const res = await rpc("economy_tasks_list", { lang: i18n.lang });
    if (my !== tasksSeq) return;
    box.textContent = "";
    const items = res.items || [];
    const groups = res.groups || [];
    for (const g of groups) box.append(groupCard(g));
    if (items.length) box.append(el("div", { class: "rewards" }, items.map((x) => missionCard(x, false))));
    if (!items.length && !groups.length) box.append(el("p", { class: "muted center", text: t("tasksEmpty") }));
    const dot = $("#tasks-dot");
    dot.hidden = !res.ready;
    dot.textContent = res.ready ? num(res.ready) : "";
  } catch (err) {
    if (my !== tasksSeq) return;
    box.textContent = "";
    box.append(el("p", { class: "muted center", text: errorText(err) }));
  }
}

function missionsShow(name) {
  if (name === "home") refreshTasksDot();
  if (name === "tasks") {
    $("#tasks-list").textContent = "";
    loadTasks();
  }
}
