(function () {
'use strict';
var ND_TEXT = {
  fa: {
    open: 'باز کردن تقویم',
    clear: 'پاک کردن',
    pickDate: 'انتخاب تاریخ',
    pickDateTime: 'انتخاب تاریخ و ساعت',
    pickTime: 'انتخاب ساعت',
    pickRange: 'انتخاب بازه',
    prevMonth: 'ماه قبل',
    nextMonth: 'ماه بعد',
    prevYear: 'سال قبل',
    nextYear: 'سال بعد',
    prevYears: 'سال‌های قبل',
    nextYears: 'سال‌های بعد',
    chooseMonth: 'انتخاب ماه',
    chooseYear: 'انتخاب سال',
    today: 'امروز',
    now: 'اکنون',
    done: 'تأیید',
    cancel: 'انصراف',
    time: 'ساعت',
    hour: 'ساعت',
    minute: 'دقیقه',
    increase: 'بیشتر',
    decrease: 'کمتر',
    presets: 'بازه‌های آماده',
    last7: '۷ روز اخیر',
    last30: '۳۰ روز اخیر',
    last90: '۹۰ روز اخیر',
    rangeStart: 'روز شروع را انتخاب کنید',
    rangeEnd: 'روز پایان را انتخاب کنید',
    zone: 'منطقه‌ی زمانی: {zone}',
    zoneAt: '{time} · {zone}',
    invalid: 'تاریخ را به شکل {example} بنویسید',
    invalidTime: 'ساعت را به شکل {example} بنویسید',
    calendar: 'تقویم',
    'cal.auto': 'خودکار',
    'cal.persian': 'شمسی',
    'cal.gregory': 'میلادی',
    'cal.autoHint': 'از زبان کنسول پیروی می‌کند',
    disabled: 'خارج از بازه‌ی مجاز'
  },
  en: {
    open: 'Open calendar',
    clear: 'Clear',
    pickDate: 'Choose a date',
    pickDateTime: 'Choose a date and time',
    pickTime: 'Choose a time',
    pickRange: 'Choose a range',
    prevMonth: 'Previous month',
    nextMonth: 'Next month',
    prevYear: 'Previous year',
    nextYear: 'Next year',
    prevYears: 'Earlier years',
    nextYears: 'Later years',
    chooseMonth: 'Choose month',
    chooseYear: 'Choose year',
    today: 'Today',
    now: 'Now',
    done: 'Done',
    cancel: 'Cancel',
    time: 'Time',
    hour: 'Hour',
    minute: 'Minute',
    increase: 'Increase',
    decrease: 'Decrease',
    presets: 'Quick ranges',
    last7: 'Last 7 days',
    last30: 'Last 30 days',
    last90: 'Last 90 days',
    rangeStart: 'Choose the first day',
    rangeEnd: 'Choose the last day',
    zone: 'Time zone: {zone}',
    zoneAt: '{time} · {zone}',
    invalid: 'Write the date like {example}',
    invalidTime: 'Write the time like {example}',
    calendar: 'Calendar',
    'cal.auto': 'Auto',
    'cal.persian': 'Jalali',
    'cal.gregory': 'Gregorian',
    'cal.autoHint': 'Follows the console language',
    disabled: 'Outside the allowed range'
  }
};

function ndt(key, vars) {
  var table = ND_TEXT[ndLang()] || ND_TEXT.en;
  var s = table[key];
  if (s === undefined) s = ND_TEXT.en[key];
  if (s === undefined) return key;
  if (vars) s = s.replace(/\{(\w+)\}/g, function (m, k) { return vars[k] === undefined ? m : String(vars[k]); });
  return s;
}

var ND_BREAKS = [-61, 9, 38, 199, 426, 686, 756, 818, 1111, 1181, 1210, 1635, 2060, 2097, 2192, 2262, 2324, 2394, 2456, 3178];

function ndDiv(a, b) { return ~~(a / b); }
function ndMod(a, b) { return a - ~~(a / b) * b; }

function jalCal(jy) {
  var bl = ND_BREAKS.length;
  var gy = jy + 621;
  var leapJ = -14;
  var jp = ND_BREAKS[0];
  var jm;
  var jump = 0;
  if (jy < jp || jy >= ND_BREAKS[bl - 1]) throw new RangeError('jalali year out of range: ' + jy);
  for (var i = 1; i < bl; i++) {
    jm = ND_BREAKS[i];
    jump = jm - jp;
    if (jy < jm) break;
    leapJ = leapJ + ndDiv(jump, 33) * 8 + ndDiv(ndMod(jump, 33), 4);
    jp = jm;
  }
  var n = jy - jp;
  leapJ = leapJ + ndDiv(n, 33) * 8 + ndDiv(ndMod(n, 33) + 3, 4);
  if (ndMod(jump, 33) === 4 && jump - n === 4) leapJ += 1;
  var leapG = ndDiv(gy, 4) - ndDiv((ndDiv(gy, 100) + 1) * 3, 4) - 150;
  var march = 20 + leapJ - leapG;
  if (jump - n < 6) n = n - jump + ndDiv(jump + 4, 33) * 33;
  var leap = ndMod(ndMod(n + 1, 33) - 1, 4);
  if (leap === -1) leap = 4;
  return { leap: leap, gy: gy, march: march };
}

function g2d(gy, gm, gd) {
  var d = ndDiv((gy + ndDiv(gm - 8, 6) + 100100) * 1461, 4) + ndDiv(153 * ndMod(gm + 9, 12) + 2, 5) + gd - 34840408;
  return d - ndDiv(ndDiv(gy + 100100 + ndDiv(gm - 8, 6), 100) * 3, 4) + 752;
}

function d2g(jdn) {
  var j = 4 * jdn + 139361631;
  j = j + ndDiv(ndDiv(4 * jdn + 183187720, 146097) * 3, 4) * 4 - 3908;
  var i = ndDiv(ndMod(j, 1461), 4) * 5 + 308;
  var gd = ndDiv(ndMod(i, 153), 5) + 1;
  var gm = ndMod(ndDiv(i, 153), 12) + 1;
  var gy = ndDiv(j, 1461) - 100100 + ndDiv(8 - gm, 6);
  return { y: gy, m: gm, d: gd };
}

function j2d(jy, jm, jd) {
  var r = jalCal(jy);
  return g2d(r.gy, 3, r.march) + (jm - 1) * 31 - ndDiv(jm, 7) * (jm - 7) + jd - 1;
}

function d2j(jdn) {
  var gy = d2g(jdn).y;
  var jy = gy - 621;
  var r = jalCal(jy);
  var k = jdn - g2d(gy, 3, r.march);
  if (k >= 0) {
    if (k <= 185) return { y: jy, m: 1 + ndDiv(k, 31), d: ndMod(k, 31) + 1 };
    k -= 186;
  } else {
    jy -= 1;
    k += 179;
    if (r.leap === 1) k += 1;
  }
  return { y: jy, m: 7 + ndDiv(k, 30), d: ndMod(k, 30) + 1 };
}

function jalaliLeap(jy) { return jalCal(jy).leap === 0; }

function jalaliMonthLength(jy, jm) {
  if (jm <= 6) return 31;
  if (jm <= 11) return 30;
  return jalaliLeap(jy) ? 30 : 29;
}

function gregorianLeap(y) { return (y % 4 === 0 && y % 100 !== 0) || y % 400 === 0; }

function gregorianMonthLength(y, m) {
  return [31, gregorianLeap(y) ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31][m - 1];
}

function toJalali(gy, gm, gd) { return d2j(g2d(gy, gm, gd)); }

function toGregorian(jy, jm, jd) { return d2g(j2d(jy, jm, jd)); }

var ND_CALS = {
  persian: {
    id: 'persian',
    rtl: true,
    fromDay: d2j,
    toDay: j2d,
    monthLength: jalaliMonthLength,
    valid: function (y, m, d) { return y > 0 && y < 3178 && m >= 1 && m <= 12 && d >= 1 && d <= jalaliMonthLength(y, m); }
  },
  gregory: {
    id: 'gregory',
    rtl: false,
    fromDay: d2g,
    toDay: g2d,
    monthLength: gregorianMonthLength,
    valid: function (y, m, d) { return y > 0 && y < 10000 && m >= 1 && m <= 12 && d >= 1 && d <= gregorianMonthLength(y, m); }
  }
};

function ndPad(n, w) {
  var s = String(Math.abs(n));
  while (s.length < (w || 2)) s = '0' + s;
  return (n < 0 ? '-' : '') + s;
}

function dayToJdn(s) {
  var m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(s || '');
  if (!m) return null;
  var y = +m[1];
  var mo = +m[2];
  var d = +m[3];
  if (!ND_CALS.gregory.valid(y, mo, d)) return null;
  return g2d(y, mo, d);
}

function jdnToDay(n) {
  var g = d2g(n);
  return ndPad(g.y, 4) + '-' + ndPad(g.m) + '-' + ndPad(g.d);
}

function jdnWeekday(n) { return ndMod(ndMod(n + 1, 7) + 7, 7); }

var NDTF = Intl.DateTimeFormat;
var ndPartsCache = {};

function zoneFormatter(zone) {
  if (!ndPartsCache[zone]) ndPartsCache[zone] = new NDTF('en-US', { timeZone: zone, calendar: 'gregory', numberingSystem: 'latn', hourCycle: 'h23', year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' });
  return ndPartsCache[zone];
}

function validZone(zone) {
  if (!zone || typeof zone !== 'string') return false;
  try { zoneFormatter(zone); return true; } catch (e) { return false; }
}

function zoneParts(ms, zone) {
  var o = {};
  zoneFormatter(zone).formatToParts(new Date(ms)).forEach(function (p) { o[p.type] = p.value; });
  return { y: +o.year, m: +o.month, d: +o.day, h: +o.hour % 24, mi: +o.minute, s: +o.second };
}

function zoneOffset(ms, zone) {
  var o = zoneParts(ms, zone);
  return Date.UTC(o.y, o.m - 1, o.d, o.h, o.mi, o.s) - Math.floor(ms / 1000) * 1000;
}

function wallMs(y, m, d, h, mi, zone) {
  var guess = Date.UTC(y, m - 1, d, h, mi);
  var ms = guess - zoneOffset(guess, zone);
  return guess - zoneOffset(ms, zone);
}

function wallToIso(wall, zone) {
  var m = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})/.exec(wall || '');
  if (!m) return null;
  return new Date(wallMs(+m[1], +m[2], +m[3], +m[4], +m[5], zone || ndZone())).toISOString();
}

function isoToWall(iso, zone) {
  if (!iso) return '';
  var ms = new Date(iso).getTime();
  if (isNaN(ms)) return '';
  var o = zoneParts(ms, zone || ndZone());
  return ndPad(o.y, 4) + '-' + ndPad(o.m) + '-' + ndPad(o.d) + 'T' + ndPad(o.h) + ':' + ndPad(o.mi);
}

function isoToJdn(iso, zone) {
  var w = isoToWall(iso, zone);
  return w ? dayToJdn(w.slice(0, 10)) : null;
}

function todayJdn(zone) {
  var o = zoneParts(Date.now(), zone || ndZone());
  return g2d(o.y, o.m, o.d);
}

function nowWall(zone) {
  var o = zoneParts(Date.now(), zone || ndZone());
  return { jdn: g2d(o.y, o.m, o.d), h: o.h, mi: o.mi };
}

function jdnTimeToIso(jdn, h, mi, zone) {
  var g = d2g(jdn);
  return new Date(wallMs(g.y, g.m, g.d, h, mi, zone || ndZone())).toISOString();
}

var ND_CAL_KEY = 'nakama.date.calendar';
var ND_ZONE_KEY = 'nakama.date.zone';
var ND_DEFAULT_ZONE = 'Asia/Tehran';
var ND_SETTINGS_PATH = '/v2/console/base/console-settings';
var ND_PREFS = ['auto', 'persian', 'gregory'];
var ndState = { zone: null, loaded: false, loading: null, listeners: [], meta: null };

function ndStore(kind) {
  try { return kind === 'session' ? window.sessionStorage : window.localStorage; } catch (e) { return null; }
}

function ndGet(key) {
  var s = ndStore('local');
  try { return s ? s.getItem(key) : null; } catch (e) { return null; }
}

function ndSet(key, value) {
  var s = ndStore('local');
  try { if (s) { if (value == null) s.removeItem(key); else s.setItem(key, value); } } catch (e) { return; }
}

function ndLang() {
  var l = typeof document !== 'undefined' && document.documentElement ? document.documentElement.getAttribute('lang') || '' : '';
  return /^fa/i.test(l) ? 'fa' : 'en';
}

function ndLocale(lang) {
  return (lang || ndLang()) === 'fa' ? 'fa-IR' : 'en-US';
}

function calendarPref() {
  var v = ndGet(ND_CAL_KEY);
  return ND_PREFS.indexOf(v) === -1 ? 'auto' : v;
}

function activeCalendar(lang) {
  var p = calendarPref();
  if (p !== 'auto') return p;
  return (lang || ndLang()) === 'fa' ? 'persian' : 'gregory';
}

function setCalendarPref(v, opts) {
  if (ND_PREFS.indexOf(v) === -1) v = 'auto';
  ndSet(ND_CAL_KEY, v === 'auto' ? null : v);
  ndEmit('calendar');
  if (!(opts && opts.reload === false) && typeof location !== 'undefined') location.reload();
}

function ndZone() {
  if (ndState.zone) return ndState.zone;
  var cached = ndGet(ND_ZONE_KEY);
  return validZone(cached) ? cached : ND_DEFAULT_ZONE;
}

function setZone(z, opts) {
  if (!validZone(z)) return false;
  var changed = z !== ndZone();
  ndState.zone = z;
  ndSet(ND_ZONE_KEY, z);
  if (changed) ndEmit('zone');
  if (changed && opts && opts.reload && typeof location !== 'undefined') location.reload();
  return true;
}

function ndEmit(kind) {
  ndState.listeners.slice().forEach(function (fn) { try { fn(kind); } catch (e) { return; } });
}

function ndOnChange(fn) {
  ndState.listeners.push(fn);
  return function () { ndState.listeners = ndState.listeners.filter(function (x) { return x !== fn; }); };
}

function ndToken() {
  var stores = [ndStore('session'), ndStore('local')];
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

function loadSettings(force) {
  if (ndState.loading && !force) return ndState.loading;
  var tk = ndToken();
  if (!tk || typeof fetch !== 'function') return Promise.resolve({ timezone: ndZone() });
  ndState.loading = fetch(ND_SETTINGS_PATH, { headers: { Accept: 'application/json', Authorization: 'Bearer ' + tk }, credentials: 'same-origin' }).then(function (res) {
    if (!res.ok) throw new Error('settings ' + res.status);
    return res.json();
  }).then(function (res) {
    ndState.loaded = true;
    ndState.meta = res || null;
    if (res && res.timezone) setZone(res.timezone);
    return res;
  }, function () {
    ndState.loading = null;
    return { timezone: ndZone() };
  });
  return ndState.loading;
}

function saveSettings(body) {
  var tk = ndToken();
  return fetch(ND_SETTINGS_PATH, { method: 'PUT', headers: { Accept: 'application/json', 'Content-Type': 'application/json', Authorization: 'Bearer ' + (tk || '') }, credentials: 'same-origin', body: JSON.stringify(body || {}) }).then(function (res) {
    return res.text().then(function (text) {
      var data = null;
      try { data = text ? JSON.parse(text) : null; } catch (e) { data = null; }
      if (!res.ok) {
        var err = new Error((data && data.message) || 'save failed');
        err.status = res.status;
        err.code = data && (typeof data.error === 'string' ? data.error : data.code);
        err.field = data && data.field;
        throw err;
      }
      ndState.loaded = true;
      ndState.meta = data;
      if (data && data.timezone) setZone(data.timezone);
      return data;
    });
  });
}

function ndWatchLogin() {
  if (typeof window === 'undefined' || !window.addEventListener) return;
  var check = function () { if (!ndState.loaded && ndToken()) loadSettings(); };
  window.addEventListener('hashchange', check);
  window.addEventListener('focus', check);
  setTimeout(check, 0);
}

var ND_PERSIAN = '۰۱۲۳۴۵۶۷۸۹';
var ndFmtCache = {};

function ndLatin(s) {
  return String(s == null ? '' : s)
    .replace(/[۰-۹]/g, function (d) { return String(d.charCodeAt(0) - 0x06F0); })
    .replace(/[٠-٩]/g, function (d) { return String(d.charCodeAt(0) - 0x0660); });
}

function ndDigits(s, lang) {
  s = String(s == null ? '' : s);
  if ((lang || ndLang()) !== 'fa') return ndLatin(s);
  return s.replace(/\d/g, function (d) { return ND_PERSIAN[+d]; });
}

function ndFormatter(locale, o) {
  var key = locale + '|' + JSON.stringify(o);
  if (!ndFmtCache[key]) ndFmtCache[key] = new NDTF(locale, o);
  return ndFmtCache[key];
}

function ndStyleOptions(opts, hasDate, hasTime) {
  var style = opts.style || 'medium';
  var o = {};
  if (opts.parts) {
    Object.keys(opts.parts).forEach(function (k) { o[k] = opts.parts[k]; });
    return o;
  }
  if (hasDate) {
    o.year = 'numeric';
    o.day = style === 'numeric' ? '2-digit' : 'numeric';
    o.month = style === 'numeric' ? '2-digit' : style === 'short' ? 'numeric' : style === 'long' ? 'long' : 'short';
  }
  if (hasTime) { o.hour = '2-digit'; o.minute = '2-digit'; o.hourCycle = 'h23'; }
  return o;
}

function ndFormat(v, opts) {
  opts = opts || {};
  if (v === null || v === undefined || v === '') return '';
  var lang = opts.lang || ndLang();
  var locale = opts.locale || ndLocale(lang);
  var cal = opts.calendar || activeCalendar(lang);
  if (typeof v === 'string' && /^\d{1,2}:\d{2}(:\d{2})?$/.test(v)) return ndDigits(v.slice(0, 5), lang);
  var dayOnly = typeof v === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(v);
  var d = v instanceof Date ? v : new Date(dayOnly ? v + 'T12:00:00Z' : v);
  if (isNaN(d.getTime())) return String(v);
  var hasTime = !dayOnly && opts.time !== false && opts.mode !== 'date';
  var hasDate = opts.date !== false && opts.mode !== 'time';
  var o = ndStyleOptions(opts, hasDate, hasTime);
  o.calendar = cal;
  o.timeZone = dayOnly ? 'UTC' : (opts.zone || opts.timeZone || ndZone());
  if (opts.zoneName) o.timeZoneName = 'short';
  var weekday = hasDate && !opts.parts && opts.style === 'long';
  try {
    return (weekday ? ndWeekdayOf(d, locale, o.timeZone, lang) : '') + ndFormatter(locale, o).format(d);
  } catch (e) {
    delete o.timeZone;
    return ndFormatter(locale, o).format(d);
  }
}

function ndWeekdayOf(d, locale, zone, lang) {
  return ndFormatter(locale, { weekday: 'long', timeZone: zone }).format(d) + (lang === 'fa' ? ' ' : ', ');
}

function ndYmd(jdn, cal) {
  return ND_CALS[cal || activeCalendar()].fromDay(jdn);
}

function ndNumeric(value, mode, opts) {
  opts = opts || {};
  var cal = opts.calendar || activeCalendar();
  var lang = opts.lang || ndLang();
  var zone = opts.zone || ndZone();
  if (value === null || value === undefined || value === '') return '';
  if (mode === 'time') return ndDigits(String(value).slice(0, 5), lang);
  if (mode === 'range') {
    if (!value.from && !value.to) return '';
    return ndNumeric(value.from, 'date', opts) + ' – ' + ndNumeric(value.to, 'date', opts);
  }
  var jdn;
  var tail = '';
  if (mode === 'date') jdn = dayToJdn(value);
  else {
    var w = isoToWall(value, zone);
    if (!w) return '';
    jdn = dayToJdn(w.slice(0, 10));
    tail = ' ' + w.slice(11, 16);
  }
  if (jdn == null) return '';
  var p = ndYmd(jdn, cal);
  return ndDigits(ndPad(p.y, 4) + '/' + ndPad(p.m) + '/' + ndPad(p.d) + tail, lang);
}

function ndParseDate(text) {
  var m = /^(\d{3,4})\s*[-/.]\s*(\d{1,2})\s*[-/.]\s*(\d{1,2})$/.exec(text);
  if (!m) return null;
  var y = +m[1];
  var mo = +m[2];
  var d = +m[3];
  var cal = y >= 1700 ? 'gregory' : 'persian';
  if (!ND_CALS[cal].valid(y, mo, d)) return null;
  return ND_CALS[cal].toDay(y, mo, d);
}

function ndParseTime(text) {
  var m = /^(\d{1,2})\s*[:٫.]\s*(\d{2})(?:\s*:\s*(\d{2}))?$/.exec(text);
  if (!m) return null;
  var h = +m[1];
  var mi = +m[2];
  if (h > 23 || mi > 59) return null;
  return { h: h, mi: mi };
}

function ndParse(text, mode, zone) {
  var s = ndLatin(text).replace(/[‌‎‏]/g, '').trim();
  if (!s) return { empty: true };
  if (mode === 'time') {
    var t = ndParseTime(s);
    return t ? { value: ndPad(t.h) + ':' + ndPad(t.mi) } : null;
  }
  if (mode === 'range') {
    var parts = s.split(/\s*(?:–|—|~|تا|\bto\b)\s*|\s+-\s+/);
    if (parts.length !== 2) return null;
    var a = ndParseDate(parts[0]);
    var b = ndParseDate(parts[1]);
    if (a == null || b == null) return null;
    if (b < a) { var x = a; a = b; b = x; }
    return { value: { from: jdnToDay(a), to: jdnToDay(b) } };
  }
  if (mode === 'date') {
    var j = ndParseDate(s);
    return j == null ? null : { value: jdnToDay(j) };
  }
  var m = /^(.+?)(?:\s*[T,\s]\s*(\d{1,2}\s*[:٫.]\s*\d{2}(?:\s*:\s*\d{2})?))?$/.exec(s);
  if (!m) return null;
  var jd = ndParseDate(m[1].trim());
  if (jd == null) return null;
  var tm = m[2] ? ndParseTime(m[2]) : { h: 0, mi: 0 };
  if (!tm) return null;
  return { value: jdnTimeToIso(jd, tm.h, tm.mi, zone) };
}

function ndMonthName(y, m, cal, lang, width) {
  var jdn = ND_CALS[cal].toDay(y, m, 15);
  var g = d2g(jdn);
  return ndFormatter(ndLocale(lang), { calendar: cal, timeZone: 'UTC', month: width || 'long' }).format(new Date(Date.UTC(g.y, g.m - 1, g.d, 12)));
}

function ndYearText(y, lang) {
  return ndDigits(String(y), lang);
}

function ndWeekdayName(wd, lang, width) {
  var base = Date.UTC(2024, 0, 7 + wd, 12);
  return ndFormatter(ndLocale(lang), { weekday: width || 'short', timeZone: 'UTC' }).format(new Date(base));
}

function ndDayLabel(jdn, cal, lang) {
  var g = d2g(jdn);
  var d = new Date(Date.UTC(g.y, g.m - 1, g.d, 12));
  return ndWeekdayOf(d, ndLocale(lang), 'UTC', lang || ndLang()) + ndFormatter(ndLocale(lang), { calendar: cal, timeZone: 'UTC', year: 'numeric', month: 'long', day: 'numeric' }).format(d);
}

function ndWeekStart(cal, lang) {
  if (cal === 'persian') return 6;
  var first = null;
  try {
    var loc = new Intl.Locale(ndLocale(lang));
    var info = typeof loc.getWeekInfo === 'function' ? loc.getWeekInfo() : loc.weekInfo;
    if (info && info.firstDay) first = info.firstDay % 7;
  } catch (e) { first = null; }
  if (first === 0 || first === 1) return first;
  return (lang || ndLang()) === 'en' ? 0 : 1;
}

function ndIsDefault(l) { return l === undefined || l === null || (Array.isArray(l) && l.length === 0); }

function ndDefaults(o, kind) {
  var out = {};
  if (o) Object.keys(o).forEach(function (k) { out[k] = o[k]; });
  if (!out.calendar) out.calendar = activeCalendar();
  if (!out.timeZone) out.timeZone = ndZone();
  return out;
}

function installIntl() {
  if (typeof window === 'undefined' || !window.Intl || window.Intl.DateTimeFormat.__nakamaDate) return false;
  var Base = window.Intl.DateTimeFormat;
  var Wrapped = function (locales, options) {
    if (!ndIsDefault(locales)) return new Base(locales, options);
    try { return new Base(locales, ndDefaults(options)); } catch (e) { return new Base(locales, options); }
  };
  Wrapped.prototype = Base.prototype;
  Wrapped.supportedLocalesOf = Base.supportedLocalesOf.bind(Base);
  Object.defineProperty(Wrapped, 'name', { value: Base.name });
  Wrapped.__nakamaDate = true;
  window.Intl.DateTimeFormat = Wrapped;
  [['toLocaleString', null], ['toLocaleDateString', 'date'], ['toLocaleTimeString', null]].forEach(function (pair) {
    var native = window.Date.prototype[pair[0]];
    window.Date.prototype[pair[0]] = function (locales, options) {
      if (!ndIsDefault(locales)) return native.call(this, locales, options);
      try { return native.call(this, locales, ndDefaults(options, pair[1])); } catch (e) { return native.call(this, locales, options); }
    };
  });
  return true;
}

var ND_SVG = {
  calendar: 'M208,32H184V24a8,8,0,0,0-16,0v8H88V24a8,8,0,0,0-16,0v8H48A16,16,0,0,0,32,48V208a16,16,0,0,0,16,16H208a16,16,0,0,0,16-16V48A16,16,0,0,0,208,32ZM72,48v8a8,8,0,0,0,16,0V48h80v8a8,8,0,0,0,16,0V48h24V80H48V48ZM208,208H48V96H208V208Z',
  clock: 'M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm0,192a88,88,0,1,1,88-88A88.1,88.1,0,0,1,128,216Zm64-88a8,8,0,0,1-8,8H128a8,8,0,0,1-8-8V72a8,8,0,0,1,16,0v48h48A8,8,0,0,1,192,128Z',
  x: 'M205.66,194.34a8,8,0,0,1-11.32,11.32L128,139.31,61.66,205.66a8,8,0,0,1-11.32-11.32L116.69,128,50.34,61.66A8,8,0,0,1,61.66,50.34L128,116.69l66.34-66.35a8,8,0,0,1,11.32,11.32L139.31,128Z',
  next: 'M181.66,133.66l-80,80a8,8,0,0,1-11.32-11.32L164.69,128,90.34,53.66a8,8,0,0,1,11.32-11.32l80,80A8,8,0,0,1,181.66,133.66Z',
  prev: 'M165.66,202.34a8,8,0,0,1-11.32,11.32l-80-80a8,8,0,0,1,0-11.32l80-80a8,8,0,0,1,11.32,11.32L91.31,128Z',
  up: 'M213.66,165.66a8,8,0,0,1-11.32,0L128,91.31,53.66,165.66a8,8,0,0,1-11.32-11.32l80-80a8,8,0,0,1,11.32,0l80,80A8,8,0,0,1,213.66,165.66Z',
  down: 'M213.66,101.66l-80,80a8,8,0,0,1-11.32,0l-80-80A8,8,0,0,1,53.66,90.34L128,164.69l74.34-74.35a8,8,0,0,1,11.32,11.32Z'
};
var ndOpen = null;
var ndSeq = 0;

function ndIcon(name, cls) {
  return '<svg class="nd-ic' + (cls ? ' ' + cls : '') + '" viewBox="0 0 256 256" fill="currentColor" aria-hidden="true" focusable="false"><path d="' + ND_SVG[name] + '"/></svg>';
}

function ndEl(tag, attrs) {
  var el = document.createElement(tag);
  Object.keys(attrs || {}).forEach(function (k) {
    var v = attrs[k];
    if (v === null || v === undefined || v === false) return;
    if (k === 'text') el.textContent = v;
    else if (k === 'html') el.innerHTML = v;
    else if (k.slice(0, 2) === 'on' && typeof v === 'function') el.addEventListener(k.slice(2), v);
    else el.setAttribute(k, v === true ? '' : v);
  });
  for (var i = 2; i < arguments.length; i++) {
    var c = arguments[i];
    if (c === null || c === undefined || c === false) continue;
    el.appendChild(typeof c === 'string' ? document.createTextNode(c) : c);
  }
  return el;
}

function ndMobile() {
  return typeof window.matchMedia === 'function' && window.matchMedia('(max-width: 640px)').matches;
}

function ndReduced() {
  return typeof window.matchMedia === 'function' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

function ndOrigin(mount) {
  var probe = document.createElement('span');
  probe.style.cssText = 'position:fixed;left:0;top:0;width:0;height:0;visibility:hidden;pointer-events:none;';
  mount.appendChild(probe);
  var r = probe.getBoundingClientRect();
  probe.remove();
  return { x: r.left, y: r.top };
}

function ndSetFixed(el, x, y, w, origin) {
  var o = origin || { x: 0, y: 0 };
  el.style.left = Math.round(x - o.x) + 'px';
  el.style.top = Math.round(y - o.y) + 'px';
  if (w) el.style.width = Math.round(w) + 'px';
}

function closePicker(restoreFocus) {
  if (ndOpen) ndOpen.close(restoreFocus);
}

function openPicker(host) {
  if (ndOpen) ndOpen.close(false);
  var lang = ndLang();
  var cal = host.calendar || activeCalendar(lang);
  var C = ND_CALS[cal];
  var rtl = C.rtl;
  var ws = ndWeekStart(cal, lang);
  var zone = host.zone || ndZone();
  var mode = host.mode;
  var initial = host.getValue();
  var now = nowWall(zone);
  var today = now.jdn;
  var minJ = host.minJdn == null ? null : host.minJdn;
  var maxJ = host.maxJdn == null ? null : host.maxJdn;
  var st = { view: 'days', sel: null, a: null, b: null, hover: null, h: now.h, mi: 0, cursor: today, vy: 0, vm: 0, yearBase: 0 };
  if (mode === 'date') st.sel = dayToJdn(initial);
  else if (mode === 'datetime') {
    var w = isoToWall(initial, zone);
    if (w) { st.sel = dayToJdn(w.slice(0, 10)); st.h = +w.slice(11, 13); st.mi = +w.slice(14, 16); }
  } else if (mode === 'time') {
    var tm = initial ? ndParseTime(String(initial)) : null;
    if (tm) { st.h = tm.h; st.mi = tm.mi; } else { st.h = now.h; st.mi = now.mi - now.mi % 5; }
  } else if (mode === 'range' && initial) {
    st.a = dayToJdn(initial.from);
    st.b = dayToJdn(initial.to);
  }
  st.cursor = clampJ(st.sel != null ? st.sel : st.a != null ? st.a : today);
  syncView();

  var spins = {};
  var id = 'nd-pop-' + (++ndSeq);
  var titleId = id + '-title';
  var titles = { date: 'pickDate', datetime: 'pickDateTime', time: 'pickTime', range: 'pickRange' };
  var pop = ndEl('div', { class: 'nd-pop nd-mode-' + mode, id: id, role: 'dialog', 'aria-modal': 'true', 'aria-label': (host.label ? host.label + ' · ' : '') + ndt(titles[mode]), dir: rtl ? 'rtl' : 'ltr', lang: ndLocale(lang), 'data-nd-pop': mode, 'data-calendar': cal, 'data-l10n-skip': '' });
  var scrim = ndEl('div', { class: 'nd-scrim', 'aria-hidden': 'true', 'data-nd-scrim': '' });
  var head = ndEl('div', { class: 'nd-head' });
  var body = ndEl('div', { class: 'nd-body' });
  var timeBox = mode === 'datetime' || mode === 'time' ? ndEl('div', { class: 'nd-time', role: 'group', 'aria-label': ndt('time'), dir: 'ltr' }) : null;
  var hint = mode === 'range' ? ndEl('p', { class: 'nd-hint', 'aria-live': 'polite', 'data-nd-hint': '' }) : null;
  var foot = ndEl('div', { class: 'nd-foot' });
  var grab = ndEl('div', { class: 'nd-grab', 'aria-hidden': 'true' });
  pop.appendChild(grab);
  if (mode !== 'time') pop.appendChild(head);
  if (mode === 'range') pop.appendChild(presetBar());
  if (mode !== 'time') pop.appendChild(body);
  if (hint) pop.appendChild(hint);
  if (timeBox) { pop.appendChild(timeBox); buildTime(); }
  pop.appendChild(foot);
  buildFoot();

  function clampJ(j) {
    if (minJ != null && j < minJ) return minJ;
    if (maxJ != null && j > maxJ) return maxJ;
    return j;
  }

  function off(j) { return (minJ != null && j < minJ) || (maxJ != null && j > maxJ); }

  function syncView() {
    var p = C.fromDay(st.cursor);
    st.vy = p.y;
    st.vm = p.m;
  }

  function navButton(dirName, label, fn) {
    var icon = rtl ? (dirName === 'prev' ? 'next' : 'prev') : dirName;
    return ndEl('button', { type: 'button', class: 'nd-nav nd-' + dirName, 'aria-label': label, title: label, html: ndIcon(icon), 'data-nd-nav': dirName, onclick: fn });
  }

  function renderHead() {
    head.innerHTML = '';
    var labels = st.view === 'days' ? ['prevMonth', 'nextMonth'] : st.view === 'months' ? ['prevYear', 'nextYear'] : ['prevYears', 'nextYears'];
    var step = function (n) {
      return function () {
        if (st.view === 'days') moveMonth(n, false);
        else if (st.view === 'months') { st.vy += n; render(); }
        else { st.yearBase += 12 * n; render(); }
        var b = head.querySelector('[data-nd-nav="' + (n < 0 ? 'prev' : 'next') + '"]');
        if (b) b.focus();
      };
    };
    var title = ndEl('div', { class: 'nd-title', id: titleId, 'aria-live': 'polite' });
    if (st.view === 'years') {
      title.appendChild(ndEl('span', { class: 'nd-title-text', text: ndYearText(st.yearBase, lang) + ' – ' + ndYearText(st.yearBase + 11, lang) }));
    } else {
      if (st.view === 'days') title.appendChild(ndEl('button', { type: 'button', class: 'nd-title-btn', 'data-nd-view': 'months', 'aria-label': ndt('chooseMonth') + ': ' + ndMonthName(st.vy, st.vm, cal, lang), text: ndMonthName(st.vy, st.vm, cal, lang), onclick: function () { st.view = 'months'; render(); focusCursor(); } }));
      title.appendChild(ndEl('button', { type: 'button', class: 'nd-title-btn', 'data-nd-view': 'years', 'aria-label': ndt('chooseYear') + ': ' + ndYearText(st.vy, lang), text: ndYearText(st.vy, lang), onclick: function () { st.view = 'years'; st.yearBase = st.vy - 5; render(); focusCursor(); } }));
    }
    head.appendChild(navButton('prev', ndt(labels[0]), step(-1)));
    head.appendChild(title);
    head.appendChild(navButton('next', ndt(labels[1]), step(1)));
  }

  function inRange(j) {
    var a = st.a;
    var b = st.b != null ? st.b : st.hover;
    if (a == null || b == null) return false;
    var lo = Math.min(a, b);
    var hi = Math.max(a, b);
    return j >= lo && j <= hi;
  }

  function dayClasses(btn, j) {
    var p = C.fromDay(j);
    var cls = ['nd-day'];
    if (p.m !== st.vm) cls.push('is-out');
    if (j === today) cls.push('is-today');
    if (off(j)) cls.push('is-off');
    var selected = false;
    if (mode === 'range') {
      var a = st.a;
      var b = st.b != null ? st.b : st.hover;
      if (a != null && j === Math.min(a, b == null ? a : b)) cls.push('is-start');
      if (a != null && j === Math.max(a, b == null ? a : b)) cls.push('is-end');
      if (inRange(j)) cls.push('is-in');
      selected = j === st.a || j === st.b || (st.a != null && st.b != null && inRange(j));
    } else selected = j === st.sel;
    if (selected) cls.push('is-sel');
    btn.className = cls.join(' ');
    btn.parentNode.setAttribute('aria-selected', selected ? 'true' : 'false');
  }

  function renderDays() {
    body.innerHTML = '';
    var table = ndEl('table', { class: 'nd-grid', role: 'grid', 'aria-labelledby': titleId, 'data-nd-grid': '' });
    var hr = ndEl('tr', {});
    for (var i = 0; i < 7; i++) {
      var wd = (ws + i) % 7;
      hr.appendChild(ndEl('th', { scope: 'col', abbr: ndWeekdayName(wd, lang, 'long'), text: ndWeekdayName(wd, lang, lang === 'fa' ? 'narrow' : 'short') }));
    }
    table.appendChild(ndEl('thead', {}, hr));
    var tb = ndEl('tbody', {});
    var first = C.toDay(st.vy, st.vm, 1);
    var start = first - ((jdnWeekday(first) - ws + 7) % 7);
    for (var r = 0; r < 6; r++) {
      var tr = ndEl('tr', {});
      for (var c = 0; c < 7; c++) {
        var j = start + r * 7 + c;
        var td = ndEl('td', { role: 'gridcell' });
        var btn = ndEl('button', { type: 'button', tabindex: j === st.cursor ? '0' : '-1', 'data-day': jdnToDay(j), 'data-jdn': String(j), 'aria-label': ndDayLabel(j, cal, lang) + (off(j) ? ' · ' + ndt('disabled') : ''), text: ndDigits(String(C.fromDay(j).d), lang) });
        if (j === today) btn.setAttribute('aria-current', 'date');
        if (off(j)) btn.setAttribute('aria-disabled', 'true');
        td.appendChild(btn);
        tr.appendChild(td);
        dayClasses(btn, j);
      }
      tb.appendChild(tr);
    }
    table.appendChild(tb);
    table.addEventListener('click', function (e) {
      var b = e.target.closest && e.target.closest('button[data-jdn]');
      if (b) pickDay(+b.getAttribute('data-jdn'));
    });
    if (mode === 'range') {
      table.addEventListener('mouseover', function (e) {
        var b = e.target.closest && e.target.closest('button[data-jdn]');
        if (!b || st.a == null || st.b != null) return;
        st.hover = +b.getAttribute('data-jdn');
        paintRange();
      });
    }
    body.appendChild(table);
  }

  function paintRange() {
    Array.prototype.forEach.call(body.querySelectorAll('button[data-jdn]'), function (b) { dayClasses(b, +b.getAttribute('data-jdn')); });
  }

  function renderMonths() {
    body.innerHTML = '';
    var grid = ndEl('div', { class: 'nd-cells nd-months', role: 'group', 'aria-labelledby': titleId });
    var cur = C.fromDay(st.cursor);
    for (var m = 1; m <= 12; m++) {
      (function (m) {
        var on = cur.y === st.vy && cur.m === m;
        var sel = st.sel != null && C.fromDay(st.sel).y === st.vy && C.fromDay(st.sel).m === m;
        grid.appendChild(ndEl('button', { type: 'button', class: 'nd-cell' + (sel ? ' is-sel' : ''), 'data-nd-month': String(m), tabindex: on || (cur.y !== st.vy && m === 1) ? '0' : '-1', 'aria-pressed': sel ? 'true' : 'false', text: ndMonthName(st.vy, m, cal, lang), onclick: function () {
          var d = Math.min(C.fromDay(st.cursor).d, C.monthLength(st.vy, m));
          st.cursor = C.toDay(st.vy, m, d);
          st.view = 'days';
          syncView();
          render();
          focusCursor();
        } }));
      })(m);
    }
    body.appendChild(grid);
  }

  function renderYears() {
    body.innerHTML = '';
    var grid = ndEl('div', { class: 'nd-cells nd-years', role: 'group', 'aria-labelledby': titleId });
    for (var i = 0; i < 12; i++) {
      (function (y) {
        var on = y === st.vy;
        grid.appendChild(ndEl('button', { type: 'button', class: 'nd-cell' + (on ? ' is-sel' : ''), 'data-nd-year': String(y), tabindex: on || (i === 0 && (st.vy < st.yearBase || st.vy > st.yearBase + 11)) ? '0' : '-1', 'aria-pressed': on ? 'true' : 'false', text: ndYearText(y, lang), onclick: function () {
          var p = C.fromDay(st.cursor);
          var d = Math.min(p.d, C.monthLength(y, p.m));
          st.cursor = C.toDay(y, p.m, d);
          syncView();
          st.view = 'months';
          render();
          focusCursor();
        } }));
      })(st.yearBase + i);
    }
    body.appendChild(grid);
  }

  function render() {
    if (mode === 'time') return;
    renderHead();
    if (st.view === 'days') renderDays();
    else if (st.view === 'months') renderMonths();
    else renderYears();
    if (hint) hint.textContent = ndt(st.a != null && st.b == null ? 'rangeEnd' : 'rangeStart');
  }

  function focusCursor() {
    var t = st.view === 'days' ? body.querySelector('button[data-jdn="' + st.cursor + '"]') : body.querySelector('.nd-cell[tabindex="0"]');
    if (t) t.focus({ preventScroll: true });
  }

  function moveMonth(n, keepFocus) {
    var p = C.fromDay(st.cursor);
    var y = p.y;
    var m = p.m + n;
    while (m < 1) { m += 12; y -= 1; }
    while (m > 12) { m -= 12; y += 1; }
    st.cursor = C.toDay(y, m, Math.min(p.d, C.monthLength(y, m)));
    syncView();
    render();
    if (keepFocus) focusCursor();
  }

  function setCursor(j) {
    var before = C.fromDay(st.cursor);
    st.cursor = j;
    var p = C.fromDay(j);
    if (p.y !== before.y || p.m !== before.m || p.m !== st.vm || p.y !== st.vy) { syncView(); render(); }
    else Array.prototype.forEach.call(body.querySelectorAll('button[data-jdn]'), function (b) { b.setAttribute('tabindex', +b.getAttribute('data-jdn') === j ? '0' : '-1'); });
    if (mode === 'range' && st.a != null && st.b == null) { st.hover = j; paintRange(); }
    focusCursor();
  }

  function emitDateTime() {
    if (st.sel == null) return;
    var iso = jdnTimeToIso(st.sel, st.h, st.mi, zone);
    if (host.min && new Date(iso) < new Date(host.min)) iso = new Date(host.min).toISOString();
    if (host.max && new Date(iso) > new Date(host.max)) iso = new Date(host.max).toISOString();
    host.commit(iso, false);
  }

  function pickDay(j) {
    if (off(j)) return;
    st.cursor = j;
    if (mode === 'date') { st.sel = j; host.commit(jdnToDay(j), true); close(true); return; }
    if (mode === 'datetime') {
      st.sel = j;
      emitDateTime();
      var p = C.fromDay(j);
      if (p.m !== st.vm || p.y !== st.vy) syncView();
      render();
      focusCursor();
      paintTime();
      return;
    }
    if (st.a == null || st.b != null) { st.a = j; st.b = null; st.hover = j; render(); focusCursor(); return; }
    var lo = Math.min(st.a, j);
    var hi = Math.max(st.a, j);
    st.a = lo;
    st.b = hi;
    host.commit({ from: jdnToDay(lo), to: jdnToDay(hi) }, true);
    close(true);
  }

  function presetBar() {
    var bar = ndEl('div', { class: 'nd-presets', role: 'group', 'aria-label': ndt('presets') });
    [['today', 1], ['last7', 7], ['last30', 30], ['last90', 90]].forEach(function (p) {
      bar.appendChild(ndEl('button', { type: 'button', class: 'nd-chip', 'data-nd-preset': String(p[1]), text: ndt(p[0]), onclick: function () {
        var to = clampJ(today);
        var from = clampJ(today - (p[1] - 1));
        st.a = from;
        st.b = to;
        host.commit({ from: jdnToDay(from), to: jdnToDay(to) }, true);
        close(true);
      } }));
    });
    return bar;
  }

  function spin(kind, max) {
    var label = ndt(kind === 'h' ? 'hour' : 'minute');
    var val = ndEl('div', { class: 'nd-spin-val', role: 'spinbutton', tabindex: '0', 'aria-label': label, 'aria-valuemin': '0', 'aria-valuemax': String(max), 'data-nd-spin': kind });
    var buf = '';
    var bufAt = 0;
    var set = function (v) {
      v = ((v % (max + 1)) + (max + 1)) % (max + 1);
      if (kind === 'h') st.h = v; else st.mi = v;
      if (mode === 'datetime' && st.sel == null) st.sel = clampJ(st.cursor);
      changed();
    };
    var cur = function () { return kind === 'h' ? st.h : st.mi; };
    val.addEventListener('keydown', function (e) {
      var k = e.key;
      if (k === 'ArrowUp') set(cur() + 1);
      else if (k === 'ArrowDown') set(cur() - 1);
      else if (k === 'PageUp') set(cur() + (kind === 'h' ? 6 : 10));
      else if (k === 'PageDown') set(cur() - (kind === 'h' ? 6 : 10));
      else if (k === 'Home') set(0);
      else if (k === 'End') set(max);
      else if (/^[0-9۰-۹٠-٩]$/.test(k)) {
        var d = ndLatin(k);
        var t = Date.now();
        buf = t - bufAt < 1200 && buf.length < 2 ? buf + d : d;
        bufAt = t;
        var n = +buf;
        if (n > max) { buf = d; n = +d; }
        set(n);
      } else return;
      e.preventDefault();
      e.stopPropagation();
    });
    val.addEventListener('wheel', function (e) { e.preventDefault(); set(cur() + (e.deltaY < 0 ? 1 : -1)); }, { passive: false });
    var up = ndEl('button', { type: 'button', class: 'nd-spin-btn', tabindex: '-1', 'aria-label': label + ' · ' + ndt('increase'), html: ndIcon('up'), onclick: function () { set(cur() + 1); val.focus(); } });
    var down = ndEl('button', { type: 'button', class: 'nd-spin-btn', tabindex: '-1', 'aria-label': label + ' · ' + ndt('decrease'), html: ndIcon('down'), onclick: function () { set(cur() - 1); val.focus(); } });
    spins[kind] = val;
    return ndEl('div', { class: 'nd-spin' }, up, val, down);
  }

  function buildTime() {
    timeBox.appendChild(ndEl('span', { class: 'nd-time-label', html: ndIcon('clock'), title: ndt('time') }));
    timeBox.appendChild(spin('h', 23));
    timeBox.appendChild(ndEl('span', { class: 'nd-time-sep', 'aria-hidden': 'true', text: ':' }));
    timeBox.appendChild(spin('mi', 59));
    paintTime();
  }

  function paintTime() {
    if (!timeBox) return;
    [['h', st.h], ['mi', st.mi]].forEach(function (p) {
      var v = spins[p[0]];
      v.textContent = ndDigits(ndPad(p[1]), lang);
      v.setAttribute('aria-valuenow', String(p[1]));
      v.setAttribute('aria-valuetext', ndDigits(ndPad(st.h) + ':' + ndPad(st.mi), lang));
    });
  }

  function changed() {
    paintTime();
    if (mode === 'time') host.commit(ndPad(st.h) + ':' + ndPad(st.mi), false);
    else if (mode === 'datetime') { emitDateTime(); if (st.view === 'days') paintSel(); }
  }

  function paintSel() {
    Array.prototype.forEach.call(body.querySelectorAll('button[data-jdn]'), function (b) { dayClasses(b, +b.getAttribute('data-jdn')); });
  }

  function footButton(label, cls, fn, data) {
    var attrs = { type: 'button', class: 'nd-btn ' + cls, text: label, onclick: fn };
    attrs['data-nd-' + data] = '';
    return ndEl('button', attrs);
  }

  function buildFoot() {
    var start = ndEl('div', { class: 'nd-foot-start' });
    var end = ndEl('div', { class: 'nd-foot-end' });
    if (mode === 'date') start.appendChild(footButton(ndt('today'), 'nd-btn-ghost', function () { if (off(today)) return; host.commit(jdnToDay(today), true); close(true); }, 'today'));
    if (mode === 'datetime' || mode === 'time') start.appendChild(footButton(ndt('now'), 'nd-btn-ghost', function () {
      var n = nowWall(zone);
      st.h = n.h;
      st.mi = n.mi;
      if (mode === 'datetime') { st.sel = clampJ(n.jdn); st.cursor = st.sel; syncView(); render(); }
      changed();
    }, 'now'));
    if (!host.required) start.appendChild(footButton(ndt('clear'), 'nd-btn-ghost', function () { host.commit(null, true); close(true); }, 'erase'));
    if (mode === 'datetime' || mode === 'time') end.appendChild(footButton(ndt('done'), 'nd-btn-primary', function () { if (mode === 'datetime' && st.sel == null) { st.sel = clampJ(st.cursor); emitDateTime(); } close(true); }, 'done'));
    else end.appendChild(footButton(ndt('cancel'), 'nd-btn-ghost', function () { close(true); }, 'cancel'));
    foot.appendChild(start);
    foot.appendChild(end);
  }

  function gridKeys(e) {
    var t = e.target;
    var k = e.key;
    if (t.hasAttribute && t.hasAttribute('data-jdn')) {
      var j = +t.getAttribute('data-jdn');
      var p = C.fromDay(j);
      var fwd = rtl ? -1 : 1;
      if (k === 'ArrowRight') setCursor(j + fwd);
      else if (k === 'ArrowLeft') setCursor(j - fwd);
      else if (k === 'ArrowDown') setCursor(j + 7);
      else if (k === 'ArrowUp') setCursor(j - 7);
      else if (k === 'Home') setCursor(j - ((jdnWeekday(j) - ws + 7) % 7));
      else if (k === 'End') setCursor(j + (6 - ((jdnWeekday(j) - ws + 7) % 7)));
      else if (k === 'PageUp' || k === 'PageDown') {
        var n = (k === 'PageUp' ? -1 : 1) * (e.shiftKey ? 12 : 1);
        var y = p.y;
        var m = p.m + n;
        while (m < 1) { m += 12; y -= 1; }
        while (m > 12) { m -= 12; y += 1; }
        setCursor(C.toDay(y, m, Math.min(p.d, C.monthLength(y, m))));
      } else if (k === 'Enter' || k === ' ') pickDay(j);
      else return false;
      return true;
    }
    if (t.classList && t.classList.contains('nd-cell')) {
      var cells = Array.prototype.slice.call(body.querySelectorAll('.nd-cell'));
      var i = cells.indexOf(t);
      var fw = rtl ? -1 : 1;
      var to = k === 'ArrowRight' ? i + fw : k === 'ArrowLeft' ? i - fw : k === 'ArrowDown' ? i + 3 : k === 'ArrowUp' ? i - 3 : k === 'Home' ? 0 : k === 'End' ? cells.length - 1 : null;
      if (to === null) return false;
      to = Math.max(0, Math.min(cells.length - 1, to));
      cells.forEach(function (c, x) { c.setAttribute('tabindex', x === to ? '0' : '-1'); });
      cells[to].focus();
      return true;
    }
    return false;
  }

  function focusables() {
    return Array.prototype.filter.call(pop.querySelectorAll('button, [tabindex="0"]'), function (n) {
      return !n.disabled && n.getAttribute('tabindex') !== '-1' && n.offsetParent !== null;
    });
  }

  function escKey(e) {
    if (e.key !== 'Escape') return;
    e.preventDefault();
    e.stopPropagation();
    if (st.view !== 'days') { st.view = 'days'; render(); focusCursor(); return; }
    host.commit(initial, true);
    close(true);
  }

  pop.addEventListener('keydown', function (e) {
    if (e.key === 'Tab') {
      var list = focusables();
      if (!list.length) return;
      var i = list.indexOf(document.activeElement);
      var next = e.shiftKey ? (i <= 0 ? list.length - 1 : i - 1) : (i === -1 || i === list.length - 1 ? 0 : i + 1);
      e.preventDefault();
      list[next].focus();
      return;
    }
    if (gridKeys(e)) { e.preventDefault(); e.stopPropagation(); }
  });
  pop.addEventListener('mousedown', function (e) {
    if (!(e.target.closest && e.target.closest('button, [tabindex]'))) e.preventDefault();
  });
  pop.addEventListener('click', function (e) {
    e.stopPropagation();
    if (!(e.target.closest && e.target.closest('button'))) e.preventDefault();
  });
  scrim.addEventListener('mousedown', function (e) { e.preventDefault(); close(false); });

  function outside(e) {
    if (pop.contains(e.target) || (host.box && host.box.contains(e.target))) return;
    close(false);
  }

  function place() {
    if (!pop.isConnected) return;
    var mobile = ndMobile();
    pop.classList.toggle('nd-sheet', mobile);
    scrim.hidden = !mobile;
    var vw = document.documentElement.clientWidth || window.innerWidth;
    var vh = window.innerHeight;
    var origin = ndOrigin(host.mount);
    if (mobile) {
      pop.style.width = '';
      ndSetFixed(scrim, 0, 0, vw, origin);
      scrim.style.height = vh + 'px';
      ndSetFixed(pop, 0, vh - pop.offsetHeight, vw, origin);
      return;
    }
    pop.style.width = '';
    var anchor = host.box || host.mount;
    var r = anchor.getBoundingClientRect();
    var w = pop.offsetWidth;
    var h = pop.offsetHeight;
    var dirRtl = getComputedStyle(anchor).direction === 'rtl';
    var x = dirRtl ? r.right - w : r.left;
    x = Math.max(8, Math.min(vw - w - 8, x));
    var y = r.bottom + 6;
    if (y + h > vh - 8 && r.top - h - 6 >= 8) y = r.top - h - 6;
    y = Math.max(8, Math.min(Math.max(8, vh - h - 8), y));
    ndSetFixed(pop, x, y, 0, origin);
  }

  var closed = false;
  function close(restoreFocus) {
    if (closed) return;
    closed = true;
    document.removeEventListener('mousedown', outside, true);
    window.removeEventListener('keydown', escKey, true);
    window.removeEventListener('resize', place);
    window.removeEventListener('scroll', onScroll, true);
    if (ndOpen && ndOpen.pop === pop) ndOpen = null;
    pop.classList.add('nd-closing');
    var done = function () { pop.remove(); scrim.remove(); };
    if (ndReduced()) done(); else setTimeout(done, 120);
    if (host.onClose) host.onClose(restoreFocus);
  }

  function onScroll(e) {
    if (pop.contains(e.target)) return;
    place();
  }

  var mount = host.mount;
  mount.appendChild(scrim);
  mount.appendChild(pop);
  render();
  place();
  requestAnimationFrame(place);
  document.addEventListener('mousedown', outside, true);
  window.addEventListener('keydown', escKey, true);
  window.addEventListener('resize', place);
  window.addEventListener('scroll', onScroll, true);
  if (mode === 'time') spins.h.focus({ preventScroll: true });
  else focusCursor();
  ndOpen = { pop: pop, close: close, host: host };
  return ndOpen;
}

var ND_EXAMPLES = {
  persian: { date: '1404/10/11', datetime: '1404/10/11 14:30', range: '1404/10/01 – 1404/10/30' },
  gregory: { date: '2026/01/01', datetime: '2026/01/01 14:30', range: '2026/01/01 – 2026/01/31' }
};

function ndSame(a, b) {
  if (a === b) return true;
  if (!a || !b) return !a && !b;
  if (typeof a === 'object' || typeof b === 'object') return !!a && !!b && a.from === b.from && a.to === b.to;
  return false;
}

function ndCleanValue(v, mode) {
  if (v === undefined || v === null || v === '') return null;
  if (mode === 'range') {
    if (typeof v !== 'object') return null;
    var from = dayToJdn(v.from) != null ? v.from : null;
    var to = dayToJdn(v.to) != null ? v.to : null;
    return from || to ? { from: from, to: to } : null;
  }
  if (mode === 'date') {
    var s = String(v);
    if (dayToJdn(s.slice(0, 10)) != null && /^\d{4}-\d{2}-\d{2}$/.test(s)) return s;
    var d = new Date(s);
    return isNaN(d.getTime()) ? null : d.toISOString().slice(0, 10);
  }
  if (mode === 'time') {
    var t = ndParseTime(ndLatin(String(v)));
    return t ? ndPad(t.h) + ':' + ndPad(t.mi) : null;
  }
  var ms = new Date(v).getTime();
  return isNaN(ms) ? null : new Date(ms).toISOString();
}

function ndBoundJdn(v, mode, zone) {
  if (!v) return null;
  if (mode === 'date' || mode === 'range') return dayToJdn(String(v).slice(0, 10));
  return isoToJdn(v, zone);
}

function ndInput(o) {
  o = o || {};
  var mode = ['date', 'datetime', 'time', 'range'].indexOf(o.mode) === -1 ? 'datetime' : o.mode;
  var lang = ndLang();
  var cal = o.calendar || activeCalendar(lang);
  var zone = o.zone || ndZone();
  var value = ndCleanValue(o.value, mode);
  var min = o.min || null;
  var max = o.max || null;
  var id = o.id || ('nd-in-' + (++ndSeq));
  var zoneId = id + '-zone';
  var wrap = ndEl('div', { class: 'nd-field' + (o.className ? ' ' + o.className : ''), 'data-nd': mode, 'data-calendar': cal, 'data-l10n-skip': '' });
  var box = ndEl('div', { class: 'nd-box' });
  var input = ndEl('input', { type: 'text', id: id, class: 'nd-input' + (o.inputClass ? ' ' + o.inputClass : ''), name: o.name || null, autocomplete: 'off', spellcheck: 'false', dir: 'ltr', inputmode: ndMobile() ? 'none' : mode === 'time' ? 'numeric' : 'text', 'aria-label': o.label || o.name || null, 'aria-haspopup': 'dialog', 'aria-expanded': 'false', 'aria-describedby': o.zoneLine === false ? null : zoneId, placeholder: '\u200e' + (o.placeholder || (mode === 'time' ? ndDigits('14:30', lang) : ndDigits(ND_EXAMPLES[cal][mode], lang))) });
  Object.keys(o.attrs || {}).forEach(function (k) { input.setAttribute(k, o.attrs[k]); });
  if (o.required) input.setAttribute('aria-required', 'true');
  var clearBtn = ndEl('button', { type: 'button', class: 'nd-clear', 'aria-label': ndt('clear'), title: ndt('clear'), html: ndIcon('x'), 'data-nd-clear': '', hidden: true });
  var openBtn = ndEl('button', { type: 'button', class: 'nd-open', 'aria-label': ndt('open'), title: ndt('open'), 'aria-haspopup': 'dialog', 'aria-expanded': 'false', html: ndIcon(mode === 'time' ? 'clock' : 'calendar'), 'data-nd-open': '' });
  var line = o.zoneLine === false ? null : ndEl('p', { class: 'nd-zone', id: zoneId, 'data-nd-zone': zone, dir: lang === 'fa' ? 'rtl' : 'ltr' });
  Object.keys(o.zoneAttrs || {}).forEach(function (k) { if (line) line.setAttribute(k, o.zoneAttrs[k]); });
  var err = ndEl('p', { class: 'nd-error', hidden: true, 'data-nd-error': '' });
  box.appendChild(input);
  box.appendChild(clearBtn);
  box.appendChild(openBtn);
  wrap.appendChild(box);
  if (line) wrap.appendChild(line);
  wrap.appendChild(err);
  var disabled = false;
  var picker = null;
  var quiet = false;

  function shown() {
    return ndNumeric(value, mode, { calendar: cal, zone: zone, lang: lang });
  }

  function paintLine() {
    if (!line) return;
    var label = o.zoneLabel || zone;
    if (mode === 'datetime' && value) line.textContent = ndt('zoneAt', { time: ndFormat(value, { calendar: cal, zone: zone, style: 'long' }), zone: label });
    else if (mode === 'date' && value) line.textContent = ndt('zoneAt', { time: ndFormat(value, { calendar: cal, style: 'long' }), zone: label });
    else line.textContent = ndt('zone', { zone: label });
  }

  function paint(keepText) {
    if (!keepText) input.value = shown();
    clearBtn.hidden = !value || disabled || !!o.required;
    paintLine();
    wrap.setAttribute('data-value', value == null ? '' : typeof value === 'object' ? (value.from || '') + '/' + (value.to || '') : value);
  }

  function fire() {
    quiet = true;
    try {
      input.dispatchEvent(new Event('input', { bubbles: true }));
      input.dispatchEvent(new Event('change', { bubbles: true }));
    } finally { quiet = false; }
  }

  function set(v, opts) {
    v = ndCleanValue(v, mode);
    var changed = !ndSame(v, value);
    value = v;
    err.hidden = true;
    err.removeAttribute('role');
    input.removeAttribute('aria-invalid');
    paint(opts && opts.keepText);
    if (changed && !(opts && opts.silent)) {
      if (o.onchange) o.onchange(value);
      if (!(opts && opts.keepText)) fire();
    }
  }

  function invalid() {
    input.setAttribute('aria-invalid', 'true');
    err.textContent = ndt(mode === 'time' ? 'invalidTime' : 'invalid', { example: ndDigits(ND_EXAMPLES[cal][mode] || '14:30', lang) });
    err.hidden = false;
    err.setAttribute('role', 'alert');
  }

  function fromText(final) {
    var r = ndParse(input.value, mode, zone);
    if (!r) {
      if (final && o.passInvalid && o.onchange) { value = null; o.onchange(ndLatin(input.value).trim()); return; }
      if (final) invalid();
      return;
    }
    if (r.empty) { set(null, { keepText: !final }); return; }
    var v = r.value;
    if (mode === 'datetime' && v) {
      if (min && new Date(v) < new Date(min)) { if (final) invalid(); return; }
      if (max && new Date(v) > new Date(max)) { if (final) invalid(); return; }
    }
    set(v, { keepText: !final });
  }

  function open() {
    if (disabled || picker) return;
    if (ndParse(input.value, mode, zone) && !ndSame(ndParse(input.value, mode, zone).value, value)) fromText(true);
    picker = openPicker({
      mode: mode,
      calendar: cal,
      zone: zone,
      label: o.label || null,
      min: min,
      max: max,
      minJdn: ndBoundJdn(min, mode, zone),
      maxJdn: ndBoundJdn(max, mode, zone),
      required: !!o.required,
      box: box,
      mount: wrap,
      getValue: function () { return value; },
      commit: function (v) { set(v); },
      onClose: function (restore) {
        picker = null;
        input.setAttribute('aria-expanded', 'false');
        openBtn.setAttribute('aria-expanded', 'false');
        wrap.classList.remove('is-open');
        if (restore) (ndMobile() ? openBtn : input).focus({ preventScroll: true });
      }
    });
    input.setAttribute('aria-expanded', 'true');
    input.setAttribute('aria-controls', picker.pop.id);
    openBtn.setAttribute('aria-expanded', 'true');
    wrap.classList.add('is-open');
  }

  input.addEventListener('input', function () { if (!quiet) fromText(false); });
  input.addEventListener('change', function () { if (!quiet) fromText(true); });
  input.addEventListener('blur', function () { if (!picker && !input.getAttribute('aria-invalid')) paint(); });
  input.addEventListener('keydown', function (e) {
    if ((e.key === 'ArrowDown' && (e.altKey || !input.value)) || (e.key === ' ' && !input.value)) { e.preventDefault(); open(); }
    else if (e.key === 'Enter') { fromText(true); if (!input.getAttribute('aria-invalid')) paint(); }
  });
  input.addEventListener('click', function () { if (ndMobile()) open(); });
  input.addEventListener('focus', function () { if (ndMobile()) { input.blur(); open(); } });
  openBtn.addEventListener('click', function (e) { e.preventDefault(); if (picker) closePicker(true); else open(); });
  clearBtn.addEventListener('click', function (e) { e.preventDefault(); set(null); input.focus(); });

  Object.defineProperty(wrap, 'value', { configurable: true, get: function () { return value; }, set: function (v) { set(v, { silent: true }); } });
  Object.defineProperty(wrap, 'disabled', { configurable: true, get: function () { return disabled; }, set: function (v) {
    disabled = !!v;
    input.disabled = disabled;
    openBtn.disabled = disabled;
    wrap.classList.toggle('is-disabled', disabled);
    if (disabled && picker) closePicker(false);
    paint();
  } });
  wrap.read = function () { return value; };
  wrap.set = function (v, opts) { set(v, opts || {}); };
  wrap.input = input;
  wrap.open = open;
  wrap.close = function () { if (picker) closePicker(false); };
  wrap.focus = function () { input.focus(); };
  wrap.setMin = function (v) { min = v || null; };
  wrap.setMax = function (v) { max = v || null; };
  wrap.zone = zone;
  wrap.mode = mode;
  if (o.disabled) wrap.disabled = true;
  paint();
  return wrap;
}

var ND_NATIVE_TYPES = { date: 'date', 'datetime-local': 'datetime', time: 'time' };
var ndNative = { observer: null, items: [] };
var ndValueDesc = typeof HTMLInputElement !== 'undefined' ? Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value') : null;

function nativeToWire(raw, mode) {
  if (!raw) return null;
  if (mode === 'datetime') {
    var m = /^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2})/.exec(raw);
    return m ? m[1] + 'T' + m[2] + ':00.000Z' : null;
  }
  if (mode === 'time') return raw.slice(0, 5);
  return raw.slice(0, 10);
}

function wireToNative(v, mode, withSeconds) {
  if (!v) return '';
  if (mode === 'datetime') return new Date(v).toISOString().slice(0, 16);
  if (mode === 'time') return v + (withSeconds ? ':00' : '');
  return v;
}

function upgradeNative(el) {
  if (el.__ndUpgraded || !ndValueDesc) return;
  var mode = ND_NATIVE_TYPES[(el.getAttribute('type') || '').toLowerCase()];
  if (!mode || el.closest('.nd-field, [data-nd-skip]')) return;
  el.__ndUpgraded = true;
  var withSeconds = mode === 'time' && el.getAttribute('step') === '1';
  var syncing = false;
  var label = el.getAttribute('aria-label') || el.getAttribute('placeholder') || null;
  var lab = el.id && document.querySelector('label[for="' + (window.CSS && CSS.escape ? CSS.escape(el.id) : el.id) + '"]');
  if (lab) label = lab.textContent.trim() || label;
  var picker = ndInput({
    mode: mode,
    value: nativeToWire(ndValueDesc.get.call(el), mode),
    label: label,
    inputClass: el.className,
    zoneLine: mode !== 'time',
    onchange: function (v) {
      syncing = true;
      try {
        ndValueDesc.set.call(el, wireToNative(v, mode, withSeconds));
        el.dispatchEvent(new Event('input', { bubbles: true }));
        el.dispatchEvent(new Event('change', { bubbles: true }));
      } finally { syncing = false; }
    }
  });
  picker.classList.add('nd-native');
  picker.setAttribute('data-nd-native', el.getAttribute('type'));
  Object.defineProperty(el, 'value', {
    configurable: true,
    get: function () { return ndValueDesc.get.call(el); },
    set: function (v) {
      ndValueDesc.set.call(el, v);
      if (!syncing) picker.set(nativeToWire(v, mode), { silent: true });
    }
  });
  el.classList.add('nd-native-source');
  el.setAttribute('aria-hidden', 'true');
  el.setAttribute('tabindex', '-1');
  picker.disabled = el.disabled;
  el.parentNode.insertBefore(picker, el.nextSibling);
  ndNative.items.push({ el: el, picker: picker });
}

function scanNative(root) {
  if (!root || !root.querySelectorAll) return;
  if (root.matches && root.matches('input[type="date"], input[type="datetime-local"], input[type="time"]')) upgradeNative(root);
  Array.prototype.forEach.call(root.querySelectorAll('input[type="date"], input[type="datetime-local"], input[type="time"]'), upgradeNative);
}

function tidyNative() {
  ndNative.items = ndNative.items.filter(function (it) {
    if (it.el.isConnected) {
      if (it.picker.disabled !== it.el.disabled) it.picker.disabled = it.el.disabled;
      if (it.picker.parentNode !== it.el.parentNode || it.el.nextSibling !== it.picker) it.el.parentNode.insertBefore(it.picker, it.el.nextSibling);
      return true;
    }
    it.picker.remove();
    return false;
  });
}

function watchNative() {
  if (ndNative.observer || typeof MutationObserver === 'undefined' || !document.body) return;
  var pending = false;
  ndNative.observer = new MutationObserver(function (list) {
    for (var i = 0; i < list.length; i++) {
      var r = list[i];
      if (r.type === 'attributes') { pending = true; continue; }
      for (var j = 0; j < r.addedNodes.length; j++) if (r.addedNodes[j].nodeType === 1) scanNative(r.addedNodes[j]);
      if (r.removedNodes.length) pending = true;
    }
    if (pending && ndNative.items.length) { pending = false; tidyNative(); }
  });
  ndNative.observer.observe(document.body, { childList: true, subtree: true, attributes: true, attributeFilter: ['disabled'] });
  scanNative(document.body);
}

var NakamaDate = {
  version: 1,
  input: ndInput,
  format: ndFormat,
  numeric: ndNumeric,
  parse: ndParse,
  calendar: function () { return activeCalendar(); },
  calendarPref: calendarPref,
  setCalendarPref: setCalendarPref,
  calendarName: function (id) { return ndt('cal.' + (id || activeCalendar())); },
  t: ndt,
  zone: ndZone,
  setZone: setZone,
  defaultZone: ND_DEFAULT_ZONE,
  loadSettings: loadSettings,
  saveSettings: saveSettings,
  settings: function () { return ndState.meta; },
  onChange: ndOnChange,
  wallToIso: wallToIso,
  isoToWall: isoToWall,
  zoneOffset: zoneOffset,
  validZone: validZone,
  today: function (zone) { return jdnToDay(todayJdn(zone)); },
  digits: ndDigits,
  latin: ndLatin,
  installIntl: installIntl,
  close: closePicker,
  icon: ndIcon,
  jalali: { toJalali: toJalali, toGregorian: toGregorian, isLeap: jalaliLeap, monthLength: jalaliMonthLength },
  days: { fromDay: dayToJdn, toDay: jdnToDay, weekday: jdnWeekday, calendars: ND_CALS }
};

if (typeof window !== 'undefined') {
  window.NakamaDate = NakamaDate;
  if (window.NakamaL10n) installIntl();
  if (typeof document !== 'undefined' && document.addEventListener) {
    var ndBoot = function () { installIntl(); watchNative(); ndWatchLogin(); };
    if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', ndBoot);
    else ndBoot();
  }
}
})();
