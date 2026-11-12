Object.assign(STRINGS.fa, {
  sponsorBy: (name) => `با حمایت ${name}`,
  sponsorAbout: "دربارهٔ حامی",
  sponsorLinks: "پیوندهای حامی",
});

Object.assign(STRINGS.en, {
  sponsorBy: (name) => `Sponsored by ${name}`,
  sponsorAbout: "About the sponsor",
  sponsorLinks: "Sponsor links",
});

const SPONSOR_PLACEMENT = "luck_sponsor";
const spState = { lang: "", items: [], placement: null, sponsors: [], seen: new Set() };

function sponsorPayload(extra) {
  return Object.assign({ platform: "web", version: typeof gameVersion === "function" ? gameVersion() : GAME_VERSION, lang: i18n.lang }, extra || {});
}

async function loadSponsorBox() {
  const box = $("#sponsor-box");
  if (!box || !state.token) return;
  try {
    const [res, list] = await Promise.all([
      rpc("publication_banners", sponsorPayload({ placement: SPONSOR_PLACEMENT })),
      rpc("publication_sponsors", { lang: i18n.lang }),
    ]);
    spState.items = res.items || [];
    spState.placement = res.placement || null;
    spState.sponsors = list.items || [];
    spState.lang = i18n.lang;
  } catch (err) {
    spState.items = [];
  }
  paintSponsorBox();
}

function sponsorCredit(sp) {
  if (!sp) return null;
  return el("span", { class: "sp-credit", "data-sponsor-credit": sp.key }, [
    sp.logoUrl ? el("img", { class: "sp-logo", src: sp.logoUrl, alt: "", width: "28", height: "28" }) : null,
    el("span", { text: t("sponsorBy", sp.name) }),
  ]);
}

function paintSponsorBox() {
  const box = $("#sponsor-box");
  if (!box) return;
  box.textContent = "";
  const b = spState.items[0];
  box.hidden = !b;
  if (!b) return;
  const sp = b.sponsor || (spState.placement && spState.placement.sponsor);
  const full = spState.sponsors.find((x) => sp && x.id === sp.id);
  const card = el("button", {
    class: "sp-card bn-tap",
    type: "button",
    "data-banner": b.key,
    "data-action": b.action ? b.action.type : "none",
    style: b.bgColor ? "--bn-bg:" + b.bgColor : null,
    "aria-label": b.title + (sp ? " — " + t("sponsorBy", sp.name) : ""),
    onclick: () => openSponsorBanner(b, card),
  }, [
    b.imageUrl ? el("img", { class: "bn-img", src: b.imageUrl, alt: "", decoding: "async" }) : null,
    el("span", { class: "bn-text" }, [
      sponsorCredit(sp),
      el("span", { class: "bn-title", text: b.title }),
      b.subtitle ? el("span", { class: "bn-sub", text: b.subtitle }) : null,
      b.action && b.action.type !== "none" ? el("span", { class: "bn-cta", text: b.action.label || t("sponsorAbout") }) : null,
    ]),
  ]);
  box.append(card);
  if (full && (full.links || []).length) {
    const links = el("div", { class: "sp-links", role: "group", "aria-label": t("sponsorLinks") });
    full.links.forEach((l) => links.append(el("button", { class: "btn btn-ghost sp-link", type: "button", "data-sponsor-link": l.type, onclick: () => openSponsorLink(full, l.type) }, l.label || l.type)));
    box.append(links);
  }
  if (!spState.seen.has(b.id)) {
    spState.seen.add(b.id);
    rpc("publication_banner_view", sponsorPayload({ bannerIds: [b.id] })).catch(() => spState.seen.delete(b.id));
  }
}

async function openSponsorBanner(b, card) {
  card.disabled = true;
  try {
    const res = await rpc("publication_banner_click", sponsorPayload({ bannerId: b.id }));
    const a = res.action || {};
    if (/^https?:\/\//i.test(a.url || "")) window.open(a.url, "_blank", "noopener,noreferrer");
  } catch (err) {
    if (err.code === "BANNER_NOT_AVAILABLE") {
      toast(t("bannerGone"), "bad");
      loadSponsorBox();
    } else {
      toast(errorText(err), "bad");
    }
  } finally {
    card.disabled = false;
  }
}

async function openSponsorLink(sp, type) {
  try {
    const res = await rpc("publication_sponsor_click", { sponsorId: sp.id, link: type, lang: i18n.lang });
    if (/^https?:\/\//i.test(res.url || "")) window.open(res.url, "_blank", "noopener,noreferrer");
  } catch (err) {
    toast(errorText(err), "bad");
  }
}

function sponsorsShow(name) {
  if (name === "chance") loadSponsorBox();
}
