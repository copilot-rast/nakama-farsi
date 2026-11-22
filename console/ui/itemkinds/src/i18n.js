var IK_TEXT = {
  fa: {
    'group.itemkinds': 'دارایی‌های بازی',
    'nav.itemkinds/routes': 'مسیریاب یونیتی', 'nav.itemkinds/runs': 'گزارش مسیریابی', 'nav.itemkinds/settings': 'تنظیمات دارایی‌ها',
    'desc.itemkinds/kind': 'هر ردیف یک آیتم فروشگاه است با پروفایل این نوع. داده‌ها از یونیتی می‌آیند؛ صاحب هر فیلد (یونیتی یا پنل) در تنظیمات دارایی‌ها مشخص است.',
    'desc.itemkinds/routes': 'هر کارتی که یونیتی می‌فرستد، با این جدول به بخش درستش می‌رود. کارت «وابسته» آیتم نمی‌سازد؛ فقط کارت‌هایی را که به آن ارجاع می‌دهند دوباره می‌سازد.',
    'desc.itemkinds/runs': 'نتیجهٔ هر ارسال از یونیتی و هر مسیریابی دوباره: هر کارت به کدام بخش رفت، چه چیزی عوض شد و کدام در انتظار ماند.',
    'desc.itemkinds/settings': 'رفتار مسیریاب، صاحب هر فیلد (یونیتی یا پنل) و تنظیمات هر نوع.',
    'save': 'ذخیره', 'saved': 'ذخیره شد', 'cancel': 'انصراف', 'close': 'بستن', 'delete': 'حذف', 'search': 'جستجو', 'all': 'همه', 'add': 'افزودن', 'edit': 'ویرایش',
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'page.first': 'صفحهٔ اول', 'page.last': 'صفحهٔ آخر', 'page.of': 'صفحهٔ {page} از {pages}', 'pager': 'صفحه‌بندی', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'yes': 'بله', 'no': 'خیر',
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
    'loading': 'Loading…', 'empty': 'Nothing found', 'page.first': 'First page', 'page.last': 'Last page', 'page.of': 'Page {page} of {pages}', 'pager': 'Pagination', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'yes': 'Yes', 'no': 'No',
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
  if (window.NakamaDate) return window.NakamaDate.format(v);
  var d = new Date(v);
  if (isNaN(d)) return '';
  try {
    return d.toLocaleString(ikLang() === 'fa' ? 'fa-IR' : 'en-GB', { calendar: ikLang() === 'fa' ? 'persian' : 'gregory', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}
