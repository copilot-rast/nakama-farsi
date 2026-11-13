ICONS['event-types'] = '<path d="M13 2 3 14h9l-1 8 10-12h-9z"/>';
ICONS['event-rules'] = '<circle cx="12" cy="8" r="6"/><path d="M8.2 13.2 7 22l5-3 5 3-1.2-8.8"/>';
ICONS['meta-rules'] = '<path d="M3 3v18h18"/><path d="m7 15 4-4 3 3 5-6"/>';
ICONS['level-tables'] = '<path d="M4 20h4v-6H4zM10 20h4V9h-4zM16 20h4V4h-4z"/>';
ICONS.streaks = '<path d="M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.4-.5-2-1-3-1.1-2.1-.2-4 2-6 .5 2.5 2 4.9 4 6.5 2 1.6 3 3.5 3 5.5a7 7 0 1 1-14 0c0-1.2.4-2.3 1-3.4A2.5 2.5 0 0 0 8.5 14.5z"/>';
ICONS['event-log'] = '<path d="M12 8v4l3 2"/><circle cx="12" cy="12" r="9"/>';

var EVENT_KEY_RE = /^[a-z][a-z0-9_.]{1,47}$/;
var REF_NAMES = {};

function refName(resource, id, call) {
  var key = resource + ':' + id;
  if (!REF_NAMES[key]) {
    REF_NAMES[key] = { item: undefined };
    REF_NAMES[key].p = (call || api)('GET', '/' + resource + '/' + encodeURIComponent(id)).then(function (it) {
      REF_NAMES[key].item = it;
      refCache[resource + ':id:' + id] = it;
      return it;
    }, function () { REF_NAMES[key].item = null; return null; });
  }
  return REF_NAMES[key];
}

function refNameCell(resource, id, opts) {
  opts = opts || {};
  var box = h('span', { class: 'min-w-0 truncate inline-flex items-center gap-1.5', 'data-ref-name': resource });
  if (!id) { box.appendChild(h('span', { class: 'text-secondary-400', text: t('none') })); return box; }
  var ref = refName(resource, id, opts.api);
  function paint(it) {
    clear(box);
    var name = it ? (i18nText(it.name) || it.key) : String(id).slice(0, 8);
    box.appendChild(h('span', { class: 'truncate', text: name }));
    if (opts.withKey && it && it.key) box.appendChild(h('span', { class: 'font-mono text-xs text-secondary-500', dir: 'ltr', text: it.key }));
  }
  if (ref.item !== undefined) paint(ref.item);
  else {
    box.appendChild(bar('5rem'));
    ref.p.then(paint);
  }
  return box;
}

function forgetRefs(resource) {
  Object.keys(REF_NAMES).forEach(function (k) { if (k.indexOf(resource + ':') === 0) delete REF_NAMES[k]; });
}

function numLocal(v) {
  return v === null || v === undefined || v === '' ? '' : fmtNum(v);
}

function ruleSentence(it, eventName) {
  var parts = [t('rs.when.' + it.measure, { event: eventName, n: numLocal(it.threshold), period: t('rs.period.' + it.period) })];
  parts.push(it.repeat === 'every' ? t('rs.repeat.every', { n: numLocal(it.threshold) }) : t('rs.repeat.' + it.repeat + (it.repeat === 'period' ? '.' + it.period : '')));
  if (it.maxClaims) parts.push(t('rs.max', { n: it.maxClaims }));
  return parts.join(' · ');
}

function ruleSentenceCell(it) {
  var box = h('div', { class: 'text-sm text-secondary-700 dark:text-secondary-200 eco-clamp', 'data-rule-sentence': '' });
  if (!it.eventTypeId) return box;
  var ref = refName('event-types', it.eventTypeId);
  function paint(et) { clear(box).appendChild(document.createTextNode(ruleSentence(it, et ? i18nText(et.name) || et.key : '…'))); }
  if (ref.item !== undefined) paint(ref.item);
  else { box.appendChild(bar('70%')); ref.p.then(paint); }
  return box;
}

function grantTag(mode) {
  return tag(t('grant.' + mode), mode === 'auto' ? 'success' : 'primary');
}

function windowCell(it) {
  if (!it.startsAt && !it.endsAt) return h('span', { class: 'text-secondary-400', text: t('ev.window.always') });
  return h('div', { class: 'flex flex-col text-xs text-secondary-600 dark:text-secondary-300 whitespace-nowrap' },
    it.startsAt ? h('span', { text: t('ev.window.from', { date: fmtDate(it.startsAt) }) }) : null,
    it.endsAt ? h('span', { text: t('ev.window.to', { date: fmtDate(it.endsAt) }) }) : null);
}

function rowsEditor(value, f, onchange) {
  var rows = (value || []).map(function (r) { return Object.assign({}, r); });
  var listEl = h('div', { class: 'flex flex-col gap-2' });
  var errors = [];
  var narrow = f.columns.filter(function (c) { return !c.wide; }).length;
  function clean(r) {
    var o = {};
    f.columns.forEach(function (c) {
      var v = r[c.name];
      if (c.type === 'i18n') v = trClean(v);
      o[c.name] = v === undefined || v === '' ? null : v;
    });
    if (r.id) o.id = r.id;
    return o;
  }
  function emit() { onchange(rows.map(clean)); }
  var add = button({ variant: 'link', size: 'sm', icon: 'plus', label: t(f.addLabel || 'add.row'), onclick: function () {
    rows.push(f.newRow ? f.newRow(rows) : {});
    render();
    emit();
    var last = listEl.lastElementChild;
    var first = last && last.querySelector('input,select,button');
    if (first) first.focus();
  } });
  add.setAttribute('data-rows-add', f.name);
  function control(c, row) {
    var label = t(c.label);
    if (c.type === 'number') return numberInput(row[c.name], { int: c.int !== false, label: label, oninput: function (v) { row[c.name] = v === '' ? null : parseNumber(v, c.int !== false); emit(); } });
    if (c.type === 'ref') return refPicker({ resource: c.resource, api: c.api, filter: c.filter, valueKey: c.valueKey, value: row[c.name], required: c.required, label: label, onchange: function (v) { row[c.name] = v; emit(); } });
    if (c.type === 'text') return textInput(row[c.name] || '', { ltr: c.ltr, mono: c.mono, label: label, oninput: function (v) { row[c.name] = v.trim(); emit(); } });
    if (c.type === 'select') {
      if (row[c.name] == null) row[c.name] = c.options()[0].value;
      return selectInput(row[c.name], c.options(), function (v) { row[c.name] = v; emit(); }, { placeholder: false, label: label });
    }
    var ti = translatedInput(row[c.name], { label: label, max: c.max, oninput: function (v) { row[c.name] = v; emit(); } });
    row._tr = ti;
    return ti;
  }
  function render() {
    clear(listEl);
    errors = [];
    add.hidden = !!(f.max && rows.length >= f.max);
    if (!rows.length) listEl.appendChild(h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400 border border-dashed border-secondary-300 dark:border-secondary-700 rounded-md px-3 py-4 text-center', text: t(f.emptyText || 'rows.empty') }));
    rows.forEach(function (row, idx) {
      var cells = f.columns.map(function (c) {
        return h('div', { class: 'min-w-0' + (c.wide ? ' eco-rows-wide' : ''), 'data-col': c.name },
          h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t(c.label) + (c.optional ? ' (' + t('optional') + ')' : '') }),
          control(c, row));
      });
      var err = h('p', { class: 'text-xs text-error-600 dark:text-error-400 eco-items-err', hidden: true });
      errors.push({ row: row, el: err });
      listEl.appendChild(h('div', { class: 'border border-secondary-200 dark:border-secondary-700 rounded-md p-3', 'data-row': String(idx) },
        h('div', { class: 'eco-rows-row', style: '--eco-rows-cols:' + narrow },
          cells,
          h('div', { class: 'eco-items-remove' }, button({ variant: 'danger', size: 'iconLg', icon: 'trash', title: t('remove'), onclick: function () { rows.splice(idx, 1); render(); emit(); } }))),
        err));
    });
  }
  render();
  var el = h('div', { 'data-rows': f.name }, listEl, h('div', { class: 'mt-2' }, add));
  function show(i, msg) {
    var e = errors[i];
    if (!e) return;
    e.el.textContent = msg || '';
    e.el.hidden = !msg;
    e.el.parentElement.classList.toggle('eco-row-invalid', !!msg);
  }
  el.validate = function () {
    var bad = false;
    errors.forEach(function (e, i) {
      var msg = '';
      f.columns.forEach(function (c) {
        if (msg) return;
        var v = e.row[c.name];
        if (c.type === 'i18n') {
          var hit = trCheck(v, { max: c.max });
          if (hit) msg = t(c.label) + ': ' + t('code.' + hit.code, { lang: langName(hit.lang) });
          return;
        }
        if (v === null || v === undefined || v === '') { if (c.required) msg = t(c.label) + ': ' + t('err.required'); return; }
        if (c.type === 'number' && (isNaN(v) || (c.min !== undefined && v < c.min) || (c.max !== undefined && v > c.max))) msg = t(c.label) + ': ' + t('err.number');
      });
      show(i, msg);
      if (msg) bad = true;
    });
    if (!bad && f.check) {
      var hit2 = f.check(rows);
      if (hit2) { show(hit2.index, hit2.message); bad = true; }
    }
    if (!bad && f.min && rows.length < f.min) return t(f.emptyText || 'rows.empty');
    return bad ? t('err.form') : null;
  };
  el.showServerError = function (err) {
    var m = /\[(\d+)\]/.exec(String(err && err.field || ''));
    if (!m) return false;
    show(Number(m[1]), errorText(err));
    var e = errors[Number(m[1])];
    if (e) e.el.scrollIntoView({ block: 'center', behavior: 'smooth' });
    return !!e;
  };
  return el;
}

function duplicateCheck(name, code) {
  return function (rows) {
    var seen = {};
    for (var i = 0; i < rows.length; i++) {
      var v = rows[i][name];
      if (v === null || v === undefined) continue;
      if (seen[v]) return { index: i, message: t('code.' + code) };
      seen[v] = 1;
    }
    return null;
  };
}

function grantOptions() {
  return opts('grant.', ['claim', 'auto'])();
}

function activeFilter() {
  return { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } };
}

function playerTagsField() {
  return { name: 'audienceTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.ev.audience', hint: 'f.ev.audience.hint', optional: true, empty: [] };
}

function windowRow() {
  return { type: 'row', fields: [
    { name: 'startsAt', type: 'datetime', label: 'f.ev.startsAt', optional: true },
    { name: 'endsAt', type: 'datetime', label: 'f.ev.endsAt', optional: true, check: function (v, vals) { return v && vals.startsAt && new Date(v) <= new Date(vals.startsAt) ? t('code.INVALID_WINDOW') : null; } }
  ] };
}

function afterEventSave(resource) {
  return function () { forgetRefs(resource); };
}

RESOURCES['event-types'] = {
  filters: [activeFilter(), { name: 'clientAllowed', label: 'filter.ev.source', options: function () { return [{ value: 'true', label: t('ev.source.client') }, { value: 'false', label: t('ev.source.server') }]; } }],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '55%' },
    { label: 'col.ev.source', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, tag(t(it.clientAllowed ? 'ev.source.client' : 'ev.source.server'), it.clientAllowed ? 'warning' : 'secondary'), it.isSystem ? tag(t('ev.system'), 'primary') : null);
    }, width: 170 },
    { label: 'col.ev.range', render: function (it) {
      if (it.valueMin == null && it.valueMax == null) return h('span', { class: 'text-secondary-400', text: t('ev.range.any') });
      return h('span', { class: 'tabular-nums whitespace-nowrap', text: (it.valueMin == null ? '…' : fmtNum(it.valueMin)) + ' – ' + (it.valueMax == null ? '…' : fmtNum(it.valueMax)) });
    }, width: 120 },
    { label: 'col.ev.rate', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: t('ev.perMinute', { n: it.rateLimit }) }); }, width: 130 },
    { label: 'col.used', render: function (it) { return it.usedBy ? h('span', { class: 'tabular-nums', text: t('ev.usedBy', { n: it.usedBy }) }) : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 120 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { name: 'isSystem', type: 'static', readonly: true, when: function (v) { return !!v.isSystem; }, render: function () { return notice('primary', t('ev.system.notice')); } },
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.ev.key.hint', required: true, keyRe: EVENT_KEY_RE, keyErr: 'code.INVALID_EVENT_KEY', lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { name: 'clientAllowed', type: 'bool', inline: true, label: 'f.ev.client', hint: 'f.ev.client.hint', empty: false, when: function (v) { return !v.isSystem; } },
    { type: 'row', fields: [
      { name: 'valueMin', type: 'number', int: false, label: 'f.ev.valueMin', optional: true },
      { name: 'valueMax', type: 'number', int: false, label: 'f.ev.valueMax', optional: true, check: function (v, vals) { return v != null && vals.valueMin != null && Number(v) < Number(vals.valueMin) ? t('code.INVALID_RANGE') : null; } }
    ] },
    { name: 'rateLimit', type: 'number', label: 'f.ev.rate', hint: 'f.ev.rate.hint', required: true, min: 1, max: 100000 },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.ev.active.hint' }
  ],
  defaults: function () { return { key: '', name: {}, description: {}, clientAllowed: false, valueMin: null, valueMax: null, rateLimit: 30, isActive: true, sort: 0 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 0; return p; },
  afterSave: afterEventSave('event-types')
};

RESOURCES['event-rules'] = {
  filters: [
    { name: 'eventTypeId', label: 'filter.ev.event', type: 'ref', resource: 'event-types' },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '40%' },
    { label: 'col.ev.when', render: ruleSentenceCell, skeleton: '60%' },
    { label: 'col.ev.prize', render: function (it) { return h('div', { class: 'flex flex-col gap-1 min-w-0' }, refNameCell('prizes', it.prizeId), h('div', {}, grantTag(it.grantMode))); }, width: 170 },
    { label: 'col.ev.claims', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.claims || 0) }); }, width: 100 },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive), (it.audienceTagIds || []).length ? tag(t('audience.only'), 'primary') : null, it.endsAt && new Date(it.endsAt) < new Date() ? tag(t('ev.ended'), 'secondary') : null); }, width: 130 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { label: 'f.ev.rule.name', hint: 'f.ev.rule.name.hint', required: true, max: 128 }),
    tr('description', { label: 'f.ev.rule.description', multiline: true, max: 1000, optional: true }),
    { type: 'heading', label: 'ev.h.when' },
    { name: 'eventTypeId', type: 'ref', resource: 'event-types', label: 'f.ev.event', required: true },
    { type: 'row', fields: [
      { name: 'measure', type: 'select', label: 'f.ev.measure', hint: 'f.ev.measure.hint', required: true, options: opts('measure.', ['count', 'sum', 'max']) },
      { name: 'period', type: 'select', label: 'f.ev.period', required: true, options: opts('period.', ['total', 'day', 'week']) }
    ] },
    { type: 'row', fields: [
      { name: 'threshold', type: 'number', int: false, label: 'f.ev.threshold', required: true, min: 0.0001 },
      { name: 'repeat', type: 'select', label: 'f.ev.repeat', required: true, options: opts('repeat.', ['once', 'period', 'every']),
        check: function (v, vals) { return v === 'period' && vals.period === 'total' ? t('code.REPEAT_NEEDS_PERIOD') : v === 'every' && vals.measure === 'max' ? t('code.EVERY_NEEDS_TOTAL') : null; } }
    ] },
    { name: 'maxClaims', type: 'number', label: 'f.ev.maxClaims', hint: 'f.ev.maxClaims.hint', optional: true, min: 1, when: function (v) { return v.repeat !== 'once'; } },
    { name: 'preview', type: 'static', readonly: true, live: true, render: function (v) {
      if (!v.eventTypeId || !v.measure || !v.period || !v.repeat || v.threshold == null || v.threshold === '') return h('span', {});
      var item = { measure: v.measure, period: v.period, repeat: v.repeat, threshold: parseNumber(v.threshold, false), maxClaims: v.repeat === 'once' ? null : v.maxClaims };
      return h('div', { class: 'eco-rule-preview text-sm rounded-md px-3 py-2', 'data-rule-preview': '' }, ruleSentenceCell(Object.assign({ eventTypeId: v.eventTypeId }, item)));
    } },
    { type: 'heading', label: 'ev.h.prize' },
    { type: 'row', fields: [
      { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', required: true },
      { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', hint: 'f.ev.grantMode.hint', required: true, options: grantOptions }
    ] },
    { type: 'heading', label: 'ev.h.who' },
    playerTagsField(),
    windowRow(),
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
  ],
  defaults: function (query) { return { key: '', name: {}, description: {}, eventTypeId: (query && query.eventTypeId) || null, measure: 'count', period: 'day', threshold: 1, repeat: 'period', maxClaims: null, grantMode: 'claim', audienceTagIds: [], startsAt: null, endsAt: null, isActive: true, sort: 0 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 0; if (p.repeat === 'once') p.maxClaims = null; return p; }
};

RESOURCES['meta-rules'] = {
  filters: [
    { name: 'eventTypeId', label: 'filter.ev.event', type: 'ref', resource: 'event-types' },
    { name: 'metaId', label: 'filter.ev.meta', type: 'ref', resource: 'metas' },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '40%' },
    { label: 'col.ev.event', render: function (it) { return refNameCell('event-types', it.eventTypeId); }, width: 170 },
    { label: 'col.ev.change', render: function (it) {
      return h('div', { class: 'flex items-center gap-1.5 text-sm min-w-0', 'data-meta-change': '' }, refNameCell('metas', it.metaId),
        h('span', { class: 'tabular-nums whitespace-nowrap text-secondary-600 dark:text-secondary-300', text: t('mr.change.' + it.action + '.' + it.valueSource, { n: numLocal(it.amount) }) }));
    }, skeleton: '45%' },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive), (it.audienceTagIds || []).length ? tag(t('audience.only'), 'primary') : null); }, width: 130 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.mr.sort.hint' }
    ] },
    tr('name', { label: 'f.mr.name', hint: 'f.mr.name.hint', required: true, max: 128 }),
    { type: 'row', fields: [
      { name: 'eventTypeId', type: 'ref', resource: 'event-types', label: 'f.ev.event', required: true },
      { name: 'metaId', type: 'ref', resource: 'metas', label: 'f.mr.meta', required: true }
    ] },
    { type: 'row', fields: [
      { name: 'action', type: 'select', label: 'f.mr.action', required: true, options: opts('mr.action.', ['add', 'set', 'max']) },
      { name: 'valueSource', type: 'select', label: 'f.mr.source', hint: 'f.mr.source.hint', required: true, options: opts('mr.source.', ['fixed', 'event']) }
    ] },
    { name: 'amount', type: 'number', int: false, label: 'f.mr.amount', hint: 'f.mr.amount.hint', required: true },
    playerTagsField(),
    windowRow(),
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
  ],
  defaults: function () { return { key: '', name: {}, eventTypeId: null, metaId: null, action: 'add', valueSource: 'event', amount: 1, audienceTagIds: [], startsAt: null, endsAt: null, isActive: true, sort: 0 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 0; return p; }
};

RESOURCES['level-tables'] = {
  filters: [{ name: 'metaId', label: 'filter.ev.meta', type: 'ref', resource: 'metas' }, activeFilter()],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '40%' },
    { label: 'col.lv.meta', render: function (it) {
      return h('div', { class: 'flex items-center gap-1.5 text-sm min-w-0' }, refNameCell('metas', it.metaId), it.levelMetaId ? h('span', { class: 'text-secondary-400', text: '←' }) : null, it.levelMetaId ? refNameCell('metas', it.levelMetaId) : null);
    }, skeleton: '35%' },
    { label: 'col.lv.steps', render: function (it) {
      var steps = it.steps || [];
      var top = steps[steps.length - 1];
      return h('div', { class: 'flex flex-col text-sm' }, h('span', { class: 'tabular-nums', text: t('lv.levelsN', { n: steps.length }) }),
        top ? h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: t('lv.top', { level: top.level, n: numLocal(top.threshold) }) }) : null);
    }, width: 160 },
    { label: 'col.ev.players', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.players || 0) }); }, width: 100 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { type: 'row', fields: [
      { name: 'metaId', type: 'ref', resource: 'metas', label: 'f.lv.meta', hint: 'f.lv.meta.hint', required: true },
      { name: 'levelMetaId', type: 'ref', resource: 'metas', label: 'f.lv.levelMeta', hint: 'f.lv.levelMeta.hint', optional: true,
        check: function (v, vals) { return v && v === vals.metaId ? t('code.SAME_META') : null; } }
    ] },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' },
    { type: 'heading', label: 'lv.h.steps' },
    { name: 'steps', type: 'rows', addLabel: 'lv.add', emptyText: 'lv.empty', min: 1, max: 200,
      newRow: function (rows) {
        var last = rows[rows.length - 1];
        return { level: last ? Number(last.level || 0) + 1 : 1, threshold: last ? Number(last.threshold || 0) * 2 || 100 : 0, name: {}, prizeId: null, grantMode: 'claim' };
      },
      check: function (rows) {
        var dup = duplicateCheck('level', 'DUPLICATE_LEVEL')(rows);
        if (dup) return dup;
        var sorted = rows.map(function (r, i) { return { r: r, i: i }; }).sort(function (a, b) { return a.r.level - b.r.level; });
        for (var i = 1; i < sorted.length; i++) if (Number(sorted[i].r.threshold) < Number(sorted[i - 1].r.threshold)) return { index: sorted[i].i, message: t('code.THRESHOLDS_NOT_ORDERED') };
        return null;
      },
      columns: [
        { name: 'level', type: 'number', label: 'f.lv.level', required: true, min: 1 },
        { name: 'threshold', type: 'number', int: false, label: 'f.lv.threshold', required: true, min: 0 },
        { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', optional: true },
        { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', options: grantOptions },
        { name: 'name', type: 'i18n', label: 'f.lv.title', max: 64, wide: true }
      ] }
  ],
  defaults: function () { return { key: '', name: {}, description: {}, metaId: null, levelMetaId: null, isActive: true, sort: 0, steps: [{ level: 1, threshold: 0, name: {}, prizeId: null, grantMode: 'claim' }, { level: 2, threshold: 100, name: {}, prizeId: null, grantMode: 'claim' }] }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 0; return p; }
};

RESOURCES.streaks = {
  filters: [{ name: 'eventTypeId', label: 'filter.ev.event', type: 'ref', resource: 'event-types' }, activeFilter()],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '40%' },
    { label: 'col.ev.event', render: function (it) { return refNameCell('event-types', it.eventTypeId); }, width: 170 },
    { label: 'col.sk.rule', render: function (it) {
      return h('div', { class: 'flex flex-col text-sm gap-0.5' },
        h('span', { text: it.periodDays === 1 ? t('sk.everyDay') : t('sk.everyDays', { n: it.periodDays }) }),
        h('span', { class: 'text-xs text-secondary-500', text: (it.gracePeriods ? t('sk.graceN', { n: it.gracePeriods }) : t('sk.noGrace')) + ' · ' + t('sk.miss.' + it.missAction) }));
    }, skeleton: '40%' },
    { label: 'col.sk.rewards', render: function (it) {
      var n = (it.rewards || []).length;
      var ms = (it.milestones || []).map(function (m) { return m.day; });
      return h('div', { class: 'flex flex-col text-sm' }, h('span', { class: 'tabular-nums', text: t('sk.rewardsN', { n: n }) }),
        it.cycleLength ? h('span', { class: 'text-xs text-secondary-500', text: t('sk.cycleN', { n: it.cycleLength }) }) : null,
        ms.length ? h('span', { class: 'text-xs text-secondary-500 tabular-nums', 'data-sk-milestones': '', text: t('sk.milestonesAt', { days: ms.map(function (d) { return numLocal(d); }).join(lang() === 'fa' ? '، ' : ', ') }) }) : null);
    }, width: 170 },
    { label: 'col.ev.players', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.players || 0) }); }, width: 100 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { name: 'hiro', type: 'static', readonly: true, render: function () { return notice('secondary', t('sk.hiro.notice')); } },
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { name: 'eventTypeId', type: 'ref', resource: 'event-types', label: 'f.sk.event', hint: 'f.sk.event.hint', required: true },
    { type: 'row', fields: [
      { name: 'periodDays', type: 'number', label: 'f.sk.periodDays', hint: 'f.sk.periodDays.hint', required: true, min: 1, max: 30 },
      { name: 'cycleLength', type: 'number', label: 'f.sk.cycle', hint: 'f.sk.cycle.hint', optional: true, min: 1, max: 365 }
    ] },
    { type: 'row', fields: [
      { name: 'gracePeriods', type: 'number', label: 'f.sk.grace', hint: 'f.sk.grace.hint', required: true, min: 0, max: 30 },
      { name: 'missAction', type: 'select', label: 'f.sk.miss', required: true, options: opts('sk.miss.', ['reset', 'step_back']) }
    ] },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' },
    { type: 'heading', label: 'sk.h.rewards' },
    { name: 'rewards', type: 'rows', addLabel: 'sk.add', emptyText: 'sk.empty', max: 365,
      newRow: function (rows) {
        var last = rows[rows.length - 1];
        return { day: last ? Number(last.day || 0) + 1 : 1, prizeId: last ? last.prizeId : null, grantMode: last ? last.grantMode : 'claim' };
      },
      check: function (rows) { return duplicateCheck('day', 'DUPLICATE_DAY')(rows); },
      columns: [
        { name: 'day', type: 'number', label: 'f.sk.day', required: true, min: 1 },
        { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', required: true },
        { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', options: grantOptions }
      ] },
    { type: 'heading', label: 'sk.h.milestones' },
    { name: 'milestonesNote', type: 'static', readonly: true, render: function () { return h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', 'data-milestones-note': '', text: t('sk.milestones.note') }); } },
    { name: 'milestones', type: 'rows', addLabel: 'sk.ms.add', emptyText: 'sk.ms.empty', max: 50,
      newRow: function (rows) {
        var last = rows[rows.length - 1];
        return { day: last ? Number(last.day || 0) * 2 || 7 : 7, prizeId: null, grantMode: 'claim' };
      },
      check: function (rows) {
        var dup = duplicateCheck('day', 'DUPLICATE_MILESTONE')(rows);
        if (dup) return dup;
        for (var i = 0; i < rows.length; i++) if (Number(rows[i].day) > 3650) return { index: i, message: t('code.INVALID_MILESTONE_DAY') };
        return null;
      },
      columns: [
        { name: 'day', type: 'number', label: 'f.sk.msDay', required: true, min: 1 },
        { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.sk.msPrize', optional: true },
        { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', options: grantOptions }
      ] },
    { name: 'repayAfterFall', type: 'bool', inline: true, label: 'f.sk.repay', hint: 'f.sk.repay.hint' }
  ],
  defaults: function () { return { key: '', name: {}, description: {}, eventTypeId: null, periodDays: 1, cycleLength: 7, gracePeriods: 0, missAction: 'reset', isActive: true, sort: 0, rewards: [], milestones: [], repayAfterFall: false }; },
  fixPayload: function (p) {
    if (p.sort == null) p.sort = 0;
    return p;
  }
};

LISTS['event-log'] = {
  noRow: true,
  filters: [
    { name: 'typeId', label: 'filter.ev.event', type: 'ref', resource: 'event-types' },
    { name: 'source', label: 'filter.ev.from', options: opts('ev.from.', ['client', 'server', 'admin']) },
    { name: 'userId', label: 'filter.user', type: 'text' }
  ],
  columns: [
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 170 },
    { label: 'col.player', render: function (it) {
      return h('div', { class: 'min-w-0' }, h('a', { href: '#/players/' + encodeURIComponent(it.userId), class: 'eco-link font-medium truncate block', text: it.displayName || it.username || String(it.userId).slice(0, 8) }),
        it.username && it.displayName ? h('div', { class: 'text-xs text-secondary-500 truncate', dir: 'ltr', text: it.username }) : null);
    }, skeleton: '40%' },
    { label: 'col.ev.event', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'truncate', text: i18nText(it.name) || it.key }), h('div', { class: 'text-secondary-500' }, mono(it.key))); }, skeleton: '40%' },
    { label: 'col.ev.value', render: function (it) { return numCell(it.value); }, width: 100 },
    { label: 'col.ev.from', render: function (it) { return tag(t('ev.from.' + it.source), it.source === 'client' ? 'warning' : it.source === 'admin' ? 'primary' : 'secondary'); }, width: 110 },
    { label: 'col.ev.metadata', render: function (it) {
      var keys = it.metadata ? Object.keys(it.metadata) : [];
      if (!keys.length) return h('span', { class: 'text-secondary-400', text: t('none') });
      return h('span', { class: 'font-mono text-xs text-secondary-600 dark:text-secondary-300 eco-clamp', dir: 'ltr', text: JSON.stringify(it.metadata) });
    }, skeleton: '30%' }
  ]
};

OVERVIEW_CATALOG.push('event-types', 'event-rules', 'level-tables', 'streaks');

var TIMEZONES = ['Asia/Tehran', 'UTC', 'Asia/Dubai', 'Asia/Baghdad', 'Asia/Kabul', 'Asia/Baku', 'Asia/Yerevan', 'Europe/Istanbul', 'Europe/London', 'Europe/Berlin', 'Europe/Moscow', 'America/New_York', 'America/Los_Angeles', 'America/Toronto', 'Australia/Sydney', 'Asia/Tokyo', 'Asia/Kolkata', 'Asia/Riyadh'];

function weekdayName(i) {
  var d = new Date(Date.UTC(2026, 10, 8 + i));
  return new Intl.DateTimeFormat(locale(), { weekday: 'long', timeZone: 'UTC' }).format(d);
}

function zoneNow(zone) {
  try {
    return new Intl.DateTimeFormat(locale(), { weekday: 'long', hour: '2-digit', minute: '2-digit', timeZone: zone }).format(new Date());
  } catch (e) {
    return '';
  }
}

var EVENT_SETTINGS_FIELDS = [
  { type: 'row', fields: [
    { name: 'timezone', type: 'select', label: 'f.es.timezone', hint: 'f.es.timezone.hint', required: true, options: function () {
      return TIMEZONES.map(function (z) { return { value: z, label: z === 'UTC' ? 'UTC' : z.replace('_', ' ') }; });
    } },
    { name: 'weekStart', type: 'select', label: 'f.es.weekStart', hint: 'f.es.weekStart.hint', required: true, noPlaceholder: true, options: function () {
      return [6, 0, 1, 2, 3, 4, 5].map(function (i) { return { value: String(i), label: weekdayName(i) }; });
    } }
  ] },
  { name: 'clock', type: 'static', readonly: true, live: true, render: function (v) {
    var now = zoneNow(v.timezone || 'UTC');
    return now ? h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 -mt-2 mb-4', 'data-zone-now': '' }, t('es.now', { time: now })) : h('span', {});
  } },
  { type: 'row', fields: [
    { name: 'retentionDays', type: 'number', label: 'f.es.retention', hint: 'f.es.retention.hint', required: true, min: 1, max: 3650 },
    { name: 'playerRateLimit', type: 'number', label: 'f.es.rate', hint: 'f.es.rate.hint', required: true, min: 1, max: 100000 }
  ] }
];

function eventSettingsPage(container) {
  var frame = tabFrame(t('desc.events'));
  clear(container).appendChild(frame.el);
  var page = apiPage(api, '/event-settings', null, EVENT_SETTINGS_FIELDS, { id: 'events', fixPayload: function (p) {
    p.weekStart = Number(p.weekStart);
    return p;
  } });
  frame.content.appendChild(page.el);
  page.load();
}

var PROGRESS_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadProgressCard(id) {
  PROGRESS_CARD.loading = true;
  return api('GET', '/users/' + encodeURIComponent(id) + '/progress' + qs({ lang: lang() })).then(function (res) {
    PROGRESS_CARD.loading = false;
    PROGRESS_CARD.at = Date.now();
    PROGRESS_CARD.data = res;
    PROGRESS_CARD.err = null;
    var old = document.querySelector('[data-player-progress]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    PROGRESS_CARD.loading = false;
    PROGRESS_CARD.at = Date.now();
    PROGRESS_CARD.err = err;
    schedule();
  });
}

function claimStatusTag(c) {
  var color = c.status === 'granted' ? 'success' : c.status === 'failed' ? 'error' : 'warning';
  return tag(t('claim.status.' + (c.status === 'ready' ? (c.grantMode === 'claim' ? 'waiting' : 'ready') : c.status)), color);
}

function recordEventDialog(userId) {
  var values = { key: '', value: 1 };
  var picker = refPicker({ resource: 'event-types', valueKey: 'key', value: '', required: true, label: t('f.ev.event'), filter: { isActive: 'true' }, onchange: function (v) { values.key = v || ''; } });
  var val = numberInput(1, { int: false, label: t('f.ev.value'), oninput: function (v) { values.value = v === '' ? null : parseNumber(v, false); } });
  var keyWrap = field(t('f.ev.event'), picker, null);
  var valWrap = field(t('f.ev.value'), val, t('f.ev.value.hint'));
  return formDialog({
    id: 'record-event',
    title: t('pp.record'),
    text: t('pp.record.text'),
    icon: 'event-types',
    ok: t('pp.record.ok'),
    body: h('div', {}, keyWrap, valWrap),
    validate: function () {
      keyWrap.setError(values.key ? '' : t('err.required'));
      valWrap.setError(values.value === null || isNaN(values.value) ? t('err.number') : '');
      return !!values.key && values.value !== null && !isNaN(values.value);
    },
    submit: function () { return api('POST', '/users/' + encodeURIComponent(userId) + '/events', { key: values.key, value: values.value }); }
  });
}

function progressSection(titleKey, body, extra) {
  return h('div', { class: 'min-w-0' }, h('div', { class: 'flex items-center justify-between gap-2 text-xs text-secondary-500 dark:text-secondary-400 mb-1' }, h('span', { text: t(titleKey) }), extra || null), body);
}

function playerProgressCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-progress]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (PROGRESS_CARD.id !== id) { PROGRESS_CARD.id = id; PROGRESS_CARD.data = null; PROGRESS_CARD.err = null; PROGRESS_CARD.at = 0; }
  if (!PROGRESS_CARD.loading && Date.now() - PROGRESS_CARD.at > 20000) loadProgressCard(id);
  var p = PROGRESS_CARD.data;
  var anchor = grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify([p.totals, p.streaks, p.levels, p.claims, p.readyClaims]) : PROGRESS_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-progress': '', 'data-key': key, 'data-economy': 'player-progress', 'data-l10n-skip': '', 'aria-labelledby': 'player-progress-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-progress-title', class: 'font-medium', text: t('pp.title') });
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(PROGRESS_CARD.err ? notice('error', errorText(PROGRESS_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { PROGRESS_CARD.at = 0; loadProgressCard(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var write = aclCan('HIRO_ECONOMY', 1);
  var reload = function (res) { if (res) { PROGRESS_CARD.at = 0; loadProgressCard(id); } };
  var actions = h('div', { class: 'flex flex-wrap items-center gap-2' });
  if (write) {
    var rec = button({ variant: 'ghost', size: 'sm', icon: 'event-types', label: t('pp.record'), onclick: function () { recordEventDialog(id).then(function (res) { if (res) toast('success', t('pp.recorded')); reload(res); }); } });
    rec.setAttribute('data-record-event', '');
    actions.appendChild(rec);
  }
  actions.appendChild(h('a', { href: hashFor('event-log', null, { userId: id }), class: 'eco-link text-sm', text: t('pp.log') }));
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-3' }, title, actions));

  var levels = h('div', { class: 'flex flex-col gap-2' });
  (p.levels || []).forEach(function (lv) {
    var next = lv.next;
    levels.appendChild(h('div', { class: 'min-w-0', 'data-pp-level': lv.key },
      h('div', { class: 'font-medium truncate', text: lv.name }),
      h('div', { class: 'text-sm', text: t('pp.level', { level: lv.level }) + (lv.title ? ' · ' + lv.title : '') }),
      h('div', { class: 'text-xs text-secondary-500 tabular-nums', text: t('pp.meta', { n: numLocal(lv.value), name: lv.metaName }) + (next ? ' · ' + t('pp.toNext', { level: next.level, n: numLocal(Math.max(0, next.threshold - lv.value)) }) : '') })));
  });
  if (!(p.levels || []).length) levels.appendChild(h('span', { class: 'text-sm text-secondary-400', text: t('none') }));

  var streaks = h('div', { class: 'flex flex-col gap-2' });
  (p.streaks || []).forEach(function (sk) {
    var row = h('div', { class: 'min-w-0', 'data-pp-streak': sk.key },
      h('div', { class: 'flex items-center gap-2 min-w-0' }, h('span', { class: 'font-medium truncate', text: sk.name }), sk.broken ? tag(t('pp.broken'), 'secondary') : sk.countedThisPeriod ? tag(t('pp.today'), 'success') : null),
      h('div', { class: 'text-sm tabular-nums', text: t('pp.streak', { n: sk.count, best: sk.best }) }),
      sk.floor || sk.nextMilestone ? h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 tabular-nums', 'data-pp-floor': String(sk.floor || 0),
        text: [sk.floor ? t('pp.floor', { n: sk.floor }) : t('pp.noFloor'), sk.nextMilestone ? t('pp.nextMilestone', { n: sk.nextMilestone }) : null].filter(Boolean).join(' · ') }) : null);
    if (write && (sk.count > 0 || sk.countedThisPeriod)) {
      var reset = button({ variant: 'link', size: 'sm', label: t('pp.reset'), onclick: function () {
        confirmModal({ title: t('pp.reset.title'), text: t('pp.reset.text', { name: sk.name }), ok: t('pp.reset'), danger: true }).then(function (yes) {
          if (!yes) return;
          api('POST', '/users/' + encodeURIComponent(id) + '/streaks/' + encodeURIComponent(sk.id) + '/reset', {}).then(function () { toast('success', t('pp.reset.done')); reload(true); }, function (err) { toast('error', errorText(err)); });
        });
      } });
      reset.setAttribute('data-streak-reset', sk.key);
      row.appendChild(reset);
    }
    streaks.appendChild(row);
  });
  if (!(p.streaks || []).length) streaks.appendChild(h('span', { class: 'text-sm text-secondary-400', text: t('none') }));

  var totals = h('div', { class: 'flex flex-col gap-1' });
  (p.totals || []).slice(0, 6).forEach(function (e) {
    totals.appendChild(h('div', { class: 'flex items-center justify-between gap-2 text-sm', 'data-pp-total': e.key },
      h('span', { class: 'truncate', text: i18nText(e.name) || e.key }),
      h('span', { class: 'tabular-nums text-secondary-600 dark:text-secondary-300 whitespace-nowrap', text: t('pp.times', { n: e.count }) })));
  });
  if (!(p.totals || []).length) totals.appendChild(h('span', { class: 'text-sm text-secondary-400', text: t('pp.noEvents') }));

  body.appendChild(h('div', { class: 'eco-pc-grid' }, progressSection('pp.levels', levels), progressSection('pp.streaks', streaks), progressSection('pp.events', totals)));

  var claims = h('div', { class: 'flex flex-col mt-3 border-t border-secondary-200 dark:border-secondary-700 pt-4', 'data-pp-claims': '' });
  claims.appendChild(h('div', { class: 'flex items-center justify-between gap-2 text-xs text-secondary-500 dark:text-secondary-400 mb-1' }, h('span', { text: t('pp.claims') }), p.readyClaims ? tag(t('pp.readyN', { n: p.readyClaims }), 'warning') : null));
  (p.claims || []).slice(0, 8).forEach(function (c) {
    var line = h('div', { class: 'flex flex-wrap items-center gap-2 py-1 text-sm', 'data-pp-claim': c.claimId },
      tag(t('claim.kind.' + c.kind), 'secondary'),
      h('span', { class: 'truncate', text: c.sourceName || c.source }),
      c.level ? h('span', { class: 'text-xs text-secondary-500', text: t('pp.level', { level: c.level }) }) : null,
      h('span', { class: 'text-secondary-400', text: '←' }),
      h('span', { class: 'font-medium truncate', text: c.prize && c.prize.name }),
      claimStatusTag(c),
      h('span', { class: 'flex-1' }),
      h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: fmtDate(c.grantedAt || c.createdAt) }));
    if (write && c.status === 'failed') {
      line.appendChild(button({ variant: 'link', size: 'sm', icon: 'refresh', label: t('pay.retry'), onclick: function () {
        api('POST', '/event-claims/' + encodeURIComponent(c.claimId) + '/retry', {}).then(function () { toast('success', t('pay.retry.done')); reload(true); }, function (err) { toast('error', errorText(err)); });
      } }));
    }
    claims.appendChild(line);
  });
  if (!(p.claims || []).length) claims.appendChild(h('span', { class: 'text-sm text-secondary-400', text: t('pp.noClaims') }));
  body.appendChild(claims);
  place();
}
