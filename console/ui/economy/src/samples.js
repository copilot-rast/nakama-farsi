ICONS['sample-data'] = ICONS.trash;
PH_ICONS['sample-data'] = PH_ICONS.trash;

var SC_PATH = '/sample-data/stargame';

function scKindLabel(kind) {
  return has('sc.k.' + kind) ? t('sc.k.' + kind) : kind;
}

function scReason(reason) {
  return has('sc.reason.' + reason) ? t('sc.reason.' + reason) : reason;
}

function scKeys(keys) {
  return h('div', { class: 'flex flex-wrap gap-1', 'data-l10n-skip': '' }, keys.map(function (k) {
    return h('span', { class: 'eco-sc-key font-mono text-xs', dir: 'ltr', text: k });
  }));
}

function scStat(name, value, label, tone) {
  return h('div', { class: 'eco-sc-stat', 'data-tone': tone, 'data-sc-stat': name },
    h('div', { class: 'eco-sc-num tabular-nums', text: fmtNum(value) }),
    h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: label }));
}

function scTable(p) {
  var rows = (p.tables || []).filter(function (tb) { return tb.deleted.length || tb.kept.length; });
  var columns = [{ label: t('sc.col.part'), width: '22%' }, { label: t('sc.col.delete') }, { label: t('sc.col.keep'), width: '30%' }];
  return table(columns, rows.map(function (tb) {
    var del = h('div', { class: 'flex flex-col gap-1' },
      tb.deleted.length ? h('div', { class: 'text-xs text-secondary-500', text: t('sc.count', { n: tb.deleted.length }) + (tb.related ? ' · ' + t('sc.related', { n: tb.related }) : '') }) : h('span', { class: 'text-xs text-secondary-400', text: '-' }),
      tb.deleted.length ? scKeys(tb.deleted) : null);
    var keep = tb.kept.length ? h('ul', { class: 'flex flex-col gap-1' }, tb.kept.map(function (k) {
      return h('li', { class: 'text-xs', 'data-sc-kept': k.key, 'data-reason': k.reason },
        h('span', { class: 'eco-sc-key font-mono', dir: 'ltr', 'data-l10n-skip': '', text: k.key }),
        h('span', { class: 'text-secondary-500 dark:text-secondary-400', text: ' ' + scReason(k.reason) }));
    })) : h('span', { class: 'text-xs text-secondary-400', text: '-' });
    var cell = h('div', { class: 'flex flex-col gap-1' },
      h('span', { class: 'font-medium text-sm', text: scKindLabel(tb.kind) }),
      h('span', { class: 'font-mono text-xs text-secondary-400', dir: 'ltr', 'data-l10n-skip': '', text: tb.table }));
    return { item: tb, cells: [cell, del, keep] };
  }));
}

function scChanges(p) {
  var items = (p.changes || []).map(function (c) {
    var what = t('sc.change.' + c.kind, { key: c.key });
    var how = c.kind === 'app_version_link' ? t('sc.action.links') : c.action === 'deleted' ? t('sc.action.deleted') : c.action === 'reset' ? t('sc.action.reset') : t('sc.action.kept') + (c.reason ? '، ' + scReason(c.reason) : '');
    return h('li', { class: 'flex flex-wrap items-center gap-2 text-sm', 'data-sc-change': c.kind + ':' + c.key, 'data-action': c.action },
      tag(what, c.action === 'reset' || c.action === 'deleted' ? 'warning' : 'secondary'), h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: how }));
  });
  (p.leaderboards || []).forEach(function (id) {
    items.push(h('li', { class: 'text-sm text-secondary-600 dark:text-secondary-300', 'data-sc-board': id }, h('span', { text: t('sc.leaderboard') + ' ' }), mono(id)));
  });
  if (!items.length) return null;
  return h('div', { class: 'flex flex-col gap-2', 'data-sc-changes': '' },
    h('h3', { class: 'font-semibold text-sm', text: t('sc.changes') }),
    h('ul', { class: 'flex flex-col gap-2' }, items));
}

function scLast(last) {
  if (!last) return h('p', { class: 'text-sm text-secondary-500', 'data-sc-last': '', text: t('sc.never') });
  var by = String(last.by || '').replace(/^console:/, '').replace(/^rpc:/, 'RPC ');
  return h('p', { class: 'text-sm text-secondary-500', 'data-sc-last': last.by || '', text: t('sc.last', { by: by, at: fmtDate(last.at), n: last.deleted || 0 }) });
}

function samplesPage(container) {
  var frame = pageFrame(t('nav.sample-data'), t('desc.sample-data'));
  clear(container).appendChild(frame.el);
  var body = h('div', { class: 'flex flex-col gap-4 eco-sc', 'data-sample-page': 'stargame' });
  frame.content.appendChild(body);

  function skeleton() {
    clear(body).appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' }, bar('40%', 'h-4'), bar('85%'), bar('70%'))));
    body.appendChild(h('div', { class: 'eco-sc-stats' }, [0, 1, 2, 3].map(function () { return h('div', { class: 'eco-sc-stat' }, bar('2.5rem', 'h-6'), bar('70%')); })));
    body.appendChild(table([{ label: t('sc.col.part') }, { label: t('sc.col.delete') }, { label: t('sc.col.keep') }], skeletonRows([{}, {}, {}], 5)));
  }

  function run(data) {
    var p = data.preview;
    var n = p.deleted + p.changed + (p.leaderboards || []).length;
    return confirmModal({ danger: true, icon: 'trash', title: t('sc.confirm.title'), text: t('sc.confirm.text', { n: p.deleted, r: p.related, c: p.changed }), ok: t('sc.confirm.ok') }).then(function (ok) {
      if (!ok || !n) return;
      return baseApi('POST', SC_PATH + '/cleanup', { confirm: 'stargame' }).then(function (res) {
        toast('success', t('sc.done', { n: res && res.result ? res.result.deleted : 0 }));
        load();
      }, function (err) { toast('error', errorText(err)); });
    });
  }

  function paint(data) {
    var p = data.preview;
    var work = p.deleted + p.changed + (p.leaderboards || []).length;
    clear(body);
    var runBtn = button({ variant: data.admin && work ? 'dangerSolid' : 'disabled', size: 'footer', icon: 'trash', label: t('sc.run'), disabled: !data.admin || !work, onclick: function () { run(data); } });
    runBtn.setAttribute('data-sc-run', '');
    body.appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' },
      h('h2', { class: 'font-semibold', text: t('sc.title') }),
      h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 eco-pre-line', text: t('sc.intro') }),
      h('div', { 'data-sc-plugin': data.pluginLoaded ? 'on' : 'off' }, notice(data.pluginLoaded ? 'warning' : 'info', t(data.pluginLoaded ? 'sc.plugin.on' : 'sc.plugin.off'))),
      scLast(data.lastRun))));
    body.appendChild(h('div', { class: 'eco-sc-stats' },
      scStat('deleted', p.deleted, t('sc.stat.deleted'), 'danger'),
      scStat('related', p.related, t('sc.stat.related'), 'neutral'),
      scStat('kept', p.kept, t('sc.stat.kept'), 'neutral'),
      scStat('changed', p.changed, t('sc.stat.changed'), 'warning')));
    if (!work && !p.kept) {
      body.appendChild(card(h('div', { 'data-sc-empty': '' }, emptyBlock(t('sc.empty'), t('sc.empty.hint')))));
    } else {
      body.appendChild(h('div', { class: 'flex flex-col gap-2' },
        h('h3', { class: 'font-semibold text-sm', text: t('sc.preview') }),
        scTable(p)));
      var changes = scChanges(p);
      if (changes) body.appendChild(card(h('div', { class: 'p-4' }, changes)));
    }
    body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-end gap-3' },
      data.admin ? null : h('span', { class: 'text-sm text-secondary-500', 'data-sc-admin-only': '', text: t('sc.adminOnly') }),
      !work && data.admin ? h('span', { class: 'text-sm text-secondary-500', text: t('sc.empty.hint') }) : null,
      runBtn));
  }

  function load() {
    skeleton();
    baseApi('GET', SC_PATH).then(paint, function (err) {
      clear(body).appendChild(errorBlock(err, load));
    });
  }
  load();
}

CUSTOM_PAGES['sample-data'] = function (root) { samplesPage(root); };
