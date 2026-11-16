// Shared editor parts: economy key lists, currency rows, prize pickers, tags.
var lbRefs = null;
function loadRefs(force) {
  if (lbRefs && !force) return Promise.resolve(lbRefs);
  return api('GET', '/leaderboardx/refs' + qs({ lang: lbLang() })).then(function (r) { lbRefs = r || {}; return lbRefs; }, function () { lbRefs = {}; return lbRefs; });
}

function refOptions(kind, value, emptyLabel) {
  var list = (lbRefs && lbRefs[kind]) || [];
  var opts = emptyLabel != null ? [{ value: '', label: emptyLabel }] : [];
  list.forEach(function (r) { opts.push({ value: r.key, label: (r.name && r.name !== r.key ? r.name + ' · ' : '') + r.key }); });
  if (value && !list.some(function (r) { return r.key === value; })) opts.push({ value: value, label: value + ' (' + t('rw.missing') + ')' });
  return opts;
}

function refSelect(kind, value, name, emptyLabel) {
  var el = select(value || '', refOptions(kind, value, emptyLabel), { name: name, label: name });
  el.setAttribute('dir', 'auto');
  return el;
}

// currencyRows edits {key: amount}; .read() returns the map.
function currencyRows(values, name) {
  var box = h('div', { class: 'lb-rows', 'data-lbx-currencies': name || '' });
  var rows = [];
  function addRow(k, v) {
    var key = refSelect('currencies', k || '', 'reward-currency', '—');
    var amt = numInput(v == null ? '' : v, 'reward-amount', t('rw.amount'));
    var row = h('div', { class: 'lb-row' }, field(t('rw.currency'), key), field(t('rw.amount'), amt),
      h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r.row !== row; }); box.dispatchEvent(new Event('input', { bubbles: true })); } })));
    rows.push({ row: row, key: key, amt: amt });
    box.appendChild(row);
  }
  Object.keys(values || {}).forEach(function (k) { addRow(k, values[k]); });
  var wrap = h('div', {}, box, button({ label: '+ ' + t('rw.addCurrency'), data: { 'lbx-action': 'add-currency' }, onclick: function () { addRow(); } }));
  wrap.read = function () {
    var out = {};
    rows.forEach(function (r) { var k = r.key.value; var n = numberValue(r.amt); if (k && n) out[k] = n; });
    return out;
  };
  return wrap;
}

// prizePickers edits {prizeKey, prizeGroupKey}.
function prizePickers(value) {
  value = value || {};
  var prize = refSelect('prizes', value.prizeKey, 'reward-prize', t('rw.noPrize'));
  var group = refSelect('prizeGroups', value.prizeGroupKey, 'reward-prize-group', t('rw.noPrize'));
  var el = h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('rw.prize'), prize), field(t('rw.group'), group));
  el.read = function () { var o = {}; if (prize.value) o.prizeKey = prize.value; if (group.value) o.prizeGroupKey = group.value; return o; };
  return el;
}

function rewardText(cur, prizes) {
  var parts = Object.keys(cur || {}).map(function (k) { return fmtNum(cur[k]) + ' ' + k; });
  (prizes || []).forEach(function (p) { if (p) parts.push(p); });
  if (!parts.length) return t('none');
  return '⁦' + parts.join(' + ') + '⁩';
}

// rewardsText describes a result's rewards {currencies:[{currencyKey, amount}], prizes:[{prizeKey, prizeGroupKey}]}.
function rewardsText(r) {
  r = r || {};
  var cur = {};
  (r.currencies || []).forEach(function (c) { cur[c.currencyKey] = c.amount; });
  return rewardText(cur, (r.prizes || []).map(function (p) { return p.prizeKey || p.prizeGroupKey; }));
}

function statusTag(st) {
  return tag(t('st.' + st), st === 'active' ? 'success' : st === 'upcoming' ? 'primary' : st === 'ended' ? 'warning' : 'secondary');
}

function resultTag(st) {
  var color = { paid: 'success', pending: 'primary', claimable: 'warning', failed: 'error', expired: 'secondary', none: 'secondary' }[st] || 'secondary';
  var el = tag(t('rs.' + st), color);
  el.setAttribute('data-lbx-status', st);
  return el;
}

function tierName(b, key) {
  if (!key) return t('none');
  var tier = (b && b.tiers || []).filter(function (x) { return x.key === key; })[0];
  return tier ? textOf(tier.name, tier.key) : key;
}

function tierSentence(basis, th, prev) {
  if (basis === 'score') return t('tier.s.score', { v: fmtNum(th) });
  if (basis === 'rank') return t('tier.s.rank', { from: prev ? prev + 1 : 1, v: th });
  if (basis === 'percent') return t('tier.s.percent', { v: th });
  return '';
}

function rewardSentence(r, b) {
  var tn = r.tier ? tierName(b, r.tier) : '';
  if (r.tier && !r.toRank) return tr('rw.s.tierAll', null, { t: tn });
  if (r.tier) return tr('rw.s.tier', { from: r.fromRank || 1, to: r.toRank }, { t: tn });
  if ((r.fromRank || 1) === r.toRank) return t('rw.s.one', { from: r.toRank });
  return t('rw.s.rank', { from: r.fromRank || 1, to: r.toRank });
}

function localInput(iso, name) {
  var el = h('input', { type: 'datetime-local', class: LB_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  if (iso) {
    var d = new Date(iso);
    if (!isNaN(d)) el.value = new Date(d.getTime() - d.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
  }
  el.read = function () { return el.value ? new Date(el.value).toISOString() : null; };
  return el;
}
