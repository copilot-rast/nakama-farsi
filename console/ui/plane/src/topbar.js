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
  if (nat.length >= 2 && !nat[0].href && info) nat[0].href = info.link.href;
  var out = [];
  if (info && info.group && info.group.id !== 'settings') out.push({ label: info.group.label, href: null });
  if (nat.length >= 2) return out.concat(nat);
  var tab = info ? tabInfo() : null;
  if (tab && !(tab.tab.first && tab.tab.label === tab.hub.label)) {
    out.push({ label: tab.hub.label, href: tab.hub.tabs[0].href });
    out.push({ label: tab.tab.label, href: null });
  } else if (info) out.push({ label: info.link.label, href: null });
  else if (nat.length) out = out.concat(nat);
  return out;
}

function paintCrumbs(box, crumbs) {
  box.innerHTML = '';
  crumbs.forEach(function (c, i) {
    var last = i === crumbs.length - 1;
    if (i) box.insertAdjacentHTML('beforeend', '<span class="pl-crumb-sep" aria-hidden="true">' + plIcon('chevron', 'pl-ic pl-ic-dir') + '</span>');
    if (last) box.appendChild(el('span', { class: 'pl-crumb-current', 'aria-current': 'page', text: c.label, dir: /^[0-9a-f-]{12,}$/i.test(c.label) ? 'ltr' : null }));
    else if (c.href) box.appendChild(el('a', { class: 'pl-crumb-link', href: c.href, text: c.label }));
    else box.appendChild(el('span', { class: 'pl-crumb-link', text: c.label }));
  });
}

function segIndicator(seg) {
  var on = seg.querySelector('[aria-current="page"], [aria-checked="true"], [aria-selected="true"]');
  var ind = seg.querySelector(':scope > .pl-seg-ind');
  if (!ind) return;
  if (!on) { ind.style.opacity = '0'; return; }
  var sr = seg.getBoundingClientRect();
  var r = on.getBoundingClientRect();
  if (!sr.width || !r.width) return;
  var x = (isRtl() ? r.right - sr.right + 4 : r.left - sr.left - 4) + seg.scrollLeft;
  ind.style.width = '100px';
  ind.style.transform = 'translateX(' + x.toFixed(1) + 'px) scaleX(' + (r.width / 100).toFixed(4) + ')';
  ind.style.opacity = '1';
}

function paintTabs(box, info) {
  var g = info && info.group && (info.group.id === 'economy' || info.group.id === 'publication' || info.group.id === 'moderation' || info.group.id === 'support') ? info.group : null;
  if (g) {
    var sib = g.links.filter(function (l) { return (l.group || '') === (info.link.group || ''); });
    g = sib.length > 1 ? { id: g.id, label: info.link.groupLabel || g.label, links: sib } : null;
  }
  var key = g ? g.id + '|' + JSON.stringify(g.links.map(function (l) { return [l.href, l.label, l.badge]; })) + '|' + info.link.href : '';
  if (box.getAttribute('data-key') === key) { if (g) segIndicator(box); return; }
  var first = !box.getAttribute('data-key') || box.getAttribute('data-group') !== (g && g.id);
  box.setAttribute('data-key', key);
  box.setAttribute('data-group', g ? g.id : '');
  box.innerHTML = '';
  box.hidden = !g;
  if (!g) return;
  box.setAttribute('aria-label', g.label);
  box.appendChild(el('span', { class: 'pl-seg-ind', 'aria-hidden': 'true' }));
  g.links.forEach(function (l) {
    var a = el('a', { href: l.href, class: 'pl-seg-item', 'aria-current': l.href === info.link.href ? 'page' : null }, el('span', { text: l.label }));
    if (l.badge) a.appendChild(el('em', { class: 'pl-seg-count', text: plNum(l.badge) }));
    box.appendChild(a);
  });
  box.classList.toggle('pl-seg-still', first);
  requestAnimationFrame(function () {
    var on = box.querySelector('[aria-current="page"]');
    if (on && box.scrollWidth > box.clientWidth) {
      var br = box.getBoundingClientRect();
      var or = on.getBoundingClientRect();
      if (or.left < br.left || or.right > br.right) box.scrollLeft += (or.left + or.width / 2) - (br.left + br.width / 2);
    }
    segIndicator(box);
    if (first) requestAnimationFrame(function () { box.classList.remove('pl-seg-still'); });
  });
}

function l10n() {
  return window.NakamaL10n || null;
}

function buildLangSeg() {
  var seg = el('div', { class: 'pl-seg pl-seg-lang', role: 'radiogroup', 'data-l10n-skip': '' });
  seg.appendChild(el('span', { class: 'pl-seg-ind', 'aria-hidden': 'true' }));
  var L = l10n();
  var codes = L && L.locales ? Object.keys(L.locales) : ['fa-IR', 'en'];
  codes.forEach(function (code) {
    var info = L && L.locales ? L.locales[code] : { short: code === 'en' ? 'EN' : 'فا', lang: code };
    var on = L ? L.locale === code : (code === 'en') === (plLang() === 'en');
    var b = el('button', { type: 'button', role: 'radio', class: 'pl-seg-item', lang: info.lang, 'aria-checked': on ? 'true' : 'false', 'aria-label': info.name || code, 'data-locale': code, text: info.short || code });
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
  var label = pst(dark ? 'top.toLight' : 'top.toDark');
  if (tb.theme.getAttribute('data-mode') === (dark ? 'dark' : 'light') && tb.theme.getAttribute('aria-label') === label) return;
  tb.theme.setAttribute('data-mode', dark ? 'dark' : 'light');
  tb.theme.innerHTML = plIcon(dark ? 'sun' : 'moon');
  tb.theme.setAttribute('aria-label', label);
  tb.theme.setAttribute('title', label);
  tb.theme.setAttribute('aria-pressed', dark ? 'true' : 'false');
}

function flowValue(n) {
  var sr = n && n.shadowRoot;
  if (!sr) return (n && n.textContent || '').trim();
  var out = '';
  Array.prototype.forEach.call(sr.querySelectorAll('[part~="number"] [part~="digit"], [part~="number"] [part~="symbol"]'), function (d) {
    var cur = parseInt(d.style.getPropertyValue('--current'), 10);
    if (d.classList.contains('digit')) out += isNaN(cur) ? '?' : String(cur);
    else out += d.textContent;
  });
  out = out.replace(/\s+/g, '');
  return !out || out.indexOf('?') !== -1 ? '–' : out;
}

function nativeStats() {
  var box = document.querySelector('#main > div.sticky:first-child .hidden.md\\:flex.text-xs');
  if (!box) return [];
  var out = [];
  Array.prototype.forEach.call(box.children, function (c) {
    if (c.classList.contains('x-divider')) return;
    var label = c.querySelector('.text-secondary-400');
    var flow = c.querySelector('number-flow-vue');
    if (label) out.push({ label: label.textContent.trim(), value: flow ? flowValue(flow) : '' });
  });
  return out;
}

var statsTimer = 0;
function paintStats() {
  var pop = tb.stats;
  if (!pop) return;
  var rows = nativeStats();
  var key = JSON.stringify(rows) + plLang();
  if (pop.getAttribute('data-key') === key) return;
  pop.setAttribute('data-key', key);
  pop.innerHTML = '';
  pop.appendChild(el('div', { class: 'pl-stats-head' }, el('span', { class: 'pl-dot', 'aria-hidden': 'true' }), el('b', { text: pst('top.status') })));
  var grid = el('dl', { class: 'pl-stats-grid' });
  rows.forEach(function (r) {
    grid.appendChild(el('div', { class: 'pl-stat' }, el('dt', { text: r.label }), el('dd', { text: plNum(r.value) })));
  });
  pop.appendChild(grid);
}

function toggleStats(force) {
  var on = force === undefined ? !root.classList.contains('pl-stats-open') : force;
  root.classList.toggle('pl-stats-open', on);
  var b = tb.end && tb.end.querySelector('.pl-status');
  if (b) b.setAttribute('aria-expanded', on ? 'true' : 'false');
  clearInterval(statsTimer);
  if (on) {
    paintStats();
    placeStats(b);
    statsTimer = setInterval(paintStats, 1000);
  }
}

function placeStats(b) {
  var pop = tb.stats;
  var box = pop && pop.offsetParent;
  if (!b || !box) return;
  var br = box.getBoundingClientRect();
  var pr = b.getBoundingClientRect();
  var start = Math.max(0, Math.round(isRtl() ? br.right - pr.right : pr.left - br.left));
  var room = isRtl() ? br.right - start - 8 : window.innerWidth - 8 - (br.left + start);
  var w = pop.offsetWidth;
  if (w > room) start = Math.max(0, start - Math.ceil(w - room));
  pop.style.insetInlineStart = start + 'px';
}

function buildEnd() {
  var box = el('div', { class: 'pl-tb-end', 'data-l10n-skip': '' });
  var search = el('button', { type: 'button', class: 'pl-search', 'aria-label': pst('top.searchPages'), 'aria-haspopup': 'dialog', onclick: function () { openPalette(); } });
  search.innerHTML = plIcon('search');
  search.appendChild(el('span', { class: 'pl-search-ph', text: pst('top.search') }));
  search.appendChild(el('span', { class: 'pl-keys', dir: 'ltr' }, el('kbd', { text: 'Ctrl' }), el('kbd', { text: 'K' })));
  var status = el('button', { type: 'button', class: 'pl-status', 'aria-expanded': 'false', 'aria-label': pst('top.status'), title: pst('top.status'), onclick: function (e) { e.stopPropagation(); toggleStats(); } });
  status.appendChild(el('span', { class: 'pl-dot', 'aria-hidden': 'true' }));
  status.appendChild(el('span', { class: 'pl-status-label', text: pst('top.live') }));
  status.setAttribute('aria-controls', 'pl-stats');
  tb.stats = el('div', { class: 'pl-stats', id: 'pl-stats', role: 'region', 'aria-label': pst('top.status'), 'data-l10n-skip': '' });
  tb.up = el('a', { class: 'pl-up', href: '#/system/updates', hidden: true, 'data-pl-updates': '' });
  tb.seg = buildLangSeg();
  tb.theme = el('button', { type: 'button', class: 'pl-round pl-theme', onclick: themeClick });
  paintTheme();
  box.appendChild(tb.up);
  box.appendChild(search);
  box.appendChild(status);
  box.appendChild(tb.stats);
  box.appendChild(tb.seg);
  box.appendChild(tb.theme);
  return box;
}

function paintUpdates() {
  var a = tb.up;
  if (!a) return;
  var api = nc();
  var u = api && api.updates ? api.updates() : null;
  var n = u ? u.newCommits || 0 : 0;
  var key = plLang() + '|' + n + '|' + (u ? u.href : '');
  if (a.getAttribute('data-key') === key) return;
  a.setAttribute('data-key', key);
  a.hidden = !n;
  if (!n) return;
  if (u.href) a.setAttribute('href', u.href);
  var label = pst('top.updatesLabel', { n: plNum(n) });
  a.setAttribute('aria-label', label);
  a.setAttribute('title', label);
  a.innerHTML = plIconD('cloudDown');
  a.appendChild(el('span', { class: 'pl-up-label', text: pst('top.updates') }));
  a.appendChild(el('em', { class: 'pl-up-n', text: plNum(n) }));
}

function paintTopbar() {
  var bar = topbar();
  if (!bar) return;
  var holder = bar.querySelector(':scope > div.flex.w-full > div.flex-1');
  var end = bar.querySelector('.justify-end');
  if (!holder || !end) return;
  if (!tb.start) {
    tb.start = el('div', { class: 'pl-tb-start', 'data-l10n-skip': '' },
      el('button', { type: 'button', class: 'pl-round pl-menu', 'aria-label': pst('top.menu'), html: plIcon('menu'), onclick: function () { openSheet(); } }),
      el('nav', { class: 'pl-crumbs' }),
      el('nav', { class: 'pl-seg pl-tabs', hidden: true }));
  }
  if (!tb.end) tb.end = buildEnd();
  if (tb.start.parentNode !== holder || holder.firstChild !== tb.start) holder.insertBefore(tb.start, holder.firstChild);
  if (tb.end.parentNode !== end || end.firstChild !== tb.end) end.insertBefore(tb.end, end.firstChild);
  var info = pageInfo();
  var crumbs = crumbsFor(bar);
  var key = plLang() + '|' + JSON.stringify(crumbs);
  var cb = tb.start.querySelector('.pl-crumbs');
  cb.setAttribute('aria-label', pst('crumb.label'));
  if (tb.key !== key) {
    tb.key = key;
    paintCrumbs(cb, crumbs);
  }
  paintTabs(tb.start.querySelector('.pl-tabs'), info);
  paintTheme();
  paintUpdates();
  var lseg = tb.seg;
  if (lseg && !lseg.__spInit && lseg.offsetWidth) { lseg.__spInit = true; segIndicator(lseg); }
}

function paletteItems(q) {
  var items = allPages();
  q = (q || '').trim().toLowerCase();
  if (!q) return items;
  var rank = function (text) {
    var i = String(text || '').toLowerCase().indexOf(q);
    if (i === -1) return -1;
    return i === 0 ? 0 : /[\s\u200c]/.test(String(text).charAt(i - 1)) ? 1 : 2;
  };
  var byLabel = [];
  var byGroup = [];
  items.forEach(function (it, i) {
    var r = rank(it.label);
    if (r !== -1) byLabel.push({ it: it, r: r, i: i });
    else if (rank(it.group) !== -1) byGroup.push(it);
  });
  byLabel.sort(function (a, b) { return a.r - b.r || a.i - b.i; });
  return byLabel.length ? byLabel.map(function (x) { return x.it; }) : byGroup;
}

function openPalette() {
  if (tb.palette) { tb.palette.querySelector('input').focus(); return; }
  var prev = document.activeElement;
  var input = el('input', { type: 'text', class: 'pl-pal-input', 'aria-label': pst('top.searchPages'), placeholder: pst('top.searchPages'), autocomplete: 'off', role: 'combobox', 'aria-expanded': 'true', 'aria-controls': 'pl-pal-list' });
  var list = el('div', { class: 'pl-pal-list', id: 'pl-pal-list', role: 'listbox', 'aria-label': pst('palette.group.pages') });
  var win = el('div', { class: 'pl-pal', role: 'dialog', 'aria-modal': 'true', 'aria-label': pst('top.searchPages') },
    el('div', { class: 'pl-pal-head', html: plIcon('search') }, input, el('kbd', { class: 'pl-esc', text: 'Esc' })),
    list,
    el('div', { class: 'pl-pal-foot', text: pst('palette.hint') }));
  var shell = el('div', { class: 'pl-pal-holder', 'data-l10n-skip': '', dir: root.getAttribute('dir') || 'ltr' }, el('div', { class: 'pl-scrim' }), win);
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
    if (!rows.length) { list.appendChild(el('div', { class: 'pl-pal-empty', text: pst('top.noResults') })); input.removeAttribute('aria-activedescendant'); return; }
    rows.slice(0, 40).forEach(function (it, i) {
      var r = el('a', { href: it.href, id: 'pl-pal-' + i, role: 'option', class: 'pl-pal-row', 'aria-selected': i === sel ? 'true' : 'false' });
      r.innerHTML = plIcon(it.icon || (it.psSection ? 'settings' : 'page'));
      r.appendChild(el('span', { class: 'pl-pal-label', text: it.label }));
      if (it.group) r.appendChild(el('span', { class: 'pl-pal-group', text: it.group }));
      r.addEventListener('click', function (e) { if (e.metaKey || e.ctrlKey) return; e.preventDefault(); go(it); });
      r.addEventListener('mousemove', function () { if (sel !== i) { sel = i; mark(); } });
      list.appendChild(r);
    });
    mark();
  }
  function mark() {
    Array.prototype.forEach.call(list.querySelectorAll('[role="option"]'), function (r, i) { r.setAttribute('aria-selected', i === sel ? 'true' : 'false'); });
    input.setAttribute('aria-activedescendant', 'pl-pal-' + sel);
    var cur = list.querySelector('[aria-selected="true"]');
    if (cur && cur.scrollIntoView) cur.scrollIntoView({ block: 'nearest' });
  }
  function close() {
    if (!tb.palette) return;
    var p = tb.palette;
    tb.palette = null;
    document.removeEventListener('keydown', keys, true);
    p.classList.add('pl-closing');
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
  shell.addEventListener('mousedown', function (e) { if (e.target === shell || e.target.classList.contains('pl-scrim')) close(); });
  document.addEventListener('keydown', keys, true);
  tb.palette = shell;
  tb.palette.close = close;
  document.body.appendChild(shell);
  paint();
  input.focus();
}

function globalKeys(e) {
  if ((e.ctrlKey || e.metaKey) && !e.altKey && (e.key === 'k' || e.key === 'K')) {
    if (root.classList.contains('pl-auth')) return;
    e.preventDefault();
    openPalette();
  } else if (e.key === 'Escape' && root.classList.contains('pl-stats-open')) {
    toggleStats(false);
  }
}

document.addEventListener('click', function (e) {
  if (!root.classList.contains('pl-stats-open')) return;
  if (e.target.closest && e.target.closest('.pl-status, .pl-stats')) return;
  toggleStats(false);
}, true);
