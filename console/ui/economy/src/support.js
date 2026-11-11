var TICKET_STATUSES = ['open', 'awaiting_staff', 'awaiting_player', 'resolved', 'closed'];
var TICKET_STATUS_COLOR = { open: 'warning', awaiting_staff: 'warning', awaiting_player: 'primary', resolved: 'success', closed: 'secondary' };
var TICKET_PRIORITIES = ['low', 'normal', 'high', 'urgent'];
var PRIORITY_COLOR = { low: 'secondary', normal: 'secondary', high: 'warning', urgent: 'error' };
var SUPPORT = { summary: null, at: 0, loading: false, agents: null, agentsAt: 0, beat: null, away: false };
try { SUPPORT.away = localStorage.getItem('economy.support.away') === '1'; } catch (e) { SUPPORT.away = false; }

function supportBadge(page) {
  var s = SUPPORT.summary;
  if (!s) return 0;
  if (page === 'tickets') return s.awaitingStaff || 0;
  if (page === 'live-chat') return s.waitingChats || 0;
  return 0;
}

function supportBadgeKey() {
  return supportBadge('tickets') + ':' + supportBadge('live-chat');
}

function setSummary(res) {
  if (!res || typeof res.awaitingStaff !== 'number') return;
  var before = supportBadgeKey();
  SUPPORT.summary = res;
  if (supportBadgeKey() !== before) schedule();
}

function refreshSupportBadge(force) {
  var now = Date.now();
  if (SUPPORT.loading || (!force && now - SUPPORT.at < 30000)) return;
  if (!token() || !aclCan('ACCOUNT_NOTES', 0)) return;
  SUPPORT.at = now;
  SUPPORT.loading = true;
  quietBase('GET', '/support/summary').then(function (res) { SUPPORT.loading = false; setSummary(res); }, function () { SUPPORT.loading = false; });
}

function onSupportPage() {
  return /^#\/support\//.test(location.hash || '');
}

function sendBeat() {
  if (!token() || !aclCan('ACCOUNT_NOTES', 1)) return Promise.resolve(null);
  return quietBase('POST', '/support/heartbeat', { away: SUPPORT.away }).then(function (res) { setSummary(res); return res; });
}

function supportBeat() {
  if (SUPPORT.beat) return;
  sendBeat();
  SUPPORT.beat = setInterval(function () {
    if (!onSupportPage() || !token()) { clearInterval(SUPPORT.beat); SUPPORT.beat = null; return; }
    sendBeat();
  }, 20000);
}

function loadAgents(force) {
  if (!force && SUPPORT.agents && Date.now() - SUPPORT.agentsAt < 15000) return Promise.resolve(SUPPORT.agents);
  return baseApi('GET', '/support/agents').then(function (res) {
    SUPPORT.agents = (res && res.items) || [];
    SUPPORT.me = res && res.me;
    SUPPORT.agentsAt = Date.now();
    return SUPPORT.agents;
  }, function () { return SUPPORT.agents || []; });
}

function ticketStatusTag(status) {
  var el = tag(t('tk.status.' + status), TICKET_STATUS_COLOR[status]);
  el.setAttribute('data-ticket-status', status);
  return el;
}

function priorityTag(p) {
  var el = tag(t('tk.priority.' + p), PRIORITY_COLOR[p]);
  el.setAttribute('data-priority', p);
  return el;
}

function deptName(d) {
  return d ? (i18nText(d.name) || d.key) : t('none');
}

function agentLabel(name) {
  return name ? name : t('tk.unassigned');
}

function minutesText(n) {
  if (n == null) return '—';
  if (n < 60) return t('dur.minutes', { n: localDigits(n) });
  if (n < 1440) return t('dur.hours', { n: localDigits(Math.round(n / 60)) });
  return t('dur.days', { n: localDigits(Math.round(n / 1440)) });
}

function ticketSubjectCell(it) {
  return h('div', { class: 'min-w-0' },
    h('div', { class: 'flex items-center gap-2 min-w-0' },
      h('span', { class: 'font-medium truncate', dir: 'auto', text: it.subject }),
      it.overdue ? tag(t('tk.overdue'), 'error') : null),
    h('div', { class: 'text-xs text-secondary-500 flex items-center gap-2 min-w-0' }, mono(it.number), h('span', { class: 'truncate', dir: 'auto', text: it.lastPreview || '' })));
}

LISTS.tickets = {
  api: 'base',
  path: 'support/tickets',
  top: function () { return ticketsTop(); },
  create: function () { return ticketsCreate(); },
  filters: [
    { name: 'status', label: 'filter.status', options: function () { return [{ value: 'waiting', label: t('tk.filter.waiting') }, { value: 'active', label: t('tk.filter.active') }].concat(TICKET_STATUSES.map(function (s) { return { value: s, label: t('tk.status.' + s) }; })); } },
    { name: 'priority', label: 'filter.priority', options: function () { return TICKET_PRIORITIES.map(function (p) { return { value: p, label: t('tk.priority.' + p) }; }); } },
    { name: 'assignee', label: 'filter.assignee', options: function () { return [{ value: 'me', label: t('tk.mine') }, { value: 'none', label: t('tk.unassigned') }].concat((SUPPORT.agents || []).map(function (a) { return { value: a.username, label: a.username }; })); } },
    { name: 'departmentId', label: 'filter.department', type: 'ref', resource: 'support/departments', api: baseApi },
    { name: 'overdue', label: 'filter.overdue', options: function () { return [{ value: 'true', label: t('tk.overdueOnly') }]; } }
  ],
  columns: [
    { label: 'col.subject', render: ticketSubjectCell, skeleton: '60%' },
    { label: 'col.player', render: function (it) { return playerLink(it.userId, it.displayName || it.username); }, width: 150 },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-col items-start gap-1' }, ticketStatusTag(it.status), it.priority !== 'normal' ? priorityTag(it.priority) : null); }, width: 150 },
    { label: 'col.assignee', render: function (it) { return h('span', { class: it.assignee ? '' : 'text-secondary-400', text: agentLabel(it.assignee) }); }, width: 130 },
    { label: 'col.department', render: function (it) { return h('span', { class: 'truncate', text: deptName(it.department) }); }, width: 150 },
    { label: 'col.lastMessage', render: function (it) { return h('div', { class: 'flex flex-col' }, dateCell(it.lastMessageAt), h('span', { class: 'text-xs text-secondary-500', text: t('tk.by.' + it.lastMessageBy) })); }, width: 170 }
  ]
};

function supportStats(s) {
  var grid = h('div', { class: 'eco-stat-grid mb-4', 'data-support-stats': '' });
  grid.appendChild(statCard(t('tk.stat.waiting'), s.awaitingStaff, null, hashFor('tickets', null, { status: 'waiting' })));
  grid.appendChild(statCard(t('tk.stat.overdue'), s.overdue, null, hashFor('tickets', null, { status: 'waiting', overdue: 'true' }), { alert: s.overdue > 0 }));
  grid.appendChild(statCard(t('tk.stat.unassigned'), s.unassigned, null, hashFor('tickets', null, { status: 'active', assignee: 'none' })));
  grid.appendChild(statCard(t('tk.stat.mine'), s.mine, null, hashFor('tickets', null, { status: 'waiting', assignee: 'me' })));
  grid.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-1 h-full' },
    h('div', { class: 'font-medium text-secondary-700 dark:text-secondary-200', text: t('tk.stat.firstResponse') }),
    h('div', { class: 'text-2xl font-semibold tabular-nums', text: minutesText(s.avgFirstResponseMinutes) }),
    h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('tk.stat.30days') }))));
  grid.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-1 h-full' },
    h('div', { class: 'font-medium text-secondary-700 dark:text-secondary-200', text: t('tk.stat.rating') }),
    h('div', { class: 'text-2xl font-semibold tabular-nums', text: s.avgRating ? localDigits(s.avgRating) + ' / ' + localDigits(5) : '—' }),
    h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('tk.stat.rated', { n: s.rated || 0 }) }))));
  return grid;
}

function ticketsPage(container, query, ctx) {
  loadAgents();
  return listPage(container, 'tickets', query, ctx);
}

function ticketsTop() {
  var holder = h('div', { 'data-support-top': '' }, h('div', { class: 'eco-stat-grid mb-4' }, [1, 2, 3, 4].map(function () { return card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('6rem'), bar('3rem', 'h-7'))); })));
  baseApi('GET', '/support/summary').then(function (s) { setSummary(s); clear(holder).appendChild(supportStats(s)); }, function () { clear(holder); });
  return holder;
}

function ticketsCreate() {
  if (!aclCan('ACCOUNT_NOTES', 1)) return null;
  var btn = button({ variant: 'primary', size: 'page', icon: 'plus', label: t('tk.new'), onclick: function () { newTicketDialog().then(function (res) { if (res && res.id) go('tickets', res.id); }); } });
  btn.setAttribute('data-ticket-new', '');
  return btn;
}

function playerPicker(onchange) {
  return refPicker({ resource: 'support/players', api: baseApi, required: true, label: t('col.player'), placeholder: t('tk.pickPlayer'), onchange: onchange });
}

function departmentPicker(value, onchange) {
  return refPicker({ resource: 'support/departments', api: baseApi, value: value || '', label: t('col.department'), filter: { isActive: 'true' }, onchange: onchange });
}

function newTicketDialog() {
  var v = { userId: '', subject: '', departmentId: '', priority: 'normal', body: '' };
  var fields = {};
  fields.player = field(t('col.player'), playerPicker(function (x) { v.userId = x || ''; fields.player.setError(''); }));
  fields.subject = field(t('tk.subject'), textInput('', { label: t('tk.subject'), oninput: function (x) { v.subject = x; fields.subject.setError(''); } }));
  fields.dept = field(t('col.department'), departmentPicker('', function (x) { v.departmentId = x || ''; }), null, { optional: true });
  fields.priority = field(t('tk.priority'), segmented(TICKET_PRIORITIES.map(function (p) { return { value: p, label: t('tk.priority.' + p) }; }), 'normal', function (x) { v.priority = x; }, t('tk.priority')));
  fields.body = field(t('tk.message'), textInput('', { multiline: true, rows: 4, label: t('tk.message'), oninput: function (x) { v.body = x; fields.body.setError(''); } }), t('tk.new.hint'));
  return formDialog({
    id: 'ticket-new', title: t('tk.new'), text: t('tk.new.text'), icon: 'tickets', ok: t('tk.new.ok'),
    body: h('div', {}, fields.player, fields.subject, fields.dept, fields.priority, fields.body),
    validate: function () {
      var ok = true;
      if (!v.userId) { fields.player.setError(t('err.required')); ok = false; }
      if (!v.subject.trim()) { fields.subject.setError(t('err.required')); ok = false; }
      if (!v.body.trim()) { fields.body.setError(t('err.required')); ok = false; }
      return ok;
    },
    submit: function () { return baseApi('POST', '/support/tickets', v); }
  });
}

function fileSize(n) {
  if (n < 1024) return t('size.b', { n: localDigits(n) });
  if (n < 1048576) return t('size.kb', { n: localDigits(Math.round(n / 1024)) });
  return t('size.mb', { n: localDigits((n / 1048576).toFixed(1)) });
}

function attachmentList(list) {
  if (!list || !list.length) return null;
  return h('div', { class: 'flex flex-wrap gap-2 mt-2', 'data-attachments': '' }, list.map(function (a) {
    var isImg = /^image\//.test(a.type || '');
    return h('a', { href: a.url || '#', target: '_blank', rel: 'noopener noreferrer', class: 'eco-att', 'data-attachment': a.key, title: a.name },
      isImg && a.url ? h('img', { src: a.url, alt: a.name, class: 'eco-att-img', loading: 'lazy' }) : icon('clip', 'h-4 w-4'),
      h('span', { class: 'eco-att-name truncate', dir: 'auto', text: a.name }),
      h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: fileSize(a.size || 0) }));
  }));
}

function eventLine(m) {
  var d = m.eventData || {};
  var vars = {
    from: d.from ? (has('tk.status.' + d.from) ? t('tk.status.' + d.from) : has('tk.priority.' + d.from) ? t('tk.priority.' + d.from) : d.from) : '',
    to: d.to ? (has('tk.status.' + d.to) ? t('tk.status.' + d.to) : has('tk.priority.' + d.to) ? t('tk.priority.' + d.to) : d.to) : '',
    who: m.authorType === 'system' ? t('tk.system') : m.authorType === 'player' ? t('tk.player') : (m.author || ''),
    n: d.rating || 0
  };
  if (m.event === 'department_changed') vars.to = '';
  var key = 'tk.ev.' + m.event + (m.event === 'assigned' && !d.to ? '.none' : '') + (d.auto ? '.auto' : '');
  if (!has(key)) key = 'tk.ev.' + m.event;
  return h('li', { class: 'eco-msg-event', 'data-event': m.event }, h('span', { text: t(key, vars) }), h('span', { class: 'text-secondary-400 whitespace-nowrap', text: fmtDate(m.createdAt, true) }));
}

function messageBubble(m, opts) {
  opts = opts || {};
  if (m.kind === 'event') return eventLine(m);
  var staff = m.authorType === 'staff' || m.sender === 'agent';
  var note = m.kind === 'note';
  var cls = 'eco-msg' + (staff ? ' eco-msg--staff' : ' eco-msg--player') + (note ? ' eco-msg--note' : '') + (m.sender === 'system' ? ' eco-msg--system' : '');
  var who = note ? t('tk.note.by', { name: m.author || '' }) : staff ? (m.author || t('tk.staff')) : (opts.playerName || t('tk.player'));
  return h('li', { class: cls, 'data-msg-kind': m.kind || m.sender || '', 'data-seq': m.seq != null ? String(m.seq) : null },
    h('div', { class: 'eco-msg-head' }, h('span', { class: 'font-medium', text: who }), note ? tag(t('tk.note'), 'warning') : null, h('span', { class: 'text-secondary-500 whitespace-nowrap', text: fmtDate(m.createdAt, true) })),
    m.body ? h('div', { class: 'eco-msg-body', dir: 'auto', text: m.body }) : null,
    attachmentList(m.attachments));
}

function cannedPicker(getLang, departmentId, onpick) {
  var sel = h('select', { class: SELECT_SM + ' eco-canned', 'aria-label': t('tk.canned') }, h('option', { value: '', text: t('tk.canned.pick') }));
  var items = [];
  baseApi('GET', '/support/canned' + qs({ isActive: 'true', limit: 100, departmentId: departmentId || '' })).then(function (res) {
    items = (res && res.items) || [];
    items.forEach(function (c, i) { sel.appendChild(h('option', { value: String(i), text: i18nText(c.title) || c.key })); });
    if (!items.length) sel.appendChild(h('option', { value: '', disabled: true, text: t('tk.canned.none') }));
  }, function () { return null; });
  sel.addEventListener('change', function () {
    var c = items[Number(sel.value)];
    sel.value = '';
    if (!c) return;
    var l = getLang();
    var body = (c.body && (c.body[l] || c.body[LANGS.def])) || i18nText(c.body) || '';
    onpick(body);
  });
  return sel;
}

function insertAtCursor(el, text) {
  var start = el.selectionStart == null ? el.value.length : el.selectionStart;
  var end = el.selectionEnd == null ? el.value.length : el.selectionEnd;
  var before = el.value.slice(0, start);
  var glue = before && !/\s$/.test(before) ? '\n' : '';
  el.value = before + glue + text + el.value.slice(end);
  var pos = (before + glue + text).length;
  el.focus();
  try { el.setSelectionRange(pos, pos); } catch (e) { pos = 0; }
  el.dispatchEvent(new Event('input', { bubbles: true }));
}

function uploadButton(onfile) {
  var input = h('input', { type: 'file', accept: 'image/png,image/jpeg,image/webp,application/pdf', class: 'sr-only', 'data-attach-input': '' });
  var btn = button({ variant: 'ghost', size: 'sm', icon: 'clip', label: t('tk.attach'), onclick: function () { input.click(); } });
  input.addEventListener('change', function () {
    var f = input.files && input.files[0];
    input.value = '';
    if (!f) return;
    var fd = new FormData();
    fd.append('file', f);
    setBusy(btn, true, t('tk.uploading'));
    baseApi('POST', '/support/uploads', fd, true).then(function (a) { setBusy(btn, false); onfile(a); }, function (err) { setBusy(btn, false); toast('error', errorText(err)); });
  });
  return h('span', {}, input, btn);
}

function composer(opts) {
  var state = { body: '', internal: false, status: 'awaiting_player', files: [] };
  var area = textInput('', { multiline: true, rows: 4, label: t('tk.reply'), placeholder: t(opts.placeholder || 'tk.reply.ph'), oninput: function (v) { state.body = v; if (opts.oninput) opts.oninput(v); } });
  area.setAttribute('data-composer', '');
  var files = h('div', { class: 'flex flex-wrap gap-2' });
  function paintFiles() {
    clear(files);
    state.files.forEach(function (a, i) {
      files.appendChild(h('span', { class: 'eco-att', 'data-pending-attachment': a.key }, icon('clip', 'h-4 w-4'), h('span', { class: 'eco-att-name truncate', text: a.name }),
        button({ variant: 'muted', size: 'icon', icon: 'close', iconClass: 'h-3 w-3', title: t('remove'), onclick: function () { state.files.splice(i, 1); paintFiles(); } })));
    });
  }
  var tools = h('div', { class: 'flex flex-wrap items-center gap-2' }, cannedPicker(opts.lang, opts.departmentId, function (txt) { insertAtCursor(area, txt); }));
  if (opts.attachments) tools.appendChild(uploadButton(function (a) { state.files.push(a); paintFiles(); }));
  var after = null;
  if (opts.statusAfter) {
    after = segmented([{ value: 'awaiting_player', label: t('tk.after.awaiting_player') }, { value: 'resolved', label: t('tk.after.resolved') }, { value: 'closed', label: t('tk.after.closed') }], 'awaiting_player', function (v) { state.status = v; }, t('tk.after'));
    after.setAttribute('data-reply-status', '');
  }
  var noteToggle = opts.notes ? toggle(false, t('tk.internal'), function (v) { state.internal = v; box.classList.toggle('eco-composer--note', v); if (after) after.hidden = v; send.querySelector('span').textContent = v ? t('tk.saveNote') : t(opts.sendLabel || 'tk.send'); }, 'sm') : null;
  if (noteToggle) noteToggle.setAttribute('data-internal-toggle', '');
  var send = button({ variant: 'primary', size: 'sm', icon: 'send', label: t(opts.sendLabel || 'tk.send'), onclick: function () { submit(); } });
  send.setAttribute('data-send', '');
  var err = h('div', {});
  function submit() {
    clear(err);
    if (!state.body.trim() && !state.files.length) { err.appendChild(h('p', { class: 'text-sm text-error-600 dark:text-error-400 mb-2', text: t('err.required') })); area.focus(); return; }
    setBusy(send, true, t('saving'));
    opts.onsend({ body: state.body, internal: state.internal, status: state.status, attachments: state.files.map(function (f) { return f.key; }) }).then(function () {
      setBusy(send, false);
      state.body = '';
      area.value = '';
      state.files = [];
      paintFiles();
    }, function (e) {
      setBusy(send, false);
      err.appendChild(h('div', { class: 'mb-2' }, notice('error', errorText(e))));
    });
  }
  area.addEventListener('keydown', function (e) { if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) { e.preventDefault(); submit(); } });
  var box = h('div', { class: 'eco-composer', 'data-composer-box': '' }, err, area, files,
    h('div', { class: 'flex flex-wrap items-center gap-3 justify-between' }, tools, h('div', { class: 'flex flex-wrap items-center gap-3' }, noteToggle, after, send)));
  box.focus = function () { area.focus(); };
  return box;
}

function infoRow(label, value) {
  return h('div', { class: 'eco-info-row' }, h('dt', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: label }), h('dd', { class: 'text-sm' }, value));
}

function ticketDetail(container, id) {
  loadAgents();
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('tickets', null, lastQuery.tickets || {}); } });
  var frame = pageFrame(t('tk.detail'), '', backBtn);
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-ticket', 'data-ticket': id });
  frame.content.appendChild(grid);
  var write = aclCan('ACCOUNT_NOTES', 1);

  function skeleton() {
    clear(grid).appendChild(h('div', { class: 'eco-ticket-main' }, card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('50%', 'h-5'), bar('90%', 'h-16'), bar('80%', 'h-16'), bar('100%', 'h-24')))));
    grid.appendChild(h('div', { class: 'eco-ticket-side' }, card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('40%'), bar('70%'), bar('60%'), bar('50%')))));
  }

  function act(path, body) {
    return baseApi('POST', '/support/tickets/' + encodeURIComponent(id) + path, body).then(function (res) { render(res); refreshSupportBadge(true); return res; });
  }

  function render(tk) {
    var h1 = frame.el.querySelector('h1 span');
    if (h1) h1.textContent = tk.subject;
    clear(grid);
    var thread = h('ol', { class: 'eco-thread', 'data-thread': '' });
    (tk.messages || []).forEach(function (m) { thread.appendChild(messageBubble(m, { playerName: tk.displayName || tk.username })); });
    var main = h('div', { class: 'eco-ticket-main flex flex-col gap-4' },
      card(h('div', { class: 'p-4' },
        h('div', { class: 'flex flex-wrap items-center gap-2 mb-3' }, mono(tk.number), ticketStatusTag(tk.status), priorityTag(tk.priority), tk.overdue ? tag(t('tk.overdue'), 'error') : null, tk.source !== 'player' ? tag(t('tk.source.' + tk.source), 'secondary') : null),
        thread)));
    if (write && tk.status !== 'closed') {
      main.appendChild(card(h('div', { class: 'p-4' }, composer({ notes: true, statusAfter: true, attachments: true, departmentId: tk.departmentId, lang: function () { return tk.lang || LANGS.def; },
        onsend: function (v) { return act('/messages', v).then(function () { toast('success', v.internal ? t('tk.noteSaved') : t('tk.sent')); }); } }))));
    } else if (write) {
      main.appendChild(notice('secondary', t('tk.closedNote')));
      var noteOnly = composer({ notes: false, statusAfter: false, attachments: false, departmentId: tk.departmentId, lang: function () { return tk.lang || LANGS.def; }, sendLabel: 'tk.saveNote', placeholder: 'tk.note.ph',
        onsend: function (v) { return act('/messages', { body: v.body, internal: true }); } });
      main.appendChild(card(h('div', { class: 'p-4' }, noteOnly)));
    }
    grid.appendChild(main);

    var side = h('dl', { class: 'flex flex-col gap-3' });
    side.appendChild(infoRow(t('col.player'), h('div', { class: 'flex flex-col' }, playerLink(tk.userId, tk.displayName || tk.username), h('span', { class: 'text-xs text-secondary-500', text: t('tk.playerCounts', { tickets: (tk.player || {}).tickets || 0, chats: (tk.player || {}).chats || 0 }) }),
      h('a', { href: hashFor('tickets', null, { userId: tk.userId }), class: 'eco-link text-xs', text: t('tk.playerTickets') }),
      aclCan('ACCOUNT', 0) ? h('a', { href: hashFor('activity', null, { userId: tk.userId }), class: 'eco-link text-xs', text: t('act.playerLink') }) : null)));
    if (write) {
      var statusBtns = h('div', { class: 'flex flex-wrap gap-2', 'data-status-actions': '' });
      var moves = tk.status === 'closed' ? [] : tk.status === 'resolved' ? ['closed', 'awaiting_staff'] : ['resolved', 'closed'];
      if (tk.status === 'closed') moves = ['awaiting_staff'];
      moves.forEach(function (s) {
        var b = button({ variant: s === 'resolved' ? 'primary' : 'ghost', size: 'sm', label: t('tk.move.' + s), onclick: function () { setBusy(b, true); act('/status', { status: s }).then(function () { toast('success', t('tk.moved')); }, function (e) { setBusy(b, false); toast('error', errorText(e)); }); } });
        b.setAttribute('data-move', s);
        statusBtns.appendChild(b);
      });
      side.appendChild(infoRow(t('col.status'), h('div', { class: 'flex flex-col gap-2 items-start' }, ticketStatusTag(tk.status), statusBtns)));
      var agentSel = selectInput(tk.assignee || '', [{ value: '', label: t('tk.unassigned') }].concat((SUPPORT.agents || []).map(function (a) { return { value: a.username, label: a.username + (a.online ? ' • ' + t('lc.online') : '') }; })).concat(tk.assignee && !(SUPPORT.agents || []).some(function (a) { return a.username === tk.assignee; }) ? [{ value: tk.assignee, label: tk.assignee }] : []), function (v) {
        act('/assign', { assignee: v }).then(function () { toast('success', t('tk.assigned')); }, function (e) { toast('error', errorText(e)); });
      }, { label: t('col.assignee'), placeholder: false });
      agentSel.setAttribute('data-assignee', '');
      var mine = !tk.assignee || tk.assignee !== SUPPORT.me ? button({ variant: 'link', size: 'sm', label: t('tk.takeIt'), onclick: function () { act('/assign', { assignee: 'me' }).then(function () { toast('success', t('tk.assigned')); }); } }) : null;
      side.appendChild(infoRow(t('col.assignee'), h('div', { class: 'flex flex-col gap-1 items-start' }, agentSel, mine)));
      side.appendChild(infoRow(t('col.department'), departmentPicker(tk.departmentId, function (v) { if (v) act('/assign', { departmentId: v }).then(function () { toast('success', t('saved')); }, function (e) { toast('error', errorText(e)); }); })));
      var prio = segmented(TICKET_PRIORITIES.map(function (p) { return { value: p, label: t('tk.priority.' + p) }; }), tk.priority, function (v) { act('/assign', { priority: v }).then(function () { toast('success', t('saved')); }, function (e) { toast('error', errorText(e)); }); }, t('tk.priority'));
      prio.setAttribute('data-priority-pick', '');
      side.appendChild(infoRow(t('tk.priority'), prio));
    } else {
      side.appendChild(infoRow(t('col.status'), ticketStatusTag(tk.status)));
      side.appendChild(infoRow(t('col.assignee'), h('span', { text: agentLabel(tk.assignee) })));
      side.appendChild(infoRow(t('col.department'), h('span', { text: deptName(tk.department) })));
    }
    var times = h('div', { class: 'flex flex-col gap-1 text-sm', 'data-sla': '' },
      h('div', {}, t('tk.created') + ': ', h('span', { class: 'tabular-nums', text: fmtDate(tk.createdAt, true) })),
      tk.dueAt ? h('div', { class: tk.overdue ? 'text-error-600 dark:text-error-400' : '' }, t('tk.due') + ': ', h('span', { class: 'tabular-nums', text: fmtDate(tk.dueAt, true) })) : null,
      h('div', {}, t('tk.firstResponse') + ': ', h('span', { class: 'tabular-nums', text: tk.firstResponseAt ? fmtDate(tk.firstResponseAt, true) : t('tk.notYet') })),
      tk.resolvedAt ? h('div', {}, t('tk.resolvedAt') + ': ', h('span', { class: 'tabular-nums', text: fmtDate(tk.resolvedAt, true) })) : null,
      tk.closedAt ? h('div', {}, t('tk.closedAt') + ': ', h('span', { class: 'tabular-nums', text: fmtDate(tk.closedAt, true) })) : null);
    side.appendChild(infoRow(t('tk.times'), times));
    if (tk.rating) {
      side.appendChild(infoRow(t('tk.rating'), h('div', { 'data-rating': String(tk.rating) }, h('div', { class: 'eco-stars', 'aria-label': t('tk.ratingN', { n: tk.rating }) }, [1, 2, 3, 4, 5].map(function (i) { return h('span', { class: i <= tk.rating ? 'eco-star-on' : 'eco-star-off' }, icon('star', 'h-4 w-4')); })),
        tk.ratingComment ? h('p', { class: 'text-sm mt-1', dir: 'auto', text: tk.ratingComment }) : null)));
    }
    if (tk.chatId) side.appendChild(infoRow(t('tk.fromChat'), h('a', { href: hashFor('live-chat', tk.chatId), class: 'eco-link text-sm', text: t('tk.openChat') })));
    grid.appendChild(h('div', { class: 'eco-ticket-side' }, card(h('div', { class: 'p-4' }, side))));
  }

  function load() {
    skeleton();
    baseApi('GET', '/support/tickets/' + encodeURIComponent(id)).then(render, function (err) { clear(grid).appendChild(errorBlock(err, load)); });
  }
  load();
}

function waitText(sec) {
  if (sec < 60) return t('lc.waitSec', { n: localDigits(Math.max(1, sec)) });
  return t('lc.waitMin', { n: localDigits(Math.floor(sec / 60)) });
}

function liveChatPage(container, id, query) {
  loadAgents();
  supportBeat();
  var view = query.view || (id ? 'mine' : 'queue');
  var me = { id: id, seq: 0, stamp: '', poll: 0, chat: null, typingAt: 0 };
  var awayToggle = toggle(!SUPPORT.away, t('lc.available'), function (v) {
    SUPPORT.away = !v;
    try { localStorage.setItem('economy.support.away', SUPPORT.away ? '1' : '0'); } catch (e) { SUPPORT.away = !v; }
    sendBeat().then(function () { toast('success', t(v ? 'lc.nowOnline' : 'lc.nowAway')); });
  }, 'sm');
  awayToggle.setAttribute('data-available', '');
  var frame = pageFrame(t('nav.live-chat'), t('desc.live-chat'), awayToggle);
  clear(container).appendChild(frame.el);
  var listBox = h('div', { class: 'eco-lc-list', 'data-lc-list': '' });
  var tabs = segmented([{ value: 'queue', label: t('lc.queue') }, { value: 'mine', label: t('lc.mine') }, { value: 'all', label: t('lc.all') }, { value: 'closed', label: t('lc.closed') }], view, function (v) {
    view = v;
    replaceHash(hashFor('live-chat', me.id, v === 'queue' && !me.id ? {} : { view: v }));
    loadList();
  }, t('lc.views'));
  tabs.setAttribute('data-lc-tabs', '');
  var pane = h('div', { class: 'eco-lc-pane', 'data-lc-pane': '' });
  frame.content.appendChild(h('div', { class: 'eco-lc' }, h('div', { class: 'eco-lc-side' }, card(h('div', { class: 'p-3 flex flex-col gap-3' }, tabs, listBox))), pane));

  function alive() { return container.isConnected && frame.el.isConnected; }

  function chatItem(c) {
    var on = c.id === me.id;
    var row = h('div', { class: 'eco-lc-item' + (on ? ' eco-lc-item--on' : ''), 'data-chat': c.id, 'data-chat-status': c.status, role: 'button', tabindex: '0',
      onclick: function (e) { if (!e.target.closest('button')) openChat(c.id); },
      onkeydown: function (e) { if (e.key === 'Enter') openChat(c.id); } },
      h('div', { class: 'flex items-center gap-2 min-w-0' },
        h('span', { class: 'eco-dot' + (c.playerOnline ? ' eco-dot--on' : ''), title: c.playerOnline ? t('lc.playerOnline') : t('lc.playerOffline') }),
        h('span', { class: 'font-medium truncate', text: c.displayName || c.username }),
        c.unread && c.agent === SUPPORT.me ? h('span', { class: 'eco-nav-badge tabular-nums', text: fmtNum(c.unread) }) : null,
        h('span', { class: 'flex-1' }),
        c.status === 'waiting' ? h('span', { class: 'text-xs text-warning-700 eco-wait whitespace-nowrap tabular-nums', text: waitText(c.waitingSeconds || 0) }) : c.agent ? h('span', { class: 'text-xs text-secondary-500 truncate', text: c.agent }) : null),
      h('div', { class: 'text-xs text-secondary-500 truncate', dir: 'auto', text: c.playerTyping ? t('lc.typing') : (c.topic ? c.topic + ' · ' : '') + (c.lastPreview || '') }));
    if (c.status === 'waiting' && aclCan('ACCOUNT_NOTES', 1)) {
      var take = button({ variant: 'primary', size: 'sm', label: t('lc.take'), onclick: function () { takeChat(c.id, take); } });
      take.setAttribute('data-take', c.id);
      row.appendChild(h('div', { class: 'mt-2' }, take));
    }
    return row;
  }

  var listSeq = 0;
  function loadList() {
    var my = ++listSeq;
    var params = { view: view === 'all' ? '' : view, limit: 50 };
    baseApi('GET', '/support/chats' + qs(params)).then(function (res) {
      if (my !== listSeq || !alive()) return;
      var items = (res && res.items) || [];
      clear(listBox);
      if (!items.length) listBox.appendChild(emptyBlock(t('lc.empty.' + view), view === 'queue' ? t('lc.empty.queue.hint') : null));
      items.forEach(function (c) { listBox.appendChild(chatItem(c)); });
    }, function (err) { if (my === listSeq) clear(listBox).appendChild(errorBlock(err, loadList)); });
  }

  function takeChat(cid, btn) {
    if (btn) setBusy(btn, true);
    baseApi('POST', '/support/chats/' + encodeURIComponent(cid) + '/take', {}).then(function () {
      toast('success', t('lc.taken'));
      view = 'mine';
      tabs.set('mine');
      refreshSupportBadge(true);
      openChat(cid);
      loadList();
    }, function (err) {
      if (btn) setBusy(btn, false);
      toast('error', err.code === 'ALREADY_TAKEN' ? t('code.ALREADY_TAKEN', { name: err.field === 'agent' ? err.message : err.message }) : errorText(err));
      loadList();
    });
  }

  function openChat(cid) {
    me.id = cid;
    me.seq = 0;
    me.stamp = '';
    me.chat = null;
    me.poll++;
    replaceHash(hashFor('live-chat', cid, view === 'queue' ? {} : { view: view }));
    Array.prototype.forEach.call(listBox.querySelectorAll('[data-chat]'), function (el) { el.classList.toggle('eco-lc-item--on', el.getAttribute('data-chat') === cid); });
    paneSkeleton();
    poll(me.poll, 0);
  }

  function paneSkeleton() {
    clear(pane).appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('40%', 'h-5'), bar('70%', 'h-10'), bar('60%', 'h-10'), bar('100%', 'h-20'))));
  }

  var thread = null;
  var header = null;
  var typingEl = null;
  function renderChat(c, fresh) {
    if (fresh || !thread) {
      clear(pane);
      header = h('div', { class: 'eco-lc-head' });
      thread = h('ol', { class: 'eco-thread eco-lc-thread', 'data-thread': '', 'aria-live': 'polite' });
      typingEl = h('div', { class: 'eco-typing', 'data-typing': '', hidden: true }, h('span', { class: 'eco-typing-dots', 'aria-hidden': 'true' }, h('i', {}), h('i', {}), h('i', {})), h('span', { text: t('lc.playerTypingNow') }));
      var foot = h('div', { 'data-lc-foot': '' });
      pane.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, header, thread, typingEl, foot)));
      pane.__foot = foot;
    }
    paintHeader(c);
    (c.messages || []).forEach(function (m) {
      if (thread.querySelector('[data-seq="' + m.seq + '"]')) return;
      thread.appendChild(chatLine(m, c));
    });
    if ((c.messages || []).length) thread.scrollTop = thread.scrollHeight;
    typingEl.hidden = !c.playerTyping;
    paintFoot(c);
  }

  function chatLine(m, c) {
    if (m.sender === 'system') {
      var d = m.eventData || {};
      var key = 'lc.ev.' + (m.event || 'note');
      return h('li', { class: 'eco-msg-event', 'data-seq': String(m.seq), 'data-event': m.event || '' }, h('span', { dir: 'auto', text: m.event === 'greeting' ? m.body : t(key, { agent: d.agent || d.to || '', from: d.from || '', to: d.to || t('lc.queue'), by: d.by === 'player' ? t('tk.player') : d.by === 'system' ? t('tk.system') : (d.by || '') }) }), h('span', { class: 'text-secondary-400 whitespace-nowrap', text: fmtDate(m.createdAt, true) }));
    }
    return messageBubble({ kind: 'message', sender: m.sender, authorType: m.sender === 'agent' ? 'staff' : 'player', author: m.author, body: m.body, createdAt: m.createdAt, seq: m.seq }, { playerName: c.displayName || c.username });
  }

  function paintHeader(c) {
    clear(header);
    header.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2' },
      h('span', { class: 'eco-dot' + (c.playerOnline ? ' eco-dot--on' : ''), 'data-player-online': c.playerOnline ? 'true' : 'false' }),
      playerLink(c.userId, c.displayName || c.username, 'font-medium'),
      h('span', { class: 'text-xs text-secondary-500', text: c.playerOnline ? t('lc.playerOnline') : t('lc.playerOffline') }),
      tag(t('lc.status.' + c.status), c.status === 'waiting' ? 'warning' : c.status === 'active' ? 'success' : 'secondary'),
      c.agent ? h('span', { class: 'text-xs text-secondary-500', text: t('lc.withAgent', { name: c.agent }) }) : null,
      c.topic ? h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300', dir: 'auto', text: '· ' + c.topic }) : null,
      h('span', { class: 'flex-1' }),
      c.ticketId ? h('a', { href: hashFor('tickets', c.ticketId), class: 'eco-link text-sm', 'data-chat-ticket': '', text: t('lc.ticketN', { n: c.ticketNumber || '' }) }) : null));
  }

  function paintFoot(c) {
    var foot = pane.__foot;
    var key = c.status + '|' + c.agent + '|' + (SUPPORT.me || '');
    if (foot.__key === key) return;
    foot.__key = key;
    clear(foot);
    var write = aclCan('ACCOUNT_NOTES', 1);
    if (!write) return;
    if (c.status === 'waiting') {
      var take = button({ variant: 'primary', size: 'sm', label: t('lc.take'), onclick: function () { takeChat(c.id, take); } });
      take.setAttribute('data-take', c.id);
      foot.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2' }, take, button({ variant: 'ghost', size: 'sm', label: t('lc.close'), onclick: function () { closeDialog(c); } })));
      return;
    }
    if (c.status === 'closed') {
      foot.appendChild(notice('secondary', c.rating ? t('lc.closedRated', { n: c.rating }) : t('lc.closedNote')));
      return;
    }
    if (c.agent !== SUPPORT.me) {
      foot.appendChild(notice('secondary', t('lc.someoneElse', { name: c.agent })));
      return;
    }
    var comp = composer({ notes: false, statusAfter: false, attachments: false, lang: function () { return c.lang || LANGS.def; }, departmentId: c.departmentId, placeholder: 'lc.reply.ph',
      oninput: function () {
        var now = Date.now();
        if (now - me.typingAt < 2500) return;
        me.typingAt = now;
        baseApi('POST', '/support/chats/' + encodeURIComponent(c.id) + '/typing', {}).then(null, function () { return null; });
      },
      onsend: function (v) { return baseApi('POST', '/support/chats/' + encodeURIComponent(c.id) + '/messages', { body: v.body }).then(function (res) { merge(res); }); } });
    var transfer = button({ variant: 'ghost', size: 'sm', icon: 'users', label: t('lc.transfer'), onclick: function () { transferDialog(c); } });
    transfer.setAttribute('data-transfer', '');
    var close = button({ variant: 'danger', size: 'sm', label: t('lc.close'), onclick: function () { closeDialog(c); } });
    close.setAttribute('data-close-chat', '');
    foot.appendChild(comp);
    foot.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 mt-2' }, transfer, close));
  }

  function merge(res) {
    if (!res || res.id !== me.id) return;
    var fresh = !me.chat;
    me.chat = res;
    me.stamp = res.stamp || me.stamp;
    var msgs = res.messages || [];
    if (msgs.length) me.seq = Math.max(me.seq, msgs[msgs.length - 1].seq);
    renderChat(res, fresh);
    if (msgs.length && res.agent === SUPPORT.me && res.status === 'active') baseApi('POST', '/support/chats/' + encodeURIComponent(res.id) + '/read', { seq: me.seq }).then(null, function () { return null; });
  }

  function poll(token2, wait) {
    if (token2 !== me.poll || !alive() || !me.id) return;
    baseApi('GET', '/support/chats/' + encodeURIComponent(me.id) + qs({ after: me.seq, wait: wait, stamp: me.stamp })).then(function (res) {
      if (token2 !== me.poll || !alive()) return;
      merge(res);
      setTimeout(function () { poll(token2, 15); }, 150);
    }, function (err) {
      if (token2 !== me.poll || !alive()) return;
      if (!me.chat) { clear(pane).appendChild(errorBlock(err, function () { openChat(me.id); })); return; }
      setTimeout(function () { poll(token2, 15); }, 3000);
    });
  }

  function transferDialog(c) {
    var to = '';
    return loadAgents(true).then(function (agents) {
      var options = [{ value: '', label: t('lc.backToQueue') }].concat(agents.filter(function (a) { return a.username !== SUPPORT.me; }).map(function (a) { return { value: a.username, label: a.username + ' · ' + (a.online ? t('lc.online') : t('lc.offline')) + ' · ' + t('lc.activeN', { n: a.activeChats || 0 }) }; }));
      var sel = selectInput('', options, function (v) { to = v; }, { label: t('lc.transferTo'), placeholder: false });
      sel.setAttribute('data-transfer-to', '');
      return formDialog({ id: 'chat-transfer', title: t('lc.transfer'), text: t('lc.transfer.text'), icon: 'users', ok: t('lc.transfer.ok'), body: field(t('lc.transferTo'), sel),
        submit: function () { return baseApi('POST', '/support/chats/' + encodeURIComponent(c.id) + '/transfer', { to: to }); } });
    }).then(function (res) {
      if (!res) return;
      toast('success', t('lc.transferred'));
      me.stamp = '';
      if (pane.__foot) pane.__foot.__key = '';
      loadList();
    });
  }

  function closeDialog(c) {
    var toTicket = false;
    var box = checkbox(false, t('lc.toTicket'), function (on) { toTicket = on; });
    box.setAttribute('data-to-ticket', '');
    return formDialog({ id: 'chat-close', title: t('lc.close.title'), text: t('lc.close.text'), icon: 'live-chat', ok: t('lc.close'), danger: true, body: h('div', {}, box, h('p', { class: 'text-xs text-secondary-500 mt-1', text: t('lc.toTicket.hint') })),
      submit: function () { return baseApi('POST', '/support/chats/' + encodeURIComponent(c.id) + '/close', { toTicket: toTicket }); } }).then(function (res) {
      if (!res) return;
      toast('success', res.ticketId ? t('lc.closedTicket', { n: res.ticketNumber || '' }) : t('lc.closedDone'));
      if (pane.__foot) pane.__foot.__key = '';
      merge(res);
      loadList();
      refreshSupportBadge(true);
    });
  }

  loadList();
  var listTimer = setInterval(function () { if (!alive()) { clearInterval(listTimer); return; } loadList(); }, 4000);
  if (id) openChat(id);
  else clear(pane).appendChild(card(emptyBlock(t('lc.pick'), t('lc.pick.hint'))));
}

RESOURCES['support-departments'] = {
  api: 'base',
  path: 'support/departments',
  settings: true,
  filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, false); }, skeleton: '50%' },
    { label: 'col.firstResponse', render: function (it) { return h('span', { class: 'tabular-nums', text: t('dur.hours', { n: localDigits(it.firstResponseHours) }) }); }, width: 130 },
    { label: 'col.agents', render: function (it) { return (it.agents || []).length ? h('div', { class: 'flex flex-wrap gap-1' }, it.agents.slice(0, 4).map(function (a) { return tag(a, 'secondary'); })) : h('span', { class: 'text-secondary-400', text: t('dp.everyone') }); }, width: 200 },
    { label: 'col.open', render: function (it) { return it.open ? h('a', { href: hashFor('tickets', null, { departmentId: it.id, status: 'active' }), class: 'eco-link tabular-nums', text: fmtNum(it.open) }) : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 100 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.dp.key.hint', required: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { label: 'f.dp.name', hint: 'f.dp.name.hint', required: true, max: 100 }),
    tr('description', { label: 'f.dp.description', hint: 'f.dp.description.hint', multiline: true, max: 300, optional: true }),
    { name: 'firstResponseHours', type: 'number', label: 'f.dp.hours', hint: 'f.dp.hours.hint', required: true, min: 1, max: 720 },
    { name: 'agents', type: 'multi', label: 'f.dp.agents', hint: 'f.dp.agents.hint', options: function () {
      var names = (SUPPORT.agents || []).map(function (a) { return a.username; });
      return names.map(function (n) { return { value: n, label: n }; });
    } },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.dp.active.hint' }
  ],
  prepare: function () { return loadAgents(); },
  defaults: function () { return { key: '', name: {}, description: {}, firstResponseHours: 24, agents: [], isActive: true, sort: 100 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 100; if (!p.agents) p.agents = []; return p; }
};

RESOURCES['support-canned'] = {
  api: 'base',
  path: 'support/canned',
  settings: true,
  filters: [
    { name: 'departmentId', label: 'filter.department', type: 'ref', resource: 'support/departments', api: baseApi },
    { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }
  ],
  columns: [
    { label: 'col.title', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.title) || it.key }), h('div', { class: 'text-xs text-secondary-500 eco-clamp', dir: 'auto', text: i18nText(it.body) })); }, skeleton: '60%' },
    { label: 'col.department', render: function (it) { return h('span', { class: it.department ? '' : 'text-secondary-400', text: it.department ? deptName(it.department) : t('cn.allDepts') }); }, width: 170 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.cn.key.hint', required: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('title', { label: 'f.cn.title', hint: 'f.cn.title.hint', required: true, max: 100 }),
    tr('body', { label: 'f.cn.body', hint: 'f.cn.body.hint', required: true, multiline: true, rows: 5, max: 3000 }),
    { name: 'departmentId', type: 'ref', resource: 'support/departments', api: baseApi, label: 'f.cn.department', hint: 'f.cn.department.hint', optional: true },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.cn.active.hint' }
  ],
  defaults: function () { return { key: '', title: {}, body: {}, departmentId: '', isActive: true, sort: 100 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 100; if (!p.departmentId) p.departmentId = ''; return p; }
};

var SUPPORT_SETTINGS_FIELDS = [
  { type: 'heading', label: 'ss.tickets' },
  { type: 'row', fields: [
    { name: 'maxOpenTickets', type: 'number', label: 'f.ss.maxOpen', hint: 'f.ss.maxOpen.hint', required: true, min: 1, max: 100 },
    { name: 'ticketsPerHour', type: 'number', label: 'f.ss.perHour', hint: 'f.ss.perHour.hint', required: true, min: 1, max: 1000 }
  ] },
  { type: 'row', fields: [
    { name: 'messagesPerHour', type: 'number', label: 'f.ss.messages', hint: 'f.ss.messages.hint', required: true, min: 1, max: 10000 },
    { name: 'autoCloseDays', type: 'number', label: 'f.ss.autoClose', hint: 'f.ss.autoClose.hint', required: true, min: 1, max: 365 }
  ] },
  { type: 'row', fields: [
    { name: 'maxAttachments', type: 'number', label: 'f.ss.attachments', hint: 'f.ss.attachments.hint', required: true, min: 0, max: 10 },
    { name: 'maxAttachmentKb', type: 'number', label: 'f.ss.attachmentKb', hint: 'f.ss.attachmentKb.hint', required: true, min: 16, max: 10240 }
  ] },
  { type: 'heading', label: 'ss.chat' },
  { name: 'chatEnabled', type: 'bool', inline: true, label: 'f.ss.chatEnabled', hint: 'f.ss.chatEnabled.hint' },
  tr('chatGreeting', { label: 'f.ss.greeting', hint: 'f.ss.greeting.hint', multiline: true, rows: 2, max: 500, optional: true }),
  tr('chatOffline', { label: 'f.ss.offline', hint: 'f.ss.offline.hint', multiline: true, rows: 2, max: 500, optional: true }),
  { type: 'row', fields: [
    { name: 'chatIdleMinutes', type: 'number', label: 'f.ss.idle', hint: 'f.ss.idle.hint', required: true, min: 5, max: 1440 },
    { name: 'chatMessagesPerMinute', type: 'number', label: 'f.ss.chatRate', hint: 'f.ss.chatRate.hint', required: true, min: 1, max: 600 }
  ] },
  { type: 'heading', label: 'ss.activity' },
  { name: 'activityRetentionDays', type: 'number', label: 'f.ss.retention', hint: 'f.ss.retention.hint', required: true, min: 7, max: 3650 }
];

function supportSettingsPage(container) {
  var frame = tabFrame(t('desc.support-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(baseApi, '/support/settings', null, SUPPORT_SETTINGS_FIELDS, { id: 'support-settings' });
  frame.content.appendChild(page.el);
  page.load();
}
