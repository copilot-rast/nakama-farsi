function passText(p) {
  if (p.passCorrect != null) return t('pk.passCorrect', { n: p.passCorrect });
  if (p.passPercent != null) return t('pk.passPercent', { n: p.passPercent });
  return t('default');
}

function packsPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 50, total: 0, missionId: q.missionId || '', q: '' };
  var missions = [];
  var frame = pageFrame('packs', t('desc.packs'), [
    button({ label: t('pk.new'), variant: 'primary', data: { 'quiz-action': 'new-pack' }, onclick: function () { packEditor(null, missions, reload); } })]);
  clear(root).appendChild(frame.el);
  var missionSel = h('select', { class: QZ_INPUT + ' eco-select', name: 'missionId', 'aria-label': t('f.mission') });
  var search = input('', { name: 'q', placeholder: t('search') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx qz-filters' }, field(t('f.mission'), missionSel), field(t('search'), search)));
  frame.content.appendChild(listBox);
  missionSel.addEventListener('change', function () { state.missionId = missionSel.value; state.page = 1; reload(); });
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.q = search.value.trim(); state.page = 1; reload(); }, 300); });

  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/quiz/missions?kind=mission'), api('GET', '/quiz/packs' + qs({ missionId: state.missionId, q: state.q, page: state.page, limit: state.limit })), loadLanguages(), loadRefs(true), loadQzSettings()]).then(function (res) {
      missions = res[0].items || [];
      clear(missionSel);
      [{ value: '', label: t('all') }, { value: 'none', label: t('pk.noMission') }].concat(missions.map(function (m) { return { value: m.id, label: textOf(m.name, m.key) }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.missionId) opt.selected = true;
        missionSel.appendChild(opt);
      });
      var d = res[1];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-quiz-list': 'packs' }, table([
        { label: t('f.name') }, { label: t('f.mission') }, { label: t('pk.questions') }, { label: t('pk.pass') }, { label: t('pk.time') }, { label: t('f.reward') }, { label: '' }
      ], items.map(function (p) {
        return { id: p.key, item: p, cells: [
          h('div', { class: 'flex flex-col' }, h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(p.name, p.key) }), p.isActive ? null : tag(t('inactive'), 'warning'),
            p.campaignStageKey ? tag(t('pk.stageTag', { k: p.campaignStageKey }), 'primary') : null), mono(p.key)),
          h('span', { text: p.missionKey || t('none') }),
          h('a', { class: 'eco-link', href: '#/quiz/questions?packId=' + p.id, 'data-quiz-questions-link': p.key, text: p.poolCategory ? t('pk.fromBank', { n: fmtNum(p.questions), c: p.poolCategory }) : fmtNum(p.questions) }),
          h('span', { text: passText(p) + (p.questionCount ? ' · ' + t('pk.asked', { n: p.questionCount }) : '') }),
          h('span', { text: (p.questionSeconds ? t('pk.perQuestion', { n: p.questionSeconds }) : '') + (p.timeSeconds ? ' ' + t('pk.whole', { n: p.timeSeconds }) : '') || t('none') }),
          h('span', { class: 'text-xs', text: rewardText(p.reward) }),
          button({ label: t('delete'), variant: 'danger', data: { 'quiz-delete': p.key }, onclick: function () { deletePack(p, reload); } })
        ] };
      }), function (p) { packEditor(p, missions, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', qzErrText(err))); });
  }
  reload();
}

function deletePack(p, done) {
  confirmBox(t('pk.delete', { key: p.key }), t('delete')).then(function (ok) {
    if (!ok) return;
    api('DELETE', '/quiz/packs/' + p.id).then(function () { toast('success', t('saved')); done(); }, function (err) {
      if (err.code !== 'IN_USE') { toast('error', qzErrText(err)); return; }
      confirmBox(t('pk.inUse'), t('delete')).then(function (yes) {
        if (yes) api('DELETE', '/quiz/packs/' + p.id + '?force=true').then(function () { toast('success', t('saved')); done(); }, function (e2) { toast('error', qzErrText(e2)); });
      });
    });
  });
}

function packEditor(p, missions, done) {
  var creating = !p;
  var show = function (stats) {
    p = p || { name: {}, description: {}, isActive: true, reward: {}, tags: [], metadata: {} };
    var key = input(p.key || '', { name: 'key', ltr: true, mono: true });
    if (!creating) key.disabled = true;
    var mission = select(p.missionId || '', [{ value: '', label: t('pk.noMission') }].concat(missions.map(function (m) { return { value: m.id, label: textOf(m.name, m.key) }; })), { name: 'missionId', label: t('f.mission') });
    var name = translated(t('f.name'), 'name', p.name);
    var desc = translated(t('f.description'), 'description', p.description, { multiline: true });
    var category = input(p.category, { name: 'category' });
    var tags = tagsInput(p.tags, 'tags');
    var lang = input(p.language, { name: 'language', ltr: true, placeholder: 'fa' });
    var difficulty = numInput(p.difficulty || '', 'difficulty');
    var icon = input(p.iconUrl, { name: 'iconUrl', ltr: true });
    var sort = numInput(p.sort || 0, 'sort');
    var time = numInput(p.timeSeconds || '', 'timeSeconds');
    var qtime = numInput(p.questionSeconds || '', 'questionSeconds');
    var pct = numInput(p.passPercent == null ? '' : p.passPercent, 'passPercent');
    var correct = numInput(p.passCorrect == null ? '' : p.passCorrect, 'passCorrect');
    var count = numInput(p.questionCount || '', 'questionCount');
    var pool = input(p.poolCategory, { name: 'poolCategory' });
    var shuffleQ = triSelect(p.shuffleQuestions, 'shuffleQuestions');
    var shuffleO = triSelect(p.shuffleOptions, 'shuffleOptions');
    var send = check(p.sendAnswers, t('pk.sendAnswers'), 'sendAnswers');
    var reward = rewardEditor(p.reward, 'pack');
    var stage = refSelect('stages', p.campaignStageKey, 'campaignStageKey', t('pk.noStage'));
    var active = check(p.isActive, t('f.active'), 'isActive');
    var meta = jsonInput(p.metadata, 'metadata');
    var msg = h('div', {});
    var cfg = qzSettingsCache && qzSettingsCache.settings;
    var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
    var statsBox = stats ? h('div', { class: 'flex flex-wrap gap-2 mb-4', 'data-quiz-pack-stats': '' },
      tag(t('pk.s.players', { n: fmtNum(stats.players) })), tag(t('pk.s.plays', { n: fmtNum(stats.plays) })), tag(t('pk.s.wins', { n: fmtNum(stats.wins) }), 'success'),
      tag(t('pk.s.rate', { n: fmtNum(stats.winRate) })), tag(t('pk.s.best', { n: fmtNum(stats.avgBestScore) }))) : null;
    var body = h('div', { 'data-quiz-form': 'pack' }, statsBox,
      grid(field(t('f.key'), key, t('hint.key')), field(t('f.mission'), mission)), name, desc,
      grid(field(t('f.category'), category), field(t('f.tags'), tags)),
      grid(field(t('f.language'), lang, t('hint.language')), field(t('f.difficulty'), difficulty)),
      grid(field(t('f.icon'), icon), field(t('f.sort'), sort)), active,
      section(t('pk.play'), h('div', {},
        grid(field(t('pk.count'), count, t('hint.count')), field(t('pk.pool'), pool, t('hint.pool'))),
        grid(field(t('pk.qtime'), qtime, cfg && cfg.timing.questionSeconds ? t('hint.qtimeDefault', { n: cfg.timing.questionSeconds }) : t('hint.qtime')), field(t('pk.ptime'), time, t('hint.ptime'))),
        grid(field(t('pk.shuffleQ'), shuffleQ), field(t('pk.shuffleO'), shuffleO)), send, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-2', text: t('hint.sendAnswers') })), 'play'),
      h('div', { class: 'h-3' }), section(t('pk.pass'), h('div', {}, grid(field(t('pk.passCorrectF'), correct, t('hint.passCorrect')), field(t('pk.passPercentF'), pct, cfg ? t('hint.passPercent', { n: cfg.scoring.passPercent }) : null))), 'pass'),
      h('div', { class: 'h-3' }), section(t('pk.reward'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('hint.packReward') }), reward), 'reward'),
      h('div', { class: 'h-3' }), section(t('pk.campaign'), h('div', {}, field(t('pk.stage'), stage, t('hint.stage'))), 'campaign'),
      h('div', { class: 'h-3' }), field(t('f.metadata'), meta), msg);
    var save = button({ label: t('save'), variant: 'primary', data: { 'quiz-action': 'save-pack' }, onclick: function () {
      busy(save, true);
      clear(msg);
      var md = meta.read();
      if (md === undefined) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_JSON') + ' · metadata')); return; }
      var out = { key: key.value.trim(), missionId: mission.value, name: name.read(), description: desc.read(), category: category.value.trim(), tags: tags.read(), language: lang.value.trim(),
        difficulty: numberValue(difficulty) || 0, iconUrl: icon.value.trim(), sort: numberValue(sort) || 0, timeSeconds: numberValue(time) || 0, questionSeconds: numberValue(qtime) || 0,
        passPercent: optNumber(pct), passCorrect: optNumber(correct), questionCount: numberValue(count) || 0, poolCategory: pool.value.trim(), shuffleQuestions: shuffleQ.read(),
        shuffleOptions: shuffleO.read(), sendAnswers: send.input.checked, reward: reward.read(), campaignStageKey: stage.value, isActive: active.input.checked, metadata: md };
      api(creating ? 'POST' : 'PUT', '/quiz/packs' + (creating ? '' : '/' + p.id), out).then(function () {
        closePanel(); toast('success', t('saved')); qzRefs = null; done();
      }, function (err) { busy(save, false); msg.appendChild(notice('error', qzErrText(err))); });
    } });
    var footer = [button({ label: t('cancel'), onclick: closePanel }), save];
    if (!creating) footer.unshift(h('a', { class: QZ_BTN + ' qz-btn qz-btn--ghost', href: '#/quiz/questions?packId=' + p.id, onclick: closePanel, text: t('pk.openQuestions') }));
    sidePanel(creating ? t('pk.new') : textOf(p.name, p.key), body, footer, 'pack');
  };
  if (creating) { show(null); return; }
  api('GET', '/quiz/packs/' + p.id).then(function (d) { p = d.pack; show(d.stats); }, function (err) { toast('error', qzErrText(err)); });
}
