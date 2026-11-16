ICONS['avatar'] = '<circle cx="12" cy="8" r="4"/><path d="M4 21c0-4 4-6 8-6s8 2 8 6"/>';
PH_ICONS['avatar'] = '<path d="M230.92,212c-15.23-26.33-38.7-45.21-66.09-54.16a72,72,0,1,0-73.66,0C63.78,166.78,40.31,185.66,25.08,212a8,8,0,1,0,13.85,8c18.84-32.56,52.14-52,89.07-52s70.23,19.44,89.07,52a8,8,0,1,0,13.85-8ZM72,96a56,56,0,1,1,56,56A56.06,56.06,0,0,1,72,96Z"/>';
ICONS['avatar-parts'] = ICONS['avatar'];
PH_ICONS['avatar-parts'] = PH_ICONS['avatar'];
ICONS['avatar-settings'] = '<path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6"/>';

function avPreview(it) {
  var box = h('div', { class: 'flex items-center gap-1', 'data-av-preview': '' });
  (it.options || []).slice(0, 6).forEach(function (o) {
    var cell = h('div', { class: 'h-8 w-8 shrink-0 rounded-md bg-secondary-100 dark:bg-secondary-700 overflow-hidden flex items-center justify-center', title: o.key });
    if (o.iconUrl || o.imageUrl) cell.appendChild(img(o.iconUrl || o.imageUrl));
    else if (o.value) cell.style.background = /^#[0-9a-f]{3,8}$/i.test(o.value) ? o.value : '';
    box.appendChild(cell);
  });
  if ((it.count || 0) > 6) box.appendChild(h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: '+' + fmtNum(it.count - 6) }));
  return box;
}

function avAccessOptions() {
  return [{ value: 'free', label: t('sp.access.free') }, { value: 'item', label: t('sp.access.item') }];
}

RESOURCES['avatar-parts'] = {
  api: 'base',
  path: 'avatar/parts',
  filters: [activeFilter()],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, false); } },
    { label: 'col.av.options', render: function (it) { return h('div', { class: 'flex flex-col gap-1' }, avPreview(it), h('span', { class: 'text-xs text-secondary-500', text: t('av.countN', { n: fmtNum(it.count || 0) }) })); }, skeleton: '30%' },
    { label: 'col.av.layer', render: function (it) { return h('div', { class: 'flex flex-col', 'data-av-layer': '' }, numCell(it.layer), it.required ? h('span', { class: 'text-xs text-secondary-500', text: t('av.required') }) : null); }, width: 110 },
    { label: 'col.av.players', render: function (it) { return numCell(it.players || 0); }, width: 100 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.av.key.hint', required: true, lockOnEdit: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    tr('name', { label: 'f.name', required: true, max: 100 }),
    { type: 'row', fields: [
      { name: 'layer', type: 'number', label: 'f.av.layer', hint: 'f.av.layer.hint', min: -1000, max: 1000, empty: 0 },
      { name: 'sort', type: 'number', label: 'f.sort', min: -100000, max: 100000, empty: 0 }
    ] },
    { name: 'required', type: 'bool', inline: true, label: 'f.av.required', hint: 'f.av.required.hint' },
    { type: 'heading', label: 'av.h.options' },
    { name: 'options', type: 'rows', addLabel: 'av.add', emptyText: 'av.empty', min: 1, max: 300, empty: [], refresh: true,
      newRow: function (rows) { return { key: 'o' + (rows.length + 1), name: {}, imageUrl: '', iconUrl: '', value: '', access: 'free', itemKey: null, isDefault: rows.length === 0, isActive: true }; },
      check: function (rows) {
        var seen = {};
        var defaults = 0;
        for (var i = 0; i < rows.length; i++) {
          var r = rows[i];
          if (!r.imageUrl && !r.value) return { index: i, message: t('code.OPTION_LOOK_REQUIRED') };
          if (r.access === 'item' && !r.itemKey) return { index: i, message: t('code.ITEM_REQUIRED') };
          if (r.isDefault && r.access === 'item') return { index: i, message: t('code.DEFAULT_MUST_BE_FREE') };
          if (r.isDefault) defaults++;
          if (seen[r.key]) return { index: i, message: t('code.DUPLICATE_KEY') };
          seen[r.key] = true;
        }
        if (defaults > 1) return { index: rows.length - 1, message: t('code.ONE_DEFAULT') };
        return null;
      },
      columns: [
        { name: 'imageUrl', type: 'custom', label: 'f.av.image', build: function (cur, change) { return imageInput(cur || '', function (v) { change(v || ''); }); } },
        { name: 'key', type: 'text', ltr: true, mono: true, label: 'f.key', required: true },
        { name: 'name', type: 'i18n', label: 'f.name', max: 100 },
        { name: 'value', type: 'text', ltr: true, label: 'f.av.value' },
        { name: 'access', type: 'select', label: 'f.sp.access', options: avAccessOptions },
        { name: 'itemKey', type: 'ref', resource: 'items', valueKey: 'key', label: 'f.av.item' },
        { name: 'isDefault', type: 'custom', label: 'f.av.default', build: function (cur, change) { return checkbox(!!cur, t('f.av.default'), function (on) { change(!!on); }); } }
      ] }
  ],
  defaults: function () { return { key: '', name: {}, layer: 0, sort: 0, required: false, isActive: true, options: [] }; },
  fixPayload: function (p) {
    p.layer = p.layer || 0;
    p.sort = p.sort || 0;
    p.options = (p.options || []).map(function (o) {
      return { key: o.key, name: o.name || {}, imageUrl: o.imageUrl || '', iconUrl: o.iconUrl || '', value: toLatinDigits(String(o.value || '')).trim(), access: o.access || 'free', itemKey: o.access === 'item' ? (o.itemKey || '') : '', isDefault: !!o.isDefault, isActive: o.isActive !== false };
    });
    ['id', 'count', 'players', 'createdAt', 'createdBy', 'updatedAt'].forEach(function (k) { delete p[k]; });
    return p;
  }
};

var AV_SETTINGS_FIELDS = [
  { name: 'enabled', type: 'bool', inline: true, label: 'f.av.enabled', hint: 'f.av.enabled.hint' }
];

function avatarSettingsPage(container) {
  var frame = pageFrame(t('nav.avatar-settings'), t('desc.avatar-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(baseApi, '/avatar/settings', null, AV_SETTINGS_FIELDS, {
    id: 'avatar',
    fixPayload: function (p) {
      delete p.updatedAt;
      delete p.updatedBy;
      return p;
    }
  });
  frame.content.appendChild(page.el);
  page.load();
}

CUSTOM_PAGES['avatar-settings'] = function (root) { avatarSettingsPage(root); };
