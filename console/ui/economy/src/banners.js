PH_ICONS.banners = '<path d="M238.76,51.73A8,8,0,0,0,232,48H40a8,8,0,0,0-5.66,13.66L76.69,104,34.34,146.34A8,8,0,0,0,40,160H173.62l-28.84,60.56a8,8,0,1,0,14.44,6.88l80-168A8,8,0,0,0,238.76,51.73ZM181.23,144H59.31l34.35-34.34a8,8,0,0,0,0-11.32L59.31,64h160Z"/>';
PH_ICONS['banner-lists'] = '<path d="M192,48H64A16,16,0,0,0,48,64V192a16,16,0,0,0,16,16H192a16,16,0,0,0,16-16V64A16,16,0,0,0,192,48Zm0,144H64V64H192V192ZM240,56V200a8,8,0,0,1-16,0V56a8,8,0,0,1,16,0ZM32,56V200a8,8,0,0,1-16,0V56a8,8,0,0,1,16,0Z"/>';
PH_ICONS.bannerDuo = '<path d="M232,56l-45.71,96H40l48-48L40,56Z" opacity="0.2"/>' + PH_ICONS.banners;

if (TAG_KEY_TYPES.indexOf('banner') === -1) TAG_KEY_TYPES.splice(TAG_KEY_TYPES.indexOf('general'), 0, 'banner');

var BN_LAYOUTS = ['carousel', 'single', 'grid', 'strip'];
var BN_ACTIONS = ['none', 'screen', 'product', 'achievement', 'url', 'sponsor'];
var BN_PLATFORMS = ['android', 'ios', 'web', 'windows', 'macos', 'linux'];
var BN_STATES = ['live', 'scheduled', 'ended', 'off'];
var BN_STATE_COLOR = { live: 'success', scheduled: 'primary', ended: 'secondary', off: 'secondary' };
var BN_WEEK = [6, 0, 1, 2, 3, 4, 5];
var BN_SET = { zone: null, p: null };

function bannerSettings() {
  if (BN_SET.p) return BN_SET.p;
  BN_SET.p = pubApi('GET', '/banners/settings').then(function (res) { BN_SET.zone = (res && res.timezone) || 'UTC'; return BN_SET; }, function () { BN_SET.p = null; BN_SET.zone = BN_SET.zone || 'UTC'; return BN_SET; });
  return BN_SET.p;
}

function bnZone() { return BN_SET.zone || 'UTC'; }

function zoneParts(ms, zone) {
  var o = {};
  new Intl.DateTimeFormat('en-US', { timeZone: zone, hourCycle: 'h23', year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
    .formatToParts(new Date(ms)).forEach(function (p) { o[p.type] = p.value; });
  return o;
}

function zoneOffset(ms, zone) {
  var o = zoneParts(ms, zone);
  return Date.UTC(+o.year, +o.month - 1, +o.day, +o.hour % 24, +o.minute, +o.second) - Math.floor(ms / 1000) * 1000;
}

function wallToIso(wall, zone) {
  var m = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})$/.exec(wall || '');
  if (!m) return null;
  var guess = Date.UTC(+m[1], +m[2] - 1, +m[3], +m[4], +m[5]);
  var ms = guess - zoneOffset(guess, zone);
  ms = guess - zoneOffset(ms, zone);
  return new Date(ms).toISOString();
}

function isoToWall(iso, zone) {
  if (!iso) return '';
  var d = new Date(iso).getTime();
  if (isNaN(d)) return '';
  var o = zoneParts(d, zone);
  return o.year + '-' + o.month + '-' + o.day + 'T' + o.hour + ':' + o.minute;
}

function zoneFmt(iso, zone) {
  if (!iso) return '';
  if (window.NakamaDate) return window.NakamaDate.format(iso, { zone: zone || window.NakamaDate.zone() });
  try {
    return new Intl.DateTimeFormat(locale(), { timeZone: zone || bnZone(), calendar: calendarId(), year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(new Date(iso));
  } catch (e) {
    return fmtDate(iso, true);
  }
}

function zoneDateInput(value, onchange) {
  if (window.NakamaDate) return window.NakamaDate.input({ mode: 'datetime', value: value, inputClass: INPUT_MD, attrs: { 'data-zone-input': '' }, zoneAttrs: { 'data-zone-shown': '' }, onchange: function (v) { onchange(v); } });
  var zone = bnZone();
  var el = h('input', { type: 'datetime-local', class: INPUT_MD + ' tabular-nums', dir: 'ltr', 'data-ltr': '', 'data-zone-input': '' });
  el.value = isoToWall(value, zone);
  var shown = h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1', 'data-zone-shown': '' });
  function paint() {
    var iso = wallToIso(el.value, zone);
    shown.textContent = iso ? t('bn.zoneAt', { time: zoneFmt(iso, zone), zone: zone }) : t('bn.zoneHint', { zone: zone });
  }
  el.addEventListener('input', function () { paint(); onchange(wallToIso(el.value, zone)); });
  paint();
  return h('div', {}, el, shown);
}

function timeOfDayInput(value, onchange, opts) {
  opts = opts || {};
  var plain = !!opts.plain;
  var cur = value && value !== '24:00' ? value : value === '24:00' && !plain ? '23:59' : '';
  if (window.NakamaDate) {
    return window.NakamaDate.input({ mode: 'time', value: cur, inputClass: INPUT_MD, name: opts.name, zone: opts.zone || bnZone(), zoneLine: opts.zoneLine, passInvalid: plain, onchange: function (v) { onchange(v ? (v === '23:59' && !plain ? '24:00' : v) : null); } });
  }
  var el = h('input', { type: 'time', class: INPUT_MD + ' tabular-nums', dir: 'ltr', 'data-ltr': '', step: 60 });
  el.value = value && value !== '24:00' ? value : value === '24:00' ? '23:59' : '';
  el.addEventListener('input', function () { onchange(el.value ? (el.value === '23:59' ? '24:00' : el.value) : null); });
  return el;
}

function perLangInput(value, kind, onchange) {
  var values = Object.assign({}, value || {});
  var langs = LANGS.active.slice();
  var current = langs.length ? langs[0].code : LANGS.def;
  var tabs = h('div', { class: 'eco-tr-tabs', role: 'tablist' });
  var panels = h('div', { class: 'eco-tr-panels' });
  var parts = {};
  var root;
  function emit() {
    var out = {};
    Object.keys(values).forEach(function (k) { var s = String(values[k] || '').trim(); if (s) out[k] = s; });
    onchange(out);
    if (root) root.dispatchEvent(new Event('input', { bubbles: true }));
  }
  langs.forEach(function (l) {
    var code = l.code;
    var dot = h('span', { class: 'eco-tr-dot', 'aria-hidden': 'true' });
    var tab = h('button', { type: 'button', role: 'tab', class: 'eco-tr-tab', 'data-lang': code, 'aria-selected': 'false', title: langName(code), onclick: function () { select(code); } },
      h('span', { dir: langDir(code), text: l.nativeName || code }), code === LANGS.def ? h('span', { class: 'eco-tr-def', text: t('tr.default') }) : null, dot);
    var ctl = kind === 'image'
      ? imageInput(values[code] || '', function (v) { values[code] = v || ''; paintTab(code); emit(); })
      : textInput(values[code] || '', { ltr: true, label: langName(code), placeholder: 'https://', oninput: function (v) { values[code] = v; paintTab(code); emit(); } });
    var panel = h('div', { role: 'tabpanel', 'data-lang': code, hidden: true }, ctl);
    parts[code] = { tab: tab, dot: dot, panel: panel };
    tabs.appendChild(tab);
    panels.appendChild(panel);
  });
  function paintTab(code) {
    var p = parts[code];
    var empty = !String(values[code] || '').trim();
    p.dot.hidden = !empty;
    p.tab.classList.toggle('eco-tr-empty', empty);
  }
  function select(code) {
    current = code;
    Object.keys(parts).forEach(function (k) {
      parts[k].panel.hidden = k !== code;
      parts[k].tab.setAttribute('aria-selected', k === code ? 'true' : 'false');
      parts[k].tab.classList.toggle('eco-tr-on', k === code);
    });
  }
  Object.keys(parts).forEach(paintTab);
  select(current);
  root = h('div', { class: 'eco-tr', 'data-per-lang': kind }, tabs, panels);
  root.validate = function () {
    var any = Object.keys(values).some(function (k) { return String(values[k] || '').trim(); });
    if (any && !String(values[LANGS.def] || '').trim()) { select(LANGS.def); return t('code.DEFAULT_LANGUAGE_REQUIRED', { lang: langName(LANGS.def) }); }
    for (var k in values) {
      var v = String(values[k] || '').trim();
      if (v && (kind === 'url' ? !/^https?:\/\/[^\s"'<>`]+$/i.test(v) : !/^(https?:\/\/|\/(?!\/))[^\s"'<>`]+$/i.test(v))) { select(k); return t('code.INVALID_URL'); }
    }
    return null;
  };
  return root;
}

function bnPreviewCard(v, code, layout) {
  var title = (v.title || {})[code] || (v.title || {})[LANGS.def] || '';
  var sub = (v.subtitle || {})[code] || (v.subtitle || {})[LANGS.def] || '';
  var src = (v.imageUrl || {})[code] || (v.imageUrl || {})[LANGS.def] || '';
  var act = v.action || {};
  var label = (act.label || {})[code] || (act.label || {})[LANGS.def] || '';
  var card2 = h('div', { class: 'eco-bn-card', 'data-layout': layout || 'carousel', dir: langDir(code), lang: code, style: v.bgColor ? '--bn-bg:' + v.bgColor : null, 'data-bn-preview-card': '' });
  if (src) card2.appendChild(h('img', { class: 'eco-bn-img', src: src, alt: '', loading: 'lazy' }));
  else card2.appendChild(h('div', { class: 'eco-bn-img eco-bn-img-empty', 'aria-hidden': 'true' }, icon('bannerDuo', 'h-8 w-8')));
  var text = h('div', { class: 'eco-bn-text' },
    h('div', { class: 'eco-bn-title', 'data-bn-preview-title': '', text: title || t('bn.preview.noTitle') }),
    sub ? h('div', { class: 'eco-bn-sub', text: sub }) : null,
    act.type && act.type !== 'none' ? h('span', { class: 'eco-bn-cta', 'data-bn-preview-cta': '', text: label || t('bn.cta.' + act.type) }) : null);
  card2.appendChild(text);
  return card2;
}

function bnActionText(a) {
  a = a || {};
  if (!a.type || a.type === 'none') return t('bn.act.none');
  if (a.type === 'url') return t('bn.act.url', { url: i18nText(a.url) || '' });
  if (a.type === 'sponsor') return t('bn.act.sponsor');
  return t('bn.act.' + a.type, { name: (a.type === 'product' ? a.product : a.type === 'achievement' ? a.achievement : a.value) || a.value || '' });
}

function bannerPreviewField() {
  return { name: 'livePreview', type: 'static', readonly: true, render: function (v) {
    var langs = LANGS.active.slice();
    var code = langs.length ? langs[0].code : LANGS.def;
    var stage = h('div', { class: 'eco-bn-stage' });
    var seg = segmented(langs.map(function (l) { return { value: l.code, label: l.nativeName || l.code }; }), code, function (c) { code = c; paint(); }, t('bn.preview.lang'));
    seg.setAttribute('data-bn-langs', '');
    var meta = h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300 mt-2', 'data-bn-preview-action': '' });
    var layout = 'carousel';
    var frame = h('div', { class: 'eco-bn-preview', 'data-bn-preview': '' }, h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' },
      h('span', { class: 'text-sm font-medium', text: t('bn.preview') }), seg), stage, meta);
    function paint() {
      clear(stage).appendChild(bnPreviewCard(v, code, layout));
      meta.textContent = bnActionText(v.action);
    }
    var pending = 0;
    function later() { if (pending) return; pending = requestAnimationFrame(function () { pending = 0; paint(); }); }
    setTimeout(function () {
      var form = frame.closest('form');
      if (!form) return;
      form.addEventListener('input', later);
      form.addEventListener('change', later);
      form.addEventListener('click', later);
    }, 0);
    function loadLayout() {
      if (!v.listId) return;
      var ref = refName('banner-lists', v.listId, pubApi);
      if (ref.item !== undefined) { layout = (ref.item && ref.item.layout) || 'carousel'; paint(); }
      else ref.p.then(function (it) { layout = (it && it.layout) || 'carousel'; paint(); });
    }
    loadLayout();
    paint();
    return frame;
  } };
}

function bnTargetChips(it) {
  var out = [];
  if ((it.audienceTagIds || []).length) out.push(t('bn.tg.tags', { n: fmtNum(it.audienceTagIds.length) }));
  if ((it.excludeTagIds || []).length) out.push(t('bn.tg.exclude', { n: fmtNum(it.excludeTagIds.length) }));
  if (it.minLevel != null && it.maxLevel != null) out.push(t('bn.tg.levelRange', { min: fmtNum(it.minLevel), max: fmtNum(it.maxLevel) }));
  else if (it.minLevel != null) out.push(t('bn.tg.levelMin', { n: fmtNum(it.minLevel) }));
  else if (it.maxLevel != null) out.push(t('bn.tg.levelMax', { n: fmtNum(it.maxLevel) }));
  if ((it.platforms || []).length) out.push(it.platforms.map(function (p) { return t('bn.platform.' + p); }).join(lang() === 'fa' ? '، ' : ', '));
  if (it.minVersion || it.maxVersion) out.push(t('bn.tg.version', { v: (it.minVersion || '…') + ' - ' + (it.maxVersion || '…') }));
  if ((it.countries || []).length) out.push(it.countries.join(lang() === 'fa' ? '، ' : ', '));
  if ((it.languages || []).length) out.push(it.languages.map(langName).join(lang() === 'fa' ? '، ' : ', '));
  if (it.playerAge && it.playerAge !== 'all') out.push(t('bn.age.' + it.playerAge + '.short', { n: fmtNum(it.newDays || 7) }));
  if ((it.days || []).length) out.push(BN_WEEK.filter(function (d) { return it.days.indexOf(d) !== -1; }).map(function (d) { return t('bn.day.' + d + '.short'); }).join(lang() === 'fa' ? '، ' : ', '));
  if (it.dayStart && it.dayEnd) out.push(t('bn.tg.hours', { from: numText(it.dayStart), to: numText(it.dayEnd) }));
  return out;
}

function numText(s) {
  return lang() === 'fa' ? String(s).replace(/[0-9]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'[d]; }) : String(s);
}

function bnStateTag(state) {
  return tag(t('bn.state.' + state), BN_STATE_COLOR[state] || 'secondary');
}

function bnThumb(it) {
  var box = h('div', { class: 'eco-bn-thumb shrink-0 rounded-md overflow-hidden flex items-center justify-center' });
  var src = i18nText(it.imageUrl);
  if (src) box.appendChild(img(src, 'banners'));
  else box.appendChild(icon('bannerDuo', 'h-5 w-5'));
  return box;
}

function ctrText(views, clicks) {
  if (!views) return t('bn.ctr.none');
  return t('bn.ctr', { n: fmtNum(Math.round((clicks / views) * 1000) / 10) });
}

function bannerStatsField() {
  return { name: 'stats', type: 'static', readonly: true, when: function (v) { return !!v.id; }, render: function (v) {
    if (!v.id) return h('span', {});
    var box = h('div', { class: 'eco-bn-stats', 'data-bn-stats': '' }, bar('60%'), bar('40%'));
    pubApi('GET', '/banners/' + encodeURIComponent(v.id) + '/stats' + qs({ days: 14 })).then(function (res) {
      clear(box);
      var max = 1;
      (res.days || []).forEach(function (d) { if (d.views > max) max = d.views; if (d.clicks > max) max = d.clicks; });
      box.appendChild(h('div', { class: 'flex flex-wrap items-center gap-3 text-sm mb-2' },
        h('span', { class: 'tabular-nums', 'data-bn-stat-views': '', text: t('bn.stats.views', { n: fmtNum(res.views || 0) }) }),
        h('span', { class: 'tabular-nums', 'data-bn-stat-clicks': '', text: t('bn.stats.clicks', { n: fmtNum(res.clicks || 0) }) }),
        h('span', { class: 'tabular-nums text-secondary-600 dark:text-secondary-300', text: ctrText(res.views || 0, res.clicks || 0) }),
        h('span', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('bn.stats.all', { v: fmtNum(res.allViews || 0), c: fmtNum(res.allClicks || 0) }) })));
      var bars = h('div', { class: 'eco-bn-bars', 'aria-hidden': 'true' });
      (res.days || []).forEach(function (d) {
        bars.appendChild(h('div', { class: 'eco-bn-day', title: d.day },
          h('span', { class: 'eco-bn-v', style: 'block-size:' + Math.round((d.views / max) * 100) + '%' }),
          h('span', { class: 'eco-bn-c', style: 'block-size:' + Math.round((d.clicks / max) * 100) + '%' })));
      });
      box.appendChild(bars);
      box.appendChild(h('div', { class: 'flex items-center gap-3 text-xs text-secondary-600 dark:text-secondary-300 mt-1' },
        h('span', { class: 'eco-bn-key eco-bn-key-v', text: t('bn.stats.viewsKey') }), h('span', { class: 'eco-bn-key eco-bn-key-c', text: t('bn.stats.clicksKey') }),
        h('span', { text: t('bn.stats.note', { zone: res.timezone || bnZone() }) })));
    }, function (err) { clear(box).appendChild(notice('error', errorText(err))); });
    return box;
  } };
}

function isAction(type) { return function (v) { return (v.action || {}).type === type; }; }

RESOURCES.banners = {
  api: 'publication',
  prepare: bannerSettings,
  filters: [
    { name: 'listId', label: 'filter.bn.list', type: 'ref', resource: 'banner-lists', api: pubApi },
    { name: 'state', label: 'filter.bn.state', options: opts('bn.state.', BN_STATES) },
    { name: 'actionType', label: 'filter.bn.action', options: opts('bn.action.', BN_ACTIONS) },
    { name: 'tagId', label: 'filter.tag', type: 'ref', resource: 'tags', api: baseApi, filter: [{ keyType: 'banner' }, { keyType: 'general' }] },
    { name: 'sponsorId', label: 'filter.bn.sponsor', type: 'ref', resource: 'sponsors', api: pubApi }
  ],
  columns: [
    { label: 'col.bn.banner', render: function (it) {
      return h('div', { class: 'flex items-center gap-3 min-w-0' }, bnThumb(it),
        h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.title) || it.name || it.key }),
          h('div', { class: 'flex flex-wrap items-center gap-1.5 text-secondary-500' }, mono(it.key),
            it.sponsorId ? h('span', { class: 'eco-sp-chip text-xs', 'data-bn-sponsor': it.sponsorKey || '', text: t('bn.sponsoredBy', { name: i18nText(it.sponsorName) || it.sponsorKey || '' }) }) : null)));
    }, skeleton: '55%' },
    { label: 'col.bn.list', render: function (it) {
      return h('div', { class: 'flex flex-col gap-1 min-w-0 text-sm' },
        h('a', { href: hashFor('banner-lists', it.listId), class: 'eco-link truncate', text: i18nText(it.listName) || it.listKey }),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums', text: t('bn.priorityN', { n: fmtNum(it.priority || 0) }) }));
    }, width: 170 },
    { label: 'col.bn.when', render: function (it) {
      var box = h('div', { class: 'flex flex-col items-start gap-1 text-xs' }, bnStateTag(it.state));
      if (it.startsAt) box.appendChild(h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: t('bn.from', { date: zoneFmt(it.startsAt) }) }));
      if (it.endsAt) box.appendChild(h('span', { class: 'text-secondary-600 dark:text-secondary-300', text: t('bn.to', { date: zoneFmt(it.endsAt) }) }));
      return box;
    }, width: 190 },
    { label: 'col.bn.who', render: function (it) {
      var chips = bnTargetChips(it);
      return h('div', { class: 'text-sm text-secondary-700 dark:text-secondary-200 eco-clamp', 'data-bn-target': '', text: chips.length ? chips.join(' · ') : t('bn.tg.all') });
    }, skeleton: '50%' },
    { label: 'col.bn.stats', render: function (it) {
      return h('div', { class: 'flex flex-col gap-0.5 text-sm tabular-nums whitespace-nowrap', 'data-bn-row-stats': '' },
        h('span', { text: t('bn.viewsClicks', { v: fmtNum(it.views7 || 0), c: fmtNum(it.clicks7 || 0) }) }),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: ctrText(it.views7 || 0, it.clicks7 || 0) }));
    }, width: 150 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.bn.key.hint', required: true, lockOnEdit: true, keyRe: /^[a-z][a-z0-9_]{1,63}$/, keyErr: 'err.bn.key' },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    { type: 'row', fields: [
      { name: 'listId', type: 'ref', resource: 'banner-lists', api: pubApi, label: 'f.bn.list', hint: 'f.bn.list.hint', required: true },
      { name: 'priority', type: 'number', label: 'f.bn.priority', hint: 'f.bn.priority.hint', min: -1000, max: 1000, empty: 0 }
    ] },
    { name: 'name', type: 'text', label: 'f.bn.name', hint: 'f.bn.name.hint', optional: true },
    bannerPreviewField(),
    { type: 'heading', label: 'bn.h.content' },
    tr('title', { label: 'f.bn.title', required: true, max: 120 }),
    tr('subtitle', { label: 'f.bn.subtitle', max: 200, optional: true }),
    tr('description', { label: 'f.bn.description', multiline: true, max: 1000, optional: true }),
    { name: 'imageUrl', type: 'custom', label: 'f.bn.image', hint: 'f.bn.image.hint', optional: true, build: function (cur, change) { return perLangInput(cur, 'image', change); } },
    { name: 'bgColor', type: 'text', ltr: true, label: 'f.bn.color', hint: 'f.bn.color.hint', optional: true, check: function (v) { return v && !/^#[0-9a-fA-F]{6}$/.test(v) ? t('code.INVALID_COLOR') : null; } },
    tagsField('banner'),
    { name: 'sponsorId', type: 'ref', resource: 'sponsors', api: pubApi, label: 'f.bn.sponsor', hint: 'f.bn.sponsor.hint', optional: true,
      check: function (v, vals) { return !v && (vals.action || {}).type === 'sponsor' ? t('code.SPONSOR_REQUIRED') : null; } },
    { type: 'heading', label: 'bn.h.action' },
    { name: 'action.type', type: 'select', label: 'f.bn.action', hint: 'f.bn.action.hint', required: true, noPlaceholder: true, options: opts('bn.action.', BN_ACTIONS) },
    { name: 'action.value', type: 'text', ltr: true, label: 'f.bn.screen', hint: 'f.bn.screen.hint', required: true, when: isAction('screen'),
      check: function (v) { return v && !/^[a-z][a-z0-9_-]{0,39}$/.test(v) ? t('code.INVALID_ACTION') : null; } },
    { name: 'action.product', type: 'ref', resource: 'products', valueKey: 'key', label: 'f.bn.product', required: true, when: isAction('product') },
    { name: 'action.achievement', type: 'ref', resource: 'achievements', valueKey: 'key', label: 'f.bn.achievement', required: true, when: isAction('achievement') },
    { name: 'action.url', type: 'custom', label: 'f.bn.url', hint: 'f.bn.url.hint', when: isAction('url'), build: function (cur, change) { return perLangInput(cur, 'url', change); },
      check: function (v) { return !v || !Object.keys(v).length ? t('err.required') : null; } },
    tr('action.label', { label: 'f.bn.cta', hint: 'f.bn.cta.hint', max: 40, optional: true, when: function (v) { return (v.action || {}).type && v.action.type !== 'none'; } }),
    { type: 'heading', label: 'bn.h.when' },
    { name: 'zoneNote', type: 'static', readonly: true, render: function () { return h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300', 'data-bn-zone': '', text: t('bn.zoneNote', { console: window.NakamaDate ? window.NakamaDate.zone() : bnZone(), zone: bnZone() }) }); } },
    { type: 'row', fields: [
      { name: 'startsAt', type: 'custom', label: 'f.ev.startsAt', optional: true, build: function (cur, change) { return zoneDateInput(cur, change); } },
      { name: 'endsAt', type: 'custom', label: 'f.ev.endsAt', optional: true, build: function (cur, change) { return zoneDateInput(cur, change); },
        check: function (v, vals) { return v && vals.startsAt && new Date(v) <= new Date(vals.startsAt) ? t('code.INVALID_WINDOW') : null; } }
    ] },
    { name: 'days', type: 'multi', label: 'f.bn.days', hint: 'f.bn.days.hint', optional: true, empty: [], options: function () { return BN_WEEK.map(function (d) { return { value: d, label: t('bn.day.' + d) }; }); } },
    { type: 'row', fields: [
      { name: 'dayStart', type: 'custom', label: 'f.bn.dayStart', hint: 'f.bn.hours.hint', optional: true, build: function (cur, change) { return timeOfDayInput(cur, change); },
        check: function (v, vals) { return !v && vals.dayEnd ? t('code.INVALID_HOURS') : null; } },
      { name: 'dayEnd', type: 'custom', label: 'f.bn.dayEnd', optional: true, build: function (cur, change) { return timeOfDayInput(cur, change); },
        check: function (v, vals) { return (!v && vals.dayStart) || (v && vals.dayStart && v <= vals.dayStart) ? t('code.INVALID_HOURS') : null; } }
    ] },
    { type: 'heading', label: 'bn.h.who' },
    { name: 'audienceTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.bn.audience', hint: 'f.bn.audience.hint', optional: true, empty: [] },
    { name: 'excludeTagIds', type: 'refMulti', resource: 'tags', api: baseApi, filter: { keyType: 'user' }, quick: 'user', label: 'f.bn.exclude', hint: 'f.bn.exclude.hint', optional: true, empty: [],
      check: function (v, vals) { return (v || []).some(function (x) { return (vals.audienceTagIds || []).indexOf(x) !== -1; }) ? t('code.TAG_BOTH_WAYS') : null; } },
    { type: 'row', fields: [
      { name: 'levelTableId', type: 'ref', resource: 'level-tables', label: 'f.tk.levelTable', hint: 'f.bn.levelTable.hint', optional: true },
      { name: 'minLevel', type: 'number', label: 'f.tk.minLevel', optional: true, min: 0 },
      { name: 'maxLevel', type: 'number', label: 'f.tk.maxLevel', optional: true, min: 0,
        check: function (v, vals) { return v != null && vals.minLevel != null && Number(v) < Number(vals.minLevel) ? t('code.INVALID_LEVEL_RANGE') : null; } }
    ] },
    { type: 'row', fields: [
      { name: 'playerAge', type: 'select', label: 'f.bn.age', hint: 'f.bn.age.hint', required: true, noPlaceholder: true, options: opts('bn.age.', ['all', 'new', 'returning']) },
      { name: 'newDays', type: 'number', label: 'f.bn.newDays', min: 1, max: 365, required: true, when: function (v) { return v.playerAge && v.playerAge !== 'all'; } }
    ] },
    { name: 'platforms', type: 'multi', label: 'f.bn.platforms', hint: 'f.bn.platforms.hint', optional: true, empty: [], options: opts('bn.platform.', BN_PLATFORMS) },
    { type: 'row', fields: [
      { name: 'minVersion', type: 'text', ltr: true, label: 'f.bn.minVersion', hint: 'f.bn.version.hint', optional: true },
      { name: 'maxVersion', type: 'text', ltr: true, label: 'f.bn.maxVersion', optional: true }
    ] },
    { type: 'row', fields: [
      { name: 'countriesText', type: 'text', ltr: true, label: 'f.bn.countries', hint: 'f.bn.countries.hint', optional: true,
        check: function (v) { return v && toLatinDigits(v).split(/[\s,،]+/).filter(Boolean).some(function (c) { return !/^[a-zA-Z]{2}$/.test(c); }) ? t('code.INVALID_COUNTRY_CODE') : null; } }
    ] },
    { name: 'languages', type: 'multi', label: 'f.bn.languages', hint: 'f.bn.languages.hint', optional: true, empty: [], options: function () { return LANGS.active.map(function (l) { return { value: l.code, label: langName(l.code) }; }); } },
    { type: 'heading', label: 'bn.h.stats', when: function (v) { return !!v.id; } },
    bannerStatsField()
  ],
  defaults: function (query) {
    return { key: '', isActive: true, listId: (query && query.listId) || null, sponsorId: (query && query.sponsorId) || null, priority: 0, name: '', title: {}, subtitle: {}, description: {}, imageUrl: {}, bgColor: '', tagIds: [],
      action: { type: 'none', value: '', label: {}, url: {}, product: null, achievement: null }, startsAt: null, endsAt: null, days: [], dayStart: null, dayEnd: null,
      audienceTagIds: [], excludeTagIds: [], levelTableId: null, minLevel: null, maxLevel: null, playerAge: 'all', newDays: 7, platforms: [], minVersion: '', maxVersion: '', countriesText: '', languages: [] };
  },
  toForm: function (v) {
    v.action = Object.assign({ type: 'none', value: '', label: {}, url: {} }, v.action || {});
    v.action.product = v.action.type === 'product' ? v.action.value : null;
    v.action.achievement = v.action.type === 'achievement' ? v.action.value : null;
    if (v.action.type !== 'screen') v.action.value = '';
    v.countriesText = (v.countries || []).join(', ');
    v.bgColor = v.bgColor || '';
    return v;
  },
  fixPayload: function (p) {
    var a = p.action || {};
    var out = { type: a.type || 'none', value: '', label: a.label || {}, url: {} };
    if (out.type === 'screen') out.value = a.value || '';
    if (out.type === 'product') out.value = a.product || '';
    if (out.type === 'achievement') out.value = a.achievement || '';
    if (out.type === 'url') out.url = a.url || {};
    if (out.type === 'none') out.label = {};
    p.action = out;
    p.countries = toLatinDigits(p.countriesText || '').split(/[\s,،]+/).filter(Boolean).map(function (c) { return c.toUpperCase(); });
    delete p.countriesText;
    p.imageUrl = p.imageUrl || {};
    p.bgColor = (p.bgColor || '').toLowerCase();
    p.name = p.name || '';
    p.priority = p.priority == null ? 0 : p.priority;
    p.minVersion = toLatinDigits(p.minVersion || '');
    p.maxVersion = toLatinDigits(p.maxVersion || '');
    p.levelTableId = p.levelTableId || '';
    p.sponsorId = p.sponsorId || '';
    p.startsAt = p.startsAt || '';
    p.endsAt = p.endsAt || '';
    p.dayStart = p.dayStart || '';
    p.dayEnd = p.dayEnd || '';
    if (p.playerAge === 'all' || !p.newDays) p.newDays = 7;
    delete p.livePreview;
    delete p.zoneNote;
    delete p.stats;
    return p;
  },
  afterSave: function () { forgetRefs('banners'); forgetRefs('banner-lists'); forgetRefs('sponsors'); }
};

function bnPreviewTool(v) {
  var who = textInput('', { ltr: true, label: t('bn.try.player'), placeholder: t('bn.try.player') });
  who.setAttribute('data-bn-try-player', '');
  var plat = selectInput('', BN_PLATFORMS.map(function (p) { return { value: p, label: t('bn.platform.' + p) }; }), function () {}, { label: t('f.bn.platforms'), placeholder: t('bn.try.anyPlatform') });
  var ver = textInput('', { ltr: true, label: t('f.bn.minVersion'), placeholder: '1.4.0' });
  var out = h('div', { class: 'mt-3 flex flex-col gap-2', 'data-bn-try-out': '' });
  var run = button({ variant: 'primary', size: 'sm', icon: 'search', label: t('bn.try.run'), onclick: function () {
    var q = toLatinDigits(who.value).trim();
    if (!q) { clear(out).appendChild(notice('error', t('bn.try.need'))); return; }
    setBusy(run, true);
    pubApi('GET', '/users/search' + qs({ q: q })).then(function (list) {
      var u = (list || [])[0];
      if (!u) throw new ApiError(404, 'USER_NOT_FOUND', '');
      var pv = plat.value;
      return pubApi('GET', '/banner-lists/' + encodeURIComponent(v.id) + '/preview' + qs({ userId: u.id, platform: pv || '', version: toLatinDigits(ver.value).trim(), lang: '' })).then(function (res) { return { u: u, res: res }; });
    }).then(function (x) {
      clear(out);
      out.appendChild(h('p', { class: 'text-sm', 'data-bn-try-who': '', text: t('bn.try.for', { name: x.u.displayName || x.u.username, n: fmtNum((x.res.items || []).length) }) }));
      (x.res.items || []).forEach(function (b, i) {
        out.appendChild(h('div', { class: 'flex items-center gap-2 text-sm', 'data-bn-try-shown': b.key }, h('span', { class: 'tabular-nums text-secondary-500', text: fmtNum(i + 1) }), h('span', { class: 'font-medium', text: b.title || b.key }), mono(b.key)));
      });
      (x.res.hidden || []).forEach(function (b) {
        out.appendChild(h('div', { class: 'flex flex-wrap items-center gap-2 text-sm text-secondary-600 dark:text-secondary-300', 'data-bn-try-hidden': b.key, 'data-reason': b.reason },
          h('span', { text: b.title || b.key }), tag(t('bn.why.' + b.reason), 'secondary')));
      });
    }, function (err) { clear(out).appendChild(notice('error', errorText(err))); }).then(function () { setBusy(run, false); });
  } });
  run.setAttribute('data-bn-try-run', '');
  return h('div', { class: 'eco-bn-try', 'data-bn-try': '' },
    h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300 mb-2', text: t('bn.try.hint') }),
    h('div', { class: 'flex flex-wrap items-end gap-2' }, h('div', { class: 'min-w-0 flex-1' }, who), h('div', { style: 'inline-size:9rem' }, plat), h('div', { style: 'inline-size:7rem' }, ver), run),
    out);
}

RESOURCES['banner-lists'] = {
  api: 'publication',
  prepare: bannerSettings,
  filters: [
    { name: 'layout', label: 'filter.bn.layout', options: opts('bn.layout.', BN_LAYOUTS) },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) {
      return h('div', { class: 'flex items-center gap-3 min-w-0' }, h('div', { class: 'eco-bn-thumb shrink-0 rounded-md flex items-center justify-center' }, icon('banner-lists', 'h-5 w-5')),
        h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) || it.key }),
          h('div', { class: 'flex flex-wrap items-center gap-1.5 text-secondary-500' }, mono(it.key),
            it.sponsorId ? h('span', { class: 'eco-sp-chip text-xs', 'data-bl-sponsor': '', text: t('bn.sponsoredBy', { name: i18nText(it.sponsorName) || '' }) }) : null)));
    }, skeleton: '45%' },
    { label: 'col.bn.layout', render: function (it) {
      return h('div', { class: 'flex flex-col items-start gap-1' }, tag(t('bn.layout.' + it.layout), 'primary'),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums', text: it.layout === 'carousel' && it.rotateSeconds ? t('bn.rotateN', { n: fmtNum(it.rotateSeconds), max: fmtNum(it.maxItems) }) : t('bn.maxN', { n: fmtNum(it.maxItems) }) }));
    }, width: 170 },
    { label: 'col.bn.banners', render: function (it) {
      return h('a', { href: hashFor('banners', null, { listId: it.id }), class: 'eco-link text-sm tabular-nums whitespace-nowrap', 'data-bn-count': '', text: t('bn.liveOf', { n: fmtNum(it.liveBanners || 0), total: fmtNum(it.banners || 0) }) });
    }, width: 150 },
    { label: 'col.bn.stats', render: function (it) {
      return h('div', { class: 'flex flex-col gap-0.5 text-sm tabular-nums whitespace-nowrap' },
        h('span', { text: t('bn.viewsClicks', { v: fmtNum(it.views7 || 0), c: fmtNum(it.clicks7 || 0) }) }),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: ctrText(it.views7 || 0, it.clicks7 || 0) }));
    }, width: 150 },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.bl.key.hint', required: true, lockOnEdit: true, keyRe: /^[a-z][a-z0-9_]{1,63}$/, keyErr: 'err.bn.key' },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.bl.active.hint' }
    ] },
    tr('name', { label: 'f.bl.name', required: true, max: 100 }),
    tr('description', { multiline: true, max: 500, optional: true }),
    { type: 'row', fields: [
      { name: 'layout', type: 'select', label: 'f.bl.layout', hint: 'f.bl.layout.hint', required: true, noPlaceholder: true, options: opts('bn.layout.', BN_LAYOUTS) },
      { name: 'maxItems', type: 'number', label: 'f.bl.maxItems', hint: 'f.bl.maxItems.hint', required: true, min: 1, max: 20 },
      { name: 'rotateSeconds', type: 'number', label: 'f.bl.rotate', hint: 'f.bl.rotate.hint', required: true, min: 0, max: 60, when: function (v) { return v.layout === 'carousel'; } }
    ] },
    { name: 'sponsorId', type: 'ref', resource: 'sponsors', api: pubApi, label: 'f.bl.sponsor', hint: 'f.bl.sponsor.hint', optional: true },
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' },
    { type: 'heading', label: 'bl.h.order', when: function (v) { return !!v.id; } },
    { name: 'order', type: 'static', readonly: true, when: function (v) { return !!v.id; }, render: function (v) {
      var list = h('div', { class: 'flex flex-col eco-tk-list', 'data-bl-order': '' });
      (v.items || []).forEach(function (b, i) {
        list.appendChild(h('a', { href: hashFor('banners', b.id), class: 'flex items-center gap-2 py-2 text-sm eco-bn-order', 'data-bl-item': b.key },
          h('span', { class: 'tabular-nums text-secondary-500 w-6', text: fmtNum(i + 1) }), bnThumb(b),
          h('span', { class: 'font-medium truncate flex-1', text: i18nText(b.title) || b.key }),
          h('span', { class: 'text-xs tabular-nums text-secondary-600 dark:text-secondary-300 whitespace-nowrap', text: t('bn.priorityN', { n: fmtNum(b.priority || 0) }) }), bnStateTag(b.state)));
      });
      if (!(v.items || []).length) list.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', text: t('bl.none') }));
      return h('div', {}, h('p', { class: 'text-xs text-secondary-600 dark:text-secondary-300 mb-1', text: t('bl.order.hint') }), list,
        h('a', { href: hashFor('banners', 'new', { listId: v.id }), class: 'eco-link text-sm inline-block mt-2', 'data-bl-add': '', text: t('bl.add') }));
    } },
    { type: 'heading', label: 'bl.h.try', when: function (v) { return !!v.id; } },
    { name: 'try', type: 'static', readonly: true, when: function (v) { return !!v.id; }, render: function (v) { return bnPreviewTool(v); } }
  ],
  defaults: function () { return { key: '', isActive: true, name: {}, description: {}, layout: 'carousel', maxItems: 5, rotateSeconds: 5, sort: 0, sponsorId: null }; },
  fixPayload: function (p) {
    if (p.sort == null) p.sort = 0;
    p.sponsorId = p.sponsorId || '';
    if (p.rotateSeconds == null) p.rotateSeconds = 0;
    delete p.order;
    delete p.try;
    return p;
  },
  afterSave: function () { forgetRefs('banner-lists'); }
};
