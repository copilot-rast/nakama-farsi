function levelsPage(root) {
  var frame = pageFrame('clans/levels', t('desc.clans/levels'));
  clear(root).appendChild(frame.el);
  frame.content.appendChild(loading());
  function paint(d) {
    clear(frame.content);
    var rows = [];
    var box = h('div', { class: 'cl-rows', 'data-clan-levels': '' });
    function addRow(l) {
      var lv = numInput(l.level, 'level', t('lv.level'));
      var xp = numInput(l.xp, 'xp', t('lv.xp'));
      var cap = numInput(l.memberCap, 'memberCap', t('lv.cap'));
      var prize = refSelect(refList('clans', 'prizes'), l.prizeKey, 'prizeKey', t('none'));
      var icon = input(l.iconUrl, { name: 'iconUrl', ltr: true });
      var perks = jsonInput(l.perks, 'perks');
      var names = {};
      var nameBox = h('div', { class: 'cl-row-full grid sm:grid-cols-2 gap-x-4' });
      (clLanguages || []).forEach(function (g) {
        var el = input((l.name || {})[g.code] || '', { name: 'name-' + g.code });
        el.setAttribute('dir', g.direction === 'rtl' ? 'rtl' : 'ltr');
        names[g.code] = el;
        nameBox.appendChild(field(t('lang.text', { label: t('f.name'), lang: g.name }), el));
      });
      var rec = { lv: lv, xp: xp, cap: cap, prize: prize, icon: icon, perks: perks, names: names };
      var row = h('div', { class: 'cl-row cl-level-row', 'data-clan-level': '' },
        field(t('lv.level'), lv), field(t('lv.xp'), xp, t('lv.hint.xp')), field(t('lv.cap'), cap), field(t('lv.prize'), prize),
        nameBox, field(t('f.icon'), icon, null, 'cl-row-full'), field(t('lv.perks'), perks, t('hint.json'), 'cl-row-full'),
        h('div', { class: 'cl-row-full flex items-center justify-between gap-2' },
          h('span', { class: 'text-xs text-secondary-500', text: l.clans != null ? t('lv.clans', { n: fmtNum(l.clans) }) : t('lv.new') }),
          button({ label: t('remove'), variant: 'danger', data: { 'clan-action': 'remove-level' }, onclick: function () { row.remove(); rows = rows.filter(function (r) { return r !== rec; }); } })));
      rec.row = row;
      rows.push(rec);
      box.appendChild(row);
    }
    (d.items || []).forEach(addRow);
    var msg = h('div', { 'data-clan-msg': '' });
    var addBtn = button({ label: '+ ' + t('lv.add'), data: { 'clan-action': 'add-level' }, onclick: function () {
      var last = rows[rows.length - 1];
      var lastLv = last ? intOr(last.lv, rows.length) : 0;
      var lastXp = last ? intOr(last.xp, 0) : -100;
      addRow({ level: lastLv + 1, xp: lastXp + 100 * (lastLv + 1), memberCap: last ? intOr(last.cap, 30) : 30, name: {}, perks: {} });
    } });
    var save = button({ label: t('save'), variant: 'primary', data: { 'clan-save': 'levels' }, onclick: function () {
      clear(msg);
      var items = [];
      for (var i = 0; i < rows.length; i++) {
        var r = rows[i];
        var p = r.perks.read();
        if (p === undefined) { msg.appendChild(notice('error', t('code.INVALID_JSON') + ' · ' + t('lv.perks') + ' ' + digits(String(i + 1)))); return; }
        var name = {};
        Object.keys(r.names).forEach(function (k) { var v = r.names[k].value.trim(); if (v) name[k] = v; });
        items.push({ level: intOr(r.lv, 0), xp: intOr(r.xp, 0), memberCap: intOr(r.cap, 0), prizeKey: r.prize.value, iconUrl: r.icon.value.trim(), name: name, perks: p });
      }
      busy(save, true);
      api('PUT', '/clans/levels', { items: items }).then(function (out) { toast('success', t('saved')); paint(out); }, function (err) { busy(save, false); msg.appendChild(notice('error', clErrText(err))); });
    } });
    frame.content.appendChild(notice('secondary', t('lv.hint')));
    frame.content.appendChild(box);
    frame.content.appendChild(h('div', {}, addBtn));
    frame.content.appendChild(h('div', { class: 'cl-savebar flex flex-wrap items-center justify-between gap-3' }, h('div', { class: 'flex-1' }, msg), save));
  }
  Promise.all([api('GET', '/clans/levels'), loadLanguages(), loadRefs('clans')]).then(function (res) { paint(res[0]); }, function (err) { clear(frame.content).appendChild(notice('error', clErrText(err))); });
}
