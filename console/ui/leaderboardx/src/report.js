function reportPage(root) {
  var q = hashQuery();
  var state = { days: Number(q.days) || 30 };
  var frame = pageFrame('report', t('desc.report'));
  clear(root).appendChild(frame.el);
  var daysSel = select(String(state.days), [7, 30, 90].map(function (n) { return { value: String(n), label: t('rp.d', { n: n }) }; }), { name: 'days', label: t('rp.days') });
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx lb-filters' }, field(t('rp.days'), daysSel)));
  frame.content.appendChild(body);
  daysSel.addEventListener('change', function () { state.days = Number(daysSel.value); reload(); });
  function kpi(name, value) {
    return card(h('div', { class: 'p-4', 'data-lbx-kpi': name }, h('div', { class: 'text-xs text-secondary-500', text: t('rp.kpi.' + name) }), h('div', { class: 'text-xl font-semibold mt-1', text: fmtNum(value || 0) })));
  }
  function reload() {
    clear(body).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', '/leaderboardx/report' + qs({ days: state.days })).then(function (d) {
      clear(body);
      var tot = d.totals || {};
      body.appendChild(h('div', { class: 'grid grid-cols-2 md:grid-cols-4 gap-3' }, kpi('boards', tot.boards), kpi('players', tot.players), kpi('writes', tot.writes), kpi('refused', tot.refused),
        kpi('paid', tot.paid), kpi('open', tot.open), kpi('failed', tot.failed)));
      var daily = d.daily || [];
      if (daily.length) {
        var peak = Math.max.apply(null, daily.map(function (x) { return x.writes + x.refused; }).concat([1]));
        var bars = h('ol', { class: 'flex flex-col gap-1', 'data-lbx-daily': '' });
        daily.forEach(function (x) {
          bars.appendChild(h('li', { class: 'flex items-center gap-2 text-xs' }, h('span', { class: 'w-24 shrink-0', dir: 'ltr', text: x.day }),
            h('span', { class: 'lb-bar flex-1' }, h('span', { style: 'width:' + Math.round((x.writes / peak) * 100) + '%' })),
            h('span', { class: 'w-20 text-end', text: fmtNum(x.writes) + (x.refused ? ' / ' + fmtNum(x.refused) : '') })));
        });
        body.appendChild(section(t('rp.daily'), bars, 'daily'));
      }
      var items = d.items || [];
      if (!items.length) { body.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      body.appendChild(h('div', { 'data-lbx-list': 'report' }, table([
        { label: t('sd.board') }, { label: t('rp.kpi.players') }, { label: t('rp.kpi.writes') }, { label: t('rp.kpi.refused') }, { label: t('rp.writers') }, { label: t('rp.periods') }, { label: t('wn.status') }
      ], items.map(function (r) {
        var rw = r.rewards || {};
        return { id: r.key, item: r, cells: [
          h('div', { class: 'flex flex-col' }, h('span', { class: 'font-medium', text: textOf(r.name, r.key) }), mono(r.key)),
          h('span', { 'data-lbx-report-players': r.key, text: fmtNum(r.players) }),
          h('span', { 'data-lbx-report-writes': r.key, text: fmtNum(r.writes) }),
          h('span', { 'data-lbx-report-refused': r.key, text: fmtNum(r.refused) }),
          h('span', { text: fmtNum(r.writers) }), h('span', { text: fmtNum(r.periods) }),
          h('span', { class: 'text-xs', 'data-lbx-report-rewards': r.key, text: t('rs.paid') + ' ' + fmtNum(rw.paid) + ' · ' + t('rs.claimable') + ' ' + fmtNum(rw.claimable) + ' · ' + t('rs.failed') + ' ' + fmtNum(rw.failed) })
        ] };
      }))));
    }, function (err) { clear(body).appendChild(notice('error', lbErrText(err))); });
  }
  reload();
}
