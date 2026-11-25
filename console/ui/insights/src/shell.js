var INS_ROOT = 'ins-root';
var INS_GROUPS = { insights: ['overview', 'active', 'retention', 'funnels', 'explore', 'sessions', 'catalog', 'storage', 'imports', 'settings', 'context'] };
var INS_ORDER = ['insights'];
var INS_RENDER = { 'insights/overview': overviewPage, 'insights/active': activePage, 'insights/retention': retentionPage, 'insights/funnels': funnelsPage, 'insights/explore': explorePage, 'insights/sessions': sessionsPage, 'insights/catalog': catalogPage, 'insights/storage': storagePage, 'insights/imports': importsPage, 'insights/settings': analyticsSettingsPage, 'insights/context': contextFillPage };
var INS_HASH = /^#\/(insights)(?:\/([^/?]*))?/;
var insPending = null;
var insRouter = null;
var insMounted = null;
var insScheduled = false;

function insRoute(hash) {
  var m = INS_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  var pages = INS_GROUPS[m[1]];
  return m[1] + '/' + (pages.indexOf(m[2]) !== -1 ? m[2] : pages[0]);
}

function insAllowed() {
  return !!token();
}

function insTitle() {
  var page = insRoute();
  return page ? t('group.' + page.split('/')[0]) + ' - ' + t('nav.' + page) : '';
}

function insAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || insRouter) return !!insRouter;
  try {
    var layout = r.resolve('/').matched[0];
    INS_ORDER.forEach(function (g) {
      if (r.hasRoute(g)) return;
      r.addRoute({ path: '/' + g + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: g, component: { name: 'InsightsPage', render: function () { return null; } }, meta: { get title() { return insTitle(); } } }] });
    });
    insRouter = r;
  } catch (e) {
    return false;
  }
  if (insPending) {
    var target = insPending;
    insPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function insNavLink(g, page, active) {
  var id = g + '/' + page;
  var on = id === active;
  return h('li', {}, h('a', { href: '#/' + id, 'data-ins-nav': id, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + id) }))));
}

function insSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelectorAll('[data-ins-group]');
  if (!group || !insAllowed()) { Array.prototype.forEach.call(existing, function (e) { e.remove(); }); return; }
  var target = document.querySelector('.x-drawer [data-cup-group]') || document.querySelector('.x-drawer [data-clan-group="trophies"]') || document.querySelector('.x-drawer [data-quiz-group]') || document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = insLang() + '|' + active;
  var mine = document.querySelector('[data-ins-group="insights"]');
  if (mine && existing.length === 1 && mine.getAttribute('data-key') === key && mine.nextElementSibling === target) return;
  Array.prototype.forEach.call(existing, function (e) { e.remove(); });
  INS_ORDER.forEach(function (g) {
    target.before(h('div', { 'data-ins-group': g, 'data-plane-group': g, 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
      h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group.' + g) }),
      h('nav', { 'aria-label': t('group.' + g) }, h('ul', { class: 'space-y-0.5' }, INS_GROUPS[g].map(function (p) { return insNavLink(g, p, active); })))));
  });
}

function insShowMains() {
  var hidden = document.querySelectorAll('[data-ins-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-ins-hidden'); }
}

function insUnmount() {
  var root = document.getElementById(INS_ROOT);
  if (root) root.remove();
  insShowMains();
  closePanel();
  insMounted = null;
}

function insSync() {
  insAddRoutes();
  var page = insRoute();
  insSyncNav(page);
  if (!page) { if (insMounted || document.getElementById(INS_ROOT)) insUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-ins-hidden', ''); }
  }
  var root = document.getElementById(INS_ROOT);
  if (!root) {
    root = h('div', { id: INS_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-ins': 'root' });
    insMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + insLang() + '|' + location.hash;
  if (insMounted !== key) {
    insMounted = key;
    closePanel();
    INS_RENDER[page](root);
  }
}

function insSchedule() {
  if (insScheduled) return;
  insScheduled = true;
  requestAnimationFrame(function () { insScheduled = false; insSync(); });
}

function insBoot() {
  window.addEventListener('hashchange', insSchedule);
  window.addEventListener('popstate', insSchedule);
  window.addEventListener('nakama-console:ready', insSchedule);
  new MutationObserver(insSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { insMounted = null; insSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (insAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  insSync();
}

window.NakamaInsights = { version: 1, pages: function () {
  var out = [];
  INS_ORDER.forEach(function (g) { INS_GROUPS[g].forEach(function (p) { out.push({ id: g + '/' + p, group: g, href: '#/' + g + '/' + p, label: t('nav.' + g + '/' + p) }); }); });
  return out;
} };

if (INS_HASH.test(location.hash)) {
  insPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', insBoot);
else insBoot();
