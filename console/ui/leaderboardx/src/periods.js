function winnersPage(root) {
  var q = hashQuery();
  var state = { board: q.board || '', page: 1, limit: 20 };
  var frame = pageFrame('winners', t('desc.winners'));
  clear(root).appendChild(frame.el);
  var head = h('div', { class: 'eco-toolbar eco-tbx lb-filters' });
  var body = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(head);
  frame.content.appendChild(body);
  var boards = [];
  Promise.all([api('GET', '/leaderboardx/boards'), loadLanguages()]).then(function (res) {
    boards = res[0].items || [];
    head.appendChild(field(t('sd.board'), boardPicker(boards, state.board, function (v) { state.board = v; state.page = 1; load(); }, t('wn.allBoards'))));
    load();
  }, function (err) { body.appendChild(notice('error', lbErrText(err))); });
  function counts(c) {
    var box = h('div', { class: 'flex flex-wrap gap-1' });
    ['paid', 'pending', 'claimable', 'failed', 'expired'].forEach(function (k) {
      if (c[k]) { var el = resultTag(k); el.textContent = t('rs.' + k) + ' ' + digits(String(c[k])); box.appendChild(el); }
    });
    return box.childNodes.length ? box : h('span', { text: t('none') });
  }
  function load() {
    clear(body).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', '/leaderboardx/periods' + qs({ boardKey: state.board, page: state.page, limit: state.limit })).then(function (d) {
      clear(body);
      var items = d.items || [];
      if (!items.length) { body.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      body.appendChild(h('div', { 'data-lbx-list': 'periods' }, table([
        { label: t('sd.board') }, { label: t('wn.end') }, { label: t('wn.reason') }, { label: t('wn.players') }, { label: t('wn.rewarded') }, { label: t('wn.status') }, { label: '' }
      ], items.map(function (p) {
        return { id: p.id, item: p, cells: [
          h('div', { class: 'flex flex-col' }, h('span', { class: 'font-medium', text: textOf(p.boardName, p.boardKey) }), mono(p.boardKey)),
          h('span', { class: 'text-xs', text: fmtDate(p.periodEnd) }),
          tag(t('reason.' + p.reason), p.reason === 'reset' ? 'primary' : 'secondary'),
          h('span', { 'data-lbx-period-players': p.id, text: fmtNum(p.players) }),
          h('span', { text: fmtNum(p.rewarded) }),
          h('div', { 'data-lbx-period-counts': p.id }, counts(p.counts || {})),
          p.counts && p.counts.failed ? button({ label: t('wn.retry'), variant: 'primary', data: { 'lbx-retry': p.id }, onclick: function () {
            api('POST', '/leaderboardx/periods/' + p.id + '/retry', {}).then(function (out) { toast('success', t('wn.retried', { n: out.requeued, p: out.paid })); load(); }, function (err) { toast('error', lbErrText(err)); });
          } }) : null
        ] };
      }), function (p) { periodPanel(p, boards); })));
      body.appendChild(pager({ page: state.page, limit: state.limit, total: d.total }, function (n) { state.page = n; load(); }));
    }, function (err) { clear(body).appendChild(notice('error', lbErrText(err))); });
  }
}

function periodPanel(p, boards) {
  var x = boards.filter(function (b) { return b.board.key === p.boardKey; })[0];
  var b = x ? x.board : { tiers: [] };
  var only = check(false, t('wn.onlyRewarded'), 'onlyRewarded');
  var list = h('div', { 'data-lbx-results': p.id });
  var state = { page: 1, limit: 50 };
  function load() {
    clear(list).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', '/leaderboardx/periods/' + p.id + qs({ status: only.input.checked ? 'rewarded' : '', page: state.page, limit: state.limit })).then(function (d) {
      clear(list);
      var items = d.items || [];
      if (!items.length) { list.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('empty') })); return; }
      var ol = h('ol', { class: 'flex flex-col gap-2' });
      items.forEach(function (r) {
        ol.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-sm border-b border-secondary-200 dark:border-secondary-700 pb-2', 'data-lbx-result': r.userId },
          h('span', { class: 'font-semibold min-w-[2rem]', text: fmtNum(r.rank) }), playerLink(r.userId, r.username),
          h('span', { class: 'text-xs', text: fmtNum(r.score) }), r.tier ? tag(tierName(b, r.tier), 'primary') : null,
          h('span', { class: 'text-xs', text: rewardsText(r.rewards) }), resultTag(r.status),
          r.lastError ? h('span', { class: 'text-xs text-error-600', text: t('wn.error') + ': ' + r.lastError }) : null));
      });
      list.appendChild(ol);
      list.appendChild(pager({ page: state.page, limit: state.limit, total: d.total }, function (n) { state.page = n; load(); }));
    }, function (err) { clear(list).appendChild(notice('error', lbErrText(err))); });
  }
  only.input.addEventListener('change', function () { state.page = 1; load(); });
  sidePanel(textOf(p.boardName, p.boardKey) + ' · ' + fmtDate(p.periodEnd), h('div', {}, only, list), [button({ label: t('close'), onclick: closePanel })], 'period');
  load();
}
