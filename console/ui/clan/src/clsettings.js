var CL_PERMS = ['accept', 'invite', 'kick', 'promote', 'edit', 'war'];

function costRows(values) {
  var box = h('div', { class: 'cl-rows', 'data-clan-cost': '' });
  var rows = [];
  function addRow(k, v) {
    var key = refSelect(refList('clans', 'currencies'), k || '', 'cost-currency', '—');
    var amt = numInput(v == null ? '' : v, 'cost-amount', t('cs.amount'));
    var rec = { key: key, amt: amt };
    var row = h('div', { class: 'cl-row' }, field(t('cs.currency'), key), field(t('cs.amount'), amt),
      h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r !== rec; }); } })));
    rows.push(rec);
    box.appendChild(row);
  }
  Object.keys(values || {}).forEach(function (k) { addRow(k, values[k]); });
  var wrap = h('div', {}, box, h('div', { class: 'mt-2' }, button({ label: '+ ' + t('cs.addCost'), data: { 'clan-action': 'add-cost' }, onclick: function () { addRow(); } })));
  wrap.read = function () {
    var out = {};
    rows.forEach(function (r) { var n = numberValue(r.amt); if (r.key.value && n) out[r.key.value] = n; });
    return out;
  };
  return wrap;
}

function xpSourceRows(list) {
  var box = h('div', { class: 'cl-rows', 'data-clan-xp-sources': '' });
  var rows = [];
  function addRow(s) {
    s = s || { type: 'game_event', event: '', amount: 1 };
    var type = select(s.type, opt('src.', ['economy_event', 'game_event']), { name: 'xp-type' });
    var ev = eventInput(eventsFor(s.type), s.event, 'xp-event', 'task_completed');
    var ctx = input(s.context, { name: 'xp-context', ltr: true, mono: true });
    var fld = input(s.field, { name: 'xp-field', ltr: true, mono: true });
    var eq = input(s.equals, { name: 'xp-equals', ltr: true, mono: true });
    var amt = numInput(s.amount, 'xp-amount', t('cs.xpAmount'));
    var per = check(s.perValue, t('cs.perValue'), 'xp-perValue');
    var rec = { type: type, ev: ev, ctx: ctx, fld: fld, eq: eq, amt: amt, per: per };
    var row = h('div', { class: 'cl-row', 'data-clan-xp-source': '' }, field(t('wr.source'), type), field(t('wr.event'), ev), field(t('cs.context'), ctx), field(t('wr.field'), fld),
      field(t('wr.equals'), eq), field(t('cs.xpAmount'), amt), h('div', { class: 'cl-row-full flex items-center justify-between gap-2' }, per,
        button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r !== rec; }); } })));
    rows.push(rec);
    box.appendChild(row);
  }
  (list || []).forEach(addRow);
  var wrap = h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('cs.hint.sources') }), box,
    h('div', { class: 'mt-2' }, button({ label: '+ ' + t('cs.addSource'), data: { 'clan-action': 'add-xp-source' }, onclick: function () { addRow(); } })));
  wrap.read = function () {
    return rows.map(function (r) {
      var o = { type: r.type.value, event: r.ev.input.value.trim(), amount: intOr(r.amt, 0) };
      ['ctx', 'fld', 'eq'].forEach(function (k, i) { var v = r[k].value.trim(); if (v) o[['context', 'field', 'equals'][i]] = v; });
      if (r.per.input.checked) o.perValue = true;
      return o;
    });
  };
  return wrap;
}

function permGrid(roles) {
  var boxes = {};
  var head = h('tr', {}, h('th', { class: 'p-2 text-start text-sm font-medium', text: t('cs.action') }), CL_ROLES.map(function (r) { return h('th', { class: 'p-2 text-center text-sm font-medium', text: t('role.' + r) }); }));
  var body = h('tbody', {});
  CL_PERMS.forEach(function (p) {
    boxes[p] = {};
    var tr = h('tr', { class: 'border-t border-secondary-200 dark:border-secondary-700', 'data-clan-perm': p }, h('td', { class: 'p-2 text-sm', 'data-label': t('cs.action') }, h('div', { class: 'flex flex-col' },
      h('span', { class: 'font-medium', text: t('cs.perm.' + p) }), h('span', { class: 'text-xs text-secondary-500', text: t('cs.permHint.' + p) }))));
    CL_ROLES.forEach(function (r) {
      var c = h('input', { type: 'checkbox', class: 'cl-check', name: 'roles.' + p + '.' + r, 'aria-label': t('cs.perm.' + p) + ' · ' + t('role.' + r), checked: (roles[p] || []).indexOf(r) !== -1 });
      boxes[p][r] = c;
      tr.appendChild(h('td', { class: 'p-2 text-center', 'data-label': t('role.' + r) }, c));
    });
    body.appendChild(tr);
  });
  var el = h('div', { class: 'overflow-x-auto' }, h('table', { class: 'w-full cl-perm-table', 'data-clan-perms': '' }, h('thead', {}, head), body));
  el.read = function () {
    var out = {};
    CL_PERMS.forEach(function (p) { out[p] = CL_ROLES.filter(function (r) { return boxes[p][r].checked; }); });
    return out;
  };
  return el;
}

function tagRows(list) {
  var box = h('div', { class: 'cl-rows', 'data-clan-tag-rows': '' });
  var rows = [];
  function addRow(g) {
    g = g || { key: '', name: {} };
    var key = input(g.key, { name: 'tag-key', ltr: true, mono: true, placeholder: 'friendly' });
    var names = {};
    var cells = [field(t('f.key'), key)];
    (clLanguages || []).forEach(function (l) {
      var el = input((g.name || {})[l.code] || '', { name: 'tag-name-' + l.code });
      el.setAttribute('dir', l.direction === 'rtl' ? 'rtl' : 'ltr');
      names[l.code] = el;
      cells.push(field(t('lang.text', { label: t('f.name'), lang: l.name }), el));
    });
    var rec = { key: key, names: names };
    var row = h('div', { class: 'cl-row', 'data-clan-tag-row': '' }, cells,
      h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r !== rec; }); } })));
    rows.push(rec);
    box.appendChild(row);
  }
  (list || []).forEach(addRow);
  var wrap = h('div', {}, box, h('div', { class: 'mt-2' }, button({ label: '+ ' + t('cs.addTag'), data: { 'clan-action': 'add-tag' }, onclick: function () { addRow(); } })));
  wrap.read = function () {
    return rows.map(function (r) {
      var name = {};
      Object.keys(r.names).forEach(function (k) { var v = r.names[k].value.trim(); if (v) name[k] = v; });
      return { key: r.key.value.trim(), name: name };
    }).filter(function (g) { return g.key; });
  };
  return wrap;
}

function clanSettingsPage(root) {
  Promise.all([loadRefs('clans', true), loadLanguages()]).then(function () {
    settingsPage(root, 'clans/settings', '/clans', function (cur) {
      var cr = cur.create, bd = cur.badge, ro = cur.roles, cd = cur.cooldowns, xp = cur.xp, ch = cur.chat, ds = cur.display, ev = cur.events;
      var enabled = check(cur.enabled, t('cs.enabled'), 'enabled');
      var crOn = check(cr.enabled, t('cs.createOn'), 'create.enabled');
      var cost = costRows(cr.cost);
      var crMin = numInput(cr.minTrophies, 'create.minTrophies');
      var nameMin = numInput(cr.nameMin, 'create.nameMin');
      var nameMax = numInput(cr.nameMax, 'create.nameMax');
      var descMax = numInput(cr.descriptionMax, 'create.descriptionMax');
      var defJoin = select(cr.defaultJoinType, opt('jt.', ['open', 'closed', 'invite']), { name: 'create.defaultJoinType' });
      var maxMin = numInput(cr.maxMinTrophies, 'create.maxMinTrophies');
      var emblems = tagsInput(bd.emblems, 'badge.emblems');
      var colors = tagsInput(bd.colors, 'badge.colors', '#E53935، #1E88E5');
      var frames = tagsInput(bd.frames, 'badge.frames');
      var perms = permGrid(ro);
      var maxCo = numInput(ro.maxCoLeaders, 'roles.maxCoLeaders');
      var leave = select(ro.leaderLeave, opt('cs.leave.', ['transfer', 'block']), { name: 'roles.leaderLeave' });
      var joinAfter = numInput(cd.joinAfterLeaveMinutes, 'cooldowns.joinAfterLeaveMinutes');
      var rejoin = numInput(cd.rejoinAfterKickHours, 'cooldowns.rejoinAfterKickHours');
      var leaveAfter = numInput(cd.leaveAfterJoinMinutes, 'cooldowns.leaveAfterJoinMinutes');
      var invDays = numInput(cd.inviteDays, 'cooldowns.inviteDays');
      var reqs = numInput(cd.requestsPerUser, 'cooldowns.requestsPerUser');
      var sources = xpSourceRows(xp.sources);
      var cap = numInput(xp.memberDailyCap, 'xp.memberDailyCap');
      var warWin = numInput(xp.warWin, 'xp.warWin');
      var warDraw = numInput(xp.warDraw, 'xp.warDraw');
      var warLoss = numInput(xp.warLoss, 'xp.warLoss');
      var sys = check(ch.systemMessages, t('cs.system'), 'chat.systemMessages');
      var sLimit = numInput(ds.searchLimit, 'display.searchLimit');
      var hLimit = numInput(ds.historyLimit, 'display.historyLimit');
      var pub = check(ev.publish, t('ts.publish'), 'events.publish');
      var notify = check(ev.notify, t('cs.notify'), 'events.notify');
      var code = numInput(ev.code, 'events.code');
      var pf = cur.profile || {}, upl = cur.uploads || {};
      var tags = tagRows(pf.tags);
      var maxTags = numInput(pf.maxTags, 'profile.maxTags');
      var mottoMax = numInput(pf.mottoMax, 'profile.mottoMax');
      var colorsP = tagsInput(pf.colors, 'profile.colors', '#E53935، #1E88E5');
      var cooldown = numInput(pf.changeCooldownMinutes, 'profile.changeCooldownMinutes');
      var blocked = input((pf.blockedWords || []).join('\n'), { name: 'profile.blockedWords', multiline: true, rows: 5 });
      var chatWords = check(pf.useChatWords, t('cs.useChatWords'), 'profile.useChatWords');
      var reqEmblem = check(pf.requireEmblem, t('cs.requireEmblem'), 'profile.requireEmblem');
      var upOn = check(upl.enabled, t('cs.uploadsOn'), 'uploads.enabled');
      var approval = check(upl.needsApproval, t('cs.needsApproval'), 'uploads.needsApproval');
      var maxKb = numInput(Math.round((upl.maxBytes || 0) / 1024), 'uploads.maxKb');
      var typeBoxes = ['png', 'jpg', 'webp'].map(function (x) { var b = check((upl.types || []).indexOf(x) !== -1, x.toUpperCase(), 'uploads.types'); b.input.value = x; return b; });
      return {
        el: [
          section(t('ts.general'), h('div', {}, enabled, h('p', { class: 'text-xs text-secondary-500', text: t('cs.hint.enabled') })), 'general'),
          section(t('cs.create'), h('div', {}, crOn, field(t('cs.cost'), cost, t('cs.hint.cost')),
            grid3(field(t('cs.minTrophies'), crMin, t('cs.hint.minTrophies')), field(t('cs.nameMin'), nameMin), field(t('cs.nameMax'), nameMax)),
            grid3(field(t('cs.descMax'), descMax), field(t('cs.defJoin'), defJoin), field(t('cs.maxMin'), maxMin, t('cs.hint.maxMin')))), 'create'),
          section(t('cs.badge'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('cs.hint.badge') }), field(t('cl.emblem'), emblems), field(t('cs.colors'), colors), field(t('cl.frame'), frames)), 'badge'),
          section(t('cs.roles'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('cs.hint.roles') }), perms, h('div', { class: 'h-3' }),
            grid(field(t('cs.maxCo'), maxCo), field(t('cs.leaderLeave'), leave, t('cs.hint.leaderLeave')))), 'roles'),
          section(t('cs.profile'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('cs.hint.profile') }),
            group(t('cs.tags'), tags, t('cs.hint.tags')), grid3(field(t('cs.maxTags'), maxTags, t('cs.hint.maxTags')), field(t('cs.mottoMax'), mottoMax, t('cs.hint.mottoMax')), field(t('cs.lookCooldown'), cooldown, t('cs.hint.lookCooldown'))),
            field(t('cs.lookColors'), colorsP, t('cs.hint.lookColors')), field(t('cs.blocked'), blocked, t('cs.hint.blocked')), chatWords, reqEmblem), 'profile'),
          section(t('cs.uploads'), h('div', {}, upOn, h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: t('cs.hint.uploads') }), approval,
            grid(field(t('cs.maxKb'), maxKb, t('cs.hint.maxKb')), group(t('cs.types'), h('div', { class: 'flex flex-wrap gap-x-4', 'data-clan-types': '' }, typeBoxes)))), 'uploads'),
          section(t('cs.cooldowns'), h('div', {}, grid(field(t('cs.joinAfter'), joinAfter, t('cs.hint.joinAfter')), field(t('cs.rejoin'), rejoin, t('cs.hint.rejoin'))),
            grid3(field(t('cs.leaveAfter'), leaveAfter, t('cs.hint.leaveAfter')), field(t('cs.invDays'), invDays), field(t('cs.reqs'), reqs, t('cs.hint.reqs')))), 'cooldowns'),
          section(t('cs.xp'), h('div', {}, sources, h('div', { class: 'h-3' }), grid(field(t('cs.cap'), cap, t('cs.hint.cap')), h('div', {})),
            grid3(field(t('cs.warWin'), warWin), field(t('cs.warDraw'), warDraw), field(t('cs.warLoss'), warLoss))), 'xp'),
          section(t('cs.chat'), h('div', {}, sys, h('p', { class: 'text-xs text-secondary-500 -mt-2', text: t('cs.hint.system') })), 'chat'),
          section(t('ts.display'), grid(field(t('cs.sLimit'), sLimit, t('cs.hint.sLimit')), field(t('cs.hLimit'), hLimit, t('cs.hint.hLimit'))), 'display'),
          section(t('ts.events'), h('div', {}, pub, notify, field(t('cs.code'), code, t('cs.hint.code'), 'cl-num')), 'events')],
        read: function (c) {
          c.enabled = enabled.input.checked;
          c.create = { enabled: crOn.input.checked, cost: cost.read(), minTrophies: intOr(crMin, 0), nameMin: intOr(nameMin, 0), nameMax: intOr(nameMax, 0), descriptionMax: intOr(descMax, 0),
            defaultJoinType: defJoin.value, maxMinTrophies: intOr(maxMin, 0) };
          c.badge = { emblems: emblems.read(), colors: colors.read(), frames: frames.read() };
          var p = perms.read();
          c.roles = { accept: p.accept, invite: p.invite, kick: p.kick, promote: p.promote, edit: p.edit, war: p.war, maxCoLeaders: intOr(maxCo, 0), leaderLeave: leave.value };
          c.cooldowns = { joinAfterLeaveMinutes: intOr(joinAfter, 0), rejoinAfterKickHours: intOr(rejoin, 0), leaveAfterJoinMinutes: intOr(leaveAfter, 0), inviteDays: intOr(invDays, 0), requestsPerUser: intOr(reqs, 0) };
          c.xp = { sources: sources.read(), memberDailyCap: intOr(cap, 0), warWin: intOr(warWin, 0), warDraw: intOr(warDraw, 0), warLoss: intOr(warLoss, 0) };
          c.chat = { systemMessages: sys.input.checked };
          c.display = { searchLimit: intOr(sLimit, 0), historyLimit: intOr(hLimit, 0) };
          c.events = { publish: pub.input.checked, notify: notify.input.checked, code: intOr(code, 0) };
          c.profile = { mottoMax: intOr(mottoMax, 0), tags: tags.read(), maxTags: intOr(maxTags, 0), colors: colorsP.read().map(function (x) { return x.toUpperCase(); }), changeCooldownMinutes: intOr(cooldown, 0),
            blockedWords: blocked.value.split('\n').map(function (w) { return w.trim(); }).filter(Boolean), useChatWords: chatWords.input.checked, requireEmblem: reqEmblem.input.checked };
          c.uploads = { enabled: upOn.input.checked, needsApproval: approval.input.checked, maxBytes: intOr(maxKb, 0) * 1024, types: typeBoxes.filter(function (b) { return b.input.checked; }).map(function (b) { return b.input.value; }) };
          return null;
        },
        after: function (out) { clClanSettings = out.settings; }
      };
    });
  });
}
