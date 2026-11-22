var ND_SVG = {
  calendar: 'M208,32H184V24a8,8,0,0,0-16,0v8H88V24a8,8,0,0,0-16,0v8H48A16,16,0,0,0,32,48V208a16,16,0,0,0,16,16H208a16,16,0,0,0,16-16V48A16,16,0,0,0,208,32ZM72,48v8a8,8,0,0,0,16,0V48h80v8a8,8,0,0,0,16,0V48h24V80H48V48ZM208,208H48V96H208V208Z',
  clock: 'M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm0,192a88,88,0,1,1,88-88A88.1,88.1,0,0,1,128,216Zm64-88a8,8,0,0,1-8,8H128a8,8,0,0,1-8-8V72a8,8,0,0,1,16,0v48h48A8,8,0,0,1,192,128Z',
  x: 'M205.66,194.34a8,8,0,0,1-11.32,11.32L128,139.31,61.66,205.66a8,8,0,0,1-11.32-11.32L116.69,128,50.34,61.66A8,8,0,0,1,61.66,50.34L128,116.69l66.34-66.35a8,8,0,0,1,11.32,11.32L139.31,128Z',
  next: 'M181.66,133.66l-80,80a8,8,0,0,1-11.32-11.32L164.69,128,90.34,53.66a8,8,0,0,1,11.32-11.32l80,80A8,8,0,0,1,181.66,133.66Z',
  prev: 'M165.66,202.34a8,8,0,0,1-11.32,11.32l-80-80a8,8,0,0,1,0-11.32l80-80a8,8,0,0,1,11.32,11.32L91.31,128Z',
  up: 'M213.66,165.66a8,8,0,0,1-11.32,0L128,91.31,53.66,165.66a8,8,0,0,1-11.32-11.32l80-80a8,8,0,0,1,11.32,0l80,80A8,8,0,0,1,213.66,165.66Z',
  down: 'M213.66,101.66l-80,80a8,8,0,0,1-11.32,0l-80-80A8,8,0,0,1,53.66,90.34L128,164.69l74.34-74.35a8,8,0,0,1,11.32,11.32Z'
};
var ndOpen = null;
var ndSeq = 0;

function ndIcon(name, cls) {
  return '<svg class="nd-ic' + (cls ? ' ' + cls : '') + '" viewBox="0 0 256 256" fill="currentColor" aria-hidden="true" focusable="false"><path d="' + ND_SVG[name] + '"/></svg>';
}

function ndEl(tag, attrs) {
  var el = document.createElement(tag);
  Object.keys(attrs || {}).forEach(function (k) {
    var v = attrs[k];
    if (v === null || v === undefined || v === false) return;
    if (k === 'text') el.textContent = v;
    else if (k === 'html') el.innerHTML = v;
    else if (k.slice(0, 2) === 'on' && typeof v === 'function') el.addEventListener(k.slice(2), v);
    else el.setAttribute(k, v === true ? '' : v);
  });
  for (var i = 2; i < arguments.length; i++) {
    var c = arguments[i];
    if (c === null || c === undefined || c === false) continue;
    el.appendChild(typeof c === 'string' ? document.createTextNode(c) : c);
  }
  return el;
}

function ndMobile() {
  return typeof window.matchMedia === 'function' && window.matchMedia('(max-width: 640px)').matches;
}

function ndReduced() {
  return typeof window.matchMedia === 'function' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

function ndOrigin(mount) {
  var probe = document.createElement('span');
  probe.style.cssText = 'position:fixed;left:0;top:0;width:0;height:0;visibility:hidden;pointer-events:none;';
  mount.appendChild(probe);
  var r = probe.getBoundingClientRect();
  probe.remove();
  return { x: r.left, y: r.top };
}

function ndSetFixed(el, x, y, w, origin) {
  var o = origin || { x: 0, y: 0 };
  el.style.left = Math.round(x - o.x) + 'px';
  el.style.top = Math.round(y - o.y) + 'px';
  if (w) el.style.width = Math.round(w) + 'px';
}

function closePicker(restoreFocus) {
  if (ndOpen) ndOpen.close(restoreFocus);
}

function openPicker(host) {
  if (ndOpen) ndOpen.close(false);
  var lang = ndLang();
  var cal = host.calendar || activeCalendar(lang);
  var C = ND_CALS[cal];
  var rtl = C.rtl;
  var ws = ndWeekStart(cal, lang);
  var zone = host.zone || ndZone();
  var mode = host.mode;
  var initial = host.getValue();
  var now = nowWall(zone);
  var today = now.jdn;
  var minJ = host.minJdn == null ? null : host.minJdn;
  var maxJ = host.maxJdn == null ? null : host.maxJdn;
  var st = { view: 'days', sel: null, a: null, b: null, hover: null, h: now.h, mi: 0, cursor: today, vy: 0, vm: 0, yearBase: 0 };
  if (mode === 'date') st.sel = dayToJdn(initial);
  else if (mode === 'datetime') {
    var w = isoToWall(initial, zone);
    if (w) { st.sel = dayToJdn(w.slice(0, 10)); st.h = +w.slice(11, 13); st.mi = +w.slice(14, 16); }
  } else if (mode === 'time') {
    var tm = initial ? ndParseTime(String(initial)) : null;
    if (tm) { st.h = tm.h; st.mi = tm.mi; } else { st.h = now.h; st.mi = now.mi - now.mi % 5; }
  } else if (mode === 'range' && initial) {
    st.a = dayToJdn(initial.from);
    st.b = dayToJdn(initial.to);
  }
  st.cursor = clampJ(st.sel != null ? st.sel : st.a != null ? st.a : today);
  syncView();

  var spins = {};
  var id = 'nd-pop-' + (++ndSeq);
  var titleId = id + '-title';
  var titles = { date: 'pickDate', datetime: 'pickDateTime', time: 'pickTime', range: 'pickRange' };
  var pop = ndEl('div', { class: 'nd-pop nd-mode-' + mode, id: id, role: 'dialog', 'aria-modal': 'true', 'aria-label': (host.label ? host.label + ' · ' : '') + ndt(titles[mode]), dir: rtl ? 'rtl' : 'ltr', lang: ndLocale(lang), 'data-nd-pop': mode, 'data-calendar': cal, 'data-l10n-skip': '' });
  var scrim = ndEl('div', { class: 'nd-scrim', 'aria-hidden': 'true', 'data-nd-scrim': '' });
  var head = ndEl('div', { class: 'nd-head' });
  var body = ndEl('div', { class: 'nd-body' });
  var timeBox = mode === 'datetime' || mode === 'time' ? ndEl('div', { class: 'nd-time', role: 'group', 'aria-label': ndt('time'), dir: 'ltr' }) : null;
  var hint = mode === 'range' ? ndEl('p', { class: 'nd-hint', 'aria-live': 'polite', 'data-nd-hint': '' }) : null;
  var foot = ndEl('div', { class: 'nd-foot' });
  var grab = ndEl('div', { class: 'nd-grab', 'aria-hidden': 'true' });
  pop.appendChild(grab);
  if (mode !== 'time') pop.appendChild(head);
  if (mode === 'range') pop.appendChild(presetBar());
  if (mode !== 'time') pop.appendChild(body);
  if (hint) pop.appendChild(hint);
  if (timeBox) { pop.appendChild(timeBox); buildTime(); }
  pop.appendChild(foot);
  buildFoot();

  function clampJ(j) {
    if (minJ != null && j < minJ) return minJ;
    if (maxJ != null && j > maxJ) return maxJ;
    return j;
  }

  function off(j) { return (minJ != null && j < minJ) || (maxJ != null && j > maxJ); }

  function syncView() {
    var p = C.fromDay(st.cursor);
    st.vy = p.y;
    st.vm = p.m;
  }

  function navButton(dirName, label, fn) {
    var icon = rtl ? (dirName === 'prev' ? 'next' : 'prev') : dirName;
    return ndEl('button', { type: 'button', class: 'nd-nav nd-' + dirName, 'aria-label': label, title: label, html: ndIcon(icon), 'data-nd-nav': dirName, onclick: fn });
  }

  function renderHead() {
    head.innerHTML = '';
    var labels = st.view === 'days' ? ['prevMonth', 'nextMonth'] : st.view === 'months' ? ['prevYear', 'nextYear'] : ['prevYears', 'nextYears'];
    var step = function (n) {
      return function () {
        if (st.view === 'days') moveMonth(n, false);
        else if (st.view === 'months') { st.vy += n; render(); }
        else { st.yearBase += 12 * n; render(); }
        var b = head.querySelector('[data-nd-nav="' + (n < 0 ? 'prev' : 'next') + '"]');
        if (b) b.focus();
      };
    };
    var title = ndEl('div', { class: 'nd-title', id: titleId, 'aria-live': 'polite' });
    if (st.view === 'years') {
      title.appendChild(ndEl('span', { class: 'nd-title-text', text: ndYearText(st.yearBase, lang) + ' – ' + ndYearText(st.yearBase + 11, lang) }));
    } else {
      if (st.view === 'days') title.appendChild(ndEl('button', { type: 'button', class: 'nd-title-btn', 'data-nd-view': 'months', 'aria-label': ndt('chooseMonth') + ': ' + ndMonthName(st.vy, st.vm, cal, lang), text: ndMonthName(st.vy, st.vm, cal, lang), onclick: function () { st.view = 'months'; render(); focusCursor(); } }));
      title.appendChild(ndEl('button', { type: 'button', class: 'nd-title-btn', 'data-nd-view': 'years', 'aria-label': ndt('chooseYear') + ': ' + ndYearText(st.vy, lang), text: ndYearText(st.vy, lang), onclick: function () { st.view = 'years'; st.yearBase = st.vy - 5; render(); focusCursor(); } }));
    }
    head.appendChild(navButton('prev', ndt(labels[0]), step(-1)));
    head.appendChild(title);
    head.appendChild(navButton('next', ndt(labels[1]), step(1)));
  }

  function inRange(j) {
    var a = st.a;
    var b = st.b != null ? st.b : st.hover;
    if (a == null || b == null) return false;
    var lo = Math.min(a, b);
    var hi = Math.max(a, b);
    return j >= lo && j <= hi;
  }

  function dayClasses(btn, j) {
    var p = C.fromDay(j);
    var cls = ['nd-day'];
    if (p.m !== st.vm) cls.push('is-out');
    if (j === today) cls.push('is-today');
    if (off(j)) cls.push('is-off');
    var selected = false;
    if (mode === 'range') {
      var a = st.a;
      var b = st.b != null ? st.b : st.hover;
      if (a != null && j === Math.min(a, b == null ? a : b)) cls.push('is-start');
      if (a != null && j === Math.max(a, b == null ? a : b)) cls.push('is-end');
      if (inRange(j)) cls.push('is-in');
      selected = j === st.a || j === st.b || (st.a != null && st.b != null && inRange(j));
    } else selected = j === st.sel;
    if (selected) cls.push('is-sel');
    btn.className = cls.join(' ');
    btn.parentNode.setAttribute('aria-selected', selected ? 'true' : 'false');
  }

  function renderDays() {
    body.innerHTML = '';
    var table = ndEl('table', { class: 'nd-grid', role: 'grid', 'aria-labelledby': titleId, 'data-nd-grid': '' });
    var hr = ndEl('tr', {});
    for (var i = 0; i < 7; i++) {
      var wd = (ws + i) % 7;
      hr.appendChild(ndEl('th', { scope: 'col', abbr: ndWeekdayName(wd, lang, 'long'), text: ndWeekdayName(wd, lang, lang === 'fa' ? 'narrow' : 'short') }));
    }
    table.appendChild(ndEl('thead', {}, hr));
    var tb = ndEl('tbody', {});
    var first = C.toDay(st.vy, st.vm, 1);
    var start = first - ((jdnWeekday(first) - ws + 7) % 7);
    for (var r = 0; r < 6; r++) {
      var tr = ndEl('tr', {});
      for (var c = 0; c < 7; c++) {
        var j = start + r * 7 + c;
        var td = ndEl('td', { role: 'gridcell' });
        var btn = ndEl('button', { type: 'button', tabindex: j === st.cursor ? '0' : '-1', 'data-day': jdnToDay(j), 'data-jdn': String(j), 'aria-label': ndDayLabel(j, cal, lang) + (off(j) ? ' · ' + ndt('disabled') : ''), text: ndDigits(String(C.fromDay(j).d), lang) });
        if (j === today) btn.setAttribute('aria-current', 'date');
        if (off(j)) btn.setAttribute('aria-disabled', 'true');
        td.appendChild(btn);
        tr.appendChild(td);
        dayClasses(btn, j);
      }
      tb.appendChild(tr);
    }
    table.appendChild(tb);
    table.addEventListener('click', function (e) {
      var b = e.target.closest && e.target.closest('button[data-jdn]');
      if (b) pickDay(+b.getAttribute('data-jdn'));
    });
    if (mode === 'range') {
      table.addEventListener('mouseover', function (e) {
        var b = e.target.closest && e.target.closest('button[data-jdn]');
        if (!b || st.a == null || st.b != null) return;
        st.hover = +b.getAttribute('data-jdn');
        paintRange();
      });
    }
    body.appendChild(table);
  }

  function paintRange() {
    Array.prototype.forEach.call(body.querySelectorAll('button[data-jdn]'), function (b) { dayClasses(b, +b.getAttribute('data-jdn')); });
  }

  function renderMonths() {
    body.innerHTML = '';
    var grid = ndEl('div', { class: 'nd-cells nd-months', role: 'group', 'aria-labelledby': titleId });
    var cur = C.fromDay(st.cursor);
    for (var m = 1; m <= 12; m++) {
      (function (m) {
        var on = cur.y === st.vy && cur.m === m;
        var sel = st.sel != null && C.fromDay(st.sel).y === st.vy && C.fromDay(st.sel).m === m;
        grid.appendChild(ndEl('button', { type: 'button', class: 'nd-cell' + (sel ? ' is-sel' : ''), 'data-nd-month': String(m), tabindex: on || (cur.y !== st.vy && m === 1) ? '0' : '-1', 'aria-pressed': sel ? 'true' : 'false', text: ndMonthName(st.vy, m, cal, lang), onclick: function () {
          var d = Math.min(C.fromDay(st.cursor).d, C.monthLength(st.vy, m));
          st.cursor = C.toDay(st.vy, m, d);
          st.view = 'days';
          syncView();
          render();
          focusCursor();
        } }));
      })(m);
    }
    body.appendChild(grid);
  }

  function renderYears() {
    body.innerHTML = '';
    var grid = ndEl('div', { class: 'nd-cells nd-years', role: 'group', 'aria-labelledby': titleId });
    for (var i = 0; i < 12; i++) {
      (function (y) {
        var on = y === st.vy;
        grid.appendChild(ndEl('button', { type: 'button', class: 'nd-cell' + (on ? ' is-sel' : ''), 'data-nd-year': String(y), tabindex: on || (i === 0 && (st.vy < st.yearBase || st.vy > st.yearBase + 11)) ? '0' : '-1', 'aria-pressed': on ? 'true' : 'false', text: ndYearText(y, lang), onclick: function () {
          var p = C.fromDay(st.cursor);
          var d = Math.min(p.d, C.monthLength(y, p.m));
          st.cursor = C.toDay(y, p.m, d);
          syncView();
          st.view = 'months';
          render();
          focusCursor();
        } }));
      })(st.yearBase + i);
    }
    body.appendChild(grid);
  }

  function render() {
    if (mode === 'time') return;
    renderHead();
    if (st.view === 'days') renderDays();
    else if (st.view === 'months') renderMonths();
    else renderYears();
    if (hint) hint.textContent = ndt(st.a != null && st.b == null ? 'rangeEnd' : 'rangeStart');
  }

  function focusCursor() {
    var t = st.view === 'days' ? body.querySelector('button[data-jdn="' + st.cursor + '"]') : body.querySelector('.nd-cell[tabindex="0"]');
    if (t) t.focus({ preventScroll: true });
  }

  function moveMonth(n, keepFocus) {
    var p = C.fromDay(st.cursor);
    var y = p.y;
    var m = p.m + n;
    while (m < 1) { m += 12; y -= 1; }
    while (m > 12) { m -= 12; y += 1; }
    st.cursor = C.toDay(y, m, Math.min(p.d, C.monthLength(y, m)));
    syncView();
    render();
    if (keepFocus) focusCursor();
  }

  function setCursor(j) {
    var before = C.fromDay(st.cursor);
    st.cursor = j;
    var p = C.fromDay(j);
    if (p.y !== before.y || p.m !== before.m || p.m !== st.vm || p.y !== st.vy) { syncView(); render(); }
    else Array.prototype.forEach.call(body.querySelectorAll('button[data-jdn]'), function (b) { b.setAttribute('tabindex', +b.getAttribute('data-jdn') === j ? '0' : '-1'); });
    if (mode === 'range' && st.a != null && st.b == null) { st.hover = j; paintRange(); }
    focusCursor();
  }

  function emitDateTime() {
    if (st.sel == null) return;
    var iso = jdnTimeToIso(st.sel, st.h, st.mi, zone);
    if (host.min && new Date(iso) < new Date(host.min)) iso = new Date(host.min).toISOString();
    if (host.max && new Date(iso) > new Date(host.max)) iso = new Date(host.max).toISOString();
    host.commit(iso, false);
  }

  function pickDay(j) {
    if (off(j)) return;
    st.cursor = j;
    if (mode === 'date') { st.sel = j; host.commit(jdnToDay(j), true); close(true); return; }
    if (mode === 'datetime') {
      st.sel = j;
      emitDateTime();
      var p = C.fromDay(j);
      if (p.m !== st.vm || p.y !== st.vy) syncView();
      render();
      focusCursor();
      paintTime();
      return;
    }
    if (st.a == null || st.b != null) { st.a = j; st.b = null; st.hover = j; render(); focusCursor(); return; }
    var lo = Math.min(st.a, j);
    var hi = Math.max(st.a, j);
    st.a = lo;
    st.b = hi;
    host.commit({ from: jdnToDay(lo), to: jdnToDay(hi) }, true);
    close(true);
  }

  function presetBar() {
    var bar = ndEl('div', { class: 'nd-presets', role: 'group', 'aria-label': ndt('presets') });
    [['today', 1], ['last7', 7], ['last30', 30], ['last90', 90]].forEach(function (p) {
      bar.appendChild(ndEl('button', { type: 'button', class: 'nd-chip', 'data-nd-preset': String(p[1]), text: ndt(p[0]), onclick: function () {
        var to = clampJ(today);
        var from = clampJ(today - (p[1] - 1));
        st.a = from;
        st.b = to;
        host.commit({ from: jdnToDay(from), to: jdnToDay(to) }, true);
        close(true);
      } }));
    });
    return bar;
  }

  function spin(kind, max) {
    var label = ndt(kind === 'h' ? 'hour' : 'minute');
    var val = ndEl('div', { class: 'nd-spin-val', role: 'spinbutton', tabindex: '0', 'aria-label': label, 'aria-valuemin': '0', 'aria-valuemax': String(max), 'data-nd-spin': kind });
    var buf = '';
    var bufAt = 0;
    var set = function (v) {
      v = ((v % (max + 1)) + (max + 1)) % (max + 1);
      if (kind === 'h') st.h = v; else st.mi = v;
      if (mode === 'datetime' && st.sel == null) st.sel = clampJ(st.cursor);
      changed();
    };
    var cur = function () { return kind === 'h' ? st.h : st.mi; };
    val.addEventListener('keydown', function (e) {
      var k = e.key;
      if (k === 'ArrowUp') set(cur() + 1);
      else if (k === 'ArrowDown') set(cur() - 1);
      else if (k === 'PageUp') set(cur() + (kind === 'h' ? 6 : 10));
      else if (k === 'PageDown') set(cur() - (kind === 'h' ? 6 : 10));
      else if (k === 'Home') set(0);
      else if (k === 'End') set(max);
      else if (/^[0-9۰-۹٠-٩]$/.test(k)) {
        var d = ndLatin(k);
        var t = Date.now();
        buf = t - bufAt < 1200 && buf.length < 2 ? buf + d : d;
        bufAt = t;
        var n = +buf;
        if (n > max) { buf = d; n = +d; }
        set(n);
      } else return;
      e.preventDefault();
      e.stopPropagation();
    });
    val.addEventListener('wheel', function (e) { e.preventDefault(); set(cur() + (e.deltaY < 0 ? 1 : -1)); }, { passive: false });
    var up = ndEl('button', { type: 'button', class: 'nd-spin-btn', tabindex: '-1', 'aria-label': label + ' · ' + ndt('increase'), html: ndIcon('up'), onclick: function () { set(cur() + 1); val.focus(); } });
    var down = ndEl('button', { type: 'button', class: 'nd-spin-btn', tabindex: '-1', 'aria-label': label + ' · ' + ndt('decrease'), html: ndIcon('down'), onclick: function () { set(cur() - 1); val.focus(); } });
    spins[kind] = val;
    return ndEl('div', { class: 'nd-spin' }, up, val, down);
  }

  function buildTime() {
    timeBox.appendChild(ndEl('span', { class: 'nd-time-label', html: ndIcon('clock'), title: ndt('time') }));
    timeBox.appendChild(spin('h', 23));
    timeBox.appendChild(ndEl('span', { class: 'nd-time-sep', 'aria-hidden': 'true', text: ':' }));
    timeBox.appendChild(spin('mi', 59));
    paintTime();
  }

  function paintTime() {
    if (!timeBox) return;
    [['h', st.h], ['mi', st.mi]].forEach(function (p) {
      var v = spins[p[0]];
      v.textContent = ndDigits(ndPad(p[1]), lang);
      v.setAttribute('aria-valuenow', String(p[1]));
      v.setAttribute('aria-valuetext', ndDigits(ndPad(st.h) + ':' + ndPad(st.mi), lang));
    });
  }

  function changed() {
    paintTime();
    if (mode === 'time') host.commit(ndPad(st.h) + ':' + ndPad(st.mi), false);
    else if (mode === 'datetime') { emitDateTime(); if (st.view === 'days') paintSel(); }
  }

  function paintSel() {
    Array.prototype.forEach.call(body.querySelectorAll('button[data-jdn]'), function (b) { dayClasses(b, +b.getAttribute('data-jdn')); });
  }

  function footButton(label, cls, fn, data) {
    var attrs = { type: 'button', class: 'nd-btn ' + cls, text: label, onclick: fn };
    attrs['data-nd-' + data] = '';
    return ndEl('button', attrs);
  }

  function buildFoot() {
    var start = ndEl('div', { class: 'nd-foot-start' });
    var end = ndEl('div', { class: 'nd-foot-end' });
    if (mode === 'date') start.appendChild(footButton(ndt('today'), 'nd-btn-ghost', function () { if (off(today)) return; host.commit(jdnToDay(today), true); close(true); }, 'today'));
    if (mode === 'datetime' || mode === 'time') start.appendChild(footButton(ndt('now'), 'nd-btn-ghost', function () {
      var n = nowWall(zone);
      st.h = n.h;
      st.mi = n.mi;
      if (mode === 'datetime') { st.sel = clampJ(n.jdn); st.cursor = st.sel; syncView(); render(); }
      changed();
    }, 'now'));
    if (!host.required) start.appendChild(footButton(ndt('clear'), 'nd-btn-ghost', function () { host.commit(null, true); close(true); }, 'erase'));
    if (mode === 'datetime' || mode === 'time') end.appendChild(footButton(ndt('done'), 'nd-btn-primary', function () { if (mode === 'datetime' && st.sel == null) { st.sel = clampJ(st.cursor); emitDateTime(); } close(true); }, 'done'));
    else end.appendChild(footButton(ndt('cancel'), 'nd-btn-ghost', function () { close(true); }, 'cancel'));
    foot.appendChild(start);
    foot.appendChild(end);
  }

  function gridKeys(e) {
    var t = e.target;
    var k = e.key;
    if (t.hasAttribute && t.hasAttribute('data-jdn')) {
      var j = +t.getAttribute('data-jdn');
      var p = C.fromDay(j);
      var fwd = rtl ? -1 : 1;
      if (k === 'ArrowRight') setCursor(j + fwd);
      else if (k === 'ArrowLeft') setCursor(j - fwd);
      else if (k === 'ArrowDown') setCursor(j + 7);
      else if (k === 'ArrowUp') setCursor(j - 7);
      else if (k === 'Home') setCursor(j - ((jdnWeekday(j) - ws + 7) % 7));
      else if (k === 'End') setCursor(j + (6 - ((jdnWeekday(j) - ws + 7) % 7)));
      else if (k === 'PageUp' || k === 'PageDown') {
        var n = (k === 'PageUp' ? -1 : 1) * (e.shiftKey ? 12 : 1);
        var y = p.y;
        var m = p.m + n;
        while (m < 1) { m += 12; y -= 1; }
        while (m > 12) { m -= 12; y += 1; }
        setCursor(C.toDay(y, m, Math.min(p.d, C.monthLength(y, m))));
      } else if (k === 'Enter' || k === ' ') pickDay(j);
      else return false;
      return true;
    }
    if (t.classList && t.classList.contains('nd-cell')) {
      var cells = Array.prototype.slice.call(body.querySelectorAll('.nd-cell'));
      var i = cells.indexOf(t);
      var fw = rtl ? -1 : 1;
      var to = k === 'ArrowRight' ? i + fw : k === 'ArrowLeft' ? i - fw : k === 'ArrowDown' ? i + 3 : k === 'ArrowUp' ? i - 3 : k === 'Home' ? 0 : k === 'End' ? cells.length - 1 : null;
      if (to === null) return false;
      to = Math.max(0, Math.min(cells.length - 1, to));
      cells.forEach(function (c, x) { c.setAttribute('tabindex', x === to ? '0' : '-1'); });
      cells[to].focus();
      return true;
    }
    return false;
  }

  function focusables() {
    return Array.prototype.filter.call(pop.querySelectorAll('button, [tabindex="0"]'), function (n) {
      return !n.disabled && n.getAttribute('tabindex') !== '-1' && n.offsetParent !== null;
    });
  }

  function escKey(e) {
    if (e.key !== 'Escape') return;
    e.preventDefault();
    e.stopPropagation();
    if (st.view !== 'days') { st.view = 'days'; render(); focusCursor(); return; }
    host.commit(initial, true);
    close(true);
  }

  pop.addEventListener('keydown', function (e) {
    if (e.key === 'Tab') {
      var list = focusables();
      if (!list.length) return;
      var i = list.indexOf(document.activeElement);
      var next = e.shiftKey ? (i <= 0 ? list.length - 1 : i - 1) : (i === -1 || i === list.length - 1 ? 0 : i + 1);
      e.preventDefault();
      list[next].focus();
      return;
    }
    if (gridKeys(e)) { e.preventDefault(); e.stopPropagation(); }
  });
  pop.addEventListener('mousedown', function (e) {
    if (!(e.target.closest && e.target.closest('button, [tabindex]'))) e.preventDefault();
  });
  pop.addEventListener('click', function (e) {
    e.stopPropagation();
    if (!(e.target.closest && e.target.closest('button'))) e.preventDefault();
  });
  scrim.addEventListener('mousedown', function (e) { e.preventDefault(); close(false); });

  function outside(e) {
    if (pop.contains(e.target) || (host.box && host.box.contains(e.target))) return;
    close(false);
  }

  function place() {
    if (!pop.isConnected) return;
    var mobile = ndMobile();
    pop.classList.toggle('nd-sheet', mobile);
    scrim.hidden = !mobile;
    var vw = document.documentElement.clientWidth || window.innerWidth;
    var vh = window.innerHeight;
    var origin = ndOrigin(host.mount);
    if (mobile) {
      pop.style.width = '';
      ndSetFixed(scrim, 0, 0, vw, origin);
      scrim.style.height = vh + 'px';
      ndSetFixed(pop, 0, vh - pop.offsetHeight, vw, origin);
      return;
    }
    pop.style.width = '';
    var anchor = host.box || host.mount;
    var r = anchor.getBoundingClientRect();
    var w = pop.offsetWidth;
    var h = pop.offsetHeight;
    var dirRtl = getComputedStyle(anchor).direction === 'rtl';
    var x = dirRtl ? r.right - w : r.left;
    x = Math.max(8, Math.min(vw - w - 8, x));
    var y = r.bottom + 6;
    if (y + h > vh - 8 && r.top - h - 6 >= 8) y = r.top - h - 6;
    y = Math.max(8, Math.min(Math.max(8, vh - h - 8), y));
    ndSetFixed(pop, x, y, 0, origin);
  }

  var closed = false;
  function close(restoreFocus) {
    if (closed) return;
    closed = true;
    document.removeEventListener('mousedown', outside, true);
    window.removeEventListener('keydown', escKey, true);
    window.removeEventListener('resize', place);
    window.removeEventListener('scroll', onScroll, true);
    if (ndOpen && ndOpen.pop === pop) ndOpen = null;
    pop.classList.add('nd-closing');
    var done = function () { pop.remove(); scrim.remove(); };
    if (ndReduced()) done(); else setTimeout(done, 120);
    if (host.onClose) host.onClose(restoreFocus);
  }

  function onScroll(e) {
    if (pop.contains(e.target)) return;
    place();
  }

  var mount = host.mount;
  mount.appendChild(scrim);
  mount.appendChild(pop);
  render();
  place();
  requestAnimationFrame(place);
  document.addEventListener('mousedown', outside, true);
  window.addEventListener('keydown', escKey, true);
  window.addEventListener('resize', place);
  window.addEventListener('scroll', onScroll, true);
  if (mode === 'time') spins.h.focus({ preventScroll: true });
  else focusCursor();
  ndOpen = { pop: pop, close: close, host: host };
  return ndOpen;
}
