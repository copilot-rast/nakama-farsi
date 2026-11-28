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
  if (p[0] === 'economy' || p[0] === 'publication' || p[0] === 'moderation' || p[0] === 'support' || p[0] === 'settings') return p[0] + '-' + (p[1] || 'index');
  if (p[0] === 'players' || p[0] === 'groups' || p[0] === 'leaderboards' || p[0] === 'storage-indexes') return p.length > 1 ? p[0] + '-detail' : p[0];
  return p[0];
}

function markPage() {
  var name = routeName();
  if (root.getAttribute('data-pl-route') !== name) root.setAttribute('data-pl-route', name);
  var main = contentRoot();
  if (!main) return;
  var box = main.querySelector('div.max-w-\\[1600px\\]');
  if (!box) return;
  var cards = box.querySelectorAll(CARD_SEL + ', .bg-white.shadow.rounded-md');
  var tables = box.querySelectorAll('table');
  var tabs = box.querySelector('.x-tab-group');
  var list = !!box.querySelector('.eco-shop') || !tabs && tables.length >= 1 && cards.length === tables.length && cards.length <= 2 && !box.querySelector('[_echarts_instance_]');
  if (list !== box.hasAttribute('data-pl-list')) box.toggleAttribute('data-pl-list', list);
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
      if (cells.length === 1 && heads.length > 1) { cells[0].setAttribute('data-pl-full', ''); continue; }
      for (var c = 0; c < cells.length; c++) {
        var lab = heads[c] || '';
        if (lab) cells[c].setAttribute('data-pl-label', lab);
        else cells[c].removeAttribute('data-pl-label');
      }
    }
    t.setAttribute('data-pl-cards', '');
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
    var ind = bar.querySelector(':scope > .pl-tabind');
    var fresh = false;
    if (!ind) {
      ind = el('span', { class: 'pl-tabind pl-still', 'aria-hidden': 'true' });
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
    if (fresh) requestAnimationFrame(function (n) { return function () { requestAnimationFrame(function () { n.classList.remove('pl-still'); }); }; }(ind));
  }
}

function readNumber(text) {
  var s = String(text || '').replace(/[۰-۹]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'.indexOf(d); }).replace(/[٬,\s]/g, '').replace('٫', '.');
  return /^-?\d+(\.\d+)?$/.test(s) ? parseFloat(s) : NaN;
}

function countKpis() {
  var nodes = document.querySelectorAll('#economy-root .eco-stat-grid .text-3xl');
  if (!nodes.length) return;
  var nf = new Intl.NumberFormat(plLang() === 'fa' ? 'fa-IR' : 'en-US', { maximumFractionDigits: 0 });
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

var KPI_ICONS = [
  [/CCU|نشست|session/i, 'users', 'info'],
  [/حضور|presence/i, 'radio', 'teal'],
  [/پارتی|part(y|ies)/i, 'group', 'warn'],
  [/مسابقه|match/i, 'gamepad', 'violet'],
  [/گوروتین|goroutine/i, 'cpu', 'mute']
];

function iconKpis() {
  var flows = document.querySelectorAll('#main > main div.text-3xl > number-flow-vue');
  for (var i = 0; i < flows.length; i++) {
    var body = flows[i].closest('.px-4.py-3');
    if (!body || body.querySelector(':scope > .pl-kpi-ic')) continue;
    var label = body.querySelector('.text-secondary-600');
    var text = label ? label.textContent : '';
    var hit = KPI_ICONS.filter(function (k) { return k[0].test(text); })[0];
    if (!hit) continue;
    var tile = el('span', { class: 'pl-kpi-ic pl-tint-' + hit[2], 'aria-hidden': 'true', html: plIconD(hit[1]) });
    body.insertBefore(tile, body.firstChild);
    body.classList.add('pl-kpi');
  }
}

var ATTN_TINT = { reports: 'warn', comments: 'info', tickets: 'violet', 'live-chat': 'teal', updates: 'teal' };

function attentionCard() {
  if (root.getAttribute('data-pl-route') !== 'dashboard') return;
  var main = document.querySelector('#main > main');
  var flow = main && main.querySelector('div.text-3xl > number-flow-vue');
  var strip = flow && flow.closest('.rounded-md.shadow.bg-white');
  var grid = strip && strip.parentElement;
  var host = grid && grid.parentElement;
  if (!host) return;
  var m = rail.model || railModel();
  if (!m) return;
  var rows = [];
  m.groups.forEach(function (g) { g.links.forEach(function (l) { if (l.badge) rows.push({ l: l, g: g }); }); });
  var key = plLang() + JSON.stringify(rows.map(function (r) { return [r.l.href, r.l.badge, r.l.label]; }));
  var box = grid.nextElementSibling && grid.nextElementSibling.classList.contains('pl-attn') ? grid.nextElementSibling : null;
  if (box && box.getAttribute('data-key') === key) return;
  if (!box) {
    box = el('section', { class: 'pl-attn', 'data-l10n-skip': '' });
    grid.after(box);
  }
  box.setAttribute('data-key', key);
  box.innerHTML = '';
  var total = rows.reduce(function (n, r) { return n + r.l.badge; }, 0);
  box.appendChild(el('div', { class: 'pl-attn-head' }, el('span', { class: 'pl-kpi-ic pl-tint-warn', 'aria-hidden': 'true', html: plIconD('alert') }), el('h2', { text: pst('attn.title') }), el('span', { class: 'pl-attn-count', text: total ? pst('attn.count', { n: total }) : '' })));
  if (!rows.length) {
    box.appendChild(el('p', { class: 'pl-attn-empty', html: plIconD('checkcircle', 'pl-ic') }, el('span', { text: pst('attn.none') })));
    return;
  }
  var list = el('div', { class: 'pl-attn-list' });
  rows.forEach(function (r) {
    var seg = routePath(r.l.href).split('/').filter(Boolean)[1] || '';
    var a = el('a', { class: 'pl-attn-row', href: r.l.href });
    a.appendChild(el('span', { class: 'pl-kpi-ic pl-tint-' + (ATTN_TINT[seg] || 'info'), 'aria-hidden': 'true', html: plIconD(r.l.icon || 'page') }));
    a.appendChild(el('span', { class: 'pl-attn-txt' }, el('b', { text: r.l.label }), el('span', { text: r.g.label })));
    a.appendChild(el('span', { class: 'pl-attn-n', text: plNum(r.l.badge) }));
    a.insertAdjacentHTML('beforeend', plIcon('chevron', 'pl-ic pl-ic-dir pl-attn-chev'));
    list.appendChild(a);
  });
  box.appendChild(list);
}

function ecoKpiTint(href) {
  if (/status=failed|grant|claimsFailed/.test(href || '')) return 'bad';
  var p = routePath(href || '#/').split('/').filter(Boolean)[1] || '';
  if (/payments|methods/.test(p)) return 'ok';
  if (/event|streak|level|task|meta-rules|meta-relations/.test(p)) return 'violet';
  if (/referral/.test(p)) return 'teal';
  return 'info';
}

function iconEcoKpis() {
  var cards = document.querySelectorAll('#economy-root .eco-stat-grid > *');
  for (var i = 0; i < cards.length; i++) {
    var c = cards[i];
    var body = c.querySelector('.p-4') || c.firstElementChild;
    if (!body || body.querySelector(':scope > .pl-kpi-ic') || !body.querySelector('.text-3xl')) continue;
    var href = c.getAttribute('href') || '';
    var alert = !!body.querySelector('.text-error-600');
    var name = href ? iconFor(href) : alert ? 'alert' : 'chart';
    if (name === 'page') name = 'chart';
    var tint = alert ? 'bad' : href ? ecoKpiTint(href) : 'mute';
    body.insertBefore(el('span', { class: 'pl-kpi-ic pl-tint-' + tint, 'aria-hidden': 'true', html: plIconD(name) }), body.firstChild);
    body.classList.add('pl-eco-kpi');
  }
}
