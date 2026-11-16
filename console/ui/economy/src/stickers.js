ICONS['stickers'] = '<circle cx="12" cy="12" r="9"/><path d="M8.5 14.5c1 1 2.2 1.5 3.5 1.5s2.5-.5 3.5-1.5M9 9.5h.01M15 9.5h.01"/>';
PH_ICONS['stickers'] = '<path d="M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm0,192a88,88,0,1,1,88-88A88.1,88.1,0,0,1,128,216ZM80,108a12,12,0,1,1,12,12A12,12,0,0,1,80,108Zm96,0a12,12,0,1,1-12-12A12,12,0,0,1,176,108Zm-1.07,48c-10.29,17.79-27.4,28-46.93,28s-36.63-10.2-46.92-28a8,8,0,1,1,13.84-8c7.47,12.91,19.21,20,33.08,20s25.61-7.1,33.07-20a8,8,0,0,1,13.86,8Z"/>';
ICONS['sticker-packs'] = ICONS['stickers'];
PH_ICONS['sticker-packs'] = PH_ICONS['stickers'];
ICONS['sticker-settings'] = '<path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6"/>';

function spPreview(it) {
  var box = h('div', { class: 'flex items-center gap-1', 'data-sp-preview': '' });
  (it.stickers || []).slice(0, 5).forEach(function (s) {
    var cell = h('div', { class: 'h-8 w-8 shrink-0 rounded-md bg-secondary-100 dark:bg-secondary-700 overflow-hidden flex items-center justify-center' });
    if (s.imageUrl) cell.appendChild(img(s.imageUrl));
    box.appendChild(cell);
  });
  if ((it.count || 0) > 5) box.appendChild(h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: '+' + fmtNum(it.count - 5) }));
  return box;
}

RESOURCES['sticker-packs'] = {
  api: 'base',
  path: 'stickers/packs',
  filters: [activeFilter(), { name: 'access', label: 'filter.sp.access', options: function () { return [{ value: 'free', label: t('sp.access.free') }, { value: 'item', label: t('sp.access.item') }]; } }],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, true); } },
    { label: 'col.sp.stickers', render: function (it) { return h('div', { class: 'flex flex-col gap-1' }, spPreview(it), h('span', { class: 'text-xs text-secondary-500', text: t('sp.countN', { n: fmtNum(it.count || 0) }) })); }, skeleton: '30%' },
    { label: 'col.sp.access', render: function (it) {
      if (it.access !== 'item') return h('span', { 'data-sp-access': 'free' }, tag(t('sp.access.free'), 'success'));
      var name = it.item ? (i18nText(it.item.name) || it.item.key) : it.itemKey;
      return h('div', { class: 'flex flex-col', 'data-sp-access': 'item' }, tag(t('sp.access.item'), 'warning'), h('span', { class: 'text-xs text-secondary-500 truncate', text: name }));
    }, width: 170 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.stk.key.hint', required: true, lockOnEdit: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    tr('name', { label: 'f.name', required: true, max: 100 }),
    tr('description', { multiline: true, max: 500, optional: true }),
    { type: 'row', fields: [
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { name: 'sort', type: 'number', label: 'f.sort', min: -100000, max: 100000, empty: 0 }
    ] },
    { type: 'heading', label: 'sp.h.access' },
    { name: 'access', type: 'select', label: 'f.sp.access', hint: 'f.sp.access.hint', required: true, noPlaceholder: true, options: function () { return [{ value: 'free', label: t('sp.access.free') }, { value: 'item', label: t('sp.access.item') }]; }, refresh: true },
    { name: 'itemKey', type: 'ref', resource: 'items', valueKey: 'key', label: 'f.sp.item', hint: 'f.sp.item.hint', required: true, when: function (v) { return v.access === 'item'; } },
    { type: 'heading', label: 'sp.h.stickers' },
    { name: 'stickers', type: 'rows', addLabel: 'sp.add', emptyText: 'sp.empty', min: 1, max: 200, empty: [],
      newRow: function (rows) { return { key: 's' + (rows.length + 1), name: {}, imageUrl: '', isActive: true }; },
      check: function (rows) {
        var seen = {};
        for (var i = 0; i < rows.length; i++) {
          if (!rows[i].imageUrl) return { index: i, message: t('sp.err.image') };
          if (seen[rows[i].key]) return { index: i, message: t('code.DUPLICATE_KEY') };
          seen[rows[i].key] = true;
        }
        return null;
      },
      columns: [
        { name: 'imageUrl', type: 'custom', label: 'f.sp.image', build: function (cur, change) { return imageInput(cur || '', function (v) { change(v || ''); }); } },
        { name: 'key', type: 'text', ltr: true, mono: true, label: 'f.key', required: true },
        { name: 'name', type: 'i18n', label: 'f.name', max: 100 }
      ] }
  ],
  defaults: function () { return { key: '', name: {}, description: {}, iconUrl: '', sort: 0, access: 'free', itemKey: null, isActive: true, stickers: [] }; },
  fixPayload: function (p) {
    if (p.access !== 'item') p.itemKey = '';
    p.itemKey = p.itemKey || '';
    p.sort = p.sort || 0;
    p.stickers = (p.stickers || []).map(function (s) { return { key: s.key, name: s.name || {}, imageUrl: s.imageUrl || '', isActive: s.isActive !== false }; });
    ['id', 'count', 'item', 'createdAt', 'createdBy', 'updatedAt'].forEach(function (k) { delete p[k]; });
    return p;
  }
};

var SP_SETTINGS_FIELDS = [
  { name: 'enabled', type: 'bool', inline: true, label: 'f.sp.enabled', hint: 'f.sp.enabled.hint' },
  { type: 'heading', label: 'sp.h.chats' },
  { type: 'row', fields: [
    { name: 'direct', type: 'bool', inline: true, label: 'f.sp.direct' },
    { name: 'groups', type: 'bool', inline: true, label: 'f.sp.groups' },
    { name: 'rooms', type: 'bool', inline: true, label: 'f.sp.rooms', hint: 'f.sp.rooms.hint' }
  ] },
  { type: 'heading', label: 'sp.h.quota' },
  { name: 'countQuota', type: 'bool', inline: true, label: 'f.sp.countQuota', hint: 'f.sp.countQuota.hint' }
];

function stickerSettingsPage(container) {
  var frame = pageFrame(t('nav.sticker-settings'), t('desc.sticker-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(baseApi, '/stickers/settings', null, SP_SETTINGS_FIELDS, {
    id: 'stickers',
    fixPayload: function (p) {
      delete p.updatedAt;
      delete p.updatedBy;
      return p;
    }
  });
  frame.content.appendChild(page.el);
  page.load();
}

CUSTOM_PAGES['sticker-settings'] = function (root) { stickerSettingsPage(root); };
