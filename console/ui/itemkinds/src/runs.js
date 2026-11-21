function runsPage(root) {
  var frame = pageFrame('itemkinds/runs', t('desc.itemkinds/runs'));
  clear(root).appendChild(frame.el);
  var list = h('div', { 'data-ik-runs': '' });
  frame.content.appendChild(list);
  clear(list).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  api('GET', IK_BASE + '/runs?limit=50').then(function (d) {
    clear(list);
    if (!d.runs.length) { list.appendChild(h('div', { class: 'p-6 text-sm text-secondary-500 bg-white dark:bg-secondary-800 rounded-md shadow', text: t('run.empty') })); return; }
    var cols = [{ label: t('run.at') }, { label: t('run.source') }, { label: t('run.actor') }, { label: t('run.summary') }];
    list.appendChild(table(cols, d.runs.map(function (r) {
      return { id: String(r.id), item: r, cells: [h('span', { text: fmtDate(r.createdAt) }), h('span', { class: 'inline-flex gap-1' }, tag(t('run.' + r.source), 'secondary'), r.dryRun ? tag(t('run.dry'), 'warning') : null), mono(r.actor), summaryTags(r.summary)] };
    }), function (r) { runPanel(r.id); }));
  }, function (err) { clear(list).appendChild(notice('error', ikErrText(err))); });
}

function runPanel(id) {
  var body = h('div', { class: 'flex flex-col gap-3' }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  sidePanel(t('nav.itemkinds/runs'), body, [button({ label: t('close'), onclick: closePanel })], 'run');
  api('GET', IK_BASE + '/runs/' + id).then(function (d) {
    clear(body);
    body.appendChild(h('p', { class: 'text-sm', text: fmtDate(d.createdAt) + ' · ' + t('run.' + d.source) + ' · ' + d.actor }));
    body.appendChild(summaryTags(d.summary));
    body.appendChild(runRowsTable(d.rows));
  }, function (err) { clear(body).appendChild(notice('error', ikErrText(err))); });
}
