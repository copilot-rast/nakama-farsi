var METHODS = ['zarinpal', 'zibal', 'bazaar', 'myket', 'in_app'];
var GATEWAYS = ['zarinpal', 'zibal'];
var STORES = ['bazaar', 'myket'];
var CATEGORY_KEY_TYPES = ['product', 'prize', 'gift', 'currency', 'faq', 'news', 'task'];
var TAG_KEY_TYPES = ['product', 'prize', 'gift', 'currency', 'faq', 'news', 'task', 'general', 'user'];
var LANG_CODE_RE = /^[a-z]{2,3}(-[a-z0-9]{2,8})?$/;
var keyTypeCache = {};
var keyTypePromise = null;

function loadKeyTypes() {
  if (keyTypePromise) return keyTypePromise;
  keyTypePromise = Promise.all(['categories', 'tags'].map(function (r) {
    return baseApi('GET', '/' + r + '/key-types').then(function (res) { if (res && res.items && res.items.length) keyTypeCache[r] = res.items; }, function () { return null; });
  })).then(function () { return keyTypeCache; }, function () { keyTypePromise = null; return keyTypeCache; });
  return keyTypePromise;
}

function keyTypeLabel(resource, key) {
  var hit = (keyTypeCache[resource] || []).filter(function (k) { return k.key === key; })[0];
  return hit && hit.name && i18nText(hit.name) ? i18nText(hit.name) : has('kind.' + key) ? t('kind.' + key) : key;
}

function keyTypeOptions(resource) {
  return function () {
    var list = keyTypeCache[resource] ? keyTypeCache[resource].map(function (k) { return k.key; }) : resource === 'tags' ? TAG_KEY_TYPES : CATEGORY_KEY_TYPES;
    return list.map(function (k) { return { value: k, label: keyTypeLabel(resource, k) }; });
  };
}

function categoryField(keyType) {
  return { name: 'categoryId', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: keyType }, quick: keyType, label: 'f.category', optional: true };
}

function tagsField(keyType) {
  return { name: 'tagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: [{ keyType: keyType }, { keyType: 'general' }], quick: keyType, label: 'f.tags', optional: true, empty: [] };
}

function tr(name, opts) {
  return Object.assign({ name: name, type: 'i18n', label: 'f.' + name }, opts || {});
}
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

var taxonomyNames = { categories: {}, tags: {} };
var taxonomyPending = { categories: {}, tags: {} };

function taxonomyLookup(kind, id) {
  if (taxonomyNames[kind][id]) return Promise.resolve(taxonomyNames[kind][id]);
  if (!taxonomyPending[kind][id]) {
    taxonomyPending[kind][id] = baseApi('GET', '/' + kind + '/' + encodeURIComponent(id)).then(function (res) {
      taxonomyNames[kind][id] = res;
      return res;
    }, function () { delete taxonomyPending[kind][id]; return null; });
  }
  return taxonomyPending[kind][id];
}

function taxonomyCell(it) {
  var box = h('div', { class: 'flex flex-wrap gap-1', 'data-taxonomy': '' });
  var ids = (it.tagIds || []).slice(0, 3);
  function paint() {
    clear(box);
    var cat = it.categoryId ? taxonomyNames.categories[it.categoryId] : null;
    if (cat) box.appendChild(tag(i18nText(cat.name) || cat.key, 'secondary'));
    ids.forEach(function (id) {
      var tg = taxonomyNames.tags[id];
      if (tg) box.appendChild(tag(i18nText(tg.name) || tg.key, 'primary'));
    });
    if ((it.tagIds || []).length > 3) box.appendChild(tag('+' + fmtNum(it.tagIds.length - 3), 'secondary'));
    if (!box.childNodes.length) box.appendChild(h('span', { class: 'text-secondary-400', text: t('none') }));
  }
  var jobs = [];
  if (it.categoryId && !taxonomyNames.categories[it.categoryId]) jobs.push(taxonomyLookup('categories', it.categoryId));
  ids.forEach(function (id) { if (!taxonomyNames.tags[id]) jobs.push(taxonomyLookup('tags', id)); });
  if (jobs.length) {
    box.appendChild(bar('4rem', 'h-4'));
    Promise.all(jobs).then(paint);
  } else paint();
  return box;
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

var KEY_NAMES = { data: null, promise: null, at: 0 };

function keyNames() {
  if (KEY_NAMES.data && Date.now() - KEY_NAMES.at < 30000) return Promise.resolve(KEY_NAMES.data);
  if (KEY_NAMES.promise) return KEY_NAMES.promise;
  var pick = function (res) {
    var out = {};
    ((res && res.items) || []).forEach(function (it) { out[it.key] = it.name; });
    return out;
  };
  KEY_NAMES.promise = Promise.all([
    api('GET', '/currencies?limit=100').then(pick, function () { return {}; }),
    api('GET', '/metas?limit=100').then(pick, function () { return {}; })
  ]).then(function (r) {
    KEY_NAMES.data = { currency: r[0], meta: r[1] };
    KEY_NAMES.at = Date.now();
    KEY_NAMES.promise = null;
    return KEY_NAMES.data;
  });
  return KEY_NAMES.promise;
}

function keyName(kind, key) {
  var names = KEY_NAMES.data && KEY_NAMES.data[kind];
  var name = names && names[key] ? i18nText(names[key]) : '';
  return name || key || '';
}

function withKeyNames(build) {
  var box = h('div', { class: 'contents' });
  var paint = function () { clear(box); append(box, build()); };
  paint();
  if (!KEY_NAMES.data || Date.now() - KEY_NAMES.at >= 30000) keyNames().then(function () { if (box.isConnected || box.parentNode) paint(); });
  return box;
}

function priceText(it) {
  var parts = [];
  if (it.priceToman !== null && it.priceToman !== undefined) parts.push(t('toman', { n: it.priceToman }));
  if (it.coinPrice && it.coinPrice.amount !== undefined && it.coinPrice.amount !== null) parts.push(t('amount.of', { n: fmtNum(it.coinPrice.amount), name: keyName('currency', it.coinPrice.currencyKey) }));
  return parts;
}

function itemSummary(items) {
  return (items || []).map(function (r) {
    if (r.type === 'storage') return r.key;
    if (r.type === 'item') return t('item.item') + ': ' + r.key;
    if (r.type === 'achievement') return t('item.achievement') + ': ' + r.key;
    if (r.type === 'tag') return t('item.tag') + ': ' + r.key + ' · ' + (Number(r.value) > 0 ? t('item.tag.forDays', { n: fmtNum(r.value) }) : t('item.tag.forever'));
    var sign = r.action === 'add' && Number(r.value) > 0 ? '+' : '';
    return t('amount.of', { n: (r.action === 'set' ? '= ' : sign) + fmtNum(r.value), name: keyName(r.type === 'meta' ? 'meta' : 'currency', r.key) });
  });
}

function languageCell(it) {
  var pic = h('div', { class: 'h-8 w-8 shrink-0 rounded-md bg-secondary-100 dark:bg-secondary-700 overflow-hidden flex items-center justify-center text-secondary-400' });
  pic.appendChild(it.iconUrl ? img(it.iconUrl, 'languages') : icon('languages', 'h-4 w-4'));
  var trName = i18nText(it.name);
  return h('div', { class: 'flex items-center gap-3 min-w-0' }, pic,
    h('div', { class: 'min-w-0' },
      h('div', { class: 'font-medium truncate flex items-center gap-2' }, h('span', { dir: it.direction === 'rtl' ? 'rtl' : 'ltr', text: it.nativeName || it.code }), it.isDefault ? tag(t('lang.default'), 'primary') : null),
      h('div', { class: 'text-secondary-500 flex items-center gap-2' }, mono(it.code), trName && trName !== it.nativeName ? h('span', { class: 'text-xs truncate', text: trName }) : null)));
}

function languageBody(it, patch) {
  return Object.assign({ name: it.name || {}, nativeName: it.nativeName, direction: it.direction, iconUrl: it.iconUrl || '', isActive: !!it.isActive, isDefault: !!it.isDefault, sort: it.sort || 0 }, patch || {});
}

function languageActive(it) {
  var box = h('div', { class: 'flex items-center gap-2' });
  var sw = toggle(it.isActive, null, function (on) {
    baseApi('PUT', '/languages/' + encodeURIComponent(it.code), languageBody(it, { isActive: on })).then(function () {
      it.isActive = on;
      label.textContent = on ? t('active') : t('inactive');
      toast('success', t(on ? 'lang.activated' : 'lang.deactivated', { name: it.nativeName || it.code }));
      languagesChanged();
    }, function (err) {
      sw.set(!on);
      toast('error', errorText(err));
    });
  }, 'sm');
  sw.setAttribute('aria-label', t('f.isActive'));
  sw.setAttribute('title', t('f.isActive'));
  var label = h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: it.isActive ? t('active') : t('inactive') });
  box.appendChild(sw);
  box.appendChild(label);
  return box;
}

var languageHooks = [];
function languagesChanged() {
  loadLanguages(true).then(function () { languageHooks.forEach(function (fn) { fn(); }); });
}

var RESOURCES = {
  currencies: {
    filters: [{ name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'currency' } }],
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '60%' },
      { label: 'col.type', render: function (it) { return tag(t('type.' + it.type), it.type === 'hard' ? 'warning' : 'primary'); }, width: 110 },
      { label: 'col.category', render: taxonomyCell, skeleton: '30%' },
      { label: 'col.default', render: function (it) { return numCell(it.defaultValue); }, width: 130 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, keyRe: CURRENCY_KEY_RE, lockOnEdit: true },
        { name: 'type', type: 'select', label: 'f.type', required: true, options: opts('type.', ['soft', 'hard']) }
      ] },
      tr('name', { required: true, max: 128 }),
      { type: 'row', fields: [categoryField('currency'), tagsField('currency')] },
      { type: 'row', fields: [
        { name: 'defaultValue', type: 'number', label: 'f.defaultValue', hint: 'f.defaultValue.currency.hint', min: 0 },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function () { return { type: 'soft', name: {}, categoryId: null, tagIds: [], defaultValue: 0, sort: 0, metadata: {} }; }
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
      tr('name', { required: true, max: 128 }),
      { name: 'defaultValue', type: 'number', int: false, label: 'f.defaultValue', hint: 'f.defaultValue.meta.hint' },
      { name: 'serverOnly', type: 'bool', inline: true, label: 'f.serverOnly', hint: 'f.serverOnly.hint' },
      { name: 'settings', type: 'json', label: 'f.settings', optional: true, empty: {} }
    ],
    defaults: function () { return { name: {}, defaultValue: 0, serverOnly: false, settings: {} }; }
  },
  prizes: {
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it); }, skeleton: '55%' },
      { label: 'col.items', render: function (it) {
        return h('div', { class: 'flex flex-wrap gap-1' }, withKeyNames(function () {
          var s = itemSummary(it.items);
          return [s.slice(0, 3).map(function (x) { return tag(x, 'secondary'); }), s.length > 3 ? tag('+' + fmtNum(s.length - 3), 'secondary') : null];
        }));
      }, skeleton: '40%' },
      { label: 'col.type', render: function (it) { return tag(it.isRepeatable ? t('repeatable') : t('once'), it.isRepeatable ? 'primary' : 'secondary'); }, width: 120 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
      tr('name', { required: true, max: 128 }),
      tr('message', { multiline: true, max: 1000, optional: true }),
      { type: 'row', fields: [categoryField('prize'), tagsField('prize')] },
      { name: 'isRepeatable', type: 'bool', inline: true, label: 'f.isRepeatable', hint: 'f.isRepeatable.hint' },
      { name: 'items', type: 'items', label: 'f.items' },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function () { return { name: {}, message: {}, tagIds: [], isRepeatable: false, items: [{ type: 'currency', key: '', action: 'add', value: 0 }], metadata: {} }; }
  },
  products: {
    filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '60%' },
      { label: 'col.price', render: function (it) { return h('div', { class: 'flex flex-col tabular-nums whitespace-nowrap' }, withKeyNames(function () { return priceText(it).map(function (p) { return h('span', { text: p }); }); }), it.discountPercent ? h('span', { class: 'text-xs text-success-600', text: t('percent', { n: it.discountPercent }) }) : null); }, width: 160 },
      { label: 'col.methods', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, (it.paymentMethods || []).map(function (m) { return tag(t('method.' + m), 'secondary'); }), (it.audienceTagIds || []).length ? tag(t('audience.only'), 'primary') : null); }, skeleton: '45%' },
      { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 80 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
        { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
      ] },
      tr('name', { required: true, max: 128 }),
      tr('description', { multiline: true, max: 4000, optional: true }),
      { type: 'row', fields: [categoryField('product'), tagsField('product')] },
      { name: 'audienceTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.audience', hint: 'f.audience.hint', optional: true, empty: [] },
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
    defaults: function () { return { isActive: true, name: {}, description: {}, tagIds: [], audienceTagIds: [], paymentMethods: ['zarinpal'], discountPercent: 0, sku: {}, coinPrice: null, isSubscription: false, minLevel: 0, sort: 0, metadata: {} }; },
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
    filters: [
      { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'gift' } },
      { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }
    ],
    columns: [
      { label: 'col.name', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) }), it.code ? h('div', { class: 'text-secondary-500' }, mono(it.code)) : null); }, skeleton: '50%' },
      { label: 'col.type', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('gift.' + it.type), it.type === 'discount' ? 'warning' : 'primary'), it.type === 'discount' && it.discountPercent ? tag(t('percent', { n: it.discountPercent }), 'secondary') : null); }, width: 150 },
      { label: 'col.category', render: taxonomyCell, skeleton: '30%' },
      { label: 'col.usage', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: fmtNum(it.totalUsed || 0) + ' / ' + (it.maxTotalUses ? fmtNum(it.maxTotalUses) : t('unlimited')) }); }, width: 130 },
      { label: 'col.expire', render: function (it) { return dateCell(it.expireAt); }, width: 190 },
      { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
    ],
    fields: [
      tr('name', { label: 'f.gift.name', required: true, max: 128 }),
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' },
      { type: 'row', fields: [categoryField('gift'), tagsField('gift')] },
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
    defaults: function (query) { return { name: {}, type: 'reward', isActive: true, isPromo: false, perUserLimit: 1, maxTotalUses: null, code: '', categoryId: (query && query.categoryId) || null, tagIds: [] }; },
    fixPayload: function (p) { if (p.code === '') p.code = null; return p; }
  },
  categories: {
    api: 'base',
    filters: [{ name: 'keyType', label: 'filter.keyType', options: keyTypeOptions('categories') }],
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '55%' },
      { label: 'col.keyType', render: function (it) { return tag(keyTypeLabel('categories', it.keyType), 'secondary'); }, width: 140 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
        { name: 'keyType', type: 'select', label: 'f.keyType', hint: 'f.keyType.category.hint', required: true, options: keyTypeOptions('categories') }
      ] },
      tr('name', { required: true, max: 128 }),
      tr('description', { multiline: true, max: 1000, optional: true }),
      { type: 'row', fields: [
        { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function (query) { return { keyType: (query && query.keyType) || 'product', name: {}, description: {}, iconUrl: '', metadata: {}, sort: 0 }; },
    fixPayload: function (p) { if (!p.iconUrl) p.iconUrl = ''; if (p.sort == null) p.sort = 0; return p; }
  },
  tags: {
    api: 'base',
    filters: [{ name: 'keyType', label: 'filter.keyType', options: keyTypeOptions('tags') }],
    columns: [
      { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '55%' },
      { label: 'col.keyType', render: function (it) {
        return h('div', { class: 'flex flex-wrap gap-1' }, tag(keyTypeLabel('tags', it.keyType), it.keyType === 'general' ? 'primary' : 'secondary'),
          it.keyType === 'user' && it.isPublic ? tag(t('tag.public'), 'success') : null,
          it.keyType === 'user' && it.players ? h('a', { href: hashFor('players', null, { tagId: it.id }), class: 'eco-link text-xs whitespace-nowrap', text: t('tag.playersN', { n: it.players }) }) : null);
      }, width: 200 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true },
        { name: 'keyType', type: 'select', label: 'f.keyType', hint: 'f.keyType.tag.hint', required: true, options: keyTypeOptions('tags') }
      ] },
      tr('name', { required: true, max: 128 }),
      tr('description', { multiline: true, max: 1000, optional: true }),
      { name: 'isPublic', type: 'bool', inline: true, label: 'f.tag.public', hint: 'f.tag.public.hint', empty: false, when: function (v) { return v.keyType === 'user'; } },
      { type: 'row', fields: [
        { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
    ],
    defaults: function (query) { return { isPublic: false, keyType: (query && query.keyType) || 'general', name: {}, description: {}, iconUrl: '', metadata: {}, sort: 0 }; },
    fixPayload: function (p) { if (!p.iconUrl) p.iconUrl = ''; if (p.sort == null) p.sort = 0; return p; }
  },
  languages: {
    api: 'base',
    filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
    columns: [
      { label: 'col.language', render: languageCell, skeleton: '60%' },
      { label: 'col.status', render: languageActive, width: 150 },
      { label: 'col.direction', render: function (it) { return tag(t('dir.' + it.direction), 'secondary'); }, width: 150 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 }
    ],
    fields: [
      { type: 'row', fields: [
        { name: 'code', type: 'key', label: 'f.lang.code', hint: 'f.lang.code.hint', required: true, keyRe: LANG_CODE_RE, keyErr: 'code.INVALID_LANGUAGE_CODE', lockOnEdit: true },
        { name: 'nativeName', label: 'f.lang.nativeName', hint: 'f.lang.nativeName.hint', required: true, check: function (v) { return v && Array.from(v.trim()).length > 64 ? t('code.INVALID_NATIVE_NAME') : null; } }
      ] },
      tr('name', { label: 'f.lang.name', hint: 'f.lang.name.hint', required: true, max: 64 }),
      { type: 'row', fields: [
        { name: 'direction', type: 'select', label: 'f.lang.direction', required: true, options: opts('dir.', ['rtl', 'ltr']) },
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { type: 'row', fields: [
        { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.lang.active.hint' },
        { name: 'isDefault', type: 'bool', inline: true, label: 'f.lang.default', hint: 'f.lang.default.hint' }
      ] }
    ],
    defaults: function () { return { code: '', name: {}, nativeName: '', direction: 'ltr', iconUrl: '', isActive: true, isDefault: false, sort: 0 }; },
    fixPayload: function (p) { p.code = String(p.code || '').trim().toLowerCase().replace(/_/g, '-'); if (!p.iconUrl) p.iconUrl = ''; if (p.sort == null) p.sort = 0; return p; },
    afterSave: function () { languagesChanged(); }
  },
  faqs: {
    api: 'publication',
    filters: [
      { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'faq' } },
      { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }
    ],
    columns: [
      { label: 'col.question', render: function (it) {
        return h('div', { class: 'min-w-0' },
          h('div', { class: 'font-medium eco-clamp', text: i18nText(it.question) }),
          h('div', { class: 'text-xs text-secondary-500 eco-clamp', text: mdPlain(i18nText(it.answer)) }));
      }, skeleton: '70%' },
      { label: 'col.category', render: function (it) { return it.categoryName ? tag(i18nText(it.categoryName), 'secondary') : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 150 },
      { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 },
      { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 80 },
      { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 190 }
    ],
    fields: [
      tr('question', { required: true, max: 500 }),
      tr('answer', { markdown: true, required: true, max: 10000 }),
      { type: 'row', fields: [
        categoryField('faq'),
        { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
      ] },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.faq.active.hint' }
    ],
    defaults: function (query) { return { question: {}, answer: {}, categoryId: (query && query.categoryId) || null, isActive: true, sort: 0 }; }
  }
};
