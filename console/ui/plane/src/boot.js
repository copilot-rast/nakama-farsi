var tickQueued = false;
var lastTick = 0;

function tick() {
  tickQueued = false;
  lastTick = Date.now();
  paintSky();
  var auth = isAuthPage();
  if (auth !== root.classList.contains('pl-auth')) {
    root.classList.toggle('pl-auth', auth);
    if (auth && typeof closeFly === 'function') { closeFly(true); closeSheet(); }
  }
  if (!auth) {
    if (typeof renderRail === 'function') renderRail();
    if (typeof paintTopbar === 'function') paintTopbar();
    markPage();
    labelTables();
    paintNativeTabs();
    countKpis();
    iconKpis();
    attentionCard();
  }
}

function schedule() {
  if (tickQueued) return;
  tickQueued = true;
  var wait = Math.max(0, 90 - (Date.now() - lastTick));
  setTimeout(function () { requestAnimationFrame(tick); }, wait);
}

function boot() {
  paintSky();
  if (!hookRoutes()) {
    window.addEventListener('nakama-console:ready', hookRoutes);
    var tries = 0;
    var poll = setInterval(function () {
      tries++;
      if (hookRoutes() || tries > 50) {
        clearInterval(poll);
        if (!hookRoutes.done) window.addEventListener('hashchange', function () { routeEnter(); schedule(); });
      }
    }, 100);
  }
  window.addEventListener('hashchange', schedule);
  window.addEventListener('resize', function () {
    if (typeof segIndicator === 'function') {
      if (tb.start) segIndicator(tb.start.querySelector('.pl-tabs'));
      if (tb.seg) segIndicator(tb.seg);
      if (!mqMobile.matches) closeSheet();
    }
    schedule();
  });
  if (typeof globalKeys === 'function') document.addEventListener('keydown', globalKeys);
  new MutationObserver(function (records) {
    for (var i = 0; i < records.length; i++) {
      var t = records[i].target;
      if (t.nodeType === 1 && t.closest && t.closest('.pl-tabind, .pl-rail, .pl-rail-full, .pl-flyout, .pl-tb-start, .pl-tb-end, .pl-sky, .pl-tip, .pl-pal-holder, [_echarts_instance_], number-flow-vue')) continue;
      schedule();
      return;
    }
  }).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function (records) {
    var hit = records.some(function (r) {
      if (r.attributeName !== 'class') return r.oldValue !== root.getAttribute(r.attributeName);
      return /(^|\s)dark(\s|$)/.test(r.oldValue || '') !== root.classList.contains('dark');
    });
    if (!hit) return;
    if (typeof rail === 'object') rail.key = '';
    if (typeof tb === 'object') tb.key = '';
    schedule();
  }).observe(root, { attributes: true, attributeOldValue: true, attributeFilter: ['class', 'lang', 'dir'] });
  if (mqMobile.addEventListener) mqMobile.addEventListener('change', function () { if (typeof rail === 'object') rail.key = ''; schedule(); });
  tick();
}

if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', boot);
else boot();
