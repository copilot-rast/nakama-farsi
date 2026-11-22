function reportsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 50, total: 0, status: q.status != null ? q.status : 'open', reasonKey: '', questionId: q.questionId || '' };
  var frame = pageFrame('reports', t('desc.reports'), [button({ label: t('rp.reasons'), data: { 'quiz-action': 'reasons' }, onclick: reasonsPanel })]);
  clear(root).appendChild(frame.el);
  var statusSel = select(state.status, [{ value: '', label: t('all') }].concat(['open', 'resolved', 'dismissed'].map(function (s) { return { value: s, label: t('st.' + s) }; })), { name: 'status', label: t('rp.status') });
  var reasonSel = h('select', { class: QZ_INPUT + ' eco-select', name: 'reasonKey', 'aria-label': t('rp.reason') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx qz-filters' }, field(t('rp.status'), statusSel), field(t('rp.reason'), reasonSel)));
  frame.content.appendChild(listBox);
  statusSel.addEventListener('change', function () { state.status = statusSel.value; state.page = 1; reload(); });
  reasonSel.addEventListener('change', function () { state.reasonKey = reasonSel.value; state.page = 1; reload(); });
  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/reports' + qs({ status: state.status, reasonKey: state.reasonKey, questionId: state.questionId, page: state.page, limit: state.limit })), api('GET', '/quiz/reasons'), loadLanguages()]).then(function (res) {
      var d = res[0];
      var reasons = res[1].items || [];
      clear(reasonSel);
      [{ value: '', label: t('all') }].concat(reasons.map(function (r) { return { value: r.key, label: textOf(r.name, r.key) }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.reasonKey) opt.selected = true;
        reasonSel.appendChild(opt);
      });
      state.total = d.total || 0;
      clear(listBox);
      listBox.appendChild(h('div', { class: 'flex flex-wrap gap-2' }, tag(t('rp.openCount', { n: fmtNum(d.open) }), d.open ? 'warning' : 'success')));
      var items = d.items || [];
      listBox.appendChild(h('div', { 'data-quiz-list': 'reports' }, table([
        { label: t('rp.date') }, { label: t('rp.player') }, { label: t('qs.question') }, { label: t('rp.reason') }, { label: t('rp.note') }, { label: t('rp.status') }, { label: '' }
      ], items.map(function (r) {
        return { id: r.id, item: r, cells: [
          h('span', { class: 'text-xs', text: fmtDate(r.createdAt) }),
          playerLink(r.userId, r.username),
          h('div', { class: 'flex flex-col gap-1' }, h('span', { dir: 'auto', class: 'qz-sentence', text: textOf(r.question, r.questionId) }),
            h('div', { class: 'flex flex-wrap gap-1' }, r.packKey ? mono(r.packKey) : null, r.questionActive ? null : tag(t('inactive'), 'warning'), r.openReports > 1 ? tag(t('qs.reports', { n: r.openReports }), 'error') : null)),
          h('span', { text: textOf(r.reasonName, r.reasonKey) }),
          h('span', { class: 'text-xs qz-sentence', dir: 'auto', text: r.note || t('none') }),
          h('div', { class: 'flex flex-col gap-1' }, statusTag(r.status), r.resolution ? h('span', { class: 'text-xs text-secondary-500', dir: 'auto', text: r.resolution }) : null),
          button({ label: t('rp.review'), data: { 'quiz-review': r.id }, onclick: function () { reviewPanel(r, reload); } })
        ] };
      }), function (r) { reviewPanel(r, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('rp.empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function reviewPanel(r, done) {
  var resolution = input(r.resolution, { name: 'resolution', multiline: true, rows: 3 });
  var same = check(r.openReports > 1, t('rp.same', { n: r.openReports }), 'sameQuestion');
  var disable = check(false, t('rp.disable'), 'disableQuestion');
  var msg = h('div', {});
  var body = h('div', { 'data-quiz-form': 'review' },
    h('p', { class: 'font-medium mb-2', dir: 'auto', text: textOf(r.question, r.questionId) }),
    h('p', { class: 'text-sm mb-1', text: t('rp.reason') + ': ' + textOf(r.reasonName, r.reasonKey) }),
    r.note ? h('p', { class: 'text-sm mb-3', dir: 'auto', text: r.note }) : null,
    h('p', { class: 'mb-3' }, h('a', { class: 'eco-link', href: '#/quiz/questions?q=' + encodeURIComponent(r.questionId), 'data-quiz-edit-question': r.questionId, onclick: function (e) {
      e.preventDefault(); closePanel(); api('GET', '/quiz/packs?limit=200').then(function (d) { questionEditor({ id: r.questionId }, d.items || [], function () {}); });
    }, text: t('rp.editQuestion') })),
    field(t('rp.resolution'), resolution), same, disable, msg);
  var act = function (status) {
    return button({ label: t('rp.do.' + status), variant: status === 'resolved' ? 'primary' : 'ghost', data: { 'quiz-action': 'review-' + status }, onclick: function () {
      clear(msg);
      api('PUT', '/quiz/reports/' + r.id, { status: status, resolution: resolution.value.trim(), sameQuestion: same.input.checked, disableQuestion: disable.input.checked }).then(function () {
        closePanel(); toast('success', t('saved')); done();
      }, function (err) { msg.appendChild(notice('error', qzErrText(err))); });
    } });
  };
  sidePanel(t('rp.review'), body, [button({ label: t('cancel'), onclick: closePanel }), r.status !== 'open' ? act('open') : null, act('dismissed'), act('resolved')].filter(Boolean), 'review');
}

function reasonsPanel() {
  var body = h('div', { 'data-quiz-form': 'reasons' }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  sidePanel(t('rp.reasons'), body, [button({ label: t('close'), onclick: closePanel })], 'reasons');
  function paint() {
    Promise.all([api('GET', '/quiz/reasons'), loadLanguages()]).then(function (res) {
      clear(body);
      body.appendChild(h('p', { class: 'text-sm text-secondary-500 mb-3', text: t('rp.reasonsHint') }));
      (res[0].items || []).forEach(function (r) { body.appendChild(reasonRow(r, paint)); });
      body.appendChild(h('div', { class: 'h-3' }));
      body.appendChild(reasonRow(null, paint));
    }, function (err) { clear(body).appendChild(notice('error', qzErrText(err))); });
  }
  paint();
}

function reasonRow(r, done) {
  var creating = !r;
  r = r || { key: '', name: {}, sort: 100, isActive: true };
  var key = input(r.key, { name: 'reason-key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = translated(t('f.name'), 'reason-name', r.name);
  var sort = numInput(r.sort, 'reason-sort');
  var active = check(r.isActive, t('f.active'), 'reason-active');
  var msg = h('div', {});
  return h('div', { class: 'qz-track mb-2', 'data-quiz-reason': r.key || 'new' },
    creating ? h('p', { class: 'font-medium mb-2', text: t('rp.newReason') }) : null,
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.key'), key), field(t('f.sort'), sort)), name, active, msg,
    h('div', { class: 'flex gap-2 justify-end' },
      creating ? null : button({ label: t('delete'), variant: 'danger', onclick: function () {
        api('DELETE', '/quiz/reasons/' + r.id).then(done, function (err) { msg.appendChild(notice('error', err.code === 'IN_USE' ? t('rp.reasonInUse') : qzErrText(err))); });
      } }),
      button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-reason' }, onclick: function () {
        clear(msg);
        api(creating ? 'POST' : 'PUT', '/quiz/reasons' + (creating ? '' : '/' + r.id), { key: key.value.trim(), name: name.read(), sort: numberValue(sort) || 0, isActive: active.input.checked }).then(function () {
          toast('success', t('saved')); done();
        }, function (err) { msg.appendChild(notice('error', qzErrText(err))); });
      } })));
}
