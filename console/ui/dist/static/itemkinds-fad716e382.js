(function () {
'use strict';
var IK_TEXT = {
  fa: {
    'group.itemkinds': 'دارایی‌های بازی',
    'nav.itemkinds/routes': 'مسیریاب یونیتی', 'nav.itemkinds/runs': 'گزارش مسیریابی', 'nav.itemkinds/settings': 'تنظیمات دارایی‌ها',
    'desc.itemkinds/kind': 'هر ردیف یک آیتم فروشگاه است با پروفایل این نوع. داده‌ها از یونیتی می‌آیند؛ صاحب هر فیلد (یونیتی یا پنل) در تنظیمات دارایی‌ها مشخص است.',
    'desc.itemkinds/routes': 'هر کارتی که یونیتی می‌فرستد، با این جدول به بخش درستش می‌رود. کارت «وابسته» آیتم نمی‌سازد؛ فقط کارت‌هایی را که به آن ارجاع می‌دهند دوباره می‌سازد.',
    'desc.itemkinds/runs': 'نتیجهٔ هر ارسال از یونیتی و هر مسیریابی دوباره: هر کارت به کدام بخش رفت، چه چیزی عوض شد و کدام در انتظار ماند.',
    'desc.itemkinds/settings': 'رفتار مسیریاب، صاحب هر فیلد (یونیتی یا پنل) و تنظیمات هر نوع.',
    'save': 'ذخیره', 'saved': 'ذخیره شد', 'cancel': 'انصراف', 'close': 'بستن', 'delete': 'حذف', 'search': 'جستجو', 'all': 'همه', 'add': 'افزودن', 'edit': 'ویرایش',
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'yes': 'بله', 'no': 'خیر',
    'confirm.title': 'مطمئن هستید؟', 'lang.text': '{label} ({lang})', 'updated': 'آخرین تغییر: {by} · {at}', 'none': '—',
    'f.key': 'کلید آیتم', 'f.name': 'نام', 'f.status': 'وضعیت', 'f.price': 'قیمت', 'f.currency': 'ارز', 'f.amount': 'مبلغ', 'f.unlockLevel': 'سطح باز شدن', 'f.active': 'در فروشگاه فعال است',
    'f.source': 'کارت یونیتی', 'f.version': 'نسخه', 'f.parent': 'آیتم والد', 'f.requires': 'پیش‌نیاز', 'f.updated': 'آخرین تغییر', 'f.noPrice': 'فروشی نیست (بدون قیمت)',
    'st.ok': 'آماده', 'st.pending': 'در انتظار', 'st.deleted': 'کارتش حذف شد', 'st.unlinked': 'وصل نشده',
    'k.new': 'آیتم تازه', 'k.create': 'ساختن', 'k.created': 'آیتم ساخته شد', 'k.searchHint': 'کلید، نام، شناسهٔ کارت یا والد',
    'k.item': 'آیتم فروشگاه', 'k.profile': 'پروفایل', 'k.sourceBox': 'منبع یونیتی', 'k.view': 'محاسبه‌شده در سرور', 'k.unlockPrices': 'قیمت باز کردن زودتر (inventoryx)',
    'k.pending': 'این آیتم در انتظار است: {reason}', 'k.deleted': 'کارت یونیتی این آیتم حذف شده است؛ آیتم خاموش ماند تا مالکیت بازیکن‌ها از بین نرود.',
    'k.unlinked': 'این آیتم هنوز به کارت یونیتی وصل نشده است. با مسیریابی دوباره، اگر کارتش روی سرور باشد وصل می‌شود.',
    'k.readonly': 'از یونیتی', 'k.reroute': 'مسیریابی دوباره', 'k.empty': 'هنوز آیتمی از این نوع نیست. کارت‌ها را از Asset Sync یونیتی بفرستید یا «آیتم تازه» بزنید.',
    'rr.title': 'مسیریابی دوباره', 'rr.hint': 'همهٔ کارت‌های ذخیره‌شده دوباره از مسیریاب رد می‌شوند، بدون اینکه یونیتی دوباره بفرستد. اول پیش‌نمایش را ببینید.',
    'rr.preview': 'پیش‌نمایش', 'rr.run': 'اجرا', 'rr.force': 'حتی کارت‌هایی که عوض نشده‌اند دوباره ساخته شوند', 'rr.done': 'مسیریابی انجام شد', 'rr.previewDone': 'پیش‌نمایش آماده است؛ چیزی ذخیره نشد',
    'act.created': 'ساخته شد', 'act.updated': 'به‌روز شد', 'act.linked': 'وصل شد', 'act.unchanged': 'بدون تغییر', 'act.pending': 'در انتظار', 'act.deactivated': 'خاموش شد',
    'act.dependency': 'وابسته', 'act.stored': 'فقط ذخیره', 'act.skipped': 'رد شد', 'act.error': 'خطا',
    'r.sourceType': 'کلاس یونیتی', 'r.match': 'شرط', 'r.kind': 'نوع', 'r.role': 'نقش', 'r.priority': 'اولویت', 'r.enabled': 'روشن', 'r.note': 'یادداشت', 'r.origin': 'پیش‌فرض',
    'role.item': 'آیتم می‌سازد', 'role.dependent': 'وابسته', 'role.ignore': 'نادیده',
    'r.new': 'مسیر تازه', 'r.editTitle': 'ویرایش مسیر', 'r.matchHint': 'JSON؛ مثلاً {"category.$.$name":"Vehicle"}. «$» یعنی کارتِ ارجاع‌شده را باز کن؛ $name و $folder و $type از خود کارت‌اند. خالی یعنی همهٔ کارت‌های این کلاس.',
    'r.sourceHint': 'نام کامل کلاس C#، مثل GameData.VehicleCard', 'r.badJson': 'شرط باید JSON درست باشد.', 'r.deleteDefault': 'این مسیر پیش‌فرض است؛ حذف نمی‌شود و فقط خاموش می‌شود.',
    'r.deleteConfirm': 'این مسیر حذف شود؟', 'r.yes': 'پیش‌فرض', 'r.custom': 'دستی',
    'run.source': 'منبع', 'run.at': 'زمان', 'run.actor': 'انجام‌دهنده', 'run.summary': 'خلاصه', 'run.dry': 'پیش‌نمایش', 'run.push': 'ارسال یونیتی', 'run.reroute': 'مسیریابی دوباره',
    'run.rows': 'ردیف‌ها', 'run.asset': 'کارت', 'run.item': 'آیتم', 'run.action': 'نتیجه', 'run.reason': 'توضیح', 'run.changes': 'تغییر', 'run.empty': 'هنوز اجرایی ثبت نشده است.',
    's.routing': 'مسیریاب', 's.enabled': 'بعد از هر ارسال از یونیتی، کارت‌ها به بخش خودشان بروند', 's.onMissingRef': 'کارتی که ارجاعش هنوز نیامده',
    'mr.pending': 'در انتظار بماند (ارسال قبول شود)', 'mr.reject': 'کل ارسال رد شود',
    's.onDelete': 'کارتی که در یونیتی حذف شد', 'od.deactivate': 'آیتم خاموش شود', 'od.keep': 'آیتم دست نخورد',
    's.linkWritesUnity': 'اولین اتصال، مقدارهای یونیتی را روی آیتم بنویسد', 's.maxDepth': 'عمق جستجوی وابسته‌ها', 's.keepRuns': 'تعداد گزارش‌های نگه‌داشته',
    's.inventory': 'پروفایل در فهرست دارایی بازیکن', 's.profile': 'پروفایل در economy_inventory', 'pr.ask': 'فقط وقتی کلاینت withProfile بفرستد', 'pr.always': 'همیشه', 'pr.never': 'هیچ‌وقت',
    's.kind': 'نوع «{name}»', 's.kindEnabled': 'این نوع روشن است', 's.ownership': 'صاحب فیلدها', 'own.unity': 'یونیتی', 'own.console': 'پنل (بعد از اولین اتصال)',
    's.hint.ownership': 'صاحب «پنل» یعنی یونیتی فقط بار اول مقدار را می‌نویسد؛ بعد از آن تغییرهای پنل می‌ماند.', 's.badNumber': 'عدد درست نیست: {field}',
    'fm.addRow': 'افزودن ردیف', 'fm.removeRow': 'حذف ردیف', 'fm.tagsHint': 'با ویرگول جدا کنید', 'fm.badJson': 'JSON درست نیست: {field}', 'fm.badNumber': 'عدد درست نیست: {field}',
    'code.INVALID_PROFILE': 'دادهٔ پروفایل درست نیست', 'code.KIND_NOT_FOUND': 'این نوع ثبت نشده است', 'code.INVALID_ROUTE': 'مسیر درست نیست', 'code.ASSETS_UNAVAILABLE': 'هنوز کارتی از یونیتی روی سرور نیست',
    'code.ITEM_NOT_FOUND': 'آیتم پیدا نشد', 'code.INVALID_KEY': 'کلید درست نیست', 'code.KEY_IMMUTABLE': 'کلید آیتم عوض نمی‌شود', 'code.INVALID_SETTING': 'تنظیم درست نیست'
  },
  en: {
    'group.itemkinds': 'Game assets',
    'nav.itemkinds/routes': 'Unity router', 'nav.itemkinds/runs': 'Routing reports', 'nav.itemkinds/settings': 'Asset settings',
    'desc.itemkinds/kind': 'Each row is a shop item with this kind\'s profile. Data comes from Unity; who owns each field (Unity or the console) is set in the asset settings.',
    'desc.itemkinds/routes': 'Every card Unity pushes goes to its section through this table. A "dependent" card makes no item; it only rebuilds the cards that refer to it.',
    'desc.itemkinds/runs': 'The result of each Unity push and each re-route: where every card went, what changed and what is still pending.',
    'desc.itemkinds/settings': 'How the router behaves, who owns each field (Unity or the console) and the settings of each kind.',
    'save': 'Save', 'saved': 'Saved', 'cancel': 'Cancel', 'close': 'Close', 'delete': 'Delete', 'search': 'Search', 'all': 'All', 'add': 'Add', 'edit': 'Edit',
    'loading': 'Loading…', 'empty': 'Nothing found', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'yes': 'Yes', 'no': 'No',
    'confirm.title': 'Are you sure?', 'lang.text': '{label} ({lang})', 'updated': 'Last change: {by} · {at}', 'none': '—',
    'f.key': 'Item key', 'f.name': 'Name', 'f.status': 'Status', 'f.price': 'Price', 'f.currency': 'Currency', 'f.amount': 'Amount', 'f.unlockLevel': 'Unlock level', 'f.active': 'Active in the shop',
    'f.source': 'Unity card', 'f.version': 'Version', 'f.parent': 'Parent item', 'f.requires': 'Requires', 'f.updated': 'Last change', 'f.noPrice': 'Not for sale (no price)',
    'st.ok': 'Ready', 'st.pending': 'Pending', 'st.deleted': 'Card deleted', 'st.unlinked': 'Not linked',
    'k.new': 'New item', 'k.create': 'Create', 'k.created': 'Item created', 'k.searchHint': 'Key, name, card id or parent',
    'k.item': 'Shop item', 'k.profile': 'Profile', 'k.sourceBox': 'Unity source', 'k.view': 'Computed on the server', 'k.unlockPrices': 'Early unlock prices (inventoryx)',
    'k.pending': 'This item is pending: {reason}', 'k.deleted': 'The Unity card of this item was deleted; the item stays switched off so players keep what they own.',
    'k.unlinked': 'This item is not linked to a Unity card yet. A re-route links it when its card is on the server.',
    'k.readonly': 'from Unity', 'k.reroute': 'Re-route', 'k.empty': 'No items of this kind yet. Push the cards from Unity Asset Sync or press "New item".',
    'rr.title': 'Re-route stored entries', 'rr.hint': 'Every stored card goes through the router again without a new push from Unity. Check the preview first.',
    'rr.preview': 'Preview', 'rr.run': 'Run', 'rr.force': 'Rebuild cards that did not change too', 'rr.done': 'Re-route done', 'rr.previewDone': 'Preview ready; nothing was saved',
    'act.created': 'Created', 'act.updated': 'Updated', 'act.linked': 'Linked', 'act.unchanged': 'Unchanged', 'act.pending': 'Pending', 'act.deactivated': 'Switched off',
    'act.dependency': 'Dependent', 'act.stored': 'Stored only', 'act.skipped': 'Skipped', 'act.error': 'Error',
    'r.sourceType': 'Unity class', 'r.match': 'Condition', 'r.kind': 'Kind', 'r.role': 'Role', 'r.priority': 'Priority', 'r.enabled': 'On', 'r.note': 'Note', 'r.origin': 'Default',
    'role.item': 'Makes an item', 'role.dependent': 'Dependent', 'role.ignore': 'Ignore',
    'r.new': 'New route', 'r.editTitle': 'Edit route', 'r.matchHint': 'JSON, e.g. {"category.$.$name":"Vehicle"}. "$" opens the referenced card; $name, $folder and $type come from the card itself. Empty means every card of this class.',
    'r.sourceHint': 'Full C# class name, like GameData.VehicleCard', 'r.badJson': 'The condition must be valid JSON.', 'r.deleteDefault': 'This is a default route; it is switched off, not deleted.',
    'r.deleteConfirm': 'Delete this route?', 'r.yes': 'Default', 'r.custom': 'Custom',
    'run.source': 'Source', 'run.at': 'Time', 'run.actor': 'By', 'run.summary': 'Summary', 'run.dry': 'Preview', 'run.push': 'Unity push', 'run.reroute': 'Re-route',
    'run.rows': 'Rows', 'run.asset': 'Card', 'run.item': 'Item', 'run.action': 'Result', 'run.reason': 'Detail', 'run.changes': 'Changes', 'run.empty': 'No runs yet.',
    's.routing': 'Router', 's.enabled': 'After each Unity push, send cards to their sections', 's.onMissingRef': 'A card whose reference has not arrived',
    'mr.pending': 'Keep it pending (accept the push)', 'mr.reject': 'Reject the whole push',
    's.onDelete': 'A card deleted in Unity', 'od.deactivate': 'Switch the item off', 'od.keep': 'Leave the item as is',
    's.linkWritesUnity': 'The first link writes the Unity values on the item', 's.maxDepth': 'How deep to look for dependents', 's.keepRuns': 'Reports to keep',
    's.inventory': 'Profile in the player inventory', 's.profile': 'Profile in economy_inventory', 'pr.ask': 'Only when the client sends withProfile', 'pr.always': 'Always', 'pr.never': 'Never',
    's.kind': 'Kind "{name}"', 's.kindEnabled': 'This kind is on', 's.ownership': 'Field owners', 'own.unity': 'Unity', 'own.console': 'Console (after the first link)',
    's.hint.ownership': '"Console" means Unity writes the value only the first time; console edits stay after that.', 's.badNumber': 'Not a valid number: {field}',
    'fm.addRow': 'Add row', 'fm.removeRow': 'Remove row', 'fm.tagsHint': 'Separate with commas', 'fm.badJson': 'Invalid JSON: {field}', 'fm.badNumber': 'Not a valid number: {field}',
    'code.INVALID_PROFILE': 'The profile data is not valid', 'code.KIND_NOT_FOUND': 'This kind is not registered', 'code.INVALID_ROUTE': 'The route is not valid', 'code.ASSETS_UNAVAILABLE': 'No Unity cards are on the server yet',
    'code.ITEM_NOT_FOUND': 'Item not found', 'code.INVALID_KEY': 'Invalid key', 'code.KEY_IMMUTABLE': 'An item key cannot change', 'code.INVALID_SETTING': 'Invalid setting'
  }
};

function ikLang() {
  var nc = window.NakamaConsole;
  var l = nc && nc.lang ? nc.lang() : (document.documentElement.getAttribute('lang') || 'fa');
  return /^en/i.test(l) ? 'en' : 'fa';
}

function has(key) {
  return (IK_TEXT[ikLang()] || IK_TEXT.fa)[key] != null;
}

function t(key, vars) {
  var dict = IK_TEXT[ikLang()] || IK_TEXT.fa;
  var s = dict[key] != null ? dict[key] : (IK_TEXT.en[key] != null ? IK_TEXT.en[key] : key);
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] != null ? digits(String(vars[k])) : m; });
  return s;
}

function digits(s) {
  if (ikLang() !== 'fa') return s;
  return String(s).replace(/[0-9]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'[+d]; }).replace(/,/g, '٬');
}

function fmtNum(n) {
  if (n == null || n === '') return '';
  var s = Number(n).toLocaleString('en-US', { maximumFractionDigits: 2 });
  return digits(s);
}

function fmtDate(v) {
  if (!v) return '';
  var d = new Date(v);
  if (isNaN(d)) return '';
  try {
    return d.toLocaleString(ikLang() === 'fa' ? 'fa-IR' : 'en-GB', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

var IK_API = '/v2/console/base';
var IK_INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 px-3 py-2 text-sm';
var IK_BTN = 'ik-btn inline-flex items-center justify-center gap-1.5 font-medium whitespace-nowrap rounded-md border text-sm py-1.5 px-3';

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

function IkError(status, code, message, field, meta) {
  this.meta = meta || null;
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
  return fetch((path.indexOf('/v2/') === 0 ? '' : IK_API) + path, init).then(function (res) {
    if (res.status === 401 && !/^#\/login/.test(location.hash)) location.hash = '#/login';
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) throw new IkError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', data && data.field || '', data && data.meta);
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
  var b = h('button', { type: opts.type || 'button', class: IK_BTN + ' ik-btn--' + (opts.variant || 'ghost') + (opts.class ? ' ' + opts.class : ''), disabled: !!opts.disabled, title: opts.title, 'aria-label': opts.title, onclick: opts.onclick },
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

var ikToastBox = null;
function toast(kind, text) {
  if (!ikToastBox || !ikToastBox.isConnected) {
    ikToastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-ik': 'toasts' });
    document.body.appendChild(ikToastBox);
  }
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  ikToastBox.appendChild(el);
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
  var err = h('p', { class: 'ik-field-error text-xs text-error-600 dark:text-error-400', hidden: true });
  var wrap = h('label', { class: 'ik-field block mb-4' + (extra ? ' ' + extra : '') },
    label ? h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1', text: label }) : null,
    control,
    hint ? h('span', { class: 'block text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: hint }) : null,
    err);
  wrap.setError = function (msg) { err.textContent = msg || ''; err.hidden = !msg; };
  return wrap;
}

function input(value, opts) {
  opts = opts || {};
  var el = h(opts.multiline ? 'textarea' : 'input', { class: IK_INPUT + (opts.mono ? ' font-mono' : ''), type: opts.multiline ? null : (opts.type || 'text'), dir: opts.ltr ? 'ltr' : 'auto', rows: opts.multiline ? (opts.rows || 3) : null, placeholder: opts.placeholder, name: opts.name, autocomplete: 'off', 'aria-label': opts.label });
  el.value = value == null ? '' : value;
  return el;
}

function select(value, options, opts) {
  opts = opts || {};
  var el = h('select', { class: IK_INPUT + ' eco-select', name: opts.name, 'aria-label': opts.label });
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) opt.selected = true;
    el.appendChild(opt);
  });
  return el;
}

function check(checked, label, name) {
  var box = h('input', { type: 'checkbox', class: 'ik-check', name: name, checked: !!checked });
  var wrap = h('label', { class: 'ik-checkline flex items-start gap-2 mb-3 cursor-pointer text-sm text-secondary-800 dark:text-secondary-200' }, box, h('span', { text: label }));
  wrap.input = box;
  return wrap;
}

function numberValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? Math.round(n) : NaN;
}

var ikPanel = null;
function closePanel() {
  if (ikPanel) { ikPanel.remove(); ikPanel = null; document.removeEventListener('keydown', panelKey); }
}

function panelKey(e) { if (e.key === 'Escape') closePanel(); }

function sidePanel(title, body, footer, name) {
  closePanel();
  var holder = h('div', { class: 'ik-panel-holder' });
  var panel = h('div', { class: 'ik-panel bg-white dark:bg-secondary-900 shadow-lg', role: 'dialog', 'aria-modal': 'true', 'aria-label': title, 'data-ik-panel': name || '' },
    h('div', { class: 'ik-panel-head flex items-center justify-between border-b px-6 py-3' },
      h('h2', { class: 'text-lg font-semibold truncate', text: title }),
      button({ label: '×', title: t('close'), onclick: closePanel, class: 'ik-close' })),
    h('div', { class: 'ik-panel-body p-6' }, body),
    footer ? h('div', { class: 'ik-panel-foot flex gap-2 justify-end bg-secondary-50 dark:bg-secondary-800 p-4' }, footer) : null);
  holder.appendChild(panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) closePanel(); });
  ikPanel = h('div', { 'data-ik': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, h('div', { class: 'ik-shade' }), holder);
  document.body.appendChild(ikPanel);
  document.addEventListener('keydown', panelKey);
  return panel;
}

function confirmBox(text, okLabel) {
  return new Promise(function (resolve) {
    var root;
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-ik': 'confirm', 'data-l10n-skip': '', class: 'ik-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'ik-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('confirm.title') }),
        h('p', { class: 'text-sm mb-4', text: text }),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: okLabel || t('yes'), variant: 'danger', data: { 'ik-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root);
  });
}

function pageFrame(page, desc, actions) {
  var content = h('div', { class: 'flex flex-col gap-4' });
  var el = h('div', { class: 'ik-page flex flex-col flex-1 px-4 md:px-6 pb-6 max-w-[1600px] w-full mx-auto', 'data-ik-page': page },
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
    var tr = h('tr', { class: 'border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' + (onrow ? ' cursor-pointer' : ''), 'data-ik-row': row.id || '' });
    if (onrow) {
      tr.tabIndex = 0;
      tr.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,select,label')) onrow(row.item); });
      tr.addEventListener('keydown', function (e) { if (e.key === 'Enter') onrow(row.item); });
    }
    row.cells.forEach(function (cell, i) { tr.appendChild(h('td', { class: 'px-3 py-2 align-middle text-sm', 'data-label': columns[i].label }, cell)); });
    tbody.appendChild(tr);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow rounded-md overflow-x-auto' }, h('table', { class: 'w-full eco-table ik-table' }, thead, tbody));
}

function pager(state, onpage) {
  var total = state.total || 0;
  var from = total ? (state.page - 1) * state.limit + 1 : 0;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  return h('div', { class: 'flex items-center justify-end gap-2 text-xs text-secondary-500' },
    h('span', { 'data-ik-range': '', text: total ? t('range', { from: from, to: to, total: total }) : t('empty') }),
    button({ label: t('prev'), disabled: state.page <= 1, onclick: function () { onpage(state.page - 1); } }),
    button({ label: t('next'), disabled: state.page >= last, onclick: function () { onpage(state.page + 1); } }));
}

function loadSettings(base) {
  return api('GET', base + '/settings');
}

function floatValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).replace(/٫/g, '.').trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? n : NaN;
}

var ikLanguages = null;
function loadLanguages() {
  if (ikLanguages) return Promise.resolve(ikLanguages);
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch('/v2/console/base/languages', { headers: headers, credentials: 'same-origin' }).then(function (r) { return r.ok ? r.json() : null; }).then(function (d) {
    var list = d && d.items ? d.items.filter(function (l) { return l.isActive; }).map(function (l) { return { code: l.code, name: l.nativeName || l.code, isDefault: !!l.isDefault, direction: l.direction }; }) : [];
    if (!list.length) list = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    list.sort(function (a, b) { return (b.isDefault ? 1 : 0) - (a.isDefault ? 1 : 0); });
    ikLanguages = list;
    return list;
  }, function () {
    ikLanguages = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    return ikLanguages;
  });
}

function translated(label, name, value, opts) {
  opts = opts || {};
  var inputs = {};
  var box = h('div', { class: 'grid sm:grid-cols-2 gap-x-4', 'data-ik-text': name });
  (ikLanguages || []).forEach(function (l) {
    var el = input(value && value[l.code] || '', { name: name + '-' + l.code, multiline: opts.multiline, rows: 2 });
    el.setAttribute('dir', l.direction === 'rtl' ? 'rtl' : 'ltr');
    inputs[l.code] = el;
    box.appendChild(field(t('lang.text', { label: label, lang: l.name }), el));
  });
  box.read = function () {
    var out = {};
    Object.keys(inputs).forEach(function (k) { var v = inputs[k].value.trim(); if (v) out[k] = v; });
    return out;
  };
  return box;
}

function textOf(m, fallback) {
  if (!m) return fallback || '';
  var l = ikLang();
  return m[l] || m.fa || m.en || Object.keys(m).map(function (k) { return m[k]; }).filter(Boolean)[0] || fallback || '';
}

function fieldError(msgBox, err) {
  clear(msgBox).appendChild(notice('error', ikErrText(err)));
}

function options(list, value) {
  return list.map(function (o) { return typeof o === 'string' ? { value: o, label: o } : o; });
}

function dateInput(value, name, type) {
  var el = h('input', { type: type || 'date', class: IK_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  el.value = value || '';
  return el;
}

function isoDay(d) {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

function ikErrText(err) {
  if (err && err.code && has('code.' + err.code)) return t('code.' + err.code, err.meta || null) + (err.field ? ' · ' + err.field : '');
  return errorText(err) + (err && err.field ? ' · ' + err.field : '');
}

function section(title, children, name) {
  return card(h('div', { class: 'p-4', 'data-ik-section': name || '' },
    h('h2', { class: 'font-semibold mb-3', text: title }), children));
}

function settingsPage(root, page, base, build) {
  var frame = pageFrame(page, t('desc.' + page));
  clear(root).appendChild(frame.el);
  frame.content.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  Promise.all([loadSettings(base), loadLanguages()]).then(function (res) {
    var d = res[0];
    var cur = JSON.parse(JSON.stringify(d.settings));
    var form = build(cur, d.defaults, d);
    var msg = h('div', { 'data-ik-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', 'data-ik-updated': '' });
    var paintMeta = function (x) { meta.textContent = x.updatedBy ? t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }) : ''; };
    paintMeta(d);
    var save = button({ label: t('save'), variant: 'primary', data: { 'ik-save': page }, onclick: function () {
      busy(save, true);
      clear(msg);
      var bad = form.read(cur);
      if (bad) { busy(save, false); msg.appendChild(notice('error', bad)); return; }
      api('PUT', base + '/settings', cur).then(function (out) {
        busy(save, false);
        cur = JSON.parse(JSON.stringify(out.settings));
        paintMeta(out);
        toast('success', t('saved'));
        if (form.after) form.after(out);
      }, function (err) {
        busy(save, false);
        msg.appendChild(notice('error', ikErrText(err)));
      });
    } });
    clear(frame.content);
    form.el.forEach(function (el) { frame.content.appendChild(el); });
    frame.content.appendChild(h('div', { class: 'ik-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), save));
  }, function (err) {
    clear(frame.content).appendChild(notice('error', ikErrText(err)));
  });
}

function numInput(value, name, label) {
  var el = input(value == null ? '' : digits(String(value)), { ltr: true, name: name, label: label });
  el.setAttribute('inputmode', 'numeric');
  return el;
}

function hashQuery() {
  var i = location.hash.indexOf('?');
  var out = {};
  if (i === -1) return out;
  location.hash.slice(i + 1).split('&').forEach(function (kv) {
    var p = kv.split('=');
    if (p[0]) out[decodeURIComponent(p[0])] = decodeURIComponent(p[1] || '');
  });
  return out;
}

function playerLink(id, name) {
  if (!id) return h('span', { class: 'text-secondary-400', text: '—' });
  return h('a', { href: '#/players/' + id, class: 'eco-link', 'data-ik-player-link': id, dir: 'auto', text: name || id.slice(0, 8) });
}

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
      var fs = h('fieldset', { class: 'ik-group mb-4', 'data-ik-fieldset': path }, h('legend', { class: 'font-medium text-sm px-1', text: label }), hint ? h('p', { class: 'text-xs text-secondary-500 mb-2', text: hint }) : null, inner.el);
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
      var wrap = h('fieldset', { class: 'ik-group mb-4', 'data-ik-fieldset': path }, h('legend', { class: 'font-medium text-sm px-1', text: label }), hint ? h('p', { class: 'text-xs text-secondary-500 mb-2', text: hint }) : null, holder, addBtn);
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

var IK_BASE = '/itemkinds';
var ikKindCache = null;
var ikKindAt = 0;
var ikKindLoading = null;
var ikCurrencies = null;

function ikLoadKinds(force) {
  if (!force && ikKindCache && Date.now() - ikKindAt < 30000) return Promise.resolve(ikKindCache);
  if (ikKindLoading) return ikKindLoading;
  ikKindLoading = api('GET', IK_BASE + '/kinds').then(function (d) {
    ikKindLoading = null;
    ikKindCache = d.kinds || [];
    ikKindAt = Date.now();
    return ikKindCache;
  }, function (err) {
    ikKindLoading = null;
    throw err;
  });
  return ikKindLoading;
}

function ikKind(key) {
  return (ikKindCache || []).filter(function (k) { return k.key === key; })[0] || null;
}

function loadCurrencies() {
  if (ikCurrencies) return Promise.resolve(ikCurrencies);
  return api('GET', '/v2/console/economy/currencies?limit=100').then(function (d) {
    ikCurrencies = (d && (d.items || d.currencies) || []).map(function (c) { return c.key; });
    return ikCurrencies;
  }, function () { ikCurrencies = []; return ikCurrencies; });
}

function kindPage(root) {
  var key = (location.hash.split('?')[0].split('/')[3] || '');
  var kind = ikKind(key);
  if (!kind) { clear(root).appendChild(h('div', { class: 'px-6' }, notice('error', t('code.KIND_NOT_FOUND')))); return; }
  var q = hashQuery();
  var state = { q: q.q || '', status: q.status || '', parent: q.parent || '', page: 1, limit: 25 };
  var actions = [
    button({ label: t('k.reroute'), data: { 'ik-reroute': kind.key }, onclick: function () { reroutePanel(kind.key, function () { paint(); }); } }),
    button({ label: t('k.new'), variant: 'primary', data: { 'ik-new': kind.key }, onclick: function () { createPanel(kind, function () { paint(); }); } })
  ];
  var frame = pageFrame('itemkinds/kind', t('desc.itemkinds/kind'), actions);
  frame.el.querySelector('h1').textContent = textOf(kind.label, kind.key);
  clear(root).appendChild(frame.el);
  var counts = h('div', { class: 'flex flex-wrap gap-2', 'data-ik-counts': kind.key });
  ['ok', 'pending', 'deleted'].forEach(function (st) { counts.appendChild(h('span', { class: 'inline-flex items-center gap-1 text-sm' }, statusTag(st), h('span', { text: fmtNum(kind.counts && kind.counts[st] || 0) }))); });
  var search = input(state.q, { name: 'q', label: t('search'), placeholder: t('k.searchHint') });
  var status = select(state.status, [{ value: '', label: t('all') }, { value: 'ok', label: t('st.ok') }, { value: 'pending', label: t('st.pending') }, { value: 'deleted', label: t('st.deleted') }, { value: 'unlinked', label: t('st.unlinked') }], { name: 'status', label: t('f.status') });
  var timer = null;
  search.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(function () { state.q = search.value.trim(); state.page = 1; paint(); }, 300); });
  status.addEventListener('change', function () { state.status = status.value; state.page = 1; paint(); });
  var filters = h('div', { class: 'ik-filters flex flex-wrap gap-3 items-end' }, field(t('search'), search), field(t('f.status'), status));
  var list = h('div', { 'data-ik-list': kind.key });
  frame.content.appendChild(counts);
  frame.content.appendChild(filters);
  frame.content.appendChild(list);
  function paint() {
    clear(list).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', IK_BASE + '/kinds/' + kind.key + '/items' + qs({ q: state.q, status: state.status, parent: state.parent, page: state.page, limit: state.limit })).then(function (d) {
      clear(list);
      state.total = d.total;
      if (!d.items.length) { list.appendChild(h('div', { class: 'p-6 text-sm text-secondary-500 bg-white dark:bg-secondary-800 rounded-md shadow', 'data-ik-empty': '', text: state.q || state.status ? t('empty') : t('k.empty') })); return; }
      var cols = [{ label: t('f.key') }, { label: t('f.name') }, { label: t('f.status') }, { label: t('f.price') }, { label: t('f.unlockLevel') }];
      (kind.columns || []).forEach(function (c) { cols.push({ label: textOf(c.label, c.key) }); });
      cols.push({ label: t('f.version') });
      var rows = d.items.map(function (it) {
        var cells = [mono(it.key), h('span', { dir: 'auto', text: textOf(it.name, it.key) }), h('span', { class: 'inline-flex gap-1 items-center' }, statusTag(it.status), it.isActive ? null : tag(t('no') + ' · ' + t('f.active'), 'secondary')),
          h('span', { text: priceText(it.price) }), h('span', { text: it.unlockLevel ? fmtNum(it.unlockLevel) : t('none') })];
        (kind.columns || []).forEach(function (c) { cells.push(h('span', { dir: 'auto', text: plainValue(valueAt(it, c.path)) })); });
        cells.push(h('span', { text: it.sourceVersion ? fmtNum(it.sourceVersion) : t('none') }));
        return { id: it.key, item: it, cells: cells };
      });
      list.appendChild(table(cols, rows, function (it) { itemPanel(kind, it.id, paint); }));
      list.appendChild(h('div', { class: 'mt-3' }, pager(state, function (p) { state.page = p; paint(); })));
    }, function (err) { clear(list).appendChild(notice('error', ikErrText(err))); });
  }
  paint();
}

function itemFields(item, currencies) {
  var name = translated(t('f.name'), 'name', item.name || {});
  var cur = select(item.price && item.price.currencyKey || '', [{ value: '', label: t('none') }].concat(currencies.map(function (c) { return { value: c, label: c }; })), { name: 'price.currency', label: t('f.currency') });
  var amount = numInput(item.price ? item.price.amount : '', 'price.amount', t('f.amount'));
  var level = numInput(item.unlockLevel || '', 'unlockLevel', t('f.unlockLevel'));
  var active = check(item.isActive !== false, t('f.active'), 'isActive');
  var el = section(t('k.item'), [h('div', { class: 'ik-row' }, field(t('f.currency'), cur), field(t('f.amount'), amount), field(t('f.unlockLevel'), level)), name, active], 'item');
  return {
    el: el,
    read: function (errors) {
      var out = { name: name.read(), isActive: active.input.checked };
      var amt = latinNumber(amount.value);
      if (cur.value && amt !== '') {
        var n = Number(amt);
        if (!isFinite(n) || n < 0) errors.push(t('fm.badNumber', { field: t('f.amount') })); else out.price = { currencyKey: cur.value, amount: Math.round(n) };
      } else if (!cur.value && amt === '') out.noPrice = true;
      var lv = latinNumber(level.value);
      if (lv === '') out.unlockLevel = 0;
      else if (!isFinite(Number(lv)) || Number(lv) < 0) errors.push(t('fm.badNumber', { field: t('f.unlockLevel') }));
      else out.unlockLevel = Math.round(Number(lv));
      return out;
    }
  };
}

function itemPanel(kind, id, done) {
  var body = h('div', { class: 'flex flex-col gap-4' }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  var msg = h('div', { 'data-ik-msg': '' });
  var save = button({ label: t('save'), variant: 'primary', data: { 'ik-save': 'item' } });
  sidePanel(textOf(kind.label, kind.key), [msg, body], [button({ label: t('cancel'), onclick: closePanel }), save], 'item');
  Promise.all([api('GET', IK_BASE + '/items/' + id), loadCurrencies(), loadLanguages()]).then(function (res) {
    var d = res[0];
    var item = d.item;
    var profile = d.profile || { status: 'unlinked', data: {} };
    clear(body);
    var head = h('div', { class: 'flex flex-wrap items-center gap-2' }, mono(item.key), statusTag(profile.status || 'unlinked'));
    body.appendChild(head);
    if (profile.status === 'pending') body.appendChild(notice('warning', t('k.pending', null).replace('{reason}', profile.reason || '')));
    if (profile.status === 'deleted') body.appendChild(notice('warning', t('k.deleted')));
    if (!d.profile) body.appendChild(notice('info', t('k.unlinked')));
    var itemForm = itemFields(item, res[1]);
    body.appendChild(itemForm.el);
    var prof = schemaForm(kind.fields, profile.data || {}, { name: 'data' });
    body.appendChild(section(t('k.profile'), prof.el, 'profile'));
    if (d.view && Object.keys(d.view).length) body.appendChild(section(t('k.view'), h('pre', { class: 'ik-pre', dir: 'ltr', 'data-ik-view': '', text: JSON.stringify(d.view, null, 2) }), 'view'));
    if (d.hasUnlockPrices) body.appendChild(section(t('k.unlockPrices'), h('div', { class: 'text-sm', 'data-ik-unlock': '' }, (d.unlockPrices || []).length ? d.unlockPrices.map(function (p) { return h('div', { text: priceText(p) }); }) : t('none')), 'unlock'));
    var src = h('dl', { class: 'ik-dl text-sm' });
    var row = function (k, v) { src.appendChild(h('dt', { text: k })); src.appendChild(h('dd', {}, v)); };
    row(t('f.source'), d.source ? h('span', {}, mono(d.source.assetId), ' · ', mono(d.source.type)) : mono(profile.sourceAssetId || t('none')));
    row(t('f.version'), h('span', { text: profile.sourceVersion ? fmtNum(profile.sourceVersion) : t('none') }));
    row(t('f.parent'), mono(item.parentKey || t('none')));
    row(t('f.requires'), mono(item.requiresKey || t('none')));
    row(t('f.updated'), h('span', { text: profile.updatedBy ? t('updated', { by: profile.updatedBy, at: fmtDate(profile.updatedAt) }) : t('none') }));
    body.appendChild(section(t('k.sourceBox'), src, 'source'));
    save.addEventListener('click', function () {
      var errors = [];
      var payload = itemForm.read(errors);
      payload.data = prof.read(JSON.parse(JSON.stringify(profile.data || {})), errors);
      clear(msg);
      if (errors.length) { msg.appendChild(notice('error', errors[0])); return; }
      busy(save, true);
      api('PUT', IK_BASE + '/items/' + id + qs({ kind: kind.key }), payload).then(function () {
        busy(save, false);
        toast('success', t('saved'));
        closePanel();
        if (done) done();
      }, function (err) { busy(save, false); msg.appendChild(notice('error', ikErrText(err))); });
    });
  }, function (err) { clear(body).appendChild(notice('error', ikErrText(err))); });
}

function createPanel(kind, done) {
  var msg = h('div', { 'data-ik-msg': '' });
  var body = h('div', { class: 'flex flex-col gap-4' }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  var save = button({ label: t('k.create'), variant: 'primary', data: { 'ik-save': 'create' } });
  sidePanel(t('k.new') + ' · ' + textOf(kind.label, kind.key), [msg, body], [button({ label: t('cancel'), onclick: closePanel }), save], 'create');
  Promise.all([loadCurrencies(), loadLanguages()]).then(function (res) {
    clear(body);
    var key = input('', { ltr: true, name: 'key', label: t('f.key') });
    body.appendChild(field(t('f.key'), key));
    var itemForm = itemFields({ isActive: true }, res[0]);
    body.appendChild(itemForm.el);
    var prof = schemaForm(kind.fields, {}, { name: 'data' });
    body.appendChild(section(t('k.profile'), prof.el, 'profile'));
    save.addEventListener('click', function () {
      var errors = [];
      var payload = itemForm.read(errors);
      payload.key = key.value.trim();
      payload.data = prof.read({}, errors);
      clear(msg);
      if (errors.length) { msg.appendChild(notice('error', errors[0])); return; }
      busy(save, true);
      api('POST', IK_BASE + '/kinds/' + kind.key + '/items', payload).then(function () {
        busy(save, false);
        toast('success', t('k.created'));
        closePanel();
        ikLoadKinds(true);
        if (done) done();
      }, function (err) { busy(save, false); msg.appendChild(notice('error', ikErrText(err))); });
    });
  });
}

function runRowsTable(rows) {
  if (!rows || !rows.length) return h('p', { class: 'text-sm text-secondary-500', text: t('empty') });
  var cols = [{ label: t('run.asset') }, { label: t('r.sourceType') }, { label: t('r.kind') }, { label: t('run.item') }, { label: t('run.action') }, { label: t('run.reason') }, { label: t('run.changes') }];
  return table(cols, rows.slice(0, 500).map(function (r) {
    var ch = (r.changes || []).map(function (c) { return c.field; }).join('، ');
    return { id: r.assetId, cells: [mono(r.assetId), mono(r.type || ''), h('span', { text: r.kind || t('none') }), mono(r.itemKey || ''), actionTag(r.action), h('span', { dir: 'auto', text: (r.code ? r.code + ' ' : '') + (r.reason || (r.parents && r.parents.length ? r.parents.join(', ') : '')) }), h('span', { dir: 'ltr', class: 'text-xs', text: ch })] };
  }));
}

function summaryTags(sum) {
  var box = h('div', { class: 'flex flex-wrap gap-2', 'data-ik-summary': '' });
  Object.keys(sum || {}).sort().forEach(function (a) { box.appendChild(h('span', { class: 'inline-flex items-center gap-1 text-sm', 'data-ik-sum': a }, actionTag(a), h('span', { text: fmtNum(sum[a]) }))); });
  if (!box.firstChild) box.appendChild(h('span', { class: 'text-sm text-secondary-500', text: t('empty') }));
  return box;
}

function reroutePanel(kindKey, done) {
  var msg = h('div', { 'data-ik-msg': '' });
  var force = check(false, t('rr.force'), 'force');
  var result = h('div', { class: 'flex flex-col gap-3', 'data-ik-result': '' });
  var run = function (dry, btn) {
    clear(msg);
    busy(btn, true);
    clear(result).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('POST', IK_BASE + '/reroute', { dryRun: dry, force: force.input.checked, kind: kindKey || '' }).then(function (out) {
      busy(btn, false);
      clear(result);
      result.appendChild(notice(dry ? 'info' : 'success', dry ? t('rr.previewDone') : t('rr.done')));
      result.appendChild(summaryTags(out.summary));
      result.appendChild(runRowsTable(out.rows));
      if (!dry) { ikLoadKinds(true); if (done) done(); }
    }, function (err) { busy(btn, false); clear(result); msg.appendChild(notice('error', ikErrText(err))); });
  };
  var preview = button({ label: t('rr.preview'), data: { 'ik-preview': '' } });
  var go = button({ label: t('rr.run'), variant: 'primary', data: { 'ik-run': '' } });
  preview.addEventListener('click', function () { run(true, preview); });
  go.addEventListener('click', function () { run(false, go); });
  sidePanel(t('rr.title'), [h('p', { class: 'text-sm text-secondary-600 dark:text-secondary-300 mb-3', text: t('rr.hint') }), force, msg, result], [button({ label: t('close'), onclick: closePanel }), preview, go], 'reroute');
}

function routesPage(root) {
  var frame = pageFrame('itemkinds/routes', t('desc.itemkinds/routes'), [
    button({ label: t('rr.title'), data: { 'ik-reroute': 'all' }, onclick: function () { reroutePanel('', function () { paint(); }); } }),
    button({ label: t('r.new'), variant: 'primary', data: { 'ik-route-new': '' }, onclick: function () { routePanel(null, meta, paint); } })
  ]);
  clear(root).appendChild(frame.el);
  var list = h('div', { 'data-ik-routes': '' });
  frame.content.appendChild(list);
  var meta = { kinds: [], roles: [] };
  function paint() {
    clear(list).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
    api('GET', IK_BASE + '/routes').then(function (d) {
      meta = d;
      clear(list);
      if (!d.routes.length) { list.appendChild(h('div', { class: 'p-6 text-sm text-secondary-500 bg-white dark:bg-secondary-800 rounded-md shadow', text: t('empty') })); return; }
      var cols = [{ label: t('r.sourceType') }, { label: t('r.match') }, { label: t('r.role') }, { label: t('r.kind') }, { label: t('r.priority') }, { label: t('r.enabled') }, { label: t('r.origin') }];
      list.appendChild(table(cols, d.routes.map(function (r) {
        var match = Object.keys(r.match || {}).length ? JSON.stringify(r.match) : t('all');
        return { id: String(r.id), item: r, cells: [mono(r.sourceType), h('span', { dir: 'ltr', class: 'font-mono text-xs', text: match }), tag(t('role.' + r.role), r.role === 'item' ? 'primary' : 'secondary'),
          h('span', { text: r.kind ? textOf((ikKind(r.kind) || {}).label, r.kind) : t('none') }), h('span', { text: fmtNum(r.priority) }), tag(r.enabled ? t('yes') : t('no'), r.enabled ? 'success' : 'secondary'),
          h('span', { text: r.origin ? t('r.yes') : t('r.custom') })] };
      }), function (r) { routePanel(r, meta, paint); }));
    }, function (err) { clear(list).appendChild(notice('error', ikErrText(err))); });
  }
  paint();
}

function routePanel(route, meta, done) {
  var r = route || { sourceType: '', match: {}, role: 'item', kind: (meta.kinds || [])[0] || '', priority: 100, enabled: true, note: '' };
  var msg = h('div', { 'data-ik-msg': '' });
  var src = input(r.sourceType, { ltr: true, name: 'sourceType', label: t('r.sourceType') });
  var match = input(Object.keys(r.match || {}).length ? JSON.stringify(r.match, null, 2) : '', { multiline: true, rows: 4, ltr: true, mono: true, name: 'match', label: t('r.match') });
  var role = select(r.role, (meta.roles || ['item', 'dependent', 'ignore']).map(function (x) { return { value: x, label: t('role.' + x) }; }), { name: 'role', label: t('r.role') });
  var kind = select(r.kind || '', [{ value: '', label: t('none') }].concat((meta.kinds || []).map(function (k) { return { value: k, label: textOf((ikKind(k) || {}).label, k) }; })), { name: 'kind', label: t('r.kind') });
  var prio = numInput(r.priority, 'priority', t('r.priority'));
  var on = check(r.enabled !== false, t('r.enabled'), 'enabled');
  var note = input(r.note || '', { name: 'note', label: t('r.note') });
  var kindField = field(t('r.kind'), kind);
  var syncKind = function () { kindField.hidden = role.value !== 'item'; };
  role.addEventListener('change', syncKind);
  syncKind();
  var save = button({ label: t('save'), variant: 'primary', data: { 'ik-save': 'route' }, onclick: function () {
    clear(msg);
    var m = {};
    var raw = match.value.trim();
    if (raw) { try { m = JSON.parse(raw); } catch (e) { msg.appendChild(notice('error', t('r.badJson'))); return; } }
    if (!m || typeof m !== 'object' || Array.isArray(m)) { msg.appendChild(notice('error', t('r.badJson'))); return; }
    var p = Number(latinNumber(prio.value) || 100);
    var body = { sourceType: src.value.trim(), match: m, role: role.value, kind: role.value === 'item' ? kind.value : '', priority: isFinite(p) ? Math.round(p) : 100, enabled: on.input.checked, note: note.value.trim() };
    busy(save, true);
    api(route ? 'PUT' : 'POST', IK_BASE + '/routes' + (route ? '/' + route.id : ''), body).then(function () {
      busy(save, false); toast('success', t('saved')); closePanel(); done();
    }, function (err) { busy(save, false); msg.appendChild(notice('error', ikErrText(err))); });
  } });
  var footer = [button({ label: t('cancel'), onclick: closePanel }), save];
  if (route) footer.unshift(button({ label: t('delete'), variant: 'danger', data: { 'ik-route-delete': String(route.id) }, onclick: function () {
    confirmBox(route.origin ? t('r.deleteDefault') : t('r.deleteConfirm'), t('delete')).then(function (ok) {
      if (!ok) return;
      api('DELETE', IK_BASE + '/routes/' + route.id).then(function () { toast('success', t('saved')); closePanel(); done(); }, function (err) { msg.appendChild(notice('error', ikErrText(err))); });
    });
  } }));
  sidePanel(route ? t('r.editTitle') : t('r.new'), [msg, field(t('r.sourceType'), src, t('r.sourceHint')), field(t('r.match'), match, t('r.matchHint')), h('div', { class: 'ik-row' }, field(t('r.role'), role), kindField, field(t('r.priority'), prio)), on, field(t('r.note'), note)], footer, 'route');
}

function runsPage(root) {
  var frame = pageFrame('itemkinds/runs', t('desc.itemkinds/runs'));
  clear(root).appendChild(frame.el);
  var list = h('div', { 'data-ik-runs': '' });
  frame.content.appendChild(list);
  clear(list).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  api('GET', IK_BASE + '/runs?limit=50').then(function (d) {
    clear(list);
    if (!d.runs.length) { list.appendChild(h('div', { class: 'p-6 text-sm text-secondary-500 bg-white dark:bg-secondary-800 rounded-md shadow', text: t('run.empty') })); return; }
    var cols = [{ label: t('run.at') }, { label: t('run.source') }, { label: t('run.actor') }, { label: t('run.summary') }];
    list.appendChild(table(cols, d.runs.map(function (r) {
      return { id: String(r.id), item: r, cells: [h('span', { text: fmtDate(r.createdAt) }), h('span', { class: 'inline-flex gap-1' }, tag(t('run.' + r.source), 'secondary'), r.dryRun ? tag(t('run.dry'), 'warning') : null), mono(r.actor), summaryTags(r.summary)] };
    }), function (r) { runPanel(r.id); }));
  }, function (err) { clear(list).appendChild(notice('error', ikErrText(err))); });
}

function runPanel(id) {
  var body = h('div', { class: 'flex flex-col gap-3' }, h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  sidePanel(t('nav.itemkinds/runs'), body, [button({ label: t('close'), onclick: closePanel })], 'run');
  api('GET', IK_BASE + '/runs/' + id).then(function (d) {
    clear(body);
    body.appendChild(h('p', { class: 'text-sm', text: fmtDate(d.createdAt) + ' · ' + t('run.' + d.source) + ' · ' + d.actor }));
    body.appendChild(summaryTags(d.summary));
    body.appendChild(runRowsTable(d.rows));
  }, function (err) { clear(body).appendChild(notice('error', ikErrText(err))); });
}

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

var IK_ROOT = 'itemkinds-root';
var IK_STATIC = ['routes', 'runs', 'settings'];
var IK_HASH = /^#\/(itemkinds)(?:\/([^/?]*))?(?:\/([^/?]+))?/;
var ikPending = null;
var ikRouter = null;
var ikMounted = null;
var ikScheduled = false;
var ikFetched = false;

function ikRoute(hash) {
  var m = IK_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  if (m[2] === 'kind' && m[3]) return 'itemkinds/kind';
  if (IK_STATIC.indexOf(m[2]) !== -1) return 'itemkinds/' + m[2];
  var first = (ikKindCache || [])[0];
  return first ? 'itemkinds/kind' : 'itemkinds/routes';
}

function ikAllowed() {
  return !!token();
}

function ikNavItems() {
  var out = [];
  (ikKindCache || []).forEach(function (k) {
    if (k.enabled === false) return;
    out.push({ id: 'itemkinds/kind/' + k.key, label: textOf(k.label, k.key) });
  });
  IK_STATIC.forEach(function (p) { out.push({ id: 'itemkinds/' + p, label: t('nav.itemkinds/' + p) }); });
  return out;
}

function ikTitle() {
  var page = ikRoute();
  if (!page) return '';
  if (page === 'itemkinds/kind') {
    var k = ikKind((location.hash.split('?')[0].split('/')[3] || ''));
    return t('group.itemkinds') + ' - ' + (k ? textOf(k.label, k.key) : '');
  }
  return t('group.itemkinds') + ' - ' + t('nav.' + page);
}

function ikAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || ikRouter) return !!ikRouter;
  try {
    var layout = r.resolve('/').matched[0];
    if (!r.hasRoute('itemkinds')) {
      r.addRoute({ path: '/itemkinds/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'itemkinds', component: { name: 'ItemKindsPage', render: function () { return null; } }, meta: { get title() { return ikTitle(); } } }] });
    }
    ikRouter = r;
  } catch (e) {
    return false;
  }
  if (ikPending) {
    var target = ikPending;
    ikPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function ikActiveId() {
  var page = ikRoute();
  if (page === 'itemkinds/kind') return 'itemkinds/kind/' + (location.hash.split('?')[0].split('/')[3] || '');
  return page;
}

function ikNavLink(item, active) {
  var on = item.id === active;
  return h('li', {}, h('a', { href: '#/' + item.id, 'data-ik-nav': item.id, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: item.label }))));
}

function ikSyncNav() {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelectorAll('[data-ik-group]');
  var kinds = (ikKindCache || []).filter(function (k) { return k.enabled !== false; });
  if (!group || !ikAllowed() || !kinds.length) { Array.prototype.forEach.call(existing, function (e) { e.remove(); }); return; }
  var target = document.querySelector('.x-drawer [data-cup-group]') || document.querySelector('.x-drawer [data-clan-group="trophies"]') || document.querySelector('.x-drawer [data-quiz-group]') || document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var active = ikActiveId();
  var items = ikNavItems();
  var key = ikLang() + '|' + active + '|' + items.map(function (i) { return i.id + ':' + i.label; }).join(',');
  var mine = document.querySelector('[data-ik-group="itemkinds"]');
  if (mine && existing.length === 1 && mine.getAttribute('data-key') === key && mine.nextElementSibling === target) return;
  Array.prototype.forEach.call(existing, function (e) { e.remove(); });
  target.before(h('div', { 'data-ik-group': 'itemkinds', 'data-plane-group': 'itemkinds', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group.itemkinds') }),
    h('nav', { 'aria-label': t('group.itemkinds') }, h('ul', { class: 'space-y-0.5' }, items.map(function (i) { return ikNavLink(i, active); })))));
}

function ikShowMains() {
  var hidden = document.querySelectorAll('[data-ik-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-ik-hidden'); }
}

function ikUnmount() {
  var root = document.getElementById(IK_ROOT);
  if (root) root.remove();
  ikShowMains();
  closePanel();
  ikMounted = null;
}

function ikRender(page, root) {
  if (page === 'itemkinds/kind') {
    var key = location.hash.split('?')[0].split('/')[3] || '';
    if (!key && ikKindCache && ikKindCache[0]) { location.replace('#/itemkinds/kind/' + ikKindCache[0].key); return; }
    kindPage(root);
  } else if (page === 'itemkinds/routes') routesPage(root);
  else if (page === 'itemkinds/runs') runsPage(root);
  else ikSettingsPage(root);
}

function ikSync() {
  ikAddRoutes();
  if (ikAllowed() && !ikFetched && !ikKindLoading) {
    ikFetched = true;
    ikLoadKinds().then(function () { ikMounted = null; ikSchedule(); }, function () { ikFetched = false; });
  }
  if (!ikAllowed()) ikFetched = false;
  var page = ikRoute();
  ikSyncNav();
  if (!page) { if (ikMounted || document.getElementById(IK_ROOT)) ikUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-ik-hidden', ''); }
  }
  var root = document.getElementById(IK_ROOT);
  if (!root) {
    root = h('div', { id: IK_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-ik': 'root' });
    ikMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  if (!ikKindCache) {
    if (ikMounted !== 'loading') { ikMounted = 'loading'; clear(root).appendChild(h('p', { class: 'px-6 text-sm text-secondary-500', text: t('loading') })); }
    return;
  }
  var key = page + '|' + ikLang() + '|' + location.hash.split('?')[0];
  if (ikMounted !== key) {
    ikMounted = key;
    closePanel();
    ikRender(page, root);
  }
}

function ikSchedule() {
  if (ikScheduled) return;
  ikScheduled = true;
  requestAnimationFrame(function () { ikScheduled = false; ikSync(); });
}

function ikBoot() {
  window.addEventListener('hashchange', ikSchedule);
  window.addEventListener('popstate', ikSchedule);
  window.addEventListener('nakama-console:ready', ikSchedule);
  new MutationObserver(ikSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { ikMounted = null; ikSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (ikAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  ikSync();
}

window.NakamaItemKinds = { version: 1, pages: function () {
  return ikNavItems().map(function (i) { return { id: i.id, group: 'itemkinds', href: '#/' + i.id, label: i.label }; });
}, reload: function () { return ikLoadKinds(true).then(function () { ikMounted = null; ikSchedule(); }); } };

if (IK_HASH.test(location.hash)) {
  ikPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', ikBoot);
else ikBoot();
})();
