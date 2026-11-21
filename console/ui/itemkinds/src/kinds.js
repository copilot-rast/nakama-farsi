var IK_BASE = '/itemkinds';
var ikKindCache = null;
var ikKindAt = 0;
var ikKindLoading = null;
var ikCurrencies = null;

function ikLoadKinds(force) {
  if (!force && ikKindCache && Date.now() - ikKindAt < 30000) return Promise.resolve(ikKindCache);
  if (ikKindLoading) return ikKindLoading;
  ikKindLoading = api('GET', IK_BASE + '/kinds').then(function (d) {
    ikKindLoading = null;
    ikKindCache = d.kinds || [];
    ikKindAt = Date.now();
    return ikKindCache;
  }, function (err) {
    ikKindLoading = null;
    throw err;
  });
  return ikKindLoading;
}

function ikKind(key) {
  return (ikKindCache || []).filter(function (k) { return k.key === key; })[0] || null;
}

function loadCurrencies() {
  if (ikCurrencies) return Promise.resolve(ikCurrencies);
  return api('GET', '/v2/console/economy/currencies?limit=100').then(function (d) {
    ikCurrencies = (d && (d.items || d.currencies) || []).map(function (c) { return c.key; });
    return ikCurrencies;
  }, function () { ikCurrencies = []; return ikCurrencies; });
}

function kindPage(root) {
  var key = (location.hash.split('?')[0].split('/')[3] || '');
  var kind = ikKind(key);
  if (!kind) { clear(root).appendChild(h('div', { class: 'px-6' }, notice('error', t('code.KIND_NOT_FOUND')))); return; }
  var q = hashQuery();
  var state = { q: q.q || '', status: q.status || '', parent: q.parent || '', page: 1, limit: 25 };
  var actions = [
    button({ label: t('k.reroute'), data: { 'ik-reroute': kind.key }, onclick: function () { reroutePanel(kind.key, function () { paint(); }); } }),
    button({ label: t('k.new'), variant: 'primary', data: { 'ik-new': kind.key }, onclick: function () { createPanel(kind, function () { paint(); }); } })
  ];
  var frame = pageFrame('itemkinds/kind', t('desc.itemkinds/kind'), actions);
  frame.el.querySelector('h1').textContent = textOf(kind.label, kind.key);
  clear(root).appendChild(frame.el);
  var counts = h('div', { class: 'flex flex-wrap gap-2', 'data-ik-counts': kind.key });
  ['ok', 'pending', 'deleted'].forEach(function (st) { counts.appendChild(h('span', { class: 'inline-flex items-center gap-1 text-sm' }, statusTag(st), h('span', { text: fmtNum(kind.counts && kind.counts[st] || 0) }))); });
  var search = input(state.q, { name: 'q', label: t('search'), placeholder: t('k.searchHint') });
  var status = select(state.status, [{ value: '', label: t('all') }, { value: 'ok', label: t('st.ok') }, { value: 'pending', label: t('st.pending') }, { value: 'deleted', label: t('st.deleted') }, { value: 'unlinked', label: t('st.unlinked') }], { name: 'status', label: t('f.status') });
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.q = search.value.trim(); state.page = 1; paint(); }, 300); });
  status.addEventListener('change', function () { state.status = status.value; state.page = 1; paint(); });
  var filters = h('div', { class: 'ik-filters flex flex-wrap gap-3 items-end' }, field(t('search'), search), field(t('f.status'), status));
  var list = h('div', { 'data-ik-list': kind.key });
  frame.content.appendChild(counts);
  frame.content.appendChild(filters);
  frame.content.appendChild(list);
  function paint() {
    clear(list).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', IK_BASE + '/kinds/' + kind.key + '/items' + qs({ q: state.q, status: state.status, parent: state.parent, page: state.page, limit: state.limit })).then(function (d) {
      clear(list);
      state.total = d.total;
      if (!d.items.length) { list.appendChild(h('div', { class: 'p-6 text-sm text-secondary-500 bg-white dark:bg-secondary-800 rounded-md shadow', 'data-ik-empty': '', text: state.q || state.status ? t('empty') : t('k.empty') })); return; }
      var cols = [{ label: t('f.key') }, { label: t('f.name') }, { label: t('f.status') }, { label: t('f.price') }, { label: t('f.unlockLevel') }];
      (kind.columns || []).forEach(function (c) { cols.push({ label: textOf(c.label, c.key) }); });
      cols.push({ label: t('f.version') });
      var rows = d.items.map(function (it) {
        var cells = [mono(it.key), h('span', { dir: 'auto', text: textOf(it.name, it.key) }), h('span', { class: 'inline-flex gap-1 items-center' }, statusTag(it.status), it.isActive ? null : tag(t('no') + ' · ' + t('f.active'), 'secondary')),
          h('span', { text: priceText(it.price) }), h('span', { text: it.unlockLevel ? fmtNum(it.unlockLevel) : t('none') })];
        (kind.columns || []).forEach(function (c) { cells.push(h('span', { dir: 'auto', text: plainValue(valueAt(it, c.path)) })); });
        cells.push(h('span', { text: it.sourceVersion ? fmtNum(it.sourceVersion) : t('none') }));
        return { id: it.key, item: it, cells: cells };
      });
      list.appendChild(table(cols, rows, function (it) { itemPanel(kind, it.id, paint); }));
      list.appendChild(h('div', { class: 'mt-3' }, pager(state, function (p) { state.page = p; paint(); })));
    }, function (err) { clear(list).appendChild(notice('error', ikErrText(err))); });
  }
  paint();
}

function itemFields(item, currencies) {
  var name = translated(t('f.name'), 'name', item.name || {});
  var cur = select(item.price && item.price.currencyKey || '', [{ value: '', label: t('none') }].concat(currencies.map(function (c) { return { value: c, label: c }; })), { name: 'price.currency', label: t('f.currency') });
  var amount = numInput(item.price ? item.price.amount : '', 'price.amount', t('f.amount'));
  var level = numInput(item.unlockLevel || '', 'unlockLevel', t('f.unlockLevel'));
  var active = check(item.isActive !== false, t('f.active'), 'isActive');
  var el = section(t('k.item'), [h('div', { class: 'ik-row' }, field(t('f.currency'), cur), field(t('f.amount'), amount), field(t('f.unlockLevel'), level)), name, active], 'item');
  return {
    el: el,
    read: function (errors) {
      var out = { name: name.read(), isActive: active.input.checked };
      var amt = latinNumber(amount.value);
      if (cur.value && amt !== '') {
        var n = Number(amt);
        if (!isFinite(n) || n < 0) errors.push(t('fm.badNumber', { field: t('f.amount') })); else out.price = { currencyKey: cur.value, amount: Math.round(n) };
      } else if (!cur.value && amt === '') out.noPrice = true;
      var lv = latinNumber(level.value);
      if (lv === '') out.unlockLevel = 0;
      else if (!isFinite(Number(lv)) || Number(lv) < 0) errors.push(t('fm.badNumber', { field: t('f.unlockLevel') }));
      else out.unlockLevel = Math.round(Number(lv));
      return out;
    }
  };
}

function itemPanel(kind, id, done) {
  var body = h('div', { class: 'flex flex-col gap-4' }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  var msg = h('div', { 'data-ik-msg': '' });
  var save = button({ label: t('save'), variant: 'primary', data: { 'ik-save': 'item' } });
  sidePanel(textOf(kind.label, kind.key), [msg, body], [button({ label: t('cancel'), onclick: closePanel }), save], 'item');
  Promise.all([api('GET', IK_BASE + '/items/' + id), loadCurrencies(), loadLanguages()]).then(function (res) {
    var d = res[0];
    var item = d.item;
    var profile = d.profile || { status: 'unlinked', data: {} };
    clear(body);
    var head = h('div', { class: 'flex flex-wrap items-center gap-2' }, mono(item.key), statusTag(profile.status || 'unlinked'));
    body.appendChild(head);
    if (profile.status === 'pending') body.appendChild(notice('warning', t('k.pending', null).replace('{reason}', profile.reason || '')));
    if (profile.status === 'deleted') body.appendChild(notice('warning', t('k.deleted')));
    if (!d.profile) body.appendChild(notice('info', t('k.unlinked')));
    var itemForm = itemFields(item, res[1]);
    body.appendChild(itemForm.el);
    var prof = schemaForm(kind.fields, profile.data || {}, { name: 'data' });
    body.appendChild(section(t('k.profile'), prof.el, 'profile'));
    if (d.view && Object.keys(d.view).length) body.appendChild(section(t('k.view'), h('pre', { class: 'ik-pre', dir: 'ltr', 'data-ik-view': '', text: JSON.stringify(d.view, null, 2) }), 'view'));
    if (d.hasUnlockPrices) body.appendChild(section(t('k.unlockPrices'), h('div', { class: 'text-sm', 'data-ik-unlock': '' }, (d.unlockPrices || []).length ? d.unlockPrices.map(function (p) { return h('div', { text: priceText(p) }); }) : t('none')), 'unlock'));
    var src = h('dl', { class: 'ik-dl text-sm' });
    var row = function (k, v) { src.appendChild(h('dt', { text: k })); src.appendChild(h('dd', {}, v)); };
    row(t('f.source'), d.source ? h('span', {}, mono(d.source.assetId), ' · ', mono(d.source.type)) : mono(profile.sourceAssetId || t('none')));
    row(t('f.version'), h('span', { text: profile.sourceVersion ? fmtNum(profile.sourceVersion) : t('none') }));
    row(t('f.parent'), mono(item.parentKey || t('none')));
    row(t('f.requires'), mono(item.requiresKey || t('none')));
    row(t('f.updated'), h('span', { text: profile.updatedBy ? t('updated', { by: profile.updatedBy, at: fmtDate(profile.updatedAt) }) : t('none') }));
    body.appendChild(section(t('k.sourceBox'), src, 'source'));
    save.addEventListener('click', function () {
      var errors = [];
      var payload = itemForm.read(errors);
      payload.data = prof.read(JSON.parse(JSON.stringify(profile.data || {})), errors);
      clear(msg);
      if (errors.length) { msg.appendChild(notice('error', errors[0])); return; }
      busy(save, true);
      api('PUT', IK_BASE + '/items/' + id + qs({ kind: kind.key }), payload).then(function () {
        busy(save, false);
        toast('success', t('saved'));
        closePanel();
        if (done) done();
      }, function (err) { busy(save, false); msg.appendChild(notice('error', ikErrText(err))); });
    });
  }, function (err) { clear(body).appendChild(notice('error', ikErrText(err))); });
}

function createPanel(kind, done) {
  var msg = h('div', { 'data-ik-msg': '' });
  var body = h('div', { class: 'flex flex-col gap-4' }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  var save = button({ label: t('k.create'), variant: 'primary', data: { 'ik-save': 'create' } });
  sidePanel(t('k.new') + ' · ' + textOf(kind.label, kind.key), [msg, body], [button({ label: t('cancel'), onclick: closePanel }), save], 'create');
  Promise.all([loadCurrencies(), loadLanguages()]).then(function (res) {
    clear(body);
    var key = input('', { ltr: true, name: 'key', label: t('f.key') });
    body.appendChild(field(t('f.key'), key));
    var itemForm = itemFields({ isActive: true }, res[0]);
    body.appendChild(itemForm.el);
    var prof = schemaForm(kind.fields, {}, { name: 'data' });
    body.appendChild(section(t('k.profile'), prof.el, 'profile'));
    save.addEventListener('click', function () {
      var errors = [];
      var payload = itemForm.read(errors);
      payload.key = key.value.trim();
      payload.data = prof.read({}, errors);
      clear(msg);
      if (errors.length) { msg.appendChild(notice('error', errors[0])); return; }
      busy(save, true);
      api('POST', IK_BASE + '/kinds/' + kind.key + '/items', payload).then(function () {
        busy(save, false);
        toast('success', t('k.created'));
        closePanel();
        ikLoadKinds(true);
        if (done) done();
      }, function (err) { busy(save, false); msg.appendChild(notice('error', ikErrText(err))); });
    });
  });
}

function runRowsTable(rows) {
  if (!rows || !rows.length) return h('p', { class: 'text-sm text-secondary-500', text: t('empty') });
  var cols = [{ label: t('run.asset') }, { label: t('r.sourceType') }, { label: t('r.kind') }, { label: t('run.item') }, { label: t('run.action') }, { label: t('run.reason') }, { label: t('run.changes') }];
  return table(cols, rows.slice(0, 500).map(function (r) {
    var ch = (r.changes || []).map(function (c) { return c.field; }).join('، ');
    return { id: r.assetId, cells: [mono(r.assetId), mono(r.type || ''), h('span', { text: r.kind || t('none') }), mono(r.itemKey || ''), actionTag(r.action), h('span', { dir: 'auto', text: (r.code ? r.code + ' ' : '') + (r.reason || (r.parents && r.parents.length ? r.parents.join(', ') : '')) }), h('span', { dir: 'ltr', class: 'text-xs', text: ch })] };
  }));
}

function summaryTags(sum) {
  var box = h('div', { class: 'flex flex-wrap gap-2', 'data-ik-summary': '' });
  Object.keys(sum || {}).sort().forEach(function (a) { box.appendChild(h('span', { class: 'inline-flex items-center gap-1 text-sm', 'data-ik-sum': a }, actionTag(a), h('span', { text: fmtNum(sum[a]) }))); });
  if (!box.firstChild) box.appendChild(h('span', { class: 'text-sm text-secondary-500', text: t('empty') }));
  return box;
}

function reroutePanel(kindKey, done) {
  var msg = h('div', { 'data-ik-msg': '' });
  var force = check(false, t('rr.force'), 'force');
  var result = h('div', { class: 'flex flex-col gap-3', 'data-ik-result': '' });
  var run = function (dry, btn) {
    clear(msg);
    busy(btn, true);
    clear(result).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('POST', IK_BASE + '/reroute', { dryRun: dry, force: force.input.checked, kind: kindKey || '' }).then(function (out) {
      busy(btn, false);
      clear(result);
      result.appendChild(notice(dry ? 'info' : 'success', dry ? t('rr.previewDone') : t('rr.done')));
      result.appendChild(summaryTags(out.summary));
      result.appendChild(runRowsTable(out.rows));
      if (!dry) { ikLoadKinds(true); if (done) done(); }
    }, function (err) { busy(btn, false); clear(result); msg.appendChild(notice('error', ikErrText(err))); });
  };
  var preview = button({ label: t('rr.preview'), data: { 'ik-preview': '' } });
  var go = button({ label: t('rr.run'), variant: 'primary', data: { 'ik-run': '' } });
  preview.addEventListener('click', function () { run(true, preview); });
  go.addEventListener('click', function () { run(false, go); });
  sidePanel(t('rr.title'), [h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 mb-3', text: t('rr.hint') }), force, msg, result], [button({ label: t('close'), onclick: closePanel }), preview, go], 'reroute');
}
