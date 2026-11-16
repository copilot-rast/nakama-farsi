var LB_PLAYER = { id: null, data: null, loading: false, at: 0 };

function lbPlayerRoute() {
  var m = /^#\/players\/([0-9a-f-]{36})(?:\?.*)?$/.exec(location.hash || '');
  return m ? m[1] : null;
}

function lbLoadPlayer(id) {
  LB_PLAYER.loading = true;
  return api('GET', '/leaderboardx/players/' + id).then(function (d) { return d; }, function (err) { return { error: err }; }).then(function (d) {
    LB_PLAYER.loading = false;
    LB_PLAYER.at = Date.now();
    if (LB_PLAYER.id === id) LB_PLAYER.data = d;
    lbSchedule();
  });
}

function lbPlayerBody(d) {
  if (d.error) return notice('warning', lbErrText(d.error));
  var box = h('div', { class: 'flex flex-col gap-3' });
  var cur = d.current || [];
  var hist = d.history || [];
  if (!cur.length && !hist.length) { box.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('pc.none') })); return box; }
  if (cur.length) {
    var ol = h('ol', { class: 'flex flex-col gap-1', 'data-lbx-player-current': '' });
    cur.forEach(function (c) {
      ol.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-lbx-player-board': c.boardKey },
        h('a', { href: '#/leaderboardx/standings?board=' + encodeURIComponent(c.boardKey), class: 'eco-link font-medium', text: textOf(c.name, c.boardKey) }),
        h('span', { 'data-lbx-player-rank': c.boardKey, text: t('pc.rank', { r: c.rank }) }), h('span', { class: 'text-xs text-secondary-500', text: t('pc.score', { s: fmtNum(c.score) }) }),
        c.tier ? tag(c.tier, 'primary') : null, statusTag(c.status)));
    });
    box.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pc.current') }), ol));
  }
  if (hist.length) {
    var hl = h('ol', { class: 'flex flex-col gap-1', 'data-lbx-player-history': '' });
    hist.forEach(function (r) {
      hl.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-xs', 'data-lbx-player-result': r.periodId },
        h('span', { class: 'font-medium', text: textOf(r.boardName, r.boardKey) }), h('span', { class: 'text-secondary-500', text: fmtDate(r.periodEnd) }),
        h('span', { text: t('pc.rank', { r: r.rank }) }), r.tier ? tag(r.tier, 'primary') : null, h('span', { text: rewardsText(r.rewards) }),
        r.status !== 'none' ? resultTag(r.status) : null));
    });
    box.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pc.history') }), hl));
  }
  if (d.refused30d) box.appendChild(h('p', { class: 'text-xs text-error-600', 'data-lbx-player-refused': '', text: t('pc.refused', { n: d.refused30d }) }));
  return box;
}

function lbPlayerCard() {
  var id = lbPlayerRoute();
  var existing = document.querySelector('[data-lbx-player]');
  if (!id || !token()) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (LB_PLAYER.id !== id) { LB_PLAYER.id = id; LB_PLAYER.data = null; LB_PLAYER.at = 0; }
  if (!LB_PLAYER.loading && Date.now() - LB_PLAYER.at > 20000) lbLoadPlayer(id);
  var d = LB_PLAYER.data;
  var key = lbLang() + '|' + id + '|' + (d ? JSON.stringify(d) : 'loading');
  if (existing && existing.getAttribute('data-key') === key && existing.parentNode === grid) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' }, h('h2', { id: 'lbx-player-title', class: 'font-medium mb-3', text: t('pc.title') }));
  body.appendChild(d ? lbPlayerBody(d) : h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  grid.appendChild(h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-lbx-player': '', 'data-key': key, 'data-l10n-skip': '',
    'aria-labelledby': 'lbx-player-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body));
}
