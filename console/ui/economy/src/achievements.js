PH_ICONS.achievements = '<path d="M216,96A88,88,0,1,0,72,163.83V240a8,8,0,0,0,11.58,7.16L128,225l44.43,22.21A8.07,8.07,0,0,0,176,248a8,8,0,0,0,8-8V163.83A87.85,87.85,0,0,0,216,96ZM56,96a72,72,0,1,1,72,72A72.08,72.08,0,0,1,56,96ZM168,227.06l-36.43-18.21a8,8,0,0,0-7.16,0L88,227.06V174.37a87.89,87.89,0,0,0,80,0ZM128,152A56,56,0,1,0,72,96,56.06,56.06,0,0,0,128,152Zm0-96A40,40,0,1,1,88,96,40,40,0,0,1,128,56Z"/>';
PH_ICONS.medalDuo = '<path d="M176,96a48,48,0,1,1-48-48A48,48,0,0,1,176,96Z" opacity="0.2"/>' + PH_ICONS.achievements;
PH_ICONS.lockDuo = '<path d="M208,96V208a8,8,0,0,1-8,8H56a8,8,0,0,1-8-8V96a8,8,0,0,1,8-8H200A8,8,0,0,1,208,96Z" opacity="0.2"/><path d="M208,80H176V56a48,48,0,0,0-96,0V80H48A16,16,0,0,0,32,96V208a16,16,0,0,0,16,16H208a16,16,0,0,0,16-16V96A16,16,0,0,0,208,80ZM96,56a32,32,0,0,1,64,0V80H96ZM208,208H48V96H208V208Zm-68-56a12,12,0,1,1-12-12A12,12,0,0,1,140,152Z"/>';

if (ACTIVITY_KINDS.indexOf('achievement') === -1) ACTIVITY_KINDS.push('achievement');
ACTIVITY_COLOR.achievement = 'warning';
ACTIVITY_ICON.achievement = 'achievements';
if (CATEGORY_KEY_TYPES.indexOf('achievement') === -1) CATEGORY_KEY_TYPES.push('achievement');
if (TAG_KEY_TYPES.indexOf('achievement') === -1) TAG_KEY_TYPES.splice(TAG_KEY_TYPES.indexOf('general'), 0, 'achievement');

var ACH_SOURCES = ['event', 'task', 'manual'];



function isEventAch(v) { return v.source === 'event'; }

function achTierLine(it) {
  var tiers = it.tiers || [];
  if (it.source !== 'event') return '';
  return tiers.map(function (r, i) { return t('ac.tier.at', { name: i18nText(r.name) || t('ac.tier.n', { n: fmtNum(r.tier || i + 1) }), n: numLocal(r.threshold) }); }).join(lang() === 'fa' ? '، ' : ' · ');
}

function achSentence(it, name) {
  if (it.source === 'task') return t(it.taskGroupId ? 'ac.when.group' : 'ac.when.task', { name: name });
  if (it.source !== 'event') return t('ac.when.manual');
  var tiers = it.tiers || [];
  var first = tiers.length ? tiers[0].threshold : '';
  var parts = [t('ac.when.' + (it.metric || 'count'), { event: name, n: numLocal(first) })];
  if (it.valueMin != null && it.valueMax != null) parts.push(t('ac.filter.range', { min: numLocal(it.valueMin), max: numLocal(it.valueMax) }));
  else if (it.valueMin != null) parts.push(t('ac.filter.min', { n: numLocal(it.valueMin) }));
  else if (it.valueMax != null) parts.push(t('ac.filter.max', { n: numLocal(it.valueMax) }));
  return parts.join(' · ');
}

function achSentenceCell(it) {
  var box = h('div', { class: 'text-sm text-secondary-700 dark:text-secondary-200 eco-clamp', 'data-ach-sentence': '' });
  var res = it.source === 'event' ? ['event-types', it.eventTypeId] : it.source === 'task' ? (it.taskGroupId ? ['task-groups', it.taskGroupId] : ['tasks', it.taskId]) : null;
  if (!res || !res[1]) { box.appendChild(document.createTextNode(achSentence(it, ''))); return box; }
  var ref = refName(res[0], res[1]);
  function paint(x) { clear(box).appendChild(document.createTextNode(achSentence(it, x ? i18nText(x.name) || x.key : '…'))); }
  if (ref.item !== undefined) paint(ref.item);
  else { box.appendChild(bar('70%')); ref.p.then(paint); }
  return box;
}

function achBadge(it, unlocked) {
  var tile = h('div', { class: 'h-9 w-9 shrink-0 rounded-md overflow-hidden flex items-center justify-center eco-ach-tile', 'data-rarity': it.rarity || 'common', 'data-locked': unlocked === false ? '' : null });
  var src = unlocked === false && it.lockedIconUrl ? it.lockedIconUrl : it.iconUrl;
  if (src) tile.appendChild(img(src, 'achievements'));
  else tile.appendChild(icon(unlocked === false ? 'lockDuo' : 'medalDuo', 'h-5 w-5'));
  return tile;
}

RESOURCES.achievements = {
  filters: [
    { name: 'source', label: 'filter.ac.source', options: opts('ac.source.', ACH_SOURCES) },
    { name: 'rarity', label: 'filter.it.rarity', options: opts('rarity.', RARITIES) },
    { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'achievement' } },
    { name: 'isHidden', label: 'filter.ac.hidden', options: function () { return [{ value: 'true', label: t('yes') }, { value: 'false', label: t('no') }]; } },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) {
      var name = i18nText(it.name) || '';
      return h('div', { class: 'flex items-center gap-3 min-w-0' }, achBadge(it),
        h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: name }), h('div', { class: 'text-secondary-500' }, mono(it.key))));
    }, skeleton: '40%' },
    { label: 'col.ac.when', render: function (it) {
      var box = h('div', { class: 'flex flex-col gap-1 min-w-0' }, achSentenceCell(it));
      var line = achTierLine(it);
      if (line) box.appendChild(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums', 'data-ach-tiers': '', text: line }));
      return box;
    }, skeleton: '55%' },
    { label: 'col.ac.tiers', render: function (it) {
      var tiers = it.tiers || [];
      var prizes = tiers.filter(function (r) { return r.prizeId; });
      var box = h('div', { class: 'flex flex-col gap-1 min-w-0 text-sm' }, h('span', { class: 'tabular-nums', text: t('ac.tiersN', { n: fmtNum(tiers.length) }) }));
      if (!prizes.length) box.appendChild(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ac.noPrize') }));
      else box.appendChild(h('div', { class: 'min-w-0' }, refNameCell('prizes', prizes[0].prizeId)));
      return box;
    }, width: 170 },
    { label: 'col.ac.rarity', render: function (it) {
      return h('div', { class: 'flex flex-wrap items-center gap-1' }, rarityTag(it.rarity), h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums whitespace-nowrap', text: t('ac.pointsN', { n: fmtNum(it.points || 0) }) }));
    }, width: 160 },
    { label: 'col.ac.players', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: t('ac.unlockedN', { n: fmtNum(it.unlocked || 0) }) }); }, width: 110 },
    { label: 'col.status', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive), it.isHidden ? tag(t('ac.hidden'), 'secondary') : null, tag(t('ac.source.' + it.source), it.source === 'event' ? 'primary' : it.source === 'task' ? 'success' : 'secondary'));
    }, width: 160 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    tr('name', { label: 'f.ac.name', required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    tr('hint', { label: 'f.hint', hint: 'f.ac.hint.hint', max: 300, optional: true }),
    { type: 'heading', label: 'ac.h.look' },
    { type: 'row', fields: [
      { name: 'rarity', type: 'select', label: 'f.ac.rarity', required: true, noPlaceholder: true, options: opts('rarity.', RARITIES) },
      { name: 'points', type: 'number', label: 'f.ac.points', hint: 'f.ac.points.hint', min: 0, max: 100000, empty: 0 }
    ] },
    { type: 'row', fields: [categoryField('achievement'), tagsField('achievement')] },
    { type: 'row', fields: [
      { name: 'iconUrl', type: 'image', label: 'f.ac.icon', optional: true },
      { name: 'lockedIconUrl', type: 'image', label: 'f.ac.lockedIcon', optional: true }
    ] },
    { name: 'isHidden', type: 'bool', inline: true, label: 'f.ac.hidden', hint: 'f.ac.hidden.hint', empty: false },
    { type: 'heading', label: 'ac.h.when' },
    { name: 'source', type: 'select', label: 'f.ac.source', hint: 'f.ac.source.hint', required: true, noPlaceholder: true, options: opts('ac.source.', ACH_SOURCES) },
    { name: 'eventTypeId', type: 'ref', resource: 'event-types', label: 'f.ev.event', required: true, when: isEventAch },
    { type: 'row', when: isEventAch, fields: [
      { name: 'metric', type: 'select', label: 'f.ac.metric', hint: 'f.tk.metric.hint', required: true, noPlaceholder: true, options: opts('tk.metric.', ['count', 'sum_value', 'high_value']), when: isEventAch },
      { name: 'valueMin', type: 'number', int: false, label: 'f.tk.valueMin', hint: 'f.tk.valueFilter.hint', optional: true, when: isEventAch },
      { name: 'valueMax', type: 'number', int: false, label: 'f.tk.valueMax', optional: true, when: isEventAch,
        check: function (v, vals) { return v != null && vals.valueMin != null && Number(v) < Number(vals.valueMin) ? t('code.INVALID_RANGE') : null; } }
    ] },
    { type: 'row', when: function (v) { return v.source === 'task'; }, fields: [
      { name: 'taskId', type: 'ref', resource: 'tasks', filter: { level: 'top' }, label: 'f.ac.task', hint: 'f.ac.task.hint', optional: true, when: function (v) { return v.source === 'task'; },
        check: function (v, vals) { return !v && !vals.taskGroupId ? t('code.TASK_REQUIRED') : v && vals.taskGroupId ? t('code.TASK_AND_GROUP') : null; } },
      { name: 'taskGroupId', type: 'ref', resource: 'task-groups', label: 'f.ac.group', hint: 'f.ac.group.hint', optional: true, when: function (v) { return v.source === 'task'; } }
    ] },
    { name: 'preview', type: 'static', readonly: true, live: true, render: function (v) {
      if (v.source === 'event' && !v.eventTypeId) return h('span', {});
      var item = { source: v.source, eventTypeId: v.eventTypeId, metric: v.metric, valueMin: v.valueMin, valueMax: v.valueMax, taskId: v.taskId, taskGroupId: v.taskGroupId, tiers: v.source === 'event' ? v.tiers : [] };
      return h('div', { class: 'eco-rule-preview text-sm rounded-md px-3 py-2 flex flex-col gap-1', 'data-ach-preview': '' }, achSentenceCell(item),
        v.source === 'event' && (v.tiers || []).length ? h('span', { class: 'text-xs tabular-nums', text: achTierLine(item) }) : null);
    } },
    { type: 'heading', label: 'ac.h.tiers', when: isEventAch },
    { name: 'tiers', type: 'rows', addLabel: 'ac.tier.add', emptyText: 'ac.tier.empty', min: 1, max: 10, when: isEventAch, empty: [], refresh: true,
      newRow: function (rows) {
        var last = rows[rows.length - 1];
        return { threshold: last ? Number(last.threshold || 0) * 3 || 10 : 1, name: {}, prizeId: null, grantMode: 'claim' };
      },
      check: function (rows) {
        for (var i = 1; i < rows.length; i++) if (Number(rows[i].threshold) <= Number(rows[i - 1].threshold)) return { index: i, message: t('ac.thresholds.order') };
        return null;
      },
      columns: [
        { name: 'threshold', type: 'number', int: false, label: 'f.ac.tier.threshold', required: true, min: 0.0001 },
        { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', optional: true },
        { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', options: grantOptions },
        { name: 'name', type: 'i18n', label: 'f.ac.tier.name', max: 64, wide: true }
      ] },
    { name: 'tiersHint', type: 'static', readonly: true, when: isEventAch, render: function () { return h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('ac.tiers.hint') }); } },
    { type: 'heading', label: 'ac.h.prize', when: function (v) { return !isEventAch(v); } },
    { type: 'row', when: function (v) { return !isEventAch(v); }, fields: [
      { name: 'unlockPrizeId', type: 'ref', resource: 'prizes', label: 'f.prize', optional: true, when: function (v) { return !isEventAch(v); } },
      { name: 'unlockGrantMode', type: 'select', label: 'f.ev.grantMode', hint: 'f.ev.grantMode.hint', noPlaceholder: true, options: grantOptions, when: function (v) { return !isEventAch(v); } }
    ] },
    { name: 'notify', type: 'bool', inline: true, label: 'f.ac.notify', hint: 'f.ac.notify.hint', empty: false },
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
  ],
  defaults: function (query) {
    return { key: '', isActive: true, name: {}, description: {}, hint: {}, rarity: 'common', points: 10, categoryId: null, tagIds: [], iconUrl: '', lockedIconUrl: '', isHidden: false,
      source: (query && query.source) || 'event', eventTypeId: (query && query.eventTypeId) || null, metric: 'count', valueMin: null, valueMax: null, taskId: null, taskGroupId: null,
      tiers: [{ threshold: 1, name: {}, prizeId: null, grantMode: 'claim' }], unlockPrizeId: null, unlockGrantMode: 'claim', notify: false, sort: 0 };
  },
  toForm: function (v) {
    var first = (v.tiers || [])[0] || {};
    v.unlockPrizeId = first.prizeId || null;
    v.unlockGrantMode = first.grantMode || 'claim';
    return v;
  },
  fixPayload: function (p) {
    if (p.sort == null) p.sort = 0;
    if (p.points == null) p.points = 0;
    if (!p.iconUrl) p.iconUrl = '';
    if (!p.lockedIconUrl) p.lockedIconUrl = '';
    p.categoryId = p.categoryId || '';
    p.tagIds = p.tagIds || [];
    if (p.source === 'event') {
      p.taskId = ''; p.taskGroupId = '';
      p.tiers = (p.tiers || []).map(function (r, i) { return { tier: i + 1, threshold: r.threshold, name: r.name || {}, prizeId: r.prizeId || '', grantMode: r.grantMode || 'claim' }; });
    } else {
      p.eventTypeId = ''; p.valueMin = null; p.valueMax = null; p.metric = 'count';
      if (p.source !== 'task') { p.taskId = ''; p.taskGroupId = ''; }
      p.taskId = p.taskId || ''; p.taskGroupId = p.taskGroupId || '';
      p.tiers = [{ tier: 1, threshold: 1, name: {}, prizeId: p.unlockPrizeId || '', grantMode: p.unlockGrantMode || 'claim' }];
    }
    delete p.unlockPrizeId;
    delete p.unlockGrantMode;
    delete p.preview;
    delete p.tiersHint;
    return p;
  },
  afterSave: function () { forgetRefs('achievements'); }
};

OVERVIEW_CATALOG.push('achievements');

var PAC_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadAchCard(id) {
  PAC_CARD.loading = true;
  return api('GET', '/users/' + encodeURIComponent(id) + '/achievements' + qs({ lang: lang() })).then(function (res) {
    PAC_CARD.loading = false;
    PAC_CARD.at = Date.now();
    PAC_CARD.data = res;
    PAC_CARD.err = null;
    var old = document.querySelector('[data-player-achievements]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    PAC_CARD.loading = false;
    PAC_CARD.at = Date.now();
    PAC_CARD.err = err;
    schedule();
  });
}

function reloadAch(id) { PAC_CARD.at = 0; loadAchCard(id); }

function achStateTag(it) {
  var color = it.state === 'completed' ? 'success' : it.state === 'unlocked' ? 'primary' : 'secondary';
  return tag(t('pac.state.' + it.state), color);
}

function achByText(it) {
  if (!it.unlocked || !it.unlockedBy) return '';
  var by = t('pac.by.' + it.unlockedBy, { name: it.unlockedRef || '' });
  return it.unlockedAt ? by + (lang() === 'fa' ? '، ' : ' · ') + fmtDate(it.unlockedAt, true) : by;
}

function achLine(userId, it, write) {
  var tierText = it.tiered || it.maxTier > 1 ? t('pac.tier', { n: fmtNum(it.tier || 0), max: fmtNum(it.maxTier || 1) }) : '';
  var pct = Math.round((Number(it.progress) || 0) * 100);
  var progress = it.source === 'event' && it.next ? t('pac.progress', { n: numLocal(it.value), total: numLocal(it.next.threshold) }) : '';
  var line = h('div', { class: 'flex flex-wrap items-center gap-2 py-2 text-sm', 'data-pac': it.key, 'data-state': it.state },
    achBadge(it, !!it.unlocked),
    h('div', { class: 'min-w-0 flex flex-col' },
      h('span', { class: 'font-medium truncate', text: it.name || it.key }),
      achByText(it) ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', 'data-pac-by': '', text: achByText(it) }) : null),
    achStateTag(it),
    rarityTag(it.rarity),
    it.hidden ? tag(t('ac.secret'), 'secondary') : null,
    it.isActive === false ? tag(t('inv.inactive'), 'secondary') : null,
    tierText ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums whitespace-nowrap', 'data-pac-tier': '', text: tierText }) : null,
    it.source === 'event' && it.next ? h('span', { class: 'eco-tk-meter', 'aria-hidden': 'true' }, h('span', { style: 'inline-size:' + pct + '%' })) : null,
    progress ? h('span', { class: 'tabular-nums text-secondary-600 dark:text-secondary-300 whitespace-nowrap', 'data-pac-progress': '', text: progress }) : null,
    h('span', { class: 'flex-1' }));
  if (write) {
    var acts = h('div', { class: 'flex flex-wrap items-center gap-1' });
    if ((it.claimable || []).length) {
      var cb = button({ variant: 'link', size: 'sm', label: t('pac.claim'), onclick: function () {
        api('POST', '/users/' + encodeURIComponent(userId) + '/achievements/' + encodeURIComponent(it.id) + '/claim', {}).then(function () { toast('success', t('pac.claim.done')); reloadAch(userId); }, function (err) { toast('error', errorText(err)); });
      } });
      cb.setAttribute('data-pac-claim', it.key);
      acts.appendChild(cb);
    }
    if (it.unlocked) {
      var rb = button({ variant: 'link', size: 'sm', label: t('pac.revoke'), onclick: function () {
        confirmModal({ title: t('pac.revoke.title'), text: t('pac.revoke.text', { name: it.name || it.key }), ok: t('pac.revoke'), danger: true }).then(function (yes) {
          if (!yes) return;
          api('DELETE', '/users/' + encodeURIComponent(userId) + '/achievements/' + encodeURIComponent(it.id)).then(function () { toast('success', t('pac.revoke.done')); reloadAch(userId); }, function (err) { toast('error', errorText(err)); });
        });
      } });
      rb.setAttribute('data-pac-revoke', it.key);
      acts.appendChild(rb);
    }
    if (acts.childNodes.length) line.appendChild(acts);
  }
  return line;
}

function playerAchievementsCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-achievements]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (PAC_CARD.id !== id) { PAC_CARD.id = id; PAC_CARD.data = null; PAC_CARD.err = null; PAC_CARD.at = 0; }
  if (!PAC_CARD.loading && Date.now() - PAC_CARD.at > 20000) loadAchCard(id);
  var p = PAC_CARD.data;
  var anchor = grid.querySelector('[data-player-items]') || grid.querySelector('[data-player-tasks]') || grid.querySelector('[data-player-activity]') || grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify(p.items) : PAC_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-achievements': '', 'data-key': key, 'data-economy': 'player-achievements', 'data-l10n-skip': '', 'aria-labelledby': 'player-achievements-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-achievements-title', class: 'font-medium flex items-center gap-2' }, icon('achievements', 'h-4 w-4'), h('span', { text: t('pac.card') }));
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(PAC_CARD.err ? notice('error', errorText(PAC_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { reloadAch(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var write = aclCan('HIRO_ECONOMY', 1);
  var items = p.items || [];
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' }, title,
    h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300 tabular-nums', 'data-pac-counts': '', text: t('pac.counts', { n: fmtNum(p.unlocked || 0), total: fmtNum(p.total || 0), p: fmtNum(p.points || 0) }) }),
    p.ready ? tag(t('pac.waiting', { n: fmtNum(p.ready) }), 'warning') : null,
    h('div', { class: 'flex flex-wrap items-center gap-3' },
      aclCan('ACCOUNT', 0) ? h('a', { href: hashFor('activity', null, { userId: id, kind: 'achievement' }), class: 'eco-link text-sm', 'data-pac-log': '', text: t('pac.log') }) : null,
      h('a', { href: hashFor('achievements'), class: 'eco-link text-sm', text: t('pac.all') }))));
  var sorted = items.slice().sort(function (a, b) { return (b.unlocked ? 1 : 0) - (a.unlocked ? 1 : 0); });
  var list = h('div', { class: 'flex flex-col eco-tk-list', 'data-pac-list': '' });
  sorted.forEach(function (it) { list.appendChild(achLine(id, it, write)); });
  if (!items.length) list.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('pac.none') }));
  body.appendChild(list);
  if (write && items.length) {
    var pick = '';
    var tierNum = null;
    var picker = refPicker({ resource: 'achievements', value: '', label: t('pac.give.pick'), placeholder: t('pac.give.pick'), onchange: function (v) { pick = v || ''; } });
    var tierIn = numberInput(null, { int: true, label: t('pac.give.tier'), oninput: function (v) { tierNum = v === '' ? null : parseNumber(v, true); } });
    tierIn.setAttribute('data-pac-tier-input', '');
    var give = button({ variant: 'primary', size: 'sm', icon: 'plus', label: t('pac.give'), onclick: function () {
      if (!pick) { toast('error', t('pac.give.pick')); return; }
      var bodyIn = { achievementId: pick };
      if (tierNum) bodyIn.tier = tierNum;
      api('POST', '/users/' + encodeURIComponent(id) + '/achievements', bodyIn).then(function () { toast('success', t('pac.give.done')); reloadAch(id); }, function (err) { toast('error', errorText(err)); });
    } });
    give.setAttribute('data-pac-give', '');
    body.appendChild(h('div', { class: 'mt-3 border-t border-secondary-200 dark:border-secondary-700 pt-4 flex flex-wrap items-end gap-2', 'data-pac-give-row': '' },
      h('div', { class: 'min-w-0 flex-1 eco-inv-pick' }, picker),
      h('div', { style: 'inline-size:6rem' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('pac.give.tier') + ' (' + t('optional') + ')' }), tierIn),
      give));
  }
  place();
}
