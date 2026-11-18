ICONS['handoff'] = '<path d="M12 3v12"/><path d="M7 10l5 5 5-5"/><path d="M5 21h14"/>';
PH_ICONS['handoff'] = '<path d="M224,144v64a8,8,0,0,1-8,8H40a8,8,0,0,1-8-8V144a8,8,0,0,1,16,0v56H208V144a8,8,0,0,1,16,0Zm-101.66,5.66a8,8,0,0,0,11.32,0l40-40a8,8,0,0,0-11.32-11.32L136,124.69V32a8,8,0,0,0-16,0v92.69L93.66,98.34a8,8,0,0,0-11.32,11.32Z"/>';

var HANDOFF = { meta: null, at: 0, loading: null, panel: null, listeners: [] };

function handoffNotify() {
  HANDOFF.listeners.forEach(function (fn) { try { fn(HANDOFF.meta); } catch (e) { return; } });
}

function handoffLoad(force) {
  if (!token()) { HANDOFF.meta = null; handoffNotify(); return Promise.resolve(null); }
  if (!force && HANDOFF.meta && Date.now() - HANDOFF.at < 15000) return Promise.resolve(HANDOFF.meta);
  if (HANDOFF.loading) return HANDOFF.loading;
  HANDOFF.loading = baseApi('GET', '/handoff').then(function (m) {
    HANDOFF.meta = m || null;
    HANDOFF.at = Date.now();
    HANDOFF.loading = null;
    handoffNotify();
    return HANDOFF.meta;
  }, function (err) {
    HANDOFF.loading = null;
    if (err && (err.status === 401 || err.status === 403 || err.status === 404)) { HANDOFF.meta = null; handoffNotify(); }
    throw err;
  });
  return HANDOFF.loading;
}

function handoffTitle(entry) {
  if (!entry) return '';
  return lang() === 'fa' ? (entry.title || entry.titleEn || '') : (entry.titleEn || entry.title || '');
}

function handoffDate(d) {
  return fmtDate(String(d) + 'T12:00:00Z', false);
}

function handoffDownload(btn) {
  setBusy(btn, true, t('hf.downloading'));
  var headers = { Accept: 'text/html' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch(BASE_BASE + '/handoff/download', { headers: headers, credentials: 'same-origin' }).then(function (res) {
    if (res.status === 401) { toLogin(); throw new ApiError(401, 'UNAUTHENTICATED', ''); }
    if (!res.ok) {
      return res.json().catch(function () { return {}; }).then(function (d) { throw new ApiError(res.status, d && d.error, d && d.message); });
    }
    var cd = res.headers.get('Content-Disposition') || '';
    var m = /filename="([^"]+)"/.exec(cd);
    var name = m ? m[1] : ((HANDOFF.meta && HANDOFF.meta.fileName) || 'handoff.html');
    return res.blob().then(function (blob) {
      var url = URL.createObjectURL(blob);
      var a = h('a', { href: url, download: name, class: 'hidden', 'data-handoff-file': name });
      document.body.appendChild(a);
      a.click();
      setTimeout(function () { URL.revokeObjectURL(url); a.remove(); }, 4000);
      toast('success', t('hf.downloaded'));
    });
  }, function () { throw new ApiError(0, 'NETWORK', ''); }).catch(function (err) {
    toast('error', errorText(err));
  }).then(function () { setBusy(btn, false); });
}

function handoffPart(label, title, count, scope) {
  return h('div', { class: 'eco-hf-part', 'data-handoff-part': scope },
    h('span', { class: 'eco-hf-part-label', text: label }),
    h('span', { class: 'eco-hf-part-title', text: title }),
    h('span', { class: 'eco-hf-part-n', text: t('hf.sections', { n: count }) }));
}

function handoffChanges(meta) {
  var all = meta.changelog || [];
  var hasProject = !!meta.project;
  var filter = 'all';
  var list = h('ol', { class: 'eco-hf-list', 'data-handoff-changes': '', 'aria-live': 'polite' });
  function paint() {
    clear(list);
    var rows = all.filter(function (e) { return filter === 'all' || e.scope === filter; });
    if (!rows.length) { list.appendChild(h('li', { class: 'py-6' }, emptyBlock(t('hf.empty')))); return; }
    rows.forEach(function (e) {
      list.appendChild(h('li', { class: 'eco-hf-row', 'data-scope': e.scope },
        h('time', { class: 'eco-hf-date', datetime: e.date, text: handoffDate(e.date) }),
        h('span', { class: 'eco-hf-text', text: handoffTitle(e) }),
        tag(t(e.scope === 'project' ? 'hf.filter.project' : 'hf.filter.general'), e.scope === 'project' ? 'warning' : 'secondary')));
    });
  }
  var opts = [{ value: 'all', label: t('hf.filter.all') }, { value: 'general', label: t('hf.filter.general') }];
  if (hasProject) opts.push({ value: 'project', label: t('hf.filter.project') });
  var seg = segmented(opts, filter, function (v) { filter = v; paint(); }, t('hf.changes'));
  seg.setAttribute('data-handoff-filter', '');
  paint();
  return h('section', { class: 'eco-hf-changes', 'aria-labelledby': 'eco-hf-changes' },
    h('div', { class: 'eco-hf-head' },
      h('div', {},
        h('h3', { id: 'eco-hf-changes', class: 'eco-hf-h', text: t('hf.changes') }),
        h('p', { class: 'eco-hf-muted', text: t('hf.changes.hint') })),
      seg),
    list);
}

function handoffBody(body, meta, dl) {
  clear(body);
  var counts = { general: 0, project: 0 };
  (meta.sections || []).forEach(function (s) { counts[s.scope] = (counts[s.scope] || 0) + 1; });
  var parts = h('div', { class: 'eco-hf-parts' },
    handoffPart(t('hf.forAll'), handoffTitle(meta.general) || t('hf.general'), counts.general, 'general'),
    meta.project ? handoffPart(t('hf.forProject'), t('hf.project') + ': ' + handoffTitle(meta.project), counts.project, 'project') : null);
  body.appendChild(h('div', { class: 'eco-hf', 'data-handoff-panel': '' },
    h('p', { class: 'eco-hf-intro', text: t('hf.intro') }),
    parts,
    meta.project ? null : h('div', { 'data-handoff-no-project': '' }, notice('info', t('hf.noProject'))),
    h('div', { class: 'eco-hf-get' },
      h('div', { class: 'eco-hf-get-row' }, dl, h('span', { class: 'eco-hf-file' }, mono(meta.fileName || ''))),
      h('p', { class: 'eco-hf-muted', text: t('hf.keys') })),
    handoffChanges(meta)));
}

function handoffOpen() {
  if (HANDOFF.panel) return;
  var body = h('div', { class: 'flex flex-col gap-3' }, bar('40%'), bar('70%'), bar('90%'), bar('60%'));
  var dl = button({ variant: 'primary', size: 'footer', icon: 'handoff', label: t('hf.download'), onclick: function () { handoffDownload(dl); } });
  dl.setAttribute('data-handoff-download', '');
  var opener = document.activeElement;
  function close() {
    if (!HANDOFF.panel) return;
    document.removeEventListener('keydown', esc, true);
    HANDOFF.panel.remove();
    HANDOFF.panel = null;
    if (opener && opener.focus) opener.focus();
  }
  function esc(e) { if (e.key === 'Escape') { e.stopPropagation(); close(); } }
  var closeBtn = button({ variant: 'ghost', size: 'footer', label: t('close'), onclick: close });
  HANDOFF.panel = sidePanel(t('hf.title'), body, closeBtn, close);
  HANDOFF.panel.setAttribute('data-handoff', '');
  document.body.appendChild(HANDOFF.panel);
  document.addEventListener('keydown', esc, true);
  function show(meta) {
    if (!HANDOFF.panel) return;
    if (!meta || !meta.visible) {
      clear(body).appendChild(notice('warning', t('code.' + ((meta && meta.reason) || 'HANDOFF_OFF'))));
      return;
    }
    handoffBody(body, meta, dl);
    dl.focus();
  }
  handoffLoad(true).then(show, function (err) {
    if (!HANDOFF.panel) return;
    clear(body).appendChild(notice('error', t('hf.loadFailed') + ' ' + errorText(err), button({ variant: 'link', size: 'sm', label: t('hf.retry'), onclick: function () { close(); handoffOpen(); } })));
  });
}

window.addEventListener('hashchange', function () { handoffLoad(false).catch(function () { return null; }); });
window.addEventListener('nakama-console:ready', function () { handoffLoad(false).catch(function () { return null; }); });
