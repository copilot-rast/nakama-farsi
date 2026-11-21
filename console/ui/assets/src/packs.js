// Content packs: list per platform, which pack the game gets, activate / deactivate / roll back, files and events.
var asPackState = { platform: '', status: '' };

function rangeText(p) {
  if (!p.minAppVersion && !p.maxAppVersion) return t('pk.any');
  if (p.minAppVersion && p.maxAppVersion) return tr('pk.fromTo', {}, { a: p.minAppVersion, b: p.maxAppVersion });
  return p.minAppVersion ? tr('pk.from', {}, { v: p.minAppVersion }) : tr('pk.to', {}, { v: p.maxAppVersion });
}

function packsPage(root) {
  var frame = pageFrame('packs', t('desc.packs'), [button({ label: t('refresh'), data: { 'assets-action': 'refresh-packs' }, onclick: function () { packsPage(root); } })]);
  clear(root).appendChild(frame.el);
  var c = frame.content;
  c.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  Promise.all([api('GET', '/assets/packs' + qs(asPackState)), api('GET', '/assets/packs/events')]).then(function (res) {
    var d = res[0], events = res[1].items || [];
    clear(c);
    var platforms = (d.platforms || []).slice();
    d.items.forEach(function (p) { if (platforms.indexOf(p.platform) === -1) platforms.push(p.platform); });
    var plat = select(asPackState.platform, [{ value: '', label: t('pk.allPlatforms') }].concat(platforms.map(function (p) { return { value: p, label: p }; })), { name: 'platform', label: t('pk.platform') });
    var status = select(asPackState.status, [{ value: '', label: t('all') }, { value: 'active', label: t('pk.active') }, { value: 'ready', label: t('pk.ready') }, { value: 'uploading', label: t('pk.uploading') }], { name: 'status', label: t('pk.status') });
    var change = function () { asPackState.platform = plat.value; asPackState.status = status.value; packsPage(root); };
    plat.addEventListener('change', change);
    status.addEventListener('change', change);
    c.appendChild(notice('secondary', t('pk.upload.hint')));
    c.appendChild(resolverCard(platforms));
    c.appendChild(h('div', { class: 'as-filters flex flex-wrap gap-3 items-end' }, field(t('pk.platform'), plat), field(t('pk.status'), status)));
    if (!d.items.length) c.appendChild(notice('secondary', t('empty')));
    else c.appendChild(h('div', { 'data-assets-list': 'packs' }, table([
      { label: t('pk.platform') }, { label: t('pk.version') }, { label: t('pk.range') }, { label: t('pk.status') }, { label: t('pk.files') }, { label: t('pk.created') }, { label: t('pk.activated') }
    ], d.items.map(function (p) {
      var tags = h('div', { class: 'flex flex-wrap gap-1' }, tag(t('pk.' + p.status), p.status === 'ready' ? 'secondary' : 'warning'),
        p.active ? tag(t('pk.active'), 'success') : null, d.current && d.current[p.platform] === p.id ? tag(t('pk.current'), 'primary') : null);
      return { id: p.version, item: p, cells: [mono(p.platform), mono(p.version), h('span', { dir: 'ltr', text: rangeText(p) }), tags,
        h('span', {}, fmtNum(p.fileCount) + ' · ' + fmtBytes(p.totalBytes)), h('div', {}, h('div', { text: fmtDate(p.createdAt) }), h('div', { class: 'text-xs text-secondary-500', dir: 'auto', text: p.createdBy })),
        p.activatedAt ? h('div', {}, h('div', { text: fmtDate(p.activatedAt) }), h('div', { class: 'text-xs text-secondary-500', dir: 'auto', text: p.activatedBy })) : '—'] };
    }), function (p) { packPanel(p.id, function () { packsPage(root); }); })));
    c.appendChild(section(t('pk.events'), eventList(events), 'events'));
  }, function (err) { clear(c).appendChild(notice('error', asErrText(err))); });
}

function eventList(events) {
  if (!events.length) return h('p', { class: 'text-sm text-secondary-500', text: t('empty') });
  return h('ul', { 'data-assets-events': '' }, events.slice(0, 30).map(function (e) {
    return h('li', { class: 'flex flex-wrap gap-2 py-1.5 text-sm border-b border-secondary-200 dark:border-secondary-700' },
      mono(e.platform), mono(e.version), tag(has('ev.' + e.action) ? t('ev.' + e.action) : e.action, e.action === 'activated' ? 'success' : e.action === 'deactivated' || e.action === 'deleted' ? 'warning' : 'secondary'),
      h('span', { class: 'text-xs text-secondary-500', dir: 'auto', text: fmtDate(e.createdAt) + ' · ' + e.actor }));
  }));
}

function resolverCard(platforms) {
  var plat = select(platforms[0] || 'Android', platforms.map(function (p) { return { value: p, label: p }; }), { name: 'resolve-platform', label: t('pk.platform') });
  var ver = input('', { name: 'resolve-version', ltr: true, placeholder: '1.2.0', label: t('pk.appVersion') });
  var out = h('div', { class: 'mt-2', 'data-assets-resolved': '' });
  var go = button({ label: t('pk.resolve'), data: { 'assets-action': 'resolve' }, onclick: function () {
    clear(out);
    api('GET', '/assets/packs/preview' + qs({ platform: plat.value, appVersion: ver.value.trim() })).then(function (d) {
      out.appendChild(d.pack ? notice('success', tr('pk.resolved', {}, { v: d.pack.version, id: d.pack.id.slice(0, 8) })) : notice('warning', t('pk.noPack')));
    }, function (err) { out.appendChild(notice('error', asErrText(err))); });
  } });
  return section(t('pk.resolver'), h('div', {}, h('div', { class: 'as-filters flex flex-wrap gap-3 items-end' }, field(t('pk.platform'), plat), field(t('pk.appVersion'), ver), go), out), 'resolver');
}

function packPanel(id, done) {
  api('GET', '/assets/packs/' + id).then(function (d) {
    var p = d.item;
    var act = function (path, confirmKey, okLabel) {
      confirmBox(tr(confirmKey, {}, { v: p.version, p: p.platform }), okLabel).then(function (ok) {
        if (!ok) return;
        api(path === 'delete' ? 'DELETE' : 'POST', '/assets/packs/' + p.id + (path === 'delete' ? '' : '/' + path), path === 'delete' ? undefined : {}).then(function () {
          toast('success', t('saved')); closePanel(); done();
        }, function (err) { toast('error', asErrText(err)); });
      });
    };
    var buttons = [];
    if (p.status === 'ready' && !p.active) {
      buttons.push(button({ label: p.activatedAt ? t('pk.rollback') : t('pk.activate'), variant: 'primary', data: { 'assets-action': 'activate-pack' }, onclick: function () { act('activate', p.activatedAt ? 'pk.rollback.confirm' : 'pk.activate.confirm', t('pk.activate')); } }));
    }
    if (p.active) buttons.push(button({ label: t('pk.deactivate'), data: { 'assets-action': 'deactivate-pack' }, onclick: function () { act('deactivate', 'pk.deactivate.confirm', t('pk.deactivate')); } }));
    if (!p.active) buttons.push(button({ label: t('delete'), variant: 'danger', data: { 'assets-action': 'delete-pack' }, onclick: function () { act('delete', 'pk.delete', t('delete')); } }));
    var urls = d.index ? h('div', { class: 'flex flex-col gap-2', 'data-assets-urls': '' }, [['pk.baseUrl', d.index.baseUrl], ['pk.catalogUrl', d.index.catalogUrl], ['pk.hashUrl', d.index.hashUrl], ['pk.catalogHash', d.index.catalogHash]].filter(function (x) { return x[1]; }).map(function (x) {
      return h('div', {}, h('div', { class: 'text-xs text-secondary-500', text: t(x[0]) }), h('div', { class: 'flex items-center gap-2' }, h('code', { class: 'as-pre flex-1', dir: 'ltr', text: x[1] }), copyButton(x[1], x[0])));
    })) : null;
    var head = h('div', { class: 'flex flex-wrap gap-2 items-center mb-4' }, mono(p.platform), mono(p.version), tag(t('pk.' + p.status), 'secondary'), p.active ? tag(t('pk.active'), 'success') : tag(t('pk.inactive'), 'secondary'),
      h('span', { class: 'text-sm', dir: 'ltr', text: rangeText(p) }), h('span', { class: 'text-sm text-secondary-500', text: fmtNum(p.fileCount) + ' · ' + fmtBytes(p.totalBytes) }));
    var files = table([{ label: 'path' }, { label: t('f.size') }, { label: t('f.hash') }], (d.files || []).slice(0, 300).map(function (f) {
      return { id: f.path, cells: [mono(f.path), fmtBytes(f.size), h('span', {}, mono(f.sha256.slice(0, 12) + '…'), f.stored ? null : tag(t('pk.notStored'), 'warning'))] };
    }));
    var content = h('div', {}, head, p.note ? h('p', { class: 'text-sm mb-3', dir: 'auto', text: t('pk.note') + ': ' + p.note }) : null,
      urls ? section(t('pk.urls'), urls, 'urls') : null, h('div', { class: 'h-3' }), section(t('pk.files'), files, 'files'), h('div', { class: 'h-3' }), section(t('pk.events'), eventList(d.events || []), 'pack-events'));
    sidePanel(p.platform + ' · ' + p.version, content, [button({ label: t('close'), onclick: closePanel })].concat(buttons), 'pack');
  }, function (err) { toast('error', asErrText(err)); });
}
