ICONS['chat-quota'] = '<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/><path d="M8 9h8M8 13h4"/>';
PH_ICONS['chat-quota'] = '<path d="M216,48H40A16,16,0,0,0,24,64V224a15.84,15.84,0,0,0,9.25,14.5A16.05,16.05,0,0,0,40,240a15.89,15.89,0,0,0,10.25-3.78l.09-.07L83,208H216a16,16,0,0,0,16-16V64A16,16,0,0,0,216,48Zm0,144H80a8,8,0,0,0-5.23,1.95L40,224V64H216ZM88,112a8,8,0,0,1,8-8h64a8,8,0,0,1,0,16H96A8,8,0,0,1,88,112Zm0,32a8,8,0,0,1,8-8h32a8,8,0,0,1,0,16H96A8,8,0,0,1,88,144Z"/>';
ICONS['chat-quota-settings'] = '<path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6"/>';
ICONS['chat-quota-report'] = '<path d="M3 3v18h18"/><path d="M7 15l4-4 3 3 5-6"/>';

function cqLimit(v) {
  if (v == null) return h('span', { class: 'text-secondary-400', text: t('cq.inherit') });
  return h('span', { class: 'tabular-nums', text: fmtNum(v) });
}

RESOURCES['chat-quota'] = {
  api: 'base',
  path: 'chat-quota/rules',
  filters: [activeFilter()],
  columns: [
    { label: 'col.name', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: it.name }), h('div', { class: 'flex flex-wrap gap-1 mt-1', 'data-cq-tags': '' }, (it.tags || []).map(function (g) { return tag(i18nText(g.name) || g.key, 'primary'); }))); }, skeleton: '45%' },
    { label: 'col.cq.limits', render: function (it) {
      if (it.unlimited) return h('div', { 'data-cq-limits': '' }, tag(t('cq.unlimited'), 'success'));
      return h('div', { class: 'flex flex-col text-sm', 'data-cq-limits': '' },
        h('span', {}, h('span', { class: 'text-secondary-500', text: t('cq.lim.messages') + ': ' }), cqLimit(it.dailyMessages)),
        h('span', { class: 'text-xs' }, h('span', { class: 'text-secondary-500', text: t('cq.lim.conversations') + ': ' }), cqLimit(it.dailyConversations)));
    }, skeleton: '25%' },
    { label: 'col.fr.priority', render: function (it) { return h('div', { class: 'flex flex-col' }, numCell(it.priority), h('span', { class: 'text-xs text-secondary-500', text: t('fr.playersN', { n: fmtNum(it.players || 0) }) })); }, width: 110 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 100 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'name', type: 'text', label: 'f.cq.rule.name', hint: 'f.cq.rule.name.hint', required: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    { name: 'tagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.cq.rule.tags', hint: 'f.cq.rule.tags.hint', required: true, empty: [] },
    { name: 'priority', type: 'number', label: 'f.cq.rule.priority', hint: 'f.cq.rule.priority.hint', min: -1000, max: 1000 },
    { name: 'unlimited', type: 'bool', inline: true, label: 'f.cq.rule.unlimited', hint: 'f.cq.rule.unlimited.hint' },
    { type: 'heading', label: 'cq.h.ruleLimits', when: function (v) { return !v.unlimited; } },
    { type: 'row', when: function (v) { return !v.unlimited; }, fields: [
      { name: 'dailyMessages', type: 'number', label: 'f.cq.dailyMessages', hint: 'f.cq.rule.limit.hint', optional: true, min: 0, max: 100000 },
      { name: 'dailyConversations', type: 'number', label: 'f.cq.dailyConversations', optional: true, min: 0, max: 100000 }
    ] },
    { type: 'row', when: function (v) { return !v.unlimited; }, fields: [
      { name: 'friendDailyMessages', type: 'number', label: 'f.cq.friendDailyMessages', hint: 'f.cq.rule.friend.hint', optional: true, min: 0, max: 100000 },
      { name: 'friendDailyConversations', type: 'number', label: 'f.cq.friendDailyConversations', optional: true, min: 0, max: 100000 }
    ] }
  ],
  defaults: function () { return { name: '', tagIds: [], priority: 0, unlimited: false, dailyMessages: null, dailyConversations: null, friendDailyMessages: null, friendDailyConversations: null, isActive: true }; },
  fixPayload: function (p) {
    if (p.priority == null) p.priority = 0;
    if (p.unlimited) { p.dailyMessages = null; p.dailyConversations = null; p.friendDailyMessages = null; p.friendDailyConversations = null; }
    return p;
  }
};

function cqKeyRows(name, resource, labelKey, hintKey, addKey, emptyKey, max, when) {
  return { name: name, type: 'rows', label: labelKey, hint: hintKey, addLabel: addKey, emptyText: emptyKey, max: max, empty: [], when: when,
    newRow: function () { return { key: null }; },
    columns: [{ name: 'key', type: 'ref', resource: resource, valueKey: 'key', label: labelKey, required: true }] };
}

var CQ_SETTINGS_FIELDS = [
  { name: 'enabled', type: 'bool', inline: true, label: 'f.cq.enabled', hint: 'f.cq.enabled.hint' },
  { type: 'heading', label: 'cq.h.count' },
  { name: 'countMode', type: 'select', label: 'f.cq.countMode', hint: 'f.cq.countMode.hint', required: true, noPlaceholder: true, options: opts('cq.mode.', ['messages', 'conversations', 'both']) },
  { type: 'row', fields: [
    { name: 'dailyMessages', type: 'number', label: 'f.cq.dailyMessages', hint: 'f.cq.dailyMessages.hint', required: true, min: 0, max: 100000, when: function (v) { return v.countMode !== 'conversations'; } },
    { name: 'dailyConversations', type: 'number', label: 'f.cq.dailyConversations', hint: 'f.cq.dailyConversations.hint', required: true, min: 0, max: 100000, when: function (v) { return v.countMode !== 'messages'; } }
  ] },
  { name: 'conversationScope', type: 'select', label: 'f.cq.scope', hint: 'f.cq.scope.hint', required: true, noPlaceholder: true, options: opts('cq.scope.', ['day', 'ever']), when: function (v) { return v.countMode !== 'messages'; } },
  { type: 'row', fields: [
    { name: 'resetTime', type: 'custom', label: 'f.cq.resetTime', hint: 'f.cq.resetTime.hint', required: true, build: function (cur, change) { return timeOfDayInput(cur, change, { plain: true, name: 'resetTime', zoneLine: false }); }, check: function (v) { return v && !/^([01]\d|2[0-3]):[0-5]\d$/.test(toLatinDigits(String(v))) ? t('cq.err.time') : null; } },
    { name: 'timeZone', type: 'select', label: 'f.es.timezone', required: true, noPlaceholder: true, options: function () { return TIMEZONES.map(function (z) { return { value: z, label: z === 'UTC' ? 'UTC' : z.replace('_', ' ') }; }); } }
  ] },
  { type: 'heading', label: 'cq.h.scope' },
  { type: 'row', fields: [
    { name: 'countDirect', type: 'bool', inline: true, label: 'f.cq.direct', hint: 'f.cq.direct.hint' },
    { name: 'countGroups', type: 'bool', inline: true, label: 'f.cq.groups', hint: 'f.cq.groups.hint' },
    { name: 'countRooms', type: 'bool', inline: true, label: 'f.cq.rooms', hint: 'f.cq.rooms.hint' }
  ] },
  { name: 'replyFree', type: 'select', label: 'f.cq.replyFree', hint: 'f.cq.replyFree.hint', required: true, noPlaceholder: true, options: opts('cq.reply.', ['off', 'started', 'reply']) },
  { name: 'friendPolicy', type: 'select', label: 'f.cq.friendPolicy', hint: 'f.cq.friendPolicy.hint', required: true, noPlaceholder: true, options: opts('cq.friends.', ['same', 'separate', 'free']) },
  { type: 'row', when: function (v) { return v.friendPolicy === 'separate'; }, fields: [
    { name: 'friendDailyMessages', type: 'number', label: 'f.cq.friendDailyMessages', required: true, min: 0, max: 100000, when: function (v) { return v.countMode !== 'conversations'; } },
    { name: 'friendDailyConversations', type: 'number', label: 'f.cq.friendDailyConversations', required: true, min: 0, max: 100000, when: function (v) { return v.countMode !== 'messages'; } }
  ] },
  { type: 'heading', label: 'cq.h.packs' },
  { type: 'row', fields: [
    { name: 'packMeta', type: 'ref', resource: 'metas', valueKey: 'key', label: 'f.cq.packMeta', hint: 'f.cq.packMeta.hint', optional: true },
    { name: 'packExpiryDays', type: 'number', label: 'f.cq.packExpiry', hint: 'f.cq.packExpiry.hint', required: true, min: 0, max: 3650 }
  ] },
  { name: 'order', type: 'select', label: 'f.cq.order', hint: 'f.cq.order.hint', required: true, noPlaceholder: true, options: opts('cq.order.', ['free_first', 'packs_first']) },
  cqKeyRows('packProducts', 'products', 'f.cq.packProducts', 'f.cq.packProducts.hint', 'cq.packProducts.add', 'cq.packProducts.empty', 20),
  { name: 'bonusMeta', type: 'ref', resource: 'metas', valueKey: 'key', label: 'f.cq.bonusMeta', hint: 'f.cq.bonusMeta.hint', optional: true },
  { type: 'heading', label: 'cq.h.unlimited' },
  { name: 'matchPolicy', type: 'select', label: 'f.cq.matchPolicy', hint: 'f.cq.matchPolicy.hint', required: true, noPlaceholder: true, options: opts('fr.policy.', ['priority', 'generous']) },
  cqKeyRows('unlimitedItems', 'items', 'f.cq.unlimitedItems', 'f.cq.unlimitedItems.hint', 'cq.unlimitedItems.add', 'cq.unlimitedItems.empty', 50),
  { type: 'row', fields: [
    { name: 'unlimitedMeta', type: 'ref', resource: 'metas', valueKey: 'key', label: 'f.cq.unlimitedMeta', hint: 'f.cq.unlimitedMeta.hint', optional: true },
    { name: 'unlimitedMetaMode', type: 'select', label: 'f.cq.unlimitedMetaMode', hint: 'f.cq.unlimitedMetaMode.hint', required: true, noPlaceholder: true, options: opts('cq.metaMode.', ['positive', 'until']) }
  ] },
  { type: 'heading', label: 'cq.h.refuse' },
  tr('message', { label: 'f.cq.message', hint: 'f.cq.message.hint', multiline: true, rows: 2, max: 300, optional: true }),
  { name: 'notice', type: 'select', label: 'f.cq.notice', hint: 'f.cq.notice.hint', required: true, noPlaceholder: true, options: opts('cq.notice.', ['none', 'inbox']) }
];

function cqSummary(v) {
  var parts = [];
  if (!v.enabled) parts.push(t('cq.sum.off'));
  var mode = v.countMode || 'messages';
  if (mode !== 'conversations') parts.push(t('cq.sum.messages', { n: fmtNum(Number(v.dailyMessages) || 0) }));
  if (mode !== 'messages') parts.push(t('cq.sum.conversations', { n: fmtNum(Number(v.dailyConversations) || 0) }));
  parts.push(t('cq.sum.reset', { time: localDigits(toLatinDigits(String(v.resetTime || '00:00'))), zone: v.timeZone || '' }));
  if (v.countGroups || v.countRooms) parts.push(t('cq.sum.more'));
  if (v.packMeta) parts.push(t('cq.sum.packs', { key: v.packMeta }) + (Number(v.packExpiryDays) ? t('cq.sum.expiry', { n: fmtNum(Number(v.packExpiryDays)) }) : ''));
  else parts.push(t('cq.sum.noPacks'));
  var box = h('div', { class: 'mt-2 flex flex-col gap-3', 'data-cq-summary': '' }, notice(v.enabled ? 'info' : 'secondary', parts.join(t('fr.sum.sep'))));
  var found = v.detectedProducts || [];
  if (v.packMeta || (v.packProducts || []).length) {
    var list = h('div', { class: 'flex flex-col gap-1 text-sm', 'data-cq-products': '' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('cq.products.found') }));
    found.forEach(function (p) {
      list.appendChild(h('a', { href: hashFor('products', p.id), class: 'eco-link flex items-center gap-2', 'data-cq-product': p.key },
        h('span', { class: 'truncate', text: p.name || p.key }), mono(p.key), p.amount ? tag(t('cq.products.amount', { n: fmtNum(p.amount) }), 'success') : null));
    });
    if (!found.length) list.appendChild(h('span', { class: 'text-secondary-500', text: t('cq.products.none') }));
    box.appendChild(list);
  }
  if (v.updatedBy) box.appendChild(h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('up.savedBy', { who: v.updatedBy, when: fmtDate(v.updatedAt) }) }));
  return box;
}

function cqSettingsCall(method, path, body) {
  return baseApi(method, path, body).then(function (res) {
    var st = Object.assign({}, res || {});
    st.packProducts = (st.packProducts || []).map(function (k) { return { key: k }; });
    st.unlimitedItems = (st.unlimitedItems || []).map(function (k) { return { key: k }; });
    st.packMeta = st.packMeta || null;
    st.bonusMeta = st.bonusMeta || null;
    st.unlimitedMeta = st.unlimitedMeta || null;
    st.message = st.message || {};
    return st;
  });
}

function chatQuotaSettingsPage(container) {
  var frame = pageFrame(t('nav.chat-quota-settings'), t('desc.chat-quota-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(cqSettingsCall, '/chat-quota/settings', null, CQ_SETTINGS_FIELDS, {
    id: 'chat-quota',
    onchange: function (v, extra) { clear(extra).appendChild(cqSummary(v)); },
    fixPayload: function (p) {
      delete p.updatedAt;
      delete p.updatedBy;
      delete p.detectedProducts;
      p.resetTime = toLatinDigits(String(p.resetTime || '00:00')).trim();
      p.packProducts = (p.packProducts || []).map(function (r) { return r.key; }).filter(Boolean);
      p.unlimitedItems = (p.unlimitedItems || []).map(function (r) { return r.key; }).filter(Boolean);
      p.packMeta = p.packMeta || '';
      p.bonusMeta = p.bonusMeta || '';
      p.unlimitedMeta = p.unlimitedMeta || '';
      return p;
    }
  });
  frame.content.appendChild(page.el);
  page.load();
}

function chatQuotaReportPage(container, query) {
  var frame = pageFrame(t('nav.chat-quota-report'), t('desc.chat-quota-report'));
  clear(container).appendChild(frame.el);
  var state = { days: query.days || '30' };
  var stats = h('div', { class: 'eco-stat-grid mb-4', 'data-cq-stats': '' });
  var box = h('div', { class: 'flex flex-col gap-4' });
  var days = filterSelect(t('filter.of.days'), state.days, [1, 7, 30, 90, 365].map(function (n) { return { value: String(n), label: t('of.daysN', { n: n }) }; }), function (v) {
    state.days = v || '30';
    replaceHash(hashFor('chat-quota-report', null, state.days !== '30' ? { days: state.days } : {}));
    load();
  });
  frame.content.appendChild(h('div', { class: 'eco-filters flex flex-wrap items-center gap-2 pb-2' }, days));
  frame.content.appendChild(stats);
  frame.content.appendChild(box);
  function load() {
    clear(stats);
    for (var i = 0; i < 4; i++) stats.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem'), bar('3rem', 'h-7'))));
    clear(box).appendChild(card(h('div', { class: 'p-4 flex flex-col gap-2' }, bar('40%'), bar('70%'))));
    baseApi('GET', '/chat-quota/report' + qs({ days: state.days })).then(function (res) {
      var tt = res.totals || {};
      var pk = res.packs || {};
      clear(stats);
      stats.appendChild(statCard(t('cq.r.sent'), tt.sent, t('cq.r.senders', { n: fmtNum(tt.senders || 0) }), null));
      stats.appendChild(statCard(t('cq.r.blocked'), tt.blocked, t('cq.r.blockedPlayers', { n: fmtNum(tt.blockedPlayers || 0) }), null, { alert: (tt.blocked || 0) > 0 }));
      stats.appendChild(statCard(t('cq.r.fromPacks'), tt.fromPacks, null, null));
      stats.appendChild(statCard(t('cq.r.sold'), pk.sold, pk.toman ? t('toman', { n: pk.toman }) : null, hashFor('payments'), null));
      clear(box);
      var coins = Object.keys(pk.coins || {}).map(function (k) { return fmtNum(pk.coins[k]) + ' ' + k; });
      var granted = pk.granted || {};
      box.appendChild(section('cq.r.breakdown', dl([
        ['cq.r.daily', h('span', { class: 'tabular-nums', text: fmtNum(tt.daily || 0) })],
        ['cq.r.free', h('span', { class: 'tabular-nums', text: fmtNum(tt.free || 0) })],
        ['cq.r.unlimited', h('span', { class: 'tabular-nums', text: fmtNum(tt.unlimited || 0) })],
        ['cq.r.grantedMeta', h('span', { class: 'tabular-nums', 'data-cq-granted-meta': '', text: fmtNum(granted.meta || 0) })],
        ['cq.r.grantedAdmin', h('span', { class: 'tabular-nums', 'data-cq-granted-admin': '', text: fmtNum(granted.admin || 0) })],
        ['cq.r.coins', coins.length ? h('span', { dir: 'ltr', class: 'tabular-nums', text: coins.join(' · ') }) : null]
      ])));
      var daily = res.daily || [];
      if (daily.length) {
        var maxV = Math.max.apply(null, daily.map(function (d) { return Math.max(d.sent || 0, d.blocked || 0); }).concat([1]));
        var bars = h('div', { class: 'flex items-end gap-1 h-32 overflow-x-auto', 'data-cq-daily': '', role: 'img', 'aria-label': t('cq.r.chart') });
        daily.forEach(function (d) {
          bars.appendChild(h('div', { class: 'flex flex-col items-center justify-end gap-1 h-full', style: 'min-width:1.75rem', title: fmtDate(d.day + 'T12:00:00', false) + ' · ' + t('cq.r.sent') + ' ' + fmtNum(d.sent) + ' · ' + t('cq.r.blocked') + ' ' + fmtNum(d.blocked) },
            h('div', { class: 'flex items-end gap-px w-full h-full' },
              h('div', { class: 'flex-1 eco-fg-bar', style: 'height:' + Math.max(3, Math.round((d.sent || 0) * 100 / maxV)) + '%' }),
              h('div', { class: 'flex-1 eco-cq-bar-blocked', style: 'height:' + Math.max(d.blocked ? 3 : 0, Math.round((d.blocked || 0) * 100 / maxV)) + '%' })),
            h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: localDigits(String(d.day).slice(8, 10)) })));
        });
        box.appendChild(section('cq.r.chart', h('div', {}, bars,
          h('div', { class: 'flex gap-4 text-xs text-secondary-500 mt-2' },
            h('span', { class: 'inline-flex items-center gap-1' }, h('span', { class: 'inline-block w-3 h-3 rounded-sm eco-fg-bar' }), t('cq.r.sent')),
            h('span', { class: 'inline-flex items-center gap-1' }, h('span', { class: 'inline-block w-3 h-3 rounded-sm eco-cq-bar-blocked' }), t('cq.r.blocked'))))));
      }
      var top = res.topBlocked || [];
      var cols = [{ label: t('col.name'), skeleton: '60%' }, { label: t('cq.r.blocked'), width: 90 }];
      var sales = pk.products || [];
      var scol = [{ label: t('col.name'), skeleton: '50%' }, { label: t('cq.r.salesN'), width: 80 }, { label: t('cq.r.revenue'), width: 120 }];
      box.appendChild(h('div', { class: 'eco-detail-grid' },
        section('cq.r.topBlocked', top.length ? table(cols, top.map(function (r) { return { item: r, cells: [playerLink(r.userId, r.displayName || r.username), h('span', { class: 'tabular-nums', text: fmtNum(r.count) })] }; })) : h('p', { class: 'text-sm text-secondary-500 py-2', 'data-cq-none': 'blocked', text: t('cq.r.noneBlocked') })),
        section('cq.r.packSales', sales.length ? table(scol, sales.map(function (r) {
          var coin = Object.keys(r.coins || {}).map(function (k) { return fmtNum(r.coins[k]) + ' ' + k; }).join(' · ');
          return { item: r, cells: [h('a', { href: hashFor('products', r.productId), class: 'eco-link truncate', text: i18nText(r.name) || r.key }), h('span', { class: 'tabular-nums', text: fmtNum(r.count) }), h('span', { class: 'tabular-nums whitespace-nowrap', dir: 'auto', text: [r.toman ? t('toman', { n: r.toman }) : '', coin].filter(Boolean).join(' · ') || t('none') })] };
        })) : h('p', { class: 'text-sm text-secondary-500 py-2', 'data-cq-none': 'sales', text: (pk.packProducts || []).length ? t('cq.r.noSales') : t('cq.r.noPackProducts') }))));
    }, function (err) { clear(stats); clear(box).appendChild(errorBlock(err, load)); });
  }
  load();
}

CUSTOM_PAGES['chat-quota-settings'] = function (root) { chatQuotaSettingsPage(root); };
CUSTOM_PAGES['chat-quota-report'] = function (root, route) { chatQuotaReportPage(root, route.query || {}); };

var CQ_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadCqCard(id) {
  CQ_CARD.loading = true;
  return baseApi('GET', '/players/' + encodeURIComponent(id) + '/chat-quota' + qs({ lang: lang() })).then(function (res) {
    cqCardData(res);
  }, function (err) {
    CQ_CARD.loading = false;
    CQ_CARD.at = Date.now();
    CQ_CARD.err = err;
    schedule();
  });
}

function cqCardData(res) {
  CQ_CARD.loading = false;
  CQ_CARD.at = Date.now();
  CQ_CARD.data = res;
  CQ_CARD.err = null;
  var old = document.querySelector('[data-player-chat-quota]');
  if (old) old.removeAttribute('data-key');
  schedule();
}

function cqGrantDialog(id) {
  var amount = numberInput(10, { int: true, label: t('f.cq.grant.amount'), oninput: function () {} });
  var days = numberInput('', { int: true, label: t('f.cq.grant.days'), oninput: function () {} });
  var note = textInput('', { label: t('f.cq.grant.note') });
  amount.setAttribute('data-cq-grant', 'amount');
  days.setAttribute('data-cq-grant', 'days');
  note.setAttribute('data-cq-grant', 'note');
  var bad = h('div', {});
  return formDialog({
    id: 'chat-quota-grant',
    title: t('cq.grant.title'),
    text: t('cq.grant.text'),
    icon: 'plus',
    ok: t('cq.grant.ok'),
    body: h('div', {}, bad, field(t('f.cq.grant.amount'), amount, t('f.cq.grant.amount.hint')), field(t('f.cq.grant.days'), days, t('f.cq.grant.days.hint'), { optional: true }), field(t('f.cq.grant.note'), note, null, { optional: true })),
    validate: function () {
      clear(bad);
      var n = parseNumber(amount.value, true);
      var d = days.value.trim() === '' ? null : parseNumber(days.value, true);
      if (!n || n < 1 || n > 100000 || (d !== null && (isNaN(d) || d < 0 || d > 3650))) { bad.appendChild(h('div', { class: 'mb-4' }, notice('error', t('err.number')))); return false; }
      return true;
    },
    submit: function () {
      var d = days.value.trim() === '' ? undefined : parseNumber(days.value, true);
      return baseApi('POST', '/players/' + encodeURIComponent(id) + '/chat-quota/grant', { amount: parseNumber(amount.value, true), expiresInDays: d, note: note.value.trim() });
    }
  });
}

function cqConfirm(opts2) {
  return formDialog({ id: opts2.id, title: opts2.title, text: opts2.text, icon: opts2.icon || 'alert', ok: opts2.ok, danger: opts2.danger, body: null, submit: opts2.submit });
}

function cqPool(p, key) {
  var pool = p[key] || {};
  return t('cq.card.pool', { used: fmtNum(pool.used || 0), limit: fmtNum(pool.limit || 0) });
}

function playerChatQuotaCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-chat-quota]');
  if (!id || !token() || !aclCan('CHANNEL_MESSAGE', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (CQ_CARD.id !== id) { CQ_CARD.id = id; CQ_CARD.data = null; CQ_CARD.err = null; CQ_CARD.at = 0; }
  if (!CQ_CARD.loading && Date.now() - CQ_CARD.at > 20000) loadCqCard(id);
  var p = CQ_CARD.data;
  var anchor = grid.querySelector('[data-player-friend-gifts]') || grid.querySelector('[data-player-offers]') || grid.querySelector('[data-player-chance]') || grid.querySelector('[data-player-achievements]') || grid.querySelector('[data-player-items]') || grid.querySelector('[data-player-tasks]') || grid.querySelector('[data-player-activity]') || grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify([p.enabled, p.unlimited, p.freeLeft, p.packsLeft, p.today, (p.lots || []).map(function (l) { return l.id + l.remaining; }), (p.days || []).length]) : CQ_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-chat-quota': '', 'data-key': key, 'data-economy': 'player-chat-quota', 'data-l10n-skip': '', 'aria-labelledby': 'player-cq-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-cq-title', class: 'font-medium flex items-center gap-2' }, icon('chat-quota', 'h-4 w-4'), h('span', { text: t('cq.card') }));
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(CQ_CARD.err ? notice('error', errorText(CQ_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { CQ_CARD.at = 0; loadCqCard(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var write = aclCan('CHANNEL_MESSAGE', 1);
  var after = function (res) { if (res && res !== true) cqCardData(res); };
  var actions = h('div', { class: 'flex flex-wrap items-center gap-1' });
  if (write) {
    var grant = button({ variant: 'ghost', size: 'sm', icon: 'plus', label: t('cq.grant'), onclick: function () { cqGrantDialog(id).then(function (res) { if (res) { toast('success', t('cq.grant.done')); after(res); } }); } });
    grant.setAttribute('data-cq-action', 'grant');
    var reset = button({ variant: 'ghost', size: 'sm', icon: 'refresh', label: t('cq.reset'), onclick: function () {
      cqConfirm({ id: 'chat-quota-reset', title: t('cq.reset.title'), text: t('cq.reset.text'), icon: 'refresh', ok: t('cq.reset'), submit: function () { return baseApi('POST', '/players/' + encodeURIComponent(id) + '/chat-quota/reset', {}); } })
        .then(function (res) { if (res) { toast('success', t('cq.reset.done')); after(res); } });
    } });
    reset.setAttribute('data-cq-action', 'reset');
    actions.appendChild(grant);
    actions.appendChild(reset);
  }
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' }, title, actions));
  var today = p.today || {};
  var by = p.unlimited ? t('cq.by.' + (p.unlimitedBy || 'rule')) : '';
  body.appendChild(h('div', { class: 'flex flex-wrap gap-1 mb-3', 'data-cq-tags': '' },
    p.enabled ? null : tag(t('cq.off'), 'warning'),
    p.unlimited ? tag(t('cq.unlimited') + ' · ' + by, 'success') : null,
    p.rule ? tag(t('cq.card.rule', { name: p.rule.name }), 'primary') : null,
    p.bonus ? tag(t('cq.card.bonus', { n: fmtNum(p.bonus) }), 'success') : null,
    today.blocked ? tag(t('cq.card.blocked', { n: fmtNum(today.blocked) }), 'error') : null));
  var unit = p.countMode === 'conversations' ? 'conversations' : 'messages';
  body.appendChild(h('div', { class: 'eco-pc-grid mb-3' },
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('cq.card.free') }),
      h('div', { class: 'text-sm tabular-nums', 'data-cq-free': '' },
        h('span', { class: 'text-lg font-semibold', text: p.unlimited ? '∞' : fmtNum(p.freeLeft || 0) }),
        h('span', { class: 'text-secondary-500', text: ' · ' + cqPool(p, unit) })),
      p.countMode === 'both' ? h('div', { class: 'text-xs text-secondary-500 tabular-nums', text: t('cq.lim.conversations') + ': ' + cqPool(p, 'conversations') }) : null,
      p.friends ? h('div', { class: 'text-xs text-secondary-500 tabular-nums', text: t('cq.card.friends') + ': ' + cqPool(p.friends, unit) }) : null),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('cq.card.packs') }),
      h('div', { class: 'text-sm tabular-nums', 'data-cq-packs': '' }, h('span', { class: 'text-lg font-semibold', text: fmtNum(p.packsLeft || 0) })),
      p.packsExpireAt ? h('div', { class: 'text-xs text-secondary-500', text: t('cq.card.expires', { date: fmtDate(p.packsExpireAt, true) }) }) : null),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('cq.card.today') }),
      h('div', { class: 'text-sm tabular-nums', 'data-cq-today': '', text: t('cq.card.todayN', { sent: fmtNum(today.sent || 0), packs: fmtNum(today.fromPacks || 0) }) }),
      h('div', { class: 'text-xs text-secondary-500', text: t('cq.card.reset', { date: fmtDate(p.resetsAt, true) }) }))));
  var lots = h('div', { class: 'flex flex-col', 'data-cq-lots': '' });
  (p.lots || []).forEach(function (l) {
    var revoke = null;
    if (write && l.active) {
      revoke = button({ variant: 'link', size: 'sm', label: t('cq.revoke'), onclick: function () {
        cqConfirm({ id: 'chat-quota-revoke', title: t('cq.revoke.title'), text: t('cq.revoke.text', { n: fmtNum(l.remaining) }), danger: true, ok: t('cq.revoke'), submit: function () { return baseApi('POST', '/players/' + encodeURIComponent(id) + '/chat-quota/lots/' + encodeURIComponent(l.id) + '/revoke', {}); } })
          .then(function (res) { if (res) { toast('success', t('cq.revoke.done')); after(res); } });
      } });
      revoke.setAttribute('data-cq-revoke', l.id);
    }
    lots.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 py-1.5 px-1 text-sm eco-fg-row', 'data-cq-lot': l.id },
      tag(t('cq.source.' + l.source), l.source === 'admin' ? 'primary' : 'success'),
      h('span', { class: 'tabular-nums', text: t('cq.lot.left', { left: fmtNum(l.remaining), total: fmtNum(l.amount) }) }),
      l.note ? h('span', { class: 'text-secondary-500 truncate min-w-0', text: l.note }) : null,
      l.active ? null : tag(t(l.remaining ? 'cq.lot.expired' : 'cq.lot.used'), 'secondary'),
      h('span', { class: 'flex-1' }),
      h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: l.expiresAt ? t('cq.lot.until', { date: fmtDate(l.expiresAt, true) }) : t('cq.lot.forever') }),
      revoke));
  });
  if (!(p.lots || []).length) lots.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('cq.card.noLots') }));
  body.appendChild(progressSection('cq.card.lots', lots, h('a', { href: hashFor('chat-quota-report'), class: 'eco-link', text: t('cq.card.report') })));
  var dayRows = p.days || [];
  if (dayRows.length) {
    var tbl = h('div', { class: 'flex flex-col text-sm mt-3', 'data-cq-days': '' });
    dayRows.slice(0, 7).forEach(function (d) {
      tbl.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 py-1 px-1 eco-fg-row tabular-nums' },
        h('span', { class: 'w-28 text-secondary-500', text: fmtDate(d.day + 'T12:00:00', false) }),
        h('span', { text: t('cq.day.sent', { n: fmtNum(d.sent) }) }),
        d.fromPacks ? h('span', { class: 'text-secondary-500', text: t('cq.day.packs', { n: fmtNum(d.fromPacks) }) }) : null,
        d.blocked ? tag(t('cq.card.blocked', { n: fmtNum(d.blocked) }), 'error') : null));
    });
    body.appendChild(progressSection('cq.card.days', tbl));
  }
  place();
}
