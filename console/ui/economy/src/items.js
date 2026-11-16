var RARITIES = ['common', 'uncommon', 'rare', 'epic', 'legendary'];
var RARITY_COLOR = { common: 'secondary', uncommon: 'success', rare: 'primary', epic: 'warning', legendary: 'error' };

if (ACTIVITY_KINDS.indexOf('item') === -1) ACTIVITY_KINDS.push('item');
if (CATEGORY_KEY_TYPES.indexOf('item') === -1) CATEGORY_KEY_TYPES.push('item', 'item_type');
if (TAG_KEY_TYPES.indexOf('item') === -1) TAG_KEY_TYPES.splice(TAG_KEY_TYPES.indexOf('general'), 0, 'item');

function rarityTag(r) {
  return tag(t('rarity.' + (r || 'common')), RARITY_COLOR[r] || 'secondary');
}

function priceLabel(p) {
  if (!p || p.amount == null) return '';
  return t('amount.of', { n: fmtNum(p.amount), name: keyName('currency', p.currencyKey) });
}

function itemGetCell(it) {
  return h('div', { class: 'flex flex-col items-start gap-1 min-w-0 text-sm' }, withKeyNames(function () {
    var out = [];
    if (it.isStarter) out.push(tag(t('it.get.starter'), 'success'));
    if (it.price) out.push(h('span', { class: 'tabular-nums whitespace-nowrap', text: priceLabel(it.price) }));
    if (it.unlockLevel) out.push(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('it.get.level', { n: fmtNum(it.unlockLevel) }) }));
    if (!it.isStarter && !it.price && !it.unlockLevel) out.push(h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: t('it.get.prize') }));
    if (it.requiresKey || it.parentKey) out.push(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 truncate', text: t('it.get.requires', { key: it.requiresKey || it.parentKey }) }));
    if ((it.upgrades || []).length) out.push(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('it.upgradesN', { n: fmtNum(it.upgrades.length) }) }));
    return out;
  }));
}

RESOURCES.items = {
  filters: [
    { name: 'typeId', label: 'filter.it.type', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'item_type' } },
    { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'item' } },
    { name: 'rarity', label: 'filter.it.rarity', options: opts('rarity.', RARITIES) },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, true); }, skeleton: '45%' },
    { label: 'col.it.type', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, tag(i18nText(it.typeName) || it.typeKey, 'secondary'), rarityTag(it.rarity)); }, width: 170 },
    { label: 'col.it.get', render: itemGetCell, skeleton: '40%', width: 190 },
    { label: 'col.it.owners', render: function (it) {
      return h('div', { class: 'flex flex-col text-sm' }, h('span', { class: 'tabular-nums', text: t('it.ownersN', { n: fmtNum(it.owners || 0) }) }),
        it.equippedBy ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('it.equippedN', { n: fmtNum(it.equippedBy) }) }) : null);
    }, width: 150 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    tr('name', { required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { type: 'row', fields: [
      { name: 'typeId', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'item_type' }, quick: 'item_type', label: 'f.it.type', hint: 'f.it.type.hint', required: true },
      { name: 'rarity', type: 'select', label: 'f.it.rarity', required: true, options: opts('rarity.', RARITIES) }
    ] },
    { type: 'row', fields: [categoryField('item'), tagsField('item')] },
    { type: 'heading', label: 'h.it.get' },
    { name: 'isStarter', type: 'bool', inline: true, label: 'f.it.starter', hint: 'f.it.starter.hint' },
    { type: 'row', when: function (v) { return !v.isStarter; }, fields: [
      { name: 'price.currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.it.price.currency', optional: true,
        check: function (c, v) { return !c && v.price && v.price.amount != null ? t('err.required') : null; } },
      { name: 'price.amount', type: 'number', label: 'f.it.price.amount', hint: 'f.it.price.hint', optional: true, min: 1,
        check: function (n, v) { return n == null && v.price && v.price.currencyKey ? t('err.required') : null; } }
    ] },
    { type: 'row', when: function (v) { return !v.isStarter; }, fields: [
      { name: 'unlockLevel', type: 'number', label: 'f.it.unlockLevel', hint: 'f.it.unlockLevel.hint', optional: true, min: 1 },
      { name: 'levelTableId', type: 'ref', resource: 'level-tables', label: 'f.it.levelTable', hint: 'f.it.levelTable.hint', optional: true }
    ] },
    { type: 'row', fields: [
      { name: 'requiresKey', type: 'ref', resource: 'items', valueKey: 'key', label: 'f.it.requires', hint: 'f.it.requires.hint', optional: true,
        check: function (v, vals) { return v && v === vals.key ? t('code.SELF_REFERENCE') : null; } },
      { name: 'parentKey', type: 'ref', resource: 'items', valueKey: 'key', label: 'f.it.parent', optional: true,
        check: function (v, vals) { return v && v === vals.key ? t('code.SELF_REFERENCE') : null; } }
    ] },
    { type: 'heading', label: 'h.it.upgrades' },
    { name: 'upgrades', type: 'rows', addLabel: 'it.up.add', emptyText: 'it.up.empty', max: 50,
      newRow: function (rows) { var last = rows[rows.length - 1]; return { currencyKey: last ? last.currencyKey : null, amount: last ? Number(last.amount || 0) * 2 || 10 : 10 }; },
      columns: [
        { name: 'currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.it.up.currency', required: true },
        { name: 'amount', type: 'number', label: 'f.it.up.amount', required: true, min: 1 }
      ] },
    { type: 'heading', label: 'h.it.services' },
    { name: 'services', type: 'rows', addLabel: 'it.sv.add', emptyText: 'it.sv.empty', max: 20,
      newRow: function () { return { key: '', currencyKey: null, unitPrice: 1, maxUnits: 1, name: {} }; },
      check: duplicateCheck('key', 'DUPLICATE_SERVICE'),
      columns: [
        { name: 'key', type: 'text', ltr: true, mono: true, label: 'f.it.sv.key', required: true },
        { name: 'currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.it.sv.currency', required: true },
        { name: 'unitPrice', type: 'number', label: 'f.it.sv.price', required: true, min: 1 },
        { name: 'maxUnits', type: 'number', label: 'f.it.sv.max', required: true, min: 1, max: 1000 },
        { name: 'name', type: 'i18n', label: 'f.it.sv.name', max: 64, wide: true }
      ] },
    { type: 'heading', label: 'f.icon' },
    { type: 'row', fields: [
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { name: 'imageUrl', type: 'image', label: 'f.it.image', optional: true }
    ] },
    { type: 'row', fields: [
      { name: 'soundUrl', type: 'sound', label: 'f.it.sound', hint: 'f.it.sound.hint', optional: true }
    ] },
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' },
    { name: 'metadata', type: 'json', label: 'f.metadata', optional: true, empty: {} }
  ],
  defaults: function (query) {
    return { key: '', isActive: true, name: {}, description: {}, typeId: (query && query.typeId) || null, rarity: 'common', categoryId: null, tagIds: [], isStarter: false,
      price: null, unlockLevel: null, levelTableId: null, requiresKey: null, parentKey: null, upgrades: [], services: [], iconUrl: '', imageUrl: '', soundUrl: '', sort: 0, metadata: {} };
  },
  fixPayload: function (p) {
    if (p.isStarter || !p.price || (p.price.currencyKey == null && p.price.amount == null)) p.price = null;
    if (p.isStarter) { p.unlockLevel = null; p.levelTableId = null; }
    if (!p.iconUrl) p.iconUrl = '';
    if (!p.imageUrl) p.imageUrl = '';
    if (!p.soundUrl) p.soundUrl = '';
    if (p.sort == null) p.sort = 0;
    p.requiresKey = p.requiresKey || '';
    p.parentKey = p.parentKey || '';
    p.levelTableId = p.levelTableId || '';
    p.categoryId = p.categoryId || '';
    p.upgrades = (p.upgrades || []).map(function (r) { return { currencyKey: r.currencyKey, amount: r.amount }; });
    p.services = (p.services || []).map(function (r) { return { key: r.key, currencyKey: r.currencyKey, unitPrice: r.unitPrice, maxUnits: r.maxUnits, name: r.name || {} }; });
    return p;
  },
  afterSave: function () { forgetRefs('items'); }
};

(function productRules() {
  var def = RESOURCES.products;
  var at = -1;
  def.fields.forEach(function (f, i) { if (f.type === 'row' && f.fields.some(function (x) { return x.name === 'minLevel'; })) at = i; });
  var rules = [
    { type: 'heading', label: 'h.pr.rules' },
    { type: 'row', fields: [
      { name: 'minLevel', type: 'number', label: 'f.minLevel', min: 0, empty: 0 },
      { name: 'maxLevel', type: 'number', label: 'f.maxLevel', hint: 'f.maxLevel.hint', optional: true, min: 0,
        check: function (v, vals) { return v != null && Number(v) < Number(vals.minLevel || 0) ? t('code.INVALID_LEVEL_RANGE') : null; } }
    ] },
    { name: 'levelTableId', type: 'ref', resource: 'level-tables', label: 'f.pr.levelTable', hint: 'f.pr.levelTable.hint', optional: true },
    { type: 'row', fields: [
      { name: 'isOneTime', type: 'bool', inline: true, label: 'f.isOneTime', hint: 'f.isOneTime.hint', empty: false },
      { name: 'checkContents', type: 'bool', inline: true, label: 'f.checkContents', hint: 'f.checkContents.hint' }
    ] },
    tr('purchaseMessage', { multiline: true, max: 500, optional: true }),
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
  ];
  if (at !== -1) def.fields.splice(at, 1);
  var meta = -1;
  def.fields.forEach(function (f, i) { if (f.name === 'metadata') meta = i; });
  Array.prototype.splice.apply(def.fields, [meta === -1 ? def.fields.length : meta, 0].concat(rules));
  var baseDefaults = def.defaults;
  def.defaults = function (q) { return Object.assign(baseDefaults(q), { maxLevel: null, levelTableId: null, isOneTime: false, checkContents: true, purchaseMessage: {} }); };
  var baseFix = def.fixPayload;
  def.fixPayload = function (p) { p = baseFix(p); p.levelTableId = p.levelTableId || ''; if (p.isOneTime == null) p.isOneTime = false; if (p.checkContents == null) p.checkContents = true; return p; };
  def.columns.splice(3, 0, { label: 'h.pr.rules', render: function (it) {
    var box = h('div', { class: 'flex flex-wrap gap-1', 'data-product-rules': '' });
    if (it.isOneTime) box.appendChild(tag(t('pr.oneTime'), 'warning'));
    if (it.minLevel && it.maxLevel != null) box.appendChild(tag(t('pr.levels', { min: fmtNum(it.minLevel), max: fmtNum(it.maxLevel) }), 'secondary'));
    else if (it.minLevel) box.appendChild(tag(t('pr.levelFrom', { n: fmtNum(it.minLevel) }), 'secondary'));
    else if (it.maxLevel != null) box.appendChild(tag(t('pr.levelTo', { n: fmtNum(it.maxLevel) }), 'secondary'));
    if (!box.childNodes.length) box.appendChild(h('span', { class: 'text-secondary-400', text: t('none') }));
    return box;
  }, width: 150 });
})();

var INV_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadInvCard(id) {
  INV_CARD.loading = true;
  return api('GET', '/users/' + encodeURIComponent(id) + '/items' + qs({ lang: lang() })).then(function (res) {
    INV_CARD.loading = false;
    INV_CARD.at = Date.now();
    INV_CARD.data = res;
    INV_CARD.err = null;
    var old = document.querySelector('[data-player-items]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    INV_CARD.loading = false;
    INV_CARD.at = Date.now();
    INV_CARD.err = err;
    schedule();
  });
}

function reloadInv(id) { INV_CARD.at = 0; loadInvCard(id); }

function invLine(userId, it, write) {
  var pic = h('div', { class: 'h-8 w-8 shrink-0 rounded-md bg-secondary-100 dark:bg-secondary-700 overflow-hidden flex items-center justify-center text-secondary-500' });
  pic.appendChild(it.iconUrl ? img(it.iconUrl, 'items') : icon('items', 'h-4 w-4'));
  var line = h('div', { class: 'flex flex-wrap items-center gap-2 py-1.5 text-sm', 'data-inv-item': it.key, 'data-equipped': it.equipped ? '' : null },
    pic,
    h('span', { class: 'font-medium truncate min-w-0', text: it.name || it.key }),
    tag(it.typeName || it.type, 'secondary'),
    rarityTag(it.rarity),
    it.equipped ? tag(t('inv.equipped'), 'success') : null,
    !it.isActive ? tag(t('inv.inactive'), 'secondary') : null,
    h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 whitespace-nowrap', text: t('inv.source.' + it.source) }),
    it.maxUpgradeLevel ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 whitespace-nowrap tabular-nums', text: t('inv.upgrade', { n: fmtNum(it.upgradeLevel || 0), max: fmtNum(it.maxUpgradeLevel) }) }) : null,
    h('span', { class: 'flex-1' }));
  if (write && !it.isStarter) {
    var rb = button({ variant: 'link', size: 'sm', label: t('inv.revoke'), onclick: function () {
      confirmModal({ title: t('inv.revoke.title'), text: t('inv.revoke.text', { name: it.name || it.key }), ok: t('inv.revoke'), danger: true }).then(function (yes) {
        if (!yes) return;
        api('DELETE', '/users/' + encodeURIComponent(userId) + '/items/' + encodeURIComponent(it.id)).then(function () { toast('success', t('inv.revoke.done')); reloadInv(userId); }, function (err) { toast('error', errorText(err)); });
      });
    } });
    rb.setAttribute('data-inv-revoke', it.key);
    line.appendChild(rb);
  }
  return line;
}

function playerItemsCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-items]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (INV_CARD.id !== id) { INV_CARD.id = id; INV_CARD.data = null; INV_CARD.err = null; INV_CARD.at = 0; }
  if (!INV_CARD.loading && Date.now() - INV_CARD.at > 20000) loadInvCard(id);
  var p = INV_CARD.data;
  var anchor = grid.querySelector('[data-player-tasks]') || grid.querySelector('[data-player-activity]') || grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify([p.items, p.equipped]) : INV_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-items': '', 'data-key': key, 'data-economy': 'player-items', 'data-l10n-skip': '', 'aria-labelledby': 'player-items-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-items-title', class: 'font-medium flex items-center gap-2' }, icon('items', 'h-4 w-4'), h('span', { text: t('inv.card') }));
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(INV_CARD.err ? notice('error', errorText(INV_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { reloadInv(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var write = aclCan('HIRO_ECONOMY', 1);
  var owned = (p.items || []).filter(function (it) { return it.owned; });
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' }, title,
    h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300 tabular-nums', 'data-inv-counts': '', text: t('inv.counts', { n: fmtNum(owned.length), total: fmtNum((p.items || []).length) }) }),
    h('a', { href: hashFor('items'), class: 'eco-link text-sm', text: t('inv.all') })));
  var eq = Object.keys(p.equipped || {});
  if (eq.length) {
    var byKey = {};
    (p.items || []).forEach(function (it) { byKey[it.key] = it; });
    body.appendChild(h('div', { class: 'flex flex-wrap gap-1 mb-2', 'data-inv-equipped': '' }, eq.sort().map(function (slot) {
      var it = byKey[p.equipped[slot]];
      return tag(t('inv.slot', { slot: it ? it.typeName || slot : slot, name: it ? it.name || it.key : p.equipped[slot] }), 'success');
    })));
  }
  var list = h('div', { class: 'flex flex-col', 'data-inv-list': '' });
  owned.forEach(function (it) { list.appendChild(invLine(id, it, write)); });
  if (!owned.length) list.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('inv.none') }));
  body.appendChild(list);
  if (write) {
    var pick = '';
    var picker = refPicker({ resource: 'items', value: '', label: t('inv.grant.pick'), placeholder: t('inv.grant.pick'), onchange: function (v) { pick = v || ''; } });
    var give = button({ variant: 'primary', size: 'sm', icon: 'plus', label: t('inv.grant'), onclick: function () {
      if (!pick) { toast('error', t('inv.grant.pick')); return; }
      api('POST', '/users/' + encodeURIComponent(id) + '/items', { itemId: pick }).then(function () { toast('success', t('inv.grant.done')); reloadInv(id); }, function (err) { toast('error', errorText(err)); });
    } });
    give.setAttribute('data-inv-grant', '');
    body.appendChild(h('div', { class: 'mt-3 border-t border-secondary-200 dark:border-secondary-700 pt-4 flex flex-wrap items-center gap-2', 'data-inv-give': '' },
      h('div', { class: 'min-w-0 flex-1 eco-inv-pick' }, picker), give));
  }
  place();
}
