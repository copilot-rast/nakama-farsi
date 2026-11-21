function campaignSettingsPage(root) {
  Promise.all([loadRefs(true)]).then(function () {
    settingsPage(root, 'settings', '/campaign', function (cur, defs, d) {
      var opt = function (prefix, list) { return list.map(function (v) { return { value: v, label: t(prefix + v) }; }); };
      var notes = h('div', { class: 'flex flex-col gap-2' });
      if (d.economy === false) notes.appendChild(notice('warning', t('se.noEconomy')));
      if (d.xpMetaFound === false) notes.appendChild(notice('warning', t('se.metaMissing', { k: cur.xp.metaKey })));
      var pl = cur.play || defs.play || { costCurrency: '', costAmount: 1, refundOnWin: true };
      if (pl.costCurrency && d.costCurrencyFound === false) notes.appendChild(notice('warning', t('se.costMissing', { k: pl.costCurrency })));
      else if (pl.costCurrency && d.costCurrencyRegen === false) notes.appendChild(notice('info', t('se.costNoRegen', { k: pl.costCurrency })));
      var costCur = refSelect('currencies', pl.costCurrency, 'play.costCurrency', t('se.cost.none'));
      var costAmount = numInput(pl.costAmount, 'play.costAmount');
      var refundOnWin = check(pl.refundOnWin, t('se.refundOnWin'), 'play.refundOnWin');
      var a = cur.attempts, s = cur.stars, u = cur.unlock, r = cur.rewards, x = cur.xp, e = cur.events, f = cur.freeride;
      var hours = numInput(a.hours, 'attempts.hours'), grace = numInput(a.graceSeconds, 'attempts.graceSeconds'), keep = numInput(a.keepDays, 'attempts.keepDays'), minSec = numInput(a.defaultMinSeconds, 'attempts.defaultMinSeconds');
      var closeOthers = check(a.closeOthers, t('se.closeOthers'), 'attempts.closeOthers'), serverCloses = check(a.serverClosesOthers, t('se.serverClosesOthers'), 'attempts.serverClosesOthers');
      var maxStars = numInput(s.maxStars, 'stars.maxStars'), minOn = numInput(s.minOnSuccess, 'stars.minOnSuccess');
      var missing = select(s.missingField, opt('se.missing.', ['pass', 'fail']), { name: 'stars.missingField' });
      var mult = multiplierInputs(s.multipliers, s.maxStars, 'multipliers');
      var rmult = multiplierInputs(s.repeatableMultipliers, s.maxStars, 'repeatableMultipliers');
      var rules = starEditor(s.rules, s.maxStars);
      var stageDefault = select(u.stageDefault, opt('se.stageDefault.', ['all_previous', 'previous', 'none']), { name: 'unlock.stageDefault' });
      var chapterDefault = select(u.chapterDefault, opt('se.chapterDefault.', ['none', 'previous']), { name: 'unlock.chapterDefault' });
      var minStars = numInput(u.minStars, 'unlock.minStars');
      var showLocked = check(u.showLocked, t('se.showLocked'), 'unlock.showLocked');
      var replay = select(r.replayMode, opt('se.replay.', ['difference', 'none', 'full']), { name: 'rewards.replayMode' });
      var zone = input(r.dayZone, { name: 'rewards.dayZone', ltr: true });
      var groupMode = select(r.prizeGroupMode, opt('se.groupMode.', ['draw', 'all']), { name: 'rewards.prizeGroupMode' });
      var modes = ['ordered', 'repeatable', 'once'].map(function (m) { return check((r.milestoneModes || []).indexOf(m) !== -1, t('mode.' + m), 'milestone-' + m); });
      var metaKey = refSelect('metas', x.metaKey, 'xp.metaKey', t('se.metaKey.none'));
      var table_ = refSelect('levelTables', x.levelTableKey, 'xp.levelTableKey', t('se.levelTable.main'));
      var ecoEvent = input(e.economyEvent, { name: 'events.economyEvent', ltr: true, mono: true });
      var anEvent = input(e.analyticsEvent, { name: 'events.analyticsEvent', ltr: true, mono: true });
      var onFailure = check(e.onFailure, t('se.onFailure'), 'events.onFailure');
      var tableOn = check(f.tableEnabled, t('se.tableEnabled'), 'freeride.tableEnabled');
      var reqId = check(f.requireRequestId, t('se.requireRequestId'), 'freeride.requireRequestId');
      var tplBox = h('div', { class: 'flex flex-col gap-3', 'data-campaign-templates': '' });
      var tplItems = [];
      var addTpl = function (tp) {
        tp = tp || { key: '', name: {}, rules: [] };
        var k = input(tp.key, { name: 'template-key', ltr: true, mono: true, label: t('f.key') });
        var nm = translated(t('f.name'), 'template-name', tp.name);
        var ed = starEditor(tp.rules, s.maxStars);
        var el = h('div', { class: 'cg-track', 'data-campaign-template': tp.key || '' },
          h('div', { class: 'flex items-end gap-3' }, field(t('f.key'), k, null, 'flex-1'),
            button({ label: t('remove'), variant: 'danger', onclick: function () { el.remove(); tplItems = tplItems.filter(function (x) { return x.el !== el; }); } })), nm, ed);
        tplItems.push({ el: el, read: function () { return { key: k.value.trim(), name: nm.read(), rules: ed.read() }; } });
        tplBox.appendChild(el);
      };
      (s.templates || []).forEach(addTpl);
      var fieldsRaw = input(JSON.stringify(cur.fields || { stage: [], chapter: [] }, null, 2), { name: 'fields', ltr: true, mono: true, multiline: true, rows: 10 });
      var reqHours = numInput(f.requestHours, 'freeride.requestHours');
      var it = cur.items || {};
      var csv = function (list) { return (list || []).join(', '); };
      var splitCsv = function (el) { return el.value.split(/[,،]/).map(function (x) { return x.trim(); }).filter(Boolean); };
      var itEnforce = check(it.enforce, t('se.itemsEnforce'), 'items.enforce');
      var itRequire = check(it.requireItem, t('se.itemsRequire'), 'items.requireItem');
      var itOwned = check(it.requireOwned, t('se.itemsOwned'), 'items.requireOwned');
      var itParam = input(it.param || 'itemKey', { name: 'items.param', ltr: true });
      var itType = input(it.itemType || '', { name: 'items.itemType', ltr: true });
      var itClass = input(it.classField || '', { name: 'items.classField', ltr: true });
      var itAllow = input(csv(it.allow), { name: 'items.allow', ltr: true });
      var itDeny = input(csv(it.deny), { name: 'items.deny', ltr: true });
      var itClasses = input(csv(it.classes), { name: 'items.classes', ltr: true });
      var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
      return {
        el: [notes,
          section(t('se.play'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: t('se.play.hint') }),
            grid(field(t('se.costCurrency'), costCur), field(t('se.costAmount'), costAmount)), refundOnWin), 'play'),
          section(t('se.attempts'), h('div', {}, grid(field(t('se.hours'), hours), field(t('se.grace'), grace)), grid(field(t('se.keepDays'), keep), field(t('se.minSeconds'), minSec)), closeOthers, serverCloses), 'attempts'),
          section(t('se.stars'), h('div', {}, grid(field(t('se.maxStars'), maxStars), field(t('se.minOnSuccess'), minOn)), field(t('se.missing'), missing),
            h('div', { class: 'text-sm font-medium mb-1', text: t('se.multipliers') }), mult, h('div', { class: 'text-sm font-medium mt-3 mb-1', text: t('se.repeatMultipliers') }), rmult,
            h('div', { class: 'text-sm font-medium mt-3 mb-1', text: t('se.rules') }), rules,
            h('div', { class: 'text-sm font-medium mt-4 mb-1', text: t('se.templates') }), h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('se.templatesHint') }), tplBox,
            button({ label: '+ ' + t('se.addTemplate'), data: { 'campaign-action': 'add-template' }, onclick: function () { addTpl(); } })), 'stars'),
          section(t('se.unlock'), h('div', {}, grid(field(t('se.stageDefault'), stageDefault), field(t('se.chapterDefault'), chapterDefault)), field(t('se.minStars'), minStars, null, 'cg-num'), showLocked), 'unlock'),
          section(t('se.rewards'), h('div', {}, grid(field(t('se.replay'), replay), field(t('se.groupMode'), groupMode)), field(t('se.zone'), zone),
            h('div', { class: 'text-sm font-medium mb-1', text: t('se.milestoneModes') }), h('div', { class: 'flex flex-wrap gap-4' }, modes)), 'rewards'),
          section(t('se.xp'), grid(field(t('se.metaKey'), metaKey), field(t('se.levelTable'), table_)), 'xp'),
          section(t('se.events'), h('div', {}, grid(field(t('se.economyEvent'), ecoEvent), field(t('se.analyticsEvent'), anEvent)), onFailure), 'events'),
          section(t('se.freeride'), h('div', {}, tableOn, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('se.tableEnabledHint') }), reqId, field(t('se.requestHours'), reqHours, null, 'cg-num')), 'freeride'),
          section(t('se.items'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('se.itemsHint') }), itEnforce, itRequire, itOwned,
            grid(field(t('se.itemsParam'), itParam), field(t('se.itemsType'), itType)), grid(field(t('se.itemsAllow'), itAllow, t('se.itemsPathHint')), field(t('se.itemsDeny'), itDeny, t('se.itemsPathHint'))),
            grid(field(t('se.itemsClasses'), itClasses, t('se.itemsPathHint')), field(t('se.itemsClassField'), itClass))), 'items'),
          section(t('se.fields'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('se.fieldsHint') }), field(t('se.fieldsJson'), fieldsRaw)), 'fields')],
        read: function (c) {
          c.play = { costCurrency: costCur.value || '', costAmount: numberValue(costAmount), refundOnWin: refundOnWin.input.checked };
          c.attempts = { hours: numberValue(hours), graceSeconds: numberValue(grace), keepDays: numberValue(keep), defaultMinSeconds: numberValue(minSec), closeOthers: closeOthers.input.checked, serverClosesOthers: serverCloses.input.checked };
          c.stars = { maxStars: numberValue(maxStars), minOnSuccess: numberValue(minOn), missingField: missing.value, multipliers: mult.read(), repeatableMultipliers: rmult.read(), rules: rules.read(),
            templates: tplItems.map(function (x) { return x.read(); }) };
          try { c.fields = JSON.parse(fieldsRaw.value || '{}'); } catch (err) { return t('code.INVALID_JSON') + ' · fields'; }
          c.unlock = { stageDefault: stageDefault.value, chapterDefault: chapterDefault.value, minStars: numberValue(minStars), showLocked: showLocked.input.checked };
          c.rewards = { replayMode: replay.value, dayZone: zone.value.trim(), prizeGroupMode: groupMode.value,
            milestoneModes: ['ordered', 'repeatable', 'once'].filter(function (m, i) { return modes[i].input.checked; }) };
          c.xp = { metaKey: metaKey.value, levelTableKey: table_.value };
          c.events = { economyEvent: ecoEvent.value.trim(), analyticsEvent: anEvent.value.trim(), onFailure: onFailure.input.checked };
          c.freeride = { tableEnabled: tableOn.input.checked, requireRequestId: reqId.input.checked, requestHours: numberValue(reqHours) };
          c.items = { enforce: itEnforce.input.checked, requireItem: itRequire.input.checked, requireOwned: itOwned.input.checked, param: itParam.value.trim(), itemType: itType.value.trim(),
            classField: itClass.value.trim(), allow: splitCsv(itAllow), deny: splitCsv(itDeny), classes: splitCsv(itClasses) };
          return null;
        },
        after: function () { cgSettingsCache = null; cgNavSettings = false; cgMounted = null; cgSchedule(); }
      };
    });
  });
}
