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
