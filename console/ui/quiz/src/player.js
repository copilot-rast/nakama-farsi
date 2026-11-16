var QZ_PLAYER = { id: null, data: null, loading: false, at: 0, resetRewards: false };

function qzPlayerRoute() {
  var m = /^#\/players\/([0-9a-f-]{36})(?:\?.*)?$/.exec(location.hash || '');
  return m ? m[1] : null;
}

function qzLoadPlayer(id) {
  QZ_PLAYER.loading = true;
  return api('GET', '/quiz/players/' + id + qs({ lang: qzLang() })).then(function (d) { return d; }, function (err) { return { error: err }; }).then(function (d) {
    QZ_PLAYER.loading = false;
    QZ_PLAYER.at = Date.now();
    if (QZ_PLAYER.id === id) QZ_PLAYER.data = d;
    qzSchedule();
  });
}

function qzReloadPlayer() {
  QZ_PLAYER.at = 0;
  if (QZ_PLAYER.id) qzLoadPlayer(QZ_PLAYER.id);
}

function qzPlayerBody(id, d) {
  if (d.error) return notice('warning', t('pc.noQuiz'));
  var box = h('div', { class: 'flex flex-col gap-3' });
  box.appendChild(h('p', { class: 'text-sm', 'data-quiz-player-summary': '', text: t('pc.summary', { packs: (d.packs || []).length, hints: d.hintsBought, reports: d.reports }) }));
  var rewards = check(QZ_PLAYER.resetRewards, t('pc.reset.rewards'), 'reset-rewards');
  rewards.input.addEventListener('change', function () { QZ_PLAYER.resetRewards = rewards.input.checked; });
  if ((d.packs || []).length) {
    var list = h('ol', { class: 'flex flex-col gap-1', 'data-quiz-player-packs': '' });
    d.packs.forEach(function (p) {
      list.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-quiz-player-pack': p.key },
        h('span', { class: 'font-medium', text: p.name }), p.firstPassAt ? tag(t('pc.passed'), 'success') : tag(t('pc.notPassed'), 'secondary'),
        h('span', { class: 'text-xs text-secondary-500', text: t('pc.plays', { p: p.plays, w: p.wins, b: p.bestScore }) }),
        p.lastPlayedAt ? h('span', { class: 'text-xs text-secondary-500', text: fmtDate(p.lastPlayedAt) }) : null,
        button({ label: t('pc.reset'), variant: 'danger', data: { 'quiz-reset': p.key }, onclick: function () {
          confirmBox(t('pc.reset.confirm', { name: p.name }), t('pc.reset')).then(function (ok) {
            if (!ok) return;
            api('DELETE', '/quiz/players/' + id + '/packs/' + p.id + (QZ_PLAYER.resetRewards ? '?rewards=reset' : '')).then(function () { toast('success', t('saved')); qzReloadPlayer(); },
              function (err) { toast('error', qzErrText(err)); });
          });
        } })));
    });
    box.appendChild(list);
  } else box.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('pc.none') }));
  if ((d.missions || []).length) {
    var ml = h('ol', { class: 'flex flex-col gap-1', 'data-quiz-player-missions': '' });
    d.missions.forEach(function (m) {
      ml.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-xs' }, tag(t('ms.kind.' + m.kind)), h('span', { class: 'font-medium', text: m.name }),
        m.completedAt ? tag(t('pc.completed'), 'success') : tag(t('pc.started', { at: fmtDate(m.startedAt) })), m.rewarded ? tag(t('pc.rewarded'), 'primary') : null));
    });
    box.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pc.missions') }), ml));
  }
  if ((d.sessions || []).length) {
    var sl = h('ol', { class: 'flex flex-col gap-1', 'data-quiz-player-sessions': '' });
    d.sessions.forEach(function (s) {
      sl.appendChild(h('li', { class: 'flex flex-wrap items-center gap-2 text-xs' }, h('span', { class: 'text-secondary-500', text: fmtDate(s.startedAt) }), mono(s.packKey), statusTag(s.status),
        h('span', { text: t('pc.score', { c: s.correct, n: s.total, s: s.score }) }), s.status === 'done' ? tag(s.passed ? t('pc.passed') : t('pc.failed'), s.passed ? 'success' : 'error') : null));
    });
    box.appendChild(h('div', {}, h('div', { class: 'text-xs text-secondary-500 mb-1', text: t('pc.sessions') }), sl));
  }
  box.appendChild(rewards);
  return box;
}

function qzPlayerCard() {
  var id = qzPlayerRoute();
  var existing = document.querySelector('[data-quiz-player]');
  if (!id || !token()) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (QZ_PLAYER.id !== id) { QZ_PLAYER.id = id; QZ_PLAYER.data = null; QZ_PLAYER.at = 0; }
  if (!QZ_PLAYER.loading && Date.now() - QZ_PLAYER.at > 20000) qzLoadPlayer(id);
  var d = QZ_PLAYER.data;
  var key = qzLang() + '|' + id + '|' + (d ? JSON.stringify(d) : 'loading');
  if (existing && existing.getAttribute('data-key') === key && existing.parentNode === grid) return;
  if (existing && existing.contains(document.activeElement) && existing.parentNode === grid) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' }, h('h2', { id: 'quiz-player-title', class: 'font-medium mb-3', text: t('pc.title') }));
  body.appendChild(d ? qzPlayerBody(id, d) : h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  grid.appendChild(h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-quiz-player': '', 'data-key': key, 'data-l10n-skip': '',
    'aria-labelledby': 'quiz-player-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body));
}
