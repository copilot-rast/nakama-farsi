function reportPage(root) {
  var q = hashQuery();
  var state = { days: Number(q.days) || 30, chapterId: q.chapterId || '' };
  var frame = pageFrame('report', t('desc.report'));
  clear(root).appendChild(frame.el);
  var daysSel = select(String(state.days), [7, 30, 90].map(function (n) { return { value: String(n), label: t('rp.d', { n: n }) }; }), { name: 'days', label: t('rp.days') });
  var chapterSel = h('select', { class: CG_INPUT + ' eco-select', name: 'chapterId', 'aria-label': t('f.chapter') });
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx cg-filters' }, field(t('rp.days'), daysSel), field(t('f.chapter'), chapterSel)));
  frame.content.appendChild(body);
  daysSel.addEventListener('change', function () { state.days = Number(daysSel.value); reload(); });
  chapterSel.addEventListener('change', function () { state.chapterId = chapterSel.value; reload(); });
  function kpi(name, value) {
    return card(h('div', { class: 'p-4', 'data-campaign-kpi': name }, h('div', { class: 'text-xs text-secondary-500', text: t('rp.kpi.' + name) }), h('div', { class: 'text-xl font-semibold mt-1', text: fmtNum(value || 0) })));
  }
  function bar(pct, name) {
    return h('span', { class: 'flex items-center gap-2 min-w-[8rem]' }, h('span', { class: 'cg-bar flex-1' }, h('span', { style: 'width:' + Math.max(0, Math.min(100, pct || 0)) + '%' })),
      h('span', { class: 'text-xs whitespace-nowrap', 'data-campaign-rate': name, text: fmtNum(pct || 0) + '٪' }));
  }
  function reload() {
    clear(body).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/campaign/chapters'), api('GET', '/campaign/report' + qs({ days: state.days, chapterId: state.chapterId, lang: cgLang() }))]).then(function (res) {
      clear(chapterSel);
      [{ value: '', label: t('st.filter.chapter') }].concat((res[0].items || []).map(function (c) { return { value: c.id, label: textOf(c.name, c.key) }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.chapterId) opt.selected = true;
        chapterSel.appendChild(opt);
      });
      var d = res[1];
      clear(body);
      var tot = d.totals || {};
      body.appendChild(h('div', { class: 'grid grid-cols-2 md:grid-cols-4 gap-3' }, kpi('players', tot.players), kpi('attempts', tot.attempts), kpi('wins', tot.wins), kpi('rejected', tot.rejected)));
      var items = d.items || [];
      if (!items.length) { body.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      body.appendChild(h('div', { 'data-campaign-list': 'report' }, table([
        { label: t('rp.stage') }, { label: t('f.mode') }, { label: t('rp.players') }, { label: t('rp.attempts') }, { label: t('rp.wins') }, { label: t('rp.rate') }, { label: t('rp.avgStars') }, { label: t('rp.drop') }, { label: t('rp.rejected') }
      ], items.map(function (r) {
        return { id: r.key, item: r, cells: [
          h('div', { class: 'flex flex-col' }, h('span', { class: 'font-medium', text: r.name }), mono(r.chapterKey + ' / ' + r.key)), modeTag(r.mode),
          h('span', { 'data-campaign-players': r.key, text: fmtNum(r.players) }), h('span', { text: fmtNum(r.attempts) }), h('span', { text: fmtNum(r.wins) }),
          bar(r.completionRate, r.key), h('span', { 'data-campaign-avg': r.key, text: fmtNum(r.avgBestStars) }),
          h('span', { 'data-campaign-drop': r.key, text: r.dropOff == null ? t('none') : fmtNum(r.dropOff) + '٪' }), h('span', { text: fmtNum(r.rejected) })
        ] };
      }))));
    }, function (err) { clear(body).appendChild(notice('error', cgErrText(err))); });
  }
  reload();
}
