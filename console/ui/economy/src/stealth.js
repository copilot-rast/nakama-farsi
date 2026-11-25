PH_ICONS['stealth-bots'] = '<path d="M247.31,124.76c-.35-.79-8.82-19.58-27.65-38.41C194.57,61.26,162.88,48,128,48S61.43,61.26,36.34,86.35C17.51,105.18,9,124,8.69,124.76a8,8,0,0,0,0,6.5c.35.79,8.82,19.57,27.65,38.4C61.43,194.74,93.12,208,128,208s66.57-13.26,91.66-38.34c18.83-18.83,27.3-37.61,27.65-38.4A8,8,0,0,0,247.31,124.76ZM128,192c-30.78,0-57.67-11.19-79.93-33.25A133.47,133.47,0,0,1,25,128,133.33,133.33,0,0,1,48.07,97.25C70.33,75.19,97.22,64,128,64s57.67,11.19,79.93,33.25A133.46,133.46,0,0,1,231.05,128C223.84,141.46,192.43,192,128,192Zm0-112a48,48,0,1,0,48,48A48.05,48.05,0,0,0,128,80Zm0,80a32,32,0,1,1,32-32A32,32,0,0,1,128,160Z"/>';

SECTIONS.economy.push('stealth-bots');
HUBS.bots = { section: 'economy', tabs: [['bots', ''], ['stealth-bots', 'stealth']] };
HUB_OF.bots = 'bots';
HUB_OF['stealth-bots'] = 'bots';

function stealthFields() {
  var range = function (a, b, la, lb, hint, min, max, when) {
    return { type: 'row', when: when, fields: [
      { name: a, type: 'number', label: la, hint: hint, required: true, min: min, max: max },
      { name: b, type: 'number', label: lb, required: true, min: min, max: max,
        check: function (v, vals) { return v != null && vals[a] != null && Number(v) < Number(vals[a]) ? t('sb.err.range') : null; } }
    ] };
  };
  var on = function (v) { return !!v.enabled; };
  return [
    { name: 'enabled', type: 'bool', inline: true, label: 'f.sb.enabled', hint: 'f.sb.enabled.hint' },
    { type: 'heading', label: 'sb.h.who', when: on },
    { type: 'row', when: on, fields: [
      { name: 'poolBots', type: 'bool', inline: true, label: 'f.sb.pool', hint: 'f.sb.pool.hint' },
      { name: 'syntheticFallback', type: 'bool', inline: true, label: 'f.sb.fallback', hint: 'f.sb.fallback.hint' }
    ] },
    { name: 'levelRange', type: 'number', label: 'f.sb.level', hint: 'f.sb.level.hint', required: true, min: 0, max: 100, when: on },
    { type: 'heading', label: 'sb.h.timing', when: on },
    range('arriveMinSeconds', 'arriveMaxSeconds', 'f.sb.arriveMin', 'f.sb.arriveMax', 'f.sb.arrive.hint', 0, 300, on),
    range('thinkMinMs', 'thinkMaxMs', 'f.sb.thinkMin', 'f.sb.thinkMax', 'f.sb.think.hint', 0, 60000, on),
    { name: 'pausePercent', type: 'number', label: 'f.sb.pause', hint: 'f.sb.pause.hint', required: true, min: 0, max: 100, when: on },
    range('pauseMinMs', 'pauseMaxMs', 'f.sb.pauseMin', 'f.sb.pauseMax', null, 0, 60000, function (v) { return v.enabled && Number(v.pausePercent) > 0; }),
    { name: 'randomSide', type: 'bool', inline: true, label: 'f.sb.side', hint: 'f.sb.side.hint', when: on },
    { type: 'heading', label: 'sb.h.online', when: on },
    range('onlineAfterMinMinutes', 'onlineAfterMaxMinutes', 'f.sb.afterMin', 'f.sb.afterMax', 'f.sb.after.hint', 0, 1440, on),
    { name: 'onlineIdlePercent', type: 'number', label: 'f.sb.idle', hint: 'f.sb.idle.hint', required: true, min: 0, max: 100, when: on },
    { name: 'friendsLine', type: 'static', readonly: true, when: on, render: function () {
      return h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400' }, t('sb.friends.line') + ' ', h('a', { href: '#/' + PS_ROUTE + '/friends', class: 'eco-link', text: t('sb.friends.link') }));
    } },
    { type: 'heading', label: 'sb.h.outcome', when: on },
    { type: 'row', when: on, fields: [
      { name: 'trophies', type: 'bool', inline: true, label: 'f.sb.trophies', hint: 'f.sb.trophies.hint' },
      { name: 'fullPrize', type: 'bool', inline: true, label: 'f.sb.prize', hint: 'f.sb.prize.hint' }
    ] },
    { type: 'row', when: on, fields: [
      { name: 'leaderboard', type: 'bool', inline: true, label: 'f.sb.board', hint: 'f.sb.board.hint' },
      { name: 'progress', type: 'bool', inline: true, label: 'f.sb.progress', hint: 'f.sb.progress.hint' }
    ] },
    { type: 'heading', label: 'sb.h.guard', when: on },
    { name: 'guard', type: 'bool', inline: true, label: 'f.sb.guard', hint: 'f.sb.guard.hint', when: on },
    { type: 'row', when: function (v) { return v.enabled && v.guard; }, fields: [
      { name: 'guardDailyWins', type: 'number', label: 'f.sb.guardWins', hint: 'f.sb.guardWins.hint', required: true, min: 0, max: 10000 },
      { name: 'guardBoost', type: 'number', label: 'f.sb.guardBoost', hint: 'f.sb.guardBoost.hint', required: true, min: 0, max: 10 }
    ] }
  ];
}

function stealthSummary(v, pool) {
  var box = h('div', { class: 'mt-2 flex flex-col gap-3', 'data-sb-summary': '' });
  if (!v.enabled) {
    box.appendChild(notice('secondary', t('sb.sum.off')));
    return box;
  }
  var parts = [t('sb.sum.arrive', { a: fmtNum(Number(v.arriveMinSeconds) || 0), b: fmtNum(Number(v.arriveMaxSeconds) || 0) })];
  parts.push(t(v.poolBots ? 'sb.sum.pool' : 'sb.sum.synthetic'));
  var counted = ['trophies', 'fullPrize', 'leaderboard', 'progress'].filter(function (k) { return !v[k]; });
  parts.push(counted.length ? t('sb.sum.notCounted', { list: counted.map(function (k) { return t('sb.out.' + k); }).join('، ') }) : t('sb.sum.allCounted'));
  if (v.guard) parts.push(t('sb.sum.guard', { n: fmtNum(Number(v.guardDailyWins) || 0) }));
  box.appendChild(notice('info', parts.join(t('fr.sum.sep'))));
  if (pool) {
    box.appendChild(h('div', { class: 'eco-stat-grid', 'data-sb-pool': '' },
      [['pool', pool.enabled], ['inMatch', pool.inMatch], ['online', pool.onlineAfterMatch], ['guarded', pool.guardedPlayersToday]].map(function (p) {
        var c = statCard(t('sb.pool.' + p[0]), p[1] || 0, null, null);
        c.setAttribute('data-sb-stat', p[0]);
        return c;
      })));
    if (!pool.enabled && v.poolBots) box.appendChild(notice('warning', t('sb.pool.empty'), h('a', { href: hashFor('bots'), class: 'eco-link', text: t('nav.bots') })));
  }
  box.appendChild(h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('sb.staff') }));
  return box;
}

function stealthBotsPage(container, query) {
  var frame = pageFrame(t('nav.stealth-bots'), t('desc.stealth-bots'));
  clear(container).appendChild(frame.el);
  var head = h('div', { class: 'mb-4', 'data-sb-games': '' });
  var body = h('div', {});
  frame.content.appendChild(head);
  frame.content.appendChild(body);
  api('GET', '/game-stealth').then(function (res) {
    var items = (res && res.items) || [];
    if (!items.length) {
      body.appendChild(notice('info', t('sb.noGames'), h('a', { href: hashFor('games'), class: 'eco-link', text: t('nav.games') })));
      return;
    }
    var key = query && query.game && items.some(function (g) { return g.key === query.game; }) ? query.game : items[0].key;
    head.appendChild(listToolbar({ filters: [filterSelect(t('sb.game'), key, items.map(function (g) {
      return { value: g.key, label: (i18nText(g.name) || g.key) + (g.enabled ? ' · ' + t('sb.on') : '') };
    }), function (v) { replaceHash(hashFor('stealth-bots', null, { game: v })); stealthBotsPage(container, { game: v }); }, { def: key })] }));
    var pool = null;
    var call = function (method, path, payload) {
      return api(method, path, payload).then(function (out) {
        pool = out.pool || null;
        return Object.assign({}, out.settings || {});
      });
    };
    var page = apiPage(call, '/game-stealth/' + encodeURIComponent(key), null, stealthFields(), {
      id: 'stealth-bots',
      onchange: function (v, extra) { clear(extra).appendChild(stealthSummary(v, pool)); },
      fixPayload: function (p) { delete p.friendsLine; return p; }
    });
    body.appendChild(page.el);
    page.load();
  }, function (err) { body.appendChild(errorBlock(err, function () { stealthBotsPage(container, query); })); });
}

CUSTOM_PAGES['stealth-bots'] = function (root, route) { stealthBotsPage(root, (route && route.query) || {}); };
