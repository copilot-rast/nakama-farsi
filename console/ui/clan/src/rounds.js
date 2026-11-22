var CL_ROUND_COLOR = { signup: 'primary', active: 'warning', settled: 'success', cancelled: 'secondary' };

function roundTag(s) {
  return tag(t('rd.st.' + s), CL_ROUND_COLOR[s] || 'secondary');
}

function roundsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 25, total: 0, status: q.status || '', configKey: q.configKey || '' };
  var frame = pageFrame('clans/rounds', t('desc.clans/rounds'), [
    button({ label: t('rd.new'), variant: 'primary', data: { 'clan-action': 'new-round' }, onclick: function () { roundCreator(reload); } })]);
  clear(root).appendChild(frame.el);
  var status = select(state.status, [{ value: '', label: t('all') }].concat(opt('rd.st.', ['signup', 'active', 'settled', 'cancelled'])), { name: 'status' });
  var cfg = select(state.configKey, [{ value: '', label: t('all') }], { name: 'configKey' });
  status.addEventListener('change', function () { state.status = status.value; state.page = 1; reload(); });
  cfg.addEventListener('change', function () { state.configKey = cfg.value; state.page = 1; reload(); });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx cl-filters' }, field(t('f.state'), status), field(t('wr.config'), cfg)));
  frame.content.appendChild(listBox);
  loadWarConfigs(true).then(function (list) {
    list.forEach(function (c) { cfg.appendChild(h('option', { value: c.key, text: textOf(c.name, c.key) + ' · ' + c.key })); });
    cfg.value = state.configKey;
  }, function () {});
  function reload() {
    clear(listBox).appendChild(loading());
    Promise.all([api('GET', '/clans/wars/rounds' + qs({ status: state.status, configKey: state.configKey, page: state.page, limit: state.limit })), loadLanguages()]).then(function (res) {
      var d = res[0];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-clan-list': 'rounds' }, table([
        { label: t('f.name') }, { label: t('wr.mode') }, { label: t('f.state') }, { label: t('rd.starts') }, { label: t('rd.ends') }, { label: t('rd.wars') }, { label: t('rd.signups') }
      ], items.map(function (r) {
        return { id: r.id, item: r, cells: [
          nameCell(textOf(r.name, r.configKey), r.configKey),
          h('span', { text: t('mode.' + r.mode) }),
          roundTag(r.status),
          h('span', { class: 'text-xs', text: fmtDate(r.startsAt) }),
          h('span', { class: 'text-xs', text: fmtDate(r.endsAt) }),
          h('span', { text: fmtNum(r.wars || 0) }),
          h('span', { text: fmtNum(r.signups || 0) })
        ] };
      }), function (r) { roundPanel(r.id, reload); })));
      if (!items.length) listBox.appendChild(emptyCard(t('rd.empty')));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}

function roundCreator(done) {
  var body = h('div', { 'data-clan-form': 'round' }, loading());
  var create = button({ label: t('rd.create'), variant: 'primary', data: { 'clan-action': 'create-round' }, disabled: true });
  sidePanel(t('rd.new'), body, [button({ label: t('cancel'), onclick: closePanel }), create], 'new-round');
  loadWarConfigs(true).then(function (list) {
    clear(body);
    if (!list.length) { body.appendChild(notice('warning', t('rd.noConfigs'))); return; }
    var cfg = select(list[0].key, list.map(function (c) { return { value: c.key, label: textOf(c.name, c.key) + ' · ' + t('mode.' + c.mode) }; }), { name: 'configKey' });
    var starts = localInput(null, 'startsAt');
    var dur = numInput('', 'durationMinutes');
    var now = check(true, t('rd.startNow'), 'startNow');
    var msg = h('div', {});
    var startsField = field(t('rd.starts'), starts, t('rd.hint.starts'));
    function paint() { startsField.hidden = now.input.checked; }
    now.input.addEventListener('change', paint);
    paint();
    body.appendChild(h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('rd.hint') }));
    body.appendChild(field(t('wr.config'), cfg));
    body.appendChild(now);
    body.appendChild(startsField);
    body.appendChild(field(t('wr.duration'), dur, t('rd.hint.duration')));
    body.appendChild(msg);
    create.disabled = false;
    create.addEventListener('click', function () {
      clear(msg);
      var payload = { configKey: cfg.value, startNow: now.input.checked, durationMinutes: intOr(dur, 0) };
      if (!now.input.checked) {
        var at = starts.read();
        if (!at) { msg.appendChild(notice('error', t('rd.needStart'))); return; }
        payload.startsAt = at;
      }
      busy(create, true);
      api('POST', '/clans/wars/rounds', payload).then(function (out) { toast('success', t('rd.created')); done(); roundPanel(out.id, done, out); }, function (err) { busy(create, false); fieldError(msg, err); });
    });
  }, function (err) { clear(body).appendChild(notice('error', clErrText(err))); });
}

function roundPanel(id, done, first) {
  var body = h('div', { 'data-clan-round': id }, loading());
  var foot = h('div', { class: 'flex flex-wrap gap-2 justify-end w-full' });
  sidePanel(t('rd.detail'), body, [foot], 'round');
  function act(path, confirmText, okLabel, btn) {
    (confirmText ? confirmBox(confirmText, okLabel) : Promise.resolve(true)).then(function (ok) {
      if (!ok) return;
      busy(btn, true);
      api('POST', '/clans/wars/rounds/' + id + '/' + path, {}).then(function (out) { toast('success', t('rd.done.' + path)); paint(out); if (done) done(); }, function (err) { busy(btn, false); toast('error', clErrText(err)); });
    });
  }
  function paint(r) {
    clear(body);
    clear(foot);
    var open = r.status === 'signup' || r.status === 'active';
    if (r.status === 'signup') { var sb = button({ label: t('rd.start'), variant: 'primary', data: { 'clan-action': 'start-round' }, onclick: function () { act('start', null, null, sb); } }); foot.appendChild(sb); }
    if (r.status === 'active') { var st = button({ label: t('rd.settle'), variant: 'primary', data: { 'clan-action': 'settle-round' }, onclick: function () { act('settle', t('rd.settleConfirm'), t('rd.settle'), st); } }); foot.appendChild(st); }
    if (open) { var cb = button({ label: t('rd.cancel'), variant: 'danger', data: { 'clan-action': 'cancel-round' }, onclick: function () { act('cancel', t('rd.cancelConfirm'), t('rd.cancel'), cb); } }); foot.insertBefore(cb, foot.firstChild); }
    foot.appendChild(button({ label: t('close'), onclick: closePanel }));
    body.appendChild(h('div', { class: 'mb-4' }, dl([
      [t('f.name'), h('span', { dir: 'auto', text: textOf(r.name, r.configKey) })],
      [t('wr.config'), mono(r.configKey)],
      [t('f.state'), h('span', { 'data-clan-value': 'round-status' }, roundTag(r.status))],
      [t('wr.mode'), t('mode.' + r.mode)],
      [t('wr.signup'), t('wr.signup.' + r.signup)],
      [t('wr.pairing'), t('wr.pair.' + r.pairing)],
      r.signupAt ? [t('rd.signupAt'), fmtDate(r.signupAt)] : null,
      [t('rd.starts'), fmtDate(r.startsAt)],
      [t('rd.ends'), fmtDate(r.endsAt)],
      r.settledAt ? [t('rd.settledAt'), fmtDate(r.settledAt)] : null,
      [t('rd.actor'), h('span', { dir: 'auto', text: r.actor || t('ss.scheduler') })]
    ], 'round')));
    var wars = r.items || [];
    var duel = r.mode === 'duel';
    body.appendChild(section(t('rd.warsT', { n: fmtNum(wars.length) }), wars.length ? h('div', { class: 'cl-rows' }, wars.map(function (w) {
      var winA = w.result === 'a', winB = w.result === 'b';
      var res = w.status === 'active' ? tag(t('rd.st.active'), 'warning') : w.result === 'draw' ? tag(t('out.draw'), 'secondary') : w.result === 'bye' ? tag(t('out.bye'), 'secondary')
        : w.result ? tag(t('rd.winner', { name: winA ? w.nameA : w.nameB }), 'success') : tag(t('out.cancelled'), 'secondary');
      return h('div', { class: 'cl-war', 'data-clan-war': w.warId },
        h('div', { class: 'cl-war-side' + (winA ? ' cl-war-win' : '') }, h('span', { class: 'font-medium', dir: 'auto', text: w.nameA || w.clanA.slice(0, 8) }),
          h('span', { class: 'text-xs text-secondary-500', text: t('rd.members', { n: fmtNum(w.membersA) }) + (duel ? ' · ' + t('rd.wins', { n: fmtNum(w.winsA) }) : '') })),
        h('div', { class: 'cl-war-score', dir: 'ltr', 'data-clan-score': '', text: digits(String(w.scoreA)) + ' : ' + (w.clanB ? digits(String(w.scoreB)) : '–') }),
        h('div', { class: 'cl-war-side' + (winB ? ' cl-war-win' : '') }, w.clanB ? h('span', { class: 'font-medium', dir: 'auto', text: w.nameB || w.clanB.slice(0, 8) }) : h('span', { class: 'text-secondary-500', text: t('rd.noOpponent') }),
          w.clanB ? h('span', { class: 'text-xs text-secondary-500', text: t('rd.members', { n: fmtNum(w.membersB) }) + (duel ? ' · ' + t('rd.wins', { n: fmtNum(w.winsB) }) : '') }) : null),
        h('div', { class: 'cl-row-full flex flex-wrap gap-2 items-center' }, res, h('span', { class: 'text-xs text-secondary-500', text: t('rd.events', { n: fmtNum(w.events) }) })));
    })) : h('p', { class: 'text-sm text-secondary-500', text: r.status === 'signup' ? t('rd.notStarted') : t('rd.noWars') }), 'wars'));
    body.appendChild(h('div', { class: 'h-3' }));
    var su = r.signupList || [];
    body.appendChild(section(t('rd.signupsT', { n: fmtNum(su.length) }), su.length ? h('ul', { class: 'flex flex-col gap-1 text-sm' }, su.map(function (s) {
      return h('li', { class: 'flex justify-between gap-2' }, h('span', { dir: 'auto', text: s.name || s.clanId.slice(0, 8) }), h('span', { class: 'text-xs text-secondary-500', text: fmtDate(s.createdAt) }));
    })) : h('p', { class: 'text-sm text-secondary-500', text: r.signup === 'auto' ? t('rd.autoSignup') : t('rd.noSignups') }), 'signups'));
  }
  if (first) { loadLanguages().then(function () { paint(first); }); return; }
  Promise.all([api('GET', '/clans/wars/rounds/' + id), loadLanguages()]).then(function (res) { paint(res[0]); }, function (err) { clear(body).appendChild(notice('error', clErrText(err))); });
}
