ICONS['comm-settings'] = ICONS['sticker-settings'];
ICONS['comm-channels'] = '<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>';
ICONS['voice-rooms'] = '<path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3z"/><path d="M19 10v2a7 7 0 0 1-14 0v-2M12 19v3"/>';
PH_ICONS['voice-rooms'] = PH_ICONS['stickers'];
PH_ICONS['comm-channels'] = PH_ICONS['stickers'];
PH_ICONS['comm-settings'] = PH_ICONS['stickers'];

var CM_SETTINGS_FIELDS = [
  { name: 'enabled', type: 'bool', inline: true, label: 'f.cm.enabled', hint: 'f.cm.enabled.hint' },
  { type: 'heading', label: 'cm.h.cooldown' },
  { name: 'about.cooldown', type: 'static', readonly: true, render: function () { return notice('info', t('cm.cooldown.about')); } },
  { type: 'row', fields: [
    { name: 'minIntervalSeconds', type: 'number', label: 'f.cm.minInterval', hint: 'f.cm.minInterval.hint', required: true, min: 0, max: 60 },
    { name: 'lockSeconds', type: 'number', label: 'f.cm.lock', hint: 'f.cm.lock.hint', required: true, min: 1, max: 3600 }
  ] },
  { type: 'row', fields: [
    { name: 'maxCount', type: 'number', label: 'f.cm.maxCount', hint: 'f.cm.maxCount.hint', required: true, min: 1, max: 1000 },
    { name: 'windowSeconds', type: 'number', label: 'f.cm.window', hint: 'f.cm.window.hint', required: true, min: 1, max: 600, check: function (v, vals) { return v != null && vals.maxCount != null && vals.minIntervalSeconds != null && Number(vals.maxCount) * Number(vals.minIntervalSeconds) >= Number(v) ? t('cm.err.window') : null; } }
  ] },
  { type: 'heading', label: 'cm.h.more' },
  { type: 'row', fields: [
    { name: 'channelHours', type: 'number', label: 'f.cm.channelHours', hint: 'f.cm.channelHours.hint', required: true, min: 1, max: 48 },
    { name: 'maxFavorites', type: 'number', label: 'f.cm.maxFavorites', hint: 'f.cm.maxFavorites.hint', required: true, min: 0, max: 1000 }
  ] }
];

function commSettingsPage(container) {
  var frame = pageFrame(t('nav.comm-settings'), t('desc.comm-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(baseApi, '/comms/settings', null, CM_SETTINGS_FIELDS, {
    id: 'comms',
    fixPayload: function (p) {
      ['updatedAt', 'updatedBy', 'about'].forEach(function (k) { delete p[k]; });
      return p;
    }
  });
  frame.content.appendChild(page.el);
  page.load();
}

function cmStatusTag(status) {
  var color = status === 'open' ? 'success' : status === 'expired' ? 'warning' : 'secondary';
  return h('span', { 'data-cm-status': status }, tag(t('cm.status.' + status), color));
}

function cmMembers(ch, reload) {
  var box = h('div', { class: 'flex flex-col gap-1' });
  (ch.members || []).forEach(function (m) {
    var row = h('div', { class: 'flex items-center gap-2 flex-wrap', 'data-cm-member': m.userId },
      playerLink(m.userId, m.displayName || m.username),
      h('span', { class: 'text-xs text-secondary-500 tabular-nums', text: t('cm.sentN', { n: fmtNum(m.sent || 0) }) }),
      m.canSend ? null : tag(t('cm.listenOnly'), 'warning'));
    if (ch.status === 'open') {
      var btn = button({ variant: 'link', size: 'sm', label: t(m.canSend ? 'cm.blockSend' : 'cm.allowSend'), onclick: function () {
        setBusy(btn, true);
        baseApi('POST', '/comms/channels/' + encodeURIComponent(ch.matchId) + '/can-send', { userId: m.userId, canSend: !m.canSend }).then(function () {
          toast('success', t('saved'));
          reload();
        }, function (err) { setBusy(btn, false); toast('error', errorText(err)); });
      } });
      btn.setAttribute('data-cm-can-send', m.userId);
      row.appendChild(btn);
    }
    box.appendChild(row);
  });
  if (!(ch.members || []).length) box.appendChild(h('span', { class: 'text-xs text-secondary-500', text: t('none') }));
  return box;
}

function commChannelsPage(container, query) {
  var frame = pageFrame(t('nav.comm-channels'), t('desc.comm-channels'));
  clear(container).appendChild(frame.el);
  var state = { status: query.status || 'open', q: query.q || '', page: 1, limit: LIMIT, total: 0 };
  var box = h('div', { class: 'flex flex-col gap-3' });
  var status = filterSelect(t('filter.status'), state.status, ['open', 'closed', 'all'].map(function (s) { return { value: s, label: t('cm.filter.' + s) }; }), function (v) { state.status = v || 'open'; state.page = 1; load(); });
  var timer = null;
  var search = searchInput(state.q, t('search.comm-channels'), function (e) { clearTimeout(timer); timer = setTimeout(function () { state.q = e.target.value.trim(); state.page = 1; load(); }, 250); });
  var pagerBox = h('div', { class: 'eco-pager-slot' });
  var countBox = h('div', { class: 'eco-count-slot' });
  frame.content.appendChild(listToolbar({ search: search, filters: [status], meta: countBox, refresh: refreshButton(function () { load(); }) }));
  frame.content.appendChild(box);
  frame.content.appendChild(pagerBox);
  var cols = [{ label: t('cm.col.match'), skeleton: '60%' }, { label: t('col.status'), width: 110 }, { label: t('cm.col.members'), skeleton: '50%' }, { label: t('cm.col.messages'), width: 90 }, { label: t('cm.col.expires'), width: 150 }, { label: '', width: 120 }];
  function load() {
    clear(box).appendChild(table(cols, skeletonRows(cols, 3)));
    baseApi('GET', '/comms/channels' + qs({ status: state.status, q: state.q, page: state.page, limit: state.limit })).then(function (res) {
      state.total = res.total || 0;
      state.loaded = true;
      clear(countBox).appendChild(listCount(state));
      clear(pagerBox).appendChild(pager(state, function (p) { state.page = p; load(); }));
      var list = res.items || [];
      clear(box);
      if (!list.length) {
        box.appendChild(card(emptyBlock(t('empty.comm-channels'), t('empty.hint.comm-channels'))));
        return;
      }
      box.appendChild(table(cols, list.map(function (ch) {
        var action = null;
        if (ch.status === 'open') {
          action = button({ variant: 'danger', size: 'sm', label: t('cm.close'), onclick: function () {
            confirmModal({ title: t('cm.close.title'), text: t('cm.close.text'), ok: t('cm.close'), danger: true }).then(function (ok) {
              if (!ok) return;
              baseApi('POST', '/comms/channels/' + encodeURIComponent(ch.matchId) + '/close', {}).then(function () { toast('success', t('cm.closed')); load(); }, function (err) { toast('error', errorText(err)); });
            });
          } });
          action.setAttribute('data-cm-close', ch.matchId);
        } else {
          action = button({ variant: 'secondary', size: 'sm', label: t('cm.open'), onclick: function () {
            baseApi('POST', '/comms/channels/' + encodeURIComponent(ch.matchId) + '/open', {}).then(function () { toast('success', t('cm.opened')); load(); }, function (err) { toast('error', errorText(err)); });
          } });
          action.setAttribute('data-cm-open', ch.matchId);
        }
        return { item: ch, cells: [
          h('div', { class: 'flex flex-col min-w-0', 'data-cm-channel': ch.matchId }, mono(ch.matchId), ch.gameKey ? h('span', { class: 'text-xs text-secondary-500', text: ch.gameKey }) : null),
          h('div', { class: 'flex flex-col gap-1' }, cmStatusTag(ch.status), ch.closeReason ? h('span', { class: 'text-xs text-secondary-500', text: t('cm.reason.' + ch.closeReason) }) : null),
          cmMembers(ch, load),
          h('span', { class: 'tabular-nums', text: fmtNum(ch.messages || 0) }),
          h('span', { class: 'text-xs', text: fmtDate(ch.expiresAt, true) }),
          action
        ] };
      })));
    }, function (err) { clear(box).appendChild(errorBlock(err, load)); });
  }
  load();
}

function voiceRoomsPage(container, query) {
  var frame = pageFrame(t('nav.voice-rooms'), t('desc.voice-rooms'));
  clear(container).appendChild(frame.el);
  var state = { q: query.q || '' };
  var box = h('div', { class: 'flex flex-col gap-3' });
  var timer = null;
  var search = searchInput(state.q, t('search.voice-rooms'), function (e) { clearTimeout(timer); timer = setTimeout(function () { state.q = e.target.value.trim(); load(); }, 250); });
  frame.content.appendChild(listToolbar({ search: search, refresh: refreshButton(function () { load(); }) }));
  frame.content.appendChild(box);
  var cols = [{ label: t('vr.col.room'), skeleton: '50%' }, { label: t('vr.col.player'), skeleton: '40%' }, { label: t('vr.col.mic'), width: 150 }, { label: t('vr.col.expires'), width: 150 }, { label: '', width: 120 }];
  function load() {
    clear(box).appendChild(table(cols, skeletonRows(cols, 3)));
    baseApi('GET', '/voice/rooms' + qs({ q: state.q, live: 'true' })).then(function (res) {
      var rows = [];
      (res.items || []).forEach(function (room) {
        (room.members || []).forEach(function (m) {
          rows.push({ room: room, m: m });
        });
      });
      clear(box);
      if (!rows.length) {
        box.appendChild(card(emptyBlock(t('empty.voice-rooms'), t('empty.hint.voice-rooms'))));
        return;
      }
      box.appendChild(table(cols, rows.map(function (r) {
        var m = r.m;
        var name = m.displayName || m.username || m.name;
        var mic = m.muted ? tag(t('vr.staffMuted'), 'error') : m.canPublish ? tag(t('vr.canPublish'), 'success') : tag(t('vr.listenOnly'), 'warning');
        var btn = button({ variant: m.muted ? 'secondary' : 'danger', size: 'sm', label: t(m.muted ? 'vr.unmute' : 'vr.mute'), onclick: function () {
          var go = function () {
            setBusy(btn, true);
            baseApi('POST', '/voice/rooms/' + encodeURIComponent(r.room.room) + '/' + (m.muted ? 'unmute' : 'mute'), { userId: m.userId }).then(function (out) {
              toast(out.applied ? 'success' : 'warning', t(out.applied ? (m.muted ? 'vr.unmuted' : 'vr.muted') : 'vr.savedNotLive'));
              load();
            }, function (err) { setBusy(btn, false); toast('error', errorText(err)); });
          };
          if (m.muted) return go();
          confirmModal({ title: t('vr.mute.title', { name: name }), text: t('vr.mute.text'), ok: t('vr.mute'), danger: true }).then(function (ok) { if (ok) go(); });
        } });
        btn.setAttribute('data-vr-action', m.userId);
        return { item: r, cells: [
          h('div', { class: 'flex flex-col min-w-0', 'data-vr-room': r.room.room }, mono(r.room.room), h('span', { class: 'text-xs text-secondary-500', text: [t('vr.kind.' + r.room.kind), r.room.gameKey, r.room.provider].filter(Boolean).join(' · ') })),
          h('div', { class: 'flex items-center gap-2' }, playerLink(m.userId, name), m.connected === true ? tag(t('vr.online'), 'success') : m.connected === false ? tag(t('vr.offline'), 'secondary') : null),
          h('span', { 'data-vr-mic': m.muted ? 'muted' : 'open' }, mic),
          h('span', { class: 'text-xs', text: fmtDate(r.room.expiresAt, true) }),
          btn
        ] };
      })));
    }, function (err) { clear(box).appendChild(errorBlock(err, load)); });
  }
  load();
}

CUSTOM_PAGES['comm-settings'] = function (root) { commSettingsPage(root); };
CUSTOM_PAGES['comm-channels'] = function (root, route) { commChannelsPage(root, route.query || {}); };
CUSTOM_PAGES['voice-rooms'] = function (root, route) { voiceRoomsPage(root, route.query || {}); };
