ICONS['avatar'] = '<circle cx="12" cy="8" r="4"/><path d="M4 21c0-4 4-6 8-6s8 2 8 6"/>';
PH_ICONS['avatar'] = '<path d="M230.92,212c-15.23-26.33-38.7-45.21-66.09-54.16a72,72,0,1,0-73.66,0C63.78,166.78,40.31,185.66,25.08,212a8,8,0,1,0,13.85,8c18.84-32.56,52.14-52,89.07-52s70.23,19.44,89.07,52a8,8,0,1,0,13.85-8ZM72,96a56,56,0,1,1,56,56A56.06,56.06,0,0,1,72,96Z"/>';
['avatar-categories', 'avatar-items', 'avatar-types'].forEach(function (k) { ICONS[k] = ICONS['avatar']; PH_ICONS[k] = PH_ICONS['avatar']; });
ICONS['avatar-settings'] = '<path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6"/>';

var AVD_ACCESS = ['free', 'currency', 'product', 'item'];
var AVD_ACCESS_COLOR = { free: 'success', currency: 'primary', product: 'warning', item: 'secondary' };

function avdSwatch(o) {
  var cell = h('div', { class: 'h-8 w-8 shrink-0 rounded-md bg-secondary-100 dark:bg-secondary-700 overflow-hidden flex items-center justify-center', title: o.key || '' });
  if (o.iconUrl || o.imageUrl) cell.appendChild(img(o.iconUrl || o.imageUrl));
  else if (o.value && /^#[0-9a-f]{3,8}$/i.test(o.value)) cell.style.background = o.value;
  return cell;
}

function avdPreview(it) {
  var box = h('div', { class: 'flex flex-wrap items-center gap-1', 'data-avd-preview': '' });
  (it.preview || []).forEach(function (o) { box.appendChild(avdSwatch(o)); });
  if ((it.count || 0) > 6) box.appendChild(h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: '+' + fmtNum(it.count - 6) }));
  return box;
}

function avdGetCell(it) {
  return h('div', { class: 'flex flex-col items-start gap-1 min-w-0 text-sm', 'data-avd-access': it.access }, withKeyNames(function () {
    var out = [tag(t('avd.access.' + it.access), AVD_ACCESS_COLOR[it.access] || 'secondary')];
    if (it.access === 'currency' && it.price) out.push(h('span', { class: 'tabular-nums whitespace-nowrap', text: priceLabel(it.price) }));
    if (it.access === 'product') out.push(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 truncate', text: it.productKey || '' }));
    if (it.access === 'item') out.push(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 truncate', text: it.itemKey || '' }));
    return out;
  }));
}

RESOURCES['avatar-categories'] = {
  filters: [activeFilter()],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '35%' },
    { label: 'col.avd.items', render: function (it) { return h('div', { class: 'flex flex-col gap-1' }, avdPreview(it), h('span', { class: 'text-xs text-secondary-500', text: t('avd.itemsN', { n: fmtNum(it.count || 0) }) })); }, skeleton: '30%' },
    { label: 'col.avd.layer', render: function (it) {
      return h('div', { class: 'flex flex-col', 'data-avd-layer': '' }, numCell(it.layer),
        h('span', { class: 'text-xs text-secondary-500', text: t('avd.slotsN', { n: fmtNum(it.slots || 1) }) + (it.required ? ' · ' + t('avd.required') : '') }));
    }, width: 130 },
    { label: 'col.avd.avatars', render: function (it) { return numCell(it.avatars || 0); }, width: 100 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.avd.key.hint', required: true, lockOnEdit: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    tr('name', { label: 'f.name', required: true, max: 100 }),
    { type: 'row', fields: [
      { name: 'layer', type: 'number', label: 'f.avd.layer', hint: 'f.avd.layer.hint', min: -1000, max: 1000, empty: 0 },
      { name: 'slots', type: 'number', label: 'f.avd.slots', hint: 'f.avd.slots.hint', min: 1, max: 20, empty: 1 }
    ] },
    { name: 'required', type: 'bool', inline: true, label: 'f.avd.required', hint: 'f.avd.required.hint' },
    { type: 'row', fields: [
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint', min: -100000, max: 100000, empty: 0 }
    ] }
  ],
  defaults: function () { return { key: '', name: {}, layer: 0, slots: 1, sort: 0, required: false, isActive: true, iconUrl: '' }; },
  fixPayload: function (p) {
    p.layer = p.layer || 0;
    p.sort = p.sort || 0;
    p.slots = p.slots || 1;
    p.iconUrl = p.iconUrl || '';
    ['id', 'count', 'avatars', 'preview', 'createdAt', 'createdBy', 'updatedAt'].forEach(function (k) { delete p[k]; });
    return p;
  },
  afterSave: function () { forgetRefs('avatar-categories'); }
};

RESOURCES['avatar-items'] = {
  filters: [
    { name: 'categoryId', label: 'filter.avd.category', type: 'ref', resource: 'avatar-categories' },
    { name: 'access', label: 'filter.avd.access', options: opts('avd.access.', AVD_ACCESS) },
    { name: 'typeId', label: 'filter.avd.type', type: 'ref', resource: 'avatar-types' },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) {
      return h('div', { class: 'flex items-center gap-3 min-w-0' }, avdSwatch(it), h('div', { class: 'min-w-0' },
        h('div', { class: 'font-medium truncate', text: i18nText(it.name) || it.key }),
        h('div', { class: 'text-secondary-500' }, mono((it.categoryKey || '') + ' / ' + it.key))));
    }, skeleton: '40%' },
    { label: 'col.avd.get', render: avdGetCell, width: 170 },
    { label: 'col.avd.cond', render: function (it) {
      return h('div', { class: 'flex flex-col gap-1 text-xs text-secondary-600 dark:text-secondary-300 min-w-0', 'data-avd-cond': '' },
        h('span', { text: reqEmpty(it.requirements) ? t('avd.cond.none') : reqSummary(it.requirements, 'avatar') }),
        h('span', { text: (it.types || []).length ? (it.types || []).map(function (y) { return i18nText(y.name) || y.key; }).join('، ') : t('avd.types.all') }));
    }, width: 190 },
    { label: 'col.avd.owners', render: function (it) {
      return h('div', { class: 'flex flex-col text-sm' }, h('span', { class: 'tabular-nums', text: t('avd.ownersN', { n: fmtNum(it.owners || 0) }) }),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('avd.equippedN', { n: fmtNum(it.equipped || 0) }) }));
    }, width: 130 },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive), it.isDefault ? tag(t('avd.default'), 'primary') : null); }, width: 110 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'categoryId', type: 'ref', resource: 'avatar-categories', label: 'f.avd.category', hint: 'f.avd.category.hint', required: true, lockOnEdit: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.avd.key.hint', required: true, lockOnEdit: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint', min: -100000, max: 100000, empty: 0 }
    ] },
    tr('name', { label: 'f.name', max: 100, optional: true }),
    { type: 'heading', label: 'h.avd.look' },
    { type: 'row', fields: [
      { name: 'imageUrl', type: 'image', label: 'f.avd.image', hint: 'f.avd.image.hint', optional: true },
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true }
    ] },
    { name: 'value', type: 'text', ltr: true, label: 'f.avd.value', hint: 'f.avd.value.hint', optional: true, max: 64,
      check: function (v, vals) { return !v && !vals.imageUrl ? t('code.OPTION_LOOK_REQUIRED') : null; } },
    { type: 'heading', label: 'h.avd.get' },
    { name: 'access', type: 'select', noPlaceholder: true, label: 'f.avd.access', hint: 'f.avd.access.hint', options: opts('avd.access.', AVD_ACCESS) },
    { type: 'row', when: function (v) { return v.access === 'currency'; }, fields: [
      { name: 'price.currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.it.price.currency', required: true },
      { name: 'price.amount', type: 'number', label: 'f.it.price.amount', required: true, min: 1 }
    ] },
    { name: 'productId', type: 'ref', resource: 'products', label: 'f.avd.product', hint: 'f.avd.product.hint', required: true, when: function (v) { return v.access === 'product'; } },
    { name: 'itemKey', type: 'ref', resource: 'items', valueKey: 'key', label: 'f.avd.item', hint: 'f.avd.item.hint', required: true, when: function (v) { return v.access === 'item'; } },
    { name: 'isDefault', type: 'bool', inline: true, label: 'f.avd.default', hint: 'f.avd.default.hint', when: function (v) { return v.access === 'free'; } },
    { type: 'heading', label: 'h.avd.who' },
    { name: 'typeIds', type: 'refMulti', resource: 'avatar-types', label: 'f.avd.types', hint: 'f.avd.types.hint', optional: true, empty: [] },
    { name: 'requirements', type: 'custom', label: 'f.avd.req', hint: 'f.avd.req.hint', build: requirementsEditor('avatar') }
  ],
  defaults: function (query) {
    return { categoryId: (query && query.categoryId) || null, key: '', name: {}, imageUrl: '', iconUrl: '', value: '', access: 'free', price: null, productId: null, itemKey: null,
      typeIds: [], requirements: null, isDefault: false, sort: 0, isActive: true };
  },
  fixPayload: function (p) {
    p.value = toLatinDigits(String(p.value || '')).trim();
    p.imageUrl = p.imageUrl || '';
    p.iconUrl = p.iconUrl || '';
    p.sort = p.sort || 0;
    if (p.access !== 'currency') p.price = null;
    p.productId = p.access === 'product' ? (p.productId || '') : '';
    p.itemKey = p.access === 'item' ? (p.itemKey || '') : '';
    if (p.access !== 'free') p.isDefault = false;
    if (p.requirements && reqEmpty(p.requirements)) p.requirements = null;
    ['id', 'categoryKey', 'categoryName', 'productKey', 'types', 'owners', 'equipped', 'createdAt', 'updatedAt'].forEach(function (k) { delete p[k]; });
    return p;
  }
};

RESOURCES['avatar-types'] = {
  filters: [activeFilter()],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '45%' },
    { label: 'col.avd.avatars', render: function (it) { return numCell(it.avatars || 0); }, width: 110 },
    { label: 'col.avd.typeItems', render: function (it) { return numCell(it.items || 0); }, width: 130 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.avd.typeKey.hint', required: true, lockOnEdit: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    tr('name', { label: 'f.name', required: true, max: 100 }),
    { type: 'row', fields: [
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { name: 'imageUrl', type: 'image', label: 'f.avd.base', hint: 'f.avd.base.hint', optional: true }
    ] },
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.avd.typeSort.hint', min: -100000, max: 100000, empty: 0 }
  ],
  defaults: function () { return { key: '', name: {}, iconUrl: '', imageUrl: '', sort: 0, isActive: true }; },
  fixPayload: function (p) {
    p.iconUrl = p.iconUrl || '';
    p.imageUrl = p.imageUrl || '';
    p.sort = p.sort || 0;
    ['id', 'avatars', 'items', 'createdAt', 'updatedAt'].forEach(function (k) { delete p[k]; });
    return p;
  },
  afterSave: function () { forgetRefs('avatar-types'); }
};

var AVD_SETTINGS_FIELDS = [
  { name: 'enabled', type: 'bool', inline: true, label: 'f.avd.enabled', hint: 'f.avd.enabled.hint' },
  { type: 'heading', label: 'h.avd.count' },
  { type: 'row', fields: [
    { name: 'maxAvatars', type: 'number', label: 'f.avd.max', hint: 'f.avd.max.hint', required: true, min: 1, max: 50 },
    { name: 'nameMaxLength', type: 'number', label: 'f.avd.nameMax', hint: 'f.avd.nameMax.hint', required: true, min: 1, max: 64 }
  ] },
  { name: 'tagLimits', type: 'rows', label: 'f.avd.tagLimits', hint: 'f.avd.tagLimits.hint', addLabel: 'avd.tag.add', emptyText: 'avd.tag.empty', max: 20, empty: [],
    newRow: function () { return { tagId: null, maxAvatars: 3 }; },
    check: duplicateCheck('tagId', 'DUPLICATE_TAG'),
    columns: [
      { name: 'tagId', type: 'ref', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, label: 'f.avd.tag', required: true },
      { name: 'maxAvatars', type: 'number', label: 'f.avd.max', required: true, min: 1, max: 50 }
    ] },
  { name: 'autoCreate', type: 'bool', inline: true, label: 'f.avd.autoCreate', hint: 'f.avd.autoCreate.hint' },
  { name: 'keepLast', type: 'bool', inline: true, label: 'f.avd.keepLast', hint: 'f.avd.keepLast.hint' },
  { type: 'heading', label: 'h.avd.rules' },
  { name: 'slotFull', type: 'select', noPlaceholder: true, label: 'f.avd.slotFull', hint: 'f.avd.slotFull.hint', options: opts('avd.slotFull.', ['replace', 'refuse']) },
  { name: 'keepBought', type: 'bool', inline: true, label: 'f.avd.keepBought', hint: 'f.avd.keepBought.hint' }
];

function avatarSettingsPage(container) {
  var frame = pageFrame(t('nav.avatar-settings'), t('desc.avatar-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(api, '/avatar-settings', null, AVD_SETTINGS_FIELDS, {
    id: 'avatar',
    fixPayload: function (p) {
      delete p.updatedAt;
      p.tagLimits = (p.tagLimits || []).map(function (r) { return { tagId: r.tagId, maxAvatars: r.maxAvatars }; });
      return p;
    }
  });
  frame.content.appendChild(page.el);
  page.load();
}

CUSTOM_PAGES['avatar-settings'] = function (root) { avatarSettingsPage(root); };
