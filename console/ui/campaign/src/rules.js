var cgRefs = null;
function loadRefs(force) {
  if (cgRefs && !force) return Promise.resolve(cgRefs);
  return api('GET', '/campaign/refs' + qs({ lang: cgLang() })).then(function (r) { cgRefs = r || {}; return cgRefs; }, function () { cgRefs = {}; return cgRefs; });
}

function refOptions(kind, value, emptyLabel) {
  var list = (cgRefs && cgRefs[kind]) || [];
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
  var box = h('div', { class: 'cg-rows', 'data-campaign-currencies': name || '' });
  var rows = [];
  function addRow(k, v) {
    var key = refSelect('currencies', k || '', 'reward-currency', '—');
    var amt = numInput(v == null ? '' : v, 'reward-amount', t('rw.amount'));
    var row = h('div', { class: 'cg-row' }, field(t('rw.currency'), key), field(t('rw.amount'), amt),
      h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (r) { return r.row !== row; }); box.dispatchEvent(new Event('input', { bubbles: true })); } })));
    rows.push({ row: row, key: key, amt: amt });
    box.appendChild(row);
  }
  Object.keys(values || {}).forEach(function (k) { addRow(k, values[k]); });
  var wrap = h('div', {}, box, button({ label: '+ ' + t('rw.addCurrency'), data: { 'campaign-action': 'add-currency' }, onclick: function () { addRow(); } }));
  wrap.read = function () {
    var out = {};
    rows.forEach(function (r) { var k = r.key.value; var n = numberValue(r.amt); if (k && n) out[k] = n; });
    return out;
  };
  return wrap;
}

function prizePickers(value) {
  value = value || {};
  var prize = refSelect('prizes', value.prizeKey, 'reward-prize', t('rw.noPrize'));
  var group = refSelect('prizeGroups', value.prizeGroupKey, 'reward-prize-group', t('rw.noPrize'));
  var el = h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('rw.prize'), prize), field(t('rw.group'), group));
  el.read = function () { var o = {}; if (prize.value) o.prizeKey = prize.value; if (group.value) o.prizeGroupKey = group.value; return o; };
  return el;
}

var CG_RULE_TYPES = ['level', 'meta', 'item', 'stage', 'chapter'];

function ruleSentence(r) {
  var v = r.value || 0;
  switch (r.type) {
    case 'level': return t('rule.s.level', { v: v }) + (r.table ? ' (' + r.table + ')' : '');
    case 'meta': return tr('rule.s.meta', { v: v }, { k: r.key });
    case 'item': return tr('rule.s.item', null, { k: r.key });
    case 'stage': return v ? tr('rule.s.stage', { v: v }, { k: r.key }) : tr('rule.s.stage0', null, { k: r.key });
    case 'chapter': return v ? tr('rule.s.chapter', { v: v }, { k: r.key }) : tr('rule.s.chapter0', null, { k: r.key });
  }
  return r.type;
}

function ruleEditor(rules, name) {
  var box = h('div', { class: 'cg-rows', 'data-campaign-rules': name || '' });
  var rows = [];
  function keyKind(type) { return { meta: 'metas', item: 'items', stage: 'stages', chapter: 'chapters' }[type]; }
  function addRow(r) {
    r = r || { type: 'level', value: 1 };
    var type = select(r.type, CG_RULE_TYPES.map(function (x) { return { value: x, label: t('rule.' + x) }; }), { name: 'rule-type', label: t('rule.type') });
    var keyBox = h('div', {});
    var value = numInput(r.value == null ? '' : r.value, 'rule-value', t('rule.value'));
    var sentence = h('div', { class: 'text-xs text-secondary-500 cg-sentence', 'data-campaign-rule-sentence': '' });
    var key = null;
    var row = h('div', { class: 'cg-row' });
    function paintKey(cur) {
      clear(keyBox);
      var kind = type.value === 'level' ? 'levelTables' : keyKind(type.value);
      key = refSelect(kind, cur, type.value === 'level' ? 'rule-table' : 'rule-key', type.value === 'level' ? t('se.levelTable.main') : '—');
      keyBox.appendChild(field(type.value === 'level' ? t('rule.table') : t('rule.key'), key));
      value.disabled = type.value === 'item';
      key.addEventListener('change', paint);
    }
    function read() {
      var o = { type: type.value, value: type.value === 'item' ? 0 : (floatValue(value) || 0) };
      if (type.value === 'level') { if (key.value) o.table = key.value; } else o.key = key.value;
      return o;
    }
    function paint() { sentence.textContent = ruleSentence(read()); }
    type.addEventListener('change', function () { paintKey(''); paint(); });
    value.addEventListener('input', paint);
    paintKey(r.type === 'level' ? r.table : r.key);
    row.appendChild(field(t('rule.type'), type));
    row.appendChild(keyBox);
    row.appendChild(field(t('rule.value'), value));
    row.appendChild(h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); rows = rows.filter(function (x) { return x.row !== row; }); } })));
    row.appendChild(h('div', { class: 'cg-row-full' }, sentence));
    rows.push({ row: row, read: read });
    box.appendChild(row);
    paint();
  }
  (rules || []).forEach(addRow);
  var wrap = h('div', {}, box, button({ label: '+ ' + t('rule.add'), data: { 'campaign-action': 'add-rule' }, onclick: function () { addRow(); } }));
  wrap.read = function () { return rows.map(function (r) { return r.read(); }); };
  return wrap;
}

var CG_FIELDS = ['timeSavedPercent', 'healthPercent', 'timesSeen', 'elapsedSeconds'];
var CG_OPS = ['>=', '>', '<=', '<', '==', '!='];

function starSentence(r) {
  var conds = (r.conditions || []).map(function (c) {
    var label = has('star.f.' + c.field) ? t('star.f.' + c.field) : c.field;
    return label + ' ' + c.op + ' ' + fmtNum(c.value) + (c.missing === 'pass' ? t('star.ifMissing') : '');
  });
  return tr('star.sentence', { stars: r.stars }, { conds: conds.length ? conds.join(t('star.and')) : t('star.always') });
}

function starEditor(rules, maxStars) {
  var box = h('div', { class: 'flex flex-col gap-3', 'data-campaign-star-rules': '' });
  var items = [];
  function addRule(r) {
    r = r || { stars: maxStars || 3, conditions: [{ field: 'timeSavedPercent', op: '>=', value: 0 }] };
    var stars = numInput(r.stars, 'star-stars', t('star.stars'));
    var sentence = h('p', { class: 'text-sm font-medium cg-sentence', 'data-campaign-star-sentence': '' });
    var condBox = h('div', { class: 'cg-rows' });
    var conds = [];
    var card_ = h('div', { class: 'cg-track', 'data-campaign-star-rule': '' });
    function addCond(c) {
      c = c || { field: 'healthPercent', op: '>=', value: 0 };
      var custom = CG_FIELDS.indexOf(c.field) === -1;
      var fieldSel = select(custom ? '__custom' : c.field, CG_FIELDS.map(function (f) { return { value: f, label: t('star.f.' + f) }; }).concat([{ value: '__custom', label: t('star.f.custom') }]), { name: 'cond-field', label: t('star.field') });
      var customIn = input(custom ? c.field : '', { name: 'cond-custom', ltr: true, mono: true, placeholder: 'arrests' });
      customIn.hidden = !custom;
      var op = select(c.op, CG_OPS.map(function (o) { return { value: o, label: o }; }), { name: 'cond-op', label: t('star.op') });
      var val = numInput(c.value, 'cond-value', t('star.value'));
      var miss = select(c.missing || '', [{ value: '', label: t('star.missing.default') }, { value: 'pass', label: t('star.missing.pass') }, { value: 'fail', label: t('star.missing.fail') }], { name: 'cond-missing', label: t('star.missing') });
      var row = h('div', { class: 'cg-row' }, field(t('star.field'), h('div', { class: 'flex flex-col gap-1' }, fieldSel, customIn)), field(t('star.op'), op), field(t('star.value'), val), field(t('star.missing'), miss),
        h('div', {}, button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); conds = conds.filter(function (x) { return x.row !== row; }); paint(); } })));
      fieldSel.addEventListener('change', function () { customIn.hidden = fieldSel.value !== '__custom'; paint(); });
      [customIn, op, val, miss].forEach(function (el) { el.addEventListener('input', paint); el.addEventListener('change', paint); });
      conds.push({ row: row, read: function () { var o = { field: fieldSel.value === '__custom' ? customIn.value.trim() : fieldSel.value, op: op.value, value: floatValue(val) || 0 }; if (miss.value) o.missing = miss.value; return o; } });
      condBox.appendChild(row);
    }
    function read() { return { stars: numberValue(stars) || 0, conditions: conds.map(function (c) { return c.read(); }) }; }
    function paint() { sentence.textContent = starSentence(read()); }
    stars.addEventListener('input', paint);
    (r.conditions || []).forEach(addCond);
    card_.appendChild(h('div', { class: 'flex flex-wrap items-end gap-3' }, field(t('star.stars'), stars, null, 'cg-num'), h('div', { class: 'flex-1' }, sentence),
      button({ label: t('remove'), variant: 'danger', onclick: function () { card_.remove(); items = items.filter(function (x) { return x.el !== card_; }); } })));
    card_.appendChild(condBox);
    card_.appendChild(button({ label: '+ ' + t('star.addCond'), data: { 'campaign-action': 'add-cond' }, onclick: function () { addCond(); paint(); } }));
    items.push({ el: card_, read: read });
    box.appendChild(card_);
    paint();
  }
  (rules || []).forEach(addRule);
  var wrap = h('div', {}, h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('star.hint') }), box,
    button({ label: '+ ' + t('star.add'), data: { 'campaign-action': 'add-star-rule' }, onclick: function () { addRule(); } }));
  wrap.read = function () { return items.map(function (i) { return i.read(); }); };
  return wrap;
}

function multiplierInputs(values, maxStars, name) {
  var inputs = [];
  var box = h('div', { class: 'flex flex-wrap gap-3', 'data-campaign-multipliers': name || '' });
  for (var i = 0; i < (maxStars || 3); i++) {
    var el = input(values && values[i] != null ? digits(String(values[i])) : '', { ltr: true, name: (name || 'mult') + '-' + (i + 1), label: String(i + 1) });
    el.setAttribute('inputmode', 'decimal');
    el.classList.add('cg-num');
    inputs.push(el);
    box.appendChild(field('★'.repeat(i + 1), el, null, 'cg-num'));
  }
  box.read = function () { return inputs.map(function (el) { var v = floatValue(el); return v == null || isNaN(v) ? 0 : v; }); };
  box.inputs = inputs;
  return box;
}

function rewardText(cur) {
  var keys = Object.keys(cur || {});
  if (!keys.length) return t('none');
  return '\u2066' + keys.map(function (k) { return fmtNum(cur[k]) + ' ' + k; }).join(' + ') + '\u2069';
}

function rewardListText(list) {
  if (!list || !list.length) return t('none');
  return '\u2066' + list.map(function (r) { return fmtNum(r.amount) + ' ' + r.currencyKey; }).join(' + ') + '\u2069';
}

function modeTag(mode) {
  return tag(t('mode.' + mode), mode === 'ordered' ? 'primary' : mode === 'once' ? 'warning' : 'success');
}

function stars(n, max) {
  var s = '';
  for (var i = 0; i < (max || 3); i++) s += i < n ? '★' : '☆';
  return h('span', { class: 'cg-stars', 'data-campaign-stars': String(n), 'aria-label': String(n), text: s });
}
