var APP_PLATFORMS = ['android', 'ios', 'web', 'windows', 'macos', 'linux'];
var APP_STORES = ['google_play', 'app_store', 'bazaar', 'myket', 'microsoft_store', 'steam', 'web', 'direct'];
var VERSION_RE = /^v?\d{1,9}(\.\d{1,9}){0,3}([-+][0-9A-Za-z.-]{1,40})?$/;
var STORE_URL_RE = /^(https?|market|bazaar|myket|itms-apps):\/\/\S+$/i;
var COUNTRY_RE = /^[A-Za-z]{2}$/;
var OTP_STATUS_COLOR = { pending: 'primary', verified: 'success', expired: 'secondary', failed: 'error', replaced: 'secondary', send_failed: 'error' };

function flagOf(code) {
  code = String(code || '').toUpperCase();
  if (!/^[A-Z]{2}$/.test(code)) return '';
  return String.fromCodePoint(code.charCodeAt(0) - 65 + 0x1F1E6, code.charCodeAt(1) - 65 + 0x1F1E6);
}

function flagBox(it) {
  var box = h('div', { class: 'eco-flag', 'aria-hidden': 'true' });
  if (it.flagUrl) box.appendChild(img(it.flagUrl, 'flag'));
  else box.appendChild(h('span', { text: it.flag || flagOf(it.code) }));
  return box;
}

function countryCell(it) {
  return h('div', { class: 'flex items-center gap-3 min-w-0' }, flagBox(it),
    h('div', { class: 'min-w-0' },
      h('div', { class: 'font-medium truncate', text: i18nText(it.name) || it.code }),
      h('div', { class: 'text-secondary-500' }, mono(it.code))));
}

function countryBody(it, patch) {
  return Object.assign({ name: trClean(it.name || {}), dialCode: it.dialCode || '', flagUrl: it.flagUrl || '', isActive: !!it.isActive, sort: it.sort == null ? 100 : it.sort }, patch || {});
}

function countryActive(it) {
  var box = h('div', { class: 'flex items-center gap-2', 'data-country-toggle': it.code });
  var label = h('span', { class: 'text-xs text-secondary-600 dark:text-secondary-300', text: it.isActive ? t('active') : t('inactive') });
  var sw = toggle(it.isActive, null, function (on) {
    baseApi('PUT', '/countries/' + encodeURIComponent(it.code), countryBody(it, { isActive: on })).then(function () {
      it.isActive = on;
      label.textContent = on ? t('active') : t('inactive');
      toast('success', t(on ? 'ct.activated' : 'ct.deactivated', { name: i18nText(it.name) || it.code }));
    }, function (err) {
      sw.set(!on);
      toast('error', errorText(err));
    });
  }, 'sm');
  sw.setAttribute('aria-label', t('f.isActive') + ': ' + (i18nText(it.name) || it.code));
  box.appendChild(sw);
  box.appendChild(label);
  return box;
}

function providerName(kind) {
  return has('pv.kind.' + kind) ? t('pv.kind.' + kind) : kind;
}

function platformLabel(p) {
  return has('platform.' + p) ? t('platform.' + p) : p;
}

function storeLabel(s) {
  return has('store.' + s) ? t('store.' + s) : s;
}

function versionCell(v) {
  return v ? h('span', { class: 'font-mono text-sm', dir: 'ltr', text: v }) : h('span', { class: 'text-secondary-400', text: t('none') });
}

function versionCheck(v) {
  return v && !VERSION_RE.test(String(v).trim()) ? t('code.INVALID_VERSION') : null;
}

function versionParts(v) {
  var s = String(v || '').trim().replace(/^v/, '');
  var pre = '';
  var m = /^([0-9.]+)(?:-([0-9A-Za-z.-]+))?/.exec(s);
  if (!m) return null;
  pre = m[2] || '';
  return { nums: m[1].split('.').map(Number), pre: pre };
}

function compareVersions(a, b) {
  var x = versionParts(a);
  var y = versionParts(b);
  if (!x || !y) return 0;
  for (var i = 0; i < 4; i++) {
    var p = x.nums[i] || 0;
    var q = y.nums[i] || 0;
    if (p !== q) return p < q ? -1 : 1;
  }
  if (x.pre === y.pre) return 0;
  if (!x.pre) return 1;
  if (!y.pre) return -1;
  return x.pre < y.pre ? -1 : 1;
}

function storeLinksEditor(value, onchange) {
  var rows = (value || []).map(function (r) { return { store: r.store || 'web', url: r.url || '', label: trValue(r.label) }; });
  var listEl = h('div', { class: 'flex flex-col gap-2' });
  var errors = [];
  function emit() { onchange(rows.map(function (r) { return { store: r.store, url: r.url, label: Object.assign({}, r.label) }; })); }
  var add = button({ variant: 'link', size: 'sm', icon: 'plus', label: t('av.link.add'), onclick: function () {
    var used = rows.map(function (r) { return r.store; });
    var next = APP_STORES.filter(function (s) { return used.indexOf(s) === -1; })[0] || 'web';
    rows.push({ store: next, url: '', label: {} });
    render();
    emit();
    var inputs = listEl.querySelectorAll('[data-link-url]');
    if (inputs.length) inputs[inputs.length - 1].focus();
  } });
  function render() {
    clear(listEl);
    errors = [];
    add.hidden = rows.length >= 8;
    if (!rows.length) listEl.appendChild(h('div', { class: 'text-sm text-secondary-500 dark:text-secondary-400 border border-dashed border-secondary-300 dark:border-secondary-700 rounded-md px-3 py-4 text-center', text: t('av.link.empty') }));
    rows.forEach(function (row, idx) {
      var cell = function (labelKey, ctl) {
        return h('div', { class: 'min-w-0' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t(labelKey) }), ctl);
      };
      var storeSel = selectInput(row.store, APP_STORES.map(function (s) { return { value: s, label: storeLabel(s) }; }), function (v) { row.store = v; emit(); }, { placeholder: false, label: t('av.link.store') });
      var url = textInput(row.url, { ltr: true, mono: true, label: t('av.link.url'), placeholder: 'https://', oninput: function (v) { row.url = v.trim(); emit(); } });
      url.setAttribute('data-link-url', '');
      var label = translatedInput(row.label, { label: t('av.link.label'), oninput: function (v) { row.label = v; emit(); } });
      var err = h('p', { class: 'text-xs text-error-600 dark:text-error-400 eco-items-err', hidden: true });
      errors.push({ row: row, el: err, label: label });
      listEl.appendChild(h('div', { class: 'border border-secondary-200 dark:border-secondary-700 rounded-md p-3', 'data-store-row': '' },
        h('div', { class: 'eco-actions-row' },
          cell('av.link.store', storeSel),
          cell('av.link.url', url),
          h('div', { class: 'min-w-0 eco-actions-label' }, h('div', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-1', text: t('av.link.label') }), label),
          h('div', { class: 'eco-items-remove' }, button({ variant: 'danger', size: 'iconLg', icon: 'trash', title: t('remove'), onclick: function () { rows.splice(idx, 1); render(); emit(); } }))),
        h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mt-1.5', text: t('av.link.hint') }),
        err));
    });
  }
  render();
  var el = h('div', {}, listEl, h('div', { class: 'mt-2' }, add));
  el.validate = function () {
    var bad = false;
    var seen = {};
    errors.forEach(function (e) {
      var r = e.row;
      var msg = '';
      var hit = trCheck(r.label, { max: 64 });
      if (!r.url) msg = t('av.link.url') + ': ' + t('err.required');
      else if (!STORE_URL_RE.test(r.url)) msg = t('code.INVALID_URL');
      else if (seen[r.store]) msg = t('code.DUPLICATE_STORE');
      else if (hit) msg = t('av.link.label') + ': ' + t('code.' + hit.code, { lang: langName(hit.lang) });
      seen[r.store] = true;
      e.el.textContent = msg;
      e.el.hidden = !msg;
      e.el.parentElement.classList.toggle('eco-row-invalid', !!msg);
      e.label.markLang(hit ? hit.lang : null);
      if (msg) bad = true;
    });
    return bad ? t('err.form') : null;
  };
  return el;
}

RESOURCES.countries = {
  api: 'base',
  filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
  columns: [
    { label: 'col.country', render: countryCell, skeleton: '60%' },
    { label: 'col.dialCode', render: function (it) { return it.dialCode ? h('span', { class: 'font-mono text-sm tabular-nums', dir: 'ltr', text: it.dialCode }) : h('span', { class: 'text-secondary-400', text: t('none') }); }, width: 130 },
    { label: 'col.status', render: countryActive, width: 150 },
    { label: 'col.sort', render: function (it) { return numCell(it.sort); }, width: 90 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'code', type: 'key', label: 'f.ct.code', hint: 'f.ct.code.hint', required: true, keyRe: COUNTRY_RE, keyErr: 'code.INVALID_COUNTRY_CODE', lockOnEdit: true },
      { name: 'dialCode', type: 'text', ltr: true, label: 'f.ct.dialCode', hint: 'f.ct.dialCode.hint', optional: true, check: function (v) { return v && !/^\+?[1-9][0-9]{0,4}$/.test(String(v).trim()) ? t('code.INVALID_DIAL_CODE') : null; } }
    ] },
    tr('name', { label: 'f.ct.name', required: true, max: 100 }),
    { name: 'flagUrl', type: 'image', label: 'f.ct.flag', hint: 'f.ct.flag.hint', optional: true },
    { type: 'row', fields: [
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.ct.active.hint' },
      { name: 'sort', type: 'number', label: 'f.sort', hint: 'f.ct.sort.hint' }
    ] }
  ],
  defaults: function () { return { code: '', name: {}, dialCode: '', flagUrl: '', isActive: true, sort: 100 }; },
  fixPayload: function (p) {
    p.code = String(p.code || '').trim().toUpperCase();
    p.dialCode = String(p.dialCode || '').trim();
    if (p.dialCode && p.dialCode[0] !== '+') p.dialCode = '+' + p.dialCode;
    if (!p.flagUrl) p.flagUrl = '';
    if (p.sort == null) p.sort = 100;
    return p;
  }
};

RESOURCES['app-versions'] = {
  api: 'base',
  filters: [{ name: 'isActive', label: 'filter.active', options: function () { return [{ value: 'true', label: t('active') }, { value: 'false', label: t('inactive') }]; } }],
  columns: [
    { label: 'col.platform', render: function (it) { return h('div', { class: 'min-w-0' }, h('div', { class: 'font-medium', text: platformLabel(it.platform) }), h('div', { class: 'text-secondary-500' }, mono(it.platform))); }, skeleton: '45%' },
    { label: 'col.latest', render: function (it) { return versionCell(it.latestVersion); }, width: 130 },
    { label: 'col.minimum', render: function (it) { return versionCell(it.minVersion); }, width: 130 },
    { label: 'col.stores', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, (it.storeLinks || []).map(function (l) { return tag(i18nText(l.label) || storeLabel(l.store), 'secondary'); })); }, skeleton: '40%' },
    { label: 'col.status', render: function (it) { return activeTag(it.isActive); }, width: 110 }
  ],
  fields: [
    { type: 'row', fields: [
      { name: 'platform', type: 'select', label: 'f.av.platform', required: true, options: function () { return APP_PLATFORMS.map(function (p) { return { value: p, label: platformLabel(p) }; }); } },
      { name: 'isActive', type: 'bool', inline: true, label: 'f.isActive', hint: 'f.av.active.hint' }
    ] },
    { type: 'row', fields: [
      { name: 'latestVersion', type: 'text', ltr: true, label: 'f.av.latest', hint: 'f.av.latest.hint', required: true, check: versionCheck },
      { name: 'minVersion', type: 'text', ltr: true, label: 'f.av.min', hint: 'f.av.min.hint', optional: true, check: function (v, vals) {
        var bad = versionCheck(v);
        if (bad) return bad;
        if (v && vals.latestVersion && VERSION_RE.test(vals.latestVersion) && compareVersions(v, vals.latestVersion) > 0) return t('code.MIN_ABOVE_LATEST');
        return null;
      } }
    ] },
    { type: 'heading', label: 'f.av.stores' },
    { name: 'storeLinks', type: 'links', empty: [] },
    { type: 'heading', label: 'f.av.messages' },
    tr('updateMessage', { label: 'f.av.updateMessage', hint: 'f.av.updateMessage.hint', multiline: true, rows: 2, max: 500, optional: true }),
    tr('forceMessage', { label: 'f.av.forceMessage', hint: 'f.av.forceMessage.hint', multiline: true, rows: 2, max: 500, optional: true })
  ],
  defaults: function () { return { platform: 'android', latestVersion: '1.0.0', minVersion: '', storeLinks: [], updateMessage: {}, forceMessage: {}, isActive: true }; },
  fixPayload: function (p) {
    p.latestVersion = String(p.latestVersion || '').trim();
    p.minVersion = String(p.minVersion || '').trim();
    p.storeLinks = p.storeLinks || [];
    return p;
  }
};

LISTS.otp = {
  api: 'publication',
  noRow: true,
  filters: [{ name: 'status', label: 'filter.status', options: function () { return ['pending', 'verified', 'expired', 'failed', 'replaced', 'send_failed'].map(function (s) { return { value: s, label: t('otp.status.' + s) }; }); } }],
  columns: [
    { label: 'col.phone', render: function (it) { return h('span', { class: 'font-mono text-sm tabular-nums', dir: 'ltr', text: it.phone }); }, skeleton: '40%' },
    { label: 'col.status', render: function (it) { return h('div', { class: 'flex flex-wrap gap-1' }, tag(t('otp.status.' + it.status), OTP_STATUS_COLOR[it.status]), it.sandbox ? tag(t('otp.sandbox'), 'warning') : tag(providerName(it.provider), 'secondary')); }, width: 200 },
    { label: 'col.attempts', render: function (it) { return h('span', { class: 'tabular-nums whitespace-nowrap', text: fmtNum(it.attempts || 0) + ' / ' + fmtNum(it.maxAttempts || 0) }); }, width: 110 },
    { label: 'col.player', render: function (it) { return it.userId ? h('a', { href: '#/players/' + encodeURIComponent(it.userId), class: 'eco-link truncate', text: it.username || it.userId }) : h('span', { class: 'text-secondary-400', text: t('none') }); }, skeleton: '30%' },
    { label: 'col.created', render: function (it) { return dateCell(it.createdAt); }, width: 180 }
  ]
};

function apiPage(call, path, title, fields, opts) {
  opts = opts || {};
  var box = card(h('div', { class: 'p-4 md:p-6' }, formSkeleton({ fields: typeof fields === 'function' ? fields({}) : fields })));
  box.setAttribute('data-settings-form', opts.id || '');
  var body = box.firstChild;
  var api2 = {};
  function load() {
    Promise.all([call('GET', path), freshLanguages()]).then(function (all) {
      if (!box.isConnected) return;
      paint(all[0] || {});
    }, function (err) { clear(body).appendChild(errorBlock(err, load)); });
  }
  function paint(data) {
    var values = clone(data) || {};
    var top = h('div', {});
    var extra = h('div', {});
    var form = buildForm(typeof fields === 'function' ? fields(values) : fields, values, { onchange: function (v) { if (opts.onchange) opts.onchange(v, extra); } });
    var lead = form.el.firstElementChild;
    if (lead && lead.classList.contains('border-t')) lead.classList.add('eco-flush');
    var saveBtn = button({ variant: 'primary', size: 'sm', label: t('save'), onclick: function () { save(); } });
    saveBtn.setAttribute('data-settings-save', '');
    clear(body).appendChild(h('form', { novalidate: true, onsubmit: function (e) { e.preventDefault(); save(); } },
      title ? h('h2', { class: 'text-lg font-medium mb-1', text: t(title) }) : null,
      opts.desc ? h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mb-4', text: t(opts.desc) }) : null,
      top, form.el, extra,
      h('div', { class: 'flex justify-end pt-2' }, saveBtn),
      h('button', { type: 'submit', hidden: true })));
    if (opts.onchange) opts.onchange(values, extra);
    function save() {
      clear(top);
      if (!form.validate()) return;
      var payload = form.payload();
      if (opts.fixPayload) payload = opts.fixPayload(payload);
      setBusy(saveBtn, true, t('saving'));
      call('PUT', path, payload).then(function (res) {
        toast('success', t('saved'));
        paint(res || payload);
        if (opts.afterSave) opts.afterSave(res);
      }, function (err) {
        setBusy(saveBtn, false);
        if (!form.showServerError(err)) {
          clear(top).appendChild(h('div', { class: 'mb-5' }, notice('error', errorText(err))));
          top.scrollIntoView({ block: 'start' });
        }
      });
    }
  }
  api2.el = box;
  api2.load = load;
  return api2;
}

var RULE_FIELDS = [
  { type: 'heading', label: 'ar.username' },
  { type: 'row', fields: [
    { name: 'usernameMin', type: 'number', label: 'f.ar.usernameMin', required: true, min: 1, max: 128 },
    { name: 'usernameMax', type: 'number', label: 'f.ar.usernameMax', required: true, min: 1, max: 128, check: function (v, vals) { return v != null && vals.usernameMin != null && Number(v) < Number(vals.usernameMin) ? t('code.MAX_BELOW_MIN') : null; } }
  ] },
  { name: 'usernameCharset', type: 'select', label: 'f.ar.charset', hint: 'f.ar.charset.hint', required: true, options: function () { return [{ value: 'any', label: t('ar.charset.any') }, { value: 'latin', label: t('ar.charset.latin') }]; } },
  { type: 'heading', label: 'ar.password' },
  { name: 'passwordMin', type: 'number', label: 'f.ar.passwordMin', hint: 'f.ar.passwordMin.hint', required: true, min: 8, max: 128 },
  { type: 'row', fields: [
    { name: 'passwordLetter', type: 'bool', inline: true, label: 'f.ar.passwordLetter' },
    { name: 'passwordDigit', type: 'bool', inline: true, label: 'f.ar.passwordDigit' }
  ] },
  { type: 'heading', label: 'ar.display' },
  { name: 'displayNameMax', type: 'number', label: 'f.ar.displayNameMax', required: true, min: 1, max: 255 }
];

function ruleProblems(r, username, password) {
  var out = [];
  var n = Array.from(username || '').length;
  if (username) {
    if (n < r.usernameMin) out.push(t('code.USERNAME_TOO_SHORT', { n: r.usernameMin }));
    else if (n > r.usernameMax) out.push(t('code.USERNAME_TOO_LONG', { n: r.usernameMax }));
    else if (r.usernameCharset === 'latin' && !/^[A-Za-z0-9_.-]+$/.test(username)) out.push(t('code.USERNAME_INVALID_CHARACTERS'));
  }
  if (password) {
    if (Array.from(password).length < r.passwordMin) out.push(t('code.PASSWORD_TOO_SHORT', { n: r.passwordMin }));
    else if (r.passwordLetter && !/\p{L}/u.test(password)) out.push(t('code.PASSWORD_NEEDS_LETTER'));
    else if (r.passwordDigit && !/\p{Nd}/u.test(password)) out.push(t('code.PASSWORD_NEEDS_DIGIT'));
  }
  return out;
}

function rulesTry(extra, values) {
  if (!extra.__try) {
    var result = h('div', { class: 'flex flex-col gap-1 text-sm', 'aria-live': 'polite', 'data-rules-result': '' });
    var user = textInput('', { label: t('ar.try.username'), placeholder: t('ar.try.username') });
    var pass = textInput('', { ltr: true, label: t('ar.try.password'), placeholder: t('ar.try.password') });
    user.setAttribute('data-rules-try', 'username');
    pass.setAttribute('data-rules-try', 'password');
    extra.__try = { result: result, user: user, pass: pass };
    var paintTry = function () { rulesPaint(extra); };
    user.addEventListener('input', paintTry);
    pass.addEventListener('input', paintTry);
    extra.appendChild(h('div', { class: 'eco-try mt-2 mb-4 rounded-md border border-secondary-200 dark:border-secondary-700 p-4' },
      h('div', { class: 'text-sm font-semibold text-secondary-800 dark:text-secondary-200 mb-1', text: t('ar.try') }),
      h('p', { class: 'text-xs text-secondary-500 dark:text-secondary-400 mb-3', text: t('ar.try.hint') }),
      h('div', { class: 'eco-form-row' }, field(t('ar.try.username'), user), field(t('ar.try.password'), pass)),
      result));
  }
  extra.__rules = values;
  rulesPaint(extra);
}

function rulesPaint(extra) {
  var tr2 = extra.__try;
  var v = extra.__rules || {};
  var r = { usernameMin: Number(toLatinDigits(String(v.usernameMin == null ? 1 : v.usernameMin))) || 1, usernameMax: Number(toLatinDigits(String(v.usernameMax == null ? 128 : v.usernameMax))) || 128,
    usernameCharset: v.usernameCharset || 'any', passwordMin: Number(toLatinDigits(String(v.passwordMin == null ? 8 : v.passwordMin))) || 8, passwordLetter: !!v.passwordLetter, passwordDigit: !!v.passwordDigit };
  clear(tr2.result);
  if (!tr2.user.value && !tr2.pass.value) return;
  var problems = ruleProblems(r, tr2.user.value.trim(), tr2.pass.value);
  if (!problems.length) tr2.result.appendChild(h('span', { class: 'eco-text-success inline-flex items-center gap-1.5', 'data-rules-ok': '' }, icon('check', 'h-4 w-4'), h('span', { text: t('ar.try.ok') })));
  problems.forEach(function (p) { tr2.result.appendChild(h('span', { class: 'eco-text-error inline-flex items-center gap-1.5', 'data-rules-bad': '' }, icon('alert', 'h-4 w-4'), h('span', { text: p }))); });
}

function accountRulesPage(container) {
  var frame = tabFrame(t('desc.accounts'));
  clear(container).appendChild(frame.el);
  var page = apiPage(baseApi, '/account-rules', null, RULE_FIELDS, { id: 'accounts', onchange: function (v, extra) { rulesTry(extra, v); } });
  frame.content.appendChild(page.el);
  page.load();
}

function otpModeNotice(v, providers) {
  var list = providers || [];
  var chosen = v.providerId ? list.filter(function (p) { return p.id === v.providerId; })[0] : list.filter(function (p) { return p.isActive; })[0];
  if (!v.isEnabled) return notice('secondary', t('otp.mode.off'));
  if (v.sandbox || (chosen && chosen.sandbox)) return notice('warning', t('otp.mode.sandbox'));
  if (!chosen || !chosen.isActive) return notice('error', t('otp.mode.noProvider'), h('a', { href: hashFor('sms'), class: 'eco-link text-sm', text: t('otp.mode.setup') }));
  return notice('success', t('otp.mode.live', { name: providerName(chosen.kind) }));
}

function otpFields(values) {
  var providers = values.providers || [];
  return [
    { type: 'row', fields: [
      { name: 'isEnabled', type: 'bool', inline: true, label: 'f.otp.enabled', hint: 'f.otp.enabled.hint' },
      { name: 'sandbox', type: 'bool', inline: true, label: 'f.otp.sandbox', hint: 'f.otp.sandbox.hint' }
    ] },
    { name: 'providerId', type: 'select', label: 'f.otp.provider', hint: 'f.otp.provider.hint', options: function () {
      return [{ value: '', label: t('otp.provider.default') }].concat(providers.map(function (p) {
        var name = providerName(p.kind);
        var flags = [];
        if (!p.isActive) flags.push(t('inactive'));
        if (p.sandbox) flags.push(t('otp.sandbox'));
        return { value: p.id, label: name + (flags.length ? ' (' + flags.join('، ') + ')' : '') };
      }));
    }, noPlaceholder: true },
    { type: 'heading', label: 'otp.code' },
    { type: 'row', fields: [
      { name: 'codeLength', type: 'number', label: 'f.otp.codeLength', hint: 'f.otp.codeLength.hint', required: true, min: 4, max: 8 },
      { name: 'ttlSeconds', type: 'number', label: 'f.otp.ttl', hint: 'f.otp.ttl.hint', required: true, min: 30, max: 1800 }
    ] },
    { type: 'row', fields: [
      { name: 'maxAttempts', type: 'number', label: 'f.otp.attempts', hint: 'f.otp.attempts.hint', required: true, min: 1, max: 10 },
      { name: 'resendSeconds', type: 'number', label: 'f.otp.resend', hint: 'f.otp.resend.hint', required: true, min: 10, max: 600 }
    ] },
    { type: 'heading', label: 'otp.limits' },
    { type: 'row', fields: [
      { name: 'phoneHourly', type: 'number', label: 'f.otp.phoneHourly', required: true, min: 1, max: 100 },
      { name: 'phoneDaily', type: 'number', label: 'f.otp.phoneDaily', required: true, min: 1, max: 500, check: function (v, vals) { return v != null && vals.phoneHourly != null && Number(v) < Number(vals.phoneHourly) ? t('code.DAILY_BELOW_HOURLY') : null; } }
    ] },
    { name: 'ipHourly', type: 'number', label: 'f.otp.ipHourly', hint: 'f.otp.ipHourly.hint', required: true, min: 1, max: 10000 },
    { type: 'heading', label: 'otp.message' },
    tr('message', { label: 'f.otp.message', hint: 'f.otp.message.hint', multiline: true, rows: 3, required: true, max: 300 })
  ];
}

function otpPage(container, query, ctx) {
  var list = listPage(container, 'otp', query, ctx);
  var providers = [];
  var head = h('div', { class: 'mb-3' });
  var page = apiPage(pubApi, '/otp/settings', null, function (values) { providers = values.providers || []; return otpFields(values); }, {
    id: 'otp',
    onchange: function (v) { clear(head).appendChild(otpModeNotice(v, providers)); },
    fixPayload: function (p) { if (!p.providerId) p.providerId = ''; return p; },
    afterSave: function () { list.reload(); }
  });
  var heading = h('h2', { class: 'text-lg font-medium mt-6 mb-2', text: t('otp.recent') });
  container.insertBefore(heading, container.firstChild);
  container.insertBefore(page.el, heading);
  container.insertBefore(head, page.el);
  page.load();
  return list;
}

function appVersionsPage(container, query, ctx) {
  return listPage(container, 'app-versions', query, ctx);
}

function countriesPage(container, query, ctx) {
  return listPage(container, 'countries', query, ctx);
}
