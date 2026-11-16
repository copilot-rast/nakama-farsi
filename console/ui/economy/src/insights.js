ICONS['insights'] = '<path d="M3 3v18h18"/><path d="M7 15l4-4 3 3 5-6"/>';
ICONS['insights-overview'] = ICONS['insights'];
ICONS['insights-sessions'] = '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>';
ICONS['heatmaps'] = '<path d="M12 22c4-3.5 7-7 7-11a7 7 0 0 0-14 0c0 4 3 7.5 7 11z"/><circle cx="12" cy="11" r="2.5"/>';
ICONS['insights-settings'] = '<path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6"/>';

var IN_SERIES = { light: ['#2a78d6', '#eb6834', '#1baf7a'], dark: ['#3987e5', '#d95926', '#199e70'] };
var IN_HEAT = ['#fde9b6', '#fcc66b', '#f7953a', '#e8602a', '#c63a24', '#8f1d21'];

function inDark() { return document.documentElement.classList.contains('dark'); }
function inColor(i) { return IN_SERIES[inDark() ? 'dark' : 'light'][i % 3]; }

function inSv(tagName, attrs) {
  var el = document.createElementNS(SVG_NS, tagName);
  for (var k in attrs || {}) if (attrs[k] != null) el.setAttribute(k, attrs[k]);
  return el;
}

function inNiceMax(v) {
  if (v <= 0) return 4;
  var p = Math.pow(10, Math.floor(Math.log10(v)));
  var n = v / p;
  var m = n <= 1 ? 1 : n <= 2 ? 2 : n <= 2.5 ? 2.5 : n <= 5 ? 5 : 10;
  return m * p;
}

function inCompact(n) {
  var a = Math.abs(n);
  if (a >= 1e6) return fmtNum(Math.round(n / 1e5) / 10) + t('in.unit.m');
  if (a >= 1e4) return fmtNum(Math.round(n / 1e3)) + t('in.unit.k');
  return fmtNum(Math.round(n * 10) / 10);
}

function inDay(d) {
  var x = new Date(d + 'T12:00:00');
  if (isNaN(x.getTime())) return d;
  return new Intl.DateTimeFormat(locale(), { month: 'short', day: 'numeric' }).format(x);
}

function inDur(sec) {
  if (sec === null || sec === undefined) return '—';
  var s = Math.round(Number(sec));
  var m = Math.floor(s / 60);
  var hh = Math.floor(m / 60);
  var txt = hh ? hh + ':' + String(m % 60).padStart(2, '0') + ':' + String(s % 60).padStart(2, '0') : m + ':' + String(s % 60).padStart(2, '0');
  return localDigits(txt);
}

function inPct(r) {
  if (r === null || r === undefined) return '—';
  return fmtNum(Math.round(r * 1000) / 10, { maximumFractionDigits: 1 }) + (lang() === 'fa' ? '٪' : '%');
}

function inLineChart(opts) {
  var wrap = h('div', { class: 'in-chart', dir: 'ltr', 'data-in-chart': opts.name || '' });
  if (opts.series.length > 1) {
    wrap.appendChild(h('div', { class: 'in-legend', dir: document.documentElement.getAttribute('dir') || 'ltr' }, opts.series.map(function (s, i) {
      return h('span', { 'data-in-legend': s.key }, h('span', { class: 'in-swatch', style: 'background:' + inColor(i) }), s.name);
    })));
  }
  var holder = h('div', { style: 'position:relative' });
  wrap.appendChild(holder);
  var tip = h('div', { class: 'in-tip', hidden: true });
  holder.appendChild(tip);
  var lastW = 0;
  function draw() {
    var W = Math.max(260, Math.round(holder.clientWidth || wrap.clientWidth || 600));
    if (W === lastW && holder.querySelector('svg')) return;
    lastW = W;
    var H = opts.height || 200;
    var padL = 44, padR = 12, padT = 10, padB = 26;
    var n = opts.labels.length;
    var max = 0;
    opts.series.forEach(function (s) { s.values.forEach(function (v) { if (v > max) max = v; }); });
    var top = inNiceMax(max);
    var x = function (i) { return padL + (n <= 1 ? (W - padL - padR) / 2 : i * (W - padL - padR) / (n - 1)); };
    var y = function (v) { return padT + (H - padT - padB) * (1 - v / top); };
    var svg = inSv('svg', { width: W, height: H, viewBox: '0 0 ' + W + ' ' + H, role: 'img', 'aria-label': opts.title || '' });
    for (var g = 0; g <= 4; g++) {
      var gv = top * g / 4;
      svg.appendChild(inSv('line', { class: 'in-grid', x1: padL, x2: W - padR, y1: y(gv), y2: y(gv) }));
      var lab = inSv('text', { class: 'in-axis', x: padL - 6, y: y(gv) + 3, 'text-anchor': 'end' });
      lab.textContent = opts.format ? opts.format(gv, true) : inCompact(gv);
      svg.appendChild(lab);
    }
    var every = Math.max(1, Math.ceil(n / Math.max(2, Math.floor((W - padL - padR) / 64))));
    for (var i = 0; i < n; i += every) {
      var tl = inSv('text', { class: 'in-axis', x: x(i), y: H - 8, 'text-anchor': 'middle' });
      tl.textContent = inDay(opts.labels[i]);
      svg.appendChild(tl);
    }
    opts.series.forEach(function (s, si) {
      var color = inColor(si);
      var d = s.values.map(function (v, k) { return (k ? 'L' : 'M') + x(k).toFixed(1) + ' ' + y(v).toFixed(1); }).join(' ');
      if (si === 0 && n > 1) svg.appendChild(inSv('path', { d: d + ' L' + x(n - 1).toFixed(1) + ' ' + y(0) + ' L' + x(0).toFixed(1) + ' ' + y(0) + ' Z', fill: color, 'fill-opacity': 0.1, stroke: 'none' }));
      svg.appendChild(inSv('path', { class: 'in-line', d: d, stroke: color, 'data-in-series': s.key }));
      if (n === 1) svg.appendChild(inSv('circle', { cx: x(0), cy: y(s.values[0]), r: 4, fill: color }));
    });
    var cross = inSv('line', { class: 'in-cross', y1: padT, y2: H - padB, visibility: 'hidden' });
    svg.appendChild(cross);
    var dots = opts.series.map(function (s, si) {
      var c = inSv('circle', { r: 4, fill: inColor(si), stroke: 'var(--in-surface, #fff)', 'stroke-width': 2, visibility: 'hidden' });
      svg.appendChild(c);
      return c;
    });
    var hit = inSv('rect', { x: padL, y: padT, width: W - padL - padR, height: H - padT - padB, fill: 'transparent' });
    svg.appendChild(hit);
    function move(ev) {
      var rect = svg.getBoundingClientRect();
      var px = (ev.touches ? ev.touches[0].clientX : ev.clientX) - rect.left;
      var k = n <= 1 ? 0 : Math.round((px - padL) / ((W - padL - padR) / (n - 1)));
      k = Math.max(0, Math.min(n - 1, k));
      cross.setAttribute('x1', x(k)); cross.setAttribute('x2', x(k)); cross.setAttribute('visibility', 'visible');
      clear(tip);
      tip.appendChild(h('div', { class: 'font-medium mb-1', text: inDay(opts.labels[k]) }));
      opts.series.forEach(function (s, si) {
        dots[si].setAttribute('cx', x(k)); dots[si].setAttribute('cy', y(s.values[k])); dots[si].setAttribute('visibility', 'visible');
        tip.appendChild(h('div', {}, h('span', { class: 'in-swatch', style: 'background:' + inColor(si) }), s.name + ': ' + (opts.format ? opts.format(s.values[k]) : fmtNum(s.values[k]))));
      });
      tip.hidden = false;
      var left = x(k) + 12;
      if (left + 150 > W) left = x(k) - 12 - Math.min(180, tip.offsetWidth || 150);
      tip.style.left = Math.max(0, left) + 'px';
      tip.style.top = padT + 'px';
    }
    function leave() {
      tip.hidden = true;
      cross.setAttribute('visibility', 'hidden');
      dots.forEach(function (d) { d.setAttribute('visibility', 'hidden'); });
    }
    hit.addEventListener('mousemove', move);
    hit.addEventListener('touchstart', move, { passive: true });
    hit.addEventListener('mouseleave', leave);
    var old = holder.querySelector('svg');
    if (old) old.remove();
    holder.insertBefore(svg, tip);
  }
  requestAnimationFrame(draw);
  if (window.ResizeObserver) {
    var ro = new ResizeObserver(function () { if (!wrap.isConnected) { ro.disconnect(); return; } draw(); });
    ro.observe(holder);
  }
  return wrap;
}

function inBars(items, name) {
  var max = 0;
  items.forEach(function (it) { if (it.value > max) max = it.value; });
  var box = h('div', { class: 'in-bars', 'data-in-bars': name });
  if (!items.length) {
    box.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('in.noData') }));
    return box;
  }
  items.forEach(function (it) {
    var pct = max ? Math.max(2, Math.round(it.value / max * 100)) : 0;
    box.appendChild(h('div', { class: 'in-bar-row', 'data-in-bar': it.key, title: it.label + ' · ' + fmtNum(it.value) },
      h('span', { class: 'in-bar-label', dir: 'auto', text: it.label }),
      h('div', { class: 'in-bar-track' }, h('div', { class: 'in-bar-fill', style: 'width:' + pct + '%' })),
      h('span', { class: 'in-bar-value', text: fmtNum(it.value) + (it.sub ? ' · ' + it.sub : '') })));
  });
  return box;
}

function inTile(label, value, sub, id) {
  return card(h('div', { class: 'p-4 flex flex-col gap-1 h-full', 'data-in-kpi': id },
    h('div', { class: 'text-sm font-medium text-secondary-700 dark:text-secondary-200', text: label }),
    h('div', { class: 'text-2xl font-semibold tabular-nums', text: value }),
    sub ? h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: sub }) : null));
}

function inRangeFilter(state, onchange) {
  return filterSelect(t('filter.of.days'), state.days, [7, 14, 30, 90, 180, 365].map(function (n) { return { value: String(n), label: t('of.daysN', { n: localDigits(String(n)) }) }; }), function (v) { state.days = v || '30'; onchange(); });
}

function inIsoDay(d) {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

function inSpan(days) {
  var to = new Date();
  var from = new Date();
  from.setDate(from.getDate() - (Number(days || 30) - 1));
  return { from: inIsoDay(from), to: inIsoDay(to) };
}

var IN_DIMS = { platforms: [], versions: [] };

function insightsPage(container, query) {
  var frame = pageFrame(t('nav.insights-overview'), t('desc.insights-overview'));
  clear(container).appendChild(frame.el);
  var state = { days: query.days || '30', platform: query.platform || '', appVersion: query.appVersion || '' };
  var filters = h('div', { class: 'eco-filters flex flex-wrap items-center gap-2 pb-2', 'data-in-filters': '' });
  var body = h('div', { class: 'flex flex-col gap-4', 'data-in-overview': '' });
  frame.content.appendChild(filters);
  frame.content.appendChild(body);
  function route() {
    var q = {};
    if (state.days !== '30') q.days = state.days;
    if (state.platform) q.platform = state.platform;
    if (state.appVersion) q.appVersion = state.appVersion;
    replaceHash(hashFor('insights-overview', null, q));
    load();
  }
  function paintFilters() {
    clear(filters);
    filters.appendChild(inRangeFilter(state, route));
    filters.appendChild(filterSelect(t('in.platform'), state.platform, IN_DIMS.platforms.map(function (p) { return { value: p, label: p }; }), function (v) { state.platform = v; route(); }));
    filters.appendChild(filterSelect(t('in.version'), state.appVersion, IN_DIMS.versions.map(function (p) { return { value: p, label: p }; }), function (v) { state.appVersion = v; route(); }));
    filters.appendChild(button({ variant: 'muted', size: 'icon', icon: 'refresh', iconClass: 'h-3 w-3', title: t('refresh'), onclick: load }));
  }
  function skeleton() {
    clear(body);
    var g = h('div', { class: 'eco-stat-grid' });
    for (var i = 0; i < 6; i++) g.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem'), bar('3rem', 'h-7'))));
    body.appendChild(g);
    body.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-2' }, bar('30%'), bar('100%', 'h-40'))));
  }
  function render(res) {
    clear(body);
    var tt = res.totals || {};
    var act = res.active || {};
    (res.platforms || []).forEach(function (p) { if (p.key && IN_DIMS.platforms.indexOf(p.key) === -1) IN_DIMS.platforms.push(p.key); });
    (res.versions || []).forEach(function (p) { if (p.key && IN_DIMS.versions.indexOf(p.key) === -1) IN_DIMS.versions.push(p.key); });
    paintFilters();
    if (res.sessionSource === 'socket') body.appendChild(notice('info', t('in.source.socket')));
    var kpis = h('div', { class: 'eco-stat-grid', 'data-in-kpis': '' },
      inTile(t('in.dau'), fmtNum(act.dau || 0), t('in.dau.sub', { day: inDay(act.day) }), 'dau'),
      inTile(t('in.wau'), fmtNum(act.wau || 0), null, 'wau'),
      inTile(t('in.mau'), fmtNum(act.mau || 0), null, 'mau'),
      inTile(t('in.new'), fmtNum(tt.newPlayers || 0), t('in.avgDau', { n: fmtNum(tt.avgDau || 0) }), 'new'),
      inTile(t('in.sessions'), fmtNum(tt.sessions || 0), t('in.perPlayer', { n: tt.sessionsPerPlayer == null ? '—' : fmtNum(tt.sessionsPerPlayer) }), 'sessions'),
      inTile(t('in.length'), inDur(tt.avgSeconds), t('in.median', { v: inDur(tt.medianSeconds) }), 'length'),
      inTile(t('in.open'), fmtNum(tt.open || 0), t('in.open.sub'), 'open'));
    body.appendChild(kpis);
    var series = res.series || [];
    var labels = series.map(function (d) { return d.day; });
    body.appendChild(section('in.chart.players', inLineChart({ name: 'players', title: t('in.chart.players'), labels: labels, series: [
      { key: 'dau', name: t('in.dau'), values: series.map(function (d) { return d.dau; }) },
      { key: 'new', name: t('in.new'), values: series.map(function (d) { return d.newPlayers; }) }] })));
    body.appendChild(h('div', { class: 'eco-detail-grid' },
      section('in.chart.sessions', inLineChart({ name: 'sessions', title: t('in.chart.sessions'), labels: labels, series: [{ key: 'sessions', name: t('in.sessions'), values: series.map(function (d) { return d.sessions; }) }] })),
      section('in.chart.length', inLineChart({ name: 'length', title: t('in.chart.length'), labels: labels, format: function (v) { return inDur(v); }, series: [{ key: 'length', name: t('in.length'), values: series.map(function (d) { return d.avgSeconds; }) }] }))));
    var ret = res.retention || [];
    var retRow = h('div', { class: 'eco-stat-grid mb-4', 'data-in-retention': '' }, ret.map(function (r) {
      return inTile(t('in.dN', { n: localDigits(String(r.day)) }), inPct(r.rate), t('in.ret.sub', { back: fmtNum(r.returned), cohort: fmtNum(r.cohort) }), 'd' + r.day);
    }));
    var cohorts = res.cohorts || [];
    var cohortBox;
    if (cohorts.length) {
      var cols = [{ label: t('in.cohort.day') }, { label: t('in.cohort.size'), width: 90 }].concat((res.retentionDays || []).map(function (n) { return { label: t('in.dN', { n: localDigits(String(n)) }), width: 90 }; }));
      cohortBox = table(cols, cohorts.map(function (c) {
        return { item: c, cells: [h('span', { class: 'whitespace-nowrap', text: inDay(c.day) }), h('span', { class: 'tabular-nums', text: fmtNum(c.size) })].concat((c.cells || []).map(function (cell) {
          if (cell.rate === null || cell.rate === undefined) return h('span', { class: 'text-secondary-400', text: '—' });
          var a = Math.min(0.85, 0.1 + cell.rate * 0.9);
          return h('span', { class: 'in-cohort-cell tabular-nums', 'data-in-cohort': c.day + ':' + cell.day, style: '--in-a:' + a.toFixed(2), title: t('in.ret.sub', { back: fmtNum(cell.returned), cohort: fmtNum(c.size) }), text: inPct(cell.rate) });
        })) };
      }));
    } else {
      cohortBox = h('p', { class: 'text-sm text-secondary-500', text: t('in.noData') });
    }
    body.appendChild(section('in.retention', h('div', {},
      h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: t('in.retention.' + (res.retentionMode || 'classic')) }),
      retRow, cohortBox)));
    body.appendChild(h('div', { class: 'eco-detail-grid' },
      section('in.byPlatform', inBars((res.platforms || []).map(function (p) { return { key: p.key || '-', label: p.key || t('in.unknown'), value: p.players, sub: t('in.sessionsN', { n: fmtNum(p.sessions) }) }; }), 'platforms')),
      section('in.byVersion', inBars((res.versions || []).map(function (p) { return { key: p.key || '-', label: p.key || t('in.unknown'), value: p.players, sub: t('in.sessionsN', { n: fmtNum(p.sessions) }) }; }), 'versions'))));
  }
  function load() {
    skeleton();
    var sp = inSpan(state.days);
    baseApi('GET', '/insights/overview' + qs({ from: sp.from, to: sp.to, platform: state.platform, appVersion: state.appVersion })).then(render, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  paintFilters();
  load();
}

function insightsSessionsPage(container, query) {
  var frame = pageFrame(t('nav.insights-sessions'), t('desc.insights-sessions'));
  clear(container).appendChild(frame.el);
  var state = { page: Number(query.page) || 1, limit: 20, total: 0, status: query.status || '', platform: query.platform || '', userId: query.userId || '' };
  var filters = h('div', { class: 'eco-filters flex flex-wrap items-center gap-2 pb-2' });
  var pagerBox = h('div', {});
  var listBox = h('div', { 'data-in-sessions': '' });
  frame.content.appendChild(filters);
  frame.content.appendChild(listBox);
  var cols = [{ label: t('in.col.player'), skeleton: '40%' }, { label: t('in.col.started'), width: 150 }, { label: t('in.length'), width: 90 }, { label: t('in.col.device'), width: 200 }, { label: t('col.status'), width: 110 }];
  function route() {
    var q = {};
    ['status', 'platform', 'userId'].forEach(function (k) { if (state[k]) q[k] = state[k]; });
    if (state.page > 1) q.page = state.page;
    replaceHash(hashFor('insights-sessions', null, q));
    load();
  }
  var search = searchInput(state.userId, t('in.f.userId'), debounce(function (e) { state.userId = toLatinDigits(e.target.value).trim(); state.page = 1; route(); }, 400));
  filters.appendChild(search);
  filters.appendChild(filterSelect(t('col.status'), state.status, [{ value: 'open', label: t('in.status.open') }, { value: 'ended', label: t('in.status.ended') }], function (v) { state.status = v; state.page = 1; route(); }));
  filters.appendChild(filterSelect(t('in.platform'), state.platform, IN_DIMS.platforms.map(function (p) { return { value: p, label: p }; }), function (v) { state.platform = v; state.page = 1; route(); }));
  function load() {
    clear(listBox).appendChild(table(cols, skeletonRows(cols, 5)));
    baseApi('GET', '/insights/sessions' + qs({ page: state.page, limit: state.limit, status: state.status, platform: state.platform, userId: state.userId })).then(function (res) {
      state.total = res.total || 0;
      var items = res.items || [];
      clear(listBox);
      clear(pagerBox).appendChild(pager(state, function (p) { state.page = p; route(); }, load));
      listBox.appendChild(h('div', { class: 'flex justify-end pb-2' }, pagerBox));
      if (!items.length) { listBox.appendChild(card(emptyBlock(t('in.sessions.empty'), t('in.sessions.empty.hint')))); return; }
      listBox.appendChild(table(cols, items.map(function (s) {
        var length = s.seconds !== null && s.seconds !== undefined ? s.seconds : Math.max(0, (new Date(s.lastSeenAt) - new Date(s.startedAt)) / 1000);
        return { item: s, cells: [
          h('div', { class: 'flex flex-col min-w-0' }, playerLink(s.userId, s.displayName || s.username), h('span', { class: 'text-xs text-secondary-500 truncate', dir: 'ltr', text: s.sessionId })),
          h('span', { class: 'text-sm whitespace-nowrap', text: fmtDate(s.startedAt, true) }),
          h('span', { class: 'tabular-nums', text: inDur(length) }),
          h('span', { class: 'text-sm truncate', dir: 'ltr', text: [s.platform, s.appVersion, s.device, s.os].filter(Boolean).join(' · ') || '—' }),
          h('div', { class: 'flex flex-wrap gap-1' }, s.endedAt ? tag(t('in.status.ended'), 'secondary') : tag(t('in.status.open'), 'success'), s.source === 'socket' ? tag(t('in.src.socket'), 'primary') : null)
        ] };
      })));
    }, function (err) { clear(listBox).appendChild(errorBlock(err, load)); });
  }
  load();
  return { reload: load };
}

function inHeatColor(f) {
  var i = Math.max(0, Math.min(IN_HEAT.length - 1, Math.floor(f * IN_HEAT.length)));
  return IN_HEAT[i];
}

function inHeatCanvas(data, opts) {
  var wrap = h('div', { class: 'in-heat', dir: 'ltr', 'data-in-heat': data.map.key });
  var canvas = h('canvas', { class: 'in-heat-canvas', role: 'img', 'aria-label': t('in.hm.aria', { map: data.map.name || data.map.key, n: fmtNum((data.totals || {}).points || 0) }) });
  var tip = h('div', { class: 'in-tip', hidden: true });
  wrap.appendChild(canvas);
  wrap.appendChild(tip);
  var size = data.size;
  var cells = data.cells || [];
  var mb = data.map.bounds || {};
  var ib = data.map.imageBounds || {};
  var minX = Infinity, maxX = -Infinity, minZ = Infinity, maxZ = -Infinity;
  cells.forEach(function (c) { minX = Math.min(minX, c[0] * size); maxX = Math.max(maxX, (c[0] + 1) * size); minZ = Math.min(minZ, c[1] * size); maxZ = Math.max(maxZ, (c[1] + 1) * size); });
  (opts.trace || []).forEach(function (p) { p.points.forEach(function (q) { minX = Math.min(minX, q[1]); maxX = Math.max(maxX, q[1]); minZ = Math.min(minZ, q[2]); maxZ = Math.max(maxZ, q[2]); }); });
  if (ib.minX !== undefined && ib.minX !== null) { minX = ib.minX; maxX = ib.maxX; minZ = ib.minZ; maxZ = ib.maxZ; }
  else if (!isFinite(minX) && mb.minX !== null && mb.minX !== undefined) { minX = mb.minX; maxX = mb.maxX; minZ = mb.minZ; maxZ = mb.maxZ; }
  if (!isFinite(minX) || maxX <= minX) { minX = 0; maxX = 100; }
  if (!isFinite(minZ) || maxZ <= minZ) { minZ = 0; maxZ = 100; }
  var spanX = maxX - minX, spanZ = maxZ - minZ;
  var maxN = 1, maxKmh = 1;
  cells.forEach(function (c) { maxN = Math.max(maxN, c[2]); maxKmh = Math.max(maxKmh, c[3]); });
  var bg = null;
  if (data.map.backgroundUrl) {
    bg = new Image();
    bg.crossOrigin = 'anonymous';
    bg.onload = function () { draw(); };
    bg.src = data.map.backgroundUrl;
  }
  var geo = null;
  function draw() {
    var W = Math.max(240, Math.round(wrap.clientWidth || 600));
    var H = Math.round(Math.min(W * spanZ / spanX, Math.max(240, window.innerHeight * 0.7)));
    var scale = Math.min(W / spanX, H / spanZ);
    var dpr = window.devicePixelRatio || 1;
    canvas.width = Math.round(W * dpr);
    canvas.height = Math.round(H * dpr);
    canvas.style.width = W + 'px';
    canvas.style.height = H + 'px';
    var ctx = canvas.getContext('2d');
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    var offX = (W - spanX * scale) / 2, offY = (H - spanZ * scale) / 2;
    var px = function (x) { return offX + (x - minX) * scale; };
    var py = function (z) { return offY + (maxZ - z) * scale; };
    geo = { px: px, py: py, scale: scale, offX: offX, offY: offY };
    ctx.fillStyle = inDark() ? '#111827' : '#f3f4f6';
    ctx.fillRect(0, 0, W, H);
    if (bg && bg.complete && bg.naturalWidth) {
      ctx.globalAlpha = 0.85;
      ctx.drawImage(bg, px(minX), py(maxZ), spanX * scale, spanZ * scale);
      ctx.globalAlpha = 1;
    }
    var speed = opts.layer === 'speed';
    var cw = Math.max(1, size * scale);
    cells.forEach(function (c) {
      var f = speed ? c[3] / maxKmh : Math.log(1 + c[2]) / Math.log(1 + maxN);
      ctx.fillStyle = inHeatColor(f);
      ctx.globalAlpha = 0.35 + 0.6 * f;
      ctx.fillRect(px(c[0] * size), py((c[1] + 1) * size), cw + 0.5, cw + 0.5);
    });
    ctx.globalAlpha = 1;
    (opts.trace || []).forEach(function (p, i) {
      ctx.strokeStyle = inColor(i);
      ctx.lineWidth = 2;
      ctx.beginPath();
      p.points.forEach(function (q, k) { if (k) ctx.lineTo(px(q[1]), py(q[2])); else ctx.moveTo(px(q[1]), py(q[2])); });
      ctx.stroke();
    });
  }
  canvas.addEventListener('mousemove', function (ev) {
    if (!geo) return;
    var rect = canvas.getBoundingClientRect();
    var wx = minX + (ev.clientX - rect.left - geo.offX) / geo.scale;
    var wz = maxZ - (ev.clientY - rect.top - geo.offY) / geo.scale;
    var cx = Math.floor(wx / size), cz = Math.floor(wz / size);
    var hit = null;
    for (var i = 0; i < cells.length; i++) if (cells[i][0] === cx && cells[i][1] === cz) { hit = cells[i]; break; }
    if (!hit) { tip.hidden = true; return; }
    clear(tip);
    tip.appendChild(h('div', { class: 'font-medium', text: 'x ' + localDigits(String(Math.round(wx))) + ' · z ' + localDigits(String(Math.round(wz))) }));
    tip.appendChild(h('div', { text: t('in.hm.cell', { n: fmtNum(hit[2]), avg: fmtNum(hit[3]), max: fmtNum(hit[4]) }) }));
    tip.hidden = false;
    tip.style.left = Math.min(ev.clientX - rect.left + 12, rect.width - 160) + 'px';
    tip.style.top = Math.max(0, ev.clientY - rect.top - 40) + 'px';
  });
  canvas.addEventListener('mouseleave', function () { tip.hidden = true; });
  requestAnimationFrame(draw);
  if (window.ResizeObserver) {
    var last = 0;
    var ro = new ResizeObserver(function () { if (!wrap.isConnected) { ro.disconnect(); return; } if (wrap.clientWidth !== last) { last = wrap.clientWidth; draw(); } });
    ro.observe(wrap);
  }
  return wrap;
}

function inMapPanel(m, onSaved) {
  var v = { name: m.name || m.key, cellSize: m.cellSize ? String(m.cellSize) : '', backgroundUrl: m.backgroundUrl || '', enabled: m.enabled !== false, ib: m.imageBounds || {} };
  var body = h('div', { 'data-in-map-form': m.key });
  var footer = h('div', { class: 'flex gap-2 w-full items-center justify-end' });
  ctOpenPanel('heatmap/' + m.key, t('in.hm.edit', { map: m.key }), body, footer, closePanel);
  var top = h('div', {});
  var name = textInput(v.name, { label: t('in.hm.name'), oninput: function (x) { v.name = x; } });
  var cell = numberInput(v.cellSize, { int: false, label: t('in.hm.cellSize'), oninput: function (x) { v.cellSize = x; } });
  cell.setAttribute('data-in-map', 'cellSize');
  var bg = textInput(v.backgroundUrl, { ltr: true, label: t('in.hm.bg'), placeholder: 'https://', oninput: function (x) { v.backgroundUrl = x.trim(); } });
  var en = toggle(v.enabled, t('in.hm.enabled'), function (x) { v.enabled = x; });
  var bounds = {};
  var brow = h('div', { class: 'grid grid-cols-2 gap-2' }, ['minX', 'maxX', 'minZ', 'maxZ'].map(function (k) {
    bounds[k] = numberInput(v.ib[k] === undefined || v.ib[k] === null ? '' : v.ib[k], { int: false, label: k, placeholder: k, oninput: function () {} });
    return field(k, bounds[k]);
  }));
  body.appendChild(h('div', {}, top,
    field(t('in.hm.name'), name),
    field(t('in.hm.cellSize'), cell, t('in.hm.cellSize.hint', { n: fmtNum(m.effectiveCellSize) }), { optional: true }),
    field(t('in.hm.bg'), bg, t('in.hm.bg.hint'), { optional: true }),
    h('p', { class: 'font-medium mb-1', text: t('in.hm.bounds') }),
    h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-2', text: t('in.hm.bounds.hint') }),
    brow,
    h('div', { class: 'mb-5' }, en)));
  var del = button({ variant: 'danger', size: 'footer', icon: 'trash', label: t('delete'), onclick: function () {
    confirmModal({ title: t('in.hm.delete.title'), text: t('in.hm.delete.text', { map: m.key }), danger: true, ok: t('delete') }).then(function (yes) {
      if (!yes) return;
      baseApi('DELETE', '/insights/heatmaps/' + encodeURIComponent(m.key)).then(function () { toast('success', t('in.hm.deleted')); closePanel(); onSaved(null); }, function (err) { toast('error', errorText(err)); });
    });
  } });
  var save = button({ variant: 'primary', size: 'footer', label: t('save'), onclick: function () {
    clear(top);
    var ib = {};
    var any = false;
    var bad = false;
    ['minX', 'maxX', 'minZ', 'maxZ'].forEach(function (k) {
      var n = parseNumber(bounds[k].value, false);
      if (n !== null && !isNaN(n)) { ib[k] = n; any = true; } else if (n !== null) bad = true;
    });
    var size = parseNumber(v.cellSize, false);
    if (bad || (size !== null && isNaN(size))) { top.appendChild(h('div', { class: 'mb-4' }, notice('error', t('err.number')))); return; }
    setBusy(save, true, t('saving'));
    baseApi('PUT', '/insights/heatmaps/' + encodeURIComponent(m.key), { name: v.name.trim(), cellSize: size || 0, backgroundUrl: v.backgroundUrl, enabled: v.enabled, imageBounds: any ? ib : {} }).then(function (res) {
      toast('success', t('saved'));
      closePanel();
      onSaved(res.item);
    }, function (err) { setBusy(save, false); top.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err)))); });
  } });
  save.setAttribute('data-in-map-save', '');
  if (aclCan('SETTINGS', 1)) { footer.appendChild(del); footer.appendChild(h('span', { class: 'flex-1' })); footer.appendChild(save); }
}

function heatmapsPage(container, query) {
  var frame = pageFrame(t('nav.heatmaps'), t('desc.heatmaps'));
  clear(container).appendChild(frame.el);
  var state = { map: query.map || '', days: query.days || '30', layer: query.layer || 'density', trace: null };
  var maps = [];
  var filters = h('div', { class: 'eco-filters flex flex-wrap items-center gap-2 pb-2', 'data-in-hm-filters': '' });
  var stats = h('div', { class: 'eco-stat-grid mb-4', 'data-in-hm-stats': '' });
  var view = h('div', { class: 'mb-4' });
  var uploads = h('div', { 'data-in-hm-uploads': '' });
  frame.content.appendChild(filters);
  frame.content.appendChild(stats);
  frame.content.appendChild(view);
  frame.content.appendChild(uploads);
  function route() {
    var q = {};
    if (state.map) q.map = state.map;
    if (state.days !== '30') q.days = state.days;
    if (state.layer !== 'density') q.layer = state.layer;
    replaceHash(hashFor('heatmaps', null, q));
  }
  function current() { return maps.filter(function (m) { return m.key === state.map; })[0]; }
  function paintFilters() {
    clear(filters);
    var sel = selectInput(state.map, maps.map(function (m) { return { value: m.key, label: (m.name && m.name !== m.key ? m.name + ' · ' : '') + m.key }; }), function (v) { state.map = v; state.trace = null; route(); loadMap(); }, { label: t('in.hm.map'), placeholder: false });
    sel.setAttribute('data-in-hm-map', '');
    filters.appendChild(h('label', { class: 'x-label block relative x-input eco-ct-field', style: INPUT_VARS }, sel));
    filters.appendChild(inRangeFilter(state, function () { route(); loadMap(); }));
    var seg = segmented([{ value: 'density', label: t('in.hm.density') }, { value: 'speed', label: t('in.hm.speed') }], state.layer, function (v) { state.layer = v; route(); loadMap(); }, t('in.hm.layer'));
    seg.setAttribute('data-in-hm-layer', '');
    filters.appendChild(seg);
    var m = current();
    if (m && aclCan('SETTINGS', 1)) {
      var edit = button({ variant: 'ghost', size: 'sm', icon: 'gear', label: t('in.hm.settings'), onclick: function () { inMapPanel(m, function () { loadList(); }); } });
      edit.setAttribute('data-in-hm-edit', '');
      filters.appendChild(edit);
    }
  }
  function loadList() {
    clear(view).appendChild(card(h('div', { class: 'p-4' }, bar('100%', 'h-64'))));
    baseApi('GET', '/insights/heatmaps').then(function (res) {
      maps = res.items || [];
      if (!res.enabled) frame.content.insertBefore(h('div', { class: 'mb-3' }, notice('warning', t('in.hm.off'))), filters);
      if (!maps.length) {
        clear(filters);
        clear(stats);
        clear(uploads);
        clear(view).appendChild(card(emptyBlock(t('in.hm.empty'), t('in.hm.empty.hint'))));
        return;
      }
      if (!current()) state.map = maps[0].key;
      paintFilters();
      loadMap();
    }, function (err) { clear(view).appendChild(errorBlock(err, loadList)); });
  }
  function loadMap() {
    var sp = inSpan(state.days);
    clear(view).appendChild(card(h('div', { class: 'p-4' }, bar('100%', 'h-64'))));
    var key = encodeURIComponent(state.map);
    Promise.all([
      baseApi('GET', '/insights/heatmaps/' + key + '/cells' + qs({ from: sp.from, to: sp.to })),
      baseApi('GET', '/insights/heatmaps/' + key + '/uploads' + qs({ limit: 20 }))
    ]).then(function (all) {
      var d = all[0];
      var tt = d.totals || {};
      clear(stats);
      stats.appendChild(inTile(t('in.hm.points'), fmtNum(tt.points || 0), t('in.hm.cells', { n: fmtNum(tt.cells || 0), size: fmtNum(d.size) }), 'points'));
      stats.appendChild(inTile(t('in.hm.uploads'), fmtNum(tt.uploads || 0), null, 'uploads'));
      stats.appendChild(inTile(t('in.hm.players'), fmtNum(tt.players || 0), null, 'players'));
      clear(view);
      var legend = h('div', { class: 'flex flex-wrap items-center gap-2 text-xs text-secondary-500 mt-2' },
        h('span', { text: t(state.layer === 'speed' ? 'in.hm.legend.slow' : 'in.hm.legend.low') }),
        h('span', { class: 'in-heat-scale', 'aria-hidden': 'true' }),
        h('span', { text: t(state.layer === 'speed' ? 'in.hm.legend.fast' : 'in.hm.legend.high') }),
        state.trace ? h('span', { class: 'flex-1' }) : null,
        state.trace ? button({ variant: 'link', size: 'sm', label: t('in.hm.trace.clear'), onclick: function () { state.trace = null; loadMap(); } }) : null);
      var notes = [];
      if ((d.sizes || []).length > 1) notes.push(notice('info', t('in.hm.sizes')));
      if (!(d.cells || []).length) notes.push(notice('info', t('in.hm.noPoints')));
      view.appendChild(card(h('div', { class: 'p-4' }, notes, inHeatCanvas(d, { layer: state.layer, trace: state.trace ? state.trace.players : null }), legend,
        state.trace ? h('div', { class: 'flex flex-wrap gap-3 text-xs mt-2', 'data-in-trace': state.trace.id }, state.trace.players.map(function (p, i) { return h('span', {}, h('span', { class: 'in-swatch', style: 'background:' + inColor(i) }), (p.player || '—') + ' · ' + fmtNum(p.points.length)); })) : null)));
      var list = all[1].items || [];
      clear(uploads);
      var cols = [{ label: t('in.col.player'), skeleton: '40%' }, { label: t('in.hm.col.upload'), width: 160 }, { label: t('in.hm.points'), width: 100 }, { label: t('in.length'), width: 90 }, { label: t('in.col.started'), width: 150 }, { label: '', width: 110 }];
      uploads.appendChild(section('in.hm.recent', list.length ? table(cols, list.map(function (u) {
        var show = u.hasTrace ? button({ variant: 'link', size: 'sm', label: t('in.hm.trace'), onclick: function () {
          baseApi('GET', '/insights/heatmaps/uploads/' + encodeURIComponent(u.id) + '/trace').then(function (tr) { state.trace = tr; loadMap(); view.scrollIntoView({ block: 'start' }); }, function (err) { toast('error', errorText(err)); });
        } }) : h('span', { class: 'text-xs text-secondary-400', text: t('in.hm.trace.gone') });
        if (u.hasTrace) show.setAttribute('data-in-trace-btn', u.id);
        return { item: u, cells: [
          playerLink(u.userId, u.displayName || u.username),
          h('span', { dir: 'ltr', class: 'font-mono text-xs truncate', text: u.uploadId + (u.part ? ' #' + u.part : '') }),
          h('span', { class: 'tabular-nums', text: fmtNum(u.points) + (u.dropped ? ' (−' + fmtNum(u.dropped) + ')' : '') }),
          h('span', { class: 'tabular-nums', text: inDur(u.durationMs / 1000) }),
          h('span', { class: 'text-sm whitespace-nowrap', text: fmtDate(u.createdAt, true) }),
          show
        ] };
      })) : h('p', { class: 'text-sm text-secondary-500', text: t('in.noData') })));
    }, function (err) { clear(view).appendChild(errorBlock(err, loadMap)); });
  }
  loadList();
}

var IN_SETTINGS_FIELDS = [
  { name: 'timezone', type: 'text', ltr: true, label: 'in.s.timezone', hint: 'in.s.timezone.hint', required: true },
  { type: 'heading', label: 'in.h.sessions' },
  { name: 'sessions.source', type: 'select', label: 'in.s.source', hint: 'in.s.source.hint', required: true, options: function () { return [{ value: 'client', label: t('in.src.client') }, { value: 'socket', label: t('in.src.socket') }]; } },
  { type: 'row', fields: [
    { name: 'sessions.pingSeconds', type: 'number', label: 'in.s.ping', hint: 'in.s.ping.hint', min: 10, max: 3600, required: true },
    { name: 'sessions.timeoutSeconds', type: 'number', label: 'in.s.timeout', hint: 'in.s.timeout.hint', min: 30, max: 86400, required: true }
  ] },
  { type: 'row', fields: [
    { name: 'sessions.maxSeconds', type: 'number', label: 'in.s.max', hint: 'in.s.max.hint', min: 60, max: 604800, required: true },
    { name: 'sessions.startsPerHour', type: 'number', label: 'in.s.starts', min: 1, max: 10000, required: true }
  ] },
  { name: 'sessions.closeOtherOpen', type: 'bool', inline: true, label: 'in.s.closeOther', hint: 'in.s.closeOther.hint' },
  { name: 'sessions.keepDays', type: 'number', label: 'in.s.keepDays', hint: 'in.s.zero.keep', min: 0, max: 3650, required: true },
  { type: 'heading', label: 'in.h.retention' },
  { type: 'row', fields: [
    { name: 'retentionDays', type: 'text', ltr: true, label: 'in.s.days', hint: 'in.s.days.hint', required: true },
    { name: 'retention.mode', type: 'select', label: 'in.s.mode', required: true, options: function () { return [{ value: 'classic', label: t('in.mode.classic') }, { value: 'rolling', label: t('in.mode.rolling') }]; } }
  ] },
  { name: 'retention.cohortRows', type: 'number', label: 'in.s.cohortRows', min: 1, max: 120, required: true },
  { type: 'heading', label: 'in.h.heatmap' },
  { name: 'heatmap.enabled', type: 'bool', inline: true, label: 'in.s.hm.enabled' },
  { name: 'heatmap.unknownMaps', type: 'select', label: 'in.s.hm.unknown', hint: 'in.s.hm.unknown.hint', required: true, options: function () { return [{ value: 'accept', label: t('in.unknown.accept') }, { value: 'reject', label: t('in.unknown.reject') }]; } },
  { type: 'row', fields: [
    { name: 'heatmap.cellSize', type: 'number', int: false, label: 'in.s.hm.cell', hint: 'in.s.hm.cell.hint', min: 0.1, max: 1000, required: true },
    { name: 'heatmap.maxKmh', type: 'number', int: false, label: 'in.s.hm.kmh', hint: 'in.s.hm.kmh.hint', min: 1, max: 100000, required: true }
  ] },
  { type: 'row', fields: [
    { name: 'heatmap.maxPoints', type: 'number', label: 'in.s.hm.points', min: 10, max: 200000, required: true },
    { name: 'heatmapKB', type: 'number', label: 'in.s.hm.kb', min: 4, max: 8192, required: true }
  ] },
  { name: 'heatmap.uploadsPerHour', type: 'number', label: 'in.s.hm.rate', min: 1, max: 10000, required: true },
  { type: 'row', fields: [
    { name: 'heatmap.keepRawDays', type: 'number', label: 'in.s.hm.raw', hint: 'in.s.hm.raw.hint', min: 0, max: 3650, required: true },
    { name: 'heatmap.keepCellsDays', type: 'number', label: 'in.s.hm.cells', hint: 'in.s.zero.keep', min: 0, max: 3650, required: true }
  ] },
  { name: 'heatmap.maxTracePoints', type: 'number', label: 'in.s.hm.trace', min: 100, max: 500000, required: true }
];

function inSettingsCall(method, path, body) {
  return baseApi(method, path, body).then(function (res) {
    var s = clone((res && res.settings) || {});
    s.retentionDays = localDigits(((s.retention || {}).days || []).join(', '));
    s.heatmapKB = Math.round(((s.heatmap || {}).maxBytes || 0) / 1024);
    return s;
  });
}

function insightsSettingsPage(container) {
  var frame = pageFrame(t('nav.insights-settings'), t('desc.insights-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(inSettingsCall, '/insights/settings', null, IN_SETTINGS_FIELDS, {
    id: 'insights',
    fixPayload: function (p) {
      var days = toLatinDigits(String(p.retentionDays || '')).split(/[^0-9]+/).filter(Boolean).map(Number);
      var out = { timezone: (p.timezone || '').trim(), sessions: p.sessions, retention: Object.assign({}, p.retention, { days: days }), heatmap: Object.assign({}, p.heatmap, { maxBytes: Number(p.heatmapKB) * 1024 }) };
      return out;
    }
  });
  frame.content.appendChild(page.el);
  page.load();
}

CUSTOM_PAGES['insights-overview'] = function (root, route) { insightsPage(root, route.query || {}); };
CUSTOM_PAGES['insights-sessions'] = function (root, route, ctl) { ctl.list = insightsSessionsPage(root, route.query || {}); };
CUSTOM_PAGES['heatmaps'] = function (root, route) { heatmapsPage(root, route.query || {}); };
CUSTOM_PAGES['insights-settings'] = function (root) { insightsSettingsPage(root); };
