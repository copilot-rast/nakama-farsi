function fieldLabel(f) {
  return textOf(f.label, f.key);
}

function latinNumber(raw) {
  return String(raw == null ? '' : raw).replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).replace(/٫/g, '.').replace(/٬/g, '').trim();
}

function schemaForm(fields, data, opts) {
  opts = opts || {};
  data = data || {};
  var parts = [];
  var box = h('div', { class: 'ik-form', 'data-ik-form': opts.name || '' });
  (fields || []).forEach(function (f) {
    var part = schemaField(f, data[f.key], opts, opts.name ? opts.name + '.' + f.key : f.key);
    parts.push({ f: f, part: part });
    box.appendChild(part.el);
  });
  return {
    el: box,
    read: function (out, errors) {
      out = out || {};
      parts.forEach(function (p) {
        var v = p.part.read(errors);
        if (v === undefined) delete out[p.f.key];
        else out[p.f.key] = v;
      });
      return out;
    }
  };
}

function schemaField(f, value, opts, path) {
  var label = fieldLabel(f);
  var ro = !!(f.readOnly && opts.lockReadOnly);
  var hint = textOf(f.hint, '');
  if (f.readOnly) hint = hint ? hint + ' · ' + t('k.readonly') : t('k.readonly');
  var named = function (el) { el.setAttribute('data-ik-field', path); if (ro) el.setAttribute('disabled', ''); return el; };
  switch (f.type) {
    case 'bool': {
      var c = check(value === undefined ? !!f['default'] : !!value, label, path);
      named(c.input);
      return { el: c, read: function () { return c.input.checked; } };
    }
    case 'number':
    case 'int': {
      var n = named(numInput(value, path, label));
      return { el: field(label, n, hint), read: function (errors) {
        var raw = latinNumber(n.value);
        if (raw === '') return undefined;
        var v = Number(raw);
        if (!isFinite(v) || (f.type === 'int' && Math.round(v) !== v)) { errors.push(t('fm.badNumber', { field: label })); return value; }
        return v;
      } };
    }
    case 'select': {
      var list = [{ value: '', label: t('none') }].concat((f.options || []).map(function (o) { return { value: o, label: o }; }));
      var s = named(select(value == null ? '' : value, list, { name: path, label: label }));
      return { el: field(label, s, hint), read: function () { return s.value === '' ? undefined : s.value; } };
    }
    case 'tags': {
      var tg = named(input(Array.isArray(value) ? value.join(', ') : '', { ltr: true, name: path, label: label }));
      return { el: field(label, tg, hint || t('fm.tagsHint')), read: function () {
        var out = tg.value.split(/[,،]/).map(function (x) { return x.trim(); }).filter(Boolean);
        return out;
      } };
    }
    case 'i18n': {
      var tx = translated(label, path, value || {});
      if (ro) Array.prototype.forEach.call(tx.querySelectorAll('input,textarea'), function (x) { x.setAttribute('disabled', ''); });
      return { el: tx, read: function () { return tx.read(); } };
    }
    case 'group': {
      var inner = schemaForm(f.fields, value || {}, { name: path, lockReadOnly: opts.lockReadOnly || f.readOnly });
      var fs = h('fieldset', { class: 'ik-group mb-4', 'data-ik-group': path }, h('legend', { class: 'font-medium text-sm px-1', text: label }), hint ? h('p', { class: 'text-xs text-secondary-500 mb-2', text: hint }) : null, inner.el);
      return { el: fs, read: function (errors) { return inner.read(JSON.parse(JSON.stringify(value || {})), errors); } };
    }
    case 'list': {
      var rows = [];
      var holder = h('div', { class: 'ik-rows' });
      var addRow = function (v) {
        var form = schemaForm(f.fields, v || {}, { name: path + '[]', lockReadOnly: opts.lockReadOnly || f.readOnly });
        var rec = { form: form, value: v || {} };
        var del = ro || f.readOnly ? null : button({ label: t('fm.removeRow'), variant: 'danger', onclick: function () { rows.splice(rows.indexOf(rec), 1); rowEl.remove(); } });
        var rowEl = h('div', { class: 'ik-list-row', 'data-ik-row-of': path }, form.el, del);
        rec.el = rowEl;
        rows.push(rec);
        holder.appendChild(rowEl);
      };
      (Array.isArray(value) ? value : []).forEach(addRow);
      var addBtn = ro || f.readOnly ? null : button({ label: t('fm.addRow'), data: { 'ik-add-row': path }, onclick: function () { addRow({}); } });
      var wrap = h('fieldset', { class: 'ik-group mb-4', 'data-ik-group': path }, h('legend', { class: 'font-medium text-sm px-1', text: label }), hint ? h('p', { class: 'text-xs text-secondary-500 mb-2', text: hint }) : null, holder, addBtn);
      return { el: wrap, read: function (errors) { return rows.map(function (r) { return r.form.read(JSON.parse(JSON.stringify(r.value)), errors); }); } };
    }
    case 'json': {
      var js = named(input(value === undefined ? '' : JSON.stringify(value, null, 2), { multiline: true, rows: 5, ltr: true, mono: true, name: path, label: label }));
      return { el: field(label, js, hint), read: function (errors) {
        var raw = js.value.trim();
        if (!raw) return undefined;
        try { return JSON.parse(raw); } catch (e) { errors.push(t('fm.badJson', { field: label })); return value; }
      } };
    }
    default: {
      var tx2 = named(input(value == null ? '' : value, { name: path, label: label, ltr: !!f.ltr }));
      return { el: field(label, tx2, hint), read: function () { var v = tx2.value.trim(); return v === '' ? undefined : v; } };
    }
  }
}

function statusTag(st) {
  var color = st === 'ok' ? 'success' : st === 'pending' ? 'warning' : st === 'deleted' ? 'error' : 'secondary';
  return tag(t('st.' + st), color);
}

function actionTag(a) {
  var color = a === 'created' || a === 'linked' ? 'success' : a === 'updated' ? 'primary' : a === 'pending' ? 'warning' : a === 'error' || a === 'deactivated' ? 'error' : 'secondary';
  return tag(t('act.' + a), color);
}

function priceText(p) {
  if (!p || !p.currencyKey) return t('none');
  return fmtNum(p.amount) + ' ' + p.currencyKey;
}

function valueAt(obj, path) {
  var cur = obj;
  String(path || '').split('.').forEach(function (k) { cur = cur == null ? cur : cur[k]; });
  return cur;
}

function plainValue(v) {
  if (v == null || v === '') return t('none');
  if (Array.isArray(v)) return digits(String(v.length));
  if (typeof v === 'number') return fmtNum(v);
  if (typeof v === 'boolean') return v ? t('yes') : t('no');
  if (typeof v === 'object') return textOf(v, JSON.stringify(v));
  return String(v);
}
