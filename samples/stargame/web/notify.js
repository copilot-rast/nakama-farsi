Object.assign(STRINGS.fa, {
  notify: "اعلان‌ها",
  notifyHint: "این‌جا می‌گویی چه اعلان‌هایی بگیری. اعلان‌ها هم روی دستگاه می‌آیند و هم در صندوق پیام.",
  notifyDevice: "این دستگاه",
  notifyDeviceOn: "اعلان روی این دستگاه روشن است.",
  notifyDeviceOff: "اعلان روی این دستگاه خاموش است.",
  notifyDeviceSandbox: "نمونهٔ وب: پیام واقعی فرستاده نمی‌شود و فقط در سرور ثبت می‌شود.",
  notifyEnable: "روشن کردن اعلان این دستگاه",
  notifyDisable: "خاموش کردن اعلان این دستگاه",
  notifyKinds: "چه اعلان‌هایی بگیرم؟",
  notifyLocked: "همیشه روشن",
  notifySaved: "ذخیره شد.",
  notifyOn: (name) => `${name}: روشن`,
  notifyOff: (name) => `${name}: خاموش`,
});

Object.assign(STRINGS.en, {
  notify: "Notifications",
  notifyHint: "Choose which notifications you get. They come to this device and to your inbox.",
  notifyDevice: "This device",
  notifyDeviceOn: "Notifications are on for this device.",
  notifyDeviceOff: "Notifications are off for this device.",
  notifyDeviceSandbox: "Web sample: no real message is sent; the server only logs it.",
  notifyEnable: "Turn on notifications for this device",
  notifyDisable: "Turn off notifications for this device",
  notifyKinds: "Which notifications do you want?",
  notifyLocked: "Always on",
  notifySaved: "Saved.",
  notifyOn: (name) => `${name}: on`,
  notifyOff: (name) => `${name}: off`,
});

const PUSH_KEY = "stargame.pushToken";
const PUSH_OFF = "stargame.pushOff";
const ntState = { registered: false, loading: false };

function pushToken() {
  let tok = localStorage.getItem(PUSH_KEY);
  if (!tok) {
    const bytes = new Uint8Array(16);
    crypto.getRandomValues(bytes);
    tok = "sandbox-web-" + Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
    localStorage.setItem(PUSH_KEY, tok);
  }
  return tok;
}

async function registerPush(force) {
  if (!state.token) return;
  if (!force && localStorage.getItem(PUSH_OFF) === "1") return;
  try {
    await rpc("publication_push_register", { token: pushToken(), platform: "web", appVersion: typeof gameVersion === "function" ? gameVersion() : GAME_VERSION, deviceId: localStorage.getItem("stargame.device") || "" });
    localStorage.removeItem(PUSH_OFF);
    ntState.registered = true;
  } catch (err) {
    ntState.registered = false;
  }
}

async function unregisterPush() {
  try {
    await rpc("publication_push_unregister", { token: pushToken() });
  } catch (err) {
    toast(errorText(err), "bad");
    return;
  }
  localStorage.setItem(PUSH_OFF, "1");
  ntState.registered = false;
}

function paintDevice(box, devices) {
  box.textContent = "";
  const mine = (devices || []).find((d) => d.active && d.platform === "web");
  ntState.registered = !!mine;
  box.append(
    el("h2", { text: t("notifyDevice") }),
    el("p", { class: "nt-device-state", "data-push": mine ? "on" : "off", text: mine ? t("notifyDeviceOn") : t("notifyDeviceOff") }),
    mine ? el("p", { class: "muted small nt-token", dir: "ltr", "data-push-token": mine.token, text: mine.token }) : null,
    el("p", { class: "muted small", text: t("notifyDeviceSandbox") }),
    el("button", { class: "btn " + (mine ? "btn-ghost" : "btn-primary"), type: "button", id: "btn-push-toggle", onclick: async (e) => {
      e.currentTarget.disabled = true;
      if (mine) await unregisterPush();
      else await registerPush(true);
      loadNotify();
    } }, mine ? t("notifyDisable") : t("notifyEnable"))
  );
}

function prefRow(p) {
  const id = "pref-" + p.key;
  const input = el("input", { type: "checkbox", id, role: "switch", "data-pref": p.key, checked: p.enabled ? true : null, disabled: p.canOptOut ? null : true });
  input.addEventListener("change", async () => {
    input.disabled = true;
    try {
      const r = await rpc("publication_notify_prefs_set", { prefs: { [p.key]: input.checked }, lang: i18n.lang });
      const now = (r.items || []).find((x) => x.key === p.key);
      input.checked = !!(now && now.enabled);
      toast(input.checked ? t("notifyOn", p.name) : t("notifyOff", p.name), "ok");
    } catch (err) {
      input.checked = !input.checked;
      toast(errorText(err), "bad");
    } finally {
      input.disabled = !p.canOptOut;
    }
  });
  return el("label", { class: "nt-pref", for: id }, [
    el("span", { class: "nt-pref-text" }, [el("b", { text: p.name }), p.description ? el("span", { class: "muted small", text: p.description }) : null]),
    p.canOptOut ? input : el("span", { class: "chip", text: t("notifyLocked") }),
  ]);
}

async function loadNotify() {
  const dev = $("#notify-device");
  const list = $("#notify-prefs");
  if (!dev || ntState.loading) return;
  ntState.loading = true;
  list.textContent = "";
  list.append(...skeletonList(3, "skel-inbox"));
  try {
    const [prefs, devices] = await Promise.all([rpc("publication_notify_prefs", { lang: i18n.lang }), rpc("publication_push_devices", {})]);
    paintDevice(dev, devices.items);
    list.textContent = "";
    for (const p of prefs.items || []) list.append(prefRow(p));
  } catch (err) {
    list.textContent = "";
    list.append(el("div", { class: "empty", text: errorText(err) }));
  } finally {
    ntState.loading = false;
  }
}

function markOpened(id) {
  if (!id) return;
  rpc("publication_notification_opened", { notificationId: id }).then((r) => setUnread(r.unread)).catch(() => {});
}

function runNoteAction(a) {
  if (!a || !a.type) return;
  if (a.type === "screen" || a.type === "goto") {
    if (SCREENS.includes(a.value)) show(a.value);
    else navigate(a.value);
    return;
  }
  if (a.type === "product" || a.type === "achievement") {
    runBannerAction(a);
    return;
  }
  if (a.type === "url" && /^https?:\/\//i.test(a.value || a.url || "")) window.open(a.value || a.url, "_blank", "noopener,noreferrer");
}

function notifyShow(name) {
  if (name === "notify") loadNotify();
}
