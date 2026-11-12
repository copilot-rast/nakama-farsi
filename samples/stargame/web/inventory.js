Object.assign(STRINGS.fa, {
  inventory: "انبار",
  inventoryHint: "سبدها و چیزهایی که داری. از هر نوع یکی را برای بازی انتخاب کن.",
  invEmpty: "انبار خالی است.",
  invOwned: "مال توست",
  invEquipped: "در بازی",
  invEquip: "انتخاب",
  invUnequip: "کنار بگذار",
  invBuy: (n) => `خرید با ${num(n)} سکه`,
  invUnlock: "باز کن",
  invLocked: (n) => `از سطح ${num(n)}`,
  invNeeds: (name) => `اول ${name} را بگیر`,
  invInShop: "در فروشگاه",
  invPrizeOnly: "فقط جایزه",
  invEquippedToast: (name) => `${name} انتخاب شد`,
  invBoughtToast: (name) => `${name} مال تو شد`,
  invUnlockedToast: (name) => `${name} باز شد`,
  invCount: (n, total) => `${num(n)} از ${num(total)}`,
  invConfirm: (name, n) => `${name} را با ${num(n)} سکه می‌خری؟`,
  rarity: { common: "معمولی", uncommon: "کمی کمیاب", rare: "کمیاب", epic: "حماسی", legendary: "افسانه‌ای" },
  fromLevel: (n) => `از سطح ${num(n)}`,
  oneTime: "یک بار",
});

Object.assign(STRINGS.en, {
  inventory: "Inventory",
  inventoryHint: "Baskets and things you own. Pick one of each type for the game.",
  invEmpty: "Your inventory is empty.",
  invOwned: "Yours",
  invEquipped: "In use",
  invEquip: "Use",
  invUnequip: "Put away",
  invBuy: (n) => `Buy for ${num(n)} coins`,
  invUnlock: "Unlock",
  invLocked: (n) => `From level ${num(n)}`,
  invNeeds: (name) => `Get ${name} first`,
  invInShop: "In the shop",
  invPrizeOnly: "Prize only",
  invEquippedToast: (name) => `${name} is in use`,
  invBoughtToast: (name) => `${name} is yours`,
  invUnlockedToast: (name) => `${name} unlocked`,
  invCount: (n, total) => `${num(n)} of ${num(total)}`,
  invConfirm: (name, n) => `Buy ${name} for ${num(n)} coins?`,
  rarity: { common: "Common", uncommon: "Uncommon", rare: "Rare", epic: "Epic", legendary: "Legendary" },
  fromLevel: (n) => `From level ${num(n)}`,
  oneTime: "Once",
});

Object.assign(STRINGS.fa.errors, {
  ALREADY_OWNED: "این را از قبل داری.",
  NOT_OWNED: "این را هنوز نداری.",
  NOT_FOR_SALE: "این فروشی نیست.",
  REQUIRES_ITEM: "اول باید آیتم پیش‌نیاز را داشته باشی.",
  LEVEL_TOO_LOW: "سطحت هنوز کافی نیست.",
  LEVEL_TOO_HIGH: "این بسته برای سطح‌های پایین‌تر است.",
  PRODUCT_ALREADY_BOUGHT: "این بسته را قبلاً خریده‌ای.",
  PRIZE_ALREADY_OWNED: "همهٔ چیزهای این بسته را داری.",
});

Object.assign(STRINGS.en.errors, {
  ALREADY_OWNED: "You already own this.",
  NOT_OWNED: "You don't own this yet.",
  NOT_FOR_SALE: "This is not for sale.",
  REQUIRES_ITEM: "You need the required item first.",
  LEVEL_TOO_LOW: "Your level is not high enough yet.",
  LEVEL_TOO_HIGH: "This pack is for lower levels.",
  PRODUCT_ALREADY_BOUGHT: "You already bought this pack.",
  PRIZE_ALREADY_OWNED: "You already own everything in this pack.",
});

if (!SCREENS.includes("inventory")) SCREENS.push("inventory");

function itemArt(it) {
  const art = it && it.metadata && it.metadata.art;
  return art && SKINS[art] ? art : null;
}

function itemName(key) {
  const it = ((state.inv && state.inv.items) || []).find((x) => x.key === key);
  return it ? it.name : key;
}

function productItemArt(p) {
  for (const key of p.items || []) {
    const it = ((state.inv && state.inv.items) || []).find((x) => x.key === key);
    if (itemArt(it)) return itemArt(it);
  }
  return null;
}

function syncInventorySkin() {
  const items = (state.inv && state.inv.items) || [];
  state.invArts = items.filter((x) => x.owned && itemArt(x)).map(itemArt);
  const skin = items.find((x) => x.owned && x.equipped && x.type === "skin" && itemArt(x));
  if (skin) localStorage.setItem(LS.skin, itemArt(skin));
}

async function loadInvArts() {
  if (!state.token) return null;
  try {
    state.inv = await rpc("economy_inventory", { lang: i18n.lang });
    syncInventorySkin();
  } catch (e) {}
  return state.inv;
}

async function itemAction(it, rpcName, payload, okText, btn) {
  if (btn) btn.disabled = true;
  try {
    await rpc(rpcName, Object.assign({ key: it.key, lang: i18n.lang }, payload || {}));
    if (rpcName === "economy_item_unequip" && it.type === "skin" && itemArt(it)) localStorage.setItem(LS.skin, "default");
    if (okText) toast(okText, "ok");
    await refreshBalances().catch(() => {});
  } catch (err) {
    toast(errorText(err), "bad");
  }
  await loadInventory();
}

function invRequestId() {
  const b = new Uint8Array(8);
  crypto.getRandomValues(b);
  return "inv-" + Array.from(b, (x) => x.toString(16).padStart(2, "0")).join("");
}

function inventoryCard(it, byKey) {
  const art = itemArt(it);
  const pic = el("div", { class: "product-art inv-art" }, art ? skinPreview(art, 72, 36) : el("span", { class: "inv-frame", "aria-hidden": "true" }));
  const chips = [el("span", { class: "chip inv-rarity inv-" + it.rarity, text: (t("rarity") || {})[it.rarity] || it.rarity })];
  if (it.equipped) chips.push(el("span", { class: "chip chip-gold", "data-inv-equipped": "", text: t("invEquipped") }));
  else if (it.owned) chips.push(el("span", { class: "chip chip-new", text: t("invOwned") }));
  const actions = el("div", { class: "methods" });
  if (it.owned) {
    if (it.equipped) {
      if (!it.isStarter) actions.append(el("button", { class: "btn", type: "button", "data-inv-unequip": it.key, text: t("invUnequip"), onclick: (e) => itemAction(it, "economy_item_unequip", null, null, e.currentTarget) }));
    } else {
      actions.append(el("button", { class: "btn btn-primary", type: "button", "data-inv-equip": it.key, text: t("invEquip"), onclick: (e) => itemAction(it, "economy_item_equip", null, t("invEquippedToast", it.name), e.currentTarget) }));
    }
  } else if (it.canBuy) {
    actions.append(el("button", { class: "btn btn-primary", type: "button", "data-inv-buy": it.key, text: t("invBuy", it.price.amount), onclick: async (e) => {
      const btn = e.currentTarget;
      const ok = await confirmDialog(t("buyWithCoins"), t("invConfirm", it.name, it.price.amount), t("buy"));
      if (ok) itemAction(it, "economy_item_buy", { requestId: invRequestId() }, t("invBoughtToast", it.name), btn);
    } }));
  } else if (it.canUnlock) {
    actions.append(el("button", { class: "btn btn-primary", type: "button", "data-inv-unlock": it.key, text: t("invUnlock"), onclick: (e) => itemAction(it, "economy_item_unlock", null, t("invUnlockedToast", it.name), e.currentTarget) }));
  } else {
    let note = t("invPrizeOnly");
    if (it.locked) note = t("invLocked", it.unlockLevel);
    else if (!it.requirementMet) note = t("invNeeds", byKey[it.requiresKey || it.parentKey] ? byKey[it.requiresKey || it.parentKey].name : it.requiresKey || it.parentKey);
    else if (!it.price && (state.catalog ? state.catalog.products.some((p) => (p.items || []).includes(it.key)) : true)) note = t("invInShop");
    actions.append(el("span", { class: "chip chip-wait", "data-inv-note": "", text: note }));
  }
  return el("article", { class: "product inv-item" + (it.equipped ? " inv-on" : ""), "data-item": it.key, "data-owned": it.owned ? "" : null }, [
    el("div", { class: "product-top" }, [pic, el("div", {}, [el("h3", { text: it.name }), el("div", { class: "inv-chips" }, chips), it.description ? el("p", { class: "muted small", text: it.description }) : null])]),
    actions,
  ]);
}

let invSeq = 0;
async function loadInventory() {
  const box = $("#inventory-list");
  const my = ++invSeq;
  if (!box.childNodes.length) box.append(el("div", { class: "center-block" }, [el("div", { class: "spinner" })]));
  try {
    const res = await rpc("economy_inventory", { lang: i18n.lang });
    if (my !== invSeq) return;
    state.inv = res;
    syncInventorySkin();
    if (!state.catalog) state.catalog = await rpc("economy_catalog", { lang: i18n.lang }).catch(() => null);
    box.textContent = "";
    const items = res.items || [];
    const byKey = Object.fromEntries(items.map((x) => [x.key, x]));
    $("#inventory-count").textContent = t("invCount", res.counts ? res.counts.owned : 0, items.length);
    if (!items.length) {
      box.append(el("p", { class: "muted center", text: t("invEmpty") }));
      return;
    }
    const groups = new Map();
    for (const it of items) {
      if (!groups.has(it.type)) groups.set(it.type, { name: it.typeName || it.type, items: [] });
      groups.get(it.type).items.push(it);
    }
    for (const [type, g] of groups) {
      box.append(el("section", { class: "inv-group", "data-inv-type": type }, [
        el("h2", { class: "inv-group-title", text: g.name }),
        el("div", { class: "products" }, g.items.map((x) => inventoryCard(x, byKey))),
      ]));
    }
  } catch (err) {
    if (my !== invSeq) return;
    box.textContent = "";
    box.append(el("p", { class: "muted center", text: errorText(err) }));
  }
}

function inventoryShow(name) {
  if (name === "home") loadInvArts();
  if (name === "inventory") {
    $("#inventory-list").textContent = "";
    loadInventory();
  }
}
