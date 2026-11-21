if (ACTIVITY_KINDS.indexOf('account') === -1) ACTIVITY_KINDS.push('account');

FRIEND_FIELDS.push(
  { type: 'heading', label: 'fr.privacy' },
  { type: 'row', fields: [
    { name: 'notifyDeclined', type: 'bool', inline: true, label: 'f.fr.notifyDeclined', hint: 'f.fr.notifyDeclined.hint' },
    { name: 'notifyRemoved', type: 'bool', inline: true, label: 'f.fr.notifyRemoved', hint: 'f.fr.notifyRemoved.hint' }
  ] },
  { type: 'row', fields: [
    { name: 'presenceFriendsOnly', type: 'bool', inline: true, label: 'f.fr.presenceFriends', hint: 'f.fr.presenceFriends.hint' },
    { name: 'dmFriendsOnly', type: 'bool', inline: true, label: 'f.fr.dmFriends', hint: 'f.fr.dmFriends.hint' }
  ] },
  { type: 'heading', label: 'fr.day' },
  { type: 'row', fields: [
    { name: 'dayZone', type: 'text', ltr: true, label: 'f.fr.dayZone', hint: 'f.fr.dayZone.hint', optional: true },
    { name: 'dayStartMinute', type: 'number', label: 'f.fr.dayStart', hint: 'f.fr.dayStart.hint', required: true, min: 0, max: 1439 }
  ] },
  { name: 'botDeclineHours', type: 'number', label: 'f.fr.botDecline', hint: 'f.fr.botDecline.hint', required: true, min: 0, max: 8760 }
);

function friendRulesPayload(p) {
  return { dailyRequests: Number(p.dailyRequests) || 0, maxFriends: Number(p.maxFriends) || 0, expiryDays: Number(p.expiryDays) || 0, rejectCooldownHours: Number(p.rejectCooldownHours) || 0,
    requestsBonusMeta: p.requestsBonusMeta || '', friendsBonusMeta: p.friendsBonusMeta || '', matchPolicy: p.matchPolicy || 'priority', overLimit: p.overLimit || 'refuse',
    notifyDeclined: !!p.notifyDeclined, notifyRemoved: !!p.notifyRemoved, presenceFriendsOnly: !!p.presenceFriendsOnly, dmFriendsOnly: !!p.dmFriendsOnly,
    dayZone: (p.dayZone || '').trim(), dayStartMinute: Number(p.dayStartMinute) || 0, botDeclineHours: Number(p.botDeclineHours) || 0 };
}

function unLadder(text) {
  var parts = toLatinDigits(String(text || '')).split(/[\s,،]+/).filter(function (x) { return x !== ''; });
  return parts.map(function (x) { return Number(x); });
}

function usernameFields(values) {
  if (values && values.ladder && values.ladderText == null) values.ladderText = values.ladder.join(', ');
  return [
    { name: 'enabled', type: 'bool', inline: true, label: 'f.un.enabled', hint: 'f.un.enabled.hint' },
    { name: 'firstFree', type: 'bool', inline: true, label: 'f.un.firstFree', hint: 'f.un.firstFree.hint' },
    { type: 'row', fields: [
      { name: 'currency', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.un.currency', hint: 'f.un.currency.hint', optional: true },
      { name: 'cooldownHours', type: 'number', label: 'f.un.cooldown', hint: 'f.un.cooldown.hint', required: true, min: 0, max: 8760 }
    ] },
    { name: 'ladderText', type: 'text', ltr: true, label: 'f.un.ladder', hint: 'f.un.ladder.hint', optional: true, check: function (v) {
      var list = unLadder(v);
      if (list.length > 20) return t('un.err.ladder');
      for (var i = 0; i < list.length; i++) if (!(list[i] >= 1 && list[i] <= 1000000000 && Math.floor(list[i]) === list[i])) return t('un.err.ladder');
      return null;
    } }
  ];
}

var PLAYER_CODE_FIELDS = [
  { type: 'row', fields: [
    { name: 'length', type: 'number', label: 'f.pcs.length', hint: 'f.pcs.length.hint', required: true, min: 4, max: 16 },
    { name: 'charset', type: 'select', label: 'f.pcs.charset', hint: 'f.pcs.charset.hint', required: true, noPlaceholder: true, options: function () { return ['alnum', 'letters', 'digits'].map(function (v) { return { value: v, label: t('pcs.charset.' + v) }; }); } }
  ] }
];

var SUGGEST_FIELDS = [
  { name: 'enabled', type: 'bool', inline: true, label: 'f.fs.enabled', hint: 'f.fs.enabled.hint' },
  { type: 'row', fields: [
    { name: 'visible', type: 'number', label: 'f.fs.visible', hint: 'f.fs.visible.hint', required: true, min: 1, max: 100 },
    { name: 'reserve', type: 'number', label: 'f.fs.reserve', hint: 'f.fs.reserve.hint', required: true, min: 0, max: 100 }
  ] },
  { type: 'heading', label: 'fs.quotas' },
  { type: 'row', fields: [
    { name: 'coplayQuota', type: 'number', label: 'f.fs.coplay', required: true, min: 0, max: 100 },
    { name: 'mutualQuota', type: 'number', label: 'f.fs.mutual', required: true, min: 0, max: 100 }
  ] },
  { type: 'row', fields: [
    { name: 'activityQuota', type: 'number', label: 'f.fs.activity', required: true, min: 0, max: 100 },
    { name: 'levelQuota', type: 'number', label: 'f.fs.level', required: true, min: 0, max: 100 }
  ] },
  { type: 'heading', label: 'fs.sources' },
  { type: 'row', fields: [
    { name: 'lookbackDays', type: 'number', label: 'f.fs.lookback', hint: 'f.fs.lookback.hint', required: true, min: 1, max: 365 },
    { name: 'presenceDays', type: 'number', label: 'f.fs.presenceDays', hint: 'f.fs.presenceDays.hint', required: true, min: 1, max: 60 }
  ] },
  { type: 'row', fields: [
    { name: 'levelTable', type: 'ref', resource: 'level-tables', valueKey: 'key', label: 'f.fs.levelTable', hint: 'f.fs.levelTable.hint', optional: true },
    { name: 'levelDelta', type: 'number', label: 'f.fs.levelDelta', hint: 'f.fs.levelDelta.hint', required: true, min: 0, max: 1000 }
  ] },
  { type: 'heading', label: 'fs.rules' },
  { type: 'row', fields: [
    { name: 'decayAfter', type: 'number', label: 'f.fs.decayAfter', hint: 'f.fs.decayAfter.hint', required: true, min: 0, max: 365 },
    { name: 'candidateMaxFriends', type: 'number', label: 'f.fs.maxFriends', hint: 'f.fs.maxFriends.hint', required: true, min: 0, max: 10000 }
  ] },
  { type: 'row', fields: [
    { name: 'scanMax', type: 'number', label: 'f.fs.scanMax', hint: 'f.fs.scanMax.hint', required: true, min: 10, max: 2000 },
    { name: 'livePatch', type: 'bool', inline: true, label: 'f.fs.livePatch', hint: 'f.fs.livePatch.hint' }
  ] }
];

function socialSettingsPage(container, path, title, desc, fields, id, fix) {
  var page = apiPage(baseApi, path, title, fields, { id: id, desc: desc, fixPayload: function (p) {
    delete p.updatedBy;
    delete p.updatedAt;
    return fix ? fix(p) : p;
  } });
  var wrap = h('div', { class: 'mt-6' }, page.el);
  container.appendChild(wrap);
  page.load();
  return page;
}

var accountRulesBase = accountRulesPage;
accountRulesPage = function (container) {
  accountRulesBase(container);
  var content = container.querySelector('[data-settings-form="accounts"]');
  var host = content ? content.parentNode : container;
  socialSettingsPage(host, '/username-settings', 'un.title', 'un.desc', usernameFields, 'username-policy', function (p) {
    p.ladder = unLadder(p.ladderText);
    delete p.ladderText;
    p.currency = p.currency || '';
    p.cooldownHours = Number(p.cooldownHours) || 0;
    p.enabled = !!p.enabled;
    p.firstFree = !!p.firstFree;
    return p;
  });
  socialSettingsPage(host, '/player-codes/settings', 'pcs.title', 'pcs.desc', PLAYER_CODE_FIELDS, 'player-codes', function (p) {
    delete p.codes;
    p.length = Number(p.length) || 6;
    return p;
  });
};

var friendRulesBase = friendRulesPage;
friendRulesPage = function (container, query, ctx) {
  var list = friendRulesBase(container, query, ctx);
  var rules = container.querySelector('[data-settings-form="friends"]');
  var holder = h('div', {});
  if (rules && rules.nextSibling) rules.parentNode.insertBefore(holder, rules.nextSibling);
  else container.appendChild(holder);
  socialSettingsPage(holder, '/friend-suggestions/settings', 'fs.title', 'fs.desc', SUGGEST_FIELDS, 'friend-suggestions', function (p) {
    Object.keys(p).forEach(function (k) { if (typeof p[k] === 'string' && k !== 'levelTable') p[k] = Number(p[k]) || 0; });
    p.levelTable = p.levelTable || '';
    p.enabled = !!p.enabled;
    p.livePatch = !!p.livePatch;
    return p;
  });
  return list;
};

var SOCIAL_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadSocialCard(id) {
  SOCIAL_CARD.loading = true;
  var enc = encodeURIComponent(id);
  return Promise.all([baseApi('GET', '/players/' + enc + '/username-changes'), baseApi('GET', '/players/' + enc + '/friend-suggestions')]).then(function (all) {
    SOCIAL_CARD.loading = false;
    SOCIAL_CARD.at = Date.now();
    SOCIAL_CARD.data = { names: all[0] || {}, sugg: all[1] || {} };
    SOCIAL_CARD.err = null;
    var old = document.querySelector('[data-player-social]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    SOCIAL_CARD.loading = false;
    SOCIAL_CARD.at = Date.now();
    SOCIAL_CARD.err = err;
    schedule();
  });
}

function renameDialog(id, current) {
  var v = { username: current || '' };
  var input = textInput(v.username, { ltr: true, label: t('psoc.rename.field'), oninput: function (x) { v.username = x; wrap.setError(''); } });
  input.setAttribute('data-rename-input', '');
  var wrap = field(t('psoc.rename.field'), input, t('psoc.rename.hint'));
  return formDialog({
    id: 'player-rename',
    title: t('psoc.rename.title'),
    text: t('psoc.rename.text'),
    icon: 'users',
    ok: t('psoc.rename'),
    body: wrap,
    validate: function () {
      var ok = v.username.trim() !== '';
      wrap.setError(ok ? '' : t('err.required'));
      return ok;
    },
    showError: function (err) {
      if (err && err.field === 'username') { wrap.setError(errorText(err)); return true; }
      return false;
    },
    submit: function () { return baseApi('POST', '/players/' + encodeURIComponent(id) + '/username', { username: v.username.trim() }); }
  });
}

function socialSourceTag(src) {
  var color = { coplay: 'primary', mutual: 'success', activity: 'warning', level: 'secondary' }[src] || 'secondary';
  return tag(t('fs.src.' + src), color);
}

function socialQuote(q) {
  if (!q || !q.enabled) return h('span', { class: 'text-sm text-secondary-500', 'data-un-quote': 'off', text: t('psoc.policyOff') });
  var box = h('div', { class: 'flex flex-col gap-0.5', 'data-un-quote': q.free ? 'free' : 'paid' },
    h('span', { class: 'text-sm', text: q.free ? t('psoc.next.free') : t('psoc.next.cost', { n: fmtNum(q.cost), cur: q.currency }) }),
    h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: t('psoc.paidN', { n: fmtNum(q.paidChanges || 0) }) }));
  if (q.cooldownEndsAt) box.appendChild(h('span', { class: 'text-xs eco-text-error', text: t('psoc.cooldown', { date: fmtDate(q.cooldownEndsAt, true) }) }));
  return box;
}

function socialHistoryTable(items) {
  if (!items.length) return h('p', { class: 'text-sm text-secondary-500', 'data-un-history': 'empty', text: t('psoc.history.empty') });
  var cols = [{ label: t('psoc.col.from') }, { label: t('psoc.col.to') }, { label: t('psoc.col.source'), width: 110 }, { label: t('psoc.col.cost'), width: 110 }, { label: t('psoc.col.when'), width: 160 }];
  var rows = items.slice(0, 10).map(function (it) {
    return { item: it, cells: [h('span', { dir: 'ltr', text: it.oldUsername }), h('span', { dir: 'ltr', class: 'font-medium', text: it.newUsername }),
      tag(t('psoc.src.' + it.source) + (it.admin ? ' · ' + it.admin : ''), it.source === 'admin' ? 'warning' : it.source === 'paid' ? 'primary' : 'success'),
      h('span', { class: 'tabular-nums', text: it.cost ? fmtNum(it.cost) + ' ' + it.currency : '-' }), h('span', { class: 'text-xs', text: fmtDate(it.createdAt, true) })] };
  });
  var tb = table(cols, rows);
  tb.setAttribute('data-un-history', String(items.length));
  return tb;
}

function socialSuggestTable(list, attr) {
  var cols = [{ label: t('psoc.col.player') }, { label: t('psoc.col.source'), width: 120 }, { label: t('psoc.col.score'), width: 100 }, { label: t('psoc.col.shows'), width: 80 }, { label: t('psoc.col.played'), width: 150 }];
  var rows = list.slice(0, 20).map(function (it) {
    return { item: it, cells: [
      h('a', { href: '#/players/' + it.userId, class: 'eco-link', dir: 'ltr', text: it.username + (it.code ? ' · ' + it.code : '') }),
      socialSourceTag(it.source),
      h('span', { class: 'tabular-nums', text: it.source === 'coplay' ? '-' : fmtNum(Math.round((it.score || 0) * 100) / 100) }),
      h('span', { class: 'tabular-nums', text: fmtNum(it.shows || 0) }),
      h('span', { class: 'text-xs', text: it.lastPlayedAt ? fmtDate(it.lastPlayedAt, true) : '-' })] };
  });
  var tb = table(cols, rows);
  tb.setAttribute(attr, String(list.length));
  return tb;
}

function playerSocialCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-social]');
  if (!id || !token() || !aclCan('ACCOUNT', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (SOCIAL_CARD.id !== id) { SOCIAL_CARD.id = id; SOCIAL_CARD.data = null; SOCIAL_CARD.err = null; SOCIAL_CARD.at = 0; }
  if (!SOCIAL_CARD.loading && Date.now() - SOCIAL_CARD.at > 20000) loadSocialCard(id);
  var d = SOCIAL_CARD.data;
  var anchor = grid.querySelector('[data-player-chat-quota]') || grid.querySelector('[data-player-friend-gifts]') || grid.querySelector('[data-player-offers]') || grid.querySelector('[data-player-chance]') || grid.querySelector('[data-player-achievements]') || grid.querySelector('[data-player-items]') || grid.querySelector('[data-player-tasks]') || grid.querySelector('[data-player-activity]') || grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (d ? JSON.stringify([d.names.code, d.names.username, (d.names.items || []).length, d.names.quote, d.sugg.generatedAt, (d.sugg.visible || []).length, (d.sugg.reserve || []).length, (d.sugg.coplay || []).length]) : SOCIAL_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-social': '', 'data-key': key, 'data-economy': 'player-social', 'data-l10n-skip': '', 'aria-labelledby': 'player-social-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-social-title', class: 'font-medium flex items-center gap-2' }, icon('users', 'h-4 w-4'), h('span', { text: t('psoc.title') }));
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!d) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(SOCIAL_CARD.err ? notice('error', errorText(SOCIAL_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { SOCIAL_CARD.at = 0; loadSocialCard(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var write = aclCan('ACCOUNT', 1);
  var reload = function () { SOCIAL_CARD.at = 0; loadSocialCard(id); if (typeof PLAYER_CARD !== 'undefined') PLAYER_CARD.at = 0; };
  var actions = h('div', { class: 'flex flex-wrap items-center gap-1' });
  if (write) {
    var rename = button({ variant: 'ghost', size: 'sm', icon: 'players', label: t('psoc.rename'), onclick: function () {
      renameDialog(id, d.names.username).then(function (res) { if (res) { toast('success', t('psoc.rename.done', { name: res.username })); reload(); } });
    } });
    rename.setAttribute('data-ps-action', 'rename');
    var regen = button({ variant: 'ghost', size: 'sm', icon: 'refresh', label: t('psoc.regen'), onclick: function () {
      setBusy(regen, true, t('saving'));
      baseApi('POST', '/players/' + encodeURIComponent(id) + '/friend-suggestions/regenerate', {}).then(function () { toast('success', t('psoc.regen.done')); reload(); }, function (err) { setBusy(regen, false); toast('error', errorText(err)); });
    } });
    regen.setAttribute('data-ps-action', 'regenerate');
    actions.appendChild(rename);
    actions.appendChild(regen);
  }
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-3' }, title, actions));
  var s = d.sugg || {};
  body.appendChild(h('div', { class: 'eco-pc-grid mb-3' },
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('psoc.code') }),
      h('div', { class: 'text-lg font-semibold', dir: 'ltr', 'data-ps-code': '', text: d.names.code || '-' })),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('psoc.next') }),
      socialQuote(d.names.quote)),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('psoc.sugg') }),
      h('div', { class: 'text-sm tabular-nums', 'data-ps-sugg': '' }, s.generatedAt ? t('psoc.sugg.counts', { v: fmtNum((s.visible || []).length), r: fmtNum((s.reserve || []).length) }) : t('psoc.sugg.none')),
      s.generatedAt ? h('div', { class: 'text-xs text-secondary-500', text: t('psoc.sugg.meta', { cycle: s.cycle, when: fmtDate(s.generatedAt, true) }) + (s.generatedBy ? ' · ' + s.generatedBy : '') }) : null)));
  body.appendChild(h('h3', { class: 'text-sm font-medium mt-2 mb-2', text: t('psoc.history') }));
  body.appendChild(socialHistoryTable(d.names.items || []));
  if ((s.visible || []).length) {
    body.appendChild(h('h3', { class: 'text-sm font-medium mt-4 mb-2', text: t('psoc.visible') }));
    body.appendChild(socialSuggestTable(s.visible, 'data-ps-visible'));
  }
  if ((s.reserve || []).length) {
    body.appendChild(h('h3', { class: 'text-sm font-medium mt-4 mb-2', text: t('psoc.reserve') }));
    body.appendChild(socialSuggestTable(s.reserve, 'data-ps-reserve'));
  }
  if ((s.coplay || []).length) {
    body.appendChild(h('h3', { class: 'text-sm font-medium mt-4 mb-2', text: t('psoc.coplay') }));
    var cols = [{ label: t('psoc.col.player') }, { label: t('psoc.col.game'), width: 120 }, { label: t('psoc.col.matches'), width: 90 }, { label: t('psoc.col.played'), width: 150 }];
    var tb = table(cols, s.coplay.map(function (c) {
      return { item: c, cells: [h('a', { href: '#/players/' + c.userId, class: 'eco-link', dir: 'ltr', text: c.username || c.userId }), h('span', { dir: 'ltr', text: c.gameKey || '-' }),
        h('span', { class: 'tabular-nums', text: fmtNum(c.matches || 0) }), h('span', { class: 'text-xs', text: fmtDate(c.lastAt, true) })] };
    }));
    tb.setAttribute('data-ps-coplay', String(s.coplay.length));
    body.appendChild(tb);
  }
  place();
}
