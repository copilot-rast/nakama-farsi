ICONS.tasks = '<path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>';
ICONS['task-groups'] = '<rect x="3" y="4" width="18" height="5" rx="1"/><rect x="3" y="11" width="18" height="5" rx="1"/><path d="M7 20h10"/>';

function minutesText(n) {
  n = Number(n) || 0;
  if (n && n % 1440 === 0) return t('tk.days', { n: fmtNum(n / 1440) });
  if (n && n % 60 === 0) return t('tk.hours', { n: fmtNum(n / 60) });
  return t('tk.minutes', { n: fmtNum(n) });
}

function taskSentence(it, eventName) {
  var parts = [];
  if (it.condition === 'children') parts.push(t('tk.when.children'));
  else {
    parts.push(t('tk.when.' + (it.metric || 'count'), { event: eventName, n: numLocal(it.target) }));
    if (it.valueMin != null && it.valueMax != null) parts.push(t('tk.filter.range', { min: numLocal(it.valueMin), max: numLocal(it.valueMax) }));
    else if (it.valueMin != null) parts.push(t('tk.filter.min', { n: numLocal(it.valueMin) }));
    else if (it.valueMax != null) parts.push(t('tk.filter.max', { n: numLocal(it.valueMax) }));
  }
  if (!it.parentId) {
    parts.push(it.repeat === 'interval' ? t('tk.repeat.interval', { time: minutesText(it.repeatMinutes) }) : t('tk.repeat.' + (it.repeat || 'once')));
    if (it.repeat !== 'once' && it.maxCompletions) parts.push(t('tk.max', { n: fmtNum(it.maxCompletions) }));
  }
  return parts.join(' · ');
}

function taskSentenceCell(it) {
  var box = h('div', { class: 'text-sm text-secondary-700 dark:text-secondary-200 eco-clamp', 'data-task-sentence': '' });
  if (it.condition === 'children' || !it.eventTypeId) { box.appendChild(document.createTextNode(taskSentence(it, ''))); return box; }
  var ref = refName('event-types', it.eventTypeId);
  function paint(et) { clear(box).appendChild(document.createTextNode(taskSentence(it, et ? i18nText(et.name) || et.key : '…'))); }
  if (ref.item !== undefined) paint(ref.item);
  else { box.appendChild(bar('70%')); ref.p.then(paint); }
  return box;
}

function handoutSentence(it) {
  var parts = [t('tg.active', { n: fmtNum(it.activeCount || 1) }), t('tg.order.' + it.orderMode)];
  parts.push(it.handout === 'timer' ? t('tg.handout.timer.after', { time: minutesText(it.handoutMinutes) }) : t('tg.handout.' + it.handout));
  if (it.isRepeatable) parts.push(t('tg.repeatable'));
  return parts.join(' · ');
}

function levelRow() {
  return { type: 'row', fields: [
    { name: 'levelTableId', type: 'ref', resource: 'level-tables', label: 'f.tk.levelTable', hint: 'f.tk.levelTable.hint', optional: true,
      check: function (v, vals) { return !v && (vals.minLevel != null || vals.maxLevel != null) ? t('code.LEVEL_TABLE_REQUIRED') : null; } },
    { name: 'minLevel', type: 'number', label: 'f.tk.minLevel', optional: true, min: 0 },
    { name: 'maxLevel', type: 'number', label: 'f.tk.maxLevel', optional: true, min: 0,
      check: function (v, vals) { return v != null && vals.minLevel != null && Number(v) < Number(vals.minLevel) ? t('code.INVALID_LEVEL_RANGE') : null; } }
  ] };
}

function isEvent(v) { return v.condition !== 'children'; }
function isTop(v) { return !v.parentId; }

RESOURCES.tasks = {
  filters: [
    { name: 'level', label: 'filter.tk.level', options: function () { return [{ value: 'top', label: t('tk.level.top') }, { value: 'sub', label: t('tk.level.sub') }]; } },
    { name: 'groupId', label: 'filter.tk.group', type: 'ref', resource: 'task-groups' },
    { name: 'eventTypeId', label: 'filter.ev.event', type: 'ref', resource: 'event-types' },
    { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'task' } },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '40%' },
    { label: 'col.tk.when', render: taskSentenceCell, skeleton: '55%' },
    { label: 'col.ev.prize', render: function (it) {
      if (!it.prizeId) return h('span', { class: 'text-secondary-400', text: t('tk.noPrize') });
      return h('div', { class: 'flex flex-col gap-1 min-w-0' }, refNameCell('prizes', it.prizeId), h('div', {}, grantTag(it.grantMode)));
    }, width: 170 },
    { label: 'col.tk.place', render: function (it) {
      var box = h('div', { class: 'flex flex-col gap-1 min-w-0 text-sm' });
      if (it.parentId) box.appendChild(h('div', { class: 'flex items-center gap-1 min-w-0' }, tag(t('tk.sub'), 'secondary'), refNameCell('tasks', it.parentId)));
      else if (it.groupId) box.appendChild(h('div', { class: 'flex items-center gap-1 min-w-0' }, tag(t('tk.inGroup'), 'primary'), refNameCell('task-groups', it.groupId)));
      if (it.children) box.appendChild(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('tk.childrenN', { n: fmtNum(it.children) }) }));
      if ((it.dependsOn || []).length) box.appendChild(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('tk.depsN', { n: fmtNum(it.dependsOn.length) }) }));
      if (!box.childNodes.length) box.appendChild(h('span', { class: 'text-secondary-400', text: t('tk.standalone') }));
      return box;
    }, width: 170 },
    { label: 'col.tk.done', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: t('tk.doneN', { n: fmtNum(it.completed || 0), p: fmtNum(it.players || 0) }) }); }, width: 120 },
    { label: 'col.status', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive),
        (it.audienceTagIds || []).length || it.minLevel != null || it.maxLevel != null ? tag(t('audience.only'), 'primary') : null,
        it.endsAt && new Date(it.endsAt) < new Date() ? tag(t('ev.ended'), 'secondary') : null);
    }, width: 130 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.tk.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { label: 'f.tk.name', hint: 'f.tk.name.hint', required: true, max: 128 }),
    tr('description', { label: 'f.tk.description', multiline: true, max: 1000, optional: true }),
    { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
    { type: 'row', fields: [categoryField('task'), tagsField('task')] },
    { type: 'heading', label: 'tk.h.when' },
    { name: 'condition', type: 'select', label: 'f.tk.condition', hint: 'f.tk.condition.hint', required: true, noPlaceholder: true, options: opts('tk.condition.', ['event', 'children']) },
    { name: 'eventTypeId', type: 'ref', resource: 'event-types', label: 'f.ev.event', required: true, when: isEvent },
    { type: 'row', when: isEvent, fields: [
      { name: 'metric', type: 'select', label: 'f.tk.metric', hint: 'f.tk.metric.hint', required: true, noPlaceholder: true, options: opts('tk.metric.', ['count', 'sum_value', 'high_value']), when: isEvent },
      { name: 'target', type: 'number', int: false, label: 'f.tk.target', required: true, min: 0.0001, when: isEvent }
    ] },
    { type: 'row', when: isEvent, fields: [
      { name: 'valueMin', type: 'number', int: false, label: 'f.tk.valueMin', hint: 'f.tk.valueFilter.hint', optional: true, when: isEvent },
      { name: 'valueMax', type: 'number', int: false, label: 'f.tk.valueMax', optional: true, when: isEvent,
        check: function (v, vals) { return v != null && vals.valueMin != null && Number(v) < Number(vals.valueMin) ? t('code.INVALID_RANGE') : null; } }
    ] },
    { name: 'parentId', type: 'ref', resource: 'tasks', filter: { level: 'top' }, label: 'f.tk.parent', hint: 'f.tk.parent.hint', optional: true, when: isEvent },
    { name: 'preview', type: 'static', readonly: true, live: true, render: function (v) {
      if (isEvent(v) && (!v.eventTypeId || v.target == null || v.target === '')) return h('span', {});
      var item = { condition: v.condition, eventTypeId: v.eventTypeId, metric: v.metric, target: parseNumber(v.target, false), valueMin: v.valueMin, valueMax: v.valueMax, repeat: v.repeat, repeatMinutes: v.repeatMinutes, maxCompletions: v.maxCompletions, parentId: v.parentId };
      return h('div', { class: 'eco-rule-preview text-sm rounded-md px-3 py-2', 'data-task-preview': '' }, taskSentenceCell(item));
    } },
    { type: 'heading', label: 'tk.h.repeat', when: isTop },
    { type: 'row', when: isTop, fields: [
      { name: 'repeat', type: 'select', label: 'f.tk.repeat', required: true, noPlaceholder: true, options: opts('tk.repeatopt.', ['once', 'daily', 'weekly', 'interval']), when: isTop,
        check: function (v, vals) { return vals.groupId && v !== 'once' ? t('code.GROUP_TASK_REPEAT') : null; } },
      { name: 'repeatMinutes', type: 'number', label: 'f.tk.repeatMinutes', hint: 'f.tk.repeatMinutes.hint', required: true, min: 1, max: 525600, when: function (v) { return isTop(v) && v.repeat === 'interval'; } },
      { name: 'maxCompletions', type: 'number', label: 'f.tk.maxCompletions', hint: 'f.tk.maxCompletions.hint', optional: true, min: 1, when: function (v) { return isTop(v) && v.repeat !== 'once'; } }
    ] },
    { name: 'groupId', type: 'ref', resource: 'task-groups', label: 'f.tk.group', hint: 'f.tk.group.hint', optional: true, when: isTop },
    { name: 'dependsOn', type: 'refMulti', resource: 'tasks', label: 'f.tk.dependsOn', hint: 'f.tk.dependsOn.hint', optional: true, empty: [] },
    { type: 'heading', label: 'ev.h.prize' },
    { type: 'row', fields: [
      { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', hint: 'f.tk.prize.hint', optional: true },
      { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', hint: 'f.ev.grantMode.hint', required: true, noPlaceholder: true, options: grantOptions }
    ] },
    { name: 'notify', type: 'bool', inline: true, label: 'f.tk.notify', hint: 'f.tk.notify.hint', empty: false },
    { type: 'heading', label: 'ev.h.who', when: isTop },
    { name: 'audienceTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.ev.audience', hint: 'f.ev.audience.hint', optional: true, empty: [], when: isTop },
    Object.assign(levelRow(), { when: isTop }),
    Object.assign(windowRow(), { when: isTop }),
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
  ],
  defaults: function (query) {
    return { key: '', name: {}, description: {}, iconUrl: '', categoryId: null, tagIds: [], condition: 'event', eventTypeId: (query && query.eventTypeId) || null, metric: 'count', target: 1, valueMin: null, valueMax: null,
      repeat: 'once', repeatMinutes: null, maxCompletions: null, parentId: (query && query.parentId) || null, dependsOn: [], groupId: (query && query.groupId) || null, prizeId: null, grantMode: 'claim', notify: false,
      audienceTagIds: [], levelTableId: null, minLevel: null, maxLevel: null, startsAt: null, endsAt: null, isActive: true, sort: 0 };
  },
  fixPayload: function (p) {
    if (p.sort == null) p.sort = 0;
    if (p.condition === 'children') { p.eventTypeId = null; p.parentId = null; p.valueMin = null; p.valueMax = null; p.target = 1; p.metric = 'count'; }
    if (p.parentId) { p.groupId = null; p.repeat = 'once'; p.audienceTagIds = []; p.levelTableId = null; p.minLevel = null; p.maxLevel = null; p.startsAt = null; p.endsAt = null; }
    if (p.repeat !== 'interval') p.repeatMinutes = null;
    if (p.repeat === 'once') p.maxCompletions = null;
    if (p.minLevel == null && p.maxLevel == null) p.levelTableId = null;
    p.dependsOn = p.dependsOn || [];
    return p;
  },
  afterSave: afterEventSave('tasks')
};

RESOURCES['task-groups'] = {
  filters: [
    { name: 'handout', label: 'filter.tg.handout', options: opts('tg.handoutopt.', ['auto', 'request', 'timer']) },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '40%' },
    { label: 'col.tg.rule', render: function (it) { return h('div', { class: 'text-sm text-secondary-700 dark:text-secondary-200 eco-clamp', 'data-group-sentence': '', text: handoutSentence(it) }); }, skeleton: '55%' },
    { label: 'col.tg.tasks', render: function (it) { return h('span', { class: 'tabular-nums', text: t('tg.tasksN', { n: fmtNum((it.taskIds || []).length) }) }); }, width: 120 },
    { label: 'col.ev.players', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.players || 0) }); }, width: 100 },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive), (it.audienceTagIds || []).length || it.minLevel != null || it.maxLevel != null ? tag(t('audience.only'), 'primary') : null); }, width: 130 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.tk.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { label: 'f.tg.name', required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
    categoryField('task'),
    { type: 'heading', label: 'tg.h.handout' },
    { type: 'row', fields: [
      { name: 'activeCount', type: 'number', label: 'f.tg.activeCount', hint: 'f.tg.activeCount.hint', required: true, min: 1, max: 20 },
      { name: 'orderMode', type: 'select', label: 'f.tg.orderMode', required: true, noPlaceholder: true, options: opts('tg.order.', ['ordered', 'random']) }
    ] },
    { type: 'row', fields: [
      { name: 'handout', type: 'select', label: 'f.tg.handout', hint: 'f.tg.handout.hint', required: true, noPlaceholder: true, options: opts('tg.handoutopt.', ['auto', 'request', 'timer']) },
      { name: 'handoutMinutes', type: 'number', label: 'f.tg.handoutMinutes', required: true, min: 1, max: 525600, when: function (v) { return v.handout === 'timer'; } }
    ] },
    { name: 'isRepeatable', type: 'bool', inline: true, label: 'f.tg.repeatable', hint: 'f.tg.repeatable.hint', empty: false },
    { type: 'heading', label: 'tg.h.tasks' },
    { name: 'taskIds', type: 'refMulti', resource: 'tasks', filter: { level: 'top' }, label: 'f.tg.tasks', hint: 'f.tg.tasks.hint', optional: true, empty: [] },
    { type: 'heading', label: 'ev.h.who' },
    playerTagsField(),
    levelRow(),
    windowRow(),
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
  ],
  defaults: function () {
    return { key: '', name: {}, description: {}, iconUrl: '', categoryId: null, activeCount: 1, orderMode: 'ordered', handout: 'auto', handoutMinutes: 60, isRepeatable: false, taskIds: [],
      audienceTagIds: [], levelTableId: null, minLevel: null, maxLevel: null, startsAt: null, endsAt: null, isActive: true, sort: 0 };
  },
  fixPayload: function (p) {
    if (p.sort == null) p.sort = 0;
    if (p.handout !== 'timer') p.handoutMinutes = 0;
    if (p.minLevel == null && p.maxLevel == null) p.levelTableId = null;
    p.taskIds = p.taskIds || [];
    return p;
  },
  afterSave: function () { forgetRefs('task-groups'); forgetRefs('tasks'); }
};

OVERVIEW_CATALOG.push('tasks', 'task-groups');

var TASK_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadTaskCard(id) {
  TASK_CARD.loading = true;
  return api('GET', '/users/' + encodeURIComponent(id) + '/tasks' + qs({ lang: lang() })).then(function (res) {
    TASK_CARD.loading = false;
    TASK_CARD.at = Date.now();
    TASK_CARD.data = res;
    TASK_CARD.err = null;
    var old = document.querySelector('[data-player-tasks]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    TASK_CARD.loading = false;
    TASK_CARD.at = Date.now();
    TASK_CARD.err = err;
    schedule();
  });
}

function taskStateTag(it) {
  var s = it.state;
  var color = s === 'claimed' ? 'success' : s === 'completed' ? 'warning' : s === 'active' ? 'primary' : 'secondary';
  var label = t('tk.state.' + (s === 'claimed' && !it.prize ? 'finished' : s));
  if (s === 'hidden' && it.reason) label = t('tk.reason.' + it.reason);
  return tag(label, color);
}

function taskAction(userId, it, what) {
  var done = { reset: 'tk.reset.done', complete: 'tk.complete.done', claim: 'tk.claim.done' }[what];
  function run() {
    api('POST', '/users/' + encodeURIComponent(userId) + '/tasks/' + encodeURIComponent(it.id) + '/' + what, {}).then(function () {
      toast('success', t(done));
      TASK_CARD.at = 0;
      loadTaskCard(userId);
    }, function (err) { toast('error', errorText(err)); });
  }
  if (what === 'claim') { run(); return; }
  confirmModal({ title: t('tk.' + what + '.title'), text: t('tk.' + what + '.text', { name: it.name }), ok: t('tk.' + what), danger: what === 'reset' }).then(function (yes) { if (yes) run(); });
}

function taskLine(userId, it, write, depth) {
  var progress = it.condition === 'children' ? t('tk.progress.children', { n: fmtNum(it.value), total: fmtNum(it.target) }) : t('tk.progress', { n: numLocal(it.value), total: numLocal(it.target) });
  var line = h('div', { class: 'flex flex-wrap items-center gap-2 py-1.5 text-sm' + (depth ? ' eco-tk-sub' : ''), 'data-pp-task': it.key, 'data-state': it.state },
    h('span', { class: 'font-medium truncate min-w-0', text: it.name || it.key }),
    taskStateTag(it),
    h('span', { class: 'tabular-nums text-secondary-600 dark:text-secondary-300 whitespace-nowrap', 'data-task-progress': '', text: progress }),
    it.repeat && it.repeat !== 'once' && !depth ? h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: t('tk.repeat.' + it.repeat) + (it.completions ? ' · ' + t('tk.completionsN', { n: fmtNum(it.completions) }) : '') }) : null,
    h('span', { class: 'flex-1' }));
  if (write && it.state !== 'hidden' && it.state !== 'locked') {
    var acts = h('div', { class: 'flex flex-wrap items-center gap-1' });
    if (it.state === 'active' && it.status) acts.appendChild(Object.assign(button({ variant: 'link', size: 'sm', label: t('tk.reset'), onclick: function () { taskAction(userId, it, 'reset'); } }), { title: t('tk.reset') }));
    if (it.state === 'completed' && it.claimable) acts.appendChild(button({ variant: 'link', size: 'sm', label: t('tk.claim'), onclick: function () { taskAction(userId, it, 'claim'); } }));
    if (it.state === 'completed' && !it.claimable) acts.appendChild(button({ variant: 'link', size: 'sm', label: t('tk.reset'), onclick: function () { taskAction(userId, it, 'reset'); } }));
    if (it.state === 'active') acts.appendChild(button({ variant: 'link', size: 'sm', label: t('tk.complete'), onclick: function () { taskAction(userId, it, 'complete'); } }));
    Array.prototype.forEach.call(acts.querySelectorAll('button'), function (b) { b.setAttribute('data-task-action', b.textContent.trim()); });
    line.appendChild(acts);
  }
  var wrap = h('div', { class: 'min-w-0' }, line);
  (it.children || []).forEach(function (c) { wrap.appendChild(taskLine(userId, c, write, 1)); });
  return wrap;
}

function playerTasksCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-tasks]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (TASK_CARD.id !== id) { TASK_CARD.id = id; TASK_CARD.data = null; TASK_CARD.err = null; TASK_CARD.at = 0; }
  if (!TASK_CARD.loading && Date.now() - TASK_CARD.at > 20000) loadTaskCard(id);
  var p = TASK_CARD.data;
  var anchor = grid.querySelector('[data-player-activity]') || grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify([p.items, p.groups]) : TASK_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-tasks': '', 'data-key': key, 'data-economy': 'player-tasks', 'data-l10n-skip': '', 'aria-labelledby': 'player-tasks-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-tasks-title', class: 'font-medium', text: t('tk.card') });
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(TASK_CARD.err ? notice('error', errorText(TASK_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { TASK_CARD.at = 0; loadTaskCard(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var write = aclCan('HIRO_ECONOMY', 1);
  var visible = (p.items || []).filter(function (it) { return it.isActive && it.state !== 'hidden'; });
  var hidden = (p.items || []).length - visible.length;
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' }, title,
    h('a', { href: hashFor('tasks'), class: 'eco-link text-sm', text: t('tk.all') })));
  var list = h('div', { class: 'flex flex-col eco-tk-list', 'data-pp-tasks': '' });
  visible.forEach(function (it) { list.appendChild(taskLine(id, it, write, 0)); });
  if (!visible.length) list.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('tk.none') }));
  body.appendChild(list);
  if (hidden) body.appendChild(h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: t('tk.hiddenN', { n: fmtNum(hidden) }) }));
  (p.groups || []).filter(function (g) { return g.isActive; }).forEach(function (g) {
    var head = h('div', { class: 'flex flex-wrap items-center gap-2 text-sm' },
      h('span', { class: 'font-medium', text: g.name || g.key }),
      tag(g.open ? t('tg.round', { n: fmtNum(g.cycle) }) : t('tk.reason.' + (g.reason || 'groupClosed')), g.open ? 'primary' : 'secondary'),
      h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums', text: t('tg.progress', { n: fmtNum(g.done), total: fmtNum(g.total) }) }),
      g.nextAt ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('tg.nextAt', { date: fmtDate(g.nextAt, true) }) }) : null,
      h('span', { class: 'flex-1' }));
    if (write && g.open && g.canRequest) {
      var nx = button({ variant: 'link', size: 'sm', label: t('tg.next'), onclick: function () {
        api('POST', '/users/' + encodeURIComponent(id) + '/task-groups/' + encodeURIComponent(g.id) + '/next', {}).then(function () { toast('success', t('tg.next.done')); TASK_CARD.at = 0; loadTaskCard(id); }, function (err) { toast('error', errorText(err)); });
      } });
      nx.setAttribute('data-group-next', g.key);
      head.appendChild(nx);
    }
    var box = h('div', { class: 'mt-3 border-t border-secondary-200 dark:border-secondary-700 pt-4', 'data-pp-group': g.key }, head);
    (g.tasks || []).forEach(function (it) { box.appendChild(taskLine(id, it, write, 0)); });
    if (g.open && !(g.tasks || []).length) box.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('tg.empty') }));
    body.appendChild(box);
  });
  place();
}
