function quizSettingsPage(root) {
  loadRefs(true).then(function () {
    settingsPage(root, 'settings', '/quiz', function (cur, defs, d) {
      var opt = function (prefix, list) { return list.map(function (v) { return { value: v, label: t(prefix + v) }; }); };
      var notes = h('div', { class: 'flex flex-col gap-2' });
      if (d.prizes === false) notes.appendChild(notice('warning', t('se.noPrizes')));
      if (d.campaign === false) notes.appendChild(notice('secondary', t('se.noCampaign')));
      var p = cur.play, tm = cur.timing, sc = cur.scoring, rv = cur.reveal, rw = cur.rewards, lm = cur.limits, hn = cur.hints, rp = cur.reports, ev = cur.events, cp = cur.campaign, rt = cur.retention;
      var delivery = select(p.delivery, opt('se.delivery.', ['all', 'one_by_one']), { name: 'play.delivery' });
      var shuffleQ = check(p.shuffleQuestions, t('se.shuffleQ'), 'play.shuffleQuestions');
      var shuffleO = check(p.shuffleOptions, t('se.shuffleO'), 'play.shuffleOptions');
      var restart = select(p.onRestart, opt('se.restart.', ['resume', 'discard']), { name: 'play.onRestart' });
      var hours = numInput(p.sessionHours, 'play.sessionHours');
      var autoFinish = check(p.autoFinish, t('se.autoFinish'), 'play.autoFinish');
      var qsec = numInput(tm.questionSeconds, 'timing.questionSeconds');
      var grace = numInput(tm.graceSeconds, 'timing.graceSeconds');
      var late = select(tm.lateAnswer, opt('se.late.', ['wrong', 'reject']), { name: 'timing.lateAnswer' });
      var pass = numInput(sc.passPercent, 'scoring.passPercent');
      var multi = select(sc.multiCorrect, opt('se.multi.', ['any', 'all']), { name: 'scoring.multiCorrect' });
      var match = select(sc.textMatch, opt('se.match.', ['normalized', 'exact']), { name: 'scoring.textMatch' });
      var reveal = select(rv.mode, opt('se.reveal.', ['after_answer', 'after_finish', 'never']), { name: 'reveal.mode' });
      var before = check(rv.allowBeforeAnswer, t('se.allowBefore'), 'reveal.allowBeforeAnswer');
      var expl = check(rv.explanation, t('se.explanation'), 'reveal.explanation');
      var packMode = select(rw.packMode, opt('se.packMode.', ['first_pass', 'every_pass', 'none']), { name: 'rewards.packMode' });
      var qMode = select(rw.questionMode, opt('se.qMode.', ['first_correct', 'every_correct', 'none']), { name: 'rewards.questionMode' });
      var groupMode = select(rw.prizeGroupMode, opt('se.groupMode.', ['draw', 'all']), { name: 'rewards.prizeGroupMode' });
      var zone = input(rw.dayZone, { name: 'rewards.dayZone', ltr: true });
      var dailyPlays = numInput(lm.dailyPlays, 'limits.dailyPlays');
      var dailyPasses = numInput(lm.dailyRewardedPasses, 'limits.dailyRewardedPasses');
      var maxPlays = numInput(lm.maxPlaysPerPack, 'limits.maxPlaysPerPack');
      var replay = check(lm.replayAfterPass, t('se.replay'), 'limits.replayAfterPass');
      var cooldown = numInput(lm.retryCooldownSeconds, 'limits.retryCooldownSeconds');
      var pricing = select(hn.pricing, opt('se.pricing.', ['per_hint', 'fixed', 'free']), { name: 'hints.pricing' });
      var fixedCur = refSelect('currencies', (hn.fixedPrice || {}).currencyKey, 'hints.fixedPrice.currencyKey', t('none'));
      var fixedAmt = numInput((hn.fixedPrice || {}).amount || 0, 'hints.fixedPrice.amount');
      var keep = select(hn.keep, opt('se.keep.', ['forever', 'session']), { name: 'hints.keep' });
      var maxHints = numInput(hn.maxPerQuestion, 'hints.maxPerQuestion');
      var repOn = check(rp.enabled, t('se.reportsOn'), 'reports.enabled');
      var repDaily = numInput(rp.dailyLimit, 'reports.dailyLimit');
      var repSeen = check(rp.requireSeen, t('se.requireSeen'), 'reports.requireSeen');
      var repNote = check(rp.noteRequired, t('se.noteRequired'), 'reports.noteRequired');
      var repAuto = numInput(rp.autoDisableAt, 'reports.autoDisableAt');
      var evCompleted = input(ev.completed, { name: 'events.completed', ltr: true, mono: true });
      var evPassed = input(ev.passed, { name: 'events.passed', ltr: true, mono: true });
      var evCorrect = input(ev.correct, { name: 'events.correct', ltr: true, mono: true });
      var evAnalytics = input(ev.analytics, { name: 'events.analytics', ltr: true, mono: true });
      var evFail = check(ev.onFailure, t('se.onFailure'), 'events.onFailure');
      var cpOn = check(cp.enabled, t('se.campaignOn'), 'campaign.enabled');
      var cpFail = check(cp.sendFailures, t('se.sendFailures'), 'campaign.sendFailures');
      var keepDays = numInput(rt.sessionDays, 'retention.sessionDays');
      var delDays = numInput(rt.deliveryDays, 'retention.deliveryDays');
      var attempts = numInput(rt.webhookAttempts, 'retention.webhookAttempts');
      var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
      var duel = duelSettingsSection(cur);
      return {
        el: [notes, duel.el,
          section(t('se.play'), h('div', {}, grid(field(t('se.delivery'), delivery, t('se.hint.delivery')), field(t('se.restart'), restart)), shuffleQ, shuffleO,
            grid(field(t('se.hours'), hours, t('se.hint.hours')), h('div', {})), autoFinish), 'play'),
          section(t('se.timing'), h('div', {}, grid(field(t('se.qsec'), qsec, t('se.hint.qsec')), field(t('se.grace'), grace)), field(t('se.late'), late)), 'timing'),
          section(t('se.scoring'), h('div', {}, grid(field(t('se.pass'), pass, t('se.hint.pass')), field(t('se.multi'), multi)), field(t('se.match'), match, t('se.hint.match'))), 'scoring'),
          section(t('se.reveal'), h('div', {}, field(t('se.revealMode'), reveal), before, expl), 'reveal'),
          section(t('se.rewards'), h('div', {}, grid(field(t('se.packMode'), packMode), field(t('se.qMode'), qMode)), grid(field(t('se.groupMode'), groupMode), field(t('se.zone'), zone))), 'rewards'),
          section(t('se.limits'), h('div', {}, grid(field(t('se.dailyPlays'), dailyPlays, t('se.hint.zero')), field(t('se.dailyPasses'), dailyPasses, t('se.hint.dailyPasses'))),
            grid(field(t('se.maxPlays'), maxPlays, t('se.hint.zero')), field(t('se.cooldown'), cooldown)), replay), 'limits'),
          section(t('se.hints'), h('div', {}, grid(field(t('se.pricing'), pricing), field(t('se.keep'), keep)), grid(field(t('se.fixedCur'), fixedCur), field(t('se.fixedAmt'), fixedAmt)),
            field(t('se.maxHints'), maxHints, t('se.hint.zero'), 'qz-num')), 'hints'),
          section(t('se.reports'), h('div', {}, repOn, repSeen, repNote, grid(field(t('se.repDaily'), repDaily), field(t('se.repAuto'), repAuto, t('se.hint.repAuto')))), 'reports'),
          section(t('se.events'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('se.hint.events') }), grid(field(t('se.evCompleted'), evCompleted), field(t('se.evPassed'), evPassed)),
            grid(field(t('se.evCorrect'), evCorrect), field(t('se.evAnalytics'), evAnalytics)), evFail), 'events'),
          section(t('se.campaign'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('se.hint.campaign') }), cpOn, cpFail), 'campaign'),
          section(t('se.retention'), grid(field(t('se.keepDays'), keepDays, t('se.hint.zeroKeep')), field(t('se.delDays'), delDays), field(t('se.attempts'), attempts)), 'retention')],
        read: function (c) {
          c.play = { delivery: delivery.value, shuffleQuestions: shuffleQ.input.checked, shuffleOptions: shuffleO.input.checked, onRestart: restart.value, sessionHours: numberValue(hours), autoFinish: autoFinish.input.checked };
          c.timing = { questionSeconds: numberValue(qsec) || 0, graceSeconds: numberValue(grace) || 0, lateAnswer: late.value };
          c.scoring = { passPercent: floatValue(pass) == null ? 0 : floatValue(pass), multiCorrect: multi.value, textMatch: match.value };
          c.reveal = { mode: reveal.value, allowBeforeAnswer: before.input.checked, explanation: expl.input.checked };
          c.rewards = { packMode: packMode.value, questionMode: qMode.value, prizeGroupMode: groupMode.value, dayZone: zone.value.trim() };
          c.limits = { dailyPlays: numberValue(dailyPlays) || 0, dailyRewardedPasses: numberValue(dailyPasses) || 0, maxPlaysPerPack: numberValue(maxPlays) || 0, replayAfterPass: replay.input.checked,
            retryCooldownSeconds: numberValue(cooldown) || 0 };
          c.hints = { pricing: pricing.value, fixedPrice: { currencyKey: fixedCur.value, amount: numberValue(fixedAmt) || 0 }, keep: keep.value, maxPerQuestion: numberValue(maxHints) || 0 };
          c.reports = { enabled: repOn.input.checked, dailyLimit: numberValue(repDaily) || 0, requireSeen: repSeen.input.checked, noteRequired: repNote.input.checked, autoDisableAt: numberValue(repAuto) || 0 };
          c.events = { completed: evCompleted.value.trim(), passed: evPassed.value.trim(), correct: evCorrect.value.trim(), analytics: evAnalytics.value.trim(), onFailure: evFail.input.checked };
          c.campaign = { enabled: cpOn.input.checked, sendFailures: cpFail.input.checked };
          duel.read(c);
          c.retention = { sessionDays: numberValue(keepDays) || 0, deliveryDays: numberValue(delDays) || 0, webhookAttempts: numberValue(attempts) || 1 };
          return null;
        },
        after: function () { qzSettingsCache = null; }
      };
    });
  });
}
