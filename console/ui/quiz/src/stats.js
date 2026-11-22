function statsPage(root) {
  var q = hashQuery();
  var days = Number(q.days) || 30;
  var frame = pageFrame('stats', t('desc.stats'));
  clear(root).appendChild(frame.el);
  var daySel = select(String(days), [7, 30, 90].map(function (d) { return { value: String(d), label: t('sx.d', { n: d }) }; }), { name: 'days', label: t('sx.days') });
  daySel.addEventListener('change', function () { location.hash = '#/quiz/stats?days=' + daySel.value; });
  var box = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx qz-filters' }, field(t('sx.days'), daySel)));
  frame.content.appendChild(box);
  box.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  api('GET', '/quiz/stats' + qs({ days: days, lang: qzLang() })).then(function (d) {
    clear(box);
    var tot = d.totals || {};
    var rate = tot.finished ? Math.round(tot.passes * 1000 / tot.finished) / 10 : 0;
    var kpi = function (k, v) {
      return card(h('div', { class: 'p-4', 'data-quiz-kpi': k }, h('div', { class: 'text-xs text-secondary-500', text: t('sx.kpi.' + k) }), h('div', { class: 'text-2xl font-semibold mt-1', text: v })));
    };
    box.appendChild(h('div', { class: 'grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-3' }, kpi('players', fmtNum(tot.players)), kpi('plays', fmtNum(tot.plays)), kpi('finished', fmtNum(tot.finished)),
      kpi('rate', fmtNum(rate) + '٪'), kpi('hints', fmtNum(tot.hintsBought)), kpi('reports', fmtNum(tot.openReports))));
    var packs = d.packs || [];
    box.appendChild(h('div', { 'data-quiz-list': 'stats' }, table([
      { label: t('f.pack') }, { label: t('f.mission') }, { label: t('sx.players') }, { label: t('sx.plays') }, { label: t('sx.finished') }, { label: t('sx.passes') }, { label: t('sx.rate') }, { label: t('sx.avg') }, { label: t('sx.secs') }
    ], packs.map(function (p) {
      return { id: p.key, cells: [h('div', { class: 'flex flex-col' }, h('span', { class: 'font-medium', text: p.name }), mono(p.key)), h('span', { text: p.missionKey || t('none') }),
        h('span', { text: fmtNum(p.players) }), h('span', { text: fmtNum(p.plays) }), h('span', { text: fmtNum(p.finished) }), h('span', { text: fmtNum(p.passes) }),
        h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'qz-bar flex-1' }, h('span', { style: 'width:' + (p.passRate || 0) + '%' })), h('span', { class: 'text-xs', text: fmtNum(p.passRate) + '٪' })),
        h('span', { text: fmtNum(p.avgScore) }), h('span', { text: fmtNum(p.avgSeconds) })] };
    }))));
    if (!packs.length) box.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') })));
    var hard = d.hardest || [];
    box.appendChild(section(t('sx.hardest'), hard.length ? table([{ label: t('qs.question') }, { label: t('f.pack') }, { label: t('sx.answers') }, { label: t('sx.correctRate') }, { label: t('sx.skipped') }],
      hard.map(function (x) {
        return { id: x.id, cells: [h('span', { dir: 'auto', class: 'qz-sentence', text: x.question }), h('span', { text: x.packKey || t('qs.bank') }), h('span', { text: fmtNum(x.answers) }),
          h('span', { text: fmtNum(x.correctRate) + '٪' }), h('span', { text: fmtNum(x.skipped) })] };
      })) : h('p', { class: 'text-sm text-secondary-500', text: t('sx.noHard') }), 'hardest'));
  }, function (err) { clear(box).appendChild(notice('error', qzErrText(err))); });
}
