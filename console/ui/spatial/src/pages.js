var CARD_SEL = '.rounded-md.shadow.bg-white';

function contentRoot() {
  var eco = document.getElementById('economy-root');
  if (eco && eco.offsetParent !== null) return eco;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') return mains[i];
  return null;
}

function routeName() {
  var p = routePath().split('/').filter(Boolean);
  if (!p.length) return 'dashboard';
  if (p[0] === 'economy' || p[0] === 'publication' || p[0] === 'settings') return p[0] + '-' + (p[1] || 'index');
  if (p[0] === 'players' || p[0] === 'groups' || p[0] === 'leaderboards' || p[0] === 'storage-indexes') return p.length > 1 ? p[0] + '-detail' : p[0];
  return p[0];
}

function markPage() {
  var name = routeName();
  if (root.getAttribute('data-sp-route') !== name) root.setAttribute('data-sp-route', name);
  var main = contentRoot();
  if (!main) return;
  var box = main.querySelector('div.max-w-\\[1600px\\]');
  if (!box) return;
  var cards = box.querySelectorAll(CARD_SEL + ', .bg-white.shadow.rounded-md');
  var tables = box.querySelectorAll('table');
  var tabs = box.querySelector('.x-tab-group');
  var list = !tabs && tables.length >= 1 && cards.length === tables.length && cards.length <= 2 && !box.querySelector('[_echarts_instance_]');
  if (list !== box.hasAttribute('data-sp-list')) box.toggleAttribute('data-sp-list', list);
}

var labelSeq = 0;
function labelTables() {
  if (!mqMobile.matches) return;
  var tables = document.querySelectorAll('#main table, [data-economy] table, .ps-window table');
  for (var i = 0; i < tables.length; i++) {
    var t = tables[i];
    if (t.classList.contains('eco-table')) continue;
    var heads = Array.prototype.map.call(t.querySelectorAll('thead th'), function (th) { return th.textContent.trim(); });
    if (!heads.length) continue;
    var rows = t.tBodies.length ? t.tBodies[0].rows : [];
    var key = heads.join('|') + '#' + rows.length + '#' + (rows[0] ? rows[0].cells.length : 0);
    if (t.__spLabels === key) continue;
    t.__spLabels = key;
    labelSeq++;
    for (var r = 0; r < rows.length; r++) {
      var cells = rows[r].cells;
      if (cells.length === 1 && heads.length > 1) { cells[0].setAttribute('data-sp-full', ''); continue; }
      for (var c = 0; c < cells.length; c++) {
        var lab = heads[c] || '';
        if (lab) cells[c].setAttribute('data-sp-label', lab);
        else cells[c].removeAttribute('data-sp-label');
      }
    }
    t.setAttribute('data-sp-cards', '');
  }
}

function isAuthPage() {
  if (/^#\/(login|reset|reset-mfa|mfa)(\/|\?|$)/.test(currentHash())) return true;
  return !document.querySelector('#main') && !!document.querySelector('#app form');
}

function activeTab(bar) {
  return bar.querySelector('.x-tab.router-link-exact-active') || bar.querySelector('.x-tab[aria-current="page"], .x-tab[aria-selected="true"], .x-tab.eco-tab-on') || bar.querySelector('.x-tab.router-link-active');
}

function paintNativeTabs() {
  var bars = document.querySelectorAll('#main .x-tab-group .relative.flex.min-w-full.w-fit.border-b, [data-economy] .x-tab-group .relative.flex.min-w-full.w-fit.border-b');
  for (var i = 0; i < bars.length; i++) {
    var bar = bars[i];
    var ind = bar.querySelector(':scope > .sp-tabind');
    var fresh = false;
    if (!ind) {
      ind = el('span', { class: 'sp-tabind sp-still', 'aria-hidden': 'true' });
      bar.appendChild(ind);
      fresh = true;
    }
    var on = activeTab(bar);
    if (!on) { ind.style.opacity = '0'; continue; }
    var br = bar.getBoundingClientRect();
    var r = on.getBoundingClientRect();
    if (!r.width) continue;
    var x = r.left - br.left + bar.scrollLeft;
    var t = 'translateX(' + x.toFixed(1) + 'px) scaleX(' + (r.width / 100).toFixed(4) + ')';
    if (ind.style.transform !== t) ind.style.transform = t;
    ind.style.opacity = '1';
    if (fresh) requestAnimationFrame(function (n) { return function () { requestAnimationFrame(function () { n.classList.remove('sp-still'); }); }; }(ind));
  }
}

function readNumber(text) {
  var s = String(text || '').replace(/[۰-۹]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'.indexOf(d); }).replace(/[٬,\s]/g, '').replace('٫', '.');
  return /^-?\d+(\.\d+)?$/.test(s) ? parseFloat(s) : NaN;
}

function countKpis() {
  var nodes = document.querySelectorAll('#economy-root .eco-stat-grid .text-3xl');
  if (!nodes.length) return;
  var nf = new Intl.NumberFormat(spLang() === 'fa' ? 'fa-IR' : 'en-US', { maximumFractionDigits: 0 });
  for (var i = 0; i < nodes.length; i++) {
    var n = nodes[i];
    if (n.__spCounted) continue;
    n.__spCounted = true;
    var v = readNumber(n.textContent);
    if (isNaN(v) || v === 0) continue;
    var final = n.textContent;
    countUp(n, v, function (x) { return nf.format(Math.round(x)); });
    setTimeout(function (node, text) { return function () { if (node.isConnected) node.textContent = text; }; }(n, final), reduced() ? 0 : 560);
  }
}
