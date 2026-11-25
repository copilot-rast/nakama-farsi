var DIRECT = ['/', '/players', '/groups', '/storage', '/storage-indexes', '/matches', '/leaderboards', '/notifications', '/payments', '/chat'];
var SETTINGS_PAGES = [['users', 'settings.users', 'users'], ['general', 'settings.general', 'toggle'], ['config', 'settings.config', 'fileCode'], ['data-management', 'settings.data', 'database'], ['audit-log', 'settings.audit', 'list']];
var NATIVE_LABEL = { '/payments': 'nav.purchases', '/chat': 'nav.chat' };
var HREF_ICON = { '': 'dashboard', players: 'user', groups: 'users', storage: 'database', 'storage-indexes': 'list', matches: 'gamepad', leaderboards: 'trophy', notifications: 'bell', payments: 'card', chat: 'message', apiexplorer: 'terminal', 'api-explorer': 'terminal', modules: 'cpu', runtime: 'cpu', overview: 'dashboard', currencies: 'coins', metas: 'chart', prizes: 'gift', products: 'shopping', offers: 'tag', 'entry-offers': 'clock', 'offer-preview': 'eye', 'offer-report': 'chart', 'offer-settings': 'sliders', 'tag-discounts': 'tag', 'tag-discount-preview': 'eye', 'tag-discount-settings': 'sliders', 'friend-gifts': 'handHeart', 'friend-gift-rules': 'shield', 'friend-gift-settings': 'sliders', 'friend-gift-report': 'chart', gifts: 'ticket', 'event-types': 'zap', 'event-rules': 'trophy', 'meta-rules': 'sliders', 'meta-relations': 'share', 'level-tables': 'ladder', streaks: 'flame', tasks: 'checklist', 'task-groups': 'layers', items: 'backpack', achievements: 'medal', 'chance-games': 'dice', 'prize-groups': 'package', events: 'zap', games: 'puzzle', updates: 'gitMerge', 'event-log': 'list', referrals: 'share', faqs: 'help', news: 'newspaper', banners: 'flagBanner', 'banner-lists': 'slideshow', sponsors: 'handshake', 'auto-rules': 'zap', 'notify-categories': 'tag', comments: 'message', 'remote-config': 'sliders', reports: 'flag', 'player-reports': 'user', restrictions: 'micoff', 'chat-words': 'type', 'chat-quota': 'message', 'chat-quota-settings': 'sliders', 'chat-quota-report': 'chart', activity: 'activity', tickets: 'ticket', 'live-chat': 'headset', users: 'users', general: 'toggle', config: 'fileCode', 'data-management': 'database', 'audit-log': 'list', 'project-settings': 'settings', content: 'cloudDown', 'content-packs': 'cloudDown', 'content-settings': 'sliders', insights: 'chart', 'insights-overview': 'chart', 'insights-sessions': 'activity', heatmaps: 'flame', 'insights-settings': 'sliders' };

function iconFor(href) {
  if (!/^#\//.test(href || '')) return 'open';
  var parts = routePath(href).split('/').filter(Boolean);
  if (parts[0] === 'publication' && parts[1] === 'notifications') return 'send';
  if (parts[0] === 'economy' && parts[1] === 'payments') return 'wallet';
  if (parts[0] === 'moderation' && parts[1] === 'players') return 'user';
  if (parts[0] === 'quiz') return QUIZ_ICON[parts[1] || 'packs'] || 'help';
  if (parts[0] === 'trophies') return TROPHY_ICON[parts[1] || 'leagues'] || 'trophy';
  if (parts[0] === 'clans') return CLAN_ICON[parts[1] || 'list'] || 'group';
  if (parts[0] === 'cups') return CUP_ICON[parts[1] || 'list'] || 'medal';
  if (parts[0] === 'itemkinds') return IK_ICON[parts[1] || 'kind'] || 'package';
  var key = parts.length > 1 && /^(economy|publication|moderation|support|system|settings)$/.test(parts[0]) ? parts[1] : parts[0] || '';
  return HREF_ICON[key] || 'page';
}

window.NakamaPlane.iconFor = iconFor;

var QUIZ_ICON = { packs: 'package', questions: 'help', missions: 'checklist', duels: 'users', 'duel-categories': 'layers', reports: 'flag', results: 'fileCode', 'import': 'cloudDown', stats: 'chart', settings: 'sliders' };
var TROPHY_ICON = { leagues: 'ladder', players: 'trophy', seasons: 'calendar', settings: 'sliders' };
var CLAN_ICON = { list: 'group', levels: 'layers', wars: 'shield', rounds: 'repeat', gallery: 'image', uploads: 'inbox', settings: 'sliders' };
var CUP_ICON = { list: 'medal', view: 'medal', templates: 'repeat', settings: 'sliders' };
var IK_ICON = { kind: 'package', routes: 'repeat', runs: 'list', settings: 'sliders' };
var GROUP_ICON = { quiz: 'help', cups: 'medal', itemkinds: 'package', trophies: 'trophy', clans: 'group', economy: 'coins', publication: 'megaphone', moderation: 'shield', support: 'buoy', system: 'cloudDown', developer: 'code', help: 'help', settings: 'settings' };

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
  if (renamed) label = pst(renamed);
  return { href: a.getAttribute('href') || '', label: label, svg: svg ? svg.outerHTML : '', icon: iconFor(a.getAttribute('href')), badge: badge, external: !/^#\//.test(a.getAttribute('href') || ''), group: a.getAttribute('data-economy-group') || '', groupLabel: a.getAttribute('data-economy-group-label') || '' };
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
    var sec = g.getAttribute('data-economy-nav') || g.getAttribute('data-plane-group');
    if (sec) {
      model.groups.push({ id: sec, label: label || pst('group.' + sec), icon: GROUP_ICON[sec] || 'page', links: links });
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
  if (devLinks.length) model.groups.push({ id: 'developer', label: pst('group.developer'), icon: 'code', links: devLinks });
  var settingsLinks = [];
  if (consoleSettings) {
    SETTINGS_PAGES.forEach(function (sp) {
      settingsLinks.push({ href: '#/settings/' + sp[0], label: pst(sp[1]), icon: sp[2], badge: 0 });
    });
  }
  var psApi = nc() && nc().projectSettings;
  var psEntry = document.querySelector('[data-ps-nav] a[href]');
  if (psApi && psApi.allowed()) settingsLinks.push({ href: psApi.href(), label: pst('settings.project'), icon: 'settings', badge: 0, ps: true });
  else if (psEntry) settingsLinks.push({ href: psEntry.getAttribute('href'), label: pst('settings.project'), icon: 'settings', badge: 0, ps: true });
  if (settingsLinks.length) model.settings = { id: 'settings', label: pst('group.settings'), icon: 'settings', links: settingsLinks };
  if (helpLinks.length) model.help = { id: 'help', label: pst('group.help'), icon: 'help', links: helpLinks };
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

function psClick(e) {
  if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
  var api = nc() && nc().projectSettings;
  if (!api) return;
  e.preventDefault();
  closeFly(true);
  closeSheet();
  api.open();
}

function allPages() {
  var m = railModel() || rail.model;
  if (!m) return [];
  var out = [];
  m.direct.forEach(function (l) { out.push({ label: l.label, href: l.href, icon: l.icon, group: '' }); });
  var hubs = nc() && nc().pageTabs ? nc().pageTabs() : [];
  m.groups.forEach(function (g) {
    g.links.forEach(function (l) {
      out.push({ label: l.label, href: l.href, icon: l.icon, group: l.groupLabel ? g.label + ' · ' + l.groupLabel : g.label });
      var hub = hubOf(hubs, l.href);
      if (hub) hub.tabs.forEach(function (tab) {
        if (tab.first && tab.label === hub.label) return;
        out.push({ label: tab.label, href: tab.href, icon: HREF_ICON[tab.id] || l.icon, group: g.label + ' · ' + hub.label, tab: true });
      });
    });
  });
  if (m.settings) m.settings.links.forEach(function (l) { out.push({ label: l.label, href: l.href, icon: l.icon, group: m.settings.label, ps: l.ps }); });
  if (nc() && typeof nc().searchItems === 'function') nc().searchItems().forEach(function (it) { if (it && it.label && it.href) out.push({ label: it.label, href: it.href, icon: it.icon, group: it.group || '' }); });
  var api = nc() && nc().projectSettings;
  if (api && api.allowed()) api.sections().forEach(function (s) { out.push({ label: s.label, href: s.href, group: pst('settings.project'), psSection: true }); });
  var D = window.NakamaDate;
  if (D) CAL_IDS.forEach(function (id) {
    out.push({ label: pst('palette.calendar', { name: D.t('cal.' + id) }), href: '#' + location.hash.slice(1), icon: 'calendar', group: pst('palette.group.settings'), action: function () { if (id !== D.calendarPref()) D.setCalendarPref(id); } });
  });
  return out;
}

function hubOf(hubs, href) {
  var base = (href || '').split('?')[0];
  return hubs.filter(function (hb) { return base === hb.href || base.indexOf(hb.href + '/') === 0; })[0] || null;
}

function tabInfo(hash) {
  var api = nc();
  if (!api || !api.pageTabs) return null;
  var path = (hash || currentHash()).split('?')[0];
  var best = null;
  api.pageTabs().forEach(function (hb) {
    if (path !== hb.href && path.indexOf(hb.href + '/') !== 0) return;
    hb.tabs.forEach(function (tab) {
      if (path === tab.href || path.indexOf(tab.href + '/') === 0) {
        if (!best || tab.href.length > best.tab.href.length) best = { hub: hb, tab: tab };
      }
    });
  });
  return best;
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


function groupBadge(g) {
  return g.links.reduce(function (n, l) { return n + (l.badge || 0); }, 0);
}

var rail = { el: null, side: null, key: '', model: null, tip: null, mode: null, view: null, viewHash: '' };

function modesOf(model) {
  var out = [];
  if (model.direct.length) {
    if (!model.game) model.game = { id: 'game', label: pst('group.game'), icon: 'home', links: model.direct };
    out.push(model.game);
  }
  model.groups.forEach(function (g) { out.push(g); });
  return out;
}

function activeMode(model) {
  var all = modesOf(model).concat(model.settings ? [model.settings] : []);
  var hit = all.filter(function (g) { return groupActive(g); })[0];
  return hit || all[0];
}

function brandInfo() {
  var info = nc() && nc().brand ? nc().brand() : null;
  return info || {};
}

function brandIcon() {
  var api = nc();
  if (!api || !api.brandMark) return null;
  var icon = api.brandMark();
  icon.classList.add('pl-app-icon');
  icon.querySelectorAll('.brand-signin-initial').forEach(function (s) { s.textContent = plNum(s.textContent); });
  return icon;
}

function brandMark(cls) {
  var icon = brandIcon();
  if (!icon) return null;
  var b = el('a', { class: cls || 'pl-brand', href: '#/', 'aria-label': pst('rail.home'), 'data-pl-tip': brandLabel() || pst('rail.home') });
  b.appendChild(icon);
  return b;
}

function brandLabel() {
  var info = brandInfo();
  return plNum(String(info.name || info.shortName || '').trim());
}

function modeItem(g, on) {
  var first = g.links.filter(function (l) { return !l.external; })[0] || g.links[0];
  var a = el('a', { class: 'pl-mode', href: first.href, 'aria-label': g.label, 'data-pl-tip': g.label, 'aria-current': on ? 'true' : null, 'data-pl-mode': g.id });
  a.innerHTML = plIconD(g.icon || 'page');
  if (groupBadge(g)) a.appendChild(el('span', { class: 'pl-pip', 'aria-hidden': 'true' }));
  if (first.ps) a.addEventListener('click', psClick);
  return a;
}

function sideLink(l) {
  var on = linkActive(l);
  var a = el('a', { class: 'pl-nav', href: l.href, 'aria-current': on ? 'page' : null, 'data-pl-href': l.href, target: l.external ? '_blank' : null, rel: l.external ? 'noopener' : null });
  a.innerHTML = plIcon(l.icon || (l.external ? 'external' : 'page'));
  a.appendChild(el('span', { class: 'pl-nav-label', text: l.label }));
  if (l.badge) a.appendChild(el('span', { class: 'pl-badge', text: plNum(l.badge) }));
  if (l.ps) { a.addEventListener('click', psClick); a.setAttribute('data-ps-entry', 'rail'); }
  else a.addEventListener('click', function () { closeSheet(); });
  return a;
}

function sideGroupOpen(mode, id) {
  var api = nc() && nc().navGroups;
  return api ? api.isOpen(mode.id, id) : true;
}

function sideLinks(box, mode) {
  var cur = null;
  mode.links.forEach(function (l) {
    if (!l.group) { cur = null; box.appendChild(sideLink(l)); return; }
    if (!cur || cur.id !== l.group) {
      var open = sideGroupOpen(mode, l.group);
      var bodyId = 'pl-grp-' + mode.id + '-' + l.group;
      var head = el('button', { type: 'button', class: 'pl-group-head', 'aria-expanded': open ? 'true' : 'false', 'aria-controls': bodyId, 'data-pl-group': l.group });
      head.innerHTML = plIcon('down', 'pl-ic pl-group-caret');
      head.appendChild(el('span', { class: 'pl-group-label', text: l.groupLabel }));
      var count = el('span', { class: 'pl-badge pl-group-count', hidden: true });
      head.appendChild(count);
      var body = el('div', { class: 'pl-group-body', id: bodyId, role: 'group', 'aria-label': l.groupLabel, hidden: open ? null : '' });
      cur = { id: l.group, head: head, body: body, count: count, badge: 0 };
      var mine = cur;
      head.addEventListener('click', function () {
        var next = head.getAttribute('aria-expanded') !== 'true';
        head.setAttribute('aria-expanded', next ? 'true' : 'false');
        body.hidden = !next;
        mine.count.hidden = next || !mine.badge;
        var api = nc() && nc().navGroups;
        if (api) api.set(mode.id, l.group, next);
      });
      box.appendChild(el('div', { class: 'pl-group', 'data-pl-group': l.group }, head, body));
    }
    cur.body.appendChild(sideLink(l));
    cur.badge += l.badge || 0;
    if (cur.badge) cur.count.textContent = plNum(cur.badge);
    cur.count.hidden = !cur.body.hidden || !cur.badge;
  });
}

function buildRail(model, mode) {
  var nav = rail.el;
  nav.innerHTML = '';
  var mark = compactOn() ? brandMark() : null;
  if (mark) nav.appendChild(mark);
  modesOf(model).forEach(function (g) { nav.appendChild(modeItem(g, g === mode)); });
  nav.appendChild(el('div', { class: 'pl-grow' }));
  if (model.help) {
    var h = el('button', { type: 'button', class: 'pl-mode', 'aria-label': model.help.label, 'data-pl-tip': model.help.label, 'aria-current': mode === model.help ? 'true' : null, html: plIconD('help') });
    h.addEventListener('click', function () { rail.view = rail.view === 'help' ? null : 'help'; rail.viewHash = currentHash(); rail.key = ''; renderRail(true); });
    nav.appendChild(h);
  }
  if (model.settings) {
    var st = el('button', { type: 'button', class: 'pl-mode', 'aria-label': model.settings.label, 'data-pl-tip': model.settings.label, 'data-pl-mode': 'settings', 'aria-current': mode === model.settings ? 'true' : null, html: plIconD('settings') });
    st.addEventListener('click', function () { rail.view = rail.view === 'settings' ? null : 'settings'; rail.viewHash = currentHash(); rail.key = ''; renderRail(true); });
    nav.appendChild(st);
  }
}

function buildSide(model, mode) {
  var p = rail.side;
  p.innerHTML = '';
  var name = brandLabel();
  var ws = el('a', { class: 'pl-ws', href: '#/', 'data-pl-ws': name ? 'brand' : 'neutral' }, brandIcon(),
    el('span', { class: 'pl-ws-name' }, name ? el('b', { text: name, title: name }) : null, el('span', { text: pst('rail.console') })));
  var toggle = el('button', { type: 'button', class: 'pl-ctl pl-ctl-icon pl-side-toggle', 'aria-label': pst('rail.collapse'), title: pst('rail.collapse'), 'aria-expanded': 'true', html: plIcon('panel'), onclick: function () { setCompact(true); } });
  p.appendChild(el('div', { class: 'pl-side-head' }, ws, toggle));
  var search = el('button', { type: 'button', class: 'pl-side-search', 'aria-haspopup': 'dialog', onclick: function () { closeSheet(); openPalette(); } });
  search.innerHTML = plIcon('search');
  search.appendChild(el('span', { class: 'pl-side-search-ph', text: pst('top.search') }));
  search.appendChild(el('span', { class: 'pl-keys', dir: 'ltr' }, el('kbd', { text: 'Ctrl' }), el('kbd', { text: 'K' })));
  p.appendChild(search);
  var scroll = el('div', { class: 'pl-side-scroll' });
  scroll.appendChild(el('div', { class: 'pl-side-label', text: mode.label }));
  sideLinks(scroll, mode);
  if (mqMobile.matches) {
    modesOf(model).concat(model.settings ? [model.settings] : []).forEach(function (g) {
      if (g === mode) return;
      scroll.appendChild(el('div', { class: 'pl-side-label', text: g.label }));
      sideLinks(scroll, g);
    });
  }
  p.appendChild(scroll);
  if (mqMobile.matches) p.appendChild(sideTools());

}

function sideTools() {
  var box = el('div', { class: 'pl-side-tools', role: 'group', 'aria-label': pst('rail.tools') });
  var api = nc();
  var st = api && api.handoff ? api.handoff.state() : null;
  if (st && st.visible) {
    var h = el('button', { type: 'button', class: 'pl-nav', 'data-pl-tool': 'handoff', html: plIcon('download'), onclick: function () { closeSheet(); if (api.handoff) api.handoff.open(); } });
    h.appendChild(el('span', { class: 'pl-nav-label', text: pst('top.handoff') }));
    box.appendChild(h);
  }
  var sdk = el('a', { class: 'pl-nav', href: '#/system/sdk', 'data-pl-tool': 'sdk', 'aria-current': currentHash().indexOf('#/system/sdk') === 0 ? 'page' : null, html: plIcon('code'), onclick: function () { closeSheet(); } });
  sdk.appendChild(el('span', { class: 'pl-nav-label', text: pst('top.sdk') }));
  box.appendChild(sdk);
  var dark = root.classList.contains('dark');
  var theme = el('button', { type: 'button', class: 'pl-ctl pl-ctl-icon pl-theme', 'data-pl-tool': 'theme', 'aria-label': pst(dark ? 'top.toLight' : 'top.toDark'), title: pst(dark ? 'top.toLight' : 'top.toDark'), 'aria-pressed': dark ? 'true' : 'false', html: plIcon(dark ? 'sun' : 'moon'), onclick: themeClick });
  var seg = buildLangSeg();
  box.appendChild(el('div', { class: 'pl-side-tools-row' }, seg, theme));
  var calSeg = buildCalSeg();
  if (calSeg) box.appendChild(el('div', { class: 'pl-side-tools-row pl-side-cal' }, calSeg));
  requestAnimationFrame(function () { segIndicator(seg); if (calSeg) segIndicator(calSeg); });
  return box;
}

function railKey(model, mode) {
  var b = brandInfo();
  var ho = nc() && nc().handoff ? nc().handoff.state() : null;
  return [plLang(), currentHash(), mqMobile.matches, compactOn(), !!(ho && ho.visible), root.classList.contains('dark'), mode && mode.id, JSON.stringify(b), JSON.stringify([model.direct, model.groups, model.settings, model.help].map(function (x) {
    return x ? JSON.stringify(x.links || x, function (k, v) { return k === 'svg' ? undefined : v; }) : '';
  }))].join('|');
}

function ensureTip() {
  if (rail.tip) return rail.tip;
  rail.tip = el('div', { class: 'pl-tip', role: 'tooltip', 'aria-hidden': 'true', 'data-l10n-skip': '' });
  document.body.appendChild(rail.tip);
  return rail.tip;
}

function showTip(target) {
  var text = target.getAttribute('data-pl-tip');
  if (!text) return;
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

function renderRail(force) {
  if (root.classList.contains('pl-auth')) return;
  var model = railModel();
  if (!model) { root.classList.remove('pl-railed'); return; }
  if (rail.view && rail.viewHash !== currentHash()) rail.view = null;
  var mode = rail.view === 'help' && model.help ? model.help : rail.view === 'settings' && model.settings ? model.settings : activeMode(model);
  var key = railKey(model, mode);
  if (!force && key === rail.key && rail.el && rail.el.isConnected && rail.side && rail.side.isConnected) return;
  rail.key = key;
  rail.model = model;
  rail.mode = mode;
  if (!rail.el) {
    rail.el = el('nav', { class: 'pl-rail', 'data-l10n-skip': '' });
    rail.el.addEventListener('mouseover', function (e) { var t = e.target.closest('[data-pl-tip]'); if (t && rail.el.contains(t)) showTip(t); });
    rail.el.addEventListener('mouseout', function (e) { var t = e.target.closest('[data-pl-tip]'); if (t && !t.contains(e.relatedTarget)) hideTip(); });
    rail.el.addEventListener('focusin', function (e) { var t = e.target.closest('[data-pl-tip]'); if (t && e.target.matches(':focus-visible')) showTip(t); });
    rail.el.addEventListener('focusout', hideTip);
    rail.side = el('nav', { class: 'pl-side', 'data-l10n-skip': '' });
  }
  rail.el.setAttribute('aria-label', pst('rail.label'));
  rail.side.setAttribute('aria-label', mode.label);
  var act = document.activeElement;
  var keepHref = act && (rail.el.contains(act) || rail.side.contains(act)) ? act.getAttribute('href') : null;
  buildRail(model, mode);
  buildSide(model, mode);
  if (keepHref) {
    var back = rail.el.querySelector('[href="' + keepHref.replace(/"/g, '') + '"]') || rail.side.querySelector('[href="' + keepHref.replace(/"/g, '') + '"]');
    if (back) back.focus({ preventScroll: true });
  }
  if (rail.el.parentNode !== document.body) document.body.appendChild(rail.el);
  if (rail.side.parentNode !== document.body) document.body.appendChild(rail.side);
  root.classList.add('pl-railed');
  if (typeof paintTopbar === 'function') paintTopbar();
}

var COMPACT_KEY = 'nakama.plane.compact';

function compactOn() {
  return !mqMobile.matches && root.classList.contains('pl-compact');
}

function setCompact(on) {
  root.classList.toggle('pl-compact', !!on);
  try { if (on) localStorage.setItem(COMPACT_KEY, '1'); else localStorage.removeItem(COMPACT_KEY); } catch (e) { on = !!on; }
  hideTip();
  rail.key = '';
  renderRail(true);
  var next = on ? document.querySelector('.pl-tb-start .pl-side-toggle') : rail.side && rail.side.querySelector('.pl-side-toggle');
  if (next) next.focus({ preventScroll: true });
}

try { if (localStorage.getItem(COMPACT_KEY) === '1') root.classList.add('pl-compact'); } catch (e) { root.classList.remove('pl-compact'); }

function closeFly() {}

var sheetPrev = null;

function openSheet() {
  if (root.classList.contains('pl-sheet-open')) return;
  rail.key = '';
  renderRail(true);
  sheetPrev = document.activeElement;
  root.classList.add('pl-sheet-open');
  rail.side.setAttribute('role', 'dialog');
  rail.side.setAttribute('aria-modal', 'true');
  var scrim = document.querySelector('.pl-sheet-scrim');
  if (!scrim) {
    scrim = el('div', { class: 'pl-sheet-scrim', 'aria-hidden': 'true', onclick: closeSheet });
    document.body.appendChild(scrim);
  }
  var first = rail.side.querySelector('[aria-current="page"]') || rail.side.querySelector('a, button');
  if (first) first.focus({ preventScroll: true });
  document.addEventListener('keydown', sheetKeys, true);
}

function closeSheet() {
  if (!root.classList.contains('pl-sheet-open')) return;
  root.classList.remove('pl-sheet-open');
  if (rail.side) { rail.side.removeAttribute('role'); rail.side.removeAttribute('aria-modal'); }
  document.removeEventListener('keydown', sheetKeys, true);
  if (sheetPrev && sheetPrev.isConnected) { try { sheetPrev.focus({ preventScroll: true }); } catch (e) { sheetPrev = null; } }
}

function sheetKeys(e) {
  if (e.key === 'Escape') { e.preventDefault(); closeSheet(); return; }
  if (e.key !== 'Tab') return;
  var list = Array.prototype.slice.call(rail.el.querySelectorAll('a[href], button')).concat(Array.prototype.slice.call(rail.side.querySelectorAll('a[href], button'))).filter(function (x) { return x.offsetParent !== null; });
  if (!list.length) return;
  var i = list.indexOf(document.activeElement);
  if (e.shiftKey && i <= 0) { e.preventDefault(); list[list.length - 1].focus(); }
  else if (!e.shiftKey && i === list.length - 1) { e.preventDefault(); list[0].focus(); }
}
