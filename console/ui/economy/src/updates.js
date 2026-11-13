var UPDATES = { summary: null, at: 0, busy: false };
var UP_AREA_ICON = { migrations: 'database', core: 'server', plugins: 'puzzle', console: 'dashboard', deps: 'package', build: 'code', e2e: 'checklist', docs: 'book', other: 'files' };
var UP_WARN_ICON = { migrations: 'database', migrationsChanged: 'database', env: 'key', goVersion: 'cpu', deps: 'package', breaking: 'warn' };
var UP_STATUS_COLOR = { ok: 'success', never: 'secondary', error: 'error', rate_limited: 'warning', base_missing: 'error', base_mismatch: 'warning', no_base: 'warning' };
var UP_FILE_COLOR = { added: 'success', removed: 'error', modified: 'secondary', renamed: 'primary', copied: 'primary', changed: 'secondary' };

function updatesBadge() {
  var s = UPDATES.summary;
  return s && s.enabled && s.newCommits > 0 ? s.newCommits : 0;
}

function refreshUpdatesBadge(force) {
  var now = Date.now();
  if (UPDATES.busy || (!force && now - UPDATES.at < 300000)) return;
  if (!token() || !aclCan('SETTINGS', 0)) return;
  UPDATES.at = now;
  UPDATES.busy = true;
  quietBase('GET', '/updates/summary').then(function (res) {
    UPDATES.busy = false;
    if (!res) return;
    var before = updatesBadge();
    UPDATES.summary = res;
    if (before !== updatesBadge()) schedule();
  });
}

function upIcon(name, cls) {
  var P = window.NakamaPlane;
  if (P && P.iconD) {
    var span = h('span', { class: 'eco-up-ic ' + (cls || ''), 'aria-hidden': 'true' });
    span.innerHTML = P.iconD(name, 'pl-ic');
    return span;
  }
  return icon(name === 'gitMerge' ? 'updates' : name === 'gitCommit' ? 'gitCommit' : 'alert', 'h-4 w-4');
}

function shortSha(s) {
  return String(s || '').slice(0, 7);
}

function ghLink(href, text, extra) {
  return h('a', { href: href, target: '_blank', rel: 'noopener noreferrer', class: 'eco-link inline-flex items-center gap-1' + (extra ? ' ' + extra : ''), dir: 'ltr' }, h('span', { text: text }), icon('external', 'h-3 w-3'));
}

function copyText(text) {
  if (navigator.clipboard && window.isSecureContext) return navigator.clipboard.writeText(text);
  return new Promise(function (resolve, reject) {
    var ta = h('textarea', { 'aria-hidden': 'true', style: 'position:fixed;top:-1000px;opacity:0' });
    ta.value = text;
    document.body.appendChild(ta);
    ta.select();
    var ok = false;
    try { ok = document.execCommand('copy'); } catch (e) { ok = false; }
    ta.remove();
    if (ok) resolve(); else reject(new Error('copy failed'));
  });
}

function codeBlock(text) {
  var btn = button({ variant: 'muted', size: 'icon', icon: 'copy', title: t('up.copy'), onclick: function () {
    copyText(text).then(function () { toast('success', t('up.copied')); }, function () { toast('error', t('up.copyFailed')); });
  } });
  btn.setAttribute('data-up-copy', '');
  return h('div', { class: 'eco-up-code', 'data-l10n-skip': '' },
    h('pre', { class: 'font-mono text-xs', dir: 'ltr', text: text }),
    h('div', { class: 'eco-up-copy' }, btn));
}

function relativeHours(iso) {
  if (!iso) return '';
  var ms = new Date(iso).getTime() - Date.now();
  var mins = Math.round(Math.abs(ms) / 60000);
  var key = ms >= 0 ? 'up.in' : 'up.ago';
  if (mins < 1) return t(ms >= 0 ? 'up.soon' : 'up.justNow');
  if (mins < 60) return t(key + '.min', { n: mins });
  var hrs = Math.round(mins / 60);
  if (hrs < 48) return t(key + '.hour', { n: hrs });
  return t(key + '.day', { n: Math.round(hrs / 24) });
}

function releaseBody(text) {
  var box = h('div', { class: 'eco-up-notes text-sm', dir: 'ltr' });
  var list = null;
  String(text || '').split('\n').forEach(function (raw) {
    var line = raw.replace(/\s+$/, '');
    if (!line.trim()) { list = null; return; }
    var m = /^#{1,6}\s+(.*)$/.exec(line);
    if (m) { list = null; box.appendChild(h('h4', { class: 'font-semibold mt-3 mb-1', text: m[1] })); return; }
    var b = /^\s*[-*]\s+(.*)$/.exec(line);
    if (b) {
      if (!list) { list = h('ul', { class: 'eco-up-ul' }); box.appendChild(list); }
      list.appendChild(h('li', { text: b[1].replace(/\*\*/g, '').replace(/`/g, '') }));
      return;
    }
    list = null;
    box.appendChild(h('p', { class: 'mb-1', text: line.replace(/\*\*/g, '').replace(/`/g, '') }));
  });
  return box;
}

function bigNumber(n, labelKey, alert) {
  return h('div', { class: 'eco-up-big' + (alert ? ' eco-up-big--alert' : '') },
    h('span', { class: 'text-3xl font-semibold tabular-nums', text: fmtNum(n || 0) }),
    h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: t(labelKey) }));
}

function upCard(titleKey, iconName, body, actions, extra) {
  var c = card(h('div', { class: 'p-4' },
    h('div', { class: 'flex items-center justify-between gap-2 mb-3' },
      h('div', { class: 'flex items-center gap-2 min-w-0' }, upIcon(iconName), h('h2', { class: 'text-lg font-medium truncate', text: t(titleKey) })),
      actions || null),
    body), extra);
  return c;
}

function updatesGuide(v) {
  var s = v.settings || {};
  var b = v.build || {};
  var p = b.parent || {};
  var remote = p.remote || 'upstream';
  var url = 'https://github.com/' + (s.repo || 'heroiclabs/nakama') + '.git';
  var mine = b.branch || 'dev';
  var steps = [
    ['up.g.remote', 'git remote add ' + remote + ' ' + url + ' || git remote set-url ' + remote + ' ' + url],
    ['up.g.fetch', 'git fetch ' + remote + ' ' + s.branch],
    ['up.g.merge', 'git checkout ' + mine + '\ngit merge ' + remote + '/' + s.branch],
    ['up.g.record', 'build/upstream-sync.sh record ' + remote + ' ' + s.branch + '\ngit add build/upstream.json\ngit commit -m "chore: record the upstream base"'],
    ['up.g.build', 'go build -o nakama -ldflags "$(build/buildinfo.sh)" .\ngo build -buildmode=plugin -ldflags "$(build/buildinfo.sh)" -o modules/base.so ./integrations/base'],
    ['up.g.migrate', './nakama migrate up --database.address "$NAKAMA_DATABASE_ADDRESS"']
  ];
  var list = h('ol', { class: 'eco-up-steps' });
  steps.forEach(function (st, i) {
    list.appendChild(h('li', {},
      h('div', { class: 'flex items-center gap-2 mb-1' }, h('span', { class: 'eco-up-step tabular-nums', text: fmtNum(i + 1) }), h('span', { class: 'font-medium', text: t(st[0]) })),
      codeBlock(st[1])));
  });
  var fork = s.repo === v.forkRepo;
  return upCard('up.guide', 'book', h('div', { 'data-up-guide': '' },
    h('div', { class: 'mb-3' }, notice('info', t('up.g.displayOnly'))),
    list,
    h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mt-3', text: t(fork ? 'up.g.forkNote' : 'up.g.parentNote', { repo: v.forkRepo }) })), null, 'eco-span-all');
}

function statusNotice(v) {
  var st = v.state || {};
  var s = v.settings || {};
  if (!s.enabled) return notice('warning', t('up.n.off'));
  if (st.queued && st.status !== 'rate_limited') return notice('info', t('up.n.queued'));
  if (!v.current && st.status !== 'never') return notice('info', t('up.n.stale', { repo: st.repo + ' ' + st.branch }));
  switch (st.status) {
    case 'never': return notice('info', t('up.n.never'));
    case 'no_base': return notice('warning', t('up.n.noBase'));
    case 'base_mismatch': return notice('warning', t('up.n.mismatch', { repo: s.repo, base: (v.build.parent || {}).repo || '' }));
    case 'base_missing': return notice('error', t('up.n.missing', { repo: s.repo, branch: s.branch }));
    case 'rate_limited': return notice('warning', t('up.n.rate', { when: fmtDate(st.nextAt) }));
    case 'error': return notice('error', t('up.n.error', { msg: st.error || '' }));
  }
  return null;
}

function thisServerCard(v) {
  var b = v.build || {};
  var p = b.parent;
  var r = v.result;
  var behind = r && v.current ? r.aheadBy : null;
  var rows = [
    ['up.commit', b.commit ? h('span', { class: 'inline-flex items-center gap-2 flex-wrap' }, mono(shortSha(b.commit)), b.dirty ? tag(t('up.dirty'), 'warning') : null) : h('span', { class: 'text-secondary-500', text: t('up.unknown') })],
    ['up.branch', b.branch ? mono(b.branch) : h('span', { class: 'text-secondary-500', text: t('up.unknown') })],
    ['up.built', b.date ? h('span', { text: fmtDate(b.date) }) : h('span', { class: 'text-secondary-500', text: t('up.unknown') })],
    ['up.ownAhead', b.commit ? h('span', { class: 'tabular-nums', text: t('up.ownAheadN', { n: b.ownAhead || 0 }) }) : null],
    ['up.base', p ? h('span', { class: 'inline-flex items-center gap-2 flex-wrap' }, ghLink('https://github.com/' + p.repo + '/commit/' + p.commit, shortSha(p.commit), 'font-mono text-xs'), h('span', { class: 'text-xs text-secondary-500', text: fmtDate(p.date) })) : h('span', { class: 'text-secondary-500', text: t('up.noBase') })],
    ['up.baseOf', p ? mono(p.repo + ' · ' + p.branch) : null],
    ['up.recorded', p && p.recordedAt ? h('span', { text: fmtDate(p.recordedAt) }) : null]
  ];
  var top = behind === null || behind === undefined ? null : bigNumber(behind, behind ? 'up.behindBy' : 'up.upToDate', behind > 0);
  var c = upCard('up.this', 'server', h('div', {}, top, dl(rows)));
  c.setAttribute('data-up-this', '');
  return c;
}

function parentCard(v, onCheck) {
  var s = v.settings || {};
  var st = v.state || {};
  var r = v.current ? v.result : null;
  var check = button({ variant: 'primary', size: 'sm', icon: 'refresh', label: t('up.recheck'), disabled: !s.enabled || st.queued, onclick: function () { onCheck(check); } });
  check.setAttribute('data-up-check', '');
  var latest = r && (r.latestRelease ? r.latestRelease.tag : r.latestTag ? r.latestTag.name : '');
  var rows = [
    ['up.repo', ghLink('https://github.com/' + s.repo + '/tree/' + encodeURIComponent(s.branch), s.repo + ' · ' + s.branch)],
    ['up.status', h('span', { class: 'inline-flex flex-wrap gap-1' }, tag(t('up.st.' + (st.status || 'never')), UP_STATUS_COLOR[st.status] || 'secondary'), st.queued ? tag(t('up.st.queued'), 'primary') : null)],
    ['up.latest', latest ? mono(latest) : null],
    ['up.checked', st.checkedAt ? h('span', { title: fmtDate(st.checkedAt) }, h('span', { text: relativeHours(st.checkedAt) })) : h('span', { class: 'text-secondary-500', text: t('up.notYet') })],
    ['up.next', s.enabled && st.nextAt ? h('span', { title: fmtDate(st.nextAt) }, h('span', { text: relativeHours(st.nextAt) })) : null],
    ['up.interval', h('span', { text: t('up.everyN', { n: s.intervalHours || 6 }) })],
    ['up.rate', r && r.rate && r.rate.limit ? h('span', { class: 'tabular-nums', text: t('up.rateN', { n: r.rate.remaining, max: r.rate.limit }) }) : null]
  ];
  var top = r ? bigNumber(r.aheadBy, 'up.newCommits', r.aheadBy > 0) : null;
  var compare = r && r.compareUrl && r.aheadBy ? h('div', { class: 'mt-3' }, ghLink(r.compareUrl, t('up.compare'))) : null;
  var c = upCard('up.parent', 'gitMerge', h('div', {}, top, dl(rows), compare), check);
  c.setAttribute('data-up-parent', '');
  return c;
}

function warningsCard(r) {
  if (!r.warnings || !r.warnings.length) return null;
  var list = h('div', { class: 'flex flex-col gap-3' });
  r.warnings.forEach(function (w) {
    var items = (w.items || []).slice(0, 12);
    list.appendChild(h('div', { class: 'eco-up-warn eco-up-warn--' + (w.kind === 'breaking' || w.kind === 'migrations' ? 'strong' : 'soft'), 'data-up-warning': w.kind },
      upIcon(UP_WARN_ICON[w.kind] || 'warn'),
      h('div', { class: 'min-w-0 flex-1' },
        h('div', { class: 'font-medium', text: t('up.w.' + w.kind, { n: (w.items || []).length }) }),
        h('div', { class: 'text-sm text-secondary-600 dark:text-secondary-300', text: t('up.w.' + w.kind + '.hint') }),
        items.length ? h('div', { class: 'flex flex-wrap gap-1 mt-2' }, items.map(function (it) { return h('span', { class: 'eco-up-chip font-mono text-xs', dir: 'ltr', text: it }); })) : null)));
  });
  var c = upCard('up.warnings', 'warn', list, tag(t('up.warningsN', { n: r.warnings.length }), 'warning'), 'eco-span-all');
  c.setAttribute('data-up-warnings', '');
  return c;
}

function releasesCard(r) {
  var body = h('div', { class: 'flex flex-col gap-4' });
  if (r.tags && r.tags.length) {
    body.appendChild(h('div', { 'data-up-tags': '' },
      h('div', { class: 'text-sm font-medium mb-2', text: t('up.newTags') }),
      h('div', { class: 'flex flex-wrap gap-2' }, r.tags.map(function (tg) { return ghLink(tg.url, tg.name, 'eco-up-chip'); }))));
  }
  (r.releases || []).forEach(function (rel) {
    body.appendChild(h('article', { class: 'eco-up-release', 'data-up-release': rel.tag },
      h('div', { class: 'flex flex-wrap items-center gap-2' },
        ghLink(rel.url, rel.name || rel.tag, 'font-semibold'),
        rel.name && rel.name !== rel.tag ? mono(rel.tag) : null,
        rel.prerelease ? tag(t('up.prerelease'), 'warning') : null,
        rel.breaks ? tag(t('up.breaking'), 'error') : null,
        h('span', { class: 'text-xs text-secondary-500', text: fmtDate(rel.publishedAt, false) })),
      releaseBody(rel.body)));
  });
  if (!(r.releases || []).length && !(r.tags || []).length) {
    body.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400', text: r.latestRelease ? t('up.noNewRelease', { tag: r.latestRelease.tag }) : t('up.noRelease') }));
  }
  var c = upCard('up.releases', 'tag', body, null, 'eco-span-all');
  c.setAttribute('data-up-releases', '');
  return c;
}

function commitsCard(r) {
  var list = h('ul', { class: 'eco-up-commits' });
  (r.commits || []).forEach(function (cm) {
    list.appendChild(h('li', { 'data-up-commit': cm.sha },
      h('div', { class: 'min-w-0 flex-1' },
        h('div', { class: 'eco-up-title', dir: 'auto', text: cm.title }),
        h('div', { class: 'flex flex-wrap items-center gap-2 text-xs text-secondary-500 dark:text-secondary-400 mt-0.5' },
          h('span', { dir: 'ltr', text: cm.author || '' }),
          h('span', { text: fmtDate(cm.date) }),
          cm.merge ? tag(t('up.merge'), 'secondary') : null,
          cm.breaks ? tag(t('up.breaking'), 'error') : null)),
      ghLink(cm.url, shortSha(cm.sha), 'font-mono text-xs shrink-0')));
  });
  var more = r.truncated && r.compareUrl ? h('div', { class: 'mt-3' }, ghLink(r.compareUrl, t('up.more', { n: r.aheadBy }))) : null;
  var c = upCard('up.commits', 'gitCommit', h('div', {}, list, more), tag(t('up.commitsN', { n: r.aheadBy }), 'secondary'), 'eco-span-all');
  c.setAttribute('data-up-commits', '');
  return c;
}

function filesCard(r) {
  var box = h('div', { class: 'flex flex-col gap-2' });
  (r.areas || []).forEach(function (a) {
    var files = h('ul', { class: 'eco-up-files' });
    a.files.forEach(function (f) {
      files.appendChild(h('li', {},
        tag(t('up.fs.' + f.status), UP_FILE_COLOR[f.status] || 'secondary'),
        h('span', { class: 'font-mono text-xs eco-up-path', dir: 'ltr', text: f.previous ? f.previous + ' → ' + f.path : f.path }),
        h('span', { class: 'text-xs tabular-nums text-secondary-500 shrink-0', dir: 'ltr', text: localDigits('+' + f.additions + ' −' + f.deletions) })));
    });
    if (a.files.length < a.count) files.appendChild(h('li', { class: 'text-xs text-secondary-500', text: t('up.moreFiles', { n: a.count - a.files.length }) }));
    var det = h('details', { class: 'eco-up-area', 'data-up-area': a.key },
      h('summary', {},
        upIcon(UP_AREA_ICON[a.key] || 'files'),
        h('span', { class: 'font-medium flex-1 min-w-0', text: t('up.area.' + a.key) }),
        h('span', { class: 'text-xs tabular-nums text-secondary-500', text: t('up.filesN', { n: a.count }) }),
        h('span', { class: 'text-xs tabular-nums text-secondary-500', dir: 'ltr', text: localDigits('+' + a.additions + ' −' + a.deletions) })),
      files);
    if (a.key === 'migrations') det.open = true;
    box.appendChild(det);
  });
  if (!(r.areas || []).length) box.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('up.noFiles') }));
  var c = upCard('up.files', 'files', box, tag(t('up.filesN', { n: r.fileCount || 0 }), 'secondary'), 'eco-span-all');
  c.setAttribute('data-up-files', '');
  return c;
}

function updatesPage(container) {
  var settingsBtn = button({ variant: 'ghost', size: 'page', icon: 'gear', label: t('up.settings'), onclick: function () { psNavigate(hashFor('updates-source'), false); } });
  settingsBtn.setAttribute('data-up-settings', '');
  var frame = pageFrame(t('nav.updates'), t('desc.updates'), aclCan('SETTINGS', 1) ? settingsBtn : null);
  clear(container).appendChild(frame.el);
  var body = h('div', { class: 'flex flex-col gap-4', 'data-updates': '' });
  frame.content.appendChild(body);
  var timer = 0;
  var polls = 0;

  function skeleton() {
    clear(body);
    var grid = h('div', { class: 'eco-detail-grid' });
    [7, 7].forEach(function (n) {
      var rows = [];
      for (var i = 0; i < n; i++) rows.push([h('dt', {}, bar('6rem')), h('dd', {}, bar(['70%', '50%', '40%'][i % 3]))]);
      grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), h('div', { class: 'mb-3' }, bar('4rem', 'h-8')), h('dl', { class: 'eco-dl' }, rows))));
    });
    [4, 6].forEach(function (n) {
      var lines = [];
      for (var i = 0; i < n; i++) lines.push(h('div', { class: 'flex items-center gap-3 py-2' }, bar('60%'), bar('4rem')));
      grid.appendChild(card(h('div', { class: 'p-4' }, h('div', { class: 'mb-4' }, bar('8rem', 'h-4')), lines), 'eco-span-all'));
    });
    body.appendChild(grid);
  }

  function check(btn) {
    setBusy(btn, true);
    baseApi('POST', '/updates/check', {}).then(function (v) {
      toast('success', t('up.queuedToast'));
      polls = 0;
      render(v);
    }, function (err) {
      setBusy(btn, false);
      toast('error', err.code === 'TOO_MANY_CHECKS' ? t('code.TOO_MANY_CHECKS') : errorText(err));
    });
  }

  function render(v) {
    if (!body.isConnected) return;
    UPDATES.summary = { enabled: v.settings.enabled, newCommits: v.current && v.result ? v.result.aheadBy : 0, status: v.state.status };
    UPDATES.at = Date.now();
    clear(body);
    var top = statusNotice(v);
    if (top) { top.setAttribute('data-up-notice', v.state.status); body.appendChild(top); }
    var grid = h('div', { class: 'eco-detail-grid' });
    grid.appendChild(thisServerCard(v));
    grid.appendChild(parentCard(v, check));
    var r = v.current ? v.result : null;
    if (r) {
      var w = warningsCard(r);
      if (w) grid.appendChild(w);
      grid.appendChild(releasesCard(r));
      if (r.aheadBy) {
        grid.appendChild(commitsCard(r));
        grid.appendChild(filesCard(r));
      } else {
        grid.appendChild(card(emptyBlock(t('up.none'), t('up.none.hint')), 'eco-span-all'));
      }
    }
    grid.appendChild(updatesGuide(v));
    body.appendChild(grid);
    clearTimeout(timer);
    if (v.state.queued && v.state.status !== 'rate_limited' && polls < 40) {
      polls++;
      timer = setTimeout(load, 1500);
    }
    schedule();
  }

  function load() {
    if (!body.isConnected) return;
    baseApi('GET', '/updates').then(render, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  skeleton();
  load();
}

var UPDATE_SETTINGS_FIELDS = function (values) {
  var branches = (values.branches && values.branches[values.repo]) || [];
  return [
    { name: 'enabled', type: 'bool', inline: true, label: 'f.up.enabled', hint: 'f.up.enabled.hint' },
    { type: 'row', fields: [
      { name: 'provider', type: 'select', label: 'f.up.provider', hint: 'f.up.provider.hint', required: true, noPlaceholder: true, options: function () { return [{ value: 'github', label: 'GitHub' }]; } },
      { name: 'intervalHours', type: 'select', label: 'f.up.interval', hint: 'f.up.interval.hint', required: true, noPlaceholder: true, options: function () {
        var list = [1, 3, 6, 12, 24, 48, 168];
        if (values.intervalHours && list.indexOf(Number(values.intervalHours)) === -1) list.push(Number(values.intervalHours));
        return list.sort(function (a, b) { return a - b; }).map(function (n) { return { value: String(n), label: t('up.everyN', { n: n }) }; });
      } }
    ] },
    { name: 'repo', type: 'text', ltr: true, label: 'f.up.repo', hint: 'f.up.repo.hint', required: true },
    { name: 'branch', type: 'custom', label: 'f.up.branch', hint: 'f.up.branch.hint', required: true, build: function (cur, change, vals) {
      var listId = 'up-branches';
      var inp = textInput(cur, { ltr: true, mono: true, label: t('f.up.branch'), oninput: function (v) { change(v.trim()); } });
      inp.setAttribute('list', listId);
      var dl2 = h('datalist', { id: listId }, branches.map(function (b) { return h('option', { value: b }); }));
      var chips = h('div', { class: 'flex flex-wrap gap-1 mt-2' }, branches.map(function (b) {
        return h('button', { type: 'button', class: 'eco-up-branch eco-up-chip font-mono text-xs', dir: 'ltr', text: b, onclick: function () { inp.value = b; change(b); } });
      }));
      return h('div', {}, inp, dl2, branches.length ? chips : null);
    } },
    { name: 'token', type: 'custom', label: 'f.up.token', hint: values.canSeal === false ? 'f.up.token.noSeal' : 'f.up.token.hint', optional: true, build: function (cur, change) {
      var inp = textInput(cur || '', { ltr: true, mono: true, label: t('f.up.token'), oninput: function (v) { change(v.trim()); } });
      inp.type = 'password';
      inp.setAttribute('autocomplete', 'new-password');
      inp.classList.add('eco-secret');
      inp.setAttribute('data-up-token', '');
      return inp;
    } },
    { name: 'forkInfo', type: 'static', readonly: true, live: true, render: function (v) {
      return h('div', { class: 'flex flex-col gap-2', 'data-up-fork': '' },
        notice('info', t('up.fork', { repo: v.forkRepo || 'copilot-rast/nakama-farsi' })),
        v.updatedBy ? h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('up.savedBy', { who: v.updatedBy, when: fmtDate(v.updatedAt) }) }) : null);
    } }
  ];
};

function updatesSourcePage(container) {
  var frame = tabFrame(t('desc.updates-source'));
  clear(container).appendChild(frame.el);
  var page = apiPage(baseApi, '/updates/settings', null, UPDATE_SETTINGS_FIELDS, { id: 'updates-source', fixPayload: function (p) {
    return { enabled: !!p.enabled, provider: p.provider || 'github', repo: String(p.repo || '').trim(), branch: String(p.branch || '').trim(), intervalHours: Number(p.intervalHours), token: p.token == null ? '' : String(p.token) };
  }, afterSave: function () { UPDATES.at = 0; refreshUpdatesBadge(true); } });
  frame.content.appendChild(page.el);
  page.load();
}
