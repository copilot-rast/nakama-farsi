var BTN = 'x-button eco-btn relative transition duration-150 focus-visible:outline-secondary-300 outline-transparent outline outline-1 outline-offset-2 inline-flex items-center justify-center font-medium whitespace-nowrap overflow-hidden align-middle active:!shadow-none border appearance-none shrink-0 rounded-md';
var BTN_STYLE = {
  primary: '--x-button-text: #fff; --x-button-border: rgb(102,84,226); --x-button-dark-text: #fff; --x-button-dark-border: rgb(107 89 238); --x-button-bg: rgb(107 89 238); --x-button-bg-hover: rgb(102,84,226); --x-button-bg-active: rgb(69,58,154); --x-button-dark-bg: rgb(102,84,226); --x-button-dark-bg-hover: rgb(93,77,206); --x-button-dark-bg-active: rgb(69,58,154);',
  ghost: '--x-button-border: transparent; --x-button-text: #1f2937; --x-button-dark-text: white; --x-button-bg-hover: #f3f4f6; --x-button-bg-active: #e5e7eb; --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  muted: '--x-button-border: transparent; --x-button-text: #475569; --x-button-dark-text: #64748b; --x-button-bg-hover: #f8fafc; --x-button-bg-active: #f1f5f9; --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
  link: '--x-button-border: transparent; --x-button-text: rgb(102,84,226); --x-button-dark-text: rgb(130,117,240); --x-button-bg-hover: rgb(244,244,253); --x-button-bg-active: rgb(233,232,252); --x-button-dark-bg-hover: #1f2937; --x-button-dark-bg-active: #374151;',
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
var INPUT_VARS = '--x-input-border: rgb(130,117,240);';
var SELECT_MD = INPUT_MD + ' eco-select';
var SELECT_SM = INPUT_SM + ' eco-select';

function searchInput(value, placeholder, oninput) {
  var input = h('input', {
    class: INPUT_SM + ' !pl-10',
    type: 'text',
    value: value || '',
    placeholder: placeholder,
    'aria-label': placeholder
  });
  input.addEventListener('input', oninput);
  return h('label', { class: 'x-label block relative align-bottom focus:outline-none x-input eco-search', style: INPUT_VARS },
    h('div', { class: 'relative' },
      icon('search', 'h-4 w-4 ml-2 left-1 text-secondary-600 dark:text-secondary-300 absolute my-auto inset-y-0'),
      input));
}

function filterSelect(label, value, options, onchange) {
  var sel = h('select', { class: SELECT_SM, 'aria-label': label, onchange: function () { onchange(sel.value); } });
  sel.appendChild(h('option', { value: '', text: label + ': ' + t('all') }));
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value || '')) opt.selected = true;
    sel.appendChild(opt);
  });
  return h('label', { class: 'x-label block relative x-input', style: INPUT_VARS }, sel);
}

function toggle(checked, label, onchange, size) {
  var on = !!checked;
  var track = h('div', { class: '', style: '--x-toggle-bg: rgb(107 89 238); --x-toggle-dark-bg: rgb(102,84,226);' });
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
  var box = h('div', {});
  var mark = document.createElementNS(SVG_NS, 'svg');
  mark.setAttribute('viewBox', '0 0 20 20');
  mark.innerHTML = '<path d="M0 11l2-2 5 5L18 3l2 2L7 18z"/>';
  box.appendChild(mark);
  var wrap = h('label', { class: 'x-checkbox inline-block relative cursor-pointer align-middle', tabindex: '0', role: 'checkbox' });
  function paint() {
    wrap.setAttribute('style', on
      ? '--x-checkbox-bg: rgb(107 89 238); --x-checkbox-border: rgb(107 89 238); --x-checkbox-dark-bg: rgb(107 89 238); --x-checkbox-dark-border: rgb(107 89 238);'
      : '--x-checkbox-bg: #fff; --x-checkbox-border: #d1d5db; --x-checkbox-dark-bg: #111827; --x-checkbox-dark-border: #9ca3af;');
    box.setAttribute('class', 'rounded flex justify-center items-center shrink-0 border border-[color:var(--x-checkbox-border)] bg-[color:var(--x-checkbox-bg)] dark:border-[color:var(--x-checkbox-dark-border)] dark:bg-[color:var(--x-checkbox-dark-bg)] h-4 w-4');
    mark.setAttribute('class', 'fill-current text-white h-2.5 w-2.5' + (on ? '' : ' opacity-0'));
    wrap.setAttribute('aria-checked', on ? 'true' : 'false');
  }
  wrap.appendChild(h('div', { class: 'flex items-center' }, box,
    h('div', { class: 'inline-block text-secondary-800 dark:text-secondary-200 pl-2 font-medium' }, h('span', { text: label }))));
  function flip(e) { e.preventDefault(); on = !on; paint(); onchange(on); }
  wrap.addEventListener('click', flip);
  wrap.addEventListener('keydown', function (e) { if (e.key === ' ' || e.key === 'Enter') flip(e); });
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

function pageFrame(title, desc, actions) {
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
          actionsBox)),
      content));
  return { el: frame, content: content, actions: actionsBox };
}

function notice(kind, text, action) {
  var color = kind === 'error' ? 'error' : kind === 'success' ? 'success' : 'secondary';
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

function pager(state, onpage, onrefresh) {
  var total = state.total || 0;
  var from = total ? (state.page - 1) * state.limit + 1 : 0;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  var info = h('div', { class: 'text-xs text-secondary-500 shrink-0' });
  if (state.loading && !state.loaded) info.appendChild(bar('5rem'));
  else if (!total) info.appendChild(h('div', { text: t('range.none') }));
  else info.appendChild(h('div', { class: 'flex items-center gap-1' }, h('span', { class: 'tabular-nums', text: t('range', { from: from, to: to, total: total }) })));
  return h('div', { class: 'flex items-center gap-2 justify-end shrink-0' },
    info,
    onrefresh ? button({ variant: 'muted', size: 'icon', icon: 'refresh', iconClass: 'h-3 w-3', title: t('refresh'), onclick: onrefresh }) : null,
    button({ variant: 'muted', size: 'icon', icon: 'first', iconClass: 'h-3 w-3', title: t('page.first'), disabled: state.page <= 1, onclick: function () { onpage(1); } }),
    button({ variant: 'muted', size: 'icon', icon: 'prev', iconClass: 'h-3 w-3', title: t('page.prev'), disabled: state.page <= 1, onclick: function () { onpage(state.page - 1); } }),
    button({ variant: 'muted', size: 'icon', icon: 'next', iconClass: 'h-3 w-3', title: t('page.next'), disabled: state.page >= last, onclick: function () { onpage(state.page + 1); } }));
}

function table(columns, rows, opts) {
  opts = opts || {};
  var thead = h('thead', { class: 'x-table-head align-bottom bg-secondary-100 dark:bg-secondary-700 sticky top-0 z-10' });
  var tr = h('tr', { class: 'text-sm text-secondary-700 dark:text-secondary-200 border-b' });
  columns.forEach(function (c) {
    tr.appendChild(h('th', { class: 'x-table-header relative p-2.5 font-medium', width: c.width },
      h('div', { class: 'flex items-center gap-1 select-none ' + (c.align === 'center' ? 'justify-center' : 'justify-start') }, h('div', { class: 'flex items-center gap-1', text: c.label }))));
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
    label ? h('p', { class: 'flex items-center gap-2 font-medium text-secondary-800 dark:text-secondary-200 mb-1' },
      h('span', { text: label }),
      opts.optional ? h('span', { class: 'text-xs font-normal text-secondary-500', text: '(' + t('optional') + ')' }) : null) : null,
    h('div', { class: 'relative' }, control),
    h('div', { class: 'x-input-footer text-xs mt-1' },
      hint ? h('p', { class: 'text-secondary-500 dark:text-secondary-400', text: hint }) : null,
      err));
  wrap.setError = function (msg) {
    err.textContent = msg || '';
    err.hidden = !msg;
    wrap.classList.toggle('eco-invalid', !!msg);
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
