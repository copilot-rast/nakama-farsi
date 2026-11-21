function routesPage(root) {
  var frame = pageFrame('itemkinds/routes', t('desc.itemkinds/routes'), [
    button({ label: t('rr.title'), data: { 'ik-reroute': 'all' }, onclick: function () { reroutePanel('', function () { paint(); }); } }),
    button({ label: t('r.new'), variant: 'primary', data: { 'ik-route-new': '' }, onclick: function () { routePanel(null, meta, paint); } })
  ]);
  clear(root).appendChild(frame.el);
  var list = h('div', { 'data-ik-routes': '' });
  frame.content.appendChild(list);
  var meta = { kinds: [], roles: [] };
  function paint() {
    clear(list).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', IK_BASE + '/routes').then(function (d) {
      meta = d;
      clear(list);
      if (!d.routes.length) { list.appendChild(h('div', { class: 'p-6 text-sm text-secondary-500 bg-white dark:bg-secondary-800 rounded-md shadow', text: t('empty') })); return; }
      var cols = [{ label: t('r.sourceType') }, { label: t('r.match') }, { label: t('r.role') }, { label: t('r.kind') }, { label: t('r.priority') }, { label: t('r.enabled') }, { label: t('r.origin') }];
      list.appendChild(table(cols, d.routes.map(function (r) {
        var match = Object.keys(r.match || {}).length ? JSON.stringify(r.match) : t('all');
        return { id: String(r.id), item: r, cells: [mono(r.sourceType), h('span', { dir: 'ltr', class: 'font-mono text-xs', text: match }), tag(t('role.' + r.role), r.role === 'item' ? 'primary' : 'secondary'),
          h('span', { text: r.kind ? textOf((ikKind(r.kind) || {}).label, r.kind) : t('none') }), h('span', { text: fmtNum(r.priority) }), tag(r.enabled ? t('yes') : t('no'), r.enabled ? 'success' : 'secondary'),
          h('span', { text: r.origin ? t('r.yes') : t('r.custom') })] };
      }), function (r) { routePanel(r, meta, paint); }));
    }, function (err) { clear(list).appendChild(notice('error', ikErrText(err))); });
  }
  paint();
}

function routePanel(route, meta, done) {
  var r = route || { sourceType: '', match: {}, role: 'item', kind: (meta.kinds || [])[0] || '', priority: 100, enabled: true, note: '' };
  var msg = h('div', { 'data-ik-msg': '' });
  var src = input(r.sourceType, { ltr: true, name: 'sourceType', label: t('r.sourceType') });
  var match = input(Object.keys(r.match || {}).length ? JSON.stringify(r.match, null, 2) : '', { multiline: true, rows: 4, ltr: true, mono: true, name: 'match', label: t('r.match') });
  var role = select(r.role, (meta.roles || ['item', 'dependent', 'ignore']).map(function (x) { return { value: x, label: t('role.' + x) }; }), { name: 'role', label: t('r.role') });
  var kind = select(r.kind || '', [{ value: '', label: t('none') }].concat((meta.kinds || []).map(function (k) { return { value: k, label: textOf((ikKind(k) || {}).label, k) }; })), { name: 'kind', label: t('r.kind') });
  var prio = numInput(r.priority, 'priority', t('r.priority'));
  var on = check(r.enabled !== false, t('r.enabled'), 'enabled');
  var note = input(r.note || '', { name: 'note', label: t('r.note') });
  var kindField = field(t('r.kind'), kind);
  var syncKind = function () { kindField.hidden = role.value !== 'item'; };
  role.addEventListener('change', syncKind);
  syncKind();
  var save = button({ label: t('save'), variant: 'primary', data: { 'ik-save': 'route' }, onclick: function () {
    clear(msg);
    var m = {};
    var raw = match.value.trim();
    if (raw) { try { m = JSON.parse(raw); } catch (e) { msg.appendChild(notice('error', t('r.badJson'))); return; } }
    if (!m || typeof m !== 'object' || Array.isArray(m)) { msg.appendChild(notice('error', t('r.badJson'))); return; }
    var p = Number(latinNumber(prio.value) || 100);
    var body = { sourceType: src.value.trim(), match: m, role: role.value, kind: role.value === 'item' ? kind.value : '', priority: isFinite(p) ? Math.round(p) : 100, enabled: on.input.checked, note: note.value.trim() };
    busy(save, true);
    api(route ? 'PUT' : 'POST', IK_BASE + '/routes' + (route ? '/' + route.id : ''), body).then(function () {
      busy(save, false); toast('success', t('saved')); closePanel(); done();
    }, function (err) { busy(save, false); msg.appendChild(notice('error', ikErrText(err))); });
  } });
  var footer = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (route) footer.unshift(button({ label: t('delete'), variant: 'danger', data: { 'ik-route-delete': String(route.id) }, onclick: function () {
    confirmBox(route.origin ? t('r.deleteDefault') : t('r.deleteConfirm'), t('delete')).then(function (ok) {
      if (!ok) return;
      api('DELETE', IK_BASE + '/routes/' + route.id).then(function () { toast('success', t('saved')); closePanel(); done(); }, function (err) { msg.appendChild(notice('error', ikErrText(err))); });
    });
  } }));
  sidePanel(route ? t('r.editTitle') : t('r.new'), [msg, field(t('r.sourceType'), src, t('r.sourceHint')), field(t('r.match'), match, t('r.matchHint')), h('div', { class: 'ik-row' }, field(t('r.role'), role), kindField, field(t('r.priority'), prio)), on, field(t('r.note'), note)], footer, 'route');
}
