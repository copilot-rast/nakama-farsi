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
