var cgSettingsCache = null;
function loadCgSettings() {
  if (cgSettingsCache) return Promise.resolve(cgSettingsCache);
  return api('GET', '/campaign/settings').then(function (d) { cgSettingsCache = d; return d; });
}

function stagesPage(root) {
  var q = hashQuery();
  var state = { page: 1, limit: 50, total: 0, chapterId: q.chapterId || '', q: q.q || '', mode: '' };
  var chapters = [];
  var frame = pageFrame('stages', t('desc.stages'), [button({ label: t('st.new'), variant: 'primary', data: { 'campaign-action': 'new-stage' }, onclick: function () { stageEditor(null, chapters, reload, state.chapterId); } })]);
  clear(root).appendChild(frame.el);
  var chapterSel = h('select', { class: CG_INPUT + ' eco-select', name: 'chapterId', 'aria-label': t('f.chapter') });
  var modeSel = select('', [{ value: '', label: t('all') }].concat(['ordered', 'repeatable', 'once'].map(function (m) { return { value: m, label: t('mode.' + m) }; })), { name: 'mode', label: t('f.mode') });
  var search = input(state.q, { name: 'q', placeholder: t('search') });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx cg-filters' }, field(t('f.chapter'), chapterSel), field(t('f.mode'), modeSel), field(t('search'), search)));
  frame.content.appendChild(listBox);
  chapterSel.addEventListener('change', function () { state.chapterId = chapterSel.value; state.page = 1; reload(); });
  modeSel.addEventListener('change', function () { state.mode = modeSel.value; state.page = 1; reload(); });
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.q = search.value.trim(); state.page = 1; reload(); }, 300); });

  function reload() {
    clear(listBox).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/campaign/chapters'), api('GET', '/campaign/stages' + qs({ chapterId: state.chapterId, mode: state.mode, q: state.q, page: state.page, limit: state.limit })), loadLanguages(), loadRefs(true), loadCgSettings()]).then(function (res) {
      chapters = res[0].items || [];
      clear(chapterSel);
      [{ value: '', label: t('st.filter.chapter') }].concat(chapters.map(function (c) { return { value: c.id, label: textOf(c.name, c.key) }; })).forEach(function (o) {
        var opt = h('option', { value: o.value, text: o.label });
        if (o.value === state.chapterId) opt.selected = true;
        chapterSel.appendChild(opt);
      });
      var d = res[1];
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      var max = (cgSettingsCache && cgSettingsCache.settings.stars.maxStars) || 3;
      var move = function (i, dir) {
        var same = items.filter(function (x) { return x.chapterId === items[i].chapterId; });
        var ids = same.map(function (x) { return x.id; });
        var at = ids.indexOf(items[i].id);
        var j = at + dir;
        if (j < 0 || j >= ids.length) return;
        var x = ids[at]; ids[at] = ids[j]; ids[j] = x;
        api('PUT', '/campaign/stages/order', { ids: ids }).then(reload, function (err) { toast('error', cgErrText(err)); });
      };
      listBox.appendChild(h('div', { 'data-campaign-list': 'stages' }, table([
        { label: t('f.sort') }, { label: t('f.name') }, { label: t('f.chapter') }, { label: t('f.mode') }, { label: t('st.rewards') }, { label: t('f.timeLimit') }, { label: t('rp.players') }, { label: '' }
      ], items.map(function (s, i) {
        return { id: s.key, item: s, cells: [
          state.chapterId ? h('div', { class: 'flex gap-1' }, button({ label: '↑', title: t('st.moveUp'), data: { 'campaign-up': s.key }, onclick: function () { move(i, -1); } }),
            button({ label: '↓', title: t('st.moveDown'), data: { 'campaign-down': s.key }, onclick: function () { move(i, 1); } })) : h('span', { text: digits(String(s.sort)) }),
          h('div', { class: 'flex flex-col' }, h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(s.name, s.key) }), s.isActive ? null : tag(t('inactive'), 'warning')), mono(s.key)),
          h('span', { text: s.chapterKey }),
          modeTag(s.mode),
          h('span', { class: 'text-xs', 'data-campaign-reward': s.key, text: rewardText((s.rewards || {}).currencies) + (s.xp ? ' · \u2066' + fmtNum(s.xp) + ' XP\u2069' : '') }),
          h('span', { text: s.timeLimitSeconds ? fmtNum(s.timeLimitSeconds) : t('none') }),
          h('span', { text: fmtNum(s.players) }),
          button({ label: t('delete'), variant: 'danger', data: { 'campaign-delete': s.key }, onclick: function () {
            confirmBox(t('st.delete', { key: s.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/campaign/stages/' + s.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', err.code === 'IN_USE' ? t('st.inUse') : cgErrText(err)); });
            });
          } })
        ] };
      }), function (s) { stageEditor(s, chapters, reload); })));
      if (!items.length) listBox.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') })));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
      void max;
    }, function (err) { clear(listBox).appendChild(notice('error', cgErrText(err))); });
  }
  reload();
}

function previewLines(cur, xp, mult, max) {
  var lines = [];
  for (var n = 1; n <= max; n++) {
    var m = mult[n - 1] || 0;
    var parts = Object.keys(cur || {}).map(function (k) { return fmtNum(Math.floor(cur[k] * m + 1e-9)) + ' ' + k; });
    if (xp) parts.push(fmtNum(Math.floor(xp * m + 1e-9)) + ' XP');
    lines.push(tr('st.preview.line', { n: n }, { pay: parts.length ? '\u2066' + parts.join(' + ') + '\u2069' : t('none') }));
  }
  return lines;
}

function stageEditor(s, chapters, done, chapterId) {
  var creating = !s;
  loadCgSettings().then(function (cfgRes) {
    var cfg = cfgRes.settings;
    var max = cfg.stars.maxStars;
    s = s || { name: {}, description: {}, mode: 'ordered', isActive: true, rewards: {}, tierPrizes: [], chapterId: chapterId || (chapters[0] && chapters[0].id) || '', metadata: {}, ownUnlock: false, ownStarRules: false, ownMultipliers: false };
    var key = input(s.key || '', { name: 'key', ltr: true, mono: true });
    if (!creating) key.disabled = true;
    var chapter = select(s.chapterId, chapters.map(function (c) { return { value: c.id, label: textOf(c.name, c.key) }; }), { name: 'chapterId', label: t('f.chapter') });
    var mode = select(s.mode, ['ordered', 'repeatable', 'once'].map(function (m) { return { value: m, label: t('mode.' + m) }; }), { name: 'mode', label: t('f.mode') });
    var modeHint = h('p', { class: 'text-xs text-secondary-500 -mt-3 mb-3', 'data-campaign-mode-hint': '' });
    var name = translated(t('f.name'), 'name', s.name);
    var desc = translated(t('f.description'), 'description', s.description, { multiline: true });
    var icon = input(s.iconUrl, { name: 'iconUrl', ltr: true });
    var address = input(s.address, { name: 'address', ltr: true, mono: true });
    var active = check(s.isActive, t('f.active'), 'isActive');
    var limit = numInput(s.timeLimitSeconds, 'timeLimitSeconds', t('f.timeLimit'));
    var minSec = numInput(s.minSeconds, 'minSeconds', t('f.minSeconds'));
    var ownUnlock = check(s.ownUnlock, t('own'), 'ownUnlock');
    var rules = ruleEditor(s.ownUnlock ? s.unlock : [], 'stage');
    var unlockBox = h('div', {}, ownUnlock, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('st.unlockDefault') }), rules);
    var ownStars = check(!s.ownStarRules, t('star.useDefault'), 'useDefaultStars');
    var templates = cfg.stars.templates || [];
    var template = select(s.starTemplate || '', [{ value: '', label: t('tpl.none') }].concat(templates.map(function (x) { return { value: x.key, label: textOf(x.name, x.key) + ' · ' + x.key }; })), { name: 'starTemplate', label: t('tpl.label') });
    var templateBox = h('div', {}, field(t('tpl.label'), template, t('tpl.hint')), h('ul', { class: 'text-sm flex flex-col gap-1 mb-3', 'data-campaign-template-rules': '' }));
    var noReplay = check(s.noReplayAfterWin, t('st.noReplay'), 'noReplayAfterWin');
    var scopeOf = function (id) { return cgScope(chapters.filter(function (x) { return x.id === id; })[0]); };
    var settingsHolder = h('div', { 'data-campaign-stage-settings': '' });
    var settingsForm = null;
    var stageDefs = (cfg.fields && cfg.fields.stage) || [];
    function paintSettings() {
      var keep = settingsForm ? settingsForm.read() : (s.settings || {});
      settingsForm = fieldsForm(stageDefs, keep, scopeOf(chapter.value));
      clear(settingsHolder).appendChild(settingsForm.empty ? h('p', { class: 'text-sm text-secondary-500', text: t('fd.none') }) : settingsForm);
    }
    chapter.addEventListener('change', paintSettings);
    paintSettings();
    var starRules = starEditor(s.ownStarRules ? s.starRules : cfg.stars.rules, max);
    var ownMult = check(!s.ownMultipliers, t('star.useDefault'), 'useDefaultMultipliers');
    var mult = multiplierInputs(s.ownMultipliers ? s.starMultipliers : (s.mode === 'repeatable' ? cfg.stars.repeatableMultipliers : cfg.stars.multipliers), max, 'mult');
    var cur = currencyRows((s.rewards || {}).currencies, 'stage');
    var xp = numInput(s.xp || '', 'xp', t('f.xp'));
    var tiers = h('div', { class: 'cg-rows', 'data-campaign-tiers': '' });
    var tierRows = [];
    function addTier(tp) {
      tp = tp || { stars: max };
      var st = numInput(tp.stars, 'tier-stars', t('star.stars'));
      var pickers = prizePickers(tp);
      var row = h('div', { class: 'cg-track', 'data-campaign-tier': '' }, h('div', { class: 'flex items-end gap-3' }, field(t('star.stars'), st, null, 'cg-num'), h('div', { class: 'flex-1' }),
        button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); tierRows = tierRows.filter(function (r) { return r.row !== row; }); } })), pickers);
      tierRows.push({ row: row, read: function () { var o = pickers.read(); o.stars = numberValue(st) || 0; return o; } });
      tiers.appendChild(row);
    }
    (s.tierPrizes || []).forEach(addTier);
    var group = input(s.rewardGroup, { name: 'rewardGroup', ltr: true, mono: true });
    var cooldown = numInput(s.cooldownSeconds || '', 'cooldownSeconds', t('f.cooldown'));
    var daily = numInput(s.dailyLimit || '', 'dailyLimit', t('f.daily'));
    var xpLimited = check(s.xpWhenLimited, t('f.xpWhenLimited'), 'xpWhenLimited');
    var meta = input(JSON.stringify(s.metadata || {}, null, 2), { name: 'metadata', ltr: true, mono: true, multiline: true, rows: 4 });
    var preview = h('ul', { class: 'text-sm flex flex-col gap-1', 'data-campaign-preview': '' });
    function paint() {
      modeHint.textContent = t('mode.hint.' + mode.value);
      rules.hidden = !ownUnlock.input.checked;
      starRules.hidden = ownStars.input.checked;
      templateBox.hidden = !ownStars.input.checked;
      var tplRules = template.value ? ((templates.filter(function (x) { return x.key === template.value; })[0] || {}).rules || []) : cfg.stars.rules;
      var tplList = templateBox.querySelector('[data-campaign-template-rules]');
      clear(tplList);
      tplRules.forEach(function (rr) { tplList.appendChild(h('li', { class: 'text-secondary-600 dark:text-secondary-300', text: starSentence(rr) })); });
      mult.hidden = false;
      mult.inputs.forEach(function (el) { el.disabled = ownMult.input.checked; });
      if (ownMult.input.checked) {
        var def = mode.value === 'repeatable' ? cfg.stars.repeatableMultipliers : cfg.stars.multipliers;
        mult.inputs.forEach(function (el, i) { el.value = digits(String(def[i])); });
      }
      clear(preview);
      previewLines(cur.read(), numberValue(xp) || 0, mult.read(), max).forEach(function (l) { preview.appendChild(h('li', { text: l })); });
      if (mode.value !== 'repeatable') preview.appendChild(h('li', { class: 'text-secondary-500', text: t('st.preview.replay') }));
    }
    var msg = h('div', {});
    var body = h('div', { 'data-campaign-form': 'stage' },
      h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.key'), key, t('hint.key')), field(t('f.chapter'), chapter)),
      field(t('f.mode'), mode), modeHint, name, desc,
      h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.icon'), icon), field(t('f.address'), address)), active,
      section(t('st.timing'), h('div', {}, h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.timeLimit'), limit), field(t('f.minSeconds'), minSec, t('hint.minSeconds'))), noReplay,
        h('p', { class: 'text-xs text-secondary-500 -mt-2', text: t('st.noReplayHint') })), 'timing'),
      h('div', { class: 'h-3' }), section(t('st.settings'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('st.settingsHint') }), settingsHolder), 'settings'),
      h('div', { class: 'h-3' }), section(t('st.unlock'), unlockBox, 'unlock'),
      h('div', { class: 'h-3' }), section(t('st.stars'), h('div', {}, ownStars, templateBox, starRules), 'stars'),
      h('div', { class: 'h-3' }), section(t('st.rewards'), h('div', {}, cur, h('div', { class: 'mt-3' }, field(t('f.xp'), xp, null, 'cg-num')),
        h('div', { class: 'text-sm font-medium mt-2 mb-1', text: t('st.multipliers') }), ownMult, mult,
        h('div', { class: 'text-sm font-medium mt-3 mb-1', text: t('st.preview') }), preview), 'rewards'),
      h('div', { class: 'h-3' }), section(t('st.tierPrizes'), h('div', {}, tiers, button({ label: '+ ' + t('st.addTier'), data: { 'campaign-action': 'add-tier' }, onclick: function () { addTier(); } })), 'tiers'),
      h('div', { class: 'h-3' }), section(t('st.limits'), h('div', {}, field(t('f.rewardGroup'), group, t('hint.rewardGroup')),
        h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.cooldown'), cooldown), field(t('f.daily'), daily)), xpLimited), 'limits'),
      h('div', { class: 'h-3' }), section(t('st.advanced'), field(t('f.metadata'), meta, t('hint.metadata')), 'advanced'), msg);
    body.addEventListener('input', paint);
    body.addEventListener('change', paint);
    body.addEventListener('click', function (e) { if (e.target.closest('button')) setTimeout(paint, 0); });
    var save = button({ label: t('save'), variant: 'primary', data: { 'campaign-action': 'save-stage' }, onclick: function () {
      busy(save, true);
      clear(msg);
      var md;
      try { md = JSON.parse(meta.value || '{}'); } catch (e) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_JSON') + ' · metadata')); return; }
      var lim = floatValue(limit), mn = floatValue(minSec);
      var badField = settingsForm.invalid();
      if (badField) { busy(save, false); msg.appendChild(notice('error', t('code.INVALID_FIELD_VALUE') + ' · ' + badField)); return; }
      var out = { key: key.value.trim(), chapterId: chapter.value, mode: mode.value, name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), address: address.value.trim(),
        sort: creating ? 1000 : s.sort, isActive: active.input.checked, timeLimitSeconds: lim == null ? null : lim, minSeconds: mn == null ? null : mn,
        unlock: ownUnlock.input.checked ? rules.read() : null, starRules: ownStars.input.checked ? null : starRules.read(), starMultipliers: ownMult.input.checked ? null : mult.read(),
        rewards: { currencies: cur.read() }, xp: numberValue(xp) || 0, tierPrizes: tierRows.map(function (r) { return r.read(); }),
        rewardGroup: group.value.trim(), cooldownSeconds: numberValue(cooldown) || 0, dailyLimit: numberValue(daily) || 0, xpWhenLimited: xpLimited.input.checked, metadata: md,
        settings: settingsForm.read(), starTemplate: ownStars.input.checked ? template.value : '', noReplayAfterWin: noReplay.input.checked };
      api(creating ? 'POST' : 'PUT', '/campaign/stages' + (creating ? '' : '/' + s.id), out).then(function () {
        closePanel(); toast('success', t('saved')); done();
      }, function (err) { busy(save, false); msg.appendChild(notice('error', cgErrText(err))); });
    } });
    sidePanel(creating ? t('st.new') : textOf(s.name, s.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'stage');
    paint();
  }, function (err) { toast('error', cgErrText(err)); });
}
