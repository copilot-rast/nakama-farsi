var clLeagues = null;
function loadLeagues(force) {
  if (clLeagues && !force) return Promise.resolve(clLeagues);
  return api('GET', '/trophies/leagues').then(function (d) { clLeagues = d; return d; });
}

function leagueName(key) {
  if (!key) return t('tr.unranked');
  var l = clLeagues && (clLeagues.items || []).filter(function (x) { return x.key === key; })[0];
  return l ? textOf(l.name, l.key) : key;
}

function leaguesPage(root) {
  var frame = pageFrame('trophies/leagues', t('desc.trophies/leagues'), [
    button({ label: t('lg.new'), variant: 'primary', data: { 'clan-action': 'new-league' }, onclick: function () { leagueEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  var box = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(box);
  function reload() {
    clear(box).appendChild(loading());
    Promise.all([loadLeagues(true), loadLanguages(), loadRefs('trophies')]).then(function (res) {
      var d = res[0];
      var items = d.items || [];
      clear(box);
      box.appendChild(h('div', { class: 'flex flex-wrap gap-2', 'data-clan-summary': 'leagues' },
        tag(t('lg.players', { n: fmtNum(d.players || 0) }), 'primary'), tag(t('lg.unranked', { n: fmtNum(d.unranked || 0) }), 'secondary')));
      box.appendChild(h('div', { 'data-clan-list': 'leagues' }, table([
        { label: t('f.name') }, { label: t('lg.min') }, { label: t('lg.rules') }, { label: t('lg.prize') }, { label: t('lg.count') }, { label: t('f.state') }, { label: '' }
      ], items.map(function (l) {
        return { id: l.key, item: l, cells: [
          nameCell(textOf(l.name, l.key), l.key, l.iconUrl, l.color),
          h('span', { text: t('lg.from', { n: fmtNum(l.minTrophies) }) }),
          h('div', { class: 'flex flex-wrap gap-1' }, l.floor ? tag(t('lg.floorOn'), 'warning') : null,
            l.resetTo != null ? tag(t('lg.resetTo', { n: fmtNum(l.resetTo) }), 'secondary') : null, !l.floor && l.resetTo == null ? h('span', { class: 'text-secondary-400', text: t('none') }) : null),
          l.prizeKey ? mono(l.prizeKey) : h('span', { class: 'text-secondary-400', text: t('none') }),
          h('span', { text: fmtNum(l.players || 0) }),
          l.isActive ? tag(t('active'), 'success') : tag(t('inactive'), 'warning'),
          button({ label: t('delete'), variant: 'danger', data: { 'clan-delete': l.key }, onclick: function () {
            confirmBox(t('lg.delete', { key: l.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/trophies/leagues/' + l.id).then(function () { toast('success', t('deleted')); reload(); }, function (err) { toast('error', clErrText(err)); });
            });
          } })
        ] };
      }), function (l) { leagueEditor(l, reload); })));
      if (!items.length) box.appendChild(emptyCard(t('lg.empty')));
    }, function (err) { clear(box).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}

function leagueEditor(l, done) {
  var creating = !l;
  l = l || { name: {}, description: {}, isActive: true, floor: false, resetTo: null, minTrophies: 0, color: '', iconUrl: '', prizeKey: '', metadata: {} };
  var key = input(l.key || '', { name: 'key', ltr: true, mono: true });
  var name = translated(t('f.name'), 'name', l.name);
  var desc = translated(t('f.description'), 'description', l.description, { multiline: true });
  var icon = input(l.iconUrl, { name: 'iconUrl', ltr: true });
  var color = colorField(l.color);
  var min = numInput(l.minTrophies || 0, 'minTrophies');
  var floor = check(l.floor, t('lg.floor'), 'floor');
  var reset = numInput(l.resetTo == null ? '' : l.resetTo, 'resetTo');
  var prize = refSelect(refList('trophies', 'prizes'), l.prizeKey, 'prizeKey', t('none'));
  var active = check(l.isActive, t('f.active'), 'isActive');
  var meta = jsonInput(l.metadata, 'metadata');
  var msg = h('div', {});
  var body = h('div', { 'data-clan-form': 'league' },
    h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('lg.hint') }),
    field(t('f.key'), key, t('hint.key')), name, desc,
    grid(field(t('f.icon'), icon), field(t('f.color'), color)),
    grid(field(t('lg.min'), min, t('lg.hint.min')), field(t('lg.reset'), reset, t('lg.hint.reset'))),
    floor, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('lg.hint.floor') }),
    field(t('lg.prize'), prize, t('lg.hint.prize')), active,
    field(t('f.metadata'), meta, t('hint.json')), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'clan-action': 'save-league' }, onclick: function () {
    clear(msg);
    var m = meta.read();
    if (m === undefined) { msg.appendChild(notice('error', t('code.INVALID_JSON') + ' · metadata')); return; }
    var r = numberValue(reset);
    var out = { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), color: color.input.value.trim(), minTrophies: intOr(min, 0),
      floor: floor.input.checked, resetTo: r == null || isNaN(r) ? null : r, prizeKey: prize.value, isActive: active.input.checked, metadata: m };
    busy(save, true);
    api(creating ? 'POST' : 'PUT', '/trophies/leagues' + (creating ? '' : '/' + l.id), out).then(function (d) {
      clLeagues = d; closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  sidePanel(creating ? t('lg.new') : textOf(l.name, l.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'league');
}
