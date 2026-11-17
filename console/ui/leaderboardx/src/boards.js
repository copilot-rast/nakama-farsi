function boardsPage(root) {
  var frame = pageFrame('boards', t('desc.boards'), [button({ label: t('new'), variant: 'primary', data: { 'lbx-action': 'new-board' }, onclick: function () { boardEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  function reload() {
    clear(frame.content).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/leaderboardx/boards'), loadLanguages(), loadRefs(true)]).then(function (res) {
      var list = res[0].items || [];
      clear(frame.content);
      if (!list.length) { frame.content.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      frame.content.appendChild(h('div', { 'data-lbx-list': 'boards' }, table([
        { label: t('f.name') }, { label: t('f.key') }, { label: t('f.period') }, { label: t('f.source') }, { label: t('sec.tiers') }, { label: t('sec.rewards') }, { label: t('wn.players') }, { label: '' }
      ], list.map(function (x) {
        var b = x.board;
        var src = t('src.' + b.source.type) + (b.source.event ? ' · ' + b.source.event : '');
        return { id: b.key, item: x, cells: [
          h('div', { class: 'flex flex-wrap items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(b.name, b.key) }), statusTag(x.status), b.scope === 'clan' ? h('span', { class: 'eco-tag eco-tag--primary inline-flex leading-tight px-2 py-0.5 text-xs rounded-md whitespace-nowrap', 'data-lbx-scope': b.key, text: t('b.clan') }) : null),
          mono(b.key),
          h('div', { class: 'flex flex-col text-xs' }, h('span', { text: t('period.' + b.period) }), h('span', { class: 'text-secondary-500', text: x.periodEnd ? t('b.ends', { at: fmtDate(x.periodEnd) }) : t('b.noEnd') })),
          h('span', { class: 'text-xs', 'data-lbx-source': b.key, text: src }),
          h('span', { class: 'text-xs', text: b.tierBasis === 'none' ? t('basis.none') : t('basis.' + b.tierBasis) + ' · ' + t('b.tiers', { n: b.tiers.length }) }),
          h('span', { class: 'text-xs', 'data-lbx-rewards': b.key, text: b.rewards.length ? b.rewards.map(function (r) { return rewardSentence(r, b); }).join(' · ') : t('none') }),
          h('a', { href: '#/leaderboardx/standings?board=' + encodeURIComponent(b.key), class: 'eco-link', 'data-lbx-players': b.key, text: t('b.players', { n: x.players }) }),
          button({ label: t('delete'), variant: 'danger', data: { 'lbx-delete': b.key }, onclick: function () {
            confirmBox(tr('b.delete', null, { key: b.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/leaderboardx/boards/' + b.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', err.code === 'IN_USE' ? t('b.inUse') : lbErrText(err)); });
            });
          } })
        ] };
      }), function (x) { boardEditor(x.board, reload); })));
    }, function (err) { clear(frame.content).appendChild(notice('error', lbErrText(err))); });
  }
  reload();
}

function tierRows(b, basisSel) {
  var box = h('div', { class: 'flex flex-col gap-3', 'data-lbx-tiers': '' });
  var rows = [];
  function paint() {
    var basis = basisSel.value;
    rows.forEach(function (r, i) {
      r.label.textContent = t('tier.threshold.' + (basis === 'none' ? 'score' : basis));
      var prev = i > 0 ? numberValue(rows[i - 1].th) : 0;
      r.sentence.textContent = tierSentence(basis, floatValue(r.th) || 0, prev);
    });
    wrap.hidden = basis === 'none';
  }
  function addRow(tier) {
    tier = tier || { key: '', name: {}, threshold: '' };
    var key = input(tier.key, { name: 'tier-key', ltr: true, mono: true });
    var th = numInput(tier.threshold, 'tier-threshold');
    var icon = input(tier.iconUrl, { name: 'tier-icon', ltr: true });
    var name = translated(t('f.name'), 'tier-name', tier.name);
    var label = h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1' });
    var sentence = h('div', { class: 'text-xs text-secondary-500 lb-sentence', 'data-lbx-tier-sentence': '' });
    var el = h('div', { class: 'lb-track', 'data-lbx-tier': '' },
      h('div', { class: 'grid sm:grid-cols-3 gap-x-4 items-end' }, field(t('tier.key'), key), h('label', { class: 'lb-field block mb-4' }, label, th), field(t('f.icon'), icon)),
      name, h('div', { class: 'flex items-center justify-between gap-2' }, sentence,
        button({ label: t('remove'), variant: 'danger', onclick: function () { el.remove(); rows = rows.filter(function (r) { return r.el !== el; }); paint(); } })));
    var row = { el: el, th: th, label: label, sentence: sentence, read: function () { return { key: key.value.trim(), name: name.read(), iconUrl: icon.value.trim(), threshold: floatValue(th) || 0 }; } };
    th.addEventListener('input', paint);
    rows.push(row);
    box.appendChild(el);
    paint();
  }
  var wrap = h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('tier.hint') }), box, button({ label: '+ ' + t('tier.add'), data: { 'lbx-action': 'add-tier' }, onclick: function () { addRow(); } }));
  (b.tiers || []).forEach(addRow);
  basisSel.addEventListener('change', paint);
  paint();
  wrap.read = function () { return basisSel.value === 'none' ? [] : rows.map(function (r) { return r.read(); }); };
  wrap.keys = function () { return rows.map(function (r) { var o = r.read(); return { key: o.key, name: o.name }; }).filter(function (x) { return x.key; }); };
  return wrap;
}

function rewardRows(b, tiers) {
  var box = h('div', { class: 'flex flex-col gap-3', 'data-lbx-reward-rows': '' });
  var rows = [];
  function tierOptions(value) {
    var opts = [{ value: '', label: t('rw.allRanks') }];
    tiers.keys().forEach(function (x) { opts.push({ value: x.key, label: tr('rw.inTier', null, { t: textOf(x.name, x.key) }) }); });
    if (value && !opts.some(function (o) { return o.value === value; })) opts.push({ value: value, label: value });
    return opts;
  }
  function addRow(r) {
    r = r || { fromRank: 1, toRank: 1 };
    var tierSel = select(r.tier || '', tierOptions(r.tier), { name: 'reward-tier', label: t('rw.tier') });
    tierSel.addEventListener('focus', function () {
      var v = tierSel.value;
      clear(tierSel);
      tierOptions(v).forEach(function (o) { var opt = h('option', { value: o.value, text: o.label }); if (o.value === v) opt.selected = true; tierSel.appendChild(opt); });
    });
    var from = numInput(r.fromRank || 1, 'reward-from', t('rw.from'));
    var to = numInput(r.toRank || '', 'reward-to', t('rw.to'));
    var cur = currencyRows(r.currencies, 'reward');
    var prizes = prizePickers(r);
    var sentence = h('div', { class: 'text-xs text-secondary-500 lb-sentence', 'data-lbx-reward-sentence': '' });
    function paint() { sentence.textContent = rewardSentence({ tier: tierSel.value, fromRank: numberValue(from) || 1, toRank: numberValue(to) || 0 }, { tiers: tiers.keys() }); }
    [tierSel, from, to].forEach(function (el) { el.addEventListener('input', paint); el.addEventListener('change', paint); });
    var el = h('div', { class: 'lb-track', 'data-lbx-reward': '' },
      h('div', { class: 'grid sm:grid-cols-3 gap-x-4 items-end' }, field(t('rw.tier'), tierSel), field(t('rw.from'), from), field(t('rw.to'), to)),
      cur, prizes, h('div', { class: 'flex items-center justify-between gap-2 mt-2' }, sentence,
        button({ label: t('remove'), variant: 'danger', onclick: function () { el.remove(); rows = rows.filter(function (x) { return x.el !== el; }); } })));
    rows.push({ el: el, read: function () { var o = prizes.read(); o.tier = tierSel.value; o.fromRank = numberValue(from) || 1; o.toRank = numberValue(to) || 0; o.currencies = cur.read(); return o; } });
    box.appendChild(el);
    paint();
  }
  (b.rewards || []).forEach(addRow);
  var wrap = h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('rw.hint') }), box, button({ label: '+ ' + t('rw.add'), data: { 'lbx-action': 'add-reward' }, onclick: function () { addRow(); } }));
  wrap.read = function () { return rows.map(function (r) { return r.read(); }); };
  return wrap;
}

function boardEditor(b, done) {
  var creating = !b;
  b = b || { name: {}, description: {}, isActive: true, sortOrder: 'desc', operator: 'incr', period: 'weekly', resetSchedule: '', tierBasis: 'none', tiers: [], rewards: [],
    source: { type: 'economy_event', calc: 'value', multiplier: 1 }, limits: { onExceed: 'reject' }, payMode: '', metadata: {} };
  var opt = function (prefix, list) { return list.map(function (v) { return { value: v, label: t(prefix + v) }; }); };
  var key = input(b.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = translated(t('f.name'), 'name', b.name);
  var desc = translated(t('f.description'), 'description', b.description, { multiline: true });
  var icon = input(b.iconUrl, { name: 'iconUrl', ltr: true });
  var sort = numInput(b.sort || 0, 'sort');
  var active = check(b.isActive, t('f.active'), 'isActive');
  var scope = select(b.scope || 'player', opt('scope.', ['player', 'clan']), { name: 'scope' });
  var clanNote = h('p', { class: 'text-xs text-warning-700 dark:text-warning-400 mb-3', 'data-lbx-clan-note': '', text: t('hint.clanNoRewards') });
  var period = select(b.period, opt('period.', ['none', 'daily', 'weekly', 'monthly', 'custom']), { name: 'period' });
  var cron = input(b.period === 'custom' ? b.resetSchedule : '', { name: 'resetSchedule', ltr: true, mono: true, placeholder: '0 0 * * 5' });
  var cronField = field(t('f.cron'), cron, t('hint.cron'));
  var starts = localInput(b.startsAt, 'startsAt');
  var ends = localInput(b.endsAt, 'endsAt');
  var maxPeriods = numInput(b.maxPeriods || 0, 'maxPeriods');
  var sortOrder = select(b.sortOrder, opt('sort.', ['desc', 'asc']), { name: 'sortOrder' });
  var operator = select(b.operator, opt('op.', ['incr', 'best', 'set', 'decr']), { name: 'operator' });
  var srcType = select(b.source.type, opt('src.', ['economy_event', 'game_event', 'client', 'server']), { name: 'source.type' });
  var event = input(b.source.event, { name: 'source.event', ltr: true, mono: true });
  var events = h('datalist', { id: 'lbx-events' });
  ((lbRefs && lbRefs.economyEvents) || []).forEach(function (e) { events.appendChild(h('option', { value: e.key, text: e.name })); });
  event.setAttribute('list', 'lbx-events');
  var context = input(b.source.context, { name: 'source.context', ltr: true, mono: true });
  var fieldIn = input(b.source.field, { name: 'source.field', ltr: true, mono: true, placeholder: 'stars' });
  var calc = select(b.source.calc, opt('calc.', ['value', 'count']), { name: 'source.calc' });
  var mult = input(b.source.multiplier == null ? 1 : digits(String(b.source.multiplier)), { name: 'source.multiplier', ltr: true });
  var eventBox = h('div', {}, field(t('f.event'), event), events);
  var gameBox = h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.context'), context), field(t('f.field'), fieldIn));
  var clientSubmit = check(b.limits.clientSubmit, t('f.clientSubmit'), 'limits.clientSubmit');
  var maxPerWrite = numInput(b.limits.maxPerWrite || 0, 'limits.maxPerWrite');
  var perMinute = numInput(b.limits.perMinute || 0, 'limits.perMinute');
  var onExceed = select(b.limits.onExceed || 'reject', opt('exceed.', ['reject', 'clamp']), { name: 'limits.onExceed' });
  var basis = select(b.tierBasis, opt('basis.', ['none', 'score', 'rank', 'percent']), { name: 'tierBasis' });
  var tiers = tierRows(b, basis);
  var rewards = rewardRows(b, tiers);
  var minScore = numInput(b.rewardMinScore || 0, 'rewardMinScore');
  var payMode = select(b.payMode || '', opt('pay.', ['', 'auto', 'claim']), { name: 'payMode' });
  var clientOpt = srcType.querySelector('option[value="client"]');
  var rewardsSec = null;
  function paint() {
    var clan = scope.value === 'clan';
    if (clan && srcType.value === 'client') srcType.value = 'economy_event';
    if (clientOpt) clientOpt.disabled = clan;
    cronField.hidden = period.value !== 'custom';
    eventBox.hidden = srcType.value !== 'economy_event' && srcType.value !== 'game_event';
    gameBox.hidden = srcType.value !== 'game_event';
    if (clan) { clientSubmit.input.checked = false; clientSubmit.input.disabled = true; }
    else if (srcType.value === 'client') { clientSubmit.input.checked = true; clientSubmit.input.disabled = true; } else clientSubmit.input.disabled = false;
    clanNote.hidden = !clan;
    if (rewardsSec) rewardsSec.hidden = clan;
  }
  [period, srcType, scope].forEach(function (el) { el.addEventListener('change', paint); });
  paint();
  var msg = h('div', { 'data-lbx-msg': '' });
  var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
  rewardsSec = section(t('sec.rewards'), h('div', {}, rewards, h('div', { class: 'h-3' }), grid(field(t('f.rewardMinScore'), minScore), field(t('f.payMode'), payMode))), 'rewards');
  paint();
  var body = h('div', { 'data-lbx-form': 'board' }, field(t('f.key'), key, t('hint.key')), name, desc, grid(field(t('f.icon'), icon), field(t('f.sort'), sort)), active,
    field(t('f.scope'), scope, t('hint.scope')), clanNote,
    section(t('sec.period'), h('div', {}, field(t('f.period'), period), cronField, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.window') }),
      grid(field(t('f.startsAt'), starts), field(t('f.endsAt'), ends)), field(t('f.maxPeriods'), maxPeriods, null, 'lb-num')), 'period'), h('div', { class: 'h-3' }),
    section(t('sec.scoring'), h('div', {}, grid(field(t('f.sortOrder'), sortOrder), field(t('f.operator'), operator)), field(t('f.source'), srcType), eventBox, gameBox,
      grid(field(t('f.calc'), calc), field(t('f.multiplier'), mult, t('hint.multiplier')))), 'scoring'), h('div', { class: 'h-3' }),
    section(t('sec.limits'), h('div', {}, clientSubmit, grid(field(t('f.maxPerWrite'), maxPerWrite), field(t('f.perMinute'), perMinute)), field(t('f.onExceed'), onExceed)), 'limits'), h('div', { class: 'h-3' }),
    section(t('sec.tiers'), h('div', {}, field(t('f.tierBasis'), basis), tiers), 'tiers'), h('div', { class: 'h-3' }),
    rewardsSec, msg);
  function read() {
    var clan = scope.value === 'clan';
    return { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), sort: numberValue(sort) || 0, isActive: active.input.checked, scope: scope.value,
      period: period.value, resetSchedule: period.value === 'custom' ? cron.value.trim() : '', startsAt: starts.read(), endsAt: ends.read(), maxPeriods: numberValue(maxPeriods) || 0,
      sortOrder: sortOrder.value, operator: operator.value,
      source: { type: srcType.value, event: event.value.trim(), context: context.value.trim(), field: fieldIn.value.trim(), calc: calc.value, multiplier: floatValue(mult) || 1 },
      limits: { clientSubmit: !clan && clientSubmit.input.checked, maxPerWrite: numberValue(maxPerWrite) || 0, perMinute: numberValue(perMinute) || 0, onExceed: onExceed.value },
      tierBasis: basis.value, tiers: tiers.read(), rewards: clan ? [] : rewards.read(), rewardMinScore: numberValue(minScore) || 0, payMode: payMode.value, metadata: b.metadata || {} };
  }
  function send(extra) {
    busy(save, true);
    clear(msg);
    var out = read();
    Object.keys(extra || {}).forEach(function (k) { out[k] = extra[k]; });
    api(creating ? 'POST' : 'PUT', '/leaderboardx/boards' + (creating ? '' : '/' + b.id), out).then(function () {
      closePanel(); toast('success', t('saved')); done();
    }, function (err) {
      busy(save, false);
      if (err.code === 'LEADERBOARD_RESTART_REQUIRED') {
        var pay = check(false, t('b.restartPay'), 'payOnRestart');
        msg.appendChild(h('div', { class: 'eco-notice eco-notice--warning rounded-md px-3 py-2 text-sm flex flex-col gap-2', 'data-lbx-restart': '' },
          h('span', { text: t('b.restart', { n: (err.meta && err.meta.players) || 0 }) }), pay,
          h('div', {}, button({ label: t('b.restartGo'), variant: 'primary', data: { 'lbx-action': 'restart' }, onclick: function () { send({ restart: true, payOnRestart: pay.input.checked }); } }))));
        return;
      }
      msg.appendChild(notice('error', lbErrText(err)));
    });
  }
  var save = button({ label: t('save'), variant: 'primary', data: { 'lbx-action': 'save-board' }, onclick: function () { send(); } });
  sidePanel(creating ? t('new') : textOf(b.name, b.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'board');
}
