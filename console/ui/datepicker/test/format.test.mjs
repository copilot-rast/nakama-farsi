import assert from 'node:assert/strict';
import test from 'node:test';
import { load } from './load.mjs';

const iso = '2026-01-01T06:30:00.000Z';

test('fa defaults to Jalali with Persian digits', () => {
  const { ND } = load({ lang: 'fa-IR' });
  assert.equal(ND.calendar(), 'persian');
  const s = ND.format(iso);
  assert.match(s, /۱۴۰۴/);
  assert.match(s, /دی/);
  assert.match(s, /۱۰:۰۰/);
  assert.equal(ND.numeric(iso, 'datetime'), '۱۴۰۴/۱۰/۱۱ ۱۰:۰۰');
  assert.equal(ND.format('2025-03-21'), ND.format('2025-03-21', { time: false }));
  assert.match(ND.format('2025-03-21'), /فروردین/);
  assert.equal(ND.format('09:30'), '۰۹:۳۰');
});

test('en defaults to Gregorian with Latin digits', () => {
  const { ND } = load({ lang: 'en' });
  assert.equal(ND.calendar(), 'gregory');
  const s = ND.format(iso);
  assert.match(s, /2026/);
  assert.match(s, /Jan/);
  assert.match(s, /10:00/);
  assert.doesNotMatch(s, /[۰-۹]/);
  assert.equal(ND.numeric(iso, 'datetime'), '2026/01/01 10:00');
});

test('the override changes the calendar but never the digits or the wire value', () => {
  const fa = load({ lang: 'fa-IR', storage: { 'nakama.date.calendar': 'gregory' } }).ND;
  assert.equal(fa.calendar(), 'gregory');
  assert.equal(fa.numeric(iso, 'datetime'), '۲۰۲۶/۰۱/۰۱ ۱۰:۰۰');
  const en = load({ lang: 'en', storage: { 'nakama.date.calendar': 'persian' } }).ND;
  assert.equal(en.calendar(), 'persian');
  assert.equal(en.numeric(iso, 'datetime'), '1404/10/11 10:00');
  assert.match(en.format(iso), /1404/);
  assert.equal(en.parse(en.numeric(iso, 'datetime'), 'datetime', 'Asia/Tehran').value, iso);
});

test('an explicit calendar and zone win over the defaults', () => {
  const { ND } = load({ lang: 'fa-IR' });
  assert.match(ND.format(iso, { calendar: 'gregory' }), /۲۰۲۶/);
  assert.match(ND.format(iso, { zone: 'UTC', calendar: 'gregory', lang: 'en' }), /6:30|06:30/);
  assert.equal(ND.format(''), '');
  assert.equal(ND.format('not a date'), 'not a date');
});

test('setting the preference stores it and reloads', () => {
  const { ND, ctx, store } = load({ lang: 'fa-IR' });
  ND.setCalendarPref('gregory');
  assert.equal(store.get('nakama.date.calendar'), 'gregory');
  assert.equal(ctx.reloaded, 1);
  ND.setCalendarPref('auto', { reload: false });
  assert.equal(store.has('nakama.date.calendar'), false);
  assert.equal(ND.calendarPref(), 'auto');
});

test('the Intl patch fills in calendar and zone only for default-locale calls', () => {
  const { ND, ctx } = load({ lang: 'en', storage: { 'nakama.date.calendar': 'persian' } });
  ND.installIntl();
  const d = new ctx.Date(iso);
  assert.match(new ctx.Intl.DateTimeFormat(undefined, { year: 'numeric' }).format(d), /1404/);
  assert.match(new ctx.Intl.DateTimeFormat('en-US', { year: 'numeric' }).format(d), /2026/);
  assert.equal(new ctx.Intl.DateTimeFormat().resolvedOptions().timeZone, 'Asia/Tehran');
  assert.match(d.toLocaleString(undefined, { hour: '2-digit', minute: '2-digit', hourCycle: 'h23' }), /10:00/);
  assert.equal(ND.installIntl(), false);
});

test('the long style puts the weekday first in both languages', () => {
  const fa = load({ lang: 'fa-IR' }).ND;
  assert.equal(fa.format('2026-12-12T11:00:00Z', { style: 'long' }), 'شنبه ۲۱ آذر ۱۴۰۵ ساعت ۱۴:۳۰');
  const en = load({ lang: 'en' }).ND;
  assert.equal(en.format('2026-12-12T11:00:00Z', { style: 'long' }), 'Saturday, December 12, 2026 at 14:30');
});
