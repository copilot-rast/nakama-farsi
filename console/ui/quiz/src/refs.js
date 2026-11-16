var qzRefs = null;
function loadRefs(force) {
  if (qzRefs && !force) return Promise.resolve(qzRefs);
  return api('GET', '/quiz/refs' + qs({ lang: qzLang() })).then(function (r) { qzRefs = r || {}; return qzRefs; }, function () { qzRefs = {}; return qzRefs; });
}

var qzSettingsCache = null;
function loadQzSettings() {
  if (qzSettingsCache) return Promise.resolve(qzSettingsCache);
  return api('GET', '/quiz/settings').then(function (d) { qzSettingsCache = d; return d; });
}

function refOptions(kind, value, emptyLabel) {
  var list = (qzRefs && qzRefs[kind]) || [];
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

function currencyRows(values, name) {
  var box = h('div', { class: 'qz-rows', 'data-quiz-currencies': name || '' });
  var rows = [];
  function addRow(k, v) {
    var key = refSelect('currencies', k || '', 'reward-currency', '—');
    var amt = numInput(v == null ? '' : v, 'reward-amount', t('rw.amount'));
    var row = h('div', { class: 'qz-row' }, field(t('rw.currency'), key), field(t('rw.amount'), amt),
      h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r.row !== row; }); } })));
    rows.push({ row: row, key: key, amt: amt });
    box.appendChild(row);
  }
  Object.keys(values || {}).forEach(function (k) { addRow(k, values[k]); });
  var wrap = h('div', {}, box, button({ label: '+ ' + t('rw.addCurrency'), data: { 'quiz-action': 'add-currency-' + (name || '') }, onclick: function () { addRow(); } }));
  wrap.read = function () {
    var out = {};
    rows.forEach(function (r) { var k = r.key.value; var n = numberValue(r.amt); if (k && n) out[k] = n; });
    return out;
  };
  return wrap;
}

function rewardEditor(value, name) {
  value = value || {};
  var cur = currencyRows(value.currencies, name);
  var prize = refSelect('prizes', value.prizeKey, name + '-prize', t('rw.noPrize'));
  var group = refSelect('prizeGroups', value.prizeGroupKey, name + '-prize-group', t('rw.noPrize'));
  var el = h('div', { 'data-quiz-reward': name }, cur, h('div', { class: 'grid sm:grid-cols-2 gap-x-4 mt-3' }, field(t('rw.prize'), prize), field(t('rw.group'), group)));
  el.read = function () {
    var o = { currencies: cur.read() };
    if (prize.value) o.prizeKey = prize.value;
    if (group.value) o.prizeGroupKey = group.value;
    return o;
  };
  return el;
}

function rewardText(r) {
  r = r || {};
  var cur = r.currencies || {};
  var parts = Object.keys(cur).map(function (k) { return fmtNum(cur[k]) + ' ' + k; });
  if (r.prizeKey) parts.push('🎁 ' + r.prizeKey);
  if (r.prizeGroupKey) parts.push('🎁 ' + r.prizeGroupKey);
  return parts.length ? '⁦' + parts.join(' + ') + '⁩' : t('none');
}

function rewardListText(list) {
  if (!list || !list.length) return t('none');
  return '⁦' + list.map(function (r) { return fmtNum(r.amount) + ' ' + r.currencyKey; }).join(' + ') + '⁩';
}

function tagsInput(list, name) {
  var el = input((list || []).join('، '), { name: name, placeholder: t('hint.tags') });
  el.read = function () { return el.value.split(/[,،\n]/).map(function (s) { return s.trim(); }).filter(Boolean); };
  return el;
}

function triSelect(value, name) {
  var el = select(value === true ? 'yes' : value === false ? 'no' : '', [{ value: '', label: t('default') }, { value: 'yes', label: t('yes') }, { value: 'no', label: t('no') }], { name: name, label: name });
  el.read = function () { return el.value === 'yes' ? true : el.value === 'no' ? false : null; };
  return el;
}

function jsonInput(value, name) {
  var el = input(JSON.stringify(value || {}), { name: name, ltr: true, mono: true, multiline: true, rows: 2 });
  el.read = function () { try { return JSON.parse(el.value || '{}'); } catch (e) { return undefined; } };
  return el;
}

function optNumber(el) {
  var n = floatValue(el);
  return n == null || isNaN(n) ? null : n;
}

function statusTag(s) {
  return tag(t('st.' + s), s === 'open' ? 'warning' : s === 'resolved' || s === 'done' || s === 'sent' ? 'success' : s === 'failed' ? 'error' : 'secondary');
}
