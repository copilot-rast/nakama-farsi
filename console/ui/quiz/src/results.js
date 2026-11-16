function resultsPage(root) {
  var data = null;
  var frame = pageFrame('results', t('desc.results'), [button({ label: t('rs.new'), variant: 'primary', data: { 'quiz-action': 'new-template' }, onclick: function () { templateEditor(null, data, reload); } })]);
  clear(root).appendChild(frame.el);
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(listBox);
  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/templates'), api('GET', '/quiz/missions'), loadRefs(true)]).then(function (res) {
      data = res[0];
      data.missions = res[1].items || [];
      var items = data.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'templates' }, table([
        { label: t('f.name') }, { label: t('rs.mission') }, { label: t('rs.webhook') }, { label: '' }
      ], items.map(function (x) {
        return { id: x.key, item: x, cells: [
          h('div', { class: 'flex flex-col' }, h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: x.name }), x.isActive ? null : tag(t('inactive'), 'warning')), mono(x.key)),
          h('span', { text: x.missionKey || t('none') }),
          h('span', { class: 'text-xs', dir: 'ltr', text: x.webhookUrl || t('none') }),
          h('div', { class: 'flex gap-1' }, button({ label: t('rs.deliveries'), data: { 'quiz-deliveries': x.key }, onclick: function () { deliveriesPanel(x); } }),
            button({ label: t('delete'), variant: 'danger', data: { 'quiz-delete': x.key }, onclick: function () {
              confirmBox(t('rs.delete', { key: x.key }), t('delete')).then(function (ok) {
                if (ok) api('DELETE', '/quiz/templates/' + x.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', qzErrText(err)); });
              });
            } }))
        ] };
      }), function (x) { templateEditor(x, data, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('rs.empty') })));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function templateEditor(x, data, done) {
  var creating = !x;
  x = x || { name: '', body: '', wonText: '', lostText: '', isActive: true };
  var key = input(x.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = input(x.name, { name: 'name' });
  var mission = select(x.missionId || '', [{ value: '', label: t('none') }].concat((data.missions || []).map(function (m) { return { value: m.id, label: t('ms.kind.' + m.kind) + ' · ' + textOf(m.name, m.key) }; })), { name: 'missionId', label: t('rs.mission') });
  var bodyIn = input(x.body, { name: 'body', multiline: true, rows: 8 });
  var won = input(x.wonText, { name: 'wonText', placeholder: 'won' });
  var lost = input(x.lostText, { name: 'lostText', placeholder: 'lost' });
  var url = input(x.webhookUrl, { name: 'webhookUrl', ltr: true, placeholder: 'https://' });
  var secret = input('', { name: 'webhookSecret', ltr: true, type: 'password', placeholder: x.hasSecret ? t('rs.secretKept') : '' });
  var active = check(x.isActive, t('f.active'), 'isActive');
  var chips = h('div', { class: 'flex flex-wrap gap-1 mb-4', 'data-quiz-placeholders': '' }, (data.placeholders || []).map(function (p) {
    return h('button', { type: 'button', class: 'qz-chip', title: t('rs.insert'), onclick: function () {
      var at = bodyIn.selectionStart || bodyIn.value.length;
      bodyIn.value = bodyIn.value.slice(0, at) + '{' + p.key + '}' + bodyIn.value.slice(at);
      bodyIn.focus();
    }, text: '{' + p.key + '}' });
  }));
  var msg = h('div', {});
  var body = h('div', { 'data-quiz-form': 'template' },
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.key'), key), field(t('f.name'), name)), field(t('rs.mission'), mission, t('hint.rsMission')),
    field(t('rs.body'), bodyIn), h('p', { class: 'text-xs text-secondary-500 mb-1', text: t('rs.placeholders') }), chips,
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('rs.won'), won), field(t('rs.lost'), lost)),
    section(t('rs.webhook'), h('div', {}, field(t('rs.url'), url, t('hint.webhook')), field(t('rs.secret'), secret, data.secrets ? t('hint.secret') : t('rs.noSecrets'))), 'webhook'),
    h('div', { class: 'h-3' }), active, msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-template' }, onclick: function () {
    busy(save, true);
    clear(msg);
    var out = { key: key.value.trim(), name: name.value.trim(), missionId: mission.value, body: bodyIn.value, wonText: won.value.trim(), lostText: lost.value.trim(), webhookUrl: url.value.trim(), isActive: active.input.checked };
    if (secret.value) out.webhookSecret = secret.value;
    api(creating ? 'POST' : 'PUT', '/quiz/templates' + (creating ? '' : '/' + x.id), out).then(function () {
      closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); msg.appendChild(notice('error', qzErrText(err))); });
  } });
  sidePanel(creating ? t('rs.new') : x.name, body, [button({ label: t('cancel'), onclick: closePanel }), save], 'template');
}

function deliveriesPanel(x) {
  var body = h('div', { 'data-quiz-deliveries-list': x.key }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  sidePanel(t('rs.deliveries') + ' · ' + x.name, body, [button({ label: t('close'), onclick: closePanel })], 'deliveries');
  api('GET', '/quiz/templates/' + x.id + '/deliveries').then(function (d) {
    var items = d.items || [];
    clear(body).appendChild(items.length ? table([{ label: t('rp.date') }, { label: t('rp.player') }, { label: t('rp.status') }, { label: t('rs.attempts') }, { label: t('rs.error') }],
      items.map(function (it) {
        return { id: it.id, cells: [h('span', { class: 'text-xs', text: fmtDate(it.createdAt) }), playerLink(it.userId), statusTag(it.status), h('span', { text: fmtNum(it.attempts) }),
          h('span', { class: 'text-xs', dir: 'ltr', text: it.lastError || '' })] };
      })) : h('p', { class: 'text-sm text-secondary-500', text: t('empty') }));
  }, function (err) { clear(body).appendChild(notice('error', qzErrText(err))); });
}
