function voiceSecretInput(cur, change, label, id) {
  var inp = textInput(cur || '', { ltr: true, mono: true, label: label, oninput: function (v) { change(v.trim()); } });
  inp.type = 'password';
  inp.setAttribute('autocomplete', 'new-password');
  inp.classList.add('eco-secret');
  inp.setAttribute('data-voice-secret', id);
  return inp;
}

function voiceFields(values) {
  var providers = values.providers || [];
  var out = [
    { name: 'enabled', type: 'bool', inline: true, label: 'f.vc.enabled', hint: 'f.vc.enabled.hint' },
    { name: 'provider', type: 'select', label: 'f.vc.provider', hint: 'f.vc.provider.hint', noPlaceholder: true, options: function () {
      return [{ value: '', label: t('vc.provider.none') }].concat(providers.map(function (p) { return { value: p.kind, label: t('vc.provider.' + p.kind) }; }));
    } }
  ];
  providers.forEach(function (p) {
    var when = function (v) { return v.provider === p.kind; };
    out.push({ type: 'heading', label: 'vc.provider.' + p.kind, when: when });
    out.push({ name: 'about.' + p.kind, type: 'static', readonly: true, when: when, render: function () { return notice('info', t('vc.about.' + p.kind)); } });
    p.fields.forEach(function (f) {
      var name = 'p.' + p.kind + '.' + f.name;
      var label = 'f.vc.' + p.kind + '.' + f.name;
      if (f.secret) {
        out.push({ name: name, type: 'custom', label: label, hint: values.canSeal === false ? 'vc.noSeal' : label + '.hint', optional: !f.required, when: when, build: function (cur, change) { return voiceSecretInput(cur, change, t(label), p.kind + '.' + f.name); } });
      } else {
        out.push({ name: name, type: 'text', ltr: true, label: label, hint: label + '.hint', required: !!f.required, optional: !f.required, when: when });
      }
    });
  });
  out.push({ type: 'heading', label: 'vc.rules' });
  out.push({ name: 'ttlSeconds', type: 'number', label: 'f.vc.ttl', hint: 'f.vc.ttl.hint', required: true, min: 60, max: 86400 });
  out.push({ type: 'row', fields: [
    { name: 'canPublish', type: 'bool', inline: true, label: 'f.vc.canPublish', hint: 'f.vc.canPublish.hint' },
    { name: 'openRooms', type: 'bool', inline: true, label: 'f.vc.openRooms', hint: 'f.vc.openRooms.hint' }
  ] });
  return out;
}

function voiceValues(data) {
  var v = clone(data) || {};
  v.p = {};
  (v.providers || []).forEach(function (p) {
    v.p[p.kind] = {};
    p.fields.forEach(function (f) {
      if (v.provider !== p.kind) return;
      v.p[p.kind][f.name] = f.secret ? ((v.secrets || {})[f.name] || '') : ((v.config || {})[f.name] || '');
    });
  });
  if (!v.ttlSeconds) v.ttlSeconds = 600;
  return v;
}

function voicePayload(values, raw) {
  var out = { enabled: !!raw.enabled, provider: raw.provider || '', ttlSeconds: Number(raw.ttlSeconds) || 0, canPublish: !!raw.canPublish, openRooms: !!raw.openRooms, config: {}, secrets: {} };
  var spec = (values.providers || []).filter(function (p) { return p.kind === out.provider; })[0];
  if (spec) {
    spec.fields.forEach(function (f) {
      var v = getPath(raw, 'p.' + spec.kind + '.' + f.name);
      v = v == null ? '' : String(v).trim();
      if (f.secret) out.secrets[f.name] = v;
      else out.config[f.name] = v;
    });
  }
  return out;
}

function voiceErrorField(err, provider, providers) {
  if (!err || !err.field || !provider) return err;
  var spec = (providers || []).filter(function (p) { return p.kind === provider; })[0];
  if (!spec || !spec.fields.some(function (f) { return f.name === err.field; })) return err;
  var copy = new ApiError(err.status, err.code, err.message, 'p.' + provider + '.' + err.field);
  return copy;
}

function voiceState(v) {
  if (v.canSeal === false) return notice('warning', t('vc.state.noSeal'));
  if (!v.enabled) return notice('secondary', t('vc.state.off'));
  if (!v.provider) return notice('error', t('vc.state.noProvider'));
  return notice('success', t('vc.state.on', { name: t('vc.provider.' + v.provider) }));
}

function voiceTestResult(res) {
  if (res.ok) return notice('success', t('vc.test.ok', { name: t('vc.provider.' + res.provider), ms: fmtNum(res.ms || 0) }));
  var code = res.code || 'VOICE_PROVIDER_ERROR';
  return notice('error', (has('code.' + code) ? t('code.' + code) : code) + (res.status ? ' (' + fmtNum(res.status) + ')' : ''));
}

function voicePage(container) {
  var frame = tabFrame(t('desc.voice'));
  clear(container).appendChild(frame.el);
  var box = card(h('div', { class: 'p-4 md:p-6' }, formSkeleton({ fields: voiceFields({}) })));
  box.setAttribute('data-settings-form', 'voice');
  frame.content.appendChild(box);
  var body = box.firstChild;

  function load() {
    Promise.all([baseApi('GET', '/voice/settings'), freshLanguages()]).then(function (all) {
      if (!box.isConnected) return;
      paint(all[0] || {});
    }, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }

  function paint(data) {
    var values = voiceValues(data);
    var top = h('div', {});
    var state = h('div', { class: 'mb-4', 'data-voice-state': '' }, voiceState(data));
    var result = h('div', { class: 'mt-3', 'data-voice-result': '' });
    var form = buildForm(voiceFields(values), values, {});
    var lead = form.el.firstElementChild;
    if (lead && lead.classList.contains('border-t')) lead.classList.add('eco-flush');
    var saveBtn = button({ variant: 'primary', size: 'sm', label: t('save'), onclick: function () { save(); } });
    saveBtn.setAttribute('data-settings-save', '');
    var testBtn = button({ variant: 'secondary', size: 'sm', label: t('vc.test'), onclick: function () { test(); } });
    testBtn.setAttribute('data-voice-test', '');
    clear(body).appendChild(h('form', { novalidate: true, onsubmit: function (e) { e.preventDefault(); save(); } },
      state, top, form.el, result,
      data.updatedBy ? h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-3', text: t('up.savedBy', { who: data.updatedBy, when: fmtDate(data.updatedAt) }) }) : null,
      h('div', { class: 'flex flex-wrap justify-end gap-2 pt-2' }, testBtn, saveBtn),
      h('button', { type: 'submit', hidden: true })));

    function fail(err, btn) {
      setBusy(btn, false);
      var mapped = voiceErrorField(err, values.provider, data.providers);
      if (!form.showServerError(mapped)) {
        clear(top).appendChild(h('div', { class: 'mb-5' }, notice('error', errorText(err))));
        top.scrollIntoView({ block: 'start' });
      }
    }

    function save() {
      clear(top);
      if (!form.validate()) return;
      var payload = voicePayload(data, form.payload());
      setBusy(saveBtn, true, t('saving'));
      baseApi('PUT', '/voice/settings', payload).then(function (res) {
        toast('success', t('saved'));
        paint(res || data);
      }, function (err) { fail(err, saveBtn); });
    }

    function test() {
      clear(top);
      clear(result);
      var raw = form.payload();
      var payload = raw.provider ? voicePayload(data, raw) : {};
      setBusy(testBtn, true, t('vc.testing'));
      baseApi('POST', '/voice/test', payload).then(function (res) {
        setBusy(testBtn, false);
        result.appendChild(voiceTestResult(res || {}));
      }, function (err) { fail(err, testBtn); });
    }
  }

  load();
}
