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
    'loading': 'در حال بارگذاری…', 'empty': 'موردی پیدا نشد', 'page.first': 'صفحهٔ اول', 'page.last': 'صفحهٔ آخر', 'page.of': 'صفحهٔ {page} از {pages}', 'pager': 'صفحه‌بندی', 'range': '{from} تا {to} از {total}', 'prev': 'قبلی', 'next': 'بعدی', 'yes': 'بله', 'no': 'خیر',
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
    'loading': 'Loading…', 'empty': 'Nothing found', 'page.first': 'First page', 'page.last': 'Last page', 'page.of': 'Page {page} of {pages}', 'pager': 'Pagination', 'range': '{from}–{to} of {total}', 'prev': 'Previous', 'next': 'Next', 'yes': 'Yes', 'no': 'No',
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
  if (window.NakamaDate) return window.NakamaDate.format(v);
  var d = new Date(v);
  if (isNaN(d)) return '';
  try {
    return d.toLocaleString(asLang() === 'fa' ? 'fa-IR' : 'en-GB', { calendar: asLang() === 'fa' ? 'persian' : 'gregory', year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return d.toISOString();
  }
}

function tr(key, vars, raw) {
  var s = t(key, vars);
  Object.keys(raw || {}).forEach(function (k) { s = s.split('{' + k + '}').join(String(raw[k])); });
  return s;
}
