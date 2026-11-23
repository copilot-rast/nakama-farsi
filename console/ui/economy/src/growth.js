var ACTIVITY_KINDS = ['signin', 'purchase', 'prize', 'event', 'restriction', 'report', 'ticket', 'chat', 'referral', 'config'];
var ACTIVITY_COLOR = { signin: 'secondary', purchase: 'success', prize: 'success', event: 'primary', restriction: 'error', report: 'warning', ticket: 'primary', chat: 'primary', referral: 'success', config: 'secondary' };
var ACTIVITY_ICON = { signin: 'users', purchase: 'payments', prize: 'prizes', event: 'event-types', restriction: 'ban', report: 'reports', ticket: 'tickets', chat: 'live-chat', referral: 'referrals', config: 'remote-config' };
var CONFIG_TYPES = ['string', 'number', 'bool', 'json', 'text'];
var PLATFORMS = ['android', 'ios', 'web', 'windows', 'macos', 'linux'];
var REFERRAL_STATUS_COLOR = { pending: 'warning', held: 'error', completed: 'success', rejected: 'secondary' };

function activityText(it) {
  var d = it.data || {};
  var vars = {};
  Object.keys(d).forEach(function (k) {
    var v = d[k];
    vars[k] = typeof v === 'number' ? fmtNum(v) : Array.isArray(v) ? v.join('، ') : v == null ? '' : String(v);
  });
  if (d.method && has('method.' + d.method)) vars.method = t('method.' + d.method);
  if (d.status && has('tk.status.' + d.status)) vars.status = t('tk.status.' + d.status);
  if (d.kind && has('rs.kind.' + d.kind)) vars.kind = t('rs.kind.' + d.kind);
  if (it.kind === 'referral' && d.status && has('rf.status.' + d.status)) vars.status = t('rf.status.' + d.status);
  if (it.kind === 'config' && Array.isArray(d.variants)) vars.variants = d.variants.join('، ');
  if (it.kind === 'signin' && has('act.method.' + it.action)) vars.method = t('act.method.' + it.action);
  var key = 'act.' + it.kind + '.' + it.action;
  return has(key) ? t(key, vars) : t('act.' + it.kind) + ' · ' + it.action;
}

function activityWhat(it) {
  return h('div', { class: 'flex items-start gap-2 min-w-0', 'data-activity': it.kind + ':' + it.action },
    h('span', { class: 'eco-act-icon eco-act-icon--' + it.kind }, icon(ACTIVITY_ICON[it.kind] || 'activity', 'h-4 w-4')),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'flex items-center gap-2 flex-wrap' }, tag(t('act.' + it.kind), ACTIVITY_COLOR[it.kind]), activityLink(it)),
      h('div', { class: 'text-sm', dir: 'auto', text: activityText(it) })));
}

function activityLink(it) {
  if (it.kind === 'ticket' && it.refId && aclCan('ACCOUNT_NOTES', 0)) return h('a', { href: hashFor('tickets', it.refId), class: 'eco-link text-xs', text: t('act.open') });
  if (it.kind === 'chat' && it.refId && aclCan('ACCOUNT_NOTES', 0)) return h('a', { href: hashFor('live-chat', it.refId, { view: 'all' }), class: 'eco-link text-xs', text: t('act.open') });
  if (it.kind === 'purchase' && it.refId) return h('a', { href: hashFor('payments', it.refId), class: 'eco-link text-xs', text: t('act.open') });
  if (it.kind === 'report' && it.refId && aclCan('CHANNEL_MESSAGE', 0)) return h('a', { href: hashFor('reports', it.refId), class: 'eco-link text-xs', text: t('act.open') });
  if (it.kind === 'referral' && it.refId) return h('a', { href: hashFor('referrals', null, { q: it.username || '' }), class: 'eco-link text-xs', text: t('act.open') });
  return null;
}

function activityWhere(it) {
  var parts = [];
  if (it.country) parts.push(h('span', { class: 'font-mono text-xs', dir: 'ltr', text: it.country }));
  if (it.platform) parts.push(tag(has('pf.' + it.platform) ? t('pf.' + it.platform) : it.platform, 'secondary'));
  if (it.ip) parts.push(mono(it.ip));
  if (!parts.length) return h('span', { class: 'text-secondary-400', text: '—' });
  return h('div', { class: 'flex flex-wrap items-center gap-1' }, parts);
}

function actorText(a) {
  if (!a) return t('act.actor.system');
  if (a === 'player') return t('act.actor.player');
  if (a === 'system') return t('act.actor.system');
  if (a.indexOf('console:') === 0) return t('act.actor.console', { name: a.slice(8) || 'admin' });
  return a;
}

LISTS.activity = {
  api: 'base',
  path: 'activity',
  noRow: true,
  top: function () { return activityTop(); },
  filters: [
    { name: 'kind', label: 'filter.kind', options: function () { return ACTIVITY_KINDS.map(function (k) { return { value: k, label: t('act.' + k) }; }); } },
    { name: 'userId', label: 'filter.playerId', type: 'text' },
    { name: 'platform', label: 'filter.platform', options: function () { return PLATFORMS.map(function (p) { return { value: p, label: has('pf.' + p) ? t('pf.' + p) : p }; }); } },
    { name: 'country', label: 'filter.country', type: 'text' },
    { name: 'from', label: 'filter.from', type: 'text' },
    { name: 'to', label: 'filter.to', type: 'text' }
  ],
  columns: [
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 160 },
    { label: 'col.player', render: function (it) { return playerLink(it.userId, it.displayName || it.username); }, width: 150 },
    { label: 'col.what', render: activityWhat, skeleton: '70%' },
    { label: 'col.where', render: activityWhere, width: 200 },
    { label: 'col.by', render: function (it) { return h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: actorText(it.actor) }); }, width: 140 }
  ]
};

function activityTop() {
  var box = h('div', { class: 'mb-3', 'data-activity-top': '' });
  baseApi('GET', '/activity/summary').then(function (s) {
    if (!s) return;
    box.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: t('act.retention', { n: localDigits(s.retentionDays) }) }));
  }, function () { return null; });
  return box;
}

function activityPage(container, query, ctx) {
  return listPage(container, 'activity', query, ctx);
}

var ACTIVITY_CARD = { id: null, data: null, loading: false, at: 0, err: null, kind: '' };

function loadActivityCard(id) {
  ACTIVITY_CARD.loading = true;
  return baseApi('GET', '/activity' + qs({ userId: id, limit: 8, kind: ACTIVITY_CARD.kind })).then(function (res) {
    ACTIVITY_CARD.loading = false;
    ACTIVITY_CARD.at = Date.now();
    ACTIVITY_CARD.data = res;
    ACTIVITY_CARD.err = null;
    var old = document.querySelector('[data-player-activity]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    ACTIVITY_CARD.loading = false;
    ACTIVITY_CARD.at = Date.now();
    ACTIVITY_CARD.err = err;
    schedule();
  });
}

function playerActivityCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-activity]');
  if (!id || !token() || !aclCan('ACCOUNT', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (ACTIVITY_CARD.id !== id) { ACTIVITY_CARD.id = id; ACTIVITY_CARD.data = null; ACTIVITY_CARD.err = null; ACTIVITY_CARD.at = 0; ACTIVITY_CARD.kind = ''; }
  if (!ACTIVITY_CARD.loading && Date.now() - ACTIVITY_CARD.at > 20000) loadActivityCard(id);
  var p = ACTIVITY_CARD.data;
  var anchor = grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + ACTIVITY_CARD.kind + '|' + (p ? JSON.stringify((p.items || []).map(function (x) { return x.id; })) : ACTIVITY_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-activity': '', 'data-key': key, 'data-economy': 'player-activity', 'data-l10n-skip': '', 'aria-labelledby': 'player-activity-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  var title = h('h2', { id: 'player-activity-title', class: 'font-medium', text: t('act.card') });
  var chips = segmented([{ value: '', label: t('all') }].concat(['signin', 'purchase', 'prize', 'ticket', 'restriction'].map(function (k) { return { value: k, label: t('act.' + k) }; })), ACTIVITY_CARD.kind, function (v) {
    ACTIVITY_CARD.kind = v;
    ACTIVITY_CARD.at = 0;
    loadActivityCard(id);
  }, t('filter.kind'));
  chips.setAttribute('data-activity-chips', '');
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-3' }, title,
    h('div', { class: 'flex flex-wrap items-center gap-3' }, chips,
      h('a', { href: hashFor('activity', null, { userId: id }), class: 'eco-link text-sm', 'data-activity-all': '', text: t('act.all') }),
      aclCan('ACCOUNT_NOTES', 0) ? h('a', { href: hashFor('tickets', null, { userId: id, status: '' }), class: 'eco-link text-sm', text: t('tk.playerTickets') }) : null)));
  if (!p) {
    body.appendChild(ACTIVITY_CARD.err ? notice('error', errorText(ACTIVITY_CARD.err)) : h('div', { class: 'flex flex-col gap-2' }, bar('50%'), bar('70%'), bar('40%')));
    place();
    return;
  }
  var list = h('ol', { class: 'eco-timeline', 'data-timeline': '' });
  (p.items || []).forEach(function (it) {
    list.appendChild(h('li', { class: 'eco-timeline-item', 'data-activity-item': it.kind },
      h('span', { class: 'eco-act-icon eco-act-icon--' + it.kind }, icon(ACTIVITY_ICON[it.kind] || 'activity', 'h-4 w-4')),
      h('div', { class: 'min-w-0 flex-1' },
        h('div', { class: 'text-sm', dir: 'auto', text: activityText(it) }),
        h('div', { class: 'text-xs text-secondary-500 flex flex-wrap gap-2' }, h('span', { text: fmtDate(it.createdAt, true) }), it.country ? h('span', { dir: 'ltr', text: it.country }) : null, it.platform ? h('span', { text: has('pf.' + it.platform) ? t('pf.' + it.platform) : it.platform }) : null)),
      activityLink(it)));
  });
  if (!(p.items || []).length) list.appendChild(h('li', { class: 'text-sm text-secondary-400', text: t('act.empty') }));
  body.appendChild(list);
  place();
}

function referralStatusTag(it) {
  var el = tag(t('rf.status.' + it.status), REFERRAL_STATUS_COLOR[it.status]);
  el.setAttribute('data-referral-status', it.status);
  return el;
}

function referralFlags(flags) {
  if (!flags || !flags.length) return null;
  return h('div', { class: 'flex flex-wrap gap-1 mt-1' }, flags.map(function (f) { var el = tag(has('rf.flag.' + f) ? t('rf.flag.' + f) : f, 'warning'); el.setAttribute('data-flag', f); return el; }));
}

function referralGoal(it) {
  if (!it.goalEventKey) return h('span', { class: 'text-secondary-400', text: t('rf.noGoal') });
  var done = Number(it.goalProgress || 0);
  var need = Number(it.goalValue || 0);
  return h('div', { class: 'text-sm' }, h('span', { class: 'font-mono text-xs', dir: 'ltr', text: it.goalEventKey }), h('div', { class: 'tabular-nums text-xs text-secondary-500', text: t('rf.progress', { n: fmtNum(Math.min(done, need)), total: fmtNum(need) }) }));
}

function referralActions(it, reload) {
  var box = h('div', { class: 'flex flex-wrap gap-1' });
  var tree = button({ variant: 'link', size: 'sm', label: t('rf.tree'), onclick: function () { referralTree(it.referrerId); } });
  tree.setAttribute('data-tree', it.referrerId);
  box.appendChild(tree);
  if (!aclCan('HIRO_ECONOMY', 1)) return box;
  if (it.status === 'held') {
    var ok = button({ variant: 'primary', size: 'sm', label: t('rf.approve'), onclick: function () { decide(it, 'approve', reload); } });
    ok.setAttribute('data-approve', it.id);
    box.appendChild(ok);
  }
  if (it.status === 'held' || it.status === 'pending') {
    var no = button({ variant: 'danger', size: 'sm', label: t('rf.reject'), onclick: function () { decide(it, 'reject', reload); } });
    no.setAttribute('data-reject', it.id);
    box.appendChild(no);
  }
  return box;
}

function decide(it, what, reload) {
  var note = textInput('', { multiline: true, rows: 2, label: t('rs.note'), placeholder: t('rs.note.ph') });
  var v = { note: '' };
  note.addEventListener('input', function () { v.note = note.value; });
  return formDialog({ id: 'referral-' + what, title: t('rf.' + what + '.title'), text: t('rf.' + what + '.text', { from: it.referrerName, to: it.refereeName }), icon: what === 'approve' ? 'check' : 'ban', danger: what === 'reject', ok: t('rf.' + what),
    body: field(t('rs.note'), note, null, { optional: true }),
    submit: function () { return api('POST', '/referrals/' + encodeURIComponent(it.id) + '/' + what, v); } }).then(function (res) {
    if (res) { toast('success', t('rf.' + what + '.done')); if (reload) reload(); }
  });
}

function referralTree(userId) {
  var body = h('div', { class: 'flex flex-col gap-3', 'data-referral-tree': userId }, bar('60%'), bar('80%'));
  confirmModal({ title: t('rf.tree.title'), icon: 'referrals', body: body, ok: t('close'), cancel: t('back') });
  api('GET', '/referrals/tree/' + encodeURIComponent(userId)).then(function (tr2) {
    clear(body);
    var up = (tr2.referredBy || []).slice().reverse();
    var chain = h('div', { class: 'flex flex-wrap items-center gap-1 text-sm' });
    up.forEach(function (u) { chain.appendChild(playerLink(u.userId, u.username)); chain.appendChild(h('span', { class: 'text-secondary-400', text: '←' })); });
    chain.appendChild(h('span', { class: 'font-semibold', text: tr2.user.username }));
    chain.appendChild(mono(tr2.user.code || ''));
    body.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('rf.tree.up') }), chain));
    var list = h('ul', { class: 'flex flex-col gap-1 eco-tree-list' });
    (tr2.children || []).forEach(function (c) {
      list.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-tree-child': c.userId }, playerLink(c.userId, c.username), referralStatusTag(c),
        c.referrals ? button({ variant: 'link', size: 'sm', label: t('rf.tree.more', { n: c.referrals }), onclick: function () { referralTree(c.userId); } }) : null,
        referralFlags(c.flags), h('span', { class: 'text-xs text-secondary-500', text: fmtDate(c.createdAt) })));
    });
    if (!(tr2.children || []).length) list.appendChild(h('li', { class: 'text-sm text-secondary-400', text: t('rf.tree.none') }));
    body.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('rf.tree.down', { n: (tr2.children || []).length }) }), list));
  }, function (err) { clear(body).appendChild(notice('error', errorText(err))); });
}

var referralReload = null;

function referralPerson(id, username, display, extra) {
  var box = h('span', { class: 'inline-flex flex-col min-w-0', 'data-referral-person': id || '' }, playerLink(id, username, extra));
  if (display && display !== username) box.appendChild(h('span', { class: 'text-xs text-secondary-500 truncate', dir: 'auto', 'data-display-name': '', text: display }));
  return box;
}

LISTS.referrals = {
  path: 'referrals',
  noRow: true,
  top: function () { return referralTop(); },
  filters: [
    { name: 'status', label: 'filter.status', options: function () { return ['pending', 'held', 'completed', 'rejected'].map(function (s) { return { value: s, label: t('rf.status.' + s) }; }); } },
    { name: 'flagged', label: 'filter.flagged', options: function () { return [{ value: 'true', label: t('rf.flaggedOnly') }]; } },
    { name: 'referrerId', label: 'filter.referrerId', type: 'text' }
  ],
  columns: [
    { label: 'col.referral', render: function (it) { return h('div', { class: 'flex flex-wrap items-center gap-1 min-w-0' }, referralPerson(it.referrerId, it.referrerName, it.referrerDisplayName), h('span', { class: 'text-secondary-400', text: '←' }), referralPerson(it.refereeId, it.refereeName, it.refereeDisplayName)); }, skeleton: '60%' },
    { label: 'col.code', render: function (it) { return mono(it.code); }, width: 110 },
    { label: 'col.status', render: function (it) { return h('div', {}, referralStatusTag(it), referralFlags(it.flags)); }, width: 170 },
    { label: 'col.goal', render: referralGoal, width: 150 },
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 150 },
    { label: 'col.actions', render: function (it) { return referralActions(it, function () { if (referralReload) referralReload(); }); }, width: 200 }
  ]
};

function referralTop() {
  var box = h('div', { 'data-referral-top': '' }, h('div', { class: 'eco-stat-grid mb-4' }, [1, 2, 3, 4].map(function () { return card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem'), bar('3rem', 'h-7'))); })));
  api('GET', '/referrals/summary').then(function (s) {
    clear(box);
    var grid = h('div', { class: 'eco-stat-grid mb-4' });
    grid.appendChild(statCard(t('rf.stat.total'), s.total, t('rf.stat.today', { n: fmtNum(s.today) }), null));
    grid.appendChild(statCard(t('rf.status.completed'), s.completed, null, hashFor('referrals', null, { status: 'completed' })));
    grid.appendChild(statCard(t('rf.status.pending'), s.pending, null, hashFor('referrals', null, { status: 'pending' })));
    grid.appendChild(statCard(t('rf.status.held'), s.held, t('rf.stat.flagged', { n: fmtNum(s.flagged) }), hashFor('referrals', null, { status: 'held' }), { alert: s.held > 0 }));
    box.appendChild(grid);
    if ((s.topReferrers || []).length) {
      var top = h('ol', { class: 'flex flex-col', 'data-top-referrers': '' });
      s.topReferrers.forEach(function (r, i) {
        top.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 px-4 py-2 border-b border-secondary-200 dark:border-secondary-700 text-sm' },
          h('span', { class: 'tabular-nums text-secondary-500 w-6', text: localDigits(i + 1) }), referralPerson(r.userId, r.username, r.displayName, 'font-medium'),
          h('span', { class: 'tabular-nums', text: t('rf.top.line', { n: fmtNum(r.total), done: fmtNum(r.completed) }) }),
          r.flagged ? tag(t('rf.top.flagged', { n: fmtNum(r.flagged) }), 'warning') : null,
          h('span', { class: 'flex-1' }),
          button({ variant: 'link', size: 'sm', label: t('rf.tree'), onclick: function () { referralTree(r.userId); } })));
      });
      box.appendChild(h('div', { class: 'mb-4' }, h('h2', { class: 'text-lg font-medium mb-2', text: t('rf.top') }), card(top)));
    }
  }, function () { clear(box); });
  return box;
}

function referralsPage(container, query, ctx) {
  var list = listPage(container, 'referrals', query, ctx);
  referralReload = list.reload;
  return list;
}

var REFERRAL_FIELDS = [
  { name: 'isEnabled', type: 'bool', inline: true, label: 'f.rr.enabled', hint: 'f.rr.enabled.hint' },
  { type: 'heading', label: 'rr.rewards' },
  { type: 'row', fields: [
    { name: 'referrerPrizeId', type: 'ref', resource: 'prizes', label: 'f.rr.referrerPrize', hint: 'f.rr.referrerPrize.hint', optional: true },
    { name: 'refereePrizeId', type: 'ref', resource: 'prizes', label: 'f.rr.refereePrize', hint: 'f.rr.refereePrize.hint', optional: true }
  ] },
  { name: 'goalEventTypeId', type: 'ref', resource: 'event-types', label: 'f.rr.goal', hint: 'f.rr.goal.hint', optional: true },
  { type: 'row', fields: [
    { name: 'goalMeasure', type: 'select', label: 'f.rr.measure', required: true, noPlaceholder: true, options: function () { return ['count', 'sum', 'max'].map(function (m) { return { value: m, label: t('rr.measure.' + m) }; }); } },
    { name: 'goalValue', type: 'number', int: false, label: 'f.rr.goalValue', hint: 'f.rr.goalValue.hint', required: true }
  ] },
  { type: 'heading', label: 'rr.rules' },
  { type: 'row', fields: [
    { name: 'newAccountDays', type: 'number', label: 'f.rr.days', hint: 'f.rr.days.hint', required: true, min: 0, max: 3650 },
    { name: 'maxPerReferrer', type: 'number', label: 'f.rr.max', hint: 'f.rr.max.hint', required: true, min: 1, max: 100000 }
  ] },
  { name: 'dailyPerReferrer', type: 'number', label: 'f.rr.daily', hint: 'f.rr.daily.hint', required: true, min: 1, max: 100000 },
  { name: 'onePerDevice', type: 'bool', inline: true, label: 'f.rr.device', hint: 'f.rr.device.hint' },
  { name: 'onePerPhone', type: 'bool', inline: true, label: 'f.rr.phone', hint: 'f.rr.phone.hint' },
  { name: 'holdFlagged', type: 'bool', inline: true, label: 'f.rr.hold', hint: 'f.rr.hold.hint' }
];

function referralRulesPage(container) {
  var frame = tabFrame(t('desc.referral-rules'));
  clear(container).appendChild(frame.el);
  var page = apiPage(api, '/referral-settings', null, REFERRAL_FIELDS, { id: 'referral-rules', fixPayload: function (p) {
    ['referrerPrizeId', 'refereePrizeId', 'goalEventTypeId'].forEach(function (k) { if (!p[k]) p[k] = ''; });
    p.goalValue = Number(p.goalValue) || 1;
    return p;
  } });
  frame.content.appendChild(page.el);
  page.load();
  frame.content.appendChild(referralRebindBox());
}

function referralRebindBox() {
  var box = h('div', { class: 'mt-6', 'data-referral-rebind': '' });
  function load() {
    clear(box).appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('40%'), bar('70%'))));
    api('GET', '/referral-rebind').then(function (v) {
      clear(box);
      var body = h('div', { class: 'p-4 flex flex-col gap-3' },
        h('h2', { class: 'text-base font-semibold', text: t('rb.title') }),
        h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: t('rb.text') }));
      if (!v.total) {
        body.appendChild(h('p', { class: 'text-sm text-secondary-500', 'data-rebind-none': '', text: t('rb.none') }));
        box.appendChild(card(body));
        return;
      }
      body.appendChild(h('p', { class: 'text-sm', 'data-rebind-total': v.total, text: t('rb.count', { n: fmtNum(v.total), done: fmtNum(v.willComplete) }) }));
      var list = h('ul', { class: 'flex flex-col gap-1 text-sm' });
      (v.groups || []).forEach(function (g) {
        list.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2', 'data-rebind-group': g.goalEventKey + ':' + g.status },
          h('span', { class: 'font-mono text-xs', dir: 'ltr', text: g.goalEventKey || t('rf.noGoal') }), referralStatusTag(g),
          h('span', { class: 'tabular-nums text-secondary-500', text: t('rb.line', { n: fmtNum(g.count) }) })));
      });
      body.appendChild(list);
      if (aclCan('HIRO_ECONOMY', 1)) {
        var prizes = { on: true };
        var sw = toggle(true, t('rb.prizes'), function (on) { prizes.on = on; }, 'sm');
        body.appendChild(sw);
        var go = button({ variant: 'primary', size: 'sm', label: t('rb.apply'), onclick: function () {
          formDialog({ id: 'referral-rebind', title: t('rb.confirm.title'), text: t('rb.confirm.text', { n: fmtNum(v.total) }), icon: 'referrals', ok: t('rb.apply'),
            body: h('div', {}), submit: function () { return api('POST', '/referral-rebind', { confirm: 'rebind', prizes: prizes.on }); } }).then(function (res) {
            if (res && res !== true) { toast('success', t('rb.done', { n: fmtNum(res.rebound), done: fmtNum(res.completed) })); load(); }
          });
        } });
        go.setAttribute('data-rebind-apply', '');
        body.appendChild(h('div', {}, go));
      }
      box.appendChild(card(body));
    }, function (err) { clear(box).appendChild(notice('error', errorText(err))); });
  }
  load();
  return box;
}

function configValuePreview(it) {
  var v = it.value;
  if (it.type === 'bool') return tag(v ? t('cf.on') : t('cf.off'), v ? 'success' : 'secondary');
  if (it.type === 'text') return h('span', { class: 'truncate', dir: 'auto', text: i18nText(v) || '' });
  if (it.type === 'json') return h('span', { class: 'font-mono text-xs truncate', dir: 'ltr', text: JSON.stringify(v).slice(0, 60) });
  if (it.type === 'number') return h('span', { class: 'tabular-nums', text: fmtNum(v) });
  return h('span', { class: 'truncate', dir: 'auto', text: String(v == null ? '' : v) });
}

LISTS['remote-config'] = {
  api: 'base',
  path: 'remote-config',
  rowHref: function (it) { return hashFor('remote-config', it.id); },
  create: function () {
    if (!aclCan('SETTINGS', 1)) return null;
    var b = button({ variant: 'primary', size: 'page', icon: 'plus', label: t('cf.new'), onclick: function () { go('remote-config', 'new'); } });
    b.setAttribute('data-config-new', '');
    return b;
  },
  filters: [
    { name: 'type', label: 'filter.type', options: function () { return CONFIG_TYPES.map(function (k) { return { value: k, label: t('cf.type.' + k) }; }); } },
    { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }
  ],
  columns: [
    { label: 'col.key', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-mono text-sm', dir: 'ltr', text: it.key }), it.description ? h('div', { class: 'text-xs text-secondary-500 truncate', dir: 'auto', text: it.description }) : null); }, skeleton: '50%' },
    { label: 'col.type', render: function (it) { return tag(t('cf.type.' + it.type), 'secondary'); }, width: 110 },
    { label: 'col.value', render: configValuePreview, width: 240 },
    { label: 'col.rules', render: function (it) { return it.ruleCount ? h('span', { class: 'tabular-nums', text: t('cf.rulesN', { n: it.ruleCount }) }) : h('span', { class: 'text-secondary-400', text: t('cf.everyone') }); }, width: 120 },
    { label: 'col.version', render: function (it) { return h('span', { class: 'tabular-nums', text: t('cf.v', { n: it.version }) }); }, width: 90 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ]
};

function configPage(container, query, ctx) {
  var list = listPage(container, 'remote-config', query, ctx);
  var recent = h('div', { class: 'mt-6', 'data-config-recent': '' });
  container.appendChild(h('div', { class: 'px-4 md:px-6 max-w-[1600px] mx-auto w-full' }, recent));
  baseApi('GET', '/remote-config/history?limit=8').then(function (res) {
    var items = (res && res.items) || [];
    if (!items.length) return;
    var ol = h('ol', { class: 'flex flex-col' });
    items.forEach(function (x) {
      ol.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 px-4 py-2 border-b border-secondary-200 dark:border-secondary-700 text-sm' },
        tag(t('cf.action.' + x.action), x.action === 'delete' ? 'error' : x.action === 'rollback' ? 'warning' : 'secondary'),
        x.exists ? h('a', { href: hashFor('remote-config', x.keyId), class: 'eco-link font-mono', dir: 'ltr', text: x.key }) : h('span', { class: 'font-mono', dir: 'ltr', text: x.key }),
        h('span', { class: 'tabular-nums text-secondary-500', text: t('cf.v', { n: x.version }) }),
        h('span', { class: 'text-secondary-500', text: x.changedBy }),
        h('span', { class: 'flex-1' }),
        h('span', { class: 'text-xs text-secondary-500', text: fmtDate(x.createdAt, true) })));
    });
    recent.appendChild(h('h2', { class: 'text-lg font-medium mb-2', text: t('cf.recent') }));
    recent.appendChild(card(ol));
  }, function () { return null; });
  return list;
}

function configValueInput(type, value, onchange) {
  if (type === 'bool') return toggle(!!value, t('cf.valueOn'), function (v) { onchange(v); }, 'sm');
  if (type === 'number') return numberInput(value, { int: false, label: t('cf.value'), oninput: function (raw) { onchange(raw === '' ? null : parseNumber(raw, false)); } });
  if (type === 'text') {
    var tri = translatedInput(value || {}, { label: t('cf.value'), multiline: true, rows: 2, max: 2000, oninput: function (v) { onchange(v); } });
    return tri;
  }
  if (type === 'json') {
    var ta = textInput(value == null ? '' : JSON.stringify(value, null, 2), { ltr: true, mono: true, multiline: true, rows: 5, label: t('cf.value'), oninput: function (raw) {
      try { onchange(JSON.parse(raw)); ta.classList.remove('eco-bad'); } catch (e) { onchange({ __invalid: raw }); ta.classList.add('eco-bad'); }
    } });
    return ta;
  }
  return textInput(value == null ? '' : value, { label: t('cf.value'), oninput: function (v) { onchange(v); } });
}

function configValueValid(type, v) {
  if (v && typeof v === 'object' && v.__invalid !== undefined) return t('code.INVALID_JSON');
  if (type === 'number' && (v === null || isNaN(v))) return t('err.number');
  if (type === 'text') { var hit = trCheck(v, { required: true, max: 2000 }); if (hit) return t('code.' + hit.code, { lang: langName(hit.lang) }); }
  if (type === 'string' && (v == null)) return t('err.required');
  return null;
}

function ruleEditor(type, rule, onremove, onmove) {
  var errBox = h('p', { class: 'text-xs text-error-600 dark:text-error-400', hidden: true });
  var name = textInput(rule.name || '', { ltr: true, mono: true, label: t('cf.rule.name'), oninput: function (v) { rule.name = v.trim(); } });
  var plats = h('div', { class: 'flex flex-wrap gap-3 eco-choices' }, PLATFORMS.map(function (p) {
    return checkbox((rule.platforms || []).indexOf(p) !== -1, has('pf.' + p) ? t('pf.' + p) : p, function (on) {
      rule.platforms = (rule.platforms || []).filter(function (x) { return x !== p; });
      if (on) rule.platforms.push(p);
    });
  }));
  var minV = textInput(rule.minVersion || '', { ltr: true, mono: true, label: t('cf.rule.min'), placeholder: '1.0.0', oninput: function (v) { rule.minVersion = v.trim(); } });
  var maxV = textInput(rule.maxVersion || '', { ltr: true, mono: true, label: t('cf.rule.max'), placeholder: '2.0.0', oninput: function (v) { rule.maxVersion = v.trim(); } });
  var tags = refMulti({ resource: 'tags', api: baseApi, filter: { keyType: 'user' }, value: rule.tagIds || [], label: t('cf.rule.tags'), onchange: function (v) { rule.tagIds = v; } });
  var countries = textInput((rule.countries || []).join(', '), { ltr: true, mono: true, label: t('cf.rule.countries'), placeholder: 'IR, DE', oninput: function (v) { rule.countries = v.split(/[\s,،]+/).map(function (c) { return c.trim().toUpperCase(); }).filter(Boolean); } });
  var pct = numberInput(rule.percent == null ? '' : rule.percent, { int: false, label: t('cf.rule.percent'), placeholder: '100', oninput: function (raw) { rule.percent = raw === '' ? null : parseNumber(raw, false); } });
  var valueBox = configValueInput(type, rule.value, function (v) { rule.value = v; });
  var row = h('div', { class: 'eco-rule', 'data-rule': '' },
    h('div', { class: 'flex flex-wrap items-center gap-2 mb-3' },
      h('span', { class: 'font-medium text-sm', text: t('cf.rule') }),
      h('span', { class: 'flex-1' }),
      button({ variant: 'muted', size: 'icon', icon: 'prev', iconClass: 'h-3 w-3 eco-rot90', title: t('cf.rule.up'), onclick: function () { onmove(-1); } }),
      button({ variant: 'muted', size: 'icon', icon: 'next', iconClass: 'h-3 w-3 eco-rot90', title: t('cf.rule.down'), onclick: function () { onmove(1); } }),
      button({ variant: 'danger', size: 'icon', icon: 'trash', iconClass: 'h-3 w-3', title: t('remove'), onclick: onremove })),
    h('div', { class: 'eco-form-row' }, field(t('cf.rule.name'), name, t('cf.rule.name.hint')), field(t('cf.rule.percent'), pct, t('cf.rule.percent.hint'), { optional: true })),
    field(t('cf.rule.platforms'), plats, t('cf.rule.platforms.hint'), { optional: true }),
    h('div', { class: 'eco-form-row' }, field(t('cf.rule.min'), minV, null, { optional: true }), field(t('cf.rule.max'), maxV, null, { optional: true })),
    h('div', { class: 'eco-form-row' }, field(t('cf.rule.tags'), tags, null, { optional: true }), field(t('cf.rule.countries'), countries, t('cf.rule.countries.hint'), { optional: true })),
    field(t('cf.rule.value'), valueBox, t('cf.rule.value.hint')),
    errBox);
  row.setError = function (msg) { errBox.textContent = msg || ''; errBox.hidden = !msg; };
  return row;
}

function configEditor(container, id, query) {
  var isNew = id === 'new';
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('remote-config', null, lastQuery['remote-config'] || {}); } });
  var frame = pageFrame(t(isNew ? 'cf.new' : 'cf.edit'), t('desc.remote-config.edit'), backBtn);
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-config', 'data-config-editor': id });
  frame.content.appendChild(grid);
  var write = aclCan('SETTINGS', 1);

  function skeleton() {
    clear(grid).appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('40%', 'h-5'), bar('100%', 'h-10'), bar('100%', 'h-20'), bar('60%', 'h-10'))));
  }

  function render(item) {
    var v = clone(item);
    v.rules = v.rules || [];
    clear(grid);
    var top = h('div', {});
    var f = {};
    var key = textInput(v.key || '', { ltr: true, mono: true, label: t('cf.key'), oninput: function (x) { v.key = x.trim(); f.key.setError(''); } });
    key.setAttribute('data-config-key', '');
    f.key = field(t('cf.key'), key, t('cf.key.hint'));
    var typeSel = selectInput(v.type || 'string', CONFIG_TYPES.map(function (k) { return { value: k, label: t('cf.type.' + k) }; }), function (x) { v.type = x; v.value = x === 'bool' ? false : x === 'text' ? {} : x === 'json' ? {} : x === 'number' ? 0 : ''; v.rules = []; paintValue(); paintRules(); }, { label: t('cf.type'), placeholder: false });
    typeSel.setAttribute('data-config-type', '');
    if (!isNew) { typeSel.disabled = true; typeSel.classList.add('eco-readonly'); }
    f.type = field(t('cf.type'), typeSel, isNew ? t('cf.type.hint') : t('cf.type.locked'));
    var desc = textInput(v.description || '', { label: t('cf.description'), oninput: function (x) { v.description = x; } });
    f.desc = field(t('cf.description'), desc, t('cf.description.hint'), { optional: true });
    var valueHolder = h('div', {});
    f.value = field(t('cf.default'), valueHolder, t('cf.default.hint'));
    var active = toggle(v.isActive !== false, t('cf.active'), function (x) { v.isActive = x; }, 'sm');
    var rulesBox = h('div', { class: 'flex flex-col gap-3', 'data-rules': '' });
    var addRule = button({ variant: 'ghost', size: 'sm', icon: 'plus', label: t('cf.rule.add'), onclick: function () { v.rules.push({ name: 'rule-' + (v.rules.length + 1), value: v.type === 'bool' ? true : clone(v.value), platforms: [], countries: [], tagIds: [], percent: null }); paintRules(); } });
    addRule.setAttribute('data-rule-add', '');
    var rows = [];
    function paintValue() { clear(valueHolder).appendChild(configValueInput(v.type, v.value, function (x) { v.value = x; f.value.setError(''); })); }
    function paintRules() {
      clear(rulesBox);
      rows = v.rules.map(function (r, i) {
        var row = ruleEditor(v.type, r, function () { v.rules.splice(i, 1); paintRules(); }, function (dir) {
          var j = i + dir;
          if (j < 0 || j >= v.rules.length) return;
          var tmp = v.rules[i]; v.rules[i] = v.rules[j]; v.rules[j] = tmp;
          paintRules();
        });
        rulesBox.appendChild(row);
        return row;
      });
      if (!v.rules.length) rulesBox.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('cf.rule.none') }));
    }
    paintValue();
    paintRules();
    var save = button({ variant: 'primary', size: 'sm', label: isNew ? t('create') : t('save'), onclick: function () { doSave(); } });
    save.setAttribute('data-config-save', '');
    function doSave() {
      clear(top);
      var ok = true;
      if (!/^[a-z][a-z0-9_.]{1,63}$/.test(v.key || '')) { f.key.setError(t('code.INVALID_KEY')); ok = false; }
      var bad = configValueValid(v.type, v.value);
      if (bad) { f.value.setError(bad); ok = false; }
      v.rules.forEach(function (r, i) {
        var e = configValueValid(v.type, r.value);
        if (!e && r.percent != null && (isNaN(r.percent) || r.percent < 0 || r.percent > 100)) e = t('code.INVALID_PERCENT');
        rows[i].setError(e || '');
        if (e) ok = false;
      });
      if (!ok) return;
      var payload = { key: v.key, type: v.type, description: v.description || '', value: v.type === 'text' ? trClean(v.value) : v.value, isActive: v.isActive !== false,
        rules: v.rules.map(function (r) { return { name: r.name, value: v.type === 'text' ? trClean(r.value) : r.value, platforms: r.platforms || [], minVersion: r.minVersion || '', maxVersion: r.maxVersion || '', tagIds: r.tagIds || [], countries: r.countries || [], percent: r.percent == null ? null : r.percent }; }) };
      setBusy(save, true, t('saving'));
      baseApi(isNew ? 'POST' : 'PUT', '/remote-config' + (isNew ? '' : '/' + encodeURIComponent(id)), payload).then(function (res) {
        toast('success', t('saved'));
        if (isNew) go('remote-config', res.id);
        else render(res);
      }, function (err) {
        setBusy(save, false);
        if (err.field === 'key') f.key.setError(errorText(err));
        else top.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
      });
    }
    var del = !isNew && write ? button({ variant: 'danger', size: 'sm', icon: 'trash', label: t('delete'), onclick: function () {
      confirmModal({ title: t('cf.delete.title'), text: t('cf.delete.text', { key: v.key }), ok: t('delete'), danger: true }).then(function (yes) {
        if (!yes) return;
        baseApi('DELETE', '/remote-config/' + encodeURIComponent(id)).then(function () { toast('success', t('deleted')); go('remote-config'); }, function (err) { toast('error', errorText(err)); });
      });
    } }) : null;
    var formCard = card(h('div', { class: 'p-4 md:p-6' }, top, h('div', { class: 'eco-form-row' }, f.key, f.type), f.desc, f.value, h('div', { class: 'mb-5' }, active),
      h('div', { class: 'border-t border-secondary-200 dark:border-secondary-700 pt-4 mt-1' },
        h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-1' }, h('h2', { class: 'font-medium', text: t('cf.rules') }), addRule),
        h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mb-3', text: t('cf.rules.hint') }), rulesBox),
      write ? h('div', { class: 'flex justify-between gap-2 pt-4' }, del || h('span', {}), save) : null));
    var side = h('div', { class: 'eco-config-side flex flex-col gap-4' });
    if (!isNew) {
      side.appendChild(previewCard());
      side.appendChild(historyCard(item, function (res) { render(res); }));
    }
    grid.appendChild(h('div', { class: 'eco-config-main' }, formCard));
    grid.appendChild(side);
  }

  function load() {
    skeleton();
    if (isNew) { freshLanguages().then(function () { render({ key: query.key || '', type: 'string', description: '', value: '', rules: [], isActive: true }); }); return; }
    Promise.all([baseApi('GET', '/remote-config/' + encodeURIComponent(id)), freshLanguages()]).then(function (all) { render(all[0]); }, function (err) { clear(grid).appendChild(errorBlock(err, load)); });
  }
  load();
}

function previewCard() {
  var v = { userId: '', platform: 'web', version: '', country: '' };
  var out = h('div', { class: 'flex flex-col gap-1 mt-2', 'data-preview-out': '' });
  var uid = textInput('', { ltr: true, mono: true, label: t('cf.preview.player'), placeholder: t('cf.preview.player.ph'), oninput: function (x) { v.userId = x.trim(); } });
  var plat = selectInput('web', PLATFORMS.map(function (p) { return { value: p, label: has('pf.' + p) ? t('pf.' + p) : p }; }), function (x) { v.platform = x; }, { label: t('cf.rule.platforms'), placeholder: false });
  var ver = textInput('', { ltr: true, mono: true, label: t('cf.preview.version'), placeholder: '1.4.0', oninput: function (x) { v.version = x.trim(); } });
  var ctry = textInput('', { ltr: true, mono: true, label: t('cf.preview.country'), placeholder: 'IR', oninput: function (x) { v.country = x.trim().toUpperCase(); } });
  var run = button({ variant: 'primary', size: 'sm', label: t('cf.preview.run'), onclick: function () {
    setBusy(run, true);
    baseApi('POST', '/remote-config/preview', Object.assign({ lang: lang() }, v)).then(function (res) {
      setBusy(run, false);
      clear(out);
      Object.keys(res.values || {}).sort().forEach(function (k) {
        var val = res.values[k];
        out.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-preview-key': k },
          h('span', { class: 'font-mono text-xs', dir: 'ltr', text: k }), h('span', { class: 'text-secondary-400', text: '=' }),
          h('span', { class: 'font-medium', dir: 'auto', text: typeof val === 'object' ? JSON.stringify(val) : String(val) }),
          tag(res.variants[k] === 'default' ? t('cf.variant.default') : res.variants[k], res.variants[k] === 'default' ? 'secondary' : 'primary'),
          res.buckets && res.buckets[k] != null ? h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: t('cf.bucket', { n: localDigits((res.buckets[k] / 100).toFixed(2)) }) }) : null));
      });
    }, function (err) { setBusy(run, false); clear(out).appendChild(notice('error', errorText(err))); });
  } });
  run.setAttribute('data-preview-run', '');
  return card(h('div', { class: 'p-4', 'data-config-preview': '' }, h('h2', { class: 'font-medium mb-1', text: t('cf.preview') }), h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('cf.preview.hint') }),
    field(t('cf.preview.player'), uid, null, { optional: true }), h('div', { class: 'eco-form-row' }, field(t('cf.rule.platforms'), plat), field(t('cf.preview.version'), ver, null, { optional: true })), field(t('cf.preview.country'), ctry, null, { optional: true }), run, out));
}

function historyCard(item, onrestored) {
  var list = h('ol', { class: 'flex flex-col', 'data-config-history': '' }, h('li', { class: 'p-3' }, bar('70%')));
  baseApi('GET', '/remote-config/' + encodeURIComponent(item.id) + '/history?limit=20').then(function (res) {
    clear(list);
    ((res && res.items) || []).forEach(function (x) {
      var snap = x.snapshot || {};
      var li = h('li', { class: 'flex flex-wrap items-center gap-2 px-3 py-2 border-b border-secondary-200 dark:border-secondary-700 text-sm', 'data-history-version': String(x.version) },
        h('span', { class: 'tabular-nums font-medium', text: t('cf.v', { n: x.version }) }),
        tag(t('cf.action.' + x.action), 'secondary'),
        h('span', { class: 'text-xs text-secondary-500', text: (x.changedBy || '') + ' · ' + fmtDate(x.createdAt, true) }),
        h('span', { class: 'flex-1' }),
        h('span', { class: 'text-xs truncate eco-hist-val', dir: 'auto', text: configValuePreview({ type: item.type, value: snap.value }).textContent }));
      if (x.version !== item.version && x.action !== 'delete' && aclCan('SETTINGS', 1)) {
        var rb = button({ variant: 'link', size: 'sm', icon: 'history', label: t('cf.rollback'), onclick: function () {
          confirmModal({ title: t('cf.rollback.title', { n: x.version }), text: t('cf.rollback.text'), ok: t('cf.rollback') }).then(function (yes) {
            if (!yes) return;
            baseApi('POST', '/remote-config/' + encodeURIComponent(item.id) + '/rollback', { version: x.version }).then(function (r2) { toast('success', t('cf.rolledBack', { n: x.version })); onrestored(r2); }, function (err) { toast('error', errorText(err)); });
          });
        } });
        rb.setAttribute('data-rollback', String(x.version));
        li.appendChild(rb);
      }
      list.appendChild(li);
    });
  }, function (err) { clear(list).appendChild(h('li', { class: 'p-3' }, notice('error', errorText(err)))); });
  return card(h('div', {}, h('h2', { class: 'font-medium px-4 pt-4 pb-2', text: t('cf.history') }), list));
}
