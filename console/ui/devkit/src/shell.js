var DK_ROOT = 'devkit-root';
var DK_PAGES = ['errors', 'api-keys', 'rpc-schema'];
var DK_RENDER = { errors: errorsPage, 'api-keys': keysPage, 'rpc-schema': schemaPage };
var DK_HASH = /^#\/devkit(?:\/([^/?]*))?/;
var dkPending = null;
var dkRouter = null;
var dkMounted = null;
var dkScheduled = false;

function dkRoute(hash) {
  var m = DK_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  return DK_PAGES.indexOf(m[1]) !== -1 ? m[1] : DK_PAGES[0];
}

function dkAllowed() {
  return !!token();
}

function dkTitle() {
  var page = dkRoute();
  return t('group') + (page ? ' - ' + t('nav.' + page) : '');
}

function dkAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || dkRouter) return !!dkRouter;
  try {
    if (!r.hasRoute('devkit')) {
      var layout = r.resolve('/').matched[0];
      r.addRoute({ path: '/devkit/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'devkit', component: { name: 'DevkitPage', render: function () { return null; } }, meta: { get title() { return dkTitle(); } } }] });
    }
    dkRouter = r;
  } catch (e) {
    return false;
  }
  if (dkPending) {
    var target = dkPending;
    dkPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function dkNavLink(page, active) {
  var on = page === active;
  return h('li', {}, h('a', { href: '#/devkit/' + page, 'data-devkit-nav': page, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + page) }))));
}

function dkSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelector('[data-devkit-group]');
  if (!group || !dkAllowed()) { if (existing) existing.remove(); return; }
  var key = dkLang() + '|' + active;
  if (existing && existing.getAttribute('data-key') === key && existing.previousElementSibling === group) return;
  var built = h('div', { 'data-devkit-group': '', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group') }),
    h('nav', { 'aria-label': t('group') }, h('ul', { class: 'space-y-0.5' }, DK_PAGES.map(function (p) { return dkNavLink(p, active); }))));
  if (existing) existing.remove();
  group.after(built);
}

function dkShowMains() {
  var hidden = document.querySelectorAll('[data-devkit-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-devkit-hidden'); }
}

function dkUnmount() {
  var root = document.getElementById(DK_ROOT);
  if (root) root.remove();
  dkShowMains();
  closePanel();
  dkMounted = null;
}

function dkSync() {
  dkAddRoutes();
  var page = dkRoute();
  dkSyncNav(page);
  if (!page) { if (dkMounted || document.getElementById(DK_ROOT)) dkUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-devkit-hidden', ''); }
  }
  var root = document.getElementById(DK_ROOT);
  if (!root) {
    root = h('div', { id: DK_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-devkit': 'root' });
    dkMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + dkLang();
  if (dkMounted !== key) {
    dkMounted = key;
    closePanel();
    DK_RENDER[page](root);
  }
}

function dkSchedule() {
  if (dkScheduled) return;
  dkScheduled = true;
  requestAnimationFrame(function () { dkScheduled = false; dkSync(); });
}

function dkBoot() {
  window.addEventListener('hashchange', dkSchedule);
  window.addEventListener('popstate', dkSchedule);
  window.addEventListener('nakama-console:ready', dkSchedule);
  new MutationObserver(dkSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { dkMounted = null; dkSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (dkAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  dkSync();
}

window.NakamaDevkit = { version: 1, pages: function () { return DK_PAGES.map(function (p) { return { id: p, href: '#/devkit/' + p, label: t('nav.' + p) }; }); } };

if (DK_HASH.test(location.hash)) {
  dkPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', dkBoot);
else dkBoot();
