(function () {
'use strict';
var SP_MESSAGES = {
  fa: {
    'rail.label': 'منوی اصلی',
    'rail.pin': 'باز نگه‌داشتن منو',
    'rail.unpin': 'جمع کردن منو',
    'rail.pages': '{n} صفحه',
    'rail.more': 'صفحه‌های {name}',
    'rail.home': 'صفحهٔ اصلی',
    'group.economy': 'اقتصاد',
    'group.publication': 'انتشار',
    'group.developer': 'توسعه',
    'group.help': 'راهنما',
    'group.settings': 'تنظیمات',
    'settings.console': 'تنظیمات کنسول',
    'settings.users': 'کاربران',
    'settings.general': 'عمومی',
    'settings.config': 'پیکربندی',
    'settings.data': 'مدیریت داده',
    'settings.audit': 'گزارش رویدادها',
    'settings.project': 'تنظیمات پروژه',
    'top.search': 'جست‌وجو',
    'top.searchPages': 'جست‌وجوی صفحه‌ها',
    'top.noResults': 'صفحه‌ای پیدا نشد',
    'top.status': 'وضعیت سرور',
    'top.live': 'زنده',
    'top.lang': 'زبان',
    'top.toDark': 'حالت تیره',
    'top.toLight': 'حالت روشن',
    'top.menu': 'منو',
    'top.close': 'بستن',
    'top.sections': 'صفحه‌های این بخش',
    'crumb.label': 'مسیر صفحه',
    'palette.hint': 'برای رفتن Enter و برای بستن Esc',
    'palette.group.pages': 'صفحه‌ها',
    'palette.group.settings': 'تنظیمات'
  },
  en: {
    'rail.label': 'Main menu',
    'rail.pin': 'Keep the menu open',
    'rail.unpin': 'Collapse the menu',
    'rail.pages': '{n} pages',
    'rail.more': '{name} pages',
    'rail.home': 'Home',
    'group.economy': 'Economy',
    'group.publication': 'Publishing',
    'group.developer': 'Developer',
    'group.help': 'Help',
    'group.settings': 'Settings',
    'settings.console': 'Console settings',
    'settings.users': 'Users',
    'settings.general': 'General',
    'settings.config': 'Configuration',
    'settings.data': 'Data management',
    'settings.audit': 'Audit log',
    'settings.project': 'Project settings',
    'top.search': 'Search',
    'top.searchPages': 'Search pages',
    'top.noResults': 'No page found',
    'top.status': 'Server status',
    'top.live': 'Live',
    'top.lang': 'Language',
    'top.toDark': 'Dark mode',
    'top.toLight': 'Light mode',
    'top.menu': 'Menu',
    'top.close': 'Close',
    'top.sections': 'Pages in this section',
    'crumb.label': 'Breadcrumb',
    'palette.hint': 'Enter to open, Esc to close',
    'palette.group.pages': 'Pages',
    'palette.group.settings': 'Settings'
  }
};

function spLang() {
  var l = (document.documentElement.getAttribute('lang') || 'fa').toLowerCase();
  return l.indexOf('fa') === 0 ? 'fa' : 'en';
}

function spNum(n) {
  var s = String(n);
  if (spLang() !== 'fa') return s;
  return s.replace(/[0-9]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'[d]; });
}

function st(key, vars) {
  var cat = SP_MESSAGES[spLang()] || SP_MESSAGES.en;
  var s = cat[key] !== undefined ? cat[key] : (SP_MESSAGES.en[key] !== undefined ? SP_MESSAGES.en[key] : key);
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] === undefined ? m : typeof vars[k] === 'number' ? spNum(vars[k]) : vars[k]; });
  return s;
}

var SP_KEY = 'nakama.spatial';
var SP_RAIL_KEY = 'nakama.spatial.rail';
var root = document.documentElement;

function spEnabled() {
  var m = /[?&]spatial=([01])\b/.exec(location.search) || /[?&]spatial=([01])\b/.exec(location.hash);
  try {
    if (m && m[1] === '0') localStorage.setItem(SP_KEY, '0');
    else if (m && m[1] === '1') localStorage.removeItem(SP_KEY);
    return localStorage.getItem(SP_KEY) !== '0';
  } catch (e) {
    return !(m && m[1] === '0');
  }
}

if (!spEnabled()) return;
root.classList.add('sp');

var mqReduce = window.matchMedia ? window.matchMedia('(prefers-reduced-motion: reduce)') : { matches: false };
var mqMobile = window.matchMedia ? window.matchMedia('(max-width: 767px)') : { matches: false };

function reduced() {
  return !!mqReduce.matches;
}

function isRtl() {
  return (root.getAttribute('dir') || 'ltr') === 'rtl';
}

function nc() {
  return window.NakamaConsole || null;
}

function el(tag, props) {
  var e = document.createElement(tag);
  if (props) {
    for (var k in props) {
      var v = props[k];
      if (v === null || v === undefined || v === false) continue;
      if (k === 'text') e.textContent = v;
      else if (k === 'html') e.innerHTML = v;
      else if (k.indexOf('on') === 0 && typeof v === 'function') e.addEventListener(k.slice(2), v);
      else e.setAttribute(k, v === true ? '' : v);
    }
  }
  for (var i = 2; i < arguments.length; i++) {
    var c = arguments[i];
    if (c === null || c === undefined || c === false) continue;
    e.appendChild(typeof c === 'string' ? document.createTextNode(c) : c);
  }
  return e;
}

var SP_ICON_PATHS = {
  star: '<path d="M11.525 2.295a.53.53 0 0 1 .95 0l2.31 4.679a2.123 2.123 0 0 0 1.595 1.16l5.166.756a.53.53 0 0 1 .294.904l-3.736 3.638a2.123 2.123 0 0 0-.611 1.878l.882 5.14a.53.53 0 0 1-.771.56l-4.618-2.428a2.122 2.122 0 0 0-1.973 0L6.396 21.01a.53.53 0 0 1-.77-.56l.881-5.139a2.122 2.122 0 0 0-.611-1.879L2.16 9.795a.53.53 0 0 1 .294-.906l5.165-.755a2.122 2.122 0 0 0 1.597-1.16z" fill="currentColor" stroke="none"/>',
  coins: '<circle cx="8" cy="8" r="6"/><path d="M18.09 10.37A6 6 0 1 1 10.34 18"/><path d="M7 6h1v4"/><path d="m16.71 13.88.7.71-2.82 2.82"/>',
  megaphone: '<path d="m3 11 18-5v12L3 14v-3z"/><path d="M11.6 16.8a3 3 0 1 1-5.8-1.6"/>',
  code: '<path d="m16 18 6-6-6-6"/><path d="m8 6-6 6 6 6"/>',
  help: '<circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/>',
  gear: '<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/>',
  sliders: '<path d="M20 7h-9"/><path d="M14 17H5"/><circle cx="17" cy="17" r="3"/><circle cx="7" cy="7" r="3"/>',
  search: '<circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>',
  moon: '<path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/>',
  chevron: '<path d="m9 18 6-6-6-6"/>',
  panel: '<rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/>',
  menu: '<path d="M4 6h16"/><path d="M4 12h16"/><path d="M4 18h16"/>',
  close: '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
  external: '<path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>',
  activity: '<path d="M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2"/>',
  page: '<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/>'
};

function spIcon(name, cls) {
  var p = SP_ICON_PATHS[name] || SP_ICON_PATHS.page;
  return '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" class="' + (cls || 'sp-ic') + (name === 'chevron' ? ' sp-ic-dir' : '') + '">' + p + '</svg>';
}

function currentHash() {
  var h = location.hash || '#/';
  return h === '#' ? '#/' : h;
}

function routePath(hash) {
  return (hash || currentHash()).replace(/^#/, '').split('?')[0] || '/';
}

function hrefActive(href, hash) {
  if (!href || href.indexOf('#/') !== 0) return false;
  var p = routePath(hash);
  var base = href.split('?')[0].replace(/^#/, '');
  if (base === '/') return p === '/';
  return p === base || p.indexOf(base + '/') === 0;
}

var spCleanups = [];
function onIdle(fn) {
  if (window.requestIdleCallback) window.requestIdleCallback(fn, { timeout: 300 });
  else setTimeout(fn, 60);
}

var SKY = {
  light: { top: '#dfe7f3', mid: '#e9edf3', low: '#f4ece6', wash: 'rgb(170 196 236 / .55)', wash2: 'rgb(248 212 192 / .55)', glow: 'rgb(255 222 190 / .85)', sun: 'rgb(255 246 234 / .9)', sunR: 48, h1: '#d5dcea', h2: '#ccd5e6', h3: '#c5cfe2' },
  dark: { top: '#0b0e1a', mid: '#111528', low: '#1a1630', wash: 'rgb(70 96 190 / .30)', wash2: 'rgb(120 70 150 / .18)', glow: 'rgb(255 214 150 / .20)', sun: 'rgb(255 236 200 / .55)', sunR: 34, h1: '#171a2e', h2: '#121424', h3: '#0d0f1a' }
};

function skyStars() {
  var s = 7;
  var rnd = function () { s = (s * 16807) % 2147483647; return s / 2147483647; };
  var out = '';
  for (var i = 0; i < 140; i++) {
    var cx = (rnd() * 1440).toFixed(1);
    var cy = (rnd() * 560).toFixed(1);
    var r = rnd() < 0.85 ? 0.7 : 1.2;
    out += '<circle cx="' + cx + '" cy="' + cy + '" r="' + r + '" fill="#fff" opacity="' + (0.25 + rnd() * 0.45).toFixed(2) + '"/>';
  }
  return out;
}

function skySvg(dark) {
  var c = dark ? SKY.dark : SKY.light;
  var stop = function (id, cx, cy, r, color) {
    return '<radialGradient id="' + id + '" cx="' + cx + '" cy="' + cy + '" r="' + r + '"><stop offset="0" stop-color="' + color + '"/><stop offset="1" stop-color="' + color + '" stop-opacity="0"/></radialGradient>';
  };
  return '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice" aria-hidden="true" focusable="false"><defs>' +
    '<linearGradient id="spk" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="' + c.top + '"/><stop offset=".6" stop-color="' + c.mid + '"/><stop offset="1" stop-color="' + c.low + '"/></linearGradient>' +
    stop('spw1', 0.82, 0.2, 0.5, c.wash) + stop('spw2', 0.35, 1, 0.55, c.wash2) + stop('spg', 0.22, 0.14, 0.42, c.glow) +
    '<filter id="sph" x="-10%" y="-10%" width="120%" height="120%"><feGaussianBlur stdDeviation="18"/></filter>' +
    '<filter id="sps" x="-150%" y="-150%" width="400%" height="400%"><feGaussianBlur stdDeviation="34"/></filter></defs>' +
    '<rect width="1440" height="900" fill="url(#spk)"/><rect width="1440" height="900" fill="url(#spw1)"/><rect width="1440" height="900" fill="url(#spw2)"/><rect width="1440" height="900" fill="url(#spg)"/>' +
    (dark ? skyStars() : '') +
    '<circle cx="300" cy="128" r="' + c.sunR + '" fill="' + c.sun + '" filter="url(#sps)"/>' +
    '<g filter="url(#sph)"><path d="M0 690 C160 640 300 650 440 676 C600 706 720 640 880 630 C1040 620 1180 668 1440 640 L1440 900 L0 900Z" fill="' + c.h1 + '"/>' +
    '<path d="M0 760 C200 720 360 744 520 760 C700 778 860 720 1040 716 C1200 712 1320 744 1440 734 L1440 900 L0 900Z" fill="' + c.h2 + '"/>' +
    '<path d="M0 836 C240 806 480 830 720 828 C960 826 1200 800 1440 812 L1440 900 L0 900Z" fill="' + c.h3 + '"/></g></svg>';
}

var skyEl = null;
function paintSky() {
  if (!document.body) return;
  var mode = root.classList.contains('dark') ? 'dark' : 'light';
  if (!skyEl) {
    skyEl = el('div', { class: 'sp-sky', 'aria-hidden': 'true', 'data-l10n-skip': '' });
  }
  if (skyEl.getAttribute('data-theme') !== mode) {
    skyEl.setAttribute('data-theme', mode);
    skyEl.innerHTML = skySvg(mode === 'dark');
  }
  if (skyEl.parentNode !== document.body || document.body.firstChild !== skyEl) document.body.insertBefore(skyEl, document.body.firstChild);
}

var CARD_SEL = '.rounded-md.shadow.bg-white';

function contentRoot() {
  var eco = document.getElementById('economy-root');
  if (eco && eco.offsetParent !== null) return eco;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') return mains[i];
  return null;
}

function routeName() {
  var p = routePath().split('/').filter(Boolean);
  if (!p.length) return 'dashboard';
  if (p[0] === 'economy' || p[0] === 'publication' || p[0] === 'settings') return p[0] + '-' + (p[1] || 'index');
  if (p[0] === 'players' || p[0] === 'groups' || p[0] === 'leaderboards' || p[0] === 'storage-indexes') return p.length > 1 ? p[0] + '-detail' : p[0];
  return p[0];
}

function markPage() {
  var name = routeName();
  if (root.getAttribute('data-sp-route') !== name) root.setAttribute('data-sp-route', name);
  var main = contentRoot();
  if (!main || main.id === 'economy-root') return;
  var box = main.querySelector('div.max-w-\\[1600px\\]');
  if (!box) return;
  var cards = box.querySelectorAll(CARD_SEL + ', .bg-white.shadow.rounded-md');
  var tables = box.querySelectorAll('table');
  var tabs = box.querySelector('.x-tab-group');
  var list = !tabs && tables.length >= 1 && cards.length === tables.length && cards.length <= 2 && !box.querySelector('[_echarts_instance_]');
  if (list !== box.hasAttribute('data-sp-list')) box.toggleAttribute('data-sp-list', list);
}

var labelSeq = 0;
function labelTables() {
  if (!mqMobile.matches) return;
  var tables = document.querySelectorAll('#main table, [data-economy] table, .ps-window table');
  for (var i = 0; i < tables.length; i++) {
    var t = tables[i];
    if (t.classList.contains('eco-table')) continue;
    var heads = Array.prototype.map.call(t.querySelectorAll('thead th'), function (th) { return th.textContent.trim(); });
    if (!heads.length) continue;
    var rows = t.tBodies.length ? t.tBodies[0].rows : [];
    var key = heads.join('|') + '#' + rows.length + '#' + (rows[0] ? rows[0].cells.length : 0);
    if (t.__spLabels === key) continue;
    t.__spLabels = key;
    labelSeq++;
    for (var r = 0; r < rows.length; r++) {
      var cells = rows[r].cells;
      if (cells.length === 1 && heads.length > 1) { cells[0].setAttribute('data-sp-full', ''); continue; }
      for (var c = 0; c < cells.length; c++) {
        var lab = heads[c] || '';
        if (lab) cells[c].setAttribute('data-sp-label', lab);
        else cells[c].removeAttribute('data-sp-label');
      }
    }
    t.setAttribute('data-sp-cards', '');
  }
}

function isAuthPage() {
  if (/^#\/(login|reset|reset-mfa|mfa)(\/|\?|$)/.test(currentHash())) return true;
  return !document.querySelector('#main') && !!document.querySelector('#app form');
}

function pageKey(path) {
  var seg = String(path || '/').split('?')[0].split('/').filter(Boolean);
  if (!seg.length) return '/';
  if (seg[0] === 'project-settings') return 'ps';
  if (seg[0] === 'economy' || seg[0] === 'publication' || seg[0] === 'base') return seg[0] + '/' + (seg[1] || '');
  if (seg[0] === 'settings') return 'settings/' + (seg[1] || '');
  if (seg[0] === 'payments') return 'payments/' + (seg[1] || '');
  return seg.slice(0, 2).join('/');
}

var enterTimer = 0;
function animateEnter(el) {
  if (!el || reduced()) return;
  el.classList.remove('sp-enter');
  void el.offsetWidth;
  el.classList.add('sp-enter');
  clearTimeout(enterTimer);
  var done = function (e) {
    if (e && e.target !== el) return;
    el.classList.remove('sp-enter');
    el.removeEventListener('animationend', done);
  };
  el.addEventListener('animationend', done);
  enterTimer = setTimeout(done, 700);
}

function routeEnter() {
  requestAnimationFrame(function () {
    requestAnimationFrame(function () { animateEnter(contentRoot()); });
  });
}

function onRouteChange(to, from) {
  var a = pageKey(to && to.path);
  var b = pageKey(from && from.path);
  if (typeof closeFly === 'function') {
    closeFly(true);
    closeSheet();
    if (tb.palette && tb.palette.close) tb.palette.close();
    toggleStats(false);
  }
  if (a !== b && a !== 'ps' && b !== 'ps') routeEnter();
  schedule();
}

function countUp(node, to, fmt) {
  if (reduced() || !isFinite(to)) { node.textContent = fmt(to); return; }
  var from = node.__spValue || 0;
  node.__spValue = to;
  var start = performance.now();
  var dur = 480;
  var step = function (now) {
    var k = Math.min(1, (now - start) / dur);
    var e = 1 - Math.pow(1 - k, 3);
    node.textContent = fmt(from + (to - from) * e);
    if (k < 1) requestAnimationFrame(step);
  };
  requestAnimationFrame(step);
}

function hookRoutes() {
  var api = nc();
  if (!api || !api.onRoute || hookRoutes.done) return !!hookRoutes.done;
  hookRoutes.done = true;
  api.onRoute(onRouteChange);
  return true;
}

var tickQueued = false;
var lastTick = 0;

function tick() {
  tickQueued = false;
  lastTick = Date.now();
  paintSky();
  var auth = isAuthPage();
  if (auth !== root.classList.contains('sp-auth')) {
    root.classList.toggle('sp-auth', auth);
    if (auth && typeof closeFly === 'function') { closeFly(true); closeSheet(); }
  }
  if (!auth) {
    if (typeof renderRail === 'function') renderRail();
    if (typeof paintTopbar === 'function') paintTopbar();
    markPage();
    labelTables();
  }
}

function schedule() {
  if (tickQueued) return;
  tickQueued = true;
  var wait = Math.max(0, 90 - (Date.now() - lastTick));
  setTimeout(function () { requestAnimationFrame(tick); }, wait);
}

function boot() {
  paintSky();
  if (!hookRoutes()) {
    window.addEventListener('nakama-console:ready', hookRoutes);
    var tries = 0;
    var poll = setInterval(function () {
      tries++;
      if (hookRoutes() || tries > 50) {
        clearInterval(poll);
        if (!hookRoutes.done) window.addEventListener('hashchange', function () { routeEnter(); schedule(); });
      }
    }, 100);
  }
  window.addEventListener('hashchange', schedule);
  window.addEventListener('resize', function () {
    if (typeof segIndicator === 'function') {
      if (tb.start) segIndicator(tb.start.querySelector('.sp-tabs'));
      if (tb.seg) segIndicator(tb.seg);
      if (!mqMobile.matches) closeSheet();
    }
    schedule();
  });
  if (typeof globalKeys === 'function') document.addEventListener('keydown', globalKeys);
  new MutationObserver(function (records) {
    for (var i = 0; i < records.length; i++) {
      var t = records[i].target;
      if (t.nodeType === 1 && t.closest && t.closest('.sp-rail, .sp-rail-full, .sp-flyout, .sp-tb-start, .sp-tb-end, .sp-sky, .sp-tip, .sp-pal-holder, [_echarts_instance_], number-flow-vue')) continue;
      schedule();
      return;
    }
  }).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function (records) {
    var hit = records.some(function (r) {
      if (r.attributeName !== 'class') return r.oldValue !== root.getAttribute(r.attributeName);
      return /(^|\s)dark(\s|$)/.test(r.oldValue || '') !== root.classList.contains('dark');
    });
    if (!hit) return;
    if (typeof rail === 'object') rail.key = '';
    if (typeof tb === 'object') tb.key = '';
    schedule();
  }).observe(root, { attributes: true, attributeOldValue: true, attributeFilter: ['class', 'lang', 'dir'] });
  if (mqMobile.addEventListener) mqMobile.addEventListener('change', function () { if (typeof rail === 'object') rail.key = ''; schedule(); });
  tick();
}

if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot);
else boot();
})();
