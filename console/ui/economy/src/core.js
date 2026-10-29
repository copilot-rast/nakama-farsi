var API_BASE = '/v2/console/economy';
var PUB_BASE = '/v2/console/publication';
var BASE_BASE = '/v2/console/base';
var CONSOLE_BASE = '/v2/console';
var PERSIAN = '۰۱۲۳۴۵۶۷۸۹';
var SVG_NS = 'http://www.w3.org/2000/svg';

function lang() {
  return /^fa/i.test(document.documentElement.getAttribute('lang') || '') ? 'fa' : 'en';
}

function locale() {
  return lang() === 'fa' ? 'fa-IR' : 'en-US';
}

function toLatinDigits(s) {
  return String(s == null ? '' : s)
    .replace(/[۰-۹]/g, function (d) { return String(d.charCodeAt(0) - 0x06F0); })
    .replace(/[٠-٩]/g, function (d) { return String(d.charCodeAt(0) - 0x0660); })
    .replace(/[٬،]/g, ',')
    .replace(/٫/g, '.');
}

function localDigits(s) {
  s = String(s == null ? '' : s);
  if (lang() !== 'fa') return s;
  return s.replace(/\d/g, function (d) { return PERSIAN[+d]; });
}

function fmtNum(n, opts) {
  if (n === null || n === undefined || n === '') return '';
  var v = Number(n);
  if (!isFinite(v)) return String(n);
  return new Intl.NumberFormat(locale(), opts || { maximumFractionDigits: 4 }).format(v);
}

function fmtDate(v, withTime) {
  if (!v) return '';
  var d = new Date(v);
  if (isNaN(d.getTime())) return String(v);
  var o = withTime === false ? { dateStyle: 'medium' } : { dateStyle: 'medium', timeStyle: 'short' };
  return new Intl.DateTimeFormat(locale(), o).format(d);
}

function t(key, vars) {
  var table = MESSAGES[lang()];
  var s = table[key];
  if (s === undefined) s = MESSAGES.en[key];
  if (s === undefined) return key;
  if (vars) {
    s = s.replace(/\{(\w+)\}/g, function (m, k) {
      var v = vars[k];
      if (typeof v === 'number') return fmtNum(v);
      return v === undefined ? m : String(v);
    });
  }
  return s;
}

function has(key) {
  return MESSAGES.en[key] !== undefined;
}

var LANGS = { items: [], active: [], codes: [], def: 'fa', loaded: false, promise: null, seq: 0, at: 0 };
var FALLBACK_LANGS = [
  { code: 'fa', nativeName: 'فارسی', name: { fa: 'فارسی', en: 'Persian' }, direction: 'rtl', isActive: true, isDefault: true, sort: 0 },
  { code: 'en', nativeName: 'English', name: { fa: 'انگلیسی', en: 'English' }, direction: 'ltr', isActive: true, isDefault: false, sort: 1 }
];

function setLanguages(items) {
  LANGS.items = items;
  LANGS.active = items.filter(function (l) { return l.isActive; });
  LANGS.codes = items.map(function (l) { return l.code; });
  var d = items.filter(function (l) { return l.isDefault; })[0] || LANGS.active[0] || items[0];
  LANGS.def = d ? d.code : 'fa';
  if (d) LANGS.active.sort(function (a, b) { return (b.code === d.code) - (a.code === d.code); });
}
setLanguages(FALLBACK_LANGS);

function loadLanguages(force) {
  if (LANGS.promise && !force) return LANGS.promise;
  var my = ++LANGS.seq;
  LANGS.promise = baseApi('GET', '/languages?limit=100').then(function (res) {
    if (my !== LANGS.seq) return LANGS;
    var items = (res && res.items) || [];
    if (items.length) setLanguages(items);
    LANGS.loaded = true;
    LANGS.at = Date.now();
    return LANGS;
  }, function () {
    if (my === LANGS.seq) LANGS.promise = null;
    return LANGS;
  });
  return LANGS.promise;
}

function freshLanguages() {
  return loadLanguages(!LANGS.loaded || Date.now() - LANGS.at > 30000);
}

function langInfo(code) {
  return LANGS.items.filter(function (l) { return l.code === code; })[0] || null;
}

function langDir(code) {
  var l = langInfo(code);
  if (l) return l.direction === 'rtl' ? 'rtl' : 'ltr';
  return /^(fa|ar|he|ur|ps|ckb|ku|sd|dv|yi)(-|$)/.test(code || '') ? 'rtl' : 'ltr';
}

function langName(code) {
  var l = langInfo(code);
  if (!l) return code;
  var n = l.name || {};
  return n[lang()] || l.nativeName || n[LANGS.def] || code;
}

function i18nText(v) {
  if (v == null) return '';
  if (typeof v === 'string') return v;
  if (typeof v !== 'object') return String(v);
  var l = lang();
  if (v[l]) return v[l];
  if (v[LANGS.def]) return v[LANGS.def];
  for (var i = 0; i < LANGS.active.length; i++) if (v[LANGS.active[i].code]) return v[LANGS.active[i].code];
  var keys = Object.keys(v).sort();
  for (var j = 0; j < keys.length; j++) if (typeof v[keys[j]] === 'string' && v[keys[j]]) return v[keys[j]];
  return '';
}

function getPath(obj, path) {
  var parts = path.split('.');
  for (var i = 0; i < parts.length; i++) {
    if (obj == null) return undefined;
    obj = obj[parts[i]];
  }
  return obj;
}

function setPath(obj, path, value) {
  var parts = path.split('.');
  for (var i = 0; i < parts.length - 1; i++) {
    if (obj[parts[i]] == null || typeof obj[parts[i]] !== 'object') obj[parts[i]] = {};
    obj = obj[parts[i]];
  }
  obj[parts[parts.length - 1]] = value;
}

function clone(v) {
  return v === undefined ? undefined : JSON.parse(JSON.stringify(v));
}

function h(tag, props) {
  var el = tag.indexOf('svg:') === 0 ? document.createElementNS(SVG_NS, tag.slice(4)) : document.createElement(tag);
  if (props) {
    for (var k in props) {
      var v = props[k];
      if (v === null || v === undefined || v === false) continue;
      if (k === 'class') el.setAttribute('class', v);
      else if (k === 'text') el.textContent = v;
      else if (k === 'html') el.innerHTML = v;
      else if (k.indexOf('on') === 0 && typeof v === 'function') el.addEventListener(k.slice(2).toLowerCase(), v);
      else if (k === 'value') el.value = v;
      else if (k === 'checked') el.checked = !!v;
      else el.setAttribute(k, v === true ? '' : v);
    }
  }
  for (var i = 2; i < arguments.length; i++) append(el, arguments[i]);
  return el;
}

function append(el, child) {
  if (child === null || child === undefined || child === false) return;
  if (Array.isArray(child)) { child.forEach(function (c) { append(el, c); }); return; }
  el.appendChild(typeof child === 'object' ? child : document.createTextNode(String(child)));
}

function clear(el) {
  while (el.firstChild) el.removeChild(el.firstChild);
  return el;
}

function ApiError(status, code, message, field) {
  this.status = status;
  this.code = code;
  this.message = message;
  this.field = field;
}

function token() {
  var stores = [];
  try { stores.push(window.localStorage); } catch (e) { stores.push(null); }
  try { stores.push(window.sessionStorage); } catch (e) { stores.push(null); }
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

function toLogin() {
  if (!/^#\/login/.test(location.hash)) location.hash = '#/login';
}

function qs(params) {
  var out = [];
  Object.keys(params || {}).forEach(function (k) {
    var v = params[k];
    if (v === undefined || v === null || v === '') return;
    out.push(encodeURIComponent(k) + '=' + encodeURIComponent(toLatinDigits(v)));
  });
  return out.length ? '?' + out.join('&') : '';
}

function request(base, method, path, body, isForm) {
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  var init = { method: method, headers: headers, credentials: 'same-origin' };
  if (body !== undefined) {
    if (isForm) init.body = body;
    else { headers['Content-Type'] = 'application/json'; init.body = JSON.stringify(body); }
  }
  return fetch(base + path, init).then(function (res) {
    if (res.status === 401) {
      toLogin();
      throw new ApiError(401, 'UNAUTHENTICATED', '');
    }
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) {
        var code = data && (typeof data.error === 'string' ? data.error : (data.code && typeof data.code === 'string' ? data.code : null));
        var msg = data && (data.message || (data.error && data.error.message)) || '';
        if (!code && /^[A-Z][A-Z0-9_]+$/.test(msg)) code = msg;
        if (code === 'INVALID_LANGUAGE') loadLanguages(true);
        throw new ApiError(res.status, code, msg, data && data.field);
      }
      return data;
    });
  }, function () {
    throw new ApiError(0, 'NETWORK', '');
  });
}

function api(method, path, body, isForm) {
  return request(API_BASE, method, path, body, isForm);
}

function pubApi(method, path, body, isForm) {
  return request(PUB_BASE, method, path, body, isForm);
}

function baseApi(method, path, body, isForm) {
  return request(BASE_BASE, method, path, body, isForm);
}

function consoleApi(method, path) {
  return request(CONSOLE_BASE, method, path);
}

function errorText(err, langCode) {
  if (!err) return t('err.generic');
  if (err.code && has('code.' + err.code)) return t('code.' + err.code, { lang: langName(langCode || LANGS.def) });
  if (err.status === 0) return t('err.network');
  if (err.status === 403) return t('err.forbidden');
  if (err.status === 404 && !err.code) return t('err.notfound');
  if (err.status === 404) return t('err.item.notfound');
  if (err.message && lang() === 'en') return err.message;
  if (err.message && /[؀-ۿ]/.test(err.message)) return err.message;
  return t('err.generic') + (err.code ? ' (' + err.code + ')' : '');
}

var ICONS = {
  overview: '<path d="M3 3v18h18"/><path d="M8 17v-4M13 17V8M18 17v-7"/>',
  currencies: '<circle cx="9" cy="9" r="6"/><path d="M18.1 10.4a6 6 0 1 1-7.7 7.7"/><path d="M8 7h1.5v4"/>',
  metas: '<path d="M21 4h-7M10 4H3M21 12h-9M8 12H3M21 20h-5M12 20H3M14 2v4M8 10v4M16 18v4"/>',
  prizes: '<rect x="3" y="8" width="18" height="4" rx="1"/><path d="M12 8v13M19 12v7a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2v-7"/><path d="M7.5 8a2.5 2.5 0 0 1 0-5C10 3 12 8 12 8s2-5 4.5-5a2.5 2.5 0 0 1 0 5"/>',
  products: '<path d="M3 9l1.5-5h15L21 9"/><path d="M3 9a3 3 0 0 0 6 0a3 3 0 0 0 6 0a3 3 0 0 0 6 0"/><path d="M5 12v8h14v-8M10 20v-5h4v5"/>',
  gifts: '<path d="M2 9a3 3 0 0 1 0 6v2a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-2a3 3 0 0 1 0-6V7a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2z"/><path d="M13 5v2M13 11v2M13 17v2"/>',
  payments: '<path d="M4 2v20l2-1 2 1 2-1 2 1 2-1 2 1 2-1 2 1V2l-2 1-2-1-2 1-2-1-2 1-2-1-2 1z"/><path d="M16 8H8M16 12H8M13 16H8"/>',
  categories: '<rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><circle cx="17.5" cy="17.5" r="3.5"/>',
  languages: '<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18"/>',
  tags: '<path d="M12.6 2.6A2 2 0 0 0 11.2 2H4a2 2 0 0 0-2 2v7.2a2 2 0 0 0 .6 1.4l8.7 8.7a2.4 2.4 0 0 0 3.4 0l6.6-6.6a2.4 2.4 0 0 0 0-3.4z"/><circle cx="7.5" cy="7.5" r="1"/>',
  search: '<circle cx="11" cy="11" r="7"/><path d="M21 21l-4.3-4.3"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  close: '<path d="M18 6L6 18M6 6l12 12"/>',
  first: '<path d="M11 17l-5-5 5-5M18 17l-5-5 5-5"/>',
  prev: '<path d="M15 18l-6-6 6-6"/>',
  next: '<path d="M9 18l6-6-6-6"/>',
  back: '<path d="M19 12H5M12 19l-7-7 7-7"/>',
  refresh: '<path d="M21 12a9 9 0 1 1-2.6-6.4L21 8"/><path d="M21 3v5h-5"/>',
  trash: '<path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6"/>',
  chevron: '<path d="M6 9l6 6 6-6"/>',
  upload: '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/>',
  empty: '<path d="M22 12h-6l-2 3h-4l-2-3H2"/><path d="M5.5 5.1L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.5-6.9A2 2 0 0 0 16.8 4H7.2a2 2 0 0 0-1.7 1.1z"/>',
  alert: '<circle cx="12" cy="12" r="9"/><path d="M12 8v4M12 16h.01"/>',
  check: '<path d="M20 6L9 17l-5-5"/>',
  faqs: '<circle cx="12" cy="12" r="9"/><path d="M9.1 9a3 3 0 0 1 5.8 1c0 2-3 3-3 3M12 17h.01"/>',
  news: '<path d="M4 22h16a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v16a2 2 0 0 1-2 2zm0 0a2 2 0 0 1-2-2v-9c0-1.1.9-2 2-2h2"/><path d="M18 14h-8M15 18h-5M10 6h8v4h-8z"/>',
  comments: '<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>',
  notifications: '<path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9"/><path d="M10.3 21a1.94 1.94 0 0 0 3.4 0"/>',
  send: '<path d="M22 2L11 13"/><path d="M22 2l-7 20-4-9-9-4z"/>',
  eye: '<path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/>',
  reply: '<path d="M9 17l-5-5 5-5"/><path d="M20 18v-2a4 4 0 0 0-4-4H4"/>',
  pin: '<path d="M12 17v5M9 10.8a2 2 0 0 1-1.1 1.8l-1.8.9A2 2 0 0 0 5 15.2V17h14v-1.8a2 2 0 0 0-1.1-1.8l-1.8-.9A2 2 0 0 1 15 10.8V7a1 1 0 0 1 1-1 2 2 0 0 0 0-4H8a2 2 0 0 0 0 4 1 1 0 0 1 1 1z"/>',
  users: '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.9M16 3.1a4 4 0 0 1 0 7.8"/>',
  flask: '<path d="M9 3h6M10 3v6.5L4.5 19a1.5 1.5 0 0 0 1.3 2h12.4a1.5 1.5 0 0 0 1.3-2L14 9.5V3"/><path d="M7 15h10"/>',
  image: '<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="M21 15l-3.1-3.1a2 2 0 0 0-2.8 0L6 21"/>'
};
var MIRROR = { first: 1, prev: 1, next: 1, back: 1, reply: 1 };

function icon(name, cls) {
  var el = document.createElementNS(SVG_NS, 'svg');
  el.setAttribute('xmlns', SVG_NS);
  el.setAttribute('class', 'shrink-0 x-icon inline shrink-0 stroke-2 ' + (cls || 'h-4 w-4'));
  el.setAttribute('stroke-linejoin', 'round');
  el.setAttribute('stroke-linecap', 'round');
  el.setAttribute('stroke', 'currentColor');
  el.setAttribute('fill', 'none');
  el.setAttribute('viewBox', '0 0 24 24');
  el.setAttribute('aria-hidden', 'true');
  el.setAttribute('data-l10n-icon', MIRROR[name] ? 'mirror' : '');
  el.innerHTML = ICONS[name] || '';
  return el;
}

function img(src, fallbackIcon) {
  var el = h('img', { src: src, alt: '', class: 'h-full w-full object-cover', loading: 'lazy' });
  el.addEventListener('error', function () { el.replaceWith(icon(fallbackIcon || 'image', 'h-4 w-4')); });
  return el;
}

function debounce(fn, ms) {
  var timer = null;
  return function () {
    var args = arguments;
    var self = this;
    clearTimeout(timer);
    timer = setTimeout(function () { fn.apply(self, args); }, ms);
  };
}
