function checkList(items, chosen, name) {
  var box = h('div', { class: 'cu-scope-list', 'data-cup-checks': name });
  var all = items.slice();
  (chosen || []).forEach(function (k) { if (!all.some(function (x) { return x.key === k; })) all.push({ key: k, name: k + ' (' + t('missing') + ')' }); });
  var boxes = all.map(function (x) {
    var c = check((chosen || []).indexOf(x.key) !== -1, x.name && x.name !== x.key ? x.name : x.key, name);
    c.input.value = x.key;
    box.appendChild(c);
    return c;
  });
  if (!all.length) box.appendChild(h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('ru.noneYet') }));
  box.read = function () { return boxes.filter(function (c) { return c.input.checked; }).map(function (c) { return c.input.value; }); };
  return box;
}

function bestOfSelect(v, name) {
  return select(String(v || 1), [1, 3, 5, 7, 9].map(function (n) { return { value: String(n), label: t('ru.bestOfN', { n: n }) }; }), { name: name, label: t('ru.bestOf') });
}

function prizeSelect(v, name) {
  return refSelect(refList('cups', 'prizes'), v, name, t('ru.noPrize'));
}

function rowList(items, make, addLabel, name) {
  var box = h('div', { class: 'cu-rows', 'data-cup-rows': name });
  var rows = [];
  function push(it) {
    var r = make(it || null);
    var del = button({ label: t('remove'), variant: 'danger', data: { 'cup-remove': name }, onclick: function () { r.el.remove(); rows.splice(rows.indexOf(r), 1); } });
    r.el.appendChild(h('div', { class: 'flex items-end' }, del));
    rows.push(r);
    box.appendChild(r.el);
  }
  (items || []).forEach(push);
  var wrap = h('div', { class: 'mb-4' }, box, h('div', { class: 'mt-2' }, button({ label: addLabel, data: { 'cup-add': name }, onclick: function () { push(null); } })));
  wrap.read = function () { return rows.map(function (r) { return r.read(); }); };
  return wrap;
}

function roundOptions() {
  var out = [{ value: 'final', label: t('rk.final') }, { value: 'semi', label: t('rk.semi') }, { value: 'third', label: t('rk.third') }];
  for (var i = 1; i <= 7; i++) out.push({ value: String(i), label: t('ru.roundN', { n: i }) });
  return out;
}

function rulesForm(src, opts) {
  opts = opts || {};
  var r = JSON.parse(JSON.stringify(src || {}));
  r.timing = r.timing || {};
  r.draw = r.draw || {};
  r.entry = r.entry || {};
  r.entry.fee = r.entry.fee || {};
  r.entry.requirements = r.entry.requirements || {};
  r.prizes = r.prizes || {};
  r.prizes.pool = r.prizes.pool || {};
  r.prizes.pool.shares = r.prizes.pool.shares || {};
  var tm = r.timing, dr = r.draw, en = r.entry, rq = en.requirements, pz = r.prizes, pool = pz.pool, sh = pool.shares;
  var nums = [];
  function num(value, name, label) {
    var el = numInput(value, name, label);
    nums.push({ el: el, label: label });
    return el;
  }
  var gameOpts = refList('cups', 'games').map(function (g) { return { value: g.key, label: g.key ? (g.name && g.name !== g.key ? g.name + ' · ' + g.key : g.key) : t('ru.external') }; });
  if (r.game && !gameOpts.some(function (o) { return o.value === r.game; })) gameOpts.push({ value: r.game, label: r.game + ' (' + t('missing') + ')' });
  if (!gameOpts.length) gameOpts.push({ value: '', label: t('ru.external') });
  var game = select(r.game || '', gameOpts, { name: 'rules.game', label: t('ru.game') });
  var rated = check(r.rated, t('ru.rated'), 'rules.rated');
  var gameOptions = jsonInput(r.gameOptions, 'rules.gameOptions');
  var capacity = num(r.capacity, 'rules.capacity', t('ru.capacity'));
  var minPlayers = num(r.minPlayers, 'rules.minPlayers', t('ru.minPlayers'));
  var underfill = select(r.underfill || 'shrink', opt('ru.under.', ['shrink', 'cancel']), { name: 'rules.underfill', label: t('ru.underfill') });
  var seeding = select(r.seeding || 'random', opt('ru.seed.', ['random', 'trophies']), { name: 'rules.seeding', label: t('ru.seeding') });
  var third = check(r.thirdPlace, t('ru.thirdPlace'), 'rules.thirdPlace');
  var bestOf = bestOfSelect(r.bestOf, 'rules.bestOf');
  var roundBestOf = rowList(r.roundBestOf, function (it) {
    var round = select(it ? it.round : 'final', roundOptions(), { name: 'roundBestOf.round', label: t('ru.round') });
    var bo = bestOfSelect(it ? it.bestOf : 3, 'roundBestOf.bestOf');
    return { el: h('div', { class: 'cu-row' }, field(t('ru.round'), round), field(t('ru.bestOf'), bo)), read: function () { return { round: round.value, bestOf: Number(bo.value) }; } };
  }, t('ru.addRoundBestOf'), 'roundBestOf');
  var mode = select(tm.mode || 'window', opt('mode.', ['window', 'live']), { name: 'rules.timing.mode', label: t('ru.mode') });
  var windowMinutes = num(tm.windowMinutes, 'rules.timing.windowMinutes', t('ru.windowMinutes'));
  var readySeconds = num(tm.readySeconds, 'rules.timing.readySeconds', t('ru.readySeconds'));
  var checkInMinutes = num(tm.checkInMinutes, 'rules.timing.checkInMinutes', t('ru.checkInMinutes'));
  var roundMinutes = num(tm.roundMinutes, 'rules.timing.roundMinutes', t('ru.roundMinutes'));
  var joinSeconds = num(tm.joinSeconds, 'rules.timing.joinSeconds', t('ru.joinSeconds'));
  var gameMinutes = num(tm.gameMinutes, 'rules.timing.gameMinutes', t('ru.gameMinutes'));
  var noShow = select(tm.noShow || 'both_out', opt('ru.ns.', ['both_out', 'higher_seed', 'random']), { name: 'rules.timing.noShow', label: t('ru.noShow') });
  var unplayed = select(tm.unplayed || 'higher_seed', opt('ru.ns.', ['both_out', 'higher_seed', 'random']), { name: 'rules.timing.unplayed', label: t('ru.unplayed') });
  var windowBox = grid(field(t('ru.windowMinutes'), windowMinutes, t('ru.hint.windowMinutes')), field(t('ru.readySeconds'), readySeconds, t('ru.hint.readySeconds')));
  var liveBox = grid(field(t('ru.checkInMinutes'), checkInMinutes, t('ru.hint.checkInMinutes')), field(t('ru.roundMinutes'), roundMinutes, t('ru.hint.roundMinutes')));
  var modeHint = h('p', { class: 'text-xs text-secondary-500 mb-3', 'data-cup-mode-hint': '' });
  function paintMode() {
    windowBox.hidden = mode.value !== 'window';
    liveBox.hidden = mode.value !== 'live';
    modeHint.textContent = t('ru.hint.mode.' + mode.value);
  }
  mode.addEventListener('change', paintMode);
  paintMode();
  var drawRule = select(dr.rule || 'replay', opt('ru.draw.', ['replay', 'higher_seed', 'random']), { name: 'rules.draw.rule', label: t('ru.drawRule') });
  var maxReplays = num(dr.maxReplays, 'rules.draw.maxReplays', t('ru.maxReplays'));
  var fallback = select(dr.fallback || 'higher_seed', opt('ru.draw.', ['higher_seed', 'random']), { name: 'rules.draw.fallback', label: t('ru.fallback') });
  var gameDecides = check(dr.gameDecides, t('ru.gameDecides'), 'rules.draw.gameDecides');
  var replayBox = grid(field(t('ru.maxReplays'), maxReplays), field(t('ru.fallback'), fallback));
  function paintDraw() { replayBox.hidden = drawRule.value !== 'replay'; }
  drawRule.addEventListener('change', paintDraw);
  paintDraw();
  var currencies = refList('cups', 'currencies');
  var feeCur = refSelect(currencies, en.fee.currencyKey || '', 'rules.entry.fee.currencyKey', t('ru.noCurrency'));
  var feeAmount = num(en.fee.amount || 0, 'rules.entry.fee.amount', t('ru.feeAmount'));
  var refund = num(en.withdrawRefundPercent, 'rules.entry.withdrawRefundPercent', t('ru.withdrawRefund'));
  var minTrophies = num(rq.minTrophies || 0, 'rules.entry.requirements.minTrophies', t('ru.minTrophies'));
  var maxTrophies = num(rq.maxTrophies || 0, 'rules.entry.requirements.maxTrophies', t('ru.maxTrophies'));
  var leagues = checkList(refList('cups', 'leagues'), rq.leagues, 'rules.entry.requirements.leagues');
  var tags = refList('cups', 'tags');
  var tagIds = checkList(tags, rq.tagIds, 'rules.entry.requirements.tagIds');
  var tagsMode = select(rq.tagsMode || 'any', opt('ru.tagsMode.', ['any', 'all']), { name: 'rules.entry.requirements.tagsMode', label: t('ru.tagsMode') });
  var excludeTagIds = checkList(tags, rq.excludeTagIds, 'rules.entry.requirements.excludeTagIds');
  var pChampion = prizeSelect(pz.champion, 'rules.prizes.champion');
  var pRunner = prizeSelect(pz.runnerUp, 'rules.prizes.runnerUp');
  var pSemi = prizeSelect(pz.semi, 'rules.prizes.semi');
  var pThird = prizeSelect(pz.third, 'rules.prizes.third');
  var pFourth = prizeSelect(pz.fourth, 'rules.prizes.fourth');
  var reached = rowList(pz.reached, function (it) {
    var round = numInput(it ? it.round : 1, 'reached.round', t('ru.round'));
    var prize = prizeSelect(it ? it.prizeKey : '', 'reached.prizeKey');
    return { el: h('div', { class: 'cu-row' }, field(t('ru.reachedRound'), round), field(t('ru.prize'), prize)), read: function () { return { round: intOr(round, 0), prizeKey: prize.value }; } };
  }, t('ru.addReached'), 'reached');
  var poolPercent = num(pool.percent || 0, 'rules.prizes.pool.percent', t('ru.poolPercent'));
  var poolBase = num(pool.base || 0, 'rules.prizes.pool.base', t('ru.poolBase'));
  var shChampion = num(sh.champion || 0, 'rules.prizes.pool.shares.champion', t('pl.champion'));
  var shRunner = num(sh.runnerUp || 0, 'rules.prizes.pool.shares.runnerUp', t('pl.runnerUp'));
  var shSemi = num(sh.semi || 0, 'rules.prizes.pool.shares.semi', t('pl.semi'));
  var shThird = num(sh.third || 0, 'rules.prizes.pool.shares.third', t('pl.third'));
  var shFourth = num(sh.fourth || 0, 'rules.prizes.pool.shares.fourth', t('pl.fourth'));
  var thirdPrizes = grid(field(t('pl.third'), pThird), field(t('pl.fourth'), pFourth));
  var semiPrize = field(t('pl.semi'), pSemi, t('ru.hint.semi'));
  function paintThird() { thirdPrizes.hidden = !third.input.checked; semiPrize.hidden = third.input.checked; }
  third.input.addEventListener('change', paintThird);
  paintThird();
  var locked = !!opts.locked;
  var el = [
    section(t('ru.s.game'), [
      field(t('ru.game'), game, t('ru.hint.game')), rated,
      field(t('ru.gameOptions'), gameOptions, t('ru.hint.gameOptions'))
    ], 'game'),
    section(t('ru.s.players'), [
      grid(field(t('ru.capacity'), capacity, t('ru.hint.capacity')), field(t('ru.minPlayers'), minPlayers, t('ru.hint.minPlayers'))),
      grid(field(t('ru.underfill'), underfill, t('ru.hint.underfill')), field(t('ru.seeding'), seeding, t('ru.hint.seeding'))),
      third
    ], 'players'),
    section(t('ru.s.bestOf'), [
      field(t('ru.bestOf'), bestOf, t('ru.hint.bestOf')),
      group(t('ru.roundBestOf'), roundBestOf, t('ru.hint.roundBestOf'))
    ], 'bestOf'),
    section(t('ru.s.timing'), [
      field(t('ru.mode'), mode), modeHint, windowBox, liveBox,
      grid(field(t('ru.joinSeconds'), joinSeconds, t('ru.hint.joinSeconds')), field(t('ru.gameMinutes'), gameMinutes, t('ru.hint.gameMinutes'))),
      grid(field(t('ru.noShow'), noShow, t('ru.hint.noShow')), field(t('ru.unplayed'), unplayed, t('ru.hint.unplayed')))
    ], 'timing'),
    section(t('ru.s.draw'), [field(t('ru.drawRule'), drawRule), replayBox, gameDecides], 'draw'),
    section(t('ru.s.entry'), [
      grid(field(t('ru.feeCurrency'), feeCur, t('ru.hint.fee')), field(t('ru.feeAmount'), feeAmount)),
      field(t('ru.withdrawRefund'), refund, t('ru.hint.withdrawRefund')),
      grid(field(t('ru.minTrophies'), minTrophies, t('hint.zero')), field(t('ru.maxTrophies'), maxTrophies, t('hint.zero'))),
      group(t('ru.leagues'), leagues, t('ru.hint.leagues')),
      group(t('ru.tagIds'), tagIds), field(t('ru.tagsMode'), tagsMode),
      group(t('ru.excludeTagIds'), excludeTagIds)
    ], 'entry'),
    section(t('ru.s.prizes'), [
      grid(field(t('pl.champion'), pChampion), field(t('pl.runnerUp'), pRunner)),
      semiPrize, thirdPrizes,
      group(t('ru.reached'), reached, t('ru.hint.reached')),
      h('h3', { class: 'font-medium text-sm mt-2 mb-2', text: t('ru.pool') }),
      h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('ru.hint.pool') }),
      grid(field(t('ru.poolPercent'), poolPercent), field(t('ru.poolBase'), poolBase)),
      grid3(field(t('pl.champion'), shChampion), field(t('pl.runnerUp'), shRunner), field(t('pl.semi'), shSemi)),
      grid(field(t('pl.third'), shThird), field(t('pl.fourth'), shFourth))
    ], 'prizes')
  ];
  if (locked) el.forEach(function (s) { Array.prototype.forEach.call(s.querySelectorAll('input,select,textarea,button'), function (x) { x.disabled = true; }); });
  function read() {
    for (var i = 0; i < nums.length; i++) {
      var n = numberValue(nums[i].el);
      if (n === null || isNaN(n)) return { error: t('ru.badNumber', { field: nums[i].label }) };
    }
    var go = gameOptions.read();
    if (go === undefined) return { error: t('ru.badJson') };
    var rb = roundBestOf.read();
    var rc = reached.read();
    for (var j = 0; j < rc.length; j++) if (!rc[j].round || !rc[j].prizeKey) return { error: t('ru.badReached') };
    var out = JSON.parse(JSON.stringify(src || {}));
    out.game = game.value;
    out.rated = rated.input.checked;
    out.gameOptions = go;
    out.capacity = numberValue(capacity);
    out.minPlayers = numberValue(minPlayers);
    out.underfill = underfill.value;
    out.seeding = seeding.value;
    out.thirdPlace = third.input.checked;
    out.bestOf = Number(bestOf.value);
    out.roundBestOf = rb;
    out.timing = { mode: mode.value, windowMinutes: numberValue(windowMinutes), roundMinutes: numberValue(roundMinutes), checkInMinutes: numberValue(checkInMinutes),
      noShow: noShow.value, unplayed: unplayed.value, readySeconds: numberValue(readySeconds), joinSeconds: numberValue(joinSeconds), gameMinutes: numberValue(gameMinutes) };
    out.draw = { rule: drawRule.value, maxReplays: numberValue(maxReplays), fallback: fallback.value, gameDecides: gameDecides.input.checked };
    out.entry = { fee: { currencyKey: feeCur.value, amount: feeCur.value ? numberValue(feeAmount) : 0 }, withdrawRefundPercent: numberValue(refund),
      requirements: { minTrophies: numberValue(minTrophies), maxTrophies: numberValue(maxTrophies), leagues: leagues.read(), tagIds: tagIds.read(), tagsMode: tagsMode.value, excludeTagIds: excludeTagIds.read() } };
    out.prizes = { champion: pChampion.value, runnerUp: pRunner.value, semi: third.input.checked ? '' : pSemi.value, third: third.input.checked ? pThird.value : '', fourth: third.input.checked ? pFourth.value : '',
      reached: rc, pool: { percent: numberValue(poolPercent), base: numberValue(poolBase), shares: { champion: numberValue(shChampion), runnerUp: numberValue(shRunner), semi: numberValue(shSemi), third: numberValue(shThird), fourth: numberValue(shFourth) } } };
    return { rules: out };
  }
  return { el: el, read: read };
}

function rulesSummary(r) {
  if (!r) return '';
  var parts = [t('mode.' + (r.timing && r.timing.mode || 'window')), t('ru.bestOfN', { n: r.bestOf || 1 }), t('ru.capN', { n: r.capacity })];
  if (r.entry && r.entry.fee && r.entry.fee.amount > 0) parts.push(tr('ru.feeV', null, { amount: fmtNum(r.entry.fee.amount), currency: r.entry.fee.currencyKey }));
  else parts.push(t('ru.free'));
  return parts.join(' · ');
}
