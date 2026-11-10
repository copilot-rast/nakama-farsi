var RESTRICT_KINDS = ['suspend', 'mute', 'warn'];
var RESTRICT_COLOR = { suspend: 'error', mute: 'warning', warn: 'primary' };
var RESTRICT_STATUS_COLOR = { active: 'error', expired: 'secondary', lifted: 'success' };
var REPORT_STATUS_COLOR = { open: 'warning', dismissed: 'secondary', resolved: 'success' };
var WORD_ACTION_COLOR = { mask: 'primary', block: 'error', flag: 'warning' };
var DURATIONS = [60, 1440, 4320, 10080, 43200];
var REPORT_ACTIONS = ['dismiss', 'delete', 'warn', 'mute', 'suspend'];
var pendingReports = null;
var reportBadgeAt = 0;
var playerSel = { ids: {}, items: {}, bar: null, page: [] };

function refreshReportBadge(force) {
  var now = Date.now();
  if (!force && now - reportBadgeAt < 60000) return;
  reportBadgeAt = now;
  if (!aclCan('CHANNEL_MESSAGE', 0)) return;
  quietBase('GET', '/reports/summary').then(function (res) {
    var n = res && typeof res.open === 'number' ? res.open : null;
    if (n !== pendingReports) {
      pendingReports = n;
      schedule();
    }
  });
}

function durationLabel(min) {
  if (min < 1440) return t('dur.hours', { n: localDigits(min / 60) });
  return t('dur.days', { n: localDigits(min / 1440) });
}

function untilText(iso) {
  return t('rs.until', { date: fmtDate(iso, true) });
}

function playerLink(id, name, extra) {
  if (!id) return h('span', { class: 'text-secondary-400', text: name || t('rp.gone') });
  return h('a', { href: '#/players/' + encodeURIComponent(id), class: 'eco-link truncate' + (extra ? ' ' + extra : ''), text: name || String(id).slice(0, 8) });
}

function playerTags(tags, max) {
  var list = tags || [];
  var box = h('div', { class: 'flex flex-wrap gap-1', 'data-player-tags': '' });
  list.slice(0, max || 6).forEach(function (tg) {
    var chip = tag(i18nText(tg.name) || tg.key, tg.isPublic ? 'success' : 'primary');
    if (tg.isPublic) chip.setAttribute('title', t('tag.public'));
    box.appendChild(chip);
  });
  if (max && list.length > max) box.appendChild(tag('+' + fmtNum(list.length - max), 'secondary'));
  if (!list.length) box.appendChild(h('span', { class: 'text-secondary-400', text: t('none') }));
  return box;
}

function playerStatus(it) {
  var box = h('div', { class: 'flex flex-col items-start gap-1', 'data-player-status': '' });
  if (it.banned) box.appendChild(tag(t('pl.banned'), 'error'));
  if (it.suspension) box.appendChild(h('span', { class: 'eco-tag eco-tag--error inline-flex leading-tight max-w-full px-2 py-0.5 text-xs rounded-md whitespace-nowrap', 'data-status': 'suspended', title: untilText(it.suspension.endsAt), text: t('pl.suspendedUntil', { date: fmtDate(it.suspension.endsAt, true) }) }));
  if (it.mute) box.appendChild(h('span', { class: 'eco-tag eco-tag--warning inline-flex leading-tight max-w-full px-2 py-0.5 text-xs rounded-md whitespace-nowrap', 'data-status': 'muted', title: untilText(it.mute.endsAt), text: t('pl.mutedUntil', { date: fmtDate(it.mute.endsAt, true) }) }));
  if (!box.childNodes.length) box.appendChild(tag(t('pl.ok'), 'success'));
  return box;
}

function durationPicker(onchange) {
  var state = { minutes: 1440, endsAt: null };
  var custom = h('div', { class: 'mt-2', hidden: true });
  var date = dateInput(null, function (v) { state.endsAt = v; emit(); });
  custom.appendChild(date);
  var options = DURATIONS.map(function (m) { return { value: String(m), label: durationLabel(m) }; }).concat([{ value: 'custom', label: t('dur.custom') }]);
  var seg = segmented(options, '1440', function (v) {
    custom.hidden = v !== 'custom';
    if (v === 'custom') { state.minutes = 0; var inp = custom.querySelector('input'); if (inp) inp.focus(); }
    else { state.minutes = Number(v); state.endsAt = null; }
    emit();
  }, t('rs.duration'), 'md');
  seg.setAttribute('data-duration', '');
  function emit() { if (onchange) onchange(value()); }
  function value() { return state.minutes ? { minutes: state.minutes } : { endsAt: state.endsAt }; }
  var el = h('div', {}, seg, custom);
  el.value = value;
  el.check = function () {
    if (state.minutes) return null;
    if (!state.endsAt) return t('code.END_REQUIRED');
    if (new Date(state.endsAt).getTime() <= Date.now() + 30000) return t('code.END_IN_PAST');
    return null;
  };
  return el;
}

function defaultReason(kind) {
  var out = {};
  LANGS.active.forEach(function (l) {
    var cat = MESSAGES[l.code];
    if (cat && cat['rs.reason.default.' + kind]) out[l.code] = cat['rs.reason.default.' + kind];
  });
  return out;
}

function restrictionForm(kind, opts) {
  opts = opts || {};
  var v = { reason: defaultReason(kind), note: '', deleteMessage: false };
  var fields = {};
  var box = h('div', { 'data-restrict-form': kind });
  var dur = null;
  if (kind === 'suspend' || kind === 'mute') {
    dur = durationPicker(function () { fields.duration.setError(''); });
    fields.duration = field(t('rs.duration'), dur, t(kind === 'suspend' ? 'rs.duration.suspend.hint' : 'rs.duration.mute.hint'));
    box.appendChild(fields.duration);
  }
  var reason = translatedInput(v.reason, { label: t('rs.reason'), multiline: true, rows: 2, max: 500, name: 'reason', oninput: function (x) { v.reason = x; fields.reason.setError(''); } });
  fields.reason = field(t('rs.reason'), reason, t('rs.reason.hint.' + kind));
  box.appendChild(fields.reason);
  if (opts.withDelete) {
    box.appendChild(h('div', { class: 'mb-5' }, checkbox(false, t('rp.alsoDelete'), function (on) { v.deleteMessage = on; })));
  }
  var note = textInput('', { multiline: true, rows: 2, label: t('rs.note'), placeholder: t('rs.note.ph'), oninput: function (x) { v.note = x; } });
  fields.note = field(t('rs.note'), note, t('rs.note.hint'), { optional: true });
  box.appendChild(fields.note);
  return {
    el: box,
    validate: function () {
      var ok = true;
      if (dur) {
        var bad = dur.check();
        fields.duration.setError(bad || '');
        if (bad) ok = false;
      }
      var hit = trCheck(v.reason, { required: true, max: 500 });
      if (hit) {
        fields.reason.setError(t('code.' + hit.code, { lang: langName(hit.lang) }));
        reason.markLang(hit.lang);
        ok = false;
      }
      return ok;
    },
    payload: function () {
      var p = { kind: kind, reason: trClean(v.reason), note: v.note.trim() };
      if (dur) Object.assign(p, dur.value());
      if (opts.withDelete) p.deleteMessage = v.deleteMessage;
      return p;
    },
    showError: function (err) {
      var f = err && err.field;
      if (f === 'endsAt' && fields.duration) { fields.duration.setError(errorText(err)); return true; }
      if (f === 'reason') { fields.reason.setError(errorText(err)); return true; }
      if (f === 'note') { fields.note.setError(errorText(err)); return true; }
      return false;
    }
  };
}

function formDialog(opts) {
  return new Promise(function (resolve) {
    var done = false;
    var top = h('div', {});
    function finish(v) {
      if (done) return;
      done = true;
      root.remove();
      document.removeEventListener('keydown', esc, true);
      resolve(v);
    }
    function esc(e) {
      if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); finish(null); return; }
      if (e.key === 'Tab') trapTab(e, box);
    }
    var ok = button({ variant: opts.danger ? 'dangerSolid' : 'primary', size: 'footer', label: opts.ok, onclick: function () { submit(); } });
    ok.setAttribute('data-dialog-ok', '');
    function submit() {
      clear(top);
      if (opts.validate && !opts.validate()) return;
      setBusy(ok, true, t('saving'));
      opts.submit().then(function (res) { finish(res || true); }, function (err) {
        setBusy(ok, false);
        if (!opts.showError || !opts.showError(err)) top.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
      });
    }
    var box = h('div', { class: 'relative w-full eco-dialog eco-dialog--form bg-white dark:bg-secondary-900 rounded-md shadow-lg overflow-hidden flex flex-col', role: 'dialog', 'aria-modal': 'true', 'aria-label': opts.title, 'data-form-dialog': opts.id || '' },
      h('div', { class: 'p-6 overflow-y-auto' },
        h('div', { class: 'flex gap-3 mb-4' },
          h('div', { class: 'shrink-0 ' + (opts.danger ? 'text-error-600 dark:text-error-400' : 'eco-selected') }, icon(opts.icon || 'alert', 'h-6 w-6')),
          h('div', { class: 'min-w-0' },
            h('h2', { class: 'text-lg font-semibold mb-1', text: opts.title }),
            opts.text ? h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: opts.text }) : null)),
        top, opts.body),
      h('div', { class: 'flex gap-3 justify-end bg-secondary-50 dark:bg-secondary-800 px-6 py-3' },
        button({ variant: 'ghost', size: 'footer', label: t('cancel'), onclick: function () { finish(null); } }), ok));
    var holder = h('div', { class: 'fixed z-[70] inset-0 flex items-center justify-center p-4' }, box);
    holder.addEventListener('mousedown', function (e) { if (e.target === holder) finish(null); });
    var root = h('div', { 'data-economy': 'dialog', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'fixed z-[70] inset-0 opacity-60 dark:opacity-70 dark:bg-black bg-secondary-900' }), holder);
    document.body.appendChild(root);
    document.addEventListener('keydown', esc, true);
    var first = box.querySelector('[data-duration] button, textarea, input');
    (first || ok).focus();
  });
}

function restrictDialog(player, kind) {
  return freshLanguages().then(function () {
    var form = restrictionForm(kind);
    var name = player.displayName || player.username || String(player.id).slice(0, 8);
    return formDialog({
      id: 'restrict-' + kind,
      title: t('rs.dialog.' + kind, { name: name }),
      text: t('rs.dialog.' + kind + '.text'),
      icon: kind === 'warn' ? 'alert' : 'ban',
      danger: kind === 'suspend',
      ok: t('rs.do.' + kind),
      body: form.el,
      validate: form.validate,
      showError: form.showError,
      submit: function () { return baseApi('POST', '/players/' + encodeURIComponent(player.id) + '/restrictions', form.payload()); }
    });
  }).then(function (res) {
    if (res) toast('success', t('rs.done.' + kind, { name: player.displayName || player.username || '' }));
    return res;
  });
}

function liftDialog(r) {
  var note = textInput('', { multiline: true, rows: 2, label: t('rs.liftNote'), placeholder: t('rs.note.ph') });
  return formDialog({
    id: 'lift',
    title: t('rs.lift.' + r.kind),
    text: t('rs.lift.text', { name: r.displayName || r.username || '' }),
    icon: 'check',
    ok: t('rs.lift.ok'),
    body: field(t('rs.liftNote'), note, null, { optional: true }),
    submit: function () { return baseApi('POST', '/restrictions/' + encodeURIComponent(r.id) + '/lift', { note: note.value.trim() }); }
  }).then(function (res) {
    if (res) toast('success', t('rs.lifted'));
    return res;
  });
}

function bulkTagBar(reload) {
  var tagIds = [];
  var count = h('span', { class: 'text-sm font-medium tabular-nums', 'aria-live': 'polite' });
  var picker = refMulti({ resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', value: [], label: t('pl.bulk.tags'), placeholder: t('pl.bulk.pick'), onchange: function (v) { tagIds = v; } });
  function run(op, btn) {
    var ids = Object.keys(playerSel.ids);
    if (!ids.length) return;
    if (!tagIds.length) { toast('error', t('code.TAGS_REQUIRED')); return; }
    setBusy(btn, true, t('saving'));
    baseApi('POST', '/players/tags', { userIds: ids, tagIds: tagIds, op: op }).then(function (res) {
      setBusy(btn, false);
      toast('success', t(op === 'add' ? 'pl.bulk.added' : 'pl.bulk.removed', { n: res.players, changed: res.changed }));
      reload();
    }, function (err) {
      setBusy(btn, false);
      toast('error', errorText(err));
    });
  }
  var add = button({ variant: 'primary', size: 'sm', icon: 'plus', label: t('pl.bulk.add'), onclick: function (e) { run('add', e.currentTarget); } });
  var remove = button({ variant: 'danger', size: 'sm', icon: 'trash', label: t('pl.bulk.remove'), onclick: function (e) { run('remove', e.currentTarget); } });
  var pageAll = button({ variant: 'link', size: 'sm', label: t('pl.bulk.page'), onclick: function () {
    playerSel.page.forEach(function (it) { playerSel.ids[it.id] = true; });
    paint();
    reload();
  } });
  var none = button({ variant: 'ghost', size: 'sm', label: t('pl.bulk.clear'), onclick: function () { playerSel.ids = {}; paint(); reload(); } });
  add.setAttribute('data-bulk', 'add');
  remove.setAttribute('data-bulk', 'remove');
  var el = h('div', { class: 'eco-bulk rounded-md border border-secondary-200 dark:border-secondary-700 p-3 mb-3', 'data-bulk-bar': '' },
    h('div', { class: 'flex flex-wrap items-center gap-2 mb-2' }, count, pageAll, none),
    h('div', { class: 'eco-bulk-row' }, h('div', { class: 'min-w-0 flex-1' }, picker), h('div', { class: 'flex flex-wrap gap-2' }, add, remove)));
  function paint() {
    var n = Object.keys(playerSel.ids).length;
    count.textContent = n ? t('pl.bulk.selected', { n: n }) : t('pl.bulk.none');
    add.disabled = remove.disabled = !n;
  }
  el.paint = paint;
  paint();
  return el;
}

function playerCheck(it) {
  var box = checkbox(!!playerSel.ids[it.id], '', function (on) {
    if (on) playerSel.ids[it.id] = true;
    else delete playerSel.ids[it.id];
    if (playerSel.bar) playerSel.bar.paint();
  });
  box.setAttribute('aria-label', t('pl.select', { name: it.displayName || it.username }));
  box.setAttribute('data-select', it.id);
  return box;
}

LISTS.players = {
  api: 'base',
  rowHref: function (it) { return '#/players/' + encodeURIComponent(it.id); },
  onItems: function (items) { playerSel.page = items; },
  filters: [
    { name: 'tagId', label: 'filter.tag', type: 'ref', resource: 'tags', api: baseApi, filter: { keyType: 'user' } },
    { name: 'status', label: 'filter.status', options: function () { return ['suspended', 'muted', 'banned', 'tagged'].map(function (s) { return { value: s, label: t('pl.status.' + s) }; }); } }
  ],
  columns: [
    { label: 'col.select', render: playerCheck, width: 48 },
    { label: 'col.player', render: function (it) {
      return h('div', { class: 'min-w-0' },
        h('div', { class: 'font-medium truncate', text: it.displayName || it.username }),
        h('div', { class: 'text-secondary-500 flex items-center gap-2 min-w-0' }, it.displayName ? h('span', { class: 'truncate', text: it.username }) : null, mono(String(it.id).slice(0, 8))));
    }, skeleton: '50%' },
    { label: 'col.tags', render: function (it) { return playerTags(it.tags, 4); }, skeleton: '40%' },
    { label: 'col.status', render: playerStatus, width: 220 },
    { label: 'col.created', render: function (it) { return dateCell(it.createTime); }, width: 170 }
  ]
};

function playersPage(container, query, ctx) {
  playerSel.ids = {};
  var list = listPage(container, 'players', query, ctx);
  var bar = bulkTagBar(function () { list.reload(); });
  playerSel.bar = bar;
  var table = container.querySelector('.eco-filters');
  var anchor = table ? table.closest('.flex.flex-col.items-start') : null;
  if (anchor && anchor.parentNode) anchor.parentNode.insertBefore(bar, anchor.nextSibling);
  return list;
}

function restrictionReason(it) {
  return h('div', { class: 'min-w-0' },
    h('div', { class: 'eco-clamp', text: i18nText(it.reason) || '' }),
    it.note ? h('div', { class: 'text-xs text-secondary-500 eco-clamp', text: it.note }) : null);
}

LISTS.restrictions = {
  api: 'base',
  noRow: true,
  filters: [
    { name: 'kind', label: 'filter.kind', options: function () { return RESTRICT_KINDS.map(function (k) { return { value: k, label: t('rs.kind.' + k) }; }); } },
    { name: 'status', label: 'filter.status', options: function () { return ['active', 'expired', 'lifted'].map(function (s) { return { value: s, label: t('rs.status.' + s) }; }); } },
    { name: 'userId', label: 'filter.playerId', type: 'text' }
  ],
  columns: [
    { label: 'col.player', render: function (it) { return h('div', { class: 'min-w-0' }, playerLink(it.userId, it.displayName || it.username, 'font-medium block'), mono(String(it.userId || '').slice(0, 8))); }, skeleton: '40%' },
    { label: 'col.kind', render: function (it) { return tag(t('rs.kind.' + it.kind), RESTRICT_COLOR[it.kind]); }, width: 110 },
    { label: 'col.status', render: function (it) {
      if (it.kind === 'warn') return h('span', { class: 'text-secondary-400', text: t('none') });
      return h('div', { class: 'flex flex-col items-start gap-0.5' },
        tag(t('rs.status.' + it.status), RESTRICT_STATUS_COLOR[it.status]),
        h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: it.status === 'lifted' && it.liftedAt ? t('rs.liftedAt', { date: fmtDate(it.liftedAt, true) }) : untilText(it.endsAt) }));
    }, width: 190 },
    { label: 'col.reason', render: restrictionReason, skeleton: '60%' },
    { label: 'col.by', render: function (it) { return h('div', { class: 'flex flex-col' }, h('span', { class: 'text-sm', text: it.createdBy || t('none') }), dateCell(it.createdAt)); }, width: 170 },
    { label: 'col.actions', render: function (it) {
      if (it.kind === 'warn' || it.status !== 'active') return h('span', {});
      var b = button({ variant: 'link', size: 'sm', icon: 'check', label: t('rs.lift'), onclick: function () {
        liftDialog(it).then(function (res) { if (res) psOrPageReload('restrictions'); });
      } });
      b.setAttribute('data-lift', it.id);
      return b;
    }, width: 140 }
  ]
};

var listReloads = {};

function psOrPageReload(name) {
  if (listReloads[name]) listReloads[name]();
}

function restrictionsPage(container, query, ctx) {
  var list = listPage(container, 'restrictions', query, ctx);
  listReloads.restrictions = list.reload;
  return list;
}

function reportMessage(it) {
  return h('div', { class: 'min-w-0' },
    h('div', { class: 'eco-clamp', dir: 'auto', text: it.content || '' }),
    h('div', { class: 'text-xs text-secondary-500 flex items-center gap-1.5 min-w-0' }, icon('users', 'h-3 w-3'), h('span', { class: 'truncate', text: it.senderName || it.senderUsername || t('rp.gone') }), it.messageDeleted ? tag(t('rp.deleted'), 'secondary') : null));
}

function reportReason(it) {
  if (it.source === 'filter') {
    return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('rp.source.filter'), 'warning'), (it.matched || []).slice(0, 3).map(function (m) { return tag(m.word, 'secondary'); }));
  }
  return h('div', { class: 'flex flex-wrap gap-1' }, tag(i18nText(it.title) || it.titleKey || t('rp.noTitle'), 'primary'));
}

LISTS.reports = {
  api: 'base',
  filters: [
    { name: 'status', label: 'filter.status', options: function () { return ['open', 'dismissed', 'resolved'].map(function (s) { return { value: s, label: t('rp.status.' + s) }; }); } },
    { name: 'source', label: 'filter.source', options: function () { return ['player', 'filter'].map(function (s) { return { value: s, label: t('rp.source.' + s) }; }); } },
    { name: 'titleId', label: 'filter.reportTitle', type: 'ref', resource: 'report-titles', api: baseApi },
    { name: 'senderId', label: 'filter.playerId', type: 'text' }
  ],
  columns: [
    { label: 'col.message', render: reportMessage, skeleton: '70%' },
    { label: 'col.reason', render: reportReason, width: 200 },
    { label: 'col.reporter', render: function (it) { return it.source === 'filter' ? h('span', { class: 'text-secondary-500', text: t('rp.byFilter') }) : playerLink(it.reporterId, it.reporterName || it.reporterUsername); }, width: 160 },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-col items-start gap-0.5' }, tag(t('rp.status.' + it.status), REPORT_STATUS_COLOR[it.status]), it.reports > 1 ? h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: t('rp.reportsN', { n: it.reports }) }) : null); }, width: 130 },
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 170 }
  ]
};

function contextView(lines, reportedId) {
  var box = h('ol', { class: 'eco-chat flex flex-col gap-2', 'data-report-context': '' });
  (lines || []).forEach(function (l) {
    var hit = l.reported || l.id === reportedId;
    box.appendChild(h('li', { class: 'eco-chat-line' + (hit ? ' eco-chat-hit' : ''), 'data-reported': hit ? 'true' : null },
      h('div', { class: 'flex items-center gap-2 text-xs text-secondary-500 mb-0.5' },
        playerLink(l.senderId, l.username, 'font-medium'),
        h('span', { class: 'whitespace-nowrap', text: fmtDate(l.createdAt, true) }),
        hit ? tag(t('rp.thisMessage'), 'warning') : null),
      h('div', { class: 'text-sm text-secondary-800 dark:text-secondary-200 break-words', dir: 'auto', text: l.content })));
  });
  if (!box.childNodes.length) box.appendChild(h('li', { class: 'text-sm text-secondary-500', text: t('rp.noContext') }));
  return box;
}

function reportDetail(container, id) {
  var backBtn = button({ variant: 'ghost', size: 'sm', icon: 'back', label: t('back'), onclick: function () { go('reports', null, lastQuery.reports || {}); } });
  var frame = pageFrame(t('rp.title'), null, null);
  frame.actions.appendChild(backBtn);
  clear(container).appendChild(frame.el);
  var grid = h('div', { class: 'eco-detail-grid', 'data-report': id });
  frame.content.appendChild(grid);

  function skeleton() {
    clear(grid);
    [6, 5, 5, 4].forEach(function (n, i) {
      var rows = [];
      for (var j = 0; j < n; j++) rows.push(h('div', { class: 'py-1.5' }, bar(['70%', '50%', '40%'][j % 3])));
      grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), rows), i === 0 ? 'eco-span-all' : null));
    });
  }

  function decision(r) {
    if (r.status !== 'open') {
      return section('rp.decision', dl([
        ['col.status', tag(t('rp.status.' + r.status), REPORT_STATUS_COLOR[r.status])],
        ['rp.action', r.action ? tag(t('rp.action.' + r.action), r.action === 'dismiss' ? 'secondary' : 'primary') : null],
        ['rp.messageDeleted', r.messageDeleted ? tag(t('yes'), 'secondary') : null],
        ['rp.restriction', r.restrictionId ? h('a', { href: hashFor('restrictions', null, { userId: r.senderId }), class: 'eco-link', text: t('rp.seeRestriction') }) : null],
        ['rp.resolvedBy', r.resolvedBy ? h('span', { text: r.resolvedBy }) : null],
        ['rp.resolvedAt', r.resolvedAt ? dateCell(r.resolvedAt) : null],
        ['rs.note', r.resolutionNote ? h('span', { class: 'break-words', text: r.resolutionNote }) : null]
      ]));
    }
    var current = 'dismiss';
    var formBox = h('div', { class: 'mt-4' });
    var form = null;
    var note = null;
    var seg = segmented(REPORT_ACTIONS.map(function (a) { return { value: a, label: t('rp.action.' + a) }; }), current, function (a) { current = a; paintForm(); }, t('rp.action'), 'md');
    seg.setAttribute('data-report-actions', '');
    var hint = h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mt-2' });
    var top = h('div', {});
    var submitBtn = button({ variant: 'primary', size: 'footer', label: t('rp.submit'), onclick: function () { submit(); } });
    submitBtn.setAttribute('data-report-submit', '');
    function paintForm() {
      clear(formBox);
      clear(top);
      hint.textContent = t('rp.action.' + current + '.hint');
      form = null;
      note = null;
      if (current === 'warn' || current === 'mute' || current === 'suspend') {
        form = restrictionForm(current, { withDelete: true });
        formBox.appendChild(form.el);
      } else {
        note = textInput('', { multiline: true, rows: 2, label: t('rs.note'), placeholder: t('rs.note.ph') });
        formBox.appendChild(field(t('rs.note'), note, t('rp.note.hint'), { optional: true }));
      }
      submitBtn.querySelector('span').textContent = t('rp.submit.' + current);
    }
    function submit() {
      clear(top);
      if (form && !form.validate()) return;
      var body = form ? form.payload() : { note: note.value.trim() };
      body.action = current;
      delete body.kind;
      setBusy(submitBtn, true, t('saving'));
      baseApi('POST', '/reports/' + encodeURIComponent(id) + '/resolve', body).then(function () {
        toast('success', t('rp.done.' + current));
        refreshReportBadge(true);
        load();
      }, function (err) {
        setBusy(submitBtn, false);
        if (!form || !form.showError(err)) top.appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
      });
    }
    paintForm();
    return card(h('div', { class: 'p-4', 'data-report-decide': '' },
      h('h2', { class: 'text-lg font-medium mb-3', text: t('rp.decide') }),
      top, seg, hint, formBox,
      h('div', { class: 'flex justify-end' }, submitBtn)), 'eco-span-all');
  }

  function render(r) {
    clear(grid);
    var head = h('div', { class: 'flex flex-wrap items-center gap-2' },
      tag(t('rp.status.' + r.status), REPORT_STATUS_COLOR[r.status]),
      r.source === 'filter' ? tag(t('rp.source.filter'), 'warning') : tag(t('rp.source.player'), 'secondary'),
      r.messageDeleted ? tag(t('rp.deleted'), 'secondary') : null);
    grid.appendChild(section('rp.conversation', h('div', {}, h('div', { class: 'mb-3' }, head), contextView(r.context, r.messageId)), null));
    grid.lastChild.classList.add('eco-span-all');
    var others = (r.others || []).filter(function (o) { return o.source === 'player'; });
    grid.appendChild(section('rp.report', dl([
      ['rp.reason', reportReason(r)],
      ['col.reporter', r.source === 'filter' ? h('span', { text: t('rp.byFilter') }) : playerLink(r.reporterId, r.reporterName || r.reporterUsername)],
      ['rp.playerNote', r.note ? h('span', { class: 'break-words', dir: 'auto', text: r.note }) : null],
      ['rp.matched', (r.matched || []).length ? h('div', { class: 'flex flex-wrap gap-1' }, r.matched.map(function (m) { return tag(m.word, WORD_ACTION_COLOR[m.action] || 'secondary'); })) : null],
      ['rp.others', others.length ? h('div', { class: 'flex flex-col gap-1' }, others.map(function (o) { return h('div', { class: 'text-sm' }, playerLink(o.reporterId, o.reporterName || o.reporterUsername), h('span', { class: 'text-secondary-500', text: ' · ' + (i18nText(o.title) || o.titleKey || '') })); })) : null],
      ['col.created', dateCell(r.createdAt)],
      ['rp.channel', r.channelLabel ? mono(r.channelLabel) : null]
    ])));
    var s = r.sender;
    grid.appendChild(section('rp.sender', s ? dl([
      ['col.player', playerLink(s.id, s.displayName || s.username, 'font-medium')],
      ['col.tags', playerTags(s.tags)],
      ['col.status', playerStatus(s)],
      ['rp.history', h('div', { class: 'flex flex-wrap gap-1' },
        tag(t('rp.count.reports', { n: s.counts.reports }), 'secondary'),
        tag(t('rp.count.warn', { n: s.counts.warn }), 'primary'),
        tag(t('rp.count.mute', { n: s.counts.mute }), 'warning'),
        tag(t('rp.count.suspend', { n: s.counts.suspend }), 'error'))],
      ['rp.links', h('div', { class: 'flex flex-wrap gap-3' },
        h('a', { href: hashFor('restrictions', null, { userId: s.id }), class: 'eco-link text-sm', text: t('rp.seeHistory') }),
        h('a', { href: hashFor('reports', null, { senderId: s.id }), class: 'eco-link text-sm', text: t('rp.seeReports') }))]
    ]) : notice('secondary', t('rp.gone'))));
    grid.appendChild(decision(r));
  }

  function load() {
    skeleton();
    baseApi('GET', '/reports/' + encodeURIComponent(id)).then(function (r) {
      if (!grid.isConnected) return;
      Promise.resolve(freshLanguages()).then(function () { render(r); });
    }, function (err) {
      clear(grid).appendChild(errorBlock(err, load));
    });
  }
  load();
}

function wordTry() {
  var input = textInput('', { multiline: true, rows: 2, label: t('cw.try.label'), placeholder: t('cw.try.ph') });
  input.setAttribute('data-word-try', '');
  var result = h('div', { class: 'flex flex-col gap-2 mt-3 text-sm', 'aria-live': 'polite', 'data-word-result': '' });
  var seq = 0;
  var run = debounce(function () {
    var text = input.value;
    var my = ++seq;
    clear(result);
    if (!text.trim()) return;
    result.appendChild(bar('60%'));
    baseApi('POST', '/chat-words/test', { text: text }).then(function (res) {
      if (my !== seq) return;
      clear(result);
      var verdict = res.blocked ? notice('error', t('cw.try.blocked')) : res.flagged ? notice('warning', res.changed ? t('cw.try.maskedFlagged') : t('cw.try.flagged')) : res.changed ? notice('success', t('cw.try.masked')) : notice('secondary', t('cw.try.clean'));
      result.appendChild(verdict);
      result.appendChild(h('div', { class: 'eco-try-out rounded-md px-3 py-2' },
        h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('cw.try.shown') }),
        h('div', { class: 'break-words', dir: 'auto', 'data-word-shown': '', text: res.blocked ? t('cw.try.notSent') : res.text })));
      result.appendChild(h('div', { class: 'text-xs text-secondary-500' }, h('span', { text: t('cw.try.normalized') + ' ' }), h('span', { class: 'font-mono', dir: 'auto', 'data-word-normalized': '', text: res.normalized || '' })));
      if ((res.hits || []).length) result.appendChild(h('div', { class: 'flex flex-wrap gap-1' }, res.hits.map(function (hh) { return tag(hh.word + ' · ' + t('cw.action.' + hh.action), WORD_ACTION_COLOR[hh.action]); })));
    }, function (err) {
      if (my !== seq) return;
      clear(result).appendChild(notice('error', errorText(err)));
    });
  }, 300);
  input.addEventListener('input', run);
  return card(h('div', { class: 'p-4' },
    h('h2', { class: 'text-base font-medium mb-1', text: t('cw.try') }),
    h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mb-3', text: t('cw.try.hint') }),
    input, result), 'mb-4 eco-word-try');
}

RESOURCES['chat-words'] = {
  api: 'base',
  filters: [
    { name: 'action', label: 'filter.action', options: function () { return ['mask', 'block', 'flag'].map(function (a) { return { value: a, label: t('cw.action.' + a) }; }); } },
    { name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }
  ],
  columns: [
    { label: 'col.word', render: function (it) {
      return h('div', { class: 'min-w-0' },
        h('div', { class: 'font-medium truncate', dir: 'auto', text: it.word }),
        it.normalized && it.normalized !== it.word ? h('div', { class: 'text-xs text-secondary-500 font-mono truncate', dir: 'auto', text: it.normalized }) : null);
    }, skeleton: '40%' },
    { label: 'col.action', render: function (it) { return tag(t('cw.action.' + it.action), WORD_ACTION_COLOR[it.action]); }, width: 140 },
    { label: 'col.match', render: function (it) { return h('span', { class: 'text-sm', text: t('cw.match.' + it.match) }); }, width: 160 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 },
    { label: 'col.updated', render: function (it) { return dateCell(it.updatedAt); }, width: 170 }
  ],
  fields: [
    { name: 'word', type: 'text', label: 'f.cw.word', hint: 'f.cw.word.hint', required: true, check: function (v) { return v && Array.from(v.trim()).length > 64 ? t('code.INVALID_WORD') : null; } },
    { type: 'row', fields: [
      { name: 'action', type: 'select', label: 'f.cw.action', hint: 'f.cw.action.hint', required: true, options: function () { return ['mask', 'block', 'flag'].map(function (a) { return { value: a, label: t('cw.action.' + a) }; }); } },
      { name: 'match', type: 'select', label: 'f.cw.match', hint: 'f.cw.match.hint', required: true, options: function () { return ['word', 'contains'].map(function (a) { return { value: a, label: t('cw.match.' + a) }; }); } }
    ] },
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.cw.active.hint' },
    { name: 'note', type: 'textarea', label: 'f.cw.note', hint: 'f.cw.note.hint', optional: true }
  ],
  defaults: function () { return { word: '', action: 'mask', match: 'word', isActive: true, note: '' }; },
  fixPayload: function (p) { p.word = String(p.word || '').trim(); p.note = String(p.note || '').trim(); return p; }
};

function chatWordsPage(container, query, ctx) {
  var list = listPage(container, 'chat-words', query, ctx);
  var frameContent = container.querySelector('.flex.flex-col.flex-1.min-w-0.px-4');
  if (frameContent) frameContent.insertBefore(wordTry(), frameContent.firstChild);
  return list;
}

RESOURCES['report-titles'] = {
  api: 'base',
  settings: true,
  filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
  columns: [
    { label: 'col.name', render: function (it) { return nameCell(it, false); }, skeleton: '50%' },
    { label: 'col.used', render: function (it) { return it.used ? h('a', { href: hashFor('reports', null, { titleId: it.id }), class: 'eco-link tabular-nums', text: t('rp.reportsN', { n: it.used }) }) : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 140 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 },
    { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.rt.key.hint', required: true },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
    ] },
    tr('name', { label: 'f.rt.name', hint: 'f.rt.name.hint', required: true, max: 100 }),
    tr('description', { label: 'f.rt.description', hint: 'f.rt.description.hint', multiline: true, max: 300, optional: true }),
    { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.rt.active.hint' }
  ],
  defaults: function () { return { key: '', name: {}, description: {}, isActive: true, sort: 100 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 100; return p; }
};

var PLAYER_CARD = { id: null, data: null, loading: false, at: 0, err: null };

function playerCardRoute() {
  var m = /^#\/players\/([0-9a-f-]{36})(?:\/account)?(?:\?.*)?$/.exec(location.hash || '');
  return m ? m[1] : null;
}

function loadPlayerCard(id) {
  PLAYER_CARD.loading = true;
  return baseApi('GET', '/players/' + encodeURIComponent(id)).then(function (res) {
    PLAYER_CARD.loading = false;
    PLAYER_CARD.at = Date.now();
    PLAYER_CARD.data = res;
    PLAYER_CARD.err = null;
    var old = document.querySelector('[data-player-mod]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    PLAYER_CARD.loading = false;
    PLAYER_CARD.at = Date.now();
    PLAYER_CARD.err = err;
    schedule();
  });
}

function editPlayerTags(p) {
  var ids = (p.tags || []).map(function (x) { return x.id; });
  (p.tags || []).forEach(function (x) { refCache['tags:id:' + x.id] = x; });
  var picker = refMulti({ resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', value: ids, label: t('pc.tags'), placeholder: t('pl.bulk.pick'), onchange: function (v) { ids = v; } });
  return formDialog({
    id: 'player-tags',
    title: t('pc.tags.edit'),
    text: t('pc.tags.edit.text'),
    icon: 'tags',
    ok: t('save'),
    body: field(t('pc.tags'), picker, t('pc.tags.hint')),
    submit: function () { return baseApi('PUT', '/players/' + encodeURIComponent(p.id) + '/tags', { tagIds: ids }); }
  }).then(function (res) {
    if (res) toast('success', t('pc.tags.saved'));
    return res;
  });
}

function playerModCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-mod]');
  if (!id || !token() || !aclCan('ACCOUNT', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (PLAYER_CARD.id !== id) { PLAYER_CARD.id = id; PLAYER_CARD.data = null; PLAYER_CARD.err = null; PLAYER_CARD.at = 0; }
  if (!PLAYER_CARD.loading && Date.now() - PLAYER_CARD.at > 20000) loadPlayerCard(id);
  var p = PLAYER_CARD.data;
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify([p.tags, p.suspension, p.mute, p.banned, (p.history || []).length]) : PLAYER_CARD.err ? 'err' : 'loading');
  if (existing && existing.getAttribute('data-key') === key && existing.parentNode === grid && grid.firstChild === existing) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-mod': '', 'data-key': key, 'data-economy': 'player-mod', 'data-l10n-skip': '', 'aria-labelledby': 'player-mod-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-mod-title', class: 'font-medium', text: t('pc.title') });
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(PLAYER_CARD.err ? notice('error', errorText(PLAYER_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { PLAYER_CARD.at = 0; loadPlayerCard(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    grid.insertBefore(cardEl, grid.firstChild);
    return;
  }
  var write = aclCan('ACCOUNT', 1);
  var after = function (res) { if (res) { PLAYER_CARD.at = 0; loadPlayerCard(id); } };
  var actions = h('div', { class: 'flex flex-wrap items-center gap-2' });
  if (write) {
    if (p.suspension) actions.appendChild(liftButton(p.suspension.id, 'suspend', p, after));
    else actions.appendChild(actionButton('suspend', 'dangerSolid', function () { restrictDialog(p, 'suspend').then(after); }));
    if (p.mute) actions.appendChild(liftButton(p.mute.id, 'mute', p, after));
    else actions.appendChild(actionButton('mute', 'ghost', function () { restrictDialog(p, 'mute').then(after); }));
    actions.appendChild(actionButton('warn', 'ghost', function () { restrictDialog(p, 'warn').then(after); }));
  }
  var editTags = write ? button({ variant: 'link', size: 'sm', icon: 'tags', label: t('pc.tags.edit'), onclick: function () { editPlayerTags(p).then(after); } }) : null;
  if (editTags) editTags.setAttribute('data-player-tags-edit', '');
  var lastWarn = (p.history || []).filter(function (x) { return x.kind === 'warn'; })[0];
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-3' }, title, actions));
  body.appendChild(h('div', { class: 'eco-pc-grid' },
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('pc.tags') }),
      h('div', { class: 'flex flex-wrap items-center gap-2' }, playerTags(p.tags), editTags)),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('pc.status') }),
      playerStatus(p),
      p.suspension && i18nText(p.suspension.reason) ? h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300 mt-1 eco-clamp', text: i18nText(p.suspension.reason) }) : null,
      p.mute && i18nText(p.mute.reason) ? h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300 mt-1 eco-clamp', text: i18nText(p.mute.reason) }) : null),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('pc.history') }),
      h('div', { class: 'flex flex-wrap gap-3 text-sm' },
        h('a', { href: hashFor('restrictions', null, { userId: id }), class: 'eco-link', 'data-player-history': '', text: t('pc.historyN', { n: (p.history || []).length }) }),
        h('a', { href: hashFor('reports', null, { senderId: id }), class: 'eco-link', text: t('pc.reportsN', { n: p.reports ? p.reports.against : 0 }) })),
      lastWarn ? h('p', { class: 'text-xs text-secondary-500 mt-1', text: t('pc.lastWarn', { date: fmtDate(lastWarn.createdAt, true) }) }) : null)));
  grid.insertBefore(cardEl, grid.firstChild);
}

function actionButton(kind, variant, onclick) {
  var b = button({ variant: variant, size: 'sm', icon: kind === 'warn' ? 'alert' : 'ban', label: t('rs.do.' + kind), onclick: onclick });
  b.setAttribute('data-restrict', kind);
  return b;
}

function liftButton(rid, kind, p, after) {
  var b = button({ variant: 'link', size: 'sm', icon: 'check', label: t('rs.lift.' + kind), onclick: function () {
    liftDialog({ id: rid, kind: kind, username: p.username, displayName: p.displayName }).then(after);
  } });
  b.setAttribute('data-lift-kind', kind);
  return b;
}
