var ROOT_ID = 'economy-root';
var NAV_LINK = 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent dark:border-transparent';
var NAV_ACTIVE = '!text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700';
var rendered = null;
var router = null;
var pendingHash = null;
var EmptyPage = { name: 'EconomyPage', render: function () { return null; } };

function captureRouter(app) {
  if (router) return;
  try {
    var r = app && app.config && app.config.globalProperties && app.config.globalProperties.$router;
    if (!r || typeof r.addRoute !== 'function') return;
    var layout = r.resolve('/').matched[0];
    if (!layout || !layout.components) return;
    SECTION_ORDER.forEach(function (sec) {
      if (r.hasRoute(sec)) return;
      r.addRoute({ path: '/' + sec + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: sec, component: EmptyPage, meta: { get title() { return sectionTitle(sec); } } }] });
    });
    if (!r.hasRoute(PS_ROUTE)) r.addRoute({ path: '/' + PS_ROUTE + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: PS_ROUTE, component: EmptyPage, meta: { get title() { return psTitle(); } } }] });
    Object.keys(PS_LEGACY).forEach(function (from) {
      var name = 'legacy-' + from.replace('/', '-');
      if (!r.hasRoute(name)) r.addRoute({ path: '/' + from + '/:rest(.*)*', name: name, redirect: function (to) { return to.fullPath.replace('/' + from, '/' + PS_ROUTE + '/' + PS_LEGACY[from]); } });
    });
    router = r;
    r.beforeEach(satoriGuard);
    r.afterEach(function (to, from) { fireRoute(to, from); });
  } catch (e) {
    router = null;
    return;
  }
  try { window.dispatchEvent(new CustomEvent('nakama-console:ready')); } catch (e) { router = router || null; }
  if (pendingHash) {
    var target = pendingHash;
    pendingHash = null;
    router.replace(target.slice(1));
  }
}

function sectionTitle(sec) {
  var r = parseRoute(location.hash);
  var page = r && r.section === sec ? r.page : null;
  return t(SECTION_LABEL[sec]) + (page ? ' - ' + t('nav.' + page) : '');
}

function sectionPageTitle(v) {
  var r = parseRoute(location.hash);
  if (!r || r.section === PS_ROUTE || !SECTIONS[r.section]) return v;
  var parts = String(v == null ? '' : v).split(' | ');
  if (parts.length < 2) return v;
  parts[0] = sectionTitle(r.section);
  return parts.join(' | ');
}

function psTitle() {
  var r = parseRoute(location.hash);
  var sec = r && psSection(r.page);
  return t('ps.title') + (sec ? ' - ' + t(sec.label) : '');
}

function legacyHash(hash) {
  var m = /^#\/([^/?]+\/[^/?]+)(.*)$/.exec(hash || '');
  return m && PS_LEGACY[m[1]] ? '#/' + PS_ROUTE + '/' + PS_LEGACY[m[1]] + m[2] : null;
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
  var r = parseRoute(hash);
  if (r && r.section === PS_ROUTE) { if (ps && ps.ctl) ps.ctl.key = keyOf(r); }
  else if (rendered) rendered.key = keyOf(r);
  if (router) router.replace(hash.slice(1));
  else history.replaceState(history.state, '', location.pathname + location.search + hash);
}

function parseRoute(hash) {
  var m = /^#\/(economy|publication|moderation|support|system|project-settings)(?:\/([^/?]*))?(?:\/([^?]*))?(?:\?(.*))?$/.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  var section = m[1];
  var page;
  if (section === PS_ROUTE) page = psSection(m[2]) ? m[2] : null;
  else {
    page = m[2] || SECTIONS[section][0];
    if (SECTIONS[section].indexOf(page) === -1) page = SECTIONS[section][0];
  }
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
  var detail = (r.page === 'payments' || r.page === 'notifications' || r.page === 'reports' || r.page === 'tickets' || r.page === 'live-chat' || r.page === 'remote-config' || r.page === 'referrals') && r.id ? '/' + r.id : '';
  return lang() + '|' + r.page + detail + '|' + JSON.stringify(r.query);
}

function navPages(section) {
  return SECTIONS[section].filter(function (p) { var need = PAGE_ACL[p] || SECTION_ACL[section]; return !need || aclCan(need, 0); });
}

function buildNav(section, collapsed, active) {
  var ul = h('ul', { class: 'space-y-0.5' });
  navPages(section).forEach(function (p) {
    var on = p === active;
    var a = h('a', { href: hashFor(p, null, p === active ? null : lastQuery[p]), class: (on ? NAV_ACTIVE + ' ' : '') + NAV_LINK, 'aria-current': on ? 'page' : null, title: collapsed ? t('nav.' + p) : null, 'data-economy-page': p });
    a.appendChild(collapsed ? h('div', { class: 'x-tooltip flex' }, icon(p, 'h-4 w-4')) : icon(p, 'h-4 w-4'));
    a.appendChild(h('div', { class: 'font-medium relative truncate' + (collapsed ? '' : ' flex-1') }, h('span', { text: collapsed ? ' ' : t('nav.' + p) })));
    var badge = navBadge(p, collapsed);
    if (badge) a.appendChild(badge);
    ul.appendChild(h('li', {}, a));
  });
  var title = SECTION_LABEL[section];
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
  return !!span && span.textContent === ' ';
}

function syncSidebar(route, psOn) {
  var active = route ? route.page : null;
  var anchors = [];
  var hrefs = ['#/players', '#/chat', '#/notifications', '#/groups', '#/storage', '#/'];
  for (var hi = 0; hi < hrefs.length && !anchors.length; hi++) {
    anchors = Array.prototype.filter.call(document.querySelectorAll('.x-drawer a[href="' + hrefs[hi] + '"]'), function (x) { return !!x.closest('nav') && !x.closest('[data-economy-nav]'); });
  }
  var keep = [];
  for (var i = 0; i < anchors.length; i++) {
    var a = anchors[i];
    var drawer = a.closest('.x-drawer');
    if (!drawer || a.closest('#' + ROOT_ID)) continue;
    var nav = a.closest('nav');
    var group = nav && nav.parentElement;
    if (!group || !group.parentElement) continue;
    var collapsed = isCollapsed(a);
    var key = lang() + '|' + collapsed + '|' + active + '|' + pendingComments + '|' + pendingReports + '|' + supportBadgeKey() + '|' + updatesBadge() + '|' + JSON.stringify(lastQuery) + '|' + token();
    var prev = group;
    for (var si = 0; si < SECTION_ORDER.length; si++) {
      if (!navPages(SECTION_ORDER[si]).length) continue;
      prev = placeNav(prev, SECTION_ORDER[si], collapsed, active, key);
      keep.push(prev);
    }
    var psNav = placePsNav(drawer, collapsed, !!psOn, lang() + '|' + collapsed + '|' + !!psOn + '|' + psEntryHash());
    if (psNav) keep.push(psNav);
    refreshBadge(false);
    refreshReportBadge(false);
    refreshSupportBadge(false);
    refreshUpdatesBadge(false);
    var vueActive = group.parentElement.querySelectorAll('a[aria-current="page"]:not([data-economy-page])');
    for (var j = 0; j < vueActive.length && (active || psOn); j++) {
      if (vueActive[j].getAttribute('href') === '#/') vueActive[j].setAttribute('data-economy-dimmed', '');
    }
  }
  var all = document.querySelectorAll('[data-economy-nav],[data-ps-nav]');
  for (var k = 0; k < all.length; k++) if (keep.indexOf(all[k]) === -1) all[k].remove();
  document.documentElement.toggleAttribute('data-economy-active', !!active || !!psOn);
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

function mountPage(root, route, ctl) {
  var ps2 = psSection(route.page);
  var ctx = { id: route.id, query: route.query };
  if (ps2) { ctl.list = ps2.render(root, route, ctx) || null; return; }
  if (route.page === 'comments') refreshBadge(true);
  if (route.page === 'reports') refreshReportBadge(true);
  if (route.page === 'tickets' || route.page === 'live-chat') { refreshSupportBadge(true); supportBeat(); }
  if (route.page === 'updates') { refreshUpdatesBadge(true); updatesPage(root); }
  else if (route.page === 'overview') overviewPage(root);
  else if (route.page === 'tickets' && route.id) ticketDetail(root, route.id);
  else if (route.page === 'tickets') ctl.list = ticketsPage(root, route.query, ctx);
  else if (route.page === 'live-chat') liveChatPage(root, route.id, route.query);
  else if (route.page === 'activity') ctl.list = activityPage(root, route.query, ctx);
  else if (route.page === 'referrals') ctl.list = referralsPage(root, route.query, ctx);
  else if (route.page === 'remote-config' && route.id) configEditor(root, route.id, route.query);
  else if (route.page === 'remote-config') ctl.list = configPage(root, route.query, ctx);
  else if (route.page === 'reports' && route.id) reportDetail(root, route.id);
  else if (route.page === 'players') ctl.list = playersPage(root, route.query, ctx);
  else if (route.page === 'restrictions') ctl.list = restrictionsPage(root, route.query, ctx);
  else if (route.page === 'chat-words') ctl.list = chatWordsPage(root, route.query, ctx);
  else if (route.page === 'payments' && route.id) paymentDetail(root, route.id);
  else if (route.page === 'notifications' && route.id === 'new') composePage(root, route.query);
  else if (route.page === 'notifications' && route.id) notificationDetail(root, route.id);
  else ctl.list = listPage(root, route.page, route.query, ctx);
}

function pageCtl(prev, root, route, onNew) {
  var key = keyOf(route);
  var ctl = prev && prev.root === root ? prev : null;
  if (!ctl || ctl.key !== key) {
    closePanel();
    ctl = { root: root, key: key, id: undefined, list: null };
    onNew(ctl);
    mountPage(root, route, ctl);
  }
  if (RESOURCES[route.page] && ctl.id !== route.id) {
    ctl.id = route.id;
    var mine = ctl;
    if (route.id) openEditor(route.page, route.id, route.query, function () { if (mine.list) mine.list.reload(); });
    else closePanel();
  }
  return ctl;
}

function render(root, route) {
  pageCtl(rendered, root, route, function (ctl) {
    window.scrollTo(0, 0);
    var scroller = document.getElementById('main');
    if (scroller) scroller.scrollTop = 0;
    rendered = ctl;
  });
}

function sync() {
  dockTheme();
  applyBrand();
  maybeClaimSetup();
  nativeTweaks();
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
    loadLanguages().then(function () {
      if (!LANGS.loaded || LANGS.def === 'fa') return;
      if (rendered) rendered.key = '';
      if (ps && ps.ctl) ps.ctl.key = '';
      schedule();
    });
    loadKeyTypes();
  }
  if (route && route.section === PS_ROUTE) {
    syncSidebar(null, true);
    psShow(route);
    dockPsButton();
    return;
  }
  if (ps) psHide();
  psRemember(hash || '#/');
  syncSidebar(route, false);
  dockPsButton();
  if (!route) { if (rendered || document.getElementById(ROOT_ID)) unmount(); return; }
  var root = document.getElementById(ROOT_ID);
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
  new MutationObserver(function () { if (rendered) rendered.key = ''; if (ps) { ps.listKey = null; if (ps.ctl) ps.ctl.key = ''; } paintThemeButton(); schedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  sync();
}

var routeHooks = [notFoundHook];

function fireRoute(to, from) {
  for (var i = 0; i < routeHooks.length; i++) {
    try { routeHooks[i](to, from); } catch (e) { continue; }
  }
}

function consoleSections() {
  return SECTION_ORDER.filter(function (sec) { return navPages(sec).length > 0; }).map(function (sec) {
    return {
      id: sec,
      label: t(SECTION_LABEL[sec]),
      pages: navPages(sec).map(function (p) { return { id: p, href: hashFor(p), label: t('nav.' + p), badge: p === 'comments' ? pendingComments || 0 : p === 'reports' ? pendingReports || 0 : p === 'updates' ? updatesBadge() : supportBadge(p) }; })
    };
  });
}

window.NakamaConsole = {
  version: 1,
  get router() { return router; },
  lang: function () { return lang(); },
  t: function (key, vars) { return t(key, vars); },
  sections: consoleSections,
  badges: function () { return { comments: pendingComments || 0, reports: pendingReports || 0, tickets: supportBadge('tickets'), chats: supportBadge('live-chat'), updates: updatesBadge() }; },
  updates: function () { var s = UPDATES.summary; return s ? { newCommits: updatesBadge(), status: s.status, latestRelease: s.latestRelease || '', href: hashFor('updates') } : null; },
  onRoute: function (fn) {
    if (typeof fn !== 'function') return function () {};
    routeHooks.push(fn);
    return function () { var i = routeHooks.indexOf(fn); if (i !== -1) routeHooks.splice(i, 1); };
  },
  projectSettings: {
    allowed: function () { return !!token() && psAllowed().length > 0; },
    href: function () { return psEntryHash(); },
    isOpen: function () { return !!ps; },
    open: function () { if (!ps) psNavigate(psEntryHash(), false); },
    sections: function () { return psAllowed().map(function (s) { return { id: s.id, href: hashFor(s.id), label: t(s.label) }; }); }
  },
  brand: function () {
    var b = BRAND.data;
    if (!b) return null;
    return { name: brandName(), shortName: b.shortName || '', markUrl: b.markUrl || '', logoUrl: b.logoUrl || '', logoDarkUrl: b.logoDarkUrl || '' };
  },
  toggleTheme: function () { toggleTheme(); },
  isDark: function () { return isDarkTheme(); }
};

if (legacyHash(location.hash) || /^#\/(economy|publication|moderation|support|system|project-settings)(\/|\?|$)/.test(location.hash)) {
  pendingHash = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
trapApp();
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot);
else boot();
