var ND_CAL_KEY = 'nakama.date.calendar';
var ND_ZONE_KEY = 'nakama.date.zone';
var ND_DEFAULT_ZONE = 'Asia/Tehran';
var ND_SETTINGS_PATH = '/v2/console/base/console-settings';
var ND_PREFS = ['auto', 'persian', 'gregory'];
var ndState = { zone: null, loaded: false, loading: null, listeners: [], meta: null };

function ndStore(kind) {
  try { return kind === 'session' ? window.sessionStorage : window.localStorage; } catch (e) { return null; }
}

function ndGet(key) {
  var s = ndStore('local');
  try { return s ? s.getItem(key) : null; } catch (e) { return null; }
}

function ndSet(key, value) {
  var s = ndStore('local');
  try { if (s) { if (value == null) s.removeItem(key); else s.setItem(key, value); } } catch (e) { return; }
}

function ndLang() {
  var l = typeof document !== 'undefined' && document.documentElement ? document.documentElement.getAttribute('lang') || '' : '';
  return /^fa/i.test(l) ? 'fa' : 'en';
}

function ndLocale(lang) {
  return (lang || ndLang()) === 'fa' ? 'fa-IR' : 'en-US';
}

function calendarPref() {
  var v = ndGet(ND_CAL_KEY);
  return ND_PREFS.indexOf(v) === -1 ? 'auto' : v;
}

function activeCalendar(lang) {
  var p = calendarPref();
  if (p !== 'auto') return p;
  return (lang || ndLang()) === 'fa' ? 'persian' : 'gregory';
}

function setCalendarPref(v, opts) {
  if (ND_PREFS.indexOf(v) === -1) v = 'auto';
  ndSet(ND_CAL_KEY, v === 'auto' ? null : v);
  ndEmit('calendar');
  if (!(opts && opts.reload === false) && typeof location !== 'undefined') location.reload();
}

function ndZone() {
  if (ndState.zone) return ndState.zone;
  var cached = ndGet(ND_ZONE_KEY);
  return validZone(cached) ? cached : ND_DEFAULT_ZONE;
}

function setZone(z, opts) {
  if (!validZone(z)) return false;
  var changed = z !== ndZone();
  ndState.zone = z;
  ndSet(ND_ZONE_KEY, z);
  if (changed) ndEmit('zone');
  if (changed && opts && opts.reload && typeof location !== 'undefined') location.reload();
  return true;
}

function ndEmit(kind) {
  ndState.listeners.slice().forEach(function (fn) { try { fn(kind); } catch (e) { return; } });
}

function ndOnChange(fn) {
  ndState.listeners.push(fn);
  return function () { ndState.listeners = ndState.listeners.filter(function (x) { return x !== fn; }); };
}

function ndToken() {
  var stores = [ndStore('session'), ndStore('local')];
  for (var i = 0; i < stores.length; i++) {
    try {
      if (!stores[i]) continue;
      var raw = stores[i].getItem('SESSION_TOKENS');
      if (!raw) continue;
      var j = JSON.parse(raw);
      if (j && j.accessToken) return j.accessToken;
    } catch (e) { continue; }
  }
  return null;
}

function loadSettings(force) {
  if (ndState.loading && !force) return ndState.loading;
  var tk = ndToken();
  if (!tk || typeof fetch !== 'function') return Promise.resolve({ timezone: ndZone() });
  ndState.loading = fetch(ND_SETTINGS_PATH, { headers: { Accept: 'application/json', Authorization: 'Bearer ' + tk }, credentials: 'same-origin' }).then(function (res) {
    if (!res.ok) throw new Error('settings ' + res.status);
    return res.json();
  }).then(function (res) {
    ndState.loaded = true;
    ndState.meta = res || null;
    if (res && res.timezone) setZone(res.timezone);
    return res;
  }, function () {
    ndState.loading = null;
    return { timezone: ndZone() };
  });
  return ndState.loading;
}

function saveSettings(body) {
  var tk = ndToken();
  return fetch(ND_SETTINGS_PATH, { method: 'PUT', headers: { Accept: 'application/json', 'Content-Type': 'application/json', Authorization: 'Bearer ' + (tk || '') }, credentials: 'same-origin', body: JSON.stringify(body || {}) }).then(function (res) {
    return res.text().then(function (text) {
      var data = null;
      try { data = text ? JSON.parse(text) : null; } catch (e) { data = null; }
      if (!res.ok) {
        var err = new Error((data && data.message) || 'save failed');
        err.status = res.status;
        err.code = data && (typeof data.error === 'string' ? data.error : data.code);
        err.field = data && data.field;
        throw err;
      }
      ndState.loaded = true;
      ndState.meta = data;
      if (data && data.timezone) setZone(data.timezone);
      return data;
    });
  });
}

function ndWatchLogin() {
  if (typeof window === 'undefined' || !window.addEventListener) return;
  var check = function () { if (!ndState.loaded && ndToken()) loadSettings(); };
  window.addEventListener('hashchange', check);
  window.addEventListener('focus', check);
  setTimeout(check, 0);
}
