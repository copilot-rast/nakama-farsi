var CG_ROOT = 'campaign-root';
var CG_PAGES = ['chapters', 'stages', 'freeride', 'report', 'settings'];
var CG_RENDER = { chapters: chaptersPage, stages: stagesPage, freeride: freeridePage, report: reportPage, settings: campaignSettingsPage };
var CG_HASH = /^#\/campaign(?:\/([^/?]*))?/;
var cgPending = null;
var cgRouter = null;
var cgMounted = null;
var cgScheduled = false;

function cgRoute(hash) {
  var m = CG_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  return CG_PAGES.indexOf(m[1]) !== -1 ? m[1] : CG_PAGES[0];
}

function cgAllowed() {
  return !!token();
}

function cgTitle() {
  var page = cgRoute();
  return t('group') + (page ? ' - ' + t('nav.' + page) : '');
}

function cgAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || cgRouter) return !!cgRouter;
  try {
    if (!r.hasRoute('campaign')) {
      var layout = r.resolve('/').matched[0];
      r.addRoute({ path: '/campaign/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'campaign', component: { name: 'CampaignPage', render: function () { return null; } }, meta: { get title() { return cgTitle(); } } }] });
    }
    cgRouter = r;
  } catch (e) {
    return false;
  }
  if (cgPending) {
    var target = cgPending;
    cgPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function cgNavLink(page, active) {
  var on = page === active;
  return h('li', {}, h('a', { href: '#/campaign/' + page, 'data-campaign-nav': page, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + page) }))));
}

function cgSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelector('[data-campaign-group]');
  if (!group || !cgAllowed()) { if (existing) existing.remove(); return; }
  var target = document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = cgLang() + '|' + active;
  if (existing && existing.getAttribute('data-key') === key && existing.nextElementSibling === target) return;
  var built = h('div', { 'data-campaign-group': '', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group') }),
    h('nav', { 'aria-label': t('group') }, h('ul', { class: 'space-y-0.5' }, CG_PAGES.map(function (p) { return cgNavLink(p, active); }))));
  if (existing) existing.remove();
  target.before(built);
}

function cgShowMains() {
  var hidden = document.querySelectorAll('[data-campaign-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-campaign-hidden'); }
}

function cgUnmount() {
  var root = document.getElementById(CG_ROOT);
  if (root) root.remove();
  cgShowMains();
  closePanel();
  cgMounted = null;
}

function cgSync() {
  cgAddRoutes();
  var page = cgRoute();
  cgSyncNav(page);
  cgPlayerCard();
  if (!page) { if (cgMounted || document.getElementById(CG_ROOT)) cgUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-campaign-hidden', ''); }
  }
  var root = document.getElementById(CG_ROOT);
  if (!root) {
    root = h('div', { id: CG_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-campaign': 'root' });
    cgMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + cgLang() + (page === 'stages' || page === 'report' ? '|' + location.hash : '');
  if (cgMounted !== key) {
    cgMounted = key;
    closePanel();
    CG_RENDER[page](root);
  }
}

function cgSchedule() {
  if (cgScheduled) return;
  cgScheduled = true;
  requestAnimationFrame(function () { cgScheduled = false; cgSync(); });
}

function cgBoot() {
  window.addEventListener('hashchange', cgSchedule);
  window.addEventListener('popstate', cgSchedule);
  window.addEventListener('nakama-console:ready', cgSchedule);
  new MutationObserver(cgSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { cgMounted = null; cgSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (cgAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  cgSync();
}

window.NakamaCampaign = { version: 1, pages: function () { return CG_PAGES.map(function (p) { return { id: p, href: '#/campaign/' + p, label: t('nav.' + p) }; }); } };

if (CG_HASH.test(location.hash)) {
  cgPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', cgBoot);
else cgBoot();
