Object.assign(STRINGS.fa, {
  bannersLabel: "پیشنهادهای ویژه",
  bannerPrev: "بنر قبلی",
  bannerNext: "بنر بعدی",
  bannerGo: (n, total) => `بنر ${num(n)} از ${num(total)}`,
  bannerCta: { screen: "باز کن", product: "ببین", achievement: "ببین", url: "برو" },
  bannerGone: "این پیشنهاد دیگر در دسترس نیست.",
  bannerOpening: "در حال باز کردن…",
});

Object.assign(STRINGS.en, {
  bannersLabel: "Featured",
  bannerPrev: "Previous banner",
  bannerNext: "Next banner",
  bannerGo: (n, total) => `Banner ${num(n)} of ${num(total)}`,
  bannerCta: { screen: "Open", product: "View", achievement: "View", url: "Go" },
  bannerGone: "This offer is no longer available.",
  bannerOpening: "Opening…",
});

const BANNER_PLACEMENT = "menu_carousel";
const bnState = { items: [], placement: null, index: 0, timer: null, at: 0, lang: "", seen: new Set(), queue: new Set(), flush: null, paused: false, loading: false };

function bannerPayload(extra) {
  return Object.assign({ platform: "web", version: typeof gameVersion === "function" ? gameVersion() : GAME_VERSION, lang: i18n.lang }, extra || {});
}

async function loadMenuBanners(force) {
  if (!state.token || bnState.loading) return;
  if (!force && bnState.lang === i18n.lang && Date.now() - bnState.at < 60000) {
    paintBanners();
    return;
  }
  bnState.loading = true;
  try {
    const res = await rpc("publication_banners", bannerPayload({ placement: BANNER_PLACEMENT }));
    bnState.items = res.items || [];
    bnState.placement = res.placement || null;
    bnState.lang = i18n.lang;
    bnState.at = Date.now();
    if (bnState.index >= bnState.items.length) bnState.index = 0;
  } catch (err) {
    bnState.items = [];
  } finally {
    bnState.loading = false;
  }
  paintBanners();
}

function bannerSlide(b, i, total) {
  const cta = (b.action && b.action.type !== "none") ? (b.action.label || (t("bannerCta") || {})[b.action.type] || "") : "";
  const slide = el("button", {
    class: "bn-slide" + (b.action && b.action.type !== "none" ? " bn-tap" : ""),
    type: "button",
    "data-banner": b.key,
    "data-action": b.action ? b.action.type : "none",
    "aria-roledescription": "slide",
    "aria-label": t("bannerGo", i + 1, total) + ": " + b.title,
    style: b.bgColor ? "--bn-bg:" + b.bgColor : null,
    onclick: () => openBanner(b, slide),
  }, [
    b.imageUrl ? el("img", { class: "bn-img", src: b.imageUrl, alt: "", loading: i === 0 ? "eager" : "lazy", decoding: "async" }) : null,
    el("span", { class: "bn-text" }, [
      el("span", { class: "bn-title", text: b.title }),
      b.subtitle ? el("span", { class: "bn-sub", text: b.subtitle }) : null,
      cta ? el("span", { class: "bn-cta", text: cta }) : null,
    ]),
  ]);
  return slide;
}

function paintBanners() {
  const box = $("#banner-carousel");
  if (!box) return;
  const items = bnState.items;
  box.textContent = "";
  box.hidden = !items.length;
  stopBannerTimer();
  if (!items.length) return;
  const track = el("div", { class: "bn-track", "aria-live": bnState.paused ? "polite" : "off" });
  items.forEach((b, i) => track.append(bannerSlide(b, i, items.length)));
  box.setAttribute("aria-label", t("bannersLabel"));
  box.append(track);
  if (items.length > 1) {
    const dots = el("div", { class: "bn-dots", role: "group", "aria-label": t("bannersLabel") });
    items.forEach((b, i) => dots.append(el("button", { class: "bn-dot", type: "button", "data-dot": String(i), "aria-label": t("bannerGo", i + 1, items.length), onclick: () => goBanner(i, true) })));
    box.append(
      el("button", { class: "bn-nav bn-prev", type: "button", "aria-label": t("bannerPrev"), onclick: () => goBanner(bnState.index - 1, true) }, "‹"),
      el("button", { class: "bn-nav bn-next", type: "button", "aria-label": t("bannerNext"), onclick: () => goBanner(bnState.index + 1, true) }, "›"),
      dots
    );
  }
  goBanner(bnState.index, false);
  startBannerTimer();
}

function goBanner(i, byUser) {
  const items = bnState.items;
  if (!items.length) return;
  bnState.index = (i + items.length) % items.length;
  const box = $("#banner-carousel");
  const slides = box.querySelectorAll(".bn-slide");
  slides.forEach((s, k) => {
    const on = k === bnState.index;
    s.hidden = !on;
    s.setAttribute("aria-hidden", on ? "false" : "true");
    s.tabIndex = on ? 0 : -1;
  });
  box.querySelectorAll(".bn-dot").forEach((d, k) => d.setAttribute("aria-current", k === bnState.index ? "true" : "false"));
  box.dataset.index = String(bnState.index);
  noteBannerView(items[bnState.index]);
  if (byUser) {
    bnState.paused = true;
    stopBannerTimer();
  }
}

function bannerVisible() {
  const box = $("#banner-carousel");
  return box && !box.hidden && !$("#screen-home").hidden && document.visibilityState === "visible";
}

function noteBannerView(b) {
  if (!b || bnState.seen.has(b.id) || !bannerVisible()) return;
  bnState.seen.add(b.id);
  bnState.queue.add(b.id);
  clearTimeout(bnState.flush);
  bnState.flush = setTimeout(flushBannerViews, 800);
}

async function flushBannerViews() {
  const ids = [...bnState.queue].slice(0, 20);
  ids.forEach((id) => bnState.queue.delete(id));
  if (!ids.length) return;
  try {
    await rpc("publication_banner_view", bannerPayload({ bannerIds: ids }));
  } catch (err) {
    ids.forEach((id) => bnState.seen.delete(id));
  }
}

function stopBannerTimer() {
  clearInterval(bnState.timer);
  bnState.timer = null;
}

function startBannerTimer() {
  const sec = bnState.placement ? bnState.placement.rotateSeconds : 0;
  if (!sec || bnState.items.length < 2 || bnState.paused) return;
  if (window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  bnState.timer = setInterval(() => {
    if (!bannerVisible()) return;
    goBanner(bnState.index + 1, false);
  }, sec * 1000);
}

async function openBanner(b, slide) {
  if (!b.action || b.action.type === "none") return;
  slide.disabled = true;
  let res;
  try {
    res = await rpc("publication_banner_click", bannerPayload({ bannerId: b.id }));
  } catch (err) {
    slide.disabled = false;
    if (err.code === "BANNER_NOT_AVAILABLE") {
      toast(t("bannerGone"), "bad");
      loadMenuBanners(true);
    } else {
      toast(errorText(err), "bad");
    }
    return;
  }
  slide.disabled = false;
  runBannerAction(res.action || {});
}

function flashCard(sel) {
  let tries = 0;
  const tick = () => {
    const node = document.querySelector(sel);
    if (node) {
      node.classList.add("bn-flash");
      node.scrollIntoView({ block: "center", behavior: "smooth" });
      setTimeout(() => node.classList.remove("bn-flash"), 2400);
      return;
    }
    if (++tries < 40) setTimeout(tick, 150);
  };
  tick();
}

function runBannerAction(a) {
  if (a.type === "screen") {
    if (SCREENS.includes(a.value)) show(a.value);
    return;
  }
  if (a.type === "product") {
    state.bannerProduct = a.value;
    show("shop");
    return;
  }
  if (a.type === "achievement") {
    show("achievements");
    flashCard(`[data-achievement="${CSS.escape(a.value)}"]`);
    return;
  }
  if (a.type === "url" && /^https?:\/\//i.test(a.url || "")) {
    window.open(a.url, "_blank", "noopener,noreferrer");
  }
}

function bannerShopHook() {
  const key = state.bannerProduct;
  if (!key || !state.catalog) return;
  const p = (state.catalog.products || []).find((x) => x.key === key);
  state.bannerProduct = null;
  if (!p) {
    toast(t("bannerGone"), "bad");
    return;
  }
  if (state.shopTab !== p.categoryKey) {
    state.shopTab = p.categoryKey;
    renderShop();
  }
  flashCard(`#products [data-key="${CSS.escape(key)}"]`);
}

function bannersShow(name) {
  if (name === "home") loadMenuBanners(false);
  else stopBannerTimer();
}

document.addEventListener("visibilitychange", () => {
  if (document.visibilityState === "visible" && !$("#screen-home").hidden) {
    noteBannerView(bnState.items[bnState.index]);
  }
});
