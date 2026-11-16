function missionsPage(root) {
  var items = [];
  var frame = pageFrame('missions', t('desc.missions'), [
    button({ label: t('ms.newGame'), data: { 'quiz-action': 'new-game' }, onclick: function () { missionEditor(null, 'game', items, reload); } }),
    button({ label: t('ms.new'), variant: 'primary', data: { 'quiz-action': 'new-mission' }, onclick: function () { missionEditor(null, 'mission', items, reload); } })]);
  clear(root).appendChild(frame.el);
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(listBox);
  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/missions'), loadLanguages(), loadRefs(true)]).then(function (res) {
      items = res[0].items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'missions' }, table([
        { label: t('f.name') }, { label: t('ms.kind') }, { label: t('ms.game') }, { label: t('ms.order') }, { label: t('ms.time') }, { label: t('ms.items') }, { label: t('f.reward') }, { label: '' }
      ], items.map(function (m) {
        return { id: m.key, item: m, cells: [
          h('div', { class: 'flex flex-col' }, h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(m.name, m.key) }), m.isActive ? null : tag(t('inactive'), 'warning')), mono(m.key)),
          tag(t('ms.kind.' + m.kind), m.kind === 'game' ? 'primary' : 'secondary'),
          h('span', { text: m.parentKey || t('none') }),
          h('span', { text: t('ms.order.' + m.orderMode) }),
          h('span', { text: m.timeSeconds ? fmtNum(m.timeSeconds) : t('none') }),
          m.kind === 'mission' ? h('a', { class: 'eco-link', href: '#/quiz/packs?missionId=' + m.id, text: t('ms.packs', { n: fmtNum(m.items) }) }) : h('span', { text: t('ms.missions', { n: fmtNum(m.items) }) }),
          h('span', { class: 'text-xs', text: rewardText(m.reward) }),
          button({ label: t('delete'), variant: 'danger', data: { 'quiz-delete': m.key }, onclick: function () {
            confirmBox(t('ms.delete', { key: m.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/quiz/missions/' + m.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', err.code === 'IN_USE' ? t('ms.inUse') : qzErrText(err)); });
            });
          } })
        ] };
      }), function (m) { missionEditor(m, m.kind, items, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('ms.empty') })));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function missionEditor(m, kind, all, done) {
  var creating = !m;
  m = m || { kind: kind, name: {}, description: {}, isActive: true, reward: {}, orderMode: 'all_at_once', metadata: {} };
  var games = all.filter(function (x) { return x.kind === 'game'; });
  var key = input(m.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var parent = select(m.parentId || '', [{ value: '', label: t('ms.noGame') }].concat(games.map(function (g) { return { value: g.id, label: textOf(g.name, g.key) }; })), { name: 'parentId', label: t('ms.game') });
  var name = translated(t('f.name'), 'name', m.name);
  var desc = translated(t('f.description'), 'description', m.description, { multiline: true });
  var order = select(m.orderMode, ['all_at_once', 'sequential'].map(function (o) { return { value: o, label: t('ms.order.' + o) }; }), { name: 'orderMode', label: t('ms.order') });
  var time = numInput(m.timeSeconds || '', 'timeSeconds');
  var send = check(m.sendAnswers, t('pk.sendAnswers'), 'sendAnswers');
  var icon = input(m.iconUrl, { name: 'iconUrl', ltr: true });
  var sort = numInput(m.sort || 0, 'sort');
  var active = check(m.isActive, t('f.active'), 'isActive');
  var reward = rewardEditor(m.reward, 'mission');
  var meta = jsonInput(m.metadata, 'metadata');
  var msg = h('div', {});
  var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
  var body = h('div', { 'data-quiz-form': 'mission' },
    h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('ms.hint.' + m.kind) }),
    grid(field(t('f.key'), key, t('hint.key')), m.kind === 'mission' ? field(t('ms.game'), parent) : h('div', {})), name, desc,
    grid(field(t('ms.order'), order, t('ms.orderHint.' + m.kind)), field(t('ms.timeF'), time, t('hint.missionTime'))),
    m.kind === 'mission' ? send : null, grid(field(t('f.icon'), icon), field(t('f.sort'), sort)), active,
    section(t('ms.reward'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.missionReward') }), reward), 'reward'),
    h('div', { class: 'h-3' }), field(t('f.metadata'), meta), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-mission' }, onclick: function () {
    busy(save, true);
    clear(msg);
    var md = meta.read();
    if (md === undefined) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_JSON'))); return; }
    var out = { key: key.value.trim(), kind: m.kind, parentId: m.kind === 'mission' ? parent.value : '', name: name.read(), description: desc.read(), orderMode: order.value,
      timeSeconds: numberValue(time) || 0, sendAnswers: send.input.checked, iconUrl: icon.value.trim(), sort: numberValue(sort) || 0, isActive: active.input.checked, reward: reward.read(), metadata: md };
    api(creating ? 'POST' : 'PUT', '/quiz/missions' + (creating ? '' : '/' + m.id), out).then(function () {
      closePanel(); toast('success', t('saved')); qzRefs = null; done();
    }, function (err) { busy(save, false); msg.appendChild(notice('error', qzErrText(err))); });
  } });
  sidePanel(creating ? t(m.kind === 'game' ? 'ms.newGame' : 'ms.new') : textOf(m.name, m.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'mission');
}
