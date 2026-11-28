var QZ_ROOT = 'quiz-root';
var QZ_PAGES = ['packs', 'questions', 'missions', 'duels', 'duel-categories', 'reports', 'results', 'import', 'stats', 'settings'];
var QZ_RENDER = { packs: packsPage, questions: questionsPage, missions: missionsPage, duels: duelsPage, 'duel-categories': duelCategoriesPage, reports: reportsPage, results: resultsPage, 'import': importPage, stats: statsPage, settings: quizSettingsPage };
var QZ_HASH = /^#\/quiz(?:\/([^/?]*))?/;
var qzPending = null;
var qzRouter = null;
var qzMounted = null;
var qzScheduled = false;

function qzRoute(hash) {
  var m = QZ_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  return QZ_PAGES.indexOf(m[1]) !== -1 ? m[1] : QZ_PAGES[0];
}

function qzAllowed() {
  return !!token();
}

function qzTitle() {
  var page = qzRoute();
  return t('group') + (page ? ' - ' + t('nav.' + page) : '');
}

function qzAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || qzRouter) return !!qzRouter;
  try {
    if (!r.hasRoute('quiz')) {
      var layout = r.resolve('/').matched[0];
      r.addRoute({ path: '/quiz/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'quiz', component: { name: 'QuizPage', render: function () { return null; } }, meta: { get title() { return qzTitle(); } } }] });
    }
    qzRouter = r;
  } catch (e) {
    return false;
  }
  if (qzPending) {
    var target = qzPending;
    qzPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function qzNavLink(page, active) {
  var on = page === active;
  return h('li', {}, h('a', { href: '#/quiz/' + page, 'data-quiz-nav': page, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + page) }))));
}

function qzModOn(id) {
  var M = window.NakamaConsole && window.NakamaConsole.modules;
  return !M || !M.on || M.on(id);
}

function qzSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelector('[data-quiz-group]');
  if (!group || !qzAllowed() || !qzModOn('quiz')) { if (existing) existing.remove(); return; }
  var target = document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = qzLang() + '|' + active;
  if (existing && existing.getAttribute('data-key') === key && existing.nextElementSibling === target) return;
  var built = h('div', { 'data-quiz-group': '', 'data-plane-group': 'quiz', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group') }),
    h('nav', { 'aria-label': t('group') }, h('ul', { class: 'space-y-0.5' }, QZ_PAGES.map(function (p) { return qzNavLink(p, active); }))));
  if (existing) existing.remove();
  target.before(built);
}

function qzShowMains() {
  var hidden = document.querySelectorAll('[data-quiz-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-quiz-hidden'); }
}

function qzUnmount() {
  var root = document.getElementById(QZ_ROOT);
  if (root) root.remove();
  qzShowMains();
  closePanel();
  qzMounted = null;
}

function qzSync() {
  qzAddRoutes();
  var page = qzRoute();
  qzSyncNav(page);
  qzPlayerCard();
  if (!page) { if (qzMounted || document.getElementById(QZ_ROOT)) qzUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-quiz-hidden', ''); }
  }
  var root = document.getElementById(QZ_ROOT);
  if (!root) {
    root = h('div', { id: QZ_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-quiz': 'root' });
    qzMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + qzLang() + '|' + qzModOn('quiz') + (page === 'questions' || page === 'reports' || page === 'stats' || page === 'duels' ? '|' + location.hash : '');
  if (qzMounted !== key) {
    qzMounted = key;
    closePanel();
    if (qzModOn('quiz')) QZ_RENDER[page](root);
    else window.NakamaConsole.modules.offPage(root, 'quiz');
  }
}

function qzSchedule() {
  if (qzScheduled) return;
  qzScheduled = true;
  requestAnimationFrame(function () { qzScheduled = false; qzSync(); });
}

function qzBoot() {
  window.addEventListener('hashchange', qzSchedule);
  window.addEventListener('popstate', qzSchedule);
  window.addEventListener('nakama-console:ready', qzSchedule);
  window.addEventListener('nakama-console:modules', qzSchedule);
  new MutationObserver(qzSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { qzMounted = null; qzSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (qzAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  qzSync();
}

window.NakamaQuiz = { version: 1, pages: function () { return QZ_PAGES.map(function (p) { return { id: p, href: '#/quiz/' + p, label: t('nav.' + p) }; }); } };

if (QZ_HASH.test(location.hash)) {
  qzPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', qzBoot);
else qzBoot();
