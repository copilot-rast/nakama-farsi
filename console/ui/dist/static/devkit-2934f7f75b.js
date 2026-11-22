(function () {
'use strict';
var DK_TEXT = {
  fa: {
    'group': 'ابزار توسعه',
    'nav.errors': 'پیام‌های خطا',
    'nav.api-keys': 'کلیدهای API',
    'nav.rpc-schema': 'نقشه RPCها',
    'desc.errors': 'متنی که بازیکن برای هر کد خطا می‌بیند. متن هر زبان را می‌شود عوض کرد؛ اگر خالی بماند متن پیش‌فرض نشان داده می‌شود.',
    'desc.api-keys': 'کلیدهای ابزارهای ویرایشگر (مثل افزونه‌های یونیتی). هر کلید فقط دسترسی‌های انتخاب‌شده را دارد و فقط یک بار نشان داده می‌شود.',
    'desc.rpc-schema': 'فهرست RPCهای کلاینت با شکل درخواست و پاسخ، ساخته‌شده از کد سرور. برای ساخت خودکار SDK.',
    'save': 'ذخیره',
    'saved': 'ذخیره شد',
    'cancel': 'انصراف',
    'close': 'بستن',
    'delete': 'حذف',
    'edit': 'ویرایش',
    'reset': 'بازگشت به پیش‌فرض',
    'search': 'جستجو',
    'all': 'همه',
    'refresh': 'تازه‌سازی',
    'loading': 'در حال بارگذاری…',
    'empty': 'موردی پیدا نشد',
    'error.load': 'بارگذاری نشد',
    'retry': 'تلاش دوباره',
    'settings': 'تنظیمات',
    'page.first': 'صفحهٔ اول', 'page.last': 'صفحهٔ آخر', 'page.of': 'صفحهٔ {page} از {pages}', 'pager': 'صفحه‌بندی', 'range': '{from} تا {to} از {total}',
    'prev': 'قبلی',
    'next': 'بعدی',
    'yes': 'بله',
    'no': 'خیر',
    'optional': 'اختیاری',
    'confirm.title': 'مطمئن هستید؟',
    'err.code': 'کد',
    'err.module': 'ماژول',
    'err.status': 'وضعیت HTTP',
    'err.state': 'وضعیت متن',
    'err.state.edited': 'ویرایش‌شده',
    'err.state.custom': 'کد سفارشی',
    'err.state.default': 'پیش‌فرض',
    'err.missing': 'بدون متن در',
    'err.new': 'کد تازه',
    'err.newTitle': 'کد خطای تازه',
    'err.editTitle': 'پیام خطای {code}',
    'err.message': 'متن ({lang})',
    'err.defaultText': 'پیش‌فرض: {text}',
    'err.noDefault': 'متن پیش‌فرض ندارد',
    'err.retry': 'ثانیه صبر پیشنهادی',
    'err.retry.hint': 'اگر سرور خودش زمان صبر را نفرستد، این عدد به SDK داده می‌شود. خالی یعنی هیچ.',
    'err.status.hint': 'خالی یعنی همان وضعیت پیش‌فرض ({status}).',
    'err.placeholders': 'جای‌نگهدار‌هایی مثل {unlockLevel} یا {retryAfter} از meta خطا پر می‌شوند.',
    'err.resetDone': 'به متن پیش‌فرض برگشت',
    'err.deleteCustom': 'این کد سفارشی حذف شود؟',
    'err.count': '{n} کد',
    'set.errors': 'تنظیمات پیام خطا',
    'set.errors.mode': 'شکل خطای RPCهای تازه',
    'set.errors.mode.envelope': 'پاکت کامل JSON (کد، وضعیت، متن، meta، زمان صبر)',
    'set.errors.mode.code': 'فقط کد (SDK متن را از فهرست می‌گیرد)',
    'set.errors.moduleMode': 'شکل خطای RPCهای پایه، اقتصاد و انتشار',
    'set.errors.moduleMode.code': 'مثل قبل: فقط کد (RPCهای شانس و بازی: JSON با کد و وضعیت)',
    'set.errors.moduleMode.envelope': 'پاکت کامل JSON با متن ترجمه‌شده',
    'set.errors.moduleMode.hint': 'کلاینت‌هایی که فقط کد را می‌شناسند با پاکت کار نمی‌کنند؛ پیش از تغییر آن‌ها را به‌روز کنید.',
    'set.errors.fallbackLang': 'زبان جایگزین',
    'set.errors.fallbackLang.hint': 'وقتی متن زبان بازیکن خالی است. خالی یعنی زبان پیش‌فرض بازی.',
    'set.errors.fallbackLang.auto': 'زبان پیش‌فرض بازی',
    'set.errors.preferAccountLanguage': 'زبان حساب بازیکن بر زبان دستگاه مقدم باشد',
    'set.errors.localizeDigits': 'عددهای داخل متن فارسی با رقم فارسی نوشته شوند',
    'set.errors.publicCatalog': 'فهرست خطاها بدون ورود هم در دسترس باشد (برای صفحه ورود)',
    'set.errors.cacheSeconds': 'مدت نگه‌داری فهرست در SDK (ثانیه)',
    'set.errors.includeDetail': 'توضیح فنی سرور هم در پاکت خطا بیاید (فقط برای آزمایش)',
    'key.new': 'کلید تازه',
    'key.name': 'نام',
    'key.description': 'توضیح',
    'key.prefix': 'پیشوند',
    'key.scopes': 'دسترسی‌ها',
    'key.scopes.hint': 'کلید فقط همین کارها را می‌تواند انجام دهد.',
    'key.custom': 'دسترسی دیگر (مثل assets:*)',
    'key.ips': 'نشانی‌های مجاز',
    'key.ips.hint': 'هر خط یک IP یا بازه CIDR. خالی یعنی از هر جا.',
    'key.expiry': 'انقضا',
    'key.expiry.default': 'پیش‌فرض ({days} روز)',
    'key.expiry.never': 'بدون انقضا',
    'key.expiry.days': 'چند روز دیگر',
    'key.expiry.date': 'تاریخ مشخص',
    'key.expiry.keep': 'بدون تغییر',
    'key.days': 'تعداد روز',
    'key.date': 'تاریخ و ساعت',
    'key.status': 'وضعیت',
    'key.status.active': 'فعال',
    'key.status.revoked': 'باطل‌شده',
    'key.status.expired': 'منقضی',
    'key.lastUsed': 'آخرین استفاده',
    'key.never': 'هرگز',
    'key.uses': 'دفعات',
    'key.expires': 'انقضا',
    'key.created': 'ساخته‌شده',
    'key.revoke': 'باطل کردن',
    'key.revoke.confirm': 'کلید «{name}» باطل شود؟ ابزارهایی که از آن استفاده می‌کنند دیگر کار نمی‌کنند.',
    'key.delete.confirm': 'کلید «{name}» برای همیشه حذف شود؟',
    'key.shown.title': 'کلید ساخته شد',
    'key.shown.warn': 'این کلید فقط همین یک بار نشان داده می‌شود. همین حالا کپی و جای امنی نگه دارید.',
    'key.copy': 'کپی',
    'key.copied': 'کپی شد',
    'key.done': 'کپی کردم',
    'key.editTitle': 'کلید {name}',
    'key.newTitle': 'کلید API تازه',
    'scope.title': 'فهرست دسترسی‌ها',
    'scope.name': 'دسترسی',
    'scope.desc': 'توضیح',
    'scope.source': 'منبع',
    'scope.keys': 'کلید فعال',
    'scope.add': 'افزودن دسترسی',
    'scope.placeholder': 'resource:action',
    'set.apiKeys': 'تنظیمات کلیدها',
    'set.apiKeys.defaultExpiryDays': 'انقضای پیش‌فرض (روز، ۰ یعنی بدون انقضا)',
    'set.apiKeys.maxExpiryDays': 'بیشترین عمر کلید (روز، ۰ یعنی بدون سقف)',
    'set.apiKeys.requireExpiry': 'هر کلید باید تاریخ انقضا داشته باشد',
    'set.apiKeys.allowWildcards': 'دسترسی‌های کلی مثل assets:* و * مجاز باشد',
    'set.apiKeys.catalogScopesOnly': 'فقط دسترسی‌های فهرست‌شده قابل انتخاب باشند',
    'set.apiKeys.trustProxyHeaders': 'IP واقعی از X-Forwarded-For خوانده شود (پشت پروکسی)',
    'set.apiKeys.lastUsedEverySeconds': 'فاصله ثبت «آخرین استفاده» (ثانیه)',
    'sc.module': 'ماژول',
    'sc.count': '{n} RPC',
    'sc.request': 'درخواست',
    'sc.response': 'پاسخ',
    'sc.errors': 'کدهای خطای ممکن',
    'sc.format': 'شکل خطا',
    'sc.source': 'فایل',
    'sc.download': 'دانلود JSON',
    'sc.format.code': 'فقط کد',
    'sc.format.json': 'JSON کد/وضعیت/meta',
    'sc.format.envelope': 'پاکت کامل',
    'set.schema': 'دسترسی به نقشه',
    'set.schema.access': 'چه کسی نقشه RPCها را ببیند',
    'set.schema.access.public': 'همه (بدون ورود)',
    'set.schema.access.session': 'بازیکنان واردشده و کلیدهای schema:read',
    'set.schema.access.apikey': 'فقط کلیدهای schema:read',
    'set.schema.access.console': 'فقط کنسول'
  },
  en: {
    'group': 'Developer kit',
    'nav.errors': 'Error messages',
    'nav.api-keys': 'API keys',
    'nav.rpc-schema': 'RPC schema',
    'desc.errors': 'The text players see for each error code. Edit any language; an empty field falls back to the default text.',
    'desc.api-keys': 'Keys for editor tools (such as Unity plugins). Each key only has the scopes you pick and is shown once.',
    'desc.rpc-schema': 'Client RPCs with request and response shapes, generated from the server code. Use it to generate SDKs.',
    'save': 'Save',
    'saved': 'Saved',
    'cancel': 'Cancel',
    'close': 'Close',
    'delete': 'Delete',
    'edit': 'Edit',
    'reset': 'Reset to default',
    'search': 'Search',
    'all': 'All',
    'refresh': 'Refresh',
    'loading': 'Loading…',
    'empty': 'Nothing found',
    'error.load': 'Could not load',
    'retry': 'Try again',
    'settings': 'Settings',
    'page.first': 'First page', 'page.last': 'Last page', 'page.of': 'Page {page} of {pages}', 'pager': 'Pagination', 'range': '{from}-{to} of {total}',
    'prev': 'Previous',
    'next': 'Next',
    'yes': 'Yes',
    'no': 'No',
    'optional': 'optional',
    'confirm.title': 'Are you sure?',
    'err.code': 'Code',
    'err.module': 'Module',
    'err.status': 'HTTP status',
    'err.state': 'Text state',
    'err.state.edited': 'Edited',
    'err.state.custom': 'Custom code',
    'err.state.default': 'Default',
    'err.missing': 'Missing in',
    'err.new': 'New code',
    'err.newTitle': 'New error code',
    'err.editTitle': 'Error message {code}',
    'err.message': 'Text ({lang})',
    'err.defaultText': 'Default: {text}',
    'err.noDefault': 'No default text',
    'err.retry': 'Suggested wait (seconds)',
    'err.retry.hint': 'Sent to the SDK when the server does not give a wait time. Empty means none.',
    'err.status.hint': 'Empty keeps the default status ({status}).',
    'err.placeholders': 'Placeholders such as {unlockLevel} or {retryAfter} are filled from the error meta.',
    'err.resetDone': 'Back to the default text',
    'err.deleteCustom': 'Delete this custom code?',
    'err.count': '{n} codes',
    'set.errors': 'Error message settings',
    'set.errors.mode': 'Error shape for new RPCs',
    'set.errors.mode.envelope': 'Full JSON envelope (code, status, text, meta, wait time)',
    'set.errors.mode.code': 'Code only (the SDK takes the text from the catalog)',
    'set.errors.moduleMode': 'Error shape for base, economy and publication RPCs',
    'set.errors.moduleMode.code': 'As before: code only (chance and game RPCs: JSON with code and status)',
    'set.errors.moduleMode.envelope': 'Full JSON envelope with the translated text',
    'set.errors.moduleMode.hint': 'Clients that only know bare codes do not understand the envelope; update them before switching.',
    'set.errors.fallbackLang': 'Fallback language',
    'set.errors.fallbackLang.hint': 'Used when the player language has no text. Empty means the game default language.',
    'set.errors.fallbackLang.auto': 'Game default language',
    'set.errors.preferAccountLanguage': 'Prefer the account language over the device language',
    'set.errors.localizeDigits': 'Write numbers inside Persian text with Persian digits',
    'set.errors.publicCatalog': 'The error catalog is readable without signing in (for the login screen)',
    'set.errors.cacheSeconds': 'How long the SDK keeps the catalog (seconds)',
    'set.errors.includeDetail': 'Include the technical server detail in the envelope (testing only)',
    'key.new': 'New key',
    'key.name': 'Name',
    'key.description': 'Description',
    'key.prefix': 'Prefix',
    'key.scopes': 'Scopes',
    'key.scopes.hint': 'The key can only do these things.',
    'key.custom': 'Other scope (e.g. assets:*)',
    'key.ips': 'Allowed addresses',
    'key.ips.hint': 'One IP or CIDR range per line. Empty means anywhere.',
    'key.expiry': 'Expiry',
    'key.expiry.default': 'Default ({days} days)',
    'key.expiry.never': 'Never expires',
    'key.expiry.days': 'In some days',
    'key.expiry.date': 'On a date',
    'key.expiry.keep': 'Keep as is',
    'key.days': 'Days',
    'key.date': 'Date and time',
    'key.status': 'Status',
    'key.status.active': 'Active',
    'key.status.revoked': 'Revoked',
    'key.status.expired': 'Expired',
    'key.lastUsed': 'Last used',
    'key.never': 'Never',
    'key.uses': 'Uses',
    'key.expires': 'Expires',
    'key.created': 'Created',
    'key.revoke': 'Revoke',
    'key.revoke.confirm': 'Revoke "{name}"? Tools that use it stop working.',
    'key.delete.confirm': 'Delete "{name}" for good?',
    'key.shown.title': 'Key created',
    'key.shown.warn': 'This key is shown only once. Copy it now and keep it somewhere safe.',
    'key.copy': 'Copy',
    'key.copied': 'Copied',
    'key.done': 'I copied it',
    'key.editTitle': 'Key {name}',
    'key.newTitle': 'New API key',
    'scope.title': 'Scope list',
    'scope.name': 'Scope',
    'scope.desc': 'Description',
    'scope.source': 'Source',
    'scope.keys': 'Active keys',
    'scope.add': 'Add scope',
    'scope.placeholder': 'resource:action',
    'set.apiKeys': 'Key settings',
    'set.apiKeys.defaultExpiryDays': 'Default expiry (days, 0 means never)',
    'set.apiKeys.maxExpiryDays': 'Longest key lifetime (days, 0 means no limit)',
    'set.apiKeys.requireExpiry': 'Every key must have an expiry date',
    'set.apiKeys.allowWildcards': 'Allow wildcard scopes such as assets:* and *',
    'set.apiKeys.catalogScopesOnly': 'Only listed scopes can be picked',
    'set.apiKeys.trustProxyHeaders': 'Read the client IP from X-Forwarded-For (behind a proxy)',
    'set.apiKeys.lastUsedEverySeconds': 'How often "last used" is written (seconds)',
    'sc.module': 'Module',
    'sc.count': '{n} RPCs',
    'sc.request': 'Request',
    'sc.response': 'Response',
    'sc.errors': 'Possible error codes',
    'sc.format': 'Error shape',
    'sc.source': 'File',
    'sc.download': 'Download JSON',
    'sc.format.code': 'Code only',
    'sc.format.json': 'JSON code/status/meta',
    'sc.format.envelope': 'Full envelope',
    'set.schema': 'Schema access',
    'set.schema.access': 'Who can read the RPC schema',
    'set.schema.access.public': 'Everyone (no sign in)',
    'set.schema.access.session': 'Signed-in players and schema:read keys',
    'set.schema.access.apikey': 'Only schema:read keys',
    'set.schema.access.console': 'Console only'
  }
};

function dkLang() {
  var nc = window.NakamaConsole;
  var l = nc && nc.lang ? nc.lang() : (document.documentElement.getAttribute('lang') || 'fa');
  return /^en/i.test(l) ? 'en' : 'fa';
}

function t(key, vars) {
  var dict = DK_TEXT[dkLang()] || DK_TEXT.fa;
  var s = dict[key] != null ? dict[key] : (DK_TEXT.en[key] != null ? DK_TEXT.en[key] : key);
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] != null ? digits(String(vars[k])) : m; });
  return s;
}

function digits(s) {
  if (dkLang() !== 'fa') return s;
  return String(s).replace(/[0-9]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'[+d]; });
}

function fmtDate(v) {
  if (!v) return '';
  if (window.NakamaDate) return window.NakamaDate.format(v);
  var d = new Date(v);
  if (isNaN(d)) return '';
  try {
    return d.toLocaleString(dkLang() === 'fa' ? 'fa-IR' : 'en-GB', { calendar: dkLang() === 'fa' ? 'persian' : 'gregory', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

var DK_API = '/v2/console/base/devkit';
var DK_INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 px-3 py-2 text-sm';
var DK_BTN = 'dk-btn inline-flex items-center justify-center gap-1.5 font-medium whitespace-nowrap rounded-md border text-sm py-1.5 px-3';

function h(tag, props) {
  var el = document.createElement(tag);
  if (props) {
    for (var k in props) {
      var v = props[k];
      if (v === null || v === undefined || v === false) continue;
      if (k === 'class') el.setAttribute('class', v);
      else if (k === 'text') el.textContent = v;
      else if (k.indexOf('on') === 0 && typeof v === 'function') el.addEventListener(k.slice(2).toLowerCase(), v);
      else if (k === 'value') el.value = v;
      else if (k === 'checked') el.checked = !!v;
      else el.setAttribute(k, v === true ? '' : v);
    }
  }
  for (var i = 2; i < arguments.length; i++) add(el, arguments[i]);
  return el;
}

function add(el, child) {
  if (child === null || child === undefined || child === false) return;
  if (Array.isArray(child)) { child.forEach(function (c) { add(el, c); }); return; }
  el.appendChild(typeof child === 'object' ? child : document.createTextNode(String(child)));
}

function clear(el) {
  while (el.firstChild) el.removeChild(el.firstChild);
  return el;
}

function token() {
  var stores = [];
  try { stores.push(window.sessionStorage); } catch (e) { stores.push(null); }
  try { stores.push(window.localStorage); } catch (e) { stores.push(null); }
  for (var i = 0; i < stores.length; i++) {
    try {
      if (!stores[i]) continue;
      var raw = stores[i].getItem('SESSION_TOKENS');
      if (!raw) continue;
      var j = JSON.parse(raw);
      if (j && j.accessToken) return j.accessToken;
    } catch (e) { continue; }
  }
  return null;
}

function DkError(status, code, message, field) {
  this.status = status;
  this.code = code;
  this.message = message;
  this.field = field;
}

function api(method, path, body) {
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  var init = { method: method, headers: headers, credentials: 'same-origin' };
  if (body !== undefined) { headers['Content-Type'] = 'application/json'; init.body = JSON.stringify(body); }
  return fetch(DK_API + path, init).then(function (res) {
    if (res.status === 401 && !/^#\/login/.test(location.hash)) location.hash = '#/login';
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) throw new DkError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', data && data.field || '');
      return data;
    });
  });
}

function qs(params) {
  var out = [];
  Object.keys(params || {}).forEach(function (k) {
    var v = params[k];
    if (v === undefined || v === null || v === '') return;
    out.push(encodeURIComponent(k) + '=' + encodeURIComponent(v));
  });
  return out.length ? '?' + out.join('&') : '';
}

function button(opts) {
  var b = h('button', { type: opts.type || 'button', class: DK_BTN + ' dk-btn--' + (opts.variant || 'ghost') + (opts.class ? ' ' + opts.class : ''), disabled: !!opts.disabled, title: opts.title, 'aria-label': opts.title, onclick: opts.onclick },
    h('span', { text: opts.label }));
  if (opts.data) Object.keys(opts.data).forEach(function (k) { b.setAttribute('data-' + k, opts.data[k]); });
  return b;
}

function busy(btn, on) {
  btn.disabled = !!on;
  btn.classList.toggle('opacity-60', !!on);
}

function notice(kind, text) {
  var color = kind === 'error' ? 'error' : kind === 'success' ? 'success' : kind === 'warning' ? 'warning' : 'secondary';
  return h('div', { class: 'eco-notice eco-notice--' + color + ' flex items-center gap-2 rounded-md px-3 py-2 text-sm', role: kind === 'error' ? 'alert' : 'status' }, h('span', { class: 'flex-1', text: text }));
}

function tag(text, color) {
  return h('span', { class: 'eco-tag eco-tag--' + (color || 'secondary') + ' inline-flex leading-tight px-2 py-0.5 text-xs rounded-md whitespace-nowrap', text: text });
}

function mono(text) {
  return h('span', { class: 'font-mono text-xs', dir: 'ltr', text: text == null ? '' : String(text) });
}

var dkToastBox = null;
function toast(kind, text) {
  if (!dkToastBox || !dkToastBox.isConnected) {
    dkToastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-devkit': 'toasts' });
    document.body.appendChild(dkToastBox);
  }
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  dkToastBox.appendChild(el);
  setTimeout(function () { el.remove(); }, kind === 'error' ? 6000 : 3000);
}

function errorText(err) {
  if (!err) return '';
  var code = err.code || '';
  return (err.message && err.message !== code ? err.message + ' ' : '') + (code ? '(' + code + ')' : '');
}

function card(children, cls) {
  return h('div', { class: 'bg-white dark:bg-secondary-800 rounded-md shadow' + (cls ? ' ' + cls : '') }, children);
}

function field(label, control, hint, extra) {
  var err = h('p', { class: 'dk-field-error eco-tf-err text-xs text-error-600 dark:text-error-400', hidden: true });
  var wrap = h('label', { class: 'dk-field eco-tbf block mb-4' + (extra ? ' ' + extra : '') },
    label ? h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 eco-tf-label', text: label }) : null,
    control,
    hint ? h('span', { class: 'block text-xs text-secondary-500 dark:text-secondary-400 eco-tf-hint', text: hint }) : null,
    err);
  wrap.setError = function (msg) { err.textContent = msg || ''; err.hidden = !msg; };
  if (label && control && control.tagName === 'INPUT' && control.getAttribute('placeholder') === label) wrap.classList.add('eco-tbf-q');
  return wrap;
}

function input(value, opts) {
  opts = opts || {};
  var el = h(opts.multiline ? 'textarea' : 'input', { class: DK_INPUT + (opts.mono ? ' font-mono' : ''), type: opts.multiline ? null : (opts.type || 'text'), dir: opts.ltr ? 'ltr' : 'auto', rows: opts.multiline ? (opts.rows || 3) : null, placeholder: opts.placeholder, name: opts.name, autocomplete: 'off', 'aria-label': opts.label });
  el.value = value == null ? '' : value;
  return el;
}

function select(value, options, opts) {
  opts = opts || {};
  var el = h('select', { class: DK_INPUT + ' eco-select', name: opts.name, 'aria-label': opts.label });
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) opt.selected = true;
    el.appendChild(opt);
  });
  return el;
}

function check(checked, label, name) {
  var box = h('input', { type: 'checkbox', class: 'dk-check', name: name, checked: !!checked });
  var wrap = h('label', { class: 'dk-checkline flex items-start gap-2 mb-3 cursor-pointer text-sm text-secondary-800 dark:text-secondary-200' }, box, h('span', { text: label }));
  wrap.input = box;
  return wrap;
}

function numberValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? Math.round(n) : NaN;
}

var dkPanel = null;
function closePanel() {
  if (dkPanel) { dkPanel.remove(); dkPanel = null; document.removeEventListener('keydown', panelKey); }
}

function panelKey(e) { if (e.key === 'Escape') closePanel(); }

function sidePanel(title, body, footer, name) {
  closePanel();
  var holder = h('div', { class: 'dk-panel-holder' });
  var panel = h('div', { class: 'dk-panel bg-white dark:bg-secondary-900 shadow-lg', role: 'dialog', 'aria-modal': 'true', 'aria-label': title, 'data-devkit-panel': name || '' },
    h('div', { class: 'dk-panel-head flex items-center justify-between border-b px-6 py-3' },
      h('h2', { class: 'text-lg font-semibold truncate', text: title }),
      button({ label: '×', title: t('close'), onclick: closePanel, class: 'dk-close' })),
    h('div', { class: 'dk-panel-body p-6' }, body),
    footer ? h('div', { class: 'dk-panel-foot flex gap-2 justify-end bg-secondary-50 dark:bg-secondary-800 p-4' }, footer) : null);
  holder.appendChild(panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) closePanel(); });
  dkPanel = h('div', { 'data-devkit': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, h('div', { class: 'dk-shade' }), holder);
  document.body.appendChild(dkPanel);
  document.addEventListener('keydown', panelKey);
  return panel;
}

function confirmBox(text, okLabel) {
  return new Promise(function (resolve) {
    var root;
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-devkit': 'confirm', 'data-l10n-skip': '', class: 'dk-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'dk-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('confirm.title') }),
        h('p', { class: 'text-sm mb-4', text: text }),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: okLabel || t('yes'), variant: 'danger', data: { 'devkit-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root);
  });
}

function pageFrame(page, desc, actions) {
  var content = h('div', { class: 'flex flex-col gap-4' });
  var el = h('div', { class: 'dk-page flex flex-col flex-1 px-4 md:px-6 pb-6 max-w-[1600px] w-full mx-auto', 'data-devkit-page': page },
    h('div', { class: 'flex flex-col sm:flex-row gap-3 justify-between items-start mb-4' },
      h('div', {},
        h('h1', { class: 'text-xl font-semibold', text: t('nav.' + page) }),
        h('p', { class: 'text-sm text-secondary-500 dark:text-secondary-400 mt-1 max-w-3xl', text: desc })),
      h('div', { class: 'flex flex-wrap gap-2' }, actions || null)),
    content);
  return { el: el, content: content };
}

function table(columns, rows, onrow) {
  var thead = h('thead', {}, h('tr', {}, columns.map(function (c) { return h('th', { class: 'p-2.5 text-start text-sm font-medium', text: c.label }); })));
  var tbody = h('tbody', {});
  rows.forEach(function (row) {
    var tr = h('tr', { class: 'border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' + (onrow ? ' cursor-pointer' : ''), 'data-devkit-row': row.id || '' });
    if (onrow) {
      tr.tabIndex = 0;
      tr.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,select,label')) onrow(row.item); });
      tr.addEventListener('keydown', function (e) { if (e.key === 'Enter') onrow(row.item); });
    }
    row.cells.forEach(function (cell, i) { tr.appendChild(h('td', { class: 'px-3 py-2 align-middle text-sm', 'data-label': columns[i].label }, cell)); });
    tbody.appendChild(tr);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow rounded-md overflow-x-auto' }, h('table', { class: 'w-full eco-table dk-table' }, thead, tbody));
}

function pager(state, onpage) {
  var total = state.total || 0;
  var nav = h('nav', { class: 'eco-pager', 'aria-label': t('pager') });
  if (!total) { nav.hidden = true; return nav; }
  var from = (state.page - 1) * state.limit + 1;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  function btn(name, path, label, disabled, go) {
    var b = h('button', { type: 'button', class: 'eco-pager-btn', title: label, 'aria-label': label, disabled: !!disabled, onclick: go, 'data-page': name });
    b.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4" aria-hidden="true" data-l10n-icon="mirror">' + path + '</svg>';
    return b;
  }
  nav.appendChild(h('span', { class: 'eco-pager-range tabular-nums', 'data-devkit-range': '', text: t('range', { from: from, to: to, total: total }) }));
  if (last > 1) {
    nav.appendChild(h('div', { class: 'eco-pager-nav' },
      btn('first', '<path d="M7 6v12M17 18l-6-6 6-6"/>', t('page.first'), state.page <= 1, function () { onpage(1); }),
      btn('prev', '<path d="M15 18l-6-6 6-6"/>', t('prev'), state.page <= 1, function () { onpage(state.page - 1); }),
      h('span', { class: 'eco-pager-page tabular-nums', 'aria-current': 'page', text: t('page.of', { page: state.page, pages: last }) }),
      btn('next', '<path d="M9 18l6-6-6-6"/>', t('next'), state.page >= last, function () { onpage(state.page + 1); }),
      btn('last', '<path d="M17 6v12M7 18l6-6-6-6"/>', t('page.last'), state.page >= last, function () { onpage(last); })));
  }
  return nav;
}

var dkSettings = null;
function loadSettings() {
  return api('GET', '/settings').then(function (d) { dkSettings = d; return d; });
}

function settingsCard(section, build) {
  var open = false;
  var body = h('div', { class: 'p-4 border-t border-secondary-200 dark:border-secondary-700', hidden: true });
  var head = h('button', { type: 'button', class: 'dk-settings-toggle w-full flex items-center justify-between px-4 py-3 text-sm font-medium', 'aria-expanded': 'false', 'data-devkit-settings-toggle': section },
    h('span', { text: t('set.' + section) }), h('span', { class: 'dk-caret', text: '▾' }));
  var box = card([head, body], 'dk-settings');
  box.setAttribute('data-devkit-settings', section);
  head.addEventListener('click', function () {
    open = !open;
    head.setAttribute('aria-expanded', open ? 'true' : 'false');
    body.hidden = !open;
    if (open && !body.firstChild) paint();
  });
  function paint() {
    clear(body).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    loadSettings().then(function (d) {
      var cur = JSON.parse(JSON.stringify(d.settings));
      var fields = build(cur[section], d.defaults[section]);
      var msg = h('div', {});
      var save = button({ label: t('save'), variant: 'primary', data: { 'devkit-settings-save': section }, onclick: function () {
        busy(save, true);
        clear(msg);
        cur[section] = fields.read();
        api('PUT', '/settings', cur).then(function (res) {
          busy(save, false);
          dkSettings = res;
          toast('success', t('saved'));
          if (fields.after) fields.after(res);
        }, function (err) {
          busy(save, false);
          msg.appendChild(notice('error', errorText(err)));
        });
      } });
      clear(body).appendChild(h('div', {}, fields.el, msg,
        d.settings.updatedBy ? h('p', { class: 'text-xs text-secondary-500 mt-2', text: d.settings.updatedBy + ' · ' + fmtDate(d.settings.updatedAt) }) : null,
        h('div', { class: 'flex justify-end mt-3' }, save)));
    }, function (err) { clear(body).appendChild(notice('error', errorText(err))); });
  }
  return box;
}

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
    var mods = [{ value: '', label: t('all') }].concat((errorsPage.modules || []).map(function (m) { return { value: m, label: m }; })).concat([{ value: 'custom', label: t('err.state.custom') }]);
    var mod = select(errorQuery.module, mods, { name: 'module', label: t('err.module') });
    mod.addEventListener('change', function () { errorQuery.module = mod.value; errorQuery.page = 1; load(); });
    var state = select(errorQuery.state, [{ value: '', label: t('all') }, { value: 'edited', label: t('err.state.edited') }, { value: 'custom', label: t('err.state.custom') }, { value: 'default', label: t('err.state.default') }], { name: 'state', label: t('err.state') });
    state.addEventListener('change', function () { errorQuery.state = state.value; errorQuery.page = 1; load(); });
    var miss = select(errorQuery.missing, [{ value: '', label: t('all') }].concat((errorsPage.languages || []).map(function (l) { return { value: l.code, label: l.name }; })), { name: 'missing', label: t('err.missing') });
    miss.addEventListener('change', function () { errorQuery.missing = miss.value; errorQuery.page = 1; load(); });
    clear(bar).appendChild(h('div', { class: 'eco-toolbar eco-tbx dk-filters w-full' },
      field(t('search'), search), field(t('err.module'), mod), field(t('err.state'), state), field(t('err.missing'), miss)));
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
  var status = select(keyQuery.status, [{ value: '', label: t('all') }, { value: 'active', label: t('key.status.active') }, { value: 'revoked', label: t('key.status.revoked') }, { value: 'expired', label: t('key.status.expired') }], { name: 'status', label: t('key.status') });
  status.addEventListener('change', function () { keyQuery.status = status.value; keyQuery.page = 1; load(); });
  var search = input(keyQuery.q, { placeholder: t('search'), name: 'q', label: t('search') });
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { keyQuery.q = search.value.trim(); keyQuery.page = 1; load(); }, 250); });
  frame.content.appendChild(settingsCard('apiKeys', keysSettings));
  frame.content.appendChild(h('div', { class: 'eco-toolbar eco-tbx dk-filters' }, field(t('search'), search), field(t('key.status'), status)));
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

var schemaQuery = { module: '', q: '' };

function schemaSettings(cur) {
  var access = select(cur.access, ['public', 'session', 'apikey', 'console'].map(function (v) { return { value: v, label: t('set.schema.access.' + v) }; }), { name: 'access' });
  return { el: field(t('set.schema.access'), access), read: function () { return { access: access.value }; } };
}

function shape(s, depth) {
  depth = depth || 0;
  if (!s || typeof s !== 'object' || !s.type) return 'any';
  var pad = new Array(depth + 2).join('  ');
  var end = new Array(depth + 1).join('  ');
  if (s.type === 'object' && s.properties) {
    var keys = Object.keys(s.properties).sort();
    if (!keys.length) return '{}';
    return '{\n' + keys.map(function (k) { return pad + k + ': ' + shape(s.properties[k], depth + 1); }).join(',\n') + '\n' + end + '}' + (s.nullable ? ' | null' : '');
  }
  if (s.type === 'object' && s.additionalProperties) return '{ [key]: ' + shape(s.additionalProperties, depth + 1) + ' }';
  if (s.type === 'array') return shape(s.items, depth) + '[]' + (s.nullable ? ' | null' : '');
  return s.type + (s.format ? ' (' + s.format + ')' : '') + (s.nullable ? ' | null' : '');
}

function schemaPage(root) {
  var doc = null;
  var download = button({ label: t('sc.download'), data: { 'devkit-action': 'download-schema' }, onclick: function () {
    if (!doc) return;
    var blob = new Blob([JSON.stringify(doc, null, 2)], { type: 'application/json' });
    var a = h('a', { href: URL.createObjectURL(blob), download: 'rpc-schema.json' });
    document.body.appendChild(a);
    a.click();
    setTimeout(function () { URL.revokeObjectURL(a.href); a.remove(); }, 1000);
  } });
  var frame = pageFrame('rpc-schema', t('desc.rpc-schema'), [download]);
  clear(root).appendChild(frame.el);
  var bar = h('div', { class: 'flex flex-wrap gap-2' });
  var list = h('div', { 'data-devkit-list': 'rpc-schema', class: 'flex flex-col gap-2' });
  frame.content.appendChild(settingsCard('schema', schemaSettings));
  frame.content.appendChild(bar);
  frame.content.appendChild(list);

  api('GET', '/rpc-schema').then(function (d) {
    doc = d;
    var mods = {};
    d.rpcs.forEach(function (r) { mods[r.module] = true; });
    var search = input(schemaQuery.q, { placeholder: t('search'), name: 'q', label: t('search') });
    search.addEventListener('input', function () { schemaQuery.q = search.value.trim().toLowerCase(); paint(); });
    var mod = select(schemaQuery.module, [{ value: '', label: t('sc.module') + ': ' + t('all') }].concat(Object.keys(mods).sort().map(function (m) { return { value: m, label: m }; })), { name: 'module', label: t('sc.module') });
    mod.addEventListener('change', function () { schemaQuery.module = mod.value; paint(); });
    clear(bar).appendChild(h('div', { class: 'flex flex-wrap gap-2 w-full' }, h('div', { class: 'flex-1 min-w-[12rem]' }, search), mod));
    paint();
  }, function (err) { clear(list).appendChild(notice('error', t('error.load') + ' ' + errorText(err))); });

  function paint() {
    var items = doc.rpcs.filter(function (r) {
      return (!schemaQuery.module || r.module === schemaQuery.module) && (!schemaQuery.q || r.id.indexOf(schemaQuery.q) !== -1);
    });
    clear(list).appendChild(h('p', { class: 'text-xs text-secondary-500', 'data-devkit-count': '', text: t('sc.count', { n: items.length }) }));
    items.forEach(function (r) {
      var det = h('details', { class: 'dk-rpc bg-white dark:bg-secondary-800 rounded-md shadow', 'data-devkit-rpc': r.id },
        h('summary', { class: 'cursor-pointer px-4 py-2 flex flex-wrap items-center gap-2' }, mono(r.id), tag(r.module), tag(t('sc.format.' + r.errorFormat), r.errorFormat === 'envelope' ? 'success' : 'secondary')),
        h('div', { class: 'px-4 pb-4 grid md:grid-cols-2 gap-4' },
          h('div', {}, h('h3', { class: 'text-xs font-semibold mb-1', text: t('sc.request') }), h('pre', { class: 'dk-pre', dir: 'ltr', text: shape(r.request) })),
          h('div', {}, h('h3', { class: 'text-xs font-semibold mb-1', text: t('sc.response') }), h('pre', { class: 'dk-pre', dir: 'ltr', text: shape(r.response) })),
          h('div', { class: 'md:col-span-2 flex flex-wrap gap-1 items-center' }, h('span', { class: 'text-xs font-semibold me-2', text: t('sc.errors') }), (r.errors || []).map(function (c) { return tag(c, 'warning'); })),
          h('div', { class: 'md:col-span-2 text-xs text-secondary-500' }, t('sc.source') + ': ', mono(r.source + ' · ' + r.handler))));
      list.appendChild(det);
    });
  }
}

var DK_ROOT = 'devkit-root';
var DK_PAGES = ['errors', 'api-keys', 'rpc-schema'];
var DK_RENDER = { errors: errorsPage, 'api-keys': keysPage, 'rpc-schema': schemaPage };
var DK_HASH = /^#\/devkit(?:\/([^/?]*))?/;
var dkPending = null;
var dkRouter = null;
var dkMounted = null;
var dkScheduled = false;

function dkRoute(hash) {
  var m = DK_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  return DK_PAGES.indexOf(m[1]) !== -1 ? m[1] : DK_PAGES[0];
}

function dkAllowed() {
  return !!token();
}

function dkTitle() {
  var page = dkRoute();
  return t('group') + (page ? ' - ' + t('nav.' + page) : '');
}

function dkAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || dkRouter) return !!dkRouter;
  try {
    if (!r.hasRoute('devkit')) {
      var layout = r.resolve('/').matched[0];
      r.addRoute({ path: '/devkit/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'devkit', component: { name: 'DevkitPage', render: function () { return null; } }, meta: { get title() { return dkTitle(); } } }] });
    }
    dkRouter = r;
  } catch (e) {
    return false;
  }
  if (dkPending) {
    var target = dkPending;
    dkPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function dkNavLink(page, active) {
  var on = page === active;
  return h('li', {}, h('a', { href: '#/devkit/' + page, 'data-devkit-nav': page, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + page) }))));
}

function dkSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelector('[data-devkit-group]');
  if (!group || !dkAllowed()) { if (existing) existing.remove(); return; }
  var key = dkLang() + '|' + active;
  if (existing && existing.getAttribute('data-key') === key && existing.previousElementSibling === group) return;
  var built = h('div', { 'data-devkit-group': '', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group') }),
    h('nav', { 'aria-label': t('group') }, h('ul', { class: 'space-y-0.5' }, DK_PAGES.map(function (p) { return dkNavLink(p, active); }))));
  if (existing) existing.remove();
  group.after(built);
}

function dkShowMains() {
  var hidden = document.querySelectorAll('[data-devkit-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-devkit-hidden'); }
}

function dkUnmount() {
  var root = document.getElementById(DK_ROOT);
  if (root) root.remove();
  dkShowMains();
  closePanel();
  dkMounted = null;
}

function dkSync() {
  dkAddRoutes();
  var page = dkRoute();
  dkSyncNav(page);
  if (!page) { if (dkMounted || document.getElementById(DK_ROOT)) dkUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-devkit-hidden', ''); }
  }
  var root = document.getElementById(DK_ROOT);
  if (!root) {
    root = h('div', { id: DK_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-devkit': 'root' });
    dkMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + dkLang();
  if (dkMounted !== key) {
    dkMounted = key;
    closePanel();
    DK_RENDER[page](root);
  }
}

function dkSchedule() {
  if (dkScheduled) return;
  dkScheduled = true;
  requestAnimationFrame(function () { dkScheduled = false; dkSync(); });
}

function dkBoot() {
  window.addEventListener('hashchange', dkSchedule);
  window.addEventListener('popstate', dkSchedule);
  window.addEventListener('nakama-console:ready', dkSchedule);
  new MutationObserver(dkSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { dkMounted = null; dkSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (dkAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  dkSync();
}

window.NakamaDevkit = { version: 1, pages: function () { return DK_PAGES.map(function (p) { return { id: p, href: '#/devkit/' + p, label: t('nav.' + p) }; }); } };

if (DK_HASH.test(location.hash)) {
  dkPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', dkBoot);
else dkBoot();
})();
