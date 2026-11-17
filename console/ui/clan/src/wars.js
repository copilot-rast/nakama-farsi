var clWarDefaults = null;
var clWarConfigs = null;
function loadWarConfigs(force) {
  if (clWarConfigs && !force) return Promise.resolve(clWarConfigs);
  return api('GET', '/clans/wars/configs').then(function (d) { clWarConfigs = d.items || []; clWarDefaults = d.defaults || null; return clWarConfigs; });
}

function eventsFor(type) {
  return type === 'game_event' ? refList('clans', 'gameEvents') : refList('clans', 'economyEvents');
}

function warsPage(root) {
  var frame = pageFrame('clans/wars', t('desc.clans/wars'), [
    button({ label: t('wr.new'), variant: 'primary', data: { 'clan-action': 'new-war' }, onclick: function () { warEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  var box = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(box);
  function reload() {
    clear(box).appendChild(loading());
    Promise.all([loadWarConfigs(true), loadLanguages(), loadRefs('clans')]).then(function (res) {
      var items = res[0];
      clear(box);
      box.appendChild(h('div', { 'data-clan-list': 'war-configs' }, table([
        { label: t('f.name') }, { label: t('wr.mode') }, { label: t('wr.schedule') }, { label: t('wr.duration') }, { label: t('wr.signup') }, { label: t('wr.next') }, { label: t('f.state') }, { label: '' }
      ], items.map(function (c) {
        return { id: c.key, item: c, cells: [
          nameCell(textOf(c.name, c.key), c.key),
          tag(t('mode.' + c.mode), c.mode === 'duel' ? 'warning' : 'primary'),
          c.schedule ? mono(c.schedule) : h('span', { text: t('wr.manual') }),
          h('span', { text: t('wr.minutes', { n: fmtNum(c.durationMinutes) }) }),
          h('span', { text: t('wr.signup.' + c.signup) + (c.signupMinutes ? ' · ' + t('wr.minutes', { n: fmtNum(c.signupMinutes) }) : '') }),
          h('span', { class: 'text-xs', text: c.nextStart ? fmtDate(c.nextStart) : t('none') }),
          c.isActive ? tag(t('active'), 'success') : tag(t('inactive'), 'warning'),
          button({ label: t('delete'), variant: 'danger', data: { 'clan-delete': c.key }, onclick: function () {
            confirmBox(t('wr.delete', { key: c.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/clans/wars/configs/' + c.id).then(function () { toast('success', t('deleted')); reload(); }, function (err) { toast('error', clErrText(err)); });
            });
          } })
        ] };
      }), function (c) { warEditor(c, reload); })));
      if (!items.length) box.appendChild(emptyCard(t('wr.empty')));
    }, function (err) { clear(box).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}

function pointRows(list) {
  var box = h('div', { class: 'cl-rows', 'data-clan-rules': 'points' });
  var rows = [];
  function addRow(p) {
    p = p || { type: 'economy_event', event: '', points: 1 };
    var type = select(p.type, opt('src.', ['economy_event', 'game_event']), { name: 'point-type' });
    var ev = eventInput(eventsFor(p.type), p.event, 'point-event', 'match_end');
    type.addEventListener('change', function () {
      var list = document.getElementById(ev.input.getAttribute('list'));
      if (!list) return;
      clear(list);
      eventsFor(type.value).forEach(function (e) { list.appendChild(h('option', { value: e.key, text: e.name || e.key })); });
    });
    var fld = input(p.field, { name: 'point-field', ltr: true, mono: true, placeholder: 'won' });
    var eq = input(p.equals, { name: 'point-equals', ltr: true, mono: true, placeholder: 'true' });
    var pts = numInput(p.points, 'point-points', t('wr.points'));
    var per = check(p.perValue, t('wr.perValue'), 'point-perValue');
    var rec = { type: type, ev: ev, fld: fld, eq: eq, pts: pts, per: per };
    var row = h('div', { class: 'cl-row', 'data-clan-point': '' }, field(t('wr.source'), type), field(t('wr.event'), ev), field(t('wr.field'), fld), field(t('wr.equals'), eq),
      field(t('wr.points'), pts), h('div', { class: 'cl-row-full flex items-center justify-between gap-2' }, per,
        button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r !== rec; }); } })));
    rows.push(rec);
    box.appendChild(row);
  }
  (list || []).forEach(addRow);
  var wrap = h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('wr.hint.points') }), box,
    h('div', { class: 'mt-2' }, button({ label: '+ ' + t('wr.addPoint'), data: { 'clan-action': 'add-point' }, onclick: function () { addRow(); } })));
  wrap.read = function () {
    return rows.map(function (r) {
      var o = { type: r.type.value, event: r.ev.input.value.trim(), points: intOr(r.pts, 0) };
      if (r.fld.value.trim()) o.field = r.fld.value.trim();
      if (r.eq.value.trim()) o.equals = r.eq.value.trim();
      if (r.per.input.checked) o.perValue = true;
      return o;
    });
  };
  return wrap;
}

function warEditor(c, done) {
  var creating = !c;
  var defs = clWarDefaults || {};
  c = c ? JSON.parse(JSON.stringify(c)) : { name: {}, description: {}, mode: 'points', isActive: true, schedule: '', durationMinutes: 1440, signup: 'auto', signupMinutes: 0, pairing: 'random',
    rules: JSON.parse(JSON.stringify(defs.points || { points: [], duel: {}, eligibility: {}, prizes: {} })) };
  var r = c.rules || {};
  r.duel = r.duel && r.duel.type ? r.duel : JSON.parse(JSON.stringify((defs.duel && defs.duel.duel) || { type: 'economy_event', wonField: 'won', drawField: 'draw', opponentField: 'opponentId', winPoints: 1 }));
  r.eligibility = r.eligibility || {};
  r.prizes = r.prizes || {};
  var prizes = refList('clans', 'prizes');
  var key = input(c.key || '', { name: 'key', ltr: true, mono: true });
  var name = translated(t('f.name'), 'name', c.name);
  var desc = translated(t('f.description'), 'description', c.description, { multiline: true });
  var mode = select(c.mode, opt('mode.', ['points', 'duel']), { name: 'mode' });
  var active = check(c.isActive, t('f.active'), 'isActive');
  var sched = input(c.schedule, { name: 'schedule', ltr: true, mono: true, placeholder: '0 18 * * 5' });
  var dur = numInput(c.durationMinutes, 'durationMinutes');
  var signup = select(c.signup, opt('wr.signup.', ['auto', 'leader']), { name: 'signup' });
  var signupMin = numInput(c.signupMinutes, 'signupMinutes');
  var pairing = select(c.pairing, opt('wr.pair.', ['random', 'level', 'trophies']), { name: 'pairing' });
  var points = pointRows(r.points);
  var d = r.duel;
  var dType = select(d.type, opt('src.', ['economy_event', 'game_event']), { name: 'duel.type' });
  var dEvent = eventInput(eventsFor(d.type), d.event, 'duel.event', 'match_end');
  var dWon = input(d.wonField, { name: 'duel.wonField', ltr: true, mono: true });
  var dDraw = input(d.drawField, { name: 'duel.drawField', ltr: true, mono: true });
  var dOpp = input(d.opponentField, { name: 'duel.opponentField', ltr: true, mono: true });
  var dWin = numInput(d.winPoints, 'duel.winPoints');
  var dDrawP = numInput(d.drawPoints, 'duel.drawPoints');
  var duelBox = h('div', { 'data-clan-rules': 'duel' }, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('wr.hint.duel') }),
    grid(field(t('wr.source'), dType), field(t('wr.event'), dEvent)), grid3(field(t('wr.won'), dWon), field(t('wr.drawF'), dDraw), field(t('wr.opp'), dOpp, t('wr.hint.opp'))),
    grid(field(t('wr.winPoints'), dWin), field(t('wr.drawPoints'), dDrawP)));
  var matchF = input(r.matchField, { name: 'rules.matchField', ltr: true, mono: true });
  var attacks = numInput(r.attacksPerMember || 0, 'rules.attacksPerMember');
  var e = r.eligibility;
  var minM = numInput(e.minMembers || 1, 'eligibility.minMembers');
  var maxM = numInput(e.maxMembers || 0, 'eligibility.maxMembers');
  var minL = numInput(e.minClanLevel || 0, 'eligibility.minClanLevel');
  var minT = numInput(e.minTrophies || 0, 'eligibility.minTrophies');
  var joined = numInput(e.joinedHours || 0, 'eligibility.joinedHours');
  var pz = r.prizes;
  var pWin = refSelect(prizes, (pz.win || {}).prizeKey, 'prizes.win', t('none'));
  var pLose = refSelect(prizes, (pz.lose || {}).prizeKey, 'prizes.lose', t('none'));
  var pDraw = refSelect(prizes, (pz.draw || {}).prizeKey, 'prizes.draw', t('none'));
  var pBye = refSelect(prizes, (pz.bye || {}).prizeKey, 'prizes.bye', t('none'));
  var pTo = select(pz.to || 'participants', opt('wr.to.', ['participants', 'members']), { name: 'prizes.to' });
  var bye = select(r.byeResult || 'none', opt('wr.bye.', ['none', 'win']), { name: 'byeResult' });
  var pointsSec = section(t('wr.pointsRules'), points, 'points');
  var duelSec = section(t('wr.duelRules'), duelBox, 'duel');
  function paint() { pointsSec.hidden = mode.value !== 'points'; duelSec.hidden = mode.value !== 'duel'; }
  mode.addEventListener('change', paint);
  paint();
  var msg = h('div', {});
  var body = h('div', { 'data-clan-form': 'war' },
    h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('wr.hint') }),
    field(t('f.key'), key, t('hint.key')), name, desc,
    grid(field(t('wr.mode'), mode, t('wr.hint.mode')), h('div', { class: 'pt-6' }, active)),
    section(t('wr.timing'), h('div', {}, grid(field(t('wr.schedule'), sched, t('wr.hint.schedule')), field(t('wr.duration'), dur, t('wr.hint.duration'))),
      grid3(field(t('wr.signup'), signup, t('wr.hint.signup')), field(t('wr.signupMinutes'), signupMin, t('wr.hint.signupMinutes')), field(t('wr.pairing'), pairing))), 'timing'), h('div', { class: 'h-3' }),
    pointsSec, duelSec, h('div', { class: 'h-3' }),
    section(t('wr.scoring'), grid(field(t('wr.matchField'), matchF, t('wr.hint.matchField')), field(t('wr.attacks'), attacks, t('wr.hint.attacks'))), 'scoring'), h('div', { class: 'h-3' }),
    section(t('wr.eligibility'), h('div', {}, grid3(field(t('wr.minMembers'), minM), field(t('wr.maxMembers'), maxM, t('hint.zero')), field(t('wr.minLevel'), minL)),
      grid(field(t('wr.minTrophies'), minT), field(t('wr.joined'), joined, t('wr.hint.joined')))), 'eligibility'), h('div', { class: 'h-3' }),
    section(t('wr.prizes'), h('div', {}, grid(field(t('wr.pWin'), pWin), field(t('wr.pLose'), pLose)), grid(field(t('wr.pDraw'), pDraw), field(t('wr.pBye'), pBye)),
      grid(field(t('wr.to'), pTo), field(t('wr.byeResult'), bye, t('wr.hint.bye')))), 'prizes'), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'clan-action': 'save-war' }, onclick: function () {
    clear(msg);
    var rules = { points: points.read(), duel: { type: dType.value, event: dEvent.input.value.trim(), wonField: dWon.value.trim(), drawField: dDraw.value.trim(), opponentField: dOpp.value.trim(),
        winPoints: intOr(dWin, 0), drawPoints: intOr(dDrawP, 0) },
      matchField: matchF.value.trim(), attacksPerMember: intOr(attacks, 0),
      eligibility: { minMembers: intOr(minM, 1), maxMembers: intOr(maxM, 0), minClanLevel: intOr(minL, 0), minTrophies: intOr(minT, 0), joinedHours: intOr(joined, 0) },
      prizes: { win: { prizeKey: pWin.value }, lose: { prizeKey: pLose.value }, draw: { prizeKey: pDraw.value }, bye: { prizeKey: pBye.value }, to: pTo.value }, byeResult: bye.value };
    if (mode.value === 'points') rules.duel = { type: '', event: '', wonField: '', drawField: '', opponentField: '', winPoints: 0, drawPoints: 0 };
    var out = { key: key.value.trim(), name: name.read(), description: desc.read(), mode: mode.value, isActive: active.input.checked, schedule: sched.value.trim(), durationMinutes: intOr(dur, 0),
      signup: signup.value, signupMinutes: intOr(signupMin, 0), pairing: pairing.value, rules: rules };
    busy(save, true);
    api(creating ? 'POST' : 'PUT', '/clans/wars/configs' + (creating ? '' : '/' + c.id), out).then(function () {
      clWarConfigs = null; closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  sidePanel(creating ? t('wr.new') : textOf(c.name, c.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'war');
}
