var CL_ROOT = 'clan-root';
var CL_GROUPS = { trophies: ['leagues', 'players', 'seasons', 'settings'], clans: ['list', 'levels', 'wars', 'rounds', 'gallery', 'uploads', 'settings'] };
var CL_ORDER = ['trophies', 'clans'];
var CL_RENDER = { 'trophies/leagues': leaguesPage, 'trophies/players': trPlayersPage, 'trophies/seasons': seasonsPage, 'trophies/settings': trophySettingsPage,
  'clans/list': clansPage, 'clans/levels': levelsPage, 'clans/wars': warsPage, 'clans/rounds': roundsPage, 'clans/gallery': galleryPage, 'clans/uploads': uploadsPage, 'clans/settings': clanSettingsPage };
var CL_HASH = /^#\/(trophies|clans)(?:\/([^/?]*))?/;
var clPending = null;
var clRouter = null;
var clMounted = null;
var clScheduled = false;

function clRoute(hash) {
  var m = CL_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  var pages = CL_GROUPS[m[1]];
  return m[1] + '/' + (pages.indexOf(m[2]) !== -1 ? m[2] : pages[0]);
}

function clAllowed() {
  return !!token();
}

function clTitle() {
  var page = clRoute();
  return page ? t('group.' + page.split('/')[0]) + ' - ' + t('nav.' + page) : '';
}

function clAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || clRouter) return !!clRouter;
  try {
    var layout = r.resolve('/').matched[0];
    CL_ORDER.forEach(function (g) {
      if (r.hasRoute(g)) return;
      r.addRoute({ path: '/' + g + '/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: g, component: { name: g === 'clans' ? 'ClansPage' : 'TrophiesPage', render: function () { return null; } }, meta: { get title() { return clTitle(); } } }] });
    });
    clRouter = r;
  } catch (e) {
    return false;
  }
  if (clPending) {
    var target = clPending;
    clPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function clNavLink(g, page, active) {
  var id = g + '/' + page;
  var on = id === active;
  return h('li', {}, h('a', { href: '#/' + id, 'data-clan-nav': id, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + id) })),
    id === 'clans/uploads' && clPendingUploads ? h('span', { class: 'eco-nav-badge tabular-nums', 'aria-label': t('up.pendingN', { n: clPendingUploads }), text: fmtNum(clPendingUploads) }) : null));
}

function clModOn(id) {
  var M = window.NakamaConsole && window.NakamaConsole.modules;
  return !M || !M.on || M.on(id);
}

function clSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelectorAll('[data-clan-group]');
  if (!group || !clAllowed()) { Array.prototype.forEach.call(existing, function (e) { e.remove(); }); return; }
  var target = document.querySelector('.x-drawer [data-quiz-group]') || document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = clLang() + '|' + active + '|' + clPendingUploads + '|' + clModOn('trophies') + clModOn('clans');
  var first = document.querySelector('[data-clan-group="trophies"]');
  var second = document.querySelector('[data-clan-group="clans"]');
  var shown = CL_ORDER.filter(clModOn);
  if (!shown.length) { Array.prototype.forEach.call(existing, function (e) { e.remove(); }); return; }
  var last = document.querySelector('[data-clan-group="' + shown[shown.length - 1] + '"]');
  if (existing.length === shown.length && Array.prototype.every.call(existing, function (e) { return e.getAttribute('data-key') === key; }) && (shown.length < 2 || (first && first.nextElementSibling === second)) && last && last.nextElementSibling === target) return;
  Array.prototype.forEach.call(existing, function (e) { e.remove(); });
  CL_ORDER.forEach(function (g) {
    if (!clModOn(g)) return;
    target.before(h('div', { 'data-clan-group': g, 'data-plane-group': g, 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
      h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group.' + g) }),
      h('nav', { 'aria-label': t('group.' + g) }, h('ul', { class: 'space-y-0.5' }, CL_GROUPS[g].map(function (p) { return clNavLink(g, p, active); })))));
  });
}

function clShowMains() {
  var hidden = document.querySelectorAll('[data-clan-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-clan-hidden'); }
}

function clUnmount() {
  var root = document.getElementById(CL_ROOT);
  if (root) root.remove();
  clShowMains();
  closePanel();
  clMounted = null;
}

function clSync() {
  clAddRoutes();
  var page = clRoute();
  if (clAllowed()) refreshPendingUploads();
  clSyncNav(page);
  if (!page) { if (clMounted || document.getElementById(CL_ROOT)) clUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-clan-hidden', ''); }
  }
  var root = document.getElementById(CL_ROOT);
  if (!root) {
    root = h('div', { id: CL_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-clan': 'root' });
    clMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + clLang() + '|' + clModOn(page.split('/')[0]) + (page === 'clans/rounds' ? '|' + location.hash : '');
  if (clMounted !== key) {
    clMounted = key;
    closePanel();
    if (clModOn(page.split('/')[0])) CL_RENDER[page](root);
    else window.NakamaConsole.modules.offPage(root, page.split('/')[0]);
  }
}

function clSchedule() {
  if (clScheduled) return;
  clScheduled = true;
  requestAnimationFrame(function () { clScheduled = false; clSync(); });
}

function clBoot() {
  window.addEventListener('hashchange', clSchedule);
  window.addEventListener('popstate', clSchedule);
  window.addEventListener('nakama-console:ready', clSchedule);
  window.addEventListener('nakama-console:modules', clSchedule);
  new MutationObserver(clSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { clMounted = null; clSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (clAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  clSync();
}

window.NakamaClan = { version: 1, pages: function () {
  var out = [];
  CL_ORDER.forEach(function (g) { CL_GROUPS[g].forEach(function (p) { out.push({ id: g + '/' + p, group: g, href: '#/' + g + '/' + p, label: t('nav.' + g + '/' + p) }); }); });
  return out;
} };

if (CL_HASH.test(location.hash)) {
  clPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', clBoot);
else clBoot();
