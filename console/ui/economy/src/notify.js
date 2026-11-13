PH_ICONS['auto-rules'] = '<path d="M215.79,118.17a8,8,0,0,0-5-5.66L153.18,90.9l14.66-73.33a8,8,0,0,0-13.69-7l-112,120a8,8,0,0,0,3,13l57.63,21.61L88.16,238.43a8,8,0,0,0,13.69,7l112-120A8,8,0,0,0,215.79,118.17ZM109.37,214l10.47-52.38a8,8,0,0,0-5-9.06L62,132.71l84.62-90.66L136.16,94.43a8,8,0,0,0,5,9.06l52.8,19.8Z"/>';
PH_ICONS['notify-categories'] = '<path d="M224,128a8,8,0,0,1-8,8H128a8,8,0,0,1,0-16h88A8,8,0,0,1,224,128ZM128,72h88a8,8,0,0,0,0-16H128a8,8,0,0,0,0,16Zm88,112H128a8,8,0,0,0,0,16h88a8,8,0,0,0,0-16ZM82.34,42.34,56,68.69,45.66,58.34A8,8,0,0,0,34.34,69.66l16,16a8,8,0,0,0,11.32,0l32-32A8,8,0,0,0,82.34,42.34Zm0,64L56,132.69,45.66,122.34a8,8,0,0,0-11.32,11.32l16,16a8,8,0,0,0,11.32,0l32-32a8,8,0,0,0-11.32-11.32Zm0,64L56,196.69,45.66,186.34a8,8,0,0,0-11.32,11.32l16,16a8,8,0,0,0,11.32,0l32-32a8,8,0,0,0-11.32-11.32Z"/>';
PH_ICONS.push = '<path d="M176,16H80A24,24,0,0,0,56,40V216a24,24,0,0,0,24,24h96a24,24,0,0,0,24-24V40A24,24,0,0,0,176,16ZM72,64H184V192H72Zm8-32h96a8,8,0,0,1,8,8v8H72V40A8,8,0,0,1,80,32Zm96,192H80a8,8,0,0,1-8-8v-8H184v8A8,8,0,0,1,176,224Z"/>';

var AR_TRIGGERS = ['event', 'streak_at_risk', 'chance_ready', 'inactive'];
var AR_CHANNELS = ['push_in_app', 'push', 'in_app'];
var AR_ACTIONS = ['none', 'screen', 'product', 'achievement', 'url'];
var AR_VARS = { event: ['player', 'value'], streak_at_risk: ['player', 'streak', 'count', 'hours'], chance_ready: ['player', 'game'], inactive: ['player', 'days'] };

function isTrigger(x) { return function (v) { return v.trigger === x; }; }

function arSentence(v) {
  var p = v.params || {};
  if (v.trigger === 'event') {
    if (!v.eventKey) return t('ar.s.eventNone');
    return v.matchField && v.matchValue ? t('ar.s.eventMatch', { event: v.eventKey, field: v.matchField, value: v.matchValue }) : t('ar.s.event', { event: v.eventKey });
  }
  if (v.trigger === 'streak_at_risk') return t('ar.s.streak', { streak: p.streakKey || t('ar.anyStreak'), n: fmtNum(p.hoursBefore || 4) });
  if (v.trigger === 'chance_ready') return t('ar.s.chance', { game: p.gameKey || t('ar.anyGame') });
  if (v.trigger === 'inactive') return t('ar.s.inactive', { n: fmtNum(p.days || 3) });
  return '';
}

function arLimitText(v) {
  var out = [];
  var m = v.cooldownMinutes == null ? 1440 : Number(v.cooldownMinutes);
  out.push(m ? (m % 1440 === 0 ? t('ar.cool.days', { n: fmtNum(m / 1440) }) : m % 60 === 0 ? t('ar.cool.hours', { n: fmtNum(m / 60) }) : t('ar.cool.minutes', { n: fmtNum(m) })) : t('ar.cool.none'));
  if (v.quietStart && v.quietEnd) out.push(t('ar.quiet', { from: numText(v.quietStart), to: numText(v.quietEnd) }));
  return out.join(lang() === 'fa' ? '، ' : ', ');
}

function arSentenceField() {
  return { name: 'sentence', type: 'static', readonly: true, render: function (v) {
    var p = h('p', { class: 'eco-ar-sentence rounded-md px-3 py-2 text-sm', 'data-ar-sentence': '' });
    function paint() { p.textContent = arSentence(v) + ' ' + t('ar.s.then', { ch: t('ch.' + (v.channel || 'push_in_app')) }) + ' ' + arLimitText(v) + '.'; }
    var pending = 0;
    function later() { if (pending) return; pending = requestAnimationFrame(function () { pending = 0; paint(); }); }
    setTimeout(function () {
      var form = p.closest('form');
      if (!form) return;
      form.addEventListener('input', later);
      form.addEventListener('change', later);
      form.addEventListener('click', later);
    }, 0);
    paint();
    return p;
  } };
}

function arVarsHint(v) {
  return h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300', 'data-ar-vars': '' }, t('ar.vars') + ' ', (AR_VARS[v.trigger] || AR_VARS.event).map(function (x) { return mono('{' + x + '}'); }));
}

function arStatsField() {
  return { name: 'statsBox', type: 'static', readonly: true, when: function (v) { return !!v.id; }, render: function (v) {
    if (!v.id) return h('span', {});
    var box = h('div', { class: 'flex flex-col gap-3', 'data-ar-stats': '' });
    box.appendChild(h('div', { class: 'eco-sp-tiles' },
      h('div', { class: 'eco-sp-tile rounded-md px-3 py-2', 'data-ar-sent': '' }, h('div', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ar.st.sent') }), h('div', { class: 'text-xl font-semibold tabular-nums', text: fmtNum(v.sent7 || 0) })),
      h('div', { class: 'eco-sp-tile rounded-md px-3 py-2', 'data-ar-skipped': '' }, h('div', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ar.st.skipped') }), h('div', { class: 'text-xl font-semibold tabular-nums', text: fmtNum(v.skipped7 || 0) })),
      h('div', { class: 'eco-sp-tile rounded-md px-3 py-2', 'data-ar-pending': '' }, h('div', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ar.st.pending') }), h('div', { class: 'text-xl font-semibold tabular-nums', text: fmtNum(v.pending || 0) })),
      h('div', { class: 'eco-sp-tile rounded-md px-3 py-2', 'data-ar-opened': '' }, h('div', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ar.st.opened') }), h('div', { class: 'text-xl font-semibold tabular-nums', text: fmtNum(v.opened7 || 0) }))));
    var list = h('div', { class: 'flex flex-col eco-tk-list', 'data-ar-fires': '' }, bar('100%'), bar('80%'));
    box.appendChild(h('h4', { class: 'text-sm font-medium', text: t('ar.recent') }));
    box.appendChild(list);
    pubApi('GET', '/auto-rules/' + encodeURIComponent(v.id) + '/fires' + qs({ limit: 10 })).then(function (res) {
      clear(list);
      (res.items || []).forEach(function (f) {
        list.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 py-2 text-sm', 'data-ar-fire': f.status, 'data-reason': f.reason || '' },
          h('a', { href: '#/players/' + encodeURIComponent(f.userId), class: 'eco-link font-medium', text: f.displayName || f.username || String(f.userId).slice(0, 8) }),
          tag(t('ar.fire.' + f.status), f.status === 'sent' ? 'success' : f.status === 'pending' ? 'primary' : 'secondary'),
          f.reason ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ar.why.' + f.reason) }) : null,
          f.status === 'pending' && f.dueAt ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ar.dueAt', { t: zoneFmt(f.dueAt) }) }) : null,
          f.notificationId ? h('a', { href: hashFor('notifications', f.notificationId), class: 'eco-link text-xs', text: t('ar.openNote') }) : null,
          h('span', { class: 'text-xs text-secondary-500 dark:text-secondary-400 eco-ms-auto', text: fmtDate(f.createdAt, true) })));
      });
      if (!(res.items || []).length) list.appendChild(h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('ar.noFires') }));
    }, function (err) { clear(list).appendChild(notice('error', errorText(err))); });
    var picked = '';
    var who = refPicker({ resource: 'users', api: usersApi, value: '', required: false, label: t('ar.test.player'), placeholder: t('tg.searchUsers'), onchange: function (x) { picked = x || ''; } });
    var out = h('div', { 'aria-live': 'polite', 'data-ar-test-out': '' });
    var run = button({ variant: 'ghost', size: 'sm', icon: 'send', label: t('ar.test.send'), onclick: function () {
      clear(out);
      if (!picked) { out.appendChild(notice('error', t('ar.test.need'))); return; }
      setBusy(run, true);
      pubApi('POST', '/auto-rules/' + encodeURIComponent(v.id) + '/test', { userId: picked }).then(function (n) {
        out.appendChild(h('div', { 'data-ar-test-done': n.id }, notice('success', t('ar.test.done'), h('a', { href: hashFor('notifications', n.id), class: 'eco-link text-sm font-medium', text: t('ar.openNote') }))));
      }, function (err) { out.appendChild(notice('error', errorText(err))); }).then(function () { setBusy(run, false); });
    } });
    run.setAttribute('data-ar-test-run', '');
    box.appendChild(h('h4', { class: 'text-sm font-medium', text: t('ar.test') }));
    box.appendChild(h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ar.test.hint') }));
    box.appendChild(h('div', { class: 'flex flex-wrap items-end gap-2', 'data-ar-test': '' }, h('div', { class: 'min-w-0 flex-1' }, who), run));
    box.appendChild(out);
    return box;
  } };
}

RESOURCES['auto-rules'] = {
  api: 'publication',
  prepare: bannerSettings,
  filters: [
    { name: 'trigger', label: 'filter.ar.trigger', options: opts('ar.trigger.', AR_TRIGGERS) },
    { name: 'categoryId', label: 'filter.ar.category', type: 'ref', resource: 'notify-categories', api: pubApi },
    activeFilter()
  ],
  columns: [
    { label: 'col.ar.rule', render: function (it) {
      return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: it.name || i18nText(it.title) || it.key }), h('div', { class: 'text-secondary-500' }, mono(it.key)));
    }, skeleton: '45%' },
    { label: 'col.ar.when', render: function (it) {
      return h('div', { class: 'text-sm text-secondary-700 dark:text-secondary-200 eco-clamp', 'data-ar-when': '', text: arSentence(it) });
    }, skeleton: '55%' },
    { label: 'col.ar.how', render: function (it) {
      return h('div', { class: 'flex flex-col items-start gap-1 text-xs' }, tag(t('ch.' + it.channel), 'primary'),
        h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: i18nText(it.categoryName) || it.categoryKey }),
        h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: arLimitText(it) }));
    }, width: 190 },
    { label: 'col.bn.stats', render: function (it) {
      return h('div', { class: 'flex flex-col gap-0.5 text-sm tabular-nums whitespace-nowrap', 'data-ar-row-stats': '' },
        h('span', { text: t('ar.sentN', { n: fmtNum(it.sent7 || 0) }) }),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ar.skippedN', { n: fmtNum(it.skipped7 || 0) }) }));
    }, width: 140 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.ar.key.hint', required: true, lockOnEdit: true, keyRe: /^[a-z][a-z0-9_]{1,63}$/, keyErr: 'err.bn.key' },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.ar.active.hint' }
    ] },
    { name: 'name', type: 'text', label: 'f.ar.name', hint: 'f.ar.name.hint', optional: true },
    { type: 'heading', label: 'ar.h.when' },
    arSentenceField(),
    { name: 'trigger', type: 'select', label: 'f.ar.trigger', required: true, noPlaceholder: true, options: opts('ar.trigger.', AR_TRIGGERS) },
    { name: 'eventKey', type: 'ref', resource: 'event-types', valueKey: 'key', label: 'f.ar.event', hint: 'f.ar.event.hint', required: true, when: isTrigger('event') },
    { type: 'row', fields: [
      { name: 'matchField', type: 'text', ltr: true, label: 'f.ar.matchField', hint: 'f.ar.match.hint', optional: true, when: isTrigger('event') },
      { name: 'matchValue', type: 'text', ltr: true, label: 'f.ar.matchValue', optional: true, when: isTrigger('event'),
        check: function (v, vals) { return vals.matchField && !v ? t('err.required') : null; } }
    ] },
    { type: 'row', fields: [
      { name: 'params.streakKey', type: 'ref', resource: 'streaks', valueKey: 'key', label: 'f.ar.streak', hint: 'f.ar.streak.hint', optional: true, when: isTrigger('streak_at_risk') },
      { name: 'params.hoursBefore', type: 'number', label: 'f.ar.hoursBefore', hint: 'f.ar.hoursBefore.hint', min: 1, max: 23, required: true, when: isTrigger('streak_at_risk') }
    ] },
    { name: 'params.gameKey', type: 'ref', resource: 'chance-games', valueKey: 'key', label: 'f.ar.game', hint: 'f.ar.game.hint', optional: true, when: isTrigger('chance_ready') },
    { name: 'params.days', type: 'number', label: 'f.ar.days', hint: 'f.ar.days.hint', min: 1, max: 365, required: true, when: isTrigger('inactive') },
    { type: 'heading', label: 'ar.h.message' },
    { type: 'row', fields: [
      { name: 'channel', type: 'select', label: 'f.ar.channel', required: true, noPlaceholder: true, options: opts('ch.', AR_CHANNELS) },
      { name: 'categoryId', type: 'ref', resource: 'notify-categories', api: pubApi, label: 'f.ar.category', hint: 'f.ar.category.hint', required: true }
    ] },
    tr('title', { label: 'f.ar.title', required: true, max: 120 }),
    tr('body', { label: 'f.ar.body', multiline: true, required: true, max: 500 }),
    { name: 'varsHint', type: 'static', readonly: true, render: function (v) { return arVarsHint(v); } },
    { name: 'imageUrl', type: 'image', label: 'f.ar.image', hint: 'f.ar.image.hint', optional: true },
    { name: 'action.type', type: 'select', label: 'f.bn.action', hint: 'f.ar.action.hint', required: true, noPlaceholder: true, options: opts('bn.action.', AR_ACTIONS) },
    { name: 'action.value', type: 'text', ltr: true, label: 'f.bn.screen', hint: 'f.bn.screen.hint', required: true, when: isAction('screen'),
      check: function (v) { return v && !/^[a-z][a-z0-9_-]{0,39}$/.test(v) ? t('code.INVALID_ACTION') : null; } },
    { name: 'action.product', type: 'ref', resource: 'products', valueKey: 'key', label: 'f.bn.product', required: true, when: isAction('product') },
    { name: 'action.achievement', type: 'ref', resource: 'achievements', valueKey: 'key', label: 'f.bn.achievement', required: true, when: isAction('achievement') },
    { name: 'action.url', type: 'text', ltr: true, label: 'f.ar.url', required: true, when: isAction('url'), check: function (v) { return v && !/^https?:\/\/[^\s"'<>`]+$/i.test(v) ? t('code.INVALID_URL') : null; } },
    tr('action.label', { label: 'f.bn.cta', max: 64, optional: true, when: function (v) { return (v.action || {}).type && v.action.type !== 'none'; } }),
    { type: 'heading', label: 'ar.h.limits' },
    { name: 'cooldownMinutes', type: 'number', label: 'f.ar.cooldown', hint: 'f.ar.cooldown.hint', min: 0, max: 525600, required: true },
    { name: 'quietNote', type: 'static', readonly: true, render: function () { return h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ar.quietNote', { zone: bnZone() }) }); } },
    { type: 'row', fields: [
      { name: 'quietStart', type: 'custom', label: 'f.ar.quietStart', optional: true, build: function (cur, change) { return timeOfDayInput(cur, change); },
        check: function (v, vals) { return !v && vals.quietEnd ? t('code.INVALID_QUIET_HOURS') : null; } },
      { name: 'quietEnd', type: 'custom', label: 'f.ar.quietEnd', optional: true, build: function (cur, change) { return timeOfDayInput(cur, change); },
        check: function (v, vals) { return (!v && vals.quietStart) || (v && v === vals.quietStart) ? t('code.INVALID_QUIET_HOURS') : null; } }
    ] },
    { type: 'heading', label: 'ar.h.stats', when: function (v) { return !!v.id; } },
    arStatsField()
  ],
  defaults: function () {
    return { key: '', isActive: true, name: '', trigger: 'event', eventKey: null, matchField: '', matchValue: '', params: { streakKey: null, hoursBefore: 4, gameKey: null, days: 3 },
      channel: 'push_in_app', categoryId: null, title: {}, body: {}, imageUrl: '', action: { type: 'none', value: '', label: {}, url: '', product: null, achievement: null },
      cooldownMinutes: 1440, quietStart: null, quietEnd: null };
  },
  toForm: function (v) {
    v.params = Object.assign({ streakKey: null, hoursBefore: 4, gameKey: null, days: 3 }, v.params || {});
    var a = v.action || {};
    v.action = { type: a.type || 'none', value: '', label: a.label || {}, url: '', product: null, achievement: null };
    if (a.type === 'screen' || a.type === 'goto') { v.action.type = 'screen'; v.action.value = a.value || ''; }
    if (a.type === 'product') v.action.product = a.value;
    if (a.type === 'achievement') v.action.achievement = a.value;
    if (a.type === 'url') v.action.url = a.value;
    return v;
  },
  fixPayload: function (p) {
    var a = p.action || {};
    var out = null;
    if (a.type && a.type !== 'none') {
      out = { type: a.type, value: a.type === 'screen' ? a.value : a.type === 'product' ? a.product : a.type === 'achievement' ? a.achievement : (a.url || '').trim() };
      if (a.label && Object.keys(a.label).length) out.label = a.label;
    }
    p.action = out;
    var pr = p.params || {};
    p.params = p.trigger === 'streak_at_risk' ? { streakKey: pr.streakKey || '', hoursBefore: pr.hoursBefore || 4 } : p.trigger === 'chance_ready' ? { gameKey: pr.gameKey || '' } : p.trigger === 'inactive' ? { days: pr.days || 3 } : {};
    if (p.trigger !== 'event') { p.eventKey = ''; p.matchField = ''; p.matchValue = ''; }
    p.eventKey = p.eventKey || '';
    p.imageUrl = p.imageUrl || '';
    p.quietStart = p.quietStart || '';
    p.quietEnd = p.quietEnd || '';
    if (p.cooldownMinutes == null) p.cooldownMinutes = 0;
    delete p.sentence;
    delete p.varsHint;
    delete p.quietNote;
    delete p.statsBox;
    return p;
  },
  afterSave: function () { forgetRefs('auto-rules'); }
};

RESOURCES['notify-categories'] = {
  api: 'publication',
  filters: [activeFilter()],
  columns: [
    { label: 'col.name', render: function (it) {
      return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) || it.key }), h('div', { class: 'text-secondary-500' }, mono(it.key)));
    }, skeleton: '45%' },
    { label: 'col.nc.choice', render: function (it) {
      return h('div', { class: 'flex flex-col items-start gap-1 text-xs' },
        tag(it.canOptOut ? t('nc.canOptOut') : t('nc.locked'), it.canOptOut ? 'secondary' : 'warning'),
        h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: it.defaultOn ? t('nc.defaultOn') : t('nc.defaultOff') }));
    }, width: 190 },
    { label: 'col.nc.players', render: function (it) {
      return h('div', { class: 'flex flex-col gap-0.5 text-sm tabular-nums whitespace-nowrap', 'data-nc-counts': '' },
        h('span', { text: t('nc.optedOutN', { n: fmtNum(it.optedOut || 0) }) }),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('nc.usedN', { c: fmtNum(it.campaigns || 0), r: fmtNum(it.rules || 0) }) }));
    }, width: 190 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.nc.key.hint', required: true, lockOnEdit: true, keyRe: /^[a-z][a-z0-9_]{1,63}$/, keyErr: 'err.bn.key' },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.nc.active.hint' }
    ] },
    tr('name', { label: 'f.nc.name', required: true, max: 80 }),
    tr('description', { multiline: true, max: 300, optional: true, hint: 'f.nc.description.hint' }),
    { type: 'row', fields: [
      { name: 'canOptOut', type: 'bool', inline: true, label: 'f.nc.canOptOut', hint: 'f.nc.canOptOut.hint' },
      { name: 'defaultOn', type: 'bool', inline: true, label: 'f.nc.defaultOn', hint: 'f.nc.defaultOn.hint',
        check: function (v, vals) { return !v && !vals.canOptOut ? t('code.CATEGORY_ALWAYS_ON') : null; } }
    ] },
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
  ],
  defaults: function () { return { key: '', isActive: true, name: {}, description: {}, canOptOut: true, defaultOn: true, sort: 0 }; },
  fixPayload: function (p) { if (p.sort == null) p.sort = 0; return p; },
  afterSave: function () { forgetRefs('notify-categories'); }
};
