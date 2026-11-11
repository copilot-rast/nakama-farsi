var SP_MESSAGES = {
  fa: {
    'rail.label': 'منوی اصلی',
    'rail.pin': 'باز نگه‌داشتن منو',
    'rail.unpin': 'جمع کردن منو',
    'rail.pages': '{n} صفحه',
    'rail.more': 'صفحه‌های {name}',
    'rail.home': 'صفحهٔ اصلی',
    'rail.sections': 'بخش‌ها',
    'rail.console': 'کنسول مدیریت',
    'group.economy': 'اقتصاد',
    'group.publication': 'انتشار',
    'group.moderation': 'بازیکنان و گفتگو',
    'group.support': 'پشتیبانی',
    'group.developer': 'توسعه',
    'group.help': 'راهنما',
    'group.settings': 'تنظیمات',
    'settings.console': 'تنظیمات کنسول',
    'settings.users': 'کاربران',
    'settings.general': 'عمومی',
    'settings.config': 'پیکربندی',
    'settings.data': 'مدیریت داده‌ها',
    'settings.audit': 'گزارش ممیزی',
    'nav.purchases': 'خریدها',
    'nav.chat': 'گفتگو',
    'settings.project': 'تنظیمات پروژه',
    'top.search': 'جست‌وجو',
    'top.searchPages': 'جست‌وجوی صفحه‌ها',
    'top.noResults': 'صفحه‌ای پیدا نشد',
    'top.status': 'وضعیت سرور',
    'top.live': 'زنده',
    'top.lang': 'زبان',
    'top.toDark': 'حالت تیره',
    'top.toLight': 'حالت روشن',
    'top.menu': 'منو',
    'top.close': 'بستن',
    'top.sections': 'صفحه‌های این بخش',
    'crumb.label': 'مسیر صفحه',
    'palette.hint': 'برای رفتن Enter و برای بستن Esc',
    'palette.group.pages': 'صفحه‌ها',
    'palette.group.settings': 'تنظیمات'
  },
  en: {
    'rail.label': 'Main menu',
    'rail.pin': 'Keep the menu open',
    'rail.unpin': 'Collapse the menu',
    'rail.pages': '{n} pages',
    'rail.more': '{name} pages',
    'rail.home': 'Home',
    'rail.sections': 'Sections',
    'rail.console': 'Admin console',
    'group.economy': 'Economy',
    'group.publication': 'Publishing',
    'group.moderation': 'Players & chat',
    'group.support': 'Support',
    'group.developer': 'Developer',
    'group.help': 'Help',
    'group.settings': 'Settings',
    'settings.console': 'Console settings',
    'settings.users': 'Users',
    'settings.general': 'General',
    'settings.config': 'Configuration',
    'settings.data': 'Data management',
    'settings.audit': 'Audit log',
    'nav.purchases': 'Purchases',
    'nav.chat': 'Chat',
    'settings.project': 'Project settings',
    'top.search': 'Search',
    'top.searchPages': 'Search pages',
    'top.noResults': 'No page found',
    'top.status': 'Server status',
    'top.live': 'Live',
    'top.lang': 'Language',
    'top.toDark': 'Dark mode',
    'top.toLight': 'Light mode',
    'top.menu': 'Menu',
    'top.close': 'Close',
    'top.sections': 'Pages in this section',
    'crumb.label': 'Breadcrumb',
    'palette.hint': 'Enter to open, Esc to close',
    'palette.group.pages': 'Pages',
    'palette.group.settings': 'Settings'
  }
};

function spLang() {
  var l = (document.documentElement.getAttribute('lang') || 'fa').toLowerCase();
  return l.indexOf('fa') === 0 ? 'fa' : 'en';
}

function spNum(n) {
  var s = String(n);
  if (spLang() !== 'fa') return s;
  return s.replace(/[0-9]/g, function (d) { return '۰۱۲۳۴۵۶۷۸۹'[d]; });
}

function st(key, vars) {
  var cat = SP_MESSAGES[spLang()] || SP_MESSAGES.en;
  var s = cat[key] !== undefined ? cat[key] : (SP_MESSAGES.en[key] !== undefined ? SP_MESSAGES.en[key] : key);
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] === undefined ? m : typeof vars[k] === 'number' ? spNum(vars[k]) : vars[k]; });
  return s;
}
