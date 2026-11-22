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
