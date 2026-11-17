var clRefs = {};
function loadRefs(kind, force) {
  if (clRefs[kind] && !force) return Promise.resolve(clRefs[kind]);
  return api('GET', '/' + kind + '/refs' + qs({ lang: clLang() })).then(function (r) { clRefs[kind] = r || {}; return clRefs[kind]; }, function () { clRefs[kind] = {}; return clRefs[kind]; });
}

function refList(kind, list) {
  return ((clRefs[kind] && clRefs[kind][list]) || []).map(function (x) { return typeof x === 'string' ? { key: x, name: x } : x; });
}

function refOptions(items, value, emptyLabel) {
  var opts = emptyLabel != null ? [{ value: '', label: emptyLabel }] : [];
  items.forEach(function (r) { opts.push({ value: r.key, label: (r.name && r.name !== r.key ? r.name + ' · ' : '') + r.key }); });
  if (value && !items.some(function (r) { return r.key === value; })) opts.push({ value: value, label: value + ' (' + t('missing') + ')' });
  return opts;
}

function refSelect(items, value, name, emptyLabel) {
  var el = select(value || '', refOptions(items, value, emptyLabel), { name: name, label: name });
  el.setAttribute('dir', 'auto');
  return el;
}

var clListSeq = 0;
function eventInput(items, value, name, placeholder) {
  var el = input(value || '', { name: name, ltr: true, mono: true, placeholder: placeholder });
  var id = 'cl-list-' + (++clListSeq);
  var list = h('datalist', { id: id });
  items.forEach(function (e) { list.appendChild(h('option', { value: e.key, text: e.name || e.key })); });
  el.setAttribute('list', id);
  var wrap = h('div', {}, el, list);
  wrap.input = el;
  return wrap;
}

function grid() {
  return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments));
}

function grid3() {
  return h('div', { class: 'grid sm:grid-cols-3 gap-x-4' }, Array.prototype.slice.call(arguments));
}

function opt(prefix, list) {
  return list.map(function (v) { return { value: v, label: t(prefix + v) }; });
}

function tagsInput(list, name, placeholder) {
  var el = input((list || []).join('، '), { name: name, placeholder: placeholder || t('hint.tags'), ltr: true });
  el.read = function () { return el.value.split(/[,،\n]/).map(function (s) { return s.trim(); }).filter(Boolean); };
  return el;
}

function jsonInput(value, name) {
  var el = input(JSON.stringify(value || {}), { name: name, ltr: true, mono: true, multiline: true, rows: 2 });
  el.read = function () { try { var v = JSON.parse(el.value.trim() || '{}'); return v && typeof v === 'object' && !Array.isArray(v) ? v : undefined; } catch (e) { return undefined; } };
  return el;
}

function intOr(el, d) {
  var n = numberValue(el);
  return n == null || isNaN(n) ? d : n;
}

function colorDot(color) {
  var dot = h('span', { class: 'cl-dot' });
  if (color) dot.style.background = color;
  return dot;
}

function nameCell(title, key, icon, color) {
  return h('div', { class: 'flex items-center gap-2 min-w-0' }, icon ? h('img', { src: icon, alt: '', class: 'cl-icon' }) : colorDot(color),
    h('div', { class: 'flex flex-col min-w-0' }, h('span', { class: 'font-medium', dir: 'auto', text: title }), key ? mono(key) : null));
}

function dl(pairs, name) {
  var box = h('dl', { class: 'cl-dl text-sm', 'data-clan-dl': name || '' });
  pairs.forEach(function (p) {
    if (!p) return;
    box.appendChild(h('dt', { text: p[0] }));
    box.appendChild(h('dd', {}, p[1] == null || p[1] === '' ? t('none') : p[1]));
  });
  return box;
}

function deltaText(n) {
  var s = (n > 0 ? '+' : n < 0 ? '−' : '') + fmtNum(Math.abs(n));
  return h('span', { class: n > 0 ? 'cl-amount-in' : n < 0 ? 'cl-amount-out' : '', dir: 'ltr', text: s });
}

function searchBox(onchange) {
  var el = input('', { name: 'q', placeholder: t('search') });
  var timer = null;
  el.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { onchange(el.value.trim()); }, 300); });
  return el;
}

function loading() {
  return h('p', { class: 'text-sm text-secondary-500', text: t('loading') });
}

function emptyCard(text) {
  return card(h('p', { class: 'p-4 text-sm text-secondary-500', text: text }));
}

function colorField(value) {
  var color = input(value || '', { name: 'color', ltr: true, mono: true, placeholder: '#F5B301' });
  var picker = h('input', { type: 'color', class: 'cl-color', name: 'color-picker', 'aria-label': t('f.color'), value: /^#[0-9A-Fa-f]{6}$/.test(value || '') ? value : '#f5b301' });
  picker.addEventListener('input', function () { color.value = picker.value.toUpperCase(); });
  color.addEventListener('input', function () { if (/^#[0-9A-Fa-f]{6}$/.test(color.value.trim())) picker.value = color.value.trim(); });
  var wrap = h('div', { class: 'flex gap-2 items-center' }, picker, color);
  wrap.input = color;
  return wrap;
}

function isUUID(s) {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(s || '');
}

function localInput(v, name) {
  var el = h('input', { type: 'datetime-local', class: CL_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  if (v) {
    var d = new Date(v);
    if (!isNaN(d)) el.value = new Date(d.getTime() - d.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
  }
  el.read = function () { if (!el.value) return null; var d = new Date(el.value); return isNaN(d) ? null : d.toISOString(); };
  return el;
}

function group(label, control, hint) {
  return h('div', { class: 'cl-field block mb-4', role: 'group', 'aria-label': label },
    h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1', text: label }), control,
    hint ? h('span', { class: 'block text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: hint }) : null);
}
