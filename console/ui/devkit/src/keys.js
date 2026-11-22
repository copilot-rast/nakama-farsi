var keyQuery = { status: '', q: '', page: 1, limit: 20 };

function keysSettings(cur, defs) {
  var nums = {};
  function num(name) {
    nums[name] = input(cur[name], { name: name, ltr: true, placeholder: String(defs[name]) });
    return field(t('set.apiKeys.' + name), nums[name]);
  }
  var boxes = {};
  function box(name) {
    boxes[name] = check(cur[name], t('set.apiKeys.' + name), name);
    return boxes[name];
  }
  var el = h('div', { class: 'grid md:grid-cols-2 gap-x-6' },
    h('div', {}, num('defaultExpiryDays'), num('maxExpiryDays'), num('lastUsedEverySeconds')),
    h('div', {}, box('requireExpiry'), box('allowWildcards'), box('catalogScopesOnly'), box('trustProxyHeaders')));
  return {
    el: el,
    read: function () {
      var out = {};
      Object.keys(nums).forEach(function (k) { var n = numberValue(nums[k]); out[k] = n == null ? defs[k] : n; });
      Object.keys(boxes).forEach(function (k) { out[k] = boxes[k].input.checked; });
      return out;
    },
    after: function () { if (keysPage.reload) keysPage.reload(); }
  };
}

function statusTag(s) {
  return tag(t('key.status.' + s), s === 'active' ? 'success' : s === 'revoked' ? 'error' : 'warning');
}

function keysPage(root) {
  var frame = pageFrame('api-keys', t('desc.api-keys'), [button({ label: t('key.new'), variant: 'primary', data: { 'devkit-action': 'new-key' }, onclick: function () { keyEditor(null); } })]);
  clear(root).appendChild(frame.el);
  var list = h('div', { 'data-devkit-list': 'api-keys' });
  var scopesBox = h('div', { 'data-devkit-list': 'api-scopes' });
  var status = select(keyQuery.status, [{ value: '', label: t('key.status') + ': ' + t('all') }, { value: 'active', label: t('key.status.active') }, { value: 'revoked', label: t('key.status.revoked') }, { value: 'expired', label: t('key.status.expired') }], { name: 'status', label: t('key.status') });
  status.addEventListener('change', function () { keyQuery.status = status.value; keyQuery.page = 1; load(); });
  var search = input(keyQuery.q, { placeholder: t('search'), name: 'q', label: t('search') });
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { keyQuery.q = search.value.trim(); keyQuery.page = 1; load(); }, 250); });
  frame.content.appendChild(settingsCard('apiKeys', keysSettings));
  frame.content.appendChild(h('div', { class: 'flex flex-wrap gap-2' }, h('div', { class: 'flex-1 min-w-[12rem]' }, search), h('div', {}, status)));
  frame.content.appendChild(list);
  frame.content.appendChild(scopesBox);

  function load() {
    api('GET', '/api-keys' + qs(keyQuery)).then(function (d) {
      if (!list.isConnected) return;
      var cols = [{ label: t('key.name') }, { label: t('key.prefix') }, { label: t('key.scopes') }, { label: t('key.status') }, { label: t('key.expires') }, { label: t('key.lastUsed') }, { label: t('key.uses') }, { label: '' }];
      var rows = d.items.map(function (k) {
        var actions = h('div', { class: 'flex gap-1 justify-end' });
        if (k.status !== 'revoked') actions.appendChild(button({ label: t('key.revoke'), variant: 'danger', data: { 'devkit-action': 'revoke-key' }, onclick: function () { revoke(k); } }));
        if (k.status !== 'active') actions.appendChild(button({ label: t('delete'), variant: 'danger', data: { 'devkit-action': 'delete-key' }, onclick: function () { remove(k); } }));
        return { id: k.id, item: k, cells: [
          h('div', {}, h('div', { class: 'font-medium', text: k.name }), k.description ? h('div', { class: 'text-xs text-secondary-500', text: k.description }) : null),
          mono(k.prefix + '_…'),
          h('div', { class: 'flex flex-wrap gap-1' }, k.scopes.map(function (s) { return tag(s, 'primary'); })),
          statusTag(k.status),
          h('span', { text: k.expiresAt ? fmtDate(k.expiresAt) : t('key.expiry.never') }),
          h('span', { text: k.lastUsedAt ? fmtDate(k.lastUsedAt) + (k.lastUsedIp ? ' · ' + k.lastUsedIp : '') : t('key.never') }),
          h('span', { text: digits(String(k.useCount)) }),
          actions] };
      });
      clear(list).appendChild(h('div', { class: 'flex flex-col gap-2' },
        rows.length ? table(cols, rows, function (k) { if (k.status !== 'revoked') keyEditor(k); }) : notice('secondary', t('empty')),
        pager(d, function (p) { keyQuery.page = p; load(); })));
    }, function (err) { clear(list).appendChild(notice('error', t('error.load') + ' ' + errorText(err))); });
    loadScopes();
  }

  function loadScopes() {
    api('GET', '/api-scopes').then(function (d) {
      keysPage.scopes = d;
      var name = input('', { name: 'scope-name', ltr: true, mono: true, placeholder: t('scope.placeholder') });
      var desc = input('', { name: 'scope-description', placeholder: t('scope.desc') });
      var msg = h('div', {});
      var addBtn = button({ label: t('scope.add'), variant: 'primary', data: { 'devkit-action': 'add-scope' }, onclick: function () {
        busy(addBtn, true);
        api('POST', '/api-scopes', { name: name.value, description: desc.value }).then(function () { busy(addBtn, false); loadScopes(); }, function (err) { busy(addBtn, false); clear(msg).appendChild(notice('error', errorText(err))); });
      } });
      var rows = d.items.map(function (s) {
        return { id: s.name, item: s, cells: [mono(s.name), h('span', { text: s.description }), tag(s.source), h('span', { text: digits(String(s.keys)) }),
          s.keys ? null : button({ label: t('delete'), variant: 'danger', data: { 'devkit-action': 'delete-scope' }, onclick: function () {
            confirmBox(s.name, t('delete')).then(function (ok) { if (ok) api('DELETE', '/api-scopes/' + encodeURIComponent(s.name)).then(loadScopes, function (err) { toast('error', errorText(err)); }); });
          } })] };
      });
      clear(scopesBox).appendChild(card(h('div', { class: 'p-4 flex flex-col gap-3' },
        h('h2', { class: 'font-semibold', text: t('scope.title') }),
        table([{ label: t('scope.name') }, { label: t('scope.desc') }, { label: t('scope.source') }, { label: t('scope.keys') }, { label: '' }], rows),
        h('div', { class: 'flex flex-wrap gap-2 items-start' }, h('div', { class: 'w-56' }, name), h('div', { class: 'flex-1 min-w-[12rem]' }, desc), addBtn),
        msg)));
    }, function (err) { clear(scopesBox).appendChild(notice('error', errorText(err))); });
  }

  function revoke(k) {
    confirmBox(t('key.revoke.confirm', { name: k.name }), t('key.revoke')).then(function (ok) {
      if (!ok) return;
      api('POST', '/api-keys/' + k.id + '/revoke').then(function () { load(); }, function (err) { toast('error', errorText(err)); });
    });
  }

  function remove(k) {
    confirmBox(t('key.delete.confirm', { name: k.name }), t('delete')).then(function (ok) {
      if (!ok) return;
      api('DELETE', '/api-keys/' + k.id).then(function () { load(); }, function (err) { toast('error', errorText(err)); });
    });
  }

  keysPage.reload = load;
  load();
}

function keyEditor(k) {
  var creating = !k;
  var cfg = dkSettings && dkSettings.settings ? dkSettings.settings.apiKeys : { defaultExpiryDays: 0, allowWildcards: true, requireExpiry: false };
  var scopes = (keysPage.scopes && keysPage.scopes.items) || [];
  var name = input(k ? k.name : '', { name: 'name' });
  var desc = input(k ? k.description : '', { name: 'description' });
  var picked = {};
  (k ? k.scopes : []).forEach(function (s) { picked[s] = true; });
  var checks = scopes.map(function (s) {
    var c = check(picked[s.name], s.name + (s.description ? ' — ' + s.description : ''), 'scope:' + s.name);
    c.input.setAttribute('data-devkit-scope', s.name);
    c.input.addEventListener('change', function () { picked[s.name] = c.input.checked; });
    return c;
  });
  var extra = input((k ? k.scopes : []).filter(function (s) { return !scopes.some(function (x) { return x.name === s; }); }).join(' '), { name: 'extraScopes', ltr: true, mono: true, placeholder: 'assets:*' });
  var ips = input(k ? k.allowedIps.join('\n') : '', { name: 'allowedIps', ltr: true, mono: true, multiline: true, rows: 2, placeholder: '203.0.113.7\n10.0.0.0/8' });
  var opts = [];
  if (!creating) opts.push({ value: 'keep', label: t('key.expiry.keep') + (k.expiresAt ? ' (' + fmtDate(k.expiresAt) + ')' : '') });
  if (creating && cfg.defaultExpiryDays > 0) opts.push({ value: 'default', label: t('key.expiry.default', { days: cfg.defaultExpiryDays }) });
  if (!cfg.requireExpiry && !(cfg.maxExpiryDays > 0)) opts.push({ value: 'never', label: t('key.expiry.never') });
  opts.push({ value: 'days', label: t('key.expiry.days') });
  opts.push({ value: 'date', label: t('key.expiry.date') });
  var mode = select(opts[0].value, opts, { name: 'expiryMode' });
  var days = input(30, { name: 'expiresInDays', ltr: true });
  var date = window.NakamaDate ? window.NakamaDate.input({ mode: 'datetime', name: 'expiresAt', label: t('key.date'), inputClass: DK_INPUT, min: new Date().toISOString() }) : input('', { name: 'expiresAt', type: 'datetime-local', ltr: true });
  var daysField = field(t('key.days'), days);
  var dateField = field(t('key.date'), date);
  function sync() { daysField.hidden = mode.value !== 'days'; dateField.hidden = mode.value !== 'date'; }
  mode.addEventListener('change', sync);
  sync();
  var f = { name: field(t('key.name'), name), desc: field(t('key.description') + ' (' + t('optional') + ')', desc), scopes: field(t('key.scopes'), h('div', { class: 'dk-scope-list' }, checks), t('key.scopes.hint')),
    extra: field(t('key.custom'), extra), ips: field(t('key.ips'), ips, t('key.ips.hint')), expiry: field(t('key.expiry'), mode) };
  var msg = h('div', {});
  var body = h('div', {}, f.name, f.desc, f.scopes, cfg.allowWildcards || !cfg.catalogScopesOnly ? f.extra : null, f.ips, f.expiry, daysField, dateField, msg);
  var save = button({ label: t('save'), variant: 'primary', data: { 'devkit-action': 'save-key' }, onclick: submit });
  sidePanel(creating ? t('key.newTitle') : t('key.editTitle', { name: k.name }), body, [button({ label: t('cancel'), onclick: closePanel }), save], 'key');

  function submit() {
    var list = Object.keys(picked).filter(function (s) { return picked[s]; });
    extra.value.split(/[\s,]+/).forEach(function (s) { if (s) list.push(s); });
    var payload = { name: name.value, description: desc.value, scopes: list, allowedIps: ips.value.split(/\n+/).map(function (s) { return s.trim(); }).filter(Boolean) };
    if (mode.value === 'never') payload.neverExpires = true;
    if (mode.value === 'days') payload.expiresInDays = numberValue(days) || 0;
    if (mode.value === 'date' && date.value) payload.expiresAt = new Date(date.value).toISOString();
    busy(save, true);
    clear(msg);
    Object.keys(f).forEach(function (x) { f[x].setError(''); });
    var p = creating ? api('POST', '/api-keys', payload) : api('PUT', '/api-keys/' + k.id, payload);
    p.then(function (res) {
      closePanel();
      if (creating) showKey(res.key, res.item);
      else toast('success', t('saved'));
      if (keysPage.reload) keysPage.reload();
    }, function (err) {
      busy(save, false);
      var map = { name: f.name, scopes: f.scopes, allowedIps: f.ips, expiresAt: f.expiry };
      if (map[err.field]) map[err.field].setError(errorText(err));
      else msg.appendChild(notice('error', errorText(err)));
    });
  }
}

function showKey(full, item) {
  var box = input(full, { name: 'created-key', ltr: true, mono: true });
  box.readOnly = true;
  box.setAttribute('data-devkit-created-key', '');
  var copy = button({ label: t('key.copy'), onclick: function () {
    box.select();
    try { navigator.clipboard.writeText(full).then(function () { toast('success', t('key.copied')); }); } catch (e) { document.execCommand('copy'); }
  } });
  sidePanel(t('key.shown.title') + ' — ' + item.name, h('div', { class: 'flex flex-col gap-3' }, notice('warning', t('key.shown.warn')), box, h('div', {}, copy)), [button({ label: t('key.done'), variant: 'primary', data: { 'devkit-action': 'key-done' }, onclick: closePanel })], 'key-created');
}
