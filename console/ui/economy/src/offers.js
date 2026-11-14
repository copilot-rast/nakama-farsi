ICONS.offers = '<path d="M9 15l6-6"/><circle cx="9.5" cy="9.5" r=".5"/><circle cx="14.5" cy="14.5" r=".5"/><path d="M5 7.2A2.2 2.2 0 0 1 7.2 5h1a2.2 2.2 0 0 0 1.55-.64l.7-.7a2.2 2.2 0 0 1 3.12 0l.7.7c.42.41.98.64 1.55.64h1a2.2 2.2 0 0 1 2.2 2.2v1c0 .58.23 1.13.64 1.55l.7.7a2.2 2.2 0 0 1 0 3.12l-.7.7a2.2 2.2 0 0 0-.64 1.55v1a2.2 2.2 0 0 1-2.2 2.2h-1a2.2 2.2 0 0 0-1.55.64l-.7.7a2.2 2.2 0 0 1-3.12 0l-.7-.7a2.2 2.2 0 0 0-1.55-.64h-1a2.2 2.2 0 0 1-2.2-2.2v-1a2.2 2.2 0 0 0-.64-1.55l-.7-.7a2.2 2.2 0 0 1 0-3.12l.7-.7A2.2 2.2 0 0 0 5 8.2z"/>';
PH_ICONS.offers = '<path d="M243.31,136,144,36.69A15.86,15.86,0,0,0,132.69,32H40a8,8,0,0,0-8,8v92.69A15.86,15.86,0,0,0,36.69,144L136,243.31a16,16,0,0,0,22.63,0l84.68-84.68a16,16,0,0,0,0-22.63Zm-96,96L48,132.69V48h84.69L232,147.31ZM96,84A12,12,0,1,1,84,72,12,12,0,0,1,96,84Z"/>';
ICONS['entry-offers'] = '<circle cx="12" cy="13" r="8"/><path d="M12 9v4l2 2"/><path d="M5 3L2 6M22 6l-3-3"/>';
ICONS['offer-settings'] = '<path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6"/>';
ICONS['offer-preview'] = '<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/>';
ICONS['offer-report'] = '<path d="M3 3v18h18"/><path d="M7 15l4-4 3 3 5-6"/>';

var OFFER_TRIGGERS = ['first_entry', 'level', 'meta', 'tag', 'account_days', 'lapsed'];
var OFFER_TRIGGER_VALUE = { level: 1, meta: 1, account_days: 1, lapsed: 1 };
var OFFER_REASON_COLOR = { picked: 'success', recent: 'warning', kept: 'primary' };

function offerMinutes(n) {
  n = Number(n) || 0;
  var d = Math.floor(n / 1440), hr = Math.floor((n % 1440) / 60), m = n % 60;
  var parts = [];
  if (d) parts.push(t('of.dur.d', { n: d }));
  if (hr) parts.push(t('of.dur.h', { n: hr }));
  if (m || !parts.length) parts.push(t('of.dur.m', { n: m }));
  return parts.join(' ');
}

function offerLeft(seconds) {
  if (!seconds || seconds <= 0) return t('of.ended');
  return t('of.left', { time: offerMinutes(Math.ceil(seconds / 60)) });
}

function offerPriceCell(it) {
  var box = h('div', { class: 'flex flex-col tabular-nums whitespace-nowrap text-sm', 'data-offer-price': it.priceMode || 'discount' });
  if (it.priceMode === 'price') {
    if (it.priceToman != null) box.appendChild(h('span', { text: t('toman', { n: it.priceToman }) }));
    if (it.coinPrice && it.coinPrice.currencyKey) box.appendChild(withKeyNames(function () { return h('span', { text: t('amount.of', { n: fmtNum(it.coinPrice.amount), name: keyName('currency', it.coinPrice.currencyKey) }) }); }));
    box.appendChild(h('span', { class: 'text-xs text-secondary-500', text: t('of.mode.price') }));
  } else {
    box.appendChild(h('span', { class: 'eco-text-success font-medium', text: it.discountPercent ? t('of.off', { n: it.discountPercent }) : t('of.productPrice') }));
  }
  return box;
}

function offerStatusCell(it) {
  var ended = it.endsAt && new Date(it.endsAt) < new Date();
  return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive),
    (it.audienceTagIds || []).length ? tag(t('audience.only'), 'primary') : null,
    it.productActive === false ? tag(t('of.offerOnly'), 'secondary') : null,
    ended ? tag(t('ev.ended'), 'secondary') : null);
}

function offerTriggerText(it) {
  var tt = it.triggerType || 'first_entry';
  var v = it.triggerValue == null ? '' : numLocal(it.triggerValue);
  return t('of.trigger.' + tt + '.say', { n: v });
}

function offerSharedColumns(kind) {
  var cols = [
    { label: 'col.name', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) || i18nText(it.productName) || it.key }), h('div', { class: 'text-secondary-500' }, mono(it.key))); }, skeleton: '30%' },
    { label: 'col.of.product', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'truncate', text: i18nText(it.productName) || '' }), it.productKey ? h('div', { class: 'text-secondary-500' }, mono(it.productKey)) : null); }, skeleton: '25%' },
    { label: 'col.price', render: offerPriceCell, width: 140 }
  ];
  if (kind === 'daily') cols.push({ label: 'col.of.weight', render: function (it) { return h('span', { class: 'tabular-nums', text: fmtNum(it.weight) }); }, width: 80 });
  else cols.push({ label: 'col.of.trigger', render: function (it) { return h('div', { class: 'flex flex-col text-sm min-w-0' }, h('span', { class: 'truncate', text: offerTriggerText(it) }), h('span', { class: 'text-xs text-secondary-500', text: offerMinutes(it.durationMinutes) })); }, skeleton: '25%' });
  cols.push({ label: 'col.of.sold', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: fmtNum(it.bought || 0) + ' / ' + fmtNum(it.made || 0) }); }, width: 90 });
  cols.push({ label: 'col.status', render: offerStatusCell, width: 150 });
  return cols;
}

function offerPriceFields() {
  return [
    { type: 'heading', label: 'of.h.product' },
    { name: 'productId', type: 'ref', resource: 'products', label: 'f.of.product', hint: 'f.of.product.hint', required: true },
    { type: 'row', fields: [
      { name: 'priceMode', type: 'select', label: 'f.of.priceMode', required: true, noPlaceholder: true, options: opts('of.mode.', ['discount', 'price']) },
      { name: 'discountPercent', type: 'number', label: 'f.of.discount', hint: 'f.of.discount.hint', min: 0, max: 100, empty: 0, when: function (v) { return v.priceMode !== 'price'; } }
    ] },
    { name: 'priceToman', type: 'number', label: 'f.of.priceToman', hint: 'f.of.priceToman.hint', min: 0, optional: true, when: function (v) { return v.priceMode === 'price'; },
      check: function (n) { return n != null && n !== '' && Number(n) > 0 && Number(n) < 100 ? t('code.INVALID_PRICE') : null; } },
    { type: 'row', when: function (v) { return v.priceMode === 'price'; }, fields: [
      { name: 'coinPrice.currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.coinPrice.currency', optional: true, when: function (v) { return v.priceMode === 'price'; } },
      { name: 'coinPrice.amount', type: 'number', label: 'f.coinPrice.amount', min: 0, optional: true, when: function (v) { return v.priceMode === 'price'; } }
    ] },
    { type: 'row', fields: [
      { name: 'sku.bazaar', type: 'text', ltr: true, label: 'f.sku.bazaar', hint: 'f.of.sku.hint', optional: true },
      { name: 'sku.myket', type: 'text', ltr: true, label: 'f.sku.myket', optional: true }
    ] },
    { name: 'purchaseLimit', type: 'number', label: 'f.of.limit', hint: 'f.of.limit.hint', required: true, min: 1, max: 1000 }
  ];
}

function offerAudienceFields() {
  return [
    { type: 'heading', label: 'ev.h.who' },
    playerTagsField(),
    { name: 'excludeTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.of.exclude', hint: 'f.of.exclude.hint', optional: true, empty: [] },
    { type: 'row', fields: [
      { name: 'minLevel', type: 'number', label: 'f.minLevel', min: 0, empty: 0 },
      { name: 'maxLevel', type: 'number', label: 'f.of.maxLevel', min: 0, optional: true, check: function (v, vals) { return v != null && v !== '' && Number(v) < Number(vals.minLevel || 0) ? t('code.INVALID_LEVEL') : null; } }
    ] },
    { name: 'levelTableId', type: 'ref', resource: 'level-tables', label: 'f.of.levelTable', hint: 'f.of.levelTable.hint', optional: true },
    { name: 'metaConditions', type: 'rows', addLabel: 'of.cond.add', emptyText: 'of.cond.empty', max: 10, empty: [],
      newRow: function () { return { metaId: null, operator: 'gte', value: 1 }; },
      columns: [
        { name: 'metaId', type: 'ref', resource: 'metas', label: 'f.rel.meta', required: true },
        { name: 'operator', type: 'select', label: 'f.rel.operator', options: opts('rel.op.', REL_OPS) },
        { name: 'value', type: 'number', int: false, label: 'f.rel.value', required: true }
      ] },
    { type: 'row', fields: [
      { name: 'minAccountDays', type: 'number', label: 'f.of.minDays', hint: 'f.of.minDays.hint', min: 0, optional: true },
      { name: 'maxAccountDays', type: 'number', label: 'f.of.maxDays', min: 0, optional: true }
    ] },
    windowRow()
  ];
}

function offerHeadFields(kind) {
  return [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: kind === 'entry' ? 'f.of.sort.entry.hint' : 'f.sort.hint' }
    ] },
    tr('name', { label: 'f.of.name', hint: 'f.of.name.hint', optional: true, max: 128 }),
    tr('description', { label: 'f.of.description', multiline: true, max: 1000, optional: true }),
    tr('badge', { label: 'f.of.badge', hint: 'f.of.badge.hint', optional: true, max: 32 })
  ];
}

function offerDefaults(kind) {
  return function (query) {
    var v = { key: '', name: {}, description: {}, badge: {}, productId: (query && query.productId) || null, priceMode: 'discount', discountPercent: 20, priceToman: null, coinPrice: null, sku: {},
      purchaseLimit: 1, weight: 10, audienceTagIds: [], excludeTagIds: [], minLevel: 0, maxLevel: null, levelTableId: null, metaConditions: [], minAccountDays: null, maxAccountDays: null,
      startsAt: null, endsAt: null, isActive: true, sort: 0, metadata: {} };
    if (kind === 'entry') Object.assign(v, { triggerType: 'first_entry', triggerValue: null, triggerLevelTable: null, triggerMetaId: null, triggerTagId: null, durationMinutes: 1440, repeat: 'once', cooldownHours: 168, maxTriggers: null });
    return v;
  };
}

function offerToForm(v) {
  v.coinPrice = v.coinPrice || null;
  v.sku = v.sku || {};
  v.metaConditions = (v.metaConditions || []).map(function (c) { return { metaId: c.metaId, operator: c.operator || 'gte', value: c.value }; });
  if (v.kind === 'entry') {
    v.triggerLevelTable = v.triggerType === 'level' ? v.triggerRef || null : null;
    v.triggerMetaId = v.triggerType === 'meta' ? v.triggerRef || null : null;
    v.triggerTagId = v.triggerType === 'tag' ? v.triggerRef || null : null;
  }
  return v;
}

function offerFixPayload(kind) {
  return function (p) {
    if (p.sort == null) p.sort = 0;
    if (p.coinPrice && !p.coinPrice.currencyKey && (p.coinPrice.amount == null || p.coinPrice.amount === '')) p.coinPrice = null;
    if (p.priceMode !== 'price') { p.priceToman = null; p.coinPrice = null; }
    var sku = {};
    if (p.sku && p.sku.bazaar) sku.bazaar = p.sku.bazaar;
    if (p.sku && p.sku.myket) sku.myket = p.sku.myket;
    p.sku = sku;
    p.metaConditions = (p.metaConditions || []).map(function (c) { return { metaId: c.metaId, operator: c.operator || 'gte', value: c.value }; });
    if (kind === 'entry') {
      p.triggerRef = p.triggerType === 'level' ? p.triggerLevelTable || '' : p.triggerType === 'meta' ? p.triggerMetaId || '' : p.triggerType === 'tag' ? p.triggerTagId || '' : '';
      if (!OFFER_TRIGGER_VALUE[p.triggerType]) p.triggerValue = null;
      if (p.repeat !== 'cooldown') p.maxTriggers = null;
      delete p.triggerLevelTable;
      delete p.triggerMetaId;
      delete p.triggerTagId;
    }
    return p;
  };
}

RESOURCES.offers = {
  filters: [{ name: 'productId', label: 'filter.of.product', type: 'ref', resource: 'products' }, activeFilter()],
  columns: offerSharedColumns('daily'),
  fields: offerHeadFields('daily').concat(offerPriceFields(), [
    { type: 'heading', label: 'of.h.draw' },
    { name: 'weight', type: 'number', label: 'f.of.weight', hint: 'f.of.weight.hint', required: true, min: 0, max: 1000000 }
  ], offerAudienceFields(), [{ name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }]),
  defaults: offerDefaults('daily'),
  toForm: offerToForm,
  fixPayload: offerFixPayload('daily'),
  afterSave: function () { forgetRefs('offers'); }
};

RESOURCES['entry-offers'] = {
  filters: [{ name: 'productId', label: 'filter.of.product', type: 'ref', resource: 'products' }, activeFilter()],
  columns: offerSharedColumns('entry'),
  fields: offerHeadFields('entry').concat(offerPriceFields(), [
    { type: 'heading', label: 'of.h.trigger' },
    { type: 'row', fields: [
      { name: 'triggerType', type: 'select', label: 'f.of.trigger', hint: 'f.of.trigger.hint', required: true, noPlaceholder: true, options: opts('of.trigger.', OFFER_TRIGGERS) },
      { name: 'triggerValue', type: 'number', int: false, label: 'f.of.triggerValue', required: true, when: function (v) { return !!OFFER_TRIGGER_VALUE[v.triggerType]; } }
    ] },
    { name: 'triggerLevelTable', type: 'ref', resource: 'level-tables', label: 'f.of.levelTable', hint: 'f.of.levelTable.hint', optional: true, when: function (v) { return v.triggerType === 'level'; } },
    { name: 'triggerMetaId', type: 'ref', resource: 'metas', label: 'f.rel.meta', required: true, when: function (v) { return v.triggerType === 'meta'; } },
    { name: 'triggerTagId', type: 'ref', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.of.triggerTag', required: true, when: function (v) { return v.triggerType === 'tag'; } },
    { name: 'triggerSay', type: 'static', readonly: true, live: true, render: function (v) {
      return h('p', { class: 'eco-rule-preview text-sm rounded-md px-3 py-2', 'data-offer-trigger-say': v.triggerType || '', text: offerTriggerText(v) + ' ' + t('of.timer.say', { time: offerMinutes(v.durationMinutes) }) });
    } },
    { type: 'heading', label: 'of.h.timer' },
    { name: 'durationMinutes', type: 'number', label: 'f.of.duration', hint: 'f.of.duration.hint', required: true, min: 1, max: 525600 },
    { type: 'row', fields: [
      { name: 'repeat', type: 'select', label: 'f.of.repeat', hint: 'f.of.repeat.hint', required: true, noPlaceholder: true, options: opts('of.repeat.', ['once', 'cooldown']) },
      { name: 'cooldownHours', type: 'number', label: 'f.of.cooldown', hint: 'f.of.cooldown.hint', min: 0, max: 87600, empty: 0, when: function (v) { return v.repeat === 'cooldown'; } }
    ] },
    { name: 'maxTriggers', type: 'number', label: 'f.of.maxTriggers', hint: 'f.rel.maxFires.hint', min: 1, max: 1000000, optional: true, when: function (v) { return v.repeat === 'cooldown'; } }
  ], offerAudienceFields(), [{ name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }]),
  defaults: offerDefaults('entry'),
  toForm: offerToForm,
  fixPayload: offerFixPayload('entry'),
  afterSave: function () { forgetRefs('entry-offers'); }
};

var OFFER_SETTINGS_FIELDS = [
  { type: 'heading', label: 'os.h.daily' },
  { name: 'dailyEnabled', type: 'bool', inline: true, label: 'f.os.dailyEnabled', hint: 'f.os.dailyEnabled.hint' },
  { type: 'row', fields: [
    { name: 'timezone', type: 'select', label: 'f.es.timezone', required: true, options: function () { return TIMEZONES.map(function (z) { return { value: z, label: z === 'UTC' ? 'UTC' : z.replace('_', ' ') }; }); } },
    { name: 'periodDays', type: 'number', label: 'f.os.periodDays', hint: 'f.os.periodDays.hint', required: true, min: 1, max: 30 }
  ] },
  { type: 'row', fields: [
    { name: 'resetHour', type: 'number', label: 'f.os.resetHour', hint: 'f.os.resetHour.hint', required: true, min: 0, max: 23 },
    { name: 'resetMinute', type: 'number', label: 'f.os.resetMinute', required: true, min: 0, max: 59 }
  ] },
  { name: 'periodSay', type: 'static', readonly: true, live: true, render: function (v) {
    if (!v.periodEnd) return h('span', {});
    return h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 -mt-2 mb-4', 'data-offer-period': '' }, t('os.period', { from: fmtDate(v.periodStart), to: fmtDate(v.periodEnd) }));
  } },
  { name: 'count', type: 'number', label: 'f.os.count', hint: 'f.os.count.hint', required: true, min: 0, max: 20 },
  { name: 'countByTag', type: 'rows', label: 'f.os.countByTag', addLabel: 'os.countByTag.add', emptyText: 'os.countByTag.empty', max: 20, empty: [],
    newRow: function () { return { tagId: null, count: 4 }; },
    columns: [
      { name: 'tagId', type: 'ref', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, label: 'f.os.tag', required: true },
      { name: 'count', type: 'number', label: 'f.os.count', required: true, min: 0, max: 20 }
    ] },
  { type: 'row', fields: [
    { name: 'weighting', type: 'select', label: 'f.os.weighting', hint: 'f.os.weighting.hint', required: true, noPlaceholder: true, options: opts('os.weighting.', ['weighted', 'uniform', 'priority']) },
    { name: 'noRepeatDays', type: 'number', label: 'f.os.noRepeatDays', hint: 'f.os.noRepeatDays.hint', required: true, min: 0, max: 365 }
  ] },
  { name: 'fillWithRepeats', type: 'bool', inline: true, label: 'f.os.fillWithRepeats', hint: 'f.os.fillWithRepeats.hint' },
  { name: 'boughtCooldownDays', type: 'number', label: 'f.os.boughtCooldown', hint: 'f.os.boughtCooldown.hint', required: true, min: -1, max: 3650 },
  { name: 'excludeOwned', type: 'bool', inline: true, label: 'f.os.excludeOwned', hint: 'f.os.excludeOwned.hint' },
  { name: 'respectProductRules', type: 'bool', inline: true, label: 'f.os.productRules', hint: 'f.os.productRules.hint' },
  { name: 'inactiveProducts', type: 'select', label: 'f.os.inactive', hint: 'f.os.inactive.hint', required: true, noPlaceholder: true, options: opts('os.inactive.', ['allow', 'skip']) },
  { name: 'showSoldOut', type: 'bool', inline: true, label: 'f.os.showSoldOut', hint: 'f.os.showSoldOut.hint' },
  { name: 'giftCodes', type: 'bool', inline: true, label: 'f.os.giftCodes', hint: 'f.os.giftCodes.hint' },
  { type: 'heading', label: 'os.h.reroll' },
  { name: 'rerollEnabled', type: 'bool', inline: true, label: 'f.os.rerollEnabled', hint: 'f.os.rerollEnabled.hint' },
  { type: 'row', fields: [
    { name: 'rerollLimit', type: 'number', label: 'f.os.rerollLimit', hint: 'f.os.rerollLimit.hint', required: true, min: 0, max: 100 },
    { name: 'rerollCostStep', type: 'number', label: 'f.os.rerollStep', hint: 'f.os.rerollStep.hint', required: true, min: 0 }
  ] },
  { type: 'row', fields: [
    { name: 'rerollCost.currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.os.rerollCurrency', hint: 'f.os.rerollCurrency.hint', optional: true },
    { name: 'rerollCost.amount', type: 'number', label: 'f.os.rerollAmount', min: 1, optional: true }
  ] },
  { name: 'rerollAvoidCurrent', type: 'bool', inline: true, label: 'f.os.rerollAvoid', hint: 'f.os.rerollAvoid.hint' },
  { type: 'heading', label: 'os.h.entry' },
  { name: 'entryEnabled', type: 'bool', inline: true, label: 'f.os.entryEnabled', hint: 'f.os.entryEnabled.hint' },
  { type: 'row', fields: [
    { name: 'entryMaxActive', type: 'number', label: 'f.os.entryMax', hint: 'f.os.entryMax.hint', required: true, min: 1, max: 20 },
    { name: 'entryTimerStart', type: 'select', label: 'f.os.timerStart', hint: 'f.os.timerStart.hint', required: true, noPlaceholder: true, options: opts('os.timer.', ['list', 'seen', 'trigger']) }
  ] }
];

function offerSettingsCall(method, path, body) {
  return api(method, path, body).then(function (res) {
    var st = Object.assign({}, (res && res.settings) || {});
    st.periodStart = res && res.periodStart;
    st.periodEnd = res && res.periodEnd;
    if (!st.rerollCost) st.rerollCost = { currencyKey: null, amount: null };
    return st;
  });
}

function offerSettingsPage(container) {
  var frame = pageFrame(t('nav.offer-settings'), t('desc.offer-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(offerSettingsCall, '/offer-settings', null, OFFER_SETTINGS_FIELDS, { id: 'offers', fixPayload: function (p) {
    delete p.periodStart;
    delete p.periodEnd;
    delete p.updatedAt;
    delete p.periodSay;
    if (!p.rerollCost || !p.rerollCost.currencyKey) p.rerollCost = null;
    p.countByTag = (p.countByTag || []).map(function (r) { return { tagId: r.tagId, count: Number(r.count) }; });
    return p;
  } });
  frame.content.appendChild(page.el);
  page.load();
}

function offerReasonTag(reason) {
  if (!reason) return tag(t('of.reason.ok'), 'success');
  return tag(t('of.reason.' + reason), OFFER_REASON_COLOR[reason] || 'secondary');
}

function offerStateTag(r) {
  var color = r.state === 'active' ? 'success' : r.state === 'sold_out' ? 'primary' : r.state === 'waiting' ? 'warning' : 'secondary';
  return tag(t('of.state.' + r.state), color);
}

function offerRowPrice(r) {
  var parts = [];
  if (r.finalToman != null) parts.push(t('toman', { n: r.finalToman }));
  if (r.coinAmount != null && r.coinCurrency) parts.push(t('amount.of', { n: fmtNum(r.coinAmount), name: keyName('currency', r.coinCurrency) }));
  return h('span', { class: 'tabular-nums whitespace-nowrap text-sm', text: parts.join(' · ') + (r.discountPercent ? ' (' + t('of.off', { n: r.discountPercent }) + ')' : '') });
}

function offerCopyColumns() {
  return [
    { label: 'col.name', render: function (r) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(r.name) || r.key }), h('div', { class: 'text-secondary-500' }, mono(r.key))); }, skeleton: '30%' },
    { label: 'col.price', render: function (r) { return withKeyNames(function () { return offerRowPrice(r); }); }, skeleton: '20%' },
    { label: 'col.of.limitUse', render: function (r) { return h('span', { class: 'tabular-nums', text: fmtNum(r.bought) + ' / ' + fmtNum(r.purchaseLimit) }); }, width: 90 },
    { label: 'col.status', render: function (r) { return h('div', { class: 'flex flex-wrap gap-1' }, offerStateTag(r), r.reroll ? tag(t('of.rerollN', { n: r.reroll }), 'secondary') : null); }, width: 150 },
    { label: 'col.of.time', render: function (r) { return h('span', { class: 'text-sm whitespace-nowrap', text: r.state === 'active' ? offerLeft(r.secondsLeft) : r.expiresAt ? fmtDate(r.expiresAt) : '' }); }, width: 170 }
  ];
}

function offerPreviewPage(container, query) {
  var frame = pageFrame(t('nav.offer-preview'), t('desc.offer-preview'));
  clear(container).appendChild(frame.el);
  var userId = query.userId || '';
  var body = h('div', { class: 'flex flex-col gap-6', 'data-offer-preview': '' });
  var picker = refPicker({ resource: 'support/players', api: baseApi, value: userId, label: t('col.player'), placeholder: t('of.pickPlayer'), onchange: function (v) {
    userId = v || '';
    replaceHash(hashFor('offer-preview', null, userId ? { userId: userId } : {}));
    load();
  } });
  var idInput = h('input', { class: INPUT_SM + ' font-mono', dir: 'ltr', 'data-ltr': '', type: 'text', value: userId, placeholder: t('of.userId'), 'aria-label': t('of.userId'), 'data-offer-user': '' });
  idInput.addEventListener('change', function () {
    userId = toLatinDigits(idInput.value).trim();
    replaceHash(hashFor('offer-preview', null, userId ? { userId: userId } : {}));
    load();
  });
  frame.content.appendChild(h('div', { class: 'flex flex-col sm:flex-row gap-2 mb-4 eco-offer-pick' },
    h('div', { class: 'x-label x-input eco-filter-ref min-w-0 flex-1', style: INPUT_VARS }, picker),
    h('label', { class: 'x-label block relative x-input flex-1', style: INPUT_VARS }, idInput)));
  frame.content.appendChild(body);

  function render(res) {
    clear(body);
    body.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-offer-summary': '' },
      h('a', { href: '#/players/' + encodeURIComponent(res.userId), class: 'eco-link font-medium', text: res.username || res.userId }),
      tag(t('of.countN', { n: res.count }), 'primary'),
      !res.dailyEnabled ? tag(t('of.dailyOff'), 'warning') : null,
      !res.entryEnabled ? tag(t('of.entryOff'), 'warning') : null,
      h('span', { class: 'text-secondary-500', text: t('os.period', { from: fmtDate(res.periodStart), to: fmtDate(res.periodEnd) }) })));
    var cur = res.current || [];
    body.appendChild(section('of.pv.current', cur.length ? table(offerCopyColumns().map(function (c) { return { label: t(c.label), width: c.width }; }), cur.map(function (r) { return { item: r, cells: offerCopyColumns().map(function (c) { return c.render(r); }) }; })) : h('p', { class: 'text-sm text-secondary-500 py-2', text: t('of.pv.noCurrent') })));
    var drawCols = [{ label: t('col.name'), skeleton: '40%' }, { label: t('col.of.weight'), width: 90 }, { label: t('col.of.reason'), width: 200 }];
    body.appendChild(section('of.pv.draw', h('div', {},
      h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mb-3', text: t('of.pv.draw.hint') }),
      (res.draw || []).length ? table(drawCols, (res.draw || []).map(function (r) {
        return { item: r, cells: [
          h('div', { class: 'min-w-0', 'data-offer-draw': r.key }, h('div', { class: 'font-medium truncate', text: i18nText(r.name) || r.key }), h('div', { class: 'text-secondary-500' }, mono(r.key))),
          h('span', { class: 'tabular-nums', text: fmtNum(r.weight) }),
          h('div', { class: 'flex flex-wrap gap-1', 'data-offer-reason': r.reason || 'ok' }, offerReasonTag(r.reason), r.picked ? tag(t('of.slotN', { n: (r.slot || 0) + 1 }), 'secondary') : null)
        ] };
      })) : h('p', { class: 'text-sm text-secondary-500 py-2', text: t('of.pv.noPool') }))));
    var entryCols = [{ label: t('col.name'), skeleton: '40%' }, { label: t('col.of.trigger'), width: 180 }, { label: t('col.of.reason'), width: 220 }];
    body.appendChild(section('of.pv.entry', (res.entry || []).length ? table(entryCols, res.entry.map(function (e) {
      return { item: e, cells: [
        h('div', { class: 'min-w-0', 'data-offer-entry': e.key }, h('div', { class: 'font-medium truncate', text: i18nText(e.name) || e.key }), h('div', { class: 'text-secondary-500' }, mono(e.key))),
        h('span', { class: 'text-sm', text: t('of.trigger.' + (e.triggerType || 'first_entry')) }),
        h('div', { class: 'flex flex-wrap gap-1', 'data-offer-entry-state': e.due ? 'due' : e.open ? 'open' : e.reason || (e.holds ? 'done' : 'waiting') },
          e.open ? tag(t('of.entry.open'), 'success') : e.due ? tag(t('of.entry.due'), 'primary') : e.reason ? offerReasonTag(e.reason) : tag(t(e.holds ? 'of.entry.done' : 'of.entry.waiting'), 'secondary'),
          e.count ? tag(t('of.timesN', { n: e.count }), 'secondary') : null)
      ] };
    })) : h('p', { class: 'text-sm text-secondary-500 py-2', text: t('of.pv.noEntry') })));
  }

  function load() {
    if (!userId) { clear(body).appendChild(card(emptyBlock(t('of.pv.pick'), t('of.pv.pick.hint')))); return; }
    idInput.value = userId;
    clear(body).appendChild(h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('70%'), bar('55%')));
    api('GET', '/offer-preview' + qs({ userId: userId })).then(render, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}

function offerReportPage(container, query) {
  var frame = pageFrame(t('nav.offer-report'), t('desc.offer-report'));
  clear(container).appendChild(frame.el);
  var state = { days: query.days || '30', kind: query.kind || '' };
  var stats = h('div', { class: 'eco-stat-grid mb-4', 'data-offer-stats': '' });
  var box = h('div', {});
  function sync() {
    var q = {};
    if (state.days !== '30') q.days = state.days;
    if (state.kind) q.kind = state.kind;
    replaceHash(hashFor('offer-report', null, q));
  }
  var days = filterSelect(t('filter.of.days'), state.days, [7, 30, 90, 365].map(function (n) { return { value: String(n), label: t('of.daysN', { n: n }) }; }), function (v) { state.days = v || '30'; sync(); load(); });
  var kind = filterSelect(t('filter.of.kind'), state.kind, opts('of.kind.', ['daily', 'entry'])(), function (v) { state.kind = v || ''; sync(); load(); });
  frame.content.appendChild(h('div', { class: 'eco-filters flex flex-wrap items-center gap-2 pb-2' }, days, kind));
  frame.content.appendChild(stats);
  frame.content.appendChild(box);
  var cols = [
    { label: t('col.name'), skeleton: '30%' }, { label: t('col.of.kind'), width: 100 }, { label: t('col.of.shown'), width: 90 }, { label: t('col.of.buyers'), width: 90 },
    { label: t('col.of.purchases'), width: 90 }, { label: t('col.of.conversion'), width: 100 }, { label: t('col.of.revenue'), width: 150 }
  ];
  function pct(x) { return t('percent', { n: Math.round((x || 0) * 1000) / 10 }); }
  function load() {
    clear(stats);
    for (var i = 0; i < 5; i++) stats.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem'), bar('3rem', 'h-7'))));
    clear(box).appendChild(table(cols, skeletonRows(cols, 5)));
    api('GET', '/offer-report' + qs({ days: state.days, kind: state.kind })).then(function (res) {
      var tt = res.totals || {};
      clear(stats);
      stats.appendChild(statCard(t('col.of.shown'), tt.shown, t('of.madeN', { n: fmtNum(tt.made || 0) }), null));
      stats.appendChild(statCard(t('col.of.buyers'), tt.buyers, null, null));
      stats.appendChild(statCard(t('col.of.purchases'), tt.purchases, null, null));
      var conv = statCard(t('col.of.conversion'), 0, null, null);
      var num = conv.querySelector('.text-3xl');
      if (num) num.textContent = pct(tt.conversion);
      stats.appendChild(conv);
      stats.appendChild(statCard(t('col.of.revenue'), tt.revenueToman, t('toman', { n: '' }).trim(), null));
      stats.appendChild(statCard(t('of.rerolls'), tt.rerolls, null, null));
      var items = res.items || [];
      clear(box);
      if (!items.length) { box.appendChild(card(emptyBlock(t('empty.title'), t('of.report.empty')))); return; }
      box.appendChild(table(cols, items.map(function (r) {
        var coins = Object.keys(r.coins || {}).map(function (k) { return fmtNum(r.coins[k]) + ' ' + k; });
        return { item: r, cells: [
          h('a', { href: hashFor(r.kind === 'entry' ? 'entry-offers' : 'offers', r.id), class: 'eco-link min-w-0', 'data-offer-report': r.key }, h('div', { class: 'font-medium truncate', text: i18nText(r.name) || r.key }), h('div', { class: 'text-secondary-500' }, mono(r.key))),
          tag(t('of.kind.' + r.kind), r.kind === 'entry' ? 'warning' : 'primary'),
          h('span', { class: 'tabular-nums', text: fmtNum(r.shown) }),
          h('span', { class: 'tabular-nums', text: fmtNum(r.buyers) }),
          h('span', { class: 'tabular-nums', text: fmtNum(r.purchases) }),
          h('span', { class: 'tabular-nums', 'data-offer-conv': r.key, text: pct(r.conversion) }),
          h('div', { class: 'flex flex-col tabular-nums text-sm' }, r.revenueToman ? h('span', { text: t('toman', { n: r.revenueToman }) }) : null, coins.length ? h('span', { class: 'text-xs text-secondary-500', dir: 'ltr', text: coins.join(' · ') }) : null, !r.revenueToman && !coins.length ? h('span', { class: 'text-secondary-400', text: t('none') }) : null)
        ] };
      })));
    }, function (err) { clear(stats); clear(box).appendChild(errorBlock(err, load)); });
  }
  load();
}

CUSTOM_PAGES['offer-settings'] = function (root) { offerSettingsPage(root); };
CUSTOM_PAGES['offer-preview'] = function (root, route) { offerPreviewPage(root, route.query || {}); };
CUSTOM_PAGES['offer-report'] = function (root, route) { offerReportPage(root, route.query || {}); };

var OFFER_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadOfferCard(id) {
  OFFER_CARD.loading = true;
  return api('GET', '/users/' + encodeURIComponent(id) + '/offers').then(function (res) {
    OFFER_CARD.loading = false;
    OFFER_CARD.at = Date.now();
    OFFER_CARD.data = res;
    OFFER_CARD.err = null;
    var old = document.querySelector('[data-player-offers]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    OFFER_CARD.loading = false;
    OFFER_CARD.at = Date.now();
    OFFER_CARD.err = err;
    schedule();
  });
}

function reloadOfferCard(id) { OFFER_CARD.at = 0; loadOfferCard(id); }

function offerCardLine(userId, r, write) {
  var line = h('div', { class: 'flex flex-wrap items-center gap-2 py-1.5 text-sm', 'data-pp-offer': r.key, 'data-offer-state': r.state },
    h('span', { class: 'font-medium truncate min-w-0', text: i18nText(r.name) || r.key }),
    offerStateTag(r),
    withKeyNames(function () { return offerRowPrice(r); }),
    h('span', { class: 'flex-1' }),
    h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap tabular-nums', text: r.state === 'active' ? offerLeft(r.secondsLeft) : r.expiresAt ? fmtDate(r.expiresAt) : r.state === 'waiting' ? t('of.notStarted') : '' }));
  if (write && (r.state === 'active' || r.state === 'waiting')) {
    var close = button({ variant: 'link', size: 'sm', label: t('of.close'), onclick: function () {
      confirmModal({ title: t('of.close.title'), text: t('of.close.text', { name: i18nText(r.name) || r.key }), ok: t('of.close'), danger: true }).then(function (yes) {
        if (!yes) return;
        api('POST', '/users/' + encodeURIComponent(userId) + '/offers/' + encodeURIComponent(r.id) + '/expire', {}).then(function () { toast('success', t('of.close.done')); reloadOfferCard(userId); }, function (err) { toast('error', errorText(err)); });
      });
    } });
    close.setAttribute('data-offer-close', r.key);
    line.appendChild(close);
  }
  if (write && r.kind === 'entry' && !r.forgotten) {
    var reset = button({ variant: 'link', size: 'sm', label: t('of.entryReset'), onclick: function () {
      confirmModal({ title: t('of.entryReset.title'), text: t('of.entryReset.text', { name: i18nText(r.name) || r.key }), ok: t('of.entryReset'), danger: true }).then(function (yes) {
        if (!yes) return;
        api('POST', '/users/' + encodeURIComponent(userId) + '/entry-offers/' + encodeURIComponent(r.offerId) + '/reset', {}).then(function () { toast('success', t('of.entryReset.done')); reloadOfferCard(userId); }, function (err) { toast('error', errorText(err)); });
      });
    } });
    reset.setAttribute('data-offer-entry-reset', r.key);
    line.appendChild(reset);
  }
  return line;
}

function playerOffersCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-offers]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (OFFER_CARD.id !== id) { OFFER_CARD.id = id; OFFER_CARD.data = null; OFFER_CARD.err = null; OFFER_CARD.at = 0; }
  if (!OFFER_CARD.loading && Date.now() - OFFER_CARD.at > 20000) loadOfferCard(id);
  var p = OFFER_CARD.data;
  var anchor = grid.querySelector('[data-player-chance]') || grid.querySelector('[data-player-achievements]') || grid.querySelector('[data-player-items]') || grid.querySelector('[data-player-tasks]') || grid.querySelector('[data-player-activity]') || grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify([p.daily, p.entry, p.rerolls]) : OFFER_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-offers': '', 'data-key': key, 'data-economy': 'player-offers', 'data-l10n-skip': '', 'aria-labelledby': 'player-offers-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-offers-title', class: 'font-medium flex items-center gap-2' }, icon('offers', 'h-4 w-4'), h('span', { text: t('of.card') }));
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(OFFER_CARD.err ? notice('error', errorText(OFFER_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { reloadOfferCard(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var write = aclCan('HIRO_ECONOMY', 1);
  var actions = h('div', { class: 'flex flex-wrap items-center gap-2' });
  if (write) {
    var redraw = button({ variant: 'ghost', size: 'sm', icon: 'refresh', label: t('of.redraw'), onclick: function () {
      confirmModal({ title: t('of.redraw.title'), text: t('of.redraw.text'), ok: t('of.redraw') }).then(function (yes) {
        if (!yes) return;
        api('POST', '/users/' + encodeURIComponent(id) + '/offers/reset-daily', {}).then(function () { toast('success', t('of.redraw.done')); reloadOfferCard(id); }, function (err) { toast('error', errorText(err)); });
      });
    } });
    redraw.setAttribute('data-offer-redraw', '');
    actions.appendChild(redraw);
  }
  actions.appendChild(h('a', { href: hashFor('offer-preview', null, { userId: id }), class: 'eco-link text-sm', 'data-offer-preview-link': '', text: t('of.previewLink') }));
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' }, title,
    h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300 tabular-nums', 'data-offer-counts': '', text: t('of.cardCounts', { shown: fmtNum(p.shown || 0), bought: fmtNum(p.bought || 0) }) }),
    actions));
  var daily = (p.daily || []).filter(function (r) { return r.status !== 'replaced'; });
  var dailyBox = h('div', { class: 'flex flex-col', 'data-pp-offers-daily': '' });
  daily.forEach(function (r) { dailyBox.appendChild(offerCardLine(id, r, write)); });
  if (!daily.length) dailyBox.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: p.dailyEnabled ? t('of.card.noDaily') : t('of.dailyOff') }));
  body.appendChild(progressSection('of.card.daily', dailyBox, p.rerolls ? tag(t('of.rerollsUsed', { n: p.rerolls }), 'secondary') : null));
  var entryBox = h('div', { class: 'flex flex-col', 'data-pp-offers-entry': '' });
  (p.entry || []).slice(0, 10).forEach(function (r) { entryBox.appendChild(offerCardLine(id, r, write)); });
  if (!(p.entry || []).length) entryBox.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('of.card.noEntry') }));
  body.appendChild(h('div', { class: 'mt-3 border-t border-secondary-200 dark:border-secondary-700 pt-4' }, progressSection('of.card.entry', entryBox)));
  place();
}

OVERVIEW_CATALOG.push('offers');
