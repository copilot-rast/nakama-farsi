PH_ICONS.bots = '<path d="M200,48H136V16a8,8,0,0,0-16,0V48H56A32,32,0,0,0,24,80V192a32,32,0,0,0,32,32H200a32,32,0,0,0,32-32V80A32,32,0,0,0,200,48Zm16,144a16,16,0,0,1-16,16H56a16,16,0,0,1-16-16V80A16,16,0,0,1,56,64H200a16,16,0,0,1,16,16Zm-52-56H92a28,28,0,0,0,0,56h72a28,28,0,0,0,0-56Zm-24,16v24H116V152ZM80,164a12,12,0,0,1,12-12h8v24H92A12,12,0,0,1,80,164Zm84,12h-8V152h8a12,12,0,0,1,0,24ZM72,108a12,12,0,1,1,12,12A12,12,0,0,1,72,108Zm88,0a12,12,0,1,1,12,12A12,12,0,0,1,160,108Z"/>';
PH_ICONS['game-matches'] = '<path d="M232,64H208V48a8,8,0,0,0-8-8H56a8,8,0,0,0-8,8V64H24A16,16,0,0,0,8,80V96a40,40,0,0,0,40,40h3.65A80.13,80.13,0,0,0,120,191.61V216H96a8,8,0,0,0,0,16h64a8,8,0,0,0,0-16H136V191.58c31.94-3.23,58.44-25.64,68.08-55.58H208a40,40,0,0,0,40-40V80A16,16,0,0,0,232,64ZM48,120A24,24,0,0,1,24,96V80H48v32q0,4,.39,8Zm144-8.9c0,35.52-29,64.64-64,64.9a64,64,0,0,1-64-64V56H192ZM232,96a24,24,0,0,1-24,24h-.5a81.81,81.81,0,0,0,.5-8.9V80h24Z"/>';

SECTIONS.economy.push('game-matches', 'bots');
NAV.economy.forEach(function (g) { if (g && g.id === 'play') g.items.splice(g.items.indexOf('games') + 1, 0, 'game-matches', 'bots'); });

var BOT_SEL = { ids: {}, page: [], bar: null, reload: null };
var GM_STATUSES = ['open', 'started', 'finished', 'cancelled'];

function walletMap(rows) {
  var out = {};
  (rows || []).forEach(function (r) { if (r && r.currency && r.amount != null && r.amount !== '') out[r.currency] = Number(r.amount); });
  return out;
}

function walletRows(map) {
  return Object.keys(map || {}).map(function (k) { return { currency: k, amount: map[k] }; });
}

function walletText(w) {
  var keys = Object.keys(w || {});
  if (!keys.length) return t('none');
  return keys.map(function (k) { return fmtNum(w[k]) + ' ' + keyName('currency', k); }).join('، ');
}

function roomsText(rooms) {
  if (!rooms || !rooms.length) return t('bt.rooms.any');
  return rooms.map(function (r) { return r.game + (r.room ? '/' + r.room : ''); }).join('، ');
}

function botCheck(it) {
  var box = checkbox(!!BOT_SEL.ids[it.userId], '', function (on) {
    if (on) BOT_SEL.ids[it.userId] = true;
    else delete BOT_SEL.ids[it.userId];
    if (BOT_SEL.bar) BOT_SEL.bar.paint();
  });
  box.setAttribute('aria-label', t('bt.select', { name: it.username }));
  box.setAttribute('data-bot-select', it.userId);
  return box;
}

LISTS.bots = {
  path: 'bots',
  create: function () {
    var wrap = h('div', { class: 'flex flex-wrap gap-2' },
      button({ variant: 'primary', size: 'page', icon: 'plus', label: t('bt.bulk'), onclick: botBulkPanel }),
      button({ variant: 'ghost', size: 'page', icon: 'upload', label: t('bt.import'), onclick: botImportPanel }),
      button({ variant: 'ghost', size: 'page', icon: 'currencies', label: t('bt.topup'), onclick: function () { botTopupPanel(Object.keys(BOT_SEL.ids)); } }));
    wrap.querySelectorAll('button')[0].setAttribute('data-bot-action', 'bulk');
    wrap.querySelectorAll('button')[1].setAttribute('data-bot-action', 'import');
    wrap.querySelectorAll('button')[2].setAttribute('data-bot-action', 'topup');
    return wrap;
  },
  top: function () { return botKpis(); },
  onItems: function (items) { BOT_SEL.page = items; keyNames(); },
  onRow: function (it) { botEditPanel(it.userId); },
  filters: [
    { name: 'enabled', label: 'filter.status', options: function () { return [{ value: 'true', label: t('bt.on') }, { value: 'false', label: t('bt.off') }]; } },
    { name: 'template', label: 'filter.bt.template', type: 'text' },
    { name: 'game', label: 'filter.bt.game', type: 'text' },
    { name: 'room', label: 'filter.bt.room', type: 'text' },
    { name: 'minLevel', label: 'filter.bt.minLevel', type: 'text' },
    { name: 'maxLevel', label: 'filter.bt.maxLevel', type: 'text' },
    { name: 'sort', label: 'filter.bt.sort', options: function () { return ['level', 'games', 'recent'].map(function (s) { return { value: s, label: t('bt.sort.' + s) }; }); } }
  ],
  columns: [
    { label: 'col.select', render: botCheck, width: 48 },
    { label: 'col.bt.bot', render: function (it) {
      return h('div', { class: 'min-w-0', 'data-bot-row': it.userId },
        h('div', { class: 'font-medium truncate', dir: 'auto', text: it.displayName || it.username }),
        h('div', { class: 'text-secondary-500 flex items-center gap-2 min-w-0' }, it.displayName ? h('span', { class: 'truncate', dir: 'auto', text: it.username }) : null, mono(String(it.userId).slice(0, 8))));
    }, skeleton: '40%' },
    { label: 'col.bt.level', render: function (it) { return h('span', { class: 'tabular-nums', 'data-bot-level': '', text: fmtNum(it.level) }); }, width: 70 },
    { label: 'col.bt.template', render: function (it) { return it.template ? mono(it.template) : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 110 },
    { label: 'col.bt.rooms', render: function (it) { return h('span', { class: 'text-sm', dir: 'ltr', text: roomsText(it.rooms) }); }, skeleton: '30%' },
    { label: 'col.bt.wallet', render: function (it) { return withKeyNames(function () { return h('span', { class: 'text-sm tabular-nums', text: walletText(it.wallet) }); }); }, skeleton: '30%' },
    { label: 'col.bt.games', render: function (it) {
      return h('div', { class: 'flex flex-col text-sm' }, h('span', { class: 'tabular-nums', text: t('bt.games', { n: fmtNum(it.games || 0) }) }),
        it.lastUsedAt ? h('span', { class: 'text-xs text-secondary-500', text: fmtDate(it.lastUsedAt) }) : null);
    }, width: 140 },
    { label: 'col.status', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, it.enabled ? tag(t('bt.on'), 'success') : tag(t('bt.off'), 'secondary'), it.banned ? tag(t('bt.banned'), 'error') : null);
    }, width: 120 }
  ]
};

function botKpis() {
  var box = h('div', { class: 'eco-stat-grid mb-4', 'data-bot-kpis': '' }, [1, 2, 3].map(function () { return card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem'), bar('3rem', 'h-7'))); }));
  Promise.all([api('GET', '/bots?limit=1'), api('GET', '/bots?limit=1&enabled=true')]).then(function (r) {
    clear(box);
    var total = r[0].total || 0, on = r[1].total || 0;
    var a = statCard(t('bt.kpi.total'), total, null, null); a.setAttribute('data-bot-kpi', 'total'); box.appendChild(a);
    var b = statCard(t('bt.kpi.on'), on, null, null); b.setAttribute('data-bot-kpi', 'enabled'); box.appendChild(b);
    var c = statCard(t('bt.kpi.off'), total - on, null, null); c.setAttribute('data-bot-kpi', 'disabled'); box.appendChild(c);
    var first = (r[0].items || [])[0];
    if (!first) return;
    api('GET', '/bots/' + encodeURIComponent(first.userId)).then(function (d) {
      var tpl = (d.summary && d.summary.templates) || {};
      Object.keys(tpl).sort().forEach(function (k) {
        var s = statCard(t('bt.kpi.template', { name: k || t('none') }), tpl[k], null, hashFor('bots', null, { template: k }));
        s.setAttribute('data-bot-kpi', 'template-' + (k || 'none'));
        box.appendChild(s);
      });
    }, function () {});
  }, function () { clear(box); });
  return box;
}

function botBulkBar(reload) {
  var count = h('span', { class: 'text-sm font-medium tabular-nums', 'aria-live': 'polite', 'data-bot-selected': '' });
  function run(enabled, btn) {
    var ids = Object.keys(BOT_SEL.ids);
    if (!ids.length) return;
    setBusy(btn, true, t('saving'));
    api('POST', '/bots/status', { ids: ids, enabled: enabled }).then(function (res) {
      setBusy(btn, false);
      toast('success', t('bt.status.done', { n: fmtNum(res.updated) }));
      BOT_SEL.ids = {};
      paint();
      reload();
    }, function (err) { setBusy(btn, false); toast('error', errorText(err)); });
  }
  var on = button({ variant: 'primary', size: 'sm', label: t('bt.enable'), onclick: function (e) { run(true, e.currentTarget); } });
  var off = button({ variant: 'danger', size: 'sm', label: t('bt.disable'), onclick: function (e) { run(false, e.currentTarget); } });
  var top = button({ variant: 'ghost', size: 'sm', icon: 'currencies', label: t('bt.topup.sel'), onclick: function () { botTopupPanel(Object.keys(BOT_SEL.ids)); } });
  on.setAttribute('data-bot-bulk', 'enable');
  off.setAttribute('data-bot-bulk', 'disable');
  top.setAttribute('data-bot-bulk', 'topup');
  var pageAll = button({ variant: 'link', size: 'sm', label: t('pl.bulk.page'), onclick: function () { BOT_SEL.page.forEach(function (it) { BOT_SEL.ids[it.userId] = true; }); paint(); reload(); } });
  pageAll.setAttribute('data-bot-bulk', 'page');
  var none = button({ variant: 'ghost', size: 'sm', label: t('pl.bulk.clear'), onclick: function () { BOT_SEL.ids = {}; paint(); reload(); } });
  var el = h('div', { class: 'eco-bulk rounded-md border border-secondary-200 dark:border-secondary-700 p-3 mb-3', 'data-bot-bulkbar': '' },
    h('div', { class: 'flex flex-wrap items-center gap-2' }, count, pageAll, none, h('span', { class: 'flex-1' }), on, off, top));
  function paint() {
    var n = Object.keys(BOT_SEL.ids).length;
    count.textContent = n ? t('pl.bulk.selected', { n: n }) : t('pl.bulk.none');
    on.disabled = off.disabled = top.disabled = !n;
  }
  el.paint = paint;
  paint();
  return el;
}

function botsPage(container, query, ctx) {
  BOT_SEL.ids = {};
  var list = listPage(container, 'bots', query, ctx);
  BOT_SEL.reload = list.reload;
  var bar = botBulkBar(function () { list.reload(); });
  BOT_SEL.bar = bar;
  var table2 = container.querySelector('.eco-filters');
  var anchor = table2 ? table2.closest('.flex.flex-col.items-start') : null;
  if (anchor && anchor.parentNode) anchor.parentNode.insertBefore(bar, anchor.nextSibling);
  return list;
}

CUSTOM_PAGES.bots = function (root, route, ctl, ctx) { ctl.list = botsPage(root, route.query || {}, ctx); };

var BOT_ROOM_COLS = [
  { name: 'game', type: 'text', ltr: true, mono: true, label: 'f.bt.game', required: true },
  { name: 'room', type: 'text', ltr: true, mono: true, label: 'f.bt.room', optional: true }
];
var BOT_WALLET_COLS = [
  { name: 'currency', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.bt.currency', required: true },
  { name: 'amount', type: 'number', label: 'f.bt.amount', required: true, min: 0, max: 1000000000000 }
];
var BOT_BAND_COLS = [
  { name: 'min', type: 'number', label: 'f.bt.band.min', required: true, min: 1, max: 1000 },
  { name: 'max', type: 'number', label: 'f.bt.band.max', optional: true, min: 0, max: 1000 },
  { name: 'template', type: 'text', ltr: true, mono: true, label: 'f.bt.template', required: true }
];

function botFieldsCommon() {
  return [
    { name: 'rooms', type: 'rows', label: 'f.bt.rooms', hint: 'f.bt.rooms.hint', addLabel: 'bt.room.add', emptyText: 'bt.rooms.any', max: 50, empty: [], columns: BOT_ROOM_COLS, newRow: function () { return { game: '', room: '' }; } },
    { name: 'wallet', type: 'rows', label: 'f.bt.wallet', hint: 'f.bt.wallet.hint', addLabel: 'bt.wallet.add', emptyText: 'bt.wallet.empty', max: 20, empty: [], columns: BOT_WALLET_COLS, newRow: function () { return { currency: null, amount: null }; } }
  ];
}

function cleanBotRooms(rows) {
  return (rows || []).filter(function (r) { return r && r.game; }).map(function (r) { return { game: r.game, room: r.room || '' }; });
}

function lookMap(rows) {
  var out = {};
  (rows || []).forEach(function (r) { if (r && r.slot && r.item) out[r.slot] = r.item; });
  return out;
}

function panelOpen(title, body, footer) {
  var close = function () { document.removeEventListener('keydown', esc, true); el.remove(); };
  var esc = function (e) { if (e.key === 'Escape') { e.preventDefault(); close(); } };
  var el = sidePanel(title, body, footer, close);
  document.addEventListener('keydown', esc, true);
  document.body.appendChild(el);
  return close;
}

function botEditPanel(id) {
  var body = h('div', { class: 'flex flex-col gap-4', 'data-bot-edit': id }, h('div', { class: 'flex flex-col gap-3' }, bar('60%'), bar('80%'), bar('40%')));
  var footer = h('div', { class: 'flex gap-4 w-full items-center justify-end' });
  var close = panelOpen(t('bt.edit'), body, footer);
  api('GET', '/bots/' + encodeURIComponent(id)).then(function (b) {
    var values = { enabled: !!b.enabled, level: b.level, template: b.template || '', username: b.username, displayName: b.displayName || '', avatarUrl: b.avatarUrl || '', note: b.note || '',
      rooms: (b.rooms || []).map(function (r) { return { game: r.game, room: r.room || '' }; }),
      look: Object.keys(b.look || {}).map(function (k) { return { slot: k, item: b.look[k] }; }), wallet: walletRows(b.wallet) };
    var form = buildForm([
      { type: 'row', fields: [
        { name: 'enabled', type: 'bool', inline: true, label: 'f.bt.enabled', hint: 'f.bt.enabled.hint' },
        { name: 'level', type: 'number', label: 'f.bt.level', required: true, min: 1, max: 1000 }
      ] },
      { type: 'row', fields: [
        { name: 'template', type: 'text', ltr: true, label: 'f.bt.template', hint: 'f.bt.template.hint', optional: true },
        { name: 'username', type: 'text', ltr: true, label: 'f.bt.username', required: true }
      ] },
      { type: 'row', fields: [
        { name: 'displayName', type: 'text', label: 'f.bt.displayName', optional: true },
        { name: 'avatarUrl', type: 'text', ltr: true, label: 'f.bt.avatarUrl', optional: true }
      ] },
      { name: 'note', type: 'textarea', label: 'f.bt.note', optional: true },
      botFieldsCommon()[0],
      { name: 'look', type: 'rows', label: 'f.bt.look', hint: 'f.bt.look.hint', addLabel: 'bt.look.add', emptyText: 'bt.look.empty', max: 20, empty: [],
        columns: [{ name: 'slot', type: 'text', ltr: true, mono: true, label: 'f.bt.look.slot', required: true }, { name: 'item', type: 'text', ltr: true, mono: true, label: 'f.bt.look.item', required: true }],
        newRow: function () { return { slot: '', item: '' }; } },
      botFieldsCommon()[1]
    ], values);
    clear(body);
    body.appendChild(h('div', { class: 'flex flex-wrap gap-2 text-sm text-secondary-600 dark:text-secondary-300' }, mono(b.userId), tag(t('bt.source.' + (b.source || 'console')), 'secondary'),
      h('span', { class: 'tabular-nums', text: t('bt.games', { n: fmtNum(b.games || 0) }) }), b.updatedBy ? h('span', { text: t('bt.updatedBy', { by: b.updatedBy }) }) : null));
    body.appendChild(form.el);
    body.appendChild(gameStatsPanel(b.userId));
    var save = button({ variant: 'primary', size: 'footer', label: t('save'), onclick: function () {
      if (!form.validate()) return;
      var p = form.payload();
      var payload = { enabled: !!p.enabled, level: Number(p.level), template: (p.template || '').trim(), username: (p.username || '').trim(), displayName: (p.displayName || '').trim(),
        avatarUrl: (p.avatarUrl || '').trim(), note: p.note || '', rooms: cleanBotRooms(p.rooms), look: lookMap(p.look), wallet: walletMap(p.wallet) };
      setBusy(save, true, t('saving'));
      api('PUT', '/bots/' + encodeURIComponent(id), payload).then(function () {
        toast('success', t('saved'));
        close();
        if (BOT_SEL.reload) BOT_SEL.reload();
      }, function (err) {
        setBusy(save, false);
        if (!form.showServerError(err)) toast('error', errorText(err));
      });
    } });
    save.setAttribute('data-bot-save', '');
    clear(footer).appendChild(button({ variant: 'ghost', size: 'footer', label: t('cancel'), onclick: close }));
    footer.appendChild(save);
  }, function (err) { clear(body).appendChild(notice('error', errorText(err))); });
}

function problemList(problems) {
  if (!problems || !problems.length) return null;
  var ul = h('ul', { class: 'flex flex-col gap-1 text-sm', 'data-bot-problems': String(problems.length) });
  problems.slice(0, 200).forEach(function (p) {
    ul.appendChild(h('li', { class: 'flex flex-wrap gap-2', 'data-bot-problem': p.code },
      h('span', { class: 'tabular-nums text-secondary-500', text: t('bt.line', { n: fmtNum(p.line) }) }),
      h('span', { dir: 'auto', class: 'font-medium', text: p.username || '' }),
      h('span', { class: 'text-error-600 dark:text-error-400', text: has('code.' + p.code) ? t('code.' + p.code) : p.code })));
  });
  return ul;
}

function bandsPayload(rows) {
  return (rows || []).filter(function (r) { return r && r.min != null && r.template; }).map(function (r) { return { min: Number(r.min), max: r.max == null || r.max === '' ? 0 : Number(r.max), template: r.template }; });
}

function botBulkPanel() {
  var values = { count: 10, prefix: '', names: '', levelMin: 2, levelMax: 25, template: 'normal', bands: [], rooms: [], wallet: [], note: '' };
  var form = buildForm([
    { type: 'row', fields: [
      { name: 'count', type: 'number', label: 'f.bt.count', hint: 'f.bt.count.hint', optional: true, min: 1, max: 5000 },
      { name: 'prefix', type: 'text', ltr: true, label: 'f.bt.prefix', hint: 'f.bt.prefix.hint', optional: true }
    ] },
    { name: 'names', type: 'textarea', label: 'f.bt.names', hint: 'f.bt.names.hint', optional: true },
    { type: 'row', fields: [
      { name: 'levelMin', type: 'number', label: 'f.bt.levelMin', required: true, min: 1, max: 1000 },
      { name: 'levelMax', type: 'number', label: 'f.bt.levelMax', required: true, min: 1, max: 1000 }
    ] },
    { name: 'template', type: 'text', ltr: true, label: 'f.bt.template', hint: 'f.bt.template.hint', optional: true },
    { name: 'bands', type: 'rows', label: 'f.bt.bands', hint: 'f.bt.bands.hint', addLabel: 'bt.band.add', emptyText: 'bt.bands.empty', max: 20, empty: [], columns: BOT_BAND_COLS, newRow: function () { return { min: null, max: null, template: '' }; } },
    botFieldsCommon()[0],
    botFieldsCommon()[1],
    { name: 'note', type: 'text', label: 'f.bt.note', optional: true }
  ], values);
  var result = h('div', { 'data-bot-bulk-result': '' });
  var body = h('div', { class: 'flex flex-col gap-4', 'data-bot-bulk-form': '' }, notice('info', t('bt.bulk.note')), form.el, result);
  var close;
  var save = button({ variant: 'primary', size: 'footer', icon: 'plus', label: t('bt.bulk.go'), onclick: function () {
    if (!form.validate()) return;
    var p = form.payload();
    var names = String(p.names || '').split(/\r?\n/).map(function (s) { return s.trim(); }).filter(Boolean);
    var payload = { count: Number(p.count) || 0, prefix: (p.prefix || '').trim(), names: names, levelMin: Number(p.levelMin), levelMax: Number(p.levelMax), template: (p.template || '').trim(),
      bands: bandsPayload(p.bands), rooms: cleanBotRooms(p.rooms), wallet: walletMap(p.wallet), note: p.note || '' };
    setBusy(save, true, t('saving'));
    api('POST', '/bots/bulk', payload).then(function (res) {
      setBusy(save, false);
      clear(result).appendChild(notice('success', t('bt.bulk.done', { n: fmtNum(res.created) })));
      append(result, problemList(res.problems));
      if (BOT_SEL.reload) BOT_SEL.reload();
    }, function (err) {
      setBusy(save, false);
      if (!form.showServerError(err)) clear(result).appendChild(notice('error', errorText(err)));
    });
  } });
  save.setAttribute('data-bot-bulk-go', '');
  close = panelOpen(t('bt.bulk'), body, h('div', { class: 'flex gap-4 w-full items-center justify-end' }, button({ variant: 'ghost', size: 'footer', label: t('close'), onclick: function () { close(); } }), save));
}

function botImportPanel() {
  var values = { csv: '', levelMin: 2, levelMax: 25, template: 'normal', bands: [] };
  var form = buildForm([
    { name: 'csv', type: 'textarea', ltr: true, label: 'f.bt.csv', hint: 'f.bt.csv.hint', required: true },
    { type: 'row', fields: [
      { name: 'levelMin', type: 'number', label: 'f.bt.levelMin', hint: 'f.bt.levelCsv.hint', required: true, min: 1, max: 1000 },
      { name: 'levelMax', type: 'number', label: 'f.bt.levelMax', required: true, min: 1, max: 1000 }
    ] },
    { name: 'template', type: 'text', ltr: true, label: 'f.bt.template', hint: 'f.bt.template.hint', optional: true },
    { name: 'bands', type: 'rows', label: 'f.bt.bands', hint: 'f.bt.bands.hint', addLabel: 'bt.band.add', emptyText: 'bt.bands.empty', max: 20, empty: [], columns: BOT_BAND_COLS, newRow: function () { return { min: null, max: null, template: '' }; } }
  ], values);
  var file = h('input', { type: 'file', accept: '.csv,text/csv', class: 'text-sm', 'data-bot-csv-file': '', 'aria-label': t('f.bt.csv') });
  file.addEventListener('change', function () {
    var f = file.files && file.files[0];
    if (!f) return;
    f.text().then(function (txt) {
      values.csv = txt;
      var ta = form.el.querySelector('textarea');
      if (ta) ta.value = txt;
    });
  });
  var result = h('div', { 'data-bot-import-result': '' });
  var body = h('div', { class: 'flex flex-col gap-4', 'data-bot-import-form': '' }, notice('info', t('bt.import.note')), file, form.el, result);
  var close;
  var go = button({ variant: 'primary', size: 'footer', label: t('bt.import.go'), disabled: true, onclick: function () { send(false, go); } });
  go.setAttribute('data-bot-import-go', '');
  var dry = button({ variant: 'ghost', size: 'footer', label: t('bt.import.dry'), onclick: function () { send(true, dry); } });
  dry.setAttribute('data-bot-import-dry', '');
  function send(dryRun, btn) {
    if (!form.validate()) return;
    var p = form.payload();
    setBusy(btn, true, t('saving'));
    api('POST', '/bots/import', { csv: values.csv || p.csv, levelMin: Number(p.levelMin), levelMax: Number(p.levelMax), template: (p.template || '').trim(), bands: bandsPayload(p.bands), dryRun: dryRun }).then(function (res) {
      setBusy(btn, false);
      clear(result);
      if (dryRun) {
        result.appendChild(notice(res.problems && res.problems.length ? 'warning' : 'success', t('bt.import.preview', { n: fmtNum(res.valid), bad: fmtNum((res.problems || []).length) })));
        go.disabled = !res.valid;
      } else {
        result.appendChild(notice('success', t('bt.bulk.done', { n: fmtNum(res.created) })));
        go.disabled = true;
        if (BOT_SEL.reload) BOT_SEL.reload();
      }
      append(result, problemList(res.problems));
    }, function (err) {
      setBusy(btn, false);
      if (!form.showServerError(err)) clear(result).appendChild(notice('error', errorText(err)));
    });
  }
  close = panelOpen(t('bt.import'), body, h('div', { class: 'flex gap-4 w-full items-center justify-end' }, button({ variant: 'ghost', size: 'footer', label: t('close'), onclick: function () { close(); } }), dry, go));
}

function botTopupPanel(ids) {
  var values = { scope: ids && ids.length ? 'selected' : 'all', currency: null, below: null, amount: null };
  var form = buildForm([
    { name: 'scope', type: 'select', noPlaceholder: true, label: 'f.bt.scope', options: function () {
      var o = [{ value: 'all', label: t('bt.scope.all') }];
      if (ids && ids.length) o.unshift({ value: 'selected', label: t('bt.scope.selected', { n: fmtNum(ids.length) }) });
      return o;
    } },
    { name: 'currency', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.bt.currency', required: true },
    { type: 'row', fields: [
      { name: 'below', type: 'number', label: 'f.bt.below', hint: 'f.bt.below.hint', optional: true, min: 0, max: 1000000000000 },
      { name: 'amount', type: 'number', label: 'f.bt.topAmount', hint: 'f.bt.topAmount.hint', required: true, min: 1, max: 1000000000000 }
    ] }
  ], values);
  var result = h('div', { 'data-bot-topup-result': '' });
  var body = h('div', { class: 'flex flex-col gap-4', 'data-bot-topup-form': '' }, form.el, result);
  var close;
  var go = button({ variant: 'primary', size: 'footer', label: t('bt.topup.go'), onclick: function () {
    if (!form.validate()) return;
    var p = form.payload();
    setBusy(go, true, t('saving'));
    api('POST', '/bots/topup', { ids: p.scope === 'selected' ? ids : [], all: p.scope === 'all', currency: p.currency, below: p.below == null ? 0 : Number(p.below), amount: Number(p.amount) }).then(function (res) {
      setBusy(go, false);
      clear(result).appendChild(notice('success', t('bt.topup.done', { n: fmtNum(res.toppedUp) })));
      if (BOT_SEL.reload) BOT_SEL.reload();
    }, function (err) {
      setBusy(go, false);
      if (!form.showServerError(err)) clear(result).appendChild(notice('error', errorText(err)));
    });
  } });
  go.setAttribute('data-bot-topup-go', '');
  close = panelOpen(t('bt.topup'), body, h('div', { class: 'flex gap-4 w-full items-center justify-end' }, button({ variant: 'ghost', size: 'footer', label: t('close'), onclick: function () { close(); } }), go));
}

function fmtSeconds(s) {
  s = Number(s) || 0;
  var hrs = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60);
  return hrs ? t('gst.hm', { h: fmtNum(hrs), m: fmtNum(m) }) : t('gst.m', { m: fmtNum(m) });
}

function gameStatsPanel(userId) {
  var body = h('div', { class: 'flex flex-col gap-2' }, bar('60%'));
  var box = h('section', { class: 'border border-secondary-200 dark:border-secondary-700 rounded-md p-3', 'data-game-stats': userId },
    h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' }, h('h3', { class: 'font-medium flex items-center gap-2' }, icon('games', 'h-4 w-4'), h('span', { text: t('gst.title') })),
      (function () { var b = button({ variant: 'link', size: 'sm', icon: 'plus', label: t('gst.add'), onclick: function () { statsEditor(userId, null, load); } }); b.setAttribute('data-game-stats-add', ''); return b; })()),
    body);
  function load() {
    api('GET', '/game-stats?userId=' + encodeURIComponent(userId)).then(function (res) {
      clear(body);
      var items = res.items || [];
      if (!items.length) { body.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('gst.none') })); return; }
      items.forEach(function (s) {
        var edit = button({ variant: 'link', size: 'sm', label: t('gst.editBtn'), onclick: function () { statsEditor(userId, s, load); } });
        edit.setAttribute('data-game-stats-edit', s.game);
        body.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 text-sm tabular-nums', 'data-game-stats-row': s.game },
          mono(s.game),
          h('span', { text: t('gst.line', { games: fmtNum(s.totalGames), wins: fmtNum(s.wins), losses: fmtNum(s.losses), forfeits: fmtNum(s.forfeits), timeouts: fmtNum(s.timeouts) }) }),
          h('span', { class: 'text-secondary-500', text: fmtSeconds(s.playtimeSeconds) }),
          withKeyNames(function () { return h('span', { class: 'text-secondary-500', text: walletText(s.winnings) }); }),
          h('span', { class: 'flex-1' }), edit));
      });
    }, function (err) { clear(body).appendChild(notice('error', errorText(err))); });
  }
  load();
  return box;
}

function statsEditor(userId, cur, done) {
  var values = { game: cur ? cur.game : '', wins: cur ? cur.wins : 0, losses: cur ? cur.losses : 0, forfeits: cur ? cur.forfeits : 0, timeouts: cur ? cur.timeouts : 0,
    playtimeSeconds: cur ? cur.playtimeSeconds : 0, winnings: walletRows(cur ? cur.winnings : {}) };
  var form = buildForm([
    { name: 'game', type: 'key', label: 'f.gst.game', required: true, lockOnEdit: !!cur },
    { type: 'row', fields: [{ name: 'wins', type: 'number', label: 'f.gst.wins', required: true, min: 0, max: 2147483647 }, { name: 'losses', type: 'number', label: 'f.gst.losses', required: true, min: 0, max: 2147483647 }] },
    { type: 'row', fields: [{ name: 'forfeits', type: 'number', label: 'f.gst.forfeits', required: true, min: 0, max: 2147483647 }, { name: 'timeouts', type: 'number', label: 'f.gst.timeouts', required: true, min: 0, max: 2147483647 }] },
    { name: 'playtimeSeconds', type: 'number', label: 'f.gst.playtime', required: true, min: 0, max: 2147483647 },
    { name: 'winnings', type: 'rows', label: 'f.gst.winnings', addLabel: 'bt.wallet.add', emptyText: 'bt.wallet.empty', max: 20, empty: [], columns: BOT_WALLET_COLS, newRow: function () { return { currency: null, amount: null }; } }
  ], values);
  if (cur) { var k = form.el.querySelector('input'); if (k) { k.readOnly = true; k.classList.add('eco-readonly'); } }
  var body = h('div', { class: 'flex flex-col gap-4', 'data-game-stats-form': '' }, notice('info', t('gst.note')), form.el);
  var close;
  var save = button({ variant: 'primary', size: 'footer', label: t('save'), onclick: function () {
    if (!form.validate()) return;
    var p = form.payload();
    setBusy(save, true, t('saving'));
    api('PUT', '/game-stats/' + encodeURIComponent(userId) + '/' + encodeURIComponent(p.game), { wins: Number(p.wins), losses: Number(p.losses), forfeits: Number(p.forfeits), timeouts: Number(p.timeouts),
      playtimeSeconds: Number(p.playtimeSeconds), winnings: walletMap(p.winnings) }).then(function () {
      toast('success', t('saved'));
      close();
      done();
    }, function (err) { setBusy(save, false); if (!form.showServerError(err)) toast('error', errorText(err)); });
  } });
  save.setAttribute('data-game-stats-save', '');
  close = panelOpen(t('gst.edit'), body, h('div', { class: 'flex gap-4 w-full items-center justify-end' }, button({ variant: 'ghost', size: 'footer', label: t('cancel'), onclick: function () { close(); } }), save));
}

var GST_CARD = { id: null, el: null };

function playerGameStatsCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-game-stats]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  var key = lang() + '|' + id;
  if (existing && existing.getAttribute('data-key') === key && existing.parentNode === grid) return;
  if (existing) existing.remove();
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white px-4 py-3', 'data-player-game-stats': '', 'data-key': key, 'data-economy': 'player-game-stats', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, gameStatsPanel(id));
  grid.appendChild(cardEl);
  GST_CARD.id = id;
}

function gmStatusTag(s) {
  var c = { open: 'warning', started: 'primary', finished: 'success', cancelled: 'secondary' }[s] || 'secondary';
  return tag(t('gmm.status.' + s), c);
}

LISTS['game-matches'] = {
  path: 'game-matches',
  noCreate: true,
  rowHref: function (it) { return hashFor('game-matches', it.id, lastQuery['game-matches'] || {}); },
  filters: [
    { name: 'status', label: 'filter.status', options: function () { return GM_STATUSES.map(function (s) { return { value: s, label: t('gmm.status.' + s) }; }); } },
    { name: 'gameKey', label: 'filter.gmm.game', type: 'text' },
    { name: 'userId', label: 'filter.gmm.user', type: 'text' }
  ],
  columns: [
    { label: 'col.gmm.match', render: function (it) {
      return h('div', { class: 'min-w-0', 'data-gm-match': it.id }, h('div', { class: 'font-medium truncate', text: (i18nText(it.gameName) || it.gameKey) + (it.roomKey ? ' · ' + it.roomKey : '') }), mono(String(it.id).slice(0, 13)));
    }, skeleton: '40%' },
    { label: 'col.gmm.players', render: function (it) {
      return h('span', { class: 'tabular-nums text-sm', text: t('gmm.players', { n: fmtNum(it.players || 0), bots: fmtNum(it.bots || 0) }) });
    }, width: 130 },
    { label: 'col.status', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, gmStatusTag(it.status), it.endReason ? tag(gmReason(it.endReason), 'secondary') : null, it.forcedBy ? tag(t('gmm.forced'), 'warning') : null);
    }, width: 220 },
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 150 }
  ]
};

function gmReason(r) {
  return has('gmm.reason.' + r) ? t('gmm.reason.' + r) : r;
}

CUSTOM_PAGES['game-matches'] = function (root, route, ctl, ctx) {
  if (route.id) gameMatchDetail(root, route.id);
  else ctl.list = listPage(root, 'game-matches', route.query || {}, ctx);
};

function gameMatchDetail(container, id) {
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('game-matches', null, lastQuery['game-matches'] || {}); } });
  var frame = pageFrame(t('gmm.title'), null, null);
  frame.actions.appendChild(backBtn);
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-detail-grid', 'data-gm-detail': id });
  frame.content.appendChild(grid);
  grid.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('40%'), bar('70%'), bar('50%')), 'eco-span-all'));
  function load() {
    api('GET', '/game-matches/' + encodeURIComponent(id)).then(render, function (err) { clear(grid).appendChild(errorBlock(err, load)); });
  }
  function render(m) {
    clear(grid);
    var actions = h('div', { class: 'flex flex-wrap gap-2' });
    if (m.status === 'started') {
      var end = button({ variant: 'primary', size: 'page', icon: 'prizes', label: t('gmm.end'), onclick: function () { endDialog(m); } });
      end.setAttribute('data-gm-action', 'end');
      actions.appendChild(end);
      var live = button({ variant: 'ghost', size: 'page', icon: 'refresh', label: t('gmm.live'), onclick: function () { liveDialog(m.id); } });
      live.setAttribute('data-gm-action', 'live');
      actions.appendChild(live);
    }
    if (m.status === 'open' || m.status === 'started') {
      var cancel = button({ variant: 'danger', size: 'page', icon: 'trash', label: t('gmm.cancel'), onclick: function () { cancelDialog(m); } });
      cancel.setAttribute('data-gm-action', 'cancel');
      actions.appendChild(cancel);
    }
    clear(frame.actions).appendChild(actions);
    frame.actions.appendChild(backBtn);
    grid.appendChild(section('gmm.info', dl([
      ['col.status', h('div', { class: 'flex flex-wrap gap-1' }, gmStatusTag(m.status), m.endReason ? tag(gmReason(m.endReason), 'secondary') : null)],
      ['gmm.game', h('span', {}, mono(m.gameKey), m.roomKey ? h('span', { text: ' / ' }) : null, m.roomKey ? mono(m.roomKey) : null)],
      ['gmm.id', mono(m.id)],
      ['gmm.players.label', h('span', { class: 'tabular-nums', text: t('gmm.players', { n: fmtNum(m.players || 0), bots: fmtNum(m.bots || 0) }) })],
      ['gmm.fees', h('span', { class: 'tabular-nums', text: fmtNum(m.fees || 0) })],
      ['gmm.refunds', m.refunds ? h('span', { class: 'tabular-nums', text: fmtNum(m.refunds) }) : null],
      ['gmm.actions', h('span', { class: 'tabular-nums', text: fmtNum(m.actions || 0) })],
      ['gmm.forcedBy', m.forcedBy ? h('span', { 'data-gm-forced': '', text: m.forcedBy }) : null],
      ['gmm.note', m.note ? h('span', { dir: 'auto', text: m.note }) : null],
      ['gmm.created', dateCell(m.createdAt)],
      ['gmm.started', m.startedAt ? dateCell(m.startedAt) : null],
      ['gmm.ended', m.endedAt ? dateCell(m.endedAt) : null]
    ])));
    var rows = (m.entries || []).map(function (e) {
      var fee = e.fee && e.fee.amount ? fmtNum(e.fee.amount) + ' ' + keyName('currency', e.fee.currencyKey) : '';
      var won = e.amount && e.amount.amount ? fmtNum(e.amount.amount) + ' ' + keyName('currency', e.amount.currencyKey) : '';
      return h('div', { class: 'flex flex-wrap items-center gap-2 py-2 border-b border-secondary-200 dark:border-secondary-700 text-sm', 'data-gm-entry': e.userId },
        e.rank ? h('span', { class: 'tabular-nums text-secondary-500 w-6', text: localDigits(e.rank) }) : null,
        playerLink(e.userId, e.username, 'font-medium'), e.bot ? tag(t('gmm.bot'), 'secondary') : null,
        tag(t('gmm.entry.' + e.status), e.status === 'refunded' ? 'warning' : 'secondary'),
        e.result ? tag(t('gmm.result.' + e.result), e.result === 'won' ? 'success' : e.result === 'forfeit' ? 'error' : 'secondary') : null,
        e.released ? tag(t('gmm.released'), 'secondary') : null,
        fee ? h('span', { class: 'tabular-nums text-secondary-500', text: t('gmm.fee', { v: fee }) }) : null,
        won ? h('span', { class: 'tabular-nums text-success-600', text: t('gmm.won', { v: won }) }) : null,
        e.prizeStatus && e.prizeStatus !== 'none' ? tag(t('gmm.prize.' + e.prizeStatus), 'primary') : null,
        e.error ? h('span', { class: 'text-error-600 dark:text-error-400', text: e.error }) : null);
    });
    grid.appendChild(section('gmm.entries', withKeyNames(function () { return rows.length ? h('div', {}, rows) : h('p', { class: 'text-sm text-secondary-500', text: t('gmm.entries.none') }); })));
    if (m.results) grid.appendChild(card(h('div', { class: 'p-4' }, h('h2', { class: 'text-lg font-medium mb-3', text: t('gmm.results') }), jsonBlock(m.results)), 'eco-span-all'));
    grid.appendChild(timeline(m.id));
  }
  function endDialog(m) {
    var seated = (m.entries || []).filter(function (e) { return e.status === 'paid' || e.status === 'free'; });
    var vals = { winner: seated[0] ? seated[0].userId : null, reason: '' };
    var sel = selectInput(vals.winner, seated.map(function (e) { return { value: e.userId, label: (e.username || e.userId) + (e.bot ? ' · ' + t('gmm.bot') : '') }; }), function (v) { vals.winner = v; }, { placeholder: false, label: t('gmm.end.winner') });
    sel.setAttribute('data-gm-winner', '');
    var reason = textInput('', { label: t('gmm.reason'), oninput: function (v) { vals.reason = v; } });
    reason.setAttribute('data-gm-reason', '');
    confirmModal({ title: t('gmm.end'), text: t('gmm.end.text'), icon: 'prizes', ok: t('gmm.end.ok'), body: h('div', { class: 'mt-3 flex flex-col gap-2' }, field(t('gmm.end.winner'), sel), field(t('gmm.reason'), reason, null, { optional: true })) }).then(function (ok) {
      if (!ok) return;
      api('POST', '/game-matches/' + encodeURIComponent(m.id) + '/end', { winnerUserId: vals.winner, reason: vals.reason.trim() }).then(function () {
        toast('success', t('gmm.end.done'));
        setTimeout(load, 400);
      }, function (err) { toast('error', errorText(err)); });
    });
  }
  function cancelDialog(m) {
    var vals = { reason: '' };
    var reason = textInput('', { label: t('gmm.reason'), oninput: function (v) { vals.reason = v; } });
    confirmModal({ title: t('gmm.cancel'), text: t(m.status === 'started' ? 'gmm.cancel.started' : 'gmm.cancel.open'), danger: true, ok: t('gmm.cancel.ok'), body: h('div', { class: 'mt-3' }, field(t('gmm.reason'), reason, null, { optional: true })) }).then(function (ok) {
      if (!ok) return;
      api('POST', '/game-matches/' + encodeURIComponent(m.id) + '/cancel', { reason: vals.reason.trim() }).then(function (res) {
        toast('success', t('gmm.cancel.done', { n: fmtNum(res.refunded || 0) }));
        setTimeout(load, 400);
      }, function (err) { toast('error', errorText(err)); });
    });
  }
  load();
}

function liveDialog(id) {
  var body = h('div', { 'data-gm-live': id }, bar('60%'));
  var close;
  function load() {
    api('GET', '/game-matches/' + encodeURIComponent(id) + '/live').then(function (res) { clear(body).appendChild(jsonBlock(res)); }, function (err) { clear(body).appendChild(notice('error', errorText(err))); });
  }
  var again = button({ variant: 'ghost', size: 'footer', icon: 'refresh', label: t('refresh'), onclick: load });
  close = panelOpen(t('gmm.live'), body, h('div', { class: 'flex gap-4 w-full items-center justify-end' }, again, button({ variant: 'primary', size: 'footer', label: t('close'), onclick: function () { close(); } })));
  load();
}

function timeline(id) {
  var list = h('ol', { class: 'flex flex-col', 'data-gm-timeline': id });
  var more = button({ variant: 'link', size: 'sm', label: t('gmm.more'), onclick: function () { load(); } });
  more.setAttribute('data-gm-more', '');
  var count = h('span', { class: 'text-sm text-secondary-500 tabular-nums' });
  var after = 0;
  function load() {
    setBusy(more, true);
    api('GET', '/game-matches/' + encodeURIComponent(id) + '/actions?after=' + after + '&limit=200').then(function (res) {
      setBusy(more, false);
      (res.items || []).forEach(function (a) {
        after = Math.max(after, a.seq);
        list.appendChild(h('li', { class: 'flex flex-wrap items-start gap-2 py-1.5 border-b border-secondary-200 dark:border-secondary-700 text-sm', 'data-gm-action-row': a.kind },
          h('span', { class: 'tabular-nums text-secondary-500 w-10', text: localDigits(a.seq) }),
          tag(has('gmm.kind.' + a.kind) ? t('gmm.kind.' + a.kind) : a.kind, 'secondary'),
          a.userId ? playerLink(a.userId, a.username) : null,
          a.data && Object.keys(a.data).length ? h('span', { class: 'font-mono text-xs text-secondary-600 dark:text-secondary-300 break-all min-w-0 flex-1', dir: 'ltr', text: JSON.stringify(a.data) }) : h('span', { class: 'flex-1' }),
          h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: fmtDate(a.at, true) })));
      });
      count.textContent = t('gmm.timeline.count', { n: fmtNum(after), total: fmtNum(res.total || 0) });
      more.hidden = after >= (res.total || 0);
      if (!list.children.length) list.appendChild(h('li', { class: 'text-sm text-secondary-500', text: t('gmm.timeline.none') }));
    }, function (err) { setBusy(more, false); toast('error', errorText(err)); });
  }
  load();
  return card(h('div', { class: 'p-4' }, h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, h('h2', { class: 'text-lg font-medium', text: t('gmm.timeline') }), count), list, h('div', { class: 'mt-2' }, more)), 'eco-span-all');
}

function roomPlayEditor(cur, change, row) {
  var vals = Object.assign({ players: null, humanSeats: null, botFillSeconds: null, botTemplate: '', botsPay: false, payout: 'prize', payoutPercent: null, loserPrizeId: null, winnerXp: null, loserXp: null,
    overdraft: false, newPlayersOnly: false, replaces: '', disabled: false }, cur || {});
  var summary = h('span', { class: 'text-sm text-secondary-700 dark:text-secondary-200 min-w-0', 'data-gm-play-summary': '' });
  var panel = h('div', { class: 'mt-3 pt-3 border-t border-secondary-200 dark:border-secondary-700', hidden: true, 'data-gm-play-panel': '' });
  var form = null;
  function text() {
    var p = [];
    if (vals.players) p.push(t('gm.play.sum.players', { n: fmtNum(vals.players) }));
    if (vals.humanSeats) p.push(t('gm.play.sum.humans', { n: fmtNum(vals.humanSeats) }));
    if (vals.botFillSeconds) p.push(t('gm.play.sum.fill', { n: fmtNum(vals.botFillSeconds) }));
    if (vals.payout === 'pot') p.push(t('gm.play.sum.pot', { n: fmtNum(vals.payoutPercent || 0) }));
    if (vals.botsPay) p.push(t('gm.play.sum.botsPay'));
    if (vals.overdraft) p.push(t('gm.play.sum.overdraft'));
    if (vals.newPlayersOnly) p.push(t('gm.play.sum.np', { room: vals.replaces || t('none') }));
    if (vals.disabled) p.push(t('gm.play.sum.off'));
    return p.length ? p.join(' · ') : t('gm.play.sum.none');
  }
  function paint() {
    summary.textContent = text();
    toggleBtn.querySelector('[data-label]').textContent = t(panel.hidden ? 'gm.req.edit' : 'gm.req.close');
  }
  function emit() { change(Object.assign({}, vals)); paint(); }
  var toggleBtn = h('button', { type: 'button', class: 'inline-flex items-center gap-1 text-sm font-medium text-primary-700 dark:text-primary-300 hover:underline shrink-0 min-h-9 px-1', 'data-gm-play-toggle': '',
    onclick: function () {
      if (!form) {
        form = buildForm([
          { type: 'row', fields: [
            { name: 'players', type: 'number', label: 'f.gm.play.players', hint: 'f.gm.play.players.hint', optional: true, min: 2, max: 100 },
            { name: 'humanSeats', type: 'number', label: 'f.gm.play.humans', hint: 'f.gm.play.humans.hint', optional: true, min: 1, max: 100 }
          ] },
          { type: 'row', fields: [
            { name: 'botFillSeconds', type: 'number', label: 'f.gm.play.fill', hint: 'f.gm.play.fill.hint', optional: true, min: 0, max: 600 },
            { name: 'botTemplate', type: 'text', ltr: true, label: 'f.gm.play.template', hint: 'f.gm.play.template.hint', optional: true }
          ] },
          { name: 'botsPay', type: 'bool', inline: true, label: 'f.gm.play.botsPay', hint: 'f.gm.play.botsPay.hint' },
          { type: 'row', fields: [
            { name: 'payout', type: 'select', noPlaceholder: true, label: 'f.gm.play.payout', options: function () { return [{ value: 'prize', label: t('gm.play.payout.prize') }, { value: 'pot', label: t('gm.play.payout.pot') }]; } },
            { name: 'payoutPercent', type: 'number', label: 'f.gm.play.percent', hint: 'f.gm.play.percent.hint', required: true, min: 0, max: 100, when: function (v) { return v.payout === 'pot'; } }
          ] },
          { name: 'loserPrizeId', type: 'ref', resource: 'prizes', label: 'f.gm.play.loserPrize', hint: 'f.gm.play.loserPrize.hint', optional: true },
          { type: 'row', fields: [
            { name: 'winnerXp', type: 'number', label: 'f.gm.play.winnerXp', optional: true, min: 0, max: 1000000 },
            { name: 'loserXp', type: 'number', label: 'f.gm.play.loserXp', optional: true, min: 0, max: 1000000 }
          ] },
          { name: 'overdraft', type: 'bool', inline: true, label: 'f.gm.play.overdraft', hint: 'f.gm.play.overdraft.hint' },
          { name: 'newPlayersOnly', type: 'bool', inline: true, label: 'f.gm.play.np', hint: 'f.gm.play.np.hint' },
          { name: 'replaces', type: 'text', ltr: true, label: 'f.gm.play.replaces', hint: 'f.gm.play.replaces.hint', required: true, when: function (v) { return v.newPlayersOnly; } },
          { name: 'disabled', type: 'bool', inline: true, label: 'f.gm.play.disabled', hint: 'f.gm.play.disabled.hint' }
        ], vals, { onchange: emit });
        panel.appendChild(form.el);
        panel.addEventListener('input', function () { setTimeout(emit, 0); });
        panel.addEventListener('change', function () { setTimeout(emit, 0); });
        panel.addEventListener('click', function () { setTimeout(emit, 0); });
      }
      panel.hidden = !panel.hidden;
      paint();
    } }, icon('gear', 'h-4 w-4'), h('span', { 'data-label': '' }));
  var box = h('div', { class: 'border border-dashed border-secondary-300 dark:border-secondary-600 rounded-md px-3 py-2', 'data-gm-play': '' },
    h('div', { class: 'flex flex-wrap items-center justify-between gap-2' }, h('div', { class: 'flex items-center gap-2 min-w-0' }, icon('bots', 'h-4 w-4 shrink-0 text-secondary-500'), summary), toggleBtn), panel);
  paint();
  box.validate = function () {
    if (!form) return null;
    if (!form.validate()) { panel.hidden = false; paint(); return t('err.form'); }
    return null;
  };
  return box;
}

function roomPlayFrom(r) {
  return { players: r.players || null, humanSeats: r.humanSeats || null, botFillSeconds: r.botFillSeconds || null, botTemplate: r.botTemplate || '', botsPay: !!r.botsPay,
    payout: r.payout === 'pot' ? 'pot' : 'prize', payoutPercent: r.payoutPercent == null ? null : r.payoutPercent, loserPrizeId: r.loserPrizeId || null, winnerXp: r.winnerXp || null,
    loserXp: r.loserXp || null, overdraft: !!r.overdraft, newPlayersOnly: !!r.newPlayersOnly, replaces: r.replaces || '', disabled: !!r.disabled };
}

function roomPlayTo(p) {
  p = p || {};
  var num = function (v) { return v == null || v === '' ? null : Number(v); };
  return { players: num(p.players), humanSeats: num(p.humanSeats), botFillSeconds: num(p.botFillSeconds), botTemplate: p.botTemplate || '', botsPay: !!p.botsPay,
    payout: p.payout === 'pot' ? 'pot' : 'prize', payoutPercent: p.payout === 'pot' ? num(p.payoutPercent) : null, loserPrizeId: p.loserPrizeId || '', winnerXp: num(p.winnerXp), loserXp: num(p.loserXp),
    overdraft: !!p.overdraft, newPlayersOnly: !!p.newPlayersOnly, replaces: p.newPlayersOnly ? (p.replaces || '') : '', disabled: !!p.disabled };
}

(function () {
  var g = RESOURCES.games;
  var rooms = g.fields.filter(function (f) { return f.name === 'rooms'; })[0];
  rooms.columns.push({ name: 'play', type: 'custom', label: 'f.gm.room.play', optional: true, wide: true, build: roomPlayEditor });
  var bots = g.fields.filter(function (f) { return f.name === 'botsAllowed'; })[0];
  g.fields.splice(g.fields.indexOf(bots) + 1, 0, { name: 'oneMatchPerPlayer', type: 'bool', inline: true, label: 'f.gm.oneMatch', hint: 'f.gm.oneMatch.hint' });
  var defaults = g.defaults, toForm = g.toForm, fixPayload = g.fixPayload, newRow = rooms.newRow;
  rooms.newRow = function (rows) { var r = newRow(rows); r.play = roomPlayFrom({}); return r; };
  g.defaults = function (q) { var v = defaults(q); v.oneMatchPerPlayer = false; return v; };
  g.toForm = function (v) {
    var raw = (v.rooms || []).slice();
    var out = toForm(v);
    out.rooms.forEach(function (r, i) { r.play = roomPlayFrom(raw[i] || {}); });
    out.oneMatchPerPlayer = !!v.oneMatchPerPlayer;
    return out;
  };
  g.fixPayload = function (p) {
    var plays = (p.rooms || []).map(function (r) { return r.play; });
    var out = fixPayload(p);
    out.rooms = out.rooms.map(function (r, i) { return Object.assign(r, roomPlayTo(plays[i])); });
    out.oneMatchPerPlayer = !!p.oneMatchPerPlayer;
    return out;
  };
})();
