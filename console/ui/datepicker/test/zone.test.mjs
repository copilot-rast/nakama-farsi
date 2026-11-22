import assert from 'node:assert/strict';
import test from 'node:test';
import { load } from './load.mjs';

const { ND, store } = load();

test('Tehran has no daylight saving from 2023 on', () => {
  for (let y = 2023; y <= 2030; y++) {
    for (const md of ['01-15', '04-15', '07-15', '10-15']) {
      const iso = ND.wallToIso(`${y}-${md}T12:00`, 'Asia/Tehran');
      assert.equal(iso, `${y}-${md}T08:30:00.000Z`, `${y}-${md}`);
      assert.equal(ND.isoToWall(iso, 'Asia/Tehran'), `${y}-${md}T12:00`);
    }
  }
});

test('wall clock round-trips in other zones too', () => {
  assert.equal(ND.wallToIso('2026-07-01T09:00', 'Europe/Berlin'), '2026-07-01T07:00:00.000Z');
  assert.equal(ND.wallToIso('2026-01-01T09:00', 'Europe/Berlin'), '2026-01-01T08:00:00.000Z');
  assert.equal(ND.wallToIso('2026-01-01T09:00', 'UTC'), '2026-01-01T09:00:00.000Z');
  assert.equal(ND.isoToWall('2026-03-20T20:30:00Z', 'Asia/Tehran'), '2026-03-21T00:00');
});

test('console zone defaults to Asia/Tehran and keeps a valid choice', () => {
  assert.equal(ND.zone(), 'Asia/Tehran');
  assert.equal(ND.setZone('Mars/Base'), false);
  assert.equal(ND.zone(), 'Asia/Tehran');
  assert.equal(ND.setZone('Europe/Berlin'), true);
  assert.equal(ND.zone(), 'Europe/Berlin');
  assert.equal(store.get('nakama.date.zone'), 'Europe/Berlin');
  ND.setZone('Asia/Tehran');
});

test('typed text parses in either calendar and stays ISO on the wire', () => {
  assert.equal(ND.parse('۱۴۰۴/۱۰/۱۱ ۱۰:۰۰', 'datetime', 'Asia/Tehran').value, '2026-01-01T06:30:00.000Z');
  assert.equal(ND.parse('2026-01-01T10:00', 'datetime', 'Asia/Tehran').value, '2026-01-01T06:30:00.000Z');
  assert.equal(ND.parse('2026/01/01', 'datetime', 'Asia/Tehran').value, '2025-12-31T20:30:00.000Z');
  assert.equal(ND.parse('1404/01/01', 'date').value, '2025-03-21');
  assert.equal(ND.parse('۰۹:۰۵', 'time').value, '09:05');
  assert.equal(ND.parse('24:00', 'time'), null);
  assert.deepEqual({ ...ND.parse('1404/10/01 – 1404/10/30', 'range').value }, { from: '2025-12-22', to: '2026-01-20' });
  assert.deepEqual({ ...ND.parse('2026-01-31 to 2026-01-01', 'range').value }, { from: '2026-01-01', to: '2026-01-31' });
  assert.equal(ND.parse('1404/13/01', 'date'), null);
  assert.ok(ND.parse('', 'date').empty);
});
