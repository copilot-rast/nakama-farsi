var ROOT_ID = 'economy-root';
var NAV_LINK = 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent dark:border-transparent';
var NAV_ACTIVE = '!border-primary-600 !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700';
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
    if (!r.hasRoute('economy')) {
      r.addRoute({ path: '/economy/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'economy', component: EmptyPage, meta: { title: 'Economy' } }] });
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
  var m = /^#\/economy(?:\/([^/?]*))?(?:\/([^?]*))?(?:\?(.*))?$/.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  var page = m[1] || 'overview';
  if (PAGES.indexOf(page) === -1) page = 'overview';
  var query = {};
  (m[3] || '').split('&').forEach(function (pair) {
    if (!pair) return;
    var i = pair.indexOf('=');
    var k = decodeURIComponent(i === -1 ? pair : pair.slice(0, i));
    var v = i === -1 ? '' : decodeURIComponent(pair.slice(i + 1).replace(/\+/g, ' '));
    query[k] = v;
  });
  return { page: page, id: m[2] ? decodeURIComponent(m[2]) : null, query: query };
}

function keyOf(r) {
  if (!r) return '';
  var detail = r.page === 'payments' && r.id ? '/' + r.id : '';
  return lang() + '|' + r.page + detail + '|' + JSON.stringify(r.query);
}

function navIcon(name) {
  var svg = icon(name, 'h-4 w-4');
  return svg;
}

function buildNav(collapsed, active) {
  var ul = h('ul', { class: 'space-y-0.5' });
  PAGES.forEach(function (p) {
    var on = p === active;
    var a = h('a', { href: hashFor(p, null, p === active ? null : lastQuery[p]), class: (on ? NAV_ACTIVE + ' ' : '') + NAV_LINK, 'aria-current': on ? 'page' : null, title: collapsed ? t('nav.' + p) : null, 'data-economy-page': p });
    a.appendChild(collapsed ? h('div', { class: 'x-tooltip flex' }, navIcon(p)) : navIcon(p));
    a.appendChild(h('div', { class: 'font-medium relative truncate' }, h('span', { text: collapsed ? ' ' : t('nav.' + p) })));
    ul.appendChild(h('li', {}, a));
  });
  return h('div', { 'data-economy-nav': '', 'data-l10n-skip': '' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 ' + (collapsed ? 'text-[8px] text-center' : 'px-3 text-[10px]'), text: collapsed ? t('section.short') : t('section') }),
    h('nav', { 'aria-label': t('section') }, ul));
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
    var key = lang() + '|' + collapsed + '|' + active + '|' + JSON.stringify(lastQuery);
    var next = group.nextElementSibling;
    var existing = next && next.hasAttribute('data-economy-nav') ? next : null;
    if (!existing || existing.getAttribute('data-key') !== key) {
      var built = buildNav(collapsed, active);
      built.setAttribute('data-key', key);
      if (existing) existing.replaceWith(built);
      else group.after(built);
      existing = built;
    }
    keep.push(existing);
    var vueActive = group.parentElement.querySelectorAll('a[aria-current="page"]:not([data-economy-page])');
    for (var j = 0; j < vueActive.length && active; j++) {
      if (vueActive[j].getAttribute('href') === '#/') vueActive[j].setAttribute('data-economy-dimmed', '');
    }
  }
  var all = document.querySelectorAll('[data-economy-nav]');
  for (var k = 0; k < all.length; k++) if (keep.indexOf(all[k]) === -1) all[k].remove();
  document.documentElement.toggleAttribute('data-economy-active', !!active);
}

function unmount() {
  var root = document.getElementById(ROOT_ID);
  if (root) root.remove();
  var hidden = document.querySelectorAll('[data-economy-hidden]');
  for (var i = 0; i < hidden.length; i++) {
    hidden[i].style.display = '';
    hidden[i].removeAttribute('data-economy-hidden');
  }
  closePanel();
  rendered = null;
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
    if (route.page === 'overview') overviewPage(root);
    else if (route.page === 'payment-methods') paymentMethodsPage(root);
    else if (route.page === 'payments' && route.id) paymentDetail(root, route.id);
    else ctl.list = listPage(root, route.page, route.query, ctx);
  }
  if (RESOURCES[route.page] && ctl.id !== route.id) {
    ctl.id = route.id;
    if (route.id) openEditor(route.page, route.id, route.query, function () { if (ctl.list) ctl.list.reload(); });
    else closePanel();
  }
}

function sync() {
  var route = parseRoute();
  syncSidebar(route);
  if (!route) { if (rendered || document.getElementById(ROOT_ID)) unmount(); return; }
  var main = document.querySelector('#main > main');
  if (!main) return;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') {
      mains[i].style.display = 'none';
      mains[i].setAttribute('data-economy-hidden', '');
    }
  }
  var root = document.getElementById(ROOT_ID);
  if (!root) {
    root = h('div', { id: ROOT_ID, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-economy': 'root' });
    rendered = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  render(root, route);
}

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
  new MutationObserver(function () { if (rendered) rendered.key = ''; schedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  sync();
}

if (/^#\/economy/.test(location.hash)) {
  pendingHash = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
trapApp();
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot);
else boot();
