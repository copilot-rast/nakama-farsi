import fs from 'node:fs';
import path from 'node:path';
import type { Locator, Page } from '@playwright/test';
import { expect, test } from '../../lib/fixtures';
import { base, consoleApi, eco, uid } from '../../lib/api';
import { field, fillTranslated, open, openSettings, panel, psWindow, root, savePanel, settlePanel } from '../../lib/console';

const k = uid();
const SHOTS = process.env.E2E_SHOTS_DIR || '';
const made = { types: [] as string[], prizes: [] as string[], rules: [] as string[] };
const PERSIAN = '۰۱۲۳۴۵۶۷۸۹';
let eventKey = '';
let prizeKey = '';

test.describe.configure({ mode: 'serial' });

const fa = (s: string) => s.replace(/\d/g, (d) => PERSIAN[+d]);

function tehranIso(day: string, hm: string) {
  const [y, m, d] = day.split('-').map(Number);
  const [h, mi] = hm.split(':').map(Number);
  return new Date(Date.UTC(y, m - 1, d, h, mi) - 210 * 60000).toISOString();
}

function tehranDay(offsetDays: number) {
  const now = new Date(Date.now() + 210 * 60000 + offsetDays * 86400000);
  return now.toISOString().slice(0, 10);
}

function jalali(day: string) {
  const parts: Record<string, string> = {};
  new Intl.DateTimeFormat('en-US-u-ca-persian-nu-latn', { timeZone: 'UTC', year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(new Date(day + 'T12:00:00Z')).forEach((p) => { parts[p.type] = p.value; });
  return `${parseInt(parts.year, 10)}/${parts.month}/${parts.day}`;
}

function rangeOf(url: string) {
  const q = new URLSearchParams(url.split('?')[1] || '');
  return [q.get('from'), q.get('to')];
}

async function shot(page: Page, name: string, target?: Locator) {
  if (!SHOTS) return;
  fs.mkdirSync(SHOTS, { recursive: true });
  await page.waitForTimeout(250);
  if (target) await target.screenshot({ path: path.join(SHOTS, name + '.png') });
  else await page.screenshot({ path: path.join(SHOTS, name + '.png') });
}

async function pickIn(scope: Locator, text: string) {
  await scope.locator('button[aria-haspopup="listbox"]').first().click();
  const pop = scope.locator('.eco-pop').first();
  await expect(pop).toBeVisible();
  await pop.locator('input').fill(text);
  const opt = pop.locator('[role="option"]', { hasText: text }).first();
  await expect(opt).toBeVisible();
  await opt.click();
}

async function openPicker(f: Locator) {
  await f.locator('[data-nd-open]').click();
  const pop = f.locator('[data-nd-pop]');
  await expect(pop).toBeVisible();
  return pop;
}

async function pickDay(pop: Locator, day: string) {
  for (let i = 0; i < 26; i++) {
    const b = pop.locator(`button[data-day="${day}"]:not(.is-out)`);
    if (await b.count()) { await b.click(); return; }
    await pop.locator('[data-nd-nav="next"]').click();
  }
  throw new Error('day not reachable: ' + day);
}

async function setTime(pop: Locator, hm: string) {
  const [h, m] = hm.split(':');
  const hs = pop.locator('[data-nd-spin="h"]');
  await hs.focus();
  await hs.press('Home');
  await hs.pressSequentially(h);
  const ms = pop.locator('[data-nd-spin="mi"]');
  await ms.focus();
  await ms.press('Home');
  await ms.pressSequentially(m);
  await expect(hs).toHaveAttribute('aria-valuenow', String(+h));
  await expect(ms).toHaveAttribute('aria-valuenow', String(+m));
}

async function chooseCalendar(page: Page, id: 'auto' | 'persian' | 'gregory') {
  await page.locator('.pl-tb-end .pl-cal').click();
  const menu = page.locator('[data-pl-calendar-menu]');
  await expect(menu).toBeVisible();
  const item = menu.locator(`[data-cal="${id}"]`);
  if ((await item.getAttribute('aria-checked')) === 'true') { await page.keyboard.press('Escape'); return; }
  await Promise.all([page.waitForEvent('load'), item.click()]);
  await page.locator('.pl-tb-end').waitFor();
}

test.beforeAll(async () => {
  eventKey = 'e2e_' + k + '_dpev';
  prizeKey = 'e2e_' + k + '_dpprize';
  const et = await eco.post('/event-types', { key: eventKey, name: { fa: 'رویداد تقویم ' + k, en: 'Calendar event ' + k }, description: {}, clientAllowed: true, valueMin: 0, valueMax: 100, rateLimit: 100, isActive: true, sort: 0 });
  made.types.push(et.id);
  const pr = await eco.post('/prizes', { key: prizeKey, name: { fa: 'جایزهٔ تقویم ' + k, en: 'Calendar prize ' + k }, isRepeatable: true, items: [{ type: 'currency', key: 'coin', action: 'add', value: 1 }], tagIds: [], metadata: {} });
  made.prizes.push(pr.id);
  const st = await base.get('/console-settings');
  expect(st.timezone).toBe('Asia/Tehran');
});

test.afterAll(async () => {
  for (const id of made.rules) await eco.del('/event-rules/' + id);
  for (const id of made.prizes) await eco.del('/prizes/' + id);
  for (const id of made.types) await eco.del('/event-types/' + id);
  await base.put('/console-settings', { timezone: 'Asia/Tehran' });
});

async function newRule(page: Page, ruleKey: string, labels: { key: string; event: string; prize: RegExp; start: RegExp; end: RegExp }, start: [string, string], end: [string, string], snap?: string) {
  await open(page, '#/economy/event-rules/new');
  await settlePanel(page);
  const p = panel(page);
  await field(p, labels.key).locator('input').fill(ruleKey);
  await fillTranslated(p, 'name', { fa: 'قانون تقویم ' + k, en: 'Calendar rule ' + k });
  await pickIn(field(p, labels.event), eventKey);
  await pickIn(field(p, labels.prize), prizeKey);
  const s = field(p, labels.start);
  let pop = await openPicker(s);
  await pickDay(pop, start[0]);
  await setTime(pop, start[1]);
  if (snap) await shot(page, snap, panel(page).locator('.x-form'));
  await pop.locator('[data-nd-done]').click();
  await expect(pop).toHaveCount(0);
  const e = field(p, labels.end);
  pop = await openPicker(e);
  await pickDay(pop, end[0]);
  await setTime(pop, end[1]);
  await pop.locator('[data-nd-done]').click();
  await expect(s.locator('.nd-field')).toHaveAttribute('data-value', tehranIso(start[0], start[1]));
  await expect(e.locator('.nd-field')).toHaveAttribute('data-value', tehranIso(end[0], end[1]));
  await expect(s.locator('[data-nd-zone]')).toContainText('Asia/Tehran');
  await savePanel(page);
  const res = await eco.get('/event-rules?q=' + encodeURIComponent(ruleKey) + '&limit=10');
  const rule = (res.items || []).find((x: any) => x.key === ruleKey);
  made.rules.push(rule.id);
  return rule;
}

test.describe('date picker', () => {
  test('fa: an event rule window picked in the Jalali calendar is stored as Tehran time and shows the same after a reload', async ({ spawn }) => {
    const page = await spawn({ signedIn: true, mode: 'dark' });
    const startDay = tehranDay(20);
    const endDay = tehranDay(27);
    const rule = await newRule(page, 'e2e_' + k + '_dpfa', { key: 'کلید', event: 'رویداد', prize: /^جایزه$/, start: /^شروع$/, end: /^پایان$/ }, [startDay, '14:30'], [endDay, '09:05'], 'picker-fa-dark-datetime');
    expect(new Date(rule.startsAt).toISOString()).toBe(tehranIso(startDay, '14:30'));
    expect(new Date(rule.endsAt).toISOString()).toBe(tehranIso(endDay, '09:05'));

    await page.reload();
    await open(page, '#/economy/event-rules/' + rule.id);
    await settlePanel(page);
    const s = field(panel(page), /^شروع$/);
    await expect(s.locator('input')).toHaveValue(fa(jalali(startDay) + ' 14:30'));
    const pop = await openPicker(s);
    await expect(pop).toHaveAttribute('data-calendar', 'persian');
    await expect(pop).toHaveAttribute('dir', 'rtl');
    await expect(pop.locator('thead th').first()).toHaveText('ش');
    await expect(pop.locator(`button[data-day="${startDay}"]`)).toHaveClass(/is-sel/);
    await shot(page, 'picker-fa-dark', pop);
    await page.keyboard.press('ArrowLeft');
    await expect(pop.locator(`button[data-day="${startDay}"]`)).toHaveAttribute('tabindex', '-1');
    await page.keyboard.press('Escape');
    await expect(pop).toHaveCount(0);
    await expect(s.locator('input')).toBeFocused();
    await expect(s.locator('.nd-field')).toHaveAttribute('data-value', tehranIso(startDay, '14:30'));
  });

  test('the calendar override changes what is shown, never the stored value', async ({ spawn }) => {
    const page = await spawn({ signedIn: true });
    const id = made.rules[0];
    const before = await eco.get('/event-rules/' + id);
    await open(page, '#/economy/event-rules');
    await chooseCalendar(page, 'gregory');
    expect(await page.evaluate(() => localStorage.getItem('nakama.date.calendar'))).toBe('gregory');
    await open(page, '#/economy/event-rules/' + id);
    await settlePanel(page);
    const s = field(panel(page), /^شروع$/);
    const startDay = new Date(new Date(before.startsAt).getTime() + 210 * 60000).toISOString().slice(0, 10);
    await expect(s.locator('input')).toHaveValue(fa(startDay.replace(/-/g, '/') + ' 14:30'));
    const pop = await openPicker(s);
    await expect(pop).toHaveAttribute('data-calendar', 'gregory');
    await expect(pop).toHaveAttribute('dir', 'ltr');
    await page.keyboard.press('Escape');
    await page.keyboard.press('Escape');
    await expect(panel(page)).toHaveCount(0);
    const after = await eco.get('/event-rules/' + id);
    expect(after.startsAt).toBe(before.startsAt);
    expect(after.endsAt).toBe(before.endsAt);
    await chooseCalendar(page, 'auto');
    expect(await page.evaluate(() => localStorage.getItem('nakama.date.calendar'))).toBeNull();
    await open(page, '#/economy/event-rules/' + id);
    await settlePanel(page);
    await expect(field(panel(page), /^شروع$/).locator('input')).toHaveValue(fa(jalali(startDay) + ' 14:30'));
  });

  test('en: Gregorian with Latin digits and the week from the locale; the stored ISO is the same Tehran time', async ({ spawn }) => {
    const page = await spawn({ signedIn: true, locale: 'en' });
    const startDay = tehranDay(33);
    const endDay = tehranDay(34);
    const rule = await newRule(page, 'e2e_' + k + '_dpen', { key: 'Key', event: 'Event', prize: /^Prize$/, start: /^Starts$/, end: /^Ends$/ }, [startDay, '08:00'], [endDay, '23:45']);
    expect(new Date(rule.startsAt).toISOString()).toBe(tehranIso(startDay, '08:00'));
    expect(new Date(rule.endsAt).toISOString()).toBe(tehranIso(endDay, '23:45'));
    await open(page, '#/economy/event-rules/' + rule.id);
    await settlePanel(page);
    const s = field(panel(page), /^Starts$/);
    await expect(s.locator('input')).toHaveValue(startDay.replace(/-/g, '/') + ' 08:00');
    const pop = await openPicker(s);
    await expect(pop).toHaveAttribute('data-calendar', 'gregory');
    await expect(pop.locator('thead th').first()).toHaveText('Sun');
    await expect(pop).not.toContainText(/[۰-۹]/);
    await shot(page, 'picker-en-light', pop);
    await page.keyboard.press('Escape');
  });

  test('banner start, gift code expiry and typed text use the console zone', async ({ spawn }) => {
    const page = await spawn({ signedIn: true });
    await open(page, '#/publication/banners/new');
    await settlePanel(page);
    const b = field(panel(page), /^شروع$/);
    await expect(b.locator('[data-zone-shown]')).toContainText('Asia/Tehran');
    await b.locator('[data-zone-input]').fill('1405/01/01 10:00');
    await b.locator('[data-zone-input]').press('Enter');
    await expect(b.locator('.nd-field')).toHaveAttribute('data-value', '2026-03-21T06:30:00.000Z');
    await expect(b.locator('[data-zone-input]')).toHaveValue('۱۴۰۵/۰۱/۰۱ ۱۰:۰۰');
    await shot(page, 'picker-fa-light-field', b);
    await page.keyboard.press('Escape');

    await open(page, '#/economy/gifts/new');
    await settlePanel(page);
    const g = field(panel(page), 'پایان اعتبار');
    const day = tehranDay(10);
    const pop = await openPicker(g);
    await shot(page, 'picker-fa-light', pop);
    await pickDay(pop, day);
    await setTime(pop, '18:00');
    await pop.locator('[data-nd-done]').click();
    await expect(pop).toHaveCount(0);
    await expect(g.locator('.nd-field')).toHaveAttribute('data-value', tehranIso(day, '18:00'));
    await g.locator('[data-nd-clear]').click();
    await expect(g.locator('.nd-field')).toHaveAttribute('data-value', '');
    await expect(g.locator('input')).toHaveValue('');
  });

  test('quiet hours use the time picker and keep HH:MM', async ({ spawn }) => {
    const page = await spawn({ signedIn: true });
    await open(page, '#/publication/auto-rules/new');
    await settlePanel(page);
    const q = field(panel(page), 'سکوت از');
    const pop = await openPicker(q);
    await expect(pop).toHaveAttribute('data-nd-pop', 'time');
    await setTime(pop, '22:15');
    await shot(page, 'picker-time', pop);
    await pop.locator('[data-nd-done]').click();
    await expect(q.locator('input')).toHaveValue('۲۲:۱۵');
    await expect(q.locator('.nd-field')).toHaveAttribute('data-value', '22:15');
    await expect(q.locator('[data-nd-zone]')).toContainText('منطقه‌ی زمانی');
  });

  test('insights range: presets fill a day range in the console zone', async ({ spawn }) => {
    const page = await spawn({ signedIn: true });
    await page.goto('/#/insights/active');
    const range = page.locator('[data-ins-range]');
    await expect(range).toBeVisible({ timeout: 20_000 });
    const pop = await openPicker(range);
    await expect(pop.locator('[data-nd-preset]')).toHaveCount(4);
    await shot(page, 'picker-range-presets', pop);
    await pop.locator('[data-nd-preset="7"]').click();
    const to = tehranDay(0);
    const from = tehranDay(-6);
    await expect.poll(() => rangeOf(page.url())).toEqual([from, to]);
    await expect(page.locator('[data-ins-range] input')).toHaveValue(fa(jalali(from) + ' – ' + jalali(to)));
    const again = await openPicker(page.locator('[data-ins-range]'));
    const days = again.locator('button[data-day]:not(.is-out):not(.is-off)');
    const a = await days.nth(2).getAttribute('data-day');
    const c = await days.nth(4).getAttribute('data-day');
    await days.nth(4).click();
    await days.nth(2).click();
    await expect.poll(() => rangeOf(page.url())).toEqual([a, c]);
  });

  test('a native jsonforms date-time input is upgraded and writes UTC the way the renderer reads it', async ({ spawn }) => {
    const page = await spawn({ signedIn: true });
    await page.goto('/#/players');
    await page.locator('.pl-tb-end').waitFor();
    await page.evaluate(() => {
      const box = document.createElement('div');
      box.id = 'e2e-native-form';
      box.style.cssText = 'position:fixed;top:120px;inset-inline-start:340px;width:320px;z-index:50;padding:16px;background:var(--surface);border:1px solid var(--line-1);border-radius:8px';
      box.innerHTML = '<label for="e2e-native-dt" class="block text-sm mb-1">endTime</label>';
      const el = document.createElement('input');
      el.type = 'datetime-local';
      el.id = 'e2e-native-dt';
      el.className = 'text-xs p-1 appearance-none block w-full';
      el.value = '2026-01-01T06:30';
      (window as any).__nativeOut = [];
      el.addEventListener('change', () => (window as any).__nativeOut.push(el.value === '' ? null : el.value + ':00.000Z'));
      box.appendChild(el);
      document.querySelector('#app')!.appendChild(box);
    });
    const up = page.locator('#e2e-native-form .nd-field[data-nd-native="datetime-local"]');
    await expect(up).toBeVisible();
    await expect(up.locator('input')).toHaveValue('۱۴۰۴/۱۰/۱۱ ۱۰:۰۰');
    await expect(page.locator('#e2e-native-dt')).toBeHidden();
    await shot(page, 'picker-native-field', page.locator('#e2e-native-form'));
    await up.locator('input').fill('1404/10/12 08:00');
    await up.locator('input').press('Enter');
    expect(await page.evaluate(() => (document.getElementById('e2e-native-dt') as HTMLInputElement).value)).toBe('2026-01-02T04:30');
    expect(await page.evaluate(() => (window as any).__nativeOut.at(-1))).toBe('2026-01-02T04:30:00.000Z');
    await page.evaluate(() => { (document.getElementById('e2e-native-dt') as HTMLInputElement).value = '2026-03-20T20:30'; });
    await expect(up.locator('input')).toHaveValue('۱۴۰۵/۰۱/۰۱ ۰۰:۰۰');
  });

  test('mobile: the picker opens as a bottom sheet', async ({ spawn }) => {
    const page = await spawn({ signedIn: true, mobile: true });
    await open(page, '#/economy/gifts/new');
    await settlePanel(page);
    const g = field(panel(page), 'شروع اعتبار');
    await g.locator('[data-nd-open]').click();
    const pop = g.locator('[data-nd-pop]');
    await expect(pop).toHaveClass(/nd-sheet/);
    const box = (await pop.boundingBox())!;
    const vp = page.viewportSize()!;
    expect(Math.round(box.y + box.height)).toBeGreaterThanOrEqual(vp.height - 1);
    expect(Math.round(box.width)).toBeGreaterThanOrEqual(vp.width - 1);
    const day = pop.locator('button[data-day]').first();
    expect((await day.boundingBox())!.height).toBeGreaterThanOrEqual(43);
    await shot(page, 'picker-mobile-sheet');
    await pop.locator('[data-nd-done]').click();
    await expect(pop).toHaveCount(0);
  });

  test('the console time zone is an admin setting, found with Ctrl+K, used by the pickers', async ({ spawn }) => {
    const page = await spawn({ signedIn: true });
    await page.goto('/#/players');
    await page.locator('.pl-tb-end').waitFor();
    await page.keyboard.press('Control+k');
    await page.locator('.pl-pal-input').fill('تاریخ و ساعت');
    await expect(page.locator('.pl-pal-row').first()).toContainText('تاریخ و ساعت');
    await page.locator('.pl-pal-input').fill('تقویم');
    await expect(page.locator('.pl-pal-row')).toHaveCount(3);
    await page.keyboard.press('Escape');
    await openSettings(page, 'datetime');
    const ps = psWindow(page);
    const sel = ps.locator('[data-settings-form="datetime"] select');
    await expect(sel).toHaveValue('Asia/Tehran');
    await sel.selectOption('Europe/Berlin');
    await ps.locator('[data-settings-save]').click();
    await expect.poll(async () => (await base.get('/console-settings')).timezone).toBe('Europe/Berlin');
    await expect.poll(() => page.evaluate(() => (window as any).NakamaDate.zone())).toBe('Europe/Berlin');
    await open(page, '#/economy/gifts/new');
    await settlePanel(page);
    await expect(field(panel(page), 'پایان اعتبار').locator('[data-nd-zone]')).toContainText('Europe/Berlin');
    await base.put('/console-settings', { timezone: 'Asia/Tehran' });
    expect((await consoleApi('PUT', '/v2/console/base/console-settings', { timezone: 'Mars/Base' })).status).toBe(400);
  });
});
