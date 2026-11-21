(function () {
'use strict';
var AS_TEXT = {
  fa: {
    'group': 'داده و محتوای بازی',
    'nav.entries': 'داده‌های بازی',
    'nav.packs': 'بسته‌های محتوا',
    'nav.settings': 'تنظیمات داده و محتوا',
    'desc.entries': 'داده‌هایی که ابزار ویرایشگر (مثل ScriptableObjectها) می‌فرستد، با نسخه، تاریخچه و بازگردانی. بازی فقط تغییرات را می‌گیرد.',
    'desc.packs': 'بسته‌های محتوای ریموت (کاتالوگ و باندل‌های Addressables) برای هر پلتفرم. بازی آخرین بستهٔ فعالِ سازگار با نسخهٔ خودش را می‌گیرد.',
    'desc.settings': 'سقف اندازه و تعداد، رفتار تعارض نسخه‌ها، دسترسی بازی، نگه‌داری تاریخچه و بسته‌ها و آدرس دانلود محتوا.',
    'save': 'ذخیره', 'saved': 'ذخیره شد', 'cancel': 'انصراف', 'close': 'بستن', 'delete': 'حذف', 'search': 'جستجو', 'all': 'همه',
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'yes': 'بله', 'no': 'خیر',
    'confirm.title': 'مطمئن هستید؟', 'updated': 'آخرین تغییر: {by} · {at}', 'refresh': 'تازه‌سازی', 'copy': 'کپی', 'copied': 'کپی شد', 'none': '—',
    'f.assetId': 'شناسه', 'f.type': 'نوع', 'f.folder': 'پوشه', 'f.name': 'نام', 'f.version': 'نسخه', 'f.size': 'اندازه', 'f.updated': 'آخرین تغییر', 'f.data': 'داده (JSON)',
    'f.sourceGuid': 'فایل مبدأ (GUID)', 'f.hash': 'هش', 'f.recursive': 'با زیرپوشه‌ها', 'f.root': 'ریشه', 'f.allFolders': 'همهٔ پوشه‌ها', 'f.allTypes': 'همهٔ نوع‌ها',
    'kpi.entries': 'داده', 'kpi.size': 'حجم', 'kpi.folders': 'پوشه', 'kpi.types': 'نوع', 'kpi.packs': 'بستهٔ آماده', 'kpi.active': 'بستهٔ فعال',
    'en.new': 'دادهٔ تازه', 'en.import': 'بارگذاری JSON', 'en.export': 'خروجی JSON', 'en.tab.live': 'داده‌ها', 'en.tab.deleted': 'حذف‌شده‌ها',
    'en.v': 'نسخهٔ {v}', 'en.by': '{by}', 'en.edit': 'ویرایش {id}', 'en.create': 'دادهٔ تازه', 'en.delete': 'دادهٔ {id} حذف شود؟ از تاریخچه قابل بازگردانی است.',
    'en.badJson': 'JSON درست نیست؛ داده باید یک شیء JSON باشد.', 'en.history': 'تاریخچه', 'en.view': 'نمایش', 'en.restore': 'بازگردانی',
    'en.restore.confirm': 'نسخهٔ {v} به‌عنوان نسخهٔ تازه ذخیره شود؟', 'en.restored': 'بازگردانده شد (نسخهٔ {v})', 'en.state': 'نسخهٔ {v} · {action}',
    'en.conflict': 'این داده بعد از باز کردن شما تغییر کرده (اکنون نسخهٔ {v}). دوباره بارگذاری کنید و تغییرتان را دوباره اعمال کنید.', 'en.reload': 'بارگذاری دوباره',
    'en.deletedAt': 'حذف: {at} · {by}', 'en.noDeleted': 'دادهٔ حذف‌شده‌ای نیست.', 'en.folders': 'پوشه‌ها',
    'en.import.hint': 'همان قالب ابزار ویرایشگر: {"items":[{"assetId","type","folder","name","data","baseVersion"}],"deletes":[{"assetId","baseVersion"}]} یا فقط آرایهٔ items. اول «بررسی» را بزنید تا ببینید چه تغییری می‌کند.',
    'en.import.check': 'بررسی (بدون ذخیره)', 'en.import.apply': 'اعمال', 'en.import.force': 'بازنویسی حتی با نسخهٔ قدیمی (اگر در تنظیمات مجاز باشد)',
    'en.import.done': '{c} تازه، {u} تغییر، {s} بی‌تغییر، {d} حذف', 'en.import.conflicts': '{n} مورد تعارض نسخه دارد.',
    'act.created': 'تازه', 'act.updated': 'تغییر', 'act.unchanged': 'بی‌تغییر', 'act.deleted': 'حذف', 'act.restored': 'بازگردانی', 'act.conflict': 'تعارض', 'act.missing': 'نبود', 'act.failed': 'خطا', 'act.deactivated': 'غیرفعال',
    'pk.platform': 'پلتفرم', 'pk.allPlatforms': 'همهٔ پلتفرم‌ها', 'pk.version': 'نسخهٔ بسته', 'pk.range': 'نسخه‌های برنامه', 'pk.status': 'وضعیت', 'pk.files': 'فایل‌ها',
    'pk.created': 'بارگذاری', 'pk.activated': 'فعال‌سازی', 'pk.any': 'همه', 'pk.from': 'از {v}', 'pk.to': 'تا {v}', 'pk.fromTo': '{a} تا {b}',
    'pk.uploading': 'در حال بارگذاری', 'pk.ready': 'آماده', 'pk.active': 'فعال', 'pk.inactive': 'غیرفعال', 'pk.current': 'در حال پخش',
    'pk.activate': 'فعال کن', 'pk.deactivate': 'غیرفعال کن', 'pk.rollback': 'بازگشت به این بسته', 'pk.delete': 'بستهٔ {v} ({p}) حذف شود؟ فایل‌هایی که بستهٔ دیگری ندارد پاک می‌شوند.',
    'pk.activate.confirm': 'بستهٔ {v} برای {p} فعال شود؟ بازیکنان با نسخهٔ سازگار از همین حالا آن را دانلود می‌کنند.',
    'pk.deactivate.confirm': 'بستهٔ {v} غیرفعال شود؟ بازی سراغ بستهٔ فعال قبلی می‌رود.',
    'pk.rollback.confirm': 'بستهٔ {v} دوباره فعال شود تا جای بستهٔ فعلی پخش شود؟',
    'pk.resolver': 'بازی کدام بسته را می‌گیرد؟', 'pk.appVersion': 'نسخهٔ برنامه', 'pk.resolve': 'نشان بده', 'pk.resolved': 'بستهٔ {v} ({id})', 'pk.noPack': 'هیچ بستهٔ فعالی برای این پلتفرم و نسخه نیست.',
    'pk.urls': 'آدرس‌ها', 'pk.baseUrl': 'آدرس پایه (RemoteLoadPath)', 'pk.catalogUrl': 'کاتالوگ', 'pk.hashUrl': 'هش کاتالوگ', 'pk.catalogHash': 'مقدار هش',
    'pk.events': 'رویدادهای اخیر', 'pk.note': 'یادداشت', 'pk.notStored': 'نرسیده', 'pk.upload.hint': 'بسته‌ها از ابزار ویرایشگر (یا هر ابزاری با کلید API و دسترسی content:write) بارگذاری می‌شوند: POST /v2/assets/editor/packs، سپس تکه‌های فایل، سپس complete.',
    'ev.uploading': 'شروع بارگذاری', 'ev.ready': 'آماده شد', 'ev.activated': 'فعال شد', 'ev.deactivated': 'غیرفعال شد', 'ev.deleted': 'حذف شد', 'ev.removed': 'حذف خودکار (نگه‌داری)', 'ev.expired': 'بارگذاری نیمه‌کاره حذف شد',
    'se.entries': 'داده‌ها', 'se.maxEntryKb': 'بیشترین اندازهٔ هر داده (KB)', 'se.maxPushItems': 'بیشترین مورد در هر ارسال', 'se.maxFetchItems': 'بیشترین داده در هر دریافت بازی',
    'se.maxEntries': 'سقف تعداد داده‌ها (۰ = بی‌سقف)', 'se.conflictMode': 'وقتی یکی از موردهای ارسال نسخهٔ قدیمی دارد', 'se.conflict.reject': 'هیچ‌کدام ذخیره نشود (پیشنهادی)',
    'se.conflict.skip': 'بقیه ذخیره شوند و آن مورد گزارش شود', 'se.allowForce': 'ابزار بتواند با force نسخهٔ قدیمی را بازنویسی کند', 'se.access': 'دسترسی بازی به داده‌ها',
    'se.access.session': 'فقط با ورود بازیکن (RPC)', 'se.access.public': 'بدون ورود هم (HTTP عمومی با ETag)', 'se.cacheSeconds': 'زمان کش پاسخ عمومی (ثانیه)',
    'se.history': 'تاریخچه', 'se.keepVersions': 'نسخه‌های نگه‌داشته برای هر داده (۰ = همه)', 'se.keepDays': 'حذف نسخه‌های قدیمی‌تر از (روز، ۰ = بدون محدودیت سنی)',
    'se.keepDeletedDays': 'نگه‌داری داده‌های حذف‌شده برای بازگردانی (روز، ۰ = همیشه)',
    'se.packs': 'بسته‌های محتوا', 'se.platforms': 'پلتفرم‌های مجاز (با ویرگول؛ خالی = هر نامی)', 'se.partKb': 'اندازهٔ هر تکهٔ بارگذاری (KB)', 'se.partHint': 'کمتر از socket.max_request_size_bytes سرور باشد.',
    'se.maxFileMb': 'بیشترین اندازهٔ هر فایل (MB)', 'se.maxPackMb': 'بیشترین حجم هر بسته (MB)', 'se.maxFiles': 'بیشترین تعداد فایل هر بسته', 'se.uploadHours': 'مهلت تمام کردن بارگذاری (ساعت)',
    'se.keepInactive': 'بسته‌های غیرفعال نگه‌داشته برای هر پلتفرم (۰ = همه)', 'se.activateFromEditor': 'ابزار بتواند بسته را همان موقع فعال کند (با دسترسی content:publish)',
    'se.singleActive': 'با فعال کردن یک بسته، بسته‌های فعال دیگرِ همان بازهٔ نسخه غیرفعال شوند', 'se.publicIndex': 'فهرست محتوا بدون ورود بازیکن هم خوانده شود (Addressables پیش از ورود)',
    'se.indexCacheSeconds': 'زمان کش فهرست محتوا (ثانیه)', 'se.baseUrl': 'آدرس دانلود محتوا (CDN؛ خالی = آدرس خود سرور)', 'se.redirect': 'وقتی ذخیره‌ساز S3 آدرس عمومی دارد، دانلود مستقیم از آن باشد',
    'se.catalogPattern': 'الگوی نام فایل کاتالوگ (regex)', 'se.requireHash': 'بسته بدون فایل .hash کاتالوگ پذیرفته نشود',
    'se.info': 'اطلاعات', 'se.storage': 'ذخیره‌ساز فایل‌ها: {d}', 'se.scopes': 'دسترسی‌های کلید API: {s}', 'se.publicStorage': 'آدرس عمومی ذخیره‌ساز تنظیم شده است.',
    'code.ASSET_VERSION_CONFLICT': 'این داده بعد از دریافت شما تغییر کرده است.', 'code.ASSET_INVALID': 'دادهٔ فرستاده‌شده درست نیست.', 'code.ASSET_TOO_LARGE': 'داده از سقف اندازه بزرگ‌تر است.',
    'code.ASSET_EXISTS': 'داده‌ای با این شناسه هست.', 'code.ASSET_NOT_FOUND': 'داده پیدا نشد.', 'code.ASSET_ID_DUPLICATED': 'شناسه تکراری است.', 'code.ASSET_SOURCE_DUPLICATED': 'یک فایل با دو شناسه آمده است.',
    'code.ASSET_SOURCE_ID_CHANGED': 'شناسهٔ یک فایل عوض شده؛ اول دادهٔ قبلی را حذف کنید.', 'code.ASSET_PUSH_EMPTY': 'چیزی برای ارسال نیست.', 'code.ASSET_PUSH_TOO_MANY': 'تعداد موردها از سقف بیشتر است.',
    'code.ASSET_FORCE_NOT_ALLOWED': 'بازنویسی اجباری در تنظیمات بسته است.', 'code.ASSET_LIMIT_REACHED': 'سقف تعداد داده‌ها پر شده است.', 'code.INVALID_JSON': 'JSON درست نیست.',
    'code.PACK_ACTIVE': 'بستهٔ فعال را نمی‌توان حذف کرد.', 'code.PACK_NOT_READY': 'بارگذاری این بسته تمام نشده است.', 'code.PACK_NOT_FOUND': 'بسته پیدا نشد.',
    'code.VALUE_OUT_OF_RANGE': 'مقدار خارج از بازهٔ مجاز است.', 'code.INVALID_POLICY': 'یکی از انتخاب‌ها درست نیست.', 'code.INVALID_PLATFORM': 'نام پلتفرم درست نیست.', 'code.INVALID_URL': 'آدرس باید http یا https باشد.',
    'code.INVALID_PATTERN': 'الگوی regex درست نیست.', 'code.PLATFORM_REQUIRED': 'پلتفرم را انتخاب کنید.', 'code.PACK_PLATFORM_NOT_ALLOWED': 'این پلتفرم مجاز نیست.', 'code.PACK_INVALID': 'نسخه درست نیست.'
  },
  en: {
    'group': 'Game data & content',
    'nav.entries': 'Game data',
    'nav.packs': 'Content packs',
    'nav.settings': 'Data & content settings',
    'desc.entries': 'Data the editor tool sends (ScriptableObjects and the like) with versions, history and rollback. The game downloads only what changed.',
    'desc.packs': 'Remote content packs (Addressables catalog and bundles) per platform. The game takes the latest active pack that fits its app version.',
    'desc.settings': 'Size and count limits, version conflicts, game access, history and pack retention, and where content is downloaded from.',
    'save': 'Save', 'saved': 'Saved', 'cancel': 'Cancel', 'close': 'Close', 'delete': 'Delete', 'search': 'Search', 'all': 'All',
    'loading': 'Loading…', 'empty': 'Nothing found', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'yes': 'Yes', 'no': 'No',
    'confirm.title': 'Are you sure?', 'updated': 'Last change: {by} · {at}', 'refresh': 'Refresh', 'copy': 'Copy', 'copied': 'Copied', 'none': '—',
    'f.assetId': 'Id', 'f.type': 'Type', 'f.folder': 'Folder', 'f.name': 'Name', 'f.version': 'Version', 'f.size': 'Size', 'f.updated': 'Last change', 'f.data': 'Data (JSON)',
    'f.sourceGuid': 'Source file (GUID)', 'f.hash': 'Hash', 'f.recursive': 'Include sub-folders', 'f.root': 'Root', 'f.allFolders': 'All folders', 'f.allTypes': 'All types',
    'kpi.entries': 'Entries', 'kpi.size': 'Size', 'kpi.folders': 'Folders', 'kpi.types': 'Types', 'kpi.packs': 'Ready packs', 'kpi.active': 'Active packs',
    'en.new': 'New entry', 'en.import': 'Upload JSON', 'en.export': 'Export JSON', 'en.tab.live': 'Entries', 'en.tab.deleted': 'Deleted',
    'en.v': 'v{v}', 'en.by': '{by}', 'en.edit': 'Edit {id}', 'en.create': 'New entry', 'en.delete': 'Delete entry {id}? It can be restored from the history.',
    'en.badJson': 'Invalid JSON; the data must be a JSON object.', 'en.history': 'History', 'en.view': 'View', 'en.restore': 'Restore',
    'en.restore.confirm': 'Save version {v} as a new version?', 'en.restored': 'Restored (version {v})', 'en.state': 'Version {v} · {action}',
    'en.conflict': 'This entry changed after you opened it (now version {v}). Reload and apply your change again.', 'en.reload': 'Reload',
    'en.deletedAt': 'Deleted {at} · {by}', 'en.noDeleted': 'No deleted entries.', 'en.folders': 'Folders',
    'en.import.hint': 'The editor tool format: {"items":[{"assetId","type","folder","name","data","baseVersion"}],"deletes":[{"assetId","baseVersion"}]} or just the items array. Check first to see what changes.',
    'en.import.check': 'Check (no save)', 'en.import.apply': 'Apply', 'en.import.force': 'Overwrite even with an old version (when the settings allow it)',
    'en.import.done': '{c} new, {u} changed, {s} unchanged, {d} deleted', 'en.import.conflicts': '{n} items have a version conflict.',
    'act.created': 'new', 'act.updated': 'changed', 'act.unchanged': 'unchanged', 'act.deleted': 'deleted', 'act.restored': 'restored', 'act.conflict': 'conflict', 'act.missing': 'missing', 'act.failed': 'failed', 'act.deactivated': 'deactivated',
    'pk.platform': 'Platform', 'pk.allPlatforms': 'All platforms', 'pk.version': 'Pack version', 'pk.range': 'App versions', 'pk.status': 'Status', 'pk.files': 'Files',
    'pk.created': 'Uploaded', 'pk.activated': 'Activated', 'pk.any': 'any', 'pk.from': 'from {v}', 'pk.to': 'up to {v}', 'pk.fromTo': '{a} to {b}',
    'pk.uploading': 'uploading', 'pk.ready': 'ready', 'pk.active': 'active', 'pk.inactive': 'inactive', 'pk.current': 'serving',
    'pk.activate': 'Activate', 'pk.deactivate': 'Deactivate', 'pk.rollback': 'Roll back to this pack', 'pk.delete': 'Delete pack {v} ({p})? Files no other pack uses are removed.',
    'pk.activate.confirm': 'Activate pack {v} for {p}? Players with a matching app version download it from now on.',
    'pk.deactivate.confirm': 'Deactivate pack {v}? The game goes back to the previous active pack.',
    'pk.rollback.confirm': 'Activate pack {v} again so it is served instead of the current one?',
    'pk.resolver': 'Which pack does the game get?', 'pk.appVersion': 'App version', 'pk.resolve': 'Show', 'pk.resolved': 'Pack {v} ({id})', 'pk.noPack': 'No active pack for this platform and version.',
    'pk.urls': 'URLs', 'pk.baseUrl': 'Base URL (RemoteLoadPath)', 'pk.catalogUrl': 'Catalog', 'pk.hashUrl': 'Catalog hash', 'pk.catalogHash': 'Hash value',
    'pk.events': 'Recent events', 'pk.note': 'Note', 'pk.notStored': 'missing', 'pk.upload.hint': 'Packs are uploaded by the editor tool (or any tool with an API key holding content:write): POST /v2/assets/editor/packs, then the file parts, then complete.',
    'ev.uploading': 'upload started', 'ev.ready': 'ready', 'ev.activated': 'activated', 'ev.deactivated': 'deactivated', 'ev.deleted': 'deleted', 'ev.removed': 'removed (retention)', 'ev.expired': 'unfinished upload removed',
    'se.entries': 'Entries', 'se.maxEntryKb': 'Largest entry (KB)', 'se.maxPushItems': 'Most items per push', 'se.maxFetchItems': 'Most entries per game fetch',
    'se.maxEntries': 'Entry limit (0 = none)', 'se.conflictMode': 'When an item of a push has an old version', 'se.conflict.reject': 'save nothing (recommended)',
    'se.conflict.skip': 'save the others and report it', 'se.allowForce': 'Tools may overwrite old versions with force', 'se.access': 'Game access to entries',
    'se.access.session': 'with a player session only (RPC)', 'se.access.public': 'also without a session (public HTTP with ETag)', 'se.cacheSeconds': 'Public answer cache (seconds)',
    'se.history': 'History', 'se.keepVersions': 'Versions kept per entry (0 = all)', 'se.keepDays': 'Remove versions older than (days, 0 = no age limit)',
    'se.keepDeletedDays': 'Keep deleted entries for restore (days, 0 = forever)',
    'se.packs': 'Content packs', 'se.platforms': 'Allowed platforms (comma separated; empty = any name)', 'se.partKb': 'Upload part size (KB)', 'se.partHint': 'Keep it below the server\'s socket.max_request_size_bytes.',
    'se.maxFileMb': 'Largest file (MB)', 'se.maxPackMb': 'Largest pack (MB)', 'se.maxFiles': 'Most files per pack', 'se.uploadHours': 'Time to finish an upload (hours)',
    'se.keepInactive': 'Inactive packs kept per platform (0 = all)', 'se.activateFromEditor': 'Tools may activate a pack right away (with content:publish)',
    'se.singleActive': 'Activating a pack deactivates other active packs of the same app range', 'se.publicIndex': 'The content index is readable without a session (Addressables before login)',
    'se.indexCacheSeconds': 'Content index cache (seconds)', 'se.baseUrl': 'Content download URL (CDN; empty = this server)', 'se.redirect': 'Download straight from S3 when the store has a public URL',
    'se.catalogPattern': 'Catalog file name pattern (regex)', 'se.requireHash': 'Refuse packs without the catalog .hash file',
    'se.info': 'Info', 'se.storage': 'File storage: {d}', 'se.scopes': 'API key scopes: {s}', 'se.publicStorage': 'The storage has a public URL.',
    'code.ASSET_VERSION_CONFLICT': 'The entry changed after you read it.', 'code.ASSET_INVALID': 'The entry is not valid.', 'code.ASSET_TOO_LARGE': 'The entry is over the size limit.',
    'code.ASSET_EXISTS': 'An entry with this id exists.', 'code.ASSET_NOT_FOUND': 'Entry not found.', 'code.ASSET_ID_DUPLICATED': 'Duplicated id.', 'code.ASSET_SOURCE_DUPLICATED': 'One file came under two ids.',
    'code.ASSET_SOURCE_ID_CHANGED': 'A file changed its id; delete the old entry first.', 'code.ASSET_PUSH_EMPTY': 'Nothing to push.', 'code.ASSET_PUSH_TOO_MANY': 'Too many items.',
    'code.ASSET_FORCE_NOT_ALLOWED': 'Forced overwrites are off in the settings.', 'code.ASSET_LIMIT_REACHED': 'The entry limit is reached.', 'code.INVALID_JSON': 'Invalid JSON.',
    'code.PACK_ACTIVE': 'An active pack cannot be deleted.', 'code.PACK_NOT_READY': 'This pack has not finished uploading.', 'code.PACK_NOT_FOUND': 'Pack not found.',
    'code.VALUE_OUT_OF_RANGE': 'The value is out of range.', 'code.INVALID_POLICY': 'One of the choices is not valid.', 'code.INVALID_PLATFORM': 'Invalid platform name.', 'code.INVALID_URL': 'The URL must be http or https.',
    'code.INVALID_PATTERN': 'Invalid regex.', 'code.PLATFORM_REQUIRED': 'Pick a platform.', 'code.PACK_PLATFORM_NOT_ALLOWED': 'This platform is not allowed.', 'code.PACK_INVALID': 'Invalid version.'
  }
};

function asLang() {
  var nc = window.NakamaConsole;
  var l = nc && nc.lang ? nc.lang() : (document.documentElement.getAttribute('lang') || 'fa');
  return /^en/i.test(l) ? 'en' : 'fa';
}

function has(key) {
  return (AS_TEXT[asLang()] || AS_TEXT.fa)[key] != null;
}

function t(key, vars) {
  var dict = AS_TEXT[asLang()] || AS_TEXT.fa;
  var s = dict[key] != null ? dict[key] : (AS_TEXT.en[key] != null ? AS_TEXT.en[key] : key);
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] != null ? digits(String(vars[k])) : m; });
  return s;
}

function digits(s) {
  if (asLang() !== 'fa') return s;
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
    return d.toLocaleString(asLang() === 'fa' ? 'fa-IR' : 'en-GB', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

// tr is t() with raw values (keys) that keep their Latin digits.
function tr(key, vars, raw) {
  var s = t(key, vars);
  Object.keys(raw || {}).forEach(function (k) { s = s.split('{' + k + '}').join(String(raw[k])); });
  return s;
}

var AS_API = '/v2/console/base';
var AS_INPUT = 'appearance-none block w-full placeholder-secondary-400 dark:placeholder-secondary-500 border-secondary-300 dark:border-secondary-700 border shadow-sm rounded-md bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200 px-3 py-2 text-sm';
var AS_BTN = 'as-btn inline-flex items-center justify-center gap-1.5 font-medium whitespace-nowrap rounded-md border text-sm py-1.5 px-3';

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

function AsError(status, code, message, field) {
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
  return fetch((path.indexOf('/v2/') === 0 ? '' : AS_API) + path, init).then(function (res) {
    if (res.status === 401 && !/^#\/login/.test(location.hash)) location.hash = '#/login';
    return res.text().then(function (text) {
      var data = null;
      if (text) { try { data = JSON.parse(text); } catch (e) { data = null; } }
      if (!res.ok) { var e = new AsError(res.status, data && typeof data.error === 'string' ? data.error : 'HTTP_' + res.status, data && data.message || '', data && data.field || ''); e.meta = data && data.meta || {}; throw e; }
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
  var b = h('button', { type: opts.type || 'button', class: AS_BTN + ' as-btn--' + (opts.variant || 'ghost') + (opts.class ? ' ' + opts.class : ''), disabled: !!opts.disabled, title: opts.title, 'aria-label': opts.title, onclick: opts.onclick },
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

var asToastBox = null;
function toast(kind, text) {
  if (!asToastBox || !asToastBox.isConnected) {
    asToastBox = h('div', { class: 'eco-toasts', 'data-l10n-skip': '', 'data-assets': 'toasts' });
    document.body.appendChild(asToastBox);
  }
  var el = notice(kind, text);
  el.classList.add('shadow-lg');
  asToastBox.appendChild(el);
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
  var err = h('p', { class: 'as-field-error text-xs text-error-600 dark:text-error-400', hidden: true });
  var wrap = h('label', { class: 'as-field block mb-4' + (extra ? ' ' + extra : '') },
    label ? h('span', { class: 'block font-medium text-sm text-secondary-800 dark:text-secondary-200 mb-1', text: label }) : null,
    control,
    hint ? h('span', { class: 'block text-xs text-secondary-500 dark:text-secondary-400 mt-1', text: hint }) : null,
    err);
  wrap.setError = function (msg) { err.textContent = msg || ''; err.hidden = !msg; };
  return wrap;
}

function input(value, opts) {
  opts = opts || {};
  var el = h(opts.multiline ? 'textarea' : 'input', { class: AS_INPUT + (opts.mono ? ' font-mono' : ''), type: opts.multiline ? null : (opts.type || 'text'), dir: opts.ltr ? 'ltr' : 'auto', rows: opts.multiline ? (opts.rows || 3) : null, placeholder: opts.placeholder, name: opts.name, autocomplete: 'off', 'aria-label': opts.label });
  el.value = value == null ? '' : value;
  return el;
}

function select(value, options, opts) {
  opts = opts || {};
  var el = h('select', { class: AS_INPUT + ' eco-select', name: opts.name, 'aria-label': opts.label });
  options.forEach(function (o) {
    var opt = h('option', { value: o.value, text: o.label });
    if (String(o.value) === String(value == null ? '' : value)) opt.selected = true;
    el.appendChild(opt);
  });
  return el;
}

function check(checked, label, name) {
  var box = h('input', { type: 'checkbox', class: 'as-check', name: name, checked: !!checked });
  var wrap = h('label', { class: 'as-checkline flex items-start gap-2 mb-3 cursor-pointer text-sm text-secondary-800 dark:text-secondary-200' }, box, h('span', { text: label }));
  wrap.input = box;
  return wrap;
}

function numberValue(el) {
  var raw = String(el.value || '').replace(/[۰-۹]/g, function (d) { return String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)); }).trim();
  if (raw === '') return null;
  var n = Number(raw);
  return isFinite(n) ? Math.round(n) : NaN;
}

var asPanel = null;
function closePanel() {
  if (asPanel) { asPanel.remove(); asPanel = null; document.removeEventListener('keydown', panelKey); }
}

function panelKey(e) { if (e.key === 'Escape') closePanel(); }

function sidePanel(title, body, footer, name) {
  closePanel();
  var holder = h('div', { class: 'as-panel-holder' });
  var panel = h('div', { class: 'as-panel bg-white dark:bg-secondary-900 shadow-lg', role: 'dialog', 'aria-modal': 'true', 'aria-label': title, 'data-assets-panel': name || '' },
    h('div', { class: 'as-panel-head flex items-center justify-between border-b px-6 py-3' },
      h('h2', { class: 'text-lg font-semibold truncate', text: title }),
      button({ label: '×', title: t('close'), onclick: closePanel, class: 'as-close' })),
    h('div', { class: 'as-panel-body p-6' }, body),
    footer ? h('div', { class: 'as-panel-foot flex gap-2 justify-end bg-secondary-50 dark:bg-secondary-800 p-4' }, footer) : null);
  holder.appendChild(panel);
  holder.addEventListener('mousedown', function (e) { if (e.target === holder) closePanel(); });
  asPanel = h('div', { 'data-assets': 'panel', 'data-l10n-skip': '', dir: document.documentElement.getAttribute('dir') || 'ltr' }, h('div', { class: 'as-shade' }), holder);
  document.body.appendChild(asPanel);
  document.addEventListener('keydown', panelKey);
  return panel;
}

function confirmBox(text, okLabel) {
  return new Promise(function (resolve) {
    var root;
    function done(v) { root.remove(); resolve(v); }
    root = h('div', { 'data-assets': 'confirm', 'data-l10n-skip': '', class: 'as-confirm', dir: document.documentElement.getAttribute('dir') || 'ltr' },
      h('div', { class: 'as-confirm-box bg-white dark:bg-secondary-800 rounded-md shadow-lg p-5', role: 'alertdialog', 'aria-modal': 'true' },
        h('h3', { class: 'font-semibold mb-2', text: t('confirm.title') }),
        h('p', { class: 'text-sm mb-4', text: text }),
        h('div', { class: 'flex gap-2 justify-end' },
          button({ label: t('cancel'), onclick: function () { done(false); } }),
          button({ label: okLabel || t('yes'), variant: 'danger', data: { 'assets-confirm': 'ok' }, onclick: function () { done(true); } }))));
    document.body.appendChild(root);
  });
}

function pageFrame(page, desc, actions) {
  var content = h('div', { class: 'flex flex-col gap-4' });
  var el = h('div', { class: 'as-page flex flex-col flex-1 px-4 md:px-6 pb-6 max-w-[1600px] w-full mx-auto', 'data-assets-page': page },
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
    var tr = h('tr', { class: 'border-b border-secondary-200 dark:border-secondary-700 hover:bg-secondary-50 dark:hover:bg-secondary-700' + (onrow ? ' cursor-pointer' : ''), 'data-assets-row': row.id || '' });
    if (onrow) {
      tr.tabIndex = 0;
      tr.addEventListener('click', function (e) { if (!e.target.closest('a,button,input,select,label')) onrow(row.item); });
      tr.addEventListener('keydown', function (e) { if (e.key === 'Enter') onrow(row.item); });
    }
    row.cells.forEach(function (cell, i) { tr.appendChild(h('td', { class: 'px-3 py-2 align-middle text-sm', 'data-label': columns[i].label }, cell)); });
    tbody.appendChild(tr);
  });
  return h('div', { class: 'bg-white dark:bg-secondary-800 shadow rounded-md overflow-x-auto' }, h('table', { class: 'w-full eco-table as-table' }, thead, tbody));
}

function pager(state, onpage) {
  var total = state.total || 0;
  var from = total ? (state.page - 1) * state.limit + 1 : 0;
  var to = Math.min(total, state.page * state.limit);
  var last = Math.max(1, Math.ceil(total / state.limit));
  return h('div', { class: 'flex items-center justify-end gap-2 text-xs text-secondary-500' },
    h('span', { 'data-assets-range': '', text: total ? t('range', { from: from, to: to, total: total }) : t('empty') }),
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

var asLanguages = null;
function loadLanguages() {
  if (asLanguages) return Promise.resolve(asLanguages);
  var headers = { Accept: 'application/json' };
  var tk = token();
  if (tk) headers.Authorization = 'Bearer ' + tk;
  return fetch('/v2/console/base/languages', { headers: headers, credentials: 'same-origin' }).then(function (r) { return r.ok ? r.json() : null; }).then(function (d) {
    var list = d && d.items ? d.items.filter(function (l) { return l.isActive; }).map(function (l) { return { code: l.code, name: l.nativeName || l.code, isDefault: !!l.isDefault, direction: l.direction }; }) : [];
    if (!list.length) list = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    list.sort(function (a, b) { return (b.isDefault ? 1 : 0) - (a.isDefault ? 1 : 0); });
    asLanguages = list;
    return list;
  }, function () {
    asLanguages = [{ code: 'fa', name: 'فارسی', isDefault: true, direction: 'rtl' }, { code: 'en', name: 'English', direction: 'ltr' }];
    return asLanguages;
  });
}

// translated renders one input per active language; .read() returns {lang: text}.
function translated(label, name, value, opts) {
  opts = opts || {};
  var inputs = {};
  var box = h('div', { class: 'grid sm:grid-cols-2 gap-x-4', 'data-assets-text': name });
  (asLanguages || []).forEach(function (l) {
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
  var l = asLang();
  return m[l] || m.fa || m.en || Object.keys(m).map(function (k) { return m[k]; }).filter(Boolean)[0] || fallback || '';
}

function fieldError(msgBox, err) {
  clear(msgBox).appendChild(notice('error', errorText(err) + (err && err.field ? ' · ' + err.field : '')));
}

function options(list, value) {
  return list.map(function (o) { return typeof o === 'string' ? { value: o, label: o } : o; });
}

function dateInput(value, name, type) {
  var el = h('input', { type: type || 'date', class: AS_INPUT, name: name, dir: 'ltr', 'aria-label': name });
  el.value = value || '';
  return el;
}

function isoDay(d) {
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
}

function asErrText(err) {
  if (err && err.code && has('code.' + err.code)) return t('code.' + err.code) + (err.field ? ' · ' + err.field : '');
  return errorText(err) + (err && err.field ? ' · ' + err.field : '');
}

function section(title, children, name) {
  return card(h('div', { class: 'p-4', 'data-assets-section': name || '' },
    h('h2', { class: 'font-semibold mb-3', text: title }), children));
}

// settingsPage loads a plugin's settings once, lets build add its sections and saves everything with one button.
function settingsPage(root, page, base, build) {
  var frame = pageFrame(page, t('desc.' + page));
  clear(root).appendChild(frame.el);
  frame.content.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  Promise.all([loadSettings(base), loadLanguages()]).then(function (res) {
    var d = res[0];
    var cur = JSON.parse(JSON.stringify(d.settings));
    var form = build(cur, d.defaults, d);
    var msg = h('div', { 'data-assets-msg': '' });
    var meta = h('p', { class: 'text-xs text-secondary-500', 'data-assets-updated': '' });
    var paintMeta = function (x) { meta.textContent = x.updatedBy ? t('updated', { by: x.updatedBy, at: fmtDate(x.updatedAt) }) : ''; };
    paintMeta(d);
    var save = button({ label: t('save'), variant: 'primary', data: { 'assets-save': page }, onclick: function () {
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
        msg.appendChild(notice('error', asErrText(err)));
      });
    } });
    clear(frame.content);
    form.el.forEach(function (el) { frame.content.appendChild(el); });
    frame.content.appendChild(h('div', { class: 'as-savebar flex flex-wrap items-center justify-between gap-3' }, meta, h('div', { class: 'flex-1' }, msg), save));
  }, function (err) {
    clear(frame.content).appendChild(notice('error', asErrText(err)));
  });
}

// numInput accepts Persian and Latin digits (type=number refuses Persian ones).
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
  return h('a', { href: '#/players/' + id, class: 'eco-link', 'data-assets-player-link': id, dir: 'auto', text: name || id.slice(0, 8) });
}


function fmtBytes(n) {
  n = Number(n || 0);
  if (n < 1024) return fmtNum(n) + ' B';
  if (n < 1024 * 1024) return fmtNum(Math.round(n / 102.4) / 10) + ' KB';
  if (n < 1024 * 1024 * 1024) return fmtNum(Math.round(n / 104857.6) / 10) + ' MB';
  return fmtNum(Math.round(n / 107374182.4) / 10) + ' GB';
}

function copyButton(text, name) {
  return button({ label: t('copy'), data: { 'assets-copy': name || '' }, onclick: function () {
    try { navigator.clipboard.writeText(text).then(function () { toast('success', t('copied')); }); } catch (e) { toast('error', String(e)); }
  } });
}

function kpis(list) {
  return h('div', { class: 'grid grid-cols-2 sm:grid-cols-4 gap-3', 'data-assets-kpis': '' }, list.map(function (k) {
    return card(h('div', { class: 'p-3' }, h('div', { class: 'text-xs text-secondary-500', text: k.label }), h('div', { class: 'text-lg font-semibold', 'data-assets-kpi': k.name, dir: 'auto', text: k.value })));
  }));
}

function actionTag(action) {
  var color = { created: 'success', updated: 'primary', restored: 'primary', deleted: 'error', conflict: 'warning', failed: 'error', deactivated: 'warning', missing: 'secondary', unchanged: 'secondary' }[action] || 'secondary';
  return tag(has('act.' + action) ? t('act.' + action) : action, color);
}

// Game data entries: folder list, search, entry editor with history and rollback, deleted entries, bulk JSON upload.
var asEntryState = { folder: '', recursive: true, type: '', q: '', page: 1, limit: 50, tab: 'live' };

function entriesPage(root) {
  var frame = pageFrame('entries', t('desc.entries'), [
    button({ label: t('en.import'), data: { 'assets-action': 'import' }, onclick: function () { importPanel(function () { entriesPage(root); }); } }),
    button({ label: t('en.export'), data: { 'assets-action': 'export' }, onclick: exportEntries }),
    button({ label: t('en.new'), variant: 'primary', data: { 'assets-action': 'new-entry' }, onclick: function () { entryPanel(null, function () { entriesPage(root); }); } })
  ]);
  clear(root).appendChild(frame.el);
  var top = h('div', { class: 'flex flex-col gap-4' });
  var body = h('div', { class: 'flex flex-col gap-4' });
  frame.content.appendChild(top);
  frame.content.appendChild(body);
  top.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  api('GET', '/assets/overview').then(function (ov) {
    clear(top);
    top.appendChild(kpis([
      { name: 'entries', label: t('kpi.entries'), value: fmtNum(ov.total) },
      { name: 'size', label: t('kpi.size'), value: fmtBytes(ov.totalBytes) },
      { name: 'folders', label: t('kpi.folders'), value: fmtNum((ov.folders || []).length) },
      { name: 'types', label: t('kpi.types'), value: fmtNum((ov.types || []).length) }
    ]));
    var tabs = h('div', { class: 'flex gap-2', role: 'tablist' }, ['live', 'deleted'].map(function (tb) {
      return button({ label: t('en.tab.' + tb), variant: asEntryState.tab === tb ? 'primary' : 'ghost', data: { 'assets-tab': tb }, onclick: function () { asEntryState.tab = tb; entriesPage(root); } });
    }));
    top.appendChild(tabs);
    if (asEntryState.tab === 'deleted') { deletedList(body, root); return; }
    var folders = folderOptions(ov.folders || []);
    var folderSel = select(asEntryState.folder, [{ value: '', label: t('f.allFolders') }].concat(folders), { name: 'folder', label: t('f.folder') });
    var typeSel = select(asEntryState.type, [{ value: '', label: t('f.allTypes') }].concat((ov.types || []).map(function (x) { return { value: x.key, label: x.key + ' (' + digits(String(x.count)) + ')' }; })), { name: 'type', label: t('f.type') });
    var q = input(asEntryState.q, { name: 'q', placeholder: t('search'), label: t('search') });
    var rec = check(asEntryState.recursive, t('f.recursive'), 'recursive');
    var reload = function () { asEntryState.folder = folderSel.value; asEntryState.type = typeSel.value; asEntryState.q = q.value.trim(); asEntryState.recursive = rec.input.checked; asEntryState.page = 1; loadList(); };
    folderSel.addEventListener('change', reload);
    typeSel.addEventListener('change', reload);
    rec.input.addEventListener('change', reload);
    var timer = null;
    q.addEventListener('input', function () { clearTimeout(timer); timer = setTimeout(reload, 300); });
    body.appendChild(h('div', { class: 'as-filters flex flex-wrap gap-3 items-end' }, field(t('f.folder'), folderSel), field(t('f.type'), typeSel), field(t('search'), q), rec));
    var list = h('div', { 'data-assets-list': 'entries' });
    body.appendChild(list);
    function loadList() {
      clear(list).appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
      var s = asEntryState;
      api('GET', '/assets/entries' + qs({ folder: s.folder, recursive: (s.folder === '/' || !s.recursive) ? 'false' : '', type: s.type, q: s.q, page: s.page, limit: s.limit })).then(function (d) {
        clear(list);
        if (!d.items.length) { list.appendChild(notice('secondary', t('empty'))); return; }
        list.appendChild(table([
          { label: t('f.assetId') }, { label: t('f.name') }, { label: t('f.type') }, { label: t('f.folder') }, { label: t('f.version') }, { label: t('f.size') }, { label: t('f.updated') }
        ], d.items.map(function (e) {
          return { id: e.assetId, item: e, cells: [mono(e.assetId), h('span', { dir: 'auto', text: e.name }), mono(e.type), mono(e.folder || '/'), tag(t('en.v', { v: e.version }), 'primary'), fmtBytes(e.size),
            h('div', {}, h('div', { text: fmtDate(e.updatedAt) }), h('div', { class: 'text-xs text-secondary-500', dir: 'auto', text: e.updatedBy }))] };
        }), function (e) { entryPanel(e.assetId, loadList); }));
        list.appendChild(h('div', { class: 'mt-2' }, pager({ page: d.page, limit: d.limit, total: d.total }, function (p) { asEntryState.page = p; loadList(); })));
      }, function (err) { clear(list).appendChild(notice('error', asErrText(err))); });
    }
    loadList();
  }, function (err) { clear(top).appendChild(notice('error', asErrText(err))); });
}

// folderOptions lists every folder and its parents with entry counts (sub-folders included).
function folderOptions(rows) {
  var counts = {};
  rows.forEach(function (r) {
    var parts = r.key ? r.key.split('/') : [];
    if (!parts.length) { counts[''] = (counts[''] || 0) + r.count; return; }
    for (var i = 1; i <= parts.length; i++) { var p = parts.slice(0, i).join('/'); counts[p] = (counts[p] || 0) + r.count; }
  });
  return Object.keys(counts).filter(function (k) { return k !== ''; }).sort().map(function (k) {
    var depth = k.split('/').length - 1;
    return { value: k, label: new Array(depth + 1).join('  ') + k.split('/').pop() + ' (' + digits(String(counts[k])) + ')' };
  }).concat(counts[''] ? [{ value: '/', label: t('f.root') + ' (' + digits(String(counts[''])) + ')' }] : []);
}

function prettyJSON(v) {
  try { return JSON.stringify(v, null, 2); } catch (e) { return String(v); }
}

function entryPanel(id, done) {
  if (!id) { entryForm(null, [], done); return; }
  api('GET', '/assets/entries/' + encodeURIComponent(id)).then(function (d) { entryForm(d.item, d.history || [], done); }, function (err) { toast('error', asErrText(err)); });
}

function entryForm(item, history, done) {
  var isNew = !item;
  item = item || { assetId: '', type: '', folder: '', name: '', data: {}, version: 0 };
  var idIn = input(item.assetId, { name: 'assetId', ltr: true, mono: true, label: t('f.assetId') });
  if (!isNew) idIn.setAttribute('readonly', '');
  var typeIn = input(item.type, { name: 'type', ltr: true, mono: true, label: t('f.type') });
  var folderIn = input(item.folder, { name: 'folder', ltr: true, mono: true, label: t('f.folder') });
  var nameIn = input(item.name, { name: 'name', label: t('f.name') });
  var dataIn = input(prettyJSON(item.data), { name: 'data', multiline: true, rows: 14, ltr: true, mono: true, label: t('f.data') });
  var msg = h('div', { 'data-assets-msg': '' });
  var save = button({ label: t('save'), variant: 'primary', data: { 'assets-action': 'save-entry' }, onclick: function () {
    clear(msg);
    var data;
    try { data = JSON.parse(dataIn.value); } catch (e) { data = null; }
    if (!data || typeof data !== 'object' || Array.isArray(data)) { msg.appendChild(notice('error', t('en.badJson'))); return; }
    var body = { assetId: idIn.value.trim(), type: typeIn.value.trim(), folder: folderIn.value.trim(), name: nameIn.value.trim(), data: data };
    if (!isNew) body.baseVersion = item.version;
    busy(save, true);
    api(isNew ? 'POST' : 'PUT', isNew ? '/assets/entries' : '/assets/entries/' + encodeURIComponent(item.assetId), body).then(function () {
      busy(save, false); toast('success', t('saved')); closePanel(); done();
    }, function (err) {
      busy(save, false);
      if (err.code === 'ASSET_VERSION_CONFLICT') { msg.appendChild(conflictNotice(err, function () { entryPanel(item.assetId, done); })); return; }
      fieldError(msg, err);
    });
  } });
  var del = isNew ? null : button({ label: t('delete'), variant: 'danger', data: { 'assets-action': 'delete-entry' }, onclick: function () {
    confirmBox(tr('en.delete', {}, { id: item.assetId }), t('delete')).then(function (ok) {
      if (!ok) return;
      api('DELETE', '/assets/entries/' + encodeURIComponent(item.assetId) + '?baseVersion=' + item.version).then(function () { toast('success', t('saved')); closePanel(); done(); }, function (err) {
        if (err.code === 'ASSET_VERSION_CONFLICT') { clear(msg).appendChild(conflictNotice(err, function () { entryPanel(item.assetId, done); })); return; }
        fieldError(msg, err);
      });
    });
  } });
  var meta = isNew ? null : h('div', { class: 'flex flex-wrap gap-2 items-center mb-4 text-xs text-secondary-500' },
    tag(t('en.v', { v: item.version }), 'primary'), mono(item.hash ? item.hash.slice(0, 16) + '…' : ''), item.sourceGuid ? mono(item.sourceGuid) : null,
    h('span', { dir: 'auto', text: fmtDate(item.updatedAt) + ' · ' + (item.updatedBy || '') }));
  var hist = isNew ? null : historyBox(item, history, done);
  var content = h('div', {}, meta,
    h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, field(t('f.assetId'), idIn), field(t('f.type'), typeIn), field(t('f.folder'), folderIn), field(t('f.name'), nameIn)),
    field(t('f.data'), dataIn), msg, hist);
  sidePanel(isNew ? t('en.create') : tr('en.edit', {}, { id: item.assetId }), content, [del, h('div', { class: 'flex-1' }), button({ label: t('cancel'), onclick: closePanel }), save], 'entry');
}

function conflictNotice(err, reload) {
  var c = (err.meta && err.meta.conflicts || [])[0] || {};
  var box = notice('warning', t('en.conflict', { v: c.version == null ? '—' : c.version }));
  box.setAttribute('data-assets-conflict', '');
  box.appendChild(button({ label: t('en.reload'), data: { 'assets-action': 'reload-entry' }, onclick: reload }));
  return box;
}

function historyBox(item, history, done) {
  var view = h('div', {});
  var rows = history.map(function (hs) {
    return h('li', { class: 'flex flex-wrap items-center gap-2 py-2 border-b border-secondary-200 dark:border-secondary-700', 'data-assets-history': hs.version },
      tag(t('en.v', { v: hs.version }), 'secondary'), actionTag(hs.action),
      h('span', { class: 'text-xs text-secondary-500 flex-1', dir: 'auto', text: fmtDate(hs.createdAt) + ' · ' + hs.actor }),
      button({ label: t('en.view'), data: { 'assets-action': 'view-version' }, onclick: function () {
        api('GET', '/assets/history/' + hs.id).then(function (d) {
          clear(view).appendChild(h('div', { class: 'mt-3' }, h('div', { class: 'text-sm font-medium mb-1', text: t('en.state', { action: t('act.' + d.item.action), v: d.item.version }) }),
            h('pre', { class: 'as-pre', dir: 'ltr', 'data-assets-state': d.item.version, text: prettyJSON(d.item.data) })));
        }, function (err) { toast('error', asErrText(err)); });
      } }),
      hs.version === item.version && hs.action !== 'deleted' ? null : button({ label: t('en.restore'), data: { 'assets-action': 'restore-version' }, onclick: function () {
        confirmBox(t('en.restore.confirm', { v: hs.version }), t('en.restore')).then(function (ok) {
          if (!ok) return;
          api('POST', '/assets/entries/' + encodeURIComponent(item.assetId) + '/rollback', { historyId: hs.id, baseVersion: item.version }).then(function (d) {
            toast('success', t('en.restored', { v: d.item.version })); closePanel(); done();
          }, function (err) { toast('error', asErrText(err)); });
        });
      } }));
  });
  return section(t('en.history'), h('div', {}, h('ul', { 'data-assets-history-list': '' }, rows), view), 'history');
}

function deletedList(body, root) {
  body.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  api('GET', '/assets/deleted').then(function (d) {
    clear(body);
    if (!d.items.length) { body.appendChild(notice('secondary', t('en.noDeleted'))); return; }
    body.appendChild(table([{ label: t('f.assetId') }, { label: t('f.type') }, { label: t('f.folder') }, { label: t('f.version') }, { label: t('f.updated') }, { label: '' }],
      d.items.map(function (hs) {
        return { id: hs.assetId, item: hs, cells: [mono(hs.assetId), mono(hs.type), mono(hs.folder || '/'), tag(t('en.v', { v: hs.version }), 'secondary'),
          h('span', { dir: 'auto', text: fmtDate(hs.createdAt) + ' · ' + hs.actor }),
          button({ label: t('en.restore'), data: { 'assets-action': 'restore-deleted' }, onclick: function () {
            api('POST', '/assets/entries/' + encodeURIComponent(hs.assetId) + '/rollback', { historyId: hs.id }).then(function (r) {
              toast('success', t('en.restored', { v: r.item.version })); entriesPage(root);
            }, function (err) { toast('error', asErrText(err)); });
          } })] };
      })));
  }, function (err) { clear(body).appendChild(notice('error', asErrText(err))); });
}

function exportEntries() {
  var s = asEntryState;
  api('GET', '/assets/export' + qs({ folder: s.folder, recursive: (s.folder === '/' || !s.recursive) ? 'false' : '' })).then(function (d) {
    var blob = new Blob([JSON.stringify(d, null, 2)], { type: 'application/json' });
    var a = h('a', { href: URL.createObjectURL(blob), download: 'assets-export.json' });
    document.body.appendChild(a); a.click(); a.remove();
  }, function (err) { toast('error', asErrText(err)); });
}

function importPanel(done) {
  var area = input('', { multiline: true, rows: 12, ltr: true, mono: true, name: 'importJson', label: t('en.import') });
  area.setAttribute('placeholder', '{"items":[{"assetId":"Tiggo8","type":"VehicleCard","folder":"Cars","name":"Tiggo 8","data":{"speed":180}}]}');
  var force = check(false, t('en.import.force'), 'force');
  var msg = h('div', { 'data-assets-msg': '' });
  var report = h('div', { 'data-assets-report': '' });
  var parse = function () {
    var body;
    try { body = JSON.parse(area.value); } catch (e) { return null; }
    if (Array.isArray(body)) body = { items: body };
    return body && typeof body === 'object' ? body : null;
  };
  var run = function (dry, btn) {
    clear(msg); clear(report);
    var body = parse();
    if (!body) { msg.appendChild(notice('error', t('en.badJson'))); return; }
    body.dryRun = dry;
    body.force = force.input.checked;
    busy(btn, true);
    api('POST', '/assets/push', body).then(function (r) {
      busy(btn, false);
      report.appendChild(notice(r.conflicts.length ? 'warning' : 'success', t('en.import.done', { c: r.created, u: r.updated, s: r.unchanged, d: r.deleted }) + (r.conflicts.length ? ' · ' + t('en.import.conflicts', { n: r.conflicts.length }) : '')));
      report.appendChild(h('div', { class: 'mt-2' }, table([{ label: t('f.assetId') }, { label: '' }, { label: t('f.version') }], r.rows.map(function (x) {
        return { id: x.assetId, cells: [mono(x.assetId), actionTag(x.action), x.version == null ? '—' : t('en.v', { v: x.version })] };
      }))));
      if (!dry) { toast('success', t('saved')); done(); }
    }, function (err) {
      busy(btn, false);
      if (err.code === 'ASSET_VERSION_CONFLICT') {
        var list = (err.meta && err.meta.conflicts || []).map(function (c) { return c.assetId; }).join(', ');
        msg.appendChild(notice('warning', t('code.ASSET_VERSION_CONFLICT') + ' ' + list));
        return;
      }
      fieldError(msg, err);
    });
  };
  var checkBtn = button({ label: t('en.import.check'), data: { 'assets-action': 'import-check' }, onclick: function () { run(true, checkBtn); } });
  var applyBtn = button({ label: t('en.import.apply'), variant: 'primary', data: { 'assets-action': 'import-apply' }, onclick: function () { run(false, applyBtn); } });
  sidePanel(t('en.import'), h('div', {}, h('p', { class: 'text-sm text-secondary-500 mb-3', dir: 'auto', text: t('en.import.hint') }), field('', area), force, msg, report),
    [button({ label: t('cancel'), onclick: closePanel }), checkBtn, applyBtn], 'import');
}

// Content packs: list per platform, which pack the game gets, activate / deactivate / roll back, files and events.
var asPackState = { platform: '', status: '' };

function rangeText(p) {
  if (!p.minAppVersion && !p.maxAppVersion) return t('pk.any');
  if (p.minAppVersion && p.maxAppVersion) return tr('pk.fromTo', {}, { a: p.minAppVersion, b: p.maxAppVersion });
  return p.minAppVersion ? tr('pk.from', {}, { v: p.minAppVersion }) : tr('pk.to', {}, { v: p.maxAppVersion });
}

function packsPage(root) {
  var frame = pageFrame('packs', t('desc.packs'), [button({ label: t('refresh'), data: { 'assets-action': 'refresh-packs' }, onclick: function () { packsPage(root); } })]);
  clear(root).appendChild(frame.el);
  var c = frame.content;
  c.appendChild(h('p', { class: 'text-sm text-secondary-500', text: t('loading') }));
  Promise.all([api('GET', '/assets/packs' + qs(asPackState)), api('GET', '/assets/packs/events')]).then(function (res) {
    var d = res[0], events = res[1].items || [];
    clear(c);
    var platforms = (d.platforms || []).slice();
    d.items.forEach(function (p) { if (platforms.indexOf(p.platform) === -1) platforms.push(p.platform); });
    var plat = select(asPackState.platform, [{ value: '', label: t('pk.allPlatforms') }].concat(platforms.map(function (p) { return { value: p, label: p }; })), { name: 'platform', label: t('pk.platform') });
    var status = select(asPackState.status, [{ value: '', label: t('all') }, { value: 'active', label: t('pk.active') }, { value: 'ready', label: t('pk.ready') }, { value: 'uploading', label: t('pk.uploading') }], { name: 'status', label: t('pk.status') });
    var change = function () { asPackState.platform = plat.value; asPackState.status = status.value; packsPage(root); };
    plat.addEventListener('change', change);
    status.addEventListener('change', change);
    c.appendChild(notice('secondary', t('pk.upload.hint')));
    c.appendChild(resolverCard(platforms));
    c.appendChild(h('div', { class: 'as-filters flex flex-wrap gap-3 items-end' }, field(t('pk.platform'), plat), field(t('pk.status'), status)));
    if (!d.items.length) c.appendChild(notice('secondary', t('empty')));
    else c.appendChild(h('div', { 'data-assets-list': 'packs' }, table([
      { label: t('pk.platform') }, { label: t('pk.version') }, { label: t('pk.range') }, { label: t('pk.status') }, { label: t('pk.files') }, { label: t('pk.created') }, { label: t('pk.activated') }
    ], d.items.map(function (p) {
      var tags = h('div', { class: 'flex flex-wrap gap-1' }, tag(t('pk.' + p.status), p.status === 'ready' ? 'secondary' : 'warning'),
        p.active ? tag(t('pk.active'), 'success') : null, d.current && d.current[p.platform] === p.id ? tag(t('pk.current'), 'primary') : null);
      return { id: p.version, item: p, cells: [mono(p.platform), mono(p.version), h('span', { dir: 'ltr', text: rangeText(p) }), tags,
        h('span', {}, fmtNum(p.fileCount) + ' · ' + fmtBytes(p.totalBytes)), h('div', {}, h('div', { text: fmtDate(p.createdAt) }), h('div', { class: 'text-xs text-secondary-500', dir: 'auto', text: p.createdBy })),
        p.activatedAt ? h('div', {}, h('div', { text: fmtDate(p.activatedAt) }), h('div', { class: 'text-xs text-secondary-500', dir: 'auto', text: p.activatedBy })) : '—'] };
    }), function (p) { packPanel(p.id, function () { packsPage(root); }); })));
    c.appendChild(section(t('pk.events'), eventList(events), 'events'));
  }, function (err) { clear(c).appendChild(notice('error', asErrText(err))); });
}

function eventList(events) {
  if (!events.length) return h('p', { class: 'text-sm text-secondary-500', text: t('empty') });
  return h('ul', { 'data-assets-events': '' }, events.slice(0, 30).map(function (e) {
    return h('li', { class: 'flex flex-wrap gap-2 py-1.5 text-sm border-b border-secondary-200 dark:border-secondary-700' },
      mono(e.platform), mono(e.version), tag(has('ev.' + e.action) ? t('ev.' + e.action) : e.action, e.action === 'activated' ? 'success' : e.action === 'deactivated' || e.action === 'deleted' ? 'warning' : 'secondary'),
      h('span', { class: 'text-xs text-secondary-500', dir: 'auto', text: fmtDate(e.createdAt) + ' · ' + e.actor }));
  }));
}

function resolverCard(platforms) {
  var plat = select(platforms[0] || 'Android', platforms.map(function (p) { return { value: p, label: p }; }), { name: 'resolve-platform', label: t('pk.platform') });
  var ver = input('', { name: 'resolve-version', ltr: true, placeholder: '1.2.0', label: t('pk.appVersion') });
  var out = h('div', { class: 'mt-2', 'data-assets-resolved': '' });
  var go = button({ label: t('pk.resolve'), data: { 'assets-action': 'resolve' }, onclick: function () {
    clear(out);
    api('GET', '/assets/packs/preview' + qs({ platform: plat.value, appVersion: ver.value.trim() })).then(function (d) {
      out.appendChild(d.pack ? notice('success', tr('pk.resolved', {}, { v: d.pack.version, id: d.pack.id.slice(0, 8) })) : notice('warning', t('pk.noPack')));
    }, function (err) { out.appendChild(notice('error', asErrText(err))); });
  } });
  return section(t('pk.resolver'), h('div', {}, h('div', { class: 'as-filters flex flex-wrap gap-3 items-end' }, field(t('pk.platform'), plat), field(t('pk.appVersion'), ver), go), out), 'resolver');
}

function packPanel(id, done) {
  api('GET', '/assets/packs/' + id).then(function (d) {
    var p = d.item;
    var act = function (path, confirmKey, okLabel) {
      confirmBox(tr(confirmKey, {}, { v: p.version, p: p.platform }), okLabel).then(function (ok) {
        if (!ok) return;
        api(path === 'delete' ? 'DELETE' : 'POST', '/assets/packs/' + p.id + (path === 'delete' ? '' : '/' + path), path === 'delete' ? undefined : {}).then(function () {
          toast('success', t('saved')); closePanel(); done();
        }, function (err) { toast('error', asErrText(err)); });
      });
    };
    var buttons = [];
    if (p.status === 'ready' && !p.active) {
      buttons.push(button({ label: p.activatedAt ? t('pk.rollback') : t('pk.activate'), variant: 'primary', data: { 'assets-action': 'activate-pack' }, onclick: function () { act('activate', p.activatedAt ? 'pk.rollback.confirm' : 'pk.activate.confirm', t('pk.activate')); } }));
    }
    if (p.active) buttons.push(button({ label: t('pk.deactivate'), data: { 'assets-action': 'deactivate-pack' }, onclick: function () { act('deactivate', 'pk.deactivate.confirm', t('pk.deactivate')); } }));
    if (!p.active) buttons.push(button({ label: t('delete'), variant: 'danger', data: { 'assets-action': 'delete-pack' }, onclick: function () { act('delete', 'pk.delete', t('delete')); } }));
    var urls = d.index ? h('div', { class: 'flex flex-col gap-2', 'data-assets-urls': '' }, [['pk.baseUrl', d.index.baseUrl], ['pk.catalogUrl', d.index.catalogUrl], ['pk.hashUrl', d.index.hashUrl], ['pk.catalogHash', d.index.catalogHash]].filter(function (x) { return x[1]; }).map(function (x) {
      return h('div', {}, h('div', { class: 'text-xs text-secondary-500', text: t(x[0]) }), h('div', { class: 'flex items-center gap-2' }, h('code', { class: 'as-pre flex-1', dir: 'ltr', text: x[1] }), copyButton(x[1], x[0])));
    })) : null;
    var head = h('div', { class: 'flex flex-wrap gap-2 items-center mb-4' }, mono(p.platform), mono(p.version), tag(t('pk.' + p.status), 'secondary'), p.active ? tag(t('pk.active'), 'success') : tag(t('pk.inactive'), 'secondary'),
      h('span', { class: 'text-sm', dir: 'ltr', text: rangeText(p) }), h('span', { class: 'text-sm text-secondary-500', text: fmtNum(p.fileCount) + ' · ' + fmtBytes(p.totalBytes) }));
    var files = table([{ label: 'path' }, { label: t('f.size') }, { label: t('f.hash') }], (d.files || []).slice(0, 300).map(function (f) {
      return { id: f.path, cells: [mono(f.path), fmtBytes(f.size), h('span', {}, mono(f.sha256.slice(0, 12) + '…'), f.stored ? null : tag(t('pk.notStored'), 'warning'))] };
    }));
    var content = h('div', {}, head, p.note ? h('p', { class: 'text-sm mb-3', dir: 'auto', text: t('pk.note') + ': ' + p.note }) : null,
      urls ? section(t('pk.urls'), urls, 'urls') : null, h('div', { class: 'h-3' }), section(t('pk.files'), files, 'files'), h('div', { class: 'h-3' }), section(t('pk.events'), eventList(d.events || []), 'pack-events'));
    sidePanel(p.platform + ' · ' + p.version, content, [button({ label: t('close'), onclick: closePanel })].concat(buttons), 'pack');
  }, function (err) { toast('error', asErrText(err)); });
}

function assetsSettingsPage(root) {
  settingsPage(root, 'settings', '/assets', function (cur, defs, d) {
    var e = cur.entries, hs = cur.history, p = cur.packs;
    var kb = function (v) { return numInput(Math.round((v || 0) / 1024), ''); };
    var mb = function (v) { return numInput(Math.round((v || 0) / 1048576), ''); };
    var maxEntry = kb(e.maxEntryBytes); maxEntry.name = 'entries.maxEntryKb';
    var maxPush = numInput(e.maxPushItems, 'entries.maxPushItems'), maxFetch = numInput(e.maxFetchItems, 'entries.maxFetchItems'), maxEntries = numInput(e.maxEntries, 'entries.maxEntries');
    var conflict = select(e.conflictMode, [{ value: 'reject', label: t('se.conflict.reject') }, { value: 'skip', label: t('se.conflict.skip') }], { name: 'entries.conflictMode' });
    var force = check(e.allowForce, t('se.allowForce'), 'entries.allowForce');
    var access = select(e.access, [{ value: 'session', label: t('se.access.session') }, { value: 'public', label: t('se.access.public') }], { name: 'entries.access' });
    var cache = numInput(e.cacheSeconds, 'entries.cacheSeconds');
    var keepV = numInput(hs.keepVersions, 'history.keepVersions'), keepD = numInput(hs.keepDays, 'history.keepDays'), keepDel = numInput(hs.keepDeletedDays, 'history.keepDeletedDays');
    var platforms = input((p.platforms || []).join(', '), { name: 'packs.platforms', ltr: true });
    var part = kb(p.partBytes); part.name = 'packs.partKb';
    var maxFile = mb(p.maxFileBytes); maxFile.name = 'packs.maxFileMb';
    var maxPack = mb(p.maxPackBytes); maxPack.name = 'packs.maxPackMb';
    var maxFiles = numInput(p.maxFiles, 'packs.maxFiles'), uploadH = numInput(p.uploadHours, 'packs.uploadHours'), keepI = numInput(p.keepInactive, 'packs.keepInactive');
    var fromEditor = check(p.activateFromEditor, t('se.activateFromEditor'), 'packs.activateFromEditor');
    var single = check(p.singleActive, t('se.singleActive'), 'packs.singleActive');
    var pub = check(p.publicIndex, t('se.publicIndex'), 'packs.publicIndex');
    var idxCache = numInput(p.indexCacheSeconds, 'packs.indexCacheSeconds');
    var base = input(p.baseUrl, { name: 'packs.baseUrl', ltr: true, placeholder: d.serverBaseUrl || 'https://cdn.example.com' });
    var redirect = check(p.redirectToStorage, t('se.redirect'), 'packs.redirectToStorage');
    var pattern = input(p.catalogPattern, { name: 'packs.catalogPattern', ltr: true, mono: true });
    var reqHash = check(p.requireHash, t('se.requireHash'), 'packs.requireHash');
    var grid = function () { return h('div', { class: 'grid sm:grid-cols-2 gap-x-4' }, Array.prototype.slice.call(arguments)); };
    var info = h('div', { class: 'flex flex-col gap-1 text-sm' }, h('div', { text: tr('se.storage', {}, { d: d.storage }) }),
      d.publicStorage ? h('div', { text: t('se.publicStorage') }) : null, h('div', { dir: 'auto' }, tr('se.scopes', {}, { s: (d.scopes || []).join(', ') })));
    return {
      el: [section(t('se.info'), info, 'info'),
        section(t('se.entries'), h('div', {}, grid(field(t('se.maxEntryKb'), maxEntry), field(t('se.maxPushItems'), maxPush), field(t('se.maxFetchItems'), maxFetch), field(t('se.maxEntries'), maxEntries)),
          field(t('se.conflictMode'), conflict), force, grid(field(t('se.access'), access), field(t('se.cacheSeconds'), cache))), 'entries'),
        section(t('se.history'), grid(field(t('se.keepVersions'), keepV), field(t('se.keepDays'), keepD), field(t('se.keepDeletedDays'), keepDel)), 'history'),
        section(t('se.packs'), h('div', {}, field(t('se.platforms'), platforms), grid(field(t('se.partKb'), part, t('se.partHint')), field(t('se.maxFiles'), maxFiles), field(t('se.maxFileMb'), maxFile), field(t('se.maxPackMb'), maxPack),
          field(t('se.uploadHours'), uploadH), field(t('se.keepInactive'), keepI)), fromEditor, single, pub, field(t('se.indexCacheSeconds'), idxCache, null, 'as-num'),
          field(t('se.baseUrl'), base), redirect, field(t('se.catalogPattern'), pattern), reqHash), 'packs')],
      read: function (c) {
        c.entries = { maxEntryBytes: numberValue(maxEntry) * 1024, maxPushItems: numberValue(maxPush), maxFetchItems: numberValue(maxFetch), maxEntries: numberValue(maxEntries) || 0,
          conflictMode: conflict.value, allowForce: force.input.checked, access: access.value, cacheSeconds: numberValue(cache) || 0 };
        c.history = { keepVersions: numberValue(keepV) || 0, keepDays: numberValue(keepD) || 0, keepDeletedDays: numberValue(keepDel) || 0 };
        c.packs = { platforms: platforms.value.split(/[,،\s]+/).map(function (x) { return x.trim(); }).filter(Boolean), partBytes: numberValue(part) * 1024,
          maxFileBytes: numberValue(maxFile) * 1048576, maxPackBytes: numberValue(maxPack) * 1048576, maxFiles: numberValue(maxFiles), uploadHours: numberValue(uploadH),
          keepInactive: numberValue(keepI) || 0, activateFromEditor: fromEditor.input.checked, singleActive: single.input.checked, publicIndex: pub.input.checked,
          indexCacheSeconds: numberValue(idxCache) || 0, baseUrl: base.value.trim(), redirectToStorage: redirect.input.checked, catalogPattern: pattern.value.trim(), requireHash: reqHash.input.checked };
        return null;
      }
    };
  });
}

var AS_ROOT = 'assets-root';
var AS_PAGES = ['entries', 'packs', 'settings'];
var AS_RENDER = { entries: entriesPage, packs: packsPage, settings: assetsSettingsPage };
var AS_HASH = /^#\/assets(?:\/([^/?]*))?/;
var asPending = null;
var asRouter = null;
var asMounted = null;
var asScheduled = false;

function asRoute(hash) {
  var m = AS_HASH.exec(hash === undefined ? location.hash : hash);
  if (!m) return null;
  return AS_PAGES.indexOf(m[1]) !== -1 ? m[1] : AS_PAGES[0];
}

function asAllowed() {
  return !!token();
}

function asTitle() {
  var page = asRoute();
  return t('group') + (page ? ' - ' + t('nav.' + page) : '');
}

function asAddRoutes() {
  var nc = window.NakamaConsole;
  var r = nc && nc.router;
  if (!r || asRouter) return !!asRouter;
  try {
    if (!r.hasRoute('assets')) {
      var layout = r.resolve('/').matched[0];
      r.addRoute({ path: '/assets/:rest(.*)*', component: layout.components.default, meta: { requiresAuth: true }, children: [{ path: '', name: 'assets', component: { name: 'AssetsPage', render: function () { return null; } }, meta: { get title() { return asTitle(); } } }] });
    }
    asRouter = r;
  } catch (e) {
    return false;
  }
  if (asPending) {
    var target = asPending;
    asPending = null;
    r.replace(target.slice(1));
  }
  return true;
}

function asNavLink(page, active) {
  var on = page === active;
  return h('li', {}, h('a', { href: '#/assets/' + page, 'data-assets-nav': page, 'aria-current': on ? 'page' : null,
    class: 'h-[37px] px-3 py-2 flex items-center gap-2 text-secondary-800 hover:text-primary-700 dark:text-[#afbccb] dark:hover:text-primary-300 hover:bg-secondary-100 dark:hover:bg-secondary-800 rounded-md border-l-[2px] border-transparent' + (on ? ' !text-primary-700 dark:!text-primary-300 !bg-secondary-50 dark:!bg-secondary-800 !border-primary-700' : '') },
    h('div', { class: 'font-medium relative truncate flex-1' }, h('span', { text: t('nav.' + page) }))));
}

function asSyncNav(active) {
  var dev = document.querySelector('.x-drawer a[href="#/api-explorer"], .x-drawer a[href="#/runtime"]');
  var group = dev && dev.closest('.x-drawer .grid > div');
  var existing = document.querySelector('[data-assets-group]');
  if (!group || !asAllowed()) { if (existing) existing.remove(); return; }
  var target = document.querySelector('.x-drawer [data-campaign-group]') || document.querySelector('.x-drawer [data-shopx-group]') || document.querySelector('.x-drawer [data-account-group]') || document.querySelector('.x-drawer [data-analytics-group]') || group;
  var key = asLang() + '|' + active;
  if (existing && existing.getAttribute('data-key') === key && existing.nextElementSibling === target) return;
  var built = h('div', { 'data-assets-group': '', 'data-l10n-skip': '', 'data-key': key, class: 'eco-nav-next' },
    h('div', { class: 'mb-2 text-secondary-500 font-bold tracking-wide uppercase truncate h-4 px-3 text-[10px]', text: t('group') }),
    h('nav', { 'aria-label': t('group') }, h('ul', { class: 'space-y-0.5' }, AS_PAGES.map(function (p) { return asNavLink(p, active); }))));
  if (existing) existing.remove();
  target.before(built);
}

function asShowMains() {
  var hidden = document.querySelectorAll('[data-assets-hidden]');
  for (var i = 0; i < hidden.length; i++) { hidden[i].style.display = ''; hidden[i].removeAttribute('data-assets-hidden'); }
}

function asUnmount() {
  var root = document.getElementById(AS_ROOT);
  if (root) root.remove();
  asShowMains();
  closePanel();
  asMounted = null;
}

function asSync() {
  asAddRoutes();
  var page = asRoute();
  asSyncNav(page);
  if (!page) { if (asMounted || document.getElementById(AS_ROOT)) asUnmount(); return; }
  var mains = document.querySelectorAll('#main > main');
  if (!mains.length) return;
  for (var i = 0; i < mains.length; i++) {
    if (mains[i].style.display !== 'none') { mains[i].style.display = 'none'; mains[i].setAttribute('data-assets-hidden', ''); }
  }
  var root = document.getElementById(AS_ROOT);
  if (!root) {
    root = h('div', { id: AS_ROOT, class: 'flex flex-col flex-1 pt-4', 'data-l10n-skip': '', 'data-assets': 'root' });
    asMounted = null;
  }
  if (root.previousElementSibling !== mains[mains.length - 1]) mains[mains.length - 1].after(root);
  var key = page + '|' + asLang();
  if (asMounted !== key) {
    asMounted = key;
    closePanel();
    AS_RENDER[page](root);
  }
}

function asSchedule() {
  if (asScheduled) return;
  asScheduled = true;
  requestAnimationFrame(function () { asScheduled = false; asSync(); });
}

function asBoot() {
  window.addEventListener('hashchange', asSchedule);
  window.addEventListener('popstate', asSchedule);
  window.addEventListener('nakama-console:ready', asSchedule);
  new MutationObserver(asSchedule).observe(document.body, { childList: true, subtree: true });
  new MutationObserver(function () { asMounted = null; asSchedule(); }).observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] });
  var waited = 0;
  var poll = setInterval(function () {
    waited += 100;
    if (asAddRoutes() || waited > 8000) clearInterval(poll);
  }, 100);
  asSync();
}

window.NakamaAssets = { version: 1, pages: function () { return AS_PAGES.map(function (p) { return { id: p, href: '#/assets/' + p, label: t('nav.' + p) }; }); } };

if (AS_HASH.test(location.hash)) {
  asPending = location.hash;
  history.replaceState(history.state, '', location.pathname + location.search + '#/');
}
if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', asBoot);
else asBoot();
})();
