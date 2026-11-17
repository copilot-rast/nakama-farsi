function checkList(items, chosen, name) {
  var box = h('div', { class: 'cl-scope-list', 'data-clan-checks': name });
  var all = items.slice();
  (chosen || []).forEach(function (k) { if (!all.some(function (x) { return x.key === k; })) all.push({ key: k, name: k + ' (' + t('missing') + ')' }); });
  var boxes = all.map(function (x) {
    var c = check((chosen || []).indexOf(x.key) !== -1, (x.name && x.name !== x.key ? x.name + ' · ' : '') + x.key, name);
    c.input.value = x.key;
    box.appendChild(c);
    return c;
  });
  if (!all.length) box.appendChild(h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('none') }));
  box.read = function () { return boxes.filter(function (c) { return c.input.checked; }).map(function (c) { return c.input.value; }); };
  return box;
}

function trophySettingsPage(root) {
  loadRefs('trophies', true).then(function () {
    settingsPage(root, 'trophies/settings', '/trophies', function (cur, defs, d) {
      var notes = h('div', { class: 'flex flex-col gap-2' });
      if (d.prizes === false) notes.appendChild(notice('warning', t('noPrizes')));
      var m = cur.match, a = cur.amounts, p = cur.prizes, s = cur.season, mm = cur.matchmaking, ds = cur.display, ev = cur.events;
      var enabled = check(cur.enabled, t('ts.enabled'), 'enabled');
      var source = select(m.source, opt('src.', ['economy_event', 'game_event']), { name: 'match.source' });
      var event = eventInput(refList('trophies', 'economyEvents'), m.event, 'match.event', 'match_end');
      var won = input(m.wonField, { name: 'match.wonField', ltr: true, mono: true });
      var draw = input(m.drawField, { name: 'match.drawField', ltr: true, mono: true });
      var matchF = input(m.matchField, { name: 'match.matchField', ltr: true, mono: true });
      var opp = input(m.opponentField, { name: 'match.opponentField', ltr: true, mono: true });
      var modeF = input(m.modeField, { name: 'match.modeField', ltr: true, mono: true });
      var modes = tagsInput(m.modes, 'match.modes');
      var skip = input(m.skipField, { name: 'match.skipField', ltr: true, mono: true });
      var start = numInput(a.start, 'amounts.start');
      var win = numInput(a.win, 'amounts.win');
      var loss = numInput(a.loss, 'amounts.loss');
      var drawA = numInput(a.draw, 'amounts.draw');
      var scale = check(a.scale, t('ts.scale'), 'amounts.scale');
      var scalePer = numInput(a.scalePer, 'amounts.scalePer');
      var scaleStep = numInput(a.scaleStep, 'amounts.scaleStep');
      var minWin = numInput(a.minWin, 'amounts.minWin');
      var maxWin = numInput(a.maxWin, 'amounts.maxWin');
      var minLoss = numInput(a.minLoss, 'amounts.minLoss');
      var maxLoss = numInput(a.maxLoss, 'amounts.maxLoss');
      var floors = check(a.floors, t('ts.floors'), 'amounts.floors');
      var oncePer = select(p.oncePer, opt('ts.once.', ['ever', 'season']), { name: 'prizes.oncePer' });
      var rising = check(p.onlyRising, t('ts.onlyRising'), 'prizes.onlyRising');
      var tries = numInput(p.maxTries, 'prizes.maxTries');
      var sOn = check(s.enabled, t('ts.seasonOn'), 'season.enabled');
      var sched = input(s.schedule, { name: 'season.schedule', ltr: true, mono: true, placeholder: '0 0 1 * *' });
      var sMode = select(s.mode, opt('ss.mode.', ['cut', 'league', 'zero']), { name: 'season.mode' });
      var above = numInput(s.above, 'season.above');
      var keep = numInput(s.keepPercent, 'season.keepPercent');
      var cutBox = grid(field(t('ts.above'), above, t('ts.hint.above')), field(t('ts.keep'), keep, t('ts.hint.keep')));
      var mmOn = check(mm.enabled, t('ts.mmOn'), 'matchmaking.enabled');
      var games = checkList(refList('trophies', 'games'), mm.games, 'matchmaking.games');
      var range = numInput(mm.range, 'matchmaking.range');
      var prop = input(mm.property, { name: 'matchmaking.property', ltr: true, mono: true });
      var hist = numInput(ds.historyLimit, 'display.historyLimit');
      var logDays = numInput(ds.logDays, 'display.logDays');
      var showOpp = check(ds.showOpponent, t('ts.showOpponent'), 'display.showOpponent');
      var pub = check(ev.publish, t('ts.publish'), 'events.publish');
      var notify = check(ev.notify, t('ts.notify'), 'events.notify');
      var nCode = numInput(ev.notifyCode, 'events.notifyCode');
      var lCode = numInput(ev.leagueCode, 'events.leagueCode');
      function paint() { cutBox.hidden = sMode.value !== 'cut'; }
      sMode.addEventListener('change', paint);
      paint();
      return {
        el: [notes,
          section(t('ts.general'), h('div', {}, enabled, h('p', { class: 'text-xs text-secondary-500', text: t('ts.hint.enabled') })), 'general'),
          section(t('ts.match'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('ts.hint.match') }),
            grid(field(t('ts.source'), source), field(t('ts.event'), event, t('ts.hint.event'))),
            grid(field(t('ts.won'), won, t('ts.hint.won')), field(t('ts.draw'), draw, t('ts.hint.draw'))),
            grid(field(t('ts.matchField'), matchF, t('ts.hint.matchField')), field(t('ts.opponent'), opp, t('ts.hint.opponent'))),
            grid(field(t('ts.modeField'), modeF, t('ts.hint.modeField')), field(t('ts.modes'), modes, t('ts.hint.modes'))),
            field(t('ts.skip'), skip, t('ts.hint.skip'))), 'match'),
          section(t('ts.amounts'), h('div', {}, grid3(field(t('ts.start'), start, t('ts.hint.start')), field(t('ts.win'), win), field(t('ts.loss'), loss)),
            grid3(field(t('ts.drawA'), drawA, t('ts.hint.drawA')), h('div', {}), h('div', {})), scale, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('ts.hint.scale') }),
            grid(field(t('ts.scalePer'), scalePer), field(t('ts.scaleStep'), scaleStep)),
            grid(field(t('ts.minWin'), minWin), field(t('ts.maxWin'), maxWin)),
            grid(field(t('ts.minLoss'), minLoss), field(t('ts.maxLoss'), maxLoss)), floors, h('p', { class: 'text-xs text-secondary-500 -mt-2', text: t('ts.hint.floors') })), 'amounts'),
          section(t('ts.prizes'), h('div', {}, grid(field(t('ts.oncePer'), oncePer, t('ts.hint.oncePer')), field(t('ts.tries'), tries, t('ts.hint.tries'))), rising), 'prizes'),
          section(t('ts.season'), h('div', {}, sOn, grid(field(t('ts.schedule'), sched, t('ts.hint.schedule')), field(t('ss.modeL'), sMode, t('ts.hint.sMode'))), cutBox), 'season'),
          section(t('ts.mm'), h('div', {}, mmOn, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('ts.hint.mm') }), field(t('ts.games'), games, t('ts.hint.games')),
            grid(field(t('ts.range'), range, t('ts.hint.range')), field(t('ts.property'), prop, t('ts.hint.property')))), 'matchmaking'),
          section(t('ts.display'), h('div', {}, grid(field(t('ts.hist'), hist, t('ts.hint.hist')), field(t('ts.logDays'), logDays, t('ts.hint.logDays'))), showOpp), 'display'),
          section(t('ts.events'), h('div', {}, pub, notify, grid(field(t('ts.nCode'), nCode, t('ts.hint.nCode')), field(t('ts.lCode'), lCode, t('ts.hint.lCode')))), 'events')],
        read: function (c) {
          c.enabled = enabled.input.checked;
          c.match = { source: source.value, event: event.input.value.trim(), wonField: won.value.trim(), drawField: draw.value.trim(), matchField: matchF.value.trim(),
            opponentField: opp.value.trim(), modeField: modeF.value.trim(), modes: modes.read(), skipField: skip.value.trim() };
          c.amounts = { start: intOr(start, 0), win: intOr(win, 0), loss: intOr(loss, 0), draw: intOr(drawA, 0), scale: scale.input.checked, scalePer: intOr(scalePer, 0),
            scaleStep: intOr(scaleStep, 0), minWin: intOr(minWin, 0), maxWin: intOr(maxWin, 0), minLoss: intOr(minLoss, 0), maxLoss: intOr(maxLoss, 0), floors: floors.input.checked };
          c.prizes = { oncePer: oncePer.value, onlyRising: rising.input.checked, maxTries: intOr(tries, 0) };
          c.season = { enabled: sOn.input.checked, schedule: sched.value.trim(), mode: sMode.value, above: intOr(above, 0), keepPercent: intOr(keep, 0) };
          c.matchmaking = { enabled: mmOn.input.checked, games: games.read(), range: intOr(range, 0), property: prop.value.trim() };
          c.display = { historyLimit: intOr(hist, 0), logDays: intOr(logDays, 0), showOpponent: showOpp.input.checked };
          c.events = { publish: pub.input.checked, notify: notify.input.checked, notifyCode: intOr(nCode, 0), leagueCode: intOr(lCode, 0) };
          return null;
        }
      };
    });
  });
}
