var BTN = 'x-button eco-btn relative transition duration-150 focus-visible:outline-secondary-300 outline-transparent outline outline-1 outline-offset-2 inline-flex items-center justify-center font-medium whitespace-nowrap overflow-hidden align-middle active:!shadow-none border appearance-none shrink-0 rounded-md';
var BTN_STYLE = {
  primary: '--x-button-text: #fff; --x-button-border: rgb(var(--primary-color-600)); --x-button-dark-text: #fff; --x-button-dark-border: rgb(var(--primary-color-500)); --x-button-bg: rgb(var(--primary-color-500)); --x-button-bg-hover: rgb(var(--primary-color-600)); --x-button-bg-active: rgb(var(--primary-color-800)); --x-button-dark-bg: rgb(var(--primary-color-600)); --x-button-dark-bg-hover: rgb(var(--primary-color-700)); --x-button-dark-bg-active: rgb(var(--primary-color-800));',
  ghost: '--x-button-border: transparent; --x-button-text: #1f2937; --x-button-dark-text: white; --x-button-bg-hover: #f3f4f6; --x-button-bg-active: #e5e7eb; --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  muted: '--x-button-border: transparent; --x-button-text: #475569; --x-button-dark-text: #64748b; --x-button-bg-hover: #f8fafc; --x-button-bg-active: #f1f5f9; --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  link: '--x-button-border: transparent; --x-button-text: rgb(var(--primary-color-600)); --x-button-dark-text: rgb(var(--primary-color-400)); --x-button-bg-hover: rgb(var(--primary-color-50)); --x-button-bg-active: rgb(var(--primary-color-100)); --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  danger: '--x-button-border: transparent; --x-button-text: rgb(var(--error-color-600)); --x-button-dark-text: rgb(var(--error-color-400)); --x-button-bg-hover: rgb(var(--error-color-50)); --x-button-bg-active: rgb(var(--error-color-100)); --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  dangerSolid: '--x-button-text: #fff; --x-button-border: rgb(var(--error-color-600)); --x-button-bg: rgb(var(--error-color-600)); --x-button-bg-hover: rgb(var(--error-color-700)); --x-button-bg-active: rgb(var(--error-color-800)); --x-button-dark-text: #fff; --x-button-dark-border: rgb(var(--error-color-600)); --x-button-dark-bg: rgb(var(--error-color-600)); --x-button-dark-bg-hover: rgb(var(--error-color-500));',
  disabled: '--x-button-bg: #f9fafb; --x-button-text: #d1d5db; --x-button-border: #e5e7eb; --x-button-dark-bg: #111827; --x-button-dark-text: #4b5563; --x-button-dark-border: #374151;'
};
var BTN_SIZE = {
  page: 'shadow-sm py-1.5 text-sm px-3',
  footer: 'py-2 px-4',
  sm: 'py-1.5 text-sm px-3',
  icon: 'leading-none p-1',
  iconLg: 'leading-none p-1.5'
};

function button(opts) {
  var variant = opts.disabled ? 'disabled' : (opts.variant || 'ghost');
  var size = BTN_SIZE[opts.size || 'footer'];
  if (variant === 'primary' && opts.size === 'footer') size += ' shadow-sm';
  var b = h('button', {
    type: opts.type || 'button',
    class: BTN + ' ' + size + ' ' + (opts.disabled ? 'cursor-not-allowed' : 'cursor-pointer') + (opts.class ? ' ' + opts.class : ''),
    style: BTN_STYLE[variant],
    disabled: !!opts.disabled,
    title: opts.title,
    'aria-label': opts.title,
    onclick: opts.onclick
  });
  var inner = h('div', { class: 'inline-flex items-center justify-center' });
  if (opts.icon) inner.appendChild(icon(opts.icon, (opts.iconClass || 'h-4 w-4') + (opts.label ? ' mr-2' : ' m-0.5')));
  if (opts.label) inner.appendChild(h('span', { text: opts.label }));
  b.appendChild(inner);
  return b;
}

function setBusy(btn, busy, label) {
  btn.disabled = !!busy;
  var span = btn.querySelector('span');
  if (busy) {
    btn.dataset.label = span ? span.textContent : '';
    if (span && label) span.textContent = label;
    btn.classList.add('opacity-60');
  } else {
    if (span && btn.dataset.label) span.textContent = btn.dataset.label;
    btn.classList.remove('opacity-60');
  }
}

var INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 outline-transparent outline outline-2 outline-offset-[-1px] transition-all duration-150 ease-in-out border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md hover:border-secondary-400 dark:hover:border-secondary-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 focus:outline-[color:var(--x-input-border)]';
var INPUT_SM = INPUT + ' px-2 py-1.5 text-sm';
var INPUT_MD = INPUT + ' px-3 py-2';
var INPUT_VARS = '--x-input-border: rgb(var(--primary-color-400));';
var SELECT_MD = INPUT_MD + ' eco-select';
var SELECT_SM = INPUT_SM + ' eco-select';

function searchInput(value, placeholder, oninput) {
  var input = h('input', {
    class: INPUT_SM + ' !pl-10 eco-search-input',
    type: 'search',
    value: value || '',
    placeholder: placeholder,
    'aria-label': placeholder,
    autocomplete: 'off',
    spellcheck: 'false'
  });
  input.addEventListener('input', oninput);
  input.addEventListener('keydown', function (e) {
    if (e.key === 'Escape' && input.value) {
      e.preventDefault();
      e.stopPropagation();
      input.value = '';
      input.dispatchEvent(new Event('input'));
    }
  });
  return h('label', { class: 'x-label block relative align-bottom focus:outline-none x-input eco-search', style: INPUT_VARS },
    h('div', { class: 'relative' },
      icon('search', 'h-4 w-4 ml-2 left-1 text-secondary-600 dark:text-secondary-300 absolute my-auto inset-y-0 eco-search-ic'),
      input));
}

function filterBox(label, control, on) {
  var box = h('label', { class: 'eco-filter' + (on ? ' eco-filter-on' : '') },
    h('span', { class: 'eco-filter-label', text: label }),
    control);
  box.mark = function (v) { box.classList.toggle('eco-filter-on', !!v); };
  return box;
}

function filterSelect(label, value, options, onchange, cfg) {
  cfg = cfg || {};
  var def = cfg.def == null ? '' : String(cfg.def);
  var sel = h('select', { class: 'eco-filter-select', 'aria-label': label, onchange: function () { box.mark(sel.value !== def); onchange(sel.value); } });
  if (cfg.def == null) sel.appendChild(h('option', { value: '', text: t('all') }));
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value || '')) opt.selected = true;
    sel.appendChild(opt);
  });
  if (cfg.def != null && !sel.value) sel.value = def;
  var box = filterBox(label, sel, sel.value !== def);
  box.select = sel;
  return box;
}

function filterText(label, value, onchange, opts) {
  opts = opts || {};
  var inp = h('input', { class: 'eco-filter-input eco-filter-text' + (opts.mono === false ? '' : ' font-mono'), dir: 'auto', type: 'text', value: value || '', placeholder: opts.placeholder || t('all'), 'aria-label': label, autocomplete: 'off', spellcheck: 'false' });
  var box = filterBox(label, inp, !!value);
  inp.addEventListener('input', function () { box.mark(inp.value.trim()); });
  inp.addEventListener('input', debounce(function () { onchange(toLatinDigits(inp.value).trim()); }, opts.wait == null ? 300 : opts.wait));
  box.input = inp;
  return box;
}

function clearFiltersButton(onclick) {
  var b = button({ variant: 'muted', size: 'sm', icon: 'close', iconClass: 'h-3.5 w-3.5', label: t('filters.clear'), class: 'eco-filters-clear', onclick: onclick });
  b.setAttribute('data-filters-clear', '');
  return b;
}

function refreshButton(onclick) {
  var b = button({ variant: 'ghost', size: 'icon', icon: 'refresh', iconClass: 'h-4 w-4', title: t('refresh'), class: 'eco-iconbtn', onclick: function () {
    b.classList.remove('eco-spin');
    void b.offsetWidth;
    b.classList.add('eco-spin');
    onclick();
  } });
  b.setAttribute('data-list-refresh', '');
  return b;
}

function listCount(state) {
  var el = h('span', { class: 'eco-list-count tabular-nums', 'aria-live': 'polite', 'data-list-count': '' });
  if (state.loading && !state.loaded) el.appendChild(bar('3.5rem'));
  else if (state.total) el.textContent = t(state.total === 1 ? 'list.count.one' : 'list.count', { n: state.total });
  return el;
}

function listToolbar(parts) {
  var bar2 = h('div', { class: 'eco-toolbar' + ((parts.filters || []).length > 3 ? ' eco-toolbar--stack' : ''), role: 'search', 'aria-label': t('list.toolbar') });
  if (parts.search) bar2.appendChild(parts.search);
  var filters = h('div', { class: 'eco-filters' });
  (parts.filters || []).forEach(function (f) { filters.appendChild(f); });
  if (parts.clear) filters.appendChild(parts.clear);
  if ((parts.filters || []).length || parts.clear) bar2.appendChild(filters);
  var meta = h('div', { class: 'eco-toolbar-meta' });
  if (parts.meta) meta.appendChild(parts.meta);
  if (parts.refresh) meta.appendChild(parts.refresh);
  bar2.appendChild(meta);
  bar2.meta = meta;
  return bar2;
}

function toggle(checked, label, onchange, size) {
  var on = !!checked;
  var track = h('div', { class: '', style: '--x-toggle-bg: rgb(var(--primary-color-500)); --x-toggle-dark-bg: rgb(var(--primary-color-600));' });
  var knob = h('div', { class: '' });
  var input = h('input', { type: 'checkbox', class: 'hidden', role: 'switch' });
  var sm = size === 'sm';
  function paint() {
    track.className = 'flex items-center rounded-full transition-colors duration-300 border-[3px] shrink-0 !border-transparent ' + (on ? 'bg-[color:var(--x-toggle-bg)] dark:bg-[color:var(--x-toggle-dark-bg)]' : 'bg-secondary-300 dark:bg-secondary-500');
    knob.className = 'rounded-full shadow transform transition duration-150 shrink-0 ' + (sm ? 'h-3 w-3' : 'h-4 w-4') + ' bg-white dark:bg-secondary-800' + (on ? ' translate-x-full' : '');
    input.checked = on;
    input.setAttribute('aria-checked', on ? 'true' : 'false');
  }
  track.appendChild(h('div', { class: 'relative shrink-0 ' + (sm ? 'w-6' : 'w-8') }, input, knob));
  var wrap = h('label', { class: 'inline-block x-toggle cursor-pointer', tabindex: '0' },
    h('div', { class: 'flex items-center' }, track,
      label ? h('span', { class: 'ml-2 font-medium text-secondary-800 dark:text-secondary-200 ' + (sm ? 'text-xs' : 'text-sm'), text: label }) : null));
  function flip(e) {
    e.preventDefault();
    on = !on;
    paint();
    onchange(on);
  }
  wrap.addEventListener('click', flip);
  wrap.addEventListener('keydown', function (e) { if (e.key === ' ' || e.key === 'Enter') flip(e); });
  wrap.set = function (v) { on = !!v; paint(); };
  paint();
  return wrap;
}

function checkbox(checked, label, onchange) {
  var on = !!checked;
  var mixed = false;
  var box = h('div', {});
  var mark = document.createElementNS(SVG_NS, 'svg');
  mark.setAttribute('viewBox', '0 0 20 20');
  mark.innerHTML = '<path d="M0 11l2-2 5 5L18 3l2 2L7 18z"/>';
  box.appendChild(mark);
  var wrap = h('label', { class: 'x-checkbox inline-block relative cursor-pointer align-middle', tabindex: '0', role: 'checkbox' });
  function paint() {
    var fill = on || mixed;
    mark.innerHTML = mixed ? '<path d="M4 9h12v2H4z"/>' : '<path d="M0 11l2-2 5 5L18 3l2 2L7 18z"/>';
    wrap.setAttribute('style', fill
      ? '--x-checkbox-bg: rgb(var(--primary-color-500)); --x-checkbox-border: rgb(var(--primary-color-500)); --x-checkbox-dark-bg: rgb(var(--primary-color-500)); --x-checkbox-dark-border: rgb(var(--primary-color-500));'
      : '--x-checkbox-bg: #fff; --x-checkbox-border: #d1d5db; --x-checkbox-dark-bg: #111827; --x-checkbox-dark-border: #9ca3af;');
    box.setAttribute('class', 'rounded flex justify-center items-center shrink-0 border border-[color:var(--x-checkbox-border)] bg-[color:var(--x-checkbox-bg)] dark:border-[color:var(--x-checkbox-dark-border)] dark:bg-[color:var(--x-checkbox-dark-bg)] h-4 w-4');
    mark.setAttribute('class', 'fill-current text-white h-2.5 w-2.5' + (fill ? '' : ' opacity-0'));
    wrap.setAttribute('aria-checked', mixed ? 'mixed' : on ? 'true' : 'false');
  }
  wrap.appendChild(h('div', { class: 'flex items-center' }, box,
    label ? h('div', { class: 'inline-block text-secondary-800 dark:text-secondary-200 pl-2 font-medium' }, h('span', { text: label })) : null));
  function flip(e) { e.preventDefault(); on = mixed ? true : !on; mixed = false; paint(); onchange(on); }
  wrap.addEventListener('click', flip);
  wrap.addEventListener('keydown', function (e) { if (e.key === ' ' || e.key === 'Enter') flip(e); });
  wrap.set = function (v) { mixed = v === 'mixed'; on = v === true; paint(); };
  paint();
  return wrap;
}

function tag(text, color, extra) {
  return h('span', { class: 'eco-tag eco-tag--' + (color || 'secondary') + ' inline-flex leading-tight max-w-full px-2 py-0.5 text-xs rounded-md whitespace-nowrap' + (extra ? ' ' + extra : ''), text: text });
}

function mono(text) {
  return h('span', { class: 'font-mono text-xs', dir: 'ltr', text: text == null ? '' : String(text) });
}

function bar(width, extra) {
  return h('div', { class: 'animate-pulse bg-secondary-300 dark:bg-secondary-600 rounded-md eco-bar ' + (extra || 'h-3'), style: 'width:' + width });
}

var frameHub = null;

function pageFrame(title, desc, actions) {
  var hub = frameHub;
  frameHub = null;
  if (hub) title = hub.label;
  var content = h('div', { class: 'flex flex-col flex-1 min-w-0 px-4 md:px-6' });
  var actionsBox = h('div', { class: 'flex items-center gap-1.5 justify-end w-full sm:w-auto' }, actions || null);
  var frame = h('div', { class: 'flex flex-col flex-1' },
    h('div', { class: 'w-full min-w-0 flex flex-col flex-1 pb-4 max-w-[1600px] mx-auto overflow-x-auto' },
      h('div', { class: 'mb-4 relative px-4 md:px-6' },
        h('div', { class: 'flex flex-col sm:flex-row gap-4 justify-between items-start' },
          h('div', { class: 'flex items-center gap-1' },
            h('div', { class: 'flex gap-4 items-center' },
              h('div', {},
                h('div', { class: 'flex items-center gap-2 mb-0.5' },
                  h('h1', { class: 'text-xl font-semibold truncate max-w-xs md:max-w-sm xl:max-w-xl mr-2' }, h('span', { text: title }))),
                desc ? h('div', { class: 'flex items-center gap-2' },
                  h('div', { class: 'text-secondary-500 -mt-0.5 space-x-2' }, h('span', { text: desc }))) : null))),
          actionsBox),
        hub ? hub.el : null),
      content));
  return { el: frame, content: content, actions: actionsBox };
}

function tabFrame(desc, actions) {
  var content = h('div', { class: 'flex flex-col flex-1 min-w-0' });
  var actionsBox = h('div', { class: 'flex flex-wrap items-center gap-2' }, actions || null);
  var frame = h('div', { class: 'flex flex-col flex-1 min-w-0' },
    h('div', { class: 'flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4' },
      h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: desc }),
      actionsBox),
    content);
  return { el: frame, content: content, actions: actionsBox };
}

function notice(kind, text, action) {
  var color = kind === 'error' ? 'error' : kind === 'success' ? 'success' : kind === 'warning' ? 'warning' : 'secondary';
  return h('div', { class: 'eco-notice eco-notice--' + color + ' flex items-center gap-2 rounded-md px-3 py-2 text-sm', role: kind === 'error' ? 'alert' : 'status' },
    icon(kind === 'success' ? 'check' : 'alert', 'h-4 w-4'),
    h('span', { class: 'flex-1', text: text }),
    action || null);
}

function emptyBlock(title, hint, action) {
  return h('div', { class: 'flex flex-col items-center justify-center text-center gap-2 py-12 px-4' },
    h('div', { class: 'text-secondary-400 dark:text-secondary-500' }, icon('empty', 'h-8 w-8')),
    h('div', { class: 'font-medium text-secondary-800 dark:text-secondary-200', text: title }),
    hint ? h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: hint }) : null,
    action ? h('div', { class: 'mt-2' }, action) : null);
}

var toastBox = null;
function toast(kind, text) {
  if (!toastBox || !toastBox.isConnected) {
    toastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-economy': '' });
    document.body.appendChild(toastBox);
  }
  toastBox.setAttribute('dir', document.documentElement.getAttribute('dir') || 'ltr');
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  toastBox.appendChild(el);
  setTimeout(function () { el.remove(); }, kind === 'error' ? 6000 : 3000);
}

function card(children, extra) {
  return h('div', { class: 'bg-white dark:bg-secondary-800 overflow-hidden rounded-md shadow' + (extra ? ' ' + extra : '') }, children);
}

function pagerButton(name, title, disabled, go) {
  var b = h('button', { type: 'button', class: 'eco-pager-btn', title: title, 'aria-label': title, disabled: !!disabled, onclick: go });
  b.appendChild(icon(name, 'h-4 w-4'));
  b.setAttribute('data-page', name);
  return b;
}

function pager(state, onpage) {
  var total = state.total || 0;
  var nav = h('nav', { class: 'eco-pager', 'aria-label': t('list.pager') });
  if (state.loading && !state.loaded) { nav.appendChild(bar('8rem')); return nav; }
  if (!total) { nav.hidden = true; return nav; }
  var from = (state.page - 1) * state.limit + 1;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  nav.appendChild(h('span', { class: 'eco-pager-range tabular-nums', text: t('range', { from: from, to: to, total: total }) }));
  if (last > 1) {
    nav.appendChild(h('div', { class: 'eco-pager-nav' },
      pagerButton('first', t('page.first'), state.page <= 1, function () { onpage(1); }),
      pagerButton('prev', t('page.prev'), state.page <= 1, function () { onpage(state.page - 1); }),
      h('span', { class: 'eco-pager-page tabular-nums', 'aria-current': 'page', text: t('page.of', { page: state.page, pages: last }) }),
      pagerButton('next', t('page.next'), state.page >= last, function () { onpage(state.page + 1); }),
      pagerButton('last', t('page.last'), state.page >= last, function () { onpage(last); })));
  }
  return nav;
}

function selection(opts) {
  var sel = { ids: {}, page: [], idOf: opts.idOf || function (it) { return it.id; }, bar: null, heads: [] };
  sel.count = function () { return Object.keys(sel.ids).length; };
  sel.list = function () { return Object.keys(sel.ids); };
  sel.paint = function () {
    var n = sel.count();
    if (sel.bar) sel.bar.paint(n);
    var onPage = sel.page.filter(function (it) { return sel.ids[sel.idOf(it)]; }).length;
    sel.heads.forEach(function (hd) { if (hd.isConnected) hd.set(onPage && onPage === sel.page.length ? true : onPage ? 'mixed' : false); });
    sel.heads = sel.heads.filter(function (hd) { return hd.isConnected; });
    document.querySelectorAll('[data-economy-sel="' + opts.name + '"]').forEach(function (box) {
      var tr = box.closest('tr');
      if (tr) tr.classList.toggle('eco-row-on', !!sel.ids[box.getAttribute('data-id')]);
    });
  };
  sel.selectPage = function () { sel.page.forEach(function (it) { sel.ids[sel.idOf(it)] = true; }); sel.paint(); sel.sync(); };
  sel.clear = function () { sel.ids = {}; sel.paint(); sel.sync(); };
  sel.sync = function () { document.querySelectorAll('[data-economy-sel="' + opts.name + '"]').forEach(function (box) { box.set(!!sel.ids[box.getAttribute('data-id')]); }); };
  sel.rowCheck = function (it, label) {
    var id = sel.idOf(it);
    var box = checkbox(!!sel.ids[id], '', function (on) {
      if (on) sel.ids[id] = true;
      else delete sel.ids[id];
      sel.paint();
    });
    box.setAttribute('aria-label', label);
    box.setAttribute('data-economy-sel', opts.name);
    box.setAttribute('data-id', id);
    box.classList.add('eco-sel-box');
    return box;
  };
  sel.headCheck = function () {
    var box = checkbox(false, '', function (on) { if (on) sel.selectPage(); else { sel.page.forEach(function (it) { delete sel.ids[sel.idOf(it)]; }); sel.paint(); sel.sync(); } });
    box.setAttribute('aria-label', t('sel.all'));
    box.setAttribute('data-select-all', '');
    box.classList.add('eco-sel-box');
    sel.heads.push(box);
    setTimeout(sel.paint, 0);
    return box;
  };
  return sel;
}

function selectionBar(sel, opts) {
  var count = h('span', { class: 'eco-selbar-count tabular-nums', 'aria-live': 'polite' });
  if (opts.countAttr) count.setAttribute(opts.countAttr, '');
  var pageAll = button({ variant: 'link', size: 'sm', label: t('pl.bulk.page'), class: 'eco-selbar-link', onclick: function () { sel.selectPage(); } });
  var none = button({ variant: 'muted', size: 'icon', icon: 'close', iconClass: 'h-4 w-4', title: t('pl.bulk.clear'), class: 'eco-iconbtn', onclick: function () { sel.clear(); } });
  if (opts.pageAttr) pageAll.setAttribute(opts.pageAttr[0], opts.pageAttr[1]);
  none.setAttribute('data-sel-clear', '');
  var el = h('div', { class: 'eco-selbar', role: 'region', 'aria-label': t('sel.bar'), hidden: true },
    h('div', { class: 'eco-selbar-lead' }, none, count, pageAll),
    h('div', { class: 'eco-selbar-actions' }, opts.actions || []));
  if (opts.attr) el.setAttribute(opts.attr, '');
  el.paint = function (n) {
    var was = !el.hidden;
    el.hidden = !n;
    count.textContent = t(opts.countKey, { n: n });
    var pageIds = sel.page.map(sel.idOf);
    pageAll.hidden = !pageIds.length || pageIds.every(function (id) { return sel.ids[id]; });
    if (opts.onpaint) opts.onpaint(n);
    if (!was && n) { el.classList.remove('eco-selbar-in'); void el.offsetWidth; el.classList.add('eco-selbar-in'); }
  };
  sel.bar = el;
  el.paint(sel.count());
  return el;
}

function table(columns, rows, opts) {
  opts = opts || {};
  var thead = h('thead', { class: 'x-table-head align-bottom bg-secondary-100 dark:bg-secondary-700 sticky top-0 z-10' });
  var tr = h('tr', { class: 'text-sm text-secondary-700 dark:text-secondary-200 border-b' });
  columns.forEach(function (c) {
    tr.appendChild(h('th', { class: 'x-table-header relative p-2.5 font-medium' + (c.head ? ' eco-th-sel' : ''), width: c.width, scope: 'col', 'aria-label': c.head ? c.label : null },
      c.head ? c.head() : h('div', { class: 'flex items-center gap-1 select-none ' + (c.align === 'center' ? 'justify-center' : 'justify-start') }, h('div', { class: 'flex items-center gap-1', text: c.label }))));
  });
  thead.appendChild(tr);
  var tbody = h('tbody', {});
  rows.forEach(function (row) {
    var r = h('tr', { class: 'x-table-row border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700 align-top' + (opts.onrow ? ' cursor-pointer' : '') });
    if (opts.onrow) {
      r.tabIndex = 0;
      r.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,label')) opts.onrow(row.item); });
      r.addEventListener('keydown', function (e) { if (e.key === 'Enter') opts.onrow(row.item); });
    }
    columns.forEach(function (c, i) {
      var cell = row.cells[i];
      r.appendChild(h('td', { class: 'relative x-table-cell px-3 h-9 py-1.5 align-middle' + (c.align === 'center' ? ' text-center' : ''), 'data-label': c.label }, cell));
    });
    tbody.appendChild(r);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow border-t rounded-md overflow-hidden' },
    h('div', { class: 'x-table w-full h-full overflow-x-auto' },
      h('div', { class: 'relative' },
        h('table', { class: 'w-full relative eco-table' }, thead, tbody)),
      opts.after || null));
}

function skeletonRows(columns, n) {
  var widths = ['70%', '45%', '60%', '35%', '50%'];
  var rows = [];
  for (var i = 0; i < n; i++) {
    rows.push({ item: null, cells: columns.map(function (c, j) { return bar(c.skeleton || widths[(i + j) % widths.length]); }) });
  }
  return rows;
}

function field(label, control, hint, opts) {
  opts = opts || {};
  var err = h('p', { class: 'text-error-600 dark:text-error-400', hidden: true });
  var wrap = h(opts.tag || 'div', { class: 'x-label block relative align-bottom focus:outline-none mb-5 x-input eco-field', style: INPUT_VARS },
    label ? h('p', { class: 'flex items-center gap-2 font-medium text-secondary-800 dark:text-secondary-200 eco-field-label' },
      h('span', { text: label }),
      opts.optional ? h('span', { class: 'text-xs font-normal text-secondary-500', text: '(' + t('optional') + ')' }) : null) : null,
    h('div', { class: 'relative' }, control),
    h('div', { class: 'x-input-footer eco-field-foot' },
      hint ? h('p', { class: 'text-secondary-500 dark:text-secondary-400', text: hint }) : null,
      err));
  wrap.setError = function (msg) {
    err.textContent = msg || '';
    err.hidden = !msg;
    wrap.classList.toggle('eco-invalid', !!msg);
    var input = control && control.matches && control.matches('input,textarea,select') ? control : control && control.querySelector ? control.querySelector('input:not([type="file"]),textarea,select') : null;
    if (!input) return;
    if (msg) {
      if (!err.id) err.id = 'eco-err-' + Math.random().toString(36).slice(2, 9);
      input.setAttribute('aria-invalid', 'true');
      input.setAttribute('aria-errormessage', err.id);
    } else {
      input.removeAttribute('aria-invalid');
      input.removeAttribute('aria-errormessage');
    }
  };
  return wrap;
}

function sidePanel(title, body, footer, onclose) {
  var shade = h('div', { class: 'fixed z-[60] inset-0 opacity-60 dark:opacity-70 dark:bg-black bg-secondary-900 w-full h-full transition duration-200' });
  var panel = h('div', { class: 'x-form relative flex flex-col z-10 dark:bg-secondary-900 rounded-l-md shadow-lg overflow-hidden transition-all duration-300 h-full w-full bg-white sm:max-w-3xl', role: 'dialog', 'aria-modal': 'true', 'aria-label': title },
    h('div', { class: 'text-lg font-semibold' }, h('div', { class: 'pl-6 py-[0.69rem] border-b pr-12 truncate', text: title })),
    h('div', { class: 'relative overflow-y-auto w-full flex-1' }, h('div', { class: 'p-6' }, body)),
    h('div', { class: '!absolute flex items-center top-2 z-10 right-2' },
      button({ variant: 'ghost', size: 'iconLg', icon: 'close', title: t('close'), onclick: onclose })),
    footer ? h('div', { class: 'flex gap-4 bg-secondary-50 dark:bg-secondary-800 p-4 justify-end items-center' }, footer) : null);
  var holder = h('div', { class: 'fixed z-[60] inset-0 overflow-y-auto flex justify-end h-screen overflow-hidden transition' }, panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) onclose(); });
  var root = h('div', { 'data-economy': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, shade, holder);
  return root;
}

function segmented(options, value, onchange, label, size) {
  var wrap = h('div', { class: 'eco-seg inline-flex flex-wrap items-center gap-0.5 rounded-md p-0.5', role: 'radiogroup', 'aria-label': label || null });
  var btns = options.map(function (o) {
    return h('button', { type: 'button', role: 'radio', class: 'eco-seg-btn inline-flex items-center gap-1.5 rounded font-medium ' + (size === 'md' ? 'px-3 py-1.5 text-sm' : 'px-2 py-1 text-xs'), 'data-value': o.value,
      onclick: function () { if (String(value) === String(o.value)) return; pick(o.value); onchange(o.value); } },
      o.icon ? icon(o.icon, 'h-4 w-4') : null, h('span', { text: o.label }));
  });
  function pick(v) {
    value = v;
    btns.forEach(function (b) {
      var on = b.getAttribute('data-value') === String(v);
      b.setAttribute('aria-checked', on ? 'true' : 'false');
      b.classList.toggle('eco-seg-on', on);
    });
  }
  btns.forEach(function (b) { wrap.appendChild(b); });
  pick(value);
  wrap.set = pick;
  return wrap;
}

function confirmModal(opts) {
  return new Promise(function (resolve) {
    var done = false;
    function finish(v) {
      if (done) return;
      done = true;
      root.remove();
      document.removeEventListener('keydown', esc, true);
      resolve(v);
    }
    function esc(e) { if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); finish(false); } }
    var ok = button({ variant: opts.danger ? 'dangerSolid' : 'primary', size: 'footer', label: opts.ok || t('yes'), onclick: function () { finish(true); } });
    var box = h('div', { class: 'relative w-full eco-dialog bg-white dark:bg-secondary-900 rounded-md shadow-lg overflow-hidden', role: 'alertdialog', 'aria-modal': 'true', 'aria-label': opts.title },
      h('div', { class: 'p-6 flex gap-3' },
        h('div', { class: 'shrink-0 ' + (opts.danger ? 'text-error-600 dark:text-error-400' : 'eco-selected') }, icon(opts.icon || 'alert', 'h-6 w-6')),
        h('div', { class: 'min-w-0' },
          h('h2', { class: 'text-lg font-semibold mb-1', text: opts.title }),
          opts.text ? h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: opts.text }) : null,
          opts.body || null)),
      h('div', { class: 'flex gap-3 justify-end bg-secondary-50 dark:bg-secondary-800 px-6 py-3' },
        button({ variant: 'ghost', size: 'footer', label: opts.cancel || t('cancel'), onclick: function () { finish(false); } }), ok));
    var holder = h('div', { class: 'fixed z-[70] inset-0 flex items-center justify-center p-4' }, box);
    holder.addEventListener('mousedown', function (e) { if (e.target === holder) finish(false); });
    var root = h('div', { 'data-economy': 'dialog', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'fixed z-[70] inset-0 opacity-60 dark:opacity-70 dark:bg-black bg-secondary-900' }), holder);
    document.body.appendChild(root);
    document.addEventListener('keydown', esc, true);
    ok.focus();
  });
}
