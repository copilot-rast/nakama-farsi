var KEY_RE = /^[a-z0-9][a-z0-9_.-]{0,63}$/;
var CURRENCY_KEY_RE = /^[a-z0-9_]{2,32}$/;
var CODE_RE = /^[A-Z0-9_-]{3,32}$/;
var refCache = {};

function refLabel(item) {
  if (!item) return '';
  var name = i18nText(item.name) || i18nText(item.title) || i18nText(item.question);
  return name || item.key || item.username || item.id || '';
}

function textInput(value, opts) {
  opts = opts || {};
  var el = h(opts.multiline ? 'textarea' : 'input', {
    class: INPUT_MD + (opts.mono ? ' font-mono text-sm' : '') + (opts.multiline ? ' eco-textarea' : ''),
    type: opts.multiline ? null : 'text',
    dir: opts.ltr ? 'ltr' : 'auto',
    rows: opts.multiline ? (opts.rows || 4) : null,
    inputmode: opts.inputmode,
    placeholder: opts.placeholder,
    autocomplete: 'off',
    spellcheck: opts.ltr ? 'false' : null,
    'aria-label': opts.label
  });
  el.value = value == null ? '' : value;
  if (opts.oninput) el.addEventListener('input', function () { opts.oninput(el.value); });
  return el;
}

function numberInput(value, opts) {
  var shown = value === null || value === undefined || value === '' ? '' : localDigits(String(value));
  var el = textInput(shown, { inputmode: opts.int ? 'numeric' : 'decimal', label: opts.label, placeholder: opts.placeholder, oninput: function () {
    var raw = toLatinDigits(el.value).replace(/,/g, '').trim();
    opts.oninput(raw);
  } });
  el.classList.add('tabular-nums');
  el.addEventListener('blur', function () {
    var raw = toLatinDigits(el.value).replace(/,/g, '').trim();
    if (raw !== '' && isFinite(Number(raw))) el.value = localDigits(raw);
  });
  return el;
}

function parseNumber(raw, int) {
  if (raw === null || raw === undefined || raw === '') return null;
  var s = toLatinDigits(String(raw)).replace(/,/g, '').trim();
  if (s === '') return null;
  if (int ? !/^-?\d+$/.test(s) : !/^-?\d+(\.\d+)?$/.test(s)) return NaN;
  return Number(s);
}

function selectInput(value, options, onchange, opts) {
  opts = opts || {};
  var sel = h('select', { class: SELECT_MD, 'aria-label': opts.label });
  if (opts.placeholder !== false) sel.appendChild(h('option', { value: '', text: opts.placeholder || t('choose') }));
  options.forEach(function (o) {
    var op = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) op.selected = true;
    sel.appendChild(op);
  });
  sel.addEventListener('change', function () { onchange(sel.value); });
  return sel;
}

function refPicker(opts) {
  var call = opts.api || api;
  var value = opts.value == null ? '' : opts.value;
  var cacheKey = function (v) { return opts.resource + ':' + (opts.valueKey || 'id') + ':' + v; };
  var boxLabel = h('span', { class: 'flex items-center gap-2 min-w-0 truncate' });
  var box = h('button', { type: 'button', class: SELECT_MD + ' eco-start flex items-center gap-2 cursor-pointer', 'aria-haspopup': 'listbox', 'aria-label': opts.label });
  box.appendChild(boxLabel);
  var clearBtn = opts.required ? null : button({ variant: 'muted', size: 'icon', icon: 'close', iconClass: 'h-3 w-3', title: t('remove'), class: 'eco-ref-clear', onclick: function (e) { e.stopPropagation(); set('', null); } });
  var list = h('div', { class: 'overflow-y-auto max-h-64', role: 'listbox' });
  var search = h('input', { class: INPUT_SM, type: 'text', placeholder: t('search.short'), 'aria-label': t('search.short') });
  var pop = h('div', { class: 'eco-pop absolute z-30 w-full mt-1 bg-white dark:bg-secondary-800 border border-secondary-200 dark:border-secondary-700 rounded-md shadow-lg p-1', hidden: true },
    h('div', { class: 'p-1 mb-0.5' }, search), list);
  var wrap = h('div', { class: 'relative eco-ref' }, box, clearBtn, pop);
  var seq = 0;

  function paint(item) {
    clear(boxLabel);
    if (!value) {
      boxLabel.appendChild(h('span', { class: 'text-secondary-400', text: opts.placeholder || t('choose') }));
    } else if (item) {
      boxLabel.appendChild(h('span', { text: refLabel(item) }));
      if (item.key && refLabel(item) !== item.key) boxLabel.appendChild(h('span', { class: 'font-mono text-xs text-secondary-500', dir: 'ltr', text: item.key }));
    } else {
      boxLabel.appendChild(h('span', { class: 'font-mono text-xs', dir: 'ltr', text: value }));
    }
    if (clearBtn) clearBtn.hidden = !value;
  }

  function set(v, item) {
    value = v || '';
    if (item) refCache[cacheKey(value)] = item;
    paint(item || refCache[cacheKey(value)]);
    opts.onchange(value || null, item || null);
  }

  function load(q) {
    var my = ++seq;
    clear(list);
    for (var i = 0; i < 3; i++) list.appendChild(h('div', { class: 'px-2 py-2' }, bar(['60%', '40%', '50%'][i])));
    var params = Object.assign({ q: q, page: 1, limit: 20 }, opts.filter || {});
    call('GET', '/' + opts.resource + qs(params)).then(function (res) {
      if (my !== seq) return;
      clear(list);
      var items = (res && res.items) || [];
      if (!items.length) list.appendChild(h('div', { class: 'px-2 py-2 text-sm text-secondary-500', text: !q && opts.emptyText ? opts.emptyText : t('no.results') }));
      items.forEach(function (it) {
        var v = it[opts.valueKey || 'id'];
        refCache[cacheKey(v)] = it;
        var row = h('button', { type: 'button', role: 'option', 'aria-selected': String(v) === String(value) ? 'true' : 'false',
          class: 'w-full eco-start px-2 py-1.5 rounded-md hover:bg-secondary-100 dark:hover:bg-secondary-700 flex items-center justify-between gap-2 text-sm' + (String(v) === String(value) ? ' eco-selected font-medium' : ''),
          onclick: function () { close(); set(v, it); } },
          h('span', { class: 'truncate', text: refLabel(it) }),
          it.key ? h('span', { class: 'font-mono text-xs text-secondary-500', dir: 'ltr', text: it.key }) : null);
        list.appendChild(row);
      });
    }, function (err) {
      if (my !== seq) return;
      clear(list);
      list.appendChild(h('div', { class: 'px-2 py-2 text-sm text-error-600 dark:text-error-400', text: errorText(err) }));
    });
  }

  function outside(e) { if (!wrap.contains(e.target)) close(); }
  function open() {
    pop.hidden = false;
    search.value = '';
    load('');
    search.focus();
    document.addEventListener('mousedown', outside, true);
  }
  function close() {
    pop.hidden = true;
    document.removeEventListener('mousedown', outside, true);
  }
  box.addEventListener('click', function () { if (pop.hidden) open(); else close(); });
  search.addEventListener('input', debounce(function () { load(toLatinDigits(search.value).trim()); }, 250));
  wrap.addEventListener('keydown', function (e) { if (e.key === 'Escape' && !pop.hidden) { e.stopPropagation(); close(); box.focus(); } });

  paint(refCache[cacheKey(value)]);
  if (value && !refCache[cacheKey(value)]) {
    var req = (opts.valueKey || 'id') === 'id'
      ? call('GET', '/' + opts.resource + '/' + encodeURIComponent(value))
      : call('GET', '/' + opts.resource + qs(Object.assign({ q: value, limit: 20 }, opts.filter || {}))).then(function (res) {
        return ((res && res.items) || []).filter(function (it) { return it[opts.valueKey] === value; })[0] || null;
      });
    req.then(function (item) { if (item) { refCache[cacheKey(value)] = item; paint(item); } }, function () { return null; });
  }
  return wrap;
}

function refMulti(opts) {
  var values = (opts.value || []).slice();
  var chips = h('div', { class: 'flex flex-wrap gap-1.5 mb-2' });
  function paint() {
    clear(chips);
    chips.hidden = !values.length;
    values.forEach(function (v) {
      var item = refCache[opts.resource + ':id:' + v];
      chips.appendChild(h('span', { class: 'eco-tag eco-tag--primary inline-flex items-center gap-1 leading-tight px-2 py-0.5 text-xs rounded-md' },
        h('span', { text: item ? refLabel(item) : v }),
        h('button', { type: 'button', class: 'eco-chip-x', 'aria-label': t('remove'), onclick: function () {
          values = values.filter(function (x) { return x !== v; });
          paint();
          opts.onchange(values.slice());
        } }, icon('close', 'h-3 w-3'))));
    });
  }
  var call = opts.api || api;
  var picker = refPicker({ resource: opts.resource, api: opts.api, emptyText: opts.emptyText, filter: opts.filter, label: opts.label, placeholder: opts.placeholder || t('choose'), value: '', onchange: function (v) {
    if (v && values.indexOf(v) === -1) {
      values.push(v);
      paint();
      opts.onchange(values.slice());
    }
    picker.querySelector('.eco-ref-clear') && picker.querySelector('.eco-ref-clear').click();
  } });
  var missing = values.filter(function (v) { return !refCache[opts.resource + ':id:' + v]; });
  missing.forEach(function (v) {
    call('GET', '/' + opts.resource + '/' + encodeURIComponent(v)).then(function (it) {
      if (!it) return;
      refCache[opts.resource + ':id:' + v] = it;
      paint();
    }, function () { return null; });
  });
  paint();
  return h('div', {}, chips, picker);
}

function imageInput(value, onchange) {
  var url = value || '';
  var preview = h('div', { class: 'h-16 w-16 shrink-0 rounded-md border border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-800 flex items-center justify-center overflow-hidden text-secondary-400' });
  var urlInput = textInput(url, { ltr: true, label: t('f.uploadUrl'), placeholder: 'https://', oninput: function (v) { url = v.trim(); paint(); onchange(url || null); } });
  urlInput.classList.add('text-sm');
  var file = h('input', { type: 'file', accept: 'image/png,image/jpeg,image/webp,image/svg+xml', class: 'hidden' });
  var err = h('p', { class: 'text-xs text-error-600 dark:text-error-400 mt-1', hidden: true });
  var up = button({ variant: 'link', size: 'sm', icon: 'upload', label: t('upload'), onclick: function () { file.click(); } });
  function paint() {
    clear(preview);
    if (url) preview.appendChild(img(url));
    else preview.appendChild(icon('image', 'h-6 w-6'));
  }
  file.addEventListener('change', function () {
    var f = file.files && file.files[0];
    if (!f) return;
    err.hidden = true;
    var fd = new FormData();
    fd.append('file', f, f.name);
    setBusy(up, true, t('uploading'));
    api('POST', '/uploads', fd, true).then(function (res) {
      url = (res && res.url) || '';
      urlInput.value = url;
      paint();
      onchange(url || null);
    }, function (e) {
      err.textContent = e && e.status === 400 ? t('err.upload') : errorText(e);
      err.hidden = false;
    }).then(function () { setBusy(up, false); file.value = ''; });
  });
  paint();
  return h('div', {}, h('div', { class: 'flex items-center gap-3' }, preview, h('div', { class: 'flex-1 min-w-0 flex flex-col gap-1.5' }, urlInput, h('div', {}, up))), file, err);
}

function toLocalInput(v) {
  if (!v) return '';
  var d = new Date(v);
  if (isNaN(d.getTime())) return '';
  var p = function (n) { return (n < 10 ? '0' : '') + n; };
  return d.getFullYear() + '-' + p(d.getMonth() + 1) + '-' + p(d.getDate()) + 'T' + p(d.getHours()) + ':' + p(d.getMinutes());
}

function dateInput(value, onchange) {
  var el = h('input', { type: 'datetime-local', class: INPUT_MD + ' tabular-nums', dir: 'ltr' });
  el.value = toLocalInput(value);
  var shown = h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1' });
  function paint() { shown.textContent = el.value ? fmtDate(new Date(el.value).toISOString()) : ''; shown.hidden = !el.value || lang() !== 'fa'; }
  el.addEventListener('input', function () {
    paint();
    onchange(el.value ? new Date(el.value).toISOString() : null);
  });
  paint();
  return h('div', {}, el, shown);
}

function mdEscape(s) {
  return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&#34;').replace(/'/g, '&#39;');
}

function mdSafeUrl(u) {
  if (/[\s"'<>`]/.test(u)) return false;
  if (/^https?:\/\//i.test(u)) return true;
  return u.charAt(0) === '/' && u.charAt(1) !== '/' && u.charAt(1) !== '\\';
}

function mdLink(s, start) {
  var depth = 0;
  var i = start + 1;
  for (; i < s.length; i++) {
    if (s[i] === '[') depth++;
    else if (s[i] === ']') { if (depth === 0) break; depth--; }
  }
  if (i >= s.length - 1 || s[i + 1] !== '(') return null;
  var close = s.indexOf(')', i + 2);
  if (close < 0) return null;
  return { text: s.slice(start + 1, i), url: s.slice(i + 2, close).trim(), end: close + 1 };
}

function mdInline(s, links) {
  var out = '';
  var i = 0;
  while (i < s.length) {
    var m;
    if (s.slice(i, i + 2) === '![' && (m = mdLink(s, i + 1)) && mdSafeUrl(m.url)) {
      out += '<img src="' + mdEscape(m.url) + '" alt="' + mdEscape(m.text) + '">';
      i = m.end;
      continue;
    }
    if (s[i] === '[' && links && (m = mdLink(s, i)) && mdSafeUrl(m.url)) {
      out += '<a href="' + mdEscape(m.url) + '" rel="noopener noreferrer" target="_blank">' + mdInline(m.text, false) + '</a>';
      i = m.end;
      continue;
    }
    if (s.slice(i, i + 2) === '**') {
      var e2 = s.indexOf('**', i + 2);
      if (e2 > i + 2) { out += '<strong>' + mdInline(s.slice(i + 2, e2), links) + '</strong>'; i = e2 + 2; continue; }
    }
    if (s[i] === '*' && i + 1 < s.length && s[i + 1] !== ' ' && s[i + 1] !== '*') {
      var e1 = s.indexOf('*', i + 1);
      if (e1 > i + 1) { out += '<em>' + mdInline(s.slice(i + 1, e1), links) + '</em>'; i = e1 + 1; continue; }
    }
    var j = i + 1;
    while (j < s.length && s[j] !== '*' && s[j] !== '[' && s[j] !== '!') j++;
    out += mdEscape(s.slice(i, j));
    i = j;
  }
  return out;
}

function mdHtml(src) {
  var blocks = [];
  var cur = null;
  function flush() { if (cur && cur.lines.length) blocks.push(cur); cur = null; }
  String(src || '').replace(/\r\n?/g, '\n').split('\n').forEach(function (line) {
    var tl = line.trim();
    if (!tl) { flush(); return; }
    if (/^### /.test(tl)) { flush(); blocks.push({ kind: 'h3', lines: [tl.slice(4).trim()] }); return; }
    if (/^## /.test(tl)) { flush(); blocks.push({ kind: 'h2', lines: [tl.slice(3).trim()] }); return; }
    var kind = /^[-*] /.test(tl) ? 'ul' : /^[0-9]{1,3}[.)] /.test(tl) ? 'ol' : 'p';
    if (!cur || cur.kind !== kind) { flush(); cur = { kind: kind, lines: [] }; }
    cur.lines.push(kind === 'ul' ? tl.slice(2).trim() : kind === 'ol' ? tl.replace(/^[0-9]{1,3}[.)] /, '').trim() : tl);
  });
  flush();
  return blocks.map(function (b) {
    if (b.kind === 'h2' || b.kind === 'h3') return '<' + b.kind + '>' + mdInline(b.lines[0], true) + '</' + b.kind + '>';
    if (b.kind === 'ul' || b.kind === 'ol') return '<' + b.kind + '>' + b.lines.map(function (l) { return '<li>' + mdInline(l, true) + '</li>'; }).join('') + '</' + b.kind + '>';
    return '<p>' + b.lines.map(function (l) { return mdInline(l, true); }).join('<br>') + '</p>';
  }).join('');
}

function mdPlain(s) {
  return String(s || '').replace(/!\[[^\]]*\]\([^)]*\)/g, '').replace(/\[([^\]]*)\]\([^)]*\)/g, '$1').replace(/^\s*(#{2,3}|[-*]|\d{1,3}[.)])\s+/gm, '').replace(/\*\*|\*/g, '').replace(/\s+/g, ' ').trim();
}

function mdPreview(src, ltr) {
  var box = h('div', { class: 'eco-md text-sm text-secondary-800 dark:text-secondary-200', dir: ltr ? 'ltr' : 'rtl' });
  var html = mdHtml(src);
  if (html) box.innerHTML = html;
  else box.appendChild(h('p', { class: 'text-secondary-400', text: t('md.empty') }));
  return box;
}

function markdownInput(value, opts) {
  var text = value == null ? '' : String(value);
  var ta = textInput(text, { multiline: true, rows: opts.rows || 6, ltr: opts.ltr, label: opts.label, oninput: function (v) { text = v; opts.oninput(v); if (counter) paintCount(); } });
  var preview = h('div', { class: 'eco-md-box border border-secondary-200 dark:border-secondary-700 rounded-md px-3 py-2 bg-secondary-50 dark:bg-secondary-900', hidden: true });
  var counter = opts.max ? h('span', { class: 'text-xs tabular-nums text-secondary-500' }) : null;
  function paintCount() {
    var n = Array.from(text).length;
    counter.textContent = t('chars', { n: n, max: opts.max });
    counter.classList.toggle('text-error-600', n > opts.max);
  }
  var seg = segmented([{ value: 'write', label: t('md.write') }, { value: 'preview', label: t('md.preview') }], 'write', function (v) {
    ta.hidden = v === 'preview';
    preview.hidden = v !== 'preview';
    if (v === 'preview') clear(preview).appendChild(mdPreview(text, opts.ltr));
  }, t('md.mode'));
  if (counter) paintCount();
  return h('div', { class: 'eco-md-input' },
    h('div', { class: 'flex items-center justify-between gap-2 mb-2' }, seg, counter),
    ta, preview,
    h('div', { class: 'flex flex-wrap items-center gap-1 text-xs text-secondary-500 dark:text-secondary-400 mt-1' },
      h('span', { text: t('md.hint') }),
      ['**bold**', '*italic*', '[text](https://…)', '## heading', '- list', '![](https://…)'].map(function (s) { return h('code', { class: 'eco-md-token', dir: 'ltr', text: s }); })));
}

var GOTO_SCREENS = ['home', 'play', 'shop', 'gift', 'payments', 'news', 'faq', 'inbox'];

function actionsEditor(value, onchange) {
  var rows = (value || []).map(function (r) { return { type: r.type || 'url', value: r.value || '', label: { fa: (r.label && r.label.fa) || '', en: (r.label && r.label.en) || '' } }; });
  var listEl = h('div', { class: 'flex flex-col gap-2' });
  var errors = [];
  var listId = 'eco-goto-' + Math.random().toString(36).slice(2, 8);
  var datalist = h('datalist', { id: listId }, GOTO_SCREENS.map(function (s) { return h('option', { value: s }); }));
  function emit() { onchange(rows.map(function (r) { return { type: r.type, value: r.value, label: { fa: r.label.fa, en: r.label.en } }; })); }
  var add = button({ variant: 'link', size: 'sm', icon: 'plus', label: t('act.add'), onclick: function () {
    rows.push({ type: 'url', value: '', label: { fa: '', en: '' } });
    render();
    emit();
  } });
  function render() {
    clear(listEl);
    errors = [];
    add.hidden = rows.length >= 5;
    if (!rows.length) listEl.appendChild(h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400 border border-dashed border-secondary-300 dark:border-secondary-700 rounded-md px-3 py-4 text-center', text: t('act.empty') }));
    rows.forEach(function (row, idx) {
      var cell = function (labelKey, ctl) {
        return h('div', { class: 'min-w-0' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t(labelKey) }), ctl);
      };
      var typeSel = selectInput(row.type, [{ value: 'url', label: t('act.url') }, { value: 'goto', label: t('act.goto') }], function (v) { row.type = v; render(); emit(); }, { placeholder: false, label: t('act.type') });
      var val = textInput(row.value, { ltr: true, mono: true, label: t('act.value'), placeholder: row.type === 'url' ? 'https://' : 'shop', oninput: function (v) { row.value = v.trim(); emit(); } });
      if (row.type === 'goto') val.setAttribute('list', listId);
      var fa = textInput(row.label.fa, { label: t('act.label.fa'), oninput: function (v) { row.label.fa = v; emit(); } });
      var en = textInput(row.label.en, { ltr: true, label: t('act.label.en'), oninput: function (v) { row.label.en = v; emit(); } });
      var err = h('p', { class: 'text-xs text-error-600 dark:text-error-400 eco-items-err', hidden: true });
      errors.push({ row: row, el: err });
      listEl.appendChild(h('div', { class: 'border border-secondary-200 dark:border-secondary-700 rounded-md p-3', 'data-action-row': '' },
        h('div', { class: 'eco-actions-row' },
          cell('act.type', typeSel),
          cell('act.value', val),
          cell('act.label.fa', fa),
          cell('act.label.en', en),
          h('div', { class: 'eco-items-remove' }, button({ variant: 'danger', size: 'iconLg', icon: 'trash', title: t('remove'), onclick: function () { rows.splice(idx, 1); render(); emit(); } }))),
        h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1.5', text: t(row.type === 'url' ? 'act.url.hint' : 'act.goto.hint') }),
        err));
    });
  }
  render();
  var el = h('div', {}, listEl, datalist, h('div', { class: 'mt-2' }, add));
  el.validate = function () {
    var bad = false;
    errors.forEach(function (e) {
      var r = e.row;
      var msg = '';
      if (!r.label.fa.trim()) msg = t('act.label.fa') + ': ' + t('err.required');
      else if (!r.value) msg = t('act.value') + ': ' + t('err.required');
      else if (r.type === 'url' && !/^https?:\/\/\S+$/i.test(r.value)) msg = t('code.INVALID_ACTION');
      e.el.textContent = msg;
      e.el.hidden = !msg;
      e.el.parentElement.classList.toggle('eco-row-invalid', !!msg);
      if (msg) bad = true;
    });
    return bad ? t('err.form') : null;
  };
  return el;
}

var ITEM_ACTIONS = { currency: ['add'], meta: ['add', 'set'], storage: ['give'] };

function itemsEditor(value, onchange) {
  var rows = (value || []).map(function (r) { return Object.assign({}, r); });
  var listEl = h('div', { class: 'flex flex-col gap-2' });
  var errors = [];
  function emit() { onchange(rows.map(function (r) { return Object.assign({}, r); })); }
  function render() {
    clear(listEl);
    errors = [];
    if (!rows.length) listEl.appendChild(h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400 border border-dashed border-secondary-300 dark:border-secondary-700 rounded-md px-3 py-4 text-center', text: t('f.items.empty') }));
    rows.forEach(function (row, idx) {
      var typeSel = selectInput(row.type, ['currency', 'meta', 'storage'].map(function (k) { return { value: k, label: t('item.' + k) }; }), function (v) {
        row.type = v;
        row.key = '';
        row.action = (ITEM_ACTIONS[v] || ['add'])[0];
        row.value = v === 'storage' ? {} : 0;
        render();
        emit();
      }, { placeholder: false, label: t('f.item.type') });
      var keyCtl;
      if (row.type === 'storage') {
        keyCtl = textInput(row.key, { ltr: true, mono: true, label: t('f.item.key'), oninput: function (v) { row.key = v.trim(); emit(); } });
      } else {
        keyCtl = refPicker({ resource: row.type === 'meta' ? 'metas' : 'currencies', valueKey: 'key', value: row.key, required: true, label: t('f.item.key'), onchange: function (v) { row.key = v || ''; emit(); } });
      }
      var actSel = selectInput(row.action, (ITEM_ACTIONS[row.type] || ['add']).map(function (k) { return { value: k, label: t('action.' + k) }; }), function (v) { row.action = v; emit(); }, { placeholder: false, label: t('f.item.action') });
      var valCtl;
      if (row.type === 'storage') {
        valCtl = textInput(row.value == null ? '{}' : JSON.stringify(row.value), { ltr: true, mono: true, label: t('f.item.value'), oninput: function (v) {
          try { row.value = JSON.parse(v); row._bad = false; } catch (e) { row._bad = true; }
          emit();
        } });
      } else {
        valCtl = numberInput(row.value, { int: row.type === 'currency', label: t('f.item.value'), oninput: function (v) { row._raw = v; row.value = parseNumber(v, row.type === 'currency'); emit(); } });
      }
      var err = h('p', { class: 'text-xs text-error-600 dark:text-error-400 eco-items-err', hidden: true });
      errors.push({ row: row, el: err });
      var cell = function (labelKey, ctl) {
        return h('div', { class: 'min-w-0' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t(labelKey) }), ctl);
      };
      listEl.appendChild(h('div', { class: 'border border-secondary-200 dark:border-secondary-700 rounded-md p-3' },
        h('div', { class: 'eco-items-row' },
          cell('f.item.type', typeSel),
          cell('f.item.key', keyCtl),
          cell('f.item.action', actSel),
          cell('f.item.value', valCtl),
          h('div', { class: 'eco-items-remove' }, button({ variant: 'danger', size: 'iconLg', icon: 'trash', title: t('remove'), onclick: function () { rows.splice(idx, 1); render(); emit(); } }))),
        err));
    });
  }
  var add = button({ variant: 'link', size: 'sm', icon: 'plus', label: t('add.row'), onclick: function () {
    rows.push({ type: 'currency', key: '', action: 'add', value: 0 });
    render();
    emit();
  } });
  render();
  var el = h('div', {}, listEl, h('div', { class: 'mt-2' }, add));
  el.validate = function () {
    var bad = false;
    errors.forEach(function (e) {
      var r = e.row;
      var msg = '';
      if (!r.key) msg = t('f.item.key') + ': ' + t('err.required');
      else if (r.type === 'storage' && (r._bad || r.value === null || typeof r.value !== 'object')) msg = t('err.json');
      else if (r.type !== 'storage' && (r.value === null || r.value === undefined || isNaN(r.value))) msg = t('f.item.value') + ': ' + t('err.number');
      e.el.textContent = msg;
      e.el.hidden = !msg;
      e.el.parentElement.classList.toggle('eco-row-invalid', !!msg);
      if (msg) bad = true;
    });
    return bad ? t('err.form') : null;
  };
  return el;
}

function buildForm(fields, values, ctx) {
  var root = h('div', {});
  var entries = [];

  function addField(f, parent) {
    if (f.type === 'row') {
      var row = h('div', { class: 'eco-form-row' });
      f.fields.forEach(function (sub) { addField(sub, row); });
      var entryRow = { def: f, el: row, row: true };
      entries.push(entryRow);
      parent.appendChild(row);
      return;
    }
    if (f.type === 'heading') {
      var hd = h('div', { class: 'text-sm font-semibold text-secondary-800 dark:text-secondary-200 mb-3 mt-1 pt-4 border-t border-secondary-200 dark:border-secondary-700', text: t(f.label) });
      entries.push({ def: f, el: hd, row: true });
      parent.appendChild(hd);
      return;
    }
    var cur = getPath(values, f.name);
    var change = function (v) {
      setPath(values, f.name, v);
      if (entry.wrap.setError) entry.wrap.setError('');
      refresh();
    };
    var label = f.label ? t(f.label) : '';
    var control;
    var t2 = f.type || 'text';
    if (t2 === 'text' || t2 === 'key' || t2 === 'textarea' || t2 === 'code') {
      control = textInput(cur, { ltr: t2 === 'key' || t2 === 'code' || f.ltr, mono: t2 === 'key' || t2 === 'code', multiline: t2 === 'textarea', label: label, oninput: function (v) {
        if (t2 === 'key') v = v.trim();
        if (t2 === 'code') v = v.trim().toUpperCase();
        change(v);
      } });
      if (t2 === 'code') control.classList.add('uppercase');
      if (f.lockOnEdit && values.id) {
        control.readOnly = true;
        control.classList.add('eco-readonly');
      }
    } else if (t2 === 'json') {
      control = textInput(cur == null ? '' : JSON.stringify(cur, null, 2), { ltr: true, mono: true, multiline: true, rows: 5, label: label, oninput: function (v) {
        entry.raw = v;
        setPath(values, f.name, v);
        if (entry.wrap.setError) entry.wrap.setError('');
      } });
      entry = { raw: cur == null ? '' : JSON.stringify(cur) };
    } else if (t2 === 'number') {
      control = numberInput(cur, { int: f.int !== false, label: label, oninput: function (v) { change(v === '' ? null : v); } });
    } else if (t2 === 'bool') {
      control = toggle(cur, f.inline ? label : null, function (v) { change(v); });
    } else if (t2 === 'select') {
      control = selectInput(cur, f.options(), function (v) { change(v || null); }, { label: label, placeholder: f.required ? false : t('none') });
      if (f.required && (cur == null || cur === '')) { setPath(values, f.name, f.options()[0].value); }
    } else if (t2 === 'multi') {
      var set = (cur || []).slice();
      control = h('div', { class: 'flex flex-wrap gap-4' });
      f.options().forEach(function (o) {
        control.appendChild(checkbox(set.indexOf(o.value) !== -1, o.label, function (on) {
          set = set.filter(function (x) { return x !== o.value; });
          if (on) set.push(o.value);
          var order = f.options().map(function (x) { return x.value; });
          set.sort(function (a, b) { return order.indexOf(a) - order.indexOf(b); });
          change(set.slice());
        }));
      });
    } else if (t2 === 'ref') {
      control = refPicker({ resource: f.resource, valueKey: f.valueKey, filter: f.filter, value: cur, required: f.required, label: label, onchange: function (v) { change(v); } });
    } else if (t2 === 'refMulti') {
      control = refMulti({ resource: f.resource, filter: f.filter, value: cur, label: label, onchange: function (v) { change(v); } });
    } else if (t2 === 'image') {
      control = imageInput(cur, function (v) { change(v); });
    } else if (t2 === 'datetime') {
      control = dateInput(cur, function (v) { change(v); });
    } else if (t2 === 'items') {
      control = itemsEditor(cur, function (v) { setPath(values, f.name, v); if (entry.wrap.setError) entry.wrap.setError(''); });
    } else if (t2 === 'markdown') {
      control = markdownInput(cur, { ltr: f.ltr, rows: f.rows, max: f.max, label: label, oninput: function (v) { change(v); } });
    } else if (t2 === 'actions') {
      control = actionsEditor(cur, function (v) { setPath(values, f.name, v); if (entry.wrap.setError) entry.wrap.setError(''); });
    } else if (t2 === 'static') {
      control = h('div', { class: 'py-2 text-secondary-700 dark:text-secondary-200' }, f.render(values));
    }
    var entry = entry || {};
    entry.def = f;
    entry.control = control;
    if (t2 === 'bool' && f.inline) {
      entry.wrap = h('div', { class: 'mb-5 eco-field' }, control,
        f.hint ? h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: t(f.hint) }) : null);
      entry.wrap.setError = function () {};
    } else {
      var hintKey = f.lockOnEdit && values.id ? 'f.key.locked' : f.hint;
      entry.wrap = field(label, control, hintKey ? t(hintKey) : null, { optional: f.optional });
    }
    entries.push(entry);
    parent.appendChild(entry.wrap);
  }

  function refresh() {
    entries.forEach(function (e) {
      var visible = !e.def.when || e.def.when(values);
      var el = e.row ? e.el : e.wrap;
      el.hidden = !visible;
      e.hidden = !visible;
    });
    if (ctx && ctx.onchange) ctx.onchange(values);
  }

  function validate() {
    var first = null;
    var errs = 0;
    entries.forEach(function (e) {
      if (e.row || e.hidden) return;
      var f = e.def;
      var v = getPath(values, f.name);
      var msg = null;
      var t2 = f.type || 'text';
      if (t2 === 'json') {
        var raw = typeof v === 'string' ? v : (v == null ? '' : JSON.stringify(v));
        if (raw.trim() === '') setPath(values, f.name, f.empty !== undefined ? f.empty : {});
        else { try { setPath(values, f.name, JSON.parse(raw)); } catch (x) { msg = t('err.json'); } }
      } else if (t2 === 'number') {
        var n = parseNumber(v, f.int !== false);
        if (n === null) { if (f.required) msg = t('err.required'); setPath(values, f.name, null); }
        else if (isNaN(n) || (f.min !== undefined && n < f.min) || (f.max !== undefined && n > f.max)) msg = t('err.number');
        else setPath(values, f.name, n);
      } else if (t2 === 'items' || t2 === 'actions') {
        msg = e.control.validate();
      } else if (f.required && (v === null || v === undefined || v === '' || (Array.isArray(v) && !v.length))) {
        msg = t('err.required');
      } else if (t2 === 'key' && v && !(f.keyRe || KEY_RE).test(v)) {
        msg = t(f.keyRe ? 'err.key' : 'err.key.general');
      } else if (t2 === 'code' && v && !CODE_RE.test(v)) {
        msg = t('err.code');
      }
      if (!msg && f.check) msg = f.check(v, values);
      e.wrap.setError(msg);
      if (msg) { errs++; if (!first) first = e.wrap; }
    });
    if (first) first.scrollIntoView({ block: 'center', behavior: 'smooth' });
    return errs === 0;
  }

  function showServerError(err) {
    if (!err || !err.field) return false;
    var name = String(err.field).replace(/\[\d+\]/g, '').replace(/_([a-z])/g, function (m, c) { return c.toUpperCase(); });
    var best = null;
    entries.forEach(function (e) {
      if (e.row || e.hidden || !e.def.name) return;
      if (name === e.def.name || name.indexOf(e.def.name + '.') === 0 || e.def.name.indexOf(name + '.') === 0) { if (!best) best = e; }
    });
    if (!best) return false;
    best.wrap.setError(errorText(err));
    best.wrap.scrollIntoView({ block: 'center', behavior: 'smooth' });
    return true;
  }

  function payload() {
    var out = {};
    fields.forEach(function walk(f) {
      if (f.type === 'row') { f.fields.forEach(walk); return; }
      if (!f.name || f.readonly || f.type === 'static' || f.type === 'heading') return;
      var visible = !f.when || f.when(values);
      var v = getPath(values, f.name);
      if (!visible) v = f.empty !== undefined ? f.empty : null;
      if (f.type === 'actions' && Array.isArray(v)) v = v.map(function (r) { var label = { fa: r.label.fa.trim() }; if (r.label.en.trim()) label.en = r.label.en.trim(); return { type: r.type, value: r.value, label: label }; });
      if (f.type === 'items' && Array.isArray(v)) v = v.map(function (r) { return { type: r.type, key: r.key, action: r.action, value: r.value }; });
      if (/\.(fa|en)$/.test(f.name)) {
        var parentPath = f.name.slice(0, -3);
        if (typeof getPath(out, parentPath) !== 'object' || getPath(out, parentPath) === null) setPath(out, parentPath, {});
        if (v === null || v === undefined || String(v).trim() === '') return;
        v = String(v).trim();
      }
      setPath(out, f.name, v === undefined ? null : v);
    });
    return out;
  }

  fields.forEach(function (f) { addField(f, root); });
  refresh();
  return { el: root, validate: validate, payload: payload, showServerError: showServerError };
}
