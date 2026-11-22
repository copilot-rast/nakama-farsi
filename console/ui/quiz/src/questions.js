function questionsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 50, total: 0, packId: q.packId || '', type: '', category: '', q: '', reported: q.reported || '', active: '' };
  var packs = [];
  var frame = pageFrame('questions', t('desc.questions'), [
    button({ label: t('qs.new'), variant: 'primary', data: { 'quiz-action': 'new-question' }, onclick: function () { questionEditor(null, packs, reload, state.packId); } })]);
  clear(root).appendChild(frame.el);
  var packSel = h('select', { class: QZ_INPUT + ' eco-select', name: 'packId', 'aria-label': t('f.pack') });
  var typeSel = select('', [{ value: '', label: t('all') }, { value: 'options', label: t('qs.type.options') }, { value: 'text', label: t('qs.type.text') }], { name: 'type', label: t('qs.type') });
  var catSel = h('select', { class: QZ_INPUT + ' eco-select', name: 'category', 'aria-label': t('f.category') });
  var activeSel = select('', [{ value: '', label: t('all') }, { value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }], { name: 'active', label: t('f.active') });
  var reported = check(state.reported === 'true', t('qs.reportedOnly'), 'reported');
  var search = input('', { name: 'q', placeholder: t('search') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx qz-filters' }, field(t('f.pack'), packSel), field(t('qs.type'), typeSel), field(t('f.category'), catSel),
    field(t('f.active'), activeSel), field(t('search'), search), reported));
  frame.content.appendChild(listBox);
  packSel.addEventListener('change', function () { state.packId = packSel.value; state.page = 1; reload(); });
  typeSel.addEventListener('change', function () { state.type = typeSel.value; state.page = 1; reload(); });
  catSel.addEventListener('change', function () { state.category = catSel.value; state.page = 1; reload(); });
  activeSel.addEventListener('change', function () { state.active = activeSel.value; state.page = 1; reload(); });
  reported.input.addEventListener('change', function () { state.reported = reported.input.checked ? 'true' : ''; state.page = 1; reload(); });
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.q = search.value.trim(); state.page = 1; reload(); }, 300); });

  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/packs?limit=200'), api('GET', '/quiz/questions' + qs({ packId: state.packId, type: state.type, category: state.category, active: state.active, reported: state.reported, q: state.q, page: state.page, limit: state.limit })),
      loadLanguages(), loadRefs(true)]).then(function (res) {
      packs = res[0].items || [];
      clear(packSel);
      [{ value: '', label: t('all') }, { value: 'none', label: t('qs.bank') }].concat(packs.map(function (p) { return { value: p.id, label: textOf(p.name, p.key) }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.packId) opt.selected = true;
        packSel.appendChild(opt);
      });
      clear(catSel);
      [{ value: '', label: t('all') }].concat((qzRefs.categories || []).map(function (c) { return { value: c.key, label: c.key }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.category) opt.selected = true;
        catSel.appendChild(opt);
      });
      var d = res[1];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'questions' }, table([
        { label: t('qs.question') }, { label: t('f.pack') }, { label: t('qs.type') }, { label: t('f.difficulty') }, { label: t('qs.answer') }, { label: t('qs.hints') }, { label: '' }
      ], items.map(function (x) {
        var right = (x.options || []).filter(function (o) { return o.correct; }).map(function (o) { return textOf(o.text, '#' + o.id); });
        return { id: x.id, item: x, cells: [
          h('div', { class: 'flex flex-col gap-1' }, h('span', { class: 'font-medium qz-sentence', dir: 'auto', text: textOf(x.question, x.mediaUrl || x.id) }),
            h('div', { class: 'flex flex-wrap gap-1' }, x.key ? mono(x.key) : null, x.isActive ? null : tag(t('inactive'), 'warning'), x.openReports ? tag(t('qs.reports', { n: x.openReports }), 'error') : null,
              x.category ? tag(x.category) : null)),
          h('span', { text: x.packKey || t('qs.bank') }),
          h('span', { text: t('qs.type.' + x.type) }),
          h('span', { text: fmtNum(x.difficulty) }),
          h('span', { class: 'text-xs', dir: 'auto', text: x.type === 'options' ? right.join('، ') : Object.keys(x.answers || {}).map(function (l) { return (x.answers[l] || []).join(' | '); }).join(' · ') }),
          h('span', { text: fmtNum(x.hints) }),
          button({ label: t('delete'), variant: 'danger', data: { 'quiz-delete': x.id }, onclick: function () { deleteQuestion(x, reload); } })
        ] };
      }), function (x) { questionEditor(x, packs, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function deleteQuestion(x, done) {
  confirmBox(t('qs.delete'), t('delete')).then(function (ok) {
    if (!ok) return;
    api('DELETE', '/quiz/questions/' + x.id).then(function () { toast('success', t('saved')); done(); }, function (err) {
      if (err.code !== 'IN_USE') { toast('error', qzErrText(err)); return; }
      confirmBox(t('qs.inUse'), t('delete')).then(function (yes) {
        if (yes) api('DELETE', '/quiz/questions/' + x.id + '?force=true').then(function () { toast('success', t('saved')); done(); }, function (e2) { toast('error', qzErrText(e2)); });
      });
    });
  });
}

function optionEditor(list) {
  var box = h('div', { class: 'qz-rows', 'data-quiz-options': '' });
  var rows = [];
  var nextId = 0;
  (list || []).forEach(function (o) { if (o.id >= nextId) nextId = o.id + 1; });
  function addRow(o) {
    o = o || { id: nextId++, text: {}, correct: false };
    var texts = {};
    var right = check(o.correct, t('qs.correct'), 'option-correct');
    var media = input(o.mediaUrl, { name: 'option-media', ltr: true, placeholder: t('qs.media') });
    var inputs = (qzLanguages || []).map(function (l) {
      var el = input(o.text && o.text[l.code] || '', { name: 'option-' + l.code, label: l.name });
      el.setAttribute('dir', l.direction === 'rtl' ? 'rtl' : 'ltr');
      texts[l.code] = el;
      return field(t('lang.text', { label: t('qs.option'), lang: l.name }), el);
    });
    var row = h('div', { class: 'qz-track', 'data-quiz-option': String(o.id) },
      h('div', { class: 'flex items-center justify-between gap-2 mb-2' }, h('span', { class: 'text-xs text-secondary-500', text: t('qs.optionN', { n: rows.length + 1 }) }), right,
        button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r.row !== row; }); } })),
      h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, inputs), field(t('qs.media'), media));
    rows.push({ row: row, id: o.id, texts: texts, right: right, media: media });
    box.appendChild(row);
  }
  (list || []).forEach(addRow);
  var wrap = h('div', {}, box, button({ label: '+ ' + t('qs.addOption'), data: { 'quiz-action': 'add-option' }, onclick: function () { addRow(); } }));
  wrap.add = addRow;
  wrap.count = function () { return rows.length; };
  wrap.read = function () {
    return rows.map(function (r) {
      var txt = {};
      Object.keys(r.texts).forEach(function (k) { var v = r.texts[k].value.trim(); if (v) txt[k] = v; });
      return { id: r.id, text: txt, mediaUrl: r.media.value.trim(), correct: r.right.input.checked };
    });
  };
  return wrap;
}

function answerEditor(value) {
  var inputs = {};
  var box = h('div', { class: 'grid sm:grid-cols-2 gap-x-4', 'data-quiz-answers': '' });
  (qzLanguages || []).forEach(function (l) {
    var el = input(((value || {})[l.code] || []).join('\n'), { name: 'answers-' + l.code, multiline: true, rows: 3 });
    el.setAttribute('dir', l.direction === 'rtl' ? 'rtl' : 'ltr');
    inputs[l.code] = el;
    box.appendChild(field(t('lang.text', { label: t('qs.accepted'), lang: l.name }), el, t('hint.accepted')));
  });
  box.read = function () {
    var out = {};
    Object.keys(inputs).forEach(function (k) {
      var list = inputs[k].value.split('\n').map(function (s) { return s.trim(); }).filter(Boolean);
      if (list.length) out[k] = list;
    });
    return out;
  };
  return box;
}

function questionEditor(x, packs, done, packId) {
  var creating = !x;
  var show = function (stats) {
    x = x || { type: 'options', question: {}, explanation: {}, isActive: true, reward: {}, tags: [], packId: packId && packId !== 'none' ? packId : '', metadata: {},
      options: [{ id: 0, text: {}, correct: true }, { id: 1, text: {} }, { id: 2, text: {} }, { id: 3, text: {} }] };
    var key = input(x.key || '', { name: 'key', ltr: true, mono: true });
    var pack = select(x.packId || '', [{ value: '', label: t('qs.bank') }].concat(packs.map(function (p) { return { value: p.id, label: textOf(p.name, p.key) }; })), { name: 'packId', label: t('f.pack') });
    var type = select(x.type, [{ value: 'options', label: t('qs.type.options') }, { value: 'text', label: t('qs.type.text') }], { name: 'type', label: t('qs.type') });
    var text = translated(t('qs.question'), 'question', x.question, { multiline: true });
    var media = input(x.mediaUrl, { name: 'mediaUrl', ltr: true });
    var opts = optionEditor(x.options);
    var answers = answerEditor(x.answers);
    var expl = translated(t('qs.explanation'), 'explanation', x.explanation, { multiline: true });
    var difficulty = numInput(x.difficulty || '', 'difficulty');
    var category = input(x.category, { name: 'category' });
    var tags = tagsInput(x.tags, 'tags');
    var time = numInput(x.timeSeconds == null ? '' : x.timeSeconds, 'timeSeconds');
    var sort = numInput(x.sort || 0, 'sort');
    var active = check(x.isActive, t('f.active'), 'isActive');
    var reward = rewardEditor(x.reward, 'question');
    var meta = jsonInput(x.metadata, 'metadata');
    var optBox = section(t('qs.options'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.options') }), opts), 'options');
    var ansBox = section(t('qs.accepted'), answers, 'answers');
    var hintsBox = h('div', {});
    var msg = h('div', {});
    var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
    function paint() { optBox.hidden = type.value !== 'options'; ansBox.hidden = type.value !== 'text'; }
    type.addEventListener('change', paint);
    var statsBox = stats ? h('div', { class: 'flex flex-wrap gap-2 mb-4', 'data-quiz-question-stats': '' }, tag(t('qs.s.answers', { n: fmtNum(stats.answers) })),
      tag(t('qs.s.correct', { n: stats.answers ? fmtNum(Math.round(stats.correct * 1000 / stats.answers) / 10) : 0 }), 'success'), x.openReports ? tag(t('qs.reports', { n: x.openReports }), 'error') : null) : null;
    var body = h('div', { 'data-quiz-form': 'question' }, statsBox,
      grid(field(t('f.pack'), pack, t('hint.bank')), field(t('qs.type'), type)), text, field(t('qs.mediaF'), media, t('hint.media')),
      optBox, ansBox, h('div', { class: 'h-3' }), expl,
      grid(field(t('f.difficulty'), difficulty), field(t('qs.time'), time, t('hint.qtimeOwn'))),
      grid(field(t('f.category'), category, t('hint.category')), field(t('f.tags'), tags)),
      grid(field(t('f.key'), key, t('hint.qkey')), field(t('f.sort'), sort)), active,
      section(t('qs.reward'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.questionReward') }), reward), 'reward'),
      h('div', { class: 'h-3' }), hintsBox, h('div', { class: 'h-3' }), field(t('f.metadata'), meta), msg);
    paint();
    if (!creating) hintList(x.id, hintsBox);
    else hintsBox.appendChild(h('p', { class: 'text-xs text-secondary-500', text: t('hn.afterSave') }));
    var save = button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-question' }, onclick: function () {
      busy(save, true);
      clear(msg);
      var md = meta.read();
      if (md === undefined) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_JSON'))); return; }
      var tsec = numberValue(time);
      var out = { key: key.value.trim(), packId: pack.value, type: type.value, question: text.read(), mediaUrl: media.value.trim(), options: type.value === 'options' ? opts.read() : [],
        answers: type.value === 'text' ? answers.read() : {}, explanation: expl.read(), difficulty: numberValue(difficulty) || 0, category: category.value.trim(), tags: tags.read(),
        timeSeconds: tsec == null || isNaN(tsec) ? null : tsec, sort: numberValue(sort) || 0, isActive: active.input.checked, reward: reward.read(), metadata: md };
      api(creating ? 'POST' : 'PUT', '/quiz/questions' + (creating ? '' : '/' + x.id), out).then(function (d) {
        toast('success', t('saved'));
        done();
        if (creating) { closePanel(); questionEditor(d.question, packs, done); } else closePanel();
      }, function (err) { busy(save, false); msg.appendChild(notice('error', qzErrText(err))); });
    } });
    sidePanel(creating ? t('qs.new') : t('qs.edit'), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'question');
  };
  if (creating) { show(null); return; }
  api('GET', '/quiz/questions/' + x.id).then(function (d) { x = d.question; show(d.stats); }, function (err) { toast('error', qzErrText(err)); });
}

function hintList(questionId, box) {
  clear(box).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  api('GET', '/quiz/questions/' + questionId + '/hints').then(function (d) {
    var list = h('div', { class: 'flex flex-col gap-2', 'data-quiz-hints': '' });
    (d.items || []).forEach(function (hn) {
      list.appendChild(h('div', { class: 'qz-track flex flex-wrap items-center gap-2', 'data-quiz-hint': hn.id },
        h('span', { class: 'font-medium', text: textOf(hn.name, '') }), hn.isActive ? null : tag(t('inactive'), 'warning'),
        tag(hn.price && hn.price.amount ? '⁦' + fmtNum(hn.price.amount) + ' ' + hn.price.currencyKey + '⁩' : t('hn.free'), 'primary'),
        h('span', { class: 'text-xs text-secondary-500', text: t('hn.bought', { n: fmtNum(hn.bought) }) }), h('div', { class: 'flex-1' }),
        button({ label: t('edit'), data: { 'quiz-hint-edit': hn.id }, onclick: function () { hintForm(questionId, hn, box); } }),
        button({ label: t('delete'), variant: 'danger', onclick: function () {
          confirmBox(t('hn.delete'), t('delete')).then(function (ok) {
            if (ok) api('DELETE', '/quiz/hints/' + hn.id).then(function () { hintList(questionId, box); }, function (err) { toast('error', qzErrText(err)); });
          });
        } })));
    });
    clear(box).appendChild(section(t('hn.title'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.hints') }), list,
      h('div', { class: 'mt-2' }, button({ label: '+ ' + t('hn.add'), data: { 'quiz-action': 'add-hint' }, onclick: function () { hintForm(questionId, null, box); } }))), 'hints'));
  }, function (err) { clear(box).appendChild(notice('error', qzErrText(err))); });
}

function hintForm(questionId, hn, box) {
  var creating = !hn;
  hn = hn || { name: {}, text: {}, price: {}, isActive: true };
  var name = translated(t('f.name'), 'hint-name', hn.name);
  var text = translated(t('hn.text'), 'hint-text', hn.text, { multiline: true });
  var media = input(hn.mediaUrl, { name: 'hint-media', ltr: true });
  var cur = refSelect('currencies', (hn.price || {}).currencyKey, 'hint-currency', t('hn.free'));
  var amount = numInput((hn.price || {}).amount || '', 'hint-amount');
  var active = check(hn.isActive, t('f.active'), 'hint-active');
  var msg = h('div', {});
  var form = h('div', { class: 'qz-track', 'data-quiz-form': 'hint' }, name, text, field(t('qs.media'), media),
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('hn.currency'), cur), field(t('hn.amount'), amount)), active, msg,
    h('div', { class: 'flex gap-2 justify-end' }, button({ label: t('cancel'), onclick: function () { hintList(questionId, box); } }),
      button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-hint' }, onclick: function () {
        clear(msg);
        var out = { name: name.read(), text: text.read(), mediaUrl: media.value.trim(), price: { currencyKey: cur.value, amount: cur.value ? numberValue(amount) || 0 : 0 }, isActive: active.input.checked };
        api(creating ? 'POST' : 'PUT', creating ? '/quiz/questions/' + questionId + '/hints' : '/quiz/hints/' + hn.id, out).then(function () { toast('success', t('saved')); hintList(questionId, box); },
          function (err) { msg.appendChild(notice('error', qzErrText(err))); });
      } })));
  clear(box).appendChild(section(t('hn.title'), form, 'hints'));
}
