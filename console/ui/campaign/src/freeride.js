function freeridePage(root) {
  var frame = pageFrame('freeride', t('desc.freeride'), [button({ label: t('fr.new'), variant: 'primary', data: { 'campaign-action': 'new-freeride' }, onclick: function () { freerideEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  function reload() {
    clear(frame.content).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/campaign/freeride'), loadLanguages(), loadRefs(true)]).then(function (res) {
      var list = res[0].items || [];
      clear(frame.content);
      if (!list.length) { frame.content.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      frame.content.appendChild(h('div', { 'data-campaign-list': 'freeride' }, table([
        { label: t('f.name') }, { label: t('f.key') }, { label: t('st.rewards') }, { label: t('st.limits') }, { label: t('f.rewardGroup') }, { label: '' }
      ], list.map(function (f) {
        var r = f.rewards || {};
        var prize = [r.prizeKey, r.prizeGroupKey].filter(Boolean).join(' · ');
        return { id: f.key, item: f, cells: [
          h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(f.name, f.key) }), f.isActive ? null : tag(t('inactive'), 'warning')),
          mono(f.key),
          h('span', { class: 'text-xs', text: rewardText(r.currencies) + (prize ? ' · ' + prize : '') + (f.xp ? ' · ' + fmtNum(f.xp) + ' XP' : '') }),
          h('span', { class: 'text-xs', 'data-campaign-limits': f.key, text: tr('fr.limits', null, { c: f.cooldownSeconds ? t('fr.cooldown', { n: f.cooldownSeconds }) : t('fr.noLimit'), d: f.dailyLimit ? t('fr.daily', { n: f.dailyLimit }) : t('fr.noLimit') }) }),
          f.rewardGroup ? mono(f.rewardGroup) : h('span', { text: t('none') }),
          button({ label: t('delete'), variant: 'danger', data: { 'campaign-delete': f.key }, onclick: function () {
            confirmBox(t('fr.delete', { key: f.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/campaign/freeride/' + f.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', cgErrText(err)); });
            });
          } })
        ] };
      }), function (f) { freerideEditor(f, reload); })));
    }, function (err) { clear(frame.content).appendChild(notice('error', cgErrText(err))); });
  }
  reload();
}

function freerideEditor(f, done) {
  var creating = !f;
  f = f || { name: {}, description: {}, rewards: {}, isActive: true };
  var key = input(f.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = translated(t('f.name'), 'name', f.name);
  var desc = translated(t('f.description'), 'description', f.description, { multiline: true });
  var icon = input(f.iconUrl, { name: 'iconUrl', ltr: true });
  var active = check(f.isActive, t('f.active'), 'isActive');
  var cur = currencyRows((f.rewards || {}).currencies, 'freeride');
  var prizes = prizePickers(f.rewards || {});
  var xp = numInput(f.xp || '', 'xp', 'XP');
  var group = input(f.rewardGroup, { name: 'rewardGroup', ltr: true, mono: true });
  var cooldown = numInput(f.cooldownSeconds || '', 'cooldownSeconds', t('f.cooldown'));
  var daily = numInput(f.dailyLimit || '', 'dailyLimit', t('f.daily'));
  var minLevel = numInput(f.minLevel || '', 'minLevel', t('f.minLevel'));
  var msg = h('div', {});
  var body = h('div', { 'data-campaign-form': 'freeride' }, field(t('f.key'), key), name, desc, field(t('f.icon'), icon), active,
    section(t('st.rewards'), h('div', {}, cur, prizes, field('XP', xp, null, 'cg-num')), 'rewards'), h('div', { class: 'h-3' }),
    section(t('st.limits'), h('div', {}, field(t('f.rewardGroup'), group, t('hint.rewardGroup')), h('div', { class: 'grid sm:grid-cols-3 gap-x-4' }, field(t('f.cooldown'), cooldown), field(t('f.daily'), daily), field(t('f.minLevel'), minLevel))), 'limits'), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'campaign-action': 'save-freeride' }, onclick: function () {
    busy(save, true);
    clear(msg);
    var rw = prizes.read();
    rw.currencies = cur.read();
    var out = { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), isActive: active.input.checked, sort: f.sort || 0, rewards: rw,
      xp: numberValue(xp) || 0, rewardGroup: group.value.trim(), cooldownSeconds: numberValue(cooldown) || 0, dailyLimit: numberValue(daily) || 0, minLevel: numberValue(minLevel) || 0 };
    api(creating ? 'POST' : 'PUT', '/campaign/freeride' + (creating ? '' : '/' + f.id), out).then(function () {
      closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); msg.appendChild(notice('error', cgErrText(err))); });
  } });
  sidePanel(creating ? t('fr.new') : textOf(f.name, f.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'freeride');
}
