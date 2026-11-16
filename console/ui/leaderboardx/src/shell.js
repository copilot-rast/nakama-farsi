var LB_ROOT = 'lbx-root';
var LB_PAGES = ['boards', 'standings', 'winners', 'report', 'settings'];
var LB_RENDER = { boards: boardsPage, standings: standingsPage, winners: winnersPage, report: reportPage, settings: lbSettingsPage };
var LB_HASH = /^#\/leaderboardx(?:\/([^/?]*))?/;
var lbPending = null;
var lbRouter = null;
var lbMounted = null;
var lbScheduled = false;

function lbRoute(hash) {
  var m = LB_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  return LB_PAGES.indexOf(m[1]) !== -1 ? m[1] : LB_PAGES[0];
}

function lbAllowed() {
  return !!token();
}

function lbTitle() {
  var page = lbRoute();
  return t('group') + (page ? ' - ' + t('nav.' + page) : '');
}

function lbAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || lbRouter) return !!lbRouter;
  try {
    if (!r.hasRoute('leaderboardx')) {
      var layout = r.resolve('/').matched[0];
      r.addRoute({ path: '/leaderboardx/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'leaderboardx', component: { name: 'LeaderboardxPage', render: function () { return null; } }, meta: { get title() { return lbTitle(); } } }] });
    }
    lbRouter = r;
  } catch (e) {
    return false;
  }
  if (lbPending) {
    var target = lbPending;
    lbPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function lbNavLink(page, active) {
  var on = page === active;
  return h('li', {}, h('a', { href: '#/leaderboardx/' + page, 'data-lbx-nav': page, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + page) }))));
}

function lbSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelector('[data-lbx-group]');
  if (!group || !lbAllowed()) { if (existing) existing.remove(); return; }
  var target = document.querySelector('.x-drawer [data-assets-group]') || document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = lbLang() + '|' + active;
  if (existing && existing.getAttribute('data-key') === key && existing.nextElementSibling === target) return;
  var built = h('div', { 'data-lbx-group': '', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group') }),
    h('nav', { 'aria-label': t('group') }, h('ul', { class: 'space-y-0.5' }, LB_PAGES.map(function (p) { return lbNavLink(p, active); }))));
  if (existing) existing.remove();
  target.before(built);
}

function lbShowMains() {
  var hidden = document.querySelectorAll('[data-lbx-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-lbx-hidden'); }
}

function lbUnmount() {
  var root = document.getElementById(LB_ROOT);
  if (root) root.remove();
  lbShowMains();
  closePanel();
  lbMounted = null;
}

function lbSync() {
  lbAddRoutes();
  var page = lbRoute();
  lbSyncNav(page);
  lbPlayerCard();
  if (!page) { if (lbMounted || document.getElementById(LB_ROOT)) lbUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-lbx-hidden', ''); }
  }
  var root = document.getElementById(LB_ROOT);
  if (!root) {
    root = h('div', { id: LB_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-lbx': 'root' });
    lbMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + lbLang() + (page === 'standings' || page === 'winners' || page === 'report' ? '|' + location.hash : '');
  if (lbMounted !== key) {
    lbMounted = key;
    closePanel();
    LB_RENDER[page](root);
  }
}

function lbSchedule() {
  if (lbScheduled) return;
  lbScheduled = true;
  requestAnimationFrame(function () { lbScheduled = false; lbSync(); });
}

function lbBoot() {
  window.addEventListener('hashchange', lbSchedule);
  window.addEventListener('popstate', lbSchedule);
  window.addEventListener('nakama-console:ready', lbSchedule);
  new MutationObserver(lbSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { lbMounted = null; lbSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (lbAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  lbSync();
}

window.NakamaLeaderboardx = { version: 1, pages: function () { return LB_PAGES.map(function (p) { return { id: p, href: '#/leaderboardx/' + p, label: t('nav.' + p) }; }); } };

if (LB_HASH.test(location.hash)) {
  lbPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', lbBoot);
else lbBoot();
