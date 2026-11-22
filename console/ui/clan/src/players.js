function trPlayersPage(root) {
  var state = { page: 1, limit: 25, total: 0, q: '', league: '*' };
  var frame = pageFrame('trophies/players', t('desc.trophies/players'), [
    button({ label: t('tp.retry'), data: { 'clan-action': 'retry-prizes' }, onclick: function (e) {
      var b = e.currentTarget;
      busy(b, true);
      api('POST', '/trophies/prizes/retry', {}).then(function (d) { busy(b, false); toast('success', t('tp.retried', { n: fmtNum(d.retried || 0) })); }, function (err) { busy(b, false); toast('error', clErrText(err)); });
    } })]);
  clear(root).appendChild(frame.el);
  var search = searchBox(function (v) { state.q = v; state.page = 1; reload(); });
  var league = select('*', [{ value: '*', label: t('all') }], { name: 'league' });
  league.addEventListener('change', function () { state.league = league.value; state.page = 1; reload(); });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx cl-filters' }, field(t('tp.search'), search, t('tp.hint.search')), field(t('tp.league'), league)));
  frame.content.appendChild(listBox);
  loadLeagues(true).then(function (d) {
    clear(league);
    [{ value: '*', label: t('all') }, { value: '-', label: t('tr.unranked') }].concat((d.items || []).map(function (l) { return { value: l.key, label: textOf(l.name, l.key) }; })).forEach(function (o) {
      league.appendChild(h('option', { value: o.value, text: o.label }));
    });
    league.value = state.league;
  }, function () {});
  function reload() {
    clear(listBox).appendChild(loading());
    var q = qs({ q: state.q, page: state.page, limit: state.limit, league: state.league !== '*' && state.league !== '-' ? state.league : null });
    if (state.league === '-') q += (q ? '&' : '?') + 'league=';
    Promise.all([api('GET', '/trophies/players' + q), loadLeagues()]).then(function (res) {
      var d = res[0];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-clan-list': 'trophy-players' }, table([
        { label: t('f.player') }, { label: t('tp.trophies') }, { label: t('tp.best') }, { label: t('tp.league') }, { label: t('tp.record') }, { label: t('f.updated') }
      ], items.map(function (p) {
        return { id: p.userId, item: p, cells: [
          h('div', { class: 'flex flex-col' }, playerLink(p.userId, p.displayName || p.username), p.username ? mono(p.username) : null),
          h('span', { class: 'font-semibold', text: fmtNum(p.trophies) }),
          h('span', { text: t('tp.bestV', { b: fmtNum(p.best), s: fmtNum(p.seasonBest) }) }),
          p.league ? tag(leagueName(p.league), 'primary') : tag(t('tr.unranked'), 'secondary'),
          h('span', { text: t('tp.wld', { w: fmtNum(p.wins), l: fmtNum(p.losses), d: fmtNum(p.draws) }) }),
          h('span', { class: 'text-xs', text: fmtDate(p.updatedAt) })
        ] };
      }), function (p) { trophyPlayerPanel(p.userId, p.displayName || p.username, reload); })));
      if (!items.length) {
        listBox.appendChild(emptyCard(t('tp.empty')));
        if (isUUID(state.q)) listBox.appendChild(h('div', {}, button({ label: t('tp.open'), variant: 'primary', data: { 'clan-action': 'open-player' }, onclick: function () { trophyPlayerPanel(state.q.toLowerCase(), '', reload); } })));
      }
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}

function trophyPlayerPanel(userId, label, done) {
  var body = h('div', { 'data-clan-player': userId }, loading());
  sidePanel(label || userId.slice(0, 8), body, [button({ label: t('close'), onclick: closePanel })], 'trophy-player');
  function paint(d) {
    clear(body);
    var amount = numInput('', 'trophies', t('tp.amount'));
    var mode = select('set', opt('tp.mode.', ['set', 'add']), { name: 'mode' });
    var msg = h('div', {});
    var apply = button({ label: t('tp.apply'), variant: 'primary', data: { 'clan-action': 'set-trophies' }, onclick: function () {
      clear(msg);
      var n = numberValue(amount);
      if (n == null || isNaN(n)) { msg.appendChild(notice('error', t('code.VALUE_OUT_OF_RANGE') + ' · trophies')); return; }
      busy(apply, true);
      api('PUT', '/trophies/players/' + userId, { trophies: n, mode: mode.value }).then(function (out) {
        toast('success', t('saved')); paint(out); if (done) done();
      }, function (err) { busy(apply, false); fieldError(msg, err); });
    } });
    body.appendChild(h('div', { class: 'mb-4' }, dl([
      [t('f.player'), playerLink(userId, label)],
      [t('tp.trophies'), h('span', { class: 'font-semibold', 'data-clan-value': 'trophies', text: fmtNum(d.trophies) })],
      [t('tp.best'), t('tp.bestV', { b: fmtNum(d.best), s: fmtNum(d.seasonBest) })],
      [t('tp.league'), d.league ? leagueName(d.league) : t('tr.unranked')],
      [t('tp.record'), t('tp.wld', { w: fmtNum(d.wins), l: fmtNum(d.losses), d: fmtNum(d.draws) })]
    ], 'trophy-player')));
    body.appendChild(section(t('tp.change'), h('div', {},
      h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('tp.hint.change') }),
      h('div', { class: 'cl-row' }, field(t('tp.amount'), amount), field(t('tp.modeL'), mode), h('div', {}, apply)), msg), 'set'));
    body.appendChild(h('div', { class: 'h-3' }));
    var log = d.log || [];
    body.appendChild(section(t('tp.log'), log.length ? table([
      { label: t('f.date') }, { label: t('tp.reason') }, { label: t('tp.delta') }, { label: t('tp.after') }, { label: t('tp.league') }, { label: t('tp.opponent') }
    ], log.map(function (x) {
      return { id: String(x.id), cells: [
        h('span', { class: 'text-xs', text: fmtDate(x.createdAt) }),
        tag(has('tp.reason.' + x.reason) ? t('tp.reason.' + x.reason) : x.reason, x.reason === 'win' ? 'success' : x.reason === 'loss' ? 'error' : 'secondary'),
        deltaText(x.delta),
        h('span', { dir: 'ltr', text: digits(fmtNum(x.before) + ' → ' + fmtNum(x.after)) }),
        x.leagueBefore !== x.leagueAfter ? h('span', { text: tr('tp.move', null, { a: leagueName(x.leagueBefore), b: leagueName(x.leagueAfter) }) }) : h('span', { text: leagueName(x.leagueAfter) }),
        x.opponent ? playerLink(x.opponent.userId, x.opponent.displayName || x.opponent.username) : h('span', { class: 'text-secondary-400', text: t('none') })
      ] };
    })) : h('p', { class: 'text-sm text-secondary-500', text: t('tp.noLog') }), 'log'));
    body.appendChild(h('div', { class: 'h-3' }));
    var prizes = d.prizes || [];
    body.appendChild(section(t('tp.prizes'), prizes.length ? table([
      { label: t('tp.league') }, { label: t('tp.season') }, { label: t('lg.prize') }, { label: t('f.state') }, { label: t('f.date') }
    ], prizes.map(function (x) {
      return { id: x.claimKey, cells: [
        h('span', { text: leagueName(x.league) }),
        h('span', { text: x.season ? fmtNum(x.season) : t('tp.ever') }),
        mono(x.prizeKey),
        h('div', { class: 'flex flex-col gap-1' }, tag(t('pay.' + x.status), x.status === 'paid' ? 'success' : x.status === 'failed' ? 'error' : 'warning'),
          x.error ? h('span', { class: 'text-xs text-error-600', dir: 'ltr', text: x.error }) : null),
        h('span', { class: 'text-xs', text: fmtDate(x.paidAt || x.createdAt) })
      ] };
    })) : h('p', { class: 'text-sm text-secondary-500', text: t('tp.noPrizes') }), 'prizes'));
  }
  Promise.all([api('GET', '/trophies/players/' + userId), loadLeagues()]).then(function (res) { paint(res[0]); }, function (err) { clear(body).appendChild(notice('error', clErrText(err))); });
}

function seasonsPage(root) {
  var frame = pageFrame('trophies/seasons', t('desc.trophies/seasons'));
  clear(root).appendChild(frame.el);
  function paint(d) {
    clear(frame.content);
    var reset = button({ label: t('ss.reset'), variant: 'danger', class: 'cl-btn--outline', data: { 'clan-action': 'reset-season' }, onclick: function () {
      confirmBox(t('ss.confirm', { mode: t('ss.mode.' + d.mode) }), t('ss.reset')).then(function (ok) {
        if (!ok) return;
        busy(reset, true);
        api('POST', '/trophies/seasons/reset', { confirm: true }).then(function (out) { toast('success', t('ss.done')); paint(out); }, function (err) { busy(reset, false); toast('error', clErrText(err)); });
      });
    } });
    frame.content.appendChild(card(h('div', { class: 'p-4 flex flex-col sm:flex-row gap-4 justify-between', 'data-clan-section': 'season-info' },
      dl([
        [t('ss.auto'), d.enabled ? tag(t('on'), 'success') : tag(t('off'), 'secondary')],
        [t('ss.schedule'), d.schedule ? mono(d.schedule) : t('none')],
        [t('ss.modeL'), t('ss.mode.' + d.mode)],
        [t('ss.next'), d.enabled && d.nextRun ? h('span', { 'data-clan-value': 'next-run', text: fmtDate(d.nextRun) }) : t('ss.manual')]
      ], 'season'),
      h('div', { class: 'flex flex-col gap-2 items-start' }, reset, h('span', { class: 'text-xs text-secondary-500 max-w-xs', text: t('ss.hint.reset') })))));
    var items = d.items || [];
    frame.content.appendChild(h('div', { 'data-clan-list': 'seasons' }, table([
      { label: t('ss.number') }, { label: t('ss.started') }, { label: t('ss.ends') }, { label: t('ss.ended') }, { label: t('ss.players') }, { label: t('ss.modeL') }, { label: t('ss.actor') }
    ], items.map(function (s) {
      return { id: String(s.number), cells: [
        h('span', { class: 'font-semibold', text: t('ss.n', { n: s.number }) }),
        h('span', { class: 'text-xs', text: fmtDate(s.startedAt) }),
        h('span', { class: 'text-xs', text: s.endsAt ? fmtDate(s.endsAt) : t('none') }),
        s.endedAt ? h('span', { class: 'text-xs', text: fmtDate(s.endedAt) }) : tag(t('ss.current'), 'success'),
        h('span', { text: s.endedAt ? fmtNum(s.playersReset) : t('none') }),
        h('span', { text: s.mode ? t('ss.mode.' + s.mode) : t('none') }),
        h('span', { dir: 'auto', text: s.actor || (s.endedAt ? t('ss.scheduler') : t('none')) })
      ] };
    }))));
    if (!items.length) frame.content.appendChild(emptyCard(t('ss.empty')));
  }
  frame.content.appendChild(loading());
  api('GET', '/trophies/seasons').then(paint, function (err) { clear(frame.content).appendChild(notice('error', clErrText(err))); });
}
