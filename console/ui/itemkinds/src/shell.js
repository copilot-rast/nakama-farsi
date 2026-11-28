var IK_ROOT = 'itemkinds-root';
var IK_STATIC = ['routes', 'runs', 'settings'];
var IK_HASH = /^#\/(itemkinds)(?:\/([^/?]*))?(?:\/([^/?]+))?/;
var ikPending = null;
var ikRouter = null;
var ikMounted = null;
var ikScheduled = false;
var ikFetched = false;

function ikRoute(hash) {
  var m = IK_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  if (m[2] === 'kind' && m[3]) return 'itemkinds/kind';
  if (IK_STATIC.indexOf(m[2]) !== -1) return 'itemkinds/' + m[2];
  var first = (ikKindCache || [])[0];
  return first ? 'itemkinds/kind' : 'itemkinds/routes';
}

function ikAllowed() {
  return !!token();
}

function ikNavItems() {
  var out = [];
  (ikKindCache || []).forEach(function (k) {
    if (k.enabled === false) return;
    out.push({ id: 'itemkinds/kind/' + k.key, label: textOf(k.label, k.key) });
  });
  IK_STATIC.forEach(function (p) { out.push({ id: 'itemkinds/' + p, label: t('nav.itemkinds/' + p) }); });
  return out;
}

function ikTitle() {
  var page = ikRoute();
  if (!page) return '';
  if (page === 'itemkinds/kind') {
    var k = ikKind((location.hash.split('?')[0].split('/')[3] || ''));
    return t('group.itemkinds') + ' - ' + (k ? textOf(k.label, k.key) : '');
  }
  return t('group.itemkinds') + ' - ' + t('nav.' + page);
}

function ikAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || ikRouter) return !!ikRouter;
  try {
    var layout = r.resolve('/').matched[0];
    if (!r.hasRoute('itemkinds')) {
      r.addRoute({ path: '/itemkinds/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'itemkinds', component: { name: 'ItemKindsPage', render: function () { return null; } }, meta: { get title() { return ikTitle(); } } }] });
    }
    ikRouter = r;
  } catch (e) {
    return false;
  }
  if (ikPending) {
    var target = ikPending;
    ikPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function ikActiveId() {
  var page = ikRoute();
  if (page === 'itemkinds/kind') return 'itemkinds/kind/' + (location.hash.split('?')[0].split('/')[3] || '');
  return page;
}

function ikNavLink(item, active) {
  var on = item.id === active;
  return h('li', {}, h('a', { href: '#/' + item.id, 'data-ik-nav': item.id, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: item.label }))));
}

function ikModOn(id) {
  var M = window.NakamaConsole && window.NakamaConsole.modules;
  return !M || !M.on || M.on(id);
}

function ikSyncNav() {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelectorAll('[data-ik-group]');
  var kinds = (ikKindCache || []).filter(function (k) { return k.enabled !== false; });
  if (!group || !ikAllowed() || !ikModOn('itemkinds') || !kinds.length) { Array.prototype.forEach.call(existing, function (e) { e.remove(); }); return; }
  var target = document.querySelector('.x-drawer [data-cup-group]') || document.querySelector('.x-drawer [data-clan-group="trophies"]') || document.querySelector('.x-drawer [data-quiz-group]') || document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var active = ikActiveId();
  var items = ikNavItems();
  var key = ikLang() + '|' + active + '|' + items.map(function (i) { return i.id + ':' + i.label; }).join(',');
  var mine = document.querySelector('[data-ik-group="itemkinds"]');
  if (mine && existing.length === 1 && mine.getAttribute('data-key') === key && mine.nextElementSibling === target) return;
  Array.prototype.forEach.call(existing, function (e) { e.remove(); });
  target.before(h('div', { 'data-ik-group': 'itemkinds', 'data-plane-group': 'itemkinds', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group.itemkinds') }),
    h('nav', { 'aria-label': t('group.itemkinds') }, h('ul', { class: 'space-y-0.5' }, items.map(function (i) { return ikNavLink(i, active); })))));
}

function ikShowMains() {
  var hidden = document.querySelectorAll('[data-ik-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-ik-hidden'); }
}

function ikUnmount() {
  var root = document.getElementById(IK_ROOT);
  if (root) root.remove();
  ikShowMains();
  closePanel();
  ikMounted = null;
}

function ikRender(page, root) {
  if (page === 'itemkinds/kind') {
    var key = location.hash.split('?')[0].split('/')[3] || '';
    if (!key && ikKindCache && ikKindCache[0]) { location.replace('#/itemkinds/kind/' + ikKindCache[0].key); return; }
    kindPage(root);
  } else if (page === 'itemkinds/routes') routesPage(root);
  else if (page === 'itemkinds/runs') runsPage(root);
  else ikSettingsPage(root);
}

function ikSync() {
  ikAddRoutes();
  if (ikAllowed() && !ikFetched && !ikKindLoading) {
    ikFetched = true;
    ikLoadKinds().then(function () { ikMounted = null; ikSchedule(); }, function () { ikFetched = false; });
  }
  if (!ikAllowed()) ikFetched = false;
  var page = ikRoute();
  ikSyncNav();
  if (!page) { if (ikMounted || document.getElementById(IK_ROOT)) ikUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-ik-hidden', ''); }
  }
  var root = document.getElementById(IK_ROOT);
  if (!root) {
    root = h('div', { id: IK_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-ik': 'root' });
    ikMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  if (!ikKindCache) {
    if (ikMounted !== 'loading') { ikMounted = 'loading'; clear(root).appendChild(h('p', { class: 'px-6 text-sm text-secondary-500', text: t('loading') })); }
    return;
  }
  var key = page + '|' + ikLang() + '|' + ikModOn('itemkinds') + '|' + location.hash.split('?')[0];
  if (ikMounted !== key) {
    ikMounted = key;
    closePanel();
    if (ikModOn('itemkinds')) ikRender(page, root);
    else window.NakamaConsole.modules.offPage(root, 'itemkinds');
  }
}

function ikSchedule() {
  if (ikScheduled) return;
  ikScheduled = true;
  requestAnimationFrame(function () { ikScheduled = false; ikSync(); });
}

function ikBoot() {
  window.addEventListener('hashchange', ikSchedule);
  window.addEventListener('popstate', ikSchedule);
  window.addEventListener('nakama-console:ready', ikSchedule);
  window.addEventListener('nakama-console:modules', ikSchedule);
  new MutationObserver(ikSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { ikMounted = null; ikSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (ikAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  ikSync();
}

window.NakamaItemKinds = { version: 1, pages: function () {
  return ikNavItems().map(function (i) { return { id: i.id, group: 'itemkinds', href: '#/' + i.id, label: i.label }; });
}, reload: function () { return ikLoadKinds(true).then(function () { ikMounted = null; ikSchedule(); }); } };

if (IK_HASH.test(location.hash)) {
  ikPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', ikBoot);
else ikBoot();
