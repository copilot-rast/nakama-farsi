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
