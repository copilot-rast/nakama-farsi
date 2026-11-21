function dashFrame(root, page, actions) {
  var frame = pageFrame(page, t('desc.' + page), actions);
  clear(root).appendChild(frame.el);
  return frame;
}

function overviewPage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/overview');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(filterBar(q));
  frame.content.appendChild(body);
  loadingInto(body);
  Promise.all([api('GET', '/insights/activity' + filterQS(q)), api('GET', '/insights/retention' + filterQS(q))]).then(function (res) {
    var d = res[0], r = res[1], a = d.activity;
    clear(body);
    var tiles = [
      { key: 'dau', label: t('k.dau'), value: fmtNum(a.dau) }, { key: 'wau', label: t('k.wau'), value: fmtNum(a.wau) }, { key: 'mau', label: t('k.mau'), value: fmtNum(a.mau) },
      { key: 'stick', label: t('k.stick'), value: fmtPct(d.stickiness) }, { key: 'new', label: t('k.new'), value: fmtNum(a.newUsers) }, { key: 'events', label: t('k.events'), value: fmtNum(a.events) }];
    (r.summary || []).forEach(function (x) { tiles.push({ key: 'd' + x.day, label: t('k.d', { n: x.day }), value: fmtPct(x.rate), sub: fmtNum(x.returned) + ' / ' + fmtNum(x.cohort) }); });
    body.appendChild(kpis(tiles));
    var days = a.days.map(function (x) { return x.day; });
    body.appendChild(panelCard(t('c.dau'), lineChart(days, [{ label: t('c.dau'), values: a.days.map(function (x) { return x.dau; }) }, { label: t('c.new'), values: a.days.map(function (x) { return x.newUsers; }) }], { fmtLabel: fmtDay, name: 'overview-dau' }), null, 'overview-chart'));
    var grid = h('div', { class: 'grid gap-4 lg:grid-cols-2' });
    grid.appendChild(panelCard(t('top.events'), d.topEvents.length ? barList(d.topEvents.map(function (x) { return { label: x.keys[0], value: x.value, tip: [t('c.users') + ': ' + fmtNum(x.users)] }; }), { name: 'top-events' }) : h('p', { class: 'text-sm text-secondary-500', text: t('empty') }),
      csvButton('top-events', function () { return { columns: ['event', 'count', 'users'], rows: d.topEvents.map(function (x) { return [x.keys[0], x.value, x.users]; }) }; }), 'top-events'));
    grid.appendChild(panelCard(t('top.platforms'), d.platforms.filter(function (x) { return x.keys[0]; }).length ? barList(d.platforms.filter(function (x) { return x.keys[0]; }).map(function (x, i) { return { label: x.keys[0] || t('none'), value: x.value, color: seriesColor(i) }; }), { name: 'platforms' }) : h('p', { class: 'text-sm text-secondary-500', text: t('empty') }), null, 'platforms'));
    body.appendChild(grid);
    body.appendChild(storeNote(d.store));
  }, function (err) { failInto(body, err); });
}

function activePage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/active');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(filterBar(q));
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/activity' + filterQS(q)).then(function (d) {
    var a = d.activity;
    clear(body);
    body.appendChild(kpis([{ key: 'dau', label: t('k.dau'), value: fmtNum(a.dau) }, { key: 'wau', label: t('k.wau'), value: fmtNum(a.wau) }, { key: 'mau', label: t('k.mau'), value: fmtNum(a.mau) },
      { key: 'stick', label: t('k.stick'), value: fmtPct(d.stickiness) }, { key: 'users', label: t('k.users'), value: fmtNum(a.users) }, { key: 'new', label: t('k.new'), value: fmtNum(a.newUsers) }]));
    var labels = a.days.map(function (x) { return x.day; });
    body.appendChild(panelCard(t('nav.insights/active'), lineChart(labels, [
      { label: t('c.dau'), values: a.days.map(function (x) { return x.dau; }) },
      { label: t('c.new'), values: a.days.map(function (x) { return x.newUsers; }) },
      { label: t('c.returning'), values: a.days.map(function (x) { return x.returning; }) }], { fmtLabel: fmtDay, name: 'active' }), null, 'active-chart'));
    var rows = a.days.slice().reverse();
    body.appendChild(panelCard(t('c.day'), dataTable([t('c.day'), t('c.dau'), t('c.new'), t('c.returning'), t('c.events')], rows.map(function (x) { return [fmtDay(x.day), fmtNum(x.dau), fmtNum(x.newUsers), fmtNum(x.returning), fmtNum(x.events)]; }), 'active'),
      csvButton('active-users', function () { return { columns: ['day', 'dau', 'new', 'returning', 'events'], rows: rows.map(function (x) { return [x.day, x.dau, x.newUsers, x.returning, x.events]; }) }; }), 'active-table'));
    body.appendChild(storeNote(d.store));
  }, function (err) { failInto(body, err); });
}

function heat(rate) {
  if (rate == null) return '';
  var a = Math.min(1, Math.max(0.06, rate));
  return 'background:color-mix(in srgb, var(--ins-s1) ' + Math.round(a * 85) + '%, transparent)';
}

function retentionPage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/retention');
  var days = input(q.days || '', { ltr: true, name: 'days', placeholder: '1,7,30' });
  var mode = select(q.mode || 'classic', [{ value: 'classic', label: t('r.classic') }, { value: 'rolling', label: t('r.rolling') }], { name: 'mode', label: t('r.mode') });
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(filterBar(q));
  frame.content.appendChild(card(h('div', { class: 'ia-filters p-3 flex flex-wrap items-end gap-3' }, field(t('r.days'), days), field(t('r.mode'), mode),
    button({ label: t('f.apply'), variant: 'primary', data: { 'ins-retention-apply': '' }, onclick: function () { setQuery(Object.assign({}, q, { days: days.value.replace(/[۰-۹]/g, function (c) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(c)); }).replace(/\s/g, ''), mode: mode.value })); } }))));
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/retention' + filterQS(q, { days: q.days, mode: q.mode })).then(function (d) {
    clear(body);
    body.appendChild(kpis(d.summary.map(function (x) { return { key: 'd' + x.day, label: t('k.d', { n: x.day }), value: fmtPct(x.rate), sub: fmtNum(x.returned) + ' / ' + fmtNum(x.cohort) }; })));
    var cols = [t('c.cohort'), t('c.size')].concat(d.days.map(function (n) { return t('k.d', { n: n }); }));
    var rows = d.cohorts.map(function (c) {
      return [fmtDay(c.day), fmtNum(c.size)].concat(c.cells.map(function (cell) {
        return cell.returned == null ? h('span', { class: 'text-secondary-400', text: t('r.pending') }) : h('span', { class: 'ia-heat', style: heat(cell.rate), title: fmtNum(cell.returned), text: fmtPct(cell.rate) });
      }));
    });
    body.appendChild(panelCard(t('nav.insights/retention'), dataTable(cols, rows, 'cohorts'),
      csvButton('retention', function () { return { columns: ['first_day', 'size'].concat(d.days.map(function (n) { return 'd' + n; })), rows: d.cohorts.map(function (c) { return [c.day, c.size].concat(c.cells.map(function (x) { return x.rate == null ? '' : x.rate; })); }) }; }), 'cohorts'));
    body.appendChild(storeNote(d.store));
  }, function (err) { failInto(body, err); });
}

function parseTagText(text) {
  var out = {};
  String(text || '').split(/[;\n]/).forEach(function (part) {
    var kv = part.split('=');
    if (kv.length < 2) return;
    var g = kv[0].trim();
    var ids = kv.slice(1).join('=').split(',').map(function (x) { return x.trim(); }).filter(Boolean);
    if (g && ids.length) out[g] = ids;
  });
  return out;
}

function tagText(tags) {
  return Object.keys(tags || {}).map(function (g) { return g + '=' + tags[g].join(','); }).join(' ; ');
}

function namesList(el) {
  var dl = h('datalist', { id: 'ins-names-' + Math.random().toString(36).slice(2) });
  el.setAttribute('list', dl.id);
  api('GET', '/insights/values' + filterQS(insQuery(), { kind: 'name', limit: 200 })).then(function (d) { (d.items || []).forEach(function (v) { dl.appendChild(h('option', { value: v.value })); }); }, function () {});
  return dl;
}

function explorePage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/explore');
  var names = input(q.name || '', { ltr: true, name: 'name', placeholder: 'DistanceOverall' });
  var metric = select(q.metric || 'sum', ['count', 'sum', 'avg', 'min', 'max', 'users'].map(function (m) { return { value: m, label: t('m.' + m) }; }), { name: 'metric', label: t('ex.metric') });
  var bucket = select(q.bucket || 'none', ['none', 'hour', 'day', 'week'].map(function (b) { return { value: b, label: t('b.' + b) }; }), { name: 'bucket', label: t('ex.bucket') });
  var groupsWanted = (q.groupBy || '').split(',').filter(Boolean);
  var builtins = ['name', 'user', 'platform', 'appVersion', 'source', 'game'];
  var groupSelects = [0, 1, 2, 3].map(function (i) {
    var el = select(groupsWanted[i] || '', [{ value: '', label: t('ex.noGroup') }].concat(builtins.map(function (b) { return { value: b, label: t('g.' + b) }; })), { name: 'groupBy' + i, label: t('ex.group') });
    if (groupsWanted[i] && groupsWanted[i].indexOf('tag:') === 0) el.appendChild(h('option', { value: groupsWanted[i], text: t('g.tag', { g: groupsWanted[i].slice(4) }), selected: true }));
    return el;
  });
  api('GET', '/insights/values' + filterQS(q, { kind: 'tags', limit: 100 })).then(function (d) {
    (d.items || []).forEach(function (v) {
      groupSelects.forEach(function (el, i) {
        if (groupsWanted[i] === 'tag:' + v.value) return;
        el.appendChild(h('option', { value: 'tag:' + v.value, text: t('g.tag', { g: v.value }) }));
      });
    });
  }, function () {});
  var tagF = {};
  Object.keys(q).forEach(function (k) { if (k.indexOf('tag.') === 0) tagF[k.slice(4)] = q[k].split(','); });
  var tags = input(tagText(tagF), { ltr: true, name: 'tags', placeholder: t('ex.filtersHint') });
  var limit = numInput(q.limit || 100, 'limit', t('ex.limit'));
  var body = h('div', { class: 'flex flex-col gap-4' });
  var run = function () {
    var next = { from: q.from, to: q.to, game: q.game, platform: q.platform, appVersion: q.appVersion, includeFlagged: q.includeFlagged, name: names.value.replace(/\s/g, ''), metric: metric.value, bucket: bucket.value,
      groupBy: groupSelects.map(function (el) { return el.value; }).filter(Boolean).filter(function (v, i, a) { return a.indexOf(v) === i; }).join(','), limit: numberValue(limit) || 100 };
    var tf = parseTagText(tags.value);
    Object.keys(tf).forEach(function (g) { next['tag.' + g] = tf[g].join(','); });
    setQuery(next);
  };
  frame.content.appendChild(filterBar(q, { flagged: true }));
  frame.content.appendChild(card(h('div', { class: 'p-3 flex flex-col gap-3', 'data-ins-explore-form': '' },
    h('div', { class: 'ia-filters flex flex-wrap items-end gap-3' }, field(t('ex.names'), names, t('ex.namesHint')), namesList(names), field(t('ex.metric'), metric), field(t('ex.bucket'), bucket), field(t('ex.limit'), limit)),
    h('div', { class: 'ia-filters flex flex-wrap items-end gap-3' }, groupSelects.map(function (el) { return field(t('ex.group'), el); })),
    h('div', { class: 'ia-filters flex flex-wrap items-end gap-3' }, field(t('ex.filters'), tags), button({ label: t('ex.run'), variant: 'primary', data: { 'ins-run': '' }, onclick: run })))));
  frame.content.appendChild(body);
  var params = { name: q.name, metric: q.metric || 'sum', bucket: q.bucket || 'none', groupBy: q.groupBy, limit: q.limit };
  Object.keys(tagF).forEach(function (g) { params['tag.' + g] = tagF[g].join(','); });
  loadingInto(body);
  api('GET', '/insights/explore' + filterQS(q, params)).then(function (d) {
    clear(body);
    var gl = d.groupBy.map(function (g) { return g.indexOf('tag:') === 0 ? t('g.tag', { g: g.slice(4) }) : t('g.' + g); });
    var keyOf = function (r) { return r.keys.length ? r.keys.map(function (k) { return k === '' ? t('none') : k; }).join(' · ') : t('m.' + d.metric); };
    var chart;
    if (d.bucket !== 'none') {
      var labels = [];
      var seriesMap = {};
      var totals = {};
      d.rows.forEach(function (r) {
        if (labels.indexOf(r.bucket) === -1) labels.push(r.bucket);
        var k = keyOf(r);
        totals[k] = (totals[k] || 0) + r.value;
      });
      labels.sort();
      var top = Object.keys(totals).sort(function (a, b) { return totals[b] - totals[a]; }).slice(0, INS_SERIES);
      top.forEach(function (k) { seriesMap[k] = labels.map(function () { return 0; }); });
      d.rows.forEach(function (r) { var k = keyOf(r); if (seriesMap[k]) seriesMap[k][labels.indexOf(r.bucket)] = r.value; });
      chart = lineChart(labels, top.map(function (k) { return { label: k, values: seriesMap[k] }; }), { fmtLabel: d.bucket === 'hour' ? function (x) { return digits(x.slice(5).replace('T', ' ')); } : fmtDay, name: 'explore' });
    } else {
      chart = barList(d.rows.slice(0, 25).map(function (r) { return { label: keyOf(r), value: r.value, tip: [t('c.count') + ': ' + fmtNum(r.count), t('c.users') + ': ' + fmtNum(r.users)] }; }), { name: 'explore' });
    }
    var cols = (d.bucket !== 'none' ? [t('c.bucket')] : []).concat(gl, [t('m.' + d.metric), t('c.count'), t('c.users')]);
    var rows = d.rows.map(function (r) {
      return (d.bucket !== 'none' ? [digits(r.bucket)] : []).concat(r.keys.map(function (k) { return h('span', { dir: 'auto', text: k || t('none') }); }), [fmtNum(r.value), fmtNum(r.count), fmtNum(r.users)]);
    });
    body.appendChild(panelCard(t('nav.insights/explore'), d.rows.length ? chart : h('p', { class: 'text-sm text-secondary-500', text: t('empty') }), null, 'explore-chart'));
    body.appendChild(panelCard(t('c.value'), dataTable(cols, rows, 'explore'),
      csvButton('explore', function () { return { columns: (d.bucket !== 'none' ? ['bucket'] : []).concat(d.groupBy, [d.metric, 'count', 'users']), rows: d.rows.map(function (r) { return (d.bucket !== 'none' ? [r.bucket] : []).concat(r.keys, [r.value, r.count, r.users]); }) }; }), 'explore-table'));
    body.appendChild(storeNote(d.store));
  }, function (err) { failInto(body, err); });
}

function funnelPanel(fn, after) {
  fn = fn || { name: '', game: '', windowSeconds: 86400, steps: [{ name: '' }, { name: '' }] };
  var name = input(fn.name, { name: 'name' });
  var game = input(fn.game, { ltr: true, name: 'game' });
  var win = numInput(Math.round(fn.windowSeconds / 60), 'window', t('fn.window'));
  var rows = h('div', { class: 'ia-rows', 'data-ins-steps': '' });
  var addRow = function (st) {
    var ev = input(st.name || '', { ltr: true, name: 'step-name' });
    var tg = input(tagText(st.tags), { ltr: true, name: 'step-tags', placeholder: t('fn.hintTags') });
    var row = h('div', { class: 'ia-row', 'data-ins-step': '' }, field(t('fn.event'), ev), namesList(ev), field(t('fn.tags'), tg),
      button({ label: t('delete'), variant: 'danger', onclick: function () { if (rows.children.length > 2) row.remove(); } }));
    row.read = function () { var o = { name: ev.value.trim() }; var tags = parseTagText(tg.value); if (Object.keys(tags).length) o.tags = tags; return o; };
    rows.appendChild(row);
  };
  fn.steps.forEach(addRow);
  var msg = h('div', {});
  var save = button({ label: t('save'), variant: 'primary', data: { 'ins-funnel-save': '' }, onclick: function () {
    busy(save, true);
    var body = { name: name.value, game: game.value.trim(), windowSeconds: (numberValue(win) || 0) * 60, steps: Array.prototype.map.call(rows.children, function (r) { return r.read(); }) };
    api(fn.id ? 'PUT' : 'POST', '/insights/funnels' + (fn.id ? '/' + fn.id : ''), body).then(function (out) { closePanel(); toast('success', t('saved')); after(out); }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  var foot = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (fn.id) foot.unshift(button({ label: t('delete'), variant: 'danger', data: { 'ins-funnel-delete': '' }, onclick: function () {
    confirmBox(fn.name).then(function (ok) { if (ok) api('DELETE', '/insights/funnels/' + fn.id).then(function () { closePanel(); toast('success', t('deleted')); after(null); }, function (err) { fieldError(msg, err); }); });
  } }));
  sidePanel(fn.id ? fn.name : t('fn.new'), h('div', {}, field(t('fn.name'), name), field(t('cat.game'), game), field(t('fn.window'), win), h('h3', { class: 'font-semibold mb-2', text: t('fn.steps') }), rows,
    h('div', { class: 'mt-2' }, button({ label: t('fn.addStep'), data: { 'ins-add-step': '' }, onclick: function () { if (rows.children.length < 12) addRow({ name: '' }); } })), msg), foot, 'funnel');
}

function funnelsPage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/funnels', [button({ label: t('fn.new'), variant: 'primary', data: { 'ins-funnel-new': '' }, onclick: function () { funnelPanel(null, function (f) { setQuery(Object.assign({}, q, { funnel: f ? f.id : '' })); }); } })]);
  var layout = h('div', { class: 'grid gap-4 lg:grid-cols-[16rem_1fr]' });
  var list = h('div', { class: 'flex flex-col gap-1', 'data-ins-funnels': '' });
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(filterBar(q));
  layout.appendChild(card(h('div', { class: 'p-3' }, list)));
  layout.appendChild(body);
  frame.content.appendChild(layout);
  loadingInto(list);
  api('GET', '/insights/funnels').then(function (d) {
    clear(list);
    if (!d.items.length) list.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('fn.empty') }));
    d.items.forEach(function (f) {
      list.appendChild(h('a', { href: '#/insights/funnels?' + filterQS(q, { funnel: f.id }).slice(1), class: 'ia-navitem rounded-md px-3 py-2 text-sm' + (q.funnel === f.id ? ' ia-navitem--on' : ''), 'data-ins-funnel': f.id, dir: 'auto', text: f.name }));
    });
    var fn = d.items.filter(function (f) { return f.id === q.funnel; })[0];
    if (!fn) { clear(body).appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('fn.pick') }))); return; }
    loadingInto(body);
    api('GET', '/insights/funnels/' + fn.id + '/run' + filterQS(q)).then(function (r) {
      clear(body);
      var first = r.steps.length ? r.steps[0].users : 0;
      var bars = h('div', { class: 'flex flex-col gap-3', 'data-ins-chart': 'funnel' }, r.steps.map(function (st, i) {
        var pct = first ? st.users / first * 100 : 0;
        return h('div', { class: 'ia-step', 'data-ins-step-result': i },
          h('div', { class: 'flex justify-between gap-2 text-sm mb-1' }, h('span', { class: 'font-medium', dir: 'auto', text: t('fn.step', { n: i + 1 }) + ' · ' + st.name }), h('span', { text: fmtNum(st.users) + ' · ' + fmtPct(st.fromFirst) })),
          h('div', { class: 'ia-bartrack ia-bartrack--tall' }, h('div', { class: 'ia-bar', style: 'width:' + Math.max(pct, st.users ? 1 : 0) + '%;background:' + seriesColor(0) })),
          i ? h('div', { class: 'text-xs text-secondary-500 mt-1', text: t('fn.fromPrev') + ': ' + fmtPct(st.fromPrevious) + ' · ' + t('fn.drop') + ': ' + fmtNum(st.dropOff) + ' · ' + t('fn.median') + ': ' + fmtSeconds(st.medianSeconds) }) : null);
      }));
      body.appendChild(panelCard(fn.name, h('div', { class: 'flex flex-col gap-3' }, h('p', { class: 'text-xs text-secondary-500', text: t('fn.windowV', { n: Math.round(r.windowSeconds / 60) }) }), bars),
        [button({ label: t('edit'), data: { 'ins-funnel-edit': '' }, onclick: function () { funnelPanel(fn, function (f) { setQuery(Object.assign({}, q, { funnel: f ? f.id : '' })); funnelsPage(root); }); } }),
          csvButton('funnel', function () { return { columns: ['step', 'event', 'users', 'from_first', 'from_previous', 'drop_off', 'median_seconds'], rows: r.steps.map(function (st, i) { return [i + 1, st.name, st.users, st.fromFirst, st.fromPrevious == null ? '' : st.fromPrevious, st.dropOff, st.medianSeconds == null ? '' : st.medianSeconds]; }) }; })], 'funnel-result'));
      body.appendChild(storeNote(r.store));
    }, function (err) { failInto(body, err); });
  }, function (err) { failInto(list, err); });
}

function timelinePanel(sess) {
  var body = h('div', { class: 'flex flex-col gap-3' });
  sidePanel(t('s.timeline'), body, [button({ label: t('close'), onclick: closePanel })], 'timeline');
  loadingInto(body);
  api('GET', '/insights/sessions/' + sess.id + '/timeline').then(function (d) {
    clear(body);
    var se = d.session;
    body.appendChild(h('dl', { class: 'ia-dl text-sm' },
      h('dt', { text: t('s.player') }), h('dd', {}, playerLink(se.userId, sess.username || sess.displayName)),
      h('dt', { text: t('s.started') }), h('dd', { text: fmtDate(se.startedAt) }),
      h('dt', { text: t('s.length') }), h('dd', { text: fmtSeconds(se.seconds) }),
      h('dt', { text: t('f.platform') }), h('dd', { dir: 'ltr', text: [se.platform, se.appVersion].filter(Boolean).join(' · ') || t('none') })));
    if (d.matchedBy === 'time' && d.events.length) body.appendChild(notice('warning', t('s.byTime')));
    if (!d.events.length) { body.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('s.noEvents') })); return; }
    body.appendChild(h('ol', { class: 'ia-timeline', 'data-ins-timeline': '' }, d.events.map(function (e) {
      return h('li', { class: 'ia-tl-item', 'data-ins-event': e.name },
        h('span', { class: 'ia-tl-time text-xs text-secondary-500', text: fmtTime(e.at) }),
        h('div', { class: 'flex flex-col gap-1 min-w-0' },
          h('div', { class: 'flex flex-wrap items-center gap-2' }, h('span', { class: 'font-medium font-mono text-sm', dir: 'ltr', text: e.name }), e.value != null ? tag(fmtNum(e.value), 'primary') : null, tag(e.source, 'secondary'), e.flagged ? tag(e.flagReason || 'flagged', 'warning') : null),
          Object.keys(e.tags || {}).length ? h('div', { class: 'flex flex-wrap gap-1' }, Object.keys(e.tags).map(function (g) { return h('span', { class: 'ia-chip', dir: 'ltr', text: g + ': ' + e.tags[g] }); })) : null,
          e.fields ? h('div', { class: 'text-xs text-secondary-500 font-mono', dir: 'ltr', text: JSON.stringify(e.fields) }) : null));
    })));
  }, function (err) { failInto(body, err); });
}

function sessionsPage(root) {
  var q = insQuery();
  var frame = dashFrame(root, 'insights/sessions');
  var user = input(q.userId || '', { ltr: true, name: 'userId', placeholder: t('s.userId') });
  var status = select(q.status || '', [{ value: '', label: t('all') }, { value: 'open', label: t('s.open') }, { value: 'ended', label: t('s.ended') }], { name: 'status', label: t('s.status') });
  var platform = input(q.platform || '', { ltr: true, name: 'platform', placeholder: t('f.any') });
  var body = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(card(h('div', { class: 'ia-filters p-3 flex flex-wrap items-end gap-3' }, field(t('s.userId'), user), field(t('s.status'), status), field(t('f.platform'), platform),
    button({ label: t('f.apply'), variant: 'primary', data: { 'ins-apply': '' }, onclick: function () { setQuery({ userId: user.value.trim(), status: status.value, platform: platform.value.trim(), page: '' }); } }))));
  frame.content.appendChild(body);
  var page = Number(q.page) || 1;
  loadingInto(body);
  api('GET', '/insights/sessions' + qs({ userId: q.userId, status: q.status, platform: q.platform, page: page, limit: 25 })).then(function (d) {
    clear(body);
    var rows = d.items.map(function (x) {
      return { id: x.id, item: x, cells: [playerLink(x.userId, x.username || x.displayName), h('span', { dir: 'ltr', text: [x.platform, x.appVersion].filter(Boolean).join(' · ') || t('none') }), fmtDate(x.startedAt), fmtSeconds(x.seconds),
        tag(x.endedAt ? t('s.ended') : t('s.open'), x.endedAt ? 'secondary' : 'success')] };
    });
    if (!rows.length) body.appendChild(h('p', { class: 'text-sm text-secondary-500', 'data-ins-empty': 'sessions', text: t('empty') }));
    else {
      var tb = table([{ label: t('s.player') }, { label: t('f.platform') }, { label: t('s.started') }, { label: t('s.length') }, { label: t('s.status') }], rows, timelinePanel);
      tb.setAttribute('data-ins-table', 'sessions');
      body.appendChild(tb);
    }
    body.appendChild(h('div', { class: 'flex justify-between items-center gap-2' },
      csvButton('sessions', function () { return { columns: ['session', 'user', 'platform', 'version', 'started', 'seconds', 'ended'], rows: d.items.map(function (x) { return [x.sessionId, x.userId, x.platform, x.appVersion, x.startedAt, x.seconds == null ? '' : x.seconds, x.endedAt || '']; }) }; }),
      pager({ page: d.page, limit: d.limit, total: d.total }, function (p) { setQuery(Object.assign({}, q, { page: p })); })));
  }, function (err) { failInto(body, err); });
}
