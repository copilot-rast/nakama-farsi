var CG_PLAYER = { id: null, data: null, loading: false, at: 0, resetClaims: false };

function cgPlayerRoute() {
  var m = /^#\/players\/([0-9a-f-]{36})(?:\?.*)?$/.exec(location.hash || '');
  return m ? m[1] : null;
}

function cgLoadPlayer(id) {
  CG_PLAYER.loading = true;
  return api('GET', '/campaign/players/' + id + qs({ lang: cgLang() })).then(function (d) { return d; }, function (err) { return { error: err }; }).then(function (d) {
    CG_PLAYER.loading = false;
    CG_PLAYER.at = Date.now();
    if (CG_PLAYER.id === id) CG_PLAYER.data = d;
    cgSchedule();
  });
}

function cgReloadPlayer() {
  CG_PLAYER.at = 0;
  if (CG_PLAYER.id) cgLoadPlayer(CG_PLAYER.id);
}

function cgPlayerBody(id, d) {
  if (d.error) return notice('warning', t('pc.noCampaign'));
  var box = h('div', { class: 'flex flex-col gap-3' });
  box.appendChild(h('p', { class: 'text-sm', 'data-campaign-player-xp': '', text: t('pc.xp', { xp: fmtNum(d.xp), level: d.level }) }));
  var claims = check(CG_PLAYER.resetClaims, t('pc.reset.claims'), 'reset-claims');
  claims.input.addEventListener('change', function () { CG_PLAYER.resetClaims = claims.input.checked; });
  (d.chapters || []).forEach(function (ch) {
    var played = (ch.stages || []).filter(function (s) { return s.played; });
    var list = h('ol', { class: 'flex flex-col gap-1' });
    (ch.stages || []).forEach(function (s) {
      var reset = s.played ? button({ label: t('pc.reset'), variant: 'danger', data: { 'campaign-reset': s.key }, onclick: function () {
        confirmBox(t('pc.reset.confirm', { name: s.name }), t('pc.reset')).then(function (ok) {
          if (!ok) return;
          api('DELETE', '/campaign/players/' + id + '/stages/' + s.id + (CG_PLAYER.resetClaims ? '?claims=reset' : '')).then(function () { toast('success', t('saved')); cgReloadPlayer(); },
            function (err) { toast('error', cgErrText(err)); });
        });
      } }) : null;
      list.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-campaign-player-stage': s.key },
        h('span', { class: 'font-medium', text: s.name }), modeTag(s.mode), stars(s.bestStars, 3),
        h('span', { class: 'text-xs text-secondary-500', text: s.played ? t('pc.wins', { w: s.wins, c: s.completions }) : (s.unlocked ? t('pc.notPlayed') : t('pc.locked')) }),
        s.lastCompletedAt ? h('span', { class: 'text-xs text-secondary-500', text: fmtDate(s.lastCompletedAt) }) : null, reset));
    });
    box.appendChild(h('div', { 'data-campaign-player-chapter': ch.key },
      h('div', { class: 'flex items-center gap-2 mb-1' }, h('span', { class: 'text-xs font-semibold text-secondary-600 dark:text-secondary-300', text: ch.name }),
        h('span', { class: 'cg-bar flex-1 max-w-[10rem]' }, h('span', { style: 'width:' + (ch.starPercent || 0) + '%' })),
        h('span', { class: 'text-xs', text: digits(String(ch.stars)) + '/' + digits(String(ch.maxStars)) + ' ★' })),
      played.length || (ch.stages || []).length ? list : null));
  });
  if ((d.attempts || []).length) {
    var ol = h('ol', { class: 'flex flex-col gap-1', 'data-campaign-player-attempts': '' });
    d.attempts.forEach(function (a) {
      ol.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-xs' }, h('span', { class: 'text-secondary-500', text: fmtDate(a.startedAt) }), mono(a.stageKey),
        tag(t('att.' + a.status), a.status === 'done' ? 'success' : a.status === 'open' ? 'primary' : 'secondary'), a.status === 'done' ? stars(a.stars, 3) : null,
        a.rejected ? tag(a.lastReject + ' ×' + digits(String(a.rejected)), 'error') : null));
    });
    box.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pc.attempts') }), ol));
  }
  box.appendChild(claims);
  return box;
}

function cgPlayerCard() {
  var id = cgPlayerRoute();
  var existing = document.querySelector('[data-campaign-player]');
  if (!id || !token()) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (CG_PLAYER.id !== id) { CG_PLAYER.id = id; CG_PLAYER.data = null; CG_PLAYER.at = 0; }
  if (!CG_PLAYER.loading && Date.now() - CG_PLAYER.at > 20000) cgLoadPlayer(id);
  var d = CG_PLAYER.data;
  var key = cgLang() + '|' + id + '|' + (d ? JSON.stringify(d) : 'loading');
  if (existing && existing.getAttribute('data-key') === key && existing.parentNode === grid) return;
  if (existing && existing.contains(document.activeElement) && existing.parentNode === grid) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' }, h('h2', { id: 'campaign-player-title', class: 'font-medium mb-3', text: t('pc.title') }));
  body.appendChild(d ? cgPlayerBody(id, d) : h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  grid.appendChild(h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-campaign-player': '', 'data-key': key, 'data-l10n-skip': '',
    'aria-labelledby': 'campaign-player-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body));
}
