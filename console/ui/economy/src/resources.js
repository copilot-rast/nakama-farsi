var METHODS = ['zarinpal', 'zibal', 'bazaar', 'myket', 'in_app'];
var GATEWAYS = ['zarinpal', 'zibal'];
var STORES = ['bazaar', 'myket'];
var STATUS_COLOR = { pending: 'warning', success: 'success', failed: 'error', expired: 'secondary' };

function opts(prefix, keys) {
  return function () { return keys.map(function (k) { return { value: k, label: t(prefix + k) }; }); };
}

function hasAny(list, wanted) {
  return (list || []).some(function (m) { return wanted.indexOf(m) !== -1; });
}

function nameCell(it, withIcon) {
  var name = i18nText(it.name) || it.name || '';
  var text = h('div', { class: 'min-w-0' },
    h('div', { class: 'font-medium truncate', text: typeof name === 'string' ? name : '' }),
    it.key || it.code ? h('div', { class: 'text-secondary-500' }, mono(it.key || it.code)) : null);
  if (!withIcon) return text;
  var pic = h('div', { class: 'h-8 w-8 shrink-0 rounded-md bg-secondary-100 dark:bg-secondary-700 overflow-hidden flex items-center justify-center text-secondary-400' });
  if (it.iconUrl) pic.appendChild(img(it.iconUrl));
  else pic.appendChild(icon('image', 'h-4 w-4'));
  return h('div', { class: 'flex items-center gap-3 min-w-0' }, pic, text);
}

function numCell(v) {
  return h('span', { class: 'tabular-nums', text: v === null || v === undefined ? '' : fmtNum(v) });
}

function dateCell(v) {
  return h('span', { class: 'text-secondary-600 dark:text-secondary-300 whitespace-nowrap', text: fmtDate(v) });
}

function activeTag(on) {
  return tag(on ? t('active') : t('inactive'), on ? 'success' : 'secondary');
}

function priceText(it) {
  var parts = [];
  if (it.priceToman !== null && it.priceToman !== undefined) parts.push(t('toman', { n: it.priceToman }));
  if (it.coinPrice && it.coinPrice.amount !== undefined && it.coinPrice.amount !== null) parts.push(fmtNum(it.coinPrice.amount) + ' ' + (it.coinPrice.currencyKey || ''));
  return parts;
}

function itemSummary(items) {
  return (items || []).map(function (r) {
    if (r.type === 'storage') return r.key;
    var sign = r.action === 'add' && Number(r.value) > 0 ? '+' : '';
    return (r.action === 'set' ? '= ' : sign) + fmtNum(r.value) + ' ' + r.key;
  });
}

var nameRow = function (base, required) {
  return { type: 'row', fields: [
    { name: base + '.fa', label: 'f.' + base + '.fa', required: required !== false },
    { name: base + '.en', label: 'f.' + base + '.en', ltr: true, optional: true }
  ] };
};
var nameTextareas = function (base) {
  return { type: 'row', fields: [
    { name: base + '.fa', type: 'textarea', label: 'f.' + base + '.fa', optional: true },
    { name: base + '.en', type: 'textarea', label: 'f.' + base + '.en', ltr: true, optional: true }
  ] };
};

var RESOURCES = {
  currencies: {
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '60%' },
      { label: 'col.type', render: function (it) { return tag(t('type.' + it.type), it.type === 'hard' ? 'warning' : 'primary'); }, width: 110 },
      { label: 'col.default', render: function (it) { return numCell(it.defaultValue); }, width: 130 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, keyRe: CURRENCY_KEY_RE, lockOnEdit: true },
        { name: 'type', type: 'select', label: 'f.type', required: true, options: opts('type.', ['soft', 'hard']) }
      ] },
      nameRow('name'),
      { type: 'row', fields: [
        { name: 'defaultValue', type: 'number', label: 'f.defaultValue', hint: 'f.defaultValue.currency.hint', min: 0 },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function () { return { type: 'soft', name: { fa: '', en: '' }, defaultValue: 0, sort: 0, metadata: {} }; }
  },
  metas: {
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '55%' },
      { label: 'col.default', render: function (it) { return numCell(it.defaultValue); }, width: 140 },
      { label: 'col.type', render: function (it) { return it.serverOnly ? tag(t('server.only'), 'warning') : ''; }, width: 130 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      nameRow('name'),
      { name: 'defaultValue', type: 'number', int: false, label: 'f.defaultValue', hint: 'f.defaultValue.meta.hint' },
      { name: 'serverOnly', type: 'bool', inline: true, label: 'f.serverOnly', hint: 'f.serverOnly.hint' },
      { name: 'settings', type: 'json', label: 'f.settings', optional: true, empty: {} }
    ],
    defaults: function () { return { name: { fa: '', en: '' }, defaultValue: 0, serverOnly: false, settings: {} }; }
  },
  prizes: {
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '55%' },
      { label: 'col.items', render: function (it) {
        var s = itemSummary(it.items);
        return h('div', { class: 'flex flex-wrap gap-1' }, s.slice(0, 3).map(function (x) { return tag(x, 'secondary'); }), s.length > 3 ? tag('+' + fmtNum(s.length - 3), 'secondary') : null);
      }, skeleton: '40%' },
      { label: 'col.type', render: function (it) { return tag(it.isRepeatable ? t('repeatable') : t('once'), it.isRepeatable ? 'primary' : 'secondary'); }, width: 120 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
      nameRow('name'),
      nameTextareas('message'),
      { type: 'row', fields: [
        { name: 'categoryId', type: 'ref', resource: 'categories', filter: { kind: 'prize' }, label: 'f.category', optional: true },
        { name: 'tagIds', type: 'refMulti', resource: 'tags', label: 'f.tags', optional: true, empty: [] }
      ] },
      { name: 'isRepeatable', type: 'bool', inline: true, label: 'f.isRepeatable', hint: 'f.isRepeatable.hint' },
      { name: 'items', type: 'items', label: 'f.items' },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function () { return { name: { fa: '', en: '' }, message: { fa: '', en: '' }, tagIds: [], isRepeatable: false, items: [{ type: 'currency', key: '', action: 'add', value: 0 }], metadata: {} }; }
  },
  products: {
    filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '60%' },
      { label: 'col.price', render: function (it) { return h('div', { class: 'flex flex-col tabular-nums whitespace-nowrap' }, priceText(it).map(function (p) { return h('span', { text: p }); }), it.discountPercent ? h('span', { class: 'text-xs text-success-600', text: t('percent', { n: it.discountPercent }) }) : null); }, width: 160 },
      { label: 'col.methods', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, (it.paymentMethods || []).map(function (m) { return tag(t('method.' + m), 'secondary'); })); }, skeleton: '45%' },
      { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 80 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
        { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
      ] },
      nameRow('name'),
      nameTextareas('description'),
      { type: 'row', fields: [
        { name: 'categoryId', type: 'ref', resource: 'categories', filter: { kind: 'product' }, label: 'f.category', optional: true },
        { name: 'tagIds', type: 'refMulti', resource: 'tags', label: 'f.tags', optional: true, empty: [] }
      ] },
      { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', hint: 'f.prize.product.hint', required: true },
      { type: 'heading', label: 'f.paymentMethods' },
      { name: 'paymentMethods', type: 'multi', required: true, options: opts('method.', METHODS) },
      { type: 'row', fields: [
        { name: 'priceToman', type: 'number', label: 'f.priceToman', hint: 'f.priceToman.hint', min: 0,
          when: function (v) { return hasAny(v.paymentMethods, GATEWAYS.concat(STORES)); },
          check: function (n, v) { return hasAny(v.paymentMethods, GATEWAYS) && (n === null || n < 100) ? t('f.priceToman.hint') : null; } },
        { name: 'discountPercent', type: 'number', label: 'f.discountPercent', min: 0, max: 100, empty: 0,
          when: function (v) { return (v.paymentMethods || []).length > 0; } }
      ] },
      { type: 'row', when: function (v) { return hasAny(v.paymentMethods, ['in_app']); }, fields: [
        { name: 'coinPrice.currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.coinPrice.currency', required: true, when: function (v) { return hasAny(v.paymentMethods, ['in_app']); } },
        { name: 'coinPrice.amount', type: 'number', label: 'f.coinPrice.amount', required: true, min: 1, when: function (v) { return hasAny(v.paymentMethods, ['in_app']); } }
      ] },
      { type: 'row', when: function (v) { return hasAny(v.paymentMethods, STORES); }, fields: [
        { name: 'sku.bazaar', type: 'text', ltr: true, label: 'f.sku.bazaar', required: true, when: function (v) { return hasAny(v.paymentMethods, ['bazaar']); } },
        { name: 'sku.myket', type: 'text', ltr: true, label: 'f.sku.myket', required: true, when: function (v) { return hasAny(v.paymentMethods, ['myket']); } }
      ] },
      { name: 'isSubscription', type: 'bool', inline: true, label: 'f.isSubscription', hint: 'f.isSubscription.hint', empty: false, when: function (v) { return hasAny(v.paymentMethods, ['bazaar']); } },
      { type: 'heading', label: 'f.icon' },
      { type: 'row', fields: [
        { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
        { name: 'coverUrl', type: 'image', label: 'f.cover', optional: true }
      ] },
      { type: 'row', fields: [
        { name: 'minLevel', type: 'number', label: 'f.minLevel', min: 0, empty: 0 },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function () { return { isActive: true, name: { fa: '', en: '' }, description: { fa: '', en: '' }, tagIds: [], paymentMethods: ['zarinpal'], discountPercent: 0, sku: {}, coinPrice: null, isSubscription: false, minLevel: 0, sort: 0, metadata: {} }; },
    fixPayload: function (p) {
      if (p.coinPrice && p.coinPrice.currencyKey == null && p.coinPrice.amount == null) p.coinPrice = null;
      var sku = {};
      if (p.sku && p.sku.bazaar) sku.bazaar = p.sku.bazaar;
      if (p.sku && p.sku.myket) sku.myket = p.sku.myket;
      p.sku = sku;
      return p;
    }
  },
  gifts: {
    filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
    columns: [
      { label: 'col.name', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: it.name || '' }), it.code ? h('div', { class: 'text-secondary-500' }, mono(it.code)) : null); }, skeleton: '50%' },
      { label: 'col.type', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('gift.' + it.type), it.type === 'discount' ? 'warning' : 'primary'), it.type === 'discount' && it.discountPercent ? tag(t('percent', { n: it.discountPercent }), 'secondary') : null); }, width: 150 },
      { label: 'col.usage', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: fmtNum(it.totalUsed || 0) + ' / ' + (it.maxTotalUses ? fmtNum(it.maxTotalUses) : t('unlimited')) }); }, width: 130 },
      { label: 'col.expire', render: function (it) { return dateCell(it.expireAt); }, width: 190 },
      { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'name', label: 'f.gift.name', required: true },
        { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
      ] },
      { type: 'row', fields: [
        { name: 'type', type: 'select', label: 'f.type', required: true, options: opts('gift.', ['reward', 'discount']) },
        { name: 'code', type: 'code', label: 'f.code', hint: 'f.code.hint', optional: true, check: function (c, v) { return v.type === 'discount' && !c ? t('code.CODE_REQUIRED') : null; } }
      ] },
      { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', required: true, when: function (v) { return v.type === 'reward'; } },
      { name: 'discountPercent', type: 'number', label: 'f.discountPercent', required: true, min: 1, max: 100, when: function (v) { return v.type === 'discount'; } },
      { type: 'row', fields: [
        { name: 'perUserLimit', type: 'number', label: 'f.perUserLimit', required: true, min: 1 },
        { name: 'maxTotalUses', type: 'number', label: 'f.maxTotalUses', hint: 'f.maxTotalUses.hint', min: 1, optional: true }
      ] },
      { type: 'row', fields: [
        { name: 'validFrom', type: 'datetime', label: 'f.validFrom', optional: true },
        { name: 'expireAt', type: 'datetime', label: 'f.expireAt', optional: true }
      ] },
      { name: 'isPromo', type: 'bool', inline: true, label: 'f.isPromo' },
      { name: 'totalUsed', type: 'static', label: 'f.totalUsed', readonly: true, when: function (v) { return !!v.id; }, render: function (v) { return numCell(v.totalUsed || 0); } }
    ],
    defaults: function () { return { type: 'reward', isActive: true, isPromo: false, perUserLimit: 1, maxTotalUses: null, code: '' }; },
    fixPayload: function (p) { if (p.code === '') p.code = null; return p; }
  },
  categories: {
    filters: [{ name: 'kind', label: 'filter.kind', options: opts('kind.', ['product', 'prize', 'gift', 'currency']) }],
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '55%' },
      { label: 'col.kind', render: function (it) { return tag(t('kind.' + it.kind), 'secondary'); }, width: 130 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
        { name: 'kind', type: 'select', label: 'f.kind', required: true, options: opts('kind.', ['product', 'prize', 'gift', 'currency']) }
      ] },
      nameRow('name'),
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ],
    defaults: function (query) { return { kind: (query && query.kind) || 'product', name: { fa: '', en: '' }, sort: 0 }; }
  },
  tags: {
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '55%' },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
      nameRow('name')
    ],
    defaults: function () { return { name: { fa: '', en: '' } }; }
  }
};
