var DK_API = '/v2/console/base/devkit';
var DK_INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 px-3 py-2 text-sm';
var DK_BTN = 'dk-btn inline-flex items-center justify-center gap-1.5 font-medium whitespace-nowrap rounded-md border text-sm py-1.5 px-3';

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

function DkError(status, code, message, field) {
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
  return fetch(DK_API + path, init).then(function (res) {
    if (res.status === 401 && !/^#\/login/.test(location.hash)) location.hash = '#/login';
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) throw new DkError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', data && data.field || '');
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
  var b = h('button', { type: opts.type || 'button', class: DK_BTN + ' dk-btn--' + (opts.variant || 'ghost') + (opts.class ? ' ' + opts.class : ''), disabled: !!opts.disabled, title: opts.title, 'aria-label': opts.title, onclick: opts.onclick },
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

var dkToastBox = null;
function toast(kind, text) {
  if (!dkToastBox || !dkToastBox.isConnected) {
    dkToastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-devkit': 'toasts' });
    document.body.appendChild(dkToastBox);
  }
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  dkToastBox.appendChild(el);
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
  var err = h('p', { class: 'dk-field-error eco-tf-err text-xs text-error-600 dark:text-error-400', hidden: true });
  var wrap = h('label', { class: 'dk-field eco-tbf block mb-4' + (extra ? ' ' + extra : '') },
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
  var el = h(opts.multiline ? 'textarea' : 'input', { class: DK_INPUT + (opts.mono ? ' font-mono' : ''), type: opts.multiline ? null : (opts.type || 'text'), dir: opts.ltr ? 'ltr' : 'auto', rows: opts.multiline ? (opts.rows || 3) : null, placeholder: opts.placeholder, name: opts.name, autocomplete: 'off', 'aria-label': opts.label });
  el.value = value == null ? '' : value;
  return el;
}

function select(value, options, opts) {
  opts = opts || {};
  var el = h('select', { class: DK_INPUT + ' eco-select', name: opts.name, 'aria-label': opts.label });
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) opt.selected = true;
    el.appendChild(opt);
  });
  return el;
}

function check(checked, label, name) {
  var box = h('input', { type: 'checkbox', class: 'dk-check', name: name, checked: !!checked });
  var wrap = h('label', { class: 'dk-checkline flex items-start gap-2 mb-3 cursor-pointer text-sm text-secondary-800 dark:text-secondary-200' }, box, h('span', { text: label }));
  wrap.input = box;
  return wrap;
}

function numberValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? Math.round(n) : NaN;
}

var dkPanel = null;
function closePanel() {
  if (dkPanel) { dkPanel.remove(); dkPanel = null; document.removeEventListener('keydown', panelKey); }
}

function panelKey(e) { if (e.key === 'Escape') closePanel(); }

function sidePanel(title, body, footer, name) {
  closePanel();
  var holder = h('div', { class: 'dk-panel-holder' });
  var panel = h('div', { class: 'dk-panel bg-white dark:bg-secondary-900 shadow-lg', role: 'dialog', 'aria-modal': 'true', 'aria-label': title, 'data-devkit-panel': name || '' },
    h('div', { class: 'dk-panel-head flex items-center justify-between border-b px-6 py-3' },
      h('h2', { class: 'text-lg font-semibold truncate', text: title }),
      button({ label: '×', title: t('close'), onclick: closePanel, class: 'dk-close' })),
    h('div', { class: 'dk-panel-body p-6' }, body),
    footer ? h('div', { class: 'dk-panel-foot flex gap-2 justify-end bg-secondary-50 dark:bg-secondary-800 p-4' }, footer) : null);
  holder.appendChild(panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) closePanel(); });
  dkPanel = h('div', { 'data-devkit': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, h('div', { class: 'dk-shade' }), holder);
  document.body.appendChild(dkPanel);
  document.addEventListener('keydown', panelKey);
  return panel;
}

function confirmBox(text, okLabel) {
  return new Promise(function (resolve) {
    var root;
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-devkit': 'confirm', 'data-l10n-skip': '', class: 'dk-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'dk-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('confirm.title') }),
        h('p', { class: 'text-sm mb-4', text: text }),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: okLabel || t('yes'), variant: 'danger', data: { 'devkit-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root);
  });
}

function pageFrame(page, desc, actions) {
  var content = h('div', { class: 'flex flex-col gap-4' });
  var el = h('div', { class: 'dk-page flex flex-col flex-1 px-4 md:px-6 pb-6 max-w-[1600px] w-full mx-auto', 'data-devkit-page': page },
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
    var tr = h('tr', { class: 'border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' + (onrow ? ' cursor-pointer' : ''), 'data-devkit-row': row.id || '' });
    if (onrow) {
      tr.tabIndex = 0;
      tr.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,select,label')) onrow(row.item); });
      tr.addEventListener('keydown', function (e) { if (e.key === 'Enter') onrow(row.item); });
    }
    row.cells.forEach(function (cell, i) { tr.appendChild(h('td', { class: 'px-3 py-2 align-middle text-sm', 'data-label': columns[i].label }, cell)); });
    tbody.appendChild(tr);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow rounded-md overflow-x-auto' }, h('table', { class: 'w-full eco-table dk-table' }, thead, tbody));
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
  nav.appendChild(h('span', { class: 'eco-pager-range tabular-nums', 'data-devkit-range': '', text: t('range', { from: from, to: to, total: total }) }));
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

var dkSettings = null;
function loadSettings() {
  return api('GET', '/settings').then(function (d) { dkSettings = d; return d; });
}

function settingsCard(section, build) {
  var open = false;
  var body = h('div', { class: 'p-4 border-t border-secondary-200 dark:border-secondary-700', hidden: true });
  var head = h('button', { type: 'button', class: 'dk-settings-toggle w-full flex items-center justify-between px-4 py-3 text-sm font-medium', 'aria-expanded': 'false', 'data-devkit-settings-toggle': section },
    h('span', { text: t('set.' + section) }), h('span', { class: 'dk-caret', text: '▾' }));
  var box = card([head, body], 'dk-settings');
  box.setAttribute('data-devkit-settings', section);
  head.addEventListener('click', function () {
    open = !open;
    head.setAttribute('aria-expanded', open ? 'true' : 'false');
    body.hidden = !open;
    if (open && !body.firstChild) paint();
  });
  function paint() {
    clear(body).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    loadSettings().then(function (d) {
      var cur = JSON.parse(JSON.stringify(d.settings));
      var fields = build(cur[section], d.defaults[section]);
      var msg = h('div', {});
      var save = button({ label: t('save'), variant: 'primary', data: { 'devkit-settings-save': section }, onclick: function () {
        busy(save, true);
        clear(msg);
        cur[section] = fields.read();
        api('PUT', '/settings', cur).then(function (res) {
          busy(save, false);
          dkSettings = res;
          toast('success', t('saved'));
          if (fields.after) fields.after(res);
        }, function (err) {
          busy(save, false);
          msg.appendChild(notice('error', errorText(err)));
        });
      } });
      clear(body).appendChild(h('div', {}, fields.el, msg,
        d.settings.updatedBy ? h('p', { class: 'text-xs text-secondary-500 mt-2', text: d.settings.updatedBy + ' · ' + fmtDate(d.settings.updatedAt) }) : null,
        h('div', { class: 'flex justify-end mt-3' }, save)));
    }, function (err) { clear(body).appendChild(notice('error', errorText(err))); });
  }
  return box;
}
