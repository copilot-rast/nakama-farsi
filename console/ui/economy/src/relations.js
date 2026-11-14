ICONS['meta-relations'] = '<circle cx="6" cy="18" r="3"/><circle cx="18" cy="6" r="3"/><path d="M6 15V9a3 3 0 0 1 3-3h6"/><path d="m13 3 3 3-3 3"/>';
PH_ICONS['meta-relations'] = '<path d="M245.66,74.34l-32-32a8,8,0,0,0-11.32,11.32L220.69,72H208c-49.33,0-61.05,28.12-71.38,52.92-9.38,22.51-16.92,40.59-49.48,42.84a40,40,0,1,0,.1,16c43.26-2.65,54.34-29.15,64.14-52.69C161.41,107,169.33,88,208,88h12.69l-18.35,18.34a8,8,0,0,0,11.32,11.32l32-32A8,8,0,0,0,245.66,74.34ZM48,200a24,24,0,1,1,24-24A24,24,0,0,1,48,200Z"/>';
PH_ICONS.flowDuo = '<circle cx="48" cy="176" r="24" opacity="0.2"/>' + PH_ICONS['meta-relations'];
PH_ICONS.metasDuo = '<path d="M224,208H32V48l64,72,32-24,56,56,40-40Z" opacity="0.2"/>' + PH_ICONS.metas;
PH_ICONS.rulesDuo = '<circle cx="128" cy="88" r="24" opacity="0.2"/><circle cx="56" cy="136" r="24" opacity="0.2"/><circle cx="200" cy="168" r="24" opacity="0.2"/>' + PH_ICONS['meta-rules'];
PH_ICONS.levelsDuo = '<path d="M80.5,128h47l-11.64,32h-47Z" opacity="0.2"/>' + PH_ICONS['level-tables'];
PH_ICONS.prizesDuo = '<path d="M208,128v72a8,8,0,0,1-8,8H56a8,8,0,0,1-8-8V128Z" opacity="0.2"/>' + PH_ICONS.prizes;
PH_ICONS.arrowDownDuo = '<path d="M200,136l-72,72L56,136Z" opacity="0.2"/><path d="M205.66,130.34a8,8,0,0,0-11.32,0L136,188.69V40a8,8,0,0,0-16,0V188.69L61.66,130.34a8,8,0,0,0-11.32,11.32l72,72a8,8,0,0,0,11.32,0l72-72A8,8,0,0,0,205.66,130.34Z"/>';

var REL_OPS = ['gte', 'gt', 'eq', 'lte', 'lt'];
var REL_SYMBOL = { gt: '>', gte: '≥', lt: '<', lte: '≤', eq: '=' };

function relSigned(n, action) {
  if (action === 'set') return '= ' + numLocal(n);
  return (Number(n) < 0 ? '− ' + numLocal(Math.abs(Number(n))) : '+ ' + numLocal(n));
}

function relWhenCell(it) {
  return h('div', { class: 'flex items-center gap-1.5 text-sm min-w-0', 'data-rel-when': '' },
    it.metaId ? refNameCell('metas', it.metaId) : h('span', { class: 'text-secondary-400', text: '…' }),
    h('span', { class: 'eco-rel-op tabular-nums whitespace-nowrap', dir: 'ltr', text: (REL_SYMBOL[it.operator] || '?') + ' ' + numLocal(it.value) }));
}

function relThenCell(it) {
  var box = h('div', { class: 'flex flex-wrap items-center gap-1.5 text-sm min-w-0', 'data-rel-then': '' });
  (it.targets || []).forEach(function (tg) {
    if (!tg || !tg.metaId) return;
    box.appendChild(h('span', { class: 'eco-rel-chip inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 min-w-0', 'data-rel-target': tg.action || 'add' },
      refNameCell('metas', tg.metaId), h('span', { class: 'tabular-nums whitespace-nowrap', dir: 'ltr', text: relSigned(tg.amount, tg.action) })));
  });
  if (it.prizeType === 'prize' && it.prizeId) box.appendChild(h('span', { class: 'eco-rel-chip eco-rel-prize inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 min-w-0', 'data-rel-prize': '' }, icon('prizes', 'h-3.5 w-3.5'), refNameCell('prizes', it.prizeId)));
  if (it.prizeType === 'prizeGroup' && it.prizeGroupId) box.appendChild(h('span', { class: 'eco-rel-chip eco-rel-prize inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 min-w-0', 'data-rel-prize': 'group' }, icon('prize-groups', 'h-3.5 w-3.5'), refNameCell('prize-groups', it.prizeGroupId)));
  if (it.prizeType && it.prizeType !== 'none') box.appendChild(grantTag(it.grantMode || 'auto'));
  if (!box.firstChild) box.appendChild(h('span', { class: 'text-secondary-400', text: t('none') }));
  return box;
}

function relRepeatTag(it) {
  if (it.repeat === 'once' || !it.repeat) return tag(t('rel.repeat.once'), 'secondary');
  return tag(t('rel.repeat.' + it.repeat) + (it.maxFires ? ' · ' + t('rel.maxN', { n: it.maxFires }) : ''), 'primary');
}

RESOURCES['meta-relations'] = {
  filters: [
    { name: 'metaId', label: 'filter.rel.source', type: 'ref', resource: 'metas' },
    { name: 'targetMetaId', label: 'filter.rel.target', type: 'ref', resource: 'metas' },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '35%' },
    { label: 'col.rel.when', render: relWhenCell, skeleton: '25%' },
    { label: 'col.rel.then', render: relThenCell, skeleton: '35%' },
    { label: 'col.ev.players', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.players || 0) }); }, width: 90 },
    { label: 'col.status', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive), relRepeatTag(it), (it.audienceTagIds || []).length ? tag(t('audience.only'), 'primary') : null,
        it.clientVisible === false ? tag(t('rel.hidden'), 'secondary') : null, it.endsAt && new Date(it.endsAt) < new Date() ? tag(t('ev.ended'), 'secondary') : null);
    }, width: 150 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.rel.sort.hint' }
    ] },
    tr('name', { label: 'f.rel.name', hint: 'f.rel.name.hint', required: true, max: 128 }),
    tr('description', { label: 'f.rel.description', hint: 'f.rel.description.hint', multiline: true, max: 1000, optional: true }),
    { type: 'heading', label: 'rel.h.when' },
    { name: 'metaId', type: 'ref', resource: 'metas', label: 'f.rel.meta', hint: 'f.rel.meta.hint', required: true },
    { type: 'row', fields: [
      { name: 'operator', type: 'select', label: 'f.rel.operator', required: true, noPlaceholder: true, options: opts('rel.op.', REL_OPS) },
      { name: 'value', type: 'number', int: false, label: 'f.rel.value', required: true }
    ] },
    { type: 'heading', label: 'rel.h.then' },
    { name: 'targets', type: 'rows', addLabel: 'rel.target.add', emptyText: 'rel.target.empty', max: 20, empty: [], refresh: true,
      newRow: function () { return { metaId: null, action: 'add', amount: 1 }; },
      check: function (rows) {
        var seen = {};
        for (var i = 0; i < rows.length; i++) {
          var k2 = rows[i].metaId + '|' + rows[i].action;
          if (rows[i].metaId && seen[k2]) return { index: i, message: t('code.DUPLICATE_TARGET') };
          seen[k2] = 1;
          if (rows[i].action !== 'set' && Number(rows[i].amount) === 0) return { index: i, message: t('code.ZERO_AMOUNT') };
        }
        return null;
      },
      columns: [
        { name: 'metaId', type: 'ref', resource: 'metas', label: 'f.rel.target', required: true },
        { name: 'action', type: 'select', label: 'f.rel.action', options: opts('rel.action.', ['add', 'set']) },
        { name: 'amount', type: 'number', int: false, label: 'f.rel.amount', required: true }
      ] },
    { type: 'heading', label: 'rel.h.prize' },
    { type: 'row', fields: [
      { name: 'prizeType', type: 'select', label: 'f.rel.prizeType', required: true, noPlaceholder: true, options: opts('rel.prizeType.', ['none', 'prize', 'prizeGroup']) },
      { name: 'grantMode', type: 'select', label: 'f.ev.grantMode', hint: 'f.rel.grantMode.hint', required: true, noPlaceholder: true, options: grantOptions, when: function (v) { return v.prizeType && v.prizeType !== 'none'; } }
    ] },
    { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', required: true, when: function (v) { return v.prizeType === 'prize'; } },
    { name: 'prizeGroupId', type: 'ref', resource: 'prize-groups', label: 'f.rel.prizeGroup', hint: 'f.rel.prizeGroup.hint', required: true, when: function (v) { return v.prizeType === 'prizeGroup'; } },
    { name: 'needsResult', type: 'static', readonly: true, live: true, render: function (v) {
      var empty = !(v.targets || []).length && (!v.prizeType || v.prizeType === 'none');
      return empty ? h('p', { class: 'text-xs text-error-600 dark:text-error-400', 'data-rel-needs': '', text: t('code.RELATION_NEEDS_RESULT') }) : h('span', {});
    } },
    { type: 'heading', label: 'rel.h.repeat' },
    { type: 'row', fields: [
      { name: 'repeat', type: 'select', label: 'f.rel.repeat', hint: 'f.rel.repeat.hint', required: true, noPlaceholder: true, options: opts('rel.repeat.', ['once', 'rearm', 'every']) },
      { name: 'maxFires', type: 'number', label: 'f.rel.maxFires', hint: 'f.rel.maxFires.hint', optional: true, min: 1, max: 1000000, when: function (v) { return v.repeat && v.repeat !== 'once'; } }
    ] },
    { name: 'preview', type: 'static', readonly: true, live: true, render: function (v) {
      if (!v.metaId || v.value == null || v.value === '') return h('span', {});
      var item = { metaId: v.metaId, operator: v.operator, value: parseNumber(v.value, false), targets: v.targets || [], prizeType: v.prizeType, prizeId: v.prizeId, prizeGroupId: v.prizeGroupId, grantMode: v.grantMode };
      return h('div', { class: 'eco-rule-preview text-sm rounded-md px-3 py-2 flex flex-wrap items-center gap-2', 'data-rel-preview': '' },
        h('span', { class: 'font-medium', text: t('rel.preview.when') }), relWhenCell(item), icon('chevron', 'h-4 w-4 eco-caret'), relThenCell(item),
        h('span', { class: 'text-xs', text: t('rel.repeat.' + (v.repeat || 'once') + '.say') }));
    } },
    { type: 'heading', label: 'ev.h.who' },
    playerTagsField(),
    windowRow(),
    { name: 'clientVisible', type: 'bool', inline: true, label: 'f.rel.visible', hint: 'f.rel.visible.hint' },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
  ],
  defaults: function (query) {
    return { key: '', name: {}, description: {}, metaId: (query && query.metaId) || null, operator: 'gte', value: 100, targets: [{ metaId: null, action: 'add', amount: 1 }],
      prizeType: 'none', prizeId: null, prizeGroupId: null, grantMode: 'auto', repeat: 'once', maxFires: null, audienceTagIds: [], clientVisible: true, startsAt: null, endsAt: null, isActive: true, sort: 0 };
  },
  toForm: function (v) {
    v.targets = (v.targets || []).map(function (tg) { return { metaId: tg.metaId, action: tg.action || 'add', amount: tg.amount }; });
    if (!v.prizeType) v.prizeType = 'none';
    return v;
  },
  fixPayload: function (p) {
    if (p.sort == null) p.sort = 0;
    if (p.prizeType !== 'prize') p.prizeId = '';
    if (p.prizeType !== 'prizeGroup') p.prizeGroupId = '';
    if (!p.grantMode) p.grantMode = 'auto';
    if (p.repeat === 'once') p.maxFires = null;
    p.targets = (p.targets || []).map(function (tg) { return { metaId: tg.metaId, action: tg.action || 'add', amount: tg.amount }; });
    return p;
  },
  afterSave: function () { forgetRefs('meta-relations'); }
};

RESOURCES.metas.onRow = function (item) { openMetaStudio(item.id); };
RESOURCES.metas.mounted = function (query) { if (query && query.studio) setTimeout(function () { openMetaStudio(query.studio); }, 0); };

var STUDIO = null;

function studioHash(id) {
  var q = Object.assign({}, lastQuery.metas || {});
  delete q.studio;
  if (id) q.studio = id;
  return hashFor('metas', null, q);
}

function closeMetaStudio() {
  if (!STUDIO) return;
  STUDIO = null;
  if (panelState && panelState.key.indexOf('studio/') === 0) closePanel();
  if (/^#\/economy\/metas(\?|$)/.test(location.hash)) replaceHash(studioHash(null));
}

function openMetaStudio(id) {
  if (STUDIO && STUDIO.id === id && STUDIO.el.isConnected) return;
  closePanel();
  var body = h('div', { 'data-meta-studio': id, class: 'eco-studio' });
  var footer = h('div', { class: 'flex gap-4 w-full items-center' });
  var el = sidePanel(t('studio.title'), body, footer, function () { if (STUDIO && STUDIO.editing) studioOverview(); else closeMetaStudio(); });
  el.setAttribute('data-economy-studio', '');
  var esc = function (e) {
    if (e.key !== 'Escape' || !STUDIO || document.querySelector('[data-economy="panel"] .eco-pop:not([hidden]), [data-economy="dialog"]')) return;
    e.preventDefault();
    if (STUDIO.editing) studioOverview(); else closeMetaStudio();
  };
  document.addEventListener('keydown', esc, true);
  panelState = { el: el, key: 'studio/' + id, esc: esc };
  STUDIO = { id: id, el: el, body: body, footer: footer, meta: null, editing: null };
  document.body.appendChild(el);
  if (/^#\/economy\/metas(\?|$)/.test(location.hash)) replaceHash(studioHash(id));
  studioOverview();
}

function studioTitle(text) {
  var head = STUDIO && STUDIO.el.querySelector('[role="dialog"] > div.text-lg > div');
  if (head) head.textContent = text;
  var dlg = STUDIO && STUDIO.el.querySelector('[role="dialog"]');
  if (dlg) dlg.setAttribute('aria-label', text);
}

function studioList(path) {
  return api('GET', path).then(function (res) { return (res && res.items) || []; });
}

function studioOverview() {
  var st = STUDIO;
  if (!st) return;
  st.editing = null;
  clear(st.footer).appendChild(h('div', { class: 'flex-1' }));
  st.footer.appendChild(button({ variant: 'ghost', size: 'footer', label: t('close'), onclick: closeMetaStudio }));
  clear(st.body).appendChild(h('div', { class: 'flex flex-col gap-3' }, bar('40%', 'h-5'), bar('100%', 'h-20'), bar('100%', 'h-20'), bar('100%', 'h-20')));
  var id = encodeURIComponent(st.id);
  api('GET', '/metas/' + id).then(function (meta) {
    if (STUDIO !== st) return null;
    st.meta = meta;
    studioTitle(t('studio.titleOf', { name: i18nText(meta.name) || meta.key }));
    return Promise.all([
      studioList('/meta-rules?limit=100&metaId=' + id),
      studioList('/meta-relations?limit=100&targetMetaId=' + id),
      studioList('/prizes?limit=100&metaKey=' + encodeURIComponent(meta.key)),
      studioList('/level-tables?limit=100&levelMetaId=' + id),
      studioList('/meta-relations?limit=100&metaId=' + id),
      studioList('/level-tables?limit=100&metaId=' + id)
    ]).then(function (all) {
      if (STUDIO !== st || st.editing) return;
      paintStudio(st, meta, { rules: all[0], relIn: all[1].filter(function (r) { return r.metaId !== st.id; }), prizes: all[2], levelsIn: all[3], relOut: all[4], levels: all[5] });
    });
  }).catch(function (err) {
    if (STUDIO !== st) return;
    clear(st.body).appendChild(errorBlock(err, studioOverview));
  });
}

function studioGroup(kind, iconName, titleKey, items, rowFn, addFn) {
  var head = h('div', { class: 'eco-studio-ghead flex items-center gap-2' },
    h('span', { class: 'eco-studio-gicon shrink-0 flex items-center justify-center rounded-md' }, icon(iconName, 'h-5 w-5')),
    h('span', { class: 'font-medium flex-1 min-w-0', text: t(titleKey) }),
    h('span', { class: 'eco-studio-count tabular-nums text-xs rounded-full px-2 py-0.5', text: fmtNum(items.length) }),
    addFn ? (function () {
      var b = button({ variant: 'link', size: 'sm', icon: 'plus', label: t('studio.add'), onclick: addFn });
      b.setAttribute('data-studio-add', kind);
      return b;
    })() : null);
  var list = h('div', { class: 'flex flex-col' });
  items.forEach(function (it) { list.appendChild(rowFn(it)); });
  if (!items.length) list.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 px-3 py-2', text: t('studio.empty.' + kind) }));
  return h('section', { class: 'eco-studio-group rounded-md', 'data-studio-group': kind }, head, list);
}

function studioRow(kind, it, main, sub, onclick, href) {
  var attrs = { class: 'eco-studio-row flex items-start gap-2 w-full eco-text-start px-3 py-2 min-w-0', 'data-studio-item': it.id, 'data-studio-kind': kind };
  var inner = [h('div', { class: 'flex-1 min-w-0 flex flex-col gap-1' },
    h('div', { class: 'flex items-center gap-2 min-w-0' }, h('span', { class: 'font-medium truncate', text: i18nText(it.name) || it.key }), it.isActive === false ? tag(t('inactive'), 'secondary') : null, main || null),
    sub || null), icon(href ? 'external' : 'chevron', 'h-4 w-4 shrink-0 text-secondary-400 eco-studio-go' + (href ? '' : ' eco-caret'))];
  if (href) { attrs.href = href; return h('a', attrs, inner); }
  attrs.type = 'button';
  attrs.onclick = onclick;
  return h('button', attrs, inner);
}

function paintStudio(st, meta, d) {
  var write = aclCan('HIRO_ECONOMY', 1);
  var node = h('section', { class: 'eco-studio-node rounded-md flex items-center gap-3 px-4 py-3', 'data-studio-node': meta.key },
    h('span', { class: 'eco-studio-nicon shrink-0 flex items-center justify-center rounded-md' }, icon('metasDuo', 'h-6 w-6')),
    h('div', { class: 'flex-1 min-w-0' },
      h('div', { class: 'font-semibold truncate', text: i18nText(meta.name) || meta.key }),
      h('div', { class: 'flex flex-wrap items-center gap-2 text-xs text-secondary-500 dark:text-secondary-400' }, mono(meta.key),
        h('span', { text: t('studio.default', { n: numLocal(meta.defaultValue) }) }), meta.serverOnly ? tag(t('server.only'), 'warning') : null)),
    write ? (function () { var b = button({ variant: 'ghost', size: 'sm', icon: 'gear', label: t('studio.editMeta'), onclick: function () { closeMetaStudio(); go('metas', meta.id, lastQuery.metas || {}); } }); b.setAttribute('data-studio-edit-meta', ''); return b; })() : null);

  var add = function (name, preset, lock) { return write ? function () { studioEditor(name, null, preset, lock); } : null; };
  var edit = function (name, it, lock) { return function () { studioEditor(name, it.id, null, lock); }; };

  var inputs = [
    studioGroup('meta-rules', 'rulesDuo', 'studio.g.rules', d.rules, function (it) {
      return studioRow('meta-rules', it, null, h('div', { class: 'flex items-center gap-1.5 text-xs text-secondary-600 dark:text-secondary-300 min-w-0' }, refNameCell('event-types', it.eventTypeId),
        h('span', { class: 'tabular-nums whitespace-nowrap', text: t('mr.change.' + it.action + '.' + it.valueSource, { n: numLocal(it.amount) }) })), edit('meta-rules', it, 'metaId'));
    }, add('meta-rules', { metaId: st.id }, 'metaId')),
    studioGroup('relations-in', 'flowDuo', 'studio.g.relIn', d.relIn, function (it) {
      return studioRow('meta-relations', it, relRepeatTag(it), h('div', { class: 'flex flex-col gap-1 text-xs min-w-0' }, relWhenCell(it), relThenCell(it)), edit('meta-relations', it, null));
    }, add('meta-relations', { targets: [{ metaId: st.id, action: 'add', amount: 1 }] }, null)),
    studioGroup('prizes', 'prizesDuo', 'studio.g.prizes', d.prizes, function (it) {
      var mine = (it.items || []).filter(function (x) { return x.type === 'meta' && x.key === meta.key; })[0];
      return studioRow('prizes', it, null, mine ? h('div', { class: 'text-xs tabular-nums text-secondary-600 dark:text-secondary-300', dir: 'ltr', text: relSigned(mine.value, mine.action) }) : null, null, hashFor('prizes', it.id));
    }, null),
    studioGroup('levels-in', 'levelsDuo', 'studio.g.levelsIn', d.levelsIn, function (it) {
      return studioRow('level-tables', it, null, h('div', { class: 'flex items-center gap-1.5 text-xs min-w-0' }, refNameCell('metas', it.metaId)), edit('level-tables', it, null));
    }, null)
  ];
  var outputs = [
    studioGroup('relations-out', 'flowDuo', 'studio.g.relOut', d.relOut, function (it) {
      return studioRow('meta-relations', it, relRepeatTag(it), h('div', { class: 'flex flex-col gap-1 text-xs min-w-0' }, relWhenCell(it), relThenCell(it)), edit('meta-relations', it, 'metaId'));
    }, add('meta-relations', { metaId: st.id }, 'metaId')),
    studioGroup('level-tables', 'levelsDuo', 'studio.g.levels', d.levels, function (it) {
      var steps = it.steps || [];
      var top = steps[steps.length - 1];
      return studioRow('level-tables', it, null, h('div', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums' }, t('lv.levelsN', { n: steps.length }) + (top ? ' · ' + t('lv.top', { level: top.level, n: numLocal(top.threshold) }) : '')), edit('level-tables', it, 'metaId'));
    }, add('level-tables', { metaId: st.id }, 'metaId'))
  ];
  var arrow = function (key) {
    return h('div', { class: 'eco-studio-arrow flex items-center gap-2 text-xs text-secondary-500 dark:text-secondary-400', 'aria-hidden': 'true' }, icon('arrowDownDuo', 'h-4 w-4'), h('span', { text: t(key) }));
  };
  clear(st.body).appendChild(h('div', { class: 'flex flex-col gap-3' },
    h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: t('studio.hint') }),
    h('div', { class: 'eco-studio-lane flex flex-col gap-3', 'data-studio-lane': 'in' }, h('div', { class: 'eco-studio-lane-title text-xs font-semibold', text: t('studio.in') }), inputs),
    arrow('studio.arrow.in'),
    node,
    arrow('studio.arrow.out'),
    h('div', { class: 'eco-studio-lane flex flex-col gap-3', 'data-studio-lane': 'out' }, h('div', { class: 'eco-studio-lane-title text-xs font-semibold', text: t('studio.out') }), outputs)));
}

function lockFields(fields, lock, meta) {
  if (!lock) return fields;
  return fields.map(function walk(f) {
    if (f.type === 'row') return Object.assign({}, f, { fields: f.fields.map(walk) });
    if (f.name !== lock) return f;
    return { name: lock, type: 'static', readonly: true, label: f.label, render: function () {
      return h('div', { class: 'eco-studio-lock inline-flex items-center gap-2 rounded-md px-2 py-1 text-sm', 'data-locked': lock },
        icon('lockDuo', 'h-4 w-4'), h('span', { class: 'font-medium', text: i18nText(meta.name) || meta.key }), mono(meta.key));
    } };
  });
}

function studioEditor(name, itemId, preset, lock) {
  var st = STUDIO;
  var def = RESOURCES[name];
  if (!st || !def) return;
  st.editing = { name: name, id: itemId };
  var isNew = !itemId;
  studioTitle(t((isNew ? 'new.' : 'edit.') + name));
  var top = h('div', {});
  var back = button({ variant: 'link', size: 'sm', icon: 'back', label: t('studio.back'), onclick: studioOverview });
  back.setAttribute('data-studio-back', '');
  clear(st.body).appendChild(h('div', {}, h('div', { class: 'mb-4' }, back), formSkeleton(def)));
  clear(st.footer);
  var load = isNew ? Promise.resolve(Object.assign(def.defaults({}), preset || {})) : api('GET', '/' + name + '/' + encodeURIComponent(itemId));
  Promise.all([load, freshLanguages(), loadKeyTypes()]).then(function (all) {
    if (STUDIO !== st || !st.editing || st.editing.name !== name) return;
    var values = clone(all[0]) || {};
    if (def.toForm) values = def.toForm(values);
    if (lock) values[lock] = st.id;
    var form = buildForm(lockFields(def.fields, lock, st.meta), values);
    clear(st.body).appendChild(h('div', { 'data-studio-editor': name },
      h('div', { class: 'mb-4' }, back),
      h('form', { novalidate: true, onsubmit: function (e) { e.preventDefault(); save(); } }, top, form.el, h('button', { type: 'submit', hidden: true }))));
    var saveBtn = button({ variant: 'primary', size: 'footer', label: isNew ? t('create') : t('save'), onclick: function () { save(); } });
    saveBtn.setAttribute('data-studio-save', '');
    var del = h('div', { class: 'flex items-center gap-2' });
    if (!isNew) {
      var resetDel = function () {
        clear(del).appendChild(button({ variant: 'danger', size: 'footer', icon: 'trash', label: t('delete'), onclick: function () {
          clear(del).appendChild(h('span', { class: 'text-sm', text: t('delete.confirm') }));
          del.appendChild(button({ variant: 'dangerSolid', size: 'sm', label: t('delete.yes'), onclick: function () {
            api('DELETE', '/' + name + '/' + encodeURIComponent(itemId)).then(function () {
              toast('success', t('deleted'));
              if (def.afterSave) def.afterSave();
              studioOverview();
            }, function (err) { resetDel(); clear(top).appendChild(h('div', { class: 'mb-5' }, notice('error', errorText(err)))); });
          } }));
          del.appendChild(button({ variant: 'ghost', size: 'sm', label: t('no'), onclick: resetDel }));
        } }));
      };
      resetDel();
    }
    clear(st.footer).appendChild(h('div', { class: 'flex-1 flex items-center' }, del));
    st.footer.appendChild(button({ variant: 'ghost', size: 'footer', label: t('cancel'), onclick: studioOverview }));
    st.footer.appendChild(saveBtn);
    function save() {
      clear(top);
      if (!form.validate()) return;
      var payload = form.payload();
      if (def.fixPayload) payload = def.fixPayload(payload);
      if (lock) payload[lock] = st.id;
      setBusy(saveBtn, true, t('saving'));
      api(isNew ? 'POST' : 'PUT', '/' + name + (isNew ? '' : '/' + encodeURIComponent(itemId)), payload).then(function () {
        toast('success', t('saved'));
        if (def.afterSave) def.afterSave();
        studioOverview();
      }, function (err) {
        setBusy(saveBtn, false);
        if (!form.showServerError(err)) clear(top).appendChild(h('div', { class: 'mb-5' }, notice('error', errorText(err))));
      });
    }
    var first = st.body.querySelector('form input:not([readonly]),form textarea,form select');
    if (first) first.focus();
  }, function (err) {
    if (STUDIO !== st) return;
    clear(st.body).appendChild(h('div', {}, h('div', { class: 'mb-4' }, back), errorBlock(err, function () { studioEditor(name, itemId, preset, lock); })));
  });
}

function changeMetaDialog(userId) {
  var values = { metaId: null, action: 'add', value: 1 };
  var picker = refPicker({ resource: 'metas', value: null, required: true, label: t('f.mr.meta'), onchange: function (v) { values.metaId = v || null; } });
  var action = selectInput('add', opts('rel.action.', ['add', 'set'])(), function (v) { values.action = v; }, { placeholder: false, label: t('f.rel.action') });
  var val = numberInput(1, { int: false, label: t('f.rel.amount'), oninput: function (v) { values.value = v === '' ? null : parseNumber(v, false); } });
  var metaWrap = field(t('f.mr.meta'), picker, null);
  var valWrap = field(t('f.rel.amount'), val, t('pp.meta.hint'));
  return formDialog({
    id: 'change-meta',
    title: t('pp.meta.change'),
    text: t('pp.meta.change.text'),
    icon: 'metas',
    ok: t('pp.meta.change.ok'),
    body: h('div', {}, metaWrap, field(t('f.rel.action'), action, null), valWrap),
    validate: function () {
      metaWrap.setError(values.metaId ? '' : t('err.required'));
      valWrap.setError(values.value === null || isNaN(values.value) ? t('err.number') : '');
      return !!values.metaId && values.value !== null && !isNaN(values.value);
    },
    submit: function () { return api('POST', '/users/' + encodeURIComponent(userId) + '/metas', { metaId: values.metaId, action: values.action, value: values.value }); }
  });
}

function relationsProgressSection(p, id, write, reload) {
  var box = h('div', { class: 'flex flex-col gap-1', 'data-pp-relations': '' });
  (p.relations || []).slice(0, 8).forEach(function (r) {
    var line = h('div', { class: 'flex flex-wrap items-center gap-2 text-sm min-w-0', 'data-pp-relation': r.key },
      h('span', { class: 'truncate font-medium', text: i18nText(r.name) || r.key }),
      r.fires > 1 ? tag(t('pp.rel.times', { n: r.fires }), 'primary') : null,
      h('span', { class: 'flex-1' }),
      h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: fmtDate(r.lastDoneAt) }));
    if (write) {
      var reset = button({ variant: 'link', size: 'sm', label: t('pp.rel.reset'), onclick: function () {
        confirmModal({ title: t('pp.rel.reset.title'), text: t('pp.rel.reset.text', { name: i18nText(r.name) || r.key }), ok: t('pp.rel.reset'), danger: true }).then(function (yes) {
          if (!yes) return;
          api('DELETE', '/users/' + encodeURIComponent(id) + '/meta-relations/' + encodeURIComponent(r.relationId)).then(function () { toast('success', t('pp.rel.reset.done')); reload(true); }, function (err) { toast('error', errorText(err)); });
        });
      } });
      reset.setAttribute('data-relation-reset', r.key);
      line.appendChild(reset);
    }
    box.appendChild(line);
  });
  if (!(p.relations || []).length) box.appendChild(h('span', { class: 'text-sm text-secondary-400', text: t('pp.rel.none') }));
  return progressSection('pp.relations', box);
}

OVERVIEW_CATALOG.push('meta-relations');
