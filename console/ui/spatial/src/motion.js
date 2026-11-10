function pageKey(path) {
  var seg = String(path || '/').split('?')[0].split('/').filter(Boolean);
  if (!seg.length) return '/';
  if (seg[0] === 'project-settings') return 'ps';
  if (seg[0] === 'economy' || seg[0] === 'publication' || seg[0] === 'base') return seg[0] + '/' + (seg[1] || '');
  if (seg[0] === 'settings') return 'settings/' + (seg[1] || '');
  if (seg[0] === 'payments') return 'payments/' + (seg[1] || '');
  return seg.slice(0, 2).join('/');
}

var enterTimer = 0;
function animateEnter(el) {
  if (!el || reduced()) return;
  el.classList.remove('sp-enter');
  void el.offsetWidth;
  el.classList.add('sp-enter');
  clearTimeout(enterTimer);
  var done = function (e) {
    if (e && e.target !== el) return;
    el.classList.remove('sp-enter');
    el.removeEventListener('animationend', done);
  };
  el.addEventListener('animationend', done);
  enterTimer = setTimeout(done, 700);
}

function routeEnter() {
  requestAnimationFrame(function () {
    requestAnimationFrame(function () { animateEnter(contentRoot()); });
  });
}

function onRouteChange(to, from) {
  var a = pageKey(to && to.path);
  var b = pageKey(from && from.path);
  var moved = (to && to.fullPath) !== (from && from.fullPath);
  if (moved && typeof closeFly === 'function') {
    closeFly(true);
    closeSheet();
    toggleStats(false);
  }
  if (a !== b && a !== 'ps' && b !== 'ps') routeEnter();
  schedule();
}

function countUp(node, to, fmt) {
  if (reduced() || !isFinite(to)) return;
  var from = node.__spValue || 0;
  node.__spValue = to;
  var start = performance.now();
  var dur = 480;
  var step = function (now) {
    var k = Math.min(1, (now - start) / dur);
    var e = 1 - Math.pow(1 - k, 3);
    node.textContent = fmt(from + (to - from) * e);
    if (k < 1) requestAnimationFrame(step);
  };
  requestAnimationFrame(step);
}

function hookRoutes() {
  var api = nc();
  if (!api || !api.onRoute || hookRoutes.done) return !!hookRoutes.done;
  hookRoutes.done = true;
  api.onRoute(onRouteChange);
  return true;
}
