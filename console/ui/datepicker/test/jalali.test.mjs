import assert from 'node:assert/strict';
import test from 'node:test';
import { load } from './load.mjs';

const { ND } = load();
const J = ND.jalali;

function intlJalali(ms) {
  const parts = {};
  new Intl.DateTimeFormat('en-US-u-ca-persian-nu-latn', { timeZone: 'UTC', year: 'numeric', month: 'numeric', day: 'numeric' }).formatToParts(new Date(ms)).forEach((p) => { parts[p.type] = p.value; });
  return { y: parseInt(parts.year, 10), m: +parts.month, d: +parts.day };
}

test('every day from 1925 to 2075 matches the Persian calendar in ICU', () => {
  const start = Date.UTC(1925, 0, 1);
  const end = Date.UTC(2075, 11, 31);
  for (let ms = start; ms <= end; ms += 86400000) {
    const g = new Date(ms);
    const j = J.toJalali(g.getUTCFullYear(), g.getUTCMonth() + 1, g.getUTCDate());
    const want = intlJalali(ms);
    assert.deepEqual({ y: j.y, m: j.m, d: j.d }, want, g.toISOString().slice(0, 10));
    const back = J.toGregorian(j.y, j.m, j.d);
    assert.deepEqual([back.y, back.m, back.d], [g.getUTCFullYear(), g.getUTCMonth() + 1, g.getUTCDate()]);
  }
});

test('leap years across one 33-year cycle', () => {
  const leaps = [];
  for (let y = 1370; y <= 1403; y++) if (J.isLeap(y)) leaps.push(y);
  assert.deepEqual(leaps, [1370, 1375, 1379, 1383, 1387, 1391, 1395, 1399, 1403]);
  assert.equal(J.isLeap(1404), false);
  assert.equal(J.isLeap(1407), false);
  assert.equal(J.isLeap(1408), true);
  for (let y = 1300; y <= 1500; y++) assert.equal(J.monthLength(y, 12), J.isLeap(y) ? 30 : 29, String(y));
});

test('Nowruz boundary', () => {
  assert.deepEqual({ ...J.toJalali(2025, 3, 20) }, { y: 1403, m: 12, d: 30 });
  assert.deepEqual({ ...J.toJalali(2025, 3, 21) }, { y: 1404, m: 1, d: 1 });
  assert.deepEqual({ ...J.toJalali(2024, 3, 19) }, { y: 1402, m: 12, d: 29 });
  assert.deepEqual({ ...J.toJalali(2024, 3, 20) }, { y: 1403, m: 1, d: 1 });
  assert.deepEqual({ ...J.toGregorian(1405, 1, 1) }, { y: 2026, m: 3, d: 21 });
});

test('end of Esfand in common and leap years', () => {
  assert.deepEqual({ ...J.toGregorian(1403, 12, 30) }, { y: 2025, m: 3, d: 20 });
  assert.deepEqual({ ...J.toGregorian(1404, 12, 29) }, { y: 2026, m: 3, d: 20 });
  assert.equal(J.monthLength(1404, 12), 29);
  assert.equal(J.monthLength(1403, 12), 30);
  assert.equal(J.monthLength(1404, 6), 31);
  assert.equal(J.monthLength(1404, 7), 30);
  assert.equal(ND.parse('1404/12/30', 'date'), null);
  assert.deepEqual({ ...ND.parse('1403/12/30', 'date') }, { value: '2025-03-20' });
});

test('weekdays come from the day number', () => {
  for (const day of ['2025-03-21', '2026-01-01', '1999-12-31', '2024-02-29']) {
    assert.equal(ND.days.weekday(ND.days.fromDay(day)), new Date(day + 'T00:00:00Z').getUTCDay(), day);
  }
  assert.equal(ND.days.toDay(ND.days.fromDay('2024-02-29')), '2024-02-29');
});
