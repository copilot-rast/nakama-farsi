var CU_ROOT = 'cup-root';
var CU_GROUPS = { cups: ['list', 'templates', 'settings'] };
var CU_ORDER = ['cups'];
var CU_RENDER = { 'cups/list': cupsPage, 'cups/view': cupViewPage, 'cups/templates': templatesPage, 'cups/settings': cupSettingsPage };
var CU_HASH = /^#\/(cups)(?:\/([^/?]*))?(?:\/([^/?]+))?/;
var cuPending = null;
var cuRouter = null;
var cuMounted = null;
var cuScheduled = false;

function cuRoute(hash) {
  var m = CU_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  var pages = CU_GROUPS[m[1]];
  if (m[2] === 'list' && m[3]) return m[1] + '/view';
  return m[1] + '/' + (pages.indexOf(m[2]) !== -1 ? m[2] : pages[0]);
}

function cuAllowed() {
  return !!token();
}

function cuTitle() {
  var page = cuRoute();
  return page ? t('group.' + page.split('/')[0]) + ' - ' + t('nav.' + page) : '';
}

function cuAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || cuRouter) return !!cuRouter;
  try {
    var layout = r.resolve('/').matched[0];
    CU_ORDER.forEach(function (g) {
      if (r.hasRoute(g)) return;
      r.addRoute({ path: '/' + g + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: g, component: { name: 'CupsPage', render: function () { return null; } }, meta: { get title() { return cuTitle(); } } }] });
    });
    cuRouter = r;
  } catch (e) {
    return false;
  }
  if (cuPending) {
    var target = cuPending;
    cuPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function cuNavLink(g, page, active) {
  var id = g + '/' + page;
  var on = id === active || (page === 'list' && active === g + '/view');
  return h('li', {}, h('a', { href: '#/' + id, 'data-cup-nav': id, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + id) }))));
}

function cuSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelectorAll('[data-cup-group]');
  if (!group || !cuAllowed()) { Array.prototype.forEach.call(existing, function (e) { e.remove(); }); return; }
  var target = document.querySelector('.x-drawer [data-clan-group="trophies"]') || document.querySelector('.x-drawer [data-quiz-group]') || document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = cuLang() + '|' + active;
  var mine = document.querySelector('[data-cup-group="cups"]');
  if (mine && existing.length === 1 && mine.getAttribute('data-key') === key && mine.nextElementSibling === target) return;
  Array.prototype.forEach.call(existing, function (e) { e.remove(); });
  CU_ORDER.forEach(function (g) {
    target.before(h('div', { 'data-cup-group': g, 'data-plane-group': g, 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
      h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group.' + g) }),
      h('nav', { 'aria-label': t('group.' + g) }, h('ul', { class: 'space-y-0.5' }, CU_GROUPS[g].map(function (p) { return cuNavLink(g, p, active); })))));
  });
}

function cuShowMains() {
  var hidden = document.querySelectorAll('[data-cup-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-cup-hidden'); }
}

function cuUnmount() {
  var root = document.getElementById(CU_ROOT);
  if (root) root.remove();
  cuShowMains();
  closePanel();
  stopPoll();
  cuMounted = null;
}

function cuSync() {
  cuAddRoutes();
  var page = cuRoute();
  cuSyncNav(page);
  if (!page) { if (cuMounted || document.getElementById(CU_ROOT)) cuUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-cup-hidden', ''); }
  }
  var root = document.getElementById(CU_ROOT);
  if (!root) {
    root = h('div', { id: CU_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-cup': 'root' });
    cuMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + cuLang() + (page === 'cups/view' || page === 'cups/list' ? '|' + location.hash : '');
  if (cuMounted !== key) {
    cuMounted = key;
    closePanel();
    CU_RENDER[page](root);
  }
}

function cuSchedule() {
  if (cuScheduled) return;
  cuScheduled = true;
  requestAnimationFrame(function () { cuScheduled = false; cuSync(); });
}

function cuBoot() {
  window.addEventListener('hashchange', cuSchedule);
  window.addEventListener('popstate', cuSchedule);
  window.addEventListener('nakama-console:ready', cuSchedule);
  new MutationObserver(cuSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { cuMounted = null; cuSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (cuAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  cuSync();
}

window.NakamaCup = { version: 1, pages: function () {
  var out = [];
  CU_ORDER.forEach(function (g) { CU_GROUPS[g].forEach(function (p) { out.push({ id: g + '/' + p, group: g, href: '#/' + g + '/' + p, label: t('nav.' + g + '/' + p) }); }); });
  return out;
} };

if (CU_HASH.test(location.hash)) {
  cuPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', cuBoot);
else cuBoot();
