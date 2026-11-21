var errorQuery = { q: '', module: '', state: '', missing: '', page: 1, limit: 25 };

function errorsSettings(cur, defs) {
  var langs = (errorsPage.languages || []).map(function (l) { return { value: l.code, label: l.name + ' (' + l.code + ')' }; });
  var mode = select(cur.mode, [{ value: 'envelope', label: t('set.errors.mode.envelope') }, { value: 'code', label: t('set.errors.mode.code') }], { name: 'mode' });
  var moduleMode = select(cur.moduleMode, [{ value: 'code', label: t('set.errors.moduleMode.code') }, { value: 'envelope', label: t('set.errors.moduleMode.envelope') }], { name: 'moduleMode' });
  var fallback = select(cur.fallbackLang, [{ value: '', label: t('set.errors.fallbackLang.auto') }].concat(langs), { name: 'fallbackLang' });
  var prefer = check(cur.preferAccountLanguage, t('set.errors.preferAccountLanguage'), 'preferAccountLanguage');
  var digitsBox = check(cur.localizeDigits, t('set.errors.localizeDigits'), 'localizeDigits');
  var pub = check(cur.publicCatalog, t('set.errors.publicCatalog'), 'publicCatalog');
  var detail = check(cur.includeDetail, t('set.errors.includeDetail'), 'includeDetail');
  var cache = input(cur.cacheSeconds, { name: 'cacheSeconds', ltr: true, placeholder: String(defs.cacheSeconds) });
  var el = h('div', { class: 'grid md:grid-cols-2 gap-x-6' },
    field(t('set.errors.mode'), mode),
    field(t('set.errors.moduleMode'), moduleMode, t('set.errors.moduleMode.hint')),
    field(t('set.errors.fallbackLang'), fallback, t('set.errors.fallbackLang.hint')),
    field(t('set.errors.cacheSeconds'), cache),
    h('div', {}, prefer, digitsBox, pub, detail));
  return {
    el: el,
    read: function () {
      var n = numberValue(cache);
      return { mode: mode.value, moduleMode: moduleMode.value, fallbackLang: fallback.value, preferAccountLanguage: prefer.input.checked, localizeDigits: digitsBox.input.checked, publicCatalog: pub.input.checked, includeDetail: detail.input.checked, cacheSeconds: n == null ? defs.cacheSeconds : n };
    }
  };
}

function errorsPage(root, route) {
  var frame = pageFrame('errors', t('desc.errors'), [button({ label: t('err.new'), variant: 'primary', data: { 'devkit-action': 'new-error' }, onclick: function () { errorEditor(null); } })]);
  clear(root).appendChild(frame.el);
  var settingsBox = settingsCard('errors', errorsSettings);
  var bar = h('div', { class: 'flex flex-wrap gap-2 items-end' });
  var list = h('div', { 'data-devkit-list': 'errors' });
  frame.content.appendChild(settingsBox);
  frame.content.appendChild(bar);
  frame.content.appendChild(list);
  var searchTimer = null;

  function toolbar() {
    var search = input(errorQuery.q, { placeholder: t('search'), name: 'q', label: t('search') });
    search.addEventListener('input', function () {
      clearTimeout(searchTimer);
      searchTimer = setTimeout(function () { errorQuery.q = search.value.trim(); errorQuery.page = 1; load(); }, 250);
    });
    var mods = [{ value: '', label: t('err.module') + ': ' + t('all') }].concat((errorsPage.modules || []).map(function (m) { return { value: m, label: m }; })).concat([{ value: 'custom', label: t('err.state.custom') }]);
    var mod = select(errorQuery.module, mods, { name: 'module', label: t('err.module') });
    mod.addEventListener('change', function () { errorQuery.module = mod.value; errorQuery.page = 1; load(); });
    var state = select(errorQuery.state, [{ value: '', label: t('err.state') + ': ' + t('all') }, { value: 'edited', label: t('err.state.edited') }, { value: 'custom', label: t('err.state.custom') }, { value: 'default', label: t('err.state.default') }], { name: 'state', label: t('err.state') });
    state.addEventListener('change', function () { errorQuery.state = state.value; errorQuery.page = 1; load(); });
    var miss = select(errorQuery.missing, [{ value: '', label: t('err.missing') + ': -' }].concat((errorsPage.languages || []).map(function (l) { return { value: l.code, label: t('err.missing') + ' ' + l.name }; })), { name: 'missing', label: t('err.missing') });
    miss.addEventListener('change', function () { errorQuery.missing = miss.value; errorQuery.page = 1; load(); });
    clear(bar).appendChild(h('div', { class: 'flex flex-wrap gap-2 w-full' },
      h('div', { class: 'flex-1 min-w-[12rem]' }, search), h('div', {}, mod), h('div', {}, state), h('div', {}, miss)));
  }

  function load() {
    api('GET', '/errors' + qs(errorQuery)).then(function (d) {
      if (!list.isConnected) return;
      var firstPaint = !errorsPage.languages;
      errorsPage.languages = d.languages;
      errorsPage.modules = d.modules;
      if (firstPaint) toolbar();
      paint(d);
    }, function (err) {
      clear(list).appendChild(notice('error', t('error.load') + ' ' + errorText(err)));
    });
  }

  function paint(d) {
    var langs = d.languages || [];
    var cols = [{ label: t('err.code') }, { label: t('err.status') }].concat(langs.map(function (l) { return { label: l.name }; })).concat([{ label: t('err.module') }]);
    var rows = d.items.map(function (e) {
      var cells = [h('div', { class: 'flex flex-col gap-1' }, mono(e.code), e.custom ? tag(t('err.state.custom'), 'primary') : e.edited ? tag(t('err.state.edited'), 'warning') : null),
        mono(e.status)];
      langs.forEach(function (l) {
        var txt = e.messages[l.code] || '';
        cells.push(h('span', { class: txt ? '' : 'text-secondary-400', dir: l.direction || 'auto', 'data-devkit-text': l.code, text: txt || '—' }));
      });
      cells.push(h('div', { class: 'flex flex-wrap gap-1' }, (e.modules || []).map(function (m) { return tag(m); })));
      return { id: e.code, item: e, cells: cells };
    });
    clear(list).appendChild(h('div', { class: 'flex flex-col gap-2' },
      rows.length ? table(cols, rows, function (e) { errorEditor(e); }) : notice('secondary', t('empty')),
      pager(d, function (p) { errorQuery.page = p; load(); })));
  }

  errorsPage.reload = load;
  if (errorsPage.languages) toolbar();
  load();
}

function errorEditor(entry) {
  var creating = !entry;
  var langs = errorsPage.languages || [];
  var codeInput = creating ? input('', { name: 'code', ltr: true, mono: true, placeholder: 'MY_ERROR_CODE' }) : null;
  var texts = {};
  var fields = [];
  var codeField = creating ? field(t('err.code'), codeInput) : null;
  if (codeField) fields.push(codeField);
  langs.forEach(function (l) {
    var def = entry && entry.defaultMessages ? entry.defaultMessages[l.code] || '' : '';
    var cur = entry ? (entry.messages[l.code] || '') : '';
    var shown = cur === def ? '' : cur;
    texts[l.code] = input(shown, { multiline: true, rows: 2, name: 'message-' + l.code, placeholder: def });
    texts[l.code].setAttribute('dir', l.direction || 'auto');
    fields.push(texts[l.code].wrap = field(t('err.message', { lang: l.name }), texts[l.code], def ? t('err.defaultText', { text: def }) : (creating ? null : t('err.noDefault'))));
  });
  var status = input(entry && entry.status !== entry.defaultStatus ? entry.status : '', { name: 'status', ltr: true, placeholder: entry ? String(entry.defaultStatus || entry.status) : '400' });
  var retry = input(entry && entry.retryAfterSeconds != null ? entry.retryAfterSeconds : '', { name: 'retryAfterSeconds', ltr: true });
  var statusField = field(t('err.status'), status, entry ? t('err.status.hint', { status: entry.defaultStatus || entry.status }) : null);
  var retryField = field(t('err.retry'), retry, t('err.retry.hint'));
  var msg = h('div', {});
  var body = h('div', {}, fields, h('div', { class: 'grid grid-cols-2 gap-4' }, statusField, retryField), h('p', { class: 'text-xs text-secondary-500 mb-3', text: t('err.placeholders') }), msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'devkit-action': 'save-error' }, onclick: submit });
  var footer = [];
  if (entry && entry.edited && !entry.custom) footer.push(button({ label: t('reset'), variant: 'danger', data: { 'devkit-action': 'reset-error' }, onclick: reset }));
  if (entry && entry.custom) footer.push(button({ label: t('delete'), variant: 'danger', data: { 'devkit-action': 'delete-error' }, onclick: reset }));
  footer.push(button({ label: t('cancel'), onclick: closePanel }));
  footer.push(save);
  sidePanel(creating ? t('err.newTitle') : t('err.editTitle', { code: entry.code }), body, footer, 'error');

  function fail(err) {
    busy(save, false);
    var f = err.field || '';
    if (f === 'code' && codeField) { codeField.setError(errorText(err)); return; }
    if (f.indexOf('messages.') === 0 && texts[f.slice(9)]) { texts[f.slice(9)].wrap.setError(errorText(err)); return; }
    if (f === 'status') { statusField.setError(errorText(err)); return; }
    if (f === 'retryAfterSeconds') { retryField.setError(errorText(err)); return; }
    clear(msg).appendChild(notice('error', errorText(err)));
  }

  function submit() {
    var messages = {};
    Object.keys(texts).forEach(function (k) { messages[k] = texts[k].value; });
    var st = numberValue(status);
    var rt = numberValue(retry);
    var payload = { messages: messages, status: st, retryAfterSeconds: rt };
    busy(save, true);
    var p = creating ? api('POST', '/errors', Object.assign({ code: codeInput.value.trim() }, payload)) : api('PUT', '/errors/' + encodeURIComponent(entry.code), payload);
    p.then(function () {
      closePanel();
      toast('success', t('saved'));
      if (errorsPage.reload) errorsPage.reload();
    }, fail);
  }

  function reset() {
    var ask = entry.custom ? confirmBox(t('err.deleteCustom'), t('delete')) : Promise.resolve(true);
    ask.then(function (ok) {
      if (!ok) return;
      api('DELETE', '/errors/' + encodeURIComponent(entry.code)).then(function () {
        closePanel();
        toast('success', entry.custom ? t('saved') : t('err.resetDone'));
        if (errorsPage.reload) errorsPage.reload();
      }, fail);
    });
  }
}
