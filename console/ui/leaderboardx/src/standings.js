// boardPicker loads the boards and calls onpick with the chosen one (kept in the hash as ?board=key).
function boardPicker(list, current, onpick, allLabel) {
  var opts = allLabel ? [{ value: '', label: allLabel }] : [];
  list.forEach(function (x) { opts.push({ value: x.board.key, label: textOf(x.board.name, x.board.key) + ' · ' + x.board.key }); });
  var sel = select(current || '', opts, { name: 'board', label: t('sd.board') });
  sel.setAttribute('dir', 'auto');
  sel.addEventListener('change', function () { onpick(sel.value); });
  return sel;
}

function standingsPage(root) {
  var q = hashQuery();
  var state = { board: q.board || '', find: '', cursor: '', stack: [] };
  var frame = pageFrame('standings', t('desc.standings'));
  clear(root).appendChild(frame.el);
  var head = h('div', { class: 'flex flex-wrap gap-2 items-end lb-filters' });
  var body = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(head);
  frame.content.appendChild(body);
  var boards = [];
  function current() { return boards.filter(function (x) { return x.board.key === state.board; })[0]; }
  api('GET', '/leaderboardx/boards').then(function (d) {
    boards = d.items || [];
    if (!boards.length) { body.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('sd.noBoards') }))); return; }
    if (!current()) state.board = boards[0].board.key;
    var find = input('', { name: 'find', ltr: true, placeholder: t('sd.find') });
    find.addEventListener('keydown', function (e) { if (e.key === 'Enter') { state.find = find.value.trim(); state.cursor = ''; state.stack = []; load(); } });
    var closeBtn = button({ label: t('sd.close'), variant: 'danger', data: { 'lbx-action': 'close-period' }, onclick: closePeriod });
    head.appendChild(field(t('sd.board'), boardPicker(boards, state.board, function (v) { state.board = v; state.cursor = ''; state.stack = []; load(); })));
    head.appendChild(field(t('search'), find));
    head.appendChild(h('div', { class: 'mb-4' }, closeBtn));
    load();
  }, function (err) { body.appendChild(notice('error', lbErrText(err))); });
  function closePeriod() {
    var x = current();
    if (!x) return;
    var pay = check(true, t('sd.closePay'), 'closePay');
    var root2;
    function done(ok) {
      root2.remove();
      if (!ok) return;
      api('POST', '/leaderboardx/boards/' + x.board.id + '/close', { pay: pay.input.checked }).then(function () { toast('success', t('sd.closed')); load(); }, function (err) { toast('error', lbErrText(err)); });
    }
    root2 = h('div', { 'data-lbx': 'confirm', 'data-l10n-skip': '', class: 'lb-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'lb-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('sd.close') }), h('p', { class: 'text-sm mb-3', text: t('sd.closeConfirm') }), pay,
        h('div', { class: 'flex gap-2 justify-end' }, button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: t('sd.close'), variant: 'danger', data: { 'lbx-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root2);
  }
  function load() {
    var x = current();
    if (!x) return;
    clear(body).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', '/leaderboardx/boards/' + x.board.id + '/standings' + qs({ limit: 50, cursor: state.cursor, q: state.find })).then(function (d) {
      clear(body);
      var b = d.board;
      body.appendChild(h('div', { class: 'flex flex-wrap items-center gap-3 text-sm', 'data-lbx-standing-head': b.key },
        statusTag(d.status), h('span', { 'data-lbx-total': '', text: t('b.players', { n: d.players }) }),
        h('span', { class: 'text-secondary-500', text: d.periodEnd ? t('sd.period', { from: fmtDate(d.periodStart), to: fmtDate(d.periodEnd) }) : t('b.noEnd') })));
      var items = d.items || [];
      if (!items.length) { body.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      body.appendChild(h('div', { 'data-lbx-list': 'standings' }, table([{ label: t('sd.rank') }, { label: t('sd.player') }, { label: t('sd.score') }, { label: t('sd.tier') }, { label: '' }],
        items.map(function (r) {
          return { id: r.userId, item: r, cells: [
            h('span', { class: 'font-semibold', 'data-lbx-rank': r.userId, text: fmtNum(r.rank) }),
            playerLink(r.userId, r.displayName || r.username),
            h('span', { 'data-lbx-score': r.userId, text: fmtNum(r.score) + (r.subscore ? ' · ' + fmtNum(r.subscore) : '') }),
            r.tier ? tag(tierName(b, r.tier), 'primary') : h('span', { text: t('none') }),
            h('div', { class: 'flex gap-1' }, button({ label: t('sd.edit'), data: { 'lbx-edit': r.userId }, onclick: function () { scoreEditor(b, r, load); } }),
              button({ label: t('remove'), variant: 'danger', data: { 'lbx-remove': r.userId }, onclick: function () {
                confirmBox(tr('sd.removeScore', null, { name: r.username || r.userId }), t('remove')).then(function (ok) {
                  if (!ok) return;
                  api('DELETE', '/leaderboardx/boards/' + b.id + '/records/' + r.userId).then(function () { toast('success', t('saved')); load(); }, function (err) { toast('error', lbErrText(err)); });
                });
              } }))
          ] };
        }))));
      if (!state.find) {
        body.appendChild(h('div', { class: 'flex justify-end gap-2' },
          button({ label: t('prev'), disabled: !state.stack.length, onclick: function () { state.cursor = state.stack.pop() || ''; load(); } }),
          button({ label: t('next'), disabled: !d.nextCursor, onclick: function () { state.stack.push(state.cursor); state.cursor = d.nextCursor; load(); } })));
      }
    }, function (err) { clear(body).appendChild(notice('error', lbErrText(err))); });
  }
}

function scoreEditor(b, r, done) {
  var score = numInput(r.score, 'score', t('sd.score'));
  var mode = select('set', [{ value: 'set', label: t('sd.mode.set') }, { value: 'add', label: t('sd.mode.add') }], { name: 'mode' });
  var msg = h('div', {});
  var save = button({ label: t('save'), variant: 'primary', data: { 'lbx-action': 'save-score' }, onclick: function () {
    busy(save, true);
    api('PUT', '/leaderboardx/boards/' + b.id + '/records/' + r.userId, { score: numberValue(score) || 0, mode: mode.value }).then(function () {
      closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); clear(msg).appendChild(notice('error', lbErrText(err))); });
  } });
  sidePanel(t('sd.edit') + ' · ' + (r.username || r.userId), h('div', { 'data-lbx-form': 'score' }, field(t('sd.score'), score), field(t('sd.mode'), mode), msg), [button({ label: t('cancel'), onclick: closePanel }), save], 'score');
}
