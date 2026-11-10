var tb = { start: null, end: null, key: '', palette: null, theme: null, seg: null };

function topbar() {
  var bar = document.querySelector('#main > div.sticky:first-child');
  return bar && bar.querySelector('.justify-end') ? bar : null;
}

function nativeCrumbs(bar) {
  var items = bar.querySelectorAll(':scope > div.flex.w-full > div.flex-1 > nav li');
  var out = [];
  Array.prototype.forEach.call(items, function (li) {
    var a = li.querySelector('a[href]');
    var text = (a || li).textContent.replace(/\s*\/\s*$/, '').trim();
    if (!text || text === '/') return;
    out.push({ label: text, href: a ? a.getAttribute('href') : null });
  });
  return out;
}

function crumbsFor(bar) {
  var info = pageInfo();
  var nat = nativeCrumbs(bar);
  var out = [];
  if (info && info.group && info.group.id !== 'settings') out.push({ label: info.group.label, href: null });
  if (nat.length >= 2) return out.concat(nat);
  if (info) out.push({ label: info.link.label, href: null });
  else if (nat.length) out = out.concat(nat);
  return out;
}

function paintCrumbs(box, crumbs) {
  box.innerHTML = '';
  crumbs.forEach(function (c, i) {
    var last = i === crumbs.length - 1;
    if (i) box.insertAdjacentHTML('beforeend', '<span class="sp-crumb-sep" aria-hidden="true">' + spIcon('chevron', 'sp-ic sp-ic-dir') + '</span>');
    if (last) box.appendChild(el('span', { class: 'sp-crumb-current', 'aria-current': 'page', text: c.label, dir: /^[0-9a-f-]{12,}$/i.test(c.label) ? 'ltr' : null }));
    else if (c.href) box.appendChild(el('a', { class: 'sp-crumb-link', href: c.href, text: c.label }));
    else box.appendChild(el('span', { class: 'sp-crumb-link', text: c.label }));
  });
}

function segIndicator(seg) {
  var on = seg.querySelector('[aria-current="page"], [aria-checked="true"], [aria-selected="true"]');
  var ind = seg.querySelector(':scope > .sp-seg-ind');
  if (!ind) return;
  if (!on) { ind.style.opacity = '0'; return; }
  var sr = seg.getBoundingClientRect();
  var r = on.getBoundingClientRect();
  if (!sr.width || !r.width) return;
  var x = isRtl() ? r.right - sr.right + 4 : r.left - sr.left - 4;
  ind.style.width = '100px';
  ind.style.transform = 'translateX(' + x.toFixed(1) + 'px) scaleX(' + (r.width / 100).toFixed(4) + ')';
  ind.style.opacity = '1';
}

function paintTabs(box, info) {
  var g = info && info.group && (info.group.id === 'economy' || info.group.id === 'publication') ? info.group : null;
  var key = g ? g.id + '|' + JSON.stringify(g.links.map(function (l) { return [l.href, l.label, l.badge]; })) + '|' + info.link.href : '';
  if (box.getAttribute('data-key') === key) { if (g) segIndicator(box); return; }
  var first = !box.getAttribute('data-key') || box.getAttribute('data-group') !== (g && g.id);
  box.setAttribute('data-key', key);
  box.setAttribute('data-group', g ? g.id : '');
  box.innerHTML = '';
  box.hidden = !g;
  if (!g) return;
  box.setAttribute('aria-label', g.label);
  box.appendChild(el('span', { class: 'sp-seg-ind', 'aria-hidden': 'true' }));
  g.links.forEach(function (l) {
    var a = el('a', { href: l.href, class: 'sp-seg-item', 'aria-current': l === info.link ? 'page' : null }, el('span', { text: l.label }));
    if (l.badge) a.appendChild(el('em', { class: 'sp-seg-count', text: spNum(l.badge) }));
    box.appendChild(a);
  });
  box.classList.toggle('sp-seg-still', first);
  requestAnimationFrame(function () {
    segIndicator(box);
    if (first) requestAnimationFrame(function () { box.classList.remove('sp-seg-still'); });
  });
}

function l10n() {
  return window.NakamaL10n || null;
}

function buildLangSeg() {
  var seg = el('div', { class: 'sp-seg sp-seg-lang', role: 'radiogroup', 'data-l10n-skip': '' });
  seg.appendChild(el('span', { class: 'sp-seg-ind', 'aria-hidden': 'true' }));
  var L = l10n();
  var codes = L && L.locales ? Object.keys(L.locales) : ['fa-IR', 'en'];
  codes.forEach(function (code) {
    var info = L && L.locales ? L.locales[code] : { short: code === 'en' ? 'EN' : 'فا', lang: code };
    var on = L ? L.locale === code : (code === 'en') === (spLang() === 'en');
    var b = el('button', { type: 'button', role: 'radio', class: 'sp-seg-item', lang: info.lang, 'aria-checked': on ? 'true' : 'false', 'aria-label': info.name || code, 'data-locale': code, text: info.short || code });
    b.addEventListener('click', function () {
      if (b.getAttribute('aria-checked') === 'true') return;
      Array.prototype.forEach.call(seg.querySelectorAll('[role="radio"]'), function (x) { x.setAttribute('aria-checked', x === b ? 'true' : 'false'); });
      segIndicator(seg);
      setTimeout(function () {
        if (L && L.setLocale) L.setLocale(code);
        else { var sw = document.querySelector('.l10n-switcher:not(.economy-theme)'); if (sw) sw.click(); }
      }, reduced() ? 0 : 200);
    });
    seg.appendChild(b);
  });
  return seg;
}

function themeClick() {
  var api = nc();
  if (api && api.toggleTheme) api.toggleTheme();
  else {
    var next = !root.classList.contains('dark');
    root.classList.toggle('dark', next);
    try { localStorage.setItem('color-mode', next ? 'dark' : 'light'); } catch (e) { next = !!next; }
  }
  paintTheme();
}

function paintTheme() {
  if (!tb.theme) return;
  var dark = root.classList.contains('dark');
  var label = st(dark ? 'top.toLight' : 'top.toDark');
  if (tb.theme.getAttribute('data-mode') === (dark ? 'dark' : 'light') && tb.theme.getAttribute('aria-label') === label) return;
  tb.theme.setAttribute('data-mode', dark ? 'dark' : 'light');
  tb.theme.innerHTML = spIcon(dark ? 'sun' : 'moon');
  tb.theme.setAttribute('aria-label', label);
  tb.theme.setAttribute('title', label);
  tb.theme.setAttribute('aria-pressed', dark ? 'true' : 'false');
}

function toggleStats(force) {
  var on = force === undefined ? !root.classList.contains('sp-stats-open') : force;
  root.classList.toggle('sp-stats-open', on);
  var b = tb.end && tb.end.querySelector('.sp-status');
  if (b) b.setAttribute('aria-expanded', on ? 'true' : 'false');
}

function buildEnd() {
  var box = el('div', { class: 'sp-tb-end', 'data-l10n-skip': '' });
  var search = el('button', { type: 'button', class: 'sp-search', 'aria-label': st('top.searchPages'), 'aria-haspopup': 'dialog', onclick: function () { openPalette(); } });
  search.innerHTML = spIcon('search');
  search.appendChild(el('span', { class: 'sp-search-ph', text: st('top.search') }));
  search.appendChild(el('span', { class: 'sp-keys', dir: 'ltr' }, el('kbd', { text: 'Ctrl' }), el('kbd', { text: 'K' })));
  var status = el('button', { type: 'button', class: 'sp-status', 'aria-expanded': 'false', 'aria-label': st('top.status'), title: st('top.status'), onclick: function (e) { e.stopPropagation(); toggleStats(); } });
  status.appendChild(el('span', { class: 'sp-dot', 'aria-hidden': 'true' }));
  status.appendChild(el('span', { class: 'sp-status-label', text: st('top.live') }));
  tb.seg = buildLangSeg();
  tb.theme = el('button', { type: 'button', class: 'sp-round sp-theme', onclick: themeClick });
  paintTheme();
  box.appendChild(search);
  box.appendChild(status);
  box.appendChild(tb.seg);
  box.appendChild(tb.theme);
  return box;
}

function paintTopbar() {
  var bar = topbar();
  if (!bar) return;
  var holder = bar.querySelector(':scope > div.flex.w-full > div.flex-1');
  var end = bar.querySelector('.justify-end');
  if (!holder || !end) return;
  if (!tb.start) {
    tb.start = el('div', { class: 'sp-tb-start', 'data-l10n-skip': '' },
      el('button', { type: 'button', class: 'sp-round sp-menu', 'aria-label': st('top.menu'), html: spIcon('menu'), onclick: function () { openSheet(); } }),
      el('nav', { class: 'sp-crumbs' }),
      el('nav', { class: 'sp-seg sp-tabs', hidden: true }));
  }
  if (!tb.end) tb.end = buildEnd();
  if (tb.start.parentNode !== holder || holder.firstChild !== tb.start) holder.insertBefore(tb.start, holder.firstChild);
  if (tb.end.parentNode !== end || end.firstChild !== tb.end) end.insertBefore(tb.end, end.firstChild);
  var info = pageInfo();
  var crumbs = crumbsFor(bar);
  var key = spLang() + '|' + JSON.stringify(crumbs);
  var cb = tb.start.querySelector('.sp-crumbs');
  cb.setAttribute('aria-label', st('crumb.label'));
  if (tb.key !== key) {
    tb.key = key;
    paintCrumbs(cb, crumbs);
  }
  paintTabs(tb.start.querySelector('.sp-tabs'), info);
  paintTheme();
  var lseg = tb.seg;
  if (lseg && !lseg.__spInit && lseg.offsetWidth) { lseg.__spInit = true; segIndicator(lseg); }
}

function paletteItems(q) {
  var items = allPages();
  q = (q || '').trim().toLowerCase();
  if (!q) return items;
  return items.filter(function (it) { return (it.label + ' ' + (it.group || '')).toLowerCase().indexOf(q) !== -1; });
}

function openPalette() {
  if (tb.palette) { tb.palette.querySelector('input').focus(); return; }
  var prev = document.activeElement;
  var input = el('input', { type: 'search', class: 'sp-pal-input', 'aria-label': st('top.searchPages'), placeholder: st('top.searchPages'), autocomplete: 'off', role: 'combobox', 'aria-expanded': 'true', 'aria-controls': 'sp-pal-list' });
  var list = el('div', { class: 'sp-pal-list', id: 'sp-pal-list', role: 'listbox', 'aria-label': st('palette.group.pages') });
  var win = el('div', { class: 'sp-pal', role: 'dialog', 'aria-modal': 'true', 'aria-label': st('top.searchPages') },
    el('div', { class: 'sp-pal-head', html: spIcon('search') }, input, el('kbd', { class: 'sp-esc', text: 'Esc' })),
    list,
    el('div', { class: 'sp-pal-foot', text: st('palette.hint') }));
  var shell = el('div', { class: 'sp-pal-holder', 'data-l10n-skip': '', dir: root.getAttribute('dir') || 'ltr' }, el('div', { class: 'sp-scrim' }), win);
  var sel = 0;
  var rows = [];
  function go(it) {
    close();
    if (it.ps) { var api = nc() && nc().projectSettings; if (api) { api.open(); return; } }
    location.hash = it.href;
  }
  function paint() {
    rows = paletteItems(input.value);
    if (sel >= rows.length) sel = rows.length - 1;
    if (sel < 0) sel = 0;
    list.innerHTML = '';
    if (!rows.length) { list.appendChild(el('div', { class: 'sp-pal-empty', text: st('top.noResults') })); input.removeAttribute('aria-activedescendant'); return; }
    rows.slice(0, 40).forEach(function (it, i) {
      var r = el('a', { href: it.href, id: 'sp-pal-' + i, role: 'option', class: 'sp-pal-row', 'aria-selected': i === sel ? 'true' : 'false' });
      r.innerHTML = it.svg || spIcon(it.psSection ? 'sliders' : 'page');
      r.appendChild(el('span', { class: 'sp-pal-label', text: it.label }));
      if (it.group) r.appendChild(el('span', { class: 'sp-pal-group', text: it.group }));
      r.addEventListener('click', function (e) { if (e.metaKey || e.ctrlKey) return; e.preventDefault(); go(it); });
      r.addEventListener('mousemove', function () { if (sel !== i) { sel = i; mark(); } });
      list.appendChild(r);
    });
    mark();
  }
  function mark() {
    Array.prototype.forEach.call(list.querySelectorAll('[role="option"]'), function (r, i) { r.setAttribute('aria-selected', i === sel ? 'true' : 'false'); });
    input.setAttribute('aria-activedescendant', 'sp-pal-' + sel);
    var cur = list.querySelector('[aria-selected="true"]');
    if (cur && cur.scrollIntoView) cur.scrollIntoView({ block: 'nearest' });
  }
  function close() {
    if (!tb.palette) return;
    var p = tb.palette;
    tb.palette = null;
    document.removeEventListener('keydown', keys, true);
    p.classList.add('sp-closing');
    setTimeout(function () { p.remove(); }, reduced() ? 0 : 160);
    if (prev && prev.isConnected) { try { prev.focus({ preventScroll: true }); } catch (e) { prev = null; } }
  }
  function keys(e) {
    if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); close(); }
    else if (e.key === 'ArrowDown') { e.preventDefault(); sel = Math.min(rows.length - 1, sel + 1); mark(); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); sel = Math.max(0, sel - 1); mark(); }
    else if (e.key === 'Enter' && rows[sel]) { e.preventDefault(); go(rows[sel]); }
    else if (e.key === 'Tab') { e.preventDefault(); input.focus(); }
  }
  input.addEventListener('input', function () { sel = 0; paint(); });
  shell.addEventListener('mousedown', function (e) { if (e.target === shell || e.target.classList.contains('sp-scrim')) close(); });
  document.addEventListener('keydown', keys, true);
  tb.palette = shell;
  tb.palette.close = close;
  document.body.appendChild(shell);
  paint();
  input.focus();
}

function globalKeys(e) {
  if ((e.ctrlKey || e.metaKey) && !e.altKey && (e.key === 'k' || e.key === 'K')) {
    if (root.classList.contains('sp-auth')) return;
    e.preventDefault();
    openPalette();
  } else if (e.key === 'Escape' && root.classList.contains('sp-stats-open')) {
    toggleStats(false);
  }
}

document.addEventListener('click', function (e) {
  if (!root.classList.contains('sp-stats-open')) return;
  if (e.target.closest && e.target.closest('.sp-status')) return;
  toggleStats(false);
}, true);
