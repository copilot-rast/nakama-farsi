var CU_API = '/v2/console/base';
var CU_INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 px-3 py-2 text-sm';
var CU_BTN = 'cu-btn inline-flex items-center justify-center gap-1.5 font-medium whitespace-nowrap rounded-md border text-sm py-1.5 px-3';

function h(tag, props) {
  var el = document.createElement(tag);
  if (props) {
    for (var k in props) {
      var v = props[k];
      if (v === null || v === undefined || v === false) continue;
      if (k === 'class') el.setAttribute('class', v);
      else if (k === 'text') el.textContent = v;
      else if (k.indexOf('on') === 0 && typeof v === 'function') el.addEventListener(k.slice(2).toLowerCase(), v);
      else if (k === 'value') el.value = v;
      else if (k === 'checked') el.checked = !!v;
      else el.setAttribute(k, v === true ? '' : v);
    }
  }
  for (var i = 2; i < arguments.length; i++) add(el, arguments[i]);
  return el;
}

function add(el, child) {
  if (child === null || child === undefined || child === false) return;
  if (Array.isArray(child)) { child.forEach(function (c) { add(el, c); }); return; }
  el.appendChild(typeof child === 'object' ? child : document.createTextNode(String(child)));
}

function clear(el) {
  while (el.firstChild) el.removeChild(el.firstChild);
  return el;
}

function token() {
  var stores = [];
  try { stores.push(window.sessionStorage); } catch (e) { stores.push(null); }
  try { stores.push(window.localStorage); } catch (e) { stores.push(null); }
  for (var i = 0; i < stores.length; i++) {
    try {
      if (!stores[i]) continue;
      var raw = stores[i].getItem('SESSION_TOKENS');
      if (!raw) continue;
      var j = JSON.parse(raw);
      if (j && j.accessToken) return j.accessToken;
    } catch (e) { continue; }
  }
  return null;
}

function CuError(status, code, message, field, meta) {
  this.meta = meta || null;
  this.status = status;
  this.code = code;
  this.message = message;
  this.field = field;
}

function api(method, path, body) {
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  var init = { method: method, headers: headers, credentials: 'same-origin' };
  if (body !== undefined) { headers['Content-Type'] = 'application/json'; init.body = JSON.stringify(body); }
  return fetch((path.indexOf('/v2/') === 0 ? '' : CU_API) + path, init).then(function (res) {
    if (res.status === 401 && !/^#\/login/.test(location.hash)) location.hash = '#/login';
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) throw new CuError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', data && data.field || '', data && data.meta);
      return data;
    });
  });
}

function qs(params) {
  var out = [];
  Object.keys(params || {}).forEach(function (k) {
    var v = params[k];
    if (v === undefined || v === null || v === '') return;
    out.push(encodeURIComponent(k) + '=' + encodeURIComponent(v));
  });
  return out.length ? '?' + out.join('&') : '';
}

function button(opts) {
  var b = h('button', { type: opts.type || 'button', class: CU_BTN + ' cu-btn--' + (opts.variant || 'ghost') + (opts.class ? ' ' + opts.class : ''), disabled: !!opts.disabled, title: opts.title, 'aria-label': opts.title, onclick: opts.onclick },
    h('span', { text: opts.label }));
  if (opts.data) Object.keys(opts.data).forEach(function (k) { b.setAttribute('data-' + k, opts.data[k]); });
  return b;
}

function busy(btn, on) {
  btn.disabled = !!on;
  btn.classList.toggle('opacity-60', !!on);
}

function notice(kind, text) {
  var color = kind === 'error' ? 'error' : kind === 'success' ? 'success' : kind === 'warning' ? 'warning' : 'secondary';
  return h('div', { class: 'eco-notice eco-notice--' + color + ' flex items-center gap-2 rounded-md px-3 py-2 text-sm', role: kind === 'error' ? 'alert' : 'status' }, h('span', { class: 'flex-1', text: text }));
}

function tag(text, color) {
  return h('span', { class: 'eco-tag eco-tag--' + (color || 'secondary') + ' inline-flex leading-tight px-2 py-0.5 text-xs rounded-md whitespace-nowrap', text: text });
}

function mono(text) {
  return h('span', { class: 'font-mono text-xs', dir: 'ltr', text: text == null ? '' : String(text) });
}

var cuToastBox = null;
function toast(kind, text) {
  if (!cuToastBox || !cuToastBox.isConnected) {
    cuToastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-cup': 'toasts' });
    document.body.appendChild(cuToastBox);
  }
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  cuToastBox.appendChild(el);
  setTimeout(function () { el.remove(); }, kind === 'error' ? 6000 : 3000);
}

function errorText(err) {
  if (!err) return '';
  var code = err.code || '';
  return (err.message && err.message !== code ? err.message + ' ' : '') + (code ? '(' + code + ')' : '');
}

function card(children, cls) {
  return h('div', { class: 'bg-white dark:bg-secondary-800 rounded-md shadow' + (cls ? ' ' + cls : '') }, children);
}

function field(label, control, hint, extra) {
  var err = h('p', { class: 'cu-field-error eco-tf-err text-xs text-error-600 dark:text-error-400', hidden: true });
  var wrap = h('label', { class: 'cu-field eco-tbf block mb-4' + (extra ? ' ' + extra : '') },
    label ? h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 eco-tf-label', text: label }) : null,
    control,
    hint ? h('span', { class: 'block text-xs text-secondary-500 dark:text-secondary-400 eco-tf-hint', text: hint }) : null,
    err);
  wrap.setError = function (msg) { err.textContent = msg || ''; err.hidden = !msg; };
  if (label && control && control.tagName === 'INPUT' && control.getAttribute('placeholder') === label) wrap.classList.add('eco-tbf-q');
  return wrap;
}

function input(value, opts) {
  opts = opts || {};
  var el = h(opts.multiline ? 'textarea' : 'input', { class: CU_INPUT + (opts.mono ? ' font-mono' : ''), type: opts.multiline ? null : (opts.type || 'text'), dir: opts.ltr ? 'ltr' : 'auto', rows: opts.multiline ? (opts.rows || 3) : null, placeholder: opts.placeholder, name: opts.name, autocomplete: 'off', 'aria-label': opts.label });
  el.value = value == null ? '' : value;
  return el;
}

function select(value, options, opts) {
  opts = opts || {};
  var el = h('select', { class: CU_INPUT + ' eco-select', name: opts.name, 'aria-label': opts.label });
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) opt.selected = true;
    el.appendChild(opt);
  });
  return el;
}

function check(checked, label, name) {
  var box = h('input', { type: 'checkbox', class: 'cu-check', name: name, checked: !!checked });
  var wrap = h('label', { class: 'cu-checkline flex items-start gap-2 mb-3 cursor-pointer text-sm text-secondary-800 dark:text-secondary-200' }, box, h('span', { text: label }));
  wrap.input = box;
  return wrap;
}

function numberValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? Math.round(n) : NaN;
}

var cuPanel = null;
function closePanel() {
  if (cuPanel) { cuPanel.remove(); cuPanel = null; document.removeEventListener('keydown', panelKey); }
}

function panelKey(e) { if (e.key === 'Escape') closePanel(); }

function sidePanel(title, body, footer, name) {
  closePanel();
  var holder = h('div', { class: 'cu-panel-holder' });
  var panel = h('div', { class: 'cu-panel bg-white dark:bg-secondary-900 shadow-lg', role: 'dialog', 'aria-modal': 'true', 'aria-label': title, 'data-cup-panel': name || '' },
    h('div', { class: 'cu-panel-head flex items-center justify-between border-b px-6 py-3' },
      h('h2', { class: 'text-lg font-semibold truncate', text: title }),
      button({ label: '×', title: t('close'), onclick: closePanel, class: 'cu-close' })),
    h('div', { class: 'cu-panel-body p-6' }, body),
    footer ? h('div', { class: 'cu-panel-foot flex gap-2 justify-end bg-secondary-50 dark:bg-secondary-800 p-4' }, footer) : null);
  holder.appendChild(panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) closePanel(); });
  cuPanel = h('div', { 'data-cup': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, h('div', { class: 'cu-shade' }), holder);
  document.body.appendChild(cuPanel);
  document.addEventListener('keydown', panelKey);
  return panel;
}

function confirmBox(text, okLabel) {
  return new Promise(function (resolve) {
    var root;
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-cup': 'confirm', 'data-l10n-skip': '', class: 'cu-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'cu-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('confirm.title') }),
        h('p', { class: 'text-sm mb-4', text: text }),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: okLabel || t('yes'), variant: 'danger', data: { 'cup-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root);
  });
}

function pageFrame(page, desc, actions) {
  var content = h('div', { class: 'flex flex-col gap-4' });
  var el = h('div', { class: 'cu-page flex flex-col flex-1 px-4 md:px-6 pb-6 max-w-[1600px] w-full mx-auto', 'data-cup-page': page },
    h('div', { class: 'flex flex-col sm:flex-row gap-3 justify-between items-start mb-4' },
      h('div', {},
        h('h1', { class: 'text-xl font-semibold', text: t('nav.' + page) }),
        h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mt-1 max-w-3xl', text: desc })),
      h('div', { class: 'flex flex-wrap gap-2' }, actions || null)),
    content);
  return { el: el, content: content };
}

function table(columns, rows, onrow) {
  var thead = h('thead', {}, h('tr', {}, columns.map(function (c) { return h('th', { class: 'p-2.5 text-start text-sm font-medium', text: c.label }); })));
  var tbody = h('tbody', {});
  rows.forEach(function (row) {
    var tr = h('tr', { class: 'border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' + (onrow ? ' cursor-pointer' : ''), 'data-cup-row': row.id || '' });
    if (onrow) {
      tr.tabIndex = 0;
      tr.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,select,label')) onrow(row.item); });
      tr.addEventListener('keydown', function (e) { if (e.key === 'Enter') onrow(row.item); });
    }
    row.cells.forEach(function (cell, i) { tr.appendChild(h('td', { class: 'px-3 py-2 align-middle text-sm', 'data-label': columns[i].label }, cell)); });
    tbody.appendChild(tr);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow rounded-md overflow-x-auto' }, h('table', { class: 'w-full eco-table cu-table' }, thead, tbody));
}

function pager(state, onpage) {
  var total = state.total || 0;
  var nav = h('nav', { class: 'eco-pager', 'aria-label': t('pager') });
  if (!total) { nav.hidden = true; return nav; }
  var from = (state.page - 1) * state.limit + 1;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  function btn(name, path, label, disabled, go) {
    var b = h('button', { type: 'button', class: 'eco-pager-btn', title: label, 'aria-label': label, disabled: !!disabled, onclick: go, 'data-page': name });
    b.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4" aria-hidden="true" data-l10n-icon="mirror">' + path + '</svg>';
    return b;
  }
  nav.appendChild(h('span', { class: 'eco-pager-range tabular-nums', 'data-cup-range': '', text: t('range', { from: from, to: to, total: total }) }));
  if (last > 1) {
    nav.appendChild(h('div', { class: 'eco-pager-nav' },
      btn('first', '<path d="M7 6v12M17 18l-6-6 6-6"/>', t('page.first'), state.page <= 1, function () { onpage(1); }),
      btn('prev', '<path d="M15 18l-6-6 6-6"/>', t('prev'), state.page <= 1, function () { onpage(state.page - 1); }),
      h('span', { class: 'eco-pager-page tabular-nums', 'aria-current': 'page', text: t('page.of', { page: state.page, pages: last }) }),
      btn('next', '<path d="M9 18l6-6-6-6"/>', t('next'), state.page >= last, function () { onpage(state.page + 1); }),
      btn('last', '<path d="M17 6v12M7 18l6-6-6-6"/>', t('page.last'), state.page >= last, function () { onpage(last); })));
  }
  return nav;
}

function loadSettings(base) {
  return api('GET', base + '/settings');
}

function floatValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).replace(/٫/g, '.').trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? n : NaN;
}

var cuLanguages = null;
function loadLanguages() {
  if (cuLanguages) return Promise.resolve(cuLanguages);
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch('/v2/console/base/languages', { headers: headers, credentials: 'same-origin' }).then(function (r) { return r.ok ? r.json() : null; }).then(function (d) {
    var list = d && d.items ? d.items.filter(function (l) { return l.isActive; }).map(function (l) { return { code: l.code, name: l.nativeName || l.code, isDefault: !!l.isDefault, direction: l.direction }; }) : [];
    if (!list.length) list = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    list.sort(function (a, b) { return (b.isDefault ? 1 : 0) - (a.isDefault ? 1 : 0); });
    cuLanguages = list;
    return list;
  }, function () {
    cuLanguages = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    return cuLanguages;
  });
}

function translated(label, name, value, opts) {
  opts = opts || {};
  var inputs = {};
  var box = h('div', { class: 'grid sm:grid-cols-2 gap-x-4', 'data-cup-text': name });
  (cuLanguages || []).forEach(function (l) {
    var el = input(value && value[l.code] || '', { name: name + '-' + l.code, multiline: opts.multiline, rows: 2 });
    el.setAttribute('dir', l.direction === 'rtl' ? 'rtl' : 'ltr');
    inputs[l.code] = el;
    box.appendChild(field(t('lang.text', { label: label, lang: l.name }), el));
  });
  box.read = function () {
    var out = {};
    Object.keys(inputs).forEach(function (k) { var v = inputs[k].value.trim(); if (v) out[k] = v; });
    return out;
  };
  return box;
}

function textOf(m, fallback) {
  if (!m) return fallback || '';
  var l = cuLang();
  return m[l] || m.fa || m.en || Object.keys(m).map(function (k) { return m[k]; }).filter(Boolean)[0] || fallback || '';
}

function fieldError(msgBox, err) {
  clear(msgBox).appendChild(notice('error', cuErrText(err)));
}

function options(list, value) {
  return list.map(function (o) { return typeof o === 'string' ? { value: o, label: o } : o; });
}

function dateInput(value, name, type) {
  var el = h('input', { type: type || 'date', class: CU_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  el.value = value || '';
  return el;
}

function isoDay(d) {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

function cuErrText(err) {
  if (err && err.code && has('code.' + err.code)) return t('code.' + err.code, err.meta || null) + (err.field ? ' · ' + err.field : '');
  return errorText(err) + (err && err.field ? ' · ' + err.field : '');
}

function section(title, children, name) {
  return card(h('div', { class: 'p-4', 'data-cup-section': name || '' },
    h('h2', { class: 'font-semibold mb-3', text: title }), children));
}

function settingsPage(root, page, base, build) {
  var frame = pageFrame(page, t('desc.' + page));
  clear(root).appendChild(frame.el);
  frame.content.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  Promise.all([loadSettings(base), loadLanguages()]).then(function (res) {
    var d = res[0];
    var cur = JSON.parse(JSON.stringify(d.settings));
    var form = build(cur, d.defaults, d);
    var msg = h('div', { 'data-cup-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', 'data-cup-updated': '' });
    var paintMeta = function (x) { meta.textContent = x.updatedBy ? t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }) : ''; };
    paintMeta(d);
    var save = button({ label: t('save'), variant: 'primary', data: { 'cup-save': page }, onclick: function () {
      busy(save, true);
      clear(msg);
      var bad = form.read(cur);
      if (bad) { busy(save, false); msg.appendChild(notice('error', bad)); return; }
      api('PUT', base + '/settings', cur).then(function (out) {
        busy(save, false);
        cur = JSON.parse(JSON.stringify(out.settings));
        paintMeta(out);
        toast('success', t('saved'));
        if (form.after) form.after(out);
      }, function (err) {
        busy(save, false);
        msg.appendChild(notice('error', cuErrText(err)));
      });
    } });
    clear(frame.content);
    form.el.forEach(function (el) { frame.content.appendChild(el); });
    frame.content.appendChild(h('div', { class: 'cu-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), save));
  }, function (err) {
    clear(frame.content).appendChild(notice('error', cuErrText(err)));
  });
}

function numInput(value, name, label) {
  var el = input(value == null ? '' : digits(String(value)), { ltr: true, name: name, label: label });
  el.setAttribute('inputmode', 'numeric');
  return el;
}

function hashQuery() {
  var i = location.hash.indexOf('?');
  var out = {};
  if (i === -1) return out;
  location.hash.slice(i + 1).split('&').forEach(function (kv) {
    var p = kv.split('=');
    if (p[0]) out[decodeURIComponent(p[0])] = decodeURIComponent(p[1] || '');
  });
  return out;
}

function playerLink(id, name) {
  if (!id) return h('span', { class: 'text-secondary-400', text: '—' });
  return h('a', { href: '#/players/' + id, class: 'eco-link', 'data-cup-player-link': id, dir: 'auto', text: name || id.slice(0, 8) });
}
