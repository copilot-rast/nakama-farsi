function chaptersPage(root) {
  var frame = pageFrame('chapters', t('desc.chapters'), [button({ label: t('ch.new'), variant: 'primary', data: { 'campaign-action': 'new-chapter' }, onclick: function () { chapterEditor(null, reload); } })]);
  clear(root).appendChild(frame.el);
  function reload() {
    clear(frame.content).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    Promise.all([api('GET', '/campaign/chapters'), loadLanguages(), loadRefs(true)]).then(function (res) {
      var list = res[0].items || [];
      clear(frame.content);
      if (!list.length) { frame.content.appendChild(card(h('p', { class: 'p-4 text-sm text-secondary-500', text: t('empty') }))); return; }
      var move = function (i, d) {
        var ids = list.map(function (c) { return c.id; });
        var j = i + d;
        if (j < 0 || j >= ids.length) return;
        var x = ids[i]; ids[i] = ids[j]; ids[j] = x;
        api('PUT', '/campaign/chapters/order', { ids: ids }).then(reload, function (err) { toast('error', cgErrText(err)); });
      };
      frame.content.appendChild(h('div', { 'data-campaign-list': 'chapters' }, table([
        { label: t('f.sort') }, { label: t('f.name') }, { label: t('f.key') }, { label: t('nav.stages') }, { label: t('ch.unlock') }, { label: t('ch.milestones') }, { label: '' }
      ], list.map(function (c, i) {
        return { id: c.key, item: c, cells: [
          h('div', { class: 'flex gap-1' }, button({ label: '↑', title: t('st.moveUp'), disabled: i === 0, data: { 'campaign-up': c.key }, onclick: function () { move(i, -1); } }),
            button({ label: '↓', title: t('st.moveDown'), disabled: i === list.length - 1, data: { 'campaign-down': c.key }, onclick: function () { move(i, 1); } })),
          h('div', { class: 'flex items-center gap-2' }, h('span', { class: 'font-medium', text: textOf(c.name, c.key) }), c.isActive ? null : tag(t('inactive'), 'warning')),
          mono(c.key),
          h('a', { href: '#/campaign/stages?chapterId=' + c.id, class: 'eco-link', text: t('ch.stages', { n: c.stageCount }) }),
          h('span', { class: 'text-xs', text: (c.unlock || []).length ? c.unlock.map(ruleSentence).join(' · ') : t('ch.noUnlock') }),
          h('span', { class: 'text-xs', text: (c.milestones || []).map(function (m) { return digits(String(m.percent)) + '٪'; }).join(' · ') || t('none') }),
          button({ label: t('delete'), variant: 'danger', data: { 'campaign-delete': c.key }, onclick: function () {
            confirmBox(t('ch.delete', { key: c.key }), t('delete')).then(function (ok) {
              if (!ok) return;
              api('DELETE', '/campaign/chapters/' + c.id).then(function () { toast('success', t('saved')); reload(); }, function (err) { toast('error', err.code === 'IN_USE' ? t('ch.inUse') : cgErrText(err)); });
            });
          } })
        ] };
      }), function (c) { chapterEditor(c, reload); })));
    }, function (err) { clear(frame.content).appendChild(notice('error', cgErrText(err))); });
  }
  reload();
}

function milestoneRows(list) {
  var box = h('div', { class: 'flex flex-col gap-3', 'data-campaign-milestones': '' });
  var rows = [];
  function addRow(m) {
    m = m || { percent: 100 };
    var pct = numInput(m.percent, 'milestone-percent', t('ch.percent'));
    var cur = currencyRows(m.currencies, 'milestone');
    var prizes = prizePickers(m);
    var el = h('div', { class: 'cg-track', 'data-campaign-milestone': '' }, h('div', { class: 'flex items-end gap-3' }, field(t('ch.percent'), pct, null, 'cg-num'), h('div', { class: 'flex-1' }),
      button({ label: t('remove'), variant: 'danger', onclick: function () { el.remove(); rows = rows.filter(function (r) { return r.el !== el; }); } })), cur, prizes);
    rows.push({ el: el, read: function () { var o = prizes.read(); o.percent = floatValue(pct) || 0; o.currencies = cur.read(); return o; } });
    box.appendChild(el);
  }
  (list || []).forEach(addRow);
  var wrap = h('div', {}, box, button({ label: '+ ' + t('ch.addMilestone'), data: { 'campaign-action': 'add-milestone' }, onclick: function () { addRow(); } }));
  wrap.read = function () { return rows.map(function (r) { return r.read(); }); };
  return wrap;
}

function chapterEditor(c, done) {
  var creating = !c;
  c = c || { name: {}, description: {}, unlock: [], milestones: [], isActive: true, metadata: {} };
  var key = input(c.key || '', { name: 'key', ltr: true, mono: true });
  if (!creating) key.disabled = true;
  var name = translated(t('f.name'), 'name', c.name);
  var desc = translated(t('f.description'), 'description', c.description, { multiline: true });
  var icon = input(c.iconUrl, { name: 'iconUrl', ltr: true });
  var address = input(c.address, { name: 'address', ltr: true, mono: true });
  var active = check(c.isActive, t('f.active'), 'isActive');
  var rules = ruleEditor(c.unlock, 'chapter');
  var milestones = milestoneRows(c.milestones);
  var msg = h('div', {});
  var body = h('div', { 'data-campaign-form': 'chapter' }, field(t('f.key'), key), name, desc, h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.icon'), icon), field(t('f.address'), address)), active,
    section(t('ch.unlock'), rules, 'unlock'), h('div', { class: 'h-3' }), section(t('ch.milestones'), milestones, 'milestones'), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'campaign-action': 'save-chapter' }, onclick: function () {
    busy(save, true);
    clear(msg);
    var out = { key: key.value.trim(), name: name.read(), description: desc.read(), iconUrl: icon.value.trim(), address: address.value.trim(), sort: c.sort || 0,
      isActive: active.input.checked, unlock: rules.read(), milestones: milestones.read(), metadata: c.metadata || {} };
    if (creating) out.sort = 1000;
    api(creating ? 'POST' : 'PUT', '/campaign/chapters' + (creating ? '' : '/' + c.id), out).then(function () {
      closePanel(); toast('success', t('saved')); done();
    }, function (err) { busy(save, false); msg.appendChild(notice('error', cgErrText(err))); });
  } });
  sidePanel(creating ? t('ch.new') : textOf(c.name, c.key), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'chapter');
}
