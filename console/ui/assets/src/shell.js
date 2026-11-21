var AS_ROOT = 'assets-root';
var AS_PAGES = ['entries', 'packs', 'settings'];
var AS_RENDER = { entries: entriesPage, packs: packsPage, settings: assetsSettingsPage };
var AS_HASH = /^#\/assets(?:\/([^/?]*))?/;
var asPending = null;
var asRouter = null;
var asMounted = null;
var asScheduled = false;

function asRoute(hash) {
  var m = AS_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  return AS_PAGES.indexOf(m[1]) !== -1 ? m[1] : AS_PAGES[0];
}

function asAllowed() {
  return !!token();
}

function asTitle() {
  var page = asRoute();
  return t('group') + (page ? ' - ' + t('nav.' + page) : '');
}

function asAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || asRouter) return !!asRouter;
  try {
    if (!r.hasRoute('assets')) {
      var layout = r.resolve('/').matched[0];
      r.addRoute({ path: '/assets/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'assets', component: { name: 'AssetsPage', render: function () { return null; } }, meta: { get title() { return asTitle(); } } }] });
    }
    asRouter = r;
  } catch (e) {
    return false;
  }
  if (asPending) {
    var target = asPending;
    asPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function asNavLink(page, active) {
  var on = page === active;
  return h('li', {}, h('a', { href: '#/assets/' + page, 'data-assets-nav': page, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + page) }))));
}

function asSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelector('[data-assets-group]');
  if (!group || !asAllowed()) { if (existing) existing.remove(); return; }
  var target = document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = asLang() + '|' + active;
  if (existing && existing.getAttribute('data-key') === key && existing.nextElementSibling === target) return;
  var built = h('div', { 'data-assets-group': '', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group') }),
    h('nav', { 'aria-label': t('group') }, h('ul', { class: 'space-y-0.5' }, AS_PAGES.map(function (p) { return asNavLink(p, active); }))));
  if (existing) existing.remove();
  target.before(built);
}

function asShowMains() {
  var hidden = document.querySelectorAll('[data-assets-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-assets-hidden'); }
}

function asUnmount() {
  var root = document.getElementById(AS_ROOT);
  if (root) root.remove();
  asShowMains();
  closePanel();
  asMounted = null;
}

function asSync() {
  asAddRoutes();
  var page = asRoute();
  asSyncNav(page);
  if (!page) { if (asMounted || document.getElementById(AS_ROOT)) asUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-assets-hidden', ''); }
  }
  var root = document.getElementById(AS_ROOT);
  if (!root) {
    root = h('div', { id: AS_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-assets': 'root' });
    asMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + asLang();
  if (asMounted !== key) {
    asMounted = key;
    closePanel();
    AS_RENDER[page](root);
  }
}

function asSchedule() {
  if (asScheduled) return;
  asScheduled = true;
  requestAnimationFrame(function () { asScheduled = false; asSync(); });
}

function asBoot() {
  window.addEventListener('hashchange', asSchedule);
  window.addEventListener('popstate', asSchedule);
  window.addEventListener('nakama-console:ready', asSchedule);
  new MutationObserver(asSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { asMounted = null; asSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (asAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  asSync();
}

window.NakamaAssets = { version: 1, pages: function () { return AS_PAGES.map(function (p) { return { id: p, href: '#/assets/' + p, label: t('nav.' + p) }; }); } };

if (AS_HASH.test(location.hash)) {
  asPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', asBoot);
else asBoot();
