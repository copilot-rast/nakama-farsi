function listText(v) { return (v || []).join(', '); }

function listOf(text) { return String(text || '').split(',').map(function (x) { return x.trim(); }).filter(Boolean); }

function gameQS(game) { return '?game=' + encodeURIComponent(game || ''); }

function groupPanel(game, g, after) {
  g = g || { key: '', nameFa: '', nameEn: '', description: '', closed: false, enabled: true, values: [] };
  var key = input(g.key, { ltr: true, name: 'key' });
  if (g.key) key.disabled = true;
  var nameFa = input(g.nameFa, { name: 'nameFa' });
  var nameEn = input(g.nameEn, { ltr: true, name: 'nameEn' });
  var desc = input(g.description, { name: 'description', multiline: true, rows: 2 });
  var closed = check(g.closed, t('cat.closed'), 'closed');
  var enabled = check(g.enabled, t('cat.enabled'), 'enabled');
  var values = input(g.values.map(function (v) { return [v.id, v.labelFa, v.labelEn].join(','); }).join('\n'), { ltr: true, name: 'values', multiline: true, rows: 8 });
  var msg = h('div', {});
  var save = button({ label: t('save'), variant: 'primary', data: { 'ins-group-save': '' }, onclick: function () {
    busy(save, true);
    var vals = values.value.split('\n').map(function (l) { return l.trim(); }).filter(Boolean).map(function (l) { var p = l.split(','); return { id: p[0].trim(), labelFa: (p[1] || '').trim(), labelEn: (p[2] || '').trim() }; });
    api('PUT', '/insights/events/catalog/groups/' + encodeURIComponent(key.value.trim()) + gameQS(game), { nameFa: nameFa.value, nameEn: nameEn.value, description: desc.value, closed: closed.input.checked, enabled: enabled.input.checked, values: vals })
      .then(function () { closePanel(); toast('success', t('saved')); after(); }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  var foot = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (g.key) foot.unshift(button({ label: t('delete'), variant: 'danger', data: { 'ins-group-delete': '' }, onclick: function () {
    confirmBox(g.key).then(function (ok) { if (ok) api('DELETE', '/insights/events/catalog/groups/' + encodeURIComponent(g.key) + gameQS(game)).then(function () { closePanel(); toast('success', t('deleted')); after(); }, function (err) { fieldError(msg, err); }); });
  } }));
  sidePanel(g.key || t('cat.newGroup'), h('div', {}, field(t('cat.key'), key, t('cat.keyHint')), h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('cat.nameFa'), nameFa), field(t('cat.nameEn'), nameEn)),
    field(t('cat.desc'), desc), closed, enabled, field(t('cat.values'), values, t('cat.valuesHint')), msg), foot, 'group');
}

function eventPanel(game, e, after) {
  e = e || { name: '', labelFa: '', labelEn: '', description: '', enabled: true, pathGroups: [], requiredTags: [], allowedTags: [], value: {} };
  var v = e.value || {};
  var name = input(e.name, { ltr: true, name: 'name' });
  if (e.name) name.disabled = true;
  var labelFa = input(e.labelFa, { name: 'labelFa' });
  var labelEn = input(e.labelEn, { ltr: true, name: 'labelEn' });
  var desc = input(e.description, { name: 'description', multiline: true, rows: 2 });
  var enabled = check(e.enabled, t('cat.enabled'), 'enabled');
  var path = input(listText(e.pathGroups), { ltr: true, name: 'pathGroups' });
  var req = input(listText(e.requiredTags), { ltr: true, name: 'requiredTags' });
  var allowed = input(listText(e.allowedTags), { ltr: true, name: 'allowedTags' });
  var vreq = check(v.required, t('cat.valueRequired'), 'valueRequired');
  var vmin = input(v.min == null ? '' : digits(String(v.min)), { ltr: true, name: 'min' });
  var vmax = input(v.max == null ? '' : digits(String(v.max)), { ltr: true, name: 'max' });
  var unit = input(v.unit || '', { name: 'unit' });
  var msg = h('div', {});
  var save = button({ label: t('save'), variant: 'primary', data: { 'ins-event-save': '' }, onclick: function () {
    busy(save, true);
    var mn = floatValue(vmin), mx = floatValue(vmax);
    api('PUT', '/insights/events/catalog/events/' + encodeURIComponent(name.value.trim()) + gameQS(game), { labelFa: labelFa.value, labelEn: labelEn.value, description: desc.value, enabled: enabled.input.checked,
      pathGroups: listOf(path.value), requiredTags: listOf(req.value), allowedTags: listOf(allowed.value), value: { required: vreq.input.checked, min: mn, max: mx, unit: unit.value } })
      .then(function () { closePanel(); toast('success', t('saved')); after(); }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  var foot = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (e.name) foot.unshift(button({ label: t('delete'), variant: 'danger', data: { 'ins-event-delete': '' }, onclick: function () {
    confirmBox(e.name).then(function (ok) { if (ok) api('DELETE', '/insights/events/catalog/events/' + encodeURIComponent(e.name) + gameQS(game)).then(function () { closePanel(); toast('success', t('deleted')); after(); }, function (err) { fieldError(msg, err); }); });
  } }));
  sidePanel(e.name || t('cat.newEvent'), h('div', {}, field(t('cat.name'), name), h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('cat.nameFa'), labelFa), field(t('cat.nameEn'), labelEn)), field(t('cat.desc'), desc), enabled,
    field(t('cat.path'), path, t('cat.pathHint')), field(t('cat.required'), req), field(t('cat.allowed'), allowed), vreq,
    h('div', { class: 'grid sm:grid-cols-3 gap-x-4' }, field(t('cat.min'), vmin), field(t('cat.max'), vmax), field(t('cat.unit'), unit)), msg), foot, 'event');
}

function catalogPage(root) {
  var q = hashQuery();
  var game = q.game || '';
  var frame = dashFrame(root, 'insights/catalog');
  var gameIn = input(game, { ltr: true, name: 'game', placeholder: t('f.mainGame') });
  var body = h('div', { class: 'flex flex-col gap-4' });
  var reload = function () { catalogPage(root); };
  frame.content.appendChild(card(h('div', { class: 'ia-filters p-3 flex flex-wrap items-end gap-3' }, field(t('cat.game'), gameIn),
    button({ label: t('f.apply'), variant: 'primary', data: { 'ins-apply': '' }, onclick: function () { setQuery({ game: gameIn.value.trim() }); } }))));
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/events/catalog' + gameQS(game)).then(function (d) {
    clear(body);
    var gl = d.groups.map(function (g) { return { id: g.key, item: g, cells: [mono(g.key), h('span', { dir: 'auto', text: (insLang() === 'fa' ? g.nameFa || g.nameEn : g.nameEn || g.nameFa) || t('none') }), tag(g.closed ? t('cat.closedTag') : t('cat.open'), g.closed ? 'warning' : 'secondary'), t('cat.valuesN', { n: g.values.length }), tag(g.enabled ? t('cat.enabled') : t('no'), g.enabled ? 'success' : 'secondary')] }; });
    var gt = gl.length ? table([{ label: t('cat.key') }, { label: t('cat.label') }, { label: t('cat.closed') }, { label: t('cat.values') }, { label: t('cat.enabled') }], gl, function (g) { groupPanel(game, g, reload); }) : h('p', { class: 'text-sm text-secondary-500', text: t('cat.emptyGroups') });
    if (gl.length) gt.setAttribute('data-ins-table', 'groups');
    body.appendChild(panelCard(t('cat.groups'), gt, [button({ label: t('cat.newGroup'), variant: 'primary', data: { 'ins-group-new': '' }, onclick: function () { groupPanel(game, null, reload); } })], 'groups'));
    var range = function (v) { v = v || {}; return v.min == null && v.max == null ? t('none') : digits((v.min == null ? '…' : v.min) + ' – ' + (v.max == null ? '…' : v.max)) + (v.unit ? ' ' + v.unit : ''); };
    var el = d.events.map(function (e) { return { id: e.name, item: e, cells: [mono(e.name), h('span', { dir: 'auto', text: (insLang() === 'fa' ? e.labelFa || e.labelEn : e.labelEn || e.labelFa) || t('none') }), mono(listText(e.pathGroups) || t('none')), mono(listText(e.requiredTags) || t('none')), range(e.value), tag(e.enabled ? t('cat.enabled') : t('no'), e.enabled ? 'success' : 'secondary')] }; });
    var et = el.length ? table([{ label: t('cat.name') }, { label: t('cat.label') }, { label: t('cat.path') }, { label: t('cat.required') }, { label: t('c.value') }, { label: t('cat.enabled') }], el, function (e) { eventPanel(game, e, reload); }) : h('p', { class: 'text-sm text-secondary-500', text: t('cat.emptyEvents') });
    if (el.length) et.setAttribute('data-ins-table', 'events');
    body.appendChild(panelCard(t('cat.events'), et, [button({ label: t('cat.newEvent'), variant: 'primary', data: { 'ins-event-new': '' }, onclick: function () { eventPanel(game, null, reload); } })], 'events'));
  }, function (err) { failInto(body, err); });
}

var INS_DRIVER_FIELDS = {
  postgres: [['host', 'st.host'], ['port', 'st.port', 'num'], ['database', 'st.database'], ['user', 'st.user'], ['sslMode', 'st.sslMode', ['require', 'verify-full', 'verify-ca', 'prefer', 'disable']], ['schema', 'st.schema']],
  clickhouse: [['url', 'st.url'], ['database', 'st.database'], ['table', 'st.table'], ['user', 'st.user']],
  mongodb: [['hosts', 'st.hosts'], ['database', 'st.database'], ['collection', 'st.collection'], ['user', 'st.user'], ['authSource', 'st.authSource'], ['tls', 'st.tls', 'bool']]
};

function storagePage(root) {
  var frame = dashFrame(root, 'insights/storage');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(body);
  loadingInto(body);
  Promise.all([api('GET', '/insights/storage'), api('GET', '/insights/outbox')]).then(function (res) {
    var st = res[0], ob = res[1];
    clear(body);
    var driver = select(st.driver, st.drivers.map(function (d) { return { value: d, label: t('st.' + d) }; }), { name: 'driver', label: t('st.driver') });
    var formBox = h('div', {});
    var controls = {};
    var secret = h('input', { type: 'password', class: INS_INPUT, name: 'secret', dir: 'ltr', autocomplete: 'new-password', placeholder: st.hasSecret ? t('st.secretKeep') : '' });
    var paint = function () {
      clear(formBox);
      controls = {};
      var fields = INS_DRIVER_FIELDS[driver.value];
      if (!fields) return;
      var cfg = (st.config || {})[driver.value] || {};
      var grid = h('div', { class: 'grid sm:grid-cols-2 gap-x-4' });
      fields.forEach(function (f) {
        var el;
        if (f[2] === 'bool') { el = check(cfg[f[0]], t(f[1]), f[0]); controls[f[0]] = { read: function () { return el.input.checked; } }; grid.appendChild(el); return; }
        if (Array.isArray(f[2])) el = select(cfg[f[0]] || f[2][0], f[2], { name: f[0], label: t(f[1]) });
        else el = input(cfg[f[0]] == null ? '' : String(cfg[f[0]]), { ltr: true, name: f[0] });
        controls[f[0]] = { read: function () { return f[2] === 'num' ? numberValue(el) || 0 : el.value.trim(); } };
        grid.appendChild(field(t(f[1]), el));
      });
      formBox.appendChild(grid);
      formBox.appendChild(field(t('st.secret'), secret, st.hasSecret ? t('st.secretSet') + ' · ' + t('st.secretKeep') : null));
      if (!st.secretKey) formBox.appendChild(notice('warning', t('st.noKey')));
    };
    driver.addEventListener('change', paint);
    paint();
    var body2 = function () {
      var cfg = {};
      Object.keys(controls).forEach(function (k) { cfg[k] = controls[k].read(); });
      var out = { driver: driver.value, config: {} };
      if (INS_DRIVER_FIELDS[driver.value]) out.config[driver.value] = cfg;
      if (secret.value) out.secret = secret.value;
      return out;
    };
    var msg = h('div', { 'data-ins-msg': '' });
    var test = button({ label: t('st.test'), data: { 'ins-storage-test': '' }, onclick: function () {
      busy(test, true);
      clear(msg);
      api('POST', '/insights/storage/test', body2()).then(function (r) { busy(test, false); msg.appendChild(notice(r.ok ? 'success' : 'error', r.ok ? t('st.testOk', { ms: r.millis }) : t('st.testFail', { e: r.error }))); }, function (err) { busy(test, false); fieldError(msg, err); });
    } });
    var save = button({ label: t('save'), variant: 'primary', data: { 'ins-storage-save': '' }, onclick: function () {
      busy(save, true);
      clear(msg);
      api('PUT', '/insights/storage', body2()).then(function () { toast('success', t('saved')); storagePage(root); }, function (err) { busy(save, false); fieldError(msg, err); });
    } });
    var status = st.status === 'error' ? notice('error', t('st.error', { e: st.error })) : null;
    body.appendChild(panelCard(t('st.driver'), h('div', {}, field(t('st.driver'), driver), formBox, h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('st.switchNote') }), status, msg,
      h('div', { class: 'flex gap-2 justify-end' }, test, save), st.updatedBy ? h('p', { class: 'text-xs text-secondary-500', text: t('updated', { by: st.updatedBy, at: fmtDate(st.updatedAt) }) }) : null), null, 'storage'));
    var tiles = kpis([{ key: 'pending', label: t('ob.pending'), value: fmtNum(ob.pending), sub: t('set.maxPendingEvents') + ': ' + fmtNum(ob.maxPendingEvents) }, { key: 'rows', label: t('ob.rows'), value: fmtNum(ob.rows) },
      { key: 'dead', label: t('ob.dead'), value: fmtNum(ob.dead), sub: fmtNum(ob.deadEvents) + ' ' + t('c.events') }, { key: 'oldest', label: t('ob.oldest'), value: ob.oldest ? fmtDate(ob.oldest) : t('none') }]);
    var failing = ob.failing.length ? dataTable(['#', t('s.source'), t('c.events'), t('ob.attempts'), t('ob.error'), t('ob.next')], ob.failing.map(function (f) {
      return [digits(String(f.id)), f.source, fmtNum(f.count), fmtNum(f.attempts), h('span', { class: 'text-xs font-mono', dir: 'ltr', text: f.lastError }), f.dead ? tag(t('ob.dead'), 'error') : fmtDate(f.nextAt)];
    }), 'outbox') : notice('success', t('ob.ok'));
    body.appendChild(panelCard(t('ob.title'), h('div', { class: 'flex flex-col gap-3' }, tiles, failing), [
      button({ label: t('ob.retry'), data: { 'ins-outbox-retry': '' }, onclick: function () { api('POST', '/insights/outbox/retry').then(function () { toast('success', t('saved')); storagePage(root); }, function (err) { toast('error', insErrText(err)); }); } }),
      button({ label: t('ob.discard'), variant: 'danger', disabled: !ob.dead, data: { 'ins-outbox-discard': '' }, onclick: function () { confirmBox(t('ob.discardQ')).then(function (ok) { if (ok) api('POST', '/insights/outbox/discard').then(function () { storagePage(root); }); }); } })], 'outbox'));
  }, function (err) { failInto(body, err); });
}

var INS_CONNECTOR_TEMPLATES = {
  ga4_bigquery: { projectId: '', dataset: 'analytics_000000000', startDate: '', daysPerRun: 3, lagDays: 1, paramTags: [], valueParam: '', nameMap: {}, includeEvents: [] },
  ga4_data: { propertyId: '', startDate: '', daysPerRun: 3, lagDays: 1, dimensions: ['platform'], nameMap: {}, includeEvents: [] },
  gameanalytics: { endpoint: 'https://s3.amazonaws.com', bucket: '', prefix: '', region: 'us-east-1', accessKey: '', filesPerRun: 20, skipSessions: false }
};

function connectorPanel(c, kinds, after) {
  var isNew = !c;
  c = c || { kind: kinds[0], name: '', enabled: false, everyMinutes: 60, config: INS_CONNECTOR_TEMPLATES[kinds[0]], hasSecret: false };
  var kind = select(c.kind, kinds.map(function (k) { return { value: k, label: t('kind.' + k) }; }), { name: 'kind', label: t('im.kind') });
  if (!isNew) kind.disabled = true;
  var name = input(c.name, { name: 'name' });
  var enabled = check(c.enabled, t('cat.enabled'), 'enabled');
  var every = numInput(c.everyMinutes, 'everyMinutes', t('im.every'));
  var config = input(JSON.stringify(c.config, null, 2), { ltr: true, mono: true, multiline: true, rows: 12, name: 'config' });
  var secret = input('', { ltr: true, mono: true, multiline: true, rows: 4, name: 'secret', placeholder: c.hasSecret ? t('st.secretKeep') : '' });
  var reset = check(false, t('im.reset'), 'resetCursor');
  kind.addEventListener('change', function () { config.value = JSON.stringify(INS_CONNECTOR_TEMPLATES[kind.value], null, 2); });
  var msg = h('div', {});
  var save = button({ label: t('save'), variant: 'primary', data: { 'ins-connector-save': '' }, onclick: function () {
    var cfg;
    try { cfg = JSON.parse(config.value); } catch (e) { fieldError(msg, { code: 'INVALID_JSON', field: 'config' }); return; }
    busy(save, true);
    var body = { kind: kind.value, name: name.value, enabled: enabled.input.checked, everyMinutes: numberValue(every) || 60, config: cfg, resetCursor: reset.input.checked };
    if (secret.value.trim()) body.secret = secret.value.trim();
    api(isNew ? 'POST' : 'PUT', '/insights/connectors' + (isNew ? '' : '/' + c.id), body).then(function () { closePanel(); toast('success', t('saved')); after(); }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  var runs = h('div', {});
  var foot = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (!isNew) {
    foot.unshift(button({ label: t('delete'), variant: 'danger', data: { 'ins-connector-delete': '' }, onclick: function () { confirmBox(c.name).then(function (ok) { if (ok) api('DELETE', '/insights/connectors/' + c.id).then(function () { closePanel(); after(); }); }); } }));
    foot.unshift(button({ label: t('im.run'), data: { 'ins-connector-run': '' }, onclick: function () { api('POST', '/insights/connectors/' + c.id + '/run').then(function () { toast('success', t('im.queued')); }, function (err) { fieldError(msg, err); }); } }));
    api('GET', '/insights/connectors/' + c.id + '/runs').then(function (d) {
      runs.appendChild(h('h3', { class: 'font-semibold my-2', text: t('im.runs') }));
      runs.appendChild(dataTable([t('s.started'), t('im.status'), t('im.read'), t('im.written'), t('im.skipped'), t('ob.error')], d.items.map(function (r) {
        return [fmtDate(r.startedAt), tag(r.status, r.status === 'ok' ? 'success' : r.status === 'error' ? 'error' : 'warning'), fmtNum(r.read), fmtNum(r.written), fmtNum(r.skipped), h('span', { class: 'text-xs font-mono', dir: 'ltr', text: r.error })];
      }), 'runs'));
    });
  }
  var hint = kind.value === 'gameanalytics' ? t('im.secretHint.s3') : t('im.secretHint.ga');
  sidePanel(isNew ? t('im.new') : c.name, h('div', {}, notice('warning', t('im.untested')), h('div', { class: 'h-3' }), field(t('im.kind'), kind), field(t('im.name'), name), enabled, field(t('im.every'), every),
    field(t('im.config'), config), field(t('im.secret'), secret, (c.hasSecret ? t('st.secretSet') + ' · ' : '') + hint), isNew ? null : reset, msg, runs), foot, 'connector');
}

function importsPage(root) {
  var frame = dashFrame(root, 'insights/imports');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/connectors').then(function (d) {
    clear(body);
    var reload = function () { importsPage(root); };
    var rows = d.items.map(function (c) {
      return { id: c.id, item: c, cells: [h('span', { dir: 'auto', text: c.name }), t('kind.' + c.kind), tag(c.enabled ? t('cat.enabled') : t('no'), c.enabled ? 'success' : 'secondary'), fmtNum(c.everyMinutes),
        c.lastRunAt ? fmtDate(c.lastRunAt) : t('none'), c.running ? tag(t('im.running'), 'warning') : c.lastStatus ? tag(c.lastStatus, c.lastStatus === 'ok' ? 'success' : 'error') : t('none'), h('span', { class: 'text-xs font-mono', dir: 'ltr', text: c.lastError })] };
    });
    var tb = rows.length ? table([{ label: t('im.name') }, { label: t('im.kind') }, { label: t('cat.enabled') }, { label: t('im.every') }, { label: t('im.last') }, { label: t('im.status') }, { label: t('ob.error') }], rows, function (c) { connectorPanel(c, d.kinds, reload); })
      : h('p', { class: 'text-sm text-secondary-500', text: t('im.empty') });
    if (rows.length) tb.setAttribute('data-ins-table', 'connectors');
    if (!d.secretKey) body.appendChild(notice('warning', t('st.noKey')));
    body.appendChild(panelCard(t('nav.insights/imports'), tb, [button({ label: t('im.new'), variant: 'primary', data: { 'ins-connector-new': '' }, onclick: function () { connectorPanel(null, d.kinds, reload); } })], 'connectors'));
  }, function (err) { failInto(body, err); });
}

var INS_SETTINGS = {
  events: [
    ['set.events', [['enabled', 'bool'], ['gameEvents', 'bool'], ['sessionEvents', 'bool']]],
    ['set.limits', [['maxBatch'], ['maxTags'], ['maxNameLength'], ['maxFields'], ['maxStringLength'], ['valueMin', 'float'], ['valueMax', 'float'], ['pastHours'], ['futureMinutes'], ['receiptDays']]],
    ['set.rate', [['batchesPerMinute'], ['eventsPerMinute']]],
    ['set.policy', [['unknownEvents', ['accept', 'flag', 'reject']], ['unknownTagGroups', ['accept', 'flag', 'reject']], ['unknownTagValues', ['flag', 'reject']], ['valueOutOfRange', ['flag', 'reject']]]],
    ['set.client', [['flushSeconds'], ['maxQueue']]]
  ],
  storage: [['set.storage', [['keepDays'], ['batchSize'], ['maxPendingEvents'], ['deadAfter']]]]
};

function analyticsSettingsPage(root) {
  var frame = dashFrame(root, 'insights/settings');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/events/settings').then(function (d) {
    clear(body);
    var readers = [];
    ['events', 'storage'].forEach(function (block) {
      INS_SETTINGS[block].forEach(function (sec) {
        var grid = h('div', { class: 'grid sm:grid-cols-2 lg:grid-cols-3 gap-x-4' });
        sec[1].forEach(function (f) {
          var key = f[0], kind = f[1], cur = d[block][key], def = d.defaults[block][key];
          var el;
          if (kind === 'bool') { el = check(cur, t('set.' + key), key); readers.push(function (out) { out[block][key] = el.input.checked; }); grid.appendChild(el); return; }
          if (Array.isArray(kind)) {
            el = select(cur, kind.map(function (p) { return { value: p, label: t('p.' + p) }; }), { name: key, label: t('set.' + key) });
            readers.push(function (out) { out[block][key] = el.value; });
            grid.appendChild(field(t('set.' + key), el, t('set.default', { v: t('p.' + def) })));
            return;
          }
          el = input(digits(String(cur)), { ltr: true, name: key });
          readers.push(function (out) { var v = kind === 'float' ? floatValue(el) : numberValue(el); out[block][key] = v == null ? def : v; });
          grid.appendChild(field(t('set.' + key), el, t('set.default', { v: def })));
        });
        body.appendChild(panelCard(t(sec[0]), grid, null, 'settings-' + sec[0]));
      });
    });
    var msg = h('div', { 'data-ins-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', text: d.updatedBy ? t('updated', { by: d.updatedBy, at: fmtDate(d.updatedAt) }) : '' });
    var save = button({ label: t('save'), variant: 'primary', data: { 'ins-settings-save': '' }, onclick: function () {
      var out = { events: Object.assign({}, d.events), storage: Object.assign({}, d.storage) };
      readers.forEach(function (r) { r(out); });
      busy(save, true);
      clear(msg);
      api('PUT', '/insights/events/settings', out).then(function (x) { busy(save, false); d = x; meta.textContent = t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }); toast('success', t('saved')); }, function (err) { busy(save, false); fieldError(msg, err); });
    } });
    body.appendChild(h('div', { class: 'ia-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), save));
  }, function (err) { failInto(body, err); });
}

function contextFillPage(root) {
  var frame = dashFrame(root, 'insights/context');
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(body);
  loadingInto(body);
  api('GET', '/insights/events/settings').then(function (d) {
    clear(body);
    var c = d.context || {}, def = (d.defaults || {}).context || {};
    var sources = (d.sources || []).map(function (s) { return { value: s.key, label: s.key + ' · ' + ((insLang() === 'fa' ? s.labelFa || s.labelEn : s.labelEn || s.labelFa) || s.key) }; });
    if (!d.providers) body.appendChild(notice('warning', t('ctx.noProviders')));
    var enabled = check(c.enabled, t('ctx.enabled'), 'enabled');
    var conflict = select(c.conflict, ['override', 'keep_client', 'reject'].map(function (p) { return { value: p, label: t('ctx.c.' + p) }; }), { name: 'conflict', label: t('ctx.conflict') });
    var flag = check(c.flagMismatch, t('ctx.flagMismatch'), 'flagMismatch');
    var outside = check(c.keepClientOutside, t('ctx.keepClientOutside'), 'keepClientOutside');
    var grace = input(digits(String(c.graceSeconds == null ? def.graceSeconds : c.graceSeconds)), { ltr: true, name: 'graceSeconds' });
    body.appendChild(panelCard(t('ctx.policy'), h('div', {}, enabled,
      h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('ctx.conflict'), conflict, t('set.default', { v: t('ctx.c.' + def.conflict) })), field(t('ctx.grace'), grace, t('ctx.graceHint', { v: def.graceSeconds }))),
      flag, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('ctx.flagHint') }), outside, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-1', text: t('ctx.outsideHint') })), null, 'context-policy'));
    var rows = h('div', { class: 'flex flex-col gap-2', 'data-ins-fills': '' });
    var readers = [];
    var addRow = function (f) {
      var opts = sources.slice();
      if (f.source && !opts.some(function (o) { return o.value === f.source; })) opts.push({ value: f.source, label: f.source });
      if (!opts.length) opts.push({ value: '', label: t('none') });
      var on = check(f.enabled, t('ctx.fillOn'), 'fillEnabled');
      var group = input(f.group, { ltr: true, name: 'group', label: t('ctx.group') });
      var src = select(f.source, opts, { name: 'source', label: t('ctx.source') });
      var row = h('div', { class: 'ia-fill flex flex-wrap items-end gap-x-4 border-b border-secondary-100 dark:border-secondary-700 pb-2', 'data-ins-fill': f.group || '' },
        h('div', { class: 'pb-1' }, on), field(t('ctx.group'), group, null, 'flex-1 min-w-[8rem] !mb-1'), field(t('ctx.source'), src, null, 'flex-[2] min-w-[12rem] !mb-1'));
      var reader = function () { return { group: group.value.trim(), source: src.value, enabled: on.input.checked }; };
      readers.push(reader);
      row.appendChild(button({ label: t('delete'), variant: 'danger', data: { 'ins-fill-delete': '' }, onclick: function () { readers.splice(readers.indexOf(reader), 1); row.remove(); } }));
      rows.appendChild(row);
    };
    (c.fills || []).forEach(addRow);
    body.appendChild(panelCard(t('ctx.fills'), h('div', {}, h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 mb-3', text: t('ctx.fillsHint') }), rows,
      h('p', { class: 'text-xs text-secondary-500 mt-3', text: t('ctx.defaults', { v: (def.fills || []).map(function (f) { return f.group + ' ← ' + f.source + (f.enabled ? '' : ' (' + t('ctx.off') + ')'); }).join('، ') }) })),
      [button({ label: t('ctx.addFill'), data: { 'ins-fill-add': '' }, onclick: function () { addRow({ group: '', source: sources.length ? sources[0].value : '', enabled: true }); } })], 'context-fills'));
    var msg = h('div', { 'data-ins-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', text: d.updatedBy ? t('updated', { by: d.updatedBy, at: fmtDate(d.updatedAt) }) : '' });
    var reset = button({ label: t('ctx.reset'), data: { 'ins-context-reset': '' }, onclick: function () {
      clear(rows); readers.length = 0; (def.fills || []).forEach(addRow);
      enabled.input.checked = def.enabled; conflict.value = def.conflict; flag.input.checked = def.flagMismatch; outside.input.checked = def.keepClientOutside; grace.value = digits(String(def.graceSeconds));
    } });
    var save = button({ label: t('save'), variant: 'primary', data: { 'ins-context-save': '' }, onclick: function () {
      var g = numberValue(grace);
      var out = { context: { enabled: enabled.input.checked, conflict: conflict.value, flagMismatch: flag.input.checked, keepClientOutside: outside.input.checked, graceSeconds: g == null ? def.graceSeconds : g,
        fills: readers.map(function (r) { return r(); }).filter(function (f) { return f.group || f.source; }) } };
      busy(save, true);
      clear(msg);
      api('PUT', '/insights/events/settings', out).then(function (x) { busy(save, false); d = x; meta.textContent = t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }); toast('success', t('saved')); }, function (err) { busy(save, false); fieldError(msg, err); });
    } });
    body.appendChild(h('div', { class: 'ia-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), h('div', { class: 'flex gap-2' }, reset, save)));
  }, function (err) { failInto(body, err); });
}
