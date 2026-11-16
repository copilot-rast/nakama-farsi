ICONS['tag-discounts'] = '<path d="M20.6 13.4l-7.2 7.2a2 2 0 0 1-2.8 0L3 13V3h10l7.6 7.6a2 2 0 0 1 0 2.8z"/><path d="M9 15l6-6"/><circle cx="8.5" cy="9" r=".6"/><circle cx="14.5" cy="15" r=".6"/>';
ICONS['tag-discount-preview'] = '<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/>';
ICONS['tag-discount-settings'] = '<path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6"/>';

function tdTargets(it) {
  var parts = [];
  if (it.allItems) parts.push(t('td.allItems'));
  else if ((it.itemKeys || []).length || (it.itemTypeIds || []).length) parts.push(t('td.itemsN', { n: fmtNum((it.itemKeys || []).length + (it.itemTypeIds || []).length) }));
  if (it.allProducts) parts.push(t('td.allProducts'));
  else if ((it.productIds || []).length) parts.push(t('td.productsN', { n: fmtNum(it.productIds.length) }));
  return parts.join(' · ');
}

RESOURCES['tag-discounts'] = {
  filters: [activeFilter(), { name: 'tagId', label: 'filter.tag', type: 'ref', resource: 'tags', api: baseApi, filter: { keyType: 'user' } }],
  columns: [
    { label: 'col.name', render: function (it) {
      return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) || it.key }), h('div', { class: 'text-secondary-500' }, mono(it.key)));
    }, skeleton: '35%' },
    { label: 'col.td.percent', render: function (it) { return h('span', { class: 'font-medium tabular-nums', 'data-td-percent': '', text: t('td.percentN', { n: fmtNum(it.percent) }) }); }, width: 90 },
    { label: 'col.td.who', render: function (it) {
      return h('div', { class: 'flex flex-col gap-1 min-w-0' }, h('div', { class: 'flex flex-wrap gap-1' }, (it.tags || []).map(function (g) { return tag(i18nText(g.name) || g.key, 'primary'); })),
        h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: t('fg.rule.players', { n: fmtNum(it.players || 0) }) }));
    }, skeleton: '25%' },
    { label: 'col.td.on', render: function (it) { return h('span', { class: 'text-sm', 'data-td-targets': '', text: tdTargets(it) }); }, skeleton: '25%' },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    tr('name', { label: 'f.td.name', hint: 'f.td.name.hint', required: true, max: 128 }),
    { type: 'row', fields: [
      { name: 'percent', type: 'number', label: 'f.td.percent', hint: 'f.td.percent.hint', required: true, min: 1, max: 100 },
      { name: 'sort', type: 'number', label: 'f.sort', min: -100000, max: 100000, empty: 0 }
    ] },
    { name: 'tagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.td.tags', hint: 'f.td.tags.hint', required: true, empty: [] },
    { type: 'heading', label: 'td.h.items' },
    { name: 'allItems', type: 'bool', inline: true, label: 'f.td.allItems', hint: 'f.td.allItems.hint' },
    { name: 'itemIds', type: 'refMulti', resource: 'items', label: 'f.td.items', hint: 'f.td.items.hint', optional: true, empty: [], when: function (v) { return !v.allItems; } },
    { name: 'itemTypeIds', type: 'refMulti', resource: 'categories', api: baseApi, filter: { keyType: 'item_type' }, label: 'f.td.itemTypes', hint: 'f.td.itemTypes.hint', optional: true, empty: [], when: function (v) { return !v.allItems; } },
    { type: 'heading', label: 'td.h.products' },
    { name: 'allProducts', type: 'bool', inline: true, label: 'f.td.allProducts', hint: 'f.td.allProducts.hint' },
    { name: 'productIds', type: 'refMulti', resource: 'products', label: 'f.td.products', hint: 'f.td.products.hint', optional: true, empty: [], when: function (v) { return !v.allProducts; } },
    { type: 'heading', label: 'td.h.time' },
    windowRow(),
    { name: 'note', type: 'static', readonly: true, render: function () { return h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', 'data-td-note': '', text: t('td.note') }); } }
  ],
  defaults: function () { return { key: '', name: {}, percent: 5, tagIds: [], allItems: false, itemIds: [], itemTypeIds: [], allProducts: false, productIds: [], startsAt: null, endsAt: null, isActive: true, sort: 0 }; },
  fixPayload: function (p) {
    if (p.sort == null) p.sort = 0;
    if (p.allItems) { p.itemIds = []; p.itemTypeIds = []; }
    delete p.itemKeys;
    if (p.allProducts) p.productIds = [];
    p.startsAt = p.startsAt || '';
    p.endsAt = p.endsAt || '';
    ['id', 'tags', 'players', 'note', 'createdAt', 'updatedAt'].forEach(function (k) { delete p[k]; });
    return p;
  },
  afterSave: function () { forgetRefs('tag-discounts'); }
};

var TD_SETTINGS_FIELDS = [
  { name: 'enabled', type: 'bool', inline: true, label: 'f.td.enabled', hint: 'f.td.enabled.hint' },
  { type: 'row', fields: [
    { name: 'combine', type: 'select', label: 'f.td.combine', hint: 'f.td.combine.hint', required: true, noPlaceholder: true, options: opts('td.combine.', ['best', 'sum']) },
    { name: 'maxPercent', type: 'number', label: 'f.td.max', hint: 'f.td.max.hint', required: true, min: 1, max: 100 }
  ] },
  { name: 'withOwnDiscount', type: 'select', label: 'f.td.own', hint: 'f.td.own.hint', required: true, noPlaceholder: true, options: opts('td.own.', ['stack', 'best']) },
  { name: 'storesNote', type: 'static', readonly: true, render: function () { return h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('td.stores') }); } }
];

function tdSettingsCall(method, path, body) {
  return api(method, path, body).then(function (res) { return Object.assign({}, (res && res.settings) || {}); });
}

function tagDiscountSettingsPage(container) {
  var frame = pageFrame(t('nav.tag-discount-settings'), t('desc.tag-discount-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(tdSettingsCall, '/tag-discount-settings', null, TD_SETTINGS_FIELDS, { id: 'tag-discounts', fixPayload: function (p) {
    delete p.updatedAt;
    delete p.storesNote;
    p.maxPercent = Number(p.maxPercent);
    return p;
  } });
  frame.content.appendChild(page.el);
  page.load();
}

function tagDiscountPreviewPage(container, query) {
  var frame = pageFrame(t('nav.tag-discount-preview'), t('desc.tag-discount-preview'));
  clear(container).appendChild(frame.el);
  var userId = query.userId || '';
  var body = h('div', { class: 'flex flex-col gap-6', 'data-td-preview': '' });
  var picker = refPicker({ resource: 'support/players', api: baseApi, value: userId, label: t('col.player'), placeholder: t('of.pickPlayer'), onchange: function (v) {
    userId = v || '';
    replaceHash(hashFor('tag-discount-preview', null, userId ? { userId: userId } : {}));
    load();
  } });
  var idInput = h('input', { class: INPUT_SM + ' font-mono', dir: 'ltr', 'data-ltr': '', type: 'text', value: userId, placeholder: t('of.userId'), 'aria-label': t('of.userId'), 'data-td-user': '' });
  idInput.addEventListener('change', function () {
    userId = toLatinDigits(idInput.value).trim();
    replaceHash(hashFor('tag-discount-preview', null, userId ? { userId: userId } : {}));
    load();
  });
  frame.content.appendChild(h('div', { class: 'flex flex-col sm:flex-row gap-2 mb-4 eco-offer-pick' },
    h('div', { class: 'x-label x-input eco-filter-ref min-w-0 flex-1', style: INPUT_VARS }, picker),
    h('label', { class: 'x-label block relative x-input flex-1', style: INPUT_VARS }, idInput)));
  frame.content.appendChild(body);

  function coin(cp) { return cp ? t('amount.of', { n: fmtNum(cp.amount), name: keyName('currency', cp.currencyKey) }) : ''; }

  function render(res) {
    clear(body);
    var rules = res.rules || [];
    body.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-td-rules': String(rules.length) },
      res.settings && !res.settings.enabled ? tag(t('td.off'), 'warning') : null,
      rules.length ? rules.map(function (r) { return tag((i18nText(r.name) || r.key) + ' · ' + t('td.percentN', { n: fmtNum(r.percent) }), 'primary'); }) : h('span', { class: 'text-secondary-500', text: t('td.pv.noRules') })));
    var cols = [{ label: t('col.name'), skeleton: '40%' }, { label: t('col.td.price'), width: 220 }, { label: t('col.td.percent'), width: 90 }];
    body.appendChild(section('td.pv.items', (res.items || []).length ? table(cols, res.items.map(function (r) {
      return { item: r, cells: [
        h('div', { class: 'min-w-0', 'data-td-item': r.key }, h('div', { class: 'font-medium truncate', text: i18nText(r.name) || r.key }), h('div', { class: 'text-secondary-500' }, mono(r.key))),
        h('div', { class: 'flex flex-col tabular-nums text-sm' }, h('span', { class: 'text-secondary-500', style: 'text-decoration:line-through', text: coin(r.price) }), h('span', { class: 'font-medium', text: coin(r.finalPrice) })),
        h('span', { class: 'tabular-nums', text: t('td.percentN', { n: fmtNum(r.percent) }) })
      ] };
    })) : h('p', { class: 'text-sm text-secondary-500 py-2', text: t('td.pv.none') })));
    body.appendChild(section('td.pv.products', (res.products || []).length ? table(cols, res.products.map(function (r) {
      var lines = [];
      if (r.priceToman != null) lines.push(h('span', {}, h('span', { class: 'text-secondary-500', style: 'text-decoration:line-through', text: t('toman', { n: r.priceToman }) }), ' ', h('span', { class: 'font-medium', text: t('toman', { n: r.finalToman }) })));
      if (r.coinPrice) lines.push(h('span', {}, h('span', { class: 'text-secondary-500', style: 'text-decoration:line-through', text: coin(r.coinPrice) }), ' ', h('span', { class: 'font-medium', text: coin(r.finalCoin) })));
      return { item: r, cells: [
        h('div', { class: 'min-w-0', 'data-td-product': r.key }, h('div', { class: 'font-medium truncate', text: i18nText(r.name) || r.key }), h('div', { class: 'text-secondary-500' }, mono(r.key))),
        h('div', { class: 'flex flex-col tabular-nums text-sm' }, lines),
        h('span', { class: 'tabular-nums', text: t('td.percentN', { n: fmtNum(r.totalPercent) }) })
      ] };
    })) : h('p', { class: 'text-sm text-secondary-500 py-2', text: t('td.pv.none') })));
  }

  function load() {
    if (!userId) { clear(body).appendChild(card(emptyBlock(t('of.pv.pick'), t('td.pv.pick.hint')))); return; }
    idInput.value = userId;
    clear(body).appendChild(h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('70%'), bar('55%')));
    api('GET', '/tag-discount-preview' + qs({ userId: userId })).then(render, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}

CUSTOM_PAGES['tag-discount-settings'] = function (root) { tagDiscountSettingsPage(root); };
CUSTOM_PAGES['tag-discount-preview'] = function (root, route) { tagDiscountPreviewPage(root, route.query || {}); };
