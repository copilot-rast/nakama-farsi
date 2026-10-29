var CHANNELS = ['in_app', 'sms', 'email'];
var CHANNEL_ICON = { in_app: 'notifications', sms: 'comments', email: 'send' };
var NOTIF_STATUS = ['queued', 'sending', 'done', 'failed'];
var NOTIF_COLOR = { queued: 'secondary', sending: 'primary', done: 'success', failed: 'error' };
var RECIPIENT_STATUS = ['pending', 'sent', 'skipped', 'failed'];
var RECIPIENT_COLOR = { pending: 'secondary', sent: 'success', skipped: 'warning', failed: 'error' };
var SMS_MAX = 600;
var pendingComments = null;
var badgeAt = 0;

function kindLabel(kind) {
  return has('pv.kind.' + kind) ? t('pv.kind.' + kind) : kind;
}

function quietGet(path) {
  var tk = token();
  if (!tk) return Promise.reject(new ApiError(401, 'UNAUTHENTICATED', ''));
  return fetch(PUB_BASE + path, { headers: { Accept: 'application/json', Authorization: 'Bearer ' + tk }, credentials: 'same-origin' }).then(function (res) {
    if (!res.ok) throw new ApiError(res.status, null, '');
    return res.json();
  });
}

function refreshBadge(force) {
  var now = Date.now();
  if (!force && now - badgeAt < 60000) return;
  badgeAt = now;
  quietGet('/comments?status=pending&limit=1').then(function (res) {
    var n = (res && res.total) || 0;
    if (n !== pendingComments) {
      pendingComments = n;
      schedule();
    }
  }, function () { return null; });
}

function navBadge(page, collapsed) {
  if (page !== 'comments' || !pendingComments) return null;
  if (collapsed) return h('span', { class: 'eco-nav-dot', 'aria-label': t('cm.pendingN', { n: pendingComments }) });
  return h('span', { class: 'eco-nav-badge tabular-nums', 'aria-label': t('cm.pendingN', { n: pendingComments }), text: fmtNum(pendingComments) });
}

function userCell(id, username, displayName) {
  return h('div', { class: 'min-w-0' },
    h('a', { href: '#/players/' + encodeURIComponent(id || ''), class: 'eco-link font-medium truncate block', text: displayName || username || String(id || '').slice(0, 8) }),
    h('div', { class: 'text-xs text-secondary-500 truncate' }, mono(username && displayName ? username : String(id || '').slice(0, 8))));
}

function commentAct(btn, id, accepted) {
  setBusy(btn, true);
  pubApi('PUT', '/comments/' + encodeURIComponent(id), { isAccepted: accepted }).then(function () {
    toast('success', t(accepted ? 'cm.accepted.done' : 'cm.rejected.done'));
    refreshBadge(true);
    if (rendered && rendered.list) rendered.list.reload();
  }, function (err) {
    setBusy(btn, false);
    toast('error', errorText(err));
  });
}

function commentActions(it) {
  var box = h('div', { class: 'flex flex-wrap items-center gap-1 justify-end eco-row-actions' });
  function reset() {
    clear(box);
    if (!it.isAccepted) box.appendChild(button({ variant: 'primary', size: 'sm', icon: 'check', label: t('cm.accept'), onclick: function (e) { commentAct(e.currentTarget, it.id, true); } }));
    else box.appendChild(button({ variant: 'ghost', size: 'sm', icon: 'close', label: t('cm.reject'), title: t('cm.reject.hint'), onclick: function (e) { commentAct(e.currentTarget, it.id, false); } }));
    box.appendChild(button({ variant: 'danger', size: 'sm', icon: 'trash', title: t('delete'), onclick: ask }));
  }
  function ask() {
    clear(box);
    box.appendChild(h('span', { class: 'text-sm text-secondary-700 dark:text-secondary-200', text: t('delete.confirm') }));
    box.appendChild(button({ variant: 'dangerSolid', size: 'sm', label: t('delete.yes'), onclick: function (e) {
      setBusy(e.currentTarget, true);
      pubApi('DELETE', '/comments/' + encodeURIComponent(it.id)).then(function () {
        toast('success', t('deleted'));
        refreshBadge(true);
        if (rendered && rendered.list) rendered.list.reload();
      }, function (err) { reset(); toast('error', errorText(err)); });
    } }));
    box.appendChild(button({ variant: 'ghost', size: 'sm', label: t('no'), onclick: reset }));
  }
  reset();
  return box;
}

var COMMENT_LIST = {
  noRow: true,
  filters: [
    { name: 'status', label: 'filter.status', options: opts('cm.', ['pending', 'accepted']) },
    { name: 'newsId', label: 'filter.news', type: 'ref', resource: 'news', api: 'publication' }
  ],
  columns: [
    { label: 'col.comment', render: function (it) {
      return h('div', { class: 'min-w-0 flex flex-col gap-1' },
        h('div', { class: 'flex items-start gap-1.5' },
          it.parentId ? h('span', { class: 'text-secondary-400 mt-0.5', title: t('cm.reply') }, icon('reply', 'h-3.5 w-3.5')) : null,
          h('div', { class: 'eco-clamp-3 eco-pre-line break-words', dir: 'auto', text: it.message })),
        h('div', { class: 'flex flex-wrap items-center eco-meta text-xs text-secondary-500' },
          h('a', { href: '#/players/' + encodeURIComponent(it.userId), class: 'eco-link font-medium', text: it.displayName || it.username || String(it.userId).slice(0, 8) }),
          h('span', { text: fmtDate(it.createdAt) }),
          it.parentId ? h('span', { text: t('cm.reply') }) : null,
          it.replyCount ? h('span', { text: t('cm.replies', { n: it.replyCount }) }) : null));
    }, skeleton: '75%' },
    { label: 'col.news', render: function (it) { return h('a', { href: hashFor('news', it.newsId), class: 'eco-link eco-clamp', text: i18nText(it.newsTitle) }); }, width: 220, skeleton: '60%' },
    { label: 'col.status', render: function (it) { return tag(t(it.isAccepted ? 'cm.accepted' : 'cm.pending'), it.isAccepted ? 'success' : 'warning'); }, width: 120 },
    { label: 'col.actions', render: commentActions, width: 210, skeleton: '80%' }
  ]
};

function targetText(tg) {
  tg = tg || {};
  if (tg.mode === 'all') return t('tg.all');
  if (tg.mode === 'group') return t('tg.group');
  return t('tg.usersN', { n: (tg.userIds || []).length });
}

function notifTitle(it) {
  return i18nText(it.title) || mdPlain(i18nText(it.body));
}

var NOTIFICATION_LIST = {
  create: function () { return button({ variant: 'primary', size: 'page', icon: 'send', label: t('nt.new'), onclick: function () { go('notifications', 'new'); } }); },
  filters: [
    { name: 'channel', label: 'filter.channel', options: opts('ch.', CHANNELS) },
    { name: 'status', label: 'filter.status', options: opts('nst.', NOTIF_STATUS) }
  ],
  columns: [
    { label: 'col.notification', render: function (it) {
      return h('div', { class: 'min-w-0' },
        h('div', { class: 'font-medium eco-clamp', dir: 'auto', text: notifTitle(it) }),
        h('div', { class: 'text-xs text-secondary-500 flex flex-wrap eco-meta' }, h('span', { text: fmtDate(it.createdAt) }), it.createdBy ? h('span', { text: it.createdBy }) : null));
    }, skeleton: '65%' },
    { label: 'col.channel', render: function (it) {
      return h('div', { class: 'flex flex-wrap items-center gap-1' }, tag(t('ch.' + it.channel), it.channel === 'in_app' ? 'primary' : 'secondary'), it.providerKind ? h('span', { class: 'text-xs text-secondary-500', text: kindLabel(it.providerKind) }) : null);
    }, width: 170 },
    { label: 'col.target', render: function (it) { return h('span', { class: 'whitespace-nowrap', text: targetText(it.target) }); }, width: 120 },
    { label: 'col.delivery', render: function (it) {
      return h('div', { class: 'flex flex-wrap items-center gap-1' },
        h('span', { class: 'tabular-nums whitespace-nowrap text-sm', text: t('nt.sentOf', { sent: it.sentCount || 0, total: it.recipientCount || 0 }) }),
        it.skippedCount ? tag(t('nt.skippedN', { n: it.skippedCount }), 'warning') : null,
        it.failedCount ? tag(t('nt.failedN', { n: it.failedCount }), 'error') : null);
    }, width: 210 },
    { label: 'col.status', render: function (it) { return tag(t('nst.' + it.status), NOTIF_COLOR[it.status]); }, width: 120 }
  ]
};

var LISTS = { payments: PAYMENT_LIST, comments: COMMENT_LIST, notifications: NOTIFICATION_LIST };

function statTile(label, value, color) {
  return h('div', { class: 'rounded-md border border-secondary-200 dark:border-secondary-700 px-3 py-2' },
    h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: label }),
    h('div', { class: 'text-2xl font-semibold tabular-nums' + (color ? ' eco-text-' + color : ''), text: fmtNum(value || 0) }));
}

function i18nBlock(v, markdown) {
  v = v || {};
  var out = h('div', { class: 'flex flex-col gap-3' });
  var order = LANGS.items.map(function (l) { return l.code; });
  Object.keys(v).sort().forEach(function (k) { if (order.indexOf(k) === -1) order.push(k); });
  order.forEach(function (l) {
    if (!v[l]) return;
    var ltr = langDir(l) === 'ltr';
    out.appendChild(h('div', { 'data-lang': l },
      h('div', { class: 'text-xs text-secondary-500 mb-1', text: langName(l) + (l === LANGS.def ? ' (' + t('tr.default') + ')' : '') }),
      markdown ? mdPreview(v[l], ltr) : h('div', { class: 'text-sm eco-pre-line break-words', dir: ltr ? 'ltr' : 'rtl', text: v[l] })));
  });
  return out;
}

function notificationDetail(container, id) {
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('notifications', null, lastQuery.notifications || {}); } });
  var frame = pageFrame(t('nt.title'), null, backBtn);
  var titleBox = frame.el.querySelector('h1');
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-detail-grid' });
  frame.content.appendChild(grid);
  var rState = { page: 1, limit: 20, status: '', total: 0, loading: true, loaded: false };
  var recBox = h('div', {});
  var recPagerTop = h('div', { class: 'flex items-center justify-end' });
  var recPagerBottom = h('div', { class: 'flex items-center justify-end eco-pt-3' });
  var timer = null;
  var built = false;
  var statsBox = h('div', { class: 'eco-tile-grid' });
  var statusBox = h('div', {});

  function alive() { return container.isConnected && rendered && rendered.root === container; }

  function skeleton() {
    clear(grid);
    [5, 4].forEach(function (n) {
      var rows = [];
      for (var i = 0; i < n; i++) rows.push([h('dt', {}, bar('6rem')), h('dd', {}, bar(['70%', '50%', '40%'][i % 3]))]);
      grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), h('dl', { class: 'eco-dl' }, rows))));
    });
    var tiles = [];
    for (var k = 0; k < 6; k++) tiles.push(h('div', { class: 'rounded-md border border-secondary-200 dark:border-secondary-700 px-3 py-2' }, bar('4rem'), bar('2.5rem', 'h-7 mt-2')));
    grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), h('div', { class: 'eco-tile-grid' }, tiles)), 'eco-span-all'));
    grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), table(recColumns(), skeletonRows(recColumns(), 5))), 'eco-span-all'));
  }

  function recColumns() {
    return [
      { label: t('col.user'), skeleton: '50%' },
      { label: t('col.status'), width: 120 },
      { label: t('col.address'), width: 200, skeleton: '70%' },
      { label: t('col.error'), skeleton: '40%' },
      { label: t('col.attempts'), width: 80 },
      { label: t('col.time'), width: 180 }
    ];
  }

  function paintStats(n) {
    clear(statusBox).appendChild(h('div', { class: 'flex flex-wrap items-center gap-1.5' },
      tag(t('nst.' + n.status), NOTIF_COLOR[n.status]),
      n.status === 'queued' || n.status === 'sending' ? h('span', { class: 'text-xs text-secondary-500', text: t('nt.live') }) : null));
    clear(statsBox);
    statsBox.appendChild(statTile(t('nt.recipients'), n.recipientCount));
    statsBox.appendChild(statTile(t('rst.sent'), n.sentCount, 'success'));
    statsBox.appendChild(statTile(t('rst.skipped'), n.skippedCount, n.skippedCount ? 'warning' : null));
    statsBox.appendChild(statTile(t('rst.failed'), n.failedCount, n.failedCount ? 'error' : null));
    statsBox.appendChild(statTile(t('rst.pending'), n.pendingCount));
    if (n.channel === 'in_app') statsBox.appendChild(statTile(t('nt.seen'), n.seenCount));
  }

  function build(n) {
    built = true;
    clear(grid);
    clear(titleBox).appendChild(h('span', { text: t('nt.title') + ' ' }));
    titleBox.appendChild(h('span', { class: 'font-mono text-base text-secondary-500', dir: 'ltr', text: String(n.id).slice(0, 8) }));
    var action = n.action && n.action.type ? h('div', { class: 'flex flex-wrap items-center gap-1.5' }, tag(t('act.' + n.action.type), 'secondary'), mono(n.action.value), n.action.label && i18nText(n.action.label) ? h('span', { class: 'text-sm', text: i18nText(n.action.label) }) : null) : null;
    grid.appendChild(section('nt.info', dl([
      ['col.status', statusBox],
      ['col.channel', h('div', { class: 'flex flex-wrap items-center gap-1.5' }, tag(t('ch.' + n.channel), n.channel === 'in_app' ? 'primary' : 'secondary'), n.providerKind ? h('span', { class: 'text-sm', text: kindLabel(n.providerKind) }) : null)],
      ['col.target', h('span', { text: targetText(n.target) })],
      ['nt.action', action],
      ['nt.createdBy', n.createdBy || null],
      ['pay.createdAt', dateCell(n.createdAt)],
      ['pay.updatedAt', dateCell(n.updatedAt)]
    ])));
    var content = h('div', { class: 'flex flex-col gap-4' });
    var hasTitle = !!i18nText(n.title);
    if (hasTitle) content.appendChild(h('div', {}, h('div', { class: 'text-sm font-medium mb-1', text: t('nt.f.title') }), i18nBlock(n.title)));
    content.appendChild(h('div', {}, hasTitle ? h('div', { class: 'text-sm font-medium mb-1', text: t('nt.f.body') }) : null, i18nBlock(n.body, n.channel !== 'sms')));
    grid.appendChild(section('nt.content', content));
    var stats = section('nt.delivery', statsBox);
    stats.classList.add('eco-span-all');
    grid.appendChild(stats);
    var filter = filterSelect(t('filter.status'), rState.status, opts('rst.', RECIPIENT_STATUS)(), function (v) { rState.status = v; rState.page = 1; loadRecipients(); });
    var recs = section('nt.recipientsList', h('div', {}, recBox, recPagerBottom), h('div', { class: 'flex flex-wrap items-center gap-2 justify-end' }, filter, recPagerTop));
    recs.classList.add('eco-span-all');
    grid.appendChild(recs);
  }

  function paintRecPager() {
    clear(recPagerTop).appendChild(pager(rState, function (p) { rState.page = p; loadRecipients(); }));
    clear(recPagerBottom).appendChild(pager(rState, function (p) { rState.page = p; loadRecipients(); }));
  }

  function renderRecipients(r) {
    var items = (r && r.items) || [];
    rState.total = (r && r.total) || 0;
    rState.loaded = true;
    rState.loading = false;
    recBox.classList.remove('opacity-60');
    clear(recBox);
    if (!items.length) {
      recBox.appendChild(h('div', { class: 'text-sm text-secondary-500 py-6 text-center', text: rState.status ? t('empty.search') : t('nt.noRecipients') }));
    } else {
      var cols = recColumns();
      recBox.appendChild(table(cols, items.map(function (it) {
        var sandbox = it.error === 'sandbox';
        return { item: it, cells: [
          userCell(it.userId, it.username, it.displayName),
          h('div', { class: 'flex flex-wrap gap-1' }, tag(t('rst.' + it.status), RECIPIENT_COLOR[it.status]), sandbox ? tag(t('pv.sandbox.short'), 'warning') : null, it.lang ? h('span', { 'data-recipient-lang': it.lang }, tag(langName(it.lang), 'secondary')) : null),
          it.address ? mono(it.address) : h('span', {}),
          it.error && !sandbox && it.error !== 'no address' ? h('span', { class: 'text-xs text-error-600 dark:text-error-400 break-words', dir: 'auto', text: it.error }) : h('span', {}),
          h('span', { class: 'tabular-nums', text: fmtNum(it.attempts || 0) }),
          h('div', { class: 'flex flex-col text-xs text-secondary-600 dark:text-secondary-300' },
            it.sentAt ? h('span', { class: 'whitespace-nowrap', text: t('nt.sentAt', { t: fmtDate(it.sentAt) }) }) : null,
            it.seenAt ? h('span', { class: 'whitespace-nowrap', text: t('nt.seenAt', { t: fmtDate(it.seenAt) }) }) : null)
        ] };
      })));
    }
    paintRecPager();
  }

  function fetchDetail() {
    return pubApi('GET', '/notifications/' + encodeURIComponent(id) + qs({ rPage: rState.page, rLimit: rState.limit, rStatus: rState.status }));
  }

  function loadRecipients() {
    rState.loading = true;
    recBox.classList.add('opacity-60');
    paintRecPager();
    fetchDetail().then(function (n) {
      if (!alive()) return;
      paintStats(n);
      renderRecipients(n.recipients);
    }, function (err) { recBox.classList.remove('opacity-60'); toast('error', errorText(err)); });
  }

  function poll() {
    clearTimeout(timer);
    timer = setTimeout(function () {
      if (!alive()) return;
      fetchDetail().then(function (n) {
        if (!alive()) return;
        paintStats(n);
        renderRecipients(n.recipients);
        if (n.status === 'queued' || n.status === 'sending') poll();
      }, function () { poll(); });
    }, 2500);
  }

  function load() {
    skeleton();
    fetchDetail().then(function (n) {
      if (!alive()) return;
      build(n);
      paintStats(n);
      renderRecipients(n.recipients);
      if (n.status === 'queued' || n.status === 'sending') poll();
    }, function (err) { clear(grid).appendChild(h('div', { class: 'eco-span-all' }, errorBlock(err, load))); });
  }
  load();
}

function groupsApi(method, path) {
  var m = /^\/groups\/([^?]+)$/.exec(path);
  if (m) return consoleApi('GET', '/group/' + m[1]).then(function (g) { return g && g.id ? { id: g.id, name: g.name, key: null } : null; });
  var q = new URLSearchParams(path.split('?')[1] || '').get('q') || '';
  return consoleApi('GET', '/group' + qs({ filter: q })).then(function (res) {
    return { items: ((res && res.groups) || []).map(function (g) { return { id: g.id, name: g.name }; }) };
  });
}

function usersApi(method, path) {
  var m = /^\/users\/([^?]+)$/.exec(path);
  var q = m ? decodeURIComponent(m[1]) : new URLSearchParams(path.split('?')[1] || '').get('q') || '';
  if (!q) return Promise.resolve(m ? null : { items: [] });
  return pubApi('GET', '/users/search' + qs({ q: q })).then(function (list) {
    var items = (list || []).map(function (u) { return { id: u.id, name: u.displayName || u.username, key: u.username }; });
    return m ? items.filter(function (u) { return u.id === q; })[0] || null : { items: items };
  });
}

function composePage(container, query) {
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('notifications', null, lastQuery.notifications || {}); } });
  var frame = pageFrame(t('nt.new'), t('nt.new.desc'), backBtn);
  clear(container).appendChild(frame.el);
  var layout = h('div', { class: 'eco-compose' });
  frame.content.appendChild(layout);
  var v = {
    channel: CHANNELS.indexOf(query.channel) !== -1 ? query.channel : 'in_app',
    providerId: '',
    title: {},
    body: {},
    action: { type: '', value: '', label: {} },
    target: { mode: 'users', userIds: query.userId ? [query.userId] : [], groupId: '' }
  };
  var providers = [];
  var preview = null;
  var previewSeq = 0;

  function skeleton() {
    clear(layout);
    var main = h('div', { class: 'flex flex-col gap-4 min-w-0' });
    [2, 3, 1, 2].forEach(function (n) {
      var f = [];
      for (var i = 0; i < n; i++) f.push(skeletonField(i === 1 ? 'h-24' : 'h-10'));
      main.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), f)));
    });
    layout.appendChild(main);
    layout.appendChild(h('div', {}, card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem', 'h-4'), bar('4rem', 'h-9'), bar('80%'), bar('100%', 'h-9')))));
  }

  function build() {
    clear(layout);
    var main = h('div', { class: 'flex flex-col gap-4 min-w-0' });
    var aside = h('div', { class: 'eco-compose-aside' });
    layout.appendChild(main);
    layout.appendChild(aside);
    var errTop = h('div', {});
    var fields = {};

    var providerBox = h('div', {});
    var channelSeg = segmented(CHANNELS.map(function (c) { return { value: c, label: t('ch.' + c), icon: CHANNEL_ICON[c] }; }), v.channel, function (c) {
      v.channel = c;
      v.providerId = '';
      paintChannel();
      paintCount();
    }, t('col.channel'), 'md');
    main.appendChild(section('nt.s.channel', h('div', {}, h('div', { class: 'mb-4' }, channelSeg), providerBox)));

    var titleRow = h('div', {});
    var titleCtl = translatedInput(v.title, { label: t('nt.f.title'), oninput: function (x) { v.title = x; fields.title.setError(''); titleCtl.markLang(null); } });
    fields.title = field(t('nt.f.title'), titleCtl);
    fields.title.ctl = titleCtl;
    titleRow.appendChild(fields.title);
    var bodyBox = h('div', {});
    main.appendChild(section('nt.s.content', h('div', {}, titleRow, bodyBox)));

    var actType = selectInput('', [{ value: 'url', label: t('act.url') }, { value: 'goto', label: t('act.goto') }], function (x) {
      v.action.type = x;
      actValue.hidden = !x;
      actValue.placeholder = x === 'url' ? 'https://' : 'shop';
      if (x === 'goto') actValue.setAttribute('list', 'eco-nt-goto'); else actValue.removeAttribute('list');
      fields.action.setError('');
    }, { label: t('act.type'), placeholder: t('act.none') });
    var actValue = textInput('', { ltr: true, mono: true, label: t('act.value'), oninput: function (x) { v.action.value = x.trim(); fields.action.setError(''); } });
    actValue.hidden = true;
    fields.action = field(null, h('div', { class: 'eco-form-row' }, actType, actValue, h('datalist', { id: 'eco-nt-goto' }, GOTO_SCREENS.map(function (s) { return h('option', { value: s }); }))), t('nt.action.hint'));
    var actLabelCtl = translatedInput(v.action.label, { label: t('act.label'), oninput: function (x) { v.action.label = x; fields.actionLabel.setError(''); actLabelCtl.markLang(null); } });
    fields.actionLabel = field(t('act.label'), actLabelCtl, t('nt.action.label.hint'), { optional: true });
    fields.actionLabel.ctl = actLabelCtl;
    fields.actionLabel.hidden = true;
    var actType0 = actType;
    actType0.addEventListener('change', function () { fields.actionLabel.hidden = !actType0.value; });
    var actionSection = section('nt.s.action', h('div', {}, fields.action, fields.actionLabel));
    main.appendChild(actionSection);

    var targetBox = h('div', {});
    var modeSeg = segmented([{ value: 'users', label: t('tg.users'), icon: 'users' }, { value: 'group', label: t('tg.group'), icon: 'categories' }, { value: 'all', label: t('tg.all'), icon: 'notifications' }], v.target.mode, function (m) {
      v.target.mode = m;
      paintTarget();
      refreshPreview();
    }, t('col.target'), 'md');
    main.appendChild(section('nt.s.target', h('div', {}, h('div', { class: 'mb-4' }, modeSeg), targetBox)));

    var countBox = h('div', { class: 'flex flex-col gap-1', 'aria-live': 'polite' });
    var sendBtn = button({ variant: 'primary', size: 'footer', icon: 'send', label: t('nt.send'), class: 'w-full', onclick: function () { send(); } });
    aside.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' },
      h('h2', { class: 'text-lg font-medium', text: t('nt.summary') }),
      countBox, errTop, sendBtn)));

    function paintChannel() {
      clear(providerBox);
      titleRow.hidden = v.channel === 'sms';
      actionSection.hidden = v.channel === 'sms';
      clear(bodyBox);
      var max = v.channel === 'sms' ? SMS_MAX : v.channel === 'email' ? 20000 : 2000;
      var bodyCtl = translatedInput(v.body, { label: t(v.channel === 'sms' ? 'nt.f.sms' : 'nt.f.body'), markdown: v.channel !== 'sms', multiline: true, rows: 4, max: max, counter: v.channel === 'sms', oninput: function (x) { v.body = x; fields.body.setError(''); bodyCtl.markLang(null); } });
      fields.body = field(t(v.channel === 'sms' ? 'nt.f.sms' : 'nt.f.body'), bodyCtl, v.channel === 'sms' ? t('nt.sms.hint') : null);
      fields.body.ctl = bodyCtl;
      bodyBox.appendChild(fields.body);
      if (v.channel === 'in_app') {
        providerBox.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: t('nt.inapp.hint') }));
        return;
      }
      var list = providers.filter(function (p) { return p.channel === v.channel && p.isActive; });
      if (!list.length) {
        providerBox.appendChild(notice('error', t('nt.noProvider'), h('a', { href: hashFor(v.channel), class: 'eco-link text-sm font-medium', text: t('nt.goSettings.' + v.channel) })));
        return;
      }
      var def = list.filter(function (p) { return p.isDefault; })[0] || list[0];
      var options = list.map(function (p) { return { value: p.id, label: kindLabel(p.kind) + (p.id === def.id ? ' (' + t('pv.default') + ')' : '') + (p.sandbox ? ' · ' + t('pv.sandbox.short') : '') }; });
      if (!v.providerId) v.providerId = def.id;
      var sandboxNote = h('div', {});
      var paintNote = function () {
        var p = list.filter(function (x) { return x.id === v.providerId; })[0];
        clear(sandboxNote);
        if (!p) return;
        if (p.sandbox) sandboxNote.appendChild(h('div', { class: 'eco-sandbox-note rounded-md px-3 py-2 text-sm flex items-center gap-2' }, icon('flask', 'h-4 w-4 shrink-0'), h('span', { text: t('nt.sandbox.note') })));
        else sandboxNote.appendChild(notice('error', t('nt.live.note')));
      };
      fields.provider = field(t('nt.f.provider'), selectInput(v.providerId, options, function (x) { v.providerId = x; paintNote(); }, { placeholder: false, label: t('nt.f.provider') }));
      providerBox.appendChild(fields.provider);
      providerBox.appendChild(sandboxNote);
      paintNote();
    }

    function paintTarget() {
      clear(targetBox);
      fields.target = null;
      if (v.target.mode === 'users') {
        fields.target = field(t('tg.pickUsers'), refMulti({ resource: 'users', api: usersApi, value: v.target.userIds, label: t('tg.pickUsers'), placeholder: t('tg.searchUsers'), emptyText: t('tg.searchUsers.hint'), onchange: function (ids) {
          v.target.userIds = ids;
          fields.target.setError('');
          refreshPreview();
        } }), t('tg.users.hint'));
      } else if (v.target.mode === 'group') {
        fields.target = field(t('tg.pickGroup'), refPicker({ resource: 'groups', api: groupsApi, value: v.target.groupId, required: false, label: t('tg.pickGroup'), placeholder: t('tg.pickGroup'), onchange: function (id) {
          v.target.groupId = id || '';
          fields.target.setError('');
          refreshPreview();
        } }), t('tg.group.hint'));
      } else {
        targetBox.appendChild(notice('secondary', t('tg.all.hint')));
      }
      if (fields.target) targetBox.appendChild(fields.target);
    }

    function paintCount() {
      clear(countBox);
      if (preview === 'loading') {
        countBox.appendChild(bar('4rem', 'h-9'));
        countBox.appendChild(bar('80%'));
        return;
      }
      if (!preview || preview.error) {
        countBox.appendChild(h('div', { class: 'text-sm text-secondary-500', text: preview && preview.error ? errorText(preview.error) : t('tg.none') }));
        return;
      }
      countBox.appendChild(h('div', { class: 'flex items-end gap-2' },
        h('span', { class: 'text-3xl font-semibold tabular-nums', 'data-recipient-count': String(preview.count), text: fmtNum(preview.count) }),
        h('span', { class: 'text-sm text-secondary-500', text: t('nt.recipientsWord') })));
      var reach = v.channel === 'sms' ? preview.withPhone : v.channel === 'email' ? preview.withEmail : preview.count;
      countBox.appendChild(h('div', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: t('tg.withPhone', { n: preview.withPhone }) + '، ' + t('tg.withEmail', { n: preview.withEmail }) }));
      if (v.channel !== 'in_app' && reach < preview.count) {
        countBox.appendChild(h('div', { class: 'text-sm eco-text-warning', text: t('tg.skipped', { n: preview.count - reach }) }));
      }
    }

    var runPreview = debounce(function () {
      var my = ++previewSeq;
      var tg = v.target;
      if ((tg.mode === 'users' && !tg.userIds.length) || (tg.mode === 'group' && !tg.groupId)) { preview = null; paintCount(); return; }
      pubApi('GET', '/notifications/preview-target' + qs({ mode: tg.mode, groupId: tg.mode === 'group' ? tg.groupId : '', userIds: tg.mode === 'users' ? tg.userIds.join(',') : '' })).then(function (res) {
        if (my !== previewSeq) return;
        preview = res;
        paintCount();
      }, function (err) {
        if (my !== previewSeq) return;
        preview = { error: err };
        paintCount();
      });
    }, 250);

    function refreshPreview() {
      preview = 'loading';
      paintCount();
      runPreview();
    }

    function validate() {
      var ok = true;
      var setErr = function (f, msg) { if (f) f.setError(msg); if (msg) ok = false; };
      [fields.title, fields.body, fields.action, fields.actionLabel, fields.target].forEach(function (f) { if (f) f.setError(''); });
      var max = v.channel === 'sms' ? SMS_MAX : v.channel === 'email' ? 20000 : 2000;
      var trErr = function (f, val, o) {
        var hit = trCheck(val, o);
        f.ctl.markLang(hit ? hit.lang : null);
        if (hit) setErr(f, hit.code === 'TEXT_TOO_LONG' && v.channel === 'sms' ? t('nt.sms.tooLong') : t('code.' + hit.code, { lang: langName(hit.lang) }));
      };
      if (v.channel !== 'sms') trErr(fields.title, v.title, { required: true, max: 200 });
      trErr(fields.body, v.body, { required: true, max: max });
      if (v.channel !== 'sms' && v.action.type) trErr(fields.actionLabel, v.action.label, { max: 64 });
      if (v.channel !== 'sms' && v.action.type) {
        if (!v.action.value) setErr(fields.action, t('act.value') + ': ' + t('err.required'));
        else if (v.action.type === 'url' && !/^https?:\/\/\S+$/i.test(v.action.value)) setErr(fields.action, t('code.INVALID_ACTION'));
      }
      if (v.target.mode === 'users' && !v.target.userIds.length) setErr(fields.target, t('code.TARGET_USERS_REQUIRED'));
      if (v.target.mode === 'group' && !v.target.groupId) setErr(fields.target, t('err.required'));
      if (v.channel !== 'in_app' && !providers.some(function (p) { return p.channel === v.channel && p.isActive; })) ok = false;
      return ok;
    }

    function payload() {
      var p = { channel: v.channel, body: trClean(v.body), target: { mode: v.target.mode } };
      if (v.channel !== 'sms') {
        p.title = trClean(v.title);
        if (v.action.type) {
          p.action = { type: v.action.type, value: v.action.value };
          var lbl = trClean(v.action.label);
          if (Object.keys(lbl).length) p.action.label = lbl;
        }
      }
      if (v.target.mode === 'users') p.target.userIds = v.target.userIds.slice();
      if (v.target.mode === 'group') p.target.groupId = v.target.groupId;
      if (v.channel !== 'in_app' && v.providerId) p.providerId = v.providerId;
      return p;
    }

    function send() {
      clear(errTop);
      if (!validate()) {
        var bad = layout.querySelector('.eco-invalid');
        if (bad) bad.scrollIntoView({ block: 'center', behavior: 'smooth' });
        return;
      }
      var prov = providers.filter(function (p) { return p.id === v.providerId; })[0];
      var live = v.channel !== 'in_app' && prov && !prov.sandbox;
      var count = preview && !preview.error && preview !== 'loading' ? preview.count : null;
      var ask = Promise.resolve(true);
      if (v.target.mode === 'all' || live) {
        ask = confirmModal({
          title: v.target.mode === 'all' ? t('nt.confirmAll.title') : t('nt.confirmLive.title'),
          text: (v.target.mode === 'all' ? (count !== null ? t('nt.confirmAll.text', { n: count }) : t('nt.confirmAll.textNoCount')) : '') + (live ? (v.target.mode === 'all' ? ' ' : '') + t('nt.confirmLive.text') : ''),
          ok: v.target.mode === 'all' ? t('nt.confirmAll.ok') : t('nt.send'),
          danger: live,
          icon: 'send'
        });
      }
      ask.then(function (yes) {
        if (!yes) return;
        setBusy(sendBtn, true, t('nt.sending'));
        pubApi('POST', '/notifications', payload()).then(function (res) {
          toast('success', t('nt.queued'));
          go('notifications', res.id);
        }, function (err) {
          setBusy(sendBtn, false);
          var map = { title: fields.title, body: fields.body, action: fields.action, 'action.label': fields.actionLabel, 'target.userIds': fields.target, 'target.groupId': fields.target, providerId: fields.provider };
          var f = err && err.field && map[err.field];
          var vals = { title: v.title, body: v.body, 'action.label': v.action.label };
          var code = f && f.ctl ? trServerLang(err, vals[err.field], { max: err.field === 'title' ? 200 : err.field === 'action.label' ? 64 : v.channel === 'sms' ? SMS_MAX : v.channel === 'email' ? 20000 : 2000 }) : null;
          if (f && f.ctl) f.ctl.markLang(code);
          if (f) { f.hidden = false; f.setError(errorText(err, code)); f.scrollIntoView({ block: 'center', behavior: 'smooth' }); }
          else errTop.appendChild(notice('error', errorText(err)));
        });
      });
    }

    paintChannel();
    paintTarget();
    paintCount();
    if (v.target.userIds.length) refreshPreview();
  }

  function load() {
    skeleton();
    Promise.all([pubApi('GET', '/providers'), freshLanguages()]).then(function (all) {
      var res = all[0];
      if (!container.isConnected) return;
      providers = (res && res.items) || [];
      build();
    }, function (err) { clear(layout).appendChild(h('div', { class: 'eco-span-all' }, errorBlock(err, load))); });
  }
  load();
}

function providerField(f, cfg, onchange) {
  var label = has('pv.field.' + f.name) ? t('pv.field.' + f.name) : f.name;
  var ctl;
  if (f.type === 'select') {
    ctl = selectInput(cfg[f.name] || f.default || '', (f.options || []).map(function (o) { return { value: o, label: has('pv.opt.' + o) ? t('pv.opt.' + o) : o }; }), function (x) { cfg[f.name] = x; onchange(); }, { placeholder: false, label: label });
  } else {
    ctl = textInput(cfg[f.name] == null ? '' : toLatinDigits(String(cfg[f.name])), { ltr: true, mono: true, label: label, inputmode: f.type === 'number' ? 'numeric' : f.type === 'email' ? 'email' : null, placeholder: f.default ? toLatinDigits(f.default) : null, oninput: function (x) { cfg[f.name] = f.type === 'number' ? toLatinDigits(x).trim() : x; onchange(); } });
    ctl.classList.add('text-sm');
    if (f.type === 'secret') ctl.classList.add('eco-secret');
  }
  var hintKey = f.type === 'secret' && cfg[f.name] ? 'pm.secret.hint' : has('pv.hint.' + f.name) ? 'pv.hint.' + f.name : null;
  return field(label, ctl, hintKey ? t(hintKey) : null, { optional: !f.required });
}

function providersPage(container, channel) {
  var frame = tabFrame(t('desc.' + channel));
  clear(container).appendChild(frame.el);
  var body = h('section', { 'data-channel': channel });
  frame.content.appendChild(body);
  var catalog = [];

  function skeleton() {
    clear(frame.actions).appendChild(bar('12rem', 'h-8'));
    var grid = h('div', { class: 'eco-card-grid' });
    for (var i = 0; i < (channel === 'sms' ? 3 : 2); i++) {
      grid.appendChild(card(h('div', { class: 'p-4' },
        h('div', { class: 'flex items-center justify-between mb-4' }, bar('7rem', 'h-4'), bar('2rem', 'h-5 rounded-full')),
        bar('100%', 'h-14 mb-4'),
        skeletonField(), skeletonField(),
        h('div', { class: 'flex justify-end' }, bar('5rem', 'h-8')))));
    }
    clear(body).appendChild(grid);
  }

  function testBox(p) {
    var email = p.channel === 'email';
    var to = textInput('', { ltr: true, mono: true, label: t('pv.test.to'), placeholder: email ? 'name@example.com' : '09xxxxxxxxx', inputmode: email ? 'email' : 'tel' });
    to.classList.add('text-sm');
    var subject = email ? textInput('', { label: t('pv.test.subject'), placeholder: t('pv.test.subject.ph') }) : null;
    if (subject) subject.classList.add('text-sm');
    var msg = textInput(t('pv.test.message.default'), { multiline: true, rows: 2, label: t('pv.test.message') });
    msg.classList.add('text-sm');
    var out = h('div', { 'aria-live': 'polite' });
    var sendBtn = button({ variant: 'link', size: 'sm', icon: 'send', label: t('pv.test.send'), onclick: function () {
      clear(out);
      setBusy(sendBtn, true, t('nt.sending'));
      var payload = { to: toLatinDigits(to.value).trim(), message: msg.value };
      if (subject) payload.subject = subject.value;
      pubApi('POST', '/providers/' + encodeURIComponent(p.id) + '/test', payload).then(function (r) {
        setBusy(sendBtn, false);
        if (r && r.ok && r.sandbox) out.appendChild(h('div', { class: 'eco-sandbox-note rounded-md px-3 py-2 text-sm flex items-start gap-2', 'data-test-result': 'sandbox' }, icon('flask', 'h-4 w-4 shrink-0 mt-0.5'), h('span', {}, h('span', { text: t('pv.test.sandbox') + ' ' }), r.reply ? mono(r.reply) : null)));
        else if (r && r.ok) out.appendChild(h('div', { 'data-test-result': 'ok' }, notice('success', t('pv.test.ok') + (r.reply ? ' ' + r.reply : ''))));
        else out.appendChild(h('div', { 'data-test-result': 'error' }, notice('error', t('pv.test.failed') + ' ' + ((r && r.error) || ''))));
      }, function (err) {
        setBusy(sendBtn, false);
        out.appendChild(notice('error', errorText(err)));
      });
    } });
    var details = h('details', { class: 'eco-test border-t border-secondary-200 dark:border-secondary-700 mt-4 eco-pt-3' },
      h('summary', { class: 'cursor-pointer text-sm font-medium text-secondary-700 dark:text-secondary-200 select-none', text: t('pv.test') }),
      h('div', { class: 'mt-3' },
        h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: t('pv.test.hint') }),
        field(t('pv.test.to'), to),
        subject ? field(t('pv.test.subject'), subject, null, { optional: true }) : null,
        field(t('pv.test.message'), msg),
        h('div', { class: 'flex justify-end -mt-2 mb-2' }, sendBtn),
        out));
    return details;
  }

  function providerCard(p) {
    var cfg = clone(p.config || {}) || {};
    var st = { isActive: !!p.isActive, isDefault: !!p.isDefault, sandbox: !!p.sandbox };
    var errTop = h('div', {});
    var fieldEls = {};
    var fieldsBox = h('div', {});
    (p.fields || []).forEach(function (f) {
      var w = providerField(f, cfg, function () { if (fieldEls[f.name]) fieldEls[f.name].setError(''); });
      fieldEls[f.name] = w;
      fieldsBox.appendChild(w);
    });
    var sandboxText = h('p', { class: 'text-xs mt-1.5' });
    var sandboxBox = h('div', { class: 'rounded-md px-3 py-2.5 mb-4', 'data-sandbox-box': '' });
    var sandboxToggle = toggle(st.sandbox, t('pv.sandbox'), function (on) { st.sandbox = on; paintSandbox(); });
    sandboxBox.appendChild(h('div', { class: 'flex items-center gap-2' }, icon('flask', 'h-4 w-4 shrink-0'), sandboxToggle));
    sandboxBox.appendChild(sandboxText);
    function paintSandbox() {
      sandboxBox.className = 'rounded-md px-3 py-2.5 mb-4 ' + (st.sandbox ? 'eco-sandbox-note' : 'eco-live-note');
      sandboxText.textContent = st.sandbox ? t('pv.sandbox.on') : t('pv.sandbox.off');
    }
    paintSandbox();
    var defToggle = toggle(st.isDefault, t('pv.default'), function (on) { st.isDefault = on; if (on && !st.isActive) { st.isActive = true; activeToggle.set(true); } }, 'sm');
    var activeToggle = toggle(st.isActive, t('pm.active'), function (on) { st.isActive = on; if (!on && st.isDefault) { st.isDefault = false; defToggle.set(false); } }, 'sm');

    var saveBtn = button({ variant: 'primary', size: 'sm', label: t('save'), onclick: function () { save(); } });
    var del = h('div', { class: 'flex items-center gap-2' });
    function resetDelete() { clear(del).appendChild(button({ variant: 'danger', size: 'sm', icon: 'trash', title: t('delete'), onclick: askDelete })); }
    function askDelete() {
      clear(del);
      del.appendChild(h('span', { class: 'text-sm text-secondary-700 dark:text-secondary-200', text: t('delete.confirm') }));
      del.appendChild(button({ variant: 'dangerSolid', size: 'sm', label: t('delete.yes'), onclick: function () {
        pubApi('DELETE', '/providers/' + encodeURIComponent(p.id)).then(function () { toast('success', t('deleted')); load(); }, function (err) { resetDelete(); clear(errTop).appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err)))); });
      } }));
      del.appendChild(button({ variant: 'ghost', size: 'sm', label: t('no'), onclick: resetDelete }));
    }
    resetDelete();

    function save() {
      clear(errTop);
      Object.keys(fieldEls).forEach(function (k) { fieldEls[k].setError(''); });
      var goingLive = p.sandbox && !st.sandbox;
      var ask = goingLive ? confirmModal({ title: t('pv.live.title'), text: t('pv.live.text'), ok: t('pv.live.ok'), danger: true, icon: 'alert' }) : Promise.resolve(true);
      ask.then(function (yes) {
        if (!yes) return;
        var out = {};
        (p.fields || []).forEach(function (f) { if (cfg[f.name] !== undefined && cfg[f.name] !== null) out[f.name] = String(cfg[f.name]); });
        setBusy(saveBtn, true, t('saving'));
        pubApi('PUT', '/providers/' + encodeURIComponent(p.id), { isActive: st.isActive, isDefault: st.isDefault, sandbox: st.sandbox, config: out }).then(function () {
          toast('success', t('saved'));
          load();
        }, function (err) {
          setBusy(saveBtn, false);
          if (err.field && fieldEls[err.field]) fieldEls[err.field].setError(errorText(err));
          else errTop.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
        });
      });
    }

    var el = card(h('div', { class: 'p-4 flex flex-col h-full' },
      h('div', { class: 'flex items-start justify-between gap-3 mb-4' },
        h('div', { class: 'min-w-0' },
          h('div', { class: 'text-lg font-medium', text: kindLabel(p.kind) }),
          h('div', { class: 'flex flex-wrap items-center gap-1.5 mt-1' }, mono(p.kind),
            tag(p.configured ? t('pm.configured') : t('pm.notConfigured'), p.configured ? 'success' : 'secondary'),
            p.isDefault ? tag(t('pv.default'), 'primary') : null)),
        h('div', { class: 'flex flex-col items-end gap-2 shrink-0' }, activeToggle, defToggle)),
      sandboxBox,
      errTop,
      h('div', { class: 'flex-1' }, fieldsBox),
      h('div', { class: 'flex items-center justify-between gap-2 pt-2' }, del, saveBtn),
      testBox(p)));
    el.setAttribute('data-provider', p.kind);
    return el;
  }

  function addControl(existing) {
    var free = catalog.filter(function (k) { return k.channel === channel && existing.indexOf(k.kind) === -1; });
    if (!free.length) return h('span', { class: 'text-xs text-secondary-500', text: t('pv.allAdded') });
    var kind = free[0].kind;
    var sel = h('select', { class: SELECT_SM, 'aria-label': t('pv.kind'), onchange: function () { kind = sel.value; } }, free.map(function (k) { return h('option', { value: k.kind, text: kindLabel(k.kind) }); }));
    var add = button({ variant: 'primary', size: 'page', icon: 'plus', label: t('pv.add'), onclick: function () {
      setBusy(add, true);
      pubApi('POST', '/providers', { channel: channel, kind: kind, sandbox: true, isActive: false, isDefault: false, config: {} }).then(function () {
        toast('success', t('pv.added'));
        load();
      }, function (err) { setBusy(add, false); toast('error', errorText(err)); });
    } });
    return h('div', { class: 'flex items-center gap-2' }, h('label', { class: 'x-label block relative x-input', style: INPUT_VARS }, sel), add);
  }

  function load() {
    skeleton();
    pubApi('GET', '/providers').then(function (res) {
      if (!container.isConnected) return;
      var items = (res && res.items) || [];
      catalog = (res && res.catalog) || [];
      var list = items.filter(function (p) { return p.channel === channel; });
      var grid = h('div', { class: 'eco-card-grid' });
      if (!list.length) grid.appendChild(h('div', { class: 'eco-span-all' }, card(emptyBlock(t('pv.empty.' + channel), t('pv.empty.hint')))));
      list.forEach(function (p) { grid.appendChild(providerCard(p)); });
      clear(frame.actions).appendChild(addControl(list.map(function (p) { return p.kind; })));
      clear(body).appendChild(grid);
    }, function (err) { clear(frame.actions); clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}
