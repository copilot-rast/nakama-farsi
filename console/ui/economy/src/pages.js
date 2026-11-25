var SECTIONS = {
  economy: ['overview', 'currencies', 'metas', 'meta-relations', 'prizes', 'products', 'offers', 'entry-offers', 'offer-preview', 'offer-report', 'offer-settings', 'tag-discounts', 'tag-discount-preview', 'tag-discount-settings', 'friend-gifts', 'friend-gift-rules', 'friend-gift-settings', 'friend-gift-report', 'items', 'avatar-categories', 'avatar-items', 'avatar-types', 'avatar-settings', 'achievements', 'chance-games', 'prize-groups', 'games', 'gifts', 'event-types', 'event-rules', 'meta-rules', 'level-tables', 'streaks', 'tasks', 'task-groups', 'event-log', 'referrals', 'payments', 'payment-settings'],
  publication: ['faqs', 'news', 'news-review', 'news-sources', 'news-ingest', 'news-reactions', 'news-settings', 'banners', 'banner-lists', 'sponsors', 'comments', 'notifications', 'auto-rules', 'notify-categories', 'remote-config'],
  moderation: ['players', 'reports', 'player-reports', 'restrictions', 'chat-words', 'chat-quota', 'chat-quota-settings', 'chat-quota-report', 'sticker-packs', 'sticker-settings', 'comm-settings', 'comm-channels', 'voice-rooms', 'activity'],
  support: ['tickets', 'live-chat'],
  system: ['updates', 'server-rules', 'content-packs', 'content-settings', 'insights-overview', 'insights-sessions', 'heatmaps', 'insights-settings', 'sdk', 'sample-data', 'config-transfer']
};
var SECTION_ORDER = ['economy', 'publication', 'moderation', 'support', 'system'];
var SECTION_LABEL = { economy: 'section', publication: 'section.pub', moderation: 'section.mod', support: 'section.sup', system: 'section.sys' };
var PAGE_ACL = { 'server-rules': 'SETTINGS', 'sample-data': 'SETTINGS', 'config-transfer': 'SETTINGS', players: 'ACCOUNT', restrictions: 'ACCOUNT', reports: 'CHANNEL_MESSAGE', 'player-reports': 'CHANNEL_MESSAGE', 'chat-words': 'CHANNEL_MESSAGE', 'chat-quota': 'CHANNEL_MESSAGE', 'chat-quota-settings': 'CHANNEL_MESSAGE', 'chat-quota-report': 'CHANNEL_MESSAGE', 'sticker-packs': 'CHANNEL_MESSAGE', 'sticker-settings': 'CHANNEL_MESSAGE', 'comm-settings': 'CHANNEL_MESSAGE', 'comm-channels': 'CHANNEL_MESSAGE', 'voice-rooms': 'CHANNEL_MESSAGE', activity: 'ACCOUNT', tickets: 'ACCOUNT_NOTES', 'live-chat': 'ACCOUNT_NOTES', 'remote-config': 'SETTINGS' };
var SECTION_ACL = { economy: 'HIRO_ECONOMY', publication: 'NOTIFICATION', system: 'SETTINGS' };
var HUBS = {
  metas: { section: 'economy', tabs: [['metas', ''], ['meta-rules', 'rules'], ['meta-relations', 'relations'], ['level-tables', 'levels']] },
  tasks: { section: 'economy', tabs: [['tasks', ''], ['task-groups', 'groups']] },
  offers: { section: 'economy', tabs: [['offers', ''], ['entry-offers', 'entry'], ['offer-preview', 'preview'], ['offer-report', 'report'], ['offer-settings', 'settings']] },
  'tag-discounts': { section: 'economy', tabs: [['tag-discounts', ''], ['tag-discount-preview', 'preview'], ['tag-discount-settings', 'settings']] },
  'friend-gifts': { section: 'economy', tabs: [['friend-gifts', ''], ['friend-gift-rules', 'rules'], ['friend-gift-settings', 'settings'], ['friend-gift-report', 'report']] },
  prizes: { section: 'economy', tabs: [['prizes', ''], ['prize-groups', 'groups']] },
  payments: { section: 'economy', tabs: [['payments', ''], ['payment-settings', 'settings']] },
  'chat-quota': { section: 'moderation', tabs: [['chat-quota', ''], ['chat-quota-settings', 'settings'], ['chat-quota-report', 'report']] },
  avatar: { section: 'economy', tabs: [['avatar-categories', ''], ['avatar-items', 'items'], ['avatar-types', 'types'], ['avatar-settings', 'settings']] },
  stickers: { section: 'moderation', tabs: [['sticker-packs', ''], ['sticker-settings', 'settings'], ['comm-settings', 'match-chat'], ['comm-channels', 'channels']] },
  events: { section: 'economy', tabs: [['event-types', ''], ['event-rules', 'prizes'], ['event-log', 'report']] },
  news: { section: 'publication', tabs: [['news', ''], ['news-review', 'review'], ['news-sources', 'sources'], ['news-ingest', 'log'], ['news-reactions', 'reactions'], ['news-settings', 'settings']] },
  banners: { section: 'publication', tabs: [['banners', ''], ['banner-lists', 'placements']] },
  notifications: { section: 'publication', tabs: [['notifications', ''], ['notify-categories', 'categories']] },
  content: { section: 'system', tabs: [['content-packs', ''], ['content-settings', 'settings']] },
  insights: { section: 'system', tabs: [['insights-overview', ''], ['insights-sessions', 'sessions'], ['heatmaps', 'heatmaps'], ['insights-settings', 'settings']] }
};
var NAV = {
  economy: ['overview', { id: 'progress', items: ['currencies', 'metas', 'tasks', 'achievements', 'streaks'] }, { id: 'sales', items: ['products', 'offers', 'tag-discounts', 'items', 'avatar', 'gifts', 'friend-gifts', 'payments'] }, { id: 'play', items: ['prizes', 'games', 'chance-games'] }, { id: 'growth', items: ['events', 'referrals'] }],
  publication: [{ id: 'content', items: ['faqs', 'news', 'comments'] }, { id: 'ads', items: ['banners', 'sponsors'] }, { id: 'messages', items: ['notifications', 'auto-rules'] }, 'remote-config'],
  moderation: [{ id: 'people', items: ['players', 'restrictions', 'activity'] }, { id: 'chat', items: ['reports', 'player-reports', 'chat-words', 'chat-quota', 'stickers', 'voice-rooms'] }],
  support: ['tickets', 'live-chat'],
  system: ['updates', 'server-rules', 'content', 'insights', 'sdk', 'sample-data', 'config-transfer']
};
var HUB_OF = {};
Object.keys(HUBS).forEach(function (hub) { HUBS[hub].tabs.forEach(function (tab) { HUB_OF[tab[0]] = hub; }); });
var hubLast = {};
var CUSTOM_PAGES = {};
var LIMIT = 20;
var lastQuery = { comments: { status: 'pending' }, reports: { status: 'open' }, 'player-reports': { status: 'active' }, tickets: { status: 'waiting' } };

function sectionOf(page) {
  if (psSection(page)) return PS_ROUTE;
  if (SECTIONS.support.indexOf(page) !== -1) return 'support';
  if (SECTIONS.system.indexOf(page) !== -1) return 'system';
  if (SECTIONS.moderation.indexOf(page) !== -1) return 'moderation';
  return SECTIONS.publication.indexOf(page) !== -1 ? 'publication' : 'economy';
}

function apiFor(page) {
  var def = RESOURCES[page] || LISTS[page];
  if (def && def.api === 'base') return baseApi;
  if (def && def.api === 'publication') return pubApi;
  var sec = sectionOf(page);
  return sec === 'publication' ? pubApi : api;
}

function pageAllowed(page) {
  var need = PAGE_ACL[page] || SECTION_ACL[sectionOf(page)];
  return !need || aclCan(need, 0);
}

function hubTabs(hub) {
  return HUBS[hub] ? HUBS[hub].tabs.map(function (tab) { return tab[0]; }).filter(pageAllowed) : [];
}

function tabLabel(page) {
  return t(has('tab.' + page) ? 'tab.' + page : 'nav.' + page);
}

function routeBase(page) {
  var hub = HUB_OF[page];
  if (!hub) return page;
  var slug = HUBS[hub].tabs.filter(function (tab) { return tab[0] === page; })[0][1];
  return hub + (slug ? '/' + slug : '');
}

function hashFor(page, id, query) {
  return '#/' + sectionOf(page) + '/' + routeBase(page) + (id ? '/' + encodeURIComponent(id) : '') + qs(query || {});
}

function go(page, id, query) {
  location.hash = hashFor(page, id, query);
}

function errorBlock(err, retry) {
  return h('div', { class: 'py-6' }, notice('error', errorText(err), retry ? button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: retry }) : null));
}

function listPage(container, name, query, ctx) {
  var def = RESOURCES[name] || LISTS[name];
  var call = apiFor(name);
  var columns = def.columns.map(function (c) { return { label: t(c.label), width: c.width, align: c.align, skeleton: c.skeleton, head: c.head }; });
  var state = { q: query.q || '', page: Math.max(1, parseInt(query.page, 10) || 1), limit: LIMIT, filters: {}, total: 0, loading: true, loaded: false };
  (def.filters || []).forEach(function (f) { if (query[f.name]) state.filters[f.name] = query[f.name]; });
  lastQuery[name] = currentQuery();

  var create = def.fields ? button({ variant: 'primary', size: 'page', icon: 'plus', label: t('new.' + name), onclick: function () { go(name, 'new', currentQuery()); } }) : def.create ? def.create() : null;
  var frame = def.settings || sectionOf(name) === PS_ROUTE ? tabFrame(t('desc.' + name), create) : pageFrame(t('nav.' + name), t('desc.' + name), create);
  var pagerBottom = h('div', { class: 'eco-pager-slot' });
  var countBox = h('div', { class: 'eco-count-slot' });
  var tableBox = h('div', {});
  var muted = false;
  function changed() {
    if (muted) return;
    state.page = 1;
    paintClear();
    sync();
    load();
  }
  var search = searchInput(state.q, t(has('search.' + name) ? 'search.' + name : 'search'), debounce(function (e) {
    state.q = toLatinDigits(e.target.value).trim();
    changed();
  }, 300));
  var resets = [];
  var filters = (def.filters || []).map(function (f) {
    var opts2 = typeof f.options === 'function' ? f.options() : f.options;
    if (f.type === 'ref') {
      var picker = refPicker({ resource: f.resource, api: typeof f.api === 'function' ? f.api : f.api === 'publication' ? pubApi : api, filter: f.filter, value: state.filters[f.name] || '', label: t(f.label), placeholder: t('all'), onchange: function (v) {
        state.filters[f.name] = v || '';
        box.mark(v);
        changed();
      } });
      var box = filterBox(t(f.label), picker, !!state.filters[f.name]);
      box.classList.add('eco-filter-ref');
      box.setAttribute('data-filter', f.name);
      resets.push(function () { var x = picker.querySelector('.eco-ref-clear'); if (x && !x.hidden) x.click(); });
      return box;
    }
    if (f.type === 'text') {
      var tb = filterText(t(f.label), state.filters[f.name], function (v) { state.filters[f.name] = v; changed(); });
      tb.setAttribute('data-filter', f.name);
      resets.push(function () { tb.input.value = ''; tb.mark(''); });
      return tb;
    }
    var sb = filterSelect(t(f.label), state.filters[f.name], opts2, function (v) { state.filters[f.name] = v; changed(); });
    sb.setAttribute('data-filter', f.name);
    resets.push(function () { sb.select.value = ''; sb.mark(''); });
    return sb;
  });
  var clearBtn = filters.length ? clearFiltersButton(function () {
    muted = true;
    resets.forEach(function (r) { r(); });
    muted = false;
    state.filters = {};
    changed();
    search.querySelector('input').focus();
  }) : null;
  function paintClear() {
    if (clearBtn) clearBtn.hidden = !Object.keys(state.filters).some(function (k) { return state.filters[k]; });
  }
  paintClear();
  var toolbar = listToolbar({ search: search, filters: filters, clear: clearBtn, meta: countBox, refresh: refreshButton(function () { load(); }) });
  if (def.top) frame.content.appendChild(def.top());
  frame.content.appendChild(h('div', { class: 'eco-list' },
    toolbar,
    h('div', { class: 'eco-list-sel' }),
    tableBox,
    pagerBottom));
  clear(container).appendChild(frame.el);

  function currentQuery() {
    var q = Object.assign({}, state.filters);
    Object.keys(q).forEach(function (k) { if (!q[k]) delete q[k]; });
    if (state.q) q.q = state.q;
    if (state.page > 1) q.page = state.page;
    return q;
  }
  function sync() {
    lastQuery[name] = currentQuery();
    ctx.query = lastQuery[name];
    replaceHash(hashFor(name, ctx.id, lastQuery[name]));
  }
  function paintPager() {
    clear(countBox).appendChild(listCount(state));
    clear(pagerBottom).appendChild(pager(state, function (p) {
      state.page = p;
      sync();
      load();
      var top = toolbar.getBoundingClientRect().top;
      if (top < 0) toolbar.scrollIntoView({ block: 'start' });
    }));
  }
  function open(item) {
    if (def.onRow) { def.onRow(item, currentQuery()); return; }
    if (def.rowHref) { location.hash = def.rowHref(item); return; }
    go(name, item.id, currentQuery());
  }
  var seq = 0;
  function load() {
    var my = ++seq;
    state.loading = true;
    if (!state.loaded) {
      clear(tableBox).appendChild(table(columns, skeletonRows(def.columns, 8)));
    } else {
      tableBox.classList.add('opacity-60');
    }
    paintPager();
    var params = Object.assign({ q: state.q, page: state.page, limit: state.limit }, state.filters);
    return call('GET', '/' + (def.path || name) + qs(params)).then(function (res) {
      if (my !== seq) return;
      state.loading = false;
      state.loaded = true;
      state.total = (res && res.total) || 0;
      var items = (res && res.items) || [];
      if (def.onItems) def.onItems(items);
      tableBox.classList.remove('opacity-60');
      if (!items.length && state.page > 1 && state.total > 0) {
        state.page = Math.max(1, Math.ceil(state.total / state.limit));
        sync();
        return load();
      }
      clear(tableBox);
      if (!items.length) {
        var filtered = state.q || Object.keys(state.filters).some(function (k) { return state.filters[k]; });
        var reset = filtered && clearBtn && !clearBtn.hidden ? button({ variant: 'ghost', size: 'sm', icon: 'close', label: t('filters.clear'), onclick: function () { clearBtn.click(); } }) : null;
        tableBox.appendChild(card(emptyBlock(filtered ? t('empty.search') : t('empty.title'), filtered ? null : t('empty.hint.' + name),
          reset || (!filtered && def.fields ? button({ variant: 'primary', size: 'page', icon: 'plus', label: t('new.' + name), onclick: function () { go(name, 'new'); } }) : !filtered && def.create ? def.create() : null))));
      } else {
        tableBox.appendChild(table(columns, items.map(function (it) {
          return { item: it, cells: def.columns.map(function (c) { return c.render(it); }) };
        }), { onrow: def.noRow ? null : open }));
      }
      paintPager();
      if (def.rendered) def.rendered();
    }, function (err) {
      if (my !== seq) return;
      state.loading = false;
      tableBox.classList.remove('opacity-60');
      clear(tableBox).appendChild(errorBlock(err, load));
      paintPager();
    });
  }
  load();
  if (def.mounted) def.mounted(query);
  return { reload: load, name: name, selSlot: frame.content.querySelector('.eco-list-sel'), toolbar: toolbar };
}

function formSkeleton(def) {
  var box = h('div', {});
  var count = 0;
  def.fields.forEach(function walk(f) {
    if (count > 7) return;
    if (f.type === 'row') { box.appendChild(h('div', { class: 'eco-form-row' }, f.fields.map(function () { return skeletonField(); }))); count += 1; return; }
    if (f.type === 'heading') return;
    box.appendChild(skeletonField(f.type === 'textarea' || f.type === 'json' || f.type === 'items' ? 'h-24' : 'h-10'));
    count += 1;
  });
  return box;
}

function skeletonField(height) {
  return h('div', { class: 'mb-5' }, bar('7rem', 'h-3 mb-2'), bar('100%', height || 'h-10'));
}

var panelState = null;

function closePanel() {
  if (panelState) {
    panelState.el.remove();
    document.removeEventListener('keydown', panelState.esc, true);
    panelState = null;
  }
}

function openEditor(name, id, query, onSaved) {
  var def = RESOURCES[name];
  if (!def) return;
  if (panelState && panelState.key === name + '/' + id) return;
  closePanel();
  var isNew = id === 'new';
  var title = t((isNew ? 'new.' : 'edit.') + name);
  var body = h('div', {});
  var footer = h('div', { class: 'flex gap-4 w-full items-center' });
  var back = function () { go(name, null, lastQuery[name] || {}); };
  var el = sidePanel(title, body, footer, back);
  var esc = function (e) { if (e.key === 'Escape' && panelState && !document.querySelector('[data-economy="panel"] .eco-pop:not([hidden]), [data-economy="panel"] .eco-qc-pop:not([hidden])')) { e.preventDefault(); back(); } };
  document.addEventListener('keydown', esc, true);
  panelState = { el: el, key: name + '/' + id, esc: esc };
  document.body.appendChild(el);
  body.appendChild(formSkeleton(def));

  var call = apiFor(name);
  var item = isNew ? Promise.resolve(def.defaults(query)) : call('GET', '/' + (def.path || name) + '/' + encodeURIComponent(id));
  Promise.all([item, freshLanguages(), loadKeyTypes(), def.prepare ? def.prepare() : null]).then(function (all) { return all[0]; }).then(function (data) {
    if (!panelState || panelState.el !== el) return;
    var values = clone(data) || {};
    if (def.toForm) values = def.toForm(values);
    var form = buildForm(def.fields, values);
    var top = h('div', {});
    clear(body).appendChild(h('form', { novalidate: true, onsubmit: function (e) { e.preventDefault(); save(); } }, top, form.el, h('button', { type: 'submit', hidden: true })));
    var saveBtn = button({ variant: 'primary', size: 'footer', type: 'button', label: isNew ? t('create') : t('save'), onclick: function () { save(); } });
    var cancelBtn = button({ variant: 'ghost', size: 'footer', label: t('cancel'), onclick: back });
    var del = null;
    if (!isNew) {
      del = h('div', { class: 'flex items-center gap-2' });
      var askDelete = function () {
        clear(del).appendChild(h('span', { class: 'text-sm text-secondary-700 dark:text-secondary-200', text: t('delete.confirm') }));
        del.appendChild(button({ variant: 'dangerSolid', size: 'sm', label: t('delete.yes'), onclick: doDelete }));
        del.appendChild(button({ variant: 'ghost', size: 'sm', label: t('no'), onclick: resetDelete }));
      };
      var resetDelete = function () { clear(del).appendChild(button({ variant: 'danger', size: 'footer', icon: 'trash', label: t('delete'), onclick: askDelete })); };
      var doDelete = function () {
        call('DELETE', '/' + (def.path || name) + '/' + encodeURIComponent(id)).then(function () {
          toast('success', t('deleted'));
          if (def.afterSave) def.afterSave();
          back();
          if (onSaved) onSaved();
        }, function (err) {
          resetDelete();
          clear(top).appendChild(h('div', { class: 'mb-5' }, notice('error', errorText(err))));
          top.scrollIntoView({ block: 'start' });
        });
      };
      resetDelete();
    }
    clear(footer).appendChild(h('div', { class: 'flex-1 flex items-center' }, del));
    footer.appendChild(cancelBtn);
    footer.appendChild(saveBtn);

    function save() {
      clear(top);
      if (!form.validate()) return;
      var payload = form.payload();
      if (def.fixPayload) payload = def.fixPayload(payload);
      setBusy(saveBtn, true, t('saving'));
      call(isNew ? 'POST' : 'PUT', '/' + (def.path || name) + (isNew ? '' : '/' + encodeURIComponent(id)), payload).then(function () {
        toast('success', t('saved'));
        if (def.afterSave) def.afterSave();
        back();
        if (onSaved) onSaved();
      }, function (err) {
        setBusy(saveBtn, false);
        if (!form.showServerError(err)) {
          clear(top).appendChild(h('div', { class: 'mb-5' }, notice('error', errorText(err))));
          top.scrollIntoView({ block: 'start' });
        }
      });
    }
    var firstInput = body.querySelector('input,textarea,select');
    if (firstInput) firstInput.focus();
  }, function (err) {
    if (!panelState || panelState.el !== el) return;
    clear(body).appendChild(errorBlock(err, function () { closePanel(); openEditor(name, id, query, onSaved); }));
  });
}

var PAYMENT_LIST = {
  filters: [
    { name: 'status', label: 'filter.status', options: opts('status.', ['pending', 'success', 'failed', 'expired']) },
    { name: 'method', label: 'filter.method', options: opts('method.', METHODS) },
    { name: 'reward', label: 'filter.reward', options: opts('pay.rewardf.', ['pending', 'refund', 'settled']) },
    { name: 'userId', label: 'filter.user', type: 'text' }
  ],
  columns: [
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 170 },
    { label: 'col.user', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: it.username || '' }), h('div', { class: 'text-secondary-500 truncate' }, mono(it.userId))); }, skeleton: '50%' },
    { label: 'col.product', render: function (it) { var p = it.product || {}; return h('div', { class: 'min-w-0' }, h('div', { class: 'truncate', text: i18nText(p.name || it.productName) || '' }), h('div', { class: 'text-secondary-500' }, mono(p.key || it.productKey || ''))); }, skeleton: '45%' },
    { label: 'col.method', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('method.' + it.method), 'secondary'), it.sandbox ? tag(t('pay.sandbox'), 'warning') : null); }, width: 150 },
    { label: 'col.amount', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: it.method === 'in_app' ? fmtNum(it.coinAmount) + ' ' + (it.coinCurrency || '') : t('toman', { n: it.finalToman || 0 }) }); }, width: 150 },
    { label: 'col.status', render: function (it) { return paymentStatus(it); }, width: 150 }
  ]
};

function paymentStatus(it) {
  return h('div', { class: 'flex flex-wrap gap-1' },
    tag(t('status.' + it.status), STATUS_COLOR[it.status]),
    it.status === 'success' && !it.rewardGranted && !it.refundStatus ? tag(t('pay.reward.notGranted'), 'error') : null,
    it.refundStatus ? tag(t('pay.refund.' + it.refundStatus), it.refundStatus === 'needed' ? 'warning' : 'secondary') : null);
}

function dl(rows) {
  return h('dl', { class: 'eco-dl' }, rows.filter(function (r) { return r && r[1] !== null && r[1] !== undefined && r[1] !== ''; }).map(function (r) {
    return [h('dt', { class: 'text-secondary-500 dark:text-secondary-400', text: t(r[0]) }), h('dd', { class: 'text-secondary-800 dark:text-secondary-200 min-w-0 break-words' }, r[1])];
  }));
}

function section(titleKey, body, actions) {
  return card(h('div', { class: 'p-4' },
    h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, h('h2', { class: 'text-lg font-medium', text: t(titleKey) }), actions || null),
    body));
}

function jsonBlock(v) {
  if (v === null || v === undefined || v === '') return null;
  var text = typeof v === 'string' ? v : JSON.stringify(v, null, 2);
  return h('pre', { class: 'eco-pre font-mono text-xs bg-secondary-50 dark:bg-secondary-900 border border-secondary-200 dark:border-secondary-700 rounded-md p-2 overflow-x-auto', dir: 'ltr', text: text });
}

var STORE_METHODS = ['bazaar', 'myket', 'google_play', 'apple'];

function lastStoreAnswer(logs) {
  var last = null;
  (logs || []).forEach(function (l) { if (l.action === 'store_validate') last = l; });
  if (!last) return null;
  var r = last.response || {};
  var err = r.error;
  var desc = r.error_description || r.translatedMessage || r.errorMessage || '';
  if (err && typeof err === 'object') { desc = err.message || ''; err = err.status || ''; }
  if (!err) err = r.messageCode || r.errorCode || '';
  return storeAnswerLine({ httpStatus: last.httpStatus, storeError: err ? String(err) : '', storeDescription: desc ? String(desc) : '' }) || (last.httpStatus ? tag(localDigits(String(last.httpStatus)), 'secondary') : null);
}

function paymentDetail(container, id) {
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('payments', null, lastQuery.payments || {}); } });
  var frame = pageFrame(t('pay.title'), null, null);
  var titleBox = frame.el.querySelector('h1');
  frame.actions.appendChild(backBtn);
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-detail-grid' });
  frame.content.appendChild(grid);

  function skeleton() {
    clear(grid);
    [5, 6, 4, 4].forEach(function (n) {
      var rows = [];
      for (var i = 0; i < n; i++) rows.push([h('dt', {}, bar('6rem')), h('dd', {}, bar(['70%', '50%', '40%'][i % 3]))]);
      grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), h('dl', { class: 'eco-dl' }, rows))));
    });
    var logRows = [];
    for (var j = 0; j < 3; j++) logRows.push(h('div', { class: 'flex items-center gap-3 py-2' }, bar('5rem'), bar('40%'), bar('4rem')));
    grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), logRows), 'eco-span-all'));
  }

  function render(p) {
    clear(grid);
    clear(titleBox).appendChild(h('span', { text: t('pay.title') + ' ' }));
    titleBox.appendChild(h('span', { class: 'font-mono text-base text-secondary-500', dir: 'ltr', text: String(p.id || id).slice(0, 8) }));
    var user = p.user || {};
    var product = p.product || {};
    var gateway = GATEWAYS.indexOf(p.method) !== -1;
    var actions = h('div', { class: 'flex flex-wrap gap-2' });
    if (p.status === 'pending' && gateway) actions.appendChild(button({ variant: 'primary', size: 'page', icon: 'refresh', label: t('pay.inquire'), onclick: function (e) { act(e.currentTarget, 'inquire', 'pay.inquire.done'); } }));
    if ((p.status === 'pending' || p.status === 'expired') && STORE_METHODS.indexOf(p.method) !== -1) {
      var rb = button({ variant: 'primary', size: 'page', icon: 'refresh', label: t('pay.recheck'), onclick: function (e) { recheck(e.currentTarget); } });
      rb.setAttribute('data-payment-recheck', '');
      actions.appendChild(rb);
    }
    if (p.status === 'success' && !p.rewardGranted && !p.refundStatus) actions.appendChild(button({ variant: 'primary', size: 'page', icon: 'prizes', label: t('pay.retry'), onclick: function (e) { act(e.currentTarget, 'retry-grant', 'pay.retry.done'); } }));
    clear(frame.actions).appendChild(actions);
    frame.actions.appendChild(backBtn);

    grid.appendChild(section('col.status', dl([
      ['col.status', paymentStatus(p)],
      ['col.method', h('div', { class: 'flex flex-wrap gap-1' }, tag(t('method.' + p.method), 'secondary'), p.sandbox ? tag(t('pay.sandbox'), 'warning') : null)],
      ['pay.reward', tag(p.rewardGranted ? t('pay.reward.granted') : t('pay.reward.notGranted'), p.rewardGranted ? 'success' : 'secondary')],
      ['pay.reward.error', p.rewardError ? h('span', { class: 'text-error-600 dark:text-error-400', text: p.rewardError }) : null],
      ['pay.error', p.error ? h('span', { class: 'text-error-600 dark:text-error-400', text: p.error }) : null],
      ['pay.outcome', p.outcome ? mono(p.outcome) : null],
      ['pay.grantAttempts', p.status === 'success' && !p.rewardGranted ? h('span', { class: 'tabular-nums', text: fmtNum(p.grantAttempts || 0) }) : null],
      ['pay.grantRetryAt', p.status === 'success' && !p.rewardGranted && p.grantRetryAt ? dateCell(p.grantRetryAt) : null]
    ])));
    if (p.status === 'success' && !p.rewardGranted) grid.appendChild(refundSection(p));
    grid.appendChild(section('pay.amounts', dl([
      ['pay.original', p.originalToman != null ? h('span', { class: 'tabular-nums', text: t('toman', { n: p.originalToman }) }) : null],
      ['pay.discount', p.discountPercent ? h('span', { text: t('percent', { n: p.discountPercent }) }) : null],
      ['pay.final', p.finalToman != null ? h('span', { class: 'tabular-nums font-semibold', text: t('toman', { n: p.finalToman }) }) : null],
      ['pay.rial', p.amountRial ? h('span', { class: 'tabular-nums', text: fmtNum(p.amountRial) }) : null],
      ['pay.minor', p.amountMinor ? h('span', { class: 'tabular-nums', dir: 'ltr', text: fmtNum(p.amountMinor) + ' ' + String(p.currency || '').toUpperCase() }) : null],
      ['pay.coin', p.coinAmount != null && p.coinCurrency ? h('span', { class: 'tabular-nums', text: fmtNum(p.coinAmount) + ' ' + p.coinCurrency }) : null]
    ])));
    grid.appendChild(section('pay.user', dl([
      ['col.user', h('a', { href: '#/players/' + encodeURIComponent(p.userId || user.id || ''), class: 'eco-link', text: user.username || p.username || p.userId || '' })],
      ['filter.user', mono(p.userId || user.id)],
      ['pay.product', h('div', {}, h('div', { text: i18nText(product.name) || '' }), product.key ? mono(product.key) : null)]
    ])));
    grid.appendChild(section('pay.gateway', dl([
      ['pay.authority', p.authority ? mono(p.authority) : null],
      ['pay.trackId', p.trackId ? mono(p.trackId) : null],
      ['pay.refId', p.refId ? mono(p.refId) : null],
      ['pay.storeToken', p.storeToken ? mono(p.storeToken) : null],
      ['pay.storeAnswer', lastStoreAnswer(p.logs)],
      ['pay.cardMask', p.cardMask ? mono(p.cardMask) : null],
      ['pay.inquiries', h('span', { class: 'tabular-nums', text: fmtNum(p.inquiryCount || 0) }) ],
      ['pay.createdAt', p.createdAt ? dateCell(p.createdAt) : null],
      ['pay.verifiedAt', p.verifiedAt ? dateCell(p.verifiedAt) : null],
      ['pay.inquiredAt', p.inquiredAt ? dateCell(p.inquiredAt) : null],
      ['pay.updatedAt', p.updatedAt ? dateCell(p.updatedAt) : null]
    ])));
    var logs = p.logs || [];
    var logBody = h('div', { class: 'flex flex-col' });
    if (!logs.length) logBody.appendChild(h('div', { class: 'text-sm text-secondary-500 py-2', text: t('pay.logs.empty') }));
    logs.forEach(function (l) {
      var ok = l.httpStatus && l.httpStatus < 400;
      var detail = h('div', { class: 'eco-log-detail mt-2', hidden: true },
        l.url ? h('div', { class: 'mb-2' }, mono(l.url)) : null,
        h('div', { class: 'eco-log-grid' },
          h('div', { class: 'min-w-0' }, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pay.request') }), jsonBlock(l.request) || h('div', { class: 'text-xs text-secondary-400', text: '…' })),
          h('div', { class: 'min-w-0' }, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pay.response') }), jsonBlock(l.response) || h('div', { class: 'text-xs text-secondary-400', text: '…' }))));
      var head = h('button', { type: 'button', class: 'w-full flex flex-wrap items-center gap-2 eco-start', 'aria-expanded': 'false', onclick: function () {
        detail.hidden = !detail.hidden;
        head.setAttribute('aria-expanded', detail.hidden ? 'false' : 'true');
        chev.classList.toggle('rotate-180', !detail.hidden);
      } });
      var chev = icon('chevron', 'h-4 w-4 text-secondary-500 transition');
      head.appendChild(tag(t('log.' + l.action), 'primary'));
      if (l.httpStatus) head.appendChild(tag(localDigits(String(l.httpStatus)), ok ? 'success' : 'error'));
      if (l.durationMs != null) head.appendChild(h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: t('pay.duration', { n: l.durationMs }) }));
      head.appendChild(h('span', { class: 'flex-1' }));
      head.appendChild(h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: fmtDate(l.createdAt) }));
      head.appendChild(chev);
      logBody.appendChild(h('div', { class: 'eco-log py-2 border-b border-secondary-200 dark:border-secondary-700' }, head, detail));
    });
    var sec = section('pay.logs', logBody);
    sec.classList.add('eco-span-all');
    grid.appendChild(sec);
  }

  function refundSection(p) {
    var note = textInput(p.refundNote || '', { multiline: true, rows: 3, label: t('pay.refund.note'), placeholder: t('pay.refund.note.ph') });
    function save(btn, status) {
      setBusy(btn, true);
      api('POST', '/payments/' + encodeURIComponent(id) + '/refund', { status: status, note: note.value }).then(function () {
        toast('success', t('saved'));
        load();
      }, function (err) {
        setBusy(btn, false);
        toast('error', errorText(err));
      });
    }
    var buttons = h('div', { class: 'flex flex-wrap gap-2 mt-3' },
      button({ variant: 'primary', size: 'sm', label: t('pay.refund.markRefunded'), onclick: function (e) { save(e.currentTarget, 'refunded'); } }),
      button({ variant: 'secondary', size: 'sm', label: t('pay.refund.markKept'), onclick: function (e) { save(e.currentTarget, 'kept'); } }));
    return section('pay.refund', h('div', {},
      dl([
        ['pay.refund.status', tag(t('pay.refund.' + (p.refundStatus || 'none')), p.refundStatus === 'needed' ? 'warning' : 'secondary')],
        ['pay.refund.by', p.refundBy ? h('span', { text: p.refundBy }) : null],
        ['pay.refund.at', p.refundAt ? dateCell(p.refundAt) : null]
      ]),
      h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-2', text: t('pay.refund.hint') }),
      field(t('pay.refund.note'), note),
      buttons));
  }

  function recheck(btn) {
    setBusy(btn, true);
    api('POST', '/payments/' + encodeURIComponent(id) + '/recheck', {}).then(function (res) {
      var rc = (res && res.recheck) || {};
      if (rc.error) toast('error', t('pay.recheck.failed', { code: has('code.' + rc.error) ? t('code.' + rc.error) : rc.error }));
      else toast('success', t('pay.recheck.done'));
      render(res);
    }, function (err) {
      setBusy(btn, false);
      toast('error', errorText(err));
    });
  }

  function act(btn, path, doneKey) {
    setBusy(btn, true);
    api('POST', '/payments/' + encodeURIComponent(id) + '/' + path, {}).then(function () {
      toast('success', t(doneKey));
      load();
    }, function (err) {
      setBusy(btn, false);
      toast('error', errorText(err));
    });
  }

  function load() {
    skeleton();
    api('GET', '/payments/' + encodeURIComponent(id)).then(render, function (err) { clear(grid).appendChild(h('div', { class: 'eco-span-all' }, errorBlock(err, load))); });
  }
  load();
}

function isSecretField(f) {
  return f.type === 'secret' || f.type === 'secretJson' || f.type === 'secretText';
}

var STORE_TEST_TONE = { credentials_ok: 'success', unreachable: 'warning', not_configured: 'warning', not_testable: 'secondary' };

function storeReasonText(reason) {
  return reason && has('pm.health.reason.' + reason) ? t('pm.health.reason.' + reason) : '';
}

function storeAnswerLine(rec) {
  if (!rec) return null;
  var parts = h('div', { class: 'flex flex-wrap items-center gap-1.5' });
  if (rec.httpStatus) parts.appendChild(tag(localDigits(String(rec.httpStatus)), rec.httpStatus < 400 || rec.httpStatus === 404 ? 'secondary' : 'error'));
  if (rec.storeError) parts.appendChild(mono(rec.storeError));
  if (rec.storeDescription) parts.appendChild(h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', dir: 'auto', text: rec.storeDescription }));
  return parts.childNodes.length ? parts : null;
}

function storeHealthBlock(m, replace) {
  var hl = m.health || {};
  var box = h('div', { class: 'eco-store-health mt-4 pt-4 border-t border-secondary-200 dark:border-secondary-700', 'data-store-health': m.method });
  box.appendChild(h('div', { class: 'text-sm font-medium mb-1', text: t('pm.health.title') }));
  box.appendChild(h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: t('pm.health.hint') }));
  var resultBox = h('div', { 'data-store-test-result': '', 'aria-live': 'polite' });
  function showResult(r) {
    clear(resultBox);
    if (!r || !r.result) return;
    var tone = STORE_TEST_TONE[r.result] || 'error';
    var wrap = h('div', { class: 'mb-3 flex flex-col gap-1.5' }, notice(tone, t('pm.test.' + r.result)));
    wrap.firstChild.setAttribute('data-result', r.result);
    var line = storeAnswerLine(r);
    var meta = h('div', { class: 'flex flex-wrap items-center gap-1.5 text-xs text-secondary-500 dark:text-secondary-400' },
      r.testedAt ? h('span', { text: t('pm.health.lastTest') + ': ' + fmtDate(r.testedAt) }) : null,
      r.sku ? h('span', { text: t('pm.health.sku') + ':' }) : null, r.sku ? mono(r.sku) : null);
    wrap.appendChild(meta);
    if (line) wrap.appendChild(line);
    if (r.skuSource === 'placeholder') wrap.appendChild(h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('pm.test.placeholder') }));
    resultBox.appendChild(wrap);
  }
  showResult(hl.lastTest);
  box.appendChild(resultBox);

  var rows = [];
  var f = hl.lastFailure;
  rows.push(['pm.health.lastFailure', f ? h('div', { class: 'flex flex-col gap-1', 'data-last-failure': '' },
    h('div', { class: 'flex flex-wrap items-center gap-1.5' }, h('span', { text: fmtDate(f.at) }), storeReasonText(f.reason) ? tag(storeReasonText(f.reason), 'error') : null),
    storeAnswerLine(f)) : h('span', { class: 'text-secondary-500', text: t('pm.health.never') })]);
  var okCell = h('span', {});
  function paintOk(at) {
    clear(okCell);
    okCell.appendChild(at ? h('span', { 'data-last-ok': '', text: fmtDate(at) }) : h('span', { class: 'text-secondary-500', text: t('pm.health.never') }));
  }
  paintOk(hl.lastOkAt);
  rows.push(['pm.health.lastOk', okCell]);
  box.appendChild(dl(rows));

  var sku = textInput('', { ltr: true, mono: true, label: t('pm.health.sku'), placeholder: 'fir_oil_100' });
  sku.setAttribute('data-store-test-sku', '');
  sku.classList.add('text-sm');
  box.appendChild(field(t('pm.health.sku'), sku, t('pm.health.sku.hint'), { optional: true }));
  var actions = h('div', { class: 'flex flex-wrap justify-end gap-2' });
  var testBtn = button({ variant: 'secondary', size: 'sm', icon: 'refresh', label: t('pm.health.test'), onclick: function () {
    setBusy(testBtn, true, t('pm.health.testing'));
    api('POST', '/payment-methods/' + encodeURIComponent(m.method) + '/test', { sku: (sku.value || '').trim() }).then(function (r) {
      setBusy(testBtn, false);
      showResult(r);
      if (r.result === 'credentials_ok') paintOk(r.testedAt);
      toast(r.result === 'credentials_ok' ? 'success' : 'error', t('pm.test.' + r.result));
    }, function (err) {
      setBusy(testBtn, false);
      toast('error', errorText(err));
    });
  } });
  testBtn.setAttribute('data-store-test', m.method);
  actions.appendChild(testBtn);
  var stuck = Number(hl.stuckPayments || 0);
  if (stuck > 0) {
    var recheckBtn = button({ variant: 'primary', size: 'sm', label: t('pm.recheck', { n: fmtNum(stuck) }), onclick: function () {
      confirmModal({ title: t('pm.recheck.title'), text: t('pm.recheck.text'), ok: t('pm.recheck.ok') }).then(function (yes) {
        if (!yes) return;
        setBusy(recheckBtn, true);
        api('POST', '/payment-methods/' + encodeURIComponent(m.method) + '/recheck', {}).then(function (r) {
          toast(r.delivered ? 'success' : 'warning', t('pm.recheck.done', { checked: fmtNum(r.checked || 0), delivered: fmtNum(r.delivered || 0), stillPending: fmtNum(r.stillPending || 0) }));
          return api('GET', '/payment-methods').then(function (list) {
            var next = (Array.isArray(list) ? list : []).filter(function (x) { return x.method === m.method; })[0];
            if (next) replace(next); else setBusy(recheckBtn, false);
          });
        }, function (err) {
          setBusy(recheckBtn, false);
          toast('error', errorText(err));
        });
      });
    } });
    recheckBtn.setAttribute('data-store-recheck', m.method);
    actions.appendChild(recheckBtn);
  }
  box.appendChild(actions);
  return box;
}

function paymentMethodsPage(container) {
  var frame = tabFrame(t('desc.payment-methods'));
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-card-grid' });
  frame.content.appendChild(grid);

  function skeleton() {
    clear(grid);
    for (var i = 0; i < 5; i++) {
      grid.appendChild(card(h('div', { class: 'p-4' },
        h('div', { class: 'flex items-center justify-between mb-4' }, bar('7rem', 'h-4'), bar('2rem', 'h-5 rounded-full')),
        skeletonField(), skeletonField(),
        h('div', { class: 'flex justify-end' }, bar('5rem', 'h-8')))));
    }
  }

  function methodCard(m) {
    var cfg = clone(m.config || {}) || {};
    var active = !!m.isActive;
    var errTop = h('div', {});
    var fieldEls = {};
    var body = h('div', {});
    (m.fields || []).forEach(function (f) {
      var label = has('pm.field.' + f.name) ? t('pm.field.' + f.name) : f.name;
      var ctl;
      if (f.type === 'bool') {
        ctl = toggle(cfg[f.name], label, function (v) { cfg[f.name] = v; }, 'sm');
        ctl.setAttribute('data-field', f.name);
        body.appendChild(h('div', { class: 'mb-4' }, ctl, f.name === m.sandboxField && m.sandboxNeedsKeys && has('pm.hint.sandbox.store') && m.method !== 'stripe' ? h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1.5', text: t('pm.hint.sandbox.store') }) : null));
        fieldEls[f.name] = { setError: function () {} };
        return;
      }
      var secret = isSecretField(f);
      var multi = f.type === 'textarea' || f.type === 'secretJson' || f.type === 'secretText';
      ctl = textInput(cfg[f.name] == null ? '' : cfg[f.name], { ltr: true, mono: true, multiline: multi, rows: f.type === 'secretJson' || f.type === 'secretText' ? 5 : 3, label: label, placeholder: f.default || '', oninput: function (v) { cfg[f.name] = v; } });
      if (secret) { ctl.setAttribute('autocomplete', 'off'); ctl.setAttribute('spellcheck', 'false'); ctl.classList.add('eco-secret'); }
      ctl.setAttribute('data-field', f.name);
      ctl.classList.add('text-sm');
      var help = secret && cfg[f.name] ? t('pm.secret.hint') : has('pm.hint.' + f.name) ? t('pm.hint.' + f.name) : null;
      var wrap = field(label, ctl, help, { optional: !f.required });
      fieldEls[f.name] = wrap;
      body.appendChild(wrap);
    });
    if (!(m.fields || []).length) body.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mb-4', text: t('pm.noFields') }));
    var saveBtn = button({ variant: 'primary', size: 'sm', label: t('save'), onclick: function () {
      clear(errTop);
      Object.keys(fieldEls).forEach(function (k) { fieldEls[k].setError(''); });
      setBusy(saveBtn, true, t('saving'));
      api('PUT', '/payment-methods/' + encodeURIComponent(m.method), { isActive: active, config: cfg }).then(function (res) {
        toast('success', t('saved'));
        if (res && res.method) el.replaceWith(methodCard(res));
        else setBusy(saveBtn, false);
      }, function (err) {
        setBusy(saveBtn, false);
        if (err.field && fieldEls[err.field]) fieldEls[err.field].setError(errorText(err));
        else errTop.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
      });
    } });
    var el = card(h('div', { class: 'p-4 flex flex-col h-full' },
      h('div', { class: 'flex items-start justify-between gap-3 mb-4' },
        h('div', {},
          h('div', { class: 'text-lg font-medium', text: has('method.' + m.method) ? t('method.' + m.method) : m.method }),
          h('div', { class: 'flex items-center gap-1.5 mt-1' }, mono(m.method), tag(m.configured ? t('pm.configured') : t('pm.notConfigured'), m.configured ? 'success' : 'secondary'))),
        toggle(active, t('pm.active'), function (v) { active = v; }, 'sm')),
      errTop,
      h('div', { class: 'flex-1' }, body),
      h('div', { class: 'flex justify-end pt-2' }, saveBtn),
      m.health ? storeHealthBlock(m, function (next) { el.replaceWith(methodCard(next)); }) : null));
    el.setAttribute('data-method', m.method);
    return el;
  }

  function load() {
    skeleton();
    api('GET', '/payment-methods').then(function (res) {
      var list = Array.isArray(res) ? res : (res && (res.items || res.methods)) || [];
      clear(grid);
      if (!list.length) { grid.appendChild(h('div', { class: 'eco-span-all' }, card(emptyBlock(t('pm.empty'))))); return; }
      list.forEach(function (m) { grid.appendChild(methodCard(m)); });
    }, function (err) { clear(grid).appendChild(h('div', { class: 'eco-span-all' }, errorBlock(err, load))); });
  }
  load();
}

var COVERAGE_PAGES = { languages: 'languages', categories: 'categories', tags: 'tags', currencies: 'currencies', metas: 'metas', prizes: 'prizes', products: 'products', gifts: 'gifts', faqs: 'faqs', news: 'news' };

function coverageLabel(resource) {
  return resource === 'languages' ? t('tab.languages') : has('nav.' + resource) ? t('nav.' + resource) : resource;
}

function coveragePanel() {
  var body = h('div', {});
  var box = section('cov.title', h('div', {}, h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mb-4', text: t('cov.desc') }), body),
    button({ variant: 'muted', size: 'icon', icon: 'refresh', iconClass: 'h-3 w-3', title: t('refresh'), onclick: function () { load(); } }));
  box.classList.add('mt-6');
  box.setAttribute('data-coverage', '');
  var seq = 0;
  function cols(langs) {
    return [{ label: t('cov.resource'), skeleton: '55%' }, { label: t('cov.total'), width: 90 }].concat(langs.map(function (code) { return { label: langName(code), width: 120 }; }));
  }
  function skeleton() {
    var langs = LANGS.active.map(function (l) { return l.code; });
    var chips = h('div', { class: 'flex flex-wrap gap-2 mb-4' }, langs.map(function () { return bar('8rem', 'h-6'); }));
    clear(body).appendChild(h('div', {}, chips, table(cols(langs), skeletonRows(cols(langs), 6))));
  }
  function render(c) {
    var langs = (c && c.languages) || [];
    var missing = (c && c.missing) || {};
    var chips = h('div', { class: 'flex flex-wrap gap-2 mb-4' }, langs.map(function (code) {
      var n = missing[code] || 0;
      return h('span', { class: 'eco-tag eco-tag--' + (n ? 'warning' : 'success') + ' inline-flex items-center gap-1.5 px-2 py-1 text-xs rounded-md', 'data-cov-lang': code, 'data-cov-missing': String(n) },
        h('span', { class: 'font-medium', text: langName(code) + (code === c.default ? ' (' + t('tr.default') + ')' : '') }),
        h('span', { class: 'tabular-nums', text: n ? t('cov.missingN', { n: n }) : t('cov.complete') }));
    }));
    var rows = ((c && c.items) || []).map(function (it) {
      var page = COVERAGE_PAGES[it.resource];
      var href = page ? hashFor(page) : null;
      var name = h('div', { class: 'min-w-0' },
        href ? h('a', { href: href, class: 'eco-link font-medium', text: coverageLabel(it.resource) }) : h('span', { class: 'font-medium', text: coverageLabel(it.resource) }),
        h('div', { class: 'text-xs text-secondary-500', text: t('cov.module.' + it.module) }));
      var cells = [name, h('span', { class: 'tabular-nums', text: fmtNum(it.total || 0) })];
      langs.forEach(function (code) {
        var n = (it.missing || {})[code] || 0;
        cells.push(n
          ? h('a', { href: href || '#', class: 'eco-tag eco-tag--warning inline-flex leading-tight px-2 py-0.5 text-xs rounded-md whitespace-nowrap tabular-nums', 'data-cov': it.resource + ':' + code, text: t('cov.missingN', { n: n }), title: t('cov.open') })
          : h('span', { class: 'eco-text-success inline-flex items-center gap-1 text-xs', 'data-cov': it.resource + ':' + code }, icon('check', 'h-3.5 w-3.5'), h('span', { text: t('cov.complete') })));
      });
      return { item: it, cells: cells };
    });
    clear(body).appendChild(h('div', {}, chips, table(cols(langs), rows)));
  }
  function load() {
    var my = ++seq;
    skeleton();
    baseApi('GET', '/languages/coverage').then(function (c) { if (my === seq) render(c); }, function (err) { if (my === seq) clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
  box.reload = load;
  return box;
}

function languagesPage(container, query, ctx) {
  var list = listPage(container, 'languages', query, ctx);
  var cov = coveragePanel();
  container.appendChild(cov);
  var hook = function () {
    if (!cov.isConnected) { languageHooks.splice(languageHooks.indexOf(hook), 1); return; }
    cov.reload();
    list.reload();
  };
  languageHooks.push(hook);
  return list;
}

function statCard(label, value, sub, href, opts) {
  opts = opts || {};
  var inner = h('div', { class: 'p-4 flex flex-col gap-1 h-full' },
    h('div', { class: 'font-medium text-secondary-700 dark:text-secondary-200', text: label }),
    h('div', { class: 'text-3xl font-semibold tabular-nums' + (opts.alert ? ' text-error-600 dark:text-error-400' : ''), text: fmtNum(value || 0) }),
    sub ? h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: sub }) : null);
  if (!href) return card(inner);
  return h('a', { href: href, class: 'block bg-white dark:bg-secondary-800 overflow-hidden rounded-md shadow hover:bg-secondary-50 dark:hover:bg-secondary-700 transition' }, inner);
}

var OVERVIEW_GROUPS = [
  { id: 'progress', items: [
    { page: 'currencies' },
    { page: 'metas', more: 'meta-rules', moreKey: 'ov.sub.rules' },
    { page: 'meta-relations', active: 'activeMetaRelations' },
    { page: 'level-tables', active: 'activeLevelTables' },
    { page: 'tasks', active: 'activeTasks', more: 'task-groups', moreKey: 'ov.sub.groups' },
    { page: 'achievements', active: 'activeAchievements' },
    { page: 'streaks', active: 'activeStreaks' }
  ] },
  { id: 'sales', items: [
    { page: 'products', active: 'activeProducts' },
    { page: 'offers', active: 'activeOffers' },
    { page: 'items', active: 'activeItems' },
    { page: 'gifts', active: 'activeGifts' },
    { page: 'friend-gifts', more: 'friendGiftsPending', moreKey: 'ov.sub.waiting' }
  ] },
  { id: 'play', items: [
    { page: 'prizes' },
    { page: 'prize-groups' },
    { page: 'games', active: 'activeGames' },
    { page: 'chance-games', active: 'activeChanceGames', more: 'chancePlaysToday', moreKey: 'ov.sub.playsToday' }
  ] },
  { id: 'growth', items: [
    { page: 'event-types', active: 'activeEventTypes' },
    { page: 'event-rules', active: 'activeEventRules' },
    { page: 'referrals', more: 'referralsCompleted', moreKey: 'ov.sub.completed' }
  ] }
];
var OVERVIEW_SHARED = ['categories', 'tags'];
var OVERVIEW_TINT = { progress: 'violet', sales: 'info', play: 'warn', growth: 'teal' };
var PAY_STATUS_ORDER = ['success', 'pending', 'failed', 'expired'];

function ovAllowed(page) {
  var ps = psSection(page);
  if (ps) return !ps.acl || aclCan(ps.acl, 0);
  return pageAllowed(page);
}

function ovIcon(page, tint, duo) {
  var box = h('span', { class: 'eco-ov-ic eco-ov-tint-' + (tint || 'info'), 'aria-hidden': 'true' });
  var P = window.NakamaPlane;
  if (P && P.iconD && (duo || P.iconFor)) box.innerHTML = P.iconD(duo || P.iconFor('#/economy/' + page));
  else box.appendChild(icon(page, 'h-4 w-4'));
  return box;
}

function ovDelta(cur, prev, sinceKey) {
  cur = Number(cur || 0);
  prev = Number(prev || 0);
  if (!cur && !prev) return h('span', { class: 'eco-ov-delta', 'data-ov-delta': 'flat', text: t(sinceKey + '.none') });
  if (!prev) return h('span', { class: 'eco-ov-delta eco-ov-delta--up', 'data-ov-delta': 'new', text: t('ov.delta.new') });
  var pct = Math.round((cur - prev) / prev * 100);
  var dir = pct > 0 ? 'up' : pct < 0 ? 'down' : 'flat';
  var sign = pct > 0 ? '+' : pct < 0 ? '−' : '';
  return h('span', { class: 'eco-ov-delta eco-ov-delta--' + dir, 'data-ov-delta': dir },
    h('span', { class: 'eco-ov-delta-num', dir: 'ltr', text: sign + fmtNum(Math.abs(pct)) + ovPct() }),
    h('span', { text: ' ' + t(sinceKey) }));
}

function ovPct() {
  return lang() === 'fa' ? '٪' : '%';
}

function ovRtl() {
  return document.documentElement.getAttribute('dir') === 'rtl';
}

function ovSpark(values) {
  if (ovRtl()) values = values.slice().reverse();
  var w = 120;
  var hgt = 32;
  var max = Math.max.apply(null, values.concat([0]));
  var svg = h('svg:svg', { viewBox: '0 0 ' + w + ' ' + hgt, preserveAspectRatio: 'none', class: 'eco-ov-spark', 'aria-hidden': 'true', focusable: 'false' });
  if (!max || values.length < 2) {
    svg.appendChild(h('svg:line', { x1: 0, x2: w, y1: hgt - 1, y2: hgt - 1, class: 'eco-ov-spark-base' }));
    return svg;
  }
  var step = w / (values.length - 1);
  var pts = values.map(function (v, i) { return (i * step).toFixed(1) + ',' + (hgt - 2 - (v / max) * (hgt - 4)).toFixed(1); });
  svg.appendChild(h('svg:path', { d: 'M0,' + hgt + ' L' + pts.join(' L') + ' L' + w + ',' + hgt + ' Z', class: 'eco-ov-spark-area' }));
  svg.appendChild(h('svg:polyline', { points: pts.join(' '), class: 'eco-ov-spark-line', 'vector-effect': 'non-scaling-stroke' }));
  return svg;
}

function ovKpi(opts) {
  var head = h('div', { class: 'eco-ov-kpi-head' }, ovIcon(opts.icon, opts.tint, opts.duo),
    h('span', { class: 'eco-ov-kpi-label', text: opts.label }));
  var value = h('div', { class: 'eco-ov-kpi-value' + (opts.alert ? ' eco-ov-alert' : '') },
    h('span', { class: 'tabular-nums', text: fmtNum(opts.value || 0) }),
    opts.unit ? h('span', { class: 'eco-ov-kpi-unit', text: opts.unit }) : null);
  var inner = [head, value, opts.spark || null, h('div', { class: 'eco-ov-kpi-foot' }, opts.foot || null)];
  var attrs = { class: 'eco-ov-kpi', 'data-ov-kpi': opts.id };
  if (opts.attrs) Object.keys(opts.attrs).forEach(function (k) { attrs[k] = opts.attrs[k]; });
  if (opts.href) { attrs.href = opts.href; return h('a', attrs, inner); }
  return h('div', attrs, inner);
}

function ovTile(item, c, tint) {
  var page = item.page;
  var n = Number(c[page] || 0);
  var bits = [];
  if (item.active && c[item.active] !== undefined && n) bits.push(t('ov.sub.active', { n: localDigits(fmtNum(c[item.active])) }));
  if (item.more && c[item.more]) bits.push(t(item.moreKey, { n: localDigits(fmtNum(c[item.more])) }));
  var label = t('nav.' + page);
  var sub = n ? bits.join(t('ov.sep')) : t('ov.sub.empty');
  return h('a', { href: hashFor(page), class: 'eco-ov-tile' + (n ? '' : ' eco-ov-tile--zero'), 'data-ov-tile': page, title: label },
    ovIcon(page, n ? tint : 'mute'),
    h('span', { class: 'eco-ov-tile-text' },
      h('span', { class: 'eco-ov-tile-label', text: label }),
      h('span', { class: 'eco-ov-tile-sub', text: sub || ' ' })),
    h('span', { class: 'eco-ov-tile-num tabular-nums', text: fmtNum(n) }));
}

function ovSection(titleText, aside, body, attrs) {
  var a = { class: 'eco-ov-section' };
  if (attrs) Object.keys(attrs).forEach(function (k) { a[k] = attrs[k]; });
  return h('section', a,
    h('div', { class: 'eco-ov-section-head' }, h('h2', { class: 'eco-ov-h2', text: titleText }), aside || null),
    body);
}

function ovShortDay(day) {
  var d = new Date(day + 'T12:00:00');
  if (isNaN(d.getTime())) return day;
  return new Intl.DateTimeFormat(locale(), dateOpts({ month: 'short', day: 'numeric' })).format(d);
}

function ovChart(daily, metric) {
  var vals = daily.map(function (d) { return Number((metric === 'revenue' ? d.toman : d.payments) || 0); });
  var total = vals.reduce(function (a, b) { return a + b; }, 0);
  var max = Math.max.apply(null, vals.concat([0]));
  var box = h('div', { class: 'eco-ov-chart', 'data-ov-chart': metric });
  var unit = metric === 'revenue' ? t('ov.chart.toman') : t('ov.chart.count');
  var fmt = function (v) { return fmtNum(v) + (metric === 'revenue' ? ' ' + t('ov.chart.toman') : ''); };
  var peak = vals.indexOf(max);
  box.appendChild(h('div', { class: 'eco-ov-chart-meta' },
    h('span', {}, h('span', { class: 'eco-ov-chart-total tabular-nums', text: fmtNum(total) }), h('span', { class: 'eco-ov-muted', text: ' ' + unit + ' · ' + t('ov.chart.days') })),
    max ? h('span', { class: 'eco-ov-muted', text: t('ov.chart.peak', { v: fmt(max), day: ovShortDay(daily[peak].day) }) }) : null));
  var W = 600;
  var H = 140;
  var gap = 3;
  var bw = daily.length ? W / daily.length : W;
  var svg = h('svg:svg', { viewBox: '0 0 ' + W + ' ' + H, preserveAspectRatio: 'none', class: 'eco-ov-bars', role: 'img',
    'aria-label': max ? t('ov.chart.aria', { what: metric === 'revenue' ? t('ov.revenue') : t('ov.chart.payments'), total: fmt(total), peak: fmt(max), day: ovShortDay(daily[peak].day) }) : t('ov.chart.empty.' + metric) });
  [0.25, 0.5, 0.75].forEach(function (f) { svg.appendChild(h('svg:line', { x1: 0, x2: W, y1: (H * f).toFixed(1), y2: (H * f).toFixed(1), class: 'eco-ov-grid', 'vector-effect': 'non-scaling-stroke' })); });
  daily.forEach(function (d, i) {
    var v = vals[i];
    var bh = max ? Math.max(v ? 2 : 0, v / max * (H - 6)) : 0;
    var x = ovRtl() ? W - (i + 1) * bw : i * bw;
    var g = h('svg:g', { class: 'eco-ov-bar' + (i === daily.length - 1 ? ' eco-ov-bar--today' : '') },
      h('svg:rect', { x: (x + gap / 2).toFixed(1), y: 0, width: Math.max(1, bw - gap).toFixed(1), height: H, class: 'eco-ov-bar-hit' }),
      h('svg:rect', { x: (x + gap / 2).toFixed(1), y: (H - bh).toFixed(1), width: Math.max(1, bw - gap).toFixed(1), height: bh.toFixed(1), rx: 2, class: 'eco-ov-bar-fill' }));
    g.appendChild(h('svg:title', { text: ovShortDay(d.day) + ': ' + fmt(v) }));
    svg.appendChild(g);
  });
  var plot = h('div', { class: 'eco-ov-plot' }, svg);
  if (!max) plot.appendChild(h('div', { class: 'eco-ov-plot-empty', text: t('ov.chart.empty.' + metric) }));
  box.appendChild(plot);
  if (daily.length) {
    box.appendChild(h('div', { class: 'eco-ov-axis eco-ov-muted' },
      h('span', { text: ovShortDay(daily[0].day) }),
      h('span', { text: t('ov.today') })));
  }
  return box;
}

function ovStatus(by, grantErrors) {
  var total = PAY_STATUS_ORDER.reduce(function (a, st) { return a + Number(by[st] || 0); }, 0);
  var box = h('div', { class: 'eco-ov-status', 'data-ov-status': '' });
  var barEl = h('div', { class: 'eco-ov-stack', role: 'img', 'aria-label': PAY_STATUS_ORDER.map(function (st) { return t('status.' + st) + ' ' + fmtNum(by[st] || 0); }).join(t('ov.sep')) });
  var legend = h('ul', { class: 'eco-ov-legend' });
  PAY_STATUS_ORDER.forEach(function (st) {
    var n = Number(by[st] || 0);
    var pct = total ? n / total * 100 : 0;
    if (n) barEl.appendChild(h('span', { class: 'eco-ov-stack-seg eco-ov-st-' + st, style: 'width:' + pct.toFixed(2) + '%' }));
    legend.appendChild(h('li', {}, h('a', { href: hashFor('payments', null, { status: st }), class: 'eco-ov-legend-row', 'data-ov-status-row': st },
      h('span', { class: 'eco-ov-dot eco-ov-st-' + st, 'aria-hidden': 'true' }),
      h('span', { class: 'eco-ov-legend-label', text: t('status.' + st) }),
      h('span', { class: 'eco-ov-muted tabular-nums', text: total ? fmtNum(pct, { maximumFractionDigits: 0 }) + ovPct() : '' }),
      h('span', { class: 'eco-ov-legend-num tabular-nums', text: fmtNum(n) }))));
  });
  if (!total) barEl.classList.add('eco-ov-stack--empty');
  box.appendChild(h('div', { class: 'eco-ov-status-head' }, h('span', { class: 'eco-ov-h3', text: t('ov.status') }), h('span', { class: 'eco-ov-muted tabular-nums', text: t('ov.status.total', { n: localDigits(fmtNum(total)) }) })));
  box.appendChild(barEl);
  box.appendChild(legend);
  if (grantErrors) box.appendChild(h('a', { href: hashFor('payments', null, { reward: 'pending' }), class: 'eco-ov-alertrow', 'data-ov-grant-errors': '' },
    icon('alert', 'h-4 w-4'), h('span', { text: t('ov.grantErrors.n', { n: localDigits(fmtNum(grantErrors)) }) })));
  return box;
}

function ovEmpty(textKey, hintKey, ctaKey, href, attr) {
  var a = { class: 'eco-ov-empty' };
  a[attr] = '';
  return h('div', a,
    h('div', { class: 'eco-ov-empty-text' },
      h('div', { class: 'eco-ov-empty-title', text: t(textKey) }),
      hintKey ? h('div', { class: 'eco-ov-muted', text: t(hintKey) }) : null),
    href ? h('a', { href: href, class: 'eco-ov-cta', 'data-ov-cta': attr }, h('span', { text: t(ctaKey) }), icon('next', 'h-3.5 w-3.5 eco-ov-cta-ic')) : null);
}

function overviewPage(container) {
  var frame = pageFrame(t('nav.overview'), t('desc.overview'));
  clear(container).appendChild(frame.el);
  var body = h('div', { class: 'eco-ov', 'data-ov': '' });
  frame.content.appendChild(body);
  var metric = 'revenue';

  function skeleton() {
    var kpis = h('div', { class: 'eco-ov-kpis' });
    for (var i = 0; i < 4; i++) kpis.appendChild(h('div', { class: 'eco-ov-kpi' }, bar('6rem'), bar('4rem', 'h-7'), bar('8rem')));
    clear(body).appendChild(kpis);
    var grp = h('div', { class: 'eco-ov-tiles' });
    for (var j = 0; j < 8; j++) grp.appendChild(h('div', { class: 'eco-ov-tile' }, bar('2rem', 'h-8'), h('span', { class: 'eco-ov-tile-text' }, bar('5rem'), bar('3rem'))));
    body.appendChild(ovSection(t('ov.catalog'), null, grp));
  }

  function render(s) {
    s = s || {};
    var c = s.counts || {};
    var by = s.paymentsByStatus || {};
    var daily = s.revenueDaily || [];
    clear(body);

    var kpis = h('div', { class: 'eco-ov-kpis', 'data-ov-kpis': '' });
    var payOk = ovAllowed('payments');
    if (payOk) {
      kpis.appendChild(ovKpi({ id: 'revenue', label: t('ov.revenue30'), value: c.revenueToman30, unit: t('ov.chart.toman'), href: hashFor('payments', null, { status: 'success' }), icon: 'payments', duo: 'wallet', tint: 'ok',
        foot: ovDelta(c.revenueToman30, c.revenueTomanPrev30, 'ov.delta.vs30'), spark: ovSpark(daily.map(function (d) { return Number(d.toman || 0); })) }));
      kpis.appendChild(ovKpi({ id: 'payments', label: t('ov.paymentsToday'), value: c.paymentsToday, href: hashFor('payments'), icon: 'payments', duo: 'card', tint: 'info',
        foot: ovDelta(c.paymentsToday, c.paymentsPrevDay, 'ov.delta.vsDay'), spark: ovSpark(daily.map(function (d) { return Number(d.payments || 0); })) }));
    }
    if (c.eventsToday !== undefined && ovAllowed('event-log')) {
      kpis.appendChild(ovKpi({ id: 'events', label: t('ov.eventsToday'), value: c.eventsToday, href: hashFor('event-log'), icon: 'event-types', duo: 'zap', tint: 'violet', attrs: { 'data-ov-events': '' },
        foot: ovDelta(c.eventsToday, c.eventsPrevDay, 'ov.delta.vsDay') }));
      var failed = Number(c.claimsFailed || 0);
      kpis.appendChild(ovKpi({ id: 'claims', label: t('ov.claimsReady'), value: c.claimsReady, href: hashFor('event-log'), icon: 'event-rules', duo: 'gift', tint: failed ? 'bad' : 'warn',
        foot: failed ? h('span', { class: 'eco-ov-delta eco-ov-delta--down', 'data-ov-claims-failed': '', text: t('ov.claimsFailed.n', { n: localDigits(fmtNum(failed)) }) })
          : h('span', { class: 'eco-ov-delta', text: c.claimsReady ? t('ov.claims.waiting') : t('ov.claims.clear') }) }));
    }
    if (kpis.children.length) body.appendChild(kpis);

    var catalog = h('div', { class: 'eco-ov-groups', 'data-ov-catalog': '' });
    OVERVIEW_GROUPS.forEach(function (g) {
      var items = g.items.filter(function (it) { return ovAllowed(it.page); });
      if (!items.length) return;
      var total = items.reduce(function (a, it) { return a + Number(c[it.page] || 0); }, 0);
      var tiles = h('div', { class: 'eco-ov-tiles', style: '--ov-cols:' + Math.ceil(items.length / Math.ceil(items.length / 4)) });
      items.forEach(function (it) { tiles.appendChild(ovTile(it, c, OVERVIEW_TINT[g.id])); });
      catalog.appendChild(h('div', { class: 'eco-ov-group', 'data-ov-group': g.id },
        h('div', { class: 'eco-ov-group-head' },
          h('h3', { class: 'eco-ov-h3', text: t('navg.' + g.id) }),
          h('span', { class: 'eco-ov-muted tabular-nums', text: t('ov.group.total', { n: localDigits(fmtNum(total)) }) })),
        tiles));
    });
    var shared = OVERVIEW_SHARED.filter(ovAllowed);
    var sharedEl = shared.length ? h('div', { class: 'eco-ov-shared' }, h('span', { class: 'eco-ov-muted', text: t('ov.shared') }),
      shared.map(function (p) { return h('a', { href: hashFor(p), class: 'eco-ov-chip', 'data-ov-tile': p }, h('span', { text: t('nav.' + p) }), h('span', { class: 'tabular-nums eco-ov-chip-num', text: fmtNum(c[p] || 0) })); })) : null;
    if (catalog.children.length) body.appendChild(ovSection(t('ov.catalog'), sharedEl, catalog));

    if (payOk) {
      var methods = s.paymentMethods || [];
      var activeMethods = methods.filter(function (m) { return m.isActive; });
      var chartBox = h('div', { class: 'eco-ov-chartbox' });
      var seg = segmented([{ value: 'revenue', label: t('ov.revenue') }, { value: 'payments', label: t('ov.chart.payments') }], metric, function (v) { metric = v; paintChart(); }, t('ov.chart.metric'));
      seg.setAttribute('data-ov-metric', '');
      function paintChart() { clear(chartBox).appendChild(ovChart(daily, metric)); }
      paintChart();
      var left = h('div', { class: 'eco-ov-pane' }, h('div', { class: 'eco-ov-pane-head' }, h('span', { class: 'eco-ov-h3', text: t('ov.chart.title') }), seg), chartBox);
      var right = h('div', { class: 'eco-ov-pane' });
      if (!activeMethods.length) {
        right.appendChild(ovEmpty('ov.empty.methods', 'ov.empty.methods.hint', 'ov.empty.methods.cta', ovAllowed('payment-methods') ? hashFor('payment-methods') : null, 'data-ov-empty-methods'));
      }
      if (c.payments) right.appendChild(ovStatus(by, c.grantErrors));
      else if (activeMethods.length) right.appendChild(ovEmpty('ov.empty.payments', 'ov.empty.payments.hint', 'ov.empty.payments.cta', ovAllowed('products') ? hashFor('products') : null, 'data-ov-empty-payments'));
      if (methods.length && activeMethods.length) {
        right.appendChild(h('div', { class: 'eco-ov-methods', 'data-ov-methods': '' },
          h('span', { class: 'eco-ov-muted', text: t('ov.methods') }),
          methods.filter(function (m) { return m.isActive || m.configured; }).map(function (m) {
            var label = has('method.' + m.method) ? t('method.' + m.method) : m.method;
            var state = !m.isActive ? 'off' : !m.configured ? 'bad' : m.sandbox ? 'sandbox' : 'on';
            return h('a', { href: ovAllowed('payment-methods') ? hashFor('payment-methods') : '#', class: 'eco-ov-chip eco-ov-method--' + state, 'data-ov-method': m.method, title: t('ov.method.' + state) },
              h('span', { class: 'eco-ov-dot', 'aria-hidden': 'true' }), h('span', { text: label }), h('span', { class: 'eco-ov-muted', text: t('ov.method.' + state) }));
          })));
      }
      body.appendChild(ovSection(t('ov.payments'), h('a', { href: hashFor('payments'), class: 'eco-ov-link', text: t('ov.allPayments') }),
        h('div', { class: 'eco-ov-payments', 'data-ov-payments': '' }, left, right)));
    }
    if (!body.children.length) body.appendChild(card(emptyBlock(t('ov.empty.none'))));
  }

  function load() {
    skeleton();
    api('GET', '/summary').then(render, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}

var PAYMENT_SETTINGS_FIELDS = [
  { name: 'rewardRetry', type: 'bool', inline: true, label: 'f.ps.retry', hint: 'f.ps.retry.hint' },
  { type: 'row', fields: [
    { name: 'retryMinutes', type: 'number', label: 'f.ps.retryMinutes', hint: 'f.ps.retryMinutes.hint', required: true, min: 1, max: 1440 },
    { name: 'maxAttempts', type: 'number', label: 'f.ps.maxAttempts', hint: 'f.ps.maxAttempts.hint', required: true, min: 1, max: 1000 }
  ] },
  { name: 'grantOnNextCall', type: 'bool', inline: true, label: 'f.ps.nextCall', hint: 'f.ps.nextCall.hint' },
  { name: 'paidConflict', type: 'select', label: 'f.ps.conflict', hint: 'f.ps.conflict.hint', required: true, noPlaceholder: true, options: opts('ps.conflict.', ['refund', 'keep']) },
  { type: 'row', fields: [
    { name: 'oncePending', type: 'select', label: 'f.ps.once', hint: 'f.ps.once.hint', required: true, noPlaceholder: true, options: opts('ps.once.', ['hold', 'replace']) },
    { name: 'onceHoldMinutes', type: 'number', label: 'f.ps.onceHold', hint: 'f.ps.onceHold.hint', required: true, min: 1, max: 1440 }
  ] },
  { name: 'pendingMinutes', type: 'number', label: 'f.ps.pending', hint: 'f.ps.pending.hint', required: true, min: 0, max: 10080 },
  { name: 'expiredReceipts', type: 'select', label: 'f.ps.expired', hint: 'f.ps.expired.hint', required: true, noPlaceholder: true, options: opts('ps.expired.', ['verified', 'refuse']) },
  { name: 'closeSandboxOnOff', type: 'bool', inline: true, label: 'f.ps.sandbox', hint: 'f.ps.sandbox.hint' },
  { type: 'heading', label: 'f.ps.stores' },
  { name: 'liveAcceptsTestPurchases', type: 'bool', inline: true, label: 'f.ps.liveTest', hint: 'f.ps.liveTest.hint' },
  { type: 'row', fields: [
    { name: 'appStorePayload', type: 'select', label: 'f.ps.payload', hint: 'f.ps.payload.hint', required: true, noPlaceholder: true, options: opts('ps.payload.', ['optional', 'required']) },
    { name: 'storeConsumed', type: 'select', label: 'f.ps.consumed', hint: 'f.ps.consumed.hint', required: true, noPlaceholder: true, options: opts('ps.consumed.', ['accept', 'refuse']) }
  ] },
  { name: 'googleAcknowledge', type: 'bool', inline: true, label: 'f.ps.ack', hint: 'f.ps.ack.hint' },
  { name: 'stripeToleranceSeconds', type: 'number', label: 'f.ps.stripeTolerance', hint: 'f.ps.stripeTolerance.hint', required: true, min: 30, max: 3600 }
];

function paymentSettingsCall(method, path, body) {
  return api(method, path, body).then(function (res) { return Object.assign({}, (res && res.settings) || {}); });
}

function paymentSettingsPage(container) {
  var frame = pageFrame(t('nav.payment-settings'), t('desc.payment-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(paymentSettingsCall, '/payment-settings', null, PAYMENT_SETTINGS_FIELDS, { id: 'payments', fixPayload: function (p) {
    delete p.updatedAt;
    ['retryMinutes', 'maxAttempts', 'onceHoldMinutes', 'stripeToleranceSeconds', 'pendingMinutes'].forEach(function (k) { p[k] = Number(p[k]); });
    return p;
  } });
  frame.content.appendChild(page.el);
  page.load();
}

CUSTOM_PAGES['payment-settings'] = function (root) { paymentSettingsPage(root); };
