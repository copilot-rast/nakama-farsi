var SECTIONS = {
  base: ['categories', 'tags'],
  economy: ['overview', 'currencies', 'metas', 'prizes', 'products', 'gifts', 'payments'],
  publication: ['faqs', 'news', 'comments', 'notifications']
};
var SETTINGS_TABS = ['payment-methods', 'sms', 'email', 'languages'];
var LEGACY_HASHES = { economy: { 'payment-methods': 'payment-methods' }, publication: { providers: 'sms' } };
var SECTION_ORDER = ['base', 'economy', 'publication'];
var MOVED_PAGES = { 'economy/categories': 'base/categories', 'economy/tags': 'base/tags' };
var PAGES = SECTIONS.base.concat(SECTIONS.economy, SECTIONS.publication);
var LIMIT = 20;
var lastQuery = { comments: { status: 'pending' } };

function sectionOf(page) {
  if (SETTINGS_TABS.indexOf(page) !== -1) return 'settings';
  if (SECTIONS.base.indexOf(page) !== -1) return 'base';
  return SECTIONS.publication.indexOf(page) !== -1 ? 'publication' : 'economy';
}

function apiFor(page) {
  var def = RESOURCES[page];
  if (def && def.api === 'base') return baseApi;
  var sec = sectionOf(page);
  return sec === 'publication' ? pubApi : sec === 'base' ? baseApi : api;
}

function hashFor(page, id, query) {
  return '#/' + sectionOf(page) + '/' + page + (id ? '/' + encodeURIComponent(id) : '') + qs(query || {});
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
  var columns = def.columns.map(function (c) { return { label: t(c.label), width: c.width, align: c.align, skeleton: c.skeleton }; });
  var state = { q: query.q || '', page: Math.max(1, parseInt(query.page, 10) || 1), limit: LIMIT, filters: {}, total: 0, loading: true, loaded: false };
  (def.filters || []).forEach(function (f) { if (query[f.name]) state.filters[f.name] = query[f.name]; });
  lastQuery[name] = currentQuery();

  var create = def.fields ? button({ variant: 'primary', size: 'page', icon: 'plus', label: t('new.' + name), onclick: function () { go(name, 'new', currentQuery()); } }) : def.create ? def.create() : null;
  var frame = def.settings ? tabFrame(t('desc.' + name), create) : pageFrame(t('nav.' + name), t('desc.' + name), create);
  var pagerTop = h('div', { class: 'relative flex items-center justify-end' });
  var pagerBottom = h('div', { class: 'flex gap-2 py-4 items-center justify-end' });
  var tableBox = h('div', {});
  var search = searchInput(state.q, t(has('search.' + name) ? 'search.' + name : 'search'), debounce(function (e) {
    state.q = toLatinDigits(e.target.value).trim();
    state.page = 1;
    sync();
    load();
  }, 300));
  var filters = (def.filters || []).map(function (f) {
    var opts2 = typeof f.options === 'function' ? f.options() : f.options;
    if (f.type === 'ref') {
      var picker = refPicker({ resource: f.resource, api: typeof f.api === 'function' ? f.api : f.api === 'publication' ? pubApi : api, filter: f.filter, value: state.filters[f.name] || '', label: t(f.label), placeholder: t(f.label) + ': ' + t('all'), onchange: function (v) {
        state.filters[f.name] = v || '';
        state.page = 1;
        sync();
        load();
      } });
      return h('div', { class: 'x-label x-input eco-filter-ref', style: INPUT_VARS }, picker);
    }
    if (f.type === 'text') {
      var inp = h('input', { class: INPUT_SM + ' font-mono eco-filter-text', dir: 'ltr', type: 'text', value: state.filters[f.name] || '', placeholder: t(f.label), 'aria-label': t(f.label) });
      inp.addEventListener('input', debounce(function () { state.filters[f.name] = toLatinDigits(inp.value).trim(); state.page = 1; sync(); load(); }, 300));
      return h('label', { class: 'x-label block relative x-input', style: INPUT_VARS }, inp);
    }
    return filterSelect(t(f.label), state.filters[f.name], opts2, function (v) { state.filters[f.name] = v; state.page = 1; sync(); load(); });
  });
  frame.content.appendChild(h('div', {},
    h('div', { class: 'flex flex-col items-start pb-2 gap-2' },
      h('div', { class: 'flex flex-col lg:flex-row lg:items-center gap-2 w-full' },
        search,
        filters.length ? h('div', { class: 'eco-filters flex flex-wrap items-center gap-2' }, filters) : null,
        h('div', { class: 'flex-grow hidden lg:block' }),
        pagerTop)),
    tableBox,
    pagerBottom));
  clear(container).appendChild(frame.el);

  function currentQuery() {
    var q = Object.assign({}, state.filters);
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
    clear(pagerTop).appendChild(pager(state, function (p) { state.page = p; sync(); load(); }, function () { load(); }));
    clear(pagerBottom).appendChild(pager(state, function (p) { state.page = p; sync(); load(); window.scrollTo(0, 0); }));
  }
  function open(item) { go(name, item.id, currentQuery()); }
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
    return call('GET', '/' + name + qs(params)).then(function (res) {
      if (my !== seq) return;
      state.loading = false;
      state.loaded = true;
      state.total = (res && res.total) || 0;
      var items = (res && res.items) || [];
      tableBox.classList.remove('opacity-60');
      if (!items.length && state.page > 1 && state.total > 0) {
        state.page = Math.max(1, Math.ceil(state.total / state.limit));
        sync();
        return load();
      }
      clear(tableBox);
      if (!items.length) {
        var filtered = state.q || Object.keys(state.filters).some(function (k) { return state.filters[k]; });
        tableBox.appendChild(card(emptyBlock(filtered ? t('empty.search') : t('empty.title'), filtered ? null : t('empty.hint.' + name),
          !filtered && def.fields ? button({ variant: 'primary', size: 'page', icon: 'plus', label: t('new.' + name), onclick: function () { go(name, 'new'); } }) : !filtered && def.create ? def.create() : null)));
      } else {
        tableBox.appendChild(table(columns, items.map(function (it) {
          return { item: it, cells: def.columns.map(function (c) { return c.render(it); }) };
        }), { onrow: def.noRow ? null : open }));
      }
      paintPager();
    }, function (err) {
      if (my !== seq) return;
      state.loading = false;
      tableBox.classList.remove('opacity-60');
      clear(tableBox).appendChild(errorBlock(err, load));
      paintPager();
    });
  }
  load();
  return { reload: load, name: name };
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
  var esc = function (e) { if (e.key === 'Escape' && panelState && !document.querySelector('[data-economy="panel"] .eco-pop:not([hidden])')) { e.preventDefault(); back(); } };
  document.addEventListener('keydown', esc, true);
  panelState = { el: el, key: name + '/' + id, esc: esc };
  document.body.appendChild(el);
  body.appendChild(formSkeleton(def));

  var call = apiFor(name);
  var item = isNew ? Promise.resolve(def.defaults(query)) : call('GET', '/' + name + '/' + encodeURIComponent(id));
  Promise.all([item, freshLanguages(), loadKeyTypes()]).then(function (all) { return all[0]; }).then(function (data) {
    if (!panelState || panelState.el !== el) return;
    var values = clone(data) || {};
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
        call('DELETE', '/' + name + '/' + encodeURIComponent(id)).then(function () {
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
      call(isNew ? 'POST' : 'PUT', '/' + name + (isNew ? '' : '/' + encodeURIComponent(id)), payload).then(function () {
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
    it.status === 'success' && !it.rewardGranted ? tag(t('pay.reward.notGranted'), 'error') : null);
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
    if (p.status === 'success' && (!p.rewardGranted || p.rewardError)) actions.appendChild(button({ variant: 'primary', size: 'page', icon: 'prizes', label: t('pay.retry'), onclick: function (e) { act(e.currentTarget, 'retry-grant', 'pay.retry.done'); } }));
    clear(frame.actions).appendChild(actions);
    frame.actions.appendChild(backBtn);

    grid.appendChild(section('col.status', dl([
      ['col.status', paymentStatus(p)],
      ['col.method', h('div', { class: 'flex flex-wrap gap-1' }, tag(t('method.' + p.method), 'secondary'), p.sandbox ? tag(t('pay.sandbox'), 'warning') : null)],
      ['pay.reward', tag(p.rewardGranted ? t('pay.reward.granted') : t('pay.reward.notGranted'), p.rewardGranted ? 'success' : 'secondary')],
      ['pay.reward.error', p.rewardError ? h('span', { class: 'text-error-600 dark:text-error-400', text: p.rewardError }) : null],
      ['pay.error', p.error ? h('span', { class: 'text-error-600 dark:text-error-400', text: p.error }) : null],
      ['pay.outcome', p.outcome ? mono(p.outcome) : null]
    ])));
    grid.appendChild(section('pay.amounts', dl([
      ['pay.original', p.originalToman != null ? h('span', { class: 'tabular-nums', text: t('toman', { n: p.originalToman }) }) : null],
      ['pay.discount', p.discountPercent ? h('span', { text: t('percent', { n: p.discountPercent }) }) : null],
      ['pay.final', p.finalToman != null ? h('span', { class: 'tabular-nums font-semibold', text: t('toman', { n: p.finalToman }) }) : null],
      ['pay.rial', p.amountRial ? h('span', { class: 'tabular-nums', text: fmtNum(p.amountRial) }) : null],
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
        body.appendChild(h('div', { class: 'mb-4' }, ctl));
        fieldEls[f.name] = { setError: function () {} };
        return;
      }
      ctl = textInput(cfg[f.name] == null ? '' : cfg[f.name], { ltr: true, mono: f.type !== 'textarea' || true, multiline: f.type === 'textarea', rows: 3, label: label, oninput: function (v) { cfg[f.name] = v; } });
      if (f.type === 'secret') { ctl.setAttribute('autocomplete', 'off'); ctl.classList.add('eco-secret'); }
      ctl.classList.add('text-sm');
      var wrap = field(label, ctl, f.type === 'secret' && cfg[f.name] ? t('pm.secret.hint') : null, { optional: !f.required });
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
      h('div', { class: 'flex justify-end pt-2' }, saveBtn)));
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

var OVERVIEW_CATALOG = ['currencies', 'metas', 'prizes', 'products', 'gifts', 'categories', 'tags'];

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

function overviewPage(container) {
  var frame = pageFrame(t('nav.overview'), t('desc.overview'));
  clear(container).appendChild(frame.el);
  var body = h('div', { class: 'flex flex-col gap-6' });
  frame.content.appendChild(body);

  function skeleton() {
    clear(body);
    [OVERVIEW_CATALOG.length, 6].forEach(function (n, k) {
      var grid = h('div', { class: 'eco-stat-grid' });
      for (var i = 0; i < n; i++) grid.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem'), bar('3rem', 'h-7'), k ? bar('8rem') : null)));
      body.appendChild(h('div', {}, h('div', { class: 'mb-3' }, bar('7rem', 'h-4')), grid));
    });
  }

  function render(s) {
    s = s || {};
    var c = s.counts || {};
    var by = s.paymentsByStatus || {};
    clear(body);
    var catalog = h('div', { class: 'eco-stat-grid' });
    OVERVIEW_CATALOG.forEach(function (name) {
      var active = name === 'products' ? c.activeProducts : name === 'gifts' ? c.activeGifts : undefined;
      catalog.appendChild(statCard(t('nav.' + name), c[name], active !== undefined ? t('active') + ': ' + fmtNum(active) : null, hashFor(name)));
    });
    body.appendChild(h('div', {}, h('h2', { class: 'text-lg font-medium mb-3', text: t('ov.catalog') }), catalog));

    var payments = h('div', { class: 'eco-stat-grid' });
    payments.appendChild(statCard(t('ov.revenue30'), c.revenueToman30, t('toman', { n: '' }).trim(), null));
    payments.appendChild(statCard(t('ov.paymentsToday'), c.paymentsToday, t('ov.total') + ': ' + fmtNum(c.payments || 0), hashFor('payments')));
    ['pending', 'success', 'failed', 'expired'].forEach(function (st) {
      payments.appendChild(statCard(t('status.' + st), by[st], null, hashFor('payments', null, { status: st })));
    });
    if (c.grantErrors) payments.appendChild(statCard(t('ov.grantErrors'), c.grantErrors, null, hashFor('payments', null, { status: 'success' }), { alert: true }));
    body.appendChild(h('div', {}, h('h2', { class: 'text-lg font-medium mb-3', text: t('ov.payments') }), payments));

    var methods = s.paymentMethods || [];
    if (methods.length) {
      var list = h('div', { class: 'flex flex-col' });
      methods.forEach(function (m) {
        list.appendChild(h('a', { href: hashFor('payment-methods'), class: 'eco-log flex flex-wrap items-center gap-2 px-4 py-3 border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' },
          h('span', { class: 'font-medium flex-1', text: has('method.' + m.method) ? t('method.' + m.method) : m.method }),
          m.sandbox ? tag(t('ov.sandbox'), 'warning') : null,
          tag(m.configured ? t('pm.configured') : t('pm.notConfigured'), m.configured ? 'secondary' : 'error'),
          tag(m.isActive ? t('active') : t('inactive'), m.isActive ? 'success' : 'secondary')));
      });
      body.appendChild(h('div', {}, h('h2', { class: 'text-lg font-medium mb-3', text: t('ov.methods') }), card(list)));
    }
  }

  function load() {
    skeleton();
    api('GET', '/summary').then(render, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}
