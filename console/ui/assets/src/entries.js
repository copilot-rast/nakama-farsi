var asEntryState = { folder: '', recursive: true, type: '', q: '', page: 1, limit: 50, tab: 'live' };

function entriesPage(root) {
  var frame = pageFrame('entries', t('desc.entries'), [
    button({ label: t('en.import'), data: { 'assets-action': 'import' }, onclick: function () { importPanel(function () { entriesPage(root); }); } }),
    button({ label: t('en.export'), data: { 'assets-action': 'export' }, onclick: exportEntries }),
    button({ label: t('en.new'), variant: 'primary', data: { 'assets-action': 'new-entry' }, onclick: function () { entryPanel(null, function () { entriesPage(root); }); } })
  ]);
  clear(root).appendChild(frame.el);
  var top = h('div', { class: 'flex flex-col gap-4' });
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(top);
  frame.content.appendChild(body);
  top.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  api('GET', '/assets/overview').then(function (ov) {
    clear(top);
    top.appendChild(kpis([
      { name: 'entries', label: t('kpi.entries'), value: fmtNum(ov.total) },
      { name: 'size', label: t('kpi.size'), value: fmtBytes(ov.totalBytes) },
      { name: 'folders', label: t('kpi.folders'), value: fmtNum((ov.folders || []).length) },
      { name: 'types', label: t('kpi.types'), value: fmtNum((ov.types || []).length) }
    ]));
    var tabs = h('div', { class: 'flex gap-2', role: 'tablist' }, ['live', 'deleted'].map(function (tb) {
      return button({ label: t('en.tab.' + tb), variant: asEntryState.tab === tb ? 'primary' : 'ghost', data: { 'assets-tab': tb }, onclick: function () { asEntryState.tab = tb; entriesPage(root); } });
    }));
    top.appendChild(tabs);
    if (asEntryState.tab === 'deleted') { deletedList(body, root); return; }
    var folders = folderOptions(ov.folders || []);
    var folderSel = select(asEntryState.folder, [{ value: '', label: t('f.allFolders') }].concat(folders), { name: 'folder', label: t('f.folder') });
    var typeSel = select(asEntryState.type, [{ value: '', label: t('f.allTypes') }].concat((ov.types || []).map(function (x) { return { value: x.key, label: x.key + ' (' + digits(String(x.count)) + ')' }; })), { name: 'type', label: t('f.type') });
    var q = input(asEntryState.q, { name: 'q', placeholder: t('search'), label: t('search') });
    var rec = check(asEntryState.recursive, t('f.recursive'), 'recursive');
    var reload = function () { asEntryState.folder = folderSel.value; asEntryState.type = typeSel.value; asEntryState.q = q.value.trim(); asEntryState.recursive = rec.input.checked; asEntryState.page = 1; loadList(); };
    folderSel.addEventListener('change', reload);
    typeSel.addEventListener('change', reload);
    rec.input.addEventListener('change', reload);
    var timer = null;
    q.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(reload, 300); });
    body.appendChild(h('div', { class: 'eco-toolbar eco-tbx as-filters' }, field(t('f.folder'), folderSel), field(t('f.type'), typeSel), field(t('search'), q), rec));
    var list = h('div', { 'data-assets-list': 'entries' });
    body.appendChild(list);
    function loadList() {
      clear(list).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
      var s = asEntryState;
      api('GET', '/assets/entries' + qs({ folder: s.folder, recursive: (s.folder === '/' || !s.recursive) ? 'false' : '', type: s.type, q: s.q, page: s.page, limit: s.limit })).then(function (d) {
        clear(list);
        if (!d.items.length) { list.appendChild(notice('secondary', t('empty'))); return; }
        list.appendChild(table([
          { label: t('f.assetId') }, { label: t('f.name') }, { label: t('f.type') }, { label: t('f.folder') }, { label: t('f.version') }, { label: t('f.size') }, { label: t('f.updated') }
        ], d.items.map(function (e) {
          return { id: e.assetId, item: e, cells: [mono(e.assetId), h('span', { dir: 'auto', text: e.name }), mono(e.type), mono(e.folder || '/'), tag(t('en.v', { v: e.version }), 'primary'), fmtBytes(e.size),
            h('div', {}, h('div', { text: fmtDate(e.updatedAt) }), h('div', { class: 'text-xs text-secondary-500', dir: 'auto', text: e.updatedBy }))] };
        }), function (e) { entryPanel(e.assetId, loadList); }));
        list.appendChild(h('div', { class: 'mt-2' }, pager({ page: d.page, limit: d.limit, total: d.total }, function (p) { asEntryState.page = p; loadList(); })));
      }, function (err) { clear(list).appendChild(notice('error', asErrText(err))); });
    }
    loadList();
  }, function (err) { clear(top).appendChild(notice('error', asErrText(err))); });
}

function folderOptions(rows) {
  var counts = {};
  rows.forEach(function (r) {
    var parts = r.key ? r.key.split('/') : [];
    if (!parts.length) { counts[''] = (counts[''] || 0) + r.count; return; }
    for (var i = 1; i <= parts.length; i++) { var p = parts.slice(0, i).join('/'); counts[p] = (counts[p] || 0) + r.count; }
  });
  return Object.keys(counts).filter(function (k) { return k !== ''; }).sort().map(function (k) {
    var depth = k.split('/').length - 1;
    return { value: k, label: new Array(depth + 1).join('  ') + k.split('/').pop() + ' (' + digits(String(counts[k])) + ')' };
  }).concat(counts[''] ? [{ value: '/', label: t('f.root') + ' (' + digits(String(counts[''])) + ')' }] : []);
}

function prettyJSON(v) {
  try { return JSON.stringify(v, null, 2); } catch (e) { return String(v); }
}

function entryPanel(id, done) {
  if (!id) { entryForm(null, [], done); return; }
  api('GET', '/assets/entries/' + encodeURIComponent(id)).then(function (d) { entryForm(d.item, d.history || [], done); }, function (err) { toast('error', asErrText(err)); });
}

function entryForm(item, history, done) {
  var isNew = !item;
  item = item || { assetId: '', type: '', folder: '', name: '', data: {}, version: 0 };
  var idIn = input(item.assetId, { name: 'assetId', ltr: true, mono: true, label: t('f.assetId') });
  if (!isNew) idIn.setAttribute('readonly', '');
  var typeIn = input(item.type, { name: 'type', ltr: true, mono: true, label: t('f.type') });
  var folderIn = input(item.folder, { name: 'folder', ltr: true, mono: true, label: t('f.folder') });
  var nameIn = input(item.name, { name: 'name', label: t('f.name') });
  var dataIn = input(prettyJSON(item.data), { name: 'data', multiline: true, rows: 14, ltr: true, mono: true, label: t('f.data') });
  var msg = h('div', { 'data-assets-msg': '' });
  var save = button({ label: t('save'), variant: 'primary', data: { 'assets-action': 'save-entry' }, onclick: function () {
    clear(msg);
    var data;
    try { data = JSON.parse(dataIn.value); } catch (e) { data = null; }
    if (!data || typeof data !== 'object' || Array.isArray(data)) { msg.appendChild(notice('error', t('en.badJson'))); return; }
    var body = { assetId: idIn.value.trim(), type: typeIn.value.trim(), folder: folderIn.value.trim(), name: nameIn.value.trim(), data: data };
    if (!isNew) body.baseVersion = item.version;
    busy(save, true);
    api(isNew ? 'POST' : 'PUT', isNew ? '/assets/entries' : '/assets/entries/' + encodeURIComponent(item.assetId), body).then(function () {
      busy(save, false); toast('success', t('saved')); closePanel(); done();
    }, function (err) {
      busy(save, false);
      if (err.code === 'ASSET_VERSION_CONFLICT') { msg.appendChild(conflictNotice(err, function () { entryPanel(item.assetId, done); })); return; }
      fieldError(msg, err);
    });
  } });
  var del = isNew ? null : button({ label: t('delete'), variant: 'danger', data: { 'assets-action': 'delete-entry' }, onclick: function () {
    confirmBox(tr('en.delete', {}, { id: item.assetId }), t('delete')).then(function (ok) {
      if (!ok) return;
      api('DELETE', '/assets/entries/' + encodeURIComponent(item.assetId) + '?baseVersion=' + item.version).then(function () { toast('success', t('saved')); closePanel(); done(); }, function (err) {
        if (err.code === 'ASSET_VERSION_CONFLICT') { clear(msg).appendChild(conflictNotice(err, function () { entryPanel(item.assetId, done); })); return; }
        fieldError(msg, err);
      });
    });
  } });
  var meta = isNew ? null : h('div', { class: 'flex flex-wrap gap-2 items-center mb-4 text-xs text-secondary-500' },
    tag(t('en.v', { v: item.version }), 'primary'), mono(item.hash ? item.hash.slice(0, 16) + '…' : ''), item.sourceGuid ? mono(item.sourceGuid) : null,
    h('span', { dir: 'auto', text: fmtDate(item.updatedAt) + ' · ' + (item.updatedBy || '') }));
  var hist = isNew ? null : historyBox(item, history, done);
  var content = h('div', {}, meta,
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.assetId'), idIn), field(t('f.type'), typeIn), field(t('f.folder'), folderIn), field(t('f.name'), nameIn)),
    field(t('f.data'), dataIn), msg, hist);
  sidePanel(isNew ? t('en.create') : tr('en.edit', {}, { id: item.assetId }), content, [del, h('div', { class: 'flex-1' }), button({ label: t('cancel'), onclick: closePanel }), save], 'entry');
}

function conflictNotice(err, reload) {
  var c = (err.meta && err.meta.conflicts || [])[0] || {};
  var box = notice('warning', t('en.conflict', { v: c.version == null ? '—' : c.version }));
  box.setAttribute('data-assets-conflict', '');
  box.appendChild(button({ label: t('en.reload'), data: { 'assets-action': 'reload-entry' }, onclick: reload }));
  return box;
}

function historyBox(item, history, done) {
  var view = h('div', {});
  var rows = history.map(function (hs) {
    return h('li', { class: 'flex flex-wrap items-center gap-2 py-2 border-b border-secondary-200 dark:border-secondary-700', 'data-assets-history': hs.version },
      tag(t('en.v', { v: hs.version }), 'secondary'), actionTag(hs.action),
      h('span', { class: 'text-xs text-secondary-500 flex-1', dir: 'auto', text: fmtDate(hs.createdAt) + ' · ' + hs.actor }),
      button({ label: t('en.view'), data: { 'assets-action': 'view-version' }, onclick: function () {
        api('GET', '/assets/history/' + hs.id).then(function (d) {
          clear(view).appendChild(h('div', { class: 'mt-3' }, h('div', { class: 'text-sm font-medium mb-1', text: t('en.state', { action: t('act.' + d.item.action), v: d.item.version }) }),
            h('pre', { class: 'as-pre', dir: 'ltr', 'data-assets-state': d.item.version, text: prettyJSON(d.item.data) })));
        }, function (err) { toast('error', asErrText(err)); });
      } }),
      hs.version === item.version && hs.action !== 'deleted' ? null : button({ label: t('en.restore'), data: { 'assets-action': 'restore-version' }, onclick: function () {
        confirmBox(t('en.restore.confirm', { v: hs.version }), t('en.restore')).then(function (ok) {
          if (!ok) return;
          api('POST', '/assets/entries/' + encodeURIComponent(item.assetId) + '/rollback', { historyId: hs.id, baseVersion: item.version }).then(function (d) {
            toast('success', t('en.restored', { v: d.item.version })); closePanel(); done();
          }, function (err) { toast('error', asErrText(err)); });
        });
      } }));
  });
  return section(t('en.history'), h('div', {}, h('ul', { 'data-assets-history-list': '' }, rows), view), 'history');
}

function deletedList(body, root) {
  body.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  api('GET', '/assets/deleted').then(function (d) {
    clear(body);
    if (!d.items.length) { body.appendChild(notice('secondary', t('en.noDeleted'))); return; }
    body.appendChild(table([{ label: t('f.assetId') }, { label: t('f.type') }, { label: t('f.folder') }, { label: t('f.version') }, { label: t('f.updated') }, { label: '' }],
      d.items.map(function (hs) {
        return { id: hs.assetId, item: hs, cells: [mono(hs.assetId), mono(hs.type), mono(hs.folder || '/'), tag(t('en.v', { v: hs.version }), 'secondary'),
          h('span', { dir: 'auto', text: fmtDate(hs.createdAt) + ' · ' + hs.actor }),
          button({ label: t('en.restore'), data: { 'assets-action': 'restore-deleted' }, onclick: function () {
            api('POST', '/assets/entries/' + encodeURIComponent(hs.assetId) + '/rollback', { historyId: hs.id }).then(function (r) {
              toast('success', t('en.restored', { v: r.item.version })); entriesPage(root);
            }, function (err) { toast('error', asErrText(err)); });
          } })] };
      })));
  }, function (err) { clear(body).appendChild(notice('error', asErrText(err))); });
}

function exportEntries() {
  var s = asEntryState;
  api('GET', '/assets/export' + qs({ folder: s.folder, recursive: (s.folder === '/' || !s.recursive) ? 'false' : '' })).then(function (d) {
    var blob = new Blob([JSON.stringify(d, null, 2)], { type: 'application/json' });
    var a = h('a', { href: URL.createObjectURL(blob), download: 'assets-export.json' });
    document.body.appendChild(a); a.click(); a.remove();
  }, function (err) { toast('error', asErrText(err)); });
}

function importPanel(done) {
  var area = input('', { multiline: true, rows: 12, ltr: true, mono: true, name: 'importJson', label: t('en.import') });
  area.setAttribute('placeholder', '{"items":[{"assetId":"Tiggo8","type":"VehicleCard","folder":"Cars","name":"Tiggo 8","data":{"speed":180}}]}');
  var force = check(false, t('en.import.force'), 'force');
  var msg = h('div', { 'data-assets-msg': '' });
  var report = h('div', { 'data-assets-report': '' });
  var parse = function () {
    var body;
    try { body = JSON.parse(area.value); } catch (e) { return null; }
    if (Array.isArray(body)) body = { items: body };
    return body && typeof body === 'object' ? body : null;
  };
  var run = function (dry, btn) {
    clear(msg); clear(report);
    var body = parse();
    if (!body) { msg.appendChild(notice('error', t('en.badJson'))); return; }
    body.dryRun = dry;
    body.force = force.input.checked;
    busy(btn, true);
    api('POST', '/assets/push', body).then(function (r) {
      busy(btn, false);
      report.appendChild(notice(r.conflicts.length ? 'warning' : 'success', t('en.import.done', { c: r.created, u: r.updated, s: r.unchanged, d: r.deleted }) + (r.conflicts.length ? ' · ' + t('en.import.conflicts', { n: r.conflicts.length }) : '')));
      report.appendChild(h('div', { class: 'mt-2' }, table([{ label: t('f.assetId') }, { label: '' }, { label: t('f.version') }], r.rows.map(function (x) {
        return { id: x.assetId, cells: [mono(x.assetId), actionTag(x.action), x.version == null ? '—' : t('en.v', { v: x.version })] };
      }))));
      if (!dry) { toast('success', t('saved')); done(); }
    }, function (err) {
      busy(btn, false);
      if (err.code === 'ASSET_VERSION_CONFLICT') {
        var list = (err.meta && err.meta.conflicts || []).map(function (c) { return c.assetId; }).join(', ');
        msg.appendChild(notice('warning', t('code.ASSET_VERSION_CONFLICT') + ' ' + list));
        return;
      }
      fieldError(msg, err);
    });
  };
  var checkBtn = button({ label: t('en.import.check'), data: { 'assets-action': 'import-check' }, onclick: function () { run(true, checkBtn); } });
  var applyBtn = button({ label: t('en.import.apply'), variant: 'primary', data: { 'assets-action': 'import-apply' }, onclick: function () { run(false, applyBtn); } });
  sidePanel(t('en.import'), h('div', {}, h('p', { class: 'text-sm text-secondary-500 mb-3', dir: 'auto', text: t('en.import.hint') }), field('', area), force, msg, report),
    [button({ label: t('cancel'), onclick: closePanel }), checkBtn, applyBtn], 'import');
}
