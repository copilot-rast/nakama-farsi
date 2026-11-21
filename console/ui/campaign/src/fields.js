function cgAppliesTo(d, scope) {
  var keys = d.chapters || [], kinds = d.kinds || [];
  if (!keys.length && !kinds.length) return true;
  return keys.indexOf(scope.key) !== -1 || (!!scope.kind && kinds.indexOf(scope.kind) !== -1);
}

function cgScope(c) {
  return { key: (c && c.key) || '', kind: (c && c.metadata && c.metadata.kind) || '' };
}

function cgDefaultText(d) {
  if (d.default == null) return '';
  if (d.type === 'bool') return d.default ? t('yes') : t('no');
  if (d.type === 'tags') return (d.default || []).join(', ');
  if (typeof d.default === 'number') return fmtNum(d.default);
  return String(d.default);
}

function fieldsForm(defs, values, scope, prefix) {
  values = values || {};
  prefix = prefix || 'settings';
  var box = h('div', { class: 'cg-fields', 'data-campaign-fields': prefix });
  var readers = [];
  (defs || []).forEach(function (d) {
    if (scope && !cgAppliesTo(d, scope)) return;
    var label = textOf(d.label, d.key);
    var hint = textOf(d.hint, '');
    var name = prefix + '.' + d.key;
    var has_ = Object.prototype.hasOwnProperty.call(values, d.key) && values[d.key] != null;
    var v = values[d.key];
    var def = cgDefaultText(d);
    var withDef = def ? (hint ? hint + ' · ' : '') + t('fd.default', { v: def }) : hint;
    var el, read;
    switch (d.type) {
      case 'number':
      case 'integer':
        el = input(has_ ? digits(String(v)) : '', { ltr: true, name: name, label: label, placeholder: def });
        el.setAttribute('inputmode', d.type === 'integer' ? 'numeric' : 'decimal');
        el.classList.add('cg-num');
        read = function () { var n = d.type === 'integer' ? numberValue(el) : floatValue(el); return n == null ? undefined : n; };
        box.appendChild(field(label, el, withDef, 'cg-num-field'));
        break;
      case 'bool':
        var start = has_ ? !!v : !!d.default;
        el = check(start, label, name);
        read = function () { var c = el.input.checked; return has_ || c !== !!d.default ? c : undefined; };
        box.appendChild(h('div', {}, el, hint ? h('p', { class: 'text-xs text-secondary-500 -mt-2 mb-3', text: hint }) : null));
        break;
      case 'select':
        el = select(has_ ? v : '', [{ value: '', label: def ? t('fd.defaultOpt', { v: def }) : t('none') }].concat((d.options || []).map(function (o) { return { value: o, label: has('opt.' + o) ? t('opt.' + o) + ' · ' + o : o }; })), { name: name, label: label });
        read = function () { return el.value || undefined; };
        box.appendChild(field(label, el, hint));
        break;
      case 'tags':
        el = input(has_ ? (v || []).join(', ') : '', { ltr: true, mono: true, name: name, label: label, placeholder: def });
        read = function () {
          var raw = el.value.trim();
          if (!raw) return has_ ? [] : undefined;
          return raw.split(/[,،]/).map(function (x) { return x.trim(); }).filter(Boolean);
        };
        box.appendChild(field(label, el, withDef ? withDef + ' · ' + t('fd.tagsHint') : t('fd.tagsHint')));
        break;
      case 'i18n':
        el = translated(label, name, has_ ? v : {}, { multiline: true });
        read = function () { var o = el.read(); return Object.keys(o).length ? o : undefined; };
        box.appendChild(h('div', {}, el, hint ? h('p', { class: 'text-xs text-secondary-500 -mt-3 mb-3', text: hint }) : null));
        break;
      case 'list':
        var items = [];
        var list = h('div', { class: 'cg-rows', 'data-campaign-list-field': d.key });
        var addItem = function (item) {
          var sub = fieldsForm(d.fields, item || {}, null, d.key);
          var row = h('div', { class: 'cg-track', 'data-campaign-list-item': d.key });
          var entry = { row: row, sub: sub };
          row.appendChild(h('div', { class: 'flex items-center justify-between gap-2 mb-2' },
            h('span', { class: 'text-sm font-medium', 'data-campaign-list-title': '' }),
            button({ label: t('remove'), variant: 'danger', onclick: function () { row.remove(); items = items.filter(function (x) { return x !== entry; }); renumber(); } })));
          row.appendChild(sub);
          items.push(entry);
          list.appendChild(row);
          renumber();
        };
        var renumber = function () { items.forEach(function (x, i) { x.row.querySelector('[data-campaign-list-title]').textContent = t('fd.item', { label: label, n: i + 1 }); }); };
        (has_ ? v : []).forEach(addItem);
        var max = d.maxItems || 20;
        read = function () { var out = items.map(function (x) { return x.sub.read(); }); return has_ || out.length ? out : undefined; };
        box.appendChild(h('div', { class: 'mb-4', 'data-campaign-list': d.key },
          h('div', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1', text: label }),
          hint ? h('p', { class: 'text-xs text-secondary-500 mb-2', text: hint }) : null, list,
          button({ label: '+ ' + t('fd.add', { label: label }), data: { 'campaign-action': 'add-' + d.key }, onclick: function () { if (items.length < max) addItem({}); else toast('error', t('code.TOO_MANY_ITEMS')); } })));
        break;
      default:
        el = input(has_ ? v : '', { name: name, label: label, placeholder: def, ltr: /^[A-Za-z0-9_.:\-\s]*$/.test(String(has_ ? v : def)) });
        read = function () { var s = el.value.trim(); return s ? s : (has_ ? '' : undefined); };
        box.appendChild(field(label, el, withDef));
    }
    readers.push({ key: d.key, read: read });
  });
  box.read = function () {
    var out = {};
    readers.forEach(function (r) { var v = r.read(); if (v !== undefined && !(typeof v === 'number' && isNaN(v))) out[r.key] = v; });
    return out;
  };
  box.invalid = function () {
    var bad = null;
    readers.forEach(function (r) { var v = r.read(); if (typeof v === 'number' && isNaN(v)) bad = bad || r.key; });
    return bad;
  };
  box.empty = !readers.length;
  return box;
}
