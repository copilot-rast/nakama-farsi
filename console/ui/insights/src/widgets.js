var INS_SVG = 'http://www.w3.org/2000/svg';
var INS_SERIES = 8;

function s(tag, attrs) {
  var el = document.createElementNS(INS_SVG, tag);
  Object.keys(attrs || {}).forEach(function (k) { if (attrs[k] != null) el.setAttribute(k, attrs[k]); });
  for (var i = 2; i < arguments.length; i++) if (arguments[i]) el.appendChild(arguments[i]);
  return el;
}

function seriesColor(i) {
  return 'var(--ins-s' + ((i % INS_SERIES) + 1) + ')';
}

function niceMax(v) {
  if (!(v > 0)) return 1;
  var p = Math.pow(10, Math.floor(Math.log10(v)));
  var n = v / p;
  var step = n <= 1 ? 1 : n <= 2 ? 2 : n <= 5 ? 5 : 10;
  return step * p;
}

function shortNum(v) {
  var a = Math.abs(v);
  var out = a >= 1e9 ? (v / 1e9).toFixed(1) + 'B' : a >= 1e6 ? (v / 1e6).toFixed(1) + 'M' : a >= 1e4 ? (v / 1e3).toFixed(0) + 'K' : String(Math.round(v * 100) / 100);
  return digits(out.replace('.0', ''));
}

function legend(series) {
  if (series.length < 2) return null;
  return h('div', { class: 'ia-legend flex flex-wrap gap-x-4 gap-y-1 text-xs text-secondary-600 dark:text-secondary-300', 'data-ins-legend': '' },
    series.map(function (sr, i) {
      return h('span', { class: 'inline-flex items-center gap-1.5' }, h('span', { class: 'ia-swatch', style: 'background:' + seriesColor(i) }), h('span', { dir: 'auto', text: sr.label }));
    }));
}

function tooltipBox() {
  return h('div', { class: 'ia-tip', role: 'status', hidden: true });
}

function lineChart(labels, series, opts) {
  opts = opts || {};
  var W = 720, H = opts.height || 240, L = 48, R = 12, T = 12, B = 28;
  var max = 0;
  var ints = true;
  series.forEach(function (sr) { sr.values.forEach(function (v) { if (v > max) max = v; if (v % 1) ints = false; }); });
  var step = niceMax(max / 4);
  if (ints && step < 1) step = 1;
  max = step * 4;
  var n = labels.length;
  var x = function (i) { return n <= 1 ? L + (W - L - R) / 2 : L + i * (W - L - R) / (n - 1); };
  var y = function (v) { return T + (H - T - B) * (1 - v / max); };
  var svg = s('svg', { viewBox: '0 0 ' + W + ' ' + H, class: 'ia-chart', role: 'img', 'aria-label': opts.label || '', preserveAspectRatio: 'none' });
  for (var g = 0; g <= 4; g++) {
    var gv = max * g / 4;
    svg.appendChild(s('line', { x1: L, x2: W - R, y1: y(gv), y2: y(gv), class: 'ia-grid' }));
    var tx = s('text', { x: L - 6, y: y(gv) + 4, class: 'ia-axis', 'text-anchor': 'end' });
    tx.textContent = shortNum(gv);
    svg.appendChild(tx);
  }
  var every = Math.max(1, Math.ceil(n / 8));
  labels.forEach(function (lb, i) {
    if (i % every !== 0 && i !== n - 1) return;
    var tx = s('text', { x: x(i), y: H - 8, class: 'ia-axis', 'text-anchor': 'middle' });
    tx.textContent = opts.fmtLabel ? opts.fmtLabel(lb) : lb;
    svg.appendChild(tx);
  });
  series.forEach(function (sr, si) {
    var d = sr.values.map(function (v, i) { return (i ? 'L' : 'M') + x(i).toFixed(1) + ' ' + y(v || 0).toFixed(1); }).join(' ');
    svg.appendChild(s('path', { d: d, class: 'ia-line', stroke: seriesColor(si), 'data-ins-series': sr.label }));
    if (n <= 40) sr.values.forEach(function (v, i) { svg.appendChild(s('circle', { cx: x(i), cy: y(v || 0), r: 3, fill: seriesColor(si), class: 'ia-dot' })); });
  });
  var cross = s('line', { y1: T, y2: H - B, class: 'ia-cross', visibility: 'hidden' });
  svg.appendChild(cross);
  var tip = tooltipBox();
  var wrap = h('div', { class: 'ia-chartwrap relative', 'data-ins-chart': opts.name || 'line' }, svg, tip);
  svg.addEventListener('mousemove', function (e) {
    var r = svg.getBoundingClientRect();
    var px = (e.clientX - r.left) / r.width * W;
    var i = n <= 1 ? 0 : Math.max(0, Math.min(n - 1, Math.round((px - L) / ((W - L - R) / (n - 1)))));
    cross.setAttribute('x1', x(i));
    cross.setAttribute('x2', x(i));
    cross.setAttribute('visibility', 'visible');
    clear(tip).appendChild(h('div', { class: 'font-medium mb-1', text: opts.fmtLabel ? opts.fmtLabel(labels[i]) : labels[i] }));
    series.forEach(function (sr, si) {
      tip.appendChild(h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'ia-swatch', style: 'background:' + seriesColor(si) }), h('span', { class: 'flex-1', dir: 'auto', text: sr.label }), h('span', { class: 'font-medium', text: fmtNum(sr.values[i]) })));
    });
    tip.hidden = false;
    var left = x(i) / W * r.width;
    tip.style.left = Math.min(Math.max(0, left + 12), r.width - 180) + 'px';
    tip.style.top = '8px';
  });
  svg.addEventListener('mouseleave', function () { tip.hidden = true; cross.setAttribute('visibility', 'hidden'); });
  return h('div', { class: 'flex flex-col gap-2' }, legend(series), wrap);
}

function barList(rows, opts) {
  opts = opts || {};
  var max = 0;
  rows.forEach(function (r) { if (r.value > max) max = r.value; });
  var tip = tooltipBox();
  var list = h('div', { class: 'flex flex-col gap-1.5 relative', 'data-ins-chart': opts.name || 'bars', role: 'list' });
  rows.forEach(function (r) {
    var pct = max > 0 ? Math.max(1, r.value / max * 100) : 0;
    var row = h('div', { class: 'ia-barrow', role: 'listitem', tabindex: '0' },
      h('div', { class: 'ia-barlabel text-sm truncate', dir: 'auto', title: r.label, text: r.label }),
      h('div', { class: 'ia-bartrack' }, h('div', { class: 'ia-bar', style: 'width:' + pct + '%;background:' + (r.color || seriesColor(0)) })),
      h('div', { class: 'ia-barvalue text-sm font-medium', text: opts.fmt ? opts.fmt(r.value) : fmtNum(r.value) }));
    var show = function () {
      clear(tip).appendChild(h('div', { class: 'font-medium', dir: 'auto', text: r.label }));
      (r.tip || []).forEach(function (line) { tip.appendChild(h('div', { text: line })); });
      tip.hidden = false;
      tip.style.top = (row.offsetTop + row.offsetHeight) + 'px';
      tip.style.left = '0px';
    };
    row.addEventListener('mouseenter', show);
    row.addEventListener('focus', show);
    row.addEventListener('mouseleave', function () { tip.hidden = true; });
    row.addEventListener('blur', function () { tip.hidden = true; });
    list.appendChild(row);
  });
  list.appendChild(tip);
  return list;
}

function kpis(items) {
  return h('div', { class: 'ia-kpis grid gap-3', 'data-ins-kpis': '' }, items.map(function (k) {
    return card(h('div', { class: 'p-4', 'data-ins-kpi': k.key || '' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: k.label }),
      h('div', { class: 'text-2xl font-semibold mt-1', text: k.value }),
      k.sub ? h('div', { class: 'text-xs text-secondary-500 mt-1', text: k.sub }) : null));
  }));
}

function csvDownload(name, columns, rows) {
  var esc = function (v) {
    var x = v == null ? '' : String(v);
    return /[",\n]/.test(x) ? '"' + x.replace(/"/g, '""') + '"' : x;
  };
  var text = '﻿' + [columns.map(esc).join(',')].concat(rows.map(function (r) { return r.map(esc).join(','); })).join('\n');
  var url = URL.createObjectURL(new Blob([text], { type: 'text/csv;charset=utf-8' }));
  var a = h('a', { href: url, download: name + '.csv' });
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(function () { URL.revokeObjectURL(url); }, 1000);
}

function csvButton(name, get) {
  return button({ label: t('csv'), data: { 'ins-csv': name }, onclick: function () { var d = get(); csvDownload(name, d.columns, d.rows); } });
}

function panelCard(title, body, actions, name) {
  return card(h('div', { class: 'p-4 flex flex-col gap-3', 'data-ins-section': name || '' },
    h('div', { class: 'flex items-center justify-between gap-2 flex-wrap' }, h('h2', { class: 'font-semibold', text: title }), actions ? h('div', { class: 'flex gap-2' }, actions) : null),
    body));
}

function insQuery() {
  var q = hashQuery();
  var today = new Date();
  if (!q.to) q.to = isoDay(today);
  if (!q.from) { var f = new Date(today); f.setDate(f.getDate() - 29); q.from = isoDay(f); }
  return q;
}

function setQuery(next) {
  var page = (location.hash.split('?')[0] || '#/insights/overview');
  var parts = [];
  Object.keys(next).forEach(function (k) { if (next[k] !== '' && next[k] != null) parts.push(encodeURIComponent(k) + '=' + encodeURIComponent(next[k])); });
  location.hash = page + (parts.length ? '?' + parts.join('&') : '');
}

function filterQS(q, extra) {
  var p = { from: q.from, to: q.to, platform: q.platform, appVersion: q.appVersion, includeFlagged: q.includeFlagged };
  if (q.game != null && q.game !== '*' && q.game !== undefined) p.game = q.game === '~' ? '' : q.game;
  var out = qs(Object.assign(p, extra || {}));
  if (q.game === '~') out += (out ? '&' : '?') + 'game=';
  return out;
}

var insGames = null;
function loadGames() {
  if (insGames) return Promise.resolve(insGames);
  return api('GET', '/insights/events/catalog').then(function (d) { insGames = d.games || []; return insGames; }, function () { insGames = []; return insGames; });
}

function filterBar(q, opts) {
  opts = opts || {};
  var from = dateInput(q.from, 'from');
  var to = dateInput(q.to, 'to');
  var game = select(q.game || '*', [{ value: '*', label: t('f.allGames') }, { value: '~', label: t('f.mainGame') }], { name: 'game', label: t('f.game') });
  loadGames().then(function (list) { list.forEach(function (g) { if (g) game.appendChild(h('option', { value: g, text: g, selected: g === q.game })); }); });
  var platform = input(q.platform || '', { ltr: true, name: 'platform', placeholder: t('f.any') });
  var version = input(q.appVersion || '', { ltr: true, name: 'appVersion', placeholder: t('f.any') });
  var flagged = check(q.includeFlagged === 'true', t('f.flagged'), 'includeFlagged');
  var apply = function (extra) {
    setQuery(Object.assign({}, q, { from: from.value, to: to.value, game: game.value === '*' ? '' : game.value, platform: platform.value.trim(), appVersion: version.value.trim(), includeFlagged: flagged.input.checked ? 'true' : '' }, extra || {}));
  };
  var preset = function (n) {
    return button({ label: t('f.days', { n: n }), data: { 'ins-preset': n }, onclick: function () {
      var e = new Date();
      var f = new Date(e);
      f.setDate(f.getDate() - (n - 1));
      from.value = isoDay(f);
      to.value = isoDay(e);
      apply();
    } });
  };
  return card(h('div', { class: 'ia-filters p-3 flex flex-wrap items-end gap-3', 'data-ins-filters': '' },
    field(t('f.from'), from), field(t('f.to'), to),
    h('div', { class: 'flex gap-1 mb-0.5' }, preset(7), preset(30), preset(90)),
    opts.noGame ? null : field(t('f.game'), game), field(t('f.platform'), platform), field(t('f.version'), version),
    opts.flagged ? flagged : null,
    button({ label: t('f.apply'), variant: 'primary', data: { 'ins-apply': '' }, onclick: function () { apply(); } })));
}

function loadingInto(el) {
  clear(el).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
}

function failInto(el, err) {
  clear(el).appendChild(notice('error', insErrText(err)));
}

function storeNote(kind) {
  return h('p', { class: 'text-xs text-secondary-500', 'data-ins-store': kind || '', text: t('store', { s: t('st.' + (kind || 'nakama')) }) });
}

function dataTable(columns, rows, name) {
  if (!rows.length) return h('p', { class: 'text-sm text-secondary-500 p-2', 'data-ins-empty': name || '', text: t('empty') });
  var el = table(columns.map(function (c) { return { label: c }; }), rows.map(function (r, i) { return { id: String(i), cells: r }; }));
  if (name) el.setAttribute('data-ins-table', name);
  return el;
}
