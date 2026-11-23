ICONS['config-transfer'] = '<path d="M7 7h13M16 3l4 4-4 4"/><path d="M17 17H4M8 13l-4 4 4 4"/>';
PH_ICONS['config-transfer'] = '<path d="M213.66,181.66l-32,32a8,8,0,0,1-11.32-11.32L188.69,184H48a8,8,0,0,1,0-16H188.69l-18.35-18.34a8,8,0,0,1,11.32-11.32l32,32A8,8,0,0,1,213.66,181.66Zm-139.32-64a8,8,0,0,0,11.32-11.32L67.31,88H208a8,8,0,0,0,0-16H67.31L85.66,53.66A8,8,0,0,0,74.34,42.34l-32,32a8,8,0,0,0,0,11.32Z"/>';
if (!PH_ICONS.download) PH_ICONS.download = '<path d="M224,144v64a8,8,0,0,1-8,8H40a8,8,0,0,1-8-8V144a8,8,0,0,1,16,0v56H208V144a8,8,0,0,1,16,0Zm-101.66,5.66a8,8,0,0,0,11.32,0l40-40a8,8,0,0,0-11.32-11.32L136,124.69V32a8,8,0,0,0-16,0v92.69L93.66,98.34a8,8,0,0,0-11.32,11.32Z"/>';
if (!ICONS.download) ICONS.download = '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3"/>';

var CT_PATH = '/config-transfer';
var CT_STATUSES = ['new', 'changed', 'same', 'missing_refs', 'secret_needed', 'switched_off', 'skipped', 'error'];
var CT_TONE = { new: 'success', changed: 'warning', same: 'secondary', missing_refs: 'error', secret_needed: 'warning', switched_off: 'warning', skipped: 'secondary', error: 'error' };
var CT_PAGE = 50;

function ctIso(s) {
  return '\u2068' + s + '\u2069';
}

function ctTitle(a) {
  var tl = a && a.title;
  if (!tl) return a ? a.key : '';
  return tl[lang()] || tl.en || a.key;
}

function ctStatusTag(s) {
  return tag(t('ct.st.' + s), CT_TONE[s] || 'secondary', 'eco-cx-status');
}

function ctSize(bytes) {
  if (bytes < 1024 * 1024) return t('ct.kb', { n: Math.max(1, Math.round(bytes / 1024)) });
  return t('ct.mb', { n: Math.round(bytes / 104857.6) / 10 });
}

function ctValue(v) {
  if (v === null || v === undefined) return '-';
  if (typeof v === 'string') return v === '' ? '""' : v;
  var s = JSON.stringify(v);
  return s.length > 120 ? s.slice(0, 117) + '...' : s;
}

function ctWarn(w) {
  var i = w.indexOf(':');
  var code = i < 0 ? w : w.slice(0, i);
  var arg = i < 0 ? '' : w.slice(i + 1);
  return has('ct.warn.' + code) ? t('ct.warn.' + code, { x: arg }) : w;
}

function ctStat(name, value, label, tone) {
  return h('div', { class: 'eco-sc-stat', 'data-tone': tone, 'data-ct-stat': name },
    h('div', { class: 'eco-sc-num tabular-nums', text: fmtNum(value) }),
    h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: label }));
}

function ctSend(path, file, opts) {
  var form = new FormData();
  form.append('file', file, file.name || 'config.json');
  form.append('options', JSON.stringify(opts || {}));
  return baseApi('POST', CT_PATH + path, form, true);
}

function ctDownload(areas) {
  var headers = { Accept: 'application/json', 'Content-Type': 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch(BASE_BASE + CT_PATH + '/export', { method: 'POST', headers: headers, credentials: 'same-origin', body: JSON.stringify({ areas: areas }) }).then(function (res) {
    if (res.status === 401) { toLogin(); throw new ApiError(401, 'UNAUTHENTICATED', ''); }
    if (!res.ok) {
      return res.json().catch(function () { return {}; }).then(function (d) { throw new ApiError(res.status, d && d.error, d && d.message); });
    }
    var cd = res.headers.get('Content-Disposition') || '';
    var m = /filename="([^"]+)"/.exec(cd);
    var name = m ? m[1] : 'config.json';
    return res.blob().then(function (blob) {
      var url = URL.createObjectURL(blob);
      var a = h('a', { href: url, download: name, class: 'hidden', 'data-ct-file': name });
      document.body.appendChild(a);
      a.click();
      setTimeout(function () { URL.revokeObjectURL(url); a.remove(); }, 1000);
      return name;
    });
  }, function () { throw new ApiError(0, 'NETWORK', ''); });
}

function ctSettingsPanel(data, onsaved) {
  var st = Object.assign({}, data.defaults || {}, data.settings || {});
  var name = h('input', { type: 'text', class: 'x-input-field w-full', value: st.serverName || '', dir: 'auto', maxlength: 64, placeholder: data.serverName || '', 'data-ct-set': 'serverName' });
  var size = h('input', { type: 'number', class: 'x-input-field w-full', value: st.maxFileMB, min: 1, max: 200, dir: 'ltr', 'data-ct-set': 'maxFileMB' });
  var keep = h('input', { type: 'number', class: 'x-input-field w-full', value: st.keepRuns, min: 1, max: 500, dir: 'ltr', 'data-ct-set': 'keepRuns' });
  var exp = toggle(st.exportEnabled, t('ct.set.export'), function (v) { st.exportEnabled = v; });
  var imp = toggle(st.importEnabled, t('ct.set.import'), function (v) { st.importEnabled = v; });
  exp.setAttribute('data-ct-set', 'exportEnabled');
  imp.setAttribute('data-ct-set', 'importEnabled');
  var fName = field(t('ct.set.name'), name, t('ct.set.name.hint'), { optional: true });
  var fSize = field(t('ct.set.size'), size, t('ct.set.size.hint', { n: (data.defaults || {}).maxFileMB || 20 }));
  var fKeep = field(t('ct.set.keep'), keep, t('ct.set.keep.hint', { n: (data.defaults || {}).keepRuns || 50 }));
  var body = h('div', { class: 'flex flex-col', 'data-ct-settings': '' }, fName, fSize, fKeep,
    h('div', { class: 'flex flex-col gap-3 mb-5' }, exp, imp));
  var root;
  function close() { if (root) root.remove(); }
  var save = button({ variant: 'primary', size: 'footer', label: t('save'), onclick: function () {
    [fSize, fKeep].forEach(function (f) { f.setError(''); });
    var body2 = { serverName: name.value.trim(), maxFileMB: Number(toLatinDigits(size.value)), keepRuns: Number(toLatinDigits(keep.value)), exportEnabled: !!st.exportEnabled, importEnabled: !!st.importEnabled };
    setBusy(save, true);
    baseApi('PUT', CT_PATH + '/settings', body2).then(function (res) {
      toast('success', t('ct.set.saved'));
      close();
      onsaved(res);
    }, function (err) {
      setBusy(save, false, t('save'));
      var target = err.field === 'maxFileMB' ? fSize : err.field === 'keepRuns' ? fKeep : err.field === 'serverName' ? fName : null;
      if (target) target.setError(errorText(err));
      else toast('error', errorText(err));
    });
  } });
  save.setAttribute('data-ct-save', '');
  root = sidePanel(t('ct.set.title'), body, [button({ variant: 'ghost', size: 'footer', label: t('cancel'), onclick: close }), save], close);
  document.body.appendChild(root);
  name.focus();
}

function ctExportCard(data) {
  var picked = {};
  (data.areas || []).forEach(function (a) { picked[a.key] = true; });
  var boxes = [];
  var count = h('span', { class: 'text-xs text-secondary-500 tabular-nums', 'data-ct-export-count': '' });
  var allBox;
  var dl;
  function paint() {
    var n = Object.keys(picked).filter(function (k) { return picked[k]; }).length;
    count.textContent = t('ct.exp.picked', { n: n, all: (data.areas || []).length });
    allBox.set(n && n === boxes.length ? true : n ? 'mixed' : false);
    var off = !data.admin || !n || !data.settings.exportEnabled;
    dl.disabled = off;
    dl.className = dl.className.replace(/cursor-(pointer|not-allowed)/, off ? 'cursor-not-allowed' : 'cursor-pointer');
    dl.style.opacity = off ? '0.5' : '';
  }
  allBox = checkbox(true, t('ct.exp.all'), function (on) {
    (data.areas || []).forEach(function (a) { picked[a.key] = on; });
    boxes.forEach(function (b) { b.set(on); });
    paint();
  });
  allBox.setAttribute('data-ct-export-all', '');
  var list = h('ul', { class: 'eco-cx-areas', 'data-ct-export-areas': '' }, (data.areas || []).map(function (a) {
    var b = checkbox(true, ctTitle(a), function (on) { picked[a.key] = on; paint(); });
    b.setAttribute('data-ct-area', a.key);
    boxes.push(b);
    return h('li', { class: 'eco-cx-area' }, b);
  }));
  dl = button({ variant: 'primary', size: 'footer', icon: 'download', label: t('ct.exp.run'), onclick: function () {
    var keys = Object.keys(picked).filter(function (k) { return picked[k]; });
    if (!keys.length) return;
    setBusy(dl, true);
    ctDownload(keys.length === (data.areas || []).length ? [] : keys).then(function (name) {
      setBusy(dl, false, t('ct.exp.run'));
      paint();
      toast('success', t('ct.exp.done', { name: name }));
      if (data.reload) data.reload(true);
    }, function (err) {
      setBusy(dl, false, t('ct.exp.run'));
      paint();
      toast('error', errorText(err));
    });
  } });
  dl.setAttribute('data-ct-export', '');
  var foot = h('div', { class: 'flex flex-wrap items-center justify-between gap-3 eco-cx-foot pt-3 border-t border-secondary-200 dark:border-secondary-700' }, count, dl);
  var el = card(h('div', { class: 'p-4 flex flex-col gap-3 h-full', 'data-ct-export-card': '' },
    h('div', { class: 'flex flex-col gap-1' },
      h('h2', { class: 'font-semibold', text: t('ct.exp.title') }),
      h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: t('ct.exp.intro', { name: ctIso(data.serverName || '-') }) })),
    data.settings.exportEnabled ? null : notice('warning', t('ct.exp.off')),
    h('div', { class: 'flex items-center justify-between gap-2' }, allBox),
    list, foot), 'h-full');
  paint();
  return el;
}

function ctImportCard(data, onfile) {
  var max = (data.settings.maxFileMB || 20) * 1024 * 1024;
  var input = h('input', { type: 'file', accept: '.json,application/json', class: 'sr-only', 'data-ct-input': '' });
  var name = h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400 min-w-0 truncate', 'data-ct-file-name': '', text: t('ct.imp.none') });
  var file = null;
  var off = !data.admin || !data.settings.importEnabled;
  var pick = button({ variant: off ? 'disabled' : 'ghost', size: 'footer', icon: 'upload', label: t('ct.imp.pick'), disabled: off, onclick: function () { input.click(); } });
  pick.setAttribute('data-ct-pick', '');
  var run = button({ variant: 'disabled', size: 'footer', label: t('ct.imp.preview'), disabled: true, onclick: function () { if (file) onfile(file, run); } });
  run.setAttribute('data-ct-preview', '');
  input.addEventListener('change', function () {
    var f = input.files && input.files[0];
    input.value = '';
    if (!f) return;
    if (f.size > max) {
      toast('error', t('code.CONFIG_FILE_TOO_LARGE'));
      return;
    }
    file = f;
    name.textContent = f.name + ' · ' + ctSize(f.size);
    name.setAttribute('title', f.name);
    var fresh = button({ variant: 'primary', size: 'footer', label: t('ct.imp.preview'), onclick: function () { onfile(file, fresh); } });
    fresh.setAttribute('data-ct-preview', '');
    run.replaceWith(fresh);
    run = fresh;
    onfile(file, fresh);
  });
  return card(h('div', { class: 'p-4 flex flex-col gap-3 h-full', 'data-ct-import-card': '' },
    h('div', { class: 'flex flex-col gap-1' },
      h('h2', { class: 'font-semibold', text: t('ct.imp.title') }),
      h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 eco-pre-line', text: t('ct.imp.intro', { n: data.settings.maxFileMB || 20 }) })),
    data.settings.importEnabled ? null : notice('warning', t('ct.imp.off')),
    h('div', { class: 'eco-cx-drop' }, input, pick, name),
    h('ul', { class: 'eco-sdk-list text-sm text-secondary-600 dark:text-secondary-300' },
      h('li', { text: t('ct.imp.rule.keys') }),
      h('li', { text: t('ct.imp.rule.delete') }),
      h('li', { text: t('ct.imp.rule.tx') })),
    h('div', { class: 'flex justify-end eco-cx-foot pt-3 border-t border-secondary-200 dark:border-secondary-700' }, run)), 'h-full');
}

function ctRunsCard(runs) {
  if (!runs || !runs.length) {
    return card(h('div', { class: 'p-4 flex flex-col gap-2', 'data-ct-runs': '' },
      h('h2', { class: 'font-semibold text-sm', text: t('ct.runs') }),
      h('p', { class: 'text-sm text-secondary-500', text: t('ct.runs.none') })));
  }
  var columns = [{ label: t('ct.runs.kind'), width: '14%' }, { label: t('ct.runs.when'), width: '22%' }, { label: t('ct.runs.by') }, { label: t('ct.runs.source') }, { label: t('ct.runs.totals') }];
  return h('div', { class: 'flex flex-col gap-2', 'data-ct-runs': '' },
    h('h2', { class: 'font-semibold text-sm', text: t('ct.runs') }),
    table(columns, runs.map(function (r) {
      var tot = r.totals || {};
      var sum = r.kind === 'export'
        ? t('ct.runs.rows', { n: Object.keys(tot).reduce(function (s, k) { return s + (Number(tot[k]) || 0); }, 0) })
        : t('ct.runs.applied', { a: Number(tot.new) || 0, b: Number(tot.changed) || 0, c: Number(tot.switchedOff) || 0 });
      return { item: r, cells: [
        h('span', { 'data-ct-run': r.kind }, tag(t('ct.runs.' + r.kind), r.kind === 'import' ? 'primary' : 'secondary')),
        h('span', { class: 'text-sm tabular-nums', text: fmtDate(r.at) }),
        h('span', { class: 'text-sm', dir: 'auto', text: r.by || '-' }),
        h('span', { class: 'text-sm', dir: 'auto', text: (r.source && r.source.name) || '-' }),
        h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: sum })] };
    })));
}

function ctPreview(body, data, file, res, onapplied) {
  var state = { q: '', area: '', status: '', page: 1, limit: CT_PAGE };
  var pickedAreas = {};
  var switchOff = {};
  var excluded = {};
  res.areas.forEach(function (a) { pickedAreas[a.key] = a.selected; switchOff[a.key] = !!a.switchOff; });
  var rowsAll = [];
  res.areas.forEach(function (a) {
    a.rows.forEach(function (r) { rowsAll.push({ area: a.key, areaTitle: ctTitle(a), r: r, id: a.key + '|' + r.table + ':' + r.key }); });
  });
  var sel = selection({ name: 'ct-rows', idOf: function (it) { return it.id; } });
  var wrap = h('div', { class: 'flex flex-col gap-4', 'data-ct-preview-result': res.applied ? 'applied' : 'preview' });
  clear(body).appendChild(wrap);

  var src = res.source || {};
  var head = card(h('div', { class: 'p-4 flex flex-col gap-3' },
    h('div', { class: 'flex flex-wrap items-center justify-between gap-2' },
      h('h2', { class: 'font-semibold', text: res.applied ? t('ct.res.title') : t('ct.pv.title') }),
      res.applied ? tag(t('ct.res.applied'), 'success') : tag(t('ct.pv.dry'), 'secondary')),
    h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300', 'data-ct-source': src.name || '' }, h('span', { text: t('ct.pv.source', { name: ctIso(src.name || '-'), at: src.exportedAt ? fmtDate(src.exportedAt) : '-', by: ctIso(src.by || '-') }) })),
    res.unknownAreas && res.unknownAreas.length ? notice('warning', t('ct.pv.unknown', { list: res.unknownAreas.join('، ') })) : null));
  wrap.appendChild(head);
  var tot = res.totals;
  wrap.appendChild(h('div', { class: 'eco-sc-stats eco-cx-stats' },
    ctStat('new', tot.new, t('ct.st.new'), 'success'),
    ctStat('changed', tot.changed, t('ct.st.changed'), 'warning'),
    ctStat('same', tot.same, t('ct.st.same'), 'neutral'),
    ctStat('missing', tot.missingRefs + tot.secretNeeded + tot.errors, t('ct.stat.attention'), 'danger')));

  var areaCols = [{ label: '', width: '2.5rem' }, { label: t('ct.col.area') }, { label: t('ct.col.counts') }, { label: t('ct.col.switch'), width: '22%' }];
  var areaTable = table(areaCols, res.areas.map(function (a) {
    var c = a.counts;
    var box = checkbox(pickedAreas[a.key], '', function (on) { pickedAreas[a.key] = on; paintApply(); });
    box.setAttribute('aria-label', ctTitle(a));
    box.setAttribute('data-ct-pick-area', a.key);
    if (res.applied) box.style.pointerEvents = 'none';
    var counts = h('div', { class: 'flex flex-wrap gap-1' });
    CT_STATUSES.forEach(function (s) {
      var key = { missing_refs: 'missingRefs', secret_needed: 'secretNeeded', switched_off: 'switchedOff', error: 'errors' }[s] || s;
      if (c[key]) counts.appendChild(tag(t('ct.st.' + s) + ' ' + fmtNum(c[key]), CT_TONE[s]));
    });
    if (!counts.childNodes.length) counts.appendChild(h('span', { class: 'text-xs text-secondary-400', text: '-' }));
    var sw = a.canSwitchOff && !res.applied ? toggle(switchOff[a.key], t('ct.col.switch.on'), function (v) { switchOff[a.key] = v; }, 'sm') : h('span', { class: 'text-xs text-secondary-400', text: a.canSwitchOff ? (a.switchOff ? t('yes') : t('no')) : '-' });
    if (sw.setAttribute) sw.setAttribute('data-ct-switch', a.key);
    var info = h('div', { class: 'flex flex-col gap-0.5 min-w-0' },
      h('span', { class: 'font-medium text-sm', text: ctTitle(a) }),
      h('span', { class: 'text-xs text-secondary-500', text: t('ct.col.inFile', { n: a.inFile }) }),
      (a.warnings || []).length ? h('span', { class: 'text-xs eco-cx-warn', 'data-ct-warn': a.key, text: a.warnings.map(ctWarn).join('، ') }) : null);
    var row = h('div', { 'data-ct-area-row': a.key }, info);
    return { item: a, cells: [box, row, counts, sw] };
  }));
  wrap.appendChild(h('div', { class: 'flex flex-col gap-2' }, h('h3', { class: 'font-semibold text-sm', text: t('ct.pv.areas') }), areaTable));

  var listBox = h('div', { class: 'flex flex-col gap-2', 'data-ct-rows': '' });
  wrap.appendChild(h('div', { class: 'flex flex-col gap-2' }, h('h3', { class: 'font-semibold text-sm', text: t('ct.pv.rows') }), listBox));

  function filtered() {
    var q = state.q.toLowerCase();
    return rowsAll.filter(function (it) {
      if (state.area && it.area !== state.area) return false;
      if (state.status && it.r.status !== state.status) return false;
      if (q && (it.r.key + ' ' + it.r.table).toLowerCase().indexOf(q) < 0) return false;
      return true;
    });
  }

  function changesCell(r) {
    if (r.missing && r.missing.length) {
      return h('ul', { class: 'flex flex-col gap-0.5 text-xs' }, r.missing.map(function (m) {
        return h('li', { 'data-ct-missing': m.key }, h('span', { text: t('ct.miss', { field: m.field }) + ' ' }), mono(m.table + ' · ' + m.key));
      }));
    }
    if (r.changes && r.changes.length) {
      var shown = r.changes.slice(0, 4);
      var list = h('ul', { class: 'flex flex-col gap-0.5 text-xs' }, shown.map(function (c) {
        return h('li', { class: 'eco-cx-change', 'data-ct-change': c.field },
          h('span', { class: 'font-mono text-secondary-700 dark:text-secondary-200', dir: 'ltr', text: c.field }),
          h('span', { class: 'eco-cx-from font-mono', dir: 'ltr', 'data-l10n-skip': '', text: ctValue(c.from) }),
          h('span', { class: 'text-secondary-400', 'aria-hidden': 'true', text: '→' }),
          h('span', { class: 'eco-cx-to font-mono', dir: 'ltr', 'data-l10n-skip': '', text: ctValue(c.to) }));
      }));
      if (r.changes.length > shown.length) list.appendChild(h('li', { class: 'text-secondary-500', text: t('ct.more', { n: r.changes.length - shown.length }) }));
      return list;
    }
    if (r.reason) return h('span', { class: 'text-xs text-secondary-500', text: has('ct.reason.' + r.reason) ? t('ct.reason.' + r.reason) : r.reason });
    return h('span', { class: 'text-xs text-secondary-400', text: '-' });
  }

  function paintRows() {
    var list = filtered();
    var total = list.length;
    var pages = Math.max(1, Math.ceil(total / state.limit));
    if (state.page > pages) state.page = pages;
    var page = list.slice((state.page - 1) * state.limit, state.page * state.limit);
    sel.page = res.applied ? [] : page;
    var search = searchInput(state.q, t('ct.f.search'), function (v) { state.q = v; state.page = 1; paintRows(); });
    var areaF = filterSelect(t('ct.f.area'), state.area, res.areas.map(function (a) { return { value: a.key, label: ctTitle(a) }; }), function (v) { state.area = v; state.page = 1; paintRows(); });
    var statusF = filterSelect(t('ct.f.status'), state.status, CT_STATUSES.map(function (s) { return { value: s, label: t('ct.st.' + s) }; }), function (v) { state.status = v; state.page = 1; paintRows(); });
    areaF.setAttribute('data-ct-filter', 'area');
    statusF.setAttribute('data-ct-filter', 'status');
    var on = state.q || state.area || state.status;
    var bar2 = listToolbar({ search: search, filters: [areaF, statusF], clear: on ? clearFiltersButton(function () { state.q = ''; state.area = ''; state.status = ''; state.page = 1; paintRows(); }) : null, meta: listCount({ total: total, loaded: true }) });
    var cols = [];
    if (!res.applied) cols.push({ label: t('sel.all'), width: '2.5rem', head: sel.headCheck });
    cols.push({ label: t('ct.col.row') }, { label: t('ct.col.status'), width: '16%' }, { label: t('ct.col.changes'), width: '44%' });
    var tableEl = total ? table(cols, page.map(function (it) {
      var r = it.r;
      var cells = [];
      if (!res.applied) cells.push(sel.rowCheck(it, r.key));
      var out = excluded[it.id];
      cells.push(h('div', { class: 'flex flex-col gap-0.5 min-w-0', 'data-ct-row': r.table + ':' + r.key, 'data-status': r.status, 'data-excluded': out ? 'true' : null },
        h('span', { class: 'font-mono text-sm break-all', dir: 'ltr', 'data-l10n-skip': '', text: r.key }),
        h('span', { class: 'text-xs text-secondary-500' }, h('span', { text: it.areaTitle + ' · ' }), mono(r.table))));
      cells.push(h('div', { class: 'flex flex-wrap gap-1' }, ctStatusTag(r.status), out ? tag(t('ct.excluded'), 'secondary') : null));
      cells.push(changesCell(r));
      return { item: it, cells: cells };
    })) : card(h('div', { 'data-ct-rows-empty': '' }, emptyBlock(t('ct.rows.empty'), on ? t('ct.rows.empty.hint') : '')));
    var selBar = res.applied ? null : selectionBar(sel, { countKey: 'ct.sel.count', attr: 'data-ct-selbar', actions: [
      button({ variant: 'ghost', size: 'sm', label: t('ct.sel.exclude'), onclick: function () { sel.list().forEach(function (id) { excluded[id] = true; }); sel.clear(); paintRows(); paintApply(); } }),
      button({ variant: 'ghost', size: 'sm', label: t('ct.sel.include'), onclick: function () { sel.list().forEach(function (id) { delete excluded[id]; }); sel.clear(); paintRows(); paintApply(); } })
    ] });
    clear(listBox);
    listBox.appendChild(bar2);
    if (selBar) listBox.appendChild(selBar);
    listBox.appendChild(tableEl);
    listBox.appendChild(pager({ total: total, page: state.page, limit: state.limit, loaded: true }, function (p) { state.page = p; paintRows(); }));
    sel.paint();
  }

  function options() {
    var areas = res.areas.filter(function (a) { return pickedAreas[a.key]; }).map(function (a) { return a.key; });
    var rows = {};
    res.areas.forEach(function (a) {
      if (!pickedAreas[a.key]) return;
      var ids = rowsAll.filter(function (it) { return it.area === a.key && it.r.status !== 'switched_off'; });
      if (!ids.some(function (it) { return excluded[it.id]; })) return;
      rows[a.key] = ids.filter(function (it) { return !excluded[it.id]; }).map(function (it) { return it.r.table + ':' + it.r.key; });
    });
    var sw = {};
    Object.keys(switchOff).forEach(function (k) { if (switchOff[k] && pickedAreas[k]) sw[k] = true; });
    return { areas: areas, rows: rows, switchOff: sw };
  }

  var actions = h('div', { class: 'flex flex-wrap items-center justify-end gap-3', 'data-ct-actions': '' });
  wrap.appendChild(actions);
  var applyBtn;
  function paintApply() {
    if (res.applied) return;
    var opts = options();
    var n = opts.areas.length;
    applyBtn.disabled = !n;
    applyBtn.style.opacity = n ? '' : '0.5';
  }
  if (!res.applied) {
    var again = button({ variant: 'ghost', size: 'footer', icon: 'refresh', label: t('ct.pv.again'), onclick: function () {
      setBusy(again, true);
      ctSend('/preview', file, options()).then(function (out) { ctPreview(body, data, file, out.result, onapplied); }, function (err) { setBusy(again, false, t('ct.pv.again')); toast('error', errorText(err)); });
    } });
    again.setAttribute('data-ct-again', '');
    applyBtn = button({ variant: 'primary', size: 'footer', icon: 'check', label: t('ct.apply'), onclick: function () {
      var opts = options();
      var sum = res.totals;
      confirmModal({ title: t('ct.apply.title'), text: t('ct.apply.text', { n: opts.areas.length, a: sum.new, b: sum.changed, name: ctIso(src.name || '-') }), ok: t('ct.apply.ok') }).then(function (ok) {
        if (!ok) return;
        setBusy(applyBtn, true);
        ctSend('/apply', file, opts).then(function (out) {
          toast('success', t('ct.apply.done', { a: out.result.totals.new, b: out.result.totals.changed }));
          ctPreview(body, data, file, out.result, onapplied);
          onapplied();
        }, function (err) {
          setBusy(applyBtn, false, t('ct.apply'));
          toast('error', errorText(err));
        });
      });
    } });
    applyBtn.setAttribute('data-ct-apply', '');
    actions.appendChild(h('span', { class: 'text-sm text-secondary-500', text: t('ct.apply.hint') }));
    actions.appendChild(again);
    actions.appendChild(applyBtn);
    paintApply();
  }
  paintRows();
}

function configTransferPage(container) {
  var settingsBtn = button({ variant: 'ghost', size: 'footer', icon: 'gear', label: t('ct.set.title'), onclick: function () { if (current) ctSettingsPanel(current, function () { load(); }); } });
  settingsBtn.setAttribute('data-ct-open-settings', '');
  var frame = pageFrame(t('nav.config-transfer'), t('desc.config-transfer'), [settingsBtn]);
  clear(container).appendChild(frame.el);
  var body = h('div', { class: 'flex flex-col gap-4 eco-cx', 'data-ct-page': '' });
  frame.content.appendChild(body);
  var top = h('div', { class: 'eco-cx-grid' });
  var preview = h('div', { class: 'flex flex-col gap-4', 'data-ct-preview-box': '' });
  var runsBox = h('div', {});
  var current = null;

  function skeleton() {
    clear(body);
    body.appendChild(h('div', { class: 'eco-cx-grid' }, [0, 1].map(function () {
      return card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('40%', 'h-4'), bar('85%'), bar('70%'), bar('60%'), bar('50%')));
    })));
    body.appendChild(table([{ label: t('ct.runs.kind') }, { label: t('ct.runs.when') }, { label: t('ct.runs.by') }], skeletonRows([{}, {}, {}], 3)));
  }

  function onfile(file, btn) {
    var label = t('ct.imp.preview');
    setBusy(btn, true);
    clear(preview).appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('30%', 'h-4'), bar('90%'), bar('75%'))));
    ctSend('/preview', file, {}).then(function (out) {
      setBusy(btn, false, label);
      ctPreview(preview, current, file, out.result, function () { load(true); });
      preview.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }, function (err) {
      setBusy(btn, false, label);
      clear(preview).appendChild(h('div', { 'data-ct-error': err.code || '' }, notice('error', errorText(err))));
    });
  }

  function paint(data) {
    current = data;
    data.reload = load;
    settingsBtn.disabled = !data.admin;
    settingsBtn.style.display = data.admin ? '' : 'none';
    clear(body);
    if (!data.admin) body.appendChild(h('div', { 'data-ct-admin-only': '' }, notice('info', t('ct.adminOnly'))));
    clear(top);
    top.appendChild(ctExportCard(data));
    top.appendChild(ctImportCard(data, onfile));
    body.appendChild(top);
    body.appendChild(preview);
    clear(runsBox).appendChild(ctRunsCard(data.runs));
    body.appendChild(runsBox);
  }

  function load(runsOnly) {
    if (runsOnly && current) {
      baseApi('GET', CT_PATH).then(function (data) { current.runs = data.runs; clear(runsBox).appendChild(ctRunsCard(data.runs)); }, function () {});
      return;
    }
    skeleton();
    baseApi('GET', CT_PATH).then(paint, function (err) {
      clear(body).appendChild(errorBlock(err, load));
    });
  }
  load();
}

CUSTOM_PAGES['config-transfer'] = function (root) { configTransferPage(root); };
