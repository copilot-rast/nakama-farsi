var CL_ROLES = ['leader', 'co_leader', 'elder', 'member'];
var clClanSettings = null;
function loadClanSettings(force) {
  if (clClanSettings && !force) return Promise.resolve(clClanSettings);
  return api('GET', '/clans/settings').then(function (d) { clClanSettings = d.settings; return clClanSettings; }, function () { return null; });
}

function joinTag(j) {
  return tag(t('jt.' + j), j === 'open' ? 'success' : j === 'invite' ? 'primary' : 'warning');
}

function levelBar(c) {
  var span = (c.nextLevelXp || 0) - (c.levelXp || 0);
  var pct = span > 0 ? Math.max(0, Math.min(100, Math.round(((c.xp - c.levelXp) / span) * 100))) : 100;
  var bar = h('div', { class: 'cl-bar', role: 'progressbar', 'aria-valuenow': pct, 'aria-valuemin': 0, 'aria-valuemax': 100 }, h('span', {}));
  bar.firstChild.style.width = pct + '%';
  return h('div', { class: 'flex flex-col gap-1 min-w-[7rem]' }, h('span', { class: 'text-xs', text: c.nextLevelXp ? t('cl.xpOf', { x: fmtNum(c.xp), n: fmtNum(c.nextLevelXp) }) : t('cl.xpMax', { x: fmtNum(c.xp) }) }), bar);
}

function clansPage(root) {
  var state = { page: 1, limit: 25, total: 0, q: '', joinType: '', sort: 'trophies' };
  var frame = pageFrame('clans/list', t('desc.clans/list'));
  clear(root).appendChild(frame.el);
  var search = searchBox(function (v) { state.q = v; state.page = 1; reload(); });
  var jt = select('', [{ value: '', label: t('all') }].concat(opt('jt.', ['open', 'closed', 'invite'])), { name: 'joinType' });
  var sort = select('trophies', opt('cl.sort.', ['trophies', 'level', 'members', 'new', 'wars']), { name: 'sort' });
  jt.addEventListener('change', function () { state.joinType = jt.value; state.page = 1; reload(); });
  sort.addEventListener('change', function () { state.sort = sort.value; state.page = 1; reload(); });
  var listBox = h('div', { class: 'flex flex-col gap-3' });
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2 items-end cl-filters' }, field(t('cl.search'), search, t('cl.hint.search')), field(t('cl.joinType'), jt), field(t('cl.sortL'), sort)));
  frame.content.appendChild(listBox);
  function reload() {
    clear(listBox).appendChild(loading());
    api('GET', '/clans/list' + qs({ q: state.q, joinType: state.joinType, sort: state.sort, page: state.page, limit: state.limit })).then(function (d) {
      state.total = d.total || 0;
      var items = d.items || [];
      clear(listBox);
      listBox.appendChild(h('div', { 'data-clan-list': 'clans' }, table([
        { label: t('cl.clan') }, { label: t('cl.level') }, { label: t('cl.members') }, { label: t('tp.trophies') }, { label: t('cl.wars') }, { label: t('cl.joinType') }, { label: t('f.created') }
      ], items.map(function (c) {
        return { id: c.clanId, item: c, cells: [
          nameCell(c.name, '#' + c.tag, c.avatarUrl, c.badge && c.badge.color),
          h('div', { class: 'flex items-center gap-2' }, tag(t('cl.lv', { n: c.level }), 'primary'), levelBar(c)),
          h('span', { text: t('cl.of', { n: fmtNum(c.members), m: fmtNum(c.maxMembers) }) }),
          h('span', { class: 'font-semibold', text: fmtNum(c.trophies) }),
          h('span', { text: t('tp.wld', { w: fmtNum(c.warWins), l: fmtNum(c.warLosses), d: fmtNum(c.warDraws) }) }),
          joinTag(c.joinType),
          h('span', { class: 'text-xs', text: fmtDate(c.createdAt) })
        ] };
      }), function (c) { clanPanel(c.clanId, c.name, reload); })));
      if (!items.length) listBox.appendChild(emptyCard(t('cl.empty')));
      listBox.appendChild(pager(state, function (p) { state.page = p; reload(); }));
    }, function (err) { clear(listBox).appendChild(notice('error', clErrText(err))); });
  }
  reload();
}

function badgeControl(list, value, name) {
  if (list && list.length) {
    var opts = [{ value: '', label: t('none') }].concat(list.map(function (v) { return { value: v, label: v }; }));
    if (value && list.indexOf(value) === -1) opts.push({ value: value, label: value + ' (' + t('missing') + ')' });
    var el = select(value || '', opts, { name: name, label: name });
    el.setAttribute('dir', 'ltr');
    return el;
  }
  return input(value || '', { name: name, ltr: true, mono: true });
}

function clanPanel(clanId, label, done) {
  var body = h('div', { 'data-clan-detail': clanId }, loading());
  var del = button({ label: t('cl.delete'), variant: 'danger', data: { 'clan-action': 'delete-clan' }, onclick: function () {
    confirmBox(t('cl.deleteConfirm', { name: label }), t('delete')).then(function (ok) {
      if (!ok) return;
      busy(del, true);
      api('DELETE', '/clans/list/' + clanId).then(function () { closePanel(); toast('success', t('deleted')); if (done) done(); }, function (err) { busy(del, false); toast('error', clErrText(err)); });
    });
  } });
  sidePanel(label, body, [del, h('div', { class: 'flex-1' }), button({ label: t('close'), onclick: closePanel })], 'clan');
  function changed(d) { paint(d); if (done) done(); }
  function paint(d) {
    var c = d.clan;
    var cfg = clClanSettings || { badge: {} };
    var badge = c.badge || {};
    label = c.name;
    clear(body);
    body.appendChild(lookHeader(c));
    body.appendChild(h('div', { class: 'mb-4' }, dl([
      [t('cl.tag'), mono('#' + c.tag)],
      [t('cl.level'), h('div', { class: 'flex items-center gap-2' }, tag(t('cl.lv', { n: c.level }), 'primary'), levelBar(c))],
      [t('cl.members'), h('span', { 'data-clan-value': 'members', text: t('cl.of', { n: fmtNum(c.members), m: fmtNum(c.maxMembers) }) })],
      [t('tp.trophies'), fmtNum(c.trophies)],
      [t('cl.wars'), t('cl.warStats', { w: fmtNum(c.warWins), l: fmtNum(c.warLosses), d: fmtNum(c.warDraws), p: fmtNum(c.warPoints) })],
      [t('cl.requests'), fmtNum(d.requests || 0)],
      [t('f.created'), fmtDate(c.createdAt)]
    ], 'clan')));
    var name = input(c.name, { name: 'name' });
    var desc = input(c.description, { name: 'description', multiline: true });
    var avatar = input(c.avatarUrl, { name: 'avatarUrl', ltr: true });
    var emblem = badgeControl(cfg.badge.emblems, badge.emblem, 'badge.emblem');
    var color = badgeControl(cfg.badge.colors, badge.color, 'badge.color');
    var frameB = badgeControl(cfg.badge.frames, badge.frame, 'badge.frame');
    var jt = select(c.joinType, opt('jt.', ['open', 'closed', 'invite']), { name: 'joinType' });
    var minT = numInput(c.minTrophies || 0, 'minTrophies');
    var pmsg = h('div', {});
    var psave = button({ label: t('save'), variant: 'primary', data: { 'clan-action': 'save-profile' }, onclick: function () {
      clear(pmsg);
      busy(psave, true);
      var prof = { name: name.value.trim(), description: desc.value.trim(),
        badge: { emblem: emblem.value.trim(), color: color.value.trim(), frame: frameB.value.trim() }, joinType: jt.value, minTrophies: intOr(minT, 0) };
      if (avatar.value.trim() !== (c.avatarUrl || '')) prof.avatarUrl = avatar.value.trim();
      api('PUT', '/clans/list/' + clanId, prof).then(function (out) {
        toast('success', t('saved')); changed(out);
      }, function (err) { busy(psave, false); fieldError(pmsg, err); });
    } });
    body.appendChild(section(t('cl.profile'), h('div', { 'data-clan-form': 'profile' }, field(t('f.name'), name), field(t('f.description'), desc),
      field(t('cl.avatar'), avatar), grid3(field(t('cl.emblem'), emblem), field(t('f.color'), color), field(t('cl.frame'), frameB)),
      grid(field(t('cl.joinType'), jt), field(t('cl.minTrophies'), minT, t('cl.hint.minTrophies'))), pmsg, h('div', { class: 'flex justify-end' }, psave)), 'profile'));
    body.appendChild(h('div', { class: 'h-3' }));
    body.appendChild(lookForm(c, clanId, changed));
    body.appendChild(h('div', { class: 'h-3' }));
    var amount = numInput('', 'xp-amount', t('cl.xpAmount'));
    var xmsg = h('div', {});
    var xbtn = button({ label: t('cl.addXp'), variant: 'primary', data: { 'clan-action': 'add-xp' }, onclick: function () {
      clear(xmsg);
      var n = numberValue(amount);
      if (!n || isNaN(n) || n <= 0) { xmsg.appendChild(notice('error', t('code.VALUE_OUT_OF_RANGE') + ' · amount')); return; }
      busy(xbtn, true);
      api('POST', '/clans/list/' + clanId + '/xp', { amount: n }).then(function (out) { toast('success', t('cl.xpAdded', { n: fmtNum(n) })); changed(out); }, function (err) { busy(xbtn, false); fieldError(xmsg, err); });
    } });
    body.appendChild(section(t('cl.xp'), h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('cl.hint.xp') }),
      h('div', { class: 'cl-row' }, field(t('cl.xpAmount'), amount), h('div', {}, xbtn)), xmsg), 'add-xp'));
    body.appendChild(h('div', { class: 'h-3' }));
    var members = d.members || [];
    var mlist = h('div', { class: 'cl-rows' });
    members.forEach(function (m) {
      var role = select(m.role, opt('role.', CL_ROLES), { name: 'role', label: t('cl.role') });
      var rbtn = button({ label: t('cl.setRole'), data: { 'clan-action': 'role' }, onclick: function () {
        if (role.value === m.role) return;
        var go = role.value === 'leader' ? confirmBox(t('cl.leaderConfirm', { name: m.displayName || m.username }), t('cl.setRole')) : Promise.resolve(true);
        go.then(function (ok) {
          if (!ok) { role.value = m.role; return; }
          busy(rbtn, true);
          api('PUT', '/clans/list/' + clanId + '/members/' + m.userId, { role: role.value }).then(function (out) { toast('success', t('saved')); changed(out); }, function (err) { busy(rbtn, false); role.value = m.role; toast('error', clErrText(err)); });
        });
      } });
      var kick = button({ label: t('cl.kick'), variant: 'danger', data: { 'clan-action': 'kick' }, onclick: function () {
        confirmBox(t('cl.kickConfirm', { name: m.displayName || m.username }), t('cl.kick')).then(function (ok) {
          if (!ok) return;
          busy(kick, true);
          api('DELETE', '/clans/list/' + clanId + '/members/' + m.userId).then(function (out) { toast('success', t('cl.kicked')); changed(out); }, function (err) { busy(kick, false); toast('error', clErrText(err)); });
        });
      } });
      mlist.appendChild(h('div', { class: 'cl-member', 'data-clan-member': m.userId },
        h('div', { class: 'flex flex-col min-w-0 cl-member-who' }, playerLink(m.userId, m.displayName || m.username),
          h('span', { class: 'text-xs text-secondary-500', text: t('cl.memberLine', { x: fmtNum(m.xp), p: fmtNum(m.warPoints), t: fmtNum(m.trophies) }) + (m.joinedAt ? ' · ' + fmtDate(m.joinedAt) : '') })),
        h('div', { class: 'flex flex-wrap gap-2 items-center' }, tag(t('role.' + m.role), m.role === 'leader' ? 'warning' : m.role === 'member' ? 'secondary' : 'primary'), role, rbtn, kick)));
    });
    body.appendChild(section(t('cl.membersT', { n: fmtNum(members.length) }), members.length ? mlist : h('p', { class: 'text-sm text-secondary-500', text: t('cl.noMembers') }), 'members'));
    body.appendChild(h('div', { class: 'h-3' }));
    var xp = d.xp || [];
    body.appendChild(section(t('cl.xpLog'), xp.length ? table([{ label: t('f.date') }, { label: t('cl.source') }, { label: t('cl.amount') }, { label: t('f.player') }],
      xp.map(function (x, i) {
        return { id: String(i), cells: [h('span', { class: 'text-xs', text: fmtDate(x.createdAt) }), h('span', { text: has('xs.' + x.source) ? t('xs.' + x.source) : x.source }), deltaText(x.amount),
          x.userId ? playerLink(x.userId, x.username) : h('span', { class: 'text-secondary-400', text: t('none') })] };
      })) : h('p', { class: 'text-sm text-secondary-500', text: t('cl.noXp') }), 'xp'));
    body.appendChild(h('div', { class: 'h-3' }));
    var wars = d.wars || [];
    body.appendChild(section(t('cl.warsT'), wars.length ? table([{ label: t('wr.config') }, { label: t('wr.mode') }, { label: t('f.state') }, { label: t('rd.score') }, { label: t('rd.ends') }],
      wars.map(function (w) {
        var mine = w.clanA === clanId;
        return { id: w.warId, cells: [mono(w.configKey), h('span', { text: t('mode.' + w.mode) }),
          w.status === 'active' ? tag(t('rd.st.active'), 'warning') : !w.outcome ? tag(t('out.cancelled'), 'secondary') : tag(t('out.' + w.outcome), w.outcome === 'win' ? 'success' : w.outcome === 'lose' ? 'error' : 'secondary'),
          h('span', { dir: 'ltr', text: digits(String(mine ? w.scoreA : w.scoreB) + ' : ' + String(mine ? w.scoreB : w.scoreA)) }), h('span', { class: 'text-xs', text: fmtDate(w.endsAt) })] };
      })) : h('p', { class: 'text-sm text-secondary-500', text: t('cl.noWars') }), 'wars'));
  }
  Promise.all([api('GET', '/clans/list/' + clanId), loadClanSettings(), loadArt().catch(function () { return []; }), loadCountries(), loadLanguages()]).then(function (res) { paint(res[0]); }, function (err) { clear(body).appendChild(notice('error', clErrText(err))); });
}
