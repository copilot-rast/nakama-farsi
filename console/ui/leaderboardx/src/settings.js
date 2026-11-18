function lbSettingsPage(root) {
  settingsPage(root, 'settings', '/leaderboardx', function (cur, defs, d) {
    var opt = function (prefix, list) { return list.map(function (v) { return { value: v, label: t(prefix + v) }; }); };
    var notes = h('div', { class: 'flex flex-col gap-2' });
    if (d.prizes === false) notes.appendChild(notice('warning', t('se.noPrizes')));
    var ds = cur.display, r = cur.rewards, s = cur.scores, hi = cur.history, e = cur.events;
    var topLimit = numInput(ds.topLimit, 'display.topLimit'), maxLimit = numInput(ds.maxLimit, 'display.maxLimit'), around = numInput(ds.aroundLimit, 'display.aroundLimit');
    var showName = check(ds.showDisplayName, t('se.showDisplayName'), 'display.showDisplayName');
    var showUpcoming = check(ds.showUpcoming, t('se.showUpcoming'), 'display.showUpcoming');
    var endedDays = numInput(ds.showEndedDays, 'display.showEndedDays');
    var payMode = select(r.payMode, opt('pay.', ['auto', 'claim']), { name: 'rewards.payMode' });
    var claimDays = numInput(r.claimDays, 'rewards.claimDays');
    var stack = select(r.stack, opt('se.stack.', ['first', 'all']), { name: 'rewards.stack' });
    var groupMode = select(r.prizeGroupMode, opt('se.groupMode.', ['draw', 'all']), { name: 'rewards.prizeGroupMode' });
    var maxTries = numInput(r.maxTries, 'rewards.maxTries'), catchUp = numInput(r.catchUpDays, 'rewards.catchUpDays');
    var reqId = check(s.requireRequestId, t('se.requireRequestId'), 'scores.requireRequestId');
    var perMinute = numInput(s.playerPerMinute, 'scores.playerPerMinute'), logDays = numInput(s.logDays, 'scores.logDays'), clientMax = numInput(s.clientMaxPerWrite, 'scores.clientMaxPerWrite');
    var keepDays = numInput(hi.keepDays, 'history.keepDays'), playerLimit = numInput(hi.playerLimit, 'history.playerLimit'), winnersLimit = numInput(hi.winnersLimit, 'history.winnersLimit');
    var ecoEvent = input(e.economyEvent, { name: 'events.economyEvent', ltr: true, mono: true });
    var anEvent = input(e.analyticsEvent, { name: 'events.analyticsEvent', ltr: true, mono: true });
    var notify = check(e.notify, t('se.notify'), 'events.notify');
    var code = numInput(e.notifyCode, 'events.notifyCode');
    var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
    return {
      el: [notes,
        section(t('se.display'), h('div', {}, grid(field(t('se.topLimit'), topLimit), field(t('se.maxLimit'), maxLimit)), grid(field(t('se.aroundLimit'), around), field(t('se.showEndedDays'), endedDays)), showName, showUpcoming), 'display'),
        section(t('se.rewards'), h('div', {}, grid(field(t('se.payMode'), payMode), field(t('se.claimDays'), claimDays)), grid(field(t('se.stack'), stack), field(t('se.groupMode'), groupMode)),
          grid(field(t('se.maxTries'), maxTries), field(t('se.catchUpDays'), catchUp))), 'rewards'),
        section(t('se.scores'), h('div', {}, reqId, grid(field(t('se.playerPerMinute'), perMinute), field(t('se.logDays'), logDays)), field(t('se.clientMaxPerWrite'), clientMax, null, 'lb-num')), 'scores'),
        section(t('se.history'), h('div', {}, field(t('se.keepDays'), keepDays, null, 'lb-num'), grid(field(t('se.playerLimit'), playerLimit), field(t('se.winnersLimit'), winnersLimit))), 'history'),
        section(t('se.events'), h('div', {}, grid(field(t('se.economyEvent'), ecoEvent), field(t('se.analyticsEvent'), anEvent)), notify, field(t('se.notifyCode'), code, null, 'lb-num')), 'events')],
      read: function (c) {
        c.display = { topLimit: numberValue(topLimit), maxLimit: numberValue(maxLimit), aroundLimit: numberValue(around), showDisplayName: showName.input.checked, showUpcoming: showUpcoming.input.checked, showEndedDays: numberValue(endedDays) };
        c.rewards = { payMode: payMode.value, claimDays: numberValue(claimDays), stack: stack.value, prizeGroupMode: groupMode.value, maxTries: numberValue(maxTries), catchUpDays: numberValue(catchUp) };
        c.scores = { requireRequestId: reqId.input.checked, playerPerMinute: numberValue(perMinute), logDays: numberValue(logDays), clientMaxPerWrite: numberValue(clientMax) || 0 };
        c.history = { keepDays: numberValue(keepDays), playerLimit: numberValue(playerLimit), winnersLimit: numberValue(winnersLimit) };
        c.events = { economyEvent: ecoEvent.value.trim(), analyticsEvent: anEvent.value.trim(), notify: notify.input.checked, notifyCode: numberValue(code) || 0 };
        return null;
      }
    };
  });
}
