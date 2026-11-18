function cupSettingsPage(root) {
  loadRefs('cups', true).then(function () {
    settingsPage(root, 'cups/settings', '/cups', function (cur) {
      var ev = cur.events, ds = cur.display, lm = cur.limits;
      var enabled = check(cur.enabled, t('cs.enabled'), 'enabled');
      var notify = check(ev.notify, t('cs.notify'), 'events.notify');
      var code = numInput(ev.code, 'events.code', t('cs.code'));
      var inApp = check(ev.inApp, t('cs.inApp'), 'events.inApp');
      var publish = check(ev.publish, t('cs.publish'), 'events.publish');
      var listLimit = numInput(ds.listLimit, 'display.listLimit', t('cs.listLimit'));
      var historyLimit = numInput(ds.historyLimit, 'display.historyLimit', t('cs.historyLimit'));
      var finishedDays = numInput(ds.finishedDays, 'display.finishedDays', t('cs.finishedDays'));
      var maxCapacity = numInput(lm.maxCapacity, 'limits.maxCapacity', t('cs.maxCapacity'));
      var activePer = numInput(lm.activePerPlayer, 'limits.activePerPlayer', t('cs.activePerPlayer'));
      var tick = numInput(lm.tickSeconds, 'limits.tickSeconds', t('cs.tickSeconds'));
      var rules = rulesForm(cur.defaults);
      var nums = [[code, 'cs.code'], [listLimit, 'cs.listLimit'], [historyLimit, 'cs.historyLimit'], [finishedDays, 'cs.finishedDays'], [maxCapacity, 'cs.maxCapacity'], [activePer, 'cs.activePerPlayer'], [tick, 'cs.tickSeconds']];
      return {
        el: [
          section(t('cs.s.general'), [enabled, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('cs.hint.enabled') })], 'general'),
          section(t('cs.s.events'), [notify, field(t('cs.code'), code, t('cs.hint.code')), inApp, publish], 'events'),
          section(t('cs.s.limits'), [grid3(field(t('cs.maxCapacity'), maxCapacity, t('cs.hint.maxCapacity')), field(t('cs.activePerPlayer'), activePer, t('hint.zero')), field(t('cs.tickSeconds'), tick, t('cs.hint.tick'))),
            grid3(field(t('cs.listLimit'), listLimit), field(t('cs.historyLimit'), historyLimit), field(t('cs.finishedDays'), finishedDays, t('cs.hint.finishedDays')))], 'limits'),
          h('h2', { class: 'font-semibold mt-2', text: t('cs.s.defaults') }),
          h('p', { class: 'text-sm text-secondary-500 -mt-2', text: t('cs.hint.defaults') })
        ].concat(rules.el),
        read: function (c) {
          for (var i = 0; i < nums.length; i++) {
            var n = numberValue(nums[i][0]);
            if (n === null || isNaN(n)) return t('ru.badNumber', { field: t(nums[i][1]) });
          }
          var rr = rules.read();
          if (rr.error) return rr.error;
          c.enabled = enabled.input.checked;
          c.events = { notify: notify.input.checked, code: numberValue(code), inApp: inApp.input.checked, publish: publish.input.checked };
          c.display = { listLimit: numberValue(listLimit), historyLimit: numberValue(historyLimit), finishedDays: numberValue(finishedDays) };
          c.limits = { maxCapacity: numberValue(maxCapacity), activePerPlayer: numberValue(activePer), tickSeconds: numberValue(tick) };
          c.defaults = rr.rules;
          return null;
        }
      };
    });
  });
}
