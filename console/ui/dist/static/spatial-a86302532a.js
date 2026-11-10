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
    'rail.sections': 'بخش‌ها',
    'rail.console': 'کنسول مدیریت',
    'group.economy': 'اقتصاد',
    'group.publication': 'انتشار',
    'group.developer': 'توسعه',
    'group.help': 'راهنما',
    'group.settings': 'تنظیمات',
    'settings.console': 'تنظیمات کنسول',
    'settings.users': 'کاربران',
    'settings.general': 'عمومی',
    'settings.config': 'پیکربندی',
    'settings.data': 'مدیریت داده‌ها',
    'settings.audit': 'گزارش ممیزی',
    'nav.purchases': 'خریدها',
    'nav.chat': 'گفتگو',
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
    'rail.sections': 'Sections',
    'rail.console': 'Admin console',
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
    'nav.purchases': 'Purchases',
    'nav.chat': 'Chat',
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
  users: '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
  toggle: '<rect width="20" height="12" x="2" y="6" rx="6"/><circle cx="16" cy="12" r="2"/>',
  fileCode: '<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="m10 13-2 2 2 2"/><path d="m14 17 2-2-2-2"/>',
  database: '<ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14a9 3 0 0 0 18 0V5"/><path d="M3 12a9 3 0 0 0 18 0"/>',
  list: '<path d="M8 6h13"/><path d="M8 12h13"/><path d="M8 18h13"/><path d="M3 6h.01"/><path d="M3 12h.01"/><path d="M3 18h.01"/>',
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
  dark: { top: '#0b0e1a', mid: '#111528', low: '#1a1630', wash: 'rgb(70 96 190 / .30)', wash2: 'rgb(120 70 150 / .18)', glow: 'rgb(255 214 150 / .20)', sun: 'rgb(255 236 200 / .30)', sunR: 34, h1: '#171a2e', h2: '#121424', h3: '#0d0f1a' }
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

var DIRECT = ['/', '/players', '/groups', '/storage', '/storage-indexes', '/matches', '/leaderboards', '/notifications', '/payments', '/chat'];
var SETTINGS_PAGES = [['users', 'settings.users', 'users'], ['general', 'settings.general', 'toggle'], ['config', 'settings.config', 'fileCode'], ['data-management', 'settings.data', 'database'], ['audit-log', 'settings.audit', 'list']];
var NATIVE_LABEL = { '/payments': 'nav.purchases', '/chat': 'nav.chat' };
var GROUP_ICON = { economy: 'coins', publication: 'megaphone', developer: 'code', help: 'help', settings: 'gear' };

function linkInfo(a) {
  var badgeEl = a.querySelector('.eco-nav-badge, .eco-nav-dot');
  var label = '';
  var span = a.querySelector('span:not(.eco-nav-badge)');
  label = (span && span.textContent.trim()) || a.getAttribute('title') || a.textContent.trim();
  if (badgeEl && badgeEl.textContent) label = label.replace(badgeEl.textContent, '').trim();
  var svg = a.querySelector('svg');
  var badge = 0;
  if (badgeEl) {
    var digits = (badgeEl.textContent || '').replace(/[۰-۹]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'.indexOf(d); }).replace(/[^0-9]/g, '');
    badge = parseInt(digits, 10) || (badgeEl.classList.contains('eco-nav-dot') ? 1 : 0);
  }
  var renamed = NATIVE_LABEL[routePath(a.getAttribute('href') || '#')];
  if (renamed) label = st(renamed);
  return { href: a.getAttribute('href') || '', label: label, svg: svg ? svg.outerHTML : '', badge: badge, external: !/^#\//.test(a.getAttribute('href') || '') };
}

function railModel() {
  var drawer = document.querySelector('.x-drawer');
  var grid = drawer && drawer.querySelector('.grid');
  if (!grid) return null;
  var model = { direct: [], groups: [], settings: null, help: null, logo: null };
  var helpLinks = [];
  var devLinks = [];
  var consoleSettings = null;
  Array.prototype.forEach.call(grid.children, function (g) {
    var links = Array.prototype.map.call(g.querySelectorAll('a[href]'), linkInfo).filter(function (l) { return l.href; });
    if (!links.length) return;
    var labelEl = null;
    for (var i = 0; i < g.children.length; i++) if (g.children[i].tagName !== 'NAV' && !g.children[i].querySelector('a')) { labelEl = g.children[i]; break; }
    var label = labelEl ? labelEl.textContent.trim() : '';
    var sec = g.getAttribute('data-economy-nav');
    if (sec) {
      model.groups.push({ id: sec, label: label || st('group.' + sec), icon: GROUP_ICON[sec] || 'page', links: links });
      return;
    }
    if (links.every(function (l) { return l.external; })) { helpLinks = helpLinks.concat(links); return; }
    links.forEach(function (l) {
      if (l.external) { helpLinks.push(l); return; }
      var p = routePath(l.href);
      if (p === '/settings' || p.indexOf('/settings/') === 0) { consoleSettings = l; return; }
      if (DIRECT.indexOf(p) !== -1) model.direct.push(l);
      else devLinks.push(l);
    });
  });
  var footer = drawer.querySelectorAll(':scope > .grid a[href], :scope > div:last-child a[href^="http"]');
  Array.prototype.forEach.call(footer, function (a) {
    var info = linkInfo(a);
    if (info.external && !helpLinks.some(function (h) { return h.href === info.href; })) helpLinks.push(info);
  });
  if (devLinks.length) model.groups.push({ id: 'developer', label: st('group.developer'), icon: 'code', links: devLinks });
  var settingsLinks = [];
  if (consoleSettings) {
    SETTINGS_PAGES.forEach(function (sp) {
      settingsLinks.push({ href: '#/settings/' + sp[0], label: st(sp[1]), svg: spIcon(sp[2]), badge: 0 });
    });
  }
  var psApi = nc() && nc().projectSettings;
  var psEntry = document.querySelector('[data-ps-nav] a[href]');
  if (psApi && psApi.allowed()) settingsLinks.push({ href: psApi.href(), label: st('settings.project'), svg: spIcon('sliders'), badge: 0, ps: true });
  else if (psEntry) settingsLinks.push({ href: psEntry.getAttribute('href'), label: st('settings.project'), svg: spIcon('sliders'), badge: 0, ps: true });
  if (settingsLinks.length) model.settings = { id: 'settings', label: st('group.settings'), icon: 'gear', links: settingsLinks };
  if (helpLinks.length) model.help = { id: 'help', label: st('group.help'), icon: 'help', links: helpLinks };
  var logo = drawer.querySelector('a[href="#/"] img:not([data-brand-hidden])');
  model.logo = logo;
  return model;
}

function groupActive(g, hash) {
  return g.links.some(function (l) { return linkActive(l, hash); });
}

function linkActive(l, hash) {
  if (l.ps) return /^#\/project-settings(\/|$)/.test(hash || currentHash());
  if (/^#\/project-settings/.test(hash || currentHash())) return false;
  return hrefActive(l.href, hash);
}

function groupBadge(g) {
  return g.links.reduce(function (n, l) { return n + (l.badge || 0); }, 0);
}

var rail = { el: null, full: null, fly: null, flyGroup: null, flyBtn: null, closeTimer: 0, key: '', model: null, tip: null, subOpen: null };

function railPinned() {
  try { return localStorage.getItem(SP_RAIL_KEY) === 'pinned'; } catch (e) { return false; }
}

function setPinned(on) {
  try { if (on) localStorage.setItem(SP_RAIL_KEY, 'pinned'); else localStorage.removeItem(SP_RAIL_KEY); } catch (e) { on = !!on; }
  root.classList.toggle('sp-pinned', !!on);
  closeFly(true);
  rail.key = '';
  renderRail();
}

function brandTile() {
  var b = el('a', { class: 'sp-brand', href: '#/', 'aria-label': st('rail.home') });
  var mark = null;
  var info = nc() && nc().brand ? nc().brand() : null;
  if (info && info.markUrl) mark = el('img', { src: info.markUrl, alt: '' });
  if (mark) b.appendChild(mark);
  else b.innerHTML = spIcon('star', 'sp-ic sp-ic-star');
  return b;
}

function railItem(l, cls) {
  var on = linkActive(l);
  var a = el('a', { class: cls || 'sp-ri', href: l.href, 'aria-label': l.label, 'data-sp-tip': l.label, 'aria-current': on ? 'page' : null, 'data-sp-href': l.href });
  a.innerHTML = l.svg || spIcon('page');
  if (cls === 'sp-row') a.appendChild(el('span', { class: 'sp-row-label', text: l.label }));
  if (l.badge) a.appendChild(el('span', { class: cls === 'sp-row' ? 'sp-chip sp-chip-accent' : 'sp-count', text: spNum(l.badge) }));
  if (l.ps) { a.addEventListener('click', psClick); a.setAttribute('data-ps-entry', 'rail'); }
  return a;
}

function psClick(e) {
  if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
  var api = nc() && nc().projectSettings;
  if (!api) return;
  e.preventDefault();
  closeFly(true);
  closeSheet();
  api.open();
}

function groupButton(g, cls) {
  var on = groupActive(g);
  var b = el('button', { type: 'button', class: cls || 'sp-ri sp-ri-group', 'aria-haspopup': 'menu', 'aria-expanded': 'false', 'aria-label': g.label, 'data-sp-group': g.id, 'data-sp-tip': cls ? null : g.label, 'aria-current': on ? 'true' : null });
  b.innerHTML = spIcon(g.icon);
  var sum = groupBadge(g);
  if (cls === 'sp-row sp-row-group') {
    b.appendChild(el('span', { class: 'sp-row-label', text: g.label }));
    b.appendChild(el('span', { class: sum ? 'sp-chip sp-chip-accent' : 'sp-chip', text: spNum(sum || g.links.length) }));
    b.insertAdjacentHTML('beforeend', spIcon('chevron', 'sp-ic sp-ic-dir sp-row-chev'));
  } else {
    b.appendChild(el('span', { class: 'sp-more', 'aria-hidden': 'true' }));
    if (sum) b.appendChild(el('span', { class: 'sp-count', text: spNum(sum) }));
  }
  b.addEventListener('mouseenter', function () { if (!mqMobile.matches) openFly(b, g, false); });
  b.addEventListener('mouseleave', scheduleClose);
  b.addEventListener('focus', function () { if (!b.__spNoFocusOpen && !mqMobile.matches) openFly(b, g, false); b.__spNoFocusOpen = false; });
  b.addEventListener('click', function () {
    if (mqMobile.matches && rail.full && rail.full.contains(b)) { inlineGroup(b, g, false); return; }
    if (rail.fly && rail.fly.classList.contains('open') && rail.flyGroup && rail.flyGroup.id === g.id && rail.flyBtn === b) { if (!b.__spJustOpened) closeFly(); }
    else openFly(b, g, false);
  });
  b.addEventListener('keydown', function (e) {
    var into = isRtl() ? 'ArrowLeft' : 'ArrowRight';
    if (e.key === 'Enter' || e.key === ' ' || e.key === into || e.key === 'ArrowDown' && b.closest('.sp-rail-full')) {
      e.preventDefault();
      openFly(b, g, true);
    } else if (e.key === 'Escape') {
      closeFly();
    }
  });
  return b;
}

function ensureFly() {
  if (rail.fly) return rail.fly;
  var f = el('div', { class: 'sp-flyout', role: 'menu', 'data-l10n-skip': '' });
  f.addEventListener('mouseenter', function () { clearTimeout(rail.closeTimer); });
  f.addEventListener('mouseleave', scheduleClose);
  f.addEventListener('keydown', function (e) {
    var items = Array.prototype.slice.call(f.querySelectorAll('[role="menuitem"]'));
    var i = items.indexOf(document.activeElement);
    var back = isRtl() ? 'ArrowRight' : 'ArrowLeft';
    if (e.key === 'ArrowDown') { e.preventDefault(); items[(i + 1) % items.length].focus(); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); items[(i - 1 + items.length) % items.length].focus(); }
    else if (e.key === 'Home') { e.preventDefault(); items[0].focus(); }
    else if (e.key === 'End') { e.preventDefault(); items[items.length - 1].focus(); }
    else if (e.key === 'Escape' || e.key === back) {
      e.preventDefault();
      var btn = rail.flyBtn;
      closeFly();
      if (btn) { btn.__spNoFocusOpen = true; btn.focus(); }
    } else if (e.key === 'Tab') {
      closeFly();
    }
  });
  document.body.appendChild(f);
  rail.fly = f;
  return f;
}

function paintFly(g) {
  var f = rail.fly;
  f.innerHTML = '';
  f.setAttribute('aria-label', g.label);
  f.setAttribute('data-group', g.id);
  f.setAttribute('dir', root.getAttribute('dir') || 'ltr');
  f.appendChild(el('div', { class: 'sp-fly-head' }, el('b', { text: g.label }), el('span', { text: st('rail.pages', { n: g.links.length }) })));
  g.links.forEach(function (l) {
    var on = linkActive(l);
    var a = el('a', { role: 'menuitem', href: l.href, tabindex: '-1', 'aria-current': on ? 'page' : null, target: l.external ? '_blank' : null, rel: l.external ? 'noopener' : null });
    a.innerHTML = l.svg || spIcon(l.external ? 'external' : 'page');
    a.appendChild(el('span', { class: 'sp-fly-label', text: l.label }));
    if (l.badge) a.appendChild(el('em', { class: 'sp-fly-count', text: spNum(l.badge) }));
    if (l.ps) { a.addEventListener('click', psClick); a.setAttribute('data-ps-entry', 'rail'); }
    else a.addEventListener('click', function () { closeFly(true); closeSheet(); });
    f.appendChild(a);
  });
}

function inlineGroup(btn, g, focusFirst) {
  var box = btn.parentNode;
  var open = box.querySelector(':scope > .sp-sub[data-group="' + g.id + '"]');
  Array.prototype.forEach.call(rail.full.querySelectorAll('.sp-sub'), function (s) {
    if (s === open) return;
    var b = s.previousElementSibling;
    if (b) b.setAttribute('aria-expanded', 'false');
    s.remove();
  });
  if (open && !focusFirst) {
    open.remove();
    btn.setAttribute('aria-expanded', 'false');
    rail.subOpen = null;
    return;
  }
  if (!open) {
    open = el('div', { class: 'sp-sub', role: 'group', 'aria-label': g.label, 'data-group': g.id });
    g.links.forEach(function (l) {
      var a = railItem(l, 'sp-row');
      a.classList.add('sp-sub-row');
      if (l.external) { a.setAttribute('target', '_blank'); a.setAttribute('rel', 'noopener'); }
      if (!l.ps) a.addEventListener('click', function () { closeSheet(); });
      open.appendChild(a);
    });
    btn.after(open);
  }
  btn.setAttribute('aria-expanded', 'true');
  rail.subOpen = g.id;
  if (focusFirst) {
    var first = open.querySelector('[aria-current="page"]') || open.querySelector('a');
    if (first) first.focus();
  }
}

function openFly(btn, g, focusFirst) {
  if (mqMobile.matches && rail.full && rail.full.contains(btn)) { inlineGroup(btn, g, focusFirst); return; }
  clearTimeout(rail.closeTimer);
  var f = ensureFly();
  if (rail.flyBtn && rail.flyBtn !== btn) rail.flyBtn.setAttribute('aria-expanded', 'false');
  if (!rail.flyGroup || rail.flyGroup.id !== g.id || rail.flyBtn !== btn || f.getAttribute('data-key') !== rail.key) {
    paintFly(g);
    f.setAttribute('data-key', rail.key);
  }
  rail.flyGroup = g;
  rail.flyBtn = btn;
  hideTip();
  var r = btn.getBoundingClientRect();
  var pinned = !!btn.closest('.sp-rail-full');
  f.classList.toggle('sp-fly-pinned', pinned);
  var h = f.offsetHeight || (56 + g.links.length * 40);
  var top = Math.max(8, Math.min(r.top - 8, window.innerHeight - h - 8));
  f.style.top = top + 'px';
  var was = f.classList.contains('open');
  f.classList.add('open');
  btn.setAttribute('aria-expanded', 'true');
  if (!was) { btn.__spJustOpened = true; setTimeout(function () { btn.__spJustOpened = false; }, 250); }
  if (focusFirst) {
    var first = f.querySelector('[aria-current="page"]') || f.querySelector('[role="menuitem"]');
    if (first) first.focus();
  }
}

function closeFly(now) {
  clearTimeout(rail.closeTimer);
  if (!rail.fly) return;
  rail.fly.classList.remove('open');
  if (now) rail.fly.classList.add('sp-fly-now');
  setTimeout(function () { if (rail.fly) rail.fly.classList.remove('sp-fly-now'); }, 30);
  if (rail.flyBtn) rail.flyBtn.setAttribute('aria-expanded', 'false');
  rail.flyGroup = null;
  rail.flyBtn = null;
}

function scheduleClose() {
  clearTimeout(rail.closeTimer);
  rail.closeTimer = setTimeout(function () {
    if (rail.fly && rail.fly.contains(document.activeElement)) return;
    closeFly();
  }, 160);
}

function ensureTip() {
  if (rail.tip) return rail.tip;
  rail.tip = el('div', { class: 'sp-tip', role: 'tooltip', 'aria-hidden': 'true', 'data-l10n-skip': '' });
  document.body.appendChild(rail.tip);
  return rail.tip;
}

function showTip(target) {
  var text = target.getAttribute('data-sp-tip');
  if (!text || (rail.fly && rail.fly.classList.contains('open') && target.hasAttribute('data-sp-group'))) return;
  var tip = ensureTip();
  tip.textContent = text;
  var r = target.getBoundingClientRect();
  tip.style.top = Math.round(r.top + r.height / 2) + 'px';
  tip.setAttribute('dir', root.getAttribute('dir') || 'ltr');
  tip.classList.add('open');
}

function hideTip() {
  if (rail.tip) rail.tip.classList.remove('open');
}

function railKey(model) {
  var b = nc() && nc().brand ? nc().brand() : null;
  return [spLang(), currentHash(), railPinned(), root.classList.contains('dark'), JSON.stringify(b), JSON.stringify([model.direct, model.groups, model.settings, model.help].map(function (x) {
    return x ? JSON.stringify(x.links || x, function (k, v) { return k === 'svg' ? undefined : v; }) : '';
  }))].join('|');
}

function buildCollapsed(model) {
  var nav = rail.el;
  nav.innerHTML = '';
  nav.appendChild(brandTile());
  nav.appendChild(el('button', { type: 'button', class: 'sp-ri sp-pin', 'aria-label': st('rail.pin'), 'data-sp-tip': st('rail.pin'), html: spIcon('panel'), onclick: function () { setPinned(true); } }));
  nav.appendChild(el('hr'));
  model.direct.forEach(function (l) { nav.appendChild(railItem(l)); });
  if (model.groups.length) nav.appendChild(el('hr'));
  model.groups.forEach(function (g) { nav.appendChild(groupButton(g)); });
  nav.appendChild(el('div', { class: 'sp-grow' }));
  if (model.help) nav.appendChild(groupButton(model.help));
  if (model.settings) nav.appendChild(groupButton(model.settings));
}

function buildFull(model) {
  var p = rail.full;
  p.innerHTML = '';
  var info = nc() && nc().brand ? nc().brand() : null;
  var name = (info && (info.name || info.shortName)) || 'Nakama';
  var logo = info && (root.classList.contains('dark') ? info.logoDarkUrl || info.logoUrl : info.logoUrl);
  var title = logo ? el('img', { class: 'sp-full-logo', src: logo, alt: name }) : el('b', { text: name });
  var head = el('div', { class: 'sp-full-head' }, brandTile(), el('div', { class: 'sp-full-name' }, title, el('span', { text: st('rail.console') })),
    el('button', { type: 'button', class: 'sp-collapse', 'aria-label': st('rail.unpin'), title: st('rail.unpin'), html: spIcon('panel'), onclick: function () { setPinned(false); } }));
  p.appendChild(head);
  var scroll = el('div', { class: 'sp-full-scroll' });
  model.direct.forEach(function (l) { scroll.appendChild(railItem(l, 'sp-row')); });
  if (model.groups.length) scroll.appendChild(el('div', { class: 'sp-full-label', text: st('rail.sections') }));
  model.groups.forEach(function (g) { scroll.appendChild(groupButton(g, 'sp-row sp-row-group')); });
  p.appendChild(scroll);
  var foot = el('div', { class: 'sp-full-foot' });
  if (model.help) foot.appendChild(groupButton(model.help, 'sp-row sp-row-group'));
  if (model.settings) foot.appendChild(groupButton(model.settings, 'sp-row sp-row-group'));
  p.appendChild(foot);
  if (rail.subOpen && mqMobile.matches && root.classList.contains('sp-sheet-open')) {
    var again = p.querySelector('[data-sp-group="' + rail.subOpen + '"]');
    var grp = [].concat(model.groups, model.help ? [model.help] : [], model.settings ? [model.settings] : []).filter(function (x) { return x.id === rail.subOpen; })[0];
    if (again && grp) inlineGroup(again, grp, false);
  } else rail.subOpen = null;
}

function renderRail(force) {
  if (root.classList.contains('sp-auth')) return;
  var model = railModel();
  if (!model) {
    root.classList.remove('sp-railed');
    return;
  }
  var key = railKey(model);
  if (!force && key === rail.key && rail.el && rail.el.isConnected && rail.full && rail.full.isConnected) return;
  rail.key = key;
  rail.model = model;
  if (!rail.el) {
    rail.el = el('nav', { class: 'sp-rail', 'data-l10n-skip': '' });
    rail.el.addEventListener('mouseover', function (e) { var t = e.target.closest('[data-sp-tip]'); if (t && rail.el.contains(t)) showTip(t); });
    rail.el.addEventListener('mouseout', function (e) { var t = e.target.closest('[data-sp-tip]'); if (t && !t.contains(e.relatedTarget)) hideTip(); });
    rail.el.addEventListener('focusin', function (e) { var t = e.target.closest('[data-sp-tip]'); if (t && e.target.matches(':focus-visible')) showTip(t); });
    rail.el.addEventListener('focusout', hideTip);
    rail.el.addEventListener('keydown', railKeys);
    rail.full = el('nav', { class: 'sp-rail-full', 'data-l10n-skip': '' });
    rail.full.addEventListener('keydown', railKeys);
  }
  rail.el.setAttribute('aria-label', st('rail.label'));
  rail.full.setAttribute('aria-label', st('rail.label'));
  var act = document.activeElement;
  var keep = null;
  if (act && (rail.el.contains(act) || rail.full.contains(act))) {
    keep = { full: rail.full.contains(act), href: act.getAttribute('data-sp-href'), group: act.getAttribute('data-sp-group'), cls: act.className };
  }
  buildCollapsed(model);
  buildFull(model);
  if (keep) {
    var box = keep.full ? rail.full : rail.el;
    var back = keep.href ? box.querySelector('[data-sp-href="' + keep.href.replace(/"/g, '') + '"]') : keep.group ? box.querySelector('[data-sp-group="' + keep.group + '"]') : box.querySelector('.' + String(keep.cls).split(' ')[0]);
    if (back) { back.__spNoFocusOpen = true; back.focus({ preventScroll: true }); }
  }
  rail.el.setAttribute('aria-hidden', railPinned() ? 'true' : 'false');
  rail.el.toggleAttribute('inert', railPinned());
  rail.full.setAttribute('aria-hidden', railPinned() ? 'false' : 'true');
  rail.full.toggleAttribute('inert', !railPinned());
  if (rail.el.parentNode !== document.body) document.body.appendChild(rail.el);
  if (rail.full.parentNode !== document.body) document.body.appendChild(rail.full);
  root.classList.add('sp-railed');
  root.classList.toggle('sp-pinned', railPinned());
  if (rail.fly && rail.fly.classList.contains('open') && rail.flyGroup) {
    var g = [].concat(model.groups, model.help ? [model.help] : [], model.settings ? [model.settings] : []).filter(function (x) { return x.id === rail.flyGroup.id; })[0];
    if (g) paintFly(g);
  }
  if (typeof paintTopbar === 'function') paintTopbar();
}

function railKeys(e) {
  if (e.key !== 'ArrowDown' && e.key !== 'ArrowUp') return;
  if (e.target.closest('.sp-flyout')) return;
  if (e.key === 'ArrowDown' && e.target.closest('.sp-rail-full') && e.target.hasAttribute('data-sp-group')) return;
  var box = e.currentTarget;
  var items = Array.prototype.slice.call(box.querySelectorAll('a.sp-ri, button.sp-ri, a.sp-row, button.sp-row, .sp-brand, .sp-collapse'));
  var i = items.indexOf(document.activeElement);
  if (i === -1) return;
  e.preventDefault();
  var next = items[(i + (e.key === 'ArrowDown' ? 1 : -1) + items.length) % items.length];
  next.__spNoFocusOpen = true;
  next.focus();
}

function allPages() {
  var m = railModel() || rail.model;
  if (!m) return [];
  var out = [];
  m.direct.forEach(function (l) { out.push({ label: l.label, href: l.href, svg: l.svg, group: '' }); });
  m.groups.forEach(function (g) { g.links.forEach(function (l) { out.push({ label: l.label, href: l.href, svg: l.svg, group: g.label }); }); });
  if (m.settings) m.settings.links.forEach(function (l) { out.push({ label: l.label, href: l.href, svg: l.svg, group: m.settings.label, ps: l.ps }); });
  var api = nc() && nc().projectSettings;
  if (api && api.allowed()) api.sections().forEach(function (s) { out.push({ label: s.label, href: s.href, group: st('settings.project'), psSection: true }); });
  return out;
}

function pageInfo(hash) {
  var m = rail.model || railModel();
  if (!m) return null;
  var best = null;
  var consider = function (l, g) {
    if (l.external || !linkActive(l, hash)) return;
    var len = l.href.split('?')[0].length;
    if (!best || len > best.len) best = { link: l, group: g, len: len };
  };
  m.direct.forEach(function (l) { consider(l, null); });
  m.groups.forEach(function (g) { g.links.forEach(function (l) { consider(l, g); }); });
  if (m.settings) m.settings.links.forEach(function (l) { consider(l, m.settings); });
  return best;
}

var sheetPrev = null;

function openSheet() {
  if (root.classList.contains('sp-sheet-open')) return;
  renderRail(true);
  sheetPrev = document.activeElement;
  root.classList.add('sp-sheet-open');
  rail.full.removeAttribute('inert');
  rail.full.setAttribute('aria-hidden', 'false');
  rail.full.setAttribute('role', 'dialog');
  rail.full.setAttribute('aria-modal', 'true');
  var scrim = document.querySelector('.sp-sheet-scrim');
  if (!scrim) {
    scrim = el('div', { class: 'sp-sheet-scrim', 'aria-hidden': 'true', onclick: closeSheet });
    document.body.appendChild(scrim);
  }
  var first = rail.full.querySelector('[aria-current="page"]') || rail.full.querySelector('a, button');
  if (first) { first.__spNoFocusOpen = true; first.focus({ preventScroll: true }); }
  document.addEventListener('keydown', sheetKeys, true);
}

function closeSheet() {
  if (!root.classList.contains('sp-sheet-open')) return;
  root.classList.remove('sp-sheet-open');
  rail.subOpen = null;
  Array.prototype.forEach.call(rail.full.querySelectorAll('.sp-sub'), function (sub) {
    var b = sub.previousElementSibling;
    if (b) b.setAttribute('aria-expanded', 'false');
    sub.remove();
  });
  rail.full.removeAttribute('role');
  rail.full.removeAttribute('aria-modal');
  if (!railPinned() || mqMobile.matches) { rail.full.setAttribute('inert', ''); rail.full.setAttribute('aria-hidden', 'true'); }
  document.removeEventListener('keydown', sheetKeys, true);
  closeFly(true);
  if (sheetPrev && sheetPrev.isConnected) { try { sheetPrev.focus({ preventScroll: true }); } catch (e) { sheetPrev = null; } }
}

function sheetKeys(e) {
  if (e.key === 'Escape' && !(rail.fly && rail.fly.classList.contains('open'))) { e.preventDefault(); closeSheet(); return; }
  if (e.key !== 'Tab') return;
  var list = Array.prototype.slice.call(rail.full.querySelectorAll('a[href], button')).filter(function (x) { return x.offsetParent !== null; });
  if (rail.fly && rail.fly.classList.contains('open')) list = list.concat(Array.prototype.slice.call(rail.fly.querySelectorAll('a')));
  if (!list.length) return;
  var i = list.indexOf(document.activeElement);
  if (e.shiftKey && i <= 0) { e.preventDefault(); list[list.length - 1].focus(); }
  else if (!e.shiftKey && i === list.length - 1) { e.preventDefault(); list[0].focus(); }
}

var tb = { start: null, end: null, key: '', palette: null, theme: null, seg: null };

function topbar() {
  var bar = document.querySelector('#main > div.sticky:first-child');
  return bar && bar.querySelector('.justify-end') ? bar : null;
}

function nativeCrumbs(bar) {
  var items = bar.querySelectorAll(':scope > div.flex.w-full > div.flex-1 > nav li');
  var out = [];
  Array.prototype.forEach.call(items, function (li) {
    var a = li.querySelector('a[href]');
    var text = (a || li).textContent.replace(/\s*\/\s*$/, '').trim();
    if (!text || text === '/') return;
    out.push({ label: text, href: a ? a.getAttribute('href') : null });
  });
  return out;
}

function crumbsFor(bar) {
  var info = pageInfo();
  var nat = nativeCrumbs(bar);
  if (nat.length >= 2 && !nat[0].href && info) nat[0].href = info.link.href;
  var out = [];
  if (info && info.group && info.group.id !== 'settings') out.push({ label: info.group.label, href: null });
  if (nat.length >= 2) return out.concat(nat);
  if (info) out.push({ label: info.link.label, href: null });
  else if (nat.length) out = out.concat(nat);
  return out;
}

function paintCrumbs(box, crumbs) {
  box.innerHTML = '';
  crumbs.forEach(function (c, i) {
    var last = i === crumbs.length - 1;
    if (i) box.insertAdjacentHTML('beforeend', '<span class="sp-crumb-sep" aria-hidden="true">' + spIcon('chevron', 'sp-ic sp-ic-dir') + '</span>');
    if (last) box.appendChild(el('span', { class: 'sp-crumb-current', 'aria-current': 'page', text: c.label, dir: /^[0-9a-f-]{12,}$/i.test(c.label) ? 'ltr' : null }));
    else if (c.href) box.appendChild(el('a', { class: 'sp-crumb-link', href: c.href, text: c.label }));
    else box.appendChild(el('span', { class: 'sp-crumb-link', text: c.label }));
  });
}

function segIndicator(seg) {
  var on = seg.querySelector('[aria-current="page"], [aria-checked="true"], [aria-selected="true"]');
  var ind = seg.querySelector(':scope > .sp-seg-ind');
  if (!ind) return;
  if (!on) { ind.style.opacity = '0'; return; }
  var sr = seg.getBoundingClientRect();
  var r = on.getBoundingClientRect();
  if (!sr.width || !r.width) return;
  var x = isRtl() ? r.right - sr.right + 4 : r.left - sr.left - 4;
  ind.style.width = '100px';
  ind.style.transform = 'translateX(' + x.toFixed(1) + 'px) scaleX(' + (r.width / 100).toFixed(4) + ')';
  ind.style.opacity = '1';
}

function paintTabs(box, info) {
  var g = info && info.group && (info.group.id === 'economy' || info.group.id === 'publication') ? info.group : null;
  var key = g ? g.id + '|' + JSON.stringify(g.links.map(function (l) { return [l.href, l.label, l.badge]; })) + '|' + info.link.href : '';
  if (box.getAttribute('data-key') === key) { if (g) segIndicator(box); return; }
  var first = !box.getAttribute('data-key') || box.getAttribute('data-group') !== (g && g.id);
  box.setAttribute('data-key', key);
  box.setAttribute('data-group', g ? g.id : '');
  box.innerHTML = '';
  box.hidden = !g;
  if (!g) return;
  box.setAttribute('aria-label', g.label);
  box.appendChild(el('span', { class: 'sp-seg-ind', 'aria-hidden': 'true' }));
  g.links.forEach(function (l) {
    var a = el('a', { href: l.href, class: 'sp-seg-item', 'aria-current': l === info.link ? 'page' : null }, el('span', { text: l.label }));
    if (l.badge) a.appendChild(el('em', { class: 'sp-seg-count', text: spNum(l.badge) }));
    box.appendChild(a);
  });
  box.classList.toggle('sp-seg-still', first);
  requestAnimationFrame(function () {
    segIndicator(box);
    if (first) requestAnimationFrame(function () { box.classList.remove('sp-seg-still'); });
  });
}

function l10n() {
  return window.NakamaL10n || null;
}

function buildLangSeg() {
  var seg = el('div', { class: 'sp-seg sp-seg-lang', role: 'radiogroup', 'data-l10n-skip': '' });
  seg.appendChild(el('span', { class: 'sp-seg-ind', 'aria-hidden': 'true' }));
  var L = l10n();
  var codes = L && L.locales ? Object.keys(L.locales) : ['fa-IR', 'en'];
  codes.forEach(function (code) {
    var info = L && L.locales ? L.locales[code] : { short: code === 'en' ? 'EN' : 'فا', lang: code };
    var on = L ? L.locale === code : (code === 'en') === (spLang() === 'en');
    var b = el('button', { type: 'button', role: 'radio', class: 'sp-seg-item', lang: info.lang, 'aria-checked': on ? 'true' : 'false', 'aria-label': info.name || code, 'data-locale': code, text: info.short || code });
    b.addEventListener('click', function () {
      if (b.getAttribute('aria-checked') === 'true') return;
      Array.prototype.forEach.call(seg.querySelectorAll('[role="radio"]'), function (x) { x.setAttribute('aria-checked', x === b ? 'true' : 'false'); });
      segIndicator(seg);
      setTimeout(function () {
        if (L && L.setLocale) L.setLocale(code);
        else { var sw = document.querySelector('.l10n-switcher:not(.economy-theme)'); if (sw) sw.click(); }
      }, reduced() ? 0 : 200);
    });
    seg.appendChild(b);
  });
  return seg;
}

function themeClick() {
  var api = nc();
  if (api && api.toggleTheme) api.toggleTheme();
  else {
    var next = !root.classList.contains('dark');
    root.classList.toggle('dark', next);
    try { localStorage.setItem('color-mode', next ? 'dark' : 'light'); } catch (e) { next = !!next; }
  }
  paintTheme();
}

function paintTheme() {
  if (!tb.theme) return;
  var dark = root.classList.contains('dark');
  var label = st(dark ? 'top.toLight' : 'top.toDark');
  if (tb.theme.getAttribute('data-mode') === (dark ? 'dark' : 'light') && tb.theme.getAttribute('aria-label') === label) return;
  tb.theme.setAttribute('data-mode', dark ? 'dark' : 'light');
  tb.theme.innerHTML = spIcon(dark ? 'sun' : 'moon');
  tb.theme.setAttribute('aria-label', label);
  tb.theme.setAttribute('title', label);
  tb.theme.setAttribute('aria-pressed', dark ? 'true' : 'false');
}

function flowValue(n) {
  var sr = n && n.shadowRoot;
  if (!sr) return (n && n.textContent || '').trim();
  var out = '';
  Array.prototype.forEach.call(sr.querySelectorAll('[part~="number"] [part~="digit"], [part~="number"] [part~="symbol"]'), function (d) {
    var cur = parseInt(d.style.getPropertyValue('--current'), 10);
    if (d.classList.contains('digit')) out += isNaN(cur) ? '?' : String(cur);
    else out += d.textContent;
  });
  out = out.replace(/\s+/g, '');
  return !out || out.indexOf('?') !== -1 ? '–' : out;
}

function nativeStats() {
  var box = document.querySelector('#main > div.sticky:first-child .hidden.md\\:flex.text-xs');
  if (!box) return [];
  var out = [];
  Array.prototype.forEach.call(box.children, function (c) {
    if (c.classList.contains('x-divider')) return;
    var label = c.querySelector('.text-secondary-400');
    var flow = c.querySelector('number-flow-vue');
    if (label) out.push({ label: label.textContent.trim(), value: flow ? flowValue(flow) : '' });
  });
  return out;
}

var statsTimer = 0;
function paintStats() {
  var pop = tb.stats;
  if (!pop) return;
  var rows = nativeStats();
  var key = JSON.stringify(rows) + spLang();
  if (pop.getAttribute('data-key') === key) return;
  pop.setAttribute('data-key', key);
  pop.innerHTML = '';
  pop.appendChild(el('div', { class: 'sp-stats-head' }, el('span', { class: 'sp-dot', 'aria-hidden': 'true' }), el('b', { text: st('top.status') })));
  var grid = el('dl', { class: 'sp-stats-grid' });
  rows.forEach(function (r) {
    grid.appendChild(el('div', { class: 'sp-stat' }, el('dt', { text: r.label }), el('dd', { text: spNum(r.value) })));
  });
  pop.appendChild(grid);
}

function toggleStats(force) {
  var on = force === undefined ? !root.classList.contains('sp-stats-open') : force;
  root.classList.toggle('sp-stats-open', on);
  var b = tb.end && tb.end.querySelector('.sp-status');
  if (b) b.setAttribute('aria-expanded', on ? 'true' : 'false');
  clearInterval(statsTimer);
  if (on) {
    paintStats();
    placeStats(b);
    statsTimer = setInterval(paintStats, 1000);
  }
}

function placeStats(b) {
  var pop = tb.stats;
  var box = pop && pop.offsetParent;
  if (!b || !box) return;
  var br = box.getBoundingClientRect();
  var pr = b.getBoundingClientRect();
  var start = Math.max(0, Math.round(isRtl() ? br.right - pr.right : pr.left - br.left));
  var room = isRtl() ? br.right - start - 8 : window.innerWidth - 8 - (br.left + start);
  var w = pop.offsetWidth;
  if (w > room) start = Math.max(0, start - Math.ceil(w - room));
  pop.style.insetInlineStart = start + 'px';
}

function buildEnd() {
  var box = el('div', { class: 'sp-tb-end', 'data-l10n-skip': '' });
  var search = el('button', { type: 'button', class: 'sp-search', 'aria-label': st('top.searchPages'), 'aria-haspopup': 'dialog', onclick: function () { openPalette(); } });
  search.innerHTML = spIcon('search');
  search.appendChild(el('span', { class: 'sp-search-ph', text: st('top.search') }));
  search.appendChild(el('span', { class: 'sp-keys', dir: 'ltr' }, el('kbd', { text: 'Ctrl' }), el('kbd', { text: 'K' })));
  var status = el('button', { type: 'button', class: 'sp-status', 'aria-expanded': 'false', 'aria-label': st('top.status'), title: st('top.status'), onclick: function (e) { e.stopPropagation(); toggleStats(); } });
  status.appendChild(el('span', { class: 'sp-dot', 'aria-hidden': 'true' }));
  status.appendChild(el('span', { class: 'sp-status-label', text: st('top.live') }));
  status.setAttribute('aria-controls', 'sp-stats');
  tb.stats = el('div', { class: 'sp-stats', id: 'sp-stats', role: 'region', 'aria-label': st('top.status'), 'data-l10n-skip': '' });
  tb.seg = buildLangSeg();
  tb.theme = el('button', { type: 'button', class: 'sp-round sp-theme', onclick: themeClick });
  paintTheme();
  box.appendChild(search);
  box.appendChild(status);
  box.appendChild(tb.stats);
  box.appendChild(tb.seg);
  box.appendChild(tb.theme);
  return box;
}

function paintTopbar() {
  var bar = topbar();
  if (!bar) return;
  var holder = bar.querySelector(':scope > div.flex.w-full > div.flex-1');
  var end = bar.querySelector('.justify-end');
  if (!holder || !end) return;
  if (!tb.start) {
    tb.start = el('div', { class: 'sp-tb-start', 'data-l10n-skip': '' },
      el('button', { type: 'button', class: 'sp-round sp-menu', 'aria-label': st('top.menu'), html: spIcon('menu'), onclick: function () { openSheet(); } }),
      el('nav', { class: 'sp-crumbs' }),
      el('nav', { class: 'sp-seg sp-tabs', hidden: true }));
  }
  if (!tb.end) tb.end = buildEnd();
  if (tb.start.parentNode !== holder || holder.firstChild !== tb.start) holder.insertBefore(tb.start, holder.firstChild);
  if (tb.end.parentNode !== end || end.firstChild !== tb.end) end.insertBefore(tb.end, end.firstChild);
  var info = pageInfo();
  var crumbs = crumbsFor(bar);
  var key = spLang() + '|' + JSON.stringify(crumbs);
  var cb = tb.start.querySelector('.sp-crumbs');
  cb.setAttribute('aria-label', st('crumb.label'));
  if (tb.key !== key) {
    tb.key = key;
    paintCrumbs(cb, crumbs);
  }
  paintTabs(tb.start.querySelector('.sp-tabs'), info);
  paintTheme();
  var lseg = tb.seg;
  if (lseg && !lseg.__spInit && lseg.offsetWidth) { lseg.__spInit = true; segIndicator(lseg); }
}

function paletteItems(q) {
  var items = allPages();
  q = (q || '').trim().toLowerCase();
  if (!q) return items;
  var rank = function (text) {
    var i = String(text || '').toLowerCase().indexOf(q);
    if (i === -1) return -1;
    return i === 0 ? 0 : /[\s\u200c]/.test(String(text).charAt(i - 1)) ? 1 : 2;
  };
  var byLabel = [];
  var byGroup = [];
  items.forEach(function (it, i) {
    var r = rank(it.label);
    if (r !== -1) byLabel.push({ it: it, r: r, i: i });
    else if (rank(it.group) !== -1) byGroup.push(it);
  });
  byLabel.sort(function (a, b) { return a.r - b.r || a.i - b.i; });
  return byLabel.length ? byLabel.map(function (x) { return x.it; }) : byGroup;
}

function openPalette() {
  if (tb.palette) { tb.palette.querySelector('input').focus(); return; }
  var prev = document.activeElement;
  var input = el('input', { type: 'text', class: 'sp-pal-input', 'aria-label': st('top.searchPages'), placeholder: st('top.searchPages'), autocomplete: 'off', role: 'combobox', 'aria-expanded': 'true', 'aria-controls': 'sp-pal-list' });
  var list = el('div', { class: 'sp-pal-list', id: 'sp-pal-list', role: 'listbox', 'aria-label': st('palette.group.pages') });
  var win = el('div', { class: 'sp-pal', role: 'dialog', 'aria-modal': 'true', 'aria-label': st('top.searchPages') },
    el('div', { class: 'sp-pal-head', html: spIcon('search') }, input, el('kbd', { class: 'sp-esc', text: 'Esc' })),
    list,
    el('div', { class: 'sp-pal-foot', text: st('palette.hint') }));
  var shell = el('div', { class: 'sp-pal-holder', 'data-l10n-skip': '', dir: root.getAttribute('dir') || 'ltr' }, el('div', { class: 'sp-scrim' }), win);
  var sel = 0;
  var rows = [];
  function go(it) {
    close();
    if (it.ps) { var api = nc() && nc().projectSettings; if (api) { api.open(); return; } }
    location.hash = it.href;
  }
  function paint() {
    rows = paletteItems(input.value);
    if (sel >= rows.length) sel = rows.length - 1;
    if (sel < 0) sel = 0;
    list.innerHTML = '';
    if (!rows.length) { list.appendChild(el('div', { class: 'sp-pal-empty', text: st('top.noResults') })); input.removeAttribute('aria-activedescendant'); return; }
    rows.slice(0, 40).forEach(function (it, i) {
      var r = el('a', { href: it.href, id: 'sp-pal-' + i, role: 'option', class: 'sp-pal-row', 'aria-selected': i === sel ? 'true' : 'false' });
      r.innerHTML = it.svg || spIcon(it.psSection ? 'sliders' : 'page');
      r.appendChild(el('span', { class: 'sp-pal-label', text: it.label }));
      if (it.group) r.appendChild(el('span', { class: 'sp-pal-group', text: it.group }));
      r.addEventListener('click', function (e) { if (e.metaKey || e.ctrlKey) return; e.preventDefault(); go(it); });
      r.addEventListener('mousemove', function () { if (sel !== i) { sel = i; mark(); } });
      list.appendChild(r);
    });
    mark();
  }
  function mark() {
    Array.prototype.forEach.call(list.querySelectorAll('[role="option"]'), function (r, i) { r.setAttribute('aria-selected', i === sel ? 'true' : 'false'); });
    input.setAttribute('aria-activedescendant', 'sp-pal-' + sel);
    var cur = list.querySelector('[aria-selected="true"]');
    if (cur && cur.scrollIntoView) cur.scrollIntoView({ block: 'nearest' });
  }
  function close() {
    if (!tb.palette) return;
    var p = tb.palette;
    tb.palette = null;
    document.removeEventListener('keydown', keys, true);
    p.classList.add('sp-closing');
    setTimeout(function () { p.remove(); }, reduced() ? 0 : 160);
    if (prev && prev.isConnected) { try { prev.focus({ preventScroll: true }); } catch (e) { prev = null; } }
  }
  function keys(e) {
    if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); close(); }
    else if (e.key === 'ArrowDown') { e.preventDefault(); sel = Math.min(rows.length - 1, sel + 1); mark(); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); sel = Math.max(0, sel - 1); mark(); }
    else if (e.key === 'Enter' && rows[sel]) { e.preventDefault(); go(rows[sel]); }
    else if (e.key === 'Tab') { e.preventDefault(); input.focus(); }
  }
  input.addEventListener('input', function () { sel = 0; paint(); });
  shell.addEventListener('mousedown', function (e) { if (e.target === shell || e.target.classList.contains('sp-scrim')) close(); });
  document.addEventListener('keydown', keys, true);
  tb.palette = shell;
  tb.palette.close = close;
  document.body.appendChild(shell);
  paint();
  input.focus();
}

function globalKeys(e) {
  if ((e.ctrlKey || e.metaKey) && !e.altKey && (e.key === 'k' || e.key === 'K')) {
    if (root.classList.contains('sp-auth')) return;
    e.preventDefault();
    openPalette();
  } else if (e.key === 'Escape' && root.classList.contains('sp-stats-open')) {
    toggleStats(false);
  }
}

document.addEventListener('click', function (e) {
  if (!root.classList.contains('sp-stats-open')) return;
  if (e.target.closest && e.target.closest('.sp-status, .sp-stats')) return;
  toggleStats(false);
}, true);

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
  if (!main) return;
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

function activeTab(bar) {
  return bar.querySelector('.x-tab.router-link-exact-active') || bar.querySelector('.x-tab[aria-current="page"], .x-tab[aria-selected="true"], .x-tab.eco-tab-on') || bar.querySelector('.x-tab.router-link-active');
}

function paintNativeTabs() {
  var bars = document.querySelectorAll('#main .x-tab-group .relative.flex.min-w-full.w-fit.border-b, [data-economy] .x-tab-group .relative.flex.min-w-full.w-fit.border-b');
  for (var i = 0; i < bars.length; i++) {
    var bar = bars[i];
    var ind = bar.querySelector(':scope > .sp-tabind');
    var fresh = false;
    if (!ind) {
      ind = el('span', { class: 'sp-tabind sp-still', 'aria-hidden': 'true' });
      bar.appendChild(ind);
      fresh = true;
    }
    var on = activeTab(bar);
    if (!on) { ind.style.opacity = '0'; continue; }
    var br = bar.getBoundingClientRect();
    var r = on.getBoundingClientRect();
    if (!r.width) continue;
    var x = r.left - br.left + bar.scrollLeft;
    var t = 'translateX(' + x.toFixed(1) + 'px) scaleX(' + (r.width / 100).toFixed(4) + ')';
    if (ind.style.transform !== t) ind.style.transform = t;
    ind.style.opacity = '1';
    if (fresh) requestAnimationFrame(function (n) { return function () { requestAnimationFrame(function () { n.classList.remove('sp-still'); }); }; }(ind));
  }
}

function readNumber(text) {
  var s = String(text || '').replace(/[۰-۹]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'.indexOf(d); }).replace(/[٬,\s]/g, '').replace('٫', '.');
  return /^-?\d+(\.\d+)?$/.test(s) ? parseFloat(s) : NaN;
}

function countKpis() {
  var nodes = document.querySelectorAll('#economy-root .eco-stat-grid .text-3xl');
  if (!nodes.length) return;
  var nf = new Intl.NumberFormat(spLang() === 'fa' ? 'fa-IR' : 'en-US', { maximumFractionDigits: 0 });
  for (var i = 0; i < nodes.length; i++) {
    var n = nodes[i];
    if (n.__spCounted) continue;
    n.__spCounted = true;
    var v = readNumber(n.textContent);
    if (isNaN(v) || v === 0) continue;
    var final = n.textContent;
    countUp(n, v, function (x) { return nf.format(Math.round(x)); });
    setTimeout(function (node, text) { return function () { if (node.isConnected) node.textContent = text; }; }(n, final), reduced() ? 0 : 560);
  }
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
  var moved = (to && to.fullPath) !== (from && from.fullPath);
  if (moved && typeof closeFly === 'function') {
    closeFly(true);
    closeSheet();
    toggleStats(false);
  }
  if (a !== b && a !== 'ps' && b !== 'ps') routeEnter();
  schedule();
}

function countUp(node, to, fmt) {
  if (reduced() || !isFinite(to)) return;
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
    paintNativeTabs();
    countKpis();
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
      if (t.nodeType === 1 && t.closest && t.closest('.sp-tabind, .sp-rail, .sp-rail-full, .sp-flyout, .sp-tb-start, .sp-tb-end, .sp-sky, .sp-tip, .sp-pal-holder, [_echarts_instance_], number-flow-vue')) continue;
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
