var ROOT_ID = 'economy-root';
var NAV_LINK = 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent dark:border-transparent';
var NAV_ACTIVE = '!border-primary-600 !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700';
var rendered = null;
var router = null;
var pendingHash = null;
var EmptyPage = { name: 'EconomyPage', render: function () { return null; } };
var TAB_ON = 'text-[color:var(--x-tab-group-text)] dark:text-[color:var(--x-tab-group-dark-text)]';
var TAB_VARS = {
  off: '--x-link-text: #475569; --x-link-text-hover: #334155; --x-link-shadow: #f1f5f9; --x-link-dark-text: #94a3b8; --x-link-dark-text-hover: #cbd5e1; --x-link-dark-shadow: #0f172a;',
  on: '--x-link-text: rgb(102,84,226); --x-link-text-hover: rgb(93,77,206); --x-link-shadow: rgb(233,232,252); --x-link-dark-text: rgb(130,117,240); --x-link-dark-text-hover: rgb(181,175,245); --x-link-dark-shadow: rgb(59,49,130);'
};

function captureRouter(app) {
  if (router) return;
  try {
    var r = app && app.config && app.config.globalProperties && app.config.globalProperties.$router;
    if (!r || typeof r.addRoute !== 'function') return;
    var layout = r.resolve('/').matched[0];
    if (!layout || !layout.components) return;
    SECTION_ORDER.forEach(function (sec) {
      if (r.hasRoute(sec)) return;
      r.addRoute({ path: '/' + sec + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: sec, component: EmptyPage, meta: { title: SECTION_TITLE[sec] } }] });
    });
    Object.keys(MOVED_PAGES).forEach(function (from) {
      var name = 'moved-' + from.replace('/', '-');
      if (!r.hasRoute(name)) r.addRoute({ path: '/' + from + '/:rest(.*)*', name: name, redirect: function (to) { return to.fullPath.replace('/' + from, '/' + MOVED_PAGES[from]); } });
    });
    if (r.hasRoute('settings')) {
      SETTINGS_TABS.forEach(function (tab) {
        if (!r.hasRoute('settings-' + tab)) r.addRoute('settings', { path: tab, name: 'settings-' + tab, component: EmptyPage, meta: tabMeta(tab) });
        if (RESOURCES[tab] && !r.hasRoute('settings-' + tab + '-item')) r.addRoute('settings', { path: tab + '/:id(.*)', name: 'settings-' + tab + '-item', component: EmptyPage, meta: tabMeta(tab) });
      });
      Object.keys(LEGACY_HASHES).forEach(function (sec) {
        Object.keys(LEGACY_HASHES[sec]).forEach(function (page) {
          var name = 'legacy-' + sec + '-' + page;
          if (!r.hasRoute(name)) r.addRoute({ path: '/' + sec + '/' + page + '/:rest(.*)*', name: name, redirect: hashFor(LEGACY_HASHES[sec][page]).slice(1) });
        });
      });
    }
    router = r;
  } catch (e) {
    router = null;
    return;
  }
  if (pendingHash) {
    var target = pendingHash;
    pendingHash = null;
    router.replace(target.slice(1));
  }
}

var SECTION_TITLE = { base: 'General', economy: 'Economy', publication: 'Publishing' };

function movedHash(hash) {
  var m = /^#\/([^/?]+\/[^/?]+)(.*)$/.exec(hash || '');
  return m && MOVED_PAGES[m[1]] ? '#/' + MOVED_PAGES[m[1]] + m[2] : null;
}

function tabMeta(tab) {
  return { get title() { return t('settings') + ' - ' + t('tab.' + tab); } };
}

function legacyHash(hash) {
  var moved = movedHash(hash);
  if (moved) return moved;
  var m = /^#\/(economy|publication)\/([^/?]+)/.exec(hash);
  var to = m && LEGACY_HASHES[m[1]][m[2]];
  return to ? hashFor(to) : null;
}

function trapApp() {
  var proto = Element.prototype;
  if (Object.getOwnPropertyDescriptor(proto, '__vue_app__')) return;
  Object.defineProperty(proto, '__vue_app__', {
    configurable: true,
    get: function () { return undefined; },
    set: function (app) {
      Object.defineProperty(this, '__vue_app__', { value: app, writable: true, configurable: true });
      delete proto.__vue_app__;
      captureRouter(app);
    }
  });
}

function replaceHash(hash) {
  if (rendered) rendered.key = keyOf(parseRoute(hash));
  if (router) router.replace(hash.slice(1));
  else history.replaceState(history.state, '', location.pathname + location.search + hash);
}

function parseRoute(hash) {
  var m = /^#\/(base|economy|publication|settings)(?:\/([^/?]*))?(?:\/([^?]*))?(?:\?(.*))?$/.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  var section = m[1];
  if (section === 'settings') {
    if (SETTINGS_TABS.indexOf(m[2]) === -1) return null;
    if (!RESOURCES[m[2]]) return { section: section, page: m[2], id: null, query: {} };
  }
  var page = section === 'settings' ? m[2] : m[2] || SECTIONS[section][0];
  if (section !== 'settings' && SECTIONS[section].indexOf(page) === -1) page = SECTIONS[section][0];
  var query = {};
  (m[4] || '').split('&').forEach(function (pair) {
    if (!pair) return;
    var i = pair.indexOf('=');
    var k = decodeURIComponent(i === -1 ? pair : pair.slice(0, i));
    var v = i === -1 ? '' : decodeURIComponent(pair.slice(i + 1).replace(/\+/g, ' '));
    query[k] = v;
  });
  return { section: section, page: page, id: m[3] ? decodeURIComponent(m[3]) : null, query: query };
}

function keyOf(r) {
  if (!r) return '';
  var detail = (r.page === 'payments' || r.page === 'notifications') && r.id ? '/' + r.id : '';
  return lang() + '|' + r.page + detail + '|' + JSON.stringify(r.query);
}

function navIcon(name) {
  var svg = icon(name, 'h-4 w-4');
  return svg;
}

function buildNav(section, collapsed, active) {
  var ul = h('ul', { class: 'space-y-0.5' });
  SECTIONS[section].forEach(function (p) {
    var on = p === active;
    var a = h('a', { href: hashFor(p, null, p === active ? null : lastQuery[p]), class: (on ? NAV_ACTIVE + ' ' : '') + NAV_LINK, 'aria-current': on ? 'page' : null, title: collapsed ? t('nav.' + p) : null, 'data-economy-page': p });
    a.appendChild(collapsed ? h('div', { class: 'x-tooltip flex' }, navIcon(p)) : navIcon(p));
    a.appendChild(h('div', { class: 'font-medium relative truncate' + (collapsed ? '' : ' flex-1') }, h('span', { text: collapsed ? ' ' : t('nav.' + p) })));
    var badge = navBadge(p, collapsed);
    if (badge) a.appendChild(badge);
    ul.appendChild(h('li', {}, a));
  });
  var title = section === 'economy' ? 'section' : section === 'base' ? 'section.base' : 'section.pub';
  return h('div', { 'data-economy-nav': section, 'data-l10n-skip': '', class: section === SECTION_ORDER[0] ? null : 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 ' + (collapsed ? 'text-[8px] text-center' : 'px-3 text-[10px]'), text: collapsed ? t(title + '.short') : t(title) }),
    h('nav', { 'aria-label': t(title) }, ul));
}

function placeNav(prev, section, collapsed, active, key) {
  var next = prev.nextElementSibling;
  var existing = next && next.getAttribute('data-economy-nav') === section ? next : null;
  if (!existing || existing.getAttribute('data-key') !== key) {
    var built = buildNav(section, collapsed, active);
    built.setAttribute('data-key', key);
    if (existing) existing.replaceWith(built);
    else prev.after(built);
    existing = built;
  }
  return existing;
}

function isCollapsed(anchor) {
  var drawer = anchor.closest('.x-drawer');
  if (drawer && drawer.getBoundingClientRect().width > 0 && drawer.getBoundingClientRect().width < 100) return true;
  var span = anchor.querySelector('span');
  return !!span && span.textContent === ' ';
}

function syncSidebar(route) {
  var active = route ? route.page : null;
  var anchors = document.querySelectorAll('a[href="#/players"]');
  var keep = [];
  for (var i = 0; i < anchors.length; i++) {
    var a = anchors[i];
    if (!a.closest('.x-drawer') || a.closest('#' + ROOT_ID)) continue;
    var nav = a.closest('nav');
    var group = nav && nav.parentElement;
    if (!group || !group.parentElement) continue;
    var collapsed = isCollapsed(a);
    var key = lang() + '|' + collapsed + '|' + active + '|' + pendingComments + '|' + JSON.stringify(lastQuery);
    var prev = group;
    for (var si = 0; si < SECTION_ORDER.length; si++) {
      prev = placeNav(prev, SECTION_ORDER[si], collapsed, active, key);
      keep.push(prev);
    }
    refreshBadge(false);
    var vueActive = group.parentElement.querySelectorAll('a[aria-current="page"]:not([data-economy-page])');
    for (var j = 0; j < vueActive.length && active; j++) {
      if (vueActive[j].getAttribute('href') === '#/') vueActive[j].setAttribute('data-economy-dimmed', '');
    }
  }
  var all = document.querySelectorAll('[data-economy-nav]');
  for (var k = 0; k < all.length; k++) if (keep.indexOf(all[k]) === -1) all[k].remove();
  document.documentElement.toggleAttribute('data-economy-active', !!active);
}

function showMains() {
  var hidden = document.querySelectorAll('[data-economy-hidden]');
  for (var i = 0; i < hidden.length; i++) {
    hidden[i].style.display = '';
    hidden[i].removeAttribute('data-economy-hidden');
  }
}

function unmount() {
  var root = document.getElementById(ROOT_ID);
  if (root) root.remove();
  showMains();
  closePanel();
  rendered = null;
}

function settingsTabRow() {
  var native = document.querySelector('#main > main .x-tab-group a.x-tab[href^="#/settings/"]:not([data-settings-tab])');
  return native ? native.parentElement : null;
}

function settingsTab(base, tab, on) {
  var a = base.cloneNode(true);
  a.removeAttribute('aria-current');
  a.classList.remove('router-link-active', 'router-link-exact-active');
  TAB_ON.split(' ').forEach(function (c) { a.classList.remove(c); });
  if (on) TAB_ON.split(' ').forEach(function (c) { a.classList.add(c); });
  a.classList.add('eco-tab');
  a.classList.toggle('eco-tab-on', on);
  a.setAttribute('href', hashFor(tab));
  a.setAttribute('aria-selected', on ? 'true' : 'false');
  if (on) a.setAttribute('aria-current', 'page');
  a.setAttribute('data-value', location.origin + location.pathname + hashFor(tab));
  a.setAttribute('data-settings-tab', tab);
  a.setAttribute('data-l10n-skip', '');
  a.setAttribute('style', on ? TAB_VARS.on : TAB_VARS.off);
  var label = a.querySelector('.font-medium');
  if (label) label.textContent = t('tab.' + tab);
  else a.textContent = t('tab.' + tab);
  return a;
}

function syncSettingsTabs(route) {
  var row = settingsTabRow();
  if (!row) return null;
  var active = route && route.section === 'settings' ? route.page : null;
  var key = lang() + '|' + active;
  var ours = row.querySelectorAll('[data-settings-tab]');
  var natives = row.querySelectorAll('a.x-tab:not([data-settings-tab])');
  var last = natives[natives.length - 1];
  var inOrder = ours.length === SETTINGS_TABS.length && last.nextElementSibling === ours[0];
  if (inOrder && ours[0].getAttribute('data-key') === key) return row;
  for (var i = 0; i < ours.length; i++) ours[i].remove();
  var base = natives[0];
  var anchor = last;
  SETTINGS_TABS.forEach(function (tab) {
    var a = settingsTab(base, tab, tab === active);
    a.setAttribute('data-key', key);
    anchor.after(a);
    anchor = a;
  });
  return row;
}

function settingsSlot(row) {
  var group = row && row.closest('.x-tab-group');
  var box = group && group.parentElement && group.parentElement.parentElement;
  var slot = box && box.nextElementSibling;
  return slot && slot.tagName === 'DIV' ? slot : null;
}

function render(root, route) {
  var key = keyOf(route);
  var ctl = rendered && rendered.root === root ? rendered : null;
  if (!ctl || ctl.key !== key) {
    closePanel();
    window.scrollTo(0, 0);
    var scroller = document.getElementById('main');
    if (scroller) scroller.scrollTop = 0;
    ctl = { root: root, key: key, id: undefined, list: null };
    rendered = ctl;
    var ctx = { id: route.id, query: route.query };
    if (route.page === 'comments') refreshBadge(true);
    if (route.page === 'overview') overviewPage(root);
    else if (route.page === 'payment-methods') paymentMethodsPage(root);
    else if (route.page === 'sms' || route.page === 'email') providersPage(root, route.page);
    else if (route.page === 'languages') ctl.list = languagesPage(root, route.query, ctx);
    else if (route.page === 'payments' && route.id) paymentDetail(root, route.id);
    else if (route.page === 'notifications' && route.id === 'new') composePage(root, route.query);
    else if (route.page === 'notifications' && route.id) notificationDetail(root, route.id);
    else ctl.list = listPage(root, route.page, route.query, ctx);
  }
  if (RESOURCES[route.page] && ctl.id !== route.id) {
    ctl.id = route.id;
    if (route.id) openEditor(route.page, route.id, route.query, function () { if (ctl.list) ctl.list.reload(); });
    else closePanel();
  }
}

function sync() {
  dockTheme();
  var hash = location.hash;
  var moved = legacyHash(hash);
  if (moved) {
    hash = moved;
    if (router) router.replace(moved.slice(1));
    else history.replaceState(history.state, '', location.pathname + location.search + moved);
  }
  var route = parseRoute(hash);
  if (route && !LANGS.tried && token()) {
    LANGS.tried = true;
    loadLanguages().then(function () { if (LANGS.loaded && LANGS.def !== 'fa' && rendered) { rendered.key = ''; schedule(); } });
    loadKeyTypes();
  }
  var settings = route && route.section === 'settings';
  syncSidebar(settings ? null : route);
  var row = syncSettingsTabs(route);
  if (!route) { if (rendered || document.getElementById(ROOT_ID)) unmount(); return; }
  var root = document.getElementById(ROOT_ID);
  if (settings) {
    if (root && root.getAttribute('data-economy') !== 'settings') {
      root.remove();
      closePanel();
      root = null;
    }
    showMains();
    var slot = settingsSlot(row);
    if (!slot) return;
    if (!root) {
      root = h('div', { id: ROOT_ID, class: 'flex flex-col flex-1 min-w-0', 'data-l10n-skip': '', 'data-economy': 'settings' });
      rendered = null;
    }
    if (root.parentElement !== slot) slot.appendChild(root);
    render(root, route);
    return;
  }
  if (root && root.getAttribute('data-economy') !== 'root') {
    root.remove();
    root = null;
    rendered = null;
  }
  var main = document.querySelector('#main > main');
  if (!main) return;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') {
      mains[i].style.display = 'none';
      mains[i].setAttribute('data-economy-hidden', '');
    }
  }
  if (!root) {
    root = h('div', { id: ROOT_ID, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-economy': 'root' });
    rendered = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  render(root, route);
}

var THEME_KEY = 'color-mode';
var THEME_ICONS = {
  moon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M21.5 14.1A9.5 9.5 0 0 1 9.9 2.5a9.5 9.5 0 1 0 11.6 11.6z"/></svg>',
  sun: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" aria-hidden="true"><circle cx="12" cy="12" r="4.5" fill="currentColor"/><path d="M12 1.5v2.5M12 20v2.5M4.6 4.6l1.8 1.8M17.6 17.6l1.8 1.8M1.5 12H4M20 12h2.5M4.6 19.4l1.8-1.8M17.6 6.4l1.8-1.8"/></svg>'
};
var themeButton = null;

function storedTheme() {
  try {
    var v = localStorage.getItem(THEME_KEY);
    if (v === 'dark' || v === 'light') return v;
  } catch (e) {}
  return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function isDarkTheme() {
  return document.documentElement.classList.contains('dark');
}

function paintThemeButton() {
  if (!themeButton) return;
  var dark = isDarkTheme();
  var label = t(dark ? 'theme.toLight' : 'theme.toDark');
  themeButton.innerHTML = dark ? THEME_ICONS.sun : THEME_ICONS.moon;
  themeButton.setAttribute('title', label);
  themeButton.setAttribute('aria-label', label);
  themeButton.setAttribute('aria-pressed', dark ? 'true' : 'false');
}

function toggleTheme() {
  var next = isDarkTheme() ? 'light' : 'dark';
  try { localStorage.setItem(THEME_KEY, next); } catch (e) {}
  document.documentElement.classList.toggle('dark', next === 'dark');
  paintThemeButton();
}

function dockTheme() {
  if (!themeButton) {
    themeButton = document.createElement('button');
    themeButton.type = 'button';
    themeButton.className = 'l10n-switcher economy-theme';
    themeButton.setAttribute('data-l10n-skip', '');
    themeButton.addEventListener('click', toggleTheme);
    paintThemeButton();
  }
  var switcher = document.querySelector('.l10n-switcher:not(.economy-theme)');
  if (switcher) {
    if (themeButton.nextElementSibling !== switcher) switcher.before(themeButton);
    themeButton.classList.toggle('l10n-switcher--docked', switcher.classList.contains('l10n-switcher--docked'));
  } else if (!themeButton.parentElement) {
    document.body.appendChild(themeButton);
  }
}

document.documentElement.classList.toggle('dark', storedTheme() === 'dark');

var scheduled = false;
function schedule() {
  if (scheduled) return;
  scheduled = true;
  requestAnimationFrame(function () {
    scheduled = false;
    sync();
  });
}

function boot() {
  var app = document.getElementById('app');
  if (!router && app && app.__vue_app__) captureRouter(app.__vue_app__);
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    var el = document.getElementById('app');
    if (!router && el && el.__vue_app__) captureRouter(el.__vue_app__);
    if (router || waited >= 5000) {
      clearInterval(poll);
      if (pendingHash) { location.hash = pendingHash; pendingHash = null; }
    }
  }, 100);
  window.addEventListener('hashchange', sync);
  window.addEventListener('popstate', schedule);
  new MutationObserver(schedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { if (rendered) rendered.key = ''; paintThemeButton(); schedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  sync();
}

if (/^#\/(base|economy|publication)/.test(location.hash) || parseRoute(location.hash)) {
  pendingHash = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
trapApp();
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot);
else boot();
