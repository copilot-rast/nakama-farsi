function colorDot(c) {
  return h('span', { class: 'qz-dot', style: c ? 'background:' + c : '', 'aria-hidden': 'true' });
}

function duelCategoriesPage(root) {
  var state = { page: 1, limit: 50, total: 0, q: '' };
  var frame = pageFrame('duel-categories', t('desc.duel-categories'), [
    button({ label: t('dc.new'), variant: 'primary', data: { 'quiz-action': 'new-duel-category' }, onclick: function () { duelCategoryEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  var search = input('', { name: 'q', placeholder: t('search') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx qz-filters' }, field(t('search'), search)));
  frame.content.appendChild(listBox);
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.q = search.value.trim(); state.page = 1; reload(); }, 300); });
  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/duel-categories' + qs({ q: state.q, page: state.page, limit: state.limit })), loadLanguages()]).then(function (res) {
      var d = res[0];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'duel-categories' }, table([
        { label: t('f.name') }, { label: t('dc.questions') }, { label: t('dc.picks') }, { label: t('dc.state') }, { label: '' }
      ], items.map(function (c) {
        return { id: c.key, item: c, cells: [
          h('div', { class: 'flex items-center gap-2' }, c.iconUrl ? h('img', { src: c.iconUrl, alt: '', class: 'qz-cat-icon' }) : colorDot(c.color),
            h('div', { class: 'flex flex-col' }, h('span', { class: 'font-medium', dir: 'auto', text: textOf(c.name, c.key) }), mono(c.key))),
          h('div', { class: 'flex flex-col gap-1' }, h('span', { text: t('dc.count', { n: fmtNum(c.questions), e: fmtNum(c.eligible) }) }),
            c.playable ? tag(t('dc.playable'), 'success') : tag(t('dc.notEnough', { n: fmtNum(d.perRound) }), 'warning')),
          h('span', { text: fmtNum(c.rounds) }),
          h('div', { class: 'flex flex-wrap gap-1' }, c.isActive ? tag(t('active'), 'success') : tag(t('inactive'), 'warning'), c.duel ? tag(t('dc.inDuel'), 'primary') : tag(t('dc.notInDuel'), 'secondary')),
          button({ label: t('delete'), variant: 'danger', data: { 'quiz-delete': c.key }, onclick: function () {
            confirmBox(t('dc.delete', { key: c.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/quiz/duel-categories/' + c.id).then(function () { toast('success', t('saved')); reload(); }, function (err) {
                if (err.code !== 'IN_USE') { toast('error', qzErrText(err)); return; }
                confirmBox(t('dc.inUse'), t('delete')).then(function (force) {
                  if (force) api('DELETE', '/quiz/duel-categories/' + c.id + '?force=true').then(function () { toast('success', t('saved')); reload(); }, function (e2) { toast('error', qzErrText(e2)); });
                });
              });
            });
          } })
        ] };
      }), function (c) { duelCategoryEditor(c, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('dc.empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function duelCategoryEditor(c, done) {
  var creating = !c;
  c = c || { name: {}, description: {}, isActive: true, duel: true, color: '', sort: 0, iconUrl: '' };
  var key = input(c.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = translated(t('f.name'), 'name', c.name);
  var desc = translated(t('f.description'), 'description', c.description, { multiline: true });
  var icon = input(c.iconUrl, { name: 'iconUrl', ltr: true });
  var color = input(c.color || '', { name: 'color', ltr: true, mono: true, placeholder: '#22AA66' });
  var picker = h('input', { type: 'color', class: 'qz-color', name: 'color-picker', 'aria-label': t('dc.color'), value: c.color || '#3b82f6' });
  picker.addEventListener('input', function () { color.value = picker.value.toUpperCase(); });
  color.addEventListener('input', function () { if (/^#[0-9A-Fa-f]{6}$/.test(color.value.trim())) picker.value = color.value.trim(); });
  var sort = numInput(c.sort || 0, 'sort');
  var active = check(c.isActive, t('f.active'), 'isActive');
  var duel = check(c.duel, t('dc.duelOn'), 'duel');
  var msg = h('div', {});
  var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
  var body = h('div', { 'data-quiz-form': 'duel-category' },
    h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('dc.hint') }),
    field(t('f.key'), key, t('dc.hint.key')), name, desc,
    grid(field(t('f.icon'), icon), field(t('dc.color'), h('div', { class: 'flex gap-2 items-center' }, picker, color))),
    grid(field(t('f.sort'), sort), h('div', {})), active, duel, msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-duel-category' }, onclick: function () {
    busy(save, true);
    clear(msg);
    var out = { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), color: color.value.trim(), sort: numberValue(sort) || 0,
      isActive: active.input.checked, duel: duel.input.checked };
    api(creating ? 'POST' : 'PUT', '/quiz/duel-categories' + (creating ? '' : '/' + c.id), out).then(function () {
      closePanel(); toast('success', t('saved')); qzRefs = null; done();
    }, function (err) { busy(save, false); fieldError(msg, err); });
  } });
  sidePanel(creating ? t('dc.new') : textOf(c.name, c.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'duel-category');
}

function duelStatusTag(s) {
  var color = s === 'active' ? 'primary' : s === 'finished' ? 'success' : s === 'waiting' || s === 'invited' ? 'warning' : 'secondary';
  return tag(t('du.st.' + s), color);
}

function duelPlayer(p, winner) {
  if (!p) return h('span', { class: 'text-xs text-secondary-500', text: t('du.noOpponent') });
  return h('div', { class: 'flex items-center gap-1.5' }, playerLink(p.userId, p.displayName || p.username), h('span', { class: 'font-semibold', text: fmtNum(p.score) }),
    winner && winner === p.userId ? tag(t('du.winner'), 'success') : null);
}

function duelsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 20, total: 0, status: q.status || 'all', mode: '', player: q.player || '' };
  var frame = pageFrame('duels', t('desc.duels'));
  clear(root).appendChild(frame.el);
  var tiles = h('div', { class: 'grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-3', 'data-quiz-kpis': 'duels' });
  var statusSel = select(state.status, ['all', 'open', 'active', 'waiting', 'invited', 'finished', 'expired', 'declined', 'cancelled'].map(function (s) { return { value: s, label: s === 'all' ? t('all') : t('du.st.' + s) }; }), { name: 'status', label: t('du.status') });
  var modeSel = select('', [{ value: '', label: t('all') }, { value: 'random', label: t('du.mode.random') }, { value: 'friend', label: t('du.mode.friend') }], { name: 'mode', label: t('du.mode') });
  var player = input(state.player, { name: 'player', placeholder: t('du.playerHint') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(tiles);
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx qz-filters' }, field(t('du.status'), statusSel), field(t('du.mode'), modeSel), field(t('du.player'), player)));
  frame.content.appendChild(listBox);
  statusSel.addEventListener('change', function () { state.status = statusSel.value; state.page = 1; reload(); });
  modeSel.addEventListener('change', function () { state.mode = modeSel.value; state.page = 1; reload(); });
  var timer = null;
  player.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.player = player.value.trim(); state.page = 1; reload(); }, 300); });
  api('GET', '/quiz/duel-stats' + qs({ days: 30, lang: qzLang() })).then(function (s) {
    var kpi = function (k, v) {
      return card(h('div', { class: 'p-4', 'data-quiz-kpi': k }, h('div', { class: 'text-xs text-secondary-500', text: t('du.kpi.' + k) }), h('div', { class: 'text-2xl font-semibold mt-1', text: v })));
    };
    clear(tiles);
    [kpi('started', fmtNum(s.started)), kpi('finished', fmtNum(s.finished)), kpi('open', fmtNum(s.open)), kpi('waiting', fmtNum(s.waiting)), kpi('timeouts', fmtNum(s.timeouts)),
      kpi('players', fmtNum(s.players))].forEach(function (x) { tiles.appendChild(x); });
  }, function () { clear(tiles); });
  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', '/quiz/duels' + qs({ status: state.status, mode: state.mode, player: state.player, page: state.page, limit: state.limit })).then(function (d) {
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'duels' }, table([
        { label: t('du.playerA') }, { label: t('du.playerB') }, { label: t('du.status') }, { label: t('du.mode') }, { label: t('du.turn') }, { label: t('du.updated') }
      ], items.map(function (x) {
        var turn = x.turnUser ? (x.playerA && x.turnUser === x.playerA.userId ? x.playerA : x.playerB) : null;
        return { id: x.id, item: x, cells: [
          duelPlayer(x.playerA, x.winner), duelPlayer(x.playerB, x.winner),
          h('div', { class: 'flex flex-col gap-1' }, duelStatusTag(x.status), x.reason ? h('span', { class: 'text-xs text-secondary-500', text: t('du.reason.' + x.reason) }) : null),
          h('span', { text: t('du.mode.' + x.mode) }),
          turn ? h('div', { class: 'flex flex-col' }, h('span', { dir: 'auto', text: turn.displayName || turn.username }), x.turnDeadline ? h('span', { class: 'text-xs text-secondary-500', text: t('du.until', { at: fmtDate(x.turnDeadline) }) }) : null) : h('span', { text: t('none') }),
          h('span', { class: 'text-xs', text: fmtDate(x.updatedAt) })
        ] };
      }), function (x) { duelPanel(x.id, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('du.empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function duelAnswerCell(a, options) {
  if (!a) return h('span', { class: 'text-xs text-secondary-400', text: t('du.notAnswered') });
  var opt = (options || []).filter(function (o) { return o.id === a.optionId; })[0];
  var color = a.outcome === 'correct' ? 'success' : a.outcome === 'pending' ? 'warning' : 'error';
  return h('div', { class: 'flex flex-col gap-0.5' }, tag(t('du.out.' + a.outcome), color),
    opt ? h('span', { class: 'text-xs', dir: 'auto', text: opt.text }) : null,
    a.answeredAt ? h('span', { class: 'text-xs text-secondary-500', text: t('du.secs', { n: fmtNum(a.seconds) }) }) : h('span', { class: 'text-xs text-secondary-500', text: t('du.served', { at: fmtDate(a.servedAt) }) }));
}

function duelPanel(id, done) {
  var body = h('div', { 'data-quiz-form': 'duel' }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  var cancel = button({ label: t('du.cancel'), variant: 'danger', data: { 'quiz-action': 'cancel-duel' }, disabled: true });
  sidePanel(t('du.detail'), body, [button({ label: t('close'), onclick: closePanel }), cancel], 'duel');
  api('GET', '/quiz/duels/' + id + qs({ lang: qzLang() })).then(function (d) {
    clear(body);
    var nameOf = function (p) { return p ? (p.displayName || p.username) : t('du.noOpponent'); };
    var open = d.status === 'waiting' || d.status === 'invited' || d.status === 'active';
    cancel.disabled = !open;
    cancel.onclick = function () {
      confirmBox(t('du.cancelConfirm'), t('du.cancel')).then(function (ok) {
        if (!ok) return;
        busy(cancel, true);
        api('POST', '/quiz/duels/' + id + '/cancel', {}).then(function () { closePanel(); toast('success', t('du.cancelled')); done(); }, function (err) { busy(cancel, false); toast('error', qzErrText(err)); });
      });
    };
    var fee = d.rules && d.rules.entryFee && d.rules.entryFee.amount ? t('du.feeV', { n: fmtNum(d.rules.entryFee.amount), c: d.rules.entryFee.currencyKey }) : t('none');
    body.appendChild(h('div', { class: 'flex flex-wrap gap-2 mb-3' }, duelStatusTag(d.status), tag(t('du.mode.' + d.mode), 'secondary'), d.reason ? tag(t('du.reason.' + d.reason), 'secondary') : null));
    body.appendChild(h('div', { class: 'grid sm:grid-cols-2 gap-3 mb-4' },
      card(h('div', { class: 'p-3' }, h('div', { class: 'text-xs text-secondary-500', text: t('du.playerA') }), duelPlayer(d.playerA, d.winner), h('div', { class: 'text-xs text-secondary-500 mt-1', text: t('du.fee.' + (d.feeA || 'none')) }))),
      card(h('div', { class: 'p-3' }, h('div', { class: 'text-xs text-secondary-500', text: t('du.playerB') }), duelPlayer(d.playerB, d.winner), h('div', { class: 'text-xs text-secondary-500 mt-1', text: t('du.fee.' + (d.feeB || 'none')) })))));
    body.appendChild(h('dl', { class: 'qz-dl text-sm mb-4' },
      h('dt', { text: t('du.created') }), h('dd', { text: fmtDate(d.createdAt) }),
      h('dt', { text: t('du.finished') }), h('dd', { text: d.finishedAt ? fmtDate(d.finishedAt) : t('none') }),
      h('dt', { text: t('du.deadline') }), h('dd', { text: d.turnDeadline ? fmtDate(d.turnDeadline) : t('none') }),
      h('dt', { text: t('du.entryFee') }), h('dd', { text: fee }),
      h('dt', { text: t('du.paid') }), h('dd', { text: d.status === 'finished' ? (d.paidA && d.paidB ? t('yes') : t('no')) : t('none') })));
    if (d.payError) body.appendChild(notice('error', d.payError));
    (d.rounds || []).forEach(function (r) {
      var picker = r.picker === (d.playerA && d.playerA.userId) ? d.playerA : d.playerB;
      var head = h('div', { class: 'flex flex-wrap items-center gap-2 mb-2' }, h('span', { class: 'font-semibold', text: t('du.round', { n: r.round }) }),
        r.category ? h('span', { class: 'flex items-center gap-1' }, colorDot(r.category.color), h('span', { dir: 'auto', text: r.category.name })) : tag(t('du.notPicked'), 'secondary'),
        h('span', { class: 'text-xs text-secondary-500', dir: 'auto', text: t('du.pickedBy', { name: nameOf(picker) }) }));
      var offers = r.offers && r.offers.length ? h('p', { class: 'text-xs text-secondary-500 mb-2', dir: 'auto', text: t('du.offers') + ': ' + r.offers.map(function (o) { return o.name; }).join('، ') }) : null;
      var rows = (r.questions || []).map(function (x, i) {
        return h('div', { class: 'qz-row', 'data-quiz-duel-question': x.questionId },
          h('div', { class: 'qz-row-full flex flex-col gap-1' }, h('span', { class: 'font-medium qz-sentence', dir: 'auto', text: fmtNum(i + 1) + '. ' + (x.text || x.questionId) }),
            h('div', { class: 'flex flex-wrap gap-1' }, (x.options || []).map(function (o) { return o.correct ? tag(o.text, 'success') : h('span', { class: 'text-xs text-secondary-500', dir: 'auto', text: o.text }); }))),
          h('div', {}, h('div', { class: 'text-xs text-secondary-500', dir: 'auto', text: nameOf(d.playerA) }), duelAnswerCell(x.a, x.options)),
          h('div', {}, h('div', { class: 'text-xs text-secondary-500', dir: 'auto', text: nameOf(d.playerB) }), duelAnswerCell(x.b, x.options)));
      });
      body.appendChild(section(t('du.round', { n: r.round }), h('div', {}, head, offers, h('div', { class: 'qz-rows' }, rows)), 'round-' + r.round));
    });
  }, function (err) { clear(body).appendChild(notice('error', qzErrText(err))); });
}

function duelSettingsSection(cur) {
  var d = cur.duel || {};
  var ev = d.events || {};
  var fee = d.entryFee || {};
  var pc = d.perCorrect || {};
  var opt = function (prefix, list) { return list.map(function (v) { return { value: v, label: t(prefix + v) }; }); };
  var enabled = check(d.enabled, t('ds.enabled'), 'duel.enabled');
  var rounds = numInput(d.rounds, 'duel.rounds');
  var per = numInput(d.questionsPerRound, 'duel.questionsPerRound');
  var choices = numInput(d.categoryChoices, 'duel.categoryChoices');
  var qsec = numInput(d.questionSeconds, 'duel.questionSeconds');
  var grace = numInput(d.graceSeconds, 'duel.graceSeconds');
  var source = select(d.questionSource, opt('ds.source.', ['any', 'bank']), { name: 'duel.questionSource' });
  var avoid = numInput(d.avoidRepeatDays, 'duel.avoidRepeatDays');
  var shuffle = check(d.shuffleOptions, t('ds.shuffle'), 'duel.shuffleOptions');
  var serveNext = check(d.serveNextOnAnswer, t('ds.serveNext'), 'duel.serveNextOnAnswer');
  var turn = numInput(d.turnHours, 'duel.turnHours');
  var timeout = select(d.timeoutResult, opt('ds.timeout.', ['opponent_wins', 'score']), { name: 'duel.timeoutResult' });
  var waiting = numInput(d.waitingHours, 'duel.waitingHours');
  var invite = numInput(d.inviteHours, 'duel.inviteHours');
  var friends = check(d.friendsOnly, t('ds.friendsOnly'), 'duel.friendsOnly');
  var maxActive = numInput(d.maxActive, 'duel.maxActive');
  var daily = numInput(d.dailyStarts, 'duel.dailyStarts');
  var feeCur = refSelect('currencies', fee.currencyKey, 'duel.entryFee.currencyKey', t('none'));
  var feeAmt = numInput(fee.amount || 0, 'duel.entryFee.amount');
  var win = refSelect('prizes', d.winPrize, 'duel.winPrize', t('none'));
  var lose = refSelect('prizes', d.losePrize, 'duel.losePrize', t('none'));
  var draw = refSelect('prizes', d.drawPrize, 'duel.drawPrize', t('none'));
  var pcCur = refSelect('currencies', pc.currencyKey, 'duel.perCorrect.currencyKey', t('none'));
  var pcAmt = numInput(pc.amount || 0, 'duel.perCorrect.amount');
  var evPlayed = input(ev.played, { name: 'duel.events.played', ltr: true, mono: true });
  var evWon = input(ev.won, { name: 'duel.events.won', ltr: true, mono: true });
  var evCorrect = input(ev.correct, { name: 'duel.events.correct', ltr: true, mono: true });
  var notify = check(d.notify, t('ds.notify'), 'duel.notify');
  var history = numInput(d.historyDays, 'duel.historyDays');
  var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
  return {
    el: section(t('ds.title'), h('div', { 'data-quiz-settings': 'duel' },
      h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('ds.hint') }), enabled,
      grid(field(t('ds.rounds'), rounds), field(t('ds.per'), per)),
      grid(field(t('ds.choices'), choices, t('ds.hint.choices')), field(t('ds.source'), source)),
      grid(field(t('ds.qsec'), qsec), field(t('ds.grace'), grace, t('ds.hint.grace'))),
      grid(field(t('ds.avoid'), avoid, t('ds.hint.avoid')), h('div', {})), shuffle, serveNext,
      grid(field(t('ds.turn'), turn, t('ds.hint.turn')), field(t('ds.timeout'), timeout)),
      grid(field(t('ds.waiting'), waiting, t('ds.hint.waiting')), field(t('ds.invite'), invite)), friends,
      grid(field(t('ds.maxActive'), maxActive), field(t('ds.daily'), daily, t('se.hint.zero'))),
      grid(field(t('ds.feeCur'), feeCur, t('ds.hint.fee')), field(t('ds.feeAmt'), feeAmt)),
      grid(field(t('ds.win'), win), field(t('ds.lose'), lose)),
      grid(field(t('ds.draw'), draw), h('div', {})),
      grid(field(t('ds.pcCur'), pcCur, t('ds.hint.pc')), field(t('ds.pcAmt'), pcAmt)),
      grid(field(t('ds.evPlayed'), evPlayed), field(t('ds.evWon'), evWon)),
      grid(field(t('ds.evCorrect'), evCorrect), field(t('ds.history'), history, t('se.hint.zeroKeep'))), notify), 'duel'),
    read: function (c) {
      c.duel = { enabled: enabled.input.checked, rounds: numberValue(rounds) || 0, questionsPerRound: numberValue(per) || 0, categoryChoices: numberValue(choices) || 0,
        questionSeconds: numberValue(qsec) || 0, graceSeconds: numberValue(grace) || 0, questionSource: source.value, avoidRepeatDays: numberValue(avoid) || 0,
        shuffleOptions: shuffle.input.checked, serveNextOnAnswer: serveNext.input.checked, turnHours: numberValue(turn) || 0, timeoutResult: timeout.value,
        waitingHours: numberValue(waiting) || 0, inviteHours: numberValue(invite) || 0, friendsOnly: friends.input.checked, maxActive: numberValue(maxActive) || 0,
        dailyStarts: numberValue(daily) || 0, entryFee: { currencyKey: feeCur.value, amount: feeCur.value ? numberValue(feeAmt) || 0 : 0 },
        winPrize: win.value, losePrize: lose.value, drawPrize: draw.value, perCorrect: { currencyKey: pcCur.value, amount: pcCur.value ? numberValue(pcAmt) || 0 : 0 },
        events: { played: evPlayed.value.trim(), won: evWon.value.trim(), correct: evCorrect.value.trim() }, notify: notify.input.checked, historyDays: numberValue(history) || 0 };
    }
  };
}
