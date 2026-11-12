Object.assign(STRINGS.fa, {
  achievements: "دستاوردها",
  achievementsHint: "نشان‌هایی که با بازی کردن باز می‌کنی. بعضی‌ها چند رده دارند و هر رده جایزهٔ خودش را.",
  achSummary: (n, total, p) => `${num(n)} از ${num(total)} دستاورد · ${num(p)} امتیاز`,
  achLocked: "قفل",
  achUnlocked: "باز شد",
  achCompleted: "همهٔ رده‌ها",
  achSecret: "دستاورد پنهان",
  achSecretName: "؟؟؟",
  achTier: (n, max) => `ردهٔ ${num(n)} از ${num(max)}`,
  achNext: (name, n) => `تا ${name}: ${n}`,
  achPoints: (n) => `${num(n)} امتیاز`,
  achClaim: "جایزه را بگیر",
  achClaimed: "گرفته شد",
  achClaimedToast: (x) => `جایزهٔ دستاورد: ${x}`,
  achAt: (d) => `باز شده در ${d}`,
  achEmpty: "هنوز دستاوردی نیست.",
  achNew: (n) => (n === 1 ? "یک دستاورد تازه باز شد!" : `${num(n)} دستاورد تازه باز شد!`),
  achTierUp: "به ردهٔ تازه‌ای از یک دستاورد رسیدی!",
  achTierName: (n) => `ردهٔ ${num(n)}`,
});

Object.assign(STRINGS.en, {
  achievements: "Achievements",
  achievementsHint: "Badges you unlock by playing. Some have several tiers, each with its own prize.",
  achSummary: (n, total, p) => `${num(n)} of ${num(total)} achievements · ${num(p)} points`,
  achLocked: "Locked",
  achUnlocked: "Unlocked",
  achCompleted: "All tiers",
  achSecret: "Secret achievement",
  achSecretName: "???",
  achTier: (n, max) => `Tier ${num(n)} of ${num(max)}`,
  achNext: (name, n) => `To ${name}: ${n}`,
  achPoints: (n) => `${num(n)} points`,
  achClaim: "Collect prize",
  achClaimed: "Collected",
  achClaimedToast: (x) => `Achievement prize: ${x}`,
  achAt: (d) => `Unlocked ${d}`,
  achEmpty: "No achievements yet.",
  achNew: (n) => (n === 1 ? "A new achievement unlocked!" : `${num(n)} new achievements unlocked!`),
  achTierUp: "You reached a new achievement tier!",
  achTierName: (n) => `Tier ${num(n)}`,
});

Object.assign(STRINGS.fa.errors, {
  ACHIEVEMENT_LOCKED: "این دستاورد هنوز باز نشده.",
  TIER_NOT_REACHED: "هنوز به این رده نرسیده‌ای.",
  NOTHING_TO_CLAIM: "جایزه‌ای برای گرفتن نمانده.",
  ACHIEVEMENT_NOT_FOUND: "این دستاورد پیدا نشد.",
});

Object.assign(STRINGS.en.errors, {
  ACHIEVEMENT_LOCKED: "This achievement is not unlocked yet.",
  TIER_NOT_REACHED: "You have not reached this tier yet.",
  NOTHING_TO_CLAIM: "There is no prize left to collect.",
  ACHIEVEMENT_NOT_FOUND: "This achievement was not found.",
});

if (!SCREENS.includes("achievements")) SCREENS.push("achievements");

const ACH_SVG = {
  medal: '<svg viewBox="0 0 256 256" width="34" height="34" fill="currentColor" aria-hidden="true"><path d="M176,96a48,48,0,1,1-48-48A48,48,0,0,1,176,96Z" opacity="0.25"/><path d="M216,96A88,88,0,1,0,72,163.83V240a8,8,0,0,0,11.58,7.16L128,225l44.43,22.21A8.07,8.07,0,0,0,176,248a8,8,0,0,0,8-8V163.83A87.85,87.85,0,0,0,216,96ZM56,96a72,72,0,1,1,72,72A72.08,72.08,0,0,1,56,96ZM168,227.06l-36.43-18.21a8,8,0,0,0-7.16,0L88,227.06V174.37a87.89,87.89,0,0,0,80,0ZM128,152A56,56,0,1,0,72,96,56.06,56.06,0,0,0,128,152Zm0-96A40,40,0,1,1,88,96,40,40,0,0,1,128,56Z"/></svg>',
  lock: '<svg viewBox="0 0 256 256" width="30" height="30" fill="currentColor" aria-hidden="true"><path d="M208,96V208a8,8,0,0,1-8,8H56a8,8,0,0,1-8-8V96a8,8,0,0,1,8-8H200A8,8,0,0,1,208,96Z" opacity="0.25"/><path d="M208,80H176V56a48,48,0,0,0-96,0V80H48A16,16,0,0,0,32,96V208a16,16,0,0,0,16,16H208a16,16,0,0,0,16-16V96A16,16,0,0,0,208,80ZM96,56a32,32,0,0,1,64,0V80H96ZM208,208H48V96H208V208Zm-68-56a12,12,0,1,1-12-12A12,12,0,0,1,140,152Z"/></svg>',
  secret: '<svg viewBox="0 0 256 256" width="30" height="30" fill="currentColor" aria-hidden="true"><path d="M224,128a96,96,0,1,1-96-96A96,96,0,0,1,224,128Z" opacity="0.25"/><path d="M140,180a12,12,0,1,1-12-12A12,12,0,0,1,140,180ZM128,72c-22.06,0-40,16.15-40,36v4a8,8,0,0,0,16,0v-4c0-11,10.77-20,24-20s24,9,24,20-10.77,20-24,20a8,8,0,0,0-8,8v8a8,8,0,0,0,16,0v-.72c18.24-3.35,32-17.9,32-35.28C168,88.15,150.06,72,128,72Zm104,56A104,104,0,1,1,128,24,104.11,104.11,0,0,1,232,128Zm-16,0a88,88,0,1,0-88,88A88.1,88.1,0,0,0,216,128Z"/></svg>',
};

function achBadgeEl(it) {
  const locked = !it.unlocked;
  const box = el("div", { class: "ach-badge ach-" + (it.rarity || "common") + (locked ? " ach-off" : ""), "aria-hidden": "true" });
  const src = locked && it.lockedIconUrl ? it.lockedIconUrl : it.iconUrl;
  if (src && !it.hidden) box.append(el("img", { src, alt: "" }));
  else box.innerHTML = it.state === "hidden" ? ACH_SVG.secret : locked ? ACH_SVG.lock : ACH_SVG.medal;
  return box;
}

function achievementNotes(ev) {
  const list = (ev && ev.achievements) || [];
  const fresh = list.filter((x) => x.unlocked).length;
  if (fresh) return t("achNew", fresh);
  if (list.length) return t("achTierUp");
  return null;
}

async function refreshAchDot() {
  if (!state.token) return;
  try {
    const res = await rpc("economy_achievements", { lang: i18n.lang });
    const dot = $("#achievements-dot");
    dot.hidden = !res.ready;
    dot.textContent = res.ready ? num(res.ready) : "";
  } catch (e) {}
}

async function claimAchievement(it, btn) {
  btn.disabled = true;
  try {
    const r = await rpc("economy_achievement_claim", { key: it.key, lang: i18n.lang });
    const got = (r.claims || []).map((c) => rewardText(c.rewards) || (c.prize && c.prize.name) || "").filter(Boolean).join(i18n.dir === "rtl" ? "، " : ", ");
    toast(t("achClaimedToast", got), "ok");
    await refreshBalances().catch(() => {});
  } catch (err) {
    toast(errorText(err), "bad");
  }
  loadAchievements();
}

function achievementCard(it) {
  if (it.state === "hidden") {
    return el("article", { class: "reward ach-card ach-hidden", "data-achievement": "hidden", "data-state": "hidden" }, [
      el("div", { class: "ach-top" }, [achBadgeEl(it), el("div", { class: "ach-body" }, [
        el("div", { class: "reward-head" }, [el("b", { text: t("achSecretName") }), el("span", { class: "chip chip-wait", text: t("achSecret") })]),
        it.hint ? el("p", { class: "muted small", "data-ach-hint": "", text: it.hint }) : null,
        el("span", { class: "chip inv-rarity inv-" + it.rarity, text: (t("rarity") || {})[it.rarity] || it.rarity }),
      ])]),
    ]);
  }
  const sep = i18n.dir === "rtl" ? "، " : " · ";
  const chips = [el("span", { class: "chip inv-rarity inv-" + it.rarity, text: (t("rarity") || {})[it.rarity] || it.rarity }), el("span", { class: "chip", text: t("achPoints", it.points || 0) })];
  const stateText = it.state === "completed" ? t("achCompleted") : it.unlocked ? t("achUnlocked") : t("achLocked");
  const stateChip = el("span", { class: it.unlocked ? "chip chip-gold" : "chip chip-wait", "data-ach-state": "", text: stateText });
  const lines = [];
  if (it.maxTier > 1) lines.push(el("p", { class: "small ach-tier", "data-ach-tier": "", text: t("achTier", it.tier || 0, it.maxTier) + (it.tierName ? sep + it.tierName : "") }));
  if (it.source === "event" && it.next) {
    const pct = Math.max(0, Math.min(100, (Number(it.progress) || 0) * 100));
    const target = it.next.name || t("achTierName", it.next.tier);
    lines.push(el("div", { class: "mission-line" }, [
      el("div", { class: "bar", role: "progressbar", "aria-valuemin": "0", "aria-valuemax": "100", "aria-valuenow": String(Math.round(pct)), "aria-label": it.name }, [el("i", { style: `width:${pct.toFixed(1)}%` })]),
      el("span", { class: "muted small", "data-progress": "", text: t("achNext", target, t("progressOf", Math.min(it.value || 0, it.next.threshold), it.next.threshold)) }),
    ]));
  }
  if (!it.unlocked && it.hint) lines.push(el("p", { class: "muted small", "data-ach-hint": "", text: it.hint }));
  const tiers = (it.tiers || []).filter((x) => x.prize);
  const tierRow = tiers.length > 0 ? el("div", { class: "ach-tiers" }, tiers.map((x) => el("span", { class: "chip ach-tier-chip ach-t-" + x.status, "data-tier": String(x.tier), "data-tier-status": x.status, text: (x.name || t("achTierName", x.tier)) + ": " + (x.prize.name || "") }))) : null;
  const foot = el("div", { class: "reward-foot" });
  if (it.unlockedAt) foot.append(el("span", { class: "muted small", text: t("achAt", fmtDay(it.unlockedAt)) }));
  if ((it.claimable || []).length) {
    const b = el("button", { class: "btn btn-primary", type: "button", "data-ach-claim": it.key, text: t("achClaim") });
    b.addEventListener("click", () => claimAchievement(it, b));
    foot.append(b);
  }
  return el("article", { class: "reward ach-card" + ((it.claimable || []).length ? " ready" : "") + (it.unlocked ? " ach-on" : ""), "data-achievement": it.key, "data-state": it.state, "data-tier": String(it.tier || 0) }, [
    el("div", { class: "ach-top" }, [achBadgeEl(it), el("div", { class: "ach-body" }, [
      el("div", { class: "reward-head" }, [el("b", { text: it.name }), stateChip]),
      it.description ? el("p", { class: "muted small", text: it.description }) : null,
      el("div", { class: "inv-chips" }, chips),
    ])]),
    ...lines,
    tierRow,
    foot,
  ]);
}

function fmtDay(iso) {
  try {
    return new Intl.DateTimeFormat(i18n.locale, { year: "numeric", month: "short", day: "numeric" }).format(new Date(iso));
  } catch (e) {
    return iso;
  }
}

let achSeq = 0;
async function loadAchievements() {
  const box = $("#achievements-list");
  const my = ++achSeq;
  if (!box.childNodes.length) box.append(el("div", { class: "center-block" }, [el("div", { class: "spinner" })]));
  try {
    const res = await rpc("economy_achievements", { lang: i18n.lang });
    if (my !== achSeq) return;
    box.textContent = "";
    $("#achievements-summary").textContent = t("achSummary", res.unlocked || 0, res.total || 0, res.points || 0);
    const dot = $("#achievements-dot");
    dot.hidden = !res.ready;
    dot.textContent = res.ready ? num(res.ready) : "";
    const items = res.items || [];
    if (!items.length) {
      box.append(el("p", { class: "muted center", text: t("achEmpty") }));
      return;
    }
    const sorted = items.slice().sort((a, b) => ((b.claimable || []).length - (a.claimable || []).length) || ((b.unlocked ? 1 : 0) - (a.unlocked ? 1 : 0)) || ((a.sort || 0) - (b.sort || 0)));
    box.append(el("div", { class: "ach-grid" }, sorted.map(achievementCard)));
  } catch (err) {
    if (my !== achSeq) return;
    box.textContent = "";
    box.append(el("p", { class: "muted center", text: errorText(err) }));
  }
}

function achievementsShow(name) {
  if (name === "home") refreshAchDot();
  if (name === "achievements") {
    $("#achievements-list").textContent = "";
    loadAchievements();
  }
}
