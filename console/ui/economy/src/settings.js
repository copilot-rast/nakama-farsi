var PS_ROUTE = 'project-settings';
var PS_SECTIONS = [
  { id: 'brand', label: 'tab.brand', icon: 'image', acl: 'SETTINGS', render: function (box) { brandPage(box); } },
  { id: 'setup', label: 'tab.setup', icon: 'flask', acl: 'SETTINGS', render: function (box) { setupPage(box); } },
  { id: 'modules', label: 'tab.modules', icon: 'modules', acl: 'SETTINGS', render: function (box) { modulesPage(box); } },
  { id: 'languages', label: 'tab.languages', icon: 'languages', acl: 'SETTINGS', render: function (box, route, ctx) { return languagesPage(box, route.query, ctx); } },
  { id: 'datetime', label: 'tab.datetime', icon: 'history', acl: 'SETTINGS', render: function (box) { dateSettingsPage(box); } },
  { id: 'countries', label: 'tab.countries', icon: 'flag', acl: 'SETTINGS', render: function (box, route, ctx) { return countriesPage(box, route.query, ctx); } },
  { id: 'accounts', label: 'tab.accounts', icon: 'users', acl: 'SETTINGS', render: function (box) { accountRulesPage(box); } },
  { id: 'friends', label: 'tab.friends', icon: 'users', acl: 'SETTINGS', render: function (box, route, ctx) { return friendRulesPage(box, route.query, ctx); } },
  { id: 'otp', label: 'tab.otp', icon: 'phone', acl: 'NOTIFICATION', render: function (box, route, ctx) { return otpPage(box, route.query, ctx); } },
  { id: 'app-versions', label: 'tab.app-versions', icon: 'versions', acl: 'SETTINGS', render: function (box, route, ctx) { return appVersionsPage(box, route.query, ctx); } },
  { id: 'categories', label: 'nav.categories', icon: 'categories', acl: 'HIRO_ECONOMY', render: function (box, route, ctx) { return listPage(box, 'categories', route.query, ctx); } },
  { id: 'tags', label: 'nav.tags', icon: 'tags', acl: 'HIRO_ECONOMY', render: function (box, route, ctx) { return listPage(box, 'tags', route.query, ctx); } },
  { id: 'report-titles', label: 'tab.report-titles', icon: 'report-titles', acl: 'CHANNEL_MESSAGE', render: function (box, route, ctx) { return listPage(box, 'report-titles', route.query, ctx); } },
  { id: 'report-rules', label: 'tab.report-rules', icon: 'report-titles', acl: 'CHANNEL_MESSAGE', render: function (box) { reportRulesPage(box); } },
  { id: 'events', label: 'tab.events', icon: 'event-types', acl: 'HIRO_ECONOMY', render: function (box) { eventSettingsPage(box); } },
  { id: 'referral-rules', label: 'tab.referral-rules', icon: 'referral-rules', acl: 'HIRO_ECONOMY', render: function (box) { referralRulesPage(box); } },
  { id: 'support-departments', label: 'tab.support-departments', icon: 'support-departments', acl: 'ACCOUNT_NOTES', render: function (box, route, ctx) { loadAgents(); return listPage(box, 'support-departments', route.query, ctx); } },
  { id: 'support-canned', label: 'tab.support-canned', icon: 'support-canned', acl: 'ACCOUNT_NOTES', render: function (box, route, ctx) { return listPage(box, 'support-canned', route.query, ctx); } },
  { id: 'support-settings', label: 'tab.support-settings', icon: 'support-settings', acl: 'ACCOUNT_NOTES', render: function (box) { supportSettingsPage(box); } },
  { id: 'payment-methods', label: 'tab.payment-methods', icon: 'payments', acl: 'HIRO_ECONOMY', render: function (box) { paymentMethodsPage(box); } },
  { id: 'sms', label: 'tab.sms', icon: 'comments', acl: 'NOTIFICATION', render: function (box) { providersPage(box, 'sms'); } },
  { id: 'email', label: 'tab.email', icon: 'mail', acl: 'NOTIFICATION', render: function (box) { providersPage(box, 'email'); } },
  { id: 'push', label: 'tab.push', icon: 'push', acl: 'NOTIFICATION', render: function (box) { providersPage(box, 'push'); } },
  { id: 'voice', label: 'tab.voice', icon: 'voice', acl: 'SETTINGS', render: function (box) { voicePage(box); } },
  { id: 'updates-source', label: 'tab.updates-source', icon: 'updates-source', acl: 'SETTINGS', render: function (box) { updatesSourcePage(box); } }
];
var PS_LEGACY = {
  'settings/brand': 'brand',
  'settings/setup': 'setup',
  'settings/languages': 'languages',
  'settings/payment-methods': 'payment-methods',
  'settings/sms': 'sms',
  'settings/email': 'email',
  'settings/push': 'push',
  'base/categories': 'categories',
  'base/tags': 'tags',
  'economy/categories': 'categories',
  'economy/tags': 'tags',
  'economy/payment-methods': 'payment-methods',
  'publication/providers': 'sms'
};
var ACL_RESOURCES = { ACCOUNT: 0, ACCOUNT_NOTES: 5, CHANNEL_MESSAGE: 13, NOTIFICATION: 20, SETTINGS: 22, HIRO_ECONOMY: 27 };
var PS_RETURN_KEY = 'economy.ps.return';
var aclCache = { tk: null, bits: null };
var ps = null;
var psReturn = null;
try { psReturn = sessionStorage.getItem(PS_RETURN_KEY); } catch (e) { psReturn = null; }

function psSection(id) {
  for (var i = 0; i < PS_SECTIONS.length; i++) if (PS_SECTIONS[i].id === id) return PS_SECTIONS[i];
  return null;
}

function b64urlBytes(s) {
  s = String(s || '').replace(/-/g, '+').replace(/_/g, '/');
  while (s.length % 4) s += '=';
  var bin = atob(s);
  var out = [];
  for (var i = 0; i < bin.length; i++) out.push(bin.charCodeAt(i));
  return out;
}

function consoleAcl() {
  var tk = token();
  if (!tk) return null;
  if (aclCache.tk === tk) return aclCache.bits;
  var bits = null;
  try {
    var payload = JSON.parse(new TextDecoder().decode(new Uint8Array(b64urlBytes(tk.split('.')[1]))));
    if (payload && typeof payload.acl === 'string' && payload.acl) bits = b64urlBytes(payload.acl);
  } catch (e) { bits = null; }
  aclCache = { tk: tk, bits: bits };
  return bits;
}

function aclCan(resource, level) {
  var bits = consoleAcl();
  if (!bits || !bits.length) return true;
  if (bits.every(function (b) { return b === 0xFF; })) return true;
  var idx = ACL_RESOURCES[resource] * 3 + (level || 0);
  return !!(bits[idx >> 3] & (1 << (7 - (idx & 7))));
}

function psAllowed() {
  return PS_SECTIONS.filter(function (s) { var mod = moduleOfPs(s.id); return aclCan(s.acl, 0) && (!mod || moduleOn(mod)); });
}

function psEntryHash() {
  var allowed = psAllowed();
  return hashFor(allowed.length ? allowed[0].id : PS_SECTIONS[0].id);
}

function psRemember(hash) {
  if (!hash || /^#\/login/.test(hash)) return;
  psReturn = hash;
  try { sessionStorage.setItem(PS_RETURN_KEY, hash); } catch (e) { return; }
}

function psNavigate(hash, replace) {
  if (router) {
    var p = replace ? router.replace(hash.slice(1)) : router.push(hash.slice(1));
    Promise.resolve(p).then(schedule, schedule);
    return;
  }
  if (replace) history.replaceState(history.state, '', location.pathname + location.search + hash);
  else location.hash = hash;
  schedule();
}

function psEnter(e) {
  if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
  e.preventDefault();
  if (ps) { if (ps.list) { var tab = ps.list.querySelector('[aria-selected="true"]'); if (tab) tab.focus(); } return; }
  psNavigate(e.currentTarget.getAttribute('href'), false);
}

function psExit() {
  var back = psReturn || '#/';
  var st = history.state;
  if (st && typeof st.back === 'string' && '#' + st.back === back) history.back();
  else psNavigate(back, true);
}

function psRefresh(page) {
  if (ps && ps.ctl && ps.ctl.key && ps.ctl.key.indexOf('|' + page + '|') !== -1) { ps.ctl.key = ''; schedule(); }
}

function psKeys(e) {
  if (!ps || wiz || document.querySelector('[data-economy="dialog"]') || panelState) return;
  if (e.key === 'Tab') { trapTab(e, ps.win); return; }
  if (e.key !== 'Escape' || e.defaultPrevented) return;
  if (ps.el.querySelector('.eco-pop:not([hidden]), .eco-qc-pop:not([hidden])')) return;
  e.preventDefault();
  psExit();
}

function psFocusIn(e) {
  if (!ps || wiz || panelState || document.querySelector('[data-economy="dialog"]')) return;
  if (ps.el.contains(e.target) || (e.target.closest && e.target.closest('.eco-toasts'))) return;
  var list = focusables(ps.win).filter(function (el) { return el.getAttribute('tabindex') !== '-1'; });
  var target = ps.list.querySelector('[aria-selected="true"]') || list[0];
  if (target) target.focus({ preventScroll: true });
}

function psListKeys(e) {
  var items = Array.prototype.slice.call(ps.list.querySelectorAll('[role="tab"]'));
  var i = items.indexOf(document.activeElement);
  if (i === -1) return;
  var rtl = (document.documentElement.getAttribute('dir') || 'ltr') === 'rtl';
  var next = null;
  if (e.key === 'ArrowDown' || e.key === (rtl ? 'ArrowLeft' : 'ArrowRight')) next = items[(i + 1) % items.length];
  else if (e.key === 'ArrowUp' || e.key === (rtl ? 'ArrowRight' : 'ArrowLeft')) next = items[(i - 1 + items.length) % items.length];
  else if (e.key === 'Home') next = items[0];
  else if (e.key === 'End') next = items[items.length - 1];
  if (!next) return;
  e.preventDefault();
  next.focus();
  psNavigate(next.getAttribute('href'), true);
}

function psMount() {
  var list = h('div', { class: 'ps-list', role: 'tablist', 'aria-orientation': 'vertical' });
  list.addEventListener('keydown', psListKeys);
  var headIcon = h('span', { class: 'ps-head-icon' });
  var headTitle = h('h3', { id: 'ps-section-title', class: 'ps-head-title' });
  var body = h('div', { id: 'project-settings-body', class: 'ps-body', role: 'tabpanel', 'aria-labelledby': 'ps-section-title', 'data-l10n-skip': '' });
  var scroller = h('div', { class: 'ps-scroll' }, body);
  var title = h('h2', { id: 'ps-title', class: 'ps-title' });
  var close = button({ variant: 'ghost', size: 'iconLg', icon: 'close', title: t('close'), class: 'ps-close', onclick: psExit });
  close.setAttribute('data-ps-close', '');
  var win = h('div', { class: 'ps-window', role: 'dialog', 'aria-modal': 'true', 'aria-labelledby': 'ps-title' },
    h('aside', { class: 'ps-side' },
      h('div', { class: 'ps-side-head' }, h('span', { class: 'ps-mark' }, icon('gear', 'h-5 w-5')), title),
      list),
    h('section', { class: 'ps-main' },
      h('header', { class: 'ps-head' }, headIcon, headTitle),
      scroller),
    close);
  var holder = h('div', { class: 'ps-holder' }, win);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) psExit(); });
  var el = h('div', { 'data-economy': 'project-settings', 'data-l10n-skip': '' }, h('div', { class: 'ps-shade' }), holder);
  ps = { el: el, win: win, list: list, body: body, scroller: scroller, title: title, headIcon: headIcon, headTitle: headTitle, ctl: null, listKey: null, page: null, prevFocus: document.activeElement };
  if (rendered) rendered.id = undefined;
  closePanel();
  document.body.appendChild(el);
  document.documentElement.classList.add('ps-open');
  document.documentElement.setAttribute('data-ps-open', '');
  document.addEventListener('keydown', psKeys, true);
  document.addEventListener('focusin', psFocusIn, true);
}

function psHide() {
  if (!ps) return;
  var old = ps;
  ps = null;
  closePanel();
  old.el.remove();
  document.removeEventListener('keydown', psKeys, true);
  document.removeEventListener('focusin', psFocusIn, true);
  document.documentElement.classList.remove('ps-open');
  document.documentElement.removeAttribute('data-ps-open');
  if (old.prevFocus && old.prevFocus.isConnected && !old.el.contains(old.prevFocus)) { try { old.prevFocus.focus({ preventScroll: true }); } catch (e) { return; } }
}

function psPaintList(allowed, page) {
  var key = lang() + '|' + page + '|' + allowed.map(function (s) { return s.id; }).join(',');
  if (ps.listKey === key) return;
  ps.listKey = key;
  ps.title.textContent = t('ps.title');
  ps.list.setAttribute('aria-label', t('ps.sections'));
  ps.list.setAttribute('aria-orientation', window.matchMedia && window.matchMedia('(max-width: 767px)').matches ? 'horizontal' : 'vertical');
  var focused = ps.list.contains(document.activeElement);
  var have = Array.prototype.map.call(ps.list.children, function (a) { return a.getAttribute('data-ps-section'); }).join(',');
  if (have && have === allowed.map(function (s) { return s.id; }).join(',')) {
    var cur = null;
    Array.prototype.forEach.call(ps.list.children, function (a) {
      var id = a.getAttribute('data-ps-section');
      var on = id === page;
      a.setAttribute('aria-selected', on ? 'true' : 'false');
      a.setAttribute('tabindex', on ? '0' : '-1');
      a.querySelector('span').textContent = t(psSection(id).label);
      if (on) cur = a;
    });
    if (cur && focused && document.activeElement !== cur) cur.focus({ preventScroll: true });
    if (cur && cur.scrollIntoView && ps.list.scrollWidth > ps.list.clientWidth) cur.scrollIntoView({ block: 'nearest', inline: 'nearest' });
    return;
  }
  clear(ps.list);
  var current = null;
  allowed.forEach(function (s) {
    var on = s.id === page;
    var a = h('a', { href: hashFor(s.id), role: 'tab', id: 'ps-tab-' + s.id, class: 'ps-item', 'aria-selected': on ? 'true' : 'false', 'aria-controls': 'project-settings-body', tabindex: on ? '0' : '-1', 'data-ps-section': s.id },
      icon(s.icon, 'h-4 w-4'), h('span', { class: 'truncate', text: t(s.label) }));
    a.addEventListener('click', function (e) {
      if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
      e.preventDefault();
      if (s.id !== (ps && ps.page)) psNavigate(hashFor(s.id), true);
    });
    if (on) current = a;
    ps.list.appendChild(a);
  });
  if (current) {
    if (focused) current.focus({ preventScroll: true });
    if (current.scrollIntoView && ps.list.scrollWidth > ps.list.clientWidth) current.scrollIntoView({ block: 'nearest', inline: 'nearest' });
  }
}

function psShow(route) {
  var allowed = psAllowed();
  var sec = psSection(route.page);
  if (allowed.length && allowed.indexOf(sec) === -1) {
    var to = '#/' + PS_ROUTE + '/' + allowed[0].id;
    if (router) router.replace(to.slice(1)).then(schedule, schedule);
    else { history.replaceState(history.state, '', location.pathname + location.search + to); schedule(); }
    return;
  }
  var fresh = !ps;
  if (fresh) psMount();
  ps.el.setAttribute('dir', document.documentElement.getAttribute('dir') || 'ltr');
  psPaintList(allowed, route.page);
  if (!allowed.length) {
    if (ps.page !== null || fresh) {
      ps.page = null;
      ps.ctl = null;
      clear(ps.headIcon);
      ps.headTitle.textContent = t('ps.title');
      clear(ps.body).appendChild(card(emptyBlock(t('ps.noAccess'), t('ps.noAccess.hint'))));
    }
    if (fresh) ps.win.querySelector('[data-ps-close]').focus();
    return;
  }
  if (ps.page !== route.page || !ps.ctl || ps.headTitle.textContent !== t(sec.label)) {
    clear(ps.headIcon).appendChild(icon(sec.icon, 'h-5 w-5'));
    ps.headTitle.textContent = t(sec.label);
  }
  ps.page = route.page;
  ps.ctl = pageCtl(ps.ctl, ps.body, route, function (ctl) {
    ps.scroller.scrollTop = 0;
    ps.ctl = ctl;
  });
  if (fresh) {
    var tab = ps.list.querySelector('[aria-selected="true"]');
    if (tab) tab.focus({ preventScroll: true });
  }
}

function placePsNav(drawer, collapsed, on, key) {
  var scroll = drawer.querySelector(':scope > .x-scroll');
  var existing = drawer.querySelector(':scope > [data-ps-nav]');
  if (!scroll || !psAllowed().length) {
    if (existing) existing.remove();
    return null;
  }
  if (existing && existing.getAttribute('data-key') === key && existing.previousElementSibling === scroll) return existing;
  var a = h('a', { href: psEntryHash(), class: (on ? NAV_ACTIVE + ' ' : '') + NAV_LINK, 'aria-current': on ? 'page' : null, title: collapsed ? t('ps.title') : null, 'data-ps-entry': 'sidebar', onclick: psEnter });
  a.appendChild(collapsed ? h('div', { class: 'x-tooltip flex' }, icon('gear', 'h-4 w-4')) : icon('gear', 'h-4 w-4'));
  a.appendChild(h('div', { class: 'font-medium relative truncate' + (collapsed ? '' : ' flex-1') }, h('span', { text: collapsed ? ' ' : t('ps.title') })));
  var box = h('div', { class: 'ps-nav', 'data-ps-nav': '', 'data-l10n-skip': '', 'data-key': key }, h('nav', { 'aria-label': t('ps.title') }, a));
  if (existing) existing.replaceWith(box);
  else scroll.after(box);
  return box;
}

var psButton = null;

function dockPsButton() {
  if (!psButton) {
    psButton = h('a', { class: 'economy-ps', 'data-l10n-skip': '', 'data-ps-entry': 'topbar', onclick: psEnter }, icon('gear', 'h-4 w-4'));
  }
  var counter = document.querySelector('number-flow-vue');
  var bar = counter && counter.closest('.justify-end');
  var menu = bar && bar.querySelector(':scope > .x-popover');
  var show = !!menu && !!token() && psAllowed().length > 0;
  if (!show) {
    if (psButton.parentElement) psButton.remove();
    return;
  }
  var href = psEntryHash();
  if (psButton.getAttribute('href') !== href) psButton.setAttribute('href', href);
  if (psButton.getAttribute('aria-label') !== t('ps.title')) {
    psButton.setAttribute('title', t('ps.title'));
    psButton.setAttribute('aria-label', t('ps.title'));
  }
  if (ps) psButton.setAttribute('aria-current', 'page');
  else psButton.removeAttribute('aria-current');
  if (psButton.nextElementSibling !== menu) menu.before(psButton);
}

document.addEventListener('keydown', function (e) {
  if (e.key && e.key.length > 1 && !/^(Shift|Control|Alt|Meta|CapsLock)$/.test(e.key)) document.documentElement.setAttribute('data-eco-kbd', '');
}, true);
document.addEventListener('pointerdown', function () { document.documentElement.removeAttribute('data-eco-kbd'); }, true);
