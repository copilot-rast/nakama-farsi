ICONS['content'] = '<path d="M17.5 19a4.5 4.5 0 1 0-1.2-8.8A6 6 0 0 0 4.5 12 3.5 3.5 0 0 0 6 19"/><path d="M12 12v9M8.5 17.5 12 21l3.5-3.5"/>';
ICONS['content-packs'] = ICONS['content'];
ICONS['content-settings'] = '<path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3M1 14h6M9 8h6M17 16h6"/>';

var CT_K = [0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5, 0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174, 0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da, 0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967, 0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85, 0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070, 0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3, 0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2];

function ctSha256Js(bytes) {
  var hh = [0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19];
  var len = bytes.length;
  var total = ((len + 9 + 63) >> 6) << 6;
  var buf = new Uint8Array(total);
  buf.set(bytes);
  buf[len] = 0x80;
  var bits = len * 8;
  var dv = new DataView(buf.buffer);
  dv.setUint32(total - 8, Math.floor(bits / 0x100000000));
  dv.setUint32(total - 4, bits >>> 0);
  var w = new Uint32Array(64);
  for (var off = 0; off < total; off += 64) {
    for (var i = 0; i < 16; i++) w[i] = dv.getUint32(off + i * 4);
    for (i = 16; i < 64; i++) {
      var a0 = w[i - 15], a1 = w[i - 2];
      var s0 = ((a0 >>> 7) | (a0 << 25)) ^ ((a0 >>> 18) | (a0 << 14)) ^ (a0 >>> 3);
      var s1 = ((a1 >>> 17) | (a1 << 15)) ^ ((a1 >>> 19) | (a1 << 13)) ^ (a1 >>> 10);
      w[i] = (w[i - 16] + s0 + w[i - 7] + s1) | 0;
    }
    var a = hh[0], b = hh[1], c = hh[2], d = hh[3], e = hh[4], f = hh[5], g = hh[6], k = hh[7];
    for (i = 0; i < 64; i++) {
      var S1 = ((e >>> 6) | (e << 26)) ^ ((e >>> 11) | (e << 21)) ^ ((e >>> 25) | (e << 7));
      var t1 = (k + S1 + ((e & f) ^ (~e & g)) + CT_K[i] + w[i]) | 0;
      var S0 = ((a >>> 2) | (a << 30)) ^ ((a >>> 13) | (a << 19)) ^ ((a >>> 22) | (a << 10));
      var t2 = (S0 + ((a & b) ^ (a & c) ^ (b & c))) | 0;
      k = g; g = f; f = e; e = (d + t1) | 0; d = c; c = b; b = a; a = (t1 + t2) | 0;
    }
    hh[0] = (hh[0] + a) | 0; hh[1] = (hh[1] + b) | 0; hh[2] = (hh[2] + c) | 0; hh[3] = (hh[3] + d) | 0;
    hh[4] = (hh[4] + e) | 0; hh[5] = (hh[5] + f) | 0; hh[6] = (hh[6] + g) | 0; hh[7] = (hh[7] + k) | 0;
  }
  return hh.map(function (x) { return ('00000000' + (x >>> 0).toString(16)).slice(-8); }).join('');
}

function ctHex(buf) {
  return Array.prototype.map.call(new Uint8Array(buf), function (x) { return ('0' + x.toString(16)).slice(-2); }).join('');
}

function ctSha256(file) {
  return file.arrayBuffer().then(function (buf) {
    if (window.crypto && window.crypto.subtle && window.isSecureContext) return window.crypto.subtle.digest('SHA-256', buf).then(ctHex);
    return ctSha256Js(new Uint8Array(buf));
  });
}

function ctBytes(n) {
  n = Number(n || 0);
  var units = [['ct.unit.gb', 1 << 30], ['ct.unit.mb', 1 << 20], ['ct.unit.kb', 1 << 10]];
  for (var i = 0; i < units.length; i++) {
    if (n >= units[i][1]) return t(units[i][0], { n: fmtNum(Math.round(n / units[i][1] * 10) / 10, { maximumFractionDigits: 1 }) });
  }
  return t('ct.unit.b', { n: fmtNum(n) });
}

function ctRange(p) {
  if (!p.minAppVersion && !p.maxAppVersion) return t('ct.range.all');
  if (!p.maxAppVersion) return t('ct.range.from', { v: p.minAppVersion });
  if (!p.minAppVersion) return t('ct.range.to', { v: p.maxAppVersion });
  return t('ct.range.between', { a: p.minAppVersion, b: p.maxAppVersion });
}

function ctLtr(text, extra) {
  return h('span', { dir: 'ltr', class: 'font-mono text-xs break-all' + (extra ? ' ' + extra : ''), text: text });
}

function ctStatus(p, live) {
  return h('div', { class: 'flex flex-wrap gap-1', 'data-ct-status': p.status },
    p.status === 'ready' ? tag(t('ct.status.ready'), 'secondary') : tag(t('ct.status.uploading'), 'warning'),
    p.active ? tag(t('ct.active'), 'success') : null,
    live ? tag(t('ct.live'), 'primary') : null);
}

function ctCopyRow(label, value) {
  if (!value) return null;
  var btn = button({ variant: 'muted', size: 'icon', icon: 'copy', title: t('ct.copy'), onclick: function () {
    copyText(value).then(function () { toast('success', t('ct.copied')); }, function () { toast('error', t('err.generic')); });
  } });
  btn.setAttribute('data-ct-copy', label);
  return h('div', { class: 'flex items-start gap-2 py-1.5 eco-fg-row', 'data-ct-url': label },
    h('div', { class: 'min-w-0 flex-1' },
      h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('ct.url.' + label) }),
      ctLtr(value)),
    btn);
}

function ctOpenPanel(key, title, body, footer, onclose) {
  closePanel();
  var el = sidePanel(title, body, footer, onclose);
  var esc = function (e) { if (e.key === 'Escape' && panelState && !document.querySelector('[data-economy="dialog"], [data-form-dialog]')) { e.preventDefault(); onclose(); } };
  document.addEventListener('keydown', esc, true);
  panelState = { el: el, key: key, esc: esc };
  document.body.appendChild(el);
  return el;
}

function ctPackPanel(id, onchange) {
  var body = h('div', { class: 'flex flex-col gap-4', 'data-ct-pack': id }, bar('60%'), bar('80%'), bar('40%'));
  var footer = h('div', { class: 'flex flex-wrap gap-2 w-full items-center justify-end' });
  ctOpenPanel('content/' + id, t('ct.pack'), body, footer, closePanel);
  var write = aclCan('SETTINGS', 1);
  function act(path, method, okKey) {
    return baseApi(method || 'POST', '/content/packs/' + encodeURIComponent(id) + path, method === 'DELETE' ? undefined : {}).then(function () {
      toast('success', t(okKey));
      onchange();
      if (method === 'DELETE') closePanel(); else load();
    }, function (err) { toast('error', errorText(err)); });
  }
  function load() {
    baseApi('GET', '/content/packs/' + encodeURIComponent(id)).then(function (res) {
      var p = res.item;
      var ix = res.index || {};
      clear(body);
      clear(footer);
      body.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2' },
        h('span', { class: 'text-lg font-semibold', dir: 'ltr', text: p.platform + ' · ' + p.version }), ctStatus(p, false)));
      body.appendChild(dl([
        ['ct.f.range', h('span', { dir: 'auto', text: ctRange(p) })],
        ['ct.f.catalog', p.catalogPath ? ctLtr(p.catalogPath) : null],
        ['ct.f.hash', p.catalogHash ? ctLtr(p.catalogHash) : null],
        ['ct.f.files', h('span', { class: 'tabular-nums', text: t('ct.filesN', { n: fmtNum(p.fileCount), size: ctBytes(p.totalBytes) }) })],
        ['ct.f.note', p.note || null],
        ['ct.f.created', h('span', { text: fmtDate(p.createdAt, true) + (p.createdBy ? ' · ' + p.createdBy : '') })],
        ['ct.f.activated', p.activatedAt ? h('span', { text: fmtDate(p.activatedAt, true) + (p.activatedBy ? ' · ' + p.activatedBy : '') }) : null]
      ]));
      if (p.status === 'ready') {
        body.appendChild(h('section', { 'data-ct-urls': '' },
          h('h3', { class: 'font-medium mb-1', text: t('ct.urls') }),
          h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-2', text: t('ct.urls.hint') }),
          ctCopyRow('base', ix.baseUrl), ctCopyRow('catalog', ix.catalogUrl), ctCopyRow('hash', ix.hashUrl)));
      } else {
        body.appendChild(notice('warning', t('ct.notReady')));
      }
      var files = res.files || [];
      body.appendChild(h('section', { 'data-ct-files': '' },
        h('h3', { class: 'font-medium mb-2', text: t('ct.files') }),
        table([{ label: t('ct.col.path') }, { label: t('ct.col.size'), width: 110 }, { label: t('col.status'), width: 110 }], files.slice(0, 500).map(function (f) {
          return { item: f, cells: [ctLtr(f.path), h('span', { class: 'tabular-nums whitespace-nowrap', text: ctBytes(f.size) }), f.stored ? tag(t('ct.stored'), 'success') : tag(t('ct.missing'), 'warning')] };
        }))));
      var events = res.events || [];
      if (events.length) {
        body.appendChild(h('section', { 'data-ct-events': '' },
          h('h3', { class: 'font-medium mb-2', text: t('ct.events') }),
          h('div', { class: 'flex flex-col text-sm' }, events.map(function (ev) {
            return h('div', { class: 'flex flex-wrap items-center gap-2 py-1 eco-fg-row' },
              tag(t('ct.ev.' + ev.action), ev.action === 'activated' ? 'success' : 'secondary'),
              h('span', { class: 'text-secondary-500', text: ev.actor || '' }),
              h('span', { class: 'flex-1' }),
              h('span', { class: 'text-xs text-secondary-500 whitespace-nowrap', text: fmtDate(ev.createdAt, true) }));
          }))));
      }
      if (!write) return;
      if (p.status === 'ready' && !p.active) {
        var on = button({ variant: 'primary', size: 'footer', icon: 'check', label: t('ct.activate'), onclick: function () { setBusy(on, true); act('/activate', 'POST', 'ct.packOn'); } });
        on.setAttribute('data-ct-action', 'activate');
        footer.appendChild(on);
      }
      if (p.active) {
        var off = button({ variant: 'ghost', size: 'footer', label: t('ct.deactivate'), onclick: function () { setBusy(off, true); act('/deactivate', 'POST', 'ct.packOff'); } });
        off.setAttribute('data-ct-action', 'deactivate');
        footer.appendChild(off);
      }
      if (!p.active) {
        var del = button({ variant: 'danger', size: 'footer', icon: 'trash', label: t('delete'), onclick: function () {
          confirmModal({ title: t('ct.delete.title'), text: t('ct.delete.text', { v: p.platform + ' ' + p.version }), danger: true, ok: t('delete') }).then(function (yes) { if (yes) act('', 'DELETE', 'ct.deleted'); });
        } });
        del.setAttribute('data-ct-action', 'delete');
        footer.insertBefore(del, footer.firstChild);
      }
    }, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  load();
}

function ctCommonRoot(paths) {
  if (!paths.length) return '';
  var first = paths[0].split('/');
  if (first.length < 2) return '';
  var root = first[0];
  for (var i = 1; i < paths.length; i++) if (paths[i].split('/')[0] !== root || paths[i].indexOf('/') === -1) return '';
  return root;
}

function ctUploadPanel(info, onDone) {
  var state = { platform: info.platforms[0] || 'Android', version: '', minAppVersion: '', maxAppVersion: '', note: '', activate: true, files: [], busy: false };
  var body = h('div', { 'data-ct-upload': '' });
  var footer = h('div', { class: 'flex gap-2 w-full items-center justify-end' });
  ctOpenPanel('content/upload', t('ct.upload'), body, footer, function () { if (!state.busy) closePanel(); });
  var top = h('div', {});
  var platform = selectInput(state.platform, info.platforms.map(function (p) { return { value: p, label: p }; }), function (v) { state.platform = v; summary(); }, { label: t('ct.f.platform'), placeholder: false });
  platform.setAttribute('data-ct-field', 'platform');
  var version = textInput('', { ltr: true, mono: true, label: t('ct.f.version'), placeholder: '1.0.0', oninput: function (v) { state.version = toLatinDigits(v).trim(); } });
  version.setAttribute('data-ct-field', 'version');
  var minV = textInput('', { ltr: true, mono: true, label: t('ct.f.min'), placeholder: '1.0', oninput: function (v) { state.minAppVersion = toLatinDigits(v).trim(); } });
  minV.setAttribute('data-ct-field', 'min');
  var maxV = textInput('', { ltr: true, mono: true, label: t('ct.f.max'), placeholder: '1.9', oninput: function (v) { state.maxAppVersion = toLatinDigits(v).trim(); } });
  maxV.setAttribute('data-ct-field', 'max');
  var note = textInput('', { label: t('ct.f.note'), oninput: function (v) { state.note = v; } });
  var activate = checkbox(true, t('ct.f.activate'), function (on) { state.activate = on; });
  activate.setAttribute('data-ct-field', 'activate');
  var dirInput = h('input', { type: 'file', multiple: true, hidden: true, 'data-ct-input': 'folder' });
  dirInput.setAttribute('webkitdirectory', '');
  var fileInput = h('input', { type: 'file', multiple: true, hidden: true, 'data-ct-input': 'files' });
  function picked(list, fromFolder) {
    var arr = Array.prototype.slice.call(list || []);
    var paths = arr.map(function (f) { return (fromFolder && f.webkitRelativePath) || f.name; });
    var root = fromFolder ? ctCommonRoot(paths) : '';
    var seen = {};
    state.files = state.files.filter(function (x) { return !fromFolder && !arr.some(function (f) { return f.name === x.path; }); });
    state.files.forEach(function (x) { seen[x.path] = true; });
    arr.forEach(function (f, i) {
      var p = paths[i];
      if (root && p.indexOf(root + '/') === 0 && root !== state.platform && root !== 'ServerData') p = p.slice(root.length + 1);
      if (/(^|\/)(\.DS_Store|Thumbs\.db)$/.test(p) || /^__MACOSX\//.test(p) || seen[p]) return;
      seen[p] = true;
      state.files.push({ path: p, file: f });
    });
    summary();
  }
  dirInput.addEventListener('change', function () { picked(dirInput.files, true); dirInput.value = ''; });
  fileInput.addEventListener('change', function () { picked(fileInput.files, false); fileInput.value = ''; });
  var pickRow = h('div', { class: 'flex flex-wrap gap-2 mb-2' },
    button({ variant: 'ghost', size: 'sm', icon: 'upload', label: t('ct.pick.folder'), onclick: function () { dirInput.click(); } }),
    button({ variant: 'ghost', size: 'sm', icon: 'plus', label: t('ct.pick.files'), onclick: function () { fileInput.click(); } }),
    button({ variant: 'ghost', size: 'sm', label: t('ct.pick.clear'), onclick: function () { state.files = []; summary(); } }),
    dirInput, fileInput);
  var sum = h('div', { class: 'text-sm', 'data-ct-summary': '' });
  var progress = h('div', { class: 'flex flex-col gap-1', hidden: true, 'data-ct-progress': '' });
  var progBar = h('div', { class: 'eco-ct-progress' }, h('span', { style: 'width:0%' }));
  var progText = h('div', { class: 'text-xs text-secondary-500 tabular-nums' });
  progress.appendChild(progBar);
  progress.appendChild(progText);
  function catalogOf() {
    var re = /^catalog[^/]*\.(json|bin)$/;
    try { if (info.pattern) re = new RegExp(info.pattern); } catch (e) { re = /^catalog[^/]*\.(json|bin)$/; }
    return state.files.filter(function (f) {
      var p = f.path.split('/');
      if (p[0] === 'ServerData' && p[1] === state.platform) p = p.slice(2);
      else if (p.length > 1 && p[0] === state.platform) p = p.slice(1);
      return p.length === 1 && re.test(p[0]);
    });
  }
  function summary() {
    clear(sum);
    if (!state.files.length) { sum.appendChild(h('p', { class: 'text-secondary-500', text: t('ct.pick.none') })); return; }
    var size = state.files.reduce(function (a, f) { return a + f.file.size; }, 0);
    var cats = catalogOf();
    sum.appendChild(h('p', { class: 'tabular-nums', text: t('ct.filesN', { n: fmtNum(state.files.length), size: ctBytes(size) }) }));
    if (cats.length === 1) sum.appendChild(h('p', { class: 'text-secondary-500' }, h('span', { text: t('ct.catalogFound') + ' ' }), ctLtr(cats[0].path)));
    else sum.appendChild(h('div', { class: 'mt-2' }, notice('warning', t(cats.length ? 'code.PACK_CATALOG_AMBIGUOUS' : 'code.PACK_CATALOG_MISSING'))));
  }
  function setProgress(done, total, text) {
    progress.hidden = false;
    progBar.firstChild.style.width = (total ? Math.min(100, Math.round(done / total * 100)) : 0) + '%';
    progText.textContent = text;
  }
  body.appendChild(h('div', {}, top,
    h('div', { class: 'eco-form-row' }, field(t('ct.f.platform'), platform), field(t('ct.f.version'), version, t('ct.f.version.hint'))),
    h('div', { class: 'eco-form-row' }, field(t('ct.f.min'), minV, t('ct.f.min.hint'), { optional: true }), field(t('ct.f.max'), maxV, null, { optional: true })),
    field(t('ct.f.note'), note, null, { optional: true }),
    h('div', { class: 'mb-5' }, activate),
    h('div', { class: 'mb-5' }, h('p', { class: 'font-medium mb-1', text: t('ct.f.files') }), h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-2', text: t('ct.pick.hint') }), pickRow, sum),
    progress));
  summary();
  var send = button({ variant: 'primary', size: 'footer', icon: 'upload', label: t('ct.upload.go'), onclick: function () { run(); } });
  send.setAttribute('data-ct-send', '');
  footer.appendChild(send);
  function fail(err) {
    state.busy = false;
    setBusy(send, false);
    clear(top).appendChild(h('div', { class: 'mb-4' }, notice('error', errorText(err))));
    top.scrollIntoView({ block: 'start' });
  }
  function run() {
    clear(top);
    if (!state.version) { fail({ code: 'PACK_INVALID', message: '' }); return; }
    if (!state.files.length) { fail({ code: 'PACK_CATALOG_MISSING' }); return; }
    state.busy = true;
    setBusy(send, true, t('ct.uploading'));
    var list = [];
    var bySha = {};
    var i = 0;
    function hashNext() {
      if (i >= state.files.length) return Promise.resolve();
      var f = state.files[i];
      setProgress(i, state.files.length, t('ct.step.hash', { n: localDigits(String(i + 1)), total: localDigits(String(state.files.length)) }));
      return ctSha256(f.file).then(function (sha) {
        list.push({ path: f.path, sha256: sha, size: f.file.size });
        bySha[sha] = f.file;
        i++;
        return hashNext();
      });
    }
    var packId = null;
    hashNext().then(function () {
      setProgress(0, 1, t('ct.step.plan'));
      return baseApi('POST', '/content/packs', { platform: state.platform, version: state.version, minAppVersion: state.minAppVersion, maxAppVersion: state.maxAppVersion, note: state.note.trim(), files: list });
    }).then(function (res) {
      packId = res.pack.id;
      var jobs = [];
      var bytes = 0;
      (res.missing || []).forEach(function (m) {
        var have = {};
        (m.received || []).forEach(function (n) { have[n] = true; });
        for (var n = 0; n < m.parts; n++) {
          if (have[n]) continue;
          var start = n * m.partBytes;
          var end = Math.min(m.size, start + m.partBytes);
          jobs.push({ sha: m.sha256, n: n, start: start, end: end });
          bytes += end - start;
        }
      });
      var sent = 0;
      var j = 0;
      function next() {
        if (j >= jobs.length) return Promise.resolve();
        var job = jobs[j];
        setProgress(sent, bytes, t('ct.step.send', { done: ctBytes(sent), total: ctBytes(bytes) }));
        return bySha[job.sha].slice(job.start, job.end).arrayBuffer().then(function (buf) {
          return baseApi('PUT', '/content/packs/' + encodeURIComponent(packId) + '/blobs/' + job.sha + '/' + job.n, buf, true);
        }).then(function () {
          sent += job.end - job.start;
          j++;
          return next();
        });
      }
      return next();
    }).then(function () {
      setProgress(1, 1, t('ct.step.finish'));
      return baseApi('POST', '/content/packs/' + encodeURIComponent(packId) + '/complete', { activate: state.activate });
    }).then(function (res) {
      state.busy = false;
      toast('success', t('ct.uploaded'));
      onDone();
      ctPackPanel(res.item.id, onDone);
    }, fail);
  }
}

function contentPacksPage(container, query) {
  var write = aclCan('SETTINGS', 1);
  var state = { platform: query.platform || '', status: query.status || '' };
  var info = { platforms: [], pattern: '' };
  var upload = write ? button({ variant: 'primary', size: 'sm', icon: 'upload', label: t('ct.upload'), onclick: function () { ctUploadPanel(info, load); } }) : null;
  if (upload) upload.setAttribute('data-ct-new', '');
  var frame = pageFrame(t('nav.content-packs'), t('desc.content-packs'), upload ? [upload] : null);
  clear(container).appendChild(frame.el);
  var filters = h('div', { class: 'eco-filters flex flex-wrap items-center gap-2 pb-2' });
  var live = h('div', { class: 'eco-stat-grid mb-4', 'data-ct-live': '' });
  var previewBox = h('div', { class: 'mb-4' });
  var listBox = h('div', { 'data-ct-list': '' });
  frame.content.appendChild(filters);
  frame.content.appendChild(live);
  frame.content.appendChild(previewBox);
  frame.content.appendChild(listBox);
  var cols = [{ label: t('ct.col.pack'), skeleton: '50%' }, { label: t('ct.col.range'), width: 170 }, { label: t('col.status'), width: 170 }, { label: t('ct.col.files'), width: 130 }, { label: t('ct.col.uploaded'), width: 160 }];
  function paintFilters() {
    clear(filters);
    filters.appendChild(filterSelect(t('ct.f.platform'), state.platform, info.platforms.map(function (p) { return { value: p, label: p }; }), function (v) { state.platform = v; replaceHash(hashFor('content-packs', null, state)); load(); }));
    filters.appendChild(filterSelect(t('col.status'), state.status, [{ value: 'ready', label: t('ct.status.ready') }, { value: 'uploading', label: t('ct.status.uploading') }, { value: 'active', label: t('ct.active') }], function (v) { state.status = v; replaceHash(hashFor('content-packs', null, state)); load(); }));
  }
  function paintPreview() {
    var pv = { platform: info.platforms[0] || '', appVersion: '' };
    var out = h('div', { class: 'text-sm mt-2', 'data-ct-preview-result': '' });
    var sel = selectInput(pv.platform, info.platforms.map(function (p) { return { value: p, label: p }; }), function (v) { pv.platform = v; }, { label: t('ct.f.platform'), placeholder: false });
    var ver = textInput('', { ltr: true, mono: true, label: t('ct.f.appVersion'), placeholder: '1.2.0', oninput: function (v) { pv.appVersion = toLatinDigits(v).trim(); } });
    ver.setAttribute('data-ct-preview', 'version');
    var go = button({ variant: 'ghost', size: 'sm', label: t('ct.preview.go'), onclick: function () {
      clear(out);
      baseApi('GET', '/content/packs/preview' + qs(pv)).then(function (res) {
        clear(out);
        if (!res.pack) { out.appendChild(h('span', { class: 'text-secondary-500', 'data-ct-preview-none': '', text: t('ct.preview.none') })); return; }
        out.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2' }, h('span', { text: t('ct.preview.gets') }), h('a', { href: '#', class: 'eco-link', dir: 'ltr', text: res.pack.platform + ' · ' + res.pack.version, onclick: function (e) { e.preventDefault(); ctPackPanel(res.pack.id, load); } })));
        out.appendChild(ctLtr(res.pack.catalogUrl, 'text-secondary-500'));
      }, function (err) { out.appendChild(notice('error', errorText(err))); });
    } });
    go.setAttribute('data-ct-preview', 'go');
    clear(previewBox).appendChild(card(h('div', { class: 'p-4' },
      h('h2', { class: 'font-medium mb-1', text: t('ct.preview') }),
      h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: t('ct.preview.hint') }),
      h('div', { class: 'flex flex-wrap items-end gap-2' }, h('div', { class: 'eco-ct-field' }, sel), h('div', { class: 'eco-ct-field eco-ct-grow' }, ver), go),
      out)));
  }
  function load() {
    clear(listBox).appendChild(table(cols, skeletonRows(cols, 3)));
    baseApi('GET', '/content/packs' + qs(state)).then(function (res) {
      var first = !info.platforms.length;
      info.platforms = res.platforms || [];
      if (first) { paintFilters(); paintPreview(); }
      var current = res.current || {};
      clear(live);
      var liveKeys = Object.keys(current);
      live.appendChild(statCard(t('ct.stat.packs'), (res.items || []).length, null, null));
      live.appendChild(statCard(t('ct.stat.live'), liveKeys.length, liveKeys.join(t('ov.sep')) || t('ct.stat.liveNone'), null));
      var st = res.storage || {};
      live.appendChild(statCard(t('ct.stat.storage'), st.files, t('ct.stat.storageSub', { size: ctBytes(st.bytes), driver: st.driver || '' }), null));
      var items = res.items || [];
      clear(listBox);
      if (!items.length) {
        listBox.appendChild(card(emptyBlock(t('ct.empty'), t('ct.empty.hint'), upload ? button({ variant: 'primary', size: 'sm', icon: 'upload', label: t('ct.upload'), onclick: function () { ctUploadPanel(info, load); } }) : null)));
        return;
      }
      listBox.appendChild(table(cols, items.map(function (p) {
        var isLive = current[p.platform] === p.id;
        return { item: p, cells: [
          h('div', { class: 'flex flex-col min-w-0', 'data-ct-row': p.platform + ':' + p.version }, h('span', { class: 'font-medium', dir: 'ltr', text: p.platform + ' · ' + p.version }), p.note ? h('span', { class: 'text-xs text-secondary-500 truncate', text: p.note }) : null),
          h('span', { class: 'text-sm', dir: 'auto', text: ctRange(p) }),
          ctStatus(p, isLive),
          h('span', { class: 'tabular-nums text-sm whitespace-nowrap', text: t('ct.filesN', { n: fmtNum(p.fileCount), size: ctBytes(p.totalBytes) }) }),
          h('span', { class: 'text-sm text-secondary-500 whitespace-nowrap', text: fmtDate(p.completedAt || p.createdAt, true) })
        ] };
      }), { onrow: function (p) { ctPackPanel(p.id, load); } }));
    }, function (err) { clear(listBox).appendChild(errorBlock(err, load)); });
  }
  baseApi('GET', '/content/settings').then(function (res) { info.pattern = ((res.settings || {}).packs || {}).catalogPattern || ''; }, function () {});
  load();
  return { reload: load };
}

var CT_SETTINGS_FIELDS = [
  { type: 'heading', label: 'ct.h.upload' },
  { name: 'packs.platforms', type: 'rows', label: 'ct.s.platforms', hint: 'ct.s.platforms.hint', addLabel: 'ct.s.platforms.add', emptyText: 'ct.s.platforms.empty', max: 32, empty: [],
    newRow: function () { return { v: '' }; }, columns: [{ name: 'v', type: 'text', ltr: true, label: 'ct.f.platform', required: true }] },
  { type: 'row', fields: [
    { name: 'partKB', type: 'number', label: 'ct.s.part', hint: 'ct.s.part.hint', min: 64, max: 3072, required: true },
    { name: 'packs.maxFiles', type: 'number', label: 'ct.s.maxFiles', min: 1, max: 100000, required: true }
  ] },
  { type: 'row', fields: [
    { name: 'fileMB', type: 'number', label: 'ct.s.fileMB', min: 1, max: 4096, required: true },
    { name: 'packMB', type: 'number', label: 'ct.s.packMB', min: 1, max: 65536, required: true }
  ] },
  { name: 'packs.catalogPattern', type: 'text', ltr: true, label: 'ct.s.pattern', hint: 'ct.s.pattern.hint', required: true },
  { name: 'packs.requireHash', type: 'bool', inline: true, label: 'ct.s.requireHash' },
  { type: 'heading', label: 'ct.h.serve' },
  { name: 'packs.singleActive', type: 'bool', inline: true, label: 'ct.s.single', hint: 'ct.s.single.hint' },
  { name: 'packs.requireAppVersion', type: 'bool', inline: true, label: 'ct.s.requireVersion', hint: 'ct.s.requireVersion.hint' },
  { name: 'packs.publicIndex', type: 'bool', inline: true, label: 'ct.s.publicIndex', hint: 'ct.s.publicIndex.hint' },
  { name: 'packs.indexCacheSeconds', type: 'number', label: 'ct.s.indexCache', min: 0, max: 86400, required: true },
  { name: 'packs.baseUrl', type: 'text', ltr: true, label: 'ct.s.baseUrl', hint: 'ct.s.baseUrl.hint', optional: true },
  { name: 'packs.redirectToStorage', type: 'bool', inline: true, label: 'ct.s.redirect', hint: 'ct.s.redirect.hint' },
  { type: 'heading', label: 'ct.h.keep' },
  { type: 'row', fields: [
    { name: 'packs.uploadHours', type: 'number', label: 'ct.s.uploadHours', hint: 'ct.s.uploadHours.hint', min: 1, max: 720, required: true },
    { name: 'packs.keepInactive', type: 'number', label: 'ct.s.keepInactive', hint: 'ct.s.keepInactive.hint', min: 0, max: 1000, required: true }
  ] }
];

function ctSettingsCall(method, path, body) {
  return baseApi(method, path, body).then(function (res) {
    var s = clone((res && res.settings) || {});
    s.packs = s.packs || {};
    s.packs.platforms = (s.packs.platforms || []).map(function (p) { return { v: p }; });
    s.partKB = Math.round((s.packs.partBytes || 0) / 1024);
    s.fileMB = Math.round((s.packs.maxFileBytes || 0) / 1048576);
    s.packMB = Math.round((s.packs.maxPackBytes || 0) / 1048576);
    s._storage = res.storage;
    s._base = res.publicBaseUrl;
    return s;
  });
}

function contentSettingsPage(container) {
  var frame = pageFrame(t('nav.content-settings'), t('desc.content-settings'));
  clear(container).appendChild(frame.el);
  var page = apiPage(ctSettingsCall, '/content/settings', null, CT_SETTINGS_FIELDS, {
    id: 'content',
    onchange: function (v, extra) {
      clear(extra);
      if (v._storage) extra.appendChild(h('div', { class: 'mb-4' }, notice('info', t('ct.s.storage', { driver: v._storage, base: v._base || '—' }))));
    },
    fixPayload: function (p) {
      var out = { packs: Object.assign({}, p.packs) };
      out.packs.platforms = (p.packs.platforms || []).map(function (r) { return (r.v || '').trim(); }).filter(Boolean);
      out.packs.partBytes = Number(p.partKB) * 1024;
      out.packs.maxFileBytes = Number(p.fileMB) * 1048576;
      out.packs.maxPackBytes = Number(p.packMB) * 1048576;
      out.packs.baseUrl = (out.packs.baseUrl || '').trim();
      return out;
    }
  });
  frame.content.appendChild(page.el);
  page.load();
}

CUSTOM_PAGES['content-packs'] = function (root, route, ctl) { ctl.list = contentPacksPage(root, route.query || {}); };
CUSTOM_PAGES['content-settings'] = function (root) { contentSettingsPage(root); };
