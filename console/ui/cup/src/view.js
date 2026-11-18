var CU_MATCH_COLOR = { pending: 'secondary', waiting: 'primary', open: 'primary', playing: 'warning', finished: 'success', bye: 'secondary', void: 'secondary' };
var CU_PLAYER_COLOR = { registered: 'primary', playing: 'warning', eliminated: 'secondary', finished: 'success', disqualified: 'error', withdrawn: 'secondary', removed: 'secondary' };
var CU_POLL_MS = 5000;
var cuPoll = null;

function matchTag(s) {
  return tag(t('ms.' + s), CU_MATCH_COLOR[s] || 'secondary');
}

function playerTag(s) {
  return tag(t('ps.' + s), CU_PLAYER_COLOR[s] || 'secondary');
}

function reasonText(r) {
  if (!r) return '';
  var parts = String(r).split(':');
  var main = has('why.' + parts[0]) ? t('why.' + parts[0]) : parts[0];
  if (parts[1]) main += ' (' + (has('ru.ns.' + parts[1]) ? t('ru.ns.' + parts[1]) : parts[1]) + ')';
  return main;
}

function payoutKind(k) {
  var m = /^reached_(\d+)$/.exec(k || '');
  if (m) return t('po.reached', { n: Number(m[1]) });
  return has('po.' + k) ? t('po.' + k) : k;
}

function sideName(s) {
  if (!s) return '';
  return s.displayName || s.username || (s.userId || '').slice(0, 8);
}

function isOpen(c) {
  return c.status === 'draft' || c.status === 'upcoming' || c.status === 'signup';
}

function stopPoll() {
  if (cuPoll) { clearInterval(cuPoll); cuPoll = null; }
}

function resolveUser(text) {
  text = String(text || '').trim();
  if (!text) return Promise.reject(new CuError(400, 'PLAYER_NOT_FOUND', '', 'userId'));
  if (isUUID(text)) return Promise.resolve(text);
  return api('GET', '/v2/console/account' + qs({ filter: text })).then(function (d) {
    var list = (d && d.users) || [];
    var hit = list.filter(function (u) { var x = u.user || u; return (x.username || '').toLowerCase() === text.toLowerCase(); })[0] || (list.length === 1 ? list[0] : null);
    if (!hit) throw new CuError(404, 'PLAYER_NOT_FOUND', '', 'userId');
    return (hit.user || hit).id;
  });
}

function cupViewPage(root) {
  stopPoll();
  var hm = /^#\/cups\/list\/([^/?]+)/.exec(location.hash);
  var id = hm ? decodeURIComponent(hm[1]) : '';
  var frame = pageFrame('cups/view', '', null);
  clear(root).appendChild(frame.el);
  var head = frame.el.querySelector('h1');
  var descEl = frame.el.querySelector('h1 + p');
  var actions = frame.el.querySelector('.flex.flex-wrap.gap-2');
  var back = h('a', { href: '#/cups/list', class: 'eco-link text-sm mb-2 inline-block', 'data-cup-back': '', text: t('cu.back') });
  frame.el.insertBefore(back, frame.el.firstChild);
  if (!id) { frame.content.appendChild(notice('error', t('code.CUP_NOT_FOUND'))); return; }
  var live = h('p', { class: 'text-xs text-secondary-500', 'data-cup-live': '' });
  var summary = h('div', { 'data-cup-summary': '' });
  var bracketBox = h('div', { 'data-cup-bracket-box': '' });
  var playersBox = h('div', { 'data-cup-players-box': '' });
  var payoutsBox = h('div', { 'data-cup-payouts-box': '' });
  frame.content.appendChild(h('div', { class: 'flex flex-col gap-4' }, live, summary, bracketBox, playersBox, payoutsBox));
  frame.content.appendChild(loading());
  var data = null;
  function busyUi() {
    return !!cuPanel || !!document.querySelector('[data-cup=confirm]');
  }
  function refresh(quiet) {
    return api('GET', '/cups/list/' + encodeURIComponent(id)).then(function (d) {
      if (!root.isConnected) { stopPoll(); return; }
      data = d;
      paint();
    }, function (err) {
      if (quiet) return;
      var l = frame.content.querySelector('p.text-secondary-500:last-child');
      if (l && l.textContent === t('loading')) l.remove();
      summary.appendChild(notice('error', cuErrText(err)));
    });
  }
  function act(path, body, confirmText, okLabel, doneText) {
    return (confirmText ? confirmBox(confirmText, okLabel) : Promise.resolve(true)).then(function (ok) {
      if (!ok) return false;
      return api(path.method || 'POST', '/cups/list/' + encodeURIComponent(id) + path.path, body).then(function (out) {
        toast('success', doneText || t('done'));
        if (out && out.cup) { data = out; paint(); } else refresh(true);
        return true;
      }, function (err) { toast('error', cuErrText(err)); return false; });
    });
  }
  function paint() {
    var c = data.cup;
    var l = frame.content.querySelector(':scope > p');
    if (l) l.remove();
    head.textContent = cupName(c);
    head.setAttribute('dir', 'auto');
    descEl.textContent = rulesSummary(c.rules);
    live.textContent = (c.status === 'running' || c.status === 'signup' || c.status === 'upcoming' ? t('cu.liveOn') + ' · ' : '') + t('cu.refreshed', { at: fmtDate(data.serverTime || new Date().toISOString()) });
    clear(actions);
    paintActions(c);
    clear(summary).appendChild(summaryCard(c));
    clear(bracketBox).appendChild(bracketCard(c, data.bracket));
    clear(playersBox).appendChild(playersCard(c, data.players || []));
    clear(payoutsBox).appendChild(payoutsCard(c, data.payouts || []));
  }
  function paintActions(c) {
    var open = isOpen(c);
    actions.appendChild(button({ label: t('cu.edit'), data: { 'cup-action': 'edit-cup' }, onclick: function () { cupEditor(c, function (nc, out) { data = out; paint(); }); } }));
    if (c.status === 'draft') actions.appendChild(button({ label: t('cu.publishNow'), variant: 'primary', data: { 'cup-action': 'publish' }, onclick: function () { act({ path: '/publish' }, {}, null, null, t('cu.published')); } }));
    if (open) actions.appendChild(button({ label: t('cu.startNow'), variant: 'primary', data: { 'cup-action': 'start' }, onclick: function () { act({ path: '/start' }, {}, t('cu.startConfirm', { n: c.players }), t('cu.startNow'), t('cu.started')); } }));
    if (open) actions.appendChild(button({ label: t('cu.addPlayer'), data: { 'cup-action': 'add-player' }, onclick: function () { addPlayerPanel(); } }));
    if (c.status !== 'finished' && c.status !== 'cancelled') actions.appendChild(button({ label: t('cu.cancel'), variant: 'danger', class: 'cu-btn--outline', data: { 'cup-action': 'cancel' }, onclick: function () { cancelPanel(c); } }));
    if (c.status === 'draft' || c.status === 'cancelled') actions.appendChild(button({ label: t('delete'), variant: 'danger', data: { 'cup-action': 'delete' }, onclick: function () {
      confirmBox(t('cu.deleteConfirm'), t('delete')).then(function (ok) {
        if (!ok) return;
        api('DELETE', '/cups/list/' + encodeURIComponent(id)).then(function () { toast('success', t('deleted')); location.hash = '#/cups/list'; }, function (err) { toast('error', cuErrText(err)); });
      });
    } }));
  }
  function summaryCard(c) {
    var r = c.rules || {};
    var champ = c.championId ? (data.players || []).filter(function (p) { return p.userId === c.championId; })[0] : null;
    var pool = c.pool || {};
    return card(h('div', { class: 'p-4 grid gap-4 md:grid-cols-2' },
      dl([
        [t('f.state'), h('span', { 'data-cup-value': 'status' }, statusTag(c.status))],
        [t('ru.mode'), h('span', { 'data-cup-value': 'mode', text: t('mode.' + ((r.timing && r.timing.mode) || 'window')) })],
        [t('ru.game'), r.game ? mono(r.game) : t('ru.external')],
        [t('cu.players'), h('span', { 'data-cup-value': 'players', text: t('cu.entrantsV', { n: fmtNum(c.status === 'running' || c.status === 'finished' ? c.players : (data.players || []).filter(function (p) { return p.status === 'registered'; }).length), cap: fmtNum(r.capacity) }) })],
        c.size ? [t('cu.size'), h('span', { 'data-cup-value': 'size', text: t('cu.sizeV', { n: c.size, rounds: c.rounds }) })] : null,
        c.status === 'running' ? [t('cu.round'), h('span', { 'data-cup-value': 'round', text: t('cu.roundV', { n: c.currentRound, of: c.rounds }) })] : null,
        champ ? [t('pl.champion'), h('span', { 'data-cup-value': 'champion' }, playerLink(champ.userId, champ.displayName || champ.username))] : null,
        c.cancelReason ? [t('cu.cancelReason'), h('span', { 'data-cup-value': 'cancel-reason', dir: 'auto', text: reasonText(c.cancelReason) })] : null
      ], 'cup'),
      dl([
        [t('cu.signupAt'), fmtDate(c.signupAt)],
        [t('cu.startsAt'), fmtDate(c.startsAt)],
        c.startedAt ? [t('cu.startedAt'), fmtDate(c.startedAt)] : null,
        c.finishedAt ? [t('cu.finishedAt'), fmtDate(c.finishedAt)] : null,
        [t('ru.feeAmount'), r.entry && r.entry.fee && r.entry.fee.amount > 0 ? h('span', {}, fmtNum(r.entry.fee.amount) + ' ', mono(r.entry.fee.currencyKey)) : t('ru.free')],
        pool.amount ? [t('ru.pool'), h('span', { 'data-cup-value': 'pool' }, fmtNum(pool.amount) + ' ', mono(pool.currencyKey || ''))] : null,
        c.templateKey ? [t('tp.template'), mono(c.templateKey)] : null,
        [t('f.id'), mono(c.id)]
      ], 'times')));
  }
  function bracketCard(c, br) {
    if (!br || !br.rounds || !br.rounds.length) return section(t('cu.bracket'), h('p', { class: 'text-sm text-secondary-500', 'data-cup-no-bracket': '', text: c.status === 'cancelled' ? t('cu.noBracketCancelled') : t('cu.noBracket') }), 'bracket');
    var cols = br.rounds.map(function (rd) {
      return h('div', { class: 'cu-round', 'data-cup-round': String(rd.round) },
        h('div', { class: 'cu-round-head' },
          h('span', { class: 'font-medium', text: roundName(rd.key) }),
          h('span', { class: 'text-xs text-secondary-500', text: t('ru.bestOfN', { n: rd.bestOf }) + ' · ' + t('rs.' + rd.status) })),
        h('div', { class: 'cu-round-body' }, rd.matches.map(function (m) { return matchCard(c, m); })));
    });
    var extra = br.third ? h('div', { class: 'mt-4' }, h('h3', { class: 'font-medium text-sm mb-2', text: t('rk.third') }), h('div', { class: 'cu-third' }, matchCard(c, br.third))) : null;
    return section(t('cu.bracket'), [h('div', { class: 'cu-bracket', 'data-cup-bracket': '', tabindex: '0', role: 'region', 'aria-label': t('cu.bracket') }, cols), extra], 'bracket');
  }
  function sideRow(m, s, wins) {
    if (!s) return h('div', { class: 'cu-side cu-side--empty' }, h('span', { class: 'text-secondary-400', text: m.status === 'bye' ? t('ms.bye') : t('cu.tbd') }));
    var won = m.winnerId && m.winnerId === s.userId;
    var lost = m.winnerId && m.winnerId !== s.userId;
    return h('div', { class: 'cu-side' + (won ? ' cu-side--won' : '') + (lost ? ' cu-side--lost' : ''), 'data-cup-side': s.userId },
      h('span', { class: 'cu-seed tabular-nums', text: s.seed ? digits(String(s.seed)) : '' }),
      h('span', { class: 'cu-side-name', dir: 'auto', text: sideName(s) }),
      s.ready && !done(m) ? h('span', { class: 'cu-ready', title: t('cu.ready'), 'aria-label': t('cu.ready'), text: '●' }) : null,
      h('span', { class: 'cu-wins tabular-nums', text: m.status === 'bye' || (!(m.games || []).length && m.status !== 'playing') ? '' : digits(String(wins || 0)) }));
  }
  function done(m) { return m.status === 'finished' || m.status === 'bye' || m.status === 'void'; }
  function matchCard(c, m) {
    var btn = h('button', { type: 'button', class: 'cu-match', 'data-cup-match': m.id, 'data-cup-status': m.status, 'aria-label': t('cu.matchOf', { a: sideName(m.a) || t('cu.tbd'), b: sideName(m.b) || t('cu.tbd') }) },
      sideRow(m, m.a, m.a && m.a.wins), sideRow(m, m.b, m.b && m.b.wins),
      h('div', { class: 'cu-match-foot' }, matchTag(m.status), m.reason && m.status !== 'bye' ? h('span', { class: 'text-xs text-secondary-500 truncate', text: reasonText(m.reason) }) : null));
    btn.addEventListener('click', function () { matchPanel(c, m); });
    return btn;
  }
  function matchPanel(c, m) {
    var body = h('div', { 'data-cup-match-panel': m.id });
    var foot = h('div', { class: 'flex flex-wrap gap-2 justify-end w-full' });
    sidePanel(t('cu.match') + ' · ' + (m.kind === 'third' ? t('rk.third') : roundName(m.roundKey)), body, [foot], 'match');
    body.appendChild(dl([
      [t('f.state'), matchTag(m.status)],
      [t('ru.bestOf'), t('ru.bestOfN', { n: m.bestOf })],
      [t('cu.sideA'), m.a ? h('span', {}, playerLink(m.a.userId, sideName(m.a)), ' · ', t('cu.winsN', { n: m.a.wins })) : t('cu.tbd')],
      [t('cu.sideB'), m.b ? h('span', {}, playerLink(m.b.userId, sideName(m.b)), ' · ', t('cu.winsN', { n: m.b.wins })) : (m.status === 'bye' ? t('ms.bye') : t('cu.tbd'))],
      m.winnerId ? [t('cu.winner'), h('span', { 'data-cup-value': 'winner', dir: 'auto', text: sideName(m.winnerId === (m.a && m.a.userId) ? m.a : m.b) })] : null,
      m.reason ? [t('cu.reason'), reasonText(m.reason)] : null,
      m.draws ? [t('cu.draws'), fmtNum(m.draws)] : null,
      m.opensAt ? [t('cu.opensAt'), fmtDate(m.opensAt)] : null,
      m.startsAt ? [t('cu.matchStarts'), fmtDate(m.startsAt)] : null,
      m.closesAt ? [t('cu.closesAt'), fmtDate(m.closesAt)] : null,
      m.finishedAt ? [t('cu.finishedAt'), fmtDate(m.finishedAt)] : null,
      m.gameMatchId ? [t('cu.gameMatch'), mono(m.gameMatchId)] : null,
      [t('f.id'), mono(m.id)]
    ], 'match'));
    var games = m.games || [];
    body.appendChild(h('div', { class: 'h-3' }));
    body.appendChild(section(t('cu.games'), games.length ? h('div', { class: 'cu-rows' }, games.map(function (g) {
      var w = g.winnerId ? (m.a && m.a.userId === g.winnerId ? sideName(m.a) : m.b && m.b.userId === g.winnerId ? sideName(m.b) : g.winnerId.slice(0, 8)) : '';
      return h('div', { class: 'cu-game', 'data-cup-game': String(g.no) },
        h('span', { class: 'font-medium', text: t('cu.gameN', { n: g.no }) }), matchTag(g.status === 'launching' ? 'waiting' : g.status),
        h('span', { class: 'text-sm', dir: 'auto', text: g.draw ? t('cu.draw') : w ? t('cu.wonBy', { name: w }) : reasonText(g.reason) }),
        h('span', { class: 'text-xs text-secondary-500', text: fmtDate(g.finishedAt || g.createdAt) }));
    })) : h('p', { class: 'text-sm text-secondary-500', text: t('cu.noGames') }), 'games'));
    var canForce = c.status === 'running' && !done(m) && m.a && m.b;
    if (canForce) {
      body.appendChild(h('div', { class: 'h-3' }));
      body.appendChild(section(t('cu.force'), [h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('cu.hint.force') }),
        h('div', { class: 'flex flex-wrap gap-2' },
          [m.a, m.b].map(function (s) {
            return button({ label: t('cu.winIs', { name: sideName(s) }), variant: 'primary', data: { 'cup-force': s.userId }, onclick: function () {
              act({ path: '/matches/' + m.id + '/result' }, { winnerId: s.userId }, t('cu.forceConfirm', { name: sideName(s) }), t('cu.forceOk'), t('cu.forced')).then(function (ok) { if (ok) closePanel(); });
            } });
          }),
          button({ label: t('cu.voidMatch'), variant: 'danger', data: { 'cup-force': 'void' }, onclick: function () {
            act({ path: '/matches/' + m.id + '/result' }, { void: true }, t('cu.voidConfirm'), t('cu.voidMatch'), t('cu.voided')).then(function (ok) { if (ok) closePanel(); });
          } }))], 'force'));
    }
    foot.appendChild(button({ label: t('close'), onclick: closePanel }));
  }
  function playersCard(c, players) {
    var open = isOpen(c);
    var running = c.status === 'running';
    var list = players.filter(function (p) { return p.status !== 'removed' && p.status !== 'withdrawn'; });
    var gone = players.length - list.length;
    var rows = list.slice().sort(function (a, b) { return (a.seed || 999) - (b.seed || 999) || String(a.joinedAt).localeCompare(String(b.joinedAt)); });
    var tbl = rows.length ? table([
      { label: t('cu.seed') }, { label: t('f.player') }, { label: t('f.state') }, { label: t('cu.trophies') }, { label: t('cu.place') }, { label: t('cu.fee') }, { label: t('cu.actions') }
    ], rows.map(function (p) {
      var acts = [];
      var activeP = p.status === 'registered' || p.status === 'playing';
      if (open && p.status === 'registered') acts.push(button({ label: t('remove'), variant: 'danger', data: { 'cup-player-action': 'remove' }, onclick: function () {
        act({ method: 'DELETE', path: '/players/' + p.userId }, undefined, t('cu.removeConfirm', { name: p.username }), t('remove'), t('cu.removed'));
      } }));
      if ((open || running) && activeP) acts.push(button({ label: t('cu.swap'), data: { 'cup-player-action': 'swap' }, onclick: function () { swapPanel(p); } }));
      if (running && activeP) acts.push(button({ label: t('cu.disqualify'), variant: 'danger', data: { 'cup-player-action': 'disqualify' }, onclick: function () { disqualifyPanel(p); } }));
      return { id: p.userId, item: p, cells: [
        h('span', { class: 'tabular-nums', text: p.seed ? digits(String(p.seed)) : t('none') }),
        playerLink(p.userId, p.displayName || p.username),
        playerTag(p.status),
        h('span', { class: 'tabular-nums', text: fmtNum(p.trophies || 0) }),
        h('span', { text: p.place ? placeText(p.place) : p.reached ? t('cu.reachedV', { n: p.reached }) : t('none') }),
        h('span', { class: 'text-xs', text: feeText(p) }),
        h('div', { class: 'flex flex-wrap gap-1' }, acts.length ? acts : t('none'))
      ] };
    })) : emptyCard(open ? t('cu.noPlayersYet') : t('cu.noPlayers'));
    return section(t('cu.playersT', { n: fmtNum(list.length) }), [tbl, gone ? h('p', { class: 'text-xs text-secondary-500 mt-2', text: t('cu.goneN', { n: gone }) }) : null], 'players');
  }
  function placeText(n) {
    return n === 1 ? t('pl.champion') : n === 2 ? t('pl.runnerUp') : n === 3 ? t('pl.third') : n === 4 ? t('pl.fourth') : t('cu.placeN', { n: n });
  }
  function feeText(p) {
    var keys = Object.keys(p.fee || {});
    var amount = keys.map(function (k) { return fmtNum(p.fee[k]) + ' ' + k; }).join('، ');
    return (amount ? amount + ' · ' : '') + t('fs.' + (p.feeState || 'none'));
  }
  function payoutsCard(c, payouts) {
    if (!payouts.length) return h('div', {});
    var failed = payouts.some(function (p) { return p.status === 'failed'; });
    var retry = failed ? button({ label: t('cu.retry'), data: { 'cup-action': 'retry-payouts' }, onclick: function () {
      busy(retry, true);
      api('POST', '/cups/payouts/retry', { cupId: id }).then(function (o) { toast('success', t('cu.retried', { n: o.requeued || 0 })); refresh(true); }, function (err) { busy(retry, false); toast('error', cuErrText(err)); });
    } }) : null;
    return section(t('cu.payouts'), [retry ? h('div', { class: 'mb-3' }, retry) : null, table([
      { label: t('f.player') }, { label: t('cu.kind') }, { label: t('ru.prize') }, { label: t('cu.amounts') }, { label: t('f.state') }, { label: t('f.date') }
    ], payouts.map(function (p) {
      var am = Object.keys(p.amounts || {}).map(function (k) { return fmtNum(p.amounts[k]) + ' ' + k; }).join('، ');
      return { id: p.id, item: p, cells: [
        playerLink(p.userId, p.username),
        h('span', { text: payoutKind(p.kind) }),
        p.prizeKey ? mono(p.prizeKey) : t('none'),
        h('span', { dir: 'auto', class: 'text-xs', text: am || t('none') }),
        h('span', { title: p.lastError || '' }, tag(t('pay.' + p.status), p.status === 'paid' ? 'success' : p.status === 'failed' ? 'error' : 'warning')),
        h('span', { class: 'text-xs', text: fmtDate(p.paidAt || p.createdAt) })
      ] };
    }))], 'payouts');
  }
  function addPlayerPanel() {
    var who = input('', { name: 'userId', ltr: true, placeholder: t('cu.hint.who') });
    var msg = h('div', {});
    var ok = button({ label: t('cu.addPlayer'), variant: 'primary', data: { 'cup-action': 'save-player' } });
    sidePanel(t('cu.addPlayer'), [field(t('f.player'), who, t('cu.hint.whoLong')), h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('cu.hint.add') }), msg], [button({ label: t('cancel'), onclick: closePanel }), ok], 'add-player');
    who.focus();
    ok.addEventListener('click', function () {
      clear(msg);
      busy(ok, true);
      resolveUser(who.value).then(function (uid) {
        return api('POST', '/cups/list/' + encodeURIComponent(id) + '/players', { userId: uid }).then(function (out) { closePanel(); toast('success', t('cu.added')); data = out; paint(); });
      }).catch(function (err) { busy(ok, false); fieldError(msg, err); });
    });
  }
  function swapPanel(p) {
    var who = input('', { name: 'to', ltr: true, placeholder: t('cu.hint.who') });
    var msg = h('div', {});
    var ok = button({ label: t('cu.swap'), variant: 'primary', data: { 'cup-action': 'save-swap' } });
    sidePanel(t('cu.swapTitle', { name: p.username }), [h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('cu.hint.swap') }), field(t('cu.swapTo'), who, t('cu.hint.whoLong')), msg], [button({ label: t('cancel'), onclick: closePanel }), ok], 'swap');
    who.focus();
    ok.addEventListener('click', function () {
      clear(msg);
      busy(ok, true);
      resolveUser(who.value).then(function (uid) {
        return api('POST', '/cups/list/' + encodeURIComponent(id) + '/swap', { from: p.userId, to: uid }).then(function (out) { closePanel(); toast('success', t('cu.swapped')); data = out; paint(); });
      }).catch(function (err) { busy(ok, false); fieldError(msg, err); });
    });
  }
  function disqualifyPanel(p) {
    var refund = check(false, t('cu.dqRefund'), 'refund');
    var msg = h('div', {});
    var ok = button({ label: t('cu.disqualify'), variant: 'danger', class: 'cu-btn--outline', data: { 'cup-action': 'save-disqualify' } });
    sidePanel(t('cu.dqTitle', { name: p.username }), [h('p', { class: 'text-sm mb-3', text: t('cu.hint.dq') }), refund, msg], [button({ label: t('cancel'), onclick: closePanel }), ok], 'disqualify');
    ok.addEventListener('click', function () {
      clear(msg);
      busy(ok, true);
      api('POST', '/cups/list/' + encodeURIComponent(id) + '/players/' + p.userId + '/disqualify', { refund: refund.input.checked }).then(function (out) { closePanel(); toast('success', t('cu.disqualified')); data = out; paint(); },
        function (err) { busy(ok, false); fieldError(msg, err); });
    });
  }
  function cancelPanel(c) {
    var reason = input('', { name: 'reason', placeholder: t('cu.hint.reason') });
    var msg = h('div', {});
    var ok = button({ label: t('cu.cancel'), variant: 'danger', class: 'cu-btn--outline', data: { 'cup-action': 'save-cancel' } });
    sidePanel(t('cu.cancelTitle'), [h('p', { class: 'text-sm mb-3', text: t('cu.hint.cancel') }), field(t('cu.cancelReason'), reason), msg], [button({ label: t('close'), onclick: closePanel }), ok], 'cancel');
    ok.addEventListener('click', function () {
      clear(msg);
      busy(ok, true);
      api('POST', '/cups/list/' + encodeURIComponent(id) + '/cancel', { reason: reason.value.trim() }).then(function (out) { closePanel(); toast('success', t('cu.cancelled')); data = out; paint(); },
        function (err) { busy(ok, false); fieldError(msg, err); });
    });
  }
  Promise.all([loadLanguages(), loadRefs('cups')]).then(function () { return refresh(false); }).then(function () {
    cuPoll = setInterval(function () {
      if (!root.isConnected || !document.querySelector('[data-cup-page="cups/view"]') || location.hash.indexOf('#/cups/list/' + id) !== 0) { stopPoll(); return; }
      if (!data || busyUi() || document.hidden) return;
      var st = data.cup.status;
      if (st === 'finished' || st === 'cancelled' || st === 'draft') return;
      refresh(true);
    }, CU_POLL_MS);
  });
}
