PH_ICONS['chance-games'] = '<path d="M192,32H64A32,32,0,0,0,32,64V192a32,32,0,0,0,32,32H192a32,32,0,0,0,32-32V64A32,32,0,0,0,192,32Zm16,160a16,16,0,0,1-16,16H64a16,16,0,0,1-16-16V64A16,16,0,0,1,64,48H192a16,16,0,0,1,16,16ZM104,92A12,12,0,1,1,92,80,12,12,0,0,1,104,92Zm72,0a12,12,0,1,1-12-12A12,12,0,0,1,176,92Zm-72,72a12,12,0,1,1-12-12A12,12,0,0,1,104,164Zm36-36a12,12,0,1,1-12-12A12,12,0,0,1,140,128Zm36,36a12,12,0,1,1-12-12A12,12,0,0,1,176,164Z"/>';
PH_ICONS.diceDuo = '<path d="M216,64V192a24,24,0,0,1-24,24H64a24,24,0,0,1-24-24V64A24,24,0,0,1,64,40H192A24,24,0,0,1,216,64Z" opacity="0.2"/>' + PH_ICONS['chance-games'];
PH_ICONS['prize-groups'] = '<path d="M223.68,66.15,135.68,18a15.88,15.88,0,0,0-15.36,0l-88,48.17a16,16,0,0,0-8.32,14v95.64a16,16,0,0,0,8.32,14l88,48.17a15.88,15.88,0,0,0,15.36,0l88-48.17a16,16,0,0,0,8.32-14V80.18A16,16,0,0,0,223.68,66.15ZM128,32l80.34,44-29.77,16.3-80.35-44ZM128,120,47.66,76l33.9-18.56,80.34,44ZM40,90l80,43.78v85.79L40,175.82Zm176,85.78h0l-80,43.79V133.82l32-17.51V152a8,8,0,0,0,16,0V107.55L216,90v85.77Z"/>';
PH_ICONS.packageDuo = '<path d="M128,129.09V232a8,8,0,0,1-3.84-1l-88-48.18a8,8,0,0,1-4.16-7V80.18a8,8,0,0,1,.7-3.25Z" opacity="0.2"/>' + PH_ICONS['prize-groups'];
PH_ICONS.timerDuo = '<path d="M216,136a88,88,0,1,1-88-88A88,88,0,0,1,216,136Z" opacity="0.2"/><path d="M128,40a96,96,0,1,0,96,96A96.11,96.11,0,0,0,128,40Zm0,176a80,80,0,1,1,80-80A80.09,80.09,0,0,1,128,216ZM173.66,90.34a8,8,0,0,1,0,11.32l-40,40a8,8,0,0,1-11.32-11.32l40-40A8,8,0,0,1,173.66,90.34ZM96,16a8,8,0,0,1,8-8h48a8,8,0,0,1,0,16H104A8,8,0,0,1,96,16Z"/>';

var CHANCE_PRIZE_TYPES = ['prize', 'prizeGroup', 'slots'];
var CH_SLOT_KINDS = ['currency', 'prize', 'prizeGroup'];

function chNotSlots(v) { return v.prizeType !== 'slots'; }
function chSlots(v) { return v.prizeType === 'slots'; }

function chLadder(text) {
  return String(text || '').split(/[\s,،]+/).filter(function (x) { return x !== ''; }).map(function (x) { return parseNumber(x, true); });
}
var CH_SLICE_COLORS = 8;

function chanceSum(rows) {
  var total = 0;
  (rows || []).forEach(function (r) { var v = Number(r && r.chance); if (isFinite(v)) total += Math.round(v * 10000); });
  return total / 10000;
}

function chanceTile(cls) {
  return h('div', { class: 'h-9 w-9 shrink-0 rounded-md flex items-center justify-center eco-ch-tile' + (cls ? ' ' + cls : '') }, icon('diceDuo', 'h-5 w-5'));
}

function chanceBar(rows, prizeType) {
  var barEl = h('div', { class: 'eco-ch-bar', 'data-ch-bar': '', role: 'img', 'aria-label': t('ch.bar.label', { n: fmtNum((rows || []).length) }) });
  (rows || []).forEach(function (r, i) {
    var v = Number(r.chance) || 0;
    if (v <= 0) return;
    var empty = prizeType === 'slots' ? false : prizeType === 'prizeGroup' ? !r.prizeGroupId : !r.prizeId;
    barEl.appendChild(h('span', { 'data-slice': String(i % CH_SLICE_COLORS), 'data-empty': empty ? '' : null, style: 'flex:' + v + ' 1 0' }));
  });
  return barEl;
}

function chanceModes(it) {
  var out = [];
  if (it.intervalTime) out.push(tag(t('ch.mode.free', { n: chanceMinutes(it.intervalTime) }), 'success'));
  if (it.price && it.price.currencyId) {
    out.push(h('span', { class: 'eco-tag eco-tag--primary inline-flex items-center gap-1 leading-tight max-w-full px-2 py-0.5 text-xs rounded-md whitespace-nowrap', 'data-ch-price': '' },
      h('span', { class: 'tabular-nums', text: t('ch.mode.ticket', { n: fmtNum(it.price.amount) }) }), refNameCell('currencies', it.price.currencyId)));
  }
  if (it.priceLadder && it.priceLadder.length) out.push(h('span', { 'data-ch-ladder': '' }, tag(t('ch.ladder.tag', { list: it.priceLadder.map(function (n) { return fmtNum(n); }).join(' › ') }), 'primary')));
  if (it.pityThreshold) out.push(h('span', { 'data-ch-pity': '' }, tag(t('ch.pity.tag', { n: fmtNum(it.pityThreshold) }), 'success')));
  if (it.dailyLimit) out.push(h('span', { 'data-ch-daily': '' }, tag(t('ch.limit.daily', { n: fmtNum(it.dailyLimit) }), 'warning')));
  if (it.totalLimit) out.push(h('span', { 'data-ch-total': '' }, tag(t('ch.limit.total', { n: fmtNum(it.totalLimit) }), 'warning')));
  return out;
}

function chanceMinutes(n) {
  n = Number(n) || 0;
  if (n >= 1440 && n % 1440 === 0) return t('ch.days', { n: fmtNum(n / 1440) });
  if (n >= 60 && n % 60 === 0) return t('ch.hours', { n: fmtNum(n / 60) });
  return t('ch.minutes', { n: fmtNum(n) });
}

function chanceRowsSummary(it) {
  var rows = it.prize || [];
  if (it.prizeType === 'slots') return t('ch.slotsN', { n: fmtNum(rows.length) });
  var won = rows.filter(function (r) { return it.prizeType === 'prizeGroup' ? r.prizeGroupId : r.prizeId; }).length;
  return t('ch.rowsN', { n: fmtNum(rows.length), w: fmtNum(won) });
}

function chanceSumBlock(v) {
  var rows = v.prize || [];
  var total = chanceSum(rows);
  var ok = Math.round(total * 10000) === 1000000;
  setTimeout(function () {
    var el = document.querySelector('[data-economy="panel"] [data-rows="prize"]');
    if (el) el.setAttribute('data-type', v.prizeType || 'prize');
  }, 0);
  var left = Math.round((100 - total) * 10000) / 10000;
  return h('div', { class: 'eco-ch-sum rounded-md px-3 py-2 flex flex-col gap-2', 'data-ch-sum': ok ? 'ok' : 'bad', 'aria-live': 'polite' },
    h('div', { class: 'flex flex-wrap items-center justify-between gap-2' },
      h('span', { class: 'text-sm font-medium tabular-nums', 'data-ch-total': String(total), text: t('ch.sum', { n: fmtNum(total) }) }),
      h('span', { class: 'text-xs tabular-nums', text: ok ? t('ch.sum.ok') : left > 0 ? t('ch.sum.left', { n: fmtNum(left) }) : t('ch.sum.over', { n: fmtNum(-left) }) })),
    chanceBar(rows, v.prizeType));
}

function chanceStats(v) {
  if (!v.id) return h('span', {});
  return h('div', { class: 'flex flex-wrap gap-3 text-sm text-secondary-700 dark:text-secondary-200 tabular-nums', 'data-ch-stats': '' },
    h('span', { text: t('ch.stat.plays', { n: fmtNum(v.plays || 0) }) }),
    h('span', { text: t('ch.stat.players', { n: fmtNum(v.players || 0) }) }),
    h('span', { text: t('ch.stat.wins', { n: fmtNum(v.wins || 0) }) }),
    v.lastPlayAt ? h('span', { text: t('ch.stat.last', { date: fmtDate(v.lastPlayAt) }) }) : null);
}

RESOURCES['chance-games'] = {
  filters: [
    { name: 'kind', label: 'filter.ch.kind', options: opts('ch.kind.', ['wheel', 'chest']) },
    { name: 'prizeType', label: 'filter.ch.type', options: opts('ch.type.', CHANCE_PRIZE_TYPES) },
    { name: 'mode', label: 'filter.ch.mode', options: opts('ch.modeName.', ['free', 'ticket']) },
    activeFilter()
  ],
  columns: [
    { label: 'col.name', render: function (it) {
      return h('div', { class: 'flex items-center gap-3 min-w-0' }, chanceTile(),
        h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) || '' }), h('div', { class: 'text-secondary-500' }, mono(it.key))));
    }, skeleton: '40%' },
    { label: 'col.ch.table', render: function (it) {
      return h('div', { class: 'flex flex-col gap-1 min-w-0' }, chanceBar(it.prize, it.prizeType),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums', 'data-ch-rows': '', text: chanceRowsSummary(it) }));
    }, skeleton: '50%' },
    { label: 'col.ch.play', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, chanceModes(it)); }, width: 200 },
    { label: 'col.ch.plays', render: function (it) {
      return h('div', { class: 'flex flex-col text-sm tabular-nums whitespace-nowrap' }, h('span', { text: t('ch.stat.plays', { n: fmtNum(it.plays || 0) }) }),
        h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: t('ch.stat.players', { n: fmtNum(it.players || 0) }) }));
    }, width: 120 },
    { label: 'col.status', render: function (it) {
      return h('div', { class: 'flex flex-wrap gap-1' }, activeTag(it.isActive), h('span', { 'data-ch-kind': it.kind || 'wheel' }, tag(t('ch.kind.' + (it.kind || 'wheel')), it.kind === 'chest' ? 'primary' : 'secondary')), tag(t('ch.type.' + it.prizeType), it.prizeType === 'prizeGroup' ? 'warning' : 'secondary'));
    }, width: 150 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive' }
    ] },
    tr('name', { label: 'f.ch.name', required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { name: 'kind', type: 'select', label: 'f.ch.kind', hint: 'f.ch.kind.hint', required: true, noPlaceholder: true, options: opts('ch.kind.', ['wheel', 'chest']) },
    { name: 'stats', type: 'static', readonly: true, render: chanceStats },
    { type: 'heading', label: 'ch.h.table' },
    { name: 'prizeType', type: 'select', label: 'f.ch.prizeType', hint: 'f.ch.prizeType.hint', required: true, noPlaceholder: true, refresh: true, options: opts('ch.type.', CHANCE_PRIZE_TYPES) },
    { name: 'slots', type: 'rows', addLabel: 'ch.slot.add', emptyText: 'ch.row.empty', min: 1, max: 50, empty: [], when: chSlots,
      newRow: function () { return { chance: 50, kind: 'currency', currencyId: null, min: 1, max: 1, prizeId: null, prizeGroupId: null }; },
      check: function (rows) {
        var any = false;
        for (var i = 0; i < rows.length; i++) {
          var r = rows[i];
          if (Number(r.chance) > 0) any = true;
          if (r.kind === 'currency' && (!r.currencyId || !(Number(r.min) >= 1) || Number(r.max) < Number(r.min))) return { index: i, message: t('code.INVALID_SLOT_RANGE') };
          if (r.kind === 'prize' && !r.prizeId) return { index: i, message: t('f.ch.prize') + ': ' + t('err.required') };
          if (r.kind === 'prizeGroup' && !r.prizeGroupId) return { index: i, message: t('f.ch.group') + ': ' + t('err.required') };
        }
        return any ? null : { index: 0, message: t('code.SLOTS_ALL_ZERO') };
      },
      columns: [
        { name: 'chance', type: 'number', int: false, label: 'f.ch.chance', required: true, min: 0, max: 100 },
        { name: 'kind', type: 'select', label: 'f.ch.slotKind', options: function () { return opts('ch.slotKind.', CH_SLOT_KINDS)(); } },
        { name: 'currencyId', type: 'ref', resource: 'currencies', label: 'f.ch.currency', optional: true },
        { name: 'min', type: 'number', label: 'f.ch.min', optional: true, min: 1 },
        { name: 'max', type: 'number', label: 'f.ch.max', optional: true, min: 1 },
        { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.ch.prize', optional: true },
        { name: 'prizeGroupId', type: 'ref', resource: 'prize-groups', label: 'f.ch.group', optional: true }
      ] },
    { name: 'slotsHint', type: 'static', readonly: true, when: chSlots, render: function () { return h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', 'data-ch-slots-hint': '', text: t('ch.slots.hint') }); } },
    { name: 'rerollEmpty', type: 'bool', inline: true, label: 'f.ch.reroll', hint: 'f.ch.reroll.hint', when: chSlots },
    { name: 'prize', type: 'rows', when: chNotSlots, addLabel: 'ch.row.add', emptyText: 'ch.row.empty', min: 1, max: 50, empty: [], refresh: true,
      newRow: function (rows) {
        var left = Math.round((100 - chanceSum(rows)) * 10000) / 10000;
        return { chance: left > 0 ? left : 0, prizeId: null, prizeGroupId: null };
      },
      check: function (rows) {
        var total = chanceSum(rows);
        if (Math.round(total * 10000) !== 1000000) return { index: rows.length - 1, message: t('code.TOTAL_CHANCE_MUST_BE_100') + ' ' + t('ch.sum', { n: fmtNum(total) }) };
        return null;
      },
      columns: [
        { name: 'chance', type: 'number', int: false, label: 'f.ch.chance', required: true, min: 0, max: 100 },
        { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.ch.prize', optional: true },
        { name: 'prizeGroupId', type: 'ref', resource: 'prize-groups', label: 'f.ch.group', optional: true }
      ] },
    { name: 'sum', type: 'static', readonly: true, live: true, when: chNotSlots, render: chanceSumBlock },
    { type: 'row', when: chNotSlots, fields: [
      { name: 'jackpotRow', type: 'number', label: 'f.ch.jackpot', hint: 'f.ch.jackpot.hint', optional: true, min: 1, max: 50, when: chNotSlots },
      { name: 'pityThreshold', type: 'number', label: 'f.ch.pity', hint: 'f.ch.pity.hint', optional: true, min: 1, max: 100000, when: chNotSlots }
    ] },
    { name: 'rowsHint', type: 'static', readonly: true, when: chNotSlots, render: function () { return h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('ch.rows.hint') }); } },
    { type: 'heading', label: 'ch.h.play' },
    { name: 'intervalTime', type: 'number', label: 'f.ch.interval', hint: 'f.ch.interval.hint', optional: true, min: 1, max: 525600 },
    { type: 'row', fields: [
      { name: 'price.currencyId', type: 'ref', resource: 'currencies', label: 'f.ch.currency', optional: true,
        check: function (v, vals) { var a = vals.price && vals.price.amount; return (v && (a == null || a === '')) || (!v && a != null && a !== '') ? t('code.INVALID_PRICE_PAYLOAD') : null; } },
      { name: 'price.amount', type: 'number', label: 'f.ch.amount', hint: 'f.ch.amount.hint', optional: true, min: 1 }
    ] },
    { name: 'ladderText', type: 'text', ltr: true, label: 'f.ch.ladder', hint: 'f.ch.ladder.hint', optional: true },
    { name: 'revealAck', type: 'bool', inline: true, label: 'f.ch.revealAck', hint: 'f.ch.revealAck.hint' },
    { name: 'playHint', type: 'static', readonly: true, live: true, render: function (v) {
      var none = !v.intervalTime && !(v.price && v.price.currencyId);
      return h('p', { class: 'text-xs ' + (none ? 'text-error-600 dark:text-error-400' : 'text-secondary-500 dark:text-secondary-400'), 'data-ch-play-hint': none ? 'none' : '', text: none ? t('code.INTERVAL_OR_PRICE_REQUIRED') : t('ch.play.hint') });
    } },
    { type: 'heading', label: 'ch.h.limits' },
    { type: 'row', fields: [
      { name: 'dailyLimit', type: 'number', label: 'f.ch.dailyLimit', hint: 'f.ch.dailyLimit.hint', optional: true, min: 1, max: 100000 },
      { name: 'totalLimit', type: 'number', label: 'f.ch.totalLimit', hint: 'f.ch.totalLimit.hint', optional: true, min: 1, max: 10000000 }
    ] },
    { name: 'showOdds', type: 'bool', inline: true, label: 'f.ch.showOdds', hint: 'f.ch.showOdds.hint' },
    { type: 'heading', label: 'ch.h.more' },
    { name: 'metadata', type: 'json', label: 'f.metadata', hint: 'f.ch.metadata.hint', optional: true, empty: {} },
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
  ],
  defaults: function () {
    return { key: '', isActive: true, name: {}, description: {}, kind: 'wheel', prizeType: 'prize', metadata: {}, sort: 0, intervalTime: 1440, price: { currencyId: null, amount: null }, dailyLimit: null, totalLimit: null, showOdds: true, ladderText: '', revealAck: false, rerollEmpty: true, jackpotRow: null, pityThreshold: null,
      slots: [{ chance: 70, kind: 'currency', currencyId: null, min: 120, max: 180, prizeId: null, prizeGroupId: null }],
      prize: [{ chance: 30, prizeId: null, prizeGroupId: null }, { chance: 40, prizeId: null, prizeGroupId: null }, { chance: 30, prizeId: null, prizeGroupId: null }] };
  },
  toForm: function (v) {
    if (!v.price) v.price = { currencyId: null, amount: null };
    var rows = v.prize || [];
    v.ladderText = (v.priceLadder || []).join(', ');
    v.jackpotRow = null;
    rows.forEach(function (r, i) { if (r.jackpot) v.jackpotRow = i + 1; });
    if (v.prizeType === 'slots') {
      v.slots = rows.map(function (r) { return { chance: r.chance, kind: r.kind || 'prize', currencyId: r.currencyId || null, min: r.min || null, max: r.max || null, prizeId: r.prizeId || null, prizeGroupId: r.prizeGroupId || null }; });
      v.prize = [{ chance: 100, prizeId: null, prizeGroupId: null }];
    } else {
      v.slots = [];
      v.prize = rows.map(function (r) { return { chance: r.chance, prizeId: r.prizeId || null, prizeGroupId: r.prizeGroupId || null }; });
    }
    if (v.rerollEmpty == null) v.rerollEmpty = true;
    return v;
  },
  fixPayload: function (p) {
    var group = p.prizeType === 'prizeGroup';
    if (p.prizeType === 'slots') {
      p.prize = (p.slots || []).map(function (r) {
        var k = r.kind || 'prize';
        var o = { chance: r.chance, kind: k };
        if (k === 'currency') { o.currencyId = r.currencyId || ''; o.min = r.min; o.max = r.max; }
        if (k === 'prize') o.prizeId = r.prizeId || '';
        if (k === 'prizeGroup') o.prizeGroupId = r.prizeGroupId || '';
        return o;
      });
      p.pityThreshold = null;
    } else {
      var jr = Number(p.jackpotRow) || 0;
      p.prize = (p.prize || []).map(function (r, i) { var o = { chance: r.chance, prizeId: group ? '' : (r.prizeId || ''), prizeGroupId: group ? (r.prizeGroupId || '') : '' }; if (i + 1 === jr) o.jackpot = true; return o; });
      if (p.pityThreshold === '' || p.pityThreshold === undefined) p.pityThreshold = null;
    }
    p.priceLadder = chLadder(p.ladderText);
    p.revealAck = !!p.revealAck;
    p.rerollEmpty = p.rerollEmpty !== false;
    delete p.slots; delete p.ladderText; delete p.jackpotRow; delete p.slotsHint;
    if (!p.price || !p.price.currencyId) p.price = null;
    if (p.intervalTime === '' || p.intervalTime === undefined) p.intervalTime = null;
    if (p.sort == null) p.sort = 0;
    if (!p.metadata) p.metadata = {};
    if (!p.kind) p.kind = 'wheel';
    if (p.dailyLimit === '' || p.dailyLimit === undefined) p.dailyLimit = null;
    if (p.totalLimit === '' || p.totalLimit === undefined) p.totalLimit = null;
    p.showOdds = p.showOdds !== false;
    delete p.stats; delete p.sum; delete p.rowsHint; delete p.playHint;
    return p;
  },
  afterSave: function () { forgetRefs('chance-games'); }
};

function groupShares(v) {
  var rows = v.members || [];
  var total = 0;
  rows.forEach(function (r) { total += Number(r.weight) > 0 ? Number(r.weight) : 0; });
  return h('div', { class: 'eco-ch-sum rounded-md px-3 py-2 flex flex-col gap-1 text-sm', 'data-pg-shares': '' },
    h('span', { class: 'font-medium', text: t('pg.shares') }),
    rows.length ? rows.map(function (r, i) {
      var w = Number(r.weight) > 0 ? Number(r.weight) : 0;
      var pct = total ? Math.round(w / total * 1000) / 10 : 0;
      return h('div', { class: 'flex items-center gap-2 min-w-0', 'data-pg-share': String(i) },
        h('span', { class: 'eco-ch-dot', 'data-slice': String(i % CH_SLICE_COLORS) }),
        r.prizeId ? refNameCell('prizes', r.prizeId) : h('span', { class: 'text-secondary-500', text: t('pg.pick') }),
        h('span', { class: 'flex-1' }),
        h('span', { class: 'tabular-nums whitespace-nowrap', text: t('pg.pct', { n: fmtNum(pct) }) }));
    }) : h('span', { class: 'text-secondary-500', text: t('pg.empty') }));
}

RESOURCES['prize-groups'] = {
  filters: [
    { name: 'prizeId', label: 'filter.pg.prize', type: 'ref', resource: 'prizes' }
  ],
  columns: [
    { label: 'col.name', render: function (it) {
      return h('div', { class: 'flex items-center gap-3 min-w-0' }, h('div', { class: 'h-9 w-9 shrink-0 rounded-md flex items-center justify-center eco-ch-tile', 'data-group': '' }, icon('packageDuo', 'h-5 w-5')),
        h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium truncate', text: i18nText(it.name) || '' }), h('div', { class: 'text-secondary-500' }, mono(it.key))));
    }, skeleton: '40%' },
    { label: 'col.pg.prizes', render: function (it) {
      var m = it.members || [];
      var box = h('div', { class: 'flex flex-col gap-1 min-w-0 text-sm' }, h('span', { class: 'tabular-nums', 'data-pg-count': '', text: t('pg.prizesN', { n: fmtNum(m.length) }) }));
      m.slice(0, 3).forEach(function (r) { box.appendChild(h('div', { class: 'min-w-0 text-xs text-secondary-600 dark:text-secondary-300' }, refNameCell('prizes', r.prizeId))); });
      return box;
    }, skeleton: '45%' },
    { label: 'col.pg.games', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', 'data-pg-games': '', text: t('pg.gamesN', { n: fmtNum(it.games || 0) }) }); }, width: 140 }
  ],
  fields: [
    { name: 'key', type: 'key', label: 'f.key', hint: 'f.key.hint', required: true, lockOnEdit: true },
    tr('name', { label: 'f.pg.name', required: true, max: 128 }),
    tr('description', { multiline: true, max: 1000, optional: true }),
    { type: 'heading', label: 'pg.h.prizes' },
    { name: 'members', type: 'rows', addLabel: 'pg.row.add', emptyText: 'pg.empty', min: 1, max: 100, empty: [], refresh: true,
      newRow: function () { return { prizeId: null, weight: 1 }; },
      check: duplicateCheck('prizeId', 'DUPLICATE_PRIZE'),
      columns: [
        { name: 'prizeId', type: 'ref', resource: 'prizes', label: 'f.prize', required: true },
        { name: 'weight', type: 'number', label: 'f.pg.weight', required: true, min: 1, max: 1000000 }
      ] },
    { name: 'shares', type: 'static', readonly: true, live: true, render: groupShares },
    { name: 'membersHint', type: 'static', readonly: true, render: function () { return h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400', text: t('pg.hint') }); } },
    { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.sort.hint' }
  ],
  defaults: function () { return { key: '', name: {}, description: {}, members: [{ prizeId: null, weight: 1 }], sort: 0 }; },
  fixPayload: function (p) {
    p.members = (p.members || []).map(function (r) { return { prizeId: r.prizeId || '', weight: r.weight == null ? 1 : r.weight }; });
    if (p.sort == null) p.sort = 0;
    delete p.shares; delete p.membersHint;
    return p;
  },
  afterSave: function () { forgetRefs('prize-groups'); }
};

var PCH_CARD = { id: null, data: null, loading: false, at: 0, err: null, limit: 10 };

function loadChanceCard(id) {
  PCH_CARD.loading = true;
  return api('GET', '/users/' + encodeURIComponent(id) + '/chance-plays' + qs({ limit: PCH_CARD.limit })).then(function (res) {
    PCH_CARD.loading = false;
    PCH_CARD.at = Date.now();
    PCH_CARD.data = res;
    PCH_CARD.err = null;
    var old = document.querySelector('[data-player-chance]');
    if (old) old.removeAttribute('data-key');
    schedule();
  }, function (err) {
    PCH_CARD.loading = false;
    PCH_CARD.at = Date.now();
    PCH_CARD.err = err;
    schedule();
  });
}

function chancePlayLine(p) {
  var prize = p.prizeId ? (i18nText(p.prizeName) || p.prizeKey) : '';
  var spent = p.price && p.price.amount ? t('pch.spent', { n: fmtNum(p.price.amount), name: keyName('currency', p.price.currencyKey) }) : '';
  return h('div', { class: 'flex flex-wrap items-center gap-2 py-2 text-sm', 'data-pch': p.id, 'data-won': p.prizeId ? 'true' : 'false' },
    chanceTile('eco-ch-tile-sm'),
    h('div', { class: 'min-w-0 flex flex-col' },
      h('span', { class: 'font-medium truncate', text: i18nText(p.gameName) || p.gameKey }),
      h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums', text: fmtDate(p.createdAt) })),
    tag(t('ch.modeName.' + p.mode), p.mode === 'ticket' ? 'primary' : 'success'),
    p.prizeId ? tag(t('pch.won', { name: prize }), 'warning') : tag(t('pch.nothing'), 'secondary'),
    p.groupKey ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 truncate', text: t('pch.fromGroup', { name: i18nText(p.groupName) || p.groupKey }) }) : null,
    spent ? h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300 tabular-nums whitespace-nowrap', 'data-pch-spent': '', text: spent }) : null,
    p.status === 'pending' ? tag(t('pch.pending'), 'error') : null,
    h('span', { class: 'flex-1' }),
    h('span', { class: 'text-xs text-secondary-500 dark:text-secondary-400 tabular-nums whitespace-nowrap', text: t('pch.roll', { n: fmtNum(p.roll) }) }));
}

function playerChanceCard() {
  var id = playerCardRoute();
  var existing = document.querySelector('[data-player-chance]');
  if (!id || !token() || !aclCan('HIRO_ECONOMY', 0)) { if (existing) existing.remove(); return; }
  var main = null;
  var mains = document.querySelectorAll('#main > main');
  for (var i = 0; i < mains.length; i++) if (mains[i].style.display !== 'none') main = mains[i];
  var grid = main && main.querySelector('div.flex.flex-col.flex-1.min-w-0 > div.grid.grid-cols-1.gap-4');
  if (!grid) return;
  if (PCH_CARD.id !== id) { PCH_CARD.id = id; PCH_CARD.data = null; PCH_CARD.err = null; PCH_CARD.at = 0; PCH_CARD.limit = 10; }
  if (!PCH_CARD.loading && Date.now() - PCH_CARD.at > 20000) loadChanceCard(id);
  var p = PCH_CARD.data;
  var anchor = grid.querySelector('[data-player-achievements]') || grid.querySelector('[data-player-items]') || grid.querySelector('[data-player-tasks]') || grid.querySelector('[data-player-activity]') || grid.querySelector('[data-player-progress]') || grid.querySelector('[data-player-mod]');
  var key = lang() + '|' + id + '|' + (p ? JSON.stringify(p.items) + p.total : PCH_CARD.err ? 'err' : 'loading');
  var placed = existing && existing.parentNode === grid && (anchor ? existing.previousElementSibling === anchor : grid.firstChild === existing);
  if (existing && existing.getAttribute('data-key') === key && placed) return;
  if (existing) existing.remove();
  var body = h('div', { class: 'px-4 py-3' });
  var cardEl = h('section', { class: 'flex flex-col dark:bg-secondary-800 rounded-md shadow bg-white', 'data-player-chance': '', 'data-key': key, 'data-economy': 'player-chance', 'data-l10n-skip': '', 'aria-labelledby': 'player-chance-title', dir: document.documentElement.getAttribute('dir') || 'ltr' }, body);
  var title = h('h2', { id: 'player-chance-title', class: 'font-medium flex items-center gap-2' }, icon('chance-games', 'h-4 w-4'), h('span', { text: t('pch.card') }));
  var place = function () { if (anchor) anchor.after(cardEl); else grid.insertBefore(cardEl, grid.firstChild); };
  if (!p) {
    body.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-3' }, title));
    body.appendChild(PCH_CARD.err ? notice('error', errorText(PCH_CARD.err), button({ variant: 'link', size: 'sm', label: t('try.again'), onclick: function () { PCH_CARD.at = 0; loadChanceCard(id); } })) : h('div', { class: 'flex flex-col gap-2' }, bar('40%'), bar('60%')));
    place();
    return;
  }
  var items = p.items || [];
  var wins = items.filter(function (x) { return x.prizeId; }).length;
  body.appendChild(h('div', { class: 'flex flex-wrap items-center justify-between gap-2 mb-2' }, title,
    h('span', { class: 'text-sm text-secondary-600 dark:text-secondary-300 tabular-nums', 'data-pch-counts': '', text: t('pch.counts', { n: fmtNum(p.total || 0), w: fmtNum(wins), shown: fmtNum(items.length) }) }),
    h('a', { href: hashFor('chance-games'), class: 'eco-link text-sm', text: t('pch.all') })));
  body.appendChild(withKeyNames(function () {
    var list = h('div', { class: 'flex flex-col eco-tk-list', 'data-pch-list': '' });
    items.forEach(function (x) { list.appendChild(chancePlayLine(x)); });
    if (!items.length) list.appendChild(h('span', { class: 'text-sm text-secondary-500 dark:text-secondary-400 py-1', 'data-pch-none': '', text: t('pch.none') }));
    return list;
  }));
  if ((p.total || 0) > items.length) {
    var more = button({ variant: 'link', size: 'sm', label: t('pch.more'), onclick: function () { PCH_CARD.limit = Math.min(100, PCH_CARD.limit + 20); PCH_CARD.at = 0; loadChanceCard(id); } });
    more.setAttribute('data-pch-more', '');
    body.appendChild(h('div', { class: 'mt-2' }, more));
  }
  place();
}
