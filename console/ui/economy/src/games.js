PH_ICONS.games = '<path d="M220.27,158.54a8,8,0,0,0-7.7-.46,20,20,0,1,1,0-36.16A8,8,0,0,0,224,114.69V72a16,16,0,0,0-16-16H171.78a35.36,35.36,0,0,0,.22-4,36.11,36.11,0,0,0-11.36-26.24,36,36,0,0,0-60.55,23.62,36.56,36.56,0,0,0,.14,6.62H64A16,16,0,0,0,48,72v32.22a35.36,35.36,0,0,0-4-.22,36.12,36.12,0,0,0-26.24,11.36,35.7,35.7,0,0,0-9.69,27,36.08,36.08,0,0,0,33.31,33.6,35.68,35.68,0,0,0,6.62-.14V208a16,16,0,0,0,16,16H208a16,16,0,0,0,16-16V165.31A8,8,0,0,0,220.27,158.54ZM208,208H64V165.31a8,8,0,0,0-11.43-7.23,20,20,0,1,1,0-36.16A8,8,0,0,0,64,114.69V72h46.69a8,8,0,0,0,7.23-11.43,20,20,0,1,1,36.16,0A8,8,0,0,0,161.31,72H208v32.23a35.68,35.68,0,0,0-6.62-.14A36,36,0,0,0,204,176a35.36,35.36,0,0,0,4-.22Z"/>';
PH_ICONS.gamesDuo = '<path d="M204,168a28,28,0,0,0,12-2.69V208a8,8,0,0,1-8,8H64a8,8,0,0,1-8-8V165.31a28,28,0,1,1,0-50.62V72a8,8,0,0,1,8-8h46.69a28,28,0,1,1,50.61,0H208a8,8,0,0,1,8,8v42.69A28,28,0,1,0,204,168Z" opacity="0.2"/>' + PH_ICONS.games;

var GAME_PLATFORMS = ['android', 'ios', 'web', 'windows', 'macos', 'linux'];
var GAME_REG = { items: [], unconfigured: [], missingHandler: [], at: 0 };

function loadGameRegistry(force) {
  if (!force && Date.now() - GAME_REG.at < 5000) return Promise.resolve(GAME_REG);
  return api('GET', '/games-registry').then(function (r) {
    GAME_REG = { items: r.items || [], unconfigured: r.unconfigured || [], missingHandler: r.missingHandler || [], at: Date.now() };
    setTimeout(paintRegTags, 0);
    return GAME_REG;
  }, function () { return GAME_REG; });
}

function gameDef(key) {
  for (var i = 0; i < GAME_REG.items.length; i++) if (GAME_REG.items[i].key === key) return GAME_REG.items[i];
  return null;
}

function schemaText(v) {
  if (v == null) return '';
  if (typeof v === 'string') return v;
  return v[lang()] || v.en || v.fa || Object.keys(v).map(function (k) { return v[k]; })[0] || '';
}

function gameTile(it) {
  var pic = h('div', { class: 'h-9 w-9 shrink-0 rounded-md overflow-hidden flex items-center justify-center eco-gm-tile' });
  if (it.iconUrl) pic.appendChild(img(it.iconUrl, 'gamesDuo'));
  else pic.appendChild(icon('gamesDuo', 'h-5 w-5'));
  return pic;
}

function gamePlayers(it) {
  var a = Number(it.minPlayers) || 1, b = Number(it.maxPlayers) || a;
  if (b <= 1) return t('gm.solo');
  return a === b ? t('gm.players', { n: fmtNum(a) }) : t('gm.playersRange', { a: fmtNum(a), b: fmtNum(b) });
}

function gameRoomChip(r) {
  var fee = r.entryFee && r.entryFee.amount ? t('gm.room.fee', { name: r.key, n: fmtNum(r.entryFee.amount), cur: keyName('currency', r.entryFee.currencyKey) }) : t('gm.room.free', { name: r.key });
  return h('span', { class: 'eco-tag eco-tag--secondary inline-flex items-center gap-1 leading-tight max-w-full px-2 py-0.5 text-xs rounded-md whitespace-nowrap tabular-nums', 'data-gm-room': r.key }, fee,
    r.minLevel ? h('span', { class: 'opacity-80', text: '· ' + t('gm.room.level', { n: fmtNum(r.minLevel) }) }) : null);
}

function gameRegTagFor(key) {
  var def = gameDef(key);
  if (!def) return tag(t('gm.noHandler'), 'error');
  return tag(def.realtime ? t('gm.realtime') : t('gm.registered'), def.realtime ? 'primary' : 'secondary');
}

function gameRegistryTag(it) {
  var box = h('span', { class: 'inline-flex', 'data-gm-regtag': it.key });
  if (GAME_REG.at) box.appendChild(gameRegTagFor(it.key));
  return box;
}

function paintRegTags() {
  var els = document.querySelectorAll('[data-gm-regtag]');
  for (var i = 0; i < els.length; i++) append(clear(els[i]), gameRegTagFor(els[i].getAttribute('data-gm-regtag')));
}

function gameWarnings() {
  var box = h('div', { class: 'flex flex-col gap-2 mb-4', 'data-gm-warnings': '' });
  loadGameRegistry(true).then(function (reg) {
    clear(box);
    reg.unconfigured.forEach(function (d) {
      var go = button({ variant: 'link', size: 'sm', label: t('gm.configure'), onclick: function () { location.hash = hashFor('games', 'new', { key: d.key }); } });
      go.setAttribute('data-gm-configure', d.key);
      var n = notice('warning', t('gm.warn.unconfigured', { key: d.key, name: schemaText(d.name) || d.key, plugin: d.plugin || '—' }), go);
      n.setAttribute('data-gm-unconfigured', d.key);
      box.appendChild(n);
    });
    reg.missingHandler.forEach(function (d) {
      var n = notice('error', t('gm.warn.missing', { key: d.key, name: i18nText(d.name) || d.key }));
      n.setAttribute('data-gm-missing', d.key);
      box.appendChild(n);
    });
    if (!reg.unconfigured.length && !reg.missingHandler.length && reg.items.length) {
      var ok = notice('success', t('gm.warn.ok', { n: fmtNum(reg.items.length) }));
      ok.setAttribute('data-gm-ok', '');
      box.appendChild(ok);
    }
  });
  return box;
}

function schemaForm(cur, change, values) {
  var box = h('div', { class: 'flex flex-col', 'data-gm-settings': '' });
  var state = { key: null, vals: Object.assign({}, cur || {}), wraps: {}, schema: null };
  function emit() { change(Object.assign({}, state.vals)); }
  function render() {
    clear(box);
    state.wraps = {};
    var def = gameDef(values.key);
    state.key = values.key;
    state.schema = def && def.schema ? def.schema : null;
    if (!def) {
      box.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', 'data-gm-noschema': '', text: t('gm.settings.noDef') }));
      var raw = textInput(JSON.stringify(state.vals, null, 2), { ltr: true, mono: true, multiline: true, rows: 4, label: t('gm.settings.json'), oninput: function (v) {
        try { state.vals = v.trim() ? JSON.parse(v) : {}; state.bad = false; } catch (x) { state.bad = true; }
        emit();
      } });
      box.appendChild(field(t('gm.settings.json'), raw, null, { optional: true }));
      return;
    }
    var props = (state.schema && state.schema.properties) || {};
    var names = Object.keys(props);
    if (!names.length) {
      box.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: t('gm.settings.none') }));
      return;
    }
    names.forEach(function (name) {
      var p = props[name];
      var label = schemaText(p.title) || name;
      var hint = schemaText(p.description);
      var v = state.vals[name];
      if (v === undefined && p.default !== undefined) { v = p.default; state.vals[name] = v; }
      var control;
      if (p.type === 'boolean') {
        control = toggle(!!v, label, function (on) { state.vals[name] = on; emit(); });
      } else if (p.enum && p.enum.length) {
        control = selectInput(v, p.enum.map(function (e) { return { value: e, label: String(e) }; }), function (s) { state.vals[name] = s; emit(); }, { placeholder: false, label: label });
      } else if (p.type === 'integer' || p.type === 'number') {
        control = numberInput(v, { int: p.type === 'integer', label: label, oninput: function (s) { state.vals[name] = s === '' ? null : parseNumber(s, p.type === 'integer'); emit(); } });
      } else {
        control = textInput(v == null ? '' : String(v), { label: label, oninput: function (s) { state.vals[name] = s; emit(); } });
      }
      var range = [];
      if (p.minimum !== undefined) range.push(t('gm.settings.min', { n: fmtNum(p.minimum) }));
      if (p.maximum !== undefined) range.push(t('gm.settings.max', { n: fmtNum(p.maximum) }));
      if (p.default !== undefined) range.push(t('gm.settings.default', { v: typeof p.default === 'number' ? fmtNum(p.default) : String(p.default) }));
      var wrap = field(p.type === 'boolean' ? '' : label, control, [hint, range.join(' · ')].filter(Boolean).join(' — ') || null, { optional: !(state.schema.required || []).includes(name) });
      wrap.setAttribute('data-setting', name);
      state.wraps[name] = wrap;
      box.appendChild(wrap);
    });
    setTimeout(emit, 0);
  }
  box.sync = function (vals) { if (vals.key !== state.key) { state.vals = {}; render(); } };
  box.validate = function () {
    if (state.bad) return t('err.json');
    if (!state.schema) return null;
    var props = state.schema.properties || {};
    var bad = null;
    Object.keys(state.wraps).forEach(function (name) {
      var p = props[name], v = state.vals[name], msg = '';
      if (v === null || v === undefined || v === '') { if ((state.schema.required || []).includes(name) && p.default === undefined) msg = t('err.required'); }
      else if ((p.type === 'integer' || p.type === 'number') && (typeof v !== 'number' || isNaN(v) || (p.minimum !== undefined && v < p.minimum) || (p.maximum !== undefined && v > p.maximum))) msg = t('err.number');
      else if (p.type === 'string' && p.maxLength !== undefined && String(v).length > p.maxLength) msg = t('gm.settings.tooLong');
      state.wraps[name].setError(msg);
      if (msg && !bad) bad = msg;
    });
    return bad ? t('err.form') : null;
  };
  box.showError = function (name, msg) { if (state.wraps[name]) { state.wraps[name].setError(msg); return true; } return false; };
  render();
  return box;
}

function gameStats(v) {
  if (!v.id) return h('span', {});
  return h('div', { class: 'flex flex-wrap gap-3 text-sm text-secondary-700 dark:text-secondary-200 tabular-nums', 'data-gm-stats': '' },
    h('span', { text: t('gm.stat.matches', { n: fmtNum(v.matches || 0) }) }),
    h('span', { text: t('gm.stat.finished', { n: fmtNum(v.finished || 0) }) }),
    h('span', { text: t('gm.stat.cancelled', { n: fmtNum(v.cancelled || 0) }) }),
    h('span', { text: t('gm.stat.players', { n: fmtNum(v.players || 0) }) }),
    v.lastMatchAt ? h('span', { text: t('gm.stat.last', { date: fmtDate(v.lastMatchAt) }) }) : null);
}

function gameRegLine(v) {
  var def = gameDef(v.key);
  if (!v.key) return h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('gm.key.pick') });
  if (!def) return h('p', { class: 'text-xs text-error-600 dark:text-error-400', 'data-gm-reg': 'missing', text: t('gm.key.missing') });
  return h('p', { class: 'text-xs text-success-700 dark:text-success-400', 'data-gm-reg': def.realtime ? 'realtime' : 'registered', text: t(def.realtime ? 'gm.key.realtime' : 'gm.key.registered', { plugin: def.plugin || '—' }) });
}

RESOURCES.games = {
  top: function () { return gameWarnings(); },
  prepare: function () { return loadGameRegistry(true); },
  onItems: function () { loadGameRegistry().then(function () { setTimeout(paintRegTags, 0); }); },
  filters: [
    { name: 'categoryId', label: 'filter.category', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'game' } },
    { name: 'platform', label: 'filter.gm.platform', options: opts('gm.platform.', GAME_PLATFORMS) },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) {
      return h('div', { class: 'flex items-center gap-3 min-w-0' }, gameTile(it),
        h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) || '' }), h('div', { class: 'text-secondary-500' }, mono(it.key))));
    }, skeleton: '40%' },
    { label: 'col.gm.play', render: function (it) {
      return h('div', { class: 'flex flex-col gap-1 text-sm min-w-0' },
        h('span', { class: 'tabular-nums', 'data-gm-players': '', text: gamePlayers(it) }),
        it.turnSeconds ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums', text: t('gm.turnN', { n: fmtNum(it.turnSeconds) }) }) : null);
    }, width: 140 },
    { label: 'col.gm.rooms', render: function (it) {
      var rooms = it.rooms || [];
      if (!rooms.length) return h('span', { class: 'text-sm text-secondary-500', text: t('gm.rooms.none') });
      return withKeyNames(function () { return h('div', { class: 'flex flex-wrap gap-1' }, rooms.slice(0, 4).map(gameRoomChip)); });
    }, skeleton: '45%' },
    { label: 'col.gm.matches', render: function (it) {
      return h('div', { class: 'flex flex-col text-sm tabular-nums whitespace-nowrap' }, h('span', { text: t('gm.stat.matches', { n: fmtNum(it.matches || 0) }) }),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('gm.stat.players', { n: fmtNum(it.players || 0) }) }));
    }, width: 120 },
    { label: 'col.status', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, it.isActive ? tag(t('active'), 'success') : tag(t('gm.off'), 'warning'), gameRegistryTag(it));
    }, width: 170 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.gm.key', hint: 'f.gm.key.hint', required: true, lockOnEdit: true, keyRe: /^[a-z][a-z0-9_]{1,63}$/, keyErr: 'err.gm.key' },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.gm.active', hint: 'f.gm.active.hint' }
    ] },
    { name: 'regLine', type: 'static', readonly: true, live: true, render: gameRegLine },
    tr('name', { label: 'f.gm.name', required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    tr('maintenance', { label: 'f.gm.maintenance', hint: 'f.gm.maintenance.hint', multiline: true, max: 300, optional: true }),
    { name: 'stats', type: 'static', readonly: true, render: gameStats },
    { type: 'row', fields: [categoryField('game'), tagsField('game')] },
    { type: 'row', fields: [
      { name: 'iconUrl', type: 'image', label: 'f.icon', optional: true },
      { name: 'sort', type: 'number', label: 'f.gm.sort', hint: 'f.gm.sort.hint' }
    ] },
    { type: 'heading', label: 'gm.h.play' },
    { type: 'row', fields: [
      { name: 'minPlayers', type: 'number', label: 'f.gm.minPlayers', required: true, min: 1, max: 100 },
      { name: 'maxPlayers', type: 'number', label: 'f.gm.maxPlayers', required: true, min: 1, max: 100,
        check: function (v, vals) { return v != null && vals.minPlayers != null && Number(v) < Number(vals.minPlayers) ? t('code.PLAYERS_MAX_BELOW_MIN') : null; } }
    ] },
    { type: 'row', fields: [
      { name: 'turnSeconds', type: 'number', label: 'f.gm.turn', hint: 'f.gm.turn.hint', optional: true, min: 1, max: 86400 },
      { name: 'matchSeconds', type: 'number', label: 'f.gm.limit', hint: 'f.gm.limit.hint', optional: true, min: 1, max: 604800 }
    ] },
    { name: 'botsAllowed', type: 'bool', inline: true, label: 'f.gm.bots', hint: 'f.gm.bots.hint' },
    { type: 'heading', label: 'gm.h.rooms' },
    { name: 'rooms', type: 'rows', addLabel: 'gm.room.add', emptyText: 'gm.room.empty', max: 20, empty: [],
      newRow: function (rows) {
        var keys = ['bronze', 'silver', 'gold', 'diamond'];
        var used = rows.map(function (r) { return r.key; });
        var key = keys.filter(function (k) { return used.indexOf(k) === -1; })[0] || '';
        return { key: key, name: {}, feeCurrency: null, feeAmount: null, prizeId: null, prizeGroupId: null, minLevel: 0 };
      },
      check: function (rows) {
        var dup = duplicateCheck('key', 'DUPLICATE_ROOM')(rows);
        if (dup) return dup;
        for (var i = 0; i < rows.length; i++) {
          var r = rows[i];
          if (!/^[a-z][a-z0-9_]{1,63}$/.test(r.key || '')) return { index: i, message: t('code.INVALID_ROOM_KEY') };
          if ((r.feeCurrency && (r.feeAmount == null || r.feeAmount === '')) || (!r.feeCurrency && r.feeAmount != null && r.feeAmount !== '')) return { index: i, message: t('code.INVALID_ENTRY_FEE') };
          if (r.prizeId && r.prizeGroupId) return { index: i, message: t('code.ONE_ROOM_PRIZE') };
        }
        return null;
      },
      columns: [
        { name: 'key', type: 'text', ltr: true, mono: true, label: 'f.gm.room.key', required: true },
        { name: 'feeCurrency', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.gm.room.currency', optional: true },
        { name: 'feeAmount', type: 'number', label: 'f.gm.room.amount', optional: true, min: 1, max: 1000000000 },
        { name: 'minLevel', type: 'number', label: 'f.gm.room.level', optional: true, min: 0, max: 1000 },
        { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.gm.room.prize', optional: true },
        { name: 'prizeGroupId', type: 'ref', resource: 'prize-groups', label: 'f.gm.room.group', optional: true },
        { name: 'name', type: 'i18n', label: 'f.gm.room.name', max: 64, wide: true }
      ] },
    { name: 'roomsHint', type: 'static', readonly: true, render: function () { return h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('gm.rooms.hint') }); } },
    { type: 'heading', label: 'gm.h.where' },
    { name: 'platforms', type: 'multi', label: 'f.gm.platforms', hint: 'f.gm.platforms.hint', options: opts('gm.platform.', GAME_PLATFORMS) },
    { type: 'row', fields: [
      { name: 'minVersion', type: 'text', ltr: true, label: 'f.gm.minVersion', optional: true, check: function (v) { return v && !/^v?\d+(\.\d+){0,3}([-+][0-9A-Za-z.-]+)?$/.test(v) ? t('code.INVALID_VERSION') : null; } },
      { name: 'maxVersion', type: 'text', ltr: true, label: 'f.gm.maxVersion', optional: true, check: function (v) { return v && !/^v?\d+(\.\d+){0,3}([-+][0-9A-Za-z.-]+)?$/.test(v) ? t('code.INVALID_VERSION') : null; } }
    ] },
    { name: 'leaderboardId', type: 'text', ltr: true, mono: true, label: 'f.gm.board', hint: 'f.gm.board.hint', optional: true },
    { type: 'heading', label: 'gm.h.settings' },
    { name: 'settings', type: 'custom', sync: true, label: 'f.gm.settings', optional: true, build: schemaForm }
  ],
  defaults: function (query) {
    var key = (query && query.key) || '';
    var def = key ? gameDef(key) : null;
    var name = {};
    if (def && def.name) name = Object.assign({}, def.name);
    return { key: key, isActive: true, name: name, description: {}, maintenance: {}, iconUrl: '', categoryId: null, tagIds: [], minPlayers: def && def.realtime ? 2 : 1, maxPlayers: def && def.realtime ? 2 : 1,
      turnSeconds: null, matchSeconds: null, botsAllowed: false, rooms: [], platforms: [], minVersion: '', maxVersion: '', leaderboardId: '', settings: {}, sort: 0 };
  },
  toForm: function (v) {
    v.rooms = (v.rooms || []).map(function (r) {
      return { key: r.key, name: r.name || {}, feeCurrency: r.entryFee ? r.entryFee.currencyKey : null, feeAmount: r.entryFee ? r.entryFee.amount : null, prizeId: r.prizeId || null, prizeGroupId: r.prizeGroupId || null, minLevel: r.minLevel || 0 };
    });
    ['minVersion', 'maxVersion', 'leaderboardId'].forEach(function (k) { if (v[k] == null) v[k] = ''; });
    if (!v.settings) v.settings = {};
    if (!v.platforms) v.platforms = [];
    return v;
  },
  fixPayload: function (p) {
    p.rooms = (p.rooms || []).map(function (r) {
      return { key: r.key, name: r.name || {}, entryFee: r.feeCurrency ? { currencyKey: r.feeCurrency, amount: r.feeAmount } : null, prizeId: r.prizeId || '', prizeGroupId: r.prizeGroupId || '', minLevel: r.minLevel || 0 };
    });
    p.categoryId = p.categoryId || '';
    if (!p.iconUrl) p.iconUrl = '';
    if (p.sort == null) p.sort = 0;
    p.minVersion = p.minVersion || '';
    p.maxVersion = p.maxVersion || '';
    p.leaderboardId = p.leaderboardId || '';
    p.platforms = p.platforms || [];
    if (!p.settings || typeof p.settings !== 'object') p.settings = {};
    delete p.stats; delete p.regLine; delete p.roomsHint;
    return p;
  },
  afterSave: function () { GAME_REG.at = 0; forgetRefs('games'); }
};

OVERVIEW_CATALOG.push('games');
