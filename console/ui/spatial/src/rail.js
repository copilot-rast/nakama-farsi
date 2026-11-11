var DIRECT = ['/', '/players', '/groups', '/storage', '/storage-indexes', '/matches', '/leaderboards', '/notifications', '/payments', '/chat'];
var SETTINGS_PAGES = [['users', 'settings.users', 'users'], ['general', 'settings.general', 'toggle'], ['config', 'settings.config', 'fileCode'], ['data-management', 'settings.data', 'database'], ['audit-log', 'settings.audit', 'list']];
var NATIVE_LABEL = { '/payments': 'nav.purchases', '/chat': 'nav.chat' };
var GROUP_ICON = { economy: 'coins', publication: 'megaphone', moderation: 'shield', support: 'headset', developer: 'code', help: 'help', settings: 'gear' };

function linkInfo(a) {
  var badgeEl = a.querySelector('.eco-nav-badge, .eco-nav-dot');
  var label = '';
  var span = a.querySelector('span:not(.eco-nav-badge)');
  label = (span && span.textContent.trim()) || a.getAttribute('title') || a.textContent.trim();
  if (badgeEl && badgeEl.textContent) label = label.replace(badgeEl.textContent, '').trim();
  var svg = a.querySelector('svg');
  var badge = 0;
  if (badgeEl) {
    var digits = (badgeEl.textContent || '').replace(/[۰-۹]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'.indexOf(d); }).replace(/[^0-9]/g, '');
    badge = parseInt(digits, 10) || (badgeEl.classList.contains('eco-nav-dot') ? 1 : 0);
  }
  var renamed = NATIVE_LABEL[routePath(a.getAttribute('href') || '#')];
  if (renamed) label = st(renamed);
  return { href: a.getAttribute('href') || '', label: label, svg: svg ? svg.outerHTML : '', badge: badge, external: !/^#\//.test(a.getAttribute('href') || '') };
}

function railModel() {
  var drawer = document.querySelector('.x-drawer');
  var grid = drawer && drawer.querySelector('.grid');
  if (!grid) return null;
  var model = { direct: [], groups: [], settings: null, help: null, logo: null };
  var helpLinks = [];
  var devLinks = [];
  var consoleSettings = null;
  Array.prototype.forEach.call(grid.children, function (g) {
    var links = Array.prototype.map.call(g.querySelectorAll('a[href]'), linkInfo).filter(function (l) { return l.href; });
    if (!links.length) return;
    var labelEl = null;
    for (var i = 0; i < g.children.length; i++) if (g.children[i].tagName !== 'NAV' && !g.children[i].querySelector('a')) { labelEl = g.children[i]; break; }
    var label = labelEl ? labelEl.textContent.trim() : '';
    var sec = g.getAttribute('data-economy-nav');
    if (sec) {
      model.groups.push({ id: sec, label: label || st('group.' + sec), icon: GROUP_ICON[sec] || 'page', links: links });
      return;
    }
    if (links.every(function (l) { return l.external; })) { helpLinks = helpLinks.concat(links); return; }
    links.forEach(function (l) {
      if (l.external) { helpLinks.push(l); return; }
      var p = routePath(l.href);
      if (p === '/settings' || p.indexOf('/settings/') === 0) { consoleSettings = l; return; }
      if (DIRECT.indexOf(p) !== -1) model.direct.push(l);
      else devLinks.push(l);
    });
  });
  var footer = drawer.querySelectorAll(':scope > .grid a[href], :scope > div:last-child a[href^="http"]');
  Array.prototype.forEach.call(footer, function (a) {
    var info = linkInfo(a);
    if (info.external && !helpLinks.some(function (h) { return h.href === info.href; })) helpLinks.push(info);
  });
  if (devLinks.length) model.groups.push({ id: 'developer', label: st('group.developer'), icon: 'code', links: devLinks });
  var settingsLinks = [];
  if (consoleSettings) {
    SETTINGS_PAGES.forEach(function (sp) {
      settingsLinks.push({ href: '#/settings/' + sp[0], label: st(sp[1]), svg: spIcon(sp[2]), badge: 0 });
    });
  }
  var psApi = nc() && nc().projectSettings;
  var psEntry = document.querySelector('[data-ps-nav] a[href]');
  if (psApi && psApi.allowed()) settingsLinks.push({ href: psApi.href(), label: st('settings.project'), svg: spIcon('sliders'), badge: 0, ps: true });
  else if (psEntry) settingsLinks.push({ href: psEntry.getAttribute('href'), label: st('settings.project'), svg: spIcon('sliders'), badge: 0, ps: true });
  if (settingsLinks.length) model.settings = { id: 'settings', label: st('group.settings'), icon: 'gear', links: settingsLinks };
  if (helpLinks.length) model.help = { id: 'help', label: st('group.help'), icon: 'help', links: helpLinks };
  var logo = drawer.querySelector('a[href="#/"] img:not([data-brand-hidden])');
  model.logo = logo;
  return model;
}

function groupActive(g, hash) {
  return g.links.some(function (l) { return linkActive(l, hash); });
}

function linkActive(l, hash) {
  if (l.ps) return /^#\/project-settings(\/|$)/.test(hash || currentHash());
  if (/^#\/project-settings/.test(hash || currentHash())) return false;
  return hrefActive(l.href, hash);
}

function groupBadge(g) {
  return g.links.reduce(function (n, l) { return n + (l.badge || 0); }, 0);
}

var rail = { el: null, full: null, fly: null, flyGroup: null, flyBtn: null, closeTimer: 0, key: '', model: null, tip: null, subOpen: null };

function railPinned() {
  try { return localStorage.getItem(SP_RAIL_KEY) === 'pinned'; } catch (e) { return false; }
}

function setPinned(on) {
  try { if (on) localStorage.setItem(SP_RAIL_KEY, 'pinned'); else localStorage.removeItem(SP_RAIL_KEY); } catch (e) { on = !!on; }
  root.classList.toggle('sp-pinned', !!on);
  closeFly(true);
  rail.key = '';
  renderRail();
}

function brandTile() {
  var b = el('a', { class: 'sp-brand', href: '#/', 'aria-label': st('rail.home') });
  var mark = null;
  var info = nc() && nc().brand ? nc().brand() : null;
  if (info && info.markUrl) mark = el('img', { src: info.markUrl, alt: '' });
  if (mark) b.appendChild(mark);
  else b.innerHTML = spIcon('star', 'sp-ic sp-ic-star');
  return b;
}

function railItem(l, cls) {
  var on = linkActive(l);
  var a = el('a', { class: cls || 'sp-ri', href: l.href, 'aria-label': l.label, 'data-sp-tip': l.label, 'aria-current': on ? 'page' : null, 'data-sp-href': l.href });
  a.innerHTML = l.svg || spIcon('page');
  if (cls === 'sp-row') a.appendChild(el('span', { class: 'sp-row-label', text: l.label }));
  if (l.badge) a.appendChild(el('span', { class: cls === 'sp-row' ? 'sp-chip sp-chip-accent' : 'sp-count', text: spNum(l.badge) }));
  if (l.ps) { a.addEventListener('click', psClick); a.setAttribute('data-ps-entry', 'rail'); }
  return a;
}

function psClick(e) {
  if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
  var api = nc() && nc().projectSettings;
  if (!api) return;
  e.preventDefault();
  closeFly(true);
  closeSheet();
  api.open();
}

function groupButton(g, cls) {
  var on = groupActive(g);
  var b = el('button', { type: 'button', class: cls || 'sp-ri sp-ri-group', 'aria-haspopup': 'menu', 'aria-expanded': 'false', 'aria-label': g.label, 'data-sp-group': g.id, 'data-sp-tip': cls ? null : g.label, 'aria-current': on ? 'true' : null });
  b.innerHTML = spIcon(g.icon);
  var sum = groupBadge(g);
  if (cls === 'sp-row sp-row-group') {
    b.appendChild(el('span', { class: 'sp-row-label', text: g.label }));
    b.appendChild(el('span', { class: sum ? 'sp-chip sp-chip-accent' : 'sp-chip', text: spNum(sum || g.links.length) }));
    b.insertAdjacentHTML('beforeend', spIcon('chevron', 'sp-ic sp-ic-dir sp-row-chev'));
  } else {
    b.appendChild(el('span', { class: 'sp-more', 'aria-hidden': 'true' }));
    if (sum) b.appendChild(el('span', { class: 'sp-count', text: spNum(sum) }));
  }
  b.addEventListener('mouseenter', function () { if (!mqMobile.matches) openFly(b, g, false); });
  b.addEventListener('mouseleave', scheduleClose);
  b.addEventListener('focus', function () { if (!b.__spNoFocusOpen && !mqMobile.matches) openFly(b, g, false); b.__spNoFocusOpen = false; });
  b.addEventListener('click', function () {
    if (mqMobile.matches && rail.full && rail.full.contains(b)) { inlineGroup(b, g, false); return; }
    if (rail.fly && rail.fly.classList.contains('open') && rail.flyGroup && rail.flyGroup.id === g.id && rail.flyBtn === b) { if (!b.__spJustOpened) closeFly(); }
    else openFly(b, g, false);
  });
  b.addEventListener('keydown', function (e) {
    var into = isRtl() ? 'ArrowLeft' : 'ArrowRight';
    if (e.key === 'Enter' || e.key === ' ' || e.key === into || e.key === 'ArrowDown' && b.closest('.sp-rail-full')) {
      e.preventDefault();
      openFly(b, g, true);
    } else if (e.key === 'Escape') {
      closeFly();
    }
  });
  return b;
}

function ensureFly() {
  if (rail.fly) return rail.fly;
  var f = el('div', { class: 'sp-flyout', role: 'menu', 'data-l10n-skip': '' });
  f.addEventListener('mouseenter', function () { clearTimeout(rail.closeTimer); });
  f.addEventListener('mouseleave', scheduleClose);
  f.addEventListener('keydown', function (e) {
    var items = Array.prototype.slice.call(f.querySelectorAll('[role="menuitem"]'));
    var i = items.indexOf(document.activeElement);
    var back = isRtl() ? 'ArrowRight' : 'ArrowLeft';
    if (e.key === 'ArrowDown') { e.preventDefault(); items[(i + 1) % items.length].focus(); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); items[(i - 1 + items.length) % items.length].focus(); }
    else if (e.key === 'Home') { e.preventDefault(); items[0].focus(); }
    else if (e.key === 'End') { e.preventDefault(); items[items.length - 1].focus(); }
    else if (e.key === 'Escape' || e.key === back) {
      e.preventDefault();
      var btn = rail.flyBtn;
      closeFly();
      if (btn) { btn.__spNoFocusOpen = true; btn.focus(); }
    } else if (e.key === 'Tab') {
      closeFly();
    }
  });
  document.body.appendChild(f);
  rail.fly = f;
  return f;
}

function paintFly(g) {
  var f = rail.fly;
  f.innerHTML = '';
  f.setAttribute('aria-label', g.label);
  f.setAttribute('data-group', g.id);
  f.setAttribute('dir', root.getAttribute('dir') || 'ltr');
  f.appendChild(el('div', { class: 'sp-fly-head' }, el('b', { text: g.label }), el('span', { text: st('rail.pages', { n: g.links.length }) })));
  g.links.forEach(function (l) {
    var on = linkActive(l);
    var a = el('a', { role: 'menuitem', href: l.href, tabindex: '-1', 'aria-current': on ? 'page' : null, target: l.external ? '_blank' : null, rel: l.external ? 'noopener' : null });
    a.innerHTML = l.svg || spIcon(l.external ? 'external' : 'page');
    a.appendChild(el('span', { class: 'sp-fly-label', text: l.label }));
    if (l.badge) a.appendChild(el('em', { class: 'sp-fly-count', text: spNum(l.badge) }));
    if (l.ps) { a.addEventListener('click', psClick); a.setAttribute('data-ps-entry', 'rail'); }
    else a.addEventListener('click', function () { closeFly(true); closeSheet(); });
    f.appendChild(a);
  });
}

function inlineGroup(btn, g, focusFirst) {
  var box = btn.parentNode;
  var open = box.querySelector(':scope > .sp-sub[data-group="' + g.id + '"]');
  Array.prototype.forEach.call(rail.full.querySelectorAll('.sp-sub'), function (s) {
    if (s === open) return;
    var b = s.previousElementSibling;
    if (b) b.setAttribute('aria-expanded', 'false');
    s.remove();
  });
  if (open && !focusFirst) {
    open.remove();
    btn.setAttribute('aria-expanded', 'false');
    rail.subOpen = null;
    return;
  }
  if (!open) {
    open = el('div', { class: 'sp-sub', role: 'group', 'aria-label': g.label, 'data-group': g.id });
    g.links.forEach(function (l) {
      var a = railItem(l, 'sp-row');
      a.classList.add('sp-sub-row');
      if (l.external) { a.setAttribute('target', '_blank'); a.setAttribute('rel', 'noopener'); }
      if (!l.ps) a.addEventListener('click', function () { closeSheet(); });
      open.appendChild(a);
    });
    btn.after(open);
  }
  btn.setAttribute('aria-expanded', 'true');
  rail.subOpen = g.id;
  if (focusFirst) {
    var first = open.querySelector('[aria-current="page"]') || open.querySelector('a');
    if (first) first.focus();
  }
}

function openFly(btn, g, focusFirst) {
  if (mqMobile.matches && rail.full && rail.full.contains(btn)) { inlineGroup(btn, g, focusFirst); return; }
  clearTimeout(rail.closeTimer);
  var f = ensureFly();
  if (rail.flyBtn && rail.flyBtn !== btn) rail.flyBtn.setAttribute('aria-expanded', 'false');
  if (!rail.flyGroup || rail.flyGroup.id !== g.id || rail.flyBtn !== btn || f.getAttribute('data-key') !== rail.key) {
    paintFly(g);
    f.setAttribute('data-key', rail.key);
  }
  rail.flyGroup = g;
  rail.flyBtn = btn;
  hideTip();
  var r = btn.getBoundingClientRect();
  var pinned = !!btn.closest('.sp-rail-full');
  f.classList.toggle('sp-fly-pinned', pinned);
  var h = f.offsetHeight || (56 + g.links.length * 40);
  var top = Math.max(8, Math.min(r.top - 8, window.innerHeight - h - 8));
  f.style.top = top + 'px';
  var was = f.classList.contains('open');
  f.classList.add('open');
  btn.setAttribute('aria-expanded', 'true');
  if (!was) { btn.__spJustOpened = true; setTimeout(function () { btn.__spJustOpened = false; }, 250); }
  if (focusFirst) {
    var first = f.querySelector('[aria-current="page"]') || f.querySelector('[role="menuitem"]');
    if (first) first.focus();
  }
}

function closeFly(now) {
  clearTimeout(rail.closeTimer);
  if (!rail.fly) return;
  rail.fly.classList.remove('open');
  if (now) rail.fly.classList.add('sp-fly-now');
  setTimeout(function () { if (rail.fly) rail.fly.classList.remove('sp-fly-now'); }, 30);
  if (rail.flyBtn) rail.flyBtn.setAttribute('aria-expanded', 'false');
  rail.flyGroup = null;
  rail.flyBtn = null;
}

function scheduleClose() {
  clearTimeout(rail.closeTimer);
  rail.closeTimer = setTimeout(function () {
    if (rail.fly && rail.fly.contains(document.activeElement)) return;
    closeFly();
  }, 160);
}

function ensureTip() {
  if (rail.tip) return rail.tip;
  rail.tip = el('div', { class: 'sp-tip', role: 'tooltip', 'aria-hidden': 'true', 'data-l10n-skip': '' });
  document.body.appendChild(rail.tip);
  return rail.tip;
}

function showTip(target) {
  var text = target.getAttribute('data-sp-tip');
  if (!text || (rail.fly && rail.fly.classList.contains('open') && target.hasAttribute('data-sp-group'))) return;
  var tip = ensureTip();
  tip.textContent = text;
  var r = target.getBoundingClientRect();
  tip.style.top = Math.round(r.top + r.height / 2) + 'px';
  tip.setAttribute('dir', root.getAttribute('dir') || 'ltr');
  tip.classList.add('open');
}

function hideTip() {
  if (rail.tip) rail.tip.classList.remove('open');
}

function railKey(model) {
  var b = nc() && nc().brand ? nc().brand() : null;
  return [spLang(), currentHash(), railPinned(), root.classList.contains('dark'), JSON.stringify(b), JSON.stringify([model.direct, model.groups, model.settings, model.help].map(function (x) {
    return x ? JSON.stringify(x.links || x, function (k, v) { return k === 'svg' ? undefined : v; }) : '';
  }))].join('|');
}

function buildCollapsed(model) {
  var nav = rail.el;
  nav.innerHTML = '';
  nav.appendChild(brandTile());
  nav.appendChild(el('button', { type: 'button', class: 'sp-ri sp-pin', 'aria-label': st('rail.pin'), 'data-sp-tip': st('rail.pin'), html: spIcon('panel'), onclick: function () { setPinned(true); } }));
  nav.appendChild(el('hr'));
  model.direct.forEach(function (l) { nav.appendChild(railItem(l)); });
  if (model.groups.length) nav.appendChild(el('hr'));
  model.groups.forEach(function (g) { nav.appendChild(groupButton(g)); });
  nav.appendChild(el('div', { class: 'sp-grow' }));
  if (model.help) nav.appendChild(groupButton(model.help));
  if (model.settings) nav.appendChild(groupButton(model.settings));
}

function buildFull(model) {
  var p = rail.full;
  p.innerHTML = '';
  var info = nc() && nc().brand ? nc().brand() : null;
  var name = (info && (info.name || info.shortName)) || 'Nakama';
  var logo = info && (root.classList.contains('dark') ? info.logoDarkUrl || info.logoUrl : info.logoUrl);
  var title = logo ? el('img', { class: 'sp-full-logo', src: logo, alt: name }) : el('b', { text: name });
  var head = el('div', { class: 'sp-full-head' }, brandTile(), el('div', { class: 'sp-full-name' }, title, el('span', { text: st('rail.console') })),
    el('button', { type: 'button', class: 'sp-collapse', 'aria-label': st('rail.unpin'), title: st('rail.unpin'), html: spIcon('panel'), onclick: function () { setPinned(false); } }));
  p.appendChild(head);
  var scroll = el('div', { class: 'sp-full-scroll' });
  model.direct.forEach(function (l) { scroll.appendChild(railItem(l, 'sp-row')); });
  if (model.groups.length) scroll.appendChild(el('div', { class: 'sp-full-label', text: st('rail.sections') }));
  model.groups.forEach(function (g) { scroll.appendChild(groupButton(g, 'sp-row sp-row-group')); });
  p.appendChild(scroll);
  var foot = el('div', { class: 'sp-full-foot' });
  if (model.help) foot.appendChild(groupButton(model.help, 'sp-row sp-row-group'));
  if (model.settings) foot.appendChild(groupButton(model.settings, 'sp-row sp-row-group'));
  p.appendChild(foot);
  if (rail.subOpen && mqMobile.matches && root.classList.contains('sp-sheet-open')) {
    var again = p.querySelector('[data-sp-group="' + rail.subOpen + '"]');
    var grp = [].concat(model.groups, model.help ? [model.help] : [], model.settings ? [model.settings] : []).filter(function (x) { return x.id === rail.subOpen; })[0];
    if (again && grp) inlineGroup(again, grp, false);
  } else rail.subOpen = null;
}

function renderRail(force) {
  if (root.classList.contains('sp-auth')) return;
  var model = railModel();
  if (!model) {
    root.classList.remove('sp-railed');
    return;
  }
  var key = railKey(model);
  if (!force && key === rail.key && rail.el && rail.el.isConnected && rail.full && rail.full.isConnected) return;
  rail.key = key;
  rail.model = model;
  if (!rail.el) {
    rail.el = el('nav', { class: 'sp-rail', 'data-l10n-skip': '' });
    rail.el.addEventListener('mouseover', function (e) { var t = e.target.closest('[data-sp-tip]'); if (t && rail.el.contains(t)) showTip(t); });
    rail.el.addEventListener('mouseout', function (e) { var t = e.target.closest('[data-sp-tip]'); if (t && !t.contains(e.relatedTarget)) hideTip(); });
    rail.el.addEventListener('focusin', function (e) { var t = e.target.closest('[data-sp-tip]'); if (t && e.target.matches(':focus-visible')) showTip(t); });
    rail.el.addEventListener('focusout', hideTip);
    rail.el.addEventListener('keydown', railKeys);
    rail.full = el('nav', { class: 'sp-rail-full', 'data-l10n-skip': '' });
    rail.full.addEventListener('keydown', railKeys);
  }
  rail.el.setAttribute('aria-label', st('rail.label'));
  rail.full.setAttribute('aria-label', st('rail.label'));
  var act = document.activeElement;
  var keep = null;
  if (act && (rail.el.contains(act) || rail.full.contains(act))) {
    keep = { full: rail.full.contains(act), href: act.getAttribute('data-sp-href'), group: act.getAttribute('data-sp-group'), cls: act.className };
  }
  buildCollapsed(model);
  buildFull(model);
  if (keep) {
    var box = keep.full ? rail.full : rail.el;
    var back = keep.href ? box.querySelector('[data-sp-href="' + keep.href.replace(/"/g, '') + '"]') : keep.group ? box.querySelector('[data-sp-group="' + keep.group + '"]') : box.querySelector('.' + String(keep.cls).split(' ')[0]);
    if (back) { back.__spNoFocusOpen = true; back.focus({ preventScroll: true }); }
  }
  rail.el.setAttribute('aria-hidden', railPinned() ? 'true' : 'false');
  rail.el.toggleAttribute('inert', railPinned());
  rail.full.setAttribute('aria-hidden', railPinned() ? 'false' : 'true');
  rail.full.toggleAttribute('inert', !railPinned());
  if (rail.el.parentNode !== document.body) document.body.appendChild(rail.el);
  if (rail.full.parentNode !== document.body) document.body.appendChild(rail.full);
  root.classList.add('sp-railed');
  root.classList.toggle('sp-pinned', railPinned());
  if (rail.fly && rail.fly.classList.contains('open') && rail.flyGroup) {
    var g = [].concat(model.groups, model.help ? [model.help] : [], model.settings ? [model.settings] : []).filter(function (x) { return x.id === rail.flyGroup.id; })[0];
    if (g) paintFly(g);
  }
  if (typeof paintTopbar === 'function') paintTopbar();
}

function railKeys(e) {
  if (e.key !== 'ArrowDown' && e.key !== 'ArrowUp') return;
  if (e.target.closest('.sp-flyout')) return;
  if (e.key === 'ArrowDown' && e.target.closest('.sp-rail-full') && e.target.hasAttribute('data-sp-group')) return;
  var box = e.currentTarget;
  var items = Array.prototype.slice.call(box.querySelectorAll('a.sp-ri, button.sp-ri, a.sp-row, button.sp-row, .sp-brand, .sp-collapse'));
  var i = items.indexOf(document.activeElement);
  if (i === -1) return;
  e.preventDefault();
  var next = items[(i + (e.key === 'ArrowDown' ? 1 : -1) + items.length) % items.length];
  next.__spNoFocusOpen = true;
  next.focus();
}

function allPages() {
  var m = railModel() || rail.model;
  if (!m) return [];
  var out = [];
  m.direct.forEach(function (l) { out.push({ label: l.label, href: l.href, svg: l.svg, group: '' }); });
  m.groups.forEach(function (g) { g.links.forEach(function (l) { out.push({ label: l.label, href: l.href, svg: l.svg, group: g.label }); }); });
  if (m.settings) m.settings.links.forEach(function (l) { out.push({ label: l.label, href: l.href, svg: l.svg, group: m.settings.label, ps: l.ps }); });
  var api = nc() && nc().projectSettings;
  if (api && api.allowed()) api.sections().forEach(function (s) { out.push({ label: s.label, href: s.href, group: st('settings.project'), psSection: true }); });
  return out;
}

function pageInfo(hash) {
  var m = rail.model || railModel();
  if (!m) return null;
  var best = null;
  var consider = function (l, g) {
    if (l.external || !linkActive(l, hash)) return;
    var len = l.href.split('?')[0].length;
    if (!best || len > best.len) best = { link: l, group: g, len: len };
  };
  m.direct.forEach(function (l) { consider(l, null); });
  m.groups.forEach(function (g) { g.links.forEach(function (l) { consider(l, g); }); });
  if (m.settings) m.settings.links.forEach(function (l) { consider(l, m.settings); });
  return best;
}

var sheetPrev = null;

function openSheet() {
  if (root.classList.contains('sp-sheet-open')) return;
  renderRail(true);
  sheetPrev = document.activeElement;
  root.classList.add('sp-sheet-open');
  rail.full.removeAttribute('inert');
  rail.full.setAttribute('aria-hidden', 'false');
  rail.full.setAttribute('role', 'dialog');
  rail.full.setAttribute('aria-modal', 'true');
  var scrim = document.querySelector('.sp-sheet-scrim');
  if (!scrim) {
    scrim = el('div', { class: 'sp-sheet-scrim', 'aria-hidden': 'true', onclick: closeSheet });
    document.body.appendChild(scrim);
  }
  var first = rail.full.querySelector('[aria-current="page"]') || rail.full.querySelector('a, button');
  if (first) { first.__spNoFocusOpen = true; first.focus({ preventScroll: true }); }
  document.addEventListener('keydown', sheetKeys, true);
}

function closeSheet() {
  if (!root.classList.contains('sp-sheet-open')) return;
  root.classList.remove('sp-sheet-open');
  rail.subOpen = null;
  Array.prototype.forEach.call(rail.full.querySelectorAll('.sp-sub'), function (sub) {
    var b = sub.previousElementSibling;
    if (b) b.setAttribute('aria-expanded', 'false');
    sub.remove();
  });
  rail.full.removeAttribute('role');
  rail.full.removeAttribute('aria-modal');
  if (!railPinned() || mqMobile.matches) { rail.full.setAttribute('inert', ''); rail.full.setAttribute('aria-hidden', 'true'); }
  document.removeEventListener('keydown', sheetKeys, true);
  closeFly(true);
  if (sheetPrev && sheetPrev.isConnected) { try { sheetPrev.focus({ preventScroll: true }); } catch (e) { sheetPrev = null; } }
}

function sheetKeys(e) {
  if (e.key === 'Escape' && !(rail.fly && rail.fly.classList.contains('open'))) { e.preventDefault(); closeSheet(); return; }
  if (e.key !== 'Tab') return;
  var list = Array.prototype.slice.call(rail.full.querySelectorAll('a[href], button')).filter(function (x) { return x.offsetParent !== null; });
  if (rail.fly && rail.fly.classList.contains('open')) list = list.concat(Array.prototype.slice.call(rail.fly.querySelectorAll('a')));
  if (!list.length) return;
  var i = list.indexOf(document.activeElement);
  if (e.shiftKey && i <= 0) { e.preventDefault(); list[list.length - 1].focus(); }
  else if (!e.shiftKey && i === list.length - 1) { e.preventDefault(); list[0].focus(); }
}
