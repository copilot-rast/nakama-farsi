var NATIVE_PURPLE = /^(#6b59ee|rgb\(107, ?89, ?238\))$/i;
var CHART_AXIS = '20px';
var CHART_EDGE = '24px';
var CHART = { key: '', patched: false };
var MONACO = { node: null, len: -1, el: null };

function vueFind(test) {
  var app = document.getElementById('app');
  var out = [];
  var seen = new Set();
  function walkVnode(v, depth) {
    if (!v || depth > 800) return;
    if (v.component) walkInstance(v.component, depth + 1);
    if (Array.isArray(v.children)) {
      for (var i = 0; i < v.children.length; i++) {
        var c = v.children[i];
        if (c && typeof c === 'object') walkVnode(c, depth + 1);
      }
    }
    if (v.suspense && v.suspense.activeBranch) walkVnode(v.suspense.activeBranch, depth + 1);
  }
  function walkInstance(inst, depth) {
    if (seen.has(inst)) return;
    seen.add(inst);
    if (test(inst)) out.push(inst);
    walkVnode(inst.subTree, depth + 1);
  }
  walkVnode(app && app._vnode, 0);
  return out;
}

function faDigits(s) {
  return String(s).replace(/(\d),(?=\d{3}\b)/g, '$1٬').replace(/(\d)\.(?=\d)/g, '$1٫').replace(/\d/g, function (d) { return PERSIAN[+d]; });
}

function compactFa(v) {
  var s = String(v);
  var m = /^(-?\d+(?:\.\d+)?)\s?([kKM])$/.exec(s);
  if (m) s = String(Number(m[1])) + ' ' + (m[2] === 'M' ? 'میلیون' : 'هزار');
  return faDigits(s);
}

function chartAccent() {
  var b = BRAND.data;
  return b && b.accent ? brandHex(b.accent) : null;
}

function asList(v) {
  return Array.isArray(v) ? v : [v];
}

function mapAxis(v, fn) {
  if (!v) return v;
  return Array.isArray(v) ? v.map(fn) : fn(v);
}

function wrapFn(fn, make) {
  if (typeof fn !== 'function' || fn.__nk) return fn;
  var out = make(fn);
  out.__nk = true;
  return out;
}

function htmlDigits(html) {
  return String(html).replace(/>([^<]+)</g, function (m, text) { return '>' + faDigits(text) + '<'; });
}

function accentGradient(hex) {
  var rgb = hexRgb(hex).join(',');
  return { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(' + rgb + ',0.3)' }, { offset: 1, color: 'rgba(' + rgb + ',0)' }] };
}

var CHART_SECONDS = new Intl.DateTimeFormat(undefined, { hour: 'numeric', minute: '2-digit', second: '2-digit' });

function seriesSpan(series) {
  var min = Infinity;
  var max = -Infinity;
  asList(series || []).forEach(function (s) {
    (s && Array.isArray(s.data) ? s.data : []).forEach(function (d) {
      var v = Array.isArray(d) ? d[0] : d && d.value ? d.value[0] : null;
      if (typeof v === 'string') v = Date.parse(v);
      if (typeof v !== 'number' || !isFinite(v)) return;
      if (v < 1e11) v *= 1000;
      if (v < min) min = v;
      if (v > max) max = v;
    });
  });
  return max > min ? max - min : 0;
}

function chartOption(opt) {
  if (!opt || typeof opt !== 'object' || Array.isArray(opt)) return opt;
  var rtl = document.documentElement.getAttribute('dir') === 'rtl';
  var fa = lang() === 'fa';
  var accent = chartAccent();
  var o = Object.assign({}, opt);
  if (o.xAxis) {
    o.xAxis = mapAxis(o.xAxis, function (a) {
      var x = Object.assign({}, a, { inverse: rtl });
      var span = seriesSpan(o.series);
      if (x.type === 'time' && span > 0 && span < 900000 && x.axisLabel && typeof x.axisLabel.formatter === 'function' && !x.axisLabel.formatter.__nkSec) {
        var sec = function (v) { return CHART_SECONDS.format(new Date(v)); };
        sec.__nk = true;
        sec.__nkSec = true;
        x.axisLabel = Object.assign({}, x.axisLabel, { formatter: sec });
      }
      return x;
    });
  }
  if (o.yAxis) {
    o.yAxis = mapAxis(o.yAxis, function (a) {
      var y = Object.assign({}, a, { position: rtl ? 'right' : 'left' });
      if (fa && y.axisLabel && typeof y.axisLabel.formatter === 'function') {
        y.axisLabel = Object.assign({}, y.axisLabel, { formatter: wrapFn(y.axisLabel.formatter, function (f) { return function () { return compactFa(f.apply(this, arguments)); }; }) });
      }
      return y;
    });
  }
  if (o.grid) {
    o.grid = mapAxis(o.grid, function (g) {
      if (!g || !g.containLabel) return g;
      return Object.assign({}, g, rtl ? { left: CHART_EDGE, right: CHART_AXIS } : { left: CHART_AXIS, right: CHART_EDGE });
    });
  }
  if (o.title && rtl) {
    o.title = mapAxis(o.title, function (t) { return Object.assign({}, t, { left: 'right' }); });
  }
  if (o.legend && rtl) {
    o.legend = mapAxis(o.legend, function (l) { return Object.assign({}, l, { left: 'right' }); });
  }
  if (o.tooltip) {
    o.tooltip = mapAxis(o.tooltip, function (tt) {
      var x = Object.assign({}, tt);
      if (rtl) x.extraCssText = 'direction:rtl;text-align:right;';
      if (fa && typeof x.formatter === 'function') x.formatter = wrapFn(x.formatter, function (f) { return function () { return htmlDigits(f.apply(this, arguments)); }; });
      return x;
    });
  }
  if (accent && o.series) {
    o.series = asList(o.series).map(function (s) {
      if (!s || typeof s !== 'object') return s;
      var line = s.lineStyle && s.lineStyle.color;
      var item = s.itemStyle && s.itemStyle.color;
      if (!(typeof line === 'string' && NATIVE_PURPLE.test(line)) && !(typeof item === 'string' && NATIVE_PURPLE.test(item))) return s;
      var x = Object.assign({}, s);
      if (x.lineStyle) x.lineStyle = Object.assign({}, x.lineStyle, { color: accent });
      if (x.itemStyle) x.itemStyle = Object.assign({}, x.itemStyle, { color: accent });
      if (x.areaStyle) x.areaStyle = Object.assign({}, x.areaStyle, { color: accentGradient(accent) });
      x.emphasis = Object.assign({}, x.emphasis, { lineStyle: Object.assign({}, x.emphasis && x.emphasis.lineStyle, { color: accent }), itemStyle: Object.assign({}, x.emphasis && x.emphasis.itemStyle, { color: accent }) });
      return x;
    });
  }
  return o;
}

function patchCharts(proto) {
  if (!proto || proto.__nkCharts) return;
  var setOption = proto.setOption;
  Object.defineProperty(proto, '__nkCharts', { value: true });
  proto.setOption = function (opt) {
    var args = Array.prototype.slice.call(arguments);
    args[0] = chartOption(opt);
    return setOption.apply(this, args);
  };
}

function restyleChart(chart) {
  if (!chart || chart.isDisposed && chart.isDisposed()) return;
  var o;
  try { o = chart.getOption(); } catch (e) { return; }
  if (!o) return;
  var part = {};
  ['xAxis', 'yAxis', 'grid', 'title', 'legend', 'tooltip', 'series'].forEach(function (k) { if (o[k] && o[k].length) part[k] = o[k]; });
  chart.setOption(part);
}

function syncCharts() {
  var els = document.querySelectorAll('[_echarts_instance_]');
  if (!els.length) return;
  var key = Array.prototype.map.call(els, function (e) { return e.getAttribute('_echarts_instance_'); }).join(',') + '|' + chartAccent() + '|' + lang();
  if (key === CHART.key) return;
  CHART.key = key;
  vueFind(function (inst) { return inst.type && inst.type.name === 'echarts'; }).forEach(function (inst) {
    var chart = inst.setupState && inst.setupState.chart;
    if (!chart || typeof chart.setOption !== 'function') return;
    patchCharts(Object.getPrototypeOf(chart));
    restyleChart(chart);
  });
}

function hexHsl(hex) {
  var c = hexRgb(hex).map(function (v) { return v / 255; });
  var max = Math.max(c[0], c[1], c[2]);
  var min = Math.min(c[0], c[1], c[2]);
  var l = (max + min) / 2;
  var d = max - min;
  var hue = 0;
  var s = d === 0 ? 0 : d / (1 - Math.abs(2 * l - 1));
  if (d) {
    if (max === c[0]) hue = ((c[1] - c[2]) / d) % 6;
    else if (max === c[1]) hue = (c[2] - c[0]) / d + 2;
    else hue = (c[0] - c[1]) / d + 4;
  }
  return [(hue * 60 + 360) % 360, s, l];
}

function darkToken(hex) {
  var hsl = hexHsl(hex);
  if (hsl[2] >= 0.6) return hex;
  var l = Math.min(0.86, 0.7 + (0.6 - hsl[2]) * 0.3);
  var s = hsl[1] < 0.15 ? hsl[1] : Math.min(0.85, Math.max(0.5, hsl[1]));
  return 'hsl(' + Math.round(hsl[0]) + ' ' + Math.round(s * 100) + '% ' + Math.round(l * 100) + '%)';
}

function syncMonaco() {
  var node = document.querySelector('style.monaco-colors');
  if (!node) return;
  var len = node.textContent.length;
  if (node === MONACO.node && len === MONACO.len && MONACO.el && MONACO.el.isConnected) return;
  MONACO.node = node;
  MONACO.len = len;
  var rules = [];
  var re = /\.(mtk\d+) \{ color: (#[0-9a-fA-F]{6}); \}/g;
  var m;
  while ((m = re.exec(node.textContent))) rules.push('html.dark .monaco-editor .' + m[1] + '{color:' + darkToken(m[2]) + '}');
  if (!MONACO.el) MONACO.el = h('style', { id: 'nk-monaco-dark' });
  MONACO.el.textContent = rules.join('\n');
  if (MONACO.el.parentNode !== document.head || MONACO.el.nextSibling) document.head.appendChild(MONACO.el);
}

var SETUP_NUDGE = { state: null, at: 0, loading: false };

function dashboardSetup() {
  var path = (location.hash || '#/').replace(/^#/, '').split('?')[0];
  var card = document.querySelector('[data-setup-nudge]');
  if (path !== '/' && path !== '') { if (card) card.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.max-w-\\[1600px\\] .grid.grid-cols-1.gap-6');
  if (!grid || !token()) return;
  if (!SETUP_NUDGE.loading && Date.now() - SETUP_NUDGE.at > 15000) {
    SETUP_NUDGE.loading = true;
    quietBase('GET', '/setup').then(function (st) {
      SETUP_NUDGE.loading = false;
      SETUP_NUDGE.at = Date.now();
      SETUP_NUDGE.state = st;
      schedule();
    });
  }
  var st = SETUP_NUDGE.state;
  var show = !!st && st.status !== 'completed' && !wiz;
  if (!show) { if (card) card.remove(); return; }
  var steps = SETUP_ORDER.slice(0, -1);
  var done = steps.filter(function (k) { return setupStepStatus(st, k) !== 'pending'; }).length;
  var key = lang() + '|' + st.status + '|' + done;
  if (card && card.getAttribute('data-key') === key && card.parentNode === grid && grid.firstChild === card) return;
  if (card) card.remove();
  var progress = t('setup.progress', { n: localDigits(done), total: localDigits(steps.length) });
  var go = button({ variant: 'primary', size: 'sm', icon: 'flask', label: t('setup.nudge.go'), onclick: function () { openWizard(SETUP_NUDGE.state, false); SETUP_NUDGE.at = 0; } });
  go.setAttribute('data-setup-nudge-go', '');
  card = h('section', { 'data-setup-nudge': '', 'data-key': key, 'data-economy': 'nudge', 'data-l10n-skip': '', class: 'setup-nudge', 'aria-labelledby': 'setup-nudge-title', dir: document.documentElement.getAttribute('dir') || 'ltr' },
    h('div', { class: 'setup-nudge-mark' }, icon('flask', 'h-5 w-5')),
    h('div', { class: 'setup-nudge-text' },
      h('h2', { id: 'setup-nudge-title', class: 'font-medium', text: t(st.status === 'skipped' ? 'setup.nudge.skipped' : 'setup.nudge.title') }),
      h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400' }, t('setup.nudge.desc') + ' ', h('span', { class: 'tabular-nums', text: progress }))),
    h('div', { class: 'setup-nudge-bar', role: 'progressbar', 'aria-valuemin': '0', 'aria-valuemax': String(steps.length), 'aria-valuenow': String(done), 'aria-label': progress },
      h('div', { class: 'setup-progress-bar', style: 'width:' + Math.round(done / steps.length * 100) + '%' })),
    go);
  grid.insertBefore(card, grid.firstChild);
}

var AUDIT_LABELS = /^(base|economy|publication) (create|update|delete|invoke) /;

function auditLogSources() {
  if (!/^#\/settings\/audit-log/.test(location.hash || '')) return;
  var rows = document.querySelectorAll('#main table tbody tr');
  for (var i = 0; i < rows.length; i++) {
    var row = rows[i];
    var msg = row.getAttribute('data-nk-audit');
    if (msg === null) {
      var cells = row.cells;
      msg = '';
      for (var c = 0; c < cells.length; c++) {
        var txt = cells[c].textContent.replace(/[⁦-⁩]/g, '').trim();
        if (AUDIT_LABELS.test(txt) || /^(پایه|اقتصاد|انتشار) · /.test(txt)) { msg = 'module'; break; }
      }
      row.setAttribute('data-nk-audit', msg);
    }
    if (msg !== 'module') continue;
    var walker = document.createTreeWalker(row, NodeFilter.SHOW_TEXT);
    var n;
    while ((n = walker.nextNode())) {
      var v = n.nodeValue.trim();
      if (v === 'Hiro Economy' || v === 'اقتصاد Hiro') n.nodeValue = n.nodeValue.replace(v, t('section'));
    }
  }
}

var SATORI = { state: null, promise: null };

function satoriState() {
  if (SATORI.state !== null) return Promise.resolve(SATORI.state);
  if (SATORI.promise) return SATORI.promise;
  var tk = token();
  if (!tk) return Promise.resolve(true);
  SATORI.promise = fetch(CONSOLE_BASE + '/extensions', { headers: { Accept: 'application/json', Authorization: 'Bearer ' + tk } }).then(function (res) { return res.ok ? res.json() : null; }).then(function (ext) {
    SATORI.promise = null;
    SATORI.state = ext ? !!ext.satori : true;
    return SATORI.state;
  }, function () { SATORI.promise = null; return true; });
  return SATORI.promise;
}

function satoriGuard(to) {
  if (!to || !/^\/players\/[^/]+\/satori(\/|$)/.test(to.path)) return true;
  return satoriState().then(function (on) { return on ? true : to.path.replace(/\/satori(\/.*)?$/, ''); });
}

function satoriRoutes() {
  if (!/^#\/players\/[^/?]+\/satori(\/|\?|$)/.test(location.hash || '')) return;
  satoriState().then(function (on) {
    if (on || !router) return;
    var m = /^#(\/players\/[^/?]+)\/satori/.exec(location.hash);
    if (m) router.replace(m[1]);
  });
}

var NOT_FOUND = { last: '', at: 0 };

function notFoundHook(arg) {
  var path = null;
  if (arg && arg.newURL) path = (new URL(arg.newURL).hash || '#/').replace(/^#/, '').split('?')[0];
  else if (arg && arg.redirectedFrom) path = arg.redirectedFrom.path;
  else path = (location.hash || '#/').replace(/^#/, '').split('?')[0];
  checkNotFound(path || '/');
}

function checkNotFound(path) {
  if (!router || !token()) return;
  var r;
  try { r = router.resolve(path); } catch (e) { return; }
  if (!r || r.name !== '404') return;
  if (NOT_FOUND.last === path && Date.now() - NOT_FOUND.at < 1500) return;
  NOT_FOUND.last = path;
  NOT_FOUND.at = Date.now();
  toast('info', t('notFound', { path: '\u2068' + path + '\u2069' }));
  setTimeout(function () {
    var now = (location.hash || '#/').replace(/^#/, '').split('?')[0];
    if (now === path) history.replaceState(history.state, '', location.pathname + location.search + '#/');
  }, 0);
}

window.addEventListener('hashchange', notFoundHook);

function nativeLabels() {
  var acct = document.querySelector('#main > div.sticky:first-child .justify-end > .x-popover > button.x-button');
  if (acct && acct.getAttribute('aria-label') !== t('account')) {
    acct.setAttribute('aria-label', t('account'));
    acct.setAttribute('title', t('account'));
  }
  if (document.getElementById('main')) return;
  var links = document.querySelectorAll('a[href="/"]');
  for (var i = 0; i < links.length; i++) {
    var a = links[i];
    if (!a.querySelector('img')) continue;
    var label = brandName() || t('brand.neutral');
    if (a.getAttribute('aria-label') !== label) a.setAttribute('aria-label', label);
    if (a.getAttribute('tabindex') !== '-1') a.setAttribute('tabindex', '-1');
  }
}

function nativeTweaks() {
  nativeLabels();
  syncCharts();
  syncMonaco();
  dashboardSetup();
  auditLogSources();
  satoriRoutes();
  playerModCard();
  playerSocialCard();
  playerProgressCard();
  playerActivityCard();
  playerTasksCard();
  playerItemsCard();
  playerAchievementsCard();
  playerChanceCard();
  playerOffersCard();
  playerFriendGiftsCard();
  playerChatQuotaCard();
}
