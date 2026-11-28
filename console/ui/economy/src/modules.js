var MODULE_GROUPS = ['compete', 'sales', 'progress', 'publish', 'chat', 'support', 'data'];
var MODULE_LIST = [
  { id: 'cups', group: 'compete', icon: 'medal', routes: ['cups'], api: 'NakamaCup' },
  { id: 'clans', group: 'compete', icon: 'group', routes: ['clans'], api: 'NakamaClan' },
  { id: 'trophies', group: 'compete', icon: 'trophy', routes: ['trophies'], api: 'NakamaClan' },
  { id: 'leaderboardx', group: 'compete', icon: 'ladder', routes: ['leaderboardx'], api: 'NakamaLeaderboardx' },
  { id: 'quiz', group: 'compete', icon: 'help', routes: ['quiz'], api: 'NakamaQuiz' },
  { id: 'campaign', group: 'compete', icon: 'flag', routes: ['campaign'], api: 'NakamaCampaign' },
  { id: 'chance', group: 'compete', icon: 'dice', pages: ['chance-games'] },
  { id: 'offers', group: 'sales', icon: 'tag', pages: ['offers', 'entry-offers', 'offer-preview', 'offer-report', 'offer-settings'] },
  { id: 'tag-discounts', group: 'sales', icon: 'seal', pages: ['tag-discounts', 'tag-discount-preview', 'tag-discount-settings'] },
  { id: 'avatar', group: 'sales', icon: 'user', pages: ['avatar-categories', 'avatar-items', 'avatar-types', 'avatar-settings'] },
  { id: 'friend-gifts', group: 'sales', icon: 'handHeart', pages: ['friend-gifts', 'friend-gift-rules', 'friend-gift-report', 'friend-gift-settings'] },
  { id: 'gifts', group: 'sales', icon: 'ticket', pages: ['gifts'] },
  { id: 'tasks', group: 'progress', icon: 'checklist', pages: ['tasks', 'task-groups'] },
  { id: 'achievements', group: 'progress', icon: 'star', pages: ['achievements'] },
  { id: 'streaks', group: 'progress', icon: 'flame', pages: ['streaks'] },
  { id: 'events', group: 'progress', icon: 'zap', pages: ['event-types', 'event-rules', 'event-log'], ps: ['events'] },
  { id: 'referrals', group: 'progress', icon: 'share', pages: ['referrals'], ps: ['referral-rules'] },
  { id: 'news', group: 'publish', icon: 'newspaper', pages: ['news', 'news-review', 'news-sources', 'news-ingest', 'news-reactions', 'news-settings'] },
  { id: 'comments', group: 'publish', icon: 'message', pages: ['comments'] },
  { id: 'faqs', group: 'publish', icon: 'book', pages: ['faqs'] },
  { id: 'banners', group: 'publish', icon: 'flagBanner', pages: ['banners', 'banner-lists'] },
  { id: 'sponsors', group: 'publish', icon: 'handshake', pages: ['sponsors'] },
  { id: 'auto-rules', group: 'publish', icon: 'send', pages: ['auto-rules'] },
  { id: 'voice', group: 'chat', icon: 'radio', pages: ['voice-rooms'], ps: ['voice'] },
  { id: 'stickers', group: 'chat', icon: 'image', pages: ['sticker-packs', 'sticker-settings', 'comm-settings', 'comm-channels'] },
  { id: 'chat-quota', group: 'chat', icon: 'chats', pages: ['chat-quota', 'chat-quota-settings', 'chat-quota-report'] },
  { id: 'support', group: 'support', icon: 'headset', pages: ['tickets', 'live-chat'], ps: ['support-departments', 'support-canned', 'support-settings'] },
  { id: 'insights', group: 'data', icon: 'chart', pages: ['insights-overview', 'insights-sessions', 'heatmaps', 'insights-settings'], routes: ['insights'], api: 'NakamaInsights' },
  { id: 'content', group: 'data', icon: 'cloudDown', pages: ['content-packs', 'content-settings'] },
  { id: 'itemkinds', group: 'data', icon: 'package', routes: ['itemkinds'], api: 'NakamaItemKinds', needsApi: true },
  { id: 'assets', group: 'data', icon: 'database', routes: ['assets'], api: 'NakamaAssets', needsApi: true }
];
var MODS_KEY = 'nakama-console:modules';
var MODS = { disabled: null, loading: null, token: null, saving: false };

(function () {
  try {
    var v = JSON.parse(localStorage.getItem(MODS_KEY) || 'null');
    if (Array.isArray(v)) MODS.disabled = v.filter(function (x) { return typeof x === 'string'; });
  } catch (e) {
    MODS.disabled = null;
  }
})();

function moduleById(id) {
  for (var i = 0; i < MODULE_LIST.length; i++) if (MODULE_LIST[i].id === id) return MODULE_LIST[i];
  return null;
}

function moduleOn(id) {
  return !MODS.disabled || MODS.disabled.indexOf(id) === -1;
}

function moduleOfPage(page) {
  for (var i = 0; i < MODULE_LIST.length; i++) if ((MODULE_LIST[i].pages || []).indexOf(page) !== -1) return MODULE_LIST[i].id;
  return null;
}

function moduleOfPs(section) {
  for (var i = 0; i < MODULE_LIST.length; i++) if ((MODULE_LIST[i].ps || []).indexOf(section) !== -1) return MODULE_LIST[i].id;
  return null;
}

function moduleOfRoute(hash) {
  var m = /^#\/([^/?]+)/.exec(hash || '');
  if (!m) return null;
  for (var i = 0; i < MODULE_LIST.length; i++) if ((MODULE_LIST[i].routes || []).indexOf(m[1]) !== -1) return MODULE_LIST[i].id;
  return null;
}

function pageModuleOn(page) {
  var id = moduleOfPage(page);
  return !id || moduleOn(id);
}

function modulesKey() {
  return JSON.stringify(MODS.disabled || []);
}

function modulesSet(list) {
  var next = (list || []).slice().sort();
  var changed = JSON.stringify(MODS.disabled || []) !== JSON.stringify(next);
  MODS.disabled = next;
  try { localStorage.setItem(MODS_KEY, JSON.stringify(next)); } catch (e) { MODS.saving = false; }
  if (!changed) return;
  if (rendered) rendered.key = '';
  var cur = parseRoute(location.hash);
  if (ps && ps.ctl && (!cur || cur.page !== 'modules')) ps.ctl.key = '';
  schedule();
  try { window.dispatchEvent(new CustomEvent('nakama-console:modules', { detail: { disabled: next.slice() } })); } catch (e) { MODS.saving = false; }
}

function modulesLoad(force) {
  var tk = token();
  if (!tk) return Promise.resolve(MODS.disabled || []);
  if (MODS.loading && MODS.token === tk && !force) return MODS.loading;
  MODS.token = tk;
  MODS.loading = baseApi('GET', '/console-modules').then(function (res) {
    modulesSet((res && res.disabled) || []);
    return MODS.disabled;
  }, function () {
    MODS.loading = null;
    return MODS.disabled || [];
  });
  return MODS.loading;
}

function modulesSave(list) {
  return baseApi('PUT', '/console-modules', { disabled: list }).then(function (res) {
    modulesSet((res && res.disabled) || []);
    return res;
  });
}

function moduleToggle(id, on) {
  var list = (MODS.disabled || []).filter(function (x) { return x !== id; });
  if (!on) list.push(id);
  return modulesSave(list);
}

function moduleInstalled(m) {
  if (!m.needsApi) return true;
  return !!window[m.api];
}

function moduleBundlePages(m) {
  var api = m.api ? window[m.api] : null;
  if (!api || typeof api.pages !== 'function') return [];
  try {
    return (api.pages() || []).filter(function (p) { return !p.group || p.group === m.id || m.routes.indexOf(p.group) !== -1; });
  } catch (e) {
    return [];
  }
}

function modulePageCount(m) {
  return (m.pages || []).length + (m.ps || []).length + moduleBundlePages(m).length;
}

function moduleWhere(m) {
  if (m.routes && !(m.pages || []).length) return t('mod.where.menu', { name: t('mod.name.' + m.id) });
  var page = m.pages[0];
  if (HUB_OF[page] === 'shop') return t('nav.shop') + ' › ' + t('shop.tab.' + shopTabOf(page).id);
  var sec = sectionOf(page);
  var grp = navGroupOf(sec, page);
  return t(SECTION_LABEL[sec]) + (grp ? ' › ' + t('navg.' + grp) : '');
}

function moduleIcon(name, cls) {
  var P = window.NakamaPlane;
  var box = h('span', { class: 'eco-mod-ic-svg', 'aria-hidden': 'true' });
  if (P && P.icon) box.innerHTML = P.icon(name, cls || 'pl-ic');
  else box.appendChild(icon('page', 'h-4 w-4'));
  return box;
}

function moduleCanEdit() {
  return !!token() && aclCan('SETTINGS', 1);
}

function moduleOffPage(root, id) {
  var name = t('mod.name.' + id);
  var can = moduleCanEdit();
  var on = can ? button({ variant: 'primary', size: 'page', label: t('mod.off.turnOn', { name: name }), onclick: function () {
    setBusy(on, true, t('saving'));
    moduleToggle(id, true).then(function () { toast('success', t('mod.on.done', { name: name })); }, function (err) { setBusy(on, false); toast('error', errorText(err)); });
  } }) : null;
  if (on) on.setAttribute('data-mod-turn-on', id);
  var all = button({ variant: 'ghost', size: 'page', label: t('mod.off.all'), onclick: function () { psNavigate(hashFor('modules'), false); } });
  all.classList.add('eco-mod-all');
  var brand = typeof brandName === 'function' ? brandName() : '';
  var box = h('div', { class: 'eco-mod-off', 'data-mod-off': id },
    h('span', { class: 'eco-mod-off-ic' }, moduleIcon((moduleById(id) || {}).icon || 'toggle')),
    h('h2', { class: 'eco-mod-off-title', text: t('mod.off.title', { name: name }) }),
    h('p', { class: 'eco-mod-off-text', text: brand ? t('mod.off.text', { game: brand }) : t('mod.off.textNoBrand') }),
    h('div', { class: 'eco-mod-off-acts' }, on, can ? all : null),
    h('p', { class: 'eco-mod-off-fine' }, moduleIcon('lock'), h('span', { text: can ? t('mod.off.fine') : t('mod.off.askAdmin') })));
  var page = h('div', { class: 'eco-mod-off-page' }, box);
  clear(root).appendChild(page);
}

function modulesPage(container) {
  var content = h('div', { class: 'flex flex-col flex-1 min-w-0' });
  var frame = { el: content, content: content };
  clear(container).appendChild(frame.el);
  var state = { q: '', filter: 'all' };
  var intro = h('div', { class: 'eco-mod-intro', 'data-mod-intro': '' }, moduleIcon('info'),
    h('div', {}, h('span', { text: t('mod.intro.a') + ' ' }), h('b', { text: t('mod.intro.menu') }), h('span', { text: t('mod.intro.sep') }), h('b', { text: t('mod.intro.search') }), h('span', { text: ' ' + t('mod.intro.and') + ' ' }), h('b', { text: t('mod.intro.cards') }), h('span', { text: ' ' + t('mod.intro.b') })));
  var search = searchInput('', t('mod.search'), debounce(function (e) { state.q = String(e.target.value || '').trim().toLowerCase(); paint(); }, 150));
  search.classList.add('eco-mod-search');
  var segBox = h('div', { class: 'eco-mod-seg' });
  var list = h('div', { class: 'eco-mod-list', 'data-mod-list': '' });
  var core = h('div', { class: 'eco-mod-core', 'data-mod-core': '' }, moduleIcon('lock'), h('div', {}, h('b', { text: t('mod.core.title') + ' ' }), h('span', { text: t('mod.core.text') })));
  frame.content.appendChild(h('div', { class: 'eco-mod', 'data-modules': '' }, intro, h('div', { class: 'eco-mod-bar' }, search, segBox), list, core));
  var can = moduleCanEdit();
  var installed = MODULE_LIST.filter(moduleInstalled);

  function paintSeg() {
    var onN = installed.filter(function (m) { return moduleOn(m.id); }).length;
    var opts = [['all', t('mod.filter.all'), installed.length], ['on', t('mod.filter.on'), onN], ['off', t('mod.filter.off'), installed.length - onN]];
    var seg = h('div', { class: 'eco-seg eco-mod-filter inline-flex items-center gap-0.5 rounded-md p-0.5', role: 'radiogroup', 'aria-label': t('mod.filter.label') });
    opts.forEach(function (o) {
      var pressed = state.filter === o[0];
      seg.appendChild(h('button', { type: 'button', role: 'radio', 'aria-checked': pressed ? 'true' : 'false', class: 'eco-seg-btn inline-flex items-center gap-1.5 rounded font-medium px-3 py-1.5 text-sm' + (pressed ? ' eco-seg-on' : ''), 'data-mod-filter': o[0],
        onclick: function () { state.filter = o[0]; paint(); } }, h('span', { text: o[1] }), h('span', { class: 'eco-mod-cnt tabular-nums', text: fmtNum(o[2]) })));
    });
    clear(segBox).appendChild(seg);
  }

  function row(m) {
    var on = moduleOn(m.id);
    var name = t('mod.name.' + m.id);
    var sw = h('button', { type: 'button', role: 'switch', class: 'eco-switch', 'aria-checked': on ? 'true' : 'false', 'aria-label': name, disabled: can ? null : '', 'data-mod-switch': m.id,
      onclick: function () { flip(m, !on, sw); } });
    var meta = on ? h('div', { class: 'eco-mod-meta' }, moduleIcon('list'), h('span', { text: moduleWhere(m) }), h('i', { text: '•', 'aria-hidden': 'true' }), h('span', { text: t('mod.pages', { n: localDigits(fmtNum(modulePageCount(m))) }) }),
      (m.ps || []).length ? h('i', { text: '•', 'aria-hidden': 'true' }) : null, (m.ps || []).length ? h('span', { text: t('mod.psN', { n: localDigits(fmtNum(m.ps.length)) }) }) : null) : null;
    var note = on ? null : h('div', { class: 'eco-mod-note', 'data-mod-note': m.id }, h('b', { text: t('mod.note.off') + ' ' }), h('span', { text: t('mod.note.text', { where: moduleWhere(m), n: localDigits(fmtNum(modulePageCount(m))) }) }));
    return h('div', { class: 'eco-mod-row' + (on ? '' : ' eco-mod-row--off'), 'data-mod-row': m.id },
      h('span', { class: 'eco-mod-ic' }, moduleIcon(m.icon)),
      h('div', { class: 'eco-mod-txt' },
        h('div', { class: 'eco-mod-name' }, h('span', { text: name }), on ? null : tag(t('mod.offTag'), 'secondary')),
        h('div', { class: 'eco-mod-desc', text: t('mod.desc.' + m.id) }),
        meta, note),
      sw);
  }

  function paint() {
    paintSeg();
    clear(list);
    var shown = 0;
    MODULE_GROUPS.forEach(function (g) {
      var all = installed.filter(function (m) { return m.group === g; });
      if (!all.length) return;
      var items = all.filter(function (m) {
        if (state.filter === 'on' && !moduleOn(m.id)) return false;
        if (state.filter === 'off' && moduleOn(m.id)) return false;
        if (!state.q) return true;
        return (t('mod.name.' + m.id) + ' ' + t('mod.desc.' + m.id) + ' ' + m.id).toLowerCase().indexOf(state.q) !== -1;
      });
      if (!items.length) return;
      shown += items.length;
      var onN = all.filter(function (m) { return moduleOn(m.id); }).length;
      list.appendChild(h('section', { class: 'eco-mod-group', 'data-mod-group': g },
        h('div', { class: 'eco-mod-group-head' }, h('h3', { text: t('mod.group.' + g) }), h('span', { text: t('mod.group.count', { on: localDigits(fmtNum(onN)), all: localDigits(fmtNum(all.length)) }) })),
        h('div', { class: 'eco-mod-card' }, items.map(row))));
    });
    if (!shown) list.appendChild(h('p', { class: 'eco-mod-empty', text: t('empty.search') }));
  }

  function flip(m, on, sw) {
    var before = (MODS.disabled || []).slice();
    sw.setAttribute('aria-busy', 'true');
    moduleToggle(m.id, on).then(function () {
      paint();
      toastAction('success', t(on ? 'mod.on.done' : 'mod.off.done', { name: t('mod.name.' + m.id) }), t('mod.undo'), function () {
        modulesSave(before).then(function () { paint(); }, function (err) { toast('error', errorText(err)); });
      });
    }, function (err) {
      sw.removeAttribute('aria-busy');
      toast('error', errorText(err));
    });
  }

  paint();
  modulesLoad(true).then(function () { if (list.isConnected) paint(); });
}

function toastAction(kind, text, label, fn) {
  if (!toastBox || !toastBox.isConnected) {
    toastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-economy': '' });
    document.body.appendChild(toastBox);
  }
  toastBox.setAttribute('dir', document.documentElement.getAttribute('dir') || 'ltr');
  var el = null;
  var undo = button({ variant: 'link', size: 'sm', label: label, onclick: function () { if (el) el.remove(); fn(); } });
  undo.setAttribute('data-toast-undo', '');
  el = notice(kind, text, undo);
  el.classList.add('shadow-lg');
  toastBox.appendChild(el);
  setTimeout(function () { el.remove(); }, 8000);
}

function modulesApi() {
  return {
    list: function () { return MODULE_LIST.filter(moduleInstalled).map(function (m) { return { id: m.id, on: moduleOn(m.id), label: t('mod.name.' + m.id), routes: (m.routes || []).slice() }; }); },
    on: function (id) { return moduleOn(id); },
    forRoute: function (hash) { var id = moduleOfRoute(hash); return id ? { id: id, label: t('mod.name.' + id), on: moduleOn(id) } : null; },
    routeOn: function (hash) { var id = moduleOfRoute(hash); return !id || moduleOn(id); },
    disabled: function () { return (MODS.disabled || []).slice(); },
    ready: function () { return modulesLoad(false); },
    offPage: function (root, id) { moduleOffPage(root, id); }
  };
}
