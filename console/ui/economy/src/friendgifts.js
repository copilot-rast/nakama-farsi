ICONS['friend-gifts'] = '<path d="M11 14h2a2 2 0 1 0 0-4h-3c-.6 0-1.1.2-1.4.6L3 16"/><path d="m7 20 1.6-1.4c.3-.4.8-.6 1.4-.6h4c1.1 0 2.1-.4 2.8-1.2l4.6-4.4a2 2 0 0 0-2.75-2.91l-4.2 3.9"/><path d="m2 15 6 6"/><path d="M19.5 8.5c.7-.7 1.5-1.6 1.5-2.7A2.73 2.73 0 0 0 16 4a2.78 2.78 0 0 0-5 1.8c0 1.2.8 2 1.5 2.8L16 12Z"/>';
PH_ICONS['friend-gifts'] = '<path d="M230.33,141.06a24.43,24.43,0,0,0-18.61-4.77C230.5,117.33,240,98.48,240,80c0-26.47-21.29-48-47.46-48A47.58,47.58,0,0,0,156,48.75,47.58,47.58,0,0,0,119.46,32C93.29,32,72,53.53,72,80c0,11,3.24,21.69,10.06,33a31.87,31.87,0,0,0-14.75,8.4L44.69,144H16A16,16,0,0,0,0,160v40a16,16,0,0,0,16,16H120a7.93,7.93,0,0,0,1.94-.24l64-16a6.94,6.94,0,0,0,1.19-.4L226,182.82l.44-.2a24.6,24.6,0,0,0,3.93-41.56ZM119.46,48A31.15,31.15,0,0,1,148.6,67a8,8,0,0,0,14.8,0,31.15,31.15,0,0,1,29.14-19C209.59,48,224,62.65,224,80c0,19.51-15.79,41.58-45.66,63.9l-11.09,2.55A28,28,0,0,0,140,112H100.68C92.05,100.36,88,90.12,88,80,88,62.65,102.41,48,119.46,48ZM16,160H40v40H16Zm203.43,8.21-38,16.18L119,200H56V155.31l22.63-22.62A15.86,15.86,0,0,1,89.94,128H140a12,12,0,0,1,0,24H112a8,8,0,0,0,0,16h32a8.32,8.32,0,0,0,1.79-.2l67-15.41.31-.08a8.6,8.6,0,0,1,6.3,15.9Z"/>';
ICONS['friend-gift-rules'] = '<path d="M12 2l7 4v6c0 5-3 8-7 10-4-2-7-5-7-10V6z"/><path d="M9 12l2 2 4-4"/>';
ICONS['friend-gift-settings'] = '<path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6"/>';
ICONS['friend-gift-report'] = '<path d="M3 3v18h18"/><path d="M7 15l4-4 3 3 5-6"/>';

var FG_STATUSES = ['pending', 'delivered', 'declined', 'expired', 'cancelled', 'reversed', 'failed', 'paying', 'sending'];
var FG_STATUS_COLOR = { pending: 'warning', delivered: 'success', declined: 'secondary', expired: 'secondary', cancelled: 'secondary', reversed: 'error', failed: 'error', paying: 'primary', sending: 'primary' };
var FG_OUTCOMES = ['delivered', 'returned', 'refunded', 'sender_got', 'refund_due', 'kept'];
var FG_KINDS = ['item', 'currency', 'product'];

function fgWhatName(g) {
  if (g.kind === 'item') return i18nText(g.itemName) || g.itemKey;
  if (g.kind === 'currency') return t('amount.of', { n: fmtNum(g.amount), name: i18nText(g.currencyName) || g.currencyKey });
  return i18nText(g.productName) || g.productKey || '';
}

function fgWhat(g) {
  var key = g.kind === 'item' ? g.itemKey : g.kind === 'currency' ? g.currencyKey : g.productKey;
  return h('div', { class: 'min-w-0', 'data-fg-what': key || '' },
    h('div', { class: 'flex items-center gap-1.5 min-w-0' }, tag(t('fg.kind.' + g.kind), g.kind === 'product' ? 'primary' : g.kind === 'currency' ? 'warning' : 'secondary'), h('span', { class: 'font-medium truncate', text: fgWhatName(g) })),
    key ? h('div', { class: 'text-secondary-500' }, mono(key)) : null);
}

function fgStatus(g) {
  return h('div', { class: 'flex flex-wrap gap-1', 'data-fg-status': g.status },
    tag(t('fg.status.' + g.status), FG_STATUS_COLOR[g.status] || 'secondary'),
    g.outcome && g.outcome !== 'delivered' ? tag(t('fg.outcome.' + g.outcome), g.outcome === 'refund_due' ? 'warning' : 'secondary') : null,
    g.messageFlagged ? tag(t('fg.flagged'), 'warning') : null);
}

function fgPeople(g) {
  return h('div', { class: 'flex items-center gap-1.5 min-w-0 text-sm flex-wrap' },
    playerLink(g.senderId, g.senderName || g.senderUsername, 'font-medium'),
    h('span', { class: 'text-secondary-400 inline-block eco-fg-arrow', 'aria-hidden': 'true', text: '→' }),
    playerLink(g.recipientId, g.recipientName || g.recipientUsername, 'font-medium'));
}

function fgPaid(g) {
  if (g.source !== 'shop') return h('span', { class: 'text-sm text-secondary-500', text: t('fg.source.inventory') });
  var parts = [];
  if (g.method === 'in_app' && g.coinAmount != null) parts.push(fmtNum(g.coinAmount) + ' ' + (g.coinCurrency || ''));
  else if (g.finalToman != null) parts.push(t('toman', { n: g.finalToman }));
  return h('div', { class: 'flex flex-col text-sm' },
    h('span', { class: 'tabular-nums whitespace-nowrap', dir: 'auto', text: parts.join(' ') || t('none') }),
    g.method ? h('span', { class: 'text-xs text-secondary-500', text: t('method.' + g.method) }) : null);
}

LISTS['friend-gifts'] = {
  filters: [
    { name: 'status', label: 'filter.status', options: opts('fg.status.', FG_STATUSES) },
    { name: 'kind', label: 'filter.fg.kind', options: opts('fg.kind.', FG_KINDS) },
    { name: 'source', label: 'filter.fg.source', options: opts('fg.source.', ['inventory', 'shop']) },
    { name: 'outcome', label: 'filter.fg.outcome', options: opts('fg.outcome.', FG_OUTCOMES) },
    { name: 'userId', label: 'filter.playerId', type: 'text' }
  ],
  columns: [
    { label: 'col.created', render: function (g) { return dateCell(g.createdAt); }, width: 160 },
    { label: 'col.fg.people', render: fgPeople, skeleton: '45%' },
    { label: 'col.fg.what', render: fgWhat, skeleton: '35%' },
    { label: 'col.fg.paid', render: fgPaid, width: 140 },
    { label: 'col.status', render: fgStatus, width: 170 }
  ]
};

CUSTOM_PAGES['friend-gifts'] = function (root, route, ctl, ctx) {
  if (route.id) friendGiftDetail(root, route.id);
  else ctl.list = listPage(root, 'friend-gifts', route.query || {}, ctx);
};

function fgReverseModal(g, done) {
  var returnTo = 'sender';
  var choices = ['sender'].concat(g.kind === 'product' && g.method === 'in_app' ? ['refund'] : [], ['none']);
  var select = selectInput(returnTo, choices.map(function (c) { return { value: c, label: t('fg.returnTo.' + c + (c === 'sender' && g.kind === 'product' ? '.product' : '')) }; }), function (v) { returnTo = v || 'sender'; }, { label: t('fg.returnTo'), placeholder: false });
  select.setAttribute('data-fg-return', '');
  var note = textInput('', { multiline: true, rows: 2, label: t('fg.reverse.note'), placeholder: t('fg.reverse.note.ph') });
  note.setAttribute('data-fg-note', '');
  var body = h('div', { class: 'mt-3 flex flex-col gap-3' }, field(t('fg.returnTo'), select, t('fg.returnTo.hint')), field(t('fg.reverse.note'), note, null, { optional: true }));
  confirmModal({ title: t('fg.reverse.title'), text: t(g.status === 'pending' ? 'fg.reverse.text.pending' : 'fg.reverse.text', { what: fgWhatName(g), name: g.recipientName || g.recipientUsername || '' }), body: body, ok: t('fg.reverse'), danger: true }).then(function (yes) {
    if (!yes) return;
    api('POST', '/friend-gifts/' + encodeURIComponent(g.id) + '/reverse', { returnTo: returnTo, note: note.value.trim() }).then(function () {
      toast('success', t('fg.reverse.done'));
      done();
    }, function (err) { toast('error', errorText(err)); });
  });
}

function friendGiftDetail(container, id) {
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('friend-gifts', null, lastQuery['friend-gifts'] || {}); } });
  var frame = pageFrame(t('fg.title'), null, null);
  frame.actions.appendChild(backBtn);
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-detail-grid', 'data-friend-gift': id });
  frame.content.appendChild(grid);

  function render(g) {
    clear(grid);
    var actions = h('div', { class: 'flex flex-wrap gap-2' });
    var reversible = aclCan('HIRO_ECONOMY', 1) && (g.status === 'pending' || (g.status === 'delivered' && g.outcome === 'delivered'));
    if (reversible) {
      var rev = button({ variant: 'danger', size: 'page', icon: 'refresh', label: t('fg.reverse'), onclick: function () { fgReverseModal(g, load); } });
      rev.setAttribute('data-fg-reverse', '');
      actions.appendChild(rev);
    }
    clear(frame.actions).appendChild(actions);
    frame.actions.appendChild(backBtn);
    grid.appendChild(section('col.status', dl([
      ['col.status', fgStatus(g)],
      ['fg.delivery', tag(t('fg.delivery.' + g.delivery), 'secondary')],
      ['fg.outcome', g.outcome ? h('span', { text: t('fg.outcome.' + g.outcome) }) : null],
      ['fg.error', g.error ? h('span', { class: 'text-error-600 dark:text-error-400', text: g.error }) : null],
      ['fg.expiresAt', g.status === 'pending' && g.expiresAt ? dateCell(g.expiresAt) : null],
      ['fg.deliveredAt', g.deliveredAt ? dateCell(g.deliveredAt) : null],
      ['fg.respondedAt', g.respondedAt ? dateCell(g.respondedAt) : null],
      ['col.created', dateCell(g.createdAt)]
    ])));
    grid.appendChild(section('fg.people', dl([
      ['fg.sender', playerLink(g.senderId, g.senderName || g.senderUsername)],
      ['fg.recipient', playerLink(g.recipientId, g.recipientName || g.recipientUsername)],
      ['fg.message', g.message ? h('span', { class: 'break-words', dir: 'auto', text: g.message }) : null],
      ['fg.request', g.requestId ? mono(g.requestId) : null]
    ])));
    grid.appendChild(section('col.fg.what', dl([
      ['col.fg.what', fgWhat(g)],
      ['fg.source', tag(t('fg.source.' + g.source), 'secondary')],
      ['fg.upgrade', g.kind === 'item' && g.itemState && g.itemState.upgradeLevel ? h('span', { class: 'tabular-nums', text: fmtNum(g.itemState.upgradeLevel) }) : null],
      ['col.fg.paid', g.source === 'shop' ? fgPaid(g) : null],
      ['fg.payment', g.paymentId ? h('a', { href: hashFor('payments', g.paymentId), class: 'eco-link font-mono text-sm', dir: 'ltr', text: String(g.paymentId).slice(0, 8) }) : null],
      ['fg.paymentStatus', g.paymentStatus ? tag(t('status.' + g.paymentStatus), STATUS_COLOR[g.paymentStatus] || 'secondary') : null]
    ])));
    if (g.status === 'reversed') {
      grid.appendChild(section('fg.reversed', dl([
        ['fg.reversedBy', h('span', { text: g.reversedBy || t('none') })],
        ['fg.reversedAt', g.reversedAt ? dateCell(g.reversedAt) : null],
        ['fg.reverse.note', g.reverseNote ? h('span', { class: 'break-words', text: g.reverseNote }) : null]
      ])));
    }
    var trail = g.trail || [];
    var list = h('ol', { class: 'flex flex-col', 'data-fg-trail': '' });
    trail.forEach(function (e) {
      list.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 py-1.5 text-sm border-b border-secondary-200 dark:border-secondary-700' },
        tag(has('fg.trail.' + e.action) ? t('fg.trail.' + e.action) : e.action, 'secondary'),
        playerLink(e.userId, e.userId === g.senderId ? (g.senderName || g.senderUsername) : (g.recipientName || g.recipientUsername)),
        h('span', { class: 'text-xs text-secondary-500', text: t('fg.by.' + (e.actor || 'system')) }),
        h('span', { class: 'flex-1' }),
        h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: fmtDate(e.createdAt) })));
    });
    if (!trail.length) list.appendChild(h('li', { class: 'text-sm text-secondary-500 py-2', text: t('fg.trail.empty') }));
    var sec = section('fg.trail', list);
    sec.classList.add('eco-span-all');
    grid.appendChild(sec);
  }

  function load() {
    clear(grid);
    [5, 4, 4].forEach(function () { grid.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-2' }, bar('8rem', 'h-4'), bar('60%'), bar('45%'), bar('70%')))); });
    api('GET', '/friend-gifts/' + encodeURIComponent(id)).then(render, function (err) { clear(grid).appendChild(h('div', { class: 'eco-span-all' }, errorBlock(err, load))); });
  }
  load();
}

RESOURCES['friend-gift-rules'] = {
  filters: [activeFilter()],
  columns: [
    { label: 'col.name', render: function (r) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: r.name }), h('div', { class: 'text-xs text-secondary-500', text: t('fg.rule.players', { n: fmtNum(r.players || 0) }) })); }, skeleton: '30%' },
    { label: 'col.fg.limits', render: function (r) {
      function lim(v) { return v == null ? t('fg.rule.inherit') : v === 0 ? t('fg.unlimited') : fmtNum(v); }
      return h('div', { class: 'flex flex-col text-sm tabular-nums', 'data-fg-rule-limits': '' },
        h('span', { text: t('fg.lim.sender') + ': ' + lim(r.perSenderDaily) }),
        h('span', { class: 'text-xs text-secondary-500', text: t('fg.lim.recipient') + ': ' + lim(r.perRecipientDaily) + ' · ' + t('fg.lim.pair') + ': ' + lim(r.perPairDaily) }));
    }, skeleton: '30%' },
    { label: 'col.fr.priority', render: function (r) { return h('span', { class: 'tabular-nums', text: fmtNum(r.priority) }); }, width: 80 },
    { label: 'col.status', render: function (r) { return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(r.isActive), r.canSend ? null : tag(t('fg.rule.noSend'), 'error')); }, width: 150 }
  ],
  fields: [
    { name: 'name', type: 'text', label: 'f.fg.rule.name', hint: 'f.fg.rule.name.hint', required: true },
    { name: 'tagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.fg.rule.tags', hint: 'f.fg.rule.tags.hint', required: true, empty: [] },
    { name: 'priority', type: 'number', label: 'f.fg.rule.priority', hint: 'f.fg.rule.priority.hint', min: -1000, max: 1000 },
    { type: 'heading', label: 'fg.h.limits' },
    { type: 'row', fields: [
      { name: 'perSenderDaily', type: 'number', label: 'f.fg.perSender', hint: 'f.fg.rule.limit.hint', min: 0, max: 10000, optional: true },
      { name: 'perRecipientDaily', type: 'number', label: 'f.fg.perRecipient', min: 0, max: 10000, optional: true }
    ] },
    { name: 'perPairDaily', type: 'number', label: 'f.fg.perPair', min: 0, max: 10000, optional: true },
    { name: 'canSend', type: 'bool', inline: true, label: 'f.fg.rule.canSend', hint: 'f.fg.rule.canSend.hint' },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
  ],
  defaults: function () { return { name: '', tagIds: [], priority: 0, perSenderDaily: null, perRecipientDaily: null, perPairDaily: null, canSend: true, isActive: true }; },
  fixPayload: function (p) { if (p.priority == null) p.priority = 0; return p; },
  afterSave: function () { forgetRefs('friend-gift-rules'); }
};

function fgYesNo() { return [{ value: 'true', label: t('yes') }, { value: 'false', label: t('no') }]; }

var FG_SETTINGS_FIELDS = [
  { name: 'enabled', type: 'bool', inline: true, label: 'f.fg.enabled', hint: 'f.fg.enabled.hint' },
  { type: 'heading', label: 'fg.h.who' },
  { type: 'row', fields: [
    { name: 'recipients', type: 'select', label: 'f.fg.recipients', hint: 'f.fg.recipients.hint', required: true, noPlaceholder: true, options: opts('fg.recipients.', ['friends', 'any']) },
    { name: 'minFriendDays', type: 'number', label: 'f.fg.minFriendDays', hint: 'f.fg.minFriendDays.hint', required: true, min: 0, max: 3650, when: function (v) { return v.recipients !== 'any'; } }
  ] },
  { name: 'minMutualFriends', type: 'number', label: 'f.fg.minMutual', hint: 'f.fg.minMutual.hint', required: true, min: 0, max: 1000 },
  { type: 'row', fields: [
    { name: 'minSenderAccountDays', type: 'number', label: 'f.fg.minSenderDays', required: true, min: 0, max: 3650 },
    { name: 'minRecipientAccountDays', type: 'number', label: 'f.fg.minRecipientDays', required: true, min: 0, max: 3650 }
  ] },
  { type: 'row', fields: [
    { name: 'minSenderLevel', type: 'number', label: 'f.fg.minLevel', hint: 'f.fg.minLevel.hint', required: true, min: 0, max: 10000 },
    { name: 'levelTableId', type: 'ref', resource: 'level-tables', label: 'f.of.levelTable', hint: 'f.of.levelTable.hint', optional: true }
  ] },
  { type: 'heading', label: 'fg.h.limits' },
  { type: 'row', fields: [
    { name: 'perSenderDaily', type: 'number', label: 'f.fg.perSender', hint: 'f.fg.zero', required: true, min: 0, max: 10000 },
    { name: 'perRecipientDaily', type: 'number', label: 'f.fg.perRecipient', required: true, min: 0, max: 10000 }
  ] },
  { name: 'perPairDaily', type: 'number', label: 'f.fg.perPair', hint: 'f.fg.perPair.hint', required: true, min: 0, max: 10000 },
  { name: 'matchPolicy', type: 'select', label: 'f.fg.matchPolicy', hint: 'f.fg.matchPolicy.hint', required: true, noPlaceholder: true, options: opts('fg.match.', ['priority', 'generous']) },
  { type: 'row', fields: [
    { name: 'sendBonusMeta', type: 'ref', resource: 'metas', valueKey: 'key', label: 'f.fg.sendBonus', hint: 'f.fg.sendBonus.hint', optional: true },
    { name: 'receiveBonusMeta', type: 'ref', resource: 'metas', valueKey: 'key', label: 'f.fg.receiveBonus', optional: true }
  ] },
  { type: 'heading', label: 'fg.h.inventory' },
  { name: 'inventoryEnabled', type: 'bool', inline: true, label: 'f.fg.inventory', hint: 'f.fg.inventory.hint' },
  { name: 'itemsDefault', type: 'bool', inline: true, label: 'f.fg.itemsDefault', hint: 'f.fg.itemsDefault.hint', when: function (v) { return v.inventoryEnabled; } },
  { name: 'itemTypes', type: 'rows', label: 'f.fg.itemTypes', addLabel: 'fg.itemTypes.add', emptyText: 'fg.itemTypes.empty', max: 100, empty: [], when: function (v) { return v.inventoryEnabled; },
    newRow: function () { return { typeId: null, giftable: 'false' }; },
    columns: [
      { name: 'typeId', type: 'ref', resource: 'categories', api: baseApi, filter: { keyType: 'item_type' }, label: 'f.it.type', required: true },
      { name: 'giftable', type: 'select', label: 'f.fg.giftable', options: fgYesNo }
    ] },
  { type: 'row', when: function (v) { return v.inventoryEnabled; }, fields: [
    { name: 'equipped', type: 'select', label: 'f.fg.equipped', hint: 'f.fg.equipped.hint', required: true, noPlaceholder: true, options: opts('fg.equipped.', ['block', 'unequip']) },
    { name: 'upgradeLevel', type: 'select', label: 'f.fg.upgrade', hint: 'f.fg.upgrade.hint', required: true, noPlaceholder: true, options: opts('fg.upgrade.', ['keep', 'reset']) }
  ] },
  { type: 'row', when: function (v) { return v.inventoryEnabled; }, fields: [
    { name: 'regift', type: 'select', label: 'f.fg.regift', hint: 'f.fg.regift.hint', required: true, noPlaceholder: true, options: opts('fg.regift.', ['block', 'allow']) },
    { name: 'itemRequirements', type: 'select', label: 'f.fg.requirements', hint: 'f.fg.requirements.hint', required: true, noPlaceholder: true, options: opts('fg.requirements.', ['check', 'ignore']) }
  ] },
  { name: 'currencyEnabled', type: 'bool', inline: true, label: 'f.fg.currency', hint: 'f.fg.currency.hint' },
  { name: 'currencies', type: 'rows', label: 'f.fg.currencies', addLabel: 'fg.currencies.add', emptyText: 'fg.currencies.empty', max: 20, empty: [], when: function (v) { return v.currencyEnabled; },
    newRow: function () { return { currencyKey: null, min: 1, max: 0 }; },
    columns: [
      { name: 'currencyKey', type: 'ref', resource: 'currencies', valueKey: 'key', label: 'f.coinPrice.currency', required: true },
      { name: 'min', type: 'number', label: 'f.fg.min', min: 0 },
      { name: 'max', type: 'number', label: 'f.fg.max', min: 0 }
    ] },
  { type: 'heading', label: 'fg.h.shop' },
  { name: 'shopEnabled', type: 'bool', inline: true, label: 'f.fg.shop', hint: 'f.fg.shop.hint' },
  { name: 'productsDefault', type: 'bool', inline: true, label: 'f.fg.productsDefault', hint: 'f.fg.productsDefault.hint', when: function (v) { return v.shopEnabled; } },
  { name: 'methods', type: 'multi', label: 'f.fg.methods', hint: 'f.fg.methods.hint', optional: true, empty: [], when: function (v) { return v.shopEnabled; }, options: opts('fg.method.', ['in_app', 'gateway', 'store']) },
  { type: 'row', when: function (v) { return v.shopEnabled; }, fields: [
    { name: 'partialOwned', type: 'select', label: 'f.fg.partial', hint: 'f.fg.partial.hint', required: true, noPlaceholder: true, options: opts('fg.partial.', ['block', 'allow']) },
    { name: 'stackableProducts', type: 'select', label: 'f.fg.stackable', hint: 'f.fg.stackable.hint', required: true, noPlaceholder: true, options: opts('fg.stackable.', ['allow', 'block']) }
  ] },
  { name: 'productRules', type: 'select', label: 'f.fg.productRules', hint: 'f.fg.productRules.hint', required: true, noPlaceholder: true, when: function (v) { return v.shopEnabled; }, options: opts('fg.productRules.', ['recipient', 'sender', 'both', 'none']) },
  { name: 'discountCodes', type: 'bool', inline: true, label: 'f.fg.codes', hint: 'f.fg.codes.hint', when: function (v) { return v.shopEnabled; } },
  { type: 'heading', label: 'fg.h.delivery' },
  { type: 'row', fields: [
    { name: 'delivery', type: 'select', label: 'f.fg.delivery', hint: 'f.fg.delivery.hint', required: true, noPlaceholder: true, options: opts('fg.delivery.', ['instant', 'accept']) },
    { name: 'expiryHours', type: 'number', label: 'f.fg.expiry', hint: 'f.fg.expiry.hint', required: true, min: 1, max: 8760, when: function (v) { return v.delivery === 'accept'; } }
  ] },
  { name: 'onExpiry', type: 'select', label: 'f.fg.onExpiry', hint: 'f.fg.onExpiry.hint', required: true, noPlaceholder: true, when: function (v) { return v.delivery === 'accept'; }, options: opts('fg.onExpiry.', ['return', 'accept']) },
  { type: 'row', fields: [
    { name: 'coinReturn', type: 'select', label: 'f.fg.coinReturn', hint: 'f.fg.coinReturn.hint', required: true, noPlaceholder: true, options: opts('fg.coinReturn.', ['refund', 'sender', 'none']) },
    { name: 'moneyReturn', type: 'select', label: 'f.fg.moneyReturn', hint: 'f.fg.moneyReturn.hint', required: true, noPlaceholder: true, options: opts('fg.moneyReturn.', ['sender', 'refund_due', 'none']) }
  ] },
  { name: 'senderCancel', type: 'bool', inline: true, label: 'f.fg.cancel', hint: 'f.fg.cancel.hint' },
  { type: 'heading', label: 'fg.h.message' },
  { name: 'messageEnabled', type: 'bool', inline: true, label: 'f.fg.message', hint: 'f.fg.message.hint' },
  { type: 'row', when: function (v) { return v.messageEnabled; }, fields: [
    { name: 'messageMaxLength', type: 'number', label: 'f.fg.messageMax', required: true, min: 1, max: 500 },
    { name: 'messageFilter', type: 'select', label: 'f.fg.filter', hint: 'f.fg.filter.hint', required: true, noPlaceholder: true, options: opts('fg.filter.', ['rules', 'block', 'off']) }
  ] },
  { type: 'heading', label: 'fg.h.notify' },
  { name: 'notifyChannel', type: 'select', label: 'f.fg.notify', hint: 'f.fg.notify.hint', required: true, noPlaceholder: true, options: opts('fg.notify.', ['in_app', 'push_in_app', 'none']) },
  { name: 'socketNotify', type: 'bool', inline: true, label: 'f.fg.socket', hint: 'f.fg.socket.hint' },
  { name: 'notifySender', type: 'bool', inline: true, label: 'f.fg.notifySender', hint: 'f.fg.notifySender.hint' }
];

function fgSettingsCall(method, path, body) {
  return api(method, path, body).then(function (res) {
    var st = Object.assign({}, (res && res.settings) || {});
    st.itemTypes = (st.itemTypes || []).map(function (r) { return { typeId: r.typeId, giftable: r.giftable ? 'true' : 'false' }; });
    st.currencies = st.currencies || [];
    st.methods = st.methods || [];
    st.levelTableId = st.levelTableId || null;
    st.sendBonusMeta = st.sendBonusMeta || null;
    st.receiveBonusMeta = st.receiveBonusMeta || null;
    return st;
  });
}

function friendGiftSettingsPage(container) {
  var frame = pageFrame(t('nav.friend-gift-settings'), t('desc.friend-gift-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(fgSettingsCall, '/friend-gift-settings', null, FG_SETTINGS_FIELDS, { id: 'friend-gifts', fixPayload: function (p) {
    delete p.updatedAt;
    p.itemTypes = (p.itemTypes || []).map(function (r) { return { typeId: r.typeId, giftable: r.giftable === true || r.giftable === 'true' }; });
    p.currencies = (p.currencies || []).map(function (r) { return { currencyKey: r.currencyKey, min: Number(r.min || 0), max: Number(r.max || 0) }; });
    p.methods = p.methods || [];
    p.levelTableId = p.levelTableId || '';
    p.sendBonusMeta = p.sendBonusMeta || '';
    p.receiveBonusMeta = p.receiveBonusMeta || '';
    return p;
  } });
  frame.content.appendChild(page.el);
  page.load();
}

function friendGiftReportPage(container, query) {
  var frame = pageFrame(t('nav.friend-gift-report'), t('desc.friend-gift-report'));
  clear(container).appendChild(frame.el);
  var state = { days: query.days || '30' };
  var stats = h('div', { class: 'eco-stat-grid mb-4', 'data-fg-stats': '' });
  var box = h('div', { class: 'flex flex-col gap-4' });
  var days = filterSelect(t('filter.of.days'), state.days, [1, 7, 30, 90, 365].map(function (n) { return { value: String(n), label: t('of.daysN', { n: n }) }; }), function (v) {
    state.days = v || '30';
    replaceHash(hashFor('friend-gift-report', null, state.days !== '30' ? { days: state.days } : {}));
    load();
  });
  frame.content.appendChild(h('div', { class: 'eco-filters flex flex-wrap items-center gap-2 pb-2' }, days));
  frame.content.appendChild(stats);
  frame.content.appendChild(box);
  function topTable(titleKey, rows, cell) {
    var cols = [{ label: t('col.name'), skeleton: '60%' }, { label: t('fg.count'), width: 90 }];
    return section(titleKey, rows.length ? table(cols, rows.map(function (r) { return { item: r, cells: [cell(r), h('span', { class: 'tabular-nums', text: fmtNum(r.count) })] }; })) : h('p', { class: 'text-sm text-secondary-500 py-2', text: t('fg.report.none') }));
  }
  function load() {
    clear(stats);
    for (var i = 0; i < 5; i++) stats.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem'), bar('3rem', 'h-7'))));
    clear(box).appendChild(card(h('div', { class: 'p-4 flex flex-col gap-2' }, bar('40%'), bar('70%'), bar('55%'))));
    api('GET', '/friend-gift-report' + qs({ days: state.days })).then(function (res) {
      var tt = res.totals || {};
      var st = res.statuses || {};
      clear(stats);
      stats.appendChild(statCard(t('fg.r.sent'), tt.sent, t('fg.r.people', { s: fmtNum(tt.senders || 0), r: fmtNum(tt.recipients || 0) }), null));
      stats.appendChild(statCard(t('fg.status.delivered'), st.delivered, null, hashFor('friend-gifts', null, { status: 'delivered' })));
      stats.appendChild(statCard(t('fg.status.pending'), st.pending, null, hashFor('friend-gifts', null, { status: 'pending' })));
      stats.appendChild(statCard(t('fg.r.returned'), (st.declined || 0) + (st.expired || 0) + (st.cancelled || 0), null, null));
      stats.appendChild(statCard(t('fg.status.reversed'), st.reversed, null, hashFor('friend-gifts', null, { status: 'reversed' }), { alert: (st.reversed || 0) > 0 }));
      stats.appendChild(statCard(t('fg.r.revenue'), tt.revenueToman, t('toman', { n: '' }).trim(), null));
      clear(box);
      var kinds = res.kinds || {};
      var coins = Object.keys(tt.coins || {}).map(function (k) { return fmtNum(tt.coins[k]) + ' ' + k; });
      var moved = Object.keys(tt.transferred || {}).map(function (k) { return fmtNum(tt.transferred[k]) + ' ' + k; });
      box.appendChild(section('fg.r.breakdown', dl([
        ['fg.kind.item', h('span', { class: 'tabular-nums', text: fmtNum(kinds.item || 0) })],
        ['fg.kind.currency', h('span', { class: 'tabular-nums', text: fmtNum(kinds.currency || 0) })],
        ['fg.kind.product', h('span', { class: 'tabular-nums', text: fmtNum(kinds.product || 0) })],
        ['fg.r.coins', coins.length ? h('span', { dir: 'ltr', class: 'tabular-nums', text: coins.join(' · ') }) : null],
        ['fg.r.moved', moved.length ? h('span', { dir: 'ltr', class: 'tabular-nums', text: moved.join(' · ') }) : null],
        ['fg.status.failed', st.failed ? h('span', { class: 'tabular-nums', text: fmtNum(st.failed) }) : null]
      ])));
      var person = function (r) { return playerLink(r.userId, r.displayName || r.username); };
      box.appendChild(h('div', { class: 'eco-detail-grid' },
        topTable('fg.r.topSenders', res.topSenders || [], person),
        topTable('fg.r.topRecipients', res.topRecipients || [], person),
        topTable('fg.r.topThings', res.topThings || [], function (r) { return h('div', { class: 'flex items-center gap-1.5 min-w-0' }, tag(t('fg.kind.' + r.kind), 'secondary'), h('span', { class: 'truncate', text: i18nText(r.name) || r.key }), mono(r.key)); })));
      var daily = res.daily || [];
      if (daily.length) {
        var maxV = Math.max.apply(null, daily.map(function (d) { return d.sent || 0; }).concat([1]));
        var bars = h('div', { class: 'flex items-end gap-1 h-32 overflow-x-auto', 'data-fg-daily': '', role: 'img', 'aria-label': t('fg.r.daily') });
        daily.forEach(function (d) {
          bars.appendChild(h('div', { class: 'flex flex-col items-center justify-end gap-1 h-full', style: 'min-width:1.5rem', title: fmtDate(d.day) + ' · ' + fmtNum(d.sent) },
            h('div', { class: 'w-full eco-fg-bar', style: 'height:' + Math.max(4, Math.round((d.sent || 0) * 100 / maxV)) + '%' }),
            h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: localDigits(String(d.day).slice(8, 10)) })));
        });
        box.appendChild(section('fg.r.daily', bars));
      }
    }, function (err) { clear(stats); clear(box).appendChild(errorBlock(err, load)); });
  }
  load();
}

CUSTOM_PAGES['friend-gift-settings'] = function (root) { friendGiftSettingsPage(root); };
CUSTOM_PAGES['friend-gift-report'] = function (root, route) { friendGiftReportPage(root, route.query || {}); };

(function giftableFields() {
  function add(def, after) {
    if (!def || def.fields.some(function (f) { return f.name === 'giftableChoice'; })) return;
    var at = def.fields.length;
    def.fields.forEach(function (f, i) { if (f.name === after) at = i; });
    def.fields.splice(at, 0, { name: 'giftableChoice', type: 'select', label: 'f.fg.giftableChoice', hint: 'f.fg.giftableChoice.hint', required: true, noPlaceholder: true,
      options: function () { return [{ value: 'default', label: t('fg.giftable.default') }, { value: 'true', label: t('fg.giftable.yes') }, { value: 'false', label: t('fg.giftable.no') }]; } });
    var toForm = def.toForm;
    def.toForm = function (v) { v = toForm ? toForm(v) : v; v.giftableChoice = v.giftable === true ? 'true' : v.giftable === false ? 'false' : 'default'; return v; };
    var defaults = def.defaults;
    def.defaults = function (q) { var v = defaults(q); v.giftableChoice = 'default'; return v; };
    var fix = def.fixPayload;
    def.fixPayload = function (p) {
      p = fix ? fix(p) : p;
      p.giftable = p.giftableChoice === 'true' ? true : p.giftableChoice === 'false' ? false : null;
      delete p.giftableChoice;
      return p;
    };
  }
  add(RESOURCES.items, 'sort');
  add(RESOURCES.products, 'sort');
})();

var FG_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function loadFgCard(id) {
  FG_CARD.loading = true;
  return api('GET', '/users/' + encodeURIComponent(id) + '/friend-gifts').then(function (res) {
    FG_CARD.loading = false;
    FG_CARD.at = Date.now();
    FG_CARD.data = res;
    FG_CARD.err = null;
    var old = document.querySelector('[data-player-friend-gifts]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    FG_CARD.loading = false;
    FG_CARD.at = Date.now();
    FG_CARD.err = err;
    schedule();
  });
}

function playerFriendGiftsCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-friend-gifts]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (FG_CARD.id !== id) { FG_CARD.id = id; FG_CARD.data = null; FG_CARD.err = null; FG_CARD.at = 0; }
  if (!FG_CARD.loading && Date.now() - FG_CARD.at > 20000) loadFgCard(id);
  var p = FG_CARD.data;
  var anchor = grid.querySelector('[data-player-offers]') || grid.querySelector('[data-player-chance]') || grid.querySelector('[data-player-achievements]') || grid.querySelector('[data-player-items]') || grid.querySelector('[data-player-tasks]') || grid.querySelector('[data-player-activity]') || grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify([p.sent, p.received, p.pendingIn, p.sentToday, (p.recent || []).map(function (g) { return g.id + g.status + g.outcome; })]) : FG_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-friend-gifts': '', 'data-key': key, 'data-economy': 'player-friend-gifts', 'data-l10n-skip': '', 'aria-labelledby': 'player-fg-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-fg-title', class: 'font-medium flex items-center gap-2' }, icon('friend-gifts', 'h-4 w-4'), h('span', { text: t('fg.card') }));
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(FG_CARD.err ? notice('error', errorText(FG_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { FG_CARD.at = 0; loadFgCard(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var lim = p.limits || {};
  var today = lim.sender ? t('fg.card.todayOf', { n: fmtNum(p.sentToday || 0), max: fmtNum(lim.sender) }) : t('fg.card.today', { n: fmtNum(p.sentToday || 0) });
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' }, title,
    h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300 tabular-nums', 'data-fg-counts': '', text: t('fg.card.counts', { sent: fmtNum(p.sent || 0), received: fmtNum(p.received || 0) }) }),
    h('a', { href: hashFor('friend-gifts', null, { userId: id }), class: 'eco-link text-sm', 'data-fg-log-link': '', text: t('fg.card.log') })));
  body.appendChild(h('div', { class: 'flex flex-wrap gap-1 mb-2' },
    tag(today, 'secondary'),
    p.pendingIn ? tag(t('fg.card.pending', { n: fmtNum(p.pendingIn) }), 'warning') : null,
    lim.canSend === false ? tag(t('fg.rule.noSend'), 'error') : null,
    lim.ruleName ? tag(t('fg.card.rule', { name: lim.ruleName }), 'primary') : null,
    lim.senderBonus ? tag(t('fg.card.bonus', { n: fmtNum(lim.senderBonus) }), 'success') : null,
    p.enabled === false ? tag(t('fg.off'), 'warning') : null));
  var list = h('div', { class: 'flex flex-col', 'data-fg-recent': '' });
  (p.recent || []).forEach(function (g) {
    var out = g.senderId === id;
    list.appendChild(h('a', { href: hashFor('friend-gifts', g.id), class: 'flex flex-wrap items-center gap-2 py-1.5 px-1 text-sm eco-fg-row', 'data-fg-row': g.id },
      tag(t(out ? 'fg.card.out' : 'fg.card.in'), out ? 'primary' : 'success'),
      h('span', { class: 'font-medium truncate min-w-0', text: fgWhatName(g) }),
      h('span', { class: 'text-secondary-500 truncate min-w-0', text: out ? (g.recipientName || g.recipientUsername || '') : (g.senderName || g.senderUsername || '') }),
      tag(t('fg.status.' + g.status), FG_STATUS_COLOR[g.status] || 'secondary'),
      h('span', { class: 'flex-1' }),
      h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: fmtDate(g.createdAt) })));
  });
  if (!(p.recent || []).length) list.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('fg.card.none') }));
  body.appendChild(progressSection('fg.card.recent', list));
  place();
}

