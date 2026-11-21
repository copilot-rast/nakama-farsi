function ikSettingsPage(root) {
  settingsPage(root, 'itemkinds/settings', IK_BASE, function (cur, defs, d) {
    var r = cur.routing || {};
    var on = check(r.enabled, t('s.enabled'), 'routing.enabled');
    var missing = select(r.onMissingRef, [{ value: 'pending', label: t('mr.pending') }, { value: 'reject', label: t('mr.reject') }], { name: 'routing.onMissingRef', label: t('s.onMissingRef') });
    var del = select(r.onDelete, [{ value: 'deactivate', label: t('od.deactivate') }, { value: 'keep', label: t('od.keep') }], { name: 'routing.onDelete', label: t('s.onDelete') });
    var link = check(r.linkWritesUnity, t('s.linkWritesUnity'), 'routing.linkWritesUnity');
    var depth = numInput(r.maxDepth, 'routing.maxDepth', t('s.maxDepth'));
    var keep = numInput(r.keepRuns, 'routing.keepRuns', t('s.keepRuns'));
    var prof = select((cur.inventory || {}).profile || 'ask', ['ask', 'always', 'never'].map(function (x) { return { value: x, label: t('pr.' + x) }; }), { name: 'inventory.profile', label: t('s.profile') });
    var els = [
      section(t('s.routing'), [on, h('div', { class: 'ik-row' }, field(t('s.onMissingRef'), missing), field(t('s.onDelete'), del)), link, h('div', { class: 'ik-row' }, field(t('s.maxDepth'), depth), field(t('s.keepRuns'), keep))], 'routing'),
      section(t('s.inventory'), field(t('s.profile'), prof), 'inventory')
    ];
    var kindForms = [];
    (d.kinds || []).forEach(function (k) {
      var cfg = (cur.kinds || {})[k.key] || {};
      var enabled = check(cfg.enabled !== false, t('s.kindEnabled'), 'kinds.' + k.key + '.enabled');
      var merged = {};
      Object.keys(k.defaults || {}).forEach(function (x) { merged[x] = k.defaults[x]; });
      Object.keys(cfg).forEach(function (x) { merged[x] = cfg[x]; });
      var form = schemaForm(k.settings || [], merged, { name: 'kinds.' + k.key });
      var owners = (k.ownership || []).map(function (o) {
        var s = select(((cur.ownership || {})[k.key] || {})[o.field] || o.value || o['default'] || 'unity', [{ value: 'unity', label: t('own.unity') }, { value: 'console', label: t('own.console') }], { name: 'ownership.' + k.key + '.' + o.field, label: textOf(o.label, o.field) });
        return { o: o, s: s, el: field(textOf(o.label, o.field), s) };
      });
      var ownBox = owners.length ? h('div', {}, h('h3', { class: 'font-medium text-sm mt-2 mb-1', text: t('s.ownership') }), h('p', { class: 'text-xs text-secondary-500 mb-2', text: t('s.hint.ownership') }), h('div', { class: 'ik-row' }, owners.map(function (x) { return x.el; }))) : null;
      kindForms.push({ k: k, enabled: enabled, form: form, owners: owners, base: cfg });
      els.push(section(t('s.kind', null).replace('{name}', textOf(k.label, k.key)), [enabled, form.el, ownBox], 'kind-' + k.key));
    });
    return {
      el: els,
      read: function (out) {
        var errors = [];
        var num = function (el, name) { var v = Number(latinNumber(el.value)); if (!isFinite(v)) { errors.push(t('s.badNumber', { field: name })); return 0; } return Math.round(v); };
        out.routing = { enabled: on.input.checked, onMissingRef: missing.value, onDelete: del.value, linkWritesUnity: link.input.checked, maxDepth: num(depth, t('s.maxDepth')), keepRuns: num(keep, t('s.keepRuns')) };
        out.inventory = { profile: prof.value };
        out.kinds = out.kinds || {};
        out.ownership = out.ownership || {};
        kindForms.forEach(function (kf) {
          var v = kf.form.read(JSON.parse(JSON.stringify(kf.base)), errors);
          v.enabled = kf.enabled.input.checked;
          out.kinds[kf.k.key] = v;
          var own = {};
          kf.owners.forEach(function (x) { if (x.s.value !== (x.o['default'] || 'unity')) own[x.o.field] = x.s.value; });
          out.ownership[kf.k.key] = own;
        });
        return errors[0] || null;
      },
      after: function () { ikLoadKinds(true); }
    };
  });
}
